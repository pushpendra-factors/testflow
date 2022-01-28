package tests

import (
	"encoding/json"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
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
		result, statusCode := store.GetStore().RunProfilesGroupQuery(queryGroup.Queries, projectID)
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
		result, statusCode := store.GetStore().RunProfilesGroupQuery(queryGroup.Queries, projectID)
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
		result, statusCode := store.GetStore().RunProfilesGroupQuery(queryGroup.Queries, projectID)
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
		result, statusCode := store.GetStore().RunProfilesGroupQuery(queryGroup.Queries, projectID)
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, float64(1), result.Results[0].Rows[0][1])
		assert.Equal(t, int(0), result.Results[0].Rows[0][0])
		assert.Equal(t, "india", result.Results[0].Rows[0][2])
		assert.Equal(t, float64(1), result.Results[0].Rows[1][1])
		assert.Equal(t, int(0), result.Results[0].Rows[1][0])
		assert.Equal(t, "us", result.Results[0].Rows[1][2])
		assert.Equal(t, "query_index", result.Results[0].Headers[0])
		assert.Equal(t, model.AliasAggr, result.Results[0].Headers[1])
		assert.Equal(t, "country", result.Results[0].Headers[2])
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

		result, errCode, _ := store.GetStore().ExecuteProfilesQuery(project.ID, query)
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

		result2, errCode, _ := store.GetStore().ExecuteProfilesQuery(project.ID, query2)
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

		result3, errCode, _ := store.GetStore().ExecuteProfilesQuery(project.ID, query3)
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

		result, errCode, _ := store.GetStore().ExecuteProfilesQuery(project.ID, query)
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

		result2, errCode, _ := store.GetStore().ExecuteProfilesQuery(project.ID, query2)
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

		result3, errCode, _ := store.GetStore().ExecuteProfilesQuery(project.ID, query3)
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
			GroupBys: []model.QueryGroupByProperty{
				model.QueryGroupByProperty{
					Entity:   model.PropertyEntityUser,
					Property: "$group_id",
					Type:     "categorical",
				},
			},
			GroupAnalysis: model.GROUP_NAME_HUBSPOT_COMPANY,
		}

		result, errCode, _ := store.GetStore().ExecuteProfilesQuery(project.ID, query)
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
			GroupBys: []model.QueryGroupByProperty{
				model.QueryGroupByProperty{
					Entity:   model.PropertyEntityUser,
					Property: "$group_id",
					Type:     "categorical",
				},
			},
			GroupAnalysis: model.GROUP_NAME_SALESFORCE_OPPORTUNITY,
		}

		result2, errCode, _ := store.GetStore().ExecuteProfilesQuery(project.ID, query2)
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
			GroupBys: []model.QueryGroupByProperty{
				model.QueryGroupByProperty{
					Entity:   model.PropertyEntityUser,
					Property: "$group_id",
					Type:     "categorical",
				},
			},
			GroupAnalysis: model.GROUP_NAME_HUBSPOT_DEAL,
		}

		result3, errCode, _ := store.GetStore().ExecuteProfilesQuery(project.ID, query3)
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
			GroupBys: []model.QueryGroupByProperty{
				model.QueryGroupByProperty{
					Entity:   model.PropertyEntityUser,
					Property: "$group_id",
					Type:     "categorical",
				},
			},
			GroupAnalysis: model.GROUP_NAME_SALESFORCE_ACCOUNT,
		}

		result4, errCode, _ := store.GetStore().ExecuteProfilesQuery(project.ID, query4)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result4)
		assert.Equal(t, model.AliasAggr, result4.Headers[0])
		assert.Equal(t, "$group_id", result4.Headers[1])
		assert.Equal(t, fmt.Sprintf("%d", group.ID), result4.Rows[0][1])
		assert.Equal(t, float64(len(sourceSalesforceUsers2)), result4.Rows[0][0])
	})
}
