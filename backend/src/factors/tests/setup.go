package tests

import (
	M "factors/model"
	U "factors/util"
	"fmt"
	"net/http"
)

// TODO: Use testify.suites to avoid multiple initializations across these tests.

// Todo(Dinesh): To be replaced with SetupProjectReturnDAO.
func SetupProject() (uint64, error) {
	var projectId uint64

	// Create random project.
	random_project_name := U.RandomLowerAphaNumString(15)
	project, err_code := M.CreateProjectWithDependencies(&M.Project{Name: random_project_name})
	if err_code != http.StatusCreated {
		return projectId, fmt.Errorf("Project Creation failed.")
	}
	projectId = project.ID
	return projectId, nil
}

func SetupProjectReturnDAO() (*M.Project, error) {
	// Create random project.
	random_project_name := U.RandomLowerAphaNumString(15)
	project, err_code := M.CreateProjectWithDependencies(&M.Project{Name: random_project_name})
	if err_code != http.StatusCreated {
		return nil, fmt.Errorf("Project Creation failed.")
	}
	return project, nil
}

func SetupProjectUserReturnDAO() (*M.Project, *M.User, error) {
	// Create random project and user.
	random_project_name := U.RandomLowerAphaNumString(15)
	project, err_code := M.CreateProjectWithDependencies(&M.Project{Name: random_project_name})
	if err_code != http.StatusCreated {
		return nil, nil, fmt.Errorf("Project Creation failed.")
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
	random_project_name := U.RandomLowerAphaNumString(15)
	project, err_code := M.CreateProjectWithDependencies(&M.Project{Name: random_project_name})
	if err_code != http.StatusCreated {
		return projectId, userId, eventNameId, fmt.Errorf("Project Creation failed.")
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
	random_project_name := U.RandomLowerAphaNumString(15)
	project, err_code := M.CreateProjectWithDependencies(&M.Project{Name: random_project_name})
	if err_code != http.StatusCreated {
		return nil, nil, nil, fmt.Errorf("Project Creation failed.")
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

func SetupAgentReturnDAO() (*M.Agent, error) {
	agent, errCode := M.CreateAgent(&M.Agent{
		Email: getRandomEmail(),
	})
	if errCode != http.StatusCreated {
		return nil, fmt.Errorf("Agent Creation failed.")
	}
	return agent, nil
}

func SetupProjectWithAgentDAO() (*M.Project, *M.Agent, error) {
	project, err := SetupProjectReturnDAO()
	if err != nil {
		return nil, nil, err
	}
	agent, err := SetupAgentReturnDAO()
	if err != nil {
		return nil, nil, err
	}
	_, errCode := M.CreateProjectAgentMapping(&M.ProjectAgentMapping{ProjectID: project.ID, AgentUUID: agent.UUID})
	if errCode != http.StatusCreated {
		return nil, nil, fmt.Errorf("ProjectAgentMapping Creation failed.")
	}
	return project, agent, nil
}
