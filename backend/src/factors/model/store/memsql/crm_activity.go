package memsql

import (
	"errors"
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

// isExistActivty check for existing activity by activity name, type, actor type , actor id and timestamp
func isExistActivty(projectID uint64, source model.CRMSource,
	name string, activtyType int, actorType int, actorID string, timestamp int64) (int, error) {
	logFields := log.Fields{
		"project_id":    projectID,
		"source":        source,
		"name":          name,
		"activity_type": activtyType,
		"actor_type":    actorType,
		"actor_id":      actorID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)
	if !model.AllowedCRMBySource(source) {
		logCtx.Error("Invalid source.")
		return http.StatusBadRequest, errors.New("invalid source")
	}

	if projectID == 0 || name == "" ||
		activtyType <= 0 || actorType <= 0 || actorID == "" {
		logCtx.Error("Missing required parameters.")
		return http.StatusBadRequest, errors.New("missing required field project_id, name, activity_type, actor_type, actor_id")
	}

	var activity model.CRMActivity
	db := C.GetServices().Db
	err := db.Model(&model.CRMActivity{}).Where("project_id = ? AND source = ? "+
		"AND name = ? AND type = ? AND actor_type = ? AND actor_id = ? AND timestamp =?",
		projectID, source, name, activtyType, actorType, actorID, timestamp).Select("id").Limit(1).Find(&activity).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return http.StatusNotFound, nil
		}

		logCtx.WithError(err).Error("Failed to get activity from crm_activities table.")
		return http.StatusInternalServerError, err
	}

	if activity.ID == "" {
		return http.StatusNotFound, nil
	}

	return http.StatusFound, nil
}

// CreateCRMActivity custom events which needs to be created for a user or a group
func (store *MemSQL) CreateCRMActivity(crmActivity *model.CRMActivity) (int, error) {
	logFields := log.Fields{
		"project_id": crmActivity.ProjectID,
		"source":     crmActivity.Source,
		"name":       crmActivity.Name,
		"timestamp":  crmActivity.Timestamp,
		"type":       crmActivity.Type,
		"actor_type": crmActivity.ActorType,
		"actor_id":   crmActivity.ActorID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)

	/*
		Type - refers to the entity type on which event will be based
		actorType and actorId refers the user or group for whom the activty will be created.
		For ex - user by email, user by salesforce lead id etc
		Properties - properties for the actvity event
	*/
	if crmActivity.ProjectID == 0 || crmActivity.Properties == nil ||
		crmActivity.Type <= 0 || crmActivity.ActorType <= 0 || crmActivity.ActorID == "" {
		logCtx.Error("Missing required parameters")
		return http.StatusBadRequest, errors.New("missing required fields project_id, properties, type, actor_type, actor_id")
	}

	if !model.AllowedCRMBySource(crmActivity.Source) {
		logCtx.Error("Invalid crm source")
		return http.StatusBadRequest, errors.New("missing crm source")
	}

	// activity id will be generated internally. Uniquness will be defined by name,type, actor_type, actor_id
	if crmActivity.ID != "" {
		logCtx.Error("Invalid id.")
		return http.StatusBadRequest, errors.New("missing id")
	}

	// Name of the activity event
	if crmActivity.Name == "" {
		logCtx.Error("Invalid activity name.")
		return http.StatusBadRequest, errors.New("missing name")
	}

	// Timestamp - time when activtity happened
	if crmActivity.Timestamp <= 0 {
		logCtx.Error("Missing activity timestamp.")
		return http.StatusBadRequest, errors.New("missing timstamp")
	}

	if U.IsEmptyPostgresJsonb(crmActivity.Properties) {
		logCtx.Error("Empty activty properties.")
		return http.StatusBadRequest, errors.New("missing properties")
	}

	status, err := isExistActivty(crmActivity.ProjectID, crmActivity.Source, crmActivity.Name,
		crmActivity.Type, crmActivity.ActorType, crmActivity.ActorID, crmActivity.Timestamp)
	if status != http.StatusNotFound {
		if status == http.StatusFound {
			return http.StatusConflict, nil
		}

		logCtx.WithError(err).Error("Failed to check for existing activity.")
		return status, err
	}

	crmActivity.ID = U.GetUUID()
	db := C.GetServices().Db
	err = db.Create(&crmActivity).Error
	if err != nil {
		if isDuplicateRecord(err) {
			return http.StatusConflict, nil
		}

		logCtx.WithError(err).Error("Failed to insert crm activity.")
		return http.StatusInternalServerError, errors.New("failed to insert activity record")
	}

	return http.StatusCreated, nil
}
