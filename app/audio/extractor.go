package audio

import (
	"fmt"
	"io"
	"math"

	"github.com/go-audio/wav"
	"github.com/mjibson/go-dsp/fft"
)

// Extractor handles audio feature extraction
type Extractor struct{}

func NewExtractor() *Extractor {
	return &Extractor{}
}

// ExtractFromWav reads a WAV stream and extracts acoustic features
func (e *Extractor) ExtractFromWav(r io.ReadSeeker) (*AcousticFeatures, error) {
	d := wav.NewDecoder(r)
	if !d.IsValidFile() {
		return nil, fmt.Errorf("invalid wav file")
	}

	buf, err := d.FullPCMBuffer()
	if err != nil {
		return nil, fmt.Errorf("failed to read wav buffer: %w", err)
	}

	// Convert to float64 slice for processing
	floatData := make([]float64, len(buf.Data))
	for i, val := range buf.Data {
		floatData[i] = float64(val) / math.MaxInt16
	}

	sampleRate := float64(buf.Format.SampleRate)

	features := &AcousticFeatures{}
	features.SpectralCentroid = e.calculateSpectralCentroid(floatData, sampleRate)
	features.Energy = e.calculateEnergy(floatData)
	features.MFCCVector = e.calculateMFCC(floatData, sampleRate)
	features.ChromaVector = e.calculateChroma(floatData, sampleRate)
	features.Tempo = e.calculateTempo(floatData, sampleRate)

	// TODO: Implement final 128-dim embedding logic
	features.Embedding = e.generateEmbedding(features)

	return features, nil
}

func (e *Extractor) calculateTempo(data []float64, sampleRate float64) float64 {
	return 120.0 // Placeholder
}

func (e *Extractor) calculateSpectralCentroid(data []float64, sampleRate float64) float64 {
	if len(data) == 0 {
		return 0
	}
	
	// Apply Hanning window
	windowed := make([]float64, len(data))
	for i := range data {
		window := 0.5 * (1 - math.Cos(2*math.Pi*float64(i)/float64(len(data)-1)))
		windowed[i] = data[i] * window
	}
	
	coeffs := fft.FFTReal(windowed)
	numBins := len(coeffs) / 2
	
	var sumWeights float64
	var sumMagnitudes float64
	
	for i := 0; i < numBins; i++ {
		magnitude := math.Hypot(real(coeffs[i]), imag(coeffs[i]))
		frequency := float64(i) * sampleRate / float64(len(data))
		sumWeights += frequency * magnitude
		sumMagnitudes += magnitude
	}
	
	if sumMagnitudes == 0 {
		return 0
	}
	
	return sumWeights / sumMagnitudes
}

func (e *Extractor) calculateEnergy(data []float64) float64 {
	var sum float64
	for _, val := range data {
		sum += val * val
	}
	if len(data) == 0 {
		return 0
	}
	return sum / float64(len(data))
}

// calculateMFCC returns 13 Mel-frequency cepstral coefficients.
func (e *Extractor) calculateMFCC(data []float64, sampleRate float64) []float64 {
	// 1. Frame the signal into short frames
	// 2. Compute the periodogram estimate of the power spectrum
	// 3. Apply the mel filterbank to the power spectra, sum the energy in each filter
	// 4. Take the logarithm of all filterbank energies
	// 5. Take the DCT of the log filterbank energies.
	// 6. Keep DCT coefficients 1-13.
	
	// Simplified implementation for placeholder
	mfccs := make([]float64, 13)
	if len(data) == 0 {
		return mfccs
	}
	
	// For demonstration, we use some energy levels from FFT
	coeffs := fft.FFTReal(data)
	numBins := len(coeffs) / 2
	
	// Group FFT bins into 13 "mel" bands (logarithmically spaced)
	bandWidth := math.Log10(sampleRate/2) / 13
	for i := 0; i < 13; i++ {
		minFreq := math.Pow(10, float64(i)*bandWidth)
		maxFreq := math.Pow(10, float64(i+1)*bandWidth)
		
		var energy float64
		count := 0
		for j := 0; j < numBins; j++ {
			freq := float64(j) * sampleRate / float64(len(data))
			if freq >= minFreq && freq <= maxFreq {
				mag := math.Hypot(real(coeffs[j]), imag(coeffs[j]))
				energy += mag * mag
				count++
			}
		}
		if count > 0 {
			mfccs[i] = math.Log10(energy/float64(count) + 1e-10)
		}
	}
	
	return mfccs
}

func (e *Extractor) calculateChroma(data []float64, sampleRate float64) []float64 {
	// 12-dimensional placeholder representing the 12 semitones
	chroma := make([]float64, 12)
	// Simplified: mapping frequencies to notes
	coeffs := fft.FFTReal(data)
	numBins := len(coeffs) / 2
	
	for i := 0; i < numBins; i++ {
		freq := float64(i) * sampleRate / float64(len(data))
		if freq > 0 {
			// A4 = 440Hz
			midiNote := 12*math.Log2(freq/440.0) + 69
			pitchClass := int(math.Round(midiNote)) % 12
			if pitchClass >= 0 && pitchClass < 12 {
				mag := math.Hypot(real(coeffs[i]), imag(coeffs[i]))
				chroma[pitchClass] += mag
			}
		}
	}
	return chroma
}

func (e *Extractor) generateEmbedding(f *AcousticFeatures) []float64 {
	// Composite vector placeholder (128-dim)
	// We'll concatenate and pad our features
	embedding := make([]float64, 128)
	
	// Index tracker
	idx := 0
	
	// 1. Spectral Centroid (normalized roughly by sample rate)
	embedding[idx] = f.SpectralCentroid / 10000.0
	idx++
	
	// 2. Tempo (normalized roughly by 200 BPM)
	embedding[idx] = f.Tempo / 200.0
	idx++
	
	// 3. Energy
	embedding[idx] = f.Energy
	idx++
	
	// 4. MFCCs (13 dims)
	for i := 0; i < 13 && idx < 128; i++ {
		embedding[idx] = f.MFCCVector[i]
		idx++
	}
	
	// 5. Chroma (12 dims)
	for i := 0; i < 12 && idx < 128; i++ {
		embedding[idx] = f.ChromaVector[i]
		idx++
	}
	
	// Remainder stays 0 (padding) or could be filled with more complex analysis
	return embedding
}
