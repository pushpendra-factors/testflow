package v1

import (
	"encoding/json"
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func GetProfileUsersHandler(c *gin.Context) (interface{}, int, string, string, bool) {

	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	logCtx := log.WithFields(log.Fields{
		"projectId": projectId,
	})
	if projectId == 0 {
		logCtx.Error("Invalid project_id.")
		return "", http.StatusBadRequest, "", "invalid project_id", true
	}

	req := c.Request

	getScore, err := getBoolQueryParam(c.Query("score"))
	if err != nil {
		logCtx.Error("Invalid score flag .")
	}

	getDebug, err := getBoolQueryParam(c.Query("debug"))
	if err != nil {
		logCtx.Error("Invalid debug flag.")
	}

	var payload model.TimelinePayload
	logCtx = log.WithFields(log.Fields{
		"projectId": projectId,
		"payload":   payload,
	})
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		logCtx.Error("Json decode failed.")
		message := fmt.Sprintf("Query failed. Invalid user source provided : %s", payload.Source)
		return nil, http.StatusBadRequest, "", message, true
	}

	profileUsersList, errCode := store.GetStore().GetProfilesListByProjectId(projectId, payload, model.PROFILE_TYPE_USER)
	if errCode != http.StatusFound {
		logCtx.Error("User profiles not found.")
		return nil, errCode, "", "", true
	}

	// Add user scores to the response if scoring is enabled
	if getScore {
		// Separate anonymous and known user IDs
		var userIdsUnknown []string
		var userIdsKnown []string
		for _, profile := range profileUsersList {
			if profile.IsAnonymous {
				userIdsUnknown = append(userIdsUnknown, profile.Identity)
			} else {
				userIdsKnown = append(userIdsKnown, profile.Identity)
			}
		}
		// Retrieve user scores
		scoresPerUser, err := store.GetStore().GetUserScoreOnIds(projectId, userIdsUnknown, userIdsKnown, getDebug)
		if err != nil {
			logCtx.Error("Error while fetching user scores.")
		} else {
			// Update user scores in the users list
			updateProfileUserScores(&profileUsersList, scoresPerUser)
		}
	}

	return profileUsersList, http.StatusOK, "", "", false
}

func getBoolQueryParam(value string) (bool, error) {
	if value == "" {
		return false, nil
	}
	status, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("invalid boolean value: %s", value)
	}
	return status, nil
}

func updateProfileUserScores(profileUsersList *[]model.Profile, scoresPerUser map[string]model.PerUserScoreOnDay) {
	var scores []float64
	for i := range *profileUsersList {
		if prof, ok := scoresPerUser[(*profileUsersList)[i].Identity]; ok {
			(*profileUsersList)[i].Score = float64(prof.Score)
			scores = append(scores, float64(prof.Score))
		} else {
			(*profileUsersList)[i].Score = 0
			scores = append(scores, 0)
		}
	}

	engagementLevels := GetEngagementLevels(scores)

	for i := range *profileUsersList {
		if prof, ok := scoresPerUser[(*profileUsersList)[i].Identity]; ok {
			(*profileUsersList)[i].Engagement = engagementLevels[float64(prof.Score)]
		} else {
			(*profileUsersList)[i].Engagement = ""
		}
	}
}

func GetEngagementLevels(scores []float64) map[float64]string {
	sort.Float64s(scores)
	result := make(map[float64]string)
	top10PercentIdx := int(math.Round(float64(len(scores)) * 0.9))
	top30PercentIdx := int(math.Round(float64(len(scores)) * 0.7))

	for i, score := range scores {
		if i >= top10PercentIdx {
			result[score] = "Hot"
		} else if i >= top30PercentIdx {
			result[score] = "Warm"
		} else {
			result[score] = "Cool"
		}
	}
	return result
}

