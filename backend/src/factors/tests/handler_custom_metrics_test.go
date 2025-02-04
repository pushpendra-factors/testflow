package tests

import (
	"encoding/json"
	C "factors/config"
	H "factors/handler"
	"factors/handler/helpers"
	"factors/model/model"
	U "factors/util"
	"fmt"
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

	name1 := U.RandomString(8)
	description1 := U.RandomString(8)

	project, agent, _ := SetupProjectWithAgentDAO()
	assert.NotNil(t, project)
	t.Run("CreateCustomMetricsSuccess", func(t *testing.T) {
		transformations := &postgres.Jsonb{json.RawMessage(`{"agFn": "sum", "agPr": "$hubspot_amount", "agPrTy": "numerical", "fil": [], "daFie": "$hubspot_datefield1"}`)}
		w := sendCreateCustomMetric(r, project.ID, agent, transformations, name1, description1, "hubspot_contacts", 1)
		assert.Equal(t, http.StatusOK, w.Code)
		var result model.CustomMetric
		decoder := json.NewDecoder(w.Body)
		if err := decoder.Decode(&result); err != nil {
			assert.NotNil(t, nil, err)
		}
	})

	t.Run("CreateCustomMetricFailure", func(t *testing.T) {
		transformations := &postgres.Jsonb{json.RawMessage(`{"agFn": "sum", "agPr": "$hubspot_amount", "agPrTy": "numerical", "fil": [], "daFie": "$hubspot_datefield2"}`)}
		w := sendCreateCustomMetric(r, project.ID, agent, transformations, name1, description1, "salesforce_users", 1)
		assert.Equal(t, http.StatusConflict, w.Code)
	})
}

func TestCreateCustomEventPostHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	name1 := U.RandomString(8)
	description1 := U.RandomString(8)
	timestamp := U.UnixTimeBeforeDuration(30 * 24 * time.Hour)
	page_url := "https://www.factors.ai/pricing"

	project, agent, _ := SetupProjectWithAgentDAO()
	assert.NotNil(t, project)
	t.Run("CreateCustomEventSuccess", func(t *testing.T) {
		transformationRaw := fmt.Sprintf(`{"agFn": "count", "agPr": "1", "agPrTy": "numerical", "fil": [], "daFie": "%d", "evNm": "%s", "en": "%s"}`, timestamp, page_url, model.QueryTypeEventsOccurrence)
		transformations := &postgres.Jsonb{RawMessage: json.RawMessage(transformationRaw)}
		w := sendCreateCustomMetric(r, project.ID, agent, transformations, name1, description1, model.EventsBasedDisplayCategory, 3)
		assert.Equal(t, http.StatusOK, w.Code)
		var result model.CustomMetric
		decoder := json.NewDecoder(w.Body)
		if err := decoder.Decode(&result); err != nil {
			assert.NotNil(t, nil, err)
		}
	})

	t.Run("CreateCustomMetricFailureDuplicateName", func(t *testing.T) {
		transformationRaw := fmt.Sprintf(`{"agFn": "count", "agPr": "1", "agPrTy": "numerical", "fil": [], "daFie": "%d", "evNm": "%s", "en": "%s"}`, timestamp, page_url, model.QueryTypeEventsOccurrence)
		transformations := &postgres.Jsonb{RawMessage: json.RawMessage(transformationRaw)}
		w := sendCreateCustomMetric(r, project.ID, agent, transformations, name1, description1, model.EventsBasedDisplayCategory, 3)
		assert.Equal(t, http.StatusConflict, w.Code)
	})
}

func TestCreateDerivedKPIPostHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, agent, _ := SetupProjectWithAgentDAO()
	assert.NotNil(t, project)

	name1 := U.RandomString(8)
	description1 := U.RandomString(8)
	t.Run("CreateCustomMetricsSuccess", func(t *testing.T) {
		transformations := &postgres.Jsonb{json.RawMessage(`{"cl":"kpi","for":"(a/b)+(a/c)","qG":[{"ca":"channels","dc":"google_ads_metrics","fil":[],"gBy":[],"me":["impressions"],"na":"a","pgUrl":"","tz":"Australia/Sydney"},{"ca":"channels","dc":"google_ads_metrics","fil":[],"gBy":[],"me":["clicks"],"na":"b","pgUrl":"","tz":"Australia/Sydney"},{"ca":"channels","dc":"google_ads_metrics","fil":[],"gBy":[],"me":["spend"],"na":"c","pgUrl":"","tz":"Australia/Sydney"}]}`)}
		w := sendCreateCustomMetric(r, project.ID, agent, transformations, name1, description1, "google_ads_metrics", 2)
		assert.Equal(t, http.StatusOK, w.Code)
		var result model.CustomMetric
		decoder := json.NewDecoder(w.Body)
		if err := decoder.Decode(&result); err != nil {
			assert.NotNil(t, nil, err)
		}
	})
	t.Run("CreateCustomMetricsFailureDuplicateName", func(t *testing.T) {
		transformations := &postgres.Jsonb{json.RawMessage(`{"cl":"kpi","for":"(a/b)+(a/c)","qG":[{"ca":"channels","dc":"google_ads_metrics","fil":[],"gBy":[],"me":["impressions"],"na":"a","pgUrl":"","tz":"Australia/Sydney"},{"ca":"channels","dc":"google_ads_metrics","fil":[],"gBy":[],"me":["clicks"],"na":"b","pgUrl":"","tz":"Australia/Sydney"},{"ca":"channels","dc":"google_ads_metrics","fil":[],"gBy":[],"me":["spend"],"na":"c","pgUrl":"","tz":"Australia/Sydney"}]}`)}
		w := sendCreateCustomMetric(r, project.ID, agent, transformations, name1, description1, "google_ads_metrics", 2)
		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("CreateCustomMetricsFailureFormulaBracesMismatch", func(t *testing.T) {
		transformations := &postgres.Jsonb{json.RawMessage(`{"cl":"kpi","for":"(a/b)+(a/c","qG":[{"ca":"channels","dc":"google_ads_metrics","fil":[],"gBy":[],"me":["impressions"],"na":"a","pgUrl":"","tz":"Australia/Sydney"},{"ca":"channels","dc":"google_ads_metrics","fil":[],"gBy":[],"me":["clicks"],"na":"b","pgUrl":"","tz":"Australia/Sydney"},{"ca":"channels","dc":"google_ads_metrics","fil":[],"gBy":[],"me":["spend"],"na":"c","pgUrl":"","tz":"Australia/Sydney"}]}`)}
		w := sendCreateCustomMetric(r, project.ID, agent, transformations, name1, description1, "google_ads_metrics", 2)
		assert.NotEqual(t, http.StatusOK, w.Code)
	})
	t.Run("CreateCustomMetricsFailureFormulaBracesWrongPlacement", func(t *testing.T) {
		transformations := &postgres.Jsonb{json.RawMessage(`{"cl":"kpi","for":"(a/b)(+a/c)","qG":[{"ca":"channels","dc":"google_ads_metrics","fil":[],"gBy":[],"me":["impressions"],"na":"a","pgUrl":"","tz":"Australia/Sydney"},{"ca":"channels","dc":"google_ads_metrics","fil":[],"gBy":[],"me":["clicks"],"na":"b","pgUrl":"","tz":"Australia/Sydney"},{"ca":"channels","dc":"google_ads_metrics","fil":[],"gBy":[],"me":["spend"],"na":"c","pgUrl":"","tz":"Australia/Sydney"}]}`)}
		w := sendCreateCustomMetric(r, project.ID, agent, transformations, name1, description1, "google_ads_metrics", 2)
		assert.NotEqual(t, http.StatusOK, w.Code)
	})
	t.Run("CreateCustomMetricsFailureFormulaSingleVariable", func(t *testing.T) {
		transformations := &postgres.Jsonb{json.RawMessage(`{"cl":"kpi","for":"a","qG":[{"ca":"channels","dc":"google_ads_metrics","fil":[],"gBy":[],"me":["impressions"],"na":"a","pgUrl":"","tz":"Australia/Sydney"}]}`)}
		w := sendCreateCustomMetric(r, project.ID, agent, transformations, name1, description1, "google_ads_metrics", 2)
		assert.NotEqual(t, http.StatusOK, w.Code)
	})
	t.Run("CreateCustomMetricsFailureFormulaVariableAndQueryMismatch1", func(t *testing.T) {
		transformations := &postgres.Jsonb{json.RawMessage(`{"cl":"kpi","for":"(a/b)","qG":[{"ca":"channels","dc":"google_ads_metrics","fil":[],"gBy":[],"me":["impressions"],"na":"a","pgUrl":"","tz":"Australia/Sydney"},{"ca":"channels","dc":"google_ads_metrics","fil":[],"gBy":[],"me":["clicks"],"na":"b","pgUrl":"","tz":"Australia/Sydney"},{"ca":"channels","dc":"google_ads_metrics","fil":[],"gBy":[],"me":["spend"],"na":"c","pgUrl":"","tz":"Australia/Sydney"}]}`)}
		w := sendCreateCustomMetric(r, project.ID, agent, transformations, name1, description1, "google_ads_metrics", 2)
		assert.NotEqual(t, http.StatusOK, w.Code)
	})
	t.Run("CreateCustomMetricsFailureFormulaVariableAndQueryMismatch2", func(t *testing.T) {
		transformations := &postgres.Jsonb{json.RawMessage(`{"cl":"kpi","for":"(a/b)*d/c","qG":[{"ca":"channels","dc":"google_ads_metrics","fil":[],"gBy":[],"me":["impressions"],"na":"a","pgUrl":"","tz":"Australia/Sydney"},{"ca":"channels","dc":"google_ads_metrics","fil":[],"gBy":[],"me":["clicks"],"na":"b","pgUrl":"","tz":"Australia/Sydney"},{"ca":"channels","dc":"google_ads_metrics","fil":[],"gBy":[],"me":["spend"],"na":"c","pgUrl":"","tz":"Australia/Sydney"}]}`)}
		w := sendCreateCustomMetric(r, project.ID, agent, transformations, name1, description1, "google_ads_metrics", 2)
		assert.NotEqual(t, http.StatusOK, w.Code)
	})
	t.Run("CreateCustomMetricsFailureGroupByPresent", func(t *testing.T) {
		transformations := &postgres.Jsonb{json.RawMessage(`{"cl":"kpi","for":"(a/b)*d/c","qG":[{"ca":"channels","dc":"google_ads_metrics","fil":[],"gBy":[{"dpNa": "campaign name", "en": "", "gr": "", "objTy": "campaign", "prDaTy": "categorical", "prNa": "campaign_name"}],"me":["impressions"],"na":"a","pgUrl":"","tz":"Australia/Sydney"},{"ca":"channels","dc":"google_ads_metrics","fil":[],"gBy":[],"me":["clicks"],"na":"b","pgUrl":"","tz":"Australia/Sydney"},{"ca":"channels","dc":"google_ads_metrics","fil":[],"gBy":[],"me":["spend"],"na":"c","pgUrl":"","tz":"Australia/Sydney"}]}`)}
		w := sendCreateCustomMetric(r, project.ID, agent, transformations, name1, description1, "google_ads_metrics", 2)
		assert.NotEqual(t, http.StatusOK, w.Code)
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
		transformations := &postgres.Jsonb{json.RawMessage(`{"agFn": "sum", "agPr": "$hubspot_amount", "agPrTy": "numerical", "fil": [], "daFie": "$hubspot_datefield1"}`)}
		w := sendCreateCustomMetric(r, project.ID, agent, transformations, name1, description1, "hubspot_contacts", 1)
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

	project, agent, _ = SetupProjectWithAgentDAO()
	assert.NotNil(t, project)
	t.Run("GetCustomMetricsSuccess", func(t *testing.T) {
		transformations := &postgres.Jsonb{json.RawMessage(`{"agFn": "sum", "agPr": "$salesforce_id", "agPrTy": "numerical", "fil": [], "daFie": "$salesforce_datefield1"}`)}
		w := sendCreateCustomMetric(r, project.ID, agent, transformations, name1, description1, "salesforce_accounts", 1)
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
	description string, objectType string, queryType int) *httptest.ResponseRecorder {
	payload := map[string]interface{}{
		"name":            name,
		"description":     description,
		"transformations": transformations,
		"objTy":           objectType,
		"type_of_query":   queryType,
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

func sendCreateCustomMetricWithPercentage(r *gin.Engine, project_id int64, agent *model.Agent, transformations *postgres.Jsonb, name string,
	description string, objectType string, queryType int) *httptest.ResponseRecorder {
	w := sendCreateCustomMetricWithOtherProperties(r, project_id, agent, transformations, name, description, objectType, queryType,
		model.MetricsPercentageType, "")
	return w
}

func sendCreateCustomMetricWithOtherProperties(r *gin.Engine, project_id int64, agent *model.Agent, transformations *postgres.Jsonb, name string,
	description string, objectType string, queryType int, displayResultAs string, metricType string) *httptest.ResponseRecorder {
	payload := map[string]interface{}{
		"name":              name,
		"description":       description,
		"transformations":   transformations,
		"objTy":             objectType,
		"type_of_query":     queryType,
		"display_result_as": displayResultAs,
		"metric_type":       metricType,
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
