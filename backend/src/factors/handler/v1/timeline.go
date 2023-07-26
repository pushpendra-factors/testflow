package v1

import (
	"encoding/json"
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	C "factors/config"

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

	profileUsersList, errCode, errMsg := store.GetStore().GetProfilesListByProjectId(projectId, payload, model.PROFILE_TYPE_USER)
	if errCode != http.StatusFound {
		logCtx.Error("User profiles not found. " + errMsg)
		return nil, errCode, "", errMsg, true
	}

	// Add user scores to the response if scoring is enabled
	if getScore || C.IsScoringEnabled(projectId) {
		// Separate anonymous and known user IDs
		var userIdsAnonymous []string
		var userIdsNonAnonymous []string
		for _, profile := range profileUsersList {
			if profile.IsAnonymous {
				userIdsAnonymous = append(userIdsAnonymous, profile.Identity) //anonymus=true
			} else {
				userIdsNonAnonymous = append(userIdsNonAnonymous, profile.Identity) // anonymus=false
			}
		}
		// Retrieve user scores
		scoresPerUser, err := store.GetStore().GetUserScoreOnIds(projectId, userIdsAnonymous, userIdsNonAnonymous, getDebug)
		if err != nil {
			logCtx.Error("Error while fetching user scores.")
		} else {
			// Update user scores in the users list
			updateProfileUserScores(&profileUsersList, scoresPerUser)
		}
	}

	profileUsersList = store.GetStore().AddPropertyValueLabelsToProfileResults(projectId, profileUsersList)
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
	result := make(map[float64]string)
	for _, score := range scores {
		percentile := calculatePercentile(scores, score)
		result[score] = getEngagement(percentile)
	}
	return result
}

func calculatePercentile(data []float64, value float64) float64 {
	sort.Float64s(data)                                       // Sort the data in ascending order
	index := sort.SearchFloat64s(data, value)                 // Find the index of the value
	percentile := float64(index) / float64(len(data)-1) * 100 // Calculate the percentile based on the index
	return percentile
}

func getEngagement(percentile float64) string {
	if percentile > 90 {
		return "Hot"
	} else if percentile > 70 {
		return "Warm"
	} else {
		return "Cool"
	}
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

	userDetails, errCode, errMsg := store.GetStore().GetProfileUserDetailsByID(projectId, identity, isAnonymous)
	if errCode != http.StatusFound {
		logCtx.Error("User details not found." + errMsg)
		return nil, errCode, PROCESSING_FAILED, errMsg, true
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

	startTime := time.Now().UnixMilli()
	profileAccountsList, errCode, errMsg := store.GetStore().GetProfilesListByProjectId(projectId, payload, model.PROFILE_TYPE_ACCOUNT)
	endTime := time.Now().UnixMilli()
	if timeTaken := endTime - startTime; timeTaken > 2000 {
		logCtx.Warn("Accounts time exceeded 2 seconds. Time taken is ", timeTaken)
	}
	if errCode != http.StatusFound {
		logCtx.Error("Account profiles not found. " + errMsg)
		return "", errCode, "", errMsg, true
	}

	// Add account scores to the response if scoring is enabled
	if getScore || C.IsScoringEnabled(projectId) {
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

	profileAccountsList = store.GetStore().AddPropertyValueLabelsToProfileResults(projectId, profileAccountsList)
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

	accountDetails, errCode, errMsg := store.GetStore().GetProfileAccountDetailsByID(projectId, id, group)
	if errCode != http.StatusFound {
		logCtx.Error("Account details not found. " + errMsg)
		return nil, errCode, PROCESSING_FAILED, errMsg, true
	}

	return accountDetails, http.StatusOK, "", "", false
}
