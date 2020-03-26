package integration

import (
	"encoding/json"
	"errors"

	C "factors/config"
	"factors/vendor_custom/machinery/v1/tasks"
)

const TypeSegment = "segment"
const TypeShopify = "shopify"

var types = [...]string{
	TypeSegment,
	TypeShopify,
}

const ProcessRequestTask = "process_integration_request"
const RequestQueue = "integration_request_queue"

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

	reqPayloadJson, err := json.Marshal(reqPayload)
	if err != nil {
		return err
	}

	queueClient := C.GetServices().QueueClient
	_, err = queueClient.SendTask(&tasks.Signature{
		Name:                 ProcessRequestTask,
		RoutingKey:           RequestQueue, // queue to send.
		RetryLaterOnPriority: true,         // allow delayed tasks to run on priority.
		Args: []tasks.Arg{
			{
				Type:  "string",
				Value: token,
			},
			{
				Type:  "string",
				Value: reqType,
			},
			{
				Type:  "string",
				Value: string(reqPayloadJson),
			},
		},
	})

	return err
}
