package tests

import (
	"encoding/json"
	H "factors/handler"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
)

type CustomEvent struct {
	EventName      string `json:"event_name"`
	CustomerUserID string `json:"customer_user_id"`
	Timestamp      int64  `json:"timestamp"`
	UTMCampaign    string `json:"utm_campaign"`
}

type ResultRow struct {
	AttributionKey string  `json:"attribution_key"`
	Cost           int64   `json:"cost"`
	Clicks         int64   `json:"clicks"`
	Impressions    int64   `json:"impressions"`
	Visitor        int64   `json:"visitor"`
	Conversion     float64 `json:"conversion"`
}

type CustomUsers struct {
	CustomerUserID string `json:"customer_user_id"`
	JoinTimestamp  int64  `json:"join_timestamp"`
}

func TestAttributionModelFile(t *testing.T) {

	// timestamp := int64(1614537000) 1st March 2021
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitAppRoutes(r)

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	customerAccountId := U.RandomLowerAphaNumString(5)
	// Should return error for non adwords customer account id
	result, err := store.GetStore().ExecuteAttributionQuery(project.ID, &model.AttributionQuery{})
	assert.NotNil(t, err)

	_, errCode := store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		IntAdwordsCustomerAccountId: &customerAccountId,
	})

	// Read adwords.
	adwordsDocs := getAdwordDocs()
	for _, adDoc := range adwordsDocs {
		// Over-writing the projectID
		adDoc.ProjectID = project.ID
		status := store.GetStore().CreateAdwordsDocument(&adDoc)
		assert.Equal(t, http.StatusCreated, status)
	}

	// Read users.
	users := getUsers()
	customerUserIdToUser := make(map[string]model.User)
	for _, user := range users {
		userTemp, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID,
			CustomerUserId: user.CustomerUserID, Properties: postgres.Jsonb{},
			JoinTimestamp: user.JoinTimestamp})
		customerUserIdToUser[user.CustomerUserID] = *userTemp
		assert.NotNil(t, userTemp)
		assert.Equal(t, http.StatusCreated, errCode)
	}

	// Read events.
	events := getEvents()
	for _, sessionEvent := range events {
		errCode = createEventWithSession(project.ID, sessionEvent.EventName,
			customerUserIdToUser[sessionEvent.CustomerUserID].ID, sessionEvent.Timestamp,
			customerUserIdToUser[sessionEvent.CustomerUserID].PropertiesId,
			sessionEvent.UTMCampaign)
		assert.Equal(t, http.StatusCreated, errCode)
	}

	// Read query.
	query := getQuery()

	// Read result.
	resultRows := getResult()
	attributionKeyToRow := make(map[string]ResultRow)
	for _, row := range resultRows {
		attributionKeyToRow[row.AttributionKey] = row
	}

	//Update user1 and user2 properties with latest campaign
	t.Run("AttributionQueryFileBased", func(t *testing.T) {

		result, err = store.GetStore().ExecuteAttributionQuery(project.ID, &query)
		assert.Nil(t, err)
		for _, row := range resultRows {
			assert.Equal(t, row.Conversion, getConversionUserCount(result, row.AttributionKey))
		}
	})
}

func getAdwordDocs() []model.AdwordsDocument {
	fileName := "adword_docs.json"
	byteVal := readBytesFromFile(fileName)
	var adwordDocs []model.AdwordsDocument
	err := json.Unmarshal(byteVal, &adwordDocs)
	if err != nil {
		fmt.Println(err)
		fmt.Println("Couldn't Unmarshal " + fileName + ". Exiting.")
		return adwordDocs
	}
	return adwordDocs
}

func getQuery() model.AttributionQuery {
	fileName := "query.json"
	byteVal := readBytesFromFile(fileName)
	var query model.AttributionQuery
	err := json.Unmarshal(byteVal, &query)
	if err != nil {
		fmt.Println(err)
		fmt.Println("Couldn't Unmarshal " + fileName + ". Exiting.")
		return query
	}
	return query
}

func getUsers() []CustomUsers {
	// Read User.
	fileName := "users.json"
	byteVal := readBytesFromFile(fileName)
	var users []CustomUsers
	err := json.Unmarshal(byteVal, &users)
	if err != nil {
		fmt.Println(err)
		fmt.Println("Couldn't Unmarshal " + fileName + ". Exiting.")
		return nil
	}
	return users
}

func getResult() []ResultRow {
	// Read Result.
	fileName := "result.json"
	byteVal := readBytesFromFile(fileName)
	var resultRows []ResultRow
	err := json.Unmarshal(byteVal, &resultRows)
	if err != nil {
		fmt.Println(err)
		fmt.Println("Couldn't Unmarshal " + fileName + ". Exiting.")
		return nil
	}
	return resultRows
}

func getEvents() []CustomEvent {
	// Read Events.
	fileName := "events.json"
	byteVal := readBytesFromFile(fileName)
	var events []CustomEvent
	err := json.Unmarshal(byteVal, &events)
	if err != nil {
		fmt.Println(err)
		fmt.Println("Couldn't Unmarshal " + fileName + ". Exiting.")
		return nil
	}
	return events
}

func readBytesFromFile(fileName string) []byte {
	jsonFile, err := os.Open("data/attrb/" + fileName)
	if err != nil {
		fmt.Println(err)
		fmt.Println("Couldn't open " + fileName + ". Exiting.")
		return nil
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	return byteValue
}
