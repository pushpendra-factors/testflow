package tests

import (
	"encoding/json"
	H "factors/handler"
	SP "factors/task/smart_properties"
	U "factors/util"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
)

func TestNewSmartPropertyCreation(t *testing.T) {
	r1 := gin.Default()
	r2 := gin.Default()
	H.InitAppRoutes(r1)

	H.InitDataServiceRoutes(r2)

	project, customerAccountID, agent, statusCode := createProjectAndAddAdwordsDocument(t, r1)
	assert.Equal(t, http.StatusAccepted, statusCode)

	campaignID := "1"
	value := map[string]interface{}{"id": "1", "campaign_id": campaignID, "name": "Campaign_1"}
	valueJSON, _ := U.EncodeToPostgresJsonb(&value)
	w := sendCreateAdwordsDocumentReq(r2, project.ID, customerAccountID, "campaigns", 20220101, campaignID, valueJSON)
	assert.Equal(t, http.StatusCreated, w.Code)

	name1 := U.RandomString(8)
	description := U.RandomString(20)
	rules := &postgres.Jsonb{json.RawMessage(`[{"value": "Campaign_added", "source": "adwords", "filters": [{"name": "campaign","property": "name","condition": "contains","logical_operator": "AND","value": "Campaign"}]}]`)}
	w = sendCreateSmartPropertyReq(r1, project.ID, agent, rules, "campaign", name1, description)
	assert.Equal(t, http.StatusOK, w.Code)
	SP.EnrichSmartPropertyForChangedRulesForProject(project.ID)
	// SP.EnrichSmartPropertyForCurrentDayForProject(project.ID)

}
