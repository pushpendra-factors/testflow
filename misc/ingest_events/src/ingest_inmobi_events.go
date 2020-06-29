package main

// Recursively reads all json files.
// Example usage on Terminal.
// export PATH_TO_FACTORS=~/repos
// cd $PATH_TO_FACTORS/factors/misc/ingest_events/src
// export GOPATH=$PATH_TO_FACTORS/factors/misc/ingest_events
// go run ingest_inmobi_events.go --input_file=/Users/aravindmurthy/Downloads/inmobi/new/first_million.csv --server=http://factors-dev.com:8080  --server=http://localhost:8080 --project_id=193 --project_token=1y9tqdb5ps6opnhu7fsetsiebtf5k6et

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	U "util"
)

var inputFileFlag = flag.String("input_file", "", "Input csv file.")
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

func translateEvent(event string) (map[string]interface{}, error) {
	/*
		Each line is a csv with following fields in order.
		Example:
		-1427909899|-39408865;1549862325000;adid;com.foxnews.android;a_present;2600:1700:3360:6a70:e8d4:eb1f:7f14:859c;;US;AT&T;SM-G955U;c39c3435dd9c686b1593f662ce9562918e7d5035;android;29.940453;-90.07273;2019-02-12 15:19:15;2600:387:1:809:0:0:0:94;;;AT&T;149.6;GPS;;;Operation Hope;1680691943949993555;;USA;1;SM-G955U;Reveal;cd33c9a853c93382a7a341b7e9284594;[94];[15118];[15250];[5592168, 247641];[1, 370, 20, 311, 24, 123];[]
		1) eventid             	string
		2) eventtimestamp      	string
		3) idtype              	string
		4) consumedresourceuri 	string
		5) consumptiontype     	string
		6) ipaddress           	string
		7) useragent           	string
		8) countrycode         	string
		9) network_onesignal   	string
		10) devicealias         	string
		11) userid              	string
		12) os                  	string
		13) lat                 	double
		14) lon                 	double
		15) req_ts              	string
		16) ip_address          	string
		17) wifi_bssid          	string
		18) connection_type     	string
		19) carrier             	string
		20) accuracy            	string
		21) event_type          	string
		22) gps_speed           	string
		23) gps_altitude        	string
		24) place_name          	string
		25) place_id            	string
		26) category            	string
		27) country             	string
		28) background          	string
		29) model               	string
		30) partner             	string
		31) app_id              	string
		32) countrylist         	string
		33) statelist           	string
		34) citylist            	string
		35) ziplist             	string
		36) poicategorylist     	string
		37) polygonidlist       	string
	*/
	eventRequestMap := make(map[string]interface{})

	log.Info(fmt.Sprintf("Processing line: %s", event))
	eventFields := strings.Split(event, ";")
	if len(eventFields) != 37 {
		return eventRequestMap, fmt.Errorf("Unexpected number of fields in event.")
	}
	clientUserId := eventFields[10]

	eventTimestamp, err := strconv.ParseInt(eventFields[1], 10, 64)
	if err != nil {
		log.Fatal(fmt.Sprintf("Unable to parse timestamp, %s, %v",
			eventFields[1], err))
	}
	eventTimestamp = eventTimestamp / 1000 // ms to s.

	userId, err := getUserId(clientUserId, eventTimestamp)
	if err != nil {
		log.Fatal("UserId not found.")
	}

	eventRequestMap["event_name"] = eventFields[3] // consumedresourceuri
	eventRequestMap["user_id"] = userId
	eventRequestMap["timestamp"] = eventTimestamp

	// Event and user properties.
	eventPropertiesMap := make(map[string]interface{})
	eventRequestMap["c_event_id"] = eventFields[0]
	eventPropertiesMap["consumptionType"] = eventFields[4]
	eventPropertiesMap["ipaddress1"] = eventFields[5]
	eventPropertiesMap[U.EP_LOCATION_LATITUDE] = eventFields[12]
	eventPropertiesMap[U.EP_LOCATION_LONGITUDE] = eventFields[13]
	eventPropertiesMap["ipaddress2"] = eventFields[15]
	eventPropertiesMap["category"] = eventFields[25]
	eventPropertiesMap["background"] = eventFields[27]
	eventPropertiesMap["appId"] = eventFields[30]
	// Convert to array. Use only first element since arrays are not supported.
	countryList := []int{}
	if err = json.Unmarshal([]byte(eventFields[31]), &countryList); err == nil {
		if len(countryList) > 0 {
			eventPropertiesMap["countryList"] = countryList[0]
		}
	}
	stateList := []int{}
	if err = json.Unmarshal([]byte(eventFields[32]), &stateList); err == nil {
		if len(stateList) > 0 {
			eventPropertiesMap["stateList"] = stateList[0]
		}
	}
	cityList := []int{}
	if err = json.Unmarshal([]byte(eventFields[33]), &cityList); err == nil {
		if len(cityList) > 0 {
			eventPropertiesMap["cityList"] = cityList[0]
		}
	}
	zipList := []int{}
	if err = json.Unmarshal([]byte(eventFields[34]), &zipList); err == nil {
		if len(zipList) > 0 {
			eventPropertiesMap["zipList"] = zipList[0]
		}
	}

	userPropertiesMap := make(map[string]interface{})
	userPropertiesMap[U.UP_BROWSER] = eventFields[6]
	userPropertiesMap["country1"] = eventFields[7]
	userPropertiesMap["network"] = eventFields[8]
	userPropertiesMap[U.UP_DEVICE_TYPE] = eventFields[9]
	userPropertiesMap[U.UP_OS] = eventFields[11]
	userPropertiesMap[U.UP_DEVICE_CARRIER] = eventFields[18]
	userPropertiesMap["placeName"] = eventFields[23]
	userPropertiesMap["placeId"] = eventFields[24]
	userPropertiesMap["country2"] = eventFields[26]
	userPropertiesMap[U.UP_DEVICE_MODEL] = eventFields[28]
	userPropertiesMap["partner"] = eventFields[29]

	eventRequestMap["event_properties"] = eventPropertiesMap
	eventRequestMap["user_properties"] = userPropertiesMap

	return eventRequestMap, nil
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

func getFileEvents(scanner *bufio.Scanner, batchSize int) ([]string, bool) {
	lineNumber := 1
	fileEvents := []string{}
	for scanner.Scan() && lineNumber <= batchSize {
		line := scanner.Text()
		fileEvents = append(fileEvents, line)
		lineNumber++
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return fileEvents, lineNumber < batchSize
}

func fileIngest(filepath string) {
	file, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)

	maxBatchSize := 1000
	batchNumber := 1
	eof := false
	for !eof {
		log.WithFields(log.Fields{
			"start": (batchNumber - 1) * maxBatchSize,
			"end":   batchNumber * maxBatchSize,
			"batch": batchNumber,
		}).Info("Ingesting Batch")

		eventLines, tmpEof := getFileEvents(scanner, maxBatchSize)
		translatedEvents := make([]map[string]interface{}, 0, 0)
		for _, event := range eventLines {
			if trEvent, err := translateEvent(event); err == nil {
				translatedEvents = append(translatedEvents, trEvent)
			}
		}

		err := ingestEvents(translatedEvents)
		if err != nil {
			log.WithError(err).Fatal("Error ingesting events")
		}

		eof = tmpEof
		batchNumber++
	}
}

func main() {
	flag.Parse()

	if *projectIdFlag <= 0 || *projectTokenFlag == "" {
		log.Printf("Missing args ProjectId: %d, projectToken: %s", *projectIdFlag, *projectTokenFlag)
		os.Exit(1)
	}

	fileIngest(*inputFileFlag)
	log.Info(fmt.Sprintf("Ingested events to project with id: %d", *projectIdFlag))
}
