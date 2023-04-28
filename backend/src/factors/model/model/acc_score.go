package model

import (
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

type AccScoreResult struct {
	ProjectId int64             `json:"projectid"`
	AccResult []PerAccountScore `json:"accountResult"`
}
type AccUserScoreResult struct {
	ProjectId int64               `json:"projectid"`
	AccResult []PerUserScoreOnDay `json:"userResult"`
}

type UserScoreResult struct {
	ProjectId int64           `json:"projectid"`
	AccResult []AllUsersScore `json:"accountResult"`
}

type AccAggScore struct {
	AccountId    string    `json:"accountid"`
	ProjectId    int64     `json:"pprojectid"`
	CurrentScore int64     `json:"currentscore"`
	ScoreHistory []int64   `json:"scorehistory"`
	TimeStamp    time.Time `json:"time"`
}

type AccWeights struct {
	WeightConfig []AccEventWeight `json:"WeightConfig"`
	SaleWindow   int64            `json:"salewindow"`
}

type AccWeightsRequest struct {
	WeightConfig map[string]AccEventWeight `json:"WeightConfig"`
	SaleWindow   int64                     `json:"salewindow"`
}

type AccEventWeight struct {
	WeightId     string              `json:"wid"`
	Weight_value float32             `json:"weight"`
	Is_deleted   bool                `json:"is_deleted"`
	EventName    string              `json:"event_name"`
	Rule         WeightKeyValueTuple `json:"rule"`
}

type WeightKeyValueTuple struct {
	Key        string  `json:"key"`
	Value      string  `json:"value"`
	Operator   bool    `json:"operator"`
	LowerBound float64 `json:"lower_bound"`
	UpperBound float64 `json:"upper_bound"`
	Type       string  `json:"property_type"` //event or user property
	ValueType  string  `json:"value_type"`    //  category or numerical
}

type EventAgg struct {
	EventName  string `json:"eventname"`
	EventId    string `json:"eventid"`
	EventCount int64  `json:"eventcount"`
}

type AggEventsPerProj struct {
	EventsCount map[string]*EventAgg `json:"eventscount"`
}
type AggEventsPerProjpo struct {
	EventsCount map[string]EventAgg `json:"eventscount"`
}
type GroupEventsCount struct {
	ProjectID    int64           `gorm:"column:project_id; primary_key:true" json:"project_id"`
	AccountID    string          `gorm:"column:account_id; primary_key:true" json:"account_id"`
	GroupID      string          `gorm:"column:group_id; primary_key:true" json:"group_id"`
	AggCount     *postgres.Jsonb `gorm:"column:agg_counts" json:"aggcount"`
	DayTimestamp int64           `gorm:"column:day_timestamp" json:"day_timestamp"`
	CreatedAt    time.Time       `gorm:"column:created_at; autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time       `gorm:"column:updated_at; autoUpdateTime" json:"updated_at"`
}

type GroupEventsCountResult struct {
	ProjectID    int64              `gorm:"column:project_id; primary_key:true" json:"project_id"`
	AccountID    string             `gorm:"column:account_id; primary_key:true" json:"account_id"`
	GroupID      string             `gorm:"column:group_id; primary_key:true" json:"group_id"`
	AggCount     AggEventsPerProjpo `gorm:"column:agg_counts" json:"aggcount"`
	DayTimestamp int64              `gorm:"column:day_timestamp" json:"day_timestamp"`
	CreatedAt    time.Time          `gorm:"column:created_at; autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time          `gorm:"column:updated_at; autoUpdateTime" json:"updated_at"`
}

type PerAccountScore struct {
	Id        string         `json:"id"`
	Score     float32        `json:"score"`
	Timestamp string         `json:"timestamp"`
	Debug     map[string]int `json:"debug"`
}

type PerUserScoreOnDay struct {
	Id        string         `json:"id"`
	Score     float32        `json:"score"`
	Timestamp string         `json:"timestamp"`
	Debug     map[string]int `json:"debug"`
}

type AllUsersScore struct {
	UserId      string             `json:"UserId"`
	ScorePerDay map[string]float64 `json:"score"`
	Debug       map[string]int     `json:"debug"`
}

type GroupEventsCountScore struct {
	ProjectId int64             `json:"projectid"`
	AccScores []PerAccountScore `json:"accountscore"`
}

type EventsCountScore struct {
	UserId     string           `json:"uid"`
	ProjectId  int64            `json:"pid"`
	EventScore map[string]int64 `json:"eventscore"`
	DateStamp  string           `json:"ds"`
	IsGroup    bool             `json:"ig"`
}
type PerUserScore struct {
	DayScore map[string]PerDayScore `json:"dayscore"`
}

type PerDayScore struct {
	PerEventScore map[string]int `json:"dayscore"`
}
