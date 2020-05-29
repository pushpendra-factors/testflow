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
	BucketName    string `json:"bucket_name"`
}

// BigqueryUploadTaskDetails To store metadata for bigquery upload tasks.
type BigqueryUploadTaskDetails struct {
	BigqueryProjectID string          `json:"bq_project_id"`
	BigqueryDataset   string          `json:"bq_dataset"`
	BigqueryTable     string          `json:"bq_table"`
	ArchivalTaskID    string          `json:"archival_task_id"`
	UploadStats       *postgres.Jsonb `json:"upload_stats"`
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

// GetNewArchivalFileNamesAndEndTimeForProject Lists names of files created during archival process.
func GetNewArchivalFileNamesAndEndTimeForProject(projectID uint64,
	lastRunAt int64) (map[int64]map[string]string, int) {

	db := C.GetServices().Db
	fileNameEndTimeMap := make(map[int64]map[string]string)

	rows, err := db.Model(&ScheduledTask{}).
		Where("project_id = ? AND task_type = ? AND (task_details->>'file_created')::bool=true"+
			" AND (task_details->>'from_timestamp')::bigint > ?", projectID, TASK_TYPE_EVENTS_ARCHIVAL, lastRunAt).
		Select("id, task_details->>'filepath', task_details->>'to_timestamp'").Rows()
	if err != nil {
		log.WithError(err).Error("Query failed to get filenames")
		return fileNameEndTimeMap, http.StatusInternalServerError
	}

	for rows.Next() {
		var fileName, taskID string
		var endTime int64
		err = rows.Scan(&taskID, &fileName, &endTime)
		if err != nil {
			log.WithError(err).Error("Error while scanning row")
			continue
		}
		fileNameEndTimeMap[endTime] = make(map[string]string)
		fileNameEndTimeMap[endTime]["filepath"] = fileName
		fileNameEndTimeMap[endTime]["task_id"] = taskID
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

// GetMinStartTimeForTaskType Returns the oldest run timestamp for the task type.
func GetMinStartTimeForTaskType(projectID uint64, taskType ScheduledTaskType) (int64, int) {
	db := C.GetServices().Db
	var minStartTime sql.NullInt64

	row := db.Model(&ScheduledTask{}).
		Where("project_id = ? AND task_type = ?", projectID, taskType).
		Select("MIN((task_details->>'from_timestamp')::bigint)").Row()
	err := row.Scan(&minStartTime)
	if err != nil {
		if gorm.IsRecordNotFoundError(err) || minStartTime.Valid {
			return 0, http.StatusNotFound
		}
		log.WithError(err).Errorf("Failed to get min start time for project id %d", projectID)
		return minStartTime.Int64, http.StatusInternalServerError
	}

	return minStartTime.Int64, http.StatusFound
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
