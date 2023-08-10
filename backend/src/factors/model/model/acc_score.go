package model

import (
	"math"
	"strconv"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

const DEFAULT_EVENT string = "all_events"
const LAST_EVENT string = "LAST_EVENT"
const NUM_TREND_DAYS int = 30

type AccScoreResult struct {
	ProjectId int64                  `json:"projectid"`
	AccResult []PerAccountScore      `json:"accountResult"`
	Debug     map[string]interface{} `json:"debug"`
}
type AccUserScoreResult struct {
	ProjectId int64                  `json:"projectid"`
	AccResult []PerUserScoreOnDay    `json:"userResult"`
	Debug     map[string]interface{} `json:"debug"`
}

type UserScoreResult struct {
	ProjectId int64                  `json:"projectid"`
	AccResult []AllUsersScore        `json:"accountResult"`
	Debug     map[string]interface{} `json:"debug"`
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

type AccEventWeight struct {
	WeightId     string                `json:"wid"`
	Weight_value float32               `json:"weight"`
	Is_deleted   bool                  `json:"is_deleted"`
	EventName    string                `json:"event_name"`
	Rule         []WeightKeyValueTuple `json:"rule"`
	Version      int                   `json:"vr"`
}

type WeightKeyValueTuple struct {
	Key        string   `json:"key"`
	Value      []string `json:"value"`
	Operator   string   `json:"operator"`
	LowerBound float64  `json:"lower_bound"`
	UpperBound float64  `json:"upper_bound"`
	Type       string   `json:"property_type"` //event or user property
	ValueType  string   `json:"value_type"`    //  category or numerical
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
	Id        string                 `json:"id"`
	Score     float32                `json:"score"`
	Timestamp string                 `json:"timestamp"`
	Trend     map[string]float32     `json:"trend"`
	Debug     map[string]interface{} `json:"debug"`
}

type PerUserScoreOnDay struct {
	Id        string                 `json:"id"`
	Score     float32                `json:"score"`
	Timestamp string                 `json:"timestamp"`
	Property  map[string][]string    `json:"prp"`
	Debug     map[string]interface{} `json:"debug"`
}

type AllUsersScore struct {
	UserId      string                       `json:"UserId"`
	ScorePerDay map[string]PerUserScoreOnDay `json:"score"`
	Debug       map[string]interface{}       `json:"debug"`
}

type GroupEventsCountScore struct {
	ProjectId int64             `json:"projectid"`
	AccScores []PerAccountScore `json:"accountscore"`
}

type EventsCountScore struct {
	UserId     string                      `json:"uid"`
	ProjectId  int64                       `json:"pid"`
	EventScore map[string]int64            `json:"eventscore"`
	Property   map[string]map[string]int64 `json:"prp"`
	DateStamp  string                      `json:"ds"`
	IsGroup    bool                        `json:"ig"`
}
type PerUserScore struct {
	DayScore map[string]PerDayScore `json:"dayscore"`
}

type PerDayScore struct {
	PerEventScore map[string]int `json:"dayscore"`
}

type LatestScore struct {
	Date        int64                       `json:"date"`
	EventsCount map[string]float64          `json:"events"`
	Properties  map[string]map[string]int64 `json:"prop"`
}

type DbUpdateAccScoring struct {
	Userid         string      `json:"uid"`
	TS             int64       `json:"ts"`
	Date           string      `json:"date"`
	CurrEventCount LatestScore `json:"ev"`
	Lastevent      LatestScore `json:"lev"`
	IsGroup        bool        `json:"ig"`
}

func GetDateFromString(ts string) int64 {
	year := ts[0:4]
	month := ts[4:6]
	date := ts[6:8]
	t, _ := strconv.ParseInt(year, 10, 64)
	t1, _ := strconv.ParseInt(month, 10, 64)
	t2, _ := strconv.ParseInt(date, 10, 64)
	t3 := time.Date(int(t), time.Month(t1), int(t2), 0, 0, 0, 0, time.UTC).Unix()
	return t3
}

func GetDefaultAccScoringWeights() AccWeights {
	var weights AccWeights
	var event_a AccEventWeight
	var event_b AccEventWeight
	var event_c AccEventWeight

	weights.WeightConfig = make([]AccEventWeight, 0)
	weights.SaleWindow = 10

	event_a = AccEventWeight{EventName: "$session", Weight_value: 10, Is_deleted: false,
		Version: 1}

	keyvals := []string{"Paid search"}
	filterPRoperties := WeightKeyValueTuple{Key: "$channel", Value: keyvals, Operator: EqualsOpStr,
		LowerBound: 0, UpperBound: 0, Type: "event", ValueType: "categorical"}
	event_b = AccEventWeight{EventName: "$session", Weight_value: 20, Is_deleted: false,
		Rule: []WeightKeyValueTuple{filterPRoperties}, Version: 1}

	event_c = AccEventWeight{EventName: "$form_submitted", Weight_value: 40, Is_deleted: false,
		Version: 1}

	weights.WeightConfig = append(weights.WeightConfig, event_a)
	weights.WeightConfig = append(weights.WeightConfig, event_b)
	weights.WeightConfig = append(weights.WeightConfig, event_c)

	return weights
}

func ComputeDayDifference(ts1 int64, ts2 int64) int {

	t1 := time.Unix(ts1, 0)
	t2 := time.Unix(ts2, 0)
	daydiff := t2.Sub(t1).Seconds() / float64(24*60*60)
	return int(math.Abs(daydiff))

}

func ComputeDecayValue(ts string, SaleWindow int64) float64 {
	currentTS := time.Now().Unix()
	EventTs := GetDateFromString(ts)
	decay := ComputeDecayValueGivenStartEndTS(currentTS, EventTs, SaleWindow)

	return decay
}

func ComputeDecayValueGivenStartEndTS(start int64, end int64, SaleWindow int64) float64 {
	var decay float64
	// get difference in weeks
	dayDiff := ComputeDayDifference(start, end)
	if int64(dayDiff) > SaleWindow {
		return 0
	}
	// get decay value
	decay = 1 - float64(float64(int64(dayDiff))/float64(SaleWindow))

	return decay
}
