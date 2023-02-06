package main

import (
	"errors"
	C "factors/config"
	"factors/model/model"
	"factors/model/store"
	"factors/util"
	"flag"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
)

func selectSmartEventIDANDReferenceID(projectID int64, eventNameID string, from, to int64) (string, []interface{}) {
	smartEventsStmnt := "select id as smart_event_id, uuid(properties ->> '$crm_reference_event_id') as reference_event_id " +
		"from events where project_id = ? and timestamp between ? and ? and event_name_id = ? "
	smartEventsParams := []interface{}{projectID, from, to, eventNameID}

	return smartEventsStmnt, smartEventsParams
}

func selectSmartEventContactByReferenceID(projectID int64, sourceName, docIDPropertyName string) (string, []interface{}) {
	smartEventsReferenceContact := "select smart_event_id, reference_event_id, properties ->> ? " +
		"as hubspot_contact_id from " + sourceName + " left join events on " + sourceName + ".reference_event_id = events.id and project_id = ?"
	smartEventsReferenceContactParams := []interface{}{docIDPropertyName, projectID}

	return smartEventsReferenceContact, smartEventsReferenceContactParams
}

func selectContactDocumentBySyncIDAndContactID(projectID int64, syncIDSource string, docType int) (string, []interface{}) {
	smartEventDocuments := "select id, timestamp, sync_id, value from hubspot_documents where project_id = ? and type = ? " +
		"and id in ( select hubspot_contact_id from " + syncIDSource + " ) and sync_id in ( select " +
		"reference_event_id :: text from " + syncIDSource + " ) "
	smartEventDocumentsParams := []interface{}{projectID, docType}
	return smartEventDocuments, smartEventDocumentsParams
}

func selectJoinedSmartEventToDocument(smartEventSource, documentSource, propertyName string) (string, []interface{}) {
	stmnt := "select hubspot_contact_id, smart_event_id, reference_event_id, smart_event_documents.sync_id, " +
		"smart_event_documents.timestamp as curr_timestamp, value -> 'properties' -> ? ->> 'value' as curr_value " +
		"from " + smartEventSource + " left join " + documentSource + " on " + smartEventSource + ".reference_event_id :: text = " + documentSource + ".sync_id"
	return stmnt, []interface{}{propertyName}
}

func selectContactUpdatedDocumentsProperty(sourceStep, propertyName string) (string, []interface{}) {
	stmnt := "select distinct on(smart_event_id, hubspot_contact_id) smart_event_id, hubspot_contact_id, " +
		"curr_timestamp, timestamp as prev_timestamp, curr_value, value -> 'properties' -> ? ->> 'value' " +
		" as prev_value from " + sourceStep + " where timestamp < curr_timestamp order by " +
		"smart_event_id, hubspot_contact_id, timestamp desc"

	return stmnt, []interface{}{propertyName}
}

func selectContactCreatedDocumentsProperty(sourceStep, contactUpdatedDocumentSource string) string {
	stmnt := "select distinct on (all_related_documents.smart_event_id) all_related_documents.smart_event_id, all_related_documents.hubspot_contact_id, " +
		"all_related_documents.curr_timestamp, all_related_documents.timestamp as prev_timestamp, all_related_documents.curr_value, " +
		"''::text as prev_value from " + sourceStep + ", " + contactUpdatedDocumentSource + " where " +
		sourceStep + ".smart_event_id not in ( select smart_event_id from " + contactUpdatedDocumentSource + ") " +
		"and " + sourceStep + ".timestamp = " + sourceStep + ".curr_timestamp"
	return stmnt
}

func selectAllHubspotDocumentsByContactID(projectID int64, sourceStep string, docType int) (string, []interface{}) {
	allDocumentsBySmartEventContact := "select * from " + sourceStep + " left join hubspot_documents on " +
		sourceStep + ".hubspot_contact_id = hubspot_documents.id and project_id = ? and type = ? " +
		"where project_id = ? and type = ?"
	allDocumentsBySmartEventContactParams := []interface{}{projectID, docType, projectID, docType}
	return allDocumentsBySmartEventContact, allDocumentsBySmartEventContactParams
}

func selectUnionResult(contactCreatedStmnt, contactUpdatedStmnt string) string {
	return "select * from filtered_created_contact_result union all select * from filtered_updated_contact_result"
}

