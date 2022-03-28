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
	"sync"
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

	dashboardQuery, errCode, errMsg := store.GetStore().CreateQuery(project.ID, &model.Queries{
		ProjectID: project.ID,
		Type:      model.QueryTypeDashboardQuery,
		Query:     postgres.Jsonb{json.RawMessage(`{}`)},
	})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Empty(t, errMsg)
	assert.NotNil(t, dashboardQuery)

	rName := U.RandomString(5)
	dashboard, errCode := store.GetStore().CreateDashboard(project.ID, agent.UUID, &model.Dashboard{Name: rName, Type: model.DashboardTypeProjectVisible})
	assert.NotNil(t, dashboard)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Equal(t, rName, dashboard.Name)

	t.Run("CreateDashboardUnit", func(t *testing.T) {
		dashboardUnit, errCode, errMsg := store.GetStore().CreateDashboardUnit(project.ID, agent.UUID, &model.DashboardUnit{DashboardId: dashboard.ID,
			Presentation: model.PresentationLine, QueryId: dashboardQuery.ID})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboardUnit)
		assert.Empty(t, errMsg)
		assert.NotNil(t, dashboardUnit.QueryId)
		assert.NotNil(t, dashboardUnit.Presentation)

		dashboardUnit1, errCode, errMsg := store.GetStore().CreateDashboardUnit(project.ID, agent.UUID, &model.DashboardUnit{DashboardId: dashboard.ID,
			Presentation: model.PresentationBar, QueryId: dashboardQuery.ID})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboardUnit1)
		assert.Empty(t, errMsg)
		assert.NotNil(t, dashboardUnit1.QueryId)
		assert.NotNil(t, dashboardUnit1.Presentation)

		dashboardUnit2, errCode, errMsg := store.GetStore().CreateDashboardUnit(project.ID, agent.UUID, &model.DashboardUnit{DashboardId: dashboard.ID,
			Presentation: model.PresentationCard, QueryId: dashboardQuery.ID})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboardUnit2)
		assert.Empty(t, errMsg)
		assert.NotNil(t, dashboardUnit2.QueryId)
		assert.NotNil(t, dashboardUnit2.Presentation)

	})

	t.Run("CreateDashboardUnitWithPresentation", func(t *testing.T) {
		dashboardUnit, errCode, errMsg := store.GetStore().CreateDashboardUnit(project.ID, agent.UUID, &model.DashboardUnit{DashboardId: dashboard.ID,
			Presentation: model.PresentationCard, QueryId: dashboardQuery.ID})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboardUnit)
		assert.Empty(t, errMsg)
		assert.NotNil(t, dashboardUnit.QueryId)
		assert.NotNil(t, dashboardUnit.Presentation)

		dashboardUnit1, errCode, errMsg := store.GetStore().CreateDashboardUnit(project.ID, agent.UUID, &model.DashboardUnit{DashboardId: dashboard.ID,
			Presentation: model.PresentationCard, QueryId: dashboardQuery.ID})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboardUnit1)
		assert.Empty(t, errMsg)
		assert.NotNil(t, dashboardUnit1.QueryId)
		assert.NotNil(t, dashboardUnit1.Presentation)

		dashboardUnit2, errCode, errMsg := store.GetStore().CreateDashboardUnit(project.ID, agent.UUID, &model.DashboardUnit{DashboardId: dashboard.ID,
			Presentation: model.PresentationCard, QueryId: dashboardQuery.ID})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboardUnit2)
		assert.Empty(t, errMsg)
		assert.NotNil(t, dashboardUnit2.QueryId)
		assert.NotNil(t, dashboardUnit2.Presentation)

	})

	t.Run("CreateDashboardUnit:Invalid", func(t *testing.T) {

		// invalid dashboard.
		rName = U.RandomString(5)
		dashboardUnit, errCode, _ := store.GetStore().CreateDashboardUnit(project.ID, agent.UUID, &model.DashboardUnit{DashboardId: 0,
			Presentation: model.PresentationLine, QueryId: dashboardQuery.ID})
		assert.Equal(t, http.StatusBadRequest, errCode)
		assert.Nil(t, dashboardUnit)

		// invalid project.
		rName = U.RandomString(5)
		dashboardUnit, errCode, _ = store.GetStore().CreateDashboardUnit(0, agent.UUID, &model.DashboardUnit{DashboardId: dashboard.ID,
			Presentation: model.PresentationLine, QueryId: dashboardQuery.ID})
		assert.Equal(t, http.StatusBadRequest, errCode)
		assert.Nil(t, dashboardUnit)

		// invalid agent.
		rName = U.RandomString(5)
		dashboardUnit, errCode, _ = store.GetStore().CreateDashboardUnit(project.ID, "", &model.DashboardUnit{DashboardId: dashboard.ID,
			Presentation: model.PresentationLine, QueryId: dashboardQuery.ID})
		assert.Equal(t, http.StatusBadRequest, errCode)
		assert.Nil(t, dashboardUnit)
	})

	t.Run("CreateDashboardUnit:WithoutAccessToDashboard", func(t *testing.T) {
		rName := U.RandomString(5)
		dashboard, errCode := store.GetStore().CreateDashboard(project.ID, agent.UUID, &model.Dashboard{Name: rName, Type: model.DashboardTypePrivate})
		assert.NotNil(t, dashboard)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.Equal(t, rName, dashboard.Name)

		dashboardUnit, errCode, _ := store.GetStore().CreateDashboardUnit(project.ID, agent2.UUID, &model.DashboardUnit{DashboardId: dashboard.ID,
			Presentation: model.PresentationLine, QueryId: dashboardQuery.ID})
		assert.Equal(t, http.StatusForbidden, errCode)
		assert.Nil(t, dashboardUnit)
	})

	t.Run("CreateDashboardUnit:DisallowAddingToWebAnalytics", func(t *testing.T) {
		rName := U.RandomString(5)
		dashboard, errCode := store.GetStore().CreateDashboard(
			project.ID, agent.UUID, &model.Dashboard{Name: rName, Type: model.DashboardTypeProjectVisible, Class: model.DashboardClassWebsiteAnalytics})
		assert.NotNil(t, dashboard)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.Equal(t, rName, dashboard.Name)

		dashboardUnit, errCode, _ := store.GetStore().CreateDashboardUnit(project.ID, agent2.UUID, &model.DashboardUnit{DashboardId: dashboard.ID,
			Presentation: model.PresentationLine, QueryId: dashboardQuery.ID})
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

	requestPayload1 := []model.DashboardUnitRequestPayload{{Description: U.RandomString(20)}}

	testArgs1 := args{requestPayload: requestPayload1,
		projectId:   project.ID,
		agentUUID:   agent.UUID,
		dashboardId: dashboard.ID}

	requestPayload2 := []model.DashboardUnitRequestPayload{{Description: U.RandomString(20)},
		{Description: U.RandomString(20)},
		{Description: U.RandomString(20)},
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
	requestPayload := model.DashboardUnitRequestPayload{Description: U.RandomString(20)}

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

	dashboardQuery, errCode, errMsg := store.GetStore().CreateQuery(project.ID, &model.Queries{
		ProjectID: project.ID,
		Type:      model.QueryTypeDashboardQuery,
		Query:     postgres.Jsonb{json.RawMessage(`{}`)},
	})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Empty(t, errMsg)
	assert.NotNil(t, dashboardQuery)

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
		dashboardUnit, errCode, errMsg := store.GetStore().CreateDashboardUnit(project.ID, agent.UUID, &model.DashboardUnit{DashboardId: dashboard.ID,
			Presentation: model.PresentationLine, QueryId: dashboardQuery.ID})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboardUnit)
		assert.Empty(t, errMsg)

		units, errCode := store.GetStore().GetDashboardUnits(project.ID, agent.UUID, dashboard.ID)
		assert.Equal(t, http.StatusFound, errCode)
		assert.Len(t, units, 1)
		assert.Equal(t, model.PresentationLine, units[0].Presentation)
	})

	t.Run("GetDashboardUnits:Invalid", func(t *testing.T) {
		dashboardUnit, errCode, errMsg := store.GetStore().CreateDashboardUnit(project.ID, agent.UUID, &model.DashboardUnit{DashboardId: dashboard.ID,
			Presentation: model.PresentationLine, QueryId: dashboardQuery.ID})
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

		dashboardUnit, errCode, _ := store.GetStore().CreateDashboardUnit(project.ID, agent.UUID, &model.DashboardUnit{DashboardId: dashboard.ID,
			Presentation: model.PresentationLine, QueryId: dashboardQuery.ID})
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

	dashboardQuery, errCode, errMsg := store.GetStore().CreateQuery(project.ID, &model.Queries{
		ProjectID: project.ID,
		Type:      model.QueryTypeDashboardQuery,
		Query:     postgres.Jsonb{json.RawMessage(`{}`)},
	})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Empty(t, errMsg)
	assert.NotNil(t, dashboardQuery)

	rName := U.RandomString(5)
	dashboard, errCode := store.GetStore().CreateDashboard(project.ID, agent.UUID, &model.Dashboard{Name: rName, Type: model.DashboardTypeProjectVisible})
	assert.NotNil(t, dashboard)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Equal(t, rName, dashboard.Name)

	t.Run("DeleteDashboardUnit", func(t *testing.T) {
		unit, errCode, _ := store.GetStore().CreateDashboardUnit(project.ID, agent.UUID, &model.DashboardUnit{DashboardId: dashboard.ID,
			Presentation: model.PresentationLine, QueryId: dashboardQuery.ID})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, unit)

		unit1, errCode, _ := store.GetStore().CreateDashboardUnit(project.ID, agent.UUID, &model.DashboardUnit{DashboardId: dashboard.ID,
			Presentation: model.PresentationBar, QueryId: dashboardQuery.ID})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, unit1)

		unit2, errCode, _ := store.GetStore().CreateDashboardUnit(project.ID, agent.UUID, &model.DashboardUnit{DashboardId: dashboard.ID,
			Presentation: model.PresentationCard, QueryId: dashboardQuery.ID})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, unit2)

		errCode = store.GetStore().DeleteDashboardUnit(project.ID, agent.UUID, dashboard.ID, unit.ID)
		assert.Equal(t, http.StatusAccepted, errCode)

		// should remove position given for unit on dashboard and rebalanced positions.
		_, errCode = store.GetStore().GetDashboard(project.ID, agent.UUID, dashboard.ID)
		assert.Equal(t, http.StatusFound, errCode)

		// Unit must have got soft deleted.
		_, errCode = store.GetStore().GetDashboardUnitByUnitID(unit.ProjectID, unit.ID)
		assert.Equal(t, http.StatusNotFound, errCode)
	})

	t.Run("DeleteDashboardUnit:Invalid", func(t *testing.T) {
		unit, errCode, _ := store.GetStore().CreateDashboardUnit(project.ID, agent.UUID, &model.DashboardUnit{DashboardId: dashboard.ID,
			Presentation: model.PresentationLine, QueryId: dashboardQuery.ID})
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

		dashboardUnit, errCode, _ := store.GetStore().CreateDashboardUnit(project.ID, agent.UUID, &model.DashboardUnit{DashboardId: dashboard.ID,
			Presentation: model.PresentationLine, QueryId: dashboardQuery.ID})
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
		&model.DashboardUnit{DashboardId: dashboard.ID, Presentation: model.PresentationLine,
			QueryId: dashboardQuery.ID})
	assert.NotEmpty(t, dashboardUnit)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Empty(t, errMsg)

	// Deleting dashboard unit should not delete the query of type QueryTypeDashboardQuery.
	unitID := dashboardUnit.ID
	errCode = store.GetStore().DeleteDashboardUnit(project.ID, agent.UUID, dashboard.ID, dashboardUnit.ID)
	assert.Equal(t, http.StatusAccepted, errCode)
	dashboardUnit, errCode = store.GetStore().GetDashboardUnitByUnitID(project.ID, unitID)
	assert.Empty(t, dashboardUnit)
	assert.Equal(t, http.StatusNotFound, errCode)
	query, errCode := store.GetStore().GetQueryWithQueryId(project.ID, dashboardQuery.ID)
	assert.NotEmpty(t, query)
	assert.Equal(t, http.StatusFound, errCode)

	// Dashboard unit with QueryTypeSavedQuery.
	dashboardUnit, errCode, errMsg = store.GetStore().CreateDashboardUnit(project.ID, agent.UUID,
		&model.DashboardUnit{DashboardId: dashboard.ID, Presentation: model.PresentationLine,
			QueryId: savedQuery.ID})
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

	dashboardQuery, errCode, errMsg := store.GetStore().CreateQuery(project.ID, &model.Queries{
		ProjectID: project.ID,
		Type:      model.QueryTypeDashboardQuery,
		Query:     postgres.Jsonb{json.RawMessage(`{}`)},
	})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Empty(t, errMsg)
	assert.NotNil(t, dashboardQuery)

	t.Run("UpdateDashboardUnit", func(t *testing.T) {
		rName := U.RandomString(5)
		dashboard, errCode := store.GetStore().CreateDashboard(project.ID, agent.UUID, &model.Dashboard{Name: rName, Type: model.DashboardTypeProjectVisible})
		assert.NotNil(t, dashboard)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.Equal(t, rName, dashboard.Name)

		unit, errCode, _ := store.GetStore().CreateDashboardUnit(project.ID, agent.UUID, &model.DashboardUnit{DashboardId: dashboard.ID,
			Presentation: model.PresentationLine, QueryId: dashboardQuery.ID})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, unit)

		description3 := "description3"
		presentation3 := "pr"
		updatedDashboardUnit3, errCode := store.GetStore().UpdateDashboardUnit(project.ID, agent.UUID, dashboard.ID, unit.ID, &model.DashboardUnit{Description: description3, Presentation: presentation3})
		assert.Equal(t, http.StatusAccepted, errCode)
		assert.Equal(t, description3, updatedDashboardUnit3.Description)
		assert.Equal(t, presentation3, updatedDashboardUnit3.Presentation)

		presentation4 := "pr"
		updatedDashboardUnit4, errCode := store.GetStore().UpdateDashboardUnit(project.ID, agent.UUID, dashboard.ID, unit.ID, &model.DashboardUnit{
			Presentation: presentation4})
		assert.Equal(t, http.StatusAccepted, errCode)
		assert.Equal(t, presentation4, updatedDashboardUnit4.Presentation)

		// invalid projectId.
		_, errCode = store.GetStore().UpdateDashboardUnit(0, agent.UUID, dashboard.ID, unit.ID, &model.DashboardUnit{})
		assert.Equal(t, http.StatusBadRequest, errCode)

		// invalid agentUUID.
		_, errCode = store.GetStore().UpdateDashboardUnit(project.ID, "", dashboard.ID, unit.ID, &model.DashboardUnit{})
		assert.Equal(t, http.StatusBadRequest, errCode)

		// invalid dashboardId.
		_, errCode = store.GetStore().UpdateDashboardUnit(project.ID, agent.UUID, 0, unit.ID, &model.DashboardUnit{})
		assert.Equal(t, http.StatusBadRequest, errCode)

		// invalid unitId.
		_, errCode = store.GetStore().UpdateDashboardUnit(project.ID, agent.UUID, dashboard.ID, 0, &model.DashboardUnit{})
		assert.Equal(t, http.StatusBadRequest, errCode)
	})

	t.Run("UpdateDashboardUnit:WithoutAccessToDashboard", func(t *testing.T) {
		rName := U.RandomString(5)
		dashboard, errCode := store.GetStore().CreateDashboard(project.ID, agent.UUID, &model.Dashboard{Name: rName, Type: model.DashboardTypePrivate})
		assert.NotNil(t, dashboard)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.Equal(t, rName, dashboard.Name)

		dashboardUnit, errCode, _ := store.GetStore().CreateDashboardUnit(project.ID, agent.UUID, &model.DashboardUnit{DashboardId: dashboard.ID,
			Presentation: model.PresentationLine, QueryId: dashboardQuery.ID})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboardUnit)

		_, errCode = store.GetStore().UpdateDashboardUnit(project.ID, agent2.UUID, dashboard.ID, dashboardUnit.ID, &model.DashboardUnit{})
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

func TestDashboardUnitEventForTimeZone(t *testing.T) {
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	H.InitAppRoutes(r)
	uri := "/sdk/event/track"

	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	userID1, _ := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	event_timestamp := 1575138601

	payload := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, "user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`, "s0", userID1, event_timestamp, "A", 1234)
	w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response := DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, "user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`, "s0", userID1, event_timestamp+10, "B", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	rName := U.RandomString(5)
	dashboard, _ := store.GetStore().CreateDashboard(project.ID, agent.UUID,
		&model.Dashboard{Name: rName, Type: model.DashboardTypeProjectVisible})

	query1 := model.Query{
		From: 1575138600,
		To:   1575224999,
		EventsWithProperties: []model.QueryEventWithProperties{
			model.QueryEventWithProperties{
				Name: "s0",
			},
		},
		Class:           model.QueryClassEvents,
		Type:            model.QueryTypeEventsOccurrence,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	sendCreateDashboardUnitReq(r, project.ID, agent, dashboard.ID, &model.DashboardUnitRequestPayload{
		Presentation: "pl", QueryId: 4})
	dashboardUnits, _ := store.GetStore().GetDashboardUnits(project.ID, agent.UUID, dashboard.ID)

	decChannelResult := struct {
		Cache  bool              `json:"cache"`
		Result model.QueryResult `json:"result"`
	}{}

	w = sendGetDashboardUnitResult(r, project.ID, agent, dashboardUnits[0].DashboardId, dashboardUnits[0].ID, &gin.H{"query": query1})
	assert.Equal(t, http.StatusOK, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &decChannelResult)
	assert.Equal(t, false, decChannelResult.Cache)

	query1.From = 1575158400
	query1.To = 1575244799
	query1.Timezone = "Europe/Lisbon"

	w = sendGetDashboardUnitResult(r, project.ID, agent, dashboardUnits[0].DashboardId, dashboardUnits[0].ID, &gin.H{"query": query1})
	assert.Equal(t, http.StatusOK, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &decChannelResult)
	assert.Equal(t, false, decChannelResult.Cache)
}

func TestDashboardUnitChannelForTimeZone(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, agent, _ := SetupProjectWithAgentDAO()

	customerAccountId := fmt.Sprintf("%d", U.RandomUint64())
	store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		IntAdwordsCustomerAccountId: &customerAccountId,
		IntAdwordsEnabledAgentUUID:  &agent.UUID,
	})
	rName := U.RandomString(5)
	dashboard, _ := store.GetStore().CreateDashboard(project.ID, agent.UUID,
		&model.Dashboard{Name: rName, Type: model.DashboardTypeProjectVisible})
	value := []byte(`{"id": 2061667885,"clicks":989, "campaign_id": 12,"impressions":10, "end_date": "20371230", "start_date": "20190711", "conversions":111, "cost":42.94}`)
	document := model.AdwordsDocument{
		ProjectID:         project.ID,
		CustomerAccountID: customerAccountId,
		Type:              5,
		Timestamp:         20191201,
		ID:                "2061667885",
		Value:             &postgres.Jsonb{value},
		TypeAlias:         "campaign_performance_report",
	}
	errCode := store.GetStore().CreateAdwordsDocument(&document)
	log.Warn(errCode)
	query := &model.ChannelQuery{
		Channel:     "google_ads",
		FilterKey:   "campaign",
		FilterValue: "all",
		From:        1575138600,
		To:          1575224999,
	}
	sendCreateDashboardUnitReq(r, project.ID, agent, dashboard.ID, &model.DashboardUnitRequestPayload{
		Presentation: "pc", QueryId: 5})
	decChannelResult := struct {
		Cache  bool                     `json:"cache"`
		Result model.ChannelQueryResult `json:"result"`
	}{}
	dashboardUnits, _ := store.GetStore().GetDashboardUnits(project.ID, agent.UUID, dashboard.ID)
	w := sendGetDashboardUnitChannelResult(r, project.ID, agent, dashboardUnits[0].DashboardId, dashboardUnits[0].ID, query)
	json.Unmarshal(w.Body.Bytes(), &decChannelResult)
	assert.Equal(t, false, decChannelResult.Cache)
	assert.Equal(t, float64(989), (*decChannelResult.Result.Metrics)["clicks"])

	w = sendGetDashboardUnitChannelResult(r, project.ID, agent, dashboardUnits[0].DashboardId, dashboardUnits[0].ID, query)
	json.Unmarshal(w.Body.Bytes(), &decChannelResult)
	assert.Equal(t, true, decChannelResult.Cache)
	assert.Equal(t, float64(989), (*decChannelResult.Result.Metrics)["clicks"])

	query.Timezone = "Europe/Lisbon"
	query.From = 1575158400
	query.To = 1575244799
	w = sendGetDashboardUnitChannelResult(r, project.ID, agent, dashboardUnits[0].DashboardId, dashboardUnits[0].ID, query)
	json.Unmarshal(w.Body.Bytes(), &decChannelResult)
	assert.Equal(t, false, decChannelResult.Cache)
	assert.Equal(t, float64(989), (*decChannelResult.Result.Metrics)["clicks"])

	// Evaluating for channelv1 handler.
	query1 := &model.ChannelGroupQueryV1{
		Class: "channel_v1",
		Queries: []model.ChannelQueryV1{{Channel: "google_ads", SelectMetrics: []string{"clicks"},
			Timezone: string(U.TimeZoneStringIST), From: 1575138600, To: 1575224999, GroupByTimestamp: "",
			Filters: []model.ChannelFilterV1{}, GroupBy: []model.ChannelGroupBy{}}},
	}
	sendCreateDashboardUnitReq(r, project.ID, agent, dashboard.ID, &model.DashboardUnitRequestPayload{
		Presentation: "pc", QueryId: 6})
	decChannelResult1 := struct {
		Cache  bool                       `json:"cache"`
		Result model.ChannelResultGroupV1 `json:"result"`
	}{}
	dashboardUnits, _ = store.GetStore().GetDashboardUnits(project.ID, agent.UUID, dashboard.ID)

	w = sendGetDashboardUnitChannelV1Result(r, project.ID, agent, dashboardUnits[0].DashboardId, dashboardUnits[0].ID, query1)
	json.Unmarshal(w.Body.Bytes(), &decChannelResult1)
	assert.Equal(t, false, decChannelResult.Cache)
	assert.Equal(t, float64(989), (*&decChannelResult1.Result.Results[0].Rows[0][0]))

	w = sendGetDashboardUnitChannelV1Result(r, project.ID, agent, dashboardUnits[0].DashboardId, dashboardUnits[0].ID, query1)
	json.Unmarshal(w.Body.Bytes(), &decChannelResult1)
	assert.Equal(t, true, decChannelResult1.Cache)
	assert.Equal(t, float64(989), (*&decChannelResult1.Result.Results[0].Rows[0][0]))

	query1.Queries[0].Timezone = "Europe/Lisbon"
	query1.Queries[0].From = 1575138600
	query1.Queries[0].To = 1575224999
	w = sendGetDashboardUnitChannelV1Result(r, project.ID, agent, dashboardUnits[0].DashboardId, dashboardUnits[0].ID, query1)
	json.Unmarshal(w.Body.Bytes(), &decChannelResult1)
	assert.Equal(t, true, decChannelResult1.Cache)
	assert.Equal(t, float64(989), (*&decChannelResult1.Result.Results[0].Rows[0][0]))

	query1.Queries[0].Timezone = "Europe/Lisbon"
	query1.Queries[0].From = 1575158400
	query1.Queries[0].To = 1575244799
	w = sendGetDashboardUnitChannelV1Result(r, project.ID, agent, dashboardUnits[0].DashboardId, dashboardUnits[0].ID, query1)
	json.Unmarshal(w.Body.Bytes(), &decChannelResult1)
	assert.Equal(t, false, decChannelResult1.Cache)
	assert.Equal(t, float64(989), (*&decChannelResult1.Result.Results[0].Rows[0][0]))

}

