package handler

import (
	M "factors/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Test command.
// curl -i -X GET http://localhost:8080/projects/1/event_names
func GetEventNamesHandler(c *gin.Context) {
	projectId, err := strconv.ParseUint(c.Params.ByName("project_id"), 10, 64)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	ens, errCode := M.GetEventNames(projectId)
	if errCode != M.DB_SUCCESS {
		c.AbortWithStatus(errCode)
	} else {
		eventNames := []string{}
		for _, en := range ens {
			eventNames = append(eventNames, en.Name)
		}
		c.JSON(http.StatusOK, eventNames)
	}
}
