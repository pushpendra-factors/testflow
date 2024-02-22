package v1

import (
	"encoding/json"
	C "factors/config"
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	AS "factors/task/account_scoring"
	U "factors/util"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

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
		return "", http.StatusBadRequest, INVALID_PROJECT, "invalid project_id", true
	}

	req := c.Request

	var payload model.TimelinePayload

	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&payload)

	logCtx = log.WithFields(log.Fields{
		"projectId": projectId,
		"payload":   payload,
	})
	if err != nil {
		logCtx.Error("Json decode failed.")
		message := "Query failed. Invalid payload"
		return nil, http.StatusBadRequest, INVALID_INPUT, message, true
	}

	profileUsersList, errCode, errMsg := store.GetStore().GetProfilesListByProjectId(projectId, payload, model.PROFILE_TYPE_USER, false)
	if errCode != http.StatusFound {
		logCtx.Error("User profiles not found. " + errMsg)
		return nil, errCode, "", errMsg, true
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

func updateProfileUserScores(profileUsersList *[]model.Profile, scoresPerUser map[string]model.PerUserScoreOnDay, buckets model.BucketRanges) {
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

	engagementLevels := model.GetEngagementLevels(scores, buckets)

	for i := range *profileUsersList {
		if prof, ok := scoresPerUser[(*profileUsersList)[i].Identity]; ok {
			(*profileUsersList)[i].Engagement = engagementLevels[float64(prof.Score)]
			(*profileUsersList)[i].TopEngagements = make(map[string]float64)
			for engagementKey, engengagementval := range prof.TopEvents {
				(*profileUsersList)[i].TopEngagements[engagementKey] += engengagementval
			}
		} else {
			(*profileUsersList)[i].Engagement = ""
		}
	}
}

func calculatePercentile(data []float64, value float64) float64 {
	sort.Float64s(data)                                       // Sort the data in ascending order
	index := sort.SearchFloat64s(data, value)                 // Find the index of the value
	percentile := float64(index) / float64(len(data)-1) * 100 // Calculate the percentile based on the index
	return percentile
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
		return "", http.StatusBadRequest, INVALID_PROJECT, "invalid project_id", true
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

	getUserMarker, err := getBoolQueryParam(c.Query("user_marker"))
	if err != nil {
		logCtx.Error("Invalid marker flag.")
	}

	downloadLimitGiven, err := getBoolQueryParam(c.Query("download"))
	if err != nil {
		logCtx.Error("Invalid limit flag.")
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
		return nil, http.StatusBadRequest, INVALID_INPUT, message, true
	}

	var profileAccountsList []model.Profile
	var errCode int
	var errMsg string

	startTime := time.Now().UnixMilli()
	if getUserMarker && C.UseSegmentMarker(projectId) {
		profileAccountsList, errCode, errMsg = store.GetStore().GetMarkedDomainsListByProjectId(projectId, payload, downloadLimitGiven)
	} else {
		profileAccountsList, errCode, errMsg = store.GetStore().GetProfilesListByProjectId(projectId, payload, model.PROFILE_TYPE_ACCOUNT, downloadLimitGiven)
	}
	endTime := time.Now().UnixMilli()
	if timeTaken := endTime - startTime; timeTaken > 2000 {
		logCtx.Warn("Accounts time exceeded 2 seconds. Time taken is ", timeTaken)
	}
	if errCode != http.StatusFound {
		logCtx.Error("Account profiles not found. " + errMsg)
		return "", errCode, "", errMsg, true
	}

	scoringAvailable, err := store.GetStore().GetFeatureStatusForProjectV2(projectId, model.FEATURE_ACCOUNT_SCORING, false)
	if err != nil {
		logCtx.Error("Error fetching scoring availability status for the project")
	}

	showScore := getScore || C.IsScoringEnabledForAllUsers(projectId)

	// Add account scores to the response if scoring is enabled
	if scoringAvailable && showScore {
		// Retrieve scores for account IDs
		var accountIds []string
		for _, profile := range profileAccountsList {
			accountIds = append(accountIds, profile.Identity)
		}
		scoresPerAccount, err := store.GetStore().GetAccountScoreOnIds(projectId, accountIds, getDebug)
		if err != nil {
			logCtx.Error("Error while fetching account scores.")
		} else {

			buckets, err := AS.GetEngagementBuckets(projectId, scoresPerAccount)
			if err != nil {
				logCtx.Error("Error while fetching account scoring bucket ranges.")
			}

			// Update account scores in the accounts list
			updateProfileUserScores(&profileAccountsList, scoresPerAccount, buckets)
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

func GetProfileAccountOverviewHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	id := c.Params.ByName("id")
	group := c.Params.ByName("group")

	logCtx := log.WithFields(log.Fields{
		"projectId": projectId,
		"userId":    id,
		"group":     group,
	})

	if projectId == 0 {
		logCtx.Error("Invalid project_id.")
		return "", http.StatusBadRequest, "", "invalid project_id", true
	}

	if id == "" {
		logCtx.Error("Invalid userId.")
		return nil, http.StatusBadRequest, INVALID_INPUT, "invalid userId", true
	}

	if group == "" {
		logCtx.Error("Invalid group name.")
		return nil, http.StatusBadRequest, INVALID_INPUT, "invalid group name", true
	}

	log.Warn("Parameters Passed")

	scoringAvailable, err := store.GetStore().GetFeatureStatusForProjectV2(projectId, model.FEATURE_ACCOUNT_SCORING, false)
	if err != nil {
		logCtx.Error("Error fetching scoring availability status for project ID.")
		return nil, http.StatusBadRequest, PROCESSING_FAILED, "Scoring Unavailable for this project", true
	}

	showScore := C.IsScoringEnabledForAllUsers(projectId)

	// Add user scores to the response if scoring is enabled
	if !scoringAvailable || !showScore {
		return nil, http.StatusBadRequest, PROCESSING_FAILED, "Scoring Unavailable for this project", true
	}

	log.Warn("Feature Flags Passed")

	accountOverview, errCode, errMsg := store.GetStore().GetAccountOverview(projectId, id, group)
	log.Warn("Returns from Method Call-", errCode, errMsg)
	if errCode != http.StatusOK {
		logCtx.Error("Account details not found. " + errMsg)
		return nil, errCode, PROCESSING_FAILED, errMsg, true
	}

	return accountOverview, http.StatusOK, "", "", false
}

func getNewEventConfig(c *gin.Context) (*[]string, error) {
	payload := []string{}
	err := c.BindJSON(&payload)
	if err != nil {
		return nil, err
	}
	return &payload, nil
}

func UpdateEventConfigHandler(c *gin.Context) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	logCtx := log.WithFields(log.Fields{
		"projectId": projectID,
	})

	eventName := c.Params.ByName("event_name")

	logCtx = log.WithFields(log.Fields{
		"projectId": projectID,
		"eventName": eventName,
	})

	if eventName == "" {
		logCtx.Error("Invalid event to update")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	payload, err := getNewEventConfig(c)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	errCode, err := store.GetStore().UpdateConfigForEvent(projectID, eventName, *payload)
	if errCode != http.StatusOK {
		logCtx.Errorln(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"status":     "success",
		"event_name": eventName,
	}
	c.JSON(http.StatusOK, response)
}
