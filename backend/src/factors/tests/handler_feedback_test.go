package tests

import (
	"encoding/json"
	"factors/model/model"
	"factors/model/store"
	"net/http"
	"testing"

	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
)

func TestPostFeedback(t *testing.T) {
	email := getRandomEmail()
	agent, errCode := SetupAgentReturnDAO(email, "+13425356")
	assert.Equal(t, http.StatusCreated, errCode)
	WIProperty := model.WeeklyInsightsProperty{
		Key:         "$browser",
		Value:       "Chrome",
		QueryID:     123,
		Type:        "Conversion",
		Order:       2,
		Entity:      "up",
		IsIncreased: true,
		Date:        "",
	}
	propertyBytes, err := json.Marshal(WIProperty)
	assert.Nil(t, err)
	property := postgres.Jsonb{RawMessage: json.RawMessage(propertyBytes)}
	status, errmsg := store.GetStore().PostFeedback(1, agent.UUID, "WeeklyInsights", &property, 2)
	assert.Equal(t, http.StatusCreated, status)
	assert.Equal(t, "", errmsg)

}
