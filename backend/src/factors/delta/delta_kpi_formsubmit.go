package delta

import (
	M "factors/model/model"
)

var formSubmitMetricToCalcInfo = map[string]EventMetricCalculationInfo{
	M.Count: {
		PropsInfo: []EventPropInfo{
			{
				defaultValue: 1,
			},
		},
	},
	M.UniqueUsers: {
		PropsInfo: []EventPropInfo{
			{
				defaultValue: 1,
			},
		},
		useUnique: true,
	},
	M.CountPerUser: {
		PropsInfo: []EventPropInfo{
			{
				defaultValue: 1,
			},
			{
				defaultValue: 1,
			},
		},
		useUnique: true,
	},
}
