package main

// Recursively reads all json files.
// Example usage on Terminal.
// export GOPATH=/Users/aravindmurthy/code/factors/misc/ingest_events
// go run ingest_firebase_ios_events.go --input_file=/Users/aravindmurthy/data/noctacam/sorted_select_all_09_11_2018_1541733812.json --server=http://localhost:8080

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

	log "github.com/sirupsen/logrus"
)

var inputFileFlag = flag.String("input_file", "", "Input json file.")
var serverFlag = flag.String("server", "http://localhost:8080", "Server Path.")
var projectIdFlag = flag.Int("project_id", 0, "Project Id.")
var projectTokenFlag = flag.String("project_token", "", "Needs to be passed if projectId is passed.")

var clientUserIdKey string = "user_pseudo_id"
var clientUserCreationTimeKey string = "user_first_touch_timestamp"
var eventNameKey string = "event_name"
var eventCreationTimeKey string = "event_timestamp"
var clientUserIdToUserIdMap map[string]string = make(map[string]string)

func getUserId(clientUserId string, eventMap map[string]interface{}) (string, error) {
	userId, found := clientUserIdToUserIdMap[clientUserId]
	if !found {
		// Create a user.
		var userCreatedTime int64
		userCreatedTimeData, found := eventMap[clientUserCreationTimeKey]
		if userCreatedTimeData != nil && found {
			userCreatedTime, _ = strconv.ParseInt(userCreatedTimeData.(string), 10, 64)
		} else {
			// If user created time is absent use eventCreatedTime.
			userCreatedTime, _ = strconv.ParseInt(eventMap[eventCreationTimeKey].(string), 10, 64)
		}
		userCreatedTime = userCreatedTime / 1000000

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

	var eventCreatedTime int64
	if eventCreatedTimeData, found := eventMap[eventCreationTimeKey]; eventCreatedTimeData != nil && found {
		eventCreatedTime, _ = strconv.ParseInt(eventCreatedTimeData.(string), 10, 64)
		eventCreatedTime = eventCreatedTime / 1000000
	}

	eventRequestMap := make(map[string]interface{})
	eventRequestMap["event_name"] = eventName
	eventRequestMap["user_id"] = userId
	eventRequestMap["timestamp"] = eventCreatedTime

	// Event properties from params.
	eventPropertiesMap := extractEventProperties(eventMap)
	eventRequestMap["event_properties"] = eventPropertiesMap

	// User properties currently stored along with event.
	userPropertiesMap := extractUserProperties(eventMap)

	// Device properties.
	if dp, ok := eventMap["device"]; ok {
		deviceProperties := extractFloatandStringProperties(dp.(map[string]interface{}))
		for k, value := range deviceProperties {
			switch k {
			case "mobile_brand_name":
				userPropertiesMap["$deviceBrand"] = value
			case "mobile_model_name":
				userPropertiesMap["$deviceModel"] = value
			case "mobile_os_hardware_model":
				userPropertiesMap["$deviceType"] = value
			case "vendor_id":
				userPropertiesMap["$deviceId"] = value
			case "operating_system":
				userPropertiesMap["$os"] = value
			case "operating_system_version":
				userPropertiesMap["$osVersion"] = value
			case "language":
				userPropertiesMap["$language"] = value
			default:
				userPropertiesMap["device."+k] = value
			}
		}
	}
	// Geo properties.
	if gp, ok := eventMap["geo"]; ok {
		geoProperties := extractFloatandStringProperties(gp.(map[string]interface{}))
		for k, value := range geoProperties {
			switch k {
			case "country":
				userPropertiesMap["$country"] = value
			case "region":
				userPropertiesMap["$region"] = value
			case "city":
				userPropertiesMap["$city"] = value
			default:
				userPropertiesMap["geo."+k] = value
			}
		}
	}
	// App properties.
	if ap, ok := eventMap["app_info"]; ok {
		appProperties := extractFloatandStringProperties(ap.(map[string]interface{}))
		for k, value := range appProperties {
			switch k {
			case "version":
				userPropertiesMap["$appVersion"] = value
			default:
				userPropertiesMap["app."+k] = value
			}
		}
	}
	// User LTV properties.
	if ulp, ok := eventMap["user_ltv"]; ok {
		userLTVProperties := extractFloatandStringProperties(ulp.(map[string]interface{}))
		for k, value := range userLTVProperties {
			userPropertiesMap["user.ltv."+k] = value
		}
	}
	// Platform
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

func setupProject() (int, string, error) {
	var projectId int
	var projectToken string

	// Create random project and a corresponding eventName and user.
	reqBody := []byte(`{"name": "NoctaCam"}`)
	url := fmt.Sprintf("%s/projects", *serverFlag)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBody))
	if resp == nil {
		log.Fatal("HTTP Request failed. Is your backend up?")
		return 0, "", errors.New("HTTP Request failed")
	}
	// always close the response-body, even if content is not required.
	defer resp.Body.Close()
	if err != nil {
		log.Fatal(fmt.Sprintf("Http Post user creation failed. Url: %s, reqBody: %s, response: %+v", url, reqBody, resp))
		return 0, "", err
	}
	jsonResponse, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Unable to parse http user create response.")
		return 0, "", err
	}
	var jsonResponseMap map[string]interface{}
	json.Unmarshal(jsonResponse, &jsonResponseMap)
	projectId = int(jsonResponseMap["id"].(float64))
	projectToken = jsonResponseMap["token"].(string)
	return projectId, projectToken, nil
}

func main() {
	flag.Parse()

	if *projectIdFlag <= 0 {
		projectId, projectToken, err := setupProject()
		if err != nil {
			log.Fatal("Project setup failed.")
			os.Exit(1)
		}
		projectIdFlag = &projectId
		projectTokenFlag = &projectToken
	}

	fileIngest(*inputFileFlag)
	log.Info(fmt.Sprintf("Ingested events to project Noctacam with id: %d", *projectIdFlag))
}
