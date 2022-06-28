package delta

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	P "factors/pattern"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

var TARGET_BEFORE_BASE_ERROR error = errors.New("LOGICAL ERROR: User hits target but misses base. It's recommended to change your base criteria")

// This is an enumerated list with values And or Or.
type BooleanOperator string

type Fraction struct {
	Numerator   float64
	Denominator float64
}

type MetricInfo struct {
	Global   float64
	Features map[string]map[string]float64
}

// const (
// 	And BooleanOperator = iota
// 	Or
// )

// var toID = map[string]BooleanOperator{
// 	"And": And,
// 	"Or":  Or,
// }

// var toString = map[BooleanOperator]string{
// 	And: "And",
// 	Or:  "Or",
// }

// Values allowed = "user", "event" for "userProperty" and "eventProperty" respectively.
type PropertiesMode string

// const (
// 	Up PropertiesMode = iota
// 	Ep
// )

// var toOpID = map[string]PropertiesMode{
// 	"Up": Up,
// 	"Ep": Ep,
// }

// var toOpString = map[PropertiesMode]string{
// 	Up: "Up",
// 	Ep: "Ep",
// }

// EventFilterCriterion represents a filter-criterion that an event has to follow.
// It is composed of a feature key, its allowed values, and an equality flag.
// If equality is True, it translates to "key in values", otherwise, "key not in values".
// In other words, if equality flag is True, then key "can take any of the values" (OR),
// and if it is False, key "cannot take any of the values" (AND).
type EventFilterCriterion struct {
	Id             int    `json:"id"`
	Key            string `json:"key"`
	Type           string
	Values         []OperatorValueTuple
	PropertiesMode PropertiesMode `json:"propmode"`
}

type OperatorValueTuple struct {
	Operator  string
	Value     string
	LogicalOp string
}

// EventCriterion abstracts a single event-criterion that users/events have to adhere to.
// More specifically, it contains an event name, equality flag, and a list of filter-criterion.
type EventCriterion struct {
	Id                  int                    `json:"id"`
	Name                string                 `json:"en"`
	EqualityFlag        bool                   `json:"eq"`
	FilterCriterionList []EventFilterCriterion `json:"filters"`
}

// EventsCriteria represents one period's event-criteria that users/events have to adhere to.
// It is basically an AND or an OR of multiple event-criterion.
type EventsCriteria struct {
	Id                 int              `json:"id"`
	Operator           BooleanOperator  `json:"op"`
	EventCriterionList []EventCriterion `json:"events"`
}

// type EventCriterionMatchResult struct {
// 	event P.CounterEventFormat
// 	op    BooleanOperator
// }

// type DeltaQueryMatchResult struct {
// 	userBaseMatchFlag   bool
// 	userTargetMatchFlag bool
// 	featDist            FeatureDistribution
// }

// type PerUserQueryResult struct {
// 	baseFlag   bool
// 	targetFlag bool
// }

// type FeatureDistribution map[Feature]int64

// type EventsCriteriaMatchResult struct {
// 	AtLeastOneFlag bool
// 	AllFlag        bool
// 	latestEvent    P.CounterEventFormat
// }

type PerUserQueryResult struct {
	baseResult   PerUserCriteriaResult
	targetResult PerUserCriteriaResult
}

type PerUserQueryResultSummary struct {
	BaseFlag          bool
	TargetFlag        bool
	BaseAndTargetFlag bool
	ChosenEvent       P.CounterEventFormat
}

type PerEventProperties struct {
	BaseFlag          bool
	TargetFlag        bool
	BaseAndTargetFlag bool
	EventProperties   map[string]interface{} `json:"epr"`
	UserProperties    map[string]interface{} `json:"upr"`
}

type PerUserCriteriaResult struct {
	criterionResultList []CriterionResult
	allFlag             bool
	anyFlag             bool
	mostRecentEvent     P.CounterEventFormat
	firstEvent          P.CounterEventFormat
	numCriterionMatched int
	criteriaMatchFlag   bool
}

