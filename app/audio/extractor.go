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
	if len(data) == 0 {
		return 120.0
	}

	windowSize := int(sampleRate * 0.05) // 50ms window
	stepSize := int(sampleRate * 0.025)  // 25ms step
	if windowSize == 0 || stepSize == 0 {
		return 120.0
	}

	var energies []float64
	for i := 0; i+windowSize <= len(data); i += stepSize {
		var energy float64
		for j := 0; j < windowSize; j++ {
			energy += data[i+j] * data[i+j]
		}
		energies = append(energies, energy)
	}

	if len(energies) == 0 {
		return 120.0
	}

	var sumEnergy float64
	for _, e := range energies {
		sumEnergy += e
	}
	meanEnergy := sumEnergy / float64(len(energies))

	var sumSq float64
	for _, e := range energies {
		sumSq += (e - meanEnergy) * (e - meanEnergy)
	}
	stdEnergy := math.Sqrt(sumSq / float64(len(energies)))
	threshold := meanEnergy + 1.5*stdEnergy

	var peakIndices []int
	for i := 1; i < len(energies)-1; i++ {
		if energies[i] > threshold && energies[i] > energies[i-1] && energies[i] > energies[i+1] {
			peakIndices = append(peakIndices, i)
		}
	}

	if len(peakIndices) < 2 {
		return 120.0
	}

	var sumDiff float64
	for i := 1; i < len(peakIndices); i++ {
		sumDiff += float64(peakIndices[i] - peakIndices[i-1])
	}
	avgDiffFrames := sumDiff / float64(len(peakIndices)-1)

	timePerFrame := float64(stepSize) / sampleRate
	avgDiffSecs := avgDiffFrames * timePerFrame

	if avgDiffSecs > 0 {
		bpm := 60.0 / avgDiffSecs
		for bpm < 60 && bpm > 0 {
			bpm *= 2
		}
		for bpm > 240 {
			bpm /= 2
		}
		return bpm
	}

	return 120.0
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
	mfccs := make([]float64, 13)
	if len(data) == 0 {
		return mfccs
	}
	
	coeffs := fft.FFTReal(data)
	numBins := len(coeffs) / 2
	
	powerSpec := make([]float64, numBins)
	for i := 0; i < numBins; i++ {
		mag := math.Hypot(real(coeffs[i]), imag(coeffs[i]))
		powerSpec[i] = (mag * mag) / float64(len(data))
	}
	
	numFilters := 13
	minMel := 2595.0 * math.Log10(1.0)
	maxMel := 2595.0 * math.Log10(1.0+(sampleRate/2.0)/700.0)
	
	melPoints := make([]float64, numFilters+2)
	stepMel := (maxMel - minMel) / float64(numFilters+1)
	for i := 0; i < numFilters+2; i++ {
		melPoints[i] = minMel + float64(i)*stepMel
	}
	
	binPoints := make([]int, numFilters+2)
	for i := 0; i < numFilters+2; i++ {
		hz := 700.0 * (math.Pow(10, melPoints[i]/2595.0) - 1.0)
		bin := int(math.Floor(float64(len(data)) * hz / sampleRate))
		if bin >= numBins {
			bin = numBins - 1
		}
		if bin < 0 {
			bin = 0
		}
		binPoints[i] = bin
	}
	
	filterBankEnergies := make([]float64, numFilters)
	for i := 0; i < numFilters; i++ {
		var energy float64
		for j := binPoints[i]; j < binPoints[i+1]; j++ {
			if binPoints[i+1]-binPoints[i] > 0 {
				weight := float64(j-binPoints[i]) / float64(binPoints[i+1]-binPoints[i])
				energy += weight * powerSpec[j]
			}
		}
		for j := binPoints[i+1]; j < binPoints[i+2]; j++ {
			if binPoints[i+2]-binPoints[i+1] > 0 {
				weight := float64(binPoints[i+2]-j) / float64(binPoints[i+2]-binPoints[i+1])
				energy += weight * powerSpec[j]
			}
		}
		if energy > 0 {
			filterBankEnergies[i] = math.Log(energy)
		} else {
			filterBankEnergies[i] = math.Log(1e-10)
		}
	}
	
	for i := 0; i < 13; i++ {
		var sum float64
		for j := 0; j < numFilters; j++ {
			sum += filterBankEnergies[j] * math.Cos(math.Pi*float64(i)*(float64(j)+0.5)/float64(numFilters))
		}
		mfccs[i] = sum
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
	
	// Normalize the final vector (L2 norm = 1)
	var sumSq float64
	for _, val := range embedding {
		sumSq += val * val
	}
	
	magnitude := math.Sqrt(sumSq)
	if magnitude > 0 {
		for i := range embedding {
			embedding[i] /= magnitude
		}
	}
	
	return embedding
}
