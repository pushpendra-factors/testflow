package memsql

import (
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) RegisterTaskDependency(taskId uint64, dependentTaskId uint64, offset int) (int, string) {
	logFields := log.Fields{
		"task_id":           taskId,
		"dependent_task_id": dependentTaskId,
		"offset":            offset,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	if taskId == 0 || dependentTaskId == 0 {
		logCtx.Error("Missing required field.")
		return http.StatusBadRequest, "Missing task id/dependent task id"
	}

	db := C.GetServices().Db

	taskDependencyDetails := model.TaskExecutionDependencyDetails{
		ID:               U.GetUUID(),
		TaskID:           taskId,
		DependentTaskID:  dependentTaskId,
		DependencyOffset: offset,
		CreatedAt:        U.TimeNowZ(),
		UpdatedAt:        U.TimeNowZ(),
	}

	baseTask, code1, _ := store.GetTaskDetailsById(taskId)
	dependentTask, code2, _ := store.GetTaskDetailsById(dependentTaskId)
	if code1 != http.StatusOK || code2 != http.StatusOK {
		return http.StatusBadRequest, "Invalid taskID"
	}
	if baseTask.Frequency == model.Stateless {
		return http.StatusBadRequest, "BaseTask is stateless - so it cannot depend of any task"
	}
	if dependentTask.Frequency > baseTask.Frequency {
		return http.StatusBadRequest, "BaseTask is more frequent than the dependent task"
	}
	if !((dependentTask.SkipStartIndex <= baseTask.SkipStartIndex && dependentTask.SkipEndIndex >= baseTask.SkipEndIndex) ||
		(dependentTask.SkipStartIndex == 0 && dependentTask.SkipEndIndex == -1)) {
		return http.StatusBadRequest, "skip indexes of tasks doesnt match"
	}
	if !(dependentTask.Frequency <= baseTask.Frequency) {
		return http.StatusBadRequest, "Frequency of tasks doesnt match"
	}
	if store.IsDependencyCircular(taskId, dependentTaskId) {
		return http.StatusBadRequest, "Circular Dependency detected"
	}
	existingDep := make([]model.TaskExecutionDependencyDetails, 0)
	db = db.Where("task_id = ?", taskId).Where("dependent_task_id = ? ", dependentTaskId).Find(&existingDep)
	if len(existingDep) > 0 {
		logCtx.Error("TaskName-Dependency already exist")
		return http.StatusConflict, "TaskName-Dependency already exist"
	}
	if err := db.Create(&taskDependencyDetails).Error; err != nil {
		logCtx.WithError(err).Error("Failure in RegisterTaskDependency")
		return http.StatusConflict, err.Error()
	}
	return http.StatusCreated, ""
}

func (store *MemSQL) DeregisterTaskDependency(taskId uint64, dependentTaskId uint64) (int, string) {
	logFields := log.Fields{
		"task_id":           taskId,
		"dependent_task_id": dependentTaskId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	if taskId == 0 || dependentTaskId == 0 {
		logCtx.Error("Missing required field.")
		return http.StatusBadRequest, "Missing task id/dependent task id"
	}

	db := C.GetServices().Db

	db = db.Where("task_id = ?", taskId).Where("dependent_task_id = ? ", dependentTaskId).Delete(&model.TaskExecutionDependencyDetails{})

	if db.Error != nil {
		logCtx.WithError(db.Error).Error(
			"Deleting model.TaskExecutionDependencyDetails failed.")
		return http.StatusInternalServerError, ""
	}

	if db.RowsAffected == 0 {
		return http.StatusNotFound, "Record not found"
	}

	return http.StatusAccepted, ""
}

func (store *MemSQL) GetAllDependentTasks(taskID uint64) ([]model.TaskExecutionDependencyDetails, int, string) {
	logFields := log.Fields{
		"task_id": taskID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	taskDeltas := make([]model.TaskExecutionDependencyDetails, 0)
	if taskID == 0 {
		logCtx.Error("missing taskID")
		return taskDeltas, http.StatusBadRequest, "missing taskID"
	}

	taskNameFilter := &model.TaskExecutionDependencyDetails{
		TaskID: taskID,
	}

	db := C.GetServices().Db

	if err := db.Where(taskNameFilter).Find(&taskDeltas).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			logCtx.WithError(err).Error("record with this task id is not found")
			return taskDeltas, http.StatusNotFound, "record with this taskID is not found"
		}
		logCtx.Error(err.Error())
		return taskDeltas, http.StatusInternalServerError, err.Error()
	}
	return taskDeltas, http.StatusOK, ""
}

func (store *MemSQL) IsDependencyCircular(taskId, dependentTaskId uint64) bool {
	logFields := log.Fields{
		"task_id":           taskId,
		"dependent_task_id": dependentTaskId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	dependencyTasksChain := make([]uint64, 0)
	dependencyTasksChain = append(dependencyTasksChain, dependentTaskId)
	for {
		if len(dependencyTasksChain) == 0 {
			return false
		}
		if dependencyTasksChain[0] == taskId {
			return true
		}
		depTasks, _, _ := store.GetAllDependentTasks(dependencyTasksChain[0])
		for _, depTask := range depTasks {
			dependencyTasksChain = append(dependencyTasksChain, depTask.DependentTaskID)
		}
		dependencyTasksChain = dependencyTasksChain[1:len(dependencyTasksChain)]
	}
}
