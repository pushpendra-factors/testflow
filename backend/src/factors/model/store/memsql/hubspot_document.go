package memsql

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"

	C "factors/config"
	"factors/model/model"
	"factors/util"
	U "factors/util"
)

func (store *MemSQL) satisfiesHubspotDocumentUniquenessConstraints(document *model.HubspotDocument) int {
	defer model.LogOnSlowExecutionWithParams(time.Now(), &log.Fields{"document": document})

	errCode := store.isHubspotDocumentExistByPrimaryKey(document)
	if errCode == http.StatusFound {
		return http.StatusConflict
	}
	if errCode == http.StatusNotFound {
		return http.StatusOK
	}
	return errCode
}

// Checks PRIMARY KEY constraint (project_id, id, type, action, timestamp)
func (store *MemSQL) isHubspotDocumentExistByPrimaryKey(document *model.HubspotDocument) int {
	defer model.LogOnSlowExecutionWithParams(time.Now(), &log.Fields{"document": document})

	logCtx := log.WithField("document", document)

	if document.ProjectId == 0 || document.ID == "" || document.Type == 0 ||
		document.Action == 0 || document.Timestamp == 0 {

		log.Error("Invalid hubspot document on primary constraint check.")
		return http.StatusBadRequest
	}

	var hubspotDocument model.HubspotDocument

	db := C.GetServices().Db
	if err := db.Limit(1).Where("project_id = ? AND id = ? AND type = ? AND action = ? AND timestamp = ?",
		document.ProjectId, document.ID, document.Type, document.Action, document.Timestamp,
	).Select("id").Find(&hubspotDocument).Error; err != nil {

		if gorm.IsRecordNotFoundError(err) {
			return http.StatusNotFound
		}

		logCtx.WithError(err).
			Error("Failed getting to check existence hubspot document by primary keys.")
		return http.StatusInternalServerError
	}

	if hubspotDocument.ID == "" {
		logCtx.Error("Invalid id value returned on hubspot document primary key check.")
		return http.StatusInternalServerError
	}

	return http.StatusFound
}

func getHubspotDocumentId(document *model.HubspotDocument) (string, error) {
	defer model.LogOnSlowExecutionWithParams(time.Now(), &log.Fields{"document": document})

	if document.Type == 0 {
		return "", model.ErrorHubspotInvalidHubspotDocumentType
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
		if document.Action == model.HubspotDocumentActionDeleted {
			idKey = "id"
		}
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

func isExistHubspotDocumentByIDAndType(projectId uint64, id string, docType int) int {
	argFields := log.Fields{"project_id": projectId, "id": id, "type": docType}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &argFields)

	logCtx := log.WithFields(argFields)

	var document model.HubspotDocument
	if projectId == 0 || id == "" || docType == 0 {
		logCtx.Error("Failed to get hubspot document by id and type. Invalid project_id or id or type.")
		return http.StatusBadRequest
	}

	db := C.GetServices().Db
	err := db.Where("project_id = ? AND id = ? AND type = ? AND action = ? ", projectId, id,
		docType, model.HubspotDocumentActionCreated).Select("id").Limit(1).Find(&document).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return http.StatusNotFound
		}

		logCtx.WithError(err).Error("Failed to get hubspot documents.")
		return http.StatusInternalServerError
	}

	if document.ID == "" {
		return http.StatusNotFound
	}

	return http.StatusFound
}

func (store *MemSQL) GetHubspotContactCreatedSyncIDAndUserID(projectID uint64, docID string) ([]model.HubspotDocument, int) {
	argFields := log.Fields{"project_id": projectID, "doc_id": docID}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &argFields)

	logCtx := log.WithFields(argFields)

	if projectID == 0 || docID == "" {
		logCtx.Error("Invalid parameters on GetHubspotContactCreatedSyncIDAndUserID.")
		return nil, http.StatusBadRequest
	}

	documents := []model.HubspotDocument{}

	db := C.GetServices().Db
	err := db.Select("sync_id, user_id, timestamp").Where("project_id = ? AND id = ? AND type = ? AND action = ? AND synced=true",
		projectID, docID, model.HubspotDocumentTypeContact, model.HubspotDocumentActionCreated).Find(&documents).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to get hubspot contact created document.")
		return nil, http.StatusInternalServerError
	}

	if len(documents) < 1 {
		return nil, http.StatusNotFound
	}

	if len(documents) > 1 {

		return documents, http.StatusMultipleChoices
	}

	return documents, http.StatusFound
}

