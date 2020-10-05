package util

import (
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

const SECONDS_IN_A_DAY int64 = 24 * 60 * 60
const EVENT_USER_CACHE_EXPIRY_SECS = 2592000

type TimeZoneString string

const (
	TimeZoneStringIST TimeZoneString = "Asia/Kolkata"
	TimeZoneStringUTC TimeZoneString = "UTC"
)

// QueryDateRangePresets Presets available for querying in dashboard / core query etc.
// TODO(prateek): Currently returns in IST. Change once timezone is added in project_settings.
var QueryDateRangePresets = map[string]func() (int64, int64){
	"LAST_MONTH":    GetQueryRangePresetLastMonthIST,
	"LAST_WEEK":     GetQueryRangePresetLastWeekIST,
	"CURRENT_MONTH": GetQueryRangePresetCurrentMonthIST,
	"CURRENT_WEEK":  GetQueryRangePresetCurrentWeekIST,
	"YESTERDAY":     GetQueryRangePresetYesterdayIST,
	"TODAY":         GetQueryRangePresetTodayIST,
}

// Group Names
var MostRecent string = "MOST RECENT"
var FrequentlySeen string = "FREQUENTLY SEEN"

func RandomString(n int) string {
	rand.Seed(time.Now().UnixNano())

	var letter = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	b := make([]rune, n)
	for i := range b {
		b[i] = letter[rand.Intn(len(letter))]
	}
	return string(b)
}

func RandomLowerAphaNumString(n int) string {
	rand.Seed(time.Now().UnixNano())

	var letter = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

	b := make([]rune, n)
	for i := range b {
		b[i] = letter[rand.Intn(len(letter))]
	}
	return string(b)
}

func RandomUint64() uint64 {
	return rand.Uint64()
}

func RandomUint64WithUnixNano() uint64 {
	return uint64(time.Now().UnixNano())
}

// RandomIntInRange Generates a random number in range [min, max).
func RandomIntInRange(min, max int) int {
	rand.Seed(time.Now().UTC().UnixNano())
	return min + rand.Intn(max-min)
}

func UnixTimeBeforeAWeek() int64 {
	return UnixTimeBeforeDuration(168 * time.Hour) // 7 days.
}

func UnixTimeBeforeDuration(duration time.Duration) int64 {
	return time.Now().UTC().Unix() - int64(duration.Seconds())
}

func IsNumber(num string) bool {
	// Use regex.
	_, err := strconv.ParseFloat(num, 64)
	if err != nil {
		return false
	}
	return true
}

func IsEmail(str string) bool {
	regexpEmail := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return regexpEmail.MatchString(str)
}

func GetNumberFromAnyString(str string) float64 {
	strAsBytes := []byte(str)
	re := regexp.MustCompile(`[+-]?([0-9]*[.])?[0-9]+`)
	numStr := string(re.Find(strAsBytes))

	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0
	}

	return num
}

func GetSortWeightFromAnyType(value interface{}) float64 {
	if value == nil {
		return 0
	}

	switch valueType := value.(type) {
	case float64:
		return value.(float64)
	case float32:
		return float64(value.(float32))
	case int:
		return float64(value.(int))
	case int32:
		return float64(value.(int32))
	case int64:
		return float64(value.(int64))
	case string:
		return GetNumberFromAnyString(value.(string))
	default:
		log.Info("Unsupported type used on GetSortWeightFromAnyType %+v", valueType)
		return 0
	}

	return 0
}

func TrimQuotes(str string) string {
	return strings.TrimSuffix(strings.TrimPrefix(str, "\""), "\"")
}

func GetValueAsString(value interface{}) (string, error) {
	switch value.(type) {
	case float32, float64:
		return fmt.Sprintf("%0.0f", value), nil
	case int, int32, int64:
		return fmt.Sprintf("%v", value), nil
	case string:
		return value.(string), nil
	case bool:
		return strconv.FormatBool(value.(bool)), nil
	default:
		return "", errors.New("invalid type to convert as string")
	}
}

