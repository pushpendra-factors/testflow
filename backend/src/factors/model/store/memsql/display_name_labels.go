package memsql

import (
	C "factors/config"
	"factors/model/model"
	"fmt"
	"net/http"
	"strings"

	"errors"
	U "factors/util"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) CreateOrUpdateDisplayNameLabel(projectID int64, source, propertyKey, value, label string) int {
	logFields := log.Fields{
		"project_id":   projectID,
		"source":       source,
		"property_key": propertyKey,
		"value":        value,
		"label":        label,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	if projectID == 0 || source == "" || propertyKey == "" || value == "" {
		logCtx.Error("Missing required field.")
		return http.StatusBadRequest
	}

	element, getErrCode, getErr := store.GetDisplayNameLabel(projectID, source, propertyKey, value)
	if getErrCode == http.StatusBadRequest || getErrCode == http.StatusInternalServerError {
		logCtx.WithError(getErr).Error("Failed on CreateOrUpdateDisplayNameLabel.")
		return getErrCode
	} else if getErrCode == http.StatusNotFound {
		status, err := store.CreateDisplayNameLabel(projectID, source, propertyKey, value, label)
		if status != http.StatusCreated {
			logCtx.WithError(err).Error("Failed on CreateOrUpdateDisplayNameLabel.")
			return status
		}
		return http.StatusCreated
	}

	if element.Label == label {
		return http.StatusConflict
	}

	db := C.GetServices().Db
	if err := db.Model(&model.DisplayNameLabel{}).
		Where("project_id = ? AND source = ? AND property_key = ? AND value = ?", projectID, source, propertyKey, value).
		Update(map[string]interface{}{"label": label}).Error; err != nil {

		logCtx.WithField("err", err).Error("Failed to update display name label.")

		return http.StatusInternalServerError
	}

	return http.StatusAccepted
}

func (store *MemSQL) CreateDisplayNameLabel(projectID int64, source, propertyKey, value, label string) (int, error) {
	if projectID == 0 || source == "" || propertyKey == "" || value == "" || label == "" {
		return http.StatusBadRequest, errors.New("Invalid parameters.")
	}

	displayNameLabel := model.DisplayNameLabel{
		ID:          U.GetUUID(),
		ProjectID:   projectID,
		Source:      source,
		PropertyKey: propertyKey,
		Value:       value,
		Label:       label,
	}

	db := C.GetServices().Db
	dbx := db.Create(&displayNameLabel)
	// Duplicates gracefully handled and allowed further.
	if dbx.Error != nil && !IsDuplicateRecordError(dbx.Error) {
		return http.StatusInternalServerError, errors.New("Failed to create a display name")
	}

	return http.StatusCreated, nil
}

func (store *MemSQL) GetDisplayNameLabel(projectID int64, source, propertyKey, value string) (*model.DisplayNameLabel, int, error) {
	if projectID == 0 || source == "" || propertyKey == "" || value == "" {
		return nil, http.StatusBadRequest, errors.New("Invalid parameters.")
	}

	var displayNameLabel model.DisplayNameLabel

	db := C.GetServices().Db
	dbx := db.Limit(1).Where("project_id = ? AND source = ? AND property_key = ? AND value = ?", projectID, source, propertyKey, value)

	if err := dbx.Find(&displayNameLabel).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound, nil
		}
		return nil, http.StatusInternalServerError, errors.New("Failed to get display name label on GetDisplayNameLabel.")
	}

	return &displayNameLabel, http.StatusFound, nil
}

func (store *MemSQL) GetDisplayNameLabelsByProjectIdAndSource(projectID int64, source string) ([]model.DisplayNameLabel, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"source":     source,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	if projectID == 0 || source == "" {
		logCtx.Error("Invalid parameters.")
		return nil, http.StatusBadRequest
	}

	var displayNameLabel []model.DisplayNameLabel

	db := C.GetServices().Db
	dbx := db.Where("project_id = ? AND source = ?", projectID, source)

	if err := dbx.Find(&displayNameLabel).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}

		logCtx.WithError(err).Error("Failed to get display name labels on GetDisplayNameLabelsByProjectIdAndSource.")
		return nil, http.StatusInternalServerError
	}

	return displayNameLabel, http.StatusFound
}

func (store *MemSQL) GetPropertyLabelAndValuesByProjectIdAndPropertyKey(projectID int64, source, propertyKey string) (map[string]string, error) {
	propertyValueLabelMap := make(map[string]string, 0)

	if projectID == 0 {
		return propertyValueLabelMap, errors.New("invalid project on GetPropertyLabelAndValuesByProjectIdAndPropertyKey")
	}

	if propertyKey == "" {
		return propertyValueLabelMap, errors.New("invalid property_key on GetPropertyLabelAndValuesByProjectIdAndPropertyKey")
	}

	if source == "" {
		return propertyValueLabelMap, errors.New("invalid source on GetPropertyLabelAndValuesByProjectIdAndPropertyKey")
	}

	var displayNameLabels []model.DisplayNameLabel

	db := C.GetServices().Db
	dbx := db.Where("project_id = ? AND source = ? AND property_key = ?", projectID, source, propertyKey)

	if err := dbx.Find(&displayNameLabels).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return propertyValueLabelMap, nil
		}

		return propertyValueLabelMap, errors.New("failed to get display name labels on GetPropertyLabelAndValuesByProjectIdAndPropertyKey")
	}

	if len(displayNameLabels) == 0 {
		return propertyValueLabelMap, nil
	}

	for i := range displayNameLabels {
		propertyValueLabelMap[displayNameLabels[i].Value] = displayNameLabels[i].Label
	}

	return propertyValueLabelMap, nil
}

