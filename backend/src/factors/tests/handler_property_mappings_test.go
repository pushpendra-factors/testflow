package tests

import (
	"encoding/json"
	C "factors/config"
	H "factors/handler"
	"factors/handler/helpers"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"strings"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestPropertyMappingHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	displayName1 := "Test property mapping 1"
	invalidDisplayName := ""
	project, agent, _ := SetupProjectWithAgentDAO()
	assert.NotNil(t, project)

	propertiesRaw1 := model.Property{
		Category:        "events",
		DisplayCategory: "page_views",
		ObjectType:      "",
		Name:            "$browser1",
		DataType:        "categorical",
		Entity:          "event",
		GroupByType:     "",
	}
	propertiesRaw2 := model.Property{
		Category:        "channels",
		DisplayCategory: "google_ads_metrics",
		ObjectType:      "",
		Name:            "$browser2",
		DataType:        "categorical",
		Entity:          "event",
		GroupByType:     "",
	}
	propertiesRaw3 := model.Property{
		Category:        "events",
		DisplayCategory: "website_session",
		ObjectType:      "",
		Name:            "$browser3",
		DataType:        "categorical",
		Entity:          "event",
		GroupByType:     "",
	}
	propertiesRaw4 := model.Property{
		Category:        "events",
		DisplayCategory: "website_session",
		ObjectType:      "",
		Name:            "$browser4",
		DataType:        "numerical",
		Entity:          "event",
		GroupByType:     "",
	}
	propertiesRaw5 := model.Property{
		Category:        "profiles",
		DisplayCategory: "hubspot_contacts",
		ObjectType:      "",
		Name:            "$browser4",
		DataType:        "categorical",
		Entity:          "user",
		GroupByType:     "",
	}
	invalidProperty1 := model.Property{
		Category:        "events",
		DisplayCategory: "",
		ObjectType:      "",
		Name:            "$browser",
		DataType:        "categorical",
		Entity:          "event",
		GroupByType:     "",
	}

	var firstPropertyMapping *model.PropertyMapping
	t.Run("CreatePropertyMappingSuccess", func(t *testing.T) {
		propertiesRaw := []model.Property{propertiesRaw1, propertiesRaw2, propertiesRaw3}
		properties_byte, _ := json.Marshal(propertiesRaw)
		properties := &postgres.Jsonb{RawMessage: properties_byte}
		w := sendCreatePropertyMapping(r, project.ID, agent, properties, displayName1)
		assert.Equal(t, http.StatusOK, w.Code)
		decoder := json.NewDecoder(w.Body)
		if err := decoder.Decode(&firstPropertyMapping); err != nil {
			assert.NotNil(t, nil, err)
		}
		assert.NotNil(t, firstPropertyMapping)
	})

	t.Run("CreatePropertyMappingFailureDuplicateName", func(t *testing.T) {
		propertiesRaw := []model.Property{propertiesRaw1, propertiesRaw2}
		properties_byte, _ := json.Marshal(propertiesRaw)
		properties := &postgres.Jsonb{RawMessage: properties_byte}
		w := sendCreatePropertyMapping(r, project.ID, agent, properties, displayName1)
		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Equal(t, strings.Contains(w.Body.String(), "Entity with the same name exists"), true)
	})

	t.Run("CreatePropertyMappingFailureMultipleDataType", func(t *testing.T) {
		propertiesRaw := []model.Property{propertiesRaw1, propertiesRaw2, propertiesRaw4}
		properties_byte, _ := json.Marshal(propertiesRaw)
		properties := &postgres.Jsonb{RawMessage: properties_byte}
		w := sendCreatePropertyMapping(r, project.ID, agent, properties, "temp")
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, strings.Contains(w.Body.String(), "All properties should have same data type"), true)
	})

	t.Run("CreatePropertyMappingFailureInvalidDisplayName", func(t *testing.T) {
		propertiesRaw := []model.Property{propertiesRaw1, propertiesRaw2}
		properties_byte, _ := json.Marshal(propertiesRaw)
		properties := &postgres.Jsonb{RawMessage: properties_byte}
		w := sendCreatePropertyMapping(r, project.ID, agent, properties, invalidDisplayName)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, strings.Contains(w.Body.String(), "Invalid display name"), true)
	})

	t.Run("CreatePropertyMappingFailureInvalidProperty", func(t *testing.T) {
		propertiesRaw := []model.Property{propertiesRaw2, invalidProperty1}
		properties_byte, _ := json.Marshal(propertiesRaw)
		properties := &postgres.Jsonb{RawMessage: properties_byte}
		w := sendCreatePropertyMapping(r, project.ID, agent, properties, "temp")
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, strings.Contains(w.Body.String(), "Error with values passed in properties"), true)
	})

	t.Run("CreatePropertyMappingFailureDuplicateDisplayCategoryProperty", func(t *testing.T) {
		propertiesRaw4.DataType = "categorical"
		propertiesRaw := []model.Property{propertiesRaw3, propertiesRaw4}
		properties_byte, _ := json.Marshal(propertiesRaw)
		properties := &postgres.Jsonb{RawMessage: properties_byte}
		w := sendCreatePropertyMapping(r, project.ID, agent, properties, "temp")
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, strings.Contains(w.Body.String(), "Duplicate display categor"), true)
	})

	t.Run("GetPropertyMappingsSuccess", func(t *testing.T) {
		w := sendGetPropertyMappings(r, project.ID, agent)
		assert.Equal(t, http.StatusOK, w.Code)
		var result []*model.PropertyMapping
		decoder := json.NewDecoder(w.Body)
		if err := decoder.Decode(&result); err != nil {
			assert.NotNil(t, nil, err)
		}
		assert.NotNil(t, result)
		assert.Equal(t, len(result), 1)
		assert.Equal(t, result[0].DisplayName, displayName1)
	})

	t.Run("GetCommonPropertyMappingsSuccess with direct display categores", func(t *testing.T) {
		displayName2 := "test2"
		propertiesRaw := []model.Property{propertiesRaw1, propertiesRaw2, propertiesRaw3}
		properties_byte, _ := json.Marshal(propertiesRaw)
		properties := &postgres.Jsonb{RawMessage: properties_byte}
		w := sendCreatePropertyMapping(r, project.ID, agent, properties, displayName2)
		assert.Equal(t, http.StatusOK, w.Code)

		displayName3 := "test3"
		propertiesRaw = []model.Property{propertiesRaw2, propertiesRaw3}
		properties_byte, _ = json.Marshal(propertiesRaw)
		properties = &postgres.Jsonb{RawMessage: properties_byte}
		w = sendCreatePropertyMapping(r, project.ID, agent, properties, displayName3)
		assert.Equal(t, http.StatusOK, w.Code)

		displayName4 := "test4"
		propertiesRaw = []model.Property{propertiesRaw2, propertiesRaw5}
		properties_byte, _ = json.Marshal(propertiesRaw)
		properties = &postgres.Jsonb{RawMessage: properties_byte}
		w = sendCreatePropertyMapping(r, project.ID, agent, properties, displayName4)
		assert.Equal(t, http.StatusOK, w.Code)

		payloadRaw := fmt.Sprintf(`[{"name": "%s", "derived_kpi": %s},{"name": "%s", "derived_kpi": %s}]`,
			model.WebsiteSessionDisplayCategory, "false", model.GoogleAdsDisplayCategory, "false")
		payload := &postgres.Jsonb{RawMessage: json.RawMessage(payloadRaw)}
		w = sendGetCommonPropertyMappings(r, project.ID, agent, payload)
		assert.Equal(t, http.StatusOK, w.Code)
		var result []*model.PropertyMapping
		decoder := json.NewDecoder(w.Body)
		if err := decoder.Decode(&result); err != nil {
			assert.NotNil(t, nil, err)
		}
		assert.NotNil(t, result)
		assert.Equal(t, len(result), 3)

		payloadRaw = fmt.Sprintf(`[{"name": "%s", "derived_kpi": %s},{"name": "%s", "derived_kpi": %s}]`,
								model.HubspotContactsDisplayCategory, "false", model.GoogleAdsDisplayCategory, "false")
		payload = &postgres.Jsonb{RawMessage: json.RawMessage(payloadRaw)}
		w = sendGetCommonPropertyMappings(r, project.ID, agent, payload)
		assert.Equal(t, http.StatusOK, w.Code)
		decoder = json.NewDecoder(w.Body)
		if err := decoder.Decode(&result); err != nil {
			assert.NotNil(t, nil, err)
		}
		assert.NotNil(t, result)
		assert.Equal(t, len(result), 1)
	})

	t.Run("GetCommonPropertyMappingsSuccess with derived metric", func(t *testing.T) {
		name1 := U.RandomString(8)
		description1 := U.RandomString(8)
		transformations1 := &postgres.Jsonb{
			RawMessage: json.RawMessage(`{"cl":"kpi","for":"a/b","qG":[{"ca":"channels","dc":"google_ads_metrics","fil":[],"gBy":[],"me":["impressions"],"na":"a","pgUrl":"","tz":"Australia/Sydney"},{"ca":"events","dc":"website_session","fil":[],"gBy":[],"me":["clicks"],"na":"b","pgUrl":"","tz":"Australia/Sydney"}]}`),
		}
		w := sendCreateCustomMetric(r, project.ID, agent, transformations1, name1, description1, "google_ads_metrics", 2)
		assert.Equal(t, http.StatusOK, w.Code)

		payloadRaw := fmt.Sprintf(`[{"name": "%s", "derived_kpi": %s},{"name": "%s", "derived_kpi": %s}]`,
									model.PageViewsDisplayCategory, "false", name1, "true")
		payload := &postgres.Jsonb{RawMessage: json.RawMessage(payloadRaw)}
		w = sendGetCommonPropertyMappings(r, project.ID, agent, payload)
		assert.Equal(t, http.StatusOK, w.Code)
		var result []*model.PropertyMapping
		decoder := json.NewDecoder(w.Body)
		if err := decoder.Decode(&result); err != nil {
			assert.NotNil(t, nil, err)
		}
		assert.NotNil(t, result)
		assert.Equal(t, len(result), 2)
	})

	t.Run("DeletePropertyMappingsSuccess", func(t *testing.T) {
		w := sendDeletePropertyMappings(r, project.ID, agent, firstPropertyMapping.ID)
		assert.Equal(t, http.StatusOK, w.Code)
		w = sendGetPropertyMappings(r, project.ID, agent)
		assert.Equal(t, http.StatusOK, w.Code)
		var result []*model.PropertyMapping
		decoder := json.NewDecoder(w.Body)
		if err := decoder.Decode(&result); err != nil {
			assert.NotNil(t, nil, err)
		}
		assert.NotNil(t, result)
		assert.Equal(t, len(result), 3)
	})
}

