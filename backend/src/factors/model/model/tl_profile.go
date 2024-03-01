package model

import (
	"encoding/base64"
	U "factors/util"
	"regexp"
	"strings"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

type Profile struct {
	Identity       string                 `json:"identity"`
	Properties     *postgres.Jsonb        `json:"-"`
	Name           string                 `json:"name,omitempty"`
	HostName       string                 `json:"host_name,omitempty"`
	IsAnonymous    bool                   `json:"is_anonymous"`
	LastActivity   time.Time              `json:"last_activity"`
	TableProps     map[string]interface{} `json:"table_props"`
	Score          float64                `json:"score"`
	Engagement     string                 `json:"engagement,omitempty"`
	TopEngagements map[string]float64     `json:"top_engagements,omitempty"`
}

type ListingTimeWindow struct {
	LowerBound time.Time `json:"lower_bound"`
	UpperBound time.Time `json:"upper_bound"`
}

type ContactDetails struct {
	UserId        string                 `json:"user_id"`
	IsAnonymous   bool                   `json:"is_anonymous"`
	Properties    *postgres.Jsonb        `json:"-"`
	LeftPaneProps map[string]interface{} `json:"leftpane_props"`
	Milestones    map[string]interface{} `json:"milestones"`
	Name          string                 `json:"name"`
	Company       string                 `json:"company"`
	Account       string                 `json:"account"`
	UserActivity  []UserActivity         `json:"user_activities"`
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
	Query        Query    `json:"query"`
	SearchFilter []string `json:"search_filter"`
	SegmentId    string   `json:"segment_id"`
}

type AccountDetails struct {
	Properties      *postgres.Jsonb        `json:"-"`
	HostName        string                 `json:"host_name"`
	Name            string                 `json:"name"`
	LeftPaneProps   map[string]interface{} `json:"leftpane_props"`
	Milestones      map[string]interface{} `json:"milestones"`
	AccountTimeline []UserTimeline         `json:"account_timeline"`
}

type Overview struct {
	Temperature float64            `json:"temperature"` // Normalised Score for base 100
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
const USER_PROFILES = "user_profiles"
const ACCOUNT_PROFILES = "account_profiles"

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
var NameProps = []string{
	U.GP_HUBSPOT_COMPANY_NAME,
	U.GP_SALESFORCE_ACCOUNT_NAME,
	U.SIX_SIGNAL_NAME,
	U.LI_LOCALIZED_NAME,
	U.G2_NAME,
}
var HostNameProps = []string{
	U.DP_DOMAIN_NAME,
	U.GP_HUBSPOT_COMPANY_DOMAIN,
	U.GP_SALESFORCE_ACCOUNT_WEBSITE,
	U.SIX_SIGNAL_DOMAIN,
	U.LI_DOMAIN,
	U.G2_DOMAIN,
}

// Hover Events Property Map
var TIMELINE_EVENT_PROPERTIES_CONFIG = map[string][]string{
	U.EVENT_NAME_SESSION: {
		U.EP_CHANNEL,
		U.EP_PAGE_URL,
		U.EP_REFERRER_URL,
		U.EP_PAGE_COUNT,
		U.SP_SPENT_TIME,
	},
	U.EVENT_NAME_HUBSPOT_ENGAGEMENT_EMAIL: {
		U.EP_HUBSPOT_ENGAGEMENT_SUBJECT,
		U.EP_HUBSPOT_ENGAGEMENT_FROM,
		U.EP_HUBSPOT_ENGAGEMENT_TO,
	},
	U.EVENT_NAME_HUBSPOT_ENGAGEMENT_MEETING_CREATED: {
		U.EP_HUBSPOT_ENGAGEMENT_TITLE,
		U.EP_HUBSPOT_ENGAGEMENT_STARTTIME,
		U.EP_HUBSPOT_ENGAGEMENT_ENDTIME,
		U.EP_HUBSPOT_ENGAGEMENT_MEETINGOUTCOME,
	},
	U.EVENT_NAME_HUBSPOT_ENGAGEMENT_CALL_CREATED: {
		U.EP_HUBSPOT_ENGAGEMENT_TITLE,
		U.EP_HUBSPOT_ENGAGEMENT_DISPOSITION_LABEL,
		U.EP_HUBSPOT_ENGAGEMENT_DURATIONMILLISECONDS,
		U.EP_HUBSPOT_ENGAGEMENT_STATUS,
	},
	U.EVENT_NAME_HUBSPOT_CONTACT_UPDATED: {
		U.EP_HUBSPOT_CONTACT_EMAIL,
		U.EP_HUBSPOT_CONTACT_FIRSTNAME,
		U.EP_HUBSPOT_CONTACT_LASTNAME,
	},
	U.EVENT_NAME_HUBSPOT_CONTACT_LIST: {
		U.EP_HUBSPOT_CONTACT_LIST_LIST_NAME,
		U.EP_HUBSPOT_CONTACT_LIST_LIST_TYPE,
		U.EP_HUBSPOT_CONTACT_LIST_LIST_CREATED_TIMESTAMP,
		U.EP_HUBSPOT_CONTACT_LIST_CONTACT_EMAIL,
	},
	U.EVENT_NAME_HUBSPOT_CONTACT_FORM_SUBMISSION: {
		U.EP_HUBSPOT_FORM_SUBMISSION_TITLE,
		U.EP_HUBSPOT_FORM_SUBMISSION_PAGETITLE,
		U.EP_HUBSPOT_FORM_SUBMISSION_PAGEURL,
	},
	U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_CREATED: {
		U.EP_SALESFORCE_CAMPAIGN_NAME,
		U.EP_SALESFORCE_CAMPAIGN_TYPE,
		U.EP_SALESFORCE_CAMPAIGN_STATUS,
		U.EP_SALESFORCE_CAMPAIGNMEMBER_STATUS,
	},
	U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_RESPONDED_TO_CAMPAIGN: {
		U.EP_SALESFORCE_CAMPAIGN_NAME,
		U.EP_SALESFORCE_CAMPAIGN_TYPE,
		U.EP_SALESFORCE_CAMPAIGN_STATUS,
		U.EP_SALESFORCE_CAMPAIGNMEMBER_STATUS,
	},
	U.EVENT_NAME_SALESFORCE_CONTACT_CREATED: {
		U.EP_SALESFORCE_CONTACT_NAME,
		U.EP_SALESFORCE_CONTACT_EMAIL,
	},
	U.EVENT_NAME_SALESFORCE_LEAD_CREATED: {
		U.EP_SF_LEAD_NAME,
		U.EP_SF_LEAD_EMAIL,
	},
	U.EVENT_NAME_SALESFORCE_TASK_CREATED: {
		U.EP_SF_TASK_SUBJECT,
		U.EP_SF_TASK_TYPE,
		U.EP_SF_TASK_SUBTYPE,
		U.EP_SF_TASK_STATUS,
		U.EP_SF_TASK_DESCRIPTION,
	},
	U.EVENT_NAME_SALESFORCE_TASK_UPDATED: {
		U.EP_SF_TASK_SUBJECT,
		U.EP_SF_TASK_TYPE,
		U.EP_SF_TASK_SUBTYPE,
		U.EP_SF_TASK_STATUS,
		U.EP_SF_TASK_DESCRIPTION,
	},
	U.EVENT_NAME_SALESFORCE_EVENT_CREATED: {
		U.EP_SF_EVENT_SUBJECT,
		U.EP_SF_EVENT_TYPE,
		U.EP_SF_EVENT_SUBTYPE,
		U.EP_SF_TASK_DESCRIPTION,
	},
	U.EVENT_NAME_SALESFORCE_EVENT_UPDATED: {
		U.EP_SF_EVENT_SUBJECT,
		U.EP_SF_EVENT_TYPE,
		U.EP_SF_EVENT_SUBTYPE,
		U.EP_SF_TASK_DESCRIPTION,
	},
	U.EVENT_NAME_FORM_SUBMITTED: {
		U.EP_PAGE_TITLE,
		U.EP_PAGE_URL,
		U.EP_REFERRER_URL,
	},
	U.EVENT_NAME_OFFLINE_TOUCH_POINT: {
		U.EP_CHANNEL,
		U.EP_CAMPAIGN,
	},
	U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_CREATED: {
		U.GP_HUBSPOT_COMPANY_NAME,
		U.GP_HUBSPOT_COMPANY_DOMAIN,
	},
	U.GROUP_EVENT_NAME_SALESFORCE_ACCOUNT_CREATED: {
		U.GP_SALESFORCE_ACCOUNT_NAME,
		U.GP_SALESFORCE_ACCOUNT_TYPE,
	},
	U.GROUP_EVENT_NAME_SALESFORCE_OPPORTUNITY_CREATED: {
		U.GP_SALESFORCE_OPPORTUNITY_NAME,
		U.GP_SALESFORCE_OPPORTUNITY_STAGENAME,
		U.GP_SALESFORCE_OPPORTUNITY_TYPE,
	},
	U.GROUP_EVENT_NAME_G2_CATEGORY: {
		U.EP_PAGE_TITLE,
		U.EP_PAGE_URL,
		U.EP_G2_CATEGORY_IDS,
	},
	U.GROUP_EVENT_NAME_G2_PRICING: {
		U.EP_PAGE_TITLE,
		U.EP_PAGE_URL,
		U.EP_G2_PRODUCT_IDS,
	},
	U.GROUP_EVENT_NAME_G2_SPONSORED: {
		U.EP_PAGE_TITLE,
		U.EP_PAGE_URL,
		U.EP_G2_PRODUCT_IDS,
	},
	U.GROUP_EVENT_NAME_G2_DEAL: {
		U.EP_PAGE_TITLE,
		U.EP_PAGE_URL,
		U.EP_G2_PRODUCT_IDS,
	},
	U.GROUP_EVENT_NAME_G2_REFERENCE: {
		U.EP_PAGE_TITLE,
		U.EP_PAGE_URL,
		U.EP_G2_PRODUCT_IDS,
	},
	U.GROUP_EVENT_NAME_G2_COMPARISON: {
		U.EP_PAGE_TITLE,
		U.EP_PAGE_URL,
		U.EP_G2_PRODUCT_IDS,
	},
	U.GROUP_EVENT_NAME_G2_REPORT: {
		U.EP_PAGE_TITLE,
		U.EP_PAGE_URL,
		U.EP_G2_PRODUCT_IDS,
	},
	U.GROUP_EVENT_NAME_G2_ALTERNATIVE: {
		U.EP_PAGE_TITLE,
		U.EP_PAGE_URL,
		U.EP_G2_PRODUCT_IDS,
		U.EP_G2_CATEGORY_IDS,
	},
	U.GROUP_EVENT_NAME_G2_PRODUCT_PROFILE: {
		U.EP_PAGE_TITLE,
		U.EP_PAGE_URL,
		U.EP_G2_PRODUCT_IDS,
	},
	U.GROUP_EVENT_NAME_LINKEDIN_VIEWED_AD: {
		U.EP_CAMPAIGN,
		U.EP_CAMPAIGN_ID,
		U.LI_AD_VIEW_COUNT,
	},
	U.GROUP_EVENT_NAME_LINKEDIN_CLICKED_AD: {
		U.EP_CAMPAIGN,
		U.EP_CAMPAIGN_ID,
		U.LI_AD_CLICK_COUNT,
	},
	"PageView": {
		U.EP_PAGE_URL,
		U.EP_REFERRER_URL,
		U.EP_PAGE_SPENT_TIME,
		U.EP_PAGE_SCROLL_PERCENT,
	},
}

var STANDARD_EVENT_NAME_ALIASES = map[string]string{
	U.GROUP_EVENT_NAME_G2_CATEGORY:        "Looked at Product Category",
	U.GROUP_EVENT_NAME_G2_PRICING:         "Looked at Pricing",
	U.GROUP_EVENT_NAME_G2_SPONSORED:       "Saw Ad on Competitor's Page",
	U.GROUP_EVENT_NAME_G2_DEAL:            "Looked at G2 Deal",
	U.GROUP_EVENT_NAME_G2_REFERENCE:       "Looked at Reference Page",
	U.GROUP_EVENT_NAME_G2_COMPARISON:      "Compared with Other Products",
	U.GROUP_EVENT_NAME_G2_REPORT:          "Looked at Grid Report",
	U.GROUP_EVENT_NAME_G2_ALTERNATIVE:     "Looked at Alternatives",
	U.GROUP_EVENT_NAME_G2_PRODUCT_PROFILE: "Looked at Product Page",
}

var EVENT_ICONS_MAP = map[string]string{
	U.EVENT_NAME_SESSION:                            "globepointer",
	U.EVENT_NAME_HUBSPOT_ENGAGEMENT_EMAIL:           "envelope",
	U.EVENT_NAME_HUBSPOT_ENGAGEMENT_MEETING_CREATED: "handshake",
	U.EVENT_NAME_HUBSPOT_ENGAGEMENT_MEETING_UPDATED: "handshake",
	U.EVENT_NAME_HUBSPOT_ENGAGEMENT_CALL_CREATED:    "phone",
	U.EVENT_NAME_HUBSPOT_ENGAGEMENT_CALL_UPDATED:    "phone",
	U.EVENT_NAME_HUBSPOT_CONTACT_FORM_SUBMISSION:    "hubspot",
	U.EVENT_NAME_FORM_SUBMITTED:                     "clipboard",
}

var GROUP_TO_COMPANY_NAME_MAP = map[string]string{
	U.GROUP_NAME_HUBSPOT_COMPANY:    U.GP_HUBSPOT_COMPANY_NAME,
	U.GROUP_NAME_SALESFORCE_ACCOUNT: U.GP_SALESFORCE_ACCOUNT_NAME,
	U.GROUP_NAME_SIX_SIGNAL:         U.SIX_SIGNAL_DOMAIN,
	U.GROUP_NAME_LINKEDIN_COMPANY:   U.LI_LOCALIZED_NAME,
	U.GROUP_NAME_G2:                 U.G2_NAME,
}

type TimelinesConfig struct {
	DisabledEvents []string            `json:"disabled_events"`
	UserConfig     UserConfig          `json:"user_config"`
	AccountConfig  AccountConfig       `json:"account_config"`
	EventsConfig   map[string][]string `json:"events_config"`
}

// default timeline config
var DefaultTimelineConfig = TimelinesConfig{
	DisabledEvents: []string{"Contact Updated", "Campaign Member Updated", "Engagement Meeting Updated", "Engagement Call Updated"},
	UserConfig: UserConfig{
		Milestones:    []string{},
		TableProps:    []string{U.UP_COUNTRY, U.SP_SPENT_TIME},
		LeftpaneProps: []string{U.UP_EMAIL, U.UP_COUNTRY, U.UP_PAGE_COUNT},
	},
	AccountConfig: AccountConfig{
		Milestones:    []string{},
		TableProps:    []string{U.SIX_SIGNAL_NAME, U.SIX_SIGNAL_INDUSTRY, U.SIX_SIGNAL_EMPLOYEE_RANGE, U.SIX_SIGNAL_ANNUAL_REVENUE},
		LeftpaneProps: []string{},
		UserProp:      "",
	},
	EventsConfig: TIMELINE_EVENT_PROPERTIES_CONFIG,
}

func FormatTimeToString(time time.Time) string {
	return time.Format("2006-01-02 15:04:05.000000")
}

func IsDomainGroup(group string) bool {
	return group == U.GROUP_NAME_DOMAINS
}

func IsSourceAllUsers(source string) bool {
	return source == "All"
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
	U.GROUP_NAME_HUBSPOT_DEAL,
	U.GROUP_NAME_SALESFORCE_OPPORTUNITY,
}

func UnixToLocalTime(timestamp int64) *time.Time {
	t := time.Unix(timestamp, 0)
	localTime := t.Local()
	return &localTime
}

func ConvertDomainIdToHostName(domainID string) (string, error) {
	domomainIdEncoded := strings.TrimPrefix(domainID, "dom-")

	decodedBytes, err := base64.StdEncoding.DecodeString(domomainIdEncoded)
	if err != nil {
		return "", err
	}

	// Convert the decoded bytes to a string
	decodedString := string(decodedBytes)
	resultArray := strings.SplitN(decodedString, "-", 2)

	if len(resultArray) != 2 {
		return decodedString, nil
	}

	hostName := resultArray[1]
	return hostName, nil
}

func GetDomainFromURL(url string) string {
	re := regexp.MustCompile(`^(?:https?://)?(?:www\.)?([^:/\n?]+)`)
	match := re.FindStringSubmatch(url)

	if len(match) > 1 {
		return match[1]
	} else {
		return url
	}
}

var ExcludedEvents = []string{
	U.EVENT_NAME_HUBSPOT_CONTACT_UPDATED,
	U.EVENT_NAME_SALESFORCE_CONTACT_UPDATED,
	U.EVENT_NAME_SALESFORCE_LEAD_UPDATED,
	U.EVENT_NAME_LEAD_SQUARED_LEAD_UPDATED,
	U.EVENT_NAME_MARKETO_LEAD_UPDATED,
	U.EVENT_NAME_SALESFORCE_ACCOUNT_UPDATED,
	U.EVENT_NAME_SALESFORCE_OPPORTUNITY_UPDATED,
	U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_UPDATED,
	U.EVENT_NAME_SALESFORCE_TASK_UPDATED,
	U.EVENT_NAME_SALESFORCE_EVENT_UPDATED,
	U.EVENT_NAME_HUBSPOT_ENGAGEMENT_MEETING_UPDATED,
	U.EVENT_NAME_HUBSPOT_ENGAGEMENT_CALL_UPDATED,
	U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_UPDATED,
	U.GROUP_EVENT_NAME_HUBSPOT_DEAL_UPDATED,
	U.GROUP_EVENT_NAME_SALESFORCE_ACCOUNT_UPDATED,
	U.GROUP_EVENT_NAME_SALESFORCE_OPPORTUNITY_UPDATED,
	U.GROUP_EVENT_NAME_G2_ALL,
}

var ExcludedEventsBool = map[string]bool{
	U.EVENT_NAME_HUBSPOT_CONTACT_UPDATED:              true,
	U.EVENT_NAME_SALESFORCE_CONTACT_UPDATED:           true,
	U.EVENT_NAME_SALESFORCE_LEAD_UPDATED:              true,
	U.EVENT_NAME_LEAD_SQUARED_LEAD_UPDATED:            true,
	U.EVENT_NAME_MARKETO_LEAD_UPDATED:                 true,
	U.EVENT_NAME_SALESFORCE_ACCOUNT_UPDATED:           true,
	U.EVENT_NAME_SALESFORCE_OPPORTUNITY_UPDATED:       true,
	U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_UPDATED:    true,
	U.EVENT_NAME_SALESFORCE_TASK_UPDATED:              true,
	U.EVENT_NAME_SALESFORCE_EVENT_UPDATED:             true,
	U.EVENT_NAME_HUBSPOT_ENGAGEMENT_MEETING_UPDATED:   true,
	U.EVENT_NAME_HUBSPOT_ENGAGEMENT_CALL_UPDATED:      true,
	U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_UPDATED:        true,
	U.GROUP_EVENT_NAME_HUBSPOT_DEAL_UPDATED:           true,
	U.GROUP_EVENT_NAME_SALESFORCE_ACCOUNT_UPDATED:     true,
	U.GROUP_EVENT_NAME_SALESFORCE_OPPORTUNITY_UPDATED: true,
	U.GROUP_EVENT_NAME_G2_ALL:                         true,
}
