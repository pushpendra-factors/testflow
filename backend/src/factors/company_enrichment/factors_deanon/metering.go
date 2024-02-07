package factors_deanon

import (
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// CheckingFactorsDeanonQuotaLimit compares the usage and limit for the project id.
func CheckingFactorsDeanonQuotaLimit(projectId int64) (bool, error) {

	logCtx := log.WithField("project_id", projectId)

	count, limit, err := GetFactorsDeanonCountAndLimit(projectId)
	if err != nil {
		return false, err
	}
	if count >= limit {
		logCtx.Warn("Factors Deanonymisation Limit Exhausted")
		return false, nil
	}
	return true, nil
}

// GetFactorsDeanonCountAndLimit fetches the count and limit of factors deanon enrichment
func GetFactorsDeanonCountAndLimit(projectId int64) (int64, int64, error) {

	logCtx := log.WithField("project_id", projectId)
	limit, err := store.GetStore().GetFeatureLimitForProject(projectId, model.FEATURE_FACTORS_DEANONYMISATION)
	if err != nil {
		logCtx.Error("Failed fetching sixsignal quota limit with error ", err)
		return 0, -1, err
	}

	timeZone, statusCode := store.GetStore().GetTimezoneForProject(projectId)
	if statusCode != http.StatusFound {
		timeZone = U.TimeZoneStringIST
	}
	monthYear := U.GetCurrentMonthYear(timeZone)
	count, err := model.GetSixSignalMonthlyUniqueEnrichmentCount(projectId, monthYear)
	if err != nil {
		logCtx.Error("Error while fetching Factors Deanonymisation count")
		return 0, -1, err
	}
	return count, limit, nil
}

func GetAccountLimitAndPercentageExhausted(projectId int64) (float64, int64, error) {

	count, limit, err := GetFactorsDeanonCountAndLimit(projectId)
	if err != nil {
		return 0, 0, err
	}

	percentageExhausted := float64(count) / float64(limit) * 100.0

	return percentageExhausted, limit, nil

}
