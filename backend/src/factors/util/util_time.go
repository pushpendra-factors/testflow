package util

import (
	"errors"

	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/jinzhu/now"

	log "github.com/sirupsen/logrus"
)

// Datetime related utility functions.
// General convention for date Functions - suffix Z if utc based, In if timezone is passed, no suffix if localTime.
const (
	DATETIME_FORMAT_YYYYMMDD_HYPHEN  string = "2006-01-02"
	DATETIME_FORMAT_YYYYMMDD         string = "20060102"
	DATETIME_FORMAT_DB               string = "2006-01-02 15:04:05"
	SECONDS_IN_YEAR                         = ((365 * 86400) + 20736)
	DATETIME_FORMAT_DB_WITH_TIMEZONE string = "2006-01-02T15:04:05-07:00"
)

var BaseYears = []int{2022, 2023, 2024, 2025, 2026, 2027, 2028, 2029, 2030, 2031, 2032, 2033, 2034, 2035}

type TimestampRange struct {
	Start int64
	End   int64
}

// If 15ThAug, 2020 00:00:00 Asia/Kolkata is result will be 15ThAug, 2020 00:00:00 UTC.
func ConvertEqualTimeFromOtherTimezoneToUTC(timestamp int64, timezone TimeZoneString) int64 {
	in := GetTimeLocationFor(timezone)
	locTime := time.Unix(timestamp, 0).In(in)
	utcLoc := GetTimeLocationFor(TimeZoneString("UTC"))

	t := time.Date(locTime.Year(), locTime.Month(), locTime.Day(), locTime.Hour(), locTime.Minute(), locTime.Second(), 0, utcLoc)
	return t.Unix()
}

// If 15ThAug, 2020 00:00:00 UTC is result will be 15ThAug, 2020 00:00:00 Asia/Kolkata.
func ConvertEqualTimeFromUTCToInOtherTimezone(timestamp int64, timezone TimeZoneString) int64 {
	utcTime := time.Unix(timestamp, 0).UTC()
	in := GetTimeLocationFor(timezone)

	t := time.Date(utcTime.Year(), utcTime.Month(), utcTime.Day(), utcTime.Hour(), utcTime.Minute(), utcTime.Second(), 0, in)
	return t.Unix()
}

// Returns date in YYYYMMDD format
func GetDateOnlyFromTimestampZ(timestamp int64) string {
	return time.Unix(timestamp, 0).UTC().Format(DATETIME_FORMAT_YYYYMMDD)
}

func GetDateOnlyHyphenFormatFromTimestampZ(timestamp int64) string {
	return time.Unix(timestamp, 0).UTC().Format(DATETIME_FORMAT_YYYYMMDD_HYPHEN)
}

// Returns date in YYYYMMDD format for a given timestamp and timezone
func GetDateOnlyFormatFromTimestampAndTimezone(timestamp int64, timezone TimeZoneString) string {
	in := GetTimeLocationFor(timezone)
	return time.Unix(timestamp, 0).In(in).Format(DATETIME_FORMAT_YYYYMMDD)
}

func GetDBDateFormatFromTimestampAndTimezone(timestamp int64, timezone TimeZoneString) string {
	in := GetTimeLocationFor(timezone)
	return time.Unix(timestamp, 0).In(in).Format(DATETIME_FORMAT_DB)
}

// GetDateFormatFromTimestampAndTimezone Returns date in "02 Jan 2006" format for a given timestamp and timezone
func GetDateFromTimestampAndTimezone(timestamp int64, timezone TimeZoneString) string {
	in := GetTimeLocationFor(timezone)
	return time.Unix(timestamp, 0).In(in).Format("02 Jan")
}

func GetTimestampFromDateTimeAndTimezone(dateTime string, timezone TimeZoneString) (int64, error) {

	// Parse the datetime string
	t, err := time.Parse(DATETIME_FORMAT_DB, dateTime)
	if err != nil {
		return 0, err
	}

	// Set the timezone
	loc := GetTimeLocationFor(timezone)
	t = t.In(loc)

	// Get the Unix timestamp
	timestamp := t.Unix()

	return timestamp, nil
}

// TimeNowZ Return current time in UTC. Should be used everywhere to avoid local timezone.
func TimeNowZ() time.Time {
	return time.Now().UTC()
}

// Returns Default time '1971-01-01 00:00:00.000000'
func DefaultTime() time.Time {
	defaultTime := time.Date(1971, 1, 1, 0, 0, 0, 0, time.UTC)
	return defaultTime
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

func FindOffsetInUTC(timezone TimeZoneString) int {
	timezoneLocation, _ := time.LoadLocation(string(timezone))
	_, offset := time.Now().UTC().In(timezoneLocation).Zone()
	return offset
}

func FindOffsetInUTCForTimestamp(timezone TimeZoneString, timestamp int64) int {
	timezoneLocation, _ := time.LoadLocation(string(timezone))
	_, offset := time.Unix(timestamp, 0).UTC().In(timezoneLocation).Zone()
	return offset
}

// TimeNowUnix Returns current epoch time.
func TimeNowUnix() int64 {
	return TimeNowZ().Unix()
}

func GetCurrentMonthYear(timeZone TimeZoneString) string {

	return TimeNowIn(timeZone).Format("January2006")

}

func GetPreviousMonthYear(timeZone TimeZoneString) string {

	return TimeNowIn(timeZone).AddDate(0, -1, 0).Format("January2006")

}

func GetCurrentDayTimestamp() int64 {
	currentTime := time.Now()
	return time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 0, 0, 0, 0, currentTime.Location()).Unix()
}

