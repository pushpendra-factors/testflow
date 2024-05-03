package handler

import (
	"encoding/json"
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func CreateSegmentHandler(c *gin.Context) {
	r := c.Request

	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)

	logCtx := log.WithFields(log.Fields{
		"projectId": projectID,
		"reqId":     U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
	})

	if projectID == 0 {
		logCtx.Error("Creation of Segment failed. Invalid project.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Segment Creation failed. Invalid project."})
		return
	}

	if r.Body == nil {
		logCtx.Error("Invalid request. Request body unavailable.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Segment Creation failed. Missing request body."})
		return
	}

	decoder := json.NewDecoder(r.Body)
	var request model.SegmentPayload
	if err := decoder.Decode(&request); err != nil {
		logCtx.WithError(err).Error("Segment creation failed. JSON Decoding failed.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Segment creation failed. Invalid payload."})
		return
	}

	segment, status, err := store.GetStore().CreateSegment(projectID, &request)
	if err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err.Error()})
		return
	}
	response := gin.H{
		"message": "Segment Creation Seccessful",
		"segment": segment,
	}
	c.JSON(status, response)
}

func GetSegmentsHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	logCtx := log.WithFields(log.Fields{
		"projectId": projectId,
	})
	if projectId == 0 {
		logCtx.Error("Invalid project_id.")
		return "", http.StatusBadRequest, "", "invalid project_id", true
	}

	segments, status := store.GetStore().GetAllSegments(projectId)
	if status != http.StatusFound {
		return "", http.StatusBadRequest, "", "Failed to get segments", true
	}

	return segments, http.StatusFound, "", "", false
}

func GetSegmentByIdHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	logCtx := log.WithFields(log.Fields{
		"projectId": projectId,
	})
	if projectId == 0 {
		logCtx.Error("Invalid project_id.")
		return "", http.StatusBadRequest, "", "invalid project_id", true
	}

	id := c.Params.ByName("id")
	logCtx = log.WithFields(log.Fields{
		"projectId": projectId,
		"segmentId": id,
	})
	if id == "" {
		logCtx.Error("Invalid segment Id.")
		return nil, http.StatusBadRequest, "Input Params are incorrect", "invalid segment Id", true
	}

	segment, errCode := store.GetStore().GetSegmentById(projectId, id)
	if errCode != http.StatusFound {
		logCtx.Error("Segment not found.")
		return nil, errCode, "Processing Failed", "Failed to get segment", true
	}

	return segment, http.StatusFound, "", "", false
}

func UpdateSegmentHandler(c *gin.Context) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	logCtx := log.WithFields(log.Fields{
		"projectId": projectID,
	})
	id := c.Params.ByName("id")
	logCtx = log.WithFields(log.Fields{
		"projectId": projectID,
		"segmentId": id,
	})
	if id == "" {
		logCtx.Error("Invalid segment Id.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	params, err := getUpdateSegmentParams(c)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	status, err := store.GetStore().UpdateSegmentById(projectID, id, *params)
	if err != nil {
		logCtx.Errorln("Updating Segment failed")
		c.AbortWithStatus(status)
		return
	}
	response := make(map[string]interface{})
	response["status"] = "success"
	response["id"] = id
	c.JSON(http.StatusOK, response)
}

func DeleteSegmentByIdHandler(c *gin.Context) {
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	logCtx := log.WithFields(log.Fields{
		"projectId": projectId,
	})
	if projectId == 0 {
		logCtx.Error("Deletion of Segment failed. Invalid project_id.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	id := c.Params.ByName("id")
	logCtx = log.WithFields(log.Fields{
		"projectId": projectId,
		"segmentId": id,
	})
	if id == "" {
		logCtx.Error("Deletion of Segment failed. Invalid id.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	status, err := store.GetStore().DeleteSegmentById(projectId, id)
	if status != http.StatusOK {
		logCtx.Error(err)
		c.AbortWithStatus(status)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Segment deletion successful."})
	return
}

func getUpdateSegmentParams(c *gin.Context) (*model.SegmentPayload, error) {
	params := model.SegmentPayload{}
	err := c.BindJSON(&params)
	if err != nil {
		return nil, err
	}
	return &params, nil
}
