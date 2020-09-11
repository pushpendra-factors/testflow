package helpers

import (
	C "factors/config"
	U "factors/util"
)

const (
	QueryTypeFactor    = "factor"
	QueryTypeAnalytics = "analytics"
)

func IsValidQueryType(queryType string) bool {
	return queryType == QueryTypeFactor || queryType == QueryTypeAnalytics
}

func IsProjectWhitelistedForEventUserCache(projectID uint64) bool {
	//TODO: Janani Add code to fetch from config and whitelist
	whitelistedIds := C.GetWhitelistedProjectIdsEventUserCache()
	if whitelistedIds == "*" {
		return true
	}
	projectIdMap := U.GetIntBoolMapFromStringList(&whitelistedIds)
	for id, _ := range projectIdMap {
		if id == projectID {
			return true
		}
	}
	return false
}
