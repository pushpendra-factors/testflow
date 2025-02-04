package tests

import (
	"factors/cache"
	cacheRedis "factors/cache/redis"
	"factors/company_enrichment/demandbase"
	"factors/company_enrichment/factors_deanon"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"net/http"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if m.DoFunc != nil {
		return m.DoFunc(req)
	}
	return nil, nil
}

func TestFactorsDeanonAccountLimitAlerts(t *testing.T) {

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	errCode, _, err := store.GetStore().UpdateProjectPlanMappingField(project.ID, "CUSTOM")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusAccepted, errCode)

	err = UpdateAccountLimtForTesting(project.ID, 10)
	assert.Nil(t, err)

	var factorsDeanonObj factors_deanon.FactorsDeanon
	logCtx := log.WithField("project_id", project.ID)

	t.Run("TestAccountLimitAlertForPartialLimitExceeded", func(t *testing.T) {

		AccountLimitCountIncrementForTesting(project.ID, 9)

		mockClient := &MockHTTPClient{}
		// Set up the behavior of the mock
		mockClient.DoFunc = func(req *http.Request) (*http.Response, error) {
			// Return a mock response
			return &http.Response{
				StatusCode: 200,
				Body:       http.NoBody,
			}, nil
		}

		errCode, err := factorsDeanonObj.HandleAccountLimitAlert(project.ID, mockClient, logCtx)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, errCode)

		// testing if key is set or not if the execute is done.
		alertKey, _ := factors_deanon.GetAccountLimitEmailAlertCacheKey(project.ID, 10, factors_deanon.ACCOUNT_LIMIT_PARTIAL_EXCEEDED, U.TimeZoneStringIST, logCtx)
		exists, _ := cacheRedis.ExistsPersistent(alertKey)
		assert.Equal(t, true, exists)

		//Testing by sending the alert again
		errCode, err = factorsDeanonObj.HandleAccountLimitAlert(project.ID, mockClient, logCtx)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusForbidden, errCode)

		DeleteAlertAndAccLimitRedisKeyAfterTesting(project.ID, factors_deanon.ACCOUNT_LIMIT_PARTIAL_EXCEEDED, logCtx)

	})

	t.Run("TestAccountLimitAlertForPartialLimitExceededWhenExecuteFailed", func(t *testing.T) {

		AccountLimitCountIncrementForTesting(project.ID, 9)

		mockClient := &MockHTTPClient{}
		// Set up the behavior of the mock
		mockClient.DoFunc = func(req *http.Request) (*http.Response, error) {
			// Return a mock response
			return &http.Response{
				StatusCode: 400,
				Body:       http.NoBody,
			}, nil
		}

		errCode, err := factorsDeanonObj.HandleAccountLimitAlert(project.ID, mockClient, logCtx)
		assert.NotNil(t, err)
		assert.Equal(t, http.StatusBadRequest, errCode)

		// testing if key is set or not if the execute failed
		alertKey, _ := factors_deanon.GetAccountLimitEmailAlertCacheKey(project.ID, 10, factors_deanon.ACCOUNT_LIMIT_PARTIAL_EXCEEDED, U.TimeZoneStringIST, logCtx)
		exists, _ := cacheRedis.ExistsPersistent(alertKey)
		assert.Equal(t, false, exists)

		DeleteAlertAndAccLimitRedisKeyAfterTesting(project.ID, factors_deanon.ACCOUNT_LIMIT_PARTIAL_EXCEEDED, logCtx)

	})

	t.Run("TestAccountLimitAlertForFullLimitExceeded", func(t *testing.T) {
		AccountLimitCountIncrementForTesting(project.ID, 11)

		mockClient := &MockHTTPClient{}
		// Set up the behavior of the mock
		mockClient.DoFunc = func(req *http.Request) (*http.Response, error) {
			// Return a mock response
			return &http.Response{
				StatusCode: 200,
				Body:       http.NoBody,
			}, nil
		}

		errCode, err := factorsDeanonObj.HandleAccountLimitAlert(project.ID, mockClient, logCtx)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, errCode)

		//Testing by sending the alert again
		errCode, err = factorsDeanonObj.HandleAccountLimitAlert(project.ID, mockClient, logCtx)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusForbidden, errCode)

		DeleteAlertAndAccLimitRedisKeyAfterTesting(project.ID, factors_deanon.ACCOUNT_LIMIT_FULLY_EXCEEDED, logCtx)
	})

	t.Run("TestAccountLimitAlertForFullLimitExceededWhenExecuteFailed", func(t *testing.T) {

		AccountLimitCountIncrementForTesting(project.ID, 11)

		mockClient := &MockHTTPClient{}
		// Set up the behavior of the mock
		mockClient.DoFunc = func(req *http.Request) (*http.Response, error) {
			// Return a mock response
			return &http.Response{
				StatusCode: 400,
				Body:       http.NoBody,
			}, nil
		}

		errCode, err := factorsDeanonObj.HandleAccountLimitAlert(project.ID, mockClient, logCtx)
		assert.NotNil(t, err)
		assert.Equal(t, http.StatusBadRequest, errCode)

		// testing if key is set or not if the execute failed
		alertKey, _ := factors_deanon.GetAccountLimitEmailAlertCacheKey(project.ID, 10, factors_deanon.ACCOUNT_LIMIT_FULLY_EXCEEDED, U.TimeZoneStringIST, logCtx)
		exists, _ := cacheRedis.ExistsPersistent(alertKey)
		assert.Equal(t, false, exists)

		DeleteAlertAndAccLimitRedisKeyAfterTesting(project.ID, factors_deanon.ACCOUNT_LIMIT_FULLY_EXCEEDED, logCtx)

	})

}

