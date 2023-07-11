package memsql

import (
	"database/sql"
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

const MAX_LIMIT = 10000

func (store *MemSQL) UpdateUserEventsCount(evdata []model.EventsCountScore, lastevent map[string]model.LatestScore) error {
	projectID := evdata[0].ProjectId
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	db := C.GetServices().Db
	for _, ev := range evdata {
		var countsLatest model.LatestScore
		us := model.User{}
		us.ID = ev.UserId
		us.ProjectId = ev.ProjectId
		uk := make(map[string]map[string]int64)
		uk[ev.DateStamp] = ev.EventScore
		countsLatest.Date = model.GetDateFromString(ev.DateStamp)
		countsLatest.EventsCount = make(map[string]float64)
		for k, v := range ev.EventScore {
			countsLatest.EventsCount[k] = float64(v)
		}
		eventsCountjson, err := json.Marshal(countsLatest)
		if err != nil {
			logCtx.WithError(err).Errorf("Unable to convert map to json in event counts ")
		}

		LatestScoreString, err := json.Marshal(lastevent[ev.UserId])
		if err != nil {
			logCtx.WithError(err).Errorf("Unable to marshall latest score ")
		}
		logCtx.Debugf("user last event data: %s", string(LatestScoreString))

		stmt := fmt.Sprintf("UPDATE users SET event_aggregate = case when event_aggregate is NULL THEN '{}' ELSE  event_aggregate END, event_aggregate::`%s`='%s' ,event_aggregate::`%s` = '%s' where id= ? and project_id= ?", ev.DateStamp, eventsCountjson, model.LAST_EVENT, LatestScoreString)
		err = db.Exec(stmt, ev.UserId, ev.ProjectId).Error
		if err != nil {
			logCtx.WithError(err).Errorf("Unable to update latest score in user events ")
			return err
		}

	}
	return nil

}

func (store *MemSQL) UpdateGroupEventsCount(evdata []model.EventsCountScore, lastevent map[string]model.LatestScore) error {
	projectID := evdata[0].ProjectId
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	db := C.GetServices().Db
	group_ids := make([]string, 0)
	for _, ev := range evdata {
		var countsLatest model.LatestScore
		us := model.User{}
		us.ID = ev.UserId
		us.ProjectId = ev.ProjectId
		uk := make(map[string]map[string]int64)
		uk[ev.DateStamp] = ev.EventScore
		countsLatest.Date = model.GetDateFromString(ev.DateStamp)
		countsLatest.EventsCount = make(map[string]float64)
		for k, v := range ev.EventScore {
			countsLatest.EventsCount[k] = float64(v)
		}
		eventsCountjson, err := json.Marshal(countsLatest)
		if err != nil {
			logCtx.WithError(err).Errorf("unable to convert map to json in event counts ")
		}

		if len(ev.EventScore) > 0 {

			var lt model.LatestScore
			lt.EventsCount = make(map[string]float64)

			if le, ok := lastevent[ev.UserId]; ok {
				lt = le
			} else {
				e := fmt.Errorf("unable to find, last event for user:%s", ev.UserId)
				logCtx.Error(e)
			}

			LatestScoreString, err := json.Marshal(lt)
			if err != nil {
				logCtx.WithError(err).Errorf("Unable to marshall latest score ")
			}

			stmt := fmt.Sprintf("UPDATE users SET event_aggregate = case when event_aggregate is NULL THEN '{}' ELSE  event_aggregate END, event_aggregate::`%s`='%s' ,event_aggregate::`%s` = '%s' where id= ? and project_id= ?", ev.DateStamp, eventsCountjson, model.LAST_EVENT, LatestScoreString)
			err = db.Exec(stmt, ev.UserId, ev.ProjectId).Error
			if err != nil {
				logCtx.WithError(err).Errorf("Unable to update latest score in user events ")
				return err
			}

			group_ids = append(group_ids, ev.UserId)
		}

	}
	logCtx.WithField("Num of groups added", len(group_ids)).Infof("groups :%v", group_ids)
	return nil

}

func (store *MemSQL) GetAccountsScore(projectId int64, groupId int, ts string, debug bool) ([]model.PerAccountScore, *model.AccWeights, error) {
	// get score of each account on current date

	projectID := projectId
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	db := C.GetServices().Db

	weights, _ := store.GetWeightsByProject(projectId)
	group_string := fmt.Sprintf("group_%d_user_id", groupId)
	stmt := fmt.Sprintf("select id, json_extract_json(event_aggregate, '%s' ) from users as u inner join (select DISTINCT( %s ) from users where project_id = ? and %s is not null) as s on s.%s =u.id and json_length(json_extract_json(event_aggregate,'%s'))>=1 and project_id= ?", ts, group_string, group_string, group_string, ts)
	rows, err := db.Raw(stmt, projectID, projectID).Rows()
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	result := make([]model.PerAccountScore, 0)
	for rows.Next() {
		var events_count_string string
		counts_map := map[string]int64{}
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
			return nil, nil, err
		}

		account_score, counts_map, decayValue, err := ComputeAccountScore(*weights, counts_map, ts)
		if err != nil {
			return nil, nil, err
		}
		r.Id = account_id
		r.Score = account_score
		r.Timestamp = ts // have to fill the right timestamp
		if debug {
			r.Debug = make(map[string]interface{})
			r.Debug["counts"] = counts_map
			r.Debug["decay"] = decayValue
			r.Debug["score"] = account_score
			r.Debug["timestamp"] = ts
		}
		result = append(result, r)
	}

	return result, weights, nil
}