// GetBeginningoftheDayEpochForDateAndTimezone provides the timestamp for the 00:00:00 hr of the given date and timezone
func GetBeginningoftheDayEpochForDateAndTimezone(dateStr string, timezone string) int64 {
	date, _ := time.Parse("20060102", dateStr)
	loc, _ := time.LoadLocation(timezone)
	t := time.Date(date.Year(), date.Month(), date.Day(), date.Hour(), date.Minute(), date.Second(), 0, loc)
	// Get epoch timestamp
	timestamp := t.Unix()
	return timestamp
}

// GetEndoftheDayEpochForDateAndTimezone provides the timestamp for the 23:59:59 hr of the given date and timezone
func GetEndoftheDayEpochForDateAndTimezone(dateStr string, timezone string) int64 {
	date, _ := time.Parse("20060102", dateStr)
	loc, _ := time.LoadLocation(timezone)
	// Set time
	hours := 23
	minutes := 59
	seconds := 59

	t := time.Date(date.Year(), date.Month(), date.Day(), hours, minutes, seconds, 0, loc)
	// Get epoch timestamp
	timestamp := t.Unix()
	return timestamp
}

func IsTimestampToday(timestamp int64) bool {
	return GetBeginningOfDayTimestamp(timestamp) == GetCurrentDayTimestamp()
}

func GetBeginningOfDayTimestamp(timestamp int64) int64 {
	t := time.Unix(timestamp, 0)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).Unix()
}

func GetBeginningOfDayTimestampZ(timestamp int64) int64 {
	t := time.Unix(timestamp, 0).UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).Unix()
}

// GetBeginningOfDayTimestamp Get's beginning of the day timestamp in given timezone.
func GetBeginningOfDayTimestampIn(timestamp int64, timezoneString TimeZoneString) int64 {
	return GetBeginningOfDayTimeZ(timestamp, timezoneString).Unix()
}

func GetBeginningOfDayTimeZ(timestamp int64, timezoneString TimeZoneString) time.Time {
	location, _ := time.LoadLocation(string(timezoneString))
	t := time.Unix(timestamp, 0).In(location)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// GetDateAsStringIn gets date in given timezone. Taking
func GetDateAsStringIn(timestamp int64, timezoneString TimeZoneString) int64 {
	if len(timezoneString) < 1 {
		timezoneString = TimeZoneStringIST
	}
	location, _ := time.LoadLocation(string(timezoneString))
	t := time.Unix(timestamp, 0).In(location)
	value, _ := strconv.ParseInt(t.Format(DATETIME_FORMAT_YYYYMMDD), 10, 64)
	return value
}

// To populate error instead.
func GetDateBeforeXPeriod(lastX int64, granularity string, timezoneString TimeZoneString) int64 {
	if granularity == GranularityDays {
		return getDateBeforeXDays(lastX, timezoneString)
	}
	return 0
}

func getDateBeforeXDays(numberOfDays int64, timezoneString TimeZoneString) int64 {
	todayBeginningTimestamp := GetBeginningOfTodayTime(timezoneString)
	return todayBeginningTimestamp.AddDate(0, 0, int(-1*numberOfDays)).Unix()
}

func DateAsFormattedInt(dateTime time.Time) uint64 {
	datetimeInt, _ := strconv.ParseInt(fmt.Sprintf("%d%02d%02d%02d", dateTime.Year(), int(dateTime.Month()), dateTime.Day(), dateTime.Hour()), 10, 64)
	return uint64(datetimeInt)
}

func DateAsYYYYMMDDFormat(dateTime time.Time) uint64 {

	datetimeInt, err := strconv.ParseInt(fmt.Sprintf("%d%02d%02d", dateTime.Year(), int(dateTime.Month()), dateTime.Day()), 10, 64)
	if err != nil {
		log.Warn("Error getting current Date As YYYYMMDD Format")
	}
	return uint64(datetimeInt)
}

// IsStartOfTodaysRangeIn Checks if the from value is start of todays range.
func IsStartOfTodaysRangeIn(from int64, timezoneString TimeZoneString) bool {
	return from == GetBeginningOfDayTimestampIn(TimeNowUnix(), timezoneString)
}

// Is30MinutesTimeRange Whether time range is of last 30 minutes.
func Is30MinutesTimeRange(from, to int64) bool {
	return (to - from) == 30*60
}

func IsLessThanTimeRange(from, to int64, maxTimeDuration int64) bool {
	return (to - from) < maxTimeDuration
}

func IsGreaterThanEqualTimeRange(from, to int64, minTimeDuration int64) bool {
	return (to - from) >= minTimeDuration
}

// GetEndOfDayTimestampIn Get's end of the day timestamp in given timezone.
func GetEndOfDayTimestampIn(timestamp int64, timezoneString TimeZoneString) int64 {
	location, _ := time.LoadLocation(string(timezoneString))
	t := time.Unix(timestamp, 0).In(location)
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), t.Location()).Unix()
}

