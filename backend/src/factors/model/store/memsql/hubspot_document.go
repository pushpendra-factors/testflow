package memsql

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
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

	logCtx := log.WithFields(log.Fields{"document": document})

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
		if _, ok := (*documentMap)["id"]; ok {
			idKey = "id"
		}
	case model.HubspotDocumentTypeContact:
		idKey = "vid"
		if document.Action == model.HubspotDocumentActionDeleted {
			idKey = "id"
		}
	case model.HubspotDocumentTypeDeal:
		idKey = "dealId"
		if _, ok := (*documentMap)["id"]; ok {
			idKey = "id"
		}
	case model.HubspotDocumentTypeFormSubmission:
		idKey = "formId"
	case model.HubspotDocumentTypeEngagement:
		idKey = "id"
	case model.HubspotDocumentTypeContactList:
		idKey = "listId"
	case model.HubspotDocumentTypeOwner:
		idKey = "ownerId"
	default:
		idKey = "guid"
	}

	if idKey == "" {
		return "", errors.New("invalid hubspot document key")
	}

	if document.Type == model.HubspotDocumentTypeEngagement {
		return model.GetHubspotEngagementId(*documentMap, idKey)
	}

	id, idExists := (*documentMap)[idKey]
	if !idExists {
		return "", errors.New("id key not exist on hubspot document")
	}

	idAsString := U.GetPropertyValueAsString(id)
	if idAsString == "" {
		return "", errors.New("invalid id on hubspot document")
	}

	if document.Type == model.HubspotDocumentTypeContactList {
		contactId, contactIdExists := (*documentMap)["contact_id"]
		if !contactIdExists {
			return "", errors.New("contact_id not found on contact_list document type")
		}

		contactIdAsString := U.GetPropertyValueAsString(contactId)
		idAsString = fmt.Sprintf("%s:%s", idAsString, contactIdAsString)
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

func isExistHubspotDocumentByIDAndType(projectId int64, id string, docType int) int {
	argFields := log.Fields{"project_id": projectId, "id": id, "type": docType}
	logCtx := log.WithFields(argFields)

	if projectId == 0 || id == "" || docType == 0 {
		logCtx.Error("Failed to get hubspot document by id and type. Invalid project_id or id or type.")
		return http.StatusBadRequest
	}

	documentIds, status := isExistHubspotDocumentByIDAndTypeInBatch(projectId, []string{id}, docType)
	if status != http.StatusFound {
		return status
	}

	if !documentIds[id] {
		return http.StatusNotFound
	}

	return http.StatusFound
}

func isExistHubspotDocumentByIDAndTypeInBatch(projectId int64, ids []string, docType int) (map[string]bool, int) {
	argFields := log.Fields{"project_id": projectId, "ids": ids, "type": docType}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &argFields)

	logCtx := log.WithFields(argFields)

	if projectId == 0 || len(ids) <= 0 || docType == 0 {
		logCtx.Error("Failed to get hubspot document by id and type. Invalid project_id or id or type.")
		return nil, http.StatusBadRequest
	}

	whereStmnt := "project_id = ? AND type = ? AND action = ?"
	whereParams := []interface{}{projectId, docType, model.HubspotDocumentActionCreated}
	db := C.GetServices().Db
	if len(ids) > 1 {
		whereStmnt = whereStmnt + " AND " + "id IN(?) "
		whereParams = append(whereParams, ids)
	} else {
		whereStmnt = whereStmnt + " AND " + "id = ?"
		whereParams = append(whereParams, ids[0])
		db.Limit(1)
	}

	var documents []model.HubspotDocument
	err := db.Where(whereStmnt, whereParams...).Select("id").Find(&documents).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}

		logCtx.WithError(err).Error("Failed to get hubspot documents.")
		return nil, http.StatusInternalServerError
	}

	if len(documents) <= 0 {
		return nil, http.StatusNotFound
	}

	documentIds := make(map[string]bool, 0)
	for i := range documents {
		documentIds[documents[i].ID] = true
	}

	return documentIds, http.StatusFound
}

