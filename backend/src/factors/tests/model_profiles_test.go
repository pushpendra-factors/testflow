package tests

import (
	"encoding/json"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"net/http"
	"testing"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
)

func TestProfiles(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)
	assert.NotNil(t, agent)
	projectID := project.ID
	// agentID := agent.UUID

	rCustomerUserId := U.RandomLowerAphaNumString(15)
	properties1 := postgres.Jsonb{RawMessage: json.RawMessage([]byte(`{"country": "india", "age": 30, "paid": true}`))}
	properties2 := postgres.Jsonb{RawMessage: json.RawMessage([]byte(`{"country": "us", "age": 20, "paid": true}`))}
	joinTime := time.Now().Unix()

	createUserID1, newUserErrorCode := store.GetStore().CreateUser(&model.User{ProjectId: projectID, CustomerUserId: rCustomerUserId, Properties: properties1, JoinTimestamp: joinTime, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, newUserErrorCode)
	assert.NotEqual(t, "", createUserID1)

	nextUserJoinTime := joinTime + 86400
	createUserID2, nextUserErrCode := store.GetStore().CreateUser(&model.User{ProjectId: projectID, Properties: properties2, JoinTimestamp: nextUserJoinTime, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, nextUserErrCode)
	assert.NotEqual(t, "", createUserID2)

	t.Run("No filters, no groupby", func(t *testing.T) {
		query := model.ProfileQuery{
			Type:     "all_users",
			Filters:  []model.QueryProperty{},
			GroupBys: []model.QueryGroupByProperty{},
			From:     joinTime - 100,
			To:       nextUserJoinTime + 100,
		}
		queryGroup := model.ProfileQueryGroup{
			Class:          "profiles",
			Queries:        []model.ProfileQuery{query},
			GlobalFilters:  []model.QueryProperty{},
			GlobalGroupBys: []model.QueryGroupByProperty{},
			From:           joinTime - 100,
			To:             nextUserJoinTime + 100,
		}
		result, statusCode := store.GetStore().RunProfilesGroupQuery(queryGroup.Queries, projectID)
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, float64(2), result.Results[0].Rows[0][1])
		assert.Equal(t, int(0), result.Results[0].Rows[0][0])
		assert.Equal(t, "query_index", result.Results[0].Headers[0])
		assert.Equal(t, "all_users", result.Results[0].Headers[1])
	})

	t.Run("1 filter, no groupby", func(t *testing.T) {
		query := model.ProfileQuery{
			Type: "all_users",
			Filters: []model.QueryProperty{
				{Type: "categorical", Property: "country", Operator: "equals", Value: "india", LogicalOp: "AND"},
			},
			GroupBys: []model.QueryGroupByProperty{},
			From:     joinTime - 100,
			To:       nextUserJoinTime + 100,
		}
		queryGroup := model.ProfileQueryGroup{
			Class:          "profiles",
			Queries:        []model.ProfileQuery{query},
			GlobalFilters:  []model.QueryProperty{},
			GlobalGroupBys: []model.QueryGroupByProperty{},
			From:           joinTime - 100,
			To:             nextUserJoinTime + 100,
		}
		result, statusCode := store.GetStore().RunProfilesGroupQuery(queryGroup.Queries, projectID)
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, float64(1), result.Results[0].Rows[0][1])
		assert.Equal(t, int(0), result.Results[0].Rows[0][0])
		assert.Equal(t, "query_index", result.Results[0].Headers[0])
		assert.Equal(t, "all_users", result.Results[0].Headers[1])
	})
	t.Run("joinTime check", func(t *testing.T) {
		query := model.ProfileQuery{
			Type:     "all_users",
			Filters:  []model.QueryProperty{},
			GroupBys: []model.QueryGroupByProperty{},
			From:     joinTime - 100,
			To:       joinTime + 100,
		}
		queryGroup := model.ProfileQueryGroup{
			Class:          "profiles",
			Queries:        []model.ProfileQuery{query},
			GlobalFilters:  []model.QueryProperty{},
			GlobalGroupBys: []model.QueryGroupByProperty{},
			From:           joinTime - 100,
			To:             joinTime + 100,
		}
		result, statusCode := store.GetStore().RunProfilesGroupQuery(queryGroup.Queries, projectID)
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, float64(1), result.Results[0].Rows[0][1])
		assert.Equal(t, int(0), result.Results[0].Rows[0][0])
		assert.Equal(t, "query_index", result.Results[0].Headers[0])
		assert.Equal(t, "all_users", result.Results[0].Headers[1])
	})

	t.Run("No filter, 1 group by", func(t *testing.T) {
		query := model.ProfileQuery{
			Type:     "all_users",
			Filters:  []model.QueryProperty{},
			GroupBys: []model.QueryGroupByProperty{{Entity: "user_g", Property: "country", Type: "categorical"}},
			From:     joinTime - 100,
			To:       nextUserJoinTime + 100,
		}
		queryGroup := model.ProfileQueryGroup{
			Class:          "profiles",
			Queries:        []model.ProfileQuery{query},
			GlobalFilters:  []model.QueryProperty{},
			GlobalGroupBys: []model.QueryGroupByProperty{},
			From:           joinTime - 100,
			To:             nextUserJoinTime + 100,
		}
		result, statusCode := store.GetStore().RunProfilesGroupQuery(queryGroup.Queries, projectID)
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, float64(1), result.Results[0].Rows[0][1])
		assert.Equal(t, int(0), result.Results[0].Rows[0][0])
		assert.Equal(t, "india", result.Results[0].Rows[0][2])
		assert.Equal(t, float64(1), result.Results[0].Rows[1][1])
		assert.Equal(t, int(0), result.Results[0].Rows[1][0])
		assert.Equal(t, "us", result.Results[0].Rows[1][2])
		assert.Equal(t, "query_index", result.Results[0].Headers[0])
		assert.Equal(t, "all_users", result.Results[0].Headers[1])
		assert.Equal(t, "country", result.Results[0].Headers[2])
	})

}
