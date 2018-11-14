package histogram

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHistogramMean(t *testing.T) {
	for _, maxBins := range []int{2, 3, 4, 5, 10} {
		for _, dimensions := range []int{2, 3, 4, 5, 10} {
			for _, numSamples := range []int{1, 5, 10, 50, 100} {
				// Exact Mean calculations with combinations of
				// number of bins, dimensions and number of samples.

				h, _ := NewNumericHistogram(maxBins, dimensions, nil)
				var sample = [][]float64{}

				for j := 0; j < numSamples; j++ {
					var values = []float64{}
					for i := 0; i < dimensions; i++ {
						values = append(values, float64(rand.Intn(100)))
					}
					sample = append(sample, values)
					h.Add(values)
				}

				count := h.Count()
				assert.InDelta(t, float64(len(sample)), count, 0.0001,
					fmt.Sprintf("Count mismatch %v != %v", count, len(sample)))

				var sum = make([]float64, dimensions)
				for _, values := range sample {
					for i := 0; i < dimensions; i++ {
						sum[i] = sum[i] + values[i]
					}
				}

				for k, _ := range sum {
					sum[k] = sum[k] / float64(len(sample))
				}

				mean := h.Mean()
				// 6% error from exact value with one dimension
				// and 30% for 5 dimensions. Not a stable test.
				// Can fail occassionally.
				allowedMismatchFraction := 0.06 * float64(dimensions)
				for k, _ := range sum {
					assert.InDelta(t, sum[k], mean[k], 
						sum[k] * allowedMismatchFraction,
						fmt.Sprintf("Mean mismatch %v != %v", mean, sum))
				}
			}
		}
	}
}

func TestNumericHistogramAddWithTemplate(t *testing.T) {
	maxBins := 10
	dimensions := 4
	numSamples := 100
	template := NumericHistogramTemplate{
		NumericHistogramTemplateUnit{
			Name:       "dim1",
			IsRequired: true,
		},
		NumericHistogramTemplateUnit{
			Name:       "dim2",
			IsRequired: false,
			Default:    0.0,
		},
		NumericHistogramTemplateUnit{
			Name:       "dim3",
			IsRequired: true,
		},
		NumericHistogramTemplateUnit{
			Name:       "dim4",
			IsRequired: false,
			Default:    1.0,
		},
	}

	h, _ := NewNumericHistogram(maxBins, dimensions, &template)
	var samples = [][]float64{}

	for j := 0; j < numSamples; {

		switch choice := rand.Intn(5); choice {
		case 0:
			// Add (values)
			var values = []float64{}
			for i := 0; i < dimensions; i++ {
				values = append(values, float64(rand.Intn(100)))
			}
			samples = append(samples, values)
			err := h.Add(values)
			assert.Nil(t, err)
			j++
		case 1:
			// AddMap all 4 dimensions.
			keyValues := map[string]float64{}
			keyValues["dim1"] = float64(rand.Intn(100))
			keyValues["dim2"] = float64(rand.Intn(100))
			keyValues["dim3"] = float64(rand.Intn(100))
			keyValues["dim4"] = float64(rand.Intn(100))
			err := h.AddMap(keyValues)
			assert.Nil(t, err)
			values := []float64{keyValues["dim1"], keyValues["dim2"],
				keyValues["dim3"], keyValues["dim4"]}
			samples = append(samples, values)
			j++
		case 2:
			// AddMap with default dimension.
			// dim1 and dim3 are required. dim2 and dim4 uses default values.
			keyValues := map[string]float64{}
			keyValues["dim1"] = float64(rand.Intn(100))
			keyValues["dim3"] = float64(rand.Intn(100))
			err := h.AddMap(keyValues)
			assert.Nil(t, err)
			values := []float64{keyValues["dim1"], template[1].Default,
				keyValues["dim3"], template[3].Default}
			samples = append(samples, values)
			j++
		case 3:
			// AddMap error missing required dimension.
			// dim3 is missing
			// Sample is not recorded since it is not added to histogram.
			keyValues := map[string]float64{}
			keyValues["dim1"] = float64(rand.Intn(100))
			keyValues["dim2"] = float64(rand.Intn(100))
			keyValues["dim4"] = float64(rand.Intn(100))
			err := h.AddMap(keyValues)
			assert.NotNil(t, err)
		case 4:
			// AddMap error with extra unknown keys.
			keyValues := map[string]float64{}
			keyValues["dim1"] = float64(rand.Intn(100))
			keyValues["dim2"] = float64(rand.Intn(100))
			keyValues["dim3"] = float64(rand.Intn(100))
			keyValues["dim4"] = float64(rand.Intn(100))
			keyValues["dim5"] = float64(rand.Intn(100))
			err := h.AddMap(keyValues)
			assert.NotNil(t, err)
		}

	}

	count := h.Count()
	assert.InDelta(t, float64(len(samples)), count, 0.0001,
		fmt.Sprintf("Count mismatch %v != %v", count, len(samples)))

	var sum = make([]float64, dimensions)
	for _, values := range samples {
		for i := 0; i < dimensions; i++ {
			sum[i] = sum[i] + values[i]
		}
	}

	for k, _ := range sum {
		sum[k] = sum[k] / float64(len(samples))
	}

	mean := h.Mean()
	// 6% error from exact value with one dimension
	// and 30% for 5 dimensions. Not a stable test.
	// Can fail occassionally.
	allowedMismatchFraction := 0.06 * float64(dimensions)
	for k, _ := range sum {
		assert.InDelta(t, sum[k], mean[k], 
			sum[k] * allowedMismatchFraction,
			fmt.Sprintf("Mean mismatch %v != %v", mean, sum))
	}
}