type CriterionResult struct {
	matchId int
}

// Query abstracts the concept of a delta-query, consisting of one base, and a list of target criteria.
type Query struct {
	Id     int            `json:"id"`
	Base   EventsCriteria `json:"base"`
	Target EventsCriteria `json:"target"`
}

type MultiFunnelQuery struct {
	Id           int              `json:"id"`
	Base         EventsCriteria   `json:"base"`
	Intermediate []EventsCriteria `json:"intermediate"`
	Target       EventsCriteria   `json:"target"`
}

// ValCountTable is a map where feature-values are keys and their frequency counts are values.
type ValCountTable map[string]uint64

// KeyValCountTable is a map where feature-keys are keys and their feature-value-counts are values.
type KeyValCountTable map[string]ValCountTable

// QueryKeyValCountTable is a map where query ids are keys and their feature-key-value-frequency-counts are values.
type QueryKeyValCountTable map[string]KeyValCountTable

// QueryResultsTable is a map where query ids are keys and their feature-key-value-frequency-counts are values.
type QueryResultsTable QueryKeyValCountTable

// MetricQueryKeyValCountTable maps a delta-metric to query-key-val-counts.
type MetricQueryKeyValCountTable map[string]QueryKeyValCountTable

// DeltaDataType maps a delta-metric to query results.
type DeltaDataType map[string]QueryResultsTable

// TimeInterval is used to capture a start-time (from) and end-time (to) based time interval.
type TimeInterval struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

// TODO: Change this to TimeInterval later.
type Period struct {
	From int64
	To   int64
}

// PeriodPair defines two periods for which delta insights could be computed.
type PeriodPair struct {
	First  Period `json:"first"`
	Second Period `json:"second"`
}

// Feature defines a feature_key-feature_value pair.
type Feature struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type metricFormulaInputs map[string]float64

type metricFormulaOutput float64

type metricFormula struct {
	method func(metricFormulaInputs) metricFormulaOutput
}

type metricExplanation struct {
	description string
}

type Metric struct {
	Name  string              `json:"name"`
	Value metricFormulaOutput `json:"value"`

	explanation metricExplanation
	formula     metricFormula
	inputs      metricFormulaInputs
}

// CrossInsight - a period-over-period insight.
type CrossInsight struct {
	Feature Feature           `json:"feature"`
	Metrics map[string]Metric `json:"metrics"`
}

// WithinInsight - a single-period insight.
type WithinInsight struct {
	Feature Feature           `json:"feature"`
	Metrics map[string]Metric `json:"metrics"`
}

// WithinPeriodInsights stores insights per period (say a week or a month).
type WithinPeriodInsights struct {
	Period        Period                   `json:"period"`
	Base          WithinPeriodMetrics      `json:"base"`
	Target        WithinPeriodMetrics      `json:"target"`
	BaseAndTarget WithinPeriodMetrics      `json:"base_target"`
	Conversion    WithinPeriodRatioMetrics `json:"conv"`
	Prevalence    WithinPeriodRatioMetrics `json:"prev"`
}

type WithinPeriodMetrics struct {
	GlobalMetrics  Level1CatFreqDist  `json:"global"`
	FeatureMetrics Level3CatRatioDist `json:"feat"`
}

type WithinPeriodRatioMetrics struct {
	GlobalMetrics  Level1CatRatioDist `json:"global"`
	FeatureMetrics Level2CatRatioDist `json:"feat"`
}

type Level1CatFreqDist map[string]int
type Level2CatFreqDist map[string]Level1CatFreqDist
type Level3CatFreqDist map[string]Level2CatFreqDist

type Level1CatRatioDist map[string]float64
type Level2CatRatioDist map[string]Level1CatRatioDist
type Level3CatRatioDist map[string]Level2CatRatioDist

