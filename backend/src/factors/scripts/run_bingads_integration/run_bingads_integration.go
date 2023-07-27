package main

import (
	"flag"

	C "factors/config"
	"factors/model/model"
	"factors/model/store"

	T "factors/task"

	taskWrapper "factors/task/task_wrapper"

	log "github.com/sirupsen/logrus"
)

func main() {

	envFlag := flag.String("env", C.DEVELOPMENT, "Environment. Could be development|staging|production.")
	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	isPSCHost := flag.Int("memsql_is_psc_host", C.MemSQLDefaultDBParams.IsPSCHost, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")
	lookback := flag.Int("lookback", 90, "lookback for job")
	bigqueryProjectId := flag.String("bigquery_project_id", "", "")
	bigqueryCredential := flag.String("bigquery_credential_json", "", "")
	overrideHealthcheckPingID := flag.String("healthcheck_ping_id", "", "Override default healthcheck ping id.")
	enableFeatureGatesV2 := flag.Bool("enable_feature_gates_v2", false, "")
	flag.Parse()

	appName := "load_bingads_integration_data"
	defaultHealthcheckPingID := C.HealthcheckBingAdsIntegrationPingID
	healthcheckPingID := C.GetHealthcheckPingID(defaultHealthcheckPingID, *overrideHealthcheckPingID)
	defer C.PingHealthcheckForPanic(appName, *envFlag, healthcheckPingID)

	config := &C.Configuration{
		Env: *envFlag,
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
		PrimaryDatastore:     *primaryDatastore,
		EnableFeatureGatesV2: *enableFeatureGatesV2,
	}
	C.InitConf(config)

	err := C.InitDB(*config)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize DB")
	}

	projectIdsArray := make([]int64, 0)
	mappings, err := store.GetStore().GetAllActiveFiveTranMappingByIntegration(model.BingAdsIntegration)
	for _, mapping := range mappings {
		projectIdsArray = append(projectIdsArray, mapping.ProjectID)
	}
	configs := make(map[string]interface{})
	configs["BigqueryProjectId"] = *bigqueryProjectId
	configs["BigqueryCredential"] = *bigqueryCredential
	status := taskWrapper.TaskFuncWithProjectId("BingAdsIntegration", *lookback, projectIdsArray, T.BingAdsIntegration, configs)
	log.Info(status)
	C.PingHealthcheckForSuccess(healthcheckPingID, status)

}
