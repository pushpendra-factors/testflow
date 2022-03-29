package tests

import (
	"encoding/json"
	"factors/model/model"
	"factors/model/store"
	"net/http"
	"testing"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
)

func TestCRMCreateData(t *testing.T) {
	project, _, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	t.Run("CreateCRMUser", func(t *testing.T) {
		user1Properties := postgres.Jsonb{json.RawMessage(`{"name":"abc","city":"xyz"}`)}

		user1 := &model.CRMUser{
			ProjectID:  project.ID,
			Source:     model.CRM_SOURCE_HUBSPOT,
			Type:       1,
			ID:         "123",
			Properties: &user1Properties,
			Timestamp:  time.Now().Unix(),
		}
		status, err := store.GetStore().CreateCRMUser(user1)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusCreated, status)
		assert.Equal(t, user1.Action, model.CRMActionCreated)

		status, err = store.GetStore().CreateCRMUser(user1)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusCreated, status)
		assert.Equal(t, user1.Action, model.CRMActionUpdated)

		status, err = store.GetStore().CreateCRMUser(user1)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusConflict, status)
		assert.Equal(t, user1.Action, model.CRMActionUpdated)

		user1.Timestamp = user1.Timestamp + 100
		status, err = store.GetStore().CreateCRMUser(user1)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusCreated, status)
		assert.Equal(t, user1.Action, model.CRMActionUpdated)
	},
	)

	t.Run("CreateCRMGroup", func(t *testing.T) {
		user1Properties := postgres.Jsonb{json.RawMessage(`{"company":"company1","city":"xyz"}`)}

		group1 := &model.CRMGroup{
			ProjectID:  project.ID,
			Source:     model.CRM_SOURCE_HUBSPOT,
			Type:       1,
			ID:         "123",
			Properties: &user1Properties,
			Timestamp:  time.Now().Unix(),
		}
		status, err := store.GetStore().CreateCRMGroup(group1)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusCreated, status)
		assert.Equal(t, group1.Action, model.CRMActionCreated)

		status, err = store.GetStore().CreateCRMGroup(group1)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusCreated, status)
		assert.Equal(t, group1.Action, model.CRMActionUpdated)

		status, err = store.GetStore().CreateCRMGroup(group1)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusConflict, status)
		assert.Equal(t, group1.Action, model.CRMActionUpdated)

		group1.Timestamp = group1.Timestamp + 100
		status, err = store.GetStore().CreateCRMGroup(group1)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusCreated, status)
		assert.Equal(t, group1.Action, model.CRMActionUpdated)
	},
	)

	t.Run("CreateCRMRelationship", func(t *testing.T) {

		relationship := &model.CRMRelationship{
			ProjectID: project.ID,
			Source:    model.CRM_SOURCE_HUBSPOT,
			FromType:  1,
			FromID:    "123",
			ToType:    2,
			ToID:      "234",
			Timestamp: time.Now().Unix(),
		}
		status, err := store.GetStore().CreateCRMRelationship(relationship)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusCreated, status)
		status, err = store.GetStore().CreateCRMRelationship(relationship)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusConflict, status)
	},
	)

	t.Run("CreateCRMActivity", func(t *testing.T) {
		activityProperties := postgres.Jsonb{json.RawMessage(`{"name":"abc","clicked":"true"}`)}
		activity := &model.CRMActivity{
			ProjectID:  project.ID,
			Source:     model.CRM_SOURCE_HUBSPOT,
			Name:       "test1",
			Type:       1,
			ActorType:  1,
			ActorID:    "123",
			Properties: &activityProperties,
			Timestamp:  time.Now().Unix(),
		}
		status, err := store.GetStore().CreateCRMActivity(activity)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusCreated, status)
		activity.ID = ""
		status, err = store.GetStore().CreateCRMActivity(activity)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusConflict, status)
		activity.Timestamp = activity.Timestamp + 100
		status, err = store.GetStore().CreateCRMActivity(activity)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusCreated, status)
	},
	)
}
