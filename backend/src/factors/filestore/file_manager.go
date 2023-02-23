package filestore

import (
	"io"
)

type FileManager interface {
	Create(dir, fileName string, reader io.Reader) error
	GetWriter(dir, fileName string) (io.WriteCloser, error)
	Get(path, fileName string) (io.ReadCloser, error)
	// Remove(path, filename string) error
	GetObjectSize(dir, fileName string) (int64, error)
	// Del(dir, filename string)error
	GetBucketName() string
	GetProjectModelDir(projectId int64, modelId uint64) string
	GetProjectDir(projectId int64) string
	GetModelEventInfoFilePathAndName(projectId int64, modelId uint64) (string, string)
	GetEventsFilePathAndName(projectId int64, startTimestamp, endTimestamp int64) (string, string)
	GetEventsGroupFilePathAndName(projectId int64, startTimestamp, endTimestamp int64, group int) (string, string)
	GetModelMetricsFilePathAndName(projectId int64, startTimestamp, endTimestamp int64) (string, string)
	GetChannelFilePathAndName(channel string, projectId int64, startTimestamp, endTimestamp int64) (string, string)
	GetUsersFilePathAndName(dateField string, projectId int64, startTimestamp, endTimestamp int64) (string, string)
	GetModelEventsBucketingFilePathAndName(projectId int64, startTimestamp, endTimestamp int64) (string, string)
	GetModelAlertsFilePathAndName(projectId int64, startTimestamp int64, endTimestamp int64) (string, string)
	GetModelArtifactsPath(projectId int64, modelId uint64) string
	GetMasterNumericalBucketsFile(projectId int64) (string, string)
	GetModelEventsNumericalBucketsFile(projectId int64, startTimestamp, endTimestamp int64) (string, string)
	GetPatternChunksDir(projectId int64, modelId uint64) string
	GetChunksMetaDataDir(projectId int64, modelId uint64) string
	GetPatternChunkFilePathAndName(projectId int64, modelId uint64, chunkId string) (string, string)
	GetChunksMetaDataFilePathAndName(projectId int64, modelId uint64) (string, string)
	GetEventArchiveFilePathAndName(projectID int64, startTime, endTime int64) (string, string)
	GetUsersArchiveFilePathAndName(projectID int64, startTime, endTime int64) (string, string)
	GetDailyEventArchiveFilePathAndName(projectID int64, dataTimestamp, startTime, endTime int64) (string, string)
	GetDailyUsersArchiveFilePathAndName(dateField string, projectID int64, dataTimestamp, startTime, endTime int64) (string, string)
	GetDailyChannelArchiveFilePathAndName(channel string, projectID int64, dataTimestamp, startTime, endTime int64) (string, string)
	ListFiles(path string) []string
	GetWeeklyInsightsMailerModelDir(projectId int64, dateString string, queryId int64, k int) string
	GetInsightsWpiFilePathAndName(projectId int64, dateString string, queryId int64, k int, mailerRun bool) (string, string)
	GetInsightsCpiFilePathAndName(projectId int64, dateString string, queryId int64, k int, mailerRun bool) (string, string)
	GetWeeklyInsightsModelDir(projectId int64, dateString string, queryId int64, k int) string
	GetModelUserPropertiesCategoricalFilePathAndName(projectId int64, modelId uint64) (string, string)
	GetModelEventPropertiesCategoricalFilePathAndName(projectId int64, modelId uint64) (string, string)
	GetModelUserPropertiesFilePathAndName(projectId int64, modelId uint64) (string, string)
	GetModelEventPropertiesFilePathAndName(projectId int64, modelId uint64) (string, string)

	GetAdsDataDir(projectId int64) string
	GetAdsDataFilePathAndName(projectId int64, report string, chunkNo int) (string, string)

	//predict
	GetPredictProjectDataPath(projectId int64, model_id int64) string
	GetPredictProjectDir(projectId int64, model_id int64) string

	GetWIPropertiesDir(projectId int64) string
	GetWIPropertiesPathAndName(projectId int64) (string, string)

	GetEventsUnsortedFilePathAndName(projectId int64, startTimestamp int64, endTimestmap int64) (string, string)
	GetEventsArtifactFilePathAndName(projectId int64, startTimestamp int64, endTimestmap int64) (string, string)
	GetChannelArtifactFilePathAndName(channel string, projectId int64, startTimestamp int64, endTimestmap int64) (string, string)

	GetPathAnalysisTempFileDir(id string, projectId int64) string
	GetPathAnalysisTempFilePathAndName(id string, projectId int64) (string, string)

	GetExplainV2Dir(id uint64, projectId int64) string
	GetExplainV2ModelPath(id uint64, projectId int64) (string, string)

	GetListReferenceFileNameAndPathFromCloud(projectID int64, reference string) (string, string)

	GetSixSignalAnalysisTempFileDir(id string, projectId int64) string
	GetSixSignalAnalysisTempFilePathAndName(id string, projectId int64) (string, string)

}
