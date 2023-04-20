package task

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"factors/filestore"
	"factors/model/model"
	"factors/model/store"
	"factors/pull"
	BQ "factors/services/bigquery"
	serviceDisk "factors/services/disk"
	"factors/util"

	"cloud.google.com/go/bigquery"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

var pbTaskID = "Task#PushToBigQuery"

func writeCloudArchiveFile(cloudManager *filestore.FileManager, tmpFileName, cloudFilePath, cloudFileName string) error {
	// Writing to the cloud storage from local.
	tmpOutputStream, err := os.Open(tmpFileName)
	if err != nil {
		return err
	}

	err = (*cloudManager).Create(cloudFilePath, cloudFileName, tmpOutputStream)
	if err != nil {
		return err
	}
	return nil
}

// ArchiveEvents Archives events for all the projects with archival enabled in project settings.
func ArchiveEvents(db *gorm.DB, cloudManager *filestore.FileManager,
	diskManger *serviceDisk.DiskDriver, maxLookbackDays int, startTime, endTime time.Time) (map[int64][]string, []error) {
	pbLog := taskLog.WithFields(log.Fields{
		"Prefix": pbTaskID + "ArchiveEvents",
	})

	allJobDetails := make(map[int64][]string)
	enabledProjectIDs, status := store.GetStore().GetArchiveEnabledProjectIDs()
	if status != http.StatusFound {
		return allJobDetails, []error{fmt.Errorf("Failed to get archive enabled projects from DB")}
	}

	var projectErrors []error
	for _, projectID := range enabledProjectIDs {
		pbLog.Infof("Running archival for project id %d", projectID)
		jobDetails, err := ArchiveEventsForProject(db, cloudManager, diskManger, projectID, maxLookbackDays, startTime, endTime, false)
		if err != nil {
			pbLog.WithError(err).Errorf("Archival failed for project id %d", projectID)
			projectErrors = append(projectErrors, err)
		}
		allJobDetails[projectID] = jobDetails
	}
	return allJobDetails, projectErrors
}

