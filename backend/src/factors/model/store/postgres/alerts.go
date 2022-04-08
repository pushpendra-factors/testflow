package postgres

import (
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

func (pg *Postgres) GetAlertById(id string, projectID uint64) (model.Alert, int) {
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
func (pg *Postgres) GetAllAlerts(projectID uint64) ([]model.Alert, int) {
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
func (pg *Postgres) UpdateAlert(id string, projectID uint64) (int, string) {
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
func (pg *Postgres) CreateAlert(projectID uint64, alert model.Alert) (model.Alert, int, string) {
	logFields := log.Fields{
		"project_id": projectID,
		"alert":      alert,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	var alertDescription model.AlertDescription
	err := U.DecodePostgresJsonbToStructType(alert.AlertDescription, &alertDescription)
	if err != nil {
		return model.Alert{}, http.StatusInternalServerError, "failed to decode jsonb to alert description"
	}
	if !pg.isValidOperator(alertDescription.Operator) {
		return model.Alert{}, http.StatusBadRequest, "Invalid Operator for Alert"
	}
	if alert.AlertType == model.ALERT_TYPE_SINGLE_RANGE {
		if !pg.isValidDateRange(alertDescription.DateRange) {
			return model.Alert{}, http.StatusBadRequest, "Invalid Date Range"
		}
	} else if alert.AlertType == model.ALERT_TYPE_MULTI_RANGE {
		if !pg.isValidDateRangeAndComparedTo(alertDescription.DateRange, alertDescription.ComparedTo) {
			return model.Alert{}, http.StatusBadRequest, "Invalid Date Range"
		}
	} else {
		return model.Alert{}, http.StatusBadRequest, "Invalid Alert Type"
	}
	alertRecord := model.Alert{
		ProjectID:          projectID,
		AlertName:          pg.getNameForAlert(alertDescription.Name, alertDescription.Operator, alertDescription.Value),
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
func (pg *Postgres) isValidOperator(operator string) bool {
	fmt.Println(operator)
	for _, op := range model.ValidOperators {
		if op == operator {
			return true
		}
	}
	return false
}
func (pg *Postgres) isValidDateRange(dateRange string) bool {
	for _, dr := range model.ValidDateRanges {
		if dr == dateRange {
			return true
		}
	}
	return false
}
func (pg *Postgres) isValidDateRangeAndComparedTo(dateRange, comparedTo string) bool {
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
func (pg *Postgres) getNameForAlert(metric, operator, value string) string {
	AlertName := fmt.Sprintf("%s%s%s", metric, operator, value)
	return AlertName
}
