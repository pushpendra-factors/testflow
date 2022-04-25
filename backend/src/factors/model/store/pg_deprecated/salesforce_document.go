package postgres

import (
	"encoding/json"
	"errors"
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

// GetSalesforceSyncInfo returns list of projects and their corresponding sync status
func (pg *Postgres) GetSalesforceSyncInfo() (model.SalesforceSyncInfo, int) {
	var lastSyncInfo []model.SalesforceLastSyncInfo
	var syncInfo model.SalesforceSyncInfo

	db := C.GetServices().Db
	err := db.Table("salesforce_documents").Select(
		"project_id, type, MAX(timestamp) as timestamp").Group(
		"project_id, type").Find(&lastSyncInfo).Error
	if err != nil {
		return syncInfo, http.StatusInternalServerError
	}

	lastSyncInfoByProject := make(map[uint64]map[string]int64, 0)
	for _, syncInfo := range lastSyncInfo {
		if _, projectExists := lastSyncInfoByProject[syncInfo.ProjectID]; !projectExists {
			lastSyncInfoByProject[syncInfo.ProjectID] = make(map[string]int64)
		}

		lastSyncInfoByProject[syncInfo.ProjectID][model.GetSalesforceAliasByDocType(syncInfo.Type)] = syncInfo.Timestamp
	}

	enabledProjectLastSync := make(map[uint64]map[string]int64, 0)

	projectSettings, errCode := pg.GetAllSalesforceProjectSettings()
	if errCode != http.StatusFound {
		return syncInfo, http.StatusInternalServerError
	}

	settingsByProject := make(map[uint64]*model.SalesforceProjectSettings, 0)
	for i, ps := range projectSettings {
		_, pExists := lastSyncInfoByProject[ps.ProjectID]

		if !pExists {
			// add projects not synced before.
			enabledProjectLastSync[ps.ProjectID] = make(map[string]int64, 0)
		} else {
			// add sync info if available.
			enabledProjectLastSync[ps.ProjectID] = lastSyncInfoByProject[ps.ProjectID]
		}

		// add types not synced before.
		for typ := range model.GetSalesforceDocumentTypeAlias(ps.ProjectID) {
			_, typExists := enabledProjectLastSync[ps.ProjectID][typ]
			if !typExists {
				// last sync timestamp as zero as type not synced before.
				enabledProjectLastSync[ps.ProjectID][typ] = 0
			}
		}

		settingsByProject[projectSettings[i].ProjectID] = &projectSettings[i]
	}

	syncInfo.LastSyncInfo = enabledProjectLastSync
	syncInfo.ProjectSettings = settingsByProject

	return syncInfo, http.StatusFound
}

func getSalesforceDocumentID(document *model.SalesforceDocument) (string, error) {
	documentMap, err := U.DecodePostgresJsonb(document.Value)
	if err != nil {
		return "", err
	}

	id, idExists := (*documentMap)["Id"]
	if !idExists {
		return "", errors.New("id key not exist on salesforce document")
	}

	idAsString := U.GetPropertyValueAsString(id)
	if idAsString == "" {
		return "", errors.New("invalid id on salesforce document")
	}
	return idAsString, nil
}

// GetSyncedSalesforceDocumentByType return salesforce_documents by doc type which are synced on chronological order
func (pg *Postgres) GetSyncedSalesforceDocumentByType(projectID uint64, ids []string,
	docType int, includeUnSynced bool) ([]model.SalesforceDocument, int) {

	logCtx := log.WithFields(log.Fields{"project_id": projectID, "ids": ids,
		"type": docType})

	var documents []model.SalesforceDocument
	if projectID == 0 || len(ids) == 0 || docType == 0 {
		logCtx.Error("Failed to get salesforce document by id and type. Invalid project_id or id or type.")
		return nil, http.StatusBadRequest
	}

	stmnt := "project_id = ? AND id IN (?) AND type = ?"
	if !includeUnSynced {
		stmnt = stmnt + " AND " + "synced=true "
	}

	db := C.GetServices().Db
	err := db.Order("timestamp").Where(stmnt,
		projectID, ids, docType).Find(&documents).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to get salesforce documents.")
		return nil, http.StatusInternalServerError
	}

	if len(documents) == 0 {
		return nil, http.StatusNotFound
	}

	return documents, http.StatusFound
}

