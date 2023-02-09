package main

import (
	"encoding/json"
	"errors"
	C "factors/config"
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

type DeleteEventNameStatus struct {
	ID                    string `json:"id"`
	Name                  string `json:"name"`
	TotalEvents           int    `json:"total_events"`
	TotalEmptySmartEvents int    `json:"total_empty_smart_events"`
}

func main() {
	env := flag.String("env", C.DEVELOPMENT, "")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")

	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")

	projectID := flag.Int64("project_id", 0, "Project Id.")
	eventNameIDs := flag.String("event_name_ids", "", "Event name Id.")
	startTime := flag.Int64("start_timestamp", 0, "Staring timestamp from events search.")
	endTime := flag.Int64("end_timestamp", 0, "Ending timestamp from events search. End timestamp will be excluded")
	wetRun := flag.Bool("wet", false, "Wet run")
	flag.Parse()
	defer util.NotifyOnPanic("Task#run_remove_empty_smart_events", *env)

	taskID := "run_remove_empty_smart_events"
	if *projectID == 0 {
		log.Error("projectId not provided")
		os.Exit(1)
	}

	if *eventNameIDs == "" {
		log.Panic("Invalid event_name_id")
	}

	if *startTime <= 0 || *endTime <= 0 {
		log.Panic("Invalid range.")
	}

	config := &C.Configuration{
		AppName: taskID,
		Env:     *env,
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
			Certificate: *memSQLCertificate,
			AppName:     taskID,
		},
		SentryDSN:        *sentryDSN,
		PrimaryDatastore: *primaryDatastore,
	}

	C.InitConf(config)
	C.InitSentryLogging(config.SentryDSN, config.AppName)

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

	eventNames, err := getSmartEventNamesWithAnyChangeFilter(*projectID, *eventNameIDs)
	if err != nil {
		log.WithError(err).Error("Failed to get event names.")
		return
	}

	deleteSmartEventStatus := make([]DeleteEventNameStatus, 0)
	for i := range eventNames {
		events, err := getSmartEventsByEventNameID(*projectID, eventNames[i].ID, *startTime, *endTime)
		if err != nil {
			log.WithFields(log.Fields{"project_id": projectID}).WithError(err).Error("Failed get events for deletion.")
			return
		}

		deleteEventsCount, err := deleteSmartEventsWithEmptyProperties(*projectID, &eventNames[i], events, *wetRun)
		if err != nil {
			log.WithFields(log.Fields{"project_id": projectID}).WithError(err).Error("Failed to delete smart events.")
			return
		}
		deleteSmartEventStatus = append(deleteSmartEventStatus, DeleteEventNameStatus{
			ID:                    eventNames[i].ID,
			Name:                  eventNames[i].Name,
			TotalEvents:           len(events),
			TotalEmptySmartEvents: deleteEventsCount,
		})
	}

	log.WithFields(log.Fields{"project_id": projectID, "delete_smart_event_status": deleteSmartEventStatus}).Info("Status for delete smart events.")

	log.WithFields(log.Fields{"project_id": *projectID, "event_name_ids": eventNameIDs}).Info("Completed delete")
}

func getSmartEventNamesWithAnyChangeFilter(projectID int64, eventNameIDs string) ([]model.EventName, error) {
	eventNames, status := store.GetStore().GetSmartEventFilterEventNames(projectID, false)
	if status != http.StatusFound {
		return nil, errors.New("failed to get smart event filter event names")
	}

	anyChangeEventNames := make([]model.EventName, 0)
	for i := range eventNames {
		crmEventFilter, err := model.GetDecodedSmartEventFilterExp(eventNames[i].FilterExpr)
		if err != nil {
			log.WithFields(log.Fields{"project_id": projectID}).WithError(err).Error("Failed to decode smart event filter expression.")
			continue
		}
		if crmEventFilter.FilterEvaluationType == model.FilterEvaluationTypeAny {
			anyChangeEventNames = append(anyChangeEventNames, eventNames[i])
		}
	}

	if eventNameIDs == "*" {
		return anyChangeEventNames, nil
	}

	requiredEventNameIds := strings.Split(eventNameIDs, ",")

	requiredEventNames := make([]model.EventName, 0)
	for _, requiredID := range requiredEventNameIds {
		for j := range anyChangeEventNames {
			if anyChangeEventNames[j].ID == requiredID {
				requiredEventNames = append(requiredEventNames, anyChangeEventNames[j])
				break
			}
		}
	}

	return requiredEventNames, nil
}

