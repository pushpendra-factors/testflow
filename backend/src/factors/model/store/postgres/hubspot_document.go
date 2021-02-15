package postgres

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"

	C "factors/config"
	"factors/model/model"
	U "factors/util"
)

var errorInvalidHubspotDocumentType = errors.New("invalid document type")
var errorFailedToGetCreatedAtFromHubspotDocument = errors.New("failed to get created_at from document")
var errorFailedToGetUpdatedAtFromHubspotDocument = errors.New("failed to get updated_at from document")
var errorUsingFallbackKey = errors.New("using fallback key from document")

const error_DuplicateHubspotDocument = "pq: duplicate key value violates unique constraint \"hubspot_documents_pkey\""

func isDuplicateHubspotDocumentError(err error) bool {
	return err.Error() == error_DuplicateHubspotDocument
}

func getHubspotDocumentId(document *model.HubspotDocument) (string, error) {
	if document.Type == 0 {
		return "", errorInvalidHubspotDocumentType
	}

	documentMap, err := U.DecodePostgresJsonb(document.Value)
	if err != nil {
		return "", err
	}

	var idKey string
	switch document.Type {
	case model.HubspotDocumentTypeCompany:
		idKey = "companyId"
	case model.HubspotDocumentTypeContact:
		idKey = "vid"
	case model.HubspotDocumentTypeDeal:
		idKey = "dealId"
	case model.HubspotDocumentTypeFormSubmission:
		idKey = "formId"
	default:
		idKey = "guid"
	}

	if idKey == "" {
		return "", errors.New("invalid hubspot document key")
	}

	id, idExists := (*documentMap)[idKey]
	if !idExists {
		return "", errors.New("id key not exist on hubspot document")
	}

	idAsString := U.GetPropertyValueAsString(id)
	if idAsString == "" {
		return "", errors.New("invalid id on hubspot document")
	}

	// No id on form submission doc so Id for form_submission
	// doc is <form_id>:<submitted_at>.
	if document.Type == model.HubspotDocumentTypeFormSubmission {
		submittedAt, submittedAtExists := (*documentMap)["submittedAt"]
		if !submittedAtExists {
			return "", errors.New("submitted not found on form_submission document type")
		}

		submittedAtAsString := U.GetPropertyValueAsString(submittedAt)
		idAsString = fmt.Sprintf("%s:%s", idAsString, submittedAtAsString)
	}

	return idAsString, nil
}

func getHubspotDocumentByIdAndType(projectId uint64, id string, docType int) ([]model.HubspotDocument, int) {
	logCtx := log.WithFields(log.Fields{"project_id": projectId, "id": id, "type": docType})

	var documents []model.HubspotDocument
	if projectId == 0 || id == "" || docType == 0 {
		logCtx.Error("Failed to get hubspot document by id and type. Invalid project_id or id or type.")
		return documents, http.StatusBadRequest
	}

	db := C.GetServices().Db
	err := db.Where("project_id = ? AND id = ? AND type = ?", projectId, id,
		docType).Find(&documents).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to get hubspot documents.")
		return documents, http.StatusInternalServerError
	}

	if len(documents) == 0 {
		return documents, http.StatusNotFound
	}

	return documents, http.StatusFound
}

func (pg *Postgres) GetHubspotDocumentByTypeAndActions(projectId uint64, ids []string,
	docType int, actions []int) ([]model.HubspotDocument, int) {

	logCtx := log.WithFields(log.Fields{"project_id": projectId, "ids": ids,
		"type": docType, "actions": actions})

	var documents []model.HubspotDocument
	if projectId == 0 || len(ids) == 0 || docType == 0 || len(actions) == 0 {
		logCtx.Error("Failed to get hubspot document by id and type. Invalid project_id or id or type or action.")
		return nil, http.StatusBadRequest
	}

	db := C.GetServices().Db
	err := db.Order("timestamp").Where(
		"project_id = ? AND id IN (?) AND type = ? AND action IN (?)",
		projectId, ids, docType, actions).Find(&documents).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to get hubspot documents.")
		return nil, http.StatusInternalServerError
	}

	if len(documents) == 0 {
		return nil, http.StatusNotFound
	}

	return documents, http.StatusFound
}

