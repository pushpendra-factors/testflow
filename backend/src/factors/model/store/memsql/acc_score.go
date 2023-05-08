package memsql

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	C "factors/config"
	"factors/model/model"
	"fmt"
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

type JSON json.RawMessage

// Scan scan value into Jsonb, implements sql.Scanner interface
func (j *JSON) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	result := json.RawMessage{}
	err := json.Unmarshal(bytes, &result)
	*j = JSON(result)
	return err
}

// Value return json value, implement driver.Valuer interface
func (j JSON) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return json.RawMessage(j).MarshalJSON()
}

func (store *MemSQL) UpdateUserEventsCount(evdata []model.EventsCountScore) error {
	projectID := evdata[0].ProjectId
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	db := C.GetServices().Db
	for _, ev := range evdata {
		us := model.User{}
		us.ID = ev.UserId
		us.ProjectId = ev.ProjectId
		uk := make(map[string]map[string]int64)
		uk[ev.DateStamp] = ev.EventScore
		eventsCountjson, err := json.Marshal(ev.EventScore)
		if err != nil {
			logCtx.WithError(err).Errorf("Unable to convert map to json in event counts ")
		}

		stmt := fmt.Sprintf("UPDATE users SET event_aggregate::`%s`='%s' where id='%s' and project_id=%d ", ev.DateStamp, eventsCountjson, ev.UserId, ev.ProjectId)
		err = db.Exec(stmt).Error
		if err != nil {
			return err
		}
	}
	return nil

}

func (store *MemSQL) UpdateGroupEventsCount(evdata []model.EventsCountScore) error {
	projectID := evdata[0].ProjectId
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	db := C.GetServices().Db
	group_ids := make([]string, 0)
	for _, ev := range evdata {
		us := model.User{}
		us.ID = ev.UserId
		us.ProjectId = ev.ProjectId
		uk := make(map[string]map[string]int64)
		uk[ev.DateStamp] = ev.EventScore
		eventsCountjson, err := json.Marshal(ev.EventScore)
		if err != nil {
			logCtx.WithError(err).Errorf("Unable to convert map to json in event counts ")
		}

		if len(ev.EventScore) > 0 {
			stmt := fmt.Sprintf("UPDATE users SET event_aggregate::`%s`='%s' where id='%s' and project_id=%d", ev.DateStamp, eventsCountjson, ev.UserId, ev.ProjectId)
			err = db.Exec(stmt).Error
			if err != nil {
				return err
			}
			group_ids = append(group_ids, ev.UserId)
		}

	}
	logCtx.WithField("Num of groups added", len(group_ids)).Infof("groups :%v", group_ids)
	return nil

}

func getPrevDateString() string {
	current_ts := time.Now()
	YY, MM, DD := current_ts.Date()

	current_date := time.Date(YY, MM, DD, 0, 0, 0, 0, time.UTC)
	prev_ts := current_date.AddDate(0, 0, -1)
	prev_ts_string := GetDateOnlyFromTimestamp(prev_ts.Unix())
	return prev_ts_string
}

func (store *MemSQL) GetAccountsScore(project_id int64, group_id int, ts string, debug bool) ([]model.PerAccountScore, error) {
	// get score of each account on current date

	projectID := project_id
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	db := C.GetServices().Db

	weights, _ := store.GetWeightsByProject(project_id)
	group_string := fmt.Sprintf("group_%d_user_id", group_id)
	stmt := fmt.Sprintf("select id, json_extract_json(event_aggregate, '%s' ) from users as u inner join (select DISTINCT( %s ) from users where project_id = %d  and %s is not null) as s on s.%s =u.id and json_length(json_extract_json(event_aggregate,'%s'))>=1 and project_id= %d", ts, group_string, project_id, group_string, group_string, ts, project_id)

	logCtx.WithField("sql", stmt).Info("sql statement")
	rows, err := db.Raw(stmt).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]model.PerAccountScore, 0)
	for rows.Next() {
		var events_count_string string
		counts_map := map[string]int{}
		var account_score float32
		var account_id string
		var r model.PerAccountScore
		err := rows.Scan(&account_id, &events_count_string)
		if err != nil {
			logCtx.WithError(err).Error("Failed to get event counts for groups")
		}
		err = json.Unmarshal([]byte(events_count_string), &counts_map)
		if err != nil {
			logCtx.WithError(err).Error("Unable to read counts data from DB")
			return nil, err
		}

		account_score, counts_map, err = ComputeAccountScore(*weights, counts_map)
		if err != nil {
			return nil, err
		}
		r.Id = account_id
		r.Score = account_score
		r.Timestamp = ts // have to fill the right ts
		if debug {
			r.Debug = counts_map
		}
		result = append(result, r)

	}

	return result, nil
}

