package tests

import (
	H "factors/handler"
	M "factors/model/model"
	"factors/model/store"
	U "factors/util"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func testingCreateAdwordsDocument(t *testing.T) {
	r := gin.Default()
	H.InitDataServiceRoutes(r)

	project, customerAccountID, _, statusCode := createProjectAndAddAdwordsDocument(t, r)
	if statusCode == http.StatusAccepted {
		campaignID := "1"
		value := map[string]interface{}{"cost": "100", "clicks": "50", "campaign_id": campaignID, "impressions": "1000", "campaign_name": "Campaign_1"}
		valueJSON, err := U.EncodeToPostgresJsonb(&value)
		assert.Nil(t, err)

		w := sendCreateAdwordsDocumentReq(r, project.ID, customerAccountID, "campaign_performance_report", 20210205, campaignID, valueJSON)
		assert.Equal(t, http.StatusCreated, w.Code)
	} else {
		assert.Equal(t, true, false)
	}
}

func createProjectAndAddAdwordsDocument(t *testing.T, r *gin.Engine) (*M.Project, string, *M.Agent, int) {

	project, agent, _ := SetupProjectWithAgentDAO()
	assert.NotNil(t, project)
	customerAccountID1 := U.RandomString(10)
	customerAccountID2 := U.RandomString(10)
	customerAccounts := customerAccountID1 + "," + customerAccountID2
	_, statusCode := store.GetStore().UpdateProjectSettings(project.ID, &M.ProjectSetting{
		IntAdwordsCustomerAccountId: &customerAccounts,
		IntAdwordsEnabledAgentUUID:  &agent.UUID,
	})
	assert.Equal(t, http.StatusAccepted, statusCode)
	if statusCode != http.StatusAccepted {
		return nil, "", nil, statusCode
	}
	return project, customerAccountID1, agent, statusCode
}

func sendCreateAdwordsDocumentReq(r *gin.Engine, projectID uint64, customerAccountID string, typeAlias string, timestamp int64, id string, valueJSON *postgres.Jsonb) *httptest.ResponseRecorder {
	payload := map[string]interface{}{
		"project_id":      projectID,
		"customer_acc_id": customerAccountID,
		"type_alias":      typeAlias,
		"timestamp":       timestamp,
		"id":              id,
		"value":           valueJSON,
	}

	rb := U.NewRequestBuilder(http.MethodPost, "http://localhost:8089/data_service/adwords/documents/add").
		WithPostParams(payload)

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error sending facebook document add requests to data server.")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
