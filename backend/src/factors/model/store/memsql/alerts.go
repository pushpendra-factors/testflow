package memsql

import (
	"errors"
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"

	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) SetAuthTokenforSlackIntegration(projectID int64, agentUUID string, authTokens model.SlackAccessTokens) error {
	db := C.GetServices().Db
	_, errCode := store.GetProjectAgentMapping(projectID, agentUUID)
	if errCode != http.StatusFound {
		log.WithField("err_code", errCode).Error("Project agent mapping not found.")
		return errors.New("Project agent mapping not found.")
	}
	var agent model.Agent
	err := db.Where("uuid = ?", agentUUID).Find(&agent).Error
	if err != nil {
		log.WithError(err).Error("Failure in SetAuthTokenforSlackIntegration")
		return err
	}
	var token model.SlackAuthTokens
	if IsEmptyPostgresJsonb(agent.SlackAccessTokens) {
		token = make(map[int64]model.SlackAccessTokens)
	} else {
		err = U.DecodePostgresJsonbToStructType(agent.SlackAccessTokens, &token)
		if err != nil {
			log.WithError(err).Error("Failure in SetAuthTokenforSlackIntegration")
			return err
		}
	}
	token[projectID] = authTokens
	// convert token to json
	TokenJson, err := U.EncodeStructTypeToPostgresJsonb(token)
	if err != nil {
		log.WithError(err).Error("Failure in SetAuthTokenforSlackIntegration")
		return err
	}
	// update the db
	err = db.Model(&model.Agent{}).Where("uuid = ?", agentUUID).Update("slack_access_tokens", TokenJson).Error
	if err != nil {
		log.WithError(err).Error("Failure in SetAuthTokenforSlackIntegration")
		return err
	}
	return nil
}
func (store *MemSQL) GetSlackAuthToken(projectID int64, agentUUID string) (model.SlackAccessTokens, error) {
	db := C.GetServices().Db
	var agent model.Agent
	err := db.Where("uuid = ?", agentUUID).Find(&agent).Error
	if err != nil {
		log.WithError(err).Error("Failure in GetSlackAuthToken")
		return model.SlackAccessTokens{}, err
	}
	var token model.SlackAuthTokens

	if IsEmptyPostgresJsonb(agent.SlackAccessTokens) {
		return model.SlackAccessTokens{}, errors.New("No slack auth token found")
	}

	err = U.DecodePostgresJsonbToStructType(agent.SlackAccessTokens, &token)
	if err != nil && err.Error() != "Empty jsonb object" {
		log.WithError(err).Error("Failure in GetSlackAuthToken")
		return model.SlackAccessTokens{}, err
	}
	if err != nil && err.Error() == "Empty jsonb object" {
		return model.SlackAccessTokens{}, errors.New("No slack auth token found")
	}
	if _, ok := token[projectID]; !ok {
		return model.SlackAccessTokens{}, errors.New("Slack token not found.")
	}
	return token[projectID], nil

}
func (store *MemSQL) DeleteSlackIntegration(projectID int64, agentUUID string) error {
	db := C.GetServices().Db
	var agent model.Agent
	err := db.Where("uuid = ?", agentUUID).Find(&agent).Error
	if err != nil {
		log.WithError(err).Error("Failure in DeleteSlackIntegration")
		return err
	}
	var token model.SlackAuthTokens
	err = U.DecodePostgresJsonbToStructType(agent.SlackAccessTokens, &token)
	if err != nil && err.Error() != "Empty jsonb object" {
		log.WithError(err).Error("Failure in DeleteSlackIntegration")
		return err
	}
	if err != nil && err.Error() == "Empty jsonb object" {
		return errors.New("No slack auth token found")
	}
	var newToken model.SlackAuthTokens
	newToken = make(map[int64]model.SlackAccessTokens)
	for k, v := range token {
		if k != projectID {
			newToken[k] = v
		}
	}
	TokenJson, err := U.EncodeStructTypeToPostgresJsonb(newToken)
	if err != nil {
		log.WithError(err).Error("Failure in DeleteSlackIntegration")
		return err
	}
	// update the db
	err = db.Model(&model.Agent{}).Where("uuid = ?", agentUUID).Update("slack_access_tokens", TokenJson).Error
	if err != nil {
		log.WithError(err).Error("Failure in DeleteSlackIntegration")
		return err
	}
	return nil
}
func (store *MemSQL) GetAlertById(id string, projectID int64) (model.Alert, int) {
	logFields := log.Fields{
		"id":         id,
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	var alert model.Alert
	if projectID == 0 {
		log.Error("Invalid project ID.")
		return model.Alert{}, http.StatusBadRequest
	}
	if id == "" {
		log.Error("Invalid ID for alert.")
		return model.Alert{}, http.StatusBadRequest
	}
	db := C.GetServices().Db
	err := db.Where("project_id = ? AND is_deleted != ? AND id = ?", projectID, true, id).Find(&alert).Error
	if err != nil {
		log.WithField("project_id", projectID).Warn(err)
		return model.Alert{}, http.StatusNotFound
	}

	return alert, http.StatusFound
}
func (store *MemSQL) GetAllAlerts(projectID int64, excludeSavedQueries bool) ([]model.Alert, int) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	alerts := make([]model.Alert, 0)
	if projectID == 0 {
		log.Error("Invalid project ID.")
		return make([]model.Alert, 0), http.StatusBadRequest
	}
	db := C.GetServices().Db
	if excludeSavedQueries {
		err := db.Where("project_id = ? AND alert_type != ? AND is_deleted != ?", projectID, 3, true).Find(&alerts).Error
		if err != nil {
			log.WithField("project_id", projectID).Warn(err)
			return make([]model.Alert, 0), http.StatusNotFound
		}
	} else {
		err := db.Where("project_id = ? AND alert_type = ? AND is_deleted != ?", projectID, 3, true).Find(&alerts).Error
		if err != nil {
			log.WithField("project_id", projectID).Warn(err)
			return make([]model.Alert, 0), http.StatusNotFound
		}
	}
	sort.Slice(alerts, func(i, j int) bool {
		return alerts[i].CreatedAt.After(alerts[j].CreatedAt)
	})
	return alerts, http.StatusFound
}

