package store

import (
	"bufio"
	"bytes"
	"encoding/json"
	"factors/filestore"
	"factors/pattern"
	"fmt"
	"io"
	"os"
	"runtime"
	"runtime/debug"
	"time"

	U "factors/util"

	cache "github.com/hashicorp/golang-lru"
	log "github.com/sirupsen/logrus"
)

const (
	IdSeparator = ":"
)

type PatternStore struct {
	cloudFileManger     filestore.FileManager
	newCloudFileManager filestore.FileManager
	diskFileManger      filestore.FileManager
	projectIdsV2        []int64

	modelChunkCache     *cache.Cache
	modelEventInfoCache *cache.Cache
}

type PatternWithMeta struct {
	PatternEvents []string        `json:"pe"`
	RawPattern    json.RawMessage `json:"rp"`
}

func New(chunkCacheSize, eventInfoCacheSize int, diskManager, cloudManager, newCloudManager filestore.FileManager, projectIdsV2 []int64) (*PatternStore, error) {
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
		newCloudFileManager: newCloudManager,
		projectIdsV2:        projectIdsV2,
	}, nil
}

func GetModelKey(projectId int64, modelId uint64) string {
	return fmt.Sprintf("%d%s%d", projectId, IdSeparator, modelId)
}

func GetChunkKey(projectId int64, modelId uint64, chunkId string) string {
	return fmt.Sprintf("%s%s%s", GetModelKey(projectId, modelId), IdSeparator, chunkId)
}

func getModelEventInfoCacheKey(projectId int64, modelId uint64) string {
	return fmt.Sprintf("%s%s%s", "model_event_info", IdSeparator, GetModelKey(projectId, modelId))
}

func getModelChunkCacheKey(projectId int64, modelId uint64, chunkId string) string {
	return fmt.Sprintf("%s%s%s", "chunk_", IdSeparator, GetChunkKey(projectId, modelId, chunkId))
}

func (ps *PatternStore) GetCloudManager(projectId int64) filestore.FileManager {
	if U.ContainsInt64InArray(ps.projectIdsV2, projectId) {
		return ps.newCloudFileManager
	}
	return ps.cloudFileManger
}

func (ps *PatternStore) putModelEventInfoInCache(projectId int64, modelId uint64, eventInfo pattern.UserAndEventsInfo) {

	logCtx := log.WithFields(log.Fields{
		"pid": projectId,
		"mid": modelId,
	})
	logCtx.Debugln("[PatternStore] putModelEventInfoInCache")

	modelKey := getModelEventInfoCacheKey(projectId, modelId)
	evict := ps.modelEventInfoCache.Add(modelKey, eventInfo)
	if evict {
		start := time.Now()
		runtime.GC()
		// Releasing memory back to the OS
		// https://stackoverflow.com/questions/37382600/cannot-free-memory-once-occupied-by-bytes-buffer
		debug.FreeOSMemory()
		logCtx.WithFields(log.Fields{
			"Duration (nanoseconds)": time.Since(start).Nanoseconds(),
		}).Info("GC Finished [PatternStore] putModelEventInfoInCache")
	}
}

