package main

import (
	"errors"
	IntHubspot "factors/integration/hubspot"
	IntSalesforce "factors/integration/salesforce"
	"factors/model/model"
	"factors/model/store"
	"factors/util"
	"flag"
	"fmt"
	"net/http"
	"sync"
	"time"

	C "factors/config"

	log "github.com/sirupsen/logrus"
)

func main() {
	env := flag.String("env", "development", "")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	isPSCHost := flag.Int("memsql_is_psc_host", C.MemSQLDefaultDBParams.IsPSCHost, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")

	memSQLDBMaxOpenConnections := flag.Int("memsql_max_open_connections", 50, "Max no.of open connections allowed on connection pool of memsql")
	memSQLDBMaxIdleConnections := flag.Int("memsql_max_idle_connections", 50, "Max no.of idle connections allowed on connection pool of memsql")

	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")
	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")

	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")
	sentryRollupSyncInSecs := flag.Int("sentry_rollup_sync_in_seconds", 300, "Enables to send errors to sentry in given interval.")
	cacheSortedSet := flag.Bool("cache_with_sorted_set", false, "Cache with sorted set keys")

	projectIDList := flag.String("project_ids", "*", "List of project_id to run for.")
	workers := flag.Int("workers", 1, "Number of workers per project")

	hubspotAppID := flag.String("hubspot_app_id", "", "Hubspot app id for oauth integration")
	hubspotAppSecret := flag.String("hubspot_app_secret", "", "Hubspot app secret for oauth integration")
	salesforceAppID := flag.String("salesforce_app_id", "", "")
	salesforceAppSecret := flag.String("salesforce_app_secret", "", "")
	crmJob := flag.String("crm_job", "", "")

	flag.Parse()
	if *env != "development" && *env != "staging" && *env != "production" {
		panic(fmt.Errorf("env [ %s ] not recognised", *env))
	}

	defaultAppName := "backfill_crm_objectl_url"

	config := &C.Configuration{
		AppName: defaultAppName,
		Env:     *env,
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
			IsPSCHost:   *isPSCHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
			Certificate: *memSQLCertificate,
			AppName:     defaultAppName,

			MaxOpenConnections:     *memSQLDBMaxOpenConnections,
			MaxIdleConnections:     *memSQLDBMaxIdleConnections,
			UseExactConnFromConfig: true,
		},
		PrimaryDatastore:       *primaryDatastore,
		RedisHost:              *redisHost,
		RedisPort:              *redisPort,
		RedisHostPersistent:    *redisHostPersistent,
		RedisPortPersistent:    *redisPortPersistent,
		SentryDSN:              *sentryDSN,
		SentryRollupSyncInSecs: *sentryRollupSyncInSecs,
		CacheSortedSet:         *cacheSortedSet,
		HubspotAppID:           *hubspotAppID,
		HubspotAppSecret:       *hubspotAppSecret,
		SalesforceAppID:        *salesforceAppID,
		SalesforceAppSecret:    *salesforceAppSecret,
	}

	C.InitConf(config)
	C.InitSortedSetCache(config.CacheSortedSet)

	err := C.InitDBWithMaxIdleAndMaxOpenConn(*config, 200, 100)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{"env": *env,
			"host": *memSQLHost, "port": *memSQLPort}).Panic("Failed to initialize DB.")
	}
	db := C.GetServices().Db
	defer db.Close()

	C.InitRedis(config.RedisHost, config.RedisPort)
	C.InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)
	C.InitSentryLogging(config.SentryDSN, config.AppName)
	defer C.WaitAndFlushAllCollectors(65 * time.Second)

	if *crmJob == "salesforce" {
		log.Info("Running for salesforce.")
		backfillSalesforce(*projectIDList, *workers)
	}

	if *crmJob == "hubspot" {
		log.Info("Running for hubspot.")
		backfillHubspot(*projectIDList, *workers)
	}
}

