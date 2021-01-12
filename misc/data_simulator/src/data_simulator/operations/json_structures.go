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

type AdwordsDocument struct {
	ProjectID         uint64            `json:"project_id"`
	CustomerAccountID string            `json:"customer_acc_id"`
	TypeAlias         string            `json:"type_alias"`
	Timestamp         int64             `json:"timestamp"`
	ID                string            `json:"id"`
	Value             map[string]string `json:"value"`
}
