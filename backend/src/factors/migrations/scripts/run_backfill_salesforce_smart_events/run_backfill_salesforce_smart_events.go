package main

import (
	C "factors/config"
	IntSalesforce "factors/integration/salesforce"
	"factors/model/model"
	"factors/model/store"
	"factors/util"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

func main() {
	env := flag.String("env", C.DEVELOPMENT, "")

	dbHost := flag.String("db_host", C.PostgresDefaultDBParams.Host, "")
	dbPort := flag.Int("db_port", C.PostgresDefaultDBParams.Port, "")
	dbUser := flag.String("db_user", C.PostgresDefaultDBParams.User, "")
	dbName := flag.String("db_name", C.PostgresDefaultDBParams.Name, "")
	dbPass := flag.String("db_pass", C.PostgresDefaultDBParams.Password, "")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")

	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")
	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")
	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")

	projectID := flag.Int64("project_id", 0, "Project Id.")
	eventNameIDs := flag.String("event_name_ids", "", "Event name Ids.")
	startTime := flag.Int64("start_timestamp", 0, "Staring timestamp from events search.")
	endTime := flag.Int64("end_timestamp", 0, "Ending timestamp from events search. End timestamp will be excluded")
	wetRun := flag.Bool("wet", false, "Wet run")
	propertiesTypeCacheSize := flag.Int("property_details_cache_size", 0, "Cache size for in memory property detail.")
	enablePropertyTypeFromDB := flag.Bool("enable_property_type_from_db", false, "Enable property type check from db.")
	whitelistedProjectIDPropertyTypeFromDB := flag.String("whitelisted_project_ids_property_type_check_from_db", "", "Allowed project id for property type check from db.")
	blacklistedProjectIDPropertyTypeFromDB := flag.String("blacklisted_project_ids_property_type_check_from_db", "", "Blocked project id for property type check from db.")
	flag.Parse()
	defer util.NotifyOnPanic("Task#run_backfill_salesforce_smart_events", *env)

	taskID := "run_backfill_salesforce_smart_events"
	if *projectID == 0 {
		log.Error("project_id not provided")
		os.Exit(1)
	}

	if *eventNameIDs == "" {
		log.Panic("Invalid event name ids")
	}

	if *startTime <= 0 || *endTime <= 0 {
		log.Panic("Invalid range.")
	}

	config := &C.Configuration{
		AppName: taskID,
		Env:     *env,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
			AppName:  taskID,
		},
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
			Certificate: *memSQLCertificate,
			AppName:     taskID,
		},
		SentryDSN:           *sentryDSN,
		RedisHost:           *redisHost,
		RedisPort:           *redisPort,
		RedisHostPersistent: *redisHostPersistent,
		RedisPortPersistent: *redisPortPersistent,
		PrimaryDatastore:    *primaryDatastore,
		DryRunCRMSmartEvent: !*wetRun,
	}

	C.InitConf(config)
	C.InitSentryLogging(config.SentryDSN, config.AppName)
	C.InitPropertiesTypeCache(*enablePropertyTypeFromDB, *propertiesTypeCacheSize,
		*whitelistedProjectIDPropertyTypeFromDB, *blacklistedProjectIDPropertyTypeFromDB)
	C.InitRedis(config.RedisHost, config.RedisPort)
	C.InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)

	err := C.InitDB(*config)
	if err != nil {
		log.Error("Failed to initialize DB.")
		os.Exit(1)
	}

	if !*wetRun {
		log.Info("Running in dry run")
	} else {
		log.Info("Running in wet run")
	}

	log.Info(fmt.Sprintf("Running for event_name_id %v", *eventNameIDs))

	eventNameIDsArr := strings.Split(*eventNameIDs, ",")
	salesforceSmartEventNames := IntSalesforce.GetSalesforceSmartEventNames(*projectID)

	filteredEventNames := filterSmartEventNamesByEventNameIDs((*salesforceSmartEventNames), eventNameIDsArr)

	log.WithFields(log.Fields{"project_id": *projectID, "salesforce_smart_event_names": filteredEventNames}).
		Info("Processing for following smart events.")

	for typeAlias, smartEvents := range filteredEventNames {
		documents, status := getSalesforceDocumentsBetweenTimeRange(*projectID, typeAlias, *startTime, *endTime)
		if status != http.StatusFound {
			log.WithFields(log.Fields{"project_id": projectID, "doc_type": typeAlias, "from": *startTime, "to": *endTime}).
				Error("Failed to get getSalesforceDocumentsBetweenTimeRange")
			return
		}

		// generate single batch for all ids in order
		batchedDocuments := IntSalesforce.GetBatchedOrderedDocumentsByID(documents, len(documents))[0]
		for docID := range batchedDocuments {
			backfillSalesforceSmartEventForSameID(*projectID, batchedDocuments[docID], smartEvents)
		}
	}
}

