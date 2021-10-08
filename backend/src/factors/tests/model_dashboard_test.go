package tests

import (
	"encoding/json"
	H "factors/handler"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
)

func TestCreateDashboard(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	t.Run("CreatePersonalDashboard", func(t *testing.T) {
		dashboard, errCode := store.GetStore().CreateAgentPersonalDashboardForProject(project.ID, agent.UUID)
		assert.NotNil(t, dashboard)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.Equal(t, model.AgentProjectPersonalDashboardName, dashboard.Name)
	})

	t.Run("CreateDashboardVisibleToAgents", func(t *testing.T) {
		rName := U.RandomString(5)
		dashboard, errCode := store.GetStore().CreateDashboard(project.ID, agent.UUID,
			&model.Dashboard{Name: rName, Type: model.DashboardTypeProjectVisible})
		assert.NotNil(t, dashboard)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.Equal(t, rName, dashboard.Name)
	})

	t.Run("CreateDashboard:Invalid", func(t *testing.T) {
		// invalid name.
		dashboard, errCode := store.GetStore().CreateDashboard(project.ID, agent.UUID, &model.Dashboard{Name: "", Type: model.DashboardTypePrivate})
		assert.Equal(t, http.StatusBadRequest, errCode)
		assert.Nil(t, dashboard)

		// invalid project id.
		rName := U.RandomString(5)
		dashboard, errCode = store.GetStore().CreateDashboard(0, agent.UUID, &model.Dashboard{Name: rName, Type: model.DashboardTypePrivate})
		assert.Equal(t, http.StatusBadRequest, errCode)
		assert.Nil(t, dashboard)

		// invalid agent.
		rName = U.RandomString(5)
		dashboard, errCode = store.GetStore().CreateDashboard(project.ID, "", &model.Dashboard{Name: rName, Type: model.DashboardTypePrivate})
		assert.Equal(t, http.StatusBadRequest, errCode)
		assert.Nil(t, dashboard)

		// invalid type.
		rName = U.RandomString(5)
		dashboard, errCode = store.GetStore().CreateDashboard(project.ID, agent.UUID, &model.Dashboard{Name: rName, Type: ""})
		assert.Equal(t, http.StatusBadRequest, errCode)
		assert.Nil(t, dashboard)
	})
}

func TestGetDashboards(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	agent2, errCode := SetupAgentReturnDAO(getRandomEmail(), "+13425356")
	assert.Equal(t, http.StatusCreated, errCode)
	_, errCode = store.GetStore().CreateProjectAgentMappingWithDependencies(&model.ProjectAgentMapping{
		ProjectID: project.ID, AgentUUID: agent2.UUID})
	assert.Equal(t, http.StatusCreated, errCode)

	t.Run("GetDashboards:NotCreated", func(t *testing.T) {
		dashboards, errCode := store.GetStore().GetDashboards(project.ID, agent.UUID)
		assert.Equal(t, http.StatusFound, errCode)
		assert.Len(t, dashboards, 1) // default dashboard.
	})

	t.Run("GetDashboards:AfterCreation", func(t *testing.T) {
		rName1 := U.RandomString(5)
		dashboard, errCode := store.GetStore().CreateDashboard(project.ID, agent.UUID, &model.Dashboard{Name: rName1, Type: model.DashboardTypePrivate})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboard)
		assert.Equal(t, http.StatusCreated, errCode)
		rName2 := U.RandomString(5)
		dashboard, errCode = store.GetStore().CreateDashboard(project.ID, agent.UUID, &model.Dashboard{Name: rName2, Type: model.DashboardTypePrivate})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboard)
		assert.Equal(t, http.StatusCreated, errCode)

		dashboards, errCode := store.GetStore().GetDashboards(project.ID, agent.UUID)
		assert.Equal(t, http.StatusFound, errCode)
		assert.Len(t, dashboards, 3) // default dashboard.
		// validates ordering.
		assert.Equal(t, model.AgentProjectPersonalDashboardName, dashboards[0].Name)
		assert.Equal(t, rName1, dashboards[1].Name)
		assert.Equal(t, rName2, dashboards[2].Name)
	})

	t.Run("GetDashboards:AccessPrivate", func(t *testing.T) {
		rName1 := U.RandomString(5)
		dashboard, errCode := store.GetStore().CreateDashboard(project.ID, agent.UUID, &model.Dashboard{Name: rName1, Type: model.DashboardTypePrivate})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboard)

		// Other agent sholuld not be able to access my private dashboard.
		dashboards, errCode := store.GetStore().GetDashboards(project.ID, agent2.UUID)
		assert.Equal(t, http.StatusFound, errCode)
		for _, d := range dashboards {
			assert.NotEqual(t, rName1, d.Name)
		}

		// Creator should have access to private dashboard.
		dashboards, errCode = store.GetStore().GetDashboards(project.ID, agent.UUID)
		assert.Equal(t, http.StatusFound, errCode)
		assert.Equal(t, rName1, dashboards[len(dashboards)-1].Name)
	})

	t.Run("GetDashboards:AccessProjectVisible", func(t *testing.T) {
		rName1 := U.RandomString(5)
		dashboard, errCode := store.GetStore().CreateDashboard(project.ID, agent.UUID, &model.Dashboard{Name: rName1, Type: model.DashboardTypeProjectVisible})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboard)

		// All agents should be able to access a ProjectVisible dashboard.
		dashboards, errCode := store.GetStore().GetDashboards(project.ID, agent2.UUID)
		assert.Equal(t, http.StatusFound, errCode)
		assert.Equal(t, rName1, dashboards[len(dashboards)-1].Name)

		// Creator should have access to project visible dashboard.
		dashboards, errCode = store.GetStore().GetDashboards(project.ID, agent.UUID)
		assert.Equal(t, http.StatusFound, errCode)
		assert.Equal(t, rName1, dashboards[len(dashboards)-1].Name)
	})
}

