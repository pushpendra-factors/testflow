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
	"time"

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

func getModelChunkCacheKey(projectId, modelId uint64, chunkId string) string {
	return fmt.Sprintf("%s%s%s", "chunk_", IdSeparator, GetChunkKey(projectId, modelId, chunkId))
}

func (ps *PatternStore) putModelEventInfoInCache(projectId, modelId uint64, eventInfo pattern.UserAndEventsInfo) {

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
		logCtx.WithFields(log.Fields{
			"Duration (nanoseconds)": time.Since(start).Nanoseconds(),
		}).Info("GC Finished [PatternStore] putModelEventInfoInCache")
	}
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

func putModelEventInfoToFileManager(fm filestore.FileManager, projectId, modelId uint64, eventInfo pattern.UserAndEventsInfo) error {
	path, fName := fm.GetModelEventInfoFilePathAndName(projectId, modelId)

	reader, err := CreateReaderFromEventInfo(eventInfo)
	if err != nil {
		return err
	}
	err = fm.Create(path, fName, reader)
	return err
}

func (ps *PatternStore) PutModelEventInfoInDisk(projectId, modelId uint64, eventInfo pattern.UserAndEventsInfo) error {

	log.WithFields(log.Fields{
		"pid": projectId,
		"mid": modelId,
	}).Debugln("[PatternStore] putModelEventInfoInDisk")

	return putModelEventInfoToFileManager(ps.diskFileManger, projectId, modelId, eventInfo)
}

