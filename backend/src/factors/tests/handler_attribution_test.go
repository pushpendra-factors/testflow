package tests

import (
	"encoding/json"

	H "factors/handler"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestAttributionDecode(t *testing.T) {

	// adding json attribution rule
	attributionQuery := postgres.Jsonb{RawMessage: json.RawMessage(`{"cl": "attribution", "meta": {"metrics_breakdown": true}, "query": {"ce": {"na": "$session", "pr": []}, "cm": ["Impressions", "Clicks", "Spend"], "to": 1596479399, "lbw": 1, "lfe": [], "from": 1595874600, "attribution_key": "Campaign", "attribution_methodology": "Last_Touch"}}`)}

	var attrQuery *model.AttributionQuery
	err := U.DecodePostgresJsonbToStructType(&attributionQuery, &attrQuery)
	fmt.Println(attrQuery)
	assert.Nil(t, err)
}

func TestAPIAttributionQueryHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	project.TimeZone = string(U.TimeZoneStringIST)
	store.GetStore().UpdateProject(project.ID, project)

	_, errCode := store.GetStore().CreateOrGetUserCreatedEventName(&model.EventName{ProjectId: project.ID, Name: "$session"})
	assert.Equal(t, http.StatusCreated, errCode)

	customerAccountID := U.RandomLowerAphaNumString(5)
	_, errCode = store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		IntAdwordsCustomerAccountId: &customerAccountID,
	})

	timezonestring := U.TimeZoneString(project.TimeZone)
	dashboardName := U.RandomString(5)
	dashboard, errCode := store.GetStore().CreateDashboard(project.ID, agent.UUID, &model.Dashboard{Name: dashboardName, Type: model.DashboardTypeProjectVisible})
	assert.NotNil(t, dashboard)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Equal(t, dashboardName, dashboard.Name)

	dashboardUnitQueriesMap := make(map[int64]map[string]interface{})
	var dashboardQueriesStr = map[string]string{
		model.QueryClassAttribution: `{"cl": "attribution", "meta": {"metrics_breakdown": true}, "query": {"ce": {"na": "$session", "pr": []}, "cm": ["Impressions", "Clicks", "Spend"], "to": 1596479399, "lbw": 1, "lfe": [], "from": 1595874600, "attribution_key": "Campaign", "attribution_methodology": "Last_Touch"}}`,
	}
	var dashboardQueryClassList []string
	var dashboardUnitsList []model.DashboardUnit
	for queryClass, queryString := range dashboardQueriesStr {
		dashboardQueryClassList = append(dashboardQueryClassList, queryClass)
		queryJSON := postgres.Jsonb{json.RawMessage(queryString)}
		baseQuery, err := model.DecodeQueryForClass(queryJSON, queryClass)
		assert.Nil(t, err)

		dashboardQuery, errCode, errMsg := store.GetStore().CreateQuery(project.ID, &model.Queries{
			ProjectID: project.ID,
			Type:      model.QueryTypeDashboardQuery,
			Query:     postgres.Jsonb{json.RawMessage(queryString)},
		})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.Empty(t, errMsg)
		assert.NotNil(t, dashboardQuery)

		dashboardUnit, errCode, _ := store.GetStore().CreateDashboardUnit(project.ID, agent.UUID, &model.DashboardUnit{
			DashboardId:  dashboard.ID,
			Presentation: model.PresentationCard,
			QueryId:      dashboardQuery.ID,
		})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboardUnit)
		dashboardUnitsList = append(dashboardUnitsList, *dashboardUnit)
		dashboardUnitQueriesMap[dashboardUnit.ID] = make(map[string]interface{})
		dashboardUnitQueriesMap[dashboardUnit.ID]["class"] = queryClass
		dashboardUnitQueriesMap[dashboardUnit.ID]["query"] = baseQuery
	}

	for rangeString, rangeFunction := range U.QueryDateRangePresets {

		from, to, errCode := rangeFunction(timezonestring)
		assert.Nil(t, errCode)
		for unitID, queryMap := range dashboardUnitQueriesMap {
			queryClass := queryMap["class"].(string)
			query := queryMap["query"].(model.BaseQuery)
			if queryClass == model.QueryClassAttribution {
				f, _ := model.GetEffectiveTimeRangeForDashboardUnitAttributionQuery(from, to)
				if f == 0 {
					continue
					assert.Nil(t, errCode)
				}
			}
			assertMsg := fmt.Sprintf("Failed for class:%s:range:%s", queryClass, rangeString)

			query.SetQueryDateRange(from, to)
			// Refresh is sent as false. Must return all presets range from cache.
			w := sendAnalyticsQueryReq(r, queryClass, project.ID, agent, dashboard.ID, unitID, rangeString, query, false, true)
			assert.NotNil(t, w)
			assert.Equal(t, http.StatusOK, w.Code, assertMsg)

			var result map[string]interface{}
			json.Unmarshal([]byte(w.Body.String()), &result)
			//metadata should not be nil
			assert.NotNil(t, result["cache_meta"])

			// Refresh is sent as true. Must return all recomputed presets.
			w = sendAnalyticsQueryReq(r, queryClass, project.ID, agent, dashboard.ID, unitID, rangeString, query, false, true)
			assert.NotNil(t, w)
			assert.Equal(t, http.StatusOK, w.Code, assertMsg)
			assert.NotNil(t, result["cache_meta"])
			result = nil
			json.Unmarshal([]byte(w.Body.String()), &result)
			log.Println(result)
			assert.NotNil(t, result["cache_meta"])

			// for Query

			// Refresh is sent as true. Must return all recomputed presets.
			w = sendAnalyticsQueryReq(r, queryClass, project.ID, agent, dashboard.ID, unitID, rangeString, query, false, false)
			assert.NotNil(t, w)
			assert.Equal(t, http.StatusOK, w.Code, assertMsg)
			assert.NotNil(t, result["cache_meta"])
			result = nil
			json.Unmarshal([]byte(w.Body.String()), &result)
			log.Println(result)
			assert.Equal(t, http.StatusOK, w.Code)
		}
	}

}
