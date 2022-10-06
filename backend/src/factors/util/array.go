package util

import (
	"strings"
)

const DefaultPrecision = 3

func ContainsUint64InArray(s []uint64, e uint64) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func ContainsInt64InArray(s []int64, e int64) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func ContainsStringInArray(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func AppendNonNullValues(args ...string) []string {
	result := make([]string, 0, 0)

	for _, arg := range args {
		if len(arg) != 0 {
			result = append(result, arg)
		}
	}
	return result
}

func ContainsDuplicate(args ...interface{}) bool {
	presentAlready := make(map[interface{}]struct{})
	for _, arg := range args {
		if _, exists := presentAlready[arg]; exists {
			return true
		} else {
			presentAlready[arg] = struct{}{}
		}
	}
	return false
}
func AppendNonNullValuesList(args []string, arg string) []string {
	result := make([]string, 0, 0)

	for _, argValue := range args {
		if len(argValue) != 0 {
			result = append(result, argValue)
		}
	}
	result = append(result, arg)
	return result
}

// ConvertInternalToExternal ...
// Standardising the external API response for metrics. This is used to convert the internal metrics to external API.
func ConvertInternalToExternal(internalMetrics [][]interface{}) [][]interface{} {
	externalMetrics := make([][]interface{}, 0, 0)
	for _, internalRow := range internalMetrics {
		externalRow := make([]interface{}, 0, 0)
		for _, metric := range internalRow {
			var value interface{}
			if metric == nil {
				value = 0
			} else {
				switch metric.(type) {
				case float64:
					value, _ = FloatRoundOffWithPrecision(metric.(float64), DefaultPrecision)
				case float32:
					value, _ = FloatRoundOffWithPrecision(float64(metric.(float32)), DefaultPrecision)
				default:
					value = metric
				}
			}
			externalRow = append(externalRow, value)
		}
		externalMetrics = append(externalMetrics, externalRow)
	}
	return externalMetrics
}

func StringsWithMatchingPrefix(s []string, prefix string) []string {
	finalStrings := make([]string, 0)
	if len(s) == 0 {
		return finalStrings
	}
	for _, a := range s {
		if strings.HasPrefix(a, prefix) {
			finalStrings = append(finalStrings, a)
		}
	}
	return finalStrings
}
