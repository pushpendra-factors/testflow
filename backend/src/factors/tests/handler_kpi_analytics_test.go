package tests

import (
	"encoding/json"
	C "factors/config"
	H "factors/handler"
	"factors/handler/helpers"
	v1 "factors/handler/v1"
	IntHubspot "factors/integration/hubspot"
	"factors/model/model"
	M "factors/model/model"
	"factors/model/store"
	TaskSession "factors/task/session"
	U "factors/util"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestChannelGroupDashboardUnitForTimeZone - To add this as well.
func TestKpiAnalytics(t *testing.T) {
	a := gin.Default()
	H.InitAppRoutes(a)

	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"

	r2 := gin.Default()
	H.InitDataServiceRoutes(r2)
	model.SetSmartPropertiesReservedNames()

	project, customerAccountID, agent, statusCode := createProjectAndAddAdwordsDocument(t, r2)
	log.Warn(customerAccountID)
	if statusCode != http.StatusAccepted {
		assert.Equal(t, false, true)
		return
	}

	createdUserID1, _ := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})

	user, _ := store.GetStore().GetUser(project.ID, createdUserID1)

	startTimestamp := U.UnixTimeBeforeDuration(time.Hour * 1)
	stepTimestamp := startTimestamp

	payload := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, "user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`, "s0", createdUserID1, stepTimestamp, "A", 1234)
	w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response := DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	timestamp := U.UnixTimeBeforeDuration(30 * 24 * time.Hour)
	eventName := U.RandomLowerAphaNumString(10)
	w = ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"event_name": "%s", "timestamp": %d, "event_properties": {"$referrer": "https://example.com/abc?ref=1", "$referrer_url": "https://example.com/abc", "$referrer_domain": "example.com", "$page_url": "https://example.com/xyz", "$page_raw_url": "https://example.com/xyz?utm_campaign=google", "$page_domain": "example.com", "$page_load_time": 100, "$page_spent_time": 120, "$qp_utm_campaign": "google", "$qp_utm_campaignid": "12345", "$qp_utm_ad": "ad_2021_1", "$qp_utm_ad_id": "9876543210", "$qp_utm_source": "google", "$qp_utm_medium": "email", "$qp_utm_keyword": "analytics", "$qp_utm_matchtype": "exact", "$qp_utm_content": "analytics", "$qp_utm_adgroup": "ad-xxx", "$qp_utm_adgroupid": "xyz123", "$qp_utm_creative": "creative-xxx", "$qp_gclid": "xxx123", "$qp_fbclid": "zzz123"}, "user_properties": {"$platform": "web", "$browser": "Mozilla", "$browser_version": "v0.1", "$browser_with_version": "Mozilla_v0.1", "$user_agent": "browser", "$os": "Linux", "$os_version": "v0.1", "$os_with_version": "Linux_v0.1", "$country": "india", "$region": "karnataka", "$city": "bengaluru", "$timezone": "Asia/Calcutta"}}`,
			eventName, timestamp)), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	w = ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"event_name": "%s", "timestamp": %d, "event_properties": {"$referrer": "https://example.com/abc?ref=1", "$referrer_url": "https://example.com/abc", "$referrer_domain": "example.com", "$page_url": "https://example.com/xyz", "$page_raw_url": "https://example.com/xyz?utm_campaign=google", "$page_domain": "example.com", "$page_load_time": 100, "$page_spent_time": 120, "$qp_utm_campaign": "google", "$qp_utm_campaignid": "12345", "$qp_utm_ad": "ad_2021_1", "$qp_utm_ad_id": "9876543210", "$qp_utm_source": "google", "$qp_utm_medium": "email", "$qp_utm_keyword": "analytics", "$qp_utm_matchtype": "exact", "$qp_utm_content": "analytics", "$qp_utm_adgroup": "ad-xxx", "$qp_utm_adgroupid": "xyz123", "$qp_utm_creative": "creative-xxx", "$qp_gclid": "xxx123", "$qp_fbclid": "zzz123"}, "user_properties": {"$platform": "web", "$browser": "Mozilla", "$browser_version": "v0.1", "$browser_with_version": "Mozilla_v0.1", "$user_agent": "browser", "$os": "Linux", "$os_version": "v0.1", "$os_with_version": "Linux_v0.1", "$country": "india", "$region": "karnataka", "$city": "bengaluru", "$timezone": "Asia/Calcutta"}}`,
			"123testing", timestamp)), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	contentGroupRequest := model.ContentGroup{}
	contentGroupRequest.ContentGroupName = "abc"
	contentGroupRequest.ContentGroupDescription = "description"
	value := model.ContentGroupValue{}
	value.LogicalOp = "OR"
	value.Operator = "contains"
	value.Value = "xyz"
	filters := make([]model.ContentGroupValue, 0)
	filters = append(filters, value)
	contentGroupValueArray := make([]model.ContentGroupRule, 0)
	contentGroupValue := model.ContentGroupRule{
		ContentGroupValue: "value",
		Rule:              filters,
	}
	contentGroupValueArray = append(contentGroupValueArray, contentGroupValue)
	contentGroupValueJson, _ := json.Marshal(contentGroupValueArray)
	contentGroupRequest.Rule = &postgres.Jsonb{contentGroupValueJson}
	w1 := sendCreateContentGroupRequest(a, contentGroupRequest, agent, project.ID)
	assert.Equal(t, http.StatusCreated, w1.Code)

	_, err := TaskSession.AddSession([]int64{project.ID}, timestamp-60, 0, 0, 0, 1, 1)
	assert.Nil(t, err)

	t.Run("Query with no groupby and no filter.", func(t *testing.T) {
		query := model.KPIQuery{
			Category:        "events",
			DisplayCategory: "page_views",
			PageUrl:         "s0",
			//Metrics:         []string{"page_views", "unique_users"},
			Metrics:          []string{"page_views"},
			Filters:          []model.KPIFilter{},
			From:             startTimestamp,
			To:               startTimestamp + 40,
			GroupByTimestamp: "date",
		}
		query1 := model.KPIQuery{}
		U.DeepCopy(&query, &query1)
		query1.GroupByTimestamp = ""

		kpiQueryGroup := model.KPIQueryGroup{
			Class:         "kpi",
			Queries:       []model.KPIQuery{query, query1},
			GlobalFilters: []model.KPIFilter{},
			GlobalGroupBy: []model.KPIGroupBy{},
		}

		result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
			C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, result[0].Headers, []string{"datetime", "page_views"})
		assert.Equal(t, len(result[0].Rows), 1)
		assert.Equal(t, result[1].Headers, []string{"page_views"})
		assert.Equal(t, len(result[1].Rows), 1)
		assert.Equal(t, result[0].Rows[0][1], float64(1))
	})

	t.Run("Query with multiple sub queries", func(t *testing.T) {

		query1 := model.KPIQuery{

			Category:        "events",
			DisplayCategory: "page_views",
			PageUrl:         "s0",
			Metrics:         []string{"page_views"},
			// Metrics:  []string{"page_views"},
			Filters:          nil,
			From:             startTimestamp,
			To:               startTimestamp + 40,
			Timezone:         "Asia/Kolkata",
			GroupByTimestamp: "date",
		}
		query2 := model.KPIQuery{}
		U.DeepCopy(&query1, &query2)
		query2.GroupByTimestamp = ""

		query3 := model.KPIQuery{

			Category:         "events",
			DisplayCategory:  "page_views",
			PageUrl:          "s0",
			Metrics:          []string{"unique_users"},
			Filters:          nil,
			From:             startTimestamp,
			To:               startTimestamp + 40,
			Timezone:         "Asia/Kolkata",
			GroupByTimestamp: "date",
		}
		query4 := model.KPIQuery{}
		U.DeepCopy(&query3, &query4)
		query4.GroupByTimestamp = ""

		kpiQueryGroup := model.KPIQueryGroup{
			Class:         "kpi",
			Queries:       []model.KPIQuery{query1, query2, query3, query4},
			GlobalFilters: []model.KPIFilter{},
			GlobalGroupBy: []model.KPIGroupBy{
				{
					ObjectType:       "s0",
					PropertyName:     "user_id",
					PropertyDataType: "categorical",
					GroupByType:      "",
					Granularity:      "",
					Entity:           "user",
				},
			},
		}

		result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(),
			kpiQueryGroup, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, result[0].Headers, []string{"datetime", "user_id", "page_views", "unique_users"})
		assert.Equal(t, len(result[0].Rows), 1)
		assert.Equal(t, result[0].Rows[0][1], "$none")
		assert.Equal(t, result[0].Rows[0][2], float64(1))
		assert.Equal(t, result[0].Rows[0][3], float64(1))
	})

	t.Run("Query with session", func(t *testing.T) {
		query1 := model.KPIQuery{
			Category:         "events",
			DisplayCategory:  "website_session",
			Metrics:          []string{"average_initial_page_load_time"},
			Filters:          nil,
			From:             timestamp - 60*2,
			To:               timestamp,
			GroupByTimestamp: "date",
		}

		query2 := model.KPIQuery{}
		U.DeepCopy(&query1, &query2)
		query2.GroupByTimestamp = ""

		kpiQueryGroup := model.KPIQueryGroup{
			Class:         "kpi",
			Queries:       []model.KPIQuery{query1, query2},
			GlobalFilters: []model.KPIFilter{},
			GlobalGroupBy: []model.KPIGroupBy{
				{
					ObjectType:       "s0",
					PropertyName:     "$user_id",
					PropertyDataType: "categorical",
					GroupByType:      "",
					Granularity:      "",
					Entity:           "user",
				},
			},
		}

		result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, result[0].Headers, []string{"datetime", "$user_id", "average_initial_page_load_time"})
		assert.Equal(t, len(result[0].Rows), 2)
		assert.Equal(t, result[0].Rows[0][2], float64(100))

		assert.Equal(t, result[1].Headers, []string{"$user_id", "average_initial_page_load_time"})
		assert.Equal(t, len(result[1].Rows), 2)
		assert.Equal(t, result[1].Rows[0][1], float64(100))
	})

	t.Run("Query for content group with session", func(t *testing.T) {

		query1 := model.KPIQuery{
			Category:        "events",
			DisplayCategory: "website_session",
			Metrics:         []string{"average_initial_page_load_time"},
			Filters:         nil,
			From:            timestamp - 60*2,
			To:              timestamp,
		}

		kpiQueryGroup := model.KPIQueryGroup{
			Class:         "kpi",
			Queries:       []model.KPIQuery{query1},
			GlobalFilters: []model.KPIFilter{},
			GlobalGroupBy: []model.KPIGroupBy{
				{
					ObjectType:       "s0",
					PropertyName:     "abc",
					PropertyDataType: "categorical",
					GroupByType:      "",
					Granularity:      "",
					Entity:           "event",
				},
			},
		}

		result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
			C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, statusCode)

		assert.Equal(t, result[0].Headers, []string{"abc", "average_initial_page_load_time"})
		assert.Equal(t, len(result[0].Rows), 1)
		assert.Equal(t, result[0].Rows[0][0], "value")
		assert.Equal(t, result[0].Rows[0][1], float64(100))
	})

	t.Run("Query with channel", func(t *testing.T) {

		query1 := model.KPIQuery{
			Category:         "channels",
			DisplayCategory:  "adwords_metrics",
			Metrics:          []string{"impressions"},
			Filters:          nil,
			From:             startTimestamp,
			To:               startTimestamp + 40,
			GroupByTimestamp: "date",
		}

		query2 := model.KPIQuery{}
		U.DeepCopy(&query1, &query2)
		query2.GroupByTimestamp = ""

		kpiQueryGroup := model.KPIQueryGroup{
			Class:         "kpi",
			Queries:       []model.KPIQuery{query1, query2},
			GlobalFilters: []model.KPIFilter{},
			GlobalGroupBy: []model.KPIGroupBy{},
		}

		result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
			C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, result[0].Headers, []string{"datetime", "adwords_metrics_impressions"})
		assert.Equal(t, len(result[0].Rows), 1)
	})

	t.Run("Query with channel and events at a time", func(t *testing.T) {

		query := model.KPIQuery{
			Category:        "events",
			DisplayCategory: "page_views",
			PageUrl:         "s0",
			//Metrics:         []string{"page_views", "unique_users"},
			Metrics:          []string{"page_views"},
			Filters:          []model.KPIFilter{},
			From:             startTimestamp,
			To:               startTimestamp + 2*86400,
			GroupByTimestamp: "date",
		}
		query1 := model.KPIQuery{
			Category:         "channels",
			DisplayCategory:  "adwords_metrics",
			Metrics:          []string{"impressions"},
			Filters:          nil,
			From:             startTimestamp,
			To:               startTimestamp + 2*86400,
			GroupByTimestamp: "date",
		}

		kpiQueryGroup := model.KPIQueryGroup{
			Class:         "kpi",
			Queries:       []model.KPIQuery{query, query1},
			GlobalFilters: []model.KPIFilter{},
			GlobalGroupBy: []model.KPIGroupBy{},
		}

		result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
			C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, result[0].Headers, []string{"datetime", "page_views", "adwords_metrics_impressions"})
		assert.Equal(t, len(result[0].Rows), 3)
	})

	t.Run("Query with channel and events with alias at a time", func(t *testing.T) {

		query := model.KPIQuery{
			Category:        "events",
			DisplayCategory: "page_views",
			PageUrl:         "s0",
			//Metrics:         []string{"page_views", "unique_users"},
			Metrics:          []string{"page_views"},
			Filters:          []model.KPIFilter{},
			From:             startTimestamp,
			To:               startTimestamp + 2*86400,
			AliasName:        "a1",
			GroupByTimestamp: "date",
		}
		query1 := model.KPIQuery{
			Category:         "channels",
			DisplayCategory:  "adwords_metrics",
			Metrics:          []string{"impressions"},
			Filters:          nil,
			From:             startTimestamp,
			To:               startTimestamp + 2*86400,
			AliasName:        "a2",
			GroupByTimestamp: "date",
		}

		kpiQueryGroup := model.KPIQueryGroup{
			Class:         "kpi",
			Queries:       []model.KPIQuery{query, query1},
			GlobalFilters: []model.KPIFilter{},
			GlobalGroupBy: []model.KPIGroupBy{},
		}

		result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
			C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, result[0].Headers, []string{"datetime", "a1", "adwords_metrics_a2"})
		assert.Equal(t, len(result[0].Rows), 3)
	})

	t.Run("Query for virtual Events", func(t *testing.T) {
		expr := "a.com/u1/u3/:prop1"
		name := "kpi_login"
		filterEventName, errCode := store.GetStore().CreateOrGetFilterEventName(&model.EventName{
			ProjectId:  project.ID,
			FilterExpr: expr,
			Name:       name,
		})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, filterEventName)
		assert.NotZero(t, filterEventName.ID)
		assert.Equal(t, name, filterEventName.Name)
		assert.Equal(t, expr, filterEventName.FilterExpr)
		assert.Equal(t, model.TYPE_FILTER_EVENT_NAME, filterEventName.Type)

		// Test filter_event_name hit with exact match.
		rEventName := "a.com/u1/u3/i1"
		// w := ServePostRequestWithHeaders(r, uri, []byte(fmt.Sprintf(`{"user_id": "%s", "event_name": "%s", "auto": true, "event_properties": {"mobile": "true", "$page_url": "%s"}, "user_properties": {"$os": "mac osx", "$os_version": "1_2_3"}}`,
		// user.ID, rEventName, rEventName), map[string]string{"Authorization": project.Token})
		w = ServePostRequestWithHeaders(r, uri, []byte(fmt.Sprintf(`{"user_id": "%s", "event_name": "%s", "auto": true, "timestamp": %d, "event_properties": {"mobile": "true", "$page_url": "%s"}, "user_properties": {"$os": "mac osx", "$os_version": "1_2_3"}}`,
			user.ID, rEventName, stepTimestamp, rEventName)), map[string]string{"Authorization": project.Token})
		assert.Equal(t, http.StatusOK, w.Code)
		responseMap := DecodeJSONResponseToMap(w.Body)
		assert.NotEmpty(t, responseMap)
		assert.Nil(t, responseMap["user_id"])
		rEvent, errCode := store.GetStore().GetEvent(project.ID, user.ID, responseMap["event_id"].(string))
		assert.Equal(t, http.StatusFound, errCode)
		assert.NotNil(t, rEvent)
		assert.Equal(t, filterEventName.ID, rEvent.EventNameId)
		var rEventProperties map[string]interface{}
		json.Unmarshal(rEvent.Properties.RawMessage, &rEventProperties)
		assert.NotNil(t, rEventProperties["prop1"])
		assert.Equal(t, "i1", rEventProperties["prop1"]) // Event property filled using expression.

		query := model.KPIQuery{
			Category:         "events",
			DisplayCategory:  "page_views",
			PageUrl:          "kpi_login",
			Metrics:          []string{"page_views"},
			Filters:          []model.KPIFilter{},
			From:             startTimestamp,
			To:               startTimestamp + 400,
			GroupByTimestamp: "date",
		}
		query1 := model.KPIQuery{}
		U.DeepCopy(&query, &query1)
		query1.GroupByTimestamp = ""

		kpiQueryGroup := model.KPIQueryGroup{
			Class:         "kpi",
			Queries:       []model.KPIQuery{query, query1},
			GlobalFilters: []model.KPIFilter{},
			GlobalGroupBy: []model.KPIGroupBy{},
		}

		result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
			C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, result[0].Headers, []string{"datetime", "page_views"})
		assert.Equal(t, len(result[0].Rows), 1)
		assert.Equal(t, result[1].Headers, []string{"page_views"})
		assert.Equal(t, len(result[1].Rows), 1)
		assert.Equal(t, result[0].Rows[0][1], float64(1))
	})
}

