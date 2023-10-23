package main

import (
	"encoding/json"
	cacheRedis "factors/cache/redis"
	C "factors/config"
	"factors/model/model"
	"flag"
	"fmt"
	"io"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

func main() {
	env := flag.String("env", C.DEVELOPMENT, "")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	isPSCHost := flag.Int("memsql_is_psc_host", C.MemSQLDefaultDBParams.IsPSCHost, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")

	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")

	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")

	overrideAppName := flag.String("app_name", "", "Override default app_name.")
	teamsAppTenantID := flag.String("teams_app_tenant_id", "", "")
	teamsAppClientID := flag.String("teams_app_client_id", "", "")
	teamsAppClientSecret := flag.String("teams_app_client_secret", "", "")
	teamsApplicationID := flag.String("teams_application_id", "", "")
	enableFeatureGatesV2 := flag.Bool("enable_feature_gates_v2", false, "")
	// blacklistedAlerts := flag.String("blacklisted_alerts", "", "")

	flag.Parse()

	if *env != "development" &&
		*env != "staging" &&
		*env != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}

	defaultAppName := "cache_cleanup_filter_lists"
	appName := C.GetAppName(defaultAppName, *overrideAppName)

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
		PrimaryDatastore:     *primaryDatastore,
		RedisHostPersistent:  *redisHostPersistent,
		RedisPortPersistent:  *redisPortPersistent,
		SentryDSN:            *sentryDSN,
		TeamsAppTenantID:     *teamsAppTenantID,
		TeamsAppClientID:     *teamsAppClientID,
		TeamsAppClientSecret: *teamsAppClientSecret,
		TeamsApplicationID:   *teamsApplicationID,
		EnableFeatureGatesV2: *enableFeatureGatesV2,
	}

	C.InitConf(config)
	C.InitSentryLogging(config.SentryDSN, config.AppName)
	C.InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)

	CacheCleanupHelper()

}

func CacheCleanupHelper() {
	pattern := "LIST:pid:*"
	keys, err := cacheRedis.GetKeysPersistent(pattern)
	if err != nil {
		log.Error(fmt.Errorf("get keys from cache failure"))
		return
	}

	for _, key := range keys {
		strKey, err := key.Key()
		if err != nil {
			log.Error(fmt.Errorf("cache key cannot be converted to string"))
		}

		// Delete sorted set and key
		err = cacheRedis.DelPersistent(key)
		if err != nil {
			log.Error(fmt.Errorf("failed to delete the sorted set"))
			return
		}

		// Read the file from the cloud and upload the
		splitKey := strings.Split(strKey, ":")
		pid := splitKey[2]
		projectID, err := strconv.ParseInt(pid, 0, 64)
		if err != nil {
			log.Error(fmt.Errorf("failed to retrieve project_id"))
		}
		reference := splitKey[len(splitKey)-1]

		err = AddListValuesToCache(projectID, reference)
		if err != nil {
			log.Error(fmt.Errorf("could not add sorted set to cache"))
		}
	}
}

func AddListValuesToCache(projectID int64, reference string) error {
	logFields := log.Fields{
		"project_id": projectID,
		"reference":  reference,
	}
	path, file := C.GetCloudManager().GetListReferenceFileNameAndPathFromCloud(projectID, reference)
	reader, err := C.GetCloudManager().Get(path, file)
	if err != nil {
		log.WithFields(logFields).WithError(err).Error("list File Missing")
		return fmt.Errorf("List File Missing")
	}
	valuesInFile := make([]string, 0)
	data, err := io.ReadAll(reader)
	if err != nil {
		log.WithFields(logFields).WithError(err).Error("File reader failed")
		return fmt.Errorf("File reader failed")
	}
	err = json.Unmarshal(data, &valuesInFile)
	if err != nil {
		log.WithFields(logFields).WithError(err).Error("list data unmarshall failed")
		return fmt.Errorf("list data unmarshall failed")
	}
	cacheKeyList, err := model.GetListCacheKey(projectID, reference)
	if err != nil {
		log.WithFields(logFields).WithError(err).Error("get cache key failed")
		return fmt.Errorf("get cache key failed")
	}
	for _, value := range valuesInFile {
		err = cacheRedis.ZAddPersistent(cacheKeyList, strings.TrimSpace(value), 0)
		if err != nil {
			log.WithFields(logFields).WithError(err).Error("failed to add new values to sorted set")
			return fmt.Errorf("failed to add new values to sorted set")
		}
	}

	return nil
}
