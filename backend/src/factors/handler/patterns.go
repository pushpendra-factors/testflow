package handler

import (
	"encoding/json"
	mid "factors/middleware"
	"factors/model/model"
	P "factors/pattern"
	PW "factors/pattern_service_wrapper"
	U "factors/util"
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func ParseFactorQuery(query map[string]interface{}) (
	startEvent string, startEventConstraints *P.EventConstraints,
	endEvent string, endEventConstraints *P.EventConstraints,
	countType string, err error) {
	var startEventWithProperties map[string]interface{}
	var endEventWithProperties map[string]interface{}
	if queryType, ok := query["queryType"]; ok {
		if queryType == model.QueryTypeEventsOccurrence {
			countType = P.COUNT_TYPE_PER_OCCURRENCE
		} else if queryType == model.QueryTypeUniqueUsers {
			countType = P.COUNT_TYPE_PER_USER
		} else {
			err := fmt.Errorf(fmt.Sprintf("Unknown queryType %s in query", queryType))
			return "", nil, "", nil, "", err
		}
	} else {
		err := fmt.Errorf("Missing query type")
		return "", nil, "", nil, "", err
	}
	if ewp, ok := query["eventsWithProperties"]; ok {
		eventsWithProperties := ewp.([]interface{})
		numEvents := len(eventsWithProperties)
		if numEvents == 1 {
			endEventWithProperties = eventsWithProperties[0].(map[string]interface{})
		} else if numEvents == 2 {
			startEventWithProperties = eventsWithProperties[0].((map[string]interface{}))
			endEventWithProperties = eventsWithProperties[1].(map[string]interface{})
		} else {
			err := fmt.Errorf(fmt.Sprintf(
				"Unexpected number of events in query: %d", numEvents))
			return "", nil, "", nil, "", err
		}
	} else {
		err := fmt.Errorf("Missing eventsWithProperties")
		return "", nil, "", nil, "", err
	}
	endEvent, endEventConstraints, err = parseQueryEventWithProperties(endEventWithProperties)
	if err != nil {
		return "", nil, "", nil, "", err
	}
	if len(startEventWithProperties) == 0 {
		return "", nil, endEvent, endEventConstraints, countType, nil
	}
	startEvent, startEventConstraints, err = parseQueryEventWithProperties(startEventWithProperties)
	if err != nil {
		return "", nil, "", nil, "", err
	}
	return startEvent, startEventConstraints, endEvent, endEventConstraints, countType, nil
}

func parseQueryEventWithProperties(eventWithProperties map[string]interface{}) (
	string, *P.EventConstraints, error) {
	epNumericConstraintsMap := make(map[string]*P.NumericConstraint)
	epCategoricalConstraintsMap := make(map[string]*P.CategoricalConstraint)
	upNumericConstraintsMap := make(map[string]*P.NumericConstraint)
	upCategoricalConstraintsMap := make(map[string]*P.CategoricalConstraint)
	var err error
	eventName := eventWithProperties["name"].(string)

	// Event properties constraints.
	if properties, ok := eventWithProperties["properties"]; ok {
		err = addPropertyConstraintsToMap(
			properties.([]interface{}),
			&epNumericConstraintsMap,
			&epCategoricalConstraintsMap)
		if err != nil {
			return "", nil, err
		}
	}

	if properties, ok := eventWithProperties["user_properties"]; ok {
		err = addPropertyConstraintsToMap(
			properties.([]interface{}),
			&upNumericConstraintsMap,
			&upCategoricalConstraintsMap)
		if err != nil {
			return "", nil, err
		}
	}

	eventConstraints := &P.EventConstraints{
		EPNumericConstraints:     []P.NumericConstraint{},
		EPCategoricalConstraints: []P.CategoricalConstraint{},
		UPNumericConstraints:     []P.NumericConstraint{},
		UPCategoricalConstraints: []P.CategoricalConstraint{},
	}
	for _, value := range epNumericConstraintsMap {
		eventConstraints.EPNumericConstraints = append(
			eventConstraints.EPNumericConstraints, *value)
	}
	for _, value := range epCategoricalConstraintsMap {
		eventConstraints.EPCategoricalConstraints = append(
			eventConstraints.EPCategoricalConstraints, *value)
	}
	for _, value := range upNumericConstraintsMap {
		eventConstraints.UPNumericConstraints = append(
			eventConstraints.UPNumericConstraints, *value)
	}
	for _, value := range upCategoricalConstraintsMap {
		eventConstraints.UPCategoricalConstraints = append(
			eventConstraints.UPCategoricalConstraints, *value)
	}
	return eventName, eventConstraints, err
}

func addPropertyConstraintsToMap(
	properties []interface{},
	numericConstraintsMap *map[string]*P.NumericConstraint,
	categoricalConstraintsMap *map[string]*P.CategoricalConstraint) error {
	var err error
	for _, ep := range properties {
		p := ep.(map[string]interface{})
		propertyName := p["property"].(string)
		if p["type"].(string) == "numerical" {
			nC, ok := (*numericConstraintsMap)[propertyName]
			if !ok {
				nC = &P.NumericConstraint{
					PropertyName: propertyName,
					LowerBound:   -math.MaxFloat64,
					UpperBound:   math.MaxFloat64,
					IsEquality:   false,
				}
				(*numericConstraintsMap)[propertyName] = nC
			}
			numValue := p["value"].(float64)
			switch op := p["operator"].(string); op {
			case "equals":
				if numValue == float64(int64(numValue)) && numValue > 0 {
					// Most likely an positive integer.
					nC.LowerBound = numValue - 0.5
					nC.UpperBound = numValue + 0.5
				} else {
					// Automatically choose a small numeric range around numValue.
					numRange := math.Min(0.1, 0.01*numValue)
					nC.LowerBound = numValue - numRange
					nC.UpperBound = numValue + numRange
				}
				nC.IsEquality = true
			case "greaterThan":
				nC.LowerBound = numValue
			case "lesserThan":
				nC.UpperBound = numValue
			default:
				err = fmt.Errorf(fmt.Sprintf("Unknown Operator: %s", op))
			}
		} else if p["type"].(string) == "categorical" {
			cC, ok := (*categoricalConstraintsMap)[propertyName]
			if !ok {
				cC = &P.CategoricalConstraint{PropertyName: propertyName}
				(*categoricalConstraintsMap)[propertyName] = cC
			}
			switch op := p["operator"].(string); op {
			case "equals":
				cC.PropertyValue = p["value"].(string)
			default:
				err = fmt.Errorf(fmt.Sprintf("Unknown Operator: %s", op))
			}
		} else {
			err = fmt.Errorf(fmt.Sprintf("Unknown Property type: %s", p["type"].(string)))
		}
	}
	return err
}

// TODO(Ankit): Pass req id to subsequent calls to pattern server
// FactorHandler godoc
// @Summary To run factors model for the given query.
// @Tags Factors
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param query body object true "Factors query"
// @Success 200 {array} pattern_service_wrapper.FactorGraphResults
// @Router /{project_id}/factor [post]
// TODO(prateek): Check for a better way to define query.
func FactorHandler(c *gin.Context) {

	logCtx := log.WithFields(log.Fields{
		"reqId": U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
	})

	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	reqId := U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID)

	logCtx.WithFields(log.Fields{"projectId": projectId}).Debug("Factor Query")

	modelId := uint64(0)
	modelIdParam := c.Query("model_id")
	var err error
	if modelIdParam != "" {
		modelId, err = strconv.ParseUint(modelIdParam, 10, 64)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
	}

	var requestBodyMap map[string]interface{}
	if err := json.NewDecoder(c.Request.Body).Decode(&requestBodyMap); err != nil {
		logCtx.WithError(err).Error("Query Patterns JSON Decoding failed.")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "json decoding : " + err.Error(),
			"status": http.StatusBadRequest,
		})
		return
	}

	if query, ok := requestBodyMap["query"].(map[string]interface{}); ok {
		logCtx.WithFields(log.Fields{"query": query}).Debug("Received query")
		startEvent, startEventConstraints, endEvent, endEventConstraints, countType, err := ParseFactorQuery(query)
		if err != nil {
			log.WithError(err).Error("Invalid Query.")
			c.JSON(http.StatusBadRequest, gin.H{
				"error":  err.Error(),
				"status": http.StatusBadRequest,
			})
			return
		}
		logCtx.WithFields(log.Fields{
			"startEvent":            startEvent,
			"endEvent":              endEvent,
			"startEventConstraints": startEventConstraints,
			"endEventConstraints":   endEventConstraints,
			"countType":             countType,
		}).Debug("Factor query parse")

		ps, err := PW.NewPatternServiceWrapper(reqId, projectId, modelId)
		if err != nil {
			logCtx.WithError(err).Error("Pattern Service initialization failed.")
			c.JSON(http.StatusBadRequest, gin.H{
				"error":  err.Error(),
				"status": http.StatusBadRequest,
			})
			return
		}
		if results, err := PW.Factor(reqId,
			projectId, startEvent, startEventConstraints,
			endEvent, endEventConstraints, countType, ps); err != nil {
			logCtx.WithError(err).Error("Factors failed.")
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
