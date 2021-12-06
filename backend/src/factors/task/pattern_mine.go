package task

import (
	"bufio"
	"bytes"
	"encoding/json"
	"factors/filestore"
	"factors/model/model"
	"factors/model/store"
	P "factors/pattern"
	patternStore "factors/pattern_server/store"
	serviceDisk "factors/services/disk"
	serviceEtcd "factors/services/etcd"
	U "factors/util"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"reflect"
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
const topK_patterns = 5
const topK_campaigns = 5
const topKProperties = 20
const keventsSpecial = 5
const keventsURL = 10
const max_PATTERN_LENGTH = 3

const max_CHUNK_SIZE_IN_BYTES int64 = 200 * 1000 * 1000 // 200MB

// quota counts for URLS, user defined events,smart events ,
// standard events, campaigns, source, referrer, medium, adgroup
const countURL = 25
const countUDE = 20
const countSME = 5
const countStdEvents = -1 // all events
const countCampaigns = 25
const countSource = 0
const countReferrer = 0
const countMedium = 0
const countAdgroup = 0
const USER_PROP_MIN_SUPPORT = 2.0
const EVENT_PROP_MIN_SUPPORT = 20.0

var regex_NUM = regexp.MustCompile("[0-9]+")
var mineLog = taskLog.WithField("prefix", "Task#PatternMine")

type patternProperties struct {
	pattern     *P.Pattern
	count       uint
	patternType string
}

func (pp patternProperties) Get_patternEventNames() []string {
	return pp.pattern.EventNames
}
func (pp patternProperties) Get_count() uint {
	return pp.count
}
func (pp patternProperties) Get_patternType() string {
	return pp.patternType
}

type CampaignEventLists struct {
	CampaignList []string
	MediumList   []string
	ReferrerList []string
	SourceList   []string
	AdgroupList  []string
}

func countPatternsWorker(projectID uint64, filepath string,
	patterns []*P.Pattern, wg *sync.WaitGroup, countOccurence bool, countsVersion int) {
	mineLog.Infof("Counting patterns from File: %s", filepath)
	file, err := os.Open(filepath)
	if err != nil {
		mineLog.WithField("filePath", filepath).Error("Failure on count pattern workers.")
	}

	scanner := bufio.NewScanner(file)
	buf := make([]byte, P.MAX_PATTERN_BYTES)
	scanner.Buffer(buf, P.MAX_PATTERN_BYTES)
	P.CountPatterns(projectID, scanner, patterns, countOccurence, countsVersion)
	file.Close()
	wg.Done()
}

func countPatterns(projectID uint64, filepath string, patterns []*P.Pattern, numRoutines int, countOccurence bool, countsVersion int) {
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
		go countPatternsWorker(projectID, filepath, patterns[low:high], &wg, countOccurence, countsVersion)
	}
	wg.Wait()
}

func computeAllUserPropertiesHistogram(projectID uint64, filepath string, pattern *P.Pattern, countsVersion int) error {
	file, err := os.Open(filepath)
	pattern.PatternVersion = countsVersion
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	// 10 MB buffer.
	const maxCapacity = 10 * 1024 * 1024
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)
	return P.ComputeAllUserPropertiesHistogram(projectID, scanner, pattern, countsVersion)
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
		if pattern.PerOccurrenceEventCategoricalProperties != nil && pattern.PerUserEventCategoricalProperties != nil {
			(*pattern.PerUserEventCategoricalProperties).TrimByFmapSize(trimFraction)
			(*pattern.PerUserUserCategoricalProperties).TrimByFmapSize(trimFraction)
			(*pattern.PerOccurrenceEventCategoricalProperties).TrimByFmapSize(trimFraction)
			(*pattern.PerOccurrenceUserCategoricalProperties).TrimByFmapSize(trimFraction)
		}
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
		if pattern.PerOccurrenceEventNumericProperties != nil && pattern.PerUserEventNumericProperties != nil {
			(*pattern.PerUserEventNumericProperties).TrimByBinSize(trimFraction)
			(*pattern.PerUserUserNumericProperties).TrimByBinSize(trimFraction)
			(*pattern.PerOccurrenceEventNumericProperties).TrimByBinSize(trimFraction)
			(*pattern.PerOccurrenceUserNumericProperties).TrimByBinSize(trimFraction)
		}
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
		if pattern.PerOccurrenceEventCategoricalProperties != nil && pattern.PerUserUserCategoricalProperties != nil {
			(*pattern.PerUserEventCategoricalProperties).TrimByBinSize(trimFraction)
			(*pattern.PerUserUserCategoricalProperties).TrimByBinSize(trimFraction)
			(*pattern.PerOccurrenceEventCategoricalProperties).TrimByBinSize(trimFraction)
			(*pattern.PerOccurrenceUserCategoricalProperties).TrimByBinSize(trimFraction)
		}
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
	mineLog.Info("Number of repeapted Events in genSegmentdCandidates : ", len(cyclicEvents))

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

func InitCampaignAnalyticsPatterns(smartEvents CampaignEventLists) ([]*P.Pattern, error) {

	var lenOnePatterns []*P.Pattern

	for _, eventName := range smartEvents.CampaignList {
		p, err := P.NewPattern([]string{eventName}, nil)
		if err != nil {
			return []*P.Pattern{}, fmt.Errorf("campaign Pattern initialization failed")
		}
		lenOnePatterns = append(lenOnePatterns, p)
	}

	for _, eventName := range smartEvents.MediumList {
		p, err := P.NewPattern([]string{eventName}, nil)
		if err != nil {
			return []*P.Pattern{}, fmt.Errorf("medium pattern initialization failed")
		}
		lenOnePatterns = append(lenOnePatterns, p)
	}

	for _, eventName := range smartEvents.ReferrerList {
		p, err := P.NewPattern([]string{eventName}, nil)
		if err != nil {
			return []*P.Pattern{}, fmt.Errorf("referrer pattern initialization failed")
		}
		lenOnePatterns = append(lenOnePatterns, p)
	}

	for _, eventName := range smartEvents.SourceList {
		p, err := P.NewPattern([]string{eventName}, nil)
		if err != nil {
			return []*P.Pattern{}, fmt.Errorf("source pattern initialization failed")
		}
		lenOnePatterns = append(lenOnePatterns, p)
	}

	for _, eventName := range smartEvents.AdgroupList {
		p, err := P.NewPattern([]string{eventName}, nil)
		if err != nil {
			return []*P.Pattern{}, fmt.Errorf("AdGroup Pattern initialization failed")
		}
		lenOnePatterns = append(lenOnePatterns, p)
	}

	return lenOnePatterns, nil
}

// mineAndWriteLenOnePatterns : All the len one events in the events file is counted which includes
// standard events, URLs , campaignType events
func mineAndWriteLenOnePatterns(projectID uint64, modelId uint64, cloudManager *filestore.FileManager,
	eventNames []string, filepathString string,
	userAndEventsInfo *P.UserAndEventsInfo, numRoutines int,
	chunkDir string, maxModelSize int64, cumulativePatternsSize int64,
	countOccurence bool, campaignTypeEvents CampaignEventLists, efTmpPath string, beamConfig *RunBeamConfig, countsVersion int) (
	[]*P.Pattern, int64, error) {
	var lenOnePatterns []*P.Pattern
	var filteredLenOnePatterns []*P.Pattern
	var patternsSize int64
	var err error

	for _, eventName := range eventNames {
		p, err := P.NewPattern([]string{eventName}, userAndEventsInfo)
		if err != nil {
			return []*P.Pattern{}, 0, fmt.Errorf("pattern initialization failed in mineLenOne")
		}
		lenOnePatterns = append(lenOnePatterns, p)
	}
	lenOnePatternsCampigns, _ := InitCampaignAnalyticsPatterns(campaignTypeEvents)
	lenOnePatterns = append(lenOnePatterns, lenOnePatternsCampigns...)
	// countPatterns(projectID, filepath, lenOnePatterns, numRoutines, countOccurence, countsVersion))

	if !beamConfig.RunOnBeam {
		countPatterns(projectID, filepathString, lenOnePatterns, numRoutines, countOccurence, countsVersion)
		filteredLenOnePatterns, patternsSize, err = filterAndCompressPatterns(
			lenOnePatterns, maxModelSize, cumulativePatternsSize, 1, max_PATTERN_LENGTH)

		if err != nil {
			return []*P.Pattern{}, 0, err
		}
	} else {
		//call beam
		scopeName := "Count_len_one"
		patternsFpath, err := countPatternController(beamConfig, projectID, modelId,
			cloudManager, filepathString,
			lenOnePatterns, numRoutines, userAndEventsInfo, countOccurence,
			efTmpPath, scopeName, countsVersion)
		if err != nil {
			return nil, 0, err
		}
		mineLog.Info("Reading from file")
		filteredLenOnePatterns, patternsSize, err = ReadFilterAndCompressPatternsFromFile(
			patternsFpath, cloudManager, maxModelSize, cumulativePatternsSize, 1, max_PATTERN_LENGTH)
		if err != nil {
			return []*P.Pattern{}, 0, err
		}
		mineLog.Infof("Beam number of len one patterns : %d , ", len(lenOnePatterns))
	}

	mineLog.Infof("patternSize : %d , ", patternsSize)

	if err := writePatternsAsChunks(filteredLenOnePatterns, chunkDir); err != nil {
		return []*P.Pattern{}, 0, err
	}
	return filteredLenOnePatterns, patternsSize, nil
}

