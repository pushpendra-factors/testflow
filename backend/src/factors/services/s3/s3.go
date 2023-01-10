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

func (sd *S3Driver) GetProjectDir(projectId int64) string {
	return pb.Join("projects", fmt.Sprintf("%d", projectId))
	// return fmt.Sprintf("projects/%d/", projectId)
}

func (sd *S3Driver) GetProjectModelDir(projectId int64, modelId uint64) string {
	path := sd.GetProjectDir(projectId)
	return fmt.Sprintf("%smodels/%d/", path, modelId)
}

func (sd *S3Driver) GetProjectDataFileDir(projectId int64, startTimestamp int64, dataType string) string {
	path := sd.GetProjectDir(projectId)
	dateFormatted := U.GetDateOnlyFromTimestampZ(startTimestamp)
	return fmt.Sprintf("%s%s/%s/", path, dataType, dateFormatted)
}

func (sd *S3Driver) GetModelUserPropertiesCategoricalFilePathAndName(projectId int64, modelId uint64) (string, string) {
	path := sd.GetProjectModelDir(projectId, modelId) + "properties/"
	return path, fmt.Sprintf("userPropCatgMap_%d.txt", modelId)
}

func (sd *S3Driver) GetModelEventPropertiesCategoricalFilePathAndName(projectId int64, modelId uint64) (string, string) {
	path := sd.GetProjectModelDir(projectId, modelId) + "properties/"
	return path, fmt.Sprintf("eventPropCatgMap_%d.txt", modelId)
}

func (sd *S3Driver) GetModelUserPropertiesFilePathAndName(projectId int64, modelId uint64) (string, string) {
	path := sd.GetProjectModelDir(projectId, modelId) + "properties/"
	return path, fmt.Sprintf("eventUserPropMap_%d.txt", modelId)
}

func (sd *S3Driver) GetModelEventPropertiesFilePathAndName(projectId int64, modelId uint64) (string, string) {
	path := sd.GetProjectModelDir(projectId, modelId) + "properties/"
	return path, fmt.Sprintf("eventEventPropMap_%d.txt", modelId)
}

func (sd *S3Driver) GetModelEventInfoFilePathAndName(projectId int64, modelId uint64) (string, string) {
	path := sd.GetProjectModelDir(projectId, modelId)
	return path, fmt.Sprintf("event_info_%d.txt", modelId)
}

func (sd *S3Driver) GetEventsFilePathAndName(projectId int64, startTimestamp, endTimestamp int64) (string, string) {
	path := sd.GetProjectDataFileDir(projectId, startTimestamp, U.DataTypeEvent)
	dateFormattedStart := U.GetDateOnlyFromTimestampZ(startTimestamp)
	dateFormattedEnd := U.GetDateOnlyFromTimestampZ(endTimestamp)
	fileName := fmt.Sprintf("events_%s-%s.txt", dateFormattedStart, dateFormattedEnd)
	return path, fileName
}

func (sd *S3Driver) GetEventsGroupFilePathAndName(projectId int64, startTimestamp, endTimestamp int64, group int) (string, string) {
	if group == 0 {
		return sd.GetEventsFilePathAndName(projectId, startTimestamp, endTimestamp)
	}
	var fileName string
	path := sd.GetProjectDataFileDir(projectId, startTimestamp, U.DataTypeEvent)
	dateFormattedStart := U.GetDateOnlyFromTimestampZ(startTimestamp)
	dateFormattedEnd := U.GetDateOnlyFromTimestampZ(endTimestamp)
	fileName = fmt.Sprintf("events_group%d_%s-%s.txt", group, dateFormattedStart, dateFormattedEnd)
	return path, fileName
}

func (sd *S3Driver) GetChannelFilePathAndName(channel string, projectId int64, startTimestamp, endTimestamp int64) (string, string) {
	path := sd.GetProjectDataFileDir(projectId, startTimestamp, "ad_reports")
	dateFormattedStart := U.GetDateOnlyFromTimestampZ(startTimestamp)
	dateFormattedEnd := U.GetDateOnlyFromTimestampZ(endTimestamp)
	fileName := fmt.Sprintf("%s_%s-%s.txt", channel, dateFormattedStart, dateFormattedEnd)
	return path, fileName
}

