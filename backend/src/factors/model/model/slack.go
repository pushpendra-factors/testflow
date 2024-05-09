package model

import (
	"encoding/json"
	U "factors/util"
	"io"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

type SlackChannelsAndUserGroups struct {
	SlackChannelsAndUserGroups map[string][]SlackChannel `json:"slack_channels_and_user_groups"`
}

type SlackChannel struct {
	Name      string `json:"name"`
	Id        string `json:"id"`
	IsPrivate bool   `json:"is_private"`
}

type SlackGetUsersResponse struct {
	Ok               bool                  `json:"ok"`
	Members          []SlackMember         `json:"members"`
	CacheTs          int64                 `json:"cache_ts"`
	ResponseMetadata SlackResponseMetadata `json:"response_metadata"`
	Error            string                `json:"error"`
}

type SlackResponseMetadata map[string]interface{}

type SlackMember struct {
	Id        string        `json:"id"`
	TeamId    string        `json:"team_id"`
	Name      string        `json:"name"`
	RealName  string        `json:"real_name"`
	IsAdmin   bool          `json:"is_admin"`
	IsOwner   bool          `json:"is_owner"`
	IsBot     bool          `json:"is_bot"`
	IsAppUser bool          `json:"is_app_user"`
	Deleted   bool          `json:"deleted"`
	Profile   MemberProfile `json:"profile"`
}

type MemberProfile struct {
	DisplayName string `json:"display_name"`
	RealName    string `json:"real_name"`
	Email       string `json:"email"`
	Team        string `json:"team"`
}

type SlackUsersList struct {
	ProjectID    int64           `gorm:"column:project_id; primary_key:true" json:"project_id"`
	AgentID      string          `gorm:"column:agent_id" json:"agent_id"`
	UsersList    *postgres.Jsonb `gorm:"column:users_list" json:"users_list"`
	LastSyncTime time.Time       `gorm:"column:last_sync_time" json:"last_sync_time"`
}

type SlackEventsApiURLVerificationEvent struct {
	Token     string `json:"token"`
	Challenge string `json:"challenge"`
	Type      string `json:"type"`
}

type SlackEventType struct {
	Type string `json:"type"`
}

type SlackUninstallAPIEvent struct {
	Token     string                 `json:"token"`
	TeamID    string                 `json:"team_id"`
	APIAppID  string                 `json:"api_app_id"`
	Event     SlackAppUninstallEvent `json:"event"`
	Type      string                 `json:"type"`
	EventID   string                 `json:"event_id"`
	EventTime int                    `json:"event_time"`
}

type SlackAppUninstallEvent struct {
	Type string `json:"type"`
}

func (eventStruct SlackEventType) ParseEvent(jsonBody *io.ReadCloser) (SlackEventType, error) {

	decoder := json.NewDecoder(*jsonBody)
	if err := decoder.Decode(&eventStruct); U.IsJsonError(err) {
		log.WithError(err).Error("Tracking failed. Json Decoding failed.")
		return SlackEventType{}, err
	}
	return eventStruct, nil
}

func (eventStruct SlackUninstallAPIEvent) ParseEvent(jsonBody *io.ReadCloser) (SlackUninstallAPIEvent, error) {

	decoder := json.NewDecoder(*jsonBody)
	if err := decoder.Decode(&eventStruct); U.IsJsonError(err) {
		log.WithError(err).Error("Tracking failed. Json Decoding failed.")
		return SlackUninstallAPIEvent{}, err
	}
	return eventStruct, nil
}

func (eventStruct SlackEventsApiURLVerificationEvent) ParseEvent(jsonBody *io.ReadCloser) (SlackEventsApiURLVerificationEvent, error) {

	decoder := json.NewDecoder(*jsonBody)
	if err := decoder.Decode(&eventStruct); U.IsJsonError(err) {
		log.WithError(err).Error("Tracking failed. Json Decoding failed.")
		return SlackEventsApiURLVerificationEvent{}, err
	}
	return eventStruct, nil
}
