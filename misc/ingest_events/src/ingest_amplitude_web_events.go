package main

// Recursively reads all json files sorted in date order.
// Example usage on Terminal.
// export GOPATH=/Users/aravindmurthy/code/factors/misc/ingest_events
// go run ingest_amplitude_web_events.go --input_dir=/Users/aravindmurthy/events --server=http://localhost:8080 --project_id=1 --project_token=

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	U "util"
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

func closeResp(resp *http.Response) {
	_, err := io.Copy(ioutil.Discard, resp.Body)
	resp.Body.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "fetch: reading: %v\n", err)
		os.Exit(1)
	}
}

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
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
		req.Header.Add("Authorization", *projectTokenFlag)
		httpClient := &http.Client{}
		resp, err := httpClient.Do(req)
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
		// always close the response-body, even if content is not required
		closeResp(resp)
		clientUserIdToUserIdMap[clientUserId] = userId
	}
	return userId, nil
}

// Copied from url.go of server code.
func hasProtocol(purl string) bool {
	return len(strings.Split(purl, "://")) > 1
}

func parseURLWithoutProtocol(parseURL string) (*url.URL, error) {
	return url.Parse(fmt.Sprintf("dummy://%s", parseURL))
}

func parseURLStable(parseURL string) (*url.URL, error) {
	if !hasProtocol(parseURL) {
		return parseURLWithoutProtocol(parseURL)
	}
	return url.Parse(parseURL)
}

func getURLHostAndPath(parseURL string) (string, error) {
	cURL := strings.TrimSpace(parseURL)

	if cURL == "" {
		return "", errors.New("parsing failed empty url")
	}

	pURL, err := parseURLStable(cURL)
	if err != nil {
		return "", err
	}

	// adds / as suffix for root.
	path := pURL.Path
	if path == "" {
		path = "/"
	}

	return fmt.Sprintf("%s%s", pURL.Host, path), nil
}

func translateEvent(eventMap map[string]interface{}) map[string]interface{} {
	log.Info(fmt.Sprintf("Processing event: %v", eventMap))
	clientUserId, found := eventMap[clientUserIdKey].(string)
	if !found {
		// If user_id is missing use session_id as user_id.
		clientUserIdFloat, found := eventMap[sessionIdKey].(float64)
		if !found {
			log.Fatal("Missing User Id and session Id in event.")
		}
		clientUserId = fmt.Sprintf("%f", clientUserIdFloat)
	}
	userId, err := getUserId(clientUserId, eventMap)
	if err != nil {
		log.Fatal("UserId not found.")
	}
	eventName, found := eventMap[eventNameKey].(string)
	if !found {
		log.Fatal("Missing EventName in event.")
	}

	eventCreatedTimeString, _ := eventMap[eventCreationTimeKey].(string)
	eventCreatedTimeString = strings.Replace(eventCreatedTimeString, " ", "T", -1)
	eventCreatedTimeString = fmt.Sprintf("%sZ", eventCreatedTimeString)
	eventCreatedTime, _ := time.Parse(time.RFC3339, eventCreatedTimeString)

	eventRequestMap := make(map[string]interface{})
	eventPropertiesMap := make(map[string]interface{})

	if eventName != "Loaded a Page" {
		eventRequestMap["event_name"] = eventName
		// Initialize event properties with, customer sent event properties.
		eventPropertiesMap = eventMap["event_properties"].(map[string]interface{})
	} else {
		// It is an automatic page track. event name is the url.
		pageEP := eventMap["event_properties"].(map[string]interface{})
		urlName, err := getURLHostAndPath(pageEP["url"].(string))
		if err != nil {
			log.Fatal(fmt.Sprintf("event: %s, err: %s", eventMap, err))
		}
		eventRequestMap["event_name"] = urlName
		eventPropertiesMap[U.EP_PAGE_RAW_URL] = pageEP["url"]
		eventPropertiesMap[U.EP_PAGE_TITLE] = pageEP["title"]
		eventPropertiesMap[U.EP_REFERRER] = pageEP["referrer"]
	}

	eventRequestMap["user_id"] = userId
	eventRequestMap["c_event_id"] = eventMap["$insert_id"]
	eventRequestMap["timestamp"] = eventCreatedTime.Unix()

	//eventPropertiesMap["amplitude_event_type"], _ = eventMap["amplitude_event_type"]
	//eventPropertiesMap["client_event_time"], _ = eventMap["client_event_time"]
	//eventPropertiesMap["additional_event_properties"], _ = eventMap["event_properties"]
	//eventPropertiesMap["event_id"], _ = eventMap["event_id"]
	//eventPropertiesMap["processed_time"], _ = eventMap["processed_time"]
	//eventPropertiesMap["event_type"], _ = eventMap["event_type"]
	//eventPropertiesMap["version_name"], _ = eventMap["version_name"]
	//eventPropertiesMap["session_id"], _ = eventMap["session_id"]
	eventPropertiesMap[U.EP_INTERNAL_IP], _ = eventMap["ip_address"]
	eventPropertiesMap[U.EP_LOCATION_LONGITUDE], _ = eventMap["location_lng"]
	eventPropertiesMap[U.EP_LOCATION_LATITUDE], _ = eventMap["location_lat"]
	eventRequestMap["event_properties"] = eventPropertiesMap

	// User properties associatied with event.
	// Initialize it with, customer sent user properties.
	userPropertiesMap := eventMap["user_properties"].(map[string]interface{})
	// Keys that will go into eventProperties.
	userPropertiesMap[U.UP_DEVICE_BRAND], _ = eventMap["device_brand"]
	userPropertiesMap[U.UP_DEVICE_MODEL], _ = eventMap["device_model"]
	userPropertiesMap[U.UP_COUNTRY], _ = eventMap["country"]
	userPropertiesMap[U.UP_OS], _ = eventMap["os_name"]
	userPropertiesMap[U.UP_DEVICE_ID], _ = eventMap["device_id"]
	userPropertiesMap[U.UP_DEVICE_TYPE], _ = eventMap["device_type"]
	userPropertiesMap[U.UP_LANGUAGE], _ = eventMap["language"]
	userPropertiesMap[U.UP_NETWORK_CARRIER], _ = eventMap["device_carrier"]
	userPropertiesMap[U.UP_OS_VERSION], _ = eventMap["os_version"]
	userPropertiesMap[U.UP_CITY], _ = eventMap["city"]
	userPropertiesMap[U.UP_REGION], _ = eventMap["region"]
	userPropertiesMap[U.UP_DEVICE_MANUFACTURER], _ = eventMap["device_manufacturer"]
	userPropertiesMap[U.UP_DEVICE_FAMILY], _ = eventMap["device_family"]
	userPropertiesMap[U.UP_PLATFORM], _ = eventMap["platform"]
	eventRequestMap["user_properties"] = userPropertiesMap

	return eventRequestMap
}

