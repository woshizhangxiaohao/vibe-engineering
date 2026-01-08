package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

// YouTubeService handles YouTube video operations using OpenRouter and Gemini.
type YouTubeService struct {
	openRouterAPIKey string
	youtubeAPIKey    string // YouTube Data API v3 key
	geminiModel      string
	httpClient       *http.Client
	log              *zap.Logger
}

// NewYouTubeService creates a new YouTubeService.
func NewYouTubeService(apiKey string, geminiModel string, log *zap.Logger) *YouTubeService {
	// Use default model if not provided
	if geminiModel == "" {
		geminiModel = "google/gemini-3-flash-preview"
	}
	
	// Get YouTube API key from environment
	youtubeAPIKey := os.Getenv("YOUTUBE_API_KEY")
	
	return &YouTubeService{
		openRouterAPIKey: apiKey,
		youtubeAPIKey:    youtubeAPIKey,
		geminiModel:      geminiModel,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
		log: log,
	}
}

// GetVideoMetadataFromAPI fetches video metadata using YouTube Data API v3.
// This provides accurate video information (title, author, etc.)
func (s *YouTubeService) GetVideoMetadataFromAPI(ctx context.Context, videoID string) (*VideoMetadata, error) {
	if s.youtubeAPIKey == "" {
		return nil, fmt.Errorf("YOUTUBE_API_KEY not configured")
	}
	
	apiURL := fmt.Sprintf(
		"https://www.googleapis.com/youtube/v3/videos?part=snippet,contentDetails&id=%s&key=%s",
		videoID, s.youtubeAPIKey,
	)
	
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call YouTube API: %w", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	
	if resp.StatusCode != http.StatusOK {
		s.log.Error("YouTube API error",
			zap.Int("status", resp.StatusCode),
			zap.String("response", string(body)),
		)
		return nil, fmt.Errorf("YouTube API returned status %d", resp.StatusCode)
	}
	
	var apiResponse struct {
		Items []struct {
			Snippet struct {
				Title        string `json:"title"`
				ChannelTitle string `json:"channelTitle"`
				Thumbnails   struct {
					MaxRes struct {
						URL string `json:"url"`
					} `json:"maxres"`
					High struct {
						URL string `json:"url"`
					} `json:"high"`
					Default struct {
						URL string `json:"url"`
					} `json:"default"`
				} `json:"thumbnails"`
			} `json:"snippet"`
			ContentDetails struct {
				Duration string `json:"duration"` // ISO 8601 format: PT1H2M3S
			} `json:"contentDetails"`
		} `json:"items"`
	}
	
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse YouTube API response: %w", err)
	}
	
	if len(apiResponse.Items) == 0 {
		return nil, fmt.Errorf("video not found or is private")
	}
	
	item := apiResponse.Items[0]
	
	// Get best available thumbnail
	thumbnailURL := item.Snippet.Thumbnails.MaxRes.URL
	if thumbnailURL == "" {
		thumbnailURL = item.Snippet.Thumbnails.High.URL
	}
	if thumbnailURL == "" {
		thumbnailURL = item.Snippet.Thumbnails.Default.URL
	}
	if thumbnailURL == "" {
		thumbnailURL = fmt.Sprintf("https://img.youtube.com/vi/%s/maxresdefault.jpg", videoID)
	}
	
	// Parse duration (ISO 8601 format: PT1H2M3S)
	duration := parseISO8601Duration(item.ContentDetails.Duration)
	
	return &VideoMetadata{
		VideoID:      videoID,
		Title:        item.Snippet.Title,
		Author:       item.Snippet.ChannelTitle,
		ThumbnailURL: thumbnailURL,
		Duration:     duration,
	}, nil
}

