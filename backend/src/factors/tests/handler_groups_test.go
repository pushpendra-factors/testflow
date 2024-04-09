package tests

import (
	b64 "encoding/base64"
	"encoding/json"
	C "factors/config"
	H "factors/handler"
	"factors/handler/helpers"
	"factors/model/model"
	"factors/model/store"
	"factors/task/event_user_cache"
	U "factors/util"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func sendGetGroupsRequest(projectId int64, agent *model.Agent, r *gin.Engine) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}
	rb := C.NewRequestBuilderWithPrefix(http.MethodGet, fmt.Sprintf("/projects/%d/groups", projectId)).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})
	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error Creating GetGroups Request")
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w

}

func TestAPIGroupsHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)
	assert.NotNil(t, agent)

	// No Groups
	w := sendGetGroupsRequest(project.ID, agent, r)
	assert.Equal(t, http.StatusFound, w.Code)
	jsonResponse, _ := io.ReadAll(w.Body)
	groupsList := []model.GroupName{}
	json.Unmarshal(jsonResponse, &groupsList)
	assert.Equal(t, 0, len(groupsList))

	// 2 Groups
	group1, status := store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_HUBSPOT_COMPANY, model.AllowedGroupNames)
	assert.Equal(t, http.StatusCreated, status)
	assert.NotNil(t, group1)
	group2, status := store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_HUBSPOT_DEAL, model.AllowedGroupNames)
	assert.Equal(t, http.StatusCreated, status)
	assert.NotNil(t, group2)
	w = sendGetGroupsRequest(project.ID, agent, r)
	assert.Equal(t, http.StatusFound, w.Code)
	jsonResponse, _ = io.ReadAll(w.Body)
	groupsList = []model.GroupName{}
	json.Unmarshal(jsonResponse, &groupsList)
	assert.Equal(t, 2, len(groupsList))

	// +3 Groups
	group3, status := store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_SALESFORCE_ACCOUNT, model.AllowedGroupNames)
	assert.Equal(t, http.StatusCreated, status)
	assert.NotNil(t, group3)
	group4, status := store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_SALESFORCE_OPPORTUNITY, model.AllowedGroupNames)
	assert.Equal(t, http.StatusCreated, status)
	assert.NotNil(t, group4)
	group5, status := store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_SIX_SIGNAL, model.AllowedGroupNames)
	assert.Equal(t, http.StatusCreated, status)
	assert.NotNil(t, group5)

	w = sendGetGroupsRequest(project.ID, agent, r)
	assert.Equal(t, http.StatusFound, w.Code)
	jsonResponse, _ = io.ReadAll(w.Body)
	groupsList = []model.GroupName{}
	json.Unmarshal(jsonResponse, &groupsList)
	assert.Equal(t, 5, len(groupsList))

	segments, status := store.GetStore().GetAllSegments(project.ID)
	assert.Equal(t, status, http.StatusFound)
	assert.NotNil(t, segments)
	assert.NotNil(t, segments[U.GROUP_NAME_DOMAINS])
	assert.Equal(t, len(segments[U.GROUP_NAME_DOMAINS]), 3)
}

