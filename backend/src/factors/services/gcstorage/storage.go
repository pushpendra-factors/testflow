package gcstorage

import (
	"context"
	"factors/filestore"
	U "factors/util"
	"fmt"
	"io"

	pb "path/filepath"
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
	if !strings.HasSuffix(dir, separator) {
		// Append / to the end if not present.
		dir = dir + separator
	}
	obj := gcsd.client.Bucket(gcsd.BucketName).Object(dir + fileName)

	writer := struct {
		ReadFrom int // to "disable" ReadFrom method
		*storage.Writer
	}{0, obj.NewWriter(ctx)}

	buf := make([]byte, 32*1024)
	if _, err := io.CopyBuffer(writer, reader, buf); err != nil {
		return err
	}

	err := writer.Close()
	return err
}

func (gcsd *GCSDriver) GetWriter(dir, fileName string) (io.WriteCloser, error) {
	ctx := context.Background()
	if !strings.HasSuffix(dir, separator) {
		// Append / to the end if not present.
		dir = dir + separator
	}
	obj := gcsd.client.Bucket(gcsd.BucketName).Object(dir + fileName)

	writer := struct {
		ReadFrom int // to "disable" ReadFrom method
		*storage.Writer
	}{0, obj.NewWriter(ctx)}

	return writer, nil
}

func (gcsd *GCSDriver) Get(dir, fileName string) (io.ReadCloser, error) {
	ctx := context.Background()
	if !strings.HasSuffix(dir, separator) {
		// Append / to the end if not present.
		dir = dir + separator
	}
	obj := gcsd.client.Bucket(gcsd.BucketName).Object(dir + fileName)
	rc, err := obj.NewReader(ctx)

	return rc, err
}