func (store *MemSQL) GetHubspotContactCreatedSyncIDAndUserID(projectID int64, docID string) ([]model.HubspotDocument, int) {
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

func (store *MemSQL) GetHubspotDocumentByTypeAndActions(projectId int64, ids []string,
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
func (store *MemSQL) GetSyncedHubspotDocumentByFilter(projectID int64,
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

func (store *MemSQL) getUpdatedDealAssociationDocument(projectID int64, incomingDocument *model.HubspotDocument) (*model.HubspotDocument, int) {
	logFields := log.Fields{
		"project_id":        projectID,
		"incoming_document": incomingDocument,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "incoming_document": incomingDocument})
	if projectID <= 0 || incomingDocument.Type != model.HubspotDocumentTypeDeal || incomingDocument.ID == "" {
		logCtx.Error("Invalid record on getUpdatedDealAssociationDocument.")
		return nil, http.StatusBadRequest
	}

	existingDocuments, status := store.GetHubspotDocumentByTypeAndActions(projectID, []string{incomingDocument.ID}, model.HubspotDocumentTypeDeal,
		[]int{incomingDocument.Action, model.HubspotDocumentActionAssociationsUpdated})
	if status != http.StatusNotFound && status != http.StatusFound {

		return nil, http.StatusInternalServerError
	}

	latestDocument := existingDocuments[len(existingDocuments)-1]
	updateRequired, err := model.IsDealUpdatedRequired(incomingDocument, &latestDocument)
	if err != nil {
		log.WithFields(log.Fields{"incoming_document": incomingDocument, "latest_document": latestDocument}).
			WithError(err).Error("Failed to check for IsDealUpdatedRequired.")
		return nil, http.StatusInternalServerError
	}

	if !updateRequired {
		return nil, http.StatusConflict
	}

	incomingDocument.Timestamp = latestDocument.Timestamp + 1
	incomingDocument.Action = model.HubspotDocumentActionAssociationsUpdated

	errCode := store.satisfiesHubspotDocumentUniquenessConstraints(incomingDocument)
	if errCode != http.StatusOK {
		if errCode == http.StatusConflict {
			return nil, errCode
		}

		logCtx.WithField("err_code", errCode).Error("Failed to check satisfiesHubspotDocumentUniquenessConstraints.")
		return nil, status
	}

	return incomingDocument, http.StatusOK
}

func (store *MemSQL) getExistingDocumentsForDealAssociationUpdates(projectID int64, documentIDs []string) ([]model.HubspotDocument, int) {
	logFields := log.Fields{
		"project_id":   projectID,
		"document_ids": documentIDs,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	if projectID == 0 || len(documentIDs) == 0 {
		logCtx.Error("Invalid input parameters on getExistingDocumentsForDealAssociationUpdates")
		return nil, http.StatusBadRequest
	}

	incomingDocumentActions := []int{model.HubspotDocumentActionUpdated, model.HubspotDocumentActionAssociationsUpdated}

	existingDocuments, status := store.GetHubspotDocumentByTypeAndActions(projectID, documentIDs, model.HubspotDocumentTypeDeal,
		incomingDocumentActions)
	if status != http.StatusNotFound && status != http.StatusFound {
		return nil, http.StatusInternalServerError
	}

	return existingDocuments, http.StatusOK
}

func (store *MemSQL) getUpdatedDealAssociationDocuments(projectID int64, incomingDocuments []*model.HubspotDocument, existingDocuments []model.HubspotDocument) ([]*model.HubspotDocument, int) {
	logFields := log.Fields{
		"project_id":         projectID,
		"incoming_documents": incomingDocuments,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "incoming_documents": incomingDocuments})
	if projectID <= 0 {
		logCtx.Error("Invalid projectID on getUpdatedDealAssociationDocuments.")
		return nil, http.StatusBadRequest
	}

	if len(existingDocuments) == 0 {
		return nil, http.StatusOK
	}

	var modifiedDocuments []*model.HubspotDocument
	for i := range incomingDocuments {
		if incomingDocuments[i].Type != model.HubspotDocumentTypeDeal || incomingDocuments[i].Action == model.HubspotDocumentActionCreated {
			continue
		}

		documentID := incomingDocuments[i].ID
		var latestDocument *model.HubspotDocument
		for j := range existingDocuments {
			if existingDocuments[j].ID == documentID {
				if latestDocument == nil {
					latestDocument = &existingDocuments[j]
				} else if existingDocuments[j].Timestamp > latestDocument.Timestamp {
					latestDocument = &existingDocuments[j]
				}
			}
		}

		if latestDocument == nil || incomingDocuments[i].Timestamp > latestDocument.Timestamp {
			continue
		}

		updateRequired, err := model.IsDealUpdatedRequired(incomingDocuments[i], latestDocument)
		if err != nil {
			log.WithFields(log.Fields{"incoming_document": incomingDocuments[i], "latest_document": latestDocument}).
				WithError(err).Error("Failed to check for IsDealUpdatedRequired.")
			return nil, http.StatusInternalServerError
		}

		if !updateRequired {
			continue
		}

		modifiedDocument := *incomingDocuments[i]
		modifiedDocument.Timestamp = latestDocument.Timestamp + 1
		modifiedDocument.Action = model.HubspotDocumentActionAssociationsUpdated
		modifiedDocuments = append(modifiedDocuments, &modifiedDocument)
	}

	return modifiedDocuments, http.StatusOK
}

func (store *MemSQL) createBatchedHubspotDocuments(projectID int64, documents []*model.HubspotDocument) int {
	defer model.LogOnSlowExecutionWithParams(time.Now(), &log.Fields{"project_id": projectID})

	logCtx := log.WithFields(log.Fields{"project_id": projectID, "documents": len(documents)})
	if len(documents) <= 0 {
		logCtx.Error("Empty batch for hubspot batch insert.")
		return http.StatusBadRequest
	}
	log.Info("Using hubspot batch insert.")

	batchedArguments := make([]interface{}, 0)
	insertColumns := "INSERT INTO hubspot_documents(project_id, id, type, action, timestamp, value, created_at, updated_at)"
	placeHolders := ""
	createdTime := gorm.NowFunc()
	for i := range documents {
		documents[i].ProjectId = projectID
		if placeHolders != "" {
			placeHolders = placeHolders + ","
		}
		placeHolders = placeHolders + "( ? )"
		createdTime = createdTime.Add(1 * time.Microsecond) // db precision is in microsecond
		arguments := []interface{}{
			documents[i].ProjectId,
			documents[i].ID,
			documents[i].Type,
			documents[i].Action,
			documents[i].Timestamp,
			documents[i].Value,
			createdTime,
			createdTime,
		}
		batchedArguments = append(batchedArguments, arguments)
	}
	insertStmnt := insertColumns + " VALUES " + placeHolders + " ON DUPLICATE KEY UPDATE synced=synced;"

	db := C.GetServices().Db
	err := db.Exec(insertStmnt, batchedArguments...).Error
	if err != nil {
		log.WithError(err).Error("Failed to batch insert hubspot documents.")
		return http.StatusInternalServerError
	}

	return http.StatusCreated
}

func (store *MemSQL) modifyAndCreateBatchedHubspotDocuments(projectID int64, documents []*model.HubspotDocument) int {
	defer model.LogOnSlowExecutionWithParams(time.Now(), &log.Fields{"project_id": projectID})

	logCtx := log.WithFields(log.Fields{"project_id": projectID, "documents": len(documents)})
	if len(documents) <= 0 {
		logCtx.Error("Empty batch for hubspot batch insert.")
		return http.StatusBadRequest
	}
	log.Info("Modifying hubspot batch size while insertion.")

	memsqlMaxSize := float64(C.GetDBMaxAllowedPacket())
	maxByteSize := float64(0)
	noOfColumns := float64(8)

	for i := range documents {
		value, err := documents[i].Value.RawMessage.MarshalJSON()
		if err != nil {
			logCtx.WithError(err).WithField("document", documents[i]).Error(err)
			continue
		}
		maxByteSize = math.Max(maxByteSize, float64(len(value)))
	}

	modifiedBatchSize := int(math.Min(float64(len(documents)), (memsqlMaxSize / maxByteSize / noOfColumns)))

	batchedDocuments := model.GetHubspotDocumentsListAsBatchById(documents, modifiedBatchSize)
	for i := range batchedDocuments {
		if len(batchedDocuments[i]) == 0 {
			continue
		}

		status := store.createBatchedHubspotDocuments(projectID, batchedDocuments[i])
		if status != http.StatusCreated {
			logCtx.WithFields(log.Fields{
				"DBMaxAllowedPacket":  memsqlMaxSize,
				"maxByteSize":         maxByteSize,
				"original_batch_size": len(documents),
				"modified_batch_size": modifiedBatchSize,
				"documents":           len(batchedDocuments),
				"err_code":            status,
			}).Error("Failed to insert hubspot documents after modifying batchsize.")
			return status
		}
	}

	return http.StatusCreated
}

func (store *MemSQL) getHubspotDocumentsForInsertion(projectId int64, documents []*model.HubspotDocument, existDocumentIDs map[string]bool, documentType int, existDocuments []model.HubspotDocument) ([]*model.HubspotDocument, error) {
	processDocuments := make([]*model.HubspotDocument, 0)
	batchDocumentIDs := make(map[string]bool, 0)

	for i := range documents {
		if exist := batchDocumentIDs[documents[i].ID]; exist {
			log.WithFields(log.Fields{"project_id": documents[i].ProjectId,
				"document_id": documents[i].ID, "documents": documents}).
				Warn("Duplicate hubspot document in same batch.")
		}
		batchDocumentIDs[documents[i].ID] = true

		isNew := !existDocumentIDs[documents[i].ID]
		createdTimestamp, updatedTimestamp, err := getHubspotCreatedAndUpdatedTimestamp(documents[i])
		if err != nil {
			return nil, err
		}
		if isNew {
			// Skip adding the record if deleted record is to added for
			// non-existing document.
			if documents[i].Action == model.HubspotDocumentActionDeleted {
				continue
			}
			createdDocument := documents[i]
			createdDocument.Action = model.HubspotDocumentActionCreated // created
			createdDocument.Timestamp = createdTimestamp
			processDocuments = append(processDocuments, createdDocument)

			// for create action also create updated with same timestamp
			updatedDocument := *documents[i]
			updatedDocument.Action = model.HubspotDocumentActionUpdated
			updatedDocument.Timestamp = createdTimestamp
			processDocuments = append(processDocuments, &updatedDocument)

			if updatedTimestamp > createdTimestamp { // create action updated if last modified time is greater than created
				recentUpdatedDocument := *documents[i]
				recentUpdatedDocument.Action = model.HubspotDocumentActionUpdated
				recentUpdatedDocument.Timestamp = updatedTimestamp
				processDocuments = append(processDocuments, &recentUpdatedDocument)
			}

		} else {
			if documents[i].Action != model.HubspotDocumentActionDeleted {
				documents[i].Action = model.HubspotDocumentActionUpdated // updated
			}
			// Any update on the entity would create a new hubspot document.
			// i.e, deal will be synced after updating a created deal with a
			// contact or a company.
			documents[i].Timestamp = updatedTimestamp
			processDocuments = append(processDocuments, documents[i])
		}
	}

	if documentType == model.HubspotDocumentTypeDeal {
		associationDocuments, errCode := store.getUpdatedDealAssociationDocuments(projectId, processDocuments, existDocuments)
		if errCode != http.StatusOK {
			return nil, errors.New("failed to getUpdatedDealAssociationDocuments")
		}
		processDocuments = append(processDocuments, associationDocuments...)
	}

	return processDocuments, nil
}

func allowedHubspotDocTypeForBatchInsert(docType int) bool {
	return docType == model.HubspotDocumentTypeContact || docType == model.HubspotDocumentTypeCompany ||
		docType == model.HubspotDocumentTypeEngagement || docType == model.HubspotDocumentTypeForm ||
		docType == model.HubspotDocumentTypeFormSubmission || docType == model.HubspotDocumentTypeDeal ||
		docType == model.HubspotDocumentTypeContactList || docType == model.HubspotDocumentTypeOwner
}

func (store *MemSQL) CreateHubspotDocumentInBatch(projectID int64, docType int, documents []*model.HubspotDocument, batchSize int) int {
	logFields := log.Fields{"project_id": projectID, "doc_type": docType, "batch_size": batchSize}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)
	if projectID == 0 || docType == 0 || batchSize <= 0 {
		logCtx.Error("Invalid parameters on create hubspot document in batch.")
		return http.StatusBadRequest
	}

	if len(documents) <= 0 {
		logCtx.Error("Missing documents.")
		return http.StatusBadRequest
	}

	if !allowedHubspotDocTypeForBatchInsert(docType) {
		logCtx.WithField("doc_type", docType).Error("Invalid document type.")
		return http.StatusBadRequest
	}

	for i := range documents {
		documents[i].ProjectId = projectID

		documents[i].Type = docType

		if U.IsEmptyPostgresJsonb(documents[i].Value) {
			logCtx.Error("Empty document value on create batch hubspot document. Skipped adding this record.")
		}

		documentId, err := getHubspotDocumentId(documents[i])
		if err != nil {
			logCtx.WithFields(log.Fields{"document": documents[i]}).WithError(err).Error(
				"Failed to get id for hubspot document on create.")
			return http.StatusInternalServerError
		}
		documents[i].ID = documentId
	}

	documentIDs := make([]string, 0)
	for i := range documents {
		documentIDs = append(documentIDs, documents[i].ID)
	}

	existDocumentIDs, errCode := isExistHubspotDocumentByIDAndTypeInBatch(projectID,
		documentIDs, docType)
	if errCode == http.StatusInternalServerError || errCode == http.StatusBadRequest {
		logCtx.WithField("err_code", errCode).Error("Failed to get isExistHubspotDocumentByIDAndTypeInBatch")
		return errCode
	}

	var existDocuments []model.HubspotDocument
	if docType == model.HubspotDocumentTypeDeal {
		existDocuments, errCode = store.getExistingDocumentsForDealAssociationUpdates(projectID, documentIDs)
		if errCode != http.StatusOK {
			logCtx.WithField("err_code", errCode).Error("Failed to get getExistingDocumentsForDealAssociationUpdates")
			return errCode
		}
	}

	batchedDocuments := model.GetHubspotDocumentsListAsBatch(documents, batchSize)
	for i := range batchedDocuments {
		processDocuments, err := store.getHubspotDocumentsForInsertion(projectID, batchedDocuments[i], existDocumentIDs, docType, existDocuments)
		if err != nil {
			logCtx.WithFields(log.Fields{"documents": processDocuments}).WithError(err).
				Error("Failed to get documents for processing in batch.")
			return http.StatusInternalServerError
		}

		if len(processDocuments) == 0 {
			logCtx.WithFields(log.Fields{"documents": batchedDocuments[i]}).Error("No document for processing in batch.")
			continue
		}

		status := store.modifyAndCreateBatchedHubspotDocuments(projectID, processDocuments)
		if status != http.StatusCreated {
			logCtx.WithFields(log.Fields{"documents": processDocuments, "err_code": status}).
				WithError(err).Error("Failed to insert batched hubspot documents.")
			return status
		}
	}

	// update count cache for batch of documents
	currentTime := U.TimeNowZ()
	for range documents {
		UpdateCountCacheByDocumentType(projectID, &currentTime, "hubspot")
	}

	return http.StatusCreated
}

func getHubspotCreatedAndUpdatedTimestamp(document *model.HubspotDocument) (int64, int64, error) {
	logCtx := log.WithFields(log.Fields{"project_id": document.ProjectId, "document_id": document.ID, "doc_type": document.Type})
	createdTimestamp, err := model.GetHubspotDocumentCreatedTimestamp(document)
	if err != nil {
		if err != model.ErrorHubspotUsingFallbackKey {
			logCtx.WithField("action", document.Action).WithError(err).Error(
				"Failed to get created timestamp from hubspot document on create.")
			return 0, 0, err
		}

		logCtx.WithField("action", document.Action).WithError(err).Error("Missing document key.")
	}

	updatedTimestamp, err := model.GetHubspotDocumentUpdatedTimestamp(document)
	if err != nil {
		if err != model.ErrorHubspotUsingFallbackKey {
			logCtx.WithField("action", document.Action).WithError(err).Error(
				"Failed to get updated timestamp from hubspot document on create.")
			return 0, 0, err
		}

		logCtx.WithField("action", document.Action).WithError(err).Error("Missing document key.")
	}

	return createdTimestamp, updatedTimestamp, nil
}
func (store *MemSQL) CreateHubspotDocument(projectId int64, document *model.HubspotDocument) int {

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

	createdTimestamp, updatedTimestamp, err := getHubspotCreatedAndUpdatedTimestamp(document)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get hubspot document created and updated timestamp.")
		return http.StatusInternalServerError
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

	if !C.DisableCRMUniquenessConstraintsCheckByProjectID(projectId) {
		errCode = store.satisfiesHubspotDocumentUniquenessConstraints(document)
		if errCode != http.StatusOK {
			if errCode != http.StatusConflict {
				return errCode
			}

			if document.Type != model.HubspotDocumentTypeDeal {
				return errCode
			}

			newDocument, errCode := store.getUpdatedDealAssociationDocument(projectId, document)
			if errCode != http.StatusOK {
				if errCode != http.StatusConflict {
					logCtx.WithField("errCode", errCode).Error("Failed to getUpdatedDealAssociationDocument.")
					return http.StatusInternalServerError
				}

				return errCode
			}

			document = newDocument
		}
	}

	db := C.GetServices().Db
	err = db.Create(document).Error
	if err != nil {
		if !IsDuplicateRecordError(err) {
			logCtx.WithError(err).Error("Failed to create hubspot document.")
			return http.StatusInternalServerError
		}

		if !C.DisableCRMUniquenessConstraintsCheckByProjectID(projectId) {
			return http.StatusConflict
		}

		if document.Type != model.HubspotDocumentTypeDeal {
			return http.StatusConflict
		}

		newDocument, errCode := store.getUpdatedDealAssociationDocument(projectId, document)
		if errCode != http.StatusOK {
			if errCode != http.StatusConflict {
				logCtx.WithField("errCode", errCode).Error("Failed to getUpdatedDealAssociationDocument.")
				return http.StatusInternalServerError
			}
			return errCode
		}

		err = db.Create(&newDocument).Error
		if err != nil {
			if IsDuplicateRecordError(err) {
				return http.StatusConflict
			}

			logCtx.WithError(err).Error("Failed to create hubspot deal association document.")
			return http.StatusInternalServerError
		}
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
	logFields := log.Fields{
		"t": t,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	for alias, typ := range model.HubspotDocumentTypeAlias {
		if t == typ {
			return alias
		}
	}

	return ""
}

func (store *MemSQL) updateHubspotProjectSettingsLastSyncInfo(projectID int64, incomingSyncInfo map[string]int64) error {
	defer model.LogOnSlowExecutionWithParams(time.Now(), &log.Fields{"project_id": projectID})

	logCtx := log.WithFields(log.Fields{"project_id": projectID})

	if projectID == 0 || incomingSyncInfo == nil {
		logCtx.Error("Missing required fields.")
		return errors.New("missing required fields")
	}

	projectSetting, status := store.GetProjectSetting(projectID)
	if status != http.StatusFound {
		logCtx.WithField("err_code", status).Error("Failed to get project setttings on hubspot last sync info.")
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
		logCtx.WithField("err_code", status).Error("Failed to update hubspot last sync info on success.")
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
					log.WithFields(log.Fields{"project_id": pid, "err_code": status}).
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
	enabledProjectLastSync := make(map[int64]map[string]int64, 0)

	// get project settings of hubspot enabled projects.
	projectSettings, errCode := store.GetAllHubspotProjectSettings()
	if errCode != http.StatusFound {
		return nil, http.StatusInternalServerError
	}

	settingsByProject := make(map[int64]*model.HubspotProjectSettings, 0)
	for i, ps := range projectSettings {
		if ps.IsFirstTimeSynced {
			continue
		}

		// add types not synced before.
		for typ := range model.HubspotDocumentTypeAlias {
			if !C.AllowHubspotEngagementsByProjectID(ps.ProjectId) && typ == model.HubspotDocumentTypeNameEngagement {
				continue
			}

			if !C.AllowSyncReferenceFields(ps.ProjectId) && typ == model.HubspotDocumentTypeNameOwner {
				continue
			}

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

	lastSyncInfoByProject := make(map[int64]map[string]int64, 0)
	for _, syncInfo := range lastSyncInfo {
		if syncInfo.Type == model.HubspotDocumentTypeContactList {
			continue
		}

		if _, projectExists := lastSyncInfoByProject[syncInfo.ProjectID]; !projectExists {
			lastSyncInfoByProject[syncInfo.ProjectID] = make(map[string]int64)
		}

		lastSyncInfoByProject[syncInfo.ProjectID][getHubspotTypeAlias(syncInfo.Type)] = syncInfo.Timestamp
	}

	// project sync of hubspot enable projects.
	enabledProjectLastSync := make(map[int64]map[string]int64, 0)

	// get project settings of hubspot enaled projects.
	projectSettings, errCode := store.GetAllHubspotProjectSettings()
	if errCode != http.StatusFound {
		return nil, http.StatusInternalServerError
	}

	settingsByProject := make(map[int64]*model.HubspotProjectSettings, 0)
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
			if !C.AllowHubspotEngagementsByProjectID(ps.ProjectId) && typ == model.HubspotDocumentTypeNameEngagement {
				continue
			}

			if !C.AllowSyncReferenceFields(ps.ProjectId) && typ == model.HubspotDocumentTypeNameOwner {
				continue
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

func (store *MemSQL) GetHubspotFormDocuments(projectId int64) ([]model.HubspotDocument, int) {
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

func (store *MemSQL) GetHubspotDocumentsByTypeForSync(projectId int64, typ int, maxCreatedAtSec int64) ([]model.HubspotDocument, int) {
	argFields := log.Fields{"project_id": projectId, "type": typ}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &log.Fields{"project_id": projectId, "typ": typ, "max_created_at_sec": maxCreatedAtSec})

	logCtx := log.WithFields(argFields)

	if projectId == 0 || typ == 0 || maxCreatedAtSec <= 0 {
		logCtx.Error("Invalid project_id or type or maxCreatedAtSec  on get hubspot documents by type.")
		return nil, http.StatusBadRequest
	}

	maxCreatedAtFmt := time.Unix(maxCreatedAtSec+1, 0).Format("2006-01-02 15:04:05")
	var documents []model.HubspotDocument

	db := C.GetServices().Db
	err := db.Order("timestamp, created_at ASC").Where("project_id=? AND type=? AND synced=false AND created_at < ? ",
		projectId, typ, maxCreatedAtFmt).Find(&documents).Error
	if err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			logCtx.WithError(err).Error("Failed to get hubspot documents by type.")
			return nil, http.StatusInternalServerError
		}
		return nil, http.StatusNotFound
	}

	if len(documents) == 0 {
		return nil, http.StatusNotFound
	}

	return documents, http.StatusFound
}

// GetHubspotDocumentBeginingTimestampByDocumentTypeForSync returns the minimum timestamp for unsynced document
func (store *MemSQL) GetHubspotDocumentBeginingTimestampByDocumentTypeForSync(projectID int64, docTypes []int) (int64, int) {
	defer model.LogOnSlowExecutionWithParams(time.Now(), &log.Fields{"project_id": projectID})

	logCtx := log.WithFields(log.Fields{"project_id": projectID, "doc_types": docTypes})

	if projectID == 0 || len(docTypes) < 1 {
		logCtx.Error("Invalid parameters.")
		return 0, http.StatusBadRequest
	}

	db := C.GetServices().Db
	rows, err := db.Raw("SELECT MIN(timestamp) FROM hubspot_documents WHERE project_id=? AND synced=false and type IN ( ? )", projectID, docTypes).Rows()
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

// GetMinTimestampByFirstSync() returns the minimum timestamp of first sync in hubspot documents
func (store *MemSQL) GetMinTimestampByFirstSync(projectID int64, docType int) (int64, int) {
	defer model.LogOnSlowExecutionWithParams(time.Now(), &log.Fields{"project_id": projectID})

	logCtx := log.WithFields(log.Fields{"project_id": projectID, "doc_type": docType})

	if projectID == 0 || docType == 0 {
		logCtx.Error("Invalid parameters.")
		return 0, http.StatusBadRequest
	}

	db := C.GetServices().Db
	rows, err := db.Raw("SELECT MIN(updated_documents.timestamp) as timestamp"+
		" "+"FROM hubspot_documents as created_documents"+
		" "+"LEFT JOIN hubspot_documents as updated_documents ON created_documents.id = updated_documents.id"+
		" "+"and created_documents.project_id = ? and updated_documents.project_id = ?"+
		" "+"and created_documents.type = ? and updated_documents.type = ?"+
		" "+"and created_documents.action = 1 and updated_documents.action = 2"+
		" "+"WHERE created_documents.timestamp != updated_documents.timestamp AND created_documents.synced=true", projectID, projectID, docType, docType).Rows() // only updated documents, ignoring create document=update document for duplicates
	if err != nil {
		log.WithError(err).Error("Failed to get hubspot minimum timestamp for updated documents.")
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
		log.Info("Failed to get hubspot minimum timestamp for sync.")
		return time.Now().AddDate(0, 0, -1).UnixNano() / int64(time.Millisecond), http.StatusNotFound
	}

	if *minTimestamp > 0 {
		return *minTimestamp, http.StatusFound
	}

	return time.Now().AddDate(0, 0, -1).UnixNano() / int64(time.Millisecond), http.StatusNotFound
}

// GetHubspotDocumentCountForSync returns count for records for each project
func (store *MemSQL) GetHubspotDocumentCountForSync(projectIDs []int64, docTypes []int, maxCreatedAtSec int64) ([]model.HubspotDocumentCount, int) {
	logFields := log.Fields{"project_ids": projectIDs, "doc_types": docTypes, "max_created_at": maxCreatedAtSec}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)

	if len(projectIDs) == 0 {
		logCtx.Error("Invalid parameters.")
		return nil, http.StatusBadRequest
	}

	var projectRecordCount []model.HubspotDocumentCount

	maxCreatedAtFmt := time.Unix(maxCreatedAtSec+1, 0).Format("2006-01-02 15:04:05")
	db := C.GetServices().Db
	err := db.Model(model.HubspotDocument{}).Select("project_id, count(*) as count").
		Where("project_id IN (?) AND synced=false and type IN ( ? ) AND created_at <= ? ", projectIDs, docTypes, maxCreatedAtFmt).
		Group("project_id").Scan(&projectRecordCount).Error
	if err != nil {
		log.WithError(err).Error("Failed to get hubspot minimum timestamp.")
		return nil, http.StatusInternalServerError
	}

	if len(projectRecordCount) == 0 {
		return nil, http.StatusNotFound
	}

	return projectRecordCount, http.StatusFound
}

// GetHubspotDocumentsByTypeANDRangeForSync return list of documents unsynced for given time range
func (store *MemSQL) GetHubspotDocumentsByTypeANDRangeForSync(projectID int64,
	docType int, from, to, maxCreatedAtSec int64) ([]model.HubspotDocument, int) {

	argFields := log.Fields{"project_id": projectID, "type": docType, "from": from, "to": to, "max_created_at_sec": maxCreatedAtSec}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &argFields)
	logCtx := log.WithFields(argFields)

	if projectID == 0 || docType == 0 || from < 0 || to < 0 {
		logCtx.Error("Invalid project_id or type on get hubspot documents by type.")
		return nil, http.StatusBadRequest
	}

	maxCreatedAtFmt := time.Unix(maxCreatedAtSec+1, 0).Format("2006-01-02 15:04:05")
	var documents []model.HubspotDocument

	db := C.GetServices().Db
	err := db.Order("timestamp, created_at ASC").Where("project_id=? AND type=? AND synced=false AND timestamp BETWEEN ? AND ? AND created_at < ? ",
		projectID, docType, from, to, maxCreatedAtFmt).Find(&documents).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to get hubspot documents by type.")
		return nil, http.StatusInternalServerError
	}

	return documents, http.StatusFound
}

func (store *MemSQL) GetSyncedHubspotDealDocumentByIdAndStage(projectId int64, id string,
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

func getHubspotDocumentValuesByPropertyNameAndLimitForV3Records(hubspotDocument model.HubspotDocument,
	propertyName string, valuesAggregate map[interface{}]int) (map[interface{}]int, error) {
	if propertyName == "" {
		return valuesAggregate, errors.New("empty property name on getHubspotDocumentValuesByPropertyNameAndLimitForV3Records")
	}

	var docProperties model.HubspotDocumentPropertiesV3
	err := json.Unmarshal((hubspotDocument.Value).RawMessage, &docProperties)
	if err != nil {
		return valuesAggregate, err
	}

	for name, value := range docProperties.Properties {
		if name != propertyName {
			continue
		}

		if value == nil {
			continue
		}

		valueStr := U.GetPropertyValueAsString(value)
		valuesAggregate[valueStr] = valuesAggregate[valueStr] + 1
	}

	return valuesAggregate, nil
}

func getHubspotDocumentValuesByPropertyNameAndLimit(hubspotDocuments []model.HubspotDocument,
	propertyName string, limit int) []interface{} {
	logFields := log.Fields{
		"hubspot_documents": hubspotDocuments,
		"property_name":     propertyName,
		"limit":             limit,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	if len(hubspotDocuments) < 1 || propertyName == "" {
		return nil
	}

	valuesAggregate := make(map[interface{}]int)
	for i := range hubspotDocuments {
		var err error
		valuesAggregate, err = getHubspotDocumentValuesByPropertyNameAndLimitForV3Records(hubspotDocuments[i], propertyName, valuesAggregate)
		if err == nil {
			continue
		}

		var docProperties model.HubspotDocumentProperties
		err = json.Unmarshal((hubspotDocuments[i].Value).RawMessage, &docProperties)
		if err != nil {
			log.WithFields(log.Fields{"document_id": hubspotDocuments[i].ID, "property_name": propertyName}).WithError(err).
				Error("Failed to unmarshal hubspot document on getHubspotDocumentValuesByPropertyNameAndLimit")
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

func getHubspotDocumentPropertiesNameByTypeForV3Records(hubspotDocument model.HubspotDocument, dateTimeProperties, categoricalProperties map[string]interface{}, currentTimestamp int64, logFields log.Fields) (map[string]interface{}, map[string]interface{}, error) {
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	var docProperties model.HubspotDocumentPropertiesV3
	err := json.Unmarshal((hubspotDocument.Value).RawMessage, &docProperties)
	if err != nil {
		log.WithError(err).Error("Failed to unmarshal hubspot document on getHubspotDocumentPropertiesNameByTypeForV3Records")
		return dateTimeProperties, categoricalProperties, err
	}

	for key, value := range docProperties.Properties {
		valueStr := U.GetPropertyValueAsString(value)
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

		if len(valueStr) == 20 || len(valueStr) == 24 { // datetime format - for v3 records
			timestamp, err := model.GetTimestampForV3Records(valueStr)
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

	return dateTimeProperties, categoricalProperties, nil
}

func getHubspotDocumentPropertiesNameByType(hubspotDocuments []model.HubspotDocument) ([]string, []string) {
	logFields := log.Fields{
		"hubspot_documents": hubspotDocuments,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	dateTimeProperties := make(map[string]interface{})
	categoricalProperties := make(map[string]interface{})
	currentTimestamp := U.TimeNowUnix() * 1000

	for i := range hubspotDocuments {
		var err error
		dateTimeProperties, categoricalProperties, err = getHubspotDocumentPropertiesNameByTypeForV3Records(hubspotDocuments[i], dateTimeProperties, categoricalProperties, currentTimestamp, logFields)
		if err == nil {
			continue
		}

		var docProperties model.HubspotDocumentProperties
		err = json.Unmarshal((hubspotDocuments[i].Value).RawMessage, &docProperties)
		if err != nil {
			log.WithError(err).Error("Failed to unmarshal hubspot document on getHubspotDocumentPropertiesNameByType")
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

func getLatestHubspotDocumentsByLimit(projectID int64, docType int, limit int) ([]model.HubspotDocument, error) {
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
		projectID, docType, model.HubspotDocumentActionUpdated, lookbackTimestampInMilliseconds).Order("timestamp desc").Limit(limit).Find(&hubspotDocuments).Error
	if err != nil {
		return nil, err

	}

	return hubspotDocuments, nil
}

// GetHubspotObjectPropertiesName returns property names by type
func (store *MemSQL) GetHubspotObjectPropertiesName(ProjectID int64, objectType string) ([]string, []string) {
	defer model.LogOnSlowExecutionWithParams(time.Now(),
		&log.Fields{"project_id": ProjectID, "object_type": objectType})

	if ProjectID == 0 || objectType == "" {
		return nil, nil
	}

	docType, err := model.GetHubspotTypeByAlias(objectType)
	if err != nil {
		return nil, nil
	}

	logCtx := log.WithFields(log.Fields{"project_id": ProjectID, "object_type": objectType})

	hubspotDocuments, err := getLatestHubspotDocumentsByLimit(ProjectID, docType, C.GetHubspotPropertiesLookbackLimit())
	if err != nil {
		logCtx.WithError(err).Error("Failed to GetHubspotObjectPropertiesName")
		return nil, nil
	}

	return getHubspotDocumentPropertiesNameByType(hubspotDocuments)
}

// GetAllHubspotObjectValuesByPropertyName returns all values by property name
func (store *MemSQL) GetAllHubspotObjectValuesByPropertyName(ProjectID int64,
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

	logCtx := log.WithFields(log.Fields{"project_id": ProjectID,
		"object_type": objectType, "property_name": propertyName})

	hubspotDocuments, err := getLatestHubspotDocumentsByLimit(ProjectID, docType, C.GetHubspotPropertiesLookbackLimit())
	if err != nil {
		logCtx.WithError(err).Error("Failed to GetAllHubspotObjectPropertyValues")
		return nil
	}

	return getHubspotDocumentValuesByPropertyNameAndLimit(hubspotDocuments, propertyName, 100)
}

func (store *MemSQL) UpdateHubspotDocumentAsSynced(projectId int64, id string, docType int,
	syncId string, timestamp int64, action int, userID, groupUserID string) int {

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

	if groupUserID != "" {
		updates["group_user_id"] = groupUserID
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

// GetLastSyncedHubspotUpdateDocumentByID returns latest synced record by document id with preference to the Update doc if timestamp is same.
func (store *MemSQL) GetLastSyncedHubspotUpdateDocumentByID(projectID int64, docID string, docType int) (*model.HubspotDocument, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"doc_id":     docID,
		"doc_type":   docType,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "doc_id": docID, "doc_type": docType})

	if projectID == 0 || docType == 0 || docID == "" {
		logCtx.Error("Missing required field")
		return nil, http.StatusBadRequest
	}

	db := C.GetServices().Db

	var document []model.HubspotDocument

	if err := db.Where("project_id = ? AND type = ? AND id = ? and synced=true",
		projectID, docType, docID).Order("timestamp DESC").Limit(2).Find(&document).Error; err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			logCtx.WithError(err).Error("Failed to get latest hubspot document by userID.")
			return nil, http.StatusInternalServerError
		}
		return nil, http.StatusNotFound
	}

	if len(document) == 0 {
		return nil, http.StatusNotFound
	}

	if len(document) == 2 {
		// Prefer the latest doc by time
		if document[0].Timestamp > document[1].Timestamp {
			return &document[0], http.StatusFound
		}
		// Prefer the UpdatedActionEvent over CreateActionEvent
		if document[0].Action == model.HubspotDocumentActionUpdated {
			return &document[0], http.StatusFound
		}
		if document[1].Action == model.HubspotDocumentActionUpdated {
			return &document[1], http.StatusFound
		}
	}
	// Case with just one document.
	return &document[0], http.StatusFound
}

// GetLastSyncedHubspotDocumentByID returns latest synced record by document id.
func (store *MemSQL) GetLastSyncedHubspotDocumentByID(projectID int64, docID string, docType int) (*model.HubspotDocument, int) {
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

func (store *MemSQL) CreateOrUpdateGroupPropertiesBySource(projectID int64, groupName string, groupID, groupUserID string,
	enProperties *map[string]interface{}, createdTimestamp, updatedTimestamp int64, source string) (string, error) {
	logFields := log.Fields{
		"project_id":        projectID,
		"group_name":        groupName,
		"group_id":          groupID,
		"group_user_id":     groupUserID,
		"en_properties":     enProperties,
		"created_timestamp": createdTimestamp,
		"updated_timestamp": updatedTimestamp,
		"source":            source,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)
	if projectID < 1 || enProperties == nil || createdTimestamp == 0 || updatedTimestamp == 0 {
		logCtx.Error("Invalid parameters on CreateOrUpdateGroupPropertiesBySource.")
		return "", errors.New("invalid parameters")
	}

	if source != model.UserSourceHubspotString && source != model.UserSourceSalesforceString &&
		source != model.UserSourceSixSignalString && source != model.UserSourceDomainsString &&
		source != model.UserSourceLinkedinCompanyString && source != model.UserSourceG2String {
		logCtx.Error("Invalid source.")
		return "", errors.New("invalid source")
	}

	newGroupUser := false
	if groupUserID == "" {
		newGroupUser = true
	}

	pJSONProperties, err := util.EncodeToPostgresJsonb(enProperties)
	if err != nil {
		return "", err
	}

	if !newGroupUser {
		user, status := store.GetUserWithoutProperties(projectID, groupUserID)
		if status != http.StatusFound {
			return "", errors.New("failed to get user")
		}

		if !(*user.IsGroupUser) {
			return "", errors.New("user is not group user")
		}

		_, status = store.UpdateUserGroupProperties(projectID, groupUserID, pJSONProperties, updatedTimestamp)
		if status != http.StatusAccepted {
			logCtx.WithFields(log.Fields{"err_code": status}).Error("Failed to update user group properties.")
			return "", errors.New("failed to update company group properties")
		}

		currentGroupID, columnName := model.GetCurrentGroupIdAndColumnName(user)
		if currentGroupID != groupID {
			status = store.UpdateGroupUserGroupId(projectID, user.ID, groupID, columnName)
			if status != http.StatusAccepted {
				logCtx.WithFields(log.Fields{"err_code": status}).Error("Failed to update user groupID.")
				return "", errors.New("failed to update company groupID")
			}
		}

		store.addGroupUserPropertyDetailsToCache(projectID, groupName, groupUserID, enProperties)

		return groupUserID, nil
	}

	requestSource := model.GetUserSourceByName(source)

	isGroupUser := true
	userID, status := store.CreateGroupUser(&model.User{
		ProjectId:     projectID,
		IsGroupUser:   &isGroupUser,
		JoinTimestamp: createdTimestamp,
		Source:        &requestSource,
	}, groupName, groupID)
	if status != http.StatusCreated {
		logCtx.WithFields(log.Fields{"err_code": status}).Error("Failed to create group user.")
		return userID, errors.New("failed to create company group user")
	}

	_, status = store.UpdateUserGroupProperties(projectID, userID, pJSONProperties, updatedTimestamp)
	if status != http.StatusAccepted {
		return userID, errors.New("failed to update company group properties")
	}

	store.addGroupUserPropertyDetailsToCache(projectID, groupName, userID, enProperties)

	return userID, nil
}
