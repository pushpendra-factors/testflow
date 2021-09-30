package tests

import (
	H "factors/handler"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestKpiAnalytics(t *testing.T) {
	a := gin.Default()
	H.InitAppRoutes(a)

	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	createdUserID1, _ := store.GetStore().CreateUser(&model.User{ProjectId: project.ID})

	startTimestamp := U.UnixTimeBeforeDuration(time.Hour * 1)
	stepTimestamp := startTimestamp

	payload := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, "user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`, "s0", createdUserID1, stepTimestamp, "A", 1234)
	w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response := DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	//t.Run("abc", func(t *testing.T) {
	//
	//	query := model.KPIQuery{
	//
	//		Category:        "events",
	//		DisplayCategory: "page_views",
	//		PageUrl:         "s0",
	//		//Metrics:         []string{"page_views", "unique_users"},
	//		Metrics: []string{"page_views"},
	//		Filters: []model.KPIFilter{
	//			{
	//				PropertyName:     "user_id",
	//				PropertyDataType: "categorical",
	//				Entity:           "user",
	//				Condition:        "equals",
	//				Value:            "1",
	//				LogicalOp:        "AND",
	//			},
	//		},
	//		From: startTimestamp,
	//		To:   startTimestamp + 40,
	//	}
	//
	//	kpiQueryGroup := model.KPIQueryGroup{
	//		Class:         "kpi",
	//		Queries:       []model.KPIQuery{query},
	//		GlobalFilters: []model.KPIFilter{},
	//		GlobalGroupBy: []model.KPIGroupBy{},
	//	}
	//
	//	result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup)
	//	assert.Equal(t, http.StatusOK, statusCode)
	//	log.Warn(result)
	//})

	//t.Run("abc1", func(t *testing.T) {
	//
	//	query := model.KPIQuery{
	//
	//		Category:        "events",
	//		DisplayCategory: "page_views",
	//		PageUrl:         "s0",
	//		//Metrics:         []string{"page_views", "unique_users"},
	//		Metrics:          []string{"page_views"},
	//		Filters:          nil,
	//		From:             startTimestamp,
	//		To:               startTimestamp + 40,
	//		GroupByTimestamp: "date",
	//	}
	//
	//	kpiQueryGroup := model.KPIQueryGroup{
	//		Class:         "kpi",
	//		Queries:       []model.KPIQuery{query},
	//		GlobalFilters: []model.KPIFilter{},
	//		GlobalGroupBy: []model.KPIGroupBy{},
	//	}
	//
	//	result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup)
	//	assert.Equal(t, http.StatusOK, statusCode)
	//	log.Warn(result)
	//})

	//t.Run("abc2", func(t *testing.T) {
	//
	//	query := model.KPIQuery{
	//
	//		Category:        "events",
	//		DisplayCategory: "page_views",
	//		PageUrl:         "s0",
	//		//Metrics:         []string{"page_views", "unique_users"},
	//		Metrics: []string{"page_views"},
	//		Filters: nil,
	//		From:    startTimestamp,
	//		To:      startTimestamp + 40,
	//		//GroupByTimestamp: "date",
	//	}
	//
	//	kpiQueryGroup := model.KPIQueryGroup{
	//		Class:         "kpi",
	//		Queries:       []model.KPIQuery{query},
	//		GlobalFilters: []model.KPIFilter{},
	//		GlobalGroupBy: []model.KPIGroupBy{
	//			{
	//				ObjectType:       "s0",
	//				PropertyName:     "user_id",
	//				PropertyDataType: "categorical",
	//				GroupByType:      "",
	//				Granularity:      "",
	//				Entity:           "user",
	//			},
	//		},
	//	}
	//
	//	result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup)
	//	assert.Equal(t, http.StatusOK, statusCode)
	//	log.Warn(result)
	//})

	t.Run("abc3", func(t *testing.T) {

		query := model.KPIQuery{

			Category:        "events",
			DisplayCategory: "website_session",
			Metrics: []string{"average_initial_page_load_time"},
			Filters: nil,
			From:    startTimestamp,
			To:      startTimestamp + 40,
		}

		kpiQueryGroup := model.KPIQueryGroup{
			Class:         "kpi",
			Queries:       []model.KPIQuery{query},
			GlobalFilters: []model.KPIFilter{},
			GlobalGroupBy: []model.KPIGroupBy{
				{
					ObjectType:       "s0",
					PropertyName:     "user_id",
					PropertyDataType: "categorical",
					GroupByType:      "",
					Granularity:      "",
					Entity:           "user",
				},
			},
		}

		result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup)
		assert.Equal(t, http.StatusOK, statusCode)
		log.Warn(result)
	})
}
