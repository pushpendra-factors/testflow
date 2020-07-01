package task

import (
	PMM "factors/pattern_model_meta"
	"reflect"
	"testing"
	"time"
)

func Test_makeLastBuildTimestampMap(t *testing.T) {

	currentTime := time.Now().Unix()               //now
	lookBackDaysFactor := int64(30 * 24 * 60 * 60) //30 days in secs
	epsilonTime := int64(1 * 60)                   //60 seconds

	type args struct {
		projectData          []PMM.ProjectData
		lookBackPeriodInDays int64
	}
	tests := []struct {
		name string
		args args
		want *map[uint64]map[string]int64
	}{
		//A project with endTime less than lookBackDays
		{"test1", args{[]PMM.ProjectData{{1, 1, "m",
			currentTime - lookBackDaysFactor, currentTime - lookBackDaysFactor - epsilonTime, nil},
			{}}, 30},
			&map[uint64]map[string]int64{}},

		//A project with endTime greater than lookBackDays
		{"test2", args{[]PMM.ProjectData{{1, 1, "m",
			currentTime - lookBackDaysFactor, currentTime - lookBackDaysFactor + epsilonTime, nil},
			{}}, 30},
			&map[uint64]map[string]int64{1: {"m": currentTime - lookBackDaysFactor + epsilonTime}}},

		//A project with endTime almost equal (or less) than lookBackDays
		{"test1", args{[]PMM.ProjectData{{1, 1, "m",
			currentTime - lookBackDaysFactor, currentTime - lookBackDaysFactor, nil},
			{}}, 30},
			&map[uint64]map[string]int64{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := makeLastBuildTimestampMap(tt.args.projectData, tt.args.lookBackPeriodInDays); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("makeLastBuildTimestampMap() = %v, want %v", got, tt.want)
			}
		})
	}
}