func selectSmartEventPreviousPropetyValue(sourceStep string) string {
	stmnt := "select smart_event_id, hubspot_contact_id, curr_timestamp, prev_timestamp, curr_value, " +
		" prev_value from " + sourceStep
	return stmnt
}

func getMemSQLQuery(projectID int64, eventNameID string, from, to int64, propertyName string, docType int) (string, []interface{}) {
	stmnt := `WITH smart_events AS (
		select
			id as smart_event_id,
			(JSON_EXTRACT_STRING(properties, ?)) as reference_event_id
		from
			events
		where
			project_id = ?
			and timestamp between ?
			and ?
			and event_name_id = ?
	),
	smart_event_reference_contact AS (
		select
			smart_event_id,
			reference_event_id,
			JSON_EXTRACT_STRING(properties, ?) as hubspot_contact_id
		from
			smart_events
			left join events on smart_events.reference_event_id = events.id
			and project_id = ?
	),
	smart_event_documents AS (
		select
			id,
			timestamp,
			sync_id,
			value
		from
			hubspot_documents
		where
			project_id = ?
			and type = ?
			and id in (
				select
					hubspot_contact_id
				from
					smart_event_reference_contact
			)
			and sync_id in (
				select
					reference_event_id
				from
					smart_event_reference_contact
			)
	),
	smart_events_to_hubspot_documents AS (
		select
			hubspot_contact_id,
			smart_event_id,
			reference_event_id,
			smart_event_documents.sync_id,
			smart_event_documents.timestamp as curr_timestamp,
			JSON_EXTRACT_STRING(value :: properties :: ` + propertyName + `, 'value') as curr_value
		from
			smart_event_reference_contact
			left join smart_event_documents on smart_event_reference_contact.reference_event_id = smart_event_documents.sync_id
	),
	all_related_documents AS (
		select
			hubspot_contact_id,
			smart_event_id,
			reference_event_id,
			smart_events_to_hubspot_documents.sync_id,
			curr_timestamp,
			hubspot_documents.timestamp,
			curr_value,
			hubspot_documents.value
		from
			smart_events_to_hubspot_documents
			left join hubspot_documents on smart_events_to_hubspot_documents.hubspot_contact_id = hubspot_documents.id
			and project_id = ?
			and type = ?
		where
			project_id = ?
			and type = ?
	),
	filtered_updated_contact_result AS (
		select
			smart_event_id,
			hubspot_contact_id,
			LAST(
				curr_timestamp,
				from_unixtime(all_related_documents.timestamp / 1000)
			) as curr_timestamp,
			LAST(
				all_related_documents.timestamp,
				from_unixtime(all_related_documents.timestamp / 1000)
			) as prev_timestamp,
			LAST(
				curr_value,
				from_unixtime(all_related_documents.timestamp / 1000)
			) as curr_value,
			LAST(
				JSON_EXTRACT_STRING(value :: properties :: ` + propertyName + `, 'value'),
				from_unixtime(all_related_documents.timestamp / 1000)
			) as prev_value
		from
			all_related_documents
		where
			timestamp < curr_timestamp
		group by
			smart_event_id,
			hubspot_contact_id
	),
	filtered_created_contact_result AS (
		select
			all_related_documents.smart_event_id,
			all_related_documents.hubspot_contact_id,
			LAST(
				all_related_documents.curr_timestamp,
				from_unixtime(all_related_documents.timestamp / 1000)
			) as curr_timestamp,
			LAST(
				all_related_documents.timestamp,
				from_unixtime(all_related_documents.timestamp / 1000)
			) as prev_timestamp,
			LAST(
				all_related_documents.curr_value,
				from_unixtime(all_related_documents.timestamp / 1000)
			) as curr_value,
			'' as prev_value
		from
			all_related_documents
			left join filtered_updated_contact_result on filtered_updated_contact_result.smart_event_id = all_related_documents.smart_event_id
		where
			filtered_updated_contact_result.smart_event_id is null
			and all_related_documents.timestamp = all_related_documents.curr_timestamp
		group by
			all_related_documents.smart_event_id,
			all_related_documents.hubspot_contact_id
	),
	overall_result AS (
		select
			*
		from
			filtered_created_contact_result
		union
		all
		select
			*
		from
			filtered_updated_contact_result
	)
	select
		smart_event_id,
		hubspot_contact_id,
		curr_timestamp,
		prev_timestamp,
		curr_value,
		prev_value
	from
		overall_result`

	params := []interface{}{"$reference_event_id", projectID, from, to, eventNameID, "$hubspot_contact_hs_object_id",
		projectID, projectID, docType, projectID,
		docType, projectID, docType}

	return stmnt, params
}