type JSDType struct {
	Base   Level2CatRatioDist `json:"base"`
	Target Level2CatRatioDist `json:"target"`
}

// CrossPeriodInsights stores cross-insights per period-pair.
type CrossPeriodInsights struct {
	Periods       PeriodPair         `json:"periods"`
	Base          CrossPeriodMetrics `json:"base"`
	Target        CrossPeriodMetrics `json:"target"`
	BaseAndTarget CrossPeriodMetrics `json:"base_target"`
	Conversion    CrossPeriodMetrics `json:"conv"`
	DeltaRatio    Level2CatRatioDist `json:"delrat"`
	JSDivergence  JSDType            `json:"jsd"`
}

type CrossPeriodMetrics struct {
	GlobalMetrics  Level1CatDiffDist `json:"global"`
	FeatureMetrics Level2CatDiffDist `json:"feat"`
}

type DiffMetric struct {
	First         interface{} `json:"first"`
	Second        interface{} `json:"second"`
	PercentChange float64     `json:"perc"`
	FactorChange  float64     `json:"factor"`
}

type WithinPeriodInsightsKpi struct {
	Category   string      `json:"categ"`
	MetricInfo *MetricInfo `json:"metric"`
	ScaleInfo  *MetricInfo `json:"scale"`
}

type CrossPeriodInsightsKpi struct {
	Periods       PeriodPair     `json:"periods"`
	Target        *CpiMetricInfo `json:"target"`
	BaseAndTarget *CpiMetricInfo `json:"base_target"`
	ScaleInfo     *CpiMetricInfo `json:"scale"`
	JSDivergence  JSDType        `json:"jsd"`
	Category      string         `json:"categ"`
}

type CpiMetricInfo struct {
	GlobalMetrics  DiffMetric                       `json:"global"`
	FeatureMetrics map[string]map[string]DiffMetric `json:"feat"`
}

type Level1CatDiffDist map[string]DiffMetric
type Level2CatDiffDist map[string]Level1CatDiffDist

// Insight represents a single insight.
// type Insight struct {
// 	Feature           Feature
// 	individualMetrics map[string]metric
// 	differenceMetrics map[string]metric
// }

// GetPreviousPeriodModelId automatically identifies the model ID of the previous period.
func GetPreviousPeriodModelId(projectId, modelId uint64) uint64 {
	// TODO: First find the period of modelId, then find the previous period.
	// For example, if modelId corresponds to 8-14 Jan, 2021, then identify
	// the period as a 'week', and then fetch the model ID for the 'previous
	// week', i.e., 1-7 Jan, 2021.
	return modelId
}

// WriteMap write map to file
func WriteMap(mymap map[string][]uint64) {
	w := csv.NewWriter(os.Stdout)
	for key, value := range mymap {
		r := make([]string, 0, 1+len(value))
		r = append(r, key)
		for _, v := range value {
			r = append(r, strconv.FormatUint(v, 10))
		}
		err := w.Write(r)
		if err != nil {
			// handle error
		}
	}
	w.Flush()
}

func makeRange(min, max int) []int {
	a := make([]int, max-min+1)
	for i := range a {
		a[i] = min + i
	}
	return a
}

func writeAsJSON(item interface{}, filename string) {
	os.MkdirAll(filepath.Dir(filename), 0755)
	file, err := os.Create(filename)
	fmt.Println(err)
	defer file.Close()
	jsonStr := FormatAsJSON(item)
	file.WriteString(fmt.Sprintf("%s\n", jsonStr))
}

func FormatAsJSON(item interface{}) string {
	jsonBytes, err := json.Marshal(item)
	jsonStr := ""
	if err != nil {
		fmt.Printf("There was an error decoding the json. err = %s", err)
		return jsonStr
	}
	jsonStr = string(jsonBytes)
	return jsonStr
}

