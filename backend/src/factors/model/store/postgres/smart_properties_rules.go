package postgres

import (
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"net/http"
	"strings"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

var smartPropertiesRulesTypeAlias = map[string]int{
	"campaign": 1,
	"ad_group": 2,
}
var sourceSmartProperties = map[string]bool{
	"all":      true,
	"adwords":  true,
	"facebook": true,
	"linkedin": true,
}
var smartPropertiesDisallowedNames = map[string]bool{
	"campaign":                       true,
	"ad_group":                       true,
	"ad_set":                         true,
	"adset":                          true,
	"creative":                       true,
	"campaign_group":                 true,
	"ad":                             true,
	"name":                           true,
	"keyword":                        true,
	"id":                             true,
	"status":                         true,
	approvalStatus:                   true,
	matchType:                        true,
	firstPositionCpc:                 true,
	firstPageCpc:                     true,
	isNegative:                       true,
	topOfPageCpc:                     true,
	qualityScore:                     true,
	impressions:                      true,
	clicks:                           true,
	"spend":                          true,
	conversion:                       true,
	clickThroughRate:                 true,
	conversionRate:                   true,
	costPerClick:                     true,
	costPerConversion:                true,
	searchImpressionShare:            true,
	searchClickShare:                 true,
	searchTopImpressionShare:         true,
	searchAbsoluteTopImpressionShare: true,
	searchBudgetLostAbsoluteTopImpressionShare: true,
	searchBudgetLostImpressionShare:            true,
	searchBudgetLostTopImpressionShare:         true,
	searchRankLostAbsoluteTopImpressionShare:   true,
	searchRankLostImpressionShare:              true,
	searchRankLostTopImpressionShare:           true,
}

// to do: @ashhar make it similar to channels fields ASAP
var mapOfObjectAndProperty = map[string]map[string]map[string]PropertiesAndRelated{
	adwordsCampaign: {
		adwordsCampaign: {
			"name": PropertiesAndRelated{typeOfProperty: U.PropertyTypeCategorical},
		},
	},
	adwordsAdGroup: {
		adwordsCampaign: {
			"name": PropertiesAndRelated{typeOfProperty: U.PropertyTypeCategorical},
		},
		adwordsAdGroup: {
			"name": PropertiesAndRelated{typeOfProperty: U.PropertyTypeCategorical},
		},
	},
}

var smartPropertiesObjects = []string{adwordsCampaign, adwordsAdGroup}
var smartPropertiesSources = []string{"all", "facebook", "adwords", "linkedin"}

const errorDuplicateSmartPropertiesRules = "pq: duplicate key value violates unique constraint \"smart_properties_rules_primary_key\""

func isDuplicateSmartPropertiesRulesError(err error) bool {
	return err.Error() == errorDuplicateSmartPropertiesRules
}

func (pg *Postgres) GetSmartPropertiesRulesConfig(projectID uint64, objectType string) (model.SmartPropertiesRulesConfig, int) {
	var result model.SmartPropertiesRulesConfig
	sources := make([]model.Source, 0, 0)
	objectAndProperty, isExists := mapOfObjectAndProperty[objectType]
	if !isExists {
		return result, http.StatusBadRequest
	}
	for _, sourceName := range smartPropertiesSources {
		objectsAndProperties := make([]model.ChannelObjectAndProperties, 0, 0)
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
func (pg *Postgres) checkIfRuleNameAlreadyPresentWhileCreate(projectID uint64, name string, objectType int) int {
	db := C.GetServices().Db
	smartPropertiesRules := make([]model.SmartPropertiesRules, 0, 0)
	err := db.Model(&model.SmartPropertiesRules{}).
		Where("project_id = ? AND is_deleted != ? AND name = ? AND type = ?", projectID, true, name, objectType).
		Find(&smartPropertiesRules).Error
	if err != nil || len(smartPropertiesRules) == 0 {
		return http.StatusNotFound
	}
	return http.StatusFound
}
func (pg *Postgres) checkIfRuleNameAlreadyPresentWhileUpdate(projectID uint64, name string, ruleID string, objectType int) int {
	db := C.GetServices().Db
	smartPropertiesRules := make([]model.SmartPropertiesRules, 0, 0)
	err := db.Model(&model.SmartPropertiesRules{}).
		Where("project_id = ? AND is_deleted != ? AND name = ? AND id != ? AND type = ?", projectID, true, name, ruleID, objectType).
		Find(&smartPropertiesRules).Error
	if err != nil || len(smartPropertiesRules) == 0 {
		return http.StatusNotFound
	}
	return http.StatusFound
}
func validateSmartPropertiesRules(projectID uint64, smartPropertiesRulesDoc *model.SmartPropertiesRules) (string, bool) {
	if projectID == 0 {
		return "Invalid project ID.", false
	}

	if smartPropertiesRulesDoc.Name == "" {
		return "Empty name for rule.", false
	}
	if smartPropertiesDisallowedNames[strings.ToLower(smartPropertiesRulesDoc.Name)] {
		return "Entered Name is not allowed.", false
	}
	if strings.Contains(smartPropertiesRulesDoc.Name, " ") {
		return "Space in property name is not allowed.", false
	}

	isValidRules := validationRules(smartPropertiesRulesDoc.Rules)
	if !isValidRules {
		return "Invalid rule conditions or empty rules object.", false
	}
	return "", true
}

func (pg *Postgres) CreateSmartPropertiesRules(projectID uint64, smartPropertiesRulesDoc *model.SmartPropertiesRules) (string, int) {
	logCtx := log.WithField("project_id", smartPropertiesRulesDoc.ProjectID)

	errMsg, isValidRule := validateSmartPropertiesRules(projectID, smartPropertiesRulesDoc)
	if !isValidRule {
		logCtx.Error(errMsg)
		return errMsg, http.StatusBadRequest
	}

	logCtx = logCtx.WithField("type_alias", smartPropertiesRulesDoc.TypeAlias)
	objectType, typeExists := smartPropertiesRulesTypeAlias[smartPropertiesRulesDoc.TypeAlias]
	if !typeExists {
		logCtx.Error("Invalid type alias.")
		return "Invalid type alias.", http.StatusBadRequest
	}
	errCode := pg.checkIfRuleNameAlreadyPresentWhileCreate(projectID, smartPropertiesRulesDoc.Name, objectType)
	if errCode == http.StatusFound {
		return "Name already present.", http.StatusBadRequest
	}
	smartPropertiesRulesDoc.Type = objectType
	smartPropertiesRulesDoc.Picked = false
	db := C.GetServices().Db
	err := db.Create(&smartPropertiesRulesDoc).Error
	if err != nil {
		if isDuplicateSmartPropertiesRulesError(err) {
			logCtx.WithError(err).WithField("project_id", smartPropertiesRulesDoc.ProjectID).Warn(
				"Failed to create rule object. Duplicate.")
			return "Duplicate Rule", http.StatusConflict
		}
		logCtx.WithError(err).WithField("project_id", smartPropertiesRulesDoc.ProjectID).Error(
			"Failed to create rule object.")
		return "Internal server error", http.StatusInternalServerError
	}
	return "", http.StatusCreated
}
func (pg *Postgres) UpdateSmartPropertiesRules(projectID uint64, ruleID string, smartPropertiesRulesDoc model.SmartPropertiesRules) (string, int) {
	logCtx := log.WithField("project_id", smartPropertiesRulesDoc.ProjectID)

	errMsg, isValidRule := validateSmartPropertiesRules(projectID, &smartPropertiesRulesDoc)
	if !isValidRule {
		logCtx.Error(errMsg)
		return errMsg, http.StatusBadRequest
	}

	logCtx = logCtx.WithField("type_alias", smartPropertiesRulesDoc.TypeAlias)
	objectType, typeExists := smartPropertiesRulesTypeAlias[smartPropertiesRulesDoc.TypeAlias]
	if !typeExists {
		logCtx.Error("Invalid type alias.")
		return "Invalid type alias.", http.StatusBadRequest
	}
	errCode := pg.checkIfRuleNameAlreadyPresentWhileUpdate(projectID, smartPropertiesRulesDoc.Name, ruleID, objectType)
	if errCode == http.StatusFound {
		return "Name already present.", http.StatusBadRequest
	}
	updatedFields := map[string]interface{}{
		"rules":       smartPropertiesRulesDoc.Rules,
		"type":        objectType,
		"picked":      smartPropertiesRulesDoc.Picked,
		"name":        smartPropertiesRulesDoc.Name,
		"description": smartPropertiesRulesDoc.Description,
		"updated_at":  time.Now().UTC(),
	}

	db := C.GetServices().Db
	err := db.Table("smart_properties_rules").Where("project_id = ? AND id = ?", projectID, ruleID).Updates(updatedFields).Error
	if err != nil {
		if isDuplicateSmartPropertiesRulesError(err) {
			logCtx.WithError(err).WithField("project_id", smartPropertiesRulesDoc.ProjectID).Warn(
				"Failed to update rule object. Duplicate.")
			return "Duplicate Rule", http.StatusConflict
		}
		logCtx.WithError(err).WithField("project_id", smartPropertiesRulesDoc.ProjectID).Error(
			"Failed to update rule object.")
		return "Internal server error", http.StatusInternalServerError
	}
	return "", http.StatusAccepted
}

func validationRules(rulesJsonb *postgres.Jsonb) bool {
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
		_, existsSource := sourceSmartProperties[rule.Source]
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
			_, objectTypeExists := smartPropertiesRulesTypeAlias[filter.Object]
			if !objectTypeExists {
				return false
			}
		}
	}
	return true
}

func (pg *Postgres) GetSmartPropertiesRules(projectID uint64) ([]model.SmartPropertiesRules, int) {
	smartPropertiesRules := make([]model.SmartPropertiesRules, 0, 0)
	if projectID == 0 {
		log.Error("Invalid project ID.")
		return make([]model.SmartPropertiesRules, 0, 0), http.StatusBadRequest
	}
	db := C.GetServices().Db
	err := db.Table("smart_properties_rules").Where("project_id = ? AND is_deleted != ?", projectID, true).Find(&smartPropertiesRules).Error
	if err != nil {
		log.WithField("project_id", projectID).Error(err)
		return make([]model.SmartPropertiesRules, 0, 0), http.StatusNotFound
	}
	return smartPropertiesRules, http.StatusFound
}

func (pg *Postgres) GetAllChangedSmartPropertiesRulesForProject(projectID uint64) ([]model.SmartPropertiesRules, int) {
	smartPropertiesRules := make([]model.SmartPropertiesRules, 0, 0)
	db := C.GetServices().Db
	err := db.Table("smart_properties_rules").Where("project_id = ? AND picked = ?", projectID, false).Find(&smartPropertiesRules).Error
	if err != nil {
		log.Error(err)
		return make([]model.SmartPropertiesRules, 0, 0), http.StatusNotFound
	}
	return smartPropertiesRules, http.StatusFound
}

func (pg *Postgres) GetSmartPropertiesRule(projectID uint64, ruleID string) (model.SmartPropertiesRules, int) {
	var smartPropertiesRule model.SmartPropertiesRules
	if projectID == 0 {
		log.Error("Invalid project ID.")
		return model.SmartPropertiesRules{}, http.StatusBadRequest
	}
	if ruleID == "" {
		log.Error("Invalid rule ID.")
		return model.SmartPropertiesRules{}, http.StatusBadRequest
	}
	db := C.GetServices().Db
	err := db.Table("smart_properties_rules").Where("project_id = ? AND is_deleted != ? AND id = ?", projectID, true, ruleID).Find(&smartPropertiesRule).Error
	if err != nil {
		log.WithField("project_id", projectID).Error(err)
		return model.SmartPropertiesRules{}, http.StatusNotFound
	}
	return smartPropertiesRule, http.StatusFound
}

func (pg *Postgres) DeleteSmartPropertiesRules(projectID uint64, ruleID string) int {
	if projectID == 0 {
		log.Error("Invalid project ID.")
		return http.StatusBadRequest
	}
	if ruleID == "" {
		log.Error("Invalid rule id")
		return http.StatusBadRequest
	}
	db := C.GetServices().Db
	err := db.Table("smart_properties_rules").Where("project_id = ? AND id = ?", projectID, ruleID).Updates(map[string]interface{}{"is_deleted": true, "picked": false, "updated_at": time.Now().UTC()}).Error
	if err != nil {
		log.WithField("project_id", projectID).Error(err)
		return http.StatusInternalServerError
	}
	return http.StatusAccepted
}

func (pg *Postgres) GetProjectIDsHavingSmartPropertiesRules() ([]uint64, int) {
	db := C.GetServices().Db
	var projectIDs []uint64
	rows, err := db.Table("smart_properties_rules").Select("DISTINCT(project_id)").Rows()
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