/*
WITH smart_events AS (
    select
        id as smart_event_id,
        uuid(properties ->> '$crm_reference_event_id') as reference_event_id
    from
        events
    where
        project_id = '399'
        and timestamp between '1627324200'
        and '1628101800'
        and event_name_id = '2330399'
),
smart_event_reference_contact AS (
    select
        smart_event_id,
        reference_event_id,
        properties ->> '$hubspot_contact_hs_object_id' as hubspot_contact_id
    from
        smart_events
        left join events on smart_events.reference_event_id = events.id
        and project_id = '399'
),
smart_event_documents AS (
    select
        id,
        timestamp,
        sync_id,
        value
    from
        hubspot_documents
    where
        project_id = '399'
        and type = '2'
        and id in (
            select
                hubspot_contact_id
            from
                smart_event_reference_contact
        )
        and sync_id in (
            select
                reference_event_id :: text
            from
                smart_event_reference_contact
        )
),
smart_events_to_hubspot_documents AS (
    select
        hubspot_contact_id,
        smart_event_id,
        reference_event_id,
        smart_event_documents.sync_id,
        smart_event_documents.timestamp as curr_timestamp,
        value -> 'properties' -> 'demo_booked_on' ->> 'value' as curr_value
    from
        smart_event_reference_contact
        left join smart_event_documents on smart_event_reference_contact.reference_event_id :: text = smart_event_documents.sync_id
),
all_related_documents AS (
    select
        *
    from
        smart_events_to_hubspot_documents
        left join hubspot_documents on smart_events_to_hubspot_documents.hubspot_contact_id = hubspot_documents.id
        and project_id = '399'
        and type = '2'
    where
        project_id = '399'
        and type = '2'
),
filtered_updated_contact_result AS (
    select
        distinct on(smart_event_id, hubspot_contact_id) smart_event_id,
        hubspot_contact_id,
        curr_timestamp,
        timestamp as prev_timestamp,
        curr_value,
        value -> 'properties' -> 'demo_booked_on' ->> 'value' as prev_value
    from
        all_related_documents
    where
        timestamp < curr_timestamp
    order by
        smart_event_id,
        hubspot_contact_id,
        timestamp desc
),
filtered_created_contact_result AS (
    select
        distinct on (all_related_documents.smart_event_id) all_related_documents.smart_event_id,
        all_related_documents.hubspot_contact_id,
        all_related_documents.curr_timestamp,
        all_related_documents.timestamp as prev_timestamp,
        all_related_documents.curr_value,
        '' :: text as prev_value
    from
        all_related_documents,
        filtered_updated_contact_result
    where
        all_related_documents.smart_event_id not in (
            select
                smart_event_id
            from
                filtered_updated_contact_result
        )
        and all_related_documents.timestamp = all_related_documents.curr_timestamp
),
overall_result AS (
    select
        *
    from
        filtered_created_contact_result
    union
    all
    select
        *
    from
        filtered_updated_contact_result
)
select
    smart_event_id,
    hubspot_contact_id,
    curr_timestamp,
    prev_timestamp,
    curr_value,
    prev_value
from
    overall_result
*/
func GetSmartEventMetaDataQuery(projectID int64, propertyName, eventNameID string, from, to int64) (string, []interface{}) {
	stmnt := ""
	withParams := []interface{}{}
	stmnt, params := selectSmartEventIDANDReferenceID(projectID, eventNameID, from, to)
	withStmnt := " WITH " + "smart_events" + " AS " + " ( " + stmnt + " ) "
	withParams = append(withParams, params...)

	stmnt, params = selectSmartEventContactByReferenceID(projectID, "smart_events", "$hubspot_contact_hs_object_id")
	withStmnt = withStmnt + " , " + "smart_event_reference_contact" + " AS " + "(" + stmnt + " ) "
	withParams = append(withParams, params...)
	stmnt, params = selectContactDocumentBySyncIDAndContactID(projectID, "smart_event_reference_contact", model.HubspotDocumentTypeContact)
	withStmnt = withStmnt + " , " + "smart_event_documents" + " AS " + " ( " + stmnt + " ) "
	withParams = append(withParams, params...)

	stmnt, params = selectJoinedSmartEventToDocument("smart_event_reference_contact", "smart_event_documents", propertyName)
	withStmnt = withStmnt + " , " + "smart_events_to_hubspot_documents" + " AS " + " ( " + stmnt + " ) "
	withParams = append(withParams, params...)

	stmnt, params = selectAllHubspotDocumentsByContactID(projectID, "smart_events_to_hubspot_documents", model.HubspotDocumentTypeContact)
	withStmnt = withStmnt + " , " + "all_related_documents" + " AS " + " ( " + stmnt + " ) "
	withParams = append(withParams, params...)

	stmnt, params = selectContactUpdatedDocumentsProperty("all_related_documents", propertyName)
	withStmnt = withStmnt + " , " + "filtered_updated_contact_result" + " AS " + " ( " + stmnt + " ) "
	withParams = append(withParams, params...)

	stmnt = selectContactCreatedDocumentsProperty("all_related_documents", "filtered_updated_contact_result")
	withStmnt = withStmnt + " , " + "filtered_created_contact_result" + " AS " + " ( " + stmnt + " ) "

	stmnt = selectUnionResult("filtered_created_contact_result", "filtered_updated_contact_result")
	withStmnt = withStmnt + " , " + "overall_result" + " AS " + " ( " + stmnt + " ) "

	stmnt = selectSmartEventPreviousPropetyValue("overall_result")

	selectStmnt := withStmnt + " " + stmnt
	return selectStmnt, withParams
}

