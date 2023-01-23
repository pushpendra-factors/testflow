package tests

import (
	"encoding/json"
	C "factors/config"
	H "factors/handler"
	"factors/model/model"
	M "factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func sendCreateFacebookDocumentReq(r *gin.Engine, project_id int64, customerAccountID string, valueJSON *postgres.Jsonb, id string, timestamp int64, type_alias string) *httptest.ResponseRecorder {
	payload := map[string]interface{}{
		"project_id":             project_id,
		"customer_ad_account_id": customerAccountID,
		"type_alias":             type_alias,
		"id":                     id,
		"value":                  valueJSON,
		"timestamp":              timestamp,
		"platform":               "facebook",
	}

	rb := U.NewRequestBuilder(http.MethodPost, "http://localhost:8089/data_service/facebook/documents/add").
		WithPostParams(payload)

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error sending facebook document add requests to data server.")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestExecuteKPIForFacebook(t *testing.T) {
	r := gin.Default()
	H.InitDataServiceRoutes(r)

	a := gin.Default()
	H.InitAppRoutes(a)
	model.SetSmartPropertiesReservedNames()

	//inserting sample data in facebook, also testing data service endpoint facebook/documents/add
	project, _, _ := SetupProjectWithAgentDAO()
	assert.NotNil(t, project)
	customerAccountID1 := U.RandomNumericString(10)
	customerAccountID2 := U.RandomNumericString(10)
	customerAccountID3 := U.RandomNumericString(10)
	_, errCode := store.GetStore().UpdateProjectSettings(project.ID, &M.ProjectSetting{
		IntFacebookAdAccount: customerAccountID1 + "," + customerAccountID2,
	})
	assert.Equal(t, http.StatusAccepted, errCode)

	campaignID1 := U.RandomNumericStringNonZeroStart(8)
	value := map[string]interface{}{"spend": "100", "clicks": "50", "campaign_id": campaignID1, "impressions": "1000", "campaign_name": "Campaign_1", "account_currency": "USD"}
	valueJSON, err := U.EncodeToPostgresJsonb(&value)
	assert.Nil(t, err)

	w := sendCreateFacebookDocumentReq(r, project.ID, customerAccountID1, valueJSON, campaignID1, 20210205, "campaign_insights")
	assert.Equal(t, http.StatusCreated, w.Code)

	campaignID2 := U.RandomNumericString(8)
	value = map[string]interface{}{"spend": "200", "clicks": "100", "campaign_id": campaignID2, "impressions": "2000", "campaign_name": "Campaign_2"}
	valueJSON, err = U.EncodeToPostgresJsonb(&value)
	assert.Nil(t, err)
	w = sendCreateFacebookDocumentReq(r, project.ID, customerAccountID1, valueJSON, campaignID2, 20210206, "campaign_insights")
	assert.Equal(t, http.StatusCreated, w.Code)

	adgroupID1_1 := U.RandomNumericString(8)
	value = map[string]interface{}{"spend": "30", "clicks": "30", "adset_id": adgroupID1_1, "adset_name": "Adgroup_1_1", "campaign_id": campaignID1, "impressions": "600", "campaign_name": "Campaign_1", "account_currency": "USD"}
	valueJSON, err = U.EncodeToPostgresJsonb(&value)
	assert.Nil(t, err)
	w = sendCreateFacebookDocumentReq(r, project.ID, customerAccountID2, valueJSON, adgroupID1_1, 20210205, "ad_set_insights")
	assert.Equal(t, http.StatusCreated, w.Code)

	adgroupID1_2 := U.RandomNumericString(8)
	value = map[string]interface{}{"spend": "70", "clicks": "20", "adset_id": adgroupID1_2, "adset_name": "Adgroup_1_2", "campaign_id": campaignID1, "impressions": "400", "campaign_name": "Campaign_1", "account_currency": "USD"}
	valueJSON, err = U.EncodeToPostgresJsonb(&value)
	assert.Nil(t, err)
	w = sendCreateFacebookDocumentReq(r, project.ID, customerAccountID2, valueJSON, adgroupID1_2, 20210205, "ad_set_insights")
	assert.Equal(t, http.StatusCreated, w.Code)

	adgroupID2_1 := U.RandomNumericString(8)
	value = map[string]interface{}{"spend": "120", "clicks": "25", "adset_id": adgroupID2_1, "adset_name": "Adgroup_2_1", "campaign_id": campaignID2, "impressions": "1500", "campaign_name": "Campaign_2", "account_currency": "USD"}
	valueJSON, err = U.EncodeToPostgresJsonb(&value)
	assert.Nil(t, err)
	w = sendCreateFacebookDocumentReq(r, project.ID, customerAccountID3, valueJSON, adgroupID2_1, 20210206, "ad_set_insights")
	assert.Equal(t, http.StatusCreated, w.Code)

	adgroupID2_2 := U.RandomNumericString(8)
	value = map[string]interface{}{"spend": "80", "clicks": "75", "adset_id": adgroupID1_2, "adset_name": "Adgroup_2_2", "campaign_id": campaignID2, "impressions": "500", "campaign_name": "Campaign_2", "account_currency": "USD"}
	valueJSON, err = U.EncodeToPostgresJsonb(&value)
	assert.Nil(t, err)
	w = sendCreateFacebookDocumentReq(r, project.ID, customerAccountID2, valueJSON, adgroupID2_2, 20210206, "ad_set_insights")
	assert.Equal(t, http.StatusCreated, w.Code)

	// // No filter no groupby, no gbt
	query := model.KPIQuery{
		Category:        "channels",
		DisplayCategory: "facebook_metrics",
		PageUrl:         "",
		Metrics:         []string{"impressions"},
		GroupBy:         []M.KPIGroupBy{},
		From:            1612314000,
		To:              1612746000,
	}
	kpiQueryGroup := model.KPIQueryGroup{
		Class:         "kpi",
		Queries:       []model.KPIQuery{query},
		GlobalFilters: []model.KPIFilter{},
		GlobalGroupBy: []model.KPIGroupBy{},
	}
	result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
		C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, float64(3000), result[0].Rows[0][0])

	// No filter no groupby, with gbt
	query = model.KPIQuery{
		Category:         "channels",
		DisplayCategory:  "facebook_metrics",
		PageUrl:          "",
		Metrics:          []string{"impressions"},
		GroupBy:          []M.KPIGroupBy{},
		GroupByTimestamp: "date",
		From:             1612314000,
		To:               1612746000,
	}
	kpiQueryGroup = model.KPIQueryGroup{
		Class:         "kpi",
		Queries:       []model.KPIQuery{query},
		GlobalFilters: []model.KPIFilter{},
		GlobalGroupBy: []model.KPIGroupBy{},
	}
	result, statusCode = store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
		C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, []string{"datetime", "facebook_metrics_impressions"}, result[0].Headers)
	assert.Equal(t, 6, len(result[0].Rows)) // date from 3rd Feb to 8th Feb
	assert.Equal(t, []interface{}{"2021-02-05T00:00:00+00:00", float64(1000)}, result[0].Rows[2])
	assert.Equal(t, []interface{}{"2021-02-06T00:00:00+00:00", float64(2000)}, result[0].Rows[3])

	// groupby with gbt
	query = model.KPIQuery{
		Category:         "channels",
		DisplayCategory:  "facebook_metrics",
		PageUrl:          "",
		Metrics:          []string{"impressions"},
		GroupBy:          []M.KPIGroupBy{},
		GroupByTimestamp: "date",
		From:             1612314000,
		To:               1612746000,
	}
	kpiQueryGroup = model.KPIQueryGroup{
		Class:         "kpi",
		Queries:       []model.KPIQuery{query},
		GlobalFilters: []model.KPIFilter{},
		GlobalGroupBy: []model.KPIGroupBy{
			{
				ObjectType:       "campaign",
				PropertyName:     "name",
				PropertyDataType: "categorical",
				Entity:           "",
			},
		},
	}
	result, statusCode = store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
		C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, []string{"datetime", "campaign_name", "facebook_metrics_impressions"}, result[0].Headers)
	assert.Equal(t, 12, len(result[0].Rows)) // date from 3rd Feb to 8th Feb and 2 campaigns
	assert.Equal(t, []interface{}{"2021-02-05T00:00:00+00:00", "Campaign_1", float64(1000)}, result[0].Rows[4])
	assert.Equal(t, []interface{}{"2021-02-06T00:00:00+00:00", "Campaign_2", float64(2000)}, result[0].Rows[7])

	// filters: campaignName contains '1' & adGroupName contains '1_1', groupBy: campaignID, adGroupName
	query = model.KPIQuery{
		Category:        "channels",
		DisplayCategory: "facebook_metrics",
		PageUrl:         "",
		Metrics:         []string{"impressions"},
		GroupBy:         []M.KPIGroupBy{},
		From:            1612314000,
		To:              1612746000,
	}
	kpiQueryGroup = model.KPIQueryGroup{
		Class:   "kpi",
		Queries: []model.KPIQuery{query},
		GlobalFilters: []model.KPIFilter{
			{
				ObjectType:       "campaign",
				PropertyName:     "name",
				PropertyDataType: "categorical",
				Entity:           "",
				Condition:        "contains",
				Value:            "1",
				LogicalOp:        "AND",
			},
		},
		GlobalGroupBy: []model.KPIGroupBy{
			{
				ObjectType:       "campaign",
				PropertyName:     "id",
				PropertyDataType: "categorical",
				Entity:           "",
			},
			{
				ObjectType:       "ad_group",
				PropertyName:     "name",
				PropertyDataType: "categorical",
				Entity:           "",
			},
		},
	}
	result, statusCode = store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
		C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, []string{"campaign_id", "ad_group_name", "facebook_metrics_impressions"}, result[0].Headers)
	assert.Equal(t, []interface{}{campaignID1, "Adgroup_1_1", float64(600)}, result[0].Rows[0])
	assert.Equal(t, []interface{}{campaignID1, "Adgroup_1_2", float64(400)}, result[0].Rows[1])

	// filters: campaignName equals '$none'
	query = model.KPIQuery{
		Category:        "channels",
		DisplayCategory: "facebook_metrics",
		PageUrl:         "",
		Metrics:         []string{"impressions"},
		GroupBy:         []M.KPIGroupBy{},
		From:            1612314000,
		To:              1612746000,
	}
	kpiQueryGroup = model.KPIQueryGroup{
		Class:   "kpi",
		Queries: []model.KPIQuery{query},
		GlobalFilters: []model.KPIFilter{
			{
				ObjectType:       "campaign",
				PropertyName:     "name",
				PropertyDataType: "categorical",
				Entity:           "",
				Condition:        "equals",
				Value:            "$none",
				LogicalOp:        "AND",
			},
		},
		GlobalGroupBy: []model.KPIGroupBy{},
	}
	result, statusCode = store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
		C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, []string{"facebook_metrics_impressions"}, result[0].Headers)
	assert.Equal(t, []interface{}{0}, result[0].Rows[0])

	// filters: campaignName not equals '$none'
	query = model.KPIQuery{
		Category:        "channels",
		DisplayCategory: "facebook_metrics",
		PageUrl:         "",
		Metrics:         []string{"impressions"},
		GroupBy:         []M.KPIGroupBy{},
		From:            1612314000,
		To:              1612746000,
	}
	kpiQueryGroup = model.KPIQueryGroup{
		Class:   "kpi",
		Queries: []model.KPIQuery{query},
		GlobalFilters: []model.KPIFilter{
			{
				ObjectType:       "campaign",
				PropertyName:     "name",
				PropertyDataType: "categorical",
				Entity:           "",
				Condition:        "notEqual",
				Value:            "$none",
				LogicalOp:        "AND",
			},
		},
		GlobalGroupBy: []model.KPIGroupBy{},
	}
	result, statusCode = store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
		C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, []string{"facebook_metrics_impressions"}, result[0].Headers)
	assert.Equal(t, []interface{}{float64(3000)}, result[0].Rows[0])

	// or filters
	query = model.KPIQuery{
		Category:        "channels",
		DisplayCategory: "facebook_metrics",
		PageUrl:         "",
		Metrics:         []string{"impressions"},
		GroupBy:         []M.KPIGroupBy{},
		From:            1612314000,
		To:              1612746000,
	}
	kpiQueryGroup = model.KPIQueryGroup{
		Class:   "kpi",
		Queries: []model.KPIQuery{query},
		GlobalFilters: []model.KPIFilter{
			{
				ObjectType:       "campaign",
				PropertyName:     "name",
				PropertyDataType: "categorical",
				Entity:           "",
				Condition:        "equals",
				Value:            "Campaign_1",
				LogicalOp:        "AND",
			},
			{
				ObjectType:       "campaign",
				PropertyName:     "name",
				PropertyDataType: "categorical",
				Entity:           "",
				Condition:        "equals",
				Value:            "Campaign_2",
				LogicalOp:        "OR",
			},
		},
		GlobalGroupBy: []model.KPIGroupBy{},
	}
	result, statusCode = store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
		C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, float64(3000), result[0].Rows[0][0])

	// or filters and "AND" filters combined
	query = model.KPIQuery{
		Category:        "channels",
		DisplayCategory: "facebook_metrics",
		PageUrl:         "",
		Metrics:         []string{"impressions"},
		GroupBy:         []M.KPIGroupBy{},
		From:            1612314000,
		To:              1612746000,
	}
	kpiQueryGroup = model.KPIQueryGroup{
		Class:   "kpi",
		Queries: []model.KPIQuery{query},
		GlobalFilters: []model.KPIFilter{
			{
				ObjectType:       "campaign",
				PropertyName:     "name",
				PropertyDataType: "categorical",
				Entity:           "",
				Condition:        "equals",
				Value:            "Campaign_1",
				LogicalOp:        "AND",
			},
			{
				ObjectType:       "campaign",
				PropertyName:     "name",
				PropertyDataType: "categorical",
				Entity:           "",
				Condition:        "equals",
				Value:            "Campaign_2",
				LogicalOp:        "OR",
			},
			{
				ObjectType:       "campaign",
				PropertyName:     "id",
				PropertyDataType: "categorical",
				Entity:           "",
				Condition:        "equals",
				Value:            "$none",
				LogicalOp:        "AND",
			},
		},
		GlobalGroupBy: []model.KPIGroupBy{},
	}
	result, statusCode = store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
		C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, 0, result[0].Rows[0][0])
}

