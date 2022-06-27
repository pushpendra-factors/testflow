package model

import (
	"time"
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
	Name              string            `json:"name"`
	Company           string            `json:"company"`
	Role              string            `json:"role"`
	Email             string            `json:"email"`
	Country           string            `json:"country"`
	WebSessionsCount  uint32            `json:"web_sessions_count"`
	TimeSpentOnSite   uint32            `json:"time_spent_on_site"`
	NumberOfPageViews uint32            `json:"number_of_page_views"`
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
	EventName string `json:"event_name"`
	Timestamp uint64 `json:"timestamp"`
}
