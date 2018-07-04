package handler

import (
	C "config"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// Test command.
// curl -i -X GET http://localhost:8080/projects/1/patterns?start_event=login&end_event=payment
func QueryPatternsHandler(c *gin.Context) {
	projectId, err := strconv.ParseUint(c.Params.ByName("project_id"), 10, 64)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	qParams := c.Request.URL.Query()

	var startEvent string = ""
	startEvents := qParams["start_event"]
	if startEvents != nil {
		startEvent = startEvents[0]
	}
	var endEvent string = ""
	endEvents := qParams["end_event"]
	if endEvents != nil {
		endEvent = endEvents[0]
	}

	ps := C.GetServices().PatternService
	if results, err := ps.Query(projectId, startEvent, endEvent, 0, ^uint(0), 0, ^uint(0)); err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Patterns query failed.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	} else {
		c.JSON(http.StatusOK, results)
	}
}
