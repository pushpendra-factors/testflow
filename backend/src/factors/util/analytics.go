package util

import (
	"fmt"
	"time"

	"github.com/jinzhu/now"
	log "github.com/sirupsen/logrus"
)

func GetTimeFromTimestampStr(timestampStr string) time.Time {
	ts, _ := time.Parse(time.RFC3339, timestampStr)
	return ts
}

func getTimezoneOffsetFromString(timezone string) string {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return "+00:00"
	}

	return time.Now().In(loc).Format("-07:00")
}

// GetTimestampAsStrWithTimezone - Appends timezone doesn't converts.
func GetTimestampAsStrWithTimezone(t time.Time, timezone string) string {
	return fmt.Sprintf("%d-%02d-%02dT%02d:00:00%s", t.Year(),
		t.Month(), t.Day(), t.Hour(), getTimezoneOffsetFromString(timezone))
}

func getTimeFromUnixTimestampWithZone(unix int64, timezone string) (time.Time, error) {
	// if timezone is "", uses UTC.
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		log.WithError(err).WithField("timezone", timezone).WithField("unix_timestamp",
			unix).Error("invalid unix timestamp with timezone.")
		return time.Time{}, err
	}

	return time.Unix(unix, 0).UTC().In(loc), nil
}

func GetAllDatesAsTimestamp(fromUnix int64, toUnix int64, timezone string) []time.Time {
	rTimestamps := make([]time.Time, 0, 0)

	from, err := getTimeFromUnixTimestampWithZone(fromUnix, timezone)
	if err != nil {
		return rTimestamps
	}
	from = now.New(from).BeginningOfDay()

	to, err := getTimeFromUnixTimestampWithZone(toUnix, timezone)
	if err != nil {
		return rTimestamps
	}
	to = now.New(to).BeginningOfDay()

	toStr := GetTimestampAsStrWithTimezone(to, timezone)

	for t, tStr := from, ""; tStr != toStr; {
		tStr = GetTimestampAsStrWithTimezone(t, timezone)
		rTimestamps = append(rTimestamps, t)
		t = t.AddDate(0, 0, 1) // next day.
	}

	return rTimestamps
}

// GetAllMonthsAsTimestamp buckets the days into start of weeks i.e.
// returns list of Sundays for from to to
func GetAllMonthsAsTimestamp(fromUnix int64, toUnix int64, timezone string) []time.Time {
	rTimestamps := make([]time.Time, 0, 0)

	from, err := getTimeFromUnixTimestampWithZone(fromUnix, timezone)
	if err != nil {
		return rTimestamps
	}
	from = now.New(from).BeginningOfMonth()

	to, err := getTimeFromUnixTimestampWithZone(toUnix, timezone)
	if err != nil {
		return rTimestamps
	}
	to = now.New(to).BeginningOfMonth()

	toStr := GetTimestampAsStrWithTimezone(to, timezone)
	for t, tStr := from, ""; tStr != toStr; {
		tStr = GetTimestampAsStrWithTimezone(t, timezone)
		rTimestamps = append(rTimestamps, t)
		// get the some date in next month, here it can be 4th, 5th, or 6th
		t = t.AddDate(0, 0, 35) // next week.
		t = now.New(t).BeginningOfMonth()
	}

	return rTimestamps
}

// GetAllWeeksAsTimestamp buckets the days into start of weeks i.e.
// returns list of Sundays for from to to
func GetAllWeeksAsTimestamp(fromUnix int64, toUnix int64, timezone string) []time.Time {
	rTimestamps := make([]time.Time, 0, 0)

	from, err := getTimeFromUnixTimestampWithZone(fromUnix, timezone)
	if err != nil {
		return rTimestamps
	}
	from = now.New(from).BeginningOfWeek()

	to, err := getTimeFromUnixTimestampWithZone(toUnix, timezone)
	if err != nil {
		return rTimestamps
	}
	to = now.New(to).BeginningOfWeek()

	toStr := GetTimestampAsStrWithTimezone(to, timezone)
	for t, tStr := from, ""; tStr != toStr; {
		tStr = GetTimestampAsStrWithTimezone(t, timezone)
		rTimestamps = append(rTimestamps, t)
		t = t.AddDate(0, 0, 7) // next week.
	}

	return rTimestamps
}

func GetAllHoursAsTimestamp(fromUnix int64, toUnix int64, timezone string) []time.Time {
	rTimestamps := make([]time.Time, 0, 0)

	from, err := getTimeFromUnixTimestampWithZone(fromUnix, timezone)
	if err != nil {
		return rTimestamps
	}
	from = now.New(from).BeginningOfHour()

	to, err := getTimeFromUnixTimestampWithZone(toUnix, timezone)
	if err != nil {
		return rTimestamps
	}
	to = now.New(to).BeginningOfHour()

	toStr := GetTimestampAsStrWithTimezone(to, timezone)
	for t, tStr := from, ""; tStr != toStr; {
		tStr = GetTimestampAsStrWithTimezone(t, timezone)
		rTimestamps = append(rTimestamps, t)
		t = t.Add(1 * time.Hour) // next hour.
	}

	return rTimestamps
}
