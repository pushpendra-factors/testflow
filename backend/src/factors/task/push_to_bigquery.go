package task

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sort"

	A "factors/archival"
	"factors/filestore"
	M "factors/model"
	BQ "factors/services/bigquery"
	serviceDisk "factors/services/disk"
	"factors/util"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

var pbTaskID = "Task#PushToBigQuery"

// ArchiveEvents Archives events for all the projects with archival enabled in project settings.
func ArchiveEvents(db *gorm.DB, cloudManager *filestore.FileManager,
	diskManger *serviceDisk.DiskDriver, maxLookbackDays int) (map[uint64][]string, []error) {
	pbLog := taskLog.WithFields(log.Fields{
		"Prefix": pbTaskID + "ArchiveEvents",
	})

	allJobDetails := make(map[uint64][]string)
	enabledProjectIDs, status := M.GetArchiveEnabledProjectIDs()
	if status != http.StatusFound {
		return allJobDetails, []error{fmt.Errorf("Failed to get archive enabled projects from DB")}
	}

	var projectErrors []error
	for _, projectID := range enabledProjectIDs {
		pbLog.Infof("Running archival for project id %d", projectID)
		jobDetails, err := ArchiveEventsForProject(db, cloudManager, diskManger, projectID, maxLookbackDays)
		if err != nil {
			pbLog.WithError(err).Errorf("Archival failed for project id %d", projectID)
			projectErrors = append(projectErrors, err)
		}
		allJobDetails[projectID] = jobDetails
	}
	return allJobDetails, projectErrors
}

