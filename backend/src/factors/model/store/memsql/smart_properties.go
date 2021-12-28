package memsql

import (
	"encoding/json"
	C "factors/config"
	Const "factors/constants"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) satisfiesSmartPropertyForeignConstraints(properties model.SmartProperties) int {
	_, errCode := store.GetProject(properties.ProjectID)
	if errCode != http.StatusFound {
		return http.StatusBadRequest
	}
	return http.StatusOK
}

func (store *MemSQL) CreateSmartProperty(smartPropertyDoc *model.SmartProperties) int {
	logCtx := log.WithField("project_id", smartPropertyDoc.ProjectID)

	if smartPropertyDoc.ProjectID == 0 {
		logCtx.Error("Invalid project ID.")
		return http.StatusBadRequest
	}

	if !C.IsDryRunSmartProperties() {
		if errCode := store.satisfiesSmartPropertyForeignConstraints(*smartPropertyDoc); errCode != http.StatusOK {
			return http.StatusInternalServerError
		}
		db := C.GetServices().Db
		err := db.Create(&smartPropertyDoc).Error
		if err != nil {
			if IsDuplicateRecordError(err) {
				logCtx.WithError(err).WithField("project_id", smartPropertyDoc.ProjectID).Error(
					"Failed to create smart properties. Duplicate.")
				return http.StatusConflict
			}
			logCtx.WithError(err).WithField("project_id", smartPropertyDoc.ProjectID).Error(
				"Failed to create smart properties.")
			return http.StatusInternalServerError
		}
	}
	return http.StatusCreated
}
func (store *MemSQL) UpdateSmartProperty(smartPropertyDoc *model.SmartProperties) int {
	logCtx := log.WithField("project_id", smartPropertyDoc.ProjectID)

	if smartPropertyDoc.ProjectID == 0 {
		logCtx.Error("Invalid project ID.")
		return http.StatusBadRequest
	}

	if !C.IsDryRunSmartProperties() {
		db := C.GetServices().Db
		err := db.Save(&smartPropertyDoc).Error
		if err != nil {
			if IsDuplicateRecordError(err) {
				logCtx.WithError(err).WithField("project_id", smartPropertyDoc.ProjectID).Error(
					"Failed to update smart properties. Duplicate.")
				return http.StatusConflict
			}
			logCtx.WithError(err).WithField("project_id", smartPropertyDoc.ProjectID).Error(
				"Failed to update smart properties.")
			return http.StatusInternalServerError
		}
	}
	return http.StatusAccepted
}
func (store *MemSQL) GetSmartPropertyByProjectIDAndSourceAndObjectType(projectID uint64, source string, objectType int) ([]model.SmartProperties, int) {
	db := C.GetServices().Db
	smartProperty := []model.SmartProperties{}
	err := db.Table("smart_properties").Where("project_id = ? AND source = ? AND object_type = ?", projectID, source, objectType).Find(&smartProperty).Error
	if err != nil && err.Error() == "record not found" {
		return smartProperty, http.StatusNotFound
	} else if err != nil && err.Error() != "record not found" {
		return smartProperty, http.StatusInternalServerError
	} else {
		return smartProperty, http.StatusFound
	}
}
func (store *MemSQL) DeleteSmartPropertyByProjectIDAndSourceAndObjectID(projectID uint64, source string, objectID string) int {
	if !C.IsDryRunSmartProperties() {
		db := C.GetServices().Db
		err := db.Table("smart_properties").Where("project_id = ? AND source = ? AND object_id = ?", projectID, source, objectID).Delete(&model.SmartProperties{}).Error
		if err != nil {
			return http.StatusInternalServerError
		}
	}
	return http.StatusAccepted
}

