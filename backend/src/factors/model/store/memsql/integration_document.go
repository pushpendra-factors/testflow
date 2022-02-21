package memsql

import (
	"errors"
	C "factors/config"
	"factors/model/model"

	U "factors/util"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

const error_duplicateDocument string = "pq: duplicate key value violates unique constraint \"docid_projectid_source_doctype_timestamp_custacc_unique_idx\""

func isDuplicateRecord(err error) bool {
	return err.Error() == error_duplicateDocument
}
func (store *MemSQL) UpsertIntegrationDocument(doc model.IntegrationDocument) error {
	db := C.GetServices().Db

	if ValidateIntegrationDocument(doc) == false {
		log.Error("Validation Failed for the document")
		return errors.New("Document validation failed")
	}
	doc.CreatedAt = U.TimeNowZ()
	doc.UpdatedAt = U.TimeNowZ()
	if !isUniquenessConstraintSatisfied(doc) {
		updatedFields := map[string]interface{}{
			"value":      doc.Value,
			"updated_at": doc.UpdatedAt,
		}
		dbErr := db.Model(&model.IntegrationDocument{}).Where("document_id = ? AND document_type = ? AND source = ? AND customer_account_id = ? AND project_id = ? AND timestamp = ?",
			doc.DocumentId, doc.DocumentType, doc.Source, doc.CustomerAccountID, doc.ProjectID, doc.Timestamp).Update(updatedFields).Error
		if dbErr != nil {
			log.WithError(dbErr).Error("updating integration document failed")
			return dbErr
		}
	} else {
		if err := db.Create(&doc).Error; err != nil {
			log.WithError(err).Error("Insert into integration document table failed")
			return err
		}
	}
	return nil
}

func isUniquenessConstraintSatisfied(doc model.IntegrationDocument) bool {
	db := C.GetServices().Db
	if err := db.Limit(1).Where("project_id = ? AND customer_account_id = ? AND document_type = ? AND timestamp = ? AND document_id = ? AND source = ?",
		doc.ProjectID, doc.CustomerAccountID, doc.DocumentType, doc.Timestamp, doc.DocumentId, doc.Source,
	).Select("document_id").Find(&model.IntegrationDocument{}).Error; err != nil {

		if gorm.IsRecordNotFoundError(err) {
			return true
		}

		log.WithError(err).
			Error("Failed getting to check existence bingads document by primary keys.")
		return false
	}
	return false
}

func (store *MemSQL) InsertIntegrationDocument(doc model.IntegrationDocument) error {
	db := C.GetServices().Db

	if ValidateIntegrationDocument(doc) == false {
		log.Error("Validation Failed for the document")
		return errors.New("Document validation failed")
	}
	doc.CreatedAt = U.TimeNowZ()
	doc.UpdatedAt = U.TimeNowZ()
	if err := db.Create(&doc).Error; err != nil {

		log.WithError(err).Error("Insert into integration document table failed")
		return err
	}
	return nil
}

func ValidateIntegrationDocument(doc model.IntegrationDocument) bool {
	if doc.DocumentId == "" || doc.DocumentType == 0 || doc.CustomerAccountID == "" || doc.ProjectID == 0 || doc.Source == "" || doc.Timestamp == 0 {
		return false
	}
	return true
}
