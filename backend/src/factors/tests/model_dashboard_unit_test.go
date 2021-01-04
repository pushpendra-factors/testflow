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
	"reflect"
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
			Title: rName, Presentation: M.PresentationLine,
			Query: postgres.Jsonb{json.RawMessage(`{}`)}}, M.DashboardUnitForNoQueryID)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboardUnit)
		assert.Empty(t, errMsg)
		assert.NotNil(t, dashboardUnit.QueryId)
		assert.NotNil(t, dashboardUnit.Settings)
		assert.NotNil(t, dashboardUnit.Presentation)

		rName1 := U.RandomString(5)
		dashboardUnit1, errCode, errMsg := M.CreateDashboardUnit(project.ID, agent.UUID, &M.DashboardUnit{DashboardId: dashboard.ID,
			Title: rName1, Presentation: M.PresentationBar, Query: postgres.Jsonb{json.RawMessage(`{}`)}}, M.DashboardUnitForNoQueryID)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboardUnit1)
		assert.Empty(t, errMsg)
		assert.NotNil(t, dashboardUnit1.QueryId)
		assert.NotNil(t, dashboardUnit1.Settings)
		assert.NotNil(t, dashboardUnit1.Presentation)

		rName2 := U.RandomString(5)
		dashboardUnit2, errCode, errMsg := M.CreateDashboardUnit(project.ID, agent.UUID, &M.DashboardUnit{DashboardId: dashboard.ID,
			Title: rName2, Presentation: M.PresentationCard, Query: postgres.Jsonb{json.RawMessage(`{}`)}}, M.DashboardUnitForNoQueryID)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboardUnit2)
		assert.Empty(t, errMsg)
		assert.NotNil(t, dashboardUnit2.QueryId)
		assert.NotNil(t, dashboardUnit2.Settings)
		assert.NotNil(t, dashboardUnit2.Presentation)

		// should be given a positions on dashboard.
		gDashboard, errCode := M.GetDashboard(project.ID, agent.UUID, dashboard.ID)
		assert.Equal(t, http.StatusFound, errCode)
		var currentPosition map[string]map[uint64]int
		err := json.Unmarshal((*gDashboard.UnitsPosition).RawMessage, &currentPosition)
		assert.Nil(t, err)
		assert.Contains(t, currentPosition[M.GetUnitType(M.PresentationLine)], dashboardUnit.ID)
		assert.Contains(t, currentPosition[M.GetUnitType(M.PresentationBar)], dashboardUnit1.ID)
		assert.NotNil(t, currentPosition[M.GetUnitType(M.PresentationLine)][dashboardUnit.ID])
		assert.NotNil(t, currentPosition[M.GetUnitType(M.PresentationLine)][dashboardUnit1.ID])
		assert.Contains(t, currentPosition[M.GetUnitType(M.PresentationCard)], dashboardUnit2.ID)
		assert.NotNil(t, currentPosition[M.GetUnitType(M.PresentationCard)][dashboardUnit2.ID])

	})

	t.Run("CreateDashboardUnitWithSetting", func(t *testing.T) {
		rName := U.RandomString(5)
		dashboardUnit, errCode, errMsg := M.CreateDashboardUnit(project.ID, agent.UUID, &M.DashboardUnit{DashboardId: dashboard.ID,
			Title: rName, Settings: postgres.Jsonb{json.RawMessage(`{"chart":"pc"}`)},
			Query: postgres.Jsonb{json.RawMessage(`{}`)}}, M.DashboardUnitForNoQueryID)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboardUnit)
		assert.Empty(t, errMsg)
		assert.NotNil(t, dashboardUnit.QueryId)
		assert.NotNil(t, dashboardUnit.Settings)
		assert.NotNil(t, dashboardUnit.Presentation)

		rName1 := U.RandomString(5)
		dashboardUnit1, errCode, errMsg := M.CreateDashboardUnit(project.ID, agent.UUID, &M.DashboardUnit{DashboardId: dashboard.ID,
			Title: rName1, Settings: postgres.Jsonb{json.RawMessage(`{"chart":"pc"}`)}, Query: postgres.Jsonb{json.RawMessage(`{}`)}}, M.DashboardUnitForNoQueryID)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboardUnit1)
		assert.Empty(t, errMsg)
		assert.NotNil(t, dashboardUnit1.QueryId)
		assert.NotNil(t, dashboardUnit1.Settings)
		assert.NotNil(t, dashboardUnit1.Presentation)

		rName2 := U.RandomString(5)
		dashboardUnit2, errCode, errMsg := M.CreateDashboardUnit(project.ID, agent.UUID, &M.DashboardUnit{DashboardId: dashboard.ID,
			Title: rName2, Settings: postgres.Jsonb{json.RawMessage(`{"chart":"pc"}`)}, Query: postgres.Jsonb{json.RawMessage(`{}`)}}, M.DashboardUnitForNoQueryID)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboardUnit2)
		assert.Empty(t, errMsg)
		assert.NotNil(t, dashboardUnit2.QueryId)
		assert.NotNil(t, dashboardUnit2.Settings)
		assert.NotNil(t, dashboardUnit2.Presentation)

		// should be given a positions on dashboard.
		gDashboard, errCode := M.GetDashboard(project.ID, agent.UUID, dashboard.ID)
		assert.Equal(t, http.StatusFound, errCode)
		var currentPosition map[string]map[uint64]int
		err := json.Unmarshal((*gDashboard.UnitsPosition).RawMessage, &currentPosition)
		assert.Nil(t, err)
		assert.Contains(t, currentPosition[M.GetUnitType(M.PresentationCard)], dashboardUnit.ID)
		assert.Contains(t, currentPosition[M.GetUnitType(M.PresentationCard)], dashboardUnit1.ID)
		assert.NotNil(t, currentPosition[M.GetUnitType(M.PresentationCard)][dashboardUnit.ID])
		assert.NotNil(t, currentPosition[M.GetUnitType(M.PresentationCard)][dashboardUnit1.ID])
		assert.Contains(t, currentPosition[M.GetUnitType(M.PresentationCard)], dashboardUnit2.ID)
		assert.NotNil(t, currentPosition[M.GetUnitType(M.PresentationCard)][dashboardUnit2.ID])

	})

	t.Run("CreateDashboardUnitWithPresentation", func(t *testing.T) {
		rName := U.RandomString(5)
		dashboardUnit, errCode, errMsg := M.CreateDashboardUnit(project.ID, agent.UUID, &M.DashboardUnit{DashboardId: dashboard.ID,
			Title: rName, Presentation: M.PresentationCard,
			Query: postgres.Jsonb{json.RawMessage(`{}`)}}, M.DashboardUnitForNoQueryID)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboardUnit)
		assert.Empty(t, errMsg)
		assert.NotNil(t, dashboardUnit.QueryId)
		assert.NotNil(t, dashboardUnit.Settings)
		assert.NotNil(t, dashboardUnit.Presentation)

		rName1 := U.RandomString(5)
		dashboardUnit1, errCode, errMsg := M.CreateDashboardUnit(project.ID, agent.UUID, &M.DashboardUnit{DashboardId: dashboard.ID,
			Title: rName1, Presentation: M.PresentationCard, Query: postgres.Jsonb{json.RawMessage(`{}`)}}, M.DashboardUnitForNoQueryID)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboardUnit1)
		assert.Empty(t, errMsg)
		assert.NotNil(t, dashboardUnit1.QueryId)
		assert.NotNil(t, dashboardUnit1.Settings)
		assert.NotNil(t, dashboardUnit1.Presentation)

		rName2 := U.RandomString(5)
		dashboardUnit2, errCode, errMsg := M.CreateDashboardUnit(project.ID, agent.UUID, &M.DashboardUnit{DashboardId: dashboard.ID,
			Title: rName2, Presentation: M.PresentationCard, Query: postgres.Jsonb{json.RawMessage(`{}`)}}, M.DashboardUnitForNoQueryID)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboardUnit2)
		assert.Empty(t, errMsg)
		assert.NotNil(t, dashboardUnit2.QueryId)
		assert.NotNil(t, dashboardUnit2.Settings)
		assert.NotNil(t, dashboardUnit2.Presentation)

		// should be given a positions on dashboard.
		gDashboard, errCode := M.GetDashboard(project.ID, agent.UUID, dashboard.ID)
		assert.Equal(t, http.StatusFound, errCode)
		var currentPosition map[string]map[uint64]int
		err := json.Unmarshal((*gDashboard.UnitsPosition).RawMessage, &currentPosition)
		assert.Nil(t, err)
		assert.Contains(t, currentPosition[M.GetUnitType(M.PresentationCard)], dashboardUnit.ID)
		assert.Contains(t, currentPosition[M.GetUnitType(M.PresentationCard)], dashboardUnit1.ID)
		assert.NotNil(t, currentPosition[M.GetUnitType(M.PresentationCard)][dashboardUnit.ID])
		assert.NotNil(t, currentPosition[M.GetUnitType(M.PresentationCard)][dashboardUnit1.ID])
		assert.Contains(t, currentPosition[M.GetUnitType(M.PresentationCard)], dashboardUnit2.ID)
		assert.NotNil(t, currentPosition[M.GetUnitType(M.PresentationCard)][dashboardUnit2.ID])

	})

	t.Run("CreateDashboardUnit:Invalid", func(t *testing.T) {
		// invalid title.
		dashboardUnit, errCode, _ := M.CreateDashboardUnit(project.ID, agent.UUID, &M.DashboardUnit{DashboardId: dashboard.ID,
			Title: "", Presentation: M.PresentationLine, Query: postgres.Jsonb{json.RawMessage(`{}`)}}, M.DashboardUnitForNoQueryID)
		assert.Equal(t, http.StatusBadRequest, errCode)
		assert.Nil(t, dashboardUnit)

		// invalid presentation.
		rName := U.RandomString(5)
		dashboardUnit, errCode, _ = M.CreateDashboardUnit(project.ID, agent.UUID, &M.DashboardUnit{DashboardId: dashboard.ID,
			Title: rName, Presentation: "", Query: postgres.Jsonb{json.RawMessage(`{}`)}}, M.DashboardUnitForNoQueryID)
		assert.Equal(t, http.StatusBadRequest, errCode)
		assert.Nil(t, dashboardUnit)

		// invalid dashboard.
		rName = U.RandomString(5)
		dashboardUnit, errCode, _ = M.CreateDashboardUnit(project.ID, agent.UUID, &M.DashboardUnit{DashboardId: 0,
			Title: rName, Presentation: M.PresentationLine, Query: postgres.Jsonb{json.RawMessage(`{}`)}}, M.DashboardUnitForNoQueryID)
		assert.Equal(t, http.StatusBadRequest, errCode)
		assert.Nil(t, dashboardUnit)

		// invalid project.
		rName = U.RandomString(5)
		dashboardUnit, errCode, _ = M.CreateDashboardUnit(0, agent.UUID, &M.DashboardUnit{DashboardId: dashboard.ID,
			Title: rName, Presentation: M.PresentationLine, Query: postgres.Jsonb{json.RawMessage(`{}`)}}, M.DashboardUnitForNoQueryID)
		assert.Equal(t, http.StatusBadRequest, errCode)
		assert.Nil(t, dashboardUnit)

		// invalid agent.
		rName = U.RandomString(5)
		dashboardUnit, errCode, _ = M.CreateDashboardUnit(project.ID, "", &M.DashboardUnit{DashboardId: dashboard.ID,
			Title: rName, Presentation: M.PresentationLine, Query: postgres.Jsonb{json.RawMessage(`{}`)}}, M.DashboardUnitForNoQueryID)
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
			Title: rTitle, Presentation: M.PresentationLine, Query: postgres.Jsonb{json.RawMessage(`{}`)}}, M.DashboardUnitForNoQueryID)
		assert.Equal(t, http.StatusForbidden, errCode)
		assert.Nil(t, dashboardUnit)
	})
}

