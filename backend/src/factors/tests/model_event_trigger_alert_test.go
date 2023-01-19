package tests

import (
	"encoding/json"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"net/http"
	"testing"

	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
)

func TestCreateEventTriggerAlert(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	slackChannel := model.SlackChannel{
		Name:      "channel1",
		Id:        U.GetUUID(),
		IsPrivate: true,
	}
	slackChannelJson, err := json.Marshal(slackChannel)
	assert.Nil(t, err)

	t.Run("CreateEventTriggerAlert:valid", func(t *testing.T) {
		rName1 := U.RandomString(5)
		alert, errCode, errMsg := store.GetStore().CreateEventTriggerAlert(agent.UUID, project.ID, &model.EventTriggerAlertConfig{
			Title: rName1, Event: rName1, Message: "Remember", MessageProperty: &postgres.Jsonb{}, RepeatAlerts: true, AlertLimit: 5, Notifications: true,
			Slack: true, SlackChannels: &postgres.Jsonb{RawMessage: slackChannelJson}, Filter: []model.QueryProperty{
				{Entity: "", Type: "categorical", Property: "campaign", Operator: "equals", LogicalOp: "AND"},
			}})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, alert)
		assert.Empty(t, errMsg)
	})

	t.Run("GetAllEventTriggerAlerts:valid", func(t *testing.T) {
		rName1 := U.RandomString(5)
		alert, errCode, errMsg := store.GetStore().CreateEventTriggerAlert(agent.UUID, project.ID, &model.EventTriggerAlertConfig{
			Title: rName1, Event: rName1, Message: "Remember", MessageProperty: &postgres.Jsonb{}, RepeatAlerts: true, AlertLimit: 5, Notifications: true,
			Slack: true, SlackChannels: &postgres.Jsonb{RawMessage: slackChannelJson}, Filter: []model.QueryProperty{
				{Entity: "", Type: "categorical", Property: "campaign", Operator: "equals", LogicalOp: "AND"},
			}})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, alert)
		assert.Empty(t, errMsg)

		alerts, errCode := store.GetStore().GetAllEventTriggerAlertsByProject(project.ID)
		assert.Equal(t, http.StatusFound, errCode)
		assert.NotNil(t, alerts)
	})

	t.Run("CreateEventTriggerAlert:Title already present:Invalid", func(t *testing.T) {
		rName1 := U.RandomString(5)
		alert, errCode, errMsg := store.GetStore().CreateEventTriggerAlert(agent.UUID, project.ID, &model.EventTriggerAlertConfig{
			Title: rName1, Event: rName1, Message: "Remember", MessageProperty: &postgres.Jsonb{}, RepeatAlerts: true, AlertLimit: 5, Notifications: true,
			Slack: true, SlackChannels: &postgres.Jsonb{RawMessage: slackChannelJson}, Filter: []model.QueryProperty{
				{Entity: "", Type: "categorical", Property: "campaign", Operator: "equals", LogicalOp: "AND"},
			}})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, alert)
		assert.Empty(t, errMsg)

		alert, errCode, errMsg = store.GetStore().CreateEventTriggerAlert(agent.UUID, project.ID, &model.EventTriggerAlertConfig{
			Title: rName1, Event: rName1, Message: "Remember", MessageProperty: &postgres.Jsonb{}, RepeatAlerts: true, AlertLimit: 5, Notifications: true,
			Slack: true, SlackChannels: &postgres.Jsonb{RawMessage: slackChannelJson}, Filter: []model.QueryProperty{
				{Entity: "", Type: "categorical", Property: "campaign", Operator: "equals", LogicalOp: "AND"},
			}})
		assert.Equal(t, http.StatusConflict, errCode)
		assert.Nil(t, alert)
		assert.NotEqual(t, "", errMsg)
	})
}

func TestDeleteEventTriggerAlert(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	slackChannel := model.SlackChannel{
		Name:      "channel1",
		Id:        U.GetUUID(),
		IsPrivate: true,
	}
	slackChannelJson, err := json.Marshal(slackChannel)
	assert.Nil(t, err)

	rName1 := U.RandomString(5)
	alert, errCode, errMsg := store.GetStore().CreateEventTriggerAlert(agent.UUID, project.ID, &model.EventTriggerAlertConfig{
		Title: rName1, Event: rName1, Message: "Remember", MessageProperty: &postgres.Jsonb{}, RepeatAlerts: true, AlertLimit: 5, Notifications: true,
		Slack: true, SlackChannels: &postgres.Jsonb{RawMessage: slackChannelJson}, Filter: []model.QueryProperty{
			{Entity: "", Type: "categorical", Property: "campaign", Operator: "equals", LogicalOp: "AND"},
		}})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Empty(t, errMsg)
	assert.NotNil(t, alert)

	t.Run("DeleteEventTriggerAlert:valid", func(t *testing.T) {
		errCode, errMsg = store.GetStore().DeleteEventTriggerAlert(project.ID, alert.ID)
		assert.Equal(t, http.StatusAccepted, errCode)
		assert.Empty(t, errMsg)
	})
}
