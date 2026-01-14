package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"vibe-backend/internal/models"
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
				"sessionId":    "debug-session",
				"runId":        "run1",
				"hypothesisId": "A",
				"location":     "youtube.go:127",
				"message":      "Raw Gemini response before parsing",
				"data": map[string]interface{}{"response": response, "responseLength": len(response), "firstChars": func() string {
					if len(response) > 100 {
						return response[:100]
					} else {
						return response
					}
				}()},
				"timestamp": time.Now().UnixMilli(),
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
				"sessionId":    "debug-session",
				"runId":        "run1",
				"hypothesisId": "A",
				"location":     "youtube.go:145",
				"message":      "Cleaned response after removing markdown",
				"data":         map[string]interface{}{"cleanedResponse": cleanedResponse, "cleanedLength": len(cleanedResponse)},
				"timestamp":    time.Now().UnixMilli(),
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
					"sessionId":    "debug-session",
					"runId":        "run1",
					"hypothesisId": "A",
					"location":     "youtube.go:165",
					"message":      "JSON parse error details",
					"data": map[string]interface{}{"error": err.Error(), "cleanedResponse": cleanedResponse, "responseStart": func() string {
						if len(cleanedResponse) > 200 {
							return cleanedResponse[:200]
						} else {
							return cleanedResponse
						}
					}()},
					"timestamp": time.Now().UnixMilli(),
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

