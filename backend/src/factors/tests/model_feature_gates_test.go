package tests

import (
	"factors/model/model"
	"factors/model/store"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFeatureStatusWithProjectSettings(t *testing.T) {
	project, _, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	// exluding project settings 
	status,err := store.GetStore().GetFeatureStatusForProjectV2(project.ID,model.FEATURE_GOOGLE_ADS,false)
	assert.Nil(t,err)
	assert.Equal(t,true,status)

	// including project settings
	status,err = store.GetStore().GetFeatureStatusForProjectV2(project.ID,model.FEATURE_GOOGLE_ADS,true)
	assert.Nil(t,err)
	assert.Equal(t,false,status)

}
