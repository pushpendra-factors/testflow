package gcstorage

import (
	"context"
	"factors/filestore"
	U "factors/util"
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

func (gcsd *GCSDriver) GetObjectSize(dir, fileName string) (int64, error) {
	ctx := context.Background()
	if !strings.HasSuffix(dir, "/") {
		// Append / to the end if not present.
		dir = dir + "/"
	}
	var objSize int64
	obj := gcsd.client.Bucket(gcsd.BucketName).Object(dir + fileName)
	if objAttrs, err := obj.Attrs(ctx); err != nil {
		return 0, err
	} else {
		objSize = objAttrs.Size
		return objSize, err
	}
}

func (gcsd *GCSDriver) GetBucketName() string {
	return gcsd.BucketName
}

func (gcsd *GCSDriver) GetProjectModelDir(projectId, modelId uint64) string {
	return fmt.Sprintf("projects/%d/models/%d/", projectId, modelId)
}

func (gcsd *GCSDriver) GetProjectEventFileDir(projectId uint64, startTimestamp int64, modelType string) string {
	dateFormatted := U.GetDateOnlyFromTimestampZ(startTimestamp)
	return fmt.Sprintf("projects/%d/events/%s/%s/", projectId, modelType, dateFormatted)
}

func (gcsd *GCSDriver) GetProjectDir(projectId uint64) string {
	return fmt.Sprintf("projects/%d/events/", projectId)
}

func (gcsd *GCSDriver) GetModelUserPropertiesCategoricalFilePathAndName(projectId, modelId uint64) (string, string) {
	path := gcsd.GetProjectModelDir(projectId, modelId) + "properties/"
	return path, fmt.Sprintf("userPropCatgMap_%d.txt", modelId)
}

func (gcsd *GCSDriver) GetModelEventPropertiesCategoricalFilePathAndName(projectId, modelId uint64) (string, string) {
	path := gcsd.GetProjectModelDir(projectId, modelId) + "properties/"
	return path, fmt.Sprintf("eventPropCatgMap_%d.txt", modelId)
}

func (gcsd *GCSDriver) GetModelUserPropertiesFilePathAndName(projectId, modelId uint64) (string, string) {
	path := gcsd.GetProjectModelDir(projectId, modelId) + "properties/"
	return path, fmt.Sprintf("eventUserPropMap_%d.txt", modelId)
}

func (gcsd *GCSDriver) GetModelEventPropertiesFilePathAndName(projectId, modelId uint64) (string, string) {
	path := gcsd.GetProjectModelDir(projectId, modelId) + "properties/"
	return path, fmt.Sprintf("eventEventPropMap_%d.txt", modelId)
}

func (gcsd *GCSDriver) GetModelEventInfoFilePathAndName(projectId, modelId uint64) (string, string) {
	path := gcsd.GetProjectModelDir(projectId, modelId)
	return path, fmt.Sprintf("event_info_%d.txt", modelId)
}

func (gcsd *GCSDriver) GetModelEventsFilePathAndName(projectId uint64, startTimestamp int64, modelType string) (string, string) {
	path := gcsd.GetProjectEventFileDir(projectId, startTimestamp, modelType)
	return path, "events.txt"
}

func (gcsd *GCSDriver) GetModelEventsBucketingFilePathAndName(projectId uint64, startTimestamp int64, modelType string) (string, string) {
	path := gcsd.GetProjectEventFileDir(projectId, startTimestamp, modelType)
	return path, "events_bucketed.txt"
}

func (gcsd *GCSDriver) GetMasterNumericalBucketsFile(projectId uint64) (string, string) {
	path := gcsd.GetProjectDir(projectId)
	return path, "numerical_buckets_master.txt"
}

func (gcsd *GCSDriver) GetModelEventsNumericalBucketsFile(projectId uint64, startTimestamp int64, modelType string) (string, string) {
	path := gcsd.GetProjectEventFileDir(projectId, startTimestamp, modelType)
	return path, "numerical_buckets.txt"
}

func (gcsd *GCSDriver) GetPatternChunksDir(projectId, modelId uint64) string {
	modelDir := gcsd.GetProjectModelDir(projectId, modelId)
	return fmt.Sprintf("%schunks/", modelDir)
}
func (gcsd *GCSDriver) GetChunksMetaDataDir(projectId, modelId uint64) string {
	modelDir := gcsd.GetProjectModelDir(projectId, modelId)
	return fmt.Sprintf("%smetadata/", modelDir)
}
func (gcsd *GCSDriver) GetPatternChunkFilePathAndName(projectId, modelId uint64, chunkId string) (string, string) {
	return gcsd.GetPatternChunksDir(projectId, modelId), fmt.Sprintf("chunk_%s.txt", chunkId)
}
func (gcsd *GCSDriver) GetChunksMetaDataFilePathAndName(projectId, modelId uint64) (string, string) {
	return gcsd.GetChunksMetaDataDir(projectId, modelId), "metadata.txt"
}
func (gcsd *GCSDriver) GetEventArchiveFilePathAndName(projectID uint64, startTime, endTime int64) (string, string) {
	year, month, date := time.Unix(startTime, 0).UTC().Date()
	path := fmt.Sprintf("archive/%d/%d/%d/", projectID, year, int(month))
	fileName := fmt.Sprintf("%d_events_%d-%d.txt", date, startTime, endTime)
	return path, fileName
}

func (gcsd *GCSDriver) GetUsersArchiveFilePathAndName(projectID uint64, startTime, endTime int64) (string, string) {
	year, month, date := time.Unix(startTime, 0).UTC().Date()
	path := fmt.Sprintf("archive/%d/%d/%d/", projectID, year, int(month))
	fileName := fmt.Sprintf("%d_users_%d-%d.txt", date, startTime, endTime)
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

func (gcsd *GCSDriver) GetInsightsWpiFilePathAndName(projectId uint64, dateString string, queryId uint64, k int) (string, string) {
	path := gcsd.GetWeeklyInsightsModelDir(projectId, dateString, queryId, k)
	return path, "wpi.txt"
}

func (gcsd *GCSDriver) GetInsightsCpiFilePathAndName(projectId uint64, dateString string, queryId uint64, k int) (string, string) {
	path := gcsd.GetWeeklyInsightsModelDir(projectId, dateString, queryId, k)
	return path, "cpi.txt"
}

func (gcsd *GCSDriver) GetWeeklyInsightsModelDir(projectId uint64, dateString string, queryId uint64, k int) string {
	return fmt.Sprintf("projects/%v/weeklyinsights/%v/q-%v/k-%v/", projectId, dateString, queryId, k)
}

func (gcsd *GCSDriver) GetWeeklyKPIModelDir(projectId uint64, dateString string, queryId uint64) string {
	return fmt.Sprintf("projects/%v/weeklyKPI/%v/q-%v/", projectId, dateString, queryId)
}

func (gcsd *GCSDriver) GetKPIFilePathAndName(projectId uint64, dateString string, queryId uint64) (string, string) {
	path := gcsd.GetWeeklyKPIModelDir(projectId, dateString, queryId)
	return path, "kpi.txt"
}