func (store *MemSQL) GetPerAccountScore(projectId int64, timestamp string, userId string, numDaysToTrend int, debug bool) (model.PerAccountScore, *model.AccWeights, error) {
	// get score of each account on current date

	projectID := projectId
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	var rt string
	var result model.PerAccountScore
	countsMapDays := map[string]map[string]int64{}
	scoreOnDays := make(map[string]float32)
	resultmap := make(map[string]model.PerAccountScore)

	currentDate := model.GetDateFromString(timestamp)
	prevDateTotrend := time.Unix(currentDate, 0).AddDate(0, 0, -1*numDaysToTrend).Unix()

	db := C.GetServices().Db

	weights, errStatus := store.GetWeightsByProject(projectId)
	if errStatus != http.StatusFound {
		return result, nil, fmt.Errorf("unable to get weights for project %d", projectId)
	}
	logCtx.WithField("weights", weights).Info("weights")

	stmt := "select event_aggregate from users where id=? and project_id=?"
	tx := db.Raw(stmt, userId, projectId).Row()
	err := tx.Scan(&rt)
	if err != nil {
		logCtx.WithError(err).Error("Unable to read user counts data from DB")
		return result, nil, err
	}

	err = json.Unmarshal([]byte(rt), &countsMapDays)
	if err != nil {
		logCtx.WithError(err).Error("Failed to unmarshall json counts for users per day")
	}

	for day, countsPerday := range countsMapDays {
		accountScore, _, decay, err := ComputeAccountScore(*weights, countsPerday, day)
		if err != nil {
			return model.PerAccountScore{}, nil, err
		}

		if day != model.LAST_EVENT {
			var t model.PerAccountScore
			t.Score = float32(accountScore)
			t.Timestamp = day
			unixDay := model.GetDateFromString(day)
			if unixDay >= prevDateTotrend {
				scoreOnDays[day] = accountScore
				if debug {
					t.Debug = make(map[string]interface{})
					t.Debug["counts"] = make(map[string]int64, 0)
					t.Debug["counts"] = countsPerday
					t.Debug["date"] = day
					t.Debug["score"] = accountScore
					t.Debug["decay"] = decay
				}
			}
			resultmap[day] = t
		}
	}

	if debug {
		result.Debug = make(map[string]interface{})
		result.Debug["info"] = resultmap
	}
	result.Id = userId
	result.Trend = make(map[string]float32)
	result.Trend = scoreOnDays
	result.Score = scoreOnDays[timestamp]
	result.Timestamp = timestamp
	return result, weights, nil

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
	logCtx.WithField("weights", weights).Info("weights")
	if !is_anonymous {
		logCtx.Info("anonymous")

		acc_score, all_scores, err := ComputeUserScoreOnCustomerId(db, userId, projectID, eventTS, *weights)
		if err != nil {
			return result, err
		}
		result.Id = userId
		result.Score = acc_score
		result.Timestamp = eventTS
		result.Debug = make(map[string]interface{})
		result.Debug["IDS_scores"] = all_scores
		return result, nil
	}
	logCtx.Info("not anonymous")
	stmt := "select json_extract_json(event_aggregate,?) from users where id=? and project_id=?"
	tx := db.Raw(stmt, eventTS, userId, projectId).Row()
	err := tx.Scan(&rt)
	if err != nil {
		logCtx.WithError(err).Error("Unable to read user counts data from DB")
		return result, err
	}
	countsMap := map[string]int64{}
	err = json.Unmarshal([]byte(rt), &countsMap)
	if err != nil {
		logCtx.WithError(err).Error("Unable to unamrshall user counts data")
		return result, err
	}
	accountScore, countsMap, decayValue, err := ComputeAccountScore(*weights, countsMap, eventTS)
	if err != nil {
		return result, err
	}
	result.Id = userId
	result.Score = accountScore
	result.Timestamp = eventTS
	if debug {
		result.Debug = make(map[string]interface{})
		result.Debug["counts"] = countsMap
		result.Debug["weights"] = weights
		result.Debug["decay"] = decayValue
	}
	return result, nil
}

