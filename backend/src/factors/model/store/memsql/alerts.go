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

const (
	KPI = "kpi_alert"
)

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

func (store *MemSQL) GetAlertByProjectId(projectId int64, excludeSavedQueries bool) ([]model.AlertInfo, int) {
	logFields := log.Fields{
		"project_id": projectId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	
	displayAlerts := make([]model.AlertInfo, 0)

	alerts, errCode := store.GetAllAlerts(projectId, excludeSavedQueries) 
	if errCode != http.StatusFound {
		return displayAlerts, errCode
	}
	
	alertArray := store.convertAlertToAlertInfo(alerts)
	return alertArray, http.StatusFound
}

func (store *MemSQL) convertAlertToAlertInfo(list []model.Alert) []model.AlertInfo {

	res := make([]model.AlertInfo, 0)

	for _, obj := range list {
		var alertConfig model.AlertConfiguration
		err := U.DecodePostgresJsonbToStructType(obj.AlertConfiguration, &alertConfig)
		if err != nil {
			log.WithError(err).Error("Problem deserializing event_trigger_alerts query.")
			return nil
		}
		deliveryOption := ""
		if alertConfig.IsSlackEnabled {
			deliveryOption += "Slack "
		}
		if alertConfig.IsEmailEnabled {
			if deliveryOption == "" {
				deliveryOption += "Email"
			} else {
				deliveryOption += "& Email"
			}
		}
		if alertConfig.IsTeamsEnabled {
			if deliveryOption == "" {
				deliveryOption += "Teams"
			} else {
				deliveryOption += "& Teams"
			}
		}

		alertJson, err := U.EncodeStructTypeToPostgresJsonb(&obj)
		if err != nil {
			return nil
		}

		e := model.AlertInfo{
			ID:              obj.ID,
			Title:           obj.AlertName,
			DeliveryOptions: deliveryOption,
			LastFailDetails: nil,
			Status:          model.Active,
			Alert:           alertJson,
			Type:            KPI,
			CreatedAt:       obj.CreatedAt,
		}
		res = append(res, e)
	}
	return res
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
	// validations
	// - Check for valid operator
	// - Check if the KPI/Metric is valid TODO
	// - Check if the date range is valid for both type 1 and type 2
	// skip validations for saved query report sharing i.e type 3
	skipValidation := false

	if alert.AlertType == model.ALERT_TYPE_QUERY_SHARING {
		skipValidation = true
	}
	err := U.DecodePostgresJsonbToStructType(alert.AlertConfiguration, &alertConfiguration)
	if err != nil {
		return false, http.StatusInternalServerError, "failed to decode jsonb to alert configuration"
	}
	if !alertConfiguration.IsEmailEnabled && !alertConfiguration.IsSlackEnabled && !alertConfiguration.IsTeamsEnabled{
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
	isTeamsIntegrated, errCode := store.IsTeamsIntegratedForProject(projectID, alert.CreatedBy)
	if errCode != http.StatusOK {
		return false, errCode, "failed to check teams integration"
	}
	if alertConfiguration.IsTeamsEnabled && !isTeamsIntegrated {
		logCtx.Error("Teams integration is not enabled for this project")
		return false, http.StatusBadRequest, "Teams integration is not enabled for this project"
	}
	if alertConfiguration.IsSlackEnabled {

		slackChannels := make(map[string][]model.SlackChannel)
		err = U.DecodePostgresJsonbToStructType(alertConfiguration.SlackChannelsAndUserGroups, &slackChannels)
		if err != nil {
			log.WithError(err).Error("failed to decode slack channels")
			return false, http.StatusBadRequest, "failed to decode slack channels"
		}

		if len(slackChannels[alert.CreatedBy]) == 0 {
			log.WithError(err).Error("Empty Slack Channel List")
			return false, http.StatusBadRequest, "Empty Slack Channel List"
		}
	}
	_, err = store.checkIfDuplicateAlertNameExists(projectID, alert.AlertName, alert.ID)
	if err != nil {
		logCtx.WithError(err).Error("Failed to create alert")
		return false, http.StatusBadRequest, err.Error()
	}
	err = U.DecodePostgresJsonbToStructType(alert.AlertDescription, &alertDescription)
	if err != nil {
		return false, http.StatusInternalServerError, "failed to decode jsonb to alert description"
	}

	if alertDescription.Name == "" && alert.AlertType != model.ALERT_TYPE_QUERY_SHARING {
		logCtx.Error("Invalid alert name")
		return false, http.StatusBadRequest, "Invalid alert name"
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