func getModelEventInfoFromFileManager(fm filestore.FileManager, projectId, modelId uint64) (pattern.UserAndEventsInfo, error) {
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

func (ps *PatternStore) getModelEventInfoFromDisk(projectId, modelId uint64) (pattern.UserAndEventsInfo, error) {

	log.WithFields(log.Fields{
		"pid": projectId,
		"mid": modelId,
	}).Debugln("[PatternStore] getModelEventInfoFromDisk")

	return getModelEventInfoFromFileManager(ps.diskFileManger, projectId, modelId)
}

func (ps *PatternStore) getModelEventInfoFromCloud(projectId, modelId uint64) (pattern.UserAndEventsInfo, error) {
	log.WithFields(log.Fields{
		"pid": projectId,
		"mid": modelId,
	}).Debugln("[PatternStore] getModelEventInfoFromCloud")

	return getModelEventInfoFromFileManager(ps.cloudFileManger, projectId, modelId)
}

func (ps *PatternStore) PutModelEventInfoInCloud(projectId, modelId uint64, eventInfo pattern.UserAndEventsInfo) error {
	return putModelEventInfoToFileManager(ps.cloudFileManger, projectId, modelId, eventInfo)
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

func getPatternsFromFileManager(fm filestore.FileManager, projectId, modelId uint64, chunkId string) ([]*pattern.Pattern, error) {
	path, fName := fm.GetPatternChunkFilePathAndName(projectId, modelId, chunkId)
	patternsReader, err := fm.Get(path, fName)
	if err != nil {
		return []*pattern.Pattern{}, err
	}
	defer patternsReader.Close()
	scanner := CreateScannerFromReader(patternsReader)
	patterns, err := CreatePatternsFromScanner(scanner)
	return patterns, err
}

func (ps *PatternStore) getPatternsFromCloud(projectId, modelId uint64, chunkId string) ([]*pattern.Pattern, error) {
	log.WithFields(log.Fields{
		"pid": projectId,
		"mid": modelId,
		"cid": chunkId,
	}).Debugln("[PatternStore] getPatternsFromCloud")

	return getPatternsFromFileManager(ps.cloudFileManger, projectId, modelId, chunkId)
}

func (ps *PatternStore) PutPatternsInCloud(projectId, modelId uint64, chunkId string, patterns []*pattern.Pattern) error {
	log.WithFields(log.Fields{
		"pid": projectId,
		"mid": modelId,
		"cid": chunkId,
	}).Debugln("[PatternStore] putPatternsInCloud")

	reader, err := CreateReaderFromPatterns(patterns)
	if err != nil {
		return err
	}

	path, fName := ps.cloudFileManger.GetPatternChunkFilePathAndName(projectId, modelId, chunkId)

	err = ps.cloudFileManger.Create(path, fName, reader)

	return err
}

func (ps *PatternStore) PutPatternsInDisk(projectId, modelId uint64, chunkId string, patterns []*pattern.Pattern) error {
	log.WithFields(log.Fields{
		"pid": projectId,
		"mid": modelId,
		"cid": chunkId,
	}).Debugln("[PatternStore] putPatternsInDisk")

	reader, err := CreateReaderFromPatterns(patterns)
	if err != nil {
		return err
	}

	path, fName := ps.diskFileManger.GetPatternChunkFilePathAndName(projectId, modelId, chunkId)

	err = ps.diskFileManger.Create(path, fName, reader)

	return err
}

func (ps *PatternStore) getPatternsFromDisk(projectId, modelId uint64, chunkId string) ([]*pattern.Pattern, error) {
	log.WithFields(log.Fields{
		"pid": projectId,
		"mid": modelId,
		"cid": chunkId,
	}).Debugln("[PatternStore] getPatternsFromDisk")

	return getPatternsFromFileManager(ps.diskFileManger, projectId, modelId, chunkId)
}

func (ps *PatternStore) putPatternsInCache(projectId, modelId uint64, chunkId string, patterns []*pattern.Pattern) {
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
		logCtx.WithFields(log.Fields{
			"Duration (nanoseconds)": time.Since(start).Nanoseconds(),
		}).Info("GC Finished [PatternStore] putPatternsInCache")
	}
}

func (ps *PatternStore) getPatternsFromCache(projectId, modelId uint64, chunkId string) ([]*pattern.Pattern, bool) {
	log.WithFields(log.Fields{
		"pid": projectId,
		"mid": modelId,
		"cid": chunkId,
	}).Debugln("[PatternStore] getPatternsFromCache")

	chunkKey := getModelChunkCacheKey(projectId, modelId, chunkId)
	patternsIface, ok := ps.modelChunkCache.Get(chunkKey)
	if !ok {
		return []*pattern.Pattern{}, ok
	}
	patterns, ok := patternsIface.([]*pattern.Pattern)
	return patterns, ok
}

func (ps *PatternStore) GetPatterns(projectId, modelId uint64, chunkId string) ([]*pattern.Pattern, error) {
	logCtx := log.WithFields(log.Fields{
		"pid": projectId,
		"mid": modelId,
		"cid": chunkId,
	})

	logCtx.Debugln("[PatternStore] GetPatterns")

	patterns, foundInCache := ps.getPatternsFromCache(projectId, modelId, chunkId)
	if foundInCache {
		return patterns, nil
	}

	writeToCache := !foundInCache
	writeToDisk := false

	patterns, err := ps.getPatternsFromDisk(projectId, modelId, chunkId)
	if err != nil {
		if os.IsNotExist(err) {
			writeToDisk = true
			patterns, err = ps.getPatternsFromCloud(projectId, modelId, chunkId)
			if err != nil {
				return []*pattern.Pattern{}, err
			}
		} else {
			return []*pattern.Pattern{}, err
		}
	}

	// This can be done in a separate go routine
	if writeToCache {
		ps.putPatternsInCache(projectId, modelId, chunkId, patterns)
	}

	if writeToDisk {
		err = ps.PutPatternsInDisk(projectId, modelId, chunkId, patterns)
	}

	return patterns, nil
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
	// Adjust scanner buffer capacity to 250MB per line.
	const maxCapacity = 250 * 1024 * 1024
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

func CreateReaderFromPatterns(patterns []*pattern.Pattern) (*bytes.Reader, error) {
	buf := new(bytes.Buffer)
	for _, pattern := range patterns {
		b, err := json.Marshal(pattern)
		if err != nil {
			return nil, err
		}
		str := string(b)
		_, err = buf.WriteString(fmt.Sprintf("%s\n", str))
		if err != nil {
			return nil, err
		}
	}
	return bytes.NewReader(buf.Bytes()), nil
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

func CreateScannerFromReader(r io.Reader) *bufio.Scanner {
	scanner := bufio.NewScanner(r)
	// Adjust scanner buffer capacity to 10MB per line.
	const maxCapacity = 10 * 1024 * 1024
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)
	return scanner
}
