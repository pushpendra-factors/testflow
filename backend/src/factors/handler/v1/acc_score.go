package v1

import (
	"encoding/json"
	mid "factors/middleware"
	"factors/model/model"
	M "factors/model/model"
	"factors/model/store"
	T "factors/task"
	U "factors/util"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// UpdateAccScoreWeights updates weights for a given project
func UpdateAccScoreWeights(c *gin.Context) (interface{}, int, string, string, bool) {
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	reqID, _ := getReqIDAndProjectID(c)

	logCtx := log.WithFields(log.Fields{
		"projectId": projectId,
		"RequestId": reqID,
	})

	var weightsRequest M.AccWeights
	var weights M.AccWeights
	weights.WeightConfig = make([]M.AccEventWeight, 0)
	r := c.Request
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&weightsRequest); err != nil {
		errMsg := "Unable to decode weights Invalid JSON."
		logCtx.WithError(err).Error(errMsg)
		return nil, http.StatusBadRequest, errMsg, "", true
	}
	weights.SaleWindow = weightsRequest.SaleWindow

	// convert incoming request to AccWeights.
	for _, wtVal := range weightsRequest.WeightConfig {
		var r M.AccEventWeight
		r.EventName = wtVal.EventName
		r.Is_deleted = wtVal.Is_deleted
		r.Rule = wtVal.Rule
		r.WeightId = wtVal.WeightId
		r.Weight_value = wtVal.Weight_value
		weights.WeightConfig = append(weights.WeightConfig, r)
	}

	dedupweights, err := T.DeduplicateWeights(weights)
	if err != nil {
		errMsg := "Unable to dedup weights."
		logCtx.WithError(err).Error(errMsg)
		return nil, http.StatusBadRequest, errMsg, "", true
	}

	logCtx.WithField("weights", weights).Infof("Updating weights for project")

	// check if project score exist
	// else create new
	// if score column exist, check if new value added and old value deleted or updated
	err = store.GetStore().UpdateAccScoreWeights(projectId, dedupweights)
	if err != nil {
		errMsg := "Unable to update weights."
		logCtx.WithError(err).Error(errMsg)
		return nil, http.StatusInternalServerError, errMsg, "", true
	}
	// if project score exist
	return dedupweights, http.StatusOK, "", "", false

}

// GetAccountScore returns account score for a given date and group_id
func GetAccountScores(c *gin.Context) (interface{}, int, string, string, bool) {
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	reqID, _ := getReqIDAndProjectID(c)

	groupIdString := c.Query("group_id")
	dateString := c.Query("date")
	debugFlag := c.Query("debug")

	logCtx := log.WithFields(log.Fields{
		"projectId": projectId,
		"RequestId": reqID,
		"group_id":  groupIdString,
		"date":      dateString,
	})

	var accountScores model.AccScoreResult
	accountScores.ProjectId = projectId
	groupId, _ := strconv.Atoi(groupIdString)
	debug, _ := strconv.ParseBool(debugFlag)

	logCtx.Info("getting account scores")
	perAccScore, weights, err := store.GetStore().GetAccountsScore(projectId, groupId, dateString, debug)
	if err != nil {
		errMsg := "Unable to get account score."
		logCtx.WithError(err).Error(errMsg)
		return nil, http.StatusInternalServerError, "", "", true
	}

	accountScores.AccResult = make([]model.PerAccountScore, len(perAccScore))
	accountScores.AccResult = perAccScore
	if debug {
		accountScores.Debug = make(map[string]interface{})
		accountScores.Debug["weights"] = weights
	}
	return accountScores, http.StatusOK, "", "", false
}

// GetAccountScore returns account score for a given date and user_id
func GetPerAccountScore(c *gin.Context) (interface{}, int, string, string, bool) {
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	reqID, _ := getReqIDAndProjectID(c)

	userId := c.Query("id")
	dateString := c.Query("date")
	numDaysTrend := c.Query("trend")
	debugFlag := c.Query("debug")

	logCtx := log.WithFields(log.Fields{
		"projectId": projectId,
		"RequestId": reqID,
		"userId":    userId,
		"date":      dateString,
	})

	var accountScores model.AccScoreResult
	var numDaysToTrend int
	var err error
	accountScores.ProjectId = projectId
	debug, _ := strconv.ParseBool(debugFlag)

	numDaysToTrend, err = strconv.Atoi(numDaysTrend)
	if err != nil {
		errMsg := "Unable to convert number of days to trend."
		logCtx.WithError(err).Error(errMsg)
		return nil, http.StatusInternalServerError, "", "", true
	}

	if numDaysToTrend == 0 {
		numDaysToTrend = model.NUM_TREND_DAYS
	}

	logCtx.Info("getting account scores")
	perAccScore, weights, err := store.GetStore().GetPerAccountScore(projectId, dateString, userId, numDaysToTrend, debug)
	if err != nil {
		errMsg := "Unable to get account score."
		logCtx.WithError(err).Error(errMsg)
		return nil, http.StatusInternalServerError, "", "", true
	}

	accountScores.AccResult = make([]model.PerAccountScore, 1)
	accountScores.AccResult[0] = perAccScore
	if debug {
		accountScores.Debug = make(map[string]interface{})
		accountScores.Debug["weights"] = weights
	}
	return accountScores, http.StatusOK, "", "", false
}

// GetUserScore returns the score for a user on a given date
func GetUserScore(c *gin.Context) (interface{}, int, string, string, bool) {
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	reqID, _ := getReqIDAndProjectID(c)

	var UsersScores model.AccUserScoreResult
	var dateString string
	var userId string

	userId = c.Query("user_id")
	dateString = c.Query("date")
	debugFlag := c.Query("debug")
	is_anonymus_string := c.Query("is_anonymus")

	logCtx := log.WithFields(log.Fields{
		"projectId": projectId,
		"RequestId": reqID,
		"user_id":   userId,
		"date":      dateString,
	})

	logCtx.Infof("Getting scores for user")
	debug, _ := strconv.ParseBool(debugFlag)
	is_anonymus, _ := strconv.ParseBool(is_anonymus_string)

	UsersScores.ProjectId = projectId
	if dateString == "" {
		current_time := time.Now()
		prev_day := current_time.AddDate(0, 0, -1)
		dateString = T.GetDateOnlyFromTimestamp(prev_day.Unix())
	}
	perAccScore, err := store.GetStore().GetUserScore(projectId, userId, dateString, debug, is_anonymus)
	if err != nil {
		errMsg := "Unable to get user score."
		logCtx.WithError(err).Error(errMsg)
		return nil, http.StatusInternalServerError, "", "", true
	}
	UsersScores.AccResult = make([]model.PerUserScoreOnDay, 0)
	UsersScores.AccResult = append(UsersScores.AccResult, perAccScore)
	// if project score exist
	return UsersScores, http.StatusOK, "", "", false

}

// GetAllUsersScores returns all users scores
func GetAllUsersScores(c *gin.Context) (interface{}, int, string, string, bool) {
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	reqID, _ := getReqIDAndProjectID(c)

	debug_flag := c.Query("debug")
	date := c.Query("date")
	logCtx := log.WithFields(log.Fields{
		"projectId": projectId,
		"RequestId": reqID,
	})

	var accountScores model.UserScoreResult
	accountScores.ProjectId = projectId
	debug, _ := strconv.ParseBool(debug_flag)
	var perAccScore []M.AllUsersScore
	var err error
	var weights *M.AccWeights
	if date == "" {
		logCtx.Info("getting account scores")
		perAccScore, weights, err = store.GetStore().GetAllUserScore(projectId, debug)
		if err != nil {
			errMsg := "Unable to get all user score."
			logCtx.WithError(err).Error(errMsg)
			return nil, http.StatusInternalServerError, "", "", true
		}
	} else if date == model.LAST_EVENT {
		logCtx.Info("getting all user scores latest scores")
		perAccScore, weights, err = store.GetStore().GetAllUserScoreLatest(projectId, debug)
		if err != nil {
			errMsg := "Unable all user latest score."
			logCtx.WithError(err).Error(errMsg)
			return nil, http.StatusInternalServerError, "", "", true
		}
	} else {
		logCtx.Info("getting all user scores on day")
		perAccScore, weights, err = store.GetStore().GetAllUserScoreOnDay(projectId, date, debug)
		if err != nil {
			errMsg := "Unable all user latest score on day."
			logCtx.WithError(err).Error(errMsg)
			return nil, http.StatusInternalServerError, "", "", true
		}
	}

	accountScores.AccResult = make([]model.AllUsersScore, len(perAccScore))
	accountScores.AccResult = perAccScore
	if debug {
		accountScores.Debug = make(map[string]interface{})
		accountScores.Debug["weights"] = weights
	}
	return accountScores, http.StatusOK, "", "", false

}
