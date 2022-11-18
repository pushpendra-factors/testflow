package tests

import (
	U "factors/util"
	"reflect"
	"testing"
	"time"
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