func TestUpdateDashboard(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	agent2, errCode := SetupAgentReturnDAO("", "")
	assert.Equal(t, http.StatusCreated, errCode)
	_, errCode = store.GetStore().CreateProjectAgentMappingWithDependencies(&model.ProjectAgentMapping{
		ProjectID: project.ID, AgentUUID: agent2.UUID})
	assert.Equal(t, http.StatusCreated, errCode)

	t.Run("UpdateDashboard:Name", func(t *testing.T) {
		rName1 := U.RandomString(5)
		dashboard, errCode := store.GetStore().CreateDashboard(project.ID, agent.UUID, &model.Dashboard{Name: rName1, Type: model.DashboardTypePrivate})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboard)
		assert.Equal(t, http.StatusCreated, errCode)

		rName2 := U.RandomString(5)
		errCode = store.GetStore().UpdateDashboard(project.ID, agent.UUID, dashboard.ID, &model.UpdatableDashboard{Name: rName2})
		assert.Equal(t, http.StatusAccepted, errCode)
		gDashboard, errCode := store.GetStore().GetDashboard(project.ID, agent.UUID, dashboard.ID)
		assert.Equal(t, http.StatusFound, errCode)
		assert.Equal(t, rName2, gDashboard.Name)
	})

	t.Run("UpdateDashboard:UnitsPosition", func(t *testing.T) {
		rName1 := U.RandomString(5)
		dashboard, errCode := store.GetStore().CreateDashboard(project.ID, agent.UUID, &model.Dashboard{Name: rName1, Type: model.DashboardTypePrivate})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboard)
		assert.Equal(t, http.StatusCreated, errCode)

		positions := map[string]map[uint64]int{
			model.UnitChart: map[uint64]int{
				1: 0,
				2: 1,
			},
		}
		errCode = store.GetStore().UpdateDashboard(project.ID, agent.UUID, dashboard.ID, &model.UpdatableDashboard{UnitsPosition: &positions})
		assert.Equal(t, http.StatusAccepted, errCode)
		gDashboard, errCode := store.GetStore().GetDashboard(project.ID, agent.UUID, dashboard.ID)
		assert.Equal(t, http.StatusFound, errCode)
		var gPositions map[string]map[uint64]int
		err := json.Unmarshal((gDashboard.UnitsPosition).RawMessage, &gPositions)
		assert.Nil(t, err)
		assert.Equal(t, positions, gPositions)

		validPositions := map[string]map[uint64]int{
			model.UnitChart: map[uint64]int{
				1: 0,
				2: 1,
			},
			model.UnitCard: map[uint64]int{
				4: 1,
				3: 0,
			},
		}
		errCode = store.GetStore().UpdateDashboard(project.ID, agent.UUID, dashboard.ID, &model.UpdatableDashboard{UnitsPosition: &validPositions})
		assert.Equal(t, http.StatusAccepted, errCode)
	})
}

