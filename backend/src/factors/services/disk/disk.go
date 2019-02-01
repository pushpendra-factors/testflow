package disk

import (
	"factors/filestore"
	"fmt"
	"io"
	"os"

	log "github.com/sirupsen/logrus"
)

const (
	// TODO Remove this once get and create have been moved to use filepath.Join
	separator = "/"
)

var _ filestore.FileManager = (*DiskDriver)(nil)

type DiskDriver struct {
	// This can be used as namespace
	// to differentiate files across multiple instances of DiskDriver
	// Analogus to bucket name
	baseDir string
}

func New(baseDir string) *DiskDriver {
	return &DiskDriver{baseDir: baseDir}
}

func MkdirAll(path string) error {
	return os.MkdirAll(path, 0755)
}

func (dd *DiskDriver) Create(path, fileName string, reader io.ReadSeeker) error {

	err := MkdirAll(path)
	if err != nil {
		log.WithError(err).Errorln("Failed to create dir")
		return err
	}

	file, err := os.Create(path + fileName)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, reader)
	return err
}

// Get opens a file in read only mode.
// Caller should take care of closing the returned io.ReadCloser.
func (dd *DiskDriver) Get(path, fileName string) (io.ReadCloser, error) {
	log.WithFields(log.Fields{
		"Path":     path,
		"FileName": fileName,
	}).Debug("DiskDriver Opening file")
	file, err := os.OpenFile(path+fileName, os.O_RDONLY, 0444)
	return file, err
}

func (dd *DiskDriver) GetProjectModelDir(projectId, modelId uint64) string {
	return fmt.Sprintf("%s/projects/%d/models/%d/", dd.baseDir, projectId, modelId)
}

func (dd *DiskDriver) GetModelEventInfoFilePathAndName(projectId, modelId uint64) (string, string) {
	path := dd.GetProjectModelDir(projectId, modelId)
	return path, fmt.Sprintf("event_info_%d.txt", modelId)
}

func (dd *DiskDriver) GetModelPatternsFilePathAndName(projectId, modelId uint64) (string, string) {
	path := dd.GetProjectModelDir(projectId, modelId)
	return path, fmt.Sprintf("patterns_%d.txt", modelId)
}

func (dd *DiskDriver) GetModelEventsFilePathAndName(projectId, modelId uint64) (string, string) {
	path := dd.GetProjectModelDir(projectId, modelId)
	return path, fmt.Sprintf("events_%d.txt", modelId)
}

func (dd *DiskDriver) GetProjectsDataFilePathAndName(version string) (string, string) {
	path := fmt.Sprintf("%s/metadata/", dd.baseDir)
	return path, fmt.Sprintf("%s.txt", version)
}

func (dd *DiskDriver) GetPatternChunkFilePathAndName(projectId, modelId uint64, chunkId string) (string, string) {
	modelDir := dd.GetProjectModelDir(projectId, modelId)
	path := fmt.Sprintf("%schunks/", modelDir)
	return path, fmt.Sprintf("chunk_%s.txt", chunkId)
}
