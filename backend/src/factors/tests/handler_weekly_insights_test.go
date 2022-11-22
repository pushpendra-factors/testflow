package tests

import (
	"encoding/json"
	C "factors/config"
	"factors/delta"
	H "factors/handler"
	v1 "factors/handler/v1"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type WeeklyInsightsParams struct {
	ProjectID       int64     `json:"project_id"`
	DashBoardUnitID int64     `json:"dashboard_unit_id"`
	BaseStartTime   time.Time `json:"base_start_time"`
	CompStartTime   time.Time `json:"comp_start_time"`
	InsightsType    string    `json:"insights_type"`
	NumberOfRecords int       `json:"number_of_records"`
}

const TOKEN_GEN_RETRY_LIMIT = 5

func TestWeeklyInsights(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)
	var data v1.WeeklyInsightsParams
	projectObj, status := createProject(data.ProjectID)
	assert.Equal(t, http.StatusCreated, status)
	data.ProjectID = projectObj.ID
	data.QueryId = 112

	queryObj, status := createQuery(data.QueryId, data.ProjectID)
	assert.Equal(t, http.StatusCreated, status)
	EventNameObj, status := createEventName(projectObj.ID, "session")
	assert.Equal(t, http.StatusCreated, status)
	status = deleteEventName(&EventNameObj)
	assert.Equal(t, http.StatusOK, status)
	data.BaseStartTime = time.Unix(1625029668, 0)
	data.CompStartTime = time.Now()
	data.InsightsType = "w"
	data.NumberOfRecords = 2
	err := createPathAndAddCpiFile(data.ProjectID, data.BaseStartTime, data.QueryId)
	assert.Nil(t, err)
	response, err := delta.GetWeeklyInsights(data.ProjectID, "", data.QueryId, &data.BaseStartTime, &data.CompStartTime, data.InsightsType, data.NumberOfRecords, 0, 0, false)
	assert.NotNil(t, response)

	outputObj := response.(delta.WeeklyInsights)

	assert.Equal(t, "ConvAndDist", outputObj.InsightsType)
	assert.Equal(t, "", outputObj.BaseLine)
	assert.Equal(t, float64(3934), outputObj.Base.W1)
	assert.Equal(t, float64(4229), outputObj.Base.W2)
	assert.True(t, outputObj.Base.IsIncreased)
	assert.Equal(t, 7.498729028978139, outputObj.Base.Percentage)

	assert.Equal(t, float64(949), outputObj.Goal.W1)
	assert.Equal(t, float64(1007), outputObj.Goal.W2)
	assert.True(t, outputObj.Goal.IsIncreased)
	assert.Equal(t, 6.111696522655427, outputObj.Goal.Percentage)

	assert.Equal(t, 24.123029994916116, outputObj.Conv.W1)
	assert.Equal(t, 23.811775833530387, outputObj.Conv.W2)
	assert.False(t, outputObj.Conv.IsIncreased)
	assert.Equal(t, 1.2902780515189254, outputObj.Conv.Percentage)

	assert.Equal(t, outputObj.Insights[0].Key, "Browser")
	assert.Equal(t, outputObj.Insights[0].Value, "Internet Explorer")
	assert.Equal(t, outputObj.Insights[0].Entity, "ep")

	assert.Equal(t, float64(236), outputObj.Insights[0].ActualValues.W1)
	assert.Equal(t, float64(257), outputObj.Insights[0].ActualValues.W2)
	assert.Equal(t, 8.898305084745763, outputObj.Insights[0].ActualValues.Percentage)
	assert.True(t, outputObj.Insights[0].ActualValues.IsIncreased)

	assert.Equal(t, 24.205128205128204, outputObj.Insights[0].ChangeInConversion.W1)
	assert.Equal(t, 24.45290199809705, outputObj.Insights[0].ChangeInConversion.W2)
	assert.Equal(t, 1.0236417294263789, outputObj.Insights[0].ChangeInConversion.Percentage)
	assert.True(t, outputObj.Insights[0].ChangeInConversion.IsIncreased)

	assert.Equal(t, float64(975), outputObj.Insights[0].ChangeInPrevalance.W1)
	assert.Equal(t, float64(1051), outputObj.Insights[0].ChangeInPrevalance.W2)
	assert.Equal(t, 7.794871794871796, outputObj.Insights[0].ChangeInPrevalance.Percentage)
	assert.True(t, outputObj.Insights[0].ChangeInPrevalance.IsIncreased)

	assert.Equal(t, "conversion", outputObj.Insights[0].Type)

	assert.Equal(t, outputObj.Insights[1].Key, "Browser")
	assert.Equal(t, outputObj.Insights[1].Value, "Chrome")
	assert.Equal(t, outputObj.Insights[1].Entity, "ep")

	assert.Equal(t, float64(471), outputObj.Insights[1].ActualValues.W1)
	assert.Equal(t, float64(455), outputObj.Insights[1].ActualValues.W2)
	assert.Equal(t, 3.397027600849257, outputObj.Insights[1].ActualValues.Percentage)
	assert.False(t, outputObj.Insights[1].ActualValues.IsIncreased)

	assert.Equal(t, 24.00611620795107, outputObj.Insights[1].ChangeInConversion.W1)
	assert.Equal(t, 23.57512953367876, outputObj.Insights[1].ChangeInConversion.W2)
	assert.Equal(t, 1.7953202864591833, outputObj.Insights[1].ChangeInConversion.Percentage)
	assert.False(t, outputObj.Insights[1].ChangeInConversion.IsIncreased)

	assert.Equal(t, float64(1962), outputObj.Insights[1].ChangeInPrevalance.W1)
	assert.Equal(t, float64(1930), outputObj.Insights[1].ChangeInPrevalance.W2)
	assert.Equal(t, 1.6309887869520898, outputObj.Insights[1].ChangeInPrevalance.Percentage)
	assert.False(t, outputObj.Insights[1].ChangeInPrevalance.IsIncreased)

	assert.Equal(t, "conversion", outputObj.Insights[1].Type)

	assert.Equal(t, outputObj.Insights[2].Key, "Browser")
	assert.Equal(t, outputObj.Insights[2].Value, "Internet Explorer")
	assert.Equal(t, outputObj.Insights[2].Entity, "ep")

	assert.Equal(t, float64(236), outputObj.Insights[2].ActualValues.W1)
	assert.Equal(t, float64(257), outputObj.Insights[2].ActualValues.W2)
	assert.Equal(t, 8.898305084745763, outputObj.Insights[2].ActualValues.Percentage)
	assert.True(t, outputObj.Insights[2].ActualValues.IsIncreased)

	assert.Equal(t, 24.868282402528976, outputObj.Insights[2].ChangeInDistribution.W1)
	assert.Equal(t, 25.521350546176762, outputObj.Insights[2].ChangeInDistribution.W2)
	assert.Equal(t, 0.6530681436477863, outputObj.Insights[2].ChangeInDistribution.Percentage)
	assert.True(t, outputObj.Insights[2].ChangeInDistribution.IsIncreased)

	assert.Equal(t, "distribution", outputObj.Insights[2].Type)

	assert.Equal(t, outputObj.Insights[3].Key, "Browser")
	assert.Equal(t, outputObj.Insights[3].Value, "Chrome")
	assert.Equal(t, outputObj.Insights[3].Entity, "ep")

	assert.Equal(t, float64(471), outputObj.Insights[3].ActualValues.W1)
	assert.Equal(t, float64(455), outputObj.Insights[3].ActualValues.W2)
	assert.Equal(t, 3.397027600849257, outputObj.Insights[3].ActualValues.Percentage)
	assert.False(t, outputObj.Insights[3].ActualValues.IsIncreased)

	assert.Equal(t, 49.63119072708114, outputObj.Insights[3].ChangeInDistribution.W1)
	assert.Equal(t, 45.1837140019861, outputObj.Insights[3].ChangeInDistribution.W2)
	assert.Equal(t, 4.4474767250950435, outputObj.Insights[3].ChangeInDistribution.Percentage)
	assert.False(t, outputObj.Insights[3].ChangeInDistribution.IsIncreased)

	assert.Equal(t, "distribution", outputObj.Insights[3].Type)

	propertyObj := CreateWIPropertyObj(data.QueryId, "Browser", "Chrome", "distribution", "ep", 4, false)
	feedbackObj, errMsg := PostFeedback(projectObj.ID, 2, propertyObj)
	assert.Equal(t, "success", errMsg)
	response, err = delta.GetWeeklyInsights(data.ProjectID, feedbackObj.CreatedBy, data.QueryId, &data.BaseStartTime, &data.CompStartTime, data.InsightsType, data.NumberOfRecords, 0, 0, false) // testing again after blacklisting
	assert.NotNil(t, response)

	outputObj = response.(delta.WeeklyInsights)
	assert.Equal(t, "ConvAndDist", outputObj.InsightsType)
	assert.Equal(t, "", outputObj.BaseLine)
	assert.Equal(t, float64(3934), outputObj.Base.W1)
	assert.Equal(t, float64(4229), outputObj.Base.W2)
	assert.True(t, outputObj.Base.IsIncreased)
	assert.Equal(t, 7.498729028978139, outputObj.Base.Percentage)

	assert.Equal(t, float64(949), outputObj.Goal.W1)
	assert.Equal(t, float64(1007), outputObj.Goal.W2)
	assert.True(t, outputObj.Goal.IsIncreased)
	assert.Equal(t, 6.111696522655427, outputObj.Goal.Percentage)

	assert.Equal(t, 24.123029994916116, outputObj.Conv.W1)
	assert.Equal(t, 23.811775833530387, outputObj.Conv.W2)
	assert.False(t, outputObj.Conv.IsIncreased)
	assert.Equal(t, 1.2902780515189254, outputObj.Conv.Percentage)

	assert.Equal(t, 0, len(outputObj.Insights)) // means element is blacklisted

	err = deletePath("projects")
	assert.Nil(t, err)
	deleteQuery(&queryObj)
	deleteFeedback(&feedbackObj)
	deleteProject(&projectObj)
}

