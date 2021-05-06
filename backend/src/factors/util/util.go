package util

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/ttacon/libphonenumber"
)

// Int64Tuple To be used at places where an int64 tuple is required to be passed or returned.
type Int64Tuple struct {
	First  int64
	Second int64
}

const SECONDS_IN_A_DAY int64 = 24 * 60 * 60
const EVENT_USER_CACHE_EXPIRY_SECS = 1728000

type TimeZoneString string

const (
	TimeZoneStringIST TimeZoneString = "Asia/Kolkata"
	TimeZoneStringUTC TimeZoneString = "UTC"

	DateRangePresetLastMonth    = "LAST_MONTH"
	DateRangePresetLastWeek     = "LAST_WEEK"
	DateRangePresetCurrentMonth = "CURRENT_MONTH"
	DateRangePresetCurrentWeek  = "CURRENT_WEEK"
	DateRangePresetYesterday    = "YESTERDAY"
	DateRangePresetToday        = "TODAY"
	DateRangePreset30Minutes    = "30MINS"
)

// NullcharBytes is null charater bytes, specific for golang json marhsal date
var NullcharBytes = []byte{0x5c, 0x75, 0x30, 0x30, 0x30, 0x30} // \u0000
var spaceBytes = []byte{0x20}

// QueryDateRangePresets Presets available for querying in dashboard / core query etc.
// TODO(prateek): Currently returns in IST. Change once timezone is added in project_settings.
var QueryDateRangePresets = map[string]func() (int64, int64){
	DateRangePresetLastMonth:    GetQueryRangePresetLastMonthIST,
	DateRangePresetLastWeek:     GetQueryRangePresetLastWeekIST,
	DateRangePresetCurrentMonth: GetQueryRangePresetCurrentMonthIST,
	DateRangePresetCurrentWeek:  GetQueryRangePresetCurrentWeekIST,
	DateRangePresetYesterday:    GetQueryRangePresetYesterdayIST,
	DateRangePresetToday:        GetQueryRangePresetTodayIST,
}

// WebAnalyticsQueryDateRangePresets Date range presets for website analytics.
var WebAnalyticsQueryDateRangePresets = map[string]func() (int64, int64){
	DateRangePresetLastMonth:    GetQueryRangePresetLastMonthIST,
	DateRangePresetLastWeek:     GetQueryRangePresetLastWeekIST,
	DateRangePresetCurrentMonth: GetQueryRangePresetCurrentMonthIST,
	DateRangePresetCurrentWeek:  GetQueryRangePresetCurrentWeekIST,
	DateRangePresetYesterday:    GetQueryRangePresetYesterdayIST,
	DateRangePresetToday:        GetQueryRangePresetTodayIST,
	DateRangePreset30Minutes:    GetQueryRangePresetLast30MinutesIST,
}

// Caching related constants for core query and dashboards.
const (
	// ImmutableDataEndDateBufferInSeconds Buffer period after which data is assumed to be immutable. Required to rerun
	// cached queries when add_session / other enrichment jobs were still running or not yet ran when data got cached.
	ImmutableDataEndDateBufferInSeconds      = 2 * SECONDS_IN_A_DAY // 2 Days.
	CacheExpiryDashboardMutableDataInSeconds = 1 * SECONDS_IN_A_DAY // 1 Days.
	CacheExpiryQueryMutableDataInSeconds     = 2 * 60 * 60          // 2 Hours.

	CacheExpiryWeeklyRangeInSeconds = 6 * 7 * SECONDS_IN_A_DAY // 6 Weeks.
	CacheExpiryDefaultInSeconds     = 62 * SECONDS_IN_A_DAY    // 62 Days.

	CacheExpiryQueryMaxInSeconds            = 2 * SECONDS_IN_A_DAY
	CacheExpiryQueryTodaysDataInSeconds     = 10 * 60      // 10 minutes.
	CacheExpiryDashboardTodaysDataInSeconds = 12 * 60 * 60 // 12 hours.
	CacheExpiryDashboard30MinutesInSeconds  = 12 * 60 * 60 // 12 hours.
)

// Group Names
var MostRecent string = "MOST RECENT"
var FrequentlySeen string = "FREQUENTLY SEEN"
var SmartEvent string = "SMART EVENTS"

