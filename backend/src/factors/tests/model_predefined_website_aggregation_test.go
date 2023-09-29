package tests

import (
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"net/http"
	"testing"

	log "github.com/sirupsen/logrus"
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
			From:              1669856400,
			To:                1675213200,
			InternalEventType: model.PredefEventTypeSession,
			WidgetInternalID:  2,
			WidgetName:        model.PredefWidGtmParams,
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
		assert.Equal(t, 1, len(result[1].Rows))
		assert.Equal(t, float64(3), result[1].Rows[0][2].(float64))

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
			From:              1669856400,
			To:                1675213200,
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
		log.WithField("result", result).Warn("kark2")
		assert.Equal(t, http.StatusOK, statusCode)

		assert.Equal(t, 2, len(result[0].Headers))
		assert.Equal(t, 1, len(result[0].Rows))
		assert.Equal(t, float64(3), result[0].Rows[0][1].(float64))

		assert.Equal(t, 3, len(result[1].Headers))
		assert.Equal(t, 1, len(result[1].Rows))
		assert.Equal(t, float64(3), result[1].Rows[0][2].(float64))

	})
}
