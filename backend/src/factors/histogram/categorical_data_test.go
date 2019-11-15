package histogram

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func buildCategoricalHistogramFromData(t *testing.T, maxBins int, dimensions int, data []string) CategoricalHistogram {
	h, _ := NewCategoricalHistogram(maxBins, dimensions, nil)

	for _, valStr := range data {
		val := strings.Split(valStr, ",")
		assert.Equal(t, dimensions, len(val),
			fmt.Sprintf("Data not matching dimension %d %s", dimensions, valStr))
		h.Add(val)
	}
	return h
}

func TestCategoricalSampleData(t *testing.T) {
	for _, data := range [][]string{
		categoricalDataSamples10, categoricalDataSamples100, categoricalDataSamples1000, categoricalDataSamples10000} {

		dimensions := len(strings.Split(data[0], ","))
		numDataSamples := len(data)

		// Evaluate with different bin sizes.
		for _, numBins := range []int{1, 2, 4, 8, 16, 32} {
			fmt.Println("------------------------------------------------------------")
			fmt.Println(fmt.Sprintf("EVALUATING FOR DIMENSIONS:%d  BINS:%d NUM_SAMPLES:%d",
				dimensions, numBins, numDataSamples))
			fmt.Println("------------------------------------------------------------")

			hist := buildCategoricalHistogramFromData(t, numBins, dimensions, data)
			assert.Equal(t, uint64(numDataSamples), hist.Count(),
				"Mismatch in number of samples.")

			// Empirically determined upper bound for distance between Histogram / Sample PDF
			// and actual generating PDF.
			// With 8 bins and 1 dimensions the maximum error is 0.093
			// With 128 bins and 1 dimensions the maximum error is 0.04
			// With 8 bins and 5 dimensions the maximum error is 0.47
			// With 128 bins and 5 dimensions the maximum error is 0.2
			maxKLDistance := 7500.0 / float64(numDataSamples) / float64(numBins)
			KLHistogramActual := 0.0
			KLSampleActual := 0.0

			sampleFrequencyMap := make(map[string]int)
			for k := range data {
				if count, ok := sampleFrequencyMap[data[k]]; ok {
					sampleFrequencyMap[data[k]] = count + 1
				} else {
					sampleFrequencyMap[data[k]] = 1
				}
			}

			for key, actualProb := range getSamplePdf() {
				evalPoint := strings.Split(key, ",")

				histProb, err := hist.PDF(evalPoint)
				assert.Nil(t, err)
				KLHistogramActual += histProb * log(histProb/actualProb) / log(2)

				sampleProb := 0.0
				if count, ok := sampleFrequencyMap[key]; ok {
					sampleProb = float64(count) / float64(numDataSamples)
				}
				KLSampleActual += sampleProb * log(sampleProb/actualProb) / log(2)

				fmt.Println(fmt.Sprintf("Probabilities at %s - Actual: %.2f, Sample: %.2f, Histogram %.2f",
					key, actualProb, sampleProb, histProb))
			}
			fmt.Println(fmt.Sprintf(
				"\nHIST_KL_DISTANCE:%.3f bits, SAMPLE_KL_DISTANCE:%.3f bits",
				KLHistogramActual, KLSampleActual))
			assert.True(t, KLHistogramActual < maxKLDistance,
				fmt.Sprintf("KLHistogramActual: %.3f, maxKLDistance: %.3f",
					KLHistogramActual, maxKLDistance))
			assert.True(t, KLSampleActual < maxKLDistance,
				fmt.Sprintf("KLSampleActual: %.3f, maxKLDistance: %.3f",
					KLSampleActual, maxKLDistance))

			fmt.Println("---------------------------------------------------")
		}
	}
}

func TestCategoricalTrimByBinSize(t *testing.T) {
	data := categoricalDataSamples1000
	dimensions := len(strings.Split(data[0], ","))
	numDataSamples := len(data)
	// First build histogram with 32 bins.
	numBins := 32

	fmt.Println("------------------------------------------------------------")
	fmt.Println(fmt.Sprintf("EVALUATING FOR DIMENSIONS:%d  BINS:%d NUM_SAMPLES:%d",
		dimensions, numBins, numDataSamples))
	fmt.Println("------------------------------------------------------------")

	hist := buildCategoricalHistogramFromData(t, numBins, dimensions, data)
	assert.Equal(t, uint64(numDataSamples), hist.Count(),
		"Mismatch in number of samples.")
	assert.Equal(t, hist.numBins(), 32, "Mismatch in number of bins.")

	// Check KLDistance before trimming.
	maxKLDistance := 6500.0 / float64(numDataSamples) / float64(numBins)
	KLHistogram := 0.0
	for key, actualProb := range getSamplePdf() {
		evalPoint := strings.Split(key, ",")

		histProb, err := hist.PDF(evalPoint)
		assert.Nil(t, err)
		KLHistogram += histProb * log(histProb/actualProb) / log(2)

		fmt.Println(fmt.Sprintf("Probabilities at %s - Actual: %.2f, Histogram %.2f",
			key, actualProb, histProb))
	}
	fmt.Println(fmt.Sprintf(
		"\nHIST_KL_DISTANCE:%.3f bits", KLHistogram))
	assert.True(t, KLHistogram < maxKLDistance,
		fmt.Sprintf("KLHistogramActual: %.3f, maxKLDistance: %.3f",
			KLHistogram, maxKLDistance))
	fmt.Println("---------------------------------------------------")

	// Trim histogram.
	hist.TrimByBinSize(0.5)
	assert.Equal(t, uint64(numDataSamples), hist.Count(),
		"Mismatch in number of samples.")
	assert.Equal(t, hist.numBins(), 16, "Mismatch in number of bins.")
	// Check KLDistance after trimming.
	KLHistogram = 0.0
	for key, actualProb := range getSamplePdf() {
		evalPoint := strings.Split(key, ",")

		histProb, err := hist.PDF(evalPoint)
		assert.Nil(t, err)
		KLHistogram += histProb * log(histProb/actualProb) / log(2)

		fmt.Println(fmt.Sprintf("Probabilities at %s - Actual: %.2f, Histogram %.2f",
			key, actualProb, histProb))
	}
	fmt.Println(fmt.Sprintf(
		"\nHIST_KL_DISTANCE:%.3f bits", KLHistogram))
	assert.True(t, KLHistogram < maxKLDistance,
		fmt.Sprintf("KLHistogramActual: %.3f, maxKLDistance: %.3f",
			KLHistogram, maxKLDistance))
	fmt.Println("---------------------------------------------------")

	// Minimum 1 bin.
	hist.TrimByBinSize(0.01)
	assert.Equal(t, uint64(numDataSamples), hist.Count(),
		"Mismatch in number of samples.")
	assert.Equal(t, hist.numBins(), 1, "Mismatch in number of bins.")
}

