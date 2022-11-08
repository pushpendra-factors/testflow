package v1

import (
	mid "factors/middleware"
	"factors/model/store"
	U "factors/util"
	"net/http"

	"github.com/gin-gonic/gin"
	"strings"
)

type CreateDisplayNamesParams struct {
	EventName    string `json:"eventName"`
	PropertyName string `json:"propertyName"`
	Tag          string `json:"tag"`
	DisplayName  string `json:"displayName"`
}

func GetcreateDisplayNamesParams(c *gin.Context) (*CreateDisplayNamesParams, error) {
	params := CreateDisplayNamesParams{}
	err := c.BindJSON(&params)
	if err != nil {
		return nil, err
	}
	return &params, nil
}

func CreateDisplayNamesHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		return nil, http.StatusBadRequest, INVALID_PROJECT, "", true
	}
	params, err := GetcreateDisplayNamesParams(c)
	if err != nil {
		return nil, http.StatusBadRequest, INVALID_INPUT, "", true
	}
	errCode := store.GetStore().CreateOrUpdateDisplayName(projectID, params.EventName, params.PropertyName, params.DisplayName, params.Tag)
	if errCode != http.StatusCreated {
		if errCode == http.StatusConflict {
			return nil, http.StatusConflict, DUPLICATE_RECORD, "", true
		}
		return nil, http.StatusInternalServerError, PROCESSING_FAILED, "", true
	}
	response := make(map[string]interface{})
	response["status"] = "success"
	return response, http.StatusCreated, "", "", false
}

func GetAllDistinctEventProperties(c *gin.Context)(interface{}, int, string, string, bool) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		return nil, http.StatusBadRequest, INVALID_PROJECT, "", true
	}
	finalDisplayNames := make(map[string]map[string]string)
	_, displayNames := store.GetStore().GetDistinctDisplayNamesForAllEventProperties(projectID)
	displayNamesOp := make(map[string]string)
	for property, displayName := range displayNames {
		displayNamesOp[property] = strings.Title(displayName)
	}
	standardPropertiesAllEvent := U.STANDARD_EVENT_PROPERTIES_DISPLAY_NAMES
	for property, displayName := range standardPropertiesAllEvent {
		displayNamesOp[property] = strings.Title(displayName)
	}
	_, displayNames = store.GetStore().GetDisplayNamesForObjectEntities(projectID)
	for property, displayName := range displayNames {
		displayNamesOp[property] = strings.Title(displayName)
	}

	finalDisplayNames["global"] = displayNamesOp
	standardPropertiesSession := U.STANDARD_SESSION_PROPERTIES_DISPLAY_NAMES
	for property, displayName := range standardPropertiesSession {
		displayNamesOp[property] = strings.Title(displayName)
	}
	
	finalDisplayNames["$session"] = displayNamesOp

	return finalDisplayNames, http.StatusOK, "", "", false
}