func TestCreateMultipleDashboardUnits(t *testing.T) {
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

	type args struct {
		requestPayload []M.DashboardUnitRequestPayload
		projectId      uint64
		agentUUID      string
		dashboardId    uint64
	}

	requestPayload1 := []M.DashboardUnitRequestPayload{{Title: U.RandomString(10), Description: U.RandomString(20),
		Settings: &postgres.Jsonb{json.RawMessage(`{"chart":"pc"}`)}, QueryId: uint64(U.RandomIntInRange(50, 100))},
	}

	testArgs1 := args{requestPayload: requestPayload1,
		projectId:   project.ID,
		agentUUID:   agent.UUID,
		dashboardId: dashboard.ID}

	requestPayload2 := []M.DashboardUnitRequestPayload{{Title: U.RandomString(10), Description: U.RandomString(20),
		Settings: &postgres.Jsonb{json.RawMessage(`{"chart":"pc"}`)}, QueryId: uint64(U.RandomIntInRange(50, 100))},
		{Title: U.RandomString(10), Description: U.RandomString(20),
			Settings: &postgres.Jsonb{json.RawMessage(`{"chart":"pc"}`)}, QueryId: uint64(U.RandomIntInRange(50, 100))},
		{Title: U.RandomString(10), Description: U.RandomString(20),
			Settings: &postgres.Jsonb{json.RawMessage(`{"chart":"pc"}`)}, QueryId: uint64(U.RandomIntInRange(50, 100))},
	}

	testArgs2 := args{requestPayload: requestPayload2,
		projectId:   project.ID,
		agentUUID:   agent.UUID,
		dashboardId: dashboard.ID}

	tests := []struct {
		name   string
		args   args
		units  int
		status int
		error  string
	}{
		{name: "SingleDashboardUnitOnOneDashboard", args: testArgs1, units: 1, status: http.StatusCreated, error: ""},
		{name: "MultiDashboardUnitsOnOneDashboard", args: testArgs2, units: 3, status: http.StatusCreated, error: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dashboardUnits, responseStatus, errorMsg := M.CreateMultipleDashboardUnits(tt.args.requestPayload, tt.args.projectId, tt.args.agentUUID, tt.args.dashboardId)

			assert.NotNil(t, dashboardUnits)
			if !reflect.DeepEqual(len(dashboardUnits), tt.units) {
				t.Errorf("CreateMultipleDashboardUnits() got = %v, want %v", len(dashboardUnits), tt.units)
			}
			if responseStatus != tt.status {
				t.Errorf("CreateMultipleDashboardUnits() got1 = %v, want %v", responseStatus, tt.status)
			}
			if errorMsg != tt.error {
				t.Errorf("CreateMultipleDashboardUnits() got2 = %v, want %v", errorMsg, tt.error)
			}
		})
	}
}