func sendCreatePropertyMapping(r *gin.Engine, project_id int64, agent *model.Agent, properties *postgres.Jsonb,
	displayName string) *httptest.ResponseRecorder {
	payload := map[string]interface{}{
		"project_id":   project_id,
		"properties":   properties,
		"display_name": displayName,
	}
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}
	url := "/projects/" + strconv.FormatUint(uint64(project_id), 10) + "/kpi/property_mappings"
	rb := C.NewRequestBuilderWithPrefix(http.MethodPost, url).
		WithPostParams(payload).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error sending create property mapping request.")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func sendGetPropertyMappings(r *gin.Engine, propject_id int64, agent *model.Agent) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}
	url := "/projects/" + strconv.FormatUint(uint64(propject_id), 10) + "/kpi/property_mappings"
	rb := C.NewRequestBuilderWithPrefix(http.MethodGet, url).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error sending get property mapping request.")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func sendGetCommonPropertyMappings(r *gin.Engine, project_id int64, agent *model.Agent, payload interface{}) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}
	url := "/projects/" + strconv.FormatUint(uint64(project_id), 10) + "/kpi/property_mappings/commom_properties"
	rb := C.NewRequestBuilderWithPrefix(http.MethodPost, url).
		WithPostParams(payload).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error sending get common property mapping request.")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func sendDeletePropertyMappings(r *gin.Engine, project_id int64, agent *model.Agent, id string) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}
	url := "/projects/" + strconv.FormatUint(uint64(project_id), 10) + "/kpi/property_mappings/" + id
	rb := C.NewRequestBuilderWithPrefix(http.MethodDelete, url).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error sending delete property mapping request.")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
