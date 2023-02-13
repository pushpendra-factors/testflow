package handler

import (
	"encoding/json"
	"factors/delta"
	V1 "factors/handler/v1"
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

func GetSixSignalReportHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	r := c.Request

	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		log.Error("Query failed. Invalid project.")
		return nil, http.StatusUnauthorized, V1.INVALID_PROJECT, "Query failed. Invalid project.", true
	}

	logCtx := log.WithFields(log.Fields{
		"projectId": projectId,
	})

	var requestPayload model.SixSignalQueryGroup

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&requestPayload); err != nil {
		logCtx.WithError(err).Error("Query failed. Json decode failed.")
		return nil, http.StatusBadRequest, V1.INVALID_INPUT, "Query failed. Json decode failed.", true
	}

	var commonQueryFrom int64
	var commonQueryTo int64
	if len(requestPayload.Queries) == 0 {
		logCtx.Error("Query failed. Empty query group.")
		return nil, http.StatusBadRequest, V1.INVALID_INPUT, "Query failed. Empty query group.", true
	} else {
		// all group queries are run for same time duration, used in dashboard unit caching
		commonQueryFrom = requestPayload.Queries[0].From
		commonQueryTo = requestPayload.Queries[0].To
	}

	var timezoneString U.TimeZoneString
	var statusCode int
	if requestPayload.Queries[0].Timezone != "" {
		_, errCode := time.LoadLocation(string(requestPayload.Queries[0].Timezone))
		if errCode != nil {
			return nil, http.StatusBadRequest, V1.INVALID_INPUT, "Query failed. Invalid Timezone provided.", true
		}
		timezoneString = U.TimeZoneString(requestPayload.Queries[0].Timezone)
	} else {
		timezoneString, statusCode = store.GetStore().GetTimezoneForProject(projectId)
		if statusCode != http.StatusFound {
			logCtx.Error("Query failed. Failed to get Timezone.")
			return nil, http.StatusBadRequest, V1.INVALID_INPUT, "Query failed. Failed to get Timezone.", true
		}
		// logCtx.WithError(err).Error("Query failed. Invalid Timezone.")
	}
	requestPayload.SetTimeZone(timezoneString)

	fromDate := U.GetDateOnlyFromTimestampZ(commonQueryFrom)
	toDate := U.GetDateOnlyFromTimestampZ(commonQueryTo)
	folderName := fmt.Sprintf("%v-%v", fromDate, toDate)
	logCtx.WithFields(log.Fields{"folder name": folderName}).Info("Folder name for reading the result")

	result := delta.GetSixSignalAnalysisData(projectId, folderName)
	return result, http.StatusOK, "", "", false
}
