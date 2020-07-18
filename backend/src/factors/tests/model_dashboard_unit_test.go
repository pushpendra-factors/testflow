package tests

import (
	"encoding/json"
	C "factors/config"
	H "factors/handler"
	"factors/handler/helpers"
	M "factors/model"
	U "factors/util"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"

	"github.com/stretchr/testify/assert"
)

func TestCreateDashboardUnit(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	agent2, errCode := SetupAgentReturnDAO(getRandomEmail(), "+13425356")
	assert.Equal(t, http.StatusCreated, errCode)
	_, errCode = M.CreateProjectAgentMappingWithDependencies(&M.ProjectAgentMapping{
		ProjectID: project.ID, AgentUUID: agent2.UUID})
	assert.Equal(t, http.StatusCreated, errCode)

	rName := U.RandomString(5)
	dashboard, errCode := M.CreateDashboard(project.ID, agent.UUID, &M.Dashboard{Name: rName, Type: M.DashboardTypeProjectVisible})
	assert.NotNil(t, dashboard)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Equal(t, rName, dashboard.Name)

	t.Run("CreateDashboardUnit", func(t *testing.T) {
		rName := U.RandomString(5)
		dashboardUnit, errCode, errMsg := M.CreateDashboardUnit(project.ID, agent.UUID, &M.DashboardUnit{DashboardId: dashboard.ID,
			Title: rName, Presentation: M.PresentationLine, Query: postgres.Jsonb{json.RawMessage(`{}`)}})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboardUnit)
		assert.Empty(t, errMsg)

		rName1 := U.RandomString(5)
		dashboardUnit1, errCode, errMsg := M.CreateDashboardUnit(project.ID, agent.UUID, &M.DashboardUnit{DashboardId: dashboard.ID,
			Title: rName1, Presentation: M.PresentationBar, Query: postgres.Jsonb{json.RawMessage(`{}`)}})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboardUnit1)
		assert.Empty(t, errMsg)

		rName2 := U.RandomString(5)
		dashboardUnit2, errCode, errMsg := M.CreateDashboardUnit(project.ID, agent.UUID, &M.DashboardUnit{DashboardId: dashboard.ID,
			Title: rName2, Presentation: M.PresentationCard, Query: postgres.Jsonb{json.RawMessage(`{}`)}})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboardUnit2)
		assert.Empty(t, errMsg)

		// should be given a positions on dashboard.
		gDashboard, errCode := M.GetDashboard(project.ID, agent.UUID, dashboard.ID)
		assert.Equal(t, http.StatusFound, errCode)
		var currentPosition map[string]map[uint64]int
		err := json.Unmarshal((*gDashboard.UnitsPosition).RawMessage, &currentPosition)
		assert.Nil(t, err)
		// inc position on unit type chart.
		assert.Contains(t, currentPosition[M.GetUnitType(M.PresentationLine)], dashboardUnit.ID)
		assert.Contains(t, currentPosition[M.GetUnitType(M.PresentationBar)], dashboardUnit1.ID)
		assert.Equal(t, currentPosition[M.GetUnitType(M.PresentationLine)][dashboardUnit.ID], 0)
		assert.Equal(t, currentPosition[M.GetUnitType(M.PresentationLine)][dashboardUnit1.ID], 1)
		// inc position on unit type card.
		assert.Contains(t, currentPosition[M.GetUnitType(M.PresentationCard)], dashboardUnit2.ID)
		assert.Equal(t, currentPosition[M.GetUnitType(M.PresentationCard)][dashboardUnit2.ID], 0)

	})

	t.Run("CreateDashboardUnit:Invalid", func(t *testing.T) {
		// invalid title.
		dashboardUnit, errCode, _ := M.CreateDashboardUnit(project.ID, agent.UUID, &M.DashboardUnit{DashboardId: dashboard.ID,
			Title: "", Presentation: M.PresentationLine, Query: postgres.Jsonb{json.RawMessage(`{}`)}})
		assert.Equal(t, http.StatusBadRequest, errCode)
		assert.Nil(t, dashboardUnit)

		// invalid presentation.
		rName := U.RandomString(5)
		dashboardUnit, errCode, _ = M.CreateDashboardUnit(project.ID, agent.UUID, &M.DashboardUnit{DashboardId: dashboard.ID,
			Title: rName, Presentation: "", Query: postgres.Jsonb{json.RawMessage(`{}`)}})
		assert.Equal(t, http.StatusBadRequest, errCode)
		assert.Nil(t, dashboardUnit)

		// invalid dashboard.
		rName = U.RandomString(5)
		dashboardUnit, errCode, _ = M.CreateDashboardUnit(project.ID, agent.UUID, &M.DashboardUnit{DashboardId: 0,
			Title: rName, Presentation: M.PresentationLine, Query: postgres.Jsonb{json.RawMessage(`{}`)}})
		assert.Equal(t, http.StatusBadRequest, errCode)
		assert.Nil(t, dashboardUnit)

		// invalid project.
		rName = U.RandomString(5)
		dashboardUnit, errCode, _ = M.CreateDashboardUnit(0, agent.UUID, &M.DashboardUnit{DashboardId: dashboard.ID,
			Title: rName, Presentation: M.PresentationLine, Query: postgres.Jsonb{json.RawMessage(`{}`)}})
		assert.Equal(t, http.StatusBadRequest, errCode)
		assert.Nil(t, dashboardUnit)

		// invalid agent.
		rName = U.RandomString(5)
		dashboardUnit, errCode, _ = M.CreateDashboardUnit(project.ID, "", &M.DashboardUnit{DashboardId: dashboard.ID,
			Title: rName, Presentation: M.PresentationLine, Query: postgres.Jsonb{json.RawMessage(`{}`)}})
		assert.Equal(t, http.StatusBadRequest, errCode)
		assert.Nil(t, dashboardUnit)
	})

	t.Run("CreateDashboardUnit:WithoutAccessToDashboard", func(t *testing.T) {
		rName := U.RandomString(5)
		dashboard, errCode := M.CreateDashboard(project.ID, agent.UUID, &M.Dashboard{Name: rName, Type: M.DashboardTypePrivate})
		assert.NotNil(t, dashboard)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.Equal(t, rName, dashboard.Name)

		rTitle := U.RandomString(5)
		dashboardUnit, errCode, _ := M.CreateDashboardUnit(project.ID, agent2.UUID, &M.DashboardUnit{DashboardId: dashboard.ID,
			Title: rTitle, Presentation: M.PresentationLine, Query: postgres.Jsonb{json.RawMessage(`{}`)}})
		assert.Equal(t, http.StatusForbidden, errCode)
		assert.Nil(t, dashboardUnit)
	})
}

