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
)

const SECONDS_IN_A_DAY int64 = 24 * 60 * 60

type TimeZoneString string

const (
	TimeZoneStringIST TimeZoneString = "Asia/Kolkata"
	TimeZoneStringUTC TimeZoneString = "UTC"
)

// QueryDateRangePresets Presets available for querying in dashboard / core query etc.
// TODO(prateek): Currently returns in IST. Change once timezone is added in project_settings.
var QueryDateRangePresets = map[string]func() (int64, int64){
	"30_DAYS":   GetQueryRangePresetLast30Days,
	"7_DAYS":    GetQueryRangePresetLast7Days,
	"YESTERDAY": GetQueryRangePresetYesterday,
	"TODAY":     GetQueryRangePresetToday,
}

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
)

// TimeNow Return current time in UTC. Should be used everywhere to avoid local timezone.
func TimeNow() time.Time {
	return time.Now().UTC()
}

// TimeNowIn Return's current time in given Timezone.
func TimeNowIn(timezone TimeZoneString) time.Time {
	timezoneLocation, _ := time.LoadLocation(string(timezone))
	return time.Now().In(timezoneLocation)
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

// GetQueryRangePresetLast30Days Returns start and end unix timestamp for last 30 days range.
func GetQueryRangePresetLast30Days() (int64, int64) {
	locationIST, _ := time.LoadLocation(string(TimeZoneStringIST))
	timeNow := time.Now().In(locationIST)
	rangeEndTime := GetBeginningOfDayTimestampZ(timeNow.Unix(), TimeZoneStringIST) - 1
	rangeStartTime := GetBeginningOfDayTimestampZ(timeNow.AddDate(0, 0, -30).Unix(), TimeZoneStringIST)
	return rangeStartTime, rangeEndTime
}

// GetQueryRangePresetLast7Days Returns start and end unix timestamp for last 7 days range.
func GetQueryRangePresetLast7Days() (int64, int64) {
	locationIST, _ := time.LoadLocation(string(TimeZoneStringIST))
	timeNow := time.Now().In(locationIST)
	rangeEndTime := GetBeginningOfDayTimestampZ(timeNow.Unix(), TimeZoneStringIST) - 1
	rangeStartTime := GetBeginningOfDayTimestampZ(timeNow.AddDate(0, 0, -7).Unix(), TimeZoneStringIST)
	return rangeStartTime, rangeEndTime
}

// GetQueryRangePresetYesterday Returns start and end unix timestamp for yesterday's range.
func GetQueryRangePresetYesterday() (int64, int64) {
	locationIST, _ := time.LoadLocation(string(TimeZoneStringIST))
	timeNow := time.Now().In(locationIST)
	rangeEndTime := GetBeginningOfDayTimestampZ(timeNow.Unix(), TimeZoneStringIST) - 1
	rangeStartTime := GetBeginningOfDayTimestampZ(timeNow.AddDate(0, 0, -1).Unix(), TimeZoneStringIST)
	return rangeStartTime, rangeEndTime
}

// GetQueryRangePresetToday Returns start and end unix timestamp for today's range.
func GetQueryRangePresetToday() (int64, int64) {
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
