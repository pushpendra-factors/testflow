package memsql

import (
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) CreateShareableURL(shareableURLParams *model.ShareableURL) (*model.ShareableURL, int) {
	logFields := log.Fields{
		"shareable_url": shareableURLParams,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	if shareableURLParams == nil || shareableURLParams.QueryID == "" {
		logCtx.Error("Invalid shareable url params")
		return nil, http.StatusBadRequest
	}

	if shareableURLParams.ID == "" {
		shareableURLParams.ID = U.GetUUID()
	}

	db := C.GetServices().Db
	if err := db.Create(&shareableURLParams).Error; err != nil {
		logCtx.WithError(err).Error("CreateShareableURL Failed")
		return nil, http.StatusInternalServerError
	}

	return shareableURLParams, http.StatusCreated
}

func (store *MemSQL) GetAllShareableURLsWithProjectIDAndAgentID(projectID int64, agentUUID string) ([]*model.ShareableURL, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"agent_uuid": agentUUID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	if projectID == 0 || agentUUID == "" {
		logCtx.Error("Invalid project id/agent uuid")
		return nil, http.StatusBadRequest
	}

	db := C.GetServices().Db
	shareableURLs := make([]*model.ShareableURL, 0)
	if err := db.Order("created_at DESC").Where("project_id = ? AND created_by = ? AND is_deleted = ?", projectID, agentUUID, false).Find(&shareableURLs).Error; err != nil {
		logCtx.WithError(err).Error("Failed to fetch rows from shareable_urls table for project")
		return shareableURLs, http.StatusInternalServerError
	}
	return shareableURLs, http.StatusFound
}

func (store *MemSQL) GetShareableURLWithShareStringAndAgentID(projectID int64, shareString, agentUUID string) (*model.ShareableURL, int) {
	logFields := log.Fields{
		"project_id":     projectID,
		"share_query_id": shareString,
		"agent_uuid":     agentUUID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	var shareableURL model.ShareableURL

	if shareString == "" || agentUUID == "" || projectID == 0 {
		logCtx.Error("Invalid share string/agent uuid/project id")
		return &shareableURL, http.StatusBadRequest
	}

	db := C.GetServices().Db
	if err := db.Where("query_id = ? AND project_id = ? AND is_deleted = ? AND expires_at > ? AND created_by = ?", shareString, projectID, false, time.Now().Unix(), agentUUID).First(&shareableURL).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return &shareableURL, http.StatusNotFound
		}
		logCtx.WithError(err).Error("GetShareableURLWithShareString Failed")
		return &shareableURL, http.StatusInternalServerError
	}
	return &shareableURL, http.StatusFound
}

