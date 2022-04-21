package tests

import (
	"errors"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"net/http"
)

// TODO: Use testify.suites to avoid multiple initializations across these tests.

func SetupProjectReturnDAO() (*model.Project, error) {

	agent, errCode := SetupAgentReturnDAO(getRandomEmail(), "+354522436")
	if errCode != http.StatusCreated {
		return nil, fmt.Errorf("Project Creation failed, agentCreation failed")
	}

	billingAccount, errCode := store.GetStore().GetBillingAccountByAgentUUID(agent.UUID)
	if errCode != http.StatusFound {
		return nil, fmt.Errorf("Project Creation failed, agent billing account not found")
	}

	// Create random project.
	random_project_name := U.RandomLowerAphaNumString(15)

	project, errCode := store.GetStore().CreateProjectWithDependencies(&model.Project{Name: random_project_name},
		agent.UUID, model.ADMIN, billingAccount.ID, true)
	if errCode != http.StatusCreated {
		return nil, fmt.Errorf("Project Creation failed.")
	}

	// Updates the next session start timestamp of project with older timestamp
	// to make the add_session to consider events with older timestamp as next
	// session start timestamp is initialized with project creation timestamp.
	errCode = store.GetStore().UpdateNextSessionStartTimestampForProject(project.ID, 1500000000)
	if errCode != http.StatusAccepted {
		return nil, errors.New("failed to update next session start timestamp")
	}

	return project, nil
}

func SetupProjectUserReturnDAO() (*model.Project, *model.User, error) {
	// Create random project and user.
	project, err := SetupProjectReturnDAO()
	if err != nil {
		return nil, nil, err
	}

	createdUserID, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	if errCode != http.StatusCreated {
		return nil, nil, errors.New("user creation failure")
	}

	user, errCode := store.GetStore().GetUser(project.ID, createdUserID)
	if errCode != http.StatusFound {
		return nil, nil, errors.New("created user not found")
	}

	return project, user, nil
}

// Todo(Dinesh): To be replaced with SetupProjectUserEventNameReturnDAO.
func SetupProjectUserEventName() (uint64, string, string, error) {
	var projectId uint64
	var userId string
	var eventNameId string

	// Create random project and a corresponding eventName and user.

	project, err := SetupProjectReturnDAO()
	if err != nil {
		return projectId, userId, eventNameId, err
	}
	createdUserID, err_code := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	if err_code != http.StatusCreated {
		return projectId, userId, eventNameId, fmt.Errorf("User Creation failed.")
	}
	en, err_code := store.GetStore().CreateOrGetUserCreatedEventName(&model.EventName{ProjectId: project.ID, Name: "login"})
	if err_code != http.StatusCreated {
		return projectId, userId, eventNameId, fmt.Errorf("EventName Creation failed.")
	}
	projectId = project.ID
	userId = createdUserID
	eventNameId = en.ID
	return projectId, userId, eventNameId, nil
}

func SetupProjectUserEventNameReturnDAO() (*model.Project, *model.User, *model.EventName, error) {

	// Create random project and a corresponding eventName and user.
	project, err := SetupProjectReturnDAO()
	if err != nil {
		return nil, nil, nil, err
	}

	createdUserID, err_code := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	if err_code != http.StatusCreated {
		return nil, nil, nil, fmt.Errorf("User Creation failed.")
	}

	user, errCode := store.GetStore().GetUser(project.ID, createdUserID)
	if errCode != http.StatusFound {
		return nil, nil, nil, errors.New("created user not found")
	}

	en, err_code := store.GetStore().CreateOrGetUserCreatedEventName(&model.EventName{ProjectId: project.ID, Name: "login"})
	if err_code != http.StatusConflict && err_code != http.StatusCreated {
		return nil, nil, nil, fmt.Errorf("EventName Creation failed.")
	}

	return project, user, en, nil
}

func getRandomEmail() string {
	email := U.RandomLowerAphaNumString(6) + "@asdfds.local"
	return email
}

func getRandomName() string {
	name := U.RandomLowerAphaNumString(8)
	return name
}

func getRandomAgentUUID() string {
	return "6ba7b814-9dad-11d1-80b4-00c04fd430c8"
}

func SetupAgentReturnDAO(email string, phone string) (*model.Agent, int) {

	if email == "" {
		email = getRandomEmail()
	}

	createAgentParams := &model.CreateAgentParams{Agent: &model.Agent{FirstName: getRandomName(),
		LastName: getRandomName(), Email: email, Phone: phone}, PlanCode: model.FreePlanCode}
	resp, errCode := store.GetStore().CreateAgentWithDependencies(createAgentParams)
	if errCode != http.StatusCreated {
		return nil, errCode
	}
	return resp.Agent, http.StatusCreated
}

