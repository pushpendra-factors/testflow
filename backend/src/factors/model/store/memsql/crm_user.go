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

// isExistCRMUserByID check for existing user by external user id, source, object type
func isExistCRMUserByID(projectID int64, source U.CRMSource, userType int, id string) (int, error) {
	logFields := log.Fields{
		"project_id": projectID,
		"source":     source,
		"type":       userType,
		"id":         id,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)
	if !model.AllowedCRMBySource(source) {
		logCtx.Error("Invalid source.")
		return http.StatusBadRequest, errors.New("invalid source")
	}

	if projectID == 0 || userType <= 0 || id == "" {
		logCtx.Error("Missing required parameters.")
		return http.StatusBadRequest, errors.New("missing required fields project_id, user_type, id, source")
	}

	var crmUser model.CRMUser
	db := C.GetServices().Db
	err := db.Model(&model.CRMUser{}).Where("project_id = ? AND source = ? "+
		"AND type = ? AND id = ? AND action = ? ",
		projectID, source, userType, id, model.CRMActionCreated).Select("id").Limit(1).Find(&crmUser).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return http.StatusNotFound, nil
		}

		logCtx.WithError(err).Error("Failed to get user from crm_user table.")
		return http.StatusInternalServerError, err
	}

	if crmUser.ID == "" {
		return http.StatusNotFound, nil
	}

	return http.StatusFound, nil
}

func (store *MemSQL) CreateCRMUser(crmUser *model.CRMUser) (int, error) {
	logFields := log.Fields{
		"project_id": crmUser.ProjectID,
		"source":     crmUser.Source,
		"id":         crmUser.ID,
		"user_type":  crmUser.Type,
		"properties": crmUser.Properties,
		"timestamp":  crmUser.Timestamp,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)

	if !model.AllowedCRMBySource(crmUser.Source) {
		logCtx.Error("Invalid source.")
		return http.StatusBadRequest, errors.New("invalid source")
	}

	/*
	 ID - External source user id, unique per user. Used to track user state by external id
	 Type - Object name defined by external source for user
	 Properties - user object properties as key and value
	*/
	if crmUser.ProjectID == 0 || crmUser.ID == "" || crmUser.Type <= 0 || crmUser.Properties == nil {
		logCtx.Error("Missing required parameters.")
		return http.StatusBadRequest, errors.New("missing required field project_id, id, type, properties")
	}

	// timestamp when the user state occured
	if crmUser.Timestamp <= 0 {
		logCtx.Error("Missing document timestamp.")
		return http.StatusBadRequest, errors.New("missing timstamp")
	}

	if U.IsEmptyPostgresJsonb(crmUser.Properties) {
		logCtx.Error("Empty document properties.")
		return http.StatusBadRequest, errors.New("empty properties")
	}

	// metadata should be used to store other information from source object if required for special case
	if crmUser.Metadata != nil && U.IsEmptyPostgresJsonb(crmUser.Metadata) {
		logCtx.Error("Empty document metadata.")
		return http.StatusBadRequest, errors.New("empty metadata")
	}

	status, err := isExistCRMUserByID(crmUser.ProjectID, crmUser.Source, crmUser.Type, crmUser.ID)
	if status == http.StatusInternalServerError {
		logCtx.WithError(err).WithField("err_code", status).Error("Failed to check existing user document.")
		return http.StatusInternalServerError, err
	}

	// first state of a user should be stored as action created representing new user, followed by only updated
	// To track the action used for user refer the same object
	isNew := status == http.StatusNotFound
	if isNew {
		crmUser.Action = model.CRMActionCreated
	} else {
		crmUser.Action = model.CRMActionUpdated
	}

	db := C.GetServices().Db
	err = db.Create(&crmUser).Error
	if err != nil {
		if IsDuplicateRecordError(err) {
			return http.StatusConflict, nil
		}

		logCtx.WithError(err).
			Error("Failed to insert crm user document.")
		return http.StatusInternalServerError, err
	}

	return http.StatusCreated, nil
}

func (store *MemSQL) GetCRMUserByTypeAndAction(projectID int64, source U.CRMSource, id string, userType int, action model.CRMAction) (*model.CRMUser, int) {

	logFields := log.Fields{
		"project_id": projectID,
		"source":     source,
		"id":         id,
		"user_type":  userType,
		"action":     action,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)

	if projectID == 0 || source == 0 || id == "" || userType == 0 || action == 0 {
		logCtx.Error("Invalid parameters")
		return nil, http.StatusBadRequest
	}

	var crmUser model.CRMUser

	db := C.GetServices().Db
	err := db.Model(&model.CRMUser{}).Where("project_id = ? AND source = ? AND id =? AND type = ? and action = ? ",
		projectID, source, id, userType, action).
		Limit(1).Find(&crmUser).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}

		logCtx.WithError(err).Error("Failed to get crm user by type and action.")
		return nil, http.StatusInternalServerError
	}

	if crmUser.ID == "" {
		return nil, http.StatusNotFound
	}

	return &crmUser, http.StatusFound
}