// ArchiveEventsForProject Archives events for a particular project to cloud storage.
func ArchiveEventsForProject(db *gorm.DB, cloudManager *filestore.FileManager,
	diskManger *serviceDisk.DiskDriver, projectID uint64, maxLookbackDays int) ([]string, error) {

	var jobDetails []string
	pbLog := taskLog.WithFields(log.Fields{
		"Prefix":    pbTaskID + "#ArchiveEventsForProject",
		"ProjectID": projectID,
	})

	projectSettings, _ := M.GetProjectSetting(projectID)
	if projectSettings == nil {
		return jobDetails, fmt.Errorf("Failed to fetch project settings")
	} else if projectSettings.ArchiveEnabled == nil || !*projectSettings.ArchiveEnabled {
		return jobDetails, fmt.Errorf("Archival not enabled for project id %d", projectID)
	}

	inProgressCount, status := M.GetScheduledTaskInProgressCount(projectID, M.TASK_TYPE_EVENTS_ARCHIVAL)
	if status != http.StatusFound {
		return jobDetails, fmt.Errorf("Failed to in progress tasks count")
	} else if inProgressCount != 0 {
		return jobDetails, fmt.Errorf("%d tasks in progress state. Mark failed or success before proceeding", inProgressCount)
	}

	parentScheduledTask := M.ScheduledTask{
		JobID:      util.GetUUID(),
		TaskType:   M.TASK_TYPE_EVENTS_ARCHIVAL,
		ProjectID:  projectID,
		TaskStatus: M.TASK_STATUS_IN_PROGRESS,
	}
	lastRunTime, status := M.GetScheduledTaskLastRunTimestamp(projectID, M.TASK_TYPE_EVENTS_ARCHIVAL)
	if status == http.StatusInternalServerError {
		return jobDetails, fmt.Errorf("Failed to get last run timestamp")
	} else if status == http.StatusNotFound {
		pbLog.Info("No previous entry found. Running full archival.")
	}

	batches := A.GetNextArchivalBatches(projectID, lastRunTime+1, maxLookbackDays)
	if len(batches) == 0 {
		logInfoAndCaptureJobDetails(pbLog, &jobDetails, "No batches found to be processed. Last run time was %v", lastRunTime)
		return jobDetails, nil
	}
	logInfoAndCaptureJobDetails(pbLog, &jobDetails, "%d batches will be processed in this run", len(batches))

	for _, batch := range batches {
		scheduledTask := parentScheduledTask
		scheduledTask.TaskStartTime = util.TimeNowUnix()
		taskDetails := M.EventArchivalTaskDetails{
			FromTimestamp: batch.StartTime,
			ToTimestamp:   batch.EndTime,
			EventCount:    batch.EventsCount,
			BucketName:    (*cloudManager).GetBucketName(),
		}
		taskDetailsJsonb, _ := util.EncodeStructTypeToPostgresJsonb(taskDetails)
		scheduledTask.TaskDetails = taskDetailsJsonb

		status := M.CreateScheduledTask(&scheduledTask)
		if status != http.StatusCreated {
			return jobDetails, fmt.Errorf("Failed to created schedule task in database")
		}

		pbLog = pbLog.WithFields(log.Fields{
			"TimeRange": fmt.Sprintf("%d-%d", batch.StartTime, batch.EndTime),
			"TaskID":    scheduledTask.ID,
		})
		pbLog.Info("Starting to process new batch.")

		if batch.EventsCount == 0 {
			logInfoAndCaptureJobDetails(
				pbLog, &jobDetails, "No events to be processed. Making empty entry in tasks table for %d-%d",
				batch.StartTime, batch.EndTime)
			taskDetails.FileCreated = false
			taskDetailsJsonb, _ = util.EncodeStructTypeToPostgresJsonb(taskDetails)
			rowsUpdated, status := M.UpdateScheduledTask(
				scheduledTask.ID, taskDetailsJsonb, util.TimeNowUnix(), M.TASK_STATUS_SUCCESS)
			if status != http.StatusAccepted || rowsUpdated == 0 {
				return jobDetails, fmt.Errorf("Failed to update scheduled task in database")
			}
			continue
		}

		tmpEventsFilePath, tmpEventsFileName := (*diskManger).GetEventArchiveFilePathAndName(projectID, batch.StartTime, batch.EndTime)
		tmpEventsFile := tmpEventsFilePath + tmpEventsFileName
		serviceDisk.MkdirAll(tmpEventsFilePath)

		pbLog.Infof("Stating to pull events to file %s", tmpEventsFile)

		rowCount, filePath, err := PullEventsForArchive(db, projectID, tmpEventsFile, batch.StartTime, batch.EndTime)
		if err != nil {
			pbLog.WithError(err).Error("Failed to pull events for archival")
			M.FailScheduleTask(scheduledTask.ID)
			return jobDetails, err
		} else if rowCount == 0 {
			pbLog.Warn("0 events found to be pushed. This shouldn't have happened after initial check")
			taskDetails.FileCreated = false
			taskDetailsJsonb, _ := util.EncodeStructTypeToPostgresJsonb(taskDetails)
			rowsUpdated, status := M.UpdateScheduledTask(
				scheduledTask.ID, taskDetailsJsonb, util.TimeNowUnix(), M.TASK_STATUS_SUCCESS)
			if status != http.StatusAccepted || rowsUpdated == 0 {
				return jobDetails, fmt.Errorf("Failed to update scheduled task in database")
			}
			continue
		}

		// Writing to the cloud storage from local.
		tmpOutputStream, err := os.Open(filePath)
		if err != nil {
			pbLog.WithError(err).Errorf("Failed to open events file %s", filePath)
			M.FailScheduleTask(scheduledTask.ID)
			return jobDetails, err
		}

		cloudEventsFilePath, cloudEventsFileName := (*cloudManager).GetEventArchiveFilePathAndName(projectID, batch.StartTime, batch.EndTime)
		cloudEventsFile := cloudEventsFilePath + cloudEventsFileName
		err = (*cloudManager).Create(cloudEventsFilePath, cloudEventsFileName, tmpOutputStream)
		if err != nil {
			pbLog.WithError(err).Error("Failed to write to cloudStorage", cloudEventsFile)
			M.FailScheduleTask(scheduledTask.ID)
			return jobDetails, err
		}

		taskDetails.FileCreated = true
		taskDetails.FilePath = cloudEventsFile
		taskDetailsJsonb, _ = util.EncodeStructTypeToPostgresJsonb(taskDetails)
		taskEndTime := util.TimeNowUnix()
		rowsUpdated, status := M.UpdateScheduledTask(
			scheduledTask.ID, taskDetailsJsonb, taskEndTime, M.TASK_STATUS_SUCCESS)
		if status != http.StatusAccepted || rowsUpdated == 0 {
			return jobDetails, fmt.Errorf("Failed to update scheduled task in database")
		}
		logInfoAndCaptureJobDetails(pbLog, &jobDetails, "%d events written to cloud file %s in %s",
			rowCount, cloudEventsFile, util.SecondsToHMSString(taskEndTime-scheduledTask.TaskStartTime))
	}
	return jobDetails, nil
}