func SetupProjectUserEventNameAgentReturnDAO() (*model.Project, *model.User, *model.EventName, *model.Agent, error) {

	// Create random project and a corresponding eventName and user.
	project, err := SetupProjectReturnDAO()
	if err != nil {
		return nil, nil, nil, nil, err
	}

	createdUserID, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	if errCode != http.StatusCreated {
		return nil, nil, nil, nil, fmt.Errorf("User Creation failed.")
	}

	user, errCode := store.GetStore().GetUser(project.ID, createdUserID)
	if errCode != http.StatusFound {
		return nil, nil, nil, nil, errors.New("created user not found")
	}

	en, errCode := store.GetStore().CreateOrGetUserCreatedEventName(&model.EventName{ProjectId: project.ID, Name: "login"})
	if errCode != http.StatusConflict && errCode != http.StatusCreated {
		return nil, nil, nil, nil, fmt.Errorf("EventName Creation failed.")
	}

	agent, errCode := SetupAgentReturnDAO(getRandomEmail(), "+1343545")
	if errCode != http.StatusCreated {
		return nil, nil, nil, nil, fmt.Errorf("Agent Creation failed.")
	}

	_, errCode = store.GetStore().CreateProjectAgentMappingWithDependencies(&model.ProjectAgentMapping{
		ProjectID: project.ID, AgentUUID: agent.UUID})
	if errCode != http.StatusCreated {
		return nil, nil, nil, nil, fmt.Errorf("ProjectAgentMapping Creation failed.")
	}
	return project, user, en, agent, nil
}

func SetupProjectWithAgentDAO() (*model.Project, *model.Agent, error) {
	project, err := SetupProjectReturnDAO()
	if err != nil {
		return nil, nil, err
	}
	agent, errCode := SetupAgentReturnDAO(getRandomEmail(), "+1343545")
	if errCode != http.StatusCreated {
		return nil, nil, fmt.Errorf("Agent Creation failed.")
	}
	_, errCode = store.GetStore().CreateProjectAgentMappingWithDependencies(&model.ProjectAgentMapping{
		ProjectID: project.ID, AgentUUID: agent.UUID})
	if errCode != http.StatusCreated {
		return nil, nil, fmt.Errorf("ProjectAgentMapping Creation failed.")
	}
	return project, agent, nil
}
func SetupProjectWithAdminAgentDAO() (*model.Project, *model.Agent, error) {
	project, err := SetupProjectReturnDAO()
	if err != nil {
		return nil, nil, err
	}
	agent, errCode := SetupAgentReturnDAO(getRandomEmail(), "+1343545")
	if errCode != http.StatusCreated {
		return nil, nil, fmt.Errorf("Agent Creation failed.")
	}
	_, errCode = store.GetStore().CreateProjectAgentMappingWithDependencies(&model.ProjectAgentMapping{
		ProjectID: project.ID, AgentUUID: agent.UUID, Role: model.ADMIN})
	if errCode != http.StatusCreated {
		return nil, nil, fmt.Errorf("ProjectAgentMapping Creation failed.")
	}
	return project, agent, nil
}
func SetupAgentWithProject(projectID uint64) (*model.Agent, error) {
	agent, errCode := SetupAgentReturnDAO(getRandomEmail(), "+1343545")
	if errCode != http.StatusCreated {
		return nil, fmt.Errorf("Agent Creation failed.")
	}
	_, errCode = store.GetStore().CreateProjectAgentMappingWithDependencies(&model.ProjectAgentMapping{
		ProjectID: projectID, AgentUUID: agent.UUID})
	if errCode != http.StatusCreated {
		return nil, fmt.Errorf("ProjectAgentMapping Creation failed.")
	}
	return agent, nil
}

type testData struct {
	Agent          *model.Agent
	Project        *model.Project
	BillingAccount *model.BillingAccount
}

func SetupTestData() (*testData, int) {
	agent, errCode := SetupAgentReturnDAO(getRandomEmail(), "+6753476")
	if errCode != http.StatusCreated {
		return nil, http.StatusInternalServerError
	}

	billingAccount, errCode := store.GetStore().GetBillingAccountByAgentUUID(agent.UUID)
	if errCode != http.StatusFound {
		return nil, http.StatusInternalServerError
	}

	// Create random project.
	random_project_name := U.RandomLowerAphaNumString(15)

	project, err_code := store.GetStore().CreateProjectWithDependencies(&model.Project{Name: random_project_name}, agent.UUID, model.ADMIN, billingAccount.ID, true)
	if err_code != http.StatusCreated {
		return nil, http.StatusInternalServerError
	}

	return &testData{
		Agent:          agent,
		Project:        project,
		BillingAccount: billingAccount,
	}, http.StatusCreated
}
