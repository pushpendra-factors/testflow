package helpers

const (
	QueryTypeFactor    = "factor"
	QueryTypeAnalytics = "analytics"
)

func IsValidQueryType(queryType string) bool {
	return queryType == QueryTypeFactor || queryType == QueryTypeAnalytics
}