func getTimestampFromPropertiesByKey(propertiesMap map[string]interface{}, key string) (int64, error) {
	propertyValue, exists := propertiesMap[key]
	if !exists || propertyValue == nil {
		return 0, errors.New("failed to get timestamp from property key")
	}

	propertyValueMap := propertyValue.(map[string]interface{})
	timestampValue, exists := propertyValueMap["timestamp"]
	if !exists || timestampValue == nil {
		return 0, errors.New("timestamp key not exist on property map")
	}

	timestamp, err := model.ReadHubspotTimestamp(timestampValue)
	if err != nil || timestamp == 0 {
		return 0, errors.New("failed to read hubspot timestamp value")
	}

	return timestamp, nil
}

func getHubspotDocumentCreatedTimestamp(document *model.HubspotDocument) (int64, error) {
	if document.Type == 0 {
		return 0, errorInvalidHubspotDocumentType
	}

	value, err := U.DecodePostgresJsonb(document.Value)
	if err != nil {
		return 0, err
	}

	if document.Type == model.HubspotDocumentTypeCompany {
		properties, exists := (*value)["properties"]
		if !exists || properties == nil {
			return 0, errorFailedToGetCreatedAtFromHubspotDocument
		}
		propertiesMap := properties.(map[string]interface{})

		createdAt, err := getTimestampFromPropertiesByKey(propertiesMap, "name")
		if err != nil {
			createdAt, err = getTimestampFromPropertiesByKey(propertiesMap, "createdate")
			if err != nil {
				return 0, err
			}
		}

		return createdAt, nil
	}

	if document.Type == model.HubspotDocumentTypeDeal {
		properties, exists := (*value)["properties"]
		if !exists || properties == nil {
			return 0, errorFailedToGetCreatedAtFromHubspotDocument
		}
		propertiesMap := properties.(map[string]interface{})

		hsCreateDate, exists := propertiesMap["hs_createdate"]
		if exists {
			hsCreateDateMap := hsCreateDate.(map[string]interface{})
			createdAtValue, exists := hsCreateDateMap["value"]
			if exists && createdAtValue != nil {
				createdAt, err := model.ReadHubspotTimestamp(createdAtValue)
				if err != nil || createdAt == 0 {
					return 0, errorFailedToGetCreatedAtFromHubspotDocument
				}

				return createdAt, nil
			}
		}

		createdAt, err := getTimestampFromPropertiesByKey(propertiesMap, "dealname")
		if err != nil {
			return 0, err
		}

		return createdAt, nil
	}

	var createdAtKey string
	if document.Type == model.HubspotDocumentTypeContact {
		createdAtKey = "createdate"
	} else if document.Type == model.HubspotDocumentTypeForm {
		createdAtKey = "createdAt"
	} else if document.Type == model.HubspotDocumentTypeFormSubmission {
		createdAtKey = "submittedAt"
	} else {
		return 0, errorFailedToGetCreatedAtFromHubspotDocument
	}

	if document.Type == model.HubspotDocumentTypeContact {

		var hubspotDocumentProperties HubspotDocumentProperties
		err := json.Unmarshal(document.Value.RawMessage, &hubspotDocumentProperties)
		if err != nil {
			return 0, errors.New("Failed to unmarshal document properties")
		}

		createdAtStr, exist := hubspotDocumentProperties.Properties[createdAtKey]
		if exist {
			createdAt, err := model.ReadHubspotTimestamp(createdAtStr.Value)
			if err != nil {
				return 0, err
			}
			return createdAt, nil
		}

		createdAtKey = "addedAt"
		createdAtInt, exists := (*value)[createdAtKey]
		if !exists {
			return 0, errorFailedToGetCreatedAtFromHubspotDocument
		}

		createdAt, err := model.ReadHubspotTimestamp(createdAtInt)
		if err != nil || createdAt == 0 {
			log.WithFields(log.Fields{"document_id": document.ID}).Warn("Failed to read addedAt from contact document.")
			return 0, errorFailedToGetCreatedAtFromHubspotDocument
		}

		return createdAt, errorUsingFallbackKey
	}

	createdAtInt, exists := (*value)[createdAtKey]
	if !exists || createdAtInt == nil {
		return 0, errorFailedToGetCreatedAtFromHubspotDocument
	}

	createdAt, err := model.ReadHubspotTimestamp(createdAtInt)
	if err != nil || createdAt == 0 {
		return 0, errorFailedToGetCreatedAtFromHubspotDocument
	}

	return createdAt, nil
}