// FetchYouTubeTranscriptStructured fetches structured transcript data from YouTube.
func (s *YouTubeService) FetchYouTubeTranscriptStructured(ctx context.Context, videoID string) (*models.YouTubeTranscriptResponse, error) {
	// Get video metadata first
	videoMetadata, err := s.GetVideoMetadataFromAPI(ctx, videoID)
	if err != nil {
		s.log.Warn("Failed to get video metadata, using defaults",
			zap.String("video_id", videoID),
			zap.Error(err),
		)
		videoMetadata = &VideoMetadata{
			VideoID:      videoID,
			Title:        "",
			Author:       "",
			ThumbnailURL: fmt.Sprintf("https://i.ytimg.com/vi/%s/maxresdefault.jpg", videoID),
			Duration:     0,
		}
	}

	// Get caption tracks and transcripts via innertube API
	apiURL := "https://www.youtube.com/youtubei/v1/player"
	apiKey := "AIzaSyAO_FJ2SlqU8Q4STEHLGCilw_Y9_11qcW8"

	payload := map[string]interface{}{
		"context": map[string]interface{}{
			"client": map[string]interface{}{
				"hl":            "en",
				"gl":            "US",
				"clientName":    "WEB",
				"clientVersion": "2.20231219.04.00",
			},
		},
		"videoId": videoID,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal innertube request: %w", err)
	}

	reqURL := fmt.Sprintf("%s?key=%s", apiURL, apiKey)
	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to create innertube request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("innertube API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read innertube response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("innertube API returned status %d", resp.StatusCode)
	}

	// Parse response to get caption tracks and video details
	var apiResult struct {
		VideoDetails struct {
			Title       string `json:"title"`
			ChannelId   string `json:"channelId"`
			LengthSeconds string `json:"lengthSeconds"`
		} `json:"videoDetails"`
		Captions struct {
			PlayerCaptionsTracklistRenderer struct {
				CaptionTracks []struct {
					BaseURL      string `json:"baseUrl"`
					LanguageCode string `json:"languageCode"`
					Name         struct {
						SimpleText string `json:"simpleText"`
					} `json:"name"`
					Kind string `json:"kind"` // "asr" for auto-generated, "" for manual
				} `json:"captionTracks"`
			} `json:"playerCaptionsTracklistRenderer"`
		} `json:"captions"`
	}

	if err := json.Unmarshal(body, &apiResult); err != nil {
		return nil, fmt.Errorf("failed to parse innertube response: %w", err)
	}

	captionTracks := apiResult.Captions.PlayerCaptionsTracklistRenderer.CaptionTracks
	if len(captionTracks) == 0 {
		return nil, fmt.Errorf("NO_CAPTIONS: 此视频没有可用的字幕")
	}

	// Build language codes list
	languageCodes := make([]models.YouTubeLanguageCode, len(captionTracks))
	for i, track := range captionTracks {
		languageCodes[i] = models.YouTubeLanguageCode{
			Code: track.LanguageCode,
			Name: track.Name.SimpleText,
		}
	}

	// Fetch transcripts for all available languages
	transcripts := make(map[string]models.TranscriptData)

	for _, track := range captionTracks {
		langCode := track.LanguageCode
		isAuto := track.Kind == "asr"

		// Fetch caption content
		captionReq, err := http.NewRequestWithContext(ctx, "GET", track.BaseURL, nil)
		if err != nil {
			s.log.Warn("Failed to create caption request",
				zap.String("lang", langCode),
				zap.Error(err),
			)
			continue
		}
		captionReq.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")

		captionResp, err := s.httpClient.Do(captionReq)
		if err != nil {
			s.log.Warn("Failed to fetch caption content",
				zap.String("lang", langCode),
				zap.Error(err),
			)
			continue
		}

		captionBody, err := io.ReadAll(captionResp.Body)
		captionResp.Body.Close()
		if err != nil || len(captionBody) == 0 {
			s.log.Warn("Failed to read caption content",
				zap.String("lang", langCode),
				zap.Error(err),
			)
			continue
		}

		// Parse structured segments
		segments, err := s.parseYouTubeCaptionsStructured(string(captionBody))
		if err != nil || len(segments) == 0 {
			s.log.Warn("Structured parsing failed, trying fallback text parsing",
				zap.String("lang", langCode),
				zap.Error(err),
				zap.Int("contentLength", len(captionBody)),
			)
			
			// Fallback: Try to parse as plain text and convert to segments
			plainText := s.parseYouTubeCaptions(string(captionBody))
			if plainText != "" {
				// Convert plain text to segments
				segments = s.convertPlainTextToSegments(plainText)
				if len(segments) > 0 {
					s.log.Info("Successfully parsed using fallback text method",
						zap.String("lang", langCode),
						zap.Int("segments", len(segments)),
					)
				} else {
					s.log.Warn("Fallback text parsing also failed",
						zap.String("lang", langCode),
					)
					continue
				}
			} else {
				s.log.Warn("Both structured and text parsing failed",
					zap.String("lang", langCode),
				)
				continue
			}
		}

		// Initialize transcript data for this language
		transcriptData := models.TranscriptData{
			Custom:  []models.TranscriptSegment{},
			Default: []models.TranscriptSegment{},
			Auto:    []models.TranscriptSegment{},
		}

		if isAuto {
			transcriptData.Auto = segments
		} else {
			transcriptData.Default = segments
			// Generate merged custom segments (combine multiple segments into longer ones)
			transcriptData.Custom = mergeTranscriptSegments(segments, 20) // Merge into ~20 second chunks
		}

		transcripts[langCode] = transcriptData
	}

	// Check if we got at least one successful transcript
	if len(transcripts) == 0 {
		return nil, fmt.Errorf("NO_CAPTIONS: 无法解析任何字幕内容")
	}

	// Build response
	response := &models.YouTubeTranscriptResponse{
		VideoID:      videoID,
		LanguageCode: languageCodes,
		Transcripts: transcripts,
		VideoInfo: models.YouTubeVideoInfo{
			Name:        videoMetadata.Title,
			ThumbnailURL: models.YouTubeThumbnailURLs{
				Hqdefault:     fmt.Sprintf("https://i.ytimg.com/vi/%s/hqdefault.jpg", videoID),
				Maxresdefault: fmt.Sprintf("https://i.ytimg.com/vi/%s/maxresdefault.jpg", videoID),
			},
			EmbedURL:    fmt.Sprintf("https://www.youtube.com/embed/%s", videoID),
			Duration:    fmt.Sprintf("%d", videoMetadata.Duration),
			Description: "",
			UploadDate:  "",
			Genre:       "",
			Author:      videoMetadata.Author,
			ChannelID:   apiResult.VideoDetails.ChannelId,
		},
	}

	// Update title from API if available
	if apiResult.VideoDetails.Title != "" {
		response.VideoInfo.Name = apiResult.VideoDetails.Title
	}
	if apiResult.VideoDetails.LengthSeconds != "" {
		if duration, err := strconv.Atoi(apiResult.VideoDetails.LengthSeconds); err == nil {
			response.VideoInfo.Duration = fmt.Sprintf("%d", duration)
		}
	}

	return response, nil
}

