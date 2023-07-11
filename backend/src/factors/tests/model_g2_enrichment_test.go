package tests

import (
	"encoding/json"
	"factors/model/model"
	"factors/model/store"
	SDK "factors/sdk"
	U "factors/util"
	"net/http"
	"testing"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
)

func TestG2Enrichment(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	value := postgres.Jsonb{RawMessage: json.RawMessage([]byte(`{"category_ids":[],"company_url":"https://data.g2.com/api/v1/ahoy/remote-event-streams/test/company","created_at":"2023-05-20T11:38:55.414-05:00","id":"12345","organization":"Collective Coffee Roasters","product_ids":["hiver"],"tag":"ad.category_competitor.product_left_sidebar","time":"2023-05-20T11:38:52.000-05:00","timestamp":1684600732,"title":"GrooveHQ Reviews 2023: Details, Pricing, & Features | G2","type":"remote_event_streams","url":"xyz.com","user_location":{"city":"Elko","country":"United States","country_code":"US","state":"Nevada","state_code":"NV"},"visitor_id":"abcd-1214"}`))}
	timestamp := U.UnixTimeBeforeDuration(time.Hour * 1)
	g2Document := model.G2Document{
		ProjectID: project.ID,
		ID:        "12345",
		Type:      1,
		TypeAlias: "event_stream",
		Timestamp: timestamp,
		Value:     &value,
		Synced:    false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	errCode := store.GetStore().CreateMultipleG2Document([]model.G2Document{g2Document})
	assert.Equal(t, http.StatusCreated, errCode)

	g2Documents, errCode := store.GetStore().GetG2DocumentsForGroupUserCreation(project.ID)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, 1, len(g2Documents))

	userPropertiesMap := U.PropertiesMap{
		U.G2_COMPANY_ID:      "54321",
		U.G2_COUNTRY:         "US",
		U.G2_DOMAIN:          "abc.com",
		U.G2_LEGAL_NAME:      "Check",
		U.G2_NAME:            "Test",
		U.G2_EMPLOYEES:       50,
		U.G2_EMPLOYEES_RANGE: "50-100",
	}
	userID, errCode := SDK.TrackGroupWithDomain(project.ID, U.GROUP_NAME_G2, "abc.com", userPropertiesMap, g2Document.Timestamp)
	assert.Equal(t, http.StatusOK, errCode)
	assert.NotEqual(t, "", userID)

	user, errCode := store.GetStore().GetUser(project.ID, userID)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, model.UserSourceG2, *(user.Source))

	eventNameG2All, errCode := store.GetStore().CreateOrGetUserCreatedEventName(&model.EventName{
		ProjectId: project.ID,
		Name:      U.GROUP_EVENT_NAME_G2_ALL,
	})
	assert.Equal(t, http.StatusCreated, errCode)
	allPageviewEvent := model.Event{
		EventNameId: eventNameG2All.ID,
		Timestamp:   timestamp,
		ProjectId:   project.ID,
		UserId:      userID,
	}
	_, errCode = store.GetStore().CreateEvent(&allPageviewEvent)
	assert.Equal(t, http.StatusCreated, errCode)

	err = store.GetStore().UpdateG2GroupUserCreationDetails(g2Document)
	assert.Nil(t, err)

	g2Documents, errCode = store.GetStore().GetG2DocumentsForGroupUserCreation(project.ID)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, 0, len(g2Documents))
}
