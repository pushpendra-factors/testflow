package task

import (
	"context"
	"encoding/json"
	L "factors/leadsquared"
	"factors/model/model"
	"factors/model/store"
	BQ "factors/services/bigquery"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	U "factors/util"

	bigquery "cloud.google.com/go/bigquery"

	"cloud.google.com/go/civil"
	log "github.com/sirupsen/logrus"
)

type AttributeValueTuple struct {
	Attribute string
	Value     string
}

type SchemaPropertyMetadataMapping struct {
	SchemaName   string
	PropertyType bigquery.FieldType
}

type PropertyMetadataObjectLeadSquared struct {
	SchemaName  string
	DataType    string
	DisplayName string
}

type IncrementalSyncResponse struct {
	RecordCount int
	Leads       []IncrementalSyncLeadResponse
}

type IncrementalSyncLeadResponse struct {
	LeadPropertyList []IncrementalSyncLeadListResponse
}

type IncrementalSyncLeadListResponse struct {
	Attribute string
	Value     interface{}
}

var DUPLICATESCHEMAERRORPREFIX = "Error 409: Already Exists"

func ifDuplicateSchema(err string) bool {
	if strings.Contains(err, DUPLICATESCHEMAERRORPREFIX) {
		return true
	} else {
		return false
	}
}
func LeadSquaredPull(projectId int64, configs map[string]interface{}) (map[string]interface{}, bool) {
	bigQuerySetting := model.BigquerySetting{
		BigqueryProjectID:       configs["BigqueryProjectId"].(string),
		BigqueryCredentialsJSON: configs["BigqueryCredential"].(string),
	}

	executionDate := configs["startTimestamp"].(int64) - U.SECONDS_IN_A_DAY
	resultStatus := make(map[string]interface{})
	ctx := context.Background()
	client, err := BQ.CreateBigqueryClient(&ctx, &bigQuerySetting)
	if err != nil {
		resultStatus["failure"] = "Failed to get bigquery client" + err.Error()
		log.WithError(err).Error("Failed to get bigquery client")
		return resultStatus, false
	}
	defer client.Close()

	setting, _ := store.GetStore().GetProjectSetting(projectId)
	var leadSquaredConfig model.LeadSquaredConfig
	PAGESIZE := 1000
	LOOKBACK := 90
	err = U.DecodePostgresJsonbToStructType(setting.LeadSquaredConfig, &leadSquaredConfig)
	if err != nil {
		resultStatus["failure"] = "Failed to decode leadsquared config" + err.Error()
		log.WithError(err).Error("Failed to decode leadsquared config")
		return resultStatus, false
	}
	leadSquaredUrlParams := map[string]string{
		"accessKey": leadSquaredConfig.AccessKey,
		"secretKey": leadSquaredConfig.SecretKey,
	}
	datasetID := leadSquaredConfig.BigqueryDataset
	for documentType, _ := range model.LeadSquaredHistoricalSyncEndpoint {
		tableID := model.LeadSquaredTableName[documentType]
		propertyMetadataList, errorStatus, msg := getMetadataDetails(documentType, leadSquaredConfig.Host, leadSquaredUrlParams)
		if errorStatus != false {
			resultStatus["error"] = msg
			log.Error(msg)
			return resultStatus, false
		}

		/*
			1. Create the dataset if doesnt exists
			2. Create table if does not exists
			3. Fetch the metadata of the table
			4. Check if there is any update needed for the schema
			5. If yes, update it
		*/
		if leadSquaredConfig.FirstTimeSync == false {
			log.Info("Creating Dataset")
			if err := client.Dataset(datasetID).Create(ctx, nil); err != nil {
				log.WithError(err).Error("Error Creating Dataset")
				resultStatus["error"] = err.Error()
				return resultStatus, false
			}
			log.Info("Creating Table")
			tableRef := client.Dataset(datasetID).Table(tableID)
			if err := tableRef.Create(ctx, nil); err != nil {
				log.WithError(err).Error("Error Creating Table")
				resultStatus["error"] = err.Error()
				return resultStatus, false
			}

		}
		log.Info("Getting Table Reference")
		tableRef := client.Dataset(datasetID).Table(tableID)
		log.Info("Getting Table CacheMeta Reference")
		meta, err := client.Dataset(datasetID).Table(tableID).Metadata(ctx)
		if err != nil {
			log.WithError(err).Error("Error Fetching CacheMeta")
			resultStatus["error"] = err.Error()
			return resultStatus, false
		}
		existingSchema := meta.Schema
		log.Info("Getting Columns")
		newSchema, needUpdate := getColumnsToBeAdded(existingSchema, propertyMetadataList)
		columnsToBeExtracted := ""
		for _, metadata := range propertyMetadataList {
			if columnsToBeExtracted == "" {
				columnsToBeExtracted = metadata.SchemaName
			} else {
				columnsToBeExtracted = columnsToBeExtracted + "," + metadata.SchemaName
			}

		}
		if needUpdate == true {
			log.Info("Updating Schema")
			update := bigquery.TableMetadataToUpdate{
				Schema: newSchema,
			}
			if _, err := tableRef.Update(ctx, update, meta.ETag); err != nil {
				log.WithError(err).Error("Error Updating Table")
				resultStatus["error"] = err.Error()
				return resultStatus, false
			}
		}
		metadataWithOrder := make([]SchemaPropertyMetadataMapping, 0)
		for _, v := range newSchema {
			metadataWithOrder = append(metadataWithOrder, SchemaPropertyMetadataMapping{
				SchemaName:   v.Name,
				PropertyType: v.Type,
			})
		}
		if leadSquaredConfig.FirstTimeSync == false {
			result, errorStatus, msg := DoHistoricalSync(leadSquaredConfig.Host, model.LeadSquaredHistoricalSyncEndpoint[documentType], leadSquaredUrlParams, columnsToBeExtracted, PAGESIZE, executionDate, LOOKBACK, datasetID, tableID, client, metadataWithOrder, ctx)
			if errorStatus == true {
				log.Error(msg)
				resultStatus["error"] = msg
				return resultStatus, false
			}
			errCode := store.GetStore().UpdateLeadSquaredFirstTimeSyncStatus(projectId)
			if errCode != http.StatusOK {
				log.Error("Failed to update first sync status in db")
				resultStatus["error"] = "Failed to update first sync status in db"
				return resultStatus, false
			}
			return result, true
		} else {
			result, errorStatus, msg := DoIncrementalSync(leadSquaredConfig.Host, model.LeadSquaredHistoricalSyncEndpoint[documentType], model.LeadSquaredDocumentEndpoint[documentType], leadSquaredUrlParams, columnsToBeExtracted, PAGESIZE, executionDate, LOOKBACK, datasetID, tableID, client, metadataWithOrder, ctx)
			if errorStatus == true {
				log.Error(msg)
				resultStatus["error"] = msg
				return resultStatus, false
			}
			return result, true
		}
	}

	return resultStatus, true
}

