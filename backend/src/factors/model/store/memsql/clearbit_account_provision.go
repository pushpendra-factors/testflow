package memsql

import (
	"encoding/json"
	"factors/config"
	"factors/integration/clear_bit"
	"factors/model/model"
	"io"

	"github.com/jinzhu/gorm/dialects/postgres"
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

	provisionAPIKey := config.GetClearbitProvisionAccountAPIKey()
	result, err := clear_bit.GetClearbitProvisionAccountResponse(API_URL, emailId, domainName, provisionAPIKey)
	if err != nil {
		return err
	}
	defer result.Body.Close()

	var response model.ClearbitProvisionAPIResponse
	body, err := io.ReadAll(result.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return err
	}

	responseJSON := postgres.Jsonb{RawMessage: body}

	//Update Project Settings Table
	store.UpdateProjectSettings(projectId, &model.ProjectSetting{FactorsClearbitKey: response.Keys.Secret, ClearbitProvisionAPIResponse: &responseJSON})

	return nil
}