func GetUUID() string {
	return uuid.New().String()
}

// StringSliceDiff Returns sliceA - sliceB set of elements.
func StringSliceDiff(sliceA, sliceB []string) []string {
	if len(sliceA) == 0 || len(sliceB) == 0 {
		return sliceA
	}
	sliceBMap := make(map[string]int)

	var diffSlice []string
	for index, value := range sliceB {
		sliceBMap[value] = index
	}

	for _, value := range sliceA {
		_, found := sliceBMap[value]
		if !found {
			diffSlice = append(diffSlice, value)
		}
	}
	return diffSlice
}

// StringValueIn Returns true if `value` is in `list` else false.
func StringValueIn(value string, list []string) bool {
	for _, val := range list {
		if val == value {
			return true
		}
	}
	return false
}

// Uint64ValueIn Returns true if `value` is in `list` else false.
func Uint64ValueIn(value uint64, list []uint64) bool {
	for _, val := range list {
		if val == value {
			return true
		}
	}
	return false
}

// FloatRoundOffWithPrecision Rounds of a float64 value to given precision. Ex: 2.667 with precision 2 -> 2.67.
func FloatRoundOffWithPrecision(value float64, precision int) (float64, error) {
	valueString := fmt.Sprintf("%0.*f", precision, value)
	roundOffValue, err := strconv.ParseFloat(valueString, 64)
	if err != nil {
		return roundOffValue, err
	}
	return roundOffValue, nil
}

// Datetime related utility functions.

const (
	DATETIME_FORMAT_YYYYMMDD_HYPHEN string = "2006-01-02"
	DATETIME_FORMAT_YYYYMMDD        string = "20060102"
)

// Returns date in YYYYMMDD format
func GetDateOnlyFromTimestamp(timestamp int64) string {
	return time.Unix(timestamp, 0).UTC().Format(DATETIME_FORMAT_YYYYMMDD)
}

// TimeNow Return current time in UTC. Should be used everywhere to avoid local timezone.
func TimeNow() time.Time {
	return time.Now().UTC()
}

// TimeNowIn Return's current time in given Timezone.
func TimeNowIn(timezone TimeZoneString) time.Time {
	timezoneLocation, _ := time.LoadLocation(string(timezone))
	return time.Now().In(timezoneLocation)
}

// ConvertTimeIn Converts given time.Time object in given timezone.
func ConvertTimeIn(t time.Time, timezone TimeZoneString) time.Time {
	timezoneLocation, _ := time.LoadLocation(string(timezone))
	return t.In(timezoneLocation)
}

// GetTimeLocationFor Returns time.Location object for given timezone.
func GetTimeLocationFor(timezone TimeZoneString) *time.Location {
	timezoneLocation, _ := time.LoadLocation(string(timezone))
	return timezoneLocation
}

// TimeNowUnix Returns current epoch time.
func TimeNowUnix() int64 {
	return TimeNow().Unix()
}

func GetCurrentDayTimestamp() int64 {
	currentTime := time.Now()
	return time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 0, 0, 0, 0, currentTime.Location()).Unix()
}

func IsTimestampToday(timestamp int64) bool {
	return GetBeginningOfDayTimestamp(timestamp) == GetCurrentDayTimestamp()
}

func GetBeginningOfDayTimestamp(timestamp int64) int64 {
	t := time.Unix(timestamp, 0)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).Unix()
}

func GetBeginningOfDayTimestampUTC(timestamp int64) int64 {
	t := time.Unix(timestamp, 0).UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).Unix()
}

// GetBeginningOfDayTimestampZ Get's beginning of the day timestamp in given timezone.
func GetBeginningOfDayTimestampZ(timestamp int64, timezoneString TimeZoneString) int64 {
	location, _ := time.LoadLocation(string(timezoneString))
	t := time.Unix(timestamp, 0).In(location)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).Unix()
}

