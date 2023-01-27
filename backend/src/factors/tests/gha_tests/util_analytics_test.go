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
