package tests

import (
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

	"github.com/stretchr/testify/assert"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func sendCreateDashboardUnitReq(r *gin.Engine, projectId uint64, agent *model.Agent, dashboardId int64, dashboardUnit *model.DashboardUnitRequestPayload) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}

	rb := C.NewRequestBuilderWithPrefix(http.MethodPost, fmt.Sprintf("/projects/%d/dashboards/%d/units", projectId, dashboardId)).
		WithPostParams(dashboardUnit).
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

func sendGetDashboardUnitResult(r *gin.Engine, projectId uint64, agent *model.Agent, dashboardId int64, dashboardUnitId int64, query *gin.H) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}

	rb := C.NewRequestBuilderWithPrefix(http.MethodPost, fmt.Sprintf("/projects/%d/query?dashboard_id=%d&dashboard_unit_id=%d", projectId, dashboardId, dashboardUnitId)).
		WithPostParams(query).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error getting dashboard unit result")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func sendGetDashboardUnitChannelResult(r *gin.Engine, projectId uint64, agent *model.Agent, dashboardId int64, dashboardUnitId int64, query *model.ChannelQuery) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}

	rb := C.NewRequestBuilderWithPrefix(http.MethodPost, fmt.Sprintf("/projects/%d/channels/query?dashboard_id=%d&dashboard_unit_id=%d", projectId, dashboardId, dashboardUnitId)).
		WithPostParams(query).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error getting dashboard unit result")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func sendGetDashboardUnitChannelV1Result(r *gin.Engine, projectId uint64, agent *model.Agent, dashboardId int64, dashboardUnitId int64, query *model.ChannelGroupQueryV1) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}

	rb := C.NewRequestBuilderWithPrefix(http.MethodPost, fmt.Sprintf("/projects/%d/v1/channels/query?dashboard_id=%d&dashboard_unit_id=%d", projectId, dashboardId, dashboardUnitId)).
		WithPostParams(query).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error getting dashboard unit result")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestAPICreateDashboardUnitHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)
	assert.NotNil(t, agent)

	rName := U.RandomString(5)
	dashboard, errCode := store.GetStore().CreateDashboard(project.ID, agent.UUID,
		&model.Dashboard{Name: rName, Type: model.DashboardTypeProjectVisible})
	assert.NotNil(t, dashboard)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Equal(t, rName, dashboard.Name)

	t.Run("CreateDashboardUnit:WithNoQuery", func(t *testing.T) {

		w := sendCreateDashboardUnitReq(r, project.ID, agent, dashboard.ID, &model.DashboardUnitRequestPayload{Presentation: model.PresentationLine})
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("CreateDashboardUnit:WithNoQuery", func(t *testing.T) {

		w := sendCreateDashboardUnitReq(r, project.ID, agent, dashboard.ID, &model.DashboardUnitRequestPayload{Presentation: model.PresentationLine})

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
