package tests

import (
	"encoding/json"
	M "factors/model"
	U "factors/util"
	"net/http"
	"testing"

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

	agent2, err := SetupAgentReturnDAO()
	assert.Nil(t, err)
	_, errCode := M.CreateProjectAgentMappingWithDependencies(&M.ProjectAgentMapping{
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

	agent2, err := SetupAgentReturnDAO()
	assert.Nil(t, err)
	_, errCode := M.CreateProjectAgentMappingWithDependencies(&M.ProjectAgentMapping{
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

		// invalid unit positions.
		invalidPositions := map[string]map[uint64]int{
			M.UnitChart: map[uint64]int{
				1: 1,
				2: 2,
			},
		}
		errCode = M.UpdateDashboard(project.ID, agent.UUID, dashboard.ID, &M.UpdatableDashboard{UnitsPosition: &invalidPositions})
		assert.Equal(t, http.StatusBadRequest, errCode)

		invalidPositions = map[string]map[uint64]int{
			M.UnitChart: map[uint64]int{
				1: 3,
				2: 1,
			},
		}
		errCode = M.UpdateDashboard(project.ID, agent.UUID, dashboard.ID, &M.UpdatableDashboard{UnitsPosition: &invalidPositions})
		assert.Equal(t, http.StatusBadRequest, errCode)

		invalidPositions = map[string]map[uint64]int{
			M.UnitChart: map[uint64]int{
				1: 0,
				2: 2, // out of order position
			},
		}
		errCode = M.UpdateDashboard(project.ID, agent.UUID, dashboard.ID, &M.UpdatableDashboard{UnitsPosition: &invalidPositions})
		assert.Equal(t, http.StatusBadRequest, errCode)

		invalidPositions = map[string]map[uint64]int{
			M.UnitChart: map[uint64]int{
				1: 0,
				2: 2,
			},
			M.UnitCard: map[uint64]int{
				1: 0, // duplicate id
				3: 1,
			},
		}
		errCode = M.UpdateDashboard(project.ID, agent.UUID, dashboard.ID, &M.UpdatableDashboard{UnitsPosition: &invalidPositions})
		assert.Equal(t, http.StatusBadRequest, errCode)

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
