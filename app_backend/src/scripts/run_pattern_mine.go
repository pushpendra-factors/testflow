package main

// Mine TOP_K Frequent patterns for every event combination (segment) at every iteration.

// Sample usage in terminal.
// export GOPATH=/Users/aravindmurthy/code/factors/app_backend/
// go run run_pattern_mine.go --project_id=1 --input_file="" --output_file=""

import (
	"bufio"
	C "config"
	"encoding/json"
	"flag"
	"fmt"
	M "model"
	"os"
	P "pattern"
	"sort"

	_ "github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

var projectIdFlag = flag.Int("project_id", 0, "Project Id.")
var inputFileFlag = flag.String("input_file", "",
	"Input file of format user_id,user_creation_time,event_id,event_creation_time sorted by user_id and event_creation_time")
var outputFileFlag = flag.String("output_file", "", "All patterns written to file with each line a JSON")

// The number of patterns generated is bounded to max_SEGMENTS * top_K per iteration.
// The amount of data and the time computed to generate this data is bounded
// by these constants.
const max_SEGMENTS = 100000
const top_K = 5

func countPatterns(filepath string, patterns []*P.Pattern) {
	file, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	P.CountPatterns(scanner, patterns)
}

// Removes all patterns with zero counts.
func filterPatterns(patterns []*P.Pattern) []*P.Pattern {
	filteredPatterns := []*P.Pattern{}
	for _, p := range patterns {
		if p.Count > 0 {
			filteredPatterns = append(filteredPatterns, p)
		}
	}
	return filteredPatterns
}

func genSegmentedCandidates(patterns []*P.Pattern) map[string][]*P.Pattern {
	pSegments := make(map[string][]*P.Pattern)
	for _, p := range patterns {
		segmentKey := fmt.Sprintf("%s,%s", p.EventNames[0], p.EventNames[len(p.EventNames)-1])
		if _, ok := pSegments[segmentKey]; !ok {
			pSegments[segmentKey] = []*P.Pattern{}
		}
		pSegments[segmentKey] = append(pSegments[segmentKey], p)
	}

	candidateSegments := make(map[string][]*P.Pattern)
	for k, patterns := range pSegments {
		cPatterns, _, err := P.GenCandidates(patterns, top_K)
		if err != nil {
			log.Fatal(err)
		}
		candidateSegments[k] = cPatterns
	}
	return candidateSegments
}

func genLenThreeSegmentedCandidates(lenTwoPatterns []*P.Pattern) map[string][]*P.Pattern {
	startPatternsMap := make(map[string][]*P.Pattern)
	endPatternsMap := make(map[string][]*P.Pattern)

	for _, p := range lenTwoPatterns {
		pEvents := p.EventNames
		if len(pEvents) != 2 {
			log.Fatal(fmt.Sprintf("Pattern %s is not of length two.", p.EventNames))
		}
		startEvent := pEvents[0]
		endEvent := pEvents[1]

		if _, ok := startPatternsMap[startEvent]; !ok {
			startPatternsMap[startEvent] = []*P.Pattern{}
		}
		startPatternsMap[startEvent] = append(startPatternsMap[startEvent], p)

		if _, ok := endPatternsMap[endEvent]; !ok {
			endPatternsMap[endEvent] = []*P.Pattern{}
		}
		endPatternsMap[endEvent] = append(endPatternsMap[endEvent], p)
	}

	// Sort the patterns in descending order.
	for _, patterns := range startPatternsMap {
		sort.Slice(patterns,
			func(i, j int) bool {
				return patterns[i].Count > patterns[j].Count
			})
	}
	for _, patterns := range endPatternsMap {
		sort.Slice(patterns,
			func(i, j int) bool {
				return patterns[i].Count > patterns[j].Count
			})
	}

	segmentedCandidates := make(map[string][]*P.Pattern)
	for _, p := range lenTwoPatterns {
		startPatterns, ok1 := startPatternsMap[p.EventNames[0]]
		endPatterns, ok2 := endPatternsMap[p.EventNames[1]]
		if startPatterns == nil || endPatterns == nil || !ok1 || !ok2 {
			continue
		}
		lenThreeCandidates, err := P.GenLenThreeCandidatePatterns(
			p, startPatterns, endPatterns, top_K)
		if err != nil {
			log.Fatal(err)
		}
		segmentedCandidates[p.String()] = lenThreeCandidates
	}
	return segmentedCandidates
}

func minePatterns(projectId int, filepath string) ([][]*P.Pattern, error) {
	allCountedPatterns := [][]*P.Pattern{}

	// Length One Patterns.
	eventNames, err_code := M.GetEventNames(uint64(projectId))
	if err_code != M.DB_SUCCESS {
		return nil, fmt.Errorf("DB read of event names failed")
	}
	var lenOnePatterns []*P.Pattern
	for _, eventName := range eventNames {
		p, err := P.NewPattern([]string{eventName.Name})
		if err != nil {
			return nil, fmt.Errorf("Pattern initialization failed")
		}
		lenOnePatterns = append(lenOnePatterns, p)
	}
	countPatterns(filepath, lenOnePatterns)
	filteredLenOnePatterns := filterPatterns(lenOnePatterns)
	allCountedPatterns = append(allCountedPatterns, filteredLenOnePatterns)
	iter := 1
	printFilteredPatterns(filteredLenOnePatterns, iter)

	// Each event combination is a segment in itself.
	lenTwoPatterns, _, err := P.GenCandidates(filteredLenOnePatterns, max_SEGMENTS)
	if err != nil {
		log.Fatal(err)
	}
	countPatterns(filepath, lenTwoPatterns)
	filteredLenTwoPatterns := filterPatterns(lenTwoPatterns)
	allCountedPatterns = append(allCountedPatterns, filteredLenTwoPatterns)
	iter++
	printFilteredPatterns(filteredLenTwoPatterns, iter)

	lenThreeSegmentedPatterns := genLenThreeSegmentedCandidates(filteredLenTwoPatterns)
	lenThreePatterns := []*P.Pattern{}
	for _, patterns := range lenThreeSegmentedPatterns {
		lenThreePatterns = append(lenThreePatterns, patterns...)
	}
	countPatterns(filepath, lenThreePatterns)
	filteredLenThreePatterns := filterPatterns(lenThreePatterns)
	allCountedPatterns = append(allCountedPatterns, filteredLenThreePatterns)

	filteredPatterns := filteredLenThreePatterns
	var candidatePatternsMap map[string][]*P.Pattern
	var candidatePatterns []*P.Pattern
	for len(filteredPatterns) > 0 {
		iter++
		printFilteredPatterns(filteredPatterns, iter)

		candidatePatternsMap = genSegmentedCandidates(filteredPatterns)
		candidatePatterns = []*P.Pattern{}
		for _, patterns := range candidatePatternsMap {
			candidatePatterns = append(candidatePatterns, patterns...)
		}
		countPatterns(filepath, candidatePatterns)
		filteredPatterns = filterPatterns(candidatePatterns)
		if len(filteredPatterns) > 0 {
			allCountedPatterns = append(allCountedPatterns, filteredPatterns)
		}
	}

	return allCountedPatterns, nil
}

func printFilteredPatterns(filteredPatterns []*P.Pattern, iter int) {
	pnum := 0
	fmt.Println("----------------------------------")
	fmt.Println(fmt.Sprintf("-------- Length %d patterns-------", iter))
	fmt.Println("----------------------------------")

	for _, p := range filteredPatterns {
		pnum++
		fmt.Printf("User Created")
		for i := 0; i < len(p.EventNames); i++ {
			fmt.Printf("--(%d min)---> %s (%.2f times)",
				int(p.Timings[i].Mean()/60.0), p.EventNames[i], p.Repeats[i].Mean())
		}
		fmt.Printf(fmt.Sprintf(" : (Count %d)\n\n\n", p.Count))
	}
	fmt.Println("----------------------------------")
}

func main() {
	// Initialize configs and connections.
	err := C.Init()
	if err != nil {
		log.Error("Failed to initialize.")
		os.Exit(1)
	}

	if *projectIdFlag <= 0 || *inputFileFlag == "" || *outputFileFlag == "" {
		log.Error("project_id and input_file are required.")
		os.Exit(1)
	}

	allCountedPatterns, err := minePatterns(*projectIdFlag, *inputFileFlag)
	if err != nil {
		log.Error("Failed to mine patterns.")
		os.Exit(1)
	}

	// Write to ouptput file.
	file, err := os.Create(*outputFileFlag)
	if err != nil {
		log.WithFields(log.Fields{"file": *outputFileFlag, "err": err}).Fatal("Unable to create file.")
		os.Exit(1)
	}
	defer file.Close()
	for _, patterns := range allCountedPatterns {
		for _, pattern := range patterns {
			b, err := json.Marshal(pattern)
			if err != nil {
				log.WithFields(log.Fields{"err": err}).Fatal("Unable to unmarshal pattern.")
			}
			pString := string(b)
			if _, err := file.WriteString(fmt.Sprintf("%s\n", pString)); err != nil {
				log.WithFields(log.Fields{"line": pString, "err": err}).Fatal("Unable to write to file.")
				os.Exit(1)
			}
		}
	}
}
