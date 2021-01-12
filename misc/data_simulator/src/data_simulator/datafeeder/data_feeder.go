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
	"net/url"
	"strings"
	"time"

	"cloud.google.com/go/storage"
)

const EventPropertyCampaign = "$campaign"

var clientUserIdToUserIdMap_staging = make(map[string]string)
var clientUserIdToUserIdMap_prod = make(map[string]string)
var campaignCounterPROD = make(map[string]int)
var campaignCounterSTAGE = make(map[string]int)

var bulkLoadUrl = "/sdk/event/track/bulk"
var getUserIdUrl = "/sdk/user/identify"
var adwordsAddDocUrl = "/data_service/adwords/documents/add"
var projectIDStage = uint64(0)
var adwordsCustomerAccountIDStage = ""
var projectIDProd = uint64(0)
var adwordsCustomerAccountIDProd = ""

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

func IngestAdwordsData(yesterday time.Time, projectId uint64, customerAccountId string, campaignCounter map[string]int, endpoint *string, authToken *string) {
	for campaignName, eventCount := range campaignCounter {

		adwordsDoc := operations.GetCampaignPerfReportAdwordsDoc(yesterday, projectId,
			customerAccountId, campaignName, eventCount)

		reqBody, _ := json.Marshal(adwordsDoc)
		url := fmt.Sprintf("%s%s", *endpoint, adwordsAddDocUrl)
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
		Log.Debug.Printf("Added adwords doc %v", resp)
	}
}

func IngestData(obj interface{}, endpoint *string, authToken *string) {
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

func getUserId(clientUserId string, eventTimestamp int64, endpoint *string, authToken *string, env string) (string, error) {
	var userId string
	var found bool
	if env == "staging" {
		userId, found = clientUserIdToUserIdMap_staging[clientUserId]
	}
	if env == "prod" {
		userId, found = clientUserIdToUserIdMap_prod[clientUserId]
	}
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
		if env == "staging" {
			clientUserIdToUserIdMap_staging[clientUserId] = userId
		}
		if env == "prod" {
			clientUserIdToUserIdMap_prod[clientUserId] = userId
		}
	}
	return userId, nil
}

func main() {

	env := flag.String("env", "development", "")
	seedDate := flag.String("seed_date", "", "")
	endpoint_prod := flag.String("endpoint_prod", "http://factors-dev.com:8085", "")
	authToken_prod := flag.String("projectkey_prod", "", "")
	project_id_prod := flag.Int("project_id_prod", 0, "Prod project Id")
	adwords_customer_id_prod := flag.String("adwords_customer_id_prod", "", "")

	endpoint_staging := flag.String("endpoint_staging", "http://factors-dev.com:8085", "")
	authToken_staging := flag.String("projectkey_staging", "", "")
	project_id_stage := flag.Int("project_id_stage", 0, "Stage project Id")
	adwords_customer_id_stage := flag.String("adwords_customer_id_stage", "", "")

	bulk_load_url := flag.String("bulk_load_url", "/sdk/event/track/bulk", "")
	get_user_id_url := flag.String("get_user_id_url", "/sdk/user/identify", "")
	adwords_add_doc_url := flag.String("adwords_add_doc_url", "/data_service/adwords/documents/add", "")
	bulkLoadUrl = *(bulk_load_url)
	getUserIdUrl = *(get_user_id_url)
	adwordsAddDocUrl = *(adwords_add_doc_url)

	dataConfig := flag.String("config", "", "")
	flag.Parse()
	//configFilePrefix := flag.String("config_filename_prefix", "", "")
	var executionDate time.Time
	if *seedDate == "" {
		executionDate = time.Now()
	} else {
		executionDate, _ = time.Parse(constants.TIMEFORMAT, *seedDate)
	}

	if *project_id_stage == 0 || *project_id_prod == 0 {
		Log.Debug.Printf("ProjectID is not provided")
		return
	}
	if *project_id_stage == 0 || *project_id_prod == 0 {
		Log.Debug.Printf("ProjectID is not provided")
		return
	}
	if *adwords_customer_id_stage == "" || *adwords_customer_id_prod == "" {
		Log.Debug.Printf("Adwords customer account is not provided")

	}
	projectIDStage = uint64(*project_id_stage)
	projectIDProd = uint64(*project_id_prod)
	adwordsCustomerAccountIDStage = *adwords_customer_id_stage
	adwordsCustomerAccountIDProd = *adwords_customer_id_prod

	filePattern := fmt.Sprintf("%v_%v_%v_%v",
		executionDate.Year(),
		executionDate.Month(),
		executionDate.Day(),
		executionDate.Hour())

	utils.CreateDirectoryIfNotExists(constants.LOCALOUTPUTFOLDER)
	utils.CreateDirectoryIfNotExists("../datagen/processed")
	Log.RegisterLogFiles(filePattern, "datafeeder")
	// TODO : Check how log can be initialized to a file
	// GET these variables from Config

	var events_staging []operations.EventOutput
	var events_prod []operations.EventOutput

	// processing events data
	// ProcessEventsFiles(env, dataConfig, endpoint_staging, authToken_staging, events_staging, endpoint_prod, authToken_prod, events_prod)

	// processing adwords data
	ProcessAdwordsDataFromEventsFiles(executionDate, env, dataConfig, endpoint_staging, authToken_staging, events_staging, endpoint_prod, authToken_prod, events_prod)
}

