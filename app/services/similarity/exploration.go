package similarity

import (
	"github.com/david22573/GoRadio/app/audio"
)

type ExplorationConfig struct {
	MinDistance    float64            // exploitation-exploration boundary
	MaxDistance    float64            // exploration-chaos boundary
	FeatureWeights map[string]float64 // weights for specific acoustic features
}

// Default config: 0.0-0.3 exploitation, 0.4-0.7 exploration
func NewDefaultExplorationConfig() ExplorationConfig {
	return ExplorationConfig{
		MinDistance: 0.4,
		MaxDistance: 0.7,
		FeatureWeights: map[string]float64{
			"spectral_centroid": 1.0,
			"tempo":             1.5,
			"energy":            1.0,
			"mfcc":              0.8,
			"chroma":            0.5,
		},
	}
}

// CalculateExplorationZone returns the distance range for finding exploration candidates
func CalculateExplorationZone(config ExplorationConfig) (minDist, maxDist float64) {
	return config.MinDistance, config.MaxDistance
}

// WeightedDistance calculates distance between two feature sets using weights
func WeightedDistance(f1, f2 *audio.AcousticFeatures, config ExplorationConfig) float64 {
	var totalDist float64
	var totalWeight float64

	// Tempo distance (highest weight usually)
	wTempo := config.FeatureWeights["tempo"]
	totalDist += wTempo * mathAbs(f1.Tempo-f2.Tempo) / 200.0
	totalWeight += wTempo

	// Energy distance
	wEnergy := config.FeatureWeights["energy"]
	totalDist += wEnergy * mathAbs(f1.Energy-f2.Energy)
	totalWeight += wEnergy

	// Spectral Centroid
	wSpec := config.FeatureWeights["spectral_centroid"]
	totalDist += wSpec * mathAbs(f1.SpectralCentroid-f2.SpectralCentroid) / 10000.0
	totalWeight += wSpec

	if totalWeight == 0 {
		return 0
	}
	return totalDist / totalWeight
}

func mathAbs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

// IsTransitionSafe ensures the transition between two tracks is not jarring
func IsTransitionSafe(f1, f2 *audio.AcousticFeatures) bool {
	// Tempo jump should not exceed 20%
	tempoDiff := mathAbs(f1.Tempo - f2.Tempo)
	if tempoDiff > (f1.Tempo * 0.2) {
		return false
	}

	// Energy jump should not exceed 0.5 (normalized)
	energyDiff := mathAbs(f1.Energy - f2.Energy)
	if energyDiff > 0.5 {
		return false
	}

	return true
}
