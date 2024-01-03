package util

import (
	"time"
	// log "github.com/sirupsen/logrus"
)

func IsAllowedToRunInThisJob(from, to time.Time, jobForLargeTimerange bool) bool {
	numberOfDays := GetNumberOfDays(from, to)
	if (numberOfDays > 15) {
		return jobForLargeTimerange
	} else {
		return !jobForLargeTimerange
	}
	return false
}

func GetNumberOfDays(from, to time.Time) float64 {
	return to.Sub(from).Hours()/24
}

// TODO: Currently relies on today for comparision. When inputFrom and inputTo are given, we need to recheck this.
func GetSplitTimerangesIn7DayRanges(inputFrom, inputTo time.Time, timezone string, location *time.Location) ([][]time.Time) {

	splitTimeRanges := make([][]time.Time, 0)
	todayBeginning, _ := GetTodayBeginningTimestampAndNowInBuffer(timezone, location)

	currentFrom := inputFrom
	currentTo := getNextTimestamp7DaysEOD(currentFrom)

	for currentTo.Unix() < todayBeginning.Unix() {
		splitTimeRanges = append(splitTimeRanges, []time.Time{ currentFrom, currentTo })
		currentFrom = currentTo.Add(1 * time.Second)
		currentTo = getNextTimestamp7DaysEOD(currentFrom)
	}
	currentTo = inputTo
	splitTimeRanges = append(splitTimeRanges, []time.Time{ currentFrom, currentTo })
	return splitTimeRanges
}

func GetTodayBeginningTimestampAndNowInBuffer(timezone string, location *time.Location) (time.Time, time.Time) {
	nowTimestamp := time.Now().In(location)
	nowTimestampWithBuffer := nowTimestamp.Add(-3*time.Minute)
	return GetBeginningOfDayTimeZ(nowTimestampWithBuffer.Unix(), TimeZoneString(timezone)), nowTimestampWithBuffer
}

func getNextTimestamp7DaysEOD(timestamp time.Time) time.Time {
	return timestamp.AddDate(0, 0, 8).Add(-1 * time.Second)
}