func (store *MemSQL) GetHubspotDocumentByTypeAndActions(projectId uint64, ids []string,
	docType int, actions []int) ([]model.HubspotDocument, int) {
	argFields := log.Fields{"project_id": projectId, "ids": ids,
		"type": docType, "actions": actions}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &argFields)

	logCtx := log.WithFields(argFields)

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

// GetSyncedHubspotDocumentByFilter get hubspot synced document by id and action
func (store *MemSQL) GetSyncedHubspotDocumentByFilter(projectID uint64,
	ID string, docType, action int) (*model.HubspotDocument, int) {

	argFields := log.Fields{"project_id": projectID, "type": docType, "action": action}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &argFields)

	logCtx := log.WithFields(argFields)

	var document model.HubspotDocument
	if projectID == 0 || ID == "" || docType == 0 || action == 0 {
		logCtx.Error("Failed to get hubspot document. Invalid params.")
		return nil, http.StatusBadRequest
	}

	db := C.GetServices().Db
	err := db.Limit(1).
		Where("project_id = ? AND id = ? AND type = ? AND action = ? AND synced=true",
			projectID, ID, docType, action).Find(&document).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}

		logCtx.WithError(err).Error("Failed to get hubspot document with user_id.")
		return nil, http.StatusInternalServerError
	}

	return &document, http.StatusFound
}

