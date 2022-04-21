package postgres

import (
	C "factors/config"
	"factors/model/model"
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

func (pg *Postgres) CreateShareableURL(shareableURLParams *model.ShareableURL) (*model.ShareableURL, int) {
	if shareableURLParams == nil || shareableURLParams.QueryID == "" {
		return nil, http.StatusBadRequest
	}

	db := C.GetServices().Db
	if err := db.Create(&shareableURLParams).Error; err != nil {
		log.WithError(err).Error("CreateShareableURL Failed")
		return nil, http.StatusInternalServerError
	}

	return shareableURLParams, http.StatusCreated
}

func (pg *Postgres) GetAllShareableURLsWithProjectIDAndAgentID(projectID uint64, agentUUID string) ([]*model.ShareableURL, int) {
	if projectID == 0 || agentUUID == "" {
		return nil, http.StatusBadRequest
	}

	db := C.GetServices().Db
	shareableURLs := make([]*model.ShareableURL, 0)
	if err := db.Order("created_at DESC").Where("project_id = ? AND created_by = ? AND is_deleted = ?", projectID, agentUUID, false).Find(&shareableURLs).Error; err != nil {
		log.WithField("project_id", projectID).Error("Failed to fetch rows from shareable_urls table for project")
		return shareableURLs, http.StatusInternalServerError
	}
	return shareableURLs, http.StatusFound
}

func (pg *Postgres) GetShareableURLWithShareStringAndAgentID(projectID uint64, shareString, agentUUID string) (*model.ShareableURL, int) {
	var shareableURL model.ShareableURL

	if projectID == 0 || shareString == "" || agentUUID == "" {
		return &shareableURL, http.StatusBadRequest
	}

	db := C.GetServices().Db
	if err := db.Where("query_id = ? AND project_id = ? AND is_deleted = ? AND expires_at > ? AND created_by = ?", shareString, projectID, false, time.Now().Unix(), agentUUID).First(&shareableURL).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return &shareableURL, http.StatusNotFound
		}
		log.WithError(err).Error("GetShareableURLWithShareString Failed")
		return &shareableURL, http.StatusInternalServerError
	}
	return &shareableURL, http.StatusFound
}

func (pg *Postgres) GetShareableURLWithShareStringWithLargestScope(projectID uint64, shareString string, entityType int) (*model.ShareableURL, int) {
	var shareableURL model.ShareableURL

	if shareString == "" || projectID == 0 || !model.ValidShareEntityTypes[entityType] {
		return &shareableURL, http.StatusBadRequest
	}

	db := C.GetServices().Db

	shareableURLs := make([]*model.ShareableURL, 0)
	rows := db.Where("query_id = ? AND project_id = ? AND entity_type = ? AND is_deleted = ? AND expires_at > ?", shareString, projectID, entityType, false, time.Now().Unix()).Find(&shareableURLs)
	if rows.Error != nil {
		return &shareableURL, http.StatusInternalServerError
	}
	if rows.RowsAffected == 0 {
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
		return &shareableURL, http.StatusInternalServerError
	}
	return &shareableURL, http.StatusFound
}

// func (pg *Postgres) GetShareableURLWithID(projectID uint64, shareId string) (*model.ShareableURL, int) {
// 	var shareableURL model.ShareableURL

// 	if shareId == "" || projectID == 0 {
// 		return &shareableURL, http.StatusBadRequest
// 	}

// 	db := C.GetServices().Db
// 	if err := db.Where("id = ? AND project_id = ? AND is_deleted = ? AND expires_at > ?", shareId, projectID, false, time.Now().Unix()).First(&shareableURL).Error; err != nil {
// 		if gorm.IsRecordNotFoundError(err) {
// 			return &shareableURL, http.StatusNotFound
// 		}
// 		return &shareableURL, http.StatusInternalServerError
// 	}
// 	return &shareableURL, http.StatusFound
// }

func (pg *Postgres) updateShareableURL(whereFields, updateFields map[string]interface{}) int {
	db := C.GetServices().Db
	rows := db.Table("shareable_urls").Where(whereFields).Where("is_deleted = ? AND expires_at > ?", false, time.Now().Unix()).Updates(updateFields)
	if rows.Error != nil {
		log.WithError(rows.Error).Error("updateShareableURL Failed")
		return http.StatusInternalServerError
	}
	if rows.RowsAffected == 0 {
		log.Error("No rows affected")
		return http.StatusNotFound
	}
	return http.StatusAccepted
}

