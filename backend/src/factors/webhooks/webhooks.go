package webhooks

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
)

func DropWebhook(url, secret string, payload interface{}) (map[string]interface{}, error) {
	if url == "" || !IsUrl(url) {
		return nil, fmt.Errorf("invalid url")
	}
	if payload == nil {
		return nil, fmt.Errorf("no payload to drop")
	}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	if secret != "" {
		h := sha256.New()
		h.Write([]byte(secret))
		request.Header.Add("factors-secret-256", base64.StdEncoding.EncodeToString(h.Sum(nil)))
	}

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		log.WithError(err).Error("failed to make request for webhook")
		return nil, err
	}
	defer resp.Body.Close()
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	response := make(map[string]interface{})
	if resp.StatusCode == 201 || resp.StatusCode == 200 {
		response["status"] = "ok"
	} else {	
		log.Error("Request:", request)
		response["error"] = string(bodyBytes)
		response["statuscode"] = resp.StatusCode
	}
	return response, nil
}

func IsUrl(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}
