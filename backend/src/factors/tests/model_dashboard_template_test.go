package tests

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm/dialects/postgres"
	"time"

	C "factors/config"
	H "factors/handler"
	"factors/handler/helpers"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/stretchr/testify/assert"
)

func TestCreateDashboardTemplate(t *testing.T) {

	rName := U.RandomString(5)
	rDesc := U.RandomString(10)
	t.Run("CreateDashboardTemplate", func(t *testing.T) {
		template, errCode, str := store.GetStore().CreateTemplate(
			&model.DashboardTemplate{
				Title:       rName,
				Description: rDesc,
			})
		assert.NotNil(t, template)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.Equal(t, "", str)
	})
}

func TestCreateDashboardTemplateData(t *testing.T) {

	rName := U.RandomString(5)
	rDesc := U.RandomString(10)
	t.Run("CreateDashboardTemplate", func(t *testing.T) {
		template, errCode, str := store.GetStore().CreateTemplate(
			&model.DashboardTemplate{
				Title:       rName,
				Description: rDesc,

				Dashboard: &postgres.Jsonb{RawMessage: json.RawMessage(`{"id":1,
        "name": "First Dashboard in Test Project",
        "type": "pv",
        "class": "template_created",
        "is_deleted":false,
        "settings":{"type":"public"}}`)},
				Units: &postgres.Jsonb{RawMessage: json.RawMessage(`[
        {
            "id": 1,
            "title": "Unit 1",
            "description": "Description 1",
            "presentation": "sp",
            "position": 11,
            "size": 111,
            "query_type": 1111,
            "query_settings": {
                "chart": "x1"
            },
            "query": {
                "cl": "events",
                "ec": "each_given_event",
                "ewp": [
                    {
                        "an": "",
                        "na": "Deal Won",
                        "pr": []
                    }
                ]
            }
        },
        {
            "id": 2,
            "title": "Unit 2",
            "description": "Description 2",
            "presentation": "pb",
            "position": 22,
            "size": 222,
            "query_type": 2222,
            "query_settings": {
                "chart": "x2"
            },
            "query": {
                "cl": "events",
                "ec": "each_given_event",
                "ewp": [
                    {
                        "an": "",
                        "na": "Deal Won",
                        "pr": []
                    }
                ]
            }
        }
    ]`)},
			})
		assert.NotNil(t, template)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.Equal(t, "", str)
	})
}

func TestDeleteDashboardTemplate(t *testing.T) {
	rName := U.RandomString(5)
	rDesc := U.RandomString(10)
	template, errCode, str := store.GetStore().CreateTemplate(&model.DashboardTemplate{Title: rName, Description: rDesc})
	assert.NotNil(t, template)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Equal(t, "", str)
	t.Run("DeleteDashboardTemplate", func(t *testing.T) {
		err := store.GetStore().DeleteTemplate(template.ID)
		assert.Equal(t, http.StatusAccepted, err)
	})
}

func TestReadDashboardTemplate(t *testing.T) {
	rName := U.RandomString(5)
	rDesc := U.RandomString(10)
	template, errCode, str := store.GetStore().CreateTemplate(&model.DashboardTemplate{Title: rName, Description: rDesc})
	assert.NotNil(t, template)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Equal(t, "", str)

	t.Run("ReadDashboardTemplate", func(t *testing.T) {
		template, errCode := store.GetStore().SearchTemplateWithTemplateID(template.ID)
		assert.NotNil(t, template)
		assert.Equal(t, 302, errCode)
		assert.Equal(t, rName, template.Title)
	})
}

func sendCreateDashboardFromTemplateReq(r *gin.Engine, projectId uint64, agent *model.Agent, templateID string, dashboardTemplate *model.DashboardTemplate) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}

	rb := C.NewRequestBuilderWithPrefix(http.MethodPost, fmt.Sprintf("/projects/%d/dashboard_template/%s/trigger", projectId, templateID)).
		WithPostParams(dashboardTemplate).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error creating create dashboard_template req")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func sendCreateTemplateFromDashboardReq(r *gin.Engine, projectId uint64, agent *model.Agent, dashboardID int64, dashboard *model.Dashboard) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}

	rb := C.NewRequestBuilderWithPrefix(http.MethodPost, fmt.Sprintf("/projects/%d/dashboards/%d/trigger", projectId, dashboard.ID)).
		WithPostParams(dashboard).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error creating create template req")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestAPICreateDashboardFromTemplate(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)
	assert.NotNil(t, agent)

	rName := U.RandomString(5)
	desc := "testing create dashboard from template abc"
	template, errCode, _ := store.GetStore().CreateTemplate(&model.DashboardTemplate{
		Title: rName, Description: desc})
	if errCode != http.StatusCreated {
		log.Error("Error creating template in database")
	}

	template2, errCode := store.GetStore().SearchTemplateWithTemplateID(template.ID)
	if errCode != 302 {
		log.Error("Error fetching dashboard from database")
	}

	t.Run("CreateDashboardFromTemplate:WithNoQuery", func(t *testing.T) {
		sendCreateDashboardFromTemplateReq(r, project.ID, agent, template2.ID, &template2)
	})
}

func TestAPICreateTemplateFromDashboard(t *testing.T) {
	//assert.Nil(t, err)
	r := gin.Default()
	H.InitAppRoutes(r)

	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)
	assert.NotNil(t, agent)

	rName := U.RandomString(5)
	desc := "testing create template from dashboard"

	dashboard, errCode := store.GetStore().CreateDashboard(project.ID, agent.UUID, &model.Dashboard{
		Name: rName, Description: desc, Type: model.DashboardTypeProjectVisible})
	assert.NotNil(t, dashboard)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Equal(t, rName, dashboard.Name)

	dashboard2, errCode := store.GetStore().GetDashboard(project.ID, agent.UUID, dashboard.ID)
	if errCode != 302 {
		log.Error("Error fetching dashboard from database")
	}

	t.Run("CreateTemplateFromDashboard:WithNoQuery", func(t *testing.T) {
		sendCreateTemplateFromDashboardReq(r, project.ID, agent, dashboard2.ID, dashboard2)
	})
}
