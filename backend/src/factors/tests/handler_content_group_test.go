package tests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	H "factors/handler"
	"factors/handler/helpers"
	"factors/model/model"
	"factors/model/store"

	C "factors/config"

	U "factors/util"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func sendCreateContentGroupRequest(r *gin.Engine, request model.ContentGroup, agent *model.Agent, projectID int64) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}

	rb := C.NewRequestBuilderWithPrefix(http.MethodPost, fmt.Sprintf("/projects/%d/v1/contentgroup", projectID)).
		WithPostParams(request).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error creating content group")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func sendUpdateContentGroupRequest(r *gin.Engine, request model.ContentGroup, agent *model.Agent, projectID int64, id string) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}

	rb := C.NewRequestBuilderWithPrefix(http.MethodPut, fmt.Sprintf("/projects/%d/v1/contentgroup/%v", projectID, id)).
		WithPostParams(request).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error updating content group")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func sendGetContentGroupsRequest(r *gin.Engine, agent *model.Agent, projectID int64) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}

	rb := C.NewRequestBuilderWithPrefix(http.MethodGet, fmt.Sprintf("/projects/%d/v1/contentgroup", projectID)).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error getting content group")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func sendDeleteContentGroupRequest(r *gin.Engine, agent *model.Agent, projectID int64, id string) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}

	rb := C.NewRequestBuilderWithPrefix(http.MethodDelete, fmt.Sprintf("/projects/%d/v1/contentgroup/%v", projectID, id)).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error deleting content group")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
func TestContentGroupAPI(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)

	request := model.ContentGroup{}
	request.ContentGroupName = fmt.Sprintf("%v-%v", "abc", U.TimeNowUnix())
	request.ContentGroupDescription = "description"
	value := model.ContentGroupValue{}
	value.Operator = "startsWith"
	value.LogicalOp = "OR"
	value.Value = "123"
	filters := make([]model.ContentGroupValue, 0)
	filters = append(filters, value)
	contentGroupValueArray := make([]model.ContentGroupRule, 0)
	contentGroupValue := model.ContentGroupRule{
		ContentGroupValue: "value",
		Rule:              filters,
	}
	contentGroupValueArray = append(contentGroupValueArray, contentGroupValue)
	contentGroupValueJson, err := json.Marshal(contentGroupValueArray)
	request.Rule = &postgres.Jsonb{contentGroupValueJson}
	w := sendCreateContentGroupRequest(r, request, agent, project.ID)
	assert.Equal(t, http.StatusCreated, w.Code)

	w = sendGetContentGroupsRequest(r, agent, project.ID)
	assert.Equal(t, http.StatusOK, w.Code)
	contentGroups := make([]model.ContentGroup, 0)
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &contentGroups)
	assert.Equal(t, len(contentGroups), 1)
	var contentGroupRuleResult []model.ContentGroupRule
	U.DecodePostgresJsonbToStructType(contentGroups[0].Rule, &contentGroupRuleResult)
	assert.Equal(t, contentGroupRuleResult[0].Rule[0].Value, "123")

	request = model.ContentGroup{}
	value = model.ContentGroupValue{}
	value.Operator = "startsWith"
	value.LogicalOp = "OR"
	value.Value = "123456"
	filters = make([]model.ContentGroupValue, 0)
	filters = append(filters, value)
	contentGroupValueArray = make([]model.ContentGroupRule, 0)
	contentGroupValue = model.ContentGroupRule{
		ContentGroupValue: "value",
		Rule:              filters,
	}
	contentGroupValueArray = append(contentGroupValueArray, contentGroupValue)
	contentGroupValueJson, err = json.Marshal(contentGroupValueArray)
	request.Rule = &postgres.Jsonb{contentGroupValueJson}
	w = sendUpdateContentGroupRequest(r, request, agent, project.ID, contentGroups[0].ID)
	assert.Equal(t, http.StatusOK, w.Code)

	w = sendGetContentGroupsRequest(r, agent, project.ID)
	assert.Equal(t, http.StatusOK, w.Code)
	contentGroups = make([]model.ContentGroup, 0)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &contentGroups)
	assert.Equal(t, len(contentGroups), 1)
	U.DecodePostgresJsonbToStructType(contentGroups[0].Rule, &contentGroupRuleResult)
	assert.Equal(t, contentGroupRuleResult[0].Rule[0].Value, "123456")

	result := store.GetStore().CheckURLContentGroupValue("123456789", project.ID)
	assert.Equal(t, result[contentGroups[0].ContentGroupName], "value")

	w = sendDeleteContentGroupRequest(r, agent, project.ID, contentGroups[0].ID)
	assert.Equal(t, http.StatusOK, w.Code)

	w = sendGetContentGroupsRequest(r, agent, project.ID)
	assert.Equal(t, http.StatusOK, w.Code)
	contentGroups = make([]model.ContentGroup, 0)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &contentGroups)
	assert.Equal(t, len(contentGroups), 0)
}