// GetQueryRangePresetCurrentWeekIn Returns start and end unix timestamp for current week (Sunday to Yesterday)
func GetQueryRangePresetCurrentWeekIn(timeZoneString TimeZoneString) (int64, int64, error) {
	location, errCode := time.LoadLocation(string(timeZoneString))
	if errCode != nil {
		return 0, 0, errCode
	}
	if isTodayFirstDayOfWeekIn(location) {
		return GetQueryRangePresetTodayIn(timeZoneString)
	}
	rangeStartTime := GetBeginningOfDayTimestampIn(BeginningOfWeekStartOfDayIn(location).Unix(), timeZoneString)
	timeNow := time.Now().In(location)
	rangeEndTime := GetBeginningOfDayTimestampIn(timeNow.Unix(), timeZoneString) - 1
	return rangeStartTime, rangeEndTime, errCode
}

func GetPresetNameByFromAndTo(from, to int64, timezoneString TimeZoneString) string {

	epsilonSeconds := int64(60)
	for rangeString, rangeFunction := range QueryDateRangePresets {
		f, t, errCode := rangeFunction(timezoneString)
		if errCode != nil {
			return ""
		}
		// If difference between from-to of query and from-to of the preset are is < 60, that is the preset!
		if to <= t+epsilonSeconds && from >= f-epsilonSeconds {
			return rangeString
		}
	}
	// In case no presets matched
	return ""
}

func GetDynamicPreviousRanges(granularity string, number int64, timeZoneString TimeZoneString) (int64, int64, error) {
	location, errCode := time.LoadLocation(string(timeZoneString))
	if errCode != nil {
		return 0, 0, errCode
	}
	var referenceTime time.Time
	var startDate time.Time
	switch granularity {
	case GranularityDays:
		startDate, referenceTime = GetRangeForXDaysToCurrentDayStart(number, location)
	case GranularityWeek:
		startDate, referenceTime = GetRangeForXWeeksToCurrentWeekStart(number, location)
	case GranularityMonth:
		startDate, referenceTime = GetRangeForXMonthsToCurrentMonthStart(number, location)
	case GranularityQuarter:
		startDate, referenceTime = GetRangeForXQuartersToCurrentQuarterStart(number, location)
	default:
		return 0, 0, errors.New("invalid granularity")
	}
	startTimestamp := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location()).Unix()
	rangeStartTime := GetBeginningOfDayTimestampIn(startTimestamp, timeZoneString)
	rangeEndTime := GetBeginningOfDayTimestampIn(referenceTime.Unix(), timeZoneString) - 1
	return rangeStartTime, rangeEndTime, errCode
}
func GetRangeForXDaysToCurrentDayStart(number int64, location *time.Location) (time.Time, time.Time) {
	t := time.Now().In(location)
	startDate := t.AddDate(0, 0, -(int(number)))
	return startDate, t
}
func GetRangeForXWeeksToCurrentWeekStart(number int64, location *time.Location) (time.Time, time.Time) {
	beginningOfCurrentWeek := BeginningOfWeekStartOfDayIn(location)
	startDate := beginningOfCurrentWeek.AddDate(0, 0, -(int(number * 7)))
	return startDate, beginningOfCurrentWeek
}
func GetRangeForXMonthsToCurrentMonthStart(number int64, location *time.Location) (time.Time, time.Time) {
	beginningOfMonth := BeginningOfMonthStartOfDayIn(location)
	startDate := beginningOfMonth.AddDate(0, -(int(number)), 0)
	return startDate, beginningOfMonth
}
func GetRangeForXQuartersToCurrentQuarterStart(number int64, location *time.Location) (time.Time, time.Time) {
	beginningOfQuarter := BeginningOfQuarterStartOfDayIn(location)
	startDate := beginningOfQuarter.AddDate(0, -(int(number * 3)), 0)
	return startDate, beginningOfQuarter
}
func GetDynamicRangesForCurrentBasedOnGranularity(granularity string, timezoneString TimeZoneString) (int64, int64, error) {
	switch granularity {
	case GranularityWeek:
		return GetCurrentWeekRange(timezoneString)
	case GranularityMonth:
		return GetCurrentMonthRange(timezoneString)
	case GranularityQuarter:
		return GetCurrentQuarterRange(timezoneString)
	}
	return 0, 0, errors.New("invalid granularity")
}

