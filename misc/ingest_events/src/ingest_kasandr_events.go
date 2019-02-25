package main

// Recursively reads all json files.
// Example usage on Terminal.
// export GOPATH=/Users/aravindmurthy/code/factors/misc/ingest_events
// go run ingest_kasandr_events.go --input_file=/Users/aravindmurthy/code/factors/sample_data/kasandr/sample_raw_data.csv --server=http://localhost:8080

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

var inputFileFlag = flag.String("input_file", "", "Input CSV file.")
var serverFlag = flag.String("server", "http://factors-dev.com:8080", "Server Path.")
var projectIdFlag = flag.Int("project_id", 0, "Project Id.")
var projectTokenFlag = flag.String("project_token", "", "Needs to be passed if projectId is passed.")

var clientUserIdToUserIdMap map[string]string = make(map[string]string)

func getUserId(clientUserId string, eventCreatedTime int64) (string, error) {
	userId, found := clientUserIdToUserIdMap[clientUserId]
	if !found {
		// Create a user.
		userCreatedTime := eventCreatedTime

		userRequestMap := make(map[string]interface{})
		userRequestMap["c_uid"] = clientUserId
		userRequestMap["join_timestamp"] = userCreatedTime

		reqBody, _ := json.Marshal(userRequestMap)
		url := fmt.Sprintf("%s/sdk/user/identify", *serverFlag)
		client := &http.Client{}
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
		req.Header.Add("Authorization", *projectTokenFlag)
		resp, err := client.Do(req)
		// always close the response-body, even if content is not required
		defer resp.Body.Close()
		if err != nil {
			log.Fatal(fmt.Sprintf("Http Post user creation failed. Url: %s, reqBody: %s, response: %+v", url, reqBody, resp))
			return "", err
		}
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

func lineIngest(line string) {
	log.Info(fmt.Sprintf("Processing event: %s", line))
	fields := strings.Fields(line)
	clientUserId := fields[0]
	offerId := fields[1]
	country := fields[2]
	category := fields[3]
	merchant := fields[4]
	eventCreatedTime1 := fields[5]
	eventCreatedTime2 := fields[6]
	eventCreatedTimeString := fmt.Sprintf("%sT%sZ", eventCreatedTime1, eventCreatedTime2)
	eventCreatedTime, _ := time.Parse(time.RFC3339, eventCreatedTimeString)
	eventType := fields[7]

	userId, err := getUserId(clientUserId, eventCreatedTime.Unix())
	if err != nil {
		log.Fatal("UserId not found.")
		return
	}

	eventName := ""
	if strings.Compare(eventType, "0") == 0 {
		eventName = "view_" + category
	} else {
		eventName = "click_" + category
	}

	eventRequestMap := make(map[string]interface{})
	eventRequestMap["event_name"] = eventName
	eventRequestMap["user_id"] = userId
	eventRequestMap["timestamp"] = eventCreatedTime.Unix()

	eventRequestPropertiesMap := make(map[string]interface{})
	eventRequestPropertiesMap["merchant"] = merchant
	eventRequestPropertiesMap["offer_id"] = offerId
	eventRequestMap["event_properties"] = eventRequestPropertiesMap

	userRequestPropertiesMap := make(map[string]interface{})
	userRequestPropertiesMap["$country"] = country
	eventRequestMap["user_properties"] = userRequestPropertiesMap

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
	file, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		if lineNumber == 1 {
			// Ignoring first line.
			continue
		}
		line := scanner.Text()
		lineIngest(line)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	flag.Parse()

	if *projectIdFlag <= 0 || *projectTokenFlag == "" {
		log.Printf("Missing args ProjectId: %d, projectToken: %s", *projectIdFlag, *projectTokenFlag)
		os.Exit(1)
	}

	fileIngest(*inputFileFlag)
	log.Info(fmt.Sprintf("Ingested events to project Ecommerce-Sample with id: %d", *projectIdFlag))
}
