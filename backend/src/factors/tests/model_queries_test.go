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

func TestModelQuery(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	agent2, errCode := SetupAgentReturnDAO(getRandomEmail(), "+13425356")
	assert.Equal(t, http.StatusCreated, errCode)
	_, errCode = M.CreateProjectAgentMappingWithDependencies(&M.ProjectAgentMapping{
		ProjectID: project.ID, AgentUUID: agent2.UUID})
	assert.Equal(t, http.StatusCreated, errCode)

	var queryId uint64
	t.Run("CreateQuery:SavedQuery:valid", func(t *testing.T) {
		rName1 := U.RandomString(5)
		query, errCode, errMsg := M.CreateQuery(project.ID, &M.Queries{ProjectID: project.ID,
			Title: rName1, Type: 2, CreatedBy: agent.UUID, Query: postgres.Jsonb{json.RawMessage(`{}`)}})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, query)
		assert.Empty(t, errMsg)
		queryId = query.ID
	})
	//No agentUUID for saved Query && empty title
	t.Run("CreateQuery:SavedQuery:invalid", func(t *testing.T) {
		rName1 := U.RandomString(5)
		_, errCode, _ := M.CreateQuery(project.ID, &M.Queries{ProjectID: project.ID,
			Title: rName1, Type: 2, CreatedBy: "", Query: postgres.Jsonb{json.RawMessage(`{}`)}})
		assert.Equal(t, http.StatusBadRequest, errCode)

		_, errCode, _ = M.CreateQuery(project.ID, &M.Queries{ProjectID: project.ID,
			Title: "", Type: 2, CreatedBy: agent.UUID, Query: postgres.Jsonb{json.RawMessage(`{}`)}})
		assert.Equal(t, http.StatusBadRequest, errCode)
	})
	// Get Query test
	query, errCode := M.GetQueryWithQueryId(project.ID, queryId)
	assert.Equal(t, http.StatusFound, errCode)

	t.Run("UpdateSavedQuery:Valid", func(t *testing.T) {
		rName1 := U.RandomString(5)
		query1, errCode := M.UpdateSavedQuery(project.ID, queryId, &M.Queries{Title: rName1})
		assert.Equal(t, http.StatusAccepted, errCode)
		assert.Equal(t, rName1, query1.Title)
		assert.NotEqual(t, query1.Title, query.Title)
	})

	t.Run("UpdateSavedQuery:Invalid", func(t *testing.T) {
		_, errCode := M.UpdateSavedQuery(project.ID, queryId, &M.Queries{Title: ""})
		assert.Equal(t, http.StatusBadRequest, errCode)
	})

	// test delete query
	errCode, errMsg := M.DeleteSavedQuery(project.ID, queryId)
	assert.Equal(t, errCode, http.StatusAccepted)
	assert.Empty(t, errMsg)

	//Check if deleted
	_, errCode = M.GetQueryWithQueryId(project.ID, queryId)
	assert.Equal(t, http.StatusNotFound, errCode)

	// test search query
	rName1 := "Hello"
	query1, errCode, errMsg := M.CreateQuery(project.ID, &M.Queries{ProjectID: project.ID,
		Title: rName1, Type: M.QueryTypeSavedQuery, CreatedBy: agent.UUID, Query: postgres.Jsonb{json.RawMessage(`{}`)}})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotNil(t, query1)
	assert.Empty(t, errMsg)

	rName2 := "World"
	query2, errCode, errMsg := M.CreateQuery(project.ID, &M.Queries{ProjectID: project.ID,
		Title: rName2, Type: M.QueryTypeDashboardQuery, CreatedBy: agent.UUID, Query: postgres.Jsonb{json.RawMessage(`{}`)}})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotNil(t, query2)
	assert.Empty(t, errMsg)

	queries, errCode := M.SearchQueriesWithProjectId(project.ID, "Hello")
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, 1, len(queries))

	queries, errCode = M.SearchQueriesWithProjectId(project.ID, "o")
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, 2, len(queries))
}

