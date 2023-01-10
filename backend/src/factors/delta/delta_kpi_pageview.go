package delta

import (
	M "factors/model/model"
	P "factors/pattern"
	U "factors/util"
)

// weekly insights calculation info for each pageview metric
var pageviewMetricToCalcInfo = map[string]EventMetricCalculationInfo{
	M.Entrances: {
		PropsInfo: []EventPropInfo{
			{
				propFunc:     checkEntrance,
				useProp:      "page",
				defaultValue: 1,
			},
		},
	},
	M.Exits: {
		PropsInfo: []EventPropInfo{
			{
				propFunc:     checkExit,
				useProp:      "page",
				defaultValue: 1,
			},
		},
	},
	M.PageViews: {
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
	M.PageviewsPerUser: {
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
	M.AvgPageLoadTime: {
		PropsInfo: []EventPropInfo{
			{
				propFunc: getPropValueEvents,
				useProp:  U.EP_PAGE_LOAD_TIME,
				setBase:  true,
			},
			{
				defaultValue: 1,
			},
		},
	},
	M.AvgVerticalScrollPercent: {
		PropsInfo: []EventPropInfo{
			{
				propFunc: getPropValueEvents,
				useProp:  U.EP_PAGE_SCROLL_PERCENT,
				setBase:  true,
			},
			{
				defaultValue: 1,
			},
		},
	},
	M.AvgTimeOnPage: {
		PropsInfo: []EventPropInfo{
			{
				propFunc: getPropValueEvents,
				useProp:  U.EP_PAGE_SPENT_TIME,
				setBase:  true,
			},
			{
				defaultValue: 1,
			},
		},
	},
	M.EngagedPageViews: {
		PropsInfo: []EventPropInfo{
			{
				propFunc:     checkEngagedPageview,
				defaultValue: 1,
			},
		},
	},
	M.EngagedUsers: {
		PropsInfo: []EventPropInfo{
			{
				propFunc:     checkEngagedPageview,
				defaultValue: 1,
			},
		},
		useUnique: true,
	},
	M.EngagementRate: {
		PropsInfo: []EventPropInfo{
			{
				propFunc:     checkEngagedPageview,
				defaultValue: 1,
				setBase:      true,
			},
			{
				defaultValue: 1,
			},
		},
	},
}

// check if given pageview event quailifies as engaged session (EP_PAGE_SCROLL_PERCENT>50 or EP_PAGE_SPENT_TIME>10)
func checkEngagedPageview(eventDetails P.CounterEventFormat, pageName string) (float64, bool, error) {
	isEngaged := false
	if spentTime, ok := existsInProps(U.EP_PAGE_SPENT_TIME, eventDetails.EventProperties, eventDetails.UserProperties, "ep"); ok {
		spentTime, err := getFloatValueFromInterface(spentTime)
		if err != nil {
			return 0, false, err
		}
		if spentTime > 10 {
			isEngaged = true
		}
	}
	if scrollPerc, ok := existsInProps(U.EP_PAGE_SCROLL_PERCENT, eventDetails.EventProperties, eventDetails.UserProperties, "ep"); ok {
		scrollPerc, err := getFloatValueFromInterface(scrollPerc)
		if err != nil {
			return 0, false, err
		}
		if scrollPerc > 50 {
			isEngaged = true
		}
	}
	return 0, isEngaged, nil
}

// check if given session event qualifies as entrance for pageName (SP_INITIAL_PAGE_URL=pageName)
func checkEntrance(eventDetails P.CounterEventFormat, pageName string) (float64, bool, error) {
	isEntrance := false
	if url, ok := existsInProps(U.SP_INITIAL_PAGE_URL, eventDetails.EventProperties, eventDetails.UserProperties, "ep"); ok {
		if url := url.(string); url == pageName {
			isEntrance = true
		}
	}
	return 0, isEntrance, nil
}

// check if given session event qualifies as exit for pageName (SP_LATEST_PAGE_URL=pageName)
func checkExit(eventDetails P.CounterEventFormat, pageName string) (float64, bool, error) {
	isEntrance := false
	if url, ok := existsInProps(U.SP_LATEST_PAGE_URL, eventDetails.EventProperties, eventDetails.UserProperties, "ep"); ok {
		if url := url.(string); url == pageName {
			isEntrance = true
		}
	}
	return 0, isEntrance, nil
}
