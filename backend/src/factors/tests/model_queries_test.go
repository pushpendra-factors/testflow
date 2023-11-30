package tests

import (
	"encoding/json"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"net/http"
	"reflect"
	"testing"

	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
)

func TestInsightsOrFunnelsGetAllQueriesGroupAnalysis(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	agent2, errCode := SetupAgentReturnDAO(getRandomEmail(), "+13425356")
	assert.Equal(t, http.StatusCreated, errCode)
	_, errCode = store.GetStore().CreateProjectAgentMappingWithDependencies(&model.ProjectAgentMapping{
		ProjectID: project.ID, AgentUUID: agent2.UUID})
	assert.Equal(t, http.StatusCreated, errCode)
	
	q1 := model.Query{
		From: 0,
		To:   0,
		EventsWithProperties: []model.QueryEventWithProperties{
			model.QueryEventWithProperties{
				Name:       "abcd",
				Properties: []model.QueryProperty{},
			},
			model.QueryEventWithProperties{
				Name:       "abcd1",
				Properties: []model.QueryProperty{},
			},
		},
		Class: model.QueryClassFunnel,

		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}
	q1InPostgresFormat, _ := U.EncodeStructTypeToPostgresJsonb(q1)
	_, errCode, errMsg := store.GetStore().CreateQuery(project.ID, &model.Queries{ProjectID: project.ID,
		Title: "abcd", Type: 2, CreatedBy: agent.UUID, Query: *q1InPostgresFormat,
		Settings: postgres.Jsonb{RawMessage: json.RawMessage(`{"size": 303}`)}})
	assert.Equal(t, "", errMsg)

	q1.Class = model.QueryClassInsights
	q2InPostgresFormat, _ := U.EncodeStructTypeToPostgresJsonb(q1)
	_, errCode, errMsg = store.GetStore().CreateQuery(project.ID, &model.Queries{ProjectID: project.ID,
		Title: "abcde", Type: 2, CreatedBy: agent.UUID, Query: *q2InPostgresFormat,
		Settings: postgres.Jsonb{RawMessage: json.RawMessage(`{"size": 303}`)}})
	assert.Equal(t, "", errMsg)

	dbQueries, _ := store.GetStore().GetALLQueriesWithProjectId(project.ID)
	for _, dbQuery := range dbQueries {
		var internalQuery model.Query
		U.DecodePostgresJsonbToStructType(&dbQuery.Query, &internalQuery)
		assert.NotEmpty(t, internalQuery.Class)
	}
}

func TestEventsGetAllQueriesGroupAnalysis(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	agent2, errCode := SetupAgentReturnDAO(getRandomEmail(), "+13425356")
	assert.Equal(t, http.StatusCreated, errCode)
	_, errCode = store.GetStore().CreateProjectAgentMappingWithDependencies(&model.ProjectAgentMapping{
		ProjectID: project.ID, AgentUUID: agent2.UUID})
	assert.Equal(t, http.StatusCreated, errCode)

	
		query1 := model.Query{
			From: 0,
			To:   0,
			EventsWithProperties: []model.QueryEventWithProperties{
				model.QueryEventWithProperties{
					Name: "abc",
					Properties: []model.QueryProperty{
					},
				},
			},
			Class: model.QueryClassEvents,

			Type:            model.QueryTypeEventsOccurrence,
			EventsCondition: model.EventCondAnyGivenEvent,
		}

		query2 := model.Query{
			From: 0,
			To:   0,
			EventsWithProperties: []model.QueryEventWithProperties{
				model.QueryEventWithProperties{
					Name: "abc1",
					Properties: []model.QueryProperty{
					},
				},
			},
			Class: model.QueryClassEvents,

			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondAnyGivenEvent,
		}

		queryGroup := model.QueryGroup{}
		queryGroup.Queries = make([]model.Query, 0)
		queryGroup.Queries = append(queryGroup.Queries, query1)
		queryGroup.Queries = append(queryGroup.Queries, query2)

		q1InPostgresFormat, _ := U.EncodeStructTypeToPostgresJsonb(queryGroup)
		_, errCode, errMsg := store.GetStore().CreateQuery(project.ID, &model.Queries{ProjectID: project.ID,
			Title: "abc", Type: 2, CreatedBy: agent.UUID, Query: *q1InPostgresFormat,
			Settings: postgres.Jsonb{RawMessage: json.RawMessage(`{"size": 303}`)}})
		assert.Equal(t, "", errMsg)
	
	dbQueries, _ := store.GetStore().GetALLQueriesWithProjectId(project.ID)
	for _, dbQuery := range dbQueries {
		var internalQuery model.QueryGroup
		U.DecodePostgresJsonbToStructType(&dbQuery.Query, &internalQuery)
		assert.NotEmpty(t, internalQuery.Queries[0].Class)
		assert.NotEmpty(t, internalQuery.Queries[1].Class)
	}

}

