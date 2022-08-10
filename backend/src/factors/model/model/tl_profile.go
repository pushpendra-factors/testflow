package model

import (
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

type Profile struct {
	Identity           string    `json:"identity"`
	Name               string    `json:"name"`
	IsAnonymous        bool      `json:"is_anonymous"`
	Country            string    `json:"country"`
	AssociatedContacts uint64    `json:"associated_contacts"`
	LastActivity       time.Time `json:"last_activity"`
}

type ContactDetails struct {
	UserId            string         `json:"user_id"`
	IsAnonymous       bool           `json:"is_anonymous"`
	Name              string         `json:"name"`
	Company           string         `json:"company"`
	Role              string         `json:"role"`
	Email             string         `json:"email"`
	Country           string         `json:"country"`
	WebSessionsCount  float64        `json:"web_sessions_count"`
	TimeSpentOnSite   float64        `json:"time_spent_on_site"`
	NumberOfPageViews float64        `json:"number_of_page_views"`
	Group1            bool           `gorm:"default:false;column:group_1" json:"group_1"`
	Group2            bool           `gorm:"default:false;column:group_2" json:"group_2"`
	Group3            bool           `gorm:"default:false;column:group_3" json:"group_3"`
	Group4            bool           `gorm:"default:false;column:group_4" json:"group_4"`
	GroupInfos        []GroupsInfo   `json:"group_infos,omitempty"`
	UserActivity      []UserActivity `json:"user_activities,omitempty"`
}

type GroupsInfo struct {
	GroupName string `json:"group_name"`
}

type UserActivity struct {
	EventName   string         `json:"event_name"`
	DisplayName string         `json:"display_name"`
	Properties  postgres.Jsonb `json:"properties,omitempty"`
	Timestamp   uint64         `json:"timestamp"`
}

type TimelinePayload struct {
	Source  string          `json:"source"`
	Filters []QueryProperty `json:"filters"`
}

type AccountDetails struct {
	Name              string         `json:"name"`
	Industry          string         `json:"industry"`
	Country           string         `json:"country"`
	NumberOfEmployees uint64         `json:"number_of_employees"`
	NumberOfUsers     uint64         `json:"number_of_users"`
	AccountTimeline   []UserTimeline `json:"account_timeline"`
}

type UserTimeline struct {
	UserId         string         `json:"user_id"`
	UserName       string         `json:"user_name"`
	UserActivities []UserActivity `json:"user_activities,omitempty"`
}

// Constants
const PROFILE_TYPE_USER = "user"
const PROFILE_TYPE_ACCOUNT = "account"
