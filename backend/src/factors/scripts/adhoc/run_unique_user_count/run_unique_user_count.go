package main

import (
	C "factors/config"
	"factors/model/model"
	"factors/model/store"
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

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	isPSCHost := flag.Int("memsql_is_psc_host", C.MemSQLDefaultDBParams.IsPSCHost, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL,
		"Primary datastore type as memsql or postgres")

	customStartTime := flag.Int64("start_time", 0, "")
	customEndTime := flag.Int64("end_time", 0, "")
	projectIdFlag := flag.Uint64("project_id", 0, "Project Id.")
	attributionKey := flag.String("attribution_key", "", "")
	attributionMethod := flag.String("attribution_method", "", "")
	eventName := flag.String("event_name", "", "")
	lookbackDays := flag.Int("lookback_days", 0, "")
	enableOptimisedFilterOnProfileQuery := flag.Bool("enable_optimised_filter_on_profile_query",
		false, "Enables filter optimisation logic for profiles query.")

	var linkedEvents linkedEvent
	flag.Var(&linkedEvents, "linked_event", "")
	flag.Parse()

	defer util.NotifyOnPanic("Task#AttributionQuery", *env)

	appName := "unique_user_count"
	config := &C.Configuration{
		Env:     *env,
		AppName: appName,
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
			IsPSCHost:   *isPSCHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
			Certificate: *memSQLCertificate,
			AppName:     appName,
		},
		PrimaryDatastore:                    *primaryDatastore,
		EnableOptimisedFilterOnProfileQuery: *enableOptimisedFilterOnProfileQuery,
	}

	// Setup.
	// Initialize configs and connections.
	C.InitConf(config)

	var err error
	err = C.InitDB(*config)
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

	query := &model.AttributionQuery{
		AnalyzeType: model.AnalyzeTypeUsers,
		//LinkedEvents:           linkedEventsArray,
		From: *customStartTime,
		To:   *customEndTime,
		//ConversionEvent:        *eventName,
		LookbackDays:           *lookbackDays,
		AttributionKey:         *attributionKey,
		AttributionMethodology: *attributionMethod,
	}
	var debugQueryKey string
	result, err := store.GetStore().ExecuteAttributionQueryV0(int64(*projectIdFlag), query, debugQueryKey, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnProfileQuery())
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
