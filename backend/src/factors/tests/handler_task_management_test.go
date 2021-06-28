package tests

import (
	H "factors/handler"
	"factors/model/model"
	"factors/model/store"
	"net/http"
	"testing"
	"time"
	"fmt"
	U "factors/util"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestDeleteTaskEndRecord(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	taskName:= fmt.Sprintf("%v_%v_1", "task_event", U.TimeNow().Unix())
	task_id, code, message := store.GetStore().RegisterTaskWithDefaultConfiguration(taskName, "Source", model.Hourly, false)
	assert.Equal(t, http.StatusCreated, code)
	assert.Equal(t, "", message)
	endDate := uint64(time.Date(2021, 4, 2, 0, 0, 0, 0, time.UTC).Unix())
	code, message = store.GetStore().InsertTaskBeginRecord(task_id,0,endDate)
	assert.Equal(t, http.StatusCreated, code)
	assert.Equal(t, "", message)

	status, err1 := store.GetStore().DeleteTaskEndRecord(task_id, 0, endDate)
	assert.NotNil(t, err1)
	assert.Equal(t, 202, status)

}
func TestGetAllProcessedIntervals(t *testing.T){
	r := gin.Default()
	H.InitAppRoutes(r)
	taskName:= fmt.Sprintf("%v_%v_1", "task_event", U.TimeNow().Unix())
	task_id, code, message := store.GetStore().RegisterTaskWithDefaultConfiguration(taskName, "Source", model.Hourly, false)
	assert.Equal(t, http.StatusCreated, code)
	assert.Equal(t, "", message)
	endDate := uint64(time.Date(2021, 4, 2, 0, 0, 0, 0, time.UTC).Unix())
	code, message = store.GetStore().InsertTaskBeginRecord(task_id,0,endDate)
	assert.Equal(t, http.StatusCreated, code)
	assert.Equal(t, "", message)

	_,status, err1 := store.GetStore().GetAllProcessedIntervals(task_id, 0, 123,&time.Time{})
	assert.Equal(t, status, http.StatusOK)
	assert.Equal(t,err1,"")
	assert.NotNil(t, err1)
	_,status,err1 = store.GetStore().GetAllProcessedIntervals(0,1,123,&time.Time{})
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Equal(t,err1,"missing taskID")
	status, err1= store.GetStore().DeleteTaskEndRecord(task_id, 0, endDate)
	assert.NotNil(t, err1)
	assert.Equal(t, 202, status)

}
