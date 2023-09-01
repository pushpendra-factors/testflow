package model

import (
	U "factors/util"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

type Profile struct {
	Identity                   string                 `json:"identity"`
	Properties                 *postgres.Jsonb        `json:"-"`
	Name                       string                 `json:"name,omitempty"`
	HostName                   string                 `json:"host_name,omitempty"`
	IsAnonymous                bool                   `json:"is_anonymous"`
	LastActivity               time.Time              `json:"last_activity"`
	PropertiesUpdatedTimestamp int64                  `json:"-"`
	TableProps                 map[string]interface{} `json:"table_props"`
	Score                      float64                `json:"score"`
	Engagement                 string                 `json:"engagement,omitempty"`
}

type MinMaxUpdatedAt struct {
	MinUpdatedAt time.Time `json:"min_updated_at"`
	MaxUpdatedAt time.Time `json:"max_updated_at"`
}

type ContactDetails struct {
	UserId        string                 `json:"user_id"`
	IsAnonymous   bool                   `json:"is_anonymous"`
	Properties    *postgres.Jsonb        `json:"-"`
	LeftPaneProps map[string]interface{} `json:"left_pane_props"`
	Milestones    map[string]interface{} `json:"milestones"`
	Name          string                 `json:"name,omitempty"`
	Company       string                 `json:"company"`
	Account       string                 `json:"account,omitempty"`
	UserActivity  []UserActivity         `json:"user_activities,omitempty"`
}

type GroupsInfo struct {
	GroupName       string `json:"group_name"`
	AssociatedGroup string `json:"associated_group"`
}

type UserActivity struct {
	EventName   string          `json:"event_name"`
	EventType   string          `json:"event_type"`
	DisplayName string          `json:"display_name"`
	AliasName   string          `json:"alias_name,omitempty"`
	Properties  *postgres.Jsonb `json:"properties,omitempty"`
	Timestamp   uint64          `json:"timestamp"`
	Icon        string          `json:"icon"`
}

type TimelinePayload struct {
	Query        Query           `json:"query"`
	SearchFilter []QueryProperty `json:"search_filter"`
}

type AccountDetails struct {
	Properties      *postgres.Jsonb        `json:"-"`
	HostName        string                 `json:"host_name"`
	Name            string                 `json:"name"`
	LeftPaneProps   map[string]interface{} `json:"left_pane_props"`
	Milestones      map[string]interface{} `json:"milestones"`
	AccountTimeline []UserTimeline         `json:"account_timeline"`
	Overview        Overview               `json:"overview"`
}

type Overview struct {
	Temperature float32            `json:"temperature"` // Normalised Score for base 100
	Engagement  string             `json:"engagement"`  // Hot, Warm, Cold
	UsersCount  int64              `json:"users_count"` // Number of Associated Users
	TimeActive  float64            `json:"time_active"` // in seconds
	ScoresList  map[string]float32 `json:"scores_list"` // Score trends list
	TopPages    []TopPage          `json:"top_pages"`
	TopUsers    []TopUser          `json:"top_users"`
}

type TopPage struct {
	PageUrl          string  `json:"page_url"`
	Views            int64   `json:"views"`
	UsersCount       int64   `json:"users_count"`
	TotalTime        float64 `json:"total_time"` // in seconds
	AvgScrollPercent float64 `json:"avg_scroll_percent"`
}

type TopUser struct {
	Name                string  `json:"name"`
	NumPageViews        int64   `json:"num_page_views"`
	AnonymousUsersCount int64   `json:"-"`
	ActiveTime          float64 `json:"active_time"` // in seconds
	NumOfPages          int64   `json:"num_of_pages"`
}

type UserTimeline struct {
	UserId         string         `json:"user_id"`
	IsAnonymous    bool           `json:"is_anonymous"`
	UserName       string         `json:"user_name"`
	AdditionalProp string         `json:"additional_prop"`
	UserActivities []UserActivity `json:"user_activities"`
}

// Constants
const (
	PROFILE_TYPE_USER    = "user"
	PROFILE_TYPE_ACCOUNT = "account"
)
const (
	COLUMN_NAME_ID               = "id"
	COLUMN_NAME_CUSTOMER_USER_ID = "customer_user_id"
)
const GROUP_ACTIVITY_USERNAME = "group_user"
const FILTER_TYPE_USERS = "users"

// Source number to source name map
var SourceGroupUser = map[int]string{
	UserSourceSalesforce:      U.GROUP_NAME_SALESFORCE_ACCOUNT,
	UserSourceHubspot:         U.GROUP_NAME_HUBSPOT_COMPANY,
	UserSourceSixSignal:       U.GROUP_NAME_SIX_SIGNAL,
	UserSourceLinkedinCompany: U.GROUP_NAME_LINKEDIN_COMPANY,
	UserSourceDomains:         U.GROUP_NAME_DOMAINS,
	UserSourceG2:              U.GROUP_NAME_G2,
}

// source name to hostname
var HostNameGroup = map[string]string{
	U.GROUP_NAME_SALESFORCE_ACCOUNT: U.GP_SALESFORCE_ACCOUNT_WEBSITE,
	U.GROUP_NAME_HUBSPOT_COMPANY:    U.GP_HUBSPOT_COMPANY_DOMAIN,
	U.GROUP_NAME_SIX_SIGNAL:         U.SIX_SIGNAL_DOMAIN,
	U.GROUP_NAME_LINKEDIN_COMPANY:   U.LI_DOMAIN,
	U.GROUP_NAME_G2:                 U.G2_DOMAIN,
}

// source name to company name
var AccountNames = map[string]string{
	U.GROUP_NAME_SALESFORCE_ACCOUNT: U.GP_SALESFORCE_ACCOUNT_NAME,
	U.GROUP_NAME_HUBSPOT_COMPANY:    U.GP_HUBSPOT_COMPANY_NAME,
	U.GROUP_NAME_SIX_SIGNAL:         U.SIX_SIGNAL_NAME,
	U.GROUP_NAME_LINKEDIN_COMPANY:   U.LI_LOCALIZED_NAME,
	U.GROUP_NAME_G2:                 U.G2_NAME,
}

// host and company name list
var NameProps = []string{U.GP_HUBSPOT_COMPANY_NAME, U.GP_SALESFORCE_ACCOUNT_NAME, U.SIX_SIGNAL_NAME, U.LI_LOCALIZED_NAME, U.G2_NAME}
var HostNameProps = []string{U.GP_HUBSPOT_COMPANY_DOMAIN, U.GP_SALESFORCE_ACCOUNT_WEBSITE, U.SIX_SIGNAL_DOMAIN, U.LI_DOMAIN, U.G2_DOMAIN}

// Hover Events Property Map
var HOVER_EVENTS_NAME_PROPERTY_MAP = map[string][]string{
	U.EVENT_NAME_SESSION:                                         {U.EP_PAGE_COUNT, U.EP_CHANNEL, U.EP_CAMPAIGN, U.SP_SPENT_TIME, U.EP_REFERRER_URL},
	U.EVENT_NAME_FORM_SUBMITTED:                                  {U.EP_FORM_NAME, U.EP_PAGE_URL},
	U.EVENT_NAME_OFFLINE_TOUCH_POINT:                             {U.EP_CHANNEL, U.EP_CAMPAIGN},
	U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_CREATED:               {U.EP_SALESFORCE_CAMPAIGN_TYPE, U.EP_SALESFORCE_CAMPAIGNMEMBER_STATUS},
	U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_RESPONDED_TO_CAMPAIGN: {U.EP_SALESFORCE_CAMPAIGN_TYPE, U.EP_SALESFORCE_CAMPAIGNMEMBER_STATUS},
	U.EVENT_NAME_HUBSPOT_CONTACT_FORM_SUBMISSION:                 {U.EP_HUBSPOT_FORM_SUBMISSION_FORMTYPE, U.EP_HUBSPOT_FORM_SUBMISSION_PAGEURL},
	U.EVENT_NAME_HUBSPOT_ENGAGEMENT_EMAIL:                        {U.EP_HUBSPOT_ENGAGEMENT_SOURCE, U.EP_HUBSPOT_ENGAGEMENT_FROM},
	U.EVENT_NAME_HUBSPOT_ENGAGEMENT_MEETING_CREATED:              {U.EP_HUBSPOT_ENGAGEMENT_TYPE, U.EP_HUBSPOT_ENGAGEMENT_MEETINGOUTCOME, U.EP_HUBSPOT_ENGAGEMENT_STARTTIME, U.EP_HUBSPOT_ENGAGEMENT_ENDTIME},
	U.EVENT_NAME_HUBSPOT_ENGAGEMENT_CALL_CREATED:                 {U.EP_HUBSPOT_ENGAGEMENT_DURATIONMILLISECONDS, U.EP_HUBSPOT_ENGAGEMENT_STATUS},
	U.EVENT_NAME_SALESFORCE_TASK_CREATED:                         {U.EP_SF_TASK_TYPE, U.EP_SF_TASK_SUBTYPE, U.EP_SF_TASK_COMPLETED_DATETIME},
	U.EVENT_NAME_SALESFORCE_EVENT_CREATED:                        {U.EP_SF_EVENT_TYPE, U.EP_SF_EVENT_SUBTYPE, U.EP_SF_EVENT_COMPLETED_DATETIME},
}

// Page View Events Hover Properties
var PAGE_VIEW_HOVERPROPS_LIST = []string{U.EP_IS_PAGE_VIEW, U.EP_PAGE_SPENT_TIME, U.EP_PAGE_SCROLL_PERCENT}

var EVENT_ICONS_MAP = map[string]string{
	U.EVENT_NAME_SESSION:                            "brand",
	U.EVENT_NAME_HUBSPOT_ENGAGEMENT_EMAIL:           "envelope",
	U.EVENT_NAME_HUBSPOT_ENGAGEMENT_MEETING_CREATED: "handshake",
	U.EVENT_NAME_HUBSPOT_ENGAGEMENT_MEETING_UPDATED: "handshake",
	U.EVENT_NAME_HUBSPOT_ENGAGEMENT_CALL_CREATED:    "phone",
	U.EVENT_NAME_HUBSPOT_ENGAGEMENT_CALL_UPDATED:    "phone",
	U.EVENT_NAME_HUBSPOT_CONTACT_FORM_SUBMISSION:    "listcheck",
	U.EVENT_NAME_FORM_SUBMITTED:                     "hand-pointer",
}

var GROUP_TO_COMPANY_NAME_MAP = map[string]string{
	U.GROUP_NAME_HUBSPOT_COMPANY:    U.GP_HUBSPOT_COMPANY_NAME,
	U.GROUP_NAME_SALESFORCE_ACCOUNT: U.GP_SALESFORCE_ACCOUNT_NAME,
	U.GROUP_NAME_SIX_SIGNAL:         U.SIX_SIGNAL_DOMAIN,
	U.GROUP_NAME_LINKEDIN_COMPANY:   U.LI_LOCALIZED_NAME,
	U.GROUP_NAME_G2:                 U.G2_NAME,
}

func FormatTimeToString(time time.Time) string {
	return time.Format("2006-01-02 15:04:05.000000")
}

func IsDomainGroup(group string) bool {
	return group == "All" || group == U.GROUP_NAME_DOMAINS
}

func IsAnyProfiles(caller string) bool {
	return (caller == PROFILE_TYPE_USER || caller == PROFILE_TYPE_ACCOUNT)
}

func IsAccountProfiles(caller string) bool {
	return caller == PROFILE_TYPE_ACCOUNT
}

func IsUserProfiles(caller string) bool {
	return caller == PROFILE_TYPE_USER
}

var GroupPropertyPrefixList = []string{
	U.GROUP_NAME_HUBSPOT_COMPANY,
	U.GROUP_NAME_SALESFORCE_ACCOUNT,
	U.GROUP_NAME_SIX_SIGNAL,
	U.LI_PROPERTIES_PREFIX,
	U.GROUP_NAME_G2,
}

func UnixToLocalTime(timestamp int64) *time.Time {
	t := time.Unix(timestamp, 0)
	localTime := t.Local()
	return &localTime
}
