package disk

import (
	"factors/filestore"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	U "factors/util"

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

	if !strings.HasSuffix(path, "/") {
		// Append / to the end if not present.
		path = path + "/"
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

	if !strings.HasSuffix(path, "/") {
		// Append / to the end if not present.
		path = path + "/"
	}
	file, err := os.OpenFile(path+fileName, os.O_RDONLY, 0444)
	return file, err
}

func (dd *DiskDriver) GetBucketName() string {
	return dd.baseDir
}

func (dd *DiskDriver) GetObjectSize(path, fileName string) (int64, error) {
	if !strings.HasSuffix(path, "/") {
		// Append / to the end if not present.
		path = path + "/"
	}
	var objInfo os.FileInfo
	var err error
	if objInfo, err = os.Stat(path + fileName); err != nil {
		return 0, err
	}
	objSize := objInfo.Size()
	return objSize, err
}

func (dd *DiskDriver) GetProjectModelDir(projectId int64, modelId uint64) string {
	return fmt.Sprintf("%s/projects/%d/models/%d/", dd.baseDir, projectId, modelId)
}

func (dd *DiskDriver) GetProjectEventFileDir(projectId int64, startTimestamp int64, modelType string) string {
	dateFormatted := U.GetDateOnlyFromTimestampZ(startTimestamp)
	return fmt.Sprintf("%s/projects/%d/events/%s/%s/", dd.baseDir, projectId, modelType, dateFormatted)
}

func (dd *DiskDriver) GetProjectDir(projectId int64) string {
	return fmt.Sprintf("%s/projects/%d/", dd.baseDir, projectId)
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

func (dd *DiskDriver) GetModelEventsFilePathAndName(projectId int64, startTimestamp int64, modelType string) (string, string) {
	path := dd.GetProjectEventFileDir(projectId, startTimestamp, modelType)
	return path, "events.txt"
}

func (dd *DiskDriver) GetModelChannelFilePathAndName(channel string, projectId int64, startTimestamp int64, modelType string) (string, string) {
	path := dd.GetProjectEventFileDir(projectId, startTimestamp, modelType)
	return path, channel + ".txt"
}

func (dd *DiskDriver) GetModelSmartPropertiesFilePathAndName(projectId int64, startTimestamp int64, modelType string) (string, string) {
	path := dd.GetProjectEventFileDir(projectId, startTimestamp, modelType)
	return path, "smart_properties.txt"
}

func (dd *DiskDriver) GetModelEventsBucketingFilePathAndName(projectId int64, startTimestamp int64, modelType string) (string, string) {
	path := dd.GetProjectEventFileDir(projectId, startTimestamp, modelType)
	return path, "events_bucketed.txt"
}

// If we have two different files for last week and this week, we might end up having non-overlapping ranges.
// So keeping one for this week and other as master. If it exists in master pick that else this week, or compute now
func (dd *DiskDriver) GetMasterNumericalBucketsFile(projectId int64) (string, string) {
	path := dd.GetProjectDir(projectId)
	return path, "numerical_buckets_master.txt"
}

func (dd *DiskDriver) GetModelEventsNumericalBucketsFile(projectId int64, startTimestamp int64, modelType string) (string, string) {
	path := dd.GetProjectEventFileDir(projectId, startTimestamp, modelType)
	return path, "numerical_buckets.txt"
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

func (dd *DiskDriver) GetInsightsWpiFilePathAndName(projectId int64, dateString string, queryId int64, k int) (string, string) {
	path := dd.GetWeeklyInsightsModelDir(projectId, dateString, queryId, k)
	return path, "wpi.txt"
}

func (dd *DiskDriver) GetInsightsCpiFilePathAndName(projectId int64, dateString string, queryId int64, k int) (string, string) {
	path := dd.GetWeeklyInsightsModelDir(projectId, dateString, queryId, k)
	return path, "cpi.txt"
}

func (dd *DiskDriver) GetWeeklyInsightsModelDir(projectId int64, dateString string, queryId int64, k int) string {
	return fmt.Sprintf("%v/projects/%d/weeklyinsights/%v/q-%v/k-%v/", dd.baseDir, projectId, dateString, queryId, k)
}

func (dd *DiskDriver) GetWeeklyKPIModelDir(projectId int64, dateString string, queryId int64) string {
	return fmt.Sprintf("%v/projects/%v/weeklyKPI/%v/q-%v/", dd.baseDir, projectId, dateString, queryId)
}

func (dd *DiskDriver) GetKPIFilePathAndName(projectId int64, dateString string, queryId int64) (string, string) {
	path := dd.GetWeeklyKPIModelDir(projectId, dateString, queryId)
	return path, "kpi.txt"
}
