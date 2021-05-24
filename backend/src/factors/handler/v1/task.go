package v1

import (
	"factors/model/store"
	"net/http"

	"github.com/gin-gonic/gin"
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
		return nil, http.StatusBadRequest, INVALID_INPUT, "", true
	}
	errCode, msg := store.GetStore().RegisterTaskDependency(params.TaskId, params.DependentTaskId, params.Delay)
	if errCode != http.StatusCreated {
		return nil, errCode, msg, "", true
	}
	response := make(map[string]interface{})
	response["status"] = "success"
	return response, http.StatusCreated, "", "", false
}
