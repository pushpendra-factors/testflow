package tests

import (
	C "factors/config"
	H "factors/handler"
	"factors/handler/helpers"
	M "factors/model"
	U "factors/util"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func sendCreateDashboardUnitReq(r *gin.Engine, projectId uint64, agent *M.Agent, dashboardId uint64, dashboardUnit *H.DashboardUnitRequestPayload) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}

	rb := U.NewRequestBuilder(http.MethodPost, fmt.Sprintf("/projects/%d/dashboards/%d/units", projectId, dashboardId)).
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

func TestAPICreateDashboardUnitHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)
	assert.NotNil(t, agent)

	rName := U.RandomString(5)
	dashboard, errCode := M.CreateDashboard(project.ID, agent.UUID, &M.Dashboard{Name: rName, Type: M.DashboardTypeProjectVisible})
	assert.NotNil(t, dashboard)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Equal(t, rName, dashboard.Name)

	t.Run("CreateDashboardUnit:WithValidQuery", func(t *testing.T) {
		rTitle := U.RandomString(5)
		query := M.Query{
			EventsCondition: M.EventCondAnyGivenEvent,
			From:            1556602834,
			To:              1557207634,
			Type:            M.QueryTypeEventsOccurrence,
			EventsWithProperties: []M.QueryEventWithProperties{
				M.QueryEventWithProperties{
					Name: "event1",
				},
			},
			OverridePeriod: true,
		}
		w := sendCreateDashboardUnitReq(r, project.ID, agent, dashboard.ID, &H.DashboardUnitRequestPayload{Title: rTitle,
			Query: query, Presentation: M.PresentationLine})
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("CreateDashboardUnit:WithInvalidQuery", func(t *testing.T) {
		rTitle := U.RandomString(5)
		query := M.Query{
			EventsCondition:      M.EventCondAnyGivenEvent,
			From:                 1556602834,
			To:                   1557207634,
			Type:                 M.QueryTypeEventsOccurrence,
			EventsWithProperties: []M.QueryEventWithProperties{}, // invalid, no events.
			OverridePeriod:       true,
		}
		w := sendCreateDashboardUnitReq(r, project.ID, agent, dashboard.ID, &H.DashboardUnitRequestPayload{Title: rTitle,
			Query: query, Presentation: M.PresentationLine})
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
