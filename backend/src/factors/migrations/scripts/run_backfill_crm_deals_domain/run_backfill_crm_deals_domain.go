package main

import (
	"flag"
	"fmt"
	"net/http"
	"sync"

	log "github.com/sirupsen/logrus"

	C "factors/config"
	IntHubspot "factors/integration/hubspot"
	IntSalesforce "factors/integration/salesforce"
	"factors/model/model"
	"factors/model/store"
	"factors/util"
)

func main() {
	projectIds := flag.String("project_ids",
		"", "List or projects ids to add dashboard units")

	env := flag.String("env", C.DEVELOPMENT, "")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")
	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")
	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")
	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")
	crmSource := flag.String("crm_source", "", "crm source to run on.")
	startTime := flag.Int64("start_time", 0, "")
	endTime := flag.Int64("end_time", 0, "")
	workers := flag.Int("workers", 1, "")

	flag.Parse()

	if *env != C.DEVELOPMENT &&
		*env != C.STAGING &&
		*env != C.PRODUCTION {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}

	taskID := "Task#BackfillCRMDealsDomain"
	defer util.NotifyOnPanic(taskID, *env)

	config := &C.Configuration{
		AppName: "backfill_crm_deals_domain",
		Env:     *env,
		MemSQLInfo: C.DBConf{
			Host:     *memSQLHost,
			Port:     *memSQLPort,
			User:     *memSQLUser,
			Name:     *memSQLName,
			Password: *memSQLPass,
		},
		RedisHost:           *redisHost,
		RedisPort:           *redisPort,
		RedisHostPersistent: *redisHostPersistent,
		RedisPortPersistent: *redisPortPersistent,
		SentryDSN:           *sentryDSN,
		PrimaryDatastore:    *primaryDatastore,
	}

	C.InitConf(config)
	C.InitSentryLogging(config.SentryDSN, config.AppName)

	err := C.InitDB(*config)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize db.")
	}

	_, projectIDMap, _ := C.GetProjectsFromListWithAllProjectSupport(*projectIds, "")

	logCtx := log.WithFields(log.Fields{"project_ids": projectIDMap, "start_time": startTime, "end_time": *endTime,
		"crm_source": *crmSource, "workers": *workers})
	if *crmSource == "hubspot" {
		logCtx.Info("Starting hubspot deals domain backfilling.")
		for projectID := range projectIDMap {
			startHubspotDealDomainAssociation(projectID, *startTime, *endTime, *workers)
		}
	}

	if *crmSource == "salesforce" {
		logCtx.Info("Starting salesforce opportunities domain backfilling.")
		for projectID := range projectIDMap {
			startSalesforceOpportunityDomainAssociation(projectID, *startTime, *endTime, *workers)
		}
	}

	logCtx.Info("Backfilling job completed.")
}

func startHubspotDealDomainAssociation(projectID int64, startTime, endTime int64, workers int) int {
	documents, status := store.GetStore().GetHubspotDocumentsByTypeAndAction(projectID, model.HubspotDocumentTypeDeal,
		model.HubspotDocumentActionCreated, startTime*1000, endTime*1000)
	if status != http.StatusFound {
		if status != http.StatusNotFound {
			log.WithFields(log.Fields{"project_id": projectID, "start_time": startTime, "end_time": endTime}).
				Error("Failed to get deal documents.")
		}

		return status
	}

	batchedDocuments := IntHubspot.GetBatchedOrderedDocumentsByID(documents, workers)
	for _, documents := range batchedDocuments {
		var wg sync.WaitGroup
		for id := range documents {
			wg.Add(1)
			go func(docID string) {
				defer wg.Done()
				associateDealToDomain(projectID, &documents[docID][0])
			}(id)
		}
		wg.Wait()
	}

	return http.StatusOK
}

func associateDealToDomain(projectID int64, document *model.HubspotDocument) {
	logCtx := log.WithFields(log.Fields{"project_id": projectID,
		"document_id": document.ID})

	if document.Synced == false {
		logCtx.Warning("Deal not synced.")
		return
	}

	groupUserID := document.GroupUserId
	if groupUserID == "" {
		enProperties, _, err := IntHubspot.GetDealProperties(projectID, document)
		if err != nil {
			logCtx.Error("Failed to get hubspot deal document properties")
			return
		}

		dealGroupUserID, _, status := IntHubspot.CreateOrUpdateHubspotGroupsProperties(projectID, document, enProperties,
			model.GROUP_NAME_HUBSPOT_DEAL, document.ID)
		if status != http.StatusOK {
			logCtx.Error("Failed to update deal group properties.")
			return
		}
		groupUserID = dealGroupUserID
	}

	_, companyIDs, err := IntHubspot.GetDealAssociatedIDs(projectID, document)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get deal company ids.")
		return
	}

	if len(companyIDs) == 0 {
		logCtx.Warning("No company id on deal record.")
		return
	}

	status := IntHubspot.AssociateDealToDomain(projectID, groupUserID, companyIDs[0])
	if status != http.StatusOK {
		logCtx.Error("Failed to AssociateDealToDomain.")
	}
}

func startSalesforceOpportunityDomainAssociation(projectID int64, startTime, endTime int64, workers int) int {
	documents, status := store.GetStore().GetSalesforceDocumentsByTypeAndAction(projectID, model.SalesforceDocumentTypeOpportunity,
		model.SalesforceDocumentCreated, startTime, endTime)
	if status != http.StatusFound {
		if status != http.StatusNotFound {
			log.WithFields(log.Fields{"project_id": projectID, "start_time": startTime, "end_time": endTime}).
				Error("Failed to get opportunity documents.")
		}

		return status
	}

	batchedDocuments := IntSalesforce.GetBatchedOrderedDocumentsByID(documents, workers)

	for _, documents := range batchedDocuments {
		var wg sync.WaitGroup
		for id := range documents {
			wg.Add(1)
			go func(docID string) {
				defer wg.Done()
				associateOpportunityToDomain(projectID, &documents[docID][0])

			}(id)
		}
		wg.Wait()
	}

	return http.StatusOK
}

func associateOpportunityToDomain(projectID int64, document *model.SalesforceDocument) {
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "document_id": document.ID})

	if document.Synced == false {
		logCtx.Warning("Opportunity not synced.")
		return
	}

	groupUserID := document.GroupUserID
	if groupUserID == "" {
		opportunityGroupUserID, status, _ := IntSalesforce.CreateOrUpdateSalesforceGroupsProperties(projectID, document,
			model.GROUP_NAME_SALESFORCE_OPPORTUNITY, document.ID)
		if status != http.StatusOK {
			logCtx.Error("Failed to create or update salesforce opportunity groups properties.")
			return
		}
		groupUserID = opportunityGroupUserID
	}

	enProperties, _, err := IntSalesforce.GetSalesforceDocumentProperties(projectID, document)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get opportunity properties.")
		return
	}

	accountID := util.GetPropertyValueAsString((*enProperties)[model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceSalesforce,
		model.GetSalesforceAliasByDocType(model.SalesforceDocumentTypeOpportunity), "accountid")])

	if accountID == "" {
		logCtx.Warning("No account id on opportunity record.")
		return
	}

	status := IntSalesforce.AssociateOpportunityToDomains(projectID, groupUserID, accountID)
	if status != http.StatusOK {
		logCtx.Error("Failed to AssociateOpportunityToDomains.")
	}
}
