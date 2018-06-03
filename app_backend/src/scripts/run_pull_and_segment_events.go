package main

// Example usage on Terminal. Process 24 hours of events ending at end time.
// export GOPATH=/Users/aravindmurthy/code/autometa/app_backend/
// go run run_pull_and_segment_events.go --project_id=1 --end_time=""--output_dir=""

import (
	C "config"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

var endTimeFlag = flag.String("end_time", time.Now().Format(time.RFC3339),
	"Events that occurred from 24 hours before end time are processed. Format is '2018-06-30T00:00:00Z'")
var projectIdFlag = flag.Int("project_id", 0, "Project Id.")
var outputDirFlag = flag.String("output_dir", "", "Results are written to output directory.")
var startTime time.Time
var endTime time.Time

func writeToEventFile(baseOutputDir string, userId string, eventName string,
	eventCreatedAt time.Time, userCreatedAt time.Time) {

	segment := int(eventCreatedAt.Sub(userCreatedAt).Seconds()) / 86400

	if segment < 0 {
		log.WithFields(
			log.Fields{"user_id": userId, "eventName": eventName,
				"e_created_at": eventCreatedAt.Format(time.RFC3339),
				"u_created_at": userCreatedAt.Format(time.RFC3339)}).Error("Event Created before user!")
		return
	}

	filename := filepath.Join(baseOutputDir, fmt.Sprintf("day_%d_events.txt", segment))
	// If the file doesn't exist, create it, or append to the file
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		log.WithFields(log.Fields{"file": filename, "err": err}).Fatal("Unable to open / create file.")
	}
	defer file.Close()

	var line string = fmt.Sprintf("%s,%s,%s,%s\n", userId, eventName,
		eventCreatedAt.Format(time.RFC3339), userCreatedAt.Format(time.RFC3339))
	if _, err := file.WriteString(line); err != nil {
		log.WithFields(log.Fields{"segment": segment, "line": line, "err": err}).Fatal("Unable to write to file.")
	}
}

func pullAndSegmentEvents(baseOutputDir string) {
	db := C.GetServices().Db

	defer db.Close()

	rows, err := db.Raw("SELECT events.user_id, events.event_name, events.created_at, users.created_at FROM events "+
		"LEFT JOIN users ON events.user_id = users.id WHERE events.project_id = ? AND events.created_at BETWEEN  ? AND ?",
		*projectIdFlag, startTime, endTime).Rows()
	defer rows.Close()

	if err != nil {
		log.WithFields(log.Fields{"err": err}).Fatal("SQL Query failed.")
		return
	}
	for rows.Next() {
		var userId string
		var eventName string
		var eventCreatedAt time.Time
		var userCreatedAt time.Time
		if err = rows.Scan(&userId, &eventName, &eventCreatedAt, &userCreatedAt); err != nil {
			log.WithFields(log.Fields{"err": err}).Error("SQL Parse failed.")
		}
		writeToEventFile(baseOutputDir, userId, eventName, eventCreatedAt, userCreatedAt)
	}
}

func setupOutputDirectory() (string, error) {
	dirName := fmt.Sprintf("meta-%d-%d-%d", *projectIdFlag, startTime.Unix(), endTime.Unix())
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
		log.Error("ProjectId and OutputDir is required.")
		os.Exit(1)
	}

	// Initialize start time and end time.
	endTime, err = time.Parse(time.RFC3339, *endTimeFlag)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Unable to parse time.")
		os.Exit(1)
	}
	startTime = endTime.Add(-24 * time.Hour)

	baseOutputDir, err := setupOutputDirectory()
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed to setup output directory")
		os.Exit(1)
	}

	pullAndSegmentEvents(baseOutputDir)
}
