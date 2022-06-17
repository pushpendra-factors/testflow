package main

import (
	"flag"

	C "factors/config"
	"factors/model/store"

	T "factors/task"

	taskWrapper "factors/task/task_wrapper"

	log "github.com/sirupsen/logrus"
)

func main() {

	envFlag := flag.String("env", C.DEVELOPMENT, "Environment. Could be development|staging|production.")
	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")
	lookback := flag.Int("lookback", 1, "lookback for job")
	bigqueryProjectId := flag.String("bigquery_project_id", "", "")
	bigqueryCredential := flag.String("bigquery_credential_json", "", "")
	overrideHealthcheckPingID := flag.String("healthcheck_ping_id", "", "Override default healthcheck ping id.")
	flag.Parse()

	appName := "load_leadsquared_integration_data"
	defaultHealthcheckPingID := C.HealthcheckLeadSquaredIntegrationSyncPingID
	healthcheckPingID := C.GetHealthcheckPingID(defaultHealthcheckPingID, *overrideHealthcheckPingID)
	defer C.PingHealthcheckForPanic(appName, *envFlag, healthcheckPingID)

	config := &C.Configuration{
		Env: *envFlag,
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
		log.WithError(err).Fatal("Failed to initialize DB")
	}

	projectIdsArray := make([]uint64, 0)
	mappings, err := store.GetStore().GetAllLeadSquaredEnabledProjects()
	if err != nil {
		C.PingHealthcheckForFailure(healthcheckPingID, "Failed to get LeadSquared Projects")
	}
	for id, _ := range mappings {
		projectIdsArray = append(projectIdsArray, id)
	}
	configs := make(map[string]interface{})
	configs["BigqueryProjectId"] = *bigqueryProjectId
	configs["BigqueryCredential"] = *bigqueryCredential
	status := taskWrapper.TaskFuncWithProjectId("LeadSquaredIntegrationSync", *lookback, projectIdsArray, T.LeadSquaredIntegration, configs)
	log.Info(status)
	C.PingHealthcheckForSuccess(healthcheckPingID, status)

}
