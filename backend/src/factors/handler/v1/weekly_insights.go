package v1

import (
	"factors/delta"
	mid "factors/middleware"
	U "factors/util"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type WeeklyInsightsParams struct {
	ProjectID       uint64    `json:"project_id"`
	DashBoardUnitID uint64    `json:"dashboard_unit_id"`
	BaseStartTime   time.Time `json:"base_start_time"`
	CompStartTime   time.Time `json:"comp_start_time"`
	InsightsType    string    `json:"insights_type"`
	NumberOfRecords int       `json:"number_of_records"`
}

func GetWeeklyInsightsParams(c *gin.Context) (*WeeklyInsightsParams, error) {
	params := WeeklyInsightsParams{}
	projectID := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
	}
	DashBoardUnitID, err := strconv.ParseUint(c.Query("dashboard_unit_id"), 10, 64)
	if err != nil {
		return nil, err
	}
	BaseStartTimeTemp, err := strconv.ParseInt(c.Query("base_start_time"), 10, 64)
	if err != nil {
		return nil, err
	}
	BaseStartTime := time.Unix(BaseStartTimeTemp, 0)

	CompStartTimeTemp, err := strconv.ParseInt(c.Query("comp_start_time"), 10, 64)
	if err != nil {
		return nil, err
	}
	CompStartTime := time.Unix(CompStartTimeTemp, 0)

	insightsType := c.Query("insights_type")
	n, err := strconv.ParseInt(c.Query("number_of_records"), 10, 64)
	NumberOfRecords := int(n)

	params.ProjectID = projectID
	params.DashBoardUnitID = DashBoardUnitID
	params.BaseStartTime = BaseStartTime
	params.CompStartTime = CompStartTime
	params.InsightsType = insightsType
	params.NumberOfRecords = NumberOfRecords

	return &params, nil

}
func GetWeeklyInsightsHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	params, err := GetWeeklyInsightsParams(c)
	if err != nil {
		log.Error(err)
		return nil, http.StatusBadRequest, INVALID_INPUT, err.Error() + "1", true

	}
	if params.InsightsType != "w" && params.InsightsType != "m" {
		return nil, http.StatusBadRequest, INVALID_INPUT, "Enter w or m ", true
	}
	if params.NumberOfRecords > 100 || params.NumberOfRecords <= 0 {
		return nil, http.StatusBadRequest, INVALID_INPUT, "number of records must be in range 1-100", true
	}
	if params.ProjectID == 0 || params.DashBoardUnitID == 0 {
		return nil, http.StatusBadRequest, INVALID_INPUT, "invalid projectId or DashboardUnitID", true
	}
	response, err := delta.GetWeeklyInsights(params.ProjectID, params.DashBoardUnitID, &params.BaseStartTime, &params.CompStartTime, params.InsightsType, params.NumberOfRecords)
	if err != nil {
		log.Error(err)
		return err, http.StatusInternalServerError, "", "", true
	}
	return response, http.StatusAccepted, "", "", false
}
