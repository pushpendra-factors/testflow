package model

import (
	"encoding/json"
	U "factors/util"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

const DEFAULT_EVENT string = "all_events"
const LAST_EVENT string = "LAST_EVENT"
const NUM_TREND_DAYS int = 30

const (
	ENGAGEMENT_LEVEL_HOT  = "Hot"
	ENGAGEMENT_LEVEL_WARM = "Warm"
	ENGAGEMENT_LEVEL_COOL = "Cool"
	ENGAGEMENT_LEVEL_ICE  = "Ice"
)

var BUCKETRANGES []float64 = []float64{100, 90, 70, 30, 0}
var BUCKETNAMES []string = []string{ENGAGEMENT_LEVEL_HOT, ENGAGEMENT_LEVEL_WARM, ENGAGEMENT_LEVEL_COOL, ENGAGEMENT_LEVEL_ICE}

const HASHKEYLENGTH = 12

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
	FilterName   string                `json:"fname"`
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
	TopEvents map[string]float64     `json:"tpe"`
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
	TopEvents   map[string]float64
}

type DbUpdateAccScoring struct {
	Userid         string      `json:"uid"`
	TS             int64       `json:"ts"`
	Date           string      `json:"date"`
	CurrEventCount LatestScore `json:"ev"`
	Lastevent      LatestScore `json:"lev"`
	IsGroup        bool        `json:"ig"`
}

type Bucket struct {
	Name string  `json:"nm"`
	High float64 `json:"high"`
	Low  float64 `json:"low"`
}

type BucketRanges struct {
	Date   string   `json:"date"`
	Ranges []Bucket `json:"bck"`
}

type AccountScoreRanges struct {
	ProjectID int64           `gorm:"column:project_id; primary_key:true" json:"project_id"`
	Date      string          `gorm:"column:date" json:"date"`
	Buckets   *postgres.Jsonb `gorm:"column:bucket" json:"bucket"`
	CreatedAt time.Time       `gorm:"column:created_at; autoCreateTime" json:"created_at"`
	UpdatedAt time.Time       `gorm:"column:updated_at; autoUpdateTime" json:"updated_at"`
}

type Weightfilter struct {
	Ename string                `json:"en"`
	Rule  []WeightKeyValueTuple `json:"rule"`
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
	log.Info("adding default account scoring weights")

	var weights AccWeights
	var event_a AccEventWeight
	var event_b AccEventWeight
	var event_c AccEventWeight

	weights.WeightConfig = make([]AccEventWeight, 0)
	weights.SaleWindow = 10
	default_version := 0

	event_a = AccEventWeight{FilterName: "Website Session 1", EventName: "$session", Weight_value: 10, Is_deleted: false,
		Version: 1}

	keyvals := []string{"Paid search"}
	filterPRoperties := WeightKeyValueTuple{Key: "$channel", Value: keyvals, Operator: EqualsOpStr,
		LowerBound: 0, UpperBound: 0, Type: "event", ValueType: "categorical"}

	event_b = AccEventWeight{FilterName: "Website Session 2", EventName: "$session", Weight_value: 20, Is_deleted: false,
		Rule: []WeightKeyValueTuple{filterPRoperties}, Version: 1}

	event_c = AccEventWeight{FilterName: "Form Submitted", EventName: "$form_submitted", Weight_value: 40, Is_deleted: false,
		Version: 1}
	keyvals = []string{"true"}
	filterPRoperties = WeightKeyValueTuple{Key: "$is_page_view",
		Value:      keyvals,
		Operator:   EqualsOpStr,
		LowerBound: 0, UpperBound: 0,
		Type:      "event",
		ValueType: "categorical"}

	event_d := AccEventWeight{FilterName: "All events", EventName: "all_events", Weight_value: 2, Is_deleted: false,
		Rule: []WeightKeyValueTuple{filterPRoperties}, Version: default_version}

	event_e := AccEventWeight{FilterName: "Linkedin ad clicked", EventName: "$linkedin_clicked_ad", Weight_value: 2, Is_deleted: false, Version: default_version}
	event_f := AccEventWeight{FilterName: "Linkedin ad viewed", EventName: "$linkedin_viewed_ad", Weight_value: 1, Is_deleted: false, Version: default_version}
	event_g := AccEventWeight{FilterName: "Form fill", EventName: "$form_fill", Weight_value: 20, Is_deleted: false, Version: default_version}

	weights.WeightConfig = append(weights.WeightConfig, event_a)
	weights.WeightConfig = append(weights.WeightConfig, event_b)
	weights.WeightConfig = append(weights.WeightConfig, event_c)
	weights.WeightConfig = append(weights.WeightConfig, event_d)
	weights.WeightConfig = append(weights.WeightConfig, event_e)
	weights.WeightConfig = append(weights.WeightConfig, event_f)
	weights.WeightConfig = append(weights.WeightConfig, event_g)
	log.WithField("weight", weights).Debugf("weights")
	return weights
}

