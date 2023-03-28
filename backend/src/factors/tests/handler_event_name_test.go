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
	TaskSession "factors/task/session"
	U "factors/util"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	b64 "encoding/base64"
	V1 "factors/handler/v1"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func sendGetEventNamesApproxRequest(projectId int64, agent *model.Agent, r *gin.Engine) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}

	req, err := buildEventNameRequest(projectId, "approx", cookieData)
	if err != nil {
		log.WithError(err).Error("Error getting event names.")
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func sendGetEventNamesExactRequest(projectId int64, agent *model.Agent, r *gin.Engine) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}
	req, err := buildEventNameRequest(projectId, "exact", cookieData)
	if err != nil {
		log.WithError(err).Error("Error getting event names.")
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func sendGetEventNamesExactRequestWithDisplayNames(projectId int64, agent *model.Agent, r *gin.Engine) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}
	req, err := buildEventNameRequestWithDisplayNames(projectId, "true", cookieData)
	if err != nil {
		log.WithError(err).Error("Error getting event names.")
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func buildEventNameRequest(projectId int64, requestType, cookieData string) (*http.Request, error) {
	rb := C.NewRequestBuilderWithPrefix(http.MethodGet, fmt.Sprintf("/projects/%d/event_names?type=%s", projectId, requestType)).
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

func sendGetEventProperties(projectId int64, event string, agent *model.Agent, r *gin.Engine) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}
	req, err := buildEventPropertiesRequest(projectId, event, "true", cookieData)
	if err != nil {
		log.WithError(err).Error("Error getting event properties.")
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func buildEventPropertiesRequest(projectId int64, event string, requestType, cookieData string) (*http.Request, error) {
	eventEncoded := b64.StdEncoding.EncodeToString([]byte(b64.StdEncoding.EncodeToString([]byte(event))))
	rb := C.NewRequestBuilderWithPrefix(http.MethodGet, fmt.Sprintf("/projects/%d/event_names/%s/properties?is_display_name_enabled=true", projectId, eventEncoded)).
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

func sendGetEventNamesByGroupRequest(projectId int64, group string, agent *model.Agent, r *gin.Engine) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}
	req, err := buildEventNamesByGroupRequest(projectId, group, cookieData)
	if err != nil {
		log.WithError(err).Error("Error getting event names")
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w

}

func buildEventNamesByGroupRequest(projectId int64, groupName string, cookieData string) (*http.Request, error) {
	rb := C.NewRequestBuilderWithPrefix(http.MethodGet, fmt.Sprintf("/projects/%d/groups/%s/event_names", projectId, groupName)).
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

func sendGetEventNamesByUserRequest(projectId int64, agent *model.Agent, r *gin.Engine) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}
	req, err := buildEventNamesByUserRequest(projectId, cookieData)
	if err != nil {
		log.WithError(err).Error("Error getting event names")
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w

}

func buildEventNamesByUserRequest(projectId int64, cookieData string) (*http.Request, error) {
	rb := C.NewRequestBuilderWithPrefix(http.MethodGet, fmt.Sprintf("/projects/%d/user/event_names", projectId)).
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

func sendGetUserProperties(projectId int64, agent *model.Agent, r *gin.Engine) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}
	req, err := buildUserPropertiesRequest(projectId, cookieData)
	if err != nil {
		log.WithError(err).Error("Error getting event properties.")
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func buildUserPropertiesRequest(projectId int64, cookieData string) (*http.Request, error) {
	rb := C.NewRequestBuilderWithPrefix(http.MethodGet, fmt.Sprintf("/projects/%d/user_properties?is_display_name_enabled=true", projectId)).
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

func buildEventNameRequestWithDisplayNames(projectId int64, displayNamesEnabled, cookieData string) (*http.Request, error) {
	rb := C.NewRequestBuilderWithPrefix(http.MethodGet, fmt.Sprintf("/projects/%d/v1/event_names?is_display_name_enabled=%s", projectId, displayNamesEnabled)).
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

func sendCreateDisplayNameRequest(r *gin.Engine, request V1.CreateDisplayNamesParams, agent *model.Agent, projectID int64) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}

	rb := C.NewRequestBuilderWithPrefix(http.MethodPost, fmt.Sprintf("/projects/%d/v1/events/displayname", projectID)).
		WithPostParams(request).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error creating display name")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func createEventWithTimestampByName(t *testing.T, project *model.Project, user *model.User, name string, timestamp int64) (*model.EventName, *model.Event) {
	eventName, errCode := store.GetStore().CreateOrGetUserCreatedEventName(&model.EventName{ProjectId: project.ID, Name: name})
	assert.NotNil(t, eventName)
	event, errCode := store.GetStore().CreateEvent(&model.Event{ProjectId: project.ID, EventNameId: eventName.ID, UserId: user.ID, Timestamp: timestamp})
	assert.Equal(t, http.StatusCreated, errCode)
	return eventName, event
}

func TestGetEventNameByGroupHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)
	C.GetConfig().LookbackWindowForEventUserCache = 10
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)

	var eventNames = struct {
		EventNames   map[string][]string `json:"event_names"`
		DisplayNames map[string]string   `json:"display_names"`
	}{}

	t.Run("TestGetEventNameByGroupHandler", func(t *testing.T) {
		w := sendGetEventNamesByGroupRequest(project.ID, model.GROUP_NAME_HUBSPOT_DEAL, agent, r)
		assert.Equal(t, http.StatusOK, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		json.Unmarshal(jsonResponse, &eventNames)
		assert.NotNil(t, eventNames.EventNames)
		assert.Contains(t, eventNames.EventNames[U.STANDARD_GROUP_DISPLAY_NAMES[model.GROUP_NAME_HUBSPOT_DEAL]], "$hubspot_deal_created", "$hubspot_deal_updated")
		assert.True(t, len(eventNames.DisplayNames) > 0)
	})
}

func TestGetEventNameByUserHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)
	C.GetConfig().LookbackWindowForEventUserCache = 10
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)

	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"

	var eventNames H.EventNamesByUserResponsePayload

	group1, status := store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_HUBSPOT_COMPANY, model.AllowedGroupNames)
	assert.Equal(t, http.StatusCreated, status)
	assert.NotNil(t, group1)
	group2, status := store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_SALESFORCE_ACCOUNT, model.AllowedGroupNames)
	assert.NotNil(t, group2)
	assert.Equal(t, http.StatusCreated, status)

	createdUserID, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.NotEmpty(t, createdUserID)
	assert.Equal(t, http.StatusCreated, errCode)

	rEventName := "$hubspot_contact_created"
	w := ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"user_id": "%s",  "event_name": "%s", "auto": true, "event_properties": {"$dollar_property": "dollarValue", "$qp_search": "mobile", "mobile": "true", "$qp_encoded": "google%%20search", "$qp_utm_keyword": "google%%20search"}, "user_properties": {"name": "Jhon"}}`, createdUserID, rEventName)),
		map[string]string{
			"Authorization": project.Token,
			"User-Agent":    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36",
		})

	assert.Equal(t, http.StatusOK, w.Code)

	rEventName = "$hubspot_contact_updated"
	w = ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"user_id": "%s",  "event_name": "%s", "auto": true, "event_properties": {"$dollar_property": "dollarValue", "$qp_search": "mobile", "mobile": "true", "$qp_encoded": "google%%20search", "$qp_utm_keyword": "google%%20search"}, "user_properties": {"name": "Jhon"}}`, createdUserID, rEventName)),
		map[string]string{
			"Authorization": project.Token,
			"User-Agent":    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36",
		})

	assert.Equal(t, http.StatusOK, w.Code)

	rEventName = "$sf_lead_created"
	w = ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"user_id": "%s",  "event_name": "%s", "auto": true, "event_properties": {"$dollar_property": "dollarValue", "$qp_search": "mobile", "mobile": "true", "$qp_encoded": "google%%20search", "$qp_utm_keyword": "google%%20search"}, "user_properties": {"name": "Jhon"}}`, createdUserID, rEventName)),
		map[string]string{
			"Authorization": project.Token,
			"User-Agent":    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36",
		})

	assert.Equal(t, http.StatusOK, w.Code)

	rEventName = "$sf_lead_updated"
	w = ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"user_id": "%s",  "event_name": "%s", "auto": true, "event_properties": {"$dollar_property": "dollarValue", "$qp_search": "mobile", "mobile": "true", "$qp_encoded": "google%%20search", "$qp_utm_keyword": "google%%20search"}, "user_properties": {"name": "Jhon"}}`, createdUserID, rEventName)),
		map[string]string{
			"Authorization": project.Token,
			"User-Agent":    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36",
		})

	assert.Equal(t, http.StatusOK, w.Code)

	rEventName = "$sf_contact_created"
	w = ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"user_id": "%s",  "event_name": "%s", "auto": true, "event_properties": {"$dollar_property": "dollarValue", "$qp_search": "mobile", "mobile": "true", "$qp_encoded": "google%%20search", "$qp_utm_keyword": "google%%20search"}, "user_properties": {"name": "Jhon"}}`, createdUserID, rEventName)),
		map[string]string{
			"Authorization": project.Token,
			"User-Agent":    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36",
		})

	assert.Equal(t, http.StatusOK, w.Code)

	rEventName = "$marketo_lead_created"
	w = ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"user_id": "%s",  "event_name": "%s", "auto": true, "event_properties": {"$dollar_property": "dollarValue", "$qp_search": "mobile", "mobile": "true", "$qp_encoded": "google%%20search", "$qp_utm_keyword": "google%%20search"}, "user_properties": {"name": "Jhon"}}`, createdUserID, rEventName)),
		map[string]string{
			"Authorization": project.Token,
			"User-Agent":    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36",
		})

	assert.Equal(t, http.StatusOK, w.Code)

	rEventName = "$sf_campaign_member_created"
	w = ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"user_id": "%s",  "event_name": "%s", "auto": true, "event_properties": {"$dollar_property": "dollarValue", "$qp_search": "mobile", "mobile": "true", "$qp_encoded": "google%%20search", "$qp_utm_keyword": "google%%20search"}, "user_properties": {"name": "Jhon"}}`, createdUserID, rEventName)),
		map[string]string{
			"Authorization": project.Token,
			"User-Agent":    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36",
		})

	assert.Equal(t, http.StatusOK, w.Code)

	// Create account smart event
	rule := &model.SmartCRMEventFilter{
		Source:               model.SmartCRMEventSourceSalesforce,
		ObjectType:           "account",
		Description:          "salesforce account",
		FilterEvaluationType: model.FilterEvaluationTypeAny,
		Filters: []model.PropertyFilter{
			{
				Name:      "Id",
				Rules:     []model.CRMFilterRule{},
				LogicalOp: model.LOGICAL_OP_AND,
			},
		},
		LogicalOp:               model.LOGICAL_OP_AND,
		TimestampReferenceField: "LastModifiedDate",
	}

	requestPayload := make(map[string]interface{})
	groupAccountSmartEventName := "Account Id set"
	requestPayload["name"] = groupAccountSmartEventName
	requestPayload["expr"] = rule

	eventName, status := store.GetStore().CreateOrGetCRMSmartEventFilterEventName(project.ID, &model.EventName{ProjectId: project.ID, Name: groupAccountSmartEventName}, rule)
	assert.Equal(t, http.StatusCreated, status)
	_, status = store.GetStore().CreateEvent(&model.Event{ProjectId: project.ID, EventNameId: eventName.ID, UserId: createdUserID, Timestamp: U.TimeNowUnix()})
	assert.Equal(t, http.StatusCreated, status)

	t.Run("TestGetEventNameByUserHandler", func(t *testing.T) {
		configs := make(map[string]interface{})
		configs["rollupLookback"] = 1
		event_user_cache.DoRollUpSortedSet(configs)
		w := sendGetEventNamesByUserRequest(project.ID, agent, r)
		assert.Equal(t, http.StatusOK, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		json.Unmarshal(jsonResponse, &eventNames)
		assert.NotNil(t, eventNames.EventNames)
		assert.True(t, len(eventNames.EventNames) > 0)
		assert.Len(t, eventNames.EventNames[U.STANDARD_GROUP_DISPLAY_NAMES[model.GROUP_NAME_HUBSPOT_COMPANY]], 2)
		for _, eventName := range []string{"$hubspot_company_created", "$hubspot_company_updated"} {
			assert.Contains(t, eventNames.EventNames[U.STANDARD_GROUP_DISPLAY_NAMES[model.GROUP_NAME_HUBSPOT_COMPANY]], eventName)
		}

		// account smart event should come under salesforce account group
		assert.Len(t, eventNames.EventNames[U.STANDARD_GROUP_DISPLAY_NAMES[model.GROUP_NAME_SALESFORCE_ACCOUNT]], 3)
		for _, eventName := range []string{"$salesforce_account_created", "$salesforce_account_updated", groupAccountSmartEventName} {
			assert.Contains(t, eventNames.EventNames[U.STANDARD_GROUP_DISPLAY_NAMES[model.GROUP_NAME_SALESFORCE_ACCOUNT]], eventName)
		}

		assert.Equal(t, len(eventNames.EventNames["Salesforce Users"]), 4)
		for _, eventName := range []string{"$sf_lead_created", "$sf_lead_updated", "$sf_contact_created", "$sf_campaign_member_created"} {
			assert.Contains(t, eventNames.EventNames["Salesforce Users"], eventName)
		}

		assert.Equal(t, len(eventNames.EventNames["Hubspot Contacts"]), 2)
		for _, eventName := range []string{"$hubspot_contact_created", "$hubspot_contact_updated"} {
			assert.Contains(t, eventNames.EventNames["Hubspot Contacts"], eventName)
		}

		assert.Equal(t, len(eventNames.EventNames["Marketo Person"]), 1)
		assert.Contains(t, eventNames.EventNames["Marketo Person"], "$marketo_lead_created")

		assert.Nil(t, eventNames.EventNames[U.SmartEvent])

		assert.Equal(t, len(eventNames.AllowedDisplayNameGroups), 5) // STANDARD_GROUP_DISPLAY_NAMES
		for displayGroupName, groupName := range U.GetStandardDisplayNameGroups() {
			assert.Equal(t, eventNames.AllowedDisplayNameGroups[displayGroupName], groupName)
		}

	})

}

func TestGetEventNamesHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)
	var eventNames = struct {
		EventNames []string `json:"event_names"`
		Exact      bool     `json:"exact"`
	}{}
	C.GetConfig().LookbackWindowForEventUserCache = 10

	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"

	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)

	w := sendGetEventNamesExactRequest(project.ID, agent, r)
	assert.Equal(t, http.StatusOK, w.Code) // Should be still 200 for no event_names with empty result set
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &eventNames)
	// should contain all event names.
	assert.Len(t, eventNames.EventNames, 0)

	createdUserID, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.NotEmpty(t, createdUserID)
	assert.Equal(t, http.StatusCreated, errCode)

	rEventName := "event1"
	w = ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"user_id": "%s",  "event_name": "%s", "auto": true, "event_properties": {"$dollar_property": "dollarValue", "$qp_search": "mobile", "mobile": "true", "$qp_encoded": "google%%20search", "$qp_utm_keyword": "google%%20search"}, "user_properties": {"name": "Jhon"}}`, createdUserID, rEventName)),
		map[string]string{
			"Authorization": project.Token,
			"User-Agent":    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36",
		})

	assert.Equal(t, http.StatusOK, w.Code)
	rEventName = "event2"
	w = ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"user_id": "%s",  "event_name": "%s", "auto": true, "event_properties": {"$dollar_property": "dollarValue", "$qp_search": "mobile", "mobile": "true", "$qp_encoded": "google%%20search", "$qp_utm_keyword": "google%%20search"}, "user_properties": {"name": "Jhon"}}`, createdUserID, rEventName)),
		map[string]string{
			"Authorization": project.Token,
			"User-Agent":    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36",
		})
	assert.Equal(t, http.StatusOK, w.Code)

	rEventName = "event1"
	w = ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"user_id": "%s",  "event_name": "%s", "auto": true, "event_properties": {"$dollar_property": "dollarValue", "$qp_search": "mobile", "mobile": "true", "$qp_encoded": "google%%20search", "$qp_utm_keyword": "google%%20search"}, "user_properties": {"name": "Jhon"}}`, createdUserID, rEventName)),
		map[string]string{
			"Authorization": project.Token,
			"User-Agent":    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36",
		})

	assert.Equal(t, http.StatusOK, w.Code)

	rEventName = "$hubspot_contact_created"
	w = ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"user_id": "%s",  "event_name": "%s", "auto": true, "event_properties": {"$dollar_property": "dollarValue", "$qp_search": "mobile", "mobile": "true", "$qp_encoded": "google%%20search", "$qp_utm_keyword": "google%%20search"}, "user_properties": {"name": "Jhon"}}`, createdUserID, rEventName)),
		map[string]string{
			"Authorization": project.Token,
			"User-Agent":    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36",
		})

	assert.Equal(t, http.StatusOK, w.Code)

	_, err = TaskSession.AddSession([]int64{project.ID}, 0, 0, 0, 0, 1, 1)
	assert.Nil(t, err)

	// Test events ingested via sdk/track call
	configs := make(map[string]interface{})
	configs["rollupLookback"] = 1
	event_user_cache.DoRollUpSortedSet(configs)
	w = sendGetEventNamesExactRequest(project.ID, agent, r)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &eventNames)
	// should contain all event names along with $session.
	assert.Len(t, eventNames.EventNames, 4)

	var eventNamesWithDisplayNames = struct {
		EventNames   map[string][]string `json:"event_names"`
		DisplayNames map[string]string   `json:"display_names"`
	}{}
	w = sendGetEventNamesExactRequestWithDisplayNames(project.ID, agent, r)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &eventNamesWithDisplayNames)
	// should contain all event names along with $session.
	assert.Len(t, eventNamesWithDisplayNames.EventNames["Most Recent"], 3)
	assert.Len(t, eventNamesWithDisplayNames.EventNames["Hubspot"], 1)
	assert.Len(t, eventNamesWithDisplayNames.DisplayNames, 38) // STANDARD_EVENTS_DISPLAY_NAMES + event1 + event2
	assert.Equal(t, eventNamesWithDisplayNames.DisplayNames["$session"], "Website Session")

	sendCreateDisplayNameRequest(r, V1.CreateDisplayNamesParams{EventName: "$session", DisplayName: "Test1"}, agent, project.ID)
	sendCreateDisplayNameRequest(r, V1.CreateDisplayNamesParams{EventName: "$hubspot_contact_created", DisplayName: "Test2", PropertyName: ""}, agent, project.ID)
	sendCreateDisplayNameRequest(r, V1.CreateDisplayNamesParams{EventName: "", DisplayName: "Test3", PropertyName: "$joinTime"}, agent, project.ID)
	sendCreateDisplayNameRequest(r, V1.CreateDisplayNamesParams{EventName: "$session", DisplayName: "Test4", PropertyName: "$is_page_view"}, agent, project.ID)
	sendCreateDisplayNameRequest(r, V1.CreateDisplayNamesParams{EventName: "$session", DisplayName: "Test5", PropertyName: "Dummy"}, agent, project.ID)
	sendCreateDisplayNameRequest(r, V1.CreateDisplayNamesParams{EventName: "", DisplayName: "Test6", PropertyName: "Dummy"}, agent, project.ID)
	sendCreateDisplayNameRequest(r, V1.CreateDisplayNamesParams{EventName: "", DisplayName: "Test6-1", PropertyName: "Dummy"}, agent, project.ID)

	status := store.GetStore().CreateOrUpdateDisplayNameByObjectType(project.ID, "$hubspot_contact_createdddate", "Contact", "Created Date", "Hubspot")
	assert.Equal(t, status, 201)
	status = store.GetStore().CreateOrUpdateDisplayNameByObjectType(project.ID, "$hubspot_contact_createdddate1", "Contact", "Created Date", "Hubspot")
	assert.Equal(t, status, 409)
	status = store.GetStore().CreateOrUpdateDisplayNameByObjectType(project.ID, "$hubspot_contact_createdddate", "Contact", "Created Date1", "Hubspot")
	assert.Equal(t, status, 201)
	status = store.GetStore().CreateOrUpdateDisplayNameByObjectType(project.ID, "$hubspot_opportunity_createdddate", "Opportunity", "Created Date1", "Hubspot")
	assert.Equal(t, status, 201)

	w = sendGetEventNamesExactRequestWithDisplayNames(project.ID, agent, r)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &eventNamesWithDisplayNames)
	// should contain all event names along with $session.
	assert.Len(t, eventNamesWithDisplayNames.EventNames["Most Recent"], 3)
	assert.Len(t, eventNamesWithDisplayNames.EventNames["Hubspot"], 1)
	assert.Len(t, eventNamesWithDisplayNames.DisplayNames, 38) // STANDARD_EVENTS_DISPLAY_NAMES + event1 + event2
	assert.Equal(t, eventNamesWithDisplayNames.DisplayNames["$session"], "Test1")
	assert.Equal(t, eventNamesWithDisplayNames.DisplayNames["$hubspot_contact_created"], "Test2")

	var properties = struct {
		Proprties    map[string][]string `json:"properties"`
		DisplayNames map[string]string   `json:"display_names"`
	}{}
	w = sendGetEventProperties(project.ID, "$session", agent, r)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &properties)
	assert.Equal(t, properties.DisplayNames["$is_page_view"], "Test4")
	assert.Equal(t, properties.DisplayNames["Dummy"], "Test5")

	w = sendGetUserProperties(project.ID, agent, r)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &properties)
	assert.Equal(t, properties.DisplayNames["$joinTime"], "Test3")
	assert.Equal(t, properties.DisplayNames["Dummy"], "Test6-1")
	assert.Equal(t, properties.DisplayNames["$hubspot_contact_createdddate"], "Hubspot Contact Created Date1")

	t.Run("DisplayNames title-case check", func(t *testing.T) {
		assert := assert.New(t)
		for event, displayName := range eventNamesWithDisplayNames.DisplayNames {
			if strings.HasPrefix(event, "$") { // Only if event is prefixed with $, displayName is capitalized
				assert.Equal(displayName, strings.Title(displayName))
			} else {
				assert.Equal(displayName, displayName)
			}
		}
		for property, displayName := range properties.DisplayNames {
			if strings.HasPrefix(property, "$") { // Only if property is prefixed with $, displayName is capitalized
				assert.Equal(displayName, strings.Title(displayName))
			} else {
				assert.Equal(displayName, displayName)
			}
		}
	})
}

func TestEnableEventLevelProperties(t *testing.T) {
	// test case with new projectID (-ve test case)
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	r := gin.Default()
	H.InitAppRoutes(r)

	status := IntHubspot.CreateOrGetHubspotEventName(project.ID)
	assert.Equal(t, http.StatusOK, status)

	createdDate := time.Now().Unix()
	eventNameCreated := U.EVENT_NAME_HUBSPOT_CONTACT_CREATED

	eventNameUpdated := U.EVENT_NAME_HUBSPOT_CONTACT_UPDATED
	dtPropertyName1 := "last_visit"
	dtPropertyValue1 := createdDate * 1000
	dtPropertyName2 := "next_visit"
	dtPropertyValue2 := createdDate * 1000

	numPropertyName1 := "vists"
	numPropertyValue1 := 15
	numPropertyName2 := "views"
	numPropertyValue2 := 10

	// datetime property
	dtEnKey1 := model.GetCRMEnrichPropertyKeyByType(
		model.SmartCRMEventSourceHubspot,
		model.HubspotDocumentTypeNameContact,
		U.GetPropertyValueAsString(dtPropertyName1),
	)
	dtEnKey2 := model.GetCRMEnrichPropertyKeyByType(
		model.SmartCRMEventSourceHubspot,
		model.HubspotDocumentTypeNameContact,
		U.GetPropertyValueAsString(dtPropertyName2),
	)

	// numerical property
	numEnKey1 := model.GetCRMEnrichPropertyKeyByType(
		model.SmartCRMEventSourceHubspot,
		model.HubspotDocumentTypeNameContact,
		U.GetPropertyValueAsString(numPropertyName1),
	)
	numEnKey2 := model.GetCRMEnrichPropertyKeyByType(
		model.SmartCRMEventSourceHubspot,
		model.HubspotDocumentTypeNameContact,
		U.GetPropertyValueAsString(numPropertyName2),
	)

	status = store.GetStore().CreatePropertyDetails(project.ID, eventNameCreated, dtEnKey1, U.PropertyTypeDateTime, false, false)
	assert.Equal(t, http.StatusCreated, status)
	status = store.GetStore().CreatePropertyDetails(project.ID, eventNameCreated, dtEnKey2, U.PropertyTypeDateTime, false, false)
	assert.Equal(t, http.StatusCreated, status)

	status = store.GetStore().CreatePropertyDetails(project.ID, "", dtEnKey1, U.PropertyTypeDateTime, true, false)
	assert.Equal(t, http.StatusCreated, status)
	status = store.GetStore().CreatePropertyDetails(project.ID, "", dtEnKey2, U.PropertyTypeDateTime, true, false)
	assert.Equal(t, http.StatusCreated, status)

	status = store.GetStore().CreatePropertyDetails(project.ID, eventNameUpdated, dtEnKey1, U.PropertyTypeDateTime, false, false)
	assert.Equal(t, http.StatusCreated, status)
	status = store.GetStore().CreatePropertyDetails(project.ID, eventNameUpdated, dtEnKey2, U.PropertyTypeDateTime, false, false)
	assert.Equal(t, http.StatusCreated, status)

	status = store.GetStore().CreatePropertyDetails(project.ID, eventNameCreated, numEnKey1, U.PropertyTypeNumerical, false, false)
	assert.Equal(t, http.StatusCreated, status)
	status = store.GetStore().CreatePropertyDetails(project.ID, eventNameCreated, numEnKey2, U.PropertyTypeNumerical, false, false)
	assert.Equal(t, http.StatusCreated, status)

	status = store.GetStore().CreatePropertyDetails(project.ID, "", numEnKey1, U.PropertyTypeNumerical, true, false)
	assert.Equal(t, http.StatusCreated, status)
	status = store.GetStore().CreatePropertyDetails(project.ID, "", numEnKey2, U.PropertyTypeNumerical, true, false)
	assert.Equal(t, http.StatusCreated, status)

	status = store.GetStore().CreatePropertyDetails(project.ID, eventNameUpdated, numEnKey1, U.PropertyTypeNumerical, false, false)
	assert.Equal(t, http.StatusCreated, status)
	status = store.GetStore().CreatePropertyDetails(project.ID, eventNameUpdated, numEnKey2, U.PropertyTypeNumerical, false, false)
	assert.Equal(t, http.StatusCreated, status)

	// create new hubspot document
	jsonContactModel := `{
		"vid": %d,
		"addedAt": %d,
		"properties": {
		"createdate": { "value": "%d" },
		  "lastmodifieddate": { "value": "%d" },
		  "lifecyclestage": { "value": "%s" },
		  "%s":{"value":"%d"},
		  "%s":{"value":"%d"},
		  "%s":{"value":"%d"},
		  "%s":{"value":"%d"}
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

	documentID := 2
	cuid := U.RandomLowerAphaNumString(5)
	updatedTime := createdDate*1000 + 100
	jsonContact := fmt.Sprintf(jsonContactModel, documentID, createdDate*1000, createdDate*1000, updatedTime, "lead", dtPropertyName1, dtPropertyValue1, dtPropertyName2, dtPropertyValue2, numPropertyName1, numPropertyValue1, numPropertyName2, numPropertyValue2, documentID, cuid, "123-45")
	contactPJson := postgres.Jsonb{json.RawMessage(jsonContact)}

	hubspotDocument := model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameContact,
		Value:     &contactPJson,
	}

	status = store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
	assert.Equal(t, http.StatusCreated, status)
	assert.Equal(t, createdDate*1000, hubspotDocument.Timestamp)

	// execute sync job
	allStatus, _ := IntHubspot.Sync(project.ID, 3, time.Now().Unix(), nil, "", 50)
	for i := range allStatus {
		assert.Equal(t, U.CRM_SYNC_STATUS_SUCCESS, allStatus[i].Status)
	}

	// execute DoRollUpSortedSet
	configs := make(map[string]interface{})
	configs["rollupLookback"] = 1
	event_user_cache.DoRollUpSortedSet(configs)
	eventEncoded := b64.StdEncoding.EncodeToString([]byte(b64.StdEncoding.EncodeToString([]byte(eventNameCreated))))
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	assert.Equal(t, err, nil)

	// invoke event name handler
	var propertyValues map[string][]string
	rb := C.NewRequestBuilderWithPrefix(http.MethodGet, fmt.Sprintf("/projects/%d/event_names/%s/properties", project.ID, eventEncoded)).
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
	err = json.Unmarshal(jsonResponse, &propertyValues)
	assert.Nil(t, err)

	// compare the returned properties
	assert.NotContains(t, propertyValues[U.PropertyTypeDateTime], dtEnKey1, dtEnKey2)
	assert.NotContains(t, propertyValues[U.PropertyTypeNumerical], numEnKey1, numEnKey2)

	// test case for which event level properties are enabled (+ve test case)
	project, agent, err = SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	C.SetEnableEventLevelEventProperties(project.ID)

	status = IntHubspot.CreateOrGetHubspotEventName(project.ID)
	assert.Equal(t, http.StatusOK, status)

	status = store.GetStore().CreatePropertyDetails(project.ID, eventNameCreated, dtEnKey1, U.PropertyTypeDateTime, false, false)
	assert.Equal(t, http.StatusCreated, status)
	status = store.GetStore().CreatePropertyDetails(project.ID, eventNameCreated, dtEnKey2, U.PropertyTypeDateTime, false, false)
	assert.Equal(t, http.StatusCreated, status)

	status = store.GetStore().CreatePropertyDetails(project.ID, "", dtEnKey1, U.PropertyTypeDateTime, true, false)
	assert.Equal(t, http.StatusCreated, status)
	status = store.GetStore().CreatePropertyDetails(project.ID, "", dtEnKey2, U.PropertyTypeDateTime, true, false)
	assert.Equal(t, http.StatusCreated, status)

	status = store.GetStore().CreatePropertyDetails(project.ID, eventNameUpdated, dtEnKey1, U.PropertyTypeDateTime, false, false)
	assert.Equal(t, http.StatusCreated, status)
	status = store.GetStore().CreatePropertyDetails(project.ID, eventNameUpdated, dtEnKey2, U.PropertyTypeDateTime, false, false)
	assert.Equal(t, http.StatusCreated, status)

	status = store.GetStore().CreatePropertyDetails(project.ID, eventNameCreated, numEnKey1, U.PropertyTypeNumerical, false, false)
	assert.Equal(t, http.StatusCreated, status)
	status = store.GetStore().CreatePropertyDetails(project.ID, eventNameCreated, numEnKey2, U.PropertyTypeNumerical, false, false)
	assert.Equal(t, http.StatusCreated, status)

	status = store.GetStore().CreatePropertyDetails(project.ID, "", numEnKey1, U.PropertyTypeNumerical, true, false)
	assert.Equal(t, http.StatusCreated, status)
	status = store.GetStore().CreatePropertyDetails(project.ID, "", numEnKey2, U.PropertyTypeNumerical, true, false)
	assert.Equal(t, http.StatusCreated, status)

	status = store.GetStore().CreatePropertyDetails(project.ID, eventNameUpdated, numEnKey1, U.PropertyTypeNumerical, false, false)
	assert.Equal(t, http.StatusCreated, status)
	status = store.GetStore().CreatePropertyDetails(project.ID, eventNameUpdated, numEnKey2, U.PropertyTypeNumerical, false, false)
	assert.Equal(t, http.StatusCreated, status)

	// create new hubspot document
	jsonContactModel = `{
		"vid": %d,
		"addedAt": %d,
		"properties": {
		"createdate": { "value": "%d" },
		  "lastmodifieddate": { "value": "%d" },
		  "lifecyclestage": { "value": "%s" },
		  "%s":{"value":"%d"},
		  "%s":{"value":"%d"},
		  "%s":{"value":"%d"},
		  "%s":{"value":"%d"}
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

	documentID = 2
	cuid = U.RandomLowerAphaNumString(5)
	updatedTime = createdDate*1000 + 100
	jsonContact = fmt.Sprintf(jsonContactModel, documentID, createdDate*1000, createdDate*1000, updatedTime, "lead", dtPropertyName1, dtPropertyValue1, dtPropertyName2, dtPropertyValue2, numPropertyName1, numPropertyValue1, numPropertyName2, numPropertyValue2, documentID, cuid, "123-45")
	contactPJson = postgres.Jsonb{json.RawMessage(jsonContact)}

	hubspotDocument = model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameContact,
		Value:     &contactPJson,
	}

	status = store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
	assert.Equal(t, http.StatusCreated, status)

	// execute sync job
	allStatus, _ = IntHubspot.Sync(project.ID, 3, time.Now().Unix(), nil, "", 50)
	for i := range allStatus {
		assert.Equal(t, U.CRM_SYNC_STATUS_SUCCESS, allStatus[i].Status)
	}

	// execute DoRollUpSortedSet
	configs = make(map[string]interface{})
	configs["rollupLookback"] = 1
	event_user_cache.DoRollUpSortedSet(configs)
	eventEncoded = b64.StdEncoding.EncodeToString([]byte(b64.StdEncoding.EncodeToString([]byte(eventNameCreated))))
	cookieData, err = helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	assert.Equal(t, err, nil)

	// invoke event name handler
	rb = C.NewRequestBuilderWithPrefix(http.MethodGet, fmt.Sprintf("/projects/%d/event_names/%s/properties", project.ID, eventEncoded)).
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

	// compare the returned properties
	assert.Contains(t, propertyValues[U.PropertyTypeDateTime], dtEnKey1, dtEnKey2)
	assert.Contains(t, propertyValues[U.PropertyTypeNumerical], numEnKey1, numEnKey2)

}

func TestDisabledEventUserProperties(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	r := gin.Default()
	H.InitAppRoutes(r)

	var properties = struct {
		DisplayNames                map[string]string `json:"display_names"`
		DisabledEventUserProperties []string          `json:"disabled_event_user_properties"`
	}{}

	w := sendGetUserProperties(project.ID, agent, r)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &properties)

	// event-level properties disabled
	assert.Len(t, properties.DisabledEventUserProperties, 3)
	assert.Equal(t, properties.DisabledEventUserProperties, U.DISABLED_EVENT_USER_PROPERTIES)

	// disabled only on event-level user_properties dropdown.
	// enabled on global user properties dropdown
	assert.Equal(t, "Initial Channel", properties.DisplayNames[U.UP_INITIAL_CHANNEL])
	assert.Equal(t, "Latest Channel", properties.DisplayNames[U.UP_LATEST_CHANNEL])

	// disabled on event-level and global user_properties dropdown.
	assert.Equal(t, "", properties.DisplayNames[U.UP_SESSION_COUNT])
}

func buildEventPropertyValuesRequest(projectId int64, eventName, propertyName string, label bool, cookieData string) (*http.Request, error) {
	eventNameEncoded := b64.StdEncoding.EncodeToString([]byte(b64.StdEncoding.EncodeToString([]byte(eventName))))
	rb := C.NewRequestBuilderWithPrefix(http.MethodGet, fmt.Sprintf("/projects/%d/event_names/%s/properties/%s/values?label=%t", projectId, eventNameEncoded, propertyName, label)).
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

func sendGetEventPropertyValues(projectId int64, eventName string, propertyName string, label bool, agent *model.Agent, r *gin.Engine) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}
	req, err := buildEventPropertyValuesRequest(projectId, eventName, propertyName, label, cookieData)
	if err != nil {
		log.WithError(err).Error("Error getting event properties.")
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestEventPropertyValuesHandler(t *testing.T) {
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

	// create new hubspot document
	jsonContactModel := `{
		"vid": %d,
		"addedAt": %d,
		"properties": {
		  	"firstname": { "value": "%s" },
		  	"lastname": { "value": "%s" },
		  	"lastmodifieddate": { "value": "%d" },
			"company_risk_profile": { "value": "%s" }
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
	jsonContact := fmt.Sprintf(jsonContactModel, documentID, createdDate*1000, "Sample", "Test", updatedTime, "blocked", documentID, cuid, "123-456")

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

	status = store.GetStore().CreateOrUpdateDisplayNameLabel(project.ID, "hubspot", "$hubspot_contact_company_risk_profile", "blocked", "Blocked")
	assert.Equal(t, http.StatusCreated, status)

	status = store.GetStore().CreateOrUpdateDisplayNameLabel(project.ID, "hubspot", "$hubspot_contact_company_risk_profile", "safe", "Safe")
	assert.Equal(t, http.StatusCreated, status)

	status = store.GetStore().CreateOrUpdateDisplayNameLabel(project.ID, "hubspot", "$hubspot_contact_company_risk_profile", "low_risk", "Low Risk")
	assert.Equal(t, http.StatusCreated, status)

	status = store.GetStore().CreateOrUpdateDisplayNameLabel(project.ID, "hubspot", "$hubspot_contact_company_risk_profile", "medium_risk", "Medium Risk")
	assert.Equal(t, http.StatusCreated, status)

	status = store.GetStore().CreateOrUpdateDisplayNameLabel(project.ID, "hubspot", "$hubspot_contact_company_risk_profile", "high_risk", "high Risk")
	assert.Equal(t, http.StatusCreated, status)

	// Returns []string when label not set
	w := sendGetEventPropertyValues(project.ID, U.EVENT_NAME_HUBSPOT_CONTACT_CREATED, "$hubspot_contact_company_risk_profile", false, agent, r)
	assert.Equal(t, http.StatusOK, w.Code)

	var propertyValues []string
	jsonResponse, err := ioutil.ReadAll(w.Body)
	assert.Nil(t, err)
	json.Unmarshal(jsonResponse, &propertyValues)
	assert.Equal(t, 1, len(propertyValues))
	assert.Contains(t, propertyValues, "blocked")
	assert.Equal(t, "blocked", propertyValues[0])

	// Returns map when label is set
	w = sendGetEventPropertyValues(project.ID, U.EVENT_NAME_HUBSPOT_CONTACT_CREATED, "$hubspot_contact_company_risk_profile", true, agent, r)
	assert.Equal(t, http.StatusOK, w.Code)

	propertyValueLabelMap := make(map[string]string, 0)
	jsonResponse, err = ioutil.ReadAll(w.Body)
	assert.Nil(t, err)
	json.Unmarshal(jsonResponse, &propertyValueLabelMap)
	assert.Equal(t, 5, len(propertyValueLabelMap))

	assert.Contains(t, propertyValueLabelMap, "blocked")
	assert.Contains(t, propertyValueLabelMap, "safe")
	assert.Contains(t, propertyValueLabelMap, "low_risk")
	assert.Contains(t, propertyValueLabelMap, "medium_risk")
	assert.Contains(t, propertyValueLabelMap, "high_risk")
	assert.Equal(t, propertyValueLabelMap["blocked"], "Blocked")
	assert.Equal(t, propertyValueLabelMap["safe"], "Safe")
	assert.Equal(t, propertyValueLabelMap["low_risk"], "Low Risk")
	assert.Equal(t, propertyValueLabelMap["medium_risk"], "Medium Risk")
	assert.Equal(t, propertyValueLabelMap["high_risk"], "high Risk")
}