// parseISO8601Duration parses ISO 8601 duration format (PT1H2M3S) to seconds.
func parseISO8601Duration(duration string) int {
	// Remove "PT" prefix
	duration = strings.TrimPrefix(duration, "PT")
	
	var hours, minutes, seconds int
	
	// Parse hours
	if idx := strings.Index(duration, "H"); idx != -1 {
		hours, _ = strconv.Atoi(duration[:idx])
		duration = duration[idx+1:]
	}
	
	// Parse minutes
	if idx := strings.Index(duration, "M"); idx != -1 {
		minutes, _ = strconv.Atoi(duration[:idx])
		duration = duration[idx+1:]
	}
	
	// Parse seconds
	if idx := strings.Index(duration, "S"); idx != -1 {
		seconds, _ = strconv.Atoi(duration[:idx])
	}
	
	return hours*3600 + minutes*60 + seconds
}

// VideoMetadata represents basic video information.
type VideoMetadata struct {
	VideoID      string
	Title        string
	Author       string
	ThumbnailURL string
	Duration     int // in seconds
}

// AnalysisResult represents the complete analysis of a video.
type AnalysisResult struct {
	Summary       string
	KeyPoints     []string
	Chapters      []ChapterData
	Transcription []TranscriptionData
}

// ChapterData represents a video chapter.
type ChapterData struct {
	Title     string
	Timestamp string
	Seconds   int
}

// TranscriptionData represents a transcription segment.
type TranscriptionData struct {
	Text      string
	Timestamp string
	Seconds   int
}

// ExtractVideoID extracts the YouTube video ID from various URL formats.
func (s *YouTubeService) ExtractVideoID(videoURL string) (string, error) {
	// Parse URL
	u, err := url.Parse(videoURL)
	if err != nil {
		return "", errors.New("invalid URL format")
	}

	// Check if it's a YouTube domain
	if !strings.Contains(u.Host, "youtube.com") && !strings.Contains(u.Host, "youtu.be") {
		return "", errors.New("not a YouTube URL")
	}

	// Handle youtu.be format
	if strings.Contains(u.Host, "youtu.be") {
		videoID := strings.TrimPrefix(u.Path, "/")
		if videoID == "" {
			return "", errors.New("invalid YouTube URL")
		}
		return videoID, nil
	}

	// Handle youtube.com format
	query := u.Query()
	videoID := query.Get("v")
	if videoID == "" {
		return "", errors.New("video ID not found in URL")
	}

	return videoID, nil
}

