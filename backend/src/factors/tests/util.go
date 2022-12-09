package tests

import (
	"bytes"
	"encoding/json"
	"factors/model/model"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm/dialects/postgres"

	U "factors/util"
)

// ServePostRequest Creates a post request and returns a ResponseRecorder object,
// which can be used to test for required results.
func ServePostRequest(r *gin.Engine, uri string,
	reqBodyString []byte) *httptest.ResponseRecorder {

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", uri, bytes.NewBuffer(reqBodyString))
	req.Header.Set("Content-UnitType", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func ServePutRequest(r *gin.Engine, uri string,
	reqBodyString []byte) *httptest.ResponseRecorder {

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", uri, bytes.NewBuffer(reqBodyString))
	req.Header.Set("Content-UnitType", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func ServeGetRequest(r *gin.Engine, uri string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", uri, bytes.NewBuffer([]byte{}))
	req.Header.Set("Content-UnitType", "application/json") // Default header.
	r.ServeHTTP(w, req)
	return w
}

func ServeDeleteRequest(r *gin.Engine, uri string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", uri, bytes.NewBuffer([]byte{}))
	req.Header.Set("Content-UnitType", "application/json") // Default header.
	r.ServeHTTP(w, req)
	return w
}

func setHeaders(req *http.Request, headers map[string]string) {
	for k, v := range headers {
		req.Header.Set(k, v)
	}
}

func ServePostRequestWithHeaders(r *gin.Engine, uri string, reqBodyString []byte,
	headers map[string]string) *httptest.ResponseRecorder {

	if len(headers) == 0 {
		log.Fatal("Please use ServePostRequest, if you don't have any custom headers to be set.")
	}
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", uri, bytes.NewBuffer(reqBodyString))
	req.Header.Set("Content-UnitType", "application/json") // Default header.
	setHeaders(req, headers)                               // Setting custom headers.
	req.RemoteAddr = "127.0.0.1"
	r.ServeHTTP(w, req)
	return w
}

func ServeGetRequestWithHeaders(r *gin.Engine, uri string, headers map[string]string) *httptest.ResponseRecorder {
	if len(headers) == 0 {
		log.Fatal("Please use ServePostRequest, if you don't have any custom headers to be set.")
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", uri, bytes.NewBuffer([]byte{}))
	req.Header.Set("Content-UnitType", "application/json") // Default header.
	setHeaders(req, headers)                               // Setting custom headers.
	r.ServeHTTP(w, req)
	return w
}

func DecodeJSONResponseToMap(body *bytes.Buffer) map[string]interface{} {
	var responseMap map[string]interface{}
	jsonResponse, err := ioutil.ReadAll(body)
	if err != nil {
		log.Fatalf("JSON decode failed : %+v", responseMap)
	}
	json.Unmarshal(jsonResponse, &responseMap)
	return responseMap
}

func DecodeJSONResponseToAnalyticsResult(body *bytes.Buffer) *model.QueryResult {
	var responseMap map[string]interface{}
	jsonResponse, err := ioutil.ReadAll(body)
	if err != nil {
		log.Fatalf("JSON decode failed : %+v", responseMap)
	}
	var result model.QueryResult
	json.Unmarshal(jsonResponse, &result)
	return &result
}

func RandomURL() string {
	return fmt.Sprintf("http://%s.com/%s",
		U.RandomLowerAphaNumString(5), U.RandomLowerAphaNumString(5))
}

func DecodePostgresJsonbWithoutError(jsonb *postgres.Jsonb) *map[string]interface{} {
	properties, _ := U.DecodePostgresJsonb(jsonb)
	return properties
}
