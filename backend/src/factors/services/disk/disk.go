package disk

import (
	"factors/filestore"
	U "factors/util"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	pb "path/filepath"
	"strings"
	"time"

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

func (dd *DiskDriver) Create(path, fileName string, reader io.Reader) error {
	err := MkdirAll(path)
	if err != nil {
		log.WithError(err).Errorln("Failed to create dir")
		return err
	}

	if !strings.HasSuffix(path, separator) {
		// Append / to the end if not present.
		path = path + separator
	}
	file, err := os.Create(path + fileName)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, reader)
	return err
}

func (dd *DiskDriver) GetWriter(path, fileName string) (io.WriteCloser, error) {
	err := MkdirAll(path)
	if err != nil {
		log.WithError(err).Errorln("Failed to create dir")
		return nil, err
	}

	if !strings.HasSuffix(path, separator) {
		// Append / to the end if not present.
		path = path + separator
	}
	file, err := os.Create(path + fileName)
	if err != nil {
		return nil, err
	}
	return file, nil
}

// Get opens a file in read only mode.
// Caller should take care of closing the returned io.ReadCloser.
func (dd *DiskDriver) Get(path, fileName string) (io.ReadCloser, error) {
	log.WithFields(log.Fields{
		"Path":     path,
		"FileName": fileName,
	}).Debug("DiskDriver Opening file")

	if !strings.HasSuffix(path, separator) {
		// Append / to the end if not present.
		path = path + separator
	}
	file, err := os.OpenFile(path+fileName, os.O_RDONLY, 0444)
	return file, err
}

func (dd *DiskDriver) GetBucketName() string {
	return dd.baseDir
}

func (dd *DiskDriver) GetObjectSize(path, fileName string) (int64, error) {
	if !strings.HasSuffix(path, separator) {
		// Append / to the end if not present.
		path = path + separator
	}
	var objInfo os.FileInfo
	var err error
	if objInfo, err = os.Stat(path + fileName); err != nil {
		return 0, err
	}
	objSize := objInfo.Size()
	return objSize, err
}

// ListFiles List files present in a directory.
func (dd *DiskDriver) ListFiles(path string) []string {
	var files []string
	fileObjects, err := ioutil.ReadDir(path)
	if err != nil {
		log.WithError(err).Errorln("Failed to read directory contents")
		return files
	}

	for _, file := range fileObjects {
		files = append(files, path+"/"+file.Name())
	}
	return files
}

func (dd *DiskDriver) GetProjectDir(projectId int64) string {
	return fmt.Sprintf("%s/projects/%d/", dd.baseDir, projectId)
}

func (dd *DiskDriver) GetProjectModelDir(projectId int64, modelId uint64) string {
	path := dd.GetProjectDir(projectId)
	return fmt.Sprintf("%smodels/%d/", path, modelId)
}

func (dd *DiskDriver) GetProjectDataFileDir(projectId int64, startTimestamp int64, dataType, modelType string) string {
	pathArr := strings.Split(dd.baseDir, "/")
	folderName := pathArr[len(pathArr)-1]
	if folderName == "cloud_storage" {
		dateFormatted := U.GetDateOnlyFromTimestampZ(startTimestamp)
		return fmt.Sprintf("projects/%d/events/%s/%s/", projectId, modelType, dateFormatted)
	} else {
		path := dd.GetProjectDir(projectId)
		dateFormatted := U.GetDateOnlyFromTimestampZ(startTimestamp)
		return fmt.Sprintf("%s%s/%s/", path, dataType, dateFormatted)
	}
}

func (dd *DiskDriver) GetModelUserPropertiesCategoricalFilePathAndName(projectId int64, modelId uint64) (string, string) {
	path := dd.GetProjectModelDir(projectId, modelId) + "properties/"
	return path, fmt.Sprintf("userPropCatgMap_%d.txt", modelId)
}

func (dd *DiskDriver) GetModelEventPropertiesCategoricalFilePathAndName(projectId int64, modelId uint64) (string, string) {
	path := dd.GetProjectModelDir(projectId, modelId) + "properties/"
	return path, fmt.Sprintf("eventPropCatgMap_%d.txt", modelId)
}

