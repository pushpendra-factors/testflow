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
			"$session_spent_time": 4.4,
			"$user_agent": "browser user agent"}`,
		))},
	})

	// Merge is not enabled for project. Should not merge.
	_, errCode := M.MergeUserPropertiesForUserID(project.ID, user1.ID, postgres.Jsonb{}, "", U.TimeNowUnix(), false, true)
	assert.Equal(t, http.StatusNotModified, errCode)
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
	assert.Equal(t, "browser user agent", (*user1PropertiesDB)["$user_agent"])
	// All properties must be present.
	for _, prop := range [...]string{"country", "age", "paid", "gender", "$initial_campaign", "$page_count", "$session_spent_time", "$user_agent"} {
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
	user1PropertiesDBRetry, _ := U.DecodePostgresJsonb(&user1DBRetry.Properties)
	user2DBRetry, _ := M.GetUser(project.ID, user2.ID)
	user2PropertiesDBRetry, _ := U.DecodePostgresJsonb(&user2DBRetry.Properties)
	assert.Equal(t, user1DB.PropertiesId, user1DBRetry.PropertiesId)
	assert.Equal(t, user2DB.PropertiesId, user2DBRetry.PropertiesId)
	fmt.Println(user1PropertiesDBRetry, user2PropertiesDBRetry)

	// Updating one of the non addable properties. Should not increase the value of addable properties.
	for i := 0; i < 5; i++ {
		cityValue := U.RandomLowerAphaNumString(5)
		propertiesUpdate := postgres.Jsonb{RawMessage: json.RawMessage(
			[]byte(fmt.Sprintf(`{"city": "%s"}`, cityValue)))}
		M.UpdateUserProperties(project.ID, user1.ID, &propertiesUpdate, U.TimeNowUnix())

		_, errCode = M.MergeUserPropertiesForUserID(project.ID, user1.ID, postgres.Jsonb{}, "", U.TimeNowUnix(), false, true)
		user1DB, _ = M.GetUser(project.ID, user1.ID)
		user1PropertiesDB, _ = U.DecodePostgresJsonb(&user1DB.Properties)
		user2DB, _ = M.GetUser(project.ID, user2.ID)
		user2PropertiesDB, _ = U.DecodePostgresJsonb(&user2DB.Properties)

		// City should have got update every time.
		assert.Equal(t, cityValue, (*user1PropertiesDB)["city"])
		assert.Equal(t, cityValue, (*user2PropertiesDB)["city"])
		// $page_count and $session_spent_time should remain same.
		assert.Equal(t, float64(25), (*user1PropertiesDB)["$page_count"])
		assert.Equal(t, float64(25), (*user2PropertiesDB)["$page_count"])
		assert.Equal(t, float64(6.6), (*user1PropertiesDB)["$session_spent_time"])
		assert.Equal(t, float64(6.6), (*user2PropertiesDB)["$session_spent_time"])
	}

	// If addable properties is updated, only difference should get added.
	previousPageCount := (*user1PropertiesDB)["$page_count"].(float64)
	previousSessionSpentTime := (*user1PropertiesDB)["$session_spent_time"].(float64)
	for i := 0; i < 5; i++ {
		// Old values for user1: $page_count = 10, $session_spent_time = 2.2. Sum: 25.
		// Old values for user2: $page_count = 15, $session_spent_time = 4.4. Sum: 6.6.
		propertiesUpdate := postgres.Jsonb{RawMessage: json.RawMessage(
			[]byte(fmt.Sprintf(`{"$page_count": %f, "$session_spent_time": %f}`,
				previousPageCount+float64(i+1), previousSessionSpentTime+float64(i)+0.5)))}
		M.UpdateUserProperties(project.ID, user1.ID, &propertiesUpdate, U.TimeNowUnix())

		_, errCode = M.MergeUserPropertiesForUserID(project.ID, user1.ID, postgres.Jsonb{}, "", U.TimeNowUnix(), false, true)
		user1DB, _ = M.GetUser(project.ID, user1.ID)
		user1PropertiesDB, _ = U.DecodePostgresJsonb(&user1DB.Properties)
		user2DB, _ = M.GetUser(project.ID, user2.ID)
		user2PropertiesDB, _ = U.DecodePostgresJsonb(&user2DB.Properties)

		assert.Equal(t, float64(previousPageCount+float64(i+1)), (*user1PropertiesDB)["$page_count"])
		assert.Equal(t, float64(previousPageCount+float64(i+1)), (*user2PropertiesDB)["$page_count"])
		assert.Equal(t, float64(previousSessionSpentTime+float64(i)+0.5), (*user1PropertiesDB)["$session_spent_time"])
		assert.Equal(t, float64(previousSessionSpentTime+float64(i)+0.5), (*user2PropertiesDB)["$session_spent_time"])
	}

	// When a new non merged user is added, entire values must be added to all users.
	previousPageCount = (*user1PropertiesDB)["$page_count"].(float64)
	previousSessionSpentTime = (*user1PropertiesDB)["$session_spent_time"].(float64)
	user3, _ := M.CreateUser(&M.User{
		ID:             U.GetUUID(),
		ProjectId:      project.ID,
		CustomerUserId: customerUserID,
		Properties: postgres.Jsonb{RawMessage: json.RawMessage([]byte(`{
			"$page_count": 15,
			"$session_spent_time": 4.5}`,
		))},
	})
	// Call merge on user3.
	_, errCode = M.MergeUserPropertiesForUserID(project.ID, user3.ID, postgres.Jsonb{}, "", U.TimeNowUnix(), false, true)
	user1DB, _ = M.GetUser(project.ID, user1.ID)
	user1PropertiesDB, _ = U.DecodePostgresJsonb(&user1DB.Properties)
	user2DB, _ = M.GetUser(project.ID, user2.ID)
	user2PropertiesDB, _ = U.DecodePostgresJsonb(&user2DB.Properties)
	user3DB, _ := M.GetUser(project.ID, user3.ID)
	user3PropertiesDB, _ := U.DecodePostgresJsonb(&user3DB.Properties)
	assert.Equal(t, float64(previousPageCount+15), (*user1PropertiesDB)["$page_count"])
	assert.Equal(t, float64(previousPageCount+15), (*user2PropertiesDB)["$page_count"])
	assert.Equal(t, float64(previousPageCount+15), (*user3PropertiesDB)["$page_count"])
	assert.Equal(t, float64(previousSessionSpentTime+4.5), (*user1PropertiesDB)["$session_spent_time"])
	assert.Equal(t, float64(previousSessionSpentTime+4.5), (*user2PropertiesDB)["$session_spent_time"])
	assert.Equal(t, float64(previousSessionSpentTime+4.5), (*user3PropertiesDB)["$session_spent_time"])
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
