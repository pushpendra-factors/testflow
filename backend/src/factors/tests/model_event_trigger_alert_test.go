package tests

import (
	"encoding/json"
	C "factors/config"
	H "factors/handler"
	"factors/handler/helpers"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
)

func TestCreateEventTriggerAlert(t *testing.T) {
	project, agent, err := SetupProjectWithSlackIntegratedAgentDAO()
	assert.Nil(t, err)

	slackChannel := model.SlackChannel{
		Name:      "channel1",
		Id:        U.GetUUID(),
		IsPrivate: true,
	}
	slackChannelJson, err := json.Marshal(slackChannel)
	assert.Nil(t, err)

	breakdownProps, err := json.Marshal(`[
        {
            "pr": "user_id",
            "en": "user",
            "pty": "categorical",
            "ena": "$hubspot_contact_created",
            "eni": 1
        }
    ]`)
	assert.Nil(t, err)

	t.Run("CreateEventTriggerAlert:valid", func(t *testing.T) {
		rName1 := U.RandomString(5)
		alert, errCode, errMsg := store.GetStore().CreateEventTriggerAlert(agent.UUID, "", project.ID, &model.EventTriggerAlertConfig{
			Title: rName1, Event: rName1, Message: "Remember", MessageProperty: &postgres.Jsonb{},
			DontRepeatAlerts: true, CoolDownTime: 1800, BreakdownProperties: &postgres.Jsonb{breakdownProps}, AlertLimit: 5, SetAlertLimit: true,
			Slack: true, SlackChannels: &postgres.Jsonb{RawMessage: slackChannelJson}, Filter: []model.QueryProperty{
				{Entity: "", Type: "categorical", Property: "campaign", Operator: "equals", LogicalOp: "AND"},
			}}, agent.UUID, agent.UUID, false)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, alert)
		assert.Empty(t, errMsg)
	})

	t.Run("GetAllEventTriggerAlerts:valid", func(t *testing.T) {
		rName1 := U.RandomString(5)
		alert, errCode, errMsg := store.GetStore().CreateEventTriggerAlert(agent.UUID, "", project.ID, &model.EventTriggerAlertConfig{
			Title: rName1, Event: rName1, Message: "Remember", MessageProperty: &postgres.Jsonb{},
			DontRepeatAlerts: true, CoolDownTime: 1800, BreakdownProperties: &postgres.Jsonb{breakdownProps}, AlertLimit: 5, SetAlertLimit: true,
			Slack: true, SlackChannels: &postgres.Jsonb{RawMessage: slackChannelJson}, Filter: []model.QueryProperty{
				{Entity: "", Type: "categorical", Property: "campaign", Operator: "equals", LogicalOp: "AND"},
			}}, agent.UUID, agent.UUID, false)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, alert)
		assert.Empty(t, errMsg)

		alerts, errCode := store.GetStore().GetAllEventTriggerAlertsByProject(project.ID)
		assert.Equal(t, http.StatusFound, errCode)
		assert.NotNil(t, alerts)
	})

	t.Run("CreateEventTriggerAlert:Title already present:Invalid", func(t *testing.T) {
		rName1 := U.RandomString(5)
		alert, errCode, errMsg := store.GetStore().CreateEventTriggerAlert(agent.UUID, "", project.ID, &model.EventTriggerAlertConfig{
			Title: rName1, Event: rName1, Message: "Remember", MessageProperty: &postgres.Jsonb{},
			DontRepeatAlerts: true, CoolDownTime: 1800, BreakdownProperties: &postgres.Jsonb{breakdownProps}, AlertLimit: 5, SetAlertLimit: true,
			Slack: true, SlackChannels: &postgres.Jsonb{RawMessage: slackChannelJson}, Filter: []model.QueryProperty{
				{Entity: "", Type: "categorical", Property: "campaign", Operator: "equals", LogicalOp: "AND"},
			}}, agent.UUID, agent.UUID, false)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, alert)
		assert.Empty(t, errMsg)

		alert, errCode, errMsg = store.GetStore().CreateEventTriggerAlert(agent.UUID, "", project.ID, &model.EventTriggerAlertConfig{
			Title: rName1, Event: rName1, Message: "Remember", MessageProperty: &postgres.Jsonb{},
			DontRepeatAlerts: true, CoolDownTime: 1800, BreakdownProperties: &postgres.Jsonb{breakdownProps}, AlertLimit: 5, SetAlertLimit: true,
			Slack: true, SlackChannels: &postgres.Jsonb{RawMessage: slackChannelJson}, Filter: []model.QueryProperty{
				{Entity: "", Type: "categorical", Property: "campaign", Operator: "equals", LogicalOp: "AND"},
			}}, agent.UUID, agent.UUID, false)
		assert.Equal(t, http.StatusConflict, errCode)
		assert.Nil(t, alert)
		assert.NotEqual(t, "", errMsg)
	})

	t.Run("CreateEventTriggerAlert:BreakdownProperties not selected for DontRepeatAlert:Invalid", func(t *testing.T) {
		rName1 := U.RandomString(5)
		alert, errCode, errMsg := store.GetStore().CreateEventTriggerAlert(agent.UUID, "", project.ID, &model.EventTriggerAlertConfig{
			Title: rName1, Event: rName1, Message: "Remember", MessageProperty: &postgres.Jsonb{},
			DontRepeatAlerts: true, CoolDownTime: 1800, BreakdownProperties: &postgres.Jsonb{}, AlertLimit: 5, SetAlertLimit: true,
			Slack: true, SlackChannels: &postgres.Jsonb{RawMessage: slackChannelJson}, Filter: []model.QueryProperty{
				{Entity: "", Type: "categorical", Property: "campaign", Operator: "equals", LogicalOp: "AND"},
			}}, agent.UUID, agent.UUID, false)
		assert.Equal(t, http.StatusBadRequest, errCode)
		assert.Nil(t, alert)
		assert.NotEmpty(t, errMsg)
	})
}

func TestDeleteEventTriggerAlert(t *testing.T) {
	project, agent, err := SetupProjectWithSlackIntegratedAgentDAO()
	assert.Nil(t, err)

	slackChannel := model.SlackChannel{
		Name:      "channel1",
		Id:        U.GetUUID(),
		IsPrivate: true,
	}
	slackChannelJson, err := json.Marshal(slackChannel)
	assert.Nil(t, err)

	breakdownProps, err := json.Marshal(`[
        {
            "pr": "user_id",
            "en": "user",
            "pty": "categorical",
            "ena": "$hubspot_contact_created",
            "eni": 1
        }
    ]`)
	assert.Nil(t, err)

	rName1 := U.RandomString(5)
	alert, errCode, errMsg := store.GetStore().CreateEventTriggerAlert(agent.UUID, "", project.ID, &model.EventTriggerAlertConfig{
		Title: rName1, Event: rName1, Message: "Remember", MessageProperty: &postgres.Jsonb{},
		DontRepeatAlerts: true, CoolDownTime: 1800, BreakdownProperties: &postgres.Jsonb{breakdownProps}, AlertLimit: 5, SetAlertLimit: true,
		Slack: true, SlackChannels: &postgres.Jsonb{RawMessage: slackChannelJson}, Filter: []model.QueryProperty{
			{Entity: "", Type: "categorical", Property: "campaign", Operator: "equals", LogicalOp: "AND"},
		}}, agent.UUID, agent.UUID, false)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Empty(t, errMsg)
	assert.NotNil(t, alert)

	t.Run("DeleteEventTriggerAlert:valid", func(t *testing.T) {
		errCode, errMsg = store.GetStore().DeleteEventTriggerAlert(project.ID, alert.ID)
		assert.Equal(t, http.StatusAccepted, errCode)
		assert.Empty(t, errMsg)
	})
}

