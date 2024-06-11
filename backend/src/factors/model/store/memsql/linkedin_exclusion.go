package memsql

import (
	C "factors/config"
	"factors/model/model"
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) GetLinkedinCappingExclusionsForRule(projectID int64, ruleID string) ([]model.LinkedinExclusion, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"rule_id":    ruleID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	linkedinExclusions := make([]model.LinkedinExclusion, 0)
	db := C.GetServices().Db

	err := db.Model(&model.LinkedinExclusion{}).Where("project_id = ? and rule_id", projectID, ruleID).Limit("1000").
		Find(&linkedinExclusions).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return make([]model.LinkedinExclusion, 0), http.StatusOK
		}
		logCtx.WithError(err).Error("Failed to get exclusion")
		return make([]model.LinkedinExclusion, 0), http.StatusInternalServerError
	}

	return linkedinExclusions, http.StatusOK
}
func (store *MemSQL) GetAllLinkedinCappingExclusionsForTimerange(projectID int64, monthStart int64, monthEnd int64) ([]model.LinkedinExclusion, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"monthStart": monthStart,
		"monthEnd":   monthEnd,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	linkedinExclusions := make([]model.LinkedinExclusion, 0)
	db := C.GetServices().Db
	err := db.Model(&model.LinkedinExclusion{}).Where("project_id = ? and timestamp between ? and ?", projectID, monthStart, monthEnd).
		Find(&linkedinExclusions).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return make([]model.LinkedinExclusion, 0), http.StatusOK
		}
		logCtx.WithError(err).Error("Failed to get exclusion")
		return make([]model.LinkedinExclusion, 0), http.StatusInternalServerError
	}
	return linkedinExclusions, http.StatusFound
}

func (store *MemSQL) CreateLinkedinExclusion(projectID int64, linkedinExclusionDoc model.LinkedinExclusion) int {
	logFields := log.Fields{
		"project_id":             projectID,
		"linkedin_exclusion_doc": linkedinExclusionDoc,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	linkedinExclusionDoc.CreatedAt = time.Now()
	linkedinExclusionDoc.UpdatedAt = time.Now()

	db := C.GetServices().Db

	err := db.Create(&linkedinExclusionDoc).Error
	if err != nil {
		if IsDuplicateRecordError(err) {
			logCtx.WithError(err).WithField("project_id", linkedinExclusionDoc.ProjectID).Warn(
				"Failed to create rule object. Duplicate.")
			return http.StatusConflict
		}
		logCtx.WithError(err).WithField("project_id", linkedinExclusionDoc.ProjectID).Error(
			"Failed to create rule object.")
		return http.StatusInternalServerError
	}

	return http.StatusCreated
}

func (store *MemSQL) GetNonPushedExclusionsForMonth(projectID int64, startOfMonthDate int64) ([]model.LinkedinExclusion, int) {
	logFields := log.Fields{
		"project_id":       projectID,
		"startOfMonthDate": startOfMonthDate,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	linkedinExclusions := make([]model.LinkedinExclusion, 0)
	db := C.GetServices().Db
	err := db.Model(&model.LinkedinExclusion{}).
		Where("project_id = ? and timestamp >= ? and is_pushed_to_linkedin = false", projectID, startOfMonthDate).
		Find(&linkedinExclusions).Error

	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return make([]model.LinkedinExclusion, 0), http.StatusOK
		}
		logCtx.WithError(err).WithField("project_id", projectID).Error(
			"Failed to get exclusions.")
		return linkedinExclusions, http.StatusInternalServerError
	}

	return linkedinExclusions, http.StatusOK
}
func (store *MemSQL) GetNonRemovedExclusionsForMonth(projectID int64, startOfMonthDate int64, endOfMonthDate int64) ([]model.LinkedinExclusion, int) {
	logFields := log.Fields{
		"project_id":       projectID,
		"startOfMonthDate": startOfMonthDate,
		"endOfMonthDate":   endOfMonthDate,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	linkedinExclusions := make([]model.LinkedinExclusion, 0)
	db := C.GetServices().Db
	err := db.Model(&model.LinkedinExclusion{}).
		Where("project_id = ? and timestamp >= ? and timestamp <= ? and is_removed_from_linkedin = false", projectID, startOfMonthDate, endOfMonthDate).
		Find(&linkedinExclusions).Error

	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return make([]model.LinkedinExclusion, 0), http.StatusOK
		}
		logCtx.WithError(err).WithField("project_id", projectID).Error(
			"Failed to get exclusions.")
		return linkedinExclusions, http.StatusInternalServerError
	}

	return linkedinExclusions, http.StatusOK
}

func (store *MemSQL) UpdateLinkedinPushSyncStatusForOrgAndRule(projectID int64, startOfMonthDate int64, orgID string, ruleID string) int {
	logFields := log.Fields{
		"project_id":       projectID,
		"startOfMonthDate": startOfMonthDate,
		"org_id":           orgID,
		"rule_id":          ruleID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	updatedFields := map[string]interface{}{
		"is_pushed_to_linkedin": true,
	}

	db := C.GetServices().Db
	err := db.Model(&model.LinkedinExclusion{}).
		Where("project_id = ? and timestamp >= ? and org_id = ? and rule_id = ?", projectID, startOfMonthDate, orgID, ruleID).
		Updates(updatedFields).Error

	if err != nil {
		logCtx.WithError(err).WithField("project_id", projectID).Error(
			"Failed to get exclusions.")
		return http.StatusInternalServerError
	}
	return http.StatusOK
}

func (store *MemSQL) UpdateLinkedinRemoveSyncStatusForOrgAndRule(projectID int64, startOfMonthDate int64, endOfMonthDate int64, orgID string, ruleID string) int {
	logFields := log.Fields{
		"project_id":       projectID,
		"startOfMonthDate": startOfMonthDate,
		"org_id":           orgID,
		"rule_id":          ruleID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	updatedFields := map[string]interface{}{
		"is_removed_from_linkedin": true,
	}

	db := C.GetServices().Db
	err := db.Model(&model.LinkedinExclusion{}).
		Where("project_id = ? and timestamp >= ? and timestamp <= ? and org_id = ? and rule_id = ?", projectID, startOfMonthDate, endOfMonthDate, orgID, ruleID).
		Updates(updatedFields).Error

	if err != nil {
		logCtx.WithError(err).WithField("project_id", projectID).Error(
			"Failed to get exclusions.")
		return http.StatusInternalServerError
	}
	return http.StatusOK
}
