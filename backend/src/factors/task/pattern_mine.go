package task

import (
	"bufio"
	"bytes"
	"encoding/json"
	"factors/filestore"
	"factors/merge"
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
	"path"
	"path/filepath"
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

// const prop_thresh_percent = 0.5
const max_prop_values_per_key = 15 // number of property values per key

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

const max_CHUNK_SIZE_IN_BYTES int64 = 200 * 1000 * 1000  // 200MB
const max_PATTERN_SIZE_IN_BYTES int64 = 20 * 1000 * 1000 //20MB

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

var BLACKLIST_USER_PROPERTIES = map[string][]string{
	"all":     {},
	"non-crm": {U.HUBSPOT_PROPERTY_PREFIX, U.SALESFORCE_PROPERTY_PREFIX},
	"crm":     {},
}

var regex_NUM = regexp.MustCompile("[0-9]+")
var mineLog = taskLog.WithField("prefix", "Task#PatternMine")

type patternProperties struct {
	pattern     *P.Pattern
	count       uint
	patternType string
}

var EventNumericPropertyKeyMap = make(map[string]bool)
var UserNumericPropertyKeyMap = make(map[string]bool)

type Properties struct {
	NumericalProperties   []string            `json:"numerical_property"`
	CategoricalProperties map[string][]string `json:"categorical_property"`
}

type ChunkMetaData struct {
	Events          []string   `json:"events"`
	EventProperties Properties `json:"event_properties"`
	UserProperties  Properties `json:"user_properties"`
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

func countPatternsWorker(projectID int64, filepath string,
	patterns []*P.Pattern, wg *sync.WaitGroup, countOccurence bool, cAlgoProps P.CountAlgoProperties) {
	mineLog.Debugf("Counting patterns from File: %s", filepath)
	file, err := os.Open(filepath)
	if err != nil {
		mineLog.WithField("filePath", filepath).Error("Failure on count pattern workers.")
	}

	scanner := bufio.NewScanner(file)
	buf := make([]byte, P.MAX_PATTERN_BYTES)
	scanner.Buffer(buf, P.MAX_PATTERN_BYTES)
	P.CountPatterns(projectID, scanner, patterns, countOccurence, cAlgoProps)
	file.Close()
	wg.Done()
}

func countPatterns(projectID int64, filepath string, patterns []*P.Pattern, numRoutines int,
	countOccurence bool, cAlgoProps P.CountAlgoProperties) {
	var wg sync.WaitGroup
	numPatterns := len(patterns)
	mineLog.Debug(fmt.Sprintf("Num patterns to count Range: %d - %d", 0, numPatterns-1))
	batchSize := int(math.Ceil(float64(numPatterns) / float64(numRoutines)))
	for i := 0; i < numRoutines; i++ {
		// Each worker gets a slice of patterns to count.
		low := int(math.Min(float64(batchSize*i), float64(numPatterns)))
		high := int(math.Min(float64(batchSize*(i+1)), float64(numPatterns)))
		mineLog.Debug(fmt.Sprintf("Batch %d patterns to count range: %d:%d", i+1, low, high))
		wg.Add(1)
		go countPatternsWorker(projectID, filepath, patterns[low:high], &wg, countOccurence, cAlgoProps)
	}
	wg.Wait()
}

func computeAllUserPropertiesHistogram(projectID int64, filepath string, pattern *P.Pattern, countsVersion int, hmineSupport float32) error {
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
	return P.ComputeAllUserPropertiesHistogram(projectID, scanner, pattern, countsVersion, hmineSupport)
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
		mineLog.Debug(fmt.Sprintf("No quota. totalConsumedBytes: %d, maxTotalBytes: %d",
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
	mineLog.Debug("Number of repeapted Events in genSegmentdCandidates : ", len(cyclicEvents))

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
func mineAndWriteLenOnePatterns(projectID int64, modelId uint64, cloudManager *filestore.FileManager, diskManager *serviceDisk.DiskDriver,
	eventNames []string, filepathString string,
	userAndEventsInfo *P.UserAndEventsInfo, numRoutines int,
	chunkDir string, maxModelSize int64, cumulativePatternsSize int64,
	countOccurence bool, campaignTypeEvents CampaignEventLists, efTmpPath string, beamConfig *merge.RunBeamConfig,
	createMetadata bool, metaDataDir string, cAlgoProps P.CountAlgoProperties) (
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

	if cAlgoProps.Counting_version == 4 {
		lenOnePatterns = nil
		lenOnePatterns, err = GenLenOneV2(cAlgoProps.Job, userAndEventsInfo)
		if err != nil {
			return nil, 0, err
		}
		mineLog.Debugf("counting patterns :%v", lenOnePatterns)
	}
	mineLog.Debugf("counting patterns :%v", lenOnePatterns)

	if !beamConfig.RunOnBeam {
		countPatterns(projectID, filepathString, lenOnePatterns, numRoutines, countOccurence, cAlgoProps)
		filteredLenOnePatterns, patternsSize, err = filterAndCompressPatterns(
			lenOnePatterns, maxModelSize, cumulativePatternsSize, 1, max_PATTERN_LENGTH)

		if err != nil {
			return []*P.Pattern{}, 0, err
		}
	} else {
		//call beam
		scopeName := "Count_len_one"
		patternsFpath, err := countPatternController(beamConfig, projectID, modelId,
			cloudManager, diskManager, filepathString,
			lenOnePatterns, numRoutines, userAndEventsInfo, countOccurence,
			efTmpPath, scopeName, cAlgoProps)
		if err != nil {
			return nil, 0, err
		}
		mineLog.Debug("Reading from file")
		filteredLenOnePatterns, patternsSize, err = ReadFilterAndCompressPatternsFromFile(
			patternsFpath, cloudManager, maxModelSize, cumulativePatternsSize, 1, max_PATTERN_LENGTH)
		if err != nil {
			return []*P.Pattern{}, 0, err
		}
		mineLog.Debugf("Beam number of len one patterns : %d , ", len(lenOnePatterns))
	}

	mineLog.Debugf("patternSize : %d , ", patternsSize)

	if err := writePatternsAsChunks(filteredLenOnePatterns, chunkDir, createMetadata, metaDataDir); err != nil {
		return []*P.Pattern{}, 0, err
	}
	return filteredLenOnePatterns, patternsSize, nil
}

// FilterCombinationPatterns filter all start events based on topK logic for URL's,UDE,standardEvents,CampAnalytics
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
		mineLog.Debug("Number of patterns in ", v.EventNames[0], "->", len(endGoalPatterns))
		if len(endGoalPatterns) > 0 {
			tmpGoalPatterns, err := FilterAllCausualEvents(endGoalPatterns, eventNamesWithType)
			if err != nil {
				mineLog.Error(err)
			}
			mineLog.Debug("Number of patterns after filtering ", v.EventNames[0], len(tmpGoalPatterns))
			allPatterns = append(allPatterns, tmpGoalPatterns...)
		} else {
			mineLog.Error("No patterns for Goal", v.EventNames[0])
		}
	}
	return allPatterns
}

func mineAndWriteLenTwoPatterns(projectId int64, modelId uint64,
	lenOnePatterns []*P.Pattern, filepathString string, cloudManager *filestore.FileManager, diskManager *serviceDisk.DiskDriver,
	userAndEventsInfo *P.UserAndEventsInfo, numRoutines int,
	chunkDir string, maxModelSize int64, cumulativePatternsSize int64, countOccurence bool,
	goalPatterns []*P.Pattern, eventNamesWithType map[string]string,
	beamConfig *merge.RunBeamConfig, efTmpPath string, createMetadata bool, metaDataDir string, cAlgoProps P.CountAlgoProperties) (
	[]*P.Pattern, int64, error) {

	var patternsSize int64
	var filteredLenTwoPatterns []*P.Pattern
	var combinationGoalPatterns []*P.Pattern
	var goalFilteredLenTwoPatterns []*P.Pattern

	var err error
	var patternsFpath string
	// var countsVersion int = cAlgoProps.Counting_version
	// var hmineSupport float32 = cAlgoProps.Hmine_support

	if cAlgoProps.Counting_version == 4 {
		combinationGoalPatterns, err = CreateCombinationEventsV2(cAlgoProps.Job, userAndEventsInfo)
		if err != nil {
			return nil, 0, err
		}
	} else {
		combinationGoalPatterns, _, err = P.GenCombinationPatternsEndingWithGoal(projectId, lenOnePatterns, goalPatterns, userAndEventsInfo)
		if err != nil {
			return nil, 0, err
		}
	}

	if !beamConfig.RunOnBeam {
		countPatterns(projectId, filepathString, combinationGoalPatterns, numRoutines, countOccurence, cAlgoProps)
		filteredLenTwoPatterns, patternsSize, err = filterAndCompressPatterns(
			combinationGoalPatterns, maxModelSize, cumulativePatternsSize, 2, max_PATTERN_LENGTH)
		if err != nil {
			return []*P.Pattern{}, 0, err
		}
	} else {

		//call beam
		scopeName := "Count_len_two"
		patternsFpath, err = countPatternController(beamConfig, projectId, modelId, cloudManager, diskManager, filepathString,
			combinationGoalPatterns, numRoutines, userAndEventsInfo, countOccurence,
			efTmpPath, scopeName, cAlgoProps)
		if err != nil {
			return nil, 0, fmt.Errorf("error in counting len two patterns in beam :%v", err)
		}
		mineLog.Debug("Reading from file")
		filteredLenTwoPatterns, patternsSize, err = ReadFilterAndCompressPatternsFromFile(
			patternsFpath, cloudManager, maxModelSize, cumulativePatternsSize, 2, max_PATTERN_LENGTH)
		if err != nil {
			return []*P.Pattern{}, 0, err
		}
		mineLog.Debugf("number len Two patterns from beam: %d , ", len(combinationGoalPatterns))

	}

	// filter all combinationGoalPatterns based on start event.
	// for each goal event based on quota logic filter patterns.
	// based on the quota  startevent filter url, sme, campaign events etc.

	if cAlgoProps.Counting_version == 4 {
		goalFilteredLenTwoPatterns = filteredLenTwoPatterns
	} else {
		goalFilteredLenTwoPatterns = FilterCombinationPatterns(filteredLenTwoPatterns, goalPatterns, eventNamesWithType)
	}
	mineLog.Debug("Total Number of Len Two Patterns :", len(goalFilteredLenTwoPatterns))

	if err := writePatternsAsChunks(goalFilteredLenTwoPatterns, chunkDir, createMetadata, metaDataDir); err != nil {
		return []*P.Pattern{}, 0, err
	}
	mineLog.Debug("Total Number of Len Two Patterns :", len(goalFilteredLenTwoPatterns))
	return goalFilteredLenTwoPatterns, patternsSize, nil
}

// GetGoalPatterns get all goalPatterns from DB
func GetGoalPatterns(projectId int64, filteredPatterns []*P.Pattern, eventNamesWithType map[string]string, campEventsType CampaignEventLists, userAndEventsInfo *P.UserAndEventsInfo) ([]*P.Pattern, error) {
	goalPatternsFromDB, errCode := store.GetStore().GetAllActiveFactorsTrackedEventsByProject(projectId)

	if errCode != http.StatusFound {
		mineLog.Debug("Failure on Get goal patterns.")
	}
	var goalPatterns []*P.Pattern
	if len(goalPatternsFromDB) > 0 {

		mineLog.Debug(fmt.Sprintf("Number of Goals from DB:%d", len(goalPatternsFromDB)))

		tmpPatterns := make(map[string]*P.Pattern)

		for _, v := range filteredPatterns {
			tmpPatterns[v.String()] = v
		}
		for _, p := range goalPatternsFromDB {
			if valPattern, ok := tmpPatterns[p.Name]; ok {
				goalPatterns = append(goalPatterns, valPattern)
				mineLog.Debugf("Goal event from db :%s", valPattern.String())
			} else {
				mineLog.Debugf("Goal event from db not available in filtered lenOne Pattern :%s", p.Name)
				if _, ok := eventNamesWithType[p.Name]; ok {
					p, err := P.NewPattern([]string{p.Name}, userAndEventsInfo)
					if err != nil {
						return []*P.Pattern{}, nil
					}
					goalPatterns = append(goalPatterns, p)
					mineLog.Debug(fmt.Sprint("Goal event from db added from events in eventsFile ", p.String()))

				}

			}

		}

		return goalPatterns, nil
	}

	goalTopKPatterns := FilterTopKEventsOnTypes(filteredPatterns, eventNamesWithType, topK_patterns, keventsSpecial, keventsURL, campEventsType)
	mineLog.Debug("Mining goals from topk events : ", len(goalTopKPatterns))

	for idx, valPat := range goalTopKPatterns {
		mineLog.Debug(fmt.Sprint("Insering in DB Goal event: ", idx, valPat.String()))
		tmpFactorsRule := model.FactorsGoalRule{EndEvent: valPat.String()}
		goalID, httpStatusTrackedEvent := store.GetStore().CreateFactorsTrackedEvent(projectId, valPat.String(), "")
		if goalID == 0 {
			mineLog.Debug("Unable to create a trackedEvent ", httpStatusTrackedEvent, " ", goalID)
		}
		mineLog.Debug("trackedEvent in db  ", httpStatusTrackedEvent, " ", valPat.String(), " ", goalID)

		_, httpstatus, err := store.GetStore().CreateFactorsGoal(projectId, valPat.String(), tmpFactorsRule, "")
		if httpstatus != http.StatusCreated {
			mineLog.Debug("Unable to create factors goal in db: ", httpstatus, " ", valPat.String(), " ", err)
		}

	}

	return goalTopKPatterns, nil

}

// GenMissingTwoLenPatterns get threeLen-twoLen
func GenMissingJourneyPatterns(goal, journey []*P.Pattern, userAndEventsInfo *P.UserAndEventsInfo) ([]*P.Pattern, error) {
	var lenGoal, lenJourney int
	if len(goal) > 0 && len(journey) > 0 {
		lenGoal = len(goal[0].EventNames)
		lenJourney = len(journey[0].EventNames)
		if lenJourney > lenGoal {
			return nil, fmt.Errorf("len of Journey is greater than goal")
		}
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

	mineLog.Debug("Number of missing patterns before filtering :", len(missingPatt))

	allMissingPatt := make([]*P.Pattern, 0)
	tmpMissingMap := make(map[string]bool)
	// add dedupe logic
	for _, v := range missingPatt {
		if !tmpMissingMap[strings.Join(v.EventNames, "_")] {
			allMissingPatt = append(allMissingPatt, v)
			tmpMissingMap[strings.Join(v.EventNames, "_")] = true
		}

	}
	mineLog.Debug("Number of missing patterns :", len(allMissingPatt))

	return allMissingPatt, nil
}

func mineAndWritePatterns(projectId int64, modelId uint64, filepath string,
	userAndEventsInfo *P.UserAndEventsInfo, cloudManager *filestore.FileManager, diskManager *serviceDisk.DiskDriver, eventNames []string,
	numRoutines int, chunkDir string,
	maxModelSize int64, countOccurence bool,
	eventNamesWithType map[string]string, repeatedEvents []string,
	campTypeEvents CampaignEventLists, efTmpPath string, beamConfig *merge.RunBeamConfig,
	createMetadata bool, metaDataDir string, cAlgoProps P.CountAlgoProperties) error {
	var filteredPatterns []*P.Pattern
	var cumulativePatternsSize int64 = 0

	patternLen := 1
	limitRoundOffFraction := 0.99

	filteredPatterns, patternsSize, err := mineAndWriteLenOnePatterns(projectId, modelId, cloudManager, diskManager,
		eventNames, filepath, userAndEventsInfo, numRoutines, chunkDir,
		maxModelSize, cumulativePatternsSize, countOccurence, campTypeEvents, efTmpPath, beamConfig, createMetadata, metaDataDir, cAlgoProps)
	if err != nil {
		return fmt.Errorf("unable to mine len one patterns :%v", err)
	}
	cumulativePatternsSize += patternsSize
	printFilteredPatterns(filteredPatterns, patternLen)
	mineLog.Debug("Number of Len-one Patterns : ", len(filteredPatterns))
	// Get all Goal Patterns => DB Patterns + CampaignAnalytics (Campaign,source,medium,referr)
	goalPatterns, err := GetGoalPatterns(projectId, filteredPatterns, eventNamesWithType, campTypeEvents, userAndEventsInfo)
	if err != nil {
		return fmt.Errorf("unable to get goal patterns :%v", err)
	}
	mineLog.Debug("Number of Goal Patterns to use in factors: ", len(goalPatterns))
	if cumulativePatternsSize >= int64(float64(maxModelSize)*limitRoundOffFraction) {
		return nil
	}
	for _, v := range goalPatterns {
		mineLog.Debug("Goal Patterns :->", v.EventNames)
	}

	patternLen++
	if patternLen > max_PATTERN_LENGTH {
		return nil
	}
	filteredTwoPatterns, patternsSize, err := mineAndWriteLenTwoPatterns(projectId, modelId,
		filteredPatterns, filepath, cloudManager, diskManager, userAndEventsInfo,
		numRoutines, chunkDir, maxModelSize, cumulativePatternsSize,
		countOccurence, goalPatterns, eventNamesWithType,
		beamConfig, efTmpPath, createMetadata, metaDataDir, cAlgoProps)
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

		if cAlgoProps.Counting_version == 4 {
			lenThreePatterns, err = GenThreeLenEventsV2(cAlgoProps.Job, userAndEventsInfo)
			if err != nil {
				return err
			}
		}

		var filteredThreePatterns []*P.Pattern
		if !beamConfig.RunOnBeam {
			countPatterns(projectId, filepath, lenThreePatterns, numRoutines, countOccurence, cAlgoProps)
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
			patternsFpath, err = countPatternController(beamConfig, projectId, modelId, cloudManager, diskManager, filepath,
				lenThreePatterns, numRoutines, userAndEventsInfo, countOccurence,
				efTmpPath, scopeName, cAlgoProps)
			if err != nil {
				return fmt.Errorf("unable to count three len patterns :%v", err)
			}
			mineLog.Debug("Reading from file")
			filteredThreePatterns, patternsSize, err = ReadFilterAndCompressPatternsFromFile(
				patternsFpath, cloudManager, maxModelSize, cumulativePatternsSize,
				patternLen, max_PATTERN_LENGTH)
			if err != nil {
				return err
			}
			mineLog.Debugf("number len Three patterns from beam: %d , ", len(lenThreePatterns))

		}

		cumulativePatternsSize += patternsSize
		if err := writePatternsAsChunks(filteredThreePatterns, chunkDir, createMetadata, metaDataDir); err != nil {
			return err
		}
		printFilteredPatterns(filteredThreePatterns, patternLen)

		// count missing two len patterns in three len patterns
		missingPatternsTwo, err := GenMissingJourneyPatterns(filteredThreePatterns, filteredTwoPatterns, userAndEventsInfo)

		if err != nil {
			mineLog.Error("Unable to create missing two len pattern")
		}
		for _, p := range missingPatternsTwo {
			if cAlgoProps.Counting_version == 4 {
				p.Segment = 0
			}
		}

		// count  - two len missing
		var filteredMissingPatterns []*P.Pattern
		if !beamConfig.RunOnBeam {
			countPatterns(projectId, filepath, missingPatternsTwo, numRoutines, countOccurence, cAlgoProps)
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
				modelId, cloudManager, diskManager, filepath,
				missingPatternsTwo, numRoutines, userAndEventsInfo,
				countOccurence, efTmpPath, scopeName, cAlgoProps)
			if err != nil {
				return fmt.Errorf("unable to count len two missing :%v", err)
			}
			mineLog.Debugf("Reading from part Directory : %s", scopeName)
			filteredMissingPatterns, patternsSize, err = ReadFilterAndCompressPatternsFromFile(patternsFpath, cloudManager, maxModelSize, cumulativePatternsSize, patternLen, max_PATTERN_LENGTH)
			if err != nil {
				return err
			}
			mineLog.Debugf("number len Two missing patterns from beam: %d , ", len(missingPatternsTwo))
		}

		cumulativePatternsSize += patternsSize
		if err := writePatternsAsChunks(filteredMissingPatterns, chunkDir, createMetadata, metaDataDir); err != nil {
			return err
		}
		printFilteredPatterns(filteredMissingPatterns, patternLen-1)

		if cumulativePatternsSize >= int64(float64(maxModelSize)*limitRoundOffFraction) {
			return nil
		}

		if cAlgoProps.Counting_version == 4 {
			// count four len patterns
			lenFourPatterns, err := CreateFourLenEventsV2(cAlgoProps.Job, userAndEventsInfo)
			if err != nil {
				return fmt.Errorf("unable to gen three level patterns v2")
			}
			mineLog.Debugf("len four patterns - %v", lenFourPatterns)
			time.Sleep(10 * time.Second)
			// count  - len four patterns
			var lenFourPatts []*P.Pattern
			if !beamConfig.RunOnBeam {
				countPatterns(projectId, filepath, lenFourPatterns, numRoutines, countOccurence, cAlgoProps)
				lenFourPatts, patternsSize, err = filterAndCompressPatterns(
					missingPatternsTwo, maxModelSize, cumulativePatternsSize,
					patternLen, max_PATTERN_LENGTH)
				if err != nil {
					return err
				}
			} else {
				//call beam
				var patternsFpath string
				scopeName := "CountFourLen"
				patternsFpath, err = countPatternController(beamConfig, projectId,
					modelId, cloudManager, diskManager, filepath,
					lenFourPatterns, numRoutines, userAndEventsInfo,
					countOccurence, efTmpPath, scopeName, cAlgoProps)
				if err != nil {
					return fmt.Errorf("unable to count len four :%v", err)
				}
				mineLog.Debugf("Reading from part Directory : %s", scopeName)
				lenFourPatts, patternsSize, err = ReadFilterAndCompressPatternsFromFile(patternsFpath, cloudManager, maxModelSize, cumulativePatternsSize, patternLen, max_PATTERN_LENGTH)
				if err != nil {
					return err
				}
				mineLog.Debugf("number len four patterns from beam: %d , ", len(missingPatternsTwo))
			}

			cumulativePatternsSize += patternsSize
			if err := writePatternsAsChunks(lenFourPatts, chunkDir, createMetadata, metaDataDir); err != nil {
				return err
			}
			printFilteredPatterns(lenFourPatts, patternLen-1)

			if cumulativePatternsSize >= int64(float64(maxModelSize)*limitRoundOffFraction) {
				return nil
			}
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
		countPatterns(projectId, filepath, candidatePatterns, numRoutines, countOccurence, cAlgoProps)
		filteredPatterns, patternsSize, err = filterAndCompressPatterns(
			candidatePatterns, maxModelSize, cumulativePatternsSize,
			patternLen, max_PATTERN_LENGTH)
		if err != nil {
			return err
		}
		if len(filteredPatterns) > 0 {
			cumulativePatternsSize += patternsSize
			if err := writePatternsAsChunks(filteredPatterns, chunkDir, createMetadata, metaDataDir); err != nil {
				return err
			}
		}
		printFilteredPatterns(filteredPatterns, patternLen)
	}

	return nil
}

func buildPropertiesInfoFromInput(cloudManager *filestore.FileManager, projectId int64, modelId uint64, eventNames []string, filepath string) (*P.UserAndEventsInfo, map[string]P.PropertiesCount, error) {
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

	fileDir, fileName := (*cloudManager).GetModelEventPropertiesFilePathAndName(projectId, modelId)
	epCount, err := P.GetPropertiesMapFromFile(cloudManager, fileDir, fileName)
	if err != nil {
		log.Error("Error reading event properties from File")
		return userAndEventsInfo, nil, err
	}

	fileDir, fileName = (*cloudManager).GetModelUserPropertiesFilePathAndName(projectId, modelId)
	upCount, err := P.GetPropertiesMapFromFile(cloudManager, fileDir, fileName)
	if err != nil {
		log.Error("Error reading user properties from File")
		return userAndEventsInfo, nil, err
	}

	allProperty, err := P.CollectPropertiesInfoFiltered(projectId, scanner, userAndEventsInfo, upCount, epCount)
	if err != nil {
		return nil, nil, err
	}
	return userAndEventsInfo, *allProperty, nil
}

func printFilteredPatterns(filteredPatterns []*P.Pattern, iter int) {
	mineLog.Debug(fmt.Sprintf("Mined %d patterns of length %d", len(filteredPatterns), iter))

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

func writeEventInfoFile(projectId int64, modelId uint64, events *bytes.Reader,
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
func writePatternsAsChunks(patterns []*P.Pattern, chunksDir string, createMetadataForChunks bool, metaDataDir string) error {
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
		if pBytesLen >= max_PATTERN_SIZE_IN_BYTES {
			// Limit is 20MB
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
	if createMetadataForChunks {
		metaData := GetChunkMetaData(patterns)
		WriteMetaDataHandler(metaData, metaDataDir)
	}
	return nil
}
func WriteMetaDataHandler(metaData []ChunkMetaData, metaDataDir string) {
	err := WriteMetaData(metaData, metaDataDir)
	if err != nil {
		log.WithError(err).Error("failed to write metadata")
		return
	}
	log.Info("metaData written succesfully")
}
func WriteMetaData(metaData []ChunkMetaData, metaDataDir string) error {
	path := metaDataDir + "metadata.txt"
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		mineLog.WithError(err).Error("Failed to open metadata file.")
		return err
	}
	defer file.Close()
	w := bufio.NewWriter(file)
	for _, data := range metaData {
		bytes, err := json.Marshal(data)
		if err != nil {
			log.WithError(err).Error("failed to marshal meta data")
			continue
		}
		lineWrite := string(bytes)
		if _, err := w.WriteString(lineWrite + "\n"); err != nil {
			log.WithError(err).Error("Unable to write metadata to file.")
			return err
		}
	}
	err = w.Flush()
	if err != nil {
		log.Error(err)
		return err
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
				mineLog.WithError(err).Error("Failed to upload chunk file to cloud.")
				return uploadedChunkIds, err
			}
			uploadedChunkIds = append(uploadedChunkIds, chunkIdStr)
		}
	}

	return uploadedChunkIds, nil
}

func uploadMetaDataToCloud(metaDataDir, cloudMetaDataDir string, cloudManager *filestore.FileManager) error {
	files, err := ioutil.ReadDir(metaDataDir)
	if err != nil {
		log.WithError(err).Error("failed to read metadata directory")
		return err
	}
	for _, cf := range files {
		mdfName := cf.Name()
		mdfPath := fmt.Sprintf("%s/%s", metaDataDir, mdfName)
		mdFile, err := os.OpenFile(mdfPath, os.O_RDWR, 0666)
		if err != nil {
			log.WithError(err).Error("Failed to open tmp metadata to upload.")
			return err
		}
		err = (*cloudManager).Create(cloudMetaDataDir, mdfName, mdFile)
		if err != nil {
			mineLog.WithError(err).Error("Failed to upload metadata file to cloud.")
			return err
		}
	}
	return nil
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

func rewriteEventsFile(tmpEventsFilePath string, tmpPath string, cloudManager *filestore.FileManager,
	userAndEventsInfo *P.UserAndEventsInfo, projectId int64, modelId uint64, cAlgoProps P.CountAlgoProperties) (CampaignEventLists, error) {
	// read events file , filter and create properties based on userProp and eventsProp
	// create encoded events based on $session and campaign eventName
	var fileDir, fileName string
	var err error
	var upCatg, epCatg map[string]string
	var numlinesv2 = 0
	var numfilterEvent = 0

	events_to_include_v2 := make(map[string]int)

	if cAlgoProps.Counting_version == 4 {
		// put all events to be included into one single map
		jb := cAlgoProps.Job
		events_to_include_v2[jb.Query.StartEvent] = 1
		events_to_include_v2[jb.Query.EndEvent] = 1
		for _, p := range jb.Query.Rule.IncludedEvents {
			events_to_include_v2[p] = 1
		}

	}

	log.Infof("Filters applied :%v", cAlgoProps.Job.Query)

	fileDir, fileName = (*cloudManager).GetModelUserPropertiesCategoricalFilePathAndName(projectId, modelId)
	upCatg, err = P.GetPropertiesCategoricalMapFromFile(cloudManager, fileDir, fileName)
	if err != nil {
		mineLog.Error("Error reading type of user properties from File")
		return CampaignEventLists{}, err
	}

	fileDir, fileName = (*cloudManager).GetModelEventPropertiesCategoricalFilePathAndName(projectId, modelId)
	epCatg, err = P.GetPropertiesCategoricalMapFromFile(cloudManager, fileDir, fileName)
	if err != nil {
		mineLog.Error("Error reading type of event properties from File")
		return CampaignEventLists{}, err
	}

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

	events, _ := store.GetStore().GetSmartEventFilterEventNames(projectId, true)
	crmEvents := make(map[string]bool)
	for _, event := range events {
		crmEvents[event.Name] = true
	}

	delEvent := 0
	delUser := 0

	log.Infof("--------------- rewriting events file ---- $$$$$---\n")
	w := bufio.NewWriter(file)
	for scanner.Scan() {
		line := scanner.Text()
		var eventDetails P.CounterEventFormat
		if err := json.Unmarshal([]byte(line), &eventDetails); err != nil {
			log.WithFields(log.Fields{"line": line, "err": err}).Error("Read failed")
			return CampaignEventLists{}, err
		}
		ename := eventDetails.EventName

		for uKey, uVal := range eventDetails.UserProperties {
			propertyType := store.GetStore().GetPropertyTypeByKeyValue(projectId, ename, uKey, uVal, true)
			if catg, ok := upCatg[uKey]; ok && catg != propertyType {
				delete(eventDetails.UserProperties, uKey)
				delUser += 1
				continue
			}
			if P.IsCustomOrCrmEvent(ename, crmEvents) {
				if U.HasPrefixFromList(uKey, BLACKLIST_USER_PROPERTIES["crm"]) {
					delete(eventDetails.UserProperties, uKey)
					delUser += 1
					continue
				}
			} else {
				if U.HasPrefixFromList(uKey, BLACKLIST_USER_PROPERTIES["non-crm"]) {
					delete(eventDetails.UserProperties, uKey)
					delUser += 1
					continue
				}
			}
			uInfo := userAndEventsInfo.UserPropertiesInfo
			if catg, ok := upCatg[uKey]; ok {
				if catg == U.PropertyTypeCategorical {
					pstring := U.GetPropertyValueAsString(uVal)
					// propString := strings.Join([]string{uKey, pstring}, "::")
					if valMap, ok := uInfo.CategoricalPropertyKeyValues[uKey]; !ok || removePropertiesTree(pstring) {
						delete(eventDetails.UserProperties, uKey)
						delUser += 1
					} else {
						if _, ok := valMap[pstring]; !ok {
							delete(eventDetails.UserProperties, uKey)
							delUser += 1
						}
					}
				} else {
					if _, ok := uInfo.NumericPropertyKeys[uKey]; !ok {
						delete(eventDetails.UserProperties, uKey)
						delUser += 1
					}
				}
			}

		}

		for eKey, eVal := range eventDetails.EventProperties {
			eInfoMap := userAndEventsInfo.EventPropertiesInfoMap
			if catg, ok := epCatg[eKey]; ok {
				if catg == U.PropertyTypeCategorical {

					pstring := U.GetPropertyValueAsString(eVal)
					// propString := strings.Join([]string{eKey, pstring}, "::")
					if valMap, ok := (*eInfoMap)[ename].CategoricalPropertyKeyValues[eKey]; !ok || removePropertiesTree(pstring) {
						delete(eventDetails.EventProperties, eKey)
						delEvent += 1
					} else {
						if _, ok := valMap[pstring]; !ok {
							delete(eventDetails.UserProperties, eKey)
							delUser += 1
						}
					}
				} else {
					if _, ok := (*eInfoMap)[ename].NumericPropertyKeys[eKey]; !ok {
						delete(eventDetails.UserProperties, eKey)
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

		if cAlgoProps.Counting_version != 4 {
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

		} else {
			if _, ok := events_to_include_v2[ename]; ok {
				if FilterEventsOnRule(eventDetails, cAlgoProps.Job, upCatg, epCatg) {

					lineWrite := string(eventDetailsBytes)
					if _, err := file.WriteString(fmt.Sprintf("%s\n", lineWrite)); err != nil {
						peLog.WithFields(log.Fields{"line": line, "err": err}).Error("Unable to write to file.")
						return CampaignEventLists{}, err
					}
					numlinesv2 += 1
				}
			} else {

				if numfilterEvent < 40 {
					if strings.Compare(ename, "$session") == 0 {
						lineWrite := string(eventDetailsBytes)
						log.Infof("$session event ---->%s", lineWrite)

					}
					numfilterEvent += 1
				}
			}

		}
	}
	w.Flush()
	if cAlgoProps.Counting_version == 4 {
		log.Infof("total number of lines rewritten is :%d", numlinesv2)
	}

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
	mineLog.Debugf("number of user prop deleted:%d", delUser)
	mineLog.Debugf("number of event prop deleted:%d", delEvent)
	return smartEvents, nil
}

func writeEncodedEvent(eventName string, property string, propertyName string, propEventsMap map[string]int, eventDetails P.CounterEventFormat, file *os.File, line string) error {
	//check if eventName and eventDetails.EventName match and property exists in eventDetails.EventProperties
	//if exists with non-empty value, write encoded event to file

	if strings.Compare(eventDetails.EventName, eventName) == 0 && eventDetails.EventProperties[property] != nil {

		var tmpEvent P.CounterEventFormat
		tmpEventId, ok := eventDetails.EventProperties[property].(string)

		if !ok {
			mineLog.Debug("Error in converting string : ", " ", strings.ToUpper(propertyName), " ", eventDetails.EventProperties[property])
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

func GetEventNamesAndType(tmpEventsFilePath string, cloudManager *filestore.FileManager, projectId int64, modelId uint64, countsVersion int) ([]string, map[string]string, map[string]int, error) {
	scanner, err := OpenEventFileAndGetScanner(tmpEventsFilePath)
	if err != nil {
		return nil, nil, nil, err
	}
	eventNames, eventsCount, err := GetEventNamesFromFile(scanner, cloudManager, projectId, modelId, countsVersion)
	if err != nil {
		mineLog.WithFields(log.Fields{"err": err, "eventFilePath": tmpEventsFilePath}).Error("Failed to read event names from file")
		return nil, nil, nil, err
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
		return nil, nil, nil, err
	}

	return eventNames, eventNamesWithType, eventsCount, nil
}

// buildWhiteListProperties build user and event properties from db ,
// if not available in DB, use counting logic to choose topk properties
func buildWhiteListProperties(projectId int64, allProperty map[string]P.PropertiesCount, numProp int) (map[string]bool, map[string]bool) {

	userPropertiesMap := make(map[string]int)
	eventPropertiesMap := make(map[string]int)
	upFilteredMap := make(map[string]bool)
	epFilteredMap := make(map[string]bool)

	for _, v := range allProperty {
		if strings.Compare(v.PropertyType, "UP") == 0 {
			if U.HasPrefixFromList(v.PropertyName, BLACKLIST_USER_PROPERTIES["all"]) {
				continue
			}
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
				mineLog.Debug("Missing user property in events File or blacklisted: ", v.UserPropertyName)
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
			mineLog.Debug("insert user property", key)
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

func buildEventsFileOnProperties(tmpEventsFilePath string, efTmpPath string, efTmpName string, cloudManager *filestore.FileManager, diskManager *serviceDisk.DiskDriver, projectId int64,
	modelId uint64, eReader io.Reader, userAndEventsInfo *P.UserAndEventsInfo, campaignLimitCount int, cAlgoProps P.CountAlgoProperties) (CampaignEventLists, error) {
	// Rewrites the events file restricting the properties to only whitelisted properties.
	//  Also returns list special campaign events created during rewrite.
	var err error
	efPath := efTmpPath + "tmpevents_" + efTmpName
	smartEvents, err := rewriteEventsFile(tmpEventsFilePath, efPath, cloudManager, userAndEventsInfo, projectId, modelId, cAlgoProps)
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
			mineLog.Debug("Intializing campaign events : ", ct[0])
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
func PatternMine(db *gorm.DB, etcdClient *serviceEtcd.EtcdClient, archiveCloudManager, tmpCloudManager, sortedCloudManager, modelCloudManager *filestore.FileManager,
	diskManager *serviceDisk.DiskDriver, numRoutines int, projectId int64,
	modelId uint64, modelType string, startTime int64, endTime int64, maxModelSize int64,
	countOccurence bool, campaignLimitCount int, beamConfig *merge.RunBeamConfig,
	createMetadata bool, cAlgoProps P.CountAlgoProperties, hardPull bool, useBucketV2 bool) (int, error) {

	var err error
	countsVersion := cAlgoProps.Counting_version
	hmineSupport := cAlgoProps.Hmine_support

	metaDataDir := diskManager.GetChunksMetaDataDir(projectId, modelId)
	if createMetadata {
		metaDataFileName := "metadata.txt"
		var data []byte
		reader := bytes.NewReader(data)
		diskManager.Create(metaDataDir, metaDataFileName, reader)
	}
	if useBucketV2 {
		if err := merge.MergeAndWriteSortedFile(projectId, U.DataTypeEvent, "", startTime, endTime,
			archiveCloudManager, tmpCloudManager, sortedCloudManager, diskManager, beamConfig, hardPull, 0); err != nil {
			mineLog.WithError(err).Error("Failed creating events file")
			return 0, err
		}
	}

	if cAlgoProps.Counting_version == 4 {
		mineLog.Debugf("explain v2 job :%v", cAlgoProps.Job)
	}

	// download events file from cloud to local
	efCloudPath, efCloudName := (*sortedCloudManager).GetEventsFilePathAndName(projectId, startTime, endTime)
	efTmpPath, efTmpName := diskManager.GetEventsFilePathAndName(projectId, startTime, endTime)
	mineLog.WithFields(log.Fields{"eventFileCloudPath": efCloudPath,
		"eventFileCloudName": efCloudName}).Info("Downloading events file from cloud.")
	eReader, err := (*sortedCloudManager).Get(efCloudPath, efCloudName)
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
	tmpChunksDir := diskManager.GetPatternChunksDir(projectId, modelId)
	if err := serviceDisk.MkdirAll(tmpChunksDir); err != nil {
		mineLog.WithFields(log.Fields{"chunkDir": tmpChunksDir, "error": err}).Error("Unable to create chunks directory.")
		return 0, err
	}

	tmpEventsFilepath := efTmpPath + efTmpName
	mineLog.Debug("Successfuly downloaded events file from cloud.", tmpEventsFilepath, efTmpPath, efTmpName)

	basePathUserProps := path.Join("/", "tmp", "fptree", "userProps")
	basePathEventProps := path.Join("/", "tmp", "fptree", "eventProps")
	err = os.MkdirAll(basePathUserProps, os.ModePerm)
	if err != nil {
		log.Fatal("unable to create temp user directory")
	}
	err = os.MkdirAll(basePathEventProps, os.ModePerm)
	if err != nil {
		log.Fatal("unable to create temp event directory")
	}

	eventNames, eventNamesWithType, events_count_v2, err := GetEventNamesAndType(tmpEventsFilepath, modelCloudManager, projectId, modelId, countsVersion)
	if err != nil {
		mineLog.WithFields(log.Fields{"err": err}).Error("Failed to get eventName and event type.")
		return 0, err
	}

	if cAlgoProps.Counting_version == 4 {
		// add top 50 events in case if events to include in nill
		jb := cAlgoProps.Job
		mineLog.Debug("before counting all:%v", jb)
		addIncludedEventsV2(&jb, events_count_v2)
		mineLog.Debug("before counting all:%v", jb)
		cAlgoProps.Job = jb
		mineLog.Debug("before counting all:%v", cAlgoProps.Job)
	}

	mineLog.Debugf("job details :%v", cAlgoProps.Job)
	userAndEventsInfo, allPropsMap, err := buildPropertiesInfoFromInput(modelCloudManager, projectId, modelId, eventNames, tmpEventsFilepath)
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
	err = writeEventInfoFile(projectId, modelId, bytes.NewReader(userAndEventsInfoBytes), (*modelCloudManager))
	if err != nil {
		mineLog.WithFields(log.Fields{"err": err}).Error("Failed to write events Info.")
		return 0, err
	}

	mineLog.Debug("Number of EventNames: ", len(eventNames))
	mineLog.Debug("Number of User Properties: ", len(userPropList))
	mineLog.Debug("Number of Event Properties: ", len(eventPropList))

	campaignAnalyticsSorted, err := buildEventsFileOnProperties(tmpEventsFilepath, efTmpPath, efTmpName, modelCloudManager, diskManager, projectId,
		modelId, eReader, userAndEventsInfo, campaignLimitCount, cAlgoProps)
	if err != nil {
		mineLog.WithFields(log.Fields{"err": err}).Error("Failed to write events data.")
		return 0, err
	}

	userAndEventsInfo = buildEventsInfoForEncodedEvents(campaignAnalyticsSorted, userAndEventsInfo)
	mineLog.Debug("Successfully Built events data and written to file.")
	mineLog.Debug("Number of Campaign events:", len(campaignAnalyticsSorted.CampaignList))
	mineLog.Debug("Number of Medium events:", len(campaignAnalyticsSorted.MediumList))
	mineLog.Debug("Number of Referrer events:", len(campaignAnalyticsSorted.ReferrerList))
	mineLog.Debug("Number of source events:", len(campaignAnalyticsSorted.SourceList))
	mineLog.Debug("Number of Adgroup events:", len(campaignAnalyticsSorted.AdgroupList))

	repeatedEvents := GetAllRepeatedEvents(eventNames, campaignAnalyticsSorted)
	mineLog.Debug("Number of repeated Events: ", len(repeatedEvents))
	mineLog.Debug("Repeated Events", repeatedEvents)

	//write events file to GCP
	modEventsFile := fmt.Sprintf("events_modified_%d.txt", modelId)
	err = writeFileToGCP(projectId, modelId, modEventsFile, tmpEventsFilepath, modelCloudManager, "")
	if err != nil {
		return 0, fmt.Errorf("unable to write modified events file to GCP")
	}

	allActiveUsersPattern, err := P.NewPattern([]string{U.SEN_ALL_ACTIVE_USERS}, userAndEventsInfo)

	if !beamConfig.RunOnBeam {
		// build histogram of all user properties.
		mineLog.WithField("tmpEventsFilePath", tmpEventsFilepath).Info("Building all user properties histogram.")
		if err != nil {
			mineLog.WithFields(log.Fields{"err": err}).Error("Failed to build pattern with histogram of all active user properties.")
			return 0, err
		}
		if err := computeAllUserPropertiesHistogram(projectId, tmpEventsFilepath, allActiveUsersPattern, countsVersion, hmineSupport); err != nil {
			mineLog.WithFields(log.Fields{"err": err}).Error("Failed to compute user properties.")
			return 0, err
		}
		if err := writePatternsAsChunks([]*P.Pattern{allActiveUsersPattern}, tmpChunksDir, createMetadata, metaDataDir); err != nil {
			mineLog.WithFields(log.Fields{"err": err}).Error("Failed to write user properties.")
			return 0, err
		}
	} else {
		//call beam
		scopeName := "Count_user_prop_hist"
		patts_all_activeUsers := make([]*P.Pattern, 0)
		patts_all_activeUsers = append(patts_all_activeUsers, allActiveUsersPattern)
		all_active_patternsFpath, err := UserPropertiesHistogramController(beamConfig, projectId, modelId, modelCloudManager, diskManager, tmpEventsFilepath,
			patts_all_activeUsers, numRoutines, userAndEventsInfo, countOccurence,
			efTmpPath, scopeName, cAlgoProps)
		if err != nil {
			return 0, fmt.Errorf("error in counting len two patterns in beam :%v", err)
		}
		mineLog.Debug("Reading from file")
		allActiveUsersPatterns, _, err := ReadFilterAndCompressPatternsFromFile(
			all_active_patternsFpath, modelCloudManager, maxModelSize, 0, 2, max_PATTERN_LENGTH)
		if err != nil {
			return 0, err
		}
		if len(allActiveUsersPatterns) > 0 {
			allActiveUsersPattern := allActiveUsersPatterns[0]
			if err := writePatternsAsChunks([]*P.Pattern{allActiveUsersPattern}, tmpChunksDir, createMetadata, metaDataDir); err != nil {
				mineLog.WithFields(log.Fields{"err": err}).Error("Failed to write user properties.")
				return 0, err
			}
		}

	}

	mineLog.Debug("Successfully built all user properties histogram.")

	// mine and write patterns as chunks
	mineLog.WithFields(log.Fields{"projectId": projectId, "tmpEventsFilepath": tmpEventsFilepath,
		"tmpChunksDir": tmpChunksDir, "routines": numRoutines}).Info("Mining patterns and writing it as chunks.")

	err = mineAndWritePatterns(projectId, modelId, tmpEventsFilepath,
		userAndEventsInfo, modelCloudManager, diskManager, eventNames, numRoutines, tmpChunksDir, maxModelSize,
		countOccurence, eventNamesWithType, repeatedEvents, campaignAnalyticsSorted,
		efTmpPath, beamConfig, createMetadata, metaDataDir, cAlgoProps)
	if err != nil {
		mineLog.WithFields(log.Fields{"err": err}).Error("Failed to mine patterns.")
		return 0, err
	}
	mineLog.Debug("Successfully mined patterns and written it as chunks.")

	var chunkIds []string
	if cAlgoProps.Counting_version != 4 {
		// upload chunks to cloud
		cloudChunksDir := (*modelCloudManager).GetPatternChunksDir(projectId, modelId)
		mineLog.WithFields(log.Fields{"tmpChunksDir": tmpChunksDir,
			"cloudChunksDir": cloudChunksDir}).Info("Uploading chunks to cloud.")
		chunkIds, err := uploadChunksToCloud(tmpChunksDir, cloudChunksDir, modelCloudManager)
		if err != nil {
			mineLog.WithFields(log.Fields{"localChunksDir": tmpChunksDir,
				"cloudChunksDir": cloudChunksDir}).Error("Failed to upload chunks to cloud.")
		}
		mineLog.Debug("Successfully uploaded chunks to cloud.")

		// upload metadata to cloud
		if createMetadata {
			cloudMetaDataDir := (*modelCloudManager).GetChunksMetaDataDir(projectId, modelId)
			mineLog.WithFields(log.Fields{"MetaDataDir": metaDataDir,
				"cloudMetaDataDir": cloudMetaDataDir}).Info("Uploading metadata to cloud.")
			err = uploadMetaDataToCloud(metaDataDir, cloudMetaDataDir, modelCloudManager)
			if err != nil {
				mineLog.WithFields(log.Fields{"localMetaDataDir": metaDataDir,
					"cloudMetaDataDir": cloudMetaDataDir}).Error("Failed to upload metadata to cloud.")
			}
			mineLog.Debug("Successfully uploaded metadata to cloud.")
		}

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
		if cAlgoProps.Counting_version != 0 {
			err = etcdClient.SetProjectVersion(newVersionId)
			if err != nil {
				mineLog.WithFields(log.Fields{"err": err}).Error("Failed to write new version id to etcd.")
				return 0, err
			}
		}
	} else {

		// upload chunks to cloud
		cloudChunksDir, _ := (*modelCloudManager).GetExplainV2ModelPath(modelId, projectId)

		mineLog.WithFields(log.Fields{"tmpChunksDir": tmpChunksDir,
			"cloudChunksDir": cloudChunksDir}).Info("Uploading chunks to cloud.")
		chunkIds, err = uploadChunksToCloud(tmpChunksDir, cloudChunksDir, modelCloudManager)
		if err != nil {
			mineLog.WithFields(log.Fields{"localChunksDir": tmpChunksDir,
				"cloudChunksDir": cloudChunksDir}).Error("Failed to upload chunks to cloud.")
		}
		mineLog.Debug("Successfully Explain v2 uploaded chunks to cloud.")
		store.GetStore().UpdateExplainV2EntityStatus(projectId, cAlgoProps.JobId, model.ACTIVE, modelId)

	}

	localEventsFilePath := filepath.Join(efTmpPath, efTmpName)
	err = deleteFile(localEventsFilePath)
	if err != nil {
		return 0, err
	}
	mineLog.Debugf("Deleted local events file :%s", localEventsFilePath)

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
	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		log.Fatal(err)
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

func FilterEventsInfoNew(userAndEventsInfo *P.UserAndEventsInfo, userProp, eventProp map[string]bool, upCount map[string]map[string]int, epCount map[string]map[string]map[string]int) *P.UserAndEventsInfo {

	userPropertiesInfo := userAndEventsInfo.UserPropertiesInfo
	eventPropertiesInfo := *userAndEventsInfo.EventPropertiesInfoMap

	//delete both categorical and numerical properties for users
	for propertyName, vals := range userPropertiesInfo.CategoricalPropertyKeyValues {

		if !userProp[propertyName] {
			delete(userPropertiesInfo.CategoricalPropertyKeyValues, propertyName)
		} else {
			for propertyValue := range vals {
				if _, ok := upCount[propertyName][propertyValue]; !ok {
					delete(userPropertiesInfo.CategoricalPropertyKeyValues, propertyName)
				}
			}
		}
	}

	for propertyName := range userPropertiesInfo.NumericPropertyKeys {

		if !userProp[propertyName] {
			delete(userPropertiesInfo.NumericPropertyKeys, propertyName)
		}
	}

	//delete both categorical and numerical properties for events
	for event, property := range eventPropertiesInfo {

		for k, vals := range property.CategoricalPropertyKeyValues {

			if !eventProp[k] {
				delete(property.CategoricalPropertyKeyValues, k)
			} else {
				for propertyValue := range vals {
					if _, ok := epCount[event][k][propertyValue]; !ok {
						delete(userPropertiesInfo.CategoricalPropertyKeyValues, k)
					}
				}
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

func FilterEventsInfo(userAndEventsInfo *P.UserAndEventsInfo, userProp, eventProp map[string]bool) *P.UserAndEventsInfo {

	userPropertiesInfo := userAndEventsInfo.UserPropertiesInfo
	eventPropertiesInfo := *userAndEventsInfo.EventPropertiesInfoMap

	//delete both categorical and numerical properties for users
	for propertyName, _ := range userPropertiesInfo.CategoricalPropertyKeyValues {

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
	for _, event := range eventPropertiesInfo {

		for k, _ := range event.CategoricalPropertyKeyValues {

			if !eventProp[k] {
				delete(event.CategoricalPropertyKeyValues, k)
			}
		}

		for k := range event.NumericPropertyKeys {

			if !eventProp[k] {
				delete(event.NumericPropertyKeys, k)
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

// GetTopStandardPatterns Get all events which starts with $ and not campaign Analytics Event
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

// GetTopSourcePatterns get all top patterns with $session and source set
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
		mineLog.Debug("Top Campaign Events ->", v.EventNames, " ", v.PerUserCount)
	}

	causalSourceEvents := GetTopSourcePatterns(allPatterns, srcNum)
	for _, v := range causalSourceEvents {
		mineLog.Debug("Top Source Events ->", v.EventNames, " ", v.PerUserCount)
	}

	causalMediumEvents := GetTopMediumPatterns(allPatterns, medNum)
	for _, v := range causalMediumEvents {
		mineLog.Debug("Top Medium Events ->", v.EventNames, " ", v.PerUserCount)
	}

	causalReferrerEvents := GetTopReferrerPatterns(allPatterns, medNum)
	for _, v := range causalReferrerEvents {
		mineLog.Debug("Top Referrer Events ->", v.EventNames, " ", v.PerUserCount)
	}

	causalAdgroupEvents := GetTopAdgroupPatterns(allPatterns, adgNum)
	for _, v := range causalAdgroupEvents {
		mineLog.Debug("Top Adgroup Events ->", v.EventNames, " ", v.PerUserCount)
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

	mineLog.Debugf("Total causal Patterns :%d", len(filteredPatterns))
	mineLog.Debugf("All causal Patterns :%v", filteredPatterns)

	return filteredPatterns, nil
}

// GenInterMediateCombinations generate cross combination patterns based on len two patt
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
		mineLog.Debug("Three len Campaign Events goal Event : ", goalName, len(targetPatt))

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
						mineLog.Debug("Three len Campaign Events : ", p.EventNames)
					}
				}
			}
		}
	}

	return allPatternsList, nil
}

// GetEventNamesFromFile read unique eventNames from Event file
func GetEventNamesFromFile_(scanner *bufio.Scanner, projectId int64) ([]string, error) {
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
		if !ok {
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
func GetEventNamesFromFile(scanner *bufio.Scanner, cloudManager *filestore.FileManager, projectId int64, modelId uint64, countsVersion int) ([]string, map[string]int, error) {
	logCtx := log.WithField("project_id", projectId)
	scanner.Split(bufio.ScanLines)
	var txtline string
	var dir, name string

	eventNames := make([]string, 0)
	eventsCount := make(map[string]int, 0)
	userPropCatgMap := make(map[string]string)
	eventPropCatgMap := make(map[string]string)

	eventUserPropMap := make(map[string]map[string]map[string]int)
	eventEventPropMap := make(map[string]map[string]map[string]int)

	blacklist := []string{}

	s := map[string]bool{}

	for scanner.Scan() {
		txtline = scanner.Text()

		var eventDetails P.CounterEventFormat
		if err := json.Unmarshal([]byte(txtline), &eventDetails); err != nil {
			log.WithFields(log.Fields{"line": txtline, "err": err}).Error("Read failed")
			return nil, nil, err
		}
		eventNameString := eventDetails.EventName
		eventsCount[eventNameString] += 1
		_, ok := s[eventNameString]
		if !ok {
			eventNames = append(eventNames, eventNameString)
			s[eventNameString] = true
		}

		for uKey, uVal := range eventDetails.UserProperties {
			if _, ok := userPropCatgMap[uKey]; !ok {
				propertyType := store.GetStore().GetPropertyTypeByKeyValue(projectId, eventNameString, uKey, uVal, true)
				userPropCatgMap[uKey] = propertyType
			}

			propType := userPropCatgMap[uKey]
			if propType == U.PropertyTypeCategorical {
				if _, ok := eventUserPropMap[eventNameString]; !ok {
					eventUserPropMap[eventNameString] = make(map[string]map[string]int)
				}
				pstring := U.GetPropertyValueAsString(uVal)
				if pstring == "" {
					pstring = "$none"
				}
				// propString := strings.Join([]string{uKey, pstring}, "::")
				if ok, _, _ := U.StringIn(blacklist, uKey); !ok {
					if _, ok := eventUserPropMap[eventNameString][uKey]; !ok {
						eventUserPropMap[eventNameString][uKey] = make(map[string]int)
					}
					eventUserPropMap[eventNameString][uKey][pstring] += 1
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
				if _, ok := eventEventPropMap[eventNameString]; !ok {
					eventEventPropMap[eventNameString] = make(map[string]map[string]int)
				}
				pstring := U.GetPropertyValueAsString(eVal)
				if pstring == "" {
					pstring = "$none"
				}
				// propString := strings.Join([]string{eKey, pstring}, "::")
				if ok, _, _ := U.StringIn(blacklist, eKey); !ok {
					if _, ok := eventEventPropMap[eventNameString][eKey]; !ok {
						eventEventPropMap[eventNameString][eKey] = make(map[string]int)
					}
					eventEventPropMap[eventNameString][eKey][pstring] += 1
				}
			}
		}
	}

	dir, name = (*cloudManager).GetModelUserPropertiesCategoricalFilePathAndName(projectId, modelId)
	CreateFileFromMap(dir, name, cloudManager, userPropCatgMap)

	dir, name = (*cloudManager).GetModelEventPropertiesCategoricalFilePathAndName(projectId, modelId)
	CreateFileFromMap(dir, name, cloudManager, eventPropCatgMap)

	err := scanner.Err()
	logCtx.Info("Extraced Unique EventNames from file")

	if err != nil {
		return nil, nil, err
	}

	for _, pMap := range eventUserPropMap {
		for _, vc := range pMap {
			U.FilterOnFrequencySpl(vc, max_prop_values_per_key)
		}
	}
	for _, pMap := range eventEventPropMap {
		for _, vc := range pMap {
			U.FilterOnFrequencySpl(vc, max_prop_values_per_key)
		}
	}

	dir, name = (*cloudManager).GetModelUserPropertiesFilePathAndName(projectId, modelId)
	CreateFileFromMap(dir, name, cloudManager, eventUserPropMap)

	_, name = (*cloudManager).GetModelEventPropertiesFilePathAndName(projectId, modelId)
	CreateFileFromMap(dir, name, cloudManager, eventEventPropMap)

	if countsVersion == 4 {
		U.FilterOnFrequency(eventsCount, 50)
	} else {

		U.FilterOnFrequency(eventsCount, 1)

	}
	return eventNames, eventsCount, nil
}

func CreateFileFromMap(fileDir, fileName string, cloudManager *filestore.FileManager, Map interface{}) error {
	if mapBytes, err := json.Marshal(Map); err != nil {
		mineLog.WithFields(log.Fields{"err": err}).Error("Failed to create ", fileName, " with error: ", err)
		return err
	} else {
		(*cloudManager).Create(fileDir, fileName, bytes.NewReader(mapBytes))
	}
	return nil
}

func removePropertiesTree(val string) bool {

	return U.IsDateTime(val)
}

func GetChunkMetaData(Patterns []*P.Pattern) []ChunkMetaData {
	filterEvents := getToBeFilteredKeysInMetaData()
	var metaData []ChunkMetaData
	for _, pattern := range Patterns {
		var metadataObj ChunkMetaData
		if pattern.PatternVersion < 2 { // histogram
			for _, eventName := range pattern.EventNames {
				for filter := range filterEvents {
					if !strings.HasPrefix(eventName, filter) {
						metadataObj.Events = append(metadataObj.Events, eventName)
					}
				}
			}
			// For Events
			EventPropertiesObj := Properties{
				CategoricalProperties: make(map[string][]string),
			}

			// numerical properties
			if pattern.PerUserEventNumericProperties != nil && pattern.PerUserEventNumericProperties.Template != nil {
				for _, eventNumericPropertyKey := range *pattern.PerUserEventNumericProperties.Template {
					if _, exists := EventNumericPropertyKeyMap[eventNumericPropertyKey.Name]; !exists {
						EventPropertiesObj.NumericalProperties = append(EventPropertiesObj.NumericalProperties, eventNumericPropertyKey.Name)
						EventNumericPropertyKeyMap[eventNumericPropertyKey.Name] = true
					}

				}
			}

			// categorical properties

			if pattern.PerUserEventCategoricalProperties != nil && pattern.PerUserEventCategoricalProperties.Template != nil {
				for idx, eventCategoricalPropertyKey := range *pattern.PerUserEventCategoricalProperties.Template {
					Key := eventCategoricalPropertyKey.Name
					eventCategoricalPropertyValues := make([]string, 0)
					for _, bin := range pattern.PerUserEventCategoricalProperties.Bins {
						valMap := make(map[string]bool)
						FreqMap := bin.FrequencyMaps[idx]
						for value := range FreqMap.Fmap {
							if _, exists := valMap[value]; !exists { // deduping values
								eventCategoricalPropertyValues = append(eventCategoricalPropertyValues, value)
								valMap[value] = true
							}
						}
					}
					if len(eventCategoricalPropertyValues) != 0 {
						EventPropertiesObj.CategoricalProperties[Key] = eventCategoricalPropertyValues
					}

				}
			}

			//For Users
			UserPropertiesObj := Properties{
				CategoricalProperties: make(map[string][]string),
			}
			// numerical properties
			if pattern.PerUserUserNumericProperties != nil && pattern.PerUserUserNumericProperties.Template != nil {
				for _, userNumericPropertyKey := range *pattern.PerUserUserNumericProperties.Template {
					if _, exists := UserNumericPropertyKeyMap[userNumericPropertyKey.Name]; !exists {
						UserPropertiesObj.NumericalProperties = append(UserPropertiesObj.NumericalProperties, userNumericPropertyKey.Name)
						UserNumericPropertyKeyMap[userNumericPropertyKey.Name] = true
					}

				}
			}

			// categorical properties
			if pattern.PerUserUserCategoricalProperties != nil && pattern.PerUserUserCategoricalProperties.Template != nil {
				for idx, userCategoricalPropertyKey := range *pattern.PerUserUserCategoricalProperties.Template {
					Key := userCategoricalPropertyKey.Name
					userCategoricalPropertyValues := make([]string, 0)
					for _, bin := range pattern.PerUserUserCategoricalProperties.Bins {
						valMap := make(map[string]bool)
						FreqMap := bin.FrequencyMaps[idx]
						for value := range FreqMap.Fmap {
							if _, exists := valMap[value]; !exists {
								userCategoricalPropertyValues = append(userCategoricalPropertyValues, value)
								valMap[value] = true
							}
						}
					}
					if len(userCategoricalPropertyValues) != 0 {
						UserPropertiesObj.CategoricalProperties[Key] = userCategoricalPropertyValues
					}

				}
			}
			metadataObj.EventProperties = EventPropertiesObj
			metadataObj.UserProperties = UserPropertiesObj

		} else if pattern.PatternVersion >= 2 { // fp tree and hmine
			for _, eventName := range pattern.EventNames {
				for filter := range filterEvents {
					if !strings.HasPrefix(eventName, filter) {
						metadataObj.Events = append(metadataObj.Events, eventName)
					}
				}
			}
			// only categorical properties
			// for Events Fp-tree
			var EventPropertiesObj Properties
			var EventCategoricalProperty = make(map[string][]string)
			if pattern.EventPropertiesPatterns != nil {
				for _, eventCategoricalProperty := range pattern.EventPropertiesPatterns {
					for _, item := range eventCategoricalProperty.Items {
						keyVal := strings.Split(item, "::") // splitting key and value
						if len(keyVal) == 2 {
							if vals, exists := EventCategoricalProperty[keyVal[0]]; exists {
								if !U.ContainsStringInArray(vals, keyVal[1]) {
									EventCategoricalProperty[keyVal[0]] = append(EventCategoricalProperty[keyVal[0]], keyVal[1])
								}
							} else {
								EventCategoricalProperty[keyVal[0]] = append(EventCategoricalProperty[keyVal[0]], keyVal[1])
							}
						}
					}
				}
			}

			EventPropertiesObj.CategoricalProperties = EventCategoricalProperty
			// only categorical properties
			// for Users Fp-tree
			var UserPropertiesObj Properties
			var UserCategoricalProperty = make(map[string][]string)
			if pattern.UserPropertiesPatterns != nil {
				for _, userCategoricalProperty := range pattern.UserPropertiesPatterns {
					for _, item := range userCategoricalProperty.Items {
						keyVal := strings.Split(item, "::")
						if len(keyVal) == 2 {
							if vals, exists := UserCategoricalProperty[keyVal[0]]; exists {
								if !U.ContainsStringInArray(vals, keyVal[1]) {
									UserCategoricalProperty[keyVal[0]] = append(UserCategoricalProperty[keyVal[0]], keyVal[1])
								}
							} else {
								UserCategoricalProperty[keyVal[0]] = append(UserCategoricalProperty[keyVal[0]], keyVal[1])
							}
						}
					}
				}
			}
			UserPropertiesObj.CategoricalProperties = UserCategoricalProperty
			metadataObj.EventProperties = EventPropertiesObj
			metadataObj.UserProperties = UserPropertiesObj
		}
		metaData = append(metaData, metadataObj)
	}
	return metaData
}
func getToBeFilteredKeysInMetaData() map[string]bool {
	keys := map[string]bool{
		"$session[":       true,
		"$AllActiveUsers": true,
	}
	return keys
}

func addIncludedEventsV2(jb *model.ExplainV2Query, ec map[string]int) {

	if jb.Query.StartEvent == "" {
		jb.Query.StartEvent = "$session"
	}

	if len(jb.Query.Rule.IncludedEvents) == 0 {
		log.Infof("Number of included events is nil , hence appending events ")
		for ename, _ := range ec {
			jb.Query.Rule.IncludedEvents = append(jb.Query.Rule.IncludedEvents, ename)
		}
	}
	if (strings.Compare(jb.Query.StartEvent, "$session") != 0) && (strings.Compare(jb.Query.EndEvent, "$session") != 0) {
		jb.Query.Rule.IncludedEvents = append(jb.Query.Rule.IncludedEvents, "$session")
	}

	log.Infof("total number of included events :%d", len(jb.Query.Rule.IncludedEvents))
	mineLog.Debugf("Recal top 50 properties :%v", jb)
}

func FilterEventsOnRule(ev P.CounterEventFormat, r model.ExplainV2Query, upCatg, epCatg map[string]string) bool {

	var frule model.FactorsGoalRule = r.Query

	ename := ev.EventName

	if strings.Compare(ename, frule.StartEvent) == 0 {
		if len(frule.Rule.StartEnEventFitler) > 0 {
			if !checkProperties("StartEventEnFilter", ev, frule, upCatg, epCatg) {
				return false
			}
		}

		if len(frule.Rule.StartEnUserFitler) > 0 {
			if !checkProperties("StartUserEnFilter", ev, frule, upCatg, epCatg) {
				return false
			}
		}

		return true

	}
	if strings.Compare(ename, frule.EndEvent) == 0 {

		if len(frule.Rule.EndEnEventFitler) > 0 {
			if !checkProperties("EndEventEnFilter", ev, frule, upCatg, epCatg) {
				return false
			}
		}

		if len(frule.Rule.EndEnUserFitler) > 0 {
			if !checkProperties("EndUserEnFilter", ev, frule, upCatg, epCatg) {
				return false
			}
		}

		return true
	}

	for _, in_ename := range frule.Rule.IncludedEvents {
		if strings.Compare(ename, in_ename) == 0 {
			if len(frule.Rule.IncludedUserProperties) > 0 {
				if !checkProperties("IncUserEnFilter", ev, frule, upCatg, epCatg) {
					return false
				}
			}

			if len(frule.Rule.IncludedEventProperties) > 0 {
				if !checkProperties("IncEventEnFilter", ev, frule, upCatg, epCatg) {
					return false
				}
			}
			return true
		}

	}

	return false
}

func checkProperties(filterString string, ev P.CounterEventFormat, ru model.FactorsGoalRule, upCatg, epCatg map[string]string) bool {

	upr := ev.UserProperties
	epr := ev.EventProperties

	if strings.Compare(filterString, "StartUserEnFilter") == 0 {
		st_us_ft := ru.Rule.StartEnUserFitler
		num_filters := len(st_us_ft)
		num_rules_matched := 0
		for ukey, uval := range upr {
			for _, fproperty := range st_us_ft {
				if strings.Compare(ukey, fproperty.Key) == 0 {
					vl := U.GetPropertyValueAsString(uval)
					vl = strings.ToLower(vl)
					prop_val := strings.ToLower(fproperty.Value)
					if fproperty.Operator == true {
						if strings.Compare(vl, prop_val) == 0 {
							num_rules_matched += 1
						}
					} else {
						if strings.Compare(vl, fproperty.Value) != 0 {
							num_rules_matched += 1
						}
					}
				}
			}
		}

		if num_rules_matched == num_filters {
			return true
		} else {
			return false
		}
	} else if strings.Compare(filterString, "StartEventEnFilter") == 0 {

		st_ev_ft := ru.Rule.StartEnEventFitler
		num_filters := len(st_ev_ft)
		num_rules_matched := 0
		for ukey, uval := range epr {
			for _, fproperty := range st_ev_ft {

				if ukey == fproperty.Key {
					vl := U.GetPropertyValueAsString(uval)
					vl = strings.ToLower(vl)
					prop_val := strings.ToLower(fproperty.Value)
					if fproperty.Operator == true {
						if strings.Compare(vl, prop_val) == 0 {
							num_rules_matched += 1
						}
					} else {
						if strings.Compare(vl, fproperty.Value) != 0 {
							num_rules_matched += 1
						}
					}
				}
			}
		}

		if num_rules_matched == num_filters {
			return true
		} else {
			return false
		}

	} else if strings.Compare(filterString, "EndUserEnFilter") == 0 {

		en_us_ft := ru.Rule.EndEnUserFitler
		num_filters := len(en_us_ft)
		num_rules_matched := 0
		for ukey, uval := range upr {
			for _, fproperty := range en_us_ft {

				if ukey == fproperty.Key {
					vl := U.GetPropertyValueAsString(uval)
					vl = strings.ToLower(vl)
					prop_val := strings.ToLower(fproperty.Value)
					if fproperty.Operator == true {
						if strings.Compare(vl, prop_val) == 0 {
							num_rules_matched += 1
						}
					} else {
						if strings.Compare(vl, fproperty.Value) != 0 {
							num_rules_matched += 1
						}
					}
				}
			}
		}

		if num_rules_matched == num_filters {
			return true
		} else {
			return false
		}
	} else if strings.Compare(filterString, "EndEventEnFilter") == 0 {

		en_ev_ft := ru.Rule.EndEnEventFitler
		num_filters := len(en_ev_ft)
		num_rules_matched := 0
		for ukey, uval := range upr {
			for _, fproperty := range en_ev_ft {
				if ukey == fproperty.Key {
					vl := U.GetPropertyValueAsString(uval)
					vl = strings.ToLower(vl)
					prop_val := strings.ToLower(fproperty.Value)
					if fproperty.Operator == true {
						if strings.Compare(vl, prop_val) == 0 {
							num_rules_matched += 1
						}
					} else {
						if strings.Compare(vl, fproperty.Value) != 0 {
							num_rules_matched += 1
						}
					}
				}
			}
		}

		if num_rules_matched == num_filters {
			return true
		} else {
			return false
		}
	} else if strings.Compare(filterString, "IncEventFilter") == 0 {

		in_ev_ft := ru.Rule.IncludedEventProperties
		num_filters := len(in_ev_ft)
		num_rules_matched := 0
		for ukey, _ := range epr {
			for _, fproperty := range in_ev_ft {
				if ukey == fproperty {
					num_rules_matched += 1
				}
			}
		}

		if num_rules_matched == num_filters {
			return true
		} else {
			return false
		}

	} else if filterString == "IncUserFilter" {

		in_us_ft := ru.Rule.IncludedUserProperties
		num_filters := len(in_us_ft)
		num_rules_matched := 0
		for ukey, _ := range epr {
			for _, fproperty := range in_us_ft {
				if ukey == fproperty {
					num_rules_matched += 1
				}
			}
		}

		if num_rules_matched == num_filters {
			return true
		} else {
			return false
		}
	}

	return false

}
