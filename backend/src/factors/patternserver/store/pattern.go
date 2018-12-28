package store

import (
	"bufio"
	"bytes"
	"encoding/json"
	"factors/filestore"
	"factors/pattern"
	"fmt"
	"os"

	cache "github.com/hashicorp/golang-lru"
	log "github.com/sirupsen/logrus"
)

const (
	IdSeparator = ":"
)

type PatternStore struct {
	cloudFileManger filestore.FileManager
	diskFileManger  filestore.FileManager

	modelChunkCache     *cache.Cache
	modelEventInfoCache *cache.Cache
}

func New(chunkCacheSize, eventInfoCacheSize int, diskManager, cloudManager filestore.FileManager) (*PatternStore, error) {
	modelChunkCache, err := cache.New(chunkCacheSize)
	if err != nil {
		return nil, err
	}

	modelEventInfoCache, err := cache.New(eventInfoCacheSize)
	if err != nil {
		return nil, err
	}

	return &PatternStore{
		modelChunkCache:     modelChunkCache,
		modelEventInfoCache: modelEventInfoCache,
		diskFileManger:      diskManager,
		cloudFileManger:     cloudManager,
	}, nil
}

func GetModelKey(projectId, modelId uint64) string {
	return fmt.Sprintf("%d%s%d", projectId, IdSeparator, modelId)
}

func GetChunkKey(projectId, modelId uint64, chunkId string) string {
	return fmt.Sprintf("%s%s%s", GetModelKey(projectId, modelId), IdSeparator, chunkId)
}

func getModelEventInfoCacheKey(projectId, modelId uint64) string {
	return fmt.Sprintf("%s%s%s", "model_event_info", IdSeparator, GetModelKey(projectId, modelId))
}

func (ps *PatternStore) putModelEventInfoInCache(projectId, modelId uint64, eventInfo pattern.UserAndEventsInfo) {

	log.WithFields(log.Fields{
		"pid": projectId,
		"mid": modelId,
	}).Debugln("[PatternStore] putModelEventInfoInCache")

	modelKey := getModelEventInfoCacheKey(projectId, modelId)
	// return evicted boolean
	// check if it can be used
	ps.modelEventInfoCache.Add(modelKey, eventInfo)
}

func (ps *PatternStore) getModelEventInfoFromCache(projectId, modelId uint64) (pattern.UserAndEventsInfo, bool) {

	log.WithFields(log.Fields{
		"pid": projectId,
		"mid": modelId,
	}).Debugln("[PatternStore] getModelEventInfoInCache")

	modelKey := getModelEventInfoCacheKey(projectId, modelId)
	modelEventInfoIface, ok := ps.modelEventInfoCache.Get(modelKey)
	if !ok {
		return pattern.UserAndEventsInfo{}, ok
	}
	modelEventInfo, ok := modelEventInfoIface.(pattern.UserAndEventsInfo)
	return modelEventInfo, ok
}

func (ps *PatternStore) PutModelEventInfoInDisk(projectId, modelId uint64, eventInfo pattern.UserAndEventsInfo) error {

	log.WithFields(log.Fields{
		"pid": projectId,
		"mid": modelId,
	}).Debugln("[PatternStore] putModelEventInfoInDisk")

	reader, err := CreateReaderFromEventInfo(eventInfo)
	if err != nil {
		return err
	}

	path, fName := ps.diskFileManger.GetModelEventInfoFilePathAndName(projectId, modelId)
	err = ps.diskFileManger.Create(path, fName, reader)

	return err
}

func (ps *PatternStore) getModelEventInfoFromDisk(projectId, modelId uint64) (pattern.UserAndEventsInfo, error) {

	log.WithFields(log.Fields{
		"pid": projectId,
		"mid": modelId,
	}).Debugln("[PatternStore] getModelEventInfoFromDisk")

	path, fName := ps.diskFileManger.GetModelEventInfoFilePathAndName(projectId, modelId)
	modelEventInfoFile, err := ps.diskFileManger.Get(path, fName)
	if err != nil {
		return pattern.UserAndEventsInfo{}, err
	}
	defer modelEventInfoFile.Close()

	scanner := bufio.NewScanner(modelEventInfoFile)

	patternEventInfo, err := CreatePatternEventInfoFromScanner(scanner)

	return patternEventInfo, err
}

