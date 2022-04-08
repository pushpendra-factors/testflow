package main

import (
	"bytes"
	"flag"
	"fmt"
	"net/http"
	"strings"

	C "factors/config"
	"factors/filestore"
	"factors/model/model"
	"factors/model/store"
	BQ "factors/services/bigquery"
	serviceDisk "factors/services/disk"
	serviceGCS "factors/services/gcstorage"
	"factors/util"

	log "github.com/sirupsen/logrus"
)

var taskID = "Script#EventsArchival"
var pbLog = log.WithField("prefix", taskID)

func main() {
	envFlag := flag.String("env", "development", "Environment. Could be development|staging|production.")
	projectIDFlag := flag.Uint64("project_id", 0, "Project id to be run for.")
	bucketNameFlag := flag.String("bucket_name", "/usr/local/var/factors/cloud_storage", "Bucket name for production.")
	bigqueryProjectIDFlag := flag.String("bq_project_id", "", "Project id to be run for.")
	bigqueryDatasetFlag := flag.String("bq_dataset", "", "Dataset for the bigquery.")
	bigqueryCredentialsFileFlag := flag.String("bq_credentials_json", "", "Filename for credentials json. Must be present in bucket at bigquery/<projectID>/")

	dbHost := flag.String("db_host", C.PostgresDefaultDBParams.Host, "")
	dbPort := flag.Int("db_port", C.PostgresDefaultDBParams.Port, "")
	dbUser := flag.String("db_user", C.PostgresDefaultDBParams.User, "")
	dbName := flag.String("db_name", C.PostgresDefaultDBParams.Name, "")
	dbPass := flag.String("db_pass", C.PostgresDefaultDBParams.Password, "")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")

	flag.Parse()
	defer util.NotifyOnPanic(taskID, *envFlag)

	if *envFlag != "development" && *envFlag != "staging" && *envFlag != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *envFlag)
		panic(err)
	} else if *projectIDFlag == 0 {
		err := fmt.Errorf("Invalid project id %d", *projectIDFlag)
		panic(err)
	}

	pbLog.Info("Starting to initialize database.")
	appName := "script_push_to_bigquery"
	config := &C.Configuration{
		AppName: appName,
		Env:     *envFlag,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
			AppName:  appName,
		},
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
			Certificate: *memSQLCertificate,
			AppName:     appName,
		},
		PrimaryDatastore: *primaryDatastore,
	}
	C.InitConf(config)

	err := C.InitDB(*config)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize DB")
	}

	var cloudManager filestore.FileManager
	var fileDir string
	if *envFlag == "development" {
		cloudManager = serviceDisk.New(*bucketNameFlag)
		fileDir = fmt.Sprintf("%s/factors-bq-cred/%d/", *bucketNameFlag, *projectIDFlag)
	} else {
		cloudManager, err = serviceGCS.New(*bucketNameFlag)
		if err != nil {
			pbLog.WithError(err).Errorln("Failed to init New GCS Client")
			panic(err)
		}
		fileDir = fmt.Sprintf("factors-bq-cred/%d/", *projectIDFlag)
	}

	pbLog.Info("Checking for existing config for project_id.")
	bigquerySetting, status := store.GetStore().GetBigquerySettingByProjectID(*projectIDFlag)
	if status == http.StatusInternalServerError {
		log.WithError(err).Fatalf("Failed to get bigquery setting for project_id")
	} else if status == http.StatusNotFound {
		pbLog.Info("Creating a new entry in BigquerySettings from given credentials.")
		if *bigqueryCredentialsFileFlag == "" || *bigqueryDatasetFlag == "" ||
			*bigqueryProjectIDFlag == "" {
			pbLog.Fatalf("Bigquery configuration details missing for adding new project")
		}

		pbLog.Info("Enabling archival and biguqery in project_settings")
		if errCode := store.GetStore().EnableBigqueryArchivalForProject(*projectIDFlag); errCode != http.StatusAccepted {
			pbLog.Fatal("Error enabling archival and biguqery in project_settings")
		}

		credentialsReader, err := cloudManager.Get(fileDir, *bigqueryCredentialsFileFlag)
		if err != nil {
			pbLog.WithError(err).Fatal("Error reading credentials file")
		}
		credentialsBuffer := new(bytes.Buffer)
		credentialsBuffer.ReadFrom(credentialsReader)

		bigquerySetting = &model.BigquerySetting{
			ProjectID:               *projectIDFlag,
			BigqueryProjectID:       *bigqueryProjectIDFlag,
			BigqueryDatasetName:     *bigqueryDatasetFlag,
			BigqueryCredentialsJSON: strings.ReplaceAll(credentialsBuffer.String(), "\n", ""),
		}
		bigquerySetting, status = store.GetStore().CreateBigquerySetting(bigquerySetting)
		if status != http.StatusCreated {
			pbLog.Fatal("Failed to create bigquery setting")
		}
		pbLog.Infof("New config %s added for project id %d.", bigquerySetting.ID, *projectIDFlag)
	} else {
		pbLog.Infof("Existing config %s found.", bigquerySetting.ID)
	}

	err = BQ.CreateBigqueryArchivalTables(*projectIDFlag)
	if err != nil {
		pbLog.WithError(err).Error("Failed to create one or more tables in bigquery.")
	}
}
