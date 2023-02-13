package memsql

import (
	"errors"
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"net/http"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) CreateGroup(projectID int64, groupName string, allowedGroupNames map[string]bool) (*model.Group, int) {
	logFields := log.Fields{
		"project_id":          projectID,
		"group_name":          groupName,
		"allowed_group_names": allowedGroupNames,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	if projectID < 1 || groupName == "" {
		logCtx.Error("Invalid parameters.")
		return nil, http.StatusBadRequest
	}

	db := C.GetServices().Db

	if _, allowed := allowedGroupNames[groupName]; !allowed {
		logCtx.Error("group name not allowed.")
		return nil, http.StatusBadRequest
	}

	_, status := store.GetGroup(projectID, groupName)
	if status != http.StatusNotFound {
		if status == http.StatusFound {
			return nil, http.StatusConflict
		}

		logCtx.WithField("err_code", status).Error("Failed to get existing groups.")
		return nil, http.StatusInternalServerError
	}

	id := struct {
		MaxID int `json:"max_id"`
	}{}

	if err := db.Table("groups").Select("max(id) as max_id").Where("project_id = ?", projectID).Find(&id).Error; err != nil {
		logCtx.WithError(err).Error("Failed to get maximum id from groups.")
		return nil, http.StatusInternalServerError
	}

	if id.MaxID >= model.AllowedGroups {
		logCtx.Error("Maximum allowed groups reached.")
		return nil, http.StatusBadRequest
	}

	group := model.Group{
		ProjectID: projectID,
		Name:      groupName,
		ID:        id.MaxID + 1,
	}

	err := db.Create(&group).Error
	if err != nil {
		if IsDuplicateRecordError(err) {
			if strings.Contains(err.Error(), "PRIMARY") {
				return nil, http.StatusConflict
			}
		}

		logCtx.WithError(err).Error("Failed to insert group.")
		return nil, http.StatusInternalServerError

	}

	return &group, http.StatusCreated
}

func (store *MemSQL) GetGroups(projectId int64) ([]model.Group, int) {
	logFields := log.Fields{
		"project_id": projectId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	if projectId < 1 {
		logCtx.Error("Invalid parameters.")
		return nil, http.StatusBadRequest
	}

	var groups []model.Group
	db := C.GetServices().Db
	err := db.Where("project_id = ?", projectId).Find(&groups).Error
	if err != nil {
		log.WithField("project_id", projectId).WithError(err).Error("Failed to get groups.")
		return groups, http.StatusInternalServerError
	}

	return groups, http.StatusFound

}
func (store *MemSQL) GetGroup(projectID int64, groupName string) (*model.Group, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"group_name": groupName,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	if projectID < 1 || groupName == "" {
		logCtx.Error("Invalid parameters.")
		return nil, http.StatusBadRequest
	}

	db := C.GetServices().Db
	group := model.Group{}
	if err := db.Model(&model.Group{}).
		Where("project_id = ? AND name = ? ", projectID, groupName).
		Find(&group).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}

		logCtx.WithError(err).Error("Failed to get group.")
		return nil, http.StatusInternalServerError
	}

	return &group, http.StatusFound
}

