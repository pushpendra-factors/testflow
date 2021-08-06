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

func selectSmartEventIDANDReferenceID(projectID uint64, eventNameID string, from, to int64) (string, []interface{}) {
	smartEventsStmnt := "select id as smart_event_id, uuid(properties ->> '$crm_reference_event_id') as reference_event_id " +
		"from events where project_id = ? and timestamp between ? and ? and event_name_id = ? "
	smartEventsParams := []interface{}{projectID, from, to, eventNameID}

	return smartEventsStmnt, smartEventsParams
}

func selectSmartEventContactByReferenceID(projectID uint64, sourceName, docIDPropertyName string) (string, []interface{}) {
	smartEventsReferenceContact := "select smart_event_id, reference_event_id, properties ->> ? " +
		"as hubspot_contact_id from " + sourceName + " left join events on " + sourceName + ".reference_event_id = events.id and project_id = ?"
	smartEventsReferenceContactParams := []interface{}{docIDPropertyName, projectID}

	return smartEventsReferenceContact, smartEventsReferenceContactParams
}

func selectContactDocumentBySyncIDAndContactID(projectID uint64, syncIDSource string, docType int) (string, []interface{}) {
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

func selectAllHubspotDocumentsByContactID(projectID uint64, sourceStep string, docType int) (string, []interface{}) {
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
func GetSmartEventMetaDataQuery(projectID uint64, propertyName, eventNameID string, from, to int64) (string, []interface{}) {
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

		smartEventMetaData[smartEventID]["curr_value"] = *currValue
		if prevValue != nil {
			smartEventMetaData[smartEventID]["prev_value"] = *prevValue
		}
	}

	return smartEventMetaData, nil
}

func main() {
	env := flag.String("env", C.DEVELOPMENT, "")
	dbHost := flag.String("db_host", C.PostgresDefaultDBParams.Host, "")
	dbPort := flag.Int("db_port", C.PostgresDefaultDBParams.Port, "")
	dbUser := flag.String("db_user", C.PostgresDefaultDBParams.User, "")
	dbName := flag.String("db_name", C.PostgresDefaultDBParams.Name, "")
	dbPass := flag.String("db_pass", C.PostgresDefaultDBParams.Password, "")
	projectID := flag.Uint64("project_id", 0, "Project id.")
	from := flag.Int64("from", 0, "Project id.")
	to := flag.Int64("to", 0, "Project id.")
	eventNameID := flag.String("event_name_id", "", "Event name id.")

	flag.Parse()

	defer util.NotifyOnPanic("Task#ValidateSmartEvent", *env)

	config := &C.Configuration{
		Env: *env,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
		},
	}

	C.InitConf(config)
	// Initialize configs and connections.
	err := C.InitDB(*config)
	if err != nil {
		log.Error("Failed to initialize DB.")
		os.Exit(1)
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

	selectStmnt, withParams := GetSmartEventMetaDataQuery(*projectID, propertyName, *eventNameID, *from, *to)

	smartEventMetaData, err := getSmartEventMetaData(selectStmnt, withParams)
	if err != nil {
		log.WithError(err).Error("Failed to get smart event meta data.")
		os.Exit(1)
	}

	totalCount := 0
	validSmartEvents := 0
	invalidSmartEvents := 0
	validSmartEventsID := map[string]bool{}
	invalidSmartEventsID := map[string]bool{}
	for eventID := range smartEventMetaData {
		currentProperties := map[string]interface{}{
			"propertyName": smartEventMetaData[eventID]["curr_value"],
		}

		prevProperties := map[string]interface{}{
			"propertyName": smartEventMetaData[eventID]["prev_value"],
		}
		valid := model.CRMFilterEvaluator(*projectID, &currentProperties, &prevProperties, filterExp, model.CompareStateBoth)
		if valid {
			validSmartEvents++
			validSmartEventsID[eventID] = true
		} else {
			invalidSmartEvents++
			invalidSmartEventsID[eventID] = true
		}
		totalCount++
	}

	log.WithFields(log.Fields{"total_count": totalCount, "valid_smart_events_count": validSmartEvents, "invalid_smart_events_count": invalidSmartEvents,
		"valid_smart_events_id": validSmartEventsID, "invalid_smart_events_id": invalidSmartEventsID}).Info("Completed validations.")
}
