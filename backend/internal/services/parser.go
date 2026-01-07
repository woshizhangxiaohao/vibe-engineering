package services

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"vibe-backend/internal/models"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

var (
	// ErrInvalidURL is returned when the URL is not a valid YouTube or Twitter link.
	ErrInvalidURL = errors.New("invalid URL: not a supported YouTube or Twitter link")
	// ErrParsingFailed is returned when parsing fails.
	ErrParsingFailed = errors.New("parsing failed: unable to extract metadata")
)

// ParserService handles URL parsing and content extraction.
type ParserService struct {
	log *zap.Logger
}

// NewParserService creates a new ParserService.
func NewParserService(log *zap.Logger) *ParserService {
	return &ParserService{log: log}
}

// Parse parses a URL and extracts metadata.
func (s *ParserService) Parse(ctx context.Context, rawURL string) (*models.ParsedContent, error) {
	// Validate and detect source
	source, err := s.detectSource(rawURL)
	if err != nil {
		return nil, err
	}

	s.log.Info("Parsing URL",
		zap.String("url", rawURL),
		zap.String("source", string(source)),
	)

	// Parse based on source
	var content *models.ParsedContent
	switch source {
	case models.SourceYouTube:
		content, err = s.parseYouTube(ctx, rawURL)
	case models.SourceTwitter:
		content, err = s.parseTwitter(ctx, rawURL)
	default:
		return nil, ErrInvalidURL
	}

	if err != nil {
		s.log.Error("Failed to parse URL",
			zap.String("url", rawURL),
			zap.Error(err),
		)
		return nil, err
	}

	// Generate unique ID
	content.ID = uuid.New().String()
	content.OriginalURL = rawURL

	return content, nil
}

// detectSource detects the source platform from the URL.
func (s *ParserService) detectSource(rawURL string) (models.ContentSource, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", ErrInvalidURL
	}

	host := strings.ToLower(parsedURL.Host)

	// Check for YouTube
	if strings.Contains(host, "youtube.com") || strings.Contains(host, "youtu.be") {
		return models.SourceYouTube, nil
	}

	// Check for Twitter/X
	if strings.Contains(host, "twitter.com") || strings.Contains(host, "x.com") {
		return models.SourceTwitter, nil
	}

	return "", ErrInvalidURL
}

// parseYouTube extracts metadata from a YouTube URL.
func (s *ParserService) parseYouTube(ctx context.Context, rawURL string) (*models.ParsedContent, error) {
	// TODO: Implement YouTube API integration or web scraping
	// For now, return mock data for demonstration purposes

	s.log.Warn("YouTube parsing using mock data - implement YouTube API integration",
		zap.String("url", rawURL),
	)

	// Extract video ID (basic implementation)
	videoID, err := s.extractYouTubeVideoID(rawURL)
	if err != nil {
		return nil, ErrParsingFailed
	}

	// Mock content - replace with actual API call
	publishedAt := time.Now().Add(-24 * time.Hour) // Mock: 1 day ago

	content := &models.ParsedContent{
		Source:       models.SourceYouTube,
		Title:        fmt.Sprintf("YouTube Video: %s", videoID),
		Author:       "Channel Name (TODO: fetch from API)",
		Summary:      s.generateMockSummary("YouTube", videoID),
		ThumbnailURL: fmt.Sprintf("https://img.youtube.com/vi/%s/maxresdefault.jpg", videoID),
		PublishedAt:  &publishedAt,
	}

	return content, nil
}

// parseTwitter extracts metadata from a Twitter/X URL.
func (s *ParserService) parseTwitter(ctx context.Context, rawURL string) (*models.ParsedContent, error) {
	// TODO: Implement Twitter API integration or web scraping
	// For now, return mock data for demonstration purposes

	s.log.Warn("Twitter parsing using mock data - implement Twitter API integration",
		zap.String("url", rawURL),
	)

	// Extract tweet ID (basic implementation)
	tweetID, err := s.extractTwitterTweetID(rawURL)
	if err != nil {
		return nil, ErrParsingFailed
	}

	// Mock content - replace with actual API call
	publishedAt := time.Now().Add(-6 * time.Hour) // Mock: 6 hours ago

	content := &models.ParsedContent{
		Source:       models.SourceTwitter,
		Title:        fmt.Sprintf("Tweet: %s", tweetID),
		Author:       "@username (TODO: fetch from API)",
		Summary:      s.generateMockSummary("Twitter", tweetID),
		ThumbnailURL: "https://abs.twimg.com/icons/apple-touch-icon-192x192.png", // Default Twitter icon
		PublishedAt:  &publishedAt,
	}

	return content, nil
}

// extractYouTubeVideoID extracts the video ID from a YouTube URL.
func (s *ParserService) extractYouTubeVideoID(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	// Handle youtu.be short links
	if strings.Contains(parsedURL.Host, "youtu.be") {
		videoID := strings.TrimPrefix(parsedURL.Path, "/")
		if videoID != "" {
			return videoID, nil
		}
	}

	// Handle youtube.com links
	if strings.Contains(parsedURL.Host, "youtube.com") {
		query := parsedURL.Query()
		videoID := query.Get("v")
		if videoID != "" {
			return videoID, nil
		}
	}

	return "", errors.New("could not extract YouTube video ID")
}

// extractTwitterTweetID extracts the tweet ID from a Twitter/X URL.
func (s *ParserService) extractTwitterTweetID(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	// Twitter URL format: https://twitter.com/username/status/1234567890
	// or: https://x.com/username/status/1234567890
	parts := strings.Split(strings.TrimPrefix(parsedURL.Path, "/"), "/")

	for i, part := range parts {
		if part == "status" && i+1 < len(parts) {
			return parts[i+1], nil
		}
	}

	return "", errors.New("could not extract Twitter tweet ID")
}

// generateMockSummary generates a mock AI summary.
// TODO: Replace with actual AI/LLM integration (OpenAI, Anthropic, etc.)
func (s *ParserService) generateMockSummary(source, id string) string {
	summaries := []string{
		"This content discusses innovative approaches to modern web development.",
		"An insightful perspective on current industry trends and best practices.",
		"Practical tips and strategies for improving productivity and workflow.",
	}

	// Simple rotation based on ID length
	index := len(id) % len(summaries)

	return fmt.Sprintf("[MOCK SUMMARY - TODO: Implement AI integration] %s", summaries[index])
}