// Note: Currently keeping the implementation specific to kpi.
func (store *MemSQL) GetAlertNamesByProjectIdTypeAndName(projectID int64, nameOfQuery string) ([]string, int) {
	rAlertNames := make([]string, 0)
	alerts, statusCode := store.GetAllAlerts(projectID, true)
	if statusCode != http.StatusFound {
		return rAlertNames, statusCode
	}

	for _, alert := range alerts {
		if alert.AlertType == model.ALERT_TYPE_SINGLE_RANGE || alert.AlertType == model.ALERT_TYPE_MULTI_RANGE {
			_, _, kpiQuery, err := model.DecodeAndFetchAlertRelatedStructs(projectID, alert)
			if err != nil {
				log.WithField("kpiQuery", kpiQuery).WithField("alert", alert).Warn("Failed to decode and fetch alert - GetAlertNamesByProjectIdTypeAndName")
				return rAlertNames, http.StatusInternalServerError
			}

			for _, metric := range kpiQuery.Metrics {
				if metric == nameOfQuery {
					rAlertNames = append(rAlertNames, alert.AlertName)
				}
			}
		} else if alert.AlertType == model.ALERT_TYPE_QUERY_SHARING {
			query, status := store.GetQueryWithQueryId(alert.ProjectID, alert.QueryID)
			if status != http.StatusFound {
				log.WithField("err_code", status).Error("Query not found for id ", alert.QueryID)
				continue
			}
			class, errMsg := store.GetQueryClassFromQueries(*query)
			if errMsg != "" {
				log.Error("Class not Found for queryID ", errMsg)
				continue
			}
			if class == model.QueryClassKPI {
				kpiQueryGroup := model.KPIQueryGroup{}
				U.DecodePostgresJsonbToStructType(&query.Query, &kpiQueryGroup)
				if len(kpiQueryGroup.Queries) == 0 {
					log.Error("Query failed. Empty query group.")
					continue
				}
				for _, kpiQuery := range kpiQueryGroup.Queries {
					for _, metric := range kpiQuery.Metrics {
						if metric == nameOfQuery {
							rAlertNames = append(rAlertNames, alert.AlertName)
						}
					}
				}
			}
		}
	}

	return rAlertNames, http.StatusFound
}