//GetLatestSalesforceDocumentByID return latest synced or unsynced document
func (pg *Postgres) GetLatestSalesforceDocumentByID(projectID uint64, documentIDs []string, docType int, maxTimestamp int64) ([]model.SalesforceDocument, int) {
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "id": documentIDs, "type": docType})

	if projectID == 0 || len(documentIDs) < 1 || docType == 0 {
		logCtx.Error("Failed to get salesforce document by id and type.")
		return nil, http.StatusBadRequest
	}

	whereStmnt := "project_id = ? AND id IN (?) AND type = ?"
	whereParams := []interface{}{projectID, documentIDs, docType}
	if maxTimestamp > 0 {
		whereStmnt = whereStmnt + " AND " + "timestamp <= ?"
		whereParams = append(whereParams, maxTimestamp)
	}
	db := C.GetServices().Db
	var documents []model.SalesforceDocument
	err := db.Select("DISTINCT ON(id) salesforce_documents.*").Where(whereStmnt, whereParams...).Order("id,timestamp desc").Find(&documents).Error
	if err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}

		logCtx.WithError(err).Error("Failed to get latest salesforce document by id and type.")
		return nil, http.StatusInternalServerError
	}

	if len(documents) < 1 {
		return nil, http.StatusNotFound
	}
	return documents, http.StatusFound
}

func getSalesforceDocumentByIDAndType(projectID uint64, id string, docType int) ([]model.SalesforceDocument, int) {
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "id": id, "type": docType})

	var documents []model.SalesforceDocument
	if projectID == 0 || id == "" || docType == 0 {
		logCtx.Error("Failed to get salesforce document by id and type. Invalid project_id or id or type.")
		return documents, http.StatusBadRequest
	}

	db := C.GetServices().Db
	err := db.Where("project_id = ? AND id = ? AND type = ?", projectID, id,
		docType).Find(&documents).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to get salesforce documents.")
		return documents, http.StatusInternalServerError
	}

	if len(documents) == 0 {
		return documents, http.StatusNotFound
	}

	return documents, http.StatusFound
}

// CreateSalesforceDocument fills required fields before inserting into salesforce_document table
func (pg *Postgres) CreateSalesforceDocument(projectID uint64, document *model.SalesforceDocument) int {
	logCtx := log.WithField("project_id", document.ProjectID)
	if projectID == 0 {
		logCtx.Error("Invalid project_id on create salesforce document.")
		return http.StatusBadRequest
	}
	document.ProjectID = projectID

	document.Type = model.GetSalesforceDocTypeByAlias(document.TypeAlias)

	if U.IsEmptyPostgresJsonb(document.Value) {
		logCtx.Error("Empty document value on create salesforce document.")
		return http.StatusBadRequest
	}

	documentID, err := getSalesforceDocumentID(document)
	if err != nil {
		logCtx.WithError(err).Error(
			"Failed to get id for salesforce document on create.")
		return http.StatusInternalServerError
	}
	document.ID = documentID

	logCtx = logCtx.WithField("type", document.Type).WithField("value", document.Value)

	newBytes := U.RemoveNullCharacterBytes(document.Value.RawMessage)

	if len(newBytes) != len(document.Value.RawMessage) {
		log.WithFields(log.Fields{"document_id": document.ID, "project_id": document.ProjectID,
			"raw_message":    string(document.Value.RawMessage),
			"sliced_message": string(newBytes)}).Warn("Using new sliced bytes for null character.")
		document.Value.RawMessage = newBytes
	}

	_, errCode := getSalesforceDocumentByIDAndType(document.ProjectID,
		document.ID, document.Type)
	if errCode == http.StatusInternalServerError || errCode == http.StatusBadRequest {
		return errCode
	}

	isNew := errCode == http.StatusNotFound
	if isNew {
		status := pg.CreateSalesforceDocumentByAction(projectID, document, model.SalesforceDocumentCreated)
		if status != http.StatusOK {
			if status != http.StatusConflict {
				logCtx.Error("Failed to create salesforce document.")
			}

			return status
		}
		UpdateCountCacheByDocumentType(projectID, &document.CreatedAt, "salesforce")
		return http.StatusCreated
	}

	status := pg.CreateSalesforceDocumentByAction(projectID, document, model.SalesforceDocumentUpdated)
	if status != http.StatusOK {
		if status != http.StatusConflict {
			logCtx.Error("Failed to create salesforce document.")
		}

		return status
	}
	UpdateCountCacheByDocumentType(projectID, &document.CreatedAt, "salesforce")
	return http.StatusCreated
}

