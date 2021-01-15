package handler

import (
	"encoding/json"
	C "factors/config"
	mid "factors/middleware"
	M "factors/model"
	U "factors/util"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type API_FilterRequestPayload struct {
	EventName  string `json:"name"`
	FilterExpr string `json:"expr"`
}

type API_FilterResponePayload struct {
	EventNameID uint64 `json:"id,omitempty"`
	ProjectID   uint64 `json:"project_id,omitempty"`
	EventName   string `json:"name,omitempty"`
	Deleted     bool   `json:"deleted,omitempty"`
	FilterExpr  string `json:"expr,omitempty"`
}

// APISmartEventFilterResponePayload implements the response payload for smart event filter
type APISmartEventFilterResponePayload struct {
	EventNameID uint64                `json:"id,omitempty"`
	ProjectID   uint64                `json:"project_id,omitempty"`
	EventName   string                `json:"name,omitempty"`
	Deleted     bool                  `json:"deleted,omitempty" swaggerignore:"true"`
	FilterExpr  M.SmartCRMEventFilter `json:"expr,omitempty"`
}

// Test command: curl -H "Content-Type: application/json" -i -X POST http://localhost:8080/projects/1/filters -d '{ "name": "login", "expr": "a.com/u1/u2"}'
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

	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Creating event_name failed. Invalid project."})
		return
	}

	eventName, errCode := M.CreateOrGetFilterEventName(
		&M.EventName{ProjectId: projectId, Name: requestPayload.EventName, FilterExpr: requestPayload.FilterExpr})
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
	EventName  string                `json:"name"`
	FilterExpr M.SmartCRMEventFilter `json:"expr"`
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

	projectID := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
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

	eventName, errCode := M.CreateOrGetCRMSmartEventFilterEventName(projectID,
		&M.EventName{ProjectId: projectID, Name: requestPayload.EventName}, &requestPayload.FilterExpr)
	if errCode != http.StatusCreated && errCode != http.StatusAccepted {
		if errCode == http.StatusBadRequest {
			c.AbortWithStatusJSON(errCode, gin.H{"error": "Invalid filter expression"})
			return
		}

		c.AbortWithStatusJSON(errCode, gin.H{"error": "Creating event_name failed"})
		return
	}

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
	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Get filters failed. Invalid project."})
		return
	}

	eventNames, errCode := M.GetFilterEventNames(projectId)
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
// @Router /{project_id}/v1/smart_event [get]
func GetSmartEventFiltersHandler(c *gin.Context) {

	projectID := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Get smart event filters failed. Invalid project."})
		return
	}

	eventNames, errCode := M.GetSmartEventFilterEventNames(projectID)
	if errCode != http.StatusFound && errCode != http.StatusNotFound {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Get smart event filters failed. Invalid project."})
		return
	}

	responsePayload := make([]*APISmartEventFilterResponePayload, len(eventNames))
	for i := 0; i < len(eventNames); i++ {
		decFilterExp, err := M.GetDecodedSmartEventFilterExp(eventNames[i].FilterExpr)
		if err != nil {
			log.WithFields(log.Fields{"project_id": projectID}).WithError(err).Error("Failed to decode smart event filter expression on GetSmartEventFiltersHandler.")
			continue
		}

		responsePayload[i] = &APISmartEventFilterResponePayload{
			ProjectID:   eventNames[i].ProjectId,
			EventNameID: eventNames[i].ID,
			EventName:   eventNames[i].Name,
			FilterExpr:  *decFilterExp,
		}
	}
	c.JSON(http.StatusOK, responsePayload)
}

// Test command: curl -H "Content-Type: application/json" -i -X PUT http://localhost:8080/projects/1/v1/smart_event/type=crm&filter_id=537 -d '{ "name": "updated_name" }'
// UpdateSmartEventFilterHandler godoc
// @Summary To update an existing smart event filter.
// @Tags V1ApiSmartEvent
// @Accept  json
// @Produce json
// @Param filter_id query integer true "Filter ID"
// @Param filter body handler.APISmartEventFilterRequestPayload true "Update filter"
// @Success 202 {object} handler.APISmartEventFilterResponePayload
// @Param type query string true "Smart event type"  Enums(crm)
// @Router /{project_id}/v1/smart_event [put]
func UpdateSmartEventFilterHandler(c *gin.Context) {
	r := c.Request

	projectID := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
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

	filterID, err := strconv.ParseUint(c.Query("filter_id"), 10, 64)
	if err != nil || filterID == 0 {
		logCtx.WithFields(log.Fields{log.ErrorKey: err}).Error("Updating smart event filter failed. filter_id parse failed.")
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

	eventName, status := M.UpdateCRMSmartEventFilter(projectID, filterID, &M.EventName{Name: requestPayload.EventName}, &requestPayload.FilterExpr)
	if status != http.StatusAccepted {
		c.JSON(status, gin.H{"error": "Failed to update smart event name"})
		return
	}

	responsePayload := &APISmartEventFilterResponePayload{
		ProjectID:   eventName.ProjectId,
		EventNameID: eventName.ID,
		EventName:   eventName.Name,
		FilterExpr:  requestPayload.FilterExpr,
	}

	c.JSON(status, responsePayload)

}

// Test command: curl -H "Content-Type: application/json" -i -X PUT http://localhost:8080/projects/1/filters/364 -d '{ "name": "updated_name" }'
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

	filterId, err := strconv.ParseUint(c.Params.ByName("filter_id"), 10, 64)
	if err != nil || filterId == 0 {
		logCtx.WithFields(log.Fields{log.ErrorKey: err}).Error("Updating filter failed. filter_id parse failed.")
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

	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Updating filter failed. Invalid project."})
		return
	}

	// Update name if there is any change.
	eventName, errCode := M.UpdateFilterEventName(projectId, filterId, &M.EventName{Name: requestPayload.EventName})
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

	filterId, err := strconv.ParseUint(c.Params.ByName("filter_id"), 10, 64)
	if err != nil || filterId == 0 {
		logCtx.WithError(err).Error("Updating filter failed. filter_id parse failed.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid filter id."})
		return
	}

	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Get filters failed. Invalid project."})
		return
	}

	errCode := M.DeleteFilterEventName(projectId, filterId)
	if errCode != http.StatusAccepted {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Updating filter failed."})
		return
	}

	c.JSON(http.StatusAccepted, &API_FilterResponePayload{
		ProjectID:   projectId,
		EventNameID: filterId,
	})
}
