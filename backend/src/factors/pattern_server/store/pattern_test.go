package store

import (
	"bufio"
	serviceDisk "factors/services/disk"
	U "factors/util"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/stretchr/testify/assert"
)

const (
	testEventName     = "click_100001"
	baseDiskDir       = "/tmp/factors-store-test"
	baseCloudDir      = "/tmp/factors-store-test-bucket"
	eventInfoTestFile = "testdata/event_info.txt"
	patternTestFile   = "testdata/pattern.txt"
)

func openFileAndGetScanner(filePath string) (*bufio.Scanner, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	scanner := CreateScannerFromReader(f)
	return scanner, nil
}

func TestCreatePatternEventInfoFromScanner(t *testing.T) {
	scanner, err := openFileAndGetScanner(eventInfoTestFile)
	assert.Nil(t, err)
	UserAndEventsInfo, err := CreatePatternEventInfoFromScanner(scanner)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(UserAndEventsInfo.UserPropertiesInfo.CategoricalPropertyKeyValues))

	assert.Equal(t, 2, len((*UserAndEventsInfo.EventPropertiesInfoMap)[testEventName].CategoricalPropertyKeyValues["merchant"]))
}

func TestGetModelKey(t *testing.T) {
	projectId := int64(time.Now().Unix())
	modelId := uint64(time.Now().Unix())
	actualK := GetModelKey(projectId, modelId)
	expectedK := fmt.Sprintf("%d%s%d", projectId, IdSeparator, modelId)
	assert.Equal(t, expectedK, actualK)
}

func TestGetChunkKey(t *testing.T) {
	projectId := int64(time.Now().Unix())
	modelId := uint64(time.Now().Unix())
	chunkId := strconv.Itoa(int(time.Now().Unix()))
	actualK := GetChunkKey(projectId, modelId, chunkId)
	expectedK := fmt.Sprintf("%d%s%d%s%s", projectId, IdSeparator, modelId, IdSeparator, chunkId)
	assert.Equal(t, expectedK, actualK)
}

func TestGetModelEventInfoCacheKey(t *testing.T) {
	projectId := int64(time.Now().Unix())
	modelId := uint64(time.Now().Unix())
	actualK := getModelEventInfoCacheKey(projectId, modelId)
	expectedK := fmt.Sprintf("%s%s%d%s%d", "model_event_info", IdSeparator, projectId, IdSeparator, modelId)
	assert.Equal(t, expectedK, actualK)
}

func TestGetModelChunkCacheKey(t *testing.T) {
	projectId := int64(time.Now().Unix())
	modelId := uint64(time.Now().Unix())
	chunkId := strconv.Itoa(int(time.Now().Unix()))
	actualK := getModelChunkCacheKey(projectId, modelId, chunkId)
	expectedK := fmt.Sprintf("%s%s%d%s%d%s%s", "chunk_", IdSeparator, projectId, IdSeparator, modelId, IdSeparator, chunkId)
	assert.Equal(t, expectedK, actualK)
}

func getOriginalPatternsWithMeta(x, y []*PatternWithMeta) ([]PatternWithMeta, []PatternWithMeta) {
	xpwms := make([]PatternWithMeta, 0, 0)
	for _, p := range x {
		xpwms = append(xpwms, *p)
	}

	ypwms := make([]PatternWithMeta, 0, 0)
	for _, p := range y {
		ypwms = append(ypwms, *p)
	}

	return xpwms, ypwms
}