func getHubspotDocumentUpdatedTimestamp(document *model.HubspotDocument) (int64, error) {
	if document.Type == 0 {
		return 0, errorInvalidHubspotDocumentType
	}

	value, err := U.DecodePostgresJsonb(document.Value)
	if err != nil {
		return 0, err
	}

	// property nested value.
	var propertyUpdateAtKey string
	if document.Type == model.HubspotDocumentTypeCompany ||
		document.Type == model.HubspotDocumentTypeDeal {

		propertyUpdateAtKey = "hs_lastmodifieddate"
	} else if document.Type == model.HubspotDocumentTypeContact {
		propertyUpdateAtKey = "lastmodifieddate"
	}
	if propertyUpdateAtKey != "" {
		properties, exists := (*value)["properties"]
		if !exists || properties == nil {
			return 0, errorFailedToGetUpdatedAtFromHubspotDocument
		}
		propertiesMap := properties.(map[string]interface{})

		propertyUpdateAt, exists := propertiesMap[propertyUpdateAtKey]
		if !exists || propertyUpdateAt == nil {
			return 0, errorFailedToGetUpdatedAtFromHubspotDocument
		}

		propertyUpdateAtMap := propertyUpdateAt.(map[string]interface{})
		value, exists := propertyUpdateAtMap["value"]
		if !exists || value == nil {
			return 0, errorFailedToGetUpdatedAtFromHubspotDocument
		}

		updatedAt, err := model.ReadHubspotTimestamp(value)
		if err != nil || updatedAt == 0 {
			return 0, errorFailedToGetUpdatedAtFromHubspotDocument
		}

		return updatedAt, nil
	}

	// direct values.
	var updatedAtKey string
	if document.Type == model.HubspotDocumentTypeForm {
		updatedAtKey = "updatedAt"
	} else if document.Type == model.HubspotDocumentTypeFormSubmission {
		updatedAtKey = "submittedAt"
	}
	if updatedAtKey != "" {
		updatedAtInt, exists := (*value)[updatedAtKey]
		if !exists || updatedAtInt == nil {
			return 0, errorFailedToGetUpdatedAtFromHubspotDocument
		}

		updatedAt, err := model.ReadHubspotTimestamp(updatedAtInt)
		if err != nil || updatedAt == 0 {
			return 0, errorFailedToGetUpdatedAtFromHubspotDocument
		}

		return updatedAt, nil
	}

	return 0, errorFailedToGetUpdatedAtFromHubspotDocument
}

