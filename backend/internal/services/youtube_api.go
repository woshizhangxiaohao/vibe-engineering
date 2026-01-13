package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"

	"vibe-backend/internal/cache"
	"vibe-backend/internal/models"
)

// YouTubeAPIService handles YouTube Data API v3 operations with caching.
type YouTubeAPIService struct {
	apiKey        string
	cache         *cache.RedisCache
	oauthService  *OAuthService
	log           *zap.Logger
	quotaUsed     int64
	quotaTotal    int64
}

const (
	// Cache TTL durations
	videoCacheTTL    = 1 * time.Hour
	playlistCacheTTL = 30 * time.Minute
	captionsCacheTTL = 2 * time.Hour

	// Quota costs (YouTube Data API v3)
	quotaVideoMetadata = 1
	quotaPlaylist      = 1
	quotaCaptions      = 50
)

// NewYouTubeAPIService creates a new YouTubeAPIService.
func NewYouTubeAPIService(apiKey string, cache *cache.RedisCache, oauthService *OAuthService, log *zap.Logger) *YouTubeAPIService {
	return &YouTubeAPIService{
		apiKey:       apiKey,
		cache:        cache,
		oauthService: oauthService,
		log:          log,
		quotaUsed:    0,
		quotaTotal:   10000, // Default daily quota
	}
}

