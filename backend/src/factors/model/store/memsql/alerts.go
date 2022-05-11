package memsql

import (
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) SetAuthTokenforSlackIntegration(projectID uint64, agentUUID string, authTokens model.SlackAccessTokens) error {
	db := C.GetServices().Db
	var agent model.Agent
	err := db.Where("uuid = ?", agentUUID).Find(&agent).Error
	if err != nil {
		log.WithFields(log.Fields{
			"uuid": agentUUID,
		}).Error(err)
		return err
	}
	var token model.SlackAuthTokens
	token[projectID] = authTokens
	// convert token to json
	TokenJson, err := U.EncodeStructTypeToPostgresJsonb(token)
	if err != nil {
		log.Error(err)
		return err
	}
	agent.SlackAuthTokens = TokenJson
	err = db.Update(&agent).Error
	if err != nil {
		log.WithFields(log.Fields{
			"uuid": agentUUID,
		}).Error(err)
		return err
	}
	return nil
}
func (store *MemSQL) GetSlackAuthToken(agentUUID string) (model.SlackAuthTokens, error) {
	db := C.GetServices().Db
	var agent model.Agent
	err := db.Where("uuid = ?", agentUUID).Find(&agent).Error
	if err != nil {
		log.WithFields(log.Fields{
			"uuid": agentUUID,
		}).Error(err)
		return nil, err
	}
	var token model.SlackAuthTokens
	err = U.DecodePostgresJsonbToStructType(agent.SlackAuthTokens, &token)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return token, nil
}
func (store *MemSQL) DeleteSlackIntegration(agentUUID string) error {
	db := C.GetServices().Db
	var agent model.Agent
	err := db.Where("uuid = ?", agentUUID).Find(&agent).Error
	if err != nil {
		log.WithFields(log.Fields{
			"uuid": agentUUID,
		}).Error(err)
		return err
	}
	agent.SlackAuthTokens = nil
	err = db.Save(&agent).Error
	if err != nil {
		log.WithFields(log.Fields{
			"uuid": agentUUID,
		}).Error(err)
		return err
	}
	return nil
}
func (store *MemSQL) GetAlertById(id string, projectID uint64) (model.Alert, int) {
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
func (store *MemSQL) GetAllAlerts(projectID uint64) ([]model.Alert, int) {
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
	err := db.Where("project_id = ? AND is_deleted != ?", projectID, true).Find(&alerts).Error
	if err != nil {
		log.WithField("project_id", projectID).Warn(err)
		return make([]model.Alert, 0), http.StatusNotFound
	}
	return alerts, http.StatusFound
}
func (store *MemSQL) DeleteAlert(id string, projectID uint64) (int, string) {
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
		log.WithField("project_id", projectID).Error(err)
		return http.StatusInternalServerError, err.Error()
	}
	return http.StatusAccepted, ""
}
func (store *MemSQL) UpdateAlert(alert model.Alert) (int, string) {
	db := C.GetServices().Db
	alert.UpdatedAt = time.Now().UTC()
	err := db.Update(&alert).Error
	if err != nil {
		log.WithField("project_id", alert.ProjectID).Error(err)
		return http.StatusInternalServerError, err.Error()
	}
	return http.StatusAccepted, ""
}
func (store *MemSQL) CreateAlert(projectID uint64, alert model.Alert) (model.Alert, int, string) {
	logFields := log.Fields{
		"project_id": projectID,
		"alert":      alert,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	var alertDescription model.AlertDescription
	var alertConfiguration model.AlertConfiguration
	err := U.DecodePostgresJsonbToStructType(alert.AlertDescription, &alertDescription)
	if err != nil {
		return model.Alert{}, http.StatusInternalServerError, "failed to decode jsonb to alert description"
	}
	// validations
	// - Check for valid operator
	// - Check if the KPI/Metric is valid TODO
	// - Check if the date range is valid for both type 1 and type 2
	if alertDescription.Name == "" {
		logCtx.Error("Invalid alert name")
		return model.Alert{}, http.StatusBadRequest, "Invalid alert name"
	}
	err = U.DecodePostgresJsonbToStructType(alert.AlertConfiguration, &alertConfiguration)
	if err != nil {
		return model.Alert{}, http.StatusInternalServerError, "failed to decode jsonb to alert configuration"
	}
	if alertConfiguration.IsEmailEnabled && len(alertConfiguration.Emails) == 0 {
		logCtx.Error("empty email list")
		return model.Alert{}, http.StatusBadRequest, "empty email list"
	}

	if !store.isValidOperator(alertDescription.Operator) {
		return model.Alert{}, http.StatusBadRequest, "Invalid Operator for Alert"
	}
	if alert.AlertType == model.ALERT_TYPE_SINGLE_RANGE {
		if !store.isValidDateRange(alertDescription.DateRange) {
			return model.Alert{}, http.StatusBadRequest, "Invalid Date Range"
		}
	} else if alert.AlertType == model.ALERT_TYPE_MULTI_RANGE {
		if !store.isValidDateRangeAndComparedTo(alertDescription.DateRange, alertDescription.ComparedTo) {
			return model.Alert{}, http.StatusBadRequest, "Invalid Date Range"
		}
	} else {
		return model.Alert{}, http.StatusBadRequest, "Invalid Alert Type"
	}
	alertRecord := model.Alert{
		ID:                 U.GetUUID(),
		ProjectID:          projectID,
		AlertName:          store.getNameForAlert(alertDescription.Name, alertDescription.Operator, alertDescription.Value),
		CreatedBy:          alert.CreatedBy,
		AlertType:          alert.AlertType,
		AlertDescription:   alert.AlertDescription,
		AlertConfiguration: alert.AlertConfiguration,
		IsDeleted:          false,
		CreatedAt:          time.Now().UTC(),
		UpdatedAt:          time.Now().UTC(),
	}
	db := C.GetServices().Db
	err = db.Create(&alertRecord).Error
	if err != nil {
		logCtx.WithError(err).WithField("project_id", alertRecord.ProjectID).Error(
			"Failed to insert alert record.")
		return model.Alert{}, http.StatusInternalServerError, "Internal server error"
	}
	return alertRecord, http.StatusCreated, ""

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
