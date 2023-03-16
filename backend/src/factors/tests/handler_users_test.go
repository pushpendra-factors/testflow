package tests

import (
	"encoding/json"
	C "factors/config"
	H "factors/handler"
	"factors/handler/helpers"
	IntHubspot "factors/integration/hubspot"
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
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func sendGetUserReq(r *gin.Engine, projectId int64, agent *model.Agent, offset, limit *int) *httptest.ResponseRecorder {

	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}

	qP := make(map[string]string)
	if offset != nil {
		qP["offset"] = fmt.Sprintf("%d", *offset)
	}
	if limit != nil {
		qP["limit"] = fmt.Sprintf("%d", *limit)
	}

	rb := C.NewRequestBuilderWithPrefix(http.MethodGet, fmt.Sprintf("/projects/%d/users", projectId)).
		WithHeader("Content-UnitType", "application/json").
		WithQueryParams(qP).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})
	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error Creating get users Req")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
func TestAPIGetUsers(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitAppRoutes(r)
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	projectId := project.ID

	// Create 100 Users.
	users := make([]model.User, 0, 0)
	numUsers := 100
	for i := 0; i < numUsers; i++ {
		createdUserID, _ := store.GetStore().CreateUser(&model.User{ProjectId: projectId, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
		user, errCode := store.GetStore().GetUser(projectId, createdUserID)
		assert.Equal(t, http.StatusFound, errCode)
		users = append(users, *user)
	}

	// Default values of offset and limit. Not sent in params.
	offset := 0
	limit := 10
	w := sendGetUserReq(r, projectId, agent, nil, nil)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	retUsers := make([]model.User, 0, 0)
	json.Unmarshal(jsonResponse, &retUsers)
	assert.Equal(t, limit, len(retUsers))
	assertUserMapsWithOffset(t, users[offset:offset+limit], retUsers)

	offset = 25
	limit = 20
	w = sendGetUserReq(r, projectId, agent, &offset, &limit)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &retUsers)
	assert.Equal(t, limit, len(retUsers))
	assertUserMapsWithOffset(t, users[offset:offset+limit], retUsers)

	// Overflow
	offset = 95
	limit = 10
	w = sendGetUserReq(r, projectId, agent, &offset, &limit)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &retUsers)
	assert.Equal(t, numUsers-95, len(retUsers))
	assertUserMapsWithOffset(t, users[offset:numUsers], retUsers)
}

func assertUserMapsWithOffset(t *testing.T, expectedUsers []model.User, actualUsers []model.User) {
	assert.Equal(t, len(expectedUsers), len(actualUsers))
	for i := 0; i < len(actualUsers); i++ {
		expectedUser := expectedUsers[i]
		actualUser := actualUsers[i]
		assert.Equal(t, expectedUser.ID, actualUser.ID)
		assert.Equal(t, expectedUser.ProjectId, actualUser.ProjectId)
		assert.Equal(t, expectedUser.CustomerUserId, actualUser.CustomerUserId)
		// Atleast join_time should be present on user_properites.
		assert.NotEqual(t, postgres.Jsonb{RawMessage: json.RawMessage([]byte(`null`))}, actualUser.Properties)
		assert.NotNil(t, actualUser.CreatedAt)
		assert.NotNil(t, actualUser.UpdatedAt)
	}
}

func buildUserPropertyValuesRequest(projectId int64, propertyName string, label bool, cookieData string) (*http.Request, error) {
	rb := C.NewRequestBuilderWithPrefix(http.MethodGet, fmt.Sprintf("/projects/%d/user_properties/%s/values?label=%t", projectId, propertyName, label)).
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

func sendGetUserPropertyValues(projectId int64, propertyName string, label bool, agent *model.Agent, r *gin.Engine) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}
	req, err := buildUserPropertyValuesRequest(projectId, propertyName, label, cookieData)
	if err != nil {
		log.WithError(err).Error("Error getting event properties.")
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestUserPropertyValuesHandler(t *testing.T) {
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

	jsonContactModel := `{
		"vid": %d,
		"addedAt": %d,
		"properties": {
		  	"firstname": { "value": "%s" },
		  	"lastname": { "value": "%s" },
		  	"lastmodifieddate": { "value": "%d" },
			"hs_analytics_source": { "value": "REFERRALS" }
		},
		"identity-profiles": [
			{
				"vid": %d,
				"identities": [
					{
					  "type": "EMAIL",
					  "value": "%s"
					},
					{
						"type": "LEAD_GUID",
						"value": "%s"
					}
				]
			}
		]
	}`

	documentID := 1
	createdDate := time.Now().Unix()
	updatedTime := createdDate*1000 + 100
	cuid := U.RandomString(10)
	jsonContact := fmt.Sprintf(jsonContactModel, documentID, createdDate*1000, "Sample", "Test", updatedTime, documentID, cuid, "123-456")

	hubspotDocument := model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameContact,
		Value:     &postgres.Jsonb{json.RawMessage(jsonContact)},
	}

	status := store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
	assert.Equal(t, http.StatusCreated, status)

	// execute sync job
	allStatus, _ := IntHubspot.Sync(project.ID, 1, time.Now().Unix(), nil, "", 50)
	for i := range allStatus {
		assert.Equal(t, U.CRM_SYNC_STATUS_SUCCESS, allStatus[i].Status)
	}

	configs = make(map[string]interface{})
	configs["rollupLookback"] = 10
	event_user_cache.DoRollUpSortedSet(configs)

	C.GetConfig().LookbackWindowForEventUserCache = 10

	status = store.GetStore().CreateOrUpdateDisplayNameLabel(project.ID, "hubspot", "$hubspot_contact_hs_analytics_source", "REFERRALS", "Referral")
	assert.Equal(t, http.StatusCreated, status)

	status = store.GetStore().CreateOrUpdateDisplayNameLabel(project.ID, "hubspot", "$hubspot_contact_hs_analytics_source", "DIRECT_TRAFFIC", "Direct Traffic")
	assert.Equal(t, http.StatusCreated, status)

	// Returns []string when label not set
	w := sendGetUserPropertyValues(project.ID, "$hubspot_contact_hs_analytics_source", false, agent, r)
	assert.Equal(t, http.StatusOK, w.Code)

	var propertyValues []string
	jsonResponse, err := ioutil.ReadAll(w.Body)
	assert.Nil(t, err)
	json.Unmarshal(jsonResponse, &propertyValues)
	assert.Equal(t, 1, len(propertyValues))
	assert.Equal(t, "REFERRALS", propertyValues[0])

	// Returns map when label is set
	w = sendGetUserPropertyValues(project.ID, "$hubspot_contact_hs_analytics_source", true, agent, r)
	assert.Equal(t, http.StatusOK, w.Code)

	propertyValueLabelMap := make(map[string]string, 0)
	jsonResponse, err = ioutil.ReadAll(w.Body)
	assert.Nil(t, err)
	json.Unmarshal(jsonResponse, &propertyValueLabelMap)
	assert.Equal(t, 2, len(propertyValueLabelMap))

	assert.Contains(t, propertyValueLabelMap, "REFERRALS")
	assert.Contains(t, propertyValueLabelMap, "DIRECT_TRAFFIC")
	assert.Equal(t, propertyValueLabelMap["REFERRALS"], "Referral")
	assert.Equal(t, propertyValueLabelMap["DIRECT_TRAFFIC"], "Direct Traffic")
}
