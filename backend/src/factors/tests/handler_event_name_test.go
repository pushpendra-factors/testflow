package tests

import (
	"encoding/json"
	C "factors/config"
	H "factors/handler"
	"factors/handler/helpers"
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

		assert.Equal(t, len(eventNames.AllowedDisplayNameGroups), 7) // STANDARD_GROUP_DISPLAY_NAMES
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
	assert.Len(t, eventNamesWithDisplayNames.DisplayNames, 47) // STANDARD_EVENTS_DISPLAY_NAMES + event1 + event2
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
	assert.Len(t, eventNamesWithDisplayNames.DisplayNames, 47) // STANDARD_EVENTS_DISPLAY_NAMES + event1 + event2
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
	assert.Len(t, properties.DisabledEventUserProperties, 5)
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
