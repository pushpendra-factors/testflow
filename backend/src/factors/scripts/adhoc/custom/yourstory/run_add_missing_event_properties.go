package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	C "factors/config"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"

	log "github.com/sirupsen/logrus"
)

// MandatoryProperties - Event properties to be added, if missing.
var MandatoryProperties = []string{
	"authorName",
	"brandName",
	"authors",
	"articleCategory",
	"tags",
	"publicationDate",
}

var nonMetaPages = []string{
	"yourstory.com",
	"yourstory.com/search",
	"yourstory.com/videos",
	"yourstory.com/companies/search",
	"yourstory.com/herstory",
	"yourstory.com/socialstory",
	"yourstory.com/hindi",
	"yourstory.com/tamil",
	"yourstory.com/category/funding",
	"yourstory.com/companies",
}

func getPropertiesForName(
	name string,
	propertiesByName *map[string]U.PropertiesMap,
) *U.PropertiesMap {

	// Give precendence to properties of non-amp page,
	// for amp page events.
	if strings.HasSuffix(name, "/amp") {
		if properties, exists := (*propertiesByName)[strings.TrimSuffix(name, "/amp")]; exists {
			return &properties
		}
	}

	if properties, exists := (*propertiesByName)[name]; exists {
		return &properties
	}

	return nil
}

func isNonMetaPageEventName(eventName string) bool {
	return U.StringValueIn(eventName, nonMetaPages)
}

