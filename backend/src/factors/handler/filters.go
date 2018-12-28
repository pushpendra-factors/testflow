package handler

import (
	"encoding/json"
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
	FilterExpr  string `json:"expr,omitempty"`
}

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

	projectIdIntf := U.GetScopeByKey(c, "projectId")
	if projectIdIntf == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Creating event_name failed. Invalid project."})
		return
	}
	projectId := projectIdIntf.(uint64)

	eventName, errCode := M.CreateOrGetFilterEventName(
		&M.EventName{ProjectId: projectId, Name: requestPayload.EventName, FilterExpr: requestPayload.FilterExpr})
	if errCode != http.StatusCreated && errCode != http.StatusConflict {
		// Todo(Dinesh): Validation errors to be added as part of JSON response change.
		// Suggested: {status: "creating event_name failed", errors: ["validation error1", 'validation error2']}
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Creating event_name failed."})
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
	projectIdIntf := U.GetScopeByKey(c, "projectId")
	if projectIdIntf == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Get filters failed. Invalid project."})
		return
	}
	projectId := projectIdIntf.(uint64)

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

	projectIdIntf := U.GetScopeByKey(c, "projectId")
	if projectIdIntf == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Updating filter failed. Invalid project."})
		return
	}
	projectId := projectIdIntf.(uint64)

	// Update name if there is any change.
	eventName, errCode := M.UpdateFilterEventName(projectId, filterId, &M.EventName{Name: requestPayload.EventName})
	if errCode != http.StatusAccepted {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Updating filter failed."})
		return
	}

	responsePayload := &API_FilterResponePayload{EventName: eventName.Name}

	c.JSON(http.StatusAccepted, responsePayload)
}
