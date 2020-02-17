package tests

import (
	M "factors/model"
	U "factors/util"
	"fmt"
	"net/http"
)

// TODO: Use testify.suites to avoid multiple initializations across these tests.

func SetupProjectReturnDAO() (*M.Project, error) {

	agent, errCode := SetupAgentReturnDAO(getRandomEmail(), "+354522436")
	if errCode != http.StatusCreated {
		return nil, fmt.Errorf("Project Creation failed, agentCreation failed")
	}

	billingAccount, errCode := M.GetBillingAccountByAgentUUID(agent.UUID)
	if errCode != http.StatusFound {
		return nil, fmt.Errorf("Project Creation failed, agent billing account not found")
	}

	// Create random project.
	random_project_name := U.RandomLowerAphaNumString(15)

	project, err_code := M.CreateProjectWithDependencies(&M.Project{Name: random_project_name}, agent.UUID, M.ADMIN, billingAccount.ID)
	if err_code != http.StatusCreated {
		return nil, fmt.Errorf("Project Creation failed.")
	}
	return project, nil
}

func SetupProjectUserReturnDAO() (*M.Project, *M.User, error) {
	// Create random project and user.
	project, err := SetupProjectReturnDAO()
	if err != nil {
		return nil, nil, err
	}

	user, err_code := M.CreateUser(&M.User{ProjectId: project.ID})
	if err_code != http.StatusCreated {
		return nil, nil, fmt.Errorf("User Creation failed.")
	}

	return project, user, nil
}

// Todo(Dinesh): To be replaced with SetupProjectUserEventNameReturnDAO.
func SetupProjectUserEventName() (uint64, string, uint64, error) {
	var projectId uint64
	var userId string
	var eventNameId uint64

	// Create random project and a corresponding eventName and user.

	project, err := SetupProjectReturnDAO()
	if err != nil {
		return projectId, userId, eventNameId, err
	}
	user, err_code := M.CreateUser(&M.User{ProjectId: project.ID})
	if err_code != http.StatusCreated {
		return projectId, userId, eventNameId, fmt.Errorf("User Creation failed.")
	}
	en, err_code := M.CreateOrGetUserCreatedEventName(&M.EventName{ProjectId: project.ID, Name: "login"})
	if err_code != http.StatusCreated {
		return projectId, userId, eventNameId, fmt.Errorf("EventName Creation failed.")
	}
	projectId = project.ID
	userId = user.ID
	eventNameId = en.ID
	return projectId, userId, eventNameId, nil
}

func SetupProjectUserEventNameReturnDAO() (*M.Project, *M.User, *M.EventName, error) {

	// Create random project and a corresponding eventName and user.
	project, err := SetupProjectReturnDAO()
	if err != nil {
		return nil, nil, nil, err
	}

	user, err_code := M.CreateUser(&M.User{ProjectId: project.ID})
	if err_code != http.StatusCreated {
		return nil, nil, nil, fmt.Errorf("User Creation failed.")
	}

	en, err_code := M.CreateOrGetUserCreatedEventName(&M.EventName{ProjectId: project.ID, Name: "login"})
	if err_code != http.StatusConflict && err_code != http.StatusCreated {
		return nil, nil, nil, fmt.Errorf("EventName Creation failed.")
	}

	return project, user, en, nil
}

func getRandomEmail() string {
	email := U.RandomLowerAphaNumString(6) + "@asdfds.local"
	return email
}

func getRandomAgentUUID() string {
	return "6ba7b814-9dad-11d1-80b4-00c04fd430c8"
}

func SetupAgentReturnDAO(email string, phone string) (*M.Agent, int) {

	if email == "" {
		email = getRandomEmail()
	}

	createAgentParams := &M.CreateAgentParams{Agent: &M.Agent{Email: email, Phone: phone}, PlanCode: M.FreePlanCode}
	resp, errCode := M.CreateAgentWithDependencies(createAgentParams)
	if errCode != http.StatusCreated {
		return nil, errCode
	}
	return resp.Agent, http.StatusCreated
}

func SetupProjectWithAgentDAO() (*M.Project, *M.Agent, error) {
	project, err := SetupProjectReturnDAO()
	if err != nil {
		return nil, nil, err
	}
	agent, errCode := SetupAgentReturnDAO(getRandomEmail(), "+1343545")
	if errCode != http.StatusCreated {
		return nil, nil, fmt.Errorf("Agent Creation failed.")
	}
	_, errCode = M.CreateProjectAgentMappingWithDependencies(&M.ProjectAgentMapping{
		ProjectID: project.ID, AgentUUID: agent.UUID})
	if errCode != http.StatusCreated {
		return nil, nil, fmt.Errorf("ProjectAgentMapping Creation failed.")
	}
	return project, agent, nil
}

type testData struct {
	Agent          *M.Agent
	Project        *M.Project
	BillingAccount *M.BillingAccount
}

func SetupTestData() (*testData, int) {
	agent, errCode := SetupAgentReturnDAO(getRandomEmail(), "+6753476")
	if errCode != http.StatusCreated {
		return nil, http.StatusInternalServerError
	}

	billingAccount, errCode := M.GetBillingAccountByAgentUUID(agent.UUID)
	if errCode != http.StatusFound {
		return nil, http.StatusInternalServerError
	}

	// Create random project.
	random_project_name := U.RandomLowerAphaNumString(15)

	project, err_code := M.CreateProjectWithDependencies(&M.Project{Name: random_project_name}, agent.UUID, M.ADMIN, billingAccount.ID)
	if err_code != http.StatusCreated {
		return nil, http.StatusInternalServerError
	}

	return &testData{
		Agent:          agent,
		Project:        project,
		BillingAccount: billingAccount,
	}, http.StatusCreated
}
