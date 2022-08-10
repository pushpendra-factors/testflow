package memsql

import (
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"github.com/jinzhu/gorm"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) satisfiesOTPRuleForeignConstraints(rule model.OTPRule) int {
	logFields := log.Fields{
		"rule": rule,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	_, errCode := store.GetProject(rule.ProjectID)
	if errCode != http.StatusFound {
		return http.StatusBadRequest
	}

	return http.StatusOK
}

func (store *MemSQL) CreateOTPRule(projectId int64, rule *model.OTPRule) (*model.OTPRule, int, string) {
	logFields := log.Fields{
		"project_id": projectId,
		"rule":       rule,
	}

	if rule == nil {
		return nil, http.StatusInternalServerError, "rule is empty"
	}

	if rule.ID == "" {
		rule.ID = U.GetUUID()
	}

	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	if projectId == 0 {
		return nil, http.StatusBadRequest, "Invalid request"
	}

	rule.ProjectID = projectId
	if errCode := store.satisfiesOTPRuleForeignConstraints(*rule); errCode != http.StatusOK {
		return nil, http.StatusInternalServerError, "Foreign constraints violation"
	}

	if err := db.Create(&rule).Error; err != nil {
		errMsg := "Failed to insert rule."
		log.WithFields(log.Fields{"Rule": rule,
			"project_id": projectId}).WithError(err).Error(errMsg)
		return nil, http.StatusInternalServerError, errMsg
	}
	return rule, http.StatusCreated, ""
}

// GetALLOTPRuleWithProjectId Get all otpRules which are active (not-deleted).
func (store *MemSQL) GetALLOTPRuleWithProjectId(projectID int64) ([]model.OTPRule, int) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	otpRules := make([]model.OTPRule, 0, 0)
	err := db.Table("otp_rules").Select("*").
		Where("project_id = ? AND is_deleted = ?", projectID, false).
		Order("created_at DESC").Find(&otpRules).Error

	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return otpRules, http.StatusNotFound
		}

		log.WithField("project_id", projectID).WithError(err).Error("Failed to fetch rows from otp_rules table for project")
		return otpRules, http.StatusInternalServerError
	}
	return otpRules, http.StatusFound
}

// GetAllRulesDeletedNotDeleted fetching deleted, non-deleted rules.
func (store *MemSQL) GetAllRulesDeletedNotDeleted(projectID int64) ([]model.OTPRule, int) {
	db := C.GetServices().Db

	otpRules := make([]model.OTPRule, 0, 0)
	err := db.Table("otp_rules").Select("*").
		Where("project_id = ? AND converted = false ", projectID).
		Order("created_at DESC").Find(&otpRules).Error

	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return otpRules, http.StatusNotFound
		}
		log.WithField("project_id", projectID).WithError(err).Error("Failed to fetch rows from otp_rules table for project")
		return otpRules, http.StatusInternalServerError
	}
	return otpRules, http.StatusFound
}

// GetOTPRuleWithRuleId Get rule for given project_id and ID which is not deleted.
func (store *MemSQL) GetOTPRuleWithRuleId(projectID int64, ruleID string) (*model.OTPRule, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"rule_id":    ruleID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	return store.getRuleWithRuleID(projectID, ruleID)
}

func (store *MemSQL) getRuleWithRuleID(projectID int64, ruleID string) (*model.OTPRule, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"rule_id":    ruleID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db
	var rule model.OTPRule
	var err error
	err = db.Table("otp_rules").Where("project_id = ? AND id=? AND is_deleted = ?",
		projectID, ruleID, false).Find(&rule).Error

	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return &model.OTPRule{}, http.StatusNotFound
		}

		log.WithField("project_id", projectID).WithError(err).Error("Failed to get otp rules table.")
		return &model.OTPRule{}, http.StatusInternalServerError
	}

	return &rule, http.StatusFound
}

// GetAnyOTPRuleWithRuleId Get rule for given project_id and ID which is not deleted.
func (store *MemSQL) GetAnyOTPRuleWithRuleId(projectID int64, ruleID string) (*model.OTPRule, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"rule_id":    ruleID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	return store.getAnyRuleWithRuleID(projectID, ruleID)
}

func (store *MemSQL) getAnyRuleWithRuleID(projectID int64, ruleID string) (*model.OTPRule, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"rule_id":    ruleID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db
	var rule model.OTPRule
	err := db.Table("otp_rules").Where("project_id = ? AND id=?",
		projectID, ruleID).Find(&rule).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return &rule, http.StatusNotFound
		}

		log.WithField("project_id", projectID).WithError(err).Error("Failed to fetch rows from otp_rules table for project")
		return &rule, http.StatusInternalServerError
	}

	return &rule, http.StatusFound
}

// DeleteOTPRule To delete an otp rule.
func (store *MemSQL) DeleteOTPRule(projectID int64, ruleID string) (int, string) {
	logFields := log.Fields{
		"project_id": projectID,
		"rule_id":    ruleID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	return deleteOTPRule(projectID, ruleID)
}

func deleteOTPRule(projectID int64, ruleID string) (int, string) {
	logFields := log.Fields{
		"project_id": projectID,
		"rule_id":    ruleID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db
	if projectID == 0 {
		return http.StatusBadRequest, "Invalid project ID"
	}
	if ruleID == "" {
		return http.StatusBadRequest, "Invalid rule ID"
	}

	var err error
	err = db.Model(&model.OTPRule{}).Where("id= ? AND project_id=?", ruleID, projectID).
		Update(map[string]interface{}{"is_deleted": true}).Error
	if err != nil {
		return http.StatusInternalServerError, "Failed to delete saved rule"
	}
	return http.StatusAccepted, ""
}

func (store *MemSQL) UpdateOTPRule(projectID int64, ruleID string, rule *model.OTPRule) (*model.OTPRule, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"rule_id":    ruleID,
		"rule":       rule,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	if ruleID == "" {
		return &model.OTPRule{}, http.StatusBadRequest
	}

	// update allowed fields.
	updateFields := make(map[string]interface{}, 0)
	if rule.RuleType != "" {
		updateFields["rule_type"] = rule.RuleType
	}
	if rule.TouchPointTimeRef != "" {
		updateFields["touch_point_time_ref"] = rule.TouchPointTimeRef
	}

	if !U.IsEmptyPostgresJsonb(&rule.Filters) {
		updateFields["filters"] = rule.Filters
	}

	if !U.IsEmptyPostgresJsonb(&rule.PropertiesMap) {
		updateFields["properties_map"] = rule.PropertiesMap
	}

	err := db.Model(&model.OTPRule{}).Where("project_id = ? AND id=? AND is_deleted = ?",
		projectID, ruleID, false).Update(updateFields).Error
	if err != nil {
		return &model.OTPRule{}, http.StatusInternalServerError
	}
	return rule, http.StatusAccepted
}
