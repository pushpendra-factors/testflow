package memsql

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"

	"errors"
	C "factors/config"
	"factors/model/model"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
)

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

		ts_unix := model.GetDateFromString(ev.DateStamp)
		var lt model.LatestScore
		lt.Date = ts_unix
		lt.EventsCount = ev.EventScore
		LatestScoreString, err := json.Marshal(lt)
		if err != nil {
			logCtx.WithError(err).Errorf("Unable to marshall latest score ")
		}

		stmt := fmt.Sprintf("UPDATE users SET event_aggregate = case when event_aggregate is NULL THEN '{}' ELSE  event_aggregate END, event_aggregate::`%s`='%s' ,event_aggregate::%s = CASE WHEN event_aggregate::%s IS NULL OR json_extract_json(event_aggregate,'%s','Date') <= %d THEN '%s' ELSE json_extract_json(event_aggregate,'%s') END where id= ? and project_id= ?", ev.DateStamp, eventsCountjson, model.LAST_EVENT, model.LAST_EVENT, model.LAST_EVENT, ts_unix, LatestScoreString, model.LAST_EVENT)
		err = db.Exec(stmt, ev.UserId, ev.ProjectId).Error
		if err != nil {
			logCtx.WithError(err).Errorf("Unable to update latest score in user events ")
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

			var lt model.LatestScore
			ts := model.GetDateFromString(ev.DateStamp)
			lt.Date = ts
			lt.EventsCount = ev.EventScore

			LatestScoreString, err := json.Marshal(lt)
			if err != nil {
				logCtx.WithError(err).Errorf("Unable to marshall latest score ")
			}

			stmt := fmt.Sprintf("UPDATE users SET event_aggregate = case when event_aggregate is NULL THEN '{}' ELSE  event_aggregate END, event_aggregate::`%s`='%s' ,event_aggregate::%s = CASE WHEN event_aggregate::%s IS NULL OR json_extract_json(event_aggregate,'%s','Date') <= %d THEN '%s' ELSE json_extract_json(event_aggregate,'%s') END where id= ? and project_id= ?", ev.DateStamp, eventsCountjson, model.LAST_EVENT, model.LAST_EVENT, model.LAST_EVENT, ts, LatestScoreString, model.LAST_EVENT)
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

func (store *MemSQL) GetAccountsScore(projectId int64, groupId int, ts string, debug bool) ([]model.PerAccountScore, error) {
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
		return nil, err
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
			return nil, err
		}

		account_score, counts_map, _, err = ComputeAccountScore(*weights, counts_map, ts)
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
	logCtx.WithField("weights", weights).Info("weights")
	if !is_anonymous {
		logCtx.Info("anonymous")

		acc_score, err := ComputeUserScoreOnCustomerId(db, userId, projectID, eventTS, *weights)
		if err != nil {
			return result, err
		}
		result.Id = userId
		result.Score = acc_score
		result.Timestamp = eventTS
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
	accountScore, countsMap, _, err := ComputeAccountScore(*weights, countsMap, eventTS)
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
		counts_map := map[string]int64{}
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

		account_score, _, _, err = ComputeAccountScore(weights, counts_map, eventTs)
		if err != nil {
			return 0, err
		}
		final_score += account_score

	}

	return final_score, nil

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

	decay_value := ComputeDecayValue(ts, saleWindow)

	account_score_after_decay := (accountScore - accountScore*float32(decay_value))

	return account_score_after_decay, eventsCountMap, decay_value, nil
}

func GetDateOnlyFromTimestamp(ts int64) string {
	year, month, date := time.Unix(ts, 0).UTC().Date()
	data := fmt.Sprintf("%d%02d%02d", year, month, date)
	return data

}

func GetDateFromString(ts string) int64 {
	year := ts[0:4]
	month := ts[4:6]
	date := ts[6:8]
	t, _ := strconv.ParseInt(year, 10, 64)
	t1, _ := strconv.ParseInt(month, 10, 64)
	t2, _ := strconv.ParseInt(date, 10, 64)
	t3 := time.Date(int(t), time.Month(t1), int(t2), 0, 0, 0, 0, time.UTC).Unix()
	return t3
}

func ComputeDayDifference(ts1 int64, ts2 int64) int {

	t1 := time.Unix(ts1, 0)
	day1 := t1.YearDay()

	t2 := time.Unix(ts2, 0)
	day2 := t2.YearDay()

	return day2 - day1
}

