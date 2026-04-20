package audio

type AcousticFeatures struct {
	TrackID          uint
	SpectralCentroid float64
	Tempo            float64
	Energy           float64
	MFCCVector       []float64 // 13-dimensional
	ChromaVector     []float64 // 12-dimensional
	Embedding        []float64 // Final 128-dim vector
}
