package gcstorage

import (
	"context"
	"factors/filestore"
	"fmt"
	"io"

	"cloud.google.com/go/storage"
)

const (
	separator = "/"
)

var _ = (*filestore.FileManager)(nil)

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

func (gcsd *GCSDriver) Create(dir, fileName string, reader io.ReadSeeker) error {
	ctx := context.Background()
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
	obj := gcsd.client.Bucket(gcsd.BucketName).Object(dir + fileName)
	rc, err := obj.NewReader(ctx)
	return rc, err
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

func (gcsd *GCSDriver) GetPatternChunkFilePathAndName(projectId, modelId uint64, chunkId string) (string, string) {
	modelDir := gcsd.GetProjectModelDir(projectId, modelId)
	path := fmt.Sprintf("%schunks/", modelDir)
	return path, fmt.Sprintf("chunk_%s.txt", chunkId)
}
