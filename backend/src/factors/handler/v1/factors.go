package v1

import (
	mid "factors/middleware"
	M "factors/model"
	P "factors/pattern"
	PW "factors/pattern_service_wrapper"
	U "factors/util"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// GetAllFactorsHandler - Factors handler
func PostFactorsHandler(c *gin.Context) {
	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	logCtx := log.WithFields(log.Fields{
		"projectId": projectId,
	})

	modelId := uint64(0)
	modelIdParam := c.Query("model_id")
	patternMode := c.Query("pattern_mode")
	propertyName := c.Query("debug_property_name")
	propertyValue := c.Query("debug_property_value")
	var err error
	if modelIdParam != "" {
		modelId, err = strconv.ParseUint(modelIdParam, 10, 64)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
	}
	inputType := c.Query("type")
	ipParams, err := GetcreateFactorsGoalParams(c)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	params := MapRule(ipParams.Rule)
	ps, err := PW.NewPatternServiceWrapper("", projectId, modelId)
	if err != nil {
		logCtx.WithError(err).Error("Pattern Service initialization failed.")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  err.Error(),
			"status": http.StatusBadRequest,
		})
		return
	}
	startConstraints, endConstraints := parseConstraints(params.Rule)
	if patternMode == "AllPatterns" {
		allEventPatterns := make([]string, 0)
		allPatterns, _ := ps.GetAllPatterns("", params.StartEvent, params.EndEvent)
		for _, eventPattern := range allPatterns {
			pattern := fmt.Sprintf("%v", eventPattern.PerUserCount)
			for _, eventName := range eventPattern.EventNames {
				pattern = pattern + "," + eventName
			}
			allEventPatterns = append(allEventPatterns, pattern)
		}
		c.JSON(http.StatusOK, allEventPatterns)
		return
	}
	if patternMode == "GetCount" {
		var count uint
		if params.StartEvent == "" {
			count, _ = ps.GetPerUserCount("", []string{params.EndEvent}, []P.EventConstraints{*endConstraints})
		} else {
			count, _ = ps.GetPerUserCount("", []string{params.StartEvent, params.EndEvent}, []P.EventConstraints{*startConstraints, *endConstraints})
		}
		c.JSON(http.StatusOK, count)
		return
	}
	if patternMode == "AllProperties" {
		var patternHistogram *P.Pattern
		if params.StartEvent == "" {
			patternHistogram = ps.GetPattern("", []string{params.EndEvent})
		} else {
			patternHistogram = ps.GetPattern("", []string{params.StartEvent, params.EndEvent})
		}
		c.JSON(http.StatusOK, patternHistogram)
		return
	}
	debugParams := make(map[string]string)
	debugParams["PropertyName"] = propertyName
	debugParams["PropertyValue"] = propertyValue
	if results, err, debugData := PW.FactorV1("",
		projectId, params.StartEvent, startConstraints,
		params.EndEvent, endConstraints, P.COUNT_TYPE_PER_USER, ps, patternMode, debugParams); err != nil {
		logCtx.WithError(err).Error("Factors failed.")
		if err.Error() == "Root node not found or frequency 0" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "No Insights Found"})
		}
		c.AbortWithStatus(http.StatusBadRequest)
		return
	} else {
		if patternMode != "" {
			c.JSON(http.StatusOK, debugData)
		}
		results.Type = inputType
		results.GoalRule = params
		c.JSON(http.StatusOK, results)
		return

	}
}

func GetFactorsHandler(c *gin.Context) {
	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	logCtx := log.WithFields(log.Fields{
		"projectId": projectId,
	})

	modelId := uint64(0)
	modelIdParam := c.Query("model_id")
	patternMode := c.Query("pattern_mode")
	var err error
	if modelIdParam != "" {
		modelId, err = strconv.ParseUint(modelIdParam, 10, 64)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
	}
	startEvent := c.Query("start_event")
	endEvent := c.Query("end_event")

	ps, err := PW.NewPatternServiceWrapper("", projectId, modelId)
	if err != nil {
		logCtx.WithError(err).Error("Pattern Service initialization failed.")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  err.Error(),
			"status": http.StatusBadRequest,
		})
		return
	}
	if patternMode == "AllPatterns" {
		allEventPatterns := make([]string, 0)
		allPatterns, _ := ps.GetAllPatterns("", startEvent, endEvent)
		for _, eventPattern := range allPatterns {
			pattern := ""
			for _, eventName := range eventPattern.EventNames {
				pattern = pattern + "," + eventName
			}
			allEventPatterns = append(allEventPatterns, pattern)
		}
		c.JSON(http.StatusOK, allEventPatterns)
		return
	}
}

