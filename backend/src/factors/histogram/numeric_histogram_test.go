package histogram

import (
	"fmt"
	"math"
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
				var allSamples = [][]float64{}

				for j := 0; j < numSamples; j++ {
					var values = []float64{}
					for i := 0; i < dimensions; i++ {
						values = append(values, float64(rand.Intn(100)))
					}
					allSamples = append(allSamples, values)
					h.Add(values)
				}

				count := h.Count()
				assert.InDelta(t, float64(len(allSamples)), count, 0.0001,
					fmt.Sprintf("Count mismatch %v != %v", count, len(allSamples)))

				for _, sample := range allSamples {
					expectedCDF := 1.0
					for i := 0; i < dimensions; i++ {
						expectedCDF *= sample[i] / 100.0
					}
					actualCDF := h.CDF(sample)
					maxDelta := 1.0 * float64(dimensions*dimensions*dimensions) / float64(numSamples)
					// Minimum of 0.15 deviation is always allowed.
					// Test is not expected to pass everytime though.
					maxDelta = math.Max(maxDelta, 0.15)
					assert.InDelta(t, expectedCDF, actualCDF, maxDelta,
						fmt.Sprintf(
							"Mismatch for vec %v, dimensions %d, "+
								"numSamples %d, numBins: %d, actualCDF %.3f, expectedCDF %.3f",
							sample, dimensions, numSamples,
							maxBins, actualCDF, expectedCDF))
				}

				var sum = make([]float64, dimensions)
				for _, sample := range allSamples {
					for i := 0; i < dimensions; i++ {
						sum[i] = sum[i] + sample[i]
					}
				}

				for k, _ := range sum {
					sum[k] = sum[k] / float64(len(allSamples))
				}

				mean := h.Mean()
				// Each dimension here is an independent uniform distribution here
				// with b=100 and a=0.
				// Mean is 50.0
				// Standard deviation is (b-a)/sqrt(12) = 8.33.
				// Actual mean can be expected to be within
				// (numSamples * numDimensions * numDimensions * numDimensions / 80 * maxBins)
				// of standard deviation away from the sample mean.
				// i.e with 100 samples and 2 dimensions and 10 bins within 1 sd.
				// with 1 sample, 10 dimensions and 2 bins within 6.2 sd.
				// A minimum deviation of one 1 sd is always allowed.
				// Can fail occassionally.
				maxDeviation := (8.33 * float64(
					numSamples*dimensions*dimensions*dimensions) / float64(80*maxBins))
				maxDeviation = math.Max(maxDeviation, 8.33)
				for k, _ := range sum {
					assert.InDelta(t, sum[k], mean[k],
						maxDeviation,
						fmt.Sprintf("Mean mismatch %v != %v, numSamples:%d, dimensions:%d, numBins:%d",
							mean, sum, numSamples, dimensions, maxBins))
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
	allSamples := [][]float64{}
	allSampleMaps := []map[string]float64{}

	for j := 0; j < numSamples; {

		switch choice := rand.Intn(5); choice {
		case 0:
			// Add (values)
			var values = []float64{}
			for i := 0; i < dimensions; i++ {
				values = append(values, float64(rand.Intn(100)))
			}
			allSamples = append(allSamples, values)
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
			allSamples = append(allSamples, values)
			allSampleMaps = append(allSampleMaps, keyValues)
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
			allSamples = append(allSamples, values)
			allSampleMaps = append(allSampleMaps, keyValues)
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
	assert.InDelta(t, float64(len(allSamples)), count, 0.0001,
		fmt.Sprintf("Count mismatch %v != %v", count, len(allSamples)))

	for _, sample := range allSampleMaps {
		expectedCDF := 1.0
		for _, v := range sample {
			expectedCDF *= v / 100.0
		}
		actualCDF := h.CDFFromMap(sample)
		assert.InDelta(t, expectedCDF, actualCDF, 0.25,
			fmt.Sprintf(
				"Mismatch for vec %v, dimensions %d, "+
					"numSamples %d, numBins: %d, actualCDF %.3f, expectedCDF %.3f",
				sample, dimensions, numSamples,
				maxBins, actualCDF, expectedCDF))
	}

	var sum = make([]float64, dimensions)
	for _, values := range allSamples {
		for i := 0; i < dimensions; i++ {
			sum[i] = sum[i] + values[i]
		}
	}

	for k, _ := range sum {
		sum[k] = sum[k] / float64(len(allSamples))
	}

	mean := h.Mean()
	// 7% error from exact value with one dimension
	// and 35% for 5 dimensions. Not a stable test.
	// Can fail occassionally.
	allowedMismatchFraction := 0.07 * float64(dimensions)
	for k, _ := range sum {
		assert.InDelta(t, sum[k], mean[k],
			sum[k]*allowedMismatchFraction,
			fmt.Sprintf("Mean mismatch %v != %v", mean, sum))
	}
}
