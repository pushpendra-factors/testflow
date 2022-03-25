package v1

import (
	"encoding/json"
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

// GetAllFactorsGoalsHandler - Get All goals handler
// GetAllFactorsGoalsHandler godoc
// @Summary Get All the saved goals for a given project
// @Tags V1FactorsApi
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Success 200 {array} model.FactorsGoal
// @Router /{project_id}/v1/factors/goals [GET]
func GetAllFactorsGoalsHandler(c *gin.Context) {
	projectID := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	goals, errCode := store.GetStore().GetAllFactorsGoals(projectID)
	if errCode != http.StatusFound {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	for index, goal := range goals {
		var ipRule model.FactorsGoalRule
		json.Unmarshal((goal.Rule).RawMessage, &ipRule)
		opRule := ReverseMapRule(ipRule)
		ruleJSON, _ := json.Marshal(opRule)
		ruleJsonb := postgres.Jsonb{ruleJSON}
		goals[index].Rule = ruleJsonb
	}
	c.JSON(http.StatusOK, goals)
}

type CreateGoalInputParams struct {
	StartEvent              model.QueryEventWithProperties `json:"st_en"`
	EndEvent                model.QueryEventWithProperties `json:"en_en"`
	GlobalFilters           []model.QueryProperty          `json:"gpr"`
	IncludedEvents          []string                       `json:"in_en"`
	IncludedUserProperties  []string                       `json:"in_upr"`
	IncludedEventProperties []string                       `json:"in_epr"`
}

type CreateFactorsGoalParams struct {
	Name string                `json:"name"`
	Rule CreateGoalInputParams `json:"rule"`
}

func GetcreateFactorsGoalParams(c *gin.Context) (*CreateFactorsGoalParams, error) {
	params := CreateFactorsGoalParams{}
	err := c.BindJSON(&params)
	if err != nil {
		return nil, err
	}
	return &params, nil
}

// CreateFactorsGoalsHandler - Create FactorsGoal Handler
// CreateFactorsGoalsHandler godoc
// @Summary Create a saved goal with specified name and rule
// @Tags V1FactorsApi
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param create body v1.CreateFactorsGoalParams true "Create"
// @Success 201 {string} json "{"id": uint64, "status": string}"
// @Router /{project_id}/v1/factors/goals [POST]
func CreateFactorsGoalsHandler(c *gin.Context) {
	loggedInAgentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	projectID := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	params, err := GetcreateFactorsGoalParams(c)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	rule, _, _, _ := MapRule(params.Rule)
	id, errCode, errMsg := store.GetStore().CreateFactorsGoal(projectID, params.Name, rule, loggedInAgentUUID)
	if errCode != http.StatusCreated {
		if errMsg != "" {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": errMsg})
			return
		}
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	response := make(map[string]interface{})
	response["status"] = "success"
	response["id"] = id
	c.JSON(http.StatusCreated, response)
}

func MapRule(ip CreateGoalInputParams) (model.FactorsGoalRule, map[string]bool, map[string]bool, map[string]bool) {
	op := model.FactorsGoalRule{}
	op.StartEvent = ip.StartEvent.Name
	op.EndEvent = ip.EndEvent.Name
	op.Rule = model.FactorsGoalFilter{}
	if len(ip.StartEvent.Properties) > 0 || len(ip.EndEvent.Properties) > 0 || len(ip.GlobalFilters) > 0 {
		op.Rule.StartEnEventFitler = make([]model.KeyValueTuple, 0)
		op.Rule.EndEnEventFitler = make([]model.KeyValueTuple, 0)
		op.Rule.StartEnUserFitler = make([]model.KeyValueTuple, 0)
		op.Rule.EndEnUserFitler = make([]model.KeyValueTuple, 0)
		op.Rule.GlobalFilters = make([]model.KeyValueTuple, 0)
	}
	for _, property := range ip.StartEvent.Properties {
		if property.Entity == "user" {
			op.Rule.StartEnUserFitler = append(op.Rule.StartEnUserFitler, mapProperty(property))
		}
		if property.Entity == "event" {
			op.Rule.StartEnEventFitler = append(op.Rule.StartEnEventFitler, mapProperty(property))
		}
	}
	for _, property := range ip.EndEvent.Properties {
		if property.Entity == "user" {
			op.Rule.EndEnUserFitler = append(op.Rule.EndEnUserFitler, mapProperty(property))
		}
		if property.Entity == "event" {
			op.Rule.EndEnEventFitler = append(op.Rule.EndEnEventFitler, mapProperty(property))
		}
	}
	for _, property := range ip.GlobalFilters {
		op.Rule.GlobalFilters = append(op.Rule.GlobalFilters, mapProperty(property))
	}
	op.Rule.IncludedEvents = ip.IncludedEvents
	op.Rule.IncludedEventProperties = ip.IncludedEventProperties
	op.Rule.IncludedUserProperties = ip.IncludedUserProperties
	includedEvents := make(map[string]bool)
	for _, event := range ip.IncludedEvents {
		includedEvents[event] = true
	}
	includedUserProperties := make(map[string]bool)
	for _, event := range ip.IncludedUserProperties {
		includedUserProperties[event] = true
	}
	includedEventProperties := make(map[string]bool)
	for _, event := range ip.IncludedEventProperties {
		includedEventProperties[event] = true
	}
	return op, includedEvents, includedEventProperties, includedUserProperties
}

func mapProperty(pr model.QueryProperty) model.KeyValueTuple {
	value := model.KeyValueTuple{}

	value.Key = pr.Property
	value.Type = pr.Type
	value.Value = pr.Value
	if pr.Type == "numerical" {
		numValue, _ := strconv.ParseFloat(pr.Value, 32)
		if pr.Operator == "equals" {
			value.Operator = true
			value.LowerBound = numValue
			value.UpperBound = numValue
		} else {
			value.Operator = false
			if pr.Operator == "greaterThan" {
				value.LowerBound = numValue
				value.UpperBound = math.MaxFloat64
			}
			if pr.Operator == "lesserThan" {
				value.LowerBound = -math.MaxFloat64
				value.UpperBound = numValue
			}
		}
	}
	if pr.Type == "categorical" {
		value.Value = pr.Value
		if pr.Operator == "equals" {
			value.Operator = true
		} else {
			value.Operator = false
		}
	}
	return value
}

type UpdateFactorsGoalParams struct {
	ID   int64                 `json:"id" binding:"required"`
	Name string                `json:"name" binding:"required"`
	Rule CreateGoalInputParams `json:"rule" binding:"required"`
}

func getUpdateFactorsGoalParams(c *gin.Context) (*UpdateFactorsGoalParams, error) {
	params := UpdateFactorsGoalParams{}
	err := c.BindJSON(&params)
	if err != nil {
		return nil, err
	}
	return &params, nil
}

// UpdateFactorsGoalsHandler - update FactorsGoal Handler
// UpdateFactorsGoalsHandler godoc
// @Summary Update a saved goal with specified name and rule
// @Tags V1FactorsApi
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param update body v1.UpdateFactorsGoalParams true "Update"
// @Success 200 {string} json "{"id": uint64, "status": string}"
// @Router /{project_id}/v1/factors/goals/update [PUT]
func UpdateFactorsGoalsHandler(c *gin.Context) {
	projectID := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	logCtx := log.WithFields(log.Fields{
		"projectId": projectID,
	})
	params, err := getUpdateFactorsGoalParams(c)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	rule, _, _, _ := MapRule(params.Rule)
	id, errCode := store.GetStore().UpdateFactorsGoal(params.ID, params.Name, rule, projectID)
	if errCode != http.StatusOK {
		logCtx.Errorln("Updating FactorsGoal failed")
		if errCode == http.StatusFound {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "FactorsGoal inactive"})
			return
		}
		if errCode == http.StatusNotFound {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "FactorsGoal Not found"})
			return
		}
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	response := make(map[string]interface{})
	response["status"] = "success"
	response["id"] = id
	c.JSON(http.StatusOK, response)
}