func (store *MemSQL) GetAlertNamesByProjectIdTypeAndNameAndPropertyMappingName(projectID int64, reqID, nameOfPropertyMappings string) ([]string, int) {
	rAlertNames := make([]string, 0)
	alerts, statusCode := store.GetAllAlerts(projectID, true)
	if statusCode != http.StatusFound {
		return rAlertNames, statusCode
	}

	for _, alert := range alerts {
		if alert.AlertType == model.ALERT_TYPE_SINGLE_RANGE || alert.AlertType == model.ALERT_TYPE_MULTI_RANGE {
			_, _, kpiQuery, err := model.DecodeAndFetchAlertRelatedStructs(projectID, alert)
			if err != nil {
				log.WithField("kpiQuery", kpiQuery).WithField("alert", alert).Warn("Failed to decode and fetch alert - GetAlertNamesByProjectIdTypeAndNameAndPropertyMappingName")
				return rAlertNames, http.StatusInternalServerError
			}

			if kpiQuery.CheckIfPropertyMappingNameIsPresent(nameOfPropertyMappings) {
				rAlertNames = append(rAlertNames, alert.AlertName)
			}
		} else if alert.AlertType == model.ALERT_TYPE_QUERY_SHARING {
			query, status := store.GetQueryWithQueryId(alert.ProjectID, alert.QueryID)
			if status != http.StatusFound {
				log.WithField("err_code", status).Error("Query not found for id ", alert.QueryID)
				continue
			}
			class, errMsg := store.GetQueryClassFromQueries(*query)
			if errMsg != "" {
				log.Error("Class not Found for queryID ", errMsg)
				continue
			}
			if class == model.QueryClassKPI {
				kpiQueryGroup := model.KPIQueryGroup{}
				U.DecodePostgresJsonbToStructType(&query.Query, &kpiQueryGroup)
				if len(kpiQueryGroup.Queries) == 0 {
					log.Error("Query failed. Empty query group.")
					continue
				}
				if kpiQueryGroup.CheckIfPropertyMappingNameIsPresent(nameOfPropertyMappings) {
					rAlertNames = append(rAlertNames, alert.AlertName)
				}
			}
		}
	}

	if len(rAlertNames) == 0 {
		return rAlertNames, http.StatusNotFound
	}
	return rAlertNames, http.StatusFound
}