func ComputeUserScoreOnCustomerId(db *gorm.DB, id string, projectId int64, eventTs string, weights model.AccWeights) (float32, []model.PerUserScoreOnDay, error) {
	logCtx := log.WithFields(log.Fields{
		"project_id": projectId,
		"user_id":    id,
	})

	stmt := "select id,json_extract_json(event_aggregate,?) from users where customer_user_id=? and project_id=?"
	tx, err := db.Raw(stmt, eventTs, id, projectId).Rows()
	if err != nil {
		return 0, nil, err
	}
	defer tx.Close()
	final_score := float32(0)
	usersOnDay := make([]model.PerUserScoreOnDay, 0)
	for tx.Next() {
		var events_count_string string
		counts_map := map[string]int64{}
		var account_score float32
		var user_id string
		err := tx.Scan(&user_id, &events_count_string)
		if err != nil {
			logCtx.WithError(err).Error("Failed to get event counts for groups")
		}
		err = json.Unmarshal([]byte(events_count_string), &counts_map)
		if err != nil {
			logCtx.WithError(err).Error("Unable to read counts data from DB")
			return 0, nil, err
		}

		account_score, counts_map, decayValue, err := ComputeAccountScore(weights, counts_map, eventTs)
		if err != nil {
			return 0, nil, err
		}
		var uday model.PerUserScoreOnDay
		uday.Id = user_id
		uday.Score = account_score
		uday.Debug = make(map[string]interface{})
		uday.Debug["customer_id"] = id
		uday.Debug["date"] = eventTs
		uday.Debug["counts"] = counts_map
		uday.Debug["weights"] = weights
		uday.Debug["decay"] = decayValue
		usersOnDay = append(usersOnDay, uday)
		final_score += account_score

	}

	return final_score, usersOnDay, nil

}

func ComputeAccountScore(weights model.AccWeights, eventsCount map[string]int64, ts string) (float32, map[string]int64, float64, error) {

	var accountScore float32
	var eventsCountMap map[string]int64 = make(map[string]int64)
	weightValue := weights.WeightConfig
	saleWindow := weights.SaleWindow
	for _, w := range weightValue {
		if ew, ok := eventsCount[w.WeightId]; ok {
			accountScore += float32(ew) * w.Weight_value
			eventsCountMap[w.WeightId] = ew
		}
	}

	decay_value := model.ComputeDecayValue(ts, saleWindow)
	account_score_after_decay := (accountScore - accountScore*float32(decay_value))
	return account_score_after_decay, eventsCountMap, decay_value, nil
}