func TestExecuteKPIForAdwords(t *testing.T) {
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
	campaignID1 := U.RandomNumericString(8)
	campaign1Value, _ := json.Marshal(map[string]interface{}{"cost": "100", "clicks": "50", "campaign_id": campaignID1, "impressions": "1000", "campaign_name": "Campaign_1"})
	campaignID2 := U.RandomNumericString(8)
	campaign2Value, _ := json.Marshal(map[string]interface{}{"cost": "200", "clicks": "100", "campaign_id": campaignID2, "impressions": "2000", "campaign_name": "Campaign_2"})
	adgroupID1_1 := U.RandomNumericString(8)
	adgroupID1_1Value, _ := json.Marshal(map[string]interface{}{"cost": "30", "clicks": "30", "ad_group_id": adgroupID1_1, "ad_group_name": "Adgroup_1_1", "campaign_id": campaignID1, "impressions": "600", "campaign_name": "Campaign_1"})
	adgroupID1_2 := U.RandomNumericString(8)
	adgroup1_2Value, _ := json.Marshal(map[string]interface{}{"cost": "70", "clicks": "20", "ad_group_id": adgroupID1_2, "ad_group_name": "Adgroup_1_2", "campaign_id": campaignID1, "impressions": "400", "campaign_name": "Campaign_1"})
	adgroupID2_1 := U.RandomNumericString(8)
	adgroup2_1Value, _ := json.Marshal(map[string]interface{}{"cost": "120", "clicks": "25", "ad_group_id": adgroupID2_1, "ad_group_name": "Adgroup_2_1", "campaign_id": campaignID2, "impressions": "1500", "campaign_name": "Campaign_2"})
	adgroupID2_2 := U.RandomNumericString(8)
	adgroup2_2Value, _ := json.Marshal(map[string]interface{}{"cost": "80", "clicks": "75", "ad_group_id": adgroupID2_2, "ad_group_name": "Adgroup_2_2", "campaign_id": campaignID2, "impressions": "500", "campaign_name": "Campaign_2"})
	adwordsDocuments := []model.AdwordsDocument{
		{ID: campaignID1, ProjectID: project.ID, CustomerAccountID: customerAccountID, TypeAlias: "campaign_performance_report", Timestamp: 20210205,
			Value: &postgres.Jsonb{RawMessage: campaign1Value}},

		{ID: campaignID2, ProjectID: project.ID, CustomerAccountID: customerAccountID, TypeAlias: "campaign_performance_report", Timestamp: 20210206,
			Value: &postgres.Jsonb{RawMessage: campaign2Value}},

		{ID: adgroupID1_1, ProjectID: project.ID, CustomerAccountID: customerAccountID, TypeAlias: "ad_group_performance_report", Timestamp: 20210205,
			Value: &postgres.Jsonb{RawMessage: adgroupID1_1Value}},

		{ID: adgroupID1_2, ProjectID: project.ID, CustomerAccountID: customerAccountID, TypeAlias: "ad_group_performance_report", Timestamp: 20210205,
			Value: &postgres.Jsonb{RawMessage: adgroup1_2Value}},

		{ID: adgroupID2_1, ProjectID: project.ID, CustomerAccountID: customerAccountID, TypeAlias: "ad_group_performance_report", Timestamp: 20210206,
			Value: &postgres.Jsonb{RawMessage: adgroup2_1Value}},

		{ID: adgroupID2_2, ProjectID: project.ID, CustomerAccountID: customerAccountID, TypeAlias: "ad_group_performance_report", Timestamp: 20210206,
			Value: &postgres.Jsonb{RawMessage: adgroup2_2Value}},
	}
	for _, adwordsDocument := range adwordsDocuments {
		status := store.GetStore().CreateAdwordsDocument(&adwordsDocument)
		assert.Equal(t, http.StatusCreated, status)
	}
	// No filter no groupby, no gbt
	query := model.KPIQuery{
		Category:        "channels",
		DisplayCategory: "google_ads_metrics",
		PageUrl:         "",
		Metrics:         []string{"impressions"},
		GroupBy:         []M.KPIGroupBy{},
		From:            1612314000,
		To:              1612746000,
	}
	kpiQueryGroup := model.KPIQueryGroup{
		Class:         "kpi",
		Queries:       []model.KPIQuery{query},
		GlobalFilters: []model.KPIFilter{},
		GlobalGroupBy: []model.KPIGroupBy{},
	}
	result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
		C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, float64(3000), result[0].Rows[0][0])

	// No filter no groupby, with gbt
	query = model.KPIQuery{
		Category:         "channels",
		DisplayCategory:  "google_ads_metrics",
		PageUrl:          "",
		Metrics:          []string{"impressions"},
		GroupBy:          []M.KPIGroupBy{},
		GroupByTimestamp: "date",
		From:             1612314000,
		To:               1612746000,
	}
	kpiQueryGroup = model.KPIQueryGroup{
		Class:         "kpi",
		Queries:       []model.KPIQuery{query},
		GlobalFilters: []model.KPIFilter{},
		GlobalGroupBy: []model.KPIGroupBy{},
	}
	result, statusCode = store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
		C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, []string{"datetime", "google_ads_metrics_impressions"}, result[0].Headers)
	assert.Equal(t, 6, len(result[0].Rows)) // date from 3rd Feb to 8th Feb
	assert.Equal(t, []interface{}{"2021-02-05T00:00:00+00:00", float64(1000)}, result[0].Rows[2])
	assert.Equal(t, []interface{}{"2021-02-06T00:00:00+00:00", float64(2000)}, result[0].Rows[3])

	// groupby with gbt
	query = model.KPIQuery{
		Category:         "channels",
		DisplayCategory:  "google_ads_metrics",
		PageUrl:          "",
		Metrics:          []string{"impressions"},
		GroupBy:          []M.KPIGroupBy{},
		GroupByTimestamp: "date",
		From:             1612314000,
		To:               1612746000,
	}
	kpiQueryGroup = model.KPIQueryGroup{
		Class:         "kpi",
		Queries:       []model.KPIQuery{query},
		GlobalFilters: []model.KPIFilter{},
		GlobalGroupBy: []model.KPIGroupBy{
			{
				ObjectType:       "campaign",
				PropertyName:     "name",
				PropertyDataType: "categorical",
				Entity:           "",
			},
		},
	}
	result, statusCode = store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
		C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, []string{"datetime", "campaign_name", "google_ads_metrics_impressions"}, result[0].Headers)
	assert.Equal(t, 12, len(result[0].Rows)) // date from 3rd Feb to 8th Feb and 2 campaigns
	assert.Equal(t, []interface{}{"2021-02-05T00:00:00+00:00", "Campaign_1", float64(1000)}, result[0].Rows[4])
	assert.Equal(t, []interface{}{"2021-02-06T00:00:00+00:00", "Campaign_2", float64(2000)}, result[0].Rows[7])

	// filters: campaignName contains '1' & adGroupName contains '1_1', groupBy: campaignID, adGroupName
	query = model.KPIQuery{
		Category:        "channels",
		DisplayCategory: "google_ads_metrics",
		PageUrl:         "",
		Metrics:         []string{"impressions"},
		GroupBy:         []M.KPIGroupBy{},
		From:            1612314000,
		To:              1612746000,
	}
	kpiQueryGroup = model.KPIQueryGroup{
		Class:   "kpi",
		Queries: []model.KPIQuery{query},
		GlobalFilters: []model.KPIFilter{
			{
				ObjectType:       "campaign",
				PropertyName:     "name",
				PropertyDataType: "categorical",
				Entity:           "",
				Condition:        "contains",
				Value:            "1",
				LogicalOp:        "AND",
			},
		},
		GlobalGroupBy: []model.KPIGroupBy{
			{
				ObjectType:       "campaign",
				PropertyName:     "id",
				PropertyDataType: "categorical",
				Entity:           "",
			},
			{
				ObjectType:       "ad_group",
				PropertyName:     "name",
				PropertyDataType: "categorical",
				Entity:           "",
			},
		},
	}
	result, statusCode = store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
		C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, []string{"campaign_id", "ad_group_name", "google_ads_metrics_impressions"}, result[0].Headers)
	assert.Equal(t, []interface{}{campaignID1, "Adgroup_1_1", float64(600)}, result[0].Rows[0])
	assert.Equal(t, []interface{}{campaignID1, "Adgroup_1_2", float64(400)}, result[0].Rows[1])

	// filters: campaignName equals '$none'
	query = model.KPIQuery{
		Category:        "channels",
		DisplayCategory: "google_ads_metrics",
		PageUrl:         "",
		Metrics:         []string{"impressions"},
		GroupBy:         []M.KPIGroupBy{},
		From:            1612314000,
		To:              1612746000,
	}
	kpiQueryGroup = model.KPIQueryGroup{
		Class:   "kpi",
		Queries: []model.KPIQuery{query},
		GlobalFilters: []model.KPIFilter{
			{
				ObjectType:       "campaign",
				PropertyName:     "name",
				PropertyDataType: "categorical",
				Entity:           "",
				Condition:        "equals",
				Value:            "$none",
				LogicalOp:        "AND",
			},
		},
		GlobalGroupBy: []model.KPIGroupBy{},
	}
	result, statusCode = store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
		C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, []string{"google_ads_metrics_impressions"}, result[0].Headers)
	assert.Equal(t, []interface{}{0}, result[0].Rows[0])

	// filters: campaignName not equals '$none'
	query = model.KPIQuery{
		Category:        "channels",
		DisplayCategory: "google_ads_metrics",
		PageUrl:         "",
		Metrics:         []string{"impressions"},
		GroupBy:         []M.KPIGroupBy{},
		From:            1612314000,
		To:              1612746000,
	}
	kpiQueryGroup = model.KPIQueryGroup{
		Class:   "kpi",
		Queries: []model.KPIQuery{query},
		GlobalFilters: []model.KPIFilter{
			{
				ObjectType:       "campaign",
				PropertyName:     "name",
				PropertyDataType: "categorical",
				Entity:           "",
				Condition:        "notEqual",
				Value:            "$none",
				LogicalOp:        "AND",
			},
		},
		GlobalGroupBy: []model.KPIGroupBy{},
	}
	result, statusCode = store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
		C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, []string{"google_ads_metrics_impressions"}, result[0].Headers)
	assert.Equal(t, []interface{}{float64(3000)}, result[0].Rows[0])

	// or filters
	query = model.KPIQuery{
		Category:        "channels",
		DisplayCategory: "google_ads_metrics",
		PageUrl:         "",
		Metrics:         []string{"impressions"},
		GroupBy:         []M.KPIGroupBy{},
		From:            1612314000,
		To:              1612746000,
	}
	kpiQueryGroup = model.KPIQueryGroup{
		Class:   "kpi",
		Queries: []model.KPIQuery{query},
		GlobalFilters: []model.KPIFilter{
			{
				ObjectType:       "campaign",
				PropertyName:     "name",
				PropertyDataType: "categorical",
				Entity:           "",
				Condition:        "equals",
				Value:            "Campaign_1",
				LogicalOp:        "AND",
			},
			{
				ObjectType:       "campaign",
				PropertyName:     "name",
				PropertyDataType: "categorical",
				Entity:           "",
				Condition:        "equals",
				Value:            "Campaign_2",
				LogicalOp:        "OR",
			},
		},
		GlobalGroupBy: []model.KPIGroupBy{},
	}
	result, statusCode = store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
		C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, float64(3000), result[0].Rows[0][0])

	// or filters and "AND" filters combined
	query = model.KPIQuery{
		Category:        "channels",
		DisplayCategory: "google_ads_metrics",
		PageUrl:         "",
		Metrics:         []string{"impressions"},
		GroupBy:         []M.KPIGroupBy{},
		From:            1612314000,
		To:              1612746000,
	}
	kpiQueryGroup = model.KPIQueryGroup{
		Class:   "kpi",
		Queries: []model.KPIQuery{query},
		GlobalFilters: []model.KPIFilter{
			{
				ObjectType:       "campaign",
				PropertyName:     "name",
				PropertyDataType: "categorical",
				Entity:           "",
				Condition:        "equals",
				Value:            "Campaign_1",
				LogicalOp:        "AND",
			},
			{
				ObjectType:       "campaign",
				PropertyName:     "name",
				PropertyDataType: "categorical",
				Entity:           "",
				Condition:        "equals",
				Value:            "Campaign_2",
				LogicalOp:        "OR",
			},
			{
				ObjectType:       "campaign",
				PropertyName:     "name",
				PropertyDataType: "categorical",
				Entity:           "",
				Condition:        "equals",
				Value:            "$none",
				LogicalOp:        "AND",
			},
		},
		GlobalGroupBy: []model.KPIGroupBy{},
	}
	result, statusCode = store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
		C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, 0, result[0].Rows[0][0])
}

