package v1

import (
	"encoding/json"
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"net/http"
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

	score := c.Query("score")
	if score == "" {
		score = "false"
	}
	get_score, err := strconv.ParseBool(score)
	if err != nil {
		logCtx.Error("Invalid score flag .")
		get_score = false
	}

	debug := c.Query("debug")
	if debug == "" {
		debug = "false"
	}
	get_debug, err := strconv.ParseBool(debug)
	if err != nil {
		logCtx.Error("Invalid debug flag.")
		get_debug = false
	}

	r := c.Request
	var payload model.TimelinePayload
	logCtx = log.WithFields(log.Fields{
		"projectId": projectId,
		"payload":   payload,
	})
	decoder := json.NewDecoder(r.Body)
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

	if get_score {
		// add score to the response based on the user scores from acc scoring
		var userIdsUnknown []string = make([]string, 0)
		var userIdsKnown []string = make([]string, 0)
		for _, profile := range profileUsersList {
			if profile.IsAnonymous {
				userIdsUnknown = append(userIdsUnknown, profile.Identity)
			} else {
				userIdsKnown = append(userIdsKnown, profile.Identity)
			}
		}
		ScoresPerUser, err := store.GetStore().GetUserScoreOnIds(projectId, userIdsUnknown, userIdsKnown, get_debug)
		if err != nil {
			logCtx.Error("Error while fetching user scores.")
			return nil, http.StatusInternalServerError, "", "", true
		}

		updateProfileUserScores(&profileUsersList, ScoresPerUser)
	}

	return profileUsersList, http.StatusOK, "", "", false
}

func updateProfileUserScores(profileUsersList *[]model.Profile, scoresPerUser map[string]model.PerUserScoreOnDay) {
	for _, profile := range *profileUsersList {
		if prof, ok := scoresPerUser[profile.Identity]; ok {
			profile.Score = float64(prof.Score)
		} else {
			profile.Score = 0
		}
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
	score := c.Query("score")
	if score == "" {
		score = "false"
	}
	get_score, err := strconv.ParseBool(score)
	if err != nil {
		logCtx.Error("Invalid score flag.")
		get_score = false
	}

	debug := c.Query("debug")
	if debug == "" {
		debug = "false"
	}
	get_debug, err := strconv.ParseBool(debug)
	if err != nil {
		logCtx.Error("Invalid debug flag.")
		get_debug = false

	}

	r := c.Request
	var payload model.TimelinePayload
	logCtx = log.WithFields(log.Fields{
		"projectId": projectId,
		"payload":   payload,
	})
	decoder := json.NewDecoder(r.Body)
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

	// add accounts scores to the response based on account scoring enabled
	if get_score {
		// get score for accountsIds from acc scoring

		var accountIds []string = make([]string, 0)
		for _, profile := range profileAccountsList {
			accountIds = append(accountIds, profile.Identity)
		}
		ScoresPerAccount, err := store.GetStore().GetAccountScoreOnIds(projectId, accountIds, get_debug)
		if err != nil {
			logCtx.Error("Error while fetching account scores.")
			return nil, http.StatusInternalServerError, "", "", true
		}
		updateProfileUserScores(&profileAccountsList, ScoresPerAccount)

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
