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

func LeadSquaredIntegration(projectId int64, configs map[string]interface{}) (map[string]interface{}, bool) {

	bigQuerySetting := model.BigquerySetting{
		BigqueryProjectID:       configs["BigqueryProjectId"].(string),
		BigqueryCredentialsJSON: configs["BigqueryCredential"].(string),
	}

	resultStatus := make(map[string]interface{})
	executionDate := configs["startTimestamp"].(int64)
	executionDateString := U.GetDateOnlyHyphenFormatFromTimestampZ(executionDate)
	ctx := context.Background()
	status := true
	client, err := BQ.CreateBigqueryClient(&ctx, &bigQuerySetting)
	if err != nil {
		status = false
		resultStatus["failure"] = "Failed to get bigquery client" + err.Error()
		log.WithError(err).Error("Failed to get bigquery client")
	}
	defer client.Close()

	setting, _ := store.GetStore().GetProjectSetting(projectId)
	var leadSquaredConfig model.LeadSquaredConfig
	err = U.DecodePostgresJsonbToStructType(setting.LeadSquaredConfig, &leadSquaredConfig)
	if err != nil {
		resultStatus["failure"] = "Failed to decode leadsquared config" + err.Error()
		log.WithError(err).Error("Failed to decode leadsquared config")
		return resultStatus, false
	}

	PAGE_SIZE := 1000
	for docType, _ := range model.LeadSquaredDocumentEndpoint {
		totalSuccess := 0
		totalFailures := 0
		propertySuccess := 0
		propertyFailures := 0
		metadataQuery, exists := model.GetLeadSquaredDocumentMetadataQuery(
			docType,
			configs["BigqueryProjectId"].(string),
			leadSquaredConfig.BigqueryDataset,
			model.LeadSquaredDataObjectColumnsQuery[docType])
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
			columnNamesFromMetadata, columnNamesFromMetadataDateTime, columnNamesFromMetadataNumerical = extractMetadataColumnsLeadSquared(metadataQueryResult)

		}
		propertySuccess, propertyFailures = InsertPropertyDataTypesLeadSquared(columnNamesFromMetadataDateTime, columnNamesFromMetadataNumerical, docType, projectId)
		leadSquaredUrlParams := map[string]string{
			"accessKey": leadSquaredConfig.AccessKey,
			"secretKey": leadSquaredConfig.SecretKey,
		}
		propertyMetadataList, errorStatus, msg := getMetadataDetails(docType, leadSquaredConfig.Host, leadSquaredUrlParams)
		if errorStatus != false {
			resultStatus["error"] = msg
			log.Error(msg)
			return resultStatus, false
		}
		disNameSuccess, disNameFailures := UpdateDisplayNames(projectId, docType, propertyMetadataList)
		resultStatus["displayname-failure-"+docType] = disNameFailures
		resultStatus["displayname-success-"+docType] = disNameSuccess
		LatestProspectAutoId := 0
		for {
			var query string
			query = model.GetLeadSquaredDocumentQuery(configs["BigqueryProjectId"].(string), leadSquaredConfig.BigqueryDataset, model.LeadSquaredDocumentToQuery[docType], executionDateString, docType, PAGE_SIZE, LatestProspectAutoId)
			var queryResult [][]string
			err = BQ.ExecuteQuery(&ctx, client, query, &queryResult)
			if err != nil {
				resultStatus["failure-"+docType] = "Error while executing query" + err.Error()
				status = false
				log.WithError(err).Error("Error while executing query")
			}
			success, failures := InsertIntegrationDocumentLeadSquared(projectId, docType, queryResult, columnNamesFromMetadata, columnNamesFromMetadataDateTime, columnNamesFromMetadataNumerical)
			totalFailures = totalFailures + failures
			totalSuccess = totalSuccess + success
			if len(queryResult) < PAGE_SIZE {
				break
			}
			LatestProspectAutoId = int(model.ConvertToNumber(model.GetLeadSquaredUserAutoId(docType, queryResult[len(queryResult)-1], columnNamesFromMetadata)))
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

func UpdateDisplayNames(projectId int64, docType string, metadata []PropertyMetadataObjectLeadSquared) (int, int) {
	success, failures := int(0), int(0)
	for _, column := range metadata {
		_, err := store.GetStore().CreateCRMProperties(&model.CRMProperty{
			ProjectID: projectId,
			Source:    U.CRM_SOURCE_LEADSQUARED,
			Type:      model.LeadSquaredDocumentTypeAlias[docType],
			Name:      column.SchemaName,
			Label:     column.DisplayName,
		})
		if err != nil {
			failures++
		} else {
			success++
		}
	}
	return success, failures
}

func InsertPropertyDataTypesLeadSquared(columnNamesFromMetadataDateTime map[string]bool, columnNamesFromMetadataNumerical map[string]bool, docType string, projectId int64) (int, int) {
	success, failures := int(0), int(0)
	for columnName, _ := range columnNamesFromMetadataDateTime {
		_, err := store.GetStore().CreateCRMProperties(&model.CRMProperty{
			ProjectID:        projectId,
			Source:           U.CRM_SOURCE_LEADSQUARED,
			Type:             model.LeadSquaredDocumentTypeAlias[docType],
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
			Source:           U.CRM_SOURCE_LEADSQUARED,
			Type:             model.LeadSquaredDocumentTypeAlias[docType],
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

func extractMetadataColumnsLeadSquared(metadataQueryResult [][]string) ([]string, map[string]bool, map[string]bool) {

	columnNamesFromMetadata := make([]string, 0)
	columnNamesFromMetadataDateTime := make(map[string]bool)
	columnNamesFromMetadataNumerical := make(map[string]bool)
	for _, metadataLine := range metadataQueryResult {
		metadataIndex := model.GetMetadataColumnNameIndex("column_name")
		columnName := metadataLine[metadataIndex]
		columnNamesFromMetadata = append(columnNamesFromMetadata, columnName)
		metadataIndexForDataType := model.GetMetadataColumnNameIndex("data_type")
		dataType := metadataLine[metadataIndexForDataType]
		if dataType == "TIMESTAMP" || dataType == "DATE" || dataType == "DATETIME" {
			columnNamesFromMetadataDateTime[columnName] = true
		}
		if dataType == "INT64" || dataType == "FLOAT64" {
			columnNamesFromMetadataNumerical[columnName] = true
		}
	}
	return columnNamesFromMetadata, columnNamesFromMetadataDateTime, columnNamesFromMetadataNumerical
}

func InsertIntegrationDocumentLeadSquared(projectId int64, docType string, queryResult [][]string, columnNamesFromMetadata []string, columnNamesFromMetadataDateTime map[string]bool, columnNamesFromMetadataNumerical map[string]bool) (int, int) {
	success := 0
	failures := 0
	for _, line := range queryResult {
		values := model.GetLeadSquaredDocumentValues(docType, line, columnNamesFromMetadata, columnNamesFromMetadataDateTime, columnNamesFromMetadataNumerical)
		valuesBlob, err := U.EncodeToPostgresJsonb(&values)
		if err != nil {
			log.WithError(err).Error("Error while encoding to jsonb")
			failures++
			break
		}
		insertionStatus := int(0)
		var errCRMStatus error
		var logIndex string
		if model.DocTypeIntegrationObjectMap[docType] == "user" {
			insertionStatus, errCRMStatus, logIndex = insertCRMUserLeadSquared(projectId, line, docType, columnNamesFromMetadata, valuesBlob)
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

func insertCRMUserLeadSquared(projectId int64, line []string, docType string, columnNamesFromMetadata []string, values *postgres.Jsonb) (int, error, string) {
	insertionStatus := int(0)
	var errCRMStatus error
	timestamps := model.GetLeadSquaredDocumentTimestamp(docType, line, columnNamesFromMetadata)
	intDocument1 := model.CRMUser{
		ID:         model.GetLeadSquaredUserId(docType, line, columnNamesFromMetadata),
		ProjectID:  projectId,
		Source:     U.CRM_SOURCE_LEADSQUARED,
		Type:       model.GetLeadSquaredDocumentDocumentType(docType),
		Timestamp:  timestamps[0],
		Properties: values,
		Email:      model.GetLeadSquaredDocumentEmail(docType, line, columnNamesFromMetadata),
		Phone:      model.GetLeadSquaredDocumentPhone(docType, line, columnNamesFromMetadata),
	}
	insertionStatus, errCRMStatus = store.GetStore().CreateCRMUser(&intDocument1)
	if insertionStatus == http.StatusCreated {
		intDocument2 := model.CRMUser{
			ID:         model.GetLeadSquaredUserId(docType, line, columnNamesFromMetadata),
			ProjectID:  projectId,
			Source:     U.CRM_SOURCE_LEADSQUARED,
			Type:       model.GetLeadSquaredDocumentDocumentType(docType),
			Timestamp:  timestamps[1],
			Properties: values,
			Email:      model.GetLeadSquaredDocumentEmail(docType, line, columnNamesFromMetadata),
			Phone:      model.GetLeadSquaredDocumentPhone(docType, line, columnNamesFromMetadata),
		}
		insertionStatus, errCRMStatus = store.GetStore().CreateCRMUser(&intDocument2)
	}
	if model.GetLeadSquaredDocumentAction(docType, line, columnNamesFromMetadata) == model.CRMActionUpdated {
		intDocument3 := model.CRMUser{
			ID:         model.GetLeadSquaredUserId(docType, line, columnNamesFromMetadata),
			ProjectID:  projectId,
			Source:     U.CRM_SOURCE_LEADSQUARED,
			Type:       model.GetLeadSquaredDocumentDocumentType(docType),
			Timestamp:  timestamps[2],
			Properties: values,
			Email:      model.GetLeadSquaredDocumentEmail(docType, line, columnNamesFromMetadata),
			Phone:      model.GetLeadSquaredDocumentPhone(docType, line, columnNamesFromMetadata),
		}
		insertionStatus, errCRMStatus = store.GetStore().CreateCRMUser(&intDocument3)
	}
	return insertionStatus, errCRMStatus, ""
}
