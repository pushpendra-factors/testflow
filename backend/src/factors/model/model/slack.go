package model

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
