package memsql

import (
	"encoding/json"
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

// CreateQueryAndSaveToDashboard Executes the following steps:
//  1. Create a query in db for given payload
//  2. Create attribution v1 dashboard, if it is not already present
//  3. Create a dashboard unit, link it to the query and save it in the dashboard
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
		log.WithField("err_code", errCode).Error("Failed to get dashboard.")
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
			ProjectID:    projectID,
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

	if errCode == http.StatusInternalServerError {
		log.WithField("err_code", errCode).Error("Failed to find attribution dashboard")
		return nil, http.StatusInternalServerError
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

// GetAttributionV1Dashboard gets  attribution v1 dashboard if exists else returns nil
func (store *MemSQL) GetAttributionV1Dashboard(projectId int64) (*model.Dashboard, int) {
	logFields := log.Fields{
		"project_id": projectId,
	}
	logCtx := log.WithFields(logFields)
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	var errCode int
	if projectId == 0 {
		log.Error("Failed to get dashboard. Invalid project_id or agent_id")
		return nil, http.StatusBadRequest
	}

	dashboard, errCode := store.GetAttributionV1DashboardByDashboardName(projectId, model.AttributionV1Name)

	if errCode == http.StatusInternalServerError {
		log.WithField("err_code", errCode).Error("Failed to find attribution dashboard")
		return nil, http.StatusInternalServerError
	}

	if errCode != http.StatusFound {
		return nil, errCode
	}
	if C.GetAttributionDebug() == 1 {
		logCtx.WithFields(log.Fields{"method": "GetOrCreateAttributionV1Dashboard", "dashboard": *dashboard}).Info("Attribution v1 dashboard debug.")
	}
	return dashboard, http.StatusFound
}

// GetAttributionDashboardUnitNamesImpactedByCustomKPI returns list of dashboards unit names (query title) which is dependent on custom KPI
func (store *MemSQL) GetAttributionDashboardUnitNamesImpactedByCustomKPI(projectID int64, customMetricName string) ([]string, int) {

	logFields := log.Fields{
		"project_id":         projectID,
		"custom_metric_name": customMetricName,
	}
	logCtx := log.WithFields(logFields)
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	impactedDashboardUnitNames := make([]string, 0)

	customMetricNameInAttributionQuery := "\"me\":[\"" + customMetricName + "\"],"

	// get the dashboard
	dashboard, errCode := store.GetAttributionV1Dashboard(projectID)

	if C.GetAttributionDebug() == 1 {
		log.WithFields(log.Fields{"dashboard": dashboard, "errCode": errCode, "customMetricName": customMetricName}).Warn("1 GetAttributionDashboardUnitNamesImpactedByCustomKPI")
	}

	if dashboard == nil || errCode != http.StatusFound {
		return impactedDashboardUnitNames, http.StatusNotFound
	}

	// get the dashboard units
	dashboardUnits, errCode := store.GetDashboardUnitByDashboardID(projectID, dashboard.ID)
	if C.GetAttributionDebug() == 1 {
		log.WithFields(log.Fields{"dashboardUnits": dashboardUnits, "errCode": errCode}).Warn("2 GetAttributionDashboardUnitNamesImpactedByCustomKPI")
	}
	if errCode != http.StatusFound || len(dashboardUnits) == 0 {
		logCtx.WithFields(log.Fields{"method": "GetAttributionV1DashboardByDashboardName", "dashboard": dashboard}).Info("Failed to get dashboard units for Attribution V1 dashboard")
		return impactedDashboardUnitNames, http.StatusNotFound
	}

	for _, unit := range dashboardUnits {

		// get the query
		queryInfo, errC := store.GetQueryWithQueryId(unit.ProjectID, unit.QueryId)

		if C.GetAttributionDebug() == 1 {
			log.WithFields(log.Fields{"queryInfo": queryInfo, "errC": errC}).Warn("3 GetAttributionDashboardUnitNamesImpactedByCustomKPI")
		}
		if errC != http.StatusFound {
			logCtx.WithField("err_code", errC).Errorf("Failed to fetch query from query_id %d", unit.QueryId)
			continue
		}

		// Create a new baseQuery instance every time to avoid overwriting from, to values in routines.
		queryClass := model.QueryClassAttributionV1
		baseQuery, err := model.DecodeQueryForClass(queryInfo.Query, queryClass)
		if err != nil {
			errMsg := fmt.Sprintf("Error decoding query, query_id %d", unit.QueryId)
			C.PingHealthcheckForFailure(C.HealthcheckDashboardDBAttributionPingID, errMsg)
			continue
		}

		attributionQuery := baseQuery.(*model.AttributionQueryUnitV1)
		var queryOriginal *model.AttributionQueryV1
		queryOriginal = attributionQuery.Query

		jsonString, err1 := json.Marshal(queryOriginal.KPIQueries)
		if C.GetAttributionDebug() == 1 {
			log.WithFields(log.Fields{"jsonString": jsonString, "string(jsonString)": string(jsonString), "err": err1}).
				Warn("4 GetAttributionDashboardUnitNamesImpactedByCustomKPI")
		}
		if err1 != nil {
			continue
		}

		// match the queries names
		if strings.Contains(string(jsonString), customMetricNameInAttributionQuery) {
			if C.GetAttributionDebug() == 1 {
				log.WithFields(log.Fields{"jsonString": string(jsonString), "string(jsonString)": string(jsonString),
					"customMetricNameInAttributionQuery": customMetricNameInAttributionQuery,
					"customMetricName":                   customMetricName,
				}).
					Warn("5 GetAttributionDashboardUnitNamesImpactedByCustomKPI")
			}
			impactedDashboardUnitNames = append(impactedDashboardUnitNames, queryInfo.Title)
		}
	}

	return impactedDashboardUnitNames, http.StatusFound
}

// GetAttributionSettingsKPIListForCustomKPI returns the custom KPI metric name if exists in the attribution's settings page.
func (store *MemSQL) GetAttributionSettingsKPIListForCustomKPI(projectID int64, customMetricName string) ([]string, int) {

	logFields := log.Fields{
		"project_id":         projectID,
		"custom_metric_name": customMetricName,
	}
	logCtx := log.WithFields(logFields)
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	impactedKPIAttributionConfigs := make([]string, 0)

	// "value": "Pipeline"}
	customMetricNameInAttributionKPIConfig := "\"value\":\"" + customMetricName + "\"}"

	// pulling project setting to build attribution query
	settings, errCode := store.GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		return impactedKPIAttributionConfigs, http.StatusNotFound
	}

	attributionConfig, err1 := decodeAttributionConfig(settings.AttributionConfig)
	if err1 != nil {
		return impactedKPIAttributionConfigs, http.StatusNotFound
	}

	jsonString, err1 := json.Marshal(attributionConfig.KpisToAttribute)
	if C.GetAttributionDebug() == 1 {
		logCtx.WithFields(log.Fields{"jsonString": jsonString, "string(jsonString)": string(jsonString), "err": err1}).
			Warn("04 GetAttributionSettingsKPIListForCustomKPI")
	}
	if err1 != nil {
		return impactedKPIAttributionConfigs, http.StatusNotFound
	}

	// match the config name with custom kpi name
	if strings.Contains(string(jsonString), customMetricNameInAttributionKPIConfig) {
		if C.GetAttributionDebug() == 1 {
			logCtx.WithFields(log.Fields{"jsonString": string(jsonString), "string(jsonString)": string(jsonString),
				"customMetricNameInAttributionKPIConfig": customMetricNameInAttributionKPIConfig,
				"customMetricName":                       customMetricName,
			}).
				Warn("05 GetAttributionSettingsKPIListForCustomKPI")
		}
		impactedKPIAttributionConfigs = append(impactedKPIAttributionConfigs, customMetricName)
	}
	return impactedKPIAttributionConfigs, http.StatusFound

}

// DeleteAttributionDashboardUnitAndQuery deletes query and corresponding dashboard unit for given project id
func (store *MemSQL) DeleteAttributionDashboardUnitAndQuery(projectID int64, queryID int64, agentUUID string, dashboardId int64, unitId int64) (int, string) {

	logFields := log.Fields{
		"project_id": projectID,
		"query_id":   queryID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	errCode := store.DeleteDashboardUnit(projectID, agentUUID, dashboardId, unitId)
	if errCode != http.StatusAccepted {
		return errCode, "Failed to delete dashboard unit."
	}

	errCode, errMsg := deleteQuery(projectID, queryID, model.QueryTypeAttributionV1Query)
	if errCode != http.StatusAccepted {
		return errCode, errMsg
	}

	errCodeShareableUrl := store.DeleteShareableURLWithEntityIDandType(projectID, queryID, model.ShareableURLEntityTypeQuery)
	if errCodeShareableUrl != http.StatusNotFound && errCodeShareableUrl != http.StatusAccepted {
		return errCodeShareableUrl, "Failed to delete shareable urls."
	}

	return http.StatusAccepted, "Successfully deleted."
}

// decodeAttributionConfig decode attribution config from project settings to map
func decodeAttributionConfig(config *postgres.Jsonb) (model.AttributionConfig, error) {
	attributionConfig := model.AttributionConfig{}
	if config == nil {
		return attributionConfig, nil
	}

	err := json.Unmarshal(config.RawMessage, &attributionConfig)
	if err != nil {
		return attributionConfig, err
	}

	return attributionConfig, nil
}
