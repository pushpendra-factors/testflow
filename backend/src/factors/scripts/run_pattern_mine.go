package main

// Mine TOP_K Frequent patterns for every event combination (segment) at every iteration.

// Sample usage in terminal.
// export GOPATH=/Users/aravindmurthy/code/factors/backend/
// go run run_pattern_mine.go --env=development --etcd=localhost:2379 --disk_dir=/tmp/factors --s3_region=us-east-1 --s3=/tmp/factors-dev --num_routines=3 --project_id=<projectId> --model_id=<modelId>
// or
// go run run_pattern_mine.go --project_id=<projectId> --model_id=<modelId>
import (
	"bufio"
	"bytes"
	"encoding/json"
	C "factors/config"
	"factors/filestore"
	M "factors/model"
	P "factors/pattern"
	serviceDisk "factors/services/disk"
	serviceEtcd "factors/services/etcd"
	serviceS3 "factors/services/s3"
	"flag"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"

	_ "github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

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

func genSegmentedCandidates(
	patterns []*P.Pattern, userAndEventsInfo *P.UserAndEventsInfo) map[string][]*P.Pattern {
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
		cPatterns, _, err := P.GenCandidates(patterns, top_K, userAndEventsInfo)
		if err != nil {
			log.Fatal(err)
		}
		candidateSegments[k] = cPatterns
	}
	return candidateSegments
}

