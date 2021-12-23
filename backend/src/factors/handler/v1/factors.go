package v1

import (
	"encoding/json"
	mid "factors/middleware"
	"factors/model/model"
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

	if projectId == 611 {
		fac := PW.Factors{}
		if err := json.Unmarshal([]byte(returnConstantData()), &fac); err != nil {
			return
		}
		c.JSON(http.StatusOK, fac)
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
		userInfo := ps.GetUserAndEventsInfo()
		c.JSON(http.StatusOK, userInfo)
		return
	}
	if patternMode == "AllHistogram" {
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
		if err.Error() == "root node not found or frequency 0" {
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

type UserDistributionInput struct {
	Events []CreateFactorsGoalParams `json:"events"`
}

func GetUserDistributionParams(c *gin.Context) ([]CreateFactorsGoalParams, error) {
	params := UserDistributionInput{}
	err := c.BindJSON(&params)
	if err != nil {
		return nil, err
	}
	return params.Events, nil
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
	ipParams, err := GetUserDistributionParams(c)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	ps, err := PW.NewPatternServiceWrapper("", projectId, modelId)
	if err != nil {
		logCtx.WithError(err).Error("Pattern Service initialization failed.")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  err.Error(),
			"status": http.StatusBadRequest,
		})
		return
	}
	params1 := MapRule(ipParams[0].Rule)
	startConstraints1, endConstraints1 := parseConstraints(params1.Rule)
	params2 := MapRule(ipParams[1].Rule)
	startConstraints2, endConstraints2 := parseConstraints(params2.Rule)
	type EventComparison struct {
		First  uint `json:"first"`
		Second uint `json:"second"`
	}
	type EventDistribution struct {
		Base   EventComparison            `json:"base"`
		After  map[string]EventComparison `json:"after"`
		Before map[string]EventComparison `json:"before"`
	}
	type PropDistribution struct {
		Base EventComparison            `json:"base"`
		Prop map[string]EventComparison `json:"prop"`
	}
	if patternMode == "EventDistribution" {
		base1, before1, after1, _ := PW.BuildUserDistribution("", params1.EndEvent, endConstraints1, ps)
		base2, before2, after2 := uint(0), make(map[string]uint), make(map[string]uint)
		if params2.EndEvent != "" {
			base2, before2, after2, _ = PW.BuildUserDistribution("", params2.EndEvent, endConstraints2, ps)
		}
		finalResult := EventDistribution{
			Base: EventComparison{
				First:  base1,
				Second: base2,
			},
			After:  make(map[string]EventComparison),
			Before: make(map[string]EventComparison),
		}
		for key, value := range before1 {
			data, exists := finalResult.Before[key]
			if !exists {
				finalResult.Before[key] = EventComparison{}
			}
			data.First = value
			finalResult.Before[key] = data
		}
		if params2.EndEvent != "" {
			for key, value := range before2 {
				data, exists := finalResult.Before[key]
				if !exists {
					finalResult.Before[key] = EventComparison{}
				}
				data.Second = value
				finalResult.Before[key] = data
			}
		}
		for key, value := range after1 {
			data, exists := finalResult.After[key]
			if !exists {
				finalResult.After[key] = EventComparison{}
			}
			data.First = value
			finalResult.After[key] = data
		}
		if params2.EndEvent != "" {
			for key, value := range after2 {
				data, exists := finalResult.After[key]
				if !exists {
					finalResult.After[key] = EventComparison{}
				}
				data.Second = value
				finalResult.After[key] = data
			}
		}
		c.JSON(http.StatusOK, finalResult)
		return
	}
	if patternMode == "UserDistribution" {
		events1 := make([]string, 0)
		if params1.StartEvent != "" {
			events1 = append(events1, params1.StartEvent)
		}
		if params1.EndEvent != "" {
			events1 = append(events1, params1.EndEvent)
		}
		overall1, propDist1, _ := PW.BuildUserDistributionWithProperties("", events1, startConstraints1, endConstraints1, ps)
		overall2, propDist2 := uint(0), make(map[string]uint)
		events2 := make([]string, 0)
		if params2.StartEvent != "" {
			events2 = append(events2, params2.StartEvent)
		}
		if params2.EndEvent != "" {
			events2 = append(events2, params2.EndEvent)
		}
		if params2.EndEvent != "" {
			overall2, propDist2, _ = PW.BuildUserDistributionWithProperties("", events2, startConstraints2, endConstraints2, ps)
		}
		finalResult := PropDistribution{
			Base: EventComparison{
				First:  overall1,
				Second: overall2,
			},
			Prop: make(map[string]EventComparison),
		}
		for key, value := range propDist1 {
			data, exists := finalResult.Prop[key]
			if !exists {
				finalResult.Prop[key] = EventComparison{}
			}
			data.First = value
			finalResult.Prop[key] = data
		}
		if params2.EndEvent != "" {
			for key, value := range propDist2 {
				data, exists := finalResult.Prop[key]
				if !exists {
					finalResult.Prop[key] = EventComparison{}
				}
				data.Second = value
				finalResult.Prop[key] = data
			}
		}
		c.JSON(http.StatusOK, finalResult)
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
			if err.Error() == "root node not found or frequency 0" {
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
			if err.Error() == "root node not found or frequency 0" {
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
			if err.Error() == "root node not found or frequency 0" {
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
			if err.Error() == "root node not found or frequency 0" {
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

func parseConstraints(filters model.FactorsGoalFilter) (*P.EventConstraints, *P.EventConstraints) {
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

func returnConstantData() string {
	factorsOutput := "{\"goal\":{\"st_en\":\"Opportunity Created\",\"en_en\":\"Deal Won\",\"st_time\":\"2020-10-01T00:00:00Z\",\"en_time\":\"2020-10-31T00:00:00Z\"},\"goal_user_count\":337,\"total_users_count\":712,\"overall_percentage\":47.33,\"overall_multiplier\":1,\"overall_percentage_text\":\"47.33% of all users have completed this goal\",\"insights\":[{\"factors_insights_key\":\"LinkedIn_Global_BOF_Remarketing\",\"factors_insights_text\":\"of which visitors from the campaign <a>LinkedIn_Global_BOF_Remarketing</a> show 1.51x goal completion\",\"factors_insights_multiplier\":1.51,\"factors_insights_percentage\":71.5,\"factors_insights_users_count\":169,\"factors_goal_users_count\":121,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"campaign\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\",\"factors_sub_insights\":[{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Country\",\"factors_attribute_value\":\"US\"}],\"factors_insights_text\":\"where <a>Country=US</a> show 1.2x goal completion\",\"factors_insights_multiplier\":1.2,\"factors_insights_percentage\":56.8,\"factors_insights_users_count\":93,\"factors_goal_users_count\":53,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Country\",\"factors_attribute_value\":\"Brazil\"}],\"factors_insights_text\":\"where <a>Country=Brazil</a> show 0.8x goal completion\",\"factors_insights_multiplier\":0.8,\"factors_insights_percentage\":37.9,\"factors_insights_users_count\":39,\"factors_goal_users_count\":15,\"factors_multiplier_increase_flag\":false,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Source/Medium\",\"factors_attribute_value\":\"Google Organic\"}],\"factors_insights_text\":\"where <a>Source/Medium=Google Organic</a> show 1.1x goal completion\",\"factors_insights_multiplier\":1.1,\"factors_insights_percentage\":52.1,\"factors_insights_users_count\":88,\"factors_goal_users_count\":46,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Page Time Spent\",\"factors_attribute_value\":\"<= 1\"}],\"factors_insights_text\":\"where <a>Page Time Spent<= 1</a> show 0.7x goal completion\",\"factors_insights_multiplier\":0.7,\"factors_insights_percentage\":33.1,\"factors_insights_users_count\":99,\"factors_goal_users_count\":33,\"factors_multiplier_increase_flag\":false,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Page Time Spent \",\"factors_attribute_value\":\">5\"}],\"factors_insights_text\":\"where <a>Page Time Spent >5</a> show 1.3x goal completion\",\"factors_insights_multiplier\":1.3,\"factors_insights_percentage\":61.5,\"factors_insights_users_count\":81,\"factors_goal_users_count\":50,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"}]},{\"factors_insights_key\":\"LinkedIn_NA_BOF_Remarketing\",\"factors_insights_text\":\"of which visitors from the campaign <a>LinkedIn_NA_BOF_Remarketing</a> show 0.6x goal completion\",\"factors_insights_multiplier\":0.6,\"factors_insights_percentage\":28.4,\"factors_insights_users_count\":151,\"factors_goal_users_count\":43,\"factors_multiplier_increase_flag\":false,\"factors_insights_type\":\"campaign\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_key\":\"Display_NA_BOF_Remarketing_CaseStudy\",\"factors_insights_text\":\"of which visitors from the campaign <a>Display_NA_BOF_Remarketing_CaseStudy</a> show 1.6x goal completion\",\"factors_insights_multiplier\":1.6,\"factors_insights_percentage\":75.7,\"factors_insights_users_count\":84,\"factors_goal_users_count\":64,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"campaign\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_key\":\"Display_Global_BOF_Remarketing_CaseStudy\",\"factors_insights_text\":\"of which visitors from the campaign <a>Display_Global_BOF_Remarketing_CaseStudy</a> show 1.2x goal completion\",\"factors_insights_multiplier\":1.2,\"factors_insights_percentage\":56.8,\"factors_insights_users_count\":228,\"factors_goal_users_count\":130,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"campaign\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\",\"factors_sub_insights\":[{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Country\",\"factors_attribute_value\":\"India\"}],\"factors_insights_text\":\"where <a>Country=India</a> show 0.6x goal completion\",\"factors_insights_multiplier\":0.6,\"factors_insights_percentage\":28.4,\"factors_insights_users_count\":66,\"factors_goal_users_count\":19,\"factors_multiplier_increase_flag\":false,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Company_Industry\",\"factors_attribute_value\":\"FinTech\"}],\"factors_insights_text\":\"where <a>Company_Industry=FinTech</a> show 1.6x goal completion\",\"factors_insights_multiplier\":1.6,\"factors_insights_percentage\":75.7,\"factors_insights_users_count\":23,\"factors_goal_users_count\":18,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Company_Industry\",\"factors_attribute_value\":\"Software\"}],\"factors_insights_text\":\"where <a>Company_Industry=Software</a> show 0.7x goal completion\",\"factors_insights_multiplier\":0.7,\"factors_insights_percentage\":33.1,\"factors_insights_users_count\":72,\"factors_goal_users_count\":24,\"factors_multiplier_increase_flag\":false,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Company_Revenue\",\"factors_attribute_value\":\"<5 mn\"}],\"factors_insights_text\":\"where <a>Company_Revenue<5 mn</a> show 1.4x goal completion\",\"factors_insights_multiplier\":1.4,\"factors_insights_percentage\":66.3,\"factors_insights_users_count\":25,\"factors_goal_users_count\":17,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Company_Revenue\",\"factors_attribute_value\":\"6-20 mn\"}],\"factors_insights_text\":\"where <a>Company_Revenue=6-20 mn</a> show 0.6x goal completion\",\"factors_insights_multiplier\":0.6,\"factors_insights_percentage\":28.4,\"factors_insights_users_count\":186,\"factors_goal_users_count\":53,\"factors_multiplier_increase_flag\":false,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Tech_stack\",\"factors_attribute_value\":\"Marketo\"}],\"factors_insights_text\":\"where <a>Tech_stack=Marketo</a> show 0.5x goal completion\",\"factors_insights_multiplier\":0.5,\"factors_insights_percentage\":23.7,\"factors_insights_users_count\":42,\"factors_goal_users_count\":10,\"factors_multiplier_increase_flag\":false,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Seniority\",\"factors_attribute_value\":\"VP\"}],\"factors_insights_text\":\"where <a>Seniority=VP</a> show 1.3x goal completion\",\"factors_insights_multiplier\":1.3,\"factors_insights_percentage\":61.5,\"factors_insights_users_count\":58,\"factors_goal_users_count\":36,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"}]},{\"factors_insights_key\":\"RollWorks_Global_Intent_Display\",\"factors_insights_text\":\"of which visitors from the campaign <a>RollWorks_Global_Intent_Display</a> show 1.8x goal completion\",\"factors_insights_multiplier\":1.8,\"factors_insights_percentage\":85.2,\"factors_insights_users_count\":91,\"factors_goal_users_count\":78,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"campaign\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_key\":\"Email_Webinar_MarketingAnalytics_DaveGerhard\",\"factors_insights_text\":\"of which visitors from the campaign <a>Email_Webinar_MarketingAnalytics_DaveGerhard</a> show 1.7x goal completion\",\"factors_insights_multiplier\":1.7,\"factors_insights_percentage\":80.5,\"factors_insights_users_count\":114,\"factors_goal_users_count\":92,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"campaign\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_key\":\"Source+Medium= LinkedIn Paid\",\"factors_insights_text\":\"of which visitors from the campaign <a>Source+Medium= LinkedIn Paid</a> show 1.1x goal completion\",\"factors_insights_multiplier\":1.1,\"factors_insights_percentage\":52.1,\"factors_insights_users_count\":351,\"factors_goal_users_count\":183,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"campaign\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_key\":\"Source+Medium= Google Paid\",\"factors_insights_text\":\"of which visitors from the campaign <a>Source+Medium= Google Paid</a> show 1.3x goal completion\",\"factors_insights_multiplier\":1.3,\"factors_insights_percentage\":61.5,\"factors_insights_users_count\":326,\"factors_goal_users_count\":201,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"campaign\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\",\"factors_sub_insights\":[{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Campaign\",\"factors_attribute_value\":\"RollWorks_Global_Intent_Display\"}],\"factors_insights_text\":\"where <a>Campaign=RollWorks_Global_Intent_Display</a> show 1.4x goal completion\",\"factors_insights_multiplier\":1.4,\"factors_insights_percentage\":66.3,\"factors_insights_users_count\":78,\"factors_goal_users_count\":52,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Source+Medium\",\"factors_attribute_value\":\"Referral G2Crowd\"}],\"factors_insights_text\":\"where <a>Source+Medium=Referral G2Crowd</a> show 1.5x goal completion\",\"factors_insights_multiplier\":1.5,\"factors_insights_percentage\":71,\"factors_insights_users_count\":43,\"factors_goal_users_count\":31,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Company_Revenue\",\"factors_attribute_value\":\"21-50 mn\"}],\"factors_insights_text\":\"where <a>Company_Revenue=21-50 mn</a> show 0.7x goal completion\",\"factors_insights_multiplier\":0.7,\"factors_insights_percentage\":33.1,\"factors_insights_users_count\":102,\"factors_goal_users_count\":34,\"factors_multiplier_increase_flag\":false,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Country\",\"factors_attribute_value\":\"Australia & New Zealand\"}],\"factors_insights_text\":\"where <a>Country=Australia & New Zealand</a> show 1.1x goal completion\",\"factors_insights_multiplier\":1.1,\"factors_insights_percentage\":52.1,\"factors_insights_users_count\":34,\"factors_goal_users_count\":18,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Country\",\"factors_attribute_value\":\"France\"}],\"factors_insights_text\":\"where <a>Country=France</a> show 0.6x goal completion\",\"factors_insights_multiplier\":0.6,\"factors_insights_percentage\":28.4,\"factors_insights_users_count\":45,\"factors_goal_users_count\":13,\"factors_multiplier_increase_flag\":false,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_key\":\"www.acme.com/pricing/\",\"factors_insights_text\":\"of which users who perform <a>www.acme.com/pricing/</a> show 1.1x goal completion\",\"factors_insights_multiplier\":1.1,\"factors_insights_percentage\":52.1,\"factors_insights_users_count\":287,\"factors_goal_users_count\":150,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"journey\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_key\":\"eBook Downloaded - The Ultimate Guide to SaaS Marketing Analytics 2021\",\"factors_insights_text\":\"of which users who perform <a>eBook Downloaded - The Ultimate Guide to SaaS Marketing Analytics 2021</a> show 1.2x goal completion\",\"factors_insights_multiplier\":1.2,\"factors_insights_percentage\":56.8,\"factors_insights_users_count\":80,\"factors_goal_users_count\":46,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"journey\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_key\":\"www.acme.com/case-study/ecommerce/\",\"factors_insights_text\":\"of which users who perform <a>www.acme.com/case-study/ecommerce/</a> show 0.3x goal completion\",\"factors_insights_multiplier\":0.3,\"factors_insights_percentage\":14.2,\"factors_insights_users_count\":91,\"factors_goal_users_count\":13,\"factors_multiplier_increase_flag\":false,\"factors_insights_type\":\"journey\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Page Count\",\"factors_attribute_value\":\">4\"}],\"factors_insights_text\":\"where <a>Page Count>4</a> show 1.4x goal completion\",\"factors_insights_multiplier\":1.4,\"factors_insights_percentage\":66.3,\"factors_insights_users_count\":129,\"factors_goal_users_count\":86,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Page Count \",\"factors_attribute_value\":\"<2\"}],\"factors_insights_text\":\"where <a>Page Count <2</a> show 0.7x goal completion\",\"factors_insights_multiplier\":0.7,\"factors_insights_percentage\":33.1,\"factors_insights_users_count\":120,\"factors_goal_users_count\":40,\"factors_multiplier_increase_flag\":false,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"}]},{\"factors_insights_key\":\"Source+Medium= RollWorks Paid\",\"factors_insights_text\":\"of which visitors from the campaign <a>Source+Medium= RollWorks Paid</a> show 1.8x goal completion\",\"factors_insights_multiplier\":1.8,\"factors_insights_percentage\":85.2,\"factors_insights_users_count\":91,\"factors_goal_users_count\":78,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"campaign\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_key\":\"Source+Medium= Direct None\",\"factors_insights_text\":\"of which visitors from the campaign <a>Source+Medium= Direct None</a> show 1.1x goal completion\",\"factors_insights_multiplier\":1.1,\"factors_insights_percentage\":52.1,\"factors_insights_users_count\":602,\"factors_goal_users_count\":314,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"campaign\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_key\":\"Source+Medium= Google Organic\",\"factors_insights_text\":\"of which visitors from the campaign <a>Source+Medium= Google Organic</a> show 1x goal completion\",\"factors_insights_multiplier\":1,\"factors_insights_percentage\":47.3,\"factors_insights_users_count\":621,\"factors_goal_users_count\":294,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"campaign\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_key\":\"Source+Medium= Referral G2Crowd\",\"factors_insights_text\":\"of which visitors from the campaign <a>Source+Medium= Referral G2Crowd</a> show 1.8x goal completion\",\"factors_insights_multiplier\":1.8,\"factors_insights_percentage\":85.2,\"factors_insights_users_count\":65,\"factors_goal_users_count\":56,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"campaign\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_key\":\"Source+Medium= Email Case Studies\",\"factors_insights_text\":\"of which visitors from the campaign <a>Source+Medium= Email Case Studies</a> show 1.4x goal completion\",\"factors_insights_multiplier\":1.4,\"factors_insights_percentage\":66.3,\"factors_insights_users_count\":319,\"factors_goal_users_count\":212,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"campaign\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\",\"factors_sub_insights\":[{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Company_Industry\",\"factors_attribute_value\":\"FinTech\"}],\"factors_insights_text\":\"where <a>Company_Industry=FinTech</a> show 1.2x goal completion\",\"factors_insights_multiplier\":1.2,\"factors_insights_percentage\":56.8,\"factors_insights_users_count\":42,\"factors_goal_users_count\":24,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Company_Industry\",\"factors_attribute_value\":\"Software\"}],\"factors_insights_text\":\"where <a>Company_Industry=Software</a> show 0.7x goal completion\",\"factors_insights_multiplier\":0.7,\"factors_insights_percentage\":33.1,\"factors_insights_users_count\":129,\"factors_goal_users_count\":43,\"factors_multiplier_increase_flag\":false,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Company_Revenue\",\"factors_attribute_value\":\"50mn+\"}],\"factors_insights_text\":\"where <a>Company_Revenue=50mn+</a> show 1.1x goal completion\",\"factors_insights_multiplier\":1.1,\"factors_insights_percentage\":52.1,\"factors_insights_users_count\":46,\"factors_goal_users_count\":24,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Company_Revenue\",\"factors_attribute_value\":\"<5 mn\"}],\"factors_insights_text\":\"where <a>Company_Revenue<5 mn</a> show 0.4x goal completion\",\"factors_insights_multiplier\":0.4,\"factors_insights_percentage\":18.9,\"factors_insights_users_count\":148,\"factors_goal_users_count\":28,\"factors_multiplier_increase_flag\":false,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Product_plan\",\"factors_attribute_value\":\"Enterprise\"}],\"factors_insights_text\":\"where <a>Product_plan=Enterprise</a> show 1.2x goal completion\",\"factors_insights_multiplier\":1.2,\"factors_insights_percentage\":56.8,\"factors_insights_users_count\":51,\"factors_goal_users_count\":29,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Product_plan\",\"factors_attribute_value\":\"Growth\"}],\"factors_insights_text\":\"where <a>Product_plan=Growth</a> show 0.7x goal completion\",\"factors_insights_multiplier\":0.7,\"factors_insights_percentage\":33.1,\"factors_insights_users_count\":453,\"factors_goal_users_count\":150,\"factors_multiplier_increase_flag\":false,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Competitor_evaluating\",\"factors_attribute_value\":\"Bizible\"}],\"factors_insights_text\":\"where <a>Competitor_evaluating=Bizible</a> show 1.3x goal completion\",\"factors_insights_multiplier\":1.3,\"factors_insights_percentage\":61.5,\"factors_insights_users_count\":115,\"factors_goal_users_count\":71,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Contact_Department\",\"factors_attribute_value\":\"Marketing Operations\"}],\"factors_insights_text\":\"where <a>Contact_Department=Marketing Operations</a> show 1.1x goal completion\",\"factors_insights_multiplier\":1.1,\"factors_insights_percentage\":52.1,\"factors_insights_users_count\":161,\"factors_goal_users_count\":84,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"}]},{\"factors_insights_key\":\"Source+Medium= Email Webinar\",\"factors_insights_text\":\"of which visitors from the campaign <a>Source+Medium= Email Webinar</a> show 1.1x goal completion\",\"factors_insights_multiplier\":1.1,\"factors_insights_percentage\":52.1,\"factors_insights_users_count\":230,\"factors_goal_users_count\":120,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"campaign\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_key\":\"Adgroup/Adset = Drift_Testimonial_Images\",\"factors_insights_text\":\"of which visitors from the campaign <a>Adgroup/Adset = Drift_Testimonial_Images</a> show 0.8x goal completion\",\"factors_insights_multiplier\":0.8,\"factors_insights_percentage\":37.9,\"factors_insights_users_count\":84,\"factors_goal_users_count\":32,\"factors_multiplier_increase_flag\":false,\"factors_insights_type\":\"campaign\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_key\":\"Adgroup/Adset = MarketingAnalytics_Webinar_DaveGerhardt_Images\",\"factors_insights_text\":\"of which visitors from the campaign <a>Adgroup/Adset = MarketingAnalytics_Webinar_DaveGerhardt_Images</a> show 1.7x goal completion\",\"factors_insights_multiplier\":1.7,\"factors_insights_percentage\":80.5,\"factors_insights_users_count\":83,\"factors_goal_users_count\":67,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"campaign\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_key\":\"Adgroup/Adset = GuestBlog_ROAS_Paid_Ads_by_ChrisWalker_Images\",\"factors_insights_text\":\"of which visitors from the campaign <a>Adgroup/Adset = GuestBlog_ROAS_Paid_Ads_by_ChrisWalker_Images</a> show 1.6x goal completion\",\"factors_insights_multiplier\":1.6,\"factors_insights_percentage\":75.7,\"factors_insights_users_count\":71,\"factors_goal_users_count\":54,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"campaign\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Company_Industry\",\"factors_attribute_value\":\"eCommerce\"}],\"factors_insights_text\":\"of which visitors with <a>Company_Industry=eCommerce</a> show 0.5x goal completion\",\"factors_insights_multiplier\":0.5,\"factors_insights_percentage\":23.7,\"factors_insights_users_count\":80,\"factors_goal_users_count\":19,\"factors_multiplier_increase_flag\":false,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Company_Industry\",\"factors_attribute_value\":\"FinTech\"}],\"factors_insights_text\":\"of which visitors with <a>Company_Industry=FinTech</a> show 1.1x goal completion\",\"factors_insights_multiplier\":1.1,\"factors_insights_percentage\":52.1,\"factors_insights_users_count\":88,\"factors_goal_users_count\":46,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Company_Industry\",\"factors_attribute_value\":\"Software\"}],\"factors_insights_text\":\"of which visitors with <a>Company_Industry=Software</a> show 1.6x goal completion\",\"factors_insights_multiplier\":1.6,\"factors_insights_percentage\":75.7,\"factors_insights_users_count\":122,\"factors_goal_users_count\":93,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\",\"factors_sub_insights\":[{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Company_Revenue\",\"factors_attribute_value\":\"6-20 mn\"}],\"factors_insights_text\":\"where <a>Company_Revenue=6-20 mn</a> show 1.2x goal completion\",\"factors_insights_multiplier\":1.2,\"factors_insights_percentage\":56.8,\"factors_insights_users_count\":68,\"factors_goal_users_count\":39,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Company_Revenue\",\"factors_attribute_value\":\"21-50 mn\"}],\"factors_insights_text\":\"where <a>Company_Revenue=21-50 mn</a> show 0.7x goal completion\",\"factors_insights_multiplier\":0.7,\"factors_insights_percentage\":33.1,\"factors_insights_users_count\":87,\"factors_goal_users_count\":29,\"factors_multiplier_increase_flag\":false,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Source+Medium\",\"factors_attribute_value\":\"Referral G2Crowd\"}],\"factors_insights_text\":\"where <a>Source+Medium=Referral G2Crowd</a> show 1x goal completion\",\"factors_insights_multiplier\":1,\"factors_insights_percentage\":47.3,\"factors_insights_users_count\":44,\"factors_goal_users_count\":21,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_key\":\"Whitepaper Download - AI & ML in Marketing Analytics 2021\",\"factors_insights_text\":\"where <a>=</a> show 1.2x goal completion\",\"factors_insights_multiplier\":1.2,\"factors_insights_percentage\":56.8,\"factors_insights_users_count\":77,\"factors_goal_users_count\":44,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"journey\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_key\":\"eBook Downloaded - The Ultimate Guide to SaaS Marketing Analytics 2021\",\"factors_insights_text\":\"where <a>=</a> show 0.8x goal completion\",\"factors_insights_multiplier\":0.8,\"factors_insights_percentage\":37.9,\"factors_insights_users_count\":81,\"factors_goal_users_count\":31,\"factors_multiplier_increase_flag\":false,\"factors_insights_type\":\"journey\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_key\":\"www.acme.com/blog/all-about-marketing-analytics-for-saas/\",\"factors_insights_text\":\"where <a>=</a> show 0.4x goal completion\",\"factors_insights_multiplier\":0.4,\"factors_insights_percentage\":18.9,\"factors_insights_users_count\":74,\"factors_goal_users_count\":14,\"factors_multiplier_increase_flag\":false,\"factors_insights_type\":\"journey\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"}]},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Company_Industry\",\"factors_attribute_value\":\"EdTech\"}],\"factors_insights_text\":\"of which visitors with <a>Company_Industry=EdTech</a> show 1x goal completion\",\"factors_insights_multiplier\":1,\"factors_insights_percentage\":47.3,\"factors_insights_users_count\":154,\"factors_goal_users_count\":73,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Company_Revenue\",\"factors_attribute_value\":\"<5 mn\"}],\"factors_insights_text\":\"of which visitors with <a>Company_Revenue<5 mn</a> show 0.6x goal completion\",\"factors_insights_multiplier\":0.6,\"factors_insights_percentage\":28.4,\"factors_insights_users_count\":144,\"factors_goal_users_count\":41,\"factors_multiplier_increase_flag\":false,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Company_Revenue\",\"factors_attribute_value\":\"6-20 mn\"}],\"factors_insights_text\":\"of which visitors with <a>Company_Revenue=6-20 mn</a> show 1.4x goal completion\",\"factors_insights_multiplier\":1.4,\"factors_insights_percentage\":66.3,\"factors_insights_users_count\":217,\"factors_goal_users_count\":144,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\",\"factors_sub_insights\":[{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Tech_stack\",\"factors_attribute_value\":\"Hubspot, Drift\"}],\"factors_insights_text\":\"where <a>Tech_stack=Hubspot, Drift</a> show 1.2x goal completion\",\"factors_insights_multiplier\":1.2,\"factors_insights_percentage\":56.8,\"factors_insights_users_count\":79,\"factors_goal_users_count\":45,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Company_Industry\",\"factors_attribute_value\":\"EdTech\"}],\"factors_insights_text\":\"where <a>Company_Industry=EdTech</a> show 0.8x goal completion\",\"factors_insights_multiplier\":0.8,\"factors_insights_percentage\":37.9,\"factors_insights_users_count\":65,\"factors_goal_users_count\":25,\"factors_multiplier_increase_flag\":false,\"factors_insights_type\":\"\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Competitor_evaluating\",\"factors_attribute_value\":\"Metadata.io\"}],\"factors_insights_text\":\"where <a>Competitor_evaluating=Metadata.io</a> show 1.3x goal completion\",\"factors_insights_multiplier\":1.3,\"factors_insights_percentage\":61.5,\"factors_insights_users_count\":60,\"factors_goal_users_count\":37,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Source+Medium\",\"factors_attribute_value\":\"Google Paid\"}],\"factors_insights_text\":\"where <a>Source+Medium=Google Paid</a> show 1x goal completion\",\"factors_insights_multiplier\":1,\"factors_insights_percentage\":47.3,\"factors_insights_users_count\":116,\"factors_goal_users_count\":55,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"}]},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Company_Revenue\",\"factors_attribute_value\":\"21-50 mn\"}],\"factors_insights_text\":\"of which visitors with <a>Company_Revenue=21-50 mn</a> show 1.1x goal completion\",\"factors_insights_multiplier\":1.1,\"factors_insights_percentage\":52.1,\"factors_insights_users_count\":186,\"factors_goal_users_count\":97,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Company_Revenue\",\"factors_attribute_value\":\"50mn+\"}],\"factors_insights_text\":\"of which visitors with <a>Company_Revenue=50mn+</a> show 0.7x goal completion\",\"factors_insights_multiplier\":0.7,\"factors_insights_percentage\":33.1,\"factors_insights_users_count\":166,\"factors_goal_users_count\":55,\"factors_multiplier_increase_flag\":false,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Number_contacts_in_company\",\"factors_attribute_value\":\"44198\"}],\"factors_insights_text\":\"of which visitors with <a>Number_contacts_in_company=44198</a> show 1.4x goal completion\",\"factors_insights_multiplier\":1.4,\"factors_insights_percentage\":66.3,\"factors_insights_users_count\":90,\"factors_goal_users_count\":60,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Number_contacts_in_company\",\"factors_attribute_value\":\"8+\"}],\"factors_insights_text\":\"of which visitors with <a>Number_contacts_in_company=8+</a> show 0.4x goal completion\",\"factors_insights_multiplier\":0.4,\"factors_insights_percentage\":18.9,\"factors_insights_users_count\":132,\"factors_goal_users_count\":25,\"factors_multiplier_increase_flag\":false,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Tech_stack\",\"factors_attribute_value\":\"Hubspot, Drift\"}],\"factors_insights_text\":\"of which visitors with <a>Tech_stack=Hubspot, Drift</a> show 1.1x goal completion\",\"factors_insights_multiplier\":1.1,\"factors_insights_percentage\":52.1,\"factors_insights_users_count\":230,\"factors_goal_users_count\":120,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Tech_stack\",\"factors_attribute_value\":\"Hotjar\"}],\"factors_insights_text\":\"of which visitors with <a>Tech_stack=Hotjar</a> show 1.9x goal completion\",\"factors_insights_multiplier\":1.9,\"factors_insights_percentage\":89.9,\"factors_insights_users_count\":65,\"factors_goal_users_count\":59,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Tech_stack\",\"factors_attribute_value\":\"Marketo\"}],\"factors_insights_text\":\"of which visitors with <a>Tech_stack=Marketo</a> show 0.7x goal completion\",\"factors_insights_multiplier\":0.7,\"factors_insights_percentage\":33.1,\"factors_insights_users_count\":123,\"factors_goal_users_count\":41,\"factors_multiplier_increase_flag\":false,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Product_plan\",\"factors_attribute_value\":\"Startup\"}],\"factors_insights_text\":\"of which visitors with <a>Product_plan=Startup</a> show 1.1x goal completion\",\"factors_insights_multiplier\":1.1,\"factors_insights_percentage\":52.1,\"factors_insights_users_count\":168,\"factors_goal_users_count\":88,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Product_plan\",\"factors_attribute_value\":\"Enterprise\"}],\"factors_insights_text\":\"of which visitors with <a>Product_plan=Enterprise</a> show 0.7x goal completion\",\"factors_insights_multiplier\":0.7,\"factors_insights_percentage\":33.1,\"factors_insights_users_count\":229,\"factors_goal_users_count\":76,\"factors_multiplier_increase_flag\":false,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Seniority\",\"factors_attribute_value\":\"VP\"}],\"factors_insights_text\":\"of which visitors with <a>Seniority=VP</a> show 1.6x goal completion\",\"factors_insights_multiplier\":1.6,\"factors_insights_percentage\":75.7,\"factors_insights_users_count\":73,\"factors_goal_users_count\":56,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Seniority\",\"factors_attribute_value\":\"Manager\"}],\"factors_insights_text\":\"of which visitors with <a>Seniority=Manager</a> show 1x goal completion\",\"factors_insights_multiplier\":1,\"factors_insights_percentage\":47.3,\"factors_insights_users_count\":397,\"factors_goal_users_count\":188,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Seniority\",\"factors_attribute_value\":\"Associate\"}],\"factors_insights_text\":\"of which visitors with <a>Seniority=Associate</a> show 0.6x goal completion\",\"factors_insights_multiplier\":0.6,\"factors_insights_percentage\":28.4,\"factors_insights_users_count\":66,\"factors_goal_users_count\":19,\"factors_multiplier_increase_flag\":false,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Competitor_evaluating\",\"factors_attribute_value\":\"Mixpanel\"}],\"factors_insights_text\":\"of which visitors with <a>Competitor_evaluating=Mixpanel</a> show 0.6x goal completion\",\"factors_insights_multiplier\":0.6,\"factors_insights_percentage\":28.4,\"factors_insights_users_count\":70,\"factors_goal_users_count\":20,\"factors_multiplier_increase_flag\":false,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Competitor_evaluating\",\"factors_attribute_value\":\"Metadata.io\"}],\"factors_insights_text\":\"of which visitors with <a>Competitor_evaluating=Metadata.io</a> show 1.6x goal completion\",\"factors_insights_multiplier\":1.6,\"factors_insights_percentage\":75.7,\"factors_insights_users_count\":105,\"factors_goal_users_count\":80,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Competitor_evaluating\",\"factors_attribute_value\":\"Bizible\"}],\"factors_insights_text\":\"of which visitors with <a>Competitor_evaluating=Bizible</a> show 1.5x goal completion\",\"factors_insights_multiplier\":1.5,\"factors_insights_percentage\":71,\"factors_insights_users_count\":149,\"factors_goal_users_count\":106,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\",\"factors_sub_insights\":[{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Country\",\"factors_attribute_value\":\"US\"}],\"factors_insights_text\":\"where <a>Country=US</a> show 1.2x goal completion\",\"factors_insights_multiplier\":1.2,\"factors_insights_percentage\":56.8,\"factors_insights_users_count\":58,\"factors_goal_users_count\":33,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Country\",\"factors_attribute_value\":\"UK\"}],\"factors_insights_text\":\"where <a>Country=UK</a> show 0.7x goal completion\",\"factors_insights_multiplier\":0.7,\"factors_insights_percentage\":33.1,\"factors_insights_users_count\":48,\"factors_goal_users_count\":16,\"factors_multiplier_increase_flag\":false,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Contact_Department\",\"factors_attribute_value\":\"Marketing Operations\"}],\"factors_insights_text\":\"where <a>Contact_Department=Marketing Operations</a> show 1.1x goal completion\",\"factors_insights_multiplier\":1.1,\"factors_insights_percentage\":52.1,\"factors_insights_users_count\":128,\"factors_goal_users_count\":67,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Tech_stack\",\"factors_attribute_value\":\"Marketo\"}],\"factors_insights_text\":\"where <a>Tech_stack=Marketo</a> show 0.8x goal completion\",\"factors_insights_multiplier\":0.8,\"factors_insights_percentage\":37.9,\"factors_insights_users_count\":79,\"factors_goal_users_count\":30,\"factors_multiplier_increase_flag\":false,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_key\":\"www.acme.com/pricing/\",\"factors_insights_text\":\"of which users who perform <a>www.acme.com/pricing/</a> show 1.3x goal completion\",\"factors_insights_multiplier\":1.3,\"factors_insights_percentage\":61.5,\"factors_insights_users_count\":82,\"factors_goal_users_count\":51,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"journey\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_key\":\"Signed up for The Marketing Analytics Academy\",\"factors_insights_text\":\"of which users who perform <a>Signed up for The Marketing Analytics Academy</a> show 1x goal completion\",\"factors_insights_multiplier\":1,\"factors_insights_percentage\":47.3,\"factors_insights_users_count\":141,\"factors_goal_users_count\":67,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"journey\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_key\":\"Webinar Attended - Dave Gerhard - Marketing Analytics for B2B SaaS in 2021\",\"factors_insights_text\":\"of which users who perform <a>Webinar Attended - Dave Gerhard - Marketing Analytics for B2B SaaS in 2021</a> show 1.3x goal completion\",\"factors_insights_multiplier\":1.3,\"factors_insights_percentage\":61.5,\"factors_insights_users_count\":126,\"factors_goal_users_count\":78,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"journey\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"}]},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Country\",\"factors_attribute_value\":\"US\"}],\"factors_insights_text\":\"of which visitors with <a>Country=US</a> show 1x goal completion\",\"factors_insights_multiplier\":1,\"factors_insights_percentage\":47.3,\"factors_insights_users_count\":200,\"factors_goal_users_count\":95,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\",\"factors_sub_insights\":[{\"factors_insights_key\":\"www.acme.com/compliance/\",\"factors_insights_text\":\"of which users who perform <a>www.acme.com/compliance/</a> show 0.9x goal completion\",\"factors_insights_multiplier\":0.9,\"factors_insights_percentage\":42.6,\"factors_insights_users_count\":157,\"factors_goal_users_count\":67,\"factors_multiplier_increase_flag\":false,\"factors_insights_type\":\"journey\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_key\":\"Webinar Attended - Optimizing Marketing ROAS for B2B SaaS\",\"factors_insights_text\":\"of which users who perform <a>Webinar Attended - Optimizing Marketing ROAS for B2B SaaS</a> show 1.8x goal completion\",\"factors_insights_multiplier\":1.8,\"factors_insights_percentage\":85.2,\"factors_insights_users_count\":36,\"factors_goal_users_count\":31,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"journey\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"}]},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Country\",\"factors_attribute_value\":\"UK\"}],\"factors_insights_text\":\"of which visitors with <a>Country=UK</a> show 1.9x goal completion\",\"factors_insights_multiplier\":1.9,\"factors_insights_percentage\":89.9,\"factors_insights_users_count\":50,\"factors_goal_users_count\":45,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Country\",\"factors_attribute_value\":\"Australia & New Zealand\"}],\"factors_insights_text\":\"of which visitors with <a>Country=Australia & New Zealand</a> show 0.7x goal completion\",\"factors_insights_multiplier\":0.7,\"factors_insights_percentage\":33.1,\"factors_insights_users_count\":102,\"factors_goal_users_count\":34,\"factors_multiplier_increase_flag\":false,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Country\",\"factors_attribute_value\":\"India\"}],\"factors_insights_text\":\"of which visitors with <a>Country=India</a> show 0.6x goal completion\",\"factors_insights_multiplier\":0.6,\"factors_insights_percentage\":28.4,\"factors_insights_users_count\":102,\"factors_goal_users_count\":29,\"factors_multiplier_increase_flag\":false,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Contact_Department\",\"factors_attribute_value\":\"Demand Generation\"}],\"factors_insights_text\":\"of which visitors with <a>Contact_Department=Demand Generation</a> show 1.1x goal completion\",\"factors_insights_multiplier\":1.1,\"factors_insights_percentage\":52.1,\"factors_insights_users_count\":193,\"factors_goal_users_count\":101,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Contact_Department\",\"factors_attribute_value\":\"Marketing Operations\"}],\"factors_insights_text\":\"of which visitors with <a>Contact_Department=Marketing Operations</a> show 1.3x goal completion\",\"factors_insights_multiplier\":1.3,\"factors_insights_percentage\":61.5,\"factors_insights_users_count\":204,\"factors_goal_users_count\":126,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Contact_Department\",\"factors_attribute_value\":\"Analytics\"}],\"factors_insights_text\":\"of which visitors with <a>Contact_Department=Analytics</a> show 0.5x goal completion\",\"factors_insights_multiplier\":0.5,\"factors_insights_percentage\":23.7,\"factors_insights_users_count\":189,\"factors_goal_users_count\":45,\"factors_multiplier_increase_flag\":false,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_key\":\"www.acme.com/\",\"factors_insights_text\":\"of which users who perform <a>www.acme.com/</a> show 1.4x goal completion\",\"factors_insights_multiplier\":1.4,\"factors_insights_percentage\":66.3,\"factors_insights_users_count\":443,\"factors_goal_users_count\":294,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"journey\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_key\":\"www.acme.com/pricing/\",\"factors_insights_text\":\"of which users who perform <a>www.acme.com/pricing/</a> show 1.4x goal completion\",\"factors_insights_multiplier\":1.2,\"factors_insights_percentage\":56.8,\"factors_insights_users_count\":529,\"factors_goal_users_count\":301,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"journey\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\",\"factors_sub_insights\":[{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Competitor_evaluating\",\"factors_attribute_value\":\"Bizible\"}],\"factors_insights_text\":\"where <a>Competitor_evaluating=Bizible</a> show 1.6x goal completion\",\"factors_insights_multiplier\":1.6,\"factors_insights_percentage\":75.7,\"factors_insights_users_count\":67,\"factors_goal_users_count\":51,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Competitor_evaluating\",\"factors_attribute_value\":\"Mixpanel\"}],\"factors_insights_text\":\"where <a>Competitor_evaluating=Mixpanel</a> show 0.5x goal completion\",\"factors_insights_multiplier\":0.5,\"factors_insights_percentage\":23.7,\"factors_insights_users_count\":75,\"factors_goal_users_count\":18,\"factors_multiplier_increase_flag\":false,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Company_Industry\",\"factors_attribute_value\":\"Software\"}],\"factors_insights_text\":\"where <a>Company_Industry=Software</a> show 1.2x goal completion\",\"factors_insights_multiplier\":1.2,\"factors_insights_percentage\":56.8,\"factors_insights_users_count\":128,\"factors_goal_users_count\":73,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Product_plan\",\"factors_attribute_value\":\"Startup\"}],\"factors_insights_text\":\"where <a>Product_plan=Startup</a> show 1.1x goal completion\",\"factors_insights_multiplier\":1.1,\"factors_insights_percentage\":52.1,\"factors_insights_users_count\":113,\"factors_goal_users_count\":59,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Seniority\",\"factors_attribute_value\":\"VP\"}],\"factors_insights_text\":\"where <a>Seniority=VP</a> show 1.4x goal completion\",\"factors_insights_multiplier\":1.4,\"factors_insights_percentage\":66.3,\"factors_insights_users_count\":67,\"factors_goal_users_count\":45,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Company_Revenue\",\"factors_attribute_value\":\"6-20 mn\"}],\"factors_insights_text\":\"where <a>Company_Revenue=6-20 mn</a> show 1.2x goal completion\",\"factors_insights_multiplier\":1.2,\"factors_insights_percentage\":56.8,\"factors_insights_users_count\":218,\"factors_goal_users_count\":124,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Company_Revenue\",\"factors_attribute_value\":\"21-50 mn\"}],\"factors_insights_text\":\"where <a>Company_Revenue=21-50 mn</a> show 0.8x goal completion\",\"factors_insights_multiplier\":0.8,\"factors_insights_percentage\":37.9,\"factors_insights_users_count\":176,\"factors_goal_users_count\":67,\"factors_multiplier_increase_flag\":false,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"}]},{\"factors_insights_key\":\"www.acme.com/customers\",\"factors_insights_text\":\"of which users who perform <a>www.acme.com/customers</a> show 1x goal completion\",\"factors_insights_multiplier\":1,\"factors_insights_percentage\":47.3,\"factors_insights_users_count\":270,\"factors_goal_users_count\":128,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"journey\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_key\":\"www.acme.com/resources/\",\"factors_insights_text\":\"of which users who perform <a>www.acme.com/resources/</a> show 1.9x goal completion\",\"factors_insights_multiplier\":1.9,\"factors_insights_percentage\":89.9,\"factors_insights_users_count\":303,\"factors_goal_users_count\":273,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"journey\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_key\":\"www.acme.com/compliance/\",\"factors_insights_text\":\"of which users who perform <a>www.acme.com/compliance/</a> show 1x goal completion\",\"factors_insights_multiplier\":1,\"factors_insights_percentage\":47.3,\"factors_insights_users_count\":655,\"factors_goal_users_count\":310,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"journey\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_key\":\"Whitepaper Download - AI & ML in Marketing Analytics 2021\",\"factors_insights_text\":\"of which users who perform <a>Whitepaper Download - AI & ML in Marketing Analytics 2021</a> show 1.9x goal completion\",\"factors_insights_multiplier\":1.9,\"factors_insights_percentage\":89.9,\"factors_insights_users_count\":88,\"factors_goal_users_count\":80,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"journey\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_key\":\"www.acme.com/blog/all-about-marketing-analytics-for-saas/\",\"factors_insights_text\":\"of which users who perform <a>www.acme.com/blog/all-about-marketing-analytics-for-saas/</a> show 1.5x goal completion\",\"factors_insights_multiplier\":1.5,\"factors_insights_percentage\":71,\"factors_insights_users_count\":80,\"factors_goal_users_count\":57,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"journey\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_key\":\"Webinar Attended - Optimizing Marketing ROAS for B2B SaaS\",\"factors_insights_text\":\"of which users who perform <a>Webinar Attended - Optimizing Marketing ROAS for B2B SaaS</a> show 1.8x goal completion\",\"factors_insights_multiplier\":1.8,\"factors_insights_percentage\":85.2,\"factors_insights_users_count\":86,\"factors_goal_users_count\":74,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"journey\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_key\":\"MQL Email Sequence 5 - Integration Steps & Guide Received\",\"factors_insights_text\":\"of which users who perform <a>MQL Email Sequence 5 - Integration Steps & Guide Received</a> show 0.8x goal completion\",\"factors_insights_multiplier\":0.8,\"factors_insights_percentage\":37.9,\"factors_insights_users_count\":390,\"factors_goal_users_count\":148,\"factors_multiplier_increase_flag\":false,\"factors_insights_type\":\"journey\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_key\":\"MQL Email Sequence 5 - Call With Acme's Customer Sucess Team for Integrations Received\",\"factors_insights_text\":\"of which users who perform <a>MQL Email Sequence 5 - Call With Acme's Customer Sucess Team for Integrations Received</a> show 1.2x goal completion\",\"factors_insights_multiplier\":1.2,\"factors_insights_percentage\":56.8,\"factors_insights_users_count\":332,\"factors_goal_users_count\":189,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"journey\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_key\":\"Signed up for The Marketing Analytics Academy\",\"factors_insights_text\":\"of which users who perform <a>Signed up for The Marketing Analytics Academy</a> show 1.5x goal completion\",\"factors_insights_multiplier\":1.5,\"factors_insights_percentage\":71,\"factors_insights_users_count\":259,\"factors_goal_users_count\":184,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"journey\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\",\"factors_sub_insights\":[{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Contact_Department\",\"factors_attribute_value\":\"Demand Generation\"}],\"factors_insights_text\":\"where <a>Contact_Department=Demand Generation</a> show 1.1x goal completion\",\"factors_insights_multiplier\":1.1,\"factors_insights_percentage\":52.1,\"factors_insights_users_count\":107,\"factors_goal_users_count\":56,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Contact_Department\",\"factors_attribute_value\":\"Analytics\"}],\"factors_insights_text\":\"where <a>Contact_Department=Analytics</a> show 0.3x goal completion\",\"factors_insights_multiplier\":0.3,\"factors_insights_percentage\":14.2,\"factors_insights_users_count\":288,\"factors_goal_users_count\":41,\"factors_multiplier_increase_flag\":false,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Seniority\",\"factors_attribute_value\":\"VP\"}],\"factors_insights_text\":\"where <a>Seniority=VP</a> show 0.7x goal completion\",\"factors_insights_multiplier\":0.7,\"factors_insights_percentage\":33.1,\"factors_insights_users_count\":36,\"factors_goal_users_count\":12,\"factors_multiplier_increase_flag\":false,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Seniority\",\"factors_attribute_value\":\"Manager\"}],\"factors_insights_text\":\"where <a>Seniority=Manager</a> show 1.3x goal completion\",\"factors_insights_multiplier\":1.3,\"factors_insights_percentage\":61.5,\"factors_insights_users_count\":146,\"factors_goal_users_count\":90,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Company_Industry\",\"factors_attribute_value\":\"FinTech\"}],\"factors_insights_text\":\"where <a>Company_Industry=FinTech</a> show 0.9x goal completion\",\"factors_insights_multiplier\":0.9,\"factors_insights_percentage\":42.6,\"factors_insights_users_count\":82,\"factors_goal_users_count\":35,\"factors_multiplier_increase_flag\":false,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_key\":\"Webinar Attended - Dave Gerhard - Marketing Analytics for B2B SaaS in 2021\",\"factors_insights_text\":\"of which users who perform <a>Webinar Attended - Dave Gerhard - Marketing Analytics for B2B SaaS in 2021</a> show 1.2x goal completion\",\"factors_insights_multiplier\":1.2,\"factors_insights_percentage\":56.8,\"factors_insights_users_count\":278,\"factors_goal_users_count\":158,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"journey\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Company_Industry\",\"factors_attribute_value\":\"EdTech\"}],\"factors_insights_text\":\"where <a>Company_Industry=EdTech</a> show 0.7x goal completion\",\"factors_insights_multiplier\":0.7,\"factors_insights_percentage\":33.1,\"factors_insights_users_count\":148,\"factors_goal_users_count\":49,\"factors_multiplier_increase_flag\":false,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"}]},{\"factors_insights_key\":\"eBook Downloaded - The Ultimate Guide to SaaS Marketing Analytics 2021\",\"factors_insights_text\":\"of which users who perform <a>eBook Downloaded - The Ultimate Guide to SaaS Marketing Analytics 2021</a> show 1.5x goal completion\",\"factors_insights_multiplier\":1.5,\"factors_insights_percentage\":71,\"factors_insights_users_count\":111,\"factors_goal_users_count\":79,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"journey\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_key\":\"Webinar Attended - Dave Gerhard - Marketing Analytics for B2B SaaS in 2021\",\"factors_insights_text\":\"of which users who perform <a>Webinar Attended - Dave Gerhard - Marketing Analytics for B2B SaaS in 2021</a> show 1.6x goal completion\",\"factors_insights_multiplier\":1.6,\"factors_insights_percentage\":75.7,\"factors_insights_users_count\":324,\"factors_goal_users_count\":246,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"journey\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\",\"factors_sub_insights\":[{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Source+Medium\",\"factors_attribute_value\":\"Google Paid\"}],\"factors_insights_text\":\"where <a>Source+Medium=Google Paid</a> show 0.8x goal completion\",\"factors_insights_multiplier\":0.8,\"factors_insights_percentage\":37.9,\"factors_insights_users_count\":211,\"factors_goal_users_count\":80,\"factors_multiplier_increase_flag\":false,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_key\":\"Email_Webinar_MarketingAnalytics_DaveGerhard\",\"factors_insights_text\":\"of which users who perform <a>Email_Webinar_MarketingAnalytics_DaveGerhard</a> show 1.1x goal completion\",\"factors_insights_multiplier\":1.1,\"factors_insights_percentage\":52.1,\"factors_insights_users_count\":176,\"factors_goal_users_count\":92,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"journey\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Seniority\",\"factors_attribute_value\":\"VP\"}],\"factors_insights_text\":\"where <a>Seniority=VP</a> show 1.2x goal completion\",\"factors_insights_multiplier\":1.2,\"factors_insights_percentage\":56.8,\"factors_insights_users_count\":59,\"factors_goal_users_count\":34,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"}]},{\"factors_insights_key\":\"www.acme.com/blog/all-about-marketing-analytics-for-saas/\",\"factors_insights_text\":\"of which users who perform <a>www.acme.com/blog/all-about-marketing-analytics-for-saas/</a> show 1x goal completion\",\"factors_insights_multiplier\":1,\"factors_insights_percentage\":47.3,\"factors_insights_users_count\":236,\"factors_goal_users_count\":112,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"journey\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_key\":\"www.acme.com/case-study/saas/\",\"factors_insights_text\":\"of which users who perform <a>www.acme.com/case-study/saas/</a> show 1.3x goal completion\",\"factors_insights_multiplier\":1.3,\"factors_insights_percentage\":61.5,\"factors_insights_users_count\":289,\"factors_goal_users_count\":178,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"journey\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_key\":\"www.acme.com/case-study/ecommerce/\",\"factors_insights_text\":\"of which users who perform <a>www.acme.com/case-study/ecommerce/</a> show 0.4x goal completion\",\"factors_insights_multiplier\":0.4,\"factors_insights_percentage\":18.9,\"factors_insights_users_count\":100,\"factors_goal_users_count\":19,\"factors_multiplier_increase_flag\":false,\"factors_insights_type\":\"journey\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_key\":\"www.acme.com/podcast/why-most-attribution-analysis-doesnt-work/\",\"factors_insights_text\":\"of which users who perform <a>www.acme.com/podcast/why-most-attribution-analysis-doesnt-work/</a> show 0.5x goal completion\",\"factors_insights_multiplier\":0.5,\"factors_insights_percentage\":23.7,\"factors_insights_users_count\":244,\"factors_goal_users_count\":58,\"factors_multiplier_increase_flag\":false,\"factors_insights_type\":\"journey\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_key\":\"www.acme.com/podcast/choosing-the-right-markops-tool/\",\"factors_insights_text\":\"of which users who perform <a>www.acme.com/podcast/choosing-the-right-markops-tool/</a> show 1.4x goal completion\",\"factors_insights_multiplier\":1.4,\"factors_insights_percentage\":66.3,\"factors_insights_users_count\":221,\"factors_goal_users_count\":147,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"journey\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\",\"factors_sub_insights\":[{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Tech_stack\",\"factors_attribute_value\":\"Hubspot\"}],\"factors_insights_text\":\"where <a>Tech_stack=Hubspot</a> show 1.1x goal completion\",\"factors_insights_multiplier\":1.1,\"factors_insights_percentage\":52.1,\"factors_insights_users_count\":147,\"factors_goal_users_count\":77,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Product_plan\",\"factors_attribute_value\":\"Startup\"}],\"factors_insights_text\":\"where <a>Product_plan=Startup</a> show 0.9x goal completion\",\"factors_insights_multiplier\":0.9,\"factors_insights_percentage\":42.6,\"factors_insights_users_count\":166,\"factors_goal_users_count\":71,\"factors_multiplier_increase_flag\":false,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Company_Industry\",\"factors_attribute_value\":\"Software\"}],\"factors_insights_text\":\"where <a>Company_Industry=Software</a> show 1x goal completion\",\"factors_insights_multiplier\":1,\"factors_insights_percentage\":47.3,\"factors_insights_users_count\":141,\"factors_goal_users_count\":67,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"},{\"factors_insights_attribute\":[{\"factors_attribute_key\":\"Contact_Department\",\"factors_attribute_value\":\"Marketing Operations\"}],\"factors_insights_text\":\"where <a>Contact_Department=Marketing Operations</a> show 1.5x goal completion\",\"factors_insights_multiplier\":1.5,\"factors_insights_percentage\":71,\"factors_insights_users_count\":61,\"factors_goal_users_count\":44,\"factors_multiplier_increase_flag\":true,\"factors_insights_type\":\"attribute\",\"factors_higher_completion_text\":\"\",\"factors_lower_completion_text\":\"\"}]}]}"
	return factorsOutput
}
