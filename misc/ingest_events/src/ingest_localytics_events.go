package main

// Recursively reads all json files.
// Example usage on Terminal.
// export GOPATH=/Users/aravindmurthy/code/factors/misc/ingest_events
// go run ingest_kasandr_events.go --input_file=/Users/aravindmurthy/code/factors/sample_data/kasandr/sample_raw_data.csv --server=http://localhost:8080

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
)

var inputFileFlag = flag.String("input_file", "", "Input json file.")
var serverFlag = flag.String("server", "http://factors-dev.com:8080", "Server Path.")
var projectIdFlag = flag.Int("project_id", 0, "Project Id.")
var projectTokenFlag = flag.String("project_token", "", "Needs to be passed if projectId is passed.")

var clientUserIdToUserIdMap map[string]string = make(map[string]string)

func getUserId(clientUserId string, eventTimestamp int64) (string, error) {
	userId, found := clientUserIdToUserIdMap[clientUserId]
	if !found {
		// Create a user.
		userRequestMap := make(map[string]interface{})
		userRequestMap["c_uid"] = clientUserId
		userRequestMap["join_timestamp"] = eventTimestamp

		reqBody, _ := json.Marshal(userRequestMap)
		url := fmt.Sprintf("%s/sdk/user/identify", *serverFlag)
		client := &http.Client{}
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
		req.Header.Add("Authorization", *projectTokenFlag)
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(fmt.Sprintf(
				"Http Post user creation failed. Url: %s, reqBody: %s, response: %+v, error: %+v", url, reqBody, resp, err))
			return "", err
		}
		// always close the response-body, even if content is not required
		defer resp.Body.Close()
		jsonResponse, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal("Unable to parse http user create response.")
			return "", err
		}
		var jsonResponseMap map[string]interface{}
		json.Unmarshal(jsonResponse, &jsonResponseMap)
		userId = jsonResponseMap["user_id"].(string)
		clientUserIdToUserIdMap[clientUserId] = userId
	}
	return userId, nil
}

func eventIngest(eventMap map[string]interface{}) {
	/*
		// Each eventMap is of below format.
		{
			"category":"Sports",
			"event_name":"View Project",
			"gender":"M",
			"age":"18-24",
			"marital_status":"married",
			"session_id":"69f62d2ae87640f5a2dde2b2e9229fe6",
			"device":"android",
			"client_time":1393632004,
			"location": {"latitude":40.189788,"city":"Lyons","state":"CO","longitude":-105.35528,"zip_code":"80540"}
		}
	*/
	log.Info(fmt.Sprintf("Processing event: %v", eventMap))

	clientUserId, found := eventMap["session_id"].(string)
	if !found {
		log.Fatal("Missing session_id in event.")
		return
	}

	eventTimestampFloat, found := eventMap["client_time"].(float64)
	if !found {
		log.Fatal("Missing timestamp in event.")
		return
	}
	eventTimestamp := int64(eventTimestampFloat)

	userId, err := getUserId(clientUserId, eventTimestamp)
	if err != nil {
		log.Fatal("UserId not found.")
		return
	}

	eventName, found := eventMap["event_name"].(string)
	if !found {
		log.Fatal("Missing EventName in event.")
		return
	}

	eventRequestMap := make(map[string]interface{})
	eventRequestMap["event_name"] = eventName
	eventRequestMap["user_id"] = userId
	eventRequestMap["timestamp"] = eventTimestamp

	// Event properties from params.
	eventPropertiesMap := make(map[string]interface{})
	userPropertiesMap := make(map[string]interface{})
	if locationInfo, ok := eventMap["location"].(map[string]interface{}); ok {
		eventPropertiesMap["$locationLat"], _ = locationInfo["latitude"]
		eventPropertiesMap["$locationLng"], _ = locationInfo["longitude"]

		userPropertiesMap["$city"], _ = locationInfo["city"]
		userPropertiesMap["$region"], _ = locationInfo["state"]
		userPropertiesMap["zip_code"], _ = locationInfo["zip_code"]
	}
	eventPropertiesMap["category"], _ = eventMap["category"]
	eventPropertiesMap["amount"], _ = eventMap["amount"]
	userPropertiesMap["gender"], _ = eventMap["gender"]
	userPropertiesMap["age"], _ = eventMap["age"]
	userPropertiesMap["marital_status"], _ = eventMap["marital_status"]
	userPropertiesMap["$platform"], _ = eventMap["device"]

	eventRequestMap["event_properties"] = eventPropertiesMap
	eventRequestMap["user_properties"] = userPropertiesMap

	reqBody, _ := json.Marshal(eventRequestMap)
	url := fmt.Sprintf("%s/sdk/event/track", *serverFlag)
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	req.Header.Add("Authorization", *projectTokenFlag)
	resp, err := client.Do(req)
	// always close the response-body, even if content is not required
	defer resp.Body.Close()
	if err != nil {
		log.Fatal(fmt.Sprintf("Http Post event creation failed. Url: %s, reqBody: %s", url, reqBody))
		return
	}
}

func fileIngest(filepath string) {
	dataBytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Fatal(err)
	}

	var eventsJson map[string][]map[string]interface{}
	if err := json.Unmarshal(dataBytes, &eventsJson); err != nil {
		log.Fatal(fmt.Sprintf("Unable to unmarshal file to json in file %s : %s", filepath, dataBytes))
	}
	eventsData, ok := eventsJson["data"]
	if !ok {
		log.Fatal(fmt.Sprintf("No data in json file %s : %s", filepath, eventsJson))
	}

	for _, eventMap := range eventsData {
		eventIngest(eventMap)
	}
}

func main() {
	flag.Parse()

	if *projectIdFlag <= 0 || *projectTokenFlag == "" {
		log.Printf("Missing args ProjectId: %d, projectToken: %s", *projectIdFlag, *projectTokenFlag)
		os.Exit(1)
	}

	fileIngest(*inputFileFlag)
	log.Info(fmt.Sprintf("Ingested events to project Localytics-Challenge with id: %d", *projectIdFlag))
}