func (store *MemSQL) DeleteAlert(id string, projectID int64) (int, string) {
	logFields := log.Fields{
		"id":         id,
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if projectID == 0 {
		log.Error("Invalid project ID.")
		return http.StatusBadRequest, "Invalid project id"
	}
	if id == "" {
		log.Error("Invalid id for alert")
		return http.StatusBadRequest, "Invalid id for alert"
	}
	db := C.GetServices().Db
	err := db.Table("alerts").Where("project_id = ? AND id = ?", projectID, id).Updates(map[string]interface{}{"is_deleted": true, "updated_at": time.Now().UTC()}).Error
	if err != nil {
		log.WithError(err).WithField("project_id", projectID).Error(err)
		return http.StatusInternalServerError, err.Error()
	}
	return http.StatusAccepted, ""
}
func (store *MemSQL) UpdateAlert(projectID int64, alertID string, alert model.Alert) (model.Alert, int, string) {
	logFields := log.Fields{
		"project_id": projectID,
		"alert_id":   alertID,
	}
	updatedFields := map[string]interface{}{
		"alert_name":          alert.AlertName,
		"alert_configuration": alert.AlertConfiguration,
		"alert_description":   alert.AlertDescription,
		"updated_at":          time.Now().UTC(),
	}
	alert.ID = alertID
	// validating alert config
	isValid, status, errMsg := store.validateAlertBody(projectID, alert)
	if !isValid {
		return model.Alert{}, status, errMsg
	}
	db := C.GetServices().Db
	err := db.Table("alerts").Where("project_id = ? AND id = ?", projectID, alertID).Updates(updatedFields).Error
	if err != nil {
		log.WithFields(logFields).WithError(err).WithField("project_id", projectID).Error(
			"Failed to update rule object.")
		return model.Alert{}, http.StatusInternalServerError, "Internal server error"
	}
	return alert, http.StatusAccepted, ""
}
func (store *MemSQL) UpdateAlertStatus(lastAlertSent bool) (int, string) {
	db := C.GetServices().Db
	err := db.Table("alerts").Where("id = ?", lastAlertSent).Updates(map[string]interface{}{"last_alert_sent": lastAlertSent, "updated_at": time.Now().UTC()}).Error
	if err != nil {
		log.WithError(err).Error("Failure in UpdateAlertStatus")
		return http.StatusInternalServerError, err.Error()
	}
	return http.StatusAccepted, ""
}
func (store *MemSQL) CreateAlert(projectID int64, alert model.Alert) (model.Alert, int, string) {
	logFields := log.Fields{
		"project_id": projectID,
		"alert":      alert,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	alertRecord := model.Alert{
		ID:                 U.GetUUID(),
		ProjectID:          projectID,
		AlertName:          alert.AlertName,
		CreatedBy:          alert.CreatedBy,
		AlertType:          alert.AlertType,
		QueryID:            alert.QueryID,
		AlertDescription:   alert.AlertDescription,
		AlertConfiguration: alert.AlertConfiguration,
		IsDeleted:          false,
		CreatedAt:          time.Now().UTC(),
		UpdatedAt:          time.Now().UTC(),
	}
	isValid, status, errMsg := store.validateAlertBody(projectID, alertRecord)
	if !isValid {
		return model.Alert{}, status, errMsg
	}
	db := C.GetServices().Db
	err := db.Create(&alertRecord).Error
	if err != nil {
		logCtx.WithError(err).WithField("project_id", alertRecord.ProjectID).Error(
			"Failed to insert alert record.")
		return model.Alert{}, http.StatusInternalServerError, "Internal server error"
	}
	return alertRecord, http.StatusCreated, ""

}

func (store *MemSQL) validateAlertBody(projectID int64, alert model.Alert) (bool, int, string) {
	logFields := log.Fields{
		"project_id": projectID,
		"alert":      alert,
	}
	logCtx := log.WithFields(logFields)

	var alertDescription model.AlertDescription
	var alertConfiguration model.AlertConfiguration
	err := U.DecodePostgresJsonbToStructType(alert.AlertDescription, &alertDescription)
	if err != nil {
		return false, http.StatusInternalServerError, "failed to decode jsonb to alert description"
	}
	// validations
	// - Check for valid operator
	// - Check if the KPI/Metric is valid TODO
	// - Check if the date range is valid for both type 1 and type 2
	// skip validations for saved query report sharing i.e type 3
	skipValidation := false

	if alert.AlertType == model.ALERT_TYPE_QUERY_SHARING {
		skipValidation = true
	}
	if alertDescription.Name == "" && alert.AlertType != model.ALERT_TYPE_QUERY_SHARING {
		logCtx.Error("Invalid alert name")
		return false, http.StatusBadRequest, "Invalid alert name"
	}
	err = U.DecodePostgresJsonbToStructType(alert.AlertConfiguration, &alertConfiguration)
	if err != nil {
		return false, http.StatusInternalServerError, "failed to decode jsonb to alert configuration"
	}
	if !alertConfiguration.IsEmailEnabled && !alertConfiguration.IsSlackEnabled {
		logCtx.Error("Select at least one notification method")
		return false, http.StatusBadRequest, "Select at least one notification method"
	}
	if alertConfiguration.IsEmailEnabled && len(alertConfiguration.Emails) == 0 {
		logCtx.Error("empty email list")
		return false, http.StatusBadRequest, "empty email list"
	}
	isSlackIntegrated, errCode := store.IsSlackIntegratedForProject(projectID, alert.CreatedBy)
	if errCode != http.StatusOK {
		return false, errCode, "failed to check slack integration"
	}
	if alertConfiguration.IsSlackEnabled && !isSlackIntegrated {
		logCtx.Error("Slack integration is not enabled for this project")
		return false, http.StatusBadRequest, "Slack integration is not enabled for this project"
	}
	_, err = store.checkIfDuplicateAlertNameExists(projectID, alert.AlertName, alert.ID)
	if err != nil {
		logCtx.WithError(err).Error("Failed to create alert")
		return false, http.StatusBadRequest, err.Error()
	}

	if !skipValidation {
		if !store.isValidOperator(alertDescription.Operator) {
			return false, http.StatusBadRequest, "Invalid Operator for Alert"
		}
	}
	if alert.AlertType == model.ALERT_TYPE_SINGLE_RANGE {
		if !store.isValidDateRange(alertDescription.DateRange) {
			return false, http.StatusBadRequest, "Invalid Date Range"
		}
	} else if alert.AlertType == model.ALERT_TYPE_MULTI_RANGE {
		if !store.isValidDateRangeAndComparedTo(alertDescription.DateRange, alertDescription.ComparedTo) {
			return false, http.StatusBadRequest, "Invalid Date Range"
		}
	} else {
		if alert.AlertType != model.ALERT_TYPE_QUERY_SHARING {
			return false, http.StatusBadRequest, "Invalid Alert Type"
		}
	}
	return true, http.StatusOK, ""
}
func (store *MemSQL) checkIfDuplicateAlertNameExists(projectID int64, alertName string, alertID string) (allowed bool, err error) {
	alerts, status := store.GetAllAlerts(projectID, true)
	if status != http.StatusFound {
		return false, errors.New(fmt.Sprintf("Failed to get Alerts for project ID %d", projectID))
	}
	for _, alert := range alerts {
		if strings.ToLower(alert.AlertName) == strings.ToLower(alertName) {
			// should allow if updating the same alert with same name
			if alert.ID == alertID {
				continue
			}
			return false, errors.New(fmt.Sprintf("Alert with name %s already exists ", alertName))
		}
	}
	return true, nil
}
func (store *MemSQL) isValidOperator(operator string) bool {
	for _, op := range model.ValidOperators {
		if op == operator {
			return true
		}
	}
	return false
}
func (store *MemSQL) isValidDateRange(dateRange string) bool {
	for _, dr := range model.ValidDateRanges {
		if dr == dateRange {
			return true
		}
	}
	return false
}
func (store *MemSQL) isValidDateRangeAndComparedTo(dateRange, comparedTo string) bool {
	flag := false
	for _, dr := range model.ValidDateRanges {
		if dr == dateRange {
			flag = true
			break
		}
	}
	if flag {
		for _, dr := range model.ValidDateRangeComparisions {
			if dr == comparedTo {
				return true
			}
		}
	}
	return false
}
func (store *MemSQL) getNameForAlert(metric, operator, value string) string {
	AlertName := fmt.Sprintf("%s%s%s", metric, operator, value)
	return AlertName
}
func IsEmptyPostgresJsonb(jsonb *postgres.Jsonb) bool {
	if jsonb == nil {
		log.Info("jsonb is nil")
		return true
	}
	strJson := string((*jsonb).RawMessage)
	return strJson == "" || strJson == "null"
}
