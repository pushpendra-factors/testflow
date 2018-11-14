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
	project, err_code := M.CreateProject(&M.Project{Name: random_project_name})
	if err_code != M.DB_SUCCESS {
		return projectId, fmt.Errorf("Project Creation failed.")
	}
	projectId = project.ID
	return projectId, nil
}

func SetupProjectReturnDAO() (*M.Project, error) {
	// Create random project.
	random_project_name := U.RandomLowerAphaNumString(15)
	project, err_code := M.CreateProject(&M.Project{Name: random_project_name})
	if err_code != M.DB_SUCCESS {
		return nil, fmt.Errorf("Project Creation failed.")
	}
	return project, nil
}

func SetupProjectUserReturnDAO() (*M.Project, *M.User, error) {
	// Create random project and user.
	random_project_name := U.RandomLowerAphaNumString(15)
	project, err_code := M.CreateProject(&M.Project{Name: random_project_name})
	if err_code != M.DB_SUCCESS {
		return nil, nil, fmt.Errorf("Project Creation failed.")
	}

	user, err_code := M.CreateUser(&M.User{ProjectId: project.ID})
	if err_code != M.DB_SUCCESS {
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
	project, err_code := M.CreateProject(&M.Project{Name: random_project_name})
	if err_code != M.DB_SUCCESS {
		return projectId, userId, eventNameId, fmt.Errorf("Project Creation failed.")
	}
	user, err_code := M.CreateUser(&M.User{ProjectId: project.ID})
	if err_code != M.DB_SUCCESS {
		return projectId, userId, eventNameId, fmt.Errorf("User Creation failed.")
	}
	en, err_code := M.CreateOrGetEventName(&M.EventName{ProjectId: project.ID, Name: "login"})
	if err_code != M.DB_SUCCESS {
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
	project, err_code := M.CreateProject(&M.Project{Name: random_project_name})
	if err_code != M.DB_SUCCESS {
		return nil, nil, nil, fmt.Errorf("Project Creation failed.")
	}

	user, err_code := M.CreateUser(&M.User{ProjectId: project.ID})
	if err_code != M.DB_SUCCESS {
		return nil, nil, nil, fmt.Errorf("User Creation failed.")
	}

	en, err_code := M.CreateOrGetEventName(&M.EventName{ProjectId: project.ID, Name: "login"})
	if err_code != http.StatusConflict && err_code != M.DB_SUCCESS {
		return nil, nil, nil, fmt.Errorf("EventName Creation failed.")
	}

	return project, user, en, nil
}

func SetupProjectDependenciesReturnDAO(project *M.Project) (*M.Project, error) {
	_, errCode := M.CreateProjectDependencies(project)
	if errCode != M.DB_SUCCESS {
		return nil, fmt.Errorf("Project depencies setup failed for project : %d", project.ID)
	}
	return project, nil
}
