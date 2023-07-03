package main

import (
	C "factors/config"
	"factors/model/model"
	"factors/model/store"
	SDK "factors/sdk"
	U "factors/util"
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

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
	projectIDs := flag.String("project_ids", "", "Projects for which the events and group user are to be populated")

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

	for _, setting := range linkedinProjectSettings {
		errMsg, errCode := createGroupUserAndEvents(setting)
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

func createGroupUserAndEvents(linkedinProjectSetting model.LinkedinProjectSettings) (string, int) {
	domainDataSet, errCode := store.GetStore().GetDomainData(linkedinProjectSetting.ProjectId)
	if errCode != http.StatusOK {
		return "Failed to get domain data from linkedin", errCode
	}
	projectID, err := strconv.ParseInt(linkedinProjectSetting.ProjectId, 10, 64)
	if err != nil {
		return err.Error(), http.StatusInternalServerError
	}
	eventNameViewedAD, errCode := store.GetStore().CreateOrGetUserCreatedEventName(&model.EventName{
		ProjectId: projectID,
		Name:      U.GROUP_EVENT_NAME_LINKEDIN_VIEWED_AD,
	})
	if errCode != http.StatusCreated && errCode != http.StatusConflict {
		return "Failed in creating viewed ad event name", errCode
	}
	eventNameClickedAD, errCode := store.GetStore().CreateOrGetUserCreatedEventName(&model.EventName{
		ProjectId: projectID,
		Name:      U.GROUP_EVENT_NAME_LINKEDIN_CLICKED_AD,
	})
	if errCode != http.StatusCreated && errCode != http.StatusConflict {
		return "Failed in creating clicked ad event name", errCode
	}
	timeZone, errCode := store.GetStore().GetTimezoneForProject(projectID)
	if errCode != http.StatusFound {
		return "Failed to get timezone", errCode
	}
	location, err := time.LoadLocation(string(timeZone))
	if err != nil {
		return "Failed to load location via timezone", http.StatusInternalServerError
	}
	for _, domainData := range domainDataSet {
		logFields := log.Fields{
			"project_id": projectID,
			"doument":    domainData,
		}
		defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
		logCtx := log.WithFields(logFields)

		if domainData.Domain != "" && domainData.Domain != "$none" {
			properties := U.PropertiesMap{
				U.LI_DOMAIN:            domainData.Domain,
				U.LI_HEADQUARTER:       domainData.HeadQuarters,
				U.LI_LOCALIZED_NAME:    domainData.LocalizedName,
				U.LI_VANITY_NAME:       domainData.VanityName,
				U.LI_PREFERRED_COUNTRY: domainData.PreferredCountry,
			}

			timestamp, err := time.ParseInLocation("20060102", domainData.Timestamp, location)
			if err != nil {
				return err.Error(), http.StatusInternalServerError
			}

			unixTimestamp := timestamp.Unix()
			userID, errCode := SDK.TrackGroupWithDomain(projectID, U.GROUP_NAME_LINKEDIN_COMPANY, domainData.Domain, properties, unixTimestamp)
			if errCode != http.StatusOK {
				logCtx.Error("Failed in TrackGroupWithDomain")
				return "Failed in TrackGroupWithDomain", errCode
			}

			if domainData.Impressions != 0 {
				viewedADEvent := model.Event{
					EventNameId: eventNameViewedAD.ID,
					Timestamp:   unixTimestamp,
					ProjectId:   projectID,
					UserId:      userID,
				}
				viewedADEventPropertiesMap := map[string]interface{}{
					U.LI_AD_VIEW_COUNT: domainData.Impressions,
					U.EP_SKIP_SESSION:  U.PROPERTY_VALUE_TRUE,
				}
				viewedADEventPropertiesJsonB, err := U.EncodeStructTypeToPostgresJsonb(&viewedADEventPropertiesMap)
				if err != nil {
					logCtx.WithError(err).Error("Failed in encoding viewed ad properties to JSONb")
					return "Failed in encoding viewed ad properties to JSONb", http.StatusInternalServerError
				}
				viewedADEvent.Properties = *viewedADEventPropertiesJsonB

				_, errCode = store.GetStore().CreateEvent(&viewedADEvent)
				if errCode != http.StatusCreated {
					logCtx.Error("Failed in creating viewed ad event")
					return "Failed in creating viewed ad event", errCode
				}
			}

			if domainData.Clicks != 0 {
				clickedADEvent := model.Event{
					EventNameId: eventNameClickedAD.ID,
					Timestamp:   unixTimestamp + 1,
					ProjectId:   projectID,
					UserId:      userID,
				}
				clickedADEventPropertiesMap := map[string]interface{}{
					U.LI_AD_CLICK_COUNT: domainData.Clicks,
					U.EP_SKIP_SESSION:   U.PROPERTY_VALUE_TRUE,
				}
				clickedADEventPropertiesJsonB, err := U.EncodeStructTypeToPostgresJsonb(&clickedADEventPropertiesMap)
				if err != nil {
					logCtx.WithError(err).Error("Failed in encoding clicked ad properties to JSONb")
					return "Failed in encoding clicked ad properties to JSONb", http.StatusInternalServerError
				}
				clickedADEvent.Properties = *clickedADEventPropertiesJsonB

				_, errCode = store.GetStore().CreateEvent(&clickedADEvent)
				if errCode != http.StatusCreated {
					logCtx.Error("Failed in creating clicked ad event")
					return "Failed in creating clicked ad event", errCode
				}
			}
		}

		err = store.GetStore().UpdateLinkedinGroupUserCreationDetails(domainData)
		if err != nil {
			logCtx.WithError(err).Error("Failed in updating user creation details")
			return "Failed in updating user creation details", http.StatusInternalServerError
		}
	}
	return "", http.StatusOK
}
