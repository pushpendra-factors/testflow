package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	C "factors/config"
	"factors/integration"
	"factors/model/store"
	mqlStore "factors/model/store/memsql"
	"factors/sdk"
	U "factors/util"
	"fmt"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
)

func Monitoring(c *gin.Context) {
	if C.GetConfig().MonitoringAPIToken == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Monitoring API is not enabled."})
	}

	sqlAdminSlowQueries, factorsSlowQueries, err := store.GetStore().MonitorSlowQueries()
	if err != nil {
		log.WithError(err).Error("Failed to run monitoring query.")
	}

	var nodeUsageStatsWithErrors mqlStore.MemSQLNodeUsageStatsWithErrors
	if C.UseMemSQLDatabaseStore() {
		nodeUsageStatsWithErrors = mqlStore.GetStore().MonitorMemSQLDiskUsage()
	}

	delayedTaskCount, sdkQueueLength, integrationQueueLength,
		isQueueDuplicationEnabled, dupDelayedTaskCount, dupSDKQueueLength,
		dupIntegrationQueueLength, sdkBody, message,
		isFailure := CheckSDKAndIntegrationProcessing(C.GetConfig().DelayedTaskThreshold,
		C.GetConfig().SdkQueueThreshold, C.GetConfig().IntegrationQueueThreshold)
	// Should not proceed with success ping, incase of failure.
	if isFailure {
		return
	}

	tableSizes := store.GetStore().CollectTableSizes()
	monitoringPayload := map[string]interface{}{
		"factors_slow_queries_count":                     len(factorsSlowQueries),
		"sqlAdmin_slow_queries_count":                    len(sqlAdminSlowQueries),
		"delayed_task_count":                             delayedTaskCount,
		"sdk_queue_length":                               sdkQueueLength,
		"integration_queue_length":                       integrationQueueLength,
		"isQueue_duplication_enabled":                    isQueueDuplicationEnabled,
		"dup_delayed_task_count":                         dupDelayedTaskCount,
		"dup_sdk_queue_length":                           dupSDKQueueLength,
		"dup_integration_queue_length":                   dupIntegrationQueueLength,
		"table_sizes":                                    tableSizes,
		"delayed_task_exceeded_threshold":                delayedTaskCount > C.GetConfig().DelayedTaskThreshold,
		"dup_delayed_taskCount_exceeded_threshold":       dupDelayedTaskCount > C.GetConfig().DelayedTaskThreshold,
		"sdk_queue_length_exceeded_threshold":            sdkQueueLength > C.GetConfig().SdkQueueThreshold,
		"dupSdk_queueLength_exceeded_threshold":          dupSDKQueueLength > C.GetConfig().SdkQueueThreshold,
		"integration_queueLength_exceeded_threshold":     integrationQueueLength > C.GetConfig().IntegrationQueueThreshold,
		"dup_integration_queueLength_exceeded_threshold": dupIntegrationQueueLength > C.GetConfig().IntegrationQueueThreshold,
		"sdk_body_length_check":                          len(sdkBody) < 20000, // Approx file size of 20k. false if less than that.
		"sdk_assets_url_status":                          message,

		// Using z_ to push the fields to end of the response. Limited to 100 queries.
		"z_factors_slow_queries":   factorsSlowQueries[:U.MinInt(100, len(factorsSlowQueries))],
		"z_sql_admin_slow_queries": sqlAdminSlowQueries[:U.MinInt(100, len(sqlAdminSlowQueries))],
	}
	if C.UseMemSQLDatabaseStore() {
		monitoringPayload["memsqlNodeUsageStats"] = nodeUsageStatsWithErrors
	}

	c.JSON(http.StatusOK, monitoringPayload)
}

func CheckSDKAndIntegrationProcessing(delayedTaskThreshold, sdkQueueThreshold, integrationQueueThreshold int) (
	int, int, int, bool, int, int, int, []byte, string, bool) {

	queueClient := C.GetServices().QueueClient
	duplicateQueueClient := C.GetServices().DuplicateQueueClient

	delayedTaskCount, err := queueClient.GetBroker().GetDelayedTasksCount()
	if err != nil {
		log.WithError(err).Error("Failed to get delayed task count from redis")
	}

	var dupDelayedTaskCount int
	if C.IsQueueDuplicationEnabled() {
		dupDelayedTaskCount, err = duplicateQueueClient.GetBroker().GetDelayedTasksCount()
		if err != nil {
			log.WithError(err).Error("Failed to get delayed task count from duplicate queue redis.")
		}
	}

	sdkQueueLength, err := queueClient.GetBroker().GetQueueLength(sdk.RequestQueue)
	if err != nil {
		log.WithError(err).Error("Failed to get sdk_request_queue length")
	}

	var dupSdkQueueLength int
	if C.IsQueueDuplicationEnabled() {
		dupSdkQueueLength, err = duplicateQueueClient.GetBroker().GetQueueLength(sdk.RequestQueueDuplicate)
		if err != nil {
			log.WithError(err).Error("Failed to get duplicate sdk_request_queue length")
		}
	}

	integrationQueueLength, err := queueClient.GetBroker().GetQueueLength(integration.RequestQueue)
	if err != nil {
		log.WithError(err).Error("Failed to get integration_request_queue length")
	}

	var dupIntegrationQueueLength int
	if C.IsQueueDuplicationEnabled() {
		dupIntegrationQueueLength, err = queueClient.GetBroker().GetQueueLength(integration.RequestQueueDuplicate)
		if err != nil {
			log.WithError(err).Error("Failed to get duplicate integration_request_queue length")
		}
	}

	res, err := http.Get(C.SDKAssetsURL)
	sdkBody, err := ioutil.ReadAll(res.Body)
	var message string
	if err != nil || res.StatusCode != http.StatusOK {

		if res == nil {
			message = fmt.Sprintf("Error '%s' and no response on getting SDK from %s", err.Error(), C.SDKAssetsURL)
		} else {
			message = fmt.Sprintf("Error '%s', Code '%d' on getting SDK from %s", err.Error(), res.StatusCode, C.SDKAssetsURL)
		}

		return delayedTaskCount, sdkQueueLength, integrationQueueLength,
			C.IsQueueDuplicationEnabled(), dupDelayedTaskCount, dupSdkQueueLength, dupIntegrationQueueLength, sdkBody, message,
			true
	}

	if err != nil || string(sdkBody[0:12]) != "var factors=" {
		return delayedTaskCount, sdkQueueLength, integrationQueueLength,
			C.IsQueueDuplicationEnabled(), dupDelayedTaskCount, dupSdkQueueLength, dupIntegrationQueueLength, sdkBody, message,
			true
	}

	return delayedTaskCount, sdkQueueLength, integrationQueueLength,
		C.IsQueueDuplicationEnabled(), dupDelayedTaskCount, dupSdkQueueLength, dupIntegrationQueueLength, sdkBody, message,
		false
}