func GetProfileUserDetailsHandler(c *gin.Context) (interface{}, int, string, string, bool) {

	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	logCtx := log.WithFields(log.Fields{
		"projectId": projectId,
	})
	if projectId == 0 {
		logCtx.Error("Invalid project_id.")
		return "", http.StatusBadRequest, "", "invalid project_id", true
	}

	identity := c.Params.ByName("id")
	logCtx = log.WithFields(log.Fields{
		"projectId": projectId,
		"userId":    identity,
	})
	if identity == "" {
		logCtx.Error("Invalid userId.")
		return nil, http.StatusBadRequest, INVALID_INPUT, "invalid userId", true
	}

	isAnonymous := c.Query("is_anonymous")
	logCtx = log.WithFields(log.Fields{
		"projectId":   projectId,
		"userId":      identity,
		"isAnonymous": isAnonymous,
	})
	if isAnonymous == "" {
		logCtx.Error("Anonymity status not valid.")
		return nil, http.StatusBadRequest, INVALID_INPUT, "", true
	}

	userDetails, errCode := store.GetStore().GetProfileUserDetailsByID(projectId, identity, isAnonymous)
	if errCode != http.StatusFound {
		logCtx.Error("User details not found.")
		return nil, errCode, PROCESSING_FAILED, "Failed to get user details", true
	}

	return userDetails, http.StatusOK, "", "", false
}

func GetProfileAccountsHandler(c *gin.Context) (interface{}, int, string, string, bool) {

	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	logCtx := log.WithFields(log.Fields{
		"projectId": projectId,
	})
	if projectId == 0 {
		logCtx.Error("Invalid project_id.")
		return "", http.StatusBadRequest, "", "invalid project_id", true
	}

	req := c.Request

	getScore, err := getBoolQueryParam(c.Query("score"))
	if err != nil {
		logCtx.Error("Invalid score flag .")
	}

	getDebug, err := getBoolQueryParam(c.Query("debug"))
	if err != nil {
		logCtx.Error("Invalid debug flag.")
	}

	var payload model.TimelinePayload
	logCtx = log.WithFields(log.Fields{
		"projectId": projectId,
		"payload":   payload,
	})
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		logCtx.Error("Json decode failed.")
		message := "Query failed. Invalid payload"
		return nil, http.StatusBadRequest, "", message, true
	}

	profileAccountsList, errCode := store.GetStore().GetProfilesListByProjectId(projectId, payload, model.PROFILE_TYPE_ACCOUNT)
	if errCode != http.StatusFound {
		logCtx.Error("Account profiles not found.")
		return "", errCode, "", "", true
	}

	// Add account scores to the response if scoring is enabled
	if getScore {
		// Retrieve scores for account IDs
		var accountIds []string
		for _, profile := range profileAccountsList {
			accountIds = append(accountIds, profile.Identity)
		}
		scoresPerAccount, err := store.GetStore().GetAccountScoreOnIds(projectId, accountIds, getDebug)
		if err != nil {
			logCtx.Error("Error while fetching account scores.")
		} else {
			// Update account scores in the accounts list
			updateProfileUserScores(&profileAccountsList, scoresPerAccount)
		}
	}

	return profileAccountsList, http.StatusOK, "", "", false
}

func GetProfileAccountDetailsHandler(c *gin.Context) (interface{}, int, string, string, bool) {

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
		"userId":    id,
	})
	if id == "" {
		logCtx.Error("Invalid userId.")
		return nil, http.StatusBadRequest, INVALID_INPUT, "invalid userId", true
	}
	group := c.Params.ByName("group")
	logCtx = log.WithFields(log.Fields{
		"projectId": projectId,
		"userId":    id,
		"group":     group,
	})
	if group == "" {
		logCtx.Error("Invalid group name.")
		return nil, http.StatusBadRequest, INVALID_INPUT, "invalid group name", true
	}

	accountDetails, errCode := store.GetStore().GetProfileAccountDetailsByID(projectId, id, group)
	if errCode != http.StatusFound {
		logCtx.Error("Account details not found.")
		return nil, errCode, PROCESSING_FAILED, "Failed to get account details", true
	}

	return accountDetails, http.StatusOK, "", "", false
}
