package accountscoring

import (
	"fmt"
	"math"
	"sort"
)

// copyslice copies a slice of float64s
func copyslice(input []float64) []float64 {
	s := make([]float64, len(input))
	copy(s, input)
	return s
}

// sortedCopy returns a sorted copy of float64s
func sortedCopy(input []float64) (copy []float64) {
	copy = copyslice(input)
	sort.Float64s(copy)
	return
}

// Mean gets the average of a slice of numbers
func Mean(input []float64) (float64, error) {
	EmptyInputErr := fmt.Errorf("slice is empty")
	if len(input) == 0 {
		return math.NaN(), EmptyInputErr
	}
	sum := 0.0
	for _, num := range input {
		sum = sum + num
	}

	return sum / float64(len(input)), nil
}

// Percentile finds the relative standing in a slice of floats
func Percentile(input []float64, percent float64) (percentile float64, err error) {
	EmptyInputErr := fmt.Errorf("slice is empty")
	BoundsErr := fmt.Errorf("out of bounds")
	length := len(input)
	if length == 0 {
		return math.NaN(), EmptyInputErr
	}

	if length == 1 {
		return input[0], nil
	}

	if percent <= 0 || percent > 100 {
		return math.NaN(), BoundsErr
	}

	// Start by sorting a copy of the slice
	c := sortedCopy(input)

	// Multiply percent by length of input
	index := (percent / 100) * float64(len(c))

	// Check if the index is a whole number
	if index == float64(int64(index)) {

		// Convert float to int
		i := int(index)

		// Find the value at the index
		percentile = c[i-1]

	} else if index > 1 {

		// Convert float to int via truncation
		i := int(index)

		// Find the average of the index and following values
		percentile, _ = Mean([]float64{c[i-1], c[i]})

	} else {
		return math.NaN(), BoundsErr
	}

	return percentile, nil

}

// PercentileNearestRank finds the relative standing in a slice of floats using the Nearest Rank method
func PercentileNearestRank(input []float64, percent float64) (percentile float64, err error) {

	// Find the length of items in the slice
	il := len(input)
	EmptyInputErr := fmt.Errorf("slice is empty")
	BoundsErr := fmt.Errorf("out of bounds")
	// Return an error for empty slices
	if il == 0 {
		return math.NaN(), EmptyInputErr
	}

	// Return error for less than 0 or greater than 100 percentages
	if percent < 0 || percent > 100 {
		return math.NaN(), BoundsErr
	}

	// Start by sorting a copy of the slice
	c := sortedCopy(input)

	// Return the last item
	if percent == 100.0 {
		return c[il-1], nil
	}

	// Find ordinal ranking
	or := int(math.Ceil(float64(il) * percent / 100))

	// Return the item that is in the place of the ordinal rank
	if or == 0 {
		return c[0], nil
	}
	return c[or-1], nil

}

func removeZeros(input []float64) []float64 {
	var result []float64

	for _, value := range input {
		if value != 0 {
			result = append(result, value)
		}
	}
	return result
}

func GetEngagementLevels(scores []float64) map[float64]string {
	result := make(map[float64]string)
	result[0] = getEngagement(0)

	nonZeroScores := removeZeros(scores)

	for _, score := range nonZeroScores {
		percentile := calculatePercentile(nonZeroScores, score)
		result[score] = getEngagement(percentile)
	}

	return result
}

func calculatePercentile(data []float64, value float64) float64 {
	sort.Float64s(data)                                       // Sort the data in ascending order
	index := sort.SearchFloat64s(data, value)                 // Find the index of the value
	percentile := float64(index) / float64(len(data)-1) * 100 // Calculate the percentile based on the index
	return percentile
}

func getEngagement(percentile float64) string {
	if percentile > 90 {
		return "Hot"
	} else if percentile > 70 {
		return "Warm"
	} else {
		return "Cool"
	}
}
