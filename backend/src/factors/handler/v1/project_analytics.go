package v1

import (
	"factors/model/store"
	"net/http"

	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func GetFactorsAnalyticsHandler(c *gin.Context) {
	noOfDays := int(7)
	noOfDaysParam := c.Query("days")
	var err error
	if noOfDaysParam != "" {
		noOfDays, err = strconv.Atoi(noOfDaysParam)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
	}
	analytics, err := store.GetStore().GetEventUserCountsOfAllProjects(noOfDays)
	if err != nil {
		log.WithError(err).Error("GetEventUserCountsOfAllProjects")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, analytics)
}