// GetEndOfDayTimestampZ Get's end of the day timestamp in given timezone.
func GetEndOfDayTimestampZ(timestamp int64, timezoneString TimeZoneString) int64 {
	location, _ := time.LoadLocation(string(timezoneString))
	t := time.Unix(timestamp, 0).In(location)
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), t.Location()).Unix()
}

// GetQueryRangePresetCurrentWeekIST Returns start and end unix timestamp for current week (Sunday to Yesterday)
func GetQueryRangePresetCurrentWeekIST() (int64, int64) {
	locationIST, _ := time.LoadLocation(string(TimeZoneStringIST))
	if isTodayFirstDayOfWeek(locationIST) {
		return GetQueryRangePresetTodayIST()
	}
	rangeStartTime := GetBeginningOfDayTimestampZ(BeginningOfWeekStartOfDay(locationIST).Unix(), TimeZoneStringIST)
	timeNow := time.Now().In(locationIST)
	rangeEndTime := GetBeginningOfDayTimestampZ(timeNow.Unix(), TimeZoneStringIST) - 1
	return rangeStartTime, rangeEndTime
}

func isTodayFirstDayOfMonth(location *time.Location) bool {
	t := time.Now().In(location)
	return t.Day() == 1
}

func isTodayFirstDayOfWeek(location *time.Location) bool {
	t := time.Now().In(location)
	return int(t.Weekday()) == 0
}

// GetQueryRangePresetCurrentMonthIST Returns start and end unix timestamp for current month (1st of month to Yesterday)
func GetQueryRangePresetCurrentMonthIST() (int64, int64) {
	locationIST, _ := time.LoadLocation(string(TimeZoneStringIST))
	if isTodayFirstDayOfMonth(locationIST) {
		return GetQueryRangePresetTodayIST()
	}
	rangeStartTime := GetBeginningOfDayTimestampZ(BeginningOfMonthStartOfDay(locationIST).Unix(), TimeZoneStringIST)
	timeNow := time.Now().In(locationIST)
	rangeEndTime := GetBeginningOfDayTimestampZ(timeNow.Unix(), TimeZoneStringIST) - 1
	return rangeStartTime, rangeEndTime
}

// GetQueryRangePresetLastWeekIST Returns start and end unix timestamp for last week (Last to Last Sunday to Last Sat)
func GetQueryRangePresetLastWeekIST() (int64, int64) {
	locationIST, _ := time.LoadLocation(string(TimeZoneStringIST))
	rangeStartTime := GetBeginningOfDayTimestampZ(BeginningOfLastWeekStartOfDay(locationIST).Unix(), TimeZoneStringIST)
	rangeEndTime := GetEndOfDayTimestampZ(EndOfLastWeekStartOfDay(locationIST).Unix(), TimeZoneStringIST)
	return rangeStartTime, rangeEndTime
}

// GetQueryRangePresetLastMonthIST Returns start and end unix timestamp for last month (1st of the last month to end of the last month)
func GetQueryRangePresetLastMonthIST() (int64, int64) {
	locationIST, _ := time.LoadLocation(string(TimeZoneStringIST))
	rangeStartTime := GetBeginningOfDayTimestampZ(BeginningOfLastMonthStartOfDay(locationIST).Unix(), TimeZoneStringIST)
	rangeEndTime := GetEndOfDayTimestampZ(EndOfLastMonthStartOfDay(locationIST).Unix(), TimeZoneStringIST)
	return rangeStartTime, rangeEndTime
}

