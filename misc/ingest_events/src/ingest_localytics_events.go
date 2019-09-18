package main

// Recursively reads all json files.
// Example usage on Terminal.
// export PATH_TO_FACTORS=~/repos
// cd $PATH_TO_FACTORS/factors/misc/ingest_events/src
// export GOPATH=$PATH_TO_FACTORS/factors/misc/ingest_events
// go run ingest_localytics_events.go --input_file=/usr/local/var/factors/localytics_data/data.json --server=http://factors-dev.com:8080  --server=http://localhost:8080 --project_id=282 --project_token=c3wycax2nd7fiqz32jqfiw21zpdhychz

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	U "./util"
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

func translateEvent(eventMap map[string]interface{}) map[string]interface{} {
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
	}

	eventTimestampFloat, found := eventMap["client_time"].(float64)
	if !found {
		log.Fatal("Missing timestamp in event.")
	}
	eventTimestamp := int64(eventTimestampFloat)

	userId, err := getUserId(clientUserId, eventTimestamp)
	if err != nil {
		log.Fatal("UserId not found.")
	}

	eventName, found := eventMap["event_name"].(string)
	if !found {
		log.Fatal("Missing EventName in event.")
	}

	eventRequestMap := make(map[string]interface{})
	eventRequestMap["event_name"] = eventName
	eventRequestMap["user_id"] = userId
	eventRequestMap["timestamp"] = eventTimestamp

	// Add customer event id if present
	// eventRequestMap["c_event_id"]= customerEventId

	// Event properties from params.
	eventPropertiesMap := make(map[string]interface{})
	userPropertiesMap := make(map[string]interface{})
	if locationInfo, ok := eventMap["location"].(map[string]interface{}); ok {
		eventPropertiesMap[U.EP_LOCATION_LATITUDE], _ = locationInfo["latitude"]
		eventPropertiesMap[U.EP_LOCATION_LONGITUDE], _ = locationInfo["longitude"]

		userPropertiesMap[U.UP_CITY], _ = locationInfo["city"]
		userPropertiesMap[U.UP_REGION], _ = locationInfo["state"]
		userPropertiesMap["zip_code"], _ = locationInfo["zip_code"]
	}
	eventPropertiesMap["category"], _ = eventMap["category"]
	eventPropertiesMap["amount"], _ = eventMap["amount"]
	userPropertiesMap["gender"], _ = eventMap["gender"]
	userPropertiesMap["age"], _ = eventMap["age"]
	userPropertiesMap["marital_status"], _ = eventMap["marital_status"]
	userPropertiesMap[U.UP_PLATFORM], _ = eventMap["device"]

	eventRequestMap["event_properties"] = eventPropertiesMap
	eventRequestMap["user_properties"] = userPropertiesMap

	return eventRequestMap
}

func ingestEvents(events []map[string]interface{}) error {
	reqBody, err := json.Marshal(events)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/sdk/event/track/bulk", *serverFlag)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", *projectTokenFlag)
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(fmt.Sprintf("Http Post event creation failed. Url: %s, reqBody: %s", url, reqBody))
		return err
	}
	defer resp.Body.Close()

	return nil
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

	maxBatchSize := 1000
	size := len(eventsData)

	noOfBatches := size / maxBatchSize

	if size%maxBatchSize != 0 {
		noOfBatches++
	}

	log.Infof("*** Total No of events to Ingest: %d", size)

	log.WithFields(log.Fields{
		"Batches":      noOfBatches,
		"MaxBatchSize": maxBatchSize,
	}).Info("No Of Batches")

	for batch := 0; batch < noOfBatches; batch++ {
		start := batch * maxBatchSize
		end := start + min(size-start, maxBatchSize)
		events := eventsData[start:end]
		log.WithFields(log.Fields{
			"start": start,
			"end":   end,
			"batch": batch,
		}).Info("Ingesting Batch")

		translatedEvents := make([]map[string]interface{}, 0, 0)
		for _, event := range events {
			translatedEvents = append(translatedEvents, translateEvent(event))
		}

		err := ingestEvents(translatedEvents)
		if err != nil {
			log.WithError(err).Fatal("Error ingesting events")
		}

	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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
