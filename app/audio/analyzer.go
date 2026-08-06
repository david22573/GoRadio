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

func (a *Analyzer) AnalyzeStream(ctx context.Context, source io.Reader) (*AcousticFeatures, error) {
	tmpFile, err := os.CreateTemp("", "stream-*.wav")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	_, err = io.Copy(tmpFile, io.LimitReader(source, 2646000))
	if err != nil && err != io.EOF {
		return nil, err
	}

	if _, err := tmpFile.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	return a.extractor.ExtractFromWav(tmpFile)
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