func TestMatchEventTriggerAlert(t *testing.T) {

	start := time.Now()

	slackChannel := model.SlackChannel{
		Name:      "channel1",
		Id:        U.GetUUID(),
		IsPrivate: false,
	}
	slackChannelJson, err := json.Marshal(slackChannel)
	assert.NotNil(t, slackChannelJson)
	assert.Nil(t, err)

	breakdownProps, err := json.Marshal(`[
        {
            "pr": "user_id",
            "en": "user",
            "pty": "categorical",
            "ena": "$hubspot_contact_created",
            "eni": 1
        }
    ]`)
	assert.Nil(t, err)

	messageProperty := []model.QueryGroupByProperty{
		{
			Entity:         "event",
			Property:       "$page_title",
			Type:           "categorical",
			EventNameIndex: 1,
		},
		{
			Entity:         "user",
			Property:       "$country",
			Type:           "categorical",
			EventNameIndex: 2,
		},
	}
	messagePropertyJson, err := json.Marshal(messageProperty)
	assert.NotNil(t, messagePropertyJson)
	assert.Nil(t, err)

	rName1 := U.RandomString(5)

	// Test for one filter with type categorical and operator equals Valid
	t.Run("MatchEventTriggerAlert:EqualsConditionMatchFound", func(t *testing.T) {
		project, user, eventName, err := SetupProjectUserEventNameReturnDAO()
		assert.NotNil(t, eventName)
		assert.NotNil(t, project)
		assert.NotNil(t, user)
		assert.Nil(t, err)

		agent, errCode := SetupAgentReturnWithSlackIntegrationDAO(getRandomEmail(), "+1343545", project.ID)
		assert.NotNil(t, agent)
		assert.Equal(t, http.StatusCreated, errCode)

		//Test for successful CreateAlert
		alert, errCode, errMsg := store.GetStore().CreateEventTriggerAlert(agent.UUID, "", project.ID, &model.EventTriggerAlertConfig{
			Title: rName1, Event: eventName.Name, Message: "Remember", MessageProperty: &postgres.Jsonb{RawMessage: messagePropertyJson},
			DontRepeatAlerts: true, CoolDownTime: 1800, BreakdownProperties: &postgres.Jsonb{breakdownProps}, AlertLimit: 5, SetAlertLimit: true,
			Slack: true, SlackChannels: &postgres.Jsonb{RawMessage: slackChannelJson}, Filter: []model.QueryProperty{
				{Entity: "event", Type: "categorical", Property: "$country", Operator: "equals", LogicalOp: "AND", Value: "US"},
			}}, agent.UUID, agent.UUID, false)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.Empty(t, errMsg)
		assert.NotNil(t, alert)

		event := &model.Event{EventNameId: eventName.ID, ProjectId: project.ID,
			UserId: agent.UUID, Timestamp: start.Unix(),
			UserProperties: &postgres.Jsonb{RawMessage: []byte{}},
			Properties:     postgres.Jsonb{RawMessage: []byte(`{"$country": "US"}`)}}

		alerts, _, errCode := store.GetStore().MatchEventTriggerAlertWithTrackPayload(project.ID, eventName.ID, &event.Properties, event.UserProperties, nil, false)
		assert.Equal(t, http.StatusFound, errCode)
		assert.NotNil(t, alerts)
	})

	// Test for one filter with type categorical and operator equals Invalid
	t.Run("MatchEventTriggerAlert:EqualsConditionMatchNotFound", func(t *testing.T) {
		project, user, eventName, err := SetupProjectUserEventNameReturnDAO()
		assert.NotNil(t, eventName)
		assert.NotNil(t, project)
		assert.NotNil(t, user)
		assert.Nil(t, err)

		agent, errCode := SetupAgentReturnWithSlackIntegrationDAO(getRandomEmail(), "+1343545", project.ID)
		assert.NotNil(t, agent)
		assert.Equal(t, http.StatusCreated, errCode)

		//Test for successful CreateAlert
		alert, errCode, errMsg := store.GetStore().CreateEventTriggerAlert(agent.UUID, "", project.ID, &model.EventTriggerAlertConfig{
			Title: rName1, Event: eventName.Name, Message: "Remember", MessageProperty: &postgres.Jsonb{RawMessage: messagePropertyJson},
			DontRepeatAlerts: true, CoolDownTime: 1800, BreakdownProperties: &postgres.Jsonb{breakdownProps}, AlertLimit: 5, SetAlertLimit: true,
			Slack: true, SlackChannels: &postgres.Jsonb{RawMessage: slackChannelJson}, Filter: []model.QueryProperty{
				{Entity: "event", Type: "categorical", Property: "$country", Operator: "equals", LogicalOp: "AND", Value: "US"},
			}}, agent.UUID, agent.UUID, false)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.Empty(t, errMsg)
		assert.NotNil(t, alert)

		event := &model.Event{EventNameId: eventName.ID, ProjectId: project.ID,
			UserId: agent.UUID, Timestamp: start.Unix(),
			UserProperties: &postgres.Jsonb{RawMessage: []byte{}},
			Properties:     postgres.Jsonb{RawMessage: []byte(`{"$country": "UK"}`)}}

		alerts, _, errCode := store.GetStore().MatchEventTriggerAlertWithTrackPayload(project.ID, eventName.ID, &event.Properties, event.UserProperties, nil, false)
		assert.Equal(t, http.StatusNotFound, errCode)
		assert.Nil(t, alerts)
	})

	// Test for one filter of type categorical and operator notEqual Valid
	t.Run("MatchEventTriggerAlert:NotEqualsConditionValid", func(t *testing.T) {
		project, user, eventName, err := SetupProjectUserEventNameReturnDAO()
		assert.NotNil(t, eventName)
		assert.NotNil(t, project)
		assert.NotNil(t, user)
		assert.Nil(t, err)

		agent, errCode := SetupAgentReturnWithSlackIntegrationDAO(getRandomEmail(), "+1343545", project.ID)
		assert.NotNil(t, agent)
		assert.Equal(t, http.StatusCreated, errCode)

		//Test for successful CreateAlert
		alert, errCode, errMsg := store.GetStore().CreateEventTriggerAlert(agent.UUID, "", project.ID, &model.EventTriggerAlertConfig{
			Title: rName1, Event: eventName.Name, Message: "Remember", MessageProperty: &postgres.Jsonb{RawMessage: messagePropertyJson},
			DontRepeatAlerts: true, CoolDownTime: 1800, BreakdownProperties: &postgres.Jsonb{breakdownProps}, AlertLimit: 5, SetAlertLimit: true,
			Slack: true, SlackChannels: &postgres.Jsonb{RawMessage: slackChannelJson}, Filter: []model.QueryProperty{
				{Entity: "event", Type: "categorical", Property: "$country", Operator: "notEqual", LogicalOp: "AND", Value: "UK"},
			}}, agent.UUID, agent.UUID, false)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.Empty(t, errMsg)
		assert.NotNil(t, alert)

		event := &model.Event{EventNameId: eventName.ID, ProjectId: project.ID,
			UserId: agent.UUID, Timestamp: start.Unix(),
			UserProperties: &postgres.Jsonb{RawMessage: []byte{}},
			Properties:     postgres.Jsonb{RawMessage: []byte(`{"$country": "UK"}`)}}

		alerts, _, errCode := store.GetStore().MatchEventTriggerAlertWithTrackPayload(project.ID, eventName.ID, &event.Properties, event.UserProperties, nil, false)
		assert.Equal(t, http.StatusNotFound, errCode)
		assert.Nil(t, alerts)
	})

	// Test for one filter of type categorical and operator notEqual Invalid
	t.Run("MatchEventTriggerAlert:NotEqualsConditionInvalid", func(t *testing.T) {
		project, user, eventName, err := SetupProjectUserEventNameReturnDAO()
		assert.NotNil(t, eventName)
		assert.NotNil(t, project)
		assert.NotNil(t, user)
		assert.Nil(t, err)

		agent, errCode := SetupAgentReturnWithSlackIntegrationDAO(getRandomEmail(), "+1343545", project.ID)
		assert.NotNil(t, agent)
		assert.Equal(t, http.StatusCreated, errCode)

		//Test for successful CreateAlert
		alert, errCode, errMsg := store.GetStore().CreateEventTriggerAlert(agent.UUID, "", project.ID, &model.EventTriggerAlertConfig{
			Title: rName1, Event: eventName.Name, Message: "Remember", MessageProperty: &postgres.Jsonb{RawMessage: messagePropertyJson},
			DontRepeatAlerts: true, CoolDownTime: 1800, BreakdownProperties: &postgres.Jsonb{breakdownProps}, AlertLimit: 5, SetAlertLimit: true,
			Slack: true, SlackChannels: &postgres.Jsonb{RawMessage: slackChannelJson}, Filter: []model.QueryProperty{
				{Entity: "event", Type: "categorical", Property: "$country", Operator: "notEqual", LogicalOp: "AND", Value: "US"},
			}}, agent.UUID, agent.UUID, false)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.Empty(t, errMsg)
		assert.NotNil(t, alert)

		event := &model.Event{EventNameId: eventName.ID, ProjectId: project.ID,
			UserId: agent.UUID, Timestamp: start.Unix(),
			UserProperties: &postgres.Jsonb{RawMessage: []byte{}},
			Properties:     postgres.Jsonb{RawMessage: []byte(`{"$country": "UK"}`)}}

		alerts, _, errCode := store.GetStore().MatchEventTriggerAlertWithTrackPayload(project.ID, eventName.ID, &event.Properties, event.UserProperties, nil, false)
		assert.Equal(t, http.StatusFound, errCode)
		assert.NotNil(t, alerts)
	})

	// Test for one filter of type categorical and operator contains Valid
	t.Run("MatchEventTriggerAlert:ContainsConditionValid", func(t *testing.T) {
		project, user, eventName, err := SetupProjectUserEventNameReturnDAO()
		assert.NotNil(t, eventName)
		assert.NotNil(t, project)
		assert.NotNil(t, user)
		assert.Nil(t, err)

		agent, errCode := SetupAgentReturnWithSlackIntegrationDAO(getRandomEmail(), "+1343545", project.ID)
		assert.NotNil(t, agent)
		assert.Equal(t, http.StatusCreated, errCode)

		//Test for successful CreateAlert
		alert, errCode, errMsg := store.GetStore().CreateEventTriggerAlert(agent.UUID, "", project.ID, &model.EventTriggerAlertConfig{
			Title: rName1, Event: eventName.Name, Message: "Remember", MessageProperty: &postgres.Jsonb{RawMessage: messagePropertyJson},
			DontRepeatAlerts: true, CoolDownTime: 1800, BreakdownProperties: &postgres.Jsonb{breakdownProps}, AlertLimit: 5, SetAlertLimit: true,
			Slack: true, SlackChannels: &postgres.Jsonb{RawMessage: slackChannelJson}, Filter: []model.QueryProperty{
				{Entity: "event", Type: "categorical", Property: "$country", Operator: "contains", LogicalOp: "AND", Value: "US"},
			}}, agent.UUID, agent.UUID, false)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.Empty(t, errMsg)
		assert.NotNil(t, alert)

		event := &model.Event{EventNameId: eventName.ID, ProjectId: project.ID,
			UserId: agent.UUID, Timestamp: start.Unix(),
			UserProperties: &postgres.Jsonb{RawMessage: []byte{}},
			Properties:     postgres.Jsonb{RawMessage: []byte(`{"$country": "USA"}`)}}

		alerts, _, errCode := store.GetStore().MatchEventTriggerAlertWithTrackPayload(project.ID, eventName.ID, &event.Properties, event.UserProperties, nil, false)
		assert.Equal(t, http.StatusFound, errCode)
		assert.NotNil(t, alerts)
	})

	// Test for one filter of type categorical and operator contains Invalid
	t.Run("MatchEventTriggerAlert:ContainsConditionInvalid", func(t *testing.T) {
		project, user, eventName, err := SetupProjectUserEventNameReturnDAO()
		assert.NotNil(t, eventName)
		assert.NotNil(t, project)
		assert.NotNil(t, user)
		assert.Nil(t, err)

		agent, errCode := SetupAgentReturnWithSlackIntegrationDAO(getRandomEmail(), "+1343545", project.ID)
		assert.NotNil(t, agent)
		assert.Equal(t, http.StatusCreated, errCode)

		//Test for successful CreateAlert
		alert, errCode, errMsg := store.GetStore().CreateEventTriggerAlert(agent.UUID, "", project.ID, &model.EventTriggerAlertConfig{
			Title: rName1, Event: eventName.Name, Message: "Remember", MessageProperty: &postgres.Jsonb{RawMessage: messagePropertyJson},
			DontRepeatAlerts: true, CoolDownTime: 1800, BreakdownProperties: &postgres.Jsonb{breakdownProps}, AlertLimit: 5, SetAlertLimit: true,
			Slack: true, SlackChannels: &postgres.Jsonb{RawMessage: slackChannelJson}, Filter: []model.QueryProperty{
				{Entity: "event", Type: "categorical", Property: "$country", Operator: "contains", LogicalOp: "AND", Value: "UK"},
			}}, agent.UUID, agent.UUID, false)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.Empty(t, errMsg)
		assert.NotNil(t, alert)

		event := &model.Event{EventNameId: eventName.ID, ProjectId: project.ID,
			UserId: agent.UUID, Timestamp: start.Unix(),
			UserProperties: &postgres.Jsonb{RawMessage: []byte{}},
			Properties:     postgres.Jsonb{RawMessage: []byte(`{"$country": "USA"}`)}}

		alerts, _, errCode := store.GetStore().MatchEventTriggerAlertWithTrackPayload(project.ID, eventName.ID, &event.Properties, event.UserProperties, nil, false)
		assert.Equal(t, http.StatusNotFound, errCode)
		assert.Nil(t, alerts)
	})

	// Test for one filter of type categorical and operator notContains Valid
	t.Run("MatchEventTriggerAlert:NotContainsConditionValid", func(t *testing.T) {
		project, user, eventName, err := SetupProjectUserEventNameReturnDAO()
		assert.NotNil(t, eventName)
		assert.NotNil(t, project)
		assert.NotNil(t, user)
		assert.Nil(t, err)

		agent, errCode := SetupAgentReturnWithSlackIntegrationDAO(getRandomEmail(), "+1343545", project.ID)
		assert.NotNil(t, agent)
		assert.Equal(t, http.StatusCreated, errCode)

		//Test for successful CreateAlert
		alert, errCode, errMsg := store.GetStore().CreateEventTriggerAlert(agent.UUID, "", project.ID, &model.EventTriggerAlertConfig{
			Title: rName1, Event: eventName.Name, Message: "Remember", MessageProperty: &postgres.Jsonb{RawMessage: messagePropertyJson},
			DontRepeatAlerts: true, CoolDownTime: 1800, BreakdownProperties: &postgres.Jsonb{breakdownProps}, AlertLimit: 5, SetAlertLimit: true,
			Slack: true, SlackChannels: &postgres.Jsonb{RawMessage: slackChannelJson}, Filter: []model.QueryProperty{
				{Entity: "event", Type: "categorical", Property: "$country", Operator: "notContains", LogicalOp: "AND", Value: "US"},
			}}, agent.UUID, agent.UUID, false)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.Empty(t, errMsg)
		assert.NotNil(t, alert)

		event := &model.Event{EventNameId: eventName.ID, ProjectId: project.ID,
			UserId: agent.UUID, Timestamp: start.Unix(),
			UserProperties: &postgres.Jsonb{RawMessage: []byte{}},
			Properties:     postgres.Jsonb{RawMessage: []byte(`{"$country": "USA"}`)}}

		alerts, _, errCode := store.GetStore().MatchEventTriggerAlertWithTrackPayload(project.ID, eventName.ID, &event.Properties, event.UserProperties, nil, false)
		assert.Equal(t, http.StatusNotFound, errCode)
		assert.Nil(t, alerts)
	})

	// Test for one filter of type categorical and operator notContains Invalid
	t.Run("MatchEventTriggerAlert:NotContainsConditionInvalid", func(t *testing.T) {
		project, user, eventName, err := SetupProjectUserEventNameReturnDAO()
		assert.NotNil(t, eventName)
		assert.NotNil(t, project)
		assert.NotNil(t, user)
		assert.Nil(t, err)

		agent, errCode := SetupAgentReturnWithSlackIntegrationDAO(getRandomEmail(), "+1343545", project.ID)
		assert.NotNil(t, agent)
		assert.Equal(t, http.StatusCreated, errCode)

		//Test for successful CreateAlert
		alert, errCode, errMsg := store.GetStore().CreateEventTriggerAlert(agent.UUID, "", project.ID, &model.EventTriggerAlertConfig{
			Title: rName1, Event: eventName.Name, Message: "Remember", MessageProperty: &postgres.Jsonb{RawMessage: messagePropertyJson},
			DontRepeatAlerts: true, CoolDownTime: 1800, BreakdownProperties: &postgres.Jsonb{breakdownProps}, AlertLimit: 5, SetAlertLimit: true,
			Slack: true, SlackChannels: &postgres.Jsonb{RawMessage: slackChannelJson}, Filter: []model.QueryProperty{
				{Entity: "event", Type: "categorical", Property: "$country", Operator: "notContains", LogicalOp: "AND", Value: "asu"},
			}}, agent.UUID, agent.UUID, false)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.Empty(t, errMsg)
		assert.NotNil(t, alert)

		event := &model.Event{EventNameId: eventName.ID, ProjectId: project.ID,
			UserId: agent.UUID, Timestamp: start.Unix(),
			UserProperties: &postgres.Jsonb{RawMessage: []byte{}},
			Properties:     postgres.Jsonb{RawMessage: []byte(`{"$country": "USA"}`)}}

		alerts, _, errCode := store.GetStore().MatchEventTriggerAlertWithTrackPayload(project.ID, eventName.ID, &event.Properties, event.UserProperties, nil, false)
		assert.Equal(t, http.StatusFound, errCode)
		assert.NotNil(t, alerts)
	})

	// Test for multiple filters of type categorical and operator equals Valid
	t.Run("MatchEventTriggerAlert:MultipleEqualsConditionsValid", func(t *testing.T) {
		project, user, eventName, err := SetupProjectUserEventNameReturnDAO()
		assert.NotNil(t, eventName)
		assert.NotNil(t, project)
		assert.NotNil(t, user)
		assert.Nil(t, err)

		agent, errCode := SetupAgentReturnWithSlackIntegrationDAO(getRandomEmail(), "+1343545", project.ID)
		assert.NotNil(t, agent)
		assert.Equal(t, http.StatusCreated, errCode)

		//Test for successful CreateAlert
		alert, errCode, errMsg := store.GetStore().CreateEventTriggerAlert(agent.UUID, "", project.ID, &model.EventTriggerAlertConfig{
			Title: rName1, Event: eventName.Name, Message: "Remember", MessageProperty: &postgres.Jsonb{RawMessage: messagePropertyJson},
			DontRepeatAlerts: true, CoolDownTime: 1800, BreakdownProperties: &postgres.Jsonb{breakdownProps}, AlertLimit: 5, SetAlertLimit: true,
			Slack: true, SlackChannels: &postgres.Jsonb{RawMessage: slackChannelJson},
			Filter: []model.QueryProperty{
				{Entity: "event", Type: "categorical", Property: "$country", Operator: "equals", LogicalOp: "AND", Value: "US"},
				{Entity: "event", Type: "categorical", Property: "$country", Operator: "equals", LogicalOp: "OR", Value: "UK"},
			}}, agent.UUID, agent.UUID, false)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.Empty(t, errMsg)
		assert.NotNil(t, alert)

		event := &model.Event{EventNameId: eventName.ID, ProjectId: project.ID,
			UserId: agent.UUID, Timestamp: start.Unix(),
			UserProperties: &postgres.Jsonb{RawMessage: []byte{}},
			Properties:     postgres.Jsonb{RawMessage: []byte(`{"$country": "US"}`)}}

		alerts, _, errCode := store.GetStore().MatchEventTriggerAlertWithTrackPayload(project.ID, eventName.ID, &event.Properties, event.UserProperties, nil, false)
		assert.Equal(t, http.StatusFound, errCode)
		assert.NotNil(t, alerts)

		event1 := &model.Event{EventNameId: eventName.ID, ProjectId: project.ID,
			UserId: agent.UUID, Timestamp: start.Unix(),
			UserProperties: &postgres.Jsonb{RawMessage: []byte{}},
			Properties:     postgres.Jsonb{RawMessage: []byte(`{"$country": "UK"}`)}}

		alerts1, _, errCode := store.GetStore().MatchEventTriggerAlertWithTrackPayload(project.ID, eventName.ID, &event1.Properties, event1.UserProperties, nil, false)
		assert.Equal(t, http.StatusFound, errCode)
		assert.NotNil(t, alerts1)
	})

	// Test for multiple filters of type categorical and operator equals Invalid
	t.Run("MatchEventTriggerAlert:MultipleEqualsConditionsInvalid", func(t *testing.T) {
		project, user, eventName, err := SetupProjectUserEventNameReturnDAO()
		assert.NotNil(t, eventName)
		assert.NotNil(t, project)
		assert.NotNil(t, user)
		assert.Nil(t, err)

		agent, errCode := SetupAgentReturnWithSlackIntegrationDAO(getRandomEmail(), "+1343545", project.ID)
		assert.NotNil(t, agent)
		assert.Equal(t, http.StatusCreated, errCode)

		//Test for successful CreateAlert
		alert, errCode, errMsg := store.GetStore().CreateEventTriggerAlert(agent.UUID, "", project.ID, &model.EventTriggerAlertConfig{
			Title: rName1, Event: eventName.Name, Message: "Remember", MessageProperty: &postgres.Jsonb{RawMessage: messagePropertyJson},
			DontRepeatAlerts: true, CoolDownTime: 1800, BreakdownProperties: &postgres.Jsonb{breakdownProps}, AlertLimit: 5, SetAlertLimit: true,
			Slack: true, SlackChannels: &postgres.Jsonb{RawMessage: slackChannelJson},
			Filter: []model.QueryProperty{
				{Entity: "event", Type: "categorical", Property: "$country", Operator: "equals", LogicalOp: "AND", Value: "US"},
				{Entity: "event", Type: "categorical", Property: "$country", Operator: "equals", LogicalOp: "OR", Value: "UK"},
			}}, agent.UUID, agent.UUID, false)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.Empty(t, errMsg)
		assert.NotNil(t, alert)

		event := &model.Event{EventNameId: eventName.ID, ProjectId: project.ID,
			UserId: agent.UUID, Timestamp: start.Unix(),
			UserProperties: &postgres.Jsonb{RawMessage: []byte{}},
			Properties:     postgres.Jsonb{RawMessage: []byte(`{"$country": "USA"}`)}}

		alerts, _, errCode := store.GetStore().MatchEventTriggerAlertWithTrackPayload(project.ID, eventName.ID, &event.Properties, event.UserProperties, nil, false)
		assert.Equal(t, http.StatusNotFound, errCode)
		assert.Nil(t, alerts)
	})

	//Test for multiple filters of categorical type and notEqual condition Valid
	t.Run("MatchEventTriggerAlert:MultipleNotEqualConditionsValid", func(t *testing.T) {
		project, user, eventName, err := SetupProjectUserEventNameReturnDAO()
		assert.NotNil(t, eventName)
		assert.NotNil(t, project)
		assert.NotNil(t, user)
		assert.Nil(t, err)

		agent, errCode := SetupAgentReturnWithSlackIntegrationDAO(getRandomEmail(), "+1343545", project.ID)
		assert.NotNil(t, agent)
		assert.Equal(t, http.StatusCreated, errCode)

		//Test for successful CreateAlert
		alert, errCode, errMsg := store.GetStore().CreateEventTriggerAlert(agent.UUID, "", project.ID, &model.EventTriggerAlertConfig{
			Title: rName1, Event: eventName.Name, Message: "Remember", MessageProperty: &postgres.Jsonb{RawMessage: messagePropertyJson},
			DontRepeatAlerts: true, CoolDownTime: 1800, BreakdownProperties: &postgres.Jsonb{breakdownProps}, AlertLimit: 5, SetAlertLimit: true,
			Slack: true, SlackChannels: &postgres.Jsonb{RawMessage: slackChannelJson},
			Filter: []model.QueryProperty{
				{Entity: "event", Type: "categorical", Property: "$country", Operator: "notEqual", LogicalOp: "AND", Value: "US"},
				{Entity: "event", Type: "categorical", Property: "$country", Operator: "notEqual", LogicalOp: "OR", Value: "UK"},
			}}, agent.UUID, agent.UUID, false)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.Empty(t, errMsg)
		assert.NotNil(t, alert)

		event := &model.Event{EventNameId: eventName.ID, ProjectId: project.ID,
			UserId: agent.UUID, Timestamp: start.Unix(),
			UserProperties: &postgres.Jsonb{RawMessage: []byte(``)},
			Properties:     postgres.Jsonb{RawMessage: []byte(`{"$country": "US"}`)}}

		alerts, _, errCode := store.GetStore().MatchEventTriggerAlertWithTrackPayload(project.ID, eventName.ID, &event.Properties, event.UserProperties, nil, false)
		assert.Equal(t, http.StatusNotFound, errCode)
		assert.Nil(t, alerts)

		event1 := &model.Event{EventNameId: eventName.ID, ProjectId: project.ID,
			UserId: agent.UUID, Timestamp: start.Unix(),
			UserProperties: &postgres.Jsonb{RawMessage: []byte(`{}`)},
			Properties:     postgres.Jsonb{RawMessage: []byte(`{"$country": "UK"}`)}}

		alerts1, _, errCode := store.GetStore().MatchEventTriggerAlertWithTrackPayload(project.ID, eventName.ID, &event1.Properties, event1.UserProperties, nil, false)
		assert.Equal(t, http.StatusNotFound, errCode)
		assert.Nil(t, alerts1)
	})

	//Test for multiple filters of categorical type and notEqual condtions Invalid
	t.Run("MatchEventTriggerAlert:MultipleNotEqualConditionsInvalid", func(t *testing.T) {
		project, user, eventName, err := SetupProjectUserEventNameReturnDAO()
		assert.NotNil(t, eventName)
		assert.NotNil(t, project)
		assert.NotNil(t, user)
		assert.Nil(t, err)

		agent, errCode := SetupAgentReturnWithSlackIntegrationDAO(getRandomEmail(), "+1343545", project.ID)
		assert.NotNil(t, agent)
		assert.Equal(t, http.StatusCreated, errCode)

		//Test for successful CreateAlert
		alert, errCode, errMsg := store.GetStore().CreateEventTriggerAlert(agent.UUID, "", project.ID, &model.EventTriggerAlertConfig{
			Title: rName1, Event: eventName.Name, Message: "Remember", MessageProperty: &postgres.Jsonb{RawMessage: messagePropertyJson},
			DontRepeatAlerts: true, CoolDownTime: 1800, BreakdownProperties: &postgres.Jsonb{breakdownProps}, AlertLimit: 5, SetAlertLimit: true,
			Slack: true, SlackChannels: &postgres.Jsonb{RawMessage: slackChannelJson},
			Filter: []model.QueryProperty{
				{Entity: "event", Type: "categorical", Property: "$country", Operator: "notEqual", LogicalOp: "AND", Value: "US"},
				{Entity: "event", Type: "categorical", Property: "$country", Operator: "notEqual", LogicalOp: "AND", Value: "UK"},
			}}, agent.UUID, agent.UUID, false)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.Empty(t, errMsg)
		assert.NotNil(t, alert)

		event := &model.Event{
			EventNameId: eventName.ID, ProjectId: project.ID,
			UserId: agent.UUID, Timestamp: start.Unix(),
			UserProperties: &postgres.Jsonb{RawMessage: []byte(`{"uuser":""}`)},
			Properties:     postgres.Jsonb{RawMessage: []byte(`{"$country":"canada"}`)},
		}

		alerts, _, errCode := store.GetStore().MatchEventTriggerAlertWithTrackPayload(project.ID, eventName.ID, &event.Properties, event.UserProperties, nil, false)
		assert.Equal(t, http.StatusFound, errCode)
		assert.NotNil(t, alerts)
	})

	//Test for multiple filters of categorical type and contains condtions valid
	t.Run("MatchEventTriggerAlert:MultipleContainsConditionsValid", func(t *testing.T) {
		project, user, eventName, err := SetupProjectUserEventNameReturnDAO()
		assert.NotNil(t, eventName)
		assert.NotNil(t, project)
		assert.NotNil(t, user)
		assert.Nil(t, err)

		agent, errCode := SetupAgentReturnWithSlackIntegrationDAO(getRandomEmail(), "+1343545", project.ID)
		assert.NotNil(t, agent)
		assert.Equal(t, http.StatusCreated, errCode)

		//Test for successful CreateAlert
		alert, errCode, errMsg := store.GetStore().CreateEventTriggerAlert(agent.UUID, "", project.ID, &model.EventTriggerAlertConfig{
			Title: rName1, Event: eventName.Name, Message: "Remember", MessageProperty: &postgres.Jsonb{RawMessage: messagePropertyJson},
			DontRepeatAlerts: true, CoolDownTime: 1800, BreakdownProperties: &postgres.Jsonb{breakdownProps}, AlertLimit: 5, SetAlertLimit: true,
			Slack: true, SlackChannels: &postgres.Jsonb{RawMessage: slackChannelJson},
			Filter: []model.QueryProperty{
				{Entity: "event", Type: "categorical", Property: "$country", Operator: "contains", LogicalOp: "AND", Value: "US"},
				{Entity: "event", Type: "categorical", Property: "$country", Operator: "contains", LogicalOp: "OR", Value: "UK"},
			}}, agent.UUID, agent.UUID, false)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.Empty(t, errMsg)
		assert.NotNil(t, alert)

		event := &model.Event{EventNameId: eventName.ID, ProjectId: project.ID,
			UserId: agent.UUID, Timestamp: start.Unix(),
			UserProperties: &postgres.Jsonb{RawMessage: []byte(``)},
			Properties:     postgres.Jsonb{RawMessage: []byte(`{"$country": "USA"}`)}}

		alerts, _, errCode := store.GetStore().MatchEventTriggerAlertWithTrackPayload(project.ID, eventName.ID, &event.Properties, event.UserProperties, nil, false)
		assert.Equal(t, http.StatusFound, errCode)
		assert.NotNil(t, alerts)

		event1 := &model.Event{
			EventNameId: eventName.ID, ProjectId: project.ID,
			UserId: agent.UUID, Timestamp: start.Unix(),
			UserProperties: &postgres.Jsonb{RawMessage: []byte(`{}`)},
			Properties:     postgres.Jsonb{RawMessage: []byte(`{"$country": "Ukraine"}`)},
		}

		alerts1, _, errCode := store.GetStore().MatchEventTriggerAlertWithTrackPayload(project.ID, eventName.ID, &event1.Properties, event1.UserProperties, nil, false)
		assert.Equal(t, http.StatusFound, errCode)
		assert.NotNil(t, alerts1)
	})

	//Test for multiple filters of categorical type and contains condtions Invalid
	t.Run("MatchEventTriggerAlert:MultipleContainsConditionsInvalid", func(t *testing.T) {
		project, user, eventName, err := SetupProjectUserEventNameReturnDAO()
		assert.NotNil(t, eventName)
		assert.NotNil(t, project)
		assert.NotNil(t, user)
		assert.Nil(t, err)

		agent, errCode := SetupAgentReturnWithSlackIntegrationDAO(getRandomEmail(), "+1343545", project.ID)
		assert.NotNil(t, agent)
		assert.Equal(t, http.StatusCreated, errCode)

		//Test for successful CreateAlert
		alert, errCode, errMsg := store.GetStore().CreateEventTriggerAlert(agent.UUID, "", project.ID, &model.EventTriggerAlertConfig{
			Title: rName1, Event: eventName.Name, Message: "Remember", MessageProperty: &postgres.Jsonb{RawMessage: messagePropertyJson},
			DontRepeatAlerts: true, CoolDownTime: 1800, BreakdownProperties: &postgres.Jsonb{breakdownProps}, AlertLimit: 5, SetAlertLimit: true,
			Slack: true, SlackChannels: &postgres.Jsonb{RawMessage: slackChannelJson},
			Filter: []model.QueryProperty{
				{Entity: "event", Type: "categorical", Property: "$country", Operator: "contains", LogicalOp: "AND", Value: "US"},
				{Entity: "event", Type: "categorical", Property: "$country", Operator: "contains", LogicalOp: "OR", Value: "UK"},
			}}, agent.UUID, agent.UUID, false)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.Empty(t, errMsg)
		assert.NotNil(t, alert)

		event := &model.Event{EventNameId: eventName.ID, ProjectId: project.ID,
			UserId: agent.UUID, Timestamp: start.Unix(),
			UserProperties: &postgres.Jsonb{RawMessage: []byte(``)},
			Properties:     postgres.Jsonb{RawMessage: []byte(`{"$country": "Canada"}`)}}

		alerts, _, errCode := store.GetStore().MatchEventTriggerAlertWithTrackPayload(project.ID, eventName.ID, &event.Properties, event.UserProperties, nil, false)
		assert.Equal(t, http.StatusNotFound, errCode)
		assert.Nil(t, alerts)
	})

	//Test for multiple filters of categorical type and notContains condtions valid and Invalid
	t.Run("MatchEventTriggerAlert:MultipleNotContainsConditions", func(t *testing.T) {
		project, user, eventName, err := SetupProjectUserEventNameReturnDAO()
		assert.NotNil(t, eventName)
		assert.NotNil(t, project)
		assert.NotNil(t, user)
		assert.Nil(t, err)

		agent, errCode := SetupAgentReturnWithSlackIntegrationDAO(getRandomEmail(), "+1343545", project.ID)
		assert.NotNil(t, agent)
		assert.Equal(t, http.StatusCreated, errCode)

		//Test for successful CreateAlert
		alert, errCode, errMsg := store.GetStore().CreateEventTriggerAlert(agent.UUID, "", project.ID, &model.EventTriggerAlertConfig{
			Title: rName1, Event: eventName.Name, Message: "Remember", MessageProperty: &postgres.Jsonb{RawMessage: messagePropertyJson},
			DontRepeatAlerts: true, CoolDownTime: 1800, BreakdownProperties: &postgres.Jsonb{breakdownProps}, AlertLimit: 5, SetAlertLimit: true,
			Slack: true, SlackChannels: &postgres.Jsonb{RawMessage: slackChannelJson},
			Filter: []model.QueryProperty{
				{Entity: "event", Type: "categorical", Property: "$country", Operator: "notContains", LogicalOp: "AND", Value: "US"},
				{Entity: "event", Type: "categorical", Property: "$country", Operator: "notContains", LogicalOp: "AND", Value: "UK"},
			}}, agent.UUID, agent.UUID, false)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.Empty(t, errMsg)
		assert.NotNil(t, alert)

		event := &model.Event{EventNameId: eventName.ID, ProjectId: project.ID,
			UserId: agent.UUID, Timestamp: start.Unix(),
			UserProperties: &postgres.Jsonb{RawMessage: []byte(``)},
			Properties:     postgres.Jsonb{RawMessage: []byte(`{"$country": "USA"}`)}}

		alerts, _, errCode := store.GetStore().MatchEventTriggerAlertWithTrackPayload(project.ID, eventName.ID, &event.Properties, event.UserProperties, nil, false)
		assert.Equal(t, http.StatusNotFound, errCode)
		assert.Nil(t, alerts)

		event1 := &model.Event{
			EventNameId: eventName.ID, ProjectId: project.ID,
			UserId: agent.UUID, Timestamp: start.Unix(),
			UserProperties: &postgres.Jsonb{RawMessage: []byte(`{}`)},
			Properties:     postgres.Jsonb{RawMessage: []byte(`{"$country": "Ukraine"}`)},
		}

		alerts1, _, errCode := store.GetStore().MatchEventTriggerAlertWithTrackPayload(project.ID, eventName.ID, &event1.Properties, event1.UserProperties, nil, false)
		assert.Equal(t, http.StatusNotFound, errCode)
		assert.Nil(t, alerts1)

		event2 := &model.Event{
			EventNameId: eventName.ID, ProjectId: project.ID,
			UserId: agent.UUID, Timestamp: start.Unix(),
			UserProperties: &postgres.Jsonb{RawMessage: []byte(``)},
			Properties:     postgres.Jsonb{RawMessage: []byte(`{"$country": "Canada"}`)},
		}

		alerts2, _, errCode := store.GetStore().MatchEventTriggerAlertWithTrackPayload(project.ID, eventName.ID, &event2.Properties, event2.UserProperties, nil, false)
		assert.Equal(t, http.StatusFound, errCode)
		assert.NotNil(t, alerts2)
	})

	//Test for multiple filters of categorical type and notContains condtions Invalid
	t.Run("MatchEventTriggerAlert:MultipleContainsConditionsValid", func(t *testing.T) {
		project, user, eventName, err := SetupProjectUserEventNameReturnDAO()
		assert.NotNil(t, eventName)
		assert.NotNil(t, project)
		assert.NotNil(t, user)
		assert.Nil(t, err)

		agent, errCode := SetupAgentReturnWithSlackIntegrationDAO(getRandomEmail(), "+1343545", project.ID)
		assert.NotNil(t, agent)
		assert.Equal(t, http.StatusCreated, errCode)

		//Test for successful CreateAlert
		alert, errCode, errMsg := store.GetStore().CreateEventTriggerAlert(agent.UUID, "", project.ID, &model.EventTriggerAlertConfig{
			Title: rName1, Event: eventName.Name, Message: "Remember", MessageProperty: &postgres.Jsonb{RawMessage: messagePropertyJson},
			DontRepeatAlerts: true, CoolDownTime: 1800, BreakdownProperties: &postgres.Jsonb{breakdownProps}, AlertLimit: 5, SetAlertLimit: true,
			Slack: true, SlackChannels: &postgres.Jsonb{RawMessage: slackChannelJson},
			Filter: []model.QueryProperty{
				{Entity: "event", Type: "categorical", Property: "$country", Operator: "notContains", LogicalOp: "AND", Value: "US"},
				{Entity: "event", Type: "categorical", Property: "$country", Operator: "notContains", LogicalOp: "OR", Value: "UK"},
			}}, agent.UUID, agent.UUID, false)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.Empty(t, errMsg)
		assert.NotNil(t, alert)

	})

	//Test for filter of numerical type and equals operator
	t.Run("MatchEventTriggerAlert:NumericalType", func(t *testing.T) {
		project, user, eventName, err := SetupProjectUserEventNameReturnDAO()
		assert.NotNil(t, eventName)
		assert.NotNil(t, project)
		assert.NotNil(t, user)
		assert.Nil(t, err)

		agent, errCode := SetupAgentReturnWithSlackIntegrationDAO(getRandomEmail(), "+1343545", project.ID)
		assert.NotNil(t, agent)
		assert.Equal(t, http.StatusCreated, errCode)

		//Test for successful CreateAlert
		alert, errCode, errMsg := store.GetStore().CreateEventTriggerAlert(agent.UUID, "", project.ID, &model.EventTriggerAlertConfig{
			Title: rName1, Event: eventName.Name, Message: "Remember", MessageProperty: &postgres.Jsonb{RawMessage: messagePropertyJson},
			DontRepeatAlerts: true, CoolDownTime: 1800, BreakdownProperties: &postgres.Jsonb{breakdownProps}, AlertLimit: 5, SetAlertLimit: true,
			Slack: true, SlackChannels: &postgres.Jsonb{RawMessage: slackChannelJson},
			Filter: []model.QueryProperty{
				{Entity: "event", Type: "numerical", Property: "$time_spent", Operator: "equals", LogicalOp: "AND", Value: "3000"},
			}}, agent.UUID, agent.UUID, false)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.Empty(t, errMsg)
		assert.NotNil(t, alert)

		event := &model.Event{EventNameId: eventName.ID, ProjectId: project.ID,
			UserId: agent.UUID, Timestamp: start.Unix(),
			UserProperties: &postgres.Jsonb{RawMessage: []byte(``)},
			Properties:     postgres.Jsonb{RawMessage: []byte(`{"$time_spent": "3000"}`)}}

		alerts, _, errCode := store.GetStore().MatchEventTriggerAlertWithTrackPayload(project.ID, eventName.ID, &event.Properties, event.UserProperties, nil, false)
		assert.Equal(t, http.StatusFound, errCode)
		assert.NotNil(t, alerts)

		event1 := &model.Event{EventNameId: eventName.ID, ProjectId: project.ID,
			UserId: agent.UUID, Timestamp: start.Unix(),
			UserProperties: &postgres.Jsonb{RawMessage: []byte(`{}`)},
			Properties:     postgres.Jsonb{RawMessage: []byte(`{"$time_spent": "3500"}`)}}

		alerts1, _, errCode := store.GetStore().MatchEventTriggerAlertWithTrackPayload(project.ID, eventName.ID, &event1.Properties, event1.UserProperties, nil, false)
		assert.Equal(t, http.StatusNotFound, errCode)
		assert.Nil(t, alerts1)
	})

	//Test for filter of numerical type and notEquals operator
	t.Run("MatchEventTriggerAlert:NumericalType", func(t *testing.T) {
		project, user, eventName, err := SetupProjectUserEventNameReturnDAO()
		assert.NotNil(t, eventName)
		assert.NotNil(t, project)
		assert.NotNil(t, user)
		assert.Nil(t, err)

		agent, errCode := SetupAgentReturnWithSlackIntegrationDAO(getRandomEmail(), "+1343545", project.ID)
		assert.NotNil(t, agent)
		assert.Equal(t, http.StatusCreated, errCode)

		//Test for successful CreateAlert
		alert, errCode, errMsg := store.GetStore().CreateEventTriggerAlert(agent.UUID, "", project.ID, &model.EventTriggerAlertConfig{
			Title: rName1, Event: eventName.Name, Message: "Remember", MessageProperty: &postgres.Jsonb{RawMessage: messagePropertyJson},
			DontRepeatAlerts: true, CoolDownTime: 1800, BreakdownProperties: &postgres.Jsonb{breakdownProps}, AlertLimit: 5, SetAlertLimit: true,
			Slack: true, SlackChannels: &postgres.Jsonb{RawMessage: slackChannelJson},
			Filter: []model.QueryProperty{
				{Entity: "event", Type: "numerical", Property: "$time_spent", Operator: "notEqual", LogicalOp: "AND", Value: "3000"},
			}}, agent.UUID, agent.UUID, false)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.Empty(t, errMsg)
		assert.NotNil(t, alert)

		event := &model.Event{EventNameId: eventName.ID, ProjectId: project.ID,
			UserId: agent.UUID, Timestamp: start.Unix(),
			UserProperties: &postgres.Jsonb{RawMessage: []byte(``)},
			Properties:     postgres.Jsonb{RawMessage: []byte(`{"$time_spent": "3500"}`)}}

		alerts, _, errCode := store.GetStore().MatchEventTriggerAlertWithTrackPayload(project.ID, eventName.ID, &event.Properties, event.UserProperties, nil, false)
		assert.Equal(t, http.StatusFound, errCode)
		assert.NotNil(t, alerts)

		event1 := &model.Event{EventNameId: eventName.ID, ProjectId: project.ID,
			UserId: agent.UUID, Timestamp: start.Unix(),
			UserProperties: &postgres.Jsonb{RawMessage: []byte(`{}`)},
			Properties:     postgres.Jsonb{RawMessage: []byte(`{"$time_spent": "3000"}`)}}

		alerts1, _, errCode := store.GetStore().MatchEventTriggerAlertWithTrackPayload(project.ID, eventName.ID, &event1.Properties, event1.UserProperties, nil, false)
		assert.Equal(t, http.StatusNotFound, errCode)
		assert.Nil(t, alerts1)
	})

	//Test for filter of numerical type and greaterThan operator
	t.Run("MatchEventTriggerAlert:NumericalTypeEquals", func(t *testing.T) {
		project, user, eventName, err := SetupProjectUserEventNameReturnDAO()
		assert.NotNil(t, eventName)
		assert.NotNil(t, project)
		assert.NotNil(t, user)
		assert.Nil(t, err)

		agent, errCode := SetupAgentReturnWithSlackIntegrationDAO(getRandomEmail(), "+1343545", project.ID)
		assert.NotNil(t, agent)
		assert.Equal(t, http.StatusCreated, errCode)

		//Test for successful CreateAlert
		alert, errCode, errMsg := store.GetStore().CreateEventTriggerAlert(agent.UUID, "", project.ID, &model.EventTriggerAlertConfig{
			Title: rName1, Event: eventName.Name, Message: "Remember", MessageProperty: &postgres.Jsonb{RawMessage: messagePropertyJson},
			DontRepeatAlerts: true, CoolDownTime: 1800, BreakdownProperties: &postgres.Jsonb{breakdownProps}, AlertLimit: 5, SetAlertLimit: true,
			Slack: true, SlackChannels: &postgres.Jsonb{RawMessage: slackChannelJson},
			Filter: []model.QueryProperty{
				{Entity: "event", Type: "numerical", Property: "$time_spent", Operator: "greaterThan", LogicalOp: "AND", Value: "3000"},
			}}, agent.UUID, agent.UUID, false)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.Empty(t, errMsg)
		assert.NotNil(t, alert)

		event := &model.Event{EventNameId: eventName.ID, ProjectId: project.ID,
			UserId: agent.UUID, Timestamp: start.Unix(),
			UserProperties: &postgres.Jsonb{RawMessage: []byte(``)},
			Properties:     postgres.Jsonb{RawMessage: []byte(`{"$time_spent": "3500"}`)}}

		alerts, _, errCode := store.GetStore().MatchEventTriggerAlertWithTrackPayload(project.ID, eventName.ID, &event.Properties, event.UserProperties, nil, false)
		assert.Equal(t, http.StatusFound, errCode)
		assert.NotNil(t, alerts)

		event1 := &model.Event{EventNameId: eventName.ID, ProjectId: project.ID,
			UserId: agent.UUID, Timestamp: start.Unix(),
			UserProperties: &postgres.Jsonb{RawMessage: []byte(`{}`)},
			Properties:     postgres.Jsonb{RawMessage: []byte(`{"$time_spent": "3000"}`)}}

		alerts1, _, errCode := store.GetStore().MatchEventTriggerAlertWithTrackPayload(project.ID, eventName.ID, &event1.Properties, event1.UserProperties, nil, false)
		assert.Equal(t, http.StatusNotFound, errCode)
		assert.Nil(t, alerts1)

		event2 := &model.Event{EventNameId: eventName.ID, ProjectId: project.ID,
			UserId: agent.UUID, Timestamp: start.Unix(),
			UserProperties: &postgres.Jsonb{RawMessage: []byte(`{}`)},
			Properties:     postgres.Jsonb{RawMessage: []byte(`{"$time_spent": "300"}`)}}

		alerts2, _, errCode := store.GetStore().MatchEventTriggerAlertWithTrackPayload(project.ID, eventName.ID, &event2.Properties, event2.UserProperties, nil, false)
		assert.Equal(t, http.StatusNotFound, errCode)
		assert.Nil(t, alerts2)
	})

	//Test for filter of numerical type and lesserThan operator
	t.Run("MatchEventTriggerAlert:NumericalTypeNotEqual", func(t *testing.T) {
		project, user, eventName, err := SetupProjectUserEventNameReturnDAO()
		assert.NotNil(t, eventName)
		assert.NotNil(t, project)
		assert.NotNil(t, user)
		assert.Nil(t, err)

		agent, errCode := SetupAgentReturnWithSlackIntegrationDAO(getRandomEmail(), "+1343545", project.ID)
		assert.NotNil(t, agent)
		assert.Equal(t, http.StatusCreated, errCode)

		//Test for successful CreateAlert
		alert, errCode, errMsg := store.GetStore().CreateEventTriggerAlert(agent.UUID, "", project.ID, &model.EventTriggerAlertConfig{
			Title: rName1, Event: eventName.Name, Message: "Remember", MessageProperty: &postgres.Jsonb{RawMessage: messagePropertyJson},
			DontRepeatAlerts: true, CoolDownTime: 1800, BreakdownProperties: &postgres.Jsonb{breakdownProps}, AlertLimit: 5, SetAlertLimit: true,
			Slack: true, SlackChannels: &postgres.Jsonb{RawMessage: slackChannelJson},
			Filter: []model.QueryProperty{
				{Entity: "event", Type: "numerical", Property: "$time_spent", Operator: "lesserThan", LogicalOp: "AND", Value: "3000"},
			}}, agent.UUID, agent.UUID, false)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.Empty(t, errMsg)
		assert.NotNil(t, alert)

		event := &model.Event{EventNameId: eventName.ID, ProjectId: project.ID,
			UserId: agent.UUID, Timestamp: start.Unix(),
			UserProperties: &postgres.Jsonb{RawMessage: []byte(``)},
			Properties:     postgres.Jsonb{RawMessage: []byte(`{"$time_spent": "2500"}`)}}

		alerts, _, errCode := store.GetStore().MatchEventTriggerAlertWithTrackPayload(project.ID, eventName.ID, &event.Properties, event.UserProperties, nil, false)
		assert.Equal(t, http.StatusFound, errCode)
		assert.NotNil(t, alerts)

		event1 := &model.Event{EventNameId: eventName.ID, ProjectId: project.ID,
			UserId: agent.UUID, Timestamp: start.Unix(),
			UserProperties: &postgres.Jsonb{RawMessage: []byte(`{}`)},
			Properties:     postgres.Jsonb{RawMessage: []byte(`{"$time_spent": "3000"}`)}}

		alerts1, _, errCode := store.GetStore().MatchEventTriggerAlertWithTrackPayload(project.ID, eventName.ID, &event1.Properties, event1.UserProperties, nil, false)
		assert.Equal(t, http.StatusNotFound, errCode)
		assert.Nil(t, alerts1)

		event2 := &model.Event{EventNameId: eventName.ID, ProjectId: project.ID,
			UserId: agent.UUID, Timestamp: start.Unix(),
			UserProperties: &postgres.Jsonb{RawMessage: []byte(`{}`)},
			Properties:     postgres.Jsonb{RawMessage: []byte(`{"$time_spent": "3500"}`)}}

		alerts2, _, errCode := store.GetStore().MatchEventTriggerAlertWithTrackPayload(project.ID, eventName.ID, &event2.Properties, event2.UserProperties, nil, false)
		assert.Equal(t, http.StatusNotFound, errCode)
		assert.Nil(t, alerts2)
	})

	//Test for filter of numerical type and greaterThanOrEqual operator
	t.Run("MatchEventTriggerAlert:NumericalTypeGreaterThanOrEqual", func(t *testing.T) {
		project, user, eventName, err := SetupProjectUserEventNameReturnDAO()
		assert.NotNil(t, eventName)
		assert.NotNil(t, project)
		assert.NotNil(t, user)
		assert.Nil(t, err)

		agent, errCode := SetupAgentReturnWithSlackIntegrationDAO(getRandomEmail(), "+1343545", project.ID)
		assert.NotNil(t, agent)
		assert.Equal(t, http.StatusCreated, errCode)

		//Test for successful CreateAlert
		alert, errCode, errMsg := store.GetStore().CreateEventTriggerAlert(agent.UUID, "", project.ID, &model.EventTriggerAlertConfig{
			Title: rName1, Event: eventName.Name, Message: "Remember", MessageProperty: &postgres.Jsonb{RawMessage: messagePropertyJson},
			DontRepeatAlerts: true, CoolDownTime: 1800, BreakdownProperties: &postgres.Jsonb{breakdownProps}, AlertLimit: 5, SetAlertLimit: true,
			Slack: true, SlackChannels: &postgres.Jsonb{RawMessage: slackChannelJson},
			Filter: []model.QueryProperty{
				{Entity: "event", Type: "numerical", Property: "$time_spent", Operator: "greaterThanOrEqual", LogicalOp: "AND", Value: "3000"},
			}}, agent.UUID, agent.UUID, false)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.Empty(t, errMsg)
		assert.NotNil(t, alert)

		event := &model.Event{EventNameId: eventName.ID, ProjectId: project.ID,
			UserId: agent.UUID, Timestamp: start.Unix(),
			UserProperties: &postgres.Jsonb{RawMessage: []byte(``)},
			Properties:     postgres.Jsonb{RawMessage: []byte(`{"$time_spent": "3500"}`)}}

		alerts, _, errCode := store.GetStore().MatchEventTriggerAlertWithTrackPayload(project.ID, eventName.ID, &event.Properties, event.UserProperties, nil, false)
		assert.Equal(t, http.StatusFound, errCode)
		assert.NotNil(t, alerts)

		event1 := &model.Event{EventNameId: eventName.ID, ProjectId: project.ID,
			UserId: agent.UUID, Timestamp: start.Unix(),
			UserProperties: &postgres.Jsonb{RawMessage: []byte(`{}`)},
			Properties:     postgres.Jsonb{RawMessage: []byte(`{"$time_spent": "3000"}`)}}

		alerts1, _, errCode := store.GetStore().MatchEventTriggerAlertWithTrackPayload(project.ID, eventName.ID, &event1.Properties, event1.UserProperties, nil, false)
		assert.Equal(t, http.StatusFound, errCode)
		assert.NotNil(t, alerts1)

		event2 := &model.Event{EventNameId: eventName.ID, ProjectId: project.ID,
			UserId: agent.UUID, Timestamp: start.Unix(),
			UserProperties: &postgres.Jsonb{RawMessage: []byte(`{}`)},
			Properties:     postgres.Jsonb{RawMessage: []byte(`{"$time_spent": "300"}`)}}

		alerts2, _, errCode := store.GetStore().MatchEventTriggerAlertWithTrackPayload(project.ID, eventName.ID, &event2.Properties, event2.UserProperties, nil, false)
		assert.Equal(t, http.StatusNotFound, errCode)
		assert.Nil(t, alerts2)
	})

	//Test for filter of numerical type and lesserThanOrEqual operator
	t.Run("MatchEventTriggerAlert:NumericalTypeLesserThanOrEqual", func(t *testing.T) {
		project, user, eventName, err := SetupProjectUserEventNameReturnDAO()
		assert.NotNil(t, eventName)
		assert.NotNil(t, project)
		assert.NotNil(t, user)
		assert.Nil(t, err)

		agent, errCode := SetupAgentReturnWithSlackIntegrationDAO(getRandomEmail(), "+1343545", project.ID)
		assert.NotNil(t, agent)
		assert.Equal(t, http.StatusCreated, errCode)

		//Test for successful CreateAlert
		alert, errCode, errMsg := store.GetStore().CreateEventTriggerAlert(agent.UUID, "", project.ID, &model.EventTriggerAlertConfig{
			Title: rName1, Event: eventName.Name, Message: "Remember", MessageProperty: &postgres.Jsonb{RawMessage: messagePropertyJson},
			DontRepeatAlerts: true, CoolDownTime: 1800, BreakdownProperties: &postgres.Jsonb{breakdownProps}, AlertLimit: 5, SetAlertLimit: true,
			Slack: true, SlackChannels: &postgres.Jsonb{RawMessage: slackChannelJson},
			Filter: []model.QueryProperty{
				{Entity: "event", Type: "numerical", Property: "$time_spent", Operator: "lesserThanOrEqual", LogicalOp: "AND", Value: "3000"},
			}}, agent.UUID, agent.UUID, false)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.Empty(t, errMsg)
		assert.NotNil(t, alert)

		event := &model.Event{EventNameId: eventName.ID, ProjectId: project.ID,
			UserId: agent.UUID, Timestamp: start.Unix(),
			UserProperties: &postgres.Jsonb{RawMessage: []byte(``)},
			Properties:     postgres.Jsonb{RawMessage: []byte(`{"$time_spent": "2500"}`)}}

		alerts, _, errCode := store.GetStore().MatchEventTriggerAlertWithTrackPayload(project.ID, eventName.ID, &event.Properties, event.UserProperties, nil, false)
		assert.Equal(t, http.StatusFound, errCode)
		assert.NotNil(t, alerts)

		event1 := &model.Event{EventNameId: eventName.ID, ProjectId: project.ID,
			UserId: agent.UUID, Timestamp: start.Unix(),
			UserProperties: &postgres.Jsonb{RawMessage: []byte(`{}`)},
			Properties:     postgres.Jsonb{RawMessage: []byte(`{"$time_spent": "3000"}`)}}

		alerts1, _, errCode := store.GetStore().MatchEventTriggerAlertWithTrackPayload(project.ID, eventName.ID, &event1.Properties, event1.UserProperties, nil, false)
		assert.Equal(t, http.StatusFound, errCode)
		assert.NotNil(t, alerts1)

		event2 := &model.Event{EventNameId: eventName.ID, ProjectId: project.ID,
			UserId: agent.UUID, Timestamp: start.Unix(),
			UserProperties: &postgres.Jsonb{RawMessage: []byte(`{}`)},
			Properties:     postgres.Jsonb{RawMessage: []byte(`{"$time_spent": "3500"}`)}}

		alerts2, _, errCode := store.GetStore().MatchEventTriggerAlertWithTrackPayload(project.ID, eventName.ID, &event2.Properties, event2.UserProperties, nil, false)
		assert.Equal(t, http.StatusNotFound, errCode)
		assert.Nil(t, alerts2)
	})

	//Test for filter of dateTime type and from-to operator
	// t.Run("MatchEventTriggerAlert:DatetimeType", func(t *testing.T) {
	// 	project, user, eventName, err := SetupProjectUserEventNameReturnDAO()
	// 	assert.NotNil(t, eventName)
	// 	assert.NotNil(t, project)
	// 	assert.NotNil(t, user)
	// 	assert.Nil(t, err)

	// 	//Test for successful CreateAlert
	// 	alert, errCode, errMsg := store.GetStore().CreateEventTriggerAlert(agent.UUID, "", project.ID, &model.EventTriggerAlertConfig{
	// 		Title: rName1, Event: eventName.Name, Message: "Remember", MessageProperty: &postgres.Jsonb{RawMessage: messagePropertyJson},
	// 		DontRepeatAlerts: true, CoolDownTime: 1800, BreakdownProperties: &postgres.Jsonb{breakdownProps}, AlertLimit: 5, SetAlertLimit: true,
	// 		Slack: true, SlackChannels: &postgres.Jsonb{RawMessage: slackChannelJson},
	// 		Filter: []model.QueryProperty{
	// 			{Entity: "user", Type: "datetime", Property: "$first_seen", Operator: "equals", LogicalOp: "AND", Value: `{"fr":"12345678", "to":"67854321"}`},
	// 		}})
	// 	assert.Equal(t, http.StatusCreated, errCode)
	// 	assert.Empty(t, errMsg)
	// 	assert.NotNil(t, alert)

	// 	event := &model.Event{EventNameId: eventName.ID, ProjectId: project.ID,
	// 		UserId: agent.UUID, Timestamp: start.Unix(),
	// 		UserProperties: &postgres.Jsonb{RawMessage: []byte(`{"$first_seen":"45678912"}`)},
	// 		Properties:     postgres.Jsonb{RawMessage: []byte(`{"$time_spent": "3500"}`)}}

	// 	alerts, errCode := store.GetStore().MatchEventTriggerAlertWithTrackPayload(project.ID, eventName.ID, &event.Properties, event.UserProperties)
	// 	assert.Equal(t, http.StatusFound, errCode)
	// 	assert.NotNil(t, alerts)

	// 	event1 := &model.Event{EventNameId: eventName.ID, ProjectId: project.ID,
	// 		UserId: agent.UUID, Timestamp: start.Unix(),
	// 		UserProperties: &postgres.Jsonb{RawMessage: []byte(`{"$first_seen":"78912999"}`)},
	// 		Properties:     postgres.Jsonb{RawMessage: []byte(`{"$time_spent": "3000"}`)}}

	// 	alerts1, errCode := store.GetStore().MatchEventTriggerAlertWithTrackPayload(project.ID, eventName.ID, &event1.Properties, event1.UserProperties)
	// 	assert.Equal(t, http.StatusNotFound, errCode)
	// 	assert.Nil(t, alerts1)
	// })

	//Test for combination filters
	t.Run("MatchEventTriggerAlert:CombinationFilters", func(t *testing.T) {
		project, user, eventName, err := SetupProjectUserEventNameReturnDAO()
		assert.NotNil(t, eventName)
		assert.NotNil(t, project)
		assert.NotNil(t, user)
		assert.Nil(t, err)

		agent, errCode := SetupAgentReturnWithSlackIntegrationDAO(getRandomEmail(), "+1343545", project.ID)
		assert.NotNil(t, agent)
		assert.Equal(t, http.StatusCreated, errCode)

		//Test for successful CreateAlert
		alert, errCode, errMsg := store.GetStore().CreateEventTriggerAlert(agent.UUID, "", project.ID, &model.EventTriggerAlertConfig{
			Title: rName1, Event: eventName.Name, Message: "Remember", MessageProperty: &postgres.Jsonb{RawMessage: messagePropertyJson},
			DontRepeatAlerts: true, CoolDownTime: 1800, BreakdownProperties: &postgres.Jsonb{breakdownProps}, AlertLimit: 5, SetAlertLimit: true,
			Slack: true, SlackChannels: &postgres.Jsonb{RawMessage: slackChannelJson},
			Filter: []model.QueryProperty{
				{Entity: "user", Type: "categorical", Property: "$Salesforce_Industry", Operator: "contains", LogicalOp: "AND", Value: "tech"},
				{Entity: "user", Type: "categorical", Property: "$country", Operator: "equals", LogicalOp: "AND", Value: "US"},
				{Entity: "event", Type: "numerical", Property: "$time_spent", Operator: "greaterThanOrEqual", LogicalOp: "AND", Value: "3000"},
			}}, agent.UUID, agent.UUID, false)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.Empty(t, errMsg)
		assert.NotNil(t, alert)

		event := &model.Event{EventNameId: eventName.ID, ProjectId: project.ID,
			UserId: agent.UUID, Timestamp: start.Unix(),
			UserProperties: &postgres.Jsonb{RawMessage: []byte(`{"$country":"US", "$Salesforce_Industry":"EdTech"}`)},
			Properties:     postgres.Jsonb{RawMessage: []byte(`{"$time_spent": "3500"}`)}}

		alerts, _, errCode := store.GetStore().MatchEventTriggerAlertWithTrackPayload(project.ID, eventName.ID, &event.Properties, event.UserProperties, nil, false)
		assert.Equal(t, http.StatusFound, errCode)
		assert.NotNil(t, alerts)

		event1 := &model.Event{EventNameId: eventName.ID, ProjectId: project.ID,
			UserId: agent.UUID, Timestamp: start.Unix(),
			UserProperties: &postgres.Jsonb{RawMessage: []byte(`{"$country":"US", "$Salesforce_Industry":"Healthcare"}`)},
			Properties:     postgres.Jsonb{RawMessage: []byte(`{"$time_spent": "3000"}`)}}

		alerts1, _, errCode := store.GetStore().MatchEventTriggerAlertWithTrackPayload(project.ID, eventName.ID, &event1.Properties, event1.UserProperties, nil, false)
		assert.Equal(t, http.StatusNotFound, errCode)
		assert.Nil(t, alerts1)
	})

	// t.Run("MatchEventTriggerAlert:CombinationFilters2", func(t *testing.T) {
	// 	project, user, eventName, err := SetupProjectUserEventNameReturnDAO()
	// 	assert.NotNil(t, eventName)
	// 	assert.NotNil(t, project)
	// 	assert.NotNil(t, user)
	// 	assert.Nil(t, err)

	// 	//Test for successful CreateAlert
	// 	alert, errCode, errMsg := store.GetStore().CreateEventTriggerAlert(agent.UUID, "", project.ID, &model.EventTriggerAlertConfig{
	// 		Title: rName1, Event: eventName.Name, Message: "Remember", MessageProperty: &postgres.Jsonb{RawMessage: messagePropertyJson}, DontRepeatAlerts: true, CoolDownTime: 1800,
	// BreakdownProperties: &postgres.Jsonb{breakdownProps}, AlertLimit: 5, SetAlertLimit: true,
	// 		Slack: true, SlackChannels: &postgres.Jsonb{RawMessage: slackChannelJson},
	// 		Filter: []model.QueryProperty{
	// 			{Entity: "user", Type: "numerical", Property: "$clicks", Operator: "lessThan", LogicalOp: "OR", Value: "350"},
	// 			{Entity: "user", Type: "categorical", Property: "$country", Operator: "contains", LogicalOp: "AND", Value: "US"},
	// 		}})
	// 	assert.Equal(t, http.StatusCreated, errCode)
	// 	assert.Empty(t, errMsg)
	// 	assert.NotNil(t, alert)

	// 	event := &model.Event{EventNameId: eventName.ID, ProjectId: project.ID,
	// 		UserId: agent.UUID, Timestamp: start.Unix(),
	// 		UserProperties: &postgres.Jsonb{RawMessage: []byte(`{"$country":"USA", "$Salesforce_Industry":"EdTech"}`)},
	// 		Properties:     postgres.Jsonb{RawMessage: []byte(`{"$time_spent": "3000"}`)}}

	// 	alerts, errCode := store.GetStore().MatchEventTriggerAlertWithTrackPayload(project.ID, eventName.ID, &event.Properties, event.UserProperties)
	// 	assert.Equal(t, http.StatusFound, errCode)
	// 	assert.NotNil(t, alerts)

	// 	event1 := &model.Event{EventNameId: eventName.ID, ProjectId: project.ID,
	// 		UserId: agent.UUID, Timestamp: start.Unix(),
	// 		UserProperties: &postgres.Jsonb{RawMessage: []byte(`{"$country":"Ukraine", "$Salesforce_Industry":"Healthcare"}`)},
	// 		Properties:     postgres.Jsonb{RawMessage: []byte(`{"$time_spent": "300"}`)}}

	// 	alerts1, errCode := store.GetStore().MatchEventTriggerAlertWithTrackPayload(project.ID, eventName.ID, &event1.Properties, event1.UserProperties)
	// 	assert.Equal(t, http.StatusNotFound, errCode)
	// 	assert.Nil(t, alerts1)
	// })

}

func TestEditEventTriggerAlertHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, agent, err := SetupProjectWithSlackIntegratedAgentDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)
	assert.NotNil(t, agent)

	slackChannel := []model.SlackChannel{
		{
			Name:      "channel1",
			Id:        U.GetUUID(),
			IsPrivate: false,
		},
	}
	slackChannelJson, err := json.Marshal(slackChannel)
	assert.NotNil(t, slackChannelJson)
	assert.Nil(t, err)

	breakdownProps, err := json.Marshal(`[
        {
            "pr": "user_id",
            "en": "user",
            "pty": "categorical",
            "ena": "$hubspot_contact_created",
            "eni": 1
        }
    ]`)
	assert.Nil(t, err)

	messageProperty := []model.QueryGroupByProperty{
		{
			Entity:         "event",
			Property:       "$page_title",
			Type:           "categorical",
			EventNameIndex: 1,
		},
		{
			Entity:         "user",
			Property:       "$country",
			Type:           "categorical",
			EventNameIndex: 2,
		},
	}
	messagePropertyJson, err := json.Marshal(messageProperty)
	assert.NotNil(t, messagePropertyJson)
	assert.Nil(t, err)

	rName1 := U.RandomString(5)

	t.Run("EditEventTriggerAlert:Valid", func(t *testing.T) {
		//Test for successful CreateAlert
		alert, errCode, errMsg := store.GetStore().CreateEventTriggerAlert(agent.UUID, "", project.ID, &model.EventTriggerAlertConfig{
			Title: rName1, Event: rName1, Message: "Remember", MessageProperty: &postgres.Jsonb{RawMessage: messagePropertyJson},
			DontRepeatAlerts: true, CoolDownTime: 1800, BreakdownProperties: &postgres.Jsonb{breakdownProps}, AlertLimit: 5, SetAlertLimit: true,
			Slack: true, SlackChannels: &postgres.Jsonb{RawMessage: slackChannelJson}, Filter: []model.QueryProperty{
				{Entity: "event", Type: "categorical", Property: "$country", Operator: "equals", LogicalOp: "AND", Value: "US"},
			}}, agent.UUID, agent.UUID, false)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.Empty(t, errMsg)
		assert.NotNil(t, alert)

		query := &model.EventTriggerAlertConfig{
			Title: "NewTitle", Event: rName1, Message: "Message Changed", MessageProperty: &postgres.Jsonb{RawMessage: messagePropertyJson},
			DontRepeatAlerts: true, CoolDownTime: 1800, BreakdownProperties: &postgres.Jsonb{breakdownProps}, AlertLimit: 5, SetAlertLimit: true,
			Slack: true, SlackChannels: &postgres.Jsonb{RawMessage: slackChannelJson}, Filter: []model.QueryProperty{
				{Entity: "event", Type: "categorical", Property: "$country", Operator: "equals", LogicalOp: "AND", Value: "US"},
			}}
		queryJson, err := json.Marshal(query)
		assert.Nil(t, err)

		w := sendEditEventTriggerAlertRequest(r, project.ID, alert.ID, agent, &postgres.Jsonb{RawMessage: queryJson})
		assert.Equal(t, http.StatusAccepted, w.Code)
		assert.NotNil(t, w)
	})
}

func sendEditEventTriggerAlertRequest(r *gin.Engine, projectId int64, id string, agent *model.Agent, query *postgres.Jsonb) *httptest.ResponseRecorder {

	cookieData, _ := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	rb := C.NewRequestBuilderWithPrefix(http.MethodPut, fmt.Sprintf("/projects/%d/v1/eventtriggeralert/%s", projectId, id)).
		WithPostParams(query).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, _ := rb.Build()

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
