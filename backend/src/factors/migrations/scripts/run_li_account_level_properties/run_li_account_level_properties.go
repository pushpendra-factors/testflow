package main

import (
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	C "factors/config"

	log "github.com/sirupsen/logrus"
)

type Status struct {
	ProjectID string `json:"project_id"`
	ErrCode   int    `json:"err_code"`
	ErrMsg    string `json:"err_msg"`
}

func main() {
	env := flag.String("env", C.DEVELOPMENT, "")
	projectIDs := flag.String("project_ids", "", "Comma separated project ids to run for. * to run for all")
	excludeProjectIDs := flag.String("exclude_project_ids", "", "Comma separated project ids to exclude the run for. * to exclude for all")
	dataMismatchRun := flag.Bool("data_mismatch_run", false, "")

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
	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")

	flag.Parse()

	if *env != C.DEVELOPMENT &&
		*env != C.STAGING &&
		*env != C.PRODUCTION {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	} else if *projectIDs == "" {
		panic(fmt.Errorf("invalid project id %s", *projectIDs))
	}

	taskID := "Script#AddLiAccountLevelProperties"
	defer U.NotifyOnPanic(taskID, *env)
	logCtx := log.WithFields(log.Fields{"Prefix": taskID})

	config := &C.Configuration{
		Env: *env,
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
			IsPSCHost:   *isPSCHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
			Certificate: *memSQLCertificate,
			AppName:     taskID,
		},
		PrimaryDatastore:              *primaryDatastore,
		EnableDomainsGroupByProjectID: "*",
		RedisHost:                     *redisHost,
		RedisPort:                     *redisPort,
		RedisHostPersistent:           *redisHostPersistent,
		RedisPortPersistent:           *redisPortPersistent,
		CaptureSourceInUsersTable:     *captureSourceInUsersTable,
		RemoveDisabledEventUserPropertiesByProjectID: *removeDisabledEventUserPropertiesByProjectID,
		SentryDSN: *sentryDSN,
	}

	C.InitConf(config)
	C.InitRedis(config.RedisHost, config.RedisPort)
	C.InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)
	C.InitSentryLogging(config.SentryDSN, config.AppName)
	// Initialize configs and connections and close with defer.
	err := C.InitDB(*config)
	if err != nil {
		logCtx.WithError(err).Fatal("Failed to run migration. Init failed.")
	}
	db := C.GetServices().Db
	defer db.Close()
	logCtx = logCtx.WithFields(log.Fields{
		"Env":       *env,
		"ProjectID": *projectIDs,
	})

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
		logCtx.Fatal("Failed to get linkedin settings")
	}

	allProjects, allowedProjectIDs, _ := C.GetProjectsFromListWithAllProjectSupport(*projectIDs, "")
	excludeAllProjects, excludedProjectIDs, _ := C.GetProjectsFromListWithAllProjectSupport(*excludeProjectIDs, "")
	for _, setting := range linkedinProjectSettings {
		errMsg, errCode := "", 0
		projectID, _ := strconv.ParseInt(setting.ProjectId, 10, 64)

		if checkIfProjectRunAllowed(projectID, excludeAllProjects, excludedProjectIDs, allProjects, allowedProjectIDs) {
			if *dataMismatchRun {
				errMsg, errCode = updateGroupUserWithMismatchedProperties(projectID)
			} else {
				errMsg, errCode = updateGroupUserWithAccountLevelProperties(projectID)
			}
		} else {
			errMsg, errCode = fmt.Sprintf("ProjectID: %d not allowed", projectID), http.StatusBadRequest
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

	logCtx.Warn("End of linkedin account level properties migration job", syncStatusSuccesses, syncStatusFailures)
}

func checkIfProjectRunAllowed(projectID int64, excludeAllProjects bool, excludedProjectIDs map[int64]bool,
	allProjects bool, allowedProjectIDs map[int64]bool) bool {
	if excludeAllProjects || excludedProjectIDs[projectID] {
		return false
	}
	if allProjects || allowedProjectIDs[projectID] {
		return true
	}
	return false
}

func updateGroupUserWithMismatchedProperties(projectID int64) (string, int) {
	db := C.GetServices().Db

	group, errCode := store.GetStore().GetGroup(projectID, model.GROUP_NAME_LINKEDIN_COMPANY)
	if errCode != http.StatusFound {
		return "Failed to get group.", http.StatusInternalServerError
	}
	source := model.GetGroupUserSourceByGroupName(U.GROUP_NAME_LINKEDIN_COMPANY)

	imprEventName, err := store.GetStore().GetEventNameIDFromEventName(U.GROUP_EVENT_NAME_LINKEDIN_VIEWED_AD, projectID)
	if err != nil {
		log.WithError(err).Error("Failed to get impr event name")
		return "Failed to get impression eventname", http.StatusInternalServerError
	}
	clickEventName, err := store.GetStore().GetEventNameIDFromEventName(U.GROUP_EVENT_NAME_LINKEDIN_CLICKED_AD, projectID)
	if err != nil {
		log.WithError(err).Error("Failed to get clicks event name")
		return "Failed to get click eventname", http.StatusInternalServerError
	}

	queryStr := "With users_0 as (SELECT id, group_%d_id, is_group_user, JSON_EXTRACT_STRING(properties, '$li_total_ad_view_count') as user_impressions, " +
		"JSON_EXTRACT_STRING(properties, '$li_total_ad_click_count') as user_clicks from users where project_id = ? and source = ?), " +
		"events_0 as (SELECT user_id, sum(JSON_EXTRACT_STRING(properties, '$li_ad_view_count')) as event_impressions, " +
		"Case when sum(JSON_EXTRACT_STRING(properties, '$li_ad_click_count')) is null then 0 else " +
		"sum(JSON_EXTRACT_STRING(properties, '$li_ad_click_count')) END as event_clicks from events where project_id = ? and event_name_id in (?,?)" +
		"group by user_id order by user_id) SELECT id, group_%d_id, is_group_user from users_0 join events_0 on id=user_id " +
		"where user_impressions != event_impressions or user_clicks != event_clicks"
	queryStr = fmt.Sprintf(queryStr, group.ID, group.ID)
	var users []model.User
	err = db.Raw(queryStr, projectID, source, projectID, imprEventName.ID, clickEventName.ID).Find(&users).Error
	if err != nil {
		log.WithError(err).Error("Failed to get group users")
		return "Failed to find group users", http.StatusInternalServerError
	}
	log.WithFields(log.Fields{"project_id": projectID, "count": len(users)}).Info("Mismatch count")
	for _, user := range users {
		err = getMetricsAndUpdateGroupUserProperties(projectID, user, group.ID, imprEventName.ID, clickEventName.ID)
		if err != nil {
			return err.Error(), http.StatusInternalServerError
		}
	}
	return "", http.StatusOK
}

func updateGroupUserWithAccountLevelProperties(projectID int64) (string, int) {
	db := C.GetServices().Db

	group, errCode := store.GetStore().GetGroup(projectID, model.GROUP_NAME_LINKEDIN_COMPANY)
	if errCode != http.StatusFound {
		return "Failed to get group.", http.StatusInternalServerError
	}
	source := model.GetGroupUserSourceByGroupName(U.GROUP_NAME_LINKEDIN_COMPANY)
	// precheckValid, errCode := precheckExitingUsersWithoutAccountLevelProperty(projectID)
	// if !precheckValid {
	// 	if errCode != http.StatusOK {
	// 		return "Failed to run precheck", errCode
	// 	}
	// 	return "Precheck failed - total users and users without property not equal", http.StatusBadRequest
	// }

	var users []model.User
	err := db.Select(fmt.Sprintf("id, group_%d_id, is_group_user", group.ID)).
		Where("project_id = ? and source = ?", projectID, source).Limit("100000").Find(&users).Error
	if err != nil {
		log.WithError(err).Error("Failed to get group users")
		return "Failed to find group users", http.StatusInternalServerError
	}

	imprEventName, err := store.GetStore().GetEventNameIDFromEventName(U.GROUP_EVENT_NAME_LINKEDIN_VIEWED_AD, projectID)
	if err != nil {
		log.WithError(err).Error("Failed to get impr event name")
		return "Failed to get impression eventname", http.StatusInternalServerError
	}
	clickEventName, err := store.GetStore().GetEventNameIDFromEventName(U.GROUP_EVENT_NAME_LINKEDIN_CLICKED_AD, projectID)
	if err != nil {
		log.WithError(err).Error("Failed to get clicks event name")
		return "Failed to get click eventname", http.StatusInternalServerError
	}
	for _, user := range users {
		err = getMetricsAndUpdateGroupUserProperties(projectID, user, group.ID, imprEventName.ID, clickEventName.ID)
		if err != nil {
			return err.Error(), http.StatusInternalServerError
		}
	}
	postCheckValid, errCode := postcheckExitingUsersWithoutAccountLevelProperty(projectID)
	if !postCheckValid {
		if errCode != http.StatusOK {
			return "Failed to run postcheck", errCode
		}
		return "Postcheck failed - total users and users with property not equal", http.StatusBadRequest
	}
	return "", http.StatusOK
}

type UserCount struct {
	UserCount int64 `json:"user_count"`
}

// func precheckExitingUsersWithoutAccountLevelProperty(projectID int64) (bool, int) {
// 	db := C.GetServices().Db

// 	source := model.GetGroupUserSourceByGroupName(U.GROUP_NAME_LINKEDIN_COMPANY)
// 	var userCountWithProperty UserCount
// 	err := db.Table("users").Select("count(*) as user_count").
// 		Where("project_id = ? and source = ? and JSON_EXTRACT_STRING(properties, ?) is not null", projectID, source, U.LI_TOTAL_AD_VIEW_COUNT).
// 		Find(&userCountWithProperty).Error
// 	if err != nil {
// 		log.WithError(err).Error("Failed running precheck query")
// 		return false, http.StatusInternalServerError
// 	}
// 	if userCountWithProperty.UserCount > 0 {
// 		return false, http.StatusOK
// 	}
// 	return true, http.StatusOK
// }

func getMetricsAndUpdateGroupUserProperties(projectID int64, user model.User, groupIndex int, imprEventNameID, clickEventNameID string) error {
	totalImpressions, totalClicks, err := getTotalImpressionsAndClicksFromEvents(projectID, user.ID, imprEventNameID, clickEventNameID)
	if err != nil {
		return err
	}

	groupID, err := model.GetGroupUserGroupID(&user, groupIndex)
	if err != nil {
		return err
	}
	sourceStr := model.GetGroupUserSourceNameByGroupName(U.GROUP_NAME_LINKEDIN_COMPANY)

	newProperties := make(U.PropertiesMap)
	newProperties[U.LI_TOTAL_AD_CLICK_COUNT] = totalClicks
	newProperties[U.LI_TOTAL_AD_VIEW_COUNT] = totalImpressions

	propertiesMap := map[string]interface{}(newProperties)

	currTime := time.Now()
	timestamp := currTime.Unix()
	_, err = store.GetStore().CreateOrUpdateGroupPropertiesBySource(projectID, U.GROUP_NAME_LINKEDIN_COMPANY, groupID, user.ID, &propertiesMap, timestamp, timestamp, sourceStr)
	if err != nil {
		return err
	}
	return nil
}

type res struct {
	Sum float64 `json:"sum"`
}

func getTotalImpressionsAndClicksFromEvents(projectID int64, userID string, imprEventNameID, clickEventNameID string) (float64, float64, error) {
	db := C.GetServices().Db
	var totalImpresionCount res
	var totalClickCount res

	err := db.Raw("Select SUM(JSON_EXTRACT_STRING(properties, '$li_ad_view_count')) as sum from events where project_id = ? and user_id = ? and event_name_id = ?",
		projectID, userID, imprEventNameID).Scan(&totalImpresionCount).Error
	if err != nil {
		return 0, 0, err
	}

	err = db.Raw("select SUM(JSON_EXTRACT_STRING(properties, '$li_ad_click_count')) as sum from events where project_id = ? and user_id = ? and event_name_id = ?",
		projectID, userID, clickEventNameID).Scan(&totalClickCount).Error

	if err != nil {
		return 0, 0, err
	}
	return totalImpresionCount.Sum, totalClickCount.Sum, nil
}

func postcheckExitingUsersWithoutAccountLevelProperty(projectID int64) (bool, int) {
	db := C.GetServices().Db

	source := model.GetGroupUserSourceByGroupName(U.GROUP_NAME_LINKEDIN_COMPANY)
	var userCountWithoutProperty UserCount
	err := db.Table("users").Select("count(*) as user_count").
		Where("project_id = ? and source = ? and JSON_EXTRACT_STRING(properties, ?) is null", projectID, source, U.LI_TOTAL_AD_VIEW_COUNT).
		Find(&userCountWithoutProperty).Error
	if err != nil {
		log.WithError(err).Error("Failed running postcheck query")
		return false, http.StatusInternalServerError
	}
	if userCountWithoutProperty.UserCount > 0 {
		return false, http.StatusOK
	}
	return true, http.StatusOK
}
