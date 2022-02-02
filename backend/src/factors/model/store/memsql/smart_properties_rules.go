package memsql

import (
	C "factors/config"
	Const "factors/constants"
	"factors/model/model"
	U "factors/util"
	"net/http"
	"strings"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

var sourceSmartProperty = map[string]bool{
	"all":      true,
	"adwords":  true,
	"facebook": true,
	"linkedin": true,
}

// to do: @ashhar make it similar to channels fields ASAP
var mapOfObjectAndProperty = map[string]map[string]map[string]model.PropertiesAndRelated{
	model.AdwordsCampaign: {
		model.AdwordsCampaign: {
			"name": model.PropertiesAndRelated{TypeOfProperty: U.PropertyTypeCategorical},
		},
	},
	model.AdwordsAdGroup: {
		model.AdwordsCampaign: {
			"name": model.PropertiesAndRelated{TypeOfProperty: U.PropertyTypeCategorical},
		},
		model.AdwordsAdGroup: {
			"name": model.PropertiesAndRelated{TypeOfProperty: U.PropertyTypeCategorical},
		},
	},
}

var smartPropertyObjects = []string{model.AdwordsCampaign, model.AdwordsAdGroup}
var smartPropertySources = []string{"all", "facebook", "adwords", "linkedin"}

func (store *MemSQL) satisfiesSmartPropertyRulesForeignConstraints(rule model.SmartPropertyRules) int {
	logFields := log.Fields{
		"rule": rule,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	// TODO: Add for project_id, user_id.
	_, errCode := store.GetProject(rule.ProjectID)
	if errCode != http.StatusFound {
		return http.StatusBadRequest
	}
	return http.StatusOK
}

func (store *MemSQL) GetSmartPropertyRulesConfig(projectID uint64, objectType string) (model.SmartPropertyRulesConfig, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"object_type": objectType,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	var result model.SmartPropertyRulesConfig
	sources := make([]model.Source, 0)
	objectAndProperty, isExists := mapOfObjectAndProperty[objectType]
	if !isExists {
		return result, http.StatusBadRequest
	}
	for _, sourceName := range smartPropertySources {
		objectsAndProperties := make([]model.ChannelObjectAndProperties, 0)
		for objectName, property := range objectAndProperty {
			currentProperties := buildProperties(property)
			objectsAndProperties = append(objectsAndProperties, buildObjectsAndProperties(currentProperties, []string{objectName})...)
		}
		var source model.Source
		source.Name = sourceName
		source.ObjectsAndProperties = objectsAndProperties
		sources = append(sources, source)
	}
	result.Name = objectType
	result.Sources = sources
	return result, http.StatusOK
}
func (store *MemSQL) checkIfRuleNameAlreadyPresentWhileCreate(projectID uint64, name string, objectType int) int {
	logFields := log.Fields{
		"project_id": projectID,
		"name": name,
		"object_type": objectType,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db
	smartPropertyRules := make([]model.SmartPropertyRules, 0)
	err := db.Model(&model.SmartPropertyRules{}).
		Where("project_id = ? AND is_deleted != ? AND name = ? AND type = ?", projectID, true, name, objectType).
		Find(&smartPropertyRules).Error
	if err != nil || len(smartPropertyRules) == 0 {
		return http.StatusNotFound
	}
	return http.StatusFound
}
func (store *MemSQL) checkIfRuleNameAlreadyPresentWhileUpdate(projectID uint64, name string, ruleID string, objectType int) int {
	logFields := log.Fields{
		"project_id": projectID,
		"name": name,
		"object_type": objectType,
		"rule_id": ruleID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db
	smartPropertyRules := make([]model.SmartPropertyRules, 0)
	err := db.Model(&model.SmartPropertyRules{}).
		Where("project_id = ? AND is_deleted != ? AND name = ? AND id != ? AND type = ?", projectID, true, name, ruleID, objectType).
		Find(&smartPropertyRules).Error
	if err != nil || len(smartPropertyRules) == 0 {
		return http.StatusNotFound
	}
	return http.StatusFound
}
func validateSmartPropertyRules(projectID uint64, smartPropertyRulesDoc *model.SmartPropertyRules) (string, bool) {
	logFields := log.Fields{
		"project_id": projectID,
		"smart_property_rules_doc": smartPropertyRulesDoc,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if projectID == 0 {
		return "Invalid project ID.", false
	}

	if smartPropertyRulesDoc.Name == "" {
		return "Empty name for rule.", false
	}
	if Const.SmartPropertyReservedNames[strings.ToLower(smartPropertyRulesDoc.Name)] {
		return "Entered Name is not allowed.", false
	}
	if strings.Contains(smartPropertyRulesDoc.Name, " ") {
		return "Space in property name is not allowed.", false
	}

	isValidRules := validationRules(smartPropertyRulesDoc.Rules)
	if !isValidRules {
		return "Invalid rule conditions or empty rules object.", false
	}
	return "", true
}

func (store *MemSQL) CreateSmartPropertyRules(projectID uint64, smartPropertyRulesDoc *model.SmartPropertyRules) (*model.SmartPropertyRules, string, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"smart_property_rules_doc": smartPropertyRulesDoc,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	errMsg, isValidRule := validateSmartPropertyRules(projectID, smartPropertyRulesDoc)
	if !isValidRule {
		logCtx.WithField("rule", smartPropertyRulesDoc).Warn(errMsg)
		return &model.SmartPropertyRules{}, errMsg, http.StatusBadRequest
	}

	logCtx = logCtx.WithField("type_alias", smartPropertyRulesDoc.TypeAlias)
	objectType, typeExists := model.SmartPropertyRulesTypeAliasToType[smartPropertyRulesDoc.TypeAlias]
	if !typeExists {
		logCtx.WithField("rule", smartPropertyRulesDoc).Warn("Invalid type alias.")
		return &model.SmartPropertyRules{}, "Invalid type alias.", http.StatusBadRequest
	}

	errCode := store.checkIfRuleNameAlreadyPresentWhileCreate(projectID, smartPropertyRulesDoc.Name, objectType)
	if errCode == http.StatusFound {
		return &model.SmartPropertyRules{}, "Name already present.", http.StatusBadRequest
	}
	smartPropertyRule := model.SmartPropertyRules{
		ID:          U.GetUUID(),
		ProjectID:   projectID,
		Type:        objectType,
		Name:        smartPropertyRulesDoc.Name,
		Description: smartPropertyRulesDoc.Description,
		Rules:       smartPropertyRulesDoc.Rules,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	if errCode := store.satisfiesSmartPropertyRulesForeignConstraints(smartPropertyRule); errCode != http.StatusOK {
		return &model.SmartPropertyRules{}, "Foreign constraints violated", http.StatusInternalServerError
	}

	db := C.GetServices().Db
	err := db.Create(&smartPropertyRule).Error
	if err != nil {
		if IsDuplicateRecordError(err) {
			logCtx.WithError(err).WithField("project_id", smartPropertyRulesDoc.ProjectID).Warn(
				"Failed to create rule object. Duplicate.")
			return &model.SmartPropertyRules{}, "Duplicate Rule", http.StatusConflict
		}
		logCtx.WithError(err).WithField("project_id", smartPropertyRulesDoc.ProjectID).Error(
			"Failed to create rule object.")
		return &model.SmartPropertyRules{}, "Internal server error", http.StatusInternalServerError
	}
	objectTypeAlias, typeAliasExists := model.SmartPropertyRulesTypeToTypeAlias[smartPropertyRule.Type]
	if !typeAliasExists {
		logCtx.WithField("rule", smartPropertyRulesDoc).Warn("Invalid type")
		return &model.SmartPropertyRules{}, "Invalid type return from db.", http.StatusBadRequest
	}
	smartPropertyRule.TypeAlias = objectTypeAlias
	return &smartPropertyRule, "", http.StatusCreated
}
func (store *MemSQL) UpdateSmartPropertyRules(projectID uint64, ruleID string, smartPropertyRulesDoc model.SmartPropertyRules) (model.SmartPropertyRules, string, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"smart_property_rules_doc": smartPropertyRulesDoc,
		"rule_id": ruleID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	errMsg, isValidRule := validateSmartPropertyRules(projectID, &smartPropertyRulesDoc)
	if !isValidRule {
		logCtx.WithField("rule", smartPropertyRulesDoc).Warn(errMsg)
		return model.SmartPropertyRules{}, errMsg, http.StatusBadRequest
	}

	logCtx = logCtx.WithField("type_alias", smartPropertyRulesDoc.TypeAlias)
	objectType, typeExists := model.SmartPropertyRulesTypeAliasToType[smartPropertyRulesDoc.TypeAlias]
	if !typeExists {
		logCtx.WithField("rule", smartPropertyRulesDoc).Warn("Invalid type alias.")
		return model.SmartPropertyRules{}, "Invalid type alias.", http.StatusBadRequest
	}
	errCode := store.checkIfRuleNameAlreadyPresentWhileUpdate(projectID, smartPropertyRulesDoc.Name, ruleID, objectType)
	if errCode == http.StatusFound && !smartPropertyRulesDoc.IsDeleted {
		return model.SmartPropertyRules{}, "Name already present.", http.StatusBadRequest
	}
	updatedFields := map[string]interface{}{
		"rules":             smartPropertyRulesDoc.Rules,
		"type":              objectType,
		"evaluation_status": smartPropertyRulesDoc.EvaluationStatus,
		"name":              smartPropertyRulesDoc.Name,
		"description":       smartPropertyRulesDoc.Description,
		"updated_at":        time.Now().UTC(),
	}

	db := C.GetServices().Db
	err := db.Table("smart_property_rules").Where("project_id = ? AND id = ?", projectID, ruleID).Updates(updatedFields).Error
	if err != nil {
		if IsDuplicateRecordError(err) {
			logCtx.WithError(err).WithField("project_id", smartPropertyRulesDoc.ProjectID).Warn(
				"Failed to update rule object. Duplicate.")
			return model.SmartPropertyRules{}, "Duplicate Rule", http.StatusConflict
		}
		logCtx.WithError(err).WithField("project_id", smartPropertyRulesDoc.ProjectID).Error(
			"Failed to update rule object.")
		return model.SmartPropertyRules{}, "Internal server error", http.StatusInternalServerError
	}
	smartPropertyRule, errCode := store.GetSmartPropertyRule(projectID, ruleID)
	if errCode != http.StatusFound {
		return model.SmartPropertyRules{}, "", http.StatusInternalServerError
	}
	return smartPropertyRule, "", http.StatusAccepted
}

func validationRules(rulesJsonb *postgres.Jsonb) bool {
	logFields := log.Fields{
		"rules_jsonb": rulesJsonb,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	var rules []model.Rule
	err := U.DecodePostgresJsonbToStructType(rulesJsonb, &rules)
	if err != nil {
		return false
	}
	if len(rules) == 0 {
		return false
	}
	for _, rule := range rules {
		if rule.Value == "" {
			return false
		}
		_, existsSource := sourceSmartProperty[rule.Source]
		if !existsSource {
			return false
		}
		if len(rule.Filters) == 0 {
			return false
		}
		for _, filter := range rule.Filters {
			if filter.Value == "" {
				return false
			}
			_, objectTypeExists := model.SmartPropertyRulesTypeAliasToType[filter.Object]
			if !objectTypeExists {
				return false
			}
		}
	}
	return true
}

func (store *MemSQL) GetSmartPropertyRules(projectID uint64) ([]model.SmartPropertyRules, int) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	smartPropertyRules := make([]model.SmartPropertyRules, 0)
	if projectID == 0 {
		log.Error("Invalid project ID.")
		return make([]model.SmartPropertyRules, 0), http.StatusBadRequest
	}
	db := C.GetServices().Db
	err := db.Table("smart_property_rules").Where("project_id = ? AND is_deleted != ?", projectID, true).Find(&smartPropertyRules).Error
	if err != nil {
		log.WithField("project_id", projectID).Warn(err)
		return make([]model.SmartPropertyRules, 0), http.StatusNotFound
	}
	for index, smartPropertyRule := range smartPropertyRules {
		objectTypeAlias, typeAliasExists := model.SmartPropertyRulesTypeToTypeAlias[smartPropertyRule.Type]
		if !typeAliasExists {
			return []model.SmartPropertyRules{}, http.StatusBadRequest
		}
		smartPropertyRules[index].TypeAlias = objectTypeAlias
	}
	return smartPropertyRules, http.StatusFound
}

func (store *MemSQL) GetAllChangedSmartPropertyRulesForProject(projectID uint64) ([]model.SmartPropertyRules, int) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	smartPropertyRules := make([]model.SmartPropertyRules, 0)
	db := C.GetServices().Db
	err := db.Table("smart_property_rules").Where("project_id = ? AND evaluation_status != ?", projectID, model.EvaluationStatusMap["picked"]).Order("updated_at asc").Find(&smartPropertyRules).Error
	if err != nil {
		log.WithField("project_id", projectID).Warn(err)
		return make([]model.SmartPropertyRules, 0), http.StatusNotFound
	}
	return smartPropertyRules, http.StatusFound
}

func (store *MemSQL) GetSmartPropertyRule(projectID uint64, ruleID string) (model.SmartPropertyRules, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"rule_id": ruleID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	var smartPropertyRule model.SmartPropertyRules
	if projectID == 0 {
		log.Error("Invalid project ID.")
		return model.SmartPropertyRules{}, http.StatusBadRequest
	}
	if ruleID == "" {
		log.Error("Invalid rule ID.")
		return model.SmartPropertyRules{}, http.StatusBadRequest
	}
	db := C.GetServices().Db
	err := db.Table("smart_property_rules").Where("project_id = ? AND is_deleted != ? AND id = ?", projectID, true, ruleID).Find(&smartPropertyRule).Error
	if err != nil {
		log.WithField("project_id", projectID).Warn(err)
		return model.SmartPropertyRules{}, http.StatusNotFound
	}
	objectTypeAlias, typeAliasExists := model.SmartPropertyRulesTypeToTypeAlias[smartPropertyRule.Type]
	if !typeAliasExists {
		return model.SmartPropertyRules{}, http.StatusBadRequest
	}
	smartPropertyRule.TypeAlias = objectTypeAlias

	return smartPropertyRule, http.StatusFound
}

func (store *MemSQL) DeleteSmartPropertyRules(projectID uint64, ruleID string) int {
	logFields := log.Fields{
		"project_id": projectID,
		"rule_id": ruleID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if projectID == 0 {
		log.Error("Invalid project ID.")
		return http.StatusBadRequest
	}
	if ruleID == "" {
		log.Error("Invalid rule id")
		return http.StatusBadRequest
	}
	db := C.GetServices().Db
	err := db.Table("smart_property_rules").Where("project_id = ? AND id = ?", projectID, ruleID).Updates(map[string]interface{}{"is_deleted": true, "evaluation_status": model.EvaluationStatusMap["not_picked"], "updated_at": time.Now().UTC()}).Error
	if err != nil {
		log.WithField("project_id", projectID).Error(err)
		return http.StatusInternalServerError
	}
	return http.StatusAccepted
}

func (store *MemSQL) GetProjectIDsHavingSmartPropertyRules() ([]uint64, int) {
	defer model.LogOnSlowExecutionWithParams(time.Now(), nil)
	db := C.GetServices().Db
	var projectIDs []uint64
	rows, err := db.Table("smart_property_rules").Select("DISTINCT(project_id)").Rows()
	if err != nil {
		return make([]uint64, 0), http.StatusInternalServerError
	}
	for rows.Next() {
		var projectID uint64
		err = rows.Scan(&projectID)
		if err != nil {
			return make([]uint64, 0), http.StatusInternalServerError
		}
		projectIDs = append(projectIDs, projectID)
	}
	return projectIDs, http.StatusFound
}
