package postgres

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	C "factors/config"
	"factors/model/model"
	U "factors/util"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

// CreateScheduledTask Creates a new task entry in scheduled_tasks table.
func (pg *Postgres) CreateScheduledTask(task *model.ScheduledTask) int {
	logCtx := log.WithFields(log.Fields{
		"Prefix":    "Model#model.ScheduledTask#Create",
		"ProjectID": task.ProjectID,
		"JobID":     task.JobID,
		"JobType":   task.TaskType,
	})

	err := validateScheduledTask(task)
	if err != nil {
		logCtx.WithError(err).Error("Task validattion failed.")
		return http.StatusBadRequest
	}

	db := C.GetServices().Db
	err = db.Create(&task).Error
	if err != nil {
		logCtx.WithError(err).Errorf("Failed to create ScheduledTask.")
		return http.StatusInternalServerError
	}

	return http.StatusCreated
}

// UpdateScheduledTask Updates runtime details for the task.
func (pg *Postgres) UpdateScheduledTask(taskID string, taskDetails *postgres.Jsonb, endTime int64, status model.ScheduledTaskStatus) (int64, int) {
	logCtx := log.WithFields(log.Fields{
		"Prefix": "Model#Update",
		"TaskID": taskID,
	})

	updates := map[string]interface{}{
		"task_end_time": endTime,
		"task_status":   status,
	}
	if taskDetails != nil {
		updates["task_details"] = taskDetails
	}

	db := C.GetServices().Db
	db = db.Model(&model.ScheduledTask{}).Where("id = ?", taskID).
		Updates(updates)

	if db.Error != nil {
		logCtx.WithError(db.Error).Error("UpdateScheduledTask Failed.")
		return 0, http.StatusInternalServerError
	}

	return db.RowsAffected, http.StatusAccepted
}

// GetScheduledTaskByID To get scheduled task by task id.
func (pg *Postgres) GetScheduledTaskByID(taskID string) (*model.ScheduledTask, int) {
	logCtx := log.WithFields(log.Fields{
		"Prefix": "Model#model.ScheduledTask#GetByID",
		"TaskID": taskID,
	})
	db := C.GetServices().Db

	var scheduledTask model.ScheduledTask
	if err := db.Where("id = ?", taskID).First(&scheduledTask).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}
		logCtx.WithError(err).Errorf("Failed to get task %s from database", taskID)
		return nil, http.StatusInternalServerError
	}
	return &scheduledTask, http.StatusFound
}

// GetScheduledTaskInProgressCount Returns the count of IN_PROGRESS jobs for particular task.
func (pg *Postgres) GetScheduledTaskInProgressCount(projectID uint64, taskType model.ScheduledTaskType) (int64, int) {
	logCtx := log.WithFields(log.Fields{
		"Prefix": "Model#model.ScheduledTask#GetInProgress",
	})
	db := C.GetServices().Db

	var inProgressCount int64
	db = db.Model(&model.ScheduledTask{}).
		Where("project_id = ? AND task_status = ? AND task_type = ?", projectID, model.TASK_STATUS_IN_PROGRESS, taskType).
		Count(&inProgressCount)

	if db.Error != nil {
		logCtx.WithError(db.Error).Error("Failed to get in progress count")
		return inProgressCount, http.StatusInternalServerError
	}
	return inProgressCount, http.StatusFound
}

// GetScheduledTaskLastRunTimestamp To get the timestamp of last for project and task_type.
func (pg *Postgres) GetScheduledTaskLastRunTimestamp(projectID uint64, taskType model.ScheduledTaskType) (int64, int) {
	db := C.GetServices().Db
	var maxTaskStartTime sql.NullInt64

	row := db.Model(&model.ScheduledTask{}).
		Where("project_id = ? AND task_type = ? AND task_status = ?", projectID, taskType, model.TASK_STATUS_SUCCESS).
		Select("MAX((task_details->>'to_timestamp')::bigint)").Row()
	err := row.Scan(&maxTaskStartTime)
	if err != nil {
		if gorm.IsRecordNotFoundError(err) || maxTaskStartTime.Valid {
			return 0, http.StatusNotFound
		}
		log.WithError(err).Errorf("Failed to get last run timestamp for projectID %d", projectID)
		return maxTaskStartTime.Int64, http.StatusInternalServerError
	}

	return maxTaskStartTime.Int64, http.StatusFound
}