func (ps *PatternStore) getModelEventInfoFromCloud(projectId, modelId uint64) (pattern.UserAndEventsInfo, error) {
	log.WithFields(log.Fields{
		"pid": projectId,
		"mid": modelId,
	}).Debugln("[PatternStore] getModelEventInfoFromCloud")

	path, fName := ps.cloudFileManger.GetModelEventInfoFilePathAndName(projectId, modelId)
	modelEventInfoFile, err := ps.cloudFileManger.Get(path, fName)
	if err != nil {
		return pattern.UserAndEventsInfo{}, err
	}
	defer modelEventInfoFile.Close()

	scanner := bufio.NewScanner(modelEventInfoFile)
	patternEventInfo, err := CreatePatternEventInfoFromScanner(scanner)

	return patternEventInfo, err
}

func (ps *PatternStore) PutModelEventInfoInCloud(projectId, modelId uint64, eventInfo pattern.UserAndEventsInfo) error {
	path, fName := ps.cloudFileManger.GetModelEventInfoFilePathAndName(projectId, modelId)

	reader, err := CreateReaderFromEventInfo(eventInfo)
	if err != nil {
		return err
	}

	err = ps.cloudFileManger.Create(path, fName, reader)

	return err
}

func (ps *PatternStore) GetModelEventInfo(projectId, modelId uint64) (pattern.UserAndEventsInfo, error) {

	logCtx := log.WithFields(log.Fields{
		"pid": projectId,
		"mid": modelId,
	})

	logCtx.Debugln("[PatternStore] GetModelEventInfo")

	ei, foundInCache := ps.getModelEventInfoFromCache(projectId, modelId)
	if foundInCache {
		return ei, nil
	}

	writeToCache := !foundInCache
	writeToDisk := false

	patternEventInfo, err := ps.getModelEventInfoFromDisk(projectId, modelId)
	if err != nil {
		if os.IsNotExist(err) {
			writeToDisk = true
			patternEventInfo, err = ps.getModelEventInfoFromCloud(projectId, modelId)
			if err != nil {
				return pattern.UserAndEventsInfo{}, err
			}
		} else {
			return pattern.UserAndEventsInfo{}, err
		}
	}

	// This can be done in a separate go routine
	if writeToCache {
		ps.putModelEventInfoInCache(projectId, modelId, patternEventInfo)
	}

	if writeToDisk {
		err = ps.PutModelEventInfoInDisk(projectId, modelId, patternEventInfo)
		if err != nil {
			logCtx.WithError(err).Error("Failed to write eventInfoMap to disk")
		}
	}

	return patternEventInfo, err
}

func CreateReaderFromEventInfo(eventInfo pattern.UserAndEventsInfo) (*bytes.Reader, error) {
	eventInfoBytes, err := json.Marshal(eventInfo)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Fatal("Failed to marshal events Info.")
		return nil, err
	}
	reader := bytes.NewReader(eventInfoBytes)
	return reader, nil
}

func CreatePatternEventInfoFromScanner(scanner *bufio.Scanner) (pattern.UserAndEventsInfo, error) {
	patternEventInfo := pattern.UserAndEventsInfo{}
	// Adjust scanner buffer capacity to 10MB per line.
	const maxCapacity = 10 * 1024 * 1024
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	for scanner.Scan() {
		line := scanner.Bytes()
		if err := json.Unmarshal(line, &patternEventInfo); err != nil {
			return pattern.UserAndEventsInfo{}, err
		}
	}
	err := scanner.Err()

	return patternEventInfo, err
}

func CreatePatternsFromScanner(scanner *bufio.Scanner) ([]*pattern.Pattern, error) {
	patterns := make([]*pattern.Pattern, 0, 0)
	for scanner.Scan() {
		line := scanner.Bytes()
		var p pattern.Pattern
		if err := json.Unmarshal(line, &p); err != nil {
			return patterns, err
		}
		patterns = append(patterns, &p)
	}
	err := scanner.Err()

	return patterns, err
}

func CreateScannerFromFile(file *os.File) *bufio.Scanner {
	scanner := bufio.NewScanner(file)
	// Adjust scanner buffer capacity to 10MB per line.
	const maxCapacity = 10 * 1024 * 1024
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)
	return scanner
}
