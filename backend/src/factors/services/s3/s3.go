package s3

import (
	"factors/filestore"
	U "factors/util"
	"fmt"
	"io"

	pb "path/filepath"

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

func (sd *S3Driver) GetProjectModelDir(projectId int64, modelId uint64) string {
	return fmt.Sprintf("projects/%d/models/%d/", projectId, modelId)
}

func (sd *S3Driver) GetProjectEventFileDir(projectId int64, startTimestamp int64, modelType string) string {
	dateFormatted := U.GetDateOnlyFromTimestampZ(startTimestamp)
	return fmt.Sprintf("projects/%d/events/%s/%s/", projectId, modelType, dateFormatted)
}

func (sd *S3Driver) GetProjectDir(projectId int64) string {
	return pb.Join("projects", fmt.Sprintf("%d", projectId))
	// return fmt.Sprintf("projects/%d/", projectId)
}

func (gcsd *S3Driver) GetModelUserPropertiesCategoricalFilePathAndName(projectId int64, modelId uint64) (string, string) {
	path := gcsd.GetProjectModelDir(projectId, modelId) + "properties/"
	return path, fmt.Sprintf("userPropCatgMap_%d.txt", modelId)
}

func (gcsd *S3Driver) GetModelEventPropertiesCategoricalFilePathAndName(projectId int64, modelId uint64) (string, string) {
	path := gcsd.GetProjectModelDir(projectId, modelId) + "properties/"
	return path, fmt.Sprintf("eventPropCatgMap_%d.txt", modelId)
}

func (gcsd *S3Driver) GetModelUserPropertiesFilePathAndName(projectId int64, modelId uint64) (string, string) {
	path := gcsd.GetProjectModelDir(projectId, modelId) + "properties/"
	return path, fmt.Sprintf("eventUserPropMap_%d.txt", modelId)
}

func (gcsd *S3Driver) GetModelEventPropertiesFilePathAndName(projectId int64, modelId uint64) (string, string) {
	path := gcsd.GetProjectModelDir(projectId, modelId) + "properties/"
	return path, fmt.Sprintf("eventEventPropMap_%d.txt", modelId)
}

func (sd *S3Driver) GetModelEventInfoFilePathAndName(projectId int64, modelId uint64) (string, string) {
	path := sd.GetProjectModelDir(projectId, modelId)
	return path, fmt.Sprintf("event_info_%d.txt", modelId)
}

func (sd *S3Driver) GetModelEventsFilePathAndName(projectId int64, startTimestamp int64, modelType string) (string, string) {
	path := sd.GetProjectEventFileDir(projectId, startTimestamp, modelType)
	return path, "events.txt"
}
func (sd *S3Driver) GetModelMetricsFilePathAndName(projectId int64, startTimestamp int64, modelType string) (string, string) {
	path := sd.GetProjectEventFileDir(projectId, startTimestamp, modelType)
	return path, "metrics.txt"
}

func (sd *S3Driver) GetModelChannelFilePathAndName(channel string, projectId int64, startTimestamp int64, modelType string) (string, string) {
	path := sd.GetProjectEventFileDir(projectId, startTimestamp, modelType)
	return path, channel + ".txt"
}

func (sd *S3Driver) GetModelUsersFilePathAndName(dateField string, projectId int64, startTimestamp int64, modelType string) (string, string) {
	path := sd.GetModelUsersDir(dateField, projectId, startTimestamp, modelType)
	return path, dateField + ".txt"
}

func (sd *S3Driver) GetModelUsersDir(dateField string, projectId int64, startTimestamp int64, modelType string) string {
	path := sd.GetProjectEventFileDir(projectId, startTimestamp, modelType)
	return fmt.Sprintf("%susers/", path)
}

func (sd *S3Driver) GetModelEventsBucketingFilePathAndName(projectId int64, startTimestamp int64, modelType string) (string, string) {
	path := sd.GetProjectEventFileDir(projectId, startTimestamp, modelType)
	return path, "events_bucketed.txt"
}

func (sd *S3Driver) GetMasterNumericalBucketsFile(projectId int64) (string, string) {
	path := sd.GetProjectDir(projectId)
	path = pb.Join(path, "events")
	return path, "numerical_buckets_master.txt"
}

func (sd *S3Driver) GetModelEventsNumericalBucketsFile(projectId int64, startTimestamp int64, modelType string) (string, string) {
	path := sd.GetProjectEventFileDir(projectId, startTimestamp, modelType)
	return path, "numerical_buckets.txt"
}

func (sd *S3Driver) GetPatternChunksDir(projectId int64, modelId uint64) string {
	modelDir := sd.GetProjectModelDir(projectId, modelId)
	return fmt.Sprintf("%schunks/", modelDir)
}
func (sd *S3Driver) GetChunksMetaDataDir(projectId int64, modelId uint64) string {
	modelDir := sd.GetProjectModelDir(projectId, modelId)
	return fmt.Sprintf("%smetadata/", modelDir)
}

// GetPatternChunkFilePathAndName - Placeholder definition. Has to be implemented.
func (sd *S3Driver) GetPatternChunkFilePathAndName(projectId int64, modelId uint64, chunkId string) (string, string) {
	return sd.GetPatternChunksDir(projectId, modelId), fmt.Sprintf("chunk_%s.txt", chunkId)
}
func (sd *S3Driver) GetChunksMetaDataFilePathAndName(projectId int64, modelId uint64) (string, string) {
	return sd.GetChunksMetaDataDir(projectId, modelId), "metadata.txt"
}

// GetEventArchiveFilePathAndName - Placeholder definition. Has to be implemented.
func (sd *S3Driver) GetEventArchiveFilePathAndName(projectID int64, startTime, endTime int64) (string, string) {
	return "", ""
}

// GetUsersArchiveFilePathAndName - Placeholder definition. Has to be implemented.
func (sd *S3Driver) GetUsersArchiveFilePathAndName(projectID int64, startTime, endTime int64) (string, string) {
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

func (sd *S3Driver) GetInsightsWpiFilePathAndName(projectId int64, dateString string, queryId int64, k int, mailerRun bool) (string, string) {
	path := ""
	if mailerRun == true {
		path = sd.GetWeeklyInsightsMailerModelDir(projectId, dateString, queryId, k)
	} else {
		path = sd.GetWeeklyInsightsModelDir(projectId, dateString, queryId, k)
	}
	return path, "wpi.txt"
}

func (sd *S3Driver) GetInsightsCpiFilePathAndName(projectId int64, dateString string, queryId int64, k int, mailerRun bool) (string, string) {
	path := ""
	if mailerRun == true {
		path = sd.GetWeeklyInsightsMailerModelDir(projectId, dateString, queryId, k)
	} else {
		path = sd.GetWeeklyInsightsModelDir(projectId, dateString, queryId, k)
	}
	return path, "cpi.txt"
}

func (sd *S3Driver) GetWeeklyInsightsModelDir(projectId int64, dateString string, queryId int64, k int) string {
	return fmt.Sprintf("projects/%v/weeklyinsights/%v/q-%v/k-%v/", projectId, dateString, queryId, k)
}

func (sd *S3Driver) GetWeeklyInsightsMailerModelDir(projectId int64, dateString string, queryId int64, k int) string {
	return fmt.Sprintf("projects/%v/weeklyinsightsmailer/%v/q-%v/k-%v/", projectId, dateString, queryId, k)
}

func (sd *S3Driver) GetWeeklyKPIModelDir(projectId int64, dateString string, queryId int64) string {
	return fmt.Sprintf("projects/%v/weeklyKPI/%v/q-%v/", projectId, dateString, queryId)
}

func (sd *S3Driver) GetKPIFilePathAndName(projectId int64, dateString string, queryId int64) (string, string) {
	path := sd.GetWeeklyKPIModelDir(projectId, dateString, queryId)
	return path, "kpi.txt"
}

func (sd *S3Driver) GetAdsDataDir(projectId int64) string {
	return fmt.Sprintf("projects/%v/AdsImport/", projectId)
}

func (sd *S3Driver) GetAdsDataFilePathAndName(projectId int64, report string, chunkNo int) (string, string) {
	path := sd.GetAdsDataDir(projectId)
	return path, fmt.Sprintf("%v-%v-%v.csv", report, projectId, chunkNo)
}

func (sd *S3Driver) GetPredictProjectDataPath(projectId int64, model_id int64) string {
	path := sd.GetPredictProjectDir(projectId, model_id)
	return pb.Join(path, "data")
}

func (sd *S3Driver) GetPredictProjectDir(projectId int64, model_id int64) string {
	path := sd.GetProjectDir(projectId)
	model_str := fmt.Sprintf("%d", model_id)
	return pb.Join(path, "predict", model_str)
}

func (sd *S3Driver) GetWIPropertiesPathAndName(projectId int64) (string, string) {
	path := sd.GetWIPropertiesDir(projectId)
	return path, "properties.txt"
}

func (sd *S3Driver) GetWIPropertiesDir(projectId int64) string {
	return fmt.Sprintf("projects/%v/weeklyinsights/", projectId)
}

func (sd *S3Driver) GetModelEventsUnsortedFilePathAndName(projectId int64, startTimestamp int64, modelType string) (string, string) {
	path := sd.GetProjectEventFileDir(projectId, startTimestamp, modelType)
	return path, "events_raw.txt"
}
func (sd *S3Driver) GetEventsArtificatFilePathAndName(projectId int64, startTimestamp int64, modelType string) (string, string) {
	path := sd.GetProjectEventFileDir(projectId, startTimestamp, modelType)
	path = pb.Join(path, "artifacts")
	return path, "users_map.txt"

}

func (sd *S3Driver) GetEventsForTimerangeFilePathAndName(projectId int64, startTimestamp int64, endTimestamp int64) (string, string) {
	path := sd.GetEventsForTimerangeFileDir(projectId, startTimestamp, endTimestamp)
	return path, "events.txt"
}

func (sd *S3Driver) GetEventsForTimerangeFileDir(projectId int64, startTimestamp int64, endTimestamp int64) string {
	dateFormattedStart := U.GetDateOnlyFromTimestampZ(startTimestamp)
	dateFormattedEnd := U.GetDateOnlyFromTimestampZ(endTimestamp)
	return fmt.Sprintf("projects/%v/%v/%v/", projectId, dateFormattedStart, dateFormattedEnd)
}

func (sd *S3Driver) GetPathAnalysisTempFileDir(id string, projectId int64) string {
	return fmt.Sprintf("projects/%v/pathanalysis/%v/", projectId, id)
}
func (sd *S3Driver) GetPathAnalysisTempFilePathAndName(id string, projectId int64) (string, string) {
	path := sd.GetPathAnalysisTempFileDir(id, projectId)
	return path, "patterns.txt"
}
