package main

import (
	"bufio"
	"errors"
	C "factors/config"
	"factors/filestore"
	"factors/model/model"
	"factors/model/store"
	serviceDisk "factors/services/disk"
	serviceGCS "factors/services/gcstorage"
	"factors/util"
	U "factors/util"
	"flag"
	"fmt"
	"net/http"

	encjson "encoding/json"

	log "github.com/sirupsen/logrus"
)

type ProjectData struct {
	ID             uint64   `json:"pid"`
	ModelID        uint64   `json:"mid"`
	ModelType      string   `json:"mt"`
	StartTimestamp int64    `json:"st"`
	EndTimestamp   int64    `json:"et"`
	Chunks         []string `json:"cs"`
}

func main() {
	envFlag := flag.String("env", "development", "")
	bucketName := flag.String("bucket_name_v2", "/usr/local/var/factors/cloud_storage", "")
	metadataVersion := flag.String("metadata_version", "1610691192", "Metadata version to read from")

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
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypePostgres, "Primary datastore type as memsql or postgres")
	flag.Parse()

	if *envFlag != "development" &&
		*envFlag != "staging" &&
		*envFlag != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *envFlag)
		panic(err)
	}

	config := &C.Configuration{
		Env: *envFlag,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
		},
		MemSQLInfo: C.DBConf{
			Host:     *memSQLHost,
			Port:     *memSQLPort,
			User:     *memSQLUser,
			Name:     *memSQLName,
			Password: *memSQLPass,
		},
		PrimaryDatastore: *primaryDatastore,
	}
	C.InitConf(config)

	err := C.InitDBWithMaxIdleAndMaxOpenConn(*config, 50, 50)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize db in add session.")
	}
	defer util.NotifyOnPanic("Task#MigrateModel", *envFlag)

	if *metadataVersion == "" {
		panic(errors.New("metadata version not given"))
	}

	var cloudManager filestore.FileManager
	if *envFlag == "development" {
		cloudManager = serviceDisk.New(*bucketName)
	} else {
		cloudManager, err = serviceGCS.New(*bucketName)
		if err != nil {
			log.WithError(err).Errorln("Failed to init New GCS Client V1")
			panic(err)
		}
	}

	path, name := cloudManager.GetProjectsDataFilePathAndName(*metadataVersion)
	log.WithFields(log.Fields{"path": path,
		"name": name}).Info("Getting project model meta using cloud manager.")

	versionFile, err := cloudManager.Get(path, name)
	if err != nil {
		log.WithFields(log.Fields{"path": path, "name": name,
			"err": err}).Error("Failed to read current version file.")
		panic(err)
	}

	scanner := bufio.NewScanner(versionFile)
	// Adjust scanner buffer capacity to 10MB per line.
	const maxCapacity = 10 * 1024 * 1024
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	projectDatas := make([]ProjectData, 0, 0)

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		var p ProjectData
		if err := encjson.Unmarshal([]byte(line), &p); err != nil {
			log.WithFields(log.Fields{"line_num": lineNum,
				"err": err}).Error("Unmarshal error. Failed to ParseProjectsDataFile.")
			continue
		}
		projectDatas = append(projectDatas, p)
	}
	err = scanner.Err()
	if err != nil {
		log.WithError(err).Errorln("Scanner error. Failed to ParseProjectsDataFile.")
	}

	success, failure := 0, 0
	for _, metadata := range projectDatas {
		chunkIdsString := ""
		for _, chunkId := range metadata.Chunks {
			if chunkIdsString != "" {
				chunkIdsString += ","
			}
			chunkIdsString += chunkId
		}
		errCode, msg := store.GetStore().CreateProjectModelMetadata(&model.ProjectModelMetadata{
			ProjectId: metadata.ID,
			ModelId:   metadata.ModelID,
			ModelType: metadata.ModelType,
			StartTime: metadata.StartTimestamp,
			EndTime:   metadata.EndTimestamp,
			Chunks:    chunkIdsString,
			CreatedAt: U.TimeNow(),
			UpdatedAt: U.TimeNow(),
		})
		if errCode != http.StatusCreated {
			failure++
			fmt.Println(msg)
		} else {
			success++
		}
	}
	fmt.Println("Success:", success, "Failure:", failure)
}