// FetchYouTubeTranscript attempts to fetch real transcript from YouTube.
// Returns the transcript text or an error if not available.
func (s *YouTubeService) FetchYouTubeTranscript(ctx context.Context, videoID string) (string, error) {
	// #region agent log
	logDebugYoutube("youtube.go:446", "FetchYouTubeTranscript entry", map[string]interface{}{
		"videoId": videoID,
	}, "C")
	// #endregion

	// Method 1: Try YouTube innertube API first (most reliable)
	transcript, err := s.fetchTranscriptViaInnertubeAPI(ctx, videoID)
	if err == nil && transcript != "" {
		s.log.Info("Successfully fetched transcript via innertube API",
			zap.String("video_id", videoID),
			zap.Int("length", len(transcript)),
		)
		return transcript, nil
	}
	if err != nil {
		s.log.Debug("Innertube API failed, trying web scraping fallback",
			zap.String("video_id", videoID),
			zap.Error(err),
		)
	}

	// Method 2: Fallback to web scraping
	// First, get the video page to extract caption track info
	videoPageURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)

	req, err := http.NewRequestWithContext(ctx, "GET", videoPageURL, nil)
	if err != nil {
		// #region agent log
		logDebugYoutube("youtube.go:450", "Failed to create request", map[string]interface{}{
			"videoId": videoID,
			"error": err.Error(),
		}, "C")
		// #endregion
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
		// #region agent log
		logDebugYoutube("youtube.go:530", "No caption URL found", map[string]interface{}{
			"videoId": videoID,
			"pageLength": len(pageContent),
		}, "C")
		// #endregion
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
				// #region agent log
				logDebugYoutube("youtube.go:580", "Fetched caption content directly", map[string]interface{}{
					"videoId": videoID,
					"captionLength": len(captionBody),
					"statusCode": captionResp.StatusCode,
					"contentType": captionResp.Header.Get("Content-Type"),
					"first500Chars": string(captionBody[:min(500, len(captionBody))]),
				}, "C")
				// #endregion

				s.log.Debug("Fetched caption content",
					zap.String("video_id", videoID),
					zap.Int("caption_length", len(captionBody)),
					zap.String("first_200_chars", string(captionBody[:min(200, len(captionBody))])),
					zap.Int("status_code", captionResp.StatusCode),
					zap.String("content_type", captionResp.Header.Get("Content-Type")),
				)

				captionText = s.parseYouTubeCaptions(string(captionBody))
				
				// #region agent log
				logDebugYoutube("youtube.go:605", "Parsed caption result", map[string]interface{}{
					"videoId": videoID,
					"parsedTextLength": len(captionText),
					"parsedTextEmpty": captionText == "",
					"parsedTextPreview": func() string {
						if len(captionText) > 200 {
							return captionText[:200]
						}
						return captionText
					}(),
				}, "C")
				// #endregion

				if captionText != "" {
					s.log.Info("Successfully parsed captions directly",
						zap.String("video_id", videoID),
						zap.Int("text_length", len(captionText)),
					)
					return captionText, nil
				}
			} else {
				// #region agent log
				logDebugYoutube("youtube.go:576", "Failed to fetch caption body", map[string]interface{}{
					"videoId": videoID,
					"hasError": err != nil,
					"error": func() string {
						if err != nil {
							return err.Error()
						}
						return ""
					}(),
					"bodyLength": len(captionBody),
				}, "C")
				// #endregion
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
			// #region agent log
			logDebugYoutube("youtube.go:658", "Empty caption body", map[string]interface{}{
				"videoId": videoID,
				"lang": lang,
			}, "C")
			// #endregion
			continue
		}

		// #region agent log
		logDebugYoutube("youtube.go:662", "Fetched caption content with lang", map[string]interface{}{
			"videoId": videoID,
			"lang": lang,
			"captionLength": len(captionBody),
			"statusCode": captionResp.StatusCode,
			"contentType": captionResp.Header.Get("Content-Type"),
			"first500Chars": string(captionBody[:min(500, len(captionBody))]),
		}, "C")
		// #endregion

		s.log.Debug("Fetched caption content",
			zap.String("video_id", videoID),
			zap.String("lang", lang),
			zap.Int("caption_length", len(captionBody)),
			zap.Int("status_code", captionResp.StatusCode),
			zap.String("content_type", captionResp.Header.Get("Content-Type")),
		)

		// Parse the XML caption content
		captionText = s.parseYouTubeCaptions(string(captionBody))
		
		// #region agent log
		logDebugYoutube("youtube.go:687", "Parsed caption result with lang", map[string]interface{}{
			"videoId": videoID,
			"lang": lang,
			"parsedTextLength": len(captionText),
			"parsedTextEmpty": captionText == "",
			"parsedTextPreview": func() string {
				if len(captionText) > 200 {
					return captionText[:200]
				}
				return captionText
			}(),
		}, "C")
		// #endregion

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

