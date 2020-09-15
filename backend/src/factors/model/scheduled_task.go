package model

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	C "factors/config"
	U "factors/util"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

type ScheduledTaskType string

const (
	TASK_TYPE_EVENTS_ARCHIVAL ScheduledTaskType = "EVENTS_ARCHIVAL"
	TASK_TYPE_BIGQUERY_UPLOAD ScheduledTaskType = "BIGQUERY_UPLOAD"
)

type ScheduledTaskStatus string

const (
	TASK_STATUS_IN_PROGRESS ScheduledTaskStatus = "IN_PROGRESS"
	TASK_STATUS_SUCCESS     ScheduledTaskStatus = "SUCCESS"
	TASK_STATUS_FAILED      ScheduledTaskStatus = "FAILED"
)

// ScheduledTask Entity to store the details for scheduled tasks.
type ScheduledTask struct {
	ID            string              `gorm:"primary_key:true;uuid;default:uuid_generate_v4()" json:"id"` // Id for the current task.
	JobID         string              `gorm:"not null" json:"job_id"`                                     // Id for the parent task.
	ProjectID     uint64              `gorm:"not null" json:"project_id"`
	TaskType      ScheduledTaskType   `gorm:"not null" json:"task_type"`
	TaskStatus    ScheduledTaskStatus `gorm:"not null" json:"task_status"`
	TaskStartTime int64               `gorm:"default:null" json:"task_start_time"` // Time when tast run started.
	TaskEndTime   int64               `gorm:"default:null" json:"task_end_time"`   // Time when task run ended.
	TaskDetails   *postgres.Jsonb     `json:"task_details"`                        // Metadata for the task.
	CreatedAt     time.Time           `json:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at"`
}

// EventArchivalTaskDetails To store metadata for individual task run.
type EventArchivalTaskDetails struct {
	FromTimestamp int64  `json:"from_timestamp"`
	ToTimestamp   int64  `json:"to_timestamp"`
	EventCount    int64  `json:"event_count"`
	FileCreated   bool   `json:"file_created"`
	FilePath      string `json:"filepath"`
	UsersFilePath string `json:"users_filepath"`
	BucketName    string `json:"bucket_name"`
}

// BigqueryUploadTaskDetails To store metadata for bigquery upload tasks.
type BigqueryUploadTaskDetails struct {
	FromTimestamp     int64           `json:"from_timestamp"`
	ToTimestamp       int64           `json:"to_timestamp"`
	BigqueryProjectID string          `json:"bq_project_id"`
	BigqueryDataset   string          `json:"bq_dataset"`
	BigqueryTable     string          `json:"bq_table"`
	ArchivalTaskID    string          `json:"archival_task_id"`
	UploadStats       *postgres.Jsonb `json:"upload_stats"`
	UsersUploadStats  *postgres.Jsonb `json:"users_supload_stats"`
}

