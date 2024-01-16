package memsql

import (
	"errors"
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) SetAuthTokenforSlackIntegration(projectID int64, agentUUID string, authTokens model.SlackAccessTokens) error {
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
	var token model.SlackAuthTokens
	if IsEmptyPostgresJsonb(agent.SlackAccessTokens) {
		token = make(map[int64]model.SlackAccessTokens)
	} else {
		err = U.DecodePostgresJsonbToStructType(agent.SlackAccessTokens, &token)
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
	err = db.Model(&model.Agent{}).Where("uuid = ?", agentUUID).Update("slack_access_tokens", TokenJson).Error
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}
func (store *MemSQL) GetSlackAuthToken(projectID int64, agentUUID string) (model.SlackAccessTokens, error) {
	db := C.GetServices().Db
	var agent model.Agent
	err := db.Where("uuid = ?", agentUUID).Find(&agent).Error
	if err != nil {
		log.Error(err)
		return model.SlackAccessTokens{}, err
	}
	var token model.SlackAuthTokens

	if IsEmptyPostgresJsonb(agent.SlackAccessTokens) {
		return model.SlackAccessTokens{}, errors.New("No slack auth token found")
	}

	err = U.DecodePostgresJsonbToStructType(agent.SlackAccessTokens, &token)
	if err != nil && err.Error() != "Empty jsonb object" {
		log.Error(err)
		return model.SlackAccessTokens{}, err
	}
	if err != nil && err.Error() == "Empty jsonb object" {
		return model.SlackAccessTokens{}, errors.New("No slack auth token found")
	}
	if _, ok := token[projectID]; !ok {
		return model.SlackAccessTokens{}, errors.New("Slack token not found.")
	}
	return token[projectID], nil

}
func (store *MemSQL) DeleteSlackIntegration(projectID int64, agentUUID string) error {
	db := C.GetServices().Db
	var agent model.Agent
	err := db.Where("uuid = ?", agentUUID).Find(&agent).Error
	if err != nil {
		log.Error(err)
		return err
	}
	var token model.SlackAuthTokens
	err = U.DecodePostgresJsonbToStructType(agent.SlackAccessTokens, &token)
	if err != nil && err.Error() != "Empty jsonb object" {
		log.Error(err)
		return err
	}
	if err != nil && err.Error() == "Empty jsonb object" {
		return errors.New("No slack auth token found")
	}
	var newToken model.SlackAuthTokens
	newToken = make(map[int64]model.SlackAccessTokens)
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
	err = db.Model(&model.Agent{}).Where("uuid = ?", agentUUID).Update("slack_access_tokens", TokenJson).Error
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func (store *MemSQL) GetSlackUsersListFromDb(projectID int64, agentID string) ([]model.SlackUser, int, error) {
	if projectID == 0 || agentID == "" {
		return nil, http.StatusBadRequest, fmt.Errorf("invalid parameters")
	}

	logFields := log.Fields{
		"project_id": projectID,
		"agent_id":   agentID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	db := C.GetServices().Db

	var usersList model.SlackUsersList
	err := db.Where("project_id = ?", projectID).Find(&usersList).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound, err
		}
		logCtx.WithError(err).Error("failed to find slack users list")
		return nil, http.StatusInternalServerError, err
	}

	slackUsers := make([]model.SlackUser, 0)
	if err = U.DecodePostgresJsonbToStructType(usersList.UsersList, &slackUsers); err != nil {
		logCtx.WithError(err).Error("failed to decode slack users list")
		return nil, http.StatusInternalServerError, err
	}

	return slackUsers, http.StatusFound, nil
}

func (store *MemSQL) UpdateSlackUsersListForProject(projectID int64, fields map[string]interface{}) (int, error) {
	if projectID == 0 {
		return http.StatusBadRequest, fmt.Errorf("invalid parameters")
	}

	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	if fields["agent_id"] == nil || fields["users_list"] == nil {
		return http.StatusBadRequest, fmt.Errorf("invalid fields provided for updation")
	}
	fields["last_sync_time"] = time.Now().Unix()

	db := C.GetServices().Db
	err := db.Model(&model.SlackUsersList{}).Where("project_id = ?", projectID).Updates(fields).Error
	if err != nil {
		log.WithFields(logFields).WithError(err).Error("failed to update slack users list table")
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}
