package histogram

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHistogramFrequencyMaps(t *testing.T) {
	for _, dimensions := range []int{2, 3, 4, 5, 10} {
		// Each dimension can take upto 1 to 5 values.
		numAllowedValues := make([]int, dimensions)
		for i := 0; i < dimensions; i++ {
			numAllowedValues[i] = rand.Intn(5) + 1
		}
		for _, maxBins := range []int{2, 3, 4, 5, 10} {
			for _, numSamples := range []int{10, 50, 100} {
				// Track and check exact frequencies.
				h, _ := NewCategoricalHistogram(maxBins, dimensions, nil)
				var symbolFrequencies = make(map[string]int)
				allSamples := [][]string{}
				for j := 0; j < numSamples; j++ {
					var sample = []string{}
					for i := 0; i < dimensions; i++ {
						// Each dimension in samples look like
						// 11, 12, ""    -> Dimension 1, numAllowedValues 2
						// 21, 22, 23, ""  -> Dimension 2, numAllowedValues 3 and so on.
						// Empty String means missing value.
						v := ""
						if rand.Intn(10) > 1 {
							// 20% chance of a symbol being an empty string.
							v = fmt.Sprintf("%d%d", i+1, rand.Intn(numAllowedValues[i])+1)
							if count, ok := symbolFrequencies[v]; ok {
								symbolFrequencies[v] = count + 1
							} else {
								symbolFrequencies[v] = 1
							}
						}
						sample = append(sample, v)

					}
					h.Add(sample)
					allSamples = append(allSamples, sample)
				}

				assert.Equal(t, numSamples, int(h.Count()),
					fmt.Sprintf("Count mismatch %v != %v", h.Count(), numSamples))
				assert.Equal(t, numSamples, int(h.totalBinCount()),
					fmt.Sprintf("Bin Total Count mismatch %v != %v", h.totalBinCount(), numSamples))

				// Check individual symbol frequencies.
				for k, expectedCount := range symbolFrequencies {
					actualCount := h.frequency(k)
					assert.Equal(t, expectedCount, int(actualCount),
						fmt.Sprintf("Count mismatch for symbol %s %v != %v",
							k, expectedCount, actualCount))
				}

				for _, vec := range allSamples {
					numVecOccurrences := 0.0
					for _, vec2 := range allSamples {
						isMatch := true
						for vecI, symbol1 := range vec {
							if symbol1 != "" && symbol1 != vec2[vecI] {
								// if symbol1 is specific and vec2[vecI] is empty,
								// it is not counted against symbol1 as it is a missing value
								// and could be any value.
								// A missing value in a query however (i.e symbol is empty)
								// means to count only combinations that have a value. (marginal distribution)
								isMatch = false
								break
							}
						}
						if isMatch {
							numVecOccurrences += 1
						}
					}
					expectedPdf := numVecOccurrences / float64(numSamples)
					actualPdf, err := h.PDF(vec)
					assert.Nil(t, err, fmt.Sprintf("Failed for string vec %v, error: %v", vec, err))
					// The actual value should not be off by more than 20% factored down by number of bins
					// give or take 5 for edge conditions.
					maxDelta := float64(numSamples)*expectedPdf*0.20/float64(maxBins) + 5.0
					assert.InDelta(t, int(expectedPdf*float64(numSamples)),
						int(actualPdf*float64(numSamples)), maxDelta,
						fmt.Sprintf(
							"Mismatch for vec %v, dimensions %d,"+
								"numSamples %d, numBins: %d, actualPDF %.3f, expectedPDF %.3f,"+
								"numAllowedValues %v",
							vec, dimensions, numSamples,
							maxBins, actualPdf, expectedPdf, numAllowedValues))
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
	var allSamples = []map[string]string{}
	for j := 0; j < numSamples; {
		switch choice := rand.Intn(5); choice {
		case 0:
			// Add (values)
			var sample = []string{}
			keyValues := map[string]string{}
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
				key := fmt.Sprintf("key%d", i+1)
				keyValues[key] = v
			}
			err := h.Add(sample)
			assert.Nil(t, err)
			allSamples = append(allSamples, keyValues)
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
			allSamples = append(allSamples, keyValues)
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
			allSamples = append(allSamples, keyValues)
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
			allSamples = append(allSamples, keyValues)
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
			allSamples = append(allSamples, keyValues)
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

	for _, mp := range allSamples {
		numSymbolCombinations := 1
		for key, symbol := range mp {
			if symbol != "" {
				// The number of combinations in the marginal
				// probability distribution function with missing
				// dimensions is lower.
				i, _ := strconv.ParseInt(string(key[3]), 10, 32)
				i -= 1
				if int(i) < dimensions {
					numSymbolCombinations *= numAllowedValues[i]
				}
			}
		}
		actualPdf, err := h.PDFFromMap(mp)
		assert.Nil(t, err, fmt.Sprintf("Failed for string vec %v, error: %v", mp, err))
		expectedPdf := 1.0 / float64(numSymbolCombinations)
		// The expected error can be calculated by modelling this as binomial distribution.
		// If p = 1/numSymbolCombinations is the probability of the sample occurring.
		// 1 - p is the probability of it not occurrint.
		// N = numSamples, is the number of trials.
		// expected mean number of times of sample occurring = np
		// expected variance = np(1-p)
		// The actual value should not be more than 4 standard deviation from the mean.
		// give or take 2 whole numbers for edge conditions.
		maxDelta := 4.0*math.Sqrt(float64(numSamples)*expectedPdf*(1.0-expectedPdf)) + 2
		assert.InDelta(t, int(expectedPdf*float64(numSamples)),
			int(actualPdf*float64(numSamples)), maxDelta,
			fmt.Sprintf(
				"Mismatch for map %v, numSymbolCombinations %d, numSamples %d,"+
					"numBins: %d, actualPDF %.3f, expectedPDF %.3f",
				mp, numSymbolCombinations, numSamples,
				maxBins, actualPdf, expectedPdf))
	}
	var nonExistingPropertyInHistogram = map[string]string{}
	nonExistingPropertyInHistogram["no_prop"] = "abc"
	actualPdf, _ := h.PDFFromMap(nonExistingPropertyInHistogram)
	assert.Equal(t, actualPdf, 0.0)
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