func (dd *DiskDriver) GetModelUserPropertiesFilePathAndName(projectId int64, modelId uint64) (string, string) {
	path := dd.GetProjectModelDir(projectId, modelId) + "properties/"
	return path, fmt.Sprintf("eventUserPropMap_%d.txt", modelId)
}

func (dd *DiskDriver) GetModelEventPropertiesFilePathAndName(projectId int64, modelId uint64) (string, string) {
	path := dd.GetProjectModelDir(projectId, modelId) + "properties/"
	return path, fmt.Sprintf("eventEventPropMap_%d.txt", modelId)
}

func (dd *DiskDriver) GetModelEventInfoFilePathAndName(projectId int64, modelId uint64) (string, string) {
	path := dd.GetProjectModelDir(projectId, modelId)
	return path, fmt.Sprintf("event_info_%d.txt", modelId)
}

func (dd *DiskDriver) GetEventsFilePathAndName(projectId int64, startTimestamp, endTimestamp int64) (string, string) {
	var fileName string
	modelType := U.GetModelType(startTimestamp, endTimestamp)
	path := dd.GetProjectDataFileDir(projectId, startTimestamp, U.DataTypeEvent, modelType)
	pathArr := strings.Split(dd.baseDir, "/")
	folderName := pathArr[len(pathArr)-1]
	if folderName == "cloud_storage" {
		fileName = "events.txt"
	} else {
		dateFormattedStart := U.GetDateOnlyFromTimestampZ(startTimestamp)
		dateFormattedEnd := U.GetDateOnlyFromTimestampZ(endTimestamp)
		fileName = fmt.Sprintf("events_%s-%s.txt", dateFormattedStart, dateFormattedEnd)
	}
	return path, fileName
}

func (dd *DiskDriver) GetModelArtifactsPath(projectId int64, modelId uint64) string {
	path := dd.GetProjectModelDir(projectId, modelId)
	path = pb.Join(path, "artifacts")
	return path
}

func (dd *DiskDriver) GetEventsGroupFilePathAndName(projectId int64, startTimestamp, endTimestamp int64, group int) (string, string) {
	if group == 0 {
		return dd.GetEventsFilePathAndName(projectId, startTimestamp, endTimestamp)
	}
	var fileName string
	modelType := U.GetModelType(startTimestamp, endTimestamp)
	path := dd.GetProjectDataFileDir(projectId, startTimestamp, U.DataTypeEvent, modelType)
	dateFormattedStart := U.GetDateOnlyFromTimestampZ(startTimestamp)
	dateFormattedEnd := U.GetDateOnlyFromTimestampZ(endTimestamp)
	fileName = fmt.Sprintf("events_group%d_%s-%s.txt", group, dateFormattedStart, dateFormattedEnd)
	return path, fileName
}

func (dd *DiskDriver) GetChannelFilePathAndName(channel string, projectId int64, startTimestamp, endTimestamp int64) (string, string) {
	var fileName string
	modelType := U.GetModelType(startTimestamp, endTimestamp)
	path := dd.GetProjectDataFileDir(projectId, startTimestamp, U.DataTypeAdReport, modelType)
	pathArr := strings.Split(dd.baseDir, "/")
	folderName := pathArr[len(pathArr)-1]
	if folderName == "cloud_storage" {
		fileName = channel + ".txt"
	} else {
		dateFormattedStart := U.GetDateOnlyFromTimestampZ(startTimestamp)
		dateFormattedEnd := U.GetDateOnlyFromTimestampZ(endTimestamp)
		fileName = fmt.Sprintf("%s_%s-%s.txt", channel, dateFormattedStart, dateFormattedEnd)
	}
	return path, fileName
}

