package handler

import (
	"encoding/json"
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

// Test command: curl -H "Content-Type: application/json" -i -X POST http://localhost:8080/projects/1/filters -d '{ "name": "login", "expr": "a.com/u1/u2"}'
func CreateFilterHandler(c *gin.Context) {
	r := c.Request

	var requestPayload API_FilterRequestPayload

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&requestPayload); err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Creating event_name failed. JSON Decoding failed.")
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

// Test command: curl -H "Content-Type: application/json" -i -X PUT http://localhost:8080/projects/1/filters/364 -d '{ "name": "updated_name" }'
func UpdateFilterHandler(c *gin.Context) {
	r := c.Request

	filterId, err := strconv.ParseUint(c.Params.ByName("filter_id"), 10, 64)
	if err != nil || filterId == 0 {
		log.WithFields(log.Fields{"error": err}).Error("Updating filter failed. filter_id parse failed.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid filter id."})
		return
	}

	var requestPayload API_FilterRequestPayload

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&requestPayload); err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Updating filter failed. JSON Decoding failed.")
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

func DeleteFilterHandler(c *gin.Context) {
	filterId, err := strconv.ParseUint(c.Params.ByName("filter_id"), 10, 64)
	if err != nil || filterId == 0 {
		log.WithFields(log.Fields{"error": err}).Error("Updating filter failed. filter_id parse failed.")
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