// func (pg *Postgres) UpdateShareableURLShareTypeWithShareIDandCreatedBy(projectID uint64, shareId, createdBy string, shareType int, allowedUsers string) int {
// 	if shareId == "" || createdBy == "" || !model.ValidShareTypes[shareType] || projectID == 0 {	
// 		return http.StatusBadRequest
// 	}

// 	whereFields := map[string]interface{}{"id": shareId, "project_id": projectID, "created_by": createdBy}

// 	updateFields := make(map[string]interface{}, 0)
// 	updateFields["share_type"] = shareType
// 	// if shareType == model.ShareableURLShareTypeAllowedUsers {
// 	// 	updateFields["allowed_users"] = allowedUsers
// 	// }
// 	return pg.updateShareableURL(whereFields, updateFields)
// }

func (pg *Postgres) DeleteShareableURLWithShareIDandAgentID(projectID uint64, shareId, createdBy string) int {
	if shareId == "" || createdBy == "" || projectID == 0 {
		return http.StatusBadRequest
	}

	whereFields := map[string]interface{}{"id": shareId, "project_id": projectID, "created_by": createdBy}
	updateFields := map[string]interface{}{"is_deleted": true}
	
	return pg.updateShareableURL(whereFields, updateFields)
}

func (pg *Postgres) DeleteShareableURLWithEntityIDandType(projectID, entityID uint64, entityType int) int {
	if !model.ValidShareEntityTypes[entityType] || entityID == 0 || projectID == 0 {
		return http.StatusBadRequest
	}

	whereFields := map[string]interface{}{"entity_id": entityID, "project_id": projectID, "entity_type": entityType}
	updateFields := map[string]interface{}{"is_deleted": true}

	return pg.updateShareableURL(whereFields, updateFields)
}

func (pg *Postgres) RevokeShareableURLsWithShareString(projectID uint64, shareString string) (int, string) {
	if shareString == "" || projectID == 0 {
		return http.StatusBadRequest, "Share string is empty"
	}

	whereFields := map[string]interface{}{"query_id": shareString, "project_id": projectID}
	updateFields := map[string]interface{}{"is_deleted": true}
	
	errCode := pg.updateShareableURL(whereFields, updateFields)
	if errCode != http.StatusAccepted && errCode != http.StatusNotFound {
		return errCode, "Failed to revoke shareable urls"
	}

	var shareableURLs []string
	shareableURLs = append(shareableURLs, shareString)
	errCode = pg.UpdateQueryIDsWithNewIDs(projectID, shareableURLs)
	if errCode == http.StatusPartialContent {
		return http.StatusPartialContent, "Failed to reset share string of the query"
	}
	return http.StatusAccepted, ""
}

func (pg *Postgres) RevokeShareableURLsWithProjectID(projectId uint64) (int, string) {
	if projectId == 0 {
		return http.StatusBadRequest, "Invalid project id"
	}

	whereFields := map[string]interface{}{"project_id": projectId}
	updateFields := map[string]interface{}{"is_deleted": true}

	// Collect all shareable urls with project id
	shareableURLs := []string{}
	db := C.GetServices().Db
	if err := db.Table("shareable_urls").Select("distinct(query_id)").Where("project_id = ? AND is_deleted = ? AND expires_at > ?", projectId, false, time.Now().Unix()).Pluck("query_id", &shareableURLs).Error; err != nil {
		return http.StatusInternalServerError, "Failed to get shareable urls"
	}

	// return pg.updateShareableURL(whereFields, updateFields)
	errCode := pg.updateShareableURL(whereFields, updateFields)
	if errCode != http.StatusAccepted && errCode != http.StatusNotFound {
		return errCode, "Failed to revoke shareable urls"
	}

	errCode = pg.UpdateQueryIDsWithNewIDs(projectId, shareableURLs)
	if errCode == http.StatusPartialContent {
		return http.StatusPartialContent, "Some share strings were not reset"
	}

	return http.StatusAccepted, ""
}