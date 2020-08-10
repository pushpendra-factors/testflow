package main

import (
	C "factors/config"
	M "factors/model"
	"factors/util"
	"flag"
	"os"

	log "github.com/sirupsen/logrus"
)

//go run run_unique_user_count.go --attribution_key=<Campaign/Source> --linked_event=<linked event1> --link_event=<linked event2> --attribution_method=<> --event_name_id=<id> --project_id=<id> --start_time=<> --end_time=<>

type linkedEvent []string

func (l *linkedEvent) String() string {
	return ""
}

func (l *linkedEvent) Set(value string) error {
	*l = append(*l, value)
	return nil
}

func main() {

	env := flag.String("env", "development", "")

	dbHost := flag.String("db_host", "localhost", "")
	dbPort := flag.Int("db_port", 5432, "")
	dbUser := flag.String("db_user", "autometa", "")
	dbName := flag.String("db_name", "autometa", "")
	dbPass := flag.String("db_pass", "@ut0me7a", "")

	customStartTime := flag.Int64("start_time", 0, "")
	customEndTime := flag.Int64("end_time", 0, "")
	projectIdFlag := flag.Uint64("project_id", 0, "Project Id.")
	attributionKey := flag.String("attribution_key", "", "")
	attributionMethod := flag.String("attribution_method", "", "")
	eventName := flag.String("event_name", "", "")
	lookbackDays := flag.Int("lookback_days", 0, "")

	var linkedEvents linkedEvent
	flag.Var(&linkedEvents, "linked_event", "")
	flag.Parse()

	defer util.NotifyOnPanic("Task#AttributionQuery", *env)

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

	// Setup.
	// Initialize configs and connections.
	C.InitConf(config.Env)

	var err error
	err = C.InitDB(config.DBInfo)
	if err != nil {
		log.Error("Failed to initialize DB.")
		os.Exit(0)
	}

	if eventName == nil || *eventName == "" {
		log.Error("No Converision event provided")
		os.Exit(0)
	}

	linkedEventsArray := make([]string, 0)
	for _, value := range linkedEvents {
		linkedEventsArray = append(linkedEventsArray, value)
	}

	query := &M.AttributionQuery{
		LinkedEvents:           linkedEventsArray,
		From:                   *customStartTime,
		To:                     *customEndTime,
		ConversionEvent:        *eventName,
		LookbackDays:           *lookbackDays,
		AttributionKey:         *attributionKey,
		AttributionMethodology: *attributionMethod,
	}

	result, err := M.ExecuteAttributionQuery(*projectIdFlag, query)
	if err != nil {
		log.Error("Failed to execute query")
	}

	if result != nil {
		log.Info(result.Headers)
		for _, row := range result.Rows {
			log.Info(row)
		}
	} else {
		log.Error("Result is Nil")
	}
}