// this includes today
func GetCurrentWeekRange(timeZoneString TimeZoneString) (int64, int64, error) {
	location, errCode := time.LoadLocation(string(timeZoneString))
	if errCode != nil {
		return 0, 0, errCode
	}
	if isTodayFirstDayOfWeekIn(location) {
		return GetQueryRangePresetTodayIn(timeZoneString)
	}
	rangeStartTime := GetBeginningOfDayTimestampIn(BeginningOfWeekStartOfDayIn(location).Unix(), timeZoneString)
	rangeEndTime := time.Now().In(location).Unix()
	return rangeStartTime, rangeEndTime, errCode
}

func isTodayFirstDayOfMonthIn(location *time.Location) bool {
	t := time.Now().In(location)
	return t.Day() == 1
}

func isTodayFirstDayOfWeekIn(location *time.Location) bool {
	t := time.Now().In(location)
	return int(t.Weekday()) == 0
}

// GetQueryRangePresetCurrentMonthIn Returns start and end unix timestamp for current month (1st of month to Yesterday)
func GetQueryRangePresetCurrentMonthIn(timezoneString TimeZoneString) (int64, int64, error) {
	location, errCode := time.LoadLocation(string(timezoneString))
	if errCode != nil {
		return 0, 0, errCode
	}
	if isTodayFirstDayOfMonthIn(location) {
		return GetQueryRangePresetTodayIn(timezoneString)
	}
	rangeStartTime := GetBeginningOfDayTimestampIn(BeginningOfMonthStartOfDayIn(location).Unix(), timezoneString)
	timeNow := time.Now().In(location)
	rangeEndTime := GetBeginningOfDayTimestampIn(timeNow.Unix(), timezoneString) - 1
	return rangeStartTime, rangeEndTime, errCode
}

// includes today
func GetCurrentMonthRange(timezoneString TimeZoneString) (int64, int64, error) {
	location, errCode := time.LoadLocation(string(timezoneString))
	if errCode != nil {
		return 0, 0, errCode
	}
	if isTodayFirstDayOfMonthIn(location) {
		return GetQueryRangePresetTodayIn(timezoneString)
	}
	rangeStartTime := GetBeginningOfDayTimestampIn(BeginningOfMonthStartOfDayIn(location).Unix(), timezoneString)
	rangeEndTime := time.Now().In(location).Unix()
	return rangeStartTime, rangeEndTime, errCode
}

// includes today
func GetCurrentQuarterRange(timezoneString TimeZoneString) (int64, int64, error) {
	location, errCode := time.LoadLocation(string(timezoneString))
	if errCode != nil {
		return 0, 0, errCode
	}
	rangeStartTime := GetBeginningOfDayTimestampIn(BeginningOfQuarterStartOfDayIn(location).Unix(), timezoneString)
	rangeEndTime := time.Now().In(location).Unix()
	return rangeStartTime, rangeEndTime, errCode
}

// GetQueryRangePresetLastWeekIn Returns start and end unix timestamp for last week (Last to Last Sunday to Last Sat)
func GetQueryRangePresetLastWeekIn(timezoneString TimeZoneString) (int64, int64, error) {
	location, errCode := time.LoadLocation(string(timezoneString))
	if errCode != nil {
		return 0, 0, errCode
	}
	rangeStartTime := GetBeginningOfDayTimestampIn(BeginningOfLastWeekStartOfDayIn(location).Unix(), timezoneString)
	rangeEndTime := GetEndOfDayTimestampIn(EndOfLastWeekStartOfDayIn(location).Unix(), timezoneString)
	return rangeStartTime, rangeEndTime, errCode
}

// GetQueryRangePresetLastMonthIn Returns start and end unix timestamp for last month (1st of the last month to end of the last month)
func GetQueryRangePresetLastMonthIn(timezoneString TimeZoneString) (int64, int64, error) {
	location, errCode := time.LoadLocation(string(timezoneString))
	if errCode != nil {
		return 0, 0, errCode
	}
	rangeStartTime := GetBeginningOfDayTimestampIn(BeginningOfLastMonthStartOfDayIn(location).Unix(), timezoneString)
	rangeEndTime := GetEndOfDayTimestampIn(EndOfLastMonthStartOfDayIn(location).Unix(), timezoneString)
	return rangeStartTime, rangeEndTime, nil
}

// GetMonthlyQueryRangesTuplesZ Return a tuple of start and end unix timestamp for last `numMonths`.
func GetMonthlyQueryRangesTuplesZ(numMonths int, timezoneString TimeZoneString) []Int64Tuple {
	monthlyRanges := make([]Int64Tuple, 0, 0)
	if numMonths == 0 {
		return monthlyRanges
	}
	lastMonthStart, lastMonthEnd, _ := GetQueryRangePresetLastMonthIn(timezoneString)
	location, _ := time.LoadLocation(string(timezoneString))
	monthlyRanges = append(monthlyRanges, Int64Tuple{lastMonthStart, lastMonthEnd})
	for i := 1; i < numMonths; i++ {
		lastMonthEnd = lastMonthStart - 1
		lastMonthEndTime := time.Unix(lastMonthEnd, 0).In(location)
		lastMonthStart = time.Date(lastMonthEndTime.Year(), lastMonthEndTime.Month(),
			1, 0, 0, 0, 0, lastMonthEndTime.Location()).Unix()
		monthlyRanges = append(monthlyRanges, Int64Tuple{lastMonthStart, lastMonthEnd})
	}

	return monthlyRanges
}

