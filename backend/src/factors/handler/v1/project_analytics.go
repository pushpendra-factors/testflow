package v1

import (
	"encoding/json"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func GetFactorsAnalyticsHandler(c *gin.Context) {
	noOfDays := int(7)
	noOfDaysParam := c.Query("days")
	isHtmlRequired := c.Query("html")
	projectID := c.Query("projectID")
	var err error
	if noOfDaysParam != "" {
		noOfDays, err = strconv.Atoi(noOfDaysParam)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
	}

	var analytics map[string][]*model.ProjectAnalytics

	if projectID != "" {
		projIdInt, _ := strconv.Atoi(projectID)
		analytics, err = store.GetStore().GetEventUserCountsByProjectID(int64(projIdInt), noOfDays)

		if err != nil {
			log.WithError(err).Error("GetEventUserCountsByProjectID")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		if isHtmlRequired == "true" {
			resultJson, err := json.Marshal(analytics)
			if err != nil {
				return
			}

			var resultMap map[string][]map[string]interface{}
			err = json.Unmarshal(resultJson, &resultMap)
			if err != nil {
				log.Error(err)
				return
			}

			U.ReturnReadableHtmlFromMaps(c, resultMap, model.ProjectAnalyticsColumnsName, model.ProjectAnalyticsColumnsNameToJsonKeys)
			return
		}

	} else {
		analytics, err = store.GetStore().GetEventUserCountsOfAllProjects(noOfDays)

		if err != nil {
			log.WithError(err).Error("GetEventUserCountsOfAllProjects")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		if isHtmlRequired == "true" {

			resultJson, err := json.Marshal(analytics)
			if err != nil {
				return
			}

			var resultMap map[string][]map[string]interface{}
			err = json.Unmarshal(resultJson, &resultMap)
			if err != nil {
				log.Error(err)
				return
			}

			U.ReturnReadableHtmlFromMaps(c, resultMap, model.AllProjectAnalyticsColumnsName, model.ProjectAnalyticsColumnsNameToJsonKeys)
			return
		}
	}

	c.JSON(http.StatusOK, analytics)
}
