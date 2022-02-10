package tests

import (
	"encoding/json"
	C "factors/config"
	Const "factors/constants"
	H "factors/handler"
	"factors/handler/helpers"
	v1 "factors/handler/v1"
	"factors/model/model"
	M "factors/model/model"
	"factors/model/store"
	TaskSession "factors/task/session"
	U "factors/util"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestKpiAnalytics(t *testing.T) {
	a := gin.Default()
	H.InitAppRoutes(a)

	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"

	r2 := gin.Default()
	H.InitDataServiceRoutes(r2)
	Const.SetSmartPropertiesReservedNames()

	project, customerAccountID, _, statusCode := createProjectAndAddAdwordsDocument(t, r2)
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

	_, err := TaskSession.AddSession([]uint64{project.ID}, timestamp-60, 0, 0, 0, 1, 1)
	assert.Nil(t, err)

	t.Run("abc", func(t *testing.T) {

		query := model.KPIQuery{

			Category:        "events",
			DisplayCategory: "page_views",
			PageUrl:         "s0",
			//Metrics:         []string{"page_views", "unique_users"},
			Metrics: []string{"page_views"},
			Filters: []model.KPIFilter{
				{
					PropertyName:     "user_id",
					PropertyDataType: "categorical",
					Entity:           "user",
					Condition:        "equals",
					Value:            "1",
					LogicalOp:        "AND",
				},
			},
			From: startTimestamp,
			To:   startTimestamp + 40,
		}

		kpiQueryGroup := model.KPIQueryGroup{
			Class:         "kpi",
			Queries:       []model.KPIQuery{query},
			GlobalFilters: []model.KPIFilter{},
			GlobalGroupBy: []model.KPIGroupBy{},
		}

		result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup)
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, result[0].Headers, []string{"page_views"})
		assert.Equal(t, len(result[0].Rows), 1)
		assert.Equal(t, result[0].Rows[0][0], int(0))
	})

	t.Run("abc1", func(t *testing.T) {

		query := model.KPIQuery{

			Category:        "events",
			DisplayCategory: "page_views",
			PageUrl:         "s0",
			//Metrics:         []string{"page_views", "unique_users"},
			Metrics:          []string{"page_views"},
			Filters:          nil,
			From:             startTimestamp,
			To:               startTimestamp + 40,
			GroupByTimestamp: "date",
		}

		kpiQueryGroup := model.KPIQueryGroup{
			Class:         "kpi",
			Queries:       []model.KPIQuery{query},
			GlobalFilters: []model.KPIFilter{},
			GlobalGroupBy: []model.KPIGroupBy{},
		}

		result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup)
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, result[0].Headers, []string{"datetime", "page_views"})
		assert.Equal(t, len(result[0].Rows), 1)
		assert.Equal(t, result[0].Rows[0][1], float64(1))
	})

	t.Run("abc2", func(t *testing.T) {

		query1 := model.KPIQuery{

			Category:        "events",
			DisplayCategory: "page_views",
			PageUrl:         "s0",
			Metrics:         []string{"page_views"},
			// Metrics:  []string{"page_views"},
			Filters:  nil,
			From:     startTimestamp,
			To:       startTimestamp + 40,
			Timezone: "Asia/Kolkata",
			//GroupByTimestamp: "date",
		}
		query2 := model.KPIQuery{
			Category:        "events",
			DisplayCategory: "page_views",
			PageUrl:         "s0",
			Metrics:         []string{"page_views"},
			// Metrics:          []string{"page_views"},
			Filters:          nil,
			From:             startTimestamp,
			To:               startTimestamp + 40,
			Timezone:         "Asia/Kolkata",
			GroupByTimestamp: "date",
		}

		query3 := model.KPIQuery{

			Category:        "events",
			DisplayCategory: "page_views",
			PageUrl:         "s0",
			Metrics:         []string{"unique_users"},
			Filters:         nil,
			From:            startTimestamp,
			To:              startTimestamp + 40,
			Timezone:        "Asia/Kolkata",
			//GroupByTimestamp: "date",
		}
		query4 := model.KPIQuery{
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

		result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup)
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, result[0].Headers, []string{"datetime", "user_id", "page_views", "unique_users"})
		assert.Equal(t, len(result[0].Rows), 1)
		assert.Equal(t, result[0].Rows[0][1], "$none")
		assert.Equal(t, result[0].Rows[0][2], float64(1))
		assert.Equal(t, result[0].Rows[0][3], float64(1))
	})

	t.Run("abc3", func(t *testing.T) {

		query := model.KPIQuery{

			Category:        "events",
			DisplayCategory: "website_session",
			Metrics:         []string{"average_initial_page_load_time"},
			Filters:         nil,
			From:            timestamp,
			To:              timestamp + (40 * 24 * 60 * 60),
		}

		kpiQueryGroup := model.KPIQueryGroup{
			Class:         "kpi",
			Queries:       []model.KPIQuery{query},
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

		result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup)
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, result[0].Headers, []string{"user_id", "average_initial_page_load_time"})
		assert.Equal(t, len(result[0].Rows), 1)
		assert.Equal(t, result[0].Rows[0][0], "$none")
		assert.Equal(t, result[0].Rows[0][1], float64(0))
	})

	t.Run("abc4", func(t *testing.T) {

		query := model.KPIQuery{

			Category:        "channels",
			DisplayCategory: "adwords_metrics",
			Metrics:         []string{"impressions"},
			Filters:         nil,
			From:            startTimestamp,
			To:              startTimestamp + 40,
		}

		kpiQueryGroup := model.KPIQueryGroup{
			Class:         "kpi",
			Queries:       []model.KPIQuery{query},
			GlobalFilters: []model.KPIFilter{},
			GlobalGroupBy: []model.KPIGroupBy{},
		}

		result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup)
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, result[0].Headers, []string{"impressions"})
		assert.Equal(t, len(result[0].Rows), 1)
		assert.Equal(t, result[0].Rows[0][0], int(0))
	})
}

