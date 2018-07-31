package handler

import (
	C "config"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func parseEventQuery(requestBodyMap map[string]interface{}) (string, string, int, int, error) {
	var startEvent string
	if se, ok := requestBodyMap["start_event"]; ok {
		if sen, ok := se.(map[string]interface{})["name"]; ok {
			startEvent = sen.(string)
		}
	}

	var endEvent string
	var endEventCardinalityLowerBound int = -1
	var endEventCardinalityUpperBound int = -1
	if ee, ok := requestBodyMap["end_event"]; ok {
		eeMap := ee.(map[string]interface{})
		if een, ok := eeMap["name"]; ok {
			endEvent = een.(string)
		}
		if eeclb, ok := eeMap["card_lower_bound"]; ok {
			tmp := int(eeclb.(float64))
			if tmp > 0 {
				endEventCardinalityLowerBound = tmp
			}
		}
		if eecub, ok := eeMap["card_upper_bound"]; ok {
			tmp := int(eecub.(float64))
			if tmp > 0 {
				endEventCardinalityUpperBound = tmp
			}
		}
	}

	var err error
	if endEventCardinalityLowerBound > 0 &&
		endEventCardinalityUpperBound > 0 &&
		endEventCardinalityLowerBound > endEventCardinalityUpperBound {
		err = fmt.Errorf(fmt.Sprintf("Unexpected Input %d is greater than %d",
			endEventCardinalityLowerBound, endEventCardinalityUpperBound))
	}
	return startEvent, endEvent, endEventCardinalityLowerBound, endEventCardinalityUpperBound, err
}

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
	var endEventCardinalityLowerBound, endEventCardinalityUpperBound int
	if startEvent, endEvent, endEventCardinalityLowerBound, endEventCardinalityUpperBound, err = parseEventQuery(requestBodyMap); err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Invalid Query.")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  err.Error(),
			"status": http.StatusBadRequest,
		})
		return
	}

	log.WithFields(log.Fields{
		"startEvent": startEvent,
		"endEvent":   endEvent,
		"eeclb":      endEventCardinalityLowerBound,
		"eecub":      endEventCardinalityUpperBound}).Info("Pattern query.")

	ps := C.GetServices().PatternService
	if results, err := ps.Query(projectId, startEvent, endEvent,
		endEventCardinalityLowerBound, endEventCardinalityUpperBound); err != nil {
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

	var startEvent, endEvent string
	var endEventCardinalityLowerBound, endEventCardinalityUpperBound int
	if startEvent, endEvent, endEventCardinalityLowerBound, endEventCardinalityUpperBound, err = parseEventQuery(requestBodyMap); err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Invalid Query.")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  err.Error(),
			"status": http.StatusBadRequest,
		})
		return
	}
	log.WithFields(log.Fields{
		"startEvent": startEvent,
		"endEvent":   endEvent,
		"eeclb":      endEventCardinalityLowerBound,
		"eecub":      endEventCardinalityUpperBound}).Info("Pattern crunch query")

	ps := C.GetServices().PatternService
	if results, err := ps.Crunch(projectId, endEvent,
		endEventCardinalityLowerBound, endEventCardinalityUpperBound); err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Patterns crunch failed.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	} else {
		c.JSON(http.StatusOK, results)
	}
}