func genLenThreeSegmentedCandidates(lenTwoPatterns []*P.Pattern,
	userAndEventsInfo *P.UserAndEventsInfo) map[string][]*P.Pattern {
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
			p, startPatterns, endPatterns, top_K, userAndEventsInfo)
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
	userAndEventsInfo *P.UserAndEventsInfo, numRoutines int,
	outputFile *os.File) ([]*P.Pattern, error) {
	var lenOnePatterns []*P.Pattern
	for _, eventName := range eventNames {
		p, err := P.NewPattern([]string{eventName.Name}, userAndEventsInfo)
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
	userAndEventsInfo *P.UserAndEventsInfo, numRoutines int,
	outputFile *os.File) ([]*P.Pattern, error) {
	// Each event combination is a segment in itself.
	lenTwoPatterns, _, err := P.GenCandidates(
		lenOnePatterns, max_SEGMENTS, userAndEventsInfo)
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

func mineAndWritePatterns(projectId uint64, filepath string,
	userAndEventsInfo *P.UserAndEventsInfo, numRoutines int, outputFile *os.File) error {
	// Length One Patterns.
	eventNames, errCode := M.GetEventNames(projectId)
	if errCode != http.StatusFound {
		return fmt.Errorf("DB read of event names failed")
	}
	var filteredPatterns []*P.Pattern

	patternLen := 1
	filteredPatterns, err := mineAndWriteLenOnePatterns(
		eventNames, filepath, userAndEventsInfo, numRoutines, outputFile)
	if err != nil {
		return err
	}
	printFilteredPatterns(filteredPatterns, patternLen)

	patternLen++
	if patternLen > max_PATTERN_LENGTH {
		return nil
	}
	filteredPatterns, err = mineAndWriteLenTwoPatterns(
		filteredPatterns, filepath, userAndEventsInfo,
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
			filteredPatterns, userAndEventsInfo)
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
			filteredPatterns, userAndEventsInfo)
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

func buildPropertiesInfoFromInput(projectId uint64, filepath string) (*P.UserAndEventsInfo, error) {
	eMap := make(map[string]*P.PropertiesInfo)

	// Length One Patterns.
	eventNames, errCode := M.GetEventNames(projectId)
	if errCode != http.StatusFound {
		return nil, fmt.Errorf("DB read of event names failed")
	}
	for _, eventName := range eventNames {
		// Initialize info.
		eMap[eventName.Name] = &P.PropertiesInfo{
			NumericPropertyKeys:          make(map[string]bool),
			CategoricalPropertyKeyValues: make(map[string]map[string]bool),
		}
	}

	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	userAndEventsInfo := &P.UserAndEventsInfo{
		UserPropertiesInfo: &P.PropertiesInfo{
			NumericPropertyKeys:          make(map[string]bool),
			CategoricalPropertyKeyValues: make(map[string]map[string]bool),
		},
		EventPropertiesInfoMap: &eMap,
	}
	scanner := bufio.NewScanner(file)
	err = P.CollectPropertiesInfo(scanner, userAndEventsInfo)
	if err != nil {
		return nil, err
	}
	return userAndEventsInfo, nil
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

// TODO(Ankit): Change write logic to write to multiple chunks
func main() {

	envFlag := flag.String("env", "development", "")
	projectIdFlag := flag.Uint64("project_id", 0, "Project Id.")
	modelIdFlag := flag.Uint64("model_id", 0, "Model Id")
	etcd := flag.String("etcd", "localhost:2379", "Comma separated list of etcd endpoints localhost:2379,localhost:2378")
	diskDirFlag := flag.String("disk_dir", "/tmp/factors", "--disk_dir=/tmp/factors pass directory")
	s3BucketFlag := flag.String("s3", "/tmp/factors-dev", "")
	s3BucketRegionFlag := flag.String("s3_region", "us-east-1", "")
	numRoutinesFlag := flag.Int("num_routines", 3, "No of routines")

	// C.Init() calls flag.Parse()
	// Remove this after initializing C.Service.DB
	err := C.Init()
	if err != nil {
		log.WithError(err).Error("Failed to initialize config")
		os.Exit(1)
	}

	log.WithFields(log.Fields{
		"Env":           *envFlag,
		"EtcdEndpoints": *etcd,
		"DiskBaseDir":   *diskDirFlag,
		"ProjectId":     *projectIdFlag,
		"ModelId":       *modelIdFlag,
		"S3Bucket":      *s3BucketFlag,
		"S3Region":      *s3BucketRegionFlag,
		"NumRoutines":   *numRoutinesFlag,
	}).Infoln("Initialising")

	if *envFlag != "development" {
		err := fmt.Errorf("env [ %s ] not recognised", *envFlag)
		panic(err)
	}

	if *projectIdFlag <= 0 || *modelIdFlag <= 0 {
		log.Error("project_id and model_id are required.")
		os.Exit(1)
	}

	if *numRoutinesFlag < 1 {
		log.Error("num_routines is less than one.")
		os.Exit(1)
	}

	env := *envFlag
	modelId := *modelIdFlag
	projectId := *projectIdFlag

	diskDir := *diskDirFlag
	bucketName := *s3BucketFlag
	region := *s3BucketRegionFlag

	diskManager := serviceDisk.New(diskDir)

	var cloudManager filestore.FileManager

	if env == "development" {
		cloudManager = serviceDisk.New(bucketName)
	} else {
		cloudManager = serviceS3.New(bucketName, region)
	}

	etcdClient, err := serviceEtcd.New([]string{*etcd})
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("failed to init etcd client")
		os.Exit(1)
	}

	// TODO(Ankit):
	// This file should be pulled from cloud first
	input, Name := diskManager.GetModelEventsFilePathAndName(projectId, modelId)
	inputFilePath := input + Name

	userAndEventsInfo, err := buildPropertiesInfoFromInput(projectId, inputFilePath)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed to build user and event Info.")
		os.Exit(1)
	}
	userAndEventsInfoBytes, err := json.Marshal(userAndEventsInfo)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Fatal("Failed to unmarshal events Info.")
		os.Exit(1)
	}

	err = writeEventInfoFile(projectId, modelId, bytes.NewReader(userAndEventsInfoBytes), cloudManager)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Fatal("Failed to write events Info.")
		os.Exit(1)
	}

	// Write intermediate results to temporary file in the same directory.
	tmpOutputFilePath := fmt.Sprintf("%s/projects/%v/models/%v/tmp.txt", diskDir, projectId, modelId)
	tmpFile, err := os.Create(tmpOutputFilePath)
	if err != nil {
		log.WithFields(log.Fields{"file": tmpOutputFilePath, "err": err}).Fatal("Unable to create file.")
		os.Exit(1)
	}

	err = mineAndWritePatterns(projectId, inputFilePath,
		userAndEventsInfo, *numRoutinesFlag, tmpFile)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed to mine patterns.")
		os.Exit(1)
	}

	log.Infoln("Renaming ModelPatterns File")
	path, fName := diskManager.GetModelPatternsFilePathAndName(projectId, modelId)
	if err = os.Rename(tmpOutputFilePath, path+"/"+fName); err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed to rename output.")
		os.Exit(1)
	}
	tmpFile.Close()
	path, fName = diskManager.GetModelPatternsFilePathAndName(projectId, modelId)
	tmpFile, err = os.Open(path + "/" + fName)
	defer tmpFile.Close()
	if err != nil {
		log.WithError(err).Error("failed to open temp file")
		os.Exit(1)
	}

	log.Infoln("Cloudmanager Creating ModelPatterns File")
	path, fName = cloudManager.GetModelPatternsFilePathAndName(projectId, modelId)
	err = cloudManager.Create(path, fName, tmpFile)
	if err != nil {
		log.WithError(err).Error("cloud manager Failed to create model patterns file")
		os.Exit(1)
	}

	log.Infoln("Etcd fetch current project version")
	curVersion, err := etcdClient.GetProjectVersion()
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed to fetch current project version")
		os.Exit(1)
	}

	log.Printf("Current Version: %s", curVersion)

	path, name := cloudManager.GetProjectsDataFilePathAndName(curVersion)
	versionFile, err := cloudManager.Get(path, name)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed to read cur version file")
		os.Exit(1)
	}

	projectDatas := parseProjectsDataFile(versionFile)

	projectDatas = append(projectDatas, projectData{
		ID:      projectId,
		ModelID: modelId,
		Chunks:  []string{"TODO"},
	})

	newVersionName := fmt.Sprintf("%v", time.Now().Unix())
	err = writeProjectDataFile(newVersionName, projectDatas, cloudManager)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed to write new version file to cloud")
		os.Exit(1)
	}

	log.WithField("newVersion", newVersionName).Info("New version")

	err = etcdClient.SetProjectVersion(newVersionName)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed to write new version file to etcd")
		os.Exit(1)
	}
}

