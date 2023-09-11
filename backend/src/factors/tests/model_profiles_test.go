package tests

import (
	"encoding/json"
	C "factors/config"
	H "factors/handler"
	"factors/handler/helpers"
	IntHubspot "factors/integration/hubspot"
	"factors/model/model"
	M "factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestProfiles(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)
	assert.NotNil(t, agent)
	projectID := project.ID

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
			Type:          "web",
			Filters:       []model.QueryProperty{},
			GroupBys:      []model.QueryGroupByProperty{},
			From:          joinTime - 100,
			To:            nextUserJoinTime + 100,
			GroupAnalysis: "users",
		}
		queryGroup := model.ProfileQueryGroup{
			Class:          "profiles",
			Queries:        []model.ProfileQuery{query},
			GlobalFilters:  []model.QueryProperty{},
			GlobalGroupBys: []model.QueryGroupByProperty{},
			From:           joinTime - 100,
			To:             nextUserJoinTime + 100,
		}
		result, statusCode := store.GetStore().RunProfilesGroupQuery(queryGroup.Queries, projectID, C.EnableOptimisedFilterOnProfileQuery())
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, float64(2), result.Results[0].Rows[0][1])
		assert.Equal(t, int(0), result.Results[0].Rows[0][0])
		assert.Equal(t, "query_index", result.Results[0].Headers[0])
		assert.Equal(t, model.AliasAggr, result.Results[0].Headers[1])
	})

	t.Run("1 filter, no groupby", func(t *testing.T) {
		query := model.ProfileQuery{
			Type: "web",
			Filters: []model.QueryProperty{
				{Type: "categorical", Property: "country", Operator: "equals", Value: "india", LogicalOp: "AND"},
			},
			GroupBys:      []model.QueryGroupByProperty{},
			From:          joinTime - 100,
			To:            nextUserJoinTime + 100,
			GroupAnalysis: "users",
		}
		queryGroup := model.ProfileQueryGroup{
			Class:          "profiles",
			Queries:        []model.ProfileQuery{query},
			GlobalFilters:  []model.QueryProperty{},
			GlobalGroupBys: []model.QueryGroupByProperty{},
			From:           joinTime - 100,
			To:             nextUserJoinTime + 100,
		}
		result, statusCode := store.GetStore().RunProfilesGroupQuery(queryGroup.Queries, projectID, C.EnableOptimisedFilterOnProfileQuery())
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, float64(1), result.Results[0].Rows[0][1])
		assert.Equal(t, int(0), result.Results[0].Rows[0][0])
		assert.Equal(t, "query_index", result.Results[0].Headers[0])
		assert.Equal(t, model.AliasAggr, result.Results[0].Headers[1])
	})

	t.Run("joinTime check", func(t *testing.T) {
		query := model.ProfileQuery{
			Type:          "web",
			Filters:       []model.QueryProperty{},
			GroupBys:      []model.QueryGroupByProperty{},
			From:          joinTime - 100,
			To:            joinTime + 100,
			GroupAnalysis: "users",
		}
		queryGroup := model.ProfileQueryGroup{
			Class:          "profiles",
			Queries:        []model.ProfileQuery{query},
			GlobalFilters:  []model.QueryProperty{},
			GlobalGroupBys: []model.QueryGroupByProperty{},
			From:           joinTime - 100,
			To:             joinTime + 100,
		}
		result, statusCode := store.GetStore().RunProfilesGroupQuery(queryGroup.Queries, projectID, C.EnableOptimisedFilterOnProfileQuery())
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, float64(1), result.Results[0].Rows[0][1])
		assert.Equal(t, int(0), result.Results[0].Rows[0][0])
		assert.Equal(t, "query_index", result.Results[0].Headers[0])
		assert.Equal(t, model.AliasAggr, result.Results[0].Headers[1])
	})

	t.Run("No filter, 1 group by", func(t *testing.T) {
		query := model.ProfileQuery{
			Type:          "web",
			Filters:       []model.QueryProperty{},
			GroupBys:      []model.QueryGroupByProperty{{Entity: "user_g", Property: "country", Type: "categorical"}},
			From:          joinTime - 100,
			To:            nextUserJoinTime + 100,
			GroupAnalysis: "users",
		}
		queryGroup := model.ProfileQueryGroup{
			Class:          "profiles",
			Queries:        []model.ProfileQuery{query},
			GlobalFilters:  []model.QueryProperty{},
			GlobalGroupBys: []model.QueryGroupByProperty{},
			From:           joinTime - 100,
			To:             nextUserJoinTime + 100,
		}
		result, statusCode := store.GetStore().RunProfilesGroupQuery(queryGroup.Queries, projectID, C.EnableOptimisedFilterOnProfileQuery())
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, "query_index", result.Results[0].Headers[0])
		assert.Equal(t, int(0), result.Results[0].Rows[0][0])
		assert.Equal(t, int(0), result.Results[0].Rows[1][0])

		assert.Equal(t, float64(1), result.Results[0].Rows[0][1])
		assert.Equal(t, float64(1), result.Results[0].Rows[1][1])
		assert.Equal(t, model.AliasAggr, result.Results[0].Headers[1])

		assert.Equal(t, "country", result.Results[0].Headers[2])
		assert.True(t, result.Results[0].Rows[0][2] == "india" || result.Results[0].Rows[0][2] == "us")
		assert.True(t, result.Results[0].Rows[1][2] == "india" || result.Results[0].Rows[1][2] == "us")

	})

	t.Run("No filter, 1 group by with user id", func(t *testing.T) {
		query := model.ProfileQuery{
			Type:          "web",
			Filters:       []model.QueryProperty{},
			GroupBys:      []model.QueryGroupByProperty{{Entity: "user_g", Property: "$user_id", Type: "categorical"}},
			From:          joinTime - 100,
			To:            nextUserJoinTime + 100,
			GroupAnalysis: "users",
		}
		queryGroup := model.ProfileQueryGroup{
			Class:          "profiles",
			Queries:        []model.ProfileQuery{query},
			GlobalFilters:  []model.QueryProperty{},
			GlobalGroupBys: []model.QueryGroupByProperty{},
			From:           joinTime - 100,
			To:             nextUserJoinTime + 100,
		}
		result, statusCode := store.GetStore().RunProfilesGroupQuery(queryGroup.Queries, projectID, C.EnableOptimisedFilterOnProfileQuery())
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, "query_index", result.Results[0].Headers[0])
		assert.Equal(t, int(0), result.Results[0].Rows[0][0])
		assert.Equal(t, int(0), result.Results[0].Rows[1][0])

		assert.Equal(t, float64(1), result.Results[0].Rows[0][1])
		assert.Equal(t, float64(1), result.Results[0].Rows[1][1])
		assert.Equal(t, model.AliasAggr, result.Results[0].Headers[1])

		assert.Equal(t, "$user_id", result.Results[0].Headers[2])
		assert.Equal(t, 2, len(result.Results[0].Rows))
	})

	t.Run("No filter, 1 group by bucketed", func(t *testing.T) {
		query := model.ProfileQuery{
			Type:          "web",
			Filters:       []model.QueryProperty{},
			GroupBys:      []model.QueryGroupByProperty{{Entity: "user", Property: "age", Type: "numerical", GroupByType: "with_buckets"}},
			From:          joinTime - 100,
			To:            nextUserJoinTime + 100,
			GroupAnalysis: "users",
		}
		queryGroup := model.ProfileQueryGroup{
			Class:          "profiles",
			Queries:        []model.ProfileQuery{query},
			GlobalFilters:  []model.QueryProperty{},
			GlobalGroupBys: []model.QueryGroupByProperty{},
			From:           joinTime - 100,
			To:             nextUserJoinTime + 100,
		}
		result, statusCode := store.GetStore().RunProfilesGroupQuery(queryGroup.Queries, projectID, C.EnableOptimisedFilterOnProfileQuery())
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, float64(1), result.Results[0].Rows[0][1])
		assert.Equal(t, int(0), result.Results[0].Rows[0][0])
		assert.Equal(t, "20", result.Results[0].Rows[0][2])
		assert.Equal(t, float64(1), result.Results[0].Rows[1][1])
		assert.Equal(t, int(0), result.Results[0].Rows[1][0])
		assert.Equal(t, "30", result.Results[0].Rows[1][2])
		assert.Equal(t, "query_index", result.Results[0].Headers[0])
		assert.Equal(t, model.AliasAggr, result.Results[0].Headers[1])
		assert.Equal(t, "age", result.Results[0].Headers[2])
	})
}