func (store *MemSQL) UpdateCRMUserAsSynced(projectID int64, source U.CRMSource, crmUser *model.CRMUser, userID, syncID string) (*model.CRMUser, int) {

	logFields := log.Fields{
		"project_id": projectID,
		"source":     source,
		"id":         crmUser.ID,
		"user_type":  crmUser.Type,
		"action":     crmUser.Action,
		"timestamp":  crmUser.Timestamp,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)
	if projectID == 0 || source == 0 || crmUser.ID == "" || crmUser.Type == 0 || crmUser.Action == 0 || crmUser.Timestamp == 0 {
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
	err := db.Model(&model.CRMUser{}).Where("project_id = ? AND source = ? AND id = ? AND type= ? AND action = ? AND timestamp= ? ",
		projectID, source, crmUser.ID, crmUser.Type, crmUser.Action, crmUser.Timestamp).Updates(updates).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to update crm user as synced.")
		return nil, http.StatusInternalServerError
	}

	crmUser.SyncID = syncID
	crmUser.UserID = userID
	return crmUser, http.StatusAccepted
}

func (store *MemSQL) GetCRMUsersInOrderForSync(projectID int64, source U.CRMSource, startTimestamp, endTimestamp int64, recordProcessLimit int) ([]model.CRMUser, int) {

	logFields := log.Fields{
		"project_id":     projectID,
		"source":         source,
		"start_timestam": startTimestamp,
		"end_timestamp":  endTimestamp,
	}

	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)
	if projectID == 0 || source == 0 || startTimestamp == 0 || endTimestamp == 0 {
		logCtx.Error("Invalid parameters")
		return nil, http.StatusBadRequest
	}

	var crmUsers []model.CRMUser
	db := C.GetServices().Db
	dbx := db.Model(&model.CRMUser{}).Where("project_id = ? AND source = ? AND synced = false AND timestamp BETWEEN ? AND ?",
		projectID, source, startTimestamp, endTimestamp).Order("timestamp, created_at")
	if recordProcessLimit > 0 {
		dbx = dbx.Limit(recordProcessLimit)
	}

	if err := dbx.Find(&crmUsers).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}

		logCtx.WithError(err).Error("Failed to get crm user records for sync.")
		return nil, http.StatusInternalServerError
	}

	if len(crmUsers) == 0 {
		return nil, http.StatusNotFound
	}

	return crmUsers, http.StatusFound
}

func (store *MemSQL) GetCRMUsersMinimumTimestampForSync(projectID int64, source U.CRMSource) (int64, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"source":     source,
	}

	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)
	if projectID == 0 || source == 0 {
		logCtx.Error("Invalid parameters")
		return 0, http.StatusBadRequest
	}

	var minTimestamp struct {
		Timestamp int64
	}

	db := C.GetServices().Db
	err := db.Model(&model.CRMUser{}).Where("project_id = ? AND source = ? AND synced = false ",
		projectID, source).Select("min(timestamp) as timestamp").Scan(&minTimestamp).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return 0, http.StatusNotFound
		}

		logCtx.WithError(err).Error("Failed to get crm user min timestamp for sync.")
		return 0, http.StatusInternalServerError
	}

	if minTimestamp.Timestamp == 0 {
		return 0, http.StatusNotFound
	}

	return minTimestamp.Timestamp, http.StatusFound
}

func (store *MemSQL) GetCRMUsersTypeAndAction(projectID int64, source U.CRMSource) ([]model.CRMUser, int) {
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

	crmUsersTypeAndActions := make([]model.CRMUser, 0, 0)

	db := C.GetServices().Db
	err := db.Table("crm_users").Where("project_id = ? AND source = ?", projectID, source).
		Select("DISTINCT(type) as type, action").Find(&crmUsersTypeAndActions).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}

		logCtx.WithError(err).Error("Failed to get all CRM users type and actions.")
		return nil, http.StatusInternalServerError
	}

	if len(crmUsersTypeAndActions) == 0 {
		return nil, http.StatusNotFound
	}

	return crmUsersTypeAndActions, http.StatusFound
}
