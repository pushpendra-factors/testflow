package model

import (
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

// type ContactsList struct {
// 	Contacts []Contact `json:"contacts"`
// }

type Contact struct {
	// UserID         string `json:"user_id"`
	Identity     string    `json:"identity"`
	IsAnonymous  bool      `json:"is_anonymous"`
	Country      string    `json:"country"`
	LastActivity time.Time `json:"last_activity"`
}

type ContactDetails struct {
	UserId            string            `json:"user_id"`
	IsAnonymous       bool              `json:"is_anonymous"`
	Name              string            `json:"name"`
	Company           string            `json:"company"`
	Role              string            `json:"role"`
	Email             string            `json:"email"`
	Country           string            `json:"country"`
	WebSessionsCount  float64           `json:"web_sessions_count"`
	TimeSpentOnSite   float64           `json:"time_spent_on_site"`
	NumberOfPageViews float64           `json:"number_of_page_views"`
	Group1            bool              `gorm:"default:false;column:group_1" json:"group_1"`
	Group2            bool              `gorm:"default:false;column:group_2" json:"group_2"`
	Group3            bool              `gorm:"default:false;column:group_3" json:"group_3"`
	Group4            bool              `gorm:"default:false;column:group_4" json:"group_4"`
	GroupInfos        []GroupsInfo      `json:"group_infos"`
	UserActivity      []ContactActivity `json:"user_activities"`
}

type GroupsInfo struct {
	GroupName string `json:"group_name"`
}

type ContactActivity struct {
	EventName   string         `json:"event_name"`
	DisplayName string         `json:"display_name"`
	Properties  postgres.Jsonb `json:"properties"`
	Timestamp   uint64         `json:"timestamp"`
}

type UTListPayload struct {
	Source  string          `json:"source"`
	Filters []QueryProperty `json:"filters"`
}