func TestKpiAnalyticsForProfile(t *testing.T) {
	a := gin.Default()
	H.InitAppRoutes(a)

	project, agent, _ := SetupProjectWithAgentDAO()
	rCustomerUserId := U.RandomLowerAphaNumString(15)
	properties1 := postgres.Jsonb{RawMessage: json.RawMessage([]byte(`{"country": "india", "age": 30, "$hubspot_amount": 300, "$hubspot_datefield1": 1640975325,  "paid": true}`))}
	properties2 := postgres.Jsonb{RawMessage: json.RawMessage([]byte(`{"country": "us", "age": 20, "$hubspot_amount": 200, "$hubspot_datefield1": 1640975525, "paid": true}`))}
	// properties2 := postgres.Jsonb{RawMessage: json.RawMessage([]byte(`{"country": "us", "age": 20, "$hubspot_amount": 300, "$hubspot_datefield1": 1640975425, "paid": true}`))}

	joinTime := int64(1640975425 - 100)

	createUserID1, newUserErrorCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: rCustomerUserId, Properties: properties1, JoinTimestamp: joinTime, Source: model.GetRequestSourcePointer(model.UserSourceHubspot)})
	assert.Equal(t, http.StatusCreated, newUserErrorCode)
	assert.NotEqual(t, "", createUserID1)

	nextUserJoinTime := joinTime + 100
	createUserID2, nextUserErrCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: properties2, JoinTimestamp: nextUserJoinTime, Source: model.GetRequestSourcePointer(model.UserSourceHubspot)})
	assert.Equal(t, http.StatusCreated, nextUserErrCode)
	assert.NotEqual(t, "", createUserID2)

	name1 := U.RandomString(8)
	description1 := U.RandomString(8)
	transformations1 := &postgres.Jsonb{json.RawMessage(`{"agFn": "sum", "agPr": "$hubspot_amount", "agPrTy": "categorical", "fil": [], "daFie": "$hubspot_datefield1"}`)}
	w := sendCreateCustomMetric(a, project.ID, agent, transformations1, name1, description1, "hubspot_contacts", 1)
	assert.Equal(t, http.StatusOK, w.Code)

	name2 := U.RandomString(8)
	description2 := U.RandomString(8)
	transformations2 := &postgres.Jsonb{json.RawMessage(`{"agFn": "sum", "agPr": "$hubspot_amount", "agPrTy": "categorical", "fil": [{"objTy": "", "prNa": "country", "prDaTy": "categorical", "en": "user", "co": "equals", "va": "india", "lOp": "AND"}], "daFie": "$hubspot_datefield1"}`)}
	w = sendCreateCustomMetric(a, project.ID, agent, transformations2, name2, description2, "hubspot_contacts", 1)
	assert.Equal(t, http.StatusOK, w.Code)

	name3 := U.RandomString(8)
	description3 := U.RandomString(8)
	transformations3 := &postgres.Jsonb{json.RawMessage(`{"agFn": "average", "agPr": "$hubspot_amount", "agPrTy": "categorical", "fil": [{"objTy": "", "prNa": "country", "prDaTy": "categorical", "en": "user", "co": "equals", "va": "india", "lOp": "AND"}], "daFie": "$hubspot_datefield1"}`)}
	w = sendCreateCustomMetric(a, project.ID, agent, transformations3, name3, description3, "hubspot_contacts", 1)
	assert.Equal(t, http.StatusOK, w.Code)

	t.Run("test hubspot contacts with no filters and no group by", func(t *testing.T) {
		query1 := model.KPIQuery{
			Category:         model.ProfileCategory,
			DisplayCategory:  "hubspot_contacts",
			PageUrl:          "",
			Metrics:          []string{name1},
			GroupBy:          []M.KPIGroupBy{},
			From:             1640975425 - 200,
			To:               1640975425 + 200,
			GroupByTimestamp: "date",
		}

		query2 := model.KPIQuery{}
		U.DeepCopy(&query1, &query2)
		query2.GroupByTimestamp = ""

		kpiQueryGroup := model.KPIQueryGroup{
			Class:         "kpi",
			Queries:       []model.KPIQuery{query1, query2},
			GlobalFilters: []model.KPIFilter{},
			GlobalGroupBy: []model.KPIGroupBy{},
		}
		result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
			C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, result[0].Headers, []string{"datetime", name1})
		assert.Equal(t, len(result[0].Rows), 1)
		assert.Equal(t, result[0].Rows[0][1], float64(500))

		assert.Equal(t, result[1].Headers, []string{name1})
		assert.Equal(t, len(result[1].Rows), 1)
		assert.Equal(t, result[1].Rows[0][0], float64(500))
	})

	t.Run("test hubspot contacts with filters only", func(t *testing.T) {
		query1 := model.KPIQuery{
			Category:         model.ProfileCategory,
			DisplayCategory:  "hubspot_contacts",
			PageUrl:          "",
			Metrics:          []string{name1},
			GroupBy:          []M.KPIGroupBy{},
			From:             1640975425 - 200,
			To:               1640975425 + 200,
			GroupByTimestamp: "date",
		}

		filter := model.KPIFilter{
			ObjectType:       "",
			PropertyName:     "country",
			PropertyDataType: "categorical",
			Entity:           "user",
			Condition:        "equals",
			Value:            "india",
			LogicalOp:        "AND",
		}

		query2 := model.KPIQuery{}
		U.DeepCopy(&query1, &query2)
		query2.GroupByTimestamp = ""

		kpiQueryGroup2 := model.KPIQueryGroup{
			Class:         "kpi",
			Queries:       []model.KPIQuery{query1, query2},
			GlobalFilters: []model.KPIFilter{filter},
			GlobalGroupBy: []model.KPIGroupBy{},
		}
		result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup2,
			C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, result[0].Headers, []string{"datetime", name1})
		assert.Equal(t, len(result[0].Rows), 1)
		assert.Equal(t, result[0].Rows[0][1], float64(300))

		assert.Equal(t, result[1].Headers, []string{name1})
		assert.Equal(t, len(result[1].Rows), 1)
		assert.Equal(t, result[1].Rows[0][0], float64(300))

	})

	t.Run("test hubspot contacts with filters only - timerange overshoot check with $none filters", func(t *testing.T) {
		query1 := model.KPIQuery{
			Category:         model.ProfileCategory,
			DisplayCategory:  "hubspot_contacts",
			PageUrl:          "",
			Metrics:          []string{name1},
			GroupBy:          []M.KPIGroupBy{},
			From:             1640975425 - 2,
			To:               1640975425 + 2,
			GroupByTimestamp: "date",
		}

		filter := model.KPIFilter{
			ObjectType:       "",
			PropertyName:     "country",
			PropertyDataType: "categorical",
			Entity:           "user",
			Condition:        "equals",
			Value:            "india",
			LogicalOp:        "AND",
		}
		filter1 := model.KPIFilter{
			ObjectType:       "",
			PropertyName:     "age",
			PropertyDataType: "categorical",
			Entity:           "user",
			Condition:        "equals",
			Value:            "20",
			LogicalOp:        "OR",
		}
		filter2 := model.KPIFilter{
			ObjectType:       "",
			PropertyName:     "country",
			PropertyDataType: "categorical",
			Entity:           "user",
			Condition:        "notEqual",
			Value:            "",
			LogicalOp:        "AND",
		}
		filter3 := model.KPIFilter{
			ObjectType:       "",
			PropertyName:     "country",
			PropertyDataType: "categorical",
			Entity:           "user",
			Condition:        "equals",
			Value:            "us",
			LogicalOp:        "AND",
		}

		query2 := model.KPIQuery{}
		U.DeepCopy(&query1, &query2)
		query2.GroupByTimestamp = ""

		kpiQueryGroup2 := model.KPIQueryGroup{
			Class:         "kpi",
			Queries:       []model.KPIQuery{query1, query2},
			GlobalFilters: []model.KPIFilter{filter, filter1, filter2, filter3},
			GlobalGroupBy: []model.KPIGroupBy{},
		}
		result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup2,
			C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, result[0].Headers, []string{"datetime", name1})
		assert.Equal(t, len(result[0].Rows), 0)
	})

	t.Run("test hubspot contacts with filters present in custom metric", func(t *testing.T) {
		// Query which supports simple function - Sum or count
		query1 := model.KPIQuery{
			Category:         model.ProfileCategory,
			DisplayCategory:  "hubspot_contacts",
			PageUrl:          "",
			Metrics:          []string{name2},
			GroupBy:          []M.KPIGroupBy{},
			From:             1640975425 - 200,
			To:               1640975425 + 200,
			GroupByTimestamp: "date",
		}

		query2 := model.KPIQuery{}
		U.DeepCopy(&query1, &query2)
		query2.GroupByTimestamp = ""

		kpiQueryGroup := model.KPIQueryGroup{
			Class:         "kpi",
			Queries:       []model.KPIQuery{query1, query2},
			GlobalFilters: []model.KPIFilter{},
			GlobalGroupBy: []model.KPIGroupBy{},
		}
		result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
			C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, result[0].Headers, []string{"datetime", name2})
		assert.Equal(t, len(result[0].Rows), 1)
		assert.Equal(t, result[0].Rows[0][1], float64(300))

		assert.Equal(t, result[1].Headers, []string{name2})
		assert.Equal(t, len(result[1].Rows), 1)
		assert.Equal(t, result[1].Rows[0][0], float64(300))

		// Query which supports complex function - Average
		query3 := model.KPIQuery{
			Category:         model.ProfileCategory,
			DisplayCategory:  "hubspot_contacts",
			PageUrl:          "",
			Metrics:          []string{name3},
			GroupBy:          []M.KPIGroupBy{},
			From:             1640975425 - 200,
			To:               1640975425 + 200,
			GroupByTimestamp: "date",
		}

		query4 := model.KPIQuery{}
		U.DeepCopy(&query3, &query4)
		query4.GroupByTimestamp = ""

		kpiQueryGroup2 := model.KPIQueryGroup{
			Class:         "kpi",
			Queries:       []model.KPIQuery{query3, query4},
			GlobalFilters: []model.KPIFilter{},
			GlobalGroupBy: []model.KPIGroupBy{},
		}
		result2, statusCode2 := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup2,
			C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, statusCode2)
		assert.Equal(t, result2[0].Headers, []string{"datetime", name3})
		assert.Equal(t, len(result2[0].Rows), 1)
		assert.Equal(t, result[0].Rows[0][1], float64(300))
	})

	t.Run("test hubspot contacts with filter and group by", func(t *testing.T) {
		// Query which supports simple function - Sum or count
		query1 := model.KPIQuery{
			Category:         model.ProfileCategory,
			DisplayCategory:  "hubspot_contacts",
			PageUrl:          "",
			Metrics:          []string{name2},
			GroupBy:          []M.KPIGroupBy{},
			From:             1640975425 - 200,
			To:               1640975425 + 200,
			GroupByTimestamp: "date",
		}

		query2 := model.KPIQuery{}
		U.DeepCopy(&query1, &query2)
		query2.GroupByTimestamp = ""

		groupBy := model.KPIGroupBy{
			ObjectType:       "",
			PropertyName:     "country",
			PropertyDataType: "categorical",
			Entity:           "user",
			GroupByType:      "",
			Granularity:      "",
		}

		kpiQueryGroup1 := model.KPIQueryGroup{
			Class:         "kpi",
			Queries:       []model.KPIQuery{query1, query2},
			GlobalFilters: []model.KPIFilter{},
			GlobalGroupBy: []model.KPIGroupBy{groupBy},
		}
		result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup1,
			C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, result[0].Headers, []string{"datetime", groupBy.PropertyName, name2})
		assert.Equal(t, len(result[0].Rows), 1)
		assert.Equal(t, result[0].Rows[0][1], "india")
		assert.Equal(t, result[0].Rows[0][2], float64(300))

		assert.Equal(t, result[1].Headers, []string{groupBy.PropertyName, name2})
		assert.Equal(t, len(result[1].Rows), 1)
		assert.Equal(t, result[1].Rows[0][0], "india")
		assert.Equal(t, result[1].Rows[0][1], float64(300))

		// Query which supports complex function - Average
		query3 := model.KPIQuery{
			Category:         model.ProfileCategory,
			DisplayCategory:  "hubspot_contacts",
			PageUrl:          "",
			Metrics:          []string{name3},
			GroupBy:          []M.KPIGroupBy{},
			From:             1640975425 - 200,
			To:               1640975425 + 200,
			GroupByTimestamp: "date",
		}

		query4 := model.KPIQuery{}
		U.DeepCopy(&query3, &query4)
		query4.GroupByTimestamp = ""

		kpiQueryGroup2 := model.KPIQueryGroup{
			Class:         "kpi",
			Queries:       []model.KPIQuery{query3, query4},
			GlobalFilters: []model.KPIFilter{},
			GlobalGroupBy: []model.KPIGroupBy{groupBy},
		}
		result2, statusCode2 := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup2,
			C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, statusCode2)
		assert.Equal(t, result2[0].Headers, []string{"datetime", groupBy.PropertyName, name3})
		assert.Equal(t, len(result2[0].Rows), 1)
		assert.Equal(t, result[0].Rows[0][1], "india")
		assert.Equal(t, result[0].Rows[0][2], float64(300))
	})

	t.Run("test alias hubspot contacts with filter and group by", func(t *testing.T) {
		// Query which supports simple function - Sum or count
		query1 := model.KPIQuery{
			Category:         model.ProfileCategory,
			DisplayCategory:  "hubspot_contacts",
			PageUrl:          "",
			Metrics:          []string{name2},
			GroupBy:          []M.KPIGroupBy{},
			From:             1640975425 - 200,
			To:               1640975425 + 200,
			AliasName:        "a1",
			GroupByTimestamp: "date",
		}

		query2 := model.KPIQuery{}
		U.DeepCopy(&query1, &query2)
		query2.GroupByTimestamp = ""

		groupBy := model.KPIGroupBy{
			ObjectType:       "",
			PropertyName:     "country",
			PropertyDataType: "categorical",
			Entity:           "user",
			GroupByType:      "",
			Granularity:      "",
		}

		kpiQueryGroup1 := model.KPIQueryGroup{
			Class:         "kpi",
			Queries:       []model.KPIQuery{query1, query2},
			GlobalFilters: []model.KPIFilter{},
			GlobalGroupBy: []model.KPIGroupBy{groupBy},
		}
		result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup1,
			C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, result[0].Headers, []string{"datetime", groupBy.PropertyName, "a1"})
		assert.Equal(t, len(result[0].Rows), 1)
		assert.Equal(t, result[0].Rows[0][1], "india")
		assert.Equal(t, result[0].Rows[0][2], float64(300))

		assert.Equal(t, result[1].Headers, []string{groupBy.PropertyName, "a1"})
		assert.Equal(t, len(result[1].Rows), 1)
		assert.Equal(t, result[1].Rows[0][0], "india")
		assert.Equal(t, result[1].Rows[0][1], float64(300))

		log.WithField("result", result).Warn("kark1")

	})
}

