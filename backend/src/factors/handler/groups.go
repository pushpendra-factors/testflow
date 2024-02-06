package handler

import (
	"encoding/base64"
	C "factors/config"
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// GetGroupsHandler curl -i -X GET http://localhost:8080/projects/1/groups
// GetGroupsHandler godoc
// @Summary Get all groups
// @Tags Groups
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Success 200 {string} json "[]json"
// @Router /{project_id}/groups [get]
func GetGroupsHandler(c *gin.Context) {
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	logCtx := log.WithFields(log.Fields{
		"projectId": projectId,
	})
	if projectId == 0 {
		logCtx.Error("invalid project_id.")
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Get groups failed. Invalid project."})
		return
	}

	groups, errCode := store.GetStore().GetGroups(projectId)
	if errCode != http.StatusFound {
		logCtx.Error("Get groups failed.")
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Get groups failed."})
		return
	}

	groups = filterOutGroupByName(groups, model.GROUP_NAME_DOMAINS)

	response := buildResponseFromGroups(groups)

	c.JSON(http.StatusFound, response)
}

func filterOutGroupByName(groups []model.Group, nameToRemove string) []model.Group {
	var updatedGroups []model.Group
	for _, group := range groups {
		if group.Name != nameToRemove {
			updatedGroups = append(updatedGroups, group)
		}
	}
	return updatedGroups
}

func buildResponseFromGroups(groups []model.Group) []model.GroupName {
	var response []model.GroupName
	for _, group := range groups {
		grp := model.GroupName{
			Name:        group.Name,
			DisplayName: getGroupDisplayName(group.Name),
			IsAccount:   model.AccountGroupNames[group.Name],
		}
		response = append(response, grp)
	}
	return response
}

func getGroupDisplayName(groupName string) string {
	if displayName, exists := U.STANDARD_GROUP_DISPLAY_NAMES[groupName]; exists {
		return displayName
	}
	return U.CreateVirtualDisplayName(groupName)
}

// GetGroupPropertiesHandler curl -i -X GET http://localhost:8080/projects/1/groups/zzxxzzz==/properties
// GetGroupPropertiesHandler godoc
// @Summary Gets property keys of the given group
// @Tags Groups
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param group_name path string true "Base64 encoded Group Name"
// @Success 200 {string} json "map[string]map[string][]string"
// @Router /{project_id}/groups/{group_name}/properties [get]
func GetGroupPropertiesHandler(c *gin.Context) {
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	logCtx := log.WithField("projectId", projectId)

	encodedGroupName := c.Params.ByName("group_name")
	if encodedGroupName == "" {
		logCtx.WithField("group_name", encodedGroupName).Error("null group_name")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var err error
	var decNameInBytes_1 []byte
	var decNameInBytes_2 []byte
	decNameInBytes_1, err = base64.StdEncoding.DecodeString(encodedGroupName)
	if err != nil {
		logCtx.WithFields(log.Fields{
			"encodedName": encodedGroupName,
			log.ErrorKey:  err,
		}).Error("Failed decoding event_name_1.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	decodedGroupName := string(decNameInBytes_1)

	decNameInBytes_2, err = base64.StdEncoding.DecodeString(decodedGroupName)
	if err != nil {
		logCtx.WithFields(log.Fields{
			"encodedName": decodedGroupName,
			log.ErrorKey:  err,
		}).Error("Failed decoding event_name_2.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	groupName := string(decNameInBytes_2)

	var propertiesFromCache map[string][]string
	var status int

	if model.IsDomainGroup(groupName) {
		allAccountProperties := getDefaultAllAccountProperties(projectId)
		response := gin.H{
			"properties":    allAccountProperties,
			"display_names": U.ALL_ACCOUNT_DEFAULT_PROPERTIES_DISPLAY_NAMES,
		}
		c.JSON(http.StatusOK, response)
		return
	} else if groupName == model.GROUP_NAME_SIX_SIGNAL {
		response := gin.H{
			"properties":    model.GetSixSignalDefaultUserProperties(),
			"display_names": model.GetSixSignalDefaultUserPropertiesDisplayNames(),
		}
		c.JSON(http.StatusOK, response)
		return
	} else {
		propertiesFromCache, status = store.GetStore().GetPropertiesByGroup(projectId, groupName, 2500,
			C.GetLookbackWindowForEventUserCache())
		if status == http.StatusInternalServerError {
			c.AbortWithStatus(status)
			return
		}
	}

	response := gin.H{"properties": U.FilterEmptyKeysAndValues(projectId, propertiesFromCache)}

	displayNamesOp := make(map[string]string)
	_, displayNames := store.GetStore().GetDisplayNamesForAllUserProperties(projectId)
	for property, displayName := range displayNames {
		displayNamesOp[property] = strings.Title(displayName)
	}

	_, displayNames = store.GetStore().GetDisplayNamesForObjectEntities(projectId)
	for property, displayName := range displayNames {
		displayNamesOp[property] = strings.Title(displayName)
	}

	for property, displayName := range U.STANDARD_USER_PROPERTIES_DISPLAY_NAMES {
		displayNamesOp[property] = strings.Title(displayName)
	}

	displayNamesMap := make(map[string]string)
	for _, props := range propertiesFromCache {
		for _, prop := range props {
			if displayName, exists := displayNamesOp[prop]; exists {
				displayNamesMap[prop] = displayName
			} else {
				displayNamesMap[prop] = U.CreateVirtualDisplayName(prop)
			}
		}
	}

	dupCheck := make(map[string]bool)
	for _, name := range displayNamesOp {
		if _, exists := dupCheck[name]; exists {
			logCtx.Warning(fmt.Sprintf("Duplicate display name %s", name))
		}
		dupCheck[name] = true
	}

	response["display_names"] = U.FilterDisplayNameEmptyKeysAndValues(projectId, displayNamesMap)

	c.JSON(http.StatusOK, response)
}

func getDefaultAllAccountProperties(projectId int64) map[string][]string {
	groups, errCode := store.GetStore().GetGroups(projectId)
	if errCode != http.StatusFound {
		log.WithField("err_code", errCode).Error("Failed to get groups while adding group info")
		return map[string][]string{}
	}

	var allAccountPropertiesCategorical, allAccountPropertiesNumerical []string
	for _, group := range groups {
		if prop, exists := U.GROUP_TO_DEFAULT_SEGMENT_MAP[group.Name]; exists {
			allAccountPropertiesCategorical = append(allAccountPropertiesCategorical, prop)
		}
	}
	allAccountPropertiesCategorical = append(allAccountPropertiesCategorical, U.VISITED_WEBSITE, U.DP_DOMAIN_NAME, U.GROUP_EVENT_NAME_ENGAGEMENT_LEVEL)

	scoringAvailable, err := store.GetStore().GetFeatureStatusForProjectV2(projectId, model.FEATURE_ACCOUNT_SCORING, false)
	if err != nil {
		log.WithField("err_code", errCode).WithField("project_id", projectId).Error("Error fetching scoring availability status for the project")
	}
	if scoringAvailable {
		allAccountPropertiesNumerical = append(allAccountPropertiesNumerical, U.GROUP_EVENT_NAME_ENGAGEMENT_SCORE, U.GROUP_EVENT_NAME_TOTAL_ENGAGEMENT_SCORE)
	}

	allAccountProperties := map[string][]string{
		"categorical": allAccountPropertiesCategorical,
		"numerical":   allAccountPropertiesNumerical,
	}
	return allAccountProperties
}

// GetGroupPropertyValuesHandler curl -i -X GET http://localhost:8080/projects/1/groups/zzxxzzz==/properties/email_id/values
// GetGroupPropertyValuesHandler godoc
// @Summary Gets values of the given group and its property.
// @Tags Groups
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param group_name path string true "Base64 encoded Group Name"
// @Param property_name path string true "Property Name"
// @Success 200 {string} json "[]string"
// @Router /{project_id}/groups/{group_name}/properties/{property_name}/values [get]
func GetGroupPropertyValuesHandler(c *gin.Context) {
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	logCtx := log.WithField("projectId", projectId)

	encodedGroupName := c.Params.ByName("group_name")
	if encodedGroupName == "" {
		logCtx.WithField("group_name", encodedGroupName).Error("null group_name")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	var err error
	var decNameInBytes_1 []byte
	var decNameInBytes_2 []byte
	decNameInBytes_1, err = base64.StdEncoding.DecodeString(encodedGroupName)
	if err != nil {
		logCtx.WithFields(log.Fields{
			"encodedName": encodedGroupName,
			log.ErrorKey:  err,
		}).Error("Failed decoding event_name_1.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	decodedGroupName := string(decNameInBytes_1)

	decNameInBytes_2, err = base64.StdEncoding.DecodeString(decodedGroupName)
	if err != nil {
		logCtx.WithFields(log.Fields{
			"encodedName": decodedGroupName,
			log.ErrorKey:  err,
		}).Error("Failed decoding event_name_2.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	groupName := string(decNameInBytes_2)
	logCtx = logCtx.WithField("group_name", groupName)

	propertyName := c.Params.ByName("property_name")
	if propertyName == "" {
		logCtx.WithField("propertyName", propertyName).Error("null property name")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	logCtx = logCtx.WithField("property_name", propertyName)
	if model.IsDomainGroup(groupName) {

		if U.ContainsStringInArray(U.ALL_ACCOUNT_DEFAULT_PROPERTIES, propertyName) {

			label := c.Query("label")
			if label == "true" {
				propertyValueLabel := map[string]string{"false": "False", "true": "True"}
				c.JSON(http.StatusOK, U.FilterDisplayNameEmptyKeysAndValues(projectId, propertyValueLabel))
				return
			}
			propertyValues := []string{"false", "true"}
			c.JSON(http.StatusOK, U.FilterEmptyArrayValues(propertyValues))
			return
		}

		if strings.EqualFold(U.GROUP_EVENT_NAME_ENGAGEMENT_LEVEL, propertyName) {
			propertyValues := []string{model.ENGAGEMENT_LEVEL_HOT, model.ENGAGEMENT_LEVEL_WARM, model.ENGAGEMENT_LEVEL_COOL, model.ENGAGEMENT_LEVEL_ICE}
			c.JSON(http.StatusOK, U.FilterEmptyArrayValues(propertyValues))
			return
		}
	}
	propertyValues, err := store.GetStore().GetPropertyValuesByGroupProperty(projectId, groupName,
		propertyName, 2500, C.GetLookbackWindowForEventUserCache())
	if err != nil {
		logCtx.WithError(err).Error("get properties values by group property")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if len(propertyValues) == 0 {
		logCtx.WithError(err).Warn("No property values returned.")
	}

	label := c.Query("label")
	if label == "true" {
		propertyValueLabel, err, isSourceEmpty := store.GetStore().GetPropertyValueLabel(projectId, propertyName, propertyValues)
		if err != nil {
			logCtx.WithError(err).Error("get group properties labels and values by property name")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		if isSourceEmpty {
			logCtx.WithField("property_name", propertyName).Warning("source is empty")
		}

		if len(propertyValueLabel) == 0 {
			logCtx.WithField("property_name", propertyName).Warn("No group property value labels returned")
		}

		c.JSON(http.StatusOK, U.FilterDisplayNameEmptyKeysAndValues(projectId, propertyValueLabel))
		return
	}

	propertyValues = U.SortRangePropertyValues(propertyName, propertyValues)
	c.JSON(http.StatusOK, U.FilterEmptyArrayValues(propertyValues))
}
