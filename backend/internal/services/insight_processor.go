package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"go.uber.org/zap"

	"vibe-backend/internal/models"
	"vibe-backend/internal/repository"
)

// InsightProcessor handles async processing of insights.
type InsightProcessor struct {
	repo               *repository.InsightRepository
	youtubeService     *YouTubeService
	translationService *TranslationService
	log                *zap.Logger
}

// NewInsightProcessor creates a new InsightProcessor.
func NewInsightProcessor(
	repo *repository.InsightRepository,
	youtubeService *YouTubeService,
	log *zap.Logger,
) *InsightProcessor {
	// Note: translationService will be nil for now, needs to be injected
	return &InsightProcessor{
		repo:           repo,
		youtubeService: youtubeService,
		log:            log,
	}
}

// SetTranslationService sets the translation service (for dependency injection).
func (p *InsightProcessor) SetTranslationService(svc *TranslationService) {
	p.translationService = svc
}

// ProcessInsightAsync starts async processing of an insight.
// This should be called in a goroutine.
func (p *InsightProcessor) ProcessInsightAsync(ctx context.Context, insightID uint) {
	p.log.Info("Starting async insight processing", zap.Uint("insight_id", insightID))

	// Get the insight
	insight, err := p.repo.GetByID(ctx, insightID)
	if err != nil {
		p.log.Error("Failed to get insight for processing",
			zap.Uint("insight_id", insightID),
			zap.Error(err),
		)
		return
	}

	// Update status to processing
	if err := p.repo.UpdateStatus(ctx, insightID, models.InsightStatusProcessing, ""); err != nil {
		p.log.Error("Failed to update insight status to processing",
			zap.Uint("insight_id", insightID),
			zap.Error(err),
		)
		return
	}

	// Detect source type and process accordingly
	sourceType, err := p.detectSourceType(insight.SourceURL)
	if err != nil {
		p.handleProcessingError(ctx, insightID, fmt.Sprintf("无法识别来源类型: %v", err))
		return
	}

	insight.SourceType = sourceType

	switch sourceType {
	case models.SourceTypeYouTube:
		p.processYouTubeInsight(ctx, insight)
	default:
		p.handleProcessingError(ctx, insightID, fmt.Sprintf("暂不支持的来源类型: %s", sourceType))
	}
}

// detectSourceType detects the source type from a URL.
func (p *InsightProcessor) detectSourceType(sourceURL string) (models.SourceType, error) {
	lowerURL := strings.ToLower(sourceURL)

	// YouTube patterns
	if strings.Contains(lowerURL, "youtube.com") || strings.Contains(lowerURL, "youtu.be") {
		return models.SourceTypeYouTube, nil
	}

	// Twitter/X patterns
	if strings.Contains(lowerURL, "twitter.com") || strings.Contains(lowerURL, "x.com") {
		return models.SourceTypeTwitter, nil
	}

	// Podcast patterns (generic detection)
	if strings.Contains(lowerURL, "podcast") ||
		strings.Contains(lowerURL, "spotify.com") ||
		strings.Contains(lowerURL, "apple.com/podcast") {
		return models.SourceTypePodcast, nil
	}

	return "", fmt.Errorf("无法从 URL 识别来源类型: %s", sourceURL)
}

// processYouTubeInsight processes a YouTube video insight.
func (p *InsightProcessor) processYouTubeInsight(ctx context.Context, insight *models.Insight) {
	p.log.Info("Processing YouTube insight",
		zap.Uint("insight_id", insight.ID),
		zap.String("source_url", insight.SourceURL),
	)

	// Extract video ID
	videoID, err := p.youtubeService.ExtractVideoID(insight.SourceURL)
	if err != nil {
		p.handleProcessingError(ctx, insight.ID, fmt.Sprintf("无效的 YouTube URL: %v", err))
		return
	}

	insight.SourceID = videoID

	// Fetch video metadata
	metadata, err := p.youtubeService.GetVideoMetadataFromAPI(ctx, videoID)
	if err != nil {
		p.log.Warn("Failed to get video metadata from API, trying alternative method",
			zap.String("video_id", videoID),
			zap.Error(err),
		)
		// Try alternative method using Gemini
		metadata, err = p.youtubeService.GetVideoMetadata(ctx, insight.SourceURL)
		if err != nil {
			p.handleProcessingError(ctx, insight.ID, fmt.Sprintf("无法获取视频元数据: %v", err))
			return
		}
	}

	// Update insight with metadata
	insight.Title = metadata.Title
	insight.Author = metadata.Author
	insight.ThumbnailURL = metadata.ThumbnailURL
	insight.Duration = metadata.Duration

	// Fetch structured transcripts
	transcriptResponse, err := p.youtubeService.FetchYouTubeTranscriptStructured(ctx, videoID)
	if err != nil {
		p.log.Warn("Failed to fetch structured transcripts",
			zap.String("video_id", videoID),
			zap.Error(err),
		)
		// Transcripts are optional, continue processing
	} else {
		// Convert transcripts to the format expected by Insight model
		transcripts, err := p.convertTranscriptsToInsightFormat(transcriptResponse, insight.TargetLang)
		if err != nil {
			p.log.Warn("Failed to convert transcripts",
				zap.String("video_id", videoID),
				zap.Error(err),
			)
		} else {
			insight.Transcripts = transcripts

			// Also store raw content (combined transcript text)
			insight.RawContent = p.extractRawContentFromTranscripts(transcriptResponse)
		}
	}

	// Update insight with all collected data
	insight.Status = models.InsightStatusCompleted

	if err := p.repo.Update(ctx, insight); err != nil {
		p.log.Error("Failed to update insight after processing",
			zap.Uint("insight_id", insight.ID),
			zap.Error(err),
		)
		p.handleProcessingError(ctx, insight.ID, fmt.Sprintf("保存处理结果失败: %v", err))
		return
	}

	p.log.Info("Successfully processed YouTube insight",
		zap.Uint("insight_id", insight.ID),
		zap.String("title", insight.Title),
		zap.Int("duration", insight.Duration),
	)
}