func TestProfilesDateRangeQuery(t *testing.T) {
	project, newUser, _, err := SetupProjectUserEventNameReturnDAO()
	assert.Nil(t, err)

	t.Run("QueryWithDateRange", func(t *testing.T) {
		initialTimestamp := time.Now().AddDate(0, 0, -10).Unix()
		var finalTimestamp int64
		var users []model.User

		// create 10 more users
		for i := 0; i < 10; i++ {
			createdUserID, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
			assert.Equal(t, http.StatusCreated, errCode)
			user, errCode := store.GetStore().GetUser(project.ID, createdUserID)
			assert.Equal(t, http.StatusFound, errCode)
			assert.True(t, len(user.ID) > 30)
			users = append(users, *user)
		}
		users = append(users, *newUser)
		finalTimestamp = time.Now().Unix()

		// normal query to fetch users from initialTimestamp to finalTimestamp
		// since a total 11 users were created, the query should return count 11 in result
		query := model.ProfileQuery{
			Type:          "web",
			From:          initialTimestamp,
			To:            finalTimestamp,
			GroupAnalysis: "users",
		}

		result, errCode, _ := store.GetStore().ExecuteProfilesQuery(project.ID, query, C.EnableOptimisedFilterOnProfileQuery())
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		assert.Equal(t, float64(11), result.Rows[0][0])

		// update userproperties of the 11 users created above. set two random browsers in "$browser" property
		// execute breakdown (groupby) query on "$browser" property and validate respective counts
		browser1 := U.RandomString(5)
		browser2 := U.RandomString(5)
		for i := 0; i < 11; i++ {
			var browser string
			if i%2 == 0 {
				browser = browser1
			} else {
				browser = browser2
			}

			newProperties := &postgres.Jsonb{RawMessage: json.RawMessage([]byte(fmt.Sprintf(
				`{"country": "india", "age": 30.1, "paid": true, "$browser": "%s"}`, browser)))}
			_, status := store.GetStore().UpdateUserPropertiesV2(project.ID, users[i].ID, newProperties, time.Now().Unix(), "", "")
			assert.Equal(t, http.StatusAccepted, status)
		}
		finalTimestamp = time.Now().Unix()

		// group by query applied on property->'$browser'
		query2 := model.ProfileQuery{
			Type: "web",
			From: initialTimestamp,
			To:   finalTimestamp,
			GroupBys: []model.QueryGroupByProperty{
				model.QueryGroupByProperty{
					Entity:   model.PropertyEntityUser,
					Property: "$browser",
					Type:     "categorical",
				},
			},
			GroupAnalysis: "users",
		}

		result2, errCode, _ := store.GetStore().ExecuteProfilesQuery(project.ID, query2, C.EnableOptimisedFilterOnProfileQuery())
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result2)
		assert.Equal(t, model.AliasAggr, result2.Headers[0])
		assert.Equal(t, "$browser", result2.Headers[1])
		assert.Equal(t, browser2, result2.Rows[0][1])
		assert.Equal(t, browser1, result2.Rows[1][1])
		assert.Equal(t, float64(5), result2.Rows[0][0])
		assert.Equal(t, float64(6), result2.Rows[1][0])

		// add userId in userproperties of the 11 users created above. set userId in "$user_id" property
		// run breakdown query on "$user_id" property to validate the newly created userIds with the userIds returned in the result
		for i := 0; i < 11; i++ {
			newProperties := &postgres.Jsonb{RawMessage: json.RawMessage([]byte(fmt.Sprintf(
				`{"$user_id": "%s"}`, users[i].ID)))}
			_, status := store.GetStore().UpdateUserPropertiesV2(project.ID, users[i].ID, newProperties, time.Now().Unix(), "", "")
			assert.Equal(t, http.StatusAccepted, status)
		}
		finalTimestamp = time.Now().Unix()

		// group by query applied on property->'$user_id'
		query3 := model.ProfileQuery{
			Type: "web",
			From: initialTimestamp,
			To:   finalTimestamp,
			GroupBys: []model.QueryGroupByProperty{
				model.QueryGroupByProperty{
					Entity:   model.PropertyEntityUser,
					Property: "$user_id",
					Type:     "categorical",
				},
			},
			GroupAnalysis: "users",
		}

		result3, errCode, _ := store.GetStore().ExecuteProfilesQuery(project.ID, query3, C.EnableOptimisedFilterOnProfileQuery())
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result3)
		var createdUsers = make(map[string]bool)
		for i := 0; i < 11; i++ {
			createdUsers[users[i].ID] = true
		}

		for i := 0; i < 11; i++ {
			assert.Equal(t, true, createdUsers[result3.Rows[i][1].(string)])
		}

	})
}

