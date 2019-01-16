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
	projectId := U.RandomUint64()
	modelId := U.RandomUint64()

	result := gcsDriver.GetProjectModelDir(projectId, modelId)
	expected := fmt.Sprintf("projects/%d/models/%d/", projectId, modelId)
	assert.Equal(t, expected, result)
}

func TestGetModelEventInfoFilePath(t *testing.T) {
	projectId := U.RandomUint64()
	modelId := U.RandomUint64()

	resultPath, resultName := gcsDriver.GetModelEventInfoFilePathAndName(projectId, modelId)
	expectedPath := gcsDriver.GetProjectModelDir(projectId, modelId)
	expectedName := fmt.Sprintf("event_info_%d.txt", modelId)

	assert.Equal(t, expectedPath, resultPath)
	assert.Equal(t, expectedName, resultName)
}

func TestGetModelPatternsFilePath(t *testing.T) {
	projectId := U.RandomUint64()
	modelId := U.RandomUint64()
	expectedPath := gcsDriver.GetProjectModelDir(projectId, modelId)
	expectedName := fmt.Sprintf("patterns_%d.txt", modelId)

	resultPath, resultName := gcsDriver.GetModelPatternsFilePathAndName(projectId, modelId)

	assert.Equal(t, expectedPath, resultPath)
	assert.Equal(t, expectedName, resultName)
}

func TestGetModelEventsFilePath(t *testing.T) {
	projectId := U.RandomUint64()
	modelId := U.RandomUint64()

	resultPath, resultName := gcsDriver.GetModelEventsFilePathAndName(projectId, modelId)
	expectedPath := gcsDriver.GetProjectModelDir(projectId, modelId)
	expectedName := fmt.Sprintf("events_%d.txt", modelId)

	assert.Equal(t, expectedPath, resultPath)
	assert.Equal(t, expectedName, resultName)
}

func TestGetProjectsDataFilePathAndName(t *testing.T) {
	version := U.RandomString(8)
	expectedPath := "metadata/"
	expectedName := fmt.Sprintf("%s.txt", version)

	resultPath, resultName := gcsDriver.GetProjectsDataFilePathAndName(version)

	assert.Equal(t, expectedPath, resultPath)
	assert.Equal(t, expectedName, resultName)
}

func TestGetPatternChunkFilePathAndName(t *testing.T) {
	projectId := U.RandomUint64()
	modelId := U.RandomUint64()
	chunkId := U.RandomString(8)
	expectedPath := gcsDriver.GetProjectModelDir(projectId, modelId) + "/chunks/"
	expectedName := fmt.Sprintf("chunk_%s.txt", chunkId)

	resultPath, resultName := gcsDriver.GetPatternChunkFilePathAndName(projectId, modelId, chunkId)

	assert.Equal(t, expectedPath, resultPath)
	assert.Equal(t, expectedName, resultName)
}
