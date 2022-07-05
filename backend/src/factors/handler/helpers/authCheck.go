package helpers

import (
	"factors/model/model"
	"factors/model/store"
	"fmt"

	"net/http"
)

// IsAdmin - To check if the logged in user is admin of the project
func IsAdmin(ProjectID int64, loggedInAgentUUID string) bool {
	loggedInAgentPAM, errCode := store.GetStore().GetProjectAgentMapping(ProjectID, loggedInAgentUUID)
	if errCode != http.StatusFound {
		return false
	}

	if loggedInAgentPAM.Role != model.ADMIN {
		return false
	}
	return true
}

func IsDemoProjectInAuthorizedProjects(authorizedProjects []uint64, id string) bool {
	for _, project := range authorizedProjects {
		projectString := fmt.Sprintf("%v", project)
		if projectString == id {
			return true
		}
	}
	return false
}
