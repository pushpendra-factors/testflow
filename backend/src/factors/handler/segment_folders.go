package handler

import (
	"encoding/json"
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func CreateSegmentFolderRouteHandler(c *gin.Context) (interface{}, int, string,string, bool) {
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		return nil, http.StatusForbidden, "Get SegmentFolders failed. Invalid project.","", true
	}
	folderType := c.Query("type")

	if !(model.IsAccountProfiles(folderType) || model.IsUserProfiles(folderType)) {
		return nil, http.StatusBadRequest, "Type Not Found.", "", true
	}

	r := c.Request

	if r.Body == nil {
		return nil, http.StatusBadRequest, "Body Not Found.", "", true
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var request model.SegmentFolderPayload
	if err := decoder.Decode(&request); err != nil {
		return nil, http.StatusBadRequest, "Invalid Payload.", "", true
	}

 	errCode := store.GetStore().CreateSegmentFolder(projectId, request.Name, folderType)
	if errCode != http.StatusCreated {
		if errCode == http.StatusConflict {
			return nil, errCode, "Segment Name Already Exists.", "", true
		}
		return nil, errCode, "Get Segment Folders failed.", "", true
	}

	return nil, http.StatusCreated, "", "", false

}

func GetAllSegmentFoldersByProjectIDHandler(c *gin.Context) (interface{}, int, string,string, bool) {
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		return nil, http.StatusForbidden, "Get SegmentFolders failed. Invalid project.","", true
	}
	folderType := c.Query("type")

	if !(model.IsAccountProfiles(folderType) || model.IsUserProfiles(folderType)) {
		return nil, http.StatusBadRequest, "Type Not Found.", "", true
	}

	allFolders, errCode := store.GetStore().GetAllSegmentFolders(projectId, folderType)
	if errCode != http.StatusFound {
		return nil, errCode, "", "", true
	}

	return allFolders, http.StatusFound, "", "", false
}

// Update Segment By ID
func UpdateSegmentFolderByIDHandler(c *gin.Context) (interface{}, int, string,string, bool) {
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		return nil, http.StatusForbidden, "Get SegmentFolders failed. Invalid project.","", true
	}

	folderType := c.Query("type")

	if !(model.IsAccountProfiles(folderType) || model.IsUserProfiles(folderType)) {
		return nil, http.StatusBadRequest, "Type Not Found.", "", true
	}
	r := c.Request

	id,err := strconv.ParseInt(c.Params.ByName("id"), 10, 64)	
	if err != nil {
		return nil, http.StatusBadRequest, "Folder ID not found", "", true
	}
	

	if r.Body == nil {
		return nil, http.StatusBadRequest, "Body Not Found.", "", true
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var request model.SegmentFolderPayload
	if err := decoder.Decode(&request); err != nil {
		return nil, http.StatusBadRequest, "Invalid Payload.", "", true
	}

 	errCode := store.GetStore().UpdateSegmentFolderByID(projectId, id, request.Name, folderType)
	if errCode != http.StatusAccepted {
		if errCode == http.StatusConflict {
			return nil, errCode, "SegmentFolder With this name already exists.", "", true
		}else if errCode == http.StatusNotFound {
			return nil, errCode, "SegmentFolder not found.", "", true
		}else if errCode == http.StatusBadRequest {
			return nil, errCode, "SegmentFolder failed to update.", "", true
		}
		return nil, errCode, "Get Segment Folders failed.", "", true
	}

	return nil, http.StatusAccepted, "", "", false
}

func DeleteSegmentFolderByIDHandler(c *gin.Context) (interface{}, int, string,string, bool) {
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		return nil, http.StatusForbidden, "PROJECT ID not found","Get SegmentFolders failed. Invalid project.", true
	}
	folderType := c.Query("type")

	if !(model.IsAccountProfiles(folderType) || model.IsUserProfiles(folderType)) {
		return nil, http.StatusBadRequest, "Type Not Found.", "", true
	}

	id,err := strconv.ParseInt(c.Params.ByName("id"), 10, 64)	
	if err != nil {
		return nil, http.StatusBadRequest, "Folder ID not found", "Folder ID is not in path parameter or INVALID", true
	}
	
	errCode := store.GetStore().DeleteSegmentFolderByID(projectId, id, folderType)
	if errCode != http.StatusAccepted {
		if errCode == http.StatusNotFound {
			return nil, errCode, "SegmentFolder not found.", "", true
		}
		return nil, errCode, "Get Segment Folders failed.", "", true
	}

	return nil, http.StatusAccepted, "", "", false
}

func MoveSegmentFolderItemHandler(c *gin.Context) (interface{}, int, string,string, bool) {
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		return nil, http.StatusForbidden, "PROJECT ID not found","Get SegmentFolders failed. Invalid project.", true
	}
	id := c.Params.ByName("id")
	if id == "" {
		return nil, http.StatusBadRequest, "SegmentID not found", "Segment ID is not in path parameter", true
	}
	folderType := c.Query("type")

	if !(model.IsAccountProfiles(folderType) || model.IsUserProfiles(folderType)) {
		return nil, http.StatusBadRequest, "Type Not Found.", "", true
	}

	r := c.Request
	if r.Body == nil {
		return nil, http.StatusBadRequest, "Body Not Found.", "", true
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var request model.MoveSegmentFolderItemPayload
	if err := decoder.Decode(&request); err != nil {
		return nil, http.StatusBadRequest, "Invalid Payload.", "", true
	}
	errCode := store.GetStore().MoveSegmentFolderItem(projectId, id, request.FolderID, folderType)
	if errCode != http.StatusAccepted {
		if errCode == http.StatusNotFound {
			return nil, http.StatusNotFound, "Invalid SegmentID or FolderID","Segment ID or Folder ID provided are wrong", true
		}
		return nil, http.StatusInternalServerError, "Unexpected Error", "Unexpected Error Found", true
	}
	return nil, errCode, "","", false
	
}

func MoveSegmentToNewFolderHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		return nil, http.StatusForbidden, "PROJECT ID not found","Get SegmentFolders failed. Invalid project.", true
	}
	id := c.Params.ByName("id")
	if id == "" {
		return nil, http.StatusBadRequest, "SegmentID not found", "Segment ID is not in path parameter", true
	}
	folderType := c.Query("type")

	if !(model.IsAccountProfiles(folderType) || model.IsUserProfiles(folderType)) {
		return nil, http.StatusBadRequest, "Type Not Found.", "", true
	}

	r := c.Request
	if r.Body == nil {
		return nil, http.StatusBadRequest, "Body Not Found.", "", true
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var request model.SegmentFolderPayload
	if err := decoder.Decode(&request); err != nil {
		return nil, http.StatusBadRequest, "Invalid Payload.", "", true
	}


	errCode := store.GetStore().MoveSegmentToNewFolder(projectId, id, request.Name, folderType)
	if errCode != http.StatusAccepted {
		if errCode == http.StatusNotFound {
			return nil, http.StatusNotFound, "Invalid SegmentID or FolderName","Segment ID or Folder Name provided are wrong", true
		}else if errCode == http.StatusConflict {
			return nil, http.StatusConflict, "Conflict With FolderName", "Conflict with FolderName", true
		}
		return nil, http.StatusInternalServerError, "Unexpected Error", "Unexpected Error Found", true
	}
	return nil, errCode, "","", false
}