func TestExecuteKPIForLinkedin(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)
	model.SetSmartPropertiesReservedNames()

	//inserting sample data in linkedin, also testing data service endpoint linkedin/documents/add
	project, _, _ := SetupProjectWithAgentDAO()
	assert.NotNil(t, project)
	customerAccountID := U.RandomNumericString(10)
	_, errCode := store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		IntLinkedinAdAccount: customerAccountID,
	})
	assert.Equal(t, http.StatusAccepted, errCode)

	campaignID1 := U.RandomNumericString(8)
	campaign1Value, _ := json.Marshal(map[string]interface{}{"costInLocalCurrency": "100", "clicks": "50", "campaign_group_id": campaignID1, "impressions": "1000", "campaign_group_name": "Campaign_1"})
	campaignID2 := U.RandomNumericString(8)
	campaign2Value, _ := json.Marshal(map[string]interface{}{"costInLocalCurrency": "200", "clicks": "100", "campaign_group_id": campaignID2, "impressions": "2000", "campaign_group_name": "Campaign_2"})
	adgroupID1_1 := U.RandomNumericString(8)
	adgroupID1_1Value, _ := json.Marshal(map[string]interface{}{"costInLocalCurrency": "30", "clicks": "30", "campaign_id": adgroupID1_1, "campaign_name": "Adgroup_1_1", "campaign_group_id": campaignID1, "impressions": "600", "campaign_group_name": "Campaign_1"})
	adgroupID1_2 := U.RandomNumericString(8)
	adgroup1_2Value, _ := json.Marshal(map[string]interface{}{"costInLocalCurrency": "70", "clicks": "20", "campaign_id": adgroupID1_2, "campaign_name": "Adgroup_1_2", "campaign_group_id": campaignID1, "impressions": "400", "campaign_group_name": "Campaign_1"})
	adgroupID2_1 := U.RandomNumericString(8)
	adgroup2_1Value, _ := json.Marshal(map[string]interface{}{"costInLocalCurrency": "120", "clicks": "25", "campaign_id": adgroupID2_1, "campaign_name": "Adgroup_2_1", "campaign_group_id": campaignID2, "impressions": "1500", "campaign_group_name": "Campaign_2"})
	adgroupID2_2 := U.RandomNumericString(8)
	adgroup2_2Value, _ := json.Marshal(map[string]interface{}{"costInLocalCurrency": "80", "clicks": "75", "campaign_id": adgroupID2_2, "campaign_name": "Adgroup_2_2", "campaign_group_id": campaignID2, "impressions": "500", "campaign_group_name": "Campaign_2"})
	linkedinDocuments := []model.LinkedinDocument{
		{ID: campaignID1, ProjectID: project.ID, CustomerAdAccountID: customerAccountID, TypeAlias: "campaign_group_insights", Timestamp: 20210205,
			Value: &postgres.Jsonb{RawMessage: campaign1Value}},

		{ID: campaignID2, ProjectID: project.ID, CustomerAdAccountID: customerAccountID, TypeAlias: "campaign_group_insights", Timestamp: 20210206,
			Value: &postgres.Jsonb{RawMessage: campaign2Value}},

		{ID: adgroupID1_1, ProjectID: project.ID, CustomerAdAccountID: customerAccountID, TypeAlias: "campaign_insights", Timestamp: 20210205,
			Value: &postgres.Jsonb{RawMessage: adgroupID1_1Value}},

		{ID: adgroupID1_2, ProjectID: project.ID, CustomerAdAccountID: customerAccountID, TypeAlias: "campaign_insights", Timestamp: 20210205,
			Value: &postgres.Jsonb{RawMessage: adgroup1_2Value}},

		{ID: adgroupID2_1, ProjectID: project.ID, CustomerAdAccountID: customerAccountID, TypeAlias: "campaign_insights", Timestamp: 20210206,
			Value: &postgres.Jsonb{RawMessage: adgroup2_1Value}},

		{ID: adgroupID2_2, ProjectID: project.ID, CustomerAdAccountID: customerAccountID, TypeAlias: "campaign_insights", Timestamp: 20210206,
			Value: &postgres.Jsonb{RawMessage: adgroup2_2Value}},
	}

	for _, linkedinDocument := range linkedinDocuments {
		status := store.GetStore().CreateLinkedinDocument(project.ID, &linkedinDocument)
		assert.Equal(t, http.StatusCreated, status)
	}

	// No filter no groupby, no gbt
	query := model.KPIQuery{
		Category:        "channels",
		DisplayCategory: "linkedin_metrics",
		PageUrl:         "",
		Metrics:         []string{"impressions"},
		GroupBy:         []M.KPIGroupBy{},
		From:            1612314000,
		To:              1612746000,
	}
	kpiQueryGroup := model.KPIQueryGroup{
		Class:         "kpi",
		Queries:       []model.KPIQuery{query},
		GlobalFilters: []model.KPIFilter{},
		GlobalGroupBy: []model.KPIGroupBy{},
	}
	result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
		C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, float64(3000), result[0].Rows[0][0])

	// No filter no groupby, with gbt
	query = model.KPIQuery{
		Category:         "channels",
		DisplayCategory:  "linkedin_metrics",
		PageUrl:          "",
		Metrics:          []string{"impressions"},
		GroupBy:          []M.KPIGroupBy{},
		GroupByTimestamp: "date",
		From:             1612314000,
		To:               1612746000,
	}
	kpiQueryGroup = model.KPIQueryGroup{
		Class:         "kpi",
		Queries:       []model.KPIQuery{query},
		GlobalFilters: []model.KPIFilter{},
		GlobalGroupBy: []model.KPIGroupBy{},
	}
	result, statusCode = store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
		C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, []string{"datetime", "linkedin_metrics_impressions"}, result[0].Headers)
	assert.Equal(t, 6, len(result[0].Rows)) // date from 3rd Feb to 8th Feb
	assert.Equal(t, []interface{}{"2021-02-05T00:00:00+00:00", float64(1000)}, result[0].Rows[2])
	assert.Equal(t, []interface{}{"2021-02-06T00:00:00+00:00", float64(2000)}, result[0].Rows[3])

	// groupby with gbt
	query = model.KPIQuery{
		Category:         "channels",
		DisplayCategory:  "linkedin_metrics",
		PageUrl:          "",
		Metrics:          []string{"impressions"},
		GroupBy:          []M.KPIGroupBy{},
		GroupByTimestamp: "date",
		From:             1612314000,
		To:               1612746000,
	}
	kpiQueryGroup = model.KPIQueryGroup{
		Class:         "kpi",
		Queries:       []model.KPIQuery{query},
		GlobalFilters: []model.KPIFilter{},
		GlobalGroupBy: []model.KPIGroupBy{
			{
				ObjectType:       "campaign",
				PropertyName:     "name",
				PropertyDataType: "categorical",
				Entity:           "",
			},
		},
	}
	result, statusCode = store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
		C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, []string{"datetime", "campaign_name", "linkedin_metrics_impressions"}, result[0].Headers)
	assert.Equal(t, 12, len(result[0].Rows)) // date from 3rd Feb to 8th Feb and 2 campaigns
	assert.Equal(t, []interface{}{"2021-02-05T00:00:00+00:00", "Campaign_1", float64(1000)}, result[0].Rows[4])
	assert.Equal(t, []interface{}{"2021-02-06T00:00:00+00:00", "Campaign_2", float64(2000)}, result[0].Rows[7])

	// filters: campaignName contains '1' & adGroupName contains '1_1', groupBy: campaignID, adGroupName
	query = model.KPIQuery{
		Category:        "channels",
		DisplayCategory: "linkedin_metrics",
		PageUrl:         "",
		Metrics:         []string{"impressions"},
		GroupBy:         []M.KPIGroupBy{},
		From:            1612314000,
		To:              1612746000,
	}
	kpiQueryGroup = model.KPIQueryGroup{
		Class:   "kpi",
		Queries: []model.KPIQuery{query},
		GlobalFilters: []model.KPIFilter{
			{
				ObjectType:       "campaign",
				PropertyName:     "name",
				PropertyDataType: "categorical",
				Entity:           "",
				Condition:        "contains",
				Value:            "1",
				LogicalOp:        "AND",
			},
		},
		GlobalGroupBy: []model.KPIGroupBy{
			{
				ObjectType:       "campaign",
				PropertyName:     "id",
				PropertyDataType: "categorical",
				Entity:           "",
			},
			{
				ObjectType:       "ad_group",
				PropertyName:     "name",
				PropertyDataType: "categorical",
				Entity:           "",
			},
		},
	}
	result, statusCode = store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
		C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, []string{"campaign_id", "ad_group_name", "linkedin_metrics_impressions"}, result[0].Headers)
	assert.Equal(t, []interface{}{campaignID1, "Adgroup_1_1", float64(600)}, result[0].Rows[0])
	assert.Equal(t, []interface{}{campaignID1, "Adgroup_1_2", float64(400)}, result[0].Rows[1])

	// filters: campaignName equals '$none'
	query = model.KPIQuery{
		Category:        "channels",
		DisplayCategory: "linkedin_metrics",
		PageUrl:         "",
		Metrics:         []string{"impressions"},
		GroupBy:         []M.KPIGroupBy{},
		From:            1612314000,
		To:              1612746000,
	}
	kpiQueryGroup = model.KPIQueryGroup{
		Class:   "kpi",
		Queries: []model.KPIQuery{query},
		GlobalFilters: []model.KPIFilter{
			{
				ObjectType:       "campaign",
				PropertyName:     "name",
				PropertyDataType: "categorical",
				Entity:           "",
				Condition:        "equals",
				Value:            "$none",
				LogicalOp:        "AND",
			},
		},
		GlobalGroupBy: []model.KPIGroupBy{},
	}
	result, statusCode = store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
		C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, []string{"linkedin_metrics_impressions"}, result[0].Headers)
	assert.Equal(t, []interface{}{0}, result[0].Rows[0])

	// filters: campaignName not equals '$none'
	query = model.KPIQuery{
		Category:        "channels",
		DisplayCategory: "linkedin_metrics",
		PageUrl:         "",
		Metrics:         []string{"impressions"},
		GroupBy:         []M.KPIGroupBy{},
		From:            1612314000,
		To:              1612746000,
	}
	kpiQueryGroup = model.KPIQueryGroup{
		Class:   "kpi",
		Queries: []model.KPIQuery{query},
		GlobalFilters: []model.KPIFilter{
			{
				ObjectType:       "campaign",
				PropertyName:     "name",
				PropertyDataType: "categorical",
				Entity:           "",
				Condition:        "notEqual",
				Value:            "$none",
				LogicalOp:        "AND",
			},
		},
		GlobalGroupBy: []model.KPIGroupBy{},
	}
	result, statusCode = store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
		C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, []string{"linkedin_metrics_impressions"}, result[0].Headers)
	assert.Equal(t, []interface{}{float64(3000)}, result[0].Rows[0])

	// or filters
	query = model.KPIQuery{
		Category:        "channels",
		DisplayCategory: "linkedin_metrics",
		PageUrl:         "",
		Metrics:         []string{"impressions"},
		GroupBy:         []M.KPIGroupBy{},
		From:            1612314000,
		To:              1612746000,
	}
	kpiQueryGroup = model.KPIQueryGroup{
		Class:   "kpi",
		Queries: []model.KPIQuery{query},
		GlobalFilters: []model.KPIFilter{
			{
				ObjectType:       "campaign",
				PropertyName:     "name",
				PropertyDataType: "categorical",
				Entity:           "",
				Condition:        "equals",
				Value:            "Campaign_1",
				LogicalOp:        "AND",
			},
			{
				ObjectType:       "campaign",
				PropertyName:     "name",
				PropertyDataType: "categorical",
				Entity:           "",
				Condition:        "equals",
				Value:            "Campaign_2",
				LogicalOp:        "OR",
			},
		},
		GlobalGroupBy: []model.KPIGroupBy{},
	}
	result, statusCode = store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
		C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, float64(3000), result[0].Rows[0][0])

	// or filters and "AND" filters combined
	query = model.KPIQuery{
		Category:        "channels",
		DisplayCategory: "linkedin_metrics",
		PageUrl:         "",
		Metrics:         []string{"impressions"},
		GroupBy:         []M.KPIGroupBy{},
		From:            1612314000,
		To:              1612746000,
	}
	kpiQueryGroup = model.KPIQueryGroup{
		Class:   "kpi",
		Queries: []model.KPIQuery{query},
		GlobalFilters: []model.KPIFilter{
			{
				ObjectType:       "campaign",
				PropertyName:     "name",
				PropertyDataType: "categorical",
				Entity:           "",
				Condition:        "equals",
				Value:            "Campaign_1",
				LogicalOp:        "AND",
			},
			{
				ObjectType:       "campaign",
				PropertyName:     "name",
				PropertyDataType: "categorical",
				Entity:           "",
				Condition:        "equals",
				Value:            "Campaign_2",
				LogicalOp:        "OR",
			},
			{
				ObjectType:       "campaign",
				PropertyName:     "id",
				PropertyDataType: "categorical",
				Entity:           "",
				Condition:        "equals",
				Value:            "$none",
				LogicalOp:        "AND",
			},
		},
		GlobalGroupBy: []model.KPIGroupBy{},
	}
	result, statusCode = store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
		C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, 0, result[0].Rows[0][0])
}

func TestExecuteKPIForLinkedinCompanyEngagements(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)
	model.SetSmartPropertiesReservedNames()

	//inserting sample data in linkedin, also testing data service endpoint linkedin/documents/add
	project, _, _ := SetupProjectWithAgentDAO()
	assert.NotNil(t, project)
	customerAccountID := U.RandomNumericString(10)
	_, errCode := store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		IntLinkedinAdAccount: customerAccountID,
	})
	assert.Equal(t, http.StatusAccepted, errCode)

	campaignID1 := U.RandomNumericString(8)
	campaign1Value, _ := json.Marshal(map[string]interface{}{"costInLocalCurrency": "100", "clicks": "50", "impressions": "1000", "vanityName": "Org1", "preferredCountry": "US", "localizedWebsite": "xyz.com", "localizedName": "Org_1", "companyHeadquarters": "US"})
	campaignID2 := U.RandomNumericString(8)
	campaign2Value, _ := json.Marshal(map[string]interface{}{"costInLocalCurrency": "200", "clicks": "100", "impressions": "2000", "vanityName": "Org2", "preferredCountry": "IN", "localizedWebsite": "abc.com", "localizedName": "Org_2", "companyHeadquarters": "IN"})
	linkedinDocuments := []model.LinkedinDocument{
		{ID: campaignID1, ProjectID: project.ID, CustomerAdAccountID: customerAccountID, TypeAlias: "member_company_insights", Timestamp: 20210205,
			Value: &postgres.Jsonb{RawMessage: campaign1Value}},

		{ID: campaignID2, ProjectID: project.ID, CustomerAdAccountID: customerAccountID, TypeAlias: "member_company_insights", Timestamp: 20210206,
			Value: &postgres.Jsonb{RawMessage: campaign2Value}},
	}

	for _, linkedinDocument := range linkedinDocuments {
		status := store.GetStore().CreateLinkedinDocument(project.ID, &linkedinDocument)
		assert.Equal(t, http.StatusCreated, status)
	}

	// No filter no groupby, no gbt
	query := model.KPIQuery{
		Category:        "channels",
		DisplayCategory: "linkedin_company_engagements",
		PageUrl:         "",
		Metrics:         []string{"impressions"},
		GroupBy:         []M.KPIGroupBy{},
		From:            1612314000,
		To:              1612746000,
	}
	kpiQueryGroup := model.KPIQueryGroup{
		Class:         "kpi",
		Queries:       []model.KPIQuery{query},
		GlobalFilters: []model.KPIFilter{},
		GlobalGroupBy: []model.KPIGroupBy{},
	}
	result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
		C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, float64(3000), result[0].Rows[0][0])

	// No filter no groupby, with gbt
	query = model.KPIQuery{
		Category:         "channels",
		DisplayCategory:  "linkedin_company_engagements",
		PageUrl:          "",
		Metrics:          []string{"impressions"},
		GroupBy:          []M.KPIGroupBy{},
		GroupByTimestamp: "date",
		From:             1612314000,
		To:               1612746000,
	}
	kpiQueryGroup = model.KPIQueryGroup{
		Class:         "kpi",
		Queries:       []model.KPIQuery{query},
		GlobalFilters: []model.KPIFilter{},
		GlobalGroupBy: []model.KPIGroupBy{},
	}
	result, statusCode = store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
		C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, []string{"datetime", "linkedin_company_engagements_impressions"}, result[0].Headers)
	assert.Equal(t, 6, len(result[0].Rows)) // date from 3rd Feb to 8th Feb
	assert.Equal(t, []interface{}{"2021-02-05T00:00:00+00:00", float64(1000)}, result[0].Rows[2])
	assert.Equal(t, []interface{}{"2021-02-06T00:00:00+00:00", float64(2000)}, result[0].Rows[3])

	// groupby with gbt
	query = model.KPIQuery{
		Category:         "channels",
		DisplayCategory:  "linkedin_company_engagements",
		PageUrl:          "",
		Metrics:          []string{"impressions"},
		GroupBy:          []M.KPIGroupBy{},
		GroupByTimestamp: "date",
		From:             1612314000,
		To:               1612746000,
	}
	kpiQueryGroup = model.KPIQueryGroup{
		Class:         "kpi",
		Queries:       []model.KPIQuery{query},
		GlobalFilters: []model.KPIFilter{},
		GlobalGroupBy: []model.KPIGroupBy{
			{
				ObjectType:       "company",
				PropertyName:     "company_vanity_name",
				PropertyDataType: "categorical",
				Entity:           "",
			},
			{
				ObjectType:       "company",
				PropertyName:     "company_localized_name",
				PropertyDataType: "categorical",
				Entity:           "",
			},
			{
				ObjectType:       "company",
				PropertyName:     "company_domain",
				PropertyDataType: "categorical",
				Entity:           "",
			},
			{
				ObjectType:       "company",
				PropertyName:     "company_preferred_country",
				PropertyDataType: "categorical",
				Entity:           "",
			},
			{
				ObjectType:       "company",
				PropertyName:     "company_headquarters",
				PropertyDataType: "categorical",
				Entity:           "",
			},
		},
	}
	result, statusCode = store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
		C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, []string{"datetime", "company_vanity_name", "company_localized_name", "company_domain", "company_preferred_country", "company_headquarters", "linkedin_company_engagements_impressions"}, result[0].Headers)
	assert.Equal(t, 12, len(result[0].Rows)) // date from 3rd Feb to 8th Feb and 2 campaigns
	assert.Equal(t, []interface{}{"2021-02-05T00:00:00+00:00", "Org1", "Org_1", "xyz.com", "US", "US", float64(1000)}, result[0].Rows[4])
	assert.Equal(t, []interface{}{"2021-02-06T00:00:00+00:00", "Org2", "Org_2", "abc.com", "IN", "IN", float64(2000)}, result[0].Rows[7])

	// filters: company vanity name contains '1', groupBy: company vanity name
	query = model.KPIQuery{
		Category:        "channels",
		DisplayCategory: "linkedin_company_engagements",
		PageUrl:         "",
		Metrics:         []string{"impressions"},
		GroupBy:         []M.KPIGroupBy{},
		From:            1612314000,
		To:              1612746000,
	}
	kpiQueryGroup = model.KPIQueryGroup{
		Class:   "kpi",
		Queries: []model.KPIQuery{query},
		GlobalFilters: []model.KPIFilter{
			{
				ObjectType:       "company",
				PropertyName:     "company_vanity_name",
				PropertyDataType: "categorical",
				Entity:           "",
				Condition:        "contains",
				Value:            "1",
				LogicalOp:        "AND",
			},
		},
		GlobalGroupBy: []model.KPIGroupBy{
			{
				ObjectType:       "company",
				PropertyName:     "company_vanity_name",
				PropertyDataType: "categorical",
				Entity:           "",
			},
			{
				ObjectType:       "company",
				PropertyName:     "company_domain",
				PropertyDataType: "categorical",
				Entity:           "",
			},
		},
	}
	result, statusCode = store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
		C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, []string{"company_vanity_name", "company_domain", "linkedin_company_engagements_impressions"}, result[0].Headers)
	assert.Equal(t, []interface{}{"Org1", "xyz.com", float64(1000)}, result[0].Rows[0])

	// filters: campaignName equals '$none'
	query = model.KPIQuery{
		Category:        "channels",
		DisplayCategory: "linkedin_company_engagements",
		PageUrl:         "",
		Metrics:         []string{"impressions"},
		GroupBy:         []M.KPIGroupBy{},
		From:            1612314000,
		To:              1612746000,
	}
	kpiQueryGroup = model.KPIQueryGroup{
		Class:   "kpi",
		Queries: []model.KPIQuery{query},
		GlobalFilters: []model.KPIFilter{
			{
				ObjectType:       "company",
				PropertyName:     "company_domain",
				PropertyDataType: "categorical",
				Entity:           "",
				Condition:        "equals",
				Value:            "$none",
				LogicalOp:        "AND",
			},
		},
		GlobalGroupBy: []model.KPIGroupBy{},
	}
	result, statusCode = store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
		C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, []string{"linkedin_company_engagements_impressions"}, result[0].Headers)
	assert.Equal(t, []interface{}{0}, result[0].Rows[0])

	// filters: campaignName not equals '$none'
	query = model.KPIQuery{
		Category:        "channels",
		DisplayCategory: "linkedin_company_engagements",
		PageUrl:         "",
		Metrics:         []string{"impressions"},
		GroupBy:         []M.KPIGroupBy{},
		From:            1612314000,
		To:              1612746000,
	}
	kpiQueryGroup = model.KPIQueryGroup{
		Class:   "kpi",
		Queries: []model.KPIQuery{query},
		GlobalFilters: []model.KPIFilter{
			{
				ObjectType:       "company",
				PropertyName:     "company_domain",
				PropertyDataType: "categorical",
				Entity:           "",
				Condition:        "notEqual",
				Value:            "$none",
				LogicalOp:        "AND",
			},
		},
		GlobalGroupBy: []model.KPIGroupBy{},
	}
	result, statusCode = store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
		C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, []string{"linkedin_company_engagements_impressions"}, result[0].Headers)
	assert.Equal(t, []interface{}{float64(3000)}, result[0].Rows[0])
}
func TestExecuteKPIForSearchConsole(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)
	model.SetSmartPropertiesReservedNames()

	// 	//inserting sample data in google_organic, also testing data service endpoint google_organic/documents/add
	project, _, _ := SetupProjectWithAgentDAO()
	assert.NotNil(t, project)
	urlPrefix := U.RandomNumericString(10)
	_, errCode := store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		IntGoogleOrganicURLPrefixes: &urlPrefix,
	})
	assert.Equal(t, http.StatusAccepted, errCode)

	id1 := U.RandomNumericString(8)
	value1, _ := json.Marshal(map[string]interface{}{"page": "factors.ai", "clicks": "50", "id": id1, "impressions": "1000"})
	id2 := U.RandomNumericString(8)
	value2, _ := json.Marshal(map[string]interface{}{"page": "factors.com", "clicks": "100", "id": id2, "impressions": "2000"})

	googleOrganicDocuments := []model.GoogleOrganicDocument{
		{ID: id1, ProjectID: project.ID, URLPrefix: urlPrefix, Timestamp: 20210205,
			Value: &postgres.Jsonb{RawMessage: value1}, Type: 2},

		{ID: id2, ProjectID: project.ID, URLPrefix: urlPrefix, Timestamp: 20210206,
			Value: &postgres.Jsonb{RawMessage: value2}, Type: 2},
	}

	for _, googleOrganicDocument := range googleOrganicDocuments {
		status := store.GetStore().CreateGoogleOrganicDocument(&googleOrganicDocument)
		assert.Equal(t, http.StatusCreated, status)
	}

	// No filter no groupby, no gbt
	query := model.KPIQuery{
		Category:        "channels",
		DisplayCategory: "google_organic_metrics",
		PageUrl:         "",
		Metrics:         []string{"impressions"},
		GroupBy:         []M.KPIGroupBy{},
		From:            1612314000,
		To:              1612746000,
	}
	kpiQueryGroup := model.KPIQueryGroup{
		Class:         "kpi",
		Queries:       []model.KPIQuery{query},
		GlobalFilters: []model.KPIFilter{},
		GlobalGroupBy: []model.KPIGroupBy{},
	}
	result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
		C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, float64(3000), result[0].Rows[0][0])

	// No filter no groupby, with gbt
	query = model.KPIQuery{
		Category:         "channels",
		DisplayCategory:  "google_organic_metrics",
		PageUrl:          "",
		Metrics:          []string{"impressions"},
		GroupBy:          []M.KPIGroupBy{},
		GroupByTimestamp: "date",
		From:             1612314000,
		To:               1612746000,
	}
	kpiQueryGroup = model.KPIQueryGroup{
		Class:         "kpi",
		Queries:       []model.KPIQuery{query},
		GlobalFilters: []model.KPIFilter{},
		GlobalGroupBy: []model.KPIGroupBy{},
	}
	result, statusCode = store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
		C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, []string{"datetime", "google_organic_metrics_impressions"}, result[0].Headers)
	assert.Equal(t, 6, len(result[0].Rows)) // date from 3rd Feb to 8th Feb
	assert.Equal(t, []interface{}{"2021-02-05T00:00:00+00:00", float64(1000)}, result[0].Rows[2])
	assert.Equal(t, []interface{}{"2021-02-06T00:00:00+00:00", float64(2000)}, result[0].Rows[3])

	// groupby with gbt
	query = model.KPIQuery{
		Category:         "channels",
		DisplayCategory:  "google_organic_metrics",
		PageUrl:          "",
		Metrics:          []string{"impressions"},
		GroupBy:          []M.KPIGroupBy{},
		GroupByTimestamp: "date",
		From:             1612314000,
		To:               1612746000,
	}
	kpiQueryGroup = model.KPIQueryGroup{
		Class:         "kpi",
		Queries:       []model.KPIQuery{query},
		GlobalFilters: []model.KPIFilter{},
		GlobalGroupBy: []model.KPIGroupBy{
			{
				ObjectType:       "organic_property",
				PropertyName:     "page",
				PropertyDataType: "categorical",
				Entity:           "",
			},
		},
	}
	result, statusCode = store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
		C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, []string{"datetime", "organic_property_page", "google_organic_metrics_impressions"}, result[0].Headers)
	assert.Equal(t, 12, len(result[0].Rows)) // date from 3rd Feb to 8th Feb and 2 pages
	assert.Equal(t, []interface{}{"2021-02-05T00:00:00+00:00", "factors.ai", float64(1000)}, result[0].Rows[4])
	assert.Equal(t, []interface{}{"2021-02-06T00:00:00+00:00", "factors.com", float64(2000)}, result[0].Rows[7])

	// filters: campaignName equals '$none'
	query = model.KPIQuery{
		Category:        "channels",
		DisplayCategory: "google_organic_metrics",
		PageUrl:         "",
		Metrics:         []string{"impressions"},
		GroupBy:         []M.KPIGroupBy{},
		From:            1612314000,
		To:              1612746000,
	}
	kpiQueryGroup = model.KPIQueryGroup{
		Class:   "kpi",
		Queries: []model.KPIQuery{query},
		GlobalFilters: []model.KPIFilter{
			{
				ObjectType:       "organic_property",
				PropertyName:     "page",
				PropertyDataType: "categorical",
				Entity:           "",
				Condition:        "equals",
				Value:            "$none",
				LogicalOp:        "AND",
			},
		},
		GlobalGroupBy: []model.KPIGroupBy{},
	}
	result, statusCode = store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
		C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, []string{"google_organic_metrics_impressions"}, result[0].Headers)
	assert.Equal(t, []interface{}{0}, result[0].Rows[0])

	// filters: campaignName not equals '$none'
	query = model.KPIQuery{
		Category:        "channels",
		DisplayCategory: "google_organic_metrics",
		PageUrl:         "",
		Metrics:         []string{"impressions"},
		GroupBy:         []M.KPIGroupBy{},
		From:            1612314000,
		To:              1612746000,
	}
	kpiQueryGroup = model.KPIQueryGroup{
		Class:   "kpi",
		Queries: []model.KPIQuery{query},
		GlobalFilters: []model.KPIFilter{
			{
				ObjectType:       "organic_property",
				PropertyName:     "page",
				PropertyDataType: "categorical",
				Entity:           "",
				Condition:        "notEqual",
				Value:            "$none",
				LogicalOp:        "AND",
			},
		},
		GlobalGroupBy: []model.KPIGroupBy{},
	}
	result, statusCode = store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
		C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, []string{"google_organic_metrics_impressions"}, result[0].Headers)
	assert.Equal(t, []interface{}{float64(3000)}, result[0].Rows[0])

	// or filters
	query = model.KPIQuery{
		Category:        "channels",
		DisplayCategory: "google_organic_metrics",
		PageUrl:         "",
		Metrics:         []string{"impressions"},
		GroupBy:         []M.KPIGroupBy{},
		From:            1612314000,
		To:              1612746000,
	}
	kpiQueryGroup = model.KPIQueryGroup{
		Class:   "kpi",
		Queries: []model.KPIQuery{query},
		GlobalFilters: []model.KPIFilter{
			{
				ObjectType:       "organic_property",
				PropertyName:     "page",
				PropertyDataType: "categorical",
				Entity:           "",
				Condition:        "equals",
				Value:            "factors.ai",
				LogicalOp:        "AND",
			},
			{
				ObjectType:       "organic_property",
				PropertyName:     "page",
				PropertyDataType: "categorical",
				Entity:           "",
				Condition:        "equals",
				Value:            "factors.com",
				LogicalOp:        "OR",
			},
		},
		GlobalGroupBy: []model.KPIGroupBy{},
	}
	result, statusCode = store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
		C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, float64(3000), result[0].Rows[0][0])

	// or filters and "AND" filters combined
	query = model.KPIQuery{
		Category:        "channels",
		DisplayCategory: "google_organic_metrics",
		PageUrl:         "",
		Metrics:         []string{"impressions"},
		GroupBy:         []M.KPIGroupBy{},
		From:            1612314000,
		To:              1612746000,
	}
	kpiQueryGroup = model.KPIQueryGroup{
		Class:   "kpi",
		Queries: []model.KPIQuery{query},
		GlobalFilters: []model.KPIFilter{
			{
				ObjectType:       "organic_property",
				PropertyName:     "page",
				PropertyDataType: "categorical",
				Entity:           "",
				Condition:        "equals",
				Value:            "factors.ai",
				LogicalOp:        "AND",
			},
			{
				ObjectType:       "organic_property",
				PropertyName:     "page",
				PropertyDataType: "categorical",
				Entity:           "",
				Condition:        "equals",
				Value:            "factors.com",
				LogicalOp:        "OR",
			},
			{
				ObjectType:       "organic_property",
				PropertyName:     "page",
				PropertyDataType: "categorical",
				Entity:           "",
				Condition:        "equals",
				Value:            "$none",
				LogicalOp:        "AND",
			},
		},
		GlobalGroupBy: []model.KPIGroupBy{},
	}
	// ** failing test case, requires fix on code
	// result, statusCode = store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
	// 	C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	// assert.Equal(t, http.StatusOK, statusCode)
	// assert.Equal(t, 0, result[0].Rows[0][0])
}

