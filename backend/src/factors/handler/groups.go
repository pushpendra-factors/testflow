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
	isAccount := c.Query("is_account")
	logCtx = log.WithFields(log.Fields{
		"projectId": projectId,
		"isAccount": isAccount,
	})
	// isAccount : true, false, ""
	groups, errCode := store.GetStore().GetGroups(projectId)
	if errCode != http.StatusFound {
		logCtx.Error("Get groups failed.")
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Get groups failed."})
		return
	}

	// isAccount : true, false, ""
	var filteredGroups []model.Group
	if isAccount == "true" || isAccount == "false" {
		isTrue := (isAccount == "true")
		for _, group := range groups {
			if model.AccountGroupNames[group.Name] == isTrue {
				filteredGroups = append(filteredGroups, group)
			}
		}
	} else {
		filteredGroups = groups
	}

	// remove $domains group
	for i, group := range filteredGroups {
		if group.Name == model.GROUP_NAME_DOMAINS {
			filteredGroups = append(filteredGroups[:i], filteredGroups[i+1:]...)
			break
		}
	}

	response := make(map[string]string)
	for _, group := range filteredGroups {
		name := group.Name
		displayName, exists := U.STANDARD_GROUP_DISPLAY_NAMES[name]
		if !exists {
			displayName = name
		}
		response[name] = displayName
	}

	c.JSON(http.StatusFound, response)
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

	if groupName == "All" || groupName == U.GROUP_NAME_DOMAINS {

		propertiesFromCache["categorical"] = U.ALL_ACCOUNT_DEFAULT_PROPERTIES

		response := gin.H{
			"properties": map[string][]string{
				"categorical": U.ALL_ACCOUNT_DEFAULT_PROPERTIES,
			},
			"display_names": U.ALL_ACCOUNT_DEFAULT_PROPERTIES_DISPLAY_NAMES,
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
	for _, props := range propertiesFromCache {
		for _, prop := range props {
			displayName := U.CreateVirtualDisplayName(prop)
			_, exist := displayNamesOp[prop]
			if !exist {
				displayNamesOp[prop] = displayName
			}
		}
	}

	dupCheck := make(map[string]bool)
	for _, name := range displayNamesOp {
		_, exists := dupCheck[name]
		if exists {
			logCtx.Warning(fmt.Sprintf("Duplicate display name %s", name))
		}
		dupCheck[name] = true
	}
	response["display_names"] = U.FilterDisplayNameEmptyKeysAndValues(projectId, displayNamesOp)

	c.JSON(http.StatusOK, response)
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
	if groupName == "All" || groupName == U.GROUP_NAME_DOMAINS {

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
		propertyValueLabel, err, isSourceEmpty := getPropertyValueLabel(projectId, propertyName, propertyValues)
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

	c.JSON(http.StatusOK, U.FilterEmptyArrayValues(propertyValues))
}
