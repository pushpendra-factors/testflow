package tests

import (
	M "factors/model"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPlanByCode(t *testing.T) {
	plan, errCode := M.GetPlanByCode(M.FreePlanCode)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, M.FreePlanID, plan.ID)
}

func TestGetPlanByID(t *testing.T) {
	plan, errCode := M.GetPlanByID(M.StartupPlanID)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, M.StartupPlanCode, plan.Code)
}