func createProject(projectID int64) (projectObj model.Project, status int) {
	var project model.Project
	project.Name = "test"
	token, err := generateUniqueToken(false)
	if err != nil {
		log.WithError(err).Error("Create project failed.")
		return project, http.StatusInternalServerError
	}
	project.Token = token
	privateToken, err := generateUniqueToken(true)
	if err != nil {
		log.WithError(err).Error("Create project failed.")
		return project, http.StatusInternalServerError
	}
	project.PrivateToken = privateToken
	db := C.GetServices()
	if err := db.Db.Create(&project).Error; err != nil {
		log.WithError(err).Error("Create project failed.")
		return project, http.StatusInternalServerError
	}
	return project, http.StatusCreated
}
func deleteProject(projectObj *model.Project) (status int) {
	db := C.GetServices()
	err := db.Db.Delete(&projectObj)
	if err != nil {
		return http.StatusNotFound
	}
	return http.StatusOK
}
func deleteQuery(queryObj *model.Queries) (status int) {
	db := C.GetServices()
	err := db.Db.Delete(&queryObj)
	if err != nil {
		return http.StatusNotFound
	}
	return http.StatusOK
}
func createQuery(QueryId int64, ProjectID int64) (queryObj model.Queries, status int) {
	var query model.Queries
	query.ID = QueryId
	query.ProjectID = ProjectID
	query.Title = "testQuery"
	query.Type = 2
	query.IsDeleted = false
	var ewp []model.QueryEventWithProperties
	temp := model.QueryEventWithProperties{
		Name: "$session",
	}
	ewp = append(ewp, temp)
	queryNested := model.Query{
		Class:                model.QueryClassFunnel,
		EventsCondition:      model.EventCondAnyGivenEvent,
		From:                 1556602834,
		To:                   1557207634,
		Type:                 model.QueryTypeEventsOccurrence,
		EventsWithProperties: ewp, // invalid, no events.
		OverridePeriod:       true,
	}
	queryJson, _ := json.Marshal(queryNested)
	query1 := &postgres.Jsonb{queryJson}
	query.Query = *query1
	db := C.GetServices()
	if err := db.Db.Create(&query).Error; err != nil {
		log.WithError(err).Error("Create query failed.")
		return query, http.StatusInternalServerError
	}
	return query, http.StatusCreated
}