//FilterCombinationPatterns filter all start events based on topK logic for URL's,UDE,standardEvents,CampAnalytics
func FilterCombinationPatterns(combinationGoalPatterns, goalPatterns []*P.Pattern, eventNamesWithType map[string]string) []*P.Pattern {
	combinationPatternsMap := make(map[string][]*P.Pattern)
	allPatterns := make([]*P.Pattern, 0)

	//get all patterns ending in goals
	for _, v := range combinationGoalPatterns {
		combinationPatternsMap[v.EventNames[1]] = append(combinationPatternsMap[v.EventNames[1]], v)
	}

	//for each goal pattern Get top URL,UDE,StandardEvents, campAnalytics
	for _, v := range goalPatterns {
		endGoalPatterns := combinationPatternsMap[v.EventNames[0]]
		mineLog.Info("Number of patterns in ", v.EventNames[0], "->", len(endGoalPatterns))
		if len(endGoalPatterns) > 0 {
			tmpGoalPatterns, err := FilterAllCausualEvents(endGoalPatterns, eventNamesWithType)
			if err != nil {
				mineLog.Error(err)
			}
			mineLog.Info("Number of patterns after filtering ", v.EventNames[0], len(tmpGoalPatterns))
			allPatterns = append(allPatterns, tmpGoalPatterns...)
		} else {
			mineLog.Error("No patterns for Goal", v.EventNames[0])
		}
	}
	return allPatterns
}

func mineAndWriteLenTwoPatterns(projectId uint64, modelId uint64,
	lenOnePatterns []*P.Pattern, filepathString string, cloudManager *filestore.FileManager,
	userAndEventsInfo *P.UserAndEventsInfo, numRoutines int,
	chunkDir string, maxModelSize int64, cumulativePatternsSize int64, countOccurence bool,
	goalPatterns []*P.Pattern, repeatedEventsList []string, eventNamesWithType map[string]string,
	beamConfig *RunBeamConfig, efTmpPath string, countsVersion int) (
	[]*P.Pattern, int64, error) {

	var patternsSize int64
	var filteredLenTwoPatterns []*P.Pattern
	var err error
	var patternsFpath string
	combinationGoalPatterns, _, err := P.GenCombinationPatternsEndingWithGoal(lenOnePatterns, goalPatterns, userAndEventsInfo)
	if err != nil {
		return nil, 0, err
	}

	if !beamConfig.RunOnBeam {
		countPatterns(projectId, filepathString, combinationGoalPatterns, numRoutines, countOccurence, countsVersion)
		filteredLenTwoPatterns, patternsSize, err = filterAndCompressPatterns(
			combinationGoalPatterns, maxModelSize, cumulativePatternsSize, 2, max_PATTERN_LENGTH)
		if err != nil {
			return []*P.Pattern{}, 0, err
		}
	} else {

		//call beam
		scopeName := "Count_len_two"
		patternsFpath, err = countPatternController(beamConfig, projectId, modelId, cloudManager, filepathString,
			combinationGoalPatterns, numRoutines, userAndEventsInfo, countOccurence,
			efTmpPath, scopeName, countsVersion)
		if err != nil {
			return nil, 0, fmt.Errorf("error in counting len two patterns in beam :%v", err)
		}
		mineLog.Info("Reading from file")
		filteredLenTwoPatterns, patternsSize, err = ReadFilterAndCompressPatternsFromFile(
			patternsFpath, cloudManager, maxModelSize, cumulativePatternsSize, 2, max_PATTERN_LENGTH)
		if err != nil {
			return []*P.Pattern{}, 0, err
		}
		mineLog.Infof("number len Two patterns from beam: %d , ", len(combinationGoalPatterns))

	}

	// filter all combinationGoalPatterns based on start event.
	// for each goal event based on quota logic filter patterns.
	// based on the quota  startevent filter url, sme, campaign events etc.
	goalFilteredLenTwoPatterns := FilterCombinationPatterns(filteredLenTwoPatterns, goalPatterns, eventNamesWithType)

	if err := writePatternsAsChunks(goalFilteredLenTwoPatterns, chunkDir); err != nil {
		return []*P.Pattern{}, 0, err
	}
	mineLog.Info("Total Numnber of Len Two Patterns :", len(goalFilteredLenTwoPatterns))
	return goalFilteredLenTwoPatterns, patternsSize, nil
}

// GetGoalPatterns get all goalPatterns from DB
func GetGoalPatterns(projectId uint64, filteredPatterns []*P.Pattern, eventNamesWithType map[string]string, campEventsType CampaignEventLists, userAndEventsInfo *P.UserAndEventsInfo) ([]*P.Pattern, error) {

	goalPatternsFromDB, errCode := store.GetStore().GetAllActiveFactorsTrackedEventsByProject(projectId)

	if errCode != http.StatusFound {
		mineLog.Info("Failure on Get goal patterns.")
	}
	var goalPatterns []*P.Pattern
	if len(goalPatternsFromDB) > 0 {

		mineLog.Info(fmt.Sprintf("Number of Goals from DB:%d", len(goalPatternsFromDB)))

		tmpPatterns := make(map[string]*P.Pattern)

		for _, v := range filteredPatterns {
			tmpPatterns[v.String()] = v
		}
		for _, p := range goalPatternsFromDB {
			if valPattern, ok := tmpPatterns[p.Name]; ok {
				goalPatterns = append(goalPatterns, valPattern)
				mineLog.Infof("Goal event from db :%s", valPattern.String())
			} else {
				mineLog.Infof("Goal event from db not available in filtered lenOne Pattern :%s", p.Name)
				if _, ok := eventNamesWithType[p.Name]; ok {
					p, err := P.NewPattern([]string{p.Name}, userAndEventsInfo)
					if err != nil {
						return []*P.Pattern{}, nil
					}
					goalPatterns = append(goalPatterns, p)
					mineLog.Info(fmt.Sprint("Goal event from db added from events in eventsFile ", valPattern.String()))

				}

			}
		}

		return goalPatterns, nil
	}

	goalTopKPatterns := FilterTopKEventsOnTypes(filteredPatterns, eventNamesWithType, topK_patterns, keventsSpecial, keventsURL, campEventsType)
	mineLog.Info("Mining goals from topk events : ", len(goalTopKPatterns))

	for idx, valPat := range goalTopKPatterns {
		mineLog.Info(fmt.Sprint("Insering in DB Goal event: ", idx, valPat.String()))
		tmpFactorsRule := model.FactorsGoalRule{EndEvent: valPat.String()}
		goalID, httpStatusTrackedEvent := store.GetStore().CreateFactorsTrackedEvent(projectId, valPat.String(), "")
		if goalID == 0 {
			mineLog.Info("Unable to create a trackedEvent ", httpStatusTrackedEvent, " ", goalID)
		}
		mineLog.Info("trackedEvent in db  ", httpStatusTrackedEvent, " ", valPat.String(), " ", goalID)

		_, httpstatus, err := store.GetStore().CreateFactorsGoal(projectId, valPat.String(), tmpFactorsRule, "")
		if httpstatus != http.StatusCreated {
			mineLog.Info("Unable to create factors goal in db: ", httpstatus, " ", valPat.String(), " ", err)
		}

	}

	return goalTopKPatterns, nil

}

//GenMissingTwoLenPatterns get threeLen-twoLen
func GenMissingJourneyPatterns(goal, journey []*P.Pattern, userAndEventsInfo *P.UserAndEventsInfo) ([]*P.Pattern, error) {
	lenGoal := len(goal[0].EventNames)
	lenJourney := len(journey[0].EventNames)
	if lenJourney > lenGoal {
		return nil, fmt.Errorf("len of Journey is greater than goal")
	}

	journeyPatt := make(map[string]*P.Pattern)
	missingPatt := make([]*P.Pattern, 0)
	sep := "_"
	for _, v := range journey {
		journeyPatt[strings.Join(v.EventNames, sep)] = v
	}

	for _, v := range goal {

		tmpString := strings.Join(v.EventNames[0:lenGoal-1], sep)
		if _, ok := journeyPatt[tmpString]; !ok {

			tmpPatt, err := P.NewPattern(v.EventNames[0:lenGoal-1], userAndEventsInfo)
			if err != nil {
				return []*P.Pattern{}, fmt.Errorf("unable to generate new n+1 len Pattern")
			}
			missingPatt = append(missingPatt, tmpPatt)
		}
	}

	mineLog.Info("Number of missing patterns before filtering :", len(missingPatt))

	allMissingPatt := make([]*P.Pattern, 0)
	tmpMissingMap := make(map[string]bool)
	// add dedupe logic
	for _, v := range missingPatt {
		if !tmpMissingMap[strings.Join(v.EventNames, "_")] {
			allMissingPatt = append(allMissingPatt, v)
			tmpMissingMap[strings.Join(v.EventNames, "_")] = true
		}

	}
	mineLog.Info("Number of missing patterns :", len(allMissingPatt))

	return allMissingPatt, nil
}

func mineAndWritePatterns(projectId uint64, modelId uint64, filepath string,
	userAndEventsInfo *P.UserAndEventsInfo, cloudManager *filestore.FileManager, eventNames []string,
	numRoutines int, chunkDir string,
	maxModelSize int64, countOccurence bool,
	eventNamesWithType map[string]string, repeatedEvents []string,
	campTypeEvents CampaignEventLists, efTmpPath string, beamConfig *RunBeamConfig, countsVersion int) error {
	var filteredPatterns []*P.Pattern
	var cumulativePatternsSize int64 = 0

	patternLen := 1
	limitRoundOffFraction := 0.99

	filteredPatterns, patternsSize, err := mineAndWriteLenOnePatterns(projectId, modelId, cloudManager,
		eventNames, filepath, userAndEventsInfo, numRoutines, chunkDir,
		maxModelSize, cumulativePatternsSize, countOccurence, campTypeEvents, efTmpPath, beamConfig, countsVersion)
	if err != nil {
		return fmt.Errorf("unable to mine len one patterns :%v", err)
	}
	cumulativePatternsSize += patternsSize
	printFilteredPatterns(filteredPatterns, patternLen)
	mineLog.Info("Number of Len-one Patterns : ", len(filteredPatterns))

	// Get all Goal Patterns => DB Patterns + CampaignAnalytics (Campaign,source,medium,referr)
	goalPatterns, err := GetGoalPatterns(projectId, filteredPatterns, eventNamesWithType, campTypeEvents, userAndEventsInfo)
	if err != nil {
		return fmt.Errorf("unable to get goal patterns :%v", err)
	}
	mineLog.Info("Number of Goal Patterns to use in factors: ", len(goalPatterns))
	if cumulativePatternsSize >= int64(float64(maxModelSize)*limitRoundOffFraction) {
		return nil
	}
	for _, v := range goalPatterns {
		mineLog.Info("Goal Patterns :->", v.EventNames)
	}

	patternLen++
	if patternLen > max_PATTERN_LENGTH {
		return nil
	}
	filteredTwoPatterns, patternsSize, err := mineAndWriteLenTwoPatterns(projectId, modelId,
		filteredPatterns, filepath, cloudManager, userAndEventsInfo,
		numRoutines, chunkDir, maxModelSize, cumulativePatternsSize,
		countOccurence, goalPatterns, repeatedEvents, eventNamesWithType,
		beamConfig, efTmpPath, countsVersion)
	if err != nil {
		return err
	}
	cumulativePatternsSize += patternsSize
	printFilteredPatterns(filteredTwoPatterns, patternLen)

	generatedThreePatterns, err := GenInterMediateCombinations(filteredTwoPatterns, userAndEventsInfo)
	if err != nil {
		return fmt.Errorf("error to creating intermediate Patterns %v", err)
	}
	generatedThreeRepeatedPatterns, err := GenRepeatedCombinations(filteredTwoPatterns, userAndEventsInfo, repeatedEvents)
	if err != nil {
		return fmt.Errorf("error to creating Repeated intermediate Patterns %v", err)
	}

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
			filteredTwoPatterns, userAndEventsInfo, repeatedEvents)
		if err != nil {
			return err
		}

		lenThreeCampaign, err := GenCampaignThreeLenCombinations(filteredTwoPatterns, goalPatterns, userAndEventsInfo, topK_campaigns)
		if err != nil {
			return fmt.Errorf("unable to gen three len combinations :%v", err)
		}
		lenThreePatterns := []*P.Pattern{}
		for _, patterns := range lenThreeSegmentedPatterns {
			lenThreePatterns = append(lenThreePatterns, patterns...)
		}

		lenThreePatterns = MergePatterns(lenThreePatterns, generatedThreePatterns)
		lenThreePatterns = MergePatterns(lenThreePatterns, generatedThreeRepeatedPatterns)
		lenThreePatterns = MergePatterns(lenThreePatterns, lenThreeCampaign)

		var filteredThreePatterns []*P.Pattern
		if !beamConfig.RunOnBeam {
			countPatterns(projectId, filepath, lenThreePatterns, numRoutines, countOccurence, countsVersion)
			filteredThreePatterns, patternsSize, err = filterAndCompressPatterns(
				lenThreePatterns, maxModelSize, cumulativePatternsSize,
				patternLen, max_PATTERN_LENGTH)
			if err != nil {
				return err
			}
		} else {
			//call beam
			var patternsFpath string
			scopeName := "Count_len_three"
			patternsFpath, err = countPatternController(beamConfig, projectId, modelId, cloudManager, filepath,
				lenThreePatterns, numRoutines, userAndEventsInfo, countOccurence,
				efTmpPath, scopeName, countsVersion)
			if err != nil {
				return fmt.Errorf("unable to count three len patterns :%v", err)
			}
			mineLog.Info("Reading from file")
			filteredThreePatterns, patternsSize, err = ReadFilterAndCompressPatternsFromFile(
				patternsFpath, cloudManager, maxModelSize, cumulativePatternsSize,
				patternLen, max_PATTERN_LENGTH)
			if err != nil {
				return err
			}
			mineLog.Infof("number len Three patterns from beam: %d , ", len(lenThreePatterns))

		}

		cumulativePatternsSize += patternsSize
		if err := writePatternsAsChunks(filteredThreePatterns, chunkDir); err != nil {
			return err
		}
		printFilteredPatterns(filteredThreePatterns, patternLen)

		// count missing two len patterns in three len patterns
		missingPatternsTwo, err := GenMissingJourneyPatterns(filteredThreePatterns, filteredTwoPatterns, userAndEventsInfo)

		if err != nil {
			mineLog.Error("Unable to create missing two len pattern")
		}

		// count  - two len missing
		var filteredMissingPatterns []*P.Pattern
		if !beamConfig.RunOnBeam {
			countPatterns(projectId, filepath, missingPatternsTwo, numRoutines, countOccurence, countsVersion)
			filteredMissingPatterns, patternsSize, err = filterAndCompressPatterns(
				missingPatternsTwo, maxModelSize, cumulativePatternsSize,
				patternLen, max_PATTERN_LENGTH)
			if err != nil {
				return err
			}
		} else {
			//call beam
			var patternsFpath string
			scopeName := "CountTwoMissing"
			patternsFpath, err = countPatternController(beamConfig, projectId,
				modelId, cloudManager, filepath,
				missingPatternsTwo, numRoutines, userAndEventsInfo,
				countOccurence, efTmpPath, scopeName, countsVersion)
			if err != nil {
				return fmt.Errorf("unable to count len two missing :%v", err)
			}
			mineLog.Infof("Reading from part Directory : %s", scopeName)
			filteredMissingPatterns, patternsSize, err = ReadFilterAndCompressPatternsFromFile(patternsFpath, cloudManager, maxModelSize, cumulativePatternsSize, patternLen, max_PATTERN_LENGTH)
			if err != nil {
				return err
			}
			mineLog.Infof("number len Two missing patterns from beam: %d , ", len(missingPatternsTwo))
		}

		cumulativePatternsSize += patternsSize
		if err := writePatternsAsChunks(filteredMissingPatterns, chunkDir); err != nil {
			return err
		}
		printFilteredPatterns(filteredMissingPatterns, patternLen-1)

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
		countPatterns(projectId, filepath, candidatePatterns, numRoutines, countOccurence, countsVersion)
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

	scanner, err := OpenEventFileAndGetScanner(filepath)
	if err != nil {
		return nil, nil, err
	}
	allProperty, err := P.CollectPropertiesInfo(projectId, scanner, userAndEventsInfo)
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

		patternWithMeta := patternStore.PatternWithMeta{
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
		if pBytesLen >= 10000000*2 {
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
			if err != nil {
				return err
			}
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

	uploadedChunkIds := make([]string, 0)
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

func FilterTopEvents(eventsMap map[string]int, limitCount int, eventclass string) []string {
	eventsList := make([]string, 0)
	eventsListWithCounts := U.RankByWordCount(eventsMap)
	if limitCount > 0 && len(eventsListWithCounts) > limitCount {
		for key, val := range eventsListWithCounts {
			eventsList = append(eventsList, val.Key)
			if key >= limitCount-1 {
				break
			}
		}
	} else {
		for _, val := range eventsListWithCounts {
			eventsList = append(eventsList, val.Key)
		}

	}
	return eventsList
}

func rewriteEventsFile(tmpEventsFilePath string, tmpPath string, userPropMap, eventPropMap map[string]bool,
	upCount, epCount map[string]map[string]int, upCatg, epCatg map[string]string) (CampaignEventLists, error) {
	// read events file , filter and create properties based on userProp and eventsProp
	// create encoded events based on $session and campaign eventName

	var smartEvents CampaignEventLists
	mineLog.WithField("path", tmpEventsFilePath).Info("Read Events from file to create encoded events")
	scanner, err := OpenEventFileAndGetScanner(tmpEventsFilePath)
	if err != nil {
		log.Error("Unable to open events File")
		return CampaignEventLists{}, err
	}

	mineLog.WithField("path", tmpPath).Info("Create a temp file to save and read events")
	file, err := os.Create(tmpPath)
	if err != nil {
		return CampaignEventLists{}, fmt.Errorf("unable to create a tmp file :%v", err)
	}
	defer file.Close()

	if err != nil {
		log.Error("Unable to create temp File")
		return CampaignEventLists{}, err
	}
	campEventsMap := make(map[string]int)
	mediumEventsMap := make(map[string]int)
	sourceEventsMap := make(map[string]int)
	referrerEventsMap := make(map[string]int)
	AdgroupEventsMap := make(map[string]int)

	mineLog.WithField("model user properties", userPropMap).Info("Final User properties to model")
	mineLog.WithField("model event Properties", eventPropMap).Info("Final Event Properties to model")

	delEvent := 0
	delUser := 0
	w := bufio.NewWriter(file)
	for scanner.Scan() {
		line := scanner.Text()
		var eventDetails P.CounterEventFormat
		if err := json.Unmarshal([]byte(line), &eventDetails); err != nil {
			log.WithFields(log.Fields{"line": line, "err": err}).Error("Read failed")
			return CampaignEventLists{}, err
		}
		ename := eventDetails.EventName

		for uKey, _ := range eventDetails.UserProperties {
			if _, ok := userPropMap[uKey]; !ok {
				delete(eventDetails.UserProperties, uKey)
			}
		}

		for uKey, uVal := range eventDetails.UserProperties {
			if catg, ok := upCatg[uKey]; ok {
				if catg == U.PropertyTypeCategorical {
					pstring := U.GetPropertyValueAsString(uVal)
					propString := strings.Join([]string{uKey, pstring}, "::")
					if _, ok := upCount[ename][propString]; !ok {
						delete(eventDetails.UserProperties, uKey)
						delUser += 1
					}
					if removePropertiesTree(pstring) {
						delete(eventDetails.UserProperties, uKey)
						delUser += 1
					}
				}

			}
		}

		for eKey, _ := range eventDetails.EventProperties {
			if _, ok := eventPropMap[eKey]; !ok {
				delete(eventDetails.EventProperties, eKey)
			}
		}

		for eKey, eVal := range eventDetails.EventProperties {
			if catg, ok := epCatg[eKey]; ok {
				if catg == U.PropertyTypeCategorical {
					pstring := U.GetPropertyValueAsString(eVal)
					propString := strings.Join([]string{eKey, pstring}, "::")
					if _, ok := epCount[ename][propString]; !ok {
						delete(eventDetails.EventProperties, eKey)
						delEvent += 1
					}

					if removePropertiesTree(pstring) {
						delete(eventDetails.EventProperties, eKey)
						delUser += 1
					}
				}

			}
		}

		eventDetailsBytes, err := json.Marshal(eventDetails)

		if err != nil {
			log.WithFields(log.Fields{"line": line, "err": err}).Error("Failed to marshal eventDetails")
			return CampaignEventLists{}, err
		}

		lineWrite := string(eventDetailsBytes)
		if _, err := file.WriteString(fmt.Sprintf("%s\n", lineWrite)); err != nil {
			peLog.WithFields(log.Fields{"line": line, "err": err}).Error("Unable to write to file.")
			return CampaignEventLists{}, err
		}

		if err := writeEncodedEvent(U.EVENT_NAME_SESSION, U.EP_CAMPAIGN, "campaign", campEventsMap, eventDetails, file, line); err != nil {
			return CampaignEventLists{}, err
		}
		if err := writeEncodedEvent(U.EVENT_NAME_SESSION, U.EP_MEDIUM, "medium", mediumEventsMap, eventDetails, file, line); err != nil {
			return CampaignEventLists{}, err
		}
		if err := writeEncodedEvent(U.EVENT_NAME_SESSION, U.EP_SOURCE, "source", sourceEventsMap, eventDetails, file, line); err != nil {
			return CampaignEventLists{}, err
		}
		if err := writeEncodedEvent(U.EVENT_NAME_SESSION, U.SP_INITIAL_REFERRER, "initial_referrer", referrerEventsMap, eventDetails, file, line); err != nil {
			return CampaignEventLists{}, err
		}
		if err := writeEncodedEvent(U.EVENT_NAME_SESSION, U.EP_ADGROUP, "adgroup", AdgroupEventsMap, eventDetails, file, line); err != nil {
			return CampaignEventLists{}, err
		}

	}
	w.Flush()

	smartEvents.CampaignList = FilterTopEvents(campEventsMap, -1, "Campaign")
	smartEvents.SourceList = FilterTopEvents(sourceEventsMap, -1, "Source")
	smartEvents.MediumList = FilterTopEvents(mediumEventsMap, -1, "Medium")
	smartEvents.ReferrerList = FilterTopEvents(referrerEventsMap, -1, "Referrer")
	smartEvents.AdgroupList = FilterTopEvents(AdgroupEventsMap, -1, "AdGroup")

	err = os.Remove(tmpEventsFilePath)
	mineLog.WithField("path", tmpEventsFilePath).Info("Remove tmpEvents File")
	if err != nil {
		mineLog.WithField("path", tmpEventsFilePath).Error("unable to remove File", err)
		return CampaignEventLists{}, err
	}
	mineLog.Infof("number of user prop deleted:%d", delUser)
	mineLog.Infof("number of event prop deleted:%d", delEvent)
	return smartEvents, nil
}

func writeEncodedEvent(eventName string, property string, propertyName string, propEventsMap map[string]int, eventDetails P.CounterEventFormat, file *os.File, line string) error {
	//check if eventName and eventDetails.EventName match and property exists in eventDetails.EventProperties
	//if exists with non-empty value, write encoded event to file

	if strings.Compare(eventDetails.EventName, eventName) == 0 && eventDetails.EventProperties[property] != nil {

		var tmpEvent P.CounterEventFormat
		tmpEventId, ok := eventDetails.EventProperties[property].(string)

		if !ok {
			mineLog.Info("Error in converting string : ", " ", strings.ToUpper(propertyName), " ", eventDetails.EventProperties[property])
		}

		if len(tmpEventId) > 0 {
			tmpEventName := eventDetails.EventName + "[" + propertyName + ":" + tmpEventId + "]"
			tmpEvent.EventName = tmpEventName
			propEventsMap[tmpEventName] = propEventsMap[tmpEventName] + 1
			tmpEvent.EventProperties = nil
			tmpEvent.UserProperties = nil
			tmpEvent.UserId = eventDetails.UserId
			tmpEvent.UserJoinTimestamp = eventDetails.UserJoinTimestamp
			tmpEvent.EventTimestamp = eventDetails.EventTimestamp
			tmpEvent.EventCardinality = eventDetails.EventCardinality
			eventDetailsBytes, _ := json.Marshal(tmpEvent)
			lineWrite := string(eventDetailsBytes)

			if _, err := file.WriteString(fmt.Sprintf("%s\n", lineWrite)); err != nil {
				peLog.WithFields(log.Fields{"line": line, "property": property, "err": err}).Error("Unable to write to file.")
				return err
			}
		}
	}
	return nil
}

func GetEventNamesAndType(tmpEventsFilePath string, projectId uint64, countsVersion int) ([]string, map[string]string, map[string]map[string]int, map[string]map[string]int, map[string]string, map[string]string, error) {
	scanner, err := OpenEventFileAndGetScanner(tmpEventsFilePath)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}
	eventNames, upCount, epCount, upCatg, epCatg, err := GetEventNamesFromFile(scanner, projectId, countsVersion)
	if err != nil {
		mineLog.WithFields(log.Fields{"err": err, "eventFilePath": tmpEventsFilePath}).Error("Failed to read event names from file")
		return nil, nil, nil, nil, nil, nil, err
	}
	mineLog.WithField("tmpEventsFilePath",
		tmpEventsFilePath).Info("Unique EventNames", eventNames)

	mineLog.WithField("tmpEventsFilePath",
		tmpEventsFilePath).Info("Building user and event properties info and writing it to file.")
	eventNamesWithType, err := store.GetStore().GetEventTypeFromDb(projectId, eventNames, 100000)

	mineLog.WithField("Event Names and type",
		eventNamesWithType).Info("Building user and type from DB.")

	if err != nil {
		mineLog.WithFields(log.Fields{"err": err}).Error("Failed to get event names and Type from DB")
		return nil, nil, nil, nil, nil, nil, err
	}

	return eventNames, eventNamesWithType, upCount, epCount, upCatg, epCatg, nil
}

// buildWhiteListProperties build user and event properties from db ,
// if not available in DB, use counting logic to choose topk properties
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

	userPropertiesList, errInt := store.GetStore().GetAllActiveFactorsTrackedUserPropertiesByProject(projectId)
	if errInt != http.StatusFound {
		mineLog.WithFields(log.Fields{"err": errInt}).Error("Unable to fetch UserProperties from db")
		return nil, nil
	}

	if len(userPropertiesList) > 0 {
		mineLog.WithFields(log.Fields{"user properties": userPropertiesList}).Info("Number of User properties from db :", len(userPropertiesList))

		for _, v := range userPropertiesList {
			if _, ok := userPropertiesMap[v.UserPropertyName]; ok {
				upFilteredMap[v.UserPropertyName] = true
			} else {
				mineLog.Info("Missing user property in events File ", v.UserPropertyName)
			}
		}
	} else {

		// if the DB is not populated , based on counting logic
		// populate the DB and use the user properties
		upSortedList := U.RankByWordCount(userPropertiesMap)
		var userPropertiesCount = 0
		for _, u := range upSortedList {
			if !upFilteredMap[u.Key] {
				upFilteredMap[u.Key] = true
				userPropertiesCount++
			}
		}

		// delete keys based on disabled_Properties
		for _, Uprop := range U.DISABLED_FACTORS_USER_PROPERTIES {
			if upFilteredMap[Uprop] {
				delete(upFilteredMap, Uprop)

			}
		}
		for key := range upFilteredMap {
			mineLog.Info("insert user property", key)
			_, errInt = store.GetStore().CreateFactorsTrackedUserProperty(projectId, key, "")
			if errInt != http.StatusCreated {
				errorString := fmt.Sprintf("unable to insert user property to db %s", key)
				mineLog.WithFields(log.Fields{"http status": errInt}).Error(errorString)

			}
		}
	}
	// ep : event Properties : addkeys based on ranking , add based on whitelist properties
	// delete based on disables properties

	epSortedList := U.RankByWordCount(eventPropertiesMap)
	var eventCountLocal = 0
	for _, u := range epSortedList {
		if !epFilteredMap[u.Key] {
			epFilteredMap[u.Key] = true
			eventCountLocal++
		}
	}

	for _, Eprop := range U.DISABLED_FACTORS_EVENT_PROPERTIES {
		delete(epFilteredMap, Eprop)
	}
	return upFilteredMap, epFilteredMap
}

