package tests

import (
	"encoding/json"
	C "factors/config"
	H "factors/handler"
	"factors/handler/helpers"
	"factors/model/model"
	"factors/model/store"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func sendGetUserReq(r *gin.Engine, projectId uint64, agent *model.Agent, offset, limit *int) *httptest.ResponseRecorder {

	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}

	qP := make(map[string]string)
	if offset != nil {
		qP["offset"] = fmt.Sprintf("%d", *offset)
	}
	if limit != nil {
		qP["limit"] = fmt.Sprintf("%d", *limit)
	}

	rb := C.NewRequestBuilderWithPrefix(http.MethodGet, fmt.Sprintf("/projects/%d/users", projectId)).
		WithHeader("Content-UnitType", "application/json").
		WithQueryParams(qP).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})
	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error Creating get users Req")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
func TestAPIGetUsers(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitAppRoutes(r)
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	projectId := project.ID

	// Create 100 Users.
	users := make([]model.User, 0, 0)
	numUsers := 100
	for i := 0; i < numUsers; i++ {
		createdUserID, _ := store.GetStore().CreateUser(&model.User{ProjectId: projectId, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
		user, errCode := store.GetStore().GetUser(projectId, createdUserID)
		assert.Equal(t, http.StatusFound, errCode)
		users = append(users, *user)
	}

	// Default values of offset and limit. Not sent in params.
	offset := 0
	limit := 10
	w := sendGetUserReq(r, projectId, agent, nil, nil)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	retUsers := make([]model.User, 0, 0)
	json.Unmarshal(jsonResponse, &retUsers)
	assert.Equal(t, limit, len(retUsers))
	assertUserMapsWithOffset(t, users[offset:offset+limit], retUsers)

	offset = 25
	limit = 20
	w = sendGetUserReq(r, projectId, agent, &offset, &limit)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &retUsers)
	assert.Equal(t, limit, len(retUsers))
	assertUserMapsWithOffset(t, users[offset:offset+limit], retUsers)

	// Overflow
	offset = 95
	limit = 10
	w = sendGetUserReq(r, projectId, agent, &offset, &limit)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &retUsers)
	assert.Equal(t, numUsers-95, len(retUsers))
	assertUserMapsWithOffset(t, users[offset:numUsers], retUsers)
}

func assertUserMapsWithOffset(t *testing.T, expectedUsers []model.User, actualUsers []model.User) {
	assert.Equal(t, len(expectedUsers), len(actualUsers))
	for i := 0; i < len(actualUsers); i++ {
		expectedUser := expectedUsers[i]
		actualUser := actualUsers[i]
		assert.Equal(t, expectedUser.ID, actualUser.ID)
		assert.Equal(t, expectedUser.ProjectId, actualUser.ProjectId)
		assert.Equal(t, expectedUser.CustomerUserId, actualUser.CustomerUserId)
		// Atleast join_time should be present on user_properites.
		assert.NotEqual(t, postgres.Jsonb{RawMessage: json.RawMessage([]byte(`null`))}, actualUser.Properties)
		assert.NotNil(t, actualUser.CreatedAt)
		assert.NotNil(t, actualUser.UpdatedAt)
	}
}