func getSmartEventMetaData(queryStmnt string, queryParams []interface{}) (map[string]map[string]interface{}, error) {
	db := C.GetServices().Db
	rows, err := db.Raw(queryStmnt, queryParams...).Rows()

	if err != nil {
		log.WithError(err).Error("Failed to getSmartEventMetaData.")
		return nil, err
	}
	defer rows.Close()

	smartEventMetaData := map[string]map[string]interface{}{}
	for rows.Next() {
		var smartEventID string
		var hubspotContactID string
		var currTimestamp, prevTimestamp *int64
		var currValue, prevValue *string
		if err := rows.Scan(&smartEventID, &hubspotContactID, &currTimestamp,
			&prevTimestamp, &currValue, &prevValue); err != nil {
			log.WithError(err).Error("Failed to get rows.")
			return nil, err
		}

		if _, exist := smartEventMetaData[smartEventID]; exist {
			log.WithField("smart_event_id", smartEventID).Error("Duplicate smart event id.")
			return nil, errors.New("Duplicate smart event id")
		}

		smartEventMetaData[smartEventID] = make(map[string]interface{})
		smartEventMetaData[smartEventID]["doc_id"] = hubspotContactID
		smartEventMetaData[smartEventID]["curr_timestamp"] = *currTimestamp
		if prevTimestamp != nil {
			smartEventMetaData[smartEventID]["prev_timestamp"] = *prevTimestamp
		}

		if currValue != nil {
			smartEventMetaData[smartEventID]["curr_value"] = *currValue
		}

		if prevValue != nil {
			smartEventMetaData[smartEventID]["prev_value"] = *prevValue
		}
	}

	return smartEventMetaData, nil
}