func filterSmartEventNamesByEventNameIDs(salesforceSmartEventNames map[string][]IntSalesforce.SalesforceSmartEventName,
	requiredEventNameIDs []string) map[string][]IntSalesforce.SalesforceSmartEventName {

	eventNamesMap := make(map[string]bool)
	for i := range requiredEventNameIDs {
		eventNamesMap[requiredEventNameIDs[i]] = true
	}

	filteredSalesforceEventNames := make(map[string][]IntSalesforce.SalesforceSmartEventName, 0)
	for docType := range salesforceSmartEventNames {

		for i := range salesforceSmartEventNames[docType] {
			eventNameID := salesforceSmartEventNames[docType][i].EventNameID
			if _, exist := eventNamesMap[eventNameID]; !exist {
				continue
			}

			filteredSalesforceEventNames[docType] = append(filteredSalesforceEventNames[docType],
				salesforceSmartEventNames[docType][i])
		}
	}

	return filteredSalesforceEventNames
}

func getSalesforceDocumentsBetweenTimeRange(projectID int64, docTypeAlias string, from, to int64) ([]model.SalesforceDocument, int) {

	docType := model.GetSalesforceDocTypeByAlias(docTypeAlias)
	documents, status := store.GetStore().GetSalesforceDocumentByType(projectID, docType, from, to)
	if status != http.StatusFound {
		log.WithFields(log.Fields{"project_id": projectID, "type_alias": docTypeAlias, "from": from, "to": to}).
			Error("Failed to salesforce documents.")
		return nil, status
	}

	return documents, status
}

func backfillSalesforceSmartEventForSameID(projectID int64, documents []model.SalesforceDocument, filteredEventNames []IntSalesforce.SalesforceSmartEventName) {

	logCtx := log.WithFields(log.Fields{"project_id": projectID})

	// set it empty to avoid smart event framework to pull last synced record
	previousDocumentProperties := make(map[string]interface{})
	for i := range documents {

		logDocCtx := logCtx.WithFields(log.Fields{"document_id": documents[i].ID, "timestamp": documents[i].Timestamp})

		if documents[i].Synced == false || documents[i].UserID == "" || documents[i].SyncID == "" {
			logDocCtx.Error("Not synced document")
			continue
		}

		// if the first document is not begenning for the id, then get the previous record to avoid smart event interface to pull last synced record
		if i == 0 && documents[i].Action != model.SalesforceDocumentCreated {
			prevDocuments, status := store.GetStore().GetLatestSalesforceDocumentByID(projectID, []string{documents[i].ID},
				documents[i].Type, documents[i].Timestamp-1)
			if status != http.StatusFound {
				logDocCtx.Error("Failed to get previous document.")
				return
			}

			_, properties, err := IntSalesforce.GetSalesforceDocumentProperties(projectID, &prevDocuments[len(prevDocuments)-1])
			if err != nil {
				logDocCtx.Error("Failed to get previous document properties.")
				return
			}

			previousDocumentProperties = *properties
		}

		_, currentDocumentProperties, err := IntSalesforce.GetSalesforceDocumentProperties(projectID, &documents[i])
		if err != nil {
			logDocCtx.WithError(err).Error("Failed to get document properties.")
			previousDocumentProperties = *currentDocumentProperties
			continue
		}

		// ALways us lastmodified timestamp for updated properties, error handling already done during event creation
		lastModifiedTimestamp, err := model.GetSalesforceDocumentTimestampByAction(&documents[i], model.SalesforceDocumentUpdated)
		if err != nil {
			logDocCtx.WithError(err).Error("Failed to get document last modified timstamp.")
			continue
		}

		for j := range filteredEventNames {
			IntSalesforce.TrackSalesforceSmartEvent(projectID, &filteredEventNames[j], documents[i].SyncID, documents[i].ID, documents[i].UserID,
				documents[i].Type, currentDocumentProperties, &previousDocumentProperties, lastModifiedTimestamp, true)
		}

		previousDocumentProperties = *currentDocumentProperties
	}

}
