package handler

import (
	"encoding/json"
	mid "factors/middleware"
	M "factors/model"
	U "factors/util"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func ChannelQueryHandler(c *gin.Context) {
	r := c.Request

	logCtx := log.WithFields(log.Fields{
		"reqId": U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
	})

	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized,
			gin.H{"error": "Channel query failed. Invalid project."})
		return
	}

	logCtx = log.WithField("project_id", projectId)

	var queryPayload M.ChannelQuery

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&queryPayload); err != nil {
		logCtx.WithError(err).Error("Channel query failed. Json decode failed.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "Channel query failed. Json decode failed."})
		return
	}

	queryResult, errCode := M.ExecuteChannelQuery(projectId, &queryPayload)
	if errCode != http.StatusOK {
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "Channel query failed. Execution failure."})
		return
	}

	c.JSON(http.StatusOK, queryResult)
}

func GetChannelFilterValuesHandler(c *gin.Context) {
	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized,
			gin.H{"error": "Channel query failed. Invalid project."})
		return
	}

	channel := c.Query("channel")
	filter := c.Query("filter")
	if channel == "" || filter == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Missing params channel and filter."})
		return
	}

	filterValues, errCode := M.GetChannelFilterValues(projectId, channel, filter)
	if errCode != http.StatusFound {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Failed to get filter values for channel."})
		return
	}

	c.JSON(http.StatusFound, gin.H{"filter_values": filterValues})
}
