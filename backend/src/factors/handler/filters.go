package handler

import (
	"encoding/json"
	C "factors/config"
	V1 "factors/handler/v1"
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type API_FilterRequestPayload struct {
	EventName  string `json:"name"`
	FilterExpr string `json:"expr"`
}

type API_FilterResponePayload struct {
	EventNameID string `json:"id,omitempty"`
	ProjectID   int64  `json:"project_id,omitempty"`
	EventName   string `json:"name,omitempty"`
	Deleted     bool   `json:"deleted,omitempty"`
	FilterExpr  string `json:"expr,omitempty"`
}

// APISmartEventFilterResponePayload implements the response payload for smart event filter
type APISmartEventFilterResponePayload struct {
	EventNameID string                    `json:"id,omitempty"`
	ProjectID   int64                     `json:"project_id,omitempty"`
	EventName   string                    `json:"name,omitempty"`
	Deleted     bool                      `json:"deleted,omitempty" swaggerignore:"true"`
	FilterExpr  model.SmartCRMEventFilter `json:"expr,omitempty"`
}

// Test command: curl -H "Content-UnitType: application/json" -i -X POST http://localhost:8080/projects/1/filters -d '{ "name": "login", "expr": "a.com/u1/u2"}'
// CreateFilterHandler godoc
// @Summary To create a new filter.
// @Tags Filters
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param filter body handler.API_FilterRequestPayload true "Create filter"
// @Success 201 {object} handler.API_FilterResponePayload
// @Router /{project_id}/filters [post]
func CreateFilterHandler(c *gin.Context) {
	r := c.Request

	var requestPayload API_FilterRequestPayload

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&requestPayload); err != nil {
		log.WithFields(log.Fields{log.ErrorKey: err}).Error("Creating event_name failed. JSON Decoding failed.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Creating event_name failed. Invalid payload."})
		return
	}

	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Creating event_name failed. Invalid project."})
		return
	}

	eventName, errCode := store.GetStore().CreateOrGetFilterEventName(
		&model.EventName{ProjectId: projectId, Name: requestPayload.EventName, FilterExpr: requestPayload.FilterExpr})
	if errCode != http.StatusCreated {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Creating event_name failed"})
		return
	}

	responsePayload := &API_FilterResponePayload{
		ProjectID:   eventName.ProjectId,
		EventNameID: eventName.ID,
		EventName:   eventName.Name,
		FilterExpr:  eventName.FilterExpr,
	}

	c.JSON(http.StatusCreated, responsePayload)
}

// APISmartEventFilterRequestPayload implements the request payload for smart event filters
type APISmartEventFilterRequestPayload struct {
	EventName  string                    `json:"name"`
	FilterExpr model.SmartCRMEventFilter `json:"expr"`
}

// CreateSmartEventFilterHandler godoc
// @Summary To create a new smart event filter.
// @Tags V1ApiSmartEvent
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param type query string true "Smart event type" Enums(crm)
// @Param filter body handler.APISmartEventFilterRequestPayload true "Create smart event filter"
// @Success 201 {object} handler.APISmartEventFilterResponePayload
// @Router /{project_id}/v1/smart_event [post]
func CreateSmartEventFilterHandler(c *gin.Context) {
	r := c.Request

	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Creating event_name failed. Invalid project."})
		return
	}

	eventType := c.Query("type")
	if eventType != "crm" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid parameter"})
		return
	}

	if !C.IsAllowedSmartEventRuleCreation() {
		c.JSON(http.StatusOK, gin.H{})
		return
	}

	var requestPayload APISmartEventFilterRequestPayload
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&requestPayload); err != nil {
		log.WithFields(log.Fields{log.ErrorKey: err}).Error("Creating event_name failed. JSON Decoding failed.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Creating event_name failed. Invalid payload."})
		return
	}

	model.HandleSmartEventNoneTypeValue(&requestPayload.FilterExpr)
	eventName, errCode := store.GetStore().CreateOrGetCRMSmartEventFilterEventName(projectID,
		&model.EventName{ProjectId: projectID, Name: requestPayload.EventName}, &requestPayload.FilterExpr)
	if errCode != http.StatusCreated && errCode != http.StatusAccepted {
		if errCode == http.StatusBadRequest {
			c.AbortWithStatusJSON(errCode, gin.H{"error": "Invalid filter expression"})
			return
		}

		if errCode == http.StatusConflict {
			c.AbortWithStatusJSON(errCode, gin.H{"error": "Duplicate rule or event_name."})
			return
		}

		c.AbortWithStatusJSON(errCode, gin.H{"error": "Creating event_name failed"})
		return
	}

	model.HandleSmartEventAnyTypeValue(&requestPayload.FilterExpr)
	responsePayload := &APISmartEventFilterResponePayload{
		ProjectID:   eventName.ProjectId,
		EventNameID: eventName.ID,
		EventName:   eventName.Name,
		FilterExpr:  requestPayload.FilterExpr,
	}

	c.JSON(http.StatusCreated, responsePayload)
}