// BeginningOfWeekStartOfDayIn beginning of current week
func BeginningOfWeekStartOfDayIn(location *time.Location) time.Time {
	t := time.Now().In(location)
	weekday := int(t.Weekday())
	t = t.AddDate(0, 0, -weekday)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// BeginningOfLastWeekStartOfDayIn beginning of last week
func BeginningOfLastWeekStartOfDayIn(location *time.Location) time.Time {
	t := time.Now().In(location)
	t = t.AddDate(0, 0, -(int(t.Weekday()) + 7))
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// EndOfLastWeekStartOfDayIn end of last week
func EndOfLastWeekStartOfDayIn(location *time.Location) time.Time {
	t := BeginningOfWeekStartOfDayIn(location)
	t = t.AddDate(0, 0, -1)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// BeginningOfMonthStartOfDayIn beginning of current month
func BeginningOfMonthStartOfDayIn(location *time.Location) time.Time {
	t := time.Now().In(location)
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}
func BeginningOfQuarterStartOfDayIn(location *time.Location) time.Time {
	t := time.Now().In(location)
	currentMonth := int(t.Month())
	monthToSubtract := (currentMonth - 1) % 3
	startDate := t.AddDate(0, -(monthToSubtract), 0)
	return time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, startDate.Location())
}

// BeginningOfLastMonthStartOfDayIn beginning of last month
func BeginningOfLastMonthStartOfDayIn(location *time.Location) time.Time {
	t := BeginningOfMonthStartOfDayIn(location)
	monthEnd := t.AddDate(0, 0, -(1))
	return time.Date(monthEnd.Year(), monthEnd.Month(), 1, 0, 0, 0, 0, monthEnd.Location())
}

// EndOfLastMonthStartOfDayIn end of last month
func EndOfLastMonthStartOfDayIn(location *time.Location) time.Time {
	t := BeginningOfMonthStartOfDayIn(location)
	return t.AddDate(0, 0, -(1))
}

// GetQueryRangePresetYesterdayIn Returns start and end unix timestamp for yesterday's range.
func GetQueryRangePresetYesterdayIn(timezoneString TimeZoneString) (int64, int64, error) {
	location, errCode := time.LoadLocation(string(timezoneString))
	if errCode != nil {
		return 0, 0, errCode
	}
	timeNow := time.Now().In(location)
	rangeEndTime := GetBeginningOfDayTimestampIn(timeNow.Unix(), timezoneString) - 1
	rangeStartTime := GetBeginningOfDayTimestampIn(timeNow.AddDate(0, 0, -1).Unix(), timezoneString)
	return rangeStartTime, rangeEndTime, nil
}

// GetQueryRangePresetDayBeforeYesterdayIn Returns start and end unix timestamp for day before yesterday's range.
func GetQueryRangePresetDayBeforeYesterdayIn(timezoneString TimeZoneString) (int64, int64, error) {
	location, errCode := time.LoadLocation(string(timezoneString))
	if errCode != nil {
		return 0, 0, errCode
	}
	timeNow := time.Now().In(location)
	rangeStartTime := GetBeginningOfDayTimestampIn(timeNow.AddDate(0, 0, -2).Unix(), timezoneString)
	rangeEndTime := GetBeginningOfDayTimestampIn(timeNow.AddDate(0, 0, -1).Unix(), timezoneString) - 1
	return rangeStartTime, rangeEndTime, nil
}

// GetQueryRangePresetTodayIn Returns start and end unix timestamp for today's range.
func GetQueryRangePresetTodayIn(timezoneString TimeZoneString) (int64, int64, error) {
	location, errCode := time.LoadLocation(string(timezoneString))
	if errCode != nil {
		return 0, 0, errCode
	}
	timeNow := time.Now().In(location)
	rangeStartTime := GetBeginningOfDayTimestampIn(timeNow.Unix(), timezoneString)
	return rangeStartTime, timeNow.Unix(), nil
}

func GetBeginningOfTodayTime(timezoneString TimeZoneString) time.Time {
	location, _ := time.LoadLocation(string(timezoneString))
	timeNow := time.Now().In(location)
	return GetBeginningOfDayTimeZ(timeNow.Unix(), timezoneString)
}

// GetQueryRangePresetLast30MinutesIn Returns start and end unix timestamp for last 30mins range.
func GetQueryRangePresetLast30MinutesIn(timezoneString TimeZoneString) (int64, int64, error) {
	location, errCode := time.LoadLocation(string(timezoneString))
	if errCode != nil {
		return 0, 0, errCode
	}
	timeNow := time.Now().In(location)
	endTime := timeNow.Unix()
	startTime := endTime - 30*60
	return startTime, endTime, nil
}

func GetBatchRangeFromStartAndEndTimestamp(startTimestamp, endTimestamp, batchRangeInSeconds int64) [][]int64 {
	batchedTimestamp := make([][]int64, 0, 0)
	for i := startTimestamp; i < endTimestamp; {
		nextTimestamp := i + batchRangeInSeconds
		if nextTimestamp > endTimestamp {
			nextTimestamp = endTimestamp
		}

		batchedTimestamp = append(batchedTimestamp, []int64{i, nextTimestamp})
		i = nextTimestamp
	}

	return batchedTimestamp
}

// format: YYYYMMDD
func GetBeginningDayTimestampFromDateString(date string) (int64, error) {
	if len(date) != 8 {
		return 0, fmt.Errorf("wrong datestring provided: %s", date)
	}
	startYear, err := strconv.ParseInt(date[0:4], 10, 64)
	if err != nil {
		return 0, err
	}
	startMonth, err := strconv.ParseInt(date[4:6], 10, 64)
	if err != nil {
		return 0, err
	}
	startDay, err := strconv.ParseInt(date[6:8], 10, 64)
	if err != nil {
		return 0, err
	}
	return time.Date(int(startYear), time.Month(startMonth), int(startDay), 0, 0, 0, 0, time.FixedZone("UTC", 0)).Unix(), nil
}

func GetTimestampInSecs(timestamp int) int {
	digitsCount := 0
	{
		timestampTmp := timestamp
		for timestampTmp > 0 {
			timestampTmp = timestampTmp / 10
			digitsCount++
		}
	}
	reqTimestamp := int(float64(timestamp) * math.Pow(10, float64(10-digitsCount)))
	return reqTimestamp
}

func startDateOfMonth(year int, month time.Month) time.Time {
	//timezoneLocation, _ := time.LoadLocation("Asia/Kolkata")
	return time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
}

func lastDateOfMonth(year int, month time.Month) time.Time {
	// timezoneLocation, _ := time.LoadLocation("Asia/Kolkata")
	return time.Date(year, month+1, 0, 0, 0, -1, 0, time.UTC)
}

type CustomPreset struct {
	From  int64
	To    int64
	Year  int
	Month string
	Day   int
	TZ    TimeZoneString
}

func GetAllDaysSinceStartTime(startTimestamp int64, zoneString TimeZoneString) []CustomPreset {
	var result []CustomPreset

	// Parse the timezone
	loc, err := time.LoadLocation(string(zoneString))
	if err != nil {
		panic(err) // Handle the error appropriately
	}

	// Start from the given startTime
	currentTime := time.Unix(startTimestamp, 0).In(loc)
	yesterday := time.Now().In(loc).AddDate(0, 0, -1) //Yesterday

	// Loop through each day until yesterday
	for !currentTime.After(yesterday) {
		year, month, day := currentTime.Date()

		// Calculate the start and end timestamps for the day
		startOfDay := time.Date(year, month, day, 0, 0, 0, 0, loc)
		endOfDay := time.Date(year, month, day, 23, 59, 59, 999999999, loc)

		// Convert to Unix timestamps
		startUnix := startOfDay.Unix()
		endUnix := endOfDay.Unix()

		// Create a CustomPreset for the current day
		preset := CustomPreset{
			From:  startUnix,
			To:    endUnix,
			Year:  year,
			Month: month.String(),
			Day:   day,
			TZ:    zoneString,
		}

		// Append the CustomPreset to the result
		result = append(result, preset)

		// Move to the next day
		currentTime = currentTime.Add(24 * time.Hour)
	}
	return result
}

func GetAllDaysTillDayB4YesterdaySinceStartTime(startTimestamp int64, zoneString TimeZoneString) []CustomPreset {
	var result []CustomPreset

	// Parse the timezone
	loc, err := time.LoadLocation(string(zoneString))
	if err != nil {
		panic(err) // Handle the error appropriately
	}

	// Start from the given startTime
	currentTime := time.Unix(startTimestamp, 0).In(loc)
	dayBeforeYesterday := time.Now().In(loc).AddDate(0, 0, -2) // n-2 day or DayBeforeYesterday

	// Loop through each day until dayBeforeYesterday
	for !currentTime.After(dayBeforeYesterday) {
		year, month, day := currentTime.Date()

		// Calculate the start and end timestamps for the day
		startOfDay := time.Date(year, month, day, 0, 0, 0, 0, loc)
		endOfDay := time.Date(year, month, day, 23, 59, 59, 999999999, loc)

		// Convert to Unix timestamps
		startUnix := startOfDay.Unix()
		endUnix := endOfDay.Unix()

		// Create a CustomPreset for the current day
		preset := CustomPreset{
			From:  startUnix,
			To:    endUnix,
			Year:  year,
			Month: month.String(),
			Day:   day,
			TZ:    zoneString,
		}

		// Append the CustomPreset to the result
		result = append(result, preset)

		// Move to the next day
		currentTime = currentTime.Add(24 * time.Hour)
	}
	return result
}

func GetAllMonthFromTo(startTime int64, zoneString TimeZoneString) []CustomPreset {

	currentTime := time.Now().Unix()
	startOfMonth := startTime
	endOfMonth := startTime
	var allRanges []CustomPreset

	for _, baseYear := range BaseYears {

		for y, m := baseYear, time.Month(1); m <= 12; m++ {
			startOfMonth = GetBeginningOfDayTimestampIn(startDateOfMonth(y, m).Unix(), zoneString)
			endOfMonth = GetEndOfDayTimestampIn(lastDateOfMonth(y, m).Unix(), zoneString)

			// skip month if it is before start time
			if endOfMonth < startTime {
				continue
			}
			// exit when the end of last month/current month is reached
			if endOfMonth > currentTime {
				break
			}

			newRange := CustomPreset{
				From:  startOfMonth,
				To:    endOfMonth,
				Year:  time.Unix(startOfMonth, 0).Year(),
				Month: time.Unix(startOfMonth, 0).Month().String(),
				Day:   time.Unix(startOfMonth, 0).Day(),
				TZ:    zoneString,
			}
			allRanges = append(allRanges, newRange)
		}
		// exit when the end of last month/current month is reached
		if endOfMonth > currentTime {
			break
		}
	}
	return allRanges
}

func GetAllWeeksFromStartTime(startTime int64, zoneString TimeZoneString) []CustomPreset {
	currentTime := time.Now().Unix()

	fromNew := startTime
	toNew := fromNew + (7 * 24 * 60 * 60) - 1
	var allRanges []CustomPreset
	for currentTime > toNew {
		newRange := CustomPreset{
			From:  fromNew,
			To:    toNew,
			Year:  time.Unix(fromNew, 0).Year(),
			Month: time.Unix(fromNew, 0).Month().String(),
			Day:   time.Unix(fromNew, 0).Day(),
			TZ:    zoneString,
		}
		allRanges = append(allRanges, newRange)
		fromNew = toNew + 1
		toNew = fromNew + (7 * 24 * 60 * 60) - 1
	}
	return allRanges
}

func GetAllDaysAsTimestamp(fromUnix int64, toUnix int64, timezone string) ([]time.Time, []string) {
	rTimestamps := make([]time.Time, 0, 0)
	rTimezoneOffsets := make([]string, 0, 0)

	from, err := GetTimeFromUnixTimestampWithZone(fromUnix, timezone)
	if err != nil {
		return rTimestamps, rTimezoneOffsets
	}
	from = now.New(from).BeginningOfDay()

	to, err := GetTimeFromUnixTimestampWithZone(toUnix, timezone)
	if err != nil {
		return rTimestamps, rTimezoneOffsets
	}
	to = now.New(to).BeginningOfDay()

	toStr := GetTimestampAsStrWithTimezone(to, timezone)
	for t, tStr := from, ""; tStr != toStr; {
		offset := GetTimezoneOffsetFromString(t, timezone)
		currTimeStrFromOffset := GetTimestampAsStrWithTimezoneGivenOffset(t, offset)
		currTimeFromOffset := GetTimeFromParseTimeStr(currTimeStrFromOffset)

		tStr = GetTimestampAsStrWithTimezone(currTimeFromOffset, timezone)
		rTimestamps = append(rTimestamps, currTimeFromOffset)
		rTimezoneOffsets = append(rTimezoneOffsets, offset)

		t = t.AddDate(0, 0, 1) // next day.
		t, _ = GetTimeFromUnixTimestampWithZone(t.Unix(), timezone)
		t = now.New(t).BeginningOfDay()
	}

	return rTimestamps, rTimezoneOffsets
}

func GetAllDaysBetweenFromTo(timezoneString TimeZoneString, effectiveFrom, effectiveTo int64) (bool, []TimestampRange) {
	var ranges []TimestampRange

	// Parse the timezone
	loc, err := time.LoadLocation(string(timezoneString))
	if err != nil {
		fmt.Println(err)
		return false, ranges
	}

	// Start from the effectiveFrom date
	currentTime := time.Unix(effectiveFrom, 0).In(loc)
	endTime := time.Unix(effectiveTo, 0).In(loc)

	// Loop through each day until the effectiveTo date
	for !currentTime.After(endTime) {
		year, month, day := currentTime.Date()

		// Calculate the start and end timestamps for the day
		startOfDay := time.Date(year, month, day, 0, 0, 0, 0, loc)
		endOfDay := time.Date(year, month, day, 23, 59, 59, 999999999, loc)

		// Convert to Unix timestamps
		startUnix := startOfDay.Unix()
		endUnix := endOfDay.Unix()

		// Create a TimestampRange for the current day
		rangeOfDay := TimestampRange{
			Start: startUnix,
			End:   endUnix,
		}

		// Append the TimestampRange to the result
		ranges = append(ranges, rangeOfDay)

		// Move to the next day
		currentTime = currentTime.Add(24 * time.Hour)
	}

	return true, ranges
}

// SanitizeWeekStart Gives the start of the week for any timestamp in the week for given timeZone
func SanitizeWeekStart(startTime int64, zoneString TimeZoneString) int64 {

	unixTimeUTC := time.Unix(startTime, 0) // gives unix time stamp in UTC
	location, errCode := time.LoadLocation(string(zoneString))
	if errCode != nil {
		return 0
	}
	unixTimeInGivenTimeZone := unixTimeUTC.In(location)
	for unixTimeInGivenTimeZone.Weekday() != 0 {
		if unixTimeInGivenTimeZone.Weekday() > 0 {
			unixTimeInGivenTimeZone = unixTimeInGivenTimeZone.AddDate(0, 0, -1)
		}
	}
	weekStart := GetBeginningOfDayTimestampIn(unixTimeInGivenTimeZone.Unix(), zoneString)
	return weekStart
}

// GenerateLast12MonthsTimestamps returns start-end of last 12 months in descending order (last month first) from now
func GenerateLast12MonthsTimestamps(timezone string) []TimestampRange {
	timestamps := make([]TimestampRange, 0)

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		fmt.Println("Invalid timezone:", timezone)
		return timestamps
	}

	now := time.Now().In(loc)
	year, month, _ := now.Date()

	for i := 0; i < 12; i++ {
		endOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, loc)
		startOfMonth := endOfMonth.AddDate(0, -1, 0)

		timestamps = append(timestamps, TimestampRange{
			Start: startOfMonth.Unix(),
			End:   endOfMonth.Unix() - 1,
		})

		year, month, _ = startOfMonth.Date()
	}

	return timestamps
}

