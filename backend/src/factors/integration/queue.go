package integration

import (
	"errors"

	C "factors/config"
	"factors/util"

	log "github.com/sirupsen/logrus"
)

const TypeSegment = "segment"
const TypeShopify = "shopify"
const TypeRudderstack = "rudderstack"

var types = [...]string{
	TypeSegment,
	TypeShopify,
	TypeRudderstack,
}

const ProcessRequestTask = "process_integration_request"
const RequestQueue = "integration_request_queue_2"
const RequestQueueDuplicate = "integration_request_queue"

func isValidRequest(token, reqType string, reqPayload interface{}) bool {
	if token == "" {
		return false
	}

	if reqPayload == nil {
		return false
	}

	var valid bool
	for _, typ := range types {
		if typ == reqType {
			valid = true
			break
		}
	}

	return valid
}

func EnqueueRequest(token, reqType string, reqPayload interface{}) error {
	if !isValidRequest(token, reqType, reqPayload) {
		return errors.New("invalid request")
	}

	taskSignature, err := util.CreateTaskSignatureForQueue(ProcessRequestTask,
		RequestQueue, token, reqType, reqPayload)
	if err != nil {
		return err
	}

	queueClient := C.GetServices().QueueClient
	_, err = queueClient.SendTask(taskSignature)
	if err != nil {
		return err
	}

	if !C.IsSDKAndIntegrationRequestQueueDuplicationEnabled() {
		return nil
	}

	dupTaskSignature, err := util.CreateTaskSignatureForQueue(ProcessRequestTask,
		RequestQueueDuplicate, token, reqType, reqPayload)
	if err != nil {
		return err
	}

	duplicateQueueClient := C.GetServices().DuplicateQueueClient
	_, err = duplicateQueueClient.SendTask(dupTaskSignature)
	if err != nil {
		// Log and return duplicate task queue failure.
		log.WithField("token", token).WithField("payload", reqPayload).
			WithError(err).Error("Failed to send integration request task to the duplicate queue.")
		return nil
	}

	return nil
}
