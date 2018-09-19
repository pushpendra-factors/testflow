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

	if query, ok := requestBodyMap["query"]; ok {
		log.WithFields(log.Fields{"query": query}).Info("Received query")
		response := map[string]interface{}{
			"query": query,
			"charts": []map[string]interface{}{
				map[string]interface{}{
					"type":   "line",
					"header": "Average PublicMessageSent per month",
					"labels": []string{"January", "February", "March", "April", "May", "June", "July"},
					"datasets": []map[string]interface{}{
						map[string]interface{}{
							"label": "Users with country:US",
							"data":  []float64{65, 59, 80, 81, 56, 55, 40},
						},
						map[string]interface{}{
							"label": "All Users",
							"data":  []float64{45, 70, 70, 101, 95, 5, 64},
						},
					},
				},
				map[string]interface{}{
					"type":   "bar",
					"header": "Users with country:US have 30% higer average PublicMessageSent than others.",
					"labels": []string{"All Users", "US", "India", "UK", "Australia", "Egypt", "Iran"},
					"datasets": []map[string]interface{}{
						map[string]interface{}{
							"label": "Average PublicMessageSent",
							"data":  []float64{65, 59, 80, 81, 56, 55, 40},
						},
						map[string]interface{}{
							"label": "Total PublicMessageSent",
							"data":  []float64{45, 70, 70, 101, 95, 5, 64},
						},
					},
				},
			},
		}
		c.JSON(http.StatusOK, response)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  fmt.Errorf("No query in request"),
			"status": http.StatusBadRequest,
		})
	}
}
