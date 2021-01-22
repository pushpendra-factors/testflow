package tests

import (
	"encoding/json"
	C "factors/config"
	H "factors/handler"
	M "factors/model"
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
		dashboard, errCode := M.CreateAgentPersonalDashboardForProject(project.ID, agent.UUID)
		assert.NotNil(t, dashboard)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.Equal(t, M.AgentProjectPersonalDashboardName, dashboard.Name)
	})

	t.Run("CreateDashboardVisibleToAgents", func(t *testing.T) {
		rName := U.RandomString(5)
		dashboard, errCode := M.CreateDashboard(project.ID, agent.UUID,
			&M.Dashboard{Name: rName, Type: M.DashboardTypeProjectVisible})
		assert.NotNil(t, dashboard)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.Equal(t, rName, dashboard.Name)
	})

	t.Run("CreateDashboard:Invalid", func(t *testing.T) {
		// invalid name.
		dashboard, errCode := M.CreateDashboard(project.ID, agent.UUID, &M.Dashboard{Name: "", Type: M.DashboardTypePrivate})
		assert.Equal(t, http.StatusBadRequest, errCode)
		assert.Nil(t, dashboard)

		// invalid project id.
		rName := U.RandomString(5)
		dashboard, errCode = M.CreateDashboard(0, agent.UUID, &M.Dashboard{Name: rName, Type: M.DashboardTypePrivate})
		assert.Equal(t, http.StatusBadRequest, errCode)
		assert.Nil(t, dashboard)

		// invalid agent.
		rName = U.RandomString(5)
		dashboard, errCode = M.CreateDashboard(project.ID, "", &M.Dashboard{Name: rName, Type: M.DashboardTypePrivate})
		assert.Equal(t, http.StatusBadRequest, errCode)
		assert.Nil(t, dashboard)

		// invalid type.
		rName = U.RandomString(5)
		dashboard, errCode = M.CreateDashboard(project.ID, agent.UUID, &M.Dashboard{Name: rName, Type: ""})
		assert.Equal(t, http.StatusBadRequest, errCode)
		assert.Nil(t, dashboard)
	})
}

func TestGetDashboards(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	agent2, errCode := SetupAgentReturnDAO(getRandomEmail(), "+13425356")
	assert.Equal(t, http.StatusCreated, errCode)
	_, errCode = M.CreateProjectAgentMappingWithDependencies(&M.ProjectAgentMapping{
		ProjectID: project.ID, AgentUUID: agent2.UUID})
	assert.Equal(t, http.StatusCreated, errCode)

	t.Run("GetDashboards:NotCreated", func(t *testing.T) {
		dashboards, errCode := M.GetDashboards(project.ID, agent.UUID)
		assert.Equal(t, http.StatusFound, errCode)
		assert.Len(t, dashboards, 1) // default dashboard.
	})

	t.Run("GetDashboards:AfterCreation", func(t *testing.T) {
		rName1 := U.RandomString(5)
		dashboard, errCode := M.CreateDashboard(project.ID, agent.UUID, &M.Dashboard{Name: rName1, Type: M.DashboardTypePrivate})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboard)
		assert.Equal(t, http.StatusCreated, errCode)
		rName2 := U.RandomString(5)
		dashboard, errCode = M.CreateDashboard(project.ID, agent.UUID, &M.Dashboard{Name: rName2, Type: M.DashboardTypePrivate})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboard)
		assert.Equal(t, http.StatusCreated, errCode)

		dashboards, errCode := M.GetDashboards(project.ID, agent.UUID)
		assert.Equal(t, http.StatusFound, errCode)
		assert.Len(t, dashboards, 3) // default dashboard.
		// validates ordering.
		assert.Equal(t, M.AgentProjectPersonalDashboardName, dashboards[0].Name)
		assert.Equal(t, rName1, dashboards[1].Name)
		assert.Equal(t, rName2, dashboards[2].Name)
	})

	t.Run("GetDashboards:AccessPrivate", func(t *testing.T) {
		rName1 := U.RandomString(5)
		dashboard, errCode := M.CreateDashboard(project.ID, agent.UUID, &M.Dashboard{Name: rName1, Type: M.DashboardTypePrivate})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboard)

		// Other agent sholuld not be able to access my private dashboard.
		dashboards, errCode := M.GetDashboards(project.ID, agent2.UUID)
		assert.Equal(t, http.StatusFound, errCode)
		for _, d := range dashboards {
			assert.NotEqual(t, rName1, d.Name)
		}

		// Creator should have access to private dashboard.
		dashboards, errCode = M.GetDashboards(project.ID, agent.UUID)
		assert.Equal(t, http.StatusFound, errCode)
		assert.Equal(t, rName1, dashboards[len(dashboards)-1].Name)
	})

	t.Run("GetDashboards:AccessProjectVisible", func(t *testing.T) {
		rName1 := U.RandomString(5)
		dashboard, errCode := M.CreateDashboard(project.ID, agent.UUID, &M.Dashboard{Name: rName1, Type: M.DashboardTypeProjectVisible})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboard)

		// All agents should be able to access a ProjectVisible dashboard.
		dashboards, errCode := M.GetDashboards(project.ID, agent2.UUID)
		assert.Equal(t, http.StatusFound, errCode)
		assert.Equal(t, rName1, dashboards[len(dashboards)-1].Name)

		// Creator should have access to project visible dashboard.
		dashboards, errCode = M.GetDashboards(project.ID, agent.UUID)
		assert.Equal(t, http.StatusFound, errCode)
		assert.Equal(t, rName1, dashboards[len(dashboards)-1].Name)
	})
}

