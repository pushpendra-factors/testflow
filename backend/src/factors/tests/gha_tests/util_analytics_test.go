package gha_tests

import (
	U "factors/util"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestGetTimeFromTimestampStr(t *testing.T) {
	str := "2011-10-02 18:48:05.123"
	time := U.GetTimeFromTimestampStr(str)
	assert.NotNil(t, time)
}

func TestGetAllDatesAsTimestamp(t *testing.T) {
	timestamp1 := 1637316219
	timestamp2 := 1639908216
	time, _ := U.GetAllDatesAndOffsetAsTimestamp(int64(timestamp1), int64(timestamp2), "")
	log.WithField("time", time).Warn("kark2")
	assert.Len(t, time, 31)
}

func TestGetAllDatesAndOffsetsAsTimestampForAustraliaTimezone(t *testing.T) {
	timestamp1 := 1648818000
	timestamp2 := 1649080799
	time, offsets := U.GetAllDatesAndOffsetAsTimestamp(int64(timestamp1), int64(timestamp2), "Australia/Sydney")
	assert.Len(t, time, 3)
	assert.Len(t, offsets, 3)
	assert.Equal(t, "+11:00", offsets[0])
	assert.Equal(t, "+10:00", offsets[1])
	assert.Equal(t, "+10:00", offsets[2])
}
