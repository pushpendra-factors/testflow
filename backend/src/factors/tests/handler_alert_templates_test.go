package tests

import (
	H "factors/handler"
	U "factors/util"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)


func TestGetAlertTemplate(t *testing.T){
	r := gin.Default()
	H.InitAppRoutes(r)

	rb := U.NewRequestBuilder(http.MethodGet, "http://factors-dev.com:8080/common/alert_templates")
	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Failed to Fetch Templates")
	}
	
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	fmt.Print(r.AppEngine)
	assert.Equal(t, http.StatusOK, w.Code)
}
func TestDeleteAlertTemplate(t *testing.T){
	r := gin.Default()
	H.InitAppRoutes(r)
	rb := U.NewRequestBuilder(http.MethodDelete, "http://factors-dev.com:8080/common/alert_templates/8")

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Failed to Delete Template")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	fmt.Print(r.AppEngine)
	assert.Equal(t, http.StatusOK, w.Code)
	



	rb = U.NewRequestBuilder(http.MethodDelete, "http://factors-dev.com:8080/common/alert_templates/account_execs")

	req, err = rb.Build()
	if err != nil {
		log.WithError(err).Error("Failed to Delete Template")
	}

	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	fmt.Print(r.AppEngine)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
}