func (pg *Postgres) CreateHubspotDocument(projectId uint64, document *model.HubspotDocument) int {
	logCtx := log.WithField("project_id", document.ProjectId)

	if projectId == 0 {
		logCtx.Error("Invalid project_id on create hubspot document.")
		return http.StatusBadRequest
	}
	document.ProjectId = projectId

	documentType, err := model.GetHubspotTypeByAlias(document.TypeAlias)
	if err != nil {
		logCtx.WithError(err).Error("Invalid type on create hubspot document.")
		return http.StatusBadRequest
	}
	document.Type = documentType

	if U.IsEmptyPostgresJsonb(document.Value) {
		logCtx.Error("Empty document value on create hubspot document.")
		return http.StatusBadRequest
	}

	documentId, err := getHubspotDocumentId(document)
	if err != nil {
		logCtx.WithError(err).Error(
			"Failed to get id for hubspot document on create.")
		return http.StatusInternalServerError
	}
	document.ID = documentId

	logCtx = logCtx.WithField("type", document.Type).WithField("value", document.Value)

	_, errCode := getHubspotDocumentByIdAndType(document.ProjectId,
		document.ID, document.Type)
	if errCode == http.StatusInternalServerError || errCode == http.StatusBadRequest {
		return errCode
	}
	isNew := errCode == http.StatusNotFound

	var timestamp int64
	if isNew {
		document.Action = model.HubspotDocumentActionCreated // created
		timestamp, err = getHubspotDocumentCreatedTimestamp(document)
	} else {
		document.Action = model.HubspotDocumentActionUpdated // updated
		// Any update on the entity would create a new hubspot document.
		// i.e, deal will be synced after updating a created deal with a
		// contact or a company.
		timestamp, err = getHubspotDocumentUpdatedTimestamp(document)
	}
	if err != nil {
		if err != errorUsingFallbackKey {
			logCtx.WithField("action", document.Action).WithError(err).Error(
				"Failed to get timestamp from hubspot document on create.")
			return http.StatusInternalServerError
		}

		logCtx.WithField("action", document.Action).WithError(err).Error("Missing document key.")
	}
	document.Timestamp = timestamp

	db := C.GetServices().Db
	err = db.Create(document).Error
	if err != nil {
		if isDuplicateHubspotDocumentError(err) {
			return http.StatusConflict
		}

		logCtx.WithError(err).Error("Failed to create hubspot document.")
		return http.StatusInternalServerError
	}

	return http.StatusCreated
}

func getHubspotTypeAlias(t int) string {
	for alias, typ := range model.HubspotDocumentTypeAlias {
		if t == typ {
			return alias
		}
	}

	return ""
}

func (pg *Postgres) GetHubspotSyncInfo() (*model.HubspotSyncInfo, int) {
	var lastSyncInfo []model.HubspotLastSyncInfo

	db := C.GetServices().Db
	err := db.Table("hubspot_documents").Select(
		"project_id, type, MAX(timestamp) as timestamp").Group(
		"project_id, type").Find(&lastSyncInfo).Error
	if err != nil {
		return nil, http.StatusInternalServerError
	}

	lastSyncInfoByProject := make(map[uint64]map[string]int64, 0)
	for _, syncInfo := range lastSyncInfo {
		if _, projectExists := lastSyncInfoByProject[syncInfo.ProjectId]; !projectExists {
			lastSyncInfoByProject[syncInfo.ProjectId] = make(map[string]int64)
		}

		lastSyncInfoByProject[syncInfo.ProjectId][getHubspotTypeAlias(syncInfo.Type)] = syncInfo.Timestamp
	}

	// project sync of hubspot enable projects.
	enabledProjectLastSync := make(map[uint64]map[string]int64, 0)

	// get project settings of hubspot enaled projects.
	projectSettings, errCode := pg.GetAllHubspotProjectSettings()
	if errCode != http.StatusFound {
		return nil, http.StatusInternalServerError
	}

	settingsByProject := make(map[uint64]*model.HubspotProjectSettings, 0)
	for i, ps := range projectSettings {
		_, pExists := lastSyncInfoByProject[ps.ProjectId]

		if !pExists {
			// add projects not synced before.
			enabledProjectLastSync[ps.ProjectId] = make(map[string]int64, 0)
		} else {
			// add sync info if avaliable.
			enabledProjectLastSync[ps.ProjectId] = lastSyncInfoByProject[ps.ProjectId]
		}

		// add types not synced before.
		for typ := range model.HubspotDocumentTypeAlias {
			_, typExists := enabledProjectLastSync[ps.ProjectId][typ]
			if !typExists {
				// last sync timestamp as zero as type not synced before.
				enabledProjectLastSync[ps.ProjectId][typ] = 0
			}
		}

		settingsByProject[projectSettings[i].ProjectId] = &projectSettings[i]
	}

	var syncInfo model.HubspotSyncInfo
	syncInfo.LastSyncInfo = enabledProjectLastSync
	syncInfo.ProjectSettings = settingsByProject

	return &syncInfo, http.StatusFound
}

