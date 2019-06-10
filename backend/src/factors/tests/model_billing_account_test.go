package tests

import (
	M "factors/model"
	U "factors/util"
	"net/http"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateGetBillingAccountByAgentUUID(t *testing.T) {
	t.Run("CreateDefaultPlan", func(t *testing.T) {
		// CreateAgent
		// Creating agent should create default billing account with freePlan
		agent, errCode := SetupAgentReturnDAO(getRandomEmail())
		assert.Equal(t, http.StatusCreated, errCode)

		ba, errCode := M.GetBillingAccountByAgentUUID(agent.UUID)
		assert.Equal(t, http.StatusFound, errCode)
		assert.Equal(t, M.FreePlanID, ba.PlanID)
	})
	t.Run("SpecificPlan", func(t *testing.T) {
		cAP := &M.CreateAgentParams{Agent: &M.Agent{Email: getRandomEmail()}, PlanCode: M.StartupPlanCode}
		resp, errCode := M.CreateAgentWithDependencies(cAP)
		assert.Equal(t, http.StatusCreated, errCode)
		ba, errCode := M.GetBillingAccountByAgentUUID(resp.Agent.UUID)
		assert.Equal(t, http.StatusFound, errCode)
		assert.Equal(t, M.StartupPlanID, ba.PlanID)
	})
}

func TestUpdateBillingAccount(t *testing.T) {

	agent, errCode := SetupAgentReturnDAO(getRandomEmail())
	assert.Equal(t, http.StatusCreated, errCode)

	ba, errCode := M.GetBillingAccountByAgentUUID(agent.UUID)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, "", ba.OrganizationName)
	assert.Equal(t, "", ba.BillingAddress)
	assert.Equal(t, "", ba.Pincode)
	assert.Equal(t, "", ba.PhoneNo)
	assert.Equal(t, M.FreePlanID, ba.PlanID)

	orgName := U.RandomString(8)
	PhoneNo := "123452"
	billingAddress := U.RandomString(20)
	pincode := "640034"
	errCode = M.UpdateBillingAccount(ba.ID, M.StartupPlanID, orgName, billingAddress, pincode, PhoneNo)
	assert.Equal(t, http.StatusAccepted, errCode)

	updatedBa, errCode := M.GetBillingAccountByAgentUUID(agent.UUID)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, orgName, updatedBa.OrganizationName)
	assert.Equal(t, billingAddress, updatedBa.BillingAddress)
	assert.Equal(t, pincode, updatedBa.Pincode)
	assert.Equal(t, PhoneNo, updatedBa.PhoneNo)
	assert.Equal(t, M.StartupPlanID, updatedBa.PlanID)
}

func TestGetProjectsUnderBillingAccount(t *testing.T) {
	td, errCode := SetupTestData()
	assert.Equal(t, http.StatusCreated, errCode)

	project, errCode := M.CreateProjectWithDependencies(&M.Project{Name: U.RandomString(6)}, td.Agent.UUID, M.ADMIN, td.BillingAccount.ID)
	assert.Equal(t, http.StatusCreated, errCode)

	expProjectIDs := []uint64{td.Project.ID, project.ID}
	sort.Slice(expProjectIDs, func(i, j int) bool { return expProjectIDs[i] < expProjectIDs[j] })

	resProjects, errCode := M.GetProjectsUnderBillingAccountID(td.BillingAccount.ID)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, 2, len(resProjects))

	resProjectIDs := []uint64{resProjects[0].ID, resProjects[1].ID}
	sort.Slice(resProjectIDs, func(i, j int) bool { return resProjectIDs[i] < resProjectIDs[j] })
	assert.Equal(t, expProjectIDs, resProjectIDs)

}

func TestGetAgentsByProjectIDs(t *testing.T) {
	td, errCode := SetupTestData()
	assert.Equal(t, http.StatusCreated, errCode)

	agent2, errCode := SetupAgentReturnDAO(getRandomEmail())
	assert.Equal(t, http.StatusCreated, errCode)

	project2, errCode := M.CreateProjectWithDependencies(&M.Project{Name: U.RandomString(6)}, agent2.UUID, M.ADMIN, td.BillingAccount.ID)
	assert.Equal(t, http.StatusCreated, errCode)

	expAgentsUUID := []string{td.Agent.UUID, agent2.UUID}

	agents, errCode := M.GetAgentsByProjectIDs([]uint64{td.Project.ID, project2.ID})
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, 2, len(agents))

	resultAgentsUUID := []string{agents[0].UUID, agents[1].UUID}

	sort.Strings(expAgentsUUID)
	sort.Strings(resultAgentsUUID)
	assert.Equal(t, expAgentsUUID, resultAgentsUUID)
}

func TestIsNewAgentCreationAllowed(t *testing.T) {
	td, errCode := SetupTestData()
	assert.Equal(t, http.StatusCreated, errCode)

	project := td.Project

	freePlan, _ := M.GetPlanByID(td.BillingAccount.PlanID)

	noOfAgentsToCreate := freePlan.MaxNoOfAgents - 1

	createdAgents := make([]*M.Agent, 0, 0)
	for i := 0; i < noOfAgentsToCreate; i++ {
		ag, errCode := SetupAgentReturnDAO(getRandomEmail())
		assert.Equal(t, http.StatusCreated, errCode)
		createdAgents = append(createdAgents, ag)
	}
	// create project agent mapping
	for _, createdAgent := range createdAgents {

		allowed, errCode := M.IsNewProjectAgentMappingCreationAllowed(td.Project.ID, createdAgent.Email)
		assert.Equal(t, http.StatusOK, errCode)

		assert.True(t, allowed)

		_, errCode = M.CreateProjectAgentMappingWithDependencies(&M.ProjectAgentMapping{
			ProjectID: project.ID,
			AgentUUID: createdAgent.UUID,
		})
		assert.Equal(t, http.StatusCreated, errCode)
	}

	// new agent creation will fail
	allowed, errCode := M.IsNewProjectAgentMappingCreationAllowed(td.Project.ID, getRandomEmail())
	assert.Equal(t, http.StatusOK, errCode)

	assert.False(t, allowed)
}

func TestGetBillingAccountByProjectID(t *testing.T) {
	td, errCode := SetupTestData()
	assert.Equal(t, http.StatusCreated, errCode)

	project := td.Project

	resultBA, errCode := M.GetBillingAccountByProjectID(project.ID)
	assert.Equal(t, http.StatusFound, errCode)

	assert.Equal(t, td.BillingAccount.ID, resultBA.ID)
}
