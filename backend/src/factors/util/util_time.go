package util

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

// Datetime related utility functions.
// General convention for date Functions - suffix Z if utc based, In if timezone is passed, no suffix if localTime.
const (
	DATETIME_FORMAT_YYYYMMDD_HYPHEN string = "2006-01-02"
	DATETIME_FORMAT_YYYYMMDD        string = "20060102"
	DATETIME_FORMAT_DB              string = "2006-01-02 15:04:05"
)

// Returns date in YYYYMMDD format
func GetDateOnlyFromTimestampZ(timestamp int64) string {
	return time.Unix(timestamp, 0).UTC().Format(DATETIME_FORMAT_YYYYMMDD)
}

// TimeNowZ Return current time in UTC. Should be used everywhere to avoid local timezone.
func TimeNowZ() time.Time {
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

func FindOffsetInUTC(timezone TimeZoneString) int {
	timezoneLocation, _ := time.LoadLocation(string(timezone))
	_, offset := time.Now().UTC().In(timezoneLocation).Zone()
	return offset
}

// TimeNowUnix Returns current epoch time.
func TimeNowUnix() int64 {
	return TimeNowZ().Unix()
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

// IsStartOfTodaysRangeIn Checks if the from value is start of todays range.
func IsStartOfTodaysRangeIn(from int64, timezoneString TimeZoneString) bool {
	return from == GetBeginningOfDayTimestampIn(TimeNowUnix(), timezoneString)
}

// Is30MinutesTimeRange Whether time range is of last 30 minutes.
func Is30MinutesTimeRange(from, to int64) bool {
	return (to - from) == 30*60
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
