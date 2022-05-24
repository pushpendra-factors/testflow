package tests

import (
	"encoding/json"
	"factors/model/model"
	"factors/model/store"
	"net/http"
	"testing"

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestGetAlertNamesByProjectIdTypeAndName(t *testing.T) {
	project, _, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	alert := model.Alert{
		ProjectID: project.ID, AlertName: "testing", AlertType: 1,
		AlertDescription:   &postgres.Jsonb{json.RawMessage(`{"name":"engaged_sessions_per_user","query":{"ca":"events","dc":"website_session","me":["engaged_sessions_per_user"],"pgURL":"abc.com","tz":"Asia/Kolkata"},"query_type":"kpi","operator":"is_greater_than","value":"122","date_range":"last_week"}`)},
		AlertConfiguration: &postgres.Jsonb{json.RawMessage(`{}`)}, IsDeleted: false,
	}
	_, a, b := store.GetStore().CreateAlert(project.ID, alert)

	alertNames, statusCode := store.GetStore().GetAlertNamesByProjectIdTypeAndName(project.ID, "engaged_sessions_per_user")
	log.WithField("alertNames", alertNames).Warn("testing")
	assert.Equal(t, statusCode, http.StatusFound)
	assert.Len(t, alertNames, 1)
}
