package tests

import (
	"factors/model/store"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMonitoringHandler(t *testing.T) {

	t.Run("TestGetProjectIdFromInfo", func(t *testing.T) {
		query1 := "SELECT * FROM WHERE project_id=100"
		query2 := "SELECT name,type FROM event_names WHERE project_id=10001 and name IN (?)"
		projectId1 := store.GetStore().GetProjectIdFromInfo(query1)
		projectId2 := store.GetStore().GetProjectIdFromInfo(query2)
		assert.Equal(t, projectId1, 100)
		assert.Equal(t, projectId2, 10001)
	})

}
