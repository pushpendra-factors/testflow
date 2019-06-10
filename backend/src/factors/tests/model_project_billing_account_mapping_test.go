package tests

import (
	M "factors/model"
	U "factors/util"
	"net/http"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetProjectBillingAccountMappings(t *testing.T) {
	testData, errCode := SetupTestData()
	assert.Equal(t, http.StatusCreated, errCode)

	billingAcc := testData.BillingAccount
	project := testData.Project

	projectBillingAccMappings, errCode := M.GetProjectBillingAccountMappings(billingAcc.ID)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, 1, len(projectBillingAccMappings))
	assert.Equal(t, project.ID, projectBillingAccMappings[0].ProjectID)

}

func TestGetProjectBillingAccountMapping(t *testing.T) {
	testData, errCode := SetupTestData()
	assert.Equal(t, http.StatusCreated, errCode)

	billingAcc := testData.BillingAccount
	project := testData.Project

	projectBillingAccMapping, errCode := M.GetProjectBillingAccountMapping(project.ID)
	assert.Equal(t, http.StatusFound, errCode)

	assert.Equal(t, billingAcc.ID, projectBillingAccMapping.BillingAccountID)
}

func TestCreateMultipleProjectsUnderSameBillingAccount(t *testing.T) {
	testData, errCode := SetupTestData()
	assert.Equal(t, http.StatusCreated, errCode)

	agent := testData.Agent
	billingAcc := testData.BillingAccount
	randProjectName := U.RandomLowerAphaNumString(15)
	newProject, errCode := M.CreateProjectWithDependencies(&M.Project{Name: randProjectName}, agent.UUID, M.ADMIN, billingAcc.ID)
	assert.Equal(t, http.StatusCreated, errCode)
	projectBillingAccMappings, errCode := M.GetProjectBillingAccountMappings(billingAcc.ID)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, 2, len(projectBillingAccMappings))

	expProjectIDs := []uint64{testData.Project.ID, newProject.ID}
	sort.Slice(expProjectIDs, func(i, j int) bool {
		return expProjectIDs[i] < expProjectIDs[j]
	})

	resultProjectIDs := []uint64{projectBillingAccMappings[0].ProjectID, projectBillingAccMappings[1].ProjectID}
	sort.Slice(resultProjectIDs, func(i, j int) bool {
		return resultProjectIDs[i] < resultProjectIDs[j]
	})

	assert.Equal(t, expProjectIDs, resultProjectIDs)
}
