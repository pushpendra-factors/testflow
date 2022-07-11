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

// isExistRelationshipByID check for existing relationship between two different object type by their external id.
func isExistRelationshipByID(projectID int64, source U.CRMSource, fromType,
	toType int, fromID, toID string) (int, error) {
	logFields := log.Fields{
		"project_id": projectID,
		"source":     source,
		"from_type":  fromType,
		"to_type":    toType,
		"from_id":    fromID,
		"to_id":      toID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)
	if !model.AllowedCRMBySource(source) {
		logCtx.Error("Invalid source.")
		return http.StatusBadRequest, errors.New("invalid source")
	}

	if projectID == 0 || fromType <= 0 || toType <= 0 ||
		fromID == "" || toID == "" {
		logCtx.Error("Missing required parameters.")
		return http.StatusBadRequest, errors.New("missing required fields project_id, from_type, to_type, from_id, to_id")
	}

	var relationship model.CRMRelationship
	db := C.GetServices().Db
	err := db.Model(&model.CRMRelationship{}).Where("project_id = ? AND source = ? "+
		"AND from_type = ? AND to_type = ? AND from_id = ? AND to_id = ?",
		projectID, source, fromType, toType, fromID, toID).Select("from_id").Limit(1).Find(&relationship).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return http.StatusNotFound, nil
		}

		logCtx.WithError(err).Error("Failed to get relationship from crm_relationship table.")
		return http.StatusInternalServerError, err
	}

	if relationship.FromID == "" {
		return http.StatusNotFound, nil
	}

	return http.StatusFound, nil
}

// CreateCRMRelationship only store relationship between two different objects
func (store *MemSQL) CreateCRMRelationship(crmRelationship *model.CRMRelationship) (int, error) {
	logFields := log.Fields{
		"project_id":                 crmRelationship.ProjectID,
		"external_relationship_id":   crmRelationship.ExternalRelationshipID,
		"external_relationship_name": crmRelationship.ExternalRelationshipName,
		"from_type":                  crmRelationship.FromType,
		"to_type":                    crmRelationship.ToType,
		"from_id":                    crmRelationship.FromID,
		"to_id":                      crmRelationship.ToID,
		"timestamp":                  crmRelationship.Timestamp,
		"skip_process":               crmRelationship.SkipProcess,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)

	if !model.AllowedCRMBySource(crmRelationship.Source) {
		logCtx.Error("Invalid source.")
		return http.StatusBadRequest, errors.New("invalid source")
	}

	/*
		fromType and toType refer relationship between two different object
		fromId and toId refer the id of the related object ids
	*/
	if crmRelationship.ProjectID == 0 || crmRelationship.FromType <= 0 ||
		crmRelationship.ToType <= 0 || crmRelationship.FromID == "" || crmRelationship.ToID == "" {
		logCtx.Error("Missing required parameters")
		return http.StatusBadRequest, errors.New("missing required fields project_id, from_type, to_type, from_id, to_id")
	}

	if crmRelationship.FromType == crmRelationship.ToType {
		logCtx.Error("From type and to type cannot be same.")
		return http.StatusBadRequest, errors.New("same type in from_type and to_type")
	}

	// timestamp when the relationship was created, can be related to source timestamp or use current timestamp
	if crmRelationship.Timestamp <= 0 {
		logCtx.Error("Invalid timestamp")
		return http.StatusBadRequest, errors.New("missing timestamp")
	}

	if crmRelationship.Properties != nil && U.IsEmptyPostgresJsonb(crmRelationship.Properties) {
		logCtx.Error("Empty relationship properties properties.")
		return http.StatusBadRequest, errors.New("empty value")
	}

	status, err := isExistRelationshipByID(crmRelationship.ProjectID, crmRelationship.Source,
		crmRelationship.FromType, crmRelationship.ToType, crmRelationship.FromID, crmRelationship.ToID)
	if status != http.StatusNotFound {
		if status == http.StatusFound {
			return http.StatusConflict, nil
		}

		logCtx.WithError(err).Error("Failed check for existing relationship.")
		return status, err
	}

	crmRelationship.ID = U.GetUUID()
	db := C.GetServices().Db
	err = db.Create(&crmRelationship).Error
	if err != nil {
		if IsDuplicateRecordError(err) {
			return http.StatusConflict, nil
		}

		logCtx.WithError(err).Error("Failed to insert crm relationship document.")
		return http.StatusInternalServerError, err
	}

	return http.StatusCreated, nil
}
