package tests

import (
	DD "factors/default_data"
	"factors/model/store"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrebuiltCustomKPI(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)
	assert.NotNil(t, agent)
	projectID := project.ID

	for _, integration := range DD.DefaultDataIntegrations {
		factory := DD.GetDefaultDataCustomKPIFactory(integration)
		statusCode2 := factory.Build(projectID)
		assert.Equal(t, http.StatusOK, statusCode2)
	}

	customMetrics, _, _ := store.GetStore().GetCustomMetricsByProjectId(projectID)
	assert.Equal(t, len(customMetrics), 24)

	for _, integration := range DD.DefaultDataIntegrations {
		factory := DD.GetDefaultDataCustomKPIFactory(integration)
		statusCode2 := factory.Build(projectID)
		assert.Equal(t, http.StatusOK, statusCode2)
	}

	customMetrics, _, _ = store.GetStore().GetCustomMetricsByProjectId(projectID)
	assert.Equal(t, len(customMetrics), 24)
}
