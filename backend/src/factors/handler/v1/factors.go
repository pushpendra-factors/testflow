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
	event := c.Query("event")
	basePropertyKey := c.Query("basePropertyKey")
	basePropertyValue := c.Query("basePropertyValue")
	basePropertyOperator := false
	if c.Query("basePropertyOperator") == "=" {
		basePropertyOperator = true
	}
	basePropertyType := c.Query("basePropertyType")
	distributionProperty := c.Query("distributionProperty")

	ps, err := PW.NewPatternServiceWrapper("", projectId, modelId)
	if err != nil {
		logCtx.WithError(err).Error("Pattern Service initialization failed.")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  err.Error(),
			"status": http.StatusBadRequest,
		})
		return
	}
	type EventDistribution struct {
		Base   interface{} `json:"base"`
		After  interface{} `json:"after"`
		Before interface{} `json:"before"`
	}
	if patternMode == "EventDistribution" {
		res, res2, res3, _ := PW.BuildUserDistribution("", event, ps)
		finalResult := EventDistribution{
			Base:   res,
			Before: res2,
			After:  res3,
		}
		c.JSON(http.StatusOK, finalResult)
		return
	}
	if patternMode == "EventDistributionWithProperties" {
		distributionProperties := make(map[string]string)
		if distributionProperty != "" {
			distributionProperties[distributionProperty] = "categorical"
		}
		baseProperty := P.EventConstraints{}
		if basePropertyType == "numerical" {

		}
		if basePropertyType == "categorical" {
			baseProperty.UPCategoricalConstraints = append(baseProperty.UPCategoricalConstraints, P.CategoricalConstraint{
				PropertyName:  basePropertyKey,
				PropertyValue: basePropertyValue,
				Operator:      extractOperator(basePropertyOperator),
			})
		}
		res, _ := PW.BuildUserDistributionWithProperties("", event, baseProperty, distributionProperties, ps)
		c.JSON(http.StatusOK, res)
		return
	}
}

type CompareFactorsGoalParams struct {
	Name string                  `json:"name"`
	Rule []CreateGoalInputParams `json:"rule"`
}

func GetcompareFactorsGoalParams(c *gin.Context) (*CompareFactorsGoalParams, error) {
	params := CompareFactorsGoalParams{}
	err := c.BindJSON(&params)
	if err != nil {
		return nil, err
	}
	return &params, nil
}