func getFileEvents(filepath string) []map[string]interface{} {
	file, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	fileEvents := []map[string]interface{}{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		var eventMap map[string]interface{}
		if err := json.Unmarshal([]byte(line), &eventMap); err != nil {
			log.Fatal(fmt.Sprintf("Unable to unmarshal line to json in file %s : %s", filepath, err))
			continue
		}
		fileEvents = append(fileEvents, eventMap)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return fileEvents
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
	closeResp(resp)
	return nil
}

func batchAndIngestEvents(eventsData []map[string]interface{}) {
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
	sort.SliceStable(fileList,
		func(i, j int) bool {
			// File names are of format Downloads/162692/162692_2019-03-14_6#754.json
			// Hour values are not in alphabetical order.
			fnameISlashSplits := strings.Split(fileList[i], "/")
			fnameI := fnameISlashSplits[len(fnameISlashSplits)-1]
			fnameIUnderscoreSplits := strings.Split(fnameI, "_")
			fnameIDate := fnameIUnderscoreSplits[1]
			fnameIHour, _ := strconv.ParseUint(strings.Split(fnameIUnderscoreSplits[2], "#")[0], 10, 32)

			fnameJSlashSplits := strings.Split(fileList[j], "/")
			fnameJ := fnameJSlashSplits[len(fnameJSlashSplits)-1]
			fnameJUnderscoreSplits := strings.Split(fnameJ, "_")
			fnameJDate := fnameJUnderscoreSplits[1]
			fnameJHour, _ := strconv.ParseUint(strings.Split(fnameJUnderscoreSplits[2], "#")[0], 10, 32)

			if fnameIDate != fnameJDate {
				return fnameIDate < fnameJDate
			} else {
				return fnameIHour < fnameJHour
			}
		})

	eventsData := []map[string]interface{}{}
	for _, file := range fileList {
		log.Printf(fmt.Sprintf("Reading file: %s", file))
		eventsData = append(eventsData, getFileEvents(file)...)
	}
	batchAndIngestEvents(eventsData)
}
