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
				h, _ := NewCategoricalHistogram(maxBins, dimensions, nil)
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

				assert.Equal(t, numSamples, int(h.Count()),
					fmt.Sprintf("Count mismatch %v != %v", h.Count(), numSamples))
				assert.Equal(t, numSamples, int(h.totalBinCount()),
					fmt.Sprintf("Bin Total Count mismatch %v != %v", h.totalBinCount(), numSamples))

				for k, expectedCount := range frequencies {
					actualCount := h.frequency(k)
					assert.Equal(t, expectedCount, int(actualCount),
						fmt.Sprintf("Count mismatch for symbol %s %v != %v",
							k, expectedCount, actualCount))
				}
			}
		}
	}
}

func TestCategoricalHistogramAddWithTemplate(t *testing.T) {
	maxBins := 10
	dimensions := 4
	numSamples := 100
	template := CategoricalHistogramTemplate{
		CategoricalHistogramTemplateUnit{
			Name:       "key1",
			IsRequired: true,
		},
		CategoricalHistogramTemplateUnit{
			Name:       "key2",
			IsRequired: false,
			Default:    "defaultValue2",
		},
		CategoricalHistogramTemplateUnit{
			Name:       "key3",
			IsRequired: true,
		},
		CategoricalHistogramTemplateUnit{
			Name:       "key4",
			IsRequired: false,
		},
	}

	// Each dimension can take upto 1 to 5 values.
	numAllowedValues := make([]int, dimensions)
	for i := 0; i < dimensions; i++ {
		numAllowedValues[i] = rand.Intn(5) + 1
	}

	// Track and check exact frequencies.
	h, _ := NewCategoricalHistogram(maxBins, dimensions, &template)
	var frequencies = make(map[string]int)
	for j := 0; j < numSamples; {
		switch choice := rand.Intn(5); choice {
		case 0:
			// Add (values)
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
			err := h.Add(sample)
			assert.Nil(t, err)
			j++
		case 1:
			// AddMap all 4 dimensions.
			keyValues := map[string]string{}
			for i := 0; i < dimensions; i++ {
				key := fmt.Sprintf("key%d", i+1)
				v := fmt.Sprintf("%d%d", i+1, rand.Intn(numAllowedValues[i])+1)
				keyValues[key] = v
				if count, ok := frequencies[v]; ok {
					frequencies[v] = count + 1
				} else {
					frequencies[v] = 1
				}
			}
			err := h.AddMap(keyValues)
			assert.Nil(t, err)
			j++
		case 2:
			// AddMap with default dimension.
			// key1 and key3 are required. key2 and key4 uses default values.
			keyValues := map[string]string{}
			v1 := fmt.Sprintf("%d%d", 1, rand.Intn(numAllowedValues[0])+1)
			keyValues["key1"] = v1
			v3 := fmt.Sprintf("%d%d", 3, rand.Intn(numAllowedValues[2])+1)
			keyValues["key3"] = v3
			err := h.AddMap(keyValues)
			assert.Nil(t, err)
			for _, v := range []string{template[0].Default, v1, template[2].Default, v3} {
				if v == "" {
					continue
				}
				if count, ok := frequencies[v]; ok {
					frequencies[v] = count + 1
				} else {
					frequencies[v] = 1
				}
			}
			j++
		case 3:
			// AddMap error missing required dimension.
			// key3 is missing
			// Frequencies are not recorded since it is not added to histogram.
			keyValues := map[string]string{}
			keyValues["key1"] = fmt.Sprintf("%d%d", 1, rand.Intn(numAllowedValues[0])+1)
			keyValues["key2"] = fmt.Sprintf("%d%d", 2, rand.Intn(numAllowedValues[1])+1)
			keyValues["key4"] = fmt.Sprintf("%d%d", 4, rand.Intn(numAllowedValues[3])+1)
			err := h.AddMap(keyValues)
			assert.NotNil(t, err)
		case 4:
			// AddMap error with extra unknown keys.
			// Frequencies are not recorded since it is not added to histogram.
			keyValues := map[string]string{}
			keyValues["key1"] = fmt.Sprintf("%d%d", 1, rand.Intn(numAllowedValues[0])+1)
			keyValues["key2"] = fmt.Sprintf("%d%d", 2, rand.Intn(numAllowedValues[1])+1)
			keyValues["key3"] = fmt.Sprintf("%d%d", 3, rand.Intn(numAllowedValues[2])+1)
			keyValues["key4"] = fmt.Sprintf("%d%d", 4, rand.Intn(numAllowedValues[3])+1)
			keyValues["key5"] = "unexpectedKey"
			err := h.AddMap(keyValues)
			assert.NotNil(t, err)
		}
	}
	assert.Equal(t, numSamples, int(h.Count()),
		fmt.Sprintf("Count mismatch %v != %v", h.Count(), numSamples))
	assert.Equal(t, numSamples, int(h.totalBinCount()),
		fmt.Sprintf("Bin Total Count mismatch %v != %v", h.totalBinCount(), numSamples))

	for k, expectedCount := range frequencies {
		actualCount := h.frequency(k)
		assert.Equal(t, expectedCount, int(actualCount),
			fmt.Sprintf("Count mismatch for symbol %s %v != %v",
				k, expectedCount, actualCount))
	}
}

func TestTrimFrequencyMap(t *testing.T) {
	maxSize := 3

	fmap := map[string]uint64{
		"key1": 10,
		"key2": 5,
		"key3": 7,
		"key4": 9,
	}
	tFmap := trimFrequencyMap(fmap, maxSize)
	fmt.Println(fmt.Sprintf("tFmap: %v", tFmap))
	assert.Equal(t, maxSize, len(tFmap))
	for _, key := range []string{"key1", "key4"} {
		actualCount, ok := tFmap[key]
		assert.True(t, ok, fmt.Sprintf("Missing key: %s", key))
		expectedCount, _ := fmap[key]
		assert.Equal(t, expectedCount, actualCount,
			fmt.Sprintf("Mismatch in count value for key: %s", key))
	}
	actualOtherCount, ok := tFmap[fMAP_OTHER_KEY]
	assert.True(t, ok)
	assert.Equal(t, uint64(12), actualOtherCount)

	fmap = map[string]uint64{
		"key1":         10,
		"key2":         5,
		"key3":         7,
		"key4":         9,
		fMAP_OTHER_KEY: 12,
	}
	tFmap = trimFrequencyMap(fmap, maxSize)
	fmt.Println(fmt.Sprintf("tFmap: %v", tFmap))
	assert.Equal(t, maxSize, len(tFmap))
	for _, key := range []string{"key1", "key4"} {
		actualCount, ok := tFmap[key]
		assert.True(t, ok, fmt.Sprintf("Missing key: %s", key))
		expectedCount, _ := fmap[key]
		assert.Equal(t, expectedCount, actualCount,
			fmt.Sprintf("Mismatch in count value for key: %s", key))
	}
	actualOtherCount, ok = tFmap[fMAP_OTHER_KEY]
	assert.True(t, ok)
	assert.Equal(t, uint64(24), actualOtherCount)
}