func TestGetDashboardResultFromCache(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	customerAccountId := fmt.Sprintf("%d", U.RandomUint64())
	_, errCode := store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		IntAdwordsCustomerAccountId: &customerAccountId,
		IntAdwordsEnabledAgentUUID:  &agent.UUID,
	})
	assert.Equal(t, http.StatusAccepted, errCode)
	rName := U.RandomString(5)
	dashboard, errCode := store.GetStore().CreateDashboard(project.ID, agent.UUID,
		&model.Dashboard{Name: rName, Type: model.DashboardTypeProjectVisible})
	assert.NotNil(t, dashboard)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Equal(t, rName, dashboard.Name)
	var from int64 = 1556602834
	var to int64 = 1557207634
	query1 := model.Query{
		EventsCondition: model.EventCondAnyGivenEvent,
		From:            from,
		To:              to,
		Type:            model.QueryTypeEventsOccurrence,
		EventsWithProperties: []model.QueryEventWithProperties{
			model.QueryEventWithProperties{
				Name: "event1",
			},
		},
		OverridePeriod: true,
	}
	query2 := model.Query{
		EventsCondition: model.EventCondAnyGivenEvent,
		From:            from + 500,
		To:              to + 500,
		Type:            model.QueryTypeEventsOccurrence,
		EventsWithProperties: []model.QueryEventWithProperties{
			model.QueryEventWithProperties{
				Name: "event1",
			},
		},
		OverridePeriod: true,
	}

	w := sendCreateDashboardUnitReq(r, project.ID, agent, dashboard.ID, &model.DashboardUnitRequestPayload{Presentation: model.PresentationLine})
	assert.Equal(t, http.StatusCreated, w.Code)
	w = sendCreateDashboardUnitReq(r, project.ID, agent, dashboard.ID, &model.DashboardUnitRequestPayload{Presentation: model.PresentationLine})
	assert.Equal(t, http.StatusCreated, w.Code)

	//For Channel query
	value := []byte(`{"id": 2061667885,"clicks":989, "campaign_id": 12,"impressions":10, "end_date": "20371230", "start_date": "20190711", "conversions":111, "cost":42.94}`)
	document := model.AdwordsDocument{
		ProjectID:         project.ID,
		CustomerAccountID: customerAccountId,
		Type:              5,
		Timestamp:         20191209,
		ID:                "2061667885",
		Value:             &postgres.Jsonb{value},
		TypeAlias:         "campaign_performance_report",
	}
	errCode = store.GetStore().CreateAdwordsDocument(&document)
	assert.Equal(t, http.StatusCreated, errCode)
	query3 := &model.ChannelQuery{
		Channel:     "google_ads",
		FilterKey:   "campaign",
		FilterValue: "all",
		From:        1575158400,
		To:          1575936000,
		Timezone:    string(U.TimeZoneStringIST),
	}

	w = sendCreateDashboardUnitReq(r, project.ID, agent, dashboard.ID, &model.DashboardUnitRequestPayload{
		Presentation: "pc",
	})
	assert.Equal(t, http.StatusCreated, w.Code)

	dashboards, errCode := store.GetStore().GetDashboards(project.ID, agent.UUID)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, dashboard.Name, dashboards[1].Name)

	// No of units should be 3
	dashboardUnits, errCode := store.GetStore().GetDashboardUnits(project.ID, agent.UUID, dashboards[1].ID)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, 3, len(dashboardUnits))

	decResult := struct {
		Cache  bool              `json:"cache"`
		Result model.QueryResult `json:"result"`
	}{}

	decChannelResult := struct {
		Cache  bool                     `json:"cache"`
		Result model.ChannelQueryResult `json:"result"`
	}{}

	//Cache should be empty
	result, errCode, errMsg := model.GetCacheResultByDashboardIdAndUnitId("", project.ID, dashboardUnits[0].DashboardId, dashboardUnits[0].ID, from, to, U.TimeZoneString(project.TimeZone))
	assert.Equal(t, http.StatusNotFound, errCode)
	assert.Nil(t, result)

	// Should set cache on first query with cache = false
	w = sendGetDashboardUnitResult(r, project.ID, agent, dashboardUnits[0].DashboardId, dashboardUnits[0].ID, &gin.H{"query": query1})
	assert.Equal(t, http.StatusOK, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &decResult)
	assert.Nil(t, err)
	assert.Equal(t, query1.To, decResult.Result.Meta.Query.To)
	assert.Equal(t, false, decResult.Cache)
	w = sendGetDashboardUnitResult(r, project.ID, agent, dashboardUnits[1].DashboardId, dashboardUnits[1].ID, &gin.H{"query": query2})
	assert.Equal(t, http.StatusOK, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &decResult)
	assert.Nil(t, err)
	assert.Equal(t, query2.To, decResult.Result.Meta.Query.To)
	assert.Equal(t, false, decResult.Cache)
	w = sendGetDashboardUnitChannelResult(r, project.ID, agent, dashboardUnits[2].DashboardId, dashboardUnits[2].ID, query3)
	assert.Equal(t, http.StatusOK, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &decChannelResult)
	assert.Nil(t, err)
	assert.Equal(t, float64(989), (*decChannelResult.Result.Metrics)["clicks"])
	assert.Equal(t, false, decChannelResult.Cache)

	// Cache should be set
	result, errCode, errMsg = model.GetCacheResultByDashboardIdAndUnitId("", project.ID, dashboardUnits[0].DashboardId, dashboardUnits[0].ID, from, to, U.TimeZoneString(project.TimeZone))
	assert.Equal(t, http.StatusFound, errCode)
	assert.Nil(t, errMsg)
	assert.Equal(t, float64(query1.To), result.Result.(map[string]interface{})["meta"].(map[string]interface{})["query"].(map[string]interface{})["to"])
	result, errCode, errMsg = model.GetCacheResultByDashboardIdAndUnitId("", project.ID, dashboardUnits[1].DashboardId, dashboardUnits[1].ID, from+500, to+500, U.TimeZoneString(project.TimeZone))
	assert.Equal(t, http.StatusFound, errCode)
	assert.Nil(t, errMsg)
	assert.Equal(t, float64(query2.To), result.Result.(map[string]interface{})["meta"].(map[string]interface{})["query"].(map[string]interface{})["to"])
	resultChannel, errCode, errMsg := model.GetCacheResultByDashboardIdAndUnitId("", project.ID, dashboardUnits[2].DashboardId, dashboardUnits[2].ID, query3.From, query3.To, U.TimeZoneString(project.TimeZone))
	assert.Equal(t, http.StatusFound, errCode)
	assert.Nil(t, errMsg)
	assert.Equal(t, float64(989), resultChannel.Result.(map[string]interface{})["metrics"].(map[string]interface{})["clicks"])

	// Cache should be set to true
	w = sendGetDashboardUnitResult(r, project.ID, agent, dashboardUnits[0].DashboardId, dashboardUnits[0].ID, &gin.H{"query": query1})
	assert.Equal(t, http.StatusOK, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &decResult)
	assert.Nil(t, err)
	assert.Equal(t, true, decResult.Cache)
}