func TestProfilesUserSourceQuery(t *testing.T) {
	project, newUser, _, err := SetupProjectUserEventNameReturnDAO()
	assert.Nil(t, err)

	t.Run("QueryWithUserSource", func(t *testing.T) {
		initialTimestamp := time.Now().AddDate(0, 0, -10).Unix()
		var finalTimestamp int64
		var users []model.User

		// create 10 more users
		for i := 0; i < 10; i++ {
			createdUserID, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID,
				Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
			assert.Equal(t, http.StatusCreated, errCode)
			user, errCode := store.GetStore().GetUser(project.ID, createdUserID)
			assert.Equal(t, http.StatusFound, errCode)
			assert.True(t, len(user.ID) > 30)
			users = append(users, *user)
		}
		users = append(users, *newUser)
		finalTimestamp = time.Now().Unix()

		// update user properties to add $source property = source of created user
		for i := 0; i < len(users); i++ {
			newProperties := &postgres.Jsonb{RawMessage: json.RawMessage([]byte(fmt.Sprintf(
				`{"$source": "%d"}`, *users[i].Source)))}
			_, status := store.GetStore().UpdateUserPropertiesV2(project.ID, users[i].ID, newProperties, time.Now().Unix(), "", "")
			assert.Equal(t, http.StatusAccepted, status)
		}

		// group query to fetch users from initialTimestamp to finalTimestamp
		// since a total 11 users were created, the query should return count 11 in result
		query := model.ProfileQuery{
			Type: "web",
			From: initialTimestamp,
			To:   finalTimestamp,
			GroupBys: []model.QueryGroupByProperty{
				model.QueryGroupByProperty{
					Entity:   model.PropertyEntityUser,
					Property: "$source",
					Type:     "categorical",
				},
			},
			GroupAnalysis: "users",
		}

		result, errCode, _ := store.GetStore().ExecuteProfilesQuery(project.ID, query, C.EnableOptimisedFilterOnProfileQuery())
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		assert.Equal(t, model.AliasAggr, result.Headers[0])
		assert.Equal(t, "$source", result.Headers[1])
		assert.Equal(t, fmt.Sprintf("%d", model.UserSourceWeb), result.Rows[0][1])
		assert.Equal(t, float64(len(users)), result.Rows[0][0])

		// create 10 more users with source hubspot
		var sourceHubspotUsers []model.User
		for i := 0; i < 10; i++ {
			createdUserID, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID,
				Source: model.GetRequestSourcePointer(model.UserSourceHubspot)})
			assert.Equal(t, http.StatusCreated, errCode)
			user, errCode := store.GetStore().GetUser(project.ID, createdUserID)
			assert.Equal(t, http.StatusFound, errCode)
			assert.True(t, len(user.ID) > 30)
			sourceHubspotUsers = append(sourceHubspotUsers, *user)
		}

		// update user properties to add $source property = source of created user
		for i := 0; i < len(sourceHubspotUsers); i++ {
			newProperties := &postgres.Jsonb{RawMessage: json.RawMessage([]byte(fmt.Sprintf(
				`{"$source": "%d"}`, *sourceHubspotUsers[i].Source)))}
			_, status := store.GetStore().UpdateUserPropertiesV2(project.ID, sourceHubspotUsers[i].ID, newProperties, time.Now().Unix(), "", "")
			assert.Equal(t, http.StatusAccepted, status)
		}
		// reset finalTimestamp
		finalTimestamp = time.Now().Unix()

		// group query to fetch users from initialTimestamp to finalTimestamp
		// total 21 users were created, 11 web users and 10 hubspot users
		// the following query should return count 10, since it is filtered for source = hubspot
		query2 := model.ProfileQuery{
			Type: "hubspot",
			From: initialTimestamp,
			To:   finalTimestamp,
			GroupBys: []model.QueryGroupByProperty{
				model.QueryGroupByProperty{
					Entity:   model.PropertyEntityUser,
					Property: "$source",
					Type:     "categorical",
				},
			},
			GroupAnalysis: "users",
		}

		result2, errCode, _ := store.GetStore().ExecuteProfilesQuery(project.ID, query2, C.EnableOptimisedFilterOnProfileQuery())
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result2)
		assert.Equal(t, model.AliasAggr, result2.Headers[0])
		assert.Equal(t, "$source", result2.Headers[1])
		assert.Equal(t, fmt.Sprintf("%d", model.UserSourceHubspot), result2.Rows[0][1])
		assert.Equal(t, float64(len(sourceHubspotUsers)), result2.Rows[0][0])

		// create 5 more users with source salesforce
		var sourceSalesforceUsers []model.User
		for i := 0; i < 5; i++ {
			createdUserID, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID,
				Source: model.GetRequestSourcePointer(model.UserSourceSalesforce)})
			assert.Equal(t, http.StatusCreated, errCode)
			user, errCode := store.GetStore().GetUser(project.ID, createdUserID)
			assert.Equal(t, http.StatusFound, errCode)
			assert.True(t, len(user.ID) > 30)
			sourceSalesforceUsers = append(sourceSalesforceUsers, *user)
		}

		// update user properties to add $source property = source of created user
		for i := 0; i < len(sourceSalesforceUsers); i++ {
			newProperties := &postgres.Jsonb{RawMessage: json.RawMessage([]byte(fmt.Sprintf(
				`{"$source": "%d"}`, *sourceSalesforceUsers[i].Source)))}
			_, status := store.GetStore().UpdateUserPropertiesV2(project.ID, sourceSalesforceUsers[i].ID, newProperties, time.Now().Unix(), "", "")
			assert.Equal(t, http.StatusAccepted, status)
		}
		// reset finalTimestamp
		finalTimestamp = time.Now().Unix()

		// group query to fetch users from initialTimestamp to finalTimestamp
		// total 26 users were created, 11 web users, 10 hubspot users and 5 salesforce users
		// the following query should return count 5, since it is filtered for source = salesforce
		query3 := model.ProfileQuery{
			Type: "salesforce",
			From: initialTimestamp,
			To:   finalTimestamp,
			GroupBys: []model.QueryGroupByProperty{
				model.QueryGroupByProperty{
					Entity:   model.PropertyEntityUser,
					Property: "$source",
					Type:     "categorical",
				},
			},
			GroupAnalysis: "users",
		}

		result3, errCode, _ := store.GetStore().ExecuteProfilesQuery(project.ID, query3, C.EnableOptimisedFilterOnProfileQuery())
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result3)
		assert.Equal(t, model.AliasAggr, result3.Headers[0])
		assert.Equal(t, "$source", result3.Headers[1])
		assert.Equal(t, fmt.Sprintf("%d", model.UserSourceSalesforce), result3.Rows[0][1])
		assert.Equal(t, float64(len(sourceSalesforceUsers)), result3.Rows[0][0])
	})
}