func TestUpdateDashboard(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	agent2, errCode := SetupAgentReturnDAO("", "")
	assert.Equal(t, http.StatusCreated, errCode)
	_, errCode = M.CreateProjectAgentMappingWithDependencies(&M.ProjectAgentMapping{
		ProjectID: project.ID, AgentUUID: agent2.UUID})
	assert.Equal(t, http.StatusCreated, errCode)

	t.Run("UpdateDashboard:Name", func(t *testing.T) {
		rName1 := U.RandomString(5)
		dashboard, errCode := M.CreateDashboard(project.ID, agent.UUID, &M.Dashboard{Name: rName1, Type: M.DashboardTypePrivate})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboard)
		assert.Equal(t, http.StatusCreated, errCode)

		rName2 := U.RandomString(5)
		errCode = M.UpdateDashboard(project.ID, agent.UUID, dashboard.ID, &M.UpdatableDashboard{Name: rName2})
		assert.Equal(t, http.StatusAccepted, errCode)
		gDashboard, errCode := M.GetDashboard(project.ID, agent.UUID, dashboard.ID)
		assert.Equal(t, http.StatusFound, errCode)
		assert.Equal(t, rName2, gDashboard.Name)
	})

	t.Run("UpdateDashboard:UnitsPosition", func(t *testing.T) {
		rName1 := U.RandomString(5)
		dashboard, errCode := M.CreateDashboard(project.ID, agent.UUID, &M.Dashboard{Name: rName1, Type: M.DashboardTypePrivate})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboard)
		assert.Equal(t, http.StatusCreated, errCode)

		positions := map[string]map[uint64]int{
			M.UnitChart: map[uint64]int{
				1: 0,
				2: 1,
			},
		}
		errCode = M.UpdateDashboard(project.ID, agent.UUID, dashboard.ID, &M.UpdatableDashboard{UnitsPosition: &positions})
		assert.Equal(t, http.StatusAccepted, errCode)
		gDashboard, errCode := M.GetDashboard(project.ID, agent.UUID, dashboard.ID)
		assert.Equal(t, http.StatusFound, errCode)
		var gPositions map[string]map[uint64]int
		err := json.Unmarshal((gDashboard.UnitsPosition).RawMessage, &gPositions)
		assert.Nil(t, err)
		assert.Equal(t, positions, gPositions)

		validPositions := map[string]map[uint64]int{
			M.UnitChart: map[uint64]int{
				1: 0,
				2: 1,
			},
			M.UnitCard: map[uint64]int{
				4: 1,
				3: 0,
			},
		}
		errCode = M.UpdateDashboard(project.ID, agent.UUID, dashboard.ID, &M.UpdatableDashboard{UnitsPosition: &validPositions})
		assert.Equal(t, http.StatusAccepted, errCode)
	})
}