func TestGetDashboardUnits(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	agent2, errCode := SetupAgentReturnDAO(getRandomEmail(), "+13425356")
	assert.Equal(t, http.StatusCreated, errCode)
	_, errCode = M.CreateProjectAgentMappingWithDependencies(&M.ProjectAgentMapping{
		ProjectID: project.ID, AgentUUID: agent2.UUID})
	assert.Equal(t, http.StatusCreated, errCode)

	rName := U.RandomString(5)
	dashboard, errCode := M.CreateDashboard(project.ID, agent.UUID, &M.Dashboard{Name: rName, Type: M.DashboardTypeProjectVisible})
	assert.NotNil(t, dashboard)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Equal(t, rName, dashboard.Name)

	t.Run("GetDashboardUnits:NotAvailable", func(t *testing.T) {
		units, errCode := M.GetDashboardUnits(project.ID, agent.UUID, dashboard.ID)
		assert.Equal(t, http.StatusFound, errCode)
		assert.Len(t, units, 0)
	})

	t.Run("GetDashboardUnits:Available", func(t *testing.T) {
		rName := U.RandomString(5)
		dashboardUnit, errCode, errMsg := M.CreateDashboardUnit(project.ID, agent.UUID, &M.DashboardUnit{DashboardId: dashboard.ID,
			Title: rName, Presentation: M.PresentationLine, Query: postgres.Jsonb{json.RawMessage(`{}`)}})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboardUnit)
		assert.Empty(t, errMsg)

		units, errCode := M.GetDashboardUnits(project.ID, agent.UUID, dashboard.ID)
		assert.Equal(t, http.StatusFound, errCode)
		assert.Len(t, units, 1)
		assert.Equal(t, rName, units[0].Title)
		assert.Equal(t, M.PresentationLine, units[0].Presentation)
	})

	t.Run("GetDashboardUnits:Invalid", func(t *testing.T) {
		rName := U.RandomString(5)
		dashboardUnit, errCode, errMsg := M.CreateDashboardUnit(project.ID, agent.UUID, &M.DashboardUnit{DashboardId: dashboard.ID,
			Title: rName, Presentation: M.PresentationLine, Query: postgres.Jsonb{json.RawMessage(`{}`)}})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboardUnit)
		assert.Empty(t, errMsg)

		// invalid project
		units, errCode := M.GetDashboardUnits(0, agent.UUID, dashboard.ID)
		assert.Equal(t, http.StatusBadRequest, errCode)
		assert.Nil(t, units)

		// invalid agent
		units, errCode = M.GetDashboardUnits(project.ID, "", dashboard.ID)
		assert.Equal(t, http.StatusBadRequest, errCode)
		assert.Nil(t, units)

		// invalid dashboard
		units, errCode = M.GetDashboardUnits(project.ID, agent.UUID, 0)
		assert.Equal(t, http.StatusBadRequest, errCode)
		assert.Nil(t, units)
	})

	t.Run("GetDashboardUnits:WithoutAccessToDashboard", func(t *testing.T) {
		rName := U.RandomString(5)
		dashboard, errCode := M.CreateDashboard(project.ID, agent.UUID, &M.Dashboard{Name: rName, Type: M.DashboardTypePrivate})
		assert.NotNil(t, dashboard)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.Equal(t, rName, dashboard.Name)

		rTitle := U.RandomString(5)
		dashboardUnit, errCode, _ := M.CreateDashboardUnit(project.ID, agent.UUID, &M.DashboardUnit{DashboardId: dashboard.ID,
			Title: rTitle, Presentation: M.PresentationLine, Query: postgres.Jsonb{json.RawMessage(`{}`)}})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboardUnit)

		units, errCode := M.GetDashboardUnits(project.ID, agent2.UUID, dashboard.ID)
		assert.Equal(t, http.StatusForbidden, errCode)
		assert.Nil(t, units)
	})
}

