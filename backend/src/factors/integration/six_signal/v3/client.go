package v3

import (
	C "factors/config"
	U "factors/util"
	"fmt"
	"net/http"
	"time"
)

const PARTNER_NAME = "factors"
const CUSTOMER_NAME = "factors-customer"

func SixSignalV3HTTPRequest(url, method, sixSignalAPIKey, clientIP string) (*http.Response, error) {

	client := &http.Client{
		Timeout: U.TimeoutOneSecond + 100*time.Millisecond,
	}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	var x6sCustomID string
	if sixSignalAPIKey == C.GetFactorsSixSignalAPIKey() {
		x6sCustomID = fmt.Sprintf("%v-%v", PARTNER_NAME, sixSignalAPIKey)
	} else {
		x6sCustomID = fmt.Sprintf("%v-%v-%v", PARTNER_NAME, CUSTOMER_NAME, sixSignalAPIKey)
	}

	sixSignalKey := "Token " + sixSignalAPIKey
	req.Header.Add("Authorization", sixSignalKey)
	req.Header.Add("X-Forwarded-For", clientIP)
	req.Header.Add("X-6s-CustomID", x6sCustomID)

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}
