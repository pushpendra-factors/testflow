package task

import (
	"context"
	"factors/model/model"
	"factors/model/store"
	BQ "factors/services/bigquery"
	"fmt"
	"strconv"

	U "factors/util"

	log "github.com/sirupsen/logrus"
	"strings"
)

func BingAdsIntegration(projectId int64, configs map[string]interface{}) (map[string]interface{}, bool) {
	available, err := store.GetStore().GetFeatureStatusForProjectV2(projectId, model.FEATURE_BING_ADS)
	if err != nil {
		log.WithError(err).Error("Failed to get feature status in bing ads integration job for project ID ", projectId)
	}
	if !available {
		log.Error("Feature Not Available... Skipping bing ads integration job for project ID ", projectId)
		return nil, false
	}
	bigQuerySetting := model.BigquerySetting{
		BigqueryProjectID:       configs["BigqueryProjectId"].(string),
		BigqueryCredentialsJSON: configs["BigqueryCredential"].(string),
	}

	resultStatus := make(map[string]interface{})
	executionDate := configs["startTimestamp"].(int64) - 86400
	executionDateString := U.GetDateOnlyHyphenFormatFromTimestampZ(executionDate)
	executionDateStringYYYYMMDD, _ := strconv.ParseInt(U.GetDateOnlyFromTimestampZ(executionDate), 10, 64)
	ctx := context.Background()
	status := true
	client, err := BQ.CreateBigqueryClient(&ctx, &bigQuerySetting)
	if err != nil {
		status = false
		resultStatus["failure"] = "Failed to get bigquery client" + err.Error()
		log.WithError(err).Error("Failed to get bigquery client")
	}
	defer client.Close()

	mapping, err := store.GetStore().GetActiveFiveTranMapping(projectId, model.BingAdsIntegration)
	if err != nil {
		status = false
		resultStatus["failure"] = "Error executing fivetran mapping query" + err.Error()
		log.WithError(err).Error("Error executing fivetran mapping query")
	}

	totalSuccess := 0
	totalFailures := 0
	PAGE_SIZE := 1000
	for docType, baseQuery := range model.BingAdsDocumentToQuery {
		offset := 0
		for {
			success := 0
			failures := 0
			query := model.GetBingAdsDocumentQuery(configs["BigqueryProjectId"].(string), mapping.SchemaID, baseQuery, executionDateString, docType, PAGE_SIZE, offset)
			var queryResult [][]string
			err = BQ.ExecuteQuery(&ctx, client, query, &queryResult)
			if err != nil {
				resultStatus["failure-"+docType] = "Error while executing query" + err.Error()
				status = false
				log.WithError(err).Error("Error while executing query")
			}
			for _, line := range queryResult {
				values := model.GetBingAdsDocumentValues(docType, line)
				valuesBlob, err := U.EncodeToPostgresJsonb(&values)
				if err != nil {
					log.WithError(err).Error("Error while encoding to jsonb")
					failures++
					break
				}
				intDocument := model.IntegrationDocument{
					DocumentId:        model.GetBingAdsDocumentDocumentId(model.BingadsDocumentTypeAlias[docType], line),
					ProjectID:         projectId,
					CustomerAccountID: model.GetBingAdsDocumentAccountId(model.BingadsDocumentTypeAlias[docType], line),
					Source:            model.BingAdsIntegration,
					DocumentType:      model.GetBingAdsDocumentDocumentType(docType),
					Timestamp:         executionDateStringYYYYMMDD,
					Value:             valuesBlob,
				}
				if docType == "campaigns" || docType == "ad_groups" || docType == "keyword" || docType == "account" {
					err := store.GetStore().UpsertIntegrationDocument(intDocument)
					if err != nil {
						log.WithError(err).Error("Error upserting integration document")
						failures++
					} else {
						success++
					}
				} else {
					err := store.GetStore().InsertIntegrationDocument(intDocument)
					if err != nil {
						log.WithError(err).Error("Error inserting integration document")
						failures++
					} else {
						success++
					}
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
	var accounts [][]string
	accountsQuery := model.GetAllAccountsQuery(configs["BigqueryProjectId"].(string), mapping.SchemaID)
	err = BQ.ExecuteQuery(&ctx, client, accountsQuery, &accounts)
	accountString := ""
	for _, account := range accounts {
		if accountString == "" {
			accountString = account[0]
		} else {
			accountString = fmt.Sprintf("%v,%v", accountString, account)
		}
	}
	accountString = strings.Replace(accountString, "[", "", -1)
	accountString = strings.Replace(accountString, "]", "", -1)
	store.GetStore().UpdateFiveTranMappingAccount(mapping.ProjectID, mapping.Integration, mapping.ConnectorID, accountString)
	if totalFailures > 0 || status == false {
		return resultStatus, false
	}
	return resultStatus, true
}
