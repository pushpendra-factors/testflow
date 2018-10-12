package histogram

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHistogramMeanAndVariance(t *testing.T) {
	for _, maxBins := range []int{2, 3, 4, 5, 10} {
		for _, dimensions := range []int{2, 3, 4, 5, 10} {
			for _, numSamples := range []int{1, 5, 10, 50, 100} {
				// Exact Mean and Variance calculations with combinations of
				// number of bins, dimensions and number of samples.

				h := NewNumericHistogram(maxBins, dimensions)
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
				for k, _ := range sum {
					assert.InDelta(t, sum[k], mean[k], 0.0001,
						fmt.Sprintf("Mean mismatch %v != %v", mean, sum))
				}

				var sumsquare = make([]float64, dimensions)
				for _, values := range sample {
					for i := 0; i < dimensions; i++ {
						sumsquare[i] = sumsquare[i] + (values[i]-sum[i])*(values[i]-sum[i])
					}
				}

				for k, _ := range sumsquare {
					sumsquare[k] = sumsquare[k] / float64(len(sample))
				}

				variance := h.Variance()
				for k, _ := range sumsquare {
					assert.InDelta(t, sumsquare[k], variance[k], 0.0001,
						fmt.Sprintf("Variance mismatch %v != %v", variance, sumsquare))
				}
			}
		}
	}
}
