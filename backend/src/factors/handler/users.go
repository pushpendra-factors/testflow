package handler

import (
	C "factors/config"
	mid "factors/middleware"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// Test command.
// curl -i -X GET http://localhost:8080/projects/1/users/bc7318e8-2b69-49b6-baf3-fdf47bcb1af9
// GetUserHandler godoc
// @Summary Get a user for the given project and user id.
// @Tags Users
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param user_id path string true "User ID"
// @Success 200 {object} model.User
// @Router /{project_id}/users/{user_id} [get]
func GetUserHandler(c *gin.Context) {
	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		log.Error("GetUser Failed. ProjectId parse failed.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	id := c.Params.ByName("user_id")
	if id == "" {
		log.WithFields(log.Fields{"project_id": projectId}).Error("GetUser Failed. Missing id.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	user, errCode := store.GetStore().GetUser(projectId, id)
	if errCode != http.StatusFound {
		c.AbortWithStatus(errCode)
	} else {
		c.JSON(http.StatusOK, user)
	}
}

// Test command.
// curl -i -X GET http://localhost:8080/projects/1/users
// curl -i -X GET http://localhost:8080/projects/1/users?offset=50&limit=10
// GetUsersHandler godoc
// @Summary Gets users for the given project id.
// @Tags Users
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param offset query integer false "Offset"
// @Param limit query integer false "Limit"
// @Success 200 {array} model.User
// @Router /{project_id}/users [get]
func GetUsersHandler(c *gin.Context) {
	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	qParams := c.Request.URL.Query()

	var offset uint64 = 0
	offsets := qParams["offset"]
	if offsets != nil {
		offsetStr := offsets[0]
		if offsetParse, err := strconv.ParseUint(offsetStr, 10, 64); err != nil {
			log.WithError(err).Error("GetUsers Failed. Offset parse failed.")
			c.AbortWithStatus(http.StatusBadRequest)
			return
		} else {
			offset = offsetParse
		}
	}

	var limit uint64 = 10
	limits := qParams["limit"]
	if limits != nil {
		limitStr := limits[0]
		if limitParse, err := strconv.ParseUint(limitStr, 10, 64); err != nil {
			log.WithError(err).Error("GetUsers Failed. Limit parse failed.")
			c.AbortWithStatus(http.StatusBadRequest)
			return
		} else {
			limit = limitParse
		}
	}

	users, errCode := store.GetStore().GetUsers(projectId, offset, limit)
	if errCode != http.StatusFound {
		c.AbortWithStatus(errCode)
	} else {
		c.JSON(http.StatusOK, users)
	}
}

// GetUserPropertiesHandler Test command.
// curl -i -X GET http://localhost:8080/projects/1/user_properties
// GetUserPropertiesHandler godoc
// @Summary Gets users properties for the given project id.
// @Tags Users
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Success 200 {string} json "map[string]string"
// @Router /{project_id}/user_properties [get]
func GetUserPropertiesHandler(c *gin.Context) {
	var err error
	var properties map[string][]string
	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	logCtx := log.WithFields(log.Fields{
		"projectId": projectId,
	})

	properties, err = store.GetStore().GetUserPropertiesByProject(projectId, 2500, C.GetLookbackWindowForEventUserCache())
	if err != nil {
		logCtx.WithError(err).Error("get user properties by project")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if len(properties) == 0 {
		logCtx.WithError(err).Error(fmt.Sprintf("No user properties Returned - ProjectID - %v", projectId))
	}
	properties = U.ClassifyDateTimePropertyKeys(&properties)
	U.FillMandatoryDefaultUserProperties(&properties)
	U.FilterDisabledCoreUserProperties(&properties)

	c.JSON(http.StatusOK, properties)
}

//GetUserPropertyValuesHandler curl -i -X GET http://localhost:8080/projects/1/user_properties/$country
// GetUserPropertiesHandler godoc
// @Summary Get property values for given property name.
// @Tags Users
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param property_name path string true "Property Name"
// @Success 200 {string} json "[]string"
// @Router /{project_id}/user_properties/{property_name}/values [get]
func GetUserPropertyValuesHandler(c *gin.Context) {
	var err error
	var propertyValues []string
	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	logCtx := log.WithFields(log.Fields{
		"projectId": projectId,
	})

	propertyName := c.Params.ByName("property_name")
	if propertyName == "" {
		logCtx.WithField("property_name", propertyName).Error("null propertyname")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	propertyValues, err = store.GetStore().GetPropertyValuesByUserProperty(projectId, propertyName, 2500, C.GetLookbackWindowForEventUserCache())
	if err != nil {
		logCtx.WithError(err).Error("get property values by user property")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	if len(propertyValues) == 0 {
		logCtx.WithError(err).Error(fmt.Sprintf("No user properties Returned - ProjectID - %v, propertyName - %s", projectId, propertyName))
	}
	c.JSON(http.StatusOK, propertyValues)
}