func addNewColumnToSchema(existingSchema bigquery.Schema, column PropertyMetadataObjectLeadSquared) bigquery.Schema {
	if column.DataType == "Number" {
		existingSchema = append(existingSchema,
			&bigquery.FieldSchema{Name: column.SchemaName, Type: bigquery.FloatFieldType},
		)
	} else if column.DataType == "Date" {
		existingSchema = append(existingSchema,
			&bigquery.FieldSchema{Name: column.SchemaName, Type: bigquery.DateTimeFieldType},
		)
	} else {
		existingSchema = append(existingSchema,
			&bigquery.FieldSchema{Name: column.SchemaName, Type: bigquery.StringFieldType},
		)
	}
	return existingSchema
}

func extractValue(value interface{}, dataType interface{}) interface{} {
	if value == nil {
		return nil
	}
	dataValueString := value.(string)
	if dataType == bigquery.FloatFieldType {
		dataValueNumber, _ := strconv.Atoi(dataValueString)
		return dataValueNumber
	} else if dataType == bigquery.DateTimeFieldType {
		dataValueTime, _ := time.Parse("2006-01-02 15:04:05.000", dataValueString)
		return civil.DateTimeOf(dataValueTime)
	} else {
		return value
	}
}

func getMetadataDetails(documentType string, host string, leadSquaredUrlParams map[string]string) ([]PropertyMetadataObjectLeadSquared, bool, string) {
	metadataEndpoint := model.LeadSquaredMetadataEndpoint[documentType]
	statusCode, responseMetadata, errorObj := L.HttpRequestWrapper(fmt.Sprintf("https://%s", host), metadataEndpoint, nil, nil, "GET", leadSquaredUrlParams)
	if statusCode != http.StatusOK || errorObj != nil {
		return nil, true, errorObj.Error()
	}
	byteSliceMetadata, err := json.Marshal(responseMetadata)
	if err != nil {
		return nil, true, err.Error()
	}
	propertyMetadataList := make([]PropertyMetadataObjectLeadSquared, 0)
	err = json.Unmarshal(byteSliceMetadata, &propertyMetadataList)
	if err != nil {
		return nil, true, err.Error()
	}
	return propertyMetadataList, false, ""
}

