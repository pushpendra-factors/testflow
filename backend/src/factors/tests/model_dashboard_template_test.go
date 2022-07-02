package tests

import (
	// "encoding/json"
	"fmt"
	"time"

	// C "factors/config"
	// "factors/handler/helpers"
	C "factors/config"
	H "factors/handler"
	"factors/handler/helpers"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"

	// U "factors/util"

	// "fmt"
	"net/http"
	"net/http/httptest"

	// "net/http/httptest"
	"testing"
	// "time"

	// "github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin"
	// "github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"

	// log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestCreateDashboardTemplate(t *testing.T){

	rName := U.RandomString(5)
	rDesc := U.RandomString(10)
	t.Run("CreateDashboardTemplate", func(t *testing.T) {
		template, errCode, str := store.GetStore().CreateTemplate(
			&model.DashboardTemplate{
				Title: rName, 
				Description: rDesc, 
	})
		assert.NotNil(t, template)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.Equal(t, "", str)
	})
}

func TestDeleteDashboardTemplate(t *testing.T){
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

func TestReadDashboardTemplate(t *testing.T){
	rName := U.RandomString(5)
	rDesc := U.RandomString(10)
	template, errCode, str := store.GetStore().CreateTemplate(&model.DashboardTemplate{Title: rName, Description: rDesc})
	assert.NotNil(t, template)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Equal(t, "", str)

	t.Run("ReadDashboardTemplate", func(t *testing.T){
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
		Name: C.GetFactorsCookieName(),
		Value: cookieData,
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

func sendCreateTemplateFromDashboardReq(r *gin.Engine, projectId uint64, agent *model.Agent, dashboardID uint64, dashboard *model.Dashboard) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}

	rb := C.NewRequestBuilderWithPrefix(http.MethodPost, fmt.Sprintf("/projects/%d/dashboards/%d/trigger", projectId, dashboard.ID)).
	WithPostParams(dashboard).
	WithCookie(&http.Cookie{
		Name: C.GetFactorsCookieName(),
		Value: cookieData,
		MaxAge: 1000,
	})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error creating create template req")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w;
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

	t.Run("CreateDashboardFromTemplate:WithNoQuery", func(t *testing.T){
		sendCreateDashboardFromTemplateReq(r, project.ID, agent, template2.ID, &template2)
	})
}

func TestAPICreateTemplateFromDashboard(t *testing.T){
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

	t.Run("CreateTemplateFromDashboard:WithNoQuery", func(t *testing.T){
		sendCreateTemplateFromDashboardReq(r, project.ID, agent, dashboard2.ID, dashboard2)
	})
}