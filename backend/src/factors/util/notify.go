package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"runtime/debug"

	log "github.com/sirupsen/logrus"
)

// sends email notification to the team.
const url_SNS_TOPIC = "https://fjnvg9a8wi.execute-api.us-east-1.amazonaws.com/v1/notify"

// NotifyThroughSNS - Send email notification to the team.
func NotifyThroughSNS(source, env, message interface{}) error {
	if env != "staging" && env != "production" && env != "development" {
		return fmt.Errorf("nofitication skipped for env %s", env)
	}

	body := map[string]interface{}{
		"source":  source,
		"env":     env,
		"message": message,
	}

	if env == "staging" {
		log.WithFields(log.Fields{"message": message, "source": source}).
			Info("Notification.")
		return nil
	}

	jsonBody, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		return err
	}

	if env == "development" {
		fmt.Println("-- Notification Template -- \n")
		fmt.Println(string(jsonBody))
		return nil
	}

	req, err := http.NewRequest("POST", url_SNS_TOPIC, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	} else if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("sns return non 200 status: %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	return nil
}

// NotifyThroughSlack - Send slack notification to the team.
func NotifyThroughSlack(source, env, message interface{}) error {
	if env != "staging" && env != "production" && env != "development" {
		return fmt.Errorf("nofitication skipped for env %s", env)
	}

	body := map[string]interface{}{
		"blocks": []interface{}{
			map[string]interface{}{
				"type": "section",
				"fields": []interface{}{
					map[string]interface{}{
						"type": "mrkdwn",
						"text": "*Source*: " + fmt.Sprintf("%v", source),
					},
					map[string]interface{}{
						"type": "mrkdwn",
						"text": "*Environment*: " + fmt.Sprintf("%v", env),
					},
				},
			},
			map[string]interface{}{
				"type": "section",
				"fields": []interface{}{
					map[string]interface{}{
						"type": "mrkdwn",
						"text": "*Error message*\n" + fmt.Sprintf("%v", message),
					},
				},
			},
		},
	}

	// Send to slack channel #panic-production
	url_Slack_TOPIC := "https://hooks.slack.com/services/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"

	if env == "staging" {
		log.WithFields(log.Fields{"message": message, "source": source}).Info("Notification.")
		// panic-notification-on-slack Slack channel
		url_Slack_TOPIC = "https://hooks.slack.com/services/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"
	}

	jsonBody, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		return err
	}

	if env == "development" {
		// Not sent to slack channel from development.
		fmt.Println("-- Notification Template -- \n")
		fmt.Println(string(jsonBody))
	}

	req, err := http.NewRequest("POST", url_Slack_TOPIC, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	} else if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("slack return non 200 status: %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	return nil
}

func notifyOnPanicWithErrorLog(appName, env string, recoveredFrom interface{}) {
	buf := make([]byte, 1024)
	runtime.Stack(buf, false)

	log.WithField("panic_message", fmt.Sprintf("%+v", recoveredFrom)).
		WithField("stack_trace", string(buf)).
		WithField("debug_stack", string(debug.Stack())).
		Error("Panic Recovered.")

	msgWithTrace := fmt.Sprintf("Panic CausedBy: %v\nStackTrace: %v\n", recoveredFrom, string(buf))
	err := NotifyThroughSNS(appName, env, msgWithTrace)
	if err != nil {
		log.WithError(err).Error("Failed to send panic message to SNS.")
	}
}

func NotifyOnPanic(taskId, env string) {
	if recoveredFrom := recover(); recoveredFrom != nil {
		notifyOnPanicWithErrorLog(taskId, env, recoveredFrom)
	}
}

func NotifyOnPanicWithError(env, appName string) {
	if recoveredFrom := recover(); recoveredFrom != nil {
		notifyOnPanicWithErrorLog(appName, env, recoveredFrom)
	}
}
