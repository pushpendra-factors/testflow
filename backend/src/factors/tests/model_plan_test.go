package tests

import (
	"factors/model/model"
	"factors/model/store"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPlanByCode(t *testing.T) {
	plan, errCode := store.GetStore().GetPlanByCode(model.FreePlanCode)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, model.FreePlanID, plan.ID)
}

func TestGetPlanByID(t *testing.T) {
	plan, errCode := store.GetStore().GetPlanByID(model.StartupPlanID)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, model.StartupPlanCode, plan.Code)
}
