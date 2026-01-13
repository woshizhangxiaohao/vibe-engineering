package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"go.uber.org/zap"
)

// TranscriptService handles YouTube transcript extraction.
type TranscriptService struct {
	log *zap.Logger
}

// NewTranscriptService creates a new TranscriptService.
func NewTranscriptService(log *zap.Logger) *TranscriptService {
	return &TranscriptService{
		log: log,
	}
}

// TranscriptSegment represents a single transcript segment.
type TranscriptSegment struct {
	Start string `json:"start"`
	End   string `json:"end"`
	Text  string `json:"text"`
}

// TranscriptResponse represents the transcript API response.
type TranscriptResponse struct {
	VideoID     string              `json:"videoId"`
	Title       string              `json:"title"`
	Author      string              `json:"author"`
	Duration    string              `json:"duration"`
	Transcripts []TranscriptSegment `json:"transcripts"`
}

// ExtractVideoID extracts YouTube video ID from URL or returns the ID if already provided.
func ExtractVideoID(input string) (string, error) {
	// If input is already a video ID (11 characters, alphanumeric)
	if matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]{11}$`, input); matched {
		return input, nil
	}

	// Extract from various YouTube URL formats
	patterns := []string{
		`(?:youtube\.com\/watch\?v=|youtu\.be\/|youtube\.com\/embed\/)([a-zA-Z0-9_-]{11})`,
		`youtube\.com\/watch\?.*v=([a-zA-Z0-9_-]{11})`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(input)
		if len(matches) > 1 {
			return matches[1], nil
		}
	}

	return "", fmt.Errorf("invalid YouTube URL or video ID: %s", input)
}

// GetTranscript fetches transcript using yt-dlp.
func (s *TranscriptService) GetTranscript(ctx context.Context, input string) (*TranscriptResponse, error) {
	// Extract video ID
	videoID, err := ExtractVideoID(input)
	if err != nil {
		return nil, err
	}

	s.log.Info("Fetching transcript",
		zap.String("video_id", videoID),
	)

	// Build yt-dlp command to get video info and subtitles
	// First, get video metadata
	metadataCmd := exec.CommandContext(ctx,
		"yt-dlp",
		"--dump-json",
		"--no-warnings",
		"--skip-download",
		fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID),
	)

	metadataOutput, err := metadataCmd.Output()
	if err != nil {
		s.log.Error("Failed to fetch video metadata",
			zap.Error(err),
			zap.String("video_id", videoID),
		)
		return nil, fmt.Errorf("failed to fetch video metadata: %w", err)
	}

	// Parse metadata
	var metadata struct {
		Title    string `json:"title"`
		Uploader string `json:"uploader"`
		Duration int    `json:"duration"`
	}

	if err := json.Unmarshal(metadataOutput, &metadata); err != nil {
		s.log.Error("Failed to parse video metadata",
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to parse video metadata: %w", err)
	}

	// Get subtitles using yt-dlp
	subsCmd := exec.CommandContext(ctx,
		"yt-dlp",
		"--write-auto-sub",
		"--sub-lang", "en,zh-Hans,zh-Hant", // Try English and Chinese
		"--sub-format", "json3",
		"--skip-download",
		"--print", "requested_subtitles",
		"--no-warnings",
		fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID),
	)

	subsOutput, err := subsCmd.Output()
	if err != nil {
		// Try getting subtitles in a different way
		s.log.Warn("First subtitle attempt failed, trying alternative method",
			zap.Error(err),
		)

		// Alternative: download subtitle file directly
		subsCmd = exec.CommandContext(ctx,
			"yt-dlp",
			"--write-auto-sub",
			"--skip-download",
			"--sub-format", "vtt",
			"--output", "/tmp/%(id)s.%(ext)s",
			"--no-warnings",
			fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID),
		)

		if _, err := subsCmd.Output(); err != nil {
			s.log.Error("Failed to fetch subtitles",
				zap.Error(err),
				zap.String("video_id", videoID),
			)
			return nil, fmt.Errorf("no subtitles available for this video")
		}

		// Parse VTT file
		segments, err := s.parseVTTFile(fmt.Sprintf("/tmp/%s.en.vtt", videoID))
		if err != nil {
			// Try other language files
			segments, err = s.parseVTTFile(fmt.Sprintf("/tmp/%s.zh-Hans.vtt", videoID))
			if err != nil {
				return nil, fmt.Errorf("failed to parse subtitle file: %w", err)
			}
		}

		return &TranscriptResponse{
			VideoID:     videoID,
			Title:       metadata.Title,
			Author:      metadata.Uploader,
			Duration:    formatDuration(metadata.Duration),
			Transcripts: segments,
		}, nil
	}

	// Parse subtitle output
	segments := s.parseSubtitleOutput(string(subsOutput))

	return &TranscriptResponse{
		VideoID:     videoID,
		Title:       metadata.Title,
		Author:      metadata.Uploader,
		Duration:    formatDuration(metadata.Duration),
		Transcripts: segments,
	}, nil
}

// parseSubtitleOutput parses yt-dlp subtitle output.
func (s *TranscriptService) parseSubtitleOutput(output string) []TranscriptSegment {
	// Parse JSON3 format from yt-dlp
	var data struct {
		Events []struct {
			TStartMs int    `json:"tStartMs"`
			DDurationMs int `json:"dDurationMs"`
			Segs     []struct {
				Utf8 string `json:"utf8"`
			} `json:"segs"`
		} `json:"events"`
	}

	if err := json.Unmarshal([]byte(output), &data); err != nil {
		s.log.Warn("Failed to parse JSON3 subtitle format", zap.Error(err))
		return []TranscriptSegment{}
	}

	segments := make([]TranscriptSegment, 0, len(data.Events))
	for _, event := range data.Events {
		if len(event.Segs) == 0 {
			continue
		}

		// Combine all text segments
		var text strings.Builder
		for _, seg := range event.Segs {
			text.WriteString(seg.Utf8)
		}

		// Convert milliseconds to HH:MM:SS format
		startMs := event.TStartMs
		endMs := startMs + event.DDurationMs

		segments = append(segments, TranscriptSegment{
			Start: formatTimestamp(startMs),
			End:   formatTimestamp(endMs),
			Text:  strings.TrimSpace(text.String()),
		})
	}

	return segments
}

// parseVTTFile parses WebVTT subtitle file.
func (s *TranscriptService) parseVTTFile(filePath string) ([]TranscriptSegment, error) {
	cmd := exec.Command("cat", filePath)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to read VTT file: %w", err)
	}

	content := string(output)
	lines := strings.Split(content, "\n")

	var segments []TranscriptSegment
	var currentText strings.Builder
	var start, end string

	// VTT format:
	// WEBVTT
	//
	// 00:00:00.000 --> 00:00:09.000
	// Text here
	//
	// 00:00:09.000 --> 00:00:15.000
	// More text

	timestampRegex := regexp.MustCompile(`(\d{2}:\d{2}:\d{2}\.\d{3})\s*-->\s*(\d{2}:\d{2}:\d{2}\.\d{3})`)

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		// Skip empty lines and WEBVTT header
		if line == "" || strings.HasPrefix(line, "WEBVTT") || strings.HasPrefix(line, "Kind:") || strings.HasPrefix(line, "Language:") {
			continue
		}

		// Check if this line contains timestamp
		matches := timestampRegex.FindStringSubmatch(line)
		if len(matches) == 3 {
			// Save previous segment if exists
			if start != "" && currentText.Len() > 0 {
				segments = append(segments, TranscriptSegment{
					Start: start,
					End:   end,
					Text:  strings.TrimSpace(currentText.String()),
				})
				currentText.Reset()
			}

			// Parse new timestamp
			start = matches[1]
			end = matches[2]
		} else if start != "" && !regexp.MustCompile(`^\d+$`).MatchString(line) {
			// This is text content (not a sequence number)
			// Remove HTML tags and formatting
			text := removeVTTFormatting(line)
			if text != "" {
				if currentText.Len() > 0 {
					currentText.WriteString(" ")
				}
				currentText.WriteString(text)
			}
		}
	}

	// Add last segment
	if start != "" && currentText.Len() > 0 {
		segments = append(segments, TranscriptSegment{
			Start: start,
			End:   end,
			Text:  strings.TrimSpace(currentText.String()),
		})
	}

	return segments, nil
}

// removeVTTFormatting removes VTT formatting tags from text.
func removeVTTFormatting(text string) string {
	// Remove <c> tags, <i> tags, etc.
	re := regexp.MustCompile(`<[^>]+>`)
	text = re.ReplaceAllString(text, "")
	return strings.TrimSpace(text)
}

// formatDuration converts seconds to HH:MM:SS format.
func formatDuration(seconds int) string {
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60

	if hours > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, secs)
	}
	return fmt.Sprintf("%02d:%02d", minutes, secs)
}

// formatTimestamp converts milliseconds to HH:MM:SS format.
func formatTimestamp(ms int) string {
	seconds := ms / 1000
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60

	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, secs)
}
