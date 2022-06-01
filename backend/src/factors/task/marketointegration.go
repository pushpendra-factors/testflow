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

	PAGE_SIZE := 1000
	for docType, _ := range model.MarketoDocumentTypeAlias {
		totalSuccess := 0
		totalFailures := 0
		propertySuccess := 0
		propertyFailures := 0
		offset := 0
		lastProcessedId := 0
		metadataQuery, exists := model.GetMarketoDocumentMetadataQuery(docType, configs["BigqueryProjectId"].(string), mapping.SchemaID, model.MarketoDataObjectColumnsQuery[docType])
		var metadataQueryResult [][]string
		columnNamesFromMetadata := make([]string, 0)
		columnNamesFromMetadataDateTime := make(map[string]bool)
		columnNamesFromMetadataNumerical := make(map[string]bool)
		if exists {
			err = BQ.ExecuteQuery(&ctx, client, metadataQuery, &metadataQueryResult)
			if err != nil {
				resultStatus["failure-"+docType] = "Error while executing metadata query" + err.Error()
				status = false
				log.WithError(err).Error("Error while executing query")
			}
			columnNamesFromMetadata, columnNamesFromMetadataDateTime, columnNamesFromMetadataNumerical = extractMetadataColumns(metadataQueryResult)

		}
		propertySuccess, propertyFailures = InsertPropertyDataTypes(columnNamesFromMetadataDateTime, columnNamesFromMetadataNumerical, docType, projectId)
		for {
			var query string
			if docType == model.MARKETO_TYPE_NAME_LEAD {
				tableRef := client.Dataset(mapping.SchemaID).Table("lead_segment")
				_, existsErr := tableRef.Metadata(ctx)
				if existsErr == nil {
					query = model.GetMarketoDocumentQuery(configs["BigqueryProjectId"].(string), mapping.SchemaID, model.MarketoDocumentToQuery[docType], executionDateString, docType, PAGE_SIZE, offset, 0)
				} else {
					query = model.GetMarketoDocumentQuery(configs["BigqueryProjectId"].(string), mapping.SchemaID, model.MarketoDocumentToQuery[model.MARKETO_TYPE_NAME_LEAD_NO_SEGMENT], executionDateString, model.MARKETO_TYPE_NAME_LEAD_NO_SEGMENT, PAGE_SIZE, offset, lastProcessedId)
				}

			} else {
				query = model.GetMarketoDocumentQuery(configs["BigqueryProjectId"].(string), mapping.SchemaID, model.MarketoDocumentToQuery[docType], executionDateString, docType, PAGE_SIZE, offset, 0)
			}
			var queryResult [][]string
			err = BQ.ExecuteQuery(&ctx, client, query, &queryResult)
			if err != nil {
				resultStatus["failure-"+docType] = "Error while executing query" + err.Error()
				status = false
				log.WithError(err).Error("Error while executing query")
			}
			success, failures := InsertIntegrationDocument(projectId, docType, queryResult, columnNamesFromMetadata, columnNamesFromMetadataDateTime, columnNamesFromMetadataNumerical)
			totalFailures = totalFailures + failures
			totalSuccess = totalSuccess + success
			if len(queryResult) < PAGE_SIZE {
				break
			}
			offset = offset + PAGE_SIZE
			lastProcessedId = int(model.ConvertToNumber(model.GetMarketoUserId(docType, queryResult[len(queryResult)-1], columnNamesFromMetadata)))
		}
		resultStatus["failure-"+docType] = totalFailures
		resultStatus["success-"+docType] = totalSuccess
		resultStatus["property-failure-"+docType] = propertyFailures
		resultStatus["property-success-"+docType] = propertySuccess
	}
	if status == false {
		return resultStatus, false
	}
	return resultStatus, true
}