func TestDeleteDashboardUnit(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	agent2, errCode := SetupAgentReturnDAO(getRandomEmail(), "+13425356")
	assert.Equal(t, http.StatusCreated, errCode)
	_, errCode = M.CreateProjectAgentMappingWithDependencies(&M.ProjectAgentMapping{
		ProjectID: project.ID, AgentUUID: agent2.UUID})
	assert.Equal(t, http.StatusCreated, errCode)

	rName := U.RandomString(5)
	dashboard, errCode := M.CreateDashboard(project.ID, agent.UUID, &M.Dashboard{Name: rName, Type: M.DashboardTypeProjectVisible})
	assert.NotNil(t, dashboard)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Equal(t, rName, dashboard.Name)

	t.Run("DeleteDashboardUnit", func(t *testing.T) {
		rName := U.RandomString(5)
		unit, errCode, _ := M.CreateDashboardUnit(project.ID, agent.UUID, &M.DashboardUnit{DashboardId: dashboard.ID,
			Title: rName, Presentation: M.PresentationLine, Query: postgres.Jsonb{json.RawMessage(`{}`)}})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, unit)

		unit1, errCode, _ := M.CreateDashboardUnit(project.ID, agent.UUID, &M.DashboardUnit{DashboardId: dashboard.ID,
			Title: rName, Presentation: M.PresentationBar, Query: postgres.Jsonb{json.RawMessage(`{}`)}})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, unit1)

		unit2, errCode, _ := M.CreateDashboardUnit(project.ID, agent.UUID, &M.DashboardUnit{DashboardId: dashboard.ID,
			Title: rName, Presentation: M.PresentationCard, Query: postgres.Jsonb{json.RawMessage(`{}`)}})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, unit2)

		errCode = M.DeleteDashboardUnit(project.ID, agent.UUID, dashboard.ID, unit.ID)
		assert.Equal(t, http.StatusAccepted, errCode)

		// should remove position given for unit on dashboard and rebalanced positions.
		gDashboard, errCode := M.GetDashboard(project.ID, agent.UUID, dashboard.ID)
		assert.Equal(t, http.StatusFound, errCode)
		var currentPosition map[string]map[uint64]int
		err := json.Unmarshal((*gDashboard.UnitsPosition).RawMessage, &currentPosition)
		assert.Nil(t, err)
		assert.NotContains(t, currentPosition[M.GetUnitType(M.PresentationLine)], unit.ID)
		assert.Contains(t, currentPosition[M.GetUnitType(M.PresentationBar)], unit1.ID)
		// unit1 should be repositioned to 0.
		assert.Equal(t, currentPosition[M.GetUnitType(M.PresentationBar)][unit1.ID], 0)
		// delete should not affect postions other unit type.
		assert.Contains(t, currentPosition, M.GetUnitType(M.PresentationCard))
		assert.Equal(t, currentPosition[M.GetUnitType(M.PresentationCard)][unit2.ID], 0)
	})

	t.Run("DeleteDashboardUnit:Invalid", func(t *testing.T) {
		rName := U.RandomString(5)
		unit, errCode, _ := M.CreateDashboardUnit(project.ID, agent.UUID, &M.DashboardUnit{DashboardId: dashboard.ID,
			Title: rName, Presentation: M.PresentationLine, Query: postgres.Jsonb{json.RawMessage(`{}`)}})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, unit)

		// invalid project.
		errCode = M.DeleteDashboardUnit(0, agent.UUID, dashboard.ID, unit.ID)
		assert.Equal(t, http.StatusBadRequest, errCode)

		// invalid dashboard.
		errCode = M.DeleteDashboardUnit(project.ID, agent.UUID, 0, unit.ID)
		assert.Equal(t, http.StatusBadRequest, errCode)

		// invalid unit.
		errCode = M.DeleteDashboardUnit(project.ID, agent.UUID, dashboard.ID, 0)
		assert.Equal(t, http.StatusBadRequest, errCode)
	})

	t.Run("DeleteDashboardUnit:WithoutAccessToDashboard", func(t *testing.T) {
		rName := U.RandomString(5)
		dashboard, errCode := M.CreateDashboard(project.ID, agent.UUID, &M.Dashboard{Name: rName, Type: M.DashboardTypePrivate})
		assert.NotNil(t, dashboard)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.Equal(t, rName, dashboard.Name)

		rTitle := U.RandomString(5)
		dashboardUnit, errCode, _ := M.CreateDashboardUnit(project.ID, agent.UUID, &M.DashboardUnit{DashboardId: dashboard.ID,
			Title: rTitle, Presentation: M.PresentationLine, Query: postgres.Jsonb{json.RawMessage(`{}`)}})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboardUnit)

		errCode = M.DeleteDashboardUnit(project.ID, agent2.UUID, dashboard.ID, dashboardUnit.ID)
		assert.Equal(t, http.StatusForbidden, errCode)
	})
}