func TestKPIProfilesForGroups(t *testing.T) {
	a := gin.Default()
	H.InitAppRoutes(a)
	project, _, _, err := SetupProjectUserEventNameReturnDAO()

	agent, _ := SetupAgentReturnDAO(getRandomEmail(), "+1343545")

	_, _ = store.GetStore().CreateProjectAgentMappingWithDependencies(&model.ProjectAgentMapping{
		ProjectID: project.ID, AgentUUID: agent.UUID})

	assert.Nil(t, err)

	t.Run("TestKPIProfilesForGroups", func(t *testing.T) {
		initialTimestamp := time.Now().AddDate(0, 0, -10).Unix()
		var finalTimestamp int64
		var sourceHubspotUsers1 []model.User

		group, status := store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_HUBSPOT_COMPANY, model.AllowedGroupNames)
		assert.Equal(t, http.StatusCreated, status)
		assert.NotNil(t, group)

		properties := postgres.Jsonb{RawMessage: json.RawMessage([]byte(fmt.Sprintf(
			`{"country": "us", "age": 20, "$hubspot_amount": 200, "$hubspot_datefield1": %d, "paid": true}`, initialTimestamp+10)))}
		// create 10 group users, source = hubspot and group_name = $hubspot_company
		for i := 0; i < 10; i++ {
			createdUserID, errCode := store.GetStore().CreateGroupUser(&model.User{ProjectId: project.ID,
				Source: model.GetRequestSourcePointer(model.UserSourceHubspot), Properties: properties}, group.Name, fmt.Sprintf("%d", group.ID))
			assert.Equal(t, http.StatusCreated, errCode)
			user, errCode := store.GetStore().GetUser(project.ID, createdUserID)
			assert.Equal(t, http.StatusFound, errCode)
			assert.True(t, len(user.ID) > 30)
			sourceHubspotUsers1 = append(sourceHubspotUsers1, *user)
		}
		finalTimestamp = time.Now().Unix()

		// update user properties to add $group_id property = group.ID of created user
		for i := 0; i < len(sourceHubspotUsers1); i++ {
			newProperties := &postgres.Jsonb{RawMessage: json.RawMessage([]byte(fmt.Sprintf(
				`{"$group_id": "%d"}`, group.ID)))}
			_, status := store.GetStore().UpdateUserPropertiesV2(project.ID, sourceHubspotUsers1[i].ID, newProperties, time.Now().Unix(), "", "")
			assert.Equal(t, http.StatusAccepted, status)
		}

		name2 := U.RandomString(8)
		description2 := U.RandomString(8)
		transformations2 := &postgres.Jsonb{json.RawMessage(`{"agFn": "unique", "agPr": "", "agPrTy": "categorical", "fil": [], "daFie": "$hubspot_datefield1"}`)}
		w := sendCreateCustomMetric(a, project.ID, agent, transformations2, name2, description2, "hubspot_companies", 1)
		assert.Equal(t, http.StatusOK, w.Code)

		query1 := model.KPIQuery{
			Category:         model.ProfileCategory,
			DisplayCategory:  "hubspot_companies",
			PageUrl:          "",
			Metrics:          []string{name2},
			GroupBy:          []M.KPIGroupBy{},
			From:             initialTimestamp,
			To:               finalTimestamp,
			GroupByTimestamp: "date",
		}

		query2 := model.KPIQuery{}
		U.DeepCopy(&query1, &query2)
		query2.GroupByTimestamp = ""

		kpiQueryGroup1 := model.KPIQueryGroup{
			Class:         "kpi",
			Queries:       []model.KPIQuery{query1, query2},
			GlobalFilters: []model.KPIFilter{},
			GlobalGroupBy: []model.KPIGroupBy{},
		}
		result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup1,
			C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, statusCode)

		assert.Equal(t, float64(10), result[0].Rows[0][1])
		assert.Equal(t, float64(10), result[1].Rows[0][0])
	})
}

