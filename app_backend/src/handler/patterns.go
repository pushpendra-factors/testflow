package handler

import (
	C "config"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// Test command.
// curl -i -X POST http://localhost:8080/projects/1/patterns/query -d '{ "start_event": "login", "end_event": "payment" }
func QueryPatternsHandler(c *gin.Context) {
	projectId, err := strconv.ParseUint(c.Params.ByName("project_id"), 10, 64)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var requestBodyMap map[string]interface{}
	if err := json.NewDecoder(c.Request.Body).Decode(&requestBodyMap); err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Query Patterns JSON Decoding failed.")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "json decoding : " + err.Error(),
			"status": http.StatusBadRequest,
		})
		return
	}

	var startEvent, endEvent string
	if se, ok := requestBodyMap["start_event"]; ok {
		startEvent = se.(string)
	}
	if ee, ok := requestBodyMap["end_event"]; ok {
		endEvent = ee.(string)
	}

	ps := C.GetServices().PatternService
	if results, err := ps.Query(projectId, startEvent, endEvent); err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Patterns query failed.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	} else {
		c.JSON(http.StatusOK, results)
	}
}

// Test command.
// curl -i -X POST http://localhost:8080/projects/1/patterns/crunch -d '{ "end_event": "payment" }

func CrunchPatternsHandler(c *gin.Context) {
	projectId, err := strconv.ParseUint(c.Params.ByName("project_id"), 10, 64)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var requestBodyMap map[string]interface{}
	if err := json.NewDecoder(c.Request.Body).Decode(&requestBodyMap); err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Query Patterns JSON Decoding failed.")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "json decoding : " + err.Error(),
			"status": http.StatusBadRequest,
		})
		return
	}

	var endEvent string
	if ee, ok := requestBodyMap["end_event"]; ok {
		endEvent = ee.(string)
	}

	ps := C.GetServices().PatternService
	if results, err := ps.Crunch(projectId, endEvent); err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Patterns crunch failed.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	} else {
		c.JSON(http.StatusOK, results)
	}
}