func (store *MemSQL) GetUserScore(projectId int64, userId string, eventTS string, debug bool, is_anonymous bool) (model.PerUserScoreOnDay, error) {
	// get score of each account on current date

	projectID := projectId
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	var rt string
	var result model.PerUserScoreOnDay

	db := C.GetServices().Db

	weights, errStatus := store.GetWeightsByProject(projectId)
	if errStatus != http.StatusFound {
		return result, fmt.Errorf("unable to get weights for project %d", projectId)
	}

	if !is_anonymous {

		acc_score, err := ComputeUserScoreOnCustomerId(db, userId, projectID, eventTS, *weights)
		if err != nil {
			return result, err
		}
		result.Id = userId
		result.Score = acc_score
		result.Timestamp = eventTS
		return result, nil
	}

	stmt := "select json_extract_json(event_aggregate,?) from users where id=? and project_id=?"
	tx := db.Raw(stmt, eventTS, userId, projectId).Row()
	err := tx.Scan(&rt)
	if err != nil {
		logCtx.WithError(err).Error("Unable to read user counts data from DB")
		return result, err
	}
	countsMap := map[string]int{}
	err = json.Unmarshal([]byte(rt), &countsMap)
	if err != nil {
		logCtx.WithError(err).Error("Unable to unamrshall user counts data")
		return result, err
	}
	accountScore, countsMap, err := ComputeAccountScore(*weights, countsMap)
	if err != nil {
		return result, err
	}
	result.Id = userId
	result.Score = accountScore
	result.Timestamp = eventTS
	if debug {
		result.Debug = countsMap
	}
	return result, nil
}

func ComputeUserScoreOnCustomerId(db *gorm.DB, id string, projectId int64, eventTs string, weights model.AccWeights) (float32, error) {
	logCtx := log.WithFields(log.Fields{
		"project_id": projectId,
		"user_id":    id,
	})

	stmt := "select json_extract_json(event_aggregate,?) from users where customer_user_id=? and project_id=?"
	tx, err := db.Raw(stmt, eventTs, id, projectId).Rows()
	if err != nil {
		return 0, err
	}
	defer tx.Close()
	final_score := float32(0)
	for tx.Next() {
		var events_count_string string
		counts_map := map[string]int{}
		var account_score float32
		var account_id string
		err := tx.Scan(&account_id, &events_count_string)
		if err != nil {
			logCtx.WithError(err).Error("Failed to get event counts for groups")
		}
		err = json.Unmarshal([]byte(events_count_string), &counts_map)
		if err != nil {
			logCtx.WithError(err).Error("Unable to read counts data from DB")
			return 0, err
		}

		account_score, counts_map, err = ComputeAccountScore(weights, counts_map)
		if err != nil {
			return 0, err
		}
		final_score += account_score

	}

	return final_score, nil

}

func ComputeAccountScore(weights model.AccWeights, eventsCount map[string]int) (float32, map[string]int, error) {

	// var accountScore int64
	var accountScore float32
	var eventsCountMap map[string]int = make(map[string]int)
	weightValue := weights.WeightConfig
	for _, w := range weightValue {
		if ew, ok := eventsCount[w.WeightId]; ok {
			accountScore += float32(ew) * w.Weight_value
			eventsCountMap[w.WeightId] = ew
		}
	}

	return accountScore, eventsCountMap, nil
}

func GetDateOnlyFromTimestamp(ts int64) string {
	year, month, date := time.Unix(ts, 0).UTC().Date()
	data := fmt.Sprintf("%d%02d%02d", year, month, date)
	return data

}

func (store *MemSQL) GetAllUserScore(projectId int64, debug bool) ([]model.AllUsersScore, error) {
	// get score of each account on current date

	projectID := projectId
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	MAX_LIMIT := 10000
	logCtx := log.WithFields(logFields)
	db := C.GetServices().Db

	weights, _ := store.GetWeightsByProject(projectId)
	stmt := fmt.Sprintf("select id, event_aggregate from users where json_length(event_aggregate)>=1 and project_id= %d LIMIT %d", projectId, MAX_LIMIT)
	logCtx.WithField("sql", stmt).Info("sql statement")
	rows, err := db.Raw(stmt).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]model.AllUsersScore, 0)
	for rows.Next() {
		var eventsCountString string
		countsMapDays := map[string]map[string]int{}
		resultmap := make(map[string]float64)
		var userId string
		var r model.AllUsersScore
		err := rows.Scan(&userId, &eventsCountString)
		if err != nil {
			logCtx.WithError(err).Error("Failed to get event counts for groups")
		}

		err = json.Unmarshal([]byte(eventsCountString), &countsMapDays)
		if err != nil {
			logCtx.WithError(err).Error("Failed to unmarshall json counts for users per day")
		}

		for day, countsPerday := range countsMapDays {
			accountScore, _, err := ComputeAccountScore(*weights, countsPerday)
			if err != nil {
				return nil, err
			}
			resultmap[day] = float64(accountScore)
		}
		r.UserId = userId
		r.ScorePerDay = resultmap
		result = append(result, r)
	}

	return result, nil
}
