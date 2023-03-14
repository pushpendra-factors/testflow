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
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func sendGetGroupsRequest(projectId int64, isAccount string, agent *model.Agent, r *gin.Engine) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}
	rb := C.NewRequestBuilderWithPrefix(http.MethodGet, fmt.Sprintf("/projects/%d/groups?is_account=%s", projectId, isAccount)).
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

	isAccountCaseList := []string{"true", "false", ""}

	// No Groups
	for _, isAccount := range isAccountCaseList {
		w := sendGetGroupsRequest(project.ID, isAccount, agent, r)
		assert.Equal(t, http.StatusFound, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		groupsList := make([]model.Group, 0)
		json.Unmarshal(jsonResponse, &groupsList)
		assert.Equal(t, 0, len(groupsList))
	}

	// 2 Groups - 1 isAccount
	group1, status := store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_HUBSPOT_COMPANY, model.AllowedGroupNames)
	assert.Equal(t, http.StatusCreated, status)
	assert.NotNil(t, group1)
	group2, status := store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_HUBSPOT_DEAL, model.AllowedGroupNames)
	assert.Equal(t, http.StatusCreated, status)
	assert.NotNil(t, group2)
	for _, isAccount := range isAccountCaseList {
		w := sendGetGroupsRequest(project.ID, isAccount, agent, r)
		assert.Equal(t, http.StatusFound, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		groupsList := make([]model.Group, 0)
		json.Unmarshal(jsonResponse, &groupsList)
		NoOfGroups := 2
		if isAccount == "true" || isAccount == "false" {
			NoOfGroups = 1
		}
		assert.Equal(t, NoOfGroups, len(groupsList))
	}

	// +3 Groups - +2 isAccount
	group3, status := store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_SALESFORCE_ACCOUNT, model.AllowedGroupNames)
	assert.Equal(t, http.StatusCreated, status)
	assert.NotNil(t, group3)
	group4, status := store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_SALESFORCE_OPPORTUNITY, model.AllowedGroupNames)
	assert.Equal(t, http.StatusCreated, status)
	assert.NotNil(t, group4)
	group5, status := store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_SIX_SIGNAL, model.AllowedGroupNames)
	assert.Equal(t, http.StatusCreated, status)
	assert.NotNil(t, group5)

	for _, isAccount := range isAccountCaseList {
		w := sendGetGroupsRequest(project.ID, isAccount, agent, r)
		assert.Equal(t, http.StatusFound, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		groupsList := make([]model.Group, 0)
		json.Unmarshal(jsonResponse, &groupsList)
		NoOfGroups := 5
		if isAccount == "true" {
			NoOfGroups = 3
		}
		if isAccount == "false" {
			NoOfGroups = 2
		}
		assert.Equal(t, NoOfGroups, len(groupsList))
	}

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
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	err = json.Unmarshal(jsonResponse, &properties)
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
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	err = json.Unmarshal(jsonResponse, &propertyValues)
	assert.Nil(t, err)
	assert.Contains(t, propertyValues, "value1")
	assert.Contains(t, propertyValues, "existingPropertyNewValue1")
}