// CreateScheduledTask Creates a new task entry in scheduled_tasks table.
func CreateScheduledTask(task *ScheduledTask) int {
	logCtx := log.WithFields(log.Fields{
		"Prefix":    "Model#ScheduledTask#Create",
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
func UpdateScheduledTask(taskID string, taskDetails *postgres.Jsonb, endTime int64, status ScheduledTaskStatus) (int64, int) {
	logCtx := log.WithFields(log.Fields{
		"Prefix": "Model#ScheduledTask#Update",
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
	db = db.Model(&ScheduledTask{}).Where("id = ?", taskID).
		Updates(updates)

	if db.Error != nil {
		logCtx.WithError(db.Error).Error("UpdateScheduledTask Failed.")
		return 0, http.StatusInternalServerError
	}

	return db.RowsAffected, http.StatusAccepted
}

// GetScheduledTaskByID To get scheduled task by task id.
func GetScheduledTaskByID(taskID string) (*ScheduledTask, int) {
	logCtx := log.WithFields(log.Fields{
		"Prefix": "Model#ScheduledTask#GetByID",
		"TaskID": taskID,
	})
	db := C.GetServices().Db

	var scheduledTask ScheduledTask
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
func GetScheduledTaskInProgressCount(projectID uint64, taskType ScheduledTaskType) (int64, int) {
	logCtx := log.WithFields(log.Fields{
		"Prefix": "Model#ScheduledTask#GetInProgress",
	})
	db := C.GetServices().Db

	var inProgressCount int64
	db = db.Model(&ScheduledTask{}).
		Where("project_id = ? AND task_status = ? AND task_type = ?", projectID, TASK_STATUS_IN_PROGRESS, taskType).
		Count(&inProgressCount)

	if db.Error != nil {
		logCtx.WithError(db.Error).Error("Failed to get in progress count")
		return inProgressCount, http.StatusInternalServerError
	}
	return inProgressCount, http.StatusFound
}

// GetScheduledTaskLastRunTimestamp To get the timestamp of last for project and task_type.
func GetScheduledTaskLastRunTimestamp(projectID uint64, taskType ScheduledTaskType) (int64, int) {
	db := C.GetServices().Db
	var maxTaskStartTime sql.NullInt64

	row := db.Model(&ScheduledTask{}).
		Where("project_id = ? AND task_type = ? AND task_status = ?", projectID, taskType, TASK_STATUS_SUCCESS).
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
func GetArchivalFileNamesForProject(projectID uint64, startTime, endTime time.Time) ([]string, int) {
	db := C.GetServices().Db

	fileNames := make([]string, 0, 0)
	rows, err := db.Model(&ScheduledTask{}).
		Where("project_id = ? AND task_type = ? AND (task_details->>'file_created')::bool=true"+
			" AND (task_details->>'from_timestamp')::bigint between ? AND ?", projectID, TASK_TYPE_EVENTS_ARCHIVAL, startTime.Unix(), endTime.Unix()).
		Select("task_details->>'filepath'").Rows()
	if err != nil {
		log.WithError(err).Error("Failed to get archived file paths")
		return fileNames, http.StatusInternalServerError
	}

	for rows.Next() {
		var fileName string
		err = rows.Scan(&fileName)
		if err != nil {
			log.WithError(err).Error("Error while scanning row")
			return fileNames, http.StatusInternalServerError
		}
		fileNames = append(fileNames, fileName)
	}
	return fileNames, http.StatusFound
}

// GetNewArchivalFileNamesAndEndTimeForProject Lists names of files created during archival process.
func GetNewArchivalFileNamesAndEndTimeForProject(projectID uint64,
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

	rows, err := db.Model(&ScheduledTask{}).
		Where("project_id = ? AND task_type = ? AND (task_details->>'file_created')::bool=true"+
			" AND (task_details->>'from_timestamp')::bigint between ? AND ?", projectID, TASK_TYPE_EVENTS_ARCHIVAL, startTime, endTime).
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
func FailScheduleTask(taskID string) {
	rowsUpdated, status := UpdateScheduledTask(taskID, nil, U.TimeNowUnix(), TASK_STATUS_FAILED)
	if status != http.StatusAccepted || rowsUpdated == 0 {
		log.Errorf("Error while marking task %s as failed", taskID)
	}
}

// GetCompletedArchivalBatches Returns completed archival batches for a given range.
func GetCompletedArchivalBatches(projectID uint64, startTime, endTime time.Time) (map[int64]int64, int) {
	db := C.GetServices().Db
	completedBatches := make(map[int64]int64)

	rows, err := db.Model(&ScheduledTask{}).
		Where("project_id = ? AND task_status = ? AND task_type = ? AND (task_details->>'from_timestamp')::bigint BETWEEN ? AND ?",
			projectID, TASK_STATUS_SUCCESS, TASK_TYPE_EVENTS_ARCHIVAL, startTime.Unix(), endTime.Unix()).
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

	rows, err := db.Model(&ScheduledTask{}).
		Where("project_id = ? AND task_status = ? AND task_type = ? AND task_details->>'archival_task_id' in (?)",
			projectID, TASK_STATUS_SUCCESS, TASK_TYPE_BIGQUERY_UPLOAD, archivalTaskIDs).
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

func validateScheduledTask(task *ScheduledTask) error {
	var validationError error
	if task.ProjectID == 0 {
		validationError = fmt.Errorf("Invalid project id %d", task.ProjectID)
	} else if task.JobID == "" {
		validationError = fmt.Errorf("JobID not initialized for task")
	} else if task.TaskType != TASK_TYPE_EVENTS_ARCHIVAL && task.TaskType != TASK_TYPE_BIGQUERY_UPLOAD {
		validationError = fmt.Errorf("Invalid TaskType")
	}

	return validationError
}
