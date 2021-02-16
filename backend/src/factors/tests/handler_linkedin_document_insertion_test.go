package tests

import (
	"encoding/json"
	H "factors/handler"
	"factors/model/model"
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

func sendCreateLinkedinDocumentReq(r *gin.Engine, linkedinDocument model.LinkedinDocument) *httptest.ResponseRecorder {
	payload := map[string]interface{}{
		"project_id":             linkedinDocument.ProjectID,
		"customer_ad_account_id": linkedinDocument.CustomerAdAccountID,
		"type_alias":             linkedinDocument.TypeAlias,
		"id":                     linkedinDocument.ID,
		"value":                  linkedinDocument.Value,
		"timestamp":              linkedinDocument.Timestamp,
	}

	rb := U.NewRequestBuilder(http.MethodPost, "http://localhost:8089/data_service/linkedin/documents/add").
		WithPostParams(payload)

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error sending linkedin document add requests to data server.")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestDataSeviceLinkedinDocumentAdd(t *testing.T) {
	r := gin.Default()
	H.InitDataServiceRoutes(r)

	//inserting sample data in linkedin, also testing data service endpoint linkedin/documents/add
	project, _ := SetupProjectReturnDAO()
	assert.NotNil(t, project)
	customerAccountID := U.RandomNumericString(10)
	_, errCode := store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		IntLinkedinAdAccount: customerAccountID,
	})
	assert.Equal(t, http.StatusAccepted, errCode)

	campaignID1 := U.RandomNumericString(8)
	campaign1Value, _ := json.Marshal(map[string]interface{}{"costInLocalCurrency": "100", "clicks": "50", "campaign_group_id": campaignID1, "impressions": "1000", "campaign_group_name": "campaign_group_1"})
	campaignID2 := U.RandomNumericString(8)
	campaign2Value, _ := json.Marshal(map[string]interface{}{"costInLocalCurrency": "200", "clicks": "100", "campaign_group_id": campaignID2, "impressions": "2000", "campaign_group_name": "campaign_group_2"})
	adgroupID1_1 := U.RandomNumericString(8)
	adgroupID1_1Value, _ := json.Marshal(map[string]interface{}{"costInLocalCurrency": "30", "clicks": "30", "campaign_id": adgroupID1_1, "campaign_name": "Adgroup_1_1", "campaign_group_id": campaignID1, "impressions": "600", "campaign_group_name": "campaign_group_1"})
	adgroupID1_2 := U.RandomNumericString(8)
	adgroup1_2Value, _ := json.Marshal(map[string]interface{}{"costInLocalCurrency": "70", "clicks": "20", "campaign_id": adgroupID1_2, "campaign_name": "Adgroup_1_2", "campaign_group_id": campaignID1, "impressions": "400", "campaign_group_name": "campaign_group_1"})
	adgroupID2_1 := U.RandomNumericString(8)
	adgroup2_1Value, _ := json.Marshal(map[string]interface{}{"costInLocalCurrency": "120", "clicks": "25", "campaign_id": adgroupID2_1, "campaign_name": "Adgroup_2_1", "campaign_group_id": campaignID2, "impressions": "1500", "campaign_group_name": "campaign_group_2"})
	adgroupID2_2 := U.RandomNumericString(8)
	adgroup2_2Value, _ := json.Marshal(map[string]interface{}{"costInLocalCurrency": "80", "clicks": "75", "campaign_id": adgroupID2_2, "campaign_name": "Adgroup_2_2", "campaign_group_id": campaignID2, "impressions": "500", "campaign_group_name": "campaign_group_2"})
	linkedinDocuments := []model.LinkedinDocument{
		{ID: campaignID1, ProjectID: project.ID, CustomerAdAccountID: customerAccountID, TypeAlias: "campaign_group_insights", Timestamp: 20210205,
			Value: &postgres.Jsonb{campaign1Value}},

		{ID: campaignID2, ProjectID: project.ID, CustomerAdAccountID: customerAccountID, TypeAlias: "campaign_group_insights", Timestamp: 20210206,
			Value: &postgres.Jsonb{campaign2Value}},

		{ID: adgroupID1_1, ProjectID: project.ID, CustomerAdAccountID: customerAccountID, TypeAlias: "campaign_insights", Timestamp: 20210205,
			Value: &postgres.Jsonb{adgroupID1_1Value}},

		{ID: adgroupID1_2, ProjectID: project.ID, CustomerAdAccountID: customerAccountID, TypeAlias: "campaign_insights", Timestamp: 20210205,
			Value: &postgres.Jsonb{adgroup1_2Value}},

		{ID: adgroupID2_1, ProjectID: project.ID, CustomerAdAccountID: customerAccountID, TypeAlias: "campaign_insights", Timestamp: 20210206,
			Value: &postgres.Jsonb{adgroup2_1Value}},

		{ID: adgroupID2_2, ProjectID: project.ID, CustomerAdAccountID: customerAccountID, TypeAlias: "campaign_insights", Timestamp: 20210206,
			Value: &postgres.Jsonb{adgroup2_2Value}},
	}

	for _, linkedinDocument := range linkedinDocuments {
		log.Warn("tx", linkedinDocument)
		w := sendCreateLinkedinDocumentReq(r, linkedinDocument)
		assert.Equal(t, http.StatusCreated, w.Code)
	}
}
