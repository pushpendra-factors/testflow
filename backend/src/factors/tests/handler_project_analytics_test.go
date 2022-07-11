package tests

import (
	"encoding/json"
	cacheRedis "factors/cache/redis"
	C "factors/config"
	H "factors/handler"
	"factors/handler/helpers"
	"factors/model/model"
	"factors/model/store/memsql"
	U "factors/util"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func sendGetProjectAnalyticsReq(r *gin.Engine, projectId int64, agent *model.Agent) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}
	rb := C.NewRequestBuilderWithPrefix(http.MethodGet, "/projectanalytics").
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	w := httptest.NewRecorder()
	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error Creating getProjectAnalytics Req")
	}
	r.ServeHTTP(w, req)
	return w

}

func CreateUserCache(dateKey string, project *model.Project) error {

	analyticsKeysInCache := make([]cacheRedis.SortedSetKeyValueTuple, 0)

	// --------unique users--------------
	uniqueUsersCountKey, err := model.UserCountAnalyticsCacheKey(dateKey)
	if err != nil {
		return err
	}
	analyticsKeysInCache = append(analyticsKeysInCache, cacheRedis.SortedSetKeyValueTuple{
		Key:   uniqueUsersCountKey,
		Value: fmt.Sprintf("%v", project.ID),
	})

	// --------unique events--------------
	uniqueEventsCountKey, err := model.UniqueEventNamesAnalyticsCacheKey(dateKey)
	if err != nil {
		return err
	}
	analyticsKeysInCache = append(analyticsKeysInCache, cacheRedis.SortedSetKeyValueTuple{
		Key:   uniqueEventsCountKey,
		Value: fmt.Sprintf("%v", project.ID)})

	// ----------total events----------------
	totalEventsCountKey, err := model.EventsCountAnalyticsCacheKey(dateKey)
	if err != nil {
		return err
	}
	analyticsKeysInCache = append(analyticsKeysInCache, cacheRedis.SortedSetKeyValueTuple{
		Key:   totalEventsCountKey,
		Value: fmt.Sprintf("%v", project.ID)})

	cacheRedis.ZincrPersistentBatch(true, analyticsKeysInCache...)

	return nil
}

func incrementCacheCountOfAllDocumentTypes(project *model.Project, currentDate time.Time) bool {
	status := true
	documentType := []string{"adwords", "linkedin", "facebook", "hubspot", "salesforce"}
	for _, doc := range documentType {
		status = memsql.UpdateCountCacheByDocumentType(project.ID, &currentDate, doc)
		if !status {
			return status
		}
	}
	return status
}

func validateData(data *model.ProjectAnalytics, projectId int64) bool {
	if data.ProjectID != projectId {
		return false
	}
	values := []int{int(data.AdwordsEvents), int(data.FacebookEvents),
		int(data.HubspotEvents), int(data.LinkedinEvents),
		int(data.SalesforceEvents), int(data.TotalEvents),
		int(data.TotalUniqueEvents), int(data.TotalUniqueUsers)}

	for i := 0; i < len(values); i++ {
		if values[i] != 1 {
			return false
		}
	}
	return true
}

func TestGetFactorsAnalyticsHandler(t *testing.T) {

	currentDate := time.Now().UTC()
	dateKey := currentDate.AddDate(0, 0, 0).Format(U.DATETIME_FORMAT_YYYYMMDD)

	r := gin.Default()
	H.InitAppRoutes(r)

	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)
	assert.NotNil(t, agent)

	C.GetConfig().ProjectAnalyticsWhitelistedUUIds = []string{agent.UUID}

	status := incrementCacheCountOfAllDocumentTypes(project, currentDate)
	assert.Equal(t, true, status)

	err = CreateUserCache(dateKey, project)
	assert.Nil(t, err)

	w := sendGetProjectAnalyticsReq(r, project.ID, agent)
	assert.Equal(t, http.StatusOK, w.Code)

	var analytics map[string][]*model.ProjectAnalytics
	decoder := json.NewDecoder(w.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&analytics); err != nil {
		assert.NotNil(t, nil, err)
	}
	length := len(analytics[dateKey])

	// sorting to get the current project id at last idx
	sort.Slice(analytics[dateKey], func(i, j int) bool {
		return analytics[dateKey][i].ProjectID < analytics[dateKey][j].ProjectID
	})
	valid := validateData(analytics[dateKey][length-1], project.ID)
	assert.Equal(t, true, valid)
	C.GetConfig().ProjectAnalyticsWhitelistedUUIds = []string{}

}
