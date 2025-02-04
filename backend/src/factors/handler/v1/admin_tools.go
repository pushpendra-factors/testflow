package v1

import (
	//"errors"
	mid "factors/middleware"
	//	"factors/model/store"
	//	"io/ioutil"

	//"factors/model/model"
	//"factors/model/store"
	U "factors/util"
	//"fmt"
	C "factors/config"
	// serviceDisk "factors/services/disk"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func GetAnalyticsMetricsFromStorage(c *gin.Context) (interface{}, int, string, string, bool) {
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return nil, http.StatusBadRequest, INVALID_PROJECT, "Invalid Project ID", true
	}
	date := c.Query("date_range")
	dateRange, err := strconv.ParseInt(date, 10, 64)
	path, fileName := C.GetCloudManager().GetModelMetricsFilePathAndName(projectId, dateRange, dateRange+U.Per_week_epoch-1)
	log.Info("Metrics file path , filename", path, fileName)
	reader, err := C.GetCloudManager().Get(path, fileName)
	if err != nil {
		log.WithError(err).Error("Error reading metrics file")
		return nil, http.StatusInternalServerError, "", err.Error(), true
	}
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		log.WithError(err).Error("Error reading metrics file")
		return nil, http.StatusInternalServerError, "", err.Error(), true
	}
	jsonData := string(data)
	return jsonData, http.StatusFound, "", "", false
}
func GetAnalyticsAlertsFromStorage(c *gin.Context) (interface{}, int, string, string, bool) {
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return nil, http.StatusBadRequest, INVALID_PROJECT, "Invalid Project ID", true
	}
	date := c.Query("date_range")
	dateRange, err := strconv.ParseInt(date, 10, 64)
	path, fileName := C.GetCloudManager().GetModelAlertsFilePathAndName(projectId, dateRange, dateRange+U.Per_week_epoch-1)
	log.Info("Alerts file path , filename", path, fileName)
	reader, err := C.GetCloudManager().Get(path, fileName)
	if err != nil {
		log.WithError(err).Error("Error reading alerts file")
		return nil, http.StatusInternalServerError, "", err.Error(), true
	}
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		log.WithError(err).Error("Error reading alerts file")
		return nil, http.StatusInternalServerError, "", err.Error(), true
	}
	jsonData := string(data)
	return jsonData, http.StatusFound, "", "", false
}

func stringToTime(s string) (time.Time, error) {
	sec, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(sec, 0), nil
}