func ComputeDecayValue(ts string, SaleWindow int64) float64 {
	currentTS := time.Now().Unix()
	EventTs := U.GetDateFromString(ts)
	decay := ComputeDecayValueGivenStartEndTS(currentTS, EventTs, SaleWindow)

	return decay
}

func ComputeDecayValueGivenStartEndTS(start int64, end int64, SaleWindow int64) float64 {
	var decay float64
	// get difference in weeks
	dayDiff := U.ComputeDayDifference(start, end)
	if int64(dayDiff) > SaleWindow {
		return 0
	}
	// get decay value
	decay = 1 - float64(float64(int64(dayDiff))/float64(SaleWindow))

	return decay
}

func removeZeros(input []float64) []float64 {
	var result []float64

	for _, value := range input {
		if value != 0 {
			result = append(result, value)
		}
	}
	return result
}

func getUniqueScores(input []float64) []float64 {
	uniqueMap := make(map[float64]bool)
	result := []float64{}

	for _, item := range input {
		if _, found := uniqueMap[item]; !found {
			uniqueMap[item] = true
			result = append(result, item)
		}
	}

	return result
}

func GetEngagementLevels(scores []float64, buckets BucketRanges) map[float64]string {
	result := make(map[float64]string)
	result[0] = GetEngagement(0, buckets)

	nonZeroScores := removeZeros(scores)
	uniqueScores := getUniqueScores(nonZeroScores)

	for _, score := range uniqueScores {
		// calculating percentile is not used in the current implementation
		// based on the config and the percentile ranges , in the account scoring job
		// we calculate the boundaries and store it in account_scoring_ranges table.
		// the accounts listing page will use these to show the engagement level.
		result[score] = GetEngagement(score, buckets)
	}
	return result
}

func GetEngagement(percentile float64, buckets BucketRanges) string {

	var maxHigh float64
	var maxLow float64
	maxHigh = float64(0)
	maxLow = math.MaxFloat64

	for _, bucket := range buckets.Ranges {

		if bucket.High > maxHigh {
			maxHigh = bucket.High
		}
		if bucket.Low < maxLow {
			maxLow = bucket.Low
		}

		if bucket.Low <= percentile && percentile <= bucket.High {

			if strings.Compare(bucket.Name, ENGAGEMENT_LEVEL_COOL) == 0 || strings.Compare(bucket.Name, "Cold") == 0 {
				return ENGAGEMENT_LEVEL_COOL
			}
			return bucket.Name
		}
	}

	if percentile > maxHigh {
		return "Hot"
	} else if percentile < maxLow {
		return "Ice"
	}
	return "Ice"
}

func DeduplicateWeights(weights AccWeights) (AccWeights, error) {

	var updatedWeights AccWeights
	var weightrulelist []AccEventWeight = make([]AccEventWeight, 0)
	weightIdMap := make(map[string]AccEventWeight)
	wt := weights.WeightConfig
	var err error
	for _, wid := range wt {
		var w_id_hash string
		// if event name is empty, set it to all_events
		// filter will be applied to all events
		if wid.EventName == "" {
			wid.EventName = DEFAULT_EVENT
		}
		w_val := Weightfilter{wid.EventName, wid.Rule}
		w_id_hash = wid.WeightId
		if w_id_hash == "" {
			w_id_hash, err = GenerateHashFromStruct(w_val)
			if err != nil {
				return AccWeights{}, err
			}
			log.Debugf("generated hash :%s", w_id_hash)
		}

		if _, ok := weightIdMap[w_id_hash]; !ok {
			wid.WeightId = w_id_hash
			weightIdMap[w_id_hash] = wid
			weightrulelist = append(weightrulelist, wid)
		} else {
			p := weightIdMap[w_id_hash]
			p.Is_deleted = wid.Is_deleted
			log.Debugf("Duplicate detected, not adding to list:%s", w_id_hash)
		}
	}

	updatedWeights.WeightConfig = weightrulelist
	updatedWeights.SaleWindow = weights.SaleWindow
	return updatedWeights, nil
}

