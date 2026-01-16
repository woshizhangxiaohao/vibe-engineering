package services

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
	"vibe-backend/internal/models"
	"vibe-backend/internal/repository"
)

// ChatService handles AI chat operations.
type ChatService struct {
	chatRepo         *repository.ChatRepository
	videoRepo        *repository.VideoRepository
	insightRepo      *repository.InsightRepository
	openRouterAPIKey string
	chatModel        string
	httpClient       *http.Client
	log              *zap.Logger
}

// NewChatService creates a new ChatService.
func NewChatService(
	chatRepo *repository.ChatRepository,
	videoRepo *repository.VideoRepository,
	insightRepo *repository.InsightRepository,
	apiKey string,
	chatModel string,
	log *zap.Logger,
) *ChatService {
	if chatModel == "" {
		chatModel = "anthropic/claude-3-5-sonnet"
	}

	return &ChatService{
		chatRepo:         chatRepo,
		videoRepo:        videoRepo,
		insightRepo:      insightRepo,
		openRouterAPIKey: apiKey,
		chatModel:        chatModel,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
		log: log,
	}
}

// ChatStream sends a message and returns a channel for streaming responses.
func (s *ChatService) ChatStream(ctx context.Context, insightID uint, message string, highlightID *uint) (<-chan models.ChatStreamEvent, error) {
	// Get the insight for context
	insight, err := s.insightRepo.GetByID(ctx, insightID)
	if err != nil {
		return nil, fmt.Errorf("insight not found: %w", err)
	}

	// Get existing chat history
	history, err := s.chatRepo.GetMessagesByAnalysisID(ctx, insightID)
	if err != nil {
		s.log.Warn("Failed to get chat history", zap.Error(err))
		history = []models.ChatMessage{}
	}

	// Save user message
	userMessage := &models.ChatMessage{
		InsightID:   insightID,
		UserID:      0, // TODO: Get from context/auth
		Role:        "user",
		Content:     message,
		HighlightID: highlightID,
	}
	if err := s.chatRepo.CreateMessage(ctx, userMessage); err != nil {
		s.log.Error("Failed to save user message", zap.Error(err))
	}

	// Build system prompt with context
	systemPrompt := s.buildSystemPrompt(insight)

	// Build messages array
	messages := s.buildMessages(systemPrompt, history, message)

	// Create response channel
	responseChan := make(chan models.ChatStreamEvent, 100)

	// Start streaming in goroutine
	go func() {
		defer close(responseChan)
		s.streamFromOpenRouter(ctx, messages, insightID, responseChan)
	}()

	return responseChan, nil
}

// GetChatHistory returns the chat history for an analysis.
func (s *ChatService) GetChatHistory(ctx context.Context, analysisID uint) (*models.ChatHistoryResponse, error) {
	messages, err := s.chatRepo.GetMessagesByAnalysisID(ctx, analysisID)
	if err != nil {
		return nil, err
	}

	response := &models.ChatHistoryResponse{
		Messages: make([]models.ChatMessageResponse, len(messages)),
	}

	for i, msg := range messages {
		response.Messages[i] = models.ChatMessageResponse{
			ID:        msg.ID,
			Role:      msg.Role,
			Content:   msg.Content,
			CreatedAt: msg.CreatedAt,
		}
	}

	return response, nil
}