func (store *MemSQL) CreateHubspotDocument(projectId uint64, document *model.HubspotDocument) int {
	defer model.LogOnSlowExecutionWithParams(time.Now(), &log.Fields{"project_id": projectId})

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

	errCode := isExistHubspotDocumentByIDAndType(document.ProjectId,
		document.ID, document.Type)
	if errCode == http.StatusInternalServerError || errCode == http.StatusBadRequest {
		return errCode
	}
	isNew := errCode == http.StatusNotFound

	createdTimestamp, err := model.GetHubspotDocumentCreatedTimestamp(document)
	if err != nil {
		if err != model.ErrorHubspotUsingFallbackKey {
			logCtx.WithField("action", document.Action).WithError(err).Error(
				"Failed to get created timestamp from hubspot document on create.")
			return http.StatusInternalServerError
		}

		logCtx.WithField("action", document.Action).WithError(err).Error("Missing document key.")
	}

	updatedTimestamp, err := model.GetHubspotDocumentUpdatedTimestamp(document)
	if err != nil {
		if err != model.ErrorHubspotUsingFallbackKey {
			logCtx.WithField("action", document.Action).WithError(err).Error(
				"Failed to get updated timestamp from hubspot document on create.")
			return http.StatusInternalServerError
		}

		logCtx.WithField("action", document.Action).WithError(err).Error("Missing document key.")
	}

	var updatedDocument model.HubspotDocument // use for duplicating new document to updated document.
	if isNew {
		// Skip adding the record if deleted record is to added for
		// non-existing document.
		if document.Action == model.HubspotDocumentActionDeleted {
			return http.StatusOK
		}
		updatedDocument = *document
		document.Action = model.HubspotDocumentActionCreated // created
		document.Timestamp = createdTimestamp
	} else {
		if document.Action != model.HubspotDocumentActionDeleted {
			document.Action = model.HubspotDocumentActionUpdated // updated
		}
		// Any update on the entity would create a new hubspot document.
		// i.e, deal will be synced after updating a created deal with a
		// contact or a company.
		document.Timestamp = updatedTimestamp
	}

	errCode = store.satisfiesHubspotDocumentUniquenessConstraints(document)
	if errCode != http.StatusOK {
		return errCode
	}

	db := C.GetServices().Db
	err = db.Create(document).Error
	if err != nil {
		if IsDuplicateRecordError(err) {
			return http.StatusConflict
		}

		logCtx.WithError(err).Error("Failed to create hubspot document.")
		return http.StatusInternalServerError
	}

	if isNew { // create updated document for new user
		updatedDocument.Action = model.HubspotDocumentActionUpdated
		updatedDocument.Timestamp = createdTimestamp
		recentUpdatedDocument := updatedDocument
		err = db.Create(&updatedDocument).Error
		if err != nil {
			if IsDuplicateRecordError(err) {
				return http.StatusConflict
			}

			logCtx.WithError(err).Error("Failed to create updated hubspot document.")
			return http.StatusInternalServerError
		}

		if updatedTimestamp > createdTimestamp {
			recentUpdatedDocument.Action = model.HubspotDocumentActionUpdated
			recentUpdatedDocument.Timestamp = updatedTimestamp
			err = db.Create(&recentUpdatedDocument).Error
			if err != nil {
				if IsDuplicateRecordError(err) {
					return http.StatusConflict
				}

				logCtx.WithError(err).Error("Failed to create recent updated hubspot document.")
				return http.StatusInternalServerError
			}
		}
	}
	UpdateCountCacheByDocumentType(projectId, &document.CreatedAt, "hubspot")
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

func (store *MemSQL) updateHubspotProjectSettingsLastSyncInfo(projectID uint64, incomingSyncInfo map[string]int64) error {
	defer model.LogOnSlowExecutionWithParams(time.Now(), &log.Fields{"project_id": projectID})

	logCtx := log.WithFields(log.Fields{"project_id": projectID})

	if projectID == 0 || incomingSyncInfo == nil {
		logCtx.Error("Missing required fields.")
		return errors.New("missing required fields")
	}

	projectSetting, status := store.GetProjectSetting(projectID)
	if status != http.StatusFound {
		logCtx.Error("Failed to get project setttings on hubspot last sync info.")
		return errors.New("failed to get project settings ")
	}

	existingSyncInfoMap, err := model.GetHubspotDecodedSyncInfo(projectSetting.IntHubspotSyncInfo)
	if err != nil {
		logCtx.WithError(err).Error("Failed to decode project setting on hubspot last sync info.")
		return err
	}

	updatedSyncInfo := model.GetHubspotSyncUpdatedInfo(&incomingSyncInfo, existingSyncInfoMap)

	enlastSyncInfo, err := json.Marshal(updatedSyncInfo)
	if err != nil {
		logCtx.WithError(err).Error("Failed to encode hubspot last sync info.")
		return err
	}

	pJSONLastSyncInfo := postgres.Jsonb{RawMessage: enlastSyncInfo}
	_, status = store.UpdateProjectSettings(projectID, &model.ProjectSetting{IntHubspotSyncInfo: &pJSONLastSyncInfo})
	if status != http.StatusAccepted {
		logCtx.Error("Failed to update hubspot last sync info on success.")
		return errors.New("Failed to update hubspot last sync info")
	}

	return nil
}

// UpdateHubspotProjectSettingsBySyncStatus update hubspot sync project settings
func (store *MemSQL) UpdateHubspotProjectSettingsBySyncStatus(success []model.HubspotProjectSyncStatus,
	failure []model.HubspotProjectSyncStatus, syncALl bool) int {
	defer model.LogOnSlowExecutionWithParams(time.Now(),
		&log.Fields{"success": success, "failure": failure, "sync_all": syncALl})

	anyFailure := false
	if syncALl {
		syncStatus, status := model.GetHubspotProjectOverAllStatus(success, failure)
		for pid, projectSuccess := range status {
			if projectSuccess {
				_, status := store.UpdateProjectSettings(pid, &model.ProjectSetting{
					IntHubspotFirstTimeSynced: true,
				})

				if status != http.StatusAccepted {
					log.WithFields(log.Fields{"project_id": pid}).
						Error("Failed to update hubspot first time sync status on success.")
					anyFailure = true
				}

				err := store.updateHubspotProjectSettingsLastSyncInfo(pid, syncStatus[pid])
				if err != nil {
					log.WithFields(log.Fields{"project_id": pid}).WithError(err).Error("Failed to update hubspot last sync info.")
					anyFailure = true
				}
			}
		}

		if anyFailure {
			return http.StatusInternalServerError
		}

		return http.StatusAccepted
	}

	syncStatus, _ := model.GetHubspotProjectOverAllStatus(success, failure)

	for pid, docTypeStatus := range syncStatus {
		err := store.updateHubspotProjectSettingsLastSyncInfo(pid, docTypeStatus)
		if err != nil {
			log.WithFields(log.Fields{"project_id": pid}).WithError(err).Error("Failed to update hubspot last sync info.")
			anyFailure = true
		}
	}

	if anyFailure {
		return http.StatusInternalServerError
	}

	return http.StatusAccepted
}

// GetHubspotFirstSyncProjectsInfo return list of projects to be synced for first time
func (store *MemSQL) GetHubspotFirstSyncProjectsInfo() (*model.HubspotSyncInfo, int) {
	defer model.LogOnSlowExecutionWithParams(time.Now(), nil)

	// project sync of hubspot enable projects.
	enabledProjectLastSync := make(map[uint64]map[string]int64, 0)

	// get project settings of hubspot enabled projects.
	projectSettings, errCode := store.GetAllHubspotProjectSettings()
	if errCode != http.StatusFound {
		return nil, http.StatusInternalServerError
	}

	settingsByProject := make(map[uint64]*model.HubspotProjectSettings, 0)
	for i, ps := range projectSettings {
		if ps.IsFirstTimeSynced {
			continue
		}

		// add types not synced before.
		for typ := range model.HubspotDocumentTypeAlias {
			if _, exist := enabledProjectLastSync[ps.ProjectId]; !exist {
				enabledProjectLastSync[ps.ProjectId] = make(map[string]int64)
			}

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

func (store *MemSQL) GetHubspotSyncInfo() (*model.HubspotSyncInfo, int) {
	defer model.LogOnSlowExecutionWithParams(time.Now(), nil)

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
		if _, projectExists := lastSyncInfoByProject[syncInfo.ProjectID]; !projectExists {
			lastSyncInfoByProject[syncInfo.ProjectID] = make(map[string]int64)
		}

		lastSyncInfoByProject[syncInfo.ProjectID][getHubspotTypeAlias(syncInfo.Type)] = syncInfo.Timestamp
	}

	// project sync of hubspot enable projects.
	enabledProjectLastSync := make(map[uint64]map[string]int64, 0)

	// get project settings of hubspot enaled projects.
	projectSettings, errCode := store.GetAllHubspotProjectSettings()
	if errCode != http.StatusFound {
		return nil, http.StatusInternalServerError
	}

	settingsByProject := make(map[uint64]*model.HubspotProjectSettings, 0)
	for i, ps := range projectSettings {
		if !ps.IsFirstTimeSynced {
			continue
		}

		_, pExists := lastSyncInfoByProject[ps.ProjectId]

		if !pExists {
			// add projects not synced before.
			enabledProjectLastSync[ps.ProjectId] = make(map[string]int64, 0)
		} else {
			// add sync info if avaliable.
			enabledProjectLastSync[ps.ProjectId] = lastSyncInfoByProject[ps.ProjectId]
		}

		// overwrite last syncinfo from project settings
		if projectSettings[i].SyncInfo != nil {
			lastSyncInfoMap, err := util.DecodePostgresJsonbAsPropertiesMap(projectSettings[i].SyncInfo)
			if err != nil {
				log.WithFields(log.Fields{"project_id": ps.ProjectId}).WithError(err).
					Error("Failed to decode hubspot last sync info.")
			} else {
				for docType, timestampInt := range *lastSyncInfoMap {
					timestamp, err := util.GetPropertyValueAsFloat64(timestampInt)
					if err != nil {
						log.WithFields(log.Fields{"project_id": ps.ProjectId}).WithError(err).
							Error("Failed to get timestamp for hubspot last sync info.")
					} else {
						enabledProjectLastSync[ps.ProjectId][docType] = int64(timestamp)
					}

				}
			}
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

func (store *MemSQL) GetHubspotFormDocuments(projectId uint64) ([]model.HubspotDocument, int) {
	defer model.LogOnSlowExecutionWithParams(time.Now(), &log.Fields{"project_id": projectId})

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

func (store *MemSQL) GetHubspotDocumentsByTypeForSync(projectId uint64, typ int) ([]model.HubspotDocument, int) {
	argFields := log.Fields{"project_id": projectId, "type": typ}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &log.Fields{"project_id": projectId, "typ": typ})

	logCtx := log.WithFields(argFields)

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

// GetHubspotDocumentBeginingTimestampByDocumentTypeForSync returns the minimum timestamp for unsynced document
func (store *MemSQL) GetHubspotDocumentBeginingTimestampByDocumentTypeForSync(projectID uint64) (int64, int) {
	defer model.LogOnSlowExecutionWithParams(time.Now(), &log.Fields{"project_id": projectID})

	logCtx := log.WithFields(log.Fields{"project_id": projectID})

	if projectID == 0 {
		logCtx.Error("Invalid project_id.")
		return 0, http.StatusBadRequest
	}

	db := C.GetServices().Db
	rows, err := db.Raw("SELECT MIN(timestamp) FROM hubspot_documents WHERE project_id=? AND synced=false", projectID).Rows()
	if err != nil {
		log.WithError(err).Error("Failed to get hubspot minimum timestamp.")
		return 0, http.StatusInternalServerError
	}

	var minTimestamp *int64
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&minTimestamp); err != nil {
			log.WithError(err).Error("Failed scanning rows on get hubspot minimum timestamp for sync.")
		}
	}

	if minTimestamp == nil {
		return 0, http.StatusNotFound
	}

	return *minTimestamp, http.StatusFound
}

// GetHubspotDocumentsByTypeANDRangeForSync return list of documents unsynced for given time range
func (store *MemSQL) GetHubspotDocumentsByTypeANDRangeForSync(projectID uint64,
	docType int, from, to int64) ([]model.HubspotDocument, int) {

	argFields := log.Fields{"project_id": projectID, "type": docType, "from": from, "to": to}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &argFields)
	logCtx := log.WithFields(argFields)

	if projectID == 0 || docType == 0 || from < 0 || to < 0 {
		logCtx.Error("Invalid project_id or type on get hubspot documents by type.")
		return nil, http.StatusBadRequest
	}

	var documents []model.HubspotDocument

	db := C.GetServices().Db
	err := db.Order("timestamp, created_at ASC").Where("project_id=? AND type=? AND synced=false AND timestamp BETWEEN ? AND ?",
		projectID, docType, from, to).Find(&documents).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to get hubspot documents by type.")
		return nil, http.StatusInternalServerError
	}

	return documents, http.StatusFound
}

func (store *MemSQL) GetSyncedHubspotDealDocumentByIdAndStage(projectId uint64, id string,
	stage string) (*model.HubspotDocument, int) {

	argFields := log.Fields{"project_id": projectId, "id": id, "stage": stage}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &argFields)
	logCtx := log.WithFields(argFields)

	if projectId == 0 || id == "" || stage == "" {
		logCtx.Error(
			"Invalid project_id or id or stage on get hubspot synced deal by id and stage.")
		return nil, http.StatusBadRequest
	}

	var documents []model.HubspotDocument

	db := C.GetServices().Db
	err := db.Limit(1).Where(
		"project_id=? AND id=? AND type=? AND synced=true AND JSON_EXTRACT_STRING(value, 'properties', 'dealstage', 'value')=?",
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

func getHubspotDocumentValuesByPropertyNameAndLimit(hubspotDocuments []model.HubspotDocument,
	propertyName string, limit int) []interface{} {

	if len(hubspotDocuments) < 1 || propertyName == "" {
		return nil
	}

	valuesAggregate := make(map[interface{}]int)
	for i := range hubspotDocuments {
		var docProperties model.HubspotDocumentProperties
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
		var docProperties model.HubspotDocumentProperties
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
	defer model.LogOnSlowExecutionWithParams(time.Now(),
		&log.Fields{"project_id": projectID, "doc_type": docType, "limit": limit})

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
func (store *MemSQL) GetHubspotObjectPropertiesName(ProjectID uint64, objectType string) ([]string, []string) {
	defer model.LogOnSlowExecutionWithParams(time.Now(),
		&log.Fields{"project_id": ProjectID, "object_type": objectType})

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
func (store *MemSQL) GetAllHubspotObjectValuesByPropertyName(ProjectID uint64,
	objectType, propertyName string) []interface{} {

	defer model.LogOnSlowExecutionWithParams(time.Now(), &log.Fields{"project_id": ProjectID,
		"object_type": objectType, "property_name": propertyName})

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

func (store *MemSQL) UpdateHubspotDocumentAsSynced(projectId uint64, id string, docType int,
	syncId string, timestamp int64, action int, userID string) int {

	defer model.LogOnSlowExecutionWithParams(time.Now(),
		&log.Fields{"project_id": projectId, "doc_type": docType, "id": id,
			"sync_id": syncId, "timestamp": timestamp, "action": action, "user_id": userID})

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
	err := db.Model(&model.HubspotDocument{}).Where("project_id = ? AND id = ? AND timestamp= ? AND action = ? AND type= ?",
		projectId, id, timestamp, action, docType).Updates(updates).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to update hubspot document as synced.")
		return http.StatusInternalServerError
	}

	return http.StatusAccepted
}

// GetLastSyncedHubspotDocumentByID returns latest synced record by document id.
func (store *MemSQL) GetLastSyncedHubspotDocumentByID(projectID uint64, docID string, docType int) (*model.HubspotDocument, int) {
	argFields := log.Fields{"project_id": projectID, "doc_id": docID, "doc_type": docType}
	model.LogOnSlowExecutionWithParams(time.Now(), &argFields)
	logCtx := log.WithFields(argFields)

	if projectID == 0 || docType == 0 || docID == "" {
		logCtx.Error("Missing required field")
		return nil, http.StatusBadRequest
	}

	db := C.GetServices().Db

	var document []model.HubspotDocument

	if err := db.Where("project_id = ? AND type = ? AND id = ? and synced=true",
		projectID, docType, docID).Order("timestamp DESC").First(&document).Error; err != nil {
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