func TestCreateDashboardUnitForMultipleDashboards(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	agent2, errCode := SetupAgentReturnDAO(getRandomEmail(), "+13425356")
	assert.Equal(t, http.StatusCreated, errCode)
	_, errCode = M.CreateProjectAgentMappingWithDependencies(&M.ProjectAgentMapping{
		ProjectID: project.ID, AgentUUID: agent2.UUID})
	assert.Equal(t, http.StatusCreated, errCode)

	rName := U.RandomString(5)
	dashboard1, errCode := M.CreateDashboard(project.ID, agent.UUID, &M.Dashboard{Name: rName, Type: M.DashboardTypeProjectVisible})
	assert.NotNil(t, dashboard1)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Equal(t, rName, dashboard1.Name)

	rName = U.RandomString(5)
	dashboard2, errCode := M.CreateDashboard(project.ID, agent.UUID, &M.Dashboard{Name: rName, Type: M.DashboardTypeProjectVisible})
	assert.NotNil(t, dashboard2)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Equal(t, rName, dashboard2.Name)

	rName = U.RandomString(5)
	dashboard3, errCode := M.CreateDashboard(project.ID, agent.UUID, &M.Dashboard{Name: rName, Type: M.DashboardTypeProjectVisible})
	assert.NotNil(t, dashboard3)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Equal(t, rName, dashboard3.Name)

	type args struct {
		dashboardIds []uint64
		projectId    uint64
		agentUUID    string
		unitPayload  M.DashboardUnitRequestPayload
	}
	requestPayload := M.DashboardUnitRequestPayload{Title: U.RandomString(10), Description: U.RandomString(20),
		Settings: &postgres.Jsonb{json.RawMessage(`{"chart":"pc"}`)}, QueryId: uint64(U.RandomIntInRange(50, 100))}

	testArgs1 := args{
		dashboardIds: []uint64{dashboard1.ID},
		projectId:    project.ID,
		agentUUID:    agent.UUID,
		unitPayload:  requestPayload}

	testArgs2 := args{
		dashboardIds: []uint64{dashboard1.ID, dashboard2.ID, dashboard3.ID},
		projectId:    project.ID,
		agentUUID:    agent.UUID,
		unitPayload:  requestPayload}

	tests := []struct {
		name   string
		args   args
		units  int
		status int
		error  string
	}{
		{name: "SingleUnitOnOneDashboards", args: testArgs1, units: 1, status: http.StatusCreated, error: ""},
		{name: "SingleUnitOnThreeDashboards", args: testArgs2, units: 3, status: http.StatusCreated, error: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dashboardUnits, got1, got2 := M.CreateDashboardUnitForMultipleDashboards(tt.args.dashboardIds, tt.args.projectId, tt.args.agentUUID, tt.args.unitPayload)

			assert.NotNil(t, dashboardUnits)
			if !reflect.DeepEqual(len(dashboardUnits), tt.units) {
				t.Errorf("CreateMultipleDashboardUnits() got = %v, want %v", len(dashboardUnits), tt.units)
			}
			if got1 != tt.status {
				t.Errorf("CreateDashboardUnitForMultipleDashboards() got1 = %v, want %v", got1, tt.status)
			}
			if got2 != tt.error {
				t.Errorf("CreateDashboardUnitForMultipleDashboards() got2 = %v, want %v", got2, tt.error)
			}
		})
	}
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
			Title: rName, Presentation: M.PresentationLine, Query: postgres.Jsonb{json.RawMessage(`{}`)}}, M.DashboardUnitForNoQueryID)
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
			Title: rName, Presentation: M.PresentationLine, Query: postgres.Jsonb{json.RawMessage(`{}`)}}, M.DashboardUnitForNoQueryID)
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
			Title: rTitle, Presentation: M.PresentationLine, Query: postgres.Jsonb{json.RawMessage(`{}`)}}, M.DashboardUnitForNoQueryID)
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
			Title: rName, Presentation: M.PresentationLine, Query: postgres.Jsonb{json.RawMessage(`{}`)}}, M.DashboardUnitForNoQueryID)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, unit)

		unit1, errCode, _ := M.CreateDashboardUnit(project.ID, agent.UUID, &M.DashboardUnit{DashboardId: dashboard.ID,
			Title: rName, Presentation: M.PresentationBar, Query: postgres.Jsonb{json.RawMessage(`{}`)}}, M.DashboardUnitForNoQueryID)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, unit1)

		unit2, errCode, _ := M.CreateDashboardUnit(project.ID, agent.UUID, &M.DashboardUnit{DashboardId: dashboard.ID,
			Title: rName, Presentation: M.PresentationCard, Query: postgres.Jsonb{json.RawMessage(`{}`)}}, M.DashboardUnitForNoQueryID)
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
			Title: rName, Presentation: M.PresentationLine, Query: postgres.Jsonb{json.RawMessage(`{}`)}}, M.DashboardUnitForNoQueryID)
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
			Title: rTitle, Presentation: M.PresentationLine, Query: postgres.Jsonb{json.RawMessage(`{}`)}}, M.DashboardUnitForNoQueryID)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboardUnit)

		errCode = M.DeleteDashboardUnit(project.ID, agent2.UUID, dashboard.ID, dashboardUnit.ID)
		assert.Equal(t, http.StatusForbidden, errCode)
	})
}