func UpdateAccountLimtForTesting(projectId int64, accLimit int64) error {

	_, addOns, err := store.GetStore().GetPlanDetailsAndAddonsForProject(projectId)
	if err != nil {
		return err
	}

	updatedFeatureList := addOns

	if _, exists := updatedFeatureList[model.FEATURE_FACTORS_DEANONYMISATION]; exists {
		feature := model.FeatureDetails{
			Limit:            accLimit,
			IsEnabledFeature: true,
		}
		updatedFeatureList[model.FEATURE_FACTORS_DEANONYMISATION] = feature
	}

	_, err = store.GetStore().UpdateAddonsForProject(projectId, updatedFeatureList)
	if err != nil {
		return err
	}
	return nil
}

func AccountLimitCountIncrementForTesting(projectId int64, count int) {

	i := 0
	for i <= count {
		val := U.RandomString(i + 5)
		err := model.SetFactorsDeanonMonthlyUniqueEnrichmentCount(projectId, val, U.TimeZoneStringIST)
		if err != nil {
			fmt.Println("Error in adding domain to redis key")
		}
		i++
	}
}

func DeleteAlertAndAccLimitRedisKeyAfterTesting(projectId int64, exhaustType string, logCtx *log.Entry) {
	alertKey, _ := factors_deanon.GetAccountLimitEmailAlertCacheKey(projectId, 10, exhaustType, U.TimeZoneStringIST, logCtx)
	limitKey, _ := model.GetFactorsDeanonMonthlyUniqueEnrichmentKey(projectId, U.GetCurrentMonthYear(U.TimeZoneStringIST))
	var keys []*cache.Key
	keys = append(keys, alertKey, limitKey)

	cacheRedis.DelPersistent(keys...)
}

func TestDemandbaseEnrich(t *testing.T) {

	var customerDemandbase demandbase.CustomerDemandbase
	var projectSettings model.ProjectSetting

	projectSettings.ProjectId = U.RandomInt64()
	projectSettings.ClientDemandbaseKey = "0lgoepFUdfbSyIIKdnJSpr2oFeT2nw4TEQFPdaVF"

	userPropertiesMap := make(U.PropertiesMap, 0)
	userId := U.RandomLowerAphaNumString(10)
	clientIP := "89.76.236.199"

	logCtx := log.WithFields(log.Fields{
		"project_id": projectSettings.ProjectId,
		"logId":      fmt.Sprintf("%v+%v", projectSettings.ProjectId, clientIP)})

	_, status := customerDemandbase.Enrich(&projectSettings, &userPropertiesMap, userId, clientIP, logCtx)

	assert.Equal(t, 1, status)

}
