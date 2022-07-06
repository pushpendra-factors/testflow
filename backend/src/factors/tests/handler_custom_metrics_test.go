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

func TestCustomMetricsPostHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, agent, _ := SetupProjectWithAgentDAO()
	assert.NotNil(t, project)

	name1 := U.RandomString(8)
	description1 := U.RandomString(8)
	t.Run("CreateCustomMetricsSuccess", func(t *testing.T) {
		transformations := &postgres.Jsonb{json.RawMessage(`{"agFn": "SUM", "agPr": "$hubspot_amount", "agPrTy": "categorical", "fil": [], "daFie": "$hubspot_datefield1"}`)}
		w := sendCreateCustomMetric(r, project.ID, agent, transformations, name1, description1, "hubspot_contacts")
		assert.Equal(t, http.StatusOK, w.Code)
		var result model.CustomMetric
		decoder := json.NewDecoder(w.Body)
		if err := decoder.Decode(&result); err != nil {
			assert.NotNil(t, nil, err)
		}
	})

}

func TestCustomMetricsGetHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, agent, _ := SetupProjectWithAgentDAO()
	assert.NotNil(t, project)

	name1 := U.RandomString(8)
	description1 := U.RandomString(8)
	t.Run("GetCustomMetricsSuccess", func(t *testing.T) {
		transformations := &postgres.Jsonb{json.RawMessage(`{"agFn": "SUM", "agPr": "$hubspot_amount", "agPrTy": "categorical", "fil": [], "daFie": "$hubspot_datefield1"}`)}
		w := sendCreateCustomMetric(r, project.ID, agent, transformations, name1, description1, "hubspot_contacts")
		assert.Equal(t, http.StatusOK, w.Code)
		w1 := sendGetCustomMetrics(r, project.ID, agent)
		assert.Equal(t, http.StatusOK, w1.Code)
		var result []model.CustomMetric
		decoder := json.NewDecoder(w1.Body)
		if err := decoder.Decode(&result); err != nil {
			assert.NotNil(t, nil, err)
		}
		assert.Equal(t, http.StatusOK, w1.Code)
		assert.Equal(t, len(result), 1)

	})
}

func sendCreateCustomMetric(r *gin.Engine, project_id int64, agent *model.Agent, transformations *postgres.Jsonb, name string,
	description string, objectType string) *httptest.ResponseRecorder {
	payload := map[string]interface{}{
		"name":            name,
		"description":     description,
		"transformations": transformations,
		"objTy":           objectType,
	}
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}
	url := "/projects/" + strconv.FormatUint(uint64(project_id), 10) + "/v1/custom_metrics"
	rb := C.NewRequestBuilderWithPrefix(http.MethodPost, url).
		WithPostParams(payload).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error sending create custom metrics request.")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func sendGetCustomMetrics(r *gin.Engine, project_id int64, agent *model.Agent) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}
	url := "/projects/" + strconv.FormatUint(uint64(project_id), 10) + "/v1/custom_metrics"
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