func (ps *PatternStore) getModelEventInfoFromCache(projectId int64, modelId uint64) (pattern.UserAndEventsInfo, bool) {

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

func putModelEventInfoToFileManager(fm filestore.FileManager, projectId int64, modelId uint64, eventInfo pattern.UserAndEventsInfo) error {
	path, fName := fm.GetModelEventInfoFilePathAndName(projectId, modelId)

	reader, err := CreateReaderFromEventInfo(eventInfo)
	if err != nil {
		return err
	}
	err = fm.Create(path, fName, reader)
	return err
}

func (ps *PatternStore) PutModelEventInfoInDisk(projectId int64, modelId uint64, eventInfo pattern.UserAndEventsInfo) error {

	log.WithFields(log.Fields{
		"pid": projectId,
		"mid": modelId,
	}).Debugln("[PatternStore] putModelEventInfoInDisk")

	return putModelEventInfoToFileManager(ps.diskFileManger, projectId, modelId, eventInfo)
}

func getModelEventInfoFromFileManager(fm filestore.FileManager, projectId int64, modelId uint64) (pattern.UserAndEventsInfo, error) {
	path, fName := fm.GetModelEventInfoFilePathAndName(projectId, modelId)
	modelEventInfoFile, err := fm.Get(path, fName)
	if err != nil {
		return pattern.UserAndEventsInfo{}, err
	}
	defer modelEventInfoFile.Close()

	scanner := bufio.NewScanner(modelEventInfoFile)

	patternEventInfo, err := CreatePatternEventInfoFromScanner(scanner)

	return patternEventInfo, err
}

func (ps *PatternStore) getModelEventInfoFromDisk(projectId int64, modelId uint64) (pattern.UserAndEventsInfo, error) {

	log.WithFields(log.Fields{
		"pid": projectId,
		"mid": modelId,
	}).Debugln("[PatternStore] getModelEventInfoFromDisk")

	return getModelEventInfoFromFileManager(ps.diskFileManger, projectId, modelId)
}

func (ps *PatternStore) getModelEventInfoFromCloud(projectId int64, modelId uint64) (pattern.UserAndEventsInfo, error) {
	log.WithFields(log.Fields{
		"pid": projectId,
		"mid": modelId,
	}).Debugln("[PatternStore] getModelEventInfoFromCloud")

	return getModelEventInfoFromFileManager(ps.GetCloudManager(projectId), projectId, modelId)
}

func (ps *PatternStore) PutModelEventInfoInCloud(projectId int64, modelId uint64, eventInfo pattern.UserAndEventsInfo) error {
	return putModelEventInfoToFileManager(ps.GetCloudManager(projectId), projectId, modelId, eventInfo)
}

func (ps *PatternStore) GetModelEventInfo(projectId int64, modelId uint64) (pattern.UserAndEventsInfo, error) {

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

func getPatternsFromFileManager(fm filestore.FileManager, projectId int64, modelId uint64, chunkId string) ([]*PatternWithMeta, error) {
	path, fName := fm.GetPatternChunkFilePathAndName(projectId, modelId, chunkId)
	patternsReader, err := fm.Get(path, fName)
	if err != nil {
		return nil, err
	}
	defer patternsReader.Close()
	scanner := CreateScannerFromReader(patternsReader)
	patterns, err := CreatePatternsWithMetaFromScanner(scanner)
	return patterns, err
}

func (ps *PatternStore) getPatternsWithMetaFromCloud(projectId int64, modelId uint64, chunkId string) ([]*PatternWithMeta, error) {
	log.WithFields(log.Fields{
		"pid": projectId,
		"mid": modelId,
		"cid": chunkId,
	}).Debugln("[PatternStore] getPatternsFromCloud")

	return getPatternsFromFileManager(ps.GetCloudManager(projectId), projectId, modelId, chunkId)
}

func (ps *PatternStore) PutPatternsWithMetaInCloud(projectId int64, modelId uint64, chunkId string, patterns []*PatternWithMeta) error {
	log.WithFields(log.Fields{
		"pid": projectId,
		"mid": modelId,
		"cid": chunkId,
	}).Debugln("[PatternStore] putPatternsInCloud")

	reader, err := CreateReaderFromPatternsWithMeta(patterns)
	if err != nil {
		return err
	}

	cloudManager := ps.GetCloudManager(projectId)

	path, fName := cloudManager.GetPatternChunkFilePathAndName(projectId, modelId, chunkId)

	err = cloudManager.Create(path, fName, reader)

	return err
}

func (ps *PatternStore) PutPatternsWithMetaInDisk(projectId int64, modelId uint64, chunkId string, patterns []*PatternWithMeta) error {
	log.WithFields(log.Fields{
		"pid": projectId,
		"mid": modelId,
		"cid": chunkId,
	}).Debugln("[PatternStore] putPatternsInDisk")

	reader, err := CreateReaderFromPatternsWithMeta(patterns)
	if err != nil {
		return err
	}

	path, fName := ps.diskFileManger.GetPatternChunkFilePathAndName(projectId, modelId, chunkId)

	err = ps.diskFileManger.Create(path, fName, reader)

	return err
}

func (ps *PatternStore) getPatternsWithMetaFromDisk(projectId int64, modelId uint64, chunkId string) ([]*PatternWithMeta, error) {
	log.WithFields(log.Fields{
		"pid": projectId,
		"mid": modelId,
		"cid": chunkId,
	}).Debugln("[PatternStore] getPatternsFromDisk")

	return getPatternsFromFileManager(ps.diskFileManger, projectId, modelId, chunkId)
}

func (ps *PatternStore) putPatternsWithMetaInCache(projectId int64, modelId uint64, chunkId string, patterns []*PatternWithMeta) {
	logCtx := log.WithFields(log.Fields{
		"pid": projectId,
		"mid": modelId,
		"cid": chunkId,
	})
	logCtx.Debugln("[PatternStore] putPatternsInCache")

	chunkKey := getModelChunkCacheKey(projectId, modelId, chunkId)
	evict := ps.modelChunkCache.Add(chunkKey, patterns)
	if evict {
		start := time.Now()
		runtime.GC()
		// Releasing memory back to the OS
		// https://stackoverflow.com/questions/37382600/cannot-free-memory-once-occupied-by-bytes-buffer
		debug.FreeOSMemory()
		logCtx.WithFields(log.Fields{
			"Duration (nanoseconds)": time.Since(start).Nanoseconds(),
		}).Info("GC Finished [PatternStore] putPatternsInCache")
	}
}

func (ps *PatternStore) getPatternsWithMetaFromCache(projectId int64, modelId uint64, chunkId string) ([]*PatternWithMeta, bool) {
	log.WithFields(log.Fields{
		"pid": projectId,
		"mid": modelId,
		"cid": chunkId,
	}).Debugln("[PatternStore] getPatternsFromCache")

	chunkKey := getModelChunkCacheKey(projectId, modelId, chunkId)
	pwmIface, ok := ps.modelChunkCache.Get(chunkKey)
	if !ok {
		return nil, false
	}
	pwm, ok := pwmIface.([]*PatternWithMeta)
	return pwm, ok
}

func (ps *PatternStore) GetPatternsWithMeta(projectId int64, modelId uint64, chunkId string) ([]*PatternWithMeta, error) {
	logCtx := log.WithFields(log.Fields{
		"pid": projectId,
		"mid": modelId,
		"cid": chunkId,
	})

	logCtx.Debugln("[PatternStore] GetPatterns")

	patternsWithMeta, foundInCache := ps.getPatternsWithMetaFromCache(projectId, modelId, chunkId)
	if foundInCache {
		return patternsWithMeta, nil
	}

	writeToCache := !foundInCache
	writeToDisk := false

	patternsWithMeta, err := ps.getPatternsWithMetaFromDisk(projectId, modelId, chunkId)
	if err != nil {
		if os.IsNotExist(err) {
			writeToDisk = true
			patternsWithMeta, err = ps.getPatternsWithMetaFromCloud(projectId, modelId, chunkId)
			if err != nil {
				return []*PatternWithMeta{}, err
			}
		} else {
			return []*PatternWithMeta{}, err
		}
	}

	// This can be done in a separate go routine
	if writeToCache {
		ps.putPatternsWithMetaInCache(projectId, modelId, chunkId, patternsWithMeta)
	}

	if writeToDisk {
		err = ps.PutPatternsWithMetaInDisk(projectId, modelId, chunkId, patternsWithMeta)
	}

	return patternsWithMeta, nil
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
	const initBufSize = 100 * 1024 // 100KB
	buf := make([]byte, initBufSize)

	// Adjust scanner buffer capacity upto 250MB per line.
	const maxCapacity = 250 * 1024 * 1024
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

func CreateReaderFromPatternsWithMeta(patterns []*PatternWithMeta) (*bytes.Reader, error) {
	buf := new(bytes.Buffer)
	for _, pattern := range patterns {
		pwmBytes, err := json.Marshal(pattern)
		if err != nil {
			return nil, err
		}

		str := string(pwmBytes)
		_, err = buf.WriteString(fmt.Sprintf("%s\n", str))
		if err != nil {
			return nil, err
		}
	}
	return bytes.NewReader(buf.Bytes()), nil
}

func CreatePatternsWithMetaFromScanner(scanner *bufio.Scanner) ([]*PatternWithMeta, error) {
	patternsWithMeta := make([]*PatternWithMeta, 0, 0)
	for scanner.Scan() {
		line := scanner.Bytes()
		var pwm PatternWithMeta
		if err := json.Unmarshal(line, &pwm); err != nil {
			return patternsWithMeta, err
		}
		patternsWithMeta = append(patternsWithMeta, &pwm)
	}
	err := scanner.Err()

	return patternsWithMeta, err
}

func CreateScannerFromReader(r io.Reader) *bufio.Scanner {
	scanner := bufio.NewScanner(r)
	// Adjust scanner buffer capacity to MAX_PATTERN_BYTES per line.
	buf := make([]byte, pattern.MAX_PATTERN_BYTES)
	scanner.Buffer(buf, pattern.MAX_PATTERN_BYTES)
	return scanner
}

func GetAllRawPatterns(patternsWithMeta []*PatternWithMeta) []*json.RawMessage {
	rawPatterns := make([]*json.RawMessage, 0, 0)

	for _, pwm := range patternsWithMeta {
		rawPatterns = append(rawPatterns, &pwm.RawPattern)
	}

	return rawPatterns
}
