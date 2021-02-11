package tests

import (
	"encoding/json"
	C "factors/config"
	H "factors/handler"
	"factors/handler/helpers"
	"factors/model/model"
	"factors/model/store"
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
	_, errCode = store.GetStore().CreateProjectAgentMappingWithDependencies(&model.ProjectAgentMapping{
		ProjectID: project.ID, AgentUUID: agent2.UUID})
	assert.Equal(t, http.StatusCreated, errCode)

	rName := U.RandomString(5)
	dashboard, errCode := store.GetStore().CreateDashboard(project.ID, agent.UUID, &model.Dashboard{Name: rName, Type: model.DashboardTypeProjectVisible})
	assert.NotNil(t, dashboard)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Equal(t, rName, dashboard.Name)

	t.Run("CreateDashboardUnit", func(t *testing.T) {
		rName := U.RandomString(5)
		dashboardUnit, errCode, errMsg := store.GetStore().CreateDashboardUnit(project.ID, agent.UUID, &model.DashboardUnit{DashboardId: dashboard.ID,
			Title: rName, Presentation: model.PresentationLine,
			Query: postgres.Jsonb{json.RawMessage(`{}`)}}, model.DashboardUnitForNoQueryID)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboardUnit)
		assert.Empty(t, errMsg)
		assert.NotNil(t, dashboardUnit.QueryId)
		assert.NotNil(t, dashboardUnit.Settings)
		assert.NotNil(t, dashboardUnit.Presentation)

		rName1 := U.RandomString(5)
		dashboardUnit1, errCode, errMsg := store.GetStore().CreateDashboardUnit(project.ID, agent.UUID, &model.DashboardUnit{DashboardId: dashboard.ID,
			Title: rName1, Presentation: model.PresentationBar, Query: postgres.Jsonb{json.RawMessage(`{}`)}}, model.DashboardUnitForNoQueryID)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboardUnit1)
		assert.Empty(t, errMsg)
		assert.NotNil(t, dashboardUnit1.QueryId)
		assert.NotNil(t, dashboardUnit1.Settings)
		assert.NotNil(t, dashboardUnit1.Presentation)

		rName2 := U.RandomString(5)
		dashboardUnit2, errCode, errMsg := store.GetStore().CreateDashboardUnit(project.ID, agent.UUID, &model.DashboardUnit{DashboardId: dashboard.ID,
			Title: rName2, Presentation: model.PresentationCard, Query: postgres.Jsonb{json.RawMessage(`{}`)}}, model.DashboardUnitForNoQueryID)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboardUnit2)
		assert.Empty(t, errMsg)
		assert.NotNil(t, dashboardUnit2.QueryId)
		assert.NotNil(t, dashboardUnit2.Settings)
		assert.NotNil(t, dashboardUnit2.Presentation)

		// should be given a positions on dashboard.
		gDashboard, errCode := store.GetStore().GetDashboard(project.ID, agent.UUID, dashboard.ID)
		assert.Equal(t, http.StatusFound, errCode)
		var currentPosition map[string]map[uint64]int
		err := json.Unmarshal((*gDashboard.UnitsPosition).RawMessage, &currentPosition)
		assert.Nil(t, err)
		assert.Contains(t, currentPosition[model.GetUnitType(model.PresentationLine)], dashboardUnit.ID)
		assert.Contains(t, currentPosition[model.GetUnitType(model.PresentationBar)], dashboardUnit1.ID)
		assert.NotNil(t, currentPosition[model.GetUnitType(model.PresentationLine)][dashboardUnit.ID])
		assert.NotNil(t, currentPosition[model.GetUnitType(model.PresentationLine)][dashboardUnit1.ID])
		assert.Contains(t, currentPosition[model.GetUnitType(model.PresentationCard)], dashboardUnit2.ID)
		assert.NotNil(t, currentPosition[model.GetUnitType(model.PresentationCard)][dashboardUnit2.ID])

	})

	t.Run("CreateDashboardUnitWithSetting", func(t *testing.T) {
		rName := U.RandomString(5)
		dashboardUnit, errCode, errMsg := store.GetStore().CreateDashboardUnit(project.ID, agent.UUID, &model.DashboardUnit{DashboardId: dashboard.ID,
			Title: rName, Settings: postgres.Jsonb{json.RawMessage(`{"chart":"pc"}`)},
			Query: postgres.Jsonb{json.RawMessage(`{}`)}}, model.DashboardUnitForNoQueryID)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboardUnit)
		assert.Empty(t, errMsg)
		assert.NotNil(t, dashboardUnit.QueryId)
		assert.NotNil(t, dashboardUnit.Settings)
		assert.NotNil(t, dashboardUnit.Presentation)

		rName1 := U.RandomString(5)
		dashboardUnit1, errCode, errMsg := store.GetStore().CreateDashboardUnit(project.ID, agent.UUID, &model.DashboardUnit{DashboardId: dashboard.ID,
			Title: rName1, Settings: postgres.Jsonb{json.RawMessage(`{"chart":"pc"}`)}, Query: postgres.Jsonb{json.RawMessage(`{}`)}}, model.DashboardUnitForNoQueryID)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboardUnit1)
		assert.Empty(t, errMsg)
		assert.NotNil(t, dashboardUnit1.QueryId)
		assert.NotNil(t, dashboardUnit1.Settings)
		assert.NotNil(t, dashboardUnit1.Presentation)

		rName2 := U.RandomString(5)
		dashboardUnit2, errCode, errMsg := store.GetStore().CreateDashboardUnit(project.ID, agent.UUID, &model.DashboardUnit{DashboardId: dashboard.ID,
			Title: rName2, Settings: postgres.Jsonb{json.RawMessage(`{"chart":"pc"}`)}, Query: postgres.Jsonb{json.RawMessage(`{}`)}}, model.DashboardUnitForNoQueryID)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboardUnit2)
		assert.Empty(t, errMsg)
		assert.NotNil(t, dashboardUnit2.QueryId)
		assert.NotNil(t, dashboardUnit2.Settings)
		assert.NotNil(t, dashboardUnit2.Presentation)

		// should be given a positions on dashboard.
		gDashboard, errCode := store.GetStore().GetDashboard(project.ID, agent.UUID, dashboard.ID)
		assert.Equal(t, http.StatusFound, errCode)
		var currentPosition map[string]map[uint64]int
		err := json.Unmarshal((*gDashboard.UnitsPosition).RawMessage, &currentPosition)
		assert.Nil(t, err)
		assert.Contains(t, currentPosition[model.GetUnitType(model.PresentationCard)], dashboardUnit.ID)
		assert.Contains(t, currentPosition[model.GetUnitType(model.PresentationCard)], dashboardUnit1.ID)
		assert.NotNil(t, currentPosition[model.GetUnitType(model.PresentationCard)][dashboardUnit.ID])
		assert.NotNil(t, currentPosition[model.GetUnitType(model.PresentationCard)][dashboardUnit1.ID])
		assert.Contains(t, currentPosition[model.GetUnitType(model.PresentationCard)], dashboardUnit2.ID)
		assert.NotNil(t, currentPosition[model.GetUnitType(model.PresentationCard)][dashboardUnit2.ID])

	})

	t.Run("CreateDashboardUnitWithPresentation", func(t *testing.T) {
		rName := U.RandomString(5)
		dashboardUnit, errCode, errMsg := store.GetStore().CreateDashboardUnit(project.ID, agent.UUID, &model.DashboardUnit{DashboardId: dashboard.ID,
			Title: rName, Presentation: model.PresentationCard,
			Query: postgres.Jsonb{json.RawMessage(`{}`)}}, model.DashboardUnitForNoQueryID)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboardUnit)
		assert.Empty(t, errMsg)
		assert.NotNil(t, dashboardUnit.QueryId)
		assert.NotNil(t, dashboardUnit.Settings)
		assert.NotNil(t, dashboardUnit.Presentation)

		rName1 := U.RandomString(5)
		dashboardUnit1, errCode, errMsg := store.GetStore().CreateDashboardUnit(project.ID, agent.UUID, &model.DashboardUnit{DashboardId: dashboard.ID,
			Title: rName1, Presentation: model.PresentationCard, Query: postgres.Jsonb{json.RawMessage(`{}`)}}, model.DashboardUnitForNoQueryID)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboardUnit1)
		assert.Empty(t, errMsg)
		assert.NotNil(t, dashboardUnit1.QueryId)
		assert.NotNil(t, dashboardUnit1.Settings)
		assert.NotNil(t, dashboardUnit1.Presentation)

		rName2 := U.RandomString(5)
		dashboardUnit2, errCode, errMsg := store.GetStore().CreateDashboardUnit(project.ID, agent.UUID, &model.DashboardUnit{DashboardId: dashboard.ID,
			Title: rName2, Presentation: model.PresentationCard, Query: postgres.Jsonb{json.RawMessage(`{}`)}}, model.DashboardUnitForNoQueryID)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboardUnit2)
		assert.Empty(t, errMsg)
		assert.NotNil(t, dashboardUnit2.QueryId)
		assert.NotNil(t, dashboardUnit2.Settings)
		assert.NotNil(t, dashboardUnit2.Presentation)

		// should be given a positions on dashboard.
		gDashboard, errCode := store.GetStore().GetDashboard(project.ID, agent.UUID, dashboard.ID)
		assert.Equal(t, http.StatusFound, errCode)
		var currentPosition map[string]map[uint64]int
		err := json.Unmarshal((*gDashboard.UnitsPosition).RawMessage, &currentPosition)
		assert.Nil(t, err)
		assert.Contains(t, currentPosition[model.GetUnitType(model.PresentationCard)], dashboardUnit.ID)
		assert.Contains(t, currentPosition[model.GetUnitType(model.PresentationCard)], dashboardUnit1.ID)
		assert.NotNil(t, currentPosition[model.GetUnitType(model.PresentationCard)][dashboardUnit.ID])
		assert.NotNil(t, currentPosition[model.GetUnitType(model.PresentationCard)][dashboardUnit1.ID])
		assert.Contains(t, currentPosition[model.GetUnitType(model.PresentationCard)], dashboardUnit2.ID)
		assert.NotNil(t, currentPosition[model.GetUnitType(model.PresentationCard)][dashboardUnit2.ID])

	})

	t.Run("CreateDashboardUnit:Invalid", func(t *testing.T) {
		// invalid title.
		dashboardUnit, errCode, _ := store.GetStore().CreateDashboardUnit(project.ID, agent.UUID, &model.DashboardUnit{DashboardId: dashboard.ID,
			Title: "", Presentation: model.PresentationLine, Query: postgres.Jsonb{json.RawMessage(`{}`)}}, model.DashboardUnitForNoQueryID)
		assert.Equal(t, http.StatusBadRequest, errCode)
		assert.Nil(t, dashboardUnit)

		// invalid presentation.
		rName := U.RandomString(5)
		dashboardUnit, errCode, _ = store.GetStore().CreateDashboardUnit(project.ID, agent.UUID, &model.DashboardUnit{DashboardId: dashboard.ID,
			Title: rName, Presentation: "", Query: postgres.Jsonb{json.RawMessage(`{}`)}}, model.DashboardUnitForNoQueryID)
		assert.Equal(t, http.StatusBadRequest, errCode)
		assert.Nil(t, dashboardUnit)

		// invalid dashboard.
		rName = U.RandomString(5)
		dashboardUnit, errCode, _ = store.GetStore().CreateDashboardUnit(project.ID, agent.UUID, &model.DashboardUnit{DashboardId: 0,
			Title: rName, Presentation: model.PresentationLine, Query: postgres.Jsonb{json.RawMessage(`{}`)}}, model.DashboardUnitForNoQueryID)
		assert.Equal(t, http.StatusBadRequest, errCode)
		assert.Nil(t, dashboardUnit)

		// invalid project.
		rName = U.RandomString(5)
		dashboardUnit, errCode, _ = store.GetStore().CreateDashboardUnit(0, agent.UUID, &model.DashboardUnit{DashboardId: dashboard.ID,
			Title: rName, Presentation: model.PresentationLine, Query: postgres.Jsonb{json.RawMessage(`{}`)}}, model.DashboardUnitForNoQueryID)
		assert.Equal(t, http.StatusBadRequest, errCode)
		assert.Nil(t, dashboardUnit)

		// invalid agent.
		rName = U.RandomString(5)
		dashboardUnit, errCode, _ = store.GetStore().CreateDashboardUnit(project.ID, "", &model.DashboardUnit{DashboardId: dashboard.ID,
			Title: rName, Presentation: model.PresentationLine, Query: postgres.Jsonb{json.RawMessage(`{}`)}}, model.DashboardUnitForNoQueryID)
		assert.Equal(t, http.StatusBadRequest, errCode)
		assert.Nil(t, dashboardUnit)
	})

	t.Run("CreateDashboardUnit:WithoutAccessToDashboard", func(t *testing.T) {
		rName := U.RandomString(5)
		dashboard, errCode := store.GetStore().CreateDashboard(project.ID, agent.UUID, &model.Dashboard{Name: rName, Type: model.DashboardTypePrivate})
		assert.NotNil(t, dashboard)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.Equal(t, rName, dashboard.Name)

		rTitle := U.RandomString(5)
		dashboardUnit, errCode, _ := store.GetStore().CreateDashboardUnit(project.ID, agent2.UUID, &model.DashboardUnit{DashboardId: dashboard.ID,
			Title: rTitle, Presentation: model.PresentationLine, Query: postgres.Jsonb{json.RawMessage(`{}`)}}, model.DashboardUnitForNoQueryID)
		assert.Equal(t, http.StatusForbidden, errCode)
		assert.Nil(t, dashboardUnit)
	})
}