func TestWeeklyTrendForChannels(t *testing.T) {
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
		{ID: "1", Timestamp: 20210205, ProjectID: project.ID, CustomerAccountID: customerAccountID, TypeAlias: "campaign_performance_report",
			Value: &postgres.Jsonb{RawMessage: json.RawMessage(`{"cost": "11","clicks": "101","campaign_id":"1","impressions": "1001", "campaign_name": "test1"}`)}},
	}
	for _, adwordsDocument := range adwordsDocuments {
		store.GetStore().CreateAdwordsDocument(&adwordsDocument)
	}
	query := model.KPIQuery{
		Category:         "channels",
		DisplayCategory:  "google_ads_metrics",
		PageUrl:          "",
		Metrics:          []string{"impressions"},
		GroupBy:          []M.KPIGroupBy{},
		GroupByTimestamp: "week",
		From:             1612314000,
		To:               1612746000,
	}
	kpiQueryGroup := model.KPIQueryGroup{
		Class:         "kpi",
		Queries:       []model.KPIQuery{query},
		GlobalFilters: []model.KPIFilter{},
		GlobalGroupBy: []model.KPIGroupBy{},
	}
	result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
		C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, []string{"datetime", "google_ads_metrics_impressions"}, result[0].Headers)
	assert.Equal(t, 2, len(result[0].Rows)) // date from 3rd Feb to 8th Feb
	assert.Equal(t, []interface{}{"2021-01-31T00:00:00+00:00", float64(1001)}, result[0].Rows[0])
	assert.Equal(t, []interface{}{"2021-02-07T00:00:00+00:00", 0}, result[0].Rows[1])
}

