package handler

import (
	"encoding/json"
	M "model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Test command.
// curl -H "Content-Type: application/json" -i -X POST http://localhost:8080/projects/1/users -d '{ "user_id": "murthy@autometa", "properties": {"city": "bengaluru", "mobile": true}}'
func CreateUserHandler(c *gin.Context) {
	r := c.Request

	projectId, err := strconv.ParseUint(c.Params.ByName("project_id"), 10, 64)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var user M.User
	err = json.NewDecoder(r.Body).Decode(&user)
	user.ProjectId = projectId
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "json decoding : " + err.Error(),
			"status": http.StatusBadRequest,
		})
		return
	}

	var errCode int
	_, errCode = M.CreateUser(&user)
	if errCode != M.DB_SUCCESS {
		c.AbortWithStatus(errCode)
	} else {
		c.JSON(http.StatusCreated, user)
	}
}

// Test command.
// curl -i -X GET http://localhost:8080/projects/1/users/bc7318e8-2b69-49b6-baf3-fdf47bcb1af9
func GetUserHandler(c *gin.Context) {
	projectId, err := strconv.ParseUint(c.Params.ByName("project_id"), 10, 64)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	id := c.Params.ByName("user_id")
	if id == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	user, errCode := M.GetUser(projectId, id)
	if errCode != M.DB_SUCCESS {
		c.AbortWithStatus(errCode)
	} else {
		c.JSON(http.StatusOK, user)
	}
}
