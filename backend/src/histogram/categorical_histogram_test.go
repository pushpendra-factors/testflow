package histogram

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHistogramFrequencyMaps(t *testing.T) {
	for _, dimensions := range []int{2, 3, 4, 5, 10} {
		for _, maxBins := range []int{2, 3, 4, 5, 10} {
			for _, numSamples := range []int{1, 5, 10, 50, 100} {
				// Each dimension can take upto 1 to 5 values.
				numAllowedValues := make([]int, dimensions)
				for i := 0; i < dimensions; i++ {
					numAllowedValues[i] = rand.Intn(5) + 1
				}

				// Track and check exact frequencies.
				h := NewCategoricalHistogram(maxBins, dimensions)
				var frequencies = make(map[string]int)
				for j := 0; j < numSamples; j++ {
					var sample = []string{}
					for i := 0; i < dimensions; i++ {
						// Samples look like
						// 11, 12, ""    -> Dimension 1, numAllowedValues 2
						// 21, 22, 23  -> Dimension 2, numAllowedValues 3 and so on.
						// Empty String means missing value.
						v := ""
						if rand.Intn(10) > 1 {
							// 20% chance of a symbol being an empty string.
							v = fmt.Sprintf("%d%d", i+1, rand.Intn(numAllowedValues[i])+1)
							if count, ok := frequencies[v]; ok {
								frequencies[v] = count + 1
							} else {
								frequencies[v] = 1
							}
						}
						sample = append(sample, v)

					}
					h.Add(sample)
				}

				assert.InDelta(t, numSamples, h.Count(), 0.0001,
					fmt.Sprintf("Count mismatch %v != %v", h.Count(), numSamples))
				assert.InDelta(t, numSamples, h.totalBinCount(), 0.0001,
					fmt.Sprintf("Bin Total Count mismatch %v != %v", h.totalBinCount(), numSamples))

				for k, expectedCount := range frequencies {
					actualCount := h.frequency(k)
					assert.InDelta(t, expectedCount, actualCount, 0.0001,
						fmt.Sprintf("Count mismatch for symbol %s %v != %v",
							k, expectedCount, actualCount))
				}
			}
		}
	}
}
