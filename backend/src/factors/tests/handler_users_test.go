package tests

import (
	"encoding/json"
	H "factors/handler"
	M "factors/model"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
)

func TestAPIGetUsers(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitAppRoutes(r)
	projectId, err := SetupProject()
	assert.Nil(t, err)

	// Create 100 Users.
	users := make([]M.User, 0, 0)
	numUsers := 100
	for i := 0; i < numUsers; i++ {
		user, _ := M.CreateUser(&M.User{ProjectId: projectId})
		users = append(users, *user)
	}

	// Default values of offset and limit. Not sent in params.
	offset := 0
	limit := 10
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET",
		fmt.Sprintf("/projects/%d/users", projectId), nil)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	retUsers := make([]M.User, 0, 0)
	json.Unmarshal(jsonResponse, &retUsers)
	assert.Equal(t, limit, len(retUsers))
	assertUserMapsWithOffset(t, users[offset:offset+limit], retUsers)

	offset = 25
	limit = 20
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET",
		fmt.Sprintf("/projects/%d/users?offset=%d&limit=%d", projectId, offset, limit), nil)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &retUsers)
	assert.Equal(t, limit, len(retUsers))
	assertUserMapsWithOffset(t, users[offset:offset+limit], retUsers)

	// Overflow
	offset = 95
	limit = 10
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET",
		fmt.Sprintf("/projects/%d/users?offset=%d&limit=%d", projectId, offset, limit), nil)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &retUsers)
	assert.Equal(t, numUsers-95, len(retUsers))
	assertUserMapsWithOffset(t, users[offset:numUsers], retUsers)
}

func assertUserMapsWithOffset(t *testing.T, expectedUsers []M.User, actualUsers []M.User) {
	assert.Equal(t, len(expectedUsers), len(actualUsers))
	for i := 0; i < len(actualUsers); i++ {
		expectedUser := expectedUsers[i]
		actualUser := actualUsers[i]
		assert.Equal(t, expectedUser.ID, actualUser.ID)
		assert.Equal(t, expectedUser.ProjectId, actualUser.ProjectId)
		assert.Equal(t, expectedUser.CustomerUserId, actualUser.CustomerUserId)
		assert.Equal(t, postgres.Jsonb{RawMessage: json.RawMessage([]byte(`null`))}, actualUser.Properties)
		assert.NotNil(t, actualUser.CreatedAt)
		assert.NotNil(t, actualUser.UpdatedAt)
	}
}
