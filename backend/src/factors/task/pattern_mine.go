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
const max_SEGMENTS = 100000
const top_K = 5
const max_PATTERN_LENGTH = 4
const max_CHUNK_SIZE_IN_BYTES int64 = 250 * 1000 * 1000

var regex_NUM = regexp.MustCompile("[0-9]+")
var mineLog = taskLog.WithField("prefix", "Task#PatternMine")

func countPatternsWorker(filepath string,
	patterns []*P.Pattern, wg *sync.WaitGroup) {
	file, err := os.Open(filepath)
	if err != nil {
		mineLog.WithField("filePath", filepath).Error("Failure on count pattern workers.")
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
	mineLog.Info(fmt.Sprintf("Num patterns to count Range: %d - %d", 0, numPatterns-1))
	batchSize := int(math.Ceil(float64(numPatterns) / float64(numRoutines)))
	for i := 0; i < numRoutines; i++ {
		// Each worker gets a slice of patterns to count.
		low := int(math.Min(float64(batchSize*i), float64(numPatterns)))
		high := int(math.Min(float64(batchSize*(i+1)), float64(numPatterns)))
		mineLog.Info(fmt.Sprintf("Batch %d patterns to count range: %d:%d", i+1, low, high))
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
	patterns []*P.Pattern, userAndEventsInfo *P.UserAndEventsInfo) (map[string][]*P.Pattern, error) {
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
			mineLog.Error("Failure on generate segemented candidates.")
			return pSegments, err
		}
		candidateSegments[k] = cPatterns
	}
	return candidateSegments, nil
}

func genLenThreeSegmentedCandidates(lenTwoPatterns []*P.Pattern,
	userAndEventsInfo *P.UserAndEventsInfo) (map[string][]*P.Pattern, error) {
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
				return patterns[i].Count > patterns[j].Count
			})
	}
	for _, patterns := range endPatternsMap {
		sort.Slice(patterns,
			func(i, j int) bool {
				return patterns[i].Count > patterns[j].Count
			})
	}

	for _, p := range lenTwoPatterns {
		startPatterns, ok1 := startPatternsMap[p.EventNames[0]]
		endPatterns, ok2 := endPatternsMap[p.EventNames[1]]
		if startPatterns == nil || endPatterns == nil || !ok1 || !ok2 {
			continue
		}
		lenThreeCandidates, err := P.GenLenThreeCandidatePatterns(
			p, startPatterns, endPatterns, top_K, userAndEventsInfo)
		if err != nil {
			mineLog.WithError(err).Error("Failed on genLenThreeSegmentedCandidates.")
			return segmentedCandidates, err
		}
		segmentedCandidates[p.String()] = lenThreeCandidates
	}
	return segmentedCandidates, nil
}

func mineAndWriteLenOnePatterns(
	eventNames []M.EventName, filepath string,
	userAndEventsInfo *P.UserAndEventsInfo, numRoutines int,
	chunkDir string) ([]*P.Pattern, error) {
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
	if err := writePatternsAsChunks(filteredLenOnePatterns, chunkDir); err != nil {
		return []*P.Pattern{}, err
	}
	return filteredLenOnePatterns, nil
}

func mineAndWriteLenTwoPatterns(
	lenOnePatterns []*P.Pattern, filepath string,
	userAndEventsInfo *P.UserAndEventsInfo, numRoutines int,
	chunkDir string) ([]*P.Pattern, error) {
	// Each event combination is a segment in itself.
	lenTwoPatterns, _, err := P.GenCandidates(
		lenOnePatterns, max_SEGMENTS, userAndEventsInfo)
	if err != nil {
		return []*P.Pattern{}, err
	}
	countPatterns(filepath, lenTwoPatterns, numRoutines)
	filteredLenTwoPatterns := filterPatterns(lenTwoPatterns)
	if err := writePatternsAsChunks(filteredLenTwoPatterns, chunkDir); err != nil {
		return []*P.Pattern{}, err
	}
	return lenTwoPatterns, nil
}

func mineAndWritePatterns(projectId uint64, filepath string,
	userAndEventsInfo *P.UserAndEventsInfo, numRoutines int, chunkDir string) error {
	// Length One Patterns.
	eventNames, errCode := M.GetEventNames(projectId)
	if errCode != http.StatusFound {
		return fmt.Errorf("DB read of event names failed")
	}
	var filteredPatterns []*P.Pattern

	patternLen := 1
	filteredPatterns, err := mineAndWriteLenOnePatterns(
		eventNames, filepath, userAndEventsInfo, numRoutines, chunkDir)
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
		numRoutines, chunkDir)
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
		lenThreeSegmentedPatterns, err := genLenThreeSegmentedCandidates(
			filteredPatterns, userAndEventsInfo)
		if err != nil {
			return err
		}
		lenThreePatterns := []*P.Pattern{}
		for _, patterns := range lenThreeSegmentedPatterns {
			lenThreePatterns = append(lenThreePatterns, patterns...)
		}
		countPatterns(filepath, lenThreePatterns, numRoutines)
		filteredPatterns = filterPatterns(lenThreePatterns)
		if err := writePatternsAsChunks(filteredPatterns, chunkDir); err != nil {
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

		candidatePatternsMap, err = genSegmentedCandidates(
			filteredPatterns, userAndEventsInfo)
		if err != nil {
			return err
		}
		candidatePatterns = []*P.Pattern{}
		for _, patterns := range candidatePatternsMap {
			candidatePatterns = append(candidatePatterns, patterns...)
		}
		countPatterns(filepath, candidatePatterns, numRoutines)
		filteredPatterns = filterPatterns(candidatePatterns)
		if len(filteredPatterns) > 0 {
			if err := writePatternsAsChunks(filteredPatterns, chunkDir); err != nil {
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
		b, err := json.Marshal(pattern)
		if err != nil {
			mineLog.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal pattern.")
			return err
		}
		pString := string(b)
		pString = pString + "\n"
		pBytes := []byte(pString)
		pBytesLen := int64(len(pBytes))

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

// PatternMine Mine TOP_K Frequent patterns for every event combination (segment) at every iteration.
func PatternMine(db *gorm.DB, etcdClient *serviceEtcd.EtcdClient, cloudManager *filestore.FileManager,
	diskManager *serviceDisk.DiskDriver, bucketName string, numRoutines int, projectId uint64,
	modelId uint64, modelType string, startTime int64, endTime int64) (string, error) {

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
		return "", err
	}
	err = diskManager.Create(efTmpPath, efTmpName, eReader)
	if err != nil {
		mineLog.WithFields(log.Fields{"err": err, "eventFilePath": efTmpPath,
			"eventFileName": efTmpName}).Error("Failed to create event file on disk.")
		return "", err
	}
	tmpEventsFilepath := efTmpPath + efTmpName
	mineLog.Info("Successfuly downloaded events file from cloud.")

	// builld user and event properites info
	mineLog.WithField("tmpEventsFilePath",
		tmpEventsFilepath).Info("Building user and event properties info and writing it to file.")
	userAndEventsInfo, err := buildPropertiesInfoFromInput(projectId, tmpEventsFilepath)
	if err != nil {
		mineLog.WithFields(log.Fields{"err": err}).Error("Failed to build user and event Info.")
		return "", err
	}
	userAndEventsInfoBytes, err := json.Marshal(userAndEventsInfo)
	if err != nil {
		mineLog.WithFields(log.Fields{"err": err}).Error("Failed to unmarshal events Info.")
		return "", err
	}
	err = writeEventInfoFile(projectId, modelId, bytes.NewReader(userAndEventsInfoBytes), (*cloudManager))
	if err != nil {
		mineLog.WithFields(log.Fields{"err": err}).Error("Failed to write events Info.")
		return "", err
	}
	mineLog.Info("Successfully Built user and event properties info and written it to file.")

	// build histogram of all user properties.
	mineLog.WithField("tmpEventsFilePath", tmpEventsFilepath).Info("Building all user properties histogram.")
	allActiveUsersPattern, err := P.NewPattern([]string{U.SEN_ALL_ACTIVE_USERS}, userAndEventsInfo)
	if err != nil {
		mineLog.WithFields(log.Fields{"err": err}).Error("Failed to build pattern with histogram of all active user properties.")
		return "", err
	}
	if err := computeAllUserPropertiesHistogram(tmpEventsFilepath, allActiveUsersPattern); err != nil {
		mineLog.WithFields(log.Fields{"err": err}).Error("Failed to compute user properties.")
		return "", err
	}
	tmpChunksDir := diskManager.GetPatternChunksDir(projectId, modelId)
	if err := serviceDisk.MkdirAll(tmpChunksDir); err != nil {
		mineLog.WithFields(log.Fields{"chunkDir": tmpChunksDir, "error": err}).Error("Unable to create chunks directory.")
		return "", err
	}
	if err := writePatternsAsChunks([]*P.Pattern{allActiveUsersPattern}, tmpChunksDir); err != nil {
		mineLog.WithFields(log.Fields{"err": err}).Error("Failed to write user properties.")
		return "", err
	}
	mineLog.Info("Successfully built all user properties histogram.")

	// mine and write patterns as chunks
	mineLog.WithFields(log.Fields{"projectId": projectId, "tmpEventsFilepath": tmpEventsFilepath,
		"tmpChunksDir": tmpChunksDir, "routines": numRoutines}).Info("Mining patterns and writing it as chunks.")
	err = mineAndWritePatterns(projectId, tmpEventsFilepath,
		userAndEventsInfo, numRoutines, tmpChunksDir)
	if err != nil {
		mineLog.WithFields(log.Fields{"err": err}).Error("Failed to mine patterns.")
		return "", err
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
		return "", err
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
		return "", err
	}
	err = etcdClient.SetProjectVersion(newVersionId)
	if err != nil {
		mineLog.WithFields(log.Fields{"err": err}).Error("Failed to write new version id to etcd.")
		return "", err
	}
	mineLog.WithField("newVersionId", newVersionId).Info("Successfully mined patterns, updated metadata and notified new version id.")

	return newVersionId, nil
}