func TestAPIGroupPropertiesAndValuesHandler(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	r := gin.Default()
	H.InitAppRoutes(r)

	allowedGroups := []string{
		"g1",
		"g2",
	}
	allowedGroupsMap := map[string]bool{
		"g1": true,
		"g2": true,
	}

	index := 1
	for _, groupName := range allowedGroups {
		group, status := store.GetStore().CreateGroup(project.ID, groupName, allowedGroupsMap)
		assert.Equal(t, http.StatusCreated, status, fmt.Sprintf("failed creating group %s", groupName))
		assert.NotNil(t, group)
		assert.Equal(t, index, group.ID)
		index++
	}

	userIDs := make([]string, len(allowedGroups))
	groupIDs := []string{"1", "2"}

	// Create group with properties.
	var properties1 map[string]interface{} = map[string]interface{}{"property1": "value1", "property2": "value2"}
	for i := range allowedGroups {
		userID, err := store.GetStore().CreateOrUpdateGroupPropertiesBySource(project.ID, allowedGroups[i], groupIDs[i], "",
			&properties1, U.TimeNowUnix(), U.TimeNowUnix(), model.SmartCRMEventSourceHubspot)
		assert.Nil(t, err)
		assert.NotEqual(t, "", userID)
		userIDs[i] = userID
	}

	// Update group properties.
	var properties2 map[string]interface{} = map[string]interface{}{"property1": "existingPropertyNewValue1", "property3": "value3"}
	userIDExisting, err := store.GetStore().CreateOrUpdateGroupPropertiesBySource(project.ID, "g1", "1", userIDs[0],
		&properties2, U.TimeNowUnix(), U.TimeNowUnix(), model.SmartCRMEventSourceHubspot)
	assert.Equal(t, userIDExisting, userIDs[0])

	configs := make(map[string]interface{})
	configs["rollupLookback"] = 1
	event_user_cache.DoRollUpSortedSet(configs)
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	assert.Equal(t, err, nil)

	C.GetConfig().LookbackWindowForEventUserCache = 1

	// Test groups properties handler.
	groupNameEncoded := b64.StdEncoding.EncodeToString([]byte(b64.StdEncoding.EncodeToString([]byte("g1"))))
	var properties map[string]map[string][]string
	rb := C.NewRequestBuilderWithPrefix(http.MethodGet, fmt.Sprintf("/projects/%d/groups/%s/properties", project.ID, groupNameEncoded)).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})
	req, err := rb.Build()
	assert.Equal(t, err, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ := io.ReadAll(w.Body)
	err = json.Unmarshal(jsonResponse, &properties)
	assert.Equal(t, err, nil)
	assert.Contains(t, properties, "properties")
	assert.Contains(t, properties["properties"], "categorical")
	assert.Contains(t, properties["properties"]["categorical"], "property1")
	assert.Contains(t, properties["properties"]["categorical"], "property2")
	assert.Contains(t, properties["properties"]["categorical"], "property3")

	// Test groups properties values handler.
	var propertyValues []string
	rb = C.NewRequestBuilderWithPrefix(http.MethodGet, fmt.Sprintf(
		"/projects/%d/groups/%s/properties/%s/values",
		project.ID, groupNameEncoded, "property1")).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})
	req, err = rb.Build()
	assert.Equal(t, err, nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = io.ReadAll(w.Body)
	err = json.Unmarshal(jsonResponse, &propertyValues)
	assert.Nil(t, err)
	assert.Contains(t, propertyValues, "value1")
	assert.Contains(t, propertyValues, "existingPropertyNewValue1")

	//Test $domains group properties
	domainGroup, status := store.GetStore().CreateOrGetDomainsGroup(project.ID)
	assert.Equal(t, http.StatusCreated, status)
	assert.NotNil(t, domainGroup)

	groupNameEncoded = b64.StdEncoding.EncodeToString([]byte(b64.StdEncoding.EncodeToString([]byte(U.GROUP_NAME_DOMAINS))))
	rb = C.NewRequestBuilderWithPrefix(http.MethodGet, fmt.Sprintf("/projects/%d/groups/%s/properties", project.ID, groupNameEncoded)).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})
	req, err = rb.Build()
	assert.Equal(t, err, nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = io.ReadAll(w.Body)
	err = json.Unmarshal(jsonResponse, &properties)
	assert.Nil(t, err)
	assert.Equal(t, U.ALL_ACCOUNT_DEFAULT_PROPERTIES, properties["properties"]["categorical"])
}

func buildGroupPropertyValuesRequest(projectId int64, groupName, propertyName string, label bool, cookieData string) (*http.Request, error) {
	groupNameEncoded := b64.StdEncoding.EncodeToString([]byte(b64.StdEncoding.EncodeToString([]byte(groupName))))
	rb := C.NewRequestBuilderWithPrefix(http.MethodGet, fmt.Sprintf("/projects/%d/groups/%s/properties/%s/values?label=%t", projectId, groupNameEncoded, propertyName, label)).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})
	req, err := rb.Build()
	if err != nil {
		return nil, err
	}
	return req, nil
}

func sendGetGroupPropertyValues(projectId int64, groupName string, propertyName string, label bool, agent *model.Agent, r *gin.Engine) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}
	req, err := buildGroupPropertyValuesRequest(projectId, groupName, propertyName, label, cookieData)
	if err != nil {
		log.WithError(err).Error("Error getting groupName property values.")
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func sendGetGroupProperties(projectID int64, groupName string, agent *model.Agent, r *gin.Engine) (*httptest.ResponseRecorder, error) {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}

	groupNameEncoded := b64.StdEncoding.EncodeToString([]byte(b64.StdEncoding.EncodeToString([]byte(groupName))))
	rb := C.NewRequestBuilderWithPrefix(http.MethodGet, fmt.Sprintf("/projects/%d/groups/%s/properties", projectID, groupNameEncoded)).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})
	req, err := rb.Build()
	if err != nil {
		return nil, err
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w, nil
}

func TestGroupPropertyValuesHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	configs := make(map[string]interface{})
	configs["eventsLimit"] = 10
	configs["propertiesLimit"] = 10
	configs["valuesLimit"] = 10
	event_user_cache.DoCleanUpSortedSet(configs)

	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)

	group, status := store.GetStore().CreateGroup(project.ID, "g1", map[string]bool{"g1": true})
	assert.Equal(t, http.StatusCreated, status)
	assert.NotNil(t, group)
	assert.Equal(t, 1, group.ID)

	var properties map[string]interface{} = map[string]interface{}{"$hubspot_deal_dealtype": "newbusiness", "$hubspot_company_hubspot_owner_id": "66"}
	userID, err := store.GetStore().CreateOrUpdateGroupPropertiesBySource(project.ID, "g1", "1", "",
		&properties, U.TimeNowUnix(), U.TimeNowUnix(), model.SmartCRMEventSourceHubspot)
	assert.Nil(t, err)
	assert.NotEqual(t, "", userID)

	configs = make(map[string]interface{})
	configs["rollupLookback"] = 10
	event_user_cache.DoRollUpSortedSet(configs)

	C.GetConfig().LookbackWindowForEventUserCache = 10

	status = store.GetStore().CreateOrUpdateDisplayNameLabel(project.ID, "hubspot", "$hubspot_deal_dealtype", "newbusiness", "New Business")
	assert.Equal(t, http.StatusCreated, status)

	status = store.GetStore().CreateOrUpdateDisplayNameLabel(project.ID, "hubspot", "$hubspot_deal_dealtype", "existingbusiness", "ExistingBusiness")
	assert.Equal(t, http.StatusCreated, status)

	status = store.GetStore().CreateOrUpdateDisplayNameLabel(project.ID, "hubspot", "$hubspot_company_hubspot_owner_id", "66", "Blog Api Test")
	assert.Equal(t, http.StatusCreated, status)

	// Returns []string when label not set
	w := sendGetGroupPropertyValues(project.ID, "g1", "$hubspot_deal_dealtype", false, agent, r)
	assert.Equal(t, http.StatusOK, w.Code)

	var propertyValues []string
	jsonResponse, err := io.ReadAll(w.Body)
	assert.Nil(t, err)
	json.Unmarshal(jsonResponse, &propertyValues)
	assert.Equal(t, 1, len(propertyValues))
	assert.Contains(t, propertyValues, "newbusiness")
	assert.Equal(t, "newbusiness", propertyValues[0])

	w = sendGetGroupPropertyValues(project.ID, "g1", "$hubspot_company_hubspot_owner_id", false, agent, r)
	assert.Equal(t, http.StatusOK, w.Code)

	jsonResponse, err = io.ReadAll(w.Body)
	assert.Nil(t, err)
	json.Unmarshal(jsonResponse, &propertyValues)
	assert.Equal(t, 1, len(propertyValues))
	assert.Contains(t, propertyValues, "66")
	assert.Equal(t, "66", propertyValues[0])

	// Returns map when label is set
	w = sendGetGroupPropertyValues(project.ID, "g1", "$hubspot_deal_dealtype", true, agent, r)
	assert.Equal(t, http.StatusOK, w.Code)

	propertyValueLabelMap := make(map[string]string, 0)
	jsonResponse, err = io.ReadAll(w.Body)
	assert.Nil(t, err)
	json.Unmarshal(jsonResponse, &propertyValueLabelMap)
	assert.Equal(t, 2, len(propertyValueLabelMap))

	assert.Contains(t, propertyValueLabelMap, "newbusiness")
	assert.Contains(t, propertyValueLabelMap, "existingbusiness")
	assert.Equal(t, propertyValueLabelMap["newbusiness"], "New Business")
	assert.Equal(t, propertyValueLabelMap["existingbusiness"], "ExistingBusiness")

	w = sendGetGroupPropertyValues(project.ID, "g1", "$hubspot_company_hubspot_owner_id", true, agent, r)
	assert.Equal(t, http.StatusOK, w.Code)

	propertyValueLabelMap = make(map[string]string, 0)
	jsonResponse, err = io.ReadAll(w.Body)
	assert.Nil(t, err)
	json.Unmarshal(jsonResponse, &propertyValueLabelMap)
	assert.Equal(t, 1, len(propertyValueLabelMap))

	assert.Contains(t, propertyValueLabelMap, "66")
	assert.Equal(t, propertyValueLabelMap["66"], "Blog Api Test")
}
