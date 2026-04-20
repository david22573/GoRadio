package audio

import (
	"context"
	"io"
	"os"
)

// Analyzer coordinates feature extraction tasks
type Analyzer struct {
	extractor *Extractor
}

func NewAnalyzer() *Analyzer {
	return &Analyzer{
		extractor: NewExtractor(),
	}
}

// AnalyzeStream samples a 30-second snippet from an audio source
func (a *Analyzer) AnalyzeStream(ctx context.Context, source io.Reader) (*AcousticFeatures, error) {
	// For now, assume we get a stream that we can sample
	// In a real implementation, we would use a library to decode different formats (mp3, etc)
	// and take a 30s snippet.
	// For this task, we'll focus on the WAV extraction part.
	return nil, nil
}

// AnalyzeFile extracts features from a local audio file
func (a *Analyzer) AnalyzeFile(path string) (*AcousticFeatures, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return a.extractor.ExtractFromWav(f)
}