func ComputeAccountScoreOnLastEvent(weights model.AccWeights, eventsCount map[string]float64) (float64, error) {

	var accountScore float64
	weightValue := weights.WeightConfig
	for _, w := range weightValue {
		if ew, ok := eventsCount[w.WeightId]; ok {
			accountScore += ew * float64(w.Weight_value)
		}
	}
	return accountScore, nil
}
func GetDateOnlyFromTimestamp(ts int64) string {
	year, month, date := time.Unix(ts, 0).UTC().Date()
	data := fmt.Sprintf("%d%02d%02d", year, month, date)
	return data

}

func (store *MemSQL) GetAllUserScore(projectId int64, debug bool) ([]model.AllUsersScore, *model.AccWeights, error) {
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
	stmt := fmt.Sprintf("select id, event_aggregate from users where json_length(event_aggregate)>=1 and project_id= ? LIMIT %d", MAX_LIMIT)
	rows, err := db.Raw(stmt, projectId).Rows()
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	result := make([]model.AllUsersScore, 0)
	for rows.Next() {
		var eventsCountString string
		countsMapDays := map[string]map[string]float64{}
		resultmap := make(map[string]model.PerUserScoreOnDay)
		var userId string
		var r model.AllUsersScore

		err := rows.Scan(&userId, &eventsCountString)
		if err != nil {
			logCtx.WithError(err).Error("Failed to get event counts for user :%s", eventsCountString)
		}

		err = json.Unmarshal([]byte(eventsCountString), &countsMapDays)
		if err != nil {
			logCtx.WithError(err).Errorf("Failed to unmarshall json counts for users per day:%s", eventsCountString)
		}

		for day, countsPerday := range countsMapDays {
			if day != model.LAST_EVENT {
				countsPerday_ := make(map[string]int64)
				for countKey, countVal := range countsPerday_ {
					countsPerday_[countKey] = int64(countVal)
				}

				accountScore, _, _, err := ComputeAccountScore(*weights, countsPerday_, day)
				if err != nil {
					return nil, nil, err
				}

				var t model.PerUserScoreOnDay
				t.Score = float32(accountScore)
				t.Timestamp = day
				if debug {
					t.Debug = make(map[string]interface{})
					t.Debug["counts"] = countsPerday
					t.Debug["date"] = day
				}
				resultmap[day] = t
			} else {

				accountScore, err := ComputeAccountScoreOnLastEvent(*weights, countsPerday)
				if err != nil {
					return nil, nil, err
				}
				var t model.PerUserScoreOnDay
				t.Score = float32(accountScore)
				t.Timestamp = day
				if debug {
					t.Debug = make(map[string]interface{})
					t.Debug["counts"] = countsPerday
					t.Debug["date"] = day
				}
				resultmap[day] = t
			}
		}
		r.UserId = userId
		r.ScorePerDay = resultmap
		result = append(result, r)
	}
	if rows.Err() != nil {
		logCtx.WithError(err).Error("Error while fetching rows to compute all user scores")

		return nil, nil, rows.Err()
	}

	return result, weights, nil
}

