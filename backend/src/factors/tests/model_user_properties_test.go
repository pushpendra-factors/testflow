package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	C "factors/config"
	M "factors/model"
	U "factors/util"

	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
)

func TestMergeUserPropertiesForUserID(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	customerUserID := getRandomEmail()
	user1, _ := M.CreateUser(&M.User{
		ID:             U.GetUUID(),
		ProjectId:      project.ID,
		CustomerUserId: customerUserID,
		Properties: postgres.Jsonb{RawMessage: json.RawMessage([]byte(`{
			"country": "india",
			"age": 30,
			"paid": true,
			"gender": "m",
			"$initial_campaign": "campaign1",
			"$page_count": 10,
			"$session_spent_time": 2.2}`,
		))},
	})

	user2, _ := M.CreateUser(&M.User{
		ID:             U.GetUUID(),
		ProjectId:      project.ID,
		CustomerUserId: customerUserID,
		Properties: postgres.Jsonb{RawMessage: json.RawMessage([]byte(`{
			"country": "canada",
			"age": 30,
			"paid": false,
			"$initial_campaign": "campaign2",
			"$page_count": 15,
			"$session_spent_time": 4.4}`,
		))},
	})

	// Merge is not enabled for project. Should not merge.
	_, errCode := M.MergeUserPropertiesForUserID(project.ID, user1.ID, postgres.Jsonb{}, "", U.TimeNowUnix(), false, true)
	assert.Equal(t, http.StatusNotAcceptable, errCode)
	user1DB, _ := M.GetUser(project.ID, user1.ID)
	user1PropertiesDB, _ := U.DecodePostgresJsonb(&user1DB.Properties)
	user2DB, _ := M.GetUser(project.ID, user2.ID)
	user2PropertiesDB, _ := U.DecodePostgresJsonb(&user2DB.Properties)
	// Both user properties must not be same.
	assert.NotEqual(t, user1PropertiesDB, user2PropertiesDB)
	// User property id must not change.
	assert.Equal(t, user1.PropertiesId, user1DB.PropertiesId)
	assert.Equal(t, user2.PropertiesId, user2DB.PropertiesId)

	// Enable merge in config.
	(*C.GetConfig()).MergeUspProjectIds = fmt.Sprint(project.ID)

	// Dryrun is set to true. Should not merge.
	_, errCode = M.MergeUserPropertiesForUserID(project.ID, user1.ID, postgres.Jsonb{}, "", U.TimeNowUnix(), true, true)
	assert.Equal(t, http.StatusNotModified, errCode)
	user1DB, _ = M.GetUser(project.ID, user1.ID)
	user1PropertiesDB, _ = U.DecodePostgresJsonb(&user1DB.Properties)
	user2DB, _ = M.GetUser(project.ID, user2.ID)
	user2PropertiesDB, _ = U.DecodePostgresJsonb(&user2DB.Properties)
	// Both user properties must not be same.
	assert.NotEqual(t, user1PropertiesDB, user2PropertiesDB)
	// User property id must not change.
	assert.Equal(t, user1.PropertiesId, user1DB.PropertiesId)
	assert.Equal(t, user2.PropertiesId, user2DB.PropertiesId)

	_, errCode = M.MergeUserPropertiesForUserID(project.ID, user1.ID, postgres.Jsonb{}, "", U.TimeNowUnix(), false, true)
	assert.Equal(t, http.StatusCreated, errCode)
	user1DB, _ = M.GetUser(project.ID, user1.ID)
	user1PropertiesDB, _ = U.DecodePostgresJsonb(&user1DB.Properties)
	user2DB, _ = M.GetUser(project.ID, user2.ID)
	user2PropertiesDB, _ = U.DecodePostgresJsonb(&user2DB.Properties)

	// Both user properties must be same.
	assert.Equal(t, user1PropertiesDB, user2PropertiesDB)
	// User property id must not be same after merge.
	assert.NotEqual(t, user1.PropertiesId, user1DB.PropertiesId)
	assert.NotEqual(t, user2.PropertiesId, user2DB.PropertiesId)

	// Property country must be canada and paid must be false.
	assert.Equal(t, "canada", (*user1PropertiesDB)["country"])
	assert.Equal(t, false, (*user1PropertiesDB)["paid"])
	assert.Equal(t, float64(30), (*user1PropertiesDB)["age"])
	// All properties must be present.
	for _, prop := range [...]string{"country", "age", "paid", "gender", "$initial_campaign", "$page_count", "$session_spent_time"} {
		_, found1 := (*user1PropertiesDB)[prop]
		assert.True(t, found1)
		_, found2 := (*user2PropertiesDB)[prop]
		assert.True(t, found2)
	}

	// Initial properties must not be overwritten by older values.
	assert.Equal(t, "campaign1", (*user2PropertiesDB)["$initial_campaign"])

	// Properties that are to be added should be sum of values.
	assert.Equal(t, float64(25), (*user1PropertiesDB)["$page_count"])
	assert.Equal(t, float64(25), (*user2PropertiesDB)["$page_count"])

	// Check if floating points are added correctly.
	// By default, 2.2 + 4.4 results in 6.6000000000000005.
	assert.Equal(t, float64(6.6), (*user1PropertiesDB)["$session_spent_time"])
	assert.Equal(t, float64(6.6), (*user2PropertiesDB)["$session_spent_time"])

	// Running merge again for the same customerID should not update user_properties.
	_, errCode = M.MergeUserPropertiesForUserID(project.ID, user1.ID, postgres.Jsonb{}, "", U.TimeNowUnix(), false, true)
	assert.Equal(t, http.StatusNotModified, errCode) // StatusNotModified.
	user1DBRetry, _ := M.GetUser(project.ID, user1.ID)
	user2DBRetry, _ := M.GetUser(project.ID, user2.ID)
	assert.Equal(t, user1DB.PropertiesId, user1DBRetry.PropertiesId)
	assert.Equal(t, user2DB.PropertiesId, user2DBRetry.PropertiesId)
}

