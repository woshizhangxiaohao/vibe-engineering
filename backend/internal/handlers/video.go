package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"vibe-backend/internal/models"
	"vibe-backend/internal/repository"
	"vibe-backend/internal/services"
)

// VideoHandler handles video analysis HTTP requests.
type VideoHandler struct {
	repo           *repository.VideoRepository
	youtubeService *services.YouTubeService
	log            *zap.Logger
}

// NewVideoHandler creates a new VideoHandler.
func NewVideoHandler(repo *repository.VideoRepository, youtubeService *services.YouTubeService, log *zap.Logger) *VideoHandler {
	return &VideoHandler{
		repo:           repo,
		youtubeService: youtubeService,
		log:            log,
	}
}

// GetMetadata fetches video metadata and AI analysis directly using Gemini.
// POST /api/v1/videos/metadata
// Request body: {"url": "https://youtube.com/watch?v=..."} or {"videoId": "..."}
func (h *VideoHandler) GetMetadata(c *gin.Context) {
	var req struct {
		URL     string `json:"url"`
		VideoID string `json:"videoId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": fmt.Sprintf("请求格式错误: %v", err),
		})
		return
	}

	// Determine video URL
	var videoURL string
	if req.URL != "" {
		videoURL = req.URL
	} else if req.VideoID != "" {
		videoURL = fmt.Sprintf("https://www.youtube.com/watch?v=%s", req.VideoID)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "必须提供 url 或 videoId",
		})
		return
	}

	// Extract video ID for validation
	videoID, err := h.youtubeService.ExtractVideoID(videoURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_URL",
			"message": fmt.Sprintf("无效的 YouTube URL: %v", err),
		})
		return
	}

	// Get metadata
	metadata, err := h.youtubeService.GetVideoMetadata(c.Request.Context(), videoURL)
	if err != nil {
		h.log.Error("Failed to get metadata",
			zap.Error(err),
			zap.String("video_id", videoID),
		)
		// Check if error is due to missing API key
		if strings.Contains(err.Error(), "OPENROUTER_API_KEY") {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    "API_KEY_MISSING",
				"message": "OPENROUTER_API_KEY 环境变量未配置，请设置该环境变量以使用视频分析功能",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "METADATA_FETCH_FAILED",
			"message": fmt.Sprintf("无法获取视频信息: %v", err),
		})
		return
	}

	// 不再检查私有视频，因为：
	// 1. API 可能返回空 title，但不代表视频是私有的
	// 2. 实际的分析会在 AnalyzeVideo 中进行，那里会真正尝试获取字幕
	// 3. 如果视频真的是私有的，在分析时会失败并返回相应错误

	c.JSON(http.StatusOK, models.MetadataResponse{
		VideoID:      metadata.VideoID,
		Title:        metadata.Title,
		Author:       metadata.Author,
		ThumbnailURL: metadata.ThumbnailURL,
		Duration:     metadata.Duration,
	})
}

// AnalyzeVideo calls Gemini to analyze video and returns jobId for frontend compatibility.
// POST /api/v1/videos/analyze
// Request body: {"videoId": "video_id", "targetLanguage": "en"} or {"url": "https://youtube.com/watch?v=..."}
func (h *VideoHandler) AnalyzeVideo(c *gin.Context) {
	var req struct {
		VideoID        string `json:"videoId"`
		URL            string `json:"url"`
		TargetLanguage string `json:"targetLanguage"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": err.Error(),
		})
		return
	}

	// Determine video URL and video ID
	var videoURL string
	var videoID string
	if req.URL != "" {
		videoURL = req.URL
		var err error
		videoID, err = h.youtubeService.ExtractVideoID(req.URL)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    "INVALID_URL",
				"message": "无效的 YouTube URL",
			})
			return
		}
	} else if req.VideoID != "" {
		videoID = req.VideoID
		videoURL = fmt.Sprintf("https://www.youtube.com/watch?v=%s", req.VideoID)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "必须提供 videoId 或 url",
		})
		return
	}

	// Generate job ID for frontend compatibility
	jobID := uuid.New().String()

	// TODO: Get user ID from JWT token
	userID := uint(1)

	// Create analysis record
	analysis := &models.VideoAnalysis{
		UserID:         userID,
		VideoID:        videoID,
		TargetLanguage: req.TargetLanguage,
		JobID:          jobID,
		Status:         "processing",
	}

	if err := h.repo.CreateAnalysis(c.Request.Context(), analysis); err != nil {
		h.log.Error("Failed to create analysis record",
			zap.Error(err),
			zap.String("video_id", videoID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "ANALYSIS_FAILED",
			"message": "无法创建解析任务",
		})
		return
	}

	// Call Gemini directly and save result (even if error, save raw response)
	go func() {
		ctx := context.Background()
		response, err := h.youtubeService.CallGeminiDirect(ctx, videoURL)

		// Get analysis record
		analysisRecord, err2 := h.repo.GetAnalysisByJobID(ctx, jobID)
		if err2 != nil {
			h.log.Error("Failed to get analysis record", zap.Error(err2))
			return
		}

		// Even if Gemini returns error, save the response as completed
		// User said: "大模型如果返回错误，也是正确的"
		if err != nil {
			h.log.Warn("Gemini returned error, but saving as completed",
				zap.Error(err),
				zap.String("video_url", videoURL),
				zap.String("job_id", jobID),
			)
			// Save error message as transcription
			analysisRecord.Status = "completed"
			analysisRecord.Summary = err.Error()
			if err := h.repo.UpdateAnalysis(ctx, analysisRecord); err != nil {
				h.log.Error("Failed to update analysis", zap.Error(err))
				return
			}
			// Save error as a single transcription entry
			transcriptions := []models.Transcription{
				{
					AnalysisID: analysisRecord.ID,
					Text:       fmt.Sprintf("Error: %v", err),
					Timestamp:  "00:00",
					Seconds:    0,
					OrderIndex: 0,
				},
			}
			h.repo.CreateTranscriptions(ctx, transcriptions)
			return
		}

		// Parse response to extract transcription
		var result struct {
			Transcription []struct {
				Text      string `json:"text"`
				Timestamp string `json:"timestamp"`
				Seconds   int    `json:"seconds"`
			} `json:"transcription"`
		}

		// Clean response before parsing
		cleanedResponse := strings.TrimSpace(response)
		if strings.HasPrefix(cleanedResponse, "```json") {
			cleanedResponse = strings.TrimPrefix(cleanedResponse, "```json")
			cleanedResponse = strings.TrimSpace(cleanedResponse)
		} else if strings.HasPrefix(cleanedResponse, "```") {
			cleanedResponse = strings.TrimPrefix(cleanedResponse, "```")
			cleanedResponse = strings.TrimSpace(cleanedResponse)
		}
		if strings.HasSuffix(cleanedResponse, "```") {
			cleanedResponse = strings.TrimSuffix(cleanedResponse, "```")
			cleanedResponse = strings.TrimSpace(cleanedResponse)
		}
		if !strings.HasPrefix(cleanedResponse, "{") {
			startIdx := strings.Index(cleanedResponse, "{")
			endIdx := strings.LastIndex(cleanedResponse, "}")
			if startIdx != -1 && endIdx != -1 && endIdx > startIdx {
				cleanedResponse = cleanedResponse[startIdx : endIdx+1]
			}
		}

		// Even if parsing fails, save raw response as completed
		if err := json.Unmarshal([]byte(cleanedResponse), &result); err != nil {
			h.log.Warn("Failed to parse Gemini response, but saving raw response as completed",
				zap.Error(err),
				zap.String("job_id", jobID),
				zap.String("raw_response", response),
			)
			// Save raw response as transcription
			analysisRecord.Status = "completed"
			analysisRecord.Summary = response // Save raw response
			if err := h.repo.UpdateAnalysis(ctx, analysisRecord); err != nil {
				h.log.Error("Failed to update analysis", zap.Error(err))
				return
			}
			// Save raw response as a single transcription entry
			transcriptions := []models.Transcription{
				{
					AnalysisID: analysisRecord.ID,
					Text:       response,
					Timestamp:  "00:00",
					Seconds:    0,
					OrderIndex: 0,
				},
			}
			h.repo.CreateTranscriptions(ctx, transcriptions)
			return
		}

		// Update analysis status to completed
		analysisRecord.Status = "completed"
		if err := h.repo.UpdateAnalysis(ctx, analysisRecord); err != nil {
			h.log.Error("Failed to update analysis", zap.Error(err))
			return
		}

		// Save transcriptions
		transcriptions := make([]models.Transcription, len(result.Transcription))
		for i, tr := range result.Transcription {
			transcriptions[i] = models.Transcription{
				AnalysisID: analysisRecord.ID,
				Text:       tr.Text,
				Timestamp:  tr.Timestamp,
				Seconds:    tr.Seconds,
				OrderIndex: i,
			}
		}
		if err := h.repo.CreateTranscriptions(ctx, transcriptions); err != nil {
			h.log.Error("Failed to save transcriptions", zap.Error(err))
		}
	}()

	// Return jobId immediately for frontend compatibility
	// #region agent log
	func() {
		logFile, _ := os.OpenFile("/Users/xiaozihao/Documents/01_Projects/Work_Code/work/Team_AI/vibe-engineering-playbook/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if logFile != nil {
			defer logFile.Close()
			logData := map[string]interface{}{
				"sessionId":    "debug-session",
				"runId":        "run1",
				"hypothesisId": "B",
				"location":     "video.go:306",
				"message":      "Returning jobId to frontend",
				"data":         map[string]interface{}{"jobId": jobID, "status": "processing", "videoId": videoID, "requestBody": map[string]interface{}{"videoId": req.VideoID, "url": req.URL, "targetLanguage": req.TargetLanguage}},
				"timestamp":    time.Now().UnixMilli(),
			}
			json.NewEncoder(logFile).Encode(logData)
		}
	}()
	// #endregion

	// Ensure response format matches frontend expectation exactly
	response := gin.H{
		"jobId":  jobID,
		"status": "processing",
	}

	h.log.Info("Returning jobId to frontend",
		zap.String("job_id", jobID),
		zap.String("video_id", videoID),
		zap.Any("response", response),
	)

	c.JSON(http.StatusOK, response)
}