func PostFactorsCompareHandler(c *gin.Context) {
	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	logCtx := log.WithFields(log.Fields{
		"projectId": projectId,
	})
	comparisonMode := c.Query("comparison_mode")
	if comparisonMode == "Week" {
		modelId1 := uint64(0)
		modelId2 := uint64(0)
		modelIdParam1 := c.Query("model_id1")
		modelIdParam2 := c.Query("model_id2")
		var err error
		if modelIdParam1 != "" && modelIdParam2 != "" {
			modelId1, err = strconv.ParseUint(modelIdParam1, 10, 64)
			modelId2, err = strconv.ParseUint(modelIdParam2, 10, 64)
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
		ps1, err := PW.NewPatternServiceWrapper("", projectId, modelId1)
		if err != nil {
			logCtx.WithError(err).Error("Pattern Service initialization failed.")
			c.JSON(http.StatusBadRequest, gin.H{
				"error":  err.Error(),
				"status": http.StatusBadRequest,
			})
			return
		}
		startConstraints, endConstraints := parseConstraints(params.Rule)
		var results1, results2 PW.Factors
		if results1, err, _ = PW.FactorV1("",
			projectId, params.StartEvent, startConstraints,
			params.EndEvent, endConstraints, P.COUNT_TYPE_PER_USER, ps1, "", nil); err != nil {
			logCtx.WithError(err).Error("Factors failed.")
			if err.Error() == "Root node not found or frequency 0" {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "No Insights Found"})
			}
			c.AbortWithStatus(http.StatusBadRequest)
			return
		} else {
			results1.Type = inputType
			results1.GoalRule = params

		}
		ps2, err := PW.NewPatternServiceWrapper("", projectId, modelId2)
		if err != nil {
			logCtx.WithError(err).Error("Pattern Service initialization failed.")
			c.JSON(http.StatusBadRequest, gin.H{
				"error":  err.Error(),
				"status": http.StatusBadRequest,
			})
			return
		}
		if results2, err, _ = PW.FactorV1("",
			projectId, params.StartEvent, startConstraints,
			params.EndEvent, endConstraints, P.COUNT_TYPE_PER_USER, ps2, "", nil); err != nil {
			logCtx.WithError(err).Error("Factors failed.")
			if err.Error() == "Root node not found or frequency 0" {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "No Insights Found"})
			}
			c.AbortWithStatus(http.StatusBadRequest)
			return
		} else {
			results2.Type = inputType
			results2.GoalRule = params

		}
		res := formatComparisonResults(results1, results2, true)
		c.JSON(http.StatusOK, res)
	}
	if comparisonMode == "Property" {
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
		inputType := c.Query("type")
		ipParams, err := GetcompareFactorsGoalParams(c)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		params1 := MapRule(ipParams.Rule[0])
		params2 := MapRule(ipParams.Rule[1])
		ps, err := PW.NewPatternServiceWrapper("", projectId, modelId)
		if err != nil {
			logCtx.WithError(err).Error("Pattern Service initialization failed.")
			c.JSON(http.StatusBadRequest, gin.H{
				"error":  err.Error(),
				"status": http.StatusBadRequest,
			})
			return
		}
		startConstraints, endConstraints := parseConstraints(params1.Rule)
		var results1, results2 PW.Factors
		if results1, err, _ = PW.FactorV1("",
			projectId, params1.StartEvent, startConstraints,
			params1.EndEvent, endConstraints, P.COUNT_TYPE_PER_USER, ps, "", nil); err != nil {
			logCtx.WithError(err).Error("Factors failed.")
			if err.Error() == "Root node not found or frequency 0" {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "No Insights Found"})
			}
			c.AbortWithStatus(http.StatusBadRequest)
			return
		} else {
			results1.Type = inputType
			results1.GoalRule = params1
		}
		startConstraints, endConstraints = parseConstraints(params2.Rule)
		if results2, err, _ = PW.FactorV1("",
			projectId, params2.StartEvent, startConstraints,
			params2.EndEvent, endConstraints, P.COUNT_TYPE_PER_USER, ps, "", nil); err != nil {
			logCtx.WithError(err).Error("Factors failed.")
			if err.Error() == "Root node not found or frequency 0" {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "No Insights Found"})
			}
			c.AbortWithStatus(http.StatusBadRequest)
			return
		} else {
			results2.Type = inputType
			results2.GoalRule = params2
		}
		res := formatComparisonResults(results1, results2, false)
		c.JSON(http.StatusOK, res)
	}
}

type ComparisonResult struct {
	FirstWeek        []float64 `json:"first"`
	SecondWeek       []float64 `json:"second"`
	PercentageChange float64
	Type             string
}

