package gcstorage

import (
	"context"
	"factors/filestore"
	"fmt"
	"io"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/iterator"
)

const (
	separator = "/"
)

var _ filestore.FileManager = (*GCSDriver)(nil)

type GCSDriver struct {
	client     *storage.Client
	BucketName string
}

func New(bucketName string) (*GCSDriver, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	d := &GCSDriver{
		BucketName: bucketName,
		client:     client,
	}
	return d, nil
}

func (gcsd *GCSDriver) Create(dir, fileName string, reader io.Reader) error {
	ctx := context.Background()
	if !strings.HasSuffix(dir, "/") {
		// Append / to the end if not present.
		dir = dir + "/"
	}

	obj := gcsd.client.Bucket(gcsd.BucketName).Object(dir + fileName)
	w := obj.NewWriter(ctx)
	if _, err := io.Copy(w, reader); err != nil {
		return err
	}
	err := w.Close()
	return err
}

func (gcsd *GCSDriver) Get(dir, fileName string) (io.ReadCloser, error) {
	ctx := context.Background()
	if !strings.HasSuffix(dir, "/") {
		// Append / to the end if not present.
		dir = dir + "/"
	}
	obj := gcsd.client.Bucket(gcsd.BucketName).Object(dir + fileName)
	rc, err := obj.NewReader(ctx)
	return rc, err
}

func (gcsd *GCSDriver) GetBucketName() string {
	return gcsd.BucketName
}

func (gcsd *GCSDriver) GetProjectModelDir(projectId, modelId uint64) string {
	return fmt.Sprintf("projects/%d/models/%d/", projectId, modelId)
}

func (gcsd *GCSDriver) GetModelEventInfoFilePathAndName(projectId, modelId uint64) (string, string) {
	path := gcsd.GetProjectModelDir(projectId, modelId)
	return path, fmt.Sprintf("event_info_%d.txt", modelId)
}

func (gcsd *GCSDriver) GetModelPatternsFilePathAndName(projectId, modelId uint64) (string, string) {
	path := gcsd.GetProjectModelDir(projectId, modelId)
	return path, fmt.Sprintf("patterns_%d.txt", modelId)
}

func (gcsd *GCSDriver) GetModelEventsFilePathAndName(projectId, modelId uint64) (string, string) {
	path := gcsd.GetProjectModelDir(projectId, modelId)
	return path, fmt.Sprintf("events_%d.txt", modelId)
}

func (gcsd *GCSDriver) GetProjectsDataFilePathAndName(version string) (string, string) {
	return "metadata/", fmt.Sprintf("%s.txt", version)
}

func (gcsd *GCSDriver) GetPatternChunksDir(projectId, modelId uint64) string {
	modelDir := gcsd.GetProjectModelDir(projectId, modelId)
	return fmt.Sprintf("%schunks/", modelDir)
}

func (gcsd *GCSDriver) GetPatternChunkFilePathAndName(projectId, modelId uint64, chunkId string) (string, string) {
	return gcsd.GetPatternChunksDir(projectId, modelId), fmt.Sprintf("chunk_%s.txt", chunkId)
}

func (gcsd *GCSDriver) GetEventArchiveFilePathAndName(projectID uint64, startTime, endTime int64) (string, string) {
	year, month, date := time.Unix(startTime, 0).UTC().Date()
	path := fmt.Sprintf("archive/%d/%d/%d/", projectID, year, int(month))
	fileName := fmt.Sprintf("%d_%d-%d.txt", date, startTime, endTime)
	return path, fileName
}

// ListFiles List files present in a folder in cloud storage. Prefix has to be without bucket name.
// Must not have leading '/' and should have trailing '/' in prefix. Ex: archive/3/.
func (gcsd *GCSDriver) ListFiles(prefix string) []string {
	var files []string
	if !strings.HasSuffix(prefix, "/") {
		prefix = prefix + "/"
	}

	ctx := context.Background()
	pathQuery := &storage.Query{Prefix: prefix}
	filesIterator := gcsd.client.Bucket(gcsd.BucketName).Objects(ctx, pathQuery)
	for {
		attributes, err := filesIterator.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			log.WithError(err).Errorf("Failed to list file. Attributes: %v\n", attributes)
			continue
		} else if attributes.Name == prefix || attributes.Name == (prefix+"/") {
			// Omit the base prefix if returned as one the objects.
			continue
		}
		files = append(files, attributes.Name)
	}

	return files
}
