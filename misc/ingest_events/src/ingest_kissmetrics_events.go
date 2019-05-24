package main

// Recursively reads all json files sorted in date order.
// Example usage on Terminal.
// export GOPATH=/Users/aravindmurthy/code/factors/misc/ingest_events
// go run ingest_kissmetrics_events.go --input_dir=/Users/aravindmurthy/Downloads/Last_3_Months_Data --server=http://localhost:8080 --project_id=1 --project_token=

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

	log "github.com/sirupsen/logrus"
)

var inputDirFlag = flag.String("input_dir", "", "Input Directory with json files.")
var serverFlag = flag.String("server", "http://factors-dev.com:8080", "Server Path.")
var projectIdFlag = flag.Int("project_id", 0, "Project Id.")
var projectTokenFlag = flag.String("project_token", "", "Needs to be passed if projectId is passed.")

var clientUserIdKey string = "_p"
var clientUserIdAliasKey string = "_p2"
var eventNameKey string = "_n"
var eventCreationTimeKey string = "_t"

var clientUserIdToUserIdMap map[string]string = make(map[string]string)

func closeResp(resp *http.Response) {
	_, err := io.Copy(ioutil.Discard, resp.Body)
	resp.Body.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "fetch: reading: %v\n", err)
		os.Exit(1)
	}
}

func getUserId(clientUserId string, clientUserIdAlias string, currentEventTimestamp int) (string, error) {
	found := false
	userId := ""
	userId, found = clientUserIdToUserIdMap[clientUserId]

	if !found && clientUserIdAlias != "" {
		// try with alias.
		userId, found = clientUserIdToUserIdMap[clientUserIdAlias]
	}

	if !found {
		userRequestMap := make(map[string]interface{})
		userRequestMap["c_uid"] = clientUserId
		userRequestMap["join_timestamp"] = currentEventTimestamp

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
		if clientUserIdAlias != "" {
			clientUserIdToUserIdMap[clientUserIdAlias] = userId
		}
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
	clientUserId, _ := eventMap[clientUserIdKey].(string)
	clientUserIdAlias, _ := eventMap[clientUserIdAliasKey].(string)
	if clientUserId == "" && clientUserIdAlias == "" {
		log.Fatal(fmt.Sprintf("Missing User Id in event: %s", eventMap))
	}

	if clientUserId == "" && clientUserIdAlias != "" {
		clientUserId = clientUserIdAlias
		clientUserIdAlias = ""
	}

	eventCreatedTimestamp := int(eventMap[eventCreationTimeKey].(float64))

	userId, err := getUserId(clientUserId, clientUserIdAlias, eventCreatedTimestamp)
	if err != nil {
		log.Fatal("UserId not found.")
	}

	eventName, found := eventMap[eventNameKey].(string)
	if !found {
		log.Info(fmt.Sprintf("Missing EventName in event: %s", eventMap))
		return nil
	}

	eventRequestMap := make(map[string]interface{})
	eventPropertiesMap := make(map[string]interface{})

	eventRequestMap["event_name"] = eventName
	eventRequestMap["user_id"] = userId
	eventRequestMap["c_event_id"] = fmt.Sprintf("%s-%s-%d", clientUserId, eventName, eventCreatedTimestamp)
	eventRequestMap["timestamp"] = eventCreatedTimestamp

	// All other key values go to event properties by default.
	for k, v := range eventMap {
		if k != clientUserIdKey && k != clientUserIdAliasKey &&
			k != eventCreationTimeKey && k != eventNameKey {
			eventPropertiesMap[k] = v
		}
	}
	eventRequestMap["event_properties"] = eventPropertiesMap

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
	maxBatchSize := 10000
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
			te := translateEvent(event)
			if te != nil {
				translatedEvents = append(translatedEvents, te)
			}
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
			// File names are of format 934.json 1001.json etc.
			fnameISlashSplits := strings.Split(fileList[i], "/")
			fnameI := fnameISlashSplits[len(fnameISlashSplits)-1]
			fnameINumber, _ := strconv.ParseUint(strings.Split(fnameI, ".")[0], 10, 32)

			fnameJSlashSplits := strings.Split(fileList[j], "/")
			fnameJ := fnameJSlashSplits[len(fnameJSlashSplits)-1]
			fnameJNumber, _ := strconv.ParseUint(strings.Split(fnameJ, ".")[0], 10, 32)
			return fnameINumber < fnameJNumber
		})

	eventsData := []map[string]interface{}{}
	for _, file := range fileList {
		log.Printf(fmt.Sprintf("Reading file: %s", file))
		eventsData = append(eventsData, getFileEvents(file)...)
	}
	batchAndIngestEvents(eventsData)
}