// processAnalysis performs the actual video analysis asynchronously.
func (h *VideoHandler) processAnalysis(ctx context.Context, analysisID uint, videoID, targetLanguage string) {
	h.log.Info("Starting video analysis",
		zap.Uint("analysis_id", analysisID),
		zap.String("video_id", videoID),
	)

	// Perform analysis
	result, err := h.youtubeService.AnalyzeVideo(ctx, videoID, targetLanguage)
	if err != nil {
		h.log.Error("Video analysis failed",
			zap.Error(err),
			zap.String("video_id", videoID),
		)
		return
	}

	// Get the analysis record from database by ID
	analysisRecord, err := h.repo.GetAnalysisByID(ctx, analysisID)
	if err != nil {
		h.log.Error("Failed to retrieve analysis record", zap.Error(err))
		return
	}

	// Update analysis with results - 只保存字幕，不保存摘要、关键点和章节
	analysisRecord.Summary = ""
	analysisRecord.Status = "completed"
	if err := h.repo.UpdateAnalysis(ctx, analysisRecord); err != nil {
		h.log.Error("Failed to update analysis", zap.Error(err))
		return
	}

	// 只保存字幕内容
	transcriptions := make([]models.Transcription, len(result.Transcription))
	for i, tr := range result.Transcription {
		transcriptions[i] = models.Transcription{
			AnalysisID: analysisID,
			Text:       tr.Text,
			Timestamp:  tr.Timestamp,
			Seconds:    tr.Seconds,
			OrderIndex: i,
		}
	}
	if err := h.repo.CreateTranscriptions(ctx, transcriptions); err != nil {
		h.log.Error("Failed to save transcriptions", zap.Error(err))
	}

	h.log.Info("Video analysis completed",
		zap.Uint("analysis_id", analysisID),
		zap.String("video_id", videoID),
	)
}

