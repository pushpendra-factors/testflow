package tests

import (
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSampleWebsiteAggregation(t *testing.T) {

	agent, errCode := SetupAgentReturnDAO(getRandomEmail(), "+13425356")
	assert.Equal(t, http.StatusCreated, errCode)
	billingAccount, errCode := store.GetStore().GetBillingAccountByAgentUUID(agent.UUID)
	assert.Equal(t, http.StatusFound, errCode)

	// Test successful create project.
	projectName := U.RandomLowerAphaNumString(15)
	project, errCode := store.GetStore().CreateProjectWithDependencies(&model.Project{Name: projectName}, agent.UUID, model.ADMIN, billingAccount.ID, true)
	assert.Equal(t, http.StatusCreated, errCode)

	date1 := int64(1672534800)
	record1 := model.WebsiteAggregation{ProjectID: project.ID, TimestampAtDay: date1, EventName: "session", EventType: "session", City: "AB", CountOfRecords: 3, SpentTime: 100.0}
	record2 := model.WebsiteAggregation{ProjectID: project.ID, TimestampAtDay: date1, EventName: "session", EventType: "session", City: "CD", CountOfRecords: 4, SpentTime: 200.0}
	record3 := model.WebsiteAggregation{ProjectID: project.ID, TimestampAtDay: date1, EventName: "abc.com", EventType: "page_view", City: "AB", CountOfRecords: 3, SpentTime: 100.0}
	record4 := model.WebsiteAggregation{ProjectID: project.ID, TimestampAtDay: date1, EventName: "abc.com", EventType: "page_view", City: "CD", CountOfRecords: 4, SpentTime: 200.0}

	_, errMsg, statusCode := store.GetStore().CreateWebsiteAggregation(record1)
	assert.Equal(t, http.StatusCreated, statusCode)
	assert.Equal(t, "", errMsg)

	_, errMsg, statusCode = store.GetStore().CreateWebsiteAggregation(record2)
	assert.Equal(t, http.StatusCreated, statusCode)
	assert.Equal(t, "", errMsg)

	_, errMsg, statusCode = store.GetStore().CreateWebsiteAggregation(record3)
	assert.Equal(t, http.StatusCreated, statusCode)
	assert.Equal(t, "", errMsg)

	_, errMsg, statusCode = store.GetStore().CreateWebsiteAggregation(record4)
	assert.Equal(t, http.StatusCreated, statusCode)
	assert.Equal(t, "", errMsg)

	t.Run("test predefined website aggregation for session", func(t *testing.T) {

		query1 := model.PredefWebsiteAggregationQuery{
			Metrics: []model.PredefinedMetric{
				{Name: model.PredefTotalSessions, DisplayName: model.PredefDispTotalSessions},
			},
			GroupBy: model.PredefinedGroupBy{
				Name: model.PredefPropCity, DisplayName: model.PredefPropDispCity,
			},
			Filters: []model.PredefinedFilter{
				{PropertyName: model.PredefPropCity, PropertyDataType: "categorical", Condition: "equals", Value: "AB", LogicalOp: ""},
			},
			GroupByTimestamp:  "",
			Timezone:          "Asia/Kolkata",
			From:              1669833000,
			To: 			   1675189800,
			InternalEventType: model.PredefEventTypeSession,
			WidgetInternalID:  2,
			WidgetName:        model.PredefWidUtmParams,
		}

		query2 := model.PredefWebsiteAggregationQuery{}
		U.DeepCopy(&query1, &query2)

		query2.GroupByTimestamp = "date"

		queries := []model.PredefWebsiteAggregationQuery{query1, query2}
		result, statusCode, _ := store.GetStore().ExecuteQueryGroupForPredefinedWebsiteAggregation(project.ID, model.PredefWebsiteAggregationQueryGroup{
			Class:   "predefined_dashboard",
			Queries: queries,
		})
		assert.Equal(t, http.StatusOK, statusCode)

		assert.Equal(t, 2, len(result[0].Headers))
		assert.Equal(t, 1, len(result[0].Rows))
		assert.Equal(t, float64(3), result[0].Rows[0][1].(float64))

		assert.Equal(t, 3, len(result[1].Headers))
		assert.Equal(t, 63, len(result[1].Rows))
		assert.Equal(t, float64(3), result[1].Rows[31][2].(float64))

	})

	t.Run("test predefined website aggregation for page views", func(t *testing.T) {

		query1 := model.PredefWebsiteAggregationQuery{
			Metrics: []model.PredefinedMetric{
				{Name: model.PredefTotalPageViews, DisplayName: model.PredefDispTotalPageViews},
			},
			GroupBy: model.PredefinedGroupBy{
				Name: model.PredefPropTopPage, DisplayName: model.PredefPropDispTopPage,
			},
			Filters: []model.PredefinedFilter{
				{PropertyName: model.PredefPropCity, PropertyDataType: "categorical", Condition: "equals", Value: "AB", LogicalOp: ""},
			},
			GroupByTimestamp:  "",
			Timezone:          "Asia/Kolkata",
			From:              1669833000,
			To:                1675189800,
			InternalEventType: model.PredefEventTypePageViews,
			WidgetInternalID:  4,
			WidgetName:        model.PredefWidPageView,
		}

		query2 := model.PredefWebsiteAggregationQuery{}
		U.DeepCopy(&query1, &query2)

		query2.GroupByTimestamp = "date"

		queries := []model.PredefWebsiteAggregationQuery{query1, query2}
		result, statusCode, _ := store.GetStore().ExecuteQueryGroupForPredefinedWebsiteAggregation(project.ID, model.PredefWebsiteAggregationQueryGroup{
			Class:   "predefined_dashboard",
			Queries: queries,
		})
		assert.Equal(t, http.StatusOK, statusCode)

		assert.Equal(t, 2, len(result[0].Headers))
		assert.Equal(t, 1, len(result[0].Rows))
		assert.Equal(t, float64(3), result[0].Rows[0][1].(float64))

		assert.Equal(t, 3, len(result[1].Headers))
		assert.Equal(t, 63, len(result[1].Rows))
		assert.Equal(t, float64(3), result[1].Rows[31][2].(float64))

	})

	t.Run("test predefined website aggregation for empty rows with group by", func(t *testing.T) {

		query1 := model.PredefWebsiteAggregationQuery{
			Metrics: []model.PredefinedMetric{
				{Name: model.PredefTotalSessions, DisplayName: model.PredefDispTotalSessions},
				{Name: model.PredefAvgSessionDuration, DisplayName: model.PredefAvgSessionDuration},
			},
			GroupBy: model.PredefinedGroupBy{
				Name: model.PredefPropCity, DisplayName: model.PredefPropDispCity,
			},
			Filters: []model.PredefinedFilter{
				{PropertyName: model.PredefPropCity, PropertyDataType: "categorical", Condition: "equals", Value: "AB", LogicalOp: ""},
			},
			GroupByTimestamp:  	"",
			Timezone:          	"Asia/Kolkata",
			From:              	1669833000,
			To: 				1669833000 + 90400,
			InternalEventType: model.PredefEventTypeSession,
			WidgetInternalID:  2,
			WidgetName:        model.PredefWidUtmParams,
		}

		query2 := model.PredefWebsiteAggregationQuery{}
		U.DeepCopy(&query1, &query2)

		query2.GroupByTimestamp = "date"

		queries := []model.PredefWebsiteAggregationQuery{query1, query2}
		result, statusCode, _ := store.GetStore().ExecuteQueryGroupForPredefinedWebsiteAggregation(project.ID, model.PredefWebsiteAggregationQueryGroup{
			Class:   "predefined_dashboard",
			Queries: queries,
		})
		assert.Equal(t, http.StatusOK, statusCode)

		assert.Equal(t, 3, len(result[0].Headers))
		assert.Equal(t, 1, len(result[0].Rows))
		assert.Equal(t, 0, result[0].Rows[0][1].(int))
		assert.Equal(t, 0, result[0].Rows[0][2].(int))

		assert.Equal(t, 4, len(result[1].Headers))
		assert.Equal(t, 2, len(result[1].Rows))
		assert.Equal(t, 0, result[1].Rows[0][2].(int))
		assert.Equal(t, 0, result[1].Rows[0][3].(int))
		assert.Equal(t, 0, result[1].Rows[1][2].(int))
		assert.Equal(t, 0, result[1].Rows[1][3].(int))

	})

	t.Run("test predefined website aggregation for empty rows without group by", func(t *testing.T) {

		query1 := model.PredefWebsiteAggregationQuery{
			Metrics: []model.PredefinedMetric{
				{Name: model.PredefTotalSessions, DisplayName: model.PredefDispTotalSessions},
				{Name: model.PredefAvgSessionDuration, DisplayName: model.PredefAvgSessionDuration},
			},
			GroupBy: model.PredefinedGroupBy{
			},
			Filters: []model.PredefinedFilter{
				{PropertyName: model.PredefPropCity, PropertyDataType: "categorical", Condition: "equals", Value: "AB", LogicalOp: ""},
			},
			GroupByTimestamp:  	"",
			Timezone:          	"Asia/Kolkata",
			From:              	1669833000,
			To: 				1669833000 + 90400,
			InternalEventType: model.PredefEventTypeSession,
			WidgetInternalID:  1,
			WidgetName:        model.PredefWidUtmParams,
		}

		query2 := model.PredefWebsiteAggregationQuery{}
		U.DeepCopy(&query1, &query2)

		query2.GroupByTimestamp = "date"

		queries := []model.PredefWebsiteAggregationQuery{query1, query2}
		result, statusCode, _ := store.GetStore().ExecuteQueryGroupForPredefinedWebsiteAggregation(project.ID, model.PredefWebsiteAggregationQueryGroup{
			Class:   "predefined_dashboard",
			Queries: queries,
		})
		assert.Equal(t, http.StatusOK, statusCode)

		assert.Equal(t, 2, len(result[0].Headers))
		assert.Equal(t, 1, len(result[0].Rows))
		assert.Equal(t, 0, result[0].Rows[0][0].(int))
		assert.Equal(t, 0, result[0].Rows[0][1].(int))

		assert.Equal(t, 3, len(result[1].Headers))
		assert.Equal(t, 2, len(result[1].Rows))
		assert.Equal(t, 0, result[1].Rows[0][1].(int))
		assert.Equal(t, 0, result[1].Rows[0][2].(int))
		assert.Equal(t, 0, result[1].Rows[1][1].(int))
		assert.Equal(t, 0, result[1].Rows[0][2].(int))
	})
}

func TestPredefWebAggDashboard(t *testing.T) {
	agent, errCode := SetupAgentReturnDAO(getRandomEmail(), "+13425356")
	assert.Equal(t, http.StatusCreated, errCode)
	billingAccount, errCode := store.GetStore().GetBillingAccountByAgentUUID(agent.UUID)
	assert.Equal(t, http.StatusFound, errCode)

	// Test successful create project.
	projectName := U.RandomLowerAphaNumString(15)
	project, errCode := store.GetStore().CreateProjectWithDependencies(&model.Project{Name: projectName}, agent.UUID, model.ADMIN, billingAccount.ID, true)
	assert.Equal(t, http.StatusCreated, errCode)

	statusCode := store.GetStore().CreatePredefWebAggDashboardIfNotExists(project.ID)
	assert.Equal(t, statusCode, http.StatusCreated)

	statusCode2 := store.GetStore().CreatePredefWebAggDashboardIfNotExists(project.ID)
	assert.Equal(t, statusCode2, http.StatusFound)

}

func TestDupReportCreation(t *testing.T) {
	j1 := getDefaultJobReport()
	j2 := getDuplicateJobReportWithoutSuccessKeys(j1)
	log.WithField("j2", j2).Warn("duplicated job")
}

func getDefaultJobReport() map[string]map[string]interface{} {
	jobReport := make(map[string]map[string]interface{})
	jobReport["success"] = make(map[string]interface{})
	jobReport["success"]["count"] = 0
	jobReport["success"]["count"] = 1
	jobReport["success"]["abc"] = ""

	jobReport["failures"] = make(map[string]interface{})
	jobReport["failures"]["def"] = "asbdfasd"	
	jobReport["long_run_projects"] = make(map[string]interface{})
	return jobReport
}

func getDuplicateJobReportWithoutSuccessKeys(jobReport map[string]map[string]interface{}) map[string]map[string]interface{} {
	dupJobReport := make(map[string]map[string]interface{})
	dupJobReport["success"] = make(map[string]interface{})
	dupJobReport["failures"] = make(map[string]interface{})
	dupJobReport["long_run_projects"] = make(map[string]interface{})

	dupJobReport["success"]["count"] = jobReport["success"]["count"]
	dupJobReport["failures"] = jobReport["failures"]
	dupJobReport["long_run_projects"] = jobReport["long_run_projects"]
	return dupJobReport
}