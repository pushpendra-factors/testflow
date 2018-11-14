package main

// Mine TOP_K Frequent patterns for every event combination (segment) at every iteration.

// Sample usage in terminal.
// export GOPATH=/Users/aravindmurthy/code/factors/backend/
// go run run_pattern_mine.go --project_id=1 --input_file="" --output_file=""

import (
	"bufio"
	"encoding/json"
	C "factors/config"
	M "factors/model"
	P "factors/pattern"
	"flag"
	"fmt"
	"math"
	"os"
	"sort"
	"sync"

	_ "github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

var projectIdFlag = flag.Int("project_id", 0, "Project Id.")
var inputFileFlag = flag.String("input_file", "",
	"Input file of format user_id,user_creation_time,event_id,event_creation_time sorted by user_id and event_creation_time")
var outputFileFlag = flag.String("output_file", "", "All patterns written to file with each line a JSON")
var numRoutinesFlag = flag.Int("num_routines", 3, "Project Id.")

// The number of patterns generated is bounded to max_SEGMENTS * top_K per iteration.
// The amount of data and the time computed to generate this data is bounded
// by these constants.
const max_SEGMENTS = 100000
const top_K = 5
const max_PATTERN_LENGTH = 4

func countPatternsWorker(filepath string,
	patterns []*P.Pattern, wg *sync.WaitGroup) {
	file, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(file)
	// 10 MB buffer.
	const maxCapacity = 10 * 1024 * 1024
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)
	P.CountPatterns(scanner, patterns)
	file.Close()
	wg.Done()
}

func countPatterns(filepath string, patterns []*P.Pattern, numRoutines int) {
	var wg sync.WaitGroup
	numPatterns := len(patterns)
	log.Info(fmt.Sprintf("Num patterns to count Range: %d - %d", 0, numPatterns-1))
	batchSize := int(math.Ceil(float64(numPatterns) / float64(numRoutines)))
	for i := 0; i < numRoutines; i++ {
		// Each worker gets a slice of patterns to count.
		low := int(math.Min(float64(batchSize*i), float64(numPatterns)))
		high := int(math.Min(float64(batchSize*(i+1)), float64(numPatterns)))
		log.Info(fmt.Sprintf("Batch %d patterns to count range: %d:%d", i+1, low, high))
		wg.Add(1)
		go countPatternsWorker(filepath, patterns[low:high], &wg)
	}
	wg.Wait()
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

func genSegmentedCandidates(patterns []*P.Pattern, eventInfoMap *P.EventInfoMap) map[string][]*P.Pattern {
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
		cPatterns, _, err := P.GenCandidates(patterns, top_K, eventInfoMap)
		if err != nil {
			log.Fatal(err)
		}
		candidateSegments[k] = cPatterns
	}
	return candidateSegments
}

func genLenThreeSegmentedCandidates(lenTwoPatterns []*P.Pattern,
	eventInfoMap *P.EventInfoMap) map[string][]*P.Pattern {
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
			p, startPatterns, endPatterns, top_K, eventInfoMap)
		if err != nil {
			log.Fatal(err)
		}
		segmentedCandidates[p.String()] = lenThreeCandidates
	}
	return segmentedCandidates
}

func writePatterns(patterns []*P.Pattern, outputFile *os.File) error {
	for _, pattern := range patterns {
		b, err := json.Marshal(pattern)
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Fatal("Unable to unmarshal pattern.")
			return err
		}
		pString := string(b)
		if _, err := outputFile.WriteString(fmt.Sprintf("%s\n", pString)); err != nil {
			log.WithFields(log.Fields{"line": pString, "err": err}).Fatal("Unable to write to file.")
			return err
		}
	}
	return nil
}

func mineAndWriteLenOnePatterns(
	eventNames []M.EventName, filepath string,
	eventInfoMap *P.EventInfoMap, numRoutines int,
	outputFile *os.File) ([]*P.Pattern, error) {
	var lenOnePatterns []*P.Pattern
	for _, eventName := range eventNames {
		p, err := P.NewPattern([]string{eventName.Name}, eventInfoMap)
		if err != nil {
			return []*P.Pattern{}, fmt.Errorf("Pattern initialization failed")
		}
		lenOnePatterns = append(lenOnePatterns, p)
	}
	countPatterns(filepath, lenOnePatterns, numRoutines)
	filteredLenOnePatterns := filterPatterns(lenOnePatterns)
	if err := writePatterns(filteredLenOnePatterns, outputFile); err != nil {
		return []*P.Pattern{}, err
	}
	return filteredLenOnePatterns, nil
}

func mineAndWriteLenTwoPatterns(
	lenOnePatterns []*P.Pattern, filepath string,
	eventInfoMap *P.EventInfoMap, numRoutines int,
	outputFile *os.File) ([]*P.Pattern, error) {
	// Each event combination is a segment in itself.
	lenTwoPatterns, _, err := P.GenCandidates(
		lenOnePatterns, max_SEGMENTS, eventInfoMap)
	if err != nil {
		return []*P.Pattern{}, err
	}
	countPatterns(filepath, lenTwoPatterns, numRoutines)
	filteredLenTwoPatterns := filterPatterns(lenTwoPatterns)
	if err := writePatterns(filteredLenTwoPatterns, outputFile); err != nil {
		return []*P.Pattern{}, err
	}
	return lenTwoPatterns, nil
}

