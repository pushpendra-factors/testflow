package clear_bit

import (
	U "factors/util"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const PLAN = "factors-pbc-customer"

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

func FormClearbitProvisionAccountRequest(apiUrl string, emailId string, domainName string, provisionAPIKey string) (*http.Client, *http.Request, error) {

	method := "POST"

	bearerToken := "Bearer " + provisionAPIKey

	formData := url.Values{}
	formData.Set("plans[]", PLAN)
	formData.Set("email", emailId)
	formData.Set("domain", domainName)

	body := strings.NewReader(formData.Encode())

	client := &http.Client{}
	req, err := http.NewRequest(method, apiUrl, body)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", bearerToken)

	return client, req, nil
}

func GetClearbitProvisionAccountResponse(apiUrl string, emailId string, domainName string, provisionAPIKey string) (*http.Response, error) {
	client, req, err := FormClearbitProvisionAccountRequest(apiUrl, emailId, domainName, provisionAPIKey)
	if err != nil {
		return nil, err
	}

	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return response, nil
}
