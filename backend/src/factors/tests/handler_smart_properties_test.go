package tests

import (
	"encoding/json"
	C "factors/config"
	H "factors/handler"
	"factors/handler/helpers"
	"factors/model/model"
	U "factors/util"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func sendCreateSmartPropertyReq(r *gin.Engine, project_id int64, agent *model.Agent, rules *postgres.Jsonb, type_alias string, name string, description string) *httptest.ResponseRecorder {
	payload := map[string]interface{}{
		"project_id":  project_id,
		"type_alias":  type_alias,
		"rules":       rules,
		"name":        name,
		"description": description,
	}
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}

	url := "/projects/" + strconv.FormatUint(uint64(project_id), 10) + "/v1/smart_properties/rules"
	rb := C.NewRequestBuilderWithPrefix(http.MethodPost, url).
		WithPostParams(payload).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error sending create smart properties request.")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
func sendGetSmartPropertyRulesReq(r *gin.Engine, projectID int64, agent *model.Agent) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}

	url := "/projects/" + strconv.FormatUint(uint64(projectID), 10) + "/v1/smart_properties/rules"
	rb := C.NewRequestBuilderWithPrefix(http.MethodGet, url).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})
	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error sending get smart properties rules request.")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
func sendGetSmartPropertyRuleReq(r *gin.Engine, projectID int64, ruleID string, agent *model.Agent) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}

	url := "/projects/" + strconv.FormatUint(uint64(projectID), 10) + "/v1/smart_properties/rules/" + ruleID
	rb := C.NewRequestBuilderWithPrefix(http.MethodGet, url).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})
	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error sending get smart properties rule request.")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
