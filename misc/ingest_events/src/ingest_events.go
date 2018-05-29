package main

// Recursively reads all json files.
// Example usage on Terminal.
// export GOPATH=/Users/aravindmurthy/code/autometa/misc/ingest_events
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

	log "github.com/sirupsen/logrus"
)

var inputDirFlag = flag.String("input_dir", "", "Input Directory with json files.")
var serverFlag = flag.String("server", "http://localhost:8080", "Server Path.")
var projectIdFlag = flag.Int("project_id", 0, "Project Id.")

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
		userCreatedTime, _ := eventMap[clientUserCreationTimeKey].(string)
		userCreatedTime = strings.Replace(userCreatedTime, " ", "T", -1)
		userCreatedTime = fmt.Sprintf("%sZ", userCreatedTime)

		userRequestMap := make(map[string]interface{})
		userRequestMap["c_uid"] = clientUserId
		userRequestMap["created_at"] = userCreatedTime
		userRequestMap["updated_at"] = userCreatedTime

		userRequestPropertiesMap := make(map[string]interface{})
		// TODO: These properties are not updated. Currently only based on the first event.
		userRequestPropertiesMap["device_brand"], _ = eventMap["device_brand"]
		userRequestPropertiesMap["location_lng"], _ = eventMap["location_lng"]
		userRequestPropertiesMap["ip_address"], _ = eventMap["ip_address"]
		userRequestPropertiesMap["device_model"], _ = eventMap["device_model"]
		userRequestPropertiesMap["language"], _ = eventMap["language"]
		userRequestPropertiesMap["country"], _ = eventMap["country"]
		userRequestPropertiesMap["os_name"], _ = eventMap["os_name"]
		userRequestPropertiesMap["device_id"], _ = eventMap["device_id"]
		userRequestPropertiesMap["device_carrier"], _ = eventMap["device_carrier"]
		userRequestPropertiesMap["os_version"], _ = eventMap["os_version"]
		userRequestPropertiesMap["city"], _ = eventMap["city"]
		userRequestPropertiesMap["device_type"], _ = eventMap["device_type"]
		userRequestPropertiesMap["version_name"], _ = eventMap["version_name"]
		userRequestPropertiesMap["region"], _ = eventMap["region"]
		userRequestPropertiesMap["device_manufacturer"], _ = eventMap["device_manufacturer"]
		userRequestPropertiesMap["location_lat"], _ = eventMap["location_lat"]
		userRequestPropertiesMap["device_family"], _ = eventMap["device_family"]
		userRequestPropertiesMap["platform"], _ = eventMap["platform"]
		userRequestPropertiesMap["library"], _ = eventMap["library"]
		userRequestPropertiesMap["user_properties"], _ = eventMap["user_properties"]
		userRequestMap["properties"] = userRequestPropertiesMap

		reqBody, _ := json.Marshal(userRequestMap)
		url := fmt.Sprintf("%s/projects/%d/users", *serverFlag, *projectIdFlag)
		resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBody))
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
		userId = jsonResponseMap["id"].(string)
		clientUserIdToUserIdMap[clientUserId] = userId
	}
	return userId, nil
}

func lineIngest(eventMap map[string]interface{}) {
	log.Info("Processing event: %v", eventMap)
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

	eventCreatedTime, _ := eventMap[eventCreationTimeKey].(string)
	eventCreatedTime = strings.Replace(eventCreatedTime, " ", "T", -1)
	eventCreatedTime = fmt.Sprintf("%sZ", eventCreatedTime)

	eventRequestMap := make(map[string]interface{})
	eventRequestMap["event_name"] = eventName
	eventRequestMap["created_at"] = eventCreatedTime
	eventRequestMap["updated_at"] = eventCreatedTime

	eventRequestPropertiesMap := make(map[string]interface{})
	// Keys that will go into eventProperties.
	eventRequestPropertiesMap["device_brand"], _ = eventMap["device_brand"]
	eventRequestPropertiesMap["location_lng"], _ = eventMap["location_lng"]
	eventRequestPropertiesMap["amplitude_event_type"], _ = eventMap["amplitude_event_type"]
	eventRequestPropertiesMap["ip_address"], _ = eventMap["ip_address"]
	eventRequestPropertiesMap["client_event_time"], _ = eventMap["client_event_time"]
	eventRequestPropertiesMap["device_model"], _ = eventMap["device_model"]
	eventRequestPropertiesMap["event_properties"], _ = eventMap["event_properties"]
	eventRequestPropertiesMap["language"], _ = eventMap["language"]
	eventRequestPropertiesMap["country"], _ = eventMap["country"]
	eventRequestPropertiesMap["event_id"], _ = eventMap["event_id"]
	eventRequestPropertiesMap["os_name"], _ = eventMap["os_name"]
	eventRequestPropertiesMap["device_id"], _ = eventMap["device_id"]
	eventRequestPropertiesMap["processed_time"], _ = eventMap["processed_time"]
	eventRequestPropertiesMap["event_type"], _ = eventMap["event_type"]
	eventRequestPropertiesMap["device_carrier"], _ = eventMap["device_carrier"]
	eventRequestPropertiesMap["os_version"], _ = eventMap["os_version"]
	eventRequestPropertiesMap["city"], _ = eventMap["city"]
	eventRequestPropertiesMap["device_type"], _ = eventMap["device_type"]
	eventRequestPropertiesMap["version_name"], _ = eventMap["version_name"]
	eventRequestPropertiesMap["region"], _ = eventMap["region"]
	eventRequestPropertiesMap["device_manufacturer"], _ = eventMap["device_manufacturer"]
	eventRequestPropertiesMap["location_lat"], _ = eventMap["location_lat"]
	eventRequestPropertiesMap["device_family"], _ = eventMap["device_family"]
	eventRequestPropertiesMap["platform"], _ = eventMap["platform"]
	eventRequestPropertiesMap["library"], _ = eventMap["library"]
	eventRequestPropertiesMap["session_id"], _ = eventMap["session_id"]
	eventRequestPropertiesMap["event_properties"], _ = eventMap["event_properties"]
	eventRequestMap["properties"] = eventRequestPropertiesMap

	reqBody, _ := json.Marshal(eventRequestMap)
	url := fmt.Sprintf("%s/projects/%d/users/%s/events", *serverFlag, *projectIdFlag, userId)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBody))
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

func setupProject() (int, error) {
	var projectId int

	// Create random project and a corresponding eventName and user.
	reqBody := []byte(`{"name": "ingest_events_project"}`)
	url := fmt.Sprintf("%s/projects", *serverFlag)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBody))
	// always close the response-body, even if content is not required.
	defer resp.Body.Close()
	if err != nil {
		log.Fatal(fmt.Sprintf("Http Post user creation failed. Url: %s, reqBody: %s, response: %+v", url, reqBody, resp))
		return 0, err
	}
	jsonResponse, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Unable to parse http user create response.")
		return 0, err
	}
	var jsonResponseMap map[string]interface{}
	json.Unmarshal(jsonResponse, &jsonResponseMap)
	projectId = int(jsonResponseMap["id"].(float64))
	return projectId, nil
}

func main() {
	flag.Parse()

	if *projectIdFlag <= 0 {
		projectId, err := setupProject()
		if err != nil {
			log.Fatal("Project setup failed.")
			os.Exit(1)
		}
		projectIdFlag = &projectId
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