// GetVideoMetadata fetches basic video metadata.
// Note: This is a simplified version. In production, you would use YouTube Data API v3.
// For now, we'll use Gemini to extract metadata from the URL.
func (s *YouTubeService) GetVideoMetadata(ctx context.Context, videoURL string) (*VideoMetadata, error) {
	videoID, err := s.ExtractVideoID(videoURL)
	if err != nil {
		return nil, err
	}

	// Call Gemini via OpenRouter to get video metadata
	prompt := fmt.Sprintf(`Extract metadata from this YouTube video URL: %s

Please provide the response in JSON format with the following structure:
{
  "title": "video title",
  "author": "channel name",
  "thumbnailUrl": "thumbnail URL",
  "duration": duration_in_seconds
}`, videoURL)

	response, err := s.callGemini(ctx, prompt)
	if err != nil {
		s.log.Error("Failed to get video metadata from Gemini", zap.Error(err))
		// Return error instead of default values, so caller can distinguish between API failures and private videos
		return nil, fmt.Errorf("failed to get video metadata: %w", err)
	}

	// #region agent log
	func() {
		logFile, _ := os.OpenFile("/Users/xiaozihao/Documents/01_Projects/Work_Code/work/Team_AI/vibe-engineering-playbook/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if logFile != nil {
			defer logFile.Close()
			logData := map[string]interface{}{
				"sessionId":     "debug-session",
				"runId":         "run1",
				"hypothesisId":  "A",
				"location":      "youtube.go:127",
				"message":       "Raw Gemini response before parsing",
				"data":          map[string]interface{}{"response": response, "responseLength": len(response), "firstChars": func() string { if len(response) > 100 { return response[:100] } else { return response } }()},
				"timestamp":     time.Now().UnixMilli(),
			}
			json.NewEncoder(logFile).Encode(logData)
		}
	}()
	// #endregion

	// Clean response: remove markdown code blocks if present
	cleanedResponse := strings.TrimSpace(response)
	// Remove ```json and ``` markers (handle both ```json and ```)
	if strings.HasPrefix(cleanedResponse, "```json") {
		cleanedResponse = strings.TrimPrefix(cleanedResponse, "```json")
		cleanedResponse = strings.TrimSpace(cleanedResponse)
	} else if strings.HasPrefix(cleanedResponse, "```") {
		cleanedResponse = strings.TrimPrefix(cleanedResponse, "```")
		cleanedResponse = strings.TrimSpace(cleanedResponse)
	}
	// Remove trailing ```
	if strings.HasSuffix(cleanedResponse, "```") {
		cleanedResponse = strings.TrimSuffix(cleanedResponse, "```")
		cleanedResponse = strings.TrimSpace(cleanedResponse)
	}
	// Try to extract JSON from text if it's wrapped in explanation
	if !strings.HasPrefix(cleanedResponse, "{") {
		// Look for JSON object in the response
		startIdx := strings.Index(cleanedResponse, "{")
		endIdx := strings.LastIndex(cleanedResponse, "}")
		if startIdx != -1 && endIdx != -1 && endIdx > startIdx {
			cleanedResponse = cleanedResponse[startIdx : endIdx+1]
		}
	}

	// #region agent log
	func() {
		logFile, _ := os.OpenFile("/Users/xiaozihao/Documents/01_Projects/Work_Code/work/Team_AI/vibe-engineering-playbook/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if logFile != nil {
			defer logFile.Close()
			logData := map[string]interface{}{
				"sessionId":     "debug-session",
				"runId":         "run1",
				"hypothesisId":  "A",
				"location":      "youtube.go:145",
				"message":       "Cleaned response after removing markdown",
				"data":          map[string]interface{}{"cleanedResponse": cleanedResponse, "cleanedLength": len(cleanedResponse)},
				"timestamp":     time.Now().UnixMilli(),
			}
			json.NewEncoder(logFile).Encode(logData)
		}
	}()
	// #endregion

	// Parse response
	var metadata struct {
		Title        string `json:"title"`
		Author       string `json:"author"`
		ThumbnailURL string `json:"thumbnailUrl"`
		Duration     int    `json:"duration"`
	}

	if err := json.Unmarshal([]byte(cleanedResponse), &metadata); err != nil {
		// #region agent log
		func() {
			logFile, _ := os.OpenFile("/Users/xiaozihao/Documents/01_Projects/Work_Code/work/Team_AI/vibe-engineering-playbook/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if logFile != nil {
				defer logFile.Close()
				logData := map[string]interface{}{
					"sessionId":     "debug-session",
					"runId":         "run1",
					"hypothesisId":  "A",
					"location":      "youtube.go:165",
					"message":       "JSON parse error details",
					"data":          map[string]interface{}{"error": err.Error(), "cleanedResponse": cleanedResponse, "responseStart": func() string { if len(cleanedResponse) > 200 { return cleanedResponse[:200] } else { return cleanedResponse } }()},
					"timestamp":     time.Now().UnixMilli(),
				}
				json.NewEncoder(logFile).Encode(logData)
			}
		}()
		// #endregion
		s.log.Warn("Failed to parse metadata response", zap.Error(err), zap.String("response", cleanedResponse))
		return nil, fmt.Errorf("failed to parse metadata response: %w", err)
	}

	// 如果 title 为空，使用默认值（可能是 API 返回格式问题，不影响后续分析）
	if metadata.Title == "" {
		metadata.Title = "Video " + videoID
		s.log.Warn("Video title is empty from API response, using default", zap.String("video_id", videoID))
	}

	// 如果 thumbnailUrl 为空，使用默认的 YouTube 缩略图 URL
	if metadata.ThumbnailURL == "" {
		metadata.ThumbnailURL = fmt.Sprintf("https://img.youtube.com/vi/%s/maxresdefault.jpg", videoID)
	}

	return &VideoMetadata{
		VideoID:      videoID,
		Title:        metadata.Title,
		Author:       metadata.Author,
		ThumbnailURL: metadata.ThumbnailURL,
		Duration:     metadata.Duration,
	}, nil
}

