package tests

import (
	"encoding/json"
	C "factors/config"
	H "factors/handler"
	"factors/handler/helpers"
	"factors/model/model"
	"factors/model/store"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func sendCreateFilterReq(r *gin.Engine, projectId int64, agent *model.Agent, name, expr string) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}
	rb := C.NewRequestBuilderWithPrefix(http.MethodPost, fmt.Sprintf("/projects/%d/filters", projectId)).
		WithPostParams(map[string]string{
			"name": name,
			"expr": expr,
		}).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})
	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error Creating CreateFilter Req")
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w

}

func TestAPICreateFilterHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)

	t.Run("GetProjectSettings1", func(t *testing.T) {
		name := "u1_u2"
		expr := "a.com/u1/u2"

		w := sendCreateFilterReq(r, project.ID, agent, name, expr)
		assert.Equal(t, http.StatusCreated, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		var jsonResponseMap map[string]interface{}
		json.Unmarshal(jsonResponse, &jsonResponseMap)
		assert.NotZero(t, jsonResponseMap["project_id"])
		assert.NotZero(t, jsonResponseMap["id"])
		assert.NotNil(t, jsonResponseMap["name"])
		assert.Equal(t, "u1_u2", jsonResponseMap["name"])
		assert.NotNil(t, jsonResponseMap["expr"])
		assert.Equal(t, expr, jsonResponseMap["expr"])
	})

	t.Run("GetProjectSettings2", func(t *testing.T) {
		name := "u1_v1"
		expr := "a.com/u1/:v1"

		w := sendCreateFilterReq(r, project.ID, agent, name, expr)
		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Equal(t, http.StatusCreated, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		var jsonResponseMap map[string]interface{}
		json.Unmarshal(jsonResponse, &jsonResponseMap)
		assert.NotZero(t, jsonResponseMap["project_id"])
		assert.NotZero(t, jsonResponseMap["id"])
		assert.NotNil(t, jsonResponseMap["name"])
		assert.Equal(t, "u1_v1", jsonResponseMap["name"])
		assert.NotNil(t, jsonResponseMap["expr"])
		assert.Equal(t, expr, jsonResponseMap["expr"])

	})

	t.Run("InvalidName", func(t *testing.T) {
		name := ""
		expr := "a.com/u1/:v2"
		w := sendCreateFilterReq(r, project.ID, agent, name, expr)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("InvalidExprAndName", func(t *testing.T) {
		name := ""
		expr := ""
		w := sendCreateFilterReq(r, project.ID, agent, name, expr)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		var jsonResponseMap map[string]interface{}
		json.Unmarshal(jsonResponse, &jsonResponseMap)
		assert.NotNil(t, jsonResponseMap["error"])
	})

	t.Run("InvalidExpr", func(t *testing.T) {
		name := "u1_u2"
		expr := ""
		w := sendCreateFilterReq(r, project.ID, agent, name, expr)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		var jsonResponseMap map[string]interface{}
		json.Unmarshal(jsonResponse, &jsonResponseMap)
		assert.NotNil(t, jsonResponseMap["error"])
	})

	t.Run("InvalidNameHasDollar", func(t *testing.T) {
		name := "$dollarName"
		expr := "a.com/dollar"
		w := sendCreateFilterReq(r, project.ID, agent, name, expr)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		var jsonResponseMap map[string]interface{}
		json.Unmarshal(jsonResponse, &jsonResponseMap)
		assert.NotNil(t, jsonResponseMap["error"])

	})

	t.Run("SanitizedExpr", func(t *testing.T) {
		name := "u1_u2"
		// user copied the url and pasted as expression.
		expr := "https://a.com/u1/u5?q=search_string"
		w := sendCreateFilterReq(r, project.ID, agent, name, expr)
		assert.Equal(t, http.StatusCreated, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		var jsonResponseMap map[string]interface{}
		json.Unmarshal(jsonResponse, &jsonResponseMap)
		assert.NotZero(t, jsonResponseMap["project_id"])
		assert.NotZero(t, jsonResponseMap["id"])
		assert.NotNil(t, jsonResponseMap["name"])
		assert.Equal(t, "u1_u2", jsonResponseMap["name"])
		assert.NotNil(t, jsonResponseMap["expr"])
		// sanitized expr from user given url.
		assert.Equal(t, "a.com/u1/u5", jsonResponseMap["expr"])
	})

}

func sendGetFilterRequest(projectId int64, agent *model.Agent, r *gin.Engine) *httptest.ResponseRecorder {

	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}

	rb := C.NewRequestBuilderWithPrefix(http.MethodGet, fmt.Sprintf("/projects/%d/filters", projectId)).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})
	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error Creating GetFilter Req")
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestAPIGetFiltersHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)

	t.Run("MissingFilters", func(t *testing.T) {
		w := sendGetFilterRequest(project.ID, agent, r)
		assert.Equal(t, http.StatusOK, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		retFilters := make([]H.API_FilterResponePayload, 0, 0)
		json.Unmarshal(jsonResponse, &retFilters)
		assert.Equal(t, 0, len(retFilters))
	})

	t.Run("HasFilters", func(t *testing.T) {
		// Filters created.
		filters := map[string]string{
			"u1_u2": "a.com/u1/u2",
			"u1_v1": "a.com/u1/:v1",
		}

		for k, v := range filters {
			store.GetStore().CreateOrGetFilterEventName(&model.EventName{
				ProjectId: project.ID, Name: k, FilterExpr: v})
		}

		w := sendGetFilterRequest(project.ID, agent, r)
		assert.Equal(t, http.StatusOK, w.Code)
		jsonResponse1, _ := ioutil.ReadAll(w.Body)
		retFilters1 := make([]H.API_FilterResponePayload, 0, 0)
		json.Unmarshal(jsonResponse1, &retFilters1)
		assert.Equal(t, 2, len(retFilters1))
	})

}

func sendUpdateFilterReq(r *gin.Engine, projectId int64, filterId string, agent *model.Agent, name, expr *string) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}

	params := make(map[string]string)
	if name != nil {
		params["name"] = *name
	}
	if expr != nil {
		params["expr"] = *expr
	}

	rb := C.NewRequestBuilderWithPrefix(http.MethodPut, fmt.Sprintf("/projects/%d/filters/%s", projectId, filterId)).
		WithPostParams(params).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})
	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error Creating CreateFilter Req")
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w

}
func TestAPIUpdateFilterHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)

	filter, _ := store.GetStore().CreateOrGetFilterEventName(&model.EventName{
		ProjectId: project.ID, Name: "u1_u2", FilterExpr: "a.com/u1/:u2"})
	assert.NotNil(t, filter)

	t.Run("GetProjectSettings", func(t *testing.T) {
		name := "new_name"
		w := sendUpdateFilterReq(r, project.ID, filter.ID, agent, &name, nil)
		assert.Equal(t, http.StatusAccepted, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		var jsonResponseMap map[string]interface{}
		json.Unmarshal(jsonResponse, &jsonResponseMap)
		assert.NotNil(t, jsonResponseMap["name"])
		assert.Equal(t, name, jsonResponseMap["name"])
		assert.Nil(t, jsonResponseMap["expr"]) // omit empty.
	})

	t.Run("EmptyName", func(t *testing.T) {
		name := ""
		w := sendUpdateFilterReq(r, project.ID, filter.ID, agent, &name, nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("TryUpdatingExpr", func(t *testing.T) {
		expr := "a.com/u1/u3"
		w := sendUpdateFilterReq(r, project.ID, filter.ID, agent, nil, &expr)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		var jsonResponseMap map[string]interface{}
		json.Unmarshal(jsonResponse, &jsonResponseMap)
		assert.NotNil(t, jsonResponseMap["error"])
	})

	t.Run("NameHasDollar", func(t *testing.T) {
		name := "$new_name"
		w := sendUpdateFilterReq(r, project.ID, filter.ID, agent, &name, nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		var jsonResponseMap map[string]interface{}
		json.Unmarshal(jsonResponse, &jsonResponseMap)
		assert.NotNil(t, jsonResponseMap["error"])
	})
}

func sendDeleteFilterReq(r *gin.Engine, projectId int64, fileterId string, agent *model.Agent) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}

	rb := C.NewRequestBuilderWithPrefix(http.MethodDelete, fmt.Sprintf("/projects/%d/filters/%s", projectId, fileterId)).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error Creating DeleteFilter Req")
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestAPIDeleteFilterHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)

	t.Run("InvalidFilter", func(t *testing.T) {
		invalidFilterId := "99999"
		w := sendDeleteFilterReq(r, project.ID, invalidFilterId, agent)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		var jsonResponseMap map[string]interface{}
		json.Unmarshal(jsonResponse, &jsonResponseMap)
		assert.NotNil(t, jsonResponseMap["error"])
	})

	t.Run("ValidFilter", func(t *testing.T) {
		filter, _ := store.GetStore().CreateOrGetFilterEventName(&model.EventName{
			ProjectId: project.ID, Name: "u1_u2", FilterExpr: "a.com/u1/:u2"})
		assert.NotNil(t, filter)

		w := sendDeleteFilterReq(r, project.ID, filter.ID, agent)
		assert.Equal(t, http.StatusAccepted, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		var jsonResponseMap map[string]interface{}
		json.Unmarshal(jsonResponse, &jsonResponseMap)
		assert.NotNil(t, jsonResponseMap["project_id"])
		assert.NotNil(t, jsonResponseMap["id"])
	})
}