// GetArchivalFileNamesForProject Get archived fileNames for a time range.
func (pg *Postgres) GetArchivalFileNamesForProject(projectID uint64, startTime, endTime time.Time) ([]string, []string, int) {
	db := C.GetServices().Db

	fileNames := make([]string, 0, 0)
	userFileNames := make([]string, 0, 0)
	rows, err := db.Model(&model.ScheduledTask{}).
		Where("project_id = ? AND task_type = ? AND (task_details->>'file_created')::bool=true"+
			" AND (task_details->>'from_timestamp')::bigint between ? AND ?", projectID, model.TASK_TYPE_EVENTS_ARCHIVAL, startTime.Unix(), endTime.Unix()).
		Select("task_details->>'filepath', task_details->>'users_filepath'").Rows()
	if err != nil {
		log.WithError(err).Error("Failed to get archived file paths")
		return fileNames, userFileNames, http.StatusInternalServerError
	}

	for rows.Next() {
		var fileName string
		var userFileName sql.NullString
		err = rows.Scan(&fileName, &userFileName)
		if err != nil {
			log.WithError(err).Error("Error while scanning row")
			return fileNames, userFileNames, http.StatusInternalServerError
		}
		fileNames = append(fileNames, fileName)
		userFileNames = append(userFileNames, userFileName.String)
	}
	return fileNames, userFileNames, http.StatusFound
}

// GetNewArchivalFileNamesAndEndTimeForProject Lists names of files created during archival process.
func (pg *Postgres) GetNewArchivalFileNamesAndEndTimeForProject(projectID uint64,
	lastRunAt int64, hardStartTime, hardEndTime time.Time) (map[int64]map[string]interface{}, int) {

	db := C.GetServices().Db
	var startTime, endTime int64
	fileNameEndTimeMap := make(map[int64]map[string]interface{})

	// Use hard start and end time if provided else lastRunAt.
	if !hardStartTime.IsZero() {
		startTime = hardStartTime.Unix()
		endTime = hardEndTime.Unix()
	} else {
		startTime = lastRunAt + 1 // Query is inclusive so increment by 1.
		endTime = U.TimeNowUnix()
	}

	rows, err := db.Model(&model.ScheduledTask{}).
		Where("project_id = ? AND task_type = ? AND (task_details->>'file_created')::bool=true"+
			" AND (task_details->>'from_timestamp')::bigint between ? AND ?", projectID, model.TASK_TYPE_EVENTS_ARCHIVAL, startTime, endTime).
		Select("id, task_details->>'filepath', task_details->>'users_filepath', task_details->>'from_timestamp', task_details->>'to_timestamp'").Rows()
	if err != nil {
		log.WithError(err).Error("Query failed to get filenames")
		return fileNameEndTimeMap, http.StatusInternalServerError
	}

	for rows.Next() {
		var fileName, taskID string
		var usersFileName sql.NullString
		var startTime, endTime int64
		err = rows.Scan(&taskID, &fileName, &usersFileName, &startTime, &endTime)
		if err != nil {
			log.WithError(err).Error("Error while scanning row")
			continue
		}
		fileNameEndTimeMap[endTime] = make(map[string]interface{})
		fileNameEndTimeMap[endTime]["filepath"] = fileName
		fileNameEndTimeMap[endTime]["users_filepath"] = usersFileName.String
		fileNameEndTimeMap[endTime]["task_id"] = taskID
		fileNameEndTimeMap[endTime]["start_time"] = startTime
	}

	if len(fileNameEndTimeMap) > 0 {
		log.Infof("Filtering out completed bigquery task in range %d-%d", startTime, endTime)
		fileNameEndTimeMap, err = filterCompletedBigqueryTasks(fileNameEndTimeMap, projectID)
		if err != nil {
			log.WithError(err).Error("Failed to filter completed tasks")
			return fileNameEndTimeMap, http.StatusInternalServerError
		}
	}

	return fileNameEndTimeMap, http.StatusFound
}

