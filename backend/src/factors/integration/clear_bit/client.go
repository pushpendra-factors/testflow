package clear_bit

import (
	U "factors/util"
	"net/http"
	"time"
)

func ClearbitHTTPRequest(clientIP string, clearbitAPIKey string) (*http.Response, error) {

	client, req, err := FormRequest(clientIP)
	if err != nil {
		return nil, err
	}
	clearbitKey := "Bearer " + clearbitAPIKey
	req.Header.Add("Authorization", clearbitKey)
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func FormRequest(clientIP string) (*http.Client, *http.Request, error) {
	baseUrl := "https://reveal.clearbit.com/v1/companies/find?ip="
	method := "GET"
	url := baseUrl + clientIP
	client := &http.Client{
		Timeout: U.TimeoutOneSecond + 100*time.Millisecond,
	}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, nil, err
	}
	return client, req, nil
}