func (store *MemSQL) GetAllUserScoreOnDay(projectId int64, ts string, debug bool) ([]model.AllUsersScore, *model.AccWeights, error) {
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
	stmt := fmt.Sprintf("select id, json_extract_json(event_aggregate,?) from users where project_id= ? and json_length(event_aggregate)>=1  LIMIT %d", MAX_LIMIT)
	rows, err := db.Raw(stmt, ts, projectId).Rows()
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	result := make([]model.AllUsersScore, 0)
	for rows.Next() {
		var eventsCountString sql.NullString
		var countsOnDays model.LatestScore
		resultmap := make(map[string]model.PerUserScoreOnDay)
		var userId string
		var r model.AllUsersScore
		err := rows.Scan(&userId, &eventsCountString)
		if err != nil {
			logCtx.WithError(err).Error("Failed to get event counts for groups")
		}

		if !eventsCountString.Valid {
			continue
		}

		err = json.Unmarshal([]byte(eventsCountString.String), &countsOnDays)
		if err != nil {
			logCtx.WithError(err).Errorf("Failed to unmarshall json counts for users per day :%s", eventsCountString.String)
		}

		countsMapDays := make(map[string]map[string]float64)
		countsMapDays[ts] = countsOnDays.EventsCount
		for day, countsPerday := range countsMapDays {
			if day != model.LAST_EVENT {
				ct := make(map[string]int64)
				for k, v := range countsPerday {
					ct[k] = int64(v)
				}

				accountScore, _, _, err := ComputeAccountScore(*weights, ct, day)
				if err != nil {
					return nil, nil, err
				}
				var t model.PerUserScoreOnDay
				t.Score = float32(accountScore)
				t.Timestamp = day
				if debug {
					t.Debug = make(map[string]interface{})
					t.Debug["counts"] = make(map[string]int64, 0)
					t.Debug["counts"] = countsPerday
					t.Debug["date"] = day
				}
				resultmap[day] = t
			}
		}
		r.UserId = userId
		r.ScorePerDay = resultmap
		result = append(result, r)
	}
	if rows.Err() != nil {
		logCtx.WithError(err).Error("Error while fetching rows to compute all user scores")

		return nil, nil, rows.Err()
	}

	return result, weights, nil
}

func (store *MemSQL) GetAllUserScoreLatest(projectId int64, debug bool) ([]model.AllUsersScore, *model.AccWeights, error) {
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
	stmt := fmt.Sprintf("select id, json_extract_json(event_aggregate,?) from users where  project_id= ? and json_length(event_aggregate)>=1  LIMIT %d", MAX_LIMIT)
	rows, err := db.Raw(stmt, model.LAST_EVENT, projectId).Rows()
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	result := make([]model.AllUsersScore, 0)
	for rows.Next() {
		var eventsCountString string
		resultmap := make(map[string]model.PerUserScoreOnDay)
		var userId string
		var r model.AllUsersScore
		var lastday model.LatestScore
		lastday.EventsCount = make(map[string]float64)
		var countsPerDay map[string]float64
		err := rows.Scan(&userId, &eventsCountString)
		if err != nil {
			logCtx.WithError(err).Error("Failed to get event counts for groups")
		}

		err = json.Unmarshal([]byte(eventsCountString), &lastday)
		if err != nil {
			logCtx.WithError(err).Error("Failed to unmarshall json counts for users per day - 2")
		}

		day := GetDateOnlyFromTimestamp(lastday.Date)
		countsPerDay = lastday.EventsCount
		accountScore, err := ComputeAccountScoreOnLastEvent(*weights, countsPerDay)
		if err != nil {
			return nil, nil, err
		}
		var t model.PerUserScoreOnDay
		t.Score = float32(accountScore)
		t.Timestamp = day

		if debug {
			t.Debug = make(map[string]interface{})
			t.Debug["date"] = day
			t.Debug["counts"] = countsPerDay
		}
		resultmap[day] = t

		r.ScorePerDay = resultmap
		r.UserId = userId
		result = append(result, r)
	}
	if rows.Err() != nil {
		logCtx.WithError(err).Error("Error while fetching rows to compute all user scores")

		return nil, nil, rows.Err()
	}

	return result, weights, nil
}

