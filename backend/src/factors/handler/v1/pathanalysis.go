package v1

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

type PathAnalysis model.PathAnalysis

const (
	buildLimit = 10
	BUILDING   = "building"
	SAVED      = "saved"
)

func GetPathAnalysisEntityHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		return nil, http.StatusForbidden, "", "Get path analysis enitity failed. Invalid project.", true
	}
	entity, errCode := store.GetStore().GetAllPathAnalysisEntityByProject(projectID)
	if errCode != http.StatusFound {
		return nil, errCode, "", "Get Saved Queries failed.", true
	}

	return entity, http.StatusOK, "", "", false
}

func CreatePathAnalysisEntityHandler(c *gin.Context) (interface{}, int, string, string, bool) {

	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	userID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)

	if projectID == 0 {
		return nil, http.StatusUnauthorized, INVALID_PROJECT, ErrorMessages[INVALID_PROJECT], true
	}
	log.Info("Create function handler triggered.")

	var entity model.PathAnalysisQuery
	r := c.Request
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&entity); err != nil {
		errMsg := "Get pathanalysis failed. Invalid JSON."
		log.WithFields(log.Fields{"project_id": projectID}).WithError(err).Error(errMsg)
		return nil, http.StatusBadRequest, errMsg, "", true
	}

	if len(entity.IncludeEvents) != 0 && len(entity.ExcludeEvents) != 0 {
		return nil, http.StatusBadRequest, PROCESSING_FAILED, "Provide either IncludeEvents or ExcludeEvents", true
	}

	err := BeforeCreate(projectID)
	if err != http.StatusOK {
		return nil, http.StatusBadRequest, PROCESSING_FAILED, "Build limit reached for creating pathanalysis", true
	}

	_, errCode, errMsg := store.GetStore().CreatePathAnalysisEntity(userID, projectID, &entity)
	if errCode != http.StatusCreated {
		log.WithFields(log.Fields{"document": entity, "err-message": errMsg}).Error("Failed to create path analysis in handler.")
		return nil, errCode, PROCESSING_FAILED, errMsg, true
	}

	c.JSON(errCode, gin.H{"Status": "successful"})
	return entity, http.StatusCreated, "", "", false
}

// Function triggered before Create handler
func BeforeCreate(projectID int64) int {

	// Checks if the there are already enough projects with BUILDING status
	status := []string{BUILDING, SAVED}
	count, _, _ := store.GetStore().GetProjectCountWithStatus(projectID, status)
	if count >= buildLimit {
		log.WithFields(log.Fields{"project_id": projectID, "err-message": count}).Error("Project BUILDING Limit reached")
		return http.StatusInternalServerError
	}
	return http.StatusOK
}

func DeleteSavedPathAnalysisEntityHandler(c *gin.Context) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Delete pathanalaysis failed. Invalid project."})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Delete failed. Invalid id provided."})
		return
	}

	errCode, errMsg := store.GetStore().DeletePathAnalysisEntity(projectID, id)
	if errCode != http.StatusAccepted {
		c.AbortWithStatusJSON(errCode, errMsg)
		return
	}

	c.JSON(errCode, gin.H{"Status": "OK"})
}
