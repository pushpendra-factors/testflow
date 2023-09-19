package main

import (
	"factors/model/store"
	"flag"
	"fmt"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"

	C "factors/config"
	T "factors/task"
)

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

	overrideHealthcheckPingID := flag.String("healthcheck_ping_id", "", "Override default healthcheck ping id.")
	overrideAppName := flag.String("app_name", "", "Override default app_name.")

	projectIdFlag := flag.String("project_ids", "",
		"Project Id. A comma separated list of project Ids and supports '*' for all projects. ex: 1,2,6,9")

	useLookbackSegmentMarker := flag.Bool("use_lookback_segment_marker", false, "Whether to compute look_back time to fetch users in last x hours.")
	lookbackSegmentMarker := flag.Int("lookback_segment_marker", 0, "Optional: Fetch users from last x hours")
	allowedGoRoutines := flag.Int("allowed_go_routines", 1, "Number of allowed to routines")
	flag.Parse()

	if *env != "development" &&
		*env != "staging" &&
		*env != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}

	const segment_markup_ping_id = "3d376ede-5fe3-40a2-a439-20ea973df73c"
	defaultAppName := "segment-markup-job"
	defaultHealthcheckPingID := segment_markup_ping_id
	healthcheckPingID := C.GetHealthcheckPingID(defaultHealthcheckPingID, *overrideHealthcheckPingID)
	appName := C.GetAppName(defaultAppName, *overrideAppName)
	defer C.PingHealthcheckForPanic(appName, *env, healthcheckPingID)

	config := &C.Configuration{
		AppName: appName,
		Env:     *env,
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
			Certificate: *memSQLCertificate,
			AppName:     appName,
		},
		PrimaryDatastore:         *primaryDatastore,
		SentryDSN:                *sentryDSN,
		UseLookbackSegmentMarker: *useLookbackSegmentMarker,
		LookbackSegmentMarker:    *lookbackSegmentMarker,
		AllowedGoRoutines:        *allowedGoRoutines,
	}

	C.InitConf(config)
	C.InitSentryLogging(config.SentryDSN, config.AppName)

	err := C.InitDB(*config)
	if err != nil {
		log.Error("Failed to initialize DB.")
		os.Exit(1)
	}

	db := C.GetServices().Db
	defer db.Close()

	isSuccess := RunSegmentMarkerForProjects(projectIdFlag)
	if !isSuccess {
		C.PingHealthcheckForFailure(healthcheckPingID, "segment_markup run failed.")
		return
	}
	C.PingHealthcheckForSuccess(healthcheckPingID, "segment_markup run success.")
}

func RunSegmentMarkerForProjects(projectIdFlag *string) bool {
	projectIdList := *projectIdFlag
	isHealthcheckSuccess := true

	allProjects, projectIDsMap, _ := C.GetProjectsFromListWithAllProjectSupport(projectIdList, "")

	if allProjects {
		projectIDs, errCode := store.GetStore().GetAllProjectIDs()
		if errCode != http.StatusFound {
			err := fmt.Errorf("failed to get all projects ids to run segment markup")
			log.WithField("err_code", err).Error(err)
			isHealthcheckSuccess = false
			return isHealthcheckSuccess
		}
		for _, projectId := range projectIDs {
			status := T.SegmentMarker(projectId)
			if status != http.StatusOK {
				log.WithField("project_id", projectId).Error("failed to run segment markup for project ID ")
				isHealthcheckSuccess = false
			}
		}
	} else {
		for projectId, _ := range projectIDsMap {
			status := T.SegmentMarker(projectId)
			if status != http.StatusOK {
				log.WithField("project_id", projectId).Error("failed to run segment markup for project ID ")
				isHealthcheckSuccess = false
			}
		}
	}
	return isHealthcheckSuccess
}
