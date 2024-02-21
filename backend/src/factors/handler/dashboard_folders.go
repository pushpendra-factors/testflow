package handler

import (
	"encoding/json"
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"net/http"

	"github.com/gin-gonic/gin"
)

// MoveDashboardToNewFolderHandler godoc
// @Summary To create dashboard folder and move the dashboard to this folder
// @Tags Dashboard Folders
// @Accept json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param dashboard folder body model.DashboardFoldersRequestPayload true "Create new dashboard folder and move dashboard to this folder"
// @Success @201 {object} model.DashboardFolders
// @Router /{project_id}/dashboard_folder/ [POST]
func MoveDashboardToNewFolderHandler(c *gin.Context) (interface{}, int, string, bool) {

	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	agentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	if projectId == 0 {
		return nil, http.StatusForbidden, "Create dashboard folder failed. Invalid project.", true
	}

	var reqPayload model.DashboardFoldersRequestPayload

	r := c.Request
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&reqPayload); err != nil {
		errMsg := "Create dashboard folder failed. Invalid JSON"
		return nil, http.StatusBadRequest, errMsg, true
	}

	// checking if agent has access to dashboard.
	hasAccess, _ := store.GetStore().HasAccessToDashboard(projectId, agentUUID, reqPayload.DashboardId)
	if !hasAccess {
		return nil, http.StatusForbidden, "Agent doesn't have access to move this dashboard", true
	}

	// checking if a folder with same name already exists.
	errCode := store.GetStore().IsFolderNameAlreadyExists(projectId, reqPayload.Name)
	if errCode == http.StatusInternalServerError {
		return nil, errCode, "Failed creating dashboard folder", true
	}
	if errCode != http.StatusNotFound {
		return nil, http.StatusForbidden, "Folder name not allowed", true
	}

	dashboardFolderReq := &model.DashboardFolders{
		Name: reqPayload.Name,
	}

	dashboardFolder, errCode := store.GetStore().CreateDashboardFolder(projectId, dashboardFolderReq)
	if errCode != http.StatusCreated {
		return nil, errCode, "Failed to create dashboard folder.", true
	}

	dashboardId := reqPayload.DashboardId
	if dashboardId != 0 {
		errCode := store.GetStore().UpdateDashboard(projectId, agentUUID, dashboardId, &model.UpdatableDashboard{FolderID: dashboardFolder.Id})
		if errCode != http.StatusAccepted {
			return nil, errCode, "Failed to add dashboard to folder.", true
		}
	}

	return dashboardFolder, http.StatusCreated, "", false
}

// UpdateDashboardFolderHandler godoc
// @Summary To update folder details (name)
// @Tags Dashboard Folders
// @Accept json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param folder_id path string true "Folder ID"
// @Param dashboard folder body model.UpdatableDashboardFolder true "Update Dashboard Folder"
// @Success @202 {string} "{"message": "Successfully updated."}"
// @Router /{project_id}/dashboard_folder/{folder_id} [PUT]
func UpdateDashboardFolderHandler(c *gin.Context) {
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Update dashboard folder failed. Invalid project."})
		return
	}

	folderId := c.Params.ByName("folder_id")
	if folderId == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Update dashboard folder failed. Invalid dashboard folder id."})
		return
	}

	var reqPayload model.UpdatableDashboardFolder
	r := c.Request
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&reqPayload); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Update dashboard folder failed. Invalid req payload."})
		return
	}

	errCode := store.GetStore().UpdateDashboardFolder(projectId, folderId, &reqPayload)
	if errCode != http.StatusAccepted {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Update dashboard failed."})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"message": "Successfully updated."})
}

// GetAllDashboardFolderHandler godoc
// @Summary To get all the dashboard folder of the given projectId.
// @Tags Dashboard Folder
// @Accept json
// @Produce json
// @Param project_id path integer true "Project ID"
// Success 302 {object} []model.DashboardFolders
// @Router /{project_id}/dashboard_folder [GET]
func GetAllDashboardFolderByProjectID(c *gin.Context) (interface{}, int, string, bool) {

	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		return nil, http.StatusForbidden, "Get dashboards failed. Invalid project.", true
	}

	dashboardFolders, errCode := store.GetStore().GetDashboardFolders(projectId)
	if errCode != http.StatusFound {
		return nil, errCode, "Get dashboard folders failed.", true
	}

	return dashboardFolders, http.StatusFound, "", false
}

// DeleteDashboardFolderHandler godoc
// @Summary To delete an existing dashboard folder.
// @Tags Dashboard folder
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param folder_id path string true "Folder ID"
// @Success 202 {string} json "{"message": "Successfully deleted."}"
// @Router /{project_id}/v1/dashboard_folder/{folder_id} [DELETE]
func DeleteDashboardFolderHandler(c *gin.Context) {
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Delete dashboard folder failed. Invalid project."})
		return
	}

	folderId := c.Params.ByName("folder_id")
	if folderId == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Delete dashboard folder failed. Invalid folder id."})
		return
	}

	errCode := store.GetStore().DeleteDashboardFolder(projectId, folderId)
	if errCode != http.StatusAccepted {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Failed to delete dashboard folder."})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"message": "Successfully deleted."})

}
