package tests

import (
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