func (pg *Postgres) GetHubspotFormDocuments(projectId uint64) ([]model.HubspotDocument, int) {
	var documents []model.HubspotDocument

	db := C.GetServices().Db
	err := db.Where("project_id=? AND type=?",
		projectId, 4).Find(&documents).Error
	if err != nil {
		log.WithField("projectId", projectId).WithError(err).Error(
			"Finding documents failed on GetHubspotFormDocuments")
		return nil, http.StatusInternalServerError
	}

	return documents, http.StatusFound
}

func (pg *Postgres) GetHubspotDocumentsByTypeForSync(projectId uint64, typ int) ([]model.HubspotDocument, int) {
	logCtx := log.WithFields(log.Fields{"project_id": projectId, "type": typ})

	if projectId == 0 || typ == 0 {
		logCtx.Error("Invalid project_id or type on get hubspot documents by type.")
		return nil, http.StatusBadRequest
	}

	var documents []model.HubspotDocument

	db := C.GetServices().Db
	err := db.Order("timestamp, created_at ASC").Where("project_id=? AND type=? AND synced=false",
		projectId, typ).Find(&documents).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to get hubspot documents by type.")
		return nil, http.StatusInternalServerError
	}

	return documents, http.StatusFound
}

func (pg *Postgres) GetSyncedHubspotDealDocumentByIdAndStage(projectId uint64, id string,
	stage string) (*model.HubspotDocument, int) {

	logCtx := log.WithFields(log.Fields{"project_id": projectId, "id": id, "stage": stage})

	if projectId == 0 || id == "" || stage == "" {
		logCtx.Error(
			"Invalid project_id or id or stage on get hubspot synced deal by id and stage.")
		return nil, http.StatusBadRequest
	}

	var documents []model.HubspotDocument

	db := C.GetServices().Db
	err := db.Limit(1).Where(
		"project_id=? AND id=? AND type=? AND synced=true AND value->'properties'->'dealstage'->>'value'=?",
		projectId, id, model.HubspotDocumentTypeDeal, stage).Find(&documents).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to get hubspot synced deal by id and stage.")
		return nil, http.StatusInternalServerError
	}

	if len(documents) == 0 {
		return nil, http.StatusNotFound
	}

	return &documents[0], http.StatusFound
}

// HubspotProperty only holds the value for hubspot document properties
type HubspotProperty struct {
	Value string `json:"value"`
}

// HubspotDocumentProperties only holds the properties object of the doucment
type HubspotDocumentProperties struct {
	Properties map[string]HubspotProperty `json:"properties"`
}

func getHubspotDocumentValuesByPropertyNameAndLimit(hubspotDocuments []model.HubspotDocument, propertyName string, limit int) []interface{} {
	if len(hubspotDocuments) < 1 || propertyName == "" {
		return nil
	}

	valuesAggregate := make(map[interface{}]int)
	for i := range hubspotDocuments {
		var docProperties HubspotDocumentProperties
		err := json.Unmarshal((hubspotDocuments[i].Value).RawMessage, &docProperties)
		if err != nil {
			log.WithFields(log.Fields{"document_id": hubspotDocuments[i].ID}).WithError(err).Error("Failed to unmarshal hubspot document on getAllHubspotDocumentPropertiesValue")
			continue
		}

		for name, value := range docProperties.Properties {
			if name != propertyName {
				continue
			}

			if value.Value == "" {
				continue
			}

			valuesAggregate[value.Value] = valuesAggregate[value.Value] + 1
		}
	}

	propertyValueTuples := getPropertyValueTuples(valuesAggregate, limit)
	propertyValues := make([]interface{}, len(propertyValueTuples))
	for i := range propertyValueTuples {
		propertyValues[i] = propertyValueTuples[i].Name
	}

	return propertyValues

}