func TestCreateMultipleDashboardUnits(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	agent2, errCode := SetupAgentReturnDAO(getRandomEmail(), "+13425356")
	assert.Equal(t, http.StatusCreated, errCode)
	_, errCode = store.GetStore().CreateProjectAgentMappingWithDependencies(&model.ProjectAgentMapping{
		ProjectID: project.ID, AgentUUID: agent2.UUID})
	assert.Equal(t, http.StatusCreated, errCode)

	rName := U.RandomString(5)
	dashboard, errCode := store.GetStore().CreateDashboard(project.ID, agent.UUID, &model.Dashboard{Name: rName, Type: model.DashboardTypeProjectVisible})
	assert.NotNil(t, dashboard)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Equal(t, rName, dashboard.Name)

	type args struct {
		requestPayload []model.DashboardUnitRequestPayload
		projectId      uint64
		agentUUID      string
		dashboardId    uint64
	}

	requestPayload1 := []model.DashboardUnitRequestPayload{{Title: U.RandomString(10), Description: U.RandomString(20),
		Settings: &postgres.Jsonb{json.RawMessage(`{"chart":"pc"}`)}},
	}

	testArgs1 := args{requestPayload: requestPayload1,
		projectId:   project.ID,
		agentUUID:   agent.UUID,
		dashboardId: dashboard.ID}

	requestPayload2 := []model.DashboardUnitRequestPayload{{Title: U.RandomString(10), Description: U.RandomString(20),
		Settings: &postgres.Jsonb{json.RawMessage(`{"chart":"pc"}`)}},
		{Title: U.RandomString(10), Description: U.RandomString(20),
			Settings: &postgres.Jsonb{json.RawMessage(`{"chart":"pc"}`)}},
		{Title: U.RandomString(10), Description: U.RandomString(20),
			Settings: &postgres.Jsonb{json.RawMessage(`{"chart":"pc"}`)}},
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
			for i := range tt.args.requestPayload {
				query, _, _ := store.GetStore().CreateQuery(project.ID, &model.Queries{
					ProjectID: project.ID,
					Type:      model.QueryTypeDashboardQuery,
					Query:     postgres.Jsonb{json.RawMessage(`{}`)},
				})
				tt.args.requestPayload[i].QueryId = query.ID
			}
			dashboardUnits, responseStatus, errorMsg := store.GetStore().CreateMultipleDashboardUnits(tt.args.requestPayload, tt.args.projectId, tt.args.agentUUID, tt.args.dashboardId)

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
	_, errCode = store.GetStore().CreateProjectAgentMappingWithDependencies(&model.ProjectAgentMapping{
		ProjectID: project.ID, AgentUUID: agent2.UUID})
	assert.Equal(t, http.StatusCreated, errCode)

	rName := U.RandomString(5)
	dashboard1, errCode := store.GetStore().CreateDashboard(project.ID, agent.UUID, &model.Dashboard{Name: rName, Type: model.DashboardTypeProjectVisible})
	assert.NotNil(t, dashboard1)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Equal(t, rName, dashboard1.Name)

	rName = U.RandomString(5)
	dashboard2, errCode := store.GetStore().CreateDashboard(project.ID, agent.UUID, &model.Dashboard{Name: rName, Type: model.DashboardTypeProjectVisible})
	assert.NotNil(t, dashboard2)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Equal(t, rName, dashboard2.Name)

	rName = U.RandomString(5)
	dashboard3, errCode := store.GetStore().CreateDashboard(project.ID, agent.UUID, &model.Dashboard{Name: rName, Type: model.DashboardTypeProjectVisible})
	assert.NotNil(t, dashboard3)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Equal(t, rName, dashboard3.Name)

	type args struct {
		dashboardIds []uint64
		projectId    uint64
		agentUUID    string
		unitPayload  model.DashboardUnitRequestPayload
	}
	requestPayload := model.DashboardUnitRequestPayload{Title: U.RandomString(10), Description: U.RandomString(20),
		Settings: &postgres.Jsonb{json.RawMessage(`{"chart":"pc"}`)}}

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
			query, _, _ := store.GetStore().CreateQuery(project.ID, &model.Queries{
				ProjectID: project.ID,
				Type:      model.QueryTypeDashboardQuery,
				Query:     postgres.Jsonb{json.RawMessage(`{}`)},
			})
			tt.args.unitPayload.QueryId = query.ID
			dashboardUnits, got1, got2 := store.GetStore().CreateDashboardUnitForMultipleDashboards(tt.args.dashboardIds, tt.args.projectId, tt.args.agentUUID, tt.args.unitPayload)

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
	_, errCode = store.GetStore().CreateProjectAgentMappingWithDependencies(&model.ProjectAgentMapping{
		ProjectID: project.ID, AgentUUID: agent2.UUID})
	assert.Equal(t, http.StatusCreated, errCode)

	rName := U.RandomString(5)
	dashboard, errCode := store.GetStore().CreateDashboard(project.ID, agent.UUID, &model.Dashboard{Name: rName, Type: model.DashboardTypeProjectVisible})
	assert.NotNil(t, dashboard)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Equal(t, rName, dashboard.Name)

	t.Run("GetDashboardUnits:NotAvailable", func(t *testing.T) {
		units, errCode := store.GetStore().GetDashboardUnits(project.ID, agent.UUID, dashboard.ID)
		assert.Equal(t, http.StatusFound, errCode)
		assert.Len(t, units, 0)
	})

	t.Run("GetDashboardUnits:Available", func(t *testing.T) {
		rName := U.RandomString(5)
		dashboardUnit, errCode, errMsg := store.GetStore().CreateDashboardUnit(project.ID, agent.UUID, &model.DashboardUnit{DashboardId: dashboard.ID,
			Title: rName, Presentation: model.PresentationLine, Query: postgres.Jsonb{json.RawMessage(`{}`)}}, model.DashboardUnitForNoQueryID)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboardUnit)
		assert.Empty(t, errMsg)

		units, errCode := store.GetStore().GetDashboardUnits(project.ID, agent.UUID, dashboard.ID)
		assert.Equal(t, http.StatusFound, errCode)
		assert.Len(t, units, 1)
		assert.Equal(t, rName, units[0].Title)
		assert.Equal(t, model.PresentationLine, units[0].Presentation)
	})

	t.Run("GetDashboardUnits:Invalid", func(t *testing.T) {
		rName := U.RandomString(5)
		dashboardUnit, errCode, errMsg := store.GetStore().CreateDashboardUnit(project.ID, agent.UUID, &model.DashboardUnit{DashboardId: dashboard.ID,
			Title: rName, Presentation: model.PresentationLine, Query: postgres.Jsonb{json.RawMessage(`{}`)}}, model.DashboardUnitForNoQueryID)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboardUnit)
		assert.Empty(t, errMsg)

		// invalid project
		units, errCode := store.GetStore().GetDashboardUnits(0, agent.UUID, dashboard.ID)
		assert.Equal(t, http.StatusBadRequest, errCode)
		assert.Nil(t, units)

		// invalid agent
		units, errCode = store.GetStore().GetDashboardUnits(project.ID, "", dashboard.ID)
		assert.Equal(t, http.StatusBadRequest, errCode)
		assert.Nil(t, units)

		// invalid dashboard
		units, errCode = store.GetStore().GetDashboardUnits(project.ID, agent.UUID, 0)
		assert.Equal(t, http.StatusBadRequest, errCode)
		assert.Nil(t, units)
	})

	t.Run("GetDashboardUnits:WithoutAccessToDashboard", func(t *testing.T) {
		rName := U.RandomString(5)
		dashboard, errCode := store.GetStore().CreateDashboard(project.ID, agent.UUID, &model.Dashboard{Name: rName, Type: model.DashboardTypePrivate})
		assert.NotNil(t, dashboard)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.Equal(t, rName, dashboard.Name)

		rTitle := U.RandomString(5)
		dashboardUnit, errCode, _ := store.GetStore().CreateDashboardUnit(project.ID, agent.UUID, &model.DashboardUnit{DashboardId: dashboard.ID,
			Title: rTitle, Presentation: model.PresentationLine, Query: postgres.Jsonb{json.RawMessage(`{}`)}}, model.DashboardUnitForNoQueryID)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboardUnit)

		units, errCode := store.GetStore().GetDashboardUnits(project.ID, agent2.UUID, dashboard.ID)
		assert.Equal(t, http.StatusForbidden, errCode)
		assert.Nil(t, units)
	})
}