func (sd *S3Driver) GetUsersFilePathAndName(dateField string, projectId int64, startTimestamp, endTimestamp int64) (string, string) {
	path := sd.GetProjectDataFileDir(projectId, startTimestamp, U.DataTypeUser)
	dateFormattedStart := U.GetDateOnlyFromTimestampZ(startTimestamp)
	dateFormattedEnd := U.GetDateOnlyFromTimestampZ(endTimestamp)
	fileName := fmt.Sprintf("%s_%s-%s.txt", dateField, dateFormattedStart, dateFormattedEnd)
	return path, fileName
}

func (sd *S3Driver) GetModelMetricsFilePathAndName(projectId int64, startTimestamp, endTimestamp int64) (string, string) {
	path := sd.GetProjectDataFileDir(projectId, startTimestamp, "metrics")
	dateFormattedStart := U.GetDateOnlyFromTimestampZ(startTimestamp)
	dateFormattedEnd := U.GetDateOnlyFromTimestampZ(endTimestamp)
	return path, fmt.Sprintf("metrics_%s-%s.txt", dateFormattedStart, dateFormattedEnd)
}

func (sd *S3Driver) GetModelAlertsFilePathAndName(projectId int64, startTimestamp int64, endTimestamp int64) (string, string) {
	path := sd.GetProjectDataFileDir(projectId, startTimestamp, "alerts")
	dateFormattedStart := U.GetDateOnlyFromTimestampZ(startTimestamp)
	dateFormattedEnd := U.GetDateOnlyFromTimestampZ(endTimestamp)
	return path, fmt.Sprintf("alerts_%s-%s.txt", dateFormattedStart, dateFormattedEnd)
}

func (sd *S3Driver) GetModelEventsBucketingFilePathAndName(projectId int64, startTimestamp, endTimestamp int64) (string, string) {
	path := sd.GetProjectDataFileDir(projectId, startTimestamp, U.DataTypeEvent)
	path = pb.Join(path, "events_bucketed")
	dateFormattedStart := U.GetDateOnlyFromTimestampZ(startTimestamp)
	dateFormattedEnd := U.GetDateOnlyFromTimestampZ(endTimestamp)
	return path, fmt.Sprintf("events_bucketed_%s-%s.txt", dateFormattedStart, dateFormattedEnd)
}

func (sd *S3Driver) GetMasterNumericalBucketsFile(projectId int64) (string, string) {
	path := sd.GetProjectDir(projectId)
	path = pb.Join(path, U.DataTypeEvent)
	return path, "numerical_buckets_master.txt"
}