func getHubspotDocumentPropertiesNameByType(hubspotDocuments []model.HubspotDocument) ([]string, []string) {
	dateTimeProperties := make(map[string]interface{})
	categoricalProperties := make(map[string]interface{})
	currentTimestamp := U.TimeNowUnix() * 1000

	for i := range hubspotDocuments {
		var docProperties HubspotDocumentProperties
		err := json.Unmarshal((hubspotDocuments[i].Value).RawMessage, &docProperties)
		if err != nil {
			log.WithError(err).Error("Failed to unmarshal hubspot document on GetHubspotObjectProperties")
			continue
		}

		for key, value := range docProperties.Properties {
			valueStr := U.GetPropertyValueAsString(value.Value)
			if valueStr == "" {
				continue
			}

			if U.IsPropertyNameContainsDateOrTime(key) {
				_, isNumber := U.ConvertDateTimeValueToNumber(value)
				if isNumber {
					dateTimeProperties[key] = true
					continue
				}
			}

			if len(valueStr) == 13 { // milliseconds format
				timestamp, err := strconv.ParseUint(valueStr, 10, 64)
				if err == nil && timestamp >= 0 && int64(timestamp) <= currentTimestamp {
					// if for some document it was passed as categorical then its not a timestamp.
					if _, exists := categoricalProperties[key]; !exists {
						dateTimeProperties[key] = true
					}
					continue
				}
			}

			// delete from datetime if already exist in it.
			if _, exists := dateTimeProperties[key]; exists {
				delete(dateTimeProperties, key)
			}

			categoricalProperties[key] = true

		}
	}

	var categoricalPropertiesArray []string
	var dateTimePropertiesArray []string
	for pName := range categoricalProperties {
		categoricalPropertiesArray = append(categoricalPropertiesArray, pName)
	}

	for pName := range dateTimeProperties {
		dateTimePropertiesArray = append(dateTimePropertiesArray, pName)
	}

	return categoricalPropertiesArray, dateTimePropertiesArray
}

func getLatestHubspotDocumentsByLimit(projectID uint64, docType int, limit int) ([]model.HubspotDocument, error) {
	if projectID == 0 {
		return nil, errors.New("invalid project_id")
	}

	if docType == 0 || limit <= 0 {
		return nil, errors.New("invalid parameters")
	}

	lookbackTimestampInMilliseconds := U.UnixTimeBeforeDuration(48*time.Hour) * 1000 //last 48 hours

	var hubspotDocuments []model.HubspotDocument
	db := C.GetServices().Db
	err := db.Model(&model.HubspotDocument{}).Where("project_id = ? AND type = ? AND action= ? AND timestamp > ?",
		projectID, docType, model.HubspotDocumentActionUpdated, lookbackTimestampInMilliseconds).Order("timestamp desc").Limit(1000).Find(&hubspotDocuments).Error
	if err != nil {
		return nil, err

	}

	return hubspotDocuments, nil
}