func TestDeleteDashboardUnitWithQuery(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	dashboardQuery, errCode, errMsg := M.CreateQuery(project.ID, &M.Queries{
		ProjectID: project.ID,
		Type:      M.QueryTypeDashboardQuery,
	})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Empty(t, errMsg)
	assert.NotNil(t, dashboardQuery)

	savedQuery, errCode, errMsg := M.CreateQuery(project.ID, &M.Queries{
		ProjectID: project.ID,
		Type:      M.QueryTypeSavedQuery,
		CreatedBy: agent.UUID,
		Title:     U.RandomString(5),
	})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Empty(t, errMsg)
	assert.NotNil(t, savedQuery)

	dashboard, errCode := M.CreateDashboard(project.ID, agent.UUID,
		&M.Dashboard{Name: U.RandomString(5), Type: M.DashboardTypeProjectVisible})
	assert.NotNil(t, dashboard)
	assert.Equal(t, http.StatusCreated, errCode)

	// Dashboard unit with QueryTypeDashboardQuery.
	dashboardUnit, errCode, errMsg := M.CreateDashboardUnit(project.ID, agent.UUID,
		&M.DashboardUnit{DashboardId: dashboard.ID, Title: U.RandomString(5), Presentation: M.PresentationLine,
			QueryId: dashboardQuery.ID, Query: postgres.Jsonb{json.RawMessage(`{}`)}},
		M.DashboardUnitWithQueryID)
	assert.NotEmpty(t, dashboardUnit)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Empty(t, errMsg)

	// Deleting dashboard unit should delete the query of type QueryTypeDashboardQuery.
	unitID := dashboardUnit.ID
	errCode = M.DeleteDashboardUnit(project.ID, agent.UUID, dashboard.ID, dashboardUnit.ID)
	assert.Equal(t, http.StatusAccepted, errCode)
	dashboardUnit, errCode = M.GetDashboardUnitByUnitID(project.ID, unitID)
	assert.Empty(t, dashboardUnit)
	assert.Equal(t, http.StatusNotFound, errCode)
	query, errCode := M.GetQueryWithQueryId(project.ID, dashboardQuery.ID)
	assert.Empty(t, query)
	assert.Equal(t, http.StatusNotFound, errCode)

	// Dashboard unit with QueryTypeSavedQuery.
	dashboardUnit, errCode, errMsg = M.CreateDashboardUnit(project.ID, agent.UUID,
		&M.DashboardUnit{DashboardId: dashboard.ID, Title: U.RandomString(5), Presentation: M.PresentationLine,
			QueryId: savedQuery.ID, Query: postgres.Jsonb{json.RawMessage(`{}`)}},
		M.DashboardUnitWithQueryID)
	assert.NotEmpty(t, dashboardUnit)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Empty(t, errMsg)

	// Deleting dashboard unit should not delete the query of type QueryTypeSavedQuery.
	unitID = dashboardUnit.ID
	errCode = M.DeleteDashboardUnit(project.ID, agent.UUID, dashboard.ID, dashboardUnit.ID)
	assert.Equal(t, http.StatusAccepted, errCode)
	dashboardUnit, errCode = M.GetDashboardUnitByUnitID(project.ID, unitID)
	assert.Empty(t, dashboardUnit)
	assert.Equal(t, http.StatusNotFound, errCode)
	query, errCode = M.GetQueryWithQueryId(project.ID, savedQuery.ID)
	assert.NotEmpty(t, query)
	assert.Equal(t, http.StatusFound, errCode)
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
			Title: unitName1, Presentation: M.PresentationLine, Query: postgres.Jsonb{json.RawMessage(`{}`)}}, M.DashboardUnitForNoQueryID)
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
			Title: rTitle, Presentation: M.PresentationLine, Query: postgres.Jsonb{json.RawMessage(`{}`)}}, M.DashboardUnitForNoQueryID)
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
			baseQuery = &M.AttributionQueryUnit{Class: queryClass, Query: &M.AttributionQuery{From: from, To: to}}
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

	_, errCode := M.CreateOrGetUserCreatedEventName(&M.EventName{ProjectId: project.ID, Name: "$session"})
	assert.Equal(t, http.StatusCreated, errCode)

	customerAccountID := U.RandomLowerAphaNumString(5)
	_, errCode = M.UpdateProjectSettings(project.ID, &M.ProjectSetting{
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
		M.QueryClassAttribution: `{"cl": "attribution", "meta": {"metrics_breakdown": true}, "query": {"ce": {"na": "$session", "pr": []}, "cm": ["Impressions", "Clicks", "Spend"], "to": 1596479399, "lbw": 1, "lfe": [], "from": 1595874600, "attribution_key": "Campaign", "attribution_methodology": "Last_Touch"}}`,
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
		}, M.DashboardUnitForNoQueryID)
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