// ArchiveEventsForProject Archives events for a particular project to cloud storage.
func ArchiveEventsForProject(db *gorm.DB, cloudManager *filestore.FileManager, diskManger *serviceDisk.DiskDriver,
	projectID int64, maxLookbackDays int, startTime, endTime time.Time, bypassSettings bool) ([]string, error) {

	var jobDetails []string
	pbLog := taskLog.WithFields(log.Fields{
		"Prefix":    pbTaskID + "#ArchiveEventsForProject",
		"ProjectID": projectID,
	})

	if !bypassSettings {
		projectSettings, _ := store.GetStore().GetProjectSetting(projectID)
		if projectSettings == nil {
			return jobDetails, fmt.Errorf("Failed to fetch project settings")
		} else if projectSettings.ArchiveEnabled == nil || !*projectSettings.ArchiveEnabled {
			return jobDetails, fmt.Errorf("Archival not enabled for project id %d", projectID)
		}
	}

	if yes, err := pull.CheckIfAddSessionCompleted(projectID, endTime.Unix()); !yes {
		if err != nil {
			return jobDetails, err
		}
		return jobDetails, fmt.Errorf("Add session job not completed for project id %d", projectID)
	}

	inProgressCount, status := store.GetStore().GetScheduledTaskInProgressCount(projectID, model.TASK_TYPE_EVENTS_ARCHIVAL)
	if status != http.StatusFound {
		return jobDetails, fmt.Errorf("Failed to in progress tasks count")
	} else if inProgressCount != 0 {
		return jobDetails, fmt.Errorf("%d tasks in progress state. Mark failed or success before proceeding", inProgressCount)
	}

	parentScheduledTask := model.ScheduledTask{
		JobID:      util.GetUUID(),
		TaskType:   model.TASK_TYPE_EVENTS_ARCHIVAL,
		ProjectID:  projectID,
		TaskStatus: model.TASK_STATUS_IN_PROGRESS,
	}
	lastRunTime, status := store.GetStore().GetScheduledTaskLastRunTimestamp(projectID, model.TASK_TYPE_EVENTS_ARCHIVAL)
	if status == http.StatusInternalServerError {
		return jobDetails, fmt.Errorf("Failed to get last run timestamp")
	} else if status == http.StatusNotFound {
		pbLog.Info("No previous entry found. Running full archival.")
	}

	batches, err := store.GetStore().GetNextArchivalBatches(projectID, lastRunTime+1, maxLookbackDays, startTime, endTime)
	if err != nil {
		return jobDetails, err
	} else if len(batches) == 0 {
		logInfoAndCaptureJobDetails(pbLog, &jobDetails, "No batches found to be processed. Last run time was %v", lastRunTime)
		return jobDetails, nil
	}
	logInfoAndCaptureJobDetails(pbLog, &jobDetails, "%d batches will be processed in this run", len(batches))

	for _, batch := range batches {
		scheduledTask := parentScheduledTask
		scheduledTask.TaskStartTime = util.TimeNowUnix()
		taskDetails := model.EventArchivalTaskDetails{
			FromTimestamp: batch.StartTime,
			ToTimestamp:   batch.EndTime,
			BucketName:    (*cloudManager).GetBucketName(),
		}
		taskDetailsJsonb, _ := util.EncodeStructTypeToPostgresJsonb(taskDetails)
		scheduledTask.TaskDetails = taskDetailsJsonb

		status := store.GetStore().CreateScheduledTask(&scheduledTask)
		if status != http.StatusCreated {
			return jobDetails, fmt.Errorf("Failed to created schedule task in database")
		}

		pbLog = pbLog.WithFields(log.Fields{
			"TimeRange": fmt.Sprintf("%d-%d", batch.StartTime, batch.EndTime),
			"TaskID":    scheduledTask.ID,
		})

		pbLog.Info("Starting to process new batch.")
		tmpEventsFilePath, tmpEventsFileName := (*diskManger).GetEventArchiveFilePathAndName(projectID, batch.StartTime, batch.EndTime)
		tmpEventsFile := tmpEventsFilePath + tmpEventsFileName
		serviceDisk.MkdirAll(tmpEventsFilePath)

		tmpUsersFilePath, tmpUsersFileName := (*diskManger).GetUsersArchiveFilePathAndName(projectID, batch.StartTime, batch.EndTime)
		tmpUsersFile := tmpUsersFilePath + tmpUsersFileName
		serviceDisk.MkdirAll(tmpUsersFilePath)

		pbLog.Infof("Stating to pull events to file %s", tmpEventsFile)

		rowCount, eventsFilePath, usersFilePath, err := pull.PullEventsForArchive(projectID,
			tmpEventsFile, tmpUsersFile, batch.StartTime, batch.EndTime)
		taskDetails.EventCount = int64(rowCount)
		if err != nil {
			pbLog.WithError(err).Error("Failed to pull events for archival")
			store.GetStore().FailScheduleTask(scheduledTask.ID)
			return jobDetails, err
		} else if rowCount == 0 {
			logInfoAndCaptureJobDetails(
				pbLog, &jobDetails, "0 events found to be processed. Making empty entry in tasks table for %d-%d",
				batch.StartTime, batch.EndTime)
			taskDetails.FileCreated = false
			taskDetailsJsonb, _ := util.EncodeStructTypeToPostgresJsonb(taskDetails)
			rowsUpdated, status := store.GetStore().UpdateScheduledTask(
				scheduledTask.ID, taskDetailsJsonb, util.TimeNowUnix(), model.TASK_STATUS_SUCCESS)
			if status != http.StatusAccepted || rowsUpdated == 0 {
				return jobDetails, fmt.Errorf("Failed to update scheduled task in database")
			}
			continue
		}

		// Writing to the cloud storage from local.
		cloudEventsFilePath, cloudEventsFileName := (*cloudManager).GetEventArchiveFilePathAndName(projectID, batch.StartTime, batch.EndTime)
		cloudUsersFilePath, cloudUsersFileName := (*cloudManager).GetUsersArchiveFilePathAndName(projectID, batch.StartTime, batch.EndTime)
		if err := writeCloudArchiveFile(cloudManager, eventsFilePath, cloudEventsFilePath, cloudEventsFileName); err != nil {
			pbLog.WithError(err).Errorf("Failed to create events file %s", eventsFilePath)
			store.GetStore().FailScheduleTask(scheduledTask.ID)
			return jobDetails, err
		} else if err := writeCloudArchiveFile(cloudManager, usersFilePath, cloudUsersFilePath, cloudUsersFileName); err != nil {
			pbLog.WithError(err).Errorf("Failed to create users file %s", eventsFilePath)
			store.GetStore().FailScheduleTask(scheduledTask.ID)
			return jobDetails, err
		}
		cloudEventsFile := cloudEventsFilePath + cloudEventsFileName
		cloudUserFile := cloudUsersFilePath + cloudUsersFileName

		taskDetails.FileCreated = true
		taskDetails.FilePath = cloudEventsFile
		taskDetails.UsersFilePath = cloudUserFile
		taskDetailsJsonb, _ = util.EncodeStructTypeToPostgresJsonb(taskDetails)
		taskEndTime := util.TimeNowUnix()
		rowsUpdated, status := store.GetStore().UpdateScheduledTask(
			scheduledTask.ID, taskDetailsJsonb, taskEndTime, model.TASK_STATUS_SUCCESS)
		if status != http.StatusAccepted || rowsUpdated == 0 {
			return jobDetails, fmt.Errorf("Failed to update scheduled task in database")
		}
		logInfoAndCaptureJobDetails(pbLog, &jobDetails, "%d events written to cloud file %s in %s",
			rowCount, cloudEventsFile, util.SecondsToHMSString(taskEndTime-scheduledTask.TaskStartTime))
	}
	return jobDetails, nil
}

