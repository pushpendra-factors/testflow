package v1

import (
	mid "factors/middleware"
	M "factors/model"
	P "factors/pattern"
	PW "factors/pattern_service_wrapper"
	U "factors/util"
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
	var err error
	if modelIdParam != "" {
		modelId, err = strconv.ParseUint(modelIdParam, 10, 64)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
	}
	inputType := c.Query("type")
	params, err := GetcreateFactorsGoalParams(c)
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
	startConstraints, endConstraints := parseConstraints(params.Rule.Rule)
	if results, err := PW.FactorV1("",
		projectId, params.Rule.StartEvent, startConstraints,
		params.Rule.EndEvent, endConstraints, P.COUNT_TYPE_PER_USER, ps); err != nil {
		logCtx.WithError(err).Error("Factors failed.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	} else {
		results.Type = inputType
		results.GoalRule = params.Rule
		c.JSON(http.StatusOK, results)
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
