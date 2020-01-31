package tests

import (
	M "factors/model"
	U "factors/util"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateProjectAgentMapping(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	email := getRandomEmail()
	agent, errCode := SetupAgentReturnDAO(email, "+13425356")
	assert.Equal(t, http.StatusCreated, errCode)

	pam := &M.ProjectAgentMapping{
		ProjectID: project.ID,
		AgentUUID: agent.UUID,
		Role:      M.AGENT,
	}
	_, errCode = M.CreateProjectAgentMappingWithDependencies(pam)
	assert.Equal(t, http.StatusCreated, errCode)
}

func TestCreateDuplicateProjectAgentMapping(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	email := getRandomEmail()
	agent, errCode := SetupAgentReturnDAO(email, "+13425356")
	assert.Equal(t, http.StatusCreated, errCode)

	pam := &M.ProjectAgentMapping{
		ProjectID: project.ID,
		AgentUUID: agent.UUID,
		Role:      M.AGENT,
	}
	_, errCode = M.CreateProjectAgentMappingWithDependencies(pam)
	assert.Equal(t, http.StatusCreated, errCode)

	_, errCode = M.CreateProjectAgentMappingWithDependencies(pam)
	assert.Equal(t, http.StatusFound, errCode)

}

func TestGetProjectAgentMapping(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	email := getRandomEmail()
	agent, errCode := SetupAgentReturnDAO(email, "+13425356")
	assert.Equal(t, http.StatusCreated, errCode)

	pam := &M.ProjectAgentMapping{
		ProjectID: project.ID,
		AgentUUID: agent.UUID,
		Role:      M.AGENT,
	}
	_, errCode = M.CreateProjectAgentMappingWithDependencies(pam)
	assert.Equal(t, http.StatusCreated, errCode)

	t.Run("MappingMissing", func(t *testing.T) {
		randomProjectId := U.RandomUint64()%10007 + 5
		_, errCode = M.GetProjectAgentMapping(randomProjectId, agent.UUID)
		assert.Equal(t, http.StatusNotFound, errCode)
	})

	t.Run("MappingFound", func(t *testing.T) {
		retPam, errCode := M.GetProjectAgentMapping(project.ID, agent.UUID)
		assert.Equal(t, http.StatusFound, errCode)
		assert.Equal(t, pam.Role, retPam.Role)
	})

}

func TestDBGetProjectAgentMappingsByProjectId(t *testing.T) {
	t.Run("MissingParams", func(t *testing.T) {
		_, errCode := M.GetProjectAgentMappingsByProjectId(0)
		assert.Equal(t, http.StatusBadRequest, errCode)
	})
	t.Run("NotFound", func(t *testing.T) {
		randProjectId := U.RandomUint64WithUnixNano()
		_, errCode := M.GetProjectAgentMappingsByProjectId(randProjectId)
		assert.Equal(t, http.StatusNotFound, errCode)
	})

	t.Run("Found", func(t *testing.T) {
		project, err := SetupProjectReturnDAO()
		assert.Nil(t, err)
		noOfAgents := int(U.RandomUint64()%10 + 5)
		createdAgents := make([]*M.Agent, 0, 0)
		for i := 0; i < noOfAgents; i++ {
			ag, errCode := SetupAgentReturnDAO(getRandomEmail(), "+13425356")
			assert.Equal(t, http.StatusCreated, errCode)
			createdAgents = append(createdAgents, ag)
		}
		// create project agent mapping
		for _, createdAgent := range createdAgents {

			_, errCode := M.CreateProjectAgentMappingWithDependencies(&M.ProjectAgentMapping{
				ProjectID: project.ID,
				AgentUUID: createdAgent.UUID,
			})
			assert.Equal(t, http.StatusCreated, errCode)
		}

		// fetch all
		retPams, errCode := M.GetProjectAgentMappingsByProjectId(project.ID)
		assert.Equal(t, http.StatusFound, errCode)
		noOfDefautsAgentsPerProject := 1
		assert.Equal(t, len(createdAgents)+noOfDefautsAgentsPerProject, len(retPams))
	})
}

func TestDBGetProjectAgentMappingsByAgentUUID(t *testing.T) {
	t.Run("MissingParams", func(t *testing.T) {
		_, errCode := M.GetProjectAgentMappingsByAgentUUID("")
		assert.Equal(t, http.StatusBadRequest, errCode)
	})
	t.Run("NotFound", func(t *testing.T) {
		_, errCode := M.GetProjectAgentMappingsByAgentUUID(getRandomAgentUUID())
		assert.Equal(t, http.StatusNotFound, errCode)
	})

	t.Run("Found", func(t *testing.T) {
		agent, errCode := SetupAgentReturnDAO(getRandomEmail(), "+13425356")
		assert.Equal(t, http.StatusCreated, errCode)
		noOfProjects := int(U.RandomUint64()%10 + 5)
		var projects []*M.Project
		for i := 0; i < noOfProjects; i++ {
			project, err := SetupProjectReturnDAO()
			assert.Nil(t, err)
			projects = append(projects, project)
		}
		for _, project := range projects {
			_, errCode := M.CreateProjectAgentMappingWithDependencies(&M.ProjectAgentMapping{
				ProjectID: project.ID,
				AgentUUID: agent.UUID,
			})
			assert.Equal(t, http.StatusCreated, errCode)
		}
		retPams, errCode := M.GetProjectAgentMappingsByAgentUUID(agent.UUID)
		assert.Equal(t, http.StatusFound, errCode)
		assert.Equal(t, len(projects), len(retPams))
	})
}

func TestDeleteProjectAgentMapping(t *testing.T) {

	t.Run("NotFound", func(t *testing.T) {
		testData, errCode := SetupTestData()
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, testData)

		project := testData.Project

		agent, errCode := SetupAgentReturnDAO(getRandomEmail(), "+13425356")
		assert.Equal(t, http.StatusCreated, errCode)
		errCode = M.DeleteProjectAgentMapping(project.ID, agent.UUID)
		assert.Equal(t, http.StatusNotFound, errCode)
	})

	t.Run("Success", func(t *testing.T) {
		testData, errCode := SetupTestData()
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, testData)

		project := testData.Project
		agent, errCode := SetupAgentReturnDAO(getRandomEmail(), "+13425356")
		assert.Equal(t, http.StatusCreated, errCode)

		_, errCode = M.CreateProjectAgentMappingWithDependencies(&M.ProjectAgentMapping{
			ProjectID: project.ID,
			AgentUUID: agent.UUID,
			Role:      M.AGENT,
		})
		assert.Equal(t, http.StatusCreated, errCode)

		errCode = M.DeleteProjectAgentMapping(project.ID, agent.UUID)
		assert.Equal(t, http.StatusAccepted, errCode)
	})

}
