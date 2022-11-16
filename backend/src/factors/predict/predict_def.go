package predict

import (
	log "github.com/sirupsen/logrus"
)

var prLog = log.WithField("prefix", "Task#PredictPullEvents")

type UsersCohort struct {
	User_id          string  `json:"uid"`
	User_customer_id string  `json:"uci"`
	Project_id       string  `json:"prj"`
	Event_name_id    string  `json:"eni"`
	Event_name       string  `json:"en"`
	Group_count      int     `json:"gpc"`
	Amount           float32 `json:"amt"`
	Timestamp        int64   `json:"ts"`
}

type UsersByEvent struct {
	User_id       string `json:"uid"`
	Event_name_id string `json:"eni"`
	Event_name    string `json:"en"`
	Timestamp     int64  `json:"ts"`
	Is_group_user int    `json:"igu"`
	Group_id      string `json:"gid"`
}
type UsersByEventName struct {
	User_id       string `json:"uid"`
	Event_name_id string `json:"eni"`
	Event_name    string `json:"en"`
	Timestamp     int64  `json:"ts"`
}
type GroupIDCounts struct {
	Project_id  string `json:"pid"`
	Group_Id    string `json:"eni"`
	Group_count string `json:"gc"`
}

type PredictProject struct {
	Project_id       int64  `json:"pid" `
	Base_event       string `json:"be"`
	Target_event     string `json:"te"`
	LookBack_in_days int    `json:"lk"`
	Target_ts        int64  `json:"ts"`
	Start_time       int64  `json:"sst"`
	End_time         int64  `json:"et"`
	Buffer_time      int64  `json:"bt"`
	Group_type       string `json:"gty"`
	Distribution     string `json:"dis"`
	Target_value     string `json:"tv"`
	Model_id         int64  `json:"mid"`
}
