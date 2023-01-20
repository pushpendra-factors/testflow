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
	Name          string                 `json:"name,omitempty"`
	Company       string                 `json:"company"`
	Group1        bool                   `gorm:"default:false;column:group_1" json:"-"`
	Group2        bool                   `gorm:"default:false;column:group_2" json:"-"`
	Group3        bool                   `gorm:"default:false;column:group_3" json:"-"`
	Group4        bool                   `gorm:"default:false;column:group_4" json:"-"`
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
	Source    string          `json:"source"`
	Filters   []QueryProperty `json:"filters"`
	SegmentId string          `json:"segment_id"`
}

type AccountDetails struct {
	Properties      *postgres.Jsonb        `json:"-"`
	HostName        string                 `json:"host_name"`
	Name            string                 `json:"name"`
	LeftPaneProps   map[string]interface{} `json:"left_pane_props"`
	AccountTimeline []UserTimeline         `json:"account_timeline"`
}

type UserTimeline struct {
	UserId         string         `json:"-"`
	IsAnonymous    bool           `json:"is_anonymous"`
	UserName       string         `json:"user_name"`
	AdditionalProp string         `json:"additional_prop"`
	UserActivities []UserActivity `json:"user_activities,omitempty"`
}

// Constants
const PROFILE_TYPE_USER = "user"
const PROFILE_TYPE_ACCOUNT = "account"

// Hover Events Property Map
var HOVER_EVENTS_NAME_PROPERTY_MAP = map[string][]string{
	U.EVENT_NAME_SESSION:                            {U.EP_PAGE_COUNT, U.EP_CHANNEL, U.EP_CAMPAIGN, U.SP_SPENT_TIME, U.EP_TIMESTAMP, U.EP_REFERRER_URL},
	U.EVENT_NAME_FORM_SUBMITTED:                     {U.EP_FORM_NAME, U.EP_PAGE_URL, U.EP_TIMESTAMP},
	U.EVENT_NAME_OFFLINE_TOUCH_POINT:                {U.EP_CHANNEL, U.EP_CAMPAIGN, U.EP_TIMESTAMP},
	U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_CREATED:  {U.EP_SALESFORCE_CAMPAIGN_TYPE, U.EP_SALESFORCE_CAMPAIGNMEMBER_STATUS, U.EP_TIMESTAMP},
	U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_UPDATED:  {U.EP_SALESFORCE_CAMPAIGN_TYPE, U.EP_SALESFORCE_CAMPAIGNMEMBER_STATUS, U.EP_TIMESTAMP},
	U.EVENT_NAME_HUBSPOT_CONTACT_FORM_SUBMISSION:    {U.EP_HUBSPOT_FORM_SUBMISSION_FORMTYPE, U.EP_HUBSPOT_FORM_SUBMISSION_PAGEURL, U.EP_HUBSPOT_FORM_SUBMISSION_TIMESTAMP},
	U.EVENT_NAME_HUBSPOT_ENGAGEMENT_EMAIL:           {U.EP_HUBSPOT_ENGAGEMENT_SOURCE, U.EP_HUBSPOT_ENGAGEMENT_TIMESTAMP},
	U.EVENT_NAME_HUBSPOT_ENGAGEMENT_MEETING_CREATED: {U.EP_HUBSPOT_ENGAGEMENT_TYPE, U.EP_HUBSPOT_ENGAGEMENT_MEETINGOUTCOME, U.EP_HUBSPOT_ENGAGEMENT_STARTTIME, U.EP_HUBSPOT_ENGAGEMENT_ENDTIME},
	U.EVENT_NAME_HUBSPOT_ENGAGEMENT_MEETING_UPDATED: {U.EP_HUBSPOT_ENGAGEMENT_TYPE, U.EP_HUBSPOT_ENGAGEMENT_MEETINGOUTCOME, U.EP_HUBSPOT_ENGAGEMENT_STARTTIME, U.EP_HUBSPOT_ENGAGEMENT_ENDTIME},
	U.EVENT_NAME_HUBSPOT_ENGAGEMENT_CALL_CREATED:    {U.EP_HUBSPOT_ENGAGEMENT_DURATIONMILLISECONDS, U.EP_HUBSPOT_ENGAGEMENT_TIMESTAMP},
	U.EVENT_NAME_HUBSPOT_ENGAGEMENT_CALL_UPDATED:    {U.EP_HUBSPOT_ENGAGEMENT_DURATIONMILLISECONDS, U.EP_HUBSPOT_ENGAGEMENT_TIMESTAMP},
}

// Page View Events Hover Properties
var PAGE_VIEW_HOVERPROPS_LIST = []string{U.EP_IS_PAGE_VIEW, U.EP_PAGE_SPENT_TIME, U.EP_PAGE_SCROLL_PERCENT, U.EP_PAGE_LOAD_TIME}

var EVENT_ICONS_MAP = map[string]string{
	U.EVENT_NAME_SESSION:                            "brand",
	U.EVENT_NAME_HUBSPOT_ENGAGEMENT_EMAIL:           "envelope",
	U.EVENT_NAME_HUBSPOT_ENGAGEMENT_MEETING_CREATED: "HandshakeOutlined",
	U.EVENT_NAME_HUBSPOT_ENGAGEMENT_MEETING_UPDATED: "HandshakeOutlined",
	U.EVENT_NAME_HUBSPOT_ENGAGEMENT_CALL_CREATED:    "phone",
	U.EVENT_NAME_HUBSPOT_ENGAGEMENT_CALL_UPDATED:    "phone",
	U.EVENT_NAME_HUBSPOT_CONTACT_FORM_SUBMISSION:    "list_check",
	U.EVENT_NAME_FORM_SUBMITTED:                     "hand_pointer",
}
