package tests

import (
	"encoding/json"
	"factors/model/model"
	"factors/model/store"
	"fmt"
	"net/http"
	"testing"

	C "factors/config"
	U "factors/util"

	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
)

func TestMergeUserPropertiesForUserID(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	timestamp := U.TimeNowUnix() - 500

	// Merge should be done on create itself as both users
	// have same customerUserID.
	customerUserID := getRandomEmail()
	timestamp = timestamp + 1
	user1, _ := store.GetStore().CreateUser(&model.User{
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
			"$session_spent_time": 2.2,
			"$latest_medium": "google",
			"$hubspot_contact_lead_guid": "lead-guid1"}`,
		))},
		JoinTimestamp: timestamp,
	})

	timestamp = timestamp + 1
	user2, _ := store.GetStore().CreateUser(&model.User{
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
			"$user_agent": "browser user agent",
			"$latest_medium": "",
			"$hubspot_contact_lead_guid": "lead-guid2"}`, // Empty. Should not overwrite.
		))},
		JoinTimestamp: timestamp,
	})

	user1DB, _ := store.GetStore().GetUser(project.ID, user1.ID)
	user1PropertiesDB, _ := U.DecodePostgresJsonb(&user1DB.Properties)
	user2DB, _ := store.GetStore().GetUser(project.ID, user2.ID)
	user2PropertiesDB, _ := U.DecodePostgresJsonb(&user2DB.Properties)

	if C.IsUserPropertiesTableWriteDeprecated(project.ID) {
		// user.properties_id should be empty as user_properties_table is deprecated.
		assert.Empty(t, user1DB.PropertiesId)
		assert.Empty(t, user2DB.PropertiesId)
	} else {
		assert.NotEmpty(t, user1DB.PropertiesId)
		assert.NotEmpty(t, user2DB.PropertiesId)
	}

	// Property country must be canada and paid must be false.
	assert.Equal(t, "canada", (*user1PropertiesDB)["country"])
	assert.Equal(t, false, (*user1PropertiesDB)["paid"])
	assert.Equal(t, float64(30), (*user1PropertiesDB)["age"])
	assert.Equal(t, "browser user agent", (*user1PropertiesDB)["$user_agent"])
	// Hubspot contact lead guid should not be considered on user_properties merge.
	assert.Equal(t, "lead-guid1", (*user1PropertiesDB)[model.UserPropertyHubspotContactLeadGUID])
	assert.Equal(t, "lead-guid2", (*user2PropertiesDB)[model.UserPropertyHubspotContactLeadGUID])

	// Remove the skipped properties on merge to check equality of others.
	for _, k := range model.UserPropertiesToSkipOnMergeByCustomerUserID {
		delete(*user2PropertiesDB, k)
		delete(*user1PropertiesDB, k)
	}
	// Both user properties must be same.
	assert.Equal(t, user1PropertiesDB, user2PropertiesDB)

	// All properties must be present.
	for _, prop := range [...]string{"country", "age", "paid", "gender", "$initial_campaign", "$page_count",
		"$session_spent_time", "$user_agent", "$latest_medium"} {
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

	// Empty property value should not overwrite old values.
	assert.Equal(t, "google", (*user1PropertiesDB)["$latest_medium"])
	assert.Equal(t, "google", (*user2PropertiesDB)["$latest_medium"])

	// Running merge again for the same customerID should not update user_properties.
	timestamp = timestamp + 1
	_, _, errCode := store.GetStore().UpdateUserProperties(project.ID, user1.ID,
		&postgres.Jsonb{RawMessage: json.RawMessage([]byte(`{}`))}, timestamp)
	assert.Equal(t, http.StatusNotModified, errCode) // StatusNotModified.
	user1DBRetry, _ := store.GetStore().GetUser(project.ID, user1.ID)
	user1PropertiesDBRetry, _ := U.DecodePostgresJsonb(&user1DBRetry.Properties)
	user2DBRetry, _ := store.GetStore().GetUser(project.ID, user2.ID)
	user2PropertiesDBRetry, _ := U.DecodePostgresJsonb(&user2DBRetry.Properties)
	assert.Equal(t, user1DB.PropertiesId, user1DBRetry.PropertiesId)
	assert.Equal(t, user2DB.PropertiesId, user2DBRetry.PropertiesId)
	fmt.Println(user1PropertiesDBRetry, user2PropertiesDBRetry)

	// Updating one of the non addable properties. Should not increase the value of addable properties.
	for i := 0; i < 5; i++ {
		cityValue := U.RandomLowerAphaNumString(5)
		propertiesUpdate := postgres.Jsonb{RawMessage: json.RawMessage(
			[]byte(fmt.Sprintf(`{"city": "%s"}`, cityValue)))}
		timestamp = timestamp + 1
		_, _, errCode := store.GetStore().UpdateUserProperties(project.ID, user1.ID, &propertiesUpdate, timestamp)
		assert.Equal(t, http.StatusAccepted, errCode)

		user1DB, _ = store.GetStore().GetUser(project.ID, user1.ID)
		user1PropertiesDB, _ = U.DecodePostgresJsonb(&user1DB.Properties)
		user2DB, _ = store.GetStore().GetUser(project.ID, user2.ID)
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
		timestamp = timestamp + 1
		_, _, errCode := store.GetStore().UpdateUserProperties(project.ID, user1.ID, &propertiesUpdate, timestamp)
		assert.Equal(t, http.StatusAccepted, errCode)

		user1DB, _ = store.GetStore().GetUser(project.ID, user1.ID)
		user1PropertiesDB, _ = U.DecodePostgresJsonb(&user1DB.Properties)
		user2DB, _ = store.GetStore().GetUser(project.ID, user2.ID)
		user2PropertiesDB, _ = U.DecodePostgresJsonb(&user2DB.Properties)

		assert.Equal(t, float64(previousPageCount+float64(i+1)), (*user1PropertiesDB)["$page_count"])
		assert.Equal(t, float64(previousPageCount+float64(i+1)), (*user2PropertiesDB)["$page_count"])
		assert.Equal(t, float64(previousSessionSpentTime+float64(i)+0.5), (*user1PropertiesDB)["$session_spent_time"])
		assert.Equal(t, float64(previousSessionSpentTime+float64(i)+0.5), (*user2PropertiesDB)["$session_spent_time"])
	}

	// When a new non merged user is added, entire values must be added to all users.
	previousPageCount = (*user1PropertiesDB)["$page_count"].(float64)
	previousSessionSpentTime = (*user1PropertiesDB)["$session_spent_time"].(float64)
	timestamp = timestamp + 1
	user3, _ := store.GetStore().CreateUser(&model.User{
		ID:             U.GetUUID(),
		ProjectId:      project.ID,
		CustomerUserId: customerUserID,
		Properties: postgres.Jsonb{RawMessage: json.RawMessage([]byte(`{
			"$page_count": 15,
			"$session_spent_time": 4.5}`,
		))},
		JoinTimestamp: timestamp,
	})

	// Call merge on user3.
	timestamp = timestamp + 1
	_, _, errCode = store.GetStore().UpdateUserProperties(project.ID, user3.ID,
		&postgres.Jsonb{RawMessage: json.RawMessage([]byte(fmt.Sprintf(`{"%s": "%s"}`,
			U.RandomNumericString(4), U.RandomNumericString(4))))}, timestamp)
	user1DB, _ = store.GetStore().GetUser(project.ID, user1.ID)
	user1PropertiesDB, _ = U.DecodePostgresJsonb(&user1DB.Properties)
	user2DB, _ = store.GetStore().GetUser(project.ID, user2.ID)
	user2PropertiesDB, _ = U.DecodePostgresJsonb(&user2DB.Properties)
	user3DB, _ := store.GetStore().GetUser(project.ID, user3.ID)
	user3PropertiesDB, _ := U.DecodePostgresJsonb(&user3DB.Properties)
	assert.Equal(t, float64(previousPageCount+15), (*user1PropertiesDB)["$page_count"])
	assert.Equal(t, float64(previousPageCount+15), (*user2PropertiesDB)["$page_count"])
	assert.Equal(t, float64(previousPageCount+15), (*user3PropertiesDB)["$page_count"])
	assert.Equal(t, float64(previousSessionSpentTime+4.5), (*user1PropertiesDB)["$session_spent_time"])
	assert.Equal(t, float64(previousSessionSpentTime+4.5), (*user2PropertiesDB)["$session_spent_time"])
	assert.Equal(t, float64(previousSessionSpentTime+4.5), (*user3PropertiesDB)["$session_spent_time"])
}

func TestSanitizeAddTypeProperties(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	user1, _ := store.GetStore().CreateUser(&model.User{
		ID:        U.GetUUID(),
		ProjectId: project.ID,
	})

	user2, _ := store.GetStore().CreateUser(&model.User{
		ID:        U.GetUUID(),
		ProjectId: project.ID,
	})

	mergedProperty1 := map[string]interface{}{
		U.UP_SESSION_COUNT:    1000000000,
		U.UP_PAGE_COUNT:       1235342430000,
		U.UP_TOTAL_SPENT_TIME: 8462088321000000,
	}
	createEventWithTimestampByName(t, project, user1, "$session", U.TimeNowUnix())
	createEventWithTimestampByName(t, project, user1, "$session", U.TimeNowUnix())
	createEventWithTimestampByName(t, project, user2, "$session", U.TimeNowUnix())

	store.GetStore().SanitizeAddTypeProperties(project.ID, []model.User{*user1, *user2}, &mergedProperty1)
	assert.Equal(t, float64(3), mergedProperty1[U.UP_SESSION_COUNT])
	assert.True(t, mergedProperty1[U.UP_PAGE_COUNT].(float64) >= float64(3*1))
	assert.True(t, mergedProperty1[U.UP_PAGE_COUNT].(float64) <= float64(3*5))
	assert.True(t, mergedProperty1[U.UP_TOTAL_SPENT_TIME].(float64) >= float64(3*60))
	assert.True(t, mergedProperty1[U.UP_TOTAL_SPENT_TIME].(float64) <= float64(3*300))

	// If session count and page count are in acceptable range, do nothing even when spent time is high.
	mergedProperty2 := map[string]interface{}{
		U.UP_SESSION_COUNT:    999,
		U.UP_PAGE_COUNT:       999,
		U.UP_TOTAL_SPENT_TIME: 8462088321000000,
	}
	store.GetStore().SanitizeAddTypeProperties(project.ID, []model.User{*user1, *user2}, &mergedProperty2)
	assert.Equal(t, 999, mergedProperty2[U.UP_SESSION_COUNT])
	assert.Equal(t, 999, mergedProperty2[U.UP_PAGE_COUNT])
	assert.Equal(t, 8462088321000000, mergedProperty2[U.UP_TOTAL_SPENT_TIME])
}