func TestUpdateDashboardUnit(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	agent2, errCode := SetupAgentReturnDAO("", "")
	assert.Equal(t, http.StatusCreated, errCode)
	_, errCode = M.CreateProjectAgentMappingWithDependencies(&M.ProjectAgentMapping{
		ProjectID: project.ID, AgentUUID: agent2.UUID})
	assert.Equal(t, http.StatusCreated, errCode)

	t.Run("UpdateDashboardUnit:Title", func(t *testing.T) {
		rName := U.RandomString(5)
		dashboard, errCode := M.CreateDashboard(project.ID, agent.UUID, &M.Dashboard{Name: rName, Type: M.DashboardTypeProjectVisible})
		assert.NotNil(t, dashboard)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.Equal(t, rName, dashboard.Name)

		unitName1 := U.RandomString(5)
		unit, errCode, _ := M.CreateDashboardUnit(project.ID, agent.UUID, &M.DashboardUnit{DashboardId: dashboard.ID,
			Title: unitName1, Presentation: M.PresentationLine, Query: postgres.Jsonb{json.RawMessage(`{}`)}})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, unit)

		unitName2 := U.RandomString(5)
		updatedDashboard, errCode := M.UpdateDashboardUnit(project.ID, agent.UUID, dashboard.ID, unit.ID, &M.DashboardUnit{Title: unitName2})
		assert.Equal(t, http.StatusAccepted, errCode)
		assert.Equal(t, unitName2, updatedDashboard.Title)

		// invalid title.
		updatedDashboard, errCode = M.UpdateDashboardUnit(project.ID, agent.UUID, dashboard.ID, unit.ID, &M.DashboardUnit{Title: ""})
		assert.Equal(t, http.StatusBadRequest, errCode)

		// invalid projectId.
		unitName3 := U.RandomString(5)
		updatedDashboard, errCode = M.UpdateDashboardUnit(0, agent.UUID, dashboard.ID, unit.ID, &M.DashboardUnit{Title: unitName3})
		assert.Equal(t, http.StatusBadRequest, errCode)

		// invalid agentUUID.
		updatedDashboard, errCode = M.UpdateDashboardUnit(project.ID, "", dashboard.ID, unit.ID, &M.DashboardUnit{Title: unitName3})
		assert.Equal(t, http.StatusBadRequest, errCode)

		// invalid dashboardId.
		updatedDashboard, errCode = M.UpdateDashboardUnit(project.ID, agent.UUID, 0, unit.ID, &M.DashboardUnit{Title: unitName3})
		assert.Equal(t, http.StatusBadRequest, errCode)

		// invalid unitId.
		updatedDashboard, errCode = M.UpdateDashboardUnit(project.ID, agent.UUID, dashboard.ID, 0, &M.DashboardUnit{Title: unitName3})
		assert.Equal(t, http.StatusBadRequest, errCode)
	})

	t.Run("UpdateDashboardUnit:WithoutAccessToDashboard", func(t *testing.T) {
		rName := U.RandomString(5)
		dashboard, errCode := M.CreateDashboard(project.ID, agent.UUID, &M.Dashboard{Name: rName, Type: M.DashboardTypePrivate})
		assert.NotNil(t, dashboard)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.Equal(t, rName, dashboard.Name)

		rTitle := U.RandomString(5)
		dashboardUnit, errCode, _ := M.CreateDashboardUnit(project.ID, agent.UUID, &M.DashboardUnit{DashboardId: dashboard.ID,
			Title: rTitle, Presentation: M.PresentationLine, Query: postgres.Jsonb{json.RawMessage(`{}`)}})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboardUnit)

		rTitle2 := U.RandomString(5)
		_, errCode = M.UpdateDashboardUnit(project.ID, agent2.UUID, dashboard.ID, dashboardUnit.ID, &M.DashboardUnit{Title: rTitle2})
		assert.Equal(t, http.StatusForbidden, errCode)
	})
}

