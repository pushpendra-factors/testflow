package helpers

import (
	"factors/model/model"
	"factors/model/store"

	"net/http"
)

// IsAdmin - To check if the logged in user is admin of the project
func IsAdmin(ProjectID uint64, loggedInAgentUUID string) bool {
	loggedInAgentPAM, errCode := store.GetStore().GetProjectAgentMapping(ProjectID, loggedInAgentUUID)
	if errCode != http.StatusFound {
		return false
	}

	if loggedInAgentPAM.Role != model.ADMIN {
		return false
	}
	return true
}

func IsDemoProjectInAuthorizedProjects(authorizedProjects []uint64, id uint64) bool {
	for _, project := range authorizedProjects {
		if project == id {
			return true
		}
	}
	return false
}
