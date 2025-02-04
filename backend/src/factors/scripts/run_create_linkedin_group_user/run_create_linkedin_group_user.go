package main

import (
	C "factors/config"
	"factors/model/model"
	"factors/model/store"
	LCE "factors/task/linkedin_company_engagements"
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
	dryRun := flag.Bool("dry_run_1", false, "Dry run mode")
	projectIDs := flag.String("project_ids", "", "Projects for which the events and group user are to be populated")
	excludeProjectIDs := flag.String("exclude_project_ids", "", "")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	isPSCHost := flag.Int("memsql_is_psc_host", C.MemSQLDefaultDBParams.IsPSCHost, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")
	captureSourceInUsersTable := flag.String("capture_source_in_users_table", "*", "")
	removeDisabledEventUserPropertiesByProjectID := flag.String("remove_disabled_event_user_properties",
		"", "List of projects to disable event user property population in events.")
	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")
	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")

	overrideHealthcheckPingID := flag.String("healthcheck_ping_id", "", "Override default healthcheck ping id.")
	overrideAppName := flag.String("app_name", "", "Override default app_name.")
	runV3Change := flag.String("run_v3_change", "", "Runs new changes with campaign group data")
	batchSize := flag.Int("batch_size", 5, "Num of parallel go routine processes")

	flag.Parse()
	if *env != "development" &&
		*env != "staging" &&
		*env != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}

	defaultAppName := "create_linkedin_engagements_group_user_job"
	defaultHealthcheckPingID := C.HealthcheckLinkedinGroupUserPingID
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
		CaptureSourceInUsersTable:     *captureSourceInUsersTable,
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

	linkedinProjectSettings := make([]model.LinkedinProjectSettings, 0)
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
	allProjects, allowedProjectIDs, disallowedProjectIDs := C.GetProjectsFromListWithAllProjectSupport(*runV3Change, *excludeProjectIDs)
	for _, setting := range linkedinProjectSettings {
		errMsg, errCode := "", 0
		projectID, _ := strconv.ParseInt(setting.ProjectId, 10, 64)
		if !disallowedProjectIDs[projectID] {
			if allProjects || allowedProjectIDs[projectID] {
				errMsg, errCode = LCE.CreateGroupUserAndEventsV3(setting, *batchSize)
			} else {
				errMsg, errCode = LCE.CreateGroupUserAndEventsV2(setting, *batchSize)
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

/* flow
1. Get all linkedin documents for which we have to create users and events
2. Create LinkedinViewedAd and LinkedinClickedAd 'eventName'. Doing it on top reduces number of times this is done
3. On step 1 we also recieve the min timestamp (Calling it t_min) from where we'll start the backfill
	3.1 Since for older events we won't have campaign id, we fetch 2 diff types of events data for each of 2 events i.e impr event and click event (total 4)
	3.2 for impr events -> map of events with campaign data where timestamp >= t_min, and map of events with campaign_data = null where timestamp >= t_min
	3.3 Same thing as above for clicks events
4. Loop thorugh all documents from step 1
	4.1 create/get group user using domain
	4.2 For impression do a check if event creation is required
	4.3 Same for click event
	4.4 Mark group user created as true, if there's no error. (Doesn't matter if event created or not)
Note: eligibity criteria for 4.2 & 4.3 -> property value i.e impressions or clicks, should be > 0
	for older events data, event with same timestamp and domain shouldn't exist
	for newer events data, event with same timestamp, domain and campaign id shouldn't exist
*/
