package memsql

import (
	"errors"
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"net/http"
	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) SetAuthTokenforTeamsIntegration(projectID int64, agentUUID string, authTokens model.TeamsAccessTokens) error {
	db := C.GetServices().Db
	_, errCode := store.GetProjectAgentMapping(projectID, agentUUID)
	if errCode != http.StatusFound {
		log.Error("Project agent mapping not found.")
		return errors.New("Project agent mapping not found.")
	}
	var agent model.Agent
	err := db.Where("uuid = ?", agentUUID).Find(&agent).Error
	if err != nil {
		log.Error(err)
		return err
	}
	var token model.TeamsAuthTokens
	if IsEmptyPostgresJsonb(agent.TeamsAccessTokens) {
		token = make(map[int64]model.TeamsAccessTokens)
	} else {
		err = U.DecodePostgresJsonbToStructType(agent.TeamsAccessTokens, &token)
		if err != nil {
			log.Error(err)
			return err
		}
	}
	token[projectID] = authTokens
	// convert token to json
	TokenJson, err := U.EncodeStructTypeToPostgresJsonb(token)
	if err != nil {
		log.Error(err)
		return err
	}
	// update the db
	err = db.Model(&model.Agent{}).Where("uuid = ?", agentUUID).Update("teams_access_tokens", TokenJson).Error
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}
func (store *MemSQL) GetTeamsAuthTokens(projectID int64, agentUUID string) (model.TeamsAccessTokens, error) {
	db := C.GetServices().Db
	var agent model.Agent
	err := db.Where("uuid = ?", agentUUID).Find(&agent).Error
	if err != nil {
		log.Error(err)
		return model.TeamsAccessTokens{}, err
	}
	var token model.TeamsAuthTokens

	if IsEmptyPostgresJsonb(agent.TeamsAccessTokens) {
		return model.TeamsAccessTokens{}, errors.New("No Teams auth token found")
	}

	err = U.DecodePostgresJsonbToStructType(agent.TeamsAccessTokens, &token)
	if err != nil && err.Error() != "Empty jsonb object" {
		log.Error(err)
		return model.TeamsAccessTokens{}, err
	}
	if err != nil && err.Error() == "Empty jsonb object" {
		return model.TeamsAccessTokens{}, errors.New("No Teams auth token found")
	}
	if _, ok := token[projectID]; !ok {
		return model.TeamsAccessTokens{}, errors.New("Teams token not found.")
	}
	return token[projectID], nil

}
// TODO : add func for using refresh token to obtain new access token.

func (store *MemSQL) DeleteTeamsIntegration(projectID int64, agentUUID string) error {
	db := C.GetServices().Db
	var agent model.Agent
	err := db.Where("uuid = ?", agentUUID).Find(&agent).Error
	if err != nil {
		log.Error(err)
		return err
	}
	var token model.TeamsAuthTokens
	err = U.DecodePostgresJsonbToStructType(agent.TeamsAccessTokens, &token)
	if err != nil && err.Error() != "Empty jsonb object" {
		log.Error(err)
		return err
	}
	if err != nil && err.Error() == "Empty jsonb object" {
		return errors.New("No teams auth token found")
	}
	var newToken model.TeamsAuthTokens
	newToken = make(map[int64]model.TeamsAccessTokens)
	for k, v := range token {
		if k != projectID {
			newToken[k] = v
		}
	}
	TokenJson, err := U.EncodeStructTypeToPostgresJsonb(newToken)
	if err != nil {
		log.Error(err)
		return err
	}
	// update the db
	err = db.Model(&model.Agent{}).Where("uuid = ?", agentUUID).Update("teams_access_tokens", TokenJson).Error
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}
