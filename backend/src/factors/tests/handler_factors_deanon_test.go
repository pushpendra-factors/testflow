package tests

import (
	C "factors/config"
	"factors/handler"
	H "factors/handler"
	"factors/handler/helpers"
	"factors/model/model"
	"factors/model/store"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func sendUpdateFactorsDeanonProviderReq(r *gin.Engine, projectId int64, agent *model.Agent, deanonProvider string) *httptest.ResponseRecorder {

	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}

	// "/:project_id/factors_deanon/provider/:name/enable"
	rb := C.NewRequestBuilderWithPrefix(http.MethodPost, fmt.Sprintf("/projects/%d/factors_deanon/provider/%v/enable", projectId, deanonProvider)).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	w := httptest.NewRecorder()
	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error Creating UpdateFactorsDeanon Req")
	}
	r.ServeHTTP(w, req)
	return w
}

func TestAPIUpdateFactorsDeanonProvider(t *testing.T) {

	r := gin.Default()
	H.InitAppRoutes(r)

	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)

	t.Run("Update6SignalAsFactorsDeanonProvider", func(t *testing.T) {
		w := sendUpdateFactorsDeanonProviderReq(r, project.ID, agent, handler.FACTORS_SIXSIGNAL)
		assert.Equal(t, http.StatusOK, w.Code)

	})

	t.Run("UpdateClearbitAsFactorsDeanonProvider", func(t *testing.T) {
		store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{FactorsClearbitKey: "test123"})
		w := sendUpdateFactorsDeanonProviderReq(r, project.ID, agent, handler.FACTORS_CLEARBIT)
		assert.Equal(t, http.StatusOK, w.Code)

	})

	t.Run("UpdateAsFactorsDeanonProviderWithWrongProvider", func(t *testing.T) {
		store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{FactorsClearbitKey: "test123"})
		w := sendUpdateFactorsDeanonProviderReq(r, project.ID, agent, "wrongProvider")
		assert.Equal(t, http.StatusBadRequest, w.Code)

	})

}
