package tests

import (
	"encoding/json"
	C "factors/config"
	H "factors/handler"
	"factors/model/model"
	M "factors/model/model"
	v1 "factors/handler/v1"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestPropertyMappingsForKPI(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	e := gin.Default()
	H.InitSDKServiceRoutes(e)
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

	// currentTime := time.Now().Unix()
	// date, _ := strconv.Atoi(time.Now().Format("20060102150405")[:8])
	// currentDate := int64(date)
	// currentDateString := time.Now().Format("2006-01-02") + "T00:00:00+00:00"
	currentTime := int64(1674084600)
	currentDate := int64(20230118)
	currentDateString := "2023-01-18T00:00:00+00:00"

	// Create Channels - adwords data
	adwordsDocuments := []M.AdwordsDocument{
		{ID: "1", Timestamp: currentDate, ProjectID: project.ID, CustomerAccountID: customerAccountID, TypeAlias: "campaign_performance_report",
			Value: &postgres.Jsonb{json.RawMessage(`{"cost": "11","clicks": "100","campaign_id":"1","impressions": "1000", "campaign_name": "india"}`)}},
		{ID: "2", Timestamp: currentDate, ProjectID: project.ID, CustomerAccountID: customerAccountID, TypeAlias: "campaign_performance_report",
			Value: &postgres.Jsonb{json.RawMessage(`{"cost": "12","clicks": "200","campaign_id":"2","impressions": "500", "campaign_name": "us"}`)}},
	}
	for _, adwordsDocument := range adwordsDocuments {
		status := store.GetStore().CreateAdwordsDocument(&adwordsDocument)
		assert.Equal(t, http.StatusCreated, status)
	}

	// Create events data
	timestamp := int64(currentTime - 82800)
	payload := fmt.Sprintf(`{"event_name": "%s", "timestamp": %d, "user_properties": {"$initial_source" : "%s", "$country":"india"}, "event_properties":{"$campaign_id":%d}}`, "s0", timestamp, "A", 1234)
	w := ServePostRequestWithHeaders(e, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response := DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "timestamp": %d, "user_properties": {"$initial_source" : "%s", "$country":"us"}, "event_properties":{"$campaign_id":%d}}`, "s0", timestamp, "A", 1234)
	w = ServePostRequestWithHeaders(e, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	// Create profiles data
	rCustomerUserId := U.RandomLowerAphaNumString(15)
	properties1 := postgres.Jsonb{RawMessage: json.RawMessage([]byte(fmt.Sprintf(`{"country": "india", "age": 30, "$hubspot_amount": 300, "$hubspot_datefield1": %d,  "paid": true}`, currentTime-100)))}
	properties2 := postgres.Jsonb{RawMessage: json.RawMessage([]byte(fmt.Sprintf(`{"country": "us", "age": 20, "$hubspot_amount": 200, "$hubspot_datefield1": %d, "paid": true}`, currentTime+100)))}

	joinTime := int64(currentTime - 100)

	createUserID1, newUserErrorCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: rCustomerUserId, Properties: properties1, JoinTimestamp: joinTime, Source: model.GetRequestSourcePointer(model.UserSourceHubspot)})
	assert.Equal(t, http.StatusCreated, newUserErrorCode)
	assert.NotEqual(t, "", createUserID1)

	nextUserJoinTime := joinTime + 100
	createUserID2, nextUserErrCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: properties2, JoinTimestamp: nextUserJoinTime, Source: model.GetRequestSourcePointer(model.UserSourceHubspot)})
	assert.Equal(t, http.StatusCreated, nextUserErrCode)
	assert.NotEqual(t, "", createUserID2)

	// Create custom metrics for profiles
	profileMetric1 := U.RandomString(8)
	description1 := U.RandomString(8)
	transformations1 := &postgres.Jsonb{json.RawMessage(`{"agFn": "sum", "agPr": "$hubspot_amount", "agPrTy": "numerical", "fil": [], "daFie": "$hubspot_datefield1"}`)}
	w = sendCreateCustomMetric(r, project.ID, agent, transformations1, profileMetric1, description1, "hubspot_contacts", 1)
	assert.Equal(t, http.StatusOK, w.Code)

	// Creating a property mapping across events and profiles on country property
	propertyMappingDisplayName := "Test property mapping 1"
	propertyMappingName := U.CreatePropertyNameFromDisplayName(propertyMappingDisplayName)
	propertiesRaw1 := model.Property{
		Category:        model.EventCategory,
		DisplayCategory: model.PageViewsDisplayCategory,
		ObjectType:      "",
		Name:            "$country",
		DataType:        "categorical",
		Entity:          "user",
	}
	propertiesRaw2 := model.Property{
		Category:        model.ProfileCategory,
		DisplayCategory: model.HubspotContactsDisplayCategory,
		ObjectType:      "",
		Name:            "country",
		DataType:        "categorical",
		Entity:          "user",
	}
	propertiesRaw3 := model.Property{
		Category:        model.ChannelCategory,
		DisplayCategory: model.GoogleAdsDisplayCategory,
		ObjectType:      "campaign",
		Name:            "campaign_name",
		DataType:        "categorical",
		Entity:          "",
	}
	propertiesRaw := []model.Property{propertiesRaw1, propertiesRaw2, propertiesRaw3}

	properties_byte, err := json.Marshal(propertiesRaw)
	if err != nil {
		log.Error(err)
	}
	properties := &postgres.Jsonb{RawMessage: properties_byte}
	w = sendCreatePropertyMapping(r, project.ID, agent, properties, propertyMappingDisplayName)
	assert.Equal(t, http.StatusOK, w.Code)

	// A profile query
	query1 := model.KPIQuery{
		Category:         model.ProfileCategory,
		DisplayCategory:  model.HubspotContactsDisplayCategory,
		PageUrl:          "",
		Metrics:          []string{profileMetric1},
		Filters:          []model.KPIFilter{},
		GroupBy:          []M.KPIGroupBy{},
		From:             currentTime - 200,
		To:               currentTime + 200,
		GroupByTimestamp: "date",
	}

	// An event query
	query2 := model.KPIQuery{
		Category:         model.EventCategory,
		DisplayCategory:  model.PageViewsDisplayCategory,
		PageUrl:          "s0",
		Metrics:          []string{"page_views"},
		Filters:          []model.KPIFilter{},
		GroupBy:          []M.KPIGroupBy{},
		From:             timestamp,
		To:               timestamp + 100,
		GroupByTimestamp: "date",
	}

	// A channel query
	query3 := model.KPIQuery{
		Category:         model.ChannelCategory,
		DisplayCategory:  model.GoogleAdsDisplayCategory,
		Metrics:          []string{"impressions"},
		Filters:          []model.KPIFilter{},
		GroupBy:          []model.KPIGroupBy{},
		From:             timestamp,
		To:               timestamp + 100,
		GroupByTimestamp: "date",
	}

	var query1Copy, query2Copy, query3Copy model.KPIQuery
	U.DeepCopy(&query1, &query1Copy)
	U.DeepCopy(&query2, &query2Copy)
	U.DeepCopy(&query3, &query3Copy)
	query1Copy.GroupByTimestamp = ""
	query2Copy.GroupByTimestamp = ""
	query3Copy.GroupByTimestamp = ""

	t.Run("Test property mapping as global filter across events and profiles", func(t *testing.T) {
		filter1 := model.KPIFilter{
			ObjectType:       "",
			PropertyName:     "country",
			PropertyDataType: "categorical",
			Entity:           "user",
			Condition:        "equals",
			Value:            "india",
			LogicalOp:        "AND",
		}

		filter2 := model.KPIFilter{
			ObjectType:       "",
			PropertyName:     "$country",
			PropertyDataType: "categorical",
			Entity:           "user",
			Condition:        "equals",
			Value:            "india",
			LogicalOp:        "AND",
		}

		// Create profiles query with country as local filter
		query1.Filters = []model.KPIFilter{filter1}
		query1Copy.Filters = []model.KPIFilter{filter1}

		// Create events query with $country as local filter
		query2.Filters = []model.KPIFilter{filter2}
		query2Copy.Filters = []model.KPIFilter{filter2}

		kpiQueryGroup := model.KPIQueryGroup{
			Class:         "kpi",
			Queries:       []model.KPIQuery{query1, query2, query1Copy, query2Copy},
			GlobalFilters: []model.KPIFilter{},
			GlobalGroupBy: []model.KPIGroupBy{},
		}

		result1, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, statusCode)

		// Create query which uses the property mapping as global filter
		query1.Filters = []model.KPIFilter{}
		query2.Filters = []model.KPIFilter{}
		query1Copy.Filters = []model.KPIFilter{}
		query2Copy.Filters = []model.KPIFilter{}

		filter := model.KPIFilter{
			ObjectType:        "",
			PropertyName:      propertyMappingName,
			IsPropertyMapping: true,
			PropertyDataType:  "categorical",
			Condition:         "equals",
			Value:             "india",
			LogicalOp:         "AND",
		}

		kpiQueryGroup = model.KPIQueryGroup{
			Class:         "kpi",
			Queries:       []model.KPIQuery{query1, query2, query1Copy, query2Copy},
			GlobalFilters: []model.KPIFilter{filter},
			GlobalGroupBy: []model.KPIGroupBy{},
		}
		result2, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, statusCode)
		// Since the property mapping which is used as a global filter works same as the two local filters, the result should be same
		assert.Equal(t, result1, result2)
	})

	t.Run("Test property mapping as global groupby across events and profiles", func(t *testing.T) {
		groupBy := model.KPIGroupBy{
			ObjectType:        "",
			PropertyName:      propertyMappingName,
			IsPropertyMapping: true,
			PropertyDataType:  "categorical",
		}

		kpiQueryGroup := model.KPIQueryGroup{
			Class:         "kpi",
			Queries:       []model.KPIQuery{query1, query2, query1Copy, query2Copy},
			GlobalFilters: []model.KPIFilter{},
			GlobalGroupBy: []model.KPIGroupBy{groupBy},
		}

		result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, []string{"datetime", propertyMappingName, profileMetric1, "page_views"}, result[0].Headers)
		assert.Equal(t, 2, len(result[0].Rows))
		assert.Equal(t, []interface{}{currentDateString, "india", float64(300), float64(1)}, result[0].Rows[0])
		assert.Equal(t, []interface{}{currentDateString, "us", float64(200), float64(1)}, result[0].Rows[1])
		
		assert.Equal(t, []string{propertyMappingName, profileMetric1, "page_views"}, result[1].Headers)
		assert.Equal(t, 2, len(result[1].Rows))
		assert.Equal(t, []interface{}{"india", float64(300), float64(1)}, result[1].Rows[0])
		assert.Equal(t, []interface{}{"us", float64(200), float64(1)}, result[1].Rows[1])
	})

	t.Run("Test property mapping as global groupby across channels and profiles", func(t *testing.T) {
		groupBy := model.KPIGroupBy{
			ObjectType:        "",
			PropertyName:      propertyMappingName,
			IsPropertyMapping: true,
			PropertyDataType:  "categorical",
		}

		kpiQueryGroup := model.KPIQueryGroup{
			Class:         "kpi",
			Queries:       []model.KPIQuery{query1, query3, query1Copy, query3Copy},
			GlobalFilters: []model.KPIFilter{},
			GlobalGroupBy: []model.KPIGroupBy{groupBy},
		}

		result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, []string{"datetime", propertyMappingName, profileMetric1, "google_ads_metrics_" + "impressions"}, result[0].Headers)
		assert.Equal(t, 2, len(result[0].Rows))
		assert.Equal(t, []interface{}{currentDateString, "india", float64(300), float64(1000)}, result[0].Rows[0])
		assert.Equal(t, []interface{}{currentDateString, "us", float64(200), float64(500)}, result[0].Rows[1])

		assert.Equal(t, []string{propertyMappingName, profileMetric1, "google_ads_metrics_" + "impressions"}, result[1].Headers)
		assert.Equal(t, 2, len(result[1].Rows))
		assert.Equal(t, []interface{}{"india", float64(300), float64(1000)}, result[1].Rows[0])
		assert.Equal(t, []interface{}{"us", float64(200), float64(500)}, result[1].Rows[1])
	})

	t.Run("Test for incorrect property mapping name and incorrect display category", func(t *testing.T) {
		filter := model.KPIFilter{
			ObjectType:        "",
			PropertyName:      "abcd",
			IsPropertyMapping: true,
			PropertyDataType:  "categorical",
			Condition:         "equals",
			Value:             "india",
			LogicalOp:         "AND",
		}

		kpiQueryGroup := model.KPIQueryGroup{
			Class:         "kpi",
			Queries:       []model.KPIQuery{query1, query2, query1Copy, query2Copy},
			GlobalFilters: []model.KPIFilter{filter},
			GlobalGroupBy: []model.KPIGroupBy{},
		}
		result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		// Since property mapping does not exist, it should return empty result and internal server error status code
		assert.Equal(t, http.StatusInternalServerError, statusCode)
		assert.Equal(t, []model.QueryResult{{}, {}}, result)

		// Adding incorrect DisplayCategory in query1
		// Which is not present in the property mapping
		query1.DisplayCategory = model.HubspotCompaniesDisplayCategory
		query1Copy.DisplayCategory = model.HubspotCompaniesDisplayCategory
		filter.PropertyName = propertyMappingName
		kpiQueryGroup.Queries = []model.KPIQuery{query1, query2, query1Copy, query2Copy}
		kpiQueryGroup.GlobalFilters = []model.KPIFilter{filter}

		result, statusCode = store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusBadRequest, statusCode)
		assert.Equal(t, []model.QueryResult{{}, {}}, result)

		// Undoing the changes to query1 and query1Copy
		query1.DisplayCategory = model.HubspotContactsDisplayCategory
		query1Copy.DisplayCategory = model.HubspotContactsDisplayCategory
	})

	t.Run("Test property mapping in derived metric", func(t *testing.T) {
		// An event query
		query2 := model.KPIQuery{
			Category:         model.EventCategory,
			DisplayCategory:  model.PageViewsDisplayCategory,
			PageUrl:          "s0",
			Metrics:          []string{"page_views"},
			Filters:          []model.KPIFilter{},
			GroupBy:          []M.KPIGroupBy{},
			Name:            "b",
			Timezone:        "Australia/Sydney",
		}
		// A channel query
		query3 := model.KPIQuery{
			Category:         model.ChannelCategory,
			DisplayCategory:  model.GoogleAdsDisplayCategory,
			Metrics:          []string{"impressions"},
			Filters:          []model.KPIFilter{},
			GroupBy:          []model.KPIGroupBy{},
			Name:            "a",
			Timezone:        "Australia/Sydney",
		}

		// Creating a derived metric using the above queries
		derivedMetric1 := U.RandomString(8)
		description1 := U.RandomString(8)
		transformationRaw, _ := json.Marshal(model.KPIQueryGroup{
			Class:   "kpi",
			Formula: "a/b",
			Queries: []model.KPIQuery{query2, query3},
		})
		transformations := &postgres.Jsonb{RawMessage: json.RawMessage(transformationRaw)}
		w = sendCreateCustomMetric(r, project.ID, agent, transformations, derivedMetric1, description1, "others", 2)
		assert.Equal(t, http.StatusOK, w.Code)

		// Creating a KPI query group using derived metric
		queryD := model.KPIQuery{
			Category:         "events",
			DisplayCategory:  "others",
			PageUrl:          "",
			Metrics:          []string{derivedMetric1},
			Filters:          []model.KPIFilter{},
			From:             timestamp,
			To:               timestamp + 100,
			GroupByTimestamp: "date",
			QueryType:        "derived",
		}
		queryDCopy := model.KPIQuery{}
		U.DeepCopy(&queryD, &queryDCopy)
		queryDCopy.GroupByTimestamp = ""

		filter := model.KPIFilter{
			PropertyName:      propertyMappingName,
			IsPropertyMapping: true,
			PropertyDataType:  "categorical",
			Condition:         "equals",
			Value:             "india",
			LogicalOp:         "AND",
		}
		groupBy := model.KPIGroupBy{
			PropertyName:      propertyMappingName,
			IsPropertyMapping: true,
			PropertyDataType:  "categorical",
		}

		kpiQueryGroup := model.KPIQueryGroup{
			Class:         "kpi",
			Queries:       []model.KPIQuery{queryD, queryDCopy},
			GlobalFilters: []model.KPIFilter{filter},
			GlobalGroupBy: []model.KPIGroupBy{groupBy},
		}
		
		result1, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, []string{"datetime", propertyMappingName, derivedMetric1}, result1[0].Headers)
		assert.Equal(t, 1, len(result1[0].Rows))
		assert.Equal(t, []interface{}{"2023-01-18T00:00:00+00:00", "india", float64(1000)}, result1[0].Rows[0])
		assert.Equal(t, []string{propertyMappingName, derivedMetric1}, result1[1].Headers)
		assert.Equal(t, 1, len(result1[1].Rows))
		assert.Equal(t, []interface{}{"india", float64(1000)}, result1[1].Rows[0])
	})

	t.Run("Test for kpi filter values fetch", func(t *testing.T) {
		filterValueRequest := v1.KPIFilterValuesRequest{
			PropertyName:      propertyMappingName,
			IsPropertyMapping: true,
		}

		w = sendKPIFilterValuesReq(r, project.ID, agent, filterValueRequest)
		assert.Equal(t, http.StatusOK, w.Code)
		log.WithField("response", w.Body).Debug("Filter Values")		
	})

}
