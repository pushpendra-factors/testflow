package operations

type UserDataOutput struct {
	UserId         string                 `json:"user_id"`
	UserAttributes map[string]interface{} `json:"user_properties"`
}

type EventOutput struct {
	UserId          string                 `json:"user_id"`
	Event           string                 `json:"event_name"`
	Timestamp       int                    `json:"timestamp"`
	UserAttributes  map[string]interface{} `json:"user_properties"`
	EventAttributes map[string]interface{} `json:"event_properties"`
}