func TestDeleteDashboard(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	dashboardQuery, errCode, errMsg := store.GetStore().CreateQuery(project.ID, &model.Queries{
		ProjectID: project.ID,
		Type:      model.QueryTypeDashboardQuery,
		Query:     postgres.Jsonb{json.RawMessage(`{}`)},
	})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Empty(t, errMsg)
	assert.NotNil(t, dashboardQuery)

	dashboard, errCode := store.GetStore().CreateDashboard(project.ID, agent.UUID,
		&model.Dashboard{Name: U.RandomString(5), Type: model.DashboardTypeProjectVisible})
	assert.NotNil(t, dashboard)
	assert.Equal(t, http.StatusCreated, errCode)

	dashboardUnit, errCode, errMsg := store.GetStore().CreateDashboardUnit(project.ID, agent.UUID,
		&model.DashboardUnit{DashboardId: dashboard.ID, Presentation: model.PresentationLine,
			QueryId: dashboardQuery.ID})
	assert.NotEmpty(t, dashboardUnit)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Empty(t, errMsg)

	// Delete a dashboard having units with queries and reports.
	errCode = store.GetStore().DeleteDashboard(project.ID, agent.UUID, dashboard.ID)
	assert.Equal(t, http.StatusAccepted, errCode)

	// Query should not get marked deleted.
	_, errCode = store.GetStore().GetQueryWithQueryId(dashboardQuery.ProjectID, dashboardQuery.ID)
	assert.Equal(t, http.StatusFound, errCode)

	// Dashboard unit should get deleted.
	_, errCode = store.GetStore().GetDashboardUnitByUnitID(dashboardUnit.ProjectID, dashboardUnit.ID)
	assert.Equal(t, http.StatusNotFound, errCode)
}