// GetResult retrieves the analysis result by job ID.
// GET /api/v1/videos/result/:jobId
func (h *VideoHandler) GetResult(c *gin.Context) {
	jobID := c.Param("jobId")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "Job ID is required",
		})
		return
	}

	// Get analysis
	analysis, err := h.repo.GetAnalysisByJobID(c.Request.Context(), jobID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    "JOB_NOT_FOUND",
				"message": "解析任务不存在",
			})
			return
		}
		h.log.Error("Failed to get analysis", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "INTERNAL_ERROR",
			"message": "获取解析结果失败",
		})
		return
	}

	// If still processing or pending, return status
	if analysis.Status == "pending" || analysis.Status == "processing" {
		c.JSON(http.StatusOK, models.AnalysisResultResponse{
			Status: analysis.Status,
		})
		return
	}

	// If failed, return error
	if analysis.Status == "failed" {
		c.JSON(http.StatusOK, gin.H{
			"code":    "ANALYSIS_FAILED",
			"message": "解析失败，请重试",
			"status":  "failed",
		})
		return
	}

	// 只获取字幕内容
	transcriptions, err := h.repo.GetTranscriptionsByAnalysisID(c.Request.Context(), analysis.ID)
	if err != nil {
		h.log.Error("Failed to get transcriptions", zap.Error(err))
	}

	// Convert to response format - 只返回字幕
	transcriptionsResp := make([]models.TranscriptionResponse, len(transcriptions))
	for i, tr := range transcriptions {
		transcriptionsResp[i] = models.TranscriptionResponse{
			Text:      tr.Text,
			Timestamp: tr.Timestamp,
			Seconds:   tr.Seconds,
		}
	}

	c.JSON(http.StatusOK, models.AnalysisResultResponse{
		AnalysisID:    analysis.ID,
		Status:        "completed",
		Summary:       "",
		KeyPoints:     []string{},
		Chapters:      []models.ChapterResponse{},
		Transcription: transcriptionsResp,
	})
}