// GetFiltersHandler godoc
// @Summary Get the list of existing filters.
// @Tags Filters
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Success 200 {array} handler.API_FilterResponePayload
// @Router /{project_id}/filters [get]
func GetFiltersHandler(c *gin.Context) {
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Get filters failed. Invalid project."})
		return
	}

	eventNames, errCode := store.GetStore().GetFilterEventNames(projectId)
	if errCode != http.StatusFound && errCode != http.StatusNotFound {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Get filters failed. Invalid project."})
		return
	}

	responsePayload := make([]*API_FilterResponePayload, len(eventNames))
	for i := 0; i < len(eventNames); i++ {
		responsePayload[i] = &API_FilterResponePayload{
			ProjectID:   eventNames[i].ProjectId,
			EventNameID: eventNames[i].ID,
			EventName:   eventNames[i].Name,
			FilterExpr:  eventNames[i].FilterExpr,
		}
	}

	c.JSON(http.StatusOK, responsePayload)
}

// GetSmartEventFiltersHandler godoc
// @Summary Get the list of existing smart event filters.
// @Tags V1ApiSmartEvent
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Success 200 {array} handler.APISmartEventFilterResponePayload
// @Param project_id path integer true "Project ID"
// @Param type query string true "Smart event type"  Enums(crm)
// @Router /{project_id}/v1/smart_event [get]
func GetSmartEventFiltersHandler(c *gin.Context) {

	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Get smart event filters failed. Invalid project."})
		return
	}

	eventNames, errCode := store.GetStore().GetSmartEventFilterEventNames(projectID, false)
	if errCode != http.StatusFound && errCode != http.StatusNotFound {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Get smart event filters failed. Invalid project."})
		return
	}

	var responsePayload []APISmartEventFilterResponePayload
	for i := 0; i < len(eventNames); i++ {
		APISmartEventFilter := APISmartEventFilterResponePayload{
			ProjectID:   eventNames[i].ProjectId,
			EventNameID: eventNames[i].ID,
			EventName:   eventNames[i].Name,
		}

		decFilterExp, err := model.GetDecodedSmartEventFilterExp(eventNames[i].FilterExpr)
		if err == nil {
			model.HandleSmartEventAnyTypeValue(decFilterExp)
			APISmartEventFilter.FilterExpr = *decFilterExp
		} else {
			log.WithFields(log.Fields{"project_id": projectID}).WithError(err).Error("Failed to decode smart event filter expression on GetSmartEventFiltersHandler.")
		}

		responsePayload = append(responsePayload, APISmartEventFilter)
	}

	c.JSON(http.StatusOK, responsePayload)
}

// Test command: curl -H "Content-UnitType: application/json" -i -X PUT http://localhost:8080/projects/1/v1/smart_event/type=crm&filter_id=537 -d '{ "name": "updated_name" }'
// UpdateSmartEventFilterHandler godoc
// @Summary To update an existing smart event filter.
// @Tags V1ApiSmartEvent
// @Accept  json
// @Produce json
// @Param filter_id query integer true "Filter ID"
// @Param project_id path integer true "Project ID"
// @Param filter body handler.APISmartEventFilterRequestPayload true "Update filter"
// @Success 202 {object} handler.APISmartEventFilterResponePayload
// @Param type query string true "Smart event type"  Enums(crm)
// @Router /{project_id}/v1/smart_event [put]
func UpdateSmartEventFilterHandler(c *gin.Context) {
	r := c.Request

	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Update event_name failed. Invalid project."})
		return
	}

	eventType := c.Query("type")
	if eventType != "crm" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid parameter"})
		return
	}

	logCtx := log.WithFields(log.Fields{
		"project_id": projectID,
		"reqId":      U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
	})

	filterID := strings.TrimSpace(c.Query("filter_id"))
	if filterID == "" {
		logCtx.Error("Updating smart event filter failed. filter_id parse failed.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid filter id."})
		return
	}

	var requestPayload APISmartEventFilterRequestPayload

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&requestPayload); err != nil {
		logCtx.WithFields(log.Fields{log.ErrorKey: err}).Error("Update event_name failed. JSON Decoding failed.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Update event_name failed. Invalid payload."})
		return
	}

	model.HandleSmartEventNoneTypeValue(&requestPayload.FilterExpr)
	eventName, status := store.GetStore().UpdateCRMSmartEventFilter(projectID, filterID, &model.EventName{Name: requestPayload.EventName}, &requestPayload.FilterExpr)
	if status != http.StatusAccepted {
		if status == http.StatusConflict {
			c.JSON(status, gin.H{"error": "Duplicate rule."})
		}

		c.JSON(status, gin.H{"error": "Failed to update smart event name"})
		return
	}

	model.HandleSmartEventAnyTypeValue(&requestPayload.FilterExpr)
	responsePayload := &APISmartEventFilterResponePayload{
		ProjectID:   eventName.ProjectId,
		EventNameID: eventName.ID,
		EventName:   eventName.Name,
		FilterExpr:  requestPayload.FilterExpr,
	}

	c.JSON(status, responsePayload)

}

