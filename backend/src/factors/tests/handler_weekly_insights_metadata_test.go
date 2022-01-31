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
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func sendWIReq(r *gin.Engine, projectId uint64, agent *model.Agent) *httptest.ResponseRecorder {

	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}

	rb := C.NewRequestBuilderWithPrefix(http.MethodGet, fmt.Sprintf("/projects/%d/weekly_insights_metadata", projectId)).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error creating create query req.")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestWIMetadata(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)
	assert.NotNil(t, agent)

	w := sendWIReq(r, project.ID, agent)
	assert.Equal(t, http.StatusOK, w.Code)

	dashboardQuery, errCode, errMsg := store.GetStore().CreateQuery(project.ID, &model.Queries{
		ProjectID: project.ID,
		Type:      model.QueryTypeDashboardQuery,
		Query:     postgres.Jsonb{RawMessage: json.RawMessage(`{}`)},
	})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Empty(t, errMsg)
	assert.NotNil(t, dashboardQuery)

	rName := U.RandomString(5)
	dashboard, errCode := store.GetStore().CreateDashboard(project.ID, agent.UUID,
		&model.Dashboard{Name: rName, Type: model.DashboardTypeProjectVisible})
	assert.NotNil(t, dashboard)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Equal(t, rName, dashboard.Name)
	assert.Equal(t, http.StatusCreated, errCode)

	dashboardUnit, errCode, errMsg := store.GetStore().CreateDashboardUnit(project.ID, agent.UUID,
		&model.DashboardUnit{DashboardId: dashboard.ID, Presentation: model.PresentationLine,
			QueryId: dashboardQuery.ID})
	assert.NotEmpty(t, dashboardUnit)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Empty(t, errMsg)

	insightId := uint64(U.TimeNowUnix())
	errCode, errMsg = store.GetStore().CreateWeeklyInsightsMetadata(&model.WeeklyInsightsMetadata{
		ProjectId:           project.ID,
		QueryId:             dashboardQuery.ID,
		BaseStartTime:       U.TimeNowUnix(),
		BaseEndTime:         U.TimeNowUnix(),
		ComparisonStartTime: U.TimeNowUnix(),
		ComparisonEndTime:   U.TimeNowUnix(),
		InsightType:         "w",
		InsightId:           insightId,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Empty(t, errMsg)

	metadata, errCode, errMsg := store.GetStore().GetWeeklyInsightsMetadata(project.ID)
	assert.NotEmpty(t, metadata)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Empty(t, errMsg)

	assert.Equal(t, project.ID, metadata[0].ProjectId)
	assert.Equal(t, dashboardQuery.ID, metadata[0].QueryId)
	assert.Equal(t, "w", metadata[0].InsightType)
	assert.Equal(t, insightId, metadata[0].InsightId)

	C.GetConfig().ProjectAnalyticsWhitelistedUUIds = []string{agent.UUID}
	assert.True(t, C.IsWeeklyInsightsWhitelisted(agent.UUID, project.ID))

	w = sendWIReq(r, project.ID, agent)
	assert.Equal(t, http.StatusOK, w.Code)

	C.GetConfig().ProjectAnalyticsWhitelistedUUIds = []string{}
}
