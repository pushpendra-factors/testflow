package main

import (
	"encoding/json"
	C "factors/config"
	"factors/model/model"
	"factors/model/store"
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

func main() {

	env := flag.String("env", C.DEVELOPMENT, "")
	isPSCHost := flag.Int("memsql_is_psc_host", C.MemSQLDefaultDBParams.IsPSCHost, "")
	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")

	queueRedisHost := flag.String("queue_redis_host", "localhost", "")
	queueRedisPort := flag.Int("queue_redis_port", 6379, "")
	useQueueRedis := flag.Bool("use_queue_redis", false, "Use queue redis for sdk related caching.")

	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")

	clearbitAccProvisionKey := flag.String("cb_acc_provision_key", "dummy", "")

	flag.Parse()

	if *env != C.DEVELOPMENT &&
		*env != C.STAGING &&
		*env != C.PRODUCTION {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}

	config := &C.Configuration{
		AppName: "free_plan_6Signal_to_clearbit_migration",
		Env:     *env,
		MemSQLInfo: C.DBConf{
			Certificate: *memSQLCertificate,
			IsPSCHost:   *isPSCHost,
			Host:        *memSQLHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
		},
		PrimaryDatastore:               *primaryDatastore,
		QueueRedisHost:                 *queueRedisHost,
		QueueRedisPort:                 *queueRedisPort,
		UseQueueRedis:                  *useQueueRedis,
		RedisHost:                      *redisHost,
		RedisPort:                      *redisPort,
		ClearbitProvisionAccountAPIKey: *clearbitAccProvisionKey,
	}

	C.InitConf(config)
	C.InitRedis(config.RedisHost, config.RedisPort)
	C.InitQueueRedis(config.QueueRedisHost, config.QueueRedisPort)
	err := C.InitDB(*config)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize db.")
	}

	defer C.WaitAndFlushAllCollectors(65 * time.Second)

	// Fetch all the projects on free plan
	projectIds, errCode, _, _ := store.GetStore().GetAllProjectIdsUsingPlanId(model.PLAN_ID_FREE)
	if errCode != http.StatusFound {
		return
	}

	count := 0
	failureMap := make(map[string][]int64)
	for _, projectId := range projectIds {

		// Check if they are active after 16 Feb, 2024 00:00:00 IST
		errCode, errMsg := store.GetStore().IsEventPresentAfterGivenTimestamp(projectId, 1708021800)
		if errCode != http.StatusFound {
			log.Warn(errMsg)
			failureMap[errMsg] = append(failureMap[errMsg], projectId)
			continue
		}

		errCode, errMsg = UpdateClearbitAsDeanonProvider(projectId)
		if errCode != http.StatusOK {
			log.Error(errMsg)
			failureMap[errMsg] = append(failureMap[errMsg], projectId)
			continue
		}
		count++
	}

	if count < len(projectIds) {
		log.Info("List of failed project ids: ", failureMap)
	}
}

func UpdateClearbitAsDeanonProvider(projectId int64) (int, string) {

	errCode, errMsg := store.GetStore().ProvisionClearbitAccountByAdminEmailAndDomain(projectId)
	if errCode != http.StatusOK {
		return errCode, errMsg
	}

	factorsDeanonConfig, err := json.Marshal(model.FactorsDeanonConfig{Clearbit: model.DeanonVendorConfig{TrafficFraction: 1.0}, SixSignal: model.DeanonVendorConfig{TrafficFraction: 0.0}})
	if err != nil {
		return http.StatusInternalServerError, "Failed Json Marshal of Deanon Config"
	}
	factorsDeanonConfigJson := postgres.Jsonb{RawMessage: factorsDeanonConfig}
	_, errCode = store.GetStore().UpdateProjectSettings(projectId, &model.ProjectSetting{FactorsDeanonConfig: &factorsDeanonConfigJson})
	if errCode != http.StatusAccepted {
		return errCode, "Failed update project settings"
	}

	return http.StatusOK, "Clearbit successfully setup as deanonymisation provider"

}