func buildEventsFileOnProperties(tmpEventsFilePath string, efTmpPath string, efTmpName string, diskManager *serviceDisk.DiskDriver, projectId uint64,
	modelId uint64, eReader io.Reader, userPropList, eventPropList map[string]bool, campaignLimitCount int,
	upCount, epCount map[string]map[string]int, upCatg, epCatg map[string]string) (CampaignEventLists, error) {
	// Rewrites the events file restricting the properties to only whitelisted properties.
	//  Also returns list special campaign events created during rewrite.

	var err error
	efPath := efTmpPath + "tmpevents_" + efTmpName
	smartEvents, err := rewriteEventsFile(tmpEventsFilePath, efPath, userPropList, eventPropList, upCount, epCount, upCatg, epCatg)
	if err != nil {
		mineLog.WithFields(log.Fields{"err": err, "tmpEventsFilePath": tmpEventsFilePath}).Error("Failed to filter disabled properties")
		return CampaignEventLists{}, err
	}
	r, _ := os.Open(efPath)
	err = diskManager.Create(efTmpPath, efTmpName, r)
	if err != nil {
		mineLog.WithFields(log.Fields{"err": err, "eventFilePath": efTmpPath,
			"eventFileName": efTmpName}).Error("Failed to create event file on disk.")
		return CampaignEventLists{}, err
	}
	mineLog.WithFields(log.Fields{"eventFilePath": efPath}).Info("Removing temp events File")
	err = os.Remove(efPath)
	if err != nil {
		mineLog.WithFields(log.Fields{"err": err, "eventFilePath": efPath}).Error("Failed to remove file")
		return CampaignEventLists{}, err

	}

	return smartEvents, nil

}

func buildEventsInfoForEncodedEvents(smartEvents CampaignEventLists, userAndEventsInfo *P.UserAndEventsInfo) *P.UserAndEventsInfo {
	// add events info for campEventsList . This create empty numeric and categorical Key values for each event in encodedList
	eMap := *userAndEventsInfo.EventPropertiesInfoMap

	listCampaignEventsNames := [][]string{smartEvents.CampaignList, smartEvents.MediumList, smartEvents.SourceList, smartEvents.ReferrerList, smartEvents.AdgroupList}

	for _, ct := range listCampaignEventsNames {
		if len(ct) > 0 {
			mineLog.Info("Intializing campaign events : ", ct[0])
			for _, eventName := range ct {
				// Initialize info.

				eMap[eventName] = &P.PropertiesInfo{
					NumericPropertyKeys:          make(map[string]bool),
					CategoricalPropertyKeyValues: make(map[string]map[string]bool),
				}
			}
		}
	}

	return userAndEventsInfo
}

// PatternMine Mine TOP_K Frequent patterns for every event combination (segment) at every iteration.
func PatternMine(db *gorm.DB, etcdClient *serviceEtcd.EtcdClient, cloudManager *filestore.FileManager,
	diskManager *serviceDisk.DiskDriver, bucketName string, numRoutines int, projectId uint64,
	modelId uint64, modelType string, startTime int64, endTime int64, maxModelSize int64,
	countOccurence bool, campaignLimitCount int, beamConfig *RunBeamConfig, countsVersion int) (int, error) {

	var err error

	// download events file from cloud to local
	efCloudPath, efCloudName := (*cloudManager).GetModelEventsFilePathAndName(projectId, startTime, modelType)
	efTmpPath, efTmpName := diskManager.GetModelEventsFilePathAndName(projectId, startTime, modelType)
	mineLog.WithFields(log.Fields{"eventFileCloudPath": efCloudPath,
		"eventFileCloudName": efCloudName}).Info("Downloading events file from cloud.")
	eReader, err := (*cloudManager).Get(efCloudPath, efCloudName)
	if err != nil {
		mineLog.WithFields(log.Fields{"err": err, "eventFilePath": efCloudPath,
			"eventFileName": efCloudName}).Error("Failed downloading events file from cloud.")
		return 0, err
	}
	err = diskManager.Create(efTmpPath, efTmpName, eReader)
	if err != nil {
		mineLog.WithFields(log.Fields{"err": err, "eventFilePath": efCloudPath,
			"eventFileName": efCloudName}).Error("Failed downloading events file from cloud.")
		return 0, err
	}
	tmpEventsFilepath := efTmpPath + efTmpName
	mineLog.Info("Successfuly downloaded events file from cloud.", tmpEventsFilepath, efTmpPath, efTmpName)
	eventNames, eventNamesWithType, upCount, epCount, upCatg, epCatg, err := GetEventNamesAndType(tmpEventsFilepath, projectId, countsVersion)
	if err != nil {
		mineLog.WithFields(log.Fields{"err": err}).Error("Failed to get eventName and event type.")
		return 0, err
	}

	userAndEventsInfo, allPropsMap, err := buildPropertiesInfoFromInput(projectId, eventNames, tmpEventsFilepath)
	if err != nil {
		mineLog.WithFields(log.Fields{"err": err}).Error("Failed to build user and event Info.")
		return 0, err
	}
	userPropList, eventPropList := buildWhiteListProperties(projectId, allPropsMap, topKProperties)
	userAndEventsInfo = FilterEventsInfo(userAndEventsInfo, userPropList, eventPropList)

	userAndEventsInfoBytes, err := json.Marshal(userAndEventsInfo)
	if err != nil {
		mineLog.WithFields(log.Fields{"err": err}).Error("Failed to unmarshal events Info.")
		return 0, err
	}

	if len(userAndEventsInfoBytes) > 249900000 {
		// Limit is 250MB
		errorString := fmt.Sprintf(
			"Too big properties info, modelId: %d, modelType: %s, projectId: %d, numBytes: %d",
			modelId, modelType, projectId, len(userAndEventsInfoBytes))
		mineLog.Error(errorString)
		return 0, fmt.Errorf(errorString)
	}
	err = writeEventInfoFile(projectId, modelId, bytes.NewReader(userAndEventsInfoBytes), (*cloudManager))
	if err != nil {
		mineLog.WithFields(log.Fields{"err": err}).Error("Failed to write events Info.")
		return 0, err
	}

	mineLog.Info("Number of EventNames: ", len(eventNames))
	mineLog.Info("Number of User Properties: ", len(userPropList))
	mineLog.Info("Number of Event Properties: ", len(eventPropList))
	campaignAnalyticsSorted, err := buildEventsFileOnProperties(tmpEventsFilepath, efTmpPath, efTmpName, diskManager, projectId,
		modelId, eReader, userPropList, eventPropList, campaignLimitCount, upCount, epCount, upCatg, epCatg)
	if err != nil {
		mineLog.WithFields(log.Fields{"err": err}).Error("Failed to write events data.")
		return 0, err
	}

	userAndEventsInfo = buildEventsInfoForEncodedEvents(campaignAnalyticsSorted, userAndEventsInfo)
	mineLog.Info("Successfully Built events data and written to file.")
	mineLog.Info("Number of Campaign events:", len(campaignAnalyticsSorted.CampaignList))
	mineLog.Info("Number of Medium events:", len(campaignAnalyticsSorted.MediumList))
	mineLog.Info("Number of Referrer events:", len(campaignAnalyticsSorted.ReferrerList))
	mineLog.Info("Number of source events:", len(campaignAnalyticsSorted.SourceList))
	mineLog.Info("Number of Adgroup events:", len(campaignAnalyticsSorted.AdgroupList))

	repeatedEvents := GetAllRepeatedEvents(eventNames, campaignAnalyticsSorted)
	mineLog.Info("Number of repeated Events: ", len(repeatedEvents))
	mineLog.Info("Repeated Events", repeatedEvents)

	// build histogram of all user properties.
	mineLog.WithField("tmpEventsFilePath", tmpEventsFilepath).Info("Building all user properties histogram.")
	allActiveUsersPattern, err := P.NewPattern([]string{U.SEN_ALL_ACTIVE_USERS}, userAndEventsInfo)
	if err != nil {
		mineLog.WithFields(log.Fields{"err": err}).Error("Failed to build pattern with histogram of all active user properties.")
		return 0, err
	}
	if err := computeAllUserPropertiesHistogram(projectId, tmpEventsFilepath, allActiveUsersPattern, countsVersion); err != nil {
		mineLog.WithFields(log.Fields{"err": err}).Error("Failed to compute user properties.")
		return 0, err
	}
	tmpChunksDir := diskManager.GetPatternChunksDir(projectId, modelId)
	if err := serviceDisk.MkdirAll(tmpChunksDir); err != nil {
		mineLog.WithFields(log.Fields{"chunkDir": tmpChunksDir, "error": err}).Error("Unable to create chunks directory.")
		return 0, err
	}
	if err := writePatternsAsChunks([]*P.Pattern{allActiveUsersPattern}, tmpChunksDir); err != nil {
		mineLog.WithFields(log.Fields{"err": err}).Error("Failed to write user properties.")
		return 0, err
	}
	mineLog.Info("Successfully built all user properties histogram.")

	//write events file to GCP
	modEventsFile := fmt.Sprintf("events_modified_%d.txt", modelId)
	err = writeFileToGCP(projectId, modelId, modEventsFile, tmpEventsFilepath, cloudManager)
	if err != nil {
		return 0, fmt.Errorf("unable to write modified events file to GCP")
	}

	// mine and write patterns as chunks
	mineLog.WithFields(log.Fields{"projectId": projectId, "tmpEventsFilepath": tmpEventsFilepath,
		"tmpChunksDir": tmpChunksDir, "routines": numRoutines}).Info("Mining patterns and writing it as chunks.")
	err = mineAndWritePatterns(projectId, modelId, tmpEventsFilepath,
		userAndEventsInfo, cloudManager, eventNames, numRoutines, tmpChunksDir, maxModelSize,
		countOccurence, eventNamesWithType, repeatedEvents, campaignAnalyticsSorted,
		efTmpPath, beamConfig, countsVersion)
	if err != nil {
		mineLog.WithFields(log.Fields{"err": err}).Error("Failed to mine patterns.")
		return 0, err
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

	chunkIdsString := ""
	for _, chunkId := range chunkIds {
		if chunkIdsString != "" {
			chunkIdsString += ","
		}
		chunkIdsString += chunkId
	}
	errCode, message := store.GetStore().CreateProjectModelMetadata(&model.ProjectModelMetadata{
		ProjectId: projectId,
		ModelId:   modelId,
		ModelType: modelType,
		StartTime: startTime,
		EndTime:   endTime,
		// CONVERT THIS TO COMMA SEPERATED STRING
		Chunks: chunkIdsString,
	})
	if errCode != http.StatusCreated {
		mineLog.Error(message)
	}
	newVersionId := fmt.Sprintf("%v", time.Now().Unix())
	err = etcdClient.SetProjectVersion(newVersionId)
	if err != nil {
		mineLog.WithFields(log.Fields{"err": err}).Error("Failed to write new version id to etcd.")
		return 0, err
	}
	return len(chunkIds), nil
}

func convert(eventNamesWithAggregation []model.EventNameWithAggregation) []model.EventName {
	eventNames := make([]model.EventName, 0)
	for _, event := range eventNamesWithAggregation {
		eventNames = append(eventNames, model.EventName{
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

// OpenEventFileAndGetScanner open file to read events
func OpenEventFileAndGetScanner(filePath string) (*bufio.Scanner, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	scanner := patternStore.CreateScannerFromReader(f)
	return scanner, nil
}

func FilterTopKEventsOnTypes(filteredPatterns []*P.Pattern, eventNamesWithType map[string]string, k, keventsSpecial, keventsURL int, campaignEventType CampaignEventLists) []*P.Pattern {

	// take topK from different event types like uc,fe,$types,url etc..
	allPatterns := make([]patternProperties, 0)

	for _, pattern_ := range filteredPatterns {
		var tmpPattern patternProperties

		tmpPattern.pattern = pattern_
		tmpPattern.count = pattern_.PerUserCount
		tmpPattern.patternType = eventNamesWithType[pattern_.EventNames[0]]

		allPatterns = append(allPatterns, tmpPattern)

	}

	//convert struct array allPatterns([]patternProperties) to interface array allPattern([]U.PatternProperties) to pass in takeTopK functions
	allPattern := make([]U.PatternProperties, len(allPatterns))
	for i, pat := range allPatterns {
		var upat U.PatternProperties = pat
		allPattern[i] = upat
	}

	//take top K for each property type
	ucTopk := U.TakeTopKUC(allPattern, k, model.TYPE_USER_CREATED_EVENT_NAME)
	feAT_Topk := U.TakeTopKpageView(allPattern, k, model.TYPE_FILTER_EVENT_NAME, model.TYPE_AUTO_TRACKED_EVENT_NAME)
	ieTopk := U.TakeTopKIE(allPattern, k, model.TYPE_INTERNAL_EVENT_NAME)
	specialTopk := U.TakeTopKspecialEvents(allPattern, keventsSpecial)
	URLTopk := U.TakeTopKAllURL(allPattern, keventsURL)

	//convert each interface array([]U.PatternProperties) back to struct array([]patternProperties)
	ucTopK := make([]patternProperties, len(ucTopk))
	for i, pat := range ucTopk {
		ucTopK[i] = pat.(patternProperties)
	}
	feAT_TopK := make([]patternProperties, len(feAT_Topk))
	for i, pat := range feAT_Topk {
		feAT_TopK[i] = pat.(patternProperties)
	}
	ieTopK := make([]patternProperties, len(ieTopk))
	for i, pat := range ieTopk {
		ieTopK[i] = pat.(patternProperties)
	}
	specialTopK := make([]patternProperties, len(specialTopk))
	for i, pat := range specialTopk {
		specialTopK[i] = pat.(patternProperties)
	}
	URLTopK := make([]patternProperties, len(URLTopk))
	for i, pat := range URLTopk {
		URLTopK[i] = pat.(patternProperties)
	}

	allPatternsFiltered := make([]patternProperties, 0)
	allPatternsFiltered = append(allPatternsFiltered, ucTopK...)
	allPatternsFiltered = append(allPatternsFiltered, feAT_TopK...)
	allPatternsFiltered = append(allPatternsFiltered, ieTopK...)
	allPatternsFiltered = append(allPatternsFiltered, specialTopK...)
	allPatternsFiltered = append(allPatternsFiltered, URLTopK...)

	allPatternsTopk := make([]*P.Pattern, 0)
	exists := make(map[string]bool)

	for _, pt := range allPatternsFiltered {
		if !exists[pt.pattern.EventNames[0]] {
			allPatternsTopk = append(allPatternsTopk, pt.pattern)
			exists[pt.pattern.EventNames[0]] = true
		}
	}

	return allPatternsTopk

}

func takeCampaignEvents(allPatterns []patternProperties, campaignEventsType CampaignEventLists) []patternProperties {

	allPatternsType := make([]patternProperties, 0)

	exists := make(map[string]bool)

	for _, v := range campaignEventsType.CampaignList {
		exists[v] = true
	}
	for _, v := range campaignEventsType.MediumList {
		exists[v] = true
	}
	for _, v := range campaignEventsType.ReferrerList {
		exists[v] = true
	}
	for _, v := range campaignEventsType.SourceList {
		exists[v] = true
	}

	for _, pt := range allPatterns {

		if exists[pt.pattern.EventNames[0]] {
			allPatternsType = append(allPatternsType, pt)
		}
	}
	return allPatternsType

}

// GetAllCyclicEvents Filter all special events
func GetAllRepeatedEvents(eventNames []string, campaignAnalyticsList CampaignEventLists) []string {
	// all
	CyclicEvents := make([]string, 0)
	eventSet := make(map[string]bool)

	for _, v := range eventNames {

		if strings.HasPrefix(v, "$") {
			if !eventSet[v] {
				CyclicEvents = append(CyclicEvents, v)
				eventSet[v] = true
			}
		}
	}

	for _, ce := range campaignAnalyticsList.CampaignList {
		if !eventSet[ce] {
			CyclicEvents = append(CyclicEvents, ce)
			eventSet[ce] = true
		}
	}

	for _, ce := range campaignAnalyticsList.MediumList {
		if !eventSet[ce] {
			CyclicEvents = append(CyclicEvents, ce)
			eventSet[ce] = true
		}
	}

	for _, ce := range campaignAnalyticsList.ReferrerList {
		if !eventSet[ce] {
			CyclicEvents = append(CyclicEvents, ce)
			eventSet[ce] = true
		}
	}
	for _, ce := range campaignAnalyticsList.SourceList {
		if !eventSet[ce] {
			CyclicEvents = append(CyclicEvents, ce)
			eventSet[ce] = true
		}
	}
	return CyclicEvents

}

func getAllRepeatedEventPatterns(allPatterns []*P.Pattern, standardEventsList []string) []*P.Pattern {
	standardEventsMap := make(map[string]bool, 0)
	filteredPatterns := make([]*P.Pattern, 0)

	for _, v := range standardEventsList {
		standardEventsMap[v] = true
	}

	for _, v := range allPatterns {
		if standardEventsMap[v.EventNames[0]] {
			filteredPatterns = append(filteredPatterns, v)

		}
	}

	return filteredPatterns

}

func MergePatterns(patternA, patternB []*P.Pattern) []*P.Pattern {
	// take union of two patterns
	allPatterns := make([]*P.Pattern, 0)
	allPatternsMap := make(map[string]bool)

	for _, patt := range patternA {
		eventNameString := strings.Join(patt.EventNames, "_")
		if _, ok := allPatternsMap[eventNameString]; !ok {
			allPatternsMap[eventNameString] = true
			allPatterns = append(allPatterns, patt)
		}
	}

	for _, patt := range patternB {
		eventNameString := strings.Join(patt.EventNames, "_")
		if _, ok := allPatternsMap[eventNameString]; !ok {
			allPatternsMap[eventNameString] = true
			allPatterns = append(allPatterns, patt)
		}
	}

	return allPatterns
}

func getTopPatterns(patterns []*P.Pattern, topKPatterns int) []*P.Pattern {

	if topKPatterns == 0 {
		return []*P.Pattern{}
	}

	if len(patterns) > 0 && topKPatterns > 1 {
		sort.Slice(patterns, func(i, j int) bool { return patterns[i].PerUserCount > patterns[j].PerUserCount })
		if len(patterns) > topKPatterns {
			return patterns[0:topKPatterns]
		}
		return patterns
	}
	return patterns

}

func GetTopURLs(allPatterns []*P.Pattern, maxNum int) []*P.Pattern {

	allUrls := make([]*P.Pattern, 0)

	for _, v := range allPatterns {

		if U.IsValidUrl(v.EventNames[0]) {
			allUrls = append(allUrls, v)
		}
	}

	filteredURLs := getTopPatterns(allUrls, maxNum)
	return filteredURLs

}

func FilterEventsInfo(userAndEventsInfo *P.UserAndEventsInfo, userProp, eventProp map[string]bool) *P.UserAndEventsInfo {

	userPropertiesInfo := userAndEventsInfo.UserPropertiesInfo
	eventPropertiesInfo := *userAndEventsInfo.EventPropertiesInfoMap

	//delete both categorical and numerical properties for users
	for propertyName := range userPropertiesInfo.CategoricalPropertyKeyValues {

		if !userProp[propertyName] {
			delete(userPropertiesInfo.CategoricalPropertyKeyValues, propertyName)
		}
	}

	for propertyName := range userPropertiesInfo.NumericPropertyKeys {

		if !userProp[propertyName] {
			delete(userPropertiesInfo.NumericPropertyKeys, propertyName)
		}
	}

	//delete both categorical and numerical properties for events
	for _, property := range eventPropertiesInfo {

		for k := range property.CategoricalPropertyKeyValues {

			if !eventProp[k] {
				delete(property.CategoricalPropertyKeyValues, k)
			}
		}

		for k := range property.NumericPropertyKeys {

			if !eventProp[k] {
				delete(property.NumericPropertyKeys, k)
			}
		}
	}

	return userAndEventsInfo
}

func GetTopUDE(allPatterns []*P.Pattern, eventNamesWithType map[string]string, maxNum int) []*P.Pattern {

	allUDE := make([]*P.Pattern, 0)

	for _, v := range allPatterns {

		if eventNamesWithType[v.EventNames[0]] == model.TYPE_USER_CREATED_EVENT_NAME {
			allUDE = append(allUDE, v)
		}
	}

	filteredUDE := getTopPatterns(allUDE, maxNum)
	return filteredUDE

}

func GetTopAT(allPatterns []*P.Pattern, eventNamesWithType map[string]string, maxNum int) []*P.Pattern {
	// get urls
	allAT := make([]*P.Pattern, 0)

	for _, v := range allPatterns {

		if eventNamesWithType[v.EventNames[0]] == model.TYPE_AUTO_TRACKED_EVENT_NAME {
			allAT = append(allAT, v)
		}
	}

	filteredAT := getTopPatterns(allAT, maxNum)
	return filteredAT

}

func GetTopSmartEvents(allPatterns []*P.Pattern, eventNamesWithType map[string]string, maxNum int) []*P.Pattern {

	allUDE := make([]*P.Pattern, 0)

	for _, v := range allPatterns {

		if eventNamesWithType[v.EventNames[0]] == model.TYPE_CRM_HUBSPOT || eventNamesWithType[v.EventNames[0]] == model.TYPE_CRM_SALESFORCE {
			allUDE = append(allUDE, v)
		}
	}

	filteredUDE := getTopPatterns(allUDE, maxNum)
	return filteredUDE

}

//GetTopStandardPatterns Get all events which starts with $ and not campaign Analytics Event
func GetTopStandardPatterns(allPatterns []*P.Pattern, maxNum int) []*P.Pattern {

	allStandardPatterns := make([]*P.Pattern, 0)

	for _, v := range allPatterns {

		if strings.HasPrefix(v.EventNames[0], "$") && !U.IsCampaignAnalytics(v.EventNames[0]) {
			allStandardPatterns = append(allStandardPatterns, v)
		}
	}

	filteredCampaignPatterns := getTopPatterns(allStandardPatterns, maxNum)
	return filteredCampaignPatterns
	// return allStandardPatterns

}

func GetTopCampaigns(allPatterns []*P.Pattern, maxNum int) []*P.Pattern {
	allCampaignPatterns := make([]*P.Pattern, 0)
	for _, v := range allPatterns {
		if U.IsCampaignEvent(v.EventNames[0]) {
			allCampaignPatterns = append(allCampaignPatterns, v)
		}
	}
	filteredCampaignPatterns := getTopPatterns(allCampaignPatterns, maxNum)
	return filteredCampaignPatterns
}

//GetTopSourcePatterns get all top patterns with $session and source set
func GetTopSourcePatterns(allPatterns []*P.Pattern, maxNum int) []*P.Pattern {
	allCampaignPatterns := make([]*P.Pattern, 0)
	for _, v := range allPatterns {
		if U.IsSourceEvent(v.EventNames[0]) {
			allCampaignPatterns = append(allCampaignPatterns, v)
		}
	}
	filteredCampaignPatterns := getTopPatterns(allCampaignPatterns, maxNum)
	return filteredCampaignPatterns
}

func GetTopMediumPatterns(allPatterns []*P.Pattern, maxNum int) []*P.Pattern {
	allCampaignPatterns := make([]*P.Pattern, 0)
	for _, v := range allPatterns {
		if U.IsMediumEvent(v.EventNames[0]) {
			allCampaignPatterns = append(allCampaignPatterns, v)
		}
	}
	filteredCampaignPatterns := getTopPatterns(allCampaignPatterns, maxNum)
	return filteredCampaignPatterns
}

func GetTopAdgroupPatterns(allPatterns []*P.Pattern, maxNum int) []*P.Pattern {
	allCampaignPatterns := make([]*P.Pattern, 0)
	for _, v := range allPatterns {
		if U.IsAdgroupEvent(v.EventNames[0]) {
			allCampaignPatterns = append(allCampaignPatterns, v)
		}
	}
	filteredCampaignPatterns := getTopPatterns(allCampaignPatterns, maxNum)
	return filteredCampaignPatterns
}

func GetTopReferrerPatterns(allPatterns []*P.Pattern, maxNum int) []*P.Pattern {
	allCampaignPatterns := make([]*P.Pattern, 0)
	for _, v := range allPatterns {
		if U.IsReferrerEvent(v.EventNames[0]) {
			allCampaignPatterns = append(allCampaignPatterns, v)
		}
	}
	filteredCampaignPatterns := getTopPatterns(allCampaignPatterns, maxNum)
	return filteredCampaignPatterns
}

func GetTopCampaignAnalyticsPatterns(allPatterns []*P.Pattern, cNum, refNum, medNum, srcNum, adgNum int) []*P.Pattern {
	causalCampaignsEvents := GetTopCampaigns(allPatterns, cNum)
	for _, v := range causalCampaignsEvents {
		mineLog.Info("Top Campaign Events ->", v.EventNames, " ", v.PerUserCount)
	}

	causalSourceEvents := GetTopSourcePatterns(allPatterns, srcNum)
	for _, v := range causalSourceEvents {
		mineLog.Info("Top Source Events ->", v.EventNames, " ", v.PerUserCount)
	}

	causalMediumEvents := GetTopMediumPatterns(allPatterns, medNum)
	for _, v := range causalMediumEvents {
		mineLog.Info("Top Medium Events ->", v.EventNames, " ", v.PerUserCount)
	}

	causalReferrerEvents := GetTopReferrerPatterns(allPatterns, medNum)
	for _, v := range causalReferrerEvents {
		mineLog.Info("Top Referrer Events ->", v.EventNames, " ", v.PerUserCount)
	}

	causalAdgroupEvents := GetTopAdgroupPatterns(allPatterns, adgNum)
	for _, v := range causalAdgroupEvents {
		mineLog.Info("Top Adgroup Events ->", v.EventNames, " ", v.PerUserCount)
	}

	filteredPatterns := MergePatterns(causalSourceEvents, causalCampaignsEvents)
	filteredPatterns = MergePatterns(filteredPatterns, causalMediumEvents)
	filteredPatterns = MergePatterns(filteredPatterns, causalAdgroupEvents)
	filteredPatterns = MergePatterns(filteredPatterns, causalReferrerEvents)
	return filteredPatterns

}

func FilterAllCausualEvents(allPatterns []*P.Pattern, eventNamesWithType map[string]string) ([]*P.Pattern, error) {

	causalURLs := GetTopURLs(allPatterns, countURL)
	causalUDE := GetTopUDE(allPatterns, eventNamesWithType, countUDE)
	causalAT := GetTopAT(allPatterns, eventNamesWithType, countURL)
	causalSME := GetTopSmartEvents(allPatterns, eventNamesWithType, countSME)
	causalStandardEvents := GetTopStandardPatterns(allPatterns, countStdEvents)
	campaignAnalyticsEvents := GetTopCampaignAnalyticsPatterns(allPatterns, countCampaigns, countReferrer, countMedium, countSource, countAdgroup)

	filteredPatterns := MergePatterns(causalURLs, causalUDE)
	filteredPatterns = MergePatterns(filteredPatterns, causalSME)
	filteredPatterns = MergePatterns(filteredPatterns, causalAT)
	filteredPatterns = MergePatterns(filteredPatterns, causalStandardEvents)
	filteredPatterns = MergePatterns(filteredPatterns, campaignAnalyticsEvents)

	mineLog.Infof("Total causal Patterns :%d", len(filteredPatterns))
	mineLog.Infof("All causal Patterns :%v", filteredPatterns)

	return filteredPatterns, nil
}

//GenInterMediateCombinations generate cross combination patterns based on len two patt
func GenInterMediateCombinations(lenTwoPatt []*P.Pattern, userAndEventsInfo *P.UserAndEventsInfo) ([]*P.Pattern, error) {
	// input-> {{"a","g"},{"b","g"}}
	// result - > {"a","b","g"} {"b","a","g"}

	for _, v := range lenTwoPatt {
		if len(v.EventNames) != 2 {
			return nil, fmt.Errorf("patterns are not of size 2")
		}
	}
	allPatterns := make([]*P.Pattern, 0)
	for i := 0; i < len(lenTwoPatt); i++ {
		for j := i + 1; j < len(lenTwoPatt); j++ {
			if !reflect.DeepEqual(lenTwoPatt[i].EventNames, lenTwoPatt[j].EventNames) {
				tmpPatt, err := P.GenCandidatesForGoals(lenTwoPatt[i], lenTwoPatt[j], userAndEventsInfo)
				if err != nil {
					return nil, fmt.Errorf("unable to create intermediate combinations")
				}
				allPatterns = append(allPatterns, tmpPatt...)
			}
		}

	}

	return allPatterns, nil

}

// GenRepeatedCombinations : from len two events , create repeated combination events
func GenRepeatedCombinations(lenTwoPatt []*P.Pattern, userAndEventsInfo *P.UserAndEventsInfo, repeatedEvents []string) ([]*P.Pattern, error) {

	// Example : lenTwo -> [[e1,g1],[e2,g1],[e3,g2],[e4,g4]]
	// 		  repeated -> [e2,e3]
	// 		  res -> [[e2,e2,g1],[e3,e3,g1]]

	for _, v := range lenTwoPatt {
		if len(v.EventNames) != 2 {
			return nil, fmt.Errorf("length of pattern is not equal to 2")
		}
	}

	repeatedEventsMap := make(map[string]bool)
	for _, v := range repeatedEvents {
		repeatedEventsMap[v] = true
	}

	allStrings := make([][]string, 0)
	var allPatterns []*P.Pattern

	for _, v := range lenTwoPatt {
		if repeatedEventsMap[v.EventNames[0]] {
			allStrings = append(allStrings, []string{v.EventNames[0], v.EventNames[0], v.EventNames[1]})
		}
	}

	//create pattern for repeated event
	for _, tmpPattEventNames := range allStrings {
		tmpPatt, err := P.NewPattern(tmpPattEventNames, userAndEventsInfo)
		if err != nil {
			return []*P.Pattern{}, fmt.Errorf("unable to generate new n+1 len Pattern")
		}
		allPatterns = append(allPatterns, tmpPatt)

	}
	return allPatterns, nil
}

// GenCampaignThreeLenCombinations Generate three Len events with only campaign events and goal
func GenCampaignThreeLenCombinations(lenTwoPatt, goalPatterns []*P.Pattern, userAndEventsInfo *P.UserAndEventsInfo, numTopK int) ([]*P.Pattern, error) {

	patternsMap := make(map[string][]*P.Pattern)
	allPatternsList := make([]*P.Pattern, 0)

	for _, pt := range lenTwoPatt {
		if len(pt.EventNames) == 2 {
			startEventName := pt.EventNames[0]
			goalEventName := pt.EventNames[1]
			if U.IsCampaignEvent(startEventName) {
				patternsMap[goalEventName] = append(patternsMap[goalEventName], pt)
			}
		}
	}
	// for each goal get top k campaigns and create comabinations within topK and goal
	for key, val := range patternsMap {
		goalName := key
		targetPatt := getTopPatterns(val, numTopK)
		mineLog.Info("Three len Campaign Events goal Event : ", goalName, len(targetPatt))

		if len(targetPatt) > 1 {
			for idx_a := 0; idx_a < len(targetPatt); idx_a++ {
				for idx_b := 0; idx_b < len(targetPatt); idx_b++ {

					event_a := targetPatt[idx_a].EventNames[0]
					event_b := targetPatt[idx_b].EventNames[0]
					event_c := goalName

					if strings.Compare(event_a, event_b) != 0 {
						p, err := P.NewPattern([]string{event_a, event_b, event_c}, userAndEventsInfo)
						if err != nil {
							return []*P.Pattern{}, fmt.Errorf("campaign Pattern initialization failed")
						}

						allPatternsList = append(allPatternsList, p)
						mineLog.Info("Three len Campaign Events : ", p.EventNames)
					}
				}
			}
		}
	}

	return allPatternsList, nil
}

// GetEventNamesFromFile read unique eventNames from Event file
func GetEventNamesFromFile_(scanner *bufio.Scanner, projectId uint64) ([]string, error) {
	logCtx := log.WithField("project_id", projectId)
	scanner.Split(bufio.ScanLines)
	var txtline string
	eventNames := make([]string, 0)
	var dat map[string]interface{}
	s := map[string]bool{}

	for scanner.Scan() {
		txtline = scanner.Text()
		if err := json.Unmarshal([]byte(txtline), &dat); err != nil {
			logCtx.Error("Unable to decode line")
		}
		eventNameString := dat["en"].(string)
		_, ok := s[eventNameString]
		if ok != true {
			eventNames = append(eventNames, eventNameString)
			s[eventNameString] = true
		}

	}
	err := scanner.Err()
	logCtx.Info("Extraced Unique EventNames from file")

	if err != nil {
		return []string{}, err
	}

	return eventNames, nil
}

// GetEventNamesFromFile read unique eventNames from Event file
func GetEventNamesFromFile(scanner *bufio.Scanner, projectId uint64, countsVersion int) ([]string, map[string]map[string]int, map[string]map[string]int, map[string]string, map[string]string, error) {
	logCtx := log.WithField("project_id", projectId)
	scanner.Split(bufio.ScanLines)
	var txtline string
	eventNames := make([]string, 0)
	userPropCatgMap := make(map[string]string)
	eventPropCatgMap := make(map[string]string)

	eventUserPropMap := make(map[string]map[string]int)
	eventEventPropMap := make(map[string]map[string]int)

	blacklist := []string{}

	s := map[string]bool{}

	for scanner.Scan() {
		txtline = scanner.Text()

		var eventDetails P.CounterEventFormat
		if err := json.Unmarshal([]byte(txtline), &eventDetails); err != nil {
			log.WithFields(log.Fields{"line": txtline, "err": err}).Error("Read failed")
			return nil, nil, nil, nil, nil, err
		}
		eventNameString := eventDetails.EventName

		if _, ok := eventEventPropMap[eventNameString]; !ok {
			eventEventPropMap[eventNameString] = make(map[string]int)
		}
		if _, ok := eventUserPropMap[eventNameString]; !ok {
			eventUserPropMap[eventNameString] = make(map[string]int)
		}

		_, ok := s[eventNameString]
		if ok != true {
			eventNames = append(eventNames, eventNameString)
			s[eventNameString] = true
		}

		for uKey, uVal := range eventDetails.UserProperties {
			if _, ok := eventPropCatgMap[uKey]; !ok {
				propertyType := store.GetStore().GetPropertyTypeByKeyValue(projectId, eventNameString, uKey, uVal, true)
				userPropCatgMap[uKey] = propertyType
			}

			propType := userPropCatgMap[uKey]
			if propType == U.PropertyTypeCategorical {
				pstring := U.GetPropertyValueAsString(uVal)
				propString := strings.Join([]string{uKey, pstring}, "::")
				if ok, _, _ := U.StringIn(blacklist, uKey); !ok {
					eventUserPropMap[eventNameString][propString] += 1
				}
			}

		}

		for eKey, eVal := range eventDetails.EventProperties {
			if _, ok := eventPropCatgMap[eKey]; !ok {
				propertyType := store.GetStore().GetPropertyTypeByKeyValue(projectId, eventNameString, eKey, eVal, false)
				eventPropCatgMap[eKey] = propertyType
			}

			propType := eventPropCatgMap[eKey]
			if propType == U.PropertyTypeCategorical {
				pstring := U.GetPropertyValueAsString(eVal)
				if pstring == "" {
					pstring = "$none"
				}
				propString := strings.Join([]string{eKey, pstring}, "::")
				if ok, _, _ := U.StringIn(blacklist, eKey); !ok {
					eventEventPropMap[eventNameString][propString] += 1
				}
			}

		}
	}
	err := scanner.Err()
	logCtx.Info("Extraced Unique EventNames from file")

	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	if countsVersion == 2 {
		mineLog.Info("Filtering based on topK")
		// filter based on min Support
		topk := 15
		userKeys := 0
		eventKeys := 0
		for _, v := range eventUserPropMap {
			U.FilterOnFrequency(v, topk)
			userKeys += len(v)
		}
		for _, v := range eventEventPropMap {
			U.FilterOnFrequency(v, topk)
			eventKeys += len(v)
		}

		mineLog.Infof("Total number of user properties key,values after filtering:%d", userKeys)
		mineLog.Infof("Total number of event properties key,values after filtering:%d", eventKeys)
	}
	return eventNames, eventUserPropMap, eventEventPropMap, userPropCatgMap, eventPropCatgMap, nil
}

func removePropertiesTree(val string) bool {

	if U.IsDateTime(val) {
		return true
	}

	return false
}