func sendUpdateSmartPropertyReq(r *gin.Engine, project_id int64, ruleID string, agent *model.Agent, rules *postgres.Jsonb, type_alias string, name string, description string) *httptest.ResponseRecorder {
	payload := map[string]interface{}{
		"project_id":  project_id,
		"id":          ruleID,
		"type_alias":  type_alias,
		"rules":       rules,
		"name":        name,
		"description": description,
	}
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}

	url := "/projects/" + strconv.FormatUint(uint64(project_id), 10) + "/v1/smart_properties/rules/" + ruleID
	rb := C.NewRequestBuilderWithPrefix(http.MethodPut, url).
		WithPostParams(payload).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error sending create smart properties request.")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func sendDeleteSmartPropertyRuleReq(r *gin.Engine, projectID int64, ruleID string, agent *model.Agent) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}

	url := "/projects/" + strconv.FormatUint(uint64(projectID), 10) + "/v1/smart_properties/rules/" + ruleID
	rb := C.NewRequestBuilderWithPrefix(http.MethodDelete, url).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})
	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error sending get smart properties rule request.")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestSmartPropertyHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, agent, _ := SetupProjectWithAgentDAO()
	assert.NotNil(t, project)

	name1 := U.RandomString(8)
	name2 := U.RandomString(8)
	ruleID1 := ""
	ruleID2 := ""
	t.Run("CreateSmartPropertySuccess1", func(t *testing.T) {
		description := U.RandomString(20)
		rules := &postgres.Jsonb{json.RawMessage(`[{"value": "bangalore", "source": "all", "filters": [{"name": "campaign","property": "name","condition": "contains","logical_operator": "AND","value": "blr"}]}]`)}
		w := sendCreateSmartPropertyReq(r, project.ID, agent, rules, "campaign", name1, description)
		assert.Equal(t, http.StatusOK, w.Code)
		var result model.SmartPropertyRules
		decoder := json.NewDecoder(w.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&result); err != nil {
			assert.NotNil(t, nil, err)
		}
		ruleID1 = result.ID
	})
	t.Run("CreateSmartPropertyFailure:EmptyRule", func(t *testing.T) {
		name := U.RandomString(8)
		description := U.RandomString(20)
		rules := &postgres.Jsonb{json.RawMessage(`[]`)}
		w := sendCreateSmartPropertyReq(r, project.ID, agent, rules, "campaign", name, description)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	t.Run("CreateSmartPropertyFailure:EmptyFilters", func(t *testing.T) {
		name := U.RandomString(8)
		description := U.RandomString(20)
		rules := &postgres.Jsonb{json.RawMessage(`[{"value": "bangalore", "source": "all", "filters": []}]`)}
		w := sendCreateSmartPropertyReq(r, project.ID, agent, rules, "campaign", name, description)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	t.Run("CreateSmartPropertyFailure:EmptyValue", func(t *testing.T) {
		name := U.RandomString(8)
		description := U.RandomString(20)
		rules := &postgres.Jsonb{json.RawMessage(`[{"value": "", "source": "all", "filters": [{"name": "campaign","property": "name","condition": "contains","logical_operator": "AND","value": "blr"}]},{"value": "mumbai", "source": "all", "filters": [{"name": "campaign","property": "name","condition": "contains","logical_operator": "AND","value": "mum"}]}]`)}
		w := sendCreateSmartPropertyReq(r, project.ID, agent, rules, "campaign", name, description)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	t.Run("CreateSmartPropertyfailure:RepeatedName", func(t *testing.T) {
		description := U.RandomString(20)
		rules := &postgres.Jsonb{json.RawMessage(`[{"value": "bangalore", "source": "all", "filters": [{"name": "campaign","property": "name","condition": "contains","logical_operator": "AND","value": "blr"}]}]`)}
		w := sendCreateSmartPropertyReq(r, project.ID, agent, rules, "campaign", name1, description)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	t.Run("GetSmartPropertyRulesReq", func(t *testing.T) {
		w := sendGetSmartPropertyRulesReq(r, project.ID, agent)
		var result []model.SmartPropertyRules
		decoder := json.NewDecoder(w.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&result); err != nil {
			assert.NotNil(t, nil, err)
		}
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, len(result), 1)
		assert.Equal(t, result[0].ID, ruleID1)
	})

	t.Run("GetSmartPropertyRuleReq", func(t *testing.T) {
		w := sendGetSmartPropertyRuleReq(r, project.ID, ruleID1, agent)
		var result model.SmartPropertyRules
		decoder := json.NewDecoder(w.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&result); err != nil {
			assert.NotNil(t, nil, err)
		}
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, result.ID, ruleID1)
	})
	t.Run("CreateSmartPropertySuccess2", func(t *testing.T) {
		description := U.RandomString(20)
		rules := &postgres.Jsonb{json.RawMessage(`[{"value": "europe", "source": "all", "filters": [{"name": "campaign","property": "name","condition": "contains","logical_operator": "AND","value": "EU"}]}]`)}
		w := sendCreateSmartPropertyReq(r, project.ID, agent, rules, "campaign", name2, description)
		assert.Equal(t, http.StatusOK, w.Code)
		var result model.SmartPropertyRules
		decoder := json.NewDecoder(w.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&result); err != nil {
			assert.NotNil(t, nil, err)
		}
		ruleID2 = result.ID
	})
	t.Run("UpdateSmartPropertyRuleFailure:RepeatedName", func(t *testing.T) {
		rules := &postgres.Jsonb{json.RawMessage(`[{"value": "bangalore", "source": "all", "filters": [{"name": "campaign","property": "name","condition": "contains","logical_operator": "AND","value": "blr"}]}]`)}
		w := sendUpdateSmartPropertyReq(r, project.ID, ruleID2, agent, rules, "campaign", name1, "description")
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	t.Run("UpdateSmartPropertyRuleSuccess", func(t *testing.T) {
		rules := &postgres.Jsonb{json.RawMessage(`[{"value": "europe", "source": "all", "filters": [{"name": "campaign","property": "name","condition": "contains","logical_operator": "AND","value": "EU"}]},{"value": "north_america", "source": "all", "filters": [{"name": "campaign","property": "name","condition": "contains","logical_operator": "AND","value": "NA"}]}]`)}
		w := sendUpdateSmartPropertyReq(r, project.ID, ruleID2, agent, rules, "campaign", name2+"123", "description")
		assert.Equal(t, http.StatusOK, w.Code)
	})
	t.Run("GetSmartPropertyRuleReq:CheckUpdatedRule", func(t *testing.T) {
		w := sendGetSmartPropertyRuleReq(r, project.ID, ruleID2, agent)
		var result model.SmartPropertyRules
		decoder := json.NewDecoder(w.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&result); err != nil {
			assert.NotNil(t, nil, err)
		}
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, result.ID, ruleID2)
		assert.Equal(t, name2+"123", result.Name)
	})
	t.Run("DeleteSmartPropertyRule", func(t *testing.T) {
		w := sendDeleteSmartPropertyRuleReq(r, project.ID, ruleID1, agent)
		assert.Equal(t, http.StatusOK, w.Code)
	})
	t.Run("GetDeletedRule", func(t *testing.T) {
		w := sendGetSmartPropertyRuleReq(r, project.ID, ruleID1, agent)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
	t.Run("CreateRuleWithDeletedRuleName", func(t *testing.T) {
		description := U.RandomString(20)
		rules := &postgres.Jsonb{json.RawMessage(`[{"value": "bangalore", "source": "all", "filters": [{"name": "campaign","property": "name","condition": "contains","logical_operator": "AND","value": "blr"}]}]`)}
		w := sendCreateSmartPropertyReq(r, project.ID, agent, rules, "campaign", name1, description)
		assert.Equal(t, http.StatusOK, w.Code)
		var result model.SmartPropertyRules
		decoder := json.NewDecoder(w.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&result); err != nil {
			assert.NotNil(t, nil, err)
		}
		assert.Equal(t, name1, result.Name)
	})
}