// fetchTranscriptViaInnertubeAPI uses YouTube's innertube API to fetch captions.
// This is more reliable than web scraping as it uses YouTube's official internal API.
func (s *YouTubeService) fetchTranscriptViaInnertubeAPI(ctx context.Context, videoID string) (string, error) {
	// YouTube innertube API endpoint
	apiURL := "https://www.youtube.com/youtubei/v1/player"
	apiKey := "AIzaSyAO_FJ2SlqU8Q4STEHLGCilw_Y9_11qcW8" // Public API key used by YouTube web client

	// Build request payload
	payload := map[string]interface{}{
		"context": map[string]interface{}{
			"client": map[string]interface{}{
				"hl":            "en",
				"gl":            "US",
				"clientName":    "WEB",
				"clientVersion": "2.20231219.04.00",
			},
		},
		"videoId": videoID,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal innertube request: %w", err)
	}

	reqURL := fmt.Sprintf("%s?key=%s", apiURL, apiKey)
	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", fmt.Errorf("failed to create innertube request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("innertube API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read innertube response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("innertube API returned status %d", resp.StatusCode)
	}

	// Parse response to get caption tracks
	var result struct {
		Captions struct {
			PlayerCaptionsTracklistRenderer struct {
				CaptionTracks []struct {
					BaseURL      string `json:"baseUrl"`
					LanguageCode string `json:"languageCode"`
					Name         struct {
						SimpleText string `json:"simpleText"`
					} `json:"name"`
				} `json:"captionTracks"`
			} `json:"playerCaptionsTracklistRenderer"`
		} `json:"captions"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse innertube response: %w", err)
	}

	captionTracks := result.Captions.PlayerCaptionsTracklistRenderer.CaptionTracks
	if len(captionTracks) == 0 {
		return "", fmt.Errorf("NO_CAPTIONS: 此视频没有可用的字幕")
	}

	s.log.Info("Found caption tracks via innertube API",
		zap.String("video_id", videoID),
		zap.Int("track_count", len(captionTracks)),
	)

	// Try to find Chinese caption first, then English, then any
	var selectedURL string
	var selectedLang string

	priorityLangs := []string{"zh", "zh-Hans", "zh-Hant", "zh-TW", "zh-CN", "en"}
	for _, lang := range priorityLangs {
		for _, track := range captionTracks {
			if strings.HasPrefix(track.LanguageCode, lang) || track.LanguageCode == lang {
				selectedURL = track.BaseURL
				selectedLang = track.LanguageCode
				break
			}
		}
		if selectedURL != "" {
			break
		}
	}

	// If no preferred language found, use first available
	if selectedURL == "" && len(captionTracks) > 0 {
		selectedURL = captionTracks[0].BaseURL
		selectedLang = captionTracks[0].LanguageCode
	}

	if selectedURL == "" {
		return "", fmt.Errorf("NO_CAPTIONS: 无法获取字幕 URL")
	}

	s.log.Info("Fetching captions from innertube URL",
		zap.String("video_id", videoID),
		zap.String("lang", selectedLang),
	)

	// Fetch caption content
	captionReq, err := http.NewRequestWithContext(ctx, "GET", selectedURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create caption request: %w", err)
	}
	captionReq.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")

	captionResp, err := s.httpClient.Do(captionReq)
	if err != nil {
		return "", fmt.Errorf("failed to fetch captions: %w", err)
	}
	defer captionResp.Body.Close()

	captionBody, err := io.ReadAll(captionResp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read caption content: %w", err)
	}

	if len(captionBody) == 0 {
		return "", fmt.Errorf("NO_CAPTIONS: 字幕内容为空")
	}

	// Parse the caption XML
	captionText := s.parseYouTubeCaptions(string(captionBody))
	if captionText == "" {
		return "", fmt.Errorf("NO_CAPTIONS: 字幕解析失败")
	}

	return captionText, nil
}

// parseYouTubeCaptionsStructured parses YouTube's XML caption format into structured segments.
func (s *YouTubeService) parseYouTubeCaptionsStructured(xmlContent string) ([]models.TranscriptSegment, error) {
	var segments []models.TranscriptSegment

	// First, try XML format with start and dur attributes
	textPattern := regexp.MustCompile(`<text[^>]*start="([^"]*)"[^>]*dur="([^"]*)"[^>]*>([\s\S]*?)</text>`)
	matches := textPattern.FindAllStringSubmatch(xmlContent, -1)

	if len(matches) == 0 {
		// Try pattern with only start attribute (need to calculate dur from next segment)
		textPattern = regexp.MustCompile(`<text[^>]*start="([^"]*)"[^>]*>([\s\S]*?)</text>`)
		matches = textPattern.FindAllStringSubmatch(xmlContent, -1)
	}

	if len(matches) == 0 {
		// Try alternative XML patterns with p tag
		altPattern := regexp.MustCompile(`<p[^>]*t="(\d+)"[^>]*d="(\d+)"[^>]*>([\s\S]*?)</p>`)
		matches = altPattern.FindAllStringSubmatch(xmlContent, -1)
	}

	if len(matches) == 0 {
		// Try pattern with t attribute only (milliseconds)
		altPattern := regexp.MustCompile(`<p[^>]*t="(\d+)"[^>]*>([\s\S]*?)</p>`)
		matches = altPattern.FindAllStringSubmatch(xmlContent, -1)
	}

	if len(matches) == 0 {
		s.log.Warn("All parsing patterns failed",
			zap.String("contentSample", xmlContent[:min(500, len(xmlContent))]),
		)
		return nil, fmt.Errorf("无法解析字幕格式")
	}

	// Parse all segments first to calculate durations
	type rawSegment struct {
		startSeconds float64
		durSeconds   float64
		text         string
		hasDur       bool
	}

	rawSegments := make([]rawSegment, 0, len(matches))
	for _, match := range matches {
		var startSeconds float64
		var durSeconds float64
		var text string
		hasDur := false

		if len(match) >= 4 && match[2] != "" {
			// Pattern with start and dur
			var err error
			startSeconds, err = strconv.ParseFloat(match[1], 64)
			if err != nil {
				continue
			}
			durSeconds, err = strconv.ParseFloat(match[2], 64)
			if err == nil {
				hasDur = true
			} else {
				durSeconds = 2.0 // Default duration
			}
			text = match[3]
		} else if len(match) >= 3 {
			// Pattern with start only or t attribute
			var err error
			startSeconds, err = strconv.ParseFloat(match[1], 64)
			if err != nil {
				continue
			}
			// Check if it's milliseconds (t attribute) or seconds (start attribute)
			if startSeconds > 10000 {
				// Likely milliseconds, convert to seconds
				startSeconds = startSeconds / 1000.0
			}
			if len(match) >= 4 && match[2] != "" {
				// Has duration attribute
				dMs, err := strconv.ParseFloat(match[2], 64)
				if err == nil {
					durSeconds = dMs / 1000.0
					hasDur = true
				} else {
					durSeconds = 2.0
				}
			} else {
				// No duration, will calculate from next segment
				durSeconds = 2.0
			}
			text = match[len(match)-1]
		} else {
			continue
		}

		// Decode HTML entities
		text = decodeHTMLEntities(text)
		text = strings.TrimSpace(text)
		text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")

		if text == "" {
			continue
		}

		rawSegments = append(rawSegments, rawSegment{
			startSeconds: startSeconds,
			durSeconds:   durSeconds,
			text:         text,
			hasDur:       hasDur,
		})
	}

	// Calculate durations for segments without dur attribute
	for i := 0; i < len(rawSegments); i++ {
		if !rawSegments[i].hasDur {
			if i < len(rawSegments)-1 {
				// Calculate duration from next segment's start time
				rawSegments[i].durSeconds = rawSegments[i+1].startSeconds - rawSegments[i].startSeconds
				if rawSegments[i].durSeconds <= 0 {
					rawSegments[i].durSeconds = 2.0 // Fallback
				}
			} else {
				// Last segment, use default duration
				rawSegments[i].durSeconds = 2.0
			}
		}
	}

	// Convert to final segments
	for _, seg := range rawSegments {
		endSeconds := seg.startSeconds + seg.durSeconds
		startTime := formatTimestampFromSeconds(seg.startSeconds)
		endTime := formatTimestampFromSeconds(endSeconds)

		segments = append(segments, models.TranscriptSegment{
			Start: startTime,
			End:   endTime,
			Text:  seg.text,
		})
	}

	return segments, nil
}