func GenerateWeeksBetween(start, end int64) []TimestampRange {
	weeks := make([]TimestampRange, 0)

	// Start at the beginning of the first week
	weekStart := start
	weekEnd := weekStart + 7*86400 - 1

	for weekEnd <= end {
		weeks = append(weeks, TimestampRange{Start: weekStart, End: weekEnd})

		// Move to the next week
		weekStart = weekEnd + 1
		weekEnd = weekEnd + 7*86400 - 1
	}
	return weeks
}

func GetAllValidRangesInBetween(from, to int64, ranges []TimestampRange) []TimestampRange {
	timestamps := make([]TimestampRange, 0)

	for _, rng := range ranges {
		if rng.Start >= from && rng.End <= to {
			timestamps = append(timestamps, rng)
		}
	}
	return timestamps
}

func IsAMonthlyRangeQuery(timezoneString TimeZoneString, effectiveFrom, effectiveTo int64) (bool, []TimestampRange) {

	last12Months := GenerateLast12MonthsTimestamps(string(timezoneString))
	stMatch := 0
	enMatch := 0
	for _, rng := range last12Months {
		if rng.Start == effectiveFrom {
			stMatch = 1
		}
		if rng.End == effectiveTo {
			enMatch = 1
		}
	}
	// both start and end timestamp belong to a month
	if stMatch == 1 && enMatch == 1 {
		return true, last12Months
	}
	return false, last12Months
}

