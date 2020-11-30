package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	C "factors/config"
	M "factors/model"
	U "factors/util"

	"github.com/jinzhu/gorm/dialects/postgres"
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

type EventWithProperties struct {
	ID            string          `json:"id"`
	Name          string          `json:"event_name"`
	PropertiesMap U.PropertiesMap `json:"properties_map"`
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

func doesPropertiesMapHaveKeys(propertiesMap U.PropertiesMap,
	keys []string) (bool, bool, U.PropertiesMap) {

	filteredPropertiesMap := U.PropertiesMap{}

	if propertiesMap == nil {
		return false, false, filteredPropertiesMap
	}

	for i := range keys {
		value, exists := propertiesMap[keys[i]]
		if exists && value != nil && value != "" {
			filteredPropertiesMap[keys[i]] = value
		}
	}

	hasAll := len(filteredPropertiesMap) == len(keys)
	hasSome := len(filteredPropertiesMap) > 0 && len(filteredPropertiesMap) < len(keys)

	return hasAll, hasSome, filteredPropertiesMap
}

type PropertiesWithCount struct {
	// Count will be based only authorName.
	// As all properties required are present when authorName is present.
	// And primary property to be fixed is authorName.
	Count      int
	Properties U.PropertiesMap
}

func getEventsWithoutPropertiesAndWithPropertiesByName(projectID uint64, from, to int64) (
	[]EventWithProperties, *map[string]U.PropertiesMap, int) {
	logCtx := log.WithField("project_id", projectID).
		WithField("from", from).
		WithField("to", to)

	eventsWithoutProperties := make([]EventWithProperties, 0, 0)
	// map[event_name]map[authorName]*PropertiesWithCount
	propertiesByNameAndOccurence := make(map[string]map[string]*PropertiesWithCount, 0)

	queryStartTimestamp := U.TimeNowUnix()
	// LIKE '%.%' is for excluding custom event_names which are not urls.
	queryStmnt := "SELECT events.id, name, properties FROM events" + " " +
		"LEFT JOIN event_names ON events.event_name_id = event_names.id" + " " +
		"WHERE events.project_id = ? AND event_names.name != '$session'" + " " +
		"AND event_names.name LIKE '%.%' AND timestamp BETWEEN ? AND ?"

	db := C.GetServices().Db
	rows, err := db.Raw(queryStmnt, projectID, from, to).Rows()
	if err != nil {
		logCtx.WithError(err).
			Error("Failed to execute raw query on getEventsWithoutPropertiesAndWithPropertiesByName.")
		return eventsWithoutProperties, nil, http.StatusInternalServerError
	}
	defer rows.Close()
	logCtx = logCtx.WithField("query_exec_time_in_secs", U.TimeNowUnix()-queryStartTimestamp)

	var rowCount int
	for rows.Next() {
		var id string
		var name string
		var properties postgres.Jsonb

		err = rows.Scan(&id, &name, &properties)
		if err != nil {
			logCtx.WithError(err).Error("Failed to scan row.")
			continue
		}

		propertiesMap, err := U.DecodePostgresJsonbAsPropertiesMap(&properties)
		if err != nil {
			logCtx.WithError(err).Error("Failed to decode properties.")
			continue
		}

		if _, exists := propertiesByNameAndOccurence[name]; !exists {
			propertiesByNameAndOccurence[name] = make(map[string]*PropertiesWithCount, 0)
		}

		hasAll, hasSome, filteredPropertiesMap := doesPropertiesMapHaveKeys(*propertiesMap, MandatoryProperties)
		if hasAll {
			authorName, asserted := filteredPropertiesMap["authorName"].(string)
			if !asserted {
				log.WithField("author", authorName).Warn("Failed to assert author name as string.")
				continue
			}

			if _, exists := propertiesByNameAndOccurence[name][authorName]; !exists {
				propertiesByNameAndOccurence[name][authorName] = &PropertiesWithCount{
					Properties: filteredPropertiesMap,
					Count:      1,
				}
			} else {
				// Always overwrite, to keep adding hasAll state.
				(*propertiesByNameAndOccurence[name][authorName]).Properties = filteredPropertiesMap
				(*propertiesByNameAndOccurence[name][authorName]).Count++
			}
		}

		if hasSome {
			propAuthorName, exists := filteredPropertiesMap["authorName"]
			if !exists && propAuthorName != nil {
				continue
			}
			authorName := propAuthorName.(string)

			if propertiesWithCount, authorExists := propertiesByNameAndOccurence[name][authorName]; !authorExists {
				propertiesByNameAndOccurence[name][authorName] = &PropertiesWithCount{
					Properties: filteredPropertiesMap,
					Count:      1,
				}
			} else {
				// Do no overwrite, hasAll state with hasSome state.
				if allKeysExist, _, _ := doesPropertiesMapHaveKeys((*propertiesWithCount).Properties,
					MandatoryProperties); allKeysExist {
					continue
				}

				// Add properties if more properties available this time.
				if len(filteredPropertiesMap) > len((*propertiesWithCount).Properties) {
					(*propertiesByNameAndOccurence[name][authorName]).Properties = filteredPropertiesMap
				}
				(*propertiesByNameAndOccurence[name][authorName]).Count++
			}
		}

		// Adds all events for update, to support update with most occurrence.
		eventsWithoutProperties = append(
			eventsWithoutProperties,
			EventWithProperties{
				ID:            id,
				Name:          name,
				PropertiesMap: *propertiesMap,
			},
		)

		rowCount++
	}

	propertiesByName := getPropertiesByNameAndMaxOccurrence(&propertiesByNameAndOccurence)

	logCtx.WithField("rows", rowCount).Info("Scanned all rows.")
	return eventsWithoutProperties, propertiesByName, http.StatusFound
}

func getPropertiesByNameAndMaxOccurrence(
	propertiesByNameAndOccurence *map[string]map[string]*PropertiesWithCount,
) *map[string]U.PropertiesMap {

	propertiesWithCount := make(map[string]PropertiesWithCount, 0)
	for name, propertiesByAuthor := range *propertiesByNameAndOccurence {
		for _, pwc := range propertiesByAuthor {
			// Select the poroeprties with max occurrence count.
			if (*pwc).Count > propertiesWithCount[name].Count &&
				// Consider only max no.of properties available.
				len((*pwc).Properties) >= len(propertiesWithCount[name].Properties) {

				propertiesWithCount[name] = *pwc
			}
		}
	}

	propertiesByName := make(map[string]U.PropertiesMap)
	for name, pwc := range propertiesWithCount {
		if pwc.Count > 0 && len(pwc.Properties) > 0 {
			propertiesByName[name] = pwc.Properties
		}
	}

	return &propertiesByName
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
	eventsWithoutProperties []EventWithProperties,
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

		errCode := M.OverwriteEventPropertiesByID(projectID, event.ID, newPropertiesJsonb)
		if errCode != http.StatusAccepted {
			logCtx.Error("Failed to update event properties after adding missing properties.")
			continue
		}
		noOfUpdates++
	}

	return http.StatusAccepted, noOfUpdates
}

func main() {
	env := flag.String("env", "development", "")
	dbHost := flag.String("db_host", "localhost", "")
	dbPort := flag.Int("db_port", 5432, "")
	dbUser := flag.String("db_user", "autometa", "")
	dbName := flag.String("db_name", "autometa", "")
	dbPass := flag.String("db_pass", "@ut0me7a", "")
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

	taskID := "Task:Yourstory:AddMissingEventProperties"
	healthcheckPingID := C.HealthcheckYourstoryAddPropertiesPingID
	defer C.PingHealthcheckForPanic(taskID, *env, healthcheckPingID)

	config := &C.Configuration{
		AppName:            "yourstory_add_missing_event_properties",
		Env:                *env,
		GCPProjectID:       *gcpProjectID,
		GCPProjectLocation: *gcpProjectLocation,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
		},
	}

	C.InitConf(config.Env)

	err := C.InitDB(config.DBInfo)
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

	events, eventNamePropertiesLookup, errCode := getEventsWithoutPropertiesAndWithPropertiesByName(*projectID, from, to)
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