func getColumnsToBeAdded(existingSchema bigquery.Schema, propertyMetadataList []PropertyMetadataObjectLeadSquared) (bigquery.Schema, bool) {
	needUpdate := false
	existingColumns := make(map[string]bool)
	for _, v := range existingSchema {
		existingColumns[v.Name] = true
	}
	for _, propertyMetadata := range propertyMetadataList {
		if (existingColumns[propertyMetadata.SchemaName]) == false {
			existingSchema = addNewColumnToSchema(existingSchema, propertyMetadata)
			needUpdate = true
		}
	}
	if existingColumns["synced_at"] == false {
		existingSchema = addNewColumnToSchema(existingSchema, PropertyMetadataObjectLeadSquared{
			SchemaName:  "synced_at",
			DataType:    "Date",
			DisplayName: "synced_at",
		})
	}
	return existingSchema, needUpdate
}

func DoHistoricalSync(host string, endpoint string, urlParams map[string]string, columns string, pageSize int, executionTimestamp int64, lookback int, datasetID string, tableId string, client *bigquery.Client, propertyMetadataList []SchemaPropertyMetadataMapping, ctx context.Context) (map[string]interface{}, bool, string) {
	index := 1
	log.Info("Starting for all records created after the lookback")
	pastTimestamp := executionTimestamp - (int64(lookback) * U.SECONDS_IN_A_DAY)
	StartDateinLeadSquaredFormat := fmt.Sprintf("%v", time.Unix(int64(pastTimestamp), 0).Format("2006-01-02 15:04:05"))
	log.Info(fmt.Sprintf("Starting Historical Sync with date > %v", StartDateinLeadSquaredFormat))
	for {
		request := model.SearchLeadsByCriteriaRequest{
			Parameter: model.LeadSearchParameterObj{
				LookupName:  "CreatedOn",
				LookupValue: StartDateinLeadSquaredFormat,
				SqlOperator: ">",
			},
			Paging: model.PagingObj{
				PageIndex: index,
				PageSize:  pageSize,
			},
			Columns: model.ColumnsObj{
				IncludeCSV: columns,
			},
			Sorting: model.SortingObj{
				ColumnName: "CreatedOn",
				Direction:  "1",
			},
		}
		headers := map[string]string{
			"Content-Type": "application/json",
		}
		statusCode, responseHistSyncData, errorObj := L.HttpRequestWrapper(fmt.Sprintf("https://%s", host), endpoint, headers, request, "POST", urlParams)
		if statusCode != http.StatusOK || errorObj != nil {
			return nil, true, errorObj.Error()
		}
		byteSliceHistSync, err := json.Marshal(responseHistSyncData)
		if err != nil {
			return nil, true, err.Error()
		}
		histSyncData := make([]interface{}, 0)
		err = json.Unmarshal(byteSliceHistSync, &histSyncData)
		if err != nil {
			return nil, true, err.Error()
		}
		log.Info(fmt.Sprintf("ModifiedOn - Inserting %v rows with index %v no of records %v", pageSize, index, len(histSyncData)))
		errorStatus, msg := insertBigQueryRow(datasetID, tableId, client, histSyncData, propertyMetadataList, ctx)
		if errorStatus != false {
			return nil, true, msg
		}
		index++
		if len(histSyncData) < pageSize {
			break
		}
	}
	return nil, false, ""
}

func insertBigQueryRow(datasetID string, tableID string, client *bigquery.Client, data []interface{}, propertyMetadataList []SchemaPropertyMetadataMapping, ctx context.Context) (bool, string) {
	ins := client.Dataset(datasetID).Table(tableID).Inserter()
	var vss []*bigquery.ValuesSaver
	for _, lead := range data {
		properties := lead.(map[string]interface{})
		jsonRow := make([]bigquery.Value, 0)
		meta, _ := client.Dataset(datasetID).Table(tableID).Metadata(ctx)
		for _, metadataColumn := range propertyMetadataList {
			if metadataColumn.SchemaName == "synced_at" {
				jsonRow = append(jsonRow, civil.DateTimeOf(U.TimeNowIn(U.TimeZoneStringUTC)))
			} else {
				attrValue := extractValue(properties[metadataColumn.SchemaName], metadataColumn.PropertyType)
				jsonRow = append(jsonRow, attrValue)
			}
		}
		vss = append(vss, &bigquery.ValuesSaver{
			Schema:   meta.Schema,
			InsertID: "",
			Row:      jsonRow,
		})
	}
	log.Info("Starting Write")
	if err := ins.Put(ctx, vss); err != nil {
		return true, err.Error()
	}
	log.Info("Ending Write")
	return false, ""
}

