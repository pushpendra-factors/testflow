package tests

import (
	C "factors/config"
	Const "factors/constants"
	H "factors/handler"
	"factors/handler/helpers"
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

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestKpiAnalytics(t *testing.T) {
	a := gin.Default()
	H.InitAppRoutes(a)

	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"

	r2 := gin.Default()
	H.InitDataServiceRoutes(r2)
	Const.SetSmartPropertiesReservedNames()

	project, customerAccountID, _, statusCode := createProjectAndAddAdwordsDocument(t, r2)
	log.Warn(customerAccountID)
	if statusCode != http.StatusAccepted {
		assert.Equal(t, false, true)
		return
	}

	createdUserID1, _ := store.GetStore().CreateUser(&model.User{ProjectId: project.ID})

	startTimestamp := U.UnixTimeBeforeDuration(time.Hour * 1)
	stepTimestamp := startTimestamp

	payload := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, "user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`, "s0", createdUserID1, stepTimestamp, "A", 1234)
	w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response := DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	t.Run("abc", func(t *testing.T) {

		query := model.KPIQuery{

			Category:        "events",
			DisplayCategory: "page_views",
			PageUrl:         "s0",
			//Metrics:         []string{"page_views", "unique_users"},
			Metrics: []string{"page_views"},
			Filters: []model.KPIFilter{
				{
					PropertyName:     "user_id",
					PropertyDataType: "categorical",
					Entity:           "user",
					Condition:        "equals",
					Value:            "1",
					LogicalOp:        "AND",
				},
			},
			From: startTimestamp,
			To:   startTimestamp + 40,
		}

		kpiQueryGroup := model.KPIQueryGroup{
			Class:         "kpi",
			Queries:       []model.KPIQuery{query},
			GlobalFilters: []model.KPIFilter{},
			GlobalGroupBy: []model.KPIGroupBy{},
		}

		result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup)
		assert.Equal(t, http.StatusOK, statusCode)
		log.Warn(result)
	})

	t.Run("abc1", func(t *testing.T) {

		query := model.KPIQuery{

			Category:        "events",
			DisplayCategory: "page_views",
			PageUrl:         "s0",
			//Metrics:         []string{"page_views", "unique_users"},
			Metrics:          []string{"page_views"},
			Filters:          nil,
			From:             startTimestamp,
			To:               startTimestamp + 40,
			GroupByTimestamp: "date",
		}

		kpiQueryGroup := model.KPIQueryGroup{
			Class:         "kpi",
			Queries:       []model.KPIQuery{query},
			GlobalFilters: []model.KPIFilter{},
			GlobalGroupBy: []model.KPIGroupBy{},
		}

		result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup)
		assert.Equal(t, http.StatusOK, statusCode)
		log.Warn(result)
	})

	t.Run("abc2", func(t *testing.T) {

		query := model.KPIQuery{

			Category:        "events",
			DisplayCategory: "page_views",
			PageUrl:         "s0",
			//Metrics:         []string{"page_views", "unique_users"},
			Metrics:  []string{"page_views"},
			Filters:  nil,
			From:     startTimestamp,
			To:       startTimestamp + 40,
			Timezone: "Asia/Kolkata",
			//GroupByTimestamp: "date",
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

	t.Run("abc3", func(t *testing.T) {

		query := model.KPIQuery{

			Category:        "events",
			DisplayCategory: "website_session",
			Metrics:         []string{"average_initial_page_load_time"},
			Filters:         nil,
			From:            startTimestamp,
			To:              startTimestamp + 40,
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

	t.Run("abc4", func(t *testing.T) {

		query := model.KPIQuery{

			Category:        "channels",
			DisplayCategory: "adwords_metrics",
			Metrics:         []string{"impressions"},
			Filters:         nil,
			From:            startTimestamp,
			To:              startTimestamp + 40,
		}

		kpiQueryGroup := model.KPIQueryGroup{
			Class:         "kpi",
			Queries:       []model.KPIQuery{query},
			GlobalFilters: []model.KPIFilter{},
			GlobalGroupBy: []model.KPIGroupBy{},
		}

		result, statusCode := store.GetStore().ExecuteKPIQueryGroup(project.ID, uuid.New().String(), kpiQueryGroup)
		assert.Equal(t, http.StatusOK, statusCode)
		log.Warn(result)
	})
}

func TestKpiAnalyticsHandler(t *testing.T) {
	a := gin.Default()
	H.InitAppRoutes(a)

	r := gin.Default()
	H.InitSDKServiceRoutes(r)

	project, agent, _ := SetupProjectWithAgentDAO()
	assert.NotNil(t, project)

	query := model.KPIQuery{
		Category:        "events",
		DisplayCategory: "page_views",
		PageUrl:         "s0",
		Metrics:         []string{"page_views"},
		GroupBy:         []M.KPIGroupBy{},
		From:            20210801,
		To:              20210801 + 40,
	}

	kpiQueryGroup := model.KPIQueryGroup{
		Class:         "kpi",
		Queries:       []model.KPIQuery{query},
		GlobalFilters: []model.KPIFilter{},
		GlobalGroupBy: []model.KPIGroupBy{},
	}

	sendKPIAnalyticsQueryReq(a, project.ID, agent, kpiQueryGroup)
}

func sendKPIAnalyticsQueryReq(r *gin.Engine, projectId uint64, agent *M.Agent, kpiqueryGroup model.KPIQueryGroup) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}

	rb := C.NewRequestBuilderWithPrefix(http.MethodPost, fmt.Sprintf("/projects/%d/v1/kpi/query", projectId)).
		WithPostParams(kpiqueryGroup).
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
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	log.Warn(jsonResponse)
	return w
}
