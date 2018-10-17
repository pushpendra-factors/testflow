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

	log "github.com/sirupsen/logrus"
)

var inputFileFlag = flag.String("input_file", "", "Input CSV file.")
var serverFlag = flag.String("server", "http://localhost:8080", "Server Path.")
var projectIdFlag = flag.Int("project_id", 0, "Project Id.")

var clientUserIdToUserIdMap map[string]string = make(map[string]string)

func getUserId(clientUserId string, eventCreatedTime string) (string, error) {
	userId, found := clientUserIdToUserIdMap[clientUserId]
	if !found {
		// Create a user.
		userCreatedTime := eventCreatedTime

		userRequestMap := make(map[string]interface{})
		userRequestMap["c_uid"] = clientUserId
		userRequestMap["created_at"] = userCreatedTime
		userRequestMap["updated_at"] = userCreatedTime
		// No Properties.
		userRequestPropertiesMap := make(map[string]interface{})
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
	eventCreatedTime := fmt.Sprintf("%sT%sZ", eventCreatedTime1, eventCreatedTime2)

	eventType := fields[7]

	userId, err := getUserId(clientUserId, eventCreatedTime)
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
	eventRequestMap["created_at"] = eventCreatedTime
	eventRequestMap["updated_at"] = eventCreatedTime

	eventRequestPropertiesMap := make(map[string]interface{})
	eventRequestPropertiesMap["country"] = country
	eventRequestPropertiesMap["merchant"] = merchant
	eventRequestPropertiesMap["offer_id"] = offerId

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

func setupProject() (int, error) {
	var projectId int

	// Create random project and a corresponding eventName and user.
	reqBody := []byte(`{"name": "ECommerce-Sample"}`)
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

	fileIngest(*inputFileFlag)
	log.Info(fmt.Sprintf("Ingested events to project Ecommerce-Sample with id: %d", *projectIdFlag))
}
