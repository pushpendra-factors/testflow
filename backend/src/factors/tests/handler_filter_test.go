package tests

import (
	"encoding/json"
	H "factors/handler"
	M "factors/model"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAPICreateFilterHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)

	filterURL := fmt.Sprintf("/projects/%d/filters", project.ID)

	// Test get project settings.
	name := "u1_u2"
	expr := "a.com/u1/u2"
	reqPayload := fmt.Sprintf(`{"name": "%s", "expr": "%s"}`, name, expr)
	w := ServePostRequest(r, filterURL, []byte(reqPayload))
	assert.Equal(t, http.StatusCreated, w.Code)
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	var jsonResponseMap map[string]interface{}
	json.Unmarshal(jsonResponse, &jsonResponseMap)
	assert.NotZero(t, jsonResponseMap["project_id"])
	assert.NotZero(t, jsonResponseMap["id"])
	assert.NotNil(t, jsonResponseMap["name"])
	assert.Equal(t, "u1_u2", jsonResponseMap["name"])
	assert.NotNil(t, jsonResponseMap["expr"])
	assert.Equal(t, expr, jsonResponseMap["expr"])

	name = "u1_v1"
	expr = "a.com/u1/:v1"
	reqPayload = fmt.Sprintf(`{"name": "%s", "expr": "%s"}`, name, expr)
	w1 := ServePostRequest(r, filterURL, []byte(reqPayload))
	assert.Equal(t, http.StatusCreated, w1.Code)
	jsonResponse1, _ := ioutil.ReadAll(w1.Body)
	var jsonResponseMap1 map[string]interface{}
	json.Unmarshal(jsonResponse1, &jsonResponseMap1)
	assert.NotZero(t, jsonResponseMap1["project_id"])
	assert.NotZero(t, jsonResponseMap1["id"])
	assert.NotNil(t, jsonResponseMap["name"])
	assert.Equal(t, "u1_v1", jsonResponseMap1["name"])
	assert.NotNil(t, jsonResponseMap1["expr"])
	assert.Equal(t, expr, jsonResponseMap1["expr"])

	// invalid name
	name = "u1_v2"
	expr = "a.com/u1/:v2"
	reqPayload = fmt.Sprintf(`{"name": "", "expr": "%s"}`, expr)
	w1 = ServePostRequest(r, filterURL, []byte(reqPayload))
	assert.Equal(t, http.StatusBadRequest, w1.Code)

	// Bad request.
	name = ""
	expr = ""
	reqPayload = fmt.Sprintf(`{"name": "%s", "expr": "%s"}`, name, expr)
	w2 := ServePostRequest(r, filterURL, []byte(reqPayload))
	assert.Equal(t, http.StatusBadRequest, w2.Code)
	jsonResponse2, _ := ioutil.ReadAll(w2.Body)
	var jsonResponseMap2 map[string]interface{}
	json.Unmarshal(jsonResponse2, &jsonResponseMap2)
	assert.NotNil(t, jsonResponseMap2["error"])

	name = "$dollarName"
	expr = "a.com/dollar"
	reqPayload = fmt.Sprintf(`{"name": "%s", "expr": "%s"}`, name, expr)
	w3 := ServePostRequest(r, filterURL, []byte(reqPayload))
	assert.Equal(t, http.StatusBadRequest, w3.Code)
	jsonResponse3, _ := ioutil.ReadAll(w3.Body)
	var jsonResponseMap3 map[string]interface{}
	json.Unmarshal(jsonResponse3, &jsonResponseMap3)
	assert.NotNil(t, jsonResponseMap3["error"])

	name = "u1_u2"
	expr = ""
	reqPayload = fmt.Sprintf(`{"name": "%s", "expr": "%s"}`, name, expr)
	w4 := ServePostRequest(r, filterURL, []byte(reqPayload))
	assert.Equal(t, http.StatusBadRequest, w4.Code)
	jsonResponse4, _ := ioutil.ReadAll(w4.Body)
	var jsonResponseMap4 map[string]interface{}
	json.Unmarshal(jsonResponse4, &jsonResponseMap4)
	assert.NotNil(t, jsonResponseMap4["error"])

	name = "u1_u2"
	// user copied the url and pasted as expression.
	expr = "https://a.com/u1/u5?q=search_string"
	reqPayload = fmt.Sprintf(`{"name": "%s", "expr": "%s"}`, name, expr)
	w5 := ServePostRequest(r, filterURL, []byte(reqPayload))
	assert.Equal(t, http.StatusCreated, w5.Code)
	jsonResponse5, _ := ioutil.ReadAll(w5.Body)
	var jsonResponseMap5 map[string]interface{}
	json.Unmarshal(jsonResponse5, &jsonResponseMap5)
	assert.NotZero(t, jsonResponseMap5["project_id"])
	assert.NotZero(t, jsonResponseMap5["id"])
	assert.NotNil(t, jsonResponseMap5["name"])
	assert.Equal(t, "u1_u2", jsonResponseMap5["name"])
	assert.NotNil(t, jsonResponseMap5["expr"])
	// sanitized expr from user given url.
	assert.Equal(t, "a.com/u1/u5", jsonResponseMap5["expr"])
}

func TestAPIGetFiltersHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)

	filterURL := fmt.Sprintf("/projects/%d/filters", project.ID)

	w := ServeGetRequest(r, filterURL)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	retFilters := make([]H.API_FilterResponePayload, 0, 0)
	json.Unmarshal(jsonResponse, &retFilters)
	assert.Equal(t, 0, len(retFilters))

	// Filters created.
	filters := map[string]string{
		"u1_u2": "a.com/u1/u2",
		"u1_v1": "a.com/u1/:v1",
	}

	for k, v := range filters {
		M.CreateOrGetFilterEventName(&M.EventName{ProjectId: project.ID, Name: k, FilterExpr: v})
	}

	w1 := ServeGetRequest(r, filterURL)
	assert.Equal(t, http.StatusOK, w1.Code)
	jsonResponse1, _ := ioutil.ReadAll(w1.Body)
	retFilters1 := make([]H.API_FilterResponePayload, 0, 0)
	json.Unmarshal(jsonResponse1, &retFilters1)
	assert.Equal(t, 2, len(retFilters1))
}

func TestAPIUpdateFilterHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)

	filter, _ := M.CreateOrGetFilterEventName(&M.EventName{ProjectId: project.ID, Name: "u1_u2", FilterExpr: "a.com/u1/:u2"})
	assert.NotNil(t, filter)

	filterURL := fmt.Sprintf("/projects/%d/filters/%d", project.ID, filter.ID)

	// Test get project settings.
	name := "new_name"
	reqPayload := fmt.Sprintf(`{"name": "%s"}`, name)
	w := ServePutRequest(r, filterURL, []byte(reqPayload))
	assert.Equal(t, http.StatusAccepted, w.Code)
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	var jsonResponseMap map[string]interface{}
	json.Unmarshal(jsonResponse, &jsonResponseMap)
	assert.NotNil(t, jsonResponseMap["name"])
	assert.Equal(t, name, jsonResponseMap["name"])
	assert.Nil(t, jsonResponseMap["expr"]) // omit empty.

	// Empty name update.
	reqPayload = fmt.Sprintf(`{"name": ""}`)
	w = ServePutRequest(r, filterURL, []byte(reqPayload))
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Try updating expr.
	expr := "a.com/u1/u3"
	reqPayload = fmt.Sprintf(`{"expr": "%s"}`, expr)
	w = ServePutRequest(r, filterURL, []byte(reqPayload))
	assert.Equal(t, http.StatusBadRequest, w.Code)
	jsonResponse1, _ := ioutil.ReadAll(w.Body)
	var jsonResponseMap1 map[string]interface{}
	json.Unmarshal(jsonResponse1, &jsonResponseMap1)
	assert.NotNil(t, jsonResponseMap1["error"])

	name = "$new_name"
	reqPayload = fmt.Sprintf(`{"name": "%s"}`, name)
	w2 := ServePutRequest(r, filterURL, []byte(reqPayload))
	assert.Equal(t, http.StatusBadRequest, w2.Code)
	jsonResponse2, _ := ioutil.ReadAll(w2.Body)
	var jsonResponseMap2 map[string]interface{}
	json.Unmarshal(jsonResponse2, &jsonResponseMap2)
	assert.NotNil(t, jsonResponseMap2["error"])
}

func TestAPIDeleteFilterHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)

	// Invalid filter.
	invalidFilterURL := fmt.Sprintf("/projects/%d/filters/%d", project.ID, 99999)
	w := ServeDeleteRequest(r, invalidFilterURL)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	var jsonResponseMap map[string]interface{}
	json.Unmarshal(jsonResponse, &jsonResponseMap)
	assert.NotNil(t, jsonResponseMap["error"])

	// Valid filter.
	filter, _ := M.CreateOrGetFilterEventName(&M.EventName{ProjectId: project.ID, Name: "u1_u2", FilterExpr: "a.com/u1/:u2"})
	assert.NotNil(t, filter)

	filterURL := fmt.Sprintf("/projects/%d/filters/%d", project.ID, filter.ID)
	w1 := ServeDeleteRequest(r, filterURL)
	assert.Equal(t, http.StatusAccepted, w1.Code)
	jsonResponse1, _ := ioutil.ReadAll(w1.Body)
	var jsonResponseMap1 map[string]interface{}
	json.Unmarshal(jsonResponse1, &jsonResponseMap1)
	assert.NotNil(t, jsonResponseMap1["project_id"])
	assert.NotNil(t, jsonResponseMap1["id"])
}
