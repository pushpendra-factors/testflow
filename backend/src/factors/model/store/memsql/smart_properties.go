package memsql

import (
	"encoding/json"
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

const errorDuplicateSmartProperties = "pq: duplicate key value violates unique constraint \"smart_properties_primary_key\""

func isDuplicateSmartPropertiesError(err error) bool {
	return err.Error() == errorDuplicateSmartProperties
}

func (store *MemSQL) CreateSmartProperties(smartPropertiesDoc *model.SmartProperties) int {
	logCtx := log.WithField("project_id", smartPropertiesDoc.ProjectID)

	if smartPropertiesDoc.ProjectID == 0 {
		logCtx.Error("Invalid project ID.")
		return http.StatusBadRequest
	}

	db := C.GetServices().Db
	err := db.Create(&smartPropertiesDoc).Error
	if err != nil {
		if isDuplicateSmartPropertiesError(err) {
			logCtx.WithError(err).WithField("project_id", smartPropertiesDoc.ProjectID).Error(
				"Failed to create smart properties. Duplicate.")
			return http.StatusConflict
		}
		logCtx.WithError(err).WithField("project_id", smartPropertiesDoc.ProjectID).Error(
			"Failed to create smart properties.")
		return http.StatusInternalServerError
	}
	return http.StatusCreated
}
func (store *MemSQL) UpdateSmartProperties(smartPropertiesDoc *model.SmartProperties) int {
	logCtx := log.WithField("project_id", smartPropertiesDoc.ProjectID)

	if smartPropertiesDoc.ProjectID == 0 {
		logCtx.Error("Invalid project ID.")
		return http.StatusBadRequest
	}

	db := C.GetServices().Db
	err := db.Save(&smartPropertiesDoc).Error
	if err != nil {
		if isDuplicateSmartPropertiesError(err) {
			logCtx.WithError(err).WithField("project_id", smartPropertiesDoc.ProjectID).Error(
				"Failed to update smart properties. Duplicate.")
			return http.StatusConflict
		}
		logCtx.WithError(err).WithField("project_id", smartPropertiesDoc.ProjectID).Error(
			"Failed to update smart properties.")
		return http.StatusInternalServerError
	}
	return http.StatusAccepted
}
func (store *MemSQL) GetSmartPropertiesByProjectIDAndSourceAndObjectType(projectID uint64, source string, objectType int) ([]model.SmartProperties, int) {
	db := C.GetServices().Db
	smartProperties := []model.SmartProperties{}
	err := db.Table("smart_properties").Where("project_id = ? AND source = ? AND object_type = ?", projectID, source, objectType).Find(&smartProperties).Error
	if err != nil && err.Error() == "record not found" {
		return smartProperties, http.StatusNotFound
	} else if err != nil && err.Error() != "record not found" {
		return smartProperties, http.StatusInternalServerError
	} else {
		return smartProperties, http.StatusFound
	}
}
func (store *MemSQL) DeleteSmartPropertyByProjectIDAndSourceAndObjectID(projectID uint64, source string, objectID string) int {
	db := C.GetServices().Db
	err := db.Table("smart_properties").Where("project_id = ? AND source = ? AND object_id = ?", projectID, source, objectID).Delete(&model.SmartProperties{}).Error
	if err != nil {
		return http.StatusInternalServerError
	}
	return http.StatusAccepted
}

func (store *MemSQL) GetSmartPropertyByProjectIDAndObjectIDAndObjectType(projectID uint64, objectID string, objectType int) (model.SmartProperties, int) {
	db := C.GetServices().Db
	smartProperties := model.SmartProperties{}
	err := db.Table("smart_properties").Where("project_id = ? AND object_id = ? AND object_type = ?", projectID, objectID, objectType).Find(&smartProperties).Error
	if err != nil && err.Error() == "record not found" {
		return smartProperties, http.StatusNotFound
	} else if err != nil && err.Error() != "record not found" {
		return smartProperties, http.StatusInternalServerError
	} else {
		return smartProperties, http.StatusFound
	}
}
func (store *MemSQL) CreateSmartPropertiesFromChannelDocumentAndRule(smartPropertiesRule *model.SmartPropertiesRules, rule model.Rule,
	channelDocument model.ChannelDocumentsWithFields, source string) int {
	var objectID string
	if smartPropertiesRule.Type == 1 {
		objectID = channelDocument.CampaignID
	} else {
		objectID = channelDocument.AdGroupID
	}
	objectPropertiesJson, err := json.Marshal(channelDocument)
	if err != nil {
		return http.StatusInternalServerError
	}
	objectPropertiesJsonb := &postgres.Jsonb{objectPropertiesJson}

	smartProperty, errCode := store.GetSmartPropertyByProjectIDAndObjectIDAndObjectType(smartPropertiesRule.ProjectID, objectID, smartPropertiesRule.Type)
	switch errCode {
	case http.StatusNotFound:
		properties := make(map[string]interface{})
		rulesRef := make(map[string]interface{})
		properties[smartPropertiesRule.Name] = rule.Value
		rulesRef[smartPropertiesRule.ID] = smartPropertiesRule.Name
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
		smartProperties := model.SmartProperties{
			ProjectID:      smartPropertiesRule.ProjectID,
			ObjectType:     smartPropertiesRule.Type,
			ObjectID:       objectID,
			ObjectProperty: objectPropertiesJsonb,
			Properties:     propertiesJsonb,
			RulesRef:       rulesRefJsonb,
			Source:         source,
			CreatedAt:      time.Now().UTC(),
			UpdatedAt:      time.Now().UTC(),
		}
		return store.CreateSmartProperties(&smartProperties)

	case http.StatusFound:
		updatedSmartProperty, errCodeGet := getUpdatedSmartPropertiesObjectForExistingSmartProperties(smartPropertiesRule, objectID, objectPropertiesJsonb, rule, smartProperty, source)
		if errCodeGet != http.StatusFound {
			return errCodeGet
		}
		errCodeUpdate := store.UpdateSmartProperties(&updatedSmartProperty)
		if errCodeUpdate != http.StatusAccepted {
			return errCodeUpdate
		} else {
			return http.StatusCreated
		}

	default:
		return errCode
	}
}

func getUpdatedSmartPropertiesObjectForExistingSmartProperties(smartPropertiesRule *model.SmartPropertiesRules, objectID string,
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

	properties[smartPropertiesRule.Name] = rule.Value
	rulesRef[smartPropertiesRule.ID] = smartPropertiesRule.Name
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

	smartProperties := model.SmartProperties{
		ProjectID:      smartPropertiesRule.ProjectID,
		ObjectType:     smartPropertiesRule.Type,
		ObjectID:       objectID,
		ObjectProperty: objectPropertiesJsonb,
		Properties:     propertiesJsonb,
		RulesRef:       rulesRefJsonb,
		Source:         source,
		CreatedAt:      smartPropertiesRule.CreatedAt,
		UpdatedAt:      time.Now().UTC(),
	}
	return smartProperties, http.StatusFound
}

func (store *MemSQL) DeleteSmartPropertiesByRuleID(projectID uint64, ruleID string) (int, int, int) {
	db := C.GetServices().Db
	smartProperties := make([]model.SmartProperties, 0, 0)

	err := db.Table("smart_properties").Where("project_id = ? AND rules_ref->>? IS NOT NULL", projectID, ruleID).Find(&smartProperties).Error
	if err != nil {
		if err.Error() == "record not found" {
			return 0, 0, http.StatusAccepted
		}
		errString := fmt.Sprintf("Failed to delete smart properties for project %d and rule %s", projectID, ruleID)
		log.Error(errString)
		return 0, 0, http.StatusInternalServerError
	}
	recordsEvaluated := 0
	recordsUpdated := 0
	for _, smartProperty := range smartProperties {
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
			err := db.Delete(&smartProperty).Error
			if err != nil {
				return recordsUpdated, recordsEvaluated, http.StatusInternalServerError
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
			err = db.Save(smartProperty).Error
			if err != nil {
				return recordsUpdated, recordsEvaluated, http.StatusInternalServerError
			}
		}
		recordsUpdated += 1
	}
	return recordsUpdated, recordsEvaluated, http.StatusAccepted
}

func checkSmartProperties(filters []model.ChannelFilterV1, groupBys []model.ChannelGroupBy) bool {
	for _, filter := range filters {
		_, isPresent := smartPropertiesDisallowedNames[filter.Property]
		if !isPresent {
			return true
		}
	}
	for _, groupBy := range groupBys {
		_, isPresent := smartPropertiesDisallowedNames[groupBy.Property]
		if !isPresent {
			return true
		}
	}
	return false
}