// GetVideoMetadata fetches video metadata with caching.
func (s *YouTubeAPIService) GetVideoMetadata(ctx context.Context, input string) (*models.YouTubeVideoResponse, error) {
	// Extract video ID from input
	videoID, err := s.extractVideoID(input)
	if err != nil {
		return nil, fmt.Errorf("INVALID_INPUT: %w", err)
	}

	// Check cache first (if cache is available)
	cacheKey := fmt.Sprintf("youtube:video:%s", videoID)
	if s.cache != nil {
		cached, err := s.cache.Get(ctx, cacheKey)
		if err == nil && cached != "" {
			var response models.YouTubeVideoResponse
			if err := json.Unmarshal([]byte(cached), &response); err == nil {
				response.CacheHit = true
				s.log.Debug("Cache hit for video metadata", zap.String("video_id", videoID))
				return &response, nil
			}
		}
	}

	// Create YouTube service
	service, err := youtube.NewService(ctx, option.WithAPIKey(s.apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create YouTube service: %w", err)
	}

	// Fetch video details
	call := service.Videos.List([]string{"snippet", "contentDetails", "status"}).Id(videoID)
	response, err := call.Do()
	if err != nil {
		s.log.Error("Failed to fetch video metadata", zap.Error(err), zap.String("video_id", videoID))
		return nil, fmt.Errorf("VIDEO_NOT_FOUND: failed to fetch video: %w", err)
	}

	if len(response.Items) == 0 {
		return nil, fmt.Errorf("VIDEO_NOT_FOUND: video does not exist or is private")
	}

	video := response.Items[0]

	// Check if video has captions
	hasCaptions := false
	if video.ContentDetails != nil {
		hasCaptions = video.ContentDetails.Caption == "true"
	}

	// Build response
	result := &models.YouTubeVideoResponse{
		ID:          videoID,
		Title:       video.Snippet.Title,
		Description: video.Snippet.Description,
		Duration:    video.ContentDetails.Duration,
		Thumbnails: models.YouTubeThumbnails{
			Default: video.Snippet.Thumbnails.Default.Url,
			High:    video.Snippet.Thumbnails.High.Url,
		},
		HasCaptions: hasCaptions,
		CacheHit:    false,
	}

	// Cache the result (if cache is available)
	if s.cache != nil {
		data, _ := json.Marshal(result)
		s.cache.Set(ctx, cacheKey, string(data), videoCacheTTL)
	}

	// Update quota
	s.incrementQuota(quotaVideoMetadata)

	return result, nil
}

// GetPlaylist fetches playlist items with caching.
func (s *YouTubeAPIService) GetPlaylist(ctx context.Context, playlistID string, token *oauth2.Token) (*models.YouTubePlaylistResponse, error) {
	// Check cache first (if cache is available)
	cacheKey := fmt.Sprintf("youtube:playlist:%s", playlistID)
	if s.cache != nil {
		cached, err := s.cache.Get(ctx, cacheKey)
		if err == nil && cached != "" {
			var response models.YouTubePlaylistResponse
			if err := json.Unmarshal([]byte(cached), &response); err == nil {
				response.CacheHit = true
				s.log.Debug("Cache hit for playlist", zap.String("playlist_id", playlistID))
				return &response, nil
			}
		}
	}

	// Create YouTube service with OAuth token if provided
	var service *youtube.Service
	var err error
	if token != nil {
		client := s.oauthService.config.Client(ctx, token)
		service, err = youtube.NewService(ctx, option.WithHTTPClient(client))
	} else {
		service, err = youtube.NewService(ctx, option.WithAPIKey(s.apiKey))
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create YouTube service: %w", err)
	}

	// Fetch playlist items
	call := service.PlaylistItems.List([]string{"snippet"}).PlaylistId(playlistID).MaxResults(50)
	response, err := call.Do()
	if err != nil {
		s.log.Error("Failed to fetch playlist", zap.Error(err), zap.String("playlist_id", playlistID))
		if strings.Contains(err.Error(), "forbidden") || strings.Contains(err.Error(), "unauthorized") {
			return nil, fmt.Errorf("UNAUTHORIZED: OAuth authorization required for private playlists")
		}
		return nil, fmt.Errorf("PLAYLIST_NOT_FOUND: failed to fetch playlist: %w", err)
	}

	// Build response
	items := make([]models.YouTubePlaylistItem, len(response.Items))
	for i, item := range response.Items {
		items[i] = models.YouTubePlaylistItem{
			VideoID:   item.Snippet.ResourceId.VideoId,
			Title:     item.Snippet.Title,
			Thumbnail: item.Snippet.Thumbnails.Default.Url,
		}
	}

	result := &models.YouTubePlaylistResponse{
		Items:    items,
		CacheHit: false,
	}

	// Cache the result (if cache is available)
	if s.cache != nil {
		data, _ := json.Marshal(result)
		s.cache.Set(ctx, cacheKey, string(data), playlistCacheTTL)
	}

	// Update quota
	s.incrementQuota(quotaPlaylist)

	return result, nil
}

// GetCaptions fetches caption tracks for a video.
func (s *YouTubeAPIService) GetCaptions(ctx context.Context, videoID string, token *oauth2.Token) (*models.YouTubeCaptionsResponse, error) {
	// Check cache first (if cache is available)
	cacheKey := fmt.Sprintf("youtube:captions:%s", videoID)
	if s.cache != nil {
		cached, err := s.cache.Get(ctx, cacheKey)
		if err == nil && cached != "" {
			var response models.YouTubeCaptionsResponse
			if err := json.Unmarshal([]byte(cached), &response); err == nil {
				s.log.Debug("Cache hit for captions", zap.String("video_id", videoID))
				return &response, nil
			}
		}
	}

	// Captions API requires OAuth
	if token == nil || token.AccessToken == "" {
		return nil, fmt.Errorf("UNAUTHORIZED: OAuth authorization required to access captions")
	}

	s.log.Debug("Creating YouTube service with OAuth token",
		zap.String("video_id", videoID),
		zap.Bool("has_token", token.AccessToken != ""),
	)

	// Create YouTube service with OAuth token
	// Use oauth2.StaticTokenSource to create a client that just uses the access token
	// This works even without GOOGLE_CLIENT_ID/SECRET configured
	tokenSource := oauth2.StaticTokenSource(token)
	client := oauth2.NewClient(ctx, tokenSource)
	service, err := youtube.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		s.log.Error("Failed to create YouTube service", zap.Error(err))
		return nil, fmt.Errorf("failed to create YouTube service: %w", err)
	}

	// Fetch caption tracks
	call := service.Captions.List([]string{"snippet"}, videoID)
	response, err := call.Do()
	if err != nil {
		s.log.Error("Failed to fetch captions", zap.Error(err), zap.String("video_id", videoID))
		return nil, fmt.Errorf("NO_CAPTIONS: failed to fetch captions: %w", err)
	}

	if len(response.Items) == 0 {
		return nil, fmt.Errorf("NO_CAPTIONS: 该视频未提供 API 可访问的字幕轨道")
	}

	// Build response
	captions := make([]models.YouTubeCaption, len(response.Items))
	for i, item := range response.Items {
		captions[i] = models.YouTubeCaption{
			ID:       item.Id,
			Language: item.Snippet.Language,
			Name:     item.Snippet.Name,
		}
	}

	result := &models.YouTubeCaptionsResponse{
		Captions: captions,
	}

	// Cache the result (if cache is available)
	if s.cache != nil {
		data, _ := json.Marshal(result)
		s.cache.Set(ctx, cacheKey, string(data), captionsCacheTTL)
	}

	// Update quota
	s.incrementQuota(quotaCaptions)

	return result, nil
}

