package util

import (
	"fmt"
	"time"

	"github.com/jinzhu/now"
	log "github.com/sirupsen/logrus"
)

const TimestampWithTimezoneFormat = "2006-01-02T15:04:05-07:00"

func GetTimeFromTimestampStr(timestampStr string) time.Time {
	ts, _ := time.Parse(time.RFC3339, timestampStr)
	return ts
}

func getTimezoneOffsetFromString(currentTime time.Time, timezone string) string {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return "+00:00"
	}

	return currentTime.In(loc).Format("-07:00")
}

func GetTimestampAsStrWithTimezoneGivenOffset(t time.Time, offset string) string {
	return fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d%s", t.Year(),
		t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), offset)
}

// GetTimestampAsStrWithTimezone - Appends timezone doesn't converts.
func GetTimestampAsStrWithTimezone(t time.Time, timezone string) string {
	return fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d%s", t.Year(),
		t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), getTimezoneOffsetFromString(t, timezone))
}

// This relies on fact of above parse.
func GetTimeFromParseTimeStr(timestampStr string) time.Time {
	ts, _ := time.Parse(TimestampWithTimezoneFormat, timestampStr)
	return ts
}

func GetTimeFromParseTimeStrWithErrorFromInterface(timestamp interface{}) (time.Time, error) {
	sTime := fmt.Sprintf("%v", timestamp)
	ts, err := time.Parse(TimestampWithTimezoneFormat, sTime)
	return ts, err
}

func GetTimeFromUnixTimestampWithZone(unix int64, timezone string) (time.Time, error) {
	// if timezone is "", uses UTC.
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		log.WithError(err).WithField("timezone", timezone).WithField("unix_timestamp",
			unix).Error("invalid unix timestamp with timezone.")
		return time.Time{}, err
	}
	curr_time := time.Unix(unix, 0).UTC().In(loc)
	parsedTimestamp := GetTimestampAsStrWithTimezone(curr_time, timezone)
	return GetTimeFromTimestampStr(parsedTimestamp), nil

}

func GetAllDatesAndOffsetAsTimestamp(fromUnix int64, toUnix int64, timezone string) ([]time.Time, []string) {
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
		tStr = GetTimestampAsStrWithTimezone(t, timezone)
		rTimestamps = append(rTimestamps, t)
		offset := fmt.Sprintf("%s", getTimezoneOffsetFromString(t.Add(4*time.Hour), timezone)) // + 4 hours to accomodate day light saving at 2:00.
		rTimezoneOffsets = append(rTimezoneOffsets, offset)

		t = t.Add(28 * time.Hour)
		t, _ = GetTimeFromUnixTimestampWithZone(t.Unix(), timezone)
		t = now.New(t).BeginningOfDay()
	}

	return rTimestamps, rTimezoneOffsets
}

// GetAllQuartersAsTimestamp buckets the days into start of quarter
func GetAllQuartersAsTimestamp(fromUnix int64, toUnix int64, timezone string) ([]time.Time, []string) {
	rTimestamps := make([]time.Time, 0, 0)
	rTimezoneOffsets := make([]string, 0, 0)

	from, err := GetTimeFromUnixTimestampWithZone(fromUnix, timezone)
	if err != nil {
		return rTimestamps, rTimezoneOffsets
	}
	from = now.New(from).BeginningOfQuarter()

	to, err := GetTimeFromUnixTimestampWithZone(toUnix, timezone)
	if err != nil {
		return rTimestamps, rTimezoneOffsets
	}
	to = now.New(to).BeginningOfQuarter()

	toStr := GetTimestampAsStrWithTimezone(to, timezone)
	for t, tStr := from, ""; tStr != toStr; {
		tStr = GetTimestampAsStrWithTimezone(t, timezone)
		rTimestamps = append(rTimestamps, t)
		offset := getTimezoneOffsetFromString(t.Add(4*time.Hour), timezone)
		rTimezoneOffsets = append(rTimezoneOffsets, offset)

		// get the some date in next quarter, here it is 120+10 i.e. 10day in next quarter
		t = t.AddDate(0, 0, 130).Add(4 * time.Hour) // next quarter.
		t, _ = GetTimeFromUnixTimestampWithZone(t.Unix(), timezone)
		t = now.New(t).BeginningOfQuarter()
	}

	return rTimestamps, rTimezoneOffsets
}

