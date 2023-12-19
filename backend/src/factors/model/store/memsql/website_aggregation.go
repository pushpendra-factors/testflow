package memsql

import (
	C "factors/config"
	"fmt"
	"net/http"
	"time"
	"strconv"
	"factors/model/model"

	log "github.com/sirupsen/logrus"
)

var DefaultDateFromObj = time.Date(0001, 01, 01, 00, 00, 00, 00, time.UTC)

type MaxTimestampStruct struct{ MaxTimestamp []uint8 }

const maxTimestampFromWebsiteAggregationSql = "select CASE WHEN max(timestamp_at_day) IS NULL THEN " + 
 	"UNIX_TIMESTAMP(date_trunc('day', NOW() - INTERVAL 32 day) ) ELSE " + 
	"max(timestamp_at_day) END AS max_timestamp from website_aggregation " +
	"WHERE timestamp_at_day >= UNIX_TIMESTAMP(date_trunc('day', NOW() - INTERVAL 32 day) ) AND project_id = %v;"

// Dbt uses this for computing min timestamp from which it has to pull data.
// Returns the max timestamp for which data is present and return.
func (store *MemSQL) GetMaxTimestampOfDataPresenceFromWebsiteAggregation(projectID int64, timezone string) (time.Time, int) {

	var maxTimestampStruct MaxTimestampStruct
	db := C.GetServices().Db

	sql := fmt.Sprintf(maxTimestampFromWebsiteAggregationSql, projectID)
	if err := db.Raw(sql).Scan(&maxTimestampStruct).Error; err != nil {
		log.WithField("projectID", projectID).WithField("err", err).Error("Failed to execute min timestamp website aggregation")
		return time.Date(2000, 01, 01, 01, 01, 01, 0, time.UTC), http.StatusInternalServerError
	}
	valueInString := string(maxTimestampStruct.MaxTimestamp)
	valueInFloat64, _ := strconv.ParseFloat(valueInString, 32)
	valueInInt64 := int64(valueInFloat64)

	if valueInInt64 == 0 {
		return time.Date(2000, 01, 01, 01, 01, 01, 0, time.UTC), http.StatusInternalServerError
	}
	epochValue := valueInInt64
	valueInTime := time.Unix(epochValue, 0).UTC()
	location, _ := time.LoadLocation(timezone)
	valueInTimeZone := time.Date(valueInTime.Year(), valueInTime.Month(), valueInTime.Day(), 0, 0, 0, 0, location)
	return valueInTimeZone, http.StatusOK
}

func (store *MemSQL) CreateWebsiteAggregation(websiteAggregation model.WebsiteAggregation) (model.WebsiteAggregation, string, int) {
	logFields := log.Fields{
		"website_aggregation": websiteAggregation,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	if websiteAggregation.ProjectID == 0 {
		logCtx.Warn("Invalid project ID for website aggregation")
		return model.WebsiteAggregation{}, "Invalid project ID for website aggregation", http.StatusBadRequest
	}

	db := C.GetServices().Db
	if err := db.Create(&websiteAggregation).Error; err != nil {
		logCtx.WithError(err).Error("Failed while creating website aggregation")
		return model.WebsiteAggregation{}, err.Error(), http.StatusInternalServerError
	}

	return websiteAggregation, "", http.StatusCreated

}
