package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	C "factors/config"
	"factors/model/model"
	"factors/model/store"
	"factors/util"

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

	useQueueRedis := flag.Bool("use_queue_redis", false, "Use queue redis for sdk related caching.")

	projectIDsFlag := flag.String("project_ids", "*", "List of project_id to run for.")
	disabledProjectIDsFlag := flag.String("disabled_project_ids", "", "List of project_ids to exclude.")
	flag.Parse()

	model.SetSmartPropertiesReservedNames()

	if *envFlag != "development" && *envFlag != "staging" && *envFlag != "production" {
		panic(fmt.Errorf("env [ %s ] not recognised", *envFlag))
	}

	allProjects, projectIdsToRun, _ := C.GetProjectsFromListWithAllProjectSupport(*projectIDsFlag, *disabledProjectIDsFlag)

	log.Info("Starting to initialize database.")
	appName := "country_name_to_iso_code_migration"

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
		PrimaryDatastore: *primaryDatastore,
		UseQueueRedis:    *useQueueRedis,
		QueueRedisHost:   *queueRedisHost,
		QueueRedisPort:   *queueRedisPort,
	}
	C.InitConf(config)
	C.InitQueueRedis(config.QueueRedisHost, config.QueueRedisPort)

	err := C.InitDB(*config)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize DB")
		os.Exit(1)
	}

	db := C.GetServices().Db
	defer db.Close()
	defer C.WaitAndFlushAllCollectors(65 * time.Second)

	projectSettings := []model.ProjectSetting{}
	if allProjects {
		err = db.Table("project_settings").Where("six_signal_config IS NOT NULL AND six_signal_config !='' AND six_signal_config !='{}'").Find(&projectSettings).Error
		if err != nil {
			log.Warn("Failed during query fetch")
			os.Exit(0)
		}
	} else {
		projectIdsArray := make([]int64, 0)
		for projectId, _ := range projectIdsToRun {
			projectIdsArray = append(projectIdsArray, projectId)
		}
		err = db.Table("project_settings").Where("six_signal_config IS NOT NULL AND six_signal_config !='' AND six_signal_config !='{}' AND project_id IN (?)", projectIdsArray).Find(&projectSettings).Error
		if err != nil {
			log.Warn("Failed during query fetch")
			os.Exit(0)
		}
	}

	errToProjectIdMap := make(map[string][]int64)

	count := 0
	for _, projectSetting := range projectSettings {

		var sixSignalConfig model.SixSignalConfig
		json.Unmarshal(projectSetting.SixSignalConfig.RawMessage, &sixSignalConfig)

		countryInclude := sixSignalConfig.CountryInclude
		countryExclude := sixSignalConfig.CountryExclude
		var failure bool

		if len(countryInclude) > 0 {
			sixSignalConfig.CountryInclude, failure = model.ReplaceCountryNameWithIsoCodeInSixSignalConfig(countryInclude)
			if failure {
				errMsg := "Failures during conversion of country name in country include"
				errToProjectIdMap[errMsg] = append(errToProjectIdMap[errMsg], projectSetting.ProjectId)
			}
		}

		if len(countryExclude) > 0 {
			sixSignalConfig.CountryExclude, failure = model.ReplaceCountryNameWithIsoCodeInSixSignalConfig(countryExclude)
			if failure {
				errMsg := "Failures during conversion of country name in country exclude"
				errToProjectIdMap[errMsg] = append(errToProjectIdMap[errMsg], projectSetting.ProjectId)
			}
		}

		sixSignalConfigJson, err := util.EncodeStructTypeToPostgresJsonb(sixSignalConfig)
		if err != nil {
			//log.WithFields(log.Fields{"sixSignalConfig": sixSignalConfig, "project_id": projectSetting.ProjectId}).Warn("failed in EncodeStructTypeToPostgresJsonb. Hence skipping.")
			errMsg := "failed in EncodeStructTypeToPostgresJsonb. Hence skipping."
			errToProjectIdMap[errMsg] = append(errToProjectIdMap[errMsg], projectSetting.ProjectId)
			continue
		}

		_, errCode := store.GetStore().UpdateProjectSettings(projectSetting.ProjectId, &model.ProjectSetting{SixSignalConfig: sixSignalConfigJson})
		if errCode != http.StatusAccepted {
			//log.WithFields(log.Fields{"project_id": projectSetting.ProjectId}).Warn("Failed to update project settings.")
			errMsg := "Failed to update project settings"
			errToProjectIdMap[errMsg] = append(errToProjectIdMap[errMsg], projectSetting.ProjectId)
			continue
		}

		count++

	}
	log.WithField("count", count).Warn("Completed with count")
	log.WithField("errMap: ", errToProjectIdMap).Info("Error map for country to iso code migration")

}