func IsAWeeklyRangeQuery(timezoneString TimeZoneString, effectiveFrom, effectiveTo int64) (bool, []TimestampRange) {

	last48Weeks := GenerateLast48WeeksTimestamps(string(timezoneString))
	stMatch := 0
	enMatch := 0
	for _, rng := range last48Weeks {
		if rng.Start == effectiveFrom {
			stMatch = 1
		}
		if rng.End == effectiveTo {
			enMatch = 1
		}
	}
	// both start and end timestamp belong to a week
	if stMatch == 1 && enMatch == 1 {
		return true, last48Weeks
	}
	return false, last48Weeks
}

// GenerateLast48WeeksTimestamps returns start-end of last 48 weeks in descending order (last week first) from now
func GenerateLast48WeeksTimestamps(timezone string) []TimestampRange {
	timestamps := make([]TimestampRange, 0)

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		fmt.Println("Invalid timezone:", timezone)
		return timestamps
	}

	now := time.Now().In(loc)

	// Set the weekday to Sunday (0) and adjust the time to 00:00:00
	weekStart := now.AddDate(0, 0, -(int(now.Weekday())))
	weekStart = time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, loc)

	for i := 0; i < 48; i++ {
		// Get the end of the week by adding 6 days, 23 hours, 59 minutes, and 59 seconds
		weekEnd := weekStart.AddDate(0, 0, 6).Add(time.Hour*23 + time.Minute*59 + time.Second*59)

		// add all the valid ranges which are smaller than today's time
		if weekStart.Unix() < now.Unix() && weekEnd.Unix() < now.Unix() {
			timestamps = append(timestamps, TimestampRange{
				Start: weekStart.Unix(),
				End:   weekEnd.Unix(),
			})
		}
		// Move to the previous week's start
		weekStart = weekStart.AddDate(0, 0, -7)
	}

	return timestamps
}
