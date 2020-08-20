package main

import (
	"flag"
	"fmt"
	"net/http"
	"strings"

	C "factors/config"
	M "factors/model"
	U "factors/util"

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

// MandatoryProperties - Event properties to be added, if missing.
var MandatoryProperties = []string{
	"authorName",
	"authors",
	"articleCategory",
	"tags",
	"brandName",
}

type EventWithProperties struct {
	ID            string          `json:"id"`
	Name          string          `json:"event_name"`
	PropertiesMap U.PropertiesMap `json:"properties_map"`
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

func getEventsWithoutPropertiesAndWithPropertiesByName(projectID uint64, from, to int64) (
	[]EventWithProperties, *map[string]U.PropertiesMap, int) {
	logCtx := log.WithField("project_id", projectID).
		WithField("from", from).
		WithField("to", to)

	eventsWithoutProperties := make([]EventWithProperties, 0, 0)
	propertiesByName := make(map[string]U.PropertiesMap, 0)

	queryStartTimestamp := U.TimeNowUnix()
	queryStmnt := "SELECT events.id, name, properties FROM events" + " " +
		"LEFT JOIN event_names ON events.event_name_id = event_names.id" + " " +
		"WHERE events.project_id = ? AND event_names.name != '$session' AND timestamp BETWEEN ? AND ?"

	db := C.GetServices().Db
	rows, err := db.Raw(queryStmnt, projectID, from, to).Rows()
	if err != nil {
		logCtx.WithError(err).
			Error("Failed to execute raw query on getEventsWithoutPropertiesAndWithPropertiesByName.")
		return eventsWithoutProperties, &propertiesByName, http.StatusInternalServerError
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

		hasAll, hasSome, filteredPropertiesMap := doesPropertiesMapHaveKeys(*propertiesMap, MandatoryProperties)
		if hasAll {
			// add to lookup if key available.
			propertiesByName[name] = filteredPropertiesMap
		} else {
			// add to list, for updating properties using lookup.
			eventsWithoutProperties = append(
				eventsWithoutProperties,
				EventWithProperties{
					ID:            id,
					Name:          name,
					PropertiesMap: *propertiesMap,
				},
			)
		}

		if hasSome {
			// Do no overwrite, hasAll state with hasSome state.
			if allKeysExist, _, _ := doesPropertiesMapHaveKeys(propertiesByName[name],
				MandatoryProperties); !allKeysExist {
				propertiesByName[name] = filteredPropertiesMap
			}
		}

		rowCount++
	}

	logCtx.WithField("rows", rowCount).Info("Scanned all rows.")
	return eventsWithoutProperties, &propertiesByName, http.StatusFound
}

func getPropertiesForName(
	name string,
	propertiesByName *map[string]U.PropertiesMap,
) *U.PropertiesMap {

	if properties, exists := (*propertiesByName)[name]; exists {
		return &properties
	}

	// Get properties from non-amp page.
	if strings.HasSuffix(name, "/amp") {
		if properties, exists := (*propertiesByName)[strings.TrimSuffix(name, "/amp")]; exists {
			return &properties
		}
	}

	return nil
}

func addEventPropertiesByName(
	projectID uint64,
	propertiesByName *map[string]U.PropertiesMap,
	eventsWithoutProperties []EventWithProperties,
) int {
	logCtx := log.WithField("project_id", projectID)

	if projectID == 0 {
		return http.StatusBadRequest
	}

	if len(eventsWithoutProperties) == 0 {
		logCtx.Error("No events without properties.")
		return http.StatusBadRequest
	}

	if propertiesByName == nil || len(*propertiesByName) == 0 {
		logCtx.Error("Empty properties by name lookup map.")
		return http.StatusInternalServerError
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

		isPropertiesAdded := false
		for i := range MandatoryProperties {
			key := MandatoryProperties[i]

			value, exists := (*propertiesFromEvent)[key]
			if !exists {
				logCtx.WithField("property", key).Error("Property not found.")
				continue
			}

			// Add properties doesn't exsits already.
			// Do not overwrite the exsiting properties.
			if _, exists := event.PropertiesMap[key]; !exists {
				event.PropertiesMap[key] = value
				isPropertiesAdded = true
			}
		}

		if !isPropertiesAdded {
			logCtx.Error("Mandatory properties not for the event. Skipping update.")
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
	}

	return http.StatusAccepted
}

func main() {
	env := flag.String("env", "development", "")
	dbHost := flag.String("db_host", "localhost", "")
	dbPort := flag.Int("db_port", 5432, "")
	dbUser := flag.String("db_user", "autometa", "")
	dbName := flag.String("db_name", "autometa", "")
	dbPass := flag.String("db_pass", "@ut0me7a", "")

	projectID := flag.Uint64("project_id", 398, "Yourstory project_id.")
	customEndTimestamp := flag.Int64("custom_end_timestamp", 0, "Custom end timestamp.")
	maxLookbackDays := flag.Int64("max_lookback_days", 1, "Fix properties for last given days. Default 1.")

	flag.Parse()

	if *env != "development" &&
		*env != "staging" &&
		*env != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}

	taskID := "Task:Yourstory:AddMissingEventProperties"
	defer U.NotifyOnPanic(taskID, *env)

	config := &C.Configuration{
		AppName: "yourstory:add_missing_event_properties",
		Env:     *env,
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
		log.WithError(err).Fatal("Failed to initialize db.")
	}

	if C.IsProduction() {
		log.SetFormatter(&log.JSONFormatter{})
	}

	if *customEndTimestamp > 0 && *customEndTimestamp < 1577836800 {
		log.WithField("end_timestamp", *customEndTimestamp).Fatal("Invalid custom end timestamp.")
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
	timerangeString := fmt.Sprintf("Timerange from %d to %d.", from, to)

	events, eventNamePropertiesLookup, errCode := getEventsWithoutPropertiesAndWithPropertiesByName(*projectID, from, to)
	if errCode != http.StatusFound {
		failureMsg = "Failed to get events without properties and properties lookup map." + " " + timerangeString
		log.WithField("err_code", errCode).Error(failureMsg)
	}

	errCode = addEventPropertiesByName(*projectID, eventNamePropertiesLookup, events)
	if errCode != http.StatusAccepted {
		failureMsg = "Failed to add missing event properties." + " " + timerangeString
		log.WithField("err_code", errCode).Error(failureMsg)
	}

	// Notify only on failure.
	if failureMsg != "" {
		if err := U.NotifyThroughSNS(taskID, *env, failureMsg); err != nil {
			log.Fatalf("Failed to notify status %+v", failureMsg)
		}
	}

	log.Info("Successfully updated missing event properties." + " " + timerangeString)
}