func TestKpiAnalyticsForCustomEvents(t *testing.T) {
	a := gin.Default()
	H.InitAppRoutes(a)

	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"

	r2 := gin.Default()
	H.InitDataServiceRoutes(r2)
	model.SetSmartPropertiesReservedNames()

	project, customerAccountID, agent, statusCode := createProjectAndAddAdwordsDocument(t, r2)
	log.Warn(customerAccountID)
	if statusCode != http.StatusAccepted {
		assert.Equal(t, false, true)
		return
	}

	createdUserID1, _ := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})

	startTimestamp := U.UnixTimeBeforeDuration(time.Hour * 1)
	stepTimestamp := startTimestamp

	payload := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, "user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`, "s0", createdUserID1, stepTimestamp, "A", 1234)
	w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response := DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	timestamp := U.UnixTimeBeforeDuration(30 * 24 * time.Hour)
	eventName := U.RandomLowerAphaNumString(10)
	w = ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"event_name": "%s", "timestamp": %d, "event_properties": {"$referrer": "https://example.com/abc?ref=1", "$referrer_url": "https://example.com/abc", "$referrer_domain": "example.com", "$page_url": "https://example.com/xyz", "$page_raw_url": "https://example.com/xyz?utm_campaign=google", "$page_domain": "example.com", "$page_load_time": 100, "$page_spent_time": 120, "$qp_utm_campaign": "google", "$qp_utm_campaignid": "12345", "$qp_utm_ad": "ad_2021_1", "$qp_utm_ad_id": "9876543210", "$qp_utm_source": "google", "$qp_utm_medium": "email", "$qp_utm_keyword": "analytics", "$qp_utm_matchtype": "exact", "$qp_utm_content": "analytics", "$qp_utm_adgroup": "ad-xxx", "$qp_utm_adgroupid": "xyz123", "$qp_utm_creative": "creative-xxx", "$qp_gclid": "xxx123", "$qp_fbclid": "zzz123"}, "user_properties": {"$platform": "web", "$browser": "Mozilla", "$browser_version": "v0.1", "$browser_with_version": "Mozilla_v0.1", "$user_agent": "browser", "$os": "Linux", "$os_version": "v0.1", "$os_with_version": "Linux_v0.1", "$country": "india", "$region": "karnataka", "$city": "bengaluru", "$timezone": "Asia/Calcutta"}}`,
			eventName, timestamp)), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	w = ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"event_name": "%s", "timestamp": %d, "event_properties": {"$referrer": "https://example.com/abc?ref=2", "$referrer_url": "https://example.com/abc", "$referrer_domain": "example.com", "$page_url": "https://example.com/xyz", "$page_raw_url": "https://example.com/xyz?utm_campaign=google", "$page_domain": "example.com", "$page_load_time": 100, "$page_spent_time": 120, "$qp_utm_campaign": "google", "$qp_utm_campaignid": "12345", "$qp_utm_ad": "ad_2021_1", "$qp_utm_ad_id": "9876543210", "$qp_utm_source": "google", "$qp_utm_medium": "email", "$qp_utm_keyword": "analytics", "$qp_utm_matchtype": "exact", "$qp_utm_content": "analytics", "$qp_utm_adgroup": "ad-xxx", "$qp_utm_adgroupid": "xyz123", "$qp_utm_creative": "creative-xxx", "$qp_gclid": "xxx123", "$qp_fbclid": "zzz123"}, "user_properties": {"$platform": "web", "$browser": "Mozilla", "$browser_version": "v0.1", "$browser_with_version": "Mozilla_v0.1", "$user_agent": "browser", "$os": "Linux", "$os_version": "v0.1", "$os_with_version": "Linux_v0.1", "$country": "india", "$region": "karnataka", "$city": "bengaluru", "$timezone": "Asia/Calcutta"}}`,
			eventName, timestamp)), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	contentGroupRequest := model.ContentGroup{}
	contentGroupRequest.ContentGroupName = "abc"
	contentGroupRequest.ContentGroupDescription = "description"
	value := model.ContentGroupValue{}
	value.LogicalOp = "OR"
	value.Operator = "contains"
	value.Value = "xyz"
	filters := make([]model.ContentGroupValue, 0)
	filters = append(filters, value)
	contentGroupValueArray := make([]model.ContentGroupRule, 0)
	contentGroupValue := model.ContentGroupRule{
		ContentGroupValue: "value",
		Rule:              filters,
	}
	contentGroupValueArray = append(contentGroupValueArray, contentGroupValue)
	contentGroupValueJson, _ := json.Marshal(contentGroupValueArray)
	contentGroupRequest.Rule = &postgres.Jsonb{contentGroupValueJson}
	w1 := sendCreateContentGroupRequest(a, contentGroupRequest, agent, project.ID)
	assert.Equal(t, http.StatusCreated, w1.Code)

	_, err := TaskSession.AddSession([]int64{project.ID}, timestamp-60, 0, 0, 0, 1, 1)
	assert.Nil(t, err)

	t.Run("Query with no groupby and no filter.", func(t *testing.T) {
		// Custom Metric Create with name name
		name := U.RandomLowerAphaNumString(10)
		description := U.RandomString(8)
		transformationRaw := fmt.Sprintf(`{"agFn": "count", "agPr": "1", "agPrTy": "categorical", "fil": [], "daFie": "%d", "evNm": "%s", "en": "%s"}`, timestamp, "s0", model.QueryTypeEventsOccurrence)
		transformations := &postgres.Jsonb{RawMessage: json.RawMessage(transformationRaw)}
		w2 := sendCreateCustomMetric(a, project.ID, agent, transformations, name, description, model.EventsBasedDisplayCategory, 3)
		assert.Equal(t, http.StatusOK, w2.Code)

		query := model.KPIQuery{
			QueryType:        model.KpiCustomQueryType,
			Category:         "events",
			DisplayCategory:  model.EventsBasedDisplayCategory,
			PageUrl:          "s0",
			Metrics:          []string{name},
			Filters:          []model.KPIFilter{},
			From:             startTimestamp,
			To:               startTimestamp + 40,
			GroupByTimestamp: "date",
		}
		query1 := model.KPIQuery{}
		U.DeepCopy(&query, &query1)
		query1.GroupByTimestamp = ""

		kpiQueryGroup := model.KPIQueryGroup{
			Class:         "kpi",
			Queries:       []model.KPIQuery{query, query1},
			GlobalFilters: []model.KPIFilter{},
			GlobalGroupBy: []model.KPIGroupBy{},
		}

		result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
			C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, result[0].Headers, []string{"datetime", name})
		assert.Equal(t, len(result[0].Rows), 1)
		assert.Equal(t, result[1].Headers, []string{name})
		assert.Equal(t, len(result[1].Rows), 1)
		assert.Equal(t, result[0].Rows[0][1], float64(1))
	})

	t.Run("Query with multiple sub queries", func(t *testing.T) {

		name1 := U.RandomLowerAphaNumString(10)
		description := U.RandomString(8)
		transformationRaw := fmt.Sprintf(`{"agFn": "count", "agPr": "1", "agPrTy": "categorical", "fil": [], "daFie": "%d", "evNm": "%s", "en": "%s"}`, timestamp, "s0", model.QueryTypeEventsOccurrence)
		transformations := &postgres.Jsonb{RawMessage: json.RawMessage(transformationRaw)}
		w2 := sendCreateCustomMetric(a, project.ID, agent, transformations, name1, description, model.EventsBasedDisplayCategory, 3)
		assert.Equal(t, http.StatusOK, w2.Code)

		query1 := model.KPIQuery{
			QueryType:        model.KpiCustomQueryType,
			Category:         "events",
			DisplayCategory:  model.EventsBasedDisplayCategory,
			PageUrl:          "s0",
			Metrics:          []string{name1},
			Filters:          nil,
			From:             startTimestamp,
			To:               startTimestamp + 40,
			Timezone:         "Asia/Kolkata",
			GroupByTimestamp: "date",
		}
		query2 := model.KPIQuery{}
		U.DeepCopy(&query1, &query2)
		query2.GroupByTimestamp = ""

		name2 := U.RandomLowerAphaNumString(10)
		description = U.RandomString(8)
		transformationRaw = fmt.Sprintf(`{"agFn": "count", "agPr": "1", "agPrTy": "categorical", "fil": [], "daFie": "%d", "evNm": "%s", "en": "%s"}`, timestamp, "s0", model.QueryTypeEventsOccurrence)
		transformations = &postgres.Jsonb{RawMessage: json.RawMessage(transformationRaw)}
		w2 = sendCreateCustomMetric(a, project.ID, agent, transformations, name2, description, model.EventsBasedDisplayCategory, 3)
		assert.Equal(t, http.StatusOK, w2.Code)

		query3 := model.KPIQuery{
			QueryType:        model.KpiCustomQueryType,
			Category:         "events",
			DisplayCategory:  model.EventsBasedDisplayCategory,
			PageUrl:          "s0",
			Metrics:          []string{name2},
			Filters:          nil,
			From:             startTimestamp,
			To:               startTimestamp + 40,
			Timezone:         "Asia/Kolkata",
			GroupByTimestamp: "date",
		}
		query4 := model.KPIQuery{}
		U.DeepCopy(&query3, &query4)
		query4.GroupByTimestamp = ""

		kpiQueryGroup := model.KPIQueryGroup{
			Class:         "kpi",
			Queries:       []model.KPIQuery{query1, query2, query3, query4},
			GlobalFilters: []model.KPIFilter{},
			GlobalGroupBy: []model.KPIGroupBy{
				{
					ObjectType:       "s0",
					PropertyName:     "user_id",
					PropertyDataType: "categorical",
					GroupByType:      "",
					Granularity:      "",
					Entity:           "user",
				},
			},
		}

		result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(),
			kpiQueryGroup, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, result[0].Headers, []string{"datetime", "user_id", name1, name2})
		assert.Equal(t, len(result[0].Rows), 1)
		assert.Equal(t, result[0].Rows[0][1], "$none")
		assert.Equal(t, result[0].Rows[0][2], float64(1))
		assert.Equal(t, result[0].Rows[0][3], float64(1))
	})

	t.Run("Query with session", func(t *testing.T) {

		name := U.RandomLowerAphaNumString(10)
		description := U.RandomString(8)
		transformationRaw := fmt.Sprintf(`{"agFn": "average", "agPr": "$initial_page_load_time", "agPrTy": "numerical", "fil": [], "daFie": "%d", "evNm": "%s", "en": "%s"}`, timestamp, "$session", model.QueryTypeEventsOccurrence)
		transformations := &postgres.Jsonb{RawMessage: json.RawMessage(transformationRaw)}
		w3 := sendCreateCustomMetric(a, project.ID, agent, transformations, name, description, model.EventsBasedDisplayCategory, 3)
		assert.Equal(t, http.StatusOK, w3.Code)

		query1 := model.KPIQuery{
			QueryType:        model.KpiCustomQueryType,
			Category:         "events",
			DisplayCategory:  model.EventsBasedDisplayCategory,
			Metrics:          []string{name},
			Filters:          nil,
			From:             timestamp - 60*2,
			To:               timestamp,
			GroupByTimestamp: "date",
		}

		query2 := model.KPIQuery{}
		U.DeepCopy(&query1, &query2)
		query2.GroupByTimestamp = ""

		kpiQueryGroup := model.KPIQueryGroup{
			Class:         "kpi",
			Queries:       []model.KPIQuery{query1, query2},
			GlobalFilters: []model.KPIFilter{},
			GlobalGroupBy: []model.KPIGroupBy{
				{
					ObjectType:       "s0",
					PropertyName:     "user_id",
					PropertyDataType: "categorical",
					GroupByType:      "",
					Granularity:      "",
					Entity:           "user",
				},
			},
		}

		result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		rw, _ := json.Marshal(result)
		fmt.Println("result", string(rw))
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, result[0].Headers, []string{"datetime", "user_id", name})
		assert.Equal(t, len(result[0].Rows), 1)
		assert.Equal(t, result[0].Rows[0][1], "$none")
		assert.Equal(t, result[0].Rows[0][2], float64(100))

		assert.Equal(t, result[1].Headers, []string{"user_id", name})
		assert.Equal(t, len(result[1].Rows), 1)
		assert.Equal(t, result[1].Rows[0][0], "$none")
		assert.Equal(t, result[1].Rows[0][1], float64(100))
	})

	t.Run("Query for content group with session", func(t *testing.T) {

		name := U.RandomLowerAphaNumString(10)
		description := U.RandomString(8)
		transformationRaw := fmt.Sprintf(`{"agFn": "average", "agPr": "$initial_page_load_time", "agPrTy": "numerical", "fil": [], "daFie": "%d", "evNm": "%s", "en": "%s"}`, timestamp, "$session", model.QueryTypeEventsOccurrence)
		transformations := &postgres.Jsonb{RawMessage: json.RawMessage(transformationRaw)}
		w3 := sendCreateCustomMetric(a, project.ID, agent, transformations, name, description, model.EventsBasedDisplayCategory, 3)
		assert.Equal(t, http.StatusOK, w3.Code)

		query1 := model.KPIQuery{
			QueryType:       model.KpiCustomQueryType,
			Category:        "events",
			DisplayCategory: model.EventsBasedDisplayCategory,
			Metrics:         []string{name},
			Filters:         nil,
			From:            timestamp - 60*2,
			To:              timestamp,
		}

		kpiQueryGroup := model.KPIQueryGroup{
			Class:         "kpi",
			Queries:       []model.KPIQuery{query1},
			GlobalFilters: []model.KPIFilter{},
			GlobalGroupBy: []model.KPIGroupBy{
				{
					ObjectType:       "s0",
					PropertyName:     "abc",
					PropertyDataType: "categorical",
					GroupByType:      "",
					Granularity:      "",
					Entity:           "event",
				},
			},
		}

		result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
			C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, statusCode)

		assert.Equal(t, result[0].Headers, []string{"abc", name})
		assert.Equal(t, len(result[0].Rows), 1)
		assert.Equal(t, result[0].Rows[0][0], "value")
		assert.Equal(t, result[0].Rows[0][1], float64(100))
	})

	t.Run("Query with no groupby and no filter for Unique.", func(t *testing.T) {
		// Custom Metric Create with name name
		name := U.RandomLowerAphaNumString(10)
		description := U.RandomString(8)
		transformationRaw := fmt.Sprintf(`{"agFn": "unique", "agPr": "1", "agPrTy": "categorical", "fil": [], "daFie": "%d", "evNm": "%s", "en": "%s"}`, timestamp, "s0", "")
		transformations := &postgres.Jsonb{RawMessage: json.RawMessage(transformationRaw)}
		w2 := sendCreateCustomMetric(a, project.ID, agent, transformations, name, description, model.EventsBasedDisplayCategory, 3)
		assert.Equal(t, http.StatusOK, w2.Code)

		query := model.KPIQuery{
			QueryType:        model.KpiCustomQueryType,
			Category:         "events",
			DisplayCategory:  model.EventsBasedDisplayCategory,
			PageUrl:          "s0",
			Metrics:          []string{name},
			Filters:          []model.KPIFilter{},
			From:             startTimestamp,
			To:               startTimestamp + 40,
			GroupByTimestamp: "date",
		}
		query1 := model.KPIQuery{}
		U.DeepCopy(&query, &query1)
		query1.GroupByTimestamp = ""

		kpiQueryGroup := model.KPIQueryGroup{
			Class:         "kpi",
			Queries:       []model.KPIQuery{query, query1},
			GlobalFilters: []model.KPIFilter{},
			GlobalGroupBy: []model.KPIGroupBy{},
		}

		result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
			C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, result[0].Headers, []string{"datetime", name})
		assert.Equal(t, len(result[0].Rows), 1)
		assert.Equal(t, result[1].Headers, []string{name})
		assert.Equal(t, len(result[1].Rows), 1)
		assert.Equal(t, result[0].Rows[0][1], float64(1))

		// Add another s0 event to the same user, so count of s0 events is 2 and unique user count is 1
		stepTimestamp += 10
		payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, "user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`, "s0", createdUserID1, stepTimestamp, "A", 1234)
		w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
		assert.Equal(t, http.StatusOK, w.Code)
		response = DecodeJSONResponseToMap(w.Body)
		assert.NotNil(t, response["event_id"])

		result, statusCode = store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
			C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, result[0].Headers, []string{"datetime", name})
		assert.Equal(t, len(result[0].Rows), 1)
		assert.Equal(t, result[1].Headers, []string{name})
		assert.Equal(t, len(result[1].Rows), 1)
		assert.Equal(t, result[0].Rows[0][1], float64(1))

		// Add another user and check if the count of unique users is 2 now and count of s0 events is 3
		createdUserID2, _ := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})

		payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, "user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`, "s0", createdUserID2, stepTimestamp, "A", 1234)
		w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
		assert.Equal(t, http.StatusOK, w.Code)
		response = DecodeJSONResponseToMap(w.Body)
		assert.NotNil(t, response["event_id"])

		result, statusCode = store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
			C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, result[0].Headers, []string{"datetime", name})
		assert.Equal(t, len(result[0].Rows), 1)
		assert.Equal(t, result[1].Headers, []string{name})
		assert.Equal(t, len(result[1].Rows), 1)
		assert.Equal(t, result[0].Rows[0][1], float64(2))
	})

	t.Run("Query with filter in custom metric transformation.", func(t *testing.T) {
		name := U.RandomLowerAphaNumString(10)
		description := U.RandomString(8)
		transformationRaw := fmt.Sprintf(`{"agFn": "count", "agPr": "1", "agPrTy": "categorical", "fil": [{"objTy": "", "prNa": "$referrer", "prDaTy": "categorical", "en": "event", "co": "equals", "va": "https://example.com/abc?ref=2", "lOp": "AND"}], "daFie": "%d", "evNm": "%s", "en": "%s"}`, timestamp, eventName, model.QueryTypeEventsOccurrence)
		transformations := &postgres.Jsonb{RawMessage: json.RawMessage(transformationRaw)}
		w2 := sendCreateCustomMetric(a, project.ID, agent, transformations, name, description, model.EventsBasedDisplayCategory, 3)
		assert.Equal(t, http.StatusOK, w2.Code)

		query := model.KPIQuery{
			QueryType:        model.KpiCustomQueryType,
			Category:         "events",
			DisplayCategory:  model.EventsBasedDisplayCategory,
			Metrics:          []string{name},
			Filters:          []model.KPIFilter{},
			From:             timestamp,
			To:               timestamp + 40,
			GroupByTimestamp: "date",
		}
		query1 := model.KPIQuery{}
		U.DeepCopy(&query, &query1)
		query1.GroupByTimestamp = ""

		kpiQueryGroup := model.KPIQueryGroup{
			Class:         "kpi",
			Queries:       []model.KPIQuery{query, query1},
			GlobalFilters: []model.KPIFilter{},
			GlobalGroupBy: []model.KPIGroupBy{},
		}

		result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
			C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, result[0].Headers, []string{"datetime", name})
		assert.Equal(t, len(result[0].Rows), 1)
		assert.Equal(t, result[1].Headers, []string{name})
		assert.Equal(t, len(result[1].Rows), 1)
		assert.Equal(t, result[0].Rows[0][1], float64(1))
	})

}