func (store *MemSQL) GetSmartPropertyByProjectIDAndObjectIDAndObjectType(projectID uint64, objectID string, objectType int) (model.SmartProperties, int) {
	db := C.GetServices().Db
	smartProperty := model.SmartProperties{}
	err := db.Table("smart_properties").Where("project_id = ? AND object_id = ? AND object_type = ?", projectID, objectID, objectType).Find(&smartProperty).Error
	if err != nil && err.Error() == "record not found" {
		return smartProperty, http.StatusNotFound
	} else if err != nil && err.Error() != "record not found" {
		return smartProperty, http.StatusInternalServerError
	} else {
		return smartProperty, http.StatusFound
	}
}
func (store *MemSQL) BuildAndCreateSmartPropertyFromChannelDocumentAndRule(smartPropertyRule *model.SmartPropertyRules, rule model.Rule,
	channelDocument model.ChannelDocumentsWithFields, source string) int {
	var objectID string
	if smartPropertyRule.Type == 1 {
		objectID = channelDocument.CampaignID
	} else {
		objectID = channelDocument.AdGroupID
	}
	objectPropertiesJson, err := json.Marshal(channelDocument)
	if err != nil {
		return http.StatusInternalServerError
	}
	objectPropertiesJsonb := &postgres.Jsonb{objectPropertiesJson}

	smartProperty, errCode := store.GetSmartPropertyByProjectIDAndObjectIDAndObjectType(smartPropertyRule.ProjectID, objectID, smartPropertyRule.Type)
	switch errCode {
	case http.StatusNotFound:
		properties := make(map[string]interface{})
		rulesRef := make(map[string]interface{})
		properties[smartPropertyRule.Name] = rule.Value
		rulesRef[smartPropertyRule.ID] = smartPropertyRule.Name
		propertiesJson, err := json.Marshal(&properties)
		if err != nil {
			return http.StatusInternalServerError
		}
		propertiesJsonb := &postgres.Jsonb{propertiesJson}

		rulesRefJson, err := json.Marshal(&rulesRef)
		if err != nil {
			return http.StatusInternalServerError
		}
		rulesRefJsonb := &postgres.Jsonb{rulesRefJson}
		smartProperty := model.SmartProperties{
			ProjectID:      smartPropertyRule.ProjectID,
			ObjectType:     smartPropertyRule.Type,
			ObjectID:       objectID,
			ObjectProperty: objectPropertiesJsonb,
			Properties:     propertiesJsonb,
			RulesRef:       rulesRefJsonb,
			Source:         source,
			CreatedAt:      time.Now().UTC(),
			UpdatedAt:      time.Now().UTC(),
		}
		return store.CreateSmartProperty(&smartProperty)

	case http.StatusFound:
		updatedSmartProperty, errCodeGet := getUpdatedSmartPropertyObjectForExistingSmartProperty(smartPropertyRule, objectID, objectPropertiesJsonb, rule, smartProperty, source)
		if errCodeGet != http.StatusFound {
			return errCodeGet
		}
		errCodeUpdate := store.UpdateSmartProperty(&updatedSmartProperty)
		if errCodeUpdate != http.StatusAccepted {
			return errCodeUpdate
		} else {
			return http.StatusCreated
		}

	default:
		return errCode
	}
}

func getUpdatedSmartPropertyObjectForExistingSmartProperty(smartPropertyRule *model.SmartPropertyRules, objectID string,
	objectPropertiesJsonb *postgres.Jsonb, rule model.Rule, smartProperty model.SmartProperties, source string) (model.SmartProperties, int) {
	properties := make(map[string]interface{})
	rulesRef := make(map[string]interface{})
	err := U.DecodePostgresJsonbToStructType(smartProperty.Properties, &properties)
	if err != nil {
		return model.SmartProperties{}, http.StatusInternalServerError
	}
	err = U.DecodePostgresJsonbToStructType(smartProperty.RulesRef, &rulesRef)
	if err != nil {
		return model.SmartProperties{}, http.StatusInternalServerError
	}

	properties[smartPropertyRule.Name] = rule.Value
	rulesRef[smartPropertyRule.ID] = smartPropertyRule.Name
	propertiesJson, err := json.Marshal(&properties)
	if err != nil {
		return model.SmartProperties{}, http.StatusInternalServerError
	}
	propertiesJsonb := &postgres.Jsonb{propertiesJson}

	rulesRefJson, err := json.Marshal(&rulesRef)
	if err != nil {
		return model.SmartProperties{}, http.StatusInternalServerError
	}
	rulesRefJsonb := &postgres.Jsonb{rulesRefJson}

	updatedSmartProperty := model.SmartProperties{
		ProjectID:      smartPropertyRule.ProjectID,
		ObjectType:     smartPropertyRule.Type,
		ObjectID:       objectID,
		ObjectProperty: objectPropertiesJsonb,
		Properties:     propertiesJsonb,
		RulesRef:       rulesRefJsonb,
		Source:         source,
		CreatedAt:      smartPropertyRule.CreatedAt,
		UpdatedAt:      time.Now().UTC(),
	}
	return updatedSmartProperty, http.StatusFound
}