func TestCacheDashboardUnitsForProjectID(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	project.TimeZone = string(U.TimeZoneStringIST)
	store.GetStore().UpdateProject(project.ID, project)

	_, errCode := store.GetStore().CreateOrGetUserCreatedEventName(&model.EventName{ProjectId: project.ID, Name: "$session"})
	assert.Equal(t, http.StatusCreated, errCode)

	customerAccountID := U.RandomLowerAphaNumString(5)
	_, errCode = store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		IntAdwordsCustomerAccountId: &customerAccountID,
	})

	timezonestring := U.TimeZoneString(project.TimeZone)
	dashboardName := U.RandomString(5)
	dashboard, errCode := store.GetStore().CreateDashboard(project.ID, agent.UUID, &model.Dashboard{Name: dashboardName, Type: model.DashboardTypeProjectVisible})
	assert.NotNil(t, dashboard)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Equal(t, dashboardName, dashboard.Name)

	dashboardUnitQueriesMap := make(map[uint64]map[string]interface{})
	var dashboardQueriesStr = map[string]string{
		model.QueryClassInsights:    `{"cl": "insights", "ec": "any_given_event", "fr": 1393612200, "to": 1396290599, "ty": "events_occurrence", "tz": "", "ewp": [{"na": "$session", "pr": []}], "gbp": [], "gbt": ""}`,
		model.QueryClassFunnel:      `{"cl": "funnel", "ec": "any_given_event", "fr": 1594492200, "to": 1594578599, "ty": "unique_users", "tz": "", "ewp": [{"na": "$session", "pr": []}, {"na": "www.chargebee.com/schedule-a-demo", "pr": []}], "gbp": [], "gbt": ""}`,
		model.QueryClassAttribution: `{"cl": "attribution", "meta": {"metrics_breakdown": true}, "query": {"ce": {"na": "$session", "pr": []}, "cm": ["Impressions", "Clicks", "Spend"], "to": 1596479399, "lbw": 1, "lfe": [], "from": 1595874600, "attribution_key": "Campaign", "attribution_methodology": "Last_Touch"}}`,
		model.QueryClassChannel:     `{"cl": "channel", "meta": {"metric": "total_cost"}, "query": {"to": 1576060774, "from": 1573468774, "channel": "google_ads", "filter_key": "campaign", "filter_value": "all"}}`,
		model.QueryClassKPI:         `{"cl":"kpi","qG":[{"ca":"events","pgUrl":"www.acme.com/pricing","dc":"page_views","me":["page_views"],"gBy":[],"fil":[],"gbt":"","fr":1633233600,"to":1633579199}],"gFil":[],"gGBy":[]}`,
	}
	var dashboardQueryClassList []string
	var dashboardUnitsList []model.DashboardUnit
	for queryClass, queryString := range dashboardQueriesStr {
		dashboardQueryClassList = append(dashboardQueryClassList, queryClass)
		queryJSON := postgres.Jsonb{json.RawMessage(queryString)}
		baseQuery, err := model.DecodeQueryForClass(queryJSON, queryClass)
		assert.Nil(t, err)

		dashboardQuery, errCode, errMsg := store.GetStore().CreateQuery(project.ID, &model.Queries{
			ProjectID: project.ID,
			Type:      model.QueryTypeDashboardQuery,
			Query:     postgres.Jsonb{json.RawMessage(queryString)},
		})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.Empty(t, errMsg)
		assert.NotNil(t, dashboardQuery)

		dashboardUnit, errCode, _ := store.GetStore().CreateDashboardUnit(project.ID, agent.UUID, &model.DashboardUnit{
			DashboardId:  dashboard.ID,
			Presentation: model.PresentationCard,
			QueryId:      dashboardQuery.ID,
		})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboardUnit)
		dashboardUnitsList = append(dashboardUnitsList, *dashboardUnit)
		dashboardUnitQueriesMap[dashboardUnit.ID] = make(map[string]interface{})
		dashboardUnitQueriesMap[dashboardUnit.ID]["class"] = queryClass
		dashboardUnitQueriesMap[dashboardUnit.ID]["query"] = baseQuery
	}
	var reportCollector sync.Map
	//dashboardUnitIDs := make([]uint64, 0)
	updatedUnitsCount := store.GetStore().CacheDashboardUnitsForProjectID(project.ID, dashboardUnitsList, dashboardQueryClassList, 1, &reportCollector)
	assert.Equal(t, 5, updatedUnitsCount)

	for rangeString, rangeFunction := range U.QueryDateRangePresets {
		from, to, errCode := rangeFunction(timezonestring)
		assert.Nil(t, errCode)
		for unitID, queryMap := range dashboardUnitQueriesMap {
			queryClass := queryMap["class"].(string)
			query := queryMap["query"].(model.BaseQuery)
			if queryClass == model.QueryClassAttribution {
				f, _ := model.GetEffectiveTimeRangeForDashboardUnitAttributionQuery(from, to)
				if f == 0 {
					continue
				}
			}
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
			if (rangeString == "CURRENT_WEEK" || rangeString == "CURRENT_MONTH") && queryClass == model.QueryClassAttribution {
				continue
			}
			w = sendAnalyticsQueryReq(r, queryClass, project.ID, agent, 0, 0, query, false, false)
			assert.NotEmpty(t, w)
			assert.Equal(t, http.StatusOK, w.Code)
			if queryClass != model.QueryClassWeb {
				// For website analytics, it returns from Dashboard cache.
				assert.Equal(t, "true", w.Header().Get(model.QueryCacheResponseFromCacheHeader), assertMsg)
			}

			if from == U.GetBeginningOfDayTimestampIn(U.TimeNowUnix(), timezonestring) {
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

	timezoneString := U.TimeZoneString(project.TimeZone)
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
	var dashboardQueryClassList []string
	var dashboardUnitsList []model.DashboardUnit
	for _, queryString := range dashboardQueriesStr {
		queryJSON := postgres.Jsonb{json.RawMessage(queryString)}
		baseQuery, err := model.DecodeQueryForClass(queryJSON, queryClass)
		assert.Nil(t, err)

		dashboardQuery, errCode, errMsg := store.GetStore().CreateQuery(project.ID, &model.Queries{
			ProjectID: project.ID,
			Type:      model.QueryTypeDashboardQuery,
			Query:     postgres.Jsonb{json.RawMessage(queryString)},
			Settings:  postgres.Jsonb{RawMessage: json.RawMessage(`{"size": 100}`)},
		})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.Empty(t, errMsg)
		assert.NotNil(t, dashboardQuery)

		dashboardUnit, errCode, _ := store.GetStore().CreateDashboardUnit(project.ID, agent.UUID, &model.DashboardUnit{
			DashboardId:  dashboard.ID,
			Presentation: model.PresentationCard,
			QueryId:      dashboardQuery.ID,
		})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboardUnit)
		dashboardUnitsList = append(dashboardUnitsList, *dashboardUnit)
		dashboardQueryClassList = append(dashboardQueryClassList, queryClass)
		dashboardUnitQueriesMap[dashboardUnit.ID] = make(map[string]interface{})
		dashboardUnitQueriesMap[dashboardUnit.ID]["class"] = queryClass
		dashboardUnitQueriesMap[dashboardUnit.ID]["queries"] = baseQuery
	}
	var reportCollector sync.Map
	//dashboardUnitIDs := make([]uint64, 0)
	updatedUnitsCount := store.GetStore().CacheDashboardUnitsForProjectID(project.ID, dashboardUnitsList, dashboardQueryClassList, 1, &reportCollector)
	assert.Equal(t, len(dashboardQueriesStr), updatedUnitsCount)
	for _, rangeFunction := range U.QueryDateRangePresets {
		from, to, errCode := rangeFunction(timezoneString)
		assert.Nil(t, errCode)
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

			if from == U.GetBeginningOfDayTimestampIn(U.TimeNowUnix(), timezoneString) {
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
	var dashboardQueryClassList []string
	var dashboardUnitsList []model.DashboardUnit
	for _, queryString := range dashboardQueriesStr {
		queryJSON := postgres.Jsonb{json.RawMessage(queryString)}
		baseQuery, err := model.DecodeQueryForClass(queryJSON, queryClass)
		assert.Nil(t, err)

		dashboardQuery, errCode, errMsg := store.GetStore().CreateQuery(project.ID, &model.Queries{
			ProjectID: project.ID,
			Type:      model.QueryTypeDashboardQuery,
			Query:     postgres.Jsonb{json.RawMessage(queryString)},
			Title:     "title_xyz",
			Settings:  postgres.Jsonb{RawMessage: json.RawMessage(`{"size": 100}`)},
		})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.Empty(t, errMsg)
		assert.NotNil(t, dashboardQuery)

		dashboardUnit, errCode, _ := store.GetStore().CreateDashboardUnit(project.ID, agent.UUID, &model.DashboardUnit{
			DashboardId:  dashboard.ID,
			Presentation: model.PresentationCard,
			QueryId:      dashboardQuery.ID,
		})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, dashboardUnit)
		dashboardUnitsList = append(dashboardUnitsList, *dashboardUnit)
		dashboardQueryClassList = append(dashboardQueryClassList, queryClass)
		dashboardUnitQueriesMap[dashboardUnit.ID] = make(map[string]interface{})
		dashboardUnitQueriesMap[dashboardUnit.ID]["class"] = queryClass
		dashboardUnitQueriesMap[dashboardUnit.ID]["queries"] = baseQuery
	}

	var reportCollector sync.Map
	//dashboardUnitIDs := make([]uint64, 0)
	updatedUnitsCount := store.GetStore().CacheDashboardUnitsForProjectID(project.ID, dashboardUnitsList, dashboardQueryClassList, 1, &reportCollector)
	assert.Equal(t, len(dashboardQueriesStr), updatedUnitsCount)
	timezonestring := U.TimeZoneString(project.TimeZone)
	for _, rangeFunction := range U.QueryDateRangePresets {
		from, to, errCode := rangeFunction(timezonestring)
		assert.Nil(t, errCode)
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

			if from == U.GetBeginningOfDayTimestampIn(U.TimeNowUnix(), U.TimeZoneString(timezonestring)) {
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

// Testing by taking lastMonth into consideration.
// Cache for lastMonth should be filled with data And normal query without lastXDays should return some values. but with lastXDays, it should return 0.
func TestDashboardUnitEventForDateTypeFilters(t *testing.T) {
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	H.InitAppRoutes(r)
	uri := "/sdk/event/track"
	lastXDays := int64(5)

	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	userID1, _ := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	event_timestamp := time.Now().AddDate(0, -1, 0).Unix()
	timezoneString, _ := store.GetStore().GetTimezoneForProject(project.ID)

	payload := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, "user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d, "$timestamp":%d}}`, "s0", userID1, event_timestamp, "A", 1234, time.Now().Unix())
	w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response := DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, "user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d, "$timestamp":%d}}`, "s0", userID1, event_timestamp+10, "B", 4321, time.Now().Unix())
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	rName := U.RandomString(5)
	dashboard, _ := store.GetStore().CreateDashboard(project.ID, agent.UUID,
		&model.Dashboard{Name: rName, Type: model.DashboardTypeProjectVisible})

	dateTimeValue := model.DateTimePropertyValue{
		From:           0,
		To:             0,
		OverridePeriod: false,
		Number:         lastXDays,
		Granularity:    U.GranularityDays,
	}
	stringifiedDateTimeValue, _ := json.Marshal(dateTimeValue)
	var query1 model.BaseQuery
	query1 = &model.Query{
		From:     1575138600,
		To:       1575224999,
		Timezone: string(timezoneString),
		EventsWithProperties: []model.QueryEventWithProperties{
			model.QueryEventWithProperties{
				Name: "s0",
				Properties: []model.QueryProperty{
					model.QueryProperty{
						Entity:    "event",
						Type:      U.PropertyTypeDateTime,
						Property:  "_$timestamp",
						LogicalOp: "AND",
						Operator:  model.InLastStr,
						Value:     string(stringifiedDateTimeValue),
					}},
			},
		},
		Class:           model.QueryClassInsights,
		Type:            model.QueryTypeEventsOccurrence,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	queryJson, _ := json.Marshal(query1)
	dashboardQuery, _, _ := store.GetStore().CreateQuery(project.ID, &model.Queries{
		ProjectID: project.ID,
		Type:      model.QueryTypeDashboardQuery,
		Query:     postgres.Jsonb{queryJson},
		Title:     "title_xyz",
		Settings:  postgres.Jsonb{RawMessage: json.RawMessage(`{"size": 100}`)},
	})

	dashboardUnit, _, _ := store.GetStore().CreateDashboardUnit(project.ID, agent.UUID,
		&model.DashboardUnit{
			DashboardId:  dashboard.ID,
			Description:  U.RandomString(5),
			QueryId:      dashboardQuery.ID,
			Presentation: model.PresentationCard,
		})
	dashboardUnitQueriesMap := make(map[uint64]map[string]interface{})
	dashboardUnitQueriesMap[dashboardUnit.ID] = make(map[string]interface{})
	dashboardUnitQueriesMap[dashboardUnit.ID]["class"] = query1.GetClass()
	dashboardUnitQueriesMap[dashboardUnit.ID]["query"] = query1
	var dashboardQueryClassList []string
	var dashboardUnitsList []model.DashboardUnit
	dashboardUnitsList = append(dashboardUnitsList, *dashboardUnit)
	dashboardQueryClassList = append(dashboardQueryClassList, query1.GetClass())
	var reportCollector sync.Map
	//dashboardUnitIDs := make([]uint64, 0)
	store.GetStore().CacheDashboardUnitsForProjectID(project.ID, dashboardUnitsList, dashboardQueryClassList, 1, &reportCollector)
	result := struct {
		Cache  bool              `json:"cache"`
		Result model.QueryResult `json:"result"`
	}{}

	for unitID, queryMap := range dashboardUnitQueriesMap {
		queryClass := queryMap["class"].(string)
		query := queryMap["query"].(model.BaseQuery)
		from, to, _ := U.GetQueryRangePresetLastMonthIn(timezoneString)
		query.SetQueryDateRange(from, to)
		w := sendAnalyticsQueryReq(r, queryClass, project.ID, agent, dashboard.ID, unitID, query, false, true)
		assert.NotNil(t, w)
		assert.Equal(t, http.StatusOK, w.Code)
		err = json.Unmarshal(w.Body.Bytes(), &result)
		assert.Equal(t, true, result.Cache)
		assert.Equal(t, result.Result.Rows[0][0], float64(2))
	}
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
	rb := C.NewRequestBuilderWithPrefix(http.MethodPost, requestURL).
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
	} else if queryClass == model.QueryClassKPI {
		queryURL = "v1/kpi/query"
		query := baseQuery.(*model.KPIQueryGroup)
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

func TestShouldCacheUnitForTimeRange(t *testing.T) {
	type args struct {
		queryClass      string
		preset          string
		from            int64
		to              int64
		onlyAttribution int
		skipAttribution int
	}
	july1Start := int64(1625077800)
	july1End := int64(1625164199)
	july2End := int64(1625250599)
	july3End := int64(1625336999)

	sundayStart := int64(1625337000)
	sundayEnd := int64(1625423399)
	mondayEnd := int64(1625509799)
	tuesdayEnd := int64(1625596199)

	tests := []struct {
		name  string
		args  args
		want  bool
		want1 int64
		want2 int64
	}{
		{"TestDateRangePresetToday", args{queryClass: model.QueryClassAttribution, preset: U.DateRangePresetToday, from: july1Start, to: july1End, onlyAttribution: 0, skipAttribution: 0}, false, 0, 0},
		{"TestDateRangePresetYesterday", args{queryClass: model.QueryClassAttribution, preset: U.DateRangePresetYesterday, from: july1Start, to: july1End, onlyAttribution: 0, skipAttribution: 0}, false, 0, 0},

		{"TestDateRangePresetCurrentMonth1", args{queryClass: model.QueryClassAttribution, preset: U.DateRangePresetCurrentMonth, from: july1Start, to: july1End, onlyAttribution: 0, skipAttribution: 0}, false, 0, 0},
		{"TestDateRangePresetCurrentMonth2", args{queryClass: model.QueryClassAttribution, preset: U.DateRangePresetCurrentMonth, from: july1Start, to: july2End, onlyAttribution: 0, skipAttribution: 0}, true, july1Start, july1End},
		{"TestDateRangePresetCurrentMonth3", args{queryClass: model.QueryClassAttribution, preset: U.DateRangePresetCurrentMonth, from: july1Start, to: july3End, onlyAttribution: 0, skipAttribution: 0}, true, july1Start, july3End - U.SECONDS_IN_A_DAY},

		{"TestDateRangePresetCurrentWeek1", args{queryClass: model.QueryClassAttribution, preset: U.DateRangePresetCurrentWeek, from: sundayStart, to: sundayEnd, onlyAttribution: 0, skipAttribution: 0}, false, 0, 0},
		{"TestDateRangePresetCurrentWeek2", args{queryClass: model.QueryClassAttribution, preset: U.DateRangePresetCurrentWeek, from: sundayStart, to: mondayEnd, onlyAttribution: 0, skipAttribution: 0}, true, sundayStart, sundayEnd},
		{"TestDateRangePresetCurrentWeek3", args{queryClass: model.QueryClassAttribution, preset: U.DateRangePresetCurrentWeek, from: sundayStart, to: tuesdayEnd, onlyAttribution: 0, skipAttribution: 0}, true, sundayStart, tuesdayEnd - U.SECONDS_IN_A_DAY},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2 := model.ShouldCacheUnitForTimeRange(tt.args.queryClass, tt.args.preset, tt.args.from, tt.args.to, tt.args.onlyAttribution, tt.args.skipAttribution)
			if got != tt.want {
				t.Errorf("ShouldCacheUnitForTimeRange() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("ShouldCacheUnitForTimeRange() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("ShouldCacheUnitForTimeRange() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func TestGetEffectiveTimeRangeForDashboardUnitAttributionQuery(t *testing.T) {
	type args struct {
		from int64
		to   int64
	}
	// Past
	july1Start := int64(1625077800)
	july1End := int64(1625164199)
	july2End := int64(1625250599)
	july3End := int64(1625336999)

	// Today
	toValid1 := U.GetBeginningOfDayTimestampIn(time.Now().Unix(), U.TimeZoneStringIST) - 1
	fromValid1 := toValid1 - 7*U.SECONDS_IN_A_DAY + 1
	toValid2 := U.GetBeginningOfDayTimestampIn(time.Now().Unix(), U.TimeZoneStringIST) - 1
	fromValid2 := toValid2 - 1*U.SECONDS_IN_A_DAY + 1

	tests := []struct {
		name  string
		args  args
		want  int64
		want1 int64
	}{
		// Past
		{"Test1", args{from: july1Start, to: july1End}, july1Start, july1End},
		{"Test2", args{from: july1Start, to: july2End}, july1Start, july2End},
		{"Test3", args{from: july1Start, to: july3End}, july1Start, july3End},
		{"Test3", args{from: july1Start, to: july3End + 10*U.SECONDS_IN_A_DAY}, july1Start, july3End + 10*U.SECONDS_IN_A_DAY},

		// Current
		{"Test4", args{from: fromValid1, to: toValid1}, fromValid1, toValid1 - U.SECONDS_IN_A_DAY},
		{"Test4", args{from: fromValid2, to: toValid2}, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := model.GetEffectiveTimeRangeForDashboardUnitAttributionQuery(tt.args.from, tt.args.to)
			if got != tt.want {
				t.Errorf("GetEffectiveTimeRangeForDashboardUnitAttributionQuery() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("GetEffectiveTimeRangeForDashboardUnitAttributionQuery() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
