package main

import (
	"factors/filestore"
	"factors/model/store"
	serviceDisk "factors/services/disk"
	serviceGCS "factors/services/gcstorage"
	"factors/util"
	"flag"
	"fmt"
	"net/http"

	C "factors/config"

	T "factors/task"

	taskWrapper "factors/task/task_wrapper"

	_ "github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

func main() {
	mainRunNumericalBucketing()
}

func mainRunNumericalBucketing() {

	projectIdFlag := flag.String("project_ids", "",
		"Optional: Project Id. A comma separated list of project Ids and supports '*' for all projects. ex: 1,2,6,9")

	envFlag := flag.String("env", "development", "")
	localDiskTmpDirFlag := flag.String("local_disk_tmp_dir", "/usr/local/var/factors/local_disk/tmp", "--local_disk_tmp_dir=/usr/local/var/factors/local_disk/tmp pass directory")
	bucketName := flag.String("bucket_name", "/usr/local/var/factors/cloud_storage", "")

	dbHost := flag.String("db_host", "localhost", "")
	dbPort := flag.Int("db_port", 5432, "")
	dbUser := flag.String("db_user", "autometa", "")
	dbName := flag.String("db_name", "autometa", "")
	dbPass := flag.String("db_pass", "@ut0me7a", "")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")

	isWeeklyEnabled := flag.Bool("weekly_enabled", false, "")
	lookback := flag.Int("lookback", 30, "lookback_for_delta lookup")

	enablePropertyTypeFromDB := flag.Bool("enable_property_type_from_db", false, "Enable property type check from db.")
	whitelistedProjectIDPropertyTypeFromDB := flag.String("whitelisted_project_ids_property_type_check_from_db", "", "Allowed project id for property type check from db.")
	blacklistedProjectIDPropertyTypeFromDB := flag.String("blacklisted_project_ids_property_type_check_from_db", "", "Blocked project id for property type check from db.")
	propertiesTypeCacheSize := flag.Int("property_details_cache_size", 0, "Cache size for in memory property detail.")

	flag.Parse()

	if *envFlag != "development" &&
		*envFlag != "staging" &&
		*envFlag != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *envFlag)
		panic(err)
	}

	defer util.NotifyOnPanic("Task#compute_numerical_bucketing", *envFlag)

	// init Config and DB.
	config := &C.Configuration{
		AppName: "compute_numerical_bucketing",
		Env:     *envFlag,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
		},
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
			Certificate: *memSQLCertificate,
			AppName:     "compute_numerical_bucketing",
		},
		PrimaryDatastore: *primaryDatastore,
	}
	C.InitConf(config)
	err := C.InitDB(*config)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize DB")
	}

	C.InitPropertiesTypeCache(*enablePropertyTypeFromDB, *propertiesTypeCacheSize, *whitelistedProjectIDPropertyTypeFromDB, *blacklistedProjectIDPropertyTypeFromDB)

	var cloudManager filestore.FileManager
	if *envFlag == "development" {
		cloudManager = serviceDisk.New(*bucketName)
	} else {
		log.Info("initializing cloud bucket")
		cloudManager, err = serviceGCS.New(*bucketName)
		if err != nil {
			log.WithError(err).Errorln("Failed to init New GCS Client")
			panic(err)
		}
	}
	diskManager := serviceDisk.New(*localDiskTmpDirFlag)

	allProjects, projectIdsToRun, _ := C.GetProjectsFromListWithAllProjectSupport(*projectIdFlag, "")
	if allProjects {
		projectIDs, errCode := store.GetStore().GetAllProjectIDs()
		if errCode != http.StatusFound {
			log.Fatal("Failed to get all projects and project_ids set to '*'.")
		}

		projectIdsToRun = make(map[int64]bool, 0)
		for _, projectID := range projectIDs {
			projectIdsToRun[projectID] = true
		}
	}

	projectIdsArray := make([]int64, 0)
	for projectId, _ := range projectIdsToRun {
		projectIdsArray = append(projectIdsArray, projectId)
	}

	configs := make(map[string]interface{})
	configs["diskManager"] = diskManager
	configs["cloudManager"] = &cloudManager

	if *isWeeklyEnabled {
		configs["modelType"] = T.ModelTypeWeek
		status := taskWrapper.TaskFuncWithProjectId("NumericalBucketingWeekly", *lookback, projectIdsArray, T.NumericalBucketing, configs)
		log.Info(status)
	}
}