func TestDerivedKPIChannels(t *testing.T) {
	r := gin.Default()
	H.InitDataServiceRoutes(r)
	model.SetSmartPropertiesReservedNames()

	a := gin.Default()
	H.InitAppRoutes(a)

	project, customerAccountID, agent, statusCode := createProjectAndAddAdwordsDocument(t, r)
	if statusCode != http.StatusAccepted {
		assert.Equal(t, false, true)
		return
	}

	adwordsDocuments := []M.AdwordsDocument{
		{ID: "1", Timestamp: 20220802, ProjectID: project.ID, CustomerAccountID: customerAccountID, TypeAlias: "campaign_performance_report",
			Value: &postgres.Jsonb{json.RawMessage(`{"cost": "11","clicks": "100","campaign_id":"1","impressions": "1000", "campaign_name": "test1"}`)}},
		{ID: "2", Timestamp: 20220802, ProjectID: project.ID, CustomerAccountID: customerAccountID, TypeAlias: "campaign_performance_report",
			Value: &postgres.Jsonb{json.RawMessage(`{"cost": "12","clicks": "200","campaign_id":"2","impressions": "500", "campaign_name": "test2"}`)}},
	}

	for _, adwordsDocument := range adwordsDocuments {
		status := store.GetStore().CreateAdwordsDocument(&adwordsDocument)
		assert.Equal(t, http.StatusCreated, status)
	}
	query := model.KPIQuery{}
	kpiQueryGroup := model.KPIQueryGroup{}

	name1 := U.RandomString(8)
	name2 := U.RandomString(8)
	description1 := U.RandomString(8)
	transformations1 := &postgres.Jsonb{json.RawMessage(`{"cl":"kpi","for":"a/b","qG":[{"ca":"channels","dc":"google_ads_metrics","fil":[],"gBy":[],"me":["impressions"],"na":"a","pgUrl":"","tz":"Australia/Sydney"},{"ca":"channels","dc":"google_ads_metrics","fil":[],"gBy":[],"me":["clicks"],"na":"b","pgUrl":"","tz":"Australia/Sydney"}]}`)}
	w := sendCreateCustomMetric(a, project.ID, agent, transformations1, name1, description1, "google_ads_metrics", 2)
	assert.Equal(t, http.StatusOK, w.Code)
	query = model.KPIQuery{
		Category:         "channels",
		DisplayCategory:  "google_ads_metrics",
		PageUrl:          "",
		Metrics:          []string{name1},
		GroupBy:          []M.KPIGroupBy{},
		From:             1659312000,
		To:               1659657600,
		QueryType:        "derived",
		GroupByTimestamp: "date",
	}
	query1 := model.KPIQuery{}
	U.DeepCopy(&query, &query1)
	query1.GroupByTimestamp = ""

	kpiQueryGroup = model.KPIQueryGroup{
		Class:         "kpi",
		Queries:       []model.KPIQuery{query, query1},
		GlobalFilters: []model.KPIFilter{},
		GlobalGroupBy: []model.KPIGroupBy{},
	}

	result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
		C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	log.WithField("result", result).Warn("kark")
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, result[0].Headers, []string{"datetime", "google_ads_metrics_" + name1})
	assert.Equal(t, result[1].Headers, []string{"google_ads_metrics_" + name1})
	assert.Equal(t, len(result[1].Rows), 1)
	assert.Equal(t, result[1].Rows[0][0], float64(5))

	t.Run("Derived kpi query with gbt, no group bys", func(t *testing.T) {
		query = model.KPIQuery{
			Category:         "channels",
			DisplayCategory:  "google_ads_metrics",
			PageUrl:          "",
			Metrics:          []string{name1},
			GroupBy:          []M.KPIGroupBy{},
			From:             1659312000,
			To:               1659657600,
			GroupByTimestamp: "date",
			QueryType:        "derived",
		}
		kpiQueryGroup = model.KPIQueryGroup{
			Class:         "kpi",
			Queries:       []model.KPIQuery{query},
			GlobalFilters: []model.KPIFilter{},
			GlobalGroupBy: []model.KPIGroupBy{},
		}

		result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
			C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, result[0].Headers, []string{"datetime", "google_ads_metrics_" + name1})
		assert.Equal(t, len(result[0].Rows), 5)
		assert.Equal(t, result[0].Rows[1][1], float64(5))
	})

	t.Run("Derived kpi query without gbt, with group by", func(t *testing.T) {
		query = model.KPIQuery{
			Category:         "channels",
			DisplayCategory:  "google_ads_metrics",
			PageUrl:          "",
			Metrics:          []string{name1},
			GroupBy:          []M.KPIGroupBy{},
			From:             1659312000,
			To:               1659657600,
			GroupByTimestamp: "",
			QueryType:        "derived",
		}
		kpiQueryGroup = model.KPIQueryGroup{
			Class:         "kpi",
			Queries:       []model.KPIQuery{query},
			GlobalFilters: []model.KPIFilter{},
			GlobalGroupBy: []model.KPIGroupBy{
				{
					ObjectType:       "campaign",
					PropertyName:     "campaign_name",
					PropertyDataType: "categorical",
					Entity:           "",
				},
			},
		}

		result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
			C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, result[0].Headers, []string{"campaign_name", "google_ads_metrics_" + name1})
		assert.Equal(t, len(result[0].Rows), 2)
	})

	t.Run("Derived kpi failure case with wrong metric name", func(t *testing.T) {
		query = model.KPIQuery{
			Category:         "channels",
			DisplayCategory:  "google_ads_metrics",
			PageUrl:          "",
			Metrics:          []string{name2},
			GroupBy:          []M.KPIGroupBy{},
			From:             1659312000,
			To:               1659657600,
			GroupByTimestamp: "date",
			QueryType:        "derived",
		}
		kpiQueryGroup = model.KPIQueryGroup{
			Class:         "kpi",
			Queries:       []model.KPIQuery{query},
			GlobalFilters: []model.KPIFilter{},
			GlobalGroupBy: []model.KPIGroupBy{},
		}

		_, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
			C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusInternalServerError, statusCode)
	})

	// tests for derived kpi with numeric value
	name3 := U.RandomString(8)
	description2 := U.RandomString(8)
	transformations2 := &postgres.Jsonb{json.RawMessage(`{"cl":"kpi","for":"(a*5)/(b*2.5)*1","qG":[{"ca":"channels","dc":"google_ads_metrics","fil":[],"gBy":[],"me":["impressions"],"na":"a","pgUrl":"","tz":"Australia/Sydney"},{"ca":"channels","dc":"google_ads_metrics","fil":[],"gBy":[],"me":["clicks"],"na":"b","pgUrl":"","tz":"Australia/Sydney"}]}`)}
	w = sendCreateCustomMetric(a, project.ID, agent, transformations2, name3, description2, "google_ads_metrics", 2)
	assert.Equal(t, http.StatusOK, w.Code)

	t.Run("Numeric Derived kpi query without gbt, no group bys", func(t *testing.T) {
		query = model.KPIQuery{
			Category:        "channels",
			DisplayCategory: "google_ads_metrics",
			PageUrl:         "",
			Metrics:         []string{name3},
			GroupBy:         []M.KPIGroupBy{},
			From:            1659312000,
			To:              1659657600,
			QueryType:       "derived",
		}
		kpiQueryGroup = model.KPIQueryGroup{
			Class:         "kpi",
			Queries:       []model.KPIQuery{query},
			GlobalFilters: []model.KPIFilter{},
			GlobalGroupBy: []model.KPIGroupBy{},
		}

		result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
			C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, result[0].Headers, []string{"google_ads_metrics_" + name3})
		assert.Equal(t, len(result[0].Rows), 1)
		assert.Equal(t, result[0].Rows[0][0], float64(10))
	})

	t.Run("Numeric Derived kpi query with gbt, no group bys", func(t *testing.T) {
		query = model.KPIQuery{
			Category:         "channels",
			DisplayCategory:  "google_ads_metrics",
			PageUrl:          "",
			Metrics:          []string{name3},
			GroupBy:          []M.KPIGroupBy{},
			From:             1659312000,
			To:               1659657600,
			GroupByTimestamp: "date",
			QueryType:        "derived",
		}
		kpiQueryGroup = model.KPIQueryGroup{
			Class:         "kpi",
			Queries:       []model.KPIQuery{query},
			GlobalFilters: []model.KPIFilter{},
			GlobalGroupBy: []model.KPIGroupBy{},
		}

		result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
			C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, result[0].Headers, []string{"datetime", "google_ads_metrics_" + name3})
		assert.Equal(t, len(result[0].Rows), 5)
		assert.Equal(t, result[0].Rows[1][1], float64(10))
	})

	name1 = U.RandomString(8)
	name2 = U.RandomString(8)
	description1 = U.RandomString(8)
	transformations1 = &postgres.Jsonb{json.RawMessage(`{"cl":"kpi","for":"a/b","qG":[{"ca":"channels","dc":"google_ads_metrics","fil":[],"gBy":[],"me":["clicks"],"na":"a","pgUrl":"","tz":"Australia/Sydney"},{"ca":"channels","dc":"google_ads_metrics","fil":[],"gBy":[],"me":["impressions"],"na":"b","pgUrl":"","tz":"Australia/Sydney"}]}`)}
	w = sendCreateCustomMetricWithPercentage(a, project.ID, agent, transformations1, name1, description1, "google_ads_metrics", 2)
	assert.Equal(t, http.StatusOK, w.Code)

	t.Run("Percentage Derived kpi query without gbt, no group bys", func(t *testing.T) {
		query = model.KPIQuery{
			Category:        "channels",
			DisplayCategory: "google_ads_metrics",
			PageUrl:         "",
			Metrics:         []string{name1},
			GroupBy:         []M.KPIGroupBy{},
			From:            1659312000,
			To:              1659657600,
			QueryType:       "derived",
		}
		kpiQueryGroup = model.KPIQueryGroup{
			Class:         "kpi",
			Queries:       []model.KPIQuery{query},
			GlobalFilters: []model.KPIFilter{},
			GlobalGroupBy: []model.KPIGroupBy{},
		}

		result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
			C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, result[0].Headers, []string{"google_ads_metrics_" + name1})
		assert.Equal(t, len(result[0].Rows), 1)
		assert.Equal(t, result[0].Rows[0][0], float64(20)) //(300/1500)*100
	})

	t.Run("Percentage Derived kpi query with gbt, no group bys", func(t *testing.T) {
		query = model.KPIQuery{
			Category:         "channels",
			DisplayCategory:  "google_ads_metrics",
			PageUrl:          "",
			Metrics:          []string{name1},
			GroupBy:          []M.KPIGroupBy{},
			From:             1659312000,
			To:               1659657600,
			GroupByTimestamp: "date",
			QueryType:        "derived",
		}
		kpiQueryGroup = model.KPIQueryGroup{
			Class:         "kpi",
			Queries:       []model.KPIQuery{query},
			GlobalFilters: []model.KPIFilter{},
			GlobalGroupBy: []model.KPIGroupBy{},
		}

		result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
			C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, result[0].Headers, []string{"datetime", "google_ads_metrics_" + name1})
		assert.Equal(t, len(result[0].Rows), 5)
		assert.Equal(t, result[0].Rows[1][1], float64(20))
	})

	t.Run("Percentage Derived kpi query without gbt, with group bys", func(t *testing.T) {
		query = model.KPIQuery{
			Category:        "channels",
			DisplayCategory: "google_ads_metrics",
			PageUrl:         "",
			Metrics:         []string{name1},
			GroupBy:         []M.KPIGroupBy{},
			From:            1659312000,
			To:              1659657600,
			QueryType:       "derived",
		}
		kpiQueryGroup = model.KPIQueryGroup{
			Class:         "kpi",
			Queries:       []model.KPIQuery{query},
			GlobalFilters: []model.KPIFilter{},
			GlobalGroupBy: []model.KPIGroupBy{
				{
					ObjectType:       "campaign",
					PropertyName:     "campaign_name",
					PropertyDataType: "categorical",
					Entity:           "",
				},
			},
		}

		result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
			C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, result[0].Headers, []string{"campaign_name", "google_ads_metrics_" + name1})
		assert.Equal(t, len(result[0].Rows), 2)
		assert.Equal(t, result[0].Rows[0], []interface{}{"test1", float64(10)})
		assert.Equal(t, result[0].Rows[1], []interface{}{"test2", float64(40)})
	})

	t.Run("Percentage Derived kpi query with gbt, with group bys", func(t *testing.T) {
		query = model.KPIQuery{
			Category:         "channels",
			DisplayCategory:  "google_ads_metrics",
			PageUrl:          "",
			Metrics:          []string{name1},
			GroupBy:          []M.KPIGroupBy{},
			From:             1659312000,
			To:               1659657600,
			GroupByTimestamp: "date",
			QueryType:        "derived",
		}
		kpiQueryGroup = model.KPIQueryGroup{
			Class:         "kpi",
			Queries:       []model.KPIQuery{query},
			GlobalFilters: []model.KPIFilter{},
			GlobalGroupBy: []model.KPIGroupBy{
				{
					ObjectType:       "campaign",
					PropertyName:     "campaign_name",
					PropertyDataType: "categorical",
					Entity:           "",
				},
			},
		}

		result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
			C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, result[0].Headers, []string{"datetime", "campaign_name", "google_ads_metrics_" + name1})
		assert.Equal(t, len(result[0].Rows), 10)
		assert.Equal(t, result[0].Rows[2][1], "test1")
		assert.Equal(t, result[0].Rows[2][2], float64(10))
		assert.Equal(t, result[0].Rows[3][1], "test2")
		assert.Equal(t, result[0].Rows[3][2], float64(40))
	})
}

