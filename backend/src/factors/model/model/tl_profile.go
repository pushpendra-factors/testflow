package model

import (
	U "factors/util"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

type Profile struct {
	Identity     string                 `json:"identity"`
	Properties   *postgres.Jsonb        `json:"-"`
	Name         string                 `json:"name,omitempty"`
	HostName     string                 `json:"host_name,omitempty"`
	IsAnonymous  bool                   `json:"is_anonymous"`
	LastActivity time.Time              `json:"last_activity"`
	TableProps   map[string]interface{} `json:"table_props"`
}

type ContactDetails struct {
	UserId        string                 `json:"user_id"`
	IsAnonymous   bool                   `json:"is_anonymous"`
	Properties    *postgres.Jsonb        `json:"-"`
	LeftPaneProps map[string]interface{} `json:"left_pane_props"`
	Milestones    map[string]interface{} `json:"milestones"`
	Name          string                 `json:"name,omitempty"`
	Company       string                 `json:"company"`
	Group1        bool                   `gorm:"default:false;column:group_1" json:"-"`
	Group2        bool                   `gorm:"default:false;column:group_2" json:"-"`
	Group3        bool                   `gorm:"default:false;column:group_3" json:"-"`
	Group4        bool                   `gorm:"default:false;column:group_4" json:"-"`
	Group5        bool                   `gorm:"default:false;column:group_5" json:"-"`
	Group6        bool                   `gorm:"default:false;column:group_6" json:"-"`
	Group7        bool                   `gorm:"default:false;column:group_7" json:"-"`
	Group8        bool                   `gorm:"default:false;column:group_8" json:"-"`
	Group1UserID  string                 `gorm:"default:null;column:group_1_user_id" json:"-"`
	Group2UserID  string                 `gorm:"default:null;column:group_2_user_id" json:"-"`
	Group3UserID  string                 `gorm:"default:null;column:group_3_user_id" json:"-"`
	Group4UserID  string                 `gorm:"default:null;column:group_4_user_id" json:"-"`
	Group5UserID  string                 `gorm:"default:null;column:group_5_user_id" json:"-"`
	Group6UserID  string                 `gorm:"default:null;column:group_6_user_id" json:"-"`
	Group7UserID  string                 `gorm:"default:null;column:group_7_user_id" json:"-"`
	Group8UserID  string                 `gorm:"default:null;column:group_8_user_id" json:"-"`
	GroupInfos    []GroupsInfo           `json:"group_infos,omitempty"`
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
	Source       string                     `json:"source"`
	SegmentId    string                     `json:"segment_id"`
	Filters      map[string][]QueryProperty `json:"filters"`
	SearchFilter map[string][]QueryProperty `json:"search_filter"`
}

type AccountDetails struct {
	Properties      *postgres.Jsonb        `json:"-"`
	HostName        string                 `json:"host_name"`
	Name            string                 `json:"name"`
	LeftPaneProps   map[string]interface{} `json:"left_pane_props"`
	Milestones      map[string]interface{} `json:"milestones"`
	AccountTimeline []UserTimeline         `json:"account_timeline"`
}

type UserTimeline struct {
	UserId         string         `json:"user_id"`
	IsAnonymous    bool           `json:"is_anonymous"`
	UserName       string         `json:"user_name"`
	AdditionalProp string         `json:"additional_prop"`
	UserActivities []UserActivity `json:"user_activities"`
}

// Constants
const PROFILE_TYPE_USER = "user"
const PROFILE_TYPE_ACCOUNT = "account"

// Profile type for Segment Events
const USER_PROFILE_CALLER = "user_profiles"
const ACCOUNT_PROFILE_CALLER = "account_profiles"

// Source number to source name map
var SourceGroupUser = map[int]string{
	UserSourceSalesforce: U.GROUP_NAME_SALESFORCE_ACCOUNT,
	UserSourceHubspot:    U.GROUP_NAME_HUBSPOT_COMPANY,
	UserSourceSixSignal:  U.GROUP_NAME_SIX_SIGNAL,
	UserSourceDomains:    U.GROUP_NAME_DOMAINS,
}

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
}