// CreateSalesforceDocumentByAction inserts salesforce_document to table by SalesforceAction
func (pg *Postgres) CreateSalesforceDocumentByAction(projectID uint64, document *model.SalesforceDocument, action model.SalesforceAction) int {
	if projectID == 0 {
		return http.StatusBadRequest
	}

	if action == 0 {
		return http.StatusBadRequest
	}

	document.Action = action
	timestamp, err := model.GetSalesforceLastModifiedTimestamp(document)
	if err != nil {
		log.WithError(err).Error("Failed to get last modified timestamp")
		return http.StatusBadRequest
	}
	document.Timestamp = timestamp

	db := C.GetServices().Db
	err = db.Create(document).Error
	if err != nil {
		if U.IsPostgresUniqueIndexViolationError("salesforce_documents_pkey", err) {
			return http.StatusConflict
		}

		log.WithError(err).Error("Failed to create salesforce document")
		return http.StatusInternalServerError
	}
	return http.StatusOK
}

func getSalesforceDocumentPropertiesByCategory(salesforceDocument []model.SalesforceDocument) ([]string, []string) {

	categoricalProperties := make(map[string]bool)
	dateTimeProperties := make(map[string]bool)

	var categoricalPropertiesArray []string
	var dateTimePropertiesArray []string

	for i := range salesforceDocument {

		var docProperties map[string]interface{}
		err := json.Unmarshal((salesforceDocument[i].Value).RawMessage, &docProperties)
		if err != nil {
			log.WithError(err).Error("Failed to unmarshal salesforce document on getSalesforceDocumentPropertiesByCategory")
			continue
		}

		for key, value := range docProperties {
			if _, err := model.GetSalesforceDocumentTimestamp(value); err == nil ||
				strings.Contains(strings.ToLower(key), "date") {

				dateTimeProperties[key] = true
			} else {
				categoricalProperties[key] = true
			}
		}

	}

	for pName := range categoricalProperties {
		categoricalPropertiesArray = append(categoricalPropertiesArray, pName)
	}

	for pName := range dateTimeProperties {
		dateTimePropertiesArray = append(dateTimePropertiesArray, pName)
	}

	return categoricalPropertiesArray, dateTimePropertiesArray
}

// ValuesCount object holds property value name and its frequency
type ValuesCount struct {
	Name  interface{}
	Count int
}

// getPropertyValueTuples return property values by limit, if distinct values is over limit most frequent is picked
func getPropertyValueTuples(valuesAggregate map[interface{}]int, limit int) []ValuesCount {

	var aggValues []ValuesCount
	for name, count := range valuesAggregate {
		aggValues = append(aggValues, ValuesCount{Name: name, Count: count})
	}

	if len(aggValues) > limit {

		sort.Slice(aggValues, func(i, j int) bool {
			return aggValues[i].Count > aggValues[j].Count
		})

		aggValues = aggValues[:limit]
	}

	return aggValues
}

// getSalesforceDocumentValuesByPropertyAndLimit return values by property name. If unique values is above limit, top n frequent value is returned
func getSalesforceDocumentValuesByPropertyAndLimit(salesforceDocument []model.SalesforceDocument, propertyName string, limit int) []interface{} {
	if len(salesforceDocument) < 1 {
		return nil
	}

	valuesAggregate := make(map[interface{}]int, 0)
	for i := range salesforceDocument {

		var docProperties map[string]interface{}
		err := json.Unmarshal((salesforceDocument[i].Value).RawMessage, &docProperties)
		if err != nil {
			log.WithFields(log.Fields{"document_id": salesforceDocument[i].ID}).WithError(err).Error("Failed to unmarshal salesforce document on getSalesforceDocumentPropertiesByCategory")
			continue
		}

		for name, value := range docProperties {
			if name != propertyName {
				continue
			}

			if value == nil || value == "" {
				continue
			}

			valuesAggregate[value] = valuesAggregate[value] + 1
		}
	}

	propertyValueTuples := getPropertyValueTuples(valuesAggregate, limit)
	propertyValues := make([]interface{}, len(propertyValueTuples))
	for i := range propertyValueTuples {
		propertyValues[i] = propertyValueTuples[i].Name
	}

	return propertyValues
}