func writeProjectDataFile(newVersionName string, projectDatas []projectData, cloudManager filestore.FileManager) error {

	path, name := cloudManager.GetProjectsDataFilePathAndName(newVersionName)

	log.Printf("Writing Projects Data %v", projectDatas)

	buffer := bytes.NewBuffer(nil)

	for _, pD := range projectDatas {
		bytes, err := json.Marshal(pD)
		if err != nil {
			return err
		}
		_, err = buffer.WriteString(fmt.Sprintf("%s\n", string(bytes)))
		if err != nil {
			return err
		}
	}
	err := cloudManager.Create(path, name, bytes.NewReader(buffer.Bytes()))
	return err
}

func writeEventInfoFile(projectId, modelId uint64, events *bytes.Reader, cloudManager filestore.FileManager) error {
	path, name := cloudManager.GetModelEventInfoFilePathAndName(projectId, modelId)
	err := cloudManager.Create(path, name, events)
	if err != nil {
		log.WithError(err).Error("writeEventInfoFile Failed to write to cloud")
		return err
	}
	return err
}

type projectData struct {
	ID        uint64    `json:"pid"`
	ModelID   uint64    `json:"mid"`
	StartDate time.Time `json:"sd"`
	EndDate   time.Time `json:"ed"`
	Chunks    []string  `json:"cs"`
}

func parseProjectsDataFile(reader io.Reader) []projectData {

	scanner := bufio.NewScanner(reader)
	// Adjust scanner buffer capacity to 10MB per line.
	const maxCapacity = 10 * 1024 * 1024
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	projectDatas := make([]projectData, 0, 0)

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		var p projectData
		if err := json.Unmarshal([]byte(line), &p); err != nil {
			log.WithFields(log.Fields{"lineNum": lineNum, "err": err}).Error("Failed to unmarshal project")
			continue
		}
		projectDatas = append(projectDatas, p)
	}
	err := scanner.Err()
	if err != nil {
		log.WithError(err).Errorln("Scanner error")
	}

	return projectDatas
}
