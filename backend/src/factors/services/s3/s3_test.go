package s3

import (
	U "factors/util"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var s3Driver *S3Driver

// TODO(Ankit):
// Add Create and Get test using localstack
func TestMain(m *testing.M) {
	s3Driver = New("factors-dev-test", "us-east-1")
	os.Exit(m.Run())
}

func TestGetProjectModelDir(t *testing.T) {
	projectId := U.RandomUint64()
	modelId := U.RandomUint64()

	result := s3Driver.GetProjectModelDir(projectId, modelId)
	expected := fmt.Sprintf("projects/%d/models/%d/", projectId, modelId)
	assert.Equal(t, expected, result)
}

func TestGetModelEventInfoFilePath(t *testing.T) {
	projectId := U.RandomUint64()
	modelId := U.RandomUint64()

	resultPath, resultName := s3Driver.GetModelEventInfoFilePathAndName(projectId, modelId)
	expectedPath := s3Driver.GetProjectModelDir(projectId, modelId)
	expectedName := fmt.Sprintf("event_info_%d.txt", modelId)

	assert.Equal(t, expectedPath, resultPath)
	assert.Equal(t, expectedName, resultName)
}

func TestGetModelPatternsFilePath(t *testing.T) {
	projectId := U.RandomUint64()
	modelId := U.RandomUint64()
	expectedPath := s3Driver.GetProjectModelDir(projectId, modelId)
	expectedName := fmt.Sprintf("patterns_%d.txt", modelId)

	resultPath, resultName := s3Driver.GetModelPatternsFilePathAndName(projectId, modelId)

	assert.Equal(t, expectedPath, resultPath)
	assert.Equal(t, expectedName, resultName)
}

func TestGetModelEventsFilePath(t *testing.T) {
	projectId := U.RandomUint64()
	modelId := U.RandomUint64()

	resultPath, resultName := s3Driver.GetModelEventsFilePathAndName(projectId, modelId)
	expectedPath := s3Driver.GetProjectModelDir(projectId, modelId)
	expectedName := fmt.Sprintf("events_%d.txt", modelId)

	assert.Equal(t, expectedPath, resultPath)
	assert.Equal(t, expectedName, resultName)
}

func TestGetProjectsDataFilePathAndName(t *testing.T) {
	version := U.RandomString(8)
	expectedPath := "metadata/"
	expectedName := fmt.Sprintf("%s.txt", version)

	resultPath, resultName := s3Driver.GetProjectsDataFilePathAndName(version)

	assert.Equal(t, expectedPath, resultPath)
	assert.Equal(t, expectedName, resultName)
}
