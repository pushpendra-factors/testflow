package main

// Mine TOP_K Frequent patterns.
// If data is visualized in below format, where U(.) are users, E(.) are events,
// T(.) are timestamps.

// U1: E1(T1), E4(T2), E1(T3), E5(T4), E1(T5), E5(T6)
// U2: E3(T7), E4(T8), E5(T9), E1(T10)
// U3: E2(T11), E1(T12), E5(T13)

// The frequency of the event E1 -> E5 is 3 - twice non overlapping
// in U1 and once in U3 - i.e. [U1: E1(T1) -> E5(T4)] [U1: E1(T5) -> E5(T6)] and
// [U3: E1(T12) -> E5(T13)].
// Further the distribution of timestamps, event properties and number of occurrences
// are stored with the patterns.

// Sample usage in terminal.
// export GOPATH=/Users/aravindmurthy/code/autometa/app_backend/
// go run run_pattern_mine.go --project_id=1 --input_file=""

import (
	"bufio"
	C "config"
	"flag"
	"fmt"
	M "model"
	"os"
	P "pattern"

	_ "github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

var projectIdFlag = flag.Int("project_id", 0, "Project Id.")
var inputFileFlag = flag.String("input_file", "",
	"Input file of format user_id,user_creation_time,event_id,event_creation_time sorted by user_id and event_creation_time")

const top_K = 20

func countPatterns(filepath string, patterns []*P.Pattern) {
	file, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	P.CountPatterns(scanner, patterns)
}

func minePatterns(projectId int, filepath string) ([]*P.Pattern, error) {
	eventNames, err_code := M.GetEventNames(uint64(projectId))
	if err_code != M.DB_SUCCESS {
		return nil, fmt.Errorf("DB read of event names failed")
	}
	var patterns []*P.Pattern
	for _, eventName := range eventNames {
		p, err := P.NewPattern([]string{eventName.Name})
		if err != nil {
			return nil, fmt.Errorf("Pattern initialization failed")
		}
		patterns = append(patterns, p)
	}
	countPatterns(filepath, patterns)
	return patterns, nil
}

func main() {
	// Initialize configs and connections.
	err := C.Init()
	if err != nil {
		log.Error("Failed to initialize.")
		os.Exit(1)
	}

	if *projectIdFlag <= 0 || *inputFileFlag == "" {
		log.Error("project_id and input_file are required.")
		os.Exit(1)
	}

	patterns, err := minePatterns(*projectIdFlag, *inputFileFlag)
	if err != nil {
		log.Error("Failed to mine patterns.")
		os.Exit(1)
	}

	for _, p := range patterns {
		fmt.Printf(fmt.Sprintf("%v", *p))
	}
}