// BeginningOfWeekStartOfDay beginning of current week
func BeginningOfWeekStartOfDay(location *time.Location) time.Time {
	t := time.Now().In(location)
	weekday := int(t.Weekday())
	t = t.AddDate(0, 0, -weekday)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// BeginningOfLastWeekStartOfDay beginning of last week
func BeginningOfLastWeekStartOfDay(location *time.Location) time.Time {
	t := time.Now().In(location)
	t = t.AddDate(0, 0, -(int(t.Weekday()) + 7))
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// EndOfLastWeekStartOfDay end of last week
func EndOfLastWeekStartOfDay(location *time.Location) time.Time {
	t := BeginningOfWeekStartOfDay(location)
	t = t.AddDate(0, 0, -1)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// BeginningOfMonthStartOfDay beginning of current month
func BeginningOfMonthStartOfDay(location *time.Location) time.Time {
	t := time.Now().In(location)
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// BeginningOfLastMonthStartOfDay beginning of last month
func BeginningOfLastMonthStartOfDay(location *time.Location) time.Time {
	t := BeginningOfMonthStartOfDay(location)
	monthEnd := t.AddDate(0, 0, -(1))
	return time.Date(monthEnd.Year(), monthEnd.Month(), 1, 0, 0, 0, 0, monthEnd.Location())
}

// EndOfLastMonthStartOfDay end of last month
func EndOfLastMonthStartOfDay(location *time.Location) time.Time {
	t := BeginningOfMonthStartOfDay(location)
	return t.AddDate(0, 0, -(1))
}

// GetQueryRangePresetYesterdayIST Returns start and end unix timestamp for yesterday's range.
func GetQueryRangePresetYesterdayIST() (int64, int64) {
	locationIST, _ := time.LoadLocation(string(TimeZoneStringIST))
	timeNow := time.Now().In(locationIST)
	rangeEndTime := GetBeginningOfDayTimestampZ(timeNow.Unix(), TimeZoneStringIST) - 1
	rangeStartTime := GetBeginningOfDayTimestampZ(timeNow.AddDate(0, 0, -1).Unix(), TimeZoneStringIST)
	return rangeStartTime, rangeEndTime
}

// GetQueryRangePresetTodayIST Returns start and end unix timestamp for today's range.
func GetQueryRangePresetTodayIST() (int64, int64) {
	locationIST, _ := time.LoadLocation(string(TimeZoneStringIST))
	timeNow := time.Now().In(locationIST)
	rangeStartTime := GetBeginningOfDayTimestampZ(timeNow.Unix(), TimeZoneStringIST)
	return rangeStartTime, timeNow.Unix()
}

// UnixToHumanTime Converts epoch to readable String timestamp.
func UnixToHumanTime(timestamp int64) string {
	return time.Unix(timestamp, 0).UTC().Format(time.RFC3339)
}

// SecondsToHMSString Converts seconds int value to hrs min secs string.
func SecondsToHMSString(totalSeconds int64) string {
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60
	return fmt.Sprintf("%d hrs %d mins %d secs", hours, minutes, seconds)
}

func Min(a int64, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func Max(a int64, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func MinInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func MaxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

// returns a string of ('?'), ... values of given batchSize
func GetValuePlaceHolder(batchSize int) string {

	concatenatedIds := ""
	for i := 0; i < batchSize; i++ {
		concatenatedIds = concatenatedIds + " (?),"
	}
	// removing that extra ',' at end
	concatenatedIds = concatenatedIds[0 : len(concatenatedIds)-1]
	return concatenatedIds
}

// returns a interface list from given string list
func GetInterfaceList(list []string) []interface{} {

	listInterface := make([]interface{}, len(list))
	for i, v := range list {
		listInterface[i] = v
	}
	return listInterface
}

// GetStringListAsBatch - Splits string list into multiple lists.
func GetStringListAsBatch(list []string, batchSize int) [][]string {
	batchList := make([][]string, 0, 0)
	listLen := len(list)
	for i := 0; i < listLen; {
		next := i + batchSize
		if next > listLen {
			next = listLen
		}

		batchList = append(batchList, list[i:next])
		i = next
	}

	return batchList
}

func GetSnakeCaseToTitleString(str string) (title string) {
	if str == "" {
		return
	}

	tokens := strings.Split(str, "_")
	for i, token := range tokens {
		if i == 0 {
			title = strings.Title(token)
			continue
		}

		title = fmt.Sprintf("%s %s", title, strings.Title(token))
	}

	return title
}

func IsContainsAnySubString(src string, sub ...string) bool {
	for _, s := range sub {
		if strings.Contains(src, s) {
			return true
		}
	}

	return false
}
