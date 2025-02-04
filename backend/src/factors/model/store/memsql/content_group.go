package memsql

import (
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"net/http"
	"reflect"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

const contentGroupsLimt = 3

func (store *MemSQL) DeleteContentGroup(id string, projectID int64) (int, string) {
	logFields := log.Fields{
		"id":         id,
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if projectID == 0 {
		log.Error("Invalid project ID.")
		return http.StatusBadRequest, "Invalid project id"
	}
	if id == "" {
		log.Error("Invalid id")
		return http.StatusBadRequest, "Invalid id"
	}
	db := C.GetServices().Db
	err := db.Table("content_groups").Where("project_id = ? AND id = ?", projectID, id).Updates(map[string]interface{}{"is_deleted": true, "updated_at": time.Now().UTC()}).Error
	if err != nil {
		log.WithField("project_id", projectID).Error(err)
		return http.StatusInternalServerError, err.Error()
	}
	return http.StatusAccepted, ""
}

func (store *MemSQL) GetContentGroupById(id string, projectID int64) (model.ContentGroup, int) {
	logFields := log.Fields{
		"id":         id,
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	var contentGroup model.ContentGroup
	if projectID == 0 {
		log.Error("Invalid project ID.")
		return model.ContentGroup{}, http.StatusBadRequest
	}
	if id == "" {
		log.Error("Invalid rule ID.")
		return model.ContentGroup{}, http.StatusBadRequest
	}
	db := C.GetServices().Db
	err := db.Table("content_groups").Where("project_id = ? AND is_deleted != ? AND id = ?", projectID, true, id).Find(&contentGroup).Error
	if err != nil {
		log.WithField("project_id", projectID).Warn(err)
		return model.ContentGroup{}, http.StatusNotFound
	}

	return contentGroup, http.StatusFound
}

func (store *MemSQL) GetAllContentGroups(projectID int64) ([]model.ContentGroup, int) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	contentGroups := make([]model.ContentGroup, 0)
	if projectID == 0 {
		log.Error("Invalid project ID.")
		return make([]model.ContentGroup, 0), http.StatusBadRequest
	}
	db := C.GetServices().Db
	err := db.Table("content_groups").Where("project_id = ? AND is_deleted != ?", projectID, true).Find(&contentGroups).Error
	if err != nil {
		log.WithField("project_id", projectID).Warn(err)
		return make([]model.ContentGroup, 0), http.StatusNotFound
	}
	return contentGroups, http.StatusFound
}

func (store *MemSQL) CreateContentGroup(projectID int64, contentGroup model.ContentGroup) (model.ContentGroup, int, string) {
	logFields := log.Fields{
		"project_id":    projectID,
		"content_group": contentGroup,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	// For a project the following validation should be followed
	// 1. distinct content group name
	// 2. distinct values inside a content group
	// 3. Rule should be distinct across all values
	// 4. max 3 content groups for a project
	// 5. no repeating filters inside the same rule - not handling for now
	// 6 n values inside a content group - not handling for now
	// 7 minimum one rule for a content group

	logCtx := log.WithFields(logFields)

	if store.IsDuplicateNameCheck(projectID, contentGroup.ContentGroupName) {
		logCtx.WithField("project_id", projectID).Error(
			"Duplicate Content Group Name")
		return model.ContentGroup{}, http.StatusBadRequest, "Duplicate Content Group Name"
	}
	validRule, errString := store.IsValidRule(contentGroup)
	if !validRule {
		logCtx.WithField("project_id", projectID).Error(
			"Invalid Rule " + errString)
		return model.ContentGroup{}, http.StatusBadRequest, "Invalid Rule " + errString
	}
	if !store.ContentGroupLimitCheck(projectID) {
		logCtx.WithField("project_id", projectID).Error(
			"Limit Exceeded")
		return model.ContentGroup{}, http.StatusBadRequest, "Limit Exceeded"
	}
	contentGroupRecord := model.ContentGroup{
		ID:                      U.GetUUID(),
		ProjectID:               projectID,
		ContentGroupName:        contentGroup.ContentGroupName,
		ContentGroupDescription: contentGroup.ContentGroupDescription,
		Rule:                    contentGroup.Rule,
		CreatedBy:               contentGroup.CreatedBy,
		CreatedAt:               time.Now().UTC(),
		UpdatedAt:               time.Now().UTC(),
		IsDeleted:               false,
	}
	db := C.GetServices().Db
	err := db.Create(&contentGroupRecord).Error
	if err != nil {
		logCtx.WithError(err).WithField("project_id", contentGroupRecord.ProjectID).Error(
			"Failed to create rule object.")
		return model.ContentGroup{}, http.StatusInternalServerError, "Internal server error"
	}
	return contentGroupRecord, http.StatusCreated, ""
}

func (store *MemSQL) IsDuplicateNameCheck(projectID int64, name string) bool {
	logFields := log.Fields{
		"project_id": projectID,
		"name":       name,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	contentGroups, _ := store.GetAllContentGroups(projectID)
	for _, contentGroup := range contentGroups {
		if contentGroup.ContentGroupName == name {
			return true
		}
	}
	return false
}

func (store *MemSQL) IsValidRule(contentGroup model.ContentGroup) (bool, string) {
	logFields := log.Fields{
		"content_group": contentGroup,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	filterConditions := map[string]bool{
		model.EqualsOpStr:      true,
		model.NotEqualOpStr:    true,
		model.ContainsOpStr:    true,
		model.NotContainsOpStr: true,
		model.StartsWith:       true,
		model.EndsWith:         true,
	}
	var contentGroupRule []model.ContentGroupRule
	err := U.DecodePostgresJsonbToStructType(contentGroup.Rule, &contentGroupRule)

	if model.CheckIfPropertyIsPresentInStaticKPIPropertyList(contentGroup.ContentGroupName) {
		return false, "Content Group name is conflicting with KPI property names."
	}
	if len(contentGroupRule) == 0 {
		return false, "Minimum one value required"
	}
	ruleValuesName := make(map[string]bool)
	rules := make([]model.ContentGroupRuleFilters, 0)
	if err == nil {
		for _, rule := range contentGroupRule {
			if ruleValuesName[rule.ContentGroupValue] == true {
				return false, "Duplicate Value Names"
			}
			ruleValuesName[rule.ContentGroupValue] = true
			rules = append(rules, rule.Rule)
		}
		for i := 0; i < len(rules)-1; i++ {
			for j := i + 1; j < len(rules); j++ {
				if reflect.DeepEqual(rules[i], rules[j]) {
					return false, "Duplicate Filters"
				}
			}
		}
		for _, filters := range rules {
			for _, rule := range filters {
				if !(rule.LogicalOp == "OR" || rule.LogicalOp == "AND") {
					return false, "Invalid Logical operator"
				}
				if !filterConditions[rule.Operator] == true {
					return false, "Invalid filter operator"
				}
			}
		}
	} else {
		return false, "Rule parsing error"
	}
	return true, ""
}

func (store *MemSQL) ContentGroupLimitCheck(projectID int64) bool {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	contentGroups, _ := store.GetAllContentGroups(projectID)
	if len(contentGroups) < contentGroupsLimt {
		return true
	}
	return false
}

func (store *MemSQL) UpdateContentGroup(id string, projectID int64, contentGroup model.ContentGroup) (model.ContentGroup, int, string) {
	logFields := log.Fields{
		"id":            id,
		"project_id":    projectID,
		"content_group": contentGroup,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	// only rule/description can be updated.
	// New value inserts in a rule
	// deletion of a value in a rule
	// edit a value in a rule
	// change order
	// validate the rule again during update

	logCtx := log.WithFields(logFields)

	validRule, errString := store.IsValidRule(contentGroup)
	if !validRule {
		logCtx.WithField("project_id", projectID).Error(
			"Invalid Rule " + errString)
		return model.ContentGroup{}, http.StatusBadRequest, "Invalid Rule " + errString
	}
	updatedFields := map[string]interface{}{
		"rule":                      contentGroup.Rule,
		"content_group_description": contentGroup.ContentGroupDescription,
		"updated_at":                time.Now().UTC(),
	}

	db := C.GetServices().Db
	err := db.Table("content_groups").Where("project_id = ? AND id = ?", projectID, id).Updates(updatedFields).Error
	if err != nil {
		logCtx.WithError(err).WithField("project_id", projectID).Error(
			"Failed to update rule object.")
		return model.ContentGroup{}, http.StatusInternalServerError, "Internal server error"
	}
	return contentGroup, http.StatusAccepted, ""
}

func (store *MemSQL) CheckURLContentGroupValue(pageUrl string, projectID int64) map[string]string {
	logFields := log.Fields{
		"page_url":   pageUrl,
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	contentGroups, _ := store.GetAllContentGroups(projectID)
	resultCg := make(map[string]string)
	for _, contentGroup := range contentGroups {
		var contentGroupRule []model.ContentGroupRule
		err := U.DecodePostgresJsonbToStructType(contentGroup.Rule, &contentGroupRule)
		if err != nil {
			return nil
		}
		for _, rule := range contentGroupRule {
			var overallResult bool
			results := make([]bool, 0)
			for _, filter := range rule.Rule {
				pageUrlCaseInsensitive := strings.ToLower(pageUrl)
				fitlerValue := strings.ToLower(filter.Value)
				if filter.Operator == model.EqualsOpStr {
					results = append(results, (pageUrlCaseInsensitive == fitlerValue))
				}
				if filter.Operator == model.NotEqualOpStr {
					results = append(results, (pageUrlCaseInsensitive != fitlerValue))
				}
				if filter.Operator == model.ContainsOpStr {
					results = append(results, strings.Contains(pageUrlCaseInsensitive, fitlerValue))
				}
				if filter.Operator == model.NotContainsOpStr {
					results = append(results, !strings.Contains(pageUrlCaseInsensitive, fitlerValue))
				}
				if filter.Operator == model.StartsWith {
					results = append(results, strings.HasPrefix(pageUrlCaseInsensitive, fitlerValue))
				}
				if filter.Operator == model.EndsWith {
					results = append(results, strings.HasSuffix(pageUrlCaseInsensitive, fitlerValue))
				}
			}
			if rule.Rule[0].LogicalOp == model.LOGICAL_OP_OR {
				soFar := false
				for _, result := range results {
					soFar = soFar || result

				}
				overallResult = soFar
			}
			if rule.Rule[0].LogicalOp == model.LOGICAL_OP_AND {
				soFar := true
				for _, result := range results {
					soFar = soFar && result

				}
				overallResult = soFar
			}
			if overallResult == true {
				resultCg[contentGroup.ContentGroupName] = rule.ContentGroupValue
				break
			}
		}
	}
	return resultCg
}