func (gcsd *GCSDriver) GetObjectSize(dir, fileName string) (int64, error) {
	ctx := context.Background()
	if !strings.HasSuffix(dir, separator) {
		// Append / to the end if not present.
		dir = dir + separator
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

// ListFiles List files present in a folder in cloud storage. Prefix has to be without bucket name.
// Must not have leading '/' and should have trailing '/' in prefix. Ex: archive/3/.
func (gcsd *GCSDriver) ListFiles(prefix string) []string {
	var files []string
	if !strings.HasSuffix(prefix, separator) {
		prefix = prefix + separator
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
		} else if attributes.Name == prefix || attributes.Name == (prefix+separator) {
			// Omit the base prefix if returned as one the objects.
			continue
		}
		files = append(files, attributes.Name)
	}

	return files
}

func (gcsd *GCSDriver) GetBucketName() string {
	return gcsd.BucketName
}

func (gcsd *GCSDriver) GetProjectDir(projectId int64) string {
	return fmt.Sprintf("projects/%d/", projectId)
}

func (gcsd *GCSDriver) GetProjectModelDir(projectId int64, modelId uint64) string {
	path := gcsd.GetProjectDir(projectId)
	return fmt.Sprintf("%smodels/%d/", path, modelId)
}

func (gcsd *GCSDriver) GetProjectDataFileDir(projectId int64, startTimestamp int64, dataType, modelType string) string {
	if gcsd.BucketName == "factors-production-v3" || gcsd.BucketName == "factors-staging-v3" {
		dateFormatted := U.GetDateOnlyFromTimestampZ(startTimestamp)
		return fmt.Sprintf("projects/%d/events/%s/%s/", projectId, modelType, dateFormatted)
	} else {
		path := gcsd.GetProjectDir(projectId)
		dateFormatted := U.GetDateOnlyFromTimestampZ(startTimestamp)
		return fmt.Sprintf("%s%s/%s/", path, dataType, dateFormatted)
	}
}

func (gcsd *GCSDriver) GetModelUserPropertiesCategoricalFilePathAndName(projectId int64, modelId uint64) (string, string) {
	path := gcsd.GetProjectModelDir(projectId, modelId) + "properties/"
	return path, fmt.Sprintf("userPropCatgMap_%d.txt", modelId)
}

func (gcsd *GCSDriver) GetModelEventPropertiesCategoricalFilePathAndName(projectId int64, modelId uint64) (string, string) {
	path := gcsd.GetProjectModelDir(projectId, modelId) + "properties/"
	return path, fmt.Sprintf("eventPropCatgMap_%d.txt", modelId)
}

func (gcsd *GCSDriver) GetModelUserPropertiesFilePathAndName(projectId int64, modelId uint64) (string, string) {
	path := gcsd.GetProjectModelDir(projectId, modelId) + "properties/"
	return path, fmt.Sprintf("eventUserPropMap_%d.txt", modelId)
}

func (gcsd *GCSDriver) GetModelEventPropertiesFilePathAndName(projectId int64, modelId uint64) (string, string) {
	path := gcsd.GetProjectModelDir(projectId, modelId) + "properties/"
	return path, fmt.Sprintf("eventEventPropMap_%d.txt", modelId)
}

func (gcsd *GCSDriver) GetModelEventInfoFilePathAndName(projectId int64, modelId uint64) (string, string) {
	path := gcsd.GetProjectModelDir(projectId, modelId)
	return path, fmt.Sprintf("event_info_%d.txt", modelId)
}

func (gcsd *GCSDriver) GetEventsFilePathAndName(projectId int64, startTimestamp, endTimestamp int64) (string, string) {
	var fileName string
	modelType := U.GetModelType(startTimestamp, endTimestamp)
	path := gcsd.GetProjectDataFileDir(projectId, startTimestamp, U.DataTypeEvent, modelType)
	if gcsd.BucketName == "factors-production-v3" || gcsd.BucketName == "factors-staging-v3" {
		fileName = "events.txt"
	} else {
		dateFormattedStart := U.GetDateOnlyFromTimestampZ(startTimestamp)
		dateFormattedEnd := U.GetDateOnlyFromTimestampZ(endTimestamp)
		fileName = fmt.Sprintf("events_%s-%s.txt", dateFormattedStart, dateFormattedEnd)
	}
	return path, fileName
}

func (gcsd *GCSDriver) GetEventsGroupFilePathAndName(projectId int64, startTimestamp, endTimestamp int64, group int) (string, string) {
	if group == 0 {
		return gcsd.GetEventsFilePathAndName(projectId, startTimestamp, endTimestamp)
	}
	var fileName string
	modelType := U.GetModelType(startTimestamp, endTimestamp)
	path := gcsd.GetProjectDataFileDir(projectId, startTimestamp, U.DataTypeEvent, modelType)
	dateFormattedStart := U.GetDateOnlyFromTimestampZ(startTimestamp)
	dateFormattedEnd := U.GetDateOnlyFromTimestampZ(endTimestamp)
	fileName = fmt.Sprintf("events_group%d_%s-%s.txt", group, dateFormattedStart, dateFormattedEnd)
	return path, fileName
}

func (gcsd *GCSDriver) GetChannelFilePathAndName(channel string, projectId int64, startTimestamp, endTimestamp int64) (string, string) {
	var fileName string
	modelType := U.GetModelType(startTimestamp, endTimestamp)
	path := gcsd.GetProjectDataFileDir(projectId, startTimestamp, U.DataTypeAdReport, modelType)
	if gcsd.BucketName == "factors-production-v3" || gcsd.BucketName == "factors-staging-v3" {
		fileName = channel + ".txt"
	} else {
		dateFormattedStart := U.GetDateOnlyFromTimestampZ(startTimestamp)
		dateFormattedEnd := U.GetDateOnlyFromTimestampZ(endTimestamp)
		fileName = fmt.Sprintf("%s_%s-%s.txt", channel, dateFormattedStart, dateFormattedEnd)
	}
	return path, fileName
}

func (gcsd *GCSDriver) GetUsersFilePathAndName(dateField string, projectId int64, startTimestamp, endTimestamp int64) (string, string) {
	var fileName string
	modelType := U.GetModelType(startTimestamp, endTimestamp)
	path := gcsd.GetProjectDataFileDir(projectId, startTimestamp, U.DataTypeUser, modelType)
	if gcsd.BucketName == "factors-production-v3" || gcsd.BucketName == "factors-staging-v3" {
		path = pb.Join(path, "users")
		fileName = dateField + ".txt"
	} else {
		dateFormattedStart := U.GetDateOnlyFromTimestampZ(startTimestamp)
		dateFormattedEnd := U.GetDateOnlyFromTimestampZ(endTimestamp)
		fileName = fmt.Sprintf("%s_%s-%s.txt", dateField, dateFormattedStart, dateFormattedEnd)
	}
	return path, fileName
}

func (gcsd *GCSDriver) GetModelMetricsFilePathAndName(projectId int64, startTimestamp, endTimestamp int64) (string, string) {
	var fileName string
	modelType := U.GetModelType(startTimestamp, endTimestamp)
	path := gcsd.GetProjectDataFileDir(projectId, startTimestamp, "metrics", modelType)
	if gcsd.BucketName == "factors-production-v3" || gcsd.BucketName == "factors-staging-v3" {
		fileName = "metrics.txt"
	} else {
		dateFormattedStart := U.GetDateOnlyFromTimestampZ(startTimestamp)
		dateFormattedEnd := U.GetDateOnlyFromTimestampZ(endTimestamp)
		fileName = fmt.Sprintf("metrics_%s-%s.txt", dateFormattedStart, dateFormattedEnd)
	}
	return path, fileName
}

func (gcsd *GCSDriver) GetModelAlertsFilePathAndName(projectId int64, startTimestamp, endTimestamp int64) (string, string) {
	var fileName string
	modelType := U.GetModelType(startTimestamp, endTimestamp)
	path := gcsd.GetProjectDataFileDir(projectId, startTimestamp, "alerts", modelType)
	if gcsd.BucketName == "factors-production-v3" || gcsd.BucketName == "factors-staging-v3" {
		fileName = "alerts.txt"
	} else {
		dateFormattedStart := U.GetDateOnlyFromTimestampZ(startTimestamp)
		dateFormattedEnd := U.GetDateOnlyFromTimestampZ(endTimestamp)
		fileName = fmt.Sprintf("alerts_%s-%s.txt", dateFormattedStart, dateFormattedEnd)
	}
	return path, fileName
}

func (gcsd *GCSDriver) GetModelEventsBucketingFilePathAndName(projectId int64, startTimestamp, endTimestamp int64) (string, string) {
	var fileName string
	modelType := U.GetModelType(startTimestamp, endTimestamp)
	path := gcsd.GetProjectDataFileDir(projectId, startTimestamp, U.DataTypeEvent, modelType)
	if gcsd.BucketName == "factors-production-v3" || gcsd.BucketName == "factors-staging-v3" {
		fileName = "events_bucketed.txt"
	} else {
		path = pb.Join(path, "events_bucketed")
		dateFormattedStart := U.GetDateOnlyFromTimestampZ(startTimestamp)
		dateFormattedEnd := U.GetDateOnlyFromTimestampZ(endTimestamp)
		fileName = fmt.Sprintf("events_bucketed_%s-%s.txt", dateFormattedStart, dateFormattedEnd)
	}
	return path, fileName
}

// Buckets
// If we have two different files for last week and this week, we might end up having non-overlapping ranges.
// So keeping one for this week and other as master. If it exists in master pick that else this week, or compute now

func (gcsd *GCSDriver) GetMasterNumericalBucketsFile(projectId int64) (string, string) {
	path := gcsd.GetProjectDir(projectId)
	path = pb.Join(path, U.DataTypeEvent)
	return path, "numerical_buckets_master.txt"
}

func (gcsd *GCSDriver) GetModelEventsNumericalBucketsFile(projectId int64, startTimestamp, endTimestamp int64) (string, string) {
	var fileName string
	modelType := U.GetModelType(startTimestamp, endTimestamp)
	path := gcsd.GetProjectDataFileDir(projectId, startTimestamp, U.DataTypeEvent, modelType)
	if gcsd.BucketName == "factors-production-v3" || gcsd.BucketName == "factors-staging-v3" {
		fileName = "numerical_buckets.txt"
	} else {
		path = pb.Join(path, "numerical_buckets")
		dateFormattedStart := U.GetDateOnlyFromTimestampZ(startTimestamp)
		dateFormattedEnd := U.GetDateOnlyFromTimestampZ(endTimestamp)
		fileName = fmt.Sprintf("numerical_buckets_%s-%s.txt", dateFormattedStart, dateFormattedEnd)
	}
	return path, fileName
}

func (gcsd *GCSDriver) GetPatternChunksDir(projectId int64, modelId uint64) string {
	modelDir := gcsd.GetProjectModelDir(projectId, modelId)
	return fmt.Sprintf("%schunks/", modelDir)
}
func (gcsd *GCSDriver) GetChunksMetaDataDir(projectId int64, modelId uint64) string {
	modelDir := gcsd.GetProjectModelDir(projectId, modelId)
	return fmt.Sprintf("%smetadata/", modelDir)
}
func (gcsd *GCSDriver) GetPatternChunkFilePathAndName(projectId int64, modelId uint64, chunkId string) (string, string) {
	return gcsd.GetPatternChunksDir(projectId, modelId), fmt.Sprintf("chunk_%s.txt", chunkId)
}
func (gcsd *GCSDriver) GetChunksMetaDataFilePathAndName(projectId int64, modelId uint64) (string, string) {
	return gcsd.GetChunksMetaDataDir(projectId, modelId), "metadata.txt"
}
func (gcsd *GCSDriver) GetEventArchiveFilePathAndName(projectID int64, startTime, endTime int64) (string, string) {
	year, month, date := time.Unix(startTime, 0).UTC().Date()
	path := fmt.Sprintf("archive/%d/%d/%d/", projectID, year, int(month))
	fileName := fmt.Sprintf("%d_events_%d-%d.txt", date, startTime, endTime)
	return path, fileName
}

func (gcsd *GCSDriver) GetUsersArchiveFilePathAndName(projectID int64, startTime, endTime int64) (string, string) {
	year, month, date := time.Unix(startTime, 0).UTC().Date()
	path := fmt.Sprintf("archive/%d/%d/%d/", projectID, year, int(month))
	fileName := fmt.Sprintf("%d_users_%d-%d.txt", date, startTime, endTime)
	return path, fileName
}

func (gcsd *GCSDriver) GetDailyArchiveFilesDir(projectID int64, dataTimestamp int64, dataType string) string {
	dateFormatted := U.GetDateOnlyFromTimestampZ(dataTimestamp)
	path := fmt.Sprintf("daily_pull/%d/%s/%s/", projectID, dateFormatted, dataType)
	return path
}

func (gcsd *GCSDriver) GetDailyEventArchiveFilePathAndName(projectID int64, dataTimestamp int64, startTime, endTime int64) (string, string) {
	path := gcsd.GetDailyArchiveFilesDir(projectID, dataTimestamp, U.DataTypeEvent)
	fileName := fmt.Sprintf("events_created_at_%d-%d.txt", startTime, endTime)
	return path, fileName
}

func (gcsd *GCSDriver) GetDailyUsersArchiveFilePathAndName(dateField string, projectID int64, dataTimestamp int64, startTime, endTime int64) (string, string) {
	path := gcsd.GetDailyArchiveFilesDir(projectID, dataTimestamp, U.DataTypeUser)
	fileName := fmt.Sprintf("%s_created_at_%d-%d.txt", dateField, startTime, endTime)
	return path, fileName
}

func (gcsd *GCSDriver) GetDailyChannelArchiveFilePathAndName(channel string, projectID int64, dataTimestamp int64, startTime, endTime int64) (string, string) {
	path := gcsd.GetDailyArchiveFilesDir(projectID, dataTimestamp, U.DataTypeAdReport)
	fileName := fmt.Sprintf("%s_created_at_%d-%d.txt", channel, startTime, endTime)
	return path, fileName
}

func (gcsd *GCSDriver) GetInsightsWpiFilePathAndName(projectId int64, dateString string, queryId int64, k int, mailerRun bool) (string, string) {
	path := ""
	if mailerRun {
		path = gcsd.GetWeeklyInsightsMailerModelDir(projectId, dateString, queryId, k)
	} else {
		path = gcsd.GetWeeklyInsightsModelDir(projectId, dateString, queryId, k)
	}
	return path, "wpi.txt"
}

func (gcsd *GCSDriver) GetInsightsCpiFilePathAndName(projectId int64, dateString string, queryId int64, k int, mailerRun bool) (string, string) {
	path := ""
	if mailerRun {
		path = gcsd.GetWeeklyInsightsMailerModelDir(projectId, dateString, queryId, k)
	} else {
		path = gcsd.GetWeeklyInsightsModelDir(projectId, dateString, queryId, k)
	}
	return path, "cpi.txt"
}

func (gcsd *GCSDriver) GetWIPropertiesDir(projectId int64) string {
	path := gcsd.GetProjectDir(projectId)
	return fmt.Sprintf("%sweeklyinsights/", path)
}

func (gcsd *GCSDriver) GetWIPropertiesPathAndName(projectId int64) (string, string) {
	path := gcsd.GetWIPropertiesDir(projectId)
	return path, "properties.txt"
}

func (gcsd *GCSDriver) GetWeeklyInsightsModelDir(projectId int64, dateString string, queryId int64, k int) string {
	path := gcsd.GetWIPropertiesDir(projectId)
	return fmt.Sprintf("%s%v/q-%v/k-%v/", path, dateString, queryId, k)
}

func (gcsd *GCSDriver) GetWeeklyInsightsMailerModelDir(projectId int64, dateString string, queryId int64, k int) string {
	path := gcsd.GetProjectDir(projectId)
	return fmt.Sprintf("%sweeklyinsightsmailer/%v/q-%v/k-%v/", path, dateString, queryId, k)
}

func (gcsd *GCSDriver) GetAdsDataDir(projectId int64) string {
	path := gcsd.GetProjectDir(projectId)
	return pb.Join(path, "AdsImport")
}

func (gcsd *GCSDriver) GetAdsDataFilePathAndName(projectId int64, report string, chunkNo int) (string, string) {
	path := gcsd.GetAdsDataDir(projectId)
	return path, fmt.Sprintf("%v-%v-%v.csv", report, projectId, chunkNo)
}

func (gcsd *GCSDriver) GetPredictProjectDataPath(projectId int64, model_id int64) string {
	path := gcsd.GetPredictProjectDir(projectId, model_id)
	return pb.Join(path, "data")
}

func (gcsd *GCSDriver) GetPredictProjectDir(projectId int64, model_id int64) string {
	path := gcsd.GetProjectDir(projectId)
	path = pb.Join(path, U.DataTypeEvent)
	model_str := fmt.Sprintf("%d", model_id)
	return pb.Join(path, "predict", model_str)
}

func (gcsd *GCSDriver) GetEventsUnsortedFilePathAndName(projectId int64, startTimestamp int64, endTimestamp int64) (string, string) {
	path, name := gcsd.GetEventsFilePathAndName(projectId, startTimestamp, endTimestamp)
	fileName := "unsorted_" + name
	return path, fileName
}

func (gcsd *GCSDriver) GetModelArtifactsPath(projectId int64, modelId uint64) string {
	path := gcsd.GetProjectModelDir(projectId, modelId)
	path = pb.Join(path, "artifacts")
	return path
}

func (gcsd *GCSDriver) GetEventsArtifactFilePathAndName(projectId int64, startTimestamp int64, endTimestamp int64, group int) (string, string) {
	fileName := "users_map.txt"
	var path string
	if gcsd.BucketName == "factors-production-v3" || gcsd.BucketName == "factors-staging-v3" {
		modelType := U.GetModelType(startTimestamp, endTimestamp)
		path = gcsd.GetProjectDataFileDir(projectId, startTimestamp, U.DataTypeEvent, modelType)
	} else {
		path = gcsd.GetEventsTempFilesDir(projectId, startTimestamp, endTimestamp, group)
	}
	return path, fileName
}

func (gcsd *GCSDriver) GetChannelArtifactFilePathAndName(channel string, projectId int64, startTimestamp int64, endTimestamp int64) (string, string) {
	fileName := "doctypes_map.txt"
	var path string
	if gcsd.BucketName == "factors-production-v3" || gcsd.BucketName == "factors-staging-v3" {
		modelType := U.GetModelType(startTimestamp, endTimestamp)
		path = gcsd.GetProjectDataFileDir(projectId, startTimestamp, U.DataTypeAdReport, modelType)
	} else {
		path = gcsd.GetChannelTempFilesDir(channel, projectId, startTimestamp, endTimestamp)
	}
	return path, fileName
}

func (gcsd *GCSDriver) GetPathAnalysisTempFileDir(id string, projectId int64) string {
	path := gcsd.GetProjectDir(projectId)
	return fmt.Sprintf("%spathanalysis/%v/", path, id)
}
func (gcsd *GCSDriver) GetPathAnalysisTempFilePathAndName(id string, projectId int64) (string, string) {
	path := gcsd.GetPathAnalysisTempFileDir(id, projectId)
	return path, "patterns.txt"
}

func (gcsd *GCSDriver) GetExplainV2Dir(id uint64, projectId int64) string {
	return fmt.Sprintf("projects/%v/explain/models/%v/", projectId, id)
}

func (gcsd *GCSDriver) GetExplainV2ModelPath(id uint64, projectId int64) (string, string) {
	path := gcsd.GetExplainV2Dir(id, projectId)
	chunksPath := pb.Join(path, "chunks")
	log.Infof("Getting explain model from (gcp): %s", chunksPath)
	return chunksPath, "chunk_1.txt"
}

func (gcsd *GCSDriver) GetListReferenceFileNameAndPathFromCloud(projectID int64, reference string) (string, string) {
	return fmt.Sprintf("projects/%v/list/%v/", projectID, reference), "list.txt"
}
func (gcsd *GCSDriver) GetSixSignalAnalysisTempFileDir(id string, projectId int64) string {
	path := gcsd.GetProjectDir(projectId)
	return fmt.Sprintf("%ssixSignal/%v/", path, id)
}

func (gcsd *GCSDriver) GetSixSignalAnalysisTempFilePathAndName(id string, projectId int64) (string, string) {
	path := gcsd.GetSixSignalAnalysisTempFileDir(id, projectId)
	return path, "results.txt"
}

func (gcsd *GCSDriver) GetAccScoreDir(projectId int64) string {
	proj_path := gcsd.GetProjectDir(projectId)
	path := fmt.Sprintf("%saccscore", proj_path)
	return path
}

func (gcsd *GCSDriver) GetAccScoreUsers(projectId int64) string {
	dirPath := gcsd.GetAccScoreDir(projectId)
	path := fmt.Sprintf("%s/users/", dirPath)
	return path
}

func (gcsd *GCSDriver) GetAccScoreAccounts(projectId int64) string {
	dirPath := gcsd.GetAccScoreDir(projectId)
	path := fmt.Sprintf("%s/groups/", dirPath)
	return path
}
func (gcsd *GCSDriver) GetEventsTempFilesDir(projectId int64, startTimestamp, endTimestamp int64, group int) string {
	path, name := gcsd.GetEventsGroupFilePathAndName(projectId, startTimestamp, endTimestamp, group)
	path = pb.Join(path, strings.Replace(name, ".txt", "", 1))
	return path
}

func (gcsd *GCSDriver) GetEventsPartFilesDir(projectId int64, startTimestamp, endTimestamp int64, sorted bool, group int) string {
	tmp := "unsorted"
	if sorted {
		tmp = "sorted"
	}
	path := gcsd.GetEventsTempFilesDir(projectId, startTimestamp, endTimestamp, group)
	path = pb.Join(path, tmp+"_parts")
	return path
}

func (gcsd *GCSDriver) GetEventsPartFilePathAndName(projectId int64, startTimestamp, endTimestamp int64, sorted bool, startIndex int, endIndex int, group int) (string, string) {
	path := gcsd.GetEventsPartFilesDir(projectId, startTimestamp, endTimestamp, sorted, group)
	name := fmt.Sprintf("%d-%d_uids.txt", startIndex, endIndex)
	return path, name
}

func (gcsd *GCSDriver) GetChannelTempFilesDir(channel string, projectId int64, startTimestamp, endTimestamp int64) string {
	path, name := gcsd.GetChannelFilePathAndName(channel, projectId, startTimestamp, endTimestamp)
	path = pb.Join(path, strings.Replace(name, ".txt", "", 1))
	return path
}

func (gcsd *GCSDriver) GetChannelPartFilesDir(channel string, projectId int64, startTimestamp, endTimestamp int64, sorted bool) string {
	tmp := "unsorted"
	if sorted {
		tmp = "sorted"
	}
	path := gcsd.GetChannelTempFilesDir(channel, projectId, startTimestamp, endTimestamp)
	path = pb.Join(path, tmp+"_parts")
	return path
}

func (gcsd *GCSDriver) GetChannelPartFilePathAndName(channel string, projectId int64, startTimestamp, endTimestamp int64, sorted bool, index int) (string, string) {
	path := gcsd.GetChannelPartFilesDir(channel, projectId, startTimestamp, endTimestamp, sorted)
	name := fmt.Sprintf("%d_doctype.txt", index)
	return path, name
}