func ProcessEventsFiles(env *string, dataConfig *string, endpoint_staging *string,
	authToken_staging *string, events_staging []operations.EventOutput, endpoint_prod *string,
	authToken_prod *string, events_prod []operations.EventOutput) {

	maxBatchSize := 1000
	counter := 0
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
			op = trimUrlParams(op)
			if op.UserId != "" {
				op.UserId, _ = getUserId(op.UserId, (int64)(op.Timestamp), endpoint_staging, authToken_staging, "staging")
				events_staging = append(events_staging, op)
				op.UserId, _ = getUserId(op.UserId, (int64)(op.Timestamp), endpoint_prod, authToken_prod, "prod")
				events_prod = append(events_prod, op)
			}
			counter++

			if counter == maxBatchSize {
				Log.Debug.Printf("Processing %v records", len(events_staging))
				IngestData(events_staging, endpoint_staging, authToken_staging)
				counter = 0
				events_staging = nil
				Log.Debug.Printf("Processing %v records", len(events_prod))
				IngestData(events_prod, endpoint_prod, authToken_prod)
				counter = 0
				events_prod = nil
			}
		}
		if counter != 0 {
			Log.Debug.Printf("Processing %v records", len(events_staging))
			IngestData(events_staging, endpoint_staging, authToken_staging)
			counter = 0
			events_staging = nil
			Log.Debug.Printf("Processing %v records", len(events_prod))
			IngestData(events_prod, endpoint_prod, authToken_prod)
			counter = 0
			events_prod = nil
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
func AddCampaignCounter(events []operations.EventOutput, campaignCounter map[string]int) {

	for _, event := range events {
		eventAttributes := event.EventAttributes
		if _, exists := eventAttributes[EventPropertyCampaign]; exists {
			campaignName := fmt.Sprintf("%v", eventAttributes[EventPropertyCampaign])
			campaignCounter[campaignName] = campaignCounter[campaignName] + 1
		}
	}
}

func ProcessAdwordsDataFromEventsFiles(eventsExecutionDate time.Time, env *string, dataConfig *string, endpoint_staging *string, authToken_staging *string, events_staging []operations.EventOutput, endpoint_prod *string, authToken_prod *string, events_prod []operations.EventOutput) {

	dateYesterday := eventsExecutionDate.AddDate(0, 0, -1).Format(constants.DATEFORMAT)
	fileForYesterday := constants.LOCALOUTPUTFOLDER + "/" + "campaign_perf_report_done" + dateYesterday + ".marker"
	if *env == "development" {

		if utils.DoesFileExist(fileForYesterday) {
			Log.Debug.Printf("Already processed at dev for yesterday %s", fileForYesterday)
			return
		}
	} else {
		if utils.DoesFileExistInCloud(fmt.Sprintf("%s/%s", constants.PROCESSEDFILESCLOUD, constants.LOCALOUTPUTFOLDER),
			constants.BUCKETNAME, fileForYesterday) {
			Log.Debug.Printf("Already processed at cloud for yesterday %s", fileForYesterday)
			return
		}
	}
	// this is just the max storage size of events_staging at one time
	storageSize := 100000
	counter := 0
	var files []string

	// pull data from processed directory
	if *env == "development" {
		files = utils.GetAllFiles(fmt.Sprintf("%s/%s", "../datagen", constants.LOCALOUTPUTFOLDER), *dataConfig)
	} else {
		files = utils.ListAllCloudFiles(fmt.Sprintf("%s/%s", constants.PROCESSEDFILESCLOUD, constants.LOCALOUTPUTFOLDER),
			constants.BUCKETNAME, *dataConfig)
	}

	workingDateYesterday := eventsExecutionDate.AddDate(0, 0, -1)
	workingFilePattern := fmt.Sprintf("_%v_%v_%v_",
		workingDateYesterday.Year(),
		workingDateYesterday.Month(),
		workingDateYesterday.Day())

	for _, element := range files {

		// filtering only yesterday's file
		if !strings.Contains(element, workingFilePattern) {
			continue
		}
		Log.Debug.Printf("Processing file %s", element)

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
			op = trimUrlParams(op)
			counter++
			events_staging = append(events_staging, op)
			events_prod = append(events_prod, op)

			if counter == storageSize {
				Log.Debug.Printf("Processing %v records", len(events_staging))
				AddCampaignCounter(events_staging, campaignCounterSTAGE)
				counter = 0
				events_staging = nil
				Log.Debug.Printf("Processing %v records", len(events_prod))
				AddCampaignCounter(events_prod, campaignCounterPROD)
				counter = 0
				events_prod = nil
			}
		}
		if counter != 0 {
			Log.Debug.Printf("Processing %v records", len(events_staging))
			AddCampaignCounter(events_staging, campaignCounterSTAGE)
			counter = 0
			events_staging = nil
			Log.Debug.Printf("Processing %v records", len(events_prod))
			AddCampaignCounter(events_prod, campaignCounterPROD)
			counter = 0
			events_prod = nil
		}

		Log.Debug.Printf("Execution complete for file: %s", element)

	}

	// run adwords ingestion
	IngestAdwordsData(workingDateYesterday, projectIDStage, adwordsCustomerAccountIDStage, campaignCounterSTAGE, endpoint_staging, authToken_staging)
	IngestAdwordsData(workingDateYesterday, projectIDProd, adwordsCustomerAccountIDProd, campaignCounterPROD, endpoint_prod, authToken_prod)

	// create marker file for today's run
	if *env == "development" {
		if !utils.CreateFile(fileForYesterday) {
			Log.Debug.Printf("Couldn't create file %s", fileForYesterday)
			return
		}
	} else {
		if !utils.CreateFileInCloud(fmt.Sprintf("%s/%s", constants.PROCESSEDFILESCLOUD,
			constants.LOCALOUTPUTFOLDER), constants.BUCKETNAME, fileForYesterday) {
			Log.Debug.Printf("Couldn't create file %s", fileForYesterday)
			return
		}
	}

	Log.Debug.Printf("Adwords data feeder execution is done for %s!", dateYesterday)
}

func trimUrlParams(op operations.EventOutput) operations.EventOutput {
	var event string
	if !(strings.HasPrefix(op.Event, "http") || strings.HasPrefix(op.Event, "https")) {
		event = "http://" + op.Event
	}
	_, err := url.ParseRequestURI(event)
	if err != nil {
		return op
	}
	u, _ := url.Parse(event)
	op.Event = u.Host + u.Path
	return op
}
