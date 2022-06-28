package v1

import (
	"encoding/json"
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetProfileUsersHandler(c *gin.Context) (interface{}, int, string, string, bool) {

	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return "", http.StatusBadRequest, "", "invalid project_id", true
	}

	r := c.Request
	var payload model.UTListPayload
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		message := fmt.Sprintf("Query failed. Invalid user source provided : %s", payload.Source)
		return nil, http.StatusBadRequest, "", message, true
	}

	profileUsersList, errCode := store.GetStore().GetProfileUsersListByProjectId(projectId, payload)

	if errCode != http.StatusFound {
		c.AbortWithStatus(errCode)
		return "", http.StatusNotFound, "", "", true
	}

	return profileUsersList, http.StatusOK, "", "", false
}

func GetProfileUserDetailsHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return "", http.StatusBadRequest, "", "", true
	}
	identity := c.Params.ByName("id")
	if identity == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return nil, http.StatusBadRequest, INVALID_INPUT, "", true
	}

	isAnonymous := c.Query("is_anonymous")

	if isAnonymous == "" {
		return nil, http.StatusBadRequest, INVALID_INPUT, "", true
	}

	userDetails, errCode := store.GetStore().GetProfileUserDetailsByID(projectId, identity, isAnonymous)
	if errCode != http.StatusFound {
		return nil, errCode, PROCESSING_FAILED, "Failed to get user details", true
	}

	return userDetails, http.StatusOK, "", "", false
}
