package media

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"
)

// SubtitleSegment represents a single subtitle entry with timing.
type SubtitleSegment struct {
	Text      string  `json:"text"`
	StartTime float64 `json:"start_time"` // seconds
	EndTime   float64 `json:"end_time"`   // seconds
}

// SubtitleResult holds the generated subtitle output.
type SubtitleResult struct {
	SRTContent string            `json:"srt_content"`
	Segments   []SubtitleSegment `json:"segments"`
}

// SubtitleGenerator creates subtitle files (SRT format) from text content.
type SubtitleGenerator struct{}

// NewSubtitleGenerator creates a new SubtitleGenerator.
func NewSubtitleGenerator() *SubtitleGenerator {
	return &SubtitleGenerator{}
}

// Generate creates SRT subtitles from pre-defined segments.
func (g *SubtitleGenerator) Generate(ctx context.Context, text string, segments []SubtitleSegment) (*SubtitleResult, error) {
	if len(segments) == 0 {
		return nil, fmt.Errorf("subtitle: no segments provided")
	}

	sb := &strings.Builder{}
	for i, seg := range segments {
		idx := i + 1
		startStr := formatSRTTime(seg.StartTime)
		endStr := formatSRTTime(seg.EndTime)
		fmt.Fprintf(sb, "%d\n%s --> %s\n%s\n\n", idx, startStr, endStr, seg.Text)
	}

	return &SubtitleResult{
		SRTContent: sb.String(),
		Segments:   segments,
	}, nil
}

// GenerateSRT automatically splits text into timed segments and generates SRT content.
// It distributes the text evenly across the audio duration, splitting at sentence boundaries.
func (g *SubtitleGenerator) GenerateSRT(ctx context.Context, text string, audioDuration float64) (*SubtitleResult, error) {
	if text == "" {
		return nil, fmt.Errorf("subtitle: empty text")
	}
	if audioDuration <= 0 {
		return nil, fmt.Errorf("subtitle: invalid audio duration %f", audioDuration)
	}

	sentences := splitSentences(text)
	if len(sentences) == 0 {
		sentences = []string{text}
	}

	segmentDuration := audioDuration / float64(len(sentences))
	const maxSegmentDuration = 8.0
	if segmentDuration > maxSegmentDuration {
		// If each segment is too long, merge adjacent sentences into fewer segments
		segments := mergeSentences(sentences, audioDuration)
		return g.buildSRT(segments, audioDuration)
	}

	return g.buildSRT(sentences, audioDuration)
}

func (g *SubtitleGenerator) buildSRT(sentences []string, totalDuration float64) (*SubtitleResult, error) {
	segCount := len(sentences)
	segmentDuration := totalDuration / float64(segCount)
	segments := make([]SubtitleSegment, segCount)

	sb := &strings.Builder{}
	for i, sentence := range sentences {
		startTime := float64(i) * segmentDuration
		endTime := startTime + segmentDuration
		if i == segCount-1 {
			endTime = totalDuration
		}

		segments[i] = SubtitleSegment{
			Text:      strings.TrimSpace(sentence),
			StartTime: math.Round(startTime*100) / 100,
			EndTime:   math.Round(endTime*100) / 100,
		}

		idx := i + 1
		startStr := formatSRTTime(startTime)
		endStr := formatSRTTime(endTime)
		fmt.Fprintf(sb, "%d\n%s --> %s\n%s\n\n", idx, startStr, endStr, segments[i].Text)
	}

	return &SubtitleResult{
		SRTContent: sb.String(),
		Segments:   segments,
	}, nil
}

// formatSRTTime converts seconds to SRT timestamp format (HH:MM:SS,mmm).
func formatSRTTime(seconds float64) string {
	d := time.Duration(seconds * float64(time.Second))
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	ms := int(d.Milliseconds()) % 1000
	return fmt.Sprintf("%02d:%02d:%02d,%03d", h, m, s, ms)
}

// splitSentences splits text into sentences at period, exclamation, question mark boundaries.
func splitSentences(text string) []string {
	var sentences []string
	current := strings.Builder{}
	for _, r := range text {
		current.WriteRune(r)
		if r == '.' || r == '!' || r == '?' || r == '。' || r == '！' || r == '？' {
			s := strings.TrimSpace(current.String())
			if s != "" {
				sentences = append(sentences, s)
			}
			current.Reset()
		}
	}
	remaining := strings.TrimSpace(current.String())
	if remaining != "" {
		sentences = append(sentences, remaining)
	}
	return sentences
}

// mergeSentences merges sentences into roughly equal-length groups to fit the duration.
func mergeSentences(sentences []string, totalDuration float64) []string {
	const segmentDuration = 5.0
	targetCount := int(math.Ceil(totalDuration / segmentDuration))
	if targetCount < 1 {
		targetCount = 1
	}
	if targetCount >= len(sentences) {
		return sentences
	}

	groupSize := float64(len(sentences)) / float64(targetCount)
	merged := make([]string, 0, targetCount)

	for i := 0; i < len(sentences); {
		end := i + int(math.Ceil(groupSize))
		if end > len(sentences) {
			end = len(sentences)
		}
		merged = append(merged, strings.Join(sentences[i:end], " "))
		i = end
	}

	return merged
}