// mergeTranscriptSegments merges short segments into longer ones for better readability.
// targetDuration is the target duration for each merged segment in seconds.
func mergeTranscriptSegments(segments []models.TranscriptSegment, targetDuration int) []models.TranscriptSegment {
	if len(segments) == 0 {
		return segments
	}

	var merged []models.TranscriptSegment
	var currentTexts []string
	var currentStart string
	var currentEnd string
	var currentDuration float64

	for i, seg := range segments {
		// Parse start and end times
		startSecs := parseTimestampToSeconds(seg.Start)
		endSecs := parseTimestampToSeconds(seg.End)
		duration := endSecs - startSecs

		if len(currentTexts) == 0 {
			// Start a new merged segment
			currentStart = seg.Start
			currentTexts = []string{seg.Text}
			currentEnd = seg.End
			currentDuration = duration
		} else {
			// Check if we should merge with current segment
			if currentDuration < float64(targetDuration) {
				// Add to current segment
				currentTexts = append(currentTexts, seg.Text)
				currentEnd = seg.End
				currentDuration += duration
			} else {
				// Finish current segment and start new one
				merged = append(merged, models.TranscriptSegment{
					Start: currentStart,
					End:   currentEnd,
					Text:  strings.Join(currentTexts, " "),
				})
				currentStart = seg.Start
				currentTexts = []string{seg.Text}
				currentEnd = seg.End
				currentDuration = duration
			}
		}

		// Don't forget the last segment
		if i == len(segments)-1 && len(currentTexts) > 0 {
			merged = append(merged, models.TranscriptSegment{
				Start: currentStart,
				End:   currentEnd,
				Text:  strings.Join(currentTexts, " "),
			})
		}
	}

	return merged
}