func TestModelQuery(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	agent2, errCode := SetupAgentReturnDAO(getRandomEmail(), "+13425356")
	assert.Equal(t, http.StatusCreated, errCode)
	_, errCode = store.GetStore().CreateProjectAgentMappingWithDependencies(&model.ProjectAgentMapping{
		ProjectID: project.ID, AgentUUID: agent2.UUID})
	assert.Equal(t, http.StatusCreated, errCode)

	var queryId int64
	t.Run("CreateQuery:SavedQuery:valid", func(t *testing.T) {
		rName1 := U.RandomString(5)
		query, errCode, errMsg := store.GetStore().CreateQuery(project.ID, &model.Queries{ProjectID: project.ID,
			Title: rName1, Type: 2, CreatedBy: agent.UUID, Query: postgres.Jsonb{RawMessage: json.RawMessage(`{}`)},
			Settings: postgres.Jsonb{RawMessage: json.RawMessage(`{"size": 303}`)}})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, query)
		assert.Empty(t, errMsg)
		queryId = query.ID
	})
	// No agentUUID for saved Query && empty title.
	t.Run("CreateQuery:SavedQuery:invalid", func(t *testing.T) {
		rName1 := U.RandomString(5)
		_, errCode, _ := store.GetStore().CreateQuery(project.ID, &model.Queries{ProjectID: project.ID,
			Title: rName1, Type: 2, CreatedBy: "", Query: postgres.Jsonb{RawMessage: json.RawMessage(`{}`)},
			Settings: postgres.Jsonb{RawMessage: json.RawMessage(`{"size": 30}`)}})
		assert.Equal(t, http.StatusBadRequest, errCode)

		_, errCode, _ = store.GetStore().CreateQuery(project.ID, &model.Queries{ProjectID: project.ID,
			Title: "", Type: 2, CreatedBy: agent.UUID, Query: postgres.Jsonb{RawMessage: json.RawMessage(`{}`)},
			Settings: postgres.Jsonb{RawMessage: json.RawMessage(`{"chart": "Line"}`)}})
		assert.Equal(t, http.StatusBadRequest, errCode)
	})
	// Get Query test
	query, errCode := store.GetStore().GetQueryWithQueryId(project.ID, queryId)
	assert.Equal(t, http.StatusFound, errCode)

	t.Run("UpdateSavedQuery:ValidForTitle", func(t *testing.T) {
		rName1 := U.RandomString(5)
		query1, errCode := store.GetStore().UpdateSavedQuery(project.ID, queryId, &model.Queries{Title: rName1,
			Settings: postgres.Jsonb{RawMessage: json.RawMessage(`{"size": 303}`)}})
		assert.Equal(t, http.StatusAccepted, errCode)
		assert.Equal(t, rName1, query1.Title)
		assert.NotEqual(t, query1.Title, query.Title)
		querySettings, _ := U.DecodePostgresJsonb(&query.Settings)
		query1Settings, _ := U.DecodePostgresJsonb(&query1.Settings)
		assert.True(t, reflect.DeepEqual(query1Settings, querySettings))
	})

	t.Run("UpdateSavedQuery:ValidForSetting", func(t *testing.T) {
		rName1 := U.RandomString(5)
		query1, errCode := store.GetStore().UpdateSavedQuery(project.ID, queryId, &model.Queries{Title: rName1,
			Settings: postgres.Jsonb{RawMessage: json.RawMessage(`{"size": 304}`)}})
		assert.Equal(t, http.StatusAccepted, errCode)
		assert.Equal(t, rName1, query1.Title)
		assert.NotEqual(t, query1.Title, query.Title)
		assert.NotEqual(t, query1.Settings, query.Settings)
		assert.NotEqual(t, string((query1.Settings).RawMessage), string((query.Settings).RawMessage))
	})

	t.Run("UpdateSavedQuery:Invalid", func(t *testing.T) {
		_, errCode := store.GetStore().UpdateSavedQuery(project.ID, queryId, &model.Queries{Title: ""})
		assert.Equal(t, http.StatusBadRequest, errCode)
	})

	// test delete query
	errCode, errMsg := store.GetStore().DeleteSavedQuery(project.ID, queryId)
	assert.Equal(t, errCode, http.StatusAccepted)
	assert.Empty(t, errMsg)

	//Check if deleted
	_, errCode = store.GetStore().GetQueryWithQueryId(project.ID, queryId)
	assert.Equal(t, http.StatusNotFound, errCode)

	// test search query
	rName1 := "Hello"
	query1, errCode, errMsg := store.GetStore().CreateQuery(project.ID, &model.Queries{ProjectID: project.ID,
		Title: rName1, Type: model.QueryTypeSavedQuery, CreatedBy: agent.UUID, Query: postgres.Jsonb{RawMessage: json.RawMessage(`{}`)},
		Settings: postgres.Jsonb{RawMessage: json.RawMessage(`{}`)}})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotNil(t, query1)
	assert.Empty(t, errMsg)

	rName2 := "World"
	query2, errCode, errMsg := store.GetStore().CreateQuery(project.ID, &model.Queries{ProjectID: project.ID,
		Title: rName2, Type: model.QueryTypeDashboardQuery, CreatedBy: agent.UUID, Query: postgres.Jsonb{RawMessage: json.RawMessage(`{}`)},
		Settings: postgres.Jsonb{RawMessage: json.RawMessage(`{}`)}})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotNil(t, query2)
	assert.Empty(t, errMsg)

	queries, errCode := store.GetStore().SearchQueriesWithProjectId(project.ID, "Hello")
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, 1, len(queries))

	queries, errCode = store.GetStore().SearchQueriesWithProjectId(project.ID, "o")
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, 2, len(queries))
}