func TestCacheDashboardUnitsForProjectIDEventsGroupQuery(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	_, errCode := M.CreateOrGetUserCreatedEventName(&M.EventName{ProjectId: project.ID, Name: "$session"})
	assert.Equal(t, http.StatusCreated, errCode)

	customerAccountID := U.RandomLowerAphaNumString(5)
	_, errCode = M.UpdateProjectSettings(project.ID, &M.ProjectSetting{
		IntAdwordsCustomerAccountId: &customerAccountID,
	})

	dashboardName := U.RandomString(5)
	dashboard, errCode := M.CreateDashboard(project.ID, agent.UUID, &M.Dashboard{Name: dashboardName, Type: M.DashboardTypeProjectVisible})
	assert.NotNil(t, dashboard)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Equal(t, dashboardName, dashboard.Name)

	dashboardUnitQueriesMap := make(map[uint64]map[string]interface{})
	dashboardQueriesStr := []string{`{"query_group":[{"cl":"events","ty":"unique_users","ec":"each_given_event","fr":1583001000,"to":1585679399,"ewp":[{"na":"$session","pr":[{"en":"event","pr":"$source","op":"equals","va":"google","ty":"categorical","lop":"AND"},{"en":"user","pr":"$country","op":"equals","va":"India","ty":"categorical","lop":"AND"}]},{"na":"MagazineViews","pr":[{"en":"event","pr":"$source","op":"equals","va":"google","ty":"categorical","lop":"AND"},{"en":"user","pr":"$country","op":"equals","va":"India","ty":"categorical","lop":"AND"}]}],"gbp":[],"gbt":"date","tz":"Asia/Calcutta"}]}`,
		`{"query_group":[{"cl":"events","ty":"unique_users","ec":"each_given_event","fr":1583001000,"to":1585679399,"ewp":[{"na":"$session","pr":[{"en":"event","pr":"$source","op":"equals","va":"google","ty":"categorical","lop":"AND"},{"en":"user","pr":"$country","op":"equals","va":"India","ty":"categorical","lop":"AND"}]},{"na":"MagazineViews","pr":[{"en":"event","pr":"$source","op":"equals","va":"google","ty":"categorical","lop":"AND"},{"en":"user","pr":"$country","op":"equals","va":"India","ty":"categorical","lop":"AND"}]}],"gbp":[{"pr":"$browser","en":"event","pty":"categorical","ena":"$session","eni":1},{"pr":"$campaign","en":"event","pty":"categorical","ena":"MagazineViews","eni":2}],"gbt":"","tz":"Asia/Calcutta"}]}`,
		`{"query_group":[{"cl":"events","ty":"unique_users","ec":"each_given_event","fr":1583001000,"to":1585679399,"ewp":[{"na":"$session","pr":[{"en":"event","pr":"$source","op":"equals","va":"google","ty":"categorical","lop":"AND"},{"en":"user","pr":"$country","op":"equals","va":"India","ty":"categorical","lop":"AND"}]},{"na":"MagazineViews","pr":[{"en":"event","pr":"$source","op":"equals","va":"google","ty":"categorical","lop":"AND"},{"en":"user","pr":"$country","op":"equals","va":"India","ty":"categorical","lop":"AND"}]}],"gbp":[{"pr":"$browser","en":"event","pty":"categorical","ena":"$session","eni":1},{"pr":"$campaign","en":"event","pty":"categorical","ena":"MagazineViews","eni":2},{"pr":"$city","en":"user","pty":"categorical","ena":"$session","eni":1},{"pr":"$city","en":"user","pty":"categorical","ena":"MagazineViews","eni":2},{"pr":"$city","en":"user","pty":"categorical","ena":"$present"}],"gbt":"date","tz":"Asia/Calcutta"}]}`,
	}
	queryClass := M.QueryClassEvents
	for _, queryString := range dashboardQueriesStr {
		queryJSON := postgres.Jsonb{json.RawMessage(queryString)}
		baseQuery, err := M.DecodeQueryForClass(queryJSON, queryClass)
		assert.Nil(t, err)

		dashboardUnit, errCode, _ := M.CreateDashboardUnit(project.ID, agent.UUID, &M.DashboardUnit{
			DashboardId:  dashboard.ID,
			Title:        U.RandomString(5),
			Query:        queryJSON,
			Presentation: M.PresentationCard,
		}, M.DashboardUnitForNoQueryID)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboardUnit)

		dashboardUnitQueriesMap[dashboardUnit.ID] = make(map[string]interface{})
		dashboardUnitQueriesMap[dashboardUnit.ID]["class"] = queryClass
		dashboardUnitQueriesMap[dashboardUnit.ID]["queries"] = baseQuery
	}

	updatedUnitsCount := M.CacheDashboardUnitsForProjectID(project.ID, 1)
	assert.Equal(t, len(dashboardQueriesStr), updatedUnitsCount)

	for key, rangeFunction := range U.QueryDateRangePresets {
		if key == "TODAY" {
			fmt.Println("RUNNING FOR TODAY")
		}
		from, to := rangeFunction()
		for unitID, queryMap := range dashboardUnitQueriesMap {
			queryClass := queryMap["class"].(string)
			queries := queryMap["queries"].(M.BaseQuery)
			queries.SetQueryDateRange(from, to)
			// Refresh is sent as false. Must return all presets range from cache.
			w := sendAnalyticsQueryReq(r, queryClass, project.ID, agent, dashboard.ID, unitID, queries, false)
			assert.NotNil(t, w)
			assert.Equal(t, http.StatusOK, w.Code)

			var result map[string]interface{}
			json.Unmarshal([]byte(w.Body.String()), &result)
			// Cache must be true in response.
			assert.True(t, result["cache"].(bool))

			// Refresh is sent as true. Still must return from cache for all presets except for todays.
			w = sendAnalyticsQueryReq(r, queryClass, project.ID, agent, dashboard.ID, unitID, queries, true)
			assert.NotNil(t, w)
			assert.Equal(t, http.StatusOK, w.Code)
			result = nil
			json.Unmarshal([]byte(w.Body.String()), &result)

			if from == U.GetBeginningOfDayTimestampZ(U.TimeNowUnix(), U.TimeZoneStringIST) {
				// Today's preset. Must not be from cache.
				assert.False(t, result["cache"].(bool))

				// If queried again with refresh as false, should return from cache.
				w := sendAnalyticsQueryReq(r, queryClass, project.ID, agent, dashboard.ID, unitID, queries, false)
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

func TestCacheDashboardUnitsForProjectIDChannelsGroupQuery(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	// _, errCode := M.CreateOrGetUserCreatedEventName(&M.EventName{ProjectId: project.ID, Name: "$session"})
	// assert.Equal(t, http.StatusCreated, errCode)

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

	dashboardQueriesStr := []string{
		`{ "query_group":[{ "channel": "google_ads", "select_metrics": ["impressions"], "filters": [], "group_by": [], "gbt": "hour", "fr": 1585679400, "to": 1585765800 }], "cl": "channel_v1" }`,
	}
	queryClass := M.QueryClassChannelV1
	for _, queryString := range dashboardQueriesStr {
		queryJSON := postgres.Jsonb{json.RawMessage(queryString)}
		baseQuery, err := M.DecodeQueryForClass(queryJSON, queryClass)
		assert.Nil(t, err)

		dashboardUnit, errCode, _ := M.CreateDashboardUnit(project.ID, agent.UUID, &M.DashboardUnit{
			DashboardId:  dashboard.ID,
			Title:        U.RandomString(5),
			Query:        queryJSON,
			Presentation: M.PresentationCard,
		}, M.DashboardUnitForNoQueryID)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboardUnit)

		dashboardUnitQueriesMap[dashboardUnit.ID] = make(map[string]interface{})
		dashboardUnitQueriesMap[dashboardUnit.ID]["class"] = queryClass
		dashboardUnitQueriesMap[dashboardUnit.ID]["queries"] = baseQuery
	}

	updatedUnitsCount := M.CacheDashboardUnitsForProjectID(project.ID, 1)
	assert.Equal(t, len(dashboardQueriesStr), updatedUnitsCount)

	for key, rangeFunction := range U.QueryDateRangePresets {
		if key == "TODAY" {
			fmt.Println("RUNNING FOR TODAY")
		}
		from, to := rangeFunction()
		for unitID, queryMap := range dashboardUnitQueriesMap {
			queryClass := queryMap["class"].(string)
			queries := queryMap["queries"].(M.BaseQuery)
			queries.SetQueryDateRange(from, to)
			// Refresh is sent as false. Must return all presets range from cache.
			w := sendAnalyticsQueryReq(r, queryClass, project.ID, agent, dashboard.ID, unitID, queries, false)
			assert.NotNil(t, w)
			assert.Equal(t, http.StatusOK, w.Code)

			var result map[string]interface{}
			json.Unmarshal([]byte(w.Body.String()), &result)
			// Cache must be true in response.
			assert.True(t, result["cache"].(bool))

			// Refresh is sent as true. Still must return from cache for all presets except for todays.
			w = sendAnalyticsQueryReq(r, queryClass, project.ID, agent, dashboard.ID, unitID, queries, true)
			assert.NotNil(t, w)
			assert.Equal(t, http.StatusOK, w.Code)
			result = nil
			json.Unmarshal([]byte(w.Body.String()), &result)

			if from == U.GetBeginningOfDayTimestampZ(U.TimeNowUnix(), U.TimeZoneStringIST) {
				// Today's preset. Must not be from cache.
				assert.False(t, result["cache"].(bool))

				// If queried again with refresh as false, should return from cache.
				w := sendAnalyticsQueryReq(r, queryClass, project.ID, agent, dashboard.ID, unitID, queries, false)
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
	} else if queryClass == M.QueryClassChannelV1 {
		queryURL = "v1/channels/query"
		query := baseQuery.(*M.ChannelGroupQueryV1)
		queryPayload = query
	} else if queryClass == M.QueryClassEvents {
		queryURL = "v1/query"
		query := baseQuery.(*M.QueryGroup)
		queryPayload = query
	} else {
		queryURL = "attribution/query"
		query := baseQuery.(*M.AttributionQueryUnit)
		queryPayload = H.AttributionRequestPayload{
			Query: query.Query,
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
