package handler

import (
	"factors/handler/helpers"
	mid "factors/middleware"
	M "factors/model"
	PC "factors/pattern_client"
	U "factors/util"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// Test command.
// curl -i -X GET http://localhost:8080/projects/1/users/bc7318e8-2b69-49b6-baf3-fdf47bcb1af9
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

	user, errCode := M.GetUser(projectId, id)
	if errCode != http.StatusFound {
		c.AbortWithStatus(errCode)
	} else {
		c.JSON(http.StatusOK, user)
	}
}

// Test command.
// curl -i -X GET http://localhost:8080/projects/1/users
// curl -i -X GET http://localhost:8080/projects/1/users?offset=50&limit=10
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

	users, errCode := M.GetUsers(projectId, offset, limit)
	if errCode != http.StatusFound {
		c.AbortWithStatus(errCode)
	} else {
		c.JSON(http.StatusOK, users)
	}
}

// Test command.
// curl -i -X GET http://localhost:8080/projects/1/user_properties
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

	if helpers.IsProjectWhitelistedForEventUserCache(projectId) {
		properties, err = M.GetUserPropertiesByProject(projectId, 2500, 30)
		if err != nil {
			logCtx.WithError(err).Error("get user properties by project")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		if len(properties) == 0 {
			logCtx.WithError(err).Error(fmt.Sprintf("No user properties Returned - ProjectID - %s", projectId))
		}
	} else {
		queryType := c.Query("query_type")
		if !helpers.IsValidQueryType(queryType) {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		reqId := U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID)
		modelId := uint64(0)
		modelIdParam := c.Query("model_id")
		if modelIdParam != "" {
			modelId, err = strconv.ParseUint(modelIdParam, 10, 64)
			if err != nil {
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}
			properties, err = PC.GetSeenUserProperties(reqId, projectId, modelId)
			if err != nil {
				log.WithFields(log.Fields{
					log.ErrorKey: err, "projectId": projectId}).Error(
					"Get User Properties from pattern servers failed.")
				properties = make(map[string][]string)
			}
		}
		var errCode int
		if len(properties) == 0 {
			properties, errCode = M.GetRecentUserPropertyKeys(projectId)
			if errCode == http.StatusInternalServerError {
				c.AbortWithStatus(errCode)
				return
			}
		}
	}
	properties = U.ClassifyDateTimePropertyKeys(&properties)
	U.FillMandatoryDefaultUserProperties(&properties)
	U.FilterDisabledCoreUserProperties(&properties)

	c.JSON(http.StatusOK, properties)
}

// curl -i -X GET http://localhost:8080/projects/1/user_properties/$country
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
	if helpers.IsProjectWhitelistedForEventUserCache(projectId) {
		propertyValues, err = M.GetPropertyValuesByUserProperty(projectId, propertyName, 2500, 30)
		if err != nil {
			logCtx.WithError(err).Error("get property values by user property")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		if len(propertyValues) == 0 {
			logCtx.WithError(err).Error(fmt.Sprintf("No user properties Returned - ProjectID - %s, propertyName - %s", projectId, propertyName))
		}
	} else {
		reqId := U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID)
		modelId := uint64(0)
		modelIdParam := c.Query("model_id")
		if modelIdParam != "" {
			modelId, err = strconv.ParseUint(modelIdParam, 10, 64)
			if err != nil {
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}

			propertyValues, err = PC.GetSeenUserPropertyValues(reqId, projectId, modelId, propertyName)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"projectId":    projectId,
					"propertyName": propertyName}).Error(
					"Get User Properties failed.")
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}
		}
		var errCode int
		if len(propertyValues) == 0 {
			propertyValues, errCode = M.GetRecentUserPropertyValues(projectId, propertyName)
			if errCode == http.StatusInternalServerError {
				c.AbortWithStatus(errCode)
			}
		}
	}
	c.JSON(http.StatusOK, propertyValues)
}