func TestExecuteAllChannelKPI(t *testing.T) {
	model.SetSmartPropertiesReservedNames()
	project, agent, _ := SetupProjectWithAgentDAO()
	assert.NotNil(t, project)
	customerAccountIDAdwords := U.RandomNumericString(10)
	_, errCode := store.GetStore().UpdateProjectSettings(project.ID, &M.ProjectSetting{
		IntAdwordsCustomerAccountId: &customerAccountIDAdwords,
		IntAdwordsEnabledAgentUUID:  &agent.UUID,
	})
	assert.Equal(t, http.StatusAccepted, errCode)
	customerAccountIDLinkedin := U.RandomNumericString(10)
	_, errCode = store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		IntLinkedinAdAccount: customerAccountIDLinkedin,
	})
	assert.Equal(t, http.StatusAccepted, errCode)

	customerAccountIDFacebook := U.RandomNumericString(10)
	_, errCode = store.GetStore().UpdateProjectSettings(project.ID, &M.ProjectSetting{
		IntFacebookAdAccount: customerAccountIDFacebook,
	})
	assert.Equal(t, http.StatusAccepted, errCode)

	adwordsDocuments := []M.AdwordsDocument{
		{ID: "1", Timestamp: 20210205, ProjectID: project.ID, CustomerAccountID: customerAccountIDAdwords, TypeAlias: "campaign_performance_report",
			Value: &postgres.Jsonb{RawMessage: json.RawMessage(`{"cost": "10000000","clicks": "1000","campaign_id":"1","impressions": "10000", "campaign_name": "test1"}`)}},
		{ID: "2", Timestamp: 20210206, ProjectID: project.ID, CustomerAccountID: customerAccountIDAdwords, TypeAlias: "campaign_performance_report",
			Value: &postgres.Jsonb{RawMessage: json.RawMessage(`{"cost": "2000000","clicks": "200","campaign_id":"2","impressions": "15000", "campaign_name": "test2"}`)}},
	}
	for _, adwordsDocument := range adwordsDocuments {
		status := store.GetStore().CreateAdwordsDocument(&adwordsDocument)
		assert.Equal(t, http.StatusCreated, status)
	}

	linkedinDocuments := []model.LinkedinDocument{
		{ID: "1", Timestamp: 20210205, ProjectID: project.ID, CustomerAdAccountID: customerAccountIDLinkedin, TypeAlias: "campaign_group_insights",
			Value: &postgres.Jsonb{RawMessage: json.RawMessage(`{"costInLocalCurrency": "1","clicks": "10","campaign_id":"1","impressions": "100", "campaign_group_name": "test1"}`)}},
		{ID: "2", Timestamp: 20210206, ProjectID: project.ID, CustomerAdAccountID: customerAccountIDLinkedin, TypeAlias: "campaign_group_insights",
			Value: &postgres.Jsonb{RawMessage: json.RawMessage(`{"costInLocalCurrency": "2","clicks": "20","campaign_id":"2","impressions": "150", "campaign_group_name": "test2"}`)}},
	}
	for _, linkedinDocument := range linkedinDocuments {
		status := store.GetStore().CreateLinkedinDocument(project.ID, &linkedinDocument)
		assert.Equal(t, http.StatusCreated, status)
	}

	facebookDocuments := []model.FacebookDocument{
		{ID: "1", Timestamp: 20210205, ProjectID: project.ID, CustomerAdAccountID: customerAccountIDFacebook, TypeAlias: "campaign_insights",
			Value: &postgres.Jsonb{RawMessage: json.RawMessage(`{"spend": "10","clicks": "100","campaign_id":"1","impressions": "1000", "campaign_name": "test1"}`)}, Platform: "facebook"},
		{ID: "2", Timestamp: 20210206, ProjectID: project.ID, CustomerAdAccountID: customerAccountIDFacebook, TypeAlias: "campaign_insights",
			Value: &postgres.Jsonb{RawMessage: json.RawMessage(`{"spend": "20","clicks": "200","campaign_id":"2","impressions": "1500", "campaign_name": "test2"}`)}, Platform: "facebook"},
	}
	for _, facebookDocument := range facebookDocuments {
		status := store.GetStore().CreateFacebookDocument(project.ID, &facebookDocument)
		assert.Equal(t, http.StatusCreated, status)
	}

	// // No filter no groupby, no gbt
	query := model.KPIQuery{
		Category:        "channels",
		DisplayCategory: "all_channels_metrics",
		PageUrl:         "",
		Metrics:         []string{"impressions"},
		GroupBy:         []M.KPIGroupBy{},
		From:            1612314000,
		To:              1612746000,
	}
	kpiQueryGroup := model.KPIQueryGroup{
		Class:         "kpi",
		Queries:       []model.KPIQuery{query},
		GlobalFilters: []model.KPIFilter{},
		GlobalGroupBy: []model.KPIGroupBy{},
	}
	result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
		C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, float64(27750), result[0].Rows[0][0])

	// No filter no groupby, with gbt
	query = model.KPIQuery{
		Category:         "channels",
		DisplayCategory:  "all_channels_metrics",
		PageUrl:          "",
		Metrics:          []string{"impressions"},
		GroupBy:          []M.KPIGroupBy{},
		GroupByTimestamp: "date",
		From:             1612314000,
		To:               1612746000,
	}
	kpiQueryGroup = model.KPIQueryGroup{
		Class:         "kpi",
		Queries:       []model.KPIQuery{query},
		GlobalFilters: []model.KPIFilter{},
		GlobalGroupBy: []model.KPIGroupBy{},
	}
	result, statusCode = store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
		C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, []string{"datetime", "all_channels_metrics_impressions"}, result[0].Headers)
	assert.Equal(t, 6, len(result[0].Rows)) // date from 3rd Feb to 8th Feb
	assert.Equal(t, []interface{}{"2021-02-05T00:00:00+00:00", float64(11100)}, result[0].Rows[2])
	assert.Equal(t, []interface{}{"2021-02-06T00:00:00+00:00", float64(16650)}, result[0].Rows[3])

	// groupby with gbt
	query = model.KPIQuery{
		Category:         "channels",
		DisplayCategory:  "all_channels_metrics",
		PageUrl:          "",
		Metrics:          []string{"impressions"},
		GroupBy:          []M.KPIGroupBy{},
		GroupByTimestamp: "date",
		From:             1612314000,
		To:               1612746000,
	}
	kpiQueryGroup = model.KPIQueryGroup{
		Class:         "kpi",
		Queries:       []model.KPIQuery{query},
		GlobalFilters: []model.KPIFilter{},
		GlobalGroupBy: []model.KPIGroupBy{
			{
				ObjectType:       "campaign",
				PropertyName:     "name",
				PropertyDataType: "categorical",
				Entity:           "",
			},
		},
	}
	result, statusCode = store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
		C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, []string{"datetime", "campaign_name", "all_channels_metrics_impressions"}, result[0].Headers)
	assert.Equal(t, 12, len(result[0].Rows)) // date from 3rd Feb to 8th Feb and 2 campaigns
	assert.Equal(t, []interface{}{"2021-02-05T00:00:00+00:00", "test1", float64(11100)}, result[0].Rows[4])
	assert.Equal(t, []interface{}{"2021-02-06T00:00:00+00:00", "test2", float64(16650)}, result[0].Rows[7])

	// filters: campaign name contains '1' groupBy: channel name, campaign name
	query = model.KPIQuery{
		Category:        "channels",
		DisplayCategory: "all_channels_metrics",
		PageUrl:         "",
		Metrics:         []string{"impressions"},
		GroupBy:         []M.KPIGroupBy{},
		From:            1612314000,
		To:              1612746000,
	}
	kpiQueryGroup = model.KPIQueryGroup{
		Class:   "kpi",
		Queries: []model.KPIQuery{query},
		GlobalFilters: []model.KPIFilter{
			{
				ObjectType:       "channel",
				PropertyName:     "name",
				PropertyDataType: "categorical",
				Entity:           "",
				Condition:        "contains",
				Value:            "Ads",
				LogicalOp:        "AND",
			},
		},
		GlobalGroupBy: []model.KPIGroupBy{
			{
				ObjectType:       "channel",
				PropertyName:     "name",
				PropertyDataType: "categorical",
				Entity:           "",
			},
			{
				ObjectType:       "campaign",
				PropertyName:     "name",
				PropertyDataType: "categorical",
				Entity:           "",
			},
		},
	}
	result, statusCode = store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
		C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, 6, len(result[0].Rows)) // 3 channels and 2 campaigns
	assert.Equal(t, []string{"channel_name", "campaign_name", "all_channels_metrics_impressions"}, result[0].Headers)
	assert.Equal(t, []interface{}{"Facebook Ads", "test1", float64(1000)}, result[0].Rows[0])
	assert.Equal(t, []interface{}{"LinkedIn Ads", "test2", float64(150)}, result[0].Rows[5])
}

