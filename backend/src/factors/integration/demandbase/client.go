package demandbase

import (
	U "factors/util"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

func DemandbaseHTTPRequest(baseurl, method, apiKey, clientIP string) (*http.Response, error) {

	parsedURL, err := url.Parse(baseurl)
	if err != nil {
		return nil, fmt.Errorf("invalid demandbase URL: %v", err)
	}

	query := parsedURL.Query()
	query.Add("auth", apiKey)
	query.Add("query", clientIP)
	parsedURL.RawQuery = query.Encode()

	client := &http.Client{
		Timeout: U.TimeoutOneSecond + 100*time.Millisecond,
	}

	req, err := http.NewRequest(method, parsedURL.String(), nil)
	if err != nil {
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return res, nil

}