// parseTimestampToSeconds converts HH:MM:SS format to seconds.
func parseTimestampToSeconds(timestamp string) float64 {
	parts := strings.Split(timestamp, ":")
	if len(parts) != 3 {
		return 0
	}
	hours, _ := strconv.Atoi(parts[0])
	minutes, _ := strconv.Atoi(parts[1])
	seconds, _ := strconv.Atoi(parts[2])
	return float64(hours*3600 + minutes*60 + seconds)
}

// formatTimestampFromSeconds converts seconds to HH:MM:SS format.
func formatTimestampFromSeconds(seconds float64) string {
	totalSeconds := int(seconds)
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	secs := totalSeconds % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, secs)
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
	// #region agent log
	logDebugYoutube("youtube.go:724", "parseYouTubeCaptions entry", map[string]interface{}{
		"contentLength": len(xmlContent),
		"contentPreview": xmlContent[:min(500, len(xmlContent))],
	}, "C")
	// #endregion

	// YouTube captions can be in different formats:
	// 1. XML format: <text start="0" dur="5.5">Caption text</text>
	// 2. JSON format (sometimes returned)

	// First, try XML format
	textPattern := regexp.MustCompile(`<text[^>]*start="([^"]*)"[^>]*>([\s\S]*?)</text>`)
	matches := textPattern.FindAllStringSubmatch(xmlContent, -1)

	// #region agent log
	logDebugYoutube("youtube.go:730", "First pattern matches", map[string]interface{}{
		"matchesCount": len(matches),
	}, "C")
	// #endregion

	if len(matches) == 0 {
		// Try alternative XML patterns
		altPattern := regexp.MustCompile(`<p[^>]*t="(\d+)"[^>]*>([\s\S]*?)</p>`)
		matches = altPattern.FindAllStringSubmatch(xmlContent, -1)
		
		// #region agent log
		logDebugYoutube("youtube.go:735", "Alternative pattern matches", map[string]interface{}{
			"matchesCount": len(matches),
		}, "C")
		// #endregion
	}

	if len(matches) == 0 {
		// #region agent log
		logDebugYoutube("youtube.go:840", "No matches found, trying more patterns", map[string]interface{}{
			"contentPreview": xmlContent[:min(1000, len(xmlContent))],
		}, "C")
		// #endregion

		// Try more patterns
		// Pattern 2: <text> without start attribute but with dur
		textPattern2 := regexp.MustCompile(`<text[^>]*dur="([^"]*)"[^>]*>([\s\S]*?)</text>`)
		matches = textPattern2.FindAllStringSubmatch(xmlContent, -1)
		
		// #region agent log
		logDebugYoutube("youtube.go:847", "Pattern 2 matches", map[string]interface{}{
			"matchesCount": len(matches),
		}, "C")
		// #endregion

		if len(matches) == 0 {
			// Pattern 3: <text> with any attributes
			textPattern3 := regexp.MustCompile(`<text[^>]*>([\s\S]*?)</text>`)
			matches = textPattern3.FindAllStringSubmatch(xmlContent, -1)
			
			// #region agent log
			logDebugYoutube("youtube.go:854", "Pattern 3 matches", map[string]interface{}{
				"matchesCount": len(matches),
			}, "C")
			// #endregion
		}

		if len(matches) == 0 {
			// Pattern 4: Try to extract all text content from XML (fallback)
			// Remove all XML tags and extract text
			textOnly := regexp.MustCompile(`<[^>]+>`).ReplaceAllString(xmlContent, "")
			textOnly = strings.TrimSpace(textOnly)
			if textOnly != "" {
				// #region agent log
				logDebugYoutube("youtube.go:864", "Using text-only fallback", map[string]interface{}{
					"textLength": len(textOnly),
					"textPreview": textOnly[:min(200, len(textOnly))],
				}, "C")
				// #endregion
				return textOnly
			}

			s.log.Debug("Failed to parse caption XML",
				zap.String("content_preview", xmlContent[:min(500, len(xmlContent))]),
			)
			return ""
		}
	}

	var result strings.Builder
	for i, match := range matches {
		// Handle different match patterns
		var timestamp string
		var text string
		
		if len(match) >= 3 {
			timestamp = match[1]
			text = match[2]
		} else if len(match) >= 2 {
			// Pattern 3: only text, no timestamp
			text = match[1]
			timestamp = fmt.Sprintf("%d", i*5) // Estimate 5 seconds per caption
		} else {
			continue
		}

		// Decode HTML entities
		text = decodeHTMLEntities(text)

		// Clean up whitespace
		text = strings.TrimSpace(text)
		text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")

		if text == "" {
			continue
		}

		// Convert timestamp to MM:SS format
		seconds, err := strconv.ParseFloat(timestamp, 64)
		if err != nil {
			// If timestamp parsing fails, use index-based estimate
			seconds = float64(i * 5)
		}
		// Handle milliseconds format (t="1234" means 1.234 seconds in some formats)
		if seconds > 10000 {
			seconds = seconds / 1000
		}
		mins := int(seconds) / 60
		secs := int(seconds) % 60

		result.WriteString(fmt.Sprintf("[%02d:%02d] %s\n", mins, secs, text))
	}

	parsedText := result.String()
	
	// #region agent log
	logDebugYoutube("youtube.go:900", "parseYouTubeCaptions result", map[string]interface{}{
		"matchesCount": len(matches),
		"parsedTextLength": len(parsedText),
		"parsedTextPreview": parsedText[:min(200, len(parsedText))],
	}, "C")
	// #endregion

	return parsedText
}

