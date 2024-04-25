package memsql

import (
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) GetAllWorkflowTemplates() ([]model.AlertTemplate, int) {

	db := C.GetServices().Db

	var alertTemplates []model.AlertTemplate
	err := db.Where("is_deleted = ?", false).
		Where("is_workflow = ?", true).
		Order("id").Find(&alertTemplates).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return alertTemplates, http.StatusNotFound
		}
		log.WithError(err).Error("Failed to get workflow templates.")
		return alertTemplates, http.StatusInternalServerError
	}

	return alertTemplates, http.StatusOK
}

func (store *MemSQL) GetAllWorklfowsByProject(projectID int64) ([]model.WorkflowAlertBody, int, error) {
	if projectID == 0 {
		return nil, http.StatusBadRequest, fmt.Errorf("invalid parameter")
	}

	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	db := C.GetServices().Db
	workflows := make([]model.Workflow, 0)
	wfAlerts := make([]model.WorkflowAlertBody, 0)
	err := db.Where("project_id = ?", projectID).Where("is_deleted = ?", false).
		Order("created_at DESC").Limit(ListLimit).Find(&workflows).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return wfAlerts, http.StatusNotFound, err
		}
		log.WithError(err).Error("Failed to fetch rows of workflows")
		return nil, http.StatusInternalServerError, err
	}

	for _, wf := range workflows {
		var alert model.WorkflowAlertBody 
		err := U.DecodePostgresJsonbToStructType(wf.AlertBody, &alert)
		if err != nil {
			log.WithError(err).Error("Failed to decode alert in the workflow object")
			continue
		}
		wfAlerts = append(wfAlerts, alert)
	}

	return wfAlerts, http.StatusFound, nil
}

func (store *MemSQL) GetWorkflowById(projectID int64, id string) (*model.Workflow, int, error) {
	if projectID == 0 || id == "" {
		return nil, http.StatusBadRequest, fmt.Errorf("invalid parameter")
	}

	logFields := log.Fields{
		"project_id":  projectID,
		"workflow_id": id,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	db := C.GetServices().Db
	var workflow model.Workflow

	err := db.Where("project_id = ?", projectID).
		Where("id = ?", id).
		Where("is_deleted = ?", false).
		Order("created_at DESC").Find(&workflow).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound, err
		}
		log.WithError(err).Error("Failed to fetch requested workflows")
		return nil, http.StatusInternalServerError, err
	}

	return &workflow, http.StatusFound, nil
}

func (store *MemSQL) CreateWorkflow(projectID int64, agentID string, alertBody model.WorkflowAlertBody) (*model.Workflow, int, error) {
	if projectID == 0 || agentID == "" {
		return nil, http.StatusBadRequest, fmt.Errorf("invalid parameter")
	}

	logFields := log.Fields{
		"project_id": projectID,
		"agent_id":   agentID,
		"workflow":   alertBody,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)
	db := C.GetServices().Db

	var workflow model.Workflow
	transTime := U.TimeNowZ()
	id := U.GetUUID()

	alertJson, err := U.EncodeStructTypeToPostgresJsonb(alertBody)
	if err != nil {
		logCtx.WithError(err).Error("Failed to encode workflow body")
		return nil, http.StatusInternalServerError, err
	}

	workflow = model.Workflow{
		ID:        id,
		ProjectID: projectID,
		Name:      alertBody.Title,
		AlertBody: alertJson,
		CreatedBy: agentID,
		CreatedAt: transTime,
		UpdatedAt: transTime,
		IsDeleted: false,
	}

	if err := db.Create(&workflow).Error; err != nil {
		logCtx.WithError(err).Error("Failed to create workflow")
		return nil, http.StatusInternalServerError, err
	}

	return &workflow, http.StatusFound, nil
}

func (store *MemSQL) UpdateWorkflow(projectID int64, id, agentID string, alertBody model.Workflow) (*model.Workflow, int, error) {
	if projectID == 0 || id == "" || agentID == "" {
		return nil, http.StatusBadRequest, fmt.Errorf("invalid parameter")
	}

	logFields := log.Fields{
		"project_id":  projectID,
		"workflow_id": id,
		"agent_id":    agentID,
		"workflow":    alertBody,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)
	db := C.GetServices().Db

	var workflow model.Workflow
	transTime := U.TimeNowZ()

	workflow = model.Workflow{
		ID:        id,
		ProjectID: projectID,
		Name:      alertBody.Name,
		AlertBody: alertBody.AlertBody,
		CreatedBy: agentID,
		CreatedAt: transTime,
		UpdatedAt: transTime,
		IsDeleted: false,
	}

	if err := db.Create(&workflow).Error; err != nil {
		logCtx.WithError(err).Error("Create failed in db")
		return &workflow, http.StatusInternalServerError, err
	}

	return &workflow, http.StatusFound, nil
}

func (store *MemSQL) DeleteWorkflow(projectID int64, id, agentID string) (int, error) {
	if projectID == 0 || id == "" || agentID == "" {
		return http.StatusBadRequest, fmt.Errorf("invalid parameter")
	}

	logFields := log.Fields{
		"project_id":  projectID,
		"workflow_id": id,
		"agent_id":    agentID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)
	db := C.GetServices().Db
	transTime := U.TimeNowZ()

	err := db.Model(&model.Workflow{}).
		Where("id = ?", id).
		Where("project_id = ?", projectID).
		Updates(map[string]interface{}{"is_deleted": true, "updated_at": transTime}).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to delete workflow.")
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}