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

	// pagerduty-high-priority Slack channel
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
		fmt.Println("-- Notification Template -- \n")
		fmt.Println(string(jsonBody))
		// panic-notification-on-slack Slack channel
		url_Slack_TOPIC = "https://hooks.slack.com/services/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"
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

func NotifyOnPanic(taskId, env string) {

	if pe := recover(); pe != nil {
		if ne := NotifyThroughSNS(taskId, env,
			map[string]interface{}{"panic_error": pe, "stacktrace": string(debug.Stack())}); ne != nil {
			log.WithField("stack_trace", string(debug.Stack())).Error(pe)
			return
		}
	}

	if pe := recover(); pe != nil {
		if ne := NotifyThroughSlack(taskId, env,
			map[string]interface{}{"panic_error": pe, "stacktrace": string(debug.Stack())}); ne != nil {
			log.WithField("stack_trace", string(debug.Stack())).Error(pe)
			return
		}
	}
}

func NotifyOnPanicWithError(env, appName string) {
	if r := recover(); r != nil {

		buf := make([]byte, 1024)
		runtime.Stack(buf, false)

		msg := fmt.Sprintf("Panic CausedBy: %v\nStackTrace: %v\n", r, string(buf))
		details := fmt.Sprintf("Debug stack: %s", string(debug.Stack()))
		log.Errorf("Recovering from panic: %v, details: %v", msg, details)

		err := NotifyThroughSNS(appName, env, msg)
		if err != nil {
			log.WithError(err).Error("failed to send message to sns")
		}

		err = NotifyThroughSlack(appName, env, msg)
		if err != nil {
			log.WithError(err).Error("failed to send message to slack")
		}
	}
}