func TestProfilesGroupSupport(t *testing.T) {
	project, _, _, err := SetupProjectUserEventNameReturnDAO()
	assert.Nil(t, err)

	t.Run("QueryForProfilesGroupSupport", func(t *testing.T) {
		initialTimestamp := time.Now().AddDate(0, 0, -10).Unix()
		var finalTimestamp int64
		var sourceHubspotUsers1 []model.User

		// create new group with name = $hubspot_company
		group, status := store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_HUBSPOT_COMPANY, model.AllowedGroupNames)
		assert.Equal(t, http.StatusCreated, status)
		assert.NotNil(t, group)

		// create 10 group users, source = hubspot and group_name = $hubspot_company
		for i := 0; i < 10; i++ {
			createdUserID, errCode := store.GetStore().CreateGroupUser(&model.User{ProjectId: project.ID,
				Source: model.GetRequestSourcePointer(model.UserSourceHubspot)}, group.Name, fmt.Sprintf("%d", group.ID))
			assert.Equal(t, http.StatusCreated, errCode)
			user, errCode := store.GetStore().GetUser(project.ID, createdUserID)
			assert.Equal(t, http.StatusFound, errCode)
			assert.True(t, len(user.ID) > 30)
			sourceHubspotUsers1 = append(sourceHubspotUsers1, *user)
		}
		finalTimestamp = time.Now().Unix()

		// update user properties to add $group_id property = group.ID of created user
		for i := 0; i < len(sourceHubspotUsers1); i++ {
			newProperties := &postgres.Jsonb{RawMessage: json.RawMessage([]byte(fmt.Sprintf(
				`{"$group_id": "%d"}`, group.ID)))}
			_, status := store.GetStore().UpdateUserPropertiesV2(project.ID, sourceHubspotUsers1[i].ID, newProperties, time.Now().Unix(), "", "")
			assert.Equal(t, http.StatusAccepted, status)
		}

		// group query to fetch users from initialTimestamp to finalTimestamp
		// since a total 10 users were created for $hubspot_company group, the query should return count 10 in result
		query := model.ProfileQuery{
			From: initialTimestamp,
			To:   finalTimestamp,
			Type: "hubspot",
			GroupBys: []model.QueryGroupByProperty{
				model.QueryGroupByProperty{
					Entity:   model.PropertyEntityUser,
					Property: "$group_id",
					Type:     "categorical",
				},
			},
			GroupAnalysis: model.GROUP_NAME_HUBSPOT_COMPANY,
		}

		result, errCode, _ := store.GetStore().ExecuteProfilesQuery(project.ID, query, C.EnableOptimisedFilterOnProfileQuery())

		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		assert.Equal(t, model.AliasAggr, result.Headers[0])
		assert.Equal(t, "$group_id", result.Headers[1])
		assert.Equal(t, fmt.Sprintf("%d", group.ID), result.Rows[0][1])
		assert.Equal(t, float64(len(sourceHubspotUsers1)), result.Rows[0][0])

		// create new group with name = $salesforce_opportunity
		var sourceSalesforceUsers1 []model.User
		group, status = store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_SALESFORCE_OPPORTUNITY, model.AllowedGroupNames)
		assert.Equal(t, http.StatusCreated, status)
		assert.NotNil(t, group)

		// create 10 group users, source = salesforce and group_name = $salesforce_opportunity
		for i := 0; i < 10; i++ {
			createdUserID, errCode := store.GetStore().CreateGroupUser(&model.User{ProjectId: project.ID,
				Source: model.GetRequestSourcePointer(model.UserSourceSalesforce)}, group.Name, fmt.Sprintf("%d", group.ID))
			assert.Equal(t, http.StatusCreated, errCode)
			user, errCode := store.GetStore().GetUser(project.ID, createdUserID)
			assert.Equal(t, http.StatusFound, errCode)
			assert.True(t, len(user.ID) > 30)
			sourceSalesforceUsers1 = append(sourceSalesforceUsers1, *user)
		}
		finalTimestamp = time.Now().Unix()

		// update user properties to add $group_id property = group.ID of created user
		for i := 0; i < len(sourceSalesforceUsers1); i++ {
			newProperties := &postgres.Jsonb{RawMessage: json.RawMessage([]byte(fmt.Sprintf(
				`{"$group_id": "%d"}`, group.ID)))}
			_, status := store.GetStore().UpdateUserPropertiesV2(project.ID, sourceSalesforceUsers1[i].ID, newProperties, time.Now().Unix(), "", "")
			assert.Equal(t, http.StatusAccepted, status)
		}

		// group query to fetch users from initialTimestamp to finalTimestamp
		// since a total 10 users were created for $salesforce_opportunity group, the query should return count 10 in result
		query2 := model.ProfileQuery{
			From: initialTimestamp,
			To:   finalTimestamp,
			Type: "salesforce",
			GroupBys: []model.QueryGroupByProperty{
				model.QueryGroupByProperty{
					Entity:   model.PropertyEntityUser,
					Property: "$group_id",
					Type:     "categorical",
				},
			},
			GroupAnalysis: model.GROUP_NAME_SALESFORCE_OPPORTUNITY,
		}

		result2, errCode, _ := store.GetStore().ExecuteProfilesQuery(project.ID, query2, C.EnableOptimisedFilterOnProfileQuery())
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result2)
		assert.Equal(t, model.AliasAggr, result2.Headers[0])
		assert.Equal(t, "$group_id", result2.Headers[1])
		assert.Equal(t, fmt.Sprintf("%d", group.ID), result2.Rows[0][1])
		assert.Equal(t, float64(len(sourceSalesforceUsers1)), result2.Rows[0][0])

		// create new group with name = $hubspot_deal
		var sourceHubspotUsers2 []model.User
		group, status = store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_HUBSPOT_DEAL, model.AllowedGroupNames)
		assert.Equal(t, http.StatusCreated, status)
		assert.NotNil(t, group)

		// create 10 group users, source = hubspot and group_name = $hubspot_deal
		for i := 0; i < 10; i++ {
			createdUserID, errCode := store.GetStore().CreateGroupUser(&model.User{ProjectId: project.ID,
				Source: model.GetRequestSourcePointer(model.UserSourceHubspot)}, group.Name, fmt.Sprintf("%d", group.ID))
			assert.Equal(t, http.StatusCreated, errCode)
			user, errCode := store.GetStore().GetUser(project.ID, createdUserID)
			assert.Equal(t, http.StatusFound, errCode)
			assert.True(t, len(user.ID) > 30)
			sourceHubspotUsers2 = append(sourceHubspotUsers2, *user)
		}
		finalTimestamp = time.Now().Unix()

		// update user properties to add $group_id property = group.ID of created user
		for i := 0; i < len(sourceHubspotUsers2); i++ {
			newProperties := &postgres.Jsonb{RawMessage: json.RawMessage([]byte(fmt.Sprintf(
				`{"$group_id": "%d"}`, group.ID)))}
			_, status := store.GetStore().UpdateUserPropertiesV2(project.ID, sourceHubspotUsers2[i].ID, newProperties, time.Now().Unix(), "", "")
			assert.Equal(t, http.StatusAccepted, status)
		}

		// group query to fetch users from initialTimestamp to finalTimestamp
		// since a total 10 users were created for $hubspot_deal group, the query should return count 10 in result
		query3 := model.ProfileQuery{
			From: initialTimestamp,
			To:   finalTimestamp,
			Type: "hubspot",
			GroupBys: []model.QueryGroupByProperty{
				model.QueryGroupByProperty{
					Entity:   model.PropertyEntityUser,
					Property: "$group_id",
					Type:     "categorical",
				},
			},
			GroupAnalysis: model.GROUP_NAME_HUBSPOT_DEAL,
		}

		result3, errCode, _ := store.GetStore().ExecuteProfilesQuery(project.ID, query3, C.EnableOptimisedFilterOnProfileQuery())
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result3)
		assert.Equal(t, model.AliasAggr, result3.Headers[0])
		assert.Equal(t, "$group_id", result3.Headers[1])
		assert.Equal(t, fmt.Sprintf("%d", group.ID), result3.Rows[0][1])
		assert.Equal(t, float64(len(sourceHubspotUsers2)), result3.Rows[0][0])

		// create new group with name = $salesforce_account
		var sourceSalesforceUsers2 []model.User
		group, status = store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_SALESFORCE_ACCOUNT, model.AllowedGroupNames)
		assert.Equal(t, http.StatusCreated, status)
		assert.NotNil(t, group)

		// create 10 group users, source = salesforce and group_name = $salesforce_account
		for i := 0; i < 10; i++ {
			createdUserID, errCode := store.GetStore().CreateGroupUser(&model.User{ProjectId: project.ID,
				Source: model.GetRequestSourcePointer(model.UserSourceSalesforce)}, group.Name, fmt.Sprintf("%d", group.ID))
			assert.Equal(t, http.StatusCreated, errCode)
			user, errCode := store.GetStore().GetUser(project.ID, createdUserID)
			assert.Equal(t, http.StatusFound, errCode)
			assert.True(t, len(user.ID) > 30)
			sourceSalesforceUsers2 = append(sourceSalesforceUsers2, *user)
		}
		finalTimestamp = time.Now().Unix()

		// update user properties to add $group_id property = group.ID of created user
		for i := 0; i < len(sourceSalesforceUsers2); i++ {
			newProperties := &postgres.Jsonb{RawMessage: json.RawMessage([]byte(fmt.Sprintf(
				`{"$group_id": "%d"}`, group.ID)))}
			_, status := store.GetStore().UpdateUserPropertiesV2(project.ID, sourceSalesforceUsers2[i].ID, newProperties, time.Now().Unix(), "", "")
			assert.Equal(t, http.StatusAccepted, status)
		}

		// group query to fetch users from initialTimestamp to finalTimestamp
		// since a total 10 users were created for $salesforce_account group, the query should return count 10 in result
		query4 := model.ProfileQuery{
			From: initialTimestamp,
			To:   finalTimestamp,
			Type: "salesforce",
			GroupBys: []model.QueryGroupByProperty{
				model.QueryGroupByProperty{
					Entity:   model.PropertyEntityUser,
					Property: "$group_id",
					Type:     "categorical",
				},
			},
			GroupAnalysis: model.GROUP_NAME_SALESFORCE_ACCOUNT,
		}

		result4, errCode, _ := store.GetStore().ExecuteProfilesQuery(project.ID, query4, C.EnableOptimisedFilterOnProfileQuery())
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result4)
		assert.Equal(t, model.AliasAggr, result4.Headers[0])
		assert.Equal(t, "$group_id", result4.Headers[1])
		assert.Equal(t, fmt.Sprintf("%d", group.ID), result4.Rows[0][1])
		assert.Equal(t, float64(len(sourceSalesforceUsers2)), result4.Rows[0][0])
	})
}