func TestKpiAnalyticsForProfile(t *testing.T) {
	a := gin.Default()
	H.InitAppRoutes(a)

	project, agent, _ := SetupProjectWithAgentDAO()
	rCustomerUserId := U.RandomLowerAphaNumString(15)
	properties1 := postgres.Jsonb{RawMessage: json.RawMessage([]byte(`{"country": "india", "age": 30, "$hubspot_amount": 300, "$hubspot_datefield1": 1640975425,  "paid": true}`))}
	properties2 := postgres.Jsonb{RawMessage: json.RawMessage([]byte(`{"country": "us", "age": 20, "$hubspot_amount": 200, "$hubspot_datefield1": 1640975425, "paid": true}`))}
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
	w := sendCreateCustomMetric(a, project.ID, agent, transformations1, name1, description1, "hubspot_contacts")
	assert.Equal(t, http.StatusOK, w.Code)

	type KPIFilter struct {
		ObjectType       string `json:"objTy"`
		PropertyName     string `json:"prNa"`
		PropertyDataType string `json:"prDaTy"`
		Entity           string `json:"en"`
		Condition        string `json:"co"`
		Value            string `json:"va"`
		LogicalOp        string `json:"lOp"`
	}

	name2 := U.RandomString(8)
	description2 := U.RandomString(8)
	transformations2 := &postgres.Jsonb{json.RawMessage(`{"agFn": "sum", "agPr": "$hubspot_amount", "agPrTy": "categorical", "fil": [{"objTy": "", "prNa": "country", "prDaTy": "categorical", "en": "user", "co": "equals", "va": "india", "lOp": "AND"}], "daFie": "$hubspot_datefield1"}`)}
	w = sendCreateCustomMetric(a, project.ID, agent, transformations2, name2, description2, "hubspot_contacts")
	assert.Equal(t, http.StatusOK, w.Code)

	t.Run("test hubspot contacts with no filters and no group by - no gbt ", func(t *testing.T) {
		query1 := model.KPIQuery{
			Category:        "profile",
			DisplayCategory: "hubspot_contacts",
			PageUrl:         "",
			Metrics:         []string{name1},
			GroupBy:         []M.KPIGroupBy{},
			From:            1640975425 - 200,
			To:              1640975425 + 200,
		}

		kpiQueryGroup := model.KPIQueryGroup{
			Class:         "kpi",
			Queries:       []model.KPIQuery{query1},
			GlobalFilters: []model.KPIFilter{},
			GlobalGroupBy: []model.KPIGroupBy{},
		}
		result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup)
		log.WithField("result", result).Warn("kark1")
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, result[0].Headers, []string{name1})
		assert.Equal(t, len(result[0].Rows), 1)
		assert.Equal(t, result[0].Rows[0][0], float64(500))
	})

	t.Run("test hubspot contacts with no filters and no group by - with gbt", func(t *testing.T) {
		query2 := model.KPIQuery{
			Category:         "profile",
			DisplayCategory:  "hubspot_contacts",
			PageUrl:          "",
			Metrics:          []string{name1},
			GroupBy:          []M.KPIGroupBy{},
			From:             1640975425 - 200,
			To:               1640975425 + 200,
			GroupByTimestamp: "date",
		}

		kpiQueryGroup2 := model.KPIQueryGroup{
			Class:         "kpi",
			Queries:       []model.KPIQuery{query2},
			GlobalFilters: []model.KPIFilter{},
			GlobalGroupBy: []model.KPIGroupBy{},
		}
		result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup2)
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, result[0].Headers, []string{"datetime", name1})
		assert.Equal(t, len(result[0].Rows), 1)
		assert.Equal(t, result[0].Rows[0][0], "2021-12-31T00:00:00+00:00")
		assert.Equal(t, result[0].Rows[0][1], float64(500))

		log.WithField("result", result).Warn("kark1")
	})

	t.Run("test hubspot contacts with filters only - no gbt", func(t *testing.T) {
		query2 := model.KPIQuery{
			Category:         "profile",
			DisplayCategory:  "hubspot_contacts",
			PageUrl:          "",
			Metrics:          []string{name1},
			GroupBy:          []M.KPIGroupBy{},
			From:             1640975425 - 200,
			To:               1640975425 + 200,
			GroupByTimestamp: "",
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
		kpiQueryGroup2 := model.KPIQueryGroup{
			Class:         "kpi",
			Queries:       []model.KPIQuery{query2},
			GlobalFilters: []model.KPIFilter{filter},
			GlobalGroupBy: []model.KPIGroupBy{},
		}
		result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup2)
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, result[0].Headers, []string{name1})
		assert.Equal(t, len(result[0].Rows), 1)
		assert.Equal(t, result[0].Rows[0][0], float64(300))

		log.WithField("result", result).Warn("kark1")
	})

	t.Run("test hubspot contacts with filters only - no gbt", func(t *testing.T) {
		query2 := model.KPIQuery{
			Category:         "profile",
			DisplayCategory:  "hubspot_contacts",
			PageUrl:          "",
			Metrics:          []string{name2},
			GroupBy:          []M.KPIGroupBy{},
			From:             1640975425 - 200,
			To:               1640975425 + 200,
			GroupByTimestamp: "",
		}

		kpiQueryGroup2 := model.KPIQueryGroup{
			Class:         "kpi",
			Queries:       []model.KPIQuery{query2},
			GlobalFilters: []model.KPIFilter{},
			GlobalGroupBy: []model.KPIGroupBy{},
		}
		result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup2)
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, result[0].Headers, []string{name2})
		assert.Equal(t, len(result[0].Rows), 1)
		assert.Equal(t, result[0].Rows[0][0], float64(300))

		log.WithField("result", result).Warn("kark1")
	})

	t.Run("test hubspot contacts with filters present in custom metric  - no gbt", func(t *testing.T) {
		query2 := model.KPIQuery{
			Category:         "profile",
			DisplayCategory:  "hubspot_contacts",
			PageUrl:          "",
			Metrics:          []string{name2},
			GroupBy:          []M.KPIGroupBy{},
			From:             1640975425 - 200,
			To:               1640975425 + 200,
			GroupByTimestamp: "",
		}

		kpiQueryGroup2 := model.KPIQueryGroup{
			Class:         "kpi",
			Queries:       []model.KPIQuery{query2},
			GlobalFilters: []model.KPIFilter{},
			GlobalGroupBy: []model.KPIGroupBy{},
		}
		result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup2)
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, result[0].Headers, []string{name2})
		assert.Equal(t, len(result[0].Rows), 1)
		assert.Equal(t, result[0].Rows[0][0], float64(300))

		log.WithField("result", result).Warn("kark1")
	})

	t.Run("test hubspot contacts with filter and group by - no gbt", func(t *testing.T) {
		query1 := model.KPIQuery{
			Category:         "profile",
			DisplayCategory:  "hubspot_contacts",
			PageUrl:          "",
			Metrics:          []string{name2},
			GroupBy:          []M.KPIGroupBy{},
			From:             1640975425 - 200,
			To:               1640975425 + 200,
			GroupByTimestamp: "",
		}

		groupBy := model.KPIGroupBy{
			ObjectType:       "",
			PropertyName:     "country",
			PropertyDataType: "categorical",
			Entity:           "user",
			GroupByType:      "",
			Granularity:      "",
		}

		kpiQueryGroup2 := model.KPIQueryGroup{
			Class:         "kpi",
			Queries:       []model.KPIQuery{query1},
			GlobalFilters: []model.KPIFilter{},
			GlobalGroupBy: []model.KPIGroupBy{groupBy},
		}
		result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup2)
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, result[0].Headers, []string{groupBy.PropertyName, name2})
		assert.Equal(t, len(result[0].Rows), 1)
		assert.Equal(t, result[0].Rows[0][0], "india")
		assert.Equal(t, result[0].Rows[0][1], float64(300))

		log.WithField("result", result).Warn("kark1")
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

func sendKPIAnalyticsQueryReq(r *gin.Engine, projectId uint64, agent *M.Agent, kpiqueryGroup model.KPIQueryGroup) *httptest.ResponseRecorder {
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
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	log.Warn(jsonResponse)
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

func sendKPIFilterValuesReq(r *gin.Engine, projectId uint64, agent *M.Agent, filterValueRequest v1.KPIFilterValuesRequest) *httptest.ResponseRecorder {
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

func sendEventNamesTypeQueryReq(r *gin.Engine, projectId uint64, agent *M.Agent) *httptest.ResponseRecorder {
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
