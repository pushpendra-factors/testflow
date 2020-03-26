package main

import (
	"bytes"
	"flag"
	"fmt"
	"net/http"

	"../util"

	log "github.com/sirupsen/logrus"
)

func main() {
	projectPrivateToken := flag.String("private_token", "", "")
	server := flag.String("server", "http://localhost:8085", "")
	flag.Parse()

	if *projectPrivateToken == "" {
		log.Fatal("Invalid project private token.")
	}

	if *server == "" {
		log.Fatal("Invalid server.")
	}

	// Page.
	pagePayload := fmt.Sprintf(`
	{
		"_metadata": {
			"bundled": [
			"Segment.io"
			],
			"unbundled": [
			
			]
		},
		"anonymousId": "80444c7e-1580-4d3c-a77a-2f3427ed7d97",
		"channel": "client",
		"context": {
			"active": true,
			"app": {
				"name": "InitechGlobal",
				"version": "545",
				"build": "3.0.1.545",
				"namespace": "com.production.segment"
			},
			"campaign": {
				"name": "TPS Innovation Newsletter",
				"source": "Newsletter",
				"medium": "email",
				"term": "tps reports",
				"content": "image link"
			},
			"device": {
				"id": "B5372DB0-C21E-11E4-8DFC-AA07A5B093DB",
				"advertisingId": "7A3CBEA0-BDF5-11E4-8DFC-AA07A5B093DB",
				"adTrackingEnabled": true,
				"manufacturer": "Apple",
				"model": "iPhone7,2",
				"name": "maguro",
				"type": "ios",
				"token": "ff15bc0c20c4aa6cd50854ff165fd265c838e5405bfeb9571066395b8c9da449"
			},
			"ip": "8.8.8.8",
			"library": {
				"name": "analytics.js",
				"version": "2.11.1"
			},
			"locale": "nl-NL",
			"location": {
				"city": "San Francisco",
				"country": "United States",
				"latitude": 40.2964197,
				"longitude": -76.9411617,
				"speed": 0
			},
			"network": {
				"bluetooth": false,
				"carrier": "T-Mobile NL",
				"cellular": true,
				"wifi": false
			},
			"os": {
				"name": "iPhone OS",
				"version": "8.1.3"
			},
			"page": {
				"path": "/academy/",
				"referrer": "https://google.com",
				"search": "",
				"title": "Analytics Academy",
				"url": "https://segment.com/academy/"
			},
			"referrer": {
				"id": "ABCD582CDEFFFF01919",
				"type": "dataxu"
			},
			"screen": {
				"width": 320,
				"height": 568,
				"density": 2
			},
			"groupId": "12345",
			"timezone": "Europe/Amsterdam",
			"userAgent": "Mozilla/5.0 (iPhone; CPU iPhone OS 9_1 like Mac OS X) AppleWebKit/601.1.46 (KHTML, like Gecko) Version/9.0 Mobile/13B143 Safari/601.1"
		},
		"integrations": {},
		"messageId": "ajs-%s",
		"originalTimestamp": "2019-01-08T16:22:06.053Z",
		"projectId": "Zzft38QJhB",
		"properties": {
			"path": "/segment.test.html",
			"referrer": "",
			"search": "?a=10",
			"title": "Segment Test",
			"url": "http://localhost:8090/segment.test.html?a=10"
		},
		"receivedAt": "2019-01-08T16:21:54.106Z",
		"sentAt": "2019-01-08T16:22:06.058Z",
		"timestamp": "2019-01-08T16:21:54.101Z",
		"type": "page",
		"userId": "",
		"version": "1.1"
		}
	`, util.RandomLowerAphaNumString(10))

	url := fmt.Sprintf("%s/integrations/segment", *server)
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(pagePayload)))
	if err != nil {
		log.WithError(err).Fatal("Failed to build request.")
	}
	req.Header.Set("Content-Type", "application/json") // Default header.
	req.Header.Set("Authorization", *projectPrivateToken)

	resp, err := client.Do(req)
	if err != nil {
		log.WithError(err).WithField("response", resp).Fatal("HTTP request failed.")
	}
	// always close the response-body, even if content is not required
	defer resp.Body.Close()

	log.WithField("response", resp.StatusCode).Info("Successfully sent event.")
}
