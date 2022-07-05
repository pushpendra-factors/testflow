package v1

import (
	"errors"
	C "factors/config"
	"factors/delta"
	mid "factors/middleware"
	"factors/model/store"
	U "factors/util"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type WeeklyInsightsParams struct {
	ProjectID       int64     `json:"project_id"`
	QueryId         int64     `json:"query_id"`
	BaseStartTime   time.Time `json:"base_start_time"`
	CompStartTime   time.Time `json:"comp_start_time"`
	InsightsType    string    `json:"insights_type"`
	NumberOfRecords int       `json:"number_of_records"`
}

func GetWeeklyInsightsParams(c *gin.Context) (*WeeklyInsightsParams, error) {
	params := WeeklyInsightsParams{}
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		return nil, errors.New("Invalid Project ID")
	}
	var QueryId int64
	if c.Query("query_id") != "" {
		Qid, err := strconv.ParseInt(c.Query("query_id"), 10, 64)
		if err != nil {
			return nil, err
		}
		QueryId = Qid
	} else {
		DashBoardUnitID, err := strconv.ParseInt(c.Query("dashboard_unit_id"), 10, 64)
		if err != nil {
			return nil, err
		}
		if DashBoardUnitID == 0 {
			return nil, errors.New("Invalid Dashboard ID")
		}
		DashBoardUnit, status := store.GetStore().GetDashboardUnitByUnitID(projectID, DashBoardUnitID)
		if status != http.StatusFound {
			return nil, errors.New("Dashboard Not found given dashboard ID")
		}
		QueryId = DashBoardUnit.QueryId
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
	n := c.Query("number_of_records")
	var NumberOfRecords int64
	if n != "" {
		NumberOfRecords, _ = strconv.ParseInt(n, 10, 64)
	} else {
		NumberOfRecords = 20 // default
	}
	params.ProjectID = projectID
	params.QueryId = QueryId
	params.BaseStartTime = BaseStartTime
	params.CompStartTime = CompStartTime
	params.InsightsType = insightsType
	params.NumberOfRecords = int(NumberOfRecords)

	return &params, nil

}
func GetWeeklyInsightsHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	var resp delta.WeeklyInsights
	params, err := GetWeeklyInsightsParams(c)
	if err != nil {
		log.Error(err)
		return resp, http.StatusBadRequest, INVALID_INPUT, err.Error() + "1", true
	}
	agentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	if !C.IsWeeklyInsightsWhitelisted(agentUUID, params.ProjectID) {
		return nil, http.StatusOK, "", "", false
	}
	if params.InsightsType != "w" && params.InsightsType != "m" {
		return nil, http.StatusBadRequest, INVALID_INPUT, "Enter w or m ", true
	}
	if params.NumberOfRecords > 100 || params.NumberOfRecords <= 0 {
		return nil, http.StatusBadRequest, INVALID_INPUT, "number of records must be in range 1-100", true
	}
	if params.ProjectID == 0 || params.QueryId == 0 {
		return nil, http.StatusBadRequest, INVALID_INPUT, "invalid projectId or QueryId", true
	}
	response, err := delta.GetWeeklyInsights(params.ProjectID, agentUUID, params.QueryId, &params.BaseStartTime, &params.CompStartTime, params.InsightsType, params.NumberOfRecords)
	if err != nil {
		log.Error(err)
		return err, http.StatusInternalServerError, "", "", true
	}
	return response, http.StatusAccepted, "", "", false
}