// PushToBigquery For all the projects with bigquery enabled in project_settings.
func PushToBigquery(cloudManager *filestore.FileManager) (map[uint64][]string, []error) {
	pbLog := taskLog.WithFields(log.Fields{
		"Prefix": pbTaskID + "#PushToBigquery",
	})

	allJobDetails := make(map[uint64][]string)
	enabledProjectIDs, status := M.GetBigqueryEnabledProjectIDs()
	if status != http.StatusFound {
		return allJobDetails, []error{fmt.Errorf("Failed to get bigquery enabled projects from DB")}
	}

	var projectErrors []error
	for _, projectID := range enabledProjectIDs {
		pbLog.Infof("Running bigquery push for project id %d", projectID)
		jobDetails, err := PushToBigqueryForProject(cloudManager, projectID)
		if err != nil {
			pbLog.WithError(err).Errorf("Push to bigquery failed for project id %d", projectID)
			projectErrors = append(projectErrors, err)
		}
		allJobDetails[projectID] = jobDetails
	}
	return allJobDetails, projectErrors
}

// PushToBigqueryForProject Pushes events to Bigquery for bigquerySetting.
func PushToBigqueryForProject(cloudManager *filestore.FileManager, projectID uint64) ([]string, error) {
	pbLog := taskLog.WithFields(log.Fields{
		"Prefix":    pbTaskID + "#PushToBigqueryForProject",
		"ProjectID": projectID,
	})

	var jobDetails []string
	projectSettings, _ := M.GetProjectSetting(projectID)
	if projectSettings == nil {
		return jobDetails, fmt.Errorf("Failed to fetch project settings")
	} else if projectSettings.BigqueryEnabled == nil || !*projectSettings.BigqueryEnabled {
		return jobDetails, fmt.Errorf("Bigquery not enabled for project id %d", projectID)
	}

	inProgressCount, status := M.GetScheduledTaskInProgressCount(projectID, M.TASK_TYPE_BIGQUERY_UPLOAD)
	if status != http.StatusFound {
		return jobDetails, fmt.Errorf("Failed to in progress tasks count")
	} else if inProgressCount != 0 {
		return jobDetails, fmt.Errorf("%d tasks in progress state. Mark failed or success before proceeding", inProgressCount)
	}

	bigquerySetting, status := M.GetBigquerySettingByProjectID(projectID)
	if status == http.StatusInternalServerError {
		return jobDetails, fmt.Errorf("Failed to get bigquery setting for project_id %d", projectID)
	} else if status == http.StatusNotFound {
		return jobDetails, fmt.Errorf("No BigQuery configuration found for project id %d in database", projectID)
	}
	pbLog = pbLog.WithFields(log.Fields{
		"BigqueryTable": fmt.Sprintf("%s.%s", bigquerySetting.BigqueryProjectID,
			bigquerySetting.BigqueryDatasetName),
	})

	newArchiveFilesMap, status := M.GetNewArchivalFileNamesAndEndTimeForProject(
		bigquerySetting.ProjectID, bigquerySetting.LastRunAt)
	if status == http.StatusInternalServerError {
		return jobDetails, fmt.Errorf("Failed to get new archive files from database")
	} else if len(newArchiveFilesMap) == 0 {
		logInfoAndCaptureJobDetails(pbLog, &jobDetails, "No new archive files found to be processed.")
		return jobDetails, nil
	}
	logInfoAndCaptureJobDetails(pbLog, &jobDetails, "%d new files found to be pushed to BigQuery", len(newArchiveFilesMap))

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

	parentScheduledTask := M.ScheduledTask{
		JobID:      util.GetUUID(),
		TaskType:   M.TASK_TYPE_BIGQUERY_UPLOAD,
		ProjectID:  projectID,
		TaskStatus: M.TASK_STATUS_IN_PROGRESS,
	}

	for _, fileEndTime := range fileEndTimes {
		archiveFile := newArchiveFilesMap[fileEndTime]["filepath"]
		taskID := newArchiveFilesMap[fileEndTime]["task_id"]

		scheduledTask := parentScheduledTask
		scheduledTask.TaskStartTime = util.TimeNowUnix()
		taskDetails := M.BigqueryUploadTaskDetails{
			BigqueryProjectID: bigquerySetting.BigqueryProjectID,
			BigqueryDataset:   bigquerySetting.BigqueryDatasetName,
			BigqueryTable:     BQ.BIGQUERY_TABLE_EVENTS,
			ArchivalTaskID:    taskID,
		}
		taskDetailsJsonb, _ := util.EncodeStructTypeToPostgresJsonb(taskDetails)
		scheduledTask.TaskDetails = taskDetailsJsonb
		status = M.CreateScheduledTask(&scheduledTask)
		if status != http.StatusCreated {
			return jobDetails, fmt.Errorf("Failed to create scheduled task for bigquery upload")
		}

		pbLog = pbLog.WithFields(log.Fields{
			"TaskID": scheduledTask.ID,
		})

		uploadStats, err := BQ.UploadFileToBigQuery(ctx, client, archiveFile, bigquerySetting, BQ.BIGQUERY_TABLE_EVENTS, pbLog, cloudManager)
		if err != nil {
			pbLog.WithError(err).Errorf("Failed to upload file %s. Aborting further processing.", archiveFile)
			M.FailScheduleTask(scheduledTask.ID)
			return jobDetails, err
		}
		pbLog.Info("Updating LastRunAt for the bigquery setting.")
		rowsAffected, status := M.UpdateBigquerySettingLastRunAt(bigquerySetting.ID, fileEndTime)
		if status != http.StatusAccepted || rowsAffected == 0 {
			pbLog.Errorf("Failed to updated LastRunAt in bigquery setting for file %s. Aborting further processing.", archiveFile)
			M.FailScheduleTask(scheduledTask.ID)
			return jobDetails, err
		}

		uploadStatsJsonb, _ := util.EncodeStructTypeToPostgresJsonb(uploadStats)
		taskDetails.UploadStats = uploadStatsJsonb
		taskDetailsJsonb, _ = util.EncodeStructTypeToPostgresJsonb(taskDetails)
		taskEndTime := util.TimeNowUnix()
		rowsUpdated, status := M.UpdateScheduledTask(scheduledTask.ID, taskDetailsJsonb, taskEndTime, M.TASK_STATUS_SUCCESS)
		if status != http.StatusAccepted || rowsUpdated == 0 {
			return jobDetails, fmt.Errorf("Failed to update scheduled task. Aborting further processing")
		}
		logInfoAndCaptureJobDetails(pbLog, &jobDetails, "File %s uploaded in %s",
			archiveFile, util.SecondsToHMSString(taskEndTime-scheduledTask.TaskStartTime))
	}
	return jobDetails, nil
}

func logInfoAndCaptureJobDetails(pbLog *log.Entry, jobDetails *[]string, message string, args ...interface{}) {
	logMessage := fmt.Sprintf(message, args...)
	pbLog.Info(logMessage)
	*jobDetails = append(*jobDetails, logMessage)
}
