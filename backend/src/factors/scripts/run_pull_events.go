package main

// Pull events that needs to be processed and write to file.
// Sample usage in terminal.
// export GOPATH=/Users/aravindmurthy/code/factors/backend/
// go run run_pull_events.go --project_id=1 --output_dir="" --end_time=""

import (
	"encoding/json"
	C "factors/config"
	P "factors/pattern"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

var projectIdFlag = flag.Int("project_id", 0, "Project Id.")
var outputDirFlag = flag.String("output_dir", "", "Results are written to output directory.")
var endTimeFlag = flag.String("end_time", time.Now().Format(time.RFC3339),
	"Events that occurred from  num_HOURS or max_events before end time are processed. Format is '2018-06-30T00:00:00Z'")

var startTime time.Time
var endTime time.Time

const max_EVENTS = 30000000 // 30 million. (million a day)
const num_HOURS = 24 * 30   // TODO(Update this to 30 days)

func pullAndWriteEventsFile(projectId int, startTime time.Time, endTime time.Time,
	baseOutputDir string) error {
	db := C.GetServices().Db

	defer db.Close()

	rows, err := db.Raw("SELECT events.user_id, events.event_name, events.created_at, events.count, events.properties, users.created_at FROM events "+
		"LEFT JOIN users ON events.user_id = users.id WHERE events.project_id = ? AND events.created_at BETWEEN  ? AND ? "+
		"ORDER BY events.user_id, events.created_at LIMIT ?",
		*projectIdFlag, startTime, endTime, max_EVENTS).Rows()
	defer rows.Close()

	if err != nil {
		log.WithFields(log.Fields{"err": err}).Fatal("SQL Query failed.")
		return err
	}
	filename := filepath.Join(baseOutputDir, "events.txt")
	file, err := os.Create(filename)
	if err != nil {
		log.WithFields(log.Fields{"file": filename, "err": err}).Fatal("Unable to create file.")
		return err
	}
	defer file.Close()

	for rows.Next() {
		var userId string
		var eventName string
		var eventCreatedAt time.Time
		var userCreatedAt time.Time
		var eventCardinality uint
		var eventProperties postgres.Jsonb
		if err = rows.Scan(&userId, &eventName, &eventCreatedAt, &eventCardinality, &eventProperties, &userCreatedAt); err != nil {
			log.WithFields(log.Fields{"err": err}).Error("SQL Parse failed.")
			return err
		}
		eventPropertiesBytes, err := eventProperties.Value()
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Fatal("Unable to unmarshal property.")
		}
		var eventPropertiesMap map[string]interface{}
		err = json.Unmarshal(eventPropertiesBytes.([]byte), &eventPropertiesMap)
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Fatal("Unable to unmarshal property.")
		}
		event := P.CounterEventFormat{
			UserId:           userId,
			UserCreatedTime:  userCreatedAt,
			EventName:        eventName,
			EventCreatedTime: eventCreatedAt,
			EventCardinality: eventCardinality,
			EventProperties:  eventPropertiesMap,
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
	}
	return nil
}

func setupOutputDirectory() (string, error) {
	dirName := fmt.Sprintf("patterns-%d-%d", *projectIdFlag, endTime.Unix())
	outputDirectory := filepath.Join(*outputDirFlag, dirName)
	err := os.Mkdir(outputDirectory, 0777)
	return outputDirectory, err
}

func main() {
	// Initialize configs and connections.
	err := C.Init()
	if err != nil {
		log.Error("Failed to initialize.")
		os.Exit(1)
	}

	if *projectIdFlag <= 0 || *outputDirFlag == "" {
		log.Error("project_id and output_dir are required.")
		os.Exit(1)
	}

	// Initialize start time and end time.
	endTime, err = time.Parse(time.RFC3339, *endTimeFlag)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Unable to parse time.")
		os.Exit(1)
	}
	startTime = endTime.Add(-num_HOURS * time.Hour)

	baseOutputDir, err := setupOutputDirectory()
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed to setup output directory.")
		os.Exit(1)
	}

	err = pullAndWriteEventsFile(*projectIdFlag, startTime, endTime, baseOutputDir)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed to mine patterns.")
		os.Exit(1)
	}
}