func TestMergeUserPropertiesForProjectID(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	(*C.GetConfig()).MergeUspProjectIds = fmt.Sprint(project.ID)

	customerUserID := getRandomEmail()
	user1, _ := M.CreateUser(&M.User{
		ID:             U.GetUUID(),
		ProjectId:      project.ID,
		CustomerUserId: customerUserID,
		Properties: postgres.Jsonb{RawMessage: json.RawMessage([]byte(`{
			"country": "india",
			"age": 30,
			"paid": true,
			"gender": "m",
			"$initial_campaign": "campaign1",
			"$page_count": 10,
			"$session_spent_time": 2.2}`,
		))},
	})

	user2, _ := M.CreateUser(&M.User{
		ID:             U.GetUUID(),
		ProjectId:      project.ID,
		CustomerUserId: customerUserID,
		Properties: postgres.Jsonb{RawMessage: json.RawMessage([]byte(`{
			"country": "canada",
			"age": 30,
			"paid": false,
			"$initial_campaign": "campaign2",
			"$page_count": 15,
			"$session_spent_time": 4.4}`,
		))},
	})

	errCode := M.MergeUserPropertiesForProjectID(project.ID, false)
	assert.Equal(t, http.StatusCreated, errCode)
	user1DB, _ := M.GetUser(project.ID, user1.ID)
	user1PropertiesDB, _ := U.DecodePostgresJsonb(&user1DB.Properties)
	user2DB, _ := M.GetUser(project.ID, user2.ID)
	user2PropertiesDB, _ := U.DecodePostgresJsonb(&user2DB.Properties)

	// Both user properties must be same.
	assert.Equal(t, user1PropertiesDB, user2PropertiesDB)
	assert.NotEqual(t, user1.PropertiesId, user1DB.PropertiesId)
	assert.NotEqual(t, user2.PropertiesId, user2DB.PropertiesId)
}