type RemoveFactorsGoalParams struct {
	ID int64 `json:"id" binding:"required"`
}

func getRemoveFactorsGoalParams(c *gin.Context) (*RemoveFactorsGoalParams, error) {
	params := RemoveFactorsGoalParams{}
	err := c.BindJSON(&params)
	if err != nil {
		return nil, err
	}
	return &params, nil
}

// RemoveFactorsGoalsHandler - remove FactorsGoal Handler
// RemoveFactorsGoalsHandler godoc
// @Summary Remove a saved goal
// @Tags V1FactorsApi
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param remove body v1.RemoveFactorsGoalParams true "Remove"
// @Success 200 {string} json "{"id": uint64, "status": string}"
// @Router /{project_id}/v1/factors/goals/remove [DELETE]
func RemoveFactorsGoalsHandler(c *gin.Context) {
	projectID := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	logCtx := log.WithFields(log.Fields{
		"projectId": projectID,
	})
	params, err := getRemoveFactorsGoalParams(c)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	id, errCode := store.GetStore().DeactivateFactorsGoal(params.ID, projectID)
	if errCode != http.StatusOK {
		logCtx.Errorln("Removing FactorsGoal failed")
		if errCode == http.StatusFound {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "FactorsGoal already deleted"})
			return
		}
		if errCode == http.StatusNotFound {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "FactorsGoal Not found"})
			return
		}
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	response := make(map[string]interface{})
	response["status"] = "success"
	response["id"] = id
	c.JSON(http.StatusOK, response)
}