func GenerateHashFromStruct(w Weightfilter) (string, error) {
	queryCacheBytes, err := json.Marshal(w)
	if err != nil {
		return "", err
	}

	hstringLowerCase := strings.ToLower(string(queryCacheBytes))
	hstring, err := GenerateHash(hstringLowerCase)
	if err != nil {
		return "", err
	}
	return hstring, err

}
func GenerateHash(bytes string) (string, error) {
	uuid := U.HashKeyUsingSha256Checksum(string(bytes))[:HASHKEYLENGTH]
	return uuid, nil
}

func UpdateWeights(projectId int64, weightsRequest AccWeights, requestID string) (AccWeights, int, error) {
	log.Info("updating  weights in DB")

	logCtx := log.WithFields(log.Fields{
		"projectId": projectId,
		"RequestId": requestID,
	})

	var weights AccWeights
	weights.WeightConfig = make([]AccEventWeight, 0)
	filterNamesMap := make(map[string]int)
	weights.SaleWindow = weightsRequest.SaleWindow

	// check for duplicate rule names first , in case of empty rule name
	// fill it with the event names
	for _, wtVal := range weightsRequest.WeightConfig {
		if len(wtVal.FilterName) > 0 {
			if _, wtOk := filterNamesMap[wtVal.FilterName]; wtOk {
				errMsg := fmt.Errorf("duplicate rule name detected")
				logCtx.WithField("duplicate name : ", wtVal.FilterName).Error(errMsg)
				return AccWeights{}, http.StatusBadRequest, errMsg
			} else {
				filterNamesMap[wtVal.FilterName] = 1
			}
		}
	}

	// convert incoming request to AccWeights.
	for _, wtVal := range weightsRequest.WeightConfig {
		var r AccEventWeight

		r.EventName = wtVal.EventName
		r.Is_deleted = wtVal.Is_deleted
		r.Rule = wtVal.Rule
		r.WeightId = wtVal.WeightId
		r.Weight_value = wtVal.Weight_value
		r.FilterName = wtVal.FilterName
		if len(r.FilterName) == 0 && len(r.EventName) > 0 {
			r.FilterName = r.EventName
		}
		weights.WeightConfig = append(weights.WeightConfig, r)
	}

	dedupweights, err := DeduplicateWeights(weights)
	if err != nil {
		errMsg := fmt.Errorf("Unable to dedup weights.")
		logCtx.WithError(err).Error(errMsg)
		return AccWeights{}, http.StatusBadRequest, errMsg
	}
	logCtx.Info("Done deduplicating weights ")
	return dedupweights, http.StatusAccepted, nil
}

func DefaultEngagementBuckets() BucketRanges {
	var defaultBucket BucketRanges
	defaultBucket.Ranges = make([]Bucket, 4)

	hotBucket := Bucket{Name: ENGAGEMENT_LEVEL_HOT, High: 100, Low: 90}
	warmBucket := Bucket{Name: ENGAGEMENT_LEVEL_WARM, High: 90, Low: 70}
	coldBucket := Bucket{Name: ENGAGEMENT_LEVEL_COOL, High: 70, Low: 30}
	iceBucket := Bucket{Name: ENGAGEMENT_LEVEL_ICE, High: 30, Low: 0}

	timeNow := time.Now().Unix()
	dateToday := U.GetDateOnlyFromTimestamp(timeNow)
	defaultBucket.Date = dateToday
	defaultBucket.Ranges = []Bucket{hotBucket, warmBucket, coldBucket, iceBucket}

	return defaultBucket

}
