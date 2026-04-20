package queue

import (
	"math"
	"time"

	"github.com/david22573/GoRadio/app/config"
	"github.com/david22573/GoRadio/app/session"
)

type ExplorationScheduler struct {
	BaseRate float64
	MinRate  float64
	MaxRate  float64
}

func NewExplorationScheduler(cfg config.PlaybackConfig) *ExplorationScheduler {
	return &ExplorationScheduler{
		BaseRate: cfg.BaseExplorationRate,
		MinRate:  cfg.MinExplorationRate,
		MaxRate:  cfg.MaxExplorationRate,
	}
}

func (es *ExplorationScheduler) CalculateRate(s *session.SessionState) float64 {
	metrics := s.CalculateMetrics()
	rate := es.BaseRate

	// Skip rate adjustments
	if metrics.SkipRate > 0.3 {
		rate = math.Min(rate*1.5, es.MaxRate)
	} else if metrics.SkipRate < 0.1 {
		rate = math.Max(rate*0.7, es.MinRate)
	}

	// Time-of-day variation: more exploration in the evening (18:00 - 24:00)
	now := time.Now()
	hour := now.Hour()
	if hour >= 18 || hour < 6 {
		rate = math.Min(rate*1.2, es.MaxRate)
	}

	return rate
}