func TestCategoricalTrimByFmapSize(t *testing.T) {
	data := categoricalDataSamples1000
	dimensions := len(strings.Split(data[0], ","))
	numDataSamples := len(data)
	// First build histogram with 32 bins.
	numBins := 32

	fmt.Println("------------------------------------------------------------")
	fmt.Println(fmt.Sprintf("EVALUATING FOR DIMENSIONS:%d  BINS:%d NUM_SAMPLES:%d",
		dimensions, numBins, numDataSamples))
	fmt.Println("------------------------------------------------------------")

	hist := buildCategoricalHistogramFromData(t, numBins, dimensions, data)
	assert.Equal(t, uint64(numDataSamples), hist.Count(),
		"Mismatch in number of samples.")
	assert.Equal(t, hist.numBins(), 32, "Mismatch in number of bins.")
	assert.Equal(t, hist.maxFmapSize(), int(fMAP_MAX_SIZE), "Mismatch in fmap size.")

	// Check KLDistance before trimming.
	maxKLDistance := 6500.0 / float64(numDataSamples) / float64(numBins)
	KLHistogram := 0.0
	for key, actualProb := range getSamplePdf() {
		evalPoint := strings.Split(key, ",")

		histProb, err := hist.PDF(evalPoint)
		assert.Nil(t, err)
		KLHistogram += histProb * log(histProb/actualProb) / log(2)

		fmt.Println(fmt.Sprintf("Probabilities at %s - Actual: %.2f, Histogram %.2f",
			key, actualProb, histProb))
	}
	fmt.Println(fmt.Sprintf(
		"\nHIST_KL_DISTANCE:%.3f bits", KLHistogram))
	assert.True(t, KLHistogram < maxKLDistance,
		fmt.Sprintf("KLHistogramActual: %.3f, maxKLDistance: %.3f",
			KLHistogram, maxKLDistance))
	fmt.Println("---------------------------------------------------")

	// Trim histogram.
	hist.TrimByFmapSize(0.5)
	assert.Equal(t, uint64(numDataSamples), hist.Count(),
		"Mismatch in number of samples.")
	assert.Equal(t, hist.numBins(), 32, "Mismatch in number of bins.")
	assert.Equal(t, hist.maxFmapSize(), int(fMAP_MAX_SIZE*0.5), "Mismatch in fmap size.")

	// Check KLDistance after trimming.
	KLHistogram = 0.0
	for key, actualProb := range getSamplePdf() {
		evalPoint := strings.Split(key, ",")

		histProb, err := hist.PDF(evalPoint)
		assert.Nil(t, err)
		KLHistogram += histProb * log(histProb/actualProb) / log(2)

		fmt.Println(fmt.Sprintf("Probabilities at %s - Actual: %.2f, Histogram %.2f",
			key, actualProb, histProb))
	}
	fmt.Println(fmt.Sprintf(
		"\nHIST_KL_DISTANCE:%.3f bits", KLHistogram))
	assert.True(t, KLHistogram < maxKLDistance,
		fmt.Sprintf("KLHistogramActual: %.3f, maxKLDistance: %.3f",
			KLHistogram, maxKLDistance))
	fmt.Println("---------------------------------------------------")

	// Minimum size is maintained.
	hist.TrimByFmapSize(0.01)
	assert.Equal(t, uint64(numDataSamples), hist.Count(),
		"Mismatch in number of samples.")
	assert.Equal(t, hist.numBins(), 32, "Mismatch in number of bins.")
	assert.Equal(t, hist.maxFmapSize(), int(fMAP_MIN_SIZE), "Mismatch in fmap size.")
}

// Use this test case to generate sample data.
/*
func TestCategoricalDataGeneration(t *testing.T) {
	data, err := generateSampleCategoricalData(10000)
	assert.Nil(t, err)
	assert.NotNil(t, data)
	for i := range data {
		fmt.Println(fmt.Sprintf("\"%s\",", data[i]))
	}
}
*/

func BenchmarkNewCategoricalHistogram(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		maxBins := 8
		dimensions := len(strings.Split(categoricalDataSamples10000[0], ","))
		h, _ := NewCategoricalHistogram(maxBins, dimensions, nil)
		h.Add(categoricalDataSamples10000)
	}
}