func backfillHubspot(projectIDList string, workers int) error {
	allProjects, projectIDs, _ := C.GetProjectsFromListWithAllProjectSupport(projectIDList, "")
	hubspotEnabledProjectSettings, errCode := store.GetStore().GetAllHubspotProjectSettings()
	if errCode != http.StatusFound {
		log.Panic("No projects enabled hubspot integration.")
	}

	featureProjectIDs, err := store.GetStore().GetAllProjectsWithFeatureEnabled(model.FEATURE_HUBSPOT, false)
	if err != nil {
		log.WithError(err).Error("Failed to get hubspot feature enabled projects.")
		return errors.New("failed to get feature projects")
	}

	featureEnabledProjectSettings := []model.HubspotProjectSettings{}
	for i := range hubspotEnabledProjectSettings {
		if util.ContainsInt64InArray(featureProjectIDs, hubspotEnabledProjectSettings[i].ProjectId) {
			featureEnabledProjectSettings = append(featureEnabledProjectSettings, hubspotEnabledProjectSettings[i])
		}
	}

	for i := range featureEnabledProjectSettings {
		if !allProjects && !projectIDs[featureEnabledProjectSettings[i].ProjectId] {
			continue
		}

		log.WithField("project_id", featureEnabledProjectSettings[i].ProjectId).Info("Running for project")
		err := backfillHubspotProject(&featureEnabledProjectSettings[i], workers)
		if err != nil {
			log.WithField("project_id", featureEnabledProjectSettings[i].ProjectId).WithError(err).Error("Failed to backfillHubspotProject.")
			return err
		}
	}

	return nil
}

func backfillHubspotProject(hubspotProjectSettings *model.HubspotProjectSettings, workers int) error {
	projectID := hubspotProjectSettings.ProjectId

	group, status := store.GetStore().GetGroup(projectID, model.GROUP_NAME_HUBSPOT_COMPANY)
	if status != http.StatusFound {
		return errors.New("failed to get group")
	}

	_, portalID, err := model.GetHubspotAccountTimezoneAndPortalID(projectID, "",
		hubspotProjectSettings.RefreshToken, C.GetHubspotAppID(), C.GetHubspotAppSecret())
	if err != nil {
		log.WithField("project_id", projectID).WithError(err).Error("Failed to get hubspot portal id.")
		return errors.New("failed to get hubspot portal id")
	}

	groupUsers, status := store.GetStore().GetGroupUsersGroupIdsByGroupName(projectID, model.GROUP_NAME_HUBSPOT_COMPANY)
	if status != http.StatusNotFound && status != http.StatusFound {
		return errors.New("failed to get group users")
	}

	groupUsersBatched := GetUsersListAsBatch(groupUsers, workers)

	for _, groupUsersbatch := range groupUsersBatched {
		wg := &sync.WaitGroup{}
		for i := range groupUsersbatch {
			wg.Add(1)
			go func(user *model.User) {
				defer wg.Done()
				updateHubspotObjectUrl(hubspotProjectSettings.ProjectId, user, group.ID, portalID)
			}(&groupUsersbatch[i])

		}
		wg.Wait()
	}

	return nil
}

func backfillSalesforce(projectIDList string, worker int) error {
	allProjects, projectIDs, _ := C.GetProjectsFromListWithAllProjectSupport(projectIDList, "")

	salesforceEnabledProjects, status := store.GetStore().GetAllSalesforceProjectSettings()
	if status != http.StatusFound {
		log.Error("Failed to get enabled salesforce integration.")
		return errors.New("failed to get salesforce project settings")
	}

	featureProjectIDs, err := store.GetStore().GetAllProjectsWithFeatureEnabled(model.FEATURE_SALESFORCE, false)
	if err != nil {
		log.WithError(err).Error("Failed to get salesforce feature enabled projects.")
		return errors.New("failed to get feature projects")
	}

	featureEnabledProjectSettings := []model.SalesforceProjectSettings{}
	for i := range salesforceEnabledProjects {
		if util.ContainsInt64InArray(featureProjectIDs, salesforceEnabledProjects[i].ProjectID) {
			featureEnabledProjectSettings = append(featureEnabledProjectSettings, salesforceEnabledProjects[i])
		}
	}

	for i := range featureEnabledProjectSettings {
		if !allProjects && !projectIDs[featureEnabledProjectSettings[i].ProjectID] {
			continue
		}

		log.WithField("project_id", featureEnabledProjectSettings[i].ProjectID).Info("Running for project")
		err := backfillSalesforceProject(featureEnabledProjectSettings[i], worker)
		if err != nil {
			log.WithField("project_id", featureEnabledProjectSettings[i].ProjectID).WithError(err).Error("Failed to backfillSalesforceProject.")
			return err
		}
	}

	return nil
}