// convertTranscriptsToInsightFormat converts YouTube transcripts to the Insight model format.
// It also translates the transcripts to the target language if translation service is available.
func (p *InsightProcessor) convertTranscriptsToInsightFormat(response *models.YouTubeTranscriptResponse, targetLang string) ([]byte, error) {
	// Convert to TranscriptItem array format expected by the Insight model
	var transcriptItems []models.TranscriptItem

	// Find the best available transcript (prefer default > auto > custom)
	for _, langData := range response.Transcripts {
		var segments []models.TranscriptSegment

		if len(langData.Default) > 0 {
			segments = langData.Default
		} else if len(langData.Auto) > 0 {
			segments = langData.Auto
		} else if len(langData.Custom) > 0 {
			segments = langData.Custom
		}

		if len(segments) == 0 {
			continue
		}

		// Convert segments to TranscriptItem format
		for _, seg := range segments {
			// Parse start time to seconds
			seconds := parseTimestampToSeconds(seg.Start)

			// Format timestamp as MM:SS
			mins := int(seconds) / 60
			secs := int(seconds) % 60
			timestamp := fmt.Sprintf("%02d:%02d", mins, secs)

			transcriptItems = append(transcriptItems, models.TranscriptItem{
				Timestamp: timestamp,
				Seconds:   int(seconds),
				Text:      seg.Text,
			})
		}

		// Only use the first language found
		break
	}

	if len(transcriptItems) == 0 {
		return nil, fmt.Errorf("no transcript segments found")
	}

	// Translate transcripts if translation service is available and target language is set
	if p.translationService != nil && targetLang != "" {
		p.log.Info("Translating transcripts",
			zap.Int("count", len(transcriptItems)),
			zap.String("target_lang", targetLang),
		)

		// Extract texts for batch translation
		texts := make([]string, len(transcriptItems))
		for i, item := range transcriptItems {
			texts[i] = item.Text
		}

		// Detect source language from first segment
		var sourceLang string
		if len(texts) > 0 {
			detected, err := p.translationService.DetectLanguage(context.Background(), texts[0])
			if err == nil {
				sourceLang = detected
			}
		}

		// Batch translate
		translations, err := p.translationService.TranslateBatch(context.Background(), texts, sourceLang, targetLang)
		if err != nil {
			p.log.Warn("Failed to translate transcripts, continuing without translation",
				zap.Error(err),
			)
		} else {
			// Add translations to transcript items
			for i, translation := range translations {
				if i < len(transcriptItems) {
					transcriptItems[i].TranslatedText = translation
				}
			}
			p.log.Info("Successfully translated transcripts")
		}
	}

	return json.Marshal(transcriptItems)
}

// extractRawContentFromTranscripts extracts plain text content from transcripts.
func (p *InsightProcessor) extractRawContentFromTranscripts(response *models.YouTubeTranscriptResponse) string {
	var textParts []string

	for _, langData := range response.Transcripts {
		var segments []models.TranscriptSegment

		if len(langData.Default) > 0 {
			segments = langData.Default
		} else if len(langData.Auto) > 0 {
			segments = langData.Auto
		} else if len(langData.Custom) > 0 {
			segments = langData.Custom
		}

		for _, seg := range segments {
			textParts = append(textParts, seg.Text)
		}

		// Only use the first language found
		break
	}

	return strings.Join(textParts, " ")
}

// handleProcessingError updates the insight status to failed with an error message.
func (p *InsightProcessor) handleProcessingError(ctx context.Context, insightID uint, errorMsg string) {
	p.log.Error("Insight processing failed",
		zap.Uint("insight_id", insightID),
		zap.String("error", errorMsg),
	)

	if err := p.repo.UpdateStatus(ctx, insightID, models.InsightStatusFailed, errorMsg); err != nil {
		p.log.Error("Failed to update insight status to failed",
			zap.Uint("insight_id", insightID),
			zap.Error(err),
		)
	}
}
