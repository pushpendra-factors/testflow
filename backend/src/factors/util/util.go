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
