package memsql

import (
	"errors"
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"fmt"
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

func (store *MemSQL) CreateOrGetGroupByName(projectID int64, groupName string, allowedGroupNames map[string]bool) (*model.Group, int) {
	if groupName == model.GROUP_NAME_DOMAINS {
		return nil, http.StatusBadRequest
	}
	return store.createOrGetGroupByName(projectID, groupName, allowedGroupNames)
}

func (store *MemSQL) CreateOrGetDomainsGroup(projectID int64) (*model.Group, int) {

	return store.createOrGetGroupByName(projectID, model.GROUP_NAME_DOMAINS, map[string]bool{model.GROUP_NAME_DOMAINS: true})
}

func (store *MemSQL) createOrGetGroupByName(projectID int64, groupName string, allowedGroupNames map[string]bool) (*model.Group, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"group_name": groupName,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	group, status := store.GetGroup(projectID, groupName)
	if status != http.StatusFound {
		group, status = store.CreateGroup(projectID, groupName, allowedGroupNames)
		if status != http.StatusCreated && status != http.StatusConflict {
			logCtx.Error("Failed to create or get group.")
		}

		return group, status
	}

	return group, http.StatusFound
}

func (store *MemSQL) GetGroups(projectId int64) ([]model.Group, int) {
	logFields := log.Fields{
		"project_id": projectId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	if projectId == 0 {
		logCtx.Error("Invalid project_id.")
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
		// Temporary fix, to be removed later on when the cache resets
		for key, value := range groupProperty.Property {
			if key == U.G2_EMPLOYEES && value.Category == U.PropertyTypeUnknown {
				value.Category = U.PropertyTypeNumerical
				groupProperty.Property[key] = value
			}
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

	var lastError error
	for i := 0; i < lastNDays; i++ {
		currentDateOnlyFormat := currentDate.AddDate(0, 0, -i).Format(U.DATETIME_FORMAT_YYYYMMDD)
		value, err := model.GetPropertyValuesByGroupPropertyFromCache(projectID,
			groupName, propertyName, currentDateOnlyFormat)
		if err != nil {
			lastError = err
			continue
		}
		values = append(values, value)
	}

	if len(values) == 0 {
		logCtx.WithError(lastError).Error("No property values from cache for last N days.")
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

func (store *MemSQL) GetGroupUserByGroupID(projectID int64, groupName string, groupID string) (*model.User, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"group_name": groupName,
		"group_id":   groupID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	if projectID == 0 || groupName == "" || groupID == "" {
		logCtx.Error("Invalid parameters.")
		return nil, http.StatusBadRequest
	}

	group, status := store.GetGroup(projectID, groupName)
	if status != http.StatusFound {
		logCtx.Error("Failed to get group on GetGroupUserByGroupID.")
		return nil, http.StatusInternalServerError
	}

	source := model.GetGroupUserSourceByGroupName(groupName)

	whereStmnt := fmt.Sprintf("project_id = ? AND group_%d_id = ? AND source = ? AND is_group_user = true ", group.ID)
	whereParams := []interface{}{projectID, groupID, source}

	db := C.GetServices().Db
	groupUser := model.User{}
	if err := db.Model(&model.User{}).
		Where(whereStmnt, whereParams...).
		Find(&groupUser).Error; err != nil {

		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}

		logCtx.WithError(err).Error("Failed to get group.")
		return nil, http.StatusInternalServerError
	}

	if groupUser.ID == "" {
		return nil, http.StatusNotFound
	}

	return &groupUser, http.StatusFound
}
