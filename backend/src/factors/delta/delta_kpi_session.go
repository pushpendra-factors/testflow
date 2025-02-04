package delta

import (
	M "factors/model/model"
	P "factors/pattern"
	U "factors/util"
)

// weekly insights calculation info for each session metric
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

// check if given session event quailifies as bounced session (SP_PAGE_COUNT<2)
func checkBouncedSession(eventDetails P.CounterEventFormat, pageName string) (float64, bool, error) {
	if cnt, ok := existsInProps(U.SP_PAGE_COUNT, eventDetails.EventProperties, eventDetails.UserProperties, "ep"); ok {
		cnt, err := U.GetFloatValueFromInterface(cnt)
		if err != nil {
			return 0, false, err
		}
		count := int64(cnt)
		if count < 2 {
			return 0, true, nil
		}
	}
	return 0, false, nil
}

// check if given session event quailifies as engaged session (SP_PAGE_COUNT>2 or SP_SPENT_TIME>10)
func checkEngagedSession(eventDetails P.CounterEventFormat, pageName string) (float64, bool, error) {
	isEngaged := false
	if timeSpent, ok := existsInProps(U.SP_SPENT_TIME, eventDetails.EventProperties, eventDetails.UserProperties, "ep"); ok {
		timeSpent, err := U.GetFloatValueFromInterface(timeSpent)
		if err != nil {
			return 0, false, err
		}
		if timeSpent > 10 {
			isEngaged = true
		}
	}
	if cnt, ok := existsInProps(U.SP_PAGE_COUNT, eventDetails.EventProperties, eventDetails.UserProperties, "ep"); ok {
		cnt, err := U.GetFloatValueFromInterface(cnt)
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

// check if given session event quailifies as new session (SP_IS_FIRST_SESSION=true)
func checkNew(eventDetails P.CounterEventFormat, prop string) (float64, bool, error) {
	var new bool
	if first, ok := existsInProps(U.SP_IS_FIRST_SESSION, eventDetails.EventProperties, eventDetails.UserProperties, "ep"); ok {
		new = first.(bool)
	}
	return 0, new, nil
}

// check if given session event quailifies as repeat session (not a new session)
func checkRepeat(eventDetails P.CounterEventFormat, prop string) (float64, bool, error) {
	val, ok, err := checkNew(eventDetails, prop)
	return val, !ok, err
}
