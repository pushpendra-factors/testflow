package task

import (
	"bufio"
	"factors/filestore"
	"factors/model/model"
	"factors/model/store"
	serviceDisk "factors/services/disk"
	U "factors/util"
	"fmt"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

var CAMPAIGN_ADS_IMPORT = "campaign"
var ADGROUP_ADS_IMPORT = "adgroup"
var KEYWORD_ADS_IMPORT = "keyword"

var CAMPAIGN_ADS_IMPORT_SUCCESS = "campaign_success"
var ADGROUP_ADS_IMPORT_SUCCESS = "adgroup_success"
var KEYWORD_ADS_IMPORT_SUCCESS = "keyword_success"

var CAMPAIGN_ADS_IMPORT_FAILURE = "campaign_failure"
var ADGROUP_ADS_IMPORT_FAILURE = "adgroup_failure"
var KEYWORD_ADS_IMPORT_FAILURE = "keyword_failure"

var CAMPAIGN_ADS_IMPORT_FAILURE_INDEX = "campaign_failure_index"
var ADGROUP_ADS_IMPORT_FAILURE_INDEX = "adgroup_failure_index"
var KEYWORD_ADS_IMPORT_FAILURE_INDEX = "keyword_failure_index"

var CAMPAIGN = 1
var ADGROUP = 2
var KEYWORD = 3
var CAMPAIGN_PERF_REPORT = 4
var ADGROUP_PERF_REPORT = 5
var KEYWORD_PERF_REPORT = 6
var ACCOUNT = 7

type CampaignPerformanceValues struct {
	AccountId      string  `json:"account_id"`
	CampaignType   string  `json:"campaign_type"`
	CampaignId     string  `json:"campaign_id"`
	CampaignName   string  `json:"campaign_name"`
	CampaignStatus string  `json:"campaign_status"`
	Impressions    int64   `json:"impressions"`
	Clicks         int64   `json:"clicks"`
	Spend          float64 `json:"spend"`
}

type CampaignMetadata struct {
	AccountId      string `json:"account_id"`
	CampaignType   string `json:"type"`
	CampaignId     string `json:"id"`
	CampaignName   string `json:"name"`
	CampaignStatus string `json:"status"`
}

type AdGroupPerformanceValues struct {
	AccountId      string  `json:"account_id"`
	CampaignType   string  `json:"campaign_type"`
	CampaignId     string  `json:"campaign_id"`
	CampaignName   string  `json:"campaign_name"`
	CampaignStatus string  `json:"campaign_status"`
	AdGroupType    string  `json:"ad_group_type"`
	AdGroupId      string  `json:"ad_group_id"`
	AdGroupName    string  `json:"ad_group_name"`
	AdGroupStatus  string  `json:"ad_group_status"`
	Impressions    int64   `json:"impressions"`
	Clicks         int64   `json:"clicks"`
	Spend          float64 `json:"spend"`
}

type AdGroupMetadata struct {
	AccountId      string `json:"account_id"`
	CampaignType   string `json:"campaign_type"`
	CampaignId     string `json:"campaign_id"`
	CampaignName   string `json:"campaign_name"`
	CampaignStatus string `json:"campaign_status"`
	AdGroupType    string `json:"type"`
	AdGroupId      string `json:"id"`
	AdGroupName    string `json:"name"`
	AdGroupStatus  string `json:"status"`
}

type KeywordPerformanceValues struct {
	AccountId      string  `json:"account_id"`
	CampaignType   string  `json:"campaign_type"`
	CampaignId     string  `json:"campaign_id"`
	CampaignName   string  `json:"campaign_name"`
	CampaignStatus string  `json:"campaign_status"`
	AdGroupType    string  `json:"ad_group_type"`
	AdGroupId      string  `json:"ad_group_id"`
	AdGroupName    string  `json:"ad_group_name"`
	AdGroupStatus  string  `json:"ad_group_status"`
	KeywordType    string  `json:"keyword_match_type"`
	KeywordId      string  `json:"keyword_id"`
	KeywordName    string  `json:"keyword_name"`
	KeywordStatus  string  `json:"keyword_status"`
	Impressions    int64   `json:"impressions"`
	Clicks         int64   `json:"clicks"`
	Spend          float64 `json:"spend"`
}

type KeywordMetadata struct {
	AccountId      string `json:"account_id"`
	CampaignType   string `json:"campaign_type"`
	CampaignId     string `json:"campaign_id"`
	CampaignName   string `json:"campaign_name"`
	CampaignStatus string `json:"campaign_status"`
	AdGroupType    string `json:"ad_group_type"`
	AdGroupId      string `json:"ad_group_id"`
	AdGroupName    string `json:"ad_group_name"`
	AdGroupStatus  string `json:"ad_group_status"`
	KeywordType    string `json:"match_type"`
	KeywordId      string `json:"id"`
	KeywordName    string `json:"name"`
	KeywordStatus  string `json:"status"`
}

type AccountMetadata struct {
	AccountId string `json:"id"`
}

func AdsImport(projectId int64, configs map[string]interface{}) (map[string]interface{}, bool) {

	diskManager := configs["diskManager"].(*serviceDisk.DiskDriver)
	cloudManager := configs["cloudManager"].(*filestore.FileManager)

	mappings, _ := store.GetStore().GetAllAdsImportEnabledProjects()
	lastProcesssedByProjectId := mappings[projectId]

	log.Info(fmt.Sprintf("Last Processed %v", lastProcesssedByProjectId))
	successStatus, failureStatus, failureIndex := make(map[string]int), make(map[string]int), make(map[string]interface{})
	toBeUpdated := make(map[string]model.LastProcessedAdsImport)
	// Read the exact file with exact chunk number
	// Keep counting the line number
	// Read with pipe seperated notation
	// convert into a object
	// Call UpsertIntegrationDocument method
	// update the marker
	toBeUpdated[CAMPAIGN_ADS_IMPORT] = lastProcesssedByProjectId[CAMPAIGN_ADS_IMPORT]
	lastProcesssed := lastProcesssedByProjectId[CAMPAIGN_ADS_IMPORT]
	i := lastProcesssed.ChunkNo
	for {
		exist := ChunkExist(projectId, i, CAMPAIGN_ADS_IMPORT, cloudManager)
		if !exist {
			log.Info(fmt.Sprintf("Campaign Chunk %v - not found", lastProcesssed.ChunkNo))
			break
		}
		scanner, err := GetAdsDataScanner(projectId, i, CAMPAIGN_ADS_IMPORT, cloudManager, diskManager)
		if err != nil {
			log.WithError(err).Error("Failed to get campaign scanner")
		}
		totalLinesRead, totalLinesProcessed, failedRecords := ScanAndParseData(scanner, CAMPAIGN_ADS_IMPORT, lastProcesssed, i, projectId)

		failureStatus[CAMPAIGN_ADS_IMPORT_FAILURE] += len(failedRecords)
		failureIndex[CAMPAIGN_ADS_IMPORT_FAILURE_INDEX] = failedRecords
		successStatus[CAMPAIGN_ADS_IMPORT_SUCCESS] += totalLinesProcessed - len(failedRecords)
		toBeUpdated[CAMPAIGN_ADS_IMPORT] = model.LastProcessedAdsImport{
			ChunkNo: i,
			LineNo:  totalLinesRead,
		}
		i++
	}
	lastProcesssed = lastProcesssedByProjectId[ADGROUP_ADS_IMPORT]
	i = lastProcesssed.ChunkNo
	toBeUpdated[ADGROUP_ADS_IMPORT] = lastProcesssedByProjectId[ADGROUP_ADS_IMPORT]
	for {
		exist := ChunkExist(projectId, i, ADGROUP_ADS_IMPORT, cloudManager)
		if !exist {
			log.Info(fmt.Sprintf("AdGroup Chunk %v - not found", lastProcesssed.ChunkNo))
			break
		}
		scanner, err := GetAdsDataScanner(projectId, i, ADGROUP_ADS_IMPORT, cloudManager, diskManager)
		if err != nil {
			log.WithError(err).Error("Failed to get adwords scanner")
		}
		totalLinesRead, totalLinesProcessed, failedRecords := ScanAndParseData(scanner, ADGROUP_ADS_IMPORT, lastProcesssed, i, projectId)
		failureStatus[ADGROUP_ADS_IMPORT_FAILURE] += len(failedRecords)
		failureIndex[ADGROUP_ADS_IMPORT_FAILURE_INDEX] = failedRecords
		successStatus[ADGROUP_ADS_IMPORT_SUCCESS] += totalLinesProcessed - len(failedRecords)
		toBeUpdated[ADGROUP_ADS_IMPORT] = model.LastProcessedAdsImport{
			ChunkNo: i,
			LineNo:  totalLinesRead,
		}
		i++
	}
	lastProcesssed = lastProcesssedByProjectId[KEYWORD_ADS_IMPORT]
	i = lastProcesssed.ChunkNo
	toBeUpdated[KEYWORD_ADS_IMPORT] = lastProcesssedByProjectId[KEYWORD_ADS_IMPORT]
	for {
		exist := ChunkExist(projectId, i, KEYWORD_ADS_IMPORT, cloudManager)
		if !exist {
			log.Info(fmt.Sprintf("Keyword Chunk %v - not found", lastProcesssed.ChunkNo))
			break
		}
		scanner, err := GetAdsDataScanner(projectId, i, KEYWORD_ADS_IMPORT, cloudManager, diskManager)
		if err != nil {
			log.WithError(err).Error("Failed to get keyword scanner")
		}
		totalLinesRead, totalLinesProcessed, failedRecords := ScanAndParseData(scanner, KEYWORD_ADS_IMPORT, lastProcesssed, i, projectId)
		failureStatus[KEYWORD_ADS_IMPORT_FAILURE] += len(failedRecords)
		failureIndex[KEYWORD_ADS_IMPORT_FAILURE_INDEX] = failedRecords
		successStatus[KEYWORD_ADS_IMPORT_SUCCESS] += totalLinesProcessed - len(failedRecords)
		toBeUpdated[KEYWORD_ADS_IMPORT] = model.LastProcessedAdsImport{
			ChunkNo: i,
			LineNo:  totalLinesRead,
		}
		i++
	}
	store.GetStore().UpdateLastProcessedAdsData(toBeUpdated, projectId)

	totalFailures := 0
	resultStatus := make(map[string]interface{})
	for tag, value := range failureStatus {
		resultStatus[tag] = value
		totalFailures += value
	}
	for tag, value := range successStatus {
		resultStatus[tag] = value
	}
	for tag, value := range failureIndex {
		resultStatus[tag] = value
	}
	if totalFailures > 0 {
		return resultStatus, false
	}
	return resultStatus, true
}

func ScanAndParseData(scanner *bufio.Scanner, report string, lastProcessed model.LastProcessedAdsImport, currentChunkNo int, projectId int64) (int, int, []int) {
	totalLinesRead := 0
	totalLinesProcessed := 0
	failedRecords := make([]int, 0)
	log.Info("Starting Scan")
	for scanner.Scan() {
		totalLinesRead++
		if lastProcessed.ChunkNo == currentChunkNo {
			if totalLinesRead <= lastProcessed.LineNo {
				continue
			}
		}
		totalLinesProcessed++
		adsDataLine := scanner.Text()
		var status bool
		if report == CAMPAIGN_ADS_IMPORT {
			status = ParseCampaignData(projectId, adsDataLine)
			if status == false {
				failedRecords = append(failedRecords, totalLinesRead)
			}
		}
		if report == ADGROUP_ADS_IMPORT {
			status = ParseAdGroupData(projectId, adsDataLine)
			if status == false {
				failedRecords = append(failedRecords, totalLinesRead)
			}
		}
		if report == KEYWORD_ADS_IMPORT {
			status = ParseKeywordData(projectId, adsDataLine)
			if status == false {
				failedRecords = append(failedRecords, totalLinesRead)
			}
		}
	}
	return totalLinesRead, totalLinesProcessed, failedRecords
}

func ParseCampaignData(projectId int64, data string) bool {
	dataSplit := strings.Split(data, ",")
	if len(dataSplit) != 11 {
		log.Error(fmt.Sprintf("Parse Campaign Data Failed - %v", data))
		return false
	}
	source := dataSplit[10]
	timestamp, _ := strconv.Atoi(dataSplit[9])
	impression, _ := strconv.Atoi(dataSplit[6])
	click, _ := strconv.Atoi(dataSplit[7])
	spend, _ := strconv.ParseFloat(dataSplit[8], 64)
	campaignPerfValues := CampaignPerformanceValues{
		Impressions:    int64(impression),
		Clicks:         int64(click),
		Spend:          spend,
		AccountId:      dataSplit[1],
		CampaignId:     dataSplit[3],
		CampaignStatus: dataSplit[5],
		CampaignType:   dataSplit[2],
		CampaignName:   dataSplit[4],
	}
	campaign := CampaignMetadata{
		AccountId:      dataSplit[1],
		CampaignId:     dataSplit[3],
		CampaignStatus: dataSplit[5],
		CampaignType:   dataSplit[2],
		CampaignName:   dataSplit[4],
	}
	account := AccountMetadata{
		AccountId: dataSplit[1],
	}

	valuesBlob, err := U.EncodeStructTypeToPostgresJsonb(&campaignPerfValues)
	if err != nil {
		log.WithError(err).Error("Error while encoding to jsonb")
		return false
	}
	intDocumentCampaignPerformance := model.IntegrationDocument{
		DocumentId:        dataSplit[3],
		ProjectID:         projectId,
		CustomerAccountID: dataSplit[1],
		Source:            source,
		DocumentType:      CAMPAIGN_PERF_REPORT,
		Timestamp:         int64(timestamp),
		Value:             valuesBlob,
	}

	valuesBlob, err = U.EncodeStructTypeToPostgresJsonb(&campaign)
	if err != nil {
		log.WithError(err).Error("Error while encoding to jsonb")
		return false
	}
	intDocumentCampaign := model.IntegrationDocument{
		DocumentId:        dataSplit[3],
		ProjectID:         projectId,
		CustomerAccountID: dataSplit[1],
		Source:            source,
		DocumentType:      CAMPAIGN,
		Timestamp:         int64(timestamp),
		Value:             valuesBlob,
	}
	valuesBlob, err = U.EncodeStructTypeToPostgresJsonb(&account)
	if err != nil {
		log.WithError(err).Error("Error while encoding to jsonb")
		return false
	}
	intDocumentAccount := model.IntegrationDocument{
		DocumentId:        dataSplit[1],
		ProjectID:         projectId,
		CustomerAccountID: dataSplit[1],
		Source:            source,
		DocumentType:      ACCOUNT,
		Timestamp:         int64(timestamp),
		Value:             valuesBlob,
	}
	err = store.GetStore().UpsertIntegrationDocument(intDocumentCampaignPerformance)
	if err != nil {
		return false
	}
	err = store.GetStore().UpsertIntegrationDocument(intDocumentCampaign)
	if err != nil {
		return false
	}
	err = store.GetStore().UpsertIntegrationDocument(intDocumentAccount)
	if err != nil {
		return false
	}
	return true
}

func ParseAdGroupData(projectId int64, data string) bool {
	dataSplit := strings.Split(data, ",")
	if len(dataSplit) != 15 {
		log.Error(fmt.Sprintf("Parse Adgroup Data Failed - %v", data))
		return false
	}
	source := dataSplit[14]
	timestamp, _ := strconv.Atoi(dataSplit[13])
	impression, _ := strconv.Atoi(dataSplit[10])
	click, _ := strconv.Atoi(dataSplit[11])
	spend, _ := strconv.ParseFloat(dataSplit[12], 64)
	adGroupPerfValues := AdGroupPerformanceValues{
		Impressions:    int64(impression),
		Clicks:         int64(click),
		Spend:          spend,
		AccountId:      dataSplit[1],
		CampaignId:     dataSplit[6],
		CampaignStatus: dataSplit[9],
		CampaignType:   dataSplit[7],
		CampaignName:   dataSplit[8],
		AdGroupType:    dataSplit[2],
		AdGroupId:      dataSplit[3],
		AdGroupName:    dataSplit[4],
		AdGroupStatus:  dataSplit[5],
	}
	adGroup := AdGroupMetadata{
		AccountId:      dataSplit[1],
		CampaignId:     dataSplit[6],
		CampaignStatus: dataSplit[9],
		CampaignType:   dataSplit[7],
		CampaignName:   dataSplit[8],
		AdGroupType:    dataSplit[2],
		AdGroupId:      dataSplit[3],
		AdGroupName:    dataSplit[4],
		AdGroupStatus:  dataSplit[5],
	}
	account := AccountMetadata{
		AccountId: dataSplit[1],
	}

	valuesBlob, err := U.EncodeStructTypeToPostgresJsonb(&adGroupPerfValues)
	if err != nil {
		log.WithError(err).Error("Error while encoding to jsonb")
		return false
	}
	intDocumentAdGroupPerformance := model.IntegrationDocument{
		DocumentId:        dataSplit[3],
		ProjectID:         projectId,
		CustomerAccountID: dataSplit[1],
		Source:            source,
		DocumentType:      ADGROUP_PERF_REPORT,
		Timestamp:         int64(timestamp),
		Value:             valuesBlob,
	}

	valuesBlob, err = U.EncodeStructTypeToPostgresJsonb(&adGroup)
	if err != nil {
		log.WithError(err).Error("Error while encoding to jsonb")
		return false
	}
	intDocumentAdGroup := model.IntegrationDocument{
		DocumentId:        dataSplit[3],
		ProjectID:         projectId,
		CustomerAccountID: dataSplit[1],
		Source:            source,
		DocumentType:      ADGROUP,
		Timestamp:         int64(timestamp),
		Value:             valuesBlob,
	}
	valuesBlob, err = U.EncodeStructTypeToPostgresJsonb(&account)
	if err != nil {
		log.WithError(err).Error("Error while encoding to jsonb")
		return false
	}
	intDocumentAccount := model.IntegrationDocument{
		DocumentId:        dataSplit[1],
		ProjectID:         projectId,
		CustomerAccountID: dataSplit[1],
		Source:            source,
		DocumentType:      ACCOUNT,
		Timestamp:         int64(timestamp),
		Value:             valuesBlob,
	}
	err = store.GetStore().UpsertIntegrationDocument(intDocumentAdGroupPerformance)
	if err != nil {
		return false
	}
	err = store.GetStore().UpsertIntegrationDocument(intDocumentAdGroup)
	if err != nil {
		return false
	}
	err = store.GetStore().UpsertIntegrationDocument(intDocumentAccount)
	if err != nil {
		return false
	}
	return true
}

func ParseKeywordData(projectId int64, data string) bool {
	dataSplit := strings.Split(data, ",")
	if len(dataSplit) != 19 {
		log.Error(fmt.Sprintf("Parse Keyword Data Failed - %v", data))
		return false
	}
	source := dataSplit[18]
	timestamp, _ := strconv.Atoi(dataSplit[17])
	impression, _ := strconv.Atoi(dataSplit[14])
	click, _ := strconv.Atoi(dataSplit[15])
	spend, _ := strconv.ParseFloat(dataSplit[16], 64)
	keywordPerfValues := KeywordPerformanceValues{
		Impressions:    int64(impression),
		Clicks:         int64(click),
		Spend:          spend,
		AccountId:      dataSplit[1],
		CampaignId:     dataSplit[10],
		CampaignStatus: dataSplit[13],
		CampaignType:   dataSplit[11],
		CampaignName:   dataSplit[12],
		AdGroupType:    dataSplit[7],
		AdGroupId:      dataSplit[6],
		AdGroupName:    dataSplit[8],
		AdGroupStatus:  dataSplit[9],
		KeywordType:    dataSplit[2],
		KeywordId:      dataSplit[3],
		KeywordName:    dataSplit[4],
		KeywordStatus:  dataSplit[5],
	}
	keyword := KeywordMetadata{
		AccountId:      dataSplit[1],
		CampaignId:     dataSplit[10],
		CampaignStatus: dataSplit[13],
		CampaignType:   dataSplit[11],
		CampaignName:   dataSplit[12],
		AdGroupType:    dataSplit[7],
		AdGroupId:      dataSplit[6],
		AdGroupName:    dataSplit[8],
		AdGroupStatus:  dataSplit[9],
		KeywordType:    dataSplit[2],
		KeywordId:      dataSplit[3],
		KeywordName:    dataSplit[4],
		KeywordStatus:  dataSplit[5],
	}
	account := AccountMetadata{
		AccountId: dataSplit[1],
	}

	valuesBlob, err := U.EncodeStructTypeToPostgresJsonb(&keywordPerfValues)
	if err != nil {
		log.WithError(err).Error("Error while encoding to jsonb")
		return false
	}
	intDocumentAdGroupPerformance := model.IntegrationDocument{
		DocumentId:        dataSplit[3],
		ProjectID:         projectId,
		CustomerAccountID: dataSplit[1],
		Source:            source,
		DocumentType:      KEYWORD_PERF_REPORT,
		Timestamp:         int64(timestamp),
		Value:             valuesBlob,
	}

	valuesBlob, err = U.EncodeStructTypeToPostgresJsonb(&keyword)
	if err != nil {
		log.WithError(err).Error("Error while encoding to jsonb")
		return false
	}
	intDocumentAdGroup := model.IntegrationDocument{
		DocumentId:        dataSplit[3],
		ProjectID:         projectId,
		CustomerAccountID: dataSplit[1],
		Source:            source,
		DocumentType:      KEYWORD,
		Timestamp:         int64(timestamp),
		Value:             valuesBlob,
	}
	valuesBlob, err = U.EncodeStructTypeToPostgresJsonb(&account)
	if err != nil {
		log.WithError(err).Error("Error while encoding to jsonb")
		return false
	}
	intDocumentAccount := model.IntegrationDocument{
		DocumentId:        dataSplit[1],
		ProjectID:         projectId,
		CustomerAccountID: dataSplit[1],
		Source:            source,
		DocumentType:      ACCOUNT,
		Timestamp:         int64(timestamp),
		Value:             valuesBlob,
	}
	err = store.GetStore().UpsertIntegrationDocument(intDocumentAdGroupPerformance)
	if err != nil {
		return false
	}
	err = store.GetStore().UpsertIntegrationDocument(intDocumentAdGroup)
	if err != nil {
		return false
	}
	err = store.GetStore().UpsertIntegrationDocument(intDocumentAccount)
	if err != nil {
		return false
	}
	return true
}

func ChunkExist(projectId int64, chunkNo int, report string, cloudManager *filestore.FileManager) bool {
	adsCloudPath, adsCloudName := (*cloudManager).GetAdsDataFilePathAndName(projectId, report, chunkNo)
	files := (*cloudManager).ListFiles(adsCloudPath)
	log.Info(fmt.Sprintf("Location: %v Files List %v", adsCloudName, files))
	for _, file := range files {
		log.Info(fmt.Sprintf("Checking File %v in %v", adsCloudName, file))
		if strings.Contains(file, adsCloudName) {
			return true
		}
	}
	return false
}

// GetAdsDataScanner Return handle to ads file scanner
func GetAdsDataScanner(projectId int64, chunkNo int, report string, cloudManager *filestore.FileManager,
	diskManager *serviceDisk.DiskDriver) (*bufio.Scanner, error) {
	var err error
	adsTmpPath, adsTmpName := diskManager.GetAdsDataFilePathAndName(projectId, report, chunkNo)
	adsFilePath := adsTmpPath + adsTmpName
	adsCloudPath, adsCloudName := (*cloudManager).GetAdsDataFilePathAndName(projectId, report, chunkNo)
	log.WithFields(log.Fields{"adsFileCloudPath": adsCloudPath,
		"adsFileCloudName": adsCloudName}).Info("Downloading ads file from cloud.")
	eReader, err := (*cloudManager).Get(adsCloudPath, adsCloudName)
	if err != nil {
		log.WithFields(log.Fields{"err": err, "adsFilePath": adsCloudPath,
			"adsFileName": adsCloudName}).Error("Failed downloading ads file from cloud.")
		return nil, err
	}
	err = diskManager.Create(adsTmpPath, adsTmpName, eReader)
	if err != nil {
		log.WithFields(log.Fields{"err": err, "adsFilePath": adsTmpPath,
			"adsFileName": adsTmpName}).Error("Failed downloading ads file from cloud.")
		return nil, err
	}

	scanner, err := OpenEventFileAndGetScanner(adsFilePath)
	if err != nil {
		log.WithFields(log.Fields{"err": err,
			"adsFilePath": adsFilePath}).Error("Failed opening event file and getting scanner.")
	}

	return scanner, err
}