func parseConstraints(filters M.FactorsGoalFilter) (*P.EventConstraints, *P.EventConstraints) {
	var startEventConstraints P.EventConstraints
	startEventConstraints.EPNumericConstraints = make([]P.NumericConstraint, 0)
	startEventConstraints.EPCategoricalConstraints = make([]P.CategoricalConstraint, 0)
	startEventConstraints.UPNumericConstraints = make([]P.NumericConstraint, 0)
	startEventConstraints.UPCategoricalConstraints = make([]P.CategoricalConstraint, 0)
	for _, filter := range filters.StartEnEventFitler {
		if filter.Type == "categorical" {
			startEventConstraints.EPCategoricalConstraints = append(startEventConstraints.EPCategoricalConstraints, P.CategoricalConstraint{
				PropertyName:  filter.Key,
				PropertyValue: filter.Value,
			})
		}
		if filter.Type == "numerical" {
			startEventConstraints.EPNumericConstraints = append(startEventConstraints.EPNumericConstraints, P.NumericConstraint{
				PropertyName: filter.Key,
				LowerBound:   filter.LowerBound,
				UpperBound:   filter.UpperBound,
				IsEquality:   filter.Operator,
			})
		}
	}
	for _, filter := range filters.StartEnUserFitler {
		if filter.Type == "categorical" {
			startEventConstraints.UPCategoricalConstraints = append(startEventConstraints.UPCategoricalConstraints, P.CategoricalConstraint{
				PropertyName:  filter.Key,
				PropertyValue: filter.Value,
			})
		}
		if filter.Type == "numerical" {
			startEventConstraints.UPNumericConstraints = append(startEventConstraints.UPNumericConstraints, P.NumericConstraint{
				PropertyName: filter.Key,
				LowerBound:   filter.LowerBound,
				UpperBound:   filter.UpperBound,
				IsEquality:   filter.Operator,
			})
		}
	}

	var endEventConstraints P.EventConstraints
	endEventConstraints.EPNumericConstraints = make([]P.NumericConstraint, 0)
	endEventConstraints.EPCategoricalConstraints = make([]P.CategoricalConstraint, 0)
	endEventConstraints.UPNumericConstraints = make([]P.NumericConstraint, 0)
	endEventConstraints.UPCategoricalConstraints = make([]P.CategoricalConstraint, 0)
	for _, filter := range filters.EndEnEventFitler {
		if filter.Type == "categorical" {
			endEventConstraints.EPCategoricalConstraints = append(endEventConstraints.EPCategoricalConstraints, P.CategoricalConstraint{
				PropertyName:  filter.Key,
				PropertyValue: filter.Value,
			})
		}
		if filter.Type == "numerical" {
			endEventConstraints.EPNumericConstraints = append(endEventConstraints.EPNumericConstraints, P.NumericConstraint{
				PropertyName: filter.Key,
				LowerBound:   filter.LowerBound,
				UpperBound:   filter.UpperBound,
				IsEquality:   filter.Operator,
			})
		}
	}
	for _, filter := range filters.EndEnUserFitler {
		if filter.Type == "categorical" {
			endEventConstraints.UPCategoricalConstraints = append(endEventConstraints.UPCategoricalConstraints, P.CategoricalConstraint{
				PropertyName:  filter.Key,
				PropertyValue: filter.Value,
			})
		}
		if filter.Type == "numerical" {
			endEventConstraints.UPNumericConstraints = append(endEventConstraints.UPNumericConstraints, P.NumericConstraint{
				PropertyName: filter.Key,
				LowerBound:   filter.LowerBound,
				UpperBound:   filter.UpperBound,
				IsEquality:   filter.Operator,
			})
		}
	}
	for _, filter := range filters.GlobalFilters {
		if filter.Type == "categorical" {
			endEventConstraints.UPCategoricalConstraints = append(endEventConstraints.UPCategoricalConstraints, P.CategoricalConstraint{
				PropertyName:  filter.Key,
				PropertyValue: filter.Value,
			})
			startEventConstraints.UPCategoricalConstraints = append(startEventConstraints.UPCategoricalConstraints, P.CategoricalConstraint{
				PropertyName:  filter.Key,
				PropertyValue: filter.Value,
			})
		}
		if filter.Type == "numerical" {
			endEventConstraints.UPNumericConstraints = append(endEventConstraints.UPNumericConstraints, P.NumericConstraint{
				PropertyName: filter.Key,
				LowerBound:   filter.LowerBound,
				UpperBound:   filter.UpperBound,
				IsEquality:   filter.Operator,
			})
			startEventConstraints.UPNumericConstraints = append(startEventConstraints.UPNumericConstraints, P.NumericConstraint{
				PropertyName: filter.Key,
				LowerBound:   filter.LowerBound,
				UpperBound:   filter.UpperBound,
				IsEquality:   filter.Operator,
			})
		}
	}
	return &startEventConstraints, &endEventConstraints
}
