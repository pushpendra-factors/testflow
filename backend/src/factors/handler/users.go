package handler

import (
	M "factors/model"
	crpc "factors/patternserver"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// Test command.
// curl -i -X GET http://localhost:8080/projects/1/users/bc7318e8-2b69-49b6-baf3-fdf47bcb1af9
func GetUserHandler(c *gin.Context) {
	projectId, err := strconv.ParseUint(c.Params.ByName("project_id"), 10, 64)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("GetUser Failed. ProjectId parse failed.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	id := c.Params.ByName("user_id")
	if id == "" {
		log.WithFields(log.Fields{"error": err}).Error("GetUser Failed. Missing id.")
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
	projectId, err := strconv.ParseUint(c.Params.ByName("project_id"), 10, 64)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	qParams := c.Request.URL.Query()

	var offset uint64 = 0
	offsets := qParams["offset"]
	if offsets != nil {
		offsetStr := offsets[0]
		if offsetParse, err := strconv.ParseUint(offsetStr, 10, 64); err != nil {
			log.WithFields(log.Fields{"error": err}).Error("GetUsers Failed. Offset parse failed.")
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
			log.WithFields(log.Fields{"error": err}).Error("GetUsers Failed. Limit parse failed.")
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
	projectID, err := strconv.ParseUint(c.Params.ByName("project_id"), 10, 64)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	modelId := uint64(0)

	modelIdParam := c.Query("model_id")
	if modelIdParam != "" {
		modelId, err = strconv.ParseUint(modelIdParam, 10, 64)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
	}

	properties, err := crpc.GetSeenUserProperties(projectID, modelId)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err, "projectId": projectID}).Error(
			"Get User Properties failed.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	c.JSON(http.StatusOK, properties)
}

// curl -i -X GET http://localhost:8080/projects/1/user_properties/$country
func GetUserPropertyValuesHandler(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Params.ByName("project_id"), 10, 64)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	modelId := uint64(0)

	modelIdParam := c.Query("model_id")
	if modelIdParam != "" {
		modelId, err = strconv.ParseUint(modelIdParam, 10, 64)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
	}

	propertyName := c.Params.ByName("property_name")
	if propertyName == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	propertyValues, err := crpc.GetSeenUserPropertyValues(projectID, modelId, propertyName)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"projectId":    projectID,
			"propertyName": propertyName}).Error(
			"Get User Properties failed.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	c.JSON(http.StatusOK, propertyValues)
}