func TestBaseQuery(t *testing.T) {
	var baseQuery M.BaseQuery
	for _, queryClass := range []string{M.QueryClassInsights, M.QueryClassFunnel, M.QueryClassChannel, M.QueryClassAttribution} {
		from, to := U.TimeNowUnix(), U.TimeNowUnix()+100
		if queryClass == M.QueryClassFunnel || queryClass == M.QueryClassInsights {
			baseQuery = &M.Query{Class: queryClass, From: from, To: to}
		} else if queryClass == M.QueryClassAttribution {
			baseQuery = &M.AttributionQuery{Class: queryClass, From: from, To: to}
		} else {
			baseQuery = &M.ChannelQueryUnit{Class: queryClass, Query: &M.ChannelQuery{From: from, To: to}}
		}
		baseQuery.SetQueryDateRange(from+10, to+15)
		updatedFrom, updatedTo := baseQuery.GetQueryDateRange()
		assert.Equal(t, from+10, updatedFrom)
		assert.Equal(t, to+15, updatedTo)
		assert.Equal(t, queryClass, baseQuery.GetClass())
	}
}

func TestCacheDashboardUnitsForProjectID(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	customerAccountID := U.RandomLowerAphaNumString(5)
	_, errCode := M.UpdateProjectSettings(project.ID, &M.ProjectSetting{
		IntAdwordsCustomerAccountId: &customerAccountID,
	})

	dashboardName := U.RandomString(5)
	dashboard, errCode := M.CreateDashboard(project.ID, agent.UUID, &M.Dashboard{Name: dashboardName, Type: M.DashboardTypeProjectVisible})
	assert.NotNil(t, dashboard)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Equal(t, dashboardName, dashboard.Name)

	dashboardUnitQueriesMap := make(map[uint64]map[string]interface{})
	var dashboardQueriesStr = map[string]string{
		M.QueryClassInsights:    `{"cl": "insights", "ec": "any_given_event", "fr": 1393612200, "to": 1396290599, "ty": "events_occurrence", "tz": "", "ewp": [{"na": "$session", "pr": []}], "gbp": [], "gbt": ""}`,
		M.QueryClassFunnel:      `{"cl": "funnel", "ec": "any_given_event", "fr": 1594492200, "to": 1594578599, "ty": "unique_users", "tz": "Asia/Calcutta", "ewp": [{"na": "$session", "pr": []}, {"na": "www.chargebee.com/schedule-a-demo", "pr": []}], "gbp": [], "gbt": ""}`,
		M.QueryClassAttribution: `{"cl": "attribution", "meta": {"metrics_breakdown": true}, "ce": "$session", "cl": "attribution", "cm": ["Impressions", "Clicks", "Spend"], "to": 1585679399, "lbw": 0, "lfe": [], "from": 1583001000, "attribution_key": "Campaign", "attribution_methodology": "First_Touch"}`,
		M.QueryClassChannel:     `{"cl": "channel", "meta": {"metric": "total_cost"}, "query": {"to": 1576060774, "from": 1573468774, "channel": "google_ads", "filter_key": "campaign", "filter_value": "all"}}`,
	}
	for queryClass, queryString := range dashboardQueriesStr {
		queryJSON := postgres.Jsonb{json.RawMessage(queryString)}
		baseQuery, err := M.DecodeQueryForClass(queryJSON, queryClass)
		assert.Nil(t, err)

		dashboardUnit, errCode, _ := M.CreateDashboardUnit(project.ID, agent.UUID, &M.DashboardUnit{
			DashboardId:  dashboard.ID,
			Title:        U.RandomString(5),
			Query:        queryJSON,
			Presentation: M.PresentationCard,
		})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboardUnit)

		dashboardUnitQueriesMap[dashboardUnit.ID] = make(map[string]interface{})
		dashboardUnitQueriesMap[dashboardUnit.ID]["class"] = queryClass
		dashboardUnitQueriesMap[dashboardUnit.ID]["query"] = baseQuery
	}

	updatedUnitsCount := M.CacheDashboardUnitsForProjectID(project.ID, 1)
	assert.Equal(t, 4, updatedUnitsCount)

	for _, rangeFunction := range U.QueryDateRangePresets {
		from, to := rangeFunction()
		for unitID, queryMap := range dashboardUnitQueriesMap {
			queryClass := queryMap["class"].(string)
			query := queryMap["query"].(M.BaseQuery)

			query.SetQueryDateRange(from, to)
			// Refresh is sent as false. Must return all presets range from cache.
			if queryClass == M.QueryClassAttribution {
				fmt.Println("Ab aayega maza")
			}
			w := sendAnalyticsQueryReq(r, queryClass, project.ID, agent, dashboard.ID, unitID, query, false)
			assert.NotNil(t, w)
			assert.Equal(t, http.StatusOK, w.Code)

			var result map[string]interface{}
			json.Unmarshal([]byte(w.Body.String()), &result)
			// Cache must be true in response.
			assert.True(t, result["cache"].(bool))

			// Refresh is sent as true. Still must return from cache for all presets except for todays.
			w = sendAnalyticsQueryReq(r, queryClass, project.ID, agent, dashboard.ID, unitID, query, true)
			assert.NotNil(t, w)
			assert.Equal(t, http.StatusOK, w.Code)
			result = nil
			json.Unmarshal([]byte(w.Body.String()), &result)

			if from == U.GetBeginningOfDayTimestampZ(U.TimeNowUnix(), U.TimeZoneStringIST) {
				// Today's preset. Must not be from cache.
				assert.False(t, result["cache"].(bool))

				// If queried again with refresh as false, should return from cache.
				w := sendAnalyticsQueryReq(r, queryClass, project.ID, agent, dashboard.ID, unitID, query, false)
				result = nil
				json.Unmarshal([]byte(w.Body.String()), &result)
				assert.True(t, result["cache"].(bool))
			} else {
				// Cache must be true in response.
				assert.True(t, result["cache"].(bool))
			}
		}
	}
}

