package main

import (
	"errors"
	C "factors/config"
	"factors/model/model"
	"factors/model/store"
	"factors/util"
	"flag"
	"net/http"
	"os"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

func main() {
	env := flag.String("env", C.DEVELOPMENT, "")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")
	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")
	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")

	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")
	sentryRollupSyncInSecs := flag.Int("sentry_rollup_sync_in_seconds", 300, "Enables to send errors to sentry in given interval.")
	projectIDList := flag.String("project_ids", "", "Project Ids.")
	startTime := flag.Int64("start_timestamp", 0, "Staring timestamp for users.")
	endTime := flag.Int64("end_timestamp", 0, "Ending timestamp for users. End timestamp will be included")
	wetRun := flag.Bool("wet", false, "Wet run")
	cacheSortedSet := flag.Bool("cache_with_sorted_set", false, "Cache with sorted set keys")
	allowEmailDomainsByProjectID := flag.String("allow_email_domain_by_project_id", "", "Allow email domains for domain group")
	flag.Parse()
	defer util.NotifyOnPanic("Task#run_backfill_users_domain", *env)

	taskID := "run_backfill_users_domain"
	if *projectIDList == "" {
		log.Error("project_ids not provided")
		os.Exit(1)
	}

	if *startTime <= 0 || *endTime <= 0 {
		log.Panic("Invalid range.")
	}

	config := &C.Configuration{
		AppName: taskID,
		Env:     *env,
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
			Certificate: *memSQLCertificate,
			AppName:     taskID,
		},
		RedisHost:                    *redisHost,
		RedisPort:                    *redisPort,
		RedisHostPersistent:          *redisHostPersistent,
		RedisPortPersistent:          *redisPortPersistent,
		SentryDSN:                    *sentryDSN,
		SentryRollupSyncInSecs:       *sentryRollupSyncInSecs,
		PrimaryDatastore:             *primaryDatastore,
		CacheSortedSet:               *cacheSortedSet,
		AllowEmailDomainsByProjectID: *allowEmailDomainsByProjectID,
	}

	C.InitConf(config)
	C.InitSortedSetCache(config.CacheSortedSet)
	C.InitSentryLogging(config.SentryDSN, config.AppName)
	C.InitRedis(config.RedisHost, config.RedisPort)
	C.InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)

	err := C.InitDB(*config)
	if err != nil {
		log.Error("Failed to initialize DB.")
		os.Exit(1)
	}

	if !*wetRun {
		log.Info("Running in dry run")
	} else {
		log.Info("Running in wet run")
	}

	_, allowedProjects, _ := C.GetProjectsFromListWithAllProjectSupport(*projectIDList, "")

	projectStatus := make(map[int64]interface{})
	for projectID := range allowedProjects {
		err, totaluniqueUsersByCustomerUserIDUpdated, totalusersWithoutCustomerUserIDUpdated := startBackfillUserDomains(projectID, *startTime, *endTime, *wetRun)
		if err != nil {
			projectStatus[projectID] = err.Error()
			break
		}

		projectStatus[projectID] = map[string]int{
			"totaluniqueUsersByCustomerUserIDUpdated": totaluniqueUsersByCustomerUserIDUpdated,
			"totalusersWithoutCustomerUserIDUpdated":  totalusersWithoutCustomerUserIDUpdated,
		}
	}

	log.WithFields(log.Fields{"project_status": projectStatus}).Info("Process completed.")
}

func startBackfillUserDomains(projectID int64, startTime, endTime int64, wetRun bool) (error, int, int) {
	log.WithFields(log.Fields{"project_id": projectID, "start_time": startTime, "end_time": endTime, "wet_run": wetRun}).Info("Running startBackfillUserDomains.")

	users, status := getUsersForDomainAssociation(projectID, startTime, endTime)
	if status != http.StatusFound {
		if status != http.StatusNotFound {
			log.WithFields(log.Fields{"project_id": projectID, "start_time": startTime, "end_time": endTime}).Error("Failed to getUsersForDomainAssociation.")
			return errors.New("failed to getUsersForDomainAssociation"), 0, 0
		}
		log.WithFields(log.Fields{"project_id": projectID, "start_time": startTime, "end_time": endTime}).Info("No users to process.")
		return nil, 0, 0
	}
	uniqueUsersByCustomerUserID := make(map[string]string)
	usersWithoutCustomerUserID := make([]string, 0)

	for i := range users {
		if users[i].CustomerUserId == "" {
			usersWithoutCustomerUserID = append(usersWithoutCustomerUserID, users[i].ID)
		} else {
			uniqueUsersByCustomerUserID[users[i].CustomerUserId] = users[i].ID
		}
	}

	totaluniqueUsersByCustomerUserIDUpdated := 0
	totalusersWithoutCustomerUserIDUpdated := 0
	for _, userID := range uniqueUsersByCustomerUserID {
		log.WithFields(log.Fields{"project_id": projectID, "user_id": userID, "wet_run": wetRun}).Info("AssociateUserDomainsGroup by customer user id.")
		if !wetRun {
			totaluniqueUsersByCustomerUserIDUpdated++
			continue
		}

		status := store.GetStore().AssociateUserDomainsGroup(projectID, userID, "", "")
		if status != http.StatusOK && status != http.StatusNotModified {
			log.WithFields(log.Fields{"project_id": projectID, "user_id": userID}).Error("Failed to AssociateUserDomainsGroup for uniqueUsersByCustomerUserID .")
		}
	}

	for _, userID := range usersWithoutCustomerUserID {
		log.WithFields(log.Fields{"project_id": projectID, "user_id": userID}).Info("AssociateUserDomainsGroup without customer user id.")
		if !wetRun {
			totalusersWithoutCustomerUserIDUpdated++
			continue
		}

		status := store.GetStore().AssociateUserDomainsGroup(projectID, userID, "", "")
		if status != http.StatusOK && status != http.StatusNotModified {
			log.WithFields(log.Fields{"project_id": projectID, "user_id": userID}).Error("Failed to AssociateUserDomainsGroup for uniqueUsersByCustomerUserID .")
		}
	}

	return nil, totaluniqueUsersByCustomerUserIDUpdated, totalusersWithoutCustomerUserIDUpdated
}

func getUsersForDomainAssociation(projectID int64, startTime, endTime int64) ([]model.User, int) {
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "start_time": startTime, "end_time": endTime})
	if projectID == 0 || startTime == 0 || endTime == 0 {
		logCtx.Error("Invalid parameters.")
		return nil, http.StatusBadRequest
	}

	db := C.GetServices().Db

	var users []model.User
	err := db.Model(&model.User{}).Select("id, customer_user_id").Where("project_id = ? AND UNIX_TIMESTAMP(updated_at) >= ? "+
		" AND UNIX_TIMESTAMP(updated_at) <= ? AND (is_group_user = FALSE OR is_group_user IS NULL)", projectID, startTime, endTime+1).Find(&users).Error
	if err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			logCtx.WithError(err).Error("Failed to get getUsersForDomainAssociation.")
			return nil, http.StatusInternalServerError
		}

		return nil, http.StatusNotFound
	}

	if len(users) == 0 {
		return nil, http.StatusNotFound
	}

	return users, http.StatusFound
}
