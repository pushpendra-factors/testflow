package main

import (
	"factors/model/store"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	C "factors/config"
	T "factors/task"
)

func main() {
	env := flag.String("env", C.DEVELOPMENT, "")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	isPSCHost := flag.Int("memsql_is_psc_host", C.MemSQLDefaultDBParams.IsPSCHost, "")
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
	batchSizeDomains := flag.Int("batch_size_domains", 50, "batch size for number of domains to be processed in a go")
	domainsLimitAllRun := flag.Int("domains_limit_all_run", 250000, "limit for domains to be processed for all run")
	processOnlyAccountSegments := flag.Bool("process_only_account_segments", false, "This flag allows only processing of all accounts type segments")
	runAllAccountsMarkerProjectIDs := flag.String("run_all_accounts_marker_project_ids", "",
		"Project Id to run all accounts marker for. A comma separated list of project Ids and supports '*' for all projects. ex: 1,2,6,9")
	runForAllAccountsInHours := flag.Int("run_for_all_accounts_in_hours", 24, "Run domains where marker_last_run_all_accounts is greater than given hours")

	memSQLUseExactConnectionsConfig := flag.Bool("memsql_use_exact_connection_config", false, "Use exact connection for open and idle as given.")
	memSQLDBMaxOpenConnections := flag.Int("memsql_max_open_connections", 100, "Max no.of open connections allowed on connection pool of memsql")
	memSQLDBMaxIdleConnections := flag.Int("memsql_max_idle_connections", 50, "Max no.of idle connections allowed on connection pool of memsql")
	bucketName := flag.String("bucket_name", "/usr/local/var/factors/cloud_storage", "")

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
			IsPSCHost:   *isPSCHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
			Certificate: *memSQLCertificate,
			AppName:     appName,

			MaxOpenConnections:     *memSQLDBMaxOpenConnections,
			MaxIdleConnections:     *memSQLDBMaxIdleConnections,
			UseExactConnFromConfig: *memSQLUseExactConnectionsConfig,
		},
		PrimaryDatastore:               *primaryDatastore,
		SentryDSN:                      *sentryDSN,
		UseLookbackSegmentMarker:       *useLookbackSegmentMarker,
		LookbackSegmentMarker:          *lookbackSegmentMarker,
		AllowedGoRoutines:              *allowedGoRoutines,
		ProcessOnlyAccountSegments:     *processOnlyAccountSegments,
		RunAllAccountsMarkerProjectIDs: *runAllAccountsMarkerProjectIDs,
		RunForAllAccountsInHours:       *runForAllAccountsInHours,
		BatchSizeDomains:               *batchSizeDomains,
		DomainsLimitAllRun:             *domainsLimitAllRun,
	}

	C.InitConf(config)
	C.InitSentryLogging(config.SentryDSN, config.AppName)
	C.InitFilemanager(*bucketName, *env, config)

	err := C.InitDB(*config)
	if err != nil {
		log.Error("Failed to initialize DB.")
		os.Exit(1)
	}

	db := C.GetServices().Db
	defer db.Close()

	startTime := time.Now().Unix()
	RunSegmentMarkerForProjects(projectIdFlag)
	endTime := time.Now().Unix()
	timeTaken := endTime - startTime
	log.Info("Time taken for job to run in sec: ", timeTaken)
	C.PingHealthcheckForSuccess(healthcheckPingID, "segment_markup run success.")
}

func RunSegmentMarkerForProjects(projectIdFlag *string) {
	projectIdList := *projectIdFlag

	projectCount, status := store.GetStore().ProjectCountToRunAllMarkerFor()

	if status != http.StatusFound {
		err := fmt.Errorf("failed to get number of projects for segment markup all run")
		log.WithField("err_code", status).Error(err)
		return
	}

	numberOfRunsPerDay := 40

	limit := int(projectCount / numberOfRunsPerDay)

	projectIdListAllRun, status := store.GetStore().GetProjectIDsListForMarker(limit)

	if status != http.StatusFound {
		err := fmt.Errorf("failed to get list of project_ids to run all marker for")
		log.WithField("err_code", status).Error(err)
	}

	allProjects, projectIDsMap, _ := C.GetProjectsFromListWithAllProjectSupport(projectIdList, "")
	failureCount := 0
	successCount := 0

	if allProjects {
		projectIDs, errCode := store.GetStore().GetAllProjectIDs()
		if errCode != http.StatusFound {
			err := fmt.Errorf("failed to get all projects ids to run segment markup")
			log.WithField("err_code", err).Error(err)
			return
		}
		for _, projectId := range projectIDs {
			status := T.SegmentMarker(projectId, projectIdListAllRun)
			if status != http.StatusOK {
				log.WithField("project_id", projectId).Error("failed to run segment markup for project ID ")
				failureCount++
			} else {
				successCount++
			}
		}
	} else {
		for projectId, _ := range projectIDsMap {
			status := T.SegmentMarker(projectId, projectIdListAllRun)
			if status != http.StatusOK {
				log.WithField("project_id", projectId).Error("failed to run segment markup for project ID ")
				failureCount++
			} else {
				successCount++
			}
		}
	}

	log.Info(fmt.Sprintf("Succesfully ran markup job for %d projects. Failures for %d projects",
		successCount, failureCount))
}
