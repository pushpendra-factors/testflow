package model

import "factors/util"

type SixSignalQueryGroup struct {
	Queries []SixSignalQuery `json:"six_signal_query_group"`
}

func (q *SixSignalQueryGroup) SetTimeZone(timezoneString util.TimeZoneString) {
	for i := 0; i < len(q.Queries); i++ {
		q.Queries[i].Timezone = timezoneString
	}
}

type SixSignalQuery struct {
	Timezone util.TimeZoneString `json:"tz"`
	From     int64               `json:"fr"`
	To       int64               `json:"to"`
}

type SixSignalResultGroup struct {
	Results     []SixSignalQueryResult `json:"result_group"`
	Query       interface{}            `json:"query"` // GroupQuery for which the result group is fetched
	IsShareable bool                   `json:"is_shareable"`
	CacheMeta   interface{}            `json:"cache_meta"` //cache info
}

type SixSignalQueryResult struct {
	Headers   []string        `json:"headers"`
	Rows      [][]interface{} `json:"rows"`
	Query     interface{}     `json:"query"`      //query for which the result is fetched
	CacheMeta interface{}     `json:"cache_meta"` //cache info will be added to this
}