// GetAllMonthsAsTimestamp buckets the days into start of weeks i.e.
// returns list of Sundays for from to to
func GetAllMonthsAsTimestamp(fromUnix int64, toUnix int64, timezone string) ([]time.Time, []string) {
	rTimestamps := make([]time.Time, 0, 0)
	rTimezoneOffsets := make([]string, 0, 0)

	from, err := GetTimeFromUnixTimestampWithZone(fromUnix, timezone)
	if err != nil {
		return rTimestamps, rTimezoneOffsets
	}
	from = now.New(from).BeginningOfMonth()

	to, err := GetTimeFromUnixTimestampWithZone(toUnix, timezone)
	if err != nil {
		return rTimestamps, rTimezoneOffsets
	}
	to = now.New(to).BeginningOfMonth()

	toStr := GetTimestampAsStrWithTimezone(to, timezone)
	for t, tStr := from, ""; tStr != toStr; {
		tStr = GetTimestampAsStrWithTimezone(t, timezone)
		rTimestamps = append(rTimestamps, t)
		offset := getTimezoneOffsetFromString(t.Add(4*time.Hour), timezone)
		rTimezoneOffsets = append(rTimezoneOffsets, offset)

		// get the some date in next month, here it can be 4th, 5th, or 6th
		t = t.AddDate(0, 0, 35).Add(4 * time.Hour) // next month.
		t, _ = GetTimeFromUnixTimestampWithZone(t.Unix(), timezone)
		t = now.New(t).BeginningOfMonth()
	}

	return rTimestamps, rTimezoneOffsets
}

// GetAllWeeksAsTimestamp buckets the days into start of weeks i.e.
// returns list of Sundays for from to to
func GetAllWeeksAsTimestamp(fromUnix int64, toUnix int64, timezone string) ([]time.Time, []string) {
	rTimestamps := make([]time.Time, 0, 0)
	rTimezoneOffsets := make([]string, 0, 0)

	from, err := GetTimeFromUnixTimestampWithZone(fromUnix, timezone)
	if err != nil {
		return rTimestamps, rTimezoneOffsets
	}
	from = now.New(from).BeginningOfWeek()

	to, err := GetTimeFromUnixTimestampWithZone(toUnix, timezone)
	if err != nil {
		return rTimestamps, rTimezoneOffsets
	}
	to = now.New(to).BeginningOfWeek()

	toStr := GetTimestampAsStrWithTimezone(to, timezone)
	for t, tStr := from, ""; tStr != toStr; {
		tStr = GetTimestampAsStrWithTimezone(t, timezone)
		rTimestamps = append(rTimestamps, t)
		offset := getTimezoneOffsetFromString(t.Add(4*time.Hour), timezone)
		rTimezoneOffsets = append(rTimezoneOffsets, offset)

		t = t.AddDate(0, 0, 7).Add(4 * time.Hour) // next week.
		t, _ = GetTimeFromUnixTimestampWithZone(t.Unix(), timezone)
		t = now.New(t).BeginningOfWeek()
	}

	return rTimestamps, rTimezoneOffsets
}

func GetAllHoursAsTimestamp(fromUnix int64, toUnix int64, timezone string) ([]time.Time, []string) {
	rTimestamps := make([]time.Time, 0, 0)
	rTimezoneOffsets := make([]string, 0, 0)

	from, err := GetTimeFromUnixTimestampWithZone(fromUnix, timezone)
	if err != nil {
		return rTimestamps, rTimezoneOffsets
	}
	from = now.New(from).BeginningOfHour()

	to, err := GetTimeFromUnixTimestampWithZone(toUnix, timezone)
	if err != nil {
		return rTimestamps, rTimezoneOffsets
	}
	to = now.New(to).BeginningOfHour()

	toStr := GetTimestampAsStrWithTimezone(to, timezone)
	for t, tStr := from, ""; tStr != toStr; {
		tStr = GetTimestampAsStrWithTimezone(t, timezone)
		rTimestamps = append(rTimestamps, t)
		offset := getTimezoneOffsetFromString(t, timezone)
		rTimezoneOffsets = append(rTimezoneOffsets, offset)

		t = t.Add(1 * time.Hour) // next hour.
		t, _ = GetTimeFromUnixTimestampWithZone(t.Unix(), timezone)
		t = now.New(t).BeginningOfHour()
	}

	return rTimestamps, rTimezoneOffsets
}