// GetQuotaStatus returns the current quota usage.
func (s *YouTubeAPIService) GetQuotaStatus(ctx context.Context) *models.QuotaResponse {
	// Try to get quota from cache (if cache is available)
	if s.cache != nil {
		cacheKey := "youtube:quota:used"
		cached, err := s.cache.Get(ctx, cacheKey)
		if err == nil && cached != "" {
			var used int64
			if _, err := fmt.Sscanf(cached, "%d", &used); err == nil {
				s.quotaUsed = used
			}
		}
	}

	remaining := s.quotaTotal - s.quotaUsed
	if remaining < 0 {
		remaining = 0
	}

	percent := float64(s.quotaUsed) / float64(s.quotaTotal) * 100
	if percent > 100 {
		percent = 100
	}

	return &models.QuotaResponse{
		Total:     s.quotaTotal,
		Used:      s.quotaUsed,
		Remaining: remaining,
		Percent:   percent,
	}
}

// incrementQuota increments the quota usage and stores in cache.
func (s *YouTubeAPIService) incrementQuota(cost int64) {
	s.quotaUsed += cost

	// Store in cache with 24-hour expiry (quota resets daily) - if cache is available
	if s.cache != nil {
		ctx := context.Background()
		cacheKey := "youtube:quota:used"
		s.cache.Set(ctx, cacheKey, fmt.Sprintf("%d", s.quotaUsed), 24*time.Hour)
	}
}

// extractVideoID extracts the YouTube video ID from various input formats.
func (s *YouTubeAPIService) extractVideoID(input string) (string, error) {
	// If input is already a valid video ID (11 characters)
	if len(input) == 11 && !strings.Contains(input, "/") && !strings.Contains(input, "?") {
		return input, nil
	}

	// Parse as URL
	u, err := url.Parse(input)
	if err != nil {
		return "", fmt.Errorf("invalid input format")
	}

	// Handle youtu.be format
	if strings.Contains(u.Host, "youtu.be") {
		videoID := strings.TrimPrefix(u.Path, "/")
		if videoID == "" {
			return "", fmt.Errorf("invalid YouTube URL")
		}
		return videoID, nil
	}

	// Handle youtube.com format
	if strings.Contains(u.Host, "youtube.com") {
		query := u.Query()
		videoID := query.Get("v")
		if videoID == "" {
			return "", fmt.Errorf("video ID not found in URL")
		}
		return videoID, nil
	}

	// Fallback: try to extract 11-character ID using regex
	re := regexp.MustCompile(`([a-zA-Z0-9_-]{11})`)
	matches := re.FindStringSubmatch(input)
	if len(matches) > 1 {
		return matches[1], nil
	}

	return "", fmt.Errorf("could not extract video ID from input")
}