// AnalyzeVideo performs complete video analysis using Gemini.
func (s *YouTubeService) AnalyzeVideo(ctx context.Context, videoID, targetLanguage string) (*AnalysisResult, error) {
	videoURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)

	prompt := fmt.Sprintf(`分析这个 YouTube 视频并返回所有字幕: %s

请返回 JSON 格式的响应，包含完整的字幕内容：
{
  "transcription": [
    {"text": "字幕文本", "timestamp": "00:00", "seconds": 0},
    {"text": "更多字幕文本", "timestamp": "00:15", "seconds": 15}
  ]
}

重要提示:
- 返回视频的所有字幕内容
- 时间戳格式为 MM:SS
- 将时间戳转换为秒数
- 字幕文本使用原始语言`, videoURL)

	response, err := s.callGemini(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze video: %w", err)
	}

	// Parse the response - 只解析字幕内容
	var result struct {
		Transcription []struct {
			Text      string `json:"text"`
			Timestamp string `json:"timestamp"`
			Seconds   int    `json:"seconds"`
		} `json:"transcription"`
	}

	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, fmt.Errorf("failed to parse analysis response: %w", err)
	}

	// Convert to our domain types - 只返回字幕，其他字段为空
	analysisResult := &AnalysisResult{
		Summary:       "",
		KeyPoints:     []string{},
		Chapters:      []ChapterData{},
		Transcription: make([]TranscriptionData, len(result.Transcription)),
	}

	for i, tr := range result.Transcription {
		analysisResult.Transcription[i] = TranscriptionData{
			Text:      tr.Text,
			Timestamp: tr.Timestamp,
			Seconds:   tr.Seconds,
		}
	}

	return analysisResult, nil
}

