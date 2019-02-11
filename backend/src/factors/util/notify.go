package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
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
