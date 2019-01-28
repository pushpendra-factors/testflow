package task

import (
	"bufio"
	"bytes"
	"encoding/json"
	"factors/filestore"
	M "factors/model"
	P "factors/pattern"
	PMM "factors/pattern_model_meta"
	serviceDisk "factors/services/disk"
	serviceEtcd "factors/services/etcd"
	U "factors/util"
	"fmt"
	"math"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
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

func computeAllUserPropertiesHistogram(filepath string, pattern *P.Pattern) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	// 10 MB buffer.
	const maxCapacity = 10 * 1024 * 1024
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)
	return P.ComputeAllUserPropertiesHistogram(scanner, pattern)
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
			log.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal pattern.")
			return err
		}
		pString := string(b)
		if _, err := outputFile.WriteString(fmt.Sprintf("%s\n", pString)); err != nil {
			log.WithFields(log.Fields{"line": pString, "err": err}).Error("Unable to write to file.")
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

func writeEventInfoFile(projectId, modelId uint64, events *bytes.Reader,
	cloudManager filestore.FileManager) error {

	path, name := cloudManager.GetModelEventInfoFilePathAndName(projectId, modelId)
	err := cloudManager.Create(path, name, events)
	if err != nil {
		log.WithError(err).Error("writeEventInfoFile Failed to write to cloud")
		return err
	}
	return err
}

// PatternMine Mine TOP_K Frequent patterns for every event combination (segment) at every iteration.
// TODO(Ankit): Change write logic to write to multiple chunks
func PatternMine(db *gorm.DB, etcdClient *serviceEtcd.EtcdClient, cloudManager *filestore.FileManager,
	diskManager *serviceDisk.DiskDriver, localDiskTmpDir string, bucketName string, numRoutines int,
	projectId uint64, modelId uint64, modelType string, startTime int64, endTime int64) (string, error) {

	var err error

	input, Name := (*cloudManager).GetModelEventsFilePathAndName(projectId, modelId)
	inputFilePath := input + Name

	userAndEventsInfo, err := buildPropertiesInfoFromInput(projectId, inputFilePath)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed to build user and event Info.")
		return "", err
	}
	userAndEventsInfoBytes, err := json.Marshal(userAndEventsInfo)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed to unmarshal events Info.")
		return "", err
	}

	err = writeEventInfoFile(projectId, modelId, bytes.NewReader(userAndEventsInfoBytes), (*cloudManager))
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed to write events Info.")
		return "", err
	}

	// Write incremental results to temporary file in local disk tmp directory and then copy final files to
	// cloud.
	chunkId := "01"
	_, fName := diskManager.GetPatternChunkFilePathAndName(projectId, modelId, chunkId)
	tmpOutputFilePath := fmt.Sprintf("%s/%s", localDiskTmpDir, fName)
	tmpFile, err := os.Create(tmpOutputFilePath)
	if err != nil {
		log.WithFields(log.Fields{"file": tmpOutputFilePath, "err": err}).Error("Unable to create file.")
		return "", err
	}
	defer tmpFile.Close()

	// First build histogram of all user properties.
	allActiveUsersPattern, err := P.NewPattern([]string{U.SEN_ALL_ACTIVE_USERS}, userAndEventsInfo)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed to build pattern with histogram of all active user properties.")
		return "", err
	}
	if err := computeAllUserPropertiesHistogram(inputFilePath, allActiveUsersPattern); err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed to compute user properties.")
		return "", err
	}
	if err := writePatterns([]*P.Pattern{allActiveUsersPattern}, tmpFile); err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed to write user properties.")
		return "", err
	}

	err = mineAndWritePatterns(projectId, inputFilePath,
		userAndEventsInfo, numRoutines, tmpFile)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed to mine patterns.")
		return "", err
	}

	log.Infoln("Cloudmanager Creating ModelPatterns File")
	tmpFile.Seek(0, 0) // Reset to the begining, before writing to cloud.
	path, fName := (*cloudManager).GetPatternChunkFilePathAndName(projectId, modelId, chunkId)
	err = (*cloudManager).Create(path, fName, tmpFile)
	if err != nil {
		log.WithError(err).Error("cloud manager Failed to create model patterns file")
		return "", err
	}

	// Update meta. Need to move this out as seperate task?
	projectDatas, err := PMM.GetProjectsMetadata(cloudManager, etcdClient)
	if err != nil {
		return "", err
	}
	projectDatas = append(projectDatas, PMM.ProjectData{
		ID:             projectId,
		ModelID:        modelId,
		ModelType:      modelType,
		StartTimestamp: startTime,
		EndTimestamp:   endTime,
		Chunks:         []string{chunkId},
	})

	newVersionName := fmt.Sprintf("%v", time.Now().Unix())
	err = PMM.WriteProjectDataFile(newVersionName, projectDatas, cloudManager)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed to write new version file to cloud")
		return "", err
	}

	err = etcdClient.SetProjectVersion(newVersionName)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed to write new version file to etcd")
		return "", err
	}

	log.WithField("newVersion", newVersionName).Info("Patterns mined successfully")
	return newVersionName, nil
}
