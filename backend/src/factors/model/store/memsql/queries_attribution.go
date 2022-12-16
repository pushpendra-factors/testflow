package memsql

import (
	"encoding/json"
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

// CreateQueryAndSaveToDashboard Executes the following steps:
//	1. Create a query in db for given payload
// 	2. Create attribution v1 dashboard, if it is not already present
//	3. Create a dashboard unit, link it to the query and save it in the dashboard
func (store *MemSQL) CreateQueryAndSaveToDashboard(projectID int64, queryInfo *model.CreateQueryAndSaveToDashboardInfo) (*model.QueryAndDashboardUnit, int, string) {
	queryRequest := &model.Queries{
		Query:     *queryInfo.Query,
		Title:     queryInfo.Title,
		Type:      queryInfo.Type,
		CreatedBy: queryInfo.CreatedBy,
		// To support empty settings value.
		Settings: postgres.Jsonb{RawMessage: json.RawMessage(`{}`)},
		IdText:   U.RandomStringForSharableQuery(50),
	}
	if queryInfo.Settings != nil && !U.IsEmptyPostgresJsonb(queryInfo.Settings) {
		queryRequest.Settings = *queryInfo.Settings
	}
	query, errCode, errMsg := store.CreateQuery(projectID, queryRequest)
	if errCode != http.StatusCreated {
		return nil, errCode, errMsg
	}

	if C.GetAttributionDebug() == 1 {
		log.WithFields(log.Fields{"method": "CreateQueryAndSaveToDashboard", "query": *query}).Info("Attribution v1 dashboard debug. After CreateQuery.")
	}
	dashboard, errCode := store.GetOrCreateAttributionV1Dashboard(projectID, queryInfo.CreatedBy)
	if errCode != http.StatusFound {
		log.Error("Failed to get dashboard.")
		return nil, errCode, "Failed to get attribution V1 dashboard."

	}
	if C.GetAttributionDebug() == 1 {
		log.WithFields(log.Fields{"method": "CreateQueryAndSaveToDashboard", "dashboard": *dashboard}).Info("Attribution v1 dashboard debug. After GetOrCreateAttributionV1Dashboard.")
	}
	if query.ID == 0 {
		return nil, http.StatusBadRequest, "Invalid queryId."
	}

	dashboardUnit, errCode, errMsg := store.CreateDashboardUnit(projectID, queryInfo.CreatedBy,
		&model.DashboardUnit{
			DashboardId:  dashboard.ID,
			Presentation: queryInfo.DashboardUnitPresentation,
			QueryId:      query.ID,
		})
	if errCode != http.StatusCreated {
		return nil, errCode, errMsg
	}
	queryAndDashboardUnit := model.QueryAndDashboardUnit{
		Query:         *query,
		DashboardUnit: *dashboardUnit,
	}
	if C.GetAttributionDebug() == 1 {
		log.WithFields(log.Fields{"method": "CreateQueryAndSaveToDashboard", "queryAndDashboardUnit": queryAndDashboardUnit}).Info("Attribution v1 dashboard debug.")
	}
	return &queryAndDashboardUnit, http.StatusCreated, ""
}

// GetOrCreateAttributionV1Dashboard gets or creates  attribution v1 dashboard for given project id
func (store *MemSQL) GetOrCreateAttributionV1Dashboard(projectId int64, agentUUID string) (*model.Dashboard, int) {
	logFields := log.Fields{

		"project_id": projectId,
		"agent_uuid": agentUUID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	var errCode int
	if projectId == 0 || agentUUID == "" {
		log.Error("Failed to get dashboard. Invalid project_id or agent_id")
		return nil, http.StatusBadRequest
	}

	dashboard, errCode := store.GetAttributionV1DashboardByDashboardName(projectId, model.AttributionV1Name)
	if C.GetAttributionDebug() == 1 {
		log.WithFields(log.Fields{"method": "GetOrCreateAttributionV1Dashboard", "dashboard": *dashboard}).Info("Attribution v1 dashboard debug. After GetAttributionV1DashboardByDashboardName")
	}
	if errCode == http.StatusNotFound {

		dashboardRequest := &model.Dashboard{
			Name:        model.AttributionV1Name,
			Description: model.AttributionV1Description,
			Type:        model.DashboardTypeAttributionV1,
			Settings:    postgres.Jsonb{RawMessage: json.RawMessage(`{}`)},
		}
		dashboard, errCode = store.CreateDashboard(projectId, agentUUID, dashboardRequest)
		{
			log.WithFields(log.Fields{"method": "GetOrCreateAttributionV1Dashboard", "dashboard": *dashboard}).Info("Attribution v1 dashboard debug. After CreateDashboard")
		}
		if errCode != http.StatusCreated {
			return nil, errCode
		}

	} else if errCode != http.StatusFound {
		return nil, errCode
	}
	if C.GetAttributionDebug() == 1 {
		log.WithFields(log.Fields{"method": "GetOrCreateAttributionV1Dashboard", "dashboard": *dashboard}).Info("Attribution v1 dashboard debug.")
	}
	return dashboard, http.StatusFound
}