func getLatestSalesforceDocumetsByLimit(projectID uint64, docType int, limit int, lookbackTimeHr int) ([]model.SalesforceDocument, error) {
	if projectID == 0 {
		return nil, errors.New("invalid project_id")
	}

	if docType == 0 || limit <= 0 {
		return nil, errors.New("invalid parameter")
	}

	var salesforceDocument []model.SalesforceDocument
	lbTimestamp := U.UnixTimeBeforeDuration(time.Duration(lookbackTimeHr) * time.Hour)
	db := C.GetServices().Db
	err := db.Model(&model.SalesforceDocument{}).Where("project_id = ? AND type = ? AND action = ? AND timestamp > ?",
		projectID, docType, model.SalesforceDocumentUpdated, lbTimestamp).Order("timestamp desc").Limit(limit).Find(&salesforceDocument).Error
	if err != nil {
		return nil, err
	}

	return salesforceDocument, nil
}

// GetSalesforceObjectPropertiesName returns object property names by type
func (pg *Postgres) GetSalesforceObjectPropertiesName(ProjectID uint64, objectType string) ([]string, []string) {
	if ProjectID == 0 || objectType == "" {
		return nil, nil
	}

	docType := model.GetSalesforceDocTypeByAlias(objectType)
	if docType == 0 {
		return nil, nil
	}

	logCtx := log.WithFields(log.Fields{"project_id": ProjectID, "doc_type": docType})
	salesforceDocument, err := getLatestSalesforceDocumetsByLimit(ProjectID, docType, 1000, C.GetSalesforcePropertyLookBackTimeHr())
	if err != nil {
		logCtx.WithError(err).Error("Failed to GetSalesforceObjectPropertiesName")
		return nil, nil
	}

	return getSalesforceDocumentPropertiesByCategory(salesforceDocument)
}

// GetSalesforceObjectValuesByPropertyName returns object values by property name
func (pg *Postgres) GetSalesforceObjectValuesByPropertyName(ProjectID uint64, objectType string, propertyName string) []interface{} {
	if ProjectID == 0 || objectType == "" || propertyName == "" {
		return nil
	}

	docType := model.GetSalesforceDocTypeByAlias(objectType)
	if docType == 0 {
		return nil
	}

	logCtx := log.WithFields(log.Fields{"project_id": ProjectID, "doc_type": docType})
	salesforceDocument, err := getLatestSalesforceDocumetsByLimit(ProjectID, docType, 1000, C.GetSalesforcePropertyLookBackTimeHr())
	if err != nil {
		logCtx.WithError(err).Error("Failed to GetSalesforceObjectPropertiesValues")
		return nil
	}

	return getSalesforceDocumentValuesByPropertyAndLimit(salesforceDocument, propertyName, 100)
}

// GetLastSyncedSalesforceDocumentByCustomerUserIDORUserID returns latest synced record by customer_user_id or user_id.
func (pg *Postgres) GetLastSyncedSalesforceDocumentByCustomerUserIDORUserID(projectID uint64, customerUserID, userID string, docType int) (*model.SalesforceDocument, int) {
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

	var document []model.SalesforceDocument

	if err := db.Where(whereStmn, whereParams...).Order("timestamp DESC").First(&document).Error; err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			logCtx.WithError(err).Error("Failed to get latest salesforce document by userID.")
			return nil, http.StatusInternalServerError
		}
		return nil, http.StatusNotFound
	}
	if len(document) != 1 {
		return nil, http.StatusNotFound
	}

	return &document[0], http.StatusFound
}

// UpdateSalesforceDocumentBySyncStatus inserts syncID and updates the status of the document as synced
func (pg *Postgres) UpdateSalesforceDocumentBySyncStatus(projectID uint64, document *model.SalesforceDocument, syncID, userID, groupUserID string, synced bool) int {
	logCtx := log.WithField("project_id", projectID).WithField("id", document.ID)

	updates := make(map[string]interface{}, 0)
	if synced {
		updates["synced"] = synced
	}

	if syncID != "" {
		updates["sync_id"] = syncID
	}

	if userID != "" {
		updates["user_id"] = userID
	}

	if groupUserID != "" {
		updates["group_user_id"] = groupUserID
	}

	db := C.GetServices().Db
	err := db.Model(&model.SalesforceDocument{}).Where("project_id = ? AND id = ? AND timestamp = ? AND type = ? AND action = ?",
		projectID, document.ID, document.Timestamp, document.Type, document.Action).Updates(updates).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to update salesforce document as synced.")
		return http.StatusInternalServerError
	}

	return http.StatusAccepted
}

