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

func NotifyOnPanic(taskId, env string) {
	if pe := recover(); pe != nil {
		if ne := NotifyThroughSNS(taskId, env,
			map[string]interface{}{"panic_error": pe, "stacktrace": string(debug.Stack())}); ne != nil {
			log.Fatal(ne, pe) // using fatal to avoid panic loop.
		}

		log.Fatal(pe) // using fatal to avoid panic loop.
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
	}
}
