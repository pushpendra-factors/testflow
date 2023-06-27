package main

import (
	C "factors/config"
	"factors/model/model"
	"factors/model/store"
	G2 "factors/task/g2"
	SP "factors/task/smart_properties"

	U "factors/util"
	"flag"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// Todo: introduce more constants
func main() {
	env := flag.String("env", C.DEVELOPMENT, "")
	projectIDs := flag.String("project_ids", "*", "List of project_id to run for.")
	dryRunG2 := flag.Bool("dry_run_smart_properties", false, "Dry run mode for smart properties job")
	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	isPSCHost := flag.Int("memsql_is_psc_host", C.MemSQLDefaultDBParams.IsPSCHost, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")

	overrideHealthcheckETLPingID := flag.String("healthcheck_etl_ping_id", "", "Override default healthcheck ping id for etl.")
	overrideHealthcheckEnrichmentPingID := flag.String("healthcheck_enrichment_ping_id", "", "Override default healthcheck ping id for enirchment.")

	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")
	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")
	enableDomainsGroupByProjectID := flag.String("enable_domains_group_by_project_id", "*", "")
	captureSourceInUsersTable := flag.String("capture_source_in_users_table", "*", "")

	flag.Parse()
	if *env != "development" &&
		*env != "staging" &&
		*env != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}

	defer U.NotifyOnPanic("Script#g2_enrichment_job", *env)
	appName := "g2_enrichment_job"
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
		},
		PrimaryDatastore:              *primaryDatastore,
		EnableDomainsGroupByProjectID: *enableDomainsGroupByProjectID,
		CaptureSourceInUsersTable:     *captureSourceInUsersTable,
		RedisHost:                     *redisHost,
		RedisPort:                     *redisPort,
		RedisHostPersistent:           *redisHostPersistent,
		RedisPortPersistent:           *redisPortPersistent,
	}
	defaultHealthcheckETLPingID := C.HeathCheckG2ETLPingID
	healthcheckETLPingID := C.GetHealthcheckPingID(defaultHealthcheckETLPingID, *overrideHealthcheckETLPingID)
	defaultHealthcheckEnrichmentPingID := C.HeathCheckG2EnrichmentPingID
	healthcheckEnrichmentPingID := C.GetHealthcheckPingID(defaultHealthcheckEnrichmentPingID, *overrideHealthcheckEnrichmentPingID)
	C.InitConf(config)
	err := C.InitDB(*config)
	if err != nil {
		log.Fatal("Init failed.")
	}
	C.InitRedis(config.RedisHost, config.RedisPort)
	C.InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)
	db := C.GetServices().Db
	defer db.Close()

	var errCode int
	g2ProjectSettings := make([]model.G2ProjectSettings, 0)

	allProjects, projectIDMap, _ := C.GetProjectsFromListWithAllProjectSupport(*projectIDs, "")
	if allProjects {
		g2ProjectSettings, errCode = store.GetStore().GetG2EnabledProjectSettings()
	} else if len(projectIDMap) > 0 {
		onlyProjectIds := make([]int64, 0, len(projectIDMap))
		for k := range projectIDMap {
			onlyProjectIds = append(onlyProjectIds, k)
		}
		g2ProjectSettings, errCode = store.GetStore().GetG2EnabledProjectSettingsForProjects(onlyProjectIds)
	} else {
		log.Fatal("No projectIDs provided in flag")
	}

	if errCode != http.StatusOK {
		log.Fatal("Failed to get project settings for g2")
	}
	syncStatusETL, syncStatusEnrichment := G2Enrichment(g2ProjectSettings)

	if !*dryRunG2 {
		if len(syncStatusETL["Failure"]) > 0 {
			C.PingHealthcheckForFailure(healthcheckETLPingID, syncStatusETL)
		} else {
			C.PingHealthcheckForSuccess(healthcheckETLPingID, syncStatusETL)
		}

		if len(syncStatusEnrichment["Failure"]) > 0 {
			C.PingHealthcheckForFailure(healthcheckEnrichmentPingID, syncStatusEnrichment)
		} else {
			C.PingHealthcheckForSuccess(healthcheckEnrichmentPingID, syncStatusEnrichment)
		}
	}
}

// G2 etl and enrichment inside the following method
func G2Enrichment(g2ProjectSettings []model.G2ProjectSettings) (map[string][]SP.Status, map[string][]SP.Status) {
	syncStatusFailures := make([]SP.Status, 0)
	syncStatusSuccesses := make([]SP.Status, 0)

	for _, setting := range g2ProjectSettings {
		errMsg := G2.PerformETLForProject(setting)
		if errMsg != "" && errMsg != G2.NO_DATA_ERROR {
			failure := SP.Status{
				ProjectID: setting.ProjectID,
				ErrMsg:    errMsg,
			}
			syncStatusFailures = append(syncStatusFailures, failure)
		} else {
			syncStatusSuccesses = append(syncStatusSuccesses, SP.Status{ProjectID: setting.ProjectID, ErrMsg: errMsg})
		}
	}

	log.Warn("End of etl part of g2 sync job")
	syncStatusETL := map[string][]SP.Status{
		"Success": syncStatusSuccesses,
		"Failure": syncStatusFailures,
	}

	syncStatusFailures = make([]SP.Status, 0)
	syncStatusSuccesses = make([]SP.Status, 0)
	for _, setting := range g2ProjectSettings {
		errMsg, errCode := G2.PerformCompanyEnrichmentAndUserAndEventCreationForProject(setting)
		if errMsg != "" {
			failure := SP.Status{
				ProjectID: setting.ProjectID,
				ErrMsg:    errMsg,
				ErrCode:   errCode,
			}
			syncStatusFailures = append(syncStatusFailures, failure)
		} else {
			syncStatusSuccesses = append(syncStatusSuccesses, SP.Status{ProjectID: setting.ProjectID, ErrMsg: errMsg})
		}
	}
	log.Warn("End of user and event creation part of g2 sync job")
	syncStatusEnrichment := map[string][]SP.Status{
		"Success": syncStatusSuccesses,
		"Failure": syncStatusFailures,
	}
	return syncStatusETL, syncStatusEnrichment
}