func TestDerivedKPIForCustomKPI(t *testing.T) {
	a := gin.Default()
	H.InitAppRoutes(a)
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"

	r2 := gin.Default()
	H.InitDataServiceRoutes(r2)
	model.SetSmartPropertiesReservedNames()

	project, customerAccountID, agent, statusCode := createProjectAndAddAdwordsDocument(t, r2)
	if statusCode != http.StatusAccepted {
		assert.Equal(t, false, true)
		return
	}
	rCustomerUserId := U.RandomLowerAphaNumString(15)
	joinTime := U.UnixTimeBeforeDuration(time.Hour * 1)
	properties1 := postgres.Jsonb{RawMessage: json.RawMessage([]byte(
		fmt.Sprintf(`{"country": "us", "age": 20, "$hubspot_amount": 200, "$hubspot_datefield1": %d, "paid": true}`, joinTime)))}
	createUserID1, _ := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: rCustomerUserId, Properties: properties1, JoinTimestamp: joinTime, Source: model.GetRequestSourcePointer(model.UserSourceHubspot)})

	startTimestamp := U.UnixTimeBeforeDuration(time.Hour * 1)
	currentDate := time.Now().UTC().Format(U.DATETIME_FORMAT_YYYYMMDD)
	currentDateInt, _ := strconv.ParseInt(currentDate, 10, 64)
	stepTimestamp := startTimestamp

	payload := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, "user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`, "s0", createUserID1, stepTimestamp, "A", 1234)
	w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response := DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	timestamp := U.UnixTimeBeforeDuration(30 * 24 * time.Hour)
	eventName := U.RandomLowerAphaNumString(10)
	w = ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"event_name": "%s", "timestamp": %d, "event_properties": {"$referrer": "https://example.com/abc?ref=1", "$referrer_url": "https://example.com/abc", "$referrer_domain": "example.com", "$page_url": "https://example.com/xyz", "$page_raw_url": "https://example.com/xyz?utm_campaign=google", "$page_domain": "example.com", "$page_load_time": 100, "$page_spent_time": 120, "$qp_utm_campaign": "google", "$qp_utm_campaignid": "12345", "$qp_utm_ad": "ad_2021_1", "$qp_utm_ad_id": "9876543210", "$qp_utm_source": "google", "$qp_utm_medium": "email", "$qp_utm_keyword": "analytics", "$qp_utm_matchtype": "exact", "$qp_utm_content": "analytics", "$qp_utm_adgroup": "ad-xxx", "$qp_utm_adgroupid": "xyz123", "$qp_utm_creative": "creative-xxx", "$qp_gclid": "xxx123", "$qp_fbclid": "zzz123"}, "user_properties": {"$platform": "web", "$browser": "Mozilla", "$browser_version": "v0.1", "$browser_with_version": "Mozilla_v0.1", "$user_agent": "browser", "$os": "Linux", "$os_version": "v0.1", "$os_with_version": "Linux_v0.1", "$country": "india", "$region": "karnataka", "$city": "bengaluru", "$timezone": "Asia/Calcutta"}}`,
			eventName, timestamp)), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	w = ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"event_name": "%s", "timestamp": %d, "event_properties": {"$referrer": "https://example.com/abc?ref=1", "$referrer_url": "https://example.com/abc", "$referrer_domain": "example.com", "$page_url": "https://example.com/xyz", "$page_raw_url": "https://example.com/xyz?utm_campaign=google", "$page_domain": "example.com", "$page_load_time": 100, "$page_spent_time": 120, "$qp_utm_campaign": "google", "$qp_utm_campaignid": "12345", "$qp_utm_ad": "ad_2021_1", "$qp_utm_ad_id": "9876543210", "$qp_utm_source": "google", "$qp_utm_medium": "email", "$qp_utm_keyword": "analytics", "$qp_utm_matchtype": "exact", "$qp_utm_content": "analytics", "$qp_utm_adgroup": "ad-xxx", "$qp_utm_adgroupid": "xyz123", "$qp_utm_creative": "creative-xxx", "$qp_gclid": "xxx123", "$qp_fbclid": "zzz123"}, "user_properties": {"$platform": "web", "$browser": "Mozilla", "$browser_version": "v0.1", "$browser_with_version": "Mozilla_v0.1", "$user_agent": "browser", "$os": "Linux", "$os_version": "v0.1", "$os_with_version": "Linux_v0.1", "$country": "india", "$region": "karnataka", "$city": "bengaluru", "$timezone": "Asia/Calcutta"}}`,
			"123testing", timestamp)), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	_, err := TaskSession.AddSession([]int64{project.ID}, timestamp-60, 0, 0, 0, 1, 1)
	assert.Nil(t, err)

	adwordsDocuments := []M.AdwordsDocument{
		{ID: "1", Timestamp: currentDateInt, ProjectID: project.ID, CustomerAccountID: customerAccountID, TypeAlias: "campaign_performance_report",
			Value: &postgres.Jsonb{json.RawMessage(`{"cost": "11","clicks": "100","campaign_id":"1","impressions": "1000", "campaign_name": "test1"}`)}},
		{ID: "2", Timestamp: currentDateInt, ProjectID: project.ID, CustomerAccountID: customerAccountID, TypeAlias: "campaign_performance_report",
			Value: &postgres.Jsonb{json.RawMessage(`{"cost": "12","clicks": "200","campaign_id":"2","impressions": "500", "campaign_name": "test2"}`)}},
	}
	for _, adwordsDocument := range adwordsDocuments {
		status := store.GetStore().CreateAdwordsDocument(&adwordsDocument)
		assert.Equal(t, http.StatusCreated, status)
	}

	t.Run("Profiles Query with no groupby and no filter.", func(t *testing.T) {

		name1 := "name1"
		description1 := U.RandomString(8)
		transformations1 := &postgres.Jsonb{json.RawMessage(`{"agFn": "sum", "agPr": "$hubspot_amount", "agPrTy": "categorical", "fil": [], "daFie": "$hubspot_datefield1"}`)}
		w = sendCreateCustomMetric(a, project.ID, agent, transformations1, name1, description1, "hubspot_contacts", 1)
		assert.Equal(t, http.StatusOK, w.Code)

		name2 := "dname2"
		description2 := U.RandomString(8)
		transformations2 := &postgres.Jsonb{json.RawMessage(`{"cl":"kpi","for":"a/b","qG":[{"ca":"channels","dc":"google_ads_metrics","fil":[],"gBy":[],"me":["impressions"],"na":"a","pgUrl":"","tz":"Australia/Sydney"},{"ca":"profiles","dc":"hubspot_contacts","fil":[],"gBy":[],"me":["name1"],"na":"b","pgUrl":"","tz":"Australia/Sydney"}]}`)}
		w = sendCreateCustomMetric(a, project.ID, agent, transformations2, name2, description2, "", 2)
		assert.Equal(t, http.StatusOK, w.Code)

		query := model.KPIQuery{
			Category:         "events",
			DisplayCategory:  "others",
			PageUrl:          "",
			Metrics:          []string{name2},
			Filters:          []model.KPIFilter{},
			From:             startTimestamp,
			To:               startTimestamp + 40,
			GroupByTimestamp: "date",
			QueryType:        "derived",
		}
		query1 := model.KPIQuery{}
		U.DeepCopy(&query, &query1)
		query1.GroupByTimestamp = ""

		kpiQueryGroup := model.KPIQueryGroup{
			Class:         "kpi",
			Queries:       []model.KPIQuery{query, query1},
			GlobalFilters: []model.KPIFilter{},
			GlobalGroupBy: []model.KPIGroupBy{},
		}

		result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
			C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, result[0].Headers, []string{"datetime", name2})
		assert.Equal(t, len(result[0].Rows), 1)
		assert.Equal(t, result[1].Headers, []string{name2})
		assert.Equal(t, len(result[1].Rows), 1)
		assert.Equal(t, result[0].Rows[0][1], float64(7.5))
	})

}