func (store *MemSQL) GetShareableURLWithShareStringWithLargestScope(projectID int64, shareString string, entityType int) (*model.ShareableURL, int) {
	logFields := log.Fields{
		"project_id":     projectID,
		"share_query_id": shareString,
		"entity_type":    entityType,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	var shareableURL model.ShareableURL

	if shareString == "" || projectID == 0 || entityType == 0 {
		logCtx.Error("Invalid share string/project id")
		return &shareableURL, http.StatusBadRequest
	}

	db := C.GetServices().Db
	shareableURLs := make([]*model.ShareableURL, 0)
	rows := db.Where("query_id = ? AND project_id = ? AND entity_type = ? AND is_deleted = ? AND expires_at > ?", shareString, projectID, entityType, false, time.Now().Unix()).Find(&shareableURLs)
	if rows.Error != nil {
		logCtx.WithError(rows.Error).Error("Error fetching shareable url")
		return &shareableURL, http.StatusInternalServerError
	}
	if rows.RowsAffected == 0 {
		logCtx.Error("No shareable url found")
		return &shareableURL, http.StatusNotFound
	}

	// Get the largest scope share type. Add allowed users logic here.
	minShareType := shareableURLs[0].ShareType
	for _, share := range shareableURLs {
		if share.ShareType < minShareType {
			minShareType = share.ShareType
		}
	}

	if err := db.Where("query_id = ? AND project_id = ? AND is_deleted = ? AND expires_at > ? AND share_type = ?", shareString, projectID, false, time.Now().Unix(), minShareType).First(&shareableURL).Error; err != nil {
		logCtx.WithError(err).Error("GetShareableURLWithShareString Failed")
		return &shareableURL, http.StatusInternalServerError
	}
	return &shareableURL, http.StatusFound
}

// func (store *MemSQL) GetShareableURLWithID(projectID uint64, shareId string) (*model.ShareableURL, int) {
// 	logFields := log.Fields{
// 		"project_id": projectID,
// 		"share_id":   shareId,
// 	}
// 	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
// 	logCtx := log.WithFields(logFields)

// 	var shareableURL model.ShareableURL

// 	if shareId == "" || projectID == 0 {
// 		logCtx.Error("Invalid share id/project id")
// 		return &shareableURL, http.StatusBadRequest
// 	}

// 	db := C.GetServices().Db
// 	if err := db.Where("id = ? AND project_id = ? AND is_deleted = ? AND expires_at > ?", shareId, projectID, false, time.Now().Unix()).First(&shareableURL).Error; err != nil {
// 		if gorm.IsRecordNotFoundError(err) {
// 			return &shareableURL, http.StatusNotFound
// 		}
// 		logCtx.WithError(err).Error("GetShareableURLWithID Failed")
// 		return &shareableURL, http.StatusInternalServerError
// 	}
// 	return &shareableURL, http.StatusFound
// }

func (store *MemSQL) updateShareableURL(whereFields, updateFields map[string]interface{}) int {
	logFields := log.Fields{
		"where_fields":  whereFields,
		"update_fields": updateFields,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	db := C.GetServices().Db
	rows := db.Table("shareable_urls").Where(whereFields).Where("is_deleted = ? AND expires_at > ?", false, time.Now().Unix()).Updates(updateFields)
	if rows.Error != nil {
		logCtx.WithError(rows.Error).Error("updateShareableURL Failed")
		return http.StatusInternalServerError
	}
	if rows.RowsAffected == 0 {
		logCtx.Error("No rows affected")
		return http.StatusNotFound
	}
	return http.StatusAccepted
}

// func (store *MemSQL) UpdateShareableURLShareTypeWithShareIDandCreatedBy(projectID uint64, shareId, createdBy string, shareType int, allowedUsers string) int {
// 	logFields := log.Fields{
// 		"project_id":    projectID,
// 		"share_id":      shareId,
// 		"share_type":    shareType,
// 		"allowed_users": allowedUsers,
// 	}
// 	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
// 	logCtx := log.WithFields(logFields)

// 	if shareId == "" || createdBy == "" || !model.ValidShareEntityTypes[shareType] || projectID == 0 {
// 		logCtx.Error("Invalid share id/created by/share type/project id")
// 		return http.StatusBadRequest
// 	}

// 	whereFields := map[string]interface{}{"id": shareId, "project_id": projectID, "created_by": createdBy}

// 	updateFields := make(map[string]interface{}, 0)
// 	updateFields["share_type"] = shareType
// 	// if shareType == model.ShareableURLShareTypeAllowedUsers {
// 	// 	updateFields["allowed_users"] = allowedUsers
// 	// }
// 	return store.updateShareableURL(whereFields, updateFields)
// }

func (store *MemSQL) DeleteShareableURLWithShareIDandAgentID(projectID int64, shareId, createdBy string) int {
	logFields := log.Fields{
		"project_id": projectID,
		"share_id":   shareId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	if shareId == "" || createdBy == "" || projectID == 0 {
		logCtx.Error("Invalid share id/created by/project id")
		return http.StatusBadRequest
	}

	whereFields := map[string]interface{}{"id": shareId, "project_id": projectID, "created_by": createdBy}
	updateFields := map[string]interface{}{"is_deleted": true}

	return store.updateShareableURL(whereFields, updateFields)
}

func (store *MemSQL) DeleteShareableURLWithEntityIDandType(projectID int64, entityID int64, entityType int) int {
	logFields := log.Fields{
		"project_id":  projectID,
		"entity_id":   entityID,
		"entity_type": entityType,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	if !model.ValidShareEntityTypes[entityType] || entityID == 0 || projectID == 0 {
		logCtx.Error("Invalid entity type/entity id/project id")
		return http.StatusBadRequest
	}

	whereFields := map[string]interface{}{"entity_id": entityID, "project_id": projectID, "entity_type": entityType}
	updateFields := map[string]interface{}{"is_deleted": true}

	return store.updateShareableURL(whereFields, updateFields)
}

func (store *MemSQL) RevokeShareableURLsWithShareString(projectID int64, shareString string) (int, string) {
	logFields := log.Fields{
		"project_id":   projectID,
		"share_string": shareString,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	if shareString == "" || projectID == 0 {
		logCtx.Error("Invalid share string/project id")
		return http.StatusBadRequest, "Share string is empty"
	}

	whereFields := map[string]interface{}{"query_id": shareString, "project_id": projectID}
	updateFields := map[string]interface{}{"is_deleted": true}

	errCode := store.updateShareableURL(whereFields, updateFields)
	if errCode != http.StatusAccepted && errCode != http.StatusNotFound {
		logCtx.WithField("err_code", errCode).Error("RevokeShareableURLsWithShareString Failed")
		return errCode, "Failed to revoke shareable urls"
	}

	var shareableURLs []string
	shareableURLs = append(shareableURLs, shareString)
	errCode = store.UpdateQueryIDsWithNewIDs(projectID, shareableURLs)
	if errCode == http.StatusPartialContent {
		logCtx.WithField("err_code", errCode).Error("Failed to reset share string of the query")
		return http.StatusPartialContent, "Failed to reset share string of the query"
	}
	return http.StatusAccepted, ""
}

func (store *MemSQL) RevokeShareableURLsWithProjectID(projectId int64) (int, string) {
	logFields := log.Fields{
		"project_id": projectId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	if projectId == 0 {
		logCtx.Error("Invalid project id")
		return http.StatusBadRequest, "Invalid project id"
	}

	whereFields := map[string]interface{}{"project_id": projectId}
	updateFields := map[string]interface{}{"is_deleted": true}

	shareableURLs := []string{}
	db := C.GetServices().Db
	if err := db.Table("shareable_urls").Select("distinct(query_id)").Where(whereFields).Where("is_deleted = ? AND expires_at > ?", false, time.Now().Unix()).Pluck("query_id", &shareableURLs).Error; err != nil {
		logCtx.WithError(err).Error("Failed to get shareable urls")
		return http.StatusInternalServerError, "Failed to get shareable urls"
	}

	errCode := store.updateShareableURL(whereFields, updateFields)
	if errCode != http.StatusAccepted && errCode != http.StatusNotFound {
		logCtx.WithField("err_code", errCode).Error("Fail to revoke shareable urls")
		return errCode, "Failed to revoke shareable urls"
	}

	errCode = store.UpdateQueryIDsWithNewIDs(projectId, shareableURLs)
	if errCode == http.StatusPartialContent {
		logCtx.WithField("err_code", errCode).Error("Some queries were not reset")
		return http.StatusPartialContent, "Some share strings were not reset"
	}

	return http.StatusAccepted, ""
}