func (sd *S3Driver) GetModelEventsNumericalBucketsFile(projectId int64, startTimestamp, endTimestamp int64) (string, string) {
	path := sd.GetProjectDataFileDir(projectId, startTimestamp, U.DataTypeEvent)
	path = pb.Join(path, "numerical_buckets")
	dateFormattedStart := U.GetDateOnlyFromTimestampZ(startTimestamp)
	dateFormattedEnd := U.GetDateOnlyFromTimestampZ(endTimestamp)
	return path, fmt.Sprintf("numerical_buckets_%s-%s.txt", dateFormattedStart, dateFormattedEnd)
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

func (sd *S3Driver) GetDailyArchiveFilesDir(projectID int64, dataTimestamp int64, dataType string) string {
	dateFormatted := U.GetDateOnlyFromTimestampZ(dataTimestamp)
	path := fmt.Sprintf("daily_pull/%d/%s/%s/", projectID, dateFormatted, dataType)
	return path
}

func (sd *S3Driver) GetDailyEventArchiveFilePathAndName(projectID int64, dataTimestamp int64, startTime, endTime int64) (string, string) {
	path := sd.GetDailyArchiveFilesDir(projectID, dataTimestamp, U.DataTypeEvent)
	fileName := fmt.Sprintf("events_created_at_%d-%d.txt", startTime, endTime)
	return path, fileName
}

func (sd *S3Driver) GetDailyUsersArchiveFilePathAndName(dateField string, projectID int64, dataTimestamp int64, startTime, endTime int64) (string, string) {
	path := sd.GetDailyArchiveFilesDir(projectID, dataTimestamp, U.DataTypeUser)
	fileName := fmt.Sprintf("%s_created_at_%d-%d.txt", dateField, startTime, endTime)
	return path, fileName
}

func (sd *S3Driver) GetDailyChannelArchiveFilePathAndName(channel string, projectID int64, dataTimestamp int64, startTime, endTime int64) (string, string) {
	path := sd.GetDailyArchiveFilesDir(projectID, dataTimestamp, U.DataTypeAdReport)
	fileName := fmt.Sprintf("%s_created_at_%d-%d.txt", channel, startTime, endTime)
	return path, fileName
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
	path := sd.GetWIPropertiesDir(projectId)
	return fmt.Sprintf("%s%v/q-%v/k-%v/", path, dateString, queryId, k)
}

func (sd *S3Driver) GetWeeklyInsightsMailerModelDir(projectId int64, dateString string, queryId int64, k int) string {
	path := sd.GetWIPropertiesDir(projectId)
	return fmt.Sprintf("%sweeklyinsightsmailer/%v/q-%v/k-%v/", path, dateString, queryId, k)
}

func (sd *S3Driver) GetAdsDataDir(projectId int64) string {
	path := sd.GetProjectDir(projectId)
	return fmt.Sprintf("%sAdsImport/", path)
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
	path = pb.Join(path, U.DataTypeEvent)
	model_str := fmt.Sprintf("%d", model_id)
	return pb.Join(path, "predict", model_str)
}

func (sd *S3Driver) GetWIPropertiesPathAndName(projectId int64) (string, string) {
	path := sd.GetWIPropertiesDir(projectId)
	return path, "properties.txt"
}

func (sd *S3Driver) GetWIPropertiesDir(projectId int64) string {
	path := sd.GetProjectDir(projectId)
	return fmt.Sprintf("%sweeklyinsights/", path)
}

func (sd *S3Driver) GetEventsUnsortedFilePathAndName(projectId int64, startTimestamp int64, endTimestamp int64) (string, string) {
	path, name := sd.GetEventsFilePathAndName(projectId, startTimestamp, endTimestamp)
	fileName := "unsorted_" + name
	return path, fileName
}

func (sd *S3Driver) GetModelArtifactsPath(projectId int64, modelId uint64) string {
	path := sd.GetProjectModelDir(projectId, modelId)
	path = pb.Join(path, "artifacts")
	return path
}

func (sd *S3Driver) GetEventsArtifactFilePathAndName(projectId int64, startTimestamp int64, endTimestamp int64) (string, string) {
	path := sd.GetProjectDataFileDir(projectId, startTimestamp, U.DataTypeEvent)
	path = pb.Join(path, "artifacts")
	dateFormattedEnd := U.GetDateOnlyFromTimestampZ(endTimestamp)
	fileName := fmt.Sprintf("users_map_%s.txt", dateFormattedEnd)
	return path, fileName
}

func (sd *S3Driver) GetChannelArtifactFilePathAndName(channel string, projectId int64, startTimestamp int64, endTimestamp int64) (string, string) {
	path := sd.GetProjectDataFileDir(projectId, startTimestamp, U.DataTypeAdReport)
	path = pb.Join(path, "artifacts")
	dateFormattedEnd := U.GetDateOnlyFromTimestampZ(endTimestamp)
	fileName := fmt.Sprintf("doctypes_map_%s_%s.txt", dateFormattedEnd, channel)
	return path, fileName
}

func (sd *S3Driver) GetPathAnalysisTempFileDir(id string, projectId int64) string {
	path := sd.GetProjectDir(projectId)
	return fmt.Sprintf("%spathanalysis/%v/", path, id)
}
func (sd *S3Driver) GetPathAnalysisTempFilePathAndName(id string, projectId int64) (string, string) {
	path := sd.GetPathAnalysisTempFileDir(id, projectId)
	return path, "patterns.txt"
}

func (sd *S3Driver) GetExplainV2Dir(id uint64, projectId int64) string {
	return fmt.Sprintf("projects/%d/explain/models/%d/", projectId, id)
}

func (sd *S3Driver) GetExplainV2ModelPath(id uint64, projectId int64) (string, string) {
	path := sd.GetExplainV2Dir(id, projectId)
	chunksPath := pb.Join(path, "chunks")
	return chunksPath, "chunk_1.txt"
}
