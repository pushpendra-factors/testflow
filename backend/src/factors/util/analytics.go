package util

import (
	"fmt"
	"time"
)

// Keeps date and reset others.
func getDateOnlyTimestamp(t time.Time) time.Time {
	ts, _ := time.Parse(time.RFC3339, GetDateOnlyTimestampStr(t))
	return ts
}

func GetDateOnlyTimestampStr(t time.Time) string {
	return fmt.Sprintf("%d-%02d-%02dT00:00:00Z", t.Year(), t.Month(), t.Day())
}

func GetAllDatesAsTimestamp(fromUnix int64, toUnix int64) []time.Time {
	from := getDateOnlyTimestamp(time.Unix(fromUnix, 0))
	to := getDateOnlyTimestamp(time.Unix(toUnix, 0))

	toStr := GetDateOnlyTimestampStr(to)
	rTimestamps := make([]time.Time, 0, 0)

	for t, tStr := from, ""; tStr != toStr; {
		tStr = GetDateOnlyTimestampStr(t)
		rTimestamps = append(rTimestamps, t)
		t = t.AddDate(0, 0, 1) // next day.
	}

	return rTimestamps
}
