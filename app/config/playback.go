package config

type PlaybackConfig struct {
	// Exploitation
	ExploitationKNN     int     // k_neighbors: 10
	SimilarityThreshold float64 // similarity_threshold: 0.7

	// Exploration
	BaseExplorationRate    float64 // base_rate: 0.15
	MinExplorationRate     float64 // min_rate: 0.10
	MaxExplorationRate     float64 // max_rate: 0.20
	ExplorationMinDistance float64 // min_distance: 0.4
	ExplorationMaxDistance float64 // max_distance: 0.7
	ControlledRatio        float64 // controlled_ratio: 0.5

	// Session
	VectorLearningRate  float64 // vector_learning_rate: 0.2
	VectorRepulsionRate float64 // Repulsion alpha
	SkipThreshold       float64 // skip_threshold: 0.3
	ListenThreshold     float64 // listen_threshold: 0.8
	HistorySize         int     // history_size: 50
	SessionTimeout      int     // timeout: 3600

	// Queue
	QueuePrefetchSize int // prefetch_size: 3
	RepeatPrevention  int // repeat_prevention: 20

	// Audio
	AnalysisDuration  int // analysis_duration: 30
	FeatureDimensions int // feature_dimensions: 128
	MFCCCoefficients  int // mfcc_coefficients: 13
	ChromaBins        int // chroma_bins: 12
}

func DefaultPlaybackConfig() PlaybackConfig {
	return PlaybackConfig{
		ExploitationKNN:        10,
		SimilarityThreshold:    0.7,
		BaseExplorationRate:    0.15,
		MinExplorationRate:     0.10,
		MaxExplorationRate:     0.20,
		ExplorationMinDistance: 0.4,
		ExplorationMaxDistance: 0.7,
		ControlledRatio:        0.5,
		VectorLearningRate:     0.2,
		VectorRepulsionRate:    -0.1,
		SkipThreshold:          0.3,
		ListenThreshold:        0.8,
		HistorySize:            50,
		SessionTimeout:         3600,
		QueuePrefetchSize:      3,
		RepeatPrevention:       20,
		AnalysisDuration:       30,
		FeatureDimensions:      128,
		MFCCCoefficients:       13,
		ChromaBins:             12,
	}
}