func updateHubspotObjectUrl(projectID int64, groupUser *model.User, groupIndex int, portalID string) error {

	logCtx := log.WithFields(log.Fields{"project_id": projectID, "user_id": groupUser.ID})
	groupID, err := model.GetGroupUserGroupID(groupUser, groupIndex)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get group user group id. Continuing execution.")
		return err
	}

	objectURL := IntHubspot.GetHubspotObjectURl(projectID, model.HubspotDocumentTypeNameCompany, portalID, groupID)
	if objectURL == "" {
		logCtx.WithError(err).Error("Failed to get hubspot portal url.")
		return errors.New("failed to get object url")
	}

	updateProperties := map[string]interface{}{}
	updateProperties[model.GetCRMObjectURLKey(projectID, model.SmartCRMEventSourceHubspot, model.HubspotDocumentTypeNameCompany)] = objectURL
	_, err = store.GetStore().CreateOrUpdateGroupPropertiesBySource(projectID, model.GROUP_NAME_HUBSPOT_COMPANY, groupID, groupUser.ID,
		&updateProperties, util.TimeNowUnix(), util.TimeNowUnix(), model.UserSourceHubspotString)
	if err != nil {
		logCtx.WithError(err).Error("Failed to update hubspot account users for object url.")
		return errors.New("failed to update group properties")
	}

	return nil
}

func backfillSalesforceProject(salesforceProjectSetting model.SalesforceProjectSettings, workers int) error {
	group, status := store.GetStore().GetGroup(salesforceProjectSetting.ProjectID, model.GROUP_NAME_SALESFORCE_ACCOUNT)
	if status != http.StatusFound {
		return errors.New("failed to get group")
	}

	groupUsers, status := store.GetStore().GetGroupUsersGroupIdsByGroupName(salesforceProjectSetting.ProjectID, model.GROUP_NAME_SALESFORCE_ACCOUNT)
	if status != http.StatusNotFound && status != http.StatusFound {
		return errors.New("failed to get groups users")
	}

	groupUsersBatched := GetUsersListAsBatch(groupUsers, workers)
	for _, groupUsersBatch := range groupUsersBatched {
		wg := &sync.WaitGroup{}
		for i := range groupUsersBatch {
			wg.Add(1)
			go func(user *model.User) {
				defer wg.Done()
				updateSalesforceObjectURL(salesforceProjectSetting.ProjectID, user, salesforceProjectSetting.InstanceURL, group.ID)
			}(&groupUsersBatch[i])
		}
		wg.Wait()
	}

	return nil
}

func updateSalesforceObjectURL(projectID int64, groupUser *model.User, instanceUrl string, groupIndex int) error {
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "user_id": groupUser.ID})

	groupID, err := model.GetGroupUserGroupID(groupUser, groupIndex)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get salesforce group user group id. Continuing execution.")
		return err
	}

	objectURL := IntSalesforce.GetSalesforceObjectURL(projectID, instanceUrl, model.SalesforceDocumentTypeNameAccount, groupID)
	if objectURL == "" {
		logCtx.WithError(err).Error("Failed to get salesforce object url.")
		return errors.New("failed to get object url")
	}

	updateProperties := map[string]interface{}{}
	updateProperties[model.GetCRMObjectURLKey(projectID, model.SmartCRMEventSourceSalesforce, model.SalesforceDocumentTypeNameAccount)] = objectURL
	_, err = store.GetStore().CreateOrUpdateGroupPropertiesBySource(projectID, model.GROUP_NAME_SALESFORCE_ACCOUNT, groupID, groupUser.ID, &updateProperties,
		util.TimeNowUnix(), util.TimeNowUnix(), model.UserSourceSalesforceString)
	if err != nil {
		logCtx.WithError(err).Error("Failed to update salesforce account users for object url.")
		return errors.New("failed to update group properties")
	}

	return nil
}

func GetUsersListAsBatch(list []model.User, batchSize int) [][]model.User {
	batchList := make([][]model.User, 0, 0)
	listLen := len(list)
	for i := 0; i < listLen; {
		next := i + batchSize
		if next > listLen {
			next = listLen
		}

		batchList = append(batchList, list[i:next])
		i = next
	}

	return batchList
}
