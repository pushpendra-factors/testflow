package task

import (
	"context"
	"encoding/json"
	L "factors/integration/leadsquared"
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

type SalesActivityMetadataObjectLeadSquared struct {
	ID          string
	Name        string
	DisplayName string
	Fields      []PropertyMetadataObjectLeadSquared
}

type IncrementalSyncResponse struct {
	RecordCount int
	Leads       []IncrementalSyncLeadResponse
}

type IncrementalSyncResponseSalesActivity struct {
	RecordCount int
	List        []interface{}
}

type IncrementalSyncLeadResponse struct {
	LeadPropertyList []IncrementalSyncLeadListResponse
}

type IncrementalSyncLeadListResponse struct {
	Attribute string
	Value     interface{}
}

var DUPLICATESCHEMAERRORPREFIX = "Error 409: Already Exists"

var LEADSQUARED_ACTIVITYCODE = map[string]int{
	model.LEADSQUARED_SALES_ACTIVITY: 30,
	model.LEADSQUARED_EMAIL_SENT:     212,
	model.LEADSQUARED_EMAIL_INFO:     355,
	model.LEADSQUARED_HAD_A_CALL:     206,
	model.LEADSQUARED_CALLED_A_CUST_NEGATIVE_REPLY : 204,
	model.LEADSQUARED_CALLED_A_CUST_POSITIVE_REPLY : 202,
	model.LEADSQUARED_CALLED_TO_COLLECT_REFERRAL : 311,
	model.LEADSQUARED_EMAIL_BOUNCED : 10,
	model.LEADSQUARED_EMAIL_LINK_CLICKED : 1,
	model.LEADSQUARED_EMAIL_MAILING_PREFERENCE_LINK_CLICKED : 8,
	model.LEADSQUARED_EMAIL_MARKED_SPAM : 9,
	model.LEASQUARED_EMAIL_NEGATIVE_RESPONSE : 13,
	model.LEASQUARED_EMAIL_NEUTRAL_RESPONSE : 14,
	model.LEASQUARED_EMAIL_POSITIVE_RESPONSE : 12,
	model.LEASQUARED_EMAIL_OPENED : 0,
	model.LEASQUARED_EMAIL_POSITVE_INBOUND_EMAIL : 18,
	model.LEASQUARED_EMAIL_RESUBSCRIBED : 15,
	model.LEADSQUARED_EMAIL_SUBSCRIBED_TO_BOOTCAMP : 43,
	model.LEADSQUARED_EMAIL_SUBSCRIBED_TO_COLLECTION : 44,
	model.LEADSQUARED_EMAIL_SUBSCRIBED_TO_EVENTS : 47,
	model.LEADSQUARED_EMAIL_SUBSCRIBED_TO_FESTIVAL : 45,
	model.LEADSQUARED_EMAIL_SUBSCRIBED_TO_INTERNNATIONAL_REACTIVATION : 46,
	model.LEADSQUARED_EMAIL_SUBSCRIBED_TO_NEWSLETTER : 51,
	model.LEADSQUARED_EMAIL_SUBSCRIBED_TO_REACTIVATION : 48,
	model.LEADSQUARED_EMAIL_SUBSCRIBED_TO_REFERRAL : 52,
	model.LEADSQUARED_EMAIL_SUBSCRIBED_TO_SURVEY : 49,
	model.LEADSQUARED_EMAIL_SUBSCRIBED_TO_TEST : 41,
	model.LEADSQUARED_EMAIL_SUBSCRIBED_TO_WORKSHOP : 50,
	model.LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_BOOTCAMP : 63,
	model.LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_COLLECTION : 64,
	model.LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_EVENTS : 67,
	model.LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_FESTIVAL : 65,
	model.LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_INTERNNATIONAL_REACTIVATION : 66,
	model.LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_NEWSLETTER : 71,
	model.LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_REACTIVATION : 68,
	model.LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_REFERRAL : 72,
	model.LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_SURVEY : 69,
	model.LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_TEST : 61,
	model.LEADSQUARED_EMAIL_UNSUBSCRIBED_TO_WORKSHOP : 70,
	model.LEADSQUARED_EMAIL_UNSUBSCRIBE_LINK_CLICKED : 6,
	model.LEADSQUARED_EMAIL_UNSUBSCRIBED : 5,
	model.LEADSQUARED_EMAIL_VIEW_IN_BROWSER_LINK_CLICKED : 11,
	model.LEADSQUARED_EMAIL_RECEIVED : 211,

}

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
	PAGESIZE := 500
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
	for documentType, _ := range model.LeadSquaredTableName {
		_, _, isDone := store.GetStore().GetLeadSquaredMarker(projectId, executionDate, documentType, "incremental_sync")
		if(isDone == true){
			continue
		}
		log.Info(fmt.Sprintf("Starting for %v", documentType))
		tableID := model.LeadSquaredTableName[documentType]
		if documentType != model.LEADSQUARED_LEAD && documentType != model.LEADSQUARED_SALES_ACTIVITY {
			if projectId == 2251799831000006 {
				continue
			}
		}
		if model.ActivityEvents[documentType] == true {
			leadSquaredUrlParams["code"] = fmt.Sprintf("%v",LEADSQUARED_ACTIVITYCODE[documentType])
		}
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
		log.Info(fmt.Sprintf("First Time Sync status - %v", leadSquaredConfig.FirstTimeSync))
		if leadSquaredConfig.FirstTimeSync == false {
			log.Info("Creating Dataset")
			if err := client.Dataset(datasetID).Create(ctx, nil); err != nil {
				if ifDuplicateSchema(err.Error()) {
					log.WithError(err).Error("Error Creating Dataset - already exist but skipping the error")
				} else {
					log.WithError(err).Error("Error Creating Dataset")
					resultStatus["error"] = err.Error()
					return resultStatus, false
				}
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
		log.Info("Getting Table Metadata Reference")
		meta, err := client.Dataset(datasetID).Table(tableID).Metadata(ctx)
		if err != nil {
			log.WithError(err).Error("Error Fetching Metadata")
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
			if documentType == model.LEADSQUARED_LEAD {
				result, errorStatus, msg := DoHistoricalSync(projectId, leadSquaredConfig.Host, model.LeadSquaredHistoricalSyncEndpoint[documentType], leadSquaredUrlParams, columnsToBeExtracted, PAGESIZE, executionDate, LOOKBACK, datasetID, tableID, client, metadataWithOrder, ctx)
				if errorStatus == true {
					log.Error(msg)
					resultStatus["error"] = msg
					return resultStatus, false
				}
				for key, value := range result {
					resultStatus[key] = value
				}
			}
		} else {
			result, errorStatus, msg := DoIncrementalSync(projectId, documentType, leadSquaredConfig.Host, model.LeadSquaredHistoricalSyncEndpoint[documentType], model.LeadSquaredDocumentEndpoint(documentType), leadSquaredUrlParams, columnsToBeExtracted, PAGESIZE, executionDate, LOOKBACK, datasetID, tableID, client, metadataWithOrder, ctx)
			if errorStatus == true {
				log.Error(msg)
				resultStatus["error"] = msg
				return resultStatus, false
			}
			for key, value := range result {
				resultStatus[key] = value
			}
		}
	}
	if leadSquaredConfig.FirstTimeSync == false {
		errCode := store.GetStore().UpdateLeadSquaredFirstTimeSyncStatus(projectId)
		if errCode != http.StatusOK {
			log.Error("Failed to update first sync status in db")
			resultStatus["error"] = "Failed to update first sync status in db"
			return resultStatus, false
		}
	}
	return resultStatus, true
}

func addNewColumnToSchema(existingSchema bigquery.Schema, column PropertyMetadataObjectLeadSquared) bigquery.Schema {
	if column.DataType == "Number" {
		existingSchema = append(existingSchema,
			&bigquery.FieldSchema{Name: column.SchemaName, Type: bigquery.FloatFieldType},
		)
	} else if column.DataType == "Date" || column.DataType == "DateTime" {
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
		dataValueTime, err := time.Parse("2006-01-02 15:04:05.000", dataValueString)
		if err != nil {
			dataValueTime, _ = time.Parse("2006-01-02 15:04:05", dataValueString)
		}
		return civil.DateTimeOf(dataValueTime)
	} else {
		return value
	}
}

func getMetadataDetails(documentType string, host string, leadSquaredUrlParams map[string]string) ([]PropertyMetadataObjectLeadSquared, bool, string) {
	metadataEndpoint := model.LeadSquaredMetadataEndpoint(documentType)
	statusCode, responseMetadata, errorObj := L.HttpRequestWrapper(fmt.Sprintf("https://%s", host), metadataEndpoint, nil, nil, "GET", leadSquaredUrlParams)
	if statusCode != http.StatusOK || errorObj != nil {
		return nil, true, errorObj.Error()
	}
	byteSliceMetadata, err := json.Marshal(responseMetadata)
	if err != nil {
		return nil, true, err.Error()
	}
	propertyMetadata := make([]PropertyMetadataObjectLeadSquared, 0)
	if documentType == model.LEADSQUARED_LEAD {
		propertyMetadataList := make([]PropertyMetadataObjectLeadSquared, 0)
		err = json.Unmarshal(byteSliceMetadata, &propertyMetadataList)
		if err != nil {
			return nil, true, err.Error()
		}
		propertyMetadata = propertyMetadataList
	}
	if model.ActivityEvents[documentType] == true {
		var salesActivityMetadata SalesActivityMetadataObjectLeadSquared
		err = json.Unmarshal(byteSliceMetadata, &salesActivityMetadata)
		if err != nil {
			return nil, true, err.Error()
		}
		propertyMetadata = salesActivityMetadata.Fields
		propertyMetadata = addConstantFieldsForActivity(documentType, propertyMetadata)
	}
	return propertyMetadata, false, ""
}

func addConstantFieldsForActivity(documentType string, property []PropertyMetadataObjectLeadSquared) []PropertyMetadataObjectLeadSquared {
	property = append(property, PropertyMetadataObjectLeadSquared{SchemaName: "ProspectActivityId", DataType: "String", DisplayName: "ProspectActivityId"})
	property = append(property, PropertyMetadataObjectLeadSquared{SchemaName: "RelatedProspectId", DataType: "String", DisplayName: "RelatedProspectId"})
	property = append(property, PropertyMetadataObjectLeadSquared{SchemaName: "ActivityType", DataType: "String", DisplayName: "ActivityType"})
	property = append(property, PropertyMetadataObjectLeadSquared{SchemaName: "ActivityEvent", DataType: "String", DisplayName: "ActivityEvent"})
	if(documentType == model.LEADSQUARED_SALES_ACTIVITY || documentType == model.LEADSQUARED_EMAIL_SENT){
		property = append(property, PropertyMetadataObjectLeadSquared{SchemaName: "ActivityEvent_Note", DataType: "String", DisplayName: "ActivityEvent_Note"})
	}
	property = append(property, PropertyMetadataObjectLeadSquared{SchemaName: "CreatedOn", DataType: "DateTime", DisplayName: "CreatedOn"})
	property = append(property, PropertyMetadataObjectLeadSquared{SchemaName: "CreatedBy", DataType: "String", DisplayName: "CreatedBy"})
	property = append(property, PropertyMetadataObjectLeadSquared{SchemaName: "CreatedByEmailAddress", DataType: "String", DisplayName: "CreatedByEmailAddress"})
	property = append(property, PropertyMetadataObjectLeadSquared{SchemaName: "CreatedByName", DataType: "String", DisplayName: "CreatedByName"})
	property = append(property, PropertyMetadataObjectLeadSquared{SchemaName: "ModifiedOn", DataType: "DateTime", DisplayName: "ModifiedOn"})
	property = append(property, PropertyMetadataObjectLeadSquared{SchemaName: "ModifiedBy", DataType: "String", DisplayName: "ModifiedBy"})
	property = append(property, PropertyMetadataObjectLeadSquared{SchemaName: "ModifiedByEmailAddress", DataType: "String", DisplayName: "ModifiedByEmailAddress"})
	property = append(property, PropertyMetadataObjectLeadSquared{SchemaName: "ModifiedByName", DataType: "String", DisplayName: "ModifiedByName"})
	property = append(property, PropertyMetadataObjectLeadSquared{SchemaName: "CreatedOnUnix", DataType: "Number", DisplayName: "CreatedOnUnix"})
	return property
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

func DoHistoricalSync(projectId int64, host string, endpoint string, urlParams map[string]string, columns string, pageSize int, executionTimestamp int64, lookback int, datasetID string, tableId string, client *bigquery.Client, propertyMetadataList []SchemaPropertyMetadataMapping, ctx context.Context) (map[string]interface{}, bool, string) {
	indexNumber, _, isDone := store.GetStore().GetLeadSquaredMarker(projectId, executionTimestamp, model.LEADSQUARED_LEAD, "historical_sync")
	if(isDone == true){
		log.Info("Done. Skipping - Historical Sync")
		return nil, false, ""
	}
	index := 0
	if indexNumber == 0 {
		index = 1
	} else {
		index = indexNumber
	}
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
			store.GetStore().CreateLeadSquaredMarker(model.LeadsquaredMarker{
				ProjectID:   projectId,
				Delta:       executionTimestamp,
				Document:    model.LEADSQUARED_LEAD,
				Tag:         "historical_sync",
				IndexNumber: index,
			})
			return nil, true, errorObj.Error()
		}
		byteSliceHistSync, err := json.Marshal(responseHistSyncData)
		if err != nil {
			store.GetStore().CreateLeadSquaredMarker(model.LeadsquaredMarker{
				ProjectID:   projectId,
				Delta:       executionTimestamp,
				Document:    model.LEADSQUARED_LEAD,
				Tag:         "historical_sync",
				IndexNumber: index,
			})
			return nil, true, err.Error()
		}
		histSyncData := make([]interface{}, 0)
		err = json.Unmarshal(byteSliceHistSync, &histSyncData)
		if err != nil {
			store.GetStore().CreateLeadSquaredMarker(model.LeadsquaredMarker{
				ProjectID:   projectId,
				Delta:       executionTimestamp,
				Document:    model.LEADSQUARED_LEAD,
				Tag:         "historical_sync",
				IndexNumber: index,
			})
			return nil, true, err.Error()
		}
		log.Info(fmt.Sprintf("ModifiedOn - Inserting %v rows with index %v no of records %v", pageSize, index, len(histSyncData)))
		errorStatus, msg := insertBigQueryRow(datasetID, tableId, client, histSyncData, propertyMetadataList, ctx)
		if errorStatus != false {
			store.GetStore().CreateLeadSquaredMarker(model.LeadsquaredMarker{
				ProjectID:   projectId,
				Delta:       executionTimestamp,
				Document:    model.LEADSQUARED_LEAD,
				Tag:         "historical_sync",
				IndexNumber: index,
			})
			return nil, true, msg
		}
		log.Info("Insert done")
		index++
		if len(histSyncData) < pageSize {
			store.GetStore().CreateLeadSquaredMarker(model.LeadsquaredMarker{
				ProjectID:   projectId,
				Delta:       executionTimestamp,
				Document:    model.LEADSQUARED_LEAD,
				Tag:         "historical_sync",
				IndexNumber: index,
				IsDone: 	 true,
			})
			break
		}
	}
	log.Info("Done historical sync for all records created after the lookback")
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

func DoIncrementalSync(projectId int64, documentType string, host string, histSyncEndpoint string, endpoint string, urlParams map[string]string, columns string, pageSize int, executionTimestamp int64, lookback int, datasetID string, tableId string, client *bigquery.Client, propertyMetadataList []SchemaPropertyMetadataMapping, ctx context.Context) (map[string]interface{}, bool, string) {
	log.Info("Starting Incremental Sync")
	indexNumber, _, isDone := store.GetStore().GetLeadSquaredMarker(projectId, executionTimestamp, documentType, "incremental_sync")
	index := 0
	if indexNumber == 0 {
		index = 1
	} else {
		index = indexNumber
	}
	totalRecordCountModifiedOn := 0
	totalRecordCountCreatedOn := 0
	startDateinLeadSquaredFormat := fmt.Sprintf("%v", time.Unix(int64(executionTimestamp), 0).Format("2006-01-02 15:04:05"))
	endDateinLeadSquaredFormat := fmt.Sprintf("%v", time.Unix(int64(executionTimestamp)+U.SECONDS_IN_A_DAY, 0).Format("2006-01-02 15:04:05"))
	log.Info(fmt.Sprintf("Starting Incremental Sync with date > %v < %v ", startDateinLeadSquaredFormat, endDateinLeadSquaredFormat))
	for {
		if(isDone == true){
			log.Info("Done. Skipping - " + documentType)
			break
		}
		var request interface{}
		if documentType == model.LEADSQUARED_LEAD {
			request = model.LeadsByDateRangeRequest{
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
		}
		if model.ActivityEvents[documentType] == true {
			request = model.SearchSalesActivityByCriteriaRequest{
				Parameter: model.SalesActivitySearchParameterObj{
					FromDate:         startDateinLeadSquaredFormat,
					ToDate:           endDateinLeadSquaredFormat,
					ActivityEvent:    LEADSQUARED_ACTIVITYCODE[documentType],
					RemoveEmptyValue: false,
				},
				Paging: model.PagingObj{
					PageIndex: index,
					PageSize:  pageSize,
				},
				Sorting: model.SortingObj{
					ColumnName: "CreatedOn",
					Direction:  "1",
				},
			}
		}
		headers := map[string]string{
			"Content-Type": "application/json",
		}
		statusCode, responseIncrSyncData, errorObj := L.HttpRequestWrapper(fmt.Sprintf("https://%s", host), endpoint, headers, request, "POST", urlParams)
		if statusCode != http.StatusOK || errorObj != nil {
			store.GetStore().CreateLeadSquaredMarker(model.LeadsquaredMarker{
				ProjectID:   projectId,
				Delta:       executionTimestamp,
				Document:    documentType,
				Tag:         "incremental_sync",
				IndexNumber: index,
			})
			return nil, true, errorObj.Error()
		}
		byteSliceIncrSync, err := json.Marshal(responseIncrSyncData)
		if err != nil {
			store.GetStore().CreateLeadSquaredMarker(model.LeadsquaredMarker{
				ProjectID:   projectId,
				Delta:       executionTimestamp,
				Document:    documentType,
				Tag:         "incremental_sync",
				IndexNumber: index,
			})
			return nil, true, err.Error()
		}
		dataForInsertion := make([]interface{}, 0)
		if documentType == model.LEADSQUARED_LEAD {
			var incrSyncData IncrementalSyncResponse
			err = json.Unmarshal(byteSliceIncrSync, &incrSyncData)
			if err != nil {
				store.GetStore().CreateLeadSquaredMarker(model.LeadsquaredMarker{
					ProjectID:   projectId,
					Delta:       executionTimestamp,
					Document:    documentType,
					Tag:         "incremental_sync",
					IndexNumber: index,
				})
				return nil, true, err.Error()
			}
			for _, lead := range incrSyncData.Leads {
				propertiesMap := make(map[string]interface{})
				for _, property := range lead.LeadPropertyList {
					propertiesMap[property.Attribute] = property.Value
				}
				dataForInsertion = append(dataForInsertion, propertiesMap)
			}
		}
		if model.ActivityEvents[documentType] == true {
			var incrSyncData IncrementalSyncResponseSalesActivity
			err = json.Unmarshal(byteSliceIncrSync, &incrSyncData)
			if err != nil {
				store.GetStore().CreateLeadSquaredMarker(model.LeadsquaredMarker{
					ProjectID:   projectId,
					Delta:       executionTimestamp,
					Document:    documentType,
					Tag:         "incremental_sync",
					IndexNumber: index,
				})
				return nil, true, err.Error()
			}
			for _, sa := range incrSyncData.List {
				propertiesMap := sa.(map[string]interface{})
				t, _ := time.Parse(U.DATETIME_FORMAT_DB, propertiesMap["CreatedOn"].(string))
				createdOnUnix := t.Unix()
				propertiesMap["CreatedOnUnix"] = fmt.Sprintf("%v", createdOnUnix)
				dataForInsertion = append(dataForInsertion, propertiesMap)
			}
		}
		log.Info(fmt.Sprintf("IncrementalSync - Inserting %v rows with index %v no of records %v", pageSize, index, len(dataForInsertion)))
		errorStatus, msg := insertBigQueryRow(datasetID, tableId, client, dataForInsertion, propertyMetadataList, ctx)
		if errorStatus != false {
			store.GetStore().CreateLeadSquaredMarker(model.LeadsquaredMarker{
				ProjectID:   projectId,
				Delta:       executionTimestamp,
				Document:    documentType,
				Tag:         "incremental_sync",
				IndexNumber: index,
			})
			return nil, true, msg
		}
		index++
		totalRecordCountModifiedOn = totalRecordCountModifiedOn + len(dataForInsertion)
		if len(dataForInsertion) < pageSize {
			store.GetStore().CreateLeadSquaredMarker(model.LeadsquaredMarker{
				ProjectID:   projectId,
				Delta:       executionTimestamp,
				Document:    documentType,
				Tag:         "incremental_sync",
				IndexNumber: index,
				IsDone: 	 true,
			})
			break
		}
	}
	if documentType == model.LEADSQUARED_LEAD {
		indexNumber, _, isDone := store.GetStore().GetLeadSquaredMarker(projectId, executionTimestamp, documentType, "incremental_sync_created_at")
		index = 0
		if indexNumber == 0 {
			index = 1
		} else {
			index = indexNumber
		}
		log.Info("Starting for all records created after the lookback")
		StartDateinLeadSquaredFormat := fmt.Sprintf("%v", time.Unix(int64(executionTimestamp), 0).Format("2006-01-02 15:04:05"))
		log.Info(fmt.Sprintf("Starting Historical Sync with date > %v", StartDateinLeadSquaredFormat))
		for {
			if(isDone == true){
				log.Info("Done. Skipping - " + documentType)
				break
			}
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
				store.GetStore().CreateLeadSquaredMarker(model.LeadsquaredMarker{
					ProjectID:   projectId,
					Delta:       executionTimestamp,
					Document:    documentType,
					Tag:         "incremental_sync_created_at",
					IndexNumber: index,
				})
				return nil, true, errorObj.Error()
			}
			byteSliceHistSync, err := json.Marshal(responseHistSyncData)
			if err != nil {
				store.GetStore().CreateLeadSquaredMarker(model.LeadsquaredMarker{
					ProjectID:   projectId,
					Delta:       executionTimestamp,
					Document:    documentType,
					Tag:         "incremental_sync_created_at",
					IndexNumber: index,
				})
				return nil, true, err.Error()
			}
			histSyncData := make([]interface{}, 0)
			err = json.Unmarshal(byteSliceHistSync, &histSyncData)
			if err != nil {
				store.GetStore().CreateLeadSquaredMarker(model.LeadsquaredMarker{
					ProjectID:   projectId,
					Delta:       executionTimestamp,
					Document:    documentType,
					Tag:         "incremental_sync_created_at",
					IndexNumber: index,
				})
				return nil, true, err.Error()
			}
			log.Info(fmt.Sprintf("ModifiedOn - Inserting %v rows with index %v no of records %v", pageSize, index, len(histSyncData)))
			errorStatus, msg := insertBigQueryRow(datasetID, tableId, client, histSyncData, propertyMetadataList, ctx)
			if errorStatus != false {
				store.GetStore().CreateLeadSquaredMarker(model.LeadsquaredMarker{
					ProjectID:   projectId,
					Delta:       executionTimestamp,
					Document:    documentType,
					Tag:         "incremental_sync_created_at",
					IndexNumber: index,
				})
				return nil, true, msg
			}
			index++
			totalRecordCountCreatedOn = totalRecordCountCreatedOn + len(histSyncData)
			if len(histSyncData) < pageSize {
				store.GetStore().CreateLeadSquaredMarker(model.LeadsquaredMarker{
					ProjectID:   projectId,
					Delta:       executionTimestamp,
					Document:    documentType,
					Tag:         "incremental_sync_created_at",
					IndexNumber: index,
					IsDone: 	 true,
				})
				break
			}
		}
	}
	resultStatus := make(map[string]interface{})
	resultStatus["ModifiedOnRecords"] = totalRecordCountModifiedOn
	resultStatus["CreatedOnRecords"] = totalRecordCountCreatedOn
	return resultStatus, false, ""
}
