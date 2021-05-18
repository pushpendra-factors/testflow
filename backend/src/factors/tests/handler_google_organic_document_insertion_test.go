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

func sendCreateGoogleOrganicDocumentReq(r *gin.Engine, googleOrganicDocument model.GoogleOrganicDocument) *httptest.ResponseRecorder {
	payload := map[string]interface{}{
		"project_id": googleOrganicDocument.ProjectID,
		"url_prefix": googleOrganicDocument.URLPrefix,
		"id":         googleOrganicDocument.ID,
		"value":      googleOrganicDocument.Value,
		"timestamp":  googleOrganicDocument.Timestamp,
	}

	rb := U.NewRequestBuilder(http.MethodPost, "http://localhost:8089/data_service/google_organic/documents/add").
		WithPostParams(payload)

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error sending googleOrganic document add requests to data server.")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestDataSeviceGoogleOrganicDocumentAdd(t *testing.T) {
	r := gin.Default()
	H.InitDataServiceRoutes(r)

	//inserting sample data in googleOrganic, also testing data service endpoint googleOrganic/documents/add
	project, _ := SetupProjectReturnDAO()
	assert.NotNil(t, project)
	urlPrefix := U.RandomNumericString(10)
	_, errCode := store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		IntGoogleOrganicURLPrefixes: &urlPrefix,
	})
	assert.Equal(t, http.StatusAccepted, errCode)

	id1 := U.RandomNumericString(8)
	value1, _ := json.Marshal(map[string]interface{}{"query": "factors ai", "clicks": "50", "id": id1, "impressions": "1000"})
	id2 := U.RandomNumericString(8)
	value2, _ := json.Marshal(map[string]interface{}{"query": "factors.ai", "clicks": "100", "id": id2, "impressions": "2000"})

	googleOrganicDocuments := []model.GoogleOrganicDocument{
		{ID: id1, ProjectID: project.ID, URLPrefix: urlPrefix, Timestamp: 20210205,
			Value: &postgres.Jsonb{value1}},

		{ID: id2, ProjectID: project.ID, URLPrefix: urlPrefix, Timestamp: 20210206,
			Value: &postgres.Jsonb{value2}},
	}

	for _, googleOrganicDocument := range googleOrganicDocuments {
		log.Warn("tx", googleOrganicDocument)
		w := sendCreateGoogleOrganicDocumentReq(r, googleOrganicDocument)
		assert.Equal(t, http.StatusCreated, w.Code)
	}
}
