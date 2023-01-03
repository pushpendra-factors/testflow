package delta

import (
	M "factors/model/model"
	P "factors/pattern"
	U "factors/util"
)

var sessionMetricToCalcInfo = map[string]EventMetricCalculationInfo{
	M.TotalSessions: {
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
	M.NewUsers: {
		PropsInfo: []EventPropInfo{
			{
				propFunc:     checkNew,
				defaultValue: 1,
			},
		},
		useUnique: true,
	},
	M.RepeatUsers: {
		PropsInfo: []EventPropInfo{
			{
				propFunc:     checkRepeat,
				defaultValue: 1,
			},
		},
		useUnique: true,
	},
	M.SessionsPerUser: {
		PropsInfo: []EventPropInfo{
			{
				defaultValue: 1,
				setBase:      true,
			},
			{
				defaultValue: 1,
			},
		},
		useUnique: true,
	},
	M.EngagedSessions: {
		PropsInfo: []EventPropInfo{
			{
				propFunc:     checkEngagedSession,
				defaultValue: 1,
			},
		},
	},
	M.EngagedUsers: {
		PropsInfo: []EventPropInfo{
			{
				propFunc:     checkEngagedSession,
				defaultValue: 1,
			},
		},
		useUnique: true,
	},
	M.EngagedSessionsPerUser: {
		PropsInfo: []EventPropInfo{
			{
				propFunc:     checkEngagedSession,
				defaultValue: 1,
				setBase:      true,
			},
			{
				defaultValue: 1,
			},
		},
		useUnique: true,
	},
	M.TotalTimeOnSite: {
		PropsInfo: []EventPropInfo{
			{
				propFunc: getPropValueEvents,
				useProp:  U.SP_SPENT_TIME,
			},
		},
	},
	M.AvgSessionDuration: {
		PropsInfo: []EventPropInfo{
			{
				propFunc: getPropValueEvents,
				useProp:  U.SP_SPENT_TIME,
				setBase:  true,
			},
			{
				defaultValue: 1,
			},
		},
	},
	M.AvgPageViewsPerSession: {
		PropsInfo: []EventPropInfo{
			{
				propFunc: getPropValueEvents,
				useProp:  U.SP_PAGE_COUNT,
				setBase:  true,
			},
			{
				defaultValue: 1,
			},
		},
	},
	M.AvgInitialPageLoadTime: {
		PropsInfo: []EventPropInfo{
			{
				propFunc: getPropValueEvents,
				useProp:  U.SP_INITIAL_PAGE_LOAD_TIME,
				setBase:  true,
			},
			{
				defaultValue: 1,
			},
		},
	},
	M.BounceRate: {
		PropsInfo: []EventPropInfo{
			{
				propFunc:     checkBouncedSession,
				defaultValue: 1,
			},
			{
				defaultValue: 1,
			},
		},
	},
	M.EngagementRate: {
		PropsInfo: []EventPropInfo{
			{
				propFunc:     checkEngagedSession,
				defaultValue: 1,
			},
			{
				defaultValue: 1,
			},
		},
	},
}

func checkBouncedSession(eventDetails P.CounterEventFormat, pageName string) (float64, bool, error) {
	if cnt, ok := ExistsInProps(U.SP_PAGE_COUNT, eventDetails.EventProperties, eventDetails.UserProperties, "ep"); ok {
		cnt, err := getFloatValueFromInterface(cnt)
		if err != nil {
			return 0, false, err
		}
		count := int64(cnt)
		if count > 2 {
			return 0, true, nil
		}
	}
	return 0, false, nil
}

func checkEngagedSession(eventDetails P.CounterEventFormat, pageName string) (float64, bool, error) {
	isEngaged := false
	if timeSpent, ok := ExistsInProps(U.SP_SPENT_TIME, eventDetails.EventProperties, eventDetails.UserProperties, "ep"); ok {
		timeSpent, err := getFloatValueFromInterface(timeSpent)
		if err != nil {
			return 0, false, err
		}
		if timeSpent > 10 {
			isEngaged = true
		}
	}
	if cnt, ok := ExistsInProps(U.SP_PAGE_COUNT, eventDetails.EventProperties, eventDetails.UserProperties, "ep"); ok {
		cnt, err := getFloatValueFromInterface(cnt)
		if err != nil {
			return 0, false, err
		}
		count := int64(cnt)
		if count > 2 {
			isEngaged = true
		}
	}
	return 0, isEngaged, nil
}

func checkNew(eventDetails P.CounterEventFormat, prop string) (float64, bool, error) {
	var new bool
	if first, ok := ExistsInProps(U.SP_IS_FIRST_SESSION, eventDetails.EventProperties, eventDetails.UserProperties, "ep"); ok {
		new = first.(bool)
	}
	return 0, new, nil
}

func checkRepeat(eventDetails P.CounterEventFormat, prop string) (float64, bool, error) {
	val, ok, err := checkNew(eventDetails, prop)
	return val, !ok, err
}
