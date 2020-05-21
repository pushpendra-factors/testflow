package filestore

import (
	"io"
)

type FileManager interface {
	Create(dir, fileName string, reader io.Reader) error
	Get(path, fileName string) (io.ReadCloser, error)
	// Del(dir, filename string)error
	GetBucketName() string
	GetProjectModelDir(projectId, modelId uint64) string
	GetModelEventInfoFilePathAndName(projectId, modelId uint64) (string, string)
	GetModelPatternsFilePathAndName(projectId, modelId uint64) (string, string)
	GetModelEventsFilePathAndName(projectId, modelId uint64) (string, string)
	GetProjectsDataFilePathAndName(version string) (string, string)
	GetPatternChunksDir(projectId, modelId uint64) string
	GetPatternChunkFilePathAndName(projectId, modelId uint64, chunkId string) (string, string)
	GetEventArchiveFilePathAndName(projectID uint64, startTime, endTime int64) (string, string)
	ListFiles(path string) []string
}
