package tests

import (
	M "factors/model"
	U "factors/util"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreateScheduledTask(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	scheduledTask := M.ScheduledTask{
		ProjectID:     project.ID,
		TaskType:      M.TASK_TYPE_EVENTS_ARCHIVAL,
		TaskStartTime: U.TimeNowUnix(),
	}

	// With empty JobID. Should fail.
	status := M.CreateScheduledTask(&scheduledTask)
	assert.Equal(t, status, http.StatusBadRequest)

	scheduledTask.JobID = U.GetUUID()
	status = M.CreateScheduledTask(&scheduledTask)
	assert.Equal(t, status, http.StatusCreated)

	scheduledTaskDB, _ := M.GetScheduledTaskByID(scheduledTask.ID)
	assert.NotNil(t, scheduledTaskDB)
	assert.Equal(t, scheduledTask.ID, scheduledTaskDB.ID)
}

func TestUpdateScheduledTaskDetailsAndEndTime(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	scheduledTask, taskDetails := getDummyValidArchivalScheduledTask(project.ID)
	status := M.CreateScheduledTask(&scheduledTask)
	assert.Equal(t, status, http.StatusCreated)

	taskDetails.FileCreated = false
	taskDetailsJsonb, _ := U.EncodeStructTypeToPostgresJsonb(taskDetails)
	taskEndTime := U.TimeNowUnix() + 1
	rowsUpdated, status := M.UpdateScheduledTask(scheduledTask.ID, taskDetailsJsonb, taskEndTime, M.TASK_STATUS_SUCCESS)
	assert.Equal(t, status, http.StatusAccepted)
	assert.NotEqual(t, rowsUpdated, 0)

	scheduledTaskDB, _ := M.GetScheduledTaskByID(scheduledTask.ID)
	assert.NotNil(t, scheduledTaskDB)
	assert.Equal(t, scheduledTask.ID, scheduledTaskDB.ID)
	assert.Equal(t, taskEndTime, scheduledTaskDB.TaskEndTime)
	assert.Equal(t, M.TASK_STATUS_SUCCESS, scheduledTaskDB.TaskStatus)

	expectedTaskDetails, _ := U.DecodePostgresJsonb(taskDetailsJsonb)
	actualTaskDetails, _ := U.DecodePostgresJsonb(scheduledTaskDB.TaskDetails)
	assert.Equal(t, expectedTaskDetails, actualTaskDetails)
}

func TestGetScheduledTaskInProgressCount(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	scheduledTask1, _ := getDummyValidBigqueryScheduledTask(project.ID)
	scheduledTask2, _ := getDummyValidBigqueryScheduledTask(project.ID)
	_ = M.CreateScheduledTask(&scheduledTask1)
	_ = M.CreateScheduledTask(&scheduledTask2)

	inProgressCount, status := M.GetScheduledTaskInProgressCount(project.ID, M.TASK_TYPE_BIGQUERY_UPLOAD)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, int64(2), inProgressCount)
}

func TestGetScheduledTaskLastRunTimestamp(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	scheduledTask1, _ := getDummyValidArchivalScheduledTask(project.ID)
	scheduledTask1.TaskStatus = M.TASK_STATUS_SUCCESS

	expectedLastRunTimestamp := scheduledTask1.TaskStartTime + 86400
	scheduledTask2, _ := getDummyValidArchivalScheduledTask(project.ID)
	scheduledTask2.TaskStatus = M.TASK_STATUS_SUCCESS
	scheduledTask2.TaskStartTime = expectedLastRunTimestamp

	// Failed task should not be checked for LastRun timestamp even if run time is later.
	scheduledTask3, _ := getDummyValidArchivalScheduledTask(project.ID)
	scheduledTask3.TaskStatus = M.TASK_STATUS_FAILED
	scheduledTask3.TaskStartTime = scheduledTask1.TaskStartTime + 2*86400

	_ = M.CreateScheduledTask(&scheduledTask1)
	_ = M.CreateScheduledTask(&scheduledTask2)
	_ = M.CreateScheduledTask(&scheduledTask3)

	actualLastRunTimestamp, status := M.GetScheduledTaskLastRunTimestamp(project.ID, scheduledTask1.TaskType)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, expectedLastRunTimestamp, actualLastRunTimestamp)
}