func DoIncrementalSync(host string, histSyncEndpoint string, endpoint string, urlParams map[string]string, columns string, pageSize int, executionTimestamp int64, lookback int, datasetID string, tableId string, client *bigquery.Client, propertyMetadataList []SchemaPropertyMetadataMapping, ctx context.Context) (map[string]interface{}, bool, string) {
	log.Info("Starting Incremental Sync")
	index := 1
	totalRecordCountModifiedOn := 0
	startDateinLeadSquaredFormat := fmt.Sprintf("%v", time.Unix(int64(executionTimestamp), 0).Format("2006-01-02 15:04:05"))
	endDateinLeadSquaredFormat := fmt.Sprintf("%v", time.Unix(int64(executionTimestamp)+U.SECONDS_IN_A_DAY, 0).Format("2006-01-02 15:04:05"))
	log.Info(fmt.Sprintf("Starting Incremental Sync with date > %v < %v ", startDateinLeadSquaredFormat, endDateinLeadSquaredFormat))
	for {
		request := model.LeadsByDateRangeRequest{
			Parameter: model.ParameterObj{
				FromDate: startDateinLeadSquaredFormat,
				ToDate:   endDateinLeadSquaredFormat,
			},
			Paging: model.PagingObj{
				PageIndex: index,
				PageSize:  pageSize,
			},
			Columns: model.ColumnsObj{
				IncludeCSV: columns,
			},
		}
		headers := map[string]string{
			"Content-Type": "application/json",
		}
		statusCode, responseIncrSyncData, errorObj := L.HttpRequestWrapper(fmt.Sprintf("https://%s", host), endpoint, headers, request, "POST", urlParams)
		if statusCode != http.StatusOK || errorObj != nil {
			return nil, true, errorObj.Error()
		}
		byteSliceIncrSync, err := json.Marshal(responseIncrSyncData)
		if err != nil {
			return nil, true, err.Error()
		}
		var incrSyncData IncrementalSyncResponse
		err = json.Unmarshal(byteSliceIncrSync, &incrSyncData)
		if err != nil {
			return nil, true, err.Error()
		}
		dataForInsertion := make([]interface{}, 0)
		for _, lead := range incrSyncData.Leads {
			propertiesMap := make(map[string]interface{})
			for _, property := range lead.LeadPropertyList {
				propertiesMap[property.Attribute] = property.Value
			}
			dataForInsertion = append(dataForInsertion, propertiesMap)
		}
		log.Info(fmt.Sprintf("IncrementalSync - Inserting %v rows with index %v no of records %v", pageSize, index, len(dataForInsertion)))
		errorStatus, msg := insertBigQueryRow(datasetID, tableId, client, dataForInsertion, propertyMetadataList, ctx)
		if errorStatus != false {
			return nil, true, msg
		}
		index++
		totalRecordCountModifiedOn = totalRecordCountModifiedOn + len(dataForInsertion)
		if len(dataForInsertion) < pageSize {
			break
		}
	}
	index = 1
	totalRecordCountCreatedOn := 0
	log.Info("Starting for all records created after the lookback")
	StartDateinLeadSquaredFormat := fmt.Sprintf("%v", time.Unix(int64(executionTimestamp), 0).Format("2006-01-02 15:04:05"))
	log.Info(fmt.Sprintf("Starting Historical Sync with date > %v", StartDateinLeadSquaredFormat))
	for {
		request := model.SearchLeadsByCriteriaRequest{
			Parameter: model.LeadSearchParameterObj{
				LookupName:  "CreatedOn",
				LookupValue: StartDateinLeadSquaredFormat,
				SqlOperator: ">",
			},
			Paging: model.PagingObj{
				PageIndex: index,
				PageSize:  pageSize,
			},
			Columns: model.ColumnsObj{
				IncludeCSV: columns,
			},
			Sorting: model.SortingObj{
				ColumnName: "CreatedOn",
				Direction:  "1",
			},
		}
		headers := map[string]string{
			"Content-Type": "application/json",
		}
		statusCode, responseHistSyncData, errorObj := L.HttpRequestWrapper(fmt.Sprintf("https://%s", host), histSyncEndpoint, headers, request, "POST", urlParams)
		if statusCode != http.StatusOK || errorObj != nil {
			return nil, true, errorObj.Error()
		}
		byteSliceHistSync, err := json.Marshal(responseHistSyncData)
		if err != nil {
			return nil, true, err.Error()
		}
		histSyncData := make([]interface{}, 0)
		err = json.Unmarshal(byteSliceHistSync, &histSyncData)
		if err != nil {
			return nil, true, err.Error()
		}
		log.Info(fmt.Sprintf("ModifiedOn - Inserting %v rows with index %v no of records %v", pageSize, index, len(histSyncData)))
		errorStatus, msg := insertBigQueryRow(datasetID, tableId, client, histSyncData, propertyMetadataList, ctx)
		if errorStatus != false {
			return nil, true, msg
		}
		index++
		totalRecordCountCreatedOn = totalRecordCountCreatedOn + len(histSyncData)
		if len(histSyncData) < pageSize {
			break
		}
	}
	resultStatus := make(map[string]interface{})
	resultStatus["ModifiedOnRecords"] = totalRecordCountModifiedOn
	resultStatus["CreatedOnRecords"] = totalRecordCountCreatedOn
	return resultStatus, false, ""
}
