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
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	rName := U.RandomString(5)
	dashboard, errCode := M.CreateSharableDashboard(project.ID, &M.Dashboard{Name: rName})
	assert.NotNil(t, dashboard)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Equal(t, rName, dashboard.Name)

	t.Run("CreateDashboardUnit", func(t *testing.T) {
		rName := U.RandomString(5)
		dashboardUnit, errCode, errMsg := M.CreateDashboardUnit(project.ID, &M.DashboardUnit{DashboardId: dashboard.ID,
			Title: rName, Presentation: M.PresentationLine, Query: postgres.Jsonb{json.RawMessage(`{}`)}})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboardUnit)
		assert.Empty(t, errMsg)
	})

	t.Run("CreateDashboardUnit:Invalid", func(t *testing.T) {
		// invalid title.
		dashboardUnit, errCode, _ := M.CreateDashboardUnit(project.ID, &M.DashboardUnit{DashboardId: dashboard.ID,
			Title: "", Presentation: M.PresentationLine, Query: postgres.Jsonb{json.RawMessage(`{}`)}})
		assert.Equal(t, http.StatusBadRequest, errCode)
		assert.Nil(t, dashboardUnit)

		// invalid presentation.
		rName := U.RandomString(5)
		dashboardUnit, errCode, _ = M.CreateDashboardUnit(project.ID, &M.DashboardUnit{DashboardId: dashboard.ID,
			Title: rName, Presentation: "", Query: postgres.Jsonb{json.RawMessage(`{}`)}})
		assert.Equal(t, http.StatusBadRequest, errCode)
		assert.Nil(t, dashboardUnit)

		// invalid dashboard.
		rName = U.RandomString(5)
		dashboardUnit, errCode, _ = M.CreateDashboardUnit(project.ID, &M.DashboardUnit{DashboardId: 0,
			Title: rName, Presentation: M.PresentationLine, Query: postgres.Jsonb{json.RawMessage(`{}`)}})
		assert.Equal(t, http.StatusBadRequest, errCode)
		assert.Nil(t, dashboardUnit)

		// invalid project.
		rName = U.RandomString(5)
		dashboardUnit, errCode, _ = M.CreateDashboardUnit(0, &M.DashboardUnit{DashboardId: dashboard.ID,
			Title: rName, Presentation: M.PresentationLine, Query: postgres.Jsonb{json.RawMessage(`{}`)}})
		assert.Equal(t, http.StatusBadRequest, errCode)
		assert.Nil(t, dashboardUnit)
	})
}

func TestGetDashboardUnits(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	rName := U.RandomString(5)
	dashboard, errCode := M.CreateSharableDashboard(project.ID, &M.Dashboard{Name: rName})
	assert.NotNil(t, dashboard)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Equal(t, rName, dashboard.Name)

	t.Run("GetDashboardUnits:NotAvailable", func(t *testing.T) {
		units, errCode := M.GetDashboardUnits(project.ID, dashboard.ID)
		assert.Equal(t, http.StatusFound, errCode)
		assert.Len(t, units, 0)
	})

	t.Run("GetDashboardUnits:Available", func(t *testing.T) {
		rName := U.RandomString(5)
		dashboardUnit, errCode, errMsg := M.CreateDashboardUnit(project.ID, &M.DashboardUnit{DashboardId: dashboard.ID,
			Title: rName, Presentation: M.PresentationLine, Query: postgres.Jsonb{json.RawMessage(`{}`)}})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboardUnit)
		assert.Empty(t, errMsg)

		units, errCode := M.GetDashboardUnits(project.ID, dashboard.ID)
		assert.Equal(t, http.StatusFound, errCode)
		assert.Len(t, units, 1)
		assert.Equal(t, rName, units[0].Title)
		assert.Equal(t, M.PresentationLine, units[0].Presentation)
	})

	t.Run("GetDashboardUnits:Invalid", func(t *testing.T) {
		rName := U.RandomString(5)
		dashboardUnit, errCode, errMsg := M.CreateDashboardUnit(project.ID, &M.DashboardUnit{DashboardId: dashboard.ID,
			Title: rName, Presentation: M.PresentationLine, Query: postgres.Jsonb{json.RawMessage(`{}`)}})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboardUnit)
		assert.Empty(t, errMsg)

		// invalid project
		units, errCode := M.GetDashboardUnits(0, dashboard.ID)
		assert.Equal(t, http.StatusBadRequest, errCode)
		assert.Nil(t, units)

		// invalid dashboard
		units, errCode = M.GetDashboardUnits(project.ID, 0)
		assert.Equal(t, http.StatusBadRequest, errCode)
		assert.Nil(t, units)
	})
}

func TestDeleteDashboardUnit(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	rName := U.RandomString(5)
	dashboard, errCode := M.CreateSharableDashboard(project.ID, &M.Dashboard{Name: rName})
	assert.NotNil(t, dashboard)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Equal(t, rName, dashboard.Name)

	t.Run("DeleteDashboardUnit", func(t *testing.T) {
		rName := U.RandomString(5)
		unit, errCode, _ := M.CreateDashboardUnit(project.ID, &M.DashboardUnit{DashboardId: dashboard.ID,
			Title: rName, Presentation: M.PresentationLine, Query: postgres.Jsonb{json.RawMessage(`{}`)}})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, unit)

		errCode = M.DeleteDashboardUnit(project.ID, dashboard.ID, unit.ID)
		assert.Equal(t, http.StatusAccepted, errCode)
	})

	t.Run("DeleteDashboardUnit:Invalid", func(t *testing.T) {
		rName := U.RandomString(5)
		unit, errCode, _ := M.CreateDashboardUnit(project.ID, &M.DashboardUnit{DashboardId: dashboard.ID,
			Title: rName, Presentation: M.PresentationLine, Query: postgres.Jsonb{json.RawMessage(`{}`)}})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, unit)

		// invalid project.
		errCode = M.DeleteDashboardUnit(0, dashboard.ID, unit.ID)
		assert.Equal(t, http.StatusBadRequest, errCode)

		// invalid dashboard.
		errCode = M.DeleteDashboardUnit(project.ID, 0, unit.ID)
		assert.Equal(t, http.StatusBadRequest, errCode)

		// invalid unit.
		errCode = M.DeleteDashboardUnit(project.ID, dashboard.ID, 0)
		assert.Equal(t, http.StatusBadRequest, errCode)
	})
}
