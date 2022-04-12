package task

import (
	"context"
	"factors/model/model"
	"factors/model/store"
	BQ "factors/services/bigquery"
	"net/http"

	U "factors/util"

	log "github.com/sirupsen/logrus"
)

func MarketoIntegration(projectId uint64, configs map[string]interface{}) (map[string]interface{}, bool) {

	bigQuerySetting := model.BigquerySetting{
		BigqueryProjectID:       configs["BigqueryProjectId"].(string),
		BigqueryCredentialsJSON: configs["BigqueryCredential"].(string),
	}

	resultStatus := make(map[string]interface{})
	executionDate := configs["startTimestamp"].(int64) - 86400
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
	for docType, baseQuery := range model.MarketoDocumentToQuery {
		offset := 0
		for {
			success := 0
			failures := 0
			query := model.GetMarketoDocumentQuery(configs["BigqueryProjectId"].(string), mapping.SchemaID, baseQuery, executionDateString, docType, PAGE_SIZE, offset)
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
			}
			for _, line := range queryResult {
				values := model.GetMarketoDocumentValues(docType, line, columnNamesFromMetadata, columnNamesFromMetadataDateTime)
				valuesBlob, err := U.EncodeToPostgresJsonb(&values)
				if err != nil {
					log.WithError(err).Error("Error while encoding to jsonb")
					failures++
					break
				}
				insertionStatus := int(0)
				if model.DocTypeIntegrationObjectMap[docType] == "activity" {
					timestamps := model.GetMarketoDocumentTimestamp(docType, line, columnNamesFromMetadata)
					intDocument := model.CRMActivity{
						ProjectID:  projectId,
						Source:     model.CRM_SOURCE_MARKETO,
						Name:       "program_membership_created",
						Type:       model.GetMarketoDocumentDocumentType(docType),
						ActorType:  model.GetMarketoActorType(docType),
						ActorID:    model.GetMarketoDocumentActorId(docType, line, columnNamesFromMetadata),
						Timestamp:  timestamps[0],
						Properties: valuesBlob,
					}

					insertionStatus, err = store.GetStore().CreateCRMActivity(&intDocument)
					if insertionStatus == http.StatusConflict {
						intDocument.Name = "program_membership_updated"
						intDocument.Timestamp = timestamps[0]
						insertionStatus, err = store.GetStore().CreateCRMActivity(&intDocument)
					}
				}
				if model.DocTypeIntegrationObjectMap[docType] == "user" {
					timestamps := model.GetMarketoDocumentTimestamp(docType, line, columnNamesFromMetadata)
					intDocument := model.CRMUser{
						ID:         model.GetMarketoUserId(docType, line, columnNamesFromMetadata),
						ProjectID:  projectId,
						Source:     model.CRM_SOURCE_MARKETO,
						Type:       model.GetMarketoDocumentDocumentType(docType),
						Timestamp:  timestamps[0],
						Properties: valuesBlob,
						Email:      model.GetMarketoDocumentEmail(docType, line, columnNamesFromMetadata),
						Phone:      model.GetMarketoDocumentPhone(docType, line, columnNamesFromMetadata),
					}
					insertionStatus, err = store.GetStore().CreateCRMUser(&intDocument)
					if insertionStatus == http.StatusCreated {
						intDocument.Timestamp = timestamps[1]
						insertionStatus, err = store.GetStore().CreateCRMUser(&intDocument)
					}
					intDocument.Timestamp = timestamps[2]
					insertionStatus, err = store.GetStore().CreateCRMUser(&intDocument)

				}
				if err != nil || insertionStatus != http.StatusCreated {
					log.WithError(err).Error("Error upserting integration document")
					failures++
				} else {
					success++
				}

			}

			resultStatus["failure-"+docType] = failures
			resultStatus["success-"+docType] = success
			totalFailures = totalFailures + failures
			totalSuccess = totalSuccess + success
			if len(queryResult) < PAGE_SIZE {
				break
			}
			offset = offset + PAGE_SIZE
		}
	}
	if totalFailures > 0 || status == false {
		return resultStatus, false
	}
	return resultStatus, true
}
