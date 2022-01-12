package s3

import (
	"factors/filestore"
	U "factors/util"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const (
	separator = "/"
)

var _ filestore.FileManager = (*S3Driver)(nil)

type S3Driver struct {
	s3         *s3.S3
	BucketName string
	Region     string
}

func New(bucketName, region string) *S3Driver {
	session := session.New()
	s3 := s3.New(session, aws.NewConfig().WithRegion(region))
	return &S3Driver{s3: s3, BucketName: bucketName, Region: region}
}

func (sd *S3Driver) Create(dir, fileName string, reader io.Reader) error {
	// log.WithFields(log.Fields{
	// 	"Dir":        dir,
	// 	"BucketName": sd.BucketName,
	// 	"Region":     sd.Region,
	// }).Debug("S3Driver Creating file")

	// // add
	// // SSE
	// // content type
	// // any key value metadata if needed
	// input := &s3.PutObjectInput{
	// 	Bucket: aws.String(sd.BucketName),
	// 	Body:   reader,
	// 	Key:    aws.String(dir + separator + fileName),
	// }
	// _, err := sd.s3.PutObject(input)
	// return err
	return nil
}

func (sd *S3Driver) Get(dir, fileName string) (io.ReadCloser, error) {
	input := s3.GetObjectInput{
		Bucket: aws.String(sd.BucketName),
		Key:    aws.String(dir + separator + fileName),
	}
	op, err := sd.s3.GetObject(&input)
	return op.Body, err
}

func (sd *S3Driver) GetObjectSize(dir, fileName string) (int64, error) {
	input := s3.GetObjectInput{
		Bucket: aws.String(sd.BucketName),
		Key:    aws.String(dir + separator + fileName),
	}
	op, err := sd.s3.GetObject(&input)
	objSize := op.ContentLength
	return *objSize, err
}

func (sd *S3Driver) GetProjectModelDir(projectId, modelId uint64) string {
	return fmt.Sprintf("projects/%d/models/%d/", projectId, modelId)
}

func (sd *S3Driver) GetProjectEventFileDir(projectId uint64, startTimestamp int64, modelType string) string {
	dateFormatted := U.GetDateOnlyFromTimestampZ(startTimestamp)
	return fmt.Sprintf("projects/%d/events/%s/%s/", projectId, modelType, dateFormatted)
}

func (sd *S3Driver) GetProjectDir(projectId uint64) string {
	return fmt.Sprintf("projects/%d/events/", projectId)
}

func (gcsd *S3Driver) GetModelUserPropertiesCategoricalFilePathAndName(projectId, modelId uint64) (string, string) {
	path := gcsd.GetProjectModelDir(projectId, modelId) + "properties/"
	return path, fmt.Sprintf("userPropCatgMap_%d.txt", modelId)
}

func (gcsd *S3Driver) GetModelEventPropertiesCategoricalFilePathAndName(projectId, modelId uint64) (string, string) {
	path := gcsd.GetProjectModelDir(projectId, modelId) + "properties/"
	return path, fmt.Sprintf("eventPropCatgMap_%d.txt", modelId)
}

func (gcsd *S3Driver) GetModelUserPropertiesFilePathAndName(projectId, modelId uint64) (string, string) {
	path := gcsd.GetProjectModelDir(projectId, modelId) + "properties/"
	return path, fmt.Sprintf("eventUserPropMap_%d.txt", modelId)
}

func (gcsd *S3Driver) GetModelEventPropertiesFilePathAndName(projectId, modelId uint64) (string, string) {
	path := gcsd.GetProjectModelDir(projectId, modelId) + "properties/"
	return path, fmt.Sprintf("eventEventPropMap_%d.txt", modelId)
}

func (sd *S3Driver) GetModelEventInfoFilePathAndName(projectId, modelId uint64) (string, string) {
	path := sd.GetProjectModelDir(projectId, modelId)
	return path, fmt.Sprintf("event_info_%d.txt", modelId)
}

func (sd *S3Driver) GetModelEventsFilePathAndName(projectId uint64, startTimestamp int64, modelType string) (string, string) {
	path := sd.GetProjectEventFileDir(projectId, startTimestamp, modelType)
	return path, fmt.Sprintf("events.txt")
}

func (sd *S3Driver) GetModelEventsBucketingFilePathAndName(projectId uint64, startTimestamp int64, modelType string) (string, string) {
	path := sd.GetProjectEventFileDir(projectId, startTimestamp, modelType)
	return path, fmt.Sprintf("events_bucketed.txt")
}

func (sd *S3Driver) GetMasterNumericalBucketsFile(projectId uint64) (string, string) {
	path := sd.GetProjectDir(projectId)
	return path, fmt.Sprintf("numerical_buckets_master.txt")
}

func (sd *S3Driver) GetModelEventsNumericalBucketsFile(projectId uint64, startTimestamp int64, modelType string) (string, string) {
	path := sd.GetProjectEventFileDir(projectId, startTimestamp, modelType)
	return path, fmt.Sprintf("numerical_buckets.txt")
}

func (sd *S3Driver) GetPatternChunksDir(projectId, modelId uint64) string {
	modelDir := sd.GetProjectModelDir(projectId, modelId)
	return fmt.Sprintf("%schunks/", modelDir)
}
func (sd *S3Driver) GetChunksMetaDataDir(projectId, modelId uint64) string {
	modelDir := sd.GetProjectModelDir(projectId, modelId)
	return fmt.Sprintf("%smetadata/", modelDir)
}

// GetPatternChunkFilePathAndName - Placeholder definition. Has to be implemented.
func (sd *S3Driver) GetPatternChunkFilePathAndName(projectId, modelId uint64, chunkId string) (string, string) {
	return sd.GetPatternChunksDir(projectId, modelId), fmt.Sprintf("chunk_%s.txt", chunkId)
}
func (sd *S3Driver) GetChunksMetaDataFilePathAndName(projectId, modelId uint64) (string, string) {
	return sd.GetChunksMetaDataDir(projectId, modelId), fmt.Sprintf("metadata.txt")
}
// GetEventArchiveFilePathAndName - Placeholder definition. Has to be implemented.
func (sd *S3Driver) GetEventArchiveFilePathAndName(projectID uint64, startTime, endTime int64) (string, string) {
	return "", ""
}

// GetUsersArchiveFilePathAndName - Placeholder definition. Has to be implemented.
func (sd *S3Driver) GetUsersArchiveFilePathAndName(projectID uint64, startTime, endTime int64) (string, string) {
	return "", ""
}

// ListFiles - Placeholder definition. Has to be implemented.
func (sd *S3Driver) ListFiles(path string) []string {
	return []string{}
}

// GetBucketName - Placeholder definition. Has to be implemented.
func (sd *S3Driver) GetBucketName() string {
	return ""
}

func (sd *S3Driver) GetInsightsWpiFilePathAndName(projectId uint64, dateString string, queryId uint64, k int) (string, string) {
	path := sd.GetWeeklyInsightsModelDir(projectId, dateString, queryId, k)
	return path, fmt.Sprintf("wpi.txt")
}

func (sd *S3Driver) GetInsightsCpiFilePathAndName(projectId uint64, dateString string, queryId uint64, k int) (string, string) {
	path := sd.GetWeeklyInsightsModelDir(projectId, dateString, queryId, k)
	return path, fmt.Sprintf("cpi.txt")
}

func (sd *S3Driver) GetWeeklyInsightsModelDir(projectId uint64, dateString string, queryId uint64, k int) string {
	return fmt.Sprintf("projects/%v/weeklyinsights/%v/q-%v/k-%v/", projectId, dateString, queryId, k)
}