<<<<<<< HEAD
// FetchYouTubeTranscript attempts to fetch real transcript from YouTube.
// Returns the transcript text or an error if not available.
func (s *YouTubeService) FetchYouTubeTranscript(ctx context.Context, videoID string) (string, error) {
	// First, get the video page to extract caption track info
	videoPageURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)
	
	req, err := http.NewRequestWithContext(ctx, "GET", videoPageURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	
	// Set headers to mimic a real browser
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7")
	req.Header.Set("Accept-Encoding", "identity") // Don't compress to make parsing easier
	
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch video page: %w", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read video page: %w", err)
	}
	
	pageContent := string(body)
	
	s.log.Debug("Fetched YouTube page",
		zap.String("video_id", videoID),
		zap.Int("page_length", len(pageContent)),
	)
	
	// Method 1: Try to find captionTracks in ytInitialPlayerResponse
	// This contains the actual subtitle tracks (not auto-translate)
	captionTracksPattern := regexp.MustCompile(`"captionTracks"\s*:\s*\[([^\]]+)\]`)
	captionTracksMatches := captionTracksPattern.FindStringSubmatch(pageContent)
	
	var captionURL string
	var captionLang string
	
	if captionTracksMatches != nil && len(captionTracksMatches) >= 2 {
		// Found captionTracks, extract the first baseUrl
		trackContent := captionTracksMatches[1]
		
		// Extract baseUrl
		baseURLPattern := regexp.MustCompile(`"baseUrl"\s*:\s*"([^"]+)"`)
		urlMatches := baseURLPattern.FindStringSubmatch(trackContent)
		if urlMatches != nil && len(urlMatches) >= 2 {
			captionURL = urlMatches[1]
		}
		
		// Extract language code of the caption
		langPattern := regexp.MustCompile(`"languageCode"\s*:\s*"([^"]+)"`)
		langMatches := langPattern.FindStringSubmatch(trackContent)
		if langMatches != nil && len(langMatches) >= 2 {
			captionLang = langMatches[1]
		}
		
		s.log.Info("Found caption track",
			zap.String("video_id", videoID),
			zap.String("lang", captionLang),
			zap.Bool("has_url", captionURL != ""),
		)
	}
	
	// Method 2: Try to find timedtext URL directly if captionTracks not found
	if captionURL == "" {
		timedtextPattern := regexp.MustCompile(`(https://www\.youtube\.com/api/timedtext[^"\\]+)`)
		timedtextMatches := timedtextPattern.FindStringSubmatch(pageContent)
		if timedtextMatches != nil && len(timedtextMatches) >= 2 {
			captionURL = timedtextMatches[1]
		}
	}
	
	// Method 3: Try ytInitialPlayerResponse
	if captionURL == "" {
		playerResponsePattern := regexp.MustCompile(`ytInitialPlayerResponse\s*=\s*(\{.+?\});`)
		playerMatches := playerResponsePattern.FindStringSubmatch(pageContent)
		if playerMatches != nil && len(playerMatches) >= 2 {
			captionURL = s.extractCaptionURLFromPlayerResponse(playerMatches[1])
		}
	}
	
	if captionURL == "" {
		s.log.Warn("No captions found in video page",
			zap.String("video_id", videoID),
		)
		return "", fmt.Errorf("NO_CAPTIONS: 此视频没有可用的字幕（视频可能未开启字幕功能）")
	}
	
	// Decode the URL (it's JSON escaped)
	captionURL = strings.ReplaceAll(captionURL, "\\u0026", "&")
	captionURL = strings.ReplaceAll(captionURL, "\\/", "/")
	
	// Add format parameter for XML output if not present
	if !strings.Contains(captionURL, "&fmt=") {
		captionURL = captionURL + "&fmt=srv3"
	}
	
	s.log.Info("Found caption URL",
		zap.String("video_id", videoID),
		zap.String("caption_url", captionURL[:min(150, len(captionURL))]),
		zap.String("caption_lang", captionLang),
	)
	
	// Define languages to try
	languages := []string{}
	seen := make(map[string]bool)
	
	// If we detected a caption language, use it first
	if captionLang != "" {
		languages = append(languages, captionLang)
		seen[captionLang] = true
	}
	
	// Try to fetch captions
	var captionText string
	var lastErr error
	
	// First, try the URL as-is (it should have the correct language already)
	s.log.Debug("Trying caption URL directly",
		zap.String("video_id", videoID),
		zap.String("url", captionURL[:min(100, len(captionURL))]),
	)
	
	captionReq, err := http.NewRequestWithContext(ctx, "GET", captionURL, nil)
	if err == nil {
		captionReq.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")
		
		captionResp, err := s.httpClient.Do(captionReq)
		if err == nil {
			captionBody, err := io.ReadAll(captionResp.Body)
			captionResp.Body.Close()
			if err == nil && len(captionBody) > 0 {
				s.log.Debug("Fetched caption content",
					zap.String("video_id", videoID),
					zap.Int("caption_length", len(captionBody)),
					zap.String("first_200_chars", string(captionBody[:min(200, len(captionBody))])),
				)
				
				captionText = s.parseYouTubeCaptions(string(captionBody))
				if captionText != "" {
					s.log.Info("Successfully parsed captions directly",
						zap.String("video_id", videoID),
						zap.Int("text_length", len(captionText)),
					)
					return captionText, nil
				}
			}
		} else {
			lastErr = err
		}
	}
	
	// If direct URL didn't work, try with different language parameters
	priorityLangs := []string{"zh-TW", "zh-Hant", "zh-Hans", "zh", "en", "ja", "ko"}
	langsToTry := []string{}
	
	// Add detected language first
	if captionLang != "" && !seen[captionLang] {
		langsToTry = append(langsToTry, captionLang)
		seen[captionLang] = true
	}
	
	// Add priority languages
	for _, lang := range priorityLangs {
		if !seen[lang] {
			langsToTry = append(langsToTry, lang)
			seen[lang] = true
		}
	}
	
	// Get base URL without lang parameter
	baseCapURL := captionURL
	if idx := strings.Index(baseCapURL, "&lang="); idx != -1 {
		endIdx := strings.Index(baseCapURL[idx+1:], "&")
		if endIdx != -1 {
			baseCapURL = baseCapURL[:idx] + baseCapURL[idx+1+endIdx:]
		} else {
			baseCapURL = baseCapURL[:idx]
		}
	}
	
	for _, lang := range langsToTry {
		tryURL := baseCapURL + "&lang=" + lang
		
		s.log.Debug("Trying caption URL with lang",
			zap.String("video_id", videoID),
			zap.String("lang", lang),
		)
		
		captionReq, err := http.NewRequestWithContext(ctx, "GET", tryURL, nil)
		if err != nil {
			lastErr = err
			continue
		}
		captionReq.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")
		
		captionResp, err := s.httpClient.Do(captionReq)
		if err != nil {
			lastErr = err
			continue
		}
		
		captionBody, err := io.ReadAll(captionResp.Body)
		captionResp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}
		
		if len(captionBody) == 0 {
			continue
		}
		
		s.log.Debug("Fetched caption content",
			zap.String("video_id", videoID),
			zap.String("lang", lang),
			zap.Int("caption_length", len(captionBody)),
		)
		
		// Parse the XML caption content
		captionText = s.parseYouTubeCaptions(string(captionBody))
		if captionText != "" {
			s.log.Info("Successfully parsed captions",
				zap.String("video_id", videoID),
				zap.String("lang", lang),
				zap.Int("text_length", len(captionText)),
			)
			return captionText, nil
		}
	}
	
	if lastErr != nil {
		return "", fmt.Errorf("NO_CAPTIONS: 字幕获取失败: %w", lastErr)
	}
	return "", fmt.Errorf("NO_CAPTIONS: 字幕内容解析失败（尝试了 %d 种语言）", len(langsToTry)+1)
}

