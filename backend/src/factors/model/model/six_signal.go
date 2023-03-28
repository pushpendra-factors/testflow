package model

import (
	"factors/util"
	"github.com/jinzhu/gorm/dialects/postgres"
)

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
	Results   []SixSignalQueryResult `json:"result_group"`
	Query     interface{}            `json:"query"`
	CacheMeta interface{}            `json:"cache_meta"`
}

type SixSignalQueryResult struct {
	Headers []string        `json:"headers"`
	Rows    [][]interface{} `json:"rows"`
	Query   interface{}     `json:"query"`
}

type SixSignalShareableURLParams struct {
	Query           *postgres.Jsonb `json:"six_signal_query"`
	EntityType      int             `json:"entity_type"`
	ShareType       int             `json:"share_type"`
	IsExpirationSet bool            `json:"is_expiration_set"`
	ExpirationTime  int64           `json:"expiration_time"`
}

type SixSignalPublicURLResponse struct {
	ProjectID    int64  `json:"project_id"`
	RouteVersion string `json:"route_version"`
	QueryID      string `json:"query_id"`
}

type SixSignalEmailAndMessage struct {
	EmailIDs []string            `json:"email_ids"`
	Url      string              `json:"url"`
	Domain   string              `json:"domain"`
	From     int64               `json:"fr"`
	To       int64               `json:"to"`
	Timezone util.TimeZoneString `json:"tz"`
}
