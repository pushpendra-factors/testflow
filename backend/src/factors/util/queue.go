package util

import (
	"encoding/json"
	"factors/vendor_custom/machinery/v1/tasks"
)

func UseQueue(token string, queueAllowedTokens []string) bool {
	// allow all for wildcard(asterisk).
	if len(queueAllowedTokens) == 1 && queueAllowedTokens[0] == "*" {
		return true
	}

	for _, allowedToken := range queueAllowedTokens {
		if token == allowedToken {
			return true
		}
	}

	return false
}

func CreateTaskSignatureForQueue(taskName, queueName, token,
	reqType string, reqPayload interface{}) (*tasks.Signature, error) {

	reqPayloadJson, err := json.Marshal(reqPayload)
	if err != nil {
		return nil, err
	}

	task := &tasks.Signature{
		Name:                 taskName,
		RoutingKey:           queueName, // queue to send.
		RetryLaterOnPriority: true,      // allow delayed tasks to run on priority.
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
	}
	return task, err
}
