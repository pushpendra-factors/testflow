package tests

import (
	"encoding/json"
	M "factors/model"
	U "factors/util"
	"net/http"
	"testing"

	"github.com/jinzhu/gorm/dialects/postgres"

	"github.com/stretchr/testify/assert"
)

func TestCreateDashboardUnit(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	agent2, err := SetupAgentReturnDAO()
	assert.Nil(t, err)
	_, errCode := M.CreateProjectAgentMappingWithDependencies(&M.ProjectAgentMapping{
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
		assert.Equal(t, http.StatusUnauthorized, errCode)
		assert.Nil(t, dashboardUnit)
	})
}

func TestGetDashboardUnits(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	agent2, err := SetupAgentReturnDAO()
	assert.Nil(t, err)
	_, errCode := M.CreateProjectAgentMappingWithDependencies(&M.ProjectAgentMapping{
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
		assert.Equal(t, http.StatusUnauthorized, errCode)
		assert.Nil(t, units)
	})
}

func TestDeleteDashboardUnit(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	agent2, err := SetupAgentReturnDAO()
	assert.Nil(t, err)
	_, errCode := M.CreateProjectAgentMappingWithDependencies(&M.ProjectAgentMapping{
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
		assert.Equal(t, http.StatusUnauthorized, errCode)
	})
}

func TestUpdateDashboardUnit(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	agent2, err := SetupAgentReturnDAO()
	assert.Nil(t, err)
	_, errCode := M.CreateProjectAgentMappingWithDependencies(&M.ProjectAgentMapping{
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
		assert.Equal(t, http.StatusUnauthorized, errCode)
	})
}