func TestDeleteDashboardUnit(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	agent2, errCode := SetupAgentReturnDAO(getRandomEmail(), "+13425356")
	assert.Equal(t, http.StatusCreated, errCode)
	_, errCode = store.GetStore().CreateProjectAgentMappingWithDependencies(&model.ProjectAgentMapping{
		ProjectID: project.ID, AgentUUID: agent2.UUID})
	assert.Equal(t, http.StatusCreated, errCode)

	rName := U.RandomString(5)
	dashboard, errCode := store.GetStore().CreateDashboard(project.ID, agent.UUID, &model.Dashboard{Name: rName, Type: model.DashboardTypeProjectVisible})
	assert.NotNil(t, dashboard)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Equal(t, rName, dashboard.Name)

	t.Run("DeleteDashboardUnit", func(t *testing.T) {
		rName := U.RandomString(5)
		unit, errCode, _ := store.GetStore().CreateDashboardUnit(project.ID, agent.UUID, &model.DashboardUnit{DashboardId: dashboard.ID,
			Title: rName, Presentation: model.PresentationLine, Query: postgres.Jsonb{json.RawMessage(`{}`)}}, model.DashboardUnitForNoQueryID)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, unit)

		unit1, errCode, _ := store.GetStore().CreateDashboardUnit(project.ID, agent.UUID, &model.DashboardUnit{DashboardId: dashboard.ID,
			Title: rName, Presentation: model.PresentationBar, Query: postgres.Jsonb{json.RawMessage(`{}`)}}, model.DashboardUnitForNoQueryID)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, unit1)

		unit2, errCode, _ := store.GetStore().CreateDashboardUnit(project.ID, agent.UUID, &model.DashboardUnit{DashboardId: dashboard.ID,
			Title: rName, Presentation: model.PresentationCard, Query: postgres.Jsonb{json.RawMessage(`{}`)}}, model.DashboardUnitForNoQueryID)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, unit2)

		errCode = store.GetStore().DeleteDashboardUnit(project.ID, agent.UUID, dashboard.ID, unit.ID)
		assert.Equal(t, http.StatusAccepted, errCode)

		// should remove position given for unit on dashboard and rebalanced positions.
		gDashboard, errCode := store.GetStore().GetDashboard(project.ID, agent.UUID, dashboard.ID)
		assert.Equal(t, http.StatusFound, errCode)
		var currentPosition map[string]map[uint64]int
		err := json.Unmarshal((*gDashboard.UnitsPosition).RawMessage, &currentPosition)
		assert.Nil(t, err)
		assert.NotContains(t, currentPosition[model.GetUnitType(model.PresentationLine)], unit.ID)
		assert.Contains(t, currentPosition[model.GetUnitType(model.PresentationBar)], unit1.ID)
		// unit1 should be repositioned to 0.
		assert.Equal(t, currentPosition[model.GetUnitType(model.PresentationBar)][unit1.ID], 0)
		// delete should not affect postions other unit type.
		assert.Contains(t, currentPosition, model.GetUnitType(model.PresentationCard))
		assert.Equal(t, currentPosition[model.GetUnitType(model.PresentationCard)][unit2.ID], 0)

		// Unit must have got soft deleted.
		var deletedUnit model.DashboardUnit
		db := C.GetServices().Db
		err = db.Model(model.DashboardUnit{}).Where("project_id = ? AND id = ?", unit.ProjectID, unit.ID).Find(&deletedUnit).Error
		assert.Nil(t, err)
		assert.Equal(t, true, deletedUnit.IsDeleted)
	})

	t.Run("DeleteDashboardUnit:Invalid", func(t *testing.T) {
		rName := U.RandomString(5)
		unit, errCode, _ := store.GetStore().CreateDashboardUnit(project.ID, agent.UUID, &model.DashboardUnit{DashboardId: dashboard.ID,
			Title: rName, Presentation: model.PresentationLine, Query: postgres.Jsonb{json.RawMessage(`{}`)}}, model.DashboardUnitForNoQueryID)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, unit)

		// invalid project.
		errCode = store.GetStore().DeleteDashboardUnit(0, agent.UUID, dashboard.ID, unit.ID)
		assert.Equal(t, http.StatusBadRequest, errCode)

		// invalid dashboard.
		errCode = store.GetStore().DeleteDashboardUnit(project.ID, agent.UUID, 0, unit.ID)
		assert.Equal(t, http.StatusBadRequest, errCode)

		// invalid unit.
		errCode = store.GetStore().DeleteDashboardUnit(project.ID, agent.UUID, dashboard.ID, 0)
		assert.Equal(t, http.StatusBadRequest, errCode)
	})

	t.Run("DeleteDashboardUnit:WithoutAccessToDashboard", func(t *testing.T) {
		rName := U.RandomString(5)
		dashboard, errCode := store.GetStore().CreateDashboard(project.ID, agent.UUID, &model.Dashboard{Name: rName, Type: model.DashboardTypePrivate})
		assert.NotNil(t, dashboard)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.Equal(t, rName, dashboard.Name)

		rTitle := U.RandomString(5)
		dashboardUnit, errCode, _ := store.GetStore().CreateDashboardUnit(project.ID, agent.UUID, &model.DashboardUnit{DashboardId: dashboard.ID,
			Title: rTitle, Presentation: model.PresentationLine, Query: postgres.Jsonb{json.RawMessage(`{}`)}}, model.DashboardUnitForNoQueryID)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboardUnit)

		errCode = store.GetStore().DeleteDashboardUnit(project.ID, agent2.UUID, dashboard.ID, dashboardUnit.ID)
		assert.Equal(t, http.StatusForbidden, errCode)
	})
}