func TestDeleteQuery(t *testing.T) {
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

	// Delete dashboard query should not delete saved type query.
	M.DeleteDashboardQuery(project.ID, savedQuery.ID)
	query, errCode := M.GetQueryWithQueryId(project.ID, savedQuery.ID)
	assert.NotEmpty(t, query)

	// Delete saved query should not delete dashboard type query.
	M.DeleteSavedQuery(project.ID, dashboardQuery.ID)
	query, errCode = M.GetQueryWithQueryId(project.ID, dashboardQuery.ID)
	assert.NotEmpty(t, query)

	// Should delete this time.
	M.DeleteSavedQuery(project.ID, savedQuery.ID)
	query, errCode = M.GetQueryWithQueryId(project.ID, savedQuery.ID)
	assert.Empty(t, query)
	assert.Equal(t, http.StatusNotFound, errCode)

	// Should delete this time.
	M.DeleteDashboardQuery(project.ID, dashboardQuery.ID)
	query, errCode = M.GetQueryWithQueryId(project.ID, dashboardQuery.ID)
	assert.Empty(t, query)
	assert.Equal(t, http.StatusNotFound, errCode)
}

func TestDeleteQueryWithDashboardUnit(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	assert.NotEmpty(t, project, agent)

	dashboardQuery, errCode, errMsg := M.CreateQuery(project.ID, &M.Queries{
		ProjectID: project.ID,
		Type:      M.QueryTypeDashboardQuery,
	})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Empty(t, errMsg)
	assert.NotNil(t, dashboardQuery)

	dashboard, errCode := M.CreateDashboard(project.ID, agent.UUID,
		&M.Dashboard{Name: U.RandomString(5), Type: M.DashboardTypeProjectVisible})
	assert.NotNil(t, dashboard)
	assert.Equal(t, http.StatusCreated, errCode)

	// Two Dashboard units with query type QueryTypeDashboardUnit.
	dashboardUnit1, errCode, errMsg := M.CreateDashboardUnit(project.ID, agent.UUID,
		&M.DashboardUnit{DashboardId: dashboard.ID, Title: U.RandomString(5), Presentation: M.PresentationLine,
			QueryId: dashboardQuery.ID, Query: postgres.Jsonb{json.RawMessage(`{}`)}},
		M.DashboardUnitWithQueryID)
	assert.NotEmpty(t, dashboardUnit1)

	dashboardUnit2, errCode, errMsg := M.CreateDashboardUnit(project.ID, agent.UUID,
		&M.DashboardUnit{DashboardId: dashboard.ID, Title: U.RandomString(5), Presentation: M.PresentationLine,
			QueryId: dashboardQuery.ID, Query: postgres.Jsonb{json.RawMessage(`{}`)}},
		M.DashboardUnitWithQueryID)
	assert.NotEmpty(t, dashboardUnit2)

	// Should not allow direct delete since units exists.
	errCode, errMsg = M.DeleteDashboardQuery(project.ID, dashboardQuery.ID)
	assert.Equal(t, http.StatusNotAcceptable, errCode)
	query, errCode := M.GetQueryWithQueryId(project.ID, dashboardQuery.ID)
	assert.NotEmpty(t, query)

	// On deleting one of the unit, should not delete the query.
	errCode = M.DeleteDashboardUnit(project.ID, agent.UUID, dashboard.ID, dashboardUnit1.ID)
	assert.Equal(t, http.StatusAccepted, errCode)
	query, errCode = M.GetQueryWithQueryId(project.ID, dashboardQuery.ID)
	assert.NotEmpty(t, query)

	// On deleting the other unit, it should now delete the undelying query.
	errCode = M.DeleteDashboardUnit(project.ID, agent.UUID, dashboard.ID, dashboardUnit2.ID)
	assert.Equal(t, http.StatusAccepted, errCode)
	query, errCode = M.GetQueryWithQueryId(project.ID, dashboardQuery.ID)
	assert.Empty(t, query)
	assert.Equal(t, http.StatusNotFound, errCode)
}