// GetPropertiesByGroup (Part of group properties caching) This method iterates for last n days to get all the
// top 'limit' properties for the given group name. Picks all last 24 hours properties and sorts the remaining by occurence
// and returns top 'limit' properties
func (store *MemSQL) GetPropertiesByGroup(projectID int64, groupName string, limit int, lastNDays int) (map[string][]string, int) {
	logFields := log.Fields{
		"project_id":  projectID,
		"group_name":  groupName,
		"limit":       limit,
		"last_N_days": lastNDays,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	properties := make(map[string][]string)
	if projectID == 0 {
		return properties, http.StatusBadRequest
	}
	currentDate := model.OverrideCacheDateRangeForProjects(projectID)
	if groupName == "" || groupName == "undefined" {
		logCtx.Error("Invalid group_name.")
		return properties, http.StatusBadRequest
	}
	groupProperties := make([]U.CachePropertyWithTimestamp, 0)
	for i := 0; i < lastNDays; i++ {
		currentDateOnlyFormat := currentDate.AddDate(0, 0, -i).Format(U.DATETIME_FORMAT_YYYYMMDD)
		groupProperty, err := model.GetPropertiesByGroupFromCache(projectID, groupName, currentDateOnlyFormat)
		if err != nil {
			logCtx.WithField("current_date", currentDateOnlyFormat).WithField("error", err).Warn("Failed to get group properties from cache.")
			continue
		}
		groupProperties = append(groupProperties, groupProperty)
	}

	groupPropertiesAggregated := U.AggregatePropertyAcrossDate(groupProperties)
	groupPropertiesSorted := U.SortByTimestampAndCount(groupPropertiesAggregated)
	if limit > 0 {
		sliceLength := len(groupPropertiesSorted)
		if sliceLength > limit {
			groupPropertiesSorted = groupPropertiesSorted[0:limit]
		}
	}

	propertyDetails, propertyDetailsStatus := store.GetAllPropertyDetailsByProjectID(projectID, groupName, true)
	for _, v := range groupPropertiesSorted {
		category := v.Category
		if propertyDetailsStatus == http.StatusFound {
			pName := model.GetPropertyNameByTrimmedSmartEventPropertyPrefix(v.Name)
			if pType, exist := (*propertyDetails)[pName]; exist {
				category = pType
			}
		}

		if properties[category] == nil {
			properties[category] = make([]string, 0)
		}
		properties[category] = append(properties[category], v.Name)
	}

	return properties, http.StatusFound
}

// GetPropertyValuesByGroupProperty (Part of event_name and properties caching) This method iterates for
// last n days to get all the top 'limit' property values for the given property/event
// Picks all last 24 hours values and sorts the remaining by occurence and returns top 'limit' values
func (store *MemSQL) GetPropertyValuesByGroupProperty(projectID int64, groupName string,
	propertyName string, limit int, lastNDays int) ([]string, error) {
	logFields := log.Fields{
		"project_id":    projectID,
		"group_name":    groupName,
		"property_name": propertyName,
		"limit":         limit,
		"last_N_days":   lastNDays,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	if projectID == 0 {
		return []string{}, errors.New("invalid project on GetPropertyValuesByGroupProperty")
	}

	if groupName == "" {
		return []string{}, errors.New("invalid event_name on GetPropertyValuesByGroupProperty")
	}

	if propertyName == "" {
		return []string{}, errors.New("invalid property_name on GetPropertyValuesByGroupProperty")
	}
	currentDate := model.OverrideCacheDateRangeForProjects(projectID)
	values := make([]U.CachePropertyValueWithTimestamp, 0)
	for i := 0; i < lastNDays; i++ {
		currentDateOnlyFormat := currentDate.AddDate(0, 0, -i).Format(U.DATETIME_FORMAT_YYYYMMDD)
		value, err := model.GetPropertyValuesByGroupPropertyFromCache(projectID,
			groupName, propertyName, currentDateOnlyFormat)
		if err != nil {
			logCtx.WithField("current_date", currentDateOnlyFormat).WithError(err).
				Error("Failed to get group property values from cache for the given date.")
			continue
		}
		values = append(values, value)
	}

	valueStrings := make([]string, 0)
	valuesAggregated := U.AggregatePropertyValuesAcrossDate(values)
	valuesSorted := U.SortByTimestampAndCount(valuesAggregated)

	for _, v := range valuesSorted {
		valueStrings = append(valueStrings, v.Name)
	}
	if limit > 0 {
		sliceLength := len(valueStrings)
		if sliceLength > limit {
			return valueStrings[0:limit], nil
		}
	}
	return valueStrings, nil
}
