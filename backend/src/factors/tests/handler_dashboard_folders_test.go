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

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestCreateDashboardFolder(t *testing.T) {

	r := gin.Default()
	H.InitAppRoutes(r)

	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	rName := U.RandomString(5)
	dashboard, errCode := store.GetStore().CreateDashboard(project.ID, agent.UUID,
		&model.Dashboard{Name: rName, Type: model.DashboardTypeProjectVisible})
	assert.NotNil(t, dashboard)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Equal(t, rName, dashboard.Name)
	t.Run("TestCreateDashboardFolder", func(t *testing.T) {

		fName := U.RandomString(5) + "_TEST"
		reqPayload := &model.DashboardFoldersRequestPayload{Name: fName, DashboardId: dashboard.ID}
		w, err := sendCreateDashboardFolderRequest(r, agent, project.ID, reqPayload)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusCreated, w.Code)

	})

	t.Run("TestNotAllowToCreateAllBoardsNameDashboardFolder", func(t *testing.T) {
		fName := model.ALL_BOARDS_FOLDER
		reqPayload := &model.DashboardFoldersRequestPayload{Name: fName, DashboardId: dashboard.ID}

		w, err := sendCreateDashboardFolderRequest(r, agent, project.ID, reqPayload)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusForbidden, w.Code)

	})
}

func TestDeleteDashboardFolder(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	//CreateDashboardFolderAndAddingDashboards
	fName := U.RandomString(5) + "_TEST"
	reqPayload := &model.DashboardFolders{Name: fName}
	dashboardFolder, errCode := store.GetStore().CreateDashboardFolder(project.ID, reqPayload)
	assert.Equal(t, http.StatusCreated, errCode)

	rName := U.RandomString(5)
	dashboard, errCode := store.GetStore().CreateDashboard(project.ID, agent.UUID,
		&model.Dashboard{Name: rName, Type: model.DashboardTypeProjectVisible, FolderID: dashboardFolder.Id})
	assert.NotNil(t, dashboard)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Equal(t, rName, dashboard.Name)

	w, err := sendDeleteDashboardFolderRequest(r, agent, project.ID, dashboardFolder.Id)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusAccepted, w.Code)

}

func TestUpdateDashboardFolder(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	//CreateDashboardFolderAndAddingDashboards
	fName := U.RandomString(5) + "_TEST"
	reqPayload := &model.DashboardFolders{Name: fName}
	dashboardFolder, errCode := store.GetStore().CreateDashboardFolder(project.ID, reqPayload)
	assert.Equal(t, http.StatusCreated, errCode)

	rName := U.RandomString(5)
	dashboard, errCode := store.GetStore().CreateDashboard(project.ID, agent.UUID,
		&model.Dashboard{Name: rName, Type: model.DashboardTypeProjectVisible, FolderID: dashboardFolder.Id})
	assert.NotNil(t, dashboard)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Equal(t, rName, dashboard.Name)

	updatedName := fName + "_UPDATE"
	updatePayload := &model.UpdatableDashboardFolder{Name: updatedName}
	w, err := sendUpdateDashboardFolderRequest(r, agent, project.ID, updatePayload, dashboardFolder.Id)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusAccepted, w.Code)
}

func TestGetAllDashboardByProjectIdHandler(t *testing.T) {

	r := gin.Default()
	H.InitAppRoutes(r)

	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	//CreateDashboardFolderAndAddingDashboards
	fName := U.RandomString(5) + "_TEST"
	reqPayload := &model.DashboardFolders{Name: fName}
	dashboardFolder, errCode := store.GetStore().CreateDashboardFolder(project.ID, reqPayload)
	assert.Equal(t, http.StatusCreated, errCode)

	rName := U.RandomString(5)
	dashboard, errCode := store.GetStore().CreateDashboard(project.ID, agent.UUID,
		&model.Dashboard{Name: rName, Type: model.DashboardTypeProjectVisible, FolderID: dashboardFolder.Id})
	assert.NotNil(t, dashboard)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Equal(t, rName, dashboard.Name)

	fName2 := U.RandomString(5) + "_TEST"
	reqPayload2 := &model.DashboardFolders{Name: fName2}
	dashboardFolder2, errCode2 := store.GetStore().CreateDashboardFolder(project.ID, reqPayload2)
	assert.Equal(t, http.StatusCreated, errCode2)

	rName2 := U.RandomString(5)
	dashboard2, errCode2 := store.GetStore().CreateDashboard(project.ID, agent.UUID,
		&model.Dashboard{Name: rName2, Type: model.DashboardTypeProjectVisible, FolderID: dashboardFolder2.Id})
	assert.NotNil(t, dashboard2)
	assert.Equal(t, http.StatusCreated, errCode2)
	assert.Equal(t, rName2, dashboard2.Name)

	w, err := sendGetAllDashboardFolderByProjectIdRequest(r, agent, project.ID)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, w.Code)

}

func sendCreateDashboardFolderRequest(r *gin.Engine, agent *model.Agent, projectId int64, reqPayload *model.DashboardFoldersRequestPayload) (*httptest.ResponseRecorder, error) {

	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
		return nil, err
	}

	rb := C.NewRequestBuilderWithPrefix(http.MethodPost, fmt.Sprintf("/projects/%d/dashboard_folder", projectId)).
		WithPostParams(reqPayload).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error creating create dashboard_folder req.")
		return nil, err
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w, nil
}

func sendDeleteDashboardFolderRequest(r *gin.Engine, agent *model.Agent, projectId int64, folderId string) (*httptest.ResponseRecorder, error) {

	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
		return nil, err
	}

	rb := C.NewRequestBuilderWithPrefix(http.MethodDelete, fmt.Sprintf("/projects/%d/dashboard_folder/%s", projectId, folderId)).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error creating delete dashboard_folder req.")
		return nil, err
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w, nil

}

func sendUpdateDashboardFolderRequest(r *gin.Engine, agent *model.Agent, projectId int64, reqPayload *model.UpdatableDashboardFolder, folderId string) (*httptest.ResponseRecorder, error) {

	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
		return nil, err
	}

	rb := C.NewRequestBuilderWithPrefix(http.MethodPut, fmt.Sprintf("/projects/%d/dashboard_folder/%s", projectId, folderId)).
		WithPostParams(reqPayload).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error creating update dashboard_folder req.")
		return nil, err
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w, nil
}

func sendGetAllDashboardFolderByProjectIdRequest(r *gin.Engine, agent *model.Agent, projectId int64) (*httptest.ResponseRecorder, error) {

	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
		return nil, err
	}

	rb := C.NewRequestBuilderWithPrefix(http.MethodGet, fmt.Sprintf("/projects/%d/dashboard_folder", projectId)).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error creating get dashboard_folder req.")
		return nil, err
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w, nil
}
