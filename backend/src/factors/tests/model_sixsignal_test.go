package tests

import (
	H "factors/handler"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSixSignalVisitorIdentificationQuery(t *testing.T) {

	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	startTimestamp := U.UnixTimeBeforeDuration(time.Hour * 1)
	stepTimestamp := startTimestamp

	createdUserID1, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID1)
	createdUserID2, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID2)
	createdUserID3, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID3)
	createdUserID4, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID4)

	/*
		user1: event s0 with property1
		user2: event s0 with property2
		user3: event s1 with property1
		user4: event s1 with property2
	*/

	//user1
	payload := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, "user_properties": {"$initial_source" : "%s", "$6Signal_name" : "%s", "$6Signal_country" : "%s", "$6Signal_domain" : "%s", "$6Signal_employee_range" : "%s", "$6Signal_revenue_range" : "%s", "$6Signal_industry" : "%s","$page_count" : "%s", "$initial_page_url" : "%s"},
	"event_properties":{"$campaign_id":%d,"$campaign":"%s","$channel":"%s","$session_spent_time":"%s"}}`, "s0", createdUserID1, stepTimestamp, "A", "Apxor", "India", "apxor.com", "20-49", "$1M - $5M", "Software and Technology", "25", "www.factors.ai", 1234, "HelloWorld", "Direct", "540")
	w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response := DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	//user2
	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, "user_properties": {"$initial_source" : "%s", "$6Signal_name" : "%s", "$6Signal_country" : "%s", "$6Signal_domain" : "%s", "$6Signal_employee_range" : "%s", "$6Signal_revenue_range" : "%s", "$6Signal_industry" : "%s","$page_count" : "%s", "$initial_page_url" : "%s"},
	"event_properties":{"$campaign_id":%d,"$campaign":"%s","$channel":"%s","$session_spent_time":"%s"}}`, "s0", createdUserID2, stepTimestamp+10, "B", "RazorQ", "India", "razorQ.com", "200-490", "$10M - $45M", "Software and Technology", "250", "www.factors.ai", 4321, "HelloIndia", "Referal", "50")
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	//user3
	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, "user_properties": {"$initial_source" : "%s", "$6Signal_name" : "%s", "$6Signal_country" : "%s", "$6Signal_domain" : "%s", "$6Signal_employee_range" : "%s", "$6Signal_revenue_range" : "%s", "$6Signal_industry" : "%s","$page_count" : "%s", "$initial_page_url" : "%s"},
	"event_properties":{"$campaign_id":%d,"$campaign":"%s","$channel":"%s","$session_spent_time":"%s"}}`, "s1", createdUserID3, stepTimestamp+20, "z", "Apxor", "India", "apxor.com", "20-49", "$1M - $5M", "Software and Technology", "235", "www.factors.ai", 4678, "HelloWorld", "Direct", "5220")
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	//user4
	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, "user_properties": {"$initial_source" : "%s", "$6Signal_name" : "%s", "$6Signal_country" : "%s", "$6Signal_domain" : "%s", "$6Signal_employee_range" : "%s", "$6Signal_revenue_range" : "%s", "$6Signal_industry" : "%s","$page_count" : "%s", "$initial_page_url" : "%s"},
	"event_properties":{"$campaign_id":%d,"$campaign":"%s","$channel":"%s","$session_spent_time":"%s"}}`, "s1", createdUserID4, stepTimestamp+30, "p", "RazorQ", "India", "razorQ.com", "200-490", "$10M - $45M", "Software and Technology", "2540", "www.factors.ai", 421, "HelloIndia", "Referal", "504")
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	t.Run("SixSignalVisitorIdentification", func(t *testing.T) {
		sixSignalQuery := model.SixSignalQuery{From: stepTimestamp - 50, To: stepTimestamp + 50, Timezone: "Asia/Kolkata"}
		result, errCode, _ := store.GetStore().ExecuteSixSignalQuery(project.ID, sixSignalQuery)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		fmt.Println(result)
	})

}

func TestSixSignalMonthlyMetering(t *testing.T) {

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	monthYear := U.GetCurrentMonthYear(U.TimeZoneStringIST)

	t.Run("TestingMeteringForValuesInDomainList", func(t *testing.T) {
		domainList := []string{"abc.com", "xyz.com", "", "xyz1.com", "a1.com", "a2.com", "a1.com",
			"a2.com", " ", "abc2.com", "abc.com"}
		for _, v := range domainList {
			if v != "" {
				err := model.SetSixSignalMonthlyUniqueEnrichmentCount(project.ID, v, U.TimeZoneStringIST)
				assert.Nil(t, err)
			}
		}
		count, err := model.GetSixSignalMonthlyUniqueEnrichmentCount(project.ID, monthYear)
		fmt.Println(count)
		assert.Nil(t, err)
		assert.NotNil(t, count)
	})

}