func TestGetNewArchivalFileNamesAndEndTimeForProject(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	scheduledTask1, taskDetails1 := getDummyValidArchivalScheduledTask(project.ID)
	taskDetails1.FileCreated = true
	taskDetails1.FilePath = "filepath1"
	taskDetails1Jsonb, _ := U.EncodeStructTypeToPostgresJsonb(taskDetails1)
	scheduledTask1.TaskDetails = taskDetails1Jsonb

	scheduledTask2, taskDetails2 := getDummyValidArchivalScheduledTask(project.ID)
	taskDetails2.FileCreated = false
	taskDetails2.FilePath = "filepath2"
	taskDetails2.FromTimestamp = taskDetails1.FromTimestamp - 86400 // One day before task1.
	taskDetails2.ToTimestamp = taskDetails1.ToTimestamp - 86400
	taskDetails2Jsonb, _ := U.EncodeStructTypeToPostgresJsonb(taskDetails2)
	scheduledTask2.TaskDetails = taskDetails2Jsonb

	scheduledTask3, taskDetails3 := getDummyValidArchivalScheduledTask(project.ID)
	taskDetails3.FileCreated = true
	taskDetails3.FilePath = "filepath3"
	taskDetails3.FromTimestamp = taskDetails2.FromTimestamp - 86400 // 2 days before task1.
	taskDetails3.ToTimestamp = taskDetails2.ToTimestamp - 86400
	taskDetails3Jsonb, _ := U.EncodeStructTypeToPostgresJsonb(taskDetails3)
	scheduledTask3.TaskDetails = taskDetails3Jsonb

	_ = M.CreateScheduledTask(&scheduledTask1)
	_ = M.CreateScheduledTask(&scheduledTask2)
	_ = M.CreateScheduledTask(&scheduledTask3)

	// Non inclusive check. Should include task1 and task3.
	bigQueryLastRunAt := taskDetails3.FromTimestamp - 1
	newFilesMap, status := M.GetNewArchivalFileNamesAndEndTimeForProject(project.ID, bigQueryLastRunAt, time.Time{}, time.Time{})
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, 2, len(newFilesMap))

	// Inclusive check. Should include only task3.
	bigQueryLastRunAt = taskDetails3.FromTimestamp
	newFilesMap, status = M.GetNewArchivalFileNamesAndEndTimeForProject(project.ID, bigQueryLastRunAt, time.Time{}, time.Time{})
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, 1, len(newFilesMap))
}

func getDummyValidArchivalScheduledTask(projectID uint64) (M.ScheduledTask, M.EventArchivalTaskDetails) {
	scheduledTask := M.ScheduledTask{
		ProjectID:     projectID,
		JobID:         U.GetUUID(),
		TaskType:      M.TASK_TYPE_EVENTS_ARCHIVAL,
		TaskStartTime: U.TimeNowUnix(),
		TaskStatus:    M.TASK_STATUS_IN_PROGRESS,
	}

	taskDetails := M.EventArchivalTaskDetails{
		FromTimestamp: U.TimeNowUnix(),
		ToTimestamp:   U.TimeNow().AddDate(0, 0, 1).Unix(),
		EventCount:    10,
		BucketName:    "/usr/local/var/factors/cloud_storage",
	}
	taskDetailsJsonb, _ := U.EncodeStructTypeToPostgresJsonb(taskDetails)
	scheduledTask.TaskDetails = taskDetailsJsonb
	return scheduledTask, taskDetails
}

func getDummyValidBigqueryScheduledTask(projectID uint64) (M.ScheduledTask, M.BigqueryUploadTaskDetails) {
	scheduledTask := M.ScheduledTask{
		ProjectID:     projectID,
		JobID:         U.GetUUID(),
		TaskType:      M.TASK_TYPE_BIGQUERY_UPLOAD,
		TaskStartTime: U.TimeNowUnix(),
		TaskStatus:    M.TASK_STATUS_IN_PROGRESS,
	}

	taskDetails := M.BigqueryUploadTaskDetails{
		BigqueryProjectID: "factors-development",
		BigqueryDataset:   "envets_archive",
		BigqueryTable:     "f_events",
	}
	taskDetailsJsonb, _ := U.EncodeStructTypeToPostgresJsonb(taskDetails)
	scheduledTask.TaskDetails = taskDetailsJsonb
	return scheduledTask, taskDetails
}
