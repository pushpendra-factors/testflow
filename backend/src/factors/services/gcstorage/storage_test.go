package gcstorage

import (
	U "factors/util"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var gcsDriver *GCSDriver

// TODO(Ankit):
// New Client fails because credentials are not provided
// fix this
func TestMain(m *testing.M) {
	var err error
	gcsDriver, err = New("factors-dev-test")
	if err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func TestGetProjectModelDir(t *testing.T) {
	projectId := U.RandomInt64()
	modelId := U.RandomUint64()

	result := gcsDriver.GetProjectModelDir(projectId, modelId)
	expected := fmt.Sprintf("projects/%d/models/%d/", projectId, modelId)
	assert.Equal(t, expected, result)
}

func TestGetModelEventInfoFilePath(t *testing.T) {
	projectId := U.RandomInt64()
	modelId := U.RandomUint64()

	resultPath, resultName := gcsDriver.GetModelEventInfoFilePathAndName(projectId, modelId)
	expectedPath := gcsDriver.GetProjectModelDir(projectId, modelId)
	expectedName := fmt.Sprintf("event_info_%d.txt", modelId)

	assert.Equal(t, expectedPath, resultPath)
	assert.Equal(t, expectedName, resultName)
}

func TestGetModelEventsFilePath(t *testing.T) {
	projectId := U.RandomInt64()
	var startTimestamp int64 = 1640995200 // 1-1-2022 0:00
	var endTimestamp int64 = 1641599999   // 7-1-2022 23:59

	resultPath, resultName := gcsDriver.GetEventsFilePathAndName(projectId, startTimestamp, endTimestamp)
	expectedPath := gcsDriver.GetProjectDir(projectId) + "events/20220101/"
	expectedName := "events_20220101_20220107.txt"

	assert.Equal(t, expectedPath, resultPath)
	assert.Equal(t, expectedName, resultName)
}

func TestGetPatternChunkFilePathAndName(t *testing.T) {
	projectId := U.RandomInt64()
	modelId := U.RandomUint64()
	chunkId := U.RandomString(8)
	expectedPath := gcsDriver.GetProjectModelDir(projectId, modelId) + "chunks/"
	expectedName := fmt.Sprintf("chunk_%s.txt", chunkId)

	resultPath, resultName := gcsDriver.GetPatternChunkFilePathAndName(projectId, modelId, chunkId)

	assert.Equal(t, expectedPath, resultPath)
	assert.Equal(t, expectedName, resultName)
}
