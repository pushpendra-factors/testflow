package task

import (
	"context"
	"factors/model/model"
	"factors/model/store"
	BQ "factors/services/bigquery"
	"net/http"

	U "factors/util"

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

func MarketoIntegration(projectId uint64, configs map[string]interface{}) (map[string]interface{}, bool) {

	bigQuerySetting := model.BigquerySetting{
		BigqueryProjectID:       configs["BigqueryProjectId"].(string),
		BigqueryCredentialsJSON: configs["BigqueryCredential"].(string),
	}

	resultStatus := make(map[string]interface{})
	executionDate := configs["startTimestamp"].(int64)
	executionDateString := U.GetDateOnlyHyphenFormatFromTimestampZ(executionDate)
	//executionDateStringYYYYMMDD, _ := strconv.ParseInt(U.GetDateOnlyFromTimestampZ(executionDate), 10, 64)
	ctx := context.Background()
	status := true
	client, err := BQ.CreateBigqueryClient(&ctx, &bigQuerySetting)
	if err != nil {
		status = false
		resultStatus["failure"] = "Failed to get bigquery client" + err.Error()
		log.WithError(err).Error("Failed to get bigquery client")
	}
	defer client.Close()

	mapping, err := store.GetStore().GetActiveFiveTranMapping(projectId, model.MarketoIntegration)
	if err != nil {
		status = false
		resultStatus["failure"] = "Error executing fivetran mapping query" + err.Error()
		log.WithError(err).Error("Error executing fivetran mapping query")
	}

	totalSuccess := 0
	totalFailures := 0
	PAGE_SIZE := 1000
	for docType, _ := range model.MarketoDocumentTypeAlias {
		offset := 0
		for {
			var query string
			if docType == model.MARKETO_TYPE_NAME_LEAD {
				tableRef := client.Dataset(mapping.SchemaID).Table("lead_segment")
				_, existsErr := tableRef.Metadata(ctx)
				if existsErr == nil {
					query = model.GetMarketoDocumentQuery(configs["BigqueryProjectId"].(string), mapping.SchemaID, model.MarketoDocumentToQuery[docType], executionDateString, docType, PAGE_SIZE, offset)
				} else {
					query = model.GetMarketoDocumentQuery(configs["BigqueryProjectId"].(string), mapping.SchemaID, model.MarketoDocumentToQuery[model.MARKETO_TYPE_NAME_LEAD_NO_SEGMENT], executionDateString, model.MARKETO_TYPE_NAME_LEAD_NO_SEGMENT, PAGE_SIZE, offset)
				}

			} else {
				query = model.GetMarketoDocumentQuery(configs["BigqueryProjectId"].(string), mapping.SchemaID, model.MarketoDocumentToQuery[docType], executionDateString, docType, PAGE_SIZE, offset)
			}
			var queryResult [][]string
			err = BQ.ExecuteQuery(&ctx, client, query, &queryResult)
			if err != nil {
				resultStatus["failure-"+docType] = "Error while executing query" + err.Error()
				status = false
				log.WithError(err).Error("Error while executing query")
			}
			metadataQuery, exists := model.GetMarketoDocumentMetadataQuery(docType, configs["BigqueryProjectId"].(string), mapping.SchemaID, model.MarketoDataObjectColumnsQuery[docType])
			var metadataQueryResult [][]string
			columnNamesFromMetadata := make([]string, 0)
			columnNamesFromMetadataDateTime := make(map[string]bool)
			if exists {
				err = BQ.ExecuteQuery(&ctx, client, metadataQuery, &metadataQueryResult)
				if err != nil {
					resultStatus["failure-"+docType] = "Error while executing metadata query" + err.Error()
					status = false
					log.WithError(err).Error("Error while executing query")
				}
				columnNamesFromMetadata, columnNamesFromMetadataDateTime = extractMetadataColumns(metadataQueryResult)
			}
			success, failures := InsertIntegrationDocument(projectId, docType, queryResult, columnNamesFromMetadata, columnNamesFromMetadataDateTime)
			totalFailures = totalFailures + failures
			totalSuccess = totalSuccess + success
			if len(queryResult) < PAGE_SIZE {
				break
			}
			offset = offset + PAGE_SIZE
		}
		resultStatus["failure-"+docType] = totalFailures
		resultStatus["success-"+docType] = totalSuccess
	}
	if status == false {
		return resultStatus, false
	}
	return resultStatus, true
}

func extractMetadataColumns(metadataQueryResult [][]string) ([]string, map[string]bool) {

	columnNamesFromMetadata := make([]string, 0)
	columnNamesFromMetadataDateTime := make(map[string]bool)
	for _, metadataLine := range metadataQueryResult {
		metadataIndex := model.GetMetadataColumnNameIndex("column_name")
		columnName := metadataLine[metadataIndex]
		columnNamesFromMetadata = append(columnNamesFromMetadata, columnName)
		metadataIndexForDataType := model.GetMetadataColumnNameIndex("data_type")
		dataType := metadataLine[metadataIndexForDataType]
		if dataType == "TIMESTAMP" {
			columnNamesFromMetadataDateTime[columnName] = true
		}
	}
	return columnNamesFromMetadata, columnNamesFromMetadataDateTime
}

func InsertIntegrationDocument(projectId uint64, docType string, queryResult [][]string, columnNamesFromMetadata []string, columnNamesFromMetadataDateTime map[string]bool) (int, int) {
	success := 0
	failures := 0
	for _, line := range queryResult {
		values := model.GetMarketoDocumentValues(docType, line, columnNamesFromMetadata, columnNamesFromMetadataDateTime)
		valuesBlob, err := U.EncodeToPostgresJsonb(&values)
		if err != nil {
			log.WithError(err).Error("Error while encoding to jsonb")
			failures++
			break
		}
		insertionStatus := int(0)
		var errCRMStatus error
		var logIndex string
		if model.DocTypeIntegrationObjectMap[docType] == "activity" {
			insertionStatus, errCRMStatus, logIndex = insertCRMActivity(projectId, line, docType, columnNamesFromMetadata, valuesBlob)
		}
		if model.DocTypeIntegrationObjectMap[docType] == "user" {
			insertionStatus, errCRMStatus, logIndex = insertCRMUser(projectId, line, docType, columnNamesFromMetadata, valuesBlob)
		}
		if errCRMStatus != nil || insertionStatus != http.StatusCreated {
			log.WithError(errCRMStatus).WithFields(log.Fields{
				"Status":   errCRMStatus,
				"DocValue": logIndex})
			failures++
		} else {
			success++
		}
	}
	return success, failures
}

func insertCRMActivity(projectId uint64, line []string, docType string, columnNamesFromMetadata []string, values *postgres.Jsonb) (int, error, string) {
	insertionStatus := int(0)
	var errCRMStatus error
	timestamps := model.GetMarketoDocumentTimestamp(docType, line, columnNamesFromMetadata)
	intDocument := model.CRMActivity{
		ProjectID:          projectId,
		ExternalActivityID: model.GetMarketoDocumentProgramId(docType, line, columnNamesFromMetadata),
		Source:             model.CRM_SOURCE_MARKETO,
		Name:               "program_membership_created",
		Type:               model.GetMarketoDocumentDocumentType(docType),
		ActorType:          model.GetMarketoActorType(docType),
		ActorID:            model.GetMarketoDocumentActorId(docType, line, columnNamesFromMetadata),
		Timestamp:          timestamps[0],
		Properties:         values,
	}

	insertionStatus, errCRMStatus = store.GetStore().CreateCRMActivity(&intDocument)
	if insertionStatus == http.StatusConflict {
		// Need to check if membership data is getting updatd to new one for a given progression status
		// Check what to do with missing leads
		intDocument.Name = "program_membership_updated"
		intDocument.Timestamp = timestamps[0]
		insertionStatus, errCRMStatus = store.GetStore().CreateCRMActivity(&intDocument)
	}
	return insertionStatus, errCRMStatus, model.GetUniqueLogValue(docType, line, columnNamesFromMetadata)
}

func insertCRMUser(projectId uint64, line []string, docType string, columnNamesFromMetadata []string, values *postgres.Jsonb) (int, error, string) {
	insertionStatus := int(0)
	var errCRMStatus error
	timestamps := model.GetMarketoDocumentTimestamp(docType, line, columnNamesFromMetadata)
	intDocument1 := model.CRMUser{
		ID:         model.GetMarketoUserId(docType, line, columnNamesFromMetadata),
		ProjectID:  projectId,
		Source:     model.CRM_SOURCE_MARKETO,
		Type:       model.GetMarketoDocumentDocumentType(docType),
		Timestamp:  timestamps[0],
		Properties: values,
		Email:      model.GetMarketoDocumentEmail(docType, line, columnNamesFromMetadata),
		Phone:      model.GetMarketoDocumentPhone(docType, line, columnNamesFromMetadata),
	}
	insertionStatus, errCRMStatus = store.GetStore().CreateCRMUser(&intDocument1)
	if insertionStatus == http.StatusCreated {
		intDocument2 := model.CRMUser{
			ID:         model.GetMarketoUserId(docType, line, columnNamesFromMetadata),
			ProjectID:  projectId,
			Source:     model.CRM_SOURCE_MARKETO,
			Type:       model.GetMarketoDocumentDocumentType(docType),
			Timestamp:  timestamps[1],
			Properties: values,
			Email:      model.GetMarketoDocumentEmail(docType, line, columnNamesFromMetadata),
			Phone:      model.GetMarketoDocumentPhone(docType, line, columnNamesFromMetadata),
		}
		insertionStatus, errCRMStatus = store.GetStore().CreateCRMUser(&intDocument2)
	}
	if model.GetMarketoDocumentAction(docType, line, columnNamesFromMetadata) == model.CRMActionUpdated {
		intDocument3 := model.CRMUser{
			ID:         model.GetMarketoUserId(docType, line, columnNamesFromMetadata),
			ProjectID:  projectId,
			Source:     model.CRM_SOURCE_MARKETO,
			Type:       model.GetMarketoDocumentDocumentType(docType),
			Timestamp:  timestamps[2],
			Properties: values,
			Email:      model.GetMarketoDocumentEmail(docType, line, columnNamesFromMetadata),
			Phone:      model.GetMarketoDocumentPhone(docType, line, columnNamesFromMetadata),
		}
		insertionStatus, errCRMStatus = store.GetStore().CreateCRMUser(&intDocument3)
	}
	return insertionStatus, errCRMStatus, model.GetUniqueLogValue(docType, line, columnNamesFromMetadata)
}
