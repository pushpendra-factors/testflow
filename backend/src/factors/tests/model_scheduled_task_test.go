package tests

import (
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreateScheduledTask(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	scheduledTask := model.ScheduledTask{
		ProjectID:     project.ID,
		TaskType:      model.TASK_TYPE_EVENTS_ARCHIVAL,
		TaskStartTime: U.TimeNowUnix(),
	}

	// With empty JobID. Should fail.
	status := store.GetStore().CreateScheduledTask(&scheduledTask)
	assert.Equal(t, status, http.StatusBadRequest)

	scheduledTask.JobID = U.GetUUID()
	status = store.GetStore().CreateScheduledTask(&scheduledTask)
	assert.Equal(t, status, http.StatusCreated)

	scheduledTaskDB, _ := store.GetStore().GetScheduledTaskByID(scheduledTask.ID)
	assert.NotNil(t, scheduledTaskDB)
	assert.Equal(t, scheduledTask.ID, scheduledTaskDB.ID)
}

func TestUpdateScheduledTaskDetailsAndEndTime(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	scheduledTask, taskDetails := getDummyValidArchivalScheduledTask(project.ID)
	status := store.GetStore().CreateScheduledTask(&scheduledTask)
	assert.Equal(t, status, http.StatusCreated)

	taskDetails.FileCreated = false
	taskDetailsJsonb, _ := U.EncodeStructTypeToPostgresJsonb(taskDetails)
	taskEndTime := U.TimeNowUnix() + 1
	rowsUpdated, status := store.GetStore().UpdateScheduledTask(scheduledTask.ID, taskDetailsJsonb, taskEndTime, model.TASK_STATUS_SUCCESS)
	assert.Equal(t, status, http.StatusAccepted)
	assert.NotEqual(t, rowsUpdated, 0)

	scheduledTaskDB, _ := store.GetStore().GetScheduledTaskByID(scheduledTask.ID)
	assert.NotNil(t, scheduledTaskDB)
	assert.Equal(t, scheduledTask.ID, scheduledTaskDB.ID)
	assert.Equal(t, taskEndTime, scheduledTaskDB.TaskEndTime)
	assert.Equal(t, model.TASK_STATUS_SUCCESS, scheduledTaskDB.TaskStatus)

	expectedTaskDetails, _ := U.DecodePostgresJsonb(taskDetailsJsonb)
	actualTaskDetails, _ := U.DecodePostgresJsonb(scheduledTaskDB.TaskDetails)
	assert.Equal(t, expectedTaskDetails, actualTaskDetails)
}

func TestGetScheduledTaskInProgressCount(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	scheduledTask1, _ := getDummyValidBigqueryScheduledTask(project.ID)
	scheduledTask2, _ := getDummyValidBigqueryScheduledTask(project.ID)
	_ = store.GetStore().CreateScheduledTask(&scheduledTask1)
	_ = store.GetStore().CreateScheduledTask(&scheduledTask2)

	inProgressCount, status := store.GetStore().GetScheduledTaskInProgressCount(project.ID, model.TASK_TYPE_BIGQUERY_UPLOAD)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, int64(2), inProgressCount)
}

func TestGetScheduledTaskLastRunTimestamp(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	scheduledTask1, _ := getDummyValidArchivalScheduledTask(project.ID)
	scheduledTask1.TaskStatus = model.TASK_STATUS_SUCCESS

	expectedLastRunTimestamp := scheduledTask1.TaskStartTime + 86400
	scheduledTask2, _ := getDummyValidArchivalScheduledTask(project.ID)
	scheduledTask2.TaskStatus = model.TASK_STATUS_SUCCESS
	scheduledTask2.TaskStartTime = expectedLastRunTimestamp

	// Failed task should not be checked for LastRun timestamp even if run time is later.
	scheduledTask3, _ := getDummyValidArchivalScheduledTask(project.ID)
	scheduledTask3.TaskStatus = model.TASK_STATUS_FAILED
	scheduledTask3.TaskStartTime = scheduledTask1.TaskStartTime + 2*86400

	_ = store.GetStore().CreateScheduledTask(&scheduledTask1)
	_ = store.GetStore().CreateScheduledTask(&scheduledTask2)
	_ = store.GetStore().CreateScheduledTask(&scheduledTask3)

	actualLastRunTimestamp, status := store.GetStore().GetScheduledTaskLastRunTimestamp(
		project.ID, scheduledTask1.TaskType)
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

	_ = store.GetStore().CreateScheduledTask(&scheduledTask1)
	_ = store.GetStore().CreateScheduledTask(&scheduledTask2)
	_ = store.GetStore().CreateScheduledTask(&scheduledTask3)

	// Non inclusive check. Should include task1 and task3.
	bigQueryLastRunAt := taskDetails3.FromTimestamp - 1
	newFilesMap, status := store.GetStore().GetNewArchivalFileNamesAndEndTimeForProject(project.ID, bigQueryLastRunAt, time.Time{}, time.Time{})
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, 2, len(newFilesMap))

	// Inclusive check. Should include only task3.
	bigQueryLastRunAt = taskDetails3.FromTimestamp
	newFilesMap, status = store.GetStore().GetNewArchivalFileNamesAndEndTimeForProject(project.ID, bigQueryLastRunAt, time.Time{}, time.Time{})
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, 1, len(newFilesMap))
}

func getDummyValidArchivalScheduledTask(projectID int64) (model.ScheduledTask, model.EventArchivalTaskDetails) {
	scheduledTask := model.ScheduledTask{
		ProjectID:     projectID,
		JobID:         U.GetUUID(),
		TaskType:      model.TASK_TYPE_EVENTS_ARCHIVAL,
		TaskStartTime: U.TimeNowUnix(),
		TaskStatus:    model.TASK_STATUS_IN_PROGRESS,
	}

	taskDetails := model.EventArchivalTaskDetails{
		FromTimestamp: U.TimeNowUnix(),
		ToTimestamp:   U.TimeNowZ().AddDate(0, 0, 1).Unix(),
		EventCount:    10,
		BucketName:    "/usr/local/var/factors/cloud_storage",
	}
	taskDetailsJsonb, _ := U.EncodeStructTypeToPostgresJsonb(taskDetails)
	scheduledTask.TaskDetails = taskDetailsJsonb
	return scheduledTask, taskDetails
}

func getDummyValidBigqueryScheduledTask(projectID int64) (model.ScheduledTask, model.BigqueryUploadTaskDetails) {
	scheduledTask := model.ScheduledTask{
		ProjectID:     projectID,
		JobID:         U.GetUUID(),
		TaskType:      model.TASK_TYPE_BIGQUERY_UPLOAD,
		TaskStartTime: U.TimeNowUnix(),
		TaskStatus:    model.TASK_STATUS_IN_PROGRESS,
	}

	taskDetails := model.BigqueryUploadTaskDetails{
		BigqueryProjectID: "factors-development",
		BigqueryDataset:   "envets_archive",
		BigqueryTable:     "f_events",
	}
	taskDetailsJsonb, _ := U.EncodeStructTypeToPostgresJsonb(taskDetails)
	scheduledTask.TaskDetails = taskDetailsJsonb
	return scheduledTask, taskDetails
}
