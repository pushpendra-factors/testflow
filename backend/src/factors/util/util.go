package util

import (
	"math/rand"
	"strconv"
	"strings"
	"time"
)

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

func TrimQuotes(str string) string {
	return strings.TrimSuffix(strings.TrimPrefix(str, "\""), "\"")
}
