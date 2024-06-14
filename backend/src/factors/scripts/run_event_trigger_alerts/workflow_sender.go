package main

import (
	"factors/cache"
	"factors/integration/linkedin_capi"
	"factors/integration/paragon"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func ProcessWorkflow(key *cache.Key, cachedWorkflow *model.CachedEventTriggerAlert,
	workflowID string, retry bool, sendTo string) (totalSuccess bool, partialSuccess bool, sendReport SendReportLogCount) {
	logCtx := log.WithFields(log.Fields{
		"key":             key,
		"cached_workflow": cachedWorkflow,
		"workflow_id":     workflowID,
		"retry":           retry,
		"send_to":         sendTo,
	})

	if cachedWorkflow.IsLinkedInCAPI {
		totalSuccess, _, sendReport = SendHelperForLinkedInCAPI(key, cachedWorkflow, workflowID, false, "", *logCtx)
	} else {
		totalSuccess, _, sendReport = SendHelperForParagonWorkflow(key, cachedWorkflow, workflowID, false, "", *logCtx)
	}

	return totalSuccess, partialSuccess, sendReport
}

func SendHelperForParagonWorkflow(key *cache.Key, cachedWorkflow *model.CachedEventTriggerAlert,
	workflowID string, retry bool, sendTo string, logCtx log.Entry) (totalSuccess bool, partialSuccess bool, sendReport SendReportLogCount) {

	errMessage := make([]string, 0)
	deliveryFailures := make([]string, 0)
	rejectedQueue := false

	if sendTo == "RejectedQueue" {
		rejectedQueue = true
	}

	url := cachedWorkflow.Message.Message //Storing the url to be hit in message for workflow
	var response = make(map[string]interface{})

	response, err := paragon.SendPayloadToParagonWorkflow(key.ProjectID, url, cachedWorkflow)
	if err != nil {
		logCtx.WithFields(log.Fields{"server_response": response}).
			WithError(err).Error("Workflow  failure")
	}
	logCtx.WithField("cached_workflow", cachedWorkflow).WithField("response", response).Info("Webhook dropped for alert.")

	stat := response["status"]
	//if atleast one property field is not null in payload the payload is considered not null
	isPayloadNull := true
	for _, val := range cachedWorkflow.Message.MessageProperty {
		if val != nil {
			isPayloadNull = false
		}
	}
	log.WithFields(log.Fields{
		"project_id":      key.ProjectID,
		"alert_id":        workflowID,
		"mode":            WEBHOOK,
		"retry":           retry,
		"is_success":      stat == "success",
		"tag":             "alert_tracker",
		"is_payload_null": isPayloadNull,
		"is_workflow":     cachedWorkflow.IsWorkflow,
	}).Info("ALERT TRACKER.")

	if stat != "success" {
		log.WithField("status", stat).WithField("response", response).Error("Workflow error details")
		sendReport.WebhookFail++
		errMessage = append(errMessage, fmt.Sprintf("Webhook host reported %v error", response["error"]))
		deliveryFailures = append(deliveryFailures, WEBHOOK)

	} else {
		sendReport.WebhookSuccess++
	}

	totalSuccess, partialSuccess = findTotalAndPartialSuccess(sendReport)
	// not total success means there has been atleast one failure
	if !totalSuccess {
		err := ParagonWorkflowFailureExecution(key, workflowID, deliveryFailures, errMessage, rejectedQueue, partialSuccess)
		if err != nil {
			logCtx.WithError(err).Error("failed while updating teams-fail flow")
		}
	}

	// partial success means there has been atleast one success
	if partialSuccess {
		status, err := store.GetStore().UpdateWorkflow(key.ProjectID, workflowID, "",
			map[string]interface{}{"last_workflow_triggered_at": U.TimeNowZ()})
		if status != http.StatusAccepted || err != nil {
			logCtx.WithError(err).Error("Failed to update db field")
		}
	}

	return totalSuccess, partialSuccess, sendReport
}

func ParagonWorkflowFailureExecution(key *cache.Key, workflowID string,
	deliveryFailures, errMsg []string, rejected, partialSuccess bool) error {

	logFields := log.Fields{
		"workflow_id": workflowID,
		"cache_key":   key,
	}
	logCtx := log.WithFields(logFields)

	if rejected && !partialSuccess {
		err := AddKeyToSortedSet(key, key.ProjectID, "RejectedQueue", rejected, partialSuccess)
		if err != nil {
			logCtx.WithError(err).Error("failed to put key in FailureSortedSet")
			return err
		}
	} else {
		for _, failPoint := range deliveryFailures {
			err := AddKeyToSortedSet(key, key.ProjectID, failPoint, rejected, partialSuccess)
			if err != nil {
				logCtx.WithError(err).Error("failed to put key in FailureSortedSet")
				return err
			}
		}
	}

	errDetails := model.LastFailDetails{
		FailTime: U.TimeNowZ(),
		FailedAt: deliveryFailures,
		Details:  errMsg,
	}
	errJson, err := U.EncodeStructTypeToPostgresJsonb(errDetails)
	if err != nil {
		logCtx.WithError(err).Error("failed to encode struct to jsonb")
		return err
	}

	status, err := store.GetStore().UpdateEventTriggerAlertField(key.ProjectID, workflowID,
		map[string]interface{}{"last_fail_details": errJson})
	if status != http.StatusAccepted || err != nil {
		logCtx.WithError(err).Error("Failed to update db field")
		return err
	}

	return nil
}

func SendHelperForLinkedInCAPI(key *cache.Key, cachedWorkflow *model.CachedEventTriggerAlert,
	workflowID string, retry bool, sendTo string, logCtx log.Entry) (totalSuccess bool, partialSuccess bool, sendReport SendReportLogCount) {

	logCtx.WithField("func", "SendHelperForLinkedInCAPI").Info("intiate SendHelperForLinkedInCAPI")
	errMessage := make([]string, 0)
	deliveryFailures := make([]string, 0)
	rejectedQueue := false

	if sendTo == "RejectedQueue" {
		rejectedQueue = true
	}

	config, err := store.GetStore().GetLinkedInCAPICofigByWorkflowId(key.ProjectID, workflowID)
	if err != nil {
		logCtx.WithError(err).Error("failed  to get linkedin configuration")
	}

	var linkedCAPIPayloadBatch model.BatchLinkedinCAPIRequestPayload
	linkedinCAPIPayloadString := U.GetPropertyValueAsString(cachedWorkflow.Message.MessageProperty["linkedCAPI_payload"])

	err = U.DecodeJSONStringToStructType(linkedinCAPIPayloadString, linkedCAPIPayloadBatch)
	if err != nil {
		logCtx.WithError(err).Error("failed to decode linkedin capi payload")
	}

	response, err := linkedin_capi.SendEventsToLinkedCAPI(config, linkedCAPIPayloadBatch)
	if err != nil {
		logCtx.WithFields(log.Fields{"server_response": response}).WithError(err).Error("LinkedIn CAPI Workflow failure.")
	}
	logCtx.WithField("cached_workflow", cachedWorkflow).WithField("response", response).Info("LinkedIn CAPI workflow sent.")

	stat := ""
	if response != nil {
		stat = "success"
	}

	//if atleast one property field is not null in payload the payload is considered not null
	isPayloadNull := true
	for _, val := range cachedWorkflow.Message.MessageProperty {
		if val != nil {
			isPayloadNull = false
		}
	}
	log.WithFields(log.Fields{
		"project_id":      key.ProjectID,
		"alert_id":        workflowID,
		"mode":            WEBHOOK,
		"retry":           retry,
		"is_success":      stat == "success",
		"tag":             "alert_tracker",
		"is_payload_null": isPayloadNull,
		"is_workflow":     cachedWorkflow.IsWorkflow,
		"is_linkedInCAPI": cachedWorkflow.IsLinkedInCAPI,
	}).Info("ALERT TRACKER.")

	if stat != "success" {
		log.WithField("status", stat).WithField("response", response).Error("linkedinCAPI error details")
		sendReport.WebhookFail++
		errMessage = append(errMessage, fmt.Sprintf("Webhook host reported %v error", response["error"]))
		deliveryFailures = append(deliveryFailures, WEBHOOK)

	} else {
		sendReport.WebhookSuccess++
	}

	totalSuccess, partialSuccess = findTotalAndPartialSuccess(sendReport)
	// not total success means there has been atleast one failure
	if !totalSuccess {
		err := ParagonWorkflowFailureExecution(key, workflowID, deliveryFailures, errMessage, rejectedQueue, partialSuccess)
		if err != nil {
			logCtx.WithError(err).Error("Failed while updating workflow failure flow")
		}
	}

	// partial success means there has been atleast one success
	if partialSuccess {
		status, err := store.GetStore().UpdateWorkflow(key.ProjectID, workflowID, "",
			map[string]interface{}{"last_workflow_triggered_at": U.TimeNowZ()})
		if status != http.StatusAccepted || err != nil {
			logCtx.WithError(err).Error("Failed to update db field")
		}
	}

	return totalSuccess, partialSuccess, sendReport
}