// convertPlainTextToSegments converts plain text with timestamps to structured segments.
func (s *YouTubeService) convertPlainTextToSegments(plainText string) []models.TranscriptSegment {
	var segments []models.TranscriptSegment
	
	// Parse format: [MM:SS] text
	// Example: [00:05] Hello world
	pattern := regexp.MustCompile(`\[(\d{2}):(\d{2})\]\s*(.+?)(?=\n\[|\z)`)
	matches := pattern.FindAllStringSubmatch(plainText, -1)
	
	for _, match := range matches {
		if len(match) < 4 {
			continue
		}
		
		minutes, err1 := strconv.Atoi(match[1])
		seconds, err2 := strconv.Atoi(match[2])
		text := strings.TrimSpace(match[3])
		
		if err1 != nil || err2 != nil || text == "" {
			continue
		}
		
		totalSeconds := minutes*60 + seconds
		startTime := formatTimestampFromSeconds(float64(totalSeconds))
		// Estimate 5 seconds duration per segment
		endTime := formatTimestampFromSeconds(float64(totalSeconds + 5))
		
		segments = append(segments, models.TranscriptSegment{
			Start: startTime,
			End:   endTime,
			Text:  text,
		})
	}
	
	return segments
}

// decodeHTMLEntities decodes HTML entities in text using Go's html package.
// Handles double-encoded entities common in YouTube captions.
func decodeHTMLEntities(text string) string {
	// YouTube captions are often double-encoded, so we unescape twice
	text = html.UnescapeString(text)
	text = html.UnescapeString(text)

	// Handle additional cleanup
	replacements := map[string]string{
		"&lrm;": "",
		"&rlm;": "",
		"\n":    " ",
		"\r":    "",
	}

	for entity, replacement := range replacements {
		text = strings.ReplaceAll(text, entity, replacement)
	}

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

// logDebugYoutube writes debug logs to file (for youtube.go)
func logDebugYoutube(location, message string, data map[string]interface{}, hypothesisIds string) {
	logEntry := map[string]interface{}{
		"location":     location,
		"message":      message,
		"data":         data,
		"timestamp":     time.Now().UnixMilli(),
		"sessionId":    "debug-session",
		"runId":        "run1",
		"hypothesisId": hypothesisIds,
	}
	logData, _ := json.Marshal(logEntry)
	
	// Try multiple log paths: env var, mounted volume, or fallback to /tmp
	logPath := os.Getenv("DEBUG_LOG_PATH")
	if logPath == "" {
		// Try workspace path first (for local development)
		workspacePath := "/Users/xiaozihao/Documents/01_Projects/Work_Code/work/Team_AI/vibe-engineering-playbook/.cursor/debug.log"
		if _, err := os.Stat("/Users/xiaozihao/Documents/01_Projects/Work_Code/work/Team_AI/vibe-engineering-playbook/.cursor"); err == nil {
			logPath = workspacePath
		} else {
			// Fallback to /tmp (always exists in Docker)
			logPath = "/tmp/debug.log"
		}
	}
	
	if f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		f.Write(logData)
		f.WriteString("\n")
		f.Close()
	}
}

// SecondsToTimestamp converts seconds to timestamp string (MM:SS).
func SecondsToTimestamp(seconds int) string {
	minutes := seconds / 60
	secs := seconds % 60
	return fmt.Sprintf("%02d:%02d", minutes, secs)
}