func TestKpiAnalyticsHandler(t *testing.T) {
	a := gin.Default()
	H.InitAppRoutes(a)

	r := gin.Default()
	H.InitSDKServiceRoutes(r)

	project, agent, _ := SetupProjectWithAgentDAO()
	assert.NotNil(t, project)

	query := model.KPIQuery{
		Category:        "events",
		DisplayCategory: "page_views",
		PageUrl:         "s0",
		Metrics:         []string{"page_views"},
		GroupBy:         []M.KPIGroupBy{},
		From:            20210801,
		To:              20210801 + 40,
	}

	kpiQueryGroup := model.KPIQueryGroup{
		Class:         "kpi",
		Queries:       []model.KPIQuery{query},
		GlobalFilters: []model.KPIFilter{},
		GlobalGroupBy: []model.KPIGroupBy{},
	}

	sendKPIAnalyticsQueryReq(a, project.ID, agent, kpiQueryGroup)
}

func sendKPIAnalyticsQueryReq(r *gin.Engine, projectId int64, agent *M.Agent, kpiqueryGroup model.KPIQueryGroup) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}

	rb := C.NewRequestBuilderWithPrefix(http.MethodPost, fmt.Sprintf("/projects/%d/v1/kpi/query", projectId)).
		WithPostParams(kpiqueryGroup).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error creating create dashboard_unit req.")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestKpiFilterValuesHandler(t *testing.T) {
	a := gin.Default()
	H.InitAppRoutes(a)

	r := gin.Default()
	H.InitSDKServiceRoutes(r)

	project, agent, _ := SetupProjectWithAgentDAO()
	assert.NotNil(t, project)

	filterValueRequest := v1.KPIFilterValuesRequest{
		Category:     "events",
		ObjectType:   "$session",
		PropertyName: "$medium",
		Entity:       "event",
	}

	sendKPIFilterValuesReq(a, project.ID, agent, filterValueRequest)
}

func sendKPIFilterValuesReq(r *gin.Engine, projectId int64, agent *M.Agent, filterValueRequest v1.KPIFilterValuesRequest) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}

	rb := C.NewRequestBuilderWithPrefix(http.MethodPost, fmt.Sprintf("/projects/%d/v1/kpi/filter_values", projectId)).
		WithPostParams(filterValueRequest).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error creating create dashboard_unit req.")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	log.Warn(jsonResponse)
	return w
}

func TestEventNamesByTypeHandler(t *testing.T) {
	a := gin.Default()
	H.InitAppRoutes(a)

	r := gin.Default()
	H.InitSDKServiceRoutes(r)

	project, agent, _ := SetupProjectWithAgentDAO()
	assert.NotNil(t, project)

	sendEventNamesTypeQueryReq(a, project.ID, agent)
}

func sendEventNamesTypeQueryReq(r *gin.Engine, projectId int64, agent *M.Agent) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}

	rb := C.NewRequestBuilderWithPrefix(http.MethodGet, fmt.Sprintf("/projects/%d/v1/event_names/page_views", projectId)).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error creating create dashboard_unit req.")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	log.Warn(jsonResponse)
	return w
}

func TestKPIChannelsMissingTimestamps(t *testing.T) {
	r := gin.Default()
	H.InitDataServiceRoutes(r)
	model.SetSmartPropertiesReservedNames()

	a := gin.Default()
	H.InitAppRoutes(a)

	project, customerAccountID, _, statusCode := createProjectAndAddAdwordsDocument(t, r)
	if statusCode != http.StatusAccepted {
		assert.Equal(t, false, true)
		return
	}

	adwordsDocuments := []M.AdwordsDocument{
		{ID: "1", Timestamp: 20220802, ProjectID: project.ID, CustomerAccountID: customerAccountID, TypeAlias: "campaign_performance_report",
			Value: &postgres.Jsonb{RawMessage: json.RawMessage(`{"cost": "11","clicks": "100","campaign_id":"1","impressions": "1000", "campaign_name": "test1"}`)}},
		{ID: "2", Timestamp: 20220803, ProjectID: project.ID, CustomerAccountID: customerAccountID, TypeAlias: "campaign_performance_report",
			Value: &postgres.Jsonb{RawMessage: json.RawMessage(`{"cost": "12","clicks": "200","campaign_id":"2","impressions": "500", "campaign_name": "test2"}`)}},
	}
	for _, adwordsDocument := range adwordsDocuments {
		status := store.GetStore().CreateAdwordsDocument(&adwordsDocument)
		assert.Equal(t, http.StatusCreated, status)
	}

	query := model.KPIQuery{
		Category:         "channels",
		DisplayCategory:  "google_ads_metrics",
		PageUrl:          "",
		Metrics:          []string{"impressions"},
		GroupBy:          []M.KPIGroupBy{},
		From:             1659312000, // 1st Aug, 2022
		To:               1659657600, // 5th Aug, 2022
		GroupByTimestamp: "date",
		Timezone:         "Asia/Kolkata",
	}
	kpiQueryGroup := model.KPIQueryGroup{
		Class:         "kpi",
		Queries:       []model.KPIQuery{query},
		GlobalFilters: []model.KPIFilter{},
		GlobalGroupBy: []model.KPIGroupBy{},
	}
	result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
		C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	assert.Equal(t, result[0].Headers, []string{"datetime", "google_ads_metrics_impressions"})
	assert.Equal(t, len(result[0].Rows), 5)
	assert.Equal(t, result[0].Rows[0][1], 0)
	assert.Equal(t, result[0].Rows[1][1], float64(1000))
	assert.Equal(t, result[0].Rows[2][1], float64(500))
	assert.Equal(t, result[0].Rows[3][1], 0)
	assert.Equal(t, result[0].Rows[4][1], 0)

	kpiQueryGroup = model.KPIQueryGroup{
		Class:         "kpi",
		Queries:       []model.KPIQuery{query},
		GlobalFilters: []model.KPIFilter{},
		GlobalGroupBy: []model.KPIGroupBy{
			{
				ObjectType:       "campaign",
				PropertyName:     "campaign_name",
				PropertyDataType: "categorical",
				Entity:           "",
			},
		},
	}
	result, statusCode = store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
		C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	assert.Equal(t, result[0].Headers, []string{"datetime", "campaign_name", "google_ads_metrics_impressions"})
	assert.Equal(t, len(result[0].Rows), 10)
	assert.Equal(t, result[0].Rows[2], []interface{}{"2022-08-02T00:00:00+05:30", "test1", float64(1000)})
	assert.Equal(t, result[0].Rows[5], []interface{}{"2022-08-03T00:00:00+05:30", "test2", float64(500)})
}

