package model

import (
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

const (
	ALERT_TYPE_SINGLE_RANGE                            = 1
	ALERT_TYPE_MULTI_RANGE                             = 2
	LAST_WEEK                                          = "last_week"
	LAST_MONTH                                         = "last_month"
	LAST_QUARTER                                       = "last_quarter"
	IS_LESS_THAN                                       = "is_less_than"
	IS_GREATER_THAN                                    = "is_greater_than"
	DECREASED_BY_MORE_THAN                             = "decreased_by_more_than"
	INCREASED_BY_MORE_THAN                             = "increased_by_more_than"
	INCREASED_OR_DECREASED_BY_MORE_THAN                = "increased_or_decreased_by_more_than"
	PERCENTAGE_HAS_DECREASED_BY_MORE_THAN              = "%_has_decreased_by_more_than"
	PERCENTAGE_HAS_INCREASED_BY_MORE_THAN              = "%_has_increased_by_more_than"
	PERCENTAGE_HAS_INCREASED_OR_DECREASED_BY_MORE_THAN = "%_has_increased_or_decreased_by_more_than"
	PREVIOUS_PERIOD                                    = "previous_period"
	SAME_PERIOD_LAST_YEAR                              = "same_period_last_year"
)

// last quarter, last month to be added
var ValidDateRanges = []string{LAST_WEEK}
var ValidDateRangeComparisions = []string{
	PREVIOUS_PERIOD, SAME_PERIOD_LAST_YEAR,
}

var ValidValues = []string{}
var ValidOperators = []string{IS_LESS_THAN, IS_GREATER_THAN, DECREASED_BY_MORE_THAN, INCREASED_BY_MORE_THAN, INCREASED_OR_DECREASED_BY_MORE_THAN, PERCENTAGE_HAS_DECREASED_BY_MORE_THAN, PERCENTAGE_HAS_INCREASED_BY_MORE_THAN, PERCENTAGE_HAS_INCREASED_OR_DECREASED_BY_MORE_THAN}

type Alert struct {
	ID                 string          `gorm:"type:uuid;default:uuid_generate_v4()" json:"id"`
	ProjectID          uint64          `json:"project_id"`
	AlertName          string          `json:"alert_name"`
	CreatedBy          string          `json:"created_by"`
	AlertType          int             `json:"alert_type"`
	AlertDescription   *postgres.Jsonb `json:"alert_description"`
	AlertConfiguration *postgres.Jsonb `json:"alert_configuration"`
	LastAlertSent      bool            `json:"last_alert_sent"`
	LastRunTime        time.Time       `json:"last_run_time"`
	IsDeleted          bool            `json:"is_deleted"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}

type AlertDescription struct {
	Name      string          `json:"name"`
	Query     *postgres.Jsonb `json:"query"`
	QueryType string          `json:"query_type"`
	Operator  string          `json:"operator"`
	Value     string          `json:"value"`
	DateRange string          `json:"date_range"`
	// only for type 2 (daterange a compared to daterange b )
	ComparedTo string `json:"compared_to"`
}
type AlertConfiguration struct {
	IsEmailEnabled             bool           `json:"email_enabled"`
	IsSlackEnabled             bool           `json:"slack_enabled"`
	Emails                     []string       `json:"emails"`
	SlackChannelsAndUserGroups *postgres.Jsonb `json:"slack_channels_and_user_groups"`
}
type SlackChannels struct {
	Channels map[string][]SlackChannel `json:"channels"`
}
type SlackChannel struct {
	ChannelName string `json:"channel_name"`
	ChannelID   string `json:"channel_id"`
	IsPrivate   bool   `json:"is_private"`
}