// extractCaptionURLFromPlayerResponse extracts caption URL from ytInitialPlayerResponse JSON.
func (s *YouTubeService) extractCaptionURLFromPlayerResponse(jsonStr string) string {
	// Try to find captionTracks in the JSON
	baseURLPattern := regexp.MustCompile(`"baseUrl"\s*:\s*"(https[^"]+timedtext[^"]+)"`)
	matches := baseURLPattern.FindStringSubmatch(jsonStr)
	if matches != nil && len(matches) >= 2 {
		return matches[1]
	}
	return ""
}

// min returns the smaller of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// parseYouTubeCaptions parses YouTube's XML caption format into readable text.
func (s *YouTubeService) parseYouTubeCaptions(xmlContent string) string {
	// YouTube captions can be in different formats:
	// 1. XML format: <text start="0" dur="5.5">Caption text</text>
	// 2. JSON format (sometimes returned)
	
	// First, try XML format
	textPattern := regexp.MustCompile(`<text[^>]*start="([^"]*)"[^>]*>([\s\S]*?)</text>`)
	matches := textPattern.FindAllStringSubmatch(xmlContent, -1)
	
	if len(matches) == 0 {
		// Try alternative XML patterns
		altPattern := regexp.MustCompile(`<p[^>]*t="(\d+)"[^>]*>([\s\S]*?)</p>`)
		matches = altPattern.FindAllStringSubmatch(xmlContent, -1)
	}
	
	if len(matches) == 0 {
		s.log.Debug("Failed to parse caption XML",
			zap.String("content_preview", xmlContent[:min(500, len(xmlContent))]),
		)
		return ""
	}
	
	var result strings.Builder
	for _, match := range matches {
		if len(match) >= 3 {
			timestamp := match[1]
			text := match[2]
			
			// Decode HTML entities
			text = decodeHTMLEntities(text)
			
			// Clean up whitespace
			text = strings.TrimSpace(text)
			text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
			
			if text == "" {
				continue
			}
			
			// Convert timestamp to MM:SS format
			seconds, _ := strconv.ParseFloat(timestamp, 64)
			// Handle milliseconds format (t="1234" means 1.234 seconds in some formats)
			if seconds > 10000 {
				seconds = seconds / 1000
			}
			mins := int(seconds) / 60
			secs := int(seconds) % 60
			
			result.WriteString(fmt.Sprintf("[%02d:%02d] %s\n", mins, secs, text))
		}
	}
	
	return result.String()
}