func TestShouldRefreshDashboardUnit(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	dashboardQuery, errCode, errMsg := store.GetStore().CreateQuery(project.ID, &model.Queries{
		ProjectID: project.ID,
		Type:      model.QueryTypeDashboardQuery,
		Query:     postgres.Jsonb{RawMessage: json.RawMessage(`{}`)},
	})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Empty(t, errMsg)
	assert.NotNil(t, dashboardQuery)

	dashboard, errCode := store.GetStore().CreateDashboard(project.ID, agent.UUID,
		&model.Dashboard{Name: U.RandomString(5), Type: model.DashboardTypeProjectVisible})
	assert.NotNil(t, dashboard)
	assert.Equal(t, http.StatusCreated, errCode)

	dashboardUnit, errCode, errMsg := store.GetStore().CreateDashboardUnit(project.ID, agent.UUID,
		&model.DashboardUnit{DashboardId: dashboard.ID, Presentation: model.PresentationLine,
			QueryId: dashboardQuery.ID})
	assert.NotEmpty(t, dashboardUnit)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Empty(t, errMsg)

	timezoneString := U.TimeZoneString(project.TimeZone)
	// 30mins range. Should allow.
	from, to, _ := U.WebAnalyticsQueryDateRangePresets[U.DateRangePreset30Minutes](timezoneString)
	assert.True(t, model.ShouldRefreshDashboardUnit(project.ID, dashboard.ID, 0, from, to, timezoneString, true))

	// Todays range. Should allow.
	from, to, _ = U.QueryDateRangePresets[U.DateRangePresetToday](timezoneString)
	assert.True(t, model.ShouldRefreshDashboardUnit(project.ID, dashboard.ID, 0, from, to, timezoneString, true))
	assert.True(t, model.ShouldRefreshDashboardUnit(project.ID, dashboard.ID, dashboardUnit.ID, from, to, timezoneString, false))

	// Yesterday's range. Should allow first time.
	from, to, _ = U.QueryDateRangePresets[U.DateRangePresetYesterday](timezoneString)
	assert.True(t, model.ShouldRefreshDashboardUnit(project.ID, dashboard.ID, 0, from, to, timezoneString, true))
	assert.True(t, model.ShouldRefreshDashboardUnit(project.ID, dashboard.ID, dashboardUnit.ID, from, to, timezoneString, false))

	// Yesterday's range. Should not allow again on same day once cache is set.
	from, to, _ = U.QueryDateRangePresets[U.DateRangePresetYesterday](timezoneString)
	model.SetCacheResultForWebAnalyticsDashboard(&model.WebAnalyticsQueryResult{}, project.ID, dashboard.ID, from, to, timezoneString)
	assert.False(t, model.ShouldRefreshDashboardUnit(project.ID, dashboard.ID, 0, from, to, timezoneString, true))
	model.SetCacheResultByDashboardIdAndUnitId("{}", project.ID, dashboard.ID, dashboardUnit.ID, from, to, timezoneString)
	assert.False(t, model.ShouldRefreshDashboardUnit(project.ID, dashboard.ID, dashboardUnit.ID, from, to, timezoneString, false))

	// More than 2 days old range. Should allow.
	from, to, _ = U.QueryDateRangePresets[U.DateRangePresetYesterday](timezoneString)
	from = from - 30*U.SECONDS_IN_A_DAY
	to = to - 2*U.SECONDS_IN_A_DAY
	assert.True(t, model.ShouldRefreshDashboardUnit(project.ID, dashboard.ID, 0, from, to, timezoneString, true))
	assert.True(t, model.ShouldRefreshDashboardUnit(project.ID, dashboard.ID, dashboardUnit.ID, from, to, timezoneString, false))
}