func (store *MemSQL) GetUserScoreOnIds(projectId int64, usersAnonymous, usersNonAnonymous []string, debug bool) (map[string]model.PerUserScoreOnDay, error) {
	// get score of each account on current date

	projectID := projectId
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	MAX_LIMIT := 10000
	logCtx := log.WithFields(logFields)
	db := C.GetServices().Db

	weights, errStatus := store.GetWeightsByProject(projectId)
	if errStatus != http.StatusFound {
		return nil, errors.New("weights not found")
	}

	stmt := fmt.Sprintf("select id, json_extract_json(event_aggregate,'%s') from users where  project_id = ? AND ID IN ( ? ) AND json_length(event_aggregate)>=1 LIMIT %d", model.LAST_EVENT, MAX_LIMIT)
	rows, err := db.Raw(stmt, projectId, usersAnonymous).Rows()
	if err != nil {

		return nil, err
	}
	defer rows.Close()

	result := make(map[string]model.PerUserScoreOnDay, 0)
	for rows.Next() {
		var resultPerUser model.PerUserScoreOnDay
		var latestCountString string
		var le model.LatestScore
		le.EventsCount = make(map[string]float64)

		var userId string
		err := rows.Scan(&userId, &latestCountString)
		if err != nil {
			logCtx.WithError(err).Error("Failed to get event counts for groups")
		}

		err = json.Unmarshal([]byte(latestCountString), &le)
		if err != nil {
			logCtx.WithError(err).Error("Failed to unmarshall json counts for users per day")
		}
		ts := GetDateOnlyFromTimestamp(le.Date)

		accountScore, err := ComputeAccountScoreOnLastEvent(*weights, le.EventsCount)
		if err != nil {
			return nil, err
		}
		resultPerUser.Id = userId
		resultPerUser.Score = float32(accountScore)
		if debug {
			resultPerUser.Debug = make(map[string]interface{})
			resultPerUser.Debug["counts"] = le.EventsCount
			resultPerUser.Debug["date"] = ts
		}
		resultPerUser.Timestamp = ts

		result[userId] = resultPerUser
	}

	resultScoresOnCustomerIds, err := ComputeUserScoreNonAnonymous(db, *weights, projectId, usersNonAnonymous, debug)
	if err != nil {
		logCtx.WithError(err).Error("Failed to compute user score for known users")
		return nil, err
	}
	for k, v := range resultScoresOnCustomerIds {
		result[k] = v
	}
	return result, nil
}

func (store *MemSQL) GetAccountScoreOnIds(projectId int64, accountIds []string, debug bool) (map[string]model.PerUserScoreOnDay, error) {
	// get score of each account on current date

	projectID := projectId
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	db := C.GetServices().Db

	weights, errStatus := store.GetWeightsByProject(projectId)
	if errStatus != http.StatusFound {
		return nil, errors.New("weights not found")
	}

	stmt := fmt.Sprintf("SELECT id, JSON_EXTRACT_JSON(event_aggregate,'%s') from users where  is_group_user=1  and project_id = ? AND ID IN ( ? ) AND JSON_LENGTH(event_aggregate)>=1 LIMIT %d", model.LAST_EVENT, MAX_LIMIT)
	rows, err := db.Raw(stmt, projectId, accountIds).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]model.PerUserScoreOnDay, 0)
	for rows.Next() {
		var resultPerUser model.PerUserScoreOnDay
		var latestCountString string
		var le model.LatestScore
		var userId string
		err := rows.Scan(&userId, &latestCountString)
		if err != nil {
			logCtx.WithError(err).Error("Failed to get event counts for groups")
		}

		err = json.Unmarshal([]byte(latestCountString), &le)
		if err != nil {
			logCtx.WithError(err).Error("Failed to unmarshall json counts for users per day")
		}

		ts := GetDateOnlyFromTimestamp(le.Date)
		resultPerUser.Timestamp = ts

		accountScore, err := ComputeAccountScoreOnLastEvent(*weights, le.EventsCount)
		if err != nil {
			return nil, err
		}
		resultPerUser.Id = userId
		resultPerUser.Score = float32(accountScore)
		if debug {
			resultPerUser.Debug = make(map[string]interface{})
			resultPerUser.Debug["counts"] = le.EventsCount
			resultPerUser.Debug["date"] = ts
		}

		result[userId] = resultPerUser
	}
	if rows.Err() != nil {
		logCtx.WithError(err).Error("Error while fetching rows to compute account scores")
		return nil, rows.Err()
	}

	return result, nil
}