func (store *MemSQL) DeleteSmartPropertyByRuleID(projectID uint64, ruleID string) (int, int, int) {
	db := C.GetServices().Db
	smartProperty := make([]model.SmartProperties, 0, 0)

	err := db.Table("smart_properties").Where("project_id = ? AND rules_ref->>? IS NOT NULL", projectID, ruleID).Find(&smartProperty).Error
	if err != nil {
		if err.Error() == "record not found" {
			return 0, 0, http.StatusAccepted
		}
		errString := fmt.Sprintf("Failed to delete smart properties for project %d and rule %s", projectID, ruleID)
		log.WithField("error string", err).Error(errString)
		return 0, 0, http.StatusInternalServerError
	}
	recordsEvaluated := 0
	recordsUpdated := 0
	for _, smartProperty := range smartProperty {
		recordsEvaluated += 1
		properties := make(map[string]interface{})
		rulesRef := make(map[string]interface{})
		err := U.DecodePostgresJsonbToStructType(smartProperty.Properties, &properties)
		if err != nil {
			return recordsUpdated, recordsEvaluated, http.StatusInternalServerError
		}
		err = U.DecodePostgresJsonbToStructType(smartProperty.RulesRef, &rulesRef)
		if err != nil {
			return recordsUpdated, recordsEvaluated, http.StatusInternalServerError
		}
		var ruleName interface{}
		newProperties := make(map[string]interface{})
		newRulesRef := make(map[string]interface{})
		for key, value := range rulesRef {
			if key == ruleID {
				ruleName = value
				continue
			}
			newRulesRef[key] = value
		}
		for key, value := range properties {
			if key == ruleName {
				continue
			}
			newProperties[key] = value
		}
		if reflect.DeepEqual(newProperties, make(map[string]interface{})) {
			if !C.IsDryRunSmartProperties() {
				err := db.Delete(&smartProperty).Error
				if err != nil {
					return recordsUpdated, recordsEvaluated, http.StatusInternalServerError
				}
			}
		} else {
			propertiesJson, err := json.Marshal(&newProperties)
			if err != nil {
				return recordsUpdated, recordsEvaluated, http.StatusInternalServerError
			}
			propertiesJsonb := &postgres.Jsonb{propertiesJson}

			rulesRefJson, err := json.Marshal(&newRulesRef)
			if err != nil {
				return recordsUpdated, recordsEvaluated, http.StatusInternalServerError
			}
			rulesRefJsonb := &postgres.Jsonb{rulesRefJson}
			smartProperty.Properties = propertiesJsonb
			smartProperty.RulesRef = rulesRefJsonb
			if !C.IsDryRunSmartProperties() {
				err = db.Save(smartProperty).Error
				if err != nil {
					return recordsUpdated, recordsEvaluated, http.StatusInternalServerError
				}
			}
		}
		recordsUpdated += 1
	}
	return recordsUpdated, recordsEvaluated, http.StatusAccepted
}

func checkSmartProperty(filters []model.ChannelFilterV1, groupBys []model.ChannelGroupBy) bool {
	for _, filter := range filters {
		_, isPresent := Const.SmartPropertyReservedNames[filter.Property]
		if !isPresent {
			return true
		}
	}
	for _, groupBy := range groupBys {
		_, isPresent := Const.SmartPropertyReservedNames[groupBy.Property]
		if !isPresent {
			return true
		}
	}
	return false
}
func checkSmartPropertyWithTypeAndSource(filters []model.ChannelFilterV1, groupBys []model.ChannelGroupBy, source string) (bool, bool) {
	campaignProperty := false
	adGroupProperty := false
	for _, filter := range filters {
		_, isPresent := Const.SmartPropertyReservedNames[filter.Property]
		if !isPresent {
			switch source {
			case "adwords":
				if filter.Object == model.AdwordsCampaign {
					campaignProperty = true
				}
				if filter.Object == model.AdwordsAdGroup {
					adGroupProperty = true
				}
			case "facebook":
				if filter.Object == model.AdwordsCampaign {
					campaignProperty = true
				}
				if filter.Object == "ad_set" {
					adGroupProperty = true
				}
			case "linkedin":
				if filter.Object == "campaign_group" {
					campaignProperty = true
				}
				if filter.Object == model.AdwordsCampaign {
					adGroupProperty = true
				}
			}
		}
	}
	for _, groupBy := range groupBys {
		_, isPresent := Const.SmartPropertyReservedNames[groupBy.Property]
		if !isPresent {
			switch source {
			case "adwords":
				if groupBy.Object == model.AdwordsCampaign {
					campaignProperty = true
				}
				if groupBy.Object == model.AdwordsAdGroup {
					adGroupProperty = true
				}
			case "facebook":
				if groupBy.Object == model.AdwordsCampaign {
					campaignProperty = true
				}
				if groupBy.Object == "ad_set" {
					adGroupProperty = true
				}
			case "linkedin":
				if groupBy.Object == "campaign_group" {
					campaignProperty = true
				}
				if groupBy.Object == model.AdwordsCampaign {
					adGroupProperty = true
				}
			}
		}
	}
	return campaignProperty, adGroupProperty
}
