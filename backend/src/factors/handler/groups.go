package handler

import (
	"encoding/base64"
	C "factors/config"
	mid "factors/middleware"
	"factors/model/store"
	U "factors/util"
	"net/http"

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
	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Get groups failed. Invalid project."})
		return
	}
	groups, errCode := store.GetStore().GetGroups(projectId)
	if errCode != http.StatusFound {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Get groups failed."})
		return
	}

	c.JSON(http.StatusFound, groups)
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
	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
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

	decGroupNameInBytes, err := base64.StdEncoding.DecodeString(encodedGroupName)
	if err != nil {
		logCtx.WithField("encodedName", encodedGroupName).Error("Failed decoding group_name.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	groupName := string(decGroupNameInBytes)

	propertiesFromCache, status := store.GetStore().GetPropertiesByGroup(projectId, groupName, 2500,
		C.GetLookbackWindowForEventUserCache())
	if status == http.StatusInternalServerError {
		c.AbortWithStatus(status)
		return
	}

	response := gin.H{"properties": propertiesFromCache}

	displayNamesOp := make(map[string]string)
	_, displayNames := store.GetStore().GetDisplayNamesForAllUserProperties(projectId)
	for property, displayName := range displayNames {
		displayNamesOp[property] = displayName
	}

	_, displayNames = store.GetStore().GetDisplayNamesForObjectEntities(projectId)
	for property, displayName := range displayNames {
		displayNamesOp[property] = displayName
	}

	response["display_names"] = displayNamesOp

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
	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
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
	decGroupNameInBytes, err := base64.StdEncoding.DecodeString(encodedGroupName)
	if err != nil {
		logCtx.WithField("encodedName", encodedGroupName).Error("Failed decoding group_name.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	groupName := string(decGroupNameInBytes)
	logCtx = logCtx.WithField("group_name", groupName)

	propertyName := c.Params.ByName("property_name")
	if propertyName == "" {
		logCtx.WithField("propertyName", propertyName).Error("null property name")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	logCtx = logCtx.WithField("property_name", propertyName)

	propertyValues, err := store.GetStore().GetPropertyValuesByGroupProperty(projectId, groupName,
		propertyName, 2500, C.GetLookbackWindowForEventUserCache())
	if err != nil {
		logCtx.WithError(err).Error("get properties values by event property")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	if len(propertyValues) == 0 {
		logCtx.WithError(err).Error("No property values returned.")
	}

	c.JSON(http.StatusOK, propertyValues)
}