func formatComparisonResults(result1 PW.Factors, result2 PW.Factors, removeSingleInsight bool) map[string]ComparisonResult {
	result := make(map[string]ComparisonResult)
	var key string
	for _, insight := range result1.Insights {
		if insight.FactorsInsightsType == PW.JOURNEYTYPE || insight.FactorsInsightsType == PW.CAMPAIGNTYPE {
			key = insight.FactorsInsightsKey
		} else if insight.FactorsInsightsType == PW.ATTRIBUTETYPE {
			if insight.FactorsInsightsAttribute[0].FactorsAttributeValue != "" {
				key = insight.FactorsInsightsAttribute[0].FactorsAttributeKey + ":" + insight.FactorsInsightsAttribute[0].FactorsAttributeValue
			} else {
				key = insight.FactorsInsightsAttribute[0].FactorsAttributeKey + ":" + fmt.Sprintf("%v", insight.FactorsInsightsAttribute[0].FactorsAttributeLowerBound) + "-" + fmt.Sprintf("%v", insight.FactorsInsightsAttribute[0].FactorsAttributeUpperBound)
			}
		}
		result[key] = ComparisonResult{
			FirstWeek: []float64{insight.FactorsInsightsMultiplier, insight.FactorsInsightsPercentage, insight.FactorsInsightsUsersCount, insight.FactorsGoalUsersCount},
			Type:      insight.FactorsInsightsType,
		}
	}
	for _, insight := range result2.Insights {
		var key string
		if insight.FactorsInsightsType == PW.JOURNEYTYPE || insight.FactorsInsightsType == PW.CAMPAIGNTYPE {
			key = insight.FactorsInsightsKey
		} else if insight.FactorsInsightsType == PW.ATTRIBUTETYPE {
			if insight.FactorsInsightsAttribute[0].FactorsAttributeValue != "" {
				key = insight.FactorsInsightsAttribute[0].FactorsAttributeKey + ":" + insight.FactorsInsightsAttribute[0].FactorsAttributeValue
			} else {
				key = insight.FactorsInsightsAttribute[0].FactorsAttributeKey + ":" + fmt.Sprintf("%v", insight.FactorsInsightsAttribute[0].FactorsAttributeLowerBound) + "-" + fmt.Sprintf("%v", insight.FactorsInsightsAttribute[0].FactorsAttributeUpperBound)
			}
		}
		comparisonObj := ComparisonResult{}
		var percentagebase float64
		if val, ok := result[key]; ok {
			comparisonObj = val
			percentagebase = comparisonObj.FirstWeek[1]
		}
		comparisonObj.SecondWeek = []float64{insight.FactorsInsightsMultiplier, insight.FactorsInsightsPercentage, insight.FactorsInsightsUsersCount, insight.FactorsGoalUsersCount}
		if percentagebase != 0.0 {
			comparisonObj.PercentageChange = insight.FactorsInsightsPercentage / percentagebase
		}
		result[key] = comparisonObj
	}
	overallPercentageChange := result2.OverallPercentage / result1.OverallPercentage
	result["base"] = ComparisonResult{
		FirstWeek:        []float64{result1.OverallMultiplier, result1.OverallPercentage, result1.TotalUsersCount, result1.GoalUserCount},
		SecondWeek:       []float64{result2.OverallMultiplier, result2.OverallPercentage, result2.TotalUsersCount, result2.GoalUserCount},
		PercentageChange: overallPercentageChange,
	}
	if removeSingleInsight == true {
		filteredResult := make(map[string]ComparisonResult)
		for key, value := range result {
			if !(value.FirstWeek == nil || value.SecondWeek == nil) {
				filteredResult[key] = value
			}
		}
		return filteredResult
	}
	return result
}

func extractOperator(isEqual bool) string {
	var op string
	if isEqual == true {
		op = P.EQUALS_OPERATOR_CONST
	} else {
		op = P.NOT_EQUALS_OPERATOR_CONST
	}
	return op
}

func parseConstraints(filters M.FactorsGoalFilter) (*P.EventConstraints, *P.EventConstraints) {
	var startEventConstraints P.EventConstraints
	startEventConstraints.EPNumericConstraints = make([]P.NumericConstraint, 0)
	startEventConstraints.EPCategoricalConstraints = make([]P.CategoricalConstraint, 0)
	startEventConstraints.UPNumericConstraints = make([]P.NumericConstraint, 0)
	startEventConstraints.UPCategoricalConstraints = make([]P.CategoricalConstraint, 0)
	for _, filter := range filters.StartEnEventFitler {
		if filter.Type == "categorical" {
			op := extractOperator(filter.Operator)
			startEventConstraints.EPCategoricalConstraints = append(startEventConstraints.EPCategoricalConstraints, P.CategoricalConstraint{
				PropertyName:  filter.Key,
				PropertyValue: filter.Value,
				Operator:      op,
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
			op := extractOperator(filter.Operator)
			startEventConstraints.UPCategoricalConstraints = append(startEventConstraints.UPCategoricalConstraints, P.CategoricalConstraint{
				PropertyName:  filter.Key,
				PropertyValue: filter.Value,
				Operator:      op,
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
			op := extractOperator(filter.Operator)
			endEventConstraints.EPCategoricalConstraints = append(endEventConstraints.EPCategoricalConstraints, P.CategoricalConstraint{
				PropertyName:  filter.Key,
				PropertyValue: filter.Value,
				Operator:      op,
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
			op := extractOperator(filter.Operator)
			endEventConstraints.UPCategoricalConstraints = append(endEventConstraints.UPCategoricalConstraints, P.CategoricalConstraint{
				PropertyName:  filter.Key,
				PropertyValue: filter.Value,
				Operator:      op,
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
			op := extractOperator(filter.Operator)
			endEventConstraints.UPCategoricalConstraints = append(endEventConstraints.UPCategoricalConstraints, P.CategoricalConstraint{
				PropertyName:  filter.Key,
				PropertyValue: filter.Value,
				Operator:      op,
			})
			startEventConstraints.UPCategoricalConstraints = append(startEventConstraints.UPCategoricalConstraints, P.CategoricalConstraint{
				PropertyName:  filter.Key,
				PropertyValue: filter.Value,
				Operator:      op,
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
