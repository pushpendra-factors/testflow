package main

// Mine TOP_K Frequent patterns of format start_event -> * -> * ... -> end_event.
// Patterns are mined from past 30 days of events from end_time of a given project or
// upto MAX_EVENTS FROM end_time whichever is lesser.
// If data is visualized in below format, where U(.) are users, E(.) are events,
// T(.) are timestamps.

// U1: E1(T1), E4(T2), E1(T3), E5(T4), E1(T5), E5(T6)
// U2: E3(T7), E4(T8), E5(T9), E1(T10)
// U3: E2(T11), E1(T12), E5(T13)

// The frequency of the event E1 -> E5 is 3 - twice non overlapping
// in U1 and once in U3 - i.e. [U1: E1(T3) -> E5(T4)] [U1: E1(T5) -> E5(T6)] and
// [U3: E1(T12) -> E5(T13)].
// Further the distribution of timestamps, event properties and number of occurrences
// are stored with the patterns.

// Sample usage in terminal.
// export GOPATH=/Users/aravindmurthy/code/autometa/app_backend/
// go run run_pattern_mine.go --project_id=1 --start_event="" --end_event="" --output_dir="" --end_time=""

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

var startEventFlag = flag.String("start_event", "", "Start event.")
var endEventFlag = flag.String("end_event", "", "End event.")
var projectIdFlag = flag.Int("project_id", 0, "Project Id.")
var outputDirFlag = flag.String("output_dir", "", "Results are written to output directory.")
var endTimeFlag = flag.String("end_time", time.Now().Format(time.RFC3339),
	"Events that occurred from  30 days or max_events before end time are processed. Format is '2018-06-30T00:00:00Z'")

var startTime time.Time
var endTime time.Time

const top_K = 20
const max_EVENTS = 30000000 // 30 million. (million a day)

type Pattern struct {
	eventNames []string,
	
}

type Patterns struct {
}

func minePatterns(projectId int, startEvent string, endEvent string,
	startTime time.Time, endTime time.Time) (*Patterns, error) {

}

func setupOutputDirectory() (string, error) {
	dirName := fmt.Sprintf("patterns-%d-%d", *projectIdFlag, endTime.Format())
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

	if *projectIdFlag <= 0 || *outputDirFlag == "" || *startEventFlag == "" || *endEventFlag == "" {
		log.Error("project_id, output_dir, start_event and end_event are required.")
		os.Exit(1)
	}

	// Initialize start time and end time.
	endTime, err = time.Parse(time.RFC3339, *endTimeFlag)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Unable to parse time.")
		os.Exit(1)
	}
	startTime = endTime.Add(-24 * time.Hour)

	patterns, err = minePatterns(*projectIdFlag, *startEventFlag,
		*endEventFlag, startTime, endTime)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed to mine patterns.")
		os.Exit(1)
	}

	baseOutputDir, err := setupOutputDirectory()
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed to setup output directory.")
		os.Exit(1)
	}
	if err = writePatterns(baseOutputDir, patterns); err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed to write patterns.")
		os.Exit(1)
	}
}
