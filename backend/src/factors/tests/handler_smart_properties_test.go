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

func sendCreateSmartPropertiesReq(r *gin.Engine, project_id uint64, agent *model.Agent, rules *postgres.Jsonb, type_alias string, name string, description string) *httptest.ResponseRecorder {
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

	url := "http://localhost:8080/projects/" + strconv.FormatUint(uint64(project_id), 10) + "/v1/smart_properties"
	log.Info("hello123 ", url)
	rb := U.NewRequestBuilder(http.MethodPost, url).
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

func TestCreateSmartPropertiesHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, agent, _ := SetupProjectWithAgentDAO()
	assert.NotNil(t, project)

	t.Run("CreateSmartPropertiesSuccess", func(t *testing.T) {
		name := U.RandomString(8)
		description := U.RandomString(20)
		rules := &postgres.Jsonb{json.RawMessage(`[{"value": "bangalore", "source": "all", "filters": [{"name": "campaign","property": "name","condition": "contains","logical_operator": "AND","value": "blr"}]}]`)}
		w := sendCreateSmartPropertiesReq(r, project.ID, agent, rules, "campaign", name, description)
		assert.Equal(t, http.StatusCreated, w.Code)
	})
	t.Run("CreateSmartPropertiesFailure:EmptyRule", func(t *testing.T) {
		name := U.RandomString(8)
		description := U.RandomString(20)
		rules := &postgres.Jsonb{json.RawMessage(`[]`)}
		w := sendCreateSmartPropertiesReq(r, project.ID, agent, rules, "campaign", name, description)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	t.Run("CreateSmartPropertiesFailure:EmptyFilters", func(t *testing.T) {
		name := U.RandomString(8)
		description := U.RandomString(20)
		rules := &postgres.Jsonb{json.RawMessage(`[{"value": "bangalore", "source": "all", "filters": []}]`)}
		w := sendCreateSmartPropertiesReq(r, project.ID, agent, rules, "campaign", name, description)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	t.Run("CreateSmartPropertiesFailure:EmptyValue", func(t *testing.T) {
		name := U.RandomString(8)
		description := U.RandomString(20)
		rules := &postgres.Jsonb{json.RawMessage(`[{"value": "", "source": "all", "filters": [{"name": "campaign","property": "name","condition": "contains","logical_operator": "AND","value": "blr"}]},{"value": "mumbai", "source": "all", "filters": [{"name": "campaign","property": "name","condition": "contains","logical_operator": "AND","value": "mum"}]}]`)}
		w := sendCreateSmartPropertiesReq(r, project.ID, agent, rules, "campaign", name, description)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
