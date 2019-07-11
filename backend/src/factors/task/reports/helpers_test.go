package reports

import (
	M "factors/model"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetWeeklyIntervals(t *testing.T) {

	st := int64(1396742400) //2014-04-06 00:00:00 +0000 UTC
	et := int64(1398556799) //2014-04-26 23:59:59 +0000 UTC

	intervals := getWeeklyIntervals(unixToUTCTime(st), unixToUTCTime(et))

	expIntervals := []M.Interval{
		M.Interval{
			StartTime: int64(1396742400), // 2014-04-06 00:00:00 +0000 UTC
			EndTime:   int64(1397347199), // 2014-04-12 23:59:59 +0000 UTC
		},
		M.Interval{
			StartTime: int64(1397347200), // 2014-04-13 00:00:00 +0000 UTC
			EndTime:   int64(1397951999), // 2014-04-19 23:59:59 +0000 UTC
		},
		M.Interval{
			StartTime: int64(1397952000), // 2014-04-20 00:00:00 +0000 UTC
			EndTime:   int64(1398556799), // 2014-04-26 23:59:59 +0000 UTC
		},
	}

	assert.Equal(t, intervals, expIntervals)
}