func TestCacheDashboardUnitsForProjectIDForAttributionQuery(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	customerAccountId := U.RandomLowerAphaNumString(5)
	_, errCode := M.UpdateProjectSettings(project.ID, &M.ProjectSetting{
		IntAdwordsCustomerAccountId: &customerAccountId,
	})

	dashboardName := U.RandomString(5)
	dashboard, errCode := M.CreateDashboard(project.ID, agent.UUID, &M.Dashboard{Name: dashboardName, Type: M.DashboardTypeProjectVisible})
	assert.NotNil(t, dashboard)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Equal(t, dashboardName, dashboard.Name)

	var query M.AttributionQuery
	queryJSON := postgres.Jsonb{json.RawMessage(`{"cl": "attribution", "meta": {"metrics_breakdown": true}, "ce": "$session", "cl": "attribution", "cm": ["Impressions", "Clicks", "Spend"], "to": 1585679399, "lbw": 0, "lfe": [], "from": 1583001000, "attribution_key": "Campaign", "attribution_methodology": "First_Touch"}`)}
	U.DecodePostgresJsonbToStructType(&queryJSON, &query)
	dashboardUnit, errCode, _ := M.CreateDashboardUnit(project.ID, agent.UUID, &M.DashboardUnit{
		DashboardId:  dashboard.ID,
		Title:        U.RandomString(5),
		Query:        queryJSON,
		Presentation: M.PresentationTable,
	})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotNil(t, dashboardUnit)

	updatedUnitsCount := M.CacheDashboardUnitsForProjectID(project.ID, 1)
	assert.Equal(t, 1, updatedUnitsCount)

	for _, rangeFunction := range U.QueryDateRangePresets {
		query.From, query.To = rangeFunction()
		// Refresh is sent as false. Must return all presets range from cache.
		w := sendAttributionQueryReq(r, project.ID, agent, dashboard.ID, dashboardUnit.ID, query, false)
		assert.NotNil(t, w)
		assert.Equal(t, http.StatusOK, w.Code)

		// Refresh is sent as true. Still must return from cache for all presets except for todays.
		w = sendAttributionQueryReq(r, project.ID, agent, dashboard.ID, dashboardUnit.ID, query, true)
		assert.NotNil(t, w)
		var result map[string]interface{}
		json.Unmarshal([]byte(w.Body.String()), &result)
		// Cache must be true in response.
		assert.True(t, result["cache"].(bool))
		break
	}
}