func sendProfilesQueryReq(r *gin.Engine, projectId int64, agent *M.Agent, profileQuery model.ProfileQueryGroup) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}

	rb := C.NewRequestBuilderWithPrefix(http.MethodPost, fmt.Sprintf("/projects/%d/profiles/query", projectId)).
		WithPostParams(profileQuery).
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

func TestProfilesUsersPropertyValueLabels(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)
	assert.NotNil(t, agent)

	startTimestamp := time.Now().Unix()

	hubspotDisplayNameLabels := map[string]string{
		"OFFFLINE":       "Offline",
		"PAID_SEARCH":    "Paid Search",
		"DIRECT_TRAFFIC": "Direct Traffic",
		"ORGANIC_SEARCH": "Organic Search",
		"SOCIAL_MEDIA":   "Social Media",
	}

	for value, label := range hubspotDisplayNameLabels {
		status := store.GetStore().CreateOrUpdateDisplayNameLabel(project.ID, "hubspot", "$hubspot_contact_hs_latest_source", value, label)
		assert.Equal(t, http.StatusCreated, status)
	}

	createdUserID1, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID1)

	createdUserID2, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, JoinTimestamp: startTimestamp, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID2)

	// create new hubspot document
	jsonContactModel := `{
		"vid": %d,
		"addedAt": %d,
		"properties": {
		  	"firstname": { "value": "%s" },
		  	"lastname": { "value": "%s" },
		  	"lastmodifieddate": { "value": "%d" },
			"hs_latest_source": { "value": "%s" }
		},
		"identity-profiles": [
			{
				"vid": %d,
				"identities": [
					{
					  "type": "EMAIL",
					  "value": "%s"
					},
					{
						"type": "LEAD_GUID",
						"value": "%s"
					}
				]
			}
		]
	}`

	latestSources := []string{"ORGANIC_SEARCH", "DIRECT_TRAFFIC", "PAID_SOCIAL"}
	hubspotDocuments := make([]*model.HubspotDocument, 0)
	for i := 0; i < len(latestSources); i++ {
		documentID := i
		createdTime := startTimestamp*1000 + int64(i*100)
		updatedTime := createdTime + 200
		cuid := U.RandomString(10)
		leadGUID := U.RandomString(15)
		jsonContact := fmt.Sprintf(jsonContactModel, documentID, createdTime, "Sample", fmt.Sprintf("Test %d", i), updatedTime, latestSources[i], documentID, cuid, leadGUID)

		document := model.HubspotDocument{
			TypeAlias: model.HubspotDocumentTypeNameContact,
			Value:     &postgres.Jsonb{json.RawMessage(jsonContact)},
		}
		hubspotDocuments = append(hubspotDocuments, &document)
	}

	status := store.GetStore().CreateHubspotDocumentInBatch(project.ID, model.HubspotDocumentTypeContact, hubspotDocuments, 3)
	assert.Equal(t, http.StatusCreated, status)

	// execute sync job
	allStatus, _ := IntHubspot.Sync(project.ID, 1, time.Now().Unix(), nil, "", 50, 3)
	for i := range allStatus {
		assert.Equal(t, U.CRM_SYNC_STATUS_SUCCESS, allStatus[i].Status)
	}

	profileQueryGroup := model.ProfileQueryGroup{
		Class: model.QueryClassProfiles,
		Queries: []model.ProfileQuery{
			model.ProfileQuery{
				Type: "hubspot",
			},
		},
		GlobalGroupBys: []model.QueryGroupByProperty{
			model.QueryGroupByProperty{
				Entity:   model.PropertyEntityUser,
				Property: "$hubspot_contact_hs_latest_source",
				Type:     "categorical",
			},
		},
		From:          startTimestamp,
		To:            startTimestamp + 500,
		GroupAnalysis: "users",
	}

	C.GetConfig().EnableSyncReferenceFieldsByProjectID = ""
	w := sendProfilesQueryReq(r, project.ID, agent, profileQueryGroup)
	jsonResponse, err := ioutil.ReadAll(w.Body)
	assert.Nil(t, err)
	var resultGroup model.ResultGroup
	json.Unmarshal(jsonResponse, &resultGroup)
	assert.Equal(t, http.StatusOK, w.Code)

	results := resultGroup.Results
	assert.NotNil(t, results)
	assert.Equal(t, 1, len(results))

	assert.Equal(t, "query_index", results[0].Headers[0])
	for i := range results[0].Rows {
		assert.Equal(t, float64(0), results[0].Rows[i][0])
	}

	assert.Equal(t, "$hubspot_contact_hs_latest_source", results[0].Headers[2])
	for i := range results[0].Rows {
		assert.Contains(t, latestSources, results[0].Rows[i][2])
	}

	assert.Equal(t, "aggregate", results[0].Headers[1])
	for i := range results[0].Rows {
		assert.Equal(t, float64(1), results[0].Rows[i][1])
	}

	C.GetConfig().EnableSyncReferenceFieldsByProjectID = "*"
	w = sendProfilesQueryReq(r, project.ID, agent, profileQueryGroup)
	jsonResponse, err = ioutil.ReadAll(w.Body)
	assert.Nil(t, err)
	resultFromCache := struct {
		ResultGroup model.ResultGroup `json:"result"`
	}{}
	json.Unmarshal(jsonResponse, &resultFromCache)
	assert.Equal(t, http.StatusOK, w.Code)

	results = resultFromCache.ResultGroup.Results
	assert.NotNil(t, results)
	assert.Equal(t, 1, len(results))

	assert.Equal(t, "query_index", results[0].Headers[0])
	for i := range results[0].Rows {
		assert.Equal(t, float64(0), results[0].Rows[i][0])
	}

	assert.Equal(t, "$hubspot_contact_hs_latest_source", results[0].Headers[2])
	for i := range results[0].Rows {
		assert.Contains(t, []string{"Organic Search", "Direct Traffic", "PAID_SOCIAL"}, results[0].Rows[i][2])
	}

	assert.Equal(t, "aggregate", results[0].Headers[1])
	for i := range results[0].Rows {
		assert.Equal(t, float64(1), results[0].Rows[i][1])
	}
}