func createEventName(projectID int64, name string) (EventName model.EventName, status int) {
	eventName, errCode := store.GetStore().CreateOrGetUserCreatedEventName(&model.EventName{ProjectId: projectID, Name: name})
	if errCode != http.StatusCreated {
		return *eventName, http.StatusNotFound
	}
	return *eventName, http.StatusCreated
}
func deleteEventName(eventNameObj *model.EventName) (status int) {
	db := C.GetServices()
	db.Db.Delete(&eventNameObj)
	return http.StatusOK
}
func deleteFeedback(feedbackObj *model.Feedback) (staus int) {
	db := C.GetServices()
	db.Db.Delete(&feedbackObj)
	return http.StatusOK
}
func generateUniqueToken(private bool) (token string, err error) {
	for tryCount := 0; tryCount < TOKEN_GEN_RETRY_LIMIT; tryCount++ {
		token = U.RandomLowerAphaNumString(32)
		tokenExists, err := isTokenExist(token, private)
		if err != nil {
			return "", err
		}
		// Break generation, if token doesn't exist already.
		if tokenExists == 0 {
			return token, nil
		}
	}
	return "", fmt.Errorf("token generation failed after %d attempts", TOKEN_GEN_RETRY_LIMIT)
}

// Checks for the existence of token already.
func isTokenExist(token string, private bool) (exists int, err error) {
	db := C.GetServices().Db

	whereCondition := "token = ?"
	if private {
		whereCondition = "private_token = ?"
	}
	var count uint64
	if err := db.Model(&model.Project{}).Where(whereCondition, token).Count(&count).Error; err != nil {
		return -1, err
	}

	if count > 0 {
		return 1, nil
	}

	return 0, nil
}
func createPathAndAddCpiFile(projectID int64, baseStartTime time.Time, QueryId int64) (err error) {
	path := fmt.Sprintf("projects/" + strconv.FormatUint(uint64(projectID), 10) + "/weeklyinsights/" + U.GetDateOnlyFromTimestampZ(baseStartTime.Unix()) + "/q-" + strconv.FormatInt(QueryId, 10) + "/k-100")
	os.MkdirAll(path, 0777)

	sourceFile, err := os.Open("data/cpi.txt")
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	newFile, err := os.Create(path + "/cpi.txt")
	if err != nil {
		return err
	}
	_, err = io.Copy(newFile, sourceFile)
	if err != nil {
		return err
	}
	return nil
}
func deletePath(path string) error {
	err := os.RemoveAll("projects")
	return err
}
func CreateWIPropertyObj(queryID int64, key, value, Type, entity string, order int, IsIncreased bool) model.WeeklyInsightsProperty {
	var WIPropertyObj model.WeeklyInsightsProperty
	WIPropertyObj.Key = key
	WIPropertyObj.Value = value
	WIPropertyObj.Type = Type
	WIPropertyObj.Entity = entity
	WIPropertyObj.Order = order
	WIPropertyObj.QueryID = queryID
	WIPropertyObj.IsIncreased = IsIncreased
	return WIPropertyObj
}
func PostFeedback(ProjectID int64, voteType int, WIPropertyObj model.WeeklyInsightsProperty) (model.Feedback, string) {
	var feedbackObj model.Feedback
	email := getRandomEmail()
	agent, errCode := SetupAgentReturnDAO(email, "+13425356")
	if errCode != http.StatusCreated {
		return feedbackObj, "failed to setup agent"
	}
	propertyBytes, err := json.Marshal(WIPropertyObj)
	if err != nil {
		return feedbackObj, error.Error(err)
	}
	property := postgres.Jsonb{RawMessage: json.RawMessage(propertyBytes)}

	feedbackObj.ProjectID = ProjectID
	feedbackObj.Property = &property
	feedbackObj.CreatedBy = agent.UUID
	feedbackObj.VoteType = voteType
	feedbackObj.Feature = "weekly_insights"

	status, _ := store.GetStore().PostFeedback(feedbackObj.ProjectID, feedbackObj.CreatedBy, feedbackObj.Feature, &property, feedbackObj.VoteType)
	if status != http.StatusCreated {
		return feedbackObj, "failed to post feedback"
	}
	return feedbackObj, "success"
}
