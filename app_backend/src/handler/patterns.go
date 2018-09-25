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

func parseFactorQuery(query map[string]interface{}) (string, string, float64, float64, error) {
	var startEvent string
	var endEvent string
	var endEventCardinalityLowerBound float64 = -1
	var endEventCardinalityUpperBound float64 = -1
	var err error

	if ewp, ok := query["eventsWithProperties"]; ok {
		eventsWithProperties := ewp.([]interface{})
		numEvents := len(eventsWithProperties)
		var endEventWithProperties map[string]interface{}
		if numEvents == 1 {
			endEventWithProperties = eventsWithProperties[0].(map[string]interface{})
		} else if numEvents == 2 {
			startEvent = eventsWithProperties[0].((map[string]interface{}))["name"].(string)
			endEventWithProperties = eventsWithProperties[1].(map[string]interface{})
		} else {
			err = fmt.Errorf(fmt.Sprintf(
				"Unexpected number of events in query: %d", numEvents))
		}
		endEvent = endEventWithProperties["name"].(string)
		for _, ep := range endEventWithProperties["properties"].([]interface{}) {
			p := ep.(map[string]interface{})
			if p["property"].(string) != "occurrence" {
				continue
			}
			switch op := p["operator"].(string); op {
			case "equals":
				tmp := p["value"].(float64)
				if tmp > 0 {
					endEventCardinalityLowerBound = tmp - 0.5
					endEventCardinalityUpperBound = tmp + 0.5
				}
			case "greaterThan":
				tmp := p["value"].(float64)
				if tmp > 0 {
					endEventCardinalityLowerBound = tmp
				}
			case "lesserThan":
				tmp := p["value"].(float64)
				if tmp > 0 {
					endEventCardinalityUpperBound = tmp
				}
			default:
				err = fmt.Errorf(fmt.Sprintf("Unknown Operator: %s", op))
			}
		}
	} else {
		err = fmt.Errorf("Missing eventsWithProperties")
	}

	if endEventCardinalityLowerBound > 0 &&
		endEventCardinalityUpperBound > 0 &&
		endEventCardinalityLowerBound > endEventCardinalityUpperBound {
		err = fmt.Errorf(fmt.Sprintf("Unexpected Input %d is greater than %d",
			endEventCardinalityLowerBound, endEventCardinalityUpperBound))
	}
	return startEvent, endEvent, endEventCardinalityLowerBound, endEventCardinalityUpperBound, err
}

func FactorHandler(c *gin.Context) {
	projectId, err := strconv.ParseUint(c.Params.ByName("project_id"), 10, 64)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	log.WithFields(log.Fields{"projectId": projectId}).Info("Factor Query")

	var requestBodyMap map[string]interface{}
	if err := json.NewDecoder(c.Request.Body).Decode(&requestBodyMap); err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Query Patterns JSON Decoding failed.")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "json decoding : " + err.Error(),
			"status": http.StatusBadRequest,
		})
		return
	}

	if query, ok := requestBodyMap["query"].(map[string]interface{}); ok {
		log.WithFields(log.Fields{"query": query}).Info("Received query")

		var startEvent, endEvent string
		var endEventCardinalityLowerBound, endEventCardinalityUpperBound float64
		var err error
		if startEvent, endEvent, endEventCardinalityLowerBound, endEventCardinalityUpperBound, err = parseFactorQuery(query); err != nil {
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
			"eecub":      endEventCardinalityUpperBound}).Info("Factor query parse")

		ps := C.GetServices().PatternService
		if results, err := ps.Factor(projectId, endEvent,
			int(endEventCardinalityLowerBound), int(endEventCardinalityUpperBound)); err != nil {
			log.WithFields(log.Fields{"error": err}).Error("Factors failed.")
			c.AbortWithStatus(http.StatusBadRequest)
			return
		} else {
			c.JSON(http.StatusOK, results)
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  fmt.Errorf("No query in request"),
			"status": http.StatusBadRequest,
		})
	}
}
