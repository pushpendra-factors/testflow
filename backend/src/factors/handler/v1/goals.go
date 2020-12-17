package v1

import (
	"factors/handler/helpers"
	mid "factors/middleware"
	M "factors/model"
	U "factors/util"
	"net/http"

	"github.com/gin-gonic/gin"
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
	goals, errCode := M.GetAllFactorsGoals(projectID)
	if errCode != http.StatusFound {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, goals)
}

type CreateFactorsGoalParams struct {
	Name string            `json:"name"`
	Rule M.FactorsGoalRule `json:"rule"`
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
	if !helpers.IsAdmin(projectID, loggedInAgentUUID) {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}
	params, err := GetcreateFactorsGoalParams(c)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	id, errCode, errMsg := M.CreateFactorsGoal(projectID, params.Name, params.Rule, loggedInAgentUUID)
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

type UpdateFactorsGoalParams struct {
	ID   int64             `json:"id" binding:"required"`
	Name string            `json:"name" binding:"required"`
	Rule M.FactorsGoalRule `json:"rule" binding:"required"`
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
	loggedInAgentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	projectID := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	if !helpers.IsAdmin(projectID, loggedInAgentUUID) {
		c.AbortWithStatus(http.StatusForbidden)
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
	id, errCode := M.UpdateFactorsGoal(params.ID, params.Name, params.Rule, projectID)
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
	loggedInAgentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	projectID := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	logCtx := log.WithFields(log.Fields{
		"projectId": projectID,
	})
	if !helpers.IsAdmin(projectID, loggedInAgentUUID) {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}
	params, err := getRemoveFactorsGoalParams(c)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	id, errCode := M.DeactivateFactorsGoal(params.ID, projectID)
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
	goals, errCode := M.GetAllFactorsGoalsWithNamePattern(projectID, params.SearchText)
	if errCode != http.StatusFound {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, goals)
}