// GetHubspotObjectPropertiesName returns property names by type
func (pg *Postgres) GetHubspotObjectPropertiesName(ProjectID uint64, objectType string) ([]string, []string) {
	if ProjectID == 0 || objectType == "" {
		return nil, nil
	}

	docType, err := model.GetHubspotTypeByAlias(objectType)
	if err != nil {
		return nil, nil
	}

	logCtx := log.WithFields(log.Fields{"project_id": ProjectID, "doc_type": docType})

	hubspotDocuments, err := getLatestHubspotDocumentsByLimit(ProjectID, docType, 1000)
	if err != nil {
		logCtx.WithError(err).Error("Failed to GetSalesforceObjectPropertiesValues")
		return nil, nil
	}

	return getHubspotDocumentPropertiesNameByType(hubspotDocuments)
}

// GetAllHubspotObjectValuesByPropertyName returns all values by property name
func (pg *Postgres) GetAllHubspotObjectValuesByPropertyName(ProjectID uint64, objectType, propertyName string) []interface{} {
	if ProjectID == 0 || objectType == "" || propertyName == "" {
		return nil
	}

	docType, err := model.GetHubspotTypeByAlias(objectType)
	if err != nil {
		return nil
	}

	logCtx := log.WithFields(log.Fields{"project_id": ProjectID, "doc_type": docType})

	hubspotDocuments, err := getLatestHubspotDocumentsByLimit(ProjectID, docType, 1000)
	if err != nil {
		logCtx.WithError(err).Error("Failed to GetAllHubspotObjectPropertyValues")
		return nil
	}

	return getHubspotDocumentValuesByPropertyNameAndLimit(hubspotDocuments, propertyName, 100)
}

func (pg *Postgres) UpdateHubspotDocumentAsSynced(projectId uint64, id string, syncId string, timestamp int64, action int, userID string) int {
	logCtx := log.WithField("project_id", projectId).WithField("id", id)

	updates := make(map[string]interface{}, 0)
	updates["synced"] = true
	if syncId != "" {
		updates["sync_id"] = syncId
	}

	if userID != "" {
		updates["user_id"] = userID
	}

	db := C.GetServices().Db
	err := db.Model(&model.HubspotDocument{}).Where("project_id = ? AND id = ? AND timestamp= ? AND action = ?",
		projectId, id, timestamp, action).Updates(updates).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to update hubspot document as synced.")
		return http.StatusInternalServerError
	}

	return http.StatusAccepted
}

// GetLastSyncedHubspotDocumentByCustomerUserIDORUserID returns latest synced record by customer_user_id or user_id.
func (pg *Postgres) GetLastSyncedHubspotDocumentByCustomerUserIDORUserID(projectID uint64, customerUserID, userID string, docType int) (*model.HubspotDocument, int) {
	if projectID == 0 {
		return nil, http.StatusBadRequest
	}

	if userID == "" || docType == 0 {
		return nil, http.StatusBadRequest
	}

	logCtx := log.WithFields(log.Fields{"project_id": projectID, "user_id": userID, "customer_user_id": customerUserID, "doc_type": docType})

	db := C.GetServices().Db

	var whereStmn string
	var whereParams []interface{}

	if customerUserID != "" {
		userIDs, status := pg.GetAllUserIDByCustomerUserID(projectID, customerUserID)
		if status == http.StatusFound {
			whereStmn = "type = ? AND project_id=? AND user_id IN(?) AND synced = true"
			whereParams = []interface{}{docType, projectID, userIDs}
		} else {
			logCtx.Error("Failed to GetAllUserIDByCustomerUserID.")
		}
	}

	if customerUserID == "" || whereStmn == "" {
		whereStmn = "type = ? AND synced = true AND project_id=? AND user_id = ? "
		whereParams = []interface{}{docType, projectID, userID}
	}

	var document []model.HubspotDocument

	if err := db.Where(whereStmn, whereParams...).Order("timestamp DESC").First(&document).Error; err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			logCtx.WithError(err).Error("Failed to get latest hubspot document by userID.")
			return nil, http.StatusInternalServerError
		}
		return nil, http.StatusNotFound
	}

	if len(document) != 1 {
		return nil, http.StatusNotFound
	}

	return &document[0], http.StatusFound
}