func TestGetDashboardResutlFromCache(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	customerAccountId := fmt.Sprintf("%d", U.RandomUint64())
	_, errCode := M.UpdateProjectSettings(project.ID, &M.ProjectSetting{
		IntAdwordsCustomerAccountId: &customerAccountId,
		IntAdwordsEnabledAgentUUID:  &agent.UUID,
	})
	assert.Equal(t, http.StatusAccepted, errCode)
	rName := U.RandomString(5)
	dashboard, errCode := M.CreateDashboard(project.ID, agent.UUID,
		&M.Dashboard{Name: rName, Type: M.DashboardTypeProjectVisible})
	assert.NotNil(t, dashboard)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Equal(t, rName, dashboard.Name)
	var from int64 = 1556602834
	var to int64 = 1557207634
	query1 := M.Query{
		EventsCondition: M.EventCondAnyGivenEvent,
		From:            from,
		To:              to,
		Type:            M.QueryTypeEventsOccurrence,
		EventsWithProperties: []M.QueryEventWithProperties{
			M.QueryEventWithProperties{
				Name: "event1",
			},
		},
		OverridePeriod: true,
	}
	query2 := M.Query{
		EventsCondition: M.EventCondAnyGivenEvent,
		From:            from + 500,
		To:              to + 500,
		Type:            M.QueryTypeEventsOccurrence,
		EventsWithProperties: []M.QueryEventWithProperties{
			M.QueryEventWithProperties{
				Name: "event1",
			},
		},
		OverridePeriod: true,
	}

	query1Json, err := json.Marshal(query1)
	assert.Nil(t, err)
	query2Json, err := json.Marshal(query2)
	assert.Nil(t, err)

	rTitle := U.RandomString(5)
	w := sendCreateDashboardUnitReq(r, project.ID, agent, dashboard.ID, &M.DashboardUnitRequestPayload{Title: rTitle,
		Query: &postgres.Jsonb{query1Json}, Presentation: M.PresentationLine})
	assert.Equal(t, http.StatusCreated, w.Code)
	rTitle = U.RandomString(5)
	w = sendCreateDashboardUnitReq(r, project.ID, agent, dashboard.ID, &M.DashboardUnitRequestPayload{Title: rTitle,
		Query: &postgres.Jsonb{query2Json}, Presentation: M.PresentationLine})
	assert.Equal(t, http.StatusCreated, w.Code)

	//For Channel query
	value := []byte(`{"id": 2061667885,"clicks":989, "campaign_id": 12,"impressions":10, "end_date": "20371230", "start_date": "20190711", "conversions":111, "cost":42.94}`)
	document := M.AdwordsDocument{
		ProjectID:         project.ID,
		CustomerAccountID: customerAccountId,
		Type:              5,
		Timestamp:         20191209,
		ID:                "2061667885",
		Value:             &postgres.Jsonb{value},
		TypeAlias:         "campaign_performance_report",
	}
	errCode = M.CreateAdwordsDocument(&document)
	assert.Equal(t, http.StatusCreated, errCode)
	query3 := &M.ChannelQuery{
		Channel:     "google_ads",
		FilterKey:   "campaign",
		FilterValue: "all",
		From:        1575158400,
		To:          1575936000,
	}

	rTitle = U.RandomString(5)
	query3Json, err := json.Marshal(query3)
	assert.Nil(t, err)

	w = sendCreateDashboardUnitReq(r, project.ID, agent, dashboard.ID, &M.DashboardUnitRequestPayload{
		Presentation: "pc",
		Query:        &postgres.Jsonb{query3Json},
		Title:        rTitle,
	})
	assert.Equal(t, http.StatusCreated, w.Code)

	dashboards, errCode := M.GetDashboards(project.ID, agent.UUID)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, dashboard.Name, dashboards[1].Name)

	// No of units should be 3
	dashboardUnits, errCode := M.GetDashboardUnits(project.ID, agent.UUID, dashboards[1].ID)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, 3, len(dashboardUnits))

	decResult := struct {
		Cache  bool          `json:"cache"`
		Result M.QueryResult `json:"result"`
	}{}

	decChannelResult := struct {
		Cache  bool                 `json:"cache"`
		Result M.ChannelQueryResult `json:"result"`
	}{}

	//Cache should be empty
	result, errCode, errMsg := M.GetCacheResultByDashboardIdAndUnitId(project.ID, dashboardUnits[0].DashboardId, dashboardUnits[0].ID, from, to)
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
	result, errCode, errMsg = M.GetCacheResultByDashboardIdAndUnitId(project.ID, dashboardUnits[0].DashboardId, dashboardUnits[0].ID, from, to)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Nil(t, errMsg)
	assert.Equal(t, float64(query1.To), result.Result.(map[string]interface{})["meta"].(map[string]interface{})["query"].(map[string]interface{})["to"])
	result, errCode, errMsg = M.GetCacheResultByDashboardIdAndUnitId(project.ID, dashboardUnits[1].DashboardId, dashboardUnits[1].ID, from+500, to+500)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Nil(t, errMsg)
	assert.Equal(t, float64(query2.To), result.Result.(map[string]interface{})["meta"].(map[string]interface{})["query"].(map[string]interface{})["to"])
	resultChannel, errCode, errMsg := M.GetCacheResultByDashboardIdAndUnitId(project.ID, dashboardUnits[2].DashboardId, dashboardUnits[2].ID, query3.From, query3.To)
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

	dashboardQuery, errCode, errMsg := M.CreateQuery(project.ID, &M.Queries{
		ProjectID: project.ID,
		Type:      M.QueryTypeDashboardQuery,
		Query:     postgres.Jsonb{json.RawMessage(`{}`)},
	})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Empty(t, errMsg)
	assert.NotNil(t, dashboardQuery)

	dashboard, errCode := M.CreateDashboard(project.ID, agent.UUID,
		&M.Dashboard{Name: U.RandomString(5), Type: M.DashboardTypeProjectVisible})
	assert.NotNil(t, dashboard)
	assert.Equal(t, http.StatusCreated, errCode)

	dashboardUnit, errCode, errMsg := M.CreateDashboardUnit(project.ID, agent.UUID,
		&M.DashboardUnit{DashboardId: dashboard.ID, Title: U.RandomString(5), Presentation: M.PresentationLine,
			QueryId: dashboardQuery.ID, Query: postgres.Jsonb{json.RawMessage(`{}`)}},
		M.DashboardUnitWithQueryID)
	assert.NotEmpty(t, dashboardUnit)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Empty(t, errMsg)

	report, errCode := M.CreateReport(&M.Report{DashboardID: dashboard.ID, ProjectID: project.ID})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, report)

	// Delete a dashboard having units with queries and reports. All should get marked deleted.
	errCode = M.DeleteDashboard(project.ID, agent.UUID, dashboard.ID)
	assert.Equal(t, http.StatusAccepted, errCode)

	var deletedQuery M.Queries
	var deletedUnit M.DashboardUnit
	var deletedReport M.DBReport

	db := C.GetServices().Db
	err = db.Model(M.Queries{}).Where("project_id = ? AND id = ?", dashboardQuery.ProjectID, dashboardQuery.ID).Find(&deletedQuery).Error
	assert.Nil(t, err)
	assert.True(t, deletedQuery.IsDeleted)
	err = db.Model(M.DashboardUnit{}).Where("project_id = ? AND id = ?", dashboardUnit.ProjectID, dashboardUnit.ID).Find(&deletedUnit).Error
	assert.Nil(t, err)
	assert.True(t, deletedUnit.IsDeleted)
	err = db.Model(M.Report{}).Where("project_id = ? AND id = ?", report.ProjectID, report.ID).Find(&deletedReport).Error
	assert.Nil(t, err)
	assert.True(t, deletedReport.IsDeleted)
}