func TestDeleteDashboardUnitWithQuery(t *testing.T) {
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

	savedQuery, errCode, errMsg := store.GetStore().CreateQuery(project.ID, &model.Queries{
		ProjectID: project.ID,
		Type:      model.QueryTypeSavedQuery,
		CreatedBy: agent.UUID,
		Title:     U.RandomString(5),
		Query:     postgres.Jsonb{json.RawMessage(`{}`)},
	})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Empty(t, errMsg)
	assert.NotNil(t, savedQuery)

	dashboard, errCode := store.GetStore().CreateDashboard(project.ID, agent.UUID,
		&model.Dashboard{Name: U.RandomString(5), Type: model.DashboardTypeProjectVisible})
	assert.NotNil(t, dashboard)
	assert.Equal(t, http.StatusCreated, errCode)

	// Dashboard unit with QueryTypeDashboardQuery.
	dashboardUnit, errCode, errMsg := store.GetStore().CreateDashboardUnit(project.ID, agent.UUID,
		&model.DashboardUnit{DashboardId: dashboard.ID, Title: U.RandomString(5), Presentation: model.PresentationLine,
			QueryId: dashboardQuery.ID, Query: postgres.Jsonb{json.RawMessage(`{}`)}},
		model.DashboardUnitWithQueryID)
	assert.NotEmpty(t, dashboardUnit)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Empty(t, errMsg)

	// Deleting dashboard unit should delete the query of type QueryTypeDashboardQuery.
	unitID := dashboardUnit.ID
	errCode = store.GetStore().DeleteDashboardUnit(project.ID, agent.UUID, dashboard.ID, dashboardUnit.ID)
	assert.Equal(t, http.StatusAccepted, errCode)
	dashboardUnit, errCode = store.GetStore().GetDashboardUnitByUnitID(project.ID, unitID)
	assert.Empty(t, dashboardUnit)
	assert.Equal(t, http.StatusNotFound, errCode)
	query, errCode := store.GetStore().GetQueryWithQueryId(project.ID, dashboardQuery.ID)
	assert.Empty(t, query)
	assert.Equal(t, http.StatusNotFound, errCode)

	// Dashboard unit with QueryTypeSavedQuery.
	dashboardUnit, errCode, errMsg = store.GetStore().CreateDashboardUnit(project.ID, agent.UUID,
		&model.DashboardUnit{DashboardId: dashboard.ID, Title: U.RandomString(5), Presentation: model.PresentationLine,
			QueryId: savedQuery.ID, Query: postgres.Jsonb{json.RawMessage(`{}`)}},
		model.DashboardUnitWithQueryID)
	assert.NotEmpty(t, dashboardUnit)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Empty(t, errMsg)

	// Deleting dashboard unit should not delete the query of type QueryTypeSavedQuery.
	unitID = dashboardUnit.ID
	errCode = store.GetStore().DeleteDashboardUnit(project.ID, agent.UUID, dashboard.ID, dashboardUnit.ID)
	assert.Equal(t, http.StatusAccepted, errCode)
	dashboardUnit, errCode = store.GetStore().GetDashboardUnitByUnitID(project.ID, unitID)
	assert.Empty(t, dashboardUnit)
	assert.Equal(t, http.StatusNotFound, errCode)
	query, errCode = store.GetStore().GetQueryWithQueryId(project.ID, savedQuery.ID)
	assert.NotEmpty(t, query)
	assert.Equal(t, http.StatusFound, errCode)
}

