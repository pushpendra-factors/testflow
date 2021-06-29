package v1

import (
	"factors/model/store"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
)

type RegisterTaskParams struct {
	TaskName         string `json:"task_name"`
	Frequency        int    `json:"frequency"`
	Source           string `json:"source"`
	IsProjectEnabled bool   `json:"is_project_enabled"`
}

func GetRegisterTaskParams(c *gin.Context) (*RegisterTaskParams, error) {
	params := RegisterTaskParams{}
	err := c.BindJSON(&params)
	if err != nil {
		return nil, err
	}
	return &params, nil
}

func RegisterTaskHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	params, err := GetRegisterTaskParams(c)
	if err != nil {
		return nil, http.StatusBadRequest, INVALID_INPUT, "", true
	}
	taskId, errCode, msg := store.GetStore().RegisterTaskWithDefaultConfiguration(params.TaskName, params.Source, params.Frequency, params.IsProjectEnabled)
	if errCode != http.StatusCreated {
		return nil, errCode, msg, "", true
	}
	response := make(map[string]interface{})
	response["status"] = "success"
	response["taskId"] = taskId
	return response, http.StatusCreated, "", "", false
}

type RegisterTaskDependencyParams struct {
	TaskId          uint64 `json:"task_id"`
	DependentTaskId uint64 `json:"dependent_task_id"`
	Delay           int    `json:"delay"`
}
type AllProcessedIntervalsParams struct {
	TaskId    uint64 `json:"task_id"`
	ProjectId uint64 `json:"project_id"`
	Lookback  int    `json:"lookback"`
}
type DeleteTaskEndRecord struct {
	TaskId    uint64 `json:"task_id"`
	ProjectId uint64 `json:"project_id"`
	Delta     uint64 `json:"delta"`
}

func GetRegisterTaskDependencyParams(c *gin.Context) (*RegisterTaskDependencyParams, error) {
	params := RegisterTaskDependencyParams{}
	err := c.BindJSON(&params)
	if err != nil {
		return nil, err
	}
	return &params, nil
}

func RegisterTaskDependencyHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	params, err := GetRegisterTaskDependencyParams(c)
	if err != nil {
		return nil, http.StatusBadRequest, INVALID_INPUT, err.Error(), true
	}
	errCode, msg := store.GetStore().RegisterTaskDependency(params.TaskId, params.DependentTaskId, params.Delay)
	if errCode != http.StatusCreated {
		return nil, errCode, msg, "", true
	}
	response := make(map[string]interface{})
	response["status"] = "success"
	return response, http.StatusCreated, "", "", false
}
func GetAllProcessedIntervalsParams(c *gin.Context) (*AllProcessedIntervalsParams, error) {
	params := AllProcessedIntervalsParams{}
	TaskId, err := strconv.ParseUint(c.Query("task_id"), 10, 64)
	if err != nil {
		return nil, err
	}
	projectId, err := strconv.ParseUint(c.Query("project_id"), 10, 64)
	if err != nil {
		return nil, err
	}
	lookback, err := strconv.ParseInt(c.Query("lookback"), 10, 64)
	if err != nil {
		return nil, err
	}
	params.TaskId = TaskId
	params.ProjectId = projectId
	params.Lookback = int(lookback)

	return &params, nil
}

func GetAllProcessedIntervalsHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	params, err := GetAllProcessedIntervalsParams(c)
	if err != nil {
		return nil, http.StatusBadRequest, INVALID_INPUT, err.Error() + "1", true
	}
	time := time.Now()
	deltas, errCode, msg := store.GetStore().GetAllProcessedIntervals(params.TaskId, params.ProjectId, params.Lookback, &time)
	if errCode != http.StatusOK {
		return nil, errCode, msg, err.Error() + "2", true
	}
	c.JSON(200, deltas)
	response := make(map[string]interface{})
	response["status"] = "success"
	return response, http.StatusCreated, "", "", false
}
func GetDeleteTaskEndRecordParams(c *gin.Context) (*DeleteTaskEndRecord, error) {
	params := DeleteTaskEndRecord{}
	TaskId, err := strconv.ParseUint(c.Query("task_id"), 10, 64)
	if err != nil {
		return nil, err
	}
	projectId, err := strconv.ParseUint(c.Query("project_id"), 10, 64)
	if err != nil {
		return nil, err
	}
	delta, err := strconv.ParseUint(c.Query("delta"), 10, 64)
	if err != nil {
		return nil, err
	}
	params.TaskId = TaskId
	params.ProjectId = projectId
	params.Delta = delta

	return &params, nil
}
func DeleteTaskEndRecordHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	params, err := GetDeleteTaskEndRecordParams(c)
	if err != nil {
		return nil, http.StatusBadRequest, INVALID_INPUT, err.Error() + "1", true
	}
	errCode, msg := store.GetStore().DeleteTaskEndRecord(params.TaskId, params.ProjectId, params.Delta)
	if errCode != http.StatusAccepted {
		return nil, errCode, msg, "", true
	}
	response := make(map[string]interface{})
	response["status"] = "success"
	return response, http.StatusAccepted, "", "", false
}