func (store *MemSQL) AddPropertyValueLabelToQueryResults(projectID int64, oldResults []model.QueryResult) ([]model.QueryResult, error) {
	newResults := make([]model.QueryResult, 0, 0)
	for i := range oldResults {
		result, err := U.EncodeStructTypeToMap(&oldResults[i])
		if err != nil {
			return oldResults, err
		}

		result, err = store.TransformQueryResultsColumnValuesToLabel(projectID, result)
		if err != nil {
			return oldResults, fmt.Errorf("Failed to set property value labels")
		}

		var newResult model.QueryResult
		err = U.DecodeInterfaceMapToStructType(result, &newResult)
		if err != nil {
			return oldResults, fmt.Errorf("Failed to encode map to query_result on AddPropertyValueLabelToQueryResults")
		}
		newResults = append(newResults, newResult)
	}
	return newResults, nil
}

func (store *MemSQL) TransformQueryResultsColumnValuesToLabel(projectID int64, result map[string]interface{}) (map[string]interface{}, error) {
	if projectID == 0 || len(result) == 0 {
		return result, errors.New("Invalid parameters")
	}

	logCtx := log.WithField("project_id", projectID)

	if _, exists := result["headers"]; !exists {
		return result, fmt.Errorf("Headers not available in map on TransformQueryResultsColumnValuesToLabel")
	}

	if _, exists := result["rows"]; !exists {
		return result, fmt.Errorf("Rows not available in map on TransformQueryResultsColumnValuesToLabel")
	}

	headers, ok := result["headers"].([]interface{})
	if !ok {
		return result, errors.New("Cannot decode headers to []interface{} on TransformQueryResultsColumnValuesToLabel")
	}
	rowsInt, ok := result["rows"].([]interface{})
	if !ok {
		return result, errors.New("Cannot decode rows to []interface{} on TransformQueryResultsColumnValuesToLabel")
	}

	rows := make([][]interface{}, 0, 0)
	for i := range rowsInt {
		if rowsInt[i] != nil {
			row, ok := rowsInt[i].([]interface{})
			if !ok {
				return result, errors.New("Cannot decode rows to [][]interface{} on TransformQueryResultsColumnValuesToLabel")
			}
			rows = append(rows, row)
		}
	}

	for _, header := range headers {
		propertyName := U.GetPropertyValueAsString(header)
		if !U.IsAllowedCRMPropertyPrefix(propertyName) {
			continue
		}

		source := strings.Split(propertyName, "_")[0]
		source = strings.TrimPrefix(source, "$")

		propertyIndex := -1
		for i, property := range headers {
			if propertyName == property {
				propertyIndex = i
			}
		}

		if propertyIndex == -1 {
			logCtx.WithFields(log.Fields{"source": source, "property_key": propertyName}).Error("Failed to get property value labels in result headers on transformResultsColumnValuesToLabel")
			continue
		}

		propertyValueLabels, err := store.GetPropertyLabelAndValuesByProjectIdAndPropertyKey(projectID, source, propertyName)
		if err != nil {
			logCtx.WithFields(log.Fields{"source": source, "property_key": propertyName}).WithError(err).Error("Failed to get property value labels on transformResultsColumnValuesToLabel")
			continue
		}

		if len(propertyValueLabels) == 0 {
			continue
		}

		for i := range rows {
			propertyValue := U.GetPropertyValueAsString(rows[i][propertyIndex])
			if label, exists := propertyValueLabels[propertyValue]; exists {
				rows[i][propertyIndex] = label
			}
		}
	}
	return result, nil
}

func (store *MemSQL) AddPropertyValueLabelsToProfileResults(projectID int64, results []model.Profile) []model.Profile {
	logCtx := log.WithField("project_id", projectID)

	tablePropsMap := make(map[string]bool, 0)
	for _, result := range results {
		for prop := range result.TableProps {
			tablePropsMap[prop] = true
		}
	}

	for propertyName := range tablePropsMap {
		if !U.IsAllowedCRMPropertyPrefix(propertyName) {
			continue
		}

		source := strings.Split(propertyName, "_")[0]
		source = strings.TrimPrefix(source, "$")

		propertyValueLabels, err := store.GetPropertyLabelAndValuesByProjectIdAndPropertyKey(projectID, source, propertyName)
		if err != nil {
			logCtx.WithFields(log.Fields{"source": source, "property_key": propertyName}).WithError(err).Error("Failed to get property value labels on AddPropertyValueLabelsToProfileResults")
			continue
		}

		if len(propertyValueLabels) == 0 {
			continue
		}

		for i := range results {
			propertyValue := U.GetPropertyValueAsString(results[i].TableProps[propertyName])
			if propertyValue == "" || propertyValueLabels[propertyValue] == "" {
				continue
			}

			results[i].TableProps[propertyName] = propertyValueLabels[propertyValue]
		}
	}

	return results
}
