package util

const DefaultPrecision = 3

func AppendNonNullValues(args ...string) []string {
	result := make([]string, 0, 0)

	for _, arg := range args {
		if len(arg) != 0 {
			result = append(result, arg)
		}
	}
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