func ComputeUserScoreNonAnonymous(db *gorm.DB, weights model.AccWeights, projectId int64, knownUsers []string, debug bool) (map[string]model.PerUserScoreOnDay, error) {

	projectID := projectId
	logFields := log.Fields{
		"project_id": projectID,
	}
	logCtx := log.WithFields(logFields)

	result := make(map[string]model.PerUserScoreOnDay, 0)

	stmt_customer := fmt.Sprintf("select customer_user_id, json_extract_json(event_aggregate,'%s') from users where project_id=? and customer_user_id IN (?) AND JSON_LENGTH(event_aggregate)>=1 LIMIT %d", model.LAST_EVENT, MAX_LIMIT)
	rows, err := db.Raw(stmt_customer, projectId, knownUsers).Rows()
	if err != nil {
		logCtx.WithError(err).Error("Error while fetching rows to compute user scores")
		return nil, err
	}

	var usercountsmap map[string]model.LatestScore = make(map[string]model.LatestScore)
	for rows.Next() {
		var latestCountString string
		var le model.LatestScore
		var userId string
		err := rows.Scan(&userId, &latestCountString)
		if err != nil {
			logCtx.WithError(err).Error("Failed to get event counts for groups")
		}

		err = json.Unmarshal([]byte(latestCountString), &le)
		if err != nil {
			logCtx.WithError(err).Error("Failed to unmarshall json counts for users per day")
		}
		if lastscore, ok := usercountsmap[userId]; !ok {
			usercountsmap[userId] = le
		} else {
			if le.Date > lastscore.Date {
				usercountsmap[userId] = le
			}
		}
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	for userKey, userCounts := range usercountsmap {
		var r model.PerUserScoreOnDay
		ts := GetDateOnlyFromTimestamp(userCounts.Date)
		r.Timestamp = ts
		accountScore, err := ComputeAccountScoreOnLastEvent(weights, userCounts.EventsCount)
		if err != nil {
			return nil, err
		}
		r.Id = userKey
		r.Score = float32(accountScore)
		if debug {
			r.Debug = make(map[string]interface{})
			r.Debug["counts"] = userCounts.EventsCount
			r.Debug["date"] = userCounts.Date
		}
		result[userKey] = r
	}

	return result, nil
}

func (store *MemSQL) GetAllUserEvents(projectId int64, debug bool) (map[string]map[string]model.LatestScore, error) {
	// get score of each account on current date

	projectID := projectId
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	MAX_LIMIT := 10000
	logCtx := log.WithFields(logFields)
	db := C.GetServices().Db

	stmt := fmt.Sprintf("select id, event_aggregate from users where json_length(event_aggregate)>=1 and project_id= ? LIMIT %d", MAX_LIMIT)
	rows, err := db.Raw(stmt, projectId).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]map[string]model.LatestScore, 0)
	for rows.Next() {
		var eventsCountString string
		var userId string
		countsMapDays := make(map[string]model.LatestScore)
		err := rows.Scan(&userId, &eventsCountString)
		if err != nil {
			logCtx.WithError(err).Error("Failed to get event counts for groups")
		}

		err = json.Unmarshal([]byte(eventsCountString), &countsMapDays)
		if err != nil {
			logCtx.WithError(err).Error("Failed to unmarshall json counts for users per day")
		}

		countsMap := make(map[string]model.LatestScore)
		for day, countsPerday := range countsMapDays {
			var t model.LatestScore
			if day != model.LAST_EVENT {
				t.Date = model.GetDateFromString(day)
				t.EventsCount = make(map[string]float64)
				for k, v := range countsPerday.EventsCount {
					t.EventsCount[k] = float64(v)
				}
			}
			countsMap[day] = t
		}
		result[userId] = countsMap
	}
	if rows.Err() != nil {
		logCtx.WithError(err).Error("Error while fetching rows to compute all user scores")

		return nil, rows.Err()
	}

	return result, nil
}
