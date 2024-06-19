package memsql

import (
	"errors"
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"net/http"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

const campaignGroupConfigFetchStr = "With campaign_group_timestamp as (Select campaign_group_id as c1, max(timestamp) as t1 from linkedin_documents where project_id = ? " +
	"and type = 2 and timestamp >= ? group by campaign_group_id) " +
	"select campaign_group_id as id, 'campaign_group' as type, false as deleted, JSON_EXTRACT_STRING(value, 'campaign_group_name') as name from linkedin_documents inner join " +
	"campaign_group_timestamp on c1 = campaign_group_id and t1=timestamp where project_id = ? and type = 2 and timestamp >= ?"

const campaignConfigFetchStr = "With campaign_timestamp as (Select campaign_id as c1, max(timestamp) as t1 from linkedin_documents where project_id = ? " +
	"and type = 3 and timestamp >= ? group by campaign_id) " +
	"select campaign_id as id, 'campaign' as type, false as deleted, JSON_EXTRACT_STRING(value, 'campaign_name') as name from linkedin_documents inner join " +
	"campaign_timestamp on c1 = campaign_id and t1=timestamp where project_id = ? and type = 3 and timestamp >= ?"

	// rethink active/paused -> and timerange
	// Show whether active or not as well
	// Comment what query is doing.
func (store *MemSQL) GetLinkedinFreqCappingConfig(projectID int64) ([]model.LinkedinCappingConfig, int) {
	campaignGroupConfig := make([]model.LinkedinCappingConfig, 0)
	campaignConfig := make([]model.LinkedinCappingConfig, 0)
	config := make([]model.LinkedinCappingConfig, 0)
	timestamp2MonthsAgo, _ := strconv.ParseInt(time.Now().AddDate(0, -2, 0).Format("20060102"), 10, 64)

	db := C.GetServices().Db
	query := campaignGroupConfigFetchStr
	err := db.Raw(query, projectID, timestamp2MonthsAgo, projectID, timestamp2MonthsAgo).Scan(&campaignGroupConfig).Error
	if err != nil {
		log.WithError(err).Error("failed to get campaign group config")
		return make([]model.LinkedinCappingConfig, 0), http.StatusInternalServerError
	}
	query = campaignConfigFetchStr
	err = db.Raw(query, projectID, timestamp2MonthsAgo, projectID, timestamp2MonthsAgo).Scan(&campaignConfig).Error
	if err != nil {
		log.WithError(err).Error("failed to get campaign config")
		return make([]model.LinkedinCappingConfig, 0), http.StatusInternalServerError
	}
	config = append(config, campaignConfig...)
	config = append(config, campaignGroupConfig...)
	return config, http.StatusOK
}

func (store *MemSQL) CreateLinkedinCappingRule(projectID int64, linkedinCappingRule *model.LinkedinCappingRule) (model.LinkedinCappingRule, string, int) {
	logFields := log.Fields{
		"project_id":             projectID,
		"linkedin_capping_rules": linkedinCappingRule,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	linkedinCappingRule.Name = model.GenerateNameFromDisplayName(linkedinCappingRule.DisplayName)
	if isPresent := store.checkIfNameAlreadyPresent(projectID, linkedinCappingRule.Name, linkedinCappingRule.ObjectType); isPresent {
		return model.LinkedinCappingRule{}, "duplicate name", http.StatusBadRequest
	}
	linkedinCappingRule.ID = U.GetUUID()
	linkedinCappingRule.CreatedAt = time.Now().UTC()
	linkedinCappingRule.UpdatedAt = time.Now().UTC()

	err := store.validateLinkedinCappingRule(projectID, linkedinCappingRule)
	if err != nil {
		logCtx.Error(err.Error())
		return model.LinkedinCappingRule{}, err.Error(), http.StatusBadRequest
	}

	db := C.GetServices().Db
	err = db.Create(&linkedinCappingRule).Error
	if err != nil {
		if IsDuplicateRecordError(err) {
			logCtx.WithError(err).WithField("project_id", linkedinCappingRule.ProjectID).Warn(
				"Failed to create rule object. Duplicate.")
			return model.LinkedinCappingRule{}, "Duplicate Rule", http.StatusConflict
		}
		logCtx.WithError(err).WithField("project_id", linkedinCappingRule.ProjectID).Error(
			"Failed to create rule object.")
		return model.LinkedinCappingRule{}, "Internal server error", http.StatusInternalServerError
	}

	return *linkedinCappingRule, "", http.StatusCreated
}

func (store *MemSQL) validateLinkedinCappingRule(projectID int64, linkedinCappingRule *model.LinkedinCappingRule) error {
	if projectID == 0 {
		return errors.New("invalid project_id")
	}
	if linkedinCappingRule.ObjectType == "" {
		return errors.New("invalid object_type")
	}
	// String util can be used directly.
	if linkedinCappingRule.ObjectType != model.LINKEDIN_ACCOUNT &&
		linkedinCappingRule.ObjectType != model.LINKEDIN_CAMPAIGN &&
		linkedinCappingRule.ObjectType != model.LINKEDIN_CAMPAIGN_GROUP {
		return errors.New("invalid object type")
	}
	objectIDs := make([]string, 0)
	if linkedinCappingRule.ObjectIDs != nil {
		err := U.DecodePostgresJsonbToStructType(linkedinCappingRule.ObjectIDs, &objectIDs)
		if err != nil {
			return err
		}
	}
	if linkedinCappingRule.ObjectType == model.LINKEDIN_ACCOUNT && len(objectIDs) > 0 {
		return errors.New("object IDs found for account level rules")
	} else if linkedinCappingRule.ObjectType != model.LINKEDIN_ACCOUNT && len(objectIDs) == 0 {
		return errors.New("object IDs not found for campaign/campaign_group level rules")
	}
	// Product/cs inform
	if linkedinCappingRule.ImpressionThreshold < 500 && linkedinCappingRule.ClickThreshold < 3 {
		return errors.New("min threshold value for impressions 500, clicks 3")
	}
	advRules := make([]model.AdvancedRuleFilters, 0)
	if linkedinCappingRule.AdvancedRules != nil {
		err := U.DecodePostgresJsonbToStructType(linkedinCappingRule.AdvancedRules, &advRules)
		if err != nil {
			return err
		}
	}
	// Here only account. Make the intention of variable clear.
	if len(advRules) > 0 && (linkedinCappingRule.AdvancedRuleType != model.LINKEDIN_ACCOUNT) {
		return errors.New("advanced rule type not defined")
	}
	for _, subRule := range advRules {
		// Check with product team
		if subRule.ImpressionThreshold < 500 && subRule.ClickThreshold < 3 {
			return errors.New("min threshold value for impressions 500, clicks 3")
		}
	}

	err := store.checkObjectIDsValidity(projectID, linkedinCappingRule.ObjectType, objectIDs)
	if err != nil {
		return err
	}
	return nil
}

func (store *MemSQL) checkIfNameAlreadyPresent(projectID int64, name string, objectType string) bool {
	logFields := log.Fields{
		"project_id":  projectID,
		"name":        name,
		"object_type": objectType,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	db := C.GetServices().Db
	linkedinCappingRules := make([]model.LinkedinCappingRule, 0)

	err := db.Table("linkedin_capping_rules").
		Where("project_id = ? AND name = ? AND object_type = ? and status != ?", projectID, name, objectType, model.LINKEDIN_STATUS_DELETED).
		Find(&linkedinCappingRules).Error
	if err != nil || len(linkedinCappingRules) == 0 {
		return false
	}

	return true
}

func (store *MemSQL) checkObjectIDsValidity(projectID int64, objectType string, objectIDs []string) error {
	if objectType == model.LINKEDIN_ACCOUNT {
		return nil
	}
	linkedinConfig, errCode := store.GetLinkedinFreqCappingConfig(projectID)
	if errCode != http.StatusOK {
		return errors.New("unable to fetch linkedin config")
	}

	// See if a util/set method probably exists.
	for _, id := range objectIDs {
		found := false
		for _, config := range linkedinConfig {
			if id == config.ID && config.Type == objectType {
				found = true
			}
		}
		if !found {
			return errors.New("invalid objectID found")
		}
	}
	return nil
}

func (store *MemSQL) GetAllLinkedinCappingRules(projectID int64) ([]model.LinkedinCappingRule, int) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	db := C.GetServices().Db
	linkedinCappingRules := make([]model.LinkedinCappingRule, 0)

	err := db.Table("linkedin_capping_rules").
		Where("project_id = ? AND status != ?", projectID, model.LINKEDIN_STATUS_DELETED).
		Find(&linkedinCappingRules).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return make([]model.LinkedinCappingRule, 0), http.StatusFound
		}
		return make([]model.LinkedinCappingRule, 0), http.StatusInternalServerError
	}

	return linkedinCappingRules, http.StatusFound
}

