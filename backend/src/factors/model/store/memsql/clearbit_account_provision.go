package memsql

import (
	"encoding/json"
	"factors/config"
	"factors/integration/clear_bit"
	"factors/model/model"
	"fmt"
	"io"
	"net/http"

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

const API_URL = "https://clearbit.com/api/v1/partnerships/account"

func (store *MemSQL) ProvisionClearbitAccount(projectIdList []int64, emailList []string, domainList []string) map[int64]interface{} {
	projectIdToErrorMap := make(map[int64]interface{})

	for i := range projectIdList {
		err := store.ProvisionClearbitAccountForSingleProject(projectIdList[i], emailList[i], domainList[i])
		if err != nil {
			projectIdToErrorMap[projectIdList[i]] = err
		}
	}
	return projectIdToErrorMap

}

func (store *MemSQL) ProvisionClearbitAccountForSingleProject(projectId int64, emailId string, domainName string) error {

	logCtx := log.WithFields(log.Fields{
		"project_id": projectId,
	})

	provisionAPIKey := config.GetClearbitProvisionAccountAPIKey()

	result, err := clear_bit.GetClearbitProvisionAccountResponse(API_URL, emailId, domainName, provisionAPIKey)
	if err != nil {
		logCtx.Error(err)
		return err
	}
	defer result.Body.Close()

	var response model.ClearbitProvisionAPIResponse
	body, err := io.ReadAll(result.Body)
	if err != nil {
		logCtx.Error(err)
		return err
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		logCtx.Error(err)
		return err
	}

	responseJSON := postgres.Jsonb{RawMessage: body}

	if result.StatusCode != http.StatusOK {
		logCtx.Error("failed provision clearbit account with response: ", &responseJSON)
		return fmt.Errorf("failed provision clearbit account")
	}

	//Update Project Settings Table
	_, errCode := store.UpdateProjectSettings(projectId, &model.ProjectSetting{FactorsClearbitKey: response.Keys.Secret, ClearbitProvisionAccResponse: &responseJSON})
	if errCode != http.StatusAccepted {
		logCtx.Error("failed to UpdateProjectSettings with clearbit response: ", &responseJSON)
		return fmt.Errorf("failed to update project settings")
	}

	return nil
}

func (store *MemSQL) IsClearbitAccountProvisioned(projectId int64) (bool, error) {

	isProvisioned := false
	settings, errCode := store.GetProjectSetting(projectId)
	if errCode != http.StatusFound {
		return isProvisioned, fmt.Errorf("Get Project Settings Failed")
	}

	if settings.FactorsClearbitKey != "" {
		isProvisioned = true
	}
	return isProvisioned, nil
}

func (store *MemSQL) ProvisionClearbitAccountByAdminEmailAndDomain(projectId int64) (int, string) {

	logCtx := log.WithField("project_id", projectId)
	isProvisioned, err := store.IsClearbitAccountProvisioned(projectId)
	if err != nil {
		logCtx.Error("Failed checking if clearbit account is provisoned")
		return http.StatusInternalServerError, "Failed checking if clearbit account is provisoned"
	}

	if !isProvisioned {

		adminMail, errCode := store.GetProjectAgentLatestAdminEmailByProjectId(projectId)
		if errCode != http.StatusFound {
			logCtx.Error("Failed fetching admin mail")
			return errCode, "Failed fetching admin mail"
		}

		project, errCode := store.GetProject(projectId)
		if errCode != http.StatusFound {
			logCtx.Error("Failed fetching projects")
			return errCode, "Failed fetching projects"
		}

		err = store.ProvisionClearbitAccountForSingleProject(projectId, adminMail, project.ClearbitDomain)
		if err != nil {
			logCtx.Error("Failed provisioning clearbit account by admin mail.")
			return http.StatusInternalServerError, "Failed provisioning clearbit account by admin mail"
		}
	}

	return http.StatusOK, ""
}