func TestKPIQueryGroupHandlerForPropertyValueLabels(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)
	assert.NotNil(t, agent)

	startTimestamp := time.Now().Unix()

	hubspotDisplayNameLabels := map[string]string{
		"OFFFLINE":       "Offline",
		"PAID_SEARCH":    "Paid Search",
		"DIRECT_TRAFFIC": "Direct Traffic",
		"ORGANIC_SEARCH": "Organic Search",
		"SOCIAL_MEDIA":   "Social Media",
	}

	for value, label := range hubspotDisplayNameLabels {
		status := store.GetStore().CreateOrUpdateDisplayNameLabel(project.ID, "hubspot", "$hubspot_contact_hs_latest_source", value, label)
		assert.Equal(t, http.StatusCreated, status)
	}

	// create new hubspot document
	jsonContactModel := `{
		"vid": %d,
		"addedAt": %d,
		"properties": {
		  	"firstname": { "value": "%s" },
		  	"lastname": { "value": "%s" },
			"createdate": { "value": "%d" },
		  	"lastmodifieddate": { "value": "%d" },
			"hs_latest_source": { "value": "%s" }
		},
		"identity-profiles": [
			{
				"vid": %d,
				"identities": [
					{
					  "type": "EMAIL",
					  "value": "%s"
					},
					{
						"type": "LEAD_GUID",
						"value": "%s"
					}
				]
			}
		]
	}`

	latestSources := []string{"ORGANIC_SEARCH", "DIRECT_TRAFFIC", "PAID_SOCIAL"}
	hubspotDocuments := make([]*model.HubspotDocument, 0)
	for i := 0; i < len(latestSources); i++ {
		documentID := i
		createdTime := startTimestamp*1000 + int64(i*100)
		updatedTime := createdTime + 200
		cuid := U.RandomString(10)
		leadGUID := U.RandomString(15)
		jsonContact := fmt.Sprintf(jsonContactModel, documentID, createdTime, "Sample", fmt.Sprintf("Test %d", i), createdTime, updatedTime, latestSources[i], documentID, cuid, leadGUID)

		document := model.HubspotDocument{
			TypeAlias: model.HubspotDocumentTypeNameContact,
			Value:     &postgres.Jsonb{json.RawMessage(jsonContact)},
		}
		hubspotDocuments = append(hubspotDocuments, &document)
	}

	status := store.GetStore().CreateHubspotDocumentInBatch(project.ID, model.HubspotDocumentTypeContact, hubspotDocuments, 3)
	assert.Equal(t, http.StatusCreated, status)

	// execute sync job
	allStatus, _ := IntHubspot.Sync(project.ID, 1, time.Now().Unix(), nil, "", 50)
	for i := range allStatus {
		assert.Equal(t, U.CRM_SYNC_STATUS_SUCCESS, allStatus[i].Status)
	}

	customMetric := &model.CustomMetric{
		ProjectID:              project.ID,
		Name:                   "Contact Created",
		Description:            "Custom KPI",
		TypeOfQuery:            model.ProfileQueryType,
		SectionDisplayCategory: "hubspot_contacts",
		Transformations:        &postgres.Jsonb{json.RawMessage(`{"agFn":"unique","daFie":"$hubspot_contact_createdate","fil":[]}`)},
	}
	customMetric, errString, statusCode := store.GetStore().CreateCustomMetric(*customMetric)
	assert.NotNil(t, customMetric)
	assert.Equal(t, "", errString)
	assert.Equal(t, http.StatusCreated, statusCode)

	kpiQueryGroup := model.KPIQueryGroup{
		Class: model.QueryClassKPI,
		Queries: []model.KPIQuery{
			model.KPIQuery{
				Category:         "profiles",
				DisplayCategory:  "hubspot_contacts",
				Metrics:          []string{"Contact Created"},
				From:             startTimestamp,
				To:               startTimestamp + 500,
				QueryType:        "custom",
				GroupByTimestamp: "date",
			},
		},
		GlobalGroupBy: []model.KPIGroupBy{
			model.KPIGroupBy{
				PropertyName:     "$hubspot_contact_hs_latest_source",
				PropertyDataType: "categorical",
				Entity:           model.PropertyEntityUser,
			},
		},
	}
	kpiQueryGroupJson, err := U.EncodeStructTypeToPostgresJsonb(kpiQueryGroup)
	assert.Nil(t, err)
	assert.NotNil(t, kpiQueryGroupJson)

	C.GetConfig().EnableSyncReferenceFieldsByProjectID = ""
	w := sendKPIAnalyticsQueryReq(r, project.ID, agent, kpiQueryGroup)
	jsonResponse, err := ioutil.ReadAll(w.Body)
	assert.Nil(t, err)
	kpiQueryGroupResult := struct {
		QueryResult []model.QueryResult `json:"result"`
		Query       model.KPIQueryGroup `json:"query"`
		IsSharable  bool                `json:"sharable"`
	}{}
	json.Unmarshal(jsonResponse, &kpiQueryGroupResult)
	assert.Equal(t, http.StatusOK, w.Code)

	assert.NotNil(t, kpiQueryGroupResult)
	assert.Equal(t, false, kpiQueryGroupResult.IsSharable)
	assert.NotNil(t, kpiQueryGroupResult.QueryResult)
	assert.NotNil(t, kpiQueryGroupResult.Query)

	assert.Equal(t, 1, len(kpiQueryGroupResult.QueryResult))

	startTimestampString := time.Unix(U.GetBeginningOfDayTimestamp(time.Now().Unix()), 0).Format(U.DATETIME_FORMAT_DB_WITH_TIMEZONE)

	assert.Equal(t, kpiQueryGroupResult.QueryResult[0].Headers, []string{"datetime", "$hubspot_contact_hs_latest_source", "Contact Created"})
	assert.Equal(t, len(latestSources), len(kpiQueryGroupResult.QueryResult[0].Rows))
	for i := range kpiQueryGroupResult.QueryResult[0].Rows {
		assert.Equal(t, startTimestampString, kpiQueryGroupResult.QueryResult[0].Rows[i][0])
		assert.Contains(t, latestSources, kpiQueryGroupResult.QueryResult[0].Rows[i][1])
		assert.Equal(t, float64(1), kpiQueryGroupResult.QueryResult[0].Rows[i][2])
	}

	C.GetConfig().EnableSyncReferenceFieldsByProjectID = "*"
	w = sendKPIAnalyticsQueryReq(r, project.ID, agent, kpiQueryGroup)
	jsonResponse, err = ioutil.ReadAll(w.Body)
	assert.Nil(t, err)

	queryResultsFromCache := make([]model.QueryResult, 0)
	json.Unmarshal(jsonResponse, &queryResultsFromCache)
	assert.Equal(t, http.StatusOK, w.Code)

	assert.NotNil(t, queryResultsFromCache)

	assert.Equal(t, 1, len(queryResultsFromCache))

	startTimestampString = time.Unix(U.GetBeginningOfDayTimestamp(time.Now().Unix()), 0).Format(U.DATETIME_FORMAT_DB_WITH_TIMEZONE)

	assert.Equal(t, queryResultsFromCache[0].Headers, []string{"datetime", "$hubspot_contact_hs_latest_source", "Contact Created"})
	assert.Equal(t, len(latestSources), len(queryResultsFromCache[0].Rows))
	for i := range queryResultsFromCache[0].Rows {
		assert.Equal(t, startTimestampString, queryResultsFromCache[0].Rows[i][0])
		assert.Contains(t, []string{"Organic Search", "Direct Traffic", "PAID_SOCIAL"}, queryResultsFromCache[0].Rows[i][1])
		assert.Equal(t, float64(1), queryResultsFromCache[0].Rows[i][2])
	}

	// Dashboard Cache Response
	dashboard, statusCode := store.GetStore().CreateDashboard(project.ID, agent.UUID, &model.Dashboard{Name: U.RandomString(5), Type: model.DashboardTypeProjectVisible})
	assert.Equal(t, http.StatusCreated, statusCode)
	assert.NotNil(t, dashboard)

	dashboardQuery, errCode, errMsg := store.GetStore().CreateQuery(project.ID, &model.Queries{
		ProjectID: project.ID,
		Type:      model.QueryTypeDashboardQuery,
		Query:     *kpiQueryGroupJson,
	})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Empty(t, errMsg)
	assert.NotNil(t, dashboardQuery)

	dashboardUnit, statusCode, errMsg := store.GetStore().CreateDashboardUnit(project.ID, agent.UUID, &model.DashboardUnit{
		DashboardId:  dashboard.ID,
		QueryId:      dashboardQuery.ID,
		Presentation: model.PresentationLine,
	})
	assert.Equal(t, "", errMsg)
	assert.Equal(t, http.StatusCreated, statusCode)
	assert.NotNil(t, dashboardUnit)

	meta := model.CacheMeta{
		Timezone:       string(U.TimeZoneStringIST),
		From:           kpiQueryGroup.Queries[0].From,
		To:             kpiQueryGroup.Queries[0].To,
		RefreshedAt:    U.TimeNowIn(U.TimeZoneStringIST).Unix(),
		LastComputedAt: U.TimeNowIn(U.TimeZoneStringIST).Unix(),
		Preset:         "",
	}
	for i := range kpiQueryGroupResult.QueryResult {
		kpiQueryGroupResult.QueryResult[i].CacheMeta = meta
	}

	model.SetCacheResultByDashboardIdAndUnitId(kpiQueryGroupResult.QueryResult, project.ID, dashboard.ID, dashboardUnit.ID, meta.From, meta.To, U.TimeZoneString(meta.Timezone), meta)

	shouldReturn, resCode, resMsgInt := helpers.GetResponseIfCachedDashboardQuery("", project.ID, dashboard.ID, dashboardUnit.ID, meta.From, meta.To, U.TimeZoneString(meta.Timezone))
	assert.Equal(t, http.StatusOK, resCode)
	assert.Equal(t, true, shouldReturn)
	assert.NotNil(t, resMsgInt)

	resMsgByte, err := json.Marshal(resMsgInt)
	assert.Nil(t, err)
	var cachedDashboardQueryResponse helpers.DashboardQueryResponsePayload
	err = json.Unmarshal(resMsgByte, &cachedDashboardQueryResponse)
	assert.Nil(t, err)

	assert.NotNil(t, cachedDashboardQueryResponse)
	assert.Equal(t, true, cachedDashboardQueryResponse.Cache)
	assert.Equal(t, meta.RefreshedAt, cachedDashboardQueryResponse.RefreshedAt)
	assert.Equal(t, meta.Timezone, cachedDashboardQueryResponse.TimeZone)
	assert.NotNil(t, cachedDashboardQueryResponse.CacheMeta)
	assert.NotNil(t, cachedDashboardQueryResponse.Result)

	result, ok := cachedDashboardQueryResponse.Result.([]interface{})
	assert.Equal(t, true, ok)
	assert.NotNil(t, result)

	var responseQueryResults []model.QueryResult
	for i := range result {
		resultByte, err := json.Marshal(result[i])
		assert.Nil(t, err)
		var queryResult model.QueryResult
		err = json.Unmarshal(resultByte, &queryResult)
		assert.Nil(t, err)

		responseQueryResults = append(responseQueryResults, queryResult)
	}
	assert.Equal(t, 1, len(responseQueryResults))

	assert.Equal(t, "datetime", responseQueryResults[0].Headers[0])
	for i := range responseQueryResults[0].Rows {
		assert.Equal(t, startTimestampString, responseQueryResults[0].Rows[i][0])
	}

	assert.Equal(t, "$hubspot_contact_hs_latest_source", responseQueryResults[0].Headers[1])
	for i := range responseQueryResults[0].Rows {
		assert.Contains(t, latestSources, responseQueryResults[0].Rows[i][1])
	}

	assert.Equal(t, "Contact Created", responseQueryResults[0].Headers[2])
	for i := range responseQueryResults[0].Rows {
		assert.Equal(t, float64(1), responseQueryResults[0].Rows[i][2])
	}

	resMsgIntTransformed, err := helpers.TransformQueryCacheResponseColumnValuesToLabel(project.ID, resMsgInt)
	assert.Nil(t, err)
	assert.NotNil(t, resMsgIntTransformed)

	resMsgTransformedByte, err := json.Marshal(resMsgIntTransformed)
	assert.Nil(t, err)
	cachedDashboardQueryResponse = helpers.DashboardQueryResponsePayload{}
	err = json.Unmarshal(resMsgTransformedByte, &cachedDashboardQueryResponse)
	assert.Nil(t, err)

	assert.NotNil(t, cachedDashboardQueryResponse)
	assert.Equal(t, true, cachedDashboardQueryResponse.Cache)
	assert.Equal(t, meta.RefreshedAt, cachedDashboardQueryResponse.RefreshedAt)
	assert.Equal(t, meta.Timezone, cachedDashboardQueryResponse.TimeZone)
	assert.NotNil(t, cachedDashboardQueryResponse.CacheMeta)
	assert.NotNil(t, cachedDashboardQueryResponse.Result)

	result, ok = cachedDashboardQueryResponse.Result.([]interface{})
	assert.Equal(t, true, ok)
	assert.NotNil(t, result)

	responseQueryResults = make([]model.QueryResult, 0)
	for i := range result {
		resultByte, err := json.Marshal(result[i])
		assert.Nil(t, err)
		var queryResult model.QueryResult
		err = json.Unmarshal(resultByte, &queryResult)
		assert.Nil(t, err)

		responseQueryResults = append(responseQueryResults, queryResult)
	}
	assert.Equal(t, 1, len(responseQueryResults))

	assert.Equal(t, "datetime", responseQueryResults[0].Headers[0])
	for i := range responseQueryResults[0].Rows {
		assert.Equal(t, startTimestampString, responseQueryResults[0].Rows[i][0])
	}

	assert.Equal(t, "$hubspot_contact_hs_latest_source", responseQueryResults[0].Headers[1])
	for i := range responseQueryResults[0].Rows {
		assert.Contains(t, []string{"Organic Search", "Direct Traffic", "PAID_SOCIAL"}, responseQueryResults[0].Rows[i][1])
	}

	assert.Equal(t, "Contact Created", responseQueryResults[0].Headers[2])
	for i := range responseQueryResults[0].Rows {
		assert.Equal(t, float64(1), responseQueryResults[0].Rows[i][2])
	}
}