// FailScheduleTask Set status as FAILED for the given taskID.
func (pg *Postgres) FailScheduleTask(taskID string) {
	rowsUpdated, status := pg.UpdateScheduledTask(taskID, nil, U.TimeNowUnix(), model.TASK_STATUS_FAILED)
	if status != http.StatusAccepted || rowsUpdated == 0 {
		log.Errorf("Error while marking task %s as failed", taskID)
	}
}

// GetCompletedArchivalBatches Returns completed archival batches for a given range.
func (pg *Postgres) GetCompletedArchivalBatches(projectID uint64, startTime, endTime time.Time) (map[int64]int64, int) {
	db := C.GetServices().Db
	completedBatches := make(map[int64]int64)

	rows, err := db.Model(&model.ScheduledTask{}).
		Where("project_id = ? AND task_status = ? AND task_type = ? AND (task_details->>'from_timestamp')::bigint BETWEEN ? AND ?",
			projectID, model.TASK_STATUS_SUCCESS, model.TASK_TYPE_EVENTS_ARCHIVAL, startTime.Unix(), endTime.Unix()).
		Select("task_details->>'from_timestamp', task_details->>'to_timestamp'").Rows()
	if err != nil {
		log.WithError(err).Error("Failed to get completed archival tasks")
		return completedBatches, http.StatusInternalServerError
	}

	for rows.Next() {
		var fromTimestamp, toTimestamp int64
		if err := rows.Scan(&fromTimestamp, &toTimestamp); err != nil {
			log.WithError(err).Error("Error while scanning completed archival timestamps")
			return completedBatches, http.StatusInternalServerError
		}
		completedBatches[fromTimestamp] = toTimestamp
	}
	return completedBatches, http.StatusFound
}

// Filters all completed bigquery tasks from the list of allTasksMap.
func filterCompletedBigqueryTasks(allTasksMap map[int64]map[string]interface{}, projectID uint64) (map[int64]map[string]interface{}, error) {
	var archivalTaskIDs, completedTaskIDs, pendingTaskIDs []string
	pendingTasksMap := make(map[int64]map[string]interface{})
	db := C.GetServices().Db

	for _, value := range allTasksMap {
		archivalTaskIDs = append(archivalTaskIDs, value["task_id"].(string))
	}

	rows, err := db.Model(&model.ScheduledTask{}).
		Where("project_id = ? AND task_status = ? AND task_type = ? AND task_details->>'archival_task_id' in (?)",
			projectID, model.TASK_STATUS_SUCCESS, model.TASK_TYPE_BIGQUERY_UPLOAD, archivalTaskIDs).
		Select("task_details->>'archival_task_id'").Rows()
	if err != nil {
		log.WithError(err).Errorf("Failed to get completed bigquery tasks list")
		return pendingTasksMap, err
	}

	for rows.Next() {
		var completedTaskID string
		if err := rows.Scan(&completedTaskID); err != nil {
			log.WithError(err).Errorf("Error while scanning bigquery completed task id")
			return pendingTasksMap, err
		}
		completedTaskIDs = append(completedTaskIDs, completedTaskID)
	}
	pendingTaskIDs = U.StringSliceDiff(archivalTaskIDs, completedTaskIDs)

	for key, value := range allTasksMap {
		if U.StringValueIn(value["task_id"].(string), pendingTaskIDs) {
			pendingTasksMap[key] = value
		} else {
			log.Infof("Filtering task %v", value)
		}
	}
	return pendingTasksMap, nil
}

func validateScheduledTask(task *model.ScheduledTask) error {
	var validationError error
	if task.ProjectID == 0 {
		validationError = fmt.Errorf("Invalid project id %d", task.ProjectID)
	} else if task.JobID == "" {
		validationError = fmt.Errorf("JobID not initialized for task")
	} else if task.TaskType != model.TASK_TYPE_EVENTS_ARCHIVAL && task.TaskType != model.TASK_TYPE_BIGQUERY_UPLOAD {
		validationError = fmt.Errorf("Invalid TaskType")
	}

	return validationError
}