// BuildAndUpsertDocument creates new salesforce_document for insertion
func (pg *Postgres) BuildAndUpsertDocument(projectID uint64, objectName string, value model.SalesforceRecord) error {
	if projectID == 0 {
		return errors.New("invalid project id")
	}
	if objectName == "" || value == nil {
		return errors.New("invalid oject name or value")
	}

	if len(value) == 0 {
		return errors.New("empty value")
	}

	var document model.SalesforceDocument
	document.ProjectID = projectID
	document.TypeAlias = objectName
	enValue, err := json.Marshal(value)
	if err != nil {
		return err
	}

	document.Value = &postgres.Jsonb{RawMessage: json.RawMessage(enValue)}
	status := pg.CreateSalesforceDocument(projectID, &document)
	if status != http.StatusCreated && status != http.StatusConflict {
		return fmt.Errorf("error while creating document Status %d", status)
	}

	if status == http.StatusConflict {
		log.WithFields(log.Fields{"project_id": projectID, "object_name": objectName}).Info("Skipped inserting salesforce record.")
	} else {
		log.WithFields(log.Fields{"project_id": projectID, "object_name": objectName}).Info("Successfully inserted salesforce record.")
	}

	return nil
}

// GetSalesforceDocumentsByTypeForSync - Pulls salesforce documents which are not synced
func (pg *Postgres) GetSalesforceDocumentsByTypeForSync(projectID uint64, typ int, from, to int64) ([]model.SalesforceDocument, int) {
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "type": typ})

	if projectID == 0 || typ == 0 {
		logCtx.Error("Invalid project_id or type on get salesforce documents by type.")
		return nil, http.StatusBadRequest
	}

	whereStmnt := "project_id=? AND type=? AND synced=false"
	whereParams := []interface{}{projectID, typ}
	if from > 0 && to > 0 {
		whereStmnt = whereStmnt + " AND " + "timestamp BETWEEN ? AND ?"
		whereParams = append(whereParams, from, to)
	}

	var documents []model.SalesforceDocument
	db := C.GetServices().Db
	err := db.Order("timestamp, created_at ASC").Where(whereStmnt, whereParams...).Find(&documents).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to get salesforce documents by type.")
		return nil, http.StatusInternalServerError
	}

	if len(documents) < 1 {
		return nil, http.StatusNotFound
	}
	return documents, http.StatusFound
}

// GetSalesforceDocumentBeginingTimestampByDocumentTypeForSync returns the minimum timestamp for unsynced document
func (pg *Postgres) GetSalesforceDocumentBeginingTimestampByDocumentTypeForSync(projectID uint64) (map[int]int64, int64, int) {
	logCtx := log.WithFields(log.Fields{"project_id": projectID})

	if projectID == 0 {
		logCtx.Error("Invalid project_id.")
		return nil, 0, http.StatusBadRequest
	}

	db := C.GetServices().Db
	rows, err := db.Raw("SELECT type, MIN(timestamp) FROM salesforce_documents WHERE project_id=? AND synced=false GROUP BY type", projectID).Rows()
	if err != nil {
		log.WithError(err).Error("Failed to get salesforce minimum timestamp.")
		return nil, 0, http.StatusInternalServerError
	}

	var docMinTimestamp map[int]int64
	var overallMinTimestamp int64

	defer rows.Close()
	for rows.Next() {
		var minTimestamp *int64
		var docType *int
		if err := rows.Scan(&docType, &minTimestamp); err != nil {
			log.WithError(err).Error("Failed scanning rows on get salesforce minimum timestamp for sync.")
			continue
		}

		if docMinTimestamp == nil {
			docMinTimestamp = make(map[int]int64)
		}

		if overallMinTimestamp == 0 || *minTimestamp < overallMinTimestamp {
			overallMinTimestamp = *minTimestamp
		}

		docMinTimestamp[*docType] = *minTimestamp
	}

	if docMinTimestamp == nil {
		return nil, 0, http.StatusNotFound
	}

	return docMinTimestamp, overallMinTimestamp, http.StatusFound
}

// batch not supported in Postgres
func (pg *Postgres) BuildAndUpsertDocumentInBatch(projectID uint64, objectName string, values []model.SalesforceRecord) error {
	return errors.New("method not allowed")
}
