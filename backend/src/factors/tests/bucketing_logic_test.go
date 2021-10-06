package tests

import (
	"factors/task"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBucketingLogicWithPercentile(t *testing.T) {
	propertyRangeMap := make(map[string]task.MinMaxTuple)
	propertyRangeMap["p1"] = task.MinMaxTuple{
		Min:     0,
		Max:     4,
		Numbers: []float64{0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 3, 4},
	}
	propertyRangeMap["p2"] = task.MinMaxTuple{
		Min:     10,
		Max:     40,
		Numbers: []float64{10, 11, 13, 20, 21, 22, 23, 24, 25, 26, 30, 31, 33, 34, 36, 39, 40, 40, 40, 40, 40, 40},
	}
	propertyRangeMap["p3"] = task.MinMaxTuple{
		Min:     6000,
		Max:     12000,
		Numbers: []float64{6005, 6007, 7032, 7045, 7070, 7092, 8000, 8001, 8002, 8003, 8004, 8005, 8006, 8007, 8008, 8009, 8010, 8012, 8059, 8080, 8086, 9000, 9090, 10000, 12000},
	}
	propertyRangeMap["p4"] = task.MinMaxTuple{
		Min:     0,
		Max:     10000000,
		Numbers: []float64{0, 1, 200, 2000, 3000, 10000, 10000, 100000, 1000001, 10000000},
	}
	propertyRangeMap["p5"] = task.MinMaxTuple{
		Min:     0,
		Max:     0,
		Numbers: []float64{0, 0, 0, 0, 0, 0, 0},
	}
	propertyRangeMap["p6"] = task.MinMaxTuple{
		Min:     1,
		Max:     1,
		Numbers: []float64{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
	}
	propertyRangeMap["p7"] = task.MinMaxTuple{
		Min:     780,
		Max:     780,
		Numbers: []float64{780, 780, 780, 780, 780, 780, 780, 780, 780, 780, 780, 780, 780, 780, 780, 780, 780, 780, 780, 780, 780, 780, 780, 780, 780, 780, 780, 780, 780, 780, 780, 780, 780, 780, 780},
	}
	propertyRangeMap["p8"] = task.MinMaxTuple{
		Min:     783,
		Max:     783,
		Numbers: []float64{783, 783},
	}
	propertyRangeMap["p9"] = task.MinMaxTuple{
		Min:     1000,
		Max:     50000,
		Numbers: []float64{1001, 1007, 1010, 1016, 1023, 1034, 1040, 1050, 1060, 1090, 2001, 2007, 2010, 2016, 2023, 2034, 2040, 2050, 2060, 2090, 3001, 3007, 3010, 3016, 3023, 3034, 3040, 3050, 3060, 3090, 30000, 40000, 50000},
	}
	result := task.BucketizeNumericalPropertiesUsingPercentile(propertyRangeMap)
	assert.Equal(t, result["p1"][0].Label, "lesserOrEquals 0")
	assert.Equal(t, result["p1"][3].Label, "greaterOrEquals 4")
	assert.Equal(t, result["p2"][0].Label, "lesserOrEquals 10")
	assert.Equal(t, result["p2"][9].Label, "greaterOrEquals 41")
	assert.Equal(t, result["p3"][0].Label, "lesserOrEquals 6000")
	assert.Equal(t, result["p3"][9].Label, "greaterOrEquals 9101")
	assert.Equal(t, result["p4"][0].Label, "lesserOrEquals 0")
	assert.Equal(t, result["p4"][9].Label, "greaterOrEquals 10000001")
	assert.Equal(t, result["p5"][0].Label, "lesserOrEquals -1")
	assert.Equal(t, result["p5"][2].Label, "greaterOrEquals 1")
	assert.Equal(t, result["p6"][0].Label, "lesserOrEquals 0")
	assert.Equal(t, result["p6"][2].Label, "greaterOrEquals 2")
	assert.Equal(t, result["p7"][0].Label, "lesserOrEquals 779")
	assert.Equal(t, result["p7"][2].Label, "greaterOrEquals 781")
	assert.Equal(t, result["p8"][0].Label, "lesserOrEquals 782")
	assert.Equal(t, result["p8"][2].Label, "greaterOrEquals 784")
	assert.Equal(t, result["p9"][0].Label, "lesserOrEquals 1000")
	assert.Equal(t, result["p9"][6].Label, "greaterOrEquals 3101")
	assert.Equal(t, 9, len(result))
}