func mineAndWritePatterns(projectId int, filepath string,
	eventInfoMap *P.EventInfoMap, numRoutines int, outputFile *os.File) error {
	// Length One Patterns.
	eventNames, errCode := M.GetEventNames(uint64(projectId))
	if errCode != M.DB_SUCCESS {
		return fmt.Errorf("DB read of event names failed")
	}
	var filteredPatterns []*P.Pattern
	patternLen := 1
	filteredPatterns, err := mineAndWriteLenOnePatterns(
		eventNames, filepath, eventInfoMap, numRoutines, outputFile)
	if err != nil {
		return err
	}
	printFilteredPatterns(filteredPatterns, patternLen)

	patternLen++
	if patternLen > max_PATTERN_LENGTH {
		return nil
	}
	filteredPatterns, err = mineAndWriteLenTwoPatterns(
		filteredPatterns, filepath, eventInfoMap,
		numRoutines, outputFile)
	if err != nil {
		return err
	}
	printFilteredPatterns(filteredPatterns, patternLen)

	// Len three patterns generation in a block to free up memory of
	// lenThreeVariables after use.
	{
		patternLen++
		if patternLen > max_PATTERN_LENGTH {
			return nil
		}
		lenThreeSegmentedPatterns := genLenThreeSegmentedCandidates(
			filteredPatterns, eventInfoMap)
		lenThreePatterns := []*P.Pattern{}
		for _, patterns := range lenThreeSegmentedPatterns {
			lenThreePatterns = append(lenThreePatterns, patterns...)
		}
		countPatterns(filepath, lenThreePatterns, numRoutines)
		filteredPatterns = filterPatterns(lenThreePatterns)
		if err := writePatterns(filteredPatterns, outputFile); err != nil {
			return err
		}
		printFilteredPatterns(filteredPatterns, patternLen)
	}

	var candidatePatternsMap map[string][]*P.Pattern
	var candidatePatterns []*P.Pattern
	for len(filteredPatterns) > 0 {
		patternLen++
		if patternLen > max_PATTERN_LENGTH {
			return nil
		}

		candidatePatternsMap = genSegmentedCandidates(
			filteredPatterns, eventInfoMap)
		candidatePatterns = []*P.Pattern{}
		for _, patterns := range candidatePatternsMap {
			candidatePatterns = append(candidatePatterns, patterns...)
		}
		countPatterns(filepath, candidatePatterns, numRoutines)
		filteredPatterns = filterPatterns(candidatePatterns)
		if len(filteredPatterns) > 0 {
			if err := writePatterns(filteredPatterns, outputFile); err != nil {
				return err
			}
		}
		printFilteredPatterns(filteredPatterns, patternLen)
	}

	return nil
}

func buildEventInfoMapFromInput(projectId int, filepath string) (*P.EventInfoMap, error) {
	eMap := &P.EventInfoMap{}

	// Length One Patterns.
	eventNames, errCode := M.GetEventNames(uint64(projectId))
	if errCode != M.DB_SUCCESS {
		return nil, fmt.Errorf("DB read of event names failed")
	}
	for _, eventName := range eventNames {
		// Initialize info.
		(*eMap)[eventName.Name] = &P.EventInfo{
			NumericPropertyKeys:          make(map[string]bool),
			CategoricalPropertyKeyValues: make(map[string]map[string]bool),
		}
	}

	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	err = P.CollectEventInfo(scanner, eMap)
	if err != nil {
		return nil, err
	}
	return eMap, nil
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
			fmt.Printf("-----> %s", p.EventNames[i])
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
	if *numRoutinesFlag < 1 {
		log.Error("num_routines is less than one.")
		os.Exit(1)
	}

	// Initialize ouptput file.
	if _, err := os.Stat(*outputFileFlag); !os.IsNotExist(err) {
		log.WithFields(log.Fields{"file": *outputFileFlag, "err": err}).Fatal("File already exists.")
		os.Exit(1)
	}
	file, err := os.Create(*outputFileFlag)
	if err != nil {
		log.WithFields(log.Fields{"file": *outputFileFlag, "err": err}).Fatal("Unable to create file.")
		os.Exit(1)
	}
	defer file.Close()

	eventInfoMap, err := buildEventInfoMapFromInput(*projectIdFlag, *inputFileFlag)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed to build event Info.")
		os.Exit(1)
	}
	// First line in output file are events information seen in the data.
	eventInfoBytes, err := json.Marshal(eventInfoMap)
	eventInfoStr := string(eventInfoBytes)
	if _, err := file.WriteString(fmt.Sprintf("%s\n", eventInfoStr)); err != nil {
		log.WithFields(log.Fields{"line": eventInfoStr, "err": err}).Fatal("Failed to write events Info.")
		os.Exit(1)
	}

	err = mineAndWritePatterns(*projectIdFlag, *inputFileFlag,
		eventInfoMap, *numRoutinesFlag, file)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed to mine patterns.")
		os.Exit(1)
	}
}