// decodeHTMLEntities decodes common HTML entities in text.
func decodeHTMLEntities(text string) string {
	replacements := map[string]string{
		"&amp;":   "&",
		"&lt;":    "<",
		"&gt;":    ">",
		"&quot;":  "\"",
		"&#39;":   "'",
		"&apos;":  "'",
		"&nbsp;":  " ",
		"&#x27;":  "'",
		"&#x2F;":  "/",
		"&#34;":   "\"",
		"&#60;":   "<",
		"&#62;":   ">",
		"&lrm;":   "",
		"&rlm;":   "",
		"\n":      " ",
		"\r":      "",
	}
	
	for entity, replacement := range replacements {
		text = strings.ReplaceAll(text, entity, replacement)
	}
	
	// Handle numeric entities like &#123;
	numericPattern := regexp.MustCompile(`&#(\d+);`)
	text = numericPattern.ReplaceAllStringFunc(text, func(match string) string {
		numStr := match[2 : len(match)-1]
		if num, err := strconv.Atoi(numStr); err == nil && num < 128 {
			return string(rune(num))
		}
		return match
	})
	
	return text
}

// CallGeminiDirect directly calls Gemini with a video URL and returns the raw response.
// It only returns real YouTube captions - no LLM hallucination allowed.
func (s *YouTubeService) CallGeminiDirect(ctx context.Context, videoURL string) (string, error) {
	// First, try to extract video ID
	videoID, err := s.ExtractVideoID(videoURL)
	if err != nil {
		return "", fmt.Errorf("invalid YouTube URL: %w", err)
	}
	
	// Try to fetch real YouTube transcript
	transcript, err := s.FetchYouTubeTranscript(ctx, videoID)
	if err == nil && transcript != "" {
		s.log.Info("Successfully fetched real YouTube transcript",
			zap.String("video_id", videoID),
			zap.Int("transcript_length", len(transcript)),
		)
		return fmt.Sprintf("# 视频字幕 (来自 YouTube 官方字幕)\n\n%s", transcript), nil
	}
	
	// Log why we couldn't get real transcript
	s.log.Warn("Could not fetch real YouTube transcript",
		zap.String("video_id", videoID),
		zap.Error(err),
	)
	
	// Return honest message - DO NOT let LLM hallucinate
	errorMsg := fmt.Sprintf(`## ⚠️ 无法获取视频字幕

**视频 ID:** %s
**视频链接:** %s

**可能的原因:**
- 该视频没有上传字幕
- 该视频禁用了字幕功能
- YouTube API 限制

**建议:**
- 请确认该视频是否有字幕（在 YouTube 播放器中点击 CC 按钮查看）
- 尝试其他有字幕的视频`, videoID, videoURL)
	
	return errorMsg, nil
}

