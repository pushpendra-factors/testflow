package main

import (
	"flag"

	C "factors/config"
	"factors/model/model"
	"factors/model/store"
	"factors/util"

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
	lookback := flag.Int("lookback", 40, "lookback for job")
	bigqueryProjectId := flag.String("bigquery_project_id", "", "")
	bigqueryCredential := flag.String("bigquery_credential_json", "", "")
	overrideHealthcheckPingID := flag.String("healthcheck_ping_id", "", "Override default healthcheck ping id.")
	flag.Parse()

	appName := "load_marketo_integration_data"
	defaultHealthcheckPingID := C.HealthcheckMarketoIntegrationPingID
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
		PrimaryDatastore: *primaryDatastore,
	}
	C.InitConf(config)

	err := C.InitDB(*config)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize DB")
	}

	projectIdsArray := make([]int64, 0)
	mappings, err := store.GetStore().GetAllActiveFiveTranMappingByIntegration(model.MarketoIntegration)

	featureProjectIDs, err := store.GetStore().GetAllProjectsWithFeatureEnabled(model.FEATURE_MARKETO, false)
	if err != nil {
		log.WithError(err).Error("Failed to get marketo feature enabled projects.")
		return
	}

	featureEnabledIntegrations := []model.FivetranMappings{}
	for i := range mappings {
		if util.ContainsInt64InArray(featureProjectIDs, mappings[i].ProjectID) {
			featureEnabledIntegrations = append(featureEnabledIntegrations, mappings[i])
		}
	}
	mappings = featureEnabledIntegrations

	for _, mapping := range mappings {
		projectIdsArray = append(projectIdsArray, mapping.ProjectID)
	}
	configs := make(map[string]interface{})
	configs["BigqueryProjectId"] = *bigqueryProjectId
	configs["BigqueryCredential"] = *bigqueryCredential
	status := taskWrapper.TaskFuncWithProjectId("MarketoIntegrationSync", *lookback, projectIdsArray, T.MarketoIntegration, configs)
	log.Info(status)
	C.PingHealthcheckForSuccess(healthcheckPingID, status)

}
