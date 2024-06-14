package main

import (
	C "factors/config"
	"factors/model/model"
	"factors/model/store"
	LFC "factors/task/linkedin_frequency_capping"
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	_ "github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

type Status struct {
	ProjectID string `json:"project_id"`
	ErrCode   int    `json:"err_code"`
	ErrMsg    string `json:"err_msg"`
}

func main() {
	env := flag.String("env", "development", "")
	dryRun := flag.Bool("dry_run", false, "Dry run mode")
	projectIDs := flag.String("project_ids", "1", "Projects for which the events and group user are to be populated")
	excludeProjectIDs := flag.String("exclude_project_ids", "", "")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	isPSCHost := flag.Int("memsql_is_psc_host", C.MemSQLDefaultDBParams.IsPSCHost, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")
	removeDisabledEventUserPropertiesByProjectID := flag.String("remove_disabled_event_user_properties",
		"", "List of projects to disable event user property population in events.")
	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")
	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")
	pushToLinkedin := flag.Bool("push_to_linkedin", false, "")
	removeFromLinkedinCustom := flag.Bool("remove_from_linkedin", false, "")

	overrideHealthcheckPingID := flag.String("healthcheck_ping_id", "", "Override default healthcheck ping id.")
	overrideAppName := flag.String("app_name", "", "Override default app_name.")

	flag.Parse()
	if *env != "development" &&
		*env != "staging" &&
		*env != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}

	defaultAppName := "create_linkedin_engagements_group_user_job"
	defaultHealthcheckPingID := C.HealthcheckLinkedinGroupUserPingID
	// change healthcheck
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
		},
		PrimaryDatastore:              *primaryDatastore,
		EnableDomainsGroupByProjectID: "*",
		RedisHost:                     *redisHost,
		RedisPort:                     *redisPort,
		RedisHostPersistent:           *redisHostPersistent,
		RedisPortPersistent:           *redisPortPersistent,
		RemoveDisabledEventUserPropertiesByProjectID: *removeDisabledEventUserPropertiesByProjectID,
	}
	C.InitConf(config)
	C.InitRedis(config.RedisHost, config.RedisPort)
	C.InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)
	C.SafeFlushAllCollectors()
	// Initialize configs and connections and close with defer.
	err := C.InitDB(*config)
	if err != nil {
		log.Fatal("Failed to create linkedin group user. Init failed.")
	}
	db := C.GetServices().Db
	defer db.Close()

	syncStatusFailures := make([]Status, 0)
	syncStatusSuccesses := make([]Status, 0)

	var linkedinProjectSettings []model.LinkedinProjectSettings
	var errCode int
	if *projectIDs == "*" {
		linkedinProjectSettings, errCode = store.GetStore().GetLinkedinEnabledProjectSettings()
	} else {
		projectIDsArray := strings.Split(*projectIDs, ",")
		linkedinProjectSettings, errCode = store.GetStore().GetLinkedinEnabledProjectSettingsForProjects(projectIDsArray)
	}
	if errCode != http.StatusOK {
		log.Fatal("Failed to get linkedin settings")
	}

	_, _, disallowedProjectIDs := C.GetProjectsFromListWithAllProjectSupport("", *excludeProjectIDs)
	for _, setting := range linkedinProjectSettings {
		errMsg, errCode := "", 0
		projectID, _ := strconv.ParseInt(setting.ProjectId, 10, 64)
		if !disallowedProjectIDs[projectID] {
			errMsg, errCode = LFC.PerformLinkedinExclusionsForProject(setting, *dryRun, *pushToLinkedin, *removeFromLinkedinCustom)
		}
		if errMsg != "" || errCode != http.StatusOK {
			syncStatusFailure := Status{
				ProjectID: setting.ProjectId,
				ErrCode:   errCode,
				ErrMsg:    errMsg,
			}
			syncStatusFailures = append(syncStatusFailures, syncStatusFailure)
		} else {
			syncStatusSuccesses = append(syncStatusSuccesses, Status{ProjectID: setting.ProjectId})
		}
	}

	log.Warn("End of linkedin company engagement user creation job", syncStatusSuccesses, syncStatusFailures)
	syncStatus := map[string]interface{}{
		"Success": syncStatusSuccesses,
		"Failure": syncStatusFailures,
	}
	if !*dryRun && *env == "production" {
		if len(syncStatusFailures) > 0 {
			C.PingHealthcheckForFailure(healthcheckPingID, syncStatus)
			return
		}
		C.PingHealthcheckForSuccess(healthcheckPingID, syncStatus)
	}
}