func ComputeDecayValue(ts string, SaleWindow int64) float64 {
	var decay float64
	// get current date
	currentTS := time.Now().Unix()
	EventTs := GetDateFromString(ts)
	// get difference in weeks
	dayDiff := ComputeDayDifference(currentTS, EventTs)

	if int64(dayDiff) > SaleWindow {
		return 1
	}
	// get decay value
	decay = float64(float64(int64(dayDiff)) / float64(SaleWindow))
	return decay
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
	stmt := fmt.Sprintf("select id, event_aggregate from users where json_length(event_aggregate)>=1 and project_id= ? LIMIT %d", MAX_LIMIT)
	rows, err := db.Raw(stmt, projectId).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]model.AllUsersScore, 0)
	for rows.Next() {
		var eventsCountString string
		countsMapDays := map[string]map[string]int64{}
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
			accountScore, _, _, err := ComputeAccountScore(*weights, countsPerday, day)
			if err != nil {
				return nil, err
			}
			if day != model.LAST_EVENT {
				resultmap[day] = float64(accountScore)
			}
		}
		r.UserId = userId
		r.ScorePerDay = resultmap
		result = append(result, r)
	}
	if rows.Err() != nil {
		logCtx.WithError(err).Error("Error while fetching rows to compute all user scores")

		return nil, rows.Err()
	}

	return result, nil
}

func (store *MemSQL) GetAllUserScoreOnDay(projectId int64, ts string, debug bool) ([]model.AllUsersScore, error) {
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
		return nil, err
	}
	defer rows.Close()

	result := make([]model.AllUsersScore, 0)
	for rows.Next() {
		var eventsCountString string
		countsMapDays := map[string]map[string]int64{}
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
			accountScore, _, _, err := ComputeAccountScore(*weights, countsPerday, day)
			if err != nil {
				return nil, err
			}
			if day != model.LAST_EVENT {
				resultmap[day] = float64(accountScore)
			}
		}
		r.UserId = userId
		r.ScorePerDay = resultmap
		result = append(result, r)
	}
	if rows.Err() != nil {
		logCtx.WithError(err).Error("Error while fetching rows to compute all user scores")

		return nil, rows.Err()
	}

	return result, nil
}

func (store *MemSQL) GetAllUserScoreLatest(projectId int64, debug bool) ([]model.AllUsersScore, error) {
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
		return nil, err
	}
	defer rows.Close()

	result := make([]model.AllUsersScore, 0)
	for rows.Next() {
		var eventsCountString string
		resultmap := make(map[string]float64)
		var userId string
		var r model.AllUsersScore
		var lastday model.LatestScore
		err := rows.Scan(&userId, &eventsCountString)
		if err != nil {
			logCtx.WithError(err).Error("Failed to get event counts for groups")
		}

		err = json.Unmarshal([]byte(eventsCountString), &lastday)
		if err != nil {
			logCtx.WithError(err).Error("Failed to unmarshall json counts for users per day")
		}

		day := GetDateOnlyFromTimestamp(lastday.Date)
		countsPerDay := lastday.EventsCount
		accountScore, _, _, err := ComputeAccountScore(*weights, countsPerDay, day)
		if err != nil {
			return nil, err
		}
		resultmap[day] = float64(accountScore)

		r.ScorePerDay = resultmap
		r.UserId = userId
		r.Debug = countsPerDay
		result = append(result, r)
	}
	if rows.Err() != nil {
		logCtx.WithError(err).Error("Error while fetching rows to compute all user scores")

		return nil, rows.Err()
	}

	return result, nil
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

		accountScore, _, _, err := ComputeAccountScore(*weights, le.EventsCount, ts)
		if err != nil {
			return nil, err
		}
		resultPerUser.Id = userId
		resultPerUser.Score = accountScore
		if debug {
			resultPerUser.Debug = le.EventsCount
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
	MAX_LIMIT := 10000
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

		accountScore, _, _, err := ComputeAccountScore(*weights, le.EventsCount, ts)
		if err != nil {
			return nil, err
		}
		resultPerUser.Id = userId
		resultPerUser.Score = accountScore
		if debug {
			resultPerUser.Debug = le.EventsCount
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

	stmt_customer := fmt.Sprintf("select customer_user_id, json_extract_json(event_aggregate,'%s') from users where customer_user_id IN (?) and project_id=?", model.LAST_EVENT)
	rows, err := db.Raw(stmt_customer, knownUsers, projectId).Rows()
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
		accountScore, _, _, err := ComputeAccountScore(weights, userCounts.EventsCount, ts)
		if err != nil {
			return nil, err
		}
		r.Id = userKey
		r.Score = accountScore
		if debug {
			r.Debug = userCounts.EventsCount
		}
		result[userKey] = r
	}

	return result, nil
}