func (store *MemSQL) GetAllActiveLinkedinCappingRules(projectID int64) ([]model.LinkedinCappingRule, int) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	db := C.GetServices().Db
	linkedinCappingRules := make([]model.LinkedinCappingRule, 0)

	err := db.Table("linkedin_capping_rules").
		Where("project_id = ? AND status = ?", projectID, model.LINKEDIN_STATUS_ACTIVE).
		Find(&linkedinCappingRules).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return make([]model.LinkedinCappingRule, 0), http.StatusNotFound
		}
		return make([]model.LinkedinCappingRule, 0), http.StatusInternalServerError
	}

	return linkedinCappingRules, http.StatusFound
}

func (store *MemSQL) GetAllActiveLinkedinCappingRulesByObjectType(projectID int64, objectType string) ([]model.LinkedinCappingRule, int) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	db := C.GetServices().Db
	linkedinCappingRules := make([]model.LinkedinCappingRule, 0)

	err := db.Table("linkedin_capping_rules").
		Where("project_id = ? AND status = ? and object_type = ?", projectID, model.LINKEDIN_STATUS_ACTIVE, objectType).
		Find(&linkedinCappingRules).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return make([]model.LinkedinCappingRule, 0), http.StatusNotFound
		}
		return make([]model.LinkedinCappingRule, 0), http.StatusInternalServerError
	}

	return linkedinCappingRules, http.StatusFound
}
func (store *MemSQL) GetLinkedinCappingRule(projectID int64, ruleID string) (model.LinkedinCappingRule, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"rule_id":    ruleID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	db := C.GetServices().Db
	var linkedinCappingRule model.LinkedinCappingRule
	err := db.Table("linkedin_capping_rules").
		Where("project_id = ? AND status != ? AND id = ?", projectID, model.LINKEDIN_STATUS_DELETED, ruleID).
		Find(&linkedinCappingRule).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return model.LinkedinCappingRule{}, http.StatusNotFound
		}
		return model.LinkedinCappingRule{}, http.StatusInternalServerError
	}

	return linkedinCappingRule, http.StatusFound
}
func (store *MemSQL) UpdateLinkedinCappingRule(projectID int64, linkedinCappingRule *model.LinkedinCappingRule) (string, int) {
	logFields := log.Fields{
		"project_id":             projectID,
		"linkedin_capping_rules": linkedinCappingRule,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	linkedinCappingRule.UpdatedAt = time.Now().UTC()

	err := store.validateLinkedinCappingRule(projectID, linkedinCappingRule)
	if err != nil {
		logCtx.Error(err.Error())
		return err.Error(), http.StatusBadRequest
	}

	updatedFields := map[string]interface{}{
		"display_name":             linkedinCappingRule.DisplayName,
		"object_type":              linkedinCappingRule.ObjectType,
		"object_ids":               linkedinCappingRule.ObjectIDs,
		"description":              linkedinCappingRule.Description,
		"status":                   linkedinCappingRule.Status,
		"granularity":              linkedinCappingRule.Granularity,
		"impression_threshold":     linkedinCappingRule.ImpressionThreshold,
		"click_threshold":          linkedinCappingRule.ClickThreshold,
		"is_advanced_rule_enabled": linkedinCappingRule.IsAdvancedRuleEnabled,
		"advanced_rule_type":       linkedinCappingRule.AdvancedRuleType,
		"advanced_rules":           linkedinCappingRule.AdvancedRules,
		"updated_at":               time.Now().UTC(),
	}

	db := C.GetServices().Db
	err = db.Table("linkedin_capping_rules").Where("project_id = ? and id = ?", projectID, linkedinCappingRule.ID).
		Updates(updatedFields).Error
	if err != nil {
		if IsDuplicateRecordError(err) {
			logCtx.WithError(err).WithField("project_id", linkedinCappingRule.ProjectID).Warn(
				"Failed to create rule object. Duplicate.")
			return "Duplicate Rule", http.StatusConflict
		}
		logCtx.WithError(err).WithField("project_id", linkedinCappingRule.ProjectID).Error(
			"Failed to create rule object.")
		return "Internal server error", http.StatusInternalServerError
	}
	return "", http.StatusAccepted
}

func (store *MemSQL) DeleteLinkedinCappingRule(projectID int64, ruleID string) int {
	logFields := log.Fields{
		"project_id": projectID,
		"rule_id":    ruleID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db
	err := db.Table("linkedin_capping_rules").
		Where("project_id = ? AND id = ?", projectID, ruleID).
		Update("status", model.LINKEDIN_STATUS_DELETED).Error
	if err != nil {
		return http.StatusInternalServerError
	}
	return http.StatusAccepted
}

func (store *MemSQL) ApplyRuleOnLinkedinCappingData(projectID int64, linkedinCappingData model.LinkedinCappingDataSet,
	ruleWithDecodedValue model.LinkedinCappingRuleWithDecodedValues, existingGroupData model.GroupRelatedData) (bool, model.RuleMatchedDataSet, error) {

	rule := ruleWithDecodedValue.Rule
	advancedRules := ruleWithDecodedValue.AdvancedRules

	if linkedinCappingData.Impressions >= rule.ImpressionThreshold || linkedinCappingData.Clicks >= rule.ClickThreshold {
		return true, model.RuleMatchedDataSet{CappingData: linkedinCappingData, Rule: rule}, nil
	} else {
		isRuleMatched, propertiesMatched, err := store.checkIfAdvancedRulesApplicable(projectID, linkedinCappingData,
			advancedRules, existingGroupData)
		if err != nil {
			return false, model.RuleMatchedDataSet{}, err
		}
		if isRuleMatched {
			return true, model.RuleMatchedDataSet{CappingData: linkedinCappingData, PropertiesMatched: propertiesMatched, Rule: rule}, nil
		}
	}
	return false, model.RuleMatchedDataSet{}, nil
}

func (store *MemSQL) checkIfAdvancedRulesApplicable(projectID int64, linkedinCappingData model.LinkedinCappingDataSet,
	advancedRules []model.AdvancedRuleFilters, groupData model.GroupRelatedData) (bool, map[string]interface{}, error) {
	propertiesMatched := make(map[string]interface{})

	for _, rule := range advancedRules {
		if linkedinCappingData.Impressions < rule.ImpressionThreshold && linkedinCappingData.Clicks < rule.ClickThreshold {
			continue
		}
		segment := model.Query{
			GlobalUserProperties: rule.Filters,
		}
		isMatched, propertiesMatchedInRule := IsRuleMatchedAllAccountsForFrequencyCapping(projectID, segment, groupData.DecodedProps,
			groupData.GroupUsers, "", groupData.UserID, make(map[string]string), make(map[string]map[string]bool))

		if isMatched {
			propertiesMatched = propertiesMatchedInRule
			return isMatched, propertiesMatched, nil
		}
	}

	return false, propertiesMatched, nil
}

// [][] what does it mean.
// add a comment.
func IsRuleMatchedAllAccountsForFrequencyCapping(projectID int64, segment model.Query, decodedProperties []map[string]interface{}, userArr []model.User,
	segmentId string, domId string, eventNameIDsMap map[string]string, fileValuesMap map[string]map[string]bool) (bool, map[string]interface{}) {
	// isMatched = all rules matched (a or b) AND (c or d)
	isMatched := false
	propertiesMatched := make(map[string]interface{})

	groupedProperties := model.GetPropertiesGrouped(segment.GlobalUserProperties)
	for index, currentGroupedProperties := range groupedProperties {
		// validity for each group like (a or b ) AND (c or d)
		groupedPropsMatched := false
		for _, p := range currentGroupedProperties {
			isValueFound, propertiesMatchedInRule := CheckPropertyInAllUsersForFrequencyCapping(projectID, p, decodedProperties, userArr, fileValuesMap)
			if isValueFound {
				groupedPropsMatched = true
				propertiesMatched = propertiesMatchedInRule
				break
			}
		}
		if index == 0 {
			isMatched = groupedPropsMatched
			continue
		}
		isMatched = groupedPropsMatched && isMatched
	}

	return isMatched, propertiesMatched
}

func CheckPropertyInAllUsersForFrequencyCapping(projectId int64, p model.QueryProperty, decodedProperties []map[string]interface{},
	userArr []model.User, fileValuesMap map[string]map[string]bool) (bool, map[string]interface{}) {
	isValueFound := false
	for index, user := range userArr {
		// skip for group user if entity is user_group or entity is user_g and is not a group user
		if (p.Entity == model.PropertyEntityUserGroup && (user.IsGroupUser != nil && *user.IsGroupUser)) ||
			(p.Entity == model.PropertyEntityUserGlobal && (user.IsGroupUser == nil || !*user.IsGroupUser)) {
			continue
		}

		// group based filtering for engagement properties
		if p.GroupName == U.GROUP_NAME_DOMAINS && user.Source != nil && *user.Source != model.UserSourceDomains {
			continue
		}

		isValueFound = CheckPropertyOfGivenType(projectId, p, &decodedProperties[index], fileValuesMap)

		// check for negative filters
		if (p.Operator == model.NotContainsOpStr && p.Value != model.PropertyValueNone) ||
			(p.Operator == model.ContainsOpStr && p.Value == model.PropertyValueNone) ||
			(p.Operator == model.NotEqualOpStr && p.Value != model.PropertyValueNone) ||
			(p.Operator == model.EqualsOpStr && p.Value == model.PropertyValueNone) ||
			(p.Operator == model.NotInList && p.Value != model.PropertyValueNone) {
			if !isValueFound {
				return isValueFound, make(map[string]interface{})
			}
			continue
		}

		if isValueFound {
			return isValueFound, decodedProperties[index]
		}
	}
	return isValueFound, make(map[string]interface{})
}

func (store *MemSQL) GetGroupRelatedData(projectID int64, groupID int, domain string, existingGroupData map[string]model.GroupRelatedData) (model.GroupRelatedData, error) {
	if domain == "$none" || domain == "" {
		return model.GroupRelatedData{}, nil
	}
	if groupData, exists := existingGroupData[domain]; exists {
		return groupData, nil
	}
	user, status := store.GetGroupUserByGroupID(projectID, model.GROUP_NAME_DOMAINS, domain)
	if status != http.StatusFound && status != http.StatusNotFound {
		return model.GroupRelatedData{}, errors.New("failed to get group users - domain user get")
	} else if status != http.StatusFound {
		return model.GroupRelatedData{}, nil
	}
	groupUsers, errCode := store.GetAllGroupPropertiesForDomain(projectID, groupID, user.ID)
	if errCode != http.StatusFound && errCode != http.StatusNotFound {
		return model.GroupRelatedData{}, errors.New("failed to get group users - group users get")
	} else if status != http.StatusFound {
		return model.GroupRelatedData{}, nil
	}

	decodedPropsArr := make([]map[string]interface{}, 0)

	for _, user := range groupUsers {
		// decoding user properties col
		decodedProps, err := U.DecodePostgresJsonb(&user.Properties)
		if err != nil {
			log.WithFields(log.Fields{"project_id": projectID, "user_id": user.ID}).Error("Unable to decode user properties for user.")
			return model.GroupRelatedData{}, errors.New("failed to get group users - user properties decode")
		}
		decodedPropsArr = append(decodedPropsArr, *decodedProps)
	}
	return model.GroupRelatedData{UserID: user.ID, GroupUsers: groupUsers, DecodedProps: decodedPropsArr}, nil
}
