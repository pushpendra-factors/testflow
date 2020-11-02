package internal

import (
	"encoding/json"
	"fmt"
	"net/http"

	"factors/metrics"
	"factors/util"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// RecordMetricRequest Structure for recording a metric through api requests.
type RecordMetricRequest struct {
	MetricType  string      `json:"type"`
	MetricName  string      `json:"name"`
	MetricValue interface{} `json:"value"`
}

// DataServiceRecordMetricHandler For taking record metrics requests from api.
func DataServiceRecordMetricHandler(c *gin.Context) {
	r := c.Request

	var recordMetricRequest RecordMetricRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&recordMetricRequest); err != nil {
		log.WithError(err).Error("Failed to decode Json request on metrics handler.")
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "Invalid request json."})
		return
	}

	typeRegisteredMetrics, isValidType := metrics.PythonServerMetrics[recordMetricRequest.MetricType]
	if !isValidType {
		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{"error": fmt.Sprintf("Metric type '%s' is not supported", recordMetricRequest.MetricType)})
		return
	}

	if _, isTrackedMetric := typeRegisteredMetrics[recordMetricRequest.MetricName]; !isTrackedMetric {
		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{"error": fmt.Sprintf("Metric '%s' not found in tracked metrics list", recordMetricRequest.MetricName)})
		return
	}

	metricValue := util.SafeConvertToFloat64(recordMetricRequest.MetricValue)
	if metricValue == 0 && fmt.Sprintf("%v", recordMetricRequest.MetricValue) != "0" {
		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{"error": fmt.Sprintf("Invalid value '%v' for metric", recordMetricRequest.MetricValue)})
		return
	}

	switch recordMetricRequest.MetricType {
	case metrics.MetricTypeIncr:
		metrics.Increment(recordMetricRequest.MetricName)
	case metrics.MetricTypeCount:
		metrics.CountFloat(recordMetricRequest.MetricName, metricValue)
	case metrics.MetricTypeLatency:
		metrics.RecordLatency(recordMetricRequest.MetricName, metricValue)
	case metrics.MetricTypeBytes:
		metrics.RecordBytesSize(recordMetricRequest.MetricName, metricValue)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully recorded metric."})
}
