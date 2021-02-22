package v1

import (
	mid "factors/middleware"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"net/http"

	C "factors/config"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

var FORCED_EVENT_NAMES = map[uint64][]string{
	215: []string{
		// Project ExpertRec.
		"cse.expertrec.com/payments/success",
	},
}

// GetEventNamesHandler godoc
// @Summary Get event names for the given project id.
// @Tags V1Api
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Success 200 {string} json "{"event_names": map[string][]string}"
// @Router /{project_id}/v1/event_names [get]
func GetEventNamesHandler(c *gin.Context) {

	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	logCtx := log.WithFields(log.Fields{
		"projectId": projectId,
	})

	// RedisGet is the only call. In case of Cache crash, job will be manually triggered to repopulate cache
	// No fallback for now.
	eventNames, err := store.GetStore().GetEventNamesOrderedByOccurenceAndRecency(projectId, 2500, C.GetLookbackWindowForEventUserCache())
	if err != nil {
		logCtx.WithError(err).Error("get event names ordered by occurence and recency")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if len(eventNames) == 0 {

		logCtx.WithError(err).Error(fmt.Sprintf("No Events Returned - ProjectID - %s", projectId))
	}

	// Force add specific events.
	if fNames, pExists := FORCED_EVENT_NAMES[projectId]; pExists {
		eventNames[U.FrequentlySeen] = append(eventNames[U.FrequentlySeen], fNames...)
	}

	// TODO: Janani Removing the IsExact property from output since its anyway backward compat with UI
	// Will remove exact/approx logic in UI as well
	c.JSON(http.StatusOK, gin.H{"event_names": eventNames})
}
