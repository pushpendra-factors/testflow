package main

// Recursively reads all json files.
// Example usage on Terminal.
// export GOPATH=/Users/aravindmurthy/code/factors/misc/ingest_events
// go run ingest_events.go --input_dir=/Users/aravindmurthy/events --server=http://localhost:8080

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

var inputDirFlag = flag.String("input_dir", "", "Input Directory with json files.")
var serverFlag = flag.String("server", "http://factors-dev.com:8080", "Server Path.")
var projectIdFlag = flag.Int("project_id", 0, "Project Id.")
var projectTokenFlag = flag.String("project_token", "", "Needs to be passed if projectId is passed.")

var clientUserIdKey string = "user_id"
var sessionIdKey string = "session_id"
var clientUserCreationTimeKey string = "user_creation_time"
var eventNameKey string = "event_type"
var eventCreationTimeKey string = "event_time"

var clientUserIdToUserIdMap map[string]string = make(map[string]string)

func getUserId(clientUserId string, eventMap map[string]interface{}) (string, error) {
	userId, found := clientUserIdToUserIdMap[clientUserId]
	if !found {
		// Create a user.
		var userCreatedTimeString string
		userCreatedTimeData, found := eventMap[clientUserCreationTimeKey]
		if userCreatedTimeData != nil && found {
			userCreatedTimeString = userCreatedTimeData.(string)
		} else {
			// If user created time is absent use eventCreatedTime.
			userCreatedTimeString, _ = eventMap[eventCreationTimeKey].(string)
		}
		userCreatedTimeString = strings.Replace(userCreatedTimeString, " ", "T", -1)
		userCreatedTimeString = strings.Replace(userCreatedTimeString, " ", "T", -1)
		userCreatedTimeString = fmt.Sprintf("%sZ", userCreatedTimeString)
		userCreatedTime, _ := time.Parse(time.RFC3339, userCreatedTimeString)

		userRequestMap := make(map[string]interface{})
		userRequestMap["c_uid"] = clientUserId
		userRequestMap["join_timestamp"] = userCreatedTime.Unix()

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

func lineIngest(eventMap map[string]interface{}) {
	log.Info(fmt.Sprintf("Processing event: %v", eventMap))
	clientUserId, found := eventMap[clientUserIdKey].(string)
	if !found {
		// If user_id is missing use session_id as user_id.
		clientUserIdFloat, found := eventMap[sessionIdKey].(float64)
		if !found {
			log.Fatal("Missing User Id and session Id in event.")
			return
		}
		clientUserId = fmt.Sprintf("%f", clientUserIdFloat)
	}
	userId, err := getUserId(clientUserId, eventMap)
	if err != nil {
		log.Fatal("UserId not found.")
		return
	}
	eventName, found := eventMap[eventNameKey].(string)
	if !found {
		log.Fatal("Missing EventName in event.")
		return
	}

	eventCreatedTimeString, _ := eventMap[eventCreationTimeKey].(string)
	eventCreatedTimeString = strings.Replace(eventCreatedTimeString, " ", "T", -1)
	eventCreatedTimeString = fmt.Sprintf("%sZ", eventCreatedTimeString)
	eventCreatedTime, _ := time.Parse(time.RFC3339, eventCreatedTimeString)

	eventRequestMap := make(map[string]interface{})
	eventRequestMap["event_name"] = eventName
	eventRequestMap["user_id"] = userId
	eventRequestMap["timestamp"] = eventCreatedTime.Unix()

	// event properties.
	eventPropertiesMap := make(map[string]interface{})
	eventPropertiesMap["amplitude_event_type"], _ = eventMap["amplitude_event_type"]
	eventPropertiesMap["client_event_time"], _ = eventMap["client_event_time"]
	eventPropertiesMap["additional_event_properties"], _ = eventMap["event_properties"]
	//eventPropertiesMap["event_id"], _ = eventMap["event_id"]
	eventPropertiesMap["processed_time"], _ = eventMap["processed_time"]
	//eventPropertiesMap["event_type"], _ = eventMap["event_type"]
	eventPropertiesMap["version_name"], _ = eventMap["version_name"]
	//eventPropertiesMap["session_id"], _ = eventMap["session_id"]
	eventPropertiesMap["$ip"], _ = eventMap["ip_address"]
	eventPropertiesMap["$locationLng"], _ = eventMap["location_lng"]
	eventPropertiesMap["$locationLat"], _ = eventMap["location_lat"]
	eventRequestMap["event_properties"] = eventPropertiesMap

	// User properties associatied with event.
	userPropertiesMap := make(map[string]interface{})
	// Keys that will go into eventProperties.
	userPropertiesMap["$deviceBrand"], _ = eventMap["device_brand"]
	userPropertiesMap["$deviceModel"], _ = eventMap["device_model"]
	userPropertiesMap["$country"], _ = eventMap["country"]
	userPropertiesMap["$os"], _ = eventMap["os_name"]
	userPropertiesMap["$deviceId"], _ = eventMap["device_id"]
	userPropertiesMap["$deviceType"], _ = eventMap["device_type"]
	userPropertiesMap["$language"], _ = eventMap["language"]
	userPropertiesMap["$deviceCarrier"], _ = eventMap["device_carrier"]
	userPropertiesMap["$osVersion"], _ = eventMap["os_version"]
	userPropertiesMap["$city"], _ = eventMap["city"]
	userPropertiesMap["$region"], _ = eventMap["region"]
	userPropertiesMap["$deviceManufacturer"], _ = eventMap["device_manufacturer"]
	userPropertiesMap["$deviceFamily"], _ = eventMap["device_family"]
	userPropertiesMap["$platform"], _ = eventMap["platform"]
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
	file, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		var eventMap map[string]interface{}
		if err := json.Unmarshal([]byte(line), &eventMap); err != nil {
			log.Fatal(fmt.Sprintf("Unable to unmarshal line to json in file %s : %s", filepath, err))
			continue
		}
		lineIngest(eventMap)
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

	fileList := []string{}
	err := filepath.Walk(*inputDirFlag, func(path string, f os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".json") {
			fileList = append(fileList, path)
		}
		return nil
	})
	if err != nil {
		log.Fatal(fmt.Sprintf("Unable to scan directories. Error %v", err))
	}

	for _, file := range fileList {
		fileIngest(file)
	}
}
