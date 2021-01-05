package tests

import (
	C "factors/config"
	H "factors/handler"
	"factors/handler/helpers"
	"factors/metrics"
	M "factors/model"
	U "factors/util"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func sendRecordMetricsRequest(r *gin.Engine, agent *M.Agent, metricType, metricName string, metricValue float64) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}

	payload := map[string]interface{}{
		"name":  metricName,
		"type":  metricType,
		"value": metricValue,
	}

	rb := U.NewRequestBuilder(http.MethodPost, "http://localhost:8089/data_service/metrics").
		WithPostParams(payload).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error sending metrics requests to data server.")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestRecordIncrementMetricType(t *testing.T) {
	r := gin.Default()
	H.InitDataServiceRoutes(r)

	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)
	assert.NotNil(t, agent)

	registeredMetricName := "data_server_dummy_incr_metric" // Type incr.
	registeredMetricType := metrics.MetricTypeIncr

	// For unknow metric type, it should fail.
	w := sendRecordMetricsRequest(r, agent, U.RandomLowerAphaNumString(5), registeredMetricName, 1)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// For random untracked metric, it should fail.
	w = sendRecordMetricsRequest(r, agent, registeredMetricType, U.RandomLowerAphaNumString(10), 1)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// For wrong operation on metric type it should fail.
	w = sendRecordMetricsRequest(r, agent, metrics.MetricTypeCount, registeredMetricName, 1)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Success on correctly tracked metric.
	w = sendRecordMetricsRequest(r, agent, registeredMetricType, registeredMetricName, 1)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAddDocument(t *testing.T) {

	r := gin.Default()
	H.InitDataServiceRoutes(r)

	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)
	assert.NotNil(t, agent)

	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}

	value := map[string]interface{}{"id": 1, "campaign_id": 2, "campaign_name": "test1", "name": "name_test"}
	jsonb, err := U.EncodeToPostgresJsonb(&value)
	assert.Nil(t, err)

	payload := map[string]interface{}{
		"project_id":      project.ID,
		"customer_acc_id": "1",
		"value":           jsonb,
		"type_alias":      "ad_groups",
		"timestamp":       20201212,
	}

	rb := U.NewRequestBuilder(http.MethodPost, "http://localhost:8089/data_service/adwords/documents/add").
		WithPostParams(payload).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error in adding adwords document to the data server.")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}