func TestProfilesHubspotDealsPropertyValueLabels(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)

	startTimestamp := time.Now().Unix()

	hubspotDisplayNameLabels := map[string]string{
		"OFFFLINE":       "Offline",
		"PAID_SEARCH":    "Paid Search",
		"DIRECT_TRAFFIC": "Direct Traffic",
		"ORGANIC_SEARCH": "Organic Search",
		"SOCIAL_MEDIA":   "Social Media",
	}

	for value, label := range hubspotDisplayNameLabels {
		status := store.GetStore().CreateOrUpdateDisplayNameLabel(project.ID, "hubspot", "$hubspot_deal_latest_source", value, label)
		assert.Equal(t, http.StatusCreated, status)
	}

	// create new hubspot document
	jsonDealModel := `{
		"dealId": %d,
		"properties": {
			"amount": { "value": "%d" },
			"createdate": { "value": "%d" },
			"hs_createdate": { "value": "%d" },
		  	"dealname": { "value": "%s" },
			"latest_source": { "value": "%s" },
		  	"hs_lastmodifieddate": { "value": "%d" }
		}
	}`

	latestSources := []string{"ORGANIC_SEARCH", "DIRECT_TRAFFIC", "PAID_SOCIAL"}
	hubspotDocuments := make([]*model.HubspotDocument, 0)
	for i := 0; i < len(latestSources); i++ {
		documentID := i
		createdTime := startTimestamp*1000 + int64(i*100)
		updatedTime := createdTime + 200
		amount := U.RandomIntInRange(1000, 2000)
		jsonDeal := fmt.Sprintf(jsonDealModel, documentID, amount, createdTime, createdTime, fmt.Sprintf("Dealname %d", i), latestSources[i], updatedTime)

		document := model.HubspotDocument{
			TypeAlias: model.HubspotDocumentTypeNameDeal,
			Value:     &postgres.Jsonb{json.RawMessage(jsonDeal)},
		}
		hubspotDocuments = append(hubspotDocuments, &document)
	}

	status := store.GetStore().CreateHubspotDocumentInBatch(project.ID, model.HubspotDocumentTypeDeal, hubspotDocuments, 3)
	assert.Equal(t, http.StatusCreated, status)

	// execute sync job
	allStatus, _ := IntHubspot.Sync(project.ID, 1, time.Now().Unix(), nil, "", 50, 3)
	for i := range allStatus {
		assert.Equal(t, U.CRM_SYNC_STATUS_SUCCESS, allStatus[i].Status)
	}

	profileQueryGroup := model.ProfileQueryGroup{
		Class: model.QueryClassProfiles,
		Queries: []model.ProfileQuery{
			model.ProfileQuery{
				Type: "hubspot",
			},
		},
		GlobalGroupBys: []model.QueryGroupByProperty{
			model.QueryGroupByProperty{
				Entity:   model.PropertyEntityUser,
				Property: "$hubspot_deal_latest_source",
				Type:     "categorical",
			},
		},
		From:          startTimestamp,
		To:            startTimestamp + 500,
		GroupAnalysis: "$hubspot_deal",
	}

	C.GetConfig().EnableSyncReferenceFieldsByProjectID = ""
	w := sendProfilesQueryReq(r, project.ID, agent, profileQueryGroup)
	jsonResponse, err := ioutil.ReadAll(w.Body)
	assert.Nil(t, err)
	var resultGroup model.ResultGroup
	json.Unmarshal(jsonResponse, &resultGroup)
	assert.Equal(t, http.StatusOK, w.Code)

	results := resultGroup.Results
	assert.NotNil(t, results)
	assert.Equal(t, 1, len(results))

	assert.Equal(t, "query_index", results[0].Headers[0])
	for i := range results[0].Rows {
		assert.Equal(t, float64(0), results[0].Rows[i][0])
	}

	assert.Equal(t, "$hubspot_deal_latest_source", results[0].Headers[2])
	for i := range results[0].Rows {
		assert.Contains(t, latestSources, results[0].Rows[i][2])
	}

	assert.Equal(t, "aggregate", results[0].Headers[1])
	for i := range results[0].Rows {
		assert.Equal(t, float64(1), results[0].Rows[i][1])
	}

	C.GetConfig().EnableSyncReferenceFieldsByProjectID = "*"
	w = sendProfilesQueryReq(r, project.ID, agent, profileQueryGroup)
	jsonResponse, err = ioutil.ReadAll(w.Body)
	assert.Nil(t, err)
	resultFromCache := struct {
		ResultGroup model.ResultGroup `json:"result"`
	}{}
	json.Unmarshal(jsonResponse, &resultFromCache)
	assert.Equal(t, http.StatusOK, w.Code)

	results = resultFromCache.ResultGroup.Results
	assert.NotNil(t, results)
	assert.Equal(t, 1, len(results))

	assert.Equal(t, "query_index", results[0].Headers[0])
	for i := range results[0].Rows {
		assert.Equal(t, float64(0), results[0].Rows[i][0])
	}

	assert.Equal(t, "$hubspot_deal_latest_source", results[0].Headers[2])
	for i := range results[0].Rows {
		assert.Contains(t, []string{"Organic Search", "Direct Traffic", "PAID_SOCIAL"}, results[0].Rows[i][2])
	}

	assert.Equal(t, "aggregate", results[0].Headers[1])
	for i := range results[0].Rows {
		assert.Equal(t, float64(1), results[0].Rows[i][1])
	}
}