// GetHistory retrieves the user's analysis history.
// GET /api/v1/history
func (h *VideoHandler) GetHistory(c *gin.Context) {
	// TODO: Get user ID from JWT token
	userID := uint(1)

	analyses, err := h.repo.GetHistoryByUserID(c.Request.Context(), userID, 20)
	if err != nil {
		h.log.Error("Failed to get history", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "INTERNAL_ERROR",
			"message": "获取历史记录失败",
		})
		return
	}

	items := make([]models.HistoryItem, len(analyses))
	for i, a := range analyses {
		items[i] = models.HistoryItem{
			VideoID:      a.VideoID,
			Title:        a.Title,
			ThumbnailURL: a.ThumbnailURL,
			CreatedAt:    a.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, models.HistoryResponse{
		Items: items,
	})
}

// ExportVideo exports the analysis results to PDF or Markdown.
// POST /api/v1/videos/export
func (h *VideoHandler) ExportVideo(c *gin.Context) {
	var req models.ExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": err.Error(),
		})
		return
	}

	// TODO: Get user ID from JWT token
	userID := uint(1)

	// Get latest analysis for this video
	analysis, err := h.repo.GetAnalysisByVideoID(c.Request.Context(), req.VideoID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    "VIDEO_NOT_FOUND",
				"message": "视频分析记录不存在",
			})
			return
		}
		h.log.Error("Failed to get analysis", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "INTERNAL_ERROR",
			"message": "导出失败",
		})
		return
	}

	// Get all related data
	keyPoints, _ := h.repo.GetKeyPointsByAnalysisID(c.Request.Context(), analysis.ID)
	chapters, _ := h.repo.GetChaptersByAnalysisID(c.Request.Context(), analysis.ID)
	transcriptions, _ := h.repo.GetTranscriptionsByAnalysisID(c.Request.Context(), analysis.ID)

	// Generate file based on format
	var content string
	var fileName string

	if req.Format == "markdown" {
		content = h.generateMarkdown(analysis, keyPoints, chapters, transcriptions)
		fileName = fmt.Sprintf("%s_%s.md", analysis.VideoID, time.Now().Format("20060102"))
	} else {
		// PDF export would require a PDF library (not implemented in this version)
		c.JSON(http.StatusNotImplemented, gin.H{
			"code":    "NOT_IMPLEMENTED",
			"message": "PDF 导出功能暂未实现",
		})
		return
	}

	// In a real implementation, you would upload to cloud storage and return a URL
	// For now, we'll return the content directly
	_ = content // content is generated but not yet used (will be uploaded to storage in future)
	downloadURL := fmt.Sprintf("/api/v1/downloads/%s", fileName)

	c.JSON(http.StatusOK, models.ExportResponse{
		DownloadURL: downloadURL,
		FileName:    fileName,
	})
}

// generateMarkdown generates Markdown content from analysis data.
func (h *VideoHandler) generateMarkdown(
	analysis *models.VideoAnalysis,
	keyPoints []models.KeyPoint,
	chapters []models.Chapter,
	transcriptions []models.Transcription,
) string {
	var md string

	md += fmt.Sprintf("# %s\n\n", analysis.Title)
	md += fmt.Sprintf("**作者**: %s\n\n", analysis.Author)
	md += fmt.Sprintf("**视频ID**: %s\n\n", analysis.VideoID)
	md += fmt.Sprintf("**分析时间**: %s\n\n", analysis.CreatedAt.Format("2006-01-02 15:04:05"))
	md += "---\n\n"

	md += "## 摘要\n\n"
	md += analysis.Summary + "\n\n"

	md += "## 核心观点\n\n"
	for i, kp := range keyPoints {
		md += fmt.Sprintf("%d. %s\n", i+1, kp.Content)
	}
	md += "\n"

	md += "## 章节\n\n"
	for _, ch := range chapters {
		md += fmt.Sprintf("### [%s] %s\n\n", ch.Timestamp, ch.Title)
	}

	md += "## 完整转录\n\n"
	for _, tr := range transcriptions {
		md += fmt.Sprintf("**[%s]** %s\n\n", tr.Timestamp, tr.Text)
	}

	return md
}
