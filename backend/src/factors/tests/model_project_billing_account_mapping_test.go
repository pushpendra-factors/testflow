package tests

import (
	"factors/model/model"
	"factors/model/store"
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

	projectBillingAccMappings, errCode := store.GetStore().GetProjectBillingAccountMappings(billingAcc.ID)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, 1, len(projectBillingAccMappings))
	assert.Equal(t, project.ID, projectBillingAccMappings[0].ProjectID)

}

func TestGetProjectBillingAccountMapping(t *testing.T) {
	testData, errCode := SetupTestData()
	assert.Equal(t, http.StatusCreated, errCode)

	billingAcc := testData.BillingAccount
	project := testData.Project

	projectBillingAccMapping, errCode := store.GetStore().GetProjectBillingAccountMapping(project.ID)
	assert.Equal(t, http.StatusFound, errCode)

	assert.Equal(t, billingAcc.ID, projectBillingAccMapping.BillingAccountID)
}

func TestCreateMultipleProjectsUnderSameBillingAccount(t *testing.T) {
	testData, errCode := SetupTestData()
	assert.Equal(t, http.StatusCreated, errCode)

	agent := testData.Agent
	billingAcc := testData.BillingAccount
	randProjectName := U.RandomLowerAphaNumString(15)
	newProject, errCode := store.GetStore().CreateProjectWithDependencies(&model.Project{Name: randProjectName}, agent.UUID, model.ADMIN, billingAcc.ID, true)
	assert.Equal(t, http.StatusCreated, errCode)
	projectBillingAccMappings, errCode := store.GetStore().GetProjectBillingAccountMappings(billingAcc.ID)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, 2, len(projectBillingAccMappings))

	expProjectIDs := []int64{testData.Project.ID, newProject.ID}
	sort.Slice(expProjectIDs, func(i, j int) bool {
		return expProjectIDs[i] < expProjectIDs[j]
	})

	resultProjectIDs := []int64{projectBillingAccMappings[0].ProjectID, projectBillingAccMappings[1].ProjectID}
	sort.Slice(resultProjectIDs, func(i, j int) bool {
		return resultProjectIDs[i] < resultProjectIDs[j]
	})

	assert.Equal(t, expProjectIDs, resultProjectIDs)
}

func TestPBAMConstraints(t *testing.T) {
	agent, errCode := SetupAgentReturnDAO(getRandomEmail(), "+91234567890")
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, agent)

	agentBillingAccount, errCode := store.GetStore().GetBillingAccountByAgentUUID(agent.UUID)
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotEmpty(t, agentBillingAccount)

	// Creating with a valid billing account id works.
	_, errCode = store.GetStore().CreateProjectWithDependencies(
		&model.Project{Name: U.RandomString(5)}, agent.UUID, model.ADMIN, agentBillingAccount.ID, true)
	assert.Equal(t, http.StatusCreated, errCode)

	// Creating new project with random billingAccountID should fail.
	billingAccountID := U.GetUUID()
	_, errCode = store.GetStore().CreateProjectWithDependencies(
		&model.Project{Name: U.RandomString(5)}, agent.UUID, model.ADMIN, billingAccountID, true)
	assert.Equal(t, http.StatusInternalServerError, errCode)
}
