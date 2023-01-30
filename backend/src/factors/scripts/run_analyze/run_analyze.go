package main

import (
	C "factors/config"
	mqlStore "factors/model/store/memsql"
	"flag"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func main() {

	env := flag.String("env", C.DEVELOPMENT, "")
	gcpProjectID := flag.String("gcp_project_id", "", "Project ID on Google Cloud")
	gcpProjectLocation := flag.String("gcp_project_location", "", "Location of google cloud project cluster")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")
	overrideHealthcheckPingID := flag.String("healthcheck_ping_id", "", "Override default healthcheck ping id.")
	overrideAppName := flag.String("app_name", "", "Override default app_name.")
	analyzeIntervalInMins := flag.Int("analyze_tables_interval", 60, "Runs analyze for table, if not analyzed in given interval.")

	defaultAppName := "analyze_job"
	appName := C.GetAppName(defaultAppName, *overrideAppName)

	defaultHealthcheckPingID := C.HealthCheckAnalyzeJobPingID
	healthcheckPingID := C.GetHealthcheckPingID(defaultHealthcheckPingID, *overrideHealthcheckPingID)

	defer C.PingHealthcheckForPanic(appName, *env, healthcheckPingID)

	flag.Parse()

	config := &C.Configuration{
		AppName:            appName,
		Env:                *env,
		GCPProjectID:       *gcpProjectID,
		GCPProjectLocation: *gcpProjectLocation,
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
	err := C.InitDB(*config)
	if err != nil {
		log.WithError(err).Fatal("Failed to initalize db.")
	}
	db := C.GetServices().Db
	defer db.Close()

	analyzeStatus := map[string]interface{}{}
	status, failedTables := mqlStore.AnalyzeTableInAnInterval(*analyzeIntervalInMins)
	if status == http.StatusInternalServerError {
		analyzeStatus["status"] = "FAILED"
		if len(failedTables) > 0 {
			analyzeStatus["failedTables"] = failedTables
		}
	}
	analyzeStatus["status"] = "SUCCESS"

	if len(analyzeStatus) > 0 {
		analyzeStatus["analyzeStatus"] = analyzeStatus
		if analyzeStatus["status"] == "FAILED" {
			C.PingHealthcheckForFailure(healthcheckPingID,
				fmt.Sprintf("failed to analyze tables with status %s", analyzeStatus["status"]))
		} else {
			C.PingHealthcheckForSuccess(healthcheckPingID,
				fmt.Sprintf("analyze job ran sucessfully with status %s", analyzeStatus["status"]))
		}
	}

}
