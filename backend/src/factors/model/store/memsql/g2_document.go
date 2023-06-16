package memsql

import (
	C "factors/config"
	"factors/model/model"
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

const insertG2Str = "INSERT INTO g2_documents (project_id,type,timestamp,id,value,created_at,updated_at) VALUES "

var G2DocumentTypeAlias = map[string]int{
	"event_stream": 1,
}

func validateG2Documents(g2Documents []model.G2Document) int {
	for index, document := range g2Documents {
		if document.TypeAlias == "" {
			log.WithField("document", document).Error("Invalid g2 document.")
			return http.StatusBadRequest
		}
		if document.ProjectID == 0 || document.Timestamp == 0 {
			log.WithField("document", document).Error("Invalid g2 document.")
			return http.StatusBadRequest
		}

		docType, docTypeExists := G2DocumentTypeAlias[document.TypeAlias]
		if !docTypeExists {
			log.WithField("document", document).Error("Invalid g2 document.")
			return http.StatusBadRequest
		}
		addG2DocType(&g2Documents[index], docType)
	}
	return http.StatusOK
}
func addG2DocType(g2Doc *model.G2Document, docType int) {
	g2Doc.Type = docType
}

func (store *MemSQL) CreateMultipleG2Document(g2Documents []model.G2Document) int {
	logFields := log.Fields{"g2_documents": g2Documents}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	status := validateG2Documents(g2Documents)
	if status != http.StatusOK {
		return status
	}

	db := C.GetServices().Db

	insertStatement := insertG2Str
	insertValuesStatement := make([]string, 0)
	insertValues := make([]interface{}, 0)
	for _, doc := range g2Documents {
		insertValuesStatement = append(insertValuesStatement, fmt.Sprintf("(?, ?, ?, ?, ?, ?, ?)"))
		insertValues = append(insertValues, doc.ProjectID,
			doc.Type, doc.Timestamp, doc.ID, doc.Value, time.Now(), time.Now())
		UpdateCountCacheByDocumentType(doc.ProjectID, &doc.CreatedAt, "g2")
	}
	insertStatement += joinWithComma(insertValuesStatement...)
	rows, err := db.Raw(insertStatement, insertValues...).Rows()

	if err != nil {
		if IsDuplicateRecordError(err) {
			log.WithError(err).WithField("g2Documents", g2Documents).Error("Failed to create an g2 doc. Duplicate.")
			return http.StatusConflict
		} else {
			log.WithError(err).WithField("g2Documents", g2Documents).Error(
				"Failed to create an linkedin doc. Continued inserting other docs.")
			return http.StatusInternalServerError
		}
	}
	defer rows.Close()

	if status != http.StatusOK {
		return status
	}
	return http.StatusCreated
}

func (store *MemSQL) GetG2LastSyncInfo(projectID int64) ([]model.G2LastSyncInfo, int) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	g2LastSyncInfos := make([]model.G2LastSyncInfo, 0)

	queryStr := "SELECT project_id, type, max(timestamp) as timestamp" +
		" FROM g2_documents WHERE project_id = ?" +
		" GROUP BY project_id, type"

	err := db.Raw(queryStr, projectID).Scan(&g2LastSyncInfos).Error
	if err != nil {
		log.WithError(err).Error("Failed to get last g2 documents by type for sync info.")
		return g2LastSyncInfos, http.StatusInternalServerError
	}

	documentTypeAliasByType := getG2DocumentTypeAliasByType()

	for i := range g2LastSyncInfos {
		logCtx := log.WithFields(logFields)
		typeAlias, typeAliasExists := documentTypeAliasByType[g2LastSyncInfos[i].Type]
		if !typeAliasExists {
			logCtx.WithField("document_type",
				g2LastSyncInfos[i].Type).Error("Invalid document type given. No type alias name.")
			continue
		}

		g2LastSyncInfos[i].TypeAlias = typeAlias
	}

	return g2LastSyncInfos, http.StatusOK
}

func getG2DocumentTypeAliasByType() map[int]string {

	defer model.LogOnSlowExecutionWithParams(time.Now(), nil)
	documentTypeMap := make(map[int]string, 0)
	for alias, typ := range G2DocumentTypeAlias {
		documentTypeMap[typ] = alias
	}

	return documentTypeMap
}

func (store *MemSQL) GetG2DocumentsForGroupUserCreation(projectID int64) ([]model.G2Document, int) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	g2Documents := make([]model.G2Document, 0, 0)

	queryStr := "SELECT * FROM g2_documents WHERE project_id = ?" +
		" AND synced = FALSE" +
		" ORDER BY timestamp ASC"

	err := db.Raw(queryStr, projectID).Scan(&g2Documents).Error
	if err != nil {
		log.WithError(err).Error("Failed to get g2 documents for group user creation.")
		return g2Documents, http.StatusInternalServerError
	}

	return g2Documents, http.StatusOK
}

func (store *MemSQL) UpdateG2GroupUserCreationDetails(g2Document model.G2Document) error {
	db := C.GetServices().Db
	err := db.Table("g2_documents").Where("project_id = ? and id = ? and type = 1 and timestamp = ?", g2Document.ProjectID, g2Document.ID, g2Document.Timestamp).Updates(map[string]interface{}{"synced": true}).Error
	if err != nil {
		return err
	}
	return nil
}
