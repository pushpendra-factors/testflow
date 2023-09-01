package metrics

import (
	"context"
	"time"

	"contrib.go.opencensus.io/exporter/stackdriver"
	log "github.com/sirupsen/logrus"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

// All tracked metrics are to be added here.
// UnitType of the metric i.e. Incr / Count / Latency / Bytes must be prefixed with each metric name.
const (
	// Metrics to event processing as sdk integration and sdk request workers.
	IncrSDKRequestOverallCount           = "sdk_request_overall_count"
	IncrSDKRequestQueueProcessed         = "sdk_request_queue_processed"
	IncrSDKRequestQueueRetry             = "sdk_request_queue_retry"
	IncrSDKRquestQueueExcludedBot        = "sdk_request_queue_excluded_bot"
	IncrIntegrationRequestOverallCount   = "integration_request_overall_count"
	IncrIntegrationRequestQueueProcessed = "integration_request_queue_processed"
	IncrIntegrationRequestQueueRetry     = "integration_request_queue_retry"

	// Metrics to to track types of sdk requests.
	IncrSDKRequestTypeTrack                    = "sdk_request_type_track"
	IncrSDKRequestTypeAMPTrack                 = "sdk_request_type_amp_track"
	IncrSDKRequestTypeUpdateEventProperties    = "sdk_request_type_update_event_properties"
	IncrSDKRequestTypeAMPUpdateEventProperties = "sdk_request_type_amp_update_event_properties"
	IncrSDKRequestTypeAddUserProperties        = "sdk_request_type_add_user_properties"
	IncrSDKRequestTypeIdentifyUser             = "sdk_request_type_identify_user"
	IncrSDKGetSettingsTimeout                  = "sdk_get_settings_timeout"

	// Metrics to track latency of sdk requests.
	LatencySDKRequestTypeTrack                    = "sdk_request_track_latency"
	LatencySDKRequestTypeAMPTrack                 = "sdk_request_amp_track_latency"
	LatencySDKRequestTypeUpdateEventProperties    = "sdk_request_update_event_properties_latency"
	LatencySDKRequestTypeAMPUpdateEventProperties = "sdk_request_amp_update_event_properties_latency"
	LatencySDKRequestTypeAddUserProperties        = "sdk_request_add_user_properties_latency"
	LatencySDKRequestTypeIdentifyUser             = "sdk_request_identify_user_latency"
	LatencySDKRequestTypeAMPIdentifyUser          = "sdk_request_identify_amp_user_latency"

	// Metrics related to event user caching.
	IncrEventCacheCounter          = "event_cache_incr"
	IncrUserCacheCounter           = "user_cache_incr"
	IncrNewUserCounter             = "new_user_incr"
	LatencyEventCache              = "event_cache_latency"
	LatencyUserCache               = "user_cache_latency"
	LatencyNewUserCache            = "new_user_cache_latency"
	IncrEventUserCleanupCounter    = "clean_up_counter_incr"
	LatencyEventUserCleanupCounter = "clean_up_counter_latency"
	IncrGroupCacheCounter          = "group_cache_incr"
	LatencyGroupCache              = "group_cache_latency"

	// Metrics to monitor size of the database tables.
	BytesTableSizeAdwordsDocuments = "table_adwords_documents_size"
	BytesTableSizeEvents           = "table_events_size"
	BytesTableSizeHubspotDocuments = "table_hubspot_documents_size"
	BytesTableSizeUserProperties   = "table_user_properties_size"
	BytesTableSizeUsers            = "table_users_size"

	IncrUserPropertiesMergeCount = "user_properties_merge_count"
	// Metrics related to user properties merge. TODO(prateek): Can be removed later since not actively tracked.
	IncrUserPropertiesMergeSanitizeCount = "user_properties_merge_sanitize_count"
)

// Metrics types defined to be used in external calls like from data server.
const (
	MetricTypeIncr    = "incr"
	MetricTypeCount   = "count"
	MetricTypeBytes   = "bytes"
	MetricTypeLatency = "latency"
)

// PythonServerMetrics All python job related metrics to be added here.
var PythonServerMetrics = map[string]map[string]bool{
	MetricTypeIncr: {
		// Dummy metric for tests.
		"data_server_dummy_incr_metric": true,
	},
	MetricTypeCount:   {},
	MetricTypeBytes:   {},
	MetricTypeLatency: {},
}

var (
	// The task latency in milliseconds.
	latencyStats    = stats.Float64("task_latency", "The task latency in milliseconds", stats.UnitMilliseconds)
	guageStatsInt   = stats.Int64("int_counter", "The number of loop iterations", stats.UnitDimensionless)
	guageStatsFloat = stats.Float64("float_counter", "The number of loop iterations", stats.UnitDimensionless)
	bytesStatsFloat = stats.Float64("bytes_size", "Size of a table or object in bytes", stats.UnitBytes)
)

var (
	// MetricNameTag Label for the metric to be updated. To be used in filter.
	MetricNameTag, _ = tag.NewKey("metric_name")
)

var (
	latencyView = &view.View{
		Name:        "latency_view",
		Measure:     latencyStats,
		Description: "The distribution of the task latencies",

		// Bucketing is not supported in stackdriver.
		// But retain this else it fails to export metrics.
		// [>=0ms, >=100ms, >=200ms, >=400ms, >=1s, >=2s, >=4s]
		Aggregation: view.Distribution(0, 100, 200, 400, 1000, 2000, 4000),
		TagKeys:     []tag.Key{MetricNameTag},
	}

	countIntView = &view.View{
		Measure:     guageStatsInt,
		Name:        "count_int_view",
		Description: "Count int view",
		Aggregation: view.Sum(),
		TagKeys:     []tag.Key{MetricNameTag},
	}

	countFloatView = &view.View{
		Measure:     guageStatsFloat,
		Name:        "count_float_view",
		Description: "Count float view",
		Aggregation: view.Sum(),
		TagKeys:     []tag.Key{MetricNameTag},
	}

	bytesSizeViewDistributed = &view.View{
		Measure:     bytesStatsFloat,
		Name:        "bytes_size_view",
		Description: "Bytes size view",
		// Bucketing is not supported in stackdriver.
		// But retain this else it fails to export metrics.
		Aggregation: view.Distribution(0, 10, 100, 1000, 10000, 100000),
		TagKeys:     []tag.Key{MetricNameTag},
	}
)

// GenericTask Resource type for custom metrics.
// Implements interface for stackdriver's monitoredresource.
// https://cloud.google.com/monitoring/api/resources#tag_generic_task
type GenericTask struct {
	ProjectID string
	Location  string
	Namespace string
	Job       string
	TaskID    string
}

// MonitoredResource returns resource type and resource labels for GenericTask
func (gt *GenericTask) MonitoredResource() (resType string, labels map[string]string) {
	labels = map[string]string{
		"project_id": gt.ProjectID,
		"location":   gt.Location,
		"namespace":  gt.Namespace,
		"job":        gt.Job,
		"task_id":    gt.TaskID,
	}
	return "generic_task", labels
}

// InitMetrics Initializes metrics exporter to collect metrics.
func InitMetrics(env, appName, projectID, projectLocation string) *stackdriver.Exporter {
	if env == "development" {
		return nil
	}
	logCtx := log.WithField("Tag", "Metrics")
	logCtx.Info("Initializing metrics exporter ...")

	ctx := context.Background()

	if err := view.Register(latencyView, countIntView, countFloatView, bytesSizeViewDistributed); err != nil {
		log.WithError(err).Error("Failed to register the view")
		return nil
	}

	monitoredResource := GenericTask{
		ProjectID: projectID,
		Location:  projectLocation,
		Namespace: env,
		Job:       appName,
		TaskID:    "generic_task",
	}

	exporter, err := stackdriver.NewExporter(stackdriver.Options{
		ProjectID:         projectID,
		MetricPrefix:      "custom.googleapis.com/" + appName + "/",
		ReportingInterval: time.Minute,
		MonitoredResource: &monitoredResource,
		Context:           ctx,
		Timeout:           30 * time.Second,
	})
	if err != nil {
		logCtx.WithError(err).Error("Error creating exporter")
		return nil
	}
	view.SetReportingPeriod(time.Minute)

	if err := exporter.StartMetricsExporter(); err != nil {
		logCtx.WithError(err).Error("Error starting metric exporter")
		return nil
	}
	return exporter
}

// Increment Increment the given metric by 1.
func Increment(metricName string) {
	CountInt(metricName, int64(1))
}

// CountInt Reports the count value for given int Metric.
func CountInt(metricName string, count int64) {
	ctx, err := tag.New(context.Background(), tag.Upsert(MetricNameTag, metricName))
	if err != nil {
		log.WithError(err).Error("Failed to record CountInt")
		return
	}
	stats.Record(ctx, guageStatsInt.M(count))
}

// CountFloat Reports the count value for given float Metric.
func CountFloat(metricName string, count float64) {
	ctx, err := tag.New(context.Background(), tag.Upsert(MetricNameTag, metricName))
	if err != nil {
		log.WithError(err).Error("Failed to record CountFloat")
		return
	}
	stats.Record(ctx, guageStatsFloat.M(count))
}

// RecordLatency Records latency as a metric in 'ms'.
func RecordLatency(metricName string, latency float64) {
	ctx, err := tag.New(context.Background(), tag.Upsert(MetricNameTag, metricName))
	if err != nil {
		log.WithError(err).Error("Failed to record Latency")
		return
	}
	stats.Record(ctx, latencyStats.M(latency))
}

// RecordBytesSize Record size in bytes for a table or an object.
func RecordBytesSize(metricName string, bytes float64) {
	ctx, err := tag.New(context.Background(), tag.Upsert(MetricNameTag, metricName))
	if err != nil {
		log.WithError(err).Error("Failed to record Bytes")
		return
	}
	stats.Record(ctx, bytesStatsFloat.M(bytes))
}