// GetMetadataWithAI directly analyzes YouTube URL using Gemini to extract all metadata and analysis.
// This bypasses traditional metadata fetching and works even for private/restricted videos.
// NOTE: This function may cause LLM hallucination. Use GetVideoMetadata instead for accurate data.
func (s *YouTubeService) GetMetadataWithAI(ctx context.Context, videoURL string) (*VideoMetadataWithAI, error) {
	videoID, err := s.ExtractVideoID(videoURL)
	if err != nil {
		return nil, err
	}

	prompt := fmt.Sprintf(`Analyze this YouTube video URL: %s

Please extract and provide comprehensive information about this video in JSON format:
{
  "title": "video title",
  "author": "channel/creator name",
  "description": "brief description of the video content",
  "aiAnalysis": "your AI analysis and summary of the video content",
  "metadata": {
    "duration": duration_in_seconds,
    "thumbnail": "thumbnail URL (use format: https://img.youtube.com/vi/VIDEO_ID/maxresdefault.jpg)"
  }
}

IMPORTANT:
- If you cannot access the actual video content, make reasonable inferences from the URL and any available public information
- Always provide all fields in the JSON response
- aiAnalysis should be a comprehensive summary of what the video is about
- duration should be a number (in seconds), use 0 if unknown
- For thumbnail, use the standard YouTube thumbnail URL format`, videoURL)

	response, err := s.callGemini(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("AI service error: %w", err)
	}

	// Parse response - try to extract JSON from the response
	var result VideoMetadataWithAI

	// Clean up response - sometimes Gemini wraps JSON in markdown code blocks
	cleanResponse := strings.TrimSpace(response)
	if strings.HasPrefix(cleanResponse, "```json") {
		cleanResponse = strings.TrimPrefix(cleanResponse, "```json")
		cleanResponse = strings.TrimSuffix(cleanResponse, "```")
		cleanResponse = strings.TrimSpace(cleanResponse)
	} else if strings.HasPrefix(cleanResponse, "```") {
		cleanResponse = strings.TrimPrefix(cleanResponse, "```")
		cleanResponse = strings.TrimSuffix(cleanResponse, "```")
		cleanResponse = strings.TrimSpace(cleanResponse)
	}

	if err := json.Unmarshal([]byte(cleanResponse), &result); err != nil {
		s.log.Error("Failed to parse AI metadata response",
			zap.Error(err),
			zap.String("response", cleanResponse),
		)
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	// Ensure we have the video ID
	result.VideoID = videoID

	// Set default thumbnail if not provided
	if result.Metadata.Thumbnail == "" {
		result.Metadata.Thumbnail = fmt.Sprintf("https://img.youtube.com/vi/%s/maxresdefault.jpg", videoID)
	}

	return &result, nil
}

// VideoMetadataWithAI represents comprehensive video metadata with AI analysis.
type VideoMetadataWithAI struct {
	VideoID     string `json:"videoId"`
	Title       string `json:"title"`
	Author      string `json:"author"`
	Description string `json:"description"`
	AIAnalysis  string `json:"aiAnalysis"`
	Metadata    struct {
		Duration  int    `json:"duration"`
		Thumbnail string `json:"thumbnail"`
	} `json:"metadata"`
}

// callGemini calls the Gemini API via OpenRouter.
func (s *YouTubeService) callGemini(ctx context.Context, prompt string) (string, error) {
	// Check if API key is configured
	if s.openRouterAPIKey == "" {
		return "", fmt.Errorf("OPENROUTER_API_KEY environment variable is not set. Please configure it to use video analysis features")
	}

	const openRouterURL = "https://openrouter.ai/api/v1/chat/completions"

	requestBody := map[string]interface{}{
		"model": s.geminiModel,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		s.log.Error("OpenRouter API error",
			zap.Int("status", resp.StatusCode),
			zap.String("response", string(body)),
		)
		return "", fmt.Errorf("OpenRouter API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var apiResponse struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return "", fmt.Errorf("failed to parse API response: %w", err)
	}

	if len(apiResponse.Choices) == 0 {
		return "", errors.New("no response from Gemini")
	}

	return apiResponse.Choices[0].Message.Content, nil
}

// TimestampToSeconds converts a timestamp string (MM:SS) to seconds.
func TimestampToSeconds(timestamp string) (int, error) {
	re := regexp.MustCompile(`^(\d+):(\d{2})$`)
	matches := re.FindStringSubmatch(timestamp)
	if matches == nil {
		return 0, errors.New("invalid timestamp format")
	}

	minutes, _ := strconv.Atoi(matches[1])
	seconds, _ := strconv.Atoi(matches[2])

	return minutes*60 + seconds, nil
}

// SecondsToTimestamp converts seconds to timestamp string (MM:SS).
func SecondsToTimestamp(seconds int) string {
	minutes := seconds / 60
	secs := seconds % 60
	return fmt.Sprintf("%02d:%02d", minutes, secs)
}
