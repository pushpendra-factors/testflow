package tests

import (
	Mems "factors/model/store/memsql"
	U "factors/util"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/jinzhu/now"
	"github.com/stretchr/testify/assert"
)

func TestGetAllWeeksAsTimestamp(t *testing.T) {
	type args struct {
		fromUnix int64
		toUnix   int64
		timezone string
	}
	// 1596220200 : Saturday, 1 August 2020 00:00:00 GMT+05:30
	// 1598898599 ; Monday, 31 August 2020 23:59:59 GMT+05:30
	aug1 := int64(1596220200)
	aug31 := int64(1598898599)
	arg1 := args{fromUnix: aug1, toUnix: aug31, timezone: string(U.TimeZoneStringIST)}

	tests := []struct {
		name string
		args args
		want []time.Time
	}{
		{"Test1", arg1, []time.Time{
			time.Unix(1595701800, 0),
			time.Unix(1596306600, 0),
			time.Unix(1596911400, 0),
			time.Unix(1597516200, 0),
			time.Unix(1598121000, 0),
			time.Unix(1598725800, 0),
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := U.GetAllWeeksAsTimestamp(tt.args.fromUnix, tt.args.toUnix, tt.args.timezone); !reflect.DeepEqual(got, tt.want) {
				for i, x := range got {
					if x.Unix() != tt.want[i].Unix() {
						t.Errorf("GetAllWeeksAsTimestamp() = %d, want %d", x.Unix(), tt.want[i].Unix())
					}
				}
			}
		})
	}
}

func TestGetAllMonthsAsTimestamp(t *testing.T) {
	type args struct {
		fromUnix int64
		toUnix   int64
		timezone string
	}
	// 1579220200 : Friday, 17 January 2020 00:00:00 GMT+05:30
	// 1598898599 : Monday, 31 August 2020 23:59:59 GMT+05:30
	aug1 := int64(1579220200)
	aug31 := int64(1598898599)
	arg1 := args{fromUnix: aug1, toUnix: aug31, timezone: string(U.TimeZoneStringIST)}

	tests := []struct {
		name string
		args args
		want []time.Time
	}{
		{"Test1", arg1, []time.Time{
			time.Unix(1577817000, 0),
			time.Unix(1580495400, 0),
			time.Unix(1583001000, 0),
			time.Unix(1585679400, 0),
			time.Unix(1588271400, 0),
			time.Unix(1590949800, 0),
			time.Unix(1593541800, 0),
			time.Unix(1596220200, 0),
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := U.GetAllMonthsAsTimestamp(tt.args.fromUnix, tt.args.toUnix, tt.args.timezone); !reflect.DeepEqual(got, tt.want) {
				for i, x := range got {
					if x.Unix() != tt.want[i].Unix() {
						t.Errorf("GetAllWeeksAsTimestamp() = %d, want %d", x.Unix(), tt.want[i].Unix())
					}
				}
			}
		})
	}
}

func TestOffsetsAndDateTimesForSpecificLocation(t1 *testing.T) {
	p := fmt.Println
	timezoneLocation, _ := time.LoadLocation("Australia/Sydney")

	then1 := time.Date(2009, 04, 01, 00, 00, 00, 00, timezoneLocation)
	p(then1)
	timestampWithTimezone := U.GetTimestampAsStrWithTimezone(then1, "Australia/Sydney")
	p(timestampWithTimezone)

	// storeSel := store.GetStore()
	timestamps, offsets := Mems.GetAllTimestampsAndOffsetBetweenByType(1640955600, 1673614799, "month", "Australia/Sydney")
	p(timestamps)
	p(offsets)

	for index, ts := range timestamps {
		p(U.GetTimestampAsStrWithTimezoneGivenOffset(ts, offsets[index]))
	}

	t := time.Date(2009, 03, 01, 00, 00, 00, 00, timezoneLocation)
	t = t.AddDate(0, 0, 35).Add(4 * time.Hour) // next month.
	// t = t.AddDate(0, 0, 30) // next month.
	t, _ = U.GetTimeFromUnixTimestampWithZone(t.Unix(), "Australia/Sydney")
	// p(t)
	// t = time.Date(t.Year(), t.Month(), t.Day(), 00, 00, 0, 0, timezoneLocation)
	// t = now.New(t).BeginningOfMonth()
	// p(t.Month())
	// p(t)
	t = now.New(t).BeginningOfMonth()
	p(t)
	tString := U.GetTimestampAsStrWithTimezoneGivenOffset(t, "+11:00")
	t2 := U.GetTimeFromParseTimeStr(tString)
	p(t2)
	// t = time.Date(t.Year(), t.Month(), 00, 00, 00, 0, 0, timezoneLocation)
	// t = now.New(t).BeginningOfMonth()
	// p(t)
	// p(t.Month())
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