func InsertPropertyDataTypes(columnNamesFromMetadataDateTime map[string]bool, columnNamesFromMetadataNumerical map[string]bool, docType string, projectId uint64) (int, int) {
	success, failures := int(0), int(0)
	for columnName, _ := range columnNamesFromMetadataDateTime {
		_, err := store.GetStore().CreateCRMProperties(&model.CRMProperty{
			ProjectID:        projectId,
			Source:           U.CRM_SOURCE_MARKETO,
			Type:             model.MarketoDocumentTypeAlias[docType],
			Name:             columnName,
			ExternalDataType: "timestamp",
			MappedDataType:   U.PropertyTypeDateTime,
		})
		if err != nil {
			failures++
		} else {
			success++
		}
	}
	for columnName, _ := range model.MarketoDataObjectColumnsDatetimeType[docType] {
		_, err := store.GetStore().CreateCRMProperties(&model.CRMProperty{
			ProjectID:        projectId,
			Source:           U.CRM_SOURCE_MARKETO,
			Type:             model.MarketoDocumentTypeAlias[docType],
			Name:             columnName,
			ExternalDataType: "timestamp",
			MappedDataType:   U.PropertyTypeDateTime,
		})
		if err != nil {
			failures++
		} else {
			success++
		}
	}

	for columnName, _ := range columnNamesFromMetadataNumerical {
		_, err := store.GetStore().CreateCRMProperties(&model.CRMProperty{
			ProjectID:        projectId,
			Source:           U.CRM_SOURCE_MARKETO,
			Type:             model.MarketoDocumentTypeAlias[docType],
			Name:             columnName,
			ExternalDataType: "float64",
			MappedDataType:   U.PropertyTypeNumerical,
		})
		if err != nil {
			failures++
		} else {
			success++
		}
	}
	for columnName, _ := range model.MarketoDataObjectColumnsNumericalType[docType] {
		_, err := store.GetStore().CreateCRMProperties(&model.CRMProperty{
			ProjectID:        projectId,
			Source:           U.CRM_SOURCE_MARKETO,
			Type:             model.MarketoDocumentTypeAlias[docType],
			Name:             columnName,
			ExternalDataType: "float64",
			MappedDataType:   U.PropertyTypeNumerical,
		})
		if err != nil {
			failures++
		} else {
			success++
		}
	}
	return success, failures

}

func extractMetadataColumns(metadataQueryResult [][]string) ([]string, map[string]bool, map[string]bool) {

	columnNamesFromMetadata := make([]string, 0)
	columnNamesFromMetadataDateTime := make(map[string]bool)
	columnNamesFromMetadataNumerical := make(map[string]bool)
	for _, metadataLine := range metadataQueryResult {
		metadataIndex := model.GetMetadataColumnNameIndex("column_name")
		columnName := metadataLine[metadataIndex]
		columnNamesFromMetadata = append(columnNamesFromMetadata, columnName)
		metadataIndexForDataType := model.GetMetadataColumnNameIndex("data_type")
		dataType := metadataLine[metadataIndexForDataType]
		if dataType == "TIMESTAMP" || dataType == "DATE" {
			columnNamesFromMetadataDateTime[columnName] = true
		}
		if dataType == "INT64" || dataType == "FLOAT64" {
			columnNamesFromMetadataNumerical[columnName] = true
		}
	}
	return columnNamesFromMetadata, columnNamesFromMetadataDateTime, columnNamesFromMetadataNumerical
}

func InsertIntegrationDocument(projectId uint64, docType string, queryResult [][]string, columnNamesFromMetadata []string, columnNamesFromMetadataDateTime map[string]bool, columnNamesFromMetadataNumerical map[string]bool) (int, int) {
	success := 0
	failures := 0
	for _, line := range queryResult {
		values := model.GetMarketoDocumentValues(docType, line, columnNamesFromMetadata, columnNamesFromMetadataDateTime, columnNamesFromMetadataNumerical)
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
		Source:             U.CRM_SOURCE_MARKETO,
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
		Source:     U.CRM_SOURCE_MARKETO,
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
			Source:     U.CRM_SOURCE_MARKETO,
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
			Source:     U.CRM_SOURCE_MARKETO,
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
