package main

import (
	C "factors/config"
	"factors/model/store"
	"flag"
	"os"
	"time"

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

	queueRedisHost := flag.String("queue_redis_host", "localhost", "")
	queueRedisPort := flag.Int("queue_redis_port", 6379, "")

	clearbitProvisionAccountAPIKey := flag.String("clearbit_provision_account_api_key", "dummy", "")

	useQueueRedis := flag.Bool("use_queue_redis", false, "Use queue redis for sdk related caching.")
	projectId := flag.String("project_id", "", "Project Id for which clearbit account will be provisoned")
	emailId := flag.String("email_id", "", "Email id required for provision account")
	domain := flag.String("domain", "", "domain name of the project")

	appName := "clearbit_acccount_provisioning"

	config := &C.Configuration{
		AppName: appName,
		Env:     *envFlag,
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
		PrimaryDatastore:               *primaryDatastore,
		UseQueueRedis:                  *useQueueRedis,
		QueueRedisHost:                 *queueRedisHost,
		QueueRedisPort:                 *queueRedisPort,
		ClearbitProvisionAccountAPIKey: *clearbitProvisionAccountAPIKey,
	}
	defaultHealthcheckPingID := C.HealthCheckClearbitAccountProvisioningJobPingID
	C.InitConf(config)
	C.InitQueueRedis(config.QueueRedisHost, config.QueueRedisPort)

	log.Info("Starting to initialize database.")
	err := C.InitDB(*config)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize DB")
		os.Exit(1)
	}

	defer C.WaitAndFlushAllCollectors(65 * time.Second)

	projectIdList := C.GetTokensFromStringListAsUint64(*projectId)
	emailList := C.GetTokensFromStringListAsString(*emailId)
	domainList := C.GetTokensFromStringListAsString(*domain)
	if len(projectIdList) == 0 || !(len(projectIdList) == len(emailList) && len(emailList) == len(domainList)) {
		log.Panic("ProjectId, Email and Domain List is not valid.")
	}

	jobReport := store.GetStore().ProvisionClearbitAccount(projectIdList, emailList, domainList)
	if len(jobReport) > 0 {
		C.PingHealthcheckForFailure(defaultHealthcheckPingID, jobReport)
	} else {
		C.PingHealthcheckForSuccess(defaultHealthcheckPingID, jobReport)
	}

}