func TestDeleteQuery(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	dashboardQuery, errCode, errMsg := store.GetStore().CreateQuery(project.ID, &model.Queries{
		ProjectID: project.ID,
		Type:      model.QueryTypeDashboardQuery,
		Query:     postgres.Jsonb{RawMessage: json.RawMessage(`{}`)},
	})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Empty(t, errMsg)
	assert.NotNil(t, dashboardQuery)

	savedQuery, errCode, errMsg := store.GetStore().CreateQuery(project.ID, &model.Queries{
		ProjectID: project.ID,
		Type:      model.QueryTypeSavedQuery,
		CreatedBy: agent.UUID,
		Title:     U.RandomString(5),
		Query:     postgres.Jsonb{RawMessage: json.RawMessage(`{}`)},
	})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Empty(t, errMsg)
	assert.NotNil(t, savedQuery)

	// Delete dashboard query should not delete saved type query.
	store.GetStore().DeleteDashboardQuery(project.ID, savedQuery.ID)
	query, errCode := store.GetStore().GetQueryWithQueryId(project.ID, savedQuery.ID)
	assert.NotEmpty(t, query)

	// Delete saved query should not delete dashboard type query.
	store.GetStore().DeleteSavedQuery(project.ID, dashboardQuery.ID)
	query, errCode = store.GetStore().GetQueryWithQueryId(project.ID, dashboardQuery.ID)
	assert.NotEmpty(t, query)

	// Should delete this time.
	store.GetStore().DeleteSavedQuery(project.ID, savedQuery.ID)
	query, errCode = store.GetStore().GetQueryWithQueryId(project.ID, savedQuery.ID)
	assert.Empty(t, query)
	assert.Equal(t, http.StatusNotFound, errCode)

	// Should delete this time.
	store.GetStore().DeleteDashboardQuery(project.ID, dashboardQuery.ID)
	query, errCode = store.GetStore().GetQueryWithQueryId(project.ID, dashboardQuery.ID)
	assert.Empty(t, query)
	assert.Equal(t, http.StatusNotFound, errCode)
}

func TestDeleteQueryWithDashboardUnit(t *testing.T) {
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

	dashboard, errCode := store.GetStore().CreateDashboard(project.ID, agent.UUID,
		&model.Dashboard{Name: U.RandomString(5), Type: model.DashboardTypeProjectVisible})
	assert.NotNil(t, dashboard)
	assert.Equal(t, http.StatusCreated, errCode)

	// Two Dashboard units with query type QueryTypeDashboardUnit.
	dashboardUnit1, errCode, errMsg := store.GetStore().CreateDashboardUnit(project.ID, agent.UUID,
		&model.DashboardUnit{DashboardId: dashboard.ID, Presentation: model.PresentationLine,
			QueryId: dashboardQuery.ID})
	assert.NotEmpty(t, dashboardUnit1)

	dashboardUnit2, errCode, errMsg := store.GetStore().CreateDashboardUnit(project.ID, agent.UUID,
		&model.DashboardUnit{DashboardId: dashboard.ID, Presentation: model.PresentationLine,
			QueryId: dashboardQuery.ID})
	assert.NotEmpty(t, dashboardUnit2)

	// Should not allow direct delete since units exists.
	errCode, errMsg = store.GetStore().DeleteDashboardQuery(project.ID, dashboardQuery.ID)
	assert.Equal(t, http.StatusNotAcceptable, errCode)
	query, errCode := store.GetStore().GetQueryWithQueryId(project.ID, dashboardQuery.ID)
	assert.NotEmpty(t, query)

	// On deleting one of the unit, should not delete the query.
	errCode = store.GetStore().DeleteDashboardUnit(project.ID, agent.UUID, dashboard.ID, dashboardUnit1.ID)
	assert.Equal(t, http.StatusAccepted, errCode)
	query, errCode = store.GetStore().GetQueryWithQueryId(project.ID, dashboardQuery.ID)
	assert.NotEmpty(t, query)

	// On deleting the other unit, it should not delete the undelying query.
	errCode = store.GetStore().DeleteDashboardUnit(project.ID, agent.UUID, dashboard.ID, dashboardUnit2.ID)
	assert.Equal(t, http.StatusAccepted, errCode)
	query, errCode = store.GetStore().GetQueryWithQueryId(project.ID, dashboardQuery.ID)
	assert.NotEmpty(t, query)
	assert.Equal(t, http.StatusFound, errCode)
}