func TestGetPutModelEventInfo(t *testing.T) {

	cloudManager := serviceDisk.New(baseCloudDir)
	diskManager := serviceDisk.New(baseDiskDir)

	scanner, err := openFileAndGetScanner(eventInfoTestFile)
	assert.Nil(t, err)

	actualEventInfoMap, err := CreatePatternEventInfoFromScanner(scanner)
	assert.Nil(t, err)

	store, err := New(5, 5, diskManager, cloudManager)
	assert.Nil(t, err)

	projectId := int64(time.Now().Unix())
	modelId := uint64(time.Now().Unix())

	// should not be present in cloud
	_, err = store.getModelEventInfoFromCloud(projectId, modelId)
	assert.NotNil(t, err)

	// put in cloud
	err = store.PutModelEventInfoInCloud(projectId, modelId, actualEventInfoMap)
	assert.Nil(t, err)

	// should not be present in cache
	_, found := store.getModelEventInfoFromCache(projectId, modelId)
	assert.False(t, found)

	// should not be present in disk
	_, err = store.getModelEventInfoFromDisk(projectId, modelId)
	assert.True(t, os.IsNotExist(err))

	eventInfo, err := store.GetModelEventInfo(projectId, modelId)
	assert.Nil(t, err)
	assert.Equal(t, actualEventInfoMap, eventInfo)

	// read from disk now, should be present on disk also
	diskEventInfo, err := store.getModelEventInfoFromDisk(projectId, modelId)
	assert.Nil(t, err)
	assert.Equal(t, actualEventInfoMap, diskEventInfo)

	// read from cache now, should be present
	cacheEventInfo, found := store.getModelEventInfoFromCache(projectId, modelId)
	assert.True(t, found)
	assert.Equal(t, actualEventInfoMap, cacheEventInfo)
}

func TestGetPutPatterns(t *testing.T) {
	cloudManager := serviceDisk.New(baseCloudDir)
	diskManager := serviceDisk.New(baseDiskDir)
	projectId := int64(time.Now().Unix())
	modelId := uint64(time.Now().Unix())
	chunkId := U.RandomString(8)

	scanner, err := openFileAndGetScanner(patternTestFile)
	assert.Nil(t, err)

	actualPatterns, err := CreatePatternsWithMetaFromScanner(scanner)
	assert.Nil(t, err)

	store, err := New(5, 5, diskManager, cloudManager)
	assert.Nil(t, err)

	// should not be present in cloud
	_, err = store.getPatternsWithMetaFromCloud(projectId, modelId, chunkId)
	assert.NotNil(t, err)

	// put in cloud
	err = store.PutPatternsWithMetaInCloud(projectId, modelId, chunkId, actualPatterns)
	assert.Nil(t, err)

	// should not be present in disk
	_, err = store.getPatternsWithMetaFromDisk(projectId, modelId, chunkId)
	assert.True(t, os.IsNotExist(err))

	// should not be present in cache
	_, found := store.getPatternsWithMetaFromCache(projectId, modelId, chunkId)
	assert.False(t, found)

	patterns, err := store.GetPatternsWithMeta(projectId, modelId, chunkId)
	assert.Nil(t, err)
	ao, po := getOriginalPatternsWithMeta(actualPatterns, patterns)
	assert.Equal(t, ao, po)
	assert.Equal(t, actualPatterns, patterns)

	// read from disk now, should be present
	diskPatterns, err := store.getPatternsWithMetaFromDisk(projectId, modelId, chunkId)
	assert.Nil(t, err)
	assert.Equal(t, actualPatterns, diskPatterns)

	// read from cache now, should be present
	cachePatterns, found := store.getPatternsWithMetaFromCache(projectId, modelId, chunkId)
	assert.True(t, found)
	assert.Equal(t, actualPatterns, cachePatterns)
}

func TestMain(m *testing.M) {
	log.SetLevel(log.DebugLevel)
	statusCode := m.Run()
	cleanup() // not using defer, defer is not called if os.Exit() is used.
	os.Exit(statusCode)
}

func cleanup() {
	dirs := []string{baseCloudDir, baseDiskDir}
	for _, dir := range dirs {
		cleanupDir(dir)
	}
}

func cleanupDir(dir string) error {
	log.Infof("Removing Dir %s\n", dir)
	err := os.RemoveAll(dir)
	if err != nil {
		log.WithError(err).Errorf("Failed to remove dir %s", dir)
	}
	return err
}
