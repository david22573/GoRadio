package config

import (
	"os"
)

type FeatureFlags struct {
	ContinuousPlayback bool
	Exploration        bool
	VectorSearch       bool
	SessionTracking    bool
}

func LoadFeatureFlags() FeatureFlags {
	return FeatureFlags{
		ContinuousPlayback: getEnvBool("ENABLE_CONTINUOUS_PLAYBACK", true),
		Exploration:        getEnvBool("ENABLE_EXPLORATION", true),
		VectorSearch:       getEnvBool("ENABLE_VECTOR_SEARCH", true),
		SessionTracking:    getEnvBool("ENABLE_SESSION_TRACKING", true),
	}
}

func getEnvBool(key string, fallback bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val == "true"
}
