package helpers

import (
	M "factors/model"
	"net/http"
)

// IsAdmin - To check if the logged in user is admin of the project
func IsAdmin(ProjectID uint64, loggedInAgentUUID string) bool {
	loggedInAgentPAM, errCode := M.GetProjectAgentMapping(ProjectID, loggedInAgentUUID)
	if errCode != http.StatusFound {
		return false
	}

	if loggedInAgentPAM.Role != M.ADMIN {
		return false
	}
	return true
}