// PushToBigquery For all the projects with bigquery enabled in project_settings.
func PushToBigquery(cloudManager *filestore.FileManager, startTime, endTime time.Time) (map[int64][]string, []error) {
	pbLog := taskLog.WithFields(log.Fields{
		"Prefix": pbTaskID + "#PushToBigquery",
	})

	allJobDetails := make(map[int64][]string)
	enabledProjectIDs, status := store.GetStore().GetBigqueryEnabledProjectIDs()
	if status != http.StatusFound {
		return allJobDetails, []error{fmt.Errorf("Failed to get bigquery enabled projects from DB")}
	}

	var projectErrors []error
	for _, projectID := range enabledProjectIDs {
		pbLog.Infof("Running bigquery push for project id %d", projectID)
		jobDetails, err := PushToBigqueryForProject(cloudManager, projectID, startTime, endTime)
		if err != nil {
			pbLog.WithError(err).Errorf("Push to bigquery failed for project id %d", projectID)
			projectErrors = append(projectErrors, err)
		}
		allJobDetails[projectID] = jobDetails
	}
	return allJobDetails, projectErrors
}

// PushToBigqueryForProject Pushes events to Bigquery for bigquerySetting.
func PushToBigqueryForProject(cloudManager *filestore.FileManager, projectID int64, startTime, endTime time.Time) ([]string, error) {
	pbLog := taskLog.WithFields(log.Fields{
		"Prefix":    pbTaskID + "#PushToBigqueryForProject",
		"ProjectID": projectID,
	})

	var jobDetails []string
	projectSettings, _ := store.GetStore().GetProjectSetting(projectID)
	if projectSettings == nil {
		return jobDetails, fmt.Errorf("Failed to fetch project settings")
	} else if projectSettings.BigqueryEnabled == nil || !*projectSettings.BigqueryEnabled {
		return jobDetails, fmt.Errorf("Bigquery not enabled for project id %d", projectID)
	}

	inProgressCount, status := store.GetStore().GetScheduledTaskInProgressCount(projectID, model.TASK_TYPE_BIGQUERY_UPLOAD)
	if status != http.StatusFound {
		return jobDetails, fmt.Errorf("Failed to in progress tasks count")
	} else if inProgressCount != 0 {
		return jobDetails, fmt.Errorf("%d tasks in progress state. Mark failed or success before proceeding", inProgressCount)
	}

	bigquerySetting, status := store.GetStore().GetBigquerySettingByProjectID(projectID)
	if status == http.StatusInternalServerError {
		return jobDetails, fmt.Errorf("Failed to get bigquery setting for project_id %d", projectID)
	} else if status == http.StatusNotFound {
		return jobDetails, fmt.Errorf("No BigQuery configuration found for project id %d in database", projectID)
	}
	pbLog = pbLog.WithFields(log.Fields{
		"BigqueryTable": fmt.Sprintf("%s.%s", bigquerySetting.BigqueryProjectID,
			bigquerySetting.BigqueryDatasetName),
	})

	lastRunAt, status := store.GetStore().GetScheduledTaskLastRunTimestamp(projectID, model.TASK_TYPE_BIGQUERY_UPLOAD)
	if status == http.StatusInternalServerError {
		return jobDetails, fmt.Errorf("Failed to get last run timestamp")
	} else if status == http.StatusNotFound {
		pbLog.Info("No previous entry found. Will process all the files.")
	}

	newArchiveFilesMap, status := store.GetStore().GetNewArchivalFileNamesAndEndTimeForProject(
		projectID, lastRunAt, startTime, endTime)
	if status == http.StatusInternalServerError {
		return jobDetails, fmt.Errorf("Failed to get new archive files from database")
	} else if len(newArchiveFilesMap) == 0 {
		logInfoAndCaptureJobDetails(pbLog, &jobDetails, "No new archive files found to be processed.")
		return jobDetails, nil
	}
	logInfoAndCaptureJobDetails(pbLog, &jobDetails, "%d new tasks found to be pushed to BigQuery starting from %d",
		len(newArchiveFilesMap), lastRunAt)

	var fileEndTimes []int64
	for endTime := range newArchiveFilesMap {
		fileEndTimes = append(fileEndTimes, endTime)
	}
	// Sort and process files in ascending order of endTimes.
	sort.Slice(fileEndTimes, func(i, j int) bool {
		return fileEndTimes[i] < fileEndTimes[j]
	})

	ctx := context.Background()
	client, err := BQ.CreateBigqueryClient(&ctx, bigquerySetting)
	if err != nil {
		pbLog.WithError(err).Error("Failed to create bigquery client")
		return jobDetails, err
	}
	defer client.Close()

	parentScheduledTask := model.ScheduledTask{
		JobID:      util.GetUUID(),
		TaskType:   model.TASK_TYPE_BIGQUERY_UPLOAD,
		ProjectID:  projectID,
		TaskStatus: model.TASK_STATUS_IN_PROGRESS,
	}

	for _, fileEndTime := range fileEndTimes {
		archiveFile := newArchiveFilesMap[fileEndTime]["filepath"].(string)
		userArchiveFile := newArchiveFilesMap[fileEndTime]["users_filepath"].(string)
		taskID := newArchiveFilesMap[fileEndTime]["task_id"].(string)
		fileStartTime := newArchiveFilesMap[fileEndTime]["start_time"].(int64)

		scheduledTask := parentScheduledTask
		scheduledTask.TaskStartTime = util.TimeNowUnix()
		taskDetails := model.BigqueryUploadTaskDetails{
			FromTimestamp:     fileStartTime,
			ToTimestamp:       fileEndTime,
			BigqueryProjectID: bigquerySetting.BigqueryProjectID,
			BigqueryDataset:   bigquerySetting.BigqueryDatasetName,
			BigqueryTable:     BQ.BIGQUERY_TABLE_EVENTS,
			ArchivalTaskID:    taskID,
		}
		taskDetailsJsonb, _ := util.EncodeStructTypeToPostgresJsonb(taskDetails)
		scheduledTask.TaskDetails = taskDetailsJsonb
		status = store.GetStore().CreateScheduledTask(&scheduledTask)
		if status != http.StatusCreated {
			return jobDetails, fmt.Errorf("Failed to create scheduled task for bigquery upload")
		}

		pbLog = pbLog.WithFields(log.Fields{
			"TaskID": scheduledTask.ID,
		})

		uploadStats, err := BQ.UploadFileToBigQuery(ctx, client, archiveFile, bigquerySetting, BQ.BIGQUERY_TABLE_EVENTS, pbLog, cloudManager)
		if err != nil {
			pbLog.WithError(err).Errorf("Failed to upload file %s. Aborting further processing.", archiveFile)
			store.GetStore().FailScheduleTask(scheduledTask.ID)
			return jobDetails, err
		}

		// TODO(prateek): Add a mechanism to handle case when one of event file is uploaded but failed for user file.
		// While reprocessing, it should not upload events file again and only process user file.
		var usersUploadStats *bigquery.JobStatus
		if userArchiveFile != "" {
			// Not empty check to ensure backward compatibility.
			usersUploadStats, err = BQ.UploadFileToBigQuery(ctx, client, userArchiveFile,
				bigquerySetting, BQ.BIGQUERY_TABLE_USERS, pbLog, cloudManager)
			if err != nil {
				pbLog.WithError(err).Errorf("Failed to upload file %s. Aborting further processing.", userArchiveFile)
				store.GetStore().FailScheduleTask(scheduledTask.ID)
				return jobDetails, err
			}
		}

		uploadStatsJsonb, _ := util.EncodeStructTypeToPostgresJsonb(uploadStats)
		usersUploadStatsJsonb, _ := util.EncodeStructTypeToPostgresJsonb(usersUploadStats)
		taskDetails.UploadStats = uploadStatsJsonb
		taskDetails.UsersUploadStats = usersUploadStatsJsonb
		taskDetailsJsonb, _ = util.EncodeStructTypeToPostgresJsonb(taskDetails)
		taskEndTime := util.TimeNowUnix()
		rowsUpdated, status := store.GetStore().UpdateScheduledTask(scheduledTask.ID, taskDetailsJsonb, taskEndTime, model.TASK_STATUS_SUCCESS)
		if status != http.StatusAccepted || rowsUpdated == 0 {
			return jobDetails, fmt.Errorf("Failed to update scheduled task. Aborting further processing")
		}
		logInfoAndCaptureJobDetails(pbLog, &jobDetails, "File %s uploaded in %s",
			strings.TrimSuffix(strings.Join([]string{archiveFile, userArchiveFile}, ","), ","),
			util.SecondsToHMSString(taskEndTime-scheduledTask.TaskStartTime))
	}

	// Dedup users table.
	pbLog.Infof("Deduping users on bigquery")
	resultSet, err := dedupUsersTable(ctx, client, bigquerySetting)
	if err != nil {
		pbLog.WithError(err).Errorf("Failed to dedup users on bigquery. Result: %v", resultSet)
	}
	return jobDetails, nil
}

func dedupUsersTable(ctx context.Context, client *bigquery.Client, bigquerySetting *model.BigquerySetting) ([][]string, error) {
	bigqueryTableName := fmt.Sprintf("%s.%s.%s", bigquerySetting.BigqueryProjectID,
		bigquerySetting.BigqueryDatasetName, BQ.BIGQUERY_TABLE_USERS)

	// Caution: If same file is processed twice, it might lead to duplicates.
	query := fmt.Sprintf("DELETE FROM %s WHERE (user_id, ingestion_date) NOT IN (SELECT (user_id, MAX(ingestion_date)) FROM "+
		"%s GROUP BY user_id);", bigqueryTableName, bigqueryTableName)

	resultSet := make([][]string, 0, 0)
	err := BQ.ExecuteQuery(&ctx, client, query, &resultSet)
	if err != nil {
		return nil, err
	}
	return resultSet, nil
}

func logInfoAndCaptureJobDetails(pbLog *log.Entry, jobDetails *[]string, message string, args ...interface{}) {
	logMessage := fmt.Sprintf(message, args...)
	pbLog.Info(logMessage)
	*jobDetails = append(*jobDetails, logMessage)
}