func RandomString(n int) string {
	rand.Seed(time.Now().UnixNano())

	var letter = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	b := make([]rune, n)
	for i := range b {
		b[i] = letter[rand.Intn(len(letter))]
	}
	return string(b)
}

func RandomNumericString(n int) string {
	rand.Seed(time.Now().UnixNano())

	var letter = []rune("0123456789")

	b := make([]rune, n)
	for i := range b {
		b[i] = letter[rand.Intn(len(letter))]
	}
	return string(b)
}

func RandomNumericStringNonZeroStart(n int) string {
	rand.Seed(time.Now().UnixNano())

	var letter = []rune("0123456789")
	var letterNonZero = []rune("123456789")

	b := make([]rune, n)
	for i := range b {
		if i == 0 {
			b[i] = letter[rand.Intn(len(letterNonZero))]
		} else {
			b[i] = letter[rand.Intn(len(letter))]
		}
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

// RemoveNullCharacterBytes return bytes by replacing null character bytes by space. Only works for json marshaled data
func RemoveNullCharacterBytes(existingbytes []byte) []byte {
	newBytes := bytes.ReplaceAll(existingbytes, NullcharBytes, spaceBytes)
	return newBytes
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
	regexpEmail := regexp.MustCompile(`^[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,7}$`)
	return regexpEmail.MatchString(str)
}

func IsValidUrl(tocheck string) bool {

	r, _ := regexp.Compile("(^(http|https)://)?(www)?.com")
	if strings.HasPrefix(tocheck, "$session") {
		return false
	}
	return r.MatchString(tocheck)
	// return true
}

// GetEmailLowerCase returns email in lower case for consistency
func GetEmailLowerCase(emailI interface{}) string {
	email := GetPropertyValueAsString(emailI)
	if !IsEmail(email) {
		return ""
	}

	return strings.ToLower(email)
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

// SafeConvertToFloat64 Converts an interface to float64 value.
func SafeConvertToFloat64(value interface{}) float64 {
	return GetSortWeightFromAnyType(value)
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

func IsValidUUID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
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

// IfThenElse Is a hack to get ternary one liners
func IfThenElse(condition bool, a interface{}, b interface{}) interface{} {
	if condition {
		return a
	}
	return b
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
	DATETIME_FORMAT_DB              string = "2006-01-02 00:00:00"
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

// GetDateAsStringZ gets date in given timezone.
func GetDateAsStringZ(timestamp int64, timezoneString TimeZoneString) int64 {
	if len(timezoneString) < 1 {
		timezoneString = TimeZoneStringIST
	}
	location, _ := time.LoadLocation(string(timezoneString))
	t := time.Unix(timestamp, 0).In(location)
	value, _ := strconv.ParseInt(t.Format(DATETIME_FORMAT_YYYYMMDD), 10, 64)
	return value
}

// IsStartOfTodaysRange Checks if the from value is start of todays range.
func IsStartOfTodaysRange(from int64, timezoneString TimeZoneString) bool {
	return from == GetBeginningOfDayTimestampZ(TimeNowUnix(), TimeZoneStringIST)
}

// Is30MinutesTimeRange Whether time range is of last 30 minutes.
func Is30MinutesTimeRange(from, to int64) bool {
	return (to - from) == 30*60
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

// GetMonthlyQueryRangesTuplesIST Return a tuple of start and end unix timestamp for last `numMonths`.
func GetMonthlyQueryRangesTuplesIST(numMonths int) []Int64Tuple {
	monthlyRanges := make([]Int64Tuple, 0, 0)
	if numMonths == 0 {
		return monthlyRanges
	}
	lastMonthStart, lastMonthEnd := GetQueryRangePresetLastMonthIST()
	locationIST, _ := time.LoadLocation(string(TimeZoneStringIST))
	monthlyRanges = append(monthlyRanges, Int64Tuple{lastMonthStart, lastMonthEnd})
	for i := 1; i < numMonths; i++ {
		lastMonthEnd = lastMonthStart - 1
		lastMonthEndTime := time.Unix(lastMonthEnd, 0).In(locationIST)
		lastMonthStart = time.Date(lastMonthEndTime.Year(), lastMonthEndTime.Month(),
			1, 0, 0, 0, 0, lastMonthEndTime.Location()).Unix()
		monthlyRanges = append(monthlyRanges, Int64Tuple{lastMonthStart, lastMonthEnd})
	}

	return monthlyRanges
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

// GetQueryRangePresetLast30MinutesIST Returns start and end unix timestamp for last 30mins range.
func GetQueryRangePresetLast30MinutesIST() (int64, int64) {
	locationIST, _ := time.LoadLocation(string(TimeZoneStringIST))
	timeNow := time.Now().In(locationIST)
	endTime := timeNow.Unix()
	startTime := endTime - 30*60
	return startTime, endTime
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

func MaxFloat64(a float64, b float64) float64 {
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

// GetUint64ListAsBatch - Returns list of uint64 as batches of uint64 list.
func GetUint64ListAsBatch(list []uint64, batchSize int) [][]uint64 {
	batchList := make([][]uint64, 0, 0)
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

func GetInterfaceListAsBatch(list []interface{}, batchSize int) [][]interface{} {
	batchList := make([][]interface{}, 0, 0)
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

// isPureisPurePhoneNumberNumber checks for pure number string
func isPurePhoneNumber(phoneNo string) bool {
	if _, err := strconv.Atoi(phoneNo); err == nil {
		return true
	}
	return false
}

func getPhoneNoSeparatorIndex(phoneNo *string) []int {
	var separatorsIndex []int
	separators := []string{" ", "-"}
	for _, seperator := range separators {
		if spIndex := strings.Index(*phoneNo, seperator); spIndex != -1 {
			separatorsIndex = append(separatorsIndex, spIndex)
		}
	}
	return separatorsIndex
}

// isPhoneValidCountryCode checks if country code exist in libphonenumber list
func isPhoneValidCountryCode(cCode string) bool {
	if code, err := strconv.Atoi(cCode); err == nil {
		if _, exist := libphonenumber.CountryCodeToRegion[code]; exist {
			return true
		}
	}
	return false
}

// maybeAddInternationalPrefix adds if missing '+' at the beginning for the libphonenumber to work
func maybeAddInternationalPrefix(phoneNo *string) bool {
	// Ex 91 1234567890
	separatorsIndex := getPhoneNoSeparatorIndex(phoneNo)

	for _, indexValue := range separatorsIndex {
		cCode := string((*phoneNo)[:indexValue])
		if isPhoneValidCountryCode(cCode) {
			*phoneNo = libphonenumber.PLUS_CHARS + *phoneNo
			return true
		}
	}
	return false
}

// getInternationalPhoneNoWithoutCountryCode removes country code
func getInternationalPhoneNoWithoutCountryCode(intPhone string) string {
	phoneNo := strings.SplitN(intPhone, " ", 2)
	return strings.Join(phoneNo[1:], "")
}

// GetPossiblePhoneNumber returns all possible phone number for specific phone number pattern
func GetPossiblePhoneNumber(phoneNo string) []string {
	var possiblePhoneNo []string
	var phonePattern string
	possiblePhoneNo = append(possiblePhoneNo, phoneNo)

	// try pure numbers if form submited also had pure numbers
	if isPurePhoneNumber(phoneNo) {
		if !strings.Contains(phoneNo, "+") {
			possiblePhoneNo = append(possiblePhoneNo, "0"+phoneNo)
			possiblePhoneNo = append(possiblePhoneNo, "+"+phoneNo)
		}

		//0123-456-789 or (012)-345-6789
		nationalPhoneNo := strings.TrimPrefix(phoneNo, "+")
		if len(nationalPhoneNo) == 10 {
			phonePattern = fmt.Sprintf("%s-%s-%s", nationalPhoneNo[:3], nationalPhoneNo[3:6], nationalPhoneNo[6:])
			possiblePhoneNo = append(possiblePhoneNo, phonePattern)
			phonePattern = fmt.Sprintf("(%s)-%s-%s", nationalPhoneNo[:3], nationalPhoneNo[3:6], nationalPhoneNo[6:])
			possiblePhoneNo = append(possiblePhoneNo, phonePattern)
			phonePattern = fmt.Sprintf("(%s) %s %s", nationalPhoneNo[:3], nationalPhoneNo[3:6], nationalPhoneNo[6:])
			possiblePhoneNo = append(possiblePhoneNo, phonePattern)
		}
	}

	//phone number having '+' will have country code attached
	num, err := libphonenumber.Parse(phoneNo, "")
	if err != nil {
		// ErrInvalidCountryCode describes missing '+', can be added if phone number in format
		if err == libphonenumber.ErrInvalidCountryCode && maybeAddInternationalPrefix(&phoneNo) {
			num, err = libphonenumber.Parse(phoneNo, "")
		}
	}

	if err == nil {

		//international format +91 1234 567 890
		intFormat := libphonenumber.Format(num, libphonenumber.INTERNATIONAL)
		if intFormat != phoneNo {
			possiblePhoneNo = append(possiblePhoneNo, intFormat)
		}

		//International without country code 1234 567 890
		possiblePhoneNo = append(possiblePhoneNo, getInternationalPhoneNoWithoutCountryCode(intFormat))
		possiblePhoneNo = append(possiblePhoneNo, libphonenumber.Format(num, libphonenumber.NATIONAL))

		//911234567890
		nationalFormat := libphonenumber.Format(num, libphonenumber.E164)[1:]
		possiblePhoneNo = append(possiblePhoneNo, nationalFormat)

		//National format 1234 567 890
		nationalNum := libphonenumber.GetNationalSignificantNumber(num)
		possiblePhoneNo = append(possiblePhoneNo, nationalNum)
		possiblePhoneNo = append(possiblePhoneNo, "0"+nationalNum)
		possiblePhoneNo = append(possiblePhoneNo, "+"+nationalNum)

		standardPhone := libphonenumber.Format(num, libphonenumber.E164)
		if standardPhone != "+91"+nationalNum {
			// standard phone number +911234567890. Assuming input phone number be +91 123 456 7890 or similar
			possiblePhoneNo = append(possiblePhoneNo, standardPhone)
		}

		//+9101234 567 890
		possiblePhoneNo = append(possiblePhoneNo, fmt.Sprintf("+%d%s", num.GetCountryCode(), libphonenumber.Format(num, libphonenumber.NATIONAL)))

	}

	return possiblePhoneNo
}

// BytesToReadableFormat Pretty prints bytes to readable KiB/MiB/GiB.. format.
func BytesToReadableFormat(bytes float64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%.1f B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB",
		float64(bytes)/float64(div), "KMGTPE"[exp])
}

// SanitizePhoneNumber currently removes only leading 0 from phone number. New functionalty will added as when required
func SanitizePhoneNumber(v interface{}) string {
	phoneNo := GetPropertyValueAsString(v)

	if phoneNo == "" || len(phoneNo) < 4 {
		return ""
	}

	if phoneNo[0] == '0' {
		return phoneNo[1:]
	}

	return phoneNo
}

func InsertRepeated(strList []string, index int, value string) []string {
	if len(strList) == index { // nil or empty slice or after last element
		return append(strList, value)
	}
	strList = append(strList[:index+1], strList[index:]...) // index < len(a)
	strList[index] = value
	return strList
}

func RankByWordCount(wordFrequencies map[string]int) PairList {
	pl := make(PairList, len(wordFrequencies))
	i := 0
	for k, v := range wordFrequencies {
		pl[i] = Pair{k, v}
		i++
	}
	sort.Sort(sort.Reverse(pl))
	return pl
}

// GenerateHash To generate hash value for given byte array.
func GenerateHash(bytes []byte) string {
	hasher := sha1.New()
	hasher.Write(bytes)
	sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	return sha
}

// GenerateHashStringForStruct Marshals the passed struct and generates a unique hash string.
func GenerateHashStringForStruct(queryPayload interface{}) (string, error) {
	queryCacheBytes, err := json.Marshal(queryPayload)
	if err != nil {
		return "", err
	}
	return GenerateHash(queryCacheBytes), nil
}

// DeepCopy Deep copies a to b using json marshaling.
func DeepCopy(a, b interface{}) {
	byt, _ := json.Marshal(a)
	json.Unmarshal(byt, b)
}

type Pair struct {
	Key   string
	Value int
}

type PairList []Pair

func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// IsStandardEvent check if eventname is standard Event
func IsStandardEvent(eventName string) bool {
	return strings.HasPrefix(eventName, "$")
}

// IsCampaignEvent check if eventname is campaign Event
func IsCampaignEvent(eventName string) bool {
	return strings.HasPrefix(eventName, "$session[campaign")
}

// IsSourceEvent check if eventname is campaign Event
func IsSourceEvent(eventName string) bool {
	return strings.HasPrefix(eventName, "$session[source")
}

// IsMediumEvent check if eventname is campaign Event
func IsMediumEvent(eventName string) bool {
	return strings.HasPrefix(eventName, "$session[medium")
}

// IsReferrerEvent check if eventname is referrer Event
func IsReferrerEvent(eventName string) bool {
	return strings.HasPrefix(eventName, "$session[initial_referrer")
}

func IsAdgroupEvent(eventName string) bool {
	return strings.HasPrefix(eventName, "$session[adgroup")
}

func IsCampaignAnalytics(eventName string) bool {
	if IsCampaignEvent(eventName) == true || IsMediumEvent(eventName) == true || IsSourceEvent(eventName) == true || IsReferrerEvent(eventName) == true || IsAdgroupEvent(eventName) == true {
		return true
	}
	return false
}

func IsItreeCampaignEvent(eventName string) bool {
	return strings.HasPrefix(eventName, "$session[campaign") || strings.HasPrefix(eventName, "$session[source") || strings.HasPrefix(eventName, "$session[medium") || strings.HasPrefix(eventName, "$session[adgroup") || strings.HasPrefix(eventName, "$session[initial_referrer")
}

// GetDashboardCacheResultExpiryInSeconds Returns expiry for cache based on query date range.
func GetDashboardCacheResultExpiryInSeconds(from, to int64) float64 {
	toStartOfDay := GetBeginningOfDayTimestampZ(to, TimeZoneStringIST)
	nowStartOfDay := GetBeginningOfDayTimestampZ(TimeNow().Unix(), TimeZoneStringIST)
	if Is30MinutesTimeRange(from, to) {
		return float64(CacheExpiryDashboard30MinutesInSeconds)
	} else if to >= nowStartOfDay {
		// End date is in today's range. Keep small expiry.
		return float64(CacheExpiryDashboardTodaysDataInSeconds)
	} else if nowStartOfDay > toStartOfDay && nowStartOfDay-toStartOfDay > ImmutableDataEndDateBufferInSeconds {
		// Data can be assumed to be immutable here after buffer (2) days.
		if to-from == (7*SECONDS_IN_A_DAY - 1) {
			// Weekly range.
			return float64(CacheExpiryWeeklyRangeInSeconds)
		} else if to-from > (27*SECONDS_IN_A_DAY - 1) {
			// Monthly range. Set no expiry.
			return 0
		} else {
			return float64(CacheExpiryDefaultInSeconds)
		}
	}
	return float64(CacheExpiryDashboardMutableDataInSeconds)
}

// GetQueryCacheResultExpiryInSeconds Returns expiry for cache based on query date range.
func GetQueryCacheResultExpiryInSeconds(from, to int64) float64 {
	toStartOfDay := GetBeginningOfDayTimestampZ(to, TimeZoneStringIST)
	nowStartOfDay := GetBeginningOfDayTimestampZ(TimeNow().Unix(), TimeZoneStringIST)
	if to >= nowStartOfDay {
		// End date is in today's range. Keep small expiry.
		return float64(CacheExpiryQueryTodaysDataInSeconds)
	} else if nowStartOfDay > toStartOfDay && nowStartOfDay-toStartOfDay > ImmutableDataEndDateBufferInSeconds {
		// Data can be assumed to be immutable here after buffer (2) days.
		return float64(CacheExpiryQueryMaxInSeconds)
	}
	return float64(CacheExpiryQueryMutableDataInSeconds)
}

func GetAggrAsFloat64(aggr interface{}) (float64, error) {
	switch aggr.(type) {
	case int:
		return float64(aggr.(int)), nil
	case int64:
		return float64(aggr.(int64)), nil
	case float32:
		return float64(aggr.(float32)), nil
	case float64:
		return aggr.(float64), nil
	case string:
		aggrInt, err := strconv.ParseInt(aggr.(string), 10, 64)
		return float64(aggrInt), err
	default:
		return float64(0), errors.New("invalid aggregate value type")
	}
}

// CleanSplitByDelimiter Splits a string by delimiter and removes any spaces.
// Ex: "a, b, c" and "a,b,c" will return same ["a", "b", "c"].
func CleanSplitByDelimiter(str string, del string) []string {
	split := strings.Split(str, del)

	cleanSplit := make([]string, 0, 0)
	for _, s := range split {
		cleanSplit = append(cleanSplit, strings.TrimSpace(s))
	}
	return cleanSplit
}