func TestChannelsV1ToKPIMigrationTransformation(t *testing.T) {
	r := gin.Default()
	H.InitDataServiceRoutes(r)
	model.SetSmartPropertiesReservedNames()

	project, _, _, statusCode := createProjectAndAddAdwordsDocument(t, r)
	if statusCode != http.StatusAccepted {
		assert.Equal(t, false, true)
		return
	}

	channelGroupQueryV1 := model.ChannelGroupQueryV1{
		Class: "channel_v1",
		Queries: []model.ChannelQueryV1{
			{
				Channel:       "google_ads",
				SelectMetrics: []string{"clicks", "impressions"},
				Filters: []M.ChannelFilterV1{
					{Object: "ad_group", Property: "name", Condition: "contains", Value: "1", LogicalOp: "AND"},
				},
				GroupBy: []M.ChannelGroupBy{
					{Object: "campaign", Property: "id"},
				},
				GroupByTimestamp: "date",
				Timezone:         "Asia/Kolkata",
				From:             1611964800,
				To:               1612310400,
			},
		},
	}
	kpiQueryGroup := model.TransformChannelsV1QueryToKPIQueryGroup(channelGroupQueryV1)

	log.WithField("kpiQueryGroup", kpiQueryGroup).Warn("testing kark1")
	result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, "",
		kpiQueryGroup, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	log.WithField("result", result).WithField("statusCode", statusCode).Warn("kark1")
	assert.NotNil(t, result[0].Headers)
	assert.NotNil(t, result[0].Rows)
}

