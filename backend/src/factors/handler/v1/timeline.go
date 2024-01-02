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

	profileUsersList, errCode, errMsg := store.GetStore().GetProfilesListByProjectId(projectId, payload, model.PROFILE_TYPE_USER)
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

	engagementLevels := GetEngagementLevels(scores, buckets)

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

func removeZeros(input []float64) []float64 {
	var result []float64

	for _, value := range input {
		if value != 0 {
			result = append(result, value)
		}
	}
	return result
}

func getUniqueScores(input []float64) []float64 {
	uniqueMap := make(map[float64]bool)
	result := []float64{}

	for _, item := range input {
		if _, found := uniqueMap[item]; !found {
			uniqueMap[item] = true
			result = append(result, item)
		}
	}

	return result
}

func GetEngagementLevels(scores []float64, buckets model.BucketRanges) map[float64]string {
	result := make(map[float64]string)
	result[0] = model.GetEngagement(0, buckets)

	nonZeroScores := removeZeros(scores)
	uniqueScores := getUniqueScores(nonZeroScores)

	for _, score := range uniqueScores {
		// calculating percentile is not used in the current implementation
		result[score] = model.GetEngagement(score, buckets)
	}

	return result
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
	if getUserMarker {
		profileAccountsList, errCode, errMsg = store.GetStore().GetMarkedDomainsListByProjectId(projectId, payload)
	} else {
		profileAccountsList, errCode, errMsg = store.GetStore().GetProfilesListByProjectId(projectId, payload, model.PROFILE_TYPE_ACCOUNT)
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

	accountOverview, errCode, errMsg := store.GetStore().GetAccountOverview(projectId, id, group)
	if errCode != http.StatusOK {
		logCtx.Error("Account details not found. " + errMsg)
		return nil, errCode, PROCESSING_FAILED, errMsg, true
	}

	return accountOverview, http.StatusOK, "", "", false
}