type SearchFactorsGoalParams struct {
	SearchText string `json:"search_text" binding:"required"`
}

func getSearchFactorsGoalParams(c *gin.Context) (*SearchFactorsGoalParams, error) {
	params := SearchFactorsGoalParams{}
	err := c.BindJSON(&params)
	if err != nil {
		return nil, err
	}
	return &params, nil
}

// SearchFactorsGoalHandler - search FactorsGoal Handler
// SearchFactorsGoalHandler godoc
// @Summary Search on saved goals
// @Tags V1FactorsApi
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param search body v1.SearchFactorsGoalParams true "Search"
// @Success 200 {array} model.FactorsGoal
// @Router /{project_id}/v1/factors/goals/search [GET]
func SearchFactorsGoalHandler(c *gin.Context) {
	projectID := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	params, err := getSearchFactorsGoalParams(c)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	goals, errCode := store.GetStore().GetAllFactorsGoalsWithNamePattern(projectID, params.SearchText)
	if errCode != http.StatusFound {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	for index, goal := range goals {
		var ipRule model.FactorsGoalRule
		json.Unmarshal((goal.Rule).RawMessage, &ipRule)
		opRule := ReverseMapRule(ipRule)
		ruleJSON, _ := json.Marshal(opRule)
		ruleJsonb := postgres.Jsonb{ruleJSON}
		goals[index].Rule = ruleJsonb
	}
	c.JSON(http.StatusOK, goals)
}

func ReverseMapRule(ip model.FactorsGoalRule) CreateGoalInputParams {
	op := CreateGoalInputParams{}
	op.StartEvent = model.QueryEventWithProperties{}
	op.EndEvent = model.QueryEventWithProperties{}
	op.StartEvent.Name = ip.StartEvent
	op.EndEvent.Name = ip.EndEvent
	for _, filter := range ip.Rule.StartEnEventFitler {
		op.StartEvent.Properties = append(op.StartEvent.Properties, ReverseMapProperty(filter, "event"))
	}
	for _, filter := range ip.Rule.EndEnEventFitler {
		op.EndEvent.Properties = append(op.EndEvent.Properties, ReverseMapProperty(filter, "event"))
	}
	for _, filter := range ip.Rule.StartEnUserFitler {
		op.StartEvent.Properties = append(op.StartEvent.Properties, ReverseMapProperty(filter, "user"))
	}
	for _, filter := range ip.Rule.EndEnUserFitler {
		op.EndEvent.Properties = append(op.EndEvent.Properties, ReverseMapProperty(filter, "user"))
	}
	for _, filter := range ip.Rule.GlobalFilters {
		op.GlobalFilters = append(op.GlobalFilters, ReverseMapProperty(filter, "user"))
	}
	op.IncludedEvents = ip.Rule.IncludedEvents
	op.IncludedEventProperties = ip.Rule.IncludedEventProperties
	op.IncludedUserProperties = ip.Rule.IncludedUserProperties
	return op
}

func ReverseMapProperty(ip model.KeyValueTuple, entity string) model.QueryProperty {
	op := model.QueryProperty{}
	op.Entity = entity
	op.Type = ip.Type
	op.Property = ip.Key
	op.Value = ip.Value
	if ip.Type == "categorical" {
		op.Value = ip.Value
	}
	if ip.Type == "numerical" {
		if ip.LowerBound != -math.MaxFloat64 {
			op.Value = fmt.Sprintf("%f", ip.LowerBound)
			op.Operator = "lowerThan"
		} else {
			op.Value = fmt.Sprintf("%f", ip.UpperBound)
			op.Operator = "greaterThan"
		}
	}
	return op
}
