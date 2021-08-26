package postgres

import (
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"net/http"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

/*
RegisterTaskDependency: Given a taskId and what this task is dependent on, make an entry in task_dependency table
	Offset: After completion of parent job, how long should i wait to start the child job

DeregisterTaskDependency: Remove the already established dependency

GetAllDependentTasks: Get all the jobs on which this particular job is dependent on

IsDependencyCircular: This is used to find if the dependency that we are trying to register is circular
*/
func (pg *Postgres) RegisterTaskDependency(taskId uint64, dependentTaskId uint64, offset int) (int, string) {
	logCtx := log.WithFields(log.Fields{"task_id": taskId, "dependentTaskId": dependentTaskId, "offset": offset})

	if taskId == 0 || dependentTaskId == 0 {
		logCtx.Error("Missing required field.")
		return http.StatusBadRequest, "Missing task id/dependent task id"
	}

	db := C.GetServices().Db

	taskDependencyDetails := model.TaskExecutionDependencyDetails{
		TaskID:           taskId,
		DependentTaskID:  dependentTaskId,
		DependencyOffset: offset,
		CreatedAt:        U.TimeNowZ(),
		UpdatedAt:        U.TimeNowZ(),
	}

	baseTask, code1, _ := pg.GetTaskDetailsById(taskId)
	dependentTask, code2, _ := pg.GetTaskDetailsById(dependentTaskId)
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
	if pg.IsDependencyCircular(taskId, dependentTaskId) {
		return http.StatusBadRequest, "Circular Dependency detected"
	}
	if err := db.Create(&taskDependencyDetails).Error; err != nil {
		if U.IsPostgresUniqueIndexViolationError("task_id_dependent_task_id_unique_idx", err) {
			logCtx.Error("TaskName-Dependency already exist")
			return http.StatusConflict, "TaskName-Dependency already exist"
		} else {
			logCtx.Error(err.Error())
			return http.StatusConflict, err.Error()
		}
	}
	return http.StatusCreated, ""
}

func (pg *Postgres) DeregisterTaskDependency(taskId uint64, dependentTaskId uint64) (int, string) {
	logCtx := log.WithFields(log.Fields{"task_id": taskId, "dependentTaskId": dependentTaskId})

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

func (pg *Postgres) GetAllDependentTasks(taskID uint64) ([]model.TaskExecutionDependencyDetails, int, string) {
	logCtx := log.WithFields(log.Fields{"taskID": taskID})
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
			logCtx.Error("record with this task id is not found")
			return taskDeltas, http.StatusNotFound, "record with this taskID is not found"
		}
		logCtx.Error(err.Error())
		return taskDeltas, http.StatusInternalServerError, err.Error()
	}
	return taskDeltas, http.StatusOK, ""
}

func (pg *Postgres) IsDependencyCircular(taskId, dependentTaskId uint64) bool {
	dependencyTasksChain := make([]uint64, 0)
	dependencyTasksChain = append(dependencyTasksChain, dependentTaskId)
	for {
		if len(dependencyTasksChain) == 0 {
			return false
		}
		if dependencyTasksChain[0] == taskId {
			return true
		}
		depTasks, _, _ := pg.GetAllDependentTasks(dependencyTasksChain[0])
		for _, depTask := range depTasks {
			dependencyTasksChain = append(dependencyTasksChain, depTask.DependentTaskID)
		}
		dependencyTasksChain = dependencyTasksChain[1:len(dependencyTasksChain)]
	}
}
