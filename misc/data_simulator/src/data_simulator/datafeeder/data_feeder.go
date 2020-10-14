package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"data_simulator/constants"
	Log "data_simulator/logger"
	"data_simulator/operations"
	"data_simulator/utils"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"cloud.google.com/go/storage"
)

var endpoint *string
var authToken *string
var bulkLoadUrl string
var getUserIdUrl string
var clientUserIdToUserIdMap map[string]string = make(map[string]string)

func ExtractEventData(data string) interface{} {
	split := strings.Split(data, "|")
	var op operations.EventOutput
	if len(split) >= 2 {
		json.Unmarshal([]byte(split[1]), &op)
	} else {
		//ignore: Fix this
	}
	return op
}

func IngestData(obj interface{}) {
	reqBody, _ := json.Marshal(obj)
	url := fmt.Sprintf("%s%s", *endpoint, bulkLoadUrl)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		Log.Error.Fatal(err)
	}
	req.Header.Add(constants.AUTHHEADERNAME, *authToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		//TODO: Janani Handle Retry
		Log.Error.Fatal(err)
	}
	Log.Debug.Printf("%v", resp)
}

func getUserId(clientUserId string, eventTimestamp int64) (string, error) {
	userId, found := clientUserIdToUserIdMap[clientUserId]
	if !found {
		// Create a user.
		userRequestMap := make(map[string]interface{})
		userRequestMap["c_uid"] = clientUserId
		userRequestMap["join_timestamp"] = eventTimestamp

		reqBody, _ := json.Marshal(userRequestMap)
		url := fmt.Sprintf("%s%s", *endpoint, getUserIdUrl)
		client := &http.Client{}
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqBody))
		if err != nil {
			Log.Error.Fatal(err)
		}
		req.Header.Add(constants.AUTHHEADERNAME, *authToken)
		resp, err := client.Do(req)
		if err != nil {
			Log.Error.Fatal(fmt.Sprintf(
				"Http Post user creation failed. Url: %s, reqBody: %s, response: %+v, error: %+v", url, reqBody, resp, err))
			return "", err
		}
		// always close the response-body, even if content is not required
		defer resp.Body.Close()
		jsonResponse, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			Log.Error.Fatal("Unable to parse http user create response.")
			return "", err
		}
		var jsonResponseMap map[string]interface{}
		json.Unmarshal(jsonResponse, &jsonResponseMap)
		userId = jsonResponseMap["user_id"].(string)
		clientUserIdToUserIdMap[clientUserId] = userId
	}
	return userId, nil
}

func main() {
	env := flag.String("env", "development", "")
	seedDate := flag.String("seed_date", "", "")
	endpoint_prod := flag.String("endpoint_prod", "", "")
	authToken_prod := flag.String("projectkey_prod", "", "")
	endpoint_staging := flag.String("endpoint_staging", "", "")
	authToken_staging := flag.String("projectkey_staging", "", "")
	dataConfig := flag.String("config", "", "")
	flag.Parse()
	//configFilePrefix := flag.String("config_filename_prefix", "", "")
	var date time.Time
	if *seedDate == "" {
		date = time.Now()
	} else {
		date, _ = time.Parse(constants.TIMEFORMAT, *seedDate)
	}
	hr, _, _ := date.Clock()
	if hr%2 != 0 {
		fmt.Println("PROD")
		endpoint = endpoint_prod
		authToken = authToken_prod
	} else {
		fmt.Println("STAGING")
		endpoint = endpoint_staging
		authToken = authToken_staging
	}
	filePattern := fmt.Sprintf("%v_%v_%v_%v",
		date.Year(),
		date.Month(),
		date.Day(),
		date.Hour())

	utils.CreateDirectoryIfNotExists(constants.LOCALOUTPUTFOLDER)
	utils.CreateDirectoryIfNotExists("../datagen/processed")
	Log.RegisterLogFiles(filePattern, "datafeeder")
	// TODO : Check how log can be initialized to a file
	//GET these variables from Config
	maxBatchSize := 1000
	counter := 0
	var events []operations.EventOutput

	bulkLoadUrl = "/sdk/event/track/bulk"
	getUserIdUrl = "/sdk/user/identify"
	var files []string
	if *env == "development" {
		files = utils.GetAllFiles(fmt.Sprintf("%s/%s", "../datagen", constants.LOCALOUTPUTFOLDER), *dataConfig)
	} else {
		files = utils.ListAllCloudFiles(fmt.Sprintf("%s/%s", constants.UNPROCESSEDFILESCLOUD, constants.LOCALOUTPUTFOLDER),
			constants.BUCKETNAME, *dataConfig)
	}
	Log.Debug.Printf("Files to be processed %v", files)
	for _, element := range files {
		var scanner *bufio.Scanner
		if strings.HasSuffix(element, ".gz") {
			var gzr *gzip.Reader
			if *env == "development" {
				gzr = utils.GetFileHandlegz(element)
			} else {
				_context := context.Background()
				_storageClient, _err := storage.NewClient(_context)
				if _err != nil {
					Log.Debug.Printf("%v", _err)
				}
				_bucket := _storageClient.Bucket(constants.BUCKETNAME)
				_object := _bucket.Object(element).ReadCompressed(true)
				_reader, _err := _object.NewReader(_context)
				if _err != nil {
					Log.Error.Fatal(_err)
				}
				defer _reader.Close()
				gzr, _err = gzip.NewReader(_reader)
				if _err != nil {
					Log.Error.Fatal(_err)
				}
				defer gzr.Close()
			}
			scanner = bufio.NewScanner(gzr)
		}
		if strings.HasSuffix(element, ".log") {
			if *env == "development" {
				_reader := utils.GetFileHandle(element)
				scanner = bufio.NewScanner(_reader)
			} else {
				_context := context.Background()
				_storageClient, _err := storage.NewClient(_context)

				if _err != nil {
					Log.Debug.Printf("%v", _err)
				}
				_bucket := _storageClient.Bucket(constants.BUCKETNAME)
				_reader, _err := _bucket.Object(element).NewReader(_context)
				defer _reader.Close()
				scanner = bufio.NewScanner(_reader)
			}
		}
		for scanner.Scan() {
			s := scanner.Text()
			op := ExtractEventData(s).(operations.EventOutput)
			if op.UserId != "" {
				op.UserId, _ = getUserId(op.UserId, (int64)(op.Timestamp))
				events = append(events, op)
			}
			counter++

			if counter == maxBatchSize {
				Log.Debug.Printf("Processing %v records", len(events))
				IngestData(events)
				counter = 0
				events = nil
			}
		}
		if counter != 0 {
			Log.Debug.Printf("Processing %v records", len(events))
			IngestData(events)
			counter = 0
			events = nil
		}
		Log.Debug.Printf("Done !!! Processing contents of File: %s", element)
		if *env != "development" {
			utils.MoveFilesInCloud(element, strings.Replace(element, "unprocessed", "processed", 1), constants.BUCKETNAME)
		} else {
			utils.MoveFiles(element, strings.Replace(element, constants.LOCALOUTPUTFOLDER, "processed", 1))
		}
	}
	if *env != "development" {
		utils.CopyFilesToCloud(constants.LOCALOUTPUTFOLDER, constants.UNPROCESSEDFILESCLOUD, constants.BUCKETNAME, true)
	}
}