func main() {
	env := flag.String("env", C.DEVELOPMENT, "")
	projectID := flag.Int64("project_id", 0, "Project id.")
	from := flag.Int64("from", 0, "Project id.")
	to := flag.Int64("to", 0, "Project id.")
	eventNameID := flag.String("event_name_id", "", "Event name id.")
	wetRun := flag.Bool("wet_run", false, "Wet run")
	batchSize := flag.Uint("batch_size", 1000, "Batch size for deleting smart events.")
	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")

	flag.Parse()

	defer util.NotifyOnPanic("Task#ValidateSmartEvent", *env)

	appName := "hubspot_smart_event_validator"
	config := &C.Configuration{
		Env:       *env,
		SentryDSN: *sentryDSN,
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
			Certificate: *memSQLCertificate,
			AppName:     appName,
		},
		PrimaryDatastore: *primaryDatastore,
	}

	C.InitConf(config)
	// Initialize configs and connections.
	err := C.InitDB(*config)
	if err != nil {
		log.Error("Failed to initialize DB.")
		os.Exit(1)
	}

	if *from < 1 || *to < 1 {
		log.Panic("Invalid start and end timestamp")
	}

	if *wetRun {
		log.Info("Wet run enabled.")
	}

	eventName, status := store.GetStore().GetSmartEventFilterEventNameByID(*projectID, *eventNameID, false)
	if status != http.StatusFound {
		log.Error("Failed to get smart event")
		os.Exit(1)
	}

	filterExp, err := model.GetDecodedSmartEventFilterExp(eventName.FilterExpr)
	if err != nil {
		log.WithError(err).Error("Failed to decode smart event filter")
		os.Exit(1)
	}

	propertyName := filterExp.Filters[0].Name

	var selectStmnt string
	var withParams []interface{}
	if *primaryDatastore == C.DatastoreTypeMemSQL {
		selectStmnt, withParams = getMemSQLQuery(*projectID, *eventNameID, *from, *to, propertyName, model.HubspotDocumentTypeContact)

	} else {
		selectStmnt, withParams = GetSmartEventMetaDataQuery(*projectID, propertyName, *eventNameID, *from, *to)

	}

	smartEventMetaData, err := getSmartEventMetaData(selectStmnt, withParams)
	if err != nil {
		log.WithError(err).Error("Failed to get smart event meta data.")
		os.Exit(1)
	}

	logCtx := log.WithFields(log.Fields{"from": *from, "to": *to, "event_name": eventName.Name, "event_name_id": eventName.ID})
	totalCount := 0
	validSmartEvents := 0
	invalidSmartEvents := 0
	validSmartEventIDs := map[string]bool{}
	invalidSmartEventIDs := map[string]bool{}
	validDocIDs := []interface{}{}
	invalidDocIDs := []interface{}{}
	for eventID := range smartEventMetaData {
		currentProperties := map[string]interface{}{
			propertyName: smartEventMetaData[eventID]["curr_value"],
		}

		prevProperties := map[string]interface{}{
			propertyName: smartEventMetaData[eventID]["prev_value"],
		}
		docID := smartEventMetaData[eventID]["doc_id"]
		valid := model.CRMFilterEvaluator(*projectID, &currentProperties, &prevProperties, filterExp, model.CompareStateBoth)
		if valid {
			validSmartEvents++
			validSmartEventIDs[eventID] = true
			validDocIDs = append(validDocIDs, docID)
			logCtx.WithFields(log.Fields{"event_id": eventID, "doc_id": docID}).Info("Valid smart event.")
		} else {
			invalidSmartEvents++
			invalidSmartEventIDs[eventID] = true
			invalidDocIDs = append(invalidDocIDs, docID)
			logCtx.WithFields(log.Fields{"event_id": eventID, "doc_id": docID}).Info("Invalid smart event.")
		}
		totalCount++
	}

	logCtx.WithFields(log.Fields{"total_count": totalCount, "valid_smart_events_count": validSmartEvents, "invalid_smart_events_count": invalidSmartEvents,
		"valid_smart_events_id": validSmartEventIDs, "invalid_smart_events_id": invalidSmartEventIDs}).Info("Completed validations.")
	logCtx.WithFields(log.Fields{"valid_doc_ids": validDocIDs, "invalid_doc_ids": invalidDocIDs}).Info("Document ids.")
	if *wetRun {
		logCtx.Info("Starting wet run.")
		eventIDs := []string{}
		for eventID := range invalidSmartEventIDs {
			eventIDs = append(eventIDs, eventID)
		}

		status := deleteSmartEventsByIDs(*projectID, *eventNameID, eventIDs, int(*batchSize))
		if status != http.StatusAccepted {
			logCtx.Error("Failed to deleteSmartEventsByIDs.")
		} else {
			logCtx.Info("Successfully deleted smart events.")
		}

		logCtx.Info("Completed wet run.")
	}

}

func deleteSmartEventsByIDs(projectID int64, eventNameID string, IDs []string, batchSize int) int {
	if eventNameID == "" || len(IDs) < 1 || projectID == 0 {
		log.WithFields(log.Fields{"project_id": projectID, "event_name_id": eventNameID}).Error("Invalid parameters.")
		return http.StatusBadRequest
	}

	status := store.GetStore().DeleteEventsByIDsInBatchForJob(projectID, eventNameID, IDs, batchSize)
	if status != http.StatusAccepted {
		log.WithFields(log.Fields{"project_id": projectID, "event_name_id": eventNameID}).Error("Failed to delete smart events.")
	}

	return status
}