// Test command: curl -H "Content-UnitType: application/json" -i -X DELETE http://localhost:8080/projects/1/v1/smart_event/type=crm&filter_id=537
// DeleteSmartEventFilterHandler godoc
// @Summary To delete an existing smart event filter.
// @Tags V1ApiSmartEvent
// @Accept  json
// @Produce json
// @Param filter_id query integer true "Filter ID"
// @Param project_id path integer true "Project ID"
// @Success 202 {object} handler.APISmartEventFilterResponePayload
// @Param type query string true "Smart event type"  Enums(crm)
// @Router /{project_id}/v1/smart_event [delete]
func DeleteSmartEventFilterHandler(c *gin.Context) (interface{}, int, string, string, bool) {

	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		return nil, http.StatusUnauthorized, V1.INVALID_PROJECT, "Delete smart event_name filter failed. Invalid project.", true
	}

	eventType := c.Query("type")
	if eventType != "crm" {
		return nil, http.StatusBadRequest, V1.INVALID_INPUT, "Delete smart event_name filter failed. Invalid query type", true
	}

	filterID := strings.TrimSpace(c.Query("filter_id"))
	if filterID == "" {
		return nil, http.StatusBadRequest, V1.INVALID_INPUT, "Delete smart event_name filter failed. Invalid rule id", true
	}

	eventName, status := store.GetStore().DeleteSmartEventFilter(projectID, filterID)
	if status != http.StatusAccepted {
		return nil, http.StatusInternalServerError, V1.PROCESSING_FAILED, "Delete smart event_name filter failed.", true
	}

	filterExp, err := model.GetDecodedSmartEventFilterExp(eventName.FilterExpr)
	if err != nil {
		return nil, http.StatusInternalServerError, V1.PROCESSING_FAILED, "Delete smart event_name filter failed. Failed to decode.", true
	}

	responsePayload := &APISmartEventFilterResponePayload{
		ProjectID:   eventName.ProjectId,
		EventNameID: eventName.ID,
		EventName:   eventName.Name,
		FilterExpr:  *filterExp,
	}

	return responsePayload, http.StatusAccepted, "", "", false
}

// Test command: curl -H "Content-UnitType: application/json" -i -X PUT http://localhost:8080/projects/1/filters/364 -d '{ "name": "updated_name" }'
// UpdateFilterHandler godoc
// @Summary To update an existing filter.
// @Tags Filters
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param filter_id path integer true "Filter ID"
// @Param filter body handler.API_FilterRequestPayload true "Update filter"
// @Success 202 {object} handler.API_FilterResponePayload
// @Router /{project_id}/filters/{filter_id} [put]
func UpdateFilterHandler(c *gin.Context) {
	r := c.Request

	logCtx := log.WithFields(log.Fields{
		"reqId": U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
	})

	filterId := strings.TrimSpace(c.Params.ByName("filter_id"))
	if filterId == "" {
		logCtx.Error("Updating filter failed. filter_id parse failed.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid filter id."})
		return
	}

	var requestPayload API_FilterRequestPayload

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&requestPayload); err != nil {
		logCtx.WithError(err).Error("Updating filter failed. JSON Decoding failed.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Updating filter failed. Invalid payload."})
		return
	}

	if requestPayload.FilterExpr != "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Updating filter_expr is not allowed."})
		return
	}

	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Updating filter failed. Invalid project."})
		return
	}

	// Update name if there is any change.
	eventName, errCode := store.GetStore().UpdateFilterEventName(projectId, filterId,
		&model.EventName{Name: requestPayload.EventName})
	if errCode != http.StatusAccepted {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Updating filter failed."})
		return
	}

	responsePayload := &API_FilterResponePayload{EventName: eventName.Name}

	c.JSON(http.StatusAccepted, responsePayload)
}

// DeleteFilterHandler godoc
// @Summary To delete an existing filter.
// @Tags Filters
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param filter_id path integer true "Filter ID"
// @Success 202 {object} handler.API_FilterResponePayload
// @Router /{project_id}/filters/{filter_id} [delete]
func DeleteFilterHandler(c *gin.Context) {

	logCtx := log.WithFields(log.Fields{
		"reqId": U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
	})

	filterId := strings.TrimSpace(c.Params.ByName("filter_id"))
	if filterId == "" {
		logCtx.Error("Updating filter failed. filter_id parse failed.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid filter id."})
		return
	}

	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Get filters failed. Invalid project."})
		return
	}

	errCode := store.GetStore().DeleteFilterEventName(projectId, filterId)
	if errCode != http.StatusAccepted {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Updating filter failed."})
		return
	}

	c.JSON(http.StatusAccepted, &API_FilterResponePayload{
		ProjectID:   projectId,
		EventNameID: filterId,
	})
}