func TestUpdateDashboardUnit(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	agent2, errCode := SetupAgentReturnDAO("", "")
	assert.Equal(t, http.StatusCreated, errCode)
	_, errCode = store.GetStore().CreateProjectAgentMappingWithDependencies(&model.ProjectAgentMapping{
		ProjectID: project.ID, AgentUUID: agent2.UUID})
	assert.Equal(t, http.StatusCreated, errCode)

	t.Run("UpdateDashboardUnit:Title", func(t *testing.T) {
		rName := U.RandomString(5)
		dashboard, errCode := store.GetStore().CreateDashboard(project.ID, agent.UUID, &model.Dashboard{Name: rName, Type: model.DashboardTypeProjectVisible})
		assert.NotNil(t, dashboard)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.Equal(t, rName, dashboard.Name)

		unitName1 := U.RandomString(5)
		unit, errCode, _ := store.GetStore().CreateDashboardUnit(project.ID, agent.UUID, &model.DashboardUnit{DashboardId: dashboard.ID,
			Title: unitName1, Presentation: model.PresentationLine, Query: postgres.Jsonb{json.RawMessage(`{}`)}}, model.DashboardUnitForNoQueryID)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, unit)

		unitName2 := U.RandomString(5)
		updatedDashboard, errCode := store.GetStore().UpdateDashboardUnit(project.ID, agent.UUID, dashboard.ID, unit.ID, &model.DashboardUnit{Title: unitName2})
		assert.Equal(t, http.StatusAccepted, errCode)
		assert.Equal(t, unitName2, updatedDashboard.Title)

		unitName33 := U.RandomString(5)
		description3 := "description3"
		presentation3 := "pr"
		settings3 := postgres.Jsonb{RawMessage: json.RawMessage(`{"Setting" : "Default"}`)}
		updatedDashboardUnit3, errCode := store.GetStore().UpdateDashboardUnit(project.ID, agent.UUID, dashboard.ID, unit.ID, &model.DashboardUnit{Title: unitName33,
			Description: description3, Presentation: presentation3, Settings: settings3})
		assert.Equal(t, http.StatusAccepted, errCode)
		assert.Equal(t, unitName33, updatedDashboardUnit3.Title)
		assert.Equal(t, description3, updatedDashboardUnit3.Description)
		assert.Equal(t, presentation3, updatedDashboardUnit3.Presentation)
		assert.Equal(t, settings3, updatedDashboardUnit3.Settings)

		unitName44 := U.RandomString(5)
		presentation4 := "pr"
		settings4 := postgres.Jsonb{RawMessage: json.RawMessage(`{"Setting" : "Default"}`)}
		updatedDashboardUnit4, errCode := store.GetStore().UpdateDashboardUnit(project.ID, agent.UUID, dashboard.ID, unit.ID, &model.DashboardUnit{Title: unitName44,
			Presentation: presentation4, Settings: settings4})
		assert.Equal(t, http.StatusAccepted, errCode)
		assert.Equal(t, unitName44, updatedDashboardUnit4.Title)
		assert.Equal(t, presentation4, updatedDashboardUnit4.Presentation)
		assert.Equal(t, settings4, updatedDashboardUnit4.Settings)

		// invalid title.
		updatedDashboard, errCode = store.GetStore().UpdateDashboardUnit(project.ID, agent.UUID, dashboard.ID, unit.ID, &model.DashboardUnit{Title: ""})
		assert.Equal(t, http.StatusBadRequest, errCode)

		// invalid projectId.
		unitName3 := U.RandomString(5)
		updatedDashboard, errCode = store.GetStore().UpdateDashboardUnit(0, agent.UUID, dashboard.ID, unit.ID, &model.DashboardUnit{Title: unitName3})
		assert.Equal(t, http.StatusBadRequest, errCode)

		// invalid agentUUID.
		updatedDashboard, errCode = store.GetStore().UpdateDashboardUnit(project.ID, "", dashboard.ID, unit.ID, &model.DashboardUnit{Title: unitName3})
		assert.Equal(t, http.StatusBadRequest, errCode)

		// invalid dashboardId.
		updatedDashboard, errCode = store.GetStore().UpdateDashboardUnit(project.ID, agent.UUID, 0, unit.ID, &model.DashboardUnit{Title: unitName3})
		assert.Equal(t, http.StatusBadRequest, errCode)

		// invalid unitId.
		updatedDashboard, errCode = store.GetStore().UpdateDashboardUnit(project.ID, agent.UUID, dashboard.ID, 0, &model.DashboardUnit{Title: unitName3})
		assert.Equal(t, http.StatusBadRequest, errCode)
	})

	t.Run("UpdateDashboardUnit:WithoutAccessToDashboard", func(t *testing.T) {
		rName := U.RandomString(5)
		dashboard, errCode := store.GetStore().CreateDashboard(project.ID, agent.UUID, &model.Dashboard{Name: rName, Type: model.DashboardTypePrivate})
		assert.NotNil(t, dashboard)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.Equal(t, rName, dashboard.Name)

		rTitle := U.RandomString(5)
		dashboardUnit, errCode, _ := store.GetStore().CreateDashboardUnit(project.ID, agent.UUID, &model.DashboardUnit{DashboardId: dashboard.ID,
			Title: rTitle, Presentation: model.PresentationLine, Query: postgres.Jsonb{json.RawMessage(`{}`)}}, model.DashboardUnitForNoQueryID)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboardUnit)

		rTitle2 := U.RandomString(5)
		_, errCode = store.GetStore().UpdateDashboardUnit(project.ID, agent2.UUID, dashboard.ID, dashboardUnit.ID, &model.DashboardUnit{Title: rTitle2})
		assert.Equal(t, http.StatusForbidden, errCode)
	})
}

