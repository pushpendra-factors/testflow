package filestore

import (
	"io"
)

type FileManager interface {
	Create(dir, fileName string, reader io.Reader) error
	Get(path, fileName string) (io.ReadCloser, error)
	GetObjectSize(dir, fileName string) (int64, error)
	// Del(dir, filename string)error
	GetBucketName() string
	GetProjectModelDir(projectId, modelId uint64) string
	GetProjectDir(projectId uint64) string
	GetProjectEventFileDir(projectId uint64, startTimestamp int64, modelType string) string
	GetModelEventInfoFilePathAndName(projectId, modelId uint64) (string, string)
	GetModelEventsFilePathAndName(projectId uint64, startTimestamp int64, modelType string) (string, string)
	GetModelEventsBucketingFilePathAndName(projectId uint64, startTimestamp int64, modelType string) (string, string)
	GetMasterNumericalBucketsFile(projectId uint64) (string, string)
	GetModelEventsNumericalBucketsFile(projectId uint64, startTimestamp int64, modelType string) (string, string)
	GetPatternChunksDir(projectId, modelId uint64) string
	GetChunksMetaDataDir(projectId, modelId uint64) string
	GetPatternChunkFilePathAndName(projectId, modelId uint64, chunkId string) (string, string)
	GetChunksMetaDataFilePathAndName(projectId, modelId uint64) (string, string)
	GetEventArchiveFilePathAndName(projectID uint64, startTime, endTime int64) (string, string)
	GetUsersArchiveFilePathAndName(projectID uint64, startTime, endTime int64) (string, string)
	ListFiles(path string) []string
	GetInsightsWpiFilePathAndName(projectId uint64, dateString string, queryId uint64, k int) (string, string)
	GetInsightsCpiFilePathAndName(projectId uint64, dateString string, queryId uint64, k int) (string, string)
	GetWeeklyInsightsModelDir(projectId uint64, dateString string, queryId uint64, k int) string
	GetModelUserPropertiesCategoricalFilePathAndName(projectId, modelId uint64) (string, string)
	GetModelEventPropertiesCategoricalFilePathAndName(projectId, modelId uint64) (string, string)
	GetModelUserPropertiesFilePathAndName(projectId, modelId uint64) (string, string)
	GetModelEventPropertiesFilePathAndName(projectId, modelId uint64) (string, string)
	GetWeeklyKPIModelDir(projectId uint64, dateString string, queryId uint64) string
	GetKPIFilePathAndName(projectId uint64, dateString string, queryId uint64) (string, string)
}
