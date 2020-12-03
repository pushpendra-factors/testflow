package task

import (
	"bufio"
	"bytes"
	"encoding/json"
	"factors/filestore"
	M "factors/model"
	P "factors/pattern"
	PMM "factors/pattern_model_meta"
	"factors/pattern_server/store"
	serviceDisk "factors/services/disk"
	serviceEtcd "factors/services/etcd"
	U "factors/util"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

// The number of patterns generated is bounded to max_SEGMENTS * top_K per iteration.
// The amount of data and the time computed to generate this data is bounded
// by these constants.
const max_SEGMENTS = 25000
const max_EVENT_NAMES = 250
const top_K = 5
const topK_patterns = 10
const topKProperties = 5
const keventsSpecial = 5
const keventsURL = 100
const max_PATTERN_LENGTH = 3

const max_CHUNK_SIZE_IN_BYTES int64 = 200 * 1000 * 1000 // 200MB

var regex_NUM = regexp.MustCompile("[0-9]+")
var mineLog = taskLog.WithField("prefix", "Task#PatternMine")

type patternProperties struct {
	pattern     *P.Pattern
	count       uint
	patternType string
}

func countPatternsWorker(filepath string,
	patterns []*P.Pattern, wg *sync.WaitGroup, countOccurence bool) {
	file, err := os.Open(filepath)
	if err != nil {
		mineLog.WithField("filePath", filepath).Error("Failure on count pattern workers.")
	}

	scanner := bufio.NewScanner(file)
	buf := make([]byte, P.MAX_PATTERN_BYTES)
	scanner.Buffer(buf, P.MAX_PATTERN_BYTES)
	P.CountPatterns(scanner, patterns, countOccurence)
	file.Close()
	wg.Done()
}

func countPatterns(filepath string, patterns []*P.Pattern, numRoutines int, countOccurence bool) {
	var wg sync.WaitGroup
	numPatterns := len(patterns)
	mineLog.Info(fmt.Sprintf("Num patterns to count Range: %d - %d", 0, numPatterns-1))
	batchSize := int(math.Ceil(float64(numPatterns) / float64(numRoutines)))
	for i := 0; i < numRoutines; i++ {
		// Each worker gets a slice of patterns to count.
		low := int(math.Min(float64(batchSize*i), float64(numPatterns)))
		high := int(math.Min(float64(batchSize*(i+1)), float64(numPatterns)))
		mineLog.Info(fmt.Sprintf("Batch %d patterns to count range: %d:%d", i+1, low, high))
		wg.Add(1)
		go countPatternsWorker(filepath, patterns[low:high], &wg, countOccurence)
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
func filterAndCompressPatterns(
	patterns []*P.Pattern, maxTotalBytes int64, totalConsumedBytes int64,
	currentPatternsLength int, maxPatternsLength int) ([]*P.Pattern, int64, error) {

	if currentPatternsLength > maxPatternsLength {
		errorString := fmt.Sprintf(
			"Current pattern length greater than max length. currentPatternsLength:%d, maxPatternsLength: %d",
			currentPatternsLength, maxPatternsLength)
		mineLog.Error(errorString)
		return []*P.Pattern{}, 0, fmt.Errorf(errorString)
	}
	if totalConsumedBytes >= maxTotalBytes {
		mineLog.Info(fmt.Sprintf("No quota. totalConsumedBytes: %d, maxTotalBytes: %d",
			totalConsumedBytes, maxTotalBytes))
		return []*P.Pattern{}, 0, nil
	}
	countFilteredPatterns := []*P.Pattern{}
	for _, p := range patterns {
		if p.PerUserCount > 0 {
			countFilteredPatterns = append(countFilteredPatterns, p)
		}
	}

	// More quota to smaller patterns.
	// Ex: maxLen = 4, maxTotalBytes = 10G
	// len1 patterns will get 10 / (4 - 1 + 1) = 2.5G
	// If len1 actually takes 2G then len2 patterns will get (10 - 2) / (4 - 2 + 1) = 2.66G
	// Compression is done on best effort. Patterns are retained as long as
	// totalConsumed Bytes does not cross over maxTotalBytes.
	// If len1 actually takes 4G then len2 patterns will get (10 - 4) / (4 - 2 + 1) = 2G
	// If len2 takes 2G then len3 gets (10 - 6) / (4 - 3 + 1) = 2G
	currentPatternsQuota := int64(float64(maxTotalBytes-totalConsumedBytes) / float64(
		maxPatternsLength-currentPatternsLength+1))
	compressedPatterns, compressedPatternsBytes, err := compressPatterns(
		countFilteredPatterns, currentPatternsQuota)
	if err != nil {
		return []*P.Pattern{}, 0, err
	}
	if (totalConsumedBytes + compressedPatternsBytes) <= maxTotalBytes {
		mineLog.WithFields(log.Fields{
			"numPatterns":          len(compressedPatterns),
			"maxTotalBytes":        maxTotalBytes,
			"totalConsumedBytes":   totalConsumedBytes,
			"currentPatternsBytes": compressedPatternsBytes,
		}).Info("Returning compressed patterns")
		return compressedPatterns, compressedPatternsBytes, nil
	}

	// Patterns are added only till it does not go over maxTotalBytes, in
	// decreasing order of count.
	// Sort the patterns in descending order.
	sort.Slice(compressedPatterns,
		func(i, j int) bool {
			return compressedPatterns[i].PerUserCount > compressedPatterns[j].PerUserCount
		})
	var cumulativeBytes int64 = 0
	compressedAndDroppedPatterns := []*P.Pattern{}
	for i, pattern := range compressedPatterns {
		b, err := json.Marshal(pattern)
		if err != nil {
			mineLog.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal pattern.")
			return []*P.Pattern{}, 0, err
		}
		pString := string(b)
		pBytes := int64(len([]byte(pString)))
		if totalConsumedBytes+cumulativeBytes+pBytes > maxTotalBytes {
			mineLog.WithFields(log.Fields{
				"numPatterns":          len(compressedPatterns),
				"numDroppedPatterns":   len(compressedPatterns) - i,
				"maxTotalBytes":        maxTotalBytes,
				"totalConsumedBytes":   totalConsumedBytes,
				"currentPatternsBytes": cumulativeBytes,
			}).Info("Dropping patterns")
			break
		}
		compressedAndDroppedPatterns = append(compressedAndDroppedPatterns, pattern)
		cumulativeBytes += pBytes
	}
	return compressedAndDroppedPatterns, cumulativeBytes, nil
}

// Compress the size of patterns in memory to the desired overall quota
// in bytes.
func compressPatterns(patterns []*P.Pattern, maxBytesSize int64) ([]*P.Pattern, int64, error) {
	if maxBytesSize <= 0 {
		return patterns, 0, fmt.Errorf(fmt.Sprintf("Incorrect maxBytesSize value. %d", maxBytesSize))
	}
	var patternsBytes int64 = 0
	for _, pattern := range patterns {
		b, err := json.Marshal(pattern)
		if err != nil {
			mineLog.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal pattern.")
			return []*P.Pattern{}, 0, err
		}
		pString := string(b)
		patternsBytes += int64(len([]byte(pString)))
	}
	// Already within quota.
	if patternsBytes <= maxBytesSize {
		return patterns, patternsBytes, nil
	}

	// First try with decreasing frequency map size of categorical histograms proportionally.
	TRIM_MULTIPLIER := 0.8
	trimFraction := float64(maxBytesSize) * TRIM_MULTIPLIER / float64(patternsBytes)

	var patternsTrim1Bytes int64 = 0
	for _, pattern := range patterns {
		(*pattern.PerUserEventCategoricalProperties).TrimByFmapSize(trimFraction)
		(*pattern.PerUserUserCategoricalProperties).TrimByFmapSize(trimFraction)
		(*pattern.PerOccurrenceEventCategoricalProperties).TrimByFmapSize(trimFraction)
		(*pattern.PerOccurrenceUserCategoricalProperties).TrimByFmapSize(trimFraction)
		b, err := json.Marshal(pattern)
		if err != nil {
			mineLog.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal pattern.")
			return nil, 0, err
		}
		pString := string(b)
		patternsTrim1Bytes += int64(len([]byte(pString)))
	}

	mineLog.WithFields(log.Fields{
		"initialSize":          patternsBytes,
		"beforeCompressSize":   patternsBytes,
		"afterCompressSize":    patternsTrim1Bytes,
		"maxSizeforCurrentSet": maxBytesSize,
	}).Info("Compression by Trim 1")

	if patternsTrim1Bytes <= maxBytesSize {
		return patterns, patternsTrim1Bytes, nil
	}
	// Next try by decreasing number of bins of numerical histograms.
	trimFraction = float64(maxBytesSize) * TRIM_MULTIPLIER / float64(patternsTrim1Bytes)
	var patternsTrim2Bytes int64 = 0.0
	for _, pattern := range patterns {
		(*pattern.PerUserEventNumericProperties).TrimByBinSize(trimFraction)
		(*pattern.PerUserUserNumericProperties).TrimByBinSize(trimFraction)
		(*pattern.PerOccurrenceEventNumericProperties).TrimByBinSize(trimFraction)
		(*pattern.PerOccurrenceUserNumericProperties).TrimByBinSize(trimFraction)
		b, err := json.Marshal(pattern)
		if err != nil {
			mineLog.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal pattern.")
			return nil, 0, err
		}
		pString := string(b)
		patternsTrim2Bytes += int64(len([]byte(pString)))
	}
	mineLog.WithFields(log.Fields{
		"initialSize":          patternsBytes,
		"beforeCompressSize":   patternsTrim1Bytes,
		"afterCompressSize":    patternsTrim2Bytes,
		"maxSizeforCurrentSet": maxBytesSize,
	}).Info("Compression by Trim 2")

	if patternsTrim2Bytes <= maxBytesSize {
		return patterns, patternsTrim2Bytes, nil
	}

	// Next try decreasing the number of bins of categorical histograms.
	trimFraction = float64(maxBytesSize) * TRIM_MULTIPLIER / float64(patternsTrim2Bytes)
	var patternsTrim3Bytes int64 = 0
	for _, pattern := range patterns {
		(*pattern.PerUserEventCategoricalProperties).TrimByBinSize(trimFraction)
		(*pattern.PerUserUserCategoricalProperties).TrimByBinSize(trimFraction)
		(*pattern.PerOccurrenceEventCategoricalProperties).TrimByBinSize(trimFraction)
		(*pattern.PerOccurrenceUserCategoricalProperties).TrimByBinSize(trimFraction)
		b, err := json.Marshal(pattern)
		if err != nil {
			mineLog.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal pattern.")
			return nil, 0, err
		}
		pString := string(b)
		patternsTrim3Bytes += int64(len([]byte(pString)))
	}
	mineLog.WithFields(log.Fields{
		"initialSize":          patternsBytes,
		"beforeCompressSize":   patternsTrim2Bytes,
		"afterCompressSize":    patternsTrim3Bytes,
		"maxSizeforCurrentSet": maxBytesSize,
	}).Info("Compression by Trim 3")

	return patterns, patternsTrim3Bytes, nil
}

func genSegmentedCandidates(
	patterns []*P.Pattern, userAndEventsInfo *P.UserAndEventsInfo, cyclicEvents []string) (map[string][]*P.Pattern, error) {
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
		cPatterns, _, err := P.GenCandidates(patterns, top_K, userAndEventsInfo, cyclicEvents)
		if err != nil {
			mineLog.Error("Failure on generate segemented candidates.")
			return pSegments, err
		}
		candidateSegments[k] = cPatterns
	}
	return candidateSegments, nil
}

func genLenThreeSegmentedCandidates(lenTwoPatterns []*P.Pattern,
	userAndEventsInfo *P.UserAndEventsInfo, cyclicEvents []string) (map[string][]*P.Pattern, error) {
	startPatternsMap := make(map[string][]*P.Pattern)
	endPatternsMap := make(map[string][]*P.Pattern)

	segmentedCandidates := make(map[string][]*P.Pattern)
	for _, p := range lenTwoPatterns {
		pEvents := p.EventNames
		if len(pEvents) != 2 {
			mineLog.Error(fmt.Sprintf("Pattern %s is not of length two.", p.EventNames))
			return segmentedCandidates, fmt.Errorf("pattern %s is not of length two", p.EventNames)
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
				return patterns[i].PerUserCount > patterns[j].PerUserCount
			})
	}
	for _, patterns := range endPatternsMap {
		sort.Slice(patterns,
			func(i, j int) bool {
				return patterns[i].PerUserCount > patterns[j].PerUserCount
			})
	}

	for _, p := range lenTwoPatterns {
		startPatterns, ok1 := startPatternsMap[p.EventNames[0]]
		endPatterns, ok2 := endPatternsMap[p.EventNames[1]]
		if startPatterns == nil || endPatterns == nil || !ok1 || !ok2 {
			continue
		}
		lenThreeCandidates, err := P.GenLenThreeCandidatePatterns(
			p, startPatterns, endPatterns, top_K, userAndEventsInfo, cyclicEvents)
		if err != nil {
			mineLog.WithError(err).Error("Failed on genLenThreeSegmentedCandidates.")
			return segmentedCandidates, err
		}
		segmentedCandidates[p.String()] = lenThreeCandidates
	}
	return segmentedCandidates, nil
}

func getUniqueCandidates(allCandidates []*P.Pattern) []*P.Pattern {

	patternDict := make(map[*P.Pattern]bool)
	allUniquePatterns := make([]*P.Pattern, 0)
	for _, c := range allCandidates {

		if _, ok := patternDict[c]; !ok {
			patternDict[c] = true
			allUniquePatterns = append(allUniquePatterns, c)
		}
	}

	return allUniquePatterns

}

func mineAndWriteLenOnePatterns(
	eventNames []string, filepath string,
	userAndEventsInfo *P.UserAndEventsInfo, numRoutines int,
	chunkDir string, maxModelSize int64, cumulativePatternsSize int64, countOccurence bool) (
	[]*P.Pattern, int64, error) {
	var lenOnePatterns []*P.Pattern
	for _, eventName := range eventNames {
		p, err := P.NewPattern([]string{eventName}, userAndEventsInfo)
		if err != nil {
			return []*P.Pattern{}, 0, fmt.Errorf("Pattern initialization failed")
		}
		lenOnePatterns = append(lenOnePatterns, p)
	}

	countPatterns(filepath, lenOnePatterns, numRoutines, countOccurence)
	filteredLenOnePatterns, patternsSize, err := filterAndCompressPatterns(
		lenOnePatterns, maxModelSize, cumulativePatternsSize, 1, max_PATTERN_LENGTH)
	if err != nil {
		return []*P.Pattern{}, 0, err
	}
	if err := writePatternsAsChunks(filteredLenOnePatterns, chunkDir); err != nil {
		return []*P.Pattern{}, 0, err
	}
	return filteredLenOnePatterns, patternsSize, nil
}

func mineAndWriteLenTwoPatterns(
	lenOnePatterns []*P.Pattern, filepath string,
	userAndEventsInfo *P.UserAndEventsInfo, numRoutines int,
	chunkDir string, maxModelSize int64, cumulativePatternsSize int64, countOccurence bool, goalPatterns []*P.Pattern) (
	[]*P.Pattern, int64, error) {
	// Each event combination is a segment in itself.
	lenTwoPatterns, _, err := P.GenSegmentsForTopGoals(
		lenOnePatterns, userAndEventsInfo, goalPatterns)

	if err != nil {
		return []*P.Pattern{}, 0, err
	}
	countPatterns(filepath, lenTwoPatterns, numRoutines, countOccurence)
	filteredLenTwoPatterns, patternsSize, err := filterAndCompressPatterns(
		lenTwoPatterns, maxModelSize, cumulativePatternsSize, 2, max_PATTERN_LENGTH)
	if err != nil {
		return []*P.Pattern{}, 0, err
	}
	if err := writePatternsAsChunks(filteredLenTwoPatterns, chunkDir); err != nil {
		return []*P.Pattern{}, 0, err
	}
	return filteredLenTwoPatterns, patternsSize, nil
}

//GetGoalPatterns get all goalPatterns from DB
func GetGoalPatterns(projectId uint64, filteredPatterns []*P.Pattern, eventNamesWithType map[string]string) ([]*P.Pattern, error) {

	goalPatternsFromDB, errCode := M.GetAllActiveFactorsGoals(projectId)

	if errCode != http.StatusFound {
		mineLog.Info("Failure on Get goal patterns.")
	}
	var goalPatterns []*P.Pattern
	if goalPatternsFromDB != nil && len(goalPatternsFromDB) > 0 {
		mineLog.Info(fmt.Sprintf("Number of Goals from DB:%d", len(goalPatternsFromDB)))

		tmpPatterns := make(map[string]*P.Pattern)

		for _, v := range filteredPatterns {
			tmpPatterns[v.String()] = v
		}
		for _, p := range goalPatternsFromDB {
			if valPattern, ok := tmpPatterns[p.Name]; ok {
				goalPatterns = append(goalPatterns, valPattern)
				mineLog.Info(fmt.Sprint("Goal event from db ", valPattern.String()))

			}
		}
		return goalPatterns, nil
	}

	goalTopKPatterns := FilterTopKEventsOnTypes(filteredPatterns, eventNamesWithType, topK_patterns, keventsSpecial, keventsURL)
	mineLog.Info(fmt.Sprintf("Mining goals from topk events"))
	for _, v := range goalTopKPatterns {
		mineLog.Info(fmt.Sprint("Goal event: ", v.String()))
	}

	for _, valPat := range goalTopKPatterns {
		tmpFactorsRule := M.FactorsGoalRule{EndEvent: valPat.String()}
		goalID, httpStatusTrackedEvent := M.CreateFactorsTrackedEvent(projectId, valPat.String(), "")
		if goalID == 0 {
			mineLog.Error("Unable to create a trackedEvent ", httpStatusTrackedEvent, " ", goalID)
			return nil, fmt.Errorf("unable to write tracked event to db")
		}
		mineLog.Info("trackedEvent in db  ", httpStatusTrackedEvent, " ", goalID)

		_, httpstatus, err := M.CreateFactorsGoal(projectId, valPat.String(), tmpFactorsRule, "")
		if httpstatus != http.StatusCreated {
			mineLog.Error("Unable to write to db ", httpstatus, err)
			return nil, fmt.Errorf("unable to write factors goal to db")

		}
	}

	return goalTopKPatterns, nil

}

func mineAndWritePatterns(projectId uint64, filepath string,
	userAndEventsInfo *P.UserAndEventsInfo, eventNames []string,
	numRoutines int, chunkDir string,
	maxModelSize int64, countOccurence bool, eventNamesWithType map[string]string, repeatedEvents []string) error {
	var filteredPatterns []*P.Pattern
	var cumulativePatternsSize int64 = 0

	patternLen := 1
	limitRoundOffFraction := 0.99

	filteredPatterns, patternsSize, err := mineAndWriteLenOnePatterns(
		eventNames, filepath, userAndEventsInfo, numRoutines, chunkDir,
		maxModelSize, cumulativePatternsSize, countOccurence)
	if err != nil {
		return err
	}
	cumulativePatternsSize += patternsSize
	printFilteredPatterns(filteredPatterns, patternLen)

	goalPatterns, err := GetGoalPatterns(projectId, filteredPatterns, eventNamesWithType)

	if cumulativePatternsSize >= int64(float64(maxModelSize)*limitRoundOffFraction) {
		return nil
	}

	patternLen++
	if patternLen > max_PATTERN_LENGTH {
		return nil
	}
	filteredPatterns, patternsSize, err = mineAndWriteLenTwoPatterns(
		filteredPatterns, filepath, userAndEventsInfo,
		numRoutines, chunkDir, maxModelSize, cumulativePatternsSize, countOccurence, goalPatterns)
	if err != nil {
		return err
	}
	cumulativePatternsSize += patternsSize
	printFilteredPatterns(filteredPatterns, patternLen)
	if cumulativePatternsSize >= int64(float64(maxModelSize)*limitRoundOffFraction) {
		return nil
	}

	// Len three patterns generation in a block to free up memory of
	// lenThreeVariables after use.
	{
		patternLen++
		if patternLen > max_PATTERN_LENGTH {
			return nil
		}
		lenThreeSegmentedPatterns, err := genLenThreeSegmentedCandidates(
			filteredPatterns, userAndEventsInfo, repeatedEvents)
		if err != nil {
			return err
		}
		lenThreePatterns := []*P.Pattern{}
		for _, patterns := range lenThreeSegmentedPatterns {
			lenThreePatterns = append(lenThreePatterns, patterns...)
		}
		countPatterns(filepath, lenThreePatterns, numRoutines, countOccurence)
		filteredPatterns, patternsSize, err = filterAndCompressPatterns(
			lenThreePatterns, maxModelSize, cumulativePatternsSize,
			patternLen, max_PATTERN_LENGTH)
		if err != nil {
			return err
		}
		cumulativePatternsSize += patternsSize
		if err := writePatternsAsChunks(filteredPatterns, chunkDir); err != nil {
			return err
		}
		printFilteredPatterns(filteredPatterns, patternLen)
		if cumulativePatternsSize >= int64(float64(maxModelSize)*limitRoundOffFraction) {
			return nil
		}
	}

	var candidatePatternsMap map[string][]*P.Pattern
	var candidatePatterns []*P.Pattern
	for len(filteredPatterns) > 0 && cumulativePatternsSize < maxModelSize {
		patternLen++
		if patternLen > max_PATTERN_LENGTH {
			return nil
		}

		candidatePatternsMap, err = genSegmentedCandidates(
			filteredPatterns, userAndEventsInfo, repeatedEvents)
		if err != nil {
			return err
		}
		candidatePatterns = []*P.Pattern{}
		for _, patterns := range candidatePatternsMap {
			candidatePatterns = append(candidatePatterns, patterns...)
		}
		countPatterns(filepath, candidatePatterns, numRoutines, countOccurence)
		filteredPatterns, patternsSize, err = filterAndCompressPatterns(
			candidatePatterns, maxModelSize, cumulativePatternsSize,
			patternLen, max_PATTERN_LENGTH)
		if err != nil {
			return err
		}
		if len(filteredPatterns) > 0 {
			cumulativePatternsSize += patternsSize
			if err := writePatternsAsChunks(filteredPatterns, chunkDir); err != nil {
				return err
			}
		}
		printFilteredPatterns(filteredPatterns, patternLen)
	}

	return nil
}

func buildPropertiesInfoFromInput(projectId uint64, eventNames []string, filepath string) (*P.UserAndEventsInfo, map[string]P.PropertiesCount, error) {
	userAndEventsInfo := P.NewUserAndEventsInfo()
	eMap := *userAndEventsInfo.EventPropertiesInfoMap
	for _, eventName := range eventNames {
		// Initialize info.
		eMap[eventName] = &P.PropertiesInfo{
			NumericPropertyKeys:          make(map[string]bool),
			CategoricalPropertyKeyValues: make(map[string]map[string]bool),
		}
	}

	file, err := os.Open(filepath)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	allProperty, err := P.CollectPropertiesInfo(scanner, userAndEventsInfo)
	if err != nil {
		return nil, nil, err
	}
	return userAndEventsInfo, *allProperty, nil
}

func printFilteredPatterns(filteredPatterns []*P.Pattern, iter int) {
	mineLog.Info(fmt.Sprintf("Mined %d patterns of length %d", len(filteredPatterns), iter))

	/*
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
	*/

}

func writeEventInfoFile(projectId, modelId uint64, events *bytes.Reader,
	cloudManager filestore.FileManager) error {

	path, name := cloudManager.GetModelEventInfoFilePathAndName(projectId, modelId)
	err := cloudManager.Create(path, name, events)
	if err != nil {
		mineLog.WithError(err).Error("writeEventInfoFile Failed to write to cloud")
		return err
	}
	return err
}

func getChunkIdFromName(chunkFileName string) string {
	if !strings.HasPrefix(chunkFileName, "chunk_") {
		return ""
	}

	return regex_NUM.FindString(chunkFileName)
}

func getLastChunkInfo(chunksDir string) (int, os.FileInfo, error) {
	cfiles, err := ioutil.ReadDir(chunksDir)
	if err != nil {
		return 0, nil, err
	}
	lastChunkIndex := 0
	var lastChunkFileInfo os.FileInfo
	for _, cf := range cfiles {
		cfName := cf.Name()
		if chunkIdStr := getChunkIdFromName(cfName); chunkIdStr != "" {
			ci, err := strconv.Atoi(chunkIdStr)
			if err != nil {
				mineLog.WithFields(log.Fields{"chunkDir": chunksDir,
					"fileName": cfName}).Error("Failed to parse chunk index")
				continue
			}
			if ci > lastChunkIndex {
				lastChunkIndex = ci
				lastChunkFileInfo = cf
			}
		}
	}

	return lastChunkIndex, lastChunkFileInfo, nil
}

func writePatternsAsChunks(patterns []*P.Pattern, chunksDir string) error {
	lastChunkIndex, lastChunkFileInfo, err := getLastChunkInfo(chunksDir)
	if err != nil {
		mineLog.WithFields(log.Fields{"err": err}).Error("Failed to read chunks dir.")
		return err
	}

	var (
		currentFilePath string
		currentFileSize int64
		currentFile     *os.File
	)

	currentFileIndex := lastChunkIndex
	// initialize with existing chunk file.
	if lastChunkIndex > 0 && lastChunkFileInfo != nil {
		currentFilePath = fmt.Sprintf("%s/%s", chunksDir, lastChunkFileInfo.Name())
		currentFileSize = lastChunkFileInfo.Size()
		currentFile, err = os.OpenFile(currentFilePath, os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			mineLog.WithError(err).Error("Failed to open chunk file.")
			return err
		}
	}

	for _, pattern := range patterns {
		patternBytes, err := json.Marshal(pattern)
		if err != nil {
			mineLog.WithFields(log.Fields{"err": err}).Error("Unable to marshal pattern.")
			return err
		}

		patternWithMeta := store.PatternWithMeta{
			PatternEvents: pattern.EventNames,
			RawPattern:    json.RawMessage(patternBytes),
		}

		pwmBytes, err := json.Marshal(patternWithMeta)
		if err != nil {
			mineLog.WithFields(log.Fields{"err": err}).Error("Unable to marshal pattern_with_meta.")
			return err
		}

		pString := string(pwmBytes)
		pString = pString + "\n"
		pBytes := []byte(pString)
		pBytesLen := int64(len(pBytes))
		if pBytesLen >= 10000000 {
			// Limit is 10MB
			errorString := fmt.Sprintf(
				"Too big pattern, chunksDir: %s, pattern: %s, numBytes: %d",
				chunksDir, pattern.String(), pBytesLen)
			mineLog.Error(errorString)
			return fmt.Errorf(errorString)
		}

		fileHasSpace := false
		if currentFileIndex > 0 {
			balanceSpace := max_CHUNK_SIZE_IN_BYTES - currentFileSize
			if balanceSpace > 0 && balanceSpace >= pBytesLen {
				fileHasSpace = true
			}
		}

		if !fileHasSpace {
			if currentFileIndex > 0 {
				err := currentFile.Close()
				if err != nil {
					mineLog.WithError(err).Error("Failed to close chunk file.")
				}
			}
			nextFileIndex := currentFileIndex + 1
			nextFileName := fmt.Sprintf("%s_%d", "chunk", nextFileIndex)

			currentFileIndex = nextFileIndex
			currentFilePath = fmt.Sprintf("%s/%s.txt", chunksDir, nextFileName)
			currentFileSize = 0
			currentFile, err = os.Create(currentFilePath)
			defer currentFile.Close()
		}

		_, err = currentFile.Write(pBytes)
		if err != nil {
			mineLog.WithFields(log.Fields{"line": pString, "err": err, "filePath": currentFilePath,
				"fileSize": currentFileSize}).Error("Failed write to chunk file.")
			return err
		}

		currentFileSize = currentFileSize + pBytesLen
	}

	return nil
}

func uploadChunksToCloud(tmpChunksDir, cloudChunksDir string, cloudManager *filestore.FileManager) ([]string, error) {
	cfiles, err := ioutil.ReadDir(tmpChunksDir)
	if err != nil {
		return nil, err
	}

	uploadedChunkIds := make([]string, 0, 0)
	for _, cf := range cfiles {
		cfName := cf.Name()
		if chunkIdStr := getChunkIdFromName(cfName); chunkIdStr != "" {
			cfPath := fmt.Sprintf("%s/%s", tmpChunksDir, cfName)
			cfFile, err := os.OpenFile(cfPath, os.O_RDWR, 0666)
			if err != nil {
				mineLog.WithError(err).Error("Failed to open tmp chunk to upload.")
				return uploadedChunkIds, err
			}
			err = (*cloudManager).Create(cloudChunksDir, cfName, cfFile)
			if err != nil {
				mineLog.WithError(err).Error("Failed to chunk upload file to cloud.")
				return uploadedChunkIds, err
			}
			uploadedChunkIds = append(uploadedChunkIds, chunkIdStr)
		}
	}

	return uploadedChunkIds, nil
}

func filterAndWriteFactorsProperties(tmpPath string, reader io.Reader, userPropMap, eventPropMap map[string]bool) error {
	scanner := bufio.NewScanner(reader)
	file, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	log.Info("Create a temp file to save and read events")
	defer file.Close()

	w := bufio.NewWriter(file)
	for scanner.Scan() {
		line := scanner.Text()
		var eventDetails P.CounterEventFormat
		if err := json.Unmarshal([]byte(line), &eventDetails); err != nil {
			log.WithFields(log.Fields{"line": line, "err": err}).Error("Read failed")
			return err
		}

		for uKey := range eventDetails.UserProperties {
			if _, ok := userPropMap[uKey]; !ok {
				delete(eventDetails.UserProperties, uKey)
			}
		}

		for eKey := range eventDetails.EventProperties {
			if _, ok := eventPropMap[eKey]; !ok {
				delete(eventDetails.EventProperties, eKey)
			}
		}

		eventDetailsBytes, err := json.Marshal(eventDetails)

		if err != nil {
			log.WithFields(log.Fields{"line": line, "err": err}).Error("Failed to marshal eventDetails")
			return err
		}

		lineWrite := string(eventDetailsBytes)
		if _, err := file.WriteString(fmt.Sprintf("%s\n", lineWrite)); err != nil {
			peLog.WithFields(log.Fields{"line": line, "err": err}).Error("Unable to write to file.")
			return err
		}

	}
	w.Flush()

	return nil
}

func GetEventNamesAndType(tmpEventsFilePath string, projectId uint64) ([]string, map[string]string, error) {
	scanner, err := OpenEventFileAndGetScanner(tmpEventsFilePath)
	eventNames, err := M.GetEventNamesFromFile(scanner, projectId)
	if err != nil {
		mineLog.WithFields(log.Fields{"err": err, "eventFilePath": tmpEventsFilePath}).Error("Failed to read event names from file")
		return nil, nil, err
	}
	mineLog.WithField("tmpEventsFilePath",
		tmpEventsFilePath).Info("Unique EventNames", eventNames)

	mineLog.WithField("tmpEventsFilePath",
		tmpEventsFilePath).Info("Building user and event properties info and writing it to file.")
	eventNamesWithType, err := M.GetEventTypeFromDb(projectId, eventNames, 100000)

	mineLog.WithField("Event Names and type",
		eventNamesWithType).Info("Building user and type from DB.")

	if err != nil {
		mineLog.WithFields(log.Fields{"err": err}).Error("Failed to get event names and Type from DB")
		return nil, nil, err
	}

	return eventNames, eventNamesWithType, nil
}

func buildWhiteListProperties(projectId uint64, allProperty map[string]P.PropertiesCount, numProp int) (map[string]bool, map[string]bool) {

	userPropertiesMap := make(map[string]int)
	eventPropertiesMap := make(map[string]int)
	upFilteredMap := make(map[string]bool)
	epFilteredMap := make(map[string]bool)

	for _, v := range allProperty {
		if strings.Compare(v.PropertyType, "UP") == 0 {
			userPropertiesMap[v.PropertyName] = int(v.Count)
		} else {
			eventPropertiesMap[v.PropertyName] = int(v.Count)
		}
	}

	userPropertiesList, errInt := M.GetAllActiveFactorsTrackedUserPropertiesByProject(projectId)
	if errInt != http.StatusFound {
		mineLog.WithFields(log.Fields{"err": errInt}).Error("Unable to fetch UserProperties from db")
		return nil, nil
	}

	if userPropertiesList != nil && len(userPropertiesList) > 0 {
		mineLog.WithFields(log.Fields{"user properties": userPropertiesList}).Info("Number of User properties from db :", len(userPropertiesList))

		for _, v := range userPropertiesList {
			upFilteredMap[v.UserPropertyName] = true
		}
	} else {
		upSortedList := U.RankByWordCount(userPropertiesMap)

		// restrict number of properties
		if len(upSortedList) > numProp {
			upSortedList = upSortedList[0:numProp]
		}
		// add keys based on ranking
		for _, u := range upSortedList {
			upFilteredMap[u.Key] = true
		}

		//add keys based on WHITELIST_Properties
		for _, v := range U.WHITELIST_FACTORS_USER_PROPERTIES {
			upFilteredMap[v] = true
		}

		//delete keys based on disabled_Properties
		for _, Uprop := range U.DISABLED_FACTORS_USER_PROPERTIES {
			delete(upFilteredMap, Uprop)
		}

		for key := range upFilteredMap {
			mineLog.Info("insert user property", key)
			_, errInt = M.CreateFactorsTrackedUserProperty(projectId, key, "")
			if errInt != http.StatusCreated {
				errorString := fmt.Sprintf("unable to insert user property to db %s", key)
				mineLog.WithFields(log.Fields{"http status": errInt}).Error(errorString)

			}
		}
	}
	//ep : event Properties : addkeys based on ranking , add based on whitelist properties
	// delete based on disables properties

	epSortedList := U.RankByWordCount(eventPropertiesMap)

	if len(epSortedList) > numProp {
		epSortedList = epSortedList[0:numProp]
	}

	for _, u := range epSortedList {
		epFilteredMap[u.Key] = true
	}
	for _, v := range U.WHITELIST_FACTORS_EVENT_PROPERTIES {
		epFilteredMap[v] = true
	}
	for _, Eprop := range U.DISABLED_FACTORS_EVENT_PROPERTIES {
		delete(epFilteredMap, Eprop)
	}

	return upFilteredMap, epFilteredMap
}

func buildEventsFileOnProperties(diskManager *serviceDisk.DiskDriver, cloudManager *filestore.FileManager, projectId uint64,
	modelId uint64, eReader io.Reader, userPropList, eventPropList map[string]bool) error {

	var err error
	efCloudPath, efCloudName := (*cloudManager).GetModelEventsFilePathAndName(projectId, modelId)
	efTmpPath, efTmpName := diskManager.GetModelEventsFilePathAndName(projectId, modelId)
	efPath := efTmpPath + "tmpEvents" + efTmpName
	err = filterAndWriteFactorsProperties(efPath, eReader, userPropList, eventPropList)
	if err != nil {
		mineLog.WithFields(log.Fields{"err": err, "eventFilePath": efCloudPath,
			"eventFileName": efCloudName}).Error("Failed to filter disabled properties")
		return err
	}
	r, err := os.Open(efPath)
	err = diskManager.Create(efTmpPath, efTmpName, r)
	if err != nil {
		mineLog.WithFields(log.Fields{"err": err, "eventFilePath": efTmpPath,
			"eventFileName": efTmpName}).Error("Failed to create event file on disk.")
		return err
	}
	err = os.Remove(efPath)
	if err != nil {
		mineLog.WithFields(log.Fields{"err": err, "eventFilePath": efPath}).Error("Failed to remove file")
		return err

	}

	return nil

}

// PatternMine Mine TOP_K Frequent patterns for every event combination (segment) at every iteration.
func PatternMine(db *gorm.DB, etcdClient *serviceEtcd.EtcdClient, cloudManager *filestore.FileManager,
	diskManager *serviceDisk.DiskDriver, bucketName string, numRoutines int, projectId uint64,
	modelId uint64, modelType string, startTime int64, endTime int64, maxModelSize int64, countOccurence bool) (string, int, error) {

	var err error

	// download events file from cloud to local
	efCloudPath, efCloudName := (*cloudManager).GetModelEventsFilePathAndName(projectId, modelId)
	efTmpPath, efTmpName := diskManager.GetModelEventsFilePathAndName(projectId, modelId)
	mineLog.WithFields(log.Fields{"eventFileCloudPath": efCloudPath,
		"eventFileCloudName": efCloudName}).Info("Downloading events file from cloud.")
	eReader, err := (*cloudManager).Get(efCloudPath, efCloudName)
	if err != nil {
		mineLog.WithFields(log.Fields{"err": err, "eventFilePath": efCloudPath,
			"eventFileName": efCloudName}).Error("Failed downloading events file from cloud.")
		return "", 0, err
	}

	tmpEventsFilepath := efTmpPath + efTmpName
	mineLog.Info("Successfuly downloaded events file from cloud.")

	eventNames, eventNamesWithType, err := GetEventNamesAndType(tmpEventsFilepath, projectId)
	if err != nil {
		mineLog.WithFields(log.Fields{"err": err}).Error("Failed to get eventName and event type.")
		return "", 0, err
	}

	userAndEventsInfo, allPropsMap, err := buildPropertiesInfoFromInput(projectId, eventNames, tmpEventsFilepath)
	userPropList, eventPropList := buildWhiteListProperties(projectId, allPropsMap, topKProperties)
	if err != nil {
		mineLog.WithFields(log.Fields{"err": err}).Error("Failed to build user and event Info.")
		return "", 0, err
	}
	userAndEventsInfoBytes, err := json.Marshal(userAndEventsInfo)
	if err != nil {
		mineLog.WithFields(log.Fields{"err": err}).Error("Failed to unmarshal events Info.")
		return "", 0, err
	}

	repeatedEvents := GetAllRepeatedEvents(eventNames)

	if len(userAndEventsInfoBytes) > 249900000 {
		// Limit is 250MB
		errorString := fmt.Sprintf(
			"Too big properties info, modelId: %d, modelType: %s, projectId: %d, numBytes: %d",
			modelId, modelType, projectId, len(userAndEventsInfoBytes))
		mineLog.Error(errorString)
		return "", 0, fmt.Errorf(errorString)
	}
	err = writeEventInfoFile(projectId, modelId, bytes.NewReader(userAndEventsInfoBytes), (*cloudManager))
	if err != nil {
		mineLog.WithFields(log.Fields{"err": err}).Error("Failed to write events Info.")
		return "", 0, err
	}
	mineLog.Info("Successfully Built user and event properties info and written it to file.")

	err = buildEventsFileOnProperties(diskManager, cloudManager, projectId,
		modelId, eReader, userPropList, eventPropList)
	if err != nil {
		mineLog.WithFields(log.Fields{"err": err}).Error("Failed to write events data.")
		return "", 0, err
	}
	mineLog.Info("Successfully Built events data and written to file.")

	// build histogram of all user properties.
	mineLog.WithField("tmpEventsFilePath", tmpEventsFilepath).Info("Building all user properties histogram.")
	allActiveUsersPattern, err := P.NewPattern([]string{U.SEN_ALL_ACTIVE_USERS}, userAndEventsInfo)
	if err != nil {
		mineLog.WithFields(log.Fields{"err": err}).Error("Failed to build pattern with histogram of all active user properties.")
		return "", 0, err
	}
	if err := computeAllUserPropertiesHistogram(tmpEventsFilepath, allActiveUsersPattern); err != nil {
		mineLog.WithFields(log.Fields{"err": err}).Error("Failed to compute user properties.")
		return "", 0, err
	}
	tmpChunksDir := diskManager.GetPatternChunksDir(projectId, modelId)
	if err := serviceDisk.MkdirAll(tmpChunksDir); err != nil {
		mineLog.WithFields(log.Fields{"chunkDir": tmpChunksDir, "error": err}).Error("Unable to create chunks directory.")
		return "", 0, err
	}
	if err := writePatternsAsChunks([]*P.Pattern{allActiveUsersPattern}, tmpChunksDir); err != nil {
		mineLog.WithFields(log.Fields{"err": err}).Error("Failed to write user properties.")
		return "", 0, err
	}
	mineLog.Info("Successfully built all user properties histogram.")

	// mine and write patterns as chunks
	mineLog.WithFields(log.Fields{"projectId": projectId, "tmpEventsFilepath": tmpEventsFilepath,
		"tmpChunksDir": tmpChunksDir, "routines": numRoutines}).Info("Mining patterns and writing it as chunks.")
	err = mineAndWritePatterns(projectId, tmpEventsFilepath,
		userAndEventsInfo, eventNames, numRoutines, tmpChunksDir, maxModelSize, countOccurence, eventNamesWithType, repeatedEvents)
	if err != nil {
		mineLog.WithFields(log.Fields{"err": err}).Error("Failed to mine patterns.")
		return "", 0, err
	}
	mineLog.Info("Successfully mined patterns and written it as chunks.")

	// upload chunks to cloud
	cloudChunksDir := (*cloudManager).GetPatternChunksDir(projectId, modelId)
	mineLog.WithFields(log.Fields{"tmpChunksDir": tmpChunksDir,
		"cloudChunksDir": cloudChunksDir}).Info("Uploading chunks to cloud.")
	chunkIds, err := uploadChunksToCloud(tmpChunksDir, cloudChunksDir, cloudManager)
	if err != nil {
		mineLog.WithFields(log.Fields{"localChunksDir": tmpChunksDir,
			"cloudChunksDir": cloudChunksDir}).Error("Failed to upload chunks to cloud.")
	}
	mineLog.Info("Successfully uploaded chunks to cloud.")

	// update metadata and notify new version through etcd.
	mineLog.WithFields(log.Fields{
		"ProjectId":      projectId,
		"ModelID":        modelId,
		"ModelType":      modelType,
		"StartTimestamp": startTime,
		"EndTimestamp":   endTime,
		"Chunks":         chunkIds,
	}).Info("Updating mined patterns info to new version of metadata.")
	projectDatas, err := PMM.GetProjectsMetadata(cloudManager, etcdClient)
	if err != nil {
		// failures logged already.
		return "", 0, err
	}
	projectDatas = append(projectDatas, PMM.ProjectData{
		ID:             projectId,
		ModelID:        modelId,
		ModelType:      modelType,
		StartTimestamp: startTime,
		EndTimestamp:   endTime,
		Chunks:         chunkIds,
	})
	newVersionId := fmt.Sprintf("%v", time.Now().Unix())
	err = PMM.WriteProjectDataFile(newVersionId, projectDatas, cloudManager)
	if err != nil {
		mineLog.WithFields(log.Fields{"err": err}).Error("Failed to write new version file to cloud.")
		return "", 0, err
	}
	err = etcdClient.SetProjectVersion(newVersionId)
	if err != nil {
		mineLog.WithFields(log.Fields{"err": err}).Error("Failed to write new version id to etcd.")
		return "", 0, err
	}
	mineLog.WithField("newVersionId", newVersionId).Info("Successfully mined patterns, updated metadata and notified new version id.")

	return newVersionId, len(chunkIds), nil
}

func convert(eventNamesWithAggregation []M.EventNameWithAggregation) []M.EventName {
	eventNames := make([]M.EventName, 0)
	for _, event := range eventNamesWithAggregation {
		eventNames = append(eventNames, M.EventName{
			ID:         event.ID,
			Name:       event.Name,
			CreatedAt:  event.CreatedAt,
			Deleted:    event.Deleted,
			FilterExpr: event.FilterExpr,
			ProjectId:  event.ProjectId,
			Type:       event.Type,
			UpdatedAt:  event.UpdatedAt,
		})
	}
	return eventNames
}

//OpenEventFileAndGetScanner open file to read events
func OpenEventFileAndGetScanner(filePath string) (*bufio.Scanner, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	scanner := store.CreateScannerFromReader(f)
	return scanner, nil
}

func FilterTopKEventsOnTypes(filteredPatterns []*P.Pattern, eventNamesWithType map[string]string, k, keventsSpecial, keventsURL int) []*P.Pattern {

	// take topK from different event types like uc,fe,$types,url etc..
	allPatterns := make([]patternProperties, 0)

	for _, pattern_ := range filteredPatterns {
		var tmpPattern patternProperties

		tmpPattern.pattern = pattern_
		tmpPattern.count = pattern_.PerUserCount
		tmpPattern.patternType = eventNamesWithType[pattern_.EventNames[0]]

		allPatterns = append(allPatterns, tmpPattern)

	}

	ucTopk := takeTopKUC(allPatterns, k)
	feAT_Topk := takeTopKpageView(allPatterns, k)
	ieTopk := takeTopKIE(allPatterns, k)
	specialTopK := takeTopKspecialEvents(allPatterns, keventsSpecial)
	URLTopK := takeTopKAllURL(allPatterns, keventsURL)

	allPatternsFiltered := make([]patternProperties, 0)

	allPatternsFiltered = append(allPatternsFiltered, ucTopk...)
	allPatternsFiltered = append(allPatternsFiltered, feAT_Topk...)
	allPatternsFiltered = append(allPatternsFiltered, ieTopk...)
	allPatternsFiltered = append(allPatternsFiltered, specialTopK...)
	allPatternsFiltered = append(allPatternsFiltered, URLTopK...)

	allPatternsTopk := make([]*P.Pattern, 0)
	exists := make(map[string]bool)
	for _, pt := range allPatternsFiltered {
		if exists[pt.pattern.EventNames[0]] == false {
			allPatternsTopk = append(allPatternsTopk, pt.pattern)
			exists[pt.pattern.EventNames[0]] = true
		}
	}

	return allPatternsTopk

}

func takeTopKUC(allPatterns []patternProperties, topK int) []patternProperties {

	allPatternsType := make([]patternProperties, 0)
	for _, pattern := range allPatterns {

		if pattern.patternType == M.TYPE_USER_CREATED_EVENT_NAME {
			allPatternsType = append(allPatternsType, pattern)
		}
	}

	if len(allPatternsType) > 0 {
		return takeTopK(allPatternsType, topK)
	}
	return allPatternsType

}

func takeTopKpageView(allPatterns []patternProperties, topK int) []patternProperties {

	allPatternsType := make([]patternProperties, 0)
	for _, pattern := range allPatterns {

		if pattern.patternType == M.TYPE_FILTER_EVENT_NAME || pattern.patternType == M.TYPE_AUTO_TRACKED_EVENT_NAME {
			allPatternsType = append(allPatternsType, pattern)
		}
	}
	if len(allPatternsType) > 0 {
		return takeTopK(allPatternsType, topK)
	}
	return allPatternsType

}

func takeTopKIE(allPatterns []patternProperties, topK int) []patternProperties {

	allPatternsType := make([]patternProperties, 0)
	for _, pattern := range allPatterns {

		if pattern.patternType == M.TYPE_INTERNAL_EVENT_NAME {
			allPatternsType = append(allPatternsType, pattern)
		}
	}
	if len(allPatternsType) > 0 {
		return takeTopK(allPatternsType, topK)
	}
	return allPatternsType

}

func takeTopKspecialEvents(allPatterns []patternProperties, topK int) []patternProperties {

	allPatternsType := make([]patternProperties, 0)
	for _, pt := range allPatterns {

		if strings.HasPrefix(pt.pattern.EventNames[0], "$") == true {
			allPatternsType = append(allPatternsType, pt)
		}
	}
	if len(allPatternsType) > 0 {
		return takeTopK(allPatternsType, topK)
	}
	return allPatternsType

}

func takeTopKAllURL(allPatterns []patternProperties, topK int) []patternProperties {

	allPatternsType := make([]patternProperties, 0)
	for _, pt := range allPatterns {

		if U.IsValidUrl(pt.pattern.EventNames[0]) == true {
			allPatternsType = append(allPatternsType, pt)
		}
	}
	if len(allPatternsType) > 0 {
		return takeTopK(allPatternsType, topK)
	}
	return allPatternsType

}

func takeTopK(patterns []patternProperties, topKPatterns int) []patternProperties {
	//rewrite with heap. can hog the memory
	if len(patterns) > 0 {
		sort.Slice(patterns, func(i, j int) bool { return patterns[i].count > patterns[j].count })
		if len(patterns) > topKPatterns {
			return patterns[0:topKPatterns]
		}
		return patterns

	}
	return patterns
}

//GetAllCyclicEvents Filter all special events
func GetAllRepeatedEvents(eventNames []string) []string {

	CyclicEvents := make([]string, 0)
	for _, v := range eventNames {

		if strings.HasPrefix(v, "$") == true {
			CyclicEvents = append(CyclicEvents, v)
		}

	}

	return CyclicEvents

}