func deleteSmartEventsWithEmptyProperties(projectID int64, eventName *model.EventName, events []model.Event, wetRun bool) (int, error) {
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "event_name": eventName, "events": len(events), "wet_run": wetRun})
	if projectID == 0 || eventName == nil {
		logCtx.Error("Invalid paremeters.")
		return 0, errors.New("invalid paremters")
	}

	crmEventFilter, err := model.GetDecodedSmartEventFilterExp(eventName.FilterExpr)
	if err != nil {
		logCtx.WithError(err).Error("Failed to decode smart event filter expression for deleting.")
		return 0, err
	}
	logCtx = logCtx.WithFields(log.Fields{"crm_event_filter": crmEventFilter})

	prevPropertyName := model.GetCurrPropertyName(crmEventFilter.Filters[0].Name, crmEventFilter.Source, crmEventFilter.ObjectType)
	currPropertyName := model.GetPrevPropertyName(crmEventFilter.Filters[0].Name, crmEventFilter.Source, crmEventFilter.ObjectType)

	deleteIDs := make([]string, 0)
	// log list of events to be deleted
	deleteEvents := make([]model.Event, 0)
	for i := range events {
		if eventName.ID != events[i].EventNameId {
			logCtx.WithFields(log.Fields{"event": events[i]}).Error("Invalid event for event name id.")
			return 0, errors.New("invalid event for event name id")
		}

		properties := map[string]interface{}{}
		err := json.Unmarshal(events[i].Properties.RawMessage, &properties)
		if err != nil {
			logCtx.WithFields(log.Fields{"event": events[i]}).WithError(err).Error("Failed decode event properties.")
			return 0, err
		}

		if (properties[prevPropertyName] == nil || properties[prevPropertyName] == "") &&
			(properties[currPropertyName] == nil || properties[currPropertyName] == "") {
			deleteIDs = append(deleteIDs, events[i].ID)
			deleteEvents = append(deleteEvents, events[i])
		}
	}

	if len(deleteEvents) <= 0 {
		logCtx.Info("No events to delete.")
		return 0, nil
	}

	for i := range deleteEvents {
		logCtx.WithFields(log.Fields{"event": deleteEvents[i], "event_name_id": deleteEvents[i].EventNameId}).
			Warning("Event will be deleted.")
	}

	if !wetRun {
		return len(deleteEvents), nil
	}

	return len(deleteEvents), deleteEventsByID(projectID, deleteIDs)
}

func deleteEventsByID(projectID int64, eventIDs []string) error {
	if projectID == 0 || len(eventIDs) == 0 {
		return errors.New("invalid parameters")
	}

	db := C.GetServices().Db

	err := db.Where("project_id = ? AND id in ( ? )", projectID, eventIDs).Delete(&model.Event{}).Error
	if err != nil {
		log.WithFields(log.Fields{"project_id": projectID, "event_ids": eventIDs}).WithError(err).Error("Failed to delete empty smart events")
		return err
	}

	return nil
}

func getSmartEventsByEventNameID(projectID int64, eventNameID string, startTime, endTime int64) ([]model.Event, error) {
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "event_name_id": eventNameID, "start_time": startTime, "end_time": endTime})
	if projectID == 0 || eventNameID == "" || startTime <= 0 || endTime <= 0 {
		logCtx.Error("Invalid paremeters")
		return nil, errors.New("invalid parameters")
	}

	db := C.GetServices().Db
	var events []model.Event
	err := db.Model(&model.Event{}).Where("project_id = ? AND event_name_id = ? AND timestamp between ? AND ?",
		projectID, eventNameID, startTime, endTime).Find(&events).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to get smart events.")
		return nil, err
	}

	return events, nil
}
