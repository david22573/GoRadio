package audio

import (
	"math"
	"testing"
)

func TestSpectralCentroid(t *testing.T) {
	e := NewExtractor()
	sampleRate := 44100.0

	// Case 1: Silent signal
	silent := make([]float64, 1024)
	scSilent := e.calculateSpectralCentroid(silent, sampleRate)
	if scSilent != 0 {
		t.Errorf("Expected centroid 0 for silent signal, got %f", scSilent)
	}

	// Case 2: Pure sine wave
	// We expect the centroid to be near the frequency of the sine wave
	freq := 1000.0
	sine := make([]float64, 1024)
	for i := range sine {
		sine[i] = math.Sin(2 * math.Pi * freq * float64(i) / sampleRate)
	}
	
	scSine := e.calculateSpectralCentroid(sine, sampleRate)
	// Allow for some binning error (delta)
	if math.Abs(scSine-freq) > 100 {
		t.Errorf("Expected centroid near %f for sine wave, got %f", freq, scSine)
	}
}

func TestEnergyCalculation(t *testing.T) {
	e := NewExtractor()
	
	// Case 1: Zero energy
	silent := make([]float64, 100)
	if e.calculateEnergy(silent) != 0 {
		t.Error("Expected 0 energy for silent signal")
	}

	// Case 2: Known energy
	// For a signal of all 1.0, energy is 1.0
	ones := make([]float64, 100)
	for i := range ones {
		ones[i] = 1.0
	}
	if e.calculateEnergy(ones) != 1.0 {
		t.Error("Expected 1.0 energy for signal of 1s")
	}
}
