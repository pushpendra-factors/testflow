package model

import (
	U "factors/util"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

const (
	ALERT_TYPE_SINGLE_RANGE                            = 1
	ALERT_TYPE_MULTI_RANGE                             = 2
	ALERT_TYPE_QUERY_SHARING                           = 3
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
	QueryID            int64           `json:"query_id"`
	ProjectID          int64           `json:"project_id"`
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
	//below 2 only for reports sharing into slack and email.
	Subject   string          `json:"subject"`
	Message   string          `json:"message"`
	// only for type 2 (daterange a compared to daterange b )
	ComparedTo string `json:"compared_to"`
}
type AlertConfiguration struct {
	IsEmailEnabled             bool            `json:"email_enabled"`
	IsSlackEnabled             bool            `json:"slack_enabled"`
	Emails                     []string        `json:"emails"`
	SlackChannelsAndUserGroups *postgres.Jsonb `json:"slack_channels_and_user_groups"`
}
type SlackChannelsAndUserGroups struct {
	SlackChannelsAndUserGroups map[string][]SlackChannel `json:"slack_channels_and_user_groups"`
}
type SlackChannel struct {
	Name      string `json:"name"`
	Id        string `json:"id"`
	IsPrivate bool   `json:"is_private"`
}

func DecodeAndFetchAlertRelatedStructs(projectID int64, alert Alert) (AlertDescription, AlertConfiguration, KPIQuery, error) {
	var alertDescription AlertDescription
	var alertConfiguration AlertConfiguration
	var kpiQuery KPIQuery

	err := U.DecodePostgresJsonbToStructType(alert.AlertDescription, &alertDescription)
	if err != nil {
		log.Errorf("failed to decode alert description for project_id: %v, alert_name: %s", projectID, alert.AlertName)
		log.Error(err)
		return alertDescription, alertConfiguration, kpiQuery, err
	}

	err = U.DecodePostgresJsonbToStructType(alert.AlertConfiguration, &alertConfiguration)
	if err != nil {
		log.Errorf("failed to decode alert configuration for project_id: %v, alert_name: %s", projectID, alert.AlertName)
		log.Error(err)
		return alertDescription, alertConfiguration, kpiQuery, err
	}

	err = U.DecodePostgresJsonbToStructType(alertDescription.Query, &kpiQuery)
	if err != nil {
		log.Errorf("Error decoding query for project_id: %v, alert_name: %s", projectID, alert.AlertName)
		log.Error(err)
		return alertDescription, alertConfiguration, kpiQuery, err
	}

	return alertDescription, alertConfiguration, kpiQuery, err
}