func TestBaseQuery(t *testing.T) {
	var baseQuery model.BaseQuery
	for _, queryClass := range []string{model.QueryClassInsights, model.QueryClassFunnel, model.QueryClassChannel, model.QueryClassAttribution} {
		from, to := U.TimeNowUnix(), U.TimeNowUnix()+100
		if queryClass == model.QueryClassFunnel || queryClass == model.QueryClassInsights {
			baseQuery = &model.Query{Class: queryClass, From: from, To: to}
		} else if queryClass == model.QueryClassAttribution {
			baseQuery = &model.AttributionQueryUnit{Class: queryClass, Query: &model.AttributionQuery{From: from, To: to}}
		} else {
			baseQuery = &model.ChannelQueryUnit{Class: queryClass, Query: &model.ChannelQuery{From: from, To: to}}
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

	_, errCode := store.GetStore().CreateOrGetUserCreatedEventName(&model.EventName{ProjectId: project.ID, Name: "$session"})
	assert.Equal(t, http.StatusCreated, errCode)

	customerAccountID := U.RandomLowerAphaNumString(5)
	_, errCode = store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		IntAdwordsCustomerAccountId: &customerAccountID,
	})

	dashboardName := U.RandomString(5)
	dashboard, errCode := store.GetStore().CreateDashboard(project.ID, agent.UUID, &model.Dashboard{Name: dashboardName, Type: model.DashboardTypeProjectVisible})
	assert.NotNil(t, dashboard)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Equal(t, dashboardName, dashboard.Name)

	dashboardUnitQueriesMap := make(map[uint64]map[string]interface{})
	var dashboardQueriesStr = map[string]string{
		model.QueryClassInsights:    `{"cl": "insights", "ec": "any_given_event", "fr": 1393612200, "to": 1396290599, "ty": "events_occurrence", "tz": "", "ewp": [{"na": "$session", "pr": []}], "gbp": [], "gbt": ""}`,
		model.QueryClassFunnel:      `{"cl": "funnel", "ec": "any_given_event", "fr": 1594492200, "to": 1594578599, "ty": "unique_users", "tz": "Asia/Calcutta", "ewp": [{"na": "$session", "pr": []}, {"na": "www.chargebee.com/schedule-a-demo", "pr": []}], "gbp": [], "gbt": ""}`,
		model.QueryClassAttribution: `{"cl": "attribution", "meta": {"metrics_breakdown": true}, "query": {"ce": {"na": "$session", "pr": []}, "cm": ["Impressions", "Clicks", "Spend"], "to": 1596479399, "lbw": 1, "lfe": [], "from": 1595874600, "attribution_key": "Campaign", "attribution_methodology": "Last_Touch"}}`,
		model.QueryClassChannel:     `{"cl": "channel", "meta": {"metric": "total_cost"}, "query": {"to": 1576060774, "from": 1573468774, "channel": "google_ads", "filter_key": "campaign", "filter_value": "all"}}`,
	}
	for queryClass, queryString := range dashboardQueriesStr {
		queryJSON := postgres.Jsonb{json.RawMessage(queryString)}
		baseQuery, err := model.DecodeQueryForClass(queryJSON, queryClass)
		assert.Nil(t, err)

		dashboardUnit, errCode, _ := store.GetStore().CreateDashboardUnit(project.ID, agent.UUID, &model.DashboardUnit{
			DashboardId:  dashboard.ID,
			Title:        U.RandomString(5),
			Query:        queryJSON,
			Presentation: model.PresentationCard,
		}, model.DashboardUnitForNoQueryID)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboardUnit)

		dashboardUnitQueriesMap[dashboardUnit.ID] = make(map[string]interface{})
		dashboardUnitQueriesMap[dashboardUnit.ID]["class"] = queryClass
		dashboardUnitQueriesMap[dashboardUnit.ID]["query"] = baseQuery
	}

	updatedUnitsCount := store.GetStore().CacheDashboardUnitsForProjectID(project.ID, 1)
	assert.Equal(t, 4, updatedUnitsCount)

	for rangeString, rangeFunction := range U.QueryDateRangePresets {
		from, to := rangeFunction()
		for unitID, queryMap := range dashboardUnitQueriesMap {
			queryClass := queryMap["class"].(string)
			query := queryMap["query"].(model.BaseQuery)
			assertMsg := fmt.Sprintf("Failed for class:%s:range:%s", queryClass, rangeString)

			query.SetQueryDateRange(from, to)
			// Refresh is sent as false. Must return all presets range from cache.
			w := sendAnalyticsQueryReq(r, queryClass, project.ID, agent, dashboard.ID, unitID, query, false, true)
			assert.NotNil(t, w)
			assert.Equal(t, http.StatusOK, w.Code)

			var result map[string]interface{}
			json.Unmarshal([]byte(w.Body.String()), &result)
			// Cache must be true in response.
			assert.True(t, result["cache"].(bool))

			// Refresh is sent as true. Still must return from cache for all presets except for todays.
			w = sendAnalyticsQueryReq(r, queryClass, project.ID, agent, dashboard.ID, unitID, query, true, true)
			assert.NotNil(t, w)
			assert.Equal(t, http.StatusOK, w.Code)
			result = nil
			json.Unmarshal([]byte(w.Body.String()), &result)

			// Send same query as core query without sending dashboardID and unitID.
			// Since cached from dashboard caching, it should also be available with direct query.
			w = sendAnalyticsQueryReq(r, queryClass, project.ID, agent, 0, 0, query, false, false)
			assert.NotEmpty(t, w)
			assert.Equal(t, http.StatusOK, w.Code)
			if queryClass != model.QueryClassWeb {
				// For website analytics, it returns from Dashboard cache.
				assert.Equal(t, "true", w.HeaderMap.Get(model.QueryCacheResponseFromCacheHeader), assertMsg)
			}

			if from == U.GetBeginningOfDayTimestampZ(U.TimeNowUnix(), U.TimeZoneStringIST) {
				// If queried again with refresh as false, should return from cache.
				w := sendAnalyticsQueryReq(r, queryClass, project.ID, agent, dashboard.ID, unitID, query, false, true)
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

	_, errCode := store.GetStore().CreateOrGetUserCreatedEventName(&model.EventName{ProjectId: project.ID, Name: "$session"})
	assert.Equal(t, http.StatusCreated, errCode)

	customerAccountID := U.RandomLowerAphaNumString(5)
	_, errCode = store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		IntAdwordsCustomerAccountId: &customerAccountID,
	})

	dashboardName := U.RandomString(5)
	dashboard, errCode := store.GetStore().CreateDashboard(project.ID, agent.UUID, &model.Dashboard{Name: dashboardName, Type: model.DashboardTypeProjectVisible})
	assert.NotNil(t, dashboard)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Equal(t, dashboardName, dashboard.Name)

	dashboardUnitQueriesMap := make(map[uint64]map[string]interface{})
	dashboardQueriesStr := []string{`{"query_group":[{"cl":"events","ty":"unique_users","ec":"each_given_event","fr":1583001000,"to":1585679399,"ewp":[{"na":"$session","pr":[{"en":"event","pr":"$source","op":"equals","va":"google","ty":"categorical","lop":"AND"},{"en":"user","pr":"$country","op":"equals","va":"India","ty":"categorical","lop":"AND"}]},{"na":"MagazineViews","pr":[{"en":"event","pr":"$source","op":"equals","va":"google","ty":"categorical","lop":"AND"},{"en":"user","pr":"$country","op":"equals","va":"India","ty":"categorical","lop":"AND"}]}],"gbp":[],"gbt":"date","tz":"Asia/Calcutta"}]}`,
		`{"query_group":[{"cl":"events","ty":"unique_users","ec":"each_given_event","fr":1583001000,"to":1585679399,"ewp":[{"na":"$session","pr":[{"en":"event","pr":"$source","op":"equals","va":"google","ty":"categorical","lop":"AND"},{"en":"user","pr":"$country","op":"equals","va":"India","ty":"categorical","lop":"AND"}]},{"na":"MagazineViews","pr":[{"en":"event","pr":"$source","op":"equals","va":"google","ty":"categorical","lop":"AND"},{"en":"user","pr":"$country","op":"equals","va":"India","ty":"categorical","lop":"AND"}]}],"gbp":[{"pr":"$browser","en":"event","pty":"categorical","ena":"$session","eni":1},{"pr":"$campaign","en":"event","pty":"categorical","ena":"MagazineViews","eni":2}],"gbt":"","tz":"Asia/Calcutta"}]}`,
		`{"query_group":[{"cl":"events","ty":"unique_users","ec":"each_given_event","fr":1583001000,"to":1585679399,"ewp":[{"na":"$session","pr":[{"en":"event","pr":"$source","op":"equals","va":"google","ty":"categorical","lop":"AND"},{"en":"user","pr":"$country","op":"equals","va":"India","ty":"categorical","lop":"AND"}]},{"na":"MagazineViews","pr":[{"en":"event","pr":"$source","op":"equals","va":"google","ty":"categorical","lop":"AND"},{"en":"user","pr":"$country","op":"equals","va":"India","ty":"categorical","lop":"AND"}]}],"gbp":[{"pr":"$browser","en":"event","pty":"categorical","ena":"$session","eni":1},{"pr":"$campaign","en":"event","pty":"categorical","ena":"MagazineViews","eni":2},{"pr":"$city","en":"user","pty":"categorical","ena":"$session","eni":1},{"pr":"$city","en":"user","pty":"categorical","ena":"MagazineViews","eni":2},{"pr":"$city","en":"user","pty":"categorical","ena":"$present"}],"gbt":"date","tz":"Asia/Calcutta"}]}`,
	}
	queryClass := model.QueryClassEvents
	for _, queryString := range dashboardQueriesStr {
		queryJSON := postgres.Jsonb{json.RawMessage(queryString)}
		baseQuery, err := model.DecodeQueryForClass(queryJSON, queryClass)
		assert.Nil(t, err)

		dashboardUnit, errCode, _ := store.GetStore().CreateDashboardUnit(project.ID, agent.UUID, &model.DashboardUnit{
			DashboardId:  dashboard.ID,
			Title:        U.RandomString(5),
			Query:        queryJSON,
			Presentation: model.PresentationCard,
		}, model.DashboardUnitForNoQueryID)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboardUnit)

		dashboardUnitQueriesMap[dashboardUnit.ID] = make(map[string]interface{})
		dashboardUnitQueriesMap[dashboardUnit.ID]["class"] = queryClass
		dashboardUnitQueriesMap[dashboardUnit.ID]["queries"] = baseQuery
	}

	updatedUnitsCount := store.GetStore().CacheDashboardUnitsForProjectID(project.ID, 1)
	assert.Equal(t, len(dashboardQueriesStr), updatedUnitsCount)

	for key, rangeFunction := range U.QueryDateRangePresets {
		if key == "TODAY" {
			fmt.Println("RUNNING FOR TODAY")
		}
		from, to := rangeFunction()
		for unitID, queryMap := range dashboardUnitQueriesMap {
			queryClass := queryMap["class"].(string)
			queries := queryMap["queries"].(model.BaseQuery)
			queries.SetQueryDateRange(from, to)
			// Refresh is sent as false. Must return all presets range from cache.
			w := sendAnalyticsQueryReq(r, queryClass, project.ID, agent, dashboard.ID, unitID, queries, false, true)
			assert.NotNil(t, w)
			assert.Equal(t, http.StatusOK, w.Code)

			var result map[string]interface{}
			json.Unmarshal([]byte(w.Body.String()), &result)
			// Cache must be true in response.
			assert.True(t, result["cache"].(bool))

			// Refresh is sent as true. Still must return from cache for all presets except for todays.
			w = sendAnalyticsQueryReq(r, queryClass, project.ID, agent, dashboard.ID, unitID, queries, true, true)
			assert.NotNil(t, w)
			assert.Equal(t, http.StatusOK, w.Code)
			result = nil
			json.Unmarshal([]byte(w.Body.String()), &result)

			if from == U.GetBeginningOfDayTimestampZ(U.TimeNowUnix(), U.TimeZoneStringIST) {
				// If queried again with refresh as false, should return from cache.
				w := sendAnalyticsQueryReq(r, queryClass, project.ID, agent, dashboard.ID, unitID, queries, false, true)
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

	// _, errCode := model.CreateOrGetUserCreatedEventName(&model.EventName{ProjectId: project.ID, Name: "$session"})
	// assert.Equal(t, http.StatusCreated, errCode)

	customerAccountID := U.RandomLowerAphaNumString(5)
	_, errCode := store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		IntAdwordsCustomerAccountId: &customerAccountID,
	})

	dashboardName := U.RandomString(5)
	dashboard, errCode := store.GetStore().CreateDashboard(project.ID, agent.UUID, &model.Dashboard{Name: dashboardName, Type: model.DashboardTypeProjectVisible})
	assert.NotNil(t, dashboard)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Equal(t, dashboardName, dashboard.Name)

	dashboardUnitQueriesMap := make(map[uint64]map[string]interface{})

	dashboardQueriesStr := []string{
		`{ "query_group":[{ "channel": "google_ads", "select_metrics": ["impressions"], "filters": [], "group_by": [], "gbt": "hour", "fr": 1585679400, "to": 1585765800 }], "cl": "channel_v1" }`,
	}
	queryClass := model.QueryClassChannelV1
	for _, queryString := range dashboardQueriesStr {
		queryJSON := postgres.Jsonb{json.RawMessage(queryString)}
		baseQuery, err := model.DecodeQueryForClass(queryJSON, queryClass)
		assert.Nil(t, err)

		dashboardUnit, errCode, _ := store.GetStore().CreateDashboardUnit(project.ID, agent.UUID, &model.DashboardUnit{
			DashboardId:  dashboard.ID,
			Title:        U.RandomString(5),
			Query:        queryJSON,
			Presentation: model.PresentationCard,
		}, model.DashboardUnitForNoQueryID)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboardUnit)

		dashboardUnitQueriesMap[dashboardUnit.ID] = make(map[string]interface{})
		dashboardUnitQueriesMap[dashboardUnit.ID]["class"] = queryClass
		dashboardUnitQueriesMap[dashboardUnit.ID]["queries"] = baseQuery
	}

	updatedUnitsCount := store.GetStore().CacheDashboardUnitsForProjectID(project.ID, 1)
	assert.Equal(t, len(dashboardQueriesStr), updatedUnitsCount)

	for key, rangeFunction := range U.QueryDateRangePresets {
		if key == "TODAY" {
			fmt.Println("RUNNING FOR TODAY")
		}
		from, to := rangeFunction()
		for unitID, queryMap := range dashboardUnitQueriesMap {
			queryClass := queryMap["class"].(string)
			queries := queryMap["queries"].(model.BaseQuery)
			queries.SetQueryDateRange(from, to)
			// Refresh is sent as false. Must return all presets range from cache.
			w := sendAnalyticsQueryReq(r, queryClass, project.ID, agent, dashboard.ID, unitID, queries, false, true)
			assert.NotNil(t, w)
			assert.Equal(t, http.StatusOK, w.Code)

			var result map[string]interface{}
			json.Unmarshal([]byte(w.Body.String()), &result)
			// Cache must be true in response.
			assert.True(t, result["cache"].(bool))

			// Refresh is sent as true. Still must return from cache for all presets except for todays.
			w = sendAnalyticsQueryReq(r, queryClass, project.ID, agent, dashboard.ID, unitID, queries, true, true)
			assert.NotNil(t, w)
			assert.Equal(t, http.StatusOK, w.Code)
			result = nil
			json.Unmarshal([]byte(w.Body.String()), &result)

			if from == U.GetBeginningOfDayTimestampZ(U.TimeNowUnix(), U.TimeZoneStringIST) {
				// If queried again with refresh as false, should return from cache.
				w := sendAnalyticsQueryReq(r, queryClass, project.ID, agent, dashboard.ID, unitID, queries, false, true)
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

func TestCreateDashboardUnitWithDeletedQuery(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	assert.NotEmpty(t, project, agent)

	dashboardQuery, errCode, errMsg := store.GetStore().CreateQuery(project.ID, &model.Queries{
		ProjectID: project.ID,
		Type:      model.QueryTypeDashboardQuery,
		Query:     postgres.Jsonb{json.RawMessage(`{}`)},
	})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Empty(t, errMsg)
	assert.NotNil(t, dashboardQuery)

	// Delete the above query.
	errCode, errMsg = store.GetStore().DeleteDashboardQuery(project.ID, dashboardQuery.ID)
	assert.Equal(t, http.StatusAccepted, errCode)

	// Try creating a new dashboard unit for the deleted query.
	dashboard, errCode := store.GetStore().CreateDashboard(project.ID, agent.UUID,
		&model.Dashboard{Name: U.RandomString(5), Type: model.DashboardTypeProjectVisible})
	assert.NotNil(t, dashboard)
	assert.Equal(t, http.StatusCreated, errCode)

	dashboardUnit, errCode, errMsg := store.GetStore().CreateDashboardUnit(project.ID, agent.UUID,
		&model.DashboardUnit{DashboardId: dashboard.ID, Title: U.RandomString(5), Presentation: model.PresentationLine,
			QueryId: dashboardQuery.ID, Query: postgres.Jsonb{json.RawMessage(`{}`)}},
		model.DashboardUnitWithQueryID)
	assert.NotEqual(t, http.StatusCreated, errCode)
	assert.Empty(t, dashboardUnit)
}

func sendAttributionQueryReq(r *gin.Engine, projectID uint64, agent *model.Agent, dashboardID, unitID uint64, query model.AttributionQuery, refresh bool) *httptest.ResponseRecorder {
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

func sendAnalyticsQueryReq(r *gin.Engine, queryClass string, projectID uint64, agent *model.Agent, dashboardID,
	unitID uint64, baseQuery model.BaseQuery, refresh bool, withDashboardParams bool) *httptest.ResponseRecorder {
	return sendAnalyticsQueryReqWithHeader(r, queryClass, projectID, agent, dashboardID, unitID,
		baseQuery, refresh, withDashboardParams, map[string]string{})
}

func sendAnalyticsQueryReqWithHeader(r *gin.Engine, queryClass string, projectID uint64, agent *model.Agent, dashboardID,
	unitID uint64, baseQuery model.BaseQuery, refresh bool, withDashboardParams bool, headers map[string]string) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}

	queryURL, queryPayload := getAnalyticsQueryUrlAandPayload(queryClass, baseQuery)
	var requestURL string
	if queryClass == model.QueryClassWeb {
		requestURL = fmt.Sprintf("/projects/%d/dashboard/%d/units/query/web_analytics?refresh=%v", projectID, dashboardID, refresh)
	} else if withDashboardParams {
		requestURL = fmt.Sprintf("/projects/%d/%s?dashboard_id=%d&dashboard_unit_id=%d&refresh=%v",
			projectID, queryURL, dashboardID, unitID, refresh)
	} else {
		requestURL = fmt.Sprintf("/projects/%d/%s?refresh=%v",
			projectID, queryURL, refresh)
	}
	rb := U.NewRequestBuilder(http.MethodPost, requestURL).
		WithPostParams(queryPayload).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	for k, v := range headers {
		rb = rb.WithHeader(k, v)
	}

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error creating create dashboard_unit req.")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func getAnalyticsQueryUrlAandPayload(queryClass string, baseQuery model.BaseQuery) (string, interface{}) {
	var queryURL string
	var queryPayload interface{}
	if queryClass == model.QueryClassFunnel || queryClass == model.QueryClassInsights {
		queryURL = "query"
		query := baseQuery.(*model.Query)
		queryPayload = H.QueryRequestPayload{
			Query: *query,
		}
	} else if queryClass == model.QueryClassChannel {
		queryURL = "channels/query"
		query := baseQuery.(*model.ChannelQueryUnit)
		queryPayload = query.Query
	} else if queryClass == model.QueryClassChannelV1 {
		queryURL = "v1/channels/query"
		query := baseQuery.(*model.ChannelGroupQueryV1)
		queryPayload = query
	} else if queryClass == model.QueryClassEvents {
		queryURL = "v1/query"
		query := baseQuery.(*model.QueryGroup)
		queryPayload = query
	} else if queryClass == model.QueryClassWeb {
		query := baseQuery.(*model.DashboardUnitsWebAnalyticsQuery)
		queryPayload = query
	} else {
		queryURL = "attribution/query"
		query := baseQuery.(*model.AttributionQueryUnit)
		queryPayload = H.AttributionRequestPayload{
			Query: query.Query,
		}
	}
	return queryURL, queryPayload
}