func (dd *DiskDriver) GetUsersFilePathAndName(dateField string, projectId int64, startTimestamp, endTimestamp int64) (string, string) {
	var fileName string
	modelType := U.GetModelType(startTimestamp, endTimestamp)
	path := dd.GetProjectDataFileDir(projectId, startTimestamp, U.DataTypeUser, modelType)
	pathArr := strings.Split(dd.baseDir, "/")
	folderName := pathArr[len(pathArr)-1]
	if folderName == "cloud_storage" {
		path = pb.Join(path, "users")
		fileName = dateField + ".txt"
	} else {
		dateFormattedStart := U.GetDateOnlyFromTimestampZ(startTimestamp)
		dateFormattedEnd := U.GetDateOnlyFromTimestampZ(endTimestamp)
		fileName = fmt.Sprintf("%s_%s-%s.txt", dateField, dateFormattedStart, dateFormattedEnd)
	}
	return path, fileName
}

func (dd *DiskDriver) GetModelMetricsFilePathAndName(projectId int64, startTimestamp, endTimestamp int64) (string, string) {
	var fileName string
	modelType := U.GetModelType(startTimestamp, endTimestamp)
	path := dd.GetProjectDataFileDir(projectId, startTimestamp, "metrics", modelType)
	pathArr := strings.Split(dd.baseDir, "/")
	folderName := pathArr[len(pathArr)-1]
	if folderName == "cloud_storage" {
		fileName = "metrics.txt"
	} else {
		dateFormattedStart := U.GetDateOnlyFromTimestampZ(startTimestamp)
		dateFormattedEnd := U.GetDateOnlyFromTimestampZ(endTimestamp)
		fileName = fmt.Sprintf("metrics_%s-%s.txt", dateFormattedStart, dateFormattedEnd)
	}
	return path, fileName
}

func (dd *DiskDriver) GetModelAlertsFilePathAndName(projectId int64, startTimestamp, endTimestamp int64) (string, string) {
	var fileName string
	modelType := U.GetModelType(startTimestamp, endTimestamp)
	path := dd.GetProjectDataFileDir(projectId, startTimestamp, "alerts", modelType)
	pathArr := strings.Split(dd.baseDir, "/")
	folderName := pathArr[len(pathArr)-1]
	if folderName == "cloud_storage" {
		fileName = "alerts.txt"
	} else {
		dateFormattedStart := U.GetDateOnlyFromTimestampZ(startTimestamp)
		dateFormattedEnd := U.GetDateOnlyFromTimestampZ(endTimestamp)
		fileName = fmt.Sprintf("alerts_%s-%s.txt", dateFormattedStart, dateFormattedEnd)
	}
	return path, fileName
}

