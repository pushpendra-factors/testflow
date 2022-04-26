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
	name string, activtyType int, actorType int, actorID string, externalActivityID string, timestamp int64) (int, error) {
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

	if projectID == 0 || name == "" || externalActivityID == "" ||
		activtyType <= 0 || actorType <= 0 || actorID == "" {
		logCtx.Error("Missing required parameters.")
		return http.StatusBadRequest, errors.New("missing required field project_id, name, activity_type, actor_type, actor_id, external_activity_id")
	}

	var activity model.CRMActivity
	db := C.GetServices().Db
	err := db.Model(&model.CRMActivity{}).Where("project_id = ? AND source = ? AND external_activity_id = ? "+
		"AND name = ? AND type = ? AND actor_type = ? AND actor_id = ? AND timestamp =?",
		projectID, source, externalActivityID, name, activtyType, actorType, actorID, timestamp).Select("id").Limit(1).Find(&activity).Error
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
		crmActivity.Type <= 0 || crmActivity.ActorType <= 0 || crmActivity.ActorID == "" || crmActivity.ExternalActivityID == "" {
		logCtx.Error("Missing required parameters")
		return http.StatusBadRequest, errors.New("missing required fields project_id, properties, type, actor_type, actor_id,external_activity_id")
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
		crmActivity.Type, crmActivity.ActorType, crmActivity.ActorID, crmActivity.ExternalActivityID, crmActivity.Timestamp)
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

func (store *MemSQL) GetCRMActivityInOrderForSync(projectID uint64, source model.CRMSource) ([]model.CRMActivity, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"source":     source,
	}

	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)
	if projectID == 0 || source == 0 {
		logCtx.Error("Invalid parameters")
		return nil, http.StatusBadRequest
	}

	var crmActivity []model.CRMActivity
	db := C.GetServices().Db
	err := db.Model(&model.CRMActivity{}).Where("project_id = ? AND source = ? AND synced = false ",
		projectID, source).Order("timestamp, created_at").Find(&crmActivity).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}

		logCtx.WithError(err).Error("Failed to get crm activity records for sync.")
		return nil, http.StatusInternalServerError
	}

	if len(crmActivity) == 0 {
		return nil, http.StatusNotFound
	}

	return crmActivity, http.StatusFound
}

func (store *MemSQL) UpdateCRMActivityAsSynced(projectID uint64, source model.CRMSource, crmActivity *model.CRMActivity, syncID, userID string) (*model.CRMActivity, int) {

	logFields := log.Fields{
		"project_id":           projectID,
		"source":               source,
		"name":                 crmActivity.Name,
		"actor_type":           crmActivity.ActorType,
		"actor_id":             crmActivity.ActorID,
		"sync_id":              syncID,
		"user_id":              userID,
		"id":                   crmActivity.ID,
		"activity_type":        crmActivity.Type,
		"timestamp":            crmActivity.Timestamp,
		"external_activity_id": crmActivity.ExternalActivityID,
	}

	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)
	if projectID == 0 || source == 0 || crmActivity.ID == "" || crmActivity.Name == "" ||
		crmActivity.ActorType == 0 || crmActivity.ActorID == "" {
		logCtx.Error("Invalid parameters")
		return nil, http.StatusBadRequest
	}

	updates := make(map[string]interface{})
	updates["synced"] = true

	if syncID != "" {
		updates["sync_id"] = syncID
	}

	if userID != "" {
		updates["user_id"] = userID
	}

	db := C.GetServices().Db
	err := db.Model(&model.CRMActivity{}).Where("project_id = ? AND source = ? AND external_activity_id = ? AND id = ? AND name = ? "+
		"AND actor_type = ? AND actor_id = ? AND type = ? AND timestamp = ? ",
		projectID, source, crmActivity.ExternalActivityID, crmActivity.ID, crmActivity.Name, crmActivity.ActorType,
		crmActivity.ActorID, crmActivity.Type, crmActivity.Timestamp).Updates(updates).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to update crm activity as synced.")
		return nil, http.StatusInternalServerError
	}
	crmActivity.SyncID = syncID
	crmActivity.UserID = userID

	return crmActivity, http.StatusAccepted
}

func (store *MemSQL) GetActivitiesDistinctEventNamesByType(projectID uint64, objectTypes []int) (map[int][]string, int) {
	logFields := log.Fields{"project_id": projectID, "object_types": objectTypes}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)

	if projectID == 0 || len(objectTypes) < 1 {
		logCtx.Error("Invalid parameters.")
		return nil, http.StatusBadRequest
	}

	db := C.GetServices().Db
	var distinctNames []struct {
		Name string
		Type int
	}

	err := db.Table("crm_activities").Where("project_id = ? AND type = ?", projectID, objectTypes).
		Select("DISTINCT(name) as name, type").Find(&distinctNames).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}

		logCtx.WithError(err).Error("Failed to get activity names by type.")
		return nil, http.StatusInternalServerError
	}

	if len(distinctNames) < 1 {
		logCtx.Error("Failed to get name for activites.")
		return nil, http.StatusNotFound
	}

	typeName := make(map[int][]string, 0)
	for i := range distinctNames {
		if _, exist := typeName[distinctNames[i].Type]; !exist {
			typeName[distinctNames[i].Type] = make([]string, 0)
		}

		typeName[distinctNames[i].Type] = append(typeName[distinctNames[i].Type], distinctNames[i].Name)
	}

	return typeName, http.StatusFound
}