func ReadFromJSONFile(filename string, targetStructPointer interface{}) error {
	jsonFile, err := os.Open(filename)
	fmt.Println(filename)
	if err != nil {
		log.WithError(err).Error("Cannot read JSON file " + filename)
		return err
	}
	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		log.WithError(err).Error("Contents of JSON file " + filename + " could not be parsed.")
		return err
	}
	err = json.Unmarshal(byteValue, targetStructPointer)
	if err != nil {
		log.WithError(err).Error("Failed to unmarshall")
		return err
	}
	return nil
}

func SmartChange(v1, v2 float64, mode string) float64 {
	var change float64
	if strings.HasPrefix(mode, "perc") || strings.HasPrefix(mode, "frac") {
		if v1 == 0 {
			if v2 == 0 {
				change = 0
			} else if v2 < 0 {
				change = -100
			} else if v2 > 0 {
				change = 100
			}
		} else {
			change = SmartDivide(v2-v1, v1) * 100
		}
		if strings.HasPrefix(mode, "frac") {
			change /= 100
		}
	} else if strings.HasPrefix(mode, "fact") {
		if v1 == 0 {
			v1 = 1
		}
		change = SmartDivide(v2, v1)
	}
	return change
}

func SmartDivide(a, b float64) float64 {
	if b == 0 {
		if a == 0 {
			return 0.0
		} else {
			return math.Inf(int(a))
		}
	} else {
		return a / b
	}
}

// func IsCandidateForANumber(s string) (bool, float64) {
// 	val, err := strconv.ParseFloat(s, 64)
// 	if err == nil {
// 		return false, 0
// 	} else {
// 		return true, val
// 	}
// }

var DATETIME_TYPE string = "DTT"
var MOBILE_NUMBER_TYPE string = "MNT"
var EMAIL_ID_TYPE string = "EIT"
var TIMESTAMP_TYPE string = "TT"
var URL_TYPE string = "URT"
var ID_TYPE string = "IDT"
var PERSON_NAME_TYPE string = "PNT"
var MISC_STRING_TYPE string = "MST"
var BOOLEAN_TYPE string = "BLT"
var INTEGER_TYPE string = "INT"
var IP_TYPE string = "IPT"
var REAL_NUM_TYPE string = "RNT"

// func InferFeatureType(keyName string, valCountMap Level1CatFreqDist, quickMode bool) string {
// 	SAMPLE_SIZE := 5
// 	allValues := reflect.ValueOf(valCountMap).MapKeys()
// 	var totalSize int = len(allValues)
// 	var sampleSize int
// 	var values []string
// 	if quickMode {
// 		rand.Shuffle(totalSize, func(i, j int) {
// 			allValues[i], allValues[j] = allValues[j], allValues[i]
// 		})
// 		sampleSize = SAMPLE_SIZE
// 	} else {
// 		sampleSize = totalSize
// 	}
// 	for i := 0; i < totalSize; i++ {
// 		if i >= sampleSize {
// 			break
// 		}
// 		values = append(values, allValues[i].String())
// 	}

// 	// for i, v := range values {
// 	// 	if  {

// 	// 	}
// 	// }
// }

func IsCandidateForANumber(stringVal string) bool {
	match, _ := regexp.MatchString("^[0-9]*.?[0-9]*$", stringVal)
	if match {
		return true
	} else {
		return false
	}
}

func IsADateFeature(keyName string, valCountMap Level1CatFreqDist) bool {
	if match, _ := regexp.MatchString("[Dd]ate", keyName); match {
		return true
	}
	return false
}

func IsANumericFeature(keyName string, valStats Level2CatRatioDist) bool {
	// TODO: Add the maxval logic later.
	// var MAXVAL int = 999999
	var nonNumericCount int = 0
	var numericCount int = 0
	for val := range valStats {
		numberFlag := IsCandidateForANumber(val)
		if numberFlag {
			numericCount++
		} else {
			nonNumericCount++
		}
	}
	if nonNumericCount == 0 {
		return true
	} else {
		return false
	}
}
