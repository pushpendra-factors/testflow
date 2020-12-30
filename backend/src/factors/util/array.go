package util

func AppendNonNullValues(args ...string) []string {
	result := make([]string, 0, 0)

	for _, arg := range args {
		if len(arg) != 0 {
			result = append(result, arg)
		}
	}
	return result
}