// AnalyzeEntities analyzes the content and returns detected entities and suggestions.
func (s *ChatService) AnalyzeEntities(ctx context.Context, insightID uint) (*models.AnalyzeEntitiesResponse, error) {
	insight, err := s.insightRepo.GetByID(ctx, insightID)
	if err != nil {
		return nil, fmt.Errorf("insight not found: %w", err)
	}

	// Build prompt for entity extraction
	prompt := fmt.Sprintf(`分析以下内容，提取所mentioned的股票、加密货币、公司等实体。

标题: %s
作者: %s
摘要: %s

请以JSON格式返回：
{
  "entities": [
    {"type": "stock|crypto|company", "name": "实体名称", "ticker": "代码"}
  ],
  "suggestions": [
    {"type": "position|prediction|analysis", "entity": "代码", "prompt": "建议的问题"}
  ]
}

只返回JSON，不要其他文字。`, insight.Title, insight.Author, insight.Summary)

	response, err := s.callOpenRouter(ctx, prompt)
	if err != nil {
		s.log.Error("Failed to analyze entities",
			zap.Uint("insight_id", insightID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to call AI service: %w", err)
	}

	// Parse response
	var result models.AnalyzeEntitiesResponse
	cleanedResponse := s.cleanJSONResponse(response)
	if err := json.Unmarshal([]byte(cleanedResponse), &result); err != nil {
		s.log.Error("Failed to parse entity response",
			zap.Uint("insight_id", insightID),
			zap.Error(err),
			zap.String("response", cleanedResponse),
		)
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return &result, nil
}

// buildSystemPrompt creates the system prompt with insight context.
func (s *ChatService) buildSystemPrompt(insight *models.Insight) string {
	return fmt.Sprintf(`你是一个智能阅读助手。用户正在阅读以下内容：

标题: %s
作者: %s
摘要: %s

请基于以上内容回答用户的问题。回答时：
1. 优先参考内容中的信息
2. 如果内容中没有相关信息，可以结合你的知识回答，但需说明
3. 保持回答简洁、有洞察力
4. 支持 Markdown 格式`, insight.Title, insight.Author, insight.Summary)
}

// buildMessages constructs the messages array for the API call.
func (s *ChatService) buildMessages(systemPrompt string, history []models.ChatMessage, newMessage string) []map[string]string {
	messages := make([]map[string]string, 0, len(history)+2)

	// Add system message
	messages = append(messages, map[string]string{
		"role":    "system",
		"content": systemPrompt,
	})

	// Add history (limit to last 10 messages to control context size)
	historyStart := 0
	if len(history) > 10 {
		historyStart = len(history) - 10
	}
	for _, msg := range history[historyStart:] {
		messages = append(messages, map[string]string{
			"role":    msg.Role,
			"content": msg.Content,
		})
	}

	// Add new user message
	messages = append(messages, map[string]string{
		"role":    "user",
		"content": newMessage,
	})

	return messages
}

// streamFromOpenRouter handles the SSE streaming from OpenRouter.
func (s *ChatService) streamFromOpenRouter(ctx context.Context, messages []map[string]string, insightID uint, responseChan chan<- models.ChatStreamEvent) {
	const openRouterURL = "https://openrouter.ai/api/v1/chat/completions"

	requestBody := map[string]interface{}{
		"model":    s.chatModel,
		"messages": messages,
		"stream":   true,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		s.log.Error("Failed to marshal request", zap.Error(err))
		responseChan <- models.ChatStreamEvent{Done: true}
		return
	}

	req, err := http.NewRequestWithContext(ctx, "POST", openRouterURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		s.log.Error("Failed to create request", zap.Error(err))
		responseChan <- models.ChatStreamEvent{Done: true}
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.openRouterAPIKey)
	req.Header.Set("Accept", "text/event-stream")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.log.Error("Failed to call OpenRouter API", zap.Error(err))
		responseChan <- models.ChatStreamEvent{Done: true}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.log.Error("OpenRouter API error", zap.Int("status", resp.StatusCode))
		responseChan <- models.ChatStreamEvent{Done: true}
		return
	}

	// Read SSE stream
	var fullContent strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}

		// Parse SSE data
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")

			// Check for stream end
			if data == "[DONE]" {
				break
			}

			// Parse chunk
			var chunk struct {
				Choices []struct {
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
					FinishReason *string `json:"finish_reason"`
				} `json:"choices"`
			}

			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				continue
			}

			if len(chunk.Choices) > 0 {
				content := chunk.Choices[0].Delta.Content
				if content != "" {
					fullContent.WriteString(content)
					responseChan <- models.ChatStreamEvent{
						Role:    "assistant",
						Content: content,
						Done:    false,
					}
				}
			}
		}
	}

	// Save assistant message
	if fullContent.Len() > 0 {
		assistantMessage := &models.ChatMessage{
			InsightID: insightID,
			UserID:    0, // TODO: Get from context/auth
			Role:      "assistant",
			Content:   fullContent.String(),
		}
		if err := s.chatRepo.CreateMessage(ctx, assistantMessage); err != nil {
			s.log.Error("Failed to save assistant message", zap.Error(err))
		}

		// Send final event with message ID
		responseChan <- models.ChatStreamEvent{
			Role:      "assistant",
			Content:   "",
			Done:      true,
			MessageID: &assistantMessage.ID,
		}
	} else {
		responseChan <- models.ChatStreamEvent{Done: true}
	}
}

// callOpenRouter makes a non-streaming call to OpenRouter.
func (s *ChatService) callOpenRouter(ctx context.Context, prompt string) (string, error) {
	const openRouterURL = "https://openrouter.ai/api/v1/chat/completions"

	requestBody := map[string]interface{}{
		"model": s.chatModel,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", openRouterURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.openRouterAPIKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call OpenRouter API: %w", err)
	}
	defer resp.Body.Close()

	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response from API")
	}

	return response.Choices[0].Message.Content, nil
}

// cleanJSONResponse removes markdown code blocks from JSON response.
func (s *ChatService) cleanJSONResponse(response string) string {
	cleaned := strings.TrimSpace(response)

	// Remove markdown code blocks
	if strings.HasPrefix(cleaned, "```json") {
		cleaned = strings.TrimPrefix(cleaned, "```json")
	} else if strings.HasPrefix(cleaned, "```") {
		cleaned = strings.TrimPrefix(cleaned, "```")
	}

	if strings.HasSuffix(cleaned, "```") {
		cleaned = strings.TrimSuffix(cleaned, "```")
	}

	cleaned = strings.TrimSpace(cleaned)

	// Extract JSON if wrapped in text
	if !strings.HasPrefix(cleaned, "{") && !strings.HasPrefix(cleaned, "[") {
		startIdx := strings.Index(cleaned, "{")
		if startIdx == -1 {
			startIdx = strings.Index(cleaned, "[")
		}
		endIdx := strings.LastIndex(cleaned, "}")
		if endIdx == -1 {
			endIdx = strings.LastIndex(cleaned, "]")
		}
		if startIdx != -1 && endIdx != -1 && endIdx > startIdx {
			cleaned = cleaned[startIdx : endIdx+1]
		}
	}

	return cleaned
}