func addEventPropertiesByName(
	projectID uint64,
	propertiesByName *map[string]U.PropertiesMap,
	eventsWithoutProperties []model.EventWithProperties,
) (int, int) {
	logCtx := log.WithField("project_id", projectID)

	noOfUpdates := 0

	if projectID == 0 {
		return http.StatusBadRequest, noOfUpdates
	}

	if len(eventsWithoutProperties) == 0 {
		logCtx.Error("No events without properties.")
		return http.StatusBadRequest, noOfUpdates
	}

	if propertiesByName == nil || len(*propertiesByName) == 0 {
		logCtx.Error("Empty properties by name lookup map.")
		return http.StatusInternalServerError, noOfUpdates
	}

	for i := range eventsWithoutProperties {
		event := eventsWithoutProperties[i]
		logCtx = logCtx.WithField("event_name", event.Name).
			WithField("id", event.ID)

		propertiesFromEvent := getPropertiesForName(event.Name, propertiesByName)
		if propertiesFromEvent == nil {
			logCtx.Error("Properties not found for event name.")
			continue
		}

		isPropertiesUpdated := false
		for i := range MandatoryProperties {
			key := MandatoryProperties[i]

			if isNonMetaPageEventName(event.Name) {
				// Removes the property if it exists already,
				// for non-meta page.
				if _, exists := event.PropertiesMap[key]; exists {
					delete(event.PropertiesMap, key)
					isPropertiesUpdated = true
				}

				// Doesn't allow adding property value,
				// for non-meta page.
				continue
			}

			valueByEventName, exists := (*propertiesFromEvent)[key]
			if !exists {
				continue
			}

			// Add property if key doesn't exsits already.
			if _, exists := event.PropertiesMap[key]; !exists {
				event.PropertiesMap[key] = valueByEventName
				isPropertiesUpdated = true
			} else {
				// Overwrite property, if the current value is not equal to max occurred value.
				if valueByEventName != nil && event.PropertiesMap[key] != valueByEventName {
					event.PropertiesMap[key] = valueByEventName
					isPropertiesUpdated = true
				}
			}
		}

		if !isPropertiesUpdated {
			continue
		}
		logCtx = logCtx.WithField("new_properties", event.PropertiesMap)

		newPropertiesMap := map[string]interface{}(event.PropertiesMap)
		newPropertiesJsonb, err := U.EncodeToPostgresJsonb(&newPropertiesMap)
		if err != nil {
			logCtx.WithError(err).
				Error("Failed to encode new properties jsonb after adding properties.")
			continue
		}

		errCode := store.GetStore().OverwriteEventPropertiesByID(projectID, event.ID, newPropertiesJsonb)
		if errCode != http.StatusAccepted {
			logCtx.Error("Failed to update event properties after adding missing properties.")
			continue
		}
		noOfUpdates++
	}

	return http.StatusAccepted, noOfUpdates
}

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
	dryRun := flag.Bool("dry_run", false, "")

	projectID := flag.Uint64("project_id", 398, "Yourstory project_id.")
	customEndTimestamp := flag.Int64("custom_end_timestamp", 0, "Custom end timestamp.")
	maxLookbackDays := flag.Int64("max_lookback_days", 1, "Fix properties for last given days. Default 1.")

	gcpProjectID := flag.String("gcp_project_id", "", "Project ID on Google Cloud")
	gcpProjectLocation := flag.String("gcp_project_location", "", "Location of google cloud project cluster")

	flag.Parse()

	if *env != "development" &&
		*env != "staging" &&
		*env != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}

	taskID := "yourstory_add_missing_event_properties"
	healthcheckPingID := C.HealthcheckYourstoryAddPropertiesPingID
	defer C.PingHealthcheckForPanic(taskID, *env, healthcheckPingID)

	config := &C.Configuration{
		AppName:            taskID,
		Env:                *env,
		GCPProjectID:       *gcpProjectID,
		GCPProjectLocation: *gcpProjectLocation,
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
		PrimaryDatastore: *primaryDatastore,
	}

	C.InitConf(config)

	err := C.InitDB(*config)
	if err != nil {
		log.WithError(err).Panic("Failed to initialize db.")
	}
	C.InitMetricsExporter(config.Env, config.AppName, config.GCPProjectID, config.GCPProjectLocation)
	defer C.WaitAndFlushAllCollectors(65 * time.Second)

	if C.IsProduction() {
		log.SetFormatter(&log.JSONFormatter{})
	}

	if *customEndTimestamp > 0 && *customEndTimestamp < 1577836800 {
		log.WithField("end_timestamp", *customEndTimestamp).Panic("Invalid custom end timestamp.")
	}

	var to int64
	if *customEndTimestamp > 0 {
		to = *customEndTimestamp
	} else {
		to = U.TimeNowUnix()
	}

	maxLookbackDaysInSeconds := 86400 * *maxLookbackDays
	from := to - maxLookbackDaysInSeconds

	var failureMsg string
	timerangeString := fmt.Sprintf("from=%d to=%d.", from, to)

	log.WithField("from", from).WithField("to", to).
		WithField("look_back_days", *maxLookbackDays).
		Info("Starting the script.")

	events, eventNamePropertiesLookup, errCode := store.GetStore().
		GetEventsWithoutPropertiesAndWithPropertiesByNameForYourStory(*projectID, from, to, MandatoryProperties)
	if errCode != http.StatusFound {
		failureMsg = "Failed to get events without properties and properties lookup map." + " " + timerangeString
		log.WithField("err_code", errCode).Error(failureMsg)
	}

	if *dryRun {
		log.WithField("lookup_size", len(*eventNamePropertiesLookup)).
			WithField("no_of_event_to_update", len(events)).
			Info("Successfull dry run.")
		os.Exit(0)
	}

	errCode, noOfUpdates := addEventPropertiesByName(*projectID, eventNamePropertiesLookup, events)
	if errCode != http.StatusAccepted {
		failureMsg = "Failed to add missing event properties." + " " + timerangeString
		log.WithField("err_code", errCode).Error(failureMsg)
	}

	// Notify only on failure.
	if failureMsg != "" {
		C.PingHealthcheckForFailure(healthcheckPingID, failureMsg)
	} else {
		C.PingHealthcheckForSuccess(healthcheckPingID, "Successfully completed")
	}

	log.WithFields(log.Fields{
		"no_of_events_without_properties": len(events),
		"size_of_lookup":                  len(*eventNamePropertiesLookup),
		"no_of_events_updated":            noOfUpdates,
	}).Info("Successfully updated missing event properties.")
}
