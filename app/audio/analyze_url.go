package audio

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// AnalyzeURL samples audio from rawURL (resolving YouTube/SoundCloud links to
// their direct stream URL via yt-dlp when needed) and extracts acoustic
// features. It is shared by the batch analyzer and the seed-time ingestion path.
// The caller is responsible for providing a context with a timeout.
func (a *Analyzer) AnalyzeURL(ctx context.Context, rawURL string, seconds int) (*AcousticFeatures, error) {
	resolved, err := resolveStreamURL(ctx, rawURL)
	if err != nil {
		return nil, err
	}

	tmp, err := os.CreateTemp("", "goradio-track-*.wav")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	if err := sampleStream(ctx, resolved, tmp.Name(), seconds); err != nil {
		return nil, err
	}

	return a.AnalyzeFile(tmp.Name())
}

// resolveStreamURL returns a direct stream URL for rawURL, resolving YouTube and
// SoundCloud links via yt-dlp. Non-streaming URLs are returned unchanged.
func resolveStreamURL(ctx context.Context, rawURL string) (string, error) {
	needsResolve := strings.Contains(rawURL, "youtube.com") ||
		strings.Contains(rawURL, "youtu.be") ||
		strings.Contains(rawURL, "soundcloud.com")
	if !needsResolve {
		return rawURL, nil
	}

	cmd := exec.CommandContext(ctx, "yt-dlp", "-g", "-f", "bestaudio", rawURL)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("yt-dlp resolve failed: %w (output: %s)", err, string(out))
	}
	resolved := strings.TrimSpace(string(out))
	if resolved == "" {
		return "", fmt.Errorf("yt-dlp returned empty stream URL")
	}
	return resolved, nil
}

// sampleStream uses ffmpeg to capture seconds of audio from a stream and save it
// as a mono 44.1kHz WAV file at outPath.
func sampleStream(ctx context.Context, url, outPath string, seconds int) error {
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-y", // Overwrite
		"-i", url,
		"-t", fmt.Sprintf("%d", seconds),
		"-acodec", "pcm_s16le",
		"-ar", "44100",
		"-ac", "1",
		outPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg error: %w (output: %s)", err, string(output))
	}

	return nil
}
