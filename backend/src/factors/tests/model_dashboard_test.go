package tests

import (
	M "factors/model"
	U "factors/util"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateDashboard(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	t.Run("CreatePersonalDashboard", func(t *testing.T) {
		dashboard, errCode := M.CreatePersonalDashboard(project.ID)
		assert.NotNil(t, dashboard)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.Equal(t, M.DefaultNamePersonalDashboard, dashboard.Name)
	})

	t.Run("CreateSharableDashboard", func(t *testing.T) {
		rName := U.RandomString(5)
		dashboard, errCode := M.CreateSharableDashboard(project.ID, &M.Dashboard{Name: rName})
		assert.NotNil(t, dashboard)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.Equal(t, rName, dashboard.Name)
	})

	t.Run("CreateSharableDashboard:Invalid", func(t *testing.T) {
		dashboard, errCode := M.CreateSharableDashboard(project.ID, &M.Dashboard{Name: ""})
		assert.Equal(t, http.StatusBadRequest, errCode)
		assert.Nil(t, dashboard)

		rName := U.RandomString(5)
		dashboard, errCode = M.CreateSharableDashboard(0, &M.Dashboard{Name: rName})
		assert.Equal(t, http.StatusBadRequest, errCode)
		assert.Nil(t, dashboard)
	})
}

func TestGetDashboards(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	t.Run("GetDashboards:NotCreated", func(t *testing.T) {
		dashboards, errCode := M.GetDashboards(project.ID)
		assert.Equal(t, http.StatusFound, errCode)
		assert.Len(t, dashboards, 1) // default dashboard.
	})

	t.Run("GetDashboards:AfterCreation", func(t *testing.T) {
		rName1 := U.RandomString(5)
		dashboard, errCode := M.CreateSharableDashboard(project.ID, &M.Dashboard{Name: rName1})
		assert.NotNil(t, dashboard)
		assert.Equal(t, http.StatusCreated, errCode)
		rName2 := U.RandomString(5)
		dashboard, errCode = M.CreateSharableDashboard(project.ID, &M.Dashboard{Name: rName2})
		assert.NotNil(t, dashboard)
		assert.Equal(t, http.StatusCreated, errCode)

		dashboards, errCode := M.GetDashboards(project.ID)
		assert.Equal(t, http.StatusFound, errCode)
		assert.Len(t, dashboards, 3) // default dashboard.
		// validates ordering.
		assert.Equal(t, M.DefaultNamePersonalDashboard, dashboards[0].Name)
		assert.Equal(t, rName1, dashboards[1].Name)
		assert.Equal(t, rName2, dashboards[2].Name)
	})
}