func sendAttributionQueryReq(r *gin.Engine, projectID uint64, agent *M.Agent, dashboardID, unitID uint64, query M.AttributionQuery, refresh bool) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}
	queryPayload := H.AttributionRequestPayload{
		Query: &query,
	}

	rb := U.NewRequestBuilder(http.MethodPost, fmt.Sprintf(
		"/projects/%d/attribution/query?dashboard_id=%d&dashboard_unit_id=%d&refresh=%v", projectID, dashboardID, unitID, refresh)).
		WithPostParams(queryPayload).
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

func sendAnalyticsQueryReq(r *gin.Engine, queryClass string, projectID uint64, agent *M.Agent, dashboardID,
	unitID uint64, baseQuery M.BaseQuery, refresh bool) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}

	var queryURL string
	var queryPayload interface{}
	if queryClass == M.QueryClassFunnel || queryClass == M.QueryClassInsights {
		queryURL = "query"
		query := baseQuery.(*M.Query)
		queryPayload = H.QueryRequestPayload{
			Query: *query,
		}
	} else if queryClass == M.QueryClassChannel {
		queryURL = "channels/query"
		query := baseQuery.(*M.ChannelQueryUnit)
		queryPayload = query.Query
	} else {
		queryURL = "attribution/query"
		query := baseQuery.(*M.AttributionQuery)
		queryPayload = H.AttributionRequestPayload{
			Query: query,
		}
	}

	rb := U.NewRequestBuilder(http.MethodPost, fmt.Sprintf(
		"/projects/%d/%s?dashboard_id=%d&dashboard_unit_id=%d&refresh=%v", projectID, queryURL, dashboardID, unitID, refresh)).
		WithPostParams(queryPayload).
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
