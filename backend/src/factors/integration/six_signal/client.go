package six_signal

import (
	U "factors/util"
	"net/http"
	"time"
)

func SixSignalHTTPRequest(clientIP string, sixSignalAPIKey string) (*http.Response, error) {

	client, req, err := FormRequest()
	if err != nil {
		return nil, err
	}
	sixSignalKey := "Token " + sixSignalAPIKey
	req.Header.Add("Authorization", sixSignalKey)
	req.Header.Add("X-Forwarded-For", clientIP)

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func FormRequest() (*http.Client, *http.Request, error) {
	url := "https://epsilon.6sense.com/v1/company/details"
	method := "GET"

	client := &http.Client{
		Timeout: U.TimeoutOneSecond + 100*time.Millisecond,
	}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, nil, err
	}
	return client, req, nil
}
