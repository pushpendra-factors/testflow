package main

// Recursively reads all json files.
// Example usage on Terminal.
// export GOPATH=/Users/aravindmurthy/code/factors/misc/ingest_noctacam_events
// go run ingest_noctacam_events.go --input_file=/Users/aravindmurthy/data/noctacam/sorted_select_all_09_11_2018_1541733812.json --server=http://localhost:8080

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

var inputFileFlag = flag.String("input_file", "", "Input json file.")
var serverFlag = flag.String("server", "http://localhost:8080", "Server Path.")
var projectIdFlag = flag.Int("project_id", 0, "Project Id.")

var clientUserIdKey string = "user_pseudo_id"
var clientUserCreationTimeKey string = "user_first_touch_timestamp"
var eventNameKey string = "event_name"
var eventCreationTimeKey string = "event_timestamp"
var clientUserIdToUserIdMap map[string]string = make(map[string]string)

func getUserId(clientUserId string, eventMap map[string]interface{}) (string, error) {
	userId, found := clientUserIdToUserIdMap[clientUserId]
	if !found {
		// Create a user.
		var userCreatedTimeInt int64
		userCreatedTimeData, found := eventMap[clientUserCreationTimeKey]
		if userCreatedTimeData != nil && found {
			userCreatedTimeInt, _ = strconv.ParseInt(userCreatedTimeData.(string), 10, 64)
		} else {
			// If user created time is absent use eventCreatedTime.
			userCreatedTimeInt, _ = strconv.ParseInt(eventMap[eventCreationTimeKey].(string), 10, 64)
		}
		userCreatedTimeInt = userCreatedTimeInt / 1000000
		userCreatedTime := time.Unix(int64(userCreatedTimeInt), 0).Format(time.RFC3339)

		userRequestMap := make(map[string]interface{})
		userRequestMap["c_uid"] = clientUserId
		userRequestMap["created_at"] = userCreatedTime
		userRequestMap["updated_at"] = userCreatedTime
		userRequestMap["properties"] = extractUserProperties(eventMap)

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

func extractEventProperties(eventMap map[string]interface{}) map[string]interface{} {
	eventPropertiesMap := make(map[string]interface{})
	up, ok := eventMap["event_params"]
	if !ok {
		return eventPropertiesMap
	}
	eventPropertiesData := up.([]interface{})
	addKeyValuesFromPropertiesArray(eventPropertiesData, eventPropertiesMap)
	return eventPropertiesMap
}

func extractUserProperties(eventMap map[string]interface{}) map[string]interface{} {
	userPropertiesMap := make(map[string]interface{})
	up, ok := eventMap["user_properties"]
	if !ok {
		return userPropertiesMap
	}
	userPropertiesData := up.([]interface{})
	addKeyValuesFromPropertiesArray(userPropertiesData, userPropertiesMap)
	return userPropertiesMap
}

func addKeyValuesFromPropertiesArray(
	propertiesData []interface{}, propertiesMap map[string]interface{}) {
	for _, pm := range propertiesData {
		propertyMap := pm.(map[string]interface{})
		key := ""
		for k, vs := range propertyMap {
			if k == "key" {
				key = vs.(string)
			}
		}
		if key == "" {
			continue
		}
		for k, vs := range propertyMap {
			if k == "value" {
				values := vs.(map[string]interface{})
				if val, ok := values["int_value"]; ok {
					if intVal, err := strconv.ParseInt(val.(string), 10, 64); err == nil {
						propertiesMap[key] = intVal
					}
				} else if val, ok := values["double_value"]; ok {
					if floatVal, ok := val.(float64); ok {
						propertiesMap[key] = floatVal
					} else if floatVal, err := strconv.ParseFloat(val.(string), 64); err == nil {
						propertiesMap[key] = floatVal
					}
				} else if val, ok := values["string_value"]; ok {
					// Try parsing int. Most of the counts are stored as string_value.
					if intVal, err := strconv.ParseInt(val.(string), 10, 64); err == nil {
						propertiesMap[key] = intVal
					} else {
						propertiesMap[key] = val.(string)
					}
				}
			}
		}
	}
}

func extractFloatandStringProperties(
	inputPropertiesMap map[string]interface{}) map[string]interface{} {
	propertiesMap := make(map[string]interface{})
	for key, v := range inputPropertiesMap {
		if floatVal, ok := v.(float64); ok {
			propertiesMap[key] = floatVal
		} else {
			val := v.(string)
			// Try parsing int, float and string in that order.
			if intVal, err := strconv.ParseInt(val, 10, 64); err == nil {
				propertiesMap[key] = intVal
			} else if floatVal, err := strconv.ParseFloat(val, 64); err == nil {
				propertiesMap[key] = floatVal
			} else {
				propertiesMap[key] = val
			}
		}
	}
	return propertiesMap
}

func lineIngest(eventMap map[string]interface{}) {
	log.Info(fmt.Sprintf("Processing event: %v", eventMap))
	clientUserId, found := eventMap[clientUserIdKey].(string)
	if !found {
		log.Fatal("Missing User Id and session Id in event.")
		return
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

	var eventCreatedTime string
	if eventCreatedTimeData, found := eventMap[eventCreationTimeKey]; eventCreatedTimeData != nil && found {
		eventCreatedTimeInt, _ := strconv.ParseInt(eventCreatedTimeData.(string), 10, 64)
		eventCreatedTimeInt = eventCreatedTimeInt / 1000000
		eventCreatedTime = time.Unix(int64(eventCreatedTimeInt), 0).Format(time.RFC3339)
	}

	eventRequestMap := make(map[string]interface{})
	eventRequestMap["event_name"] = eventName
	eventRequestMap["created_at"] = eventCreatedTime
	eventRequestMap["updated_at"] = eventCreatedTime

	// Event properties from params.
	eventRequestPropertiesMap := extractEventProperties(eventMap)
	// User properties currently stored along with event.
	userProperties := extractUserProperties(eventMap)
	for k, value := range userProperties {
		eventRequestPropertiesMap["user." + k] = value
	}
	// Device properties.
	if dp, ok := eventMap["device"]; ok {
		deviceProperties := extractFloatandStringProperties(dp.(map[string]interface{}))
		for k, value := range deviceProperties {
			eventRequestPropertiesMap["device." + k] = value
		}
	}
	// Geo properties.
	if gp, ok := eventMap["geo"]; ok {
		geoProperties := extractFloatandStringProperties(gp.(map[string]interface{}))
		for k, value := range geoProperties {
			eventRequestPropertiesMap["geo." + k] = value
		}
	}
	// App properties.
	if ap, ok := eventMap["app_info"]; ok {
		appProperties := extractFloatandStringProperties(ap.(map[string]interface{}))
		for k, value := range appProperties {
			eventRequestPropertiesMap["app." + k] = value
		}
	}
	// User LTV properties.
	if ulp, ok := eventMap["user_ltv"]; ok {
		userLTVProperties := extractFloatandStringProperties(ulp.(map[string]interface{}))
		for k, value := range userLTVProperties {
			eventRequestPropertiesMap["user.ltv." + k] = value
		}
	}
	// Platform
	eventRequestPropertiesMap["platform"], _ = eventMap["platform"]
	
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
	reqBody := []byte(`{"name": "NoctaCam"}`)
	url := fmt.Sprintf("%s/projects", *serverFlag)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBody))
	if resp == nil {
		log.Fatal("HTTP Request failed. Is your backend up?")
		return 0, errors.New("HTTP Request failed")
	}
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
	log.Info(fmt.Sprintf("Ingested events to project Noctacam with id: %d", *projectIdFlag))
}