func TestKPIChannelsWithSmartProperties(t *testing.T) {
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
		{ID: "2", Timestamp: 20220802, ProjectID: project.ID, CustomerAccountID: customerAccountID, TypeAlias: "campaign_performance_report",
			Value: &postgres.Jsonb{RawMessage: json.RawMessage(`{"cost": "12","clicks": "200","campaign_id":"2","impressions": "500", "campaign_name": "test2"}`)}},
	}
	for _, adwordsDocument := range adwordsDocuments {
		status := store.GetStore().CreateAdwordsDocument(&adwordsDocument)
		assert.Equal(t, http.StatusCreated, status)
	}
	sp1Obj := &postgres.Jsonb{RawMessage: json.RawMessage(`{"ad_group_id":"","ad_group_name":"","campaign_id":"1","campaign_name":"test1"}`)}
	sp1Name := "number1"
	sp1Properties := &postgres.Jsonb{RawMessage: json.RawMessage(fmt.Sprintf(`{"%s": "1"}`, sp1Name))}
	sp1 := model.SmartProperties{
		ProjectID:      project.ID,
		ObjectType:     1,
		ObjectID:       "1",
		ObjectProperty: sp1Obj,
		Properties:     sp1Properties,
		RulesRef:       &postgres.Jsonb{RawMessage: json.RawMessage(fmt.Sprintf(`{"%s": "%s"}`, uuid.New().String(), sp1Name))},
		Source:         "adwords",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	sp2Obj := &postgres.Jsonb{RawMessage: json.RawMessage(`{"ad_group_id":"","ad_group_name":"","campaign_id":"2","campaign_name":"test2"}`)}
	sp2Name := "number2"
	sp2Properties := &postgres.Jsonb{RawMessage: json.RawMessage(fmt.Sprintf(`{"%s": "2"}`, sp2Name))}
	sp2 := model.SmartProperties{
		ProjectID:      project.ID,
		ObjectType:     1,
		ObjectID:       "2",
		ObjectProperty: sp2Obj,
		Properties:     sp2Properties,
		RulesRef:       &postgres.Jsonb{RawMessage: json.RawMessage(fmt.Sprintf(`{"%s": "%s"}`, uuid.New().String(), sp2Name))},
		Source:         "adwords",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	errCode := store.GetStore().CreateSmartProperty(&sp1)
	assert.Equal(t, http.StatusCreated, errCode)
	errCode = store.GetStore().CreateSmartProperty(&sp2)
	assert.Equal(t, http.StatusCreated, errCode)

	query := model.KPIQuery{
		Category:        "channels",
		DisplayCategory: "google_ads_metrics",
		PageUrl:         "",
		Metrics:         []string{"impressions"},
		GroupBy:         []M.KPIGroupBy{},
		From:            1659312000,
		To:              1659657600,
	}
	kpiQueryGroup := model.KPIQueryGroup{
		Class:         "kpi",
		Queries:       []model.KPIQuery{query},
		GlobalFilters: []model.KPIFilter{},
		GlobalGroupBy: []model.KPIGroupBy{
			{
				ObjectType:       "campaign",
				PropertyName:     "campaign_number1",
				PropertyDataType: "categorical",
				Entity:           "",
			},
		},
	}
	result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
		C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	assert.Equal(t, result[0].Headers, []string{"campaign_number1", "google_ads_metrics_impressions"})
	assert.Equal(t, len(result[0].Rows), 2)
	assert.Equal(t, result[0].Rows, [][]interface{}{{"$none", float64(500)}, {"1", float64(1000)}})

	kpiQueryGroup = model.KPIQueryGroup{
		Class:   "kpi",
		Queries: []model.KPIQuery{query},
		GlobalFilters: []model.KPIFilter{
			{
				ObjectType:       "campaign",
				PropertyName:     "campaign_number1",
				PropertyDataType: "categorical",
				Entity:           "",
				Condition:        "equals",
				Value:            "$none",
				LogicalOp:        "AND",
			},
		},
		GlobalGroupBy: []model.KPIGroupBy{
			{
				ObjectType:       "campaign",
				PropertyName:     "campaign_number1",
				PropertyDataType: "categorical",
				Entity:           "",
			},
		},
	}
	result, statusCode = store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
		C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	assert.Equal(t, result[0].Headers, []string{"campaign_number1", "google_ads_metrics_impressions"})
	assert.Equal(t, len(result[0].Rows), 1)
	assert.Equal(t, result[0].Rows, [][]interface{}{{"$none", float64(500)}})

	kpiQueryGroup = model.KPIQueryGroup{
		Class:   "kpi",
		Queries: []model.KPIQuery{query},
		GlobalFilters: []model.KPIFilter{
			{
				ObjectType:       "campaign",
				PropertyName:     "campaign_number1",
				PropertyDataType: "categorical",
				Entity:           "",
				Condition:        "notEqual",
				Value:            "$none",
				LogicalOp:        "AND",
			},
		},
		GlobalGroupBy: []model.KPIGroupBy{
			{
				ObjectType:       "campaign",
				PropertyName:     "campaign_number1",
				PropertyDataType: "categorical",
				Entity:           "",
			},
		},
	}
	result, statusCode = store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
		C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	assert.Equal(t, result[0].Headers, []string{"campaign_number1", "google_ads_metrics_impressions"})
	assert.Equal(t, len(result[0].Rows), 1)
	assert.Equal(t, result[0].Rows, [][]interface{}{{"1", float64(1000)}})

	kpiQueryGroup = model.KPIQueryGroup{
		Class:   "kpi",
		Queries: []model.KPIQuery{query},
		GlobalFilters: []model.KPIFilter{
			{
				ObjectType:       "campaign",
				PropertyName:     "campaign_number1",
				PropertyDataType: "categorical",
				Entity:           "",
				Condition:        "equals",
				Value:            "1",
				LogicalOp:        "AND",
			},
		},
		GlobalGroupBy: []model.KPIGroupBy{
			{
				ObjectType:       "campaign",
				PropertyName:     "campaign_number1",
				PropertyDataType: "categorical",
				Entity:           "",
			},
		},
	}
	result, statusCode = store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup,
		C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	assert.Equal(t, result[0].Headers, []string{"campaign_number1", "google_ads_metrics_impressions"})
	assert.Equal(t, len(result[0].Rows), 1)
	assert.Equal(t, result[0].Rows, [][]interface{}{{"1", float64(1000)}})
}
