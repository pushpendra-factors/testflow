package main

// Pull events that needs to be processed and write to file.
// Sample usage in terminal.
// export GOPATH=/Users/aravindmurthy/code/factors/backend/
// go run run_pull_events.go --project_id=1 --output_dir="" --end_time=""

import (
	"encoding/json"
	C "factors/config"
	P "factors/pattern"
	serviceDisk "factors/services/disk"
	"flag"
	"fmt"
	"os"
	"time"

	_ "github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

var projectIdFlag = flag.Int("project_id", 0, "Project Id.")
var localDiskTmpDirFlag = flag.String("local_disk_tmp_dir",
	"/tmp/factors/local_disk/tmp", "--local_disk_tmp_dir=/tmp/factors/local_disk/tmp pass directory")
var bucketNameFlag = flag.String("bucket_name", "/tmp/factors/cloud_storage", "--bucket_name=/tmp/factors/cloud_storage pass bucket name")
var endTimeFlag = flag.Int64("end_time", time.Now().Unix(),
	"Events that occurred from  num_HOURS or max_events before end time are processed. Format is unix timestamp")

const max_EVENTS = 30000000        // 30 million. (million a day)
const num_SECONDS = 30 * 24 * 3600 // TODO(Update this to 30 days)

func pullAndWriteEventsFile(projectId int, startTime int64, endTime int64,
	baseOutputDir string) error {
	db := C.GetServices().Db

	defer db.Close()

	rows, err := db.Raw("SELECT COALESCE(users.customer_user_id, users.id), event_names.name, events.timestamp, events.count,"+
		" events.properties, users.join_timestamp, user_properties.properties FROM events "+
		"LEFT JOIN event_names ON events.event_name_id = event_names.id LEFT JOIN users ON events.user_id = users.id "+
		"LEFT JOIN user_properties ON events.user_properties_id = user_properties.id "+
		"WHERE events.project_id = ? AND events.timestamp BETWEEN  ? AND ? "+
		"ORDER BY events.user_id, events.timestamp LIMIT ?", *projectIdFlag, startTime, endTime, max_EVENTS).Rows()
	defer rows.Close()

	if err != nil {
		log.WithFields(log.Fields{"err": err}).Fatal("SQL Query failed.")
		return err
	}

	file, err := os.Create(baseOutputDir)
	if err != nil {
		log.WithFields(log.Fields{"file": baseOutputDir, "err": err}).Fatal("Unable to create file.")
		return err
	}
	defer file.Close()

	rowCount := 0
	for rows.Next() {
		var userId string
		var eventName string
		var eventTimestamp int64
		var userJoinTimestamp int64
		var eventCardinality uint
		var eventProperties postgres.Jsonb
		var userProperties postgres.Jsonb
		if err = rows.Scan(&userId, &eventName, &eventTimestamp,
			&eventCardinality, &eventProperties, &userJoinTimestamp, &userProperties); err != nil {
			log.WithFields(log.Fields{"err": err}).Error("SQL Parse failed.")
			return err
		}
		eventPropertiesBytes, err := eventProperties.Value()
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Fatal("Unable to unmarshal event property.")
		}
		var eventPropertiesMap map[string]interface{}
		err = json.Unmarshal(eventPropertiesBytes.([]byte), &eventPropertiesMap)
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Fatal("Unable to unmarshal event property.")
		}
		userPropertiesBytes, err := userProperties.Value()
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Fatal("Unable to unmarshal user property.")
		}
		var userPropertiesMap map[string]interface{}
		err = json.Unmarshal(userPropertiesBytes.([]byte), &userPropertiesMap)
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Fatal("Unable to unmarshal user property.")
		}
		event := P.CounterEventFormat{
			UserId:            userId,
			UserJoinTimestamp: userJoinTimestamp,
			EventName:         eventName,
			EventTimestamp:    eventTimestamp,
			EventCardinality:  eventCardinality,
			EventProperties:   eventPropertiesMap,
			UserProperties:    userPropertiesMap,
		}

		lineBytes, err := json.Marshal(event)
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Fatal("Unable to unmarshal event.")
			return err
		}
		line := string(lineBytes)
		if _, err := file.WriteString(fmt.Sprintf("%s\n", line)); err != nil {
			log.WithFields(log.Fields{"line": line, "err": err}).Fatal("Unable to write to file.")
			return err
		}
		rowCount++
	}
	log.Infof("Events pulled : %d", rowCount)
	log.Infof("Output filepath : %s", baseOutputDir)
	return nil
}

func main() {
	// Initialize configs and connections.
	err := C.Init()
	if err != nil {
		log.Error("Failed to initialize.")
		os.Exit(1)
	}

	if *projectIdFlag <= 0 {
		log.Error("project_id flag is required.")
		os.Exit(1)
	}

	projectId := uint64(*projectIdFlag)
	modelId := uint64(time.Now().Unix())

	localDiskTmpDir := *localDiskTmpDirFlag
	bucketName := *bucketNameFlag

	diskManager := serviceDisk.New(localDiskTmpDir)
	cloundManager := serviceDisk.New(bucketName)

	err = os.MkdirAll(localDiskTmpDir, 0755)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed to create local disk tmp directory.")
		os.Exit(1)
	}

	if C.IsDevelopment() {
		err = os.MkdirAll(fmt.Sprintf("%s/projects/%v/models/%v/", bucketName, projectId, modelId), 0755)
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Error("Failed to Create projects model dir.")
			os.Exit(1)
		}
	}

	var startTime int64
	var endTime int64
	// Initialize start time and end time.
	if *endTimeFlag < 0 {
		log.WithFields(log.Fields{"err": err}).Error("Incorrect end time.")
		os.Exit(1)
	}
	endTime = *endTimeFlag
	startTime = endTime - num_SECONDS

	// First write file to local disk tmp directory.
	_, fName := diskManager.GetModelEventsFilePathAndName(projectId, modelId)
	tmpOutputFilePath := fmt.Sprintf("%s/%s", localDiskTmpDir, fName)
	err = pullAndWriteEventsFile(*projectIdFlag, startTime, endTime, tmpOutputFilePath)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed to mine patterns.")
		os.Exit(1)
	}
	tmpOutputFile, err := os.Open(tmpOutputFilePath)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed to open file.")
		os.Exit(1)
	}

	cDir, cName := cloundManager.GetModelEventsFilePathAndName(projectId, modelId)
	err = cloundManager.Create(cDir, cName, tmpOutputFile)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed to upload file to cloud")
		os.Exit(1)
	}

	log.WithFields(log.Fields{
		"ProjectId": *projectIdFlag,
		"ModelId":   modelId,
	}).Info("Project Model Information")
}