func (dd *DiskDriver) GetModelEventsBucketingFilePathAndName(projectId int64, startTimestamp, endTimestamp int64) (string, string) {
	var fileName string
	modelType := U.GetModelType(startTimestamp, endTimestamp)
	path := dd.GetProjectDataFileDir(projectId, startTimestamp, U.DataTypeEvent, modelType)
	pathArr := strings.Split(dd.baseDir, "/")
	folderName := pathArr[len(pathArr)-1]
	if folderName == "cloud_storage" {
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

func (dd *DiskDriver) GetMasterNumericalBucketsFile(projectId int64) (string, string) {
	path := dd.GetProjectDir(projectId)
	path = pb.Join(path, U.DataTypeEvent)
	return path, "numerical_buckets_master.txt"
}

func (dd *DiskDriver) GetModelEventsNumericalBucketsFile(projectId int64, startTimestamp, endTimestamp int64) (string, string) {
	var fileName string
	modelType := U.GetModelType(startTimestamp, endTimestamp)
	path := dd.GetProjectDataFileDir(projectId, startTimestamp, U.DataTypeEvent, modelType)
	pathArr := strings.Split(dd.baseDir, "/")
	folderName := pathArr[len(pathArr)-1]
	if folderName == "cloud_storage" {
		fileName = "numerical_buckets.txt"
	} else {
		path = pb.Join(path, "numerical_buckets")
		dateFormattedStart := U.GetDateOnlyFromTimestampZ(startTimestamp)
		dateFormattedEnd := U.GetDateOnlyFromTimestampZ(endTimestamp)
		fileName = fmt.Sprintf("numerical_buckets_%s-%s.txt", dateFormattedStart, dateFormattedEnd)
	}
	return path, fileName
}

func (dd *DiskDriver) GetPatternChunksDir(projectId int64, modelId uint64) string {
	modelDir := dd.GetProjectModelDir(projectId, modelId)
	return fmt.Sprintf("%schunks/", modelDir)
}
func (dd *DiskDriver) GetChunksMetaDataDir(projectId int64, modelId uint64) string {
	modelDir := dd.GetProjectModelDir(projectId, modelId)
	return fmt.Sprintf("%smetadata/", modelDir)
}
func (dd *DiskDriver) GetPatternChunkFilePathAndName(projectId int64, modelId uint64, chunkId string) (string, string) {
	return dd.GetPatternChunksDir(projectId, modelId), fmt.Sprintf("chunk_%s.txt", chunkId)
}
func (dd *DiskDriver) GetChunksMetaDataFilePathAndName(projectId int64, modelId uint64) (string, string) {
	return dd.GetChunksMetaDataDir(projectId, modelId), "metadata.txt"
}
func (dd *DiskDriver) GetEventArchiveFilePathAndName(projectID int64, startTime, endTime int64) (string, string) {
	year, month, date := time.Unix(startTime, 0).UTC().Date()
	path := fmt.Sprintf("%s/archive/%d/%d/%d/", dd.baseDir, projectID, year, int(month))
	fileName := fmt.Sprintf("%d_events_%d-%d.txt", date, startTime, endTime)
	return path, fileName
}

func (dd *DiskDriver) GetUsersArchiveFilePathAndName(projectID int64, startTime, endTime int64) (string, string) {
	year, month, date := time.Unix(startTime, 0).UTC().Date()
	path := fmt.Sprintf("%s/archive/%d/%d/%d/", dd.baseDir, projectID, year, int(month))
	fileName := fmt.Sprintf("%d_users_%d-%d.txt", date, startTime, endTime)
	return path, fileName
}

func (dd *DiskDriver) GetDailyArchiveFilesDir(projectID int64, dataTimestamp int64, dataType string) string {
	dateFormatted := U.GetDateOnlyFromTimestampZ(dataTimestamp)
	path := fmt.Sprintf("%s/daily_pull/%d/%s/%s/", dd.baseDir, projectID, dateFormatted, dataType)
	return path
}

func (dd *DiskDriver) GetDailyEventArchiveFilePathAndName(projectID int64, dataTimestamp int64, startTime, endTime int64) (string, string) {
	path := dd.GetDailyArchiveFilesDir(projectID, dataTimestamp, U.DataTypeEvent)
	fileName := fmt.Sprintf("events_created_at_%d-%d.txt", startTime, endTime)
	return path, fileName
}

func (dd *DiskDriver) GetDailyUsersArchiveFilePathAndName(dateField string, projectID int64, dataTimestamp int64, startTime, endTime int64) (string, string) {
	path := dd.GetDailyArchiveFilesDir(projectID, dataTimestamp, U.DataTypeUser)
	fileName := fmt.Sprintf("%s_created_at_%d-%d.txt", dateField, startTime, endTime)
	return path, fileName
}

func (dd *DiskDriver) GetDailyChannelArchiveFilePathAndName(channel string, projectID int64, dataTimestamp int64, startTime, endTime int64) (string, string) {
	path := dd.GetDailyArchiveFilesDir(projectID, dataTimestamp, U.DataTypeAdReport)
	fileName := fmt.Sprintf("%s_created_at_%d-%d.txt", channel, startTime, endTime)
	return path, fileName
}

func (dd *DiskDriver) GetInsightsWpiFilePathAndName(projectId int64, dateString string, queryId int64, k int, mailerRun bool) (string, string) {
	path := ""
	if mailerRun {
		path = dd.GetWeeklyInsightsMailerModelDir(projectId, dateString, queryId, k)
	} else {
		path = dd.GetWeeklyInsightsModelDir(projectId, dateString, queryId, k)
	}
	return path, "wpi.txt"
}

func (dd *DiskDriver) GetInsightsCpiFilePathAndName(projectId int64, dateString string, queryId int64, k int, mailerRun bool) (string, string) {
	path := ""
	if mailerRun {
		path = dd.GetWeeklyInsightsMailerModelDir(projectId, dateString, queryId, k)
	} else {
		path = dd.GetWeeklyInsightsModelDir(projectId, dateString, queryId, k)
	}
	return path, "cpi.txt"
}

func (dd *DiskDriver) GetWIPropertiesDir(projectId int64) string {
	path := dd.GetProjectDir(projectId)
	return fmt.Sprintf("%sweeklyinsights/", path)
}

func (dd *DiskDriver) GetWIPropertiesPathAndName(projectId int64) (string, string) {
	path := dd.GetWIPropertiesDir(projectId)
	return path, "properties.txt"
}

func (dd *DiskDriver) GetWeeklyInsightsModelDir(projectId int64, dateString string, queryId int64, k int) string {
	path := dd.GetWIPropertiesDir(projectId)
	return fmt.Sprintf("%s%v/q-%v/k-%v/", path, dateString, queryId, k)
}

func (dd *DiskDriver) GetWeeklyInsightsMailerModelDir(projectId int64, dateString string, queryId int64, k int) string {
	path := dd.GetProjectDir(projectId)
	return fmt.Sprintf("%sweeklyinsightsmailer/%v/q-%v/k-%v/", path, dateString, queryId, k)
}

func (dd *DiskDriver) GetAdsDataDir(projectId int64) string {
	path := dd.GetProjectDir(projectId)
	return pb.Join(path, "AdsImport")
}

func (dd *DiskDriver) GetAdsDataFilePathAndName(projectId int64, report string, chunkNo int) (string, string) {
	path := dd.GetAdsDataDir(projectId)
	return path, fmt.Sprintf("%v-%v-%v.csv", report, projectId, chunkNo)
}

func (dd *DiskDriver) GetPredictProjectDataPath(projectId int64, model_id int64) string {
	path := dd.GetPredictProjectDir(projectId, model_id)
	return pb.Join(path, "data")
}

func (dd *DiskDriver) GetPredictProjectDir(projectId int64, model_id int64) string {
	path := dd.GetProjectDir(projectId)
	path = pb.Join(path, U.DataTypeEvent)
	model_str := fmt.Sprintf("%d", model_id)
	return pb.Join(path, "predict", model_str)
}

func (dd *DiskDriver) GetEventsUnsortedFilePathAndName(projectId int64, startTimestamp int64, endTimestamp int64) (string, string) {
	path, name := dd.GetEventsFilePathAndName(projectId, startTimestamp, endTimestamp)
	fileName := "unsorted_" + name
	return path, fileName
}

func (dd *DiskDriver) GetEventsArtifactFilePathAndName(projectId int64, startTimestamp int64, endTimestamp int64, group int) (string, string) {
	var fileName string = "users_map.txt"
	var path string
	pathArr := strings.Split(dd.baseDir, "/")
	folderName := pathArr[len(pathArr)-1]
	if folderName == "cloud_storage" {
		modelType := U.GetModelType(startTimestamp, endTimestamp)
		path = dd.GetProjectDataFileDir(projectId, startTimestamp, U.DataTypeEvent, modelType)
	} else {
		path = dd.GetEventsTempFilesDir(projectId, startTimestamp, endTimestamp, group)
	}
	return path, fileName
}

func (dd *DiskDriver) GetChannelArtifactFilePathAndName(channel string, projectId int64, startTimestamp int64, endTimestamp int64) (string, string) {
	var fileName string = "doctypes_map.txt"
	var path string
	pathArr := strings.Split(dd.baseDir, "/")
	folderName := pathArr[len(pathArr)-1]
	if folderName == "cloud_storage" {
		modelType := U.GetModelType(startTimestamp, endTimestamp)
		path = dd.GetProjectDataFileDir(projectId, startTimestamp, U.DataTypeAdReport, modelType)
	} else {
		path = dd.GetChannelTempFilesDir(channel, projectId, startTimestamp, endTimestamp)
	}
	return path, fileName
}

func (dd *DiskDriver) GetPathAnalysisTempFileDir(id string, projectId int64) string {
	path := dd.GetProjectDir(projectId)
	return fmt.Sprintf("%spathanalysis/%v/", path, id)
}
func (dd *DiskDriver) GetPathAnalysisTempFilePathAndName(id string, projectId int64) (string, string) {
	path := dd.GetPathAnalysisTempFileDir(id, projectId)
	return path, "patterns.txt"
}

func (dd *DiskDriver) GetExplainV2Dir(id uint64, projectId int64) string {
	path := fmt.Sprintf("%s/projects/%v/explain/models/%v/", dd.baseDir, projectId, id)
	return path
}

func (dd *DiskDriver) GetExplainV2ModelPath(id uint64, projectId int64) (string, string) {
	path := dd.GetExplainV2Dir(id, projectId)
	chunksPath := pb.Join(path, "chunks")
	return chunksPath, "chunk_1.txt"
}

func (dd *DiskDriver) GetListReferenceFileNameAndPathFromCloud(projectID int64, reference string) (string, string) {
	return fmt.Sprintf("%s/projects/%v/list/%v/", dd.baseDir, projectID, reference), "list.txt"
}
func (dd *DiskDriver) GetSixSignalAnalysisTempFileDir(id string, projectId int64) string {
	path := dd.GetProjectDir(projectId)
	return fmt.Sprintf("%ssixSignal/%v/", path, id)
}

func (dd *DiskDriver) GetSixSignalAnalysisTempFilePathAndName(id string, projectId int64) (string, string) {
	path := dd.GetSixSignalAnalysisTempFileDir(id, projectId)
	return path, "results.txt"
}

func (dd *DiskDriver) GetEventsTempFilesDir(projectId int64, startTimestamp, endTimestamp int64, group int) string {
	path, name := dd.GetEventsGroupFilePathAndName(projectId, startTimestamp, endTimestamp, group)
	path = pb.Join(path, strings.Replace(name, ".txt", "", 1))
	return path
}

func (dd *DiskDriver) GetEventsPartFilesDir(projectId int64, startTimestamp, endTimestamp int64, sorted bool, group int) string {
	tmp := "unsorted"
	if sorted {
		tmp = "sorted"
	}
	path := dd.GetEventsTempFilesDir(projectId, startTimestamp, endTimestamp, group)
	path = pb.Join(path, tmp+"_parts")
	return path
}

func (dd *DiskDriver) GetEventsPartFilePathAndName(projectId int64, startTimestamp, endTimestamp int64, sorted bool, startIndex, endIndex int, group int) (string, string) {
	path := dd.GetEventsPartFilesDir(projectId, startTimestamp, endTimestamp, sorted, group)
	name := fmt.Sprintf("%d-%d_uids.txt", startIndex, endIndex)
	return path, name
}

func (dd *DiskDriver) GetChannelTempFilesDir(channel string, projectId int64, startTimestamp, endTimestamp int64) string {
	path, name := dd.GetChannelFilePathAndName(channel, projectId, startTimestamp, endTimestamp)
	path = pb.Join(path, strings.Replace(name, ".txt", "", 1))
	return path
}

func (dd *DiskDriver) GetChannelPartFilesDir(channel string, projectId int64, startTimestamp, endTimestamp int64, sorted bool) string {
	tmp := "unsorted"
	if sorted {
		tmp = "sorted"
	}
	path := dd.GetChannelTempFilesDir(channel, projectId, startTimestamp, endTimestamp)
	path = pb.Join(path, tmp+"_parts")
	return path
}

func (dd *DiskDriver) GetChannelPartFilePathAndName(channel string, projectId int64, startTimestamp, endTimestamp int64, sorted bool, index int) (string, string) {
	path := dd.GetChannelPartFilesDir(channel, projectId, startTimestamp, endTimestamp, sorted)
	name := fmt.Sprintf("%d_doctype.txt", index)
	return path, name
}
