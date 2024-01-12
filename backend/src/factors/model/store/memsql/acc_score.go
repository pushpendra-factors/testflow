package memsql

import (
	"database/sql"
	"encoding/json"
	"errors"
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"sync"

	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

const MAX_LIMIT = 10000
const NORM_CONFIG = 10000
const NUM_ROUTINES = 25

func makeUserUpdateString(user map[string]model.LatestScore, score float64) (string, error) {

	updateString := make([]string, 0)
	finalString := ""
	engagement_string := U.GROUP_EVENT_NAME_ENGAGEMENT_SCORE
	for evKey, evCount := range user {
		eventsCountjson, err := json.Marshal(evCount)
		if err != nil {
			log.WithError(err).Errorf("Unable to convert map to json in event counts ")
			return "", err
		}

		if score == -1 {
			ct := fmt.Sprintf("event_aggregate::`%s`='%s'", evKey, string(eventsCountjson))
			updateString = append(updateString, ct)
		} else {

			ct := fmt.Sprintf("event_aggregate::`%s`='%s',properties::`%s`=%f", evKey, string(eventsCountjson), engagement_string, score)
			updateString = append(updateString, ct)
		}
	}

	countTotal := 0.0
	if evCounts, cok := user[model.LAST_EVENT]; cok {

		for _, v := range evCounts.EventsCount {
			countTotal += v
		}
	}

	if countTotal != 0.0 {
		finalString = strings.Join(updateString, ",")
	}

	return finalString, nil
}

func makeAccountUpdateString(user map[string]model.LatestScore, score, allEventsSCore float64, engagement_level string) (string, error) {

	updateString := make([]string, 0)
	finalString := ""
	engagement_string := U.GROUP_EVENT_NAME_ENGAGEMENT_SCORE
	total_engagement_string := U.GROUP_EVENT_NAME_TOTAL_ENGAGEMENT_SCORE
	engagement_level_string := U.GROUP_EVENT_NAME_ENGAGEMENT_LEVEL

	for evKey, evCount := range user {
		eventsCountjson, err := json.Marshal(evCount)
		if err != nil {
			log.WithError(err).Errorf("Unable to convert map to json in event counts ")
			return "", err
		}

		if score == -1 {
			ct := fmt.Sprintf("event_aggregate::`%s`='%s'", evKey, string(eventsCountjson))
			updateString = append(updateString, ct)
		} else {
			ct := fmt.Sprintf("event_aggregate::`%s`='%s',properties::`%s`=%f,properties::`%s`=%f ", evKey, string(eventsCountjson), engagement_string, score, total_engagement_string, allEventsSCore)
			updateString = append(updateString, ct)
		}
	}

	countTotal := 0.0
	if evCounts, cok := user[model.LAST_EVENT]; cok {
		for _, v := range evCounts.EventsCount {
			countTotal += v
		}
	}

	if countTotal != 0.0 {
		engagementLevelStringModified := strings.Join([]string{"$", engagement_level_string}, "")
		ct := fmt.Sprintf("properties::%s='%s'", engagementLevelStringModified, engagement_level)
		updateString = append(updateString, ct)
		finalString = strings.Join(updateString, ",")
	}

	return finalString, nil
}

func (store *MemSQL) UpdateUserEventsCount(projectId int64, evdata map[string]map[string]model.LatestScore) error {
	projectID := projectId
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	db := C.GetServices().Db
	userIDs := make(map[string]int, 0)
	for uid, _ := range evdata {
		userIDs[uid] = 1
	}

	numLines := 0
	for userId, _ := range userIDs {
		var updateCounts map[string]model.LatestScore = make(map[string]model.LatestScore)

		if ev, cok := evdata[userId]; cok {
			for evDate, evCounts := range ev {
				updateCounts[evDate] = evCounts
			}
		}

		updateString, nil := makeUserUpdateString(updateCounts, float64(-1))
		if updateString != "" {
			stmt := fmt.Sprintf("UPDATE users SET event_aggregate = case when event_aggregate is NULL THEN '{}' ELSE  event_aggregate END, %s where id= ? and project_id= ?", updateString)
			err := db.Exec(stmt, userId, projectId).Error
			if err != nil {
				logCtx.WithError(err).Errorf("Unable to update latest score in user events ")
				return err
			}
			numLines += 1
		}
	}
	logCtx.WithField("updates", numLines).Info("Numer of rows (users) written to DB")
	return nil
}

func (store *MemSQL) UpdateUserEventsCountGO(projectId int64, evdata map[string]map[string]model.LatestScore) error {
	logFields := log.Fields{
		"project_id": projectId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	// db := C.GetServices().Db
	numUserRoutines := NUM_ROUTINES
	userIDs := make(map[string]int, 0)
	usersIdsList := make([]string, 0)
	for uid, _ := range evdata {
		userIDs[uid] = 1
	}

	for uid, _ := range userIDs {
		usersIdsList = append(usersIdsList, uid)
	}
	userIDChunks := U.GetStringListAsBatch(usersIdsList, numUserRoutines)
	logCtx.WithField("counts", len(userIDs)).Info("Updating users in DB")

	for _, userIDList := range userIDChunks {
		var wg sync.WaitGroup

		wg.Add(len(userIDList))
		for i := 0; i < len(userIDList); i++ {
			go UpdatePerUser(projectId, evdata, userIDList[i], &wg)
		}
		wg.Wait()
	}
	logCtx.Info("Done updating users in DB")

	return nil

}

func UpdatePerUser(projectId int64, evdata map[string]map[string]model.LatestScore,
	userId string, wg *sync.WaitGroup) error {

	logFields := log.Fields{
		"project_id": projectId,
	}
	logCtx := log.WithFields(logFields)
	defer wg.Done()
	db := C.GetServices().Db

	var updateCounts map[string]model.LatestScore = make(map[string]model.LatestScore)

	if ev, cok := evdata[userId]; cok {
		for evDate, evCounts := range ev {
			updateCounts[evDate] = evCounts
		}
	}

	updateString, nil := makeUserUpdateString(updateCounts, float64(-1))
	if updateString != "" {
		stmt := fmt.Sprintf("UPDATE users SET event_aggregate = case when event_aggregate is NULL THEN '{}' ELSE  event_aggregate END, %s where id= ? and project_id= ?", updateString)
		err := db.Exec(stmt, userId, projectId).Error
		if err != nil {
			logCtx.WithError(err).Errorf("Unable to update latest score in user events ")
			return err
		}
	}

	return nil

}

func UpdatePerAccount(projectId int64, evdata map[string]map[string]model.LatestScore, lastevent map[string]model.LatestScore,
	allEvents map[string]model.LatestScore,
	userId string, weights model.AccWeights, bucket model.BucketRanges, db *gorm.DB, wg *sync.WaitGroup) error {

	logFields := log.Fields{
		"project_id": projectId,
	}
	logCtx := log.WithFields(logFields)
	defer wg.Done()

	var updateCounts map[string]model.LatestScore = make(map[string]model.LatestScore)
	if ev, cok := evdata[userId]; cok {
		for evDate, evCounts := range ev {
			updateCounts[evDate] = evCounts
		}
	}
	if _, lok := lastevent[userId]; lok {
		lastEvent := lastevent[userId]
		allEvent := allEvents[userId]
		updateCounts[model.LAST_EVENT] = lastEvent
		score, err := ComputeAccountScoreOnLastEvent(projectId, weights, lastEvent.EventsCount)
		if err != nil {
			logCtx.WithField("last event ", lastevent).Error("unable to compute final score ")
		}
		allEventscore, err := ComputeAccountScoreOnLastEvent(projectId, weights, allEvent.EventsCount)
		if err != nil {
			logCtx.WithField("last event ", lastevent).Error("unable to compute final score ")
		}
		engagementLevel := model.GetEngagement(allEventscore, bucket)

		updateString, nil := makeAccountUpdateString(updateCounts, score, allEventscore, engagementLevel)
		if updateString != "" {
			stmt := fmt.Sprintf("UPDATE users SET event_aggregate = case when event_aggregate is NULL THEN '{}' ELSE  event_aggregate END, %s where id= ? and project_id= ?", updateString)
			err := db.Exec(stmt, userId, projectId).Error
			if err != nil {
				logCtx.WithError(err).WithField("statement ", stmt).Errorf("Unable to update latest score in user events ")
				return err
			}
		}

	}

	return nil

}

func (store *MemSQL) UpdateGroupEventsCountGO(projectId int64, evdata map[string]map[string]model.LatestScore,
	lastevent map[string]model.LatestScore, allEvents map[string]model.LatestScore, weights model.AccWeights, bucket model.BucketRanges) error {
	logFields := log.Fields{
		"project_id": projectId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	db := C.GetServices().Db
	numUserRoutines := NUM_ROUTINES
	userIDs := make(map[string]int, 0)
	usersIdsList := make([]string, 0)
	for uid, _ := range evdata {
		userIDs[uid] = 1
	}
	for uid, _ := range lastevent {
		userIDs[uid] = 1
	}
	for uid, _ := range allEvents {
		userIDs[uid] = 1
	}
	for uid, _ := range userIDs {
		usersIdsList = append(usersIdsList, uid)
	}
	userIDChunks := U.GetStringListAsBatch(usersIdsList, numUserRoutines)
	logCtx.WithField("counts :", len(userIDs)).Info("Updating groups in DB")

	for _, userIDList := range userIDChunks {
		var wg sync.WaitGroup

		wg.Add(len(userIDList))
		for i := 0; i < len(userIDList); i++ {
			go UpdatePerAccount(projectId, evdata, lastevent, allEvents, userIDList[i], weights, bucket, db, &wg)
		}
		wg.Wait()
	}
	logCtx.Info("Done Updating groups in DB")

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
		var counts_map model.LatestScore
		var account_score float64
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

		currentDate := U.GetDateOnlyFromTimestamp(counts_map.Date)
		account_score = ComputeScoreWithWeightsAndCounts(projectId, weights, counts_map.EventsCount, currentDate)
		r.Id = account_id
		r.Score = float32(account_score)
		r.Timestamp = ts // have to fill the right timestamp
		if debug {
			r.Debug = make(map[string]interface{})
			r.Debug["counts"] = counts_map
			r.Debug["score"] = account_score
			r.Debug["timestamp"] = ts
		}
		result = append(result, r)
	}

	return result, weights, nil
}

func (store *MemSQL) GetPerAccountScore(projectId int64, timestamp string, userId string, numDaysToTrend int, debug bool) (model.PerAccountScore, *model.AccWeights, string, error) {
	// get score of each account on current date

	projectID := projectId
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	var rt string
	var fscore float64
	var result model.PerAccountScore
	countsMapDays := make(map[string]model.LatestScore)
	scoreOnDays := make(map[string]float32)
	resultmap := make(map[string]model.PerAccountScore)

	currentDate := U.GetDateFromString(timestamp)
	prevDateTotrend := time.Unix(currentDate, 0).AddDate(0, 0, -1*numDaysToTrend).Unix()

	db := C.GetServices().Db

	weights, errStatus := store.GetWeightsByProject(projectId)
	if errStatus != http.StatusFound {
		return result, nil, "", fmt.Errorf("unable to get weights for project %d", projectId)
	}

	stmt := "select event_aggregate from users where id=? and project_id=?"
	tx := db.Raw(stmt, userId, projectId).Row()
	err := tx.Scan(&rt)
	if err != nil {
		logCtx.WithError(err).Error("Unable to read user counts data from DB")
		return result, nil, "", err
	}

	err = json.Unmarshal([]byte(rt), &countsMapDays)
	if err != nil {
		logCtx.WithError(err).Error("Failed to unmarshall json counts for users per day")
	}

	resultmap, scoreOnDays, fscore, err = CalculatescoresPerAccount(projectId, weights, currentDate, prevDateTotrend, countsMapDays)
	if err != nil {
		return model.PerAccountScore{}, nil, "", err
	}

	if debug {
		result.Debug = make(map[string]interface{})
		result.Debug["info"] = resultmap
	}

	ev := countsMapDays[model.LAST_EVENT]
	dateString := U.GetDateOnlyFromTimestamp(ev.Date)

	buckets, errOnbuckets := store.GetEngagementBucketsOnProject(projectId, dateString)
	if errOnbuckets != nil {
		logCtx.WithError(errOnbuckets).Error("unable to get bucket ranges from DB")

	}

	engagementLevel := model.GetEngagement(fscore, buckets)
	result.Id = userId
	result.Trend = make(map[string]float32)
	result.Trend = scoreOnDays
	result.Score = float32(fscore)
	result.Timestamp = timestamp
	return result, weights, engagementLevel, nil
}

func (store *MemSQL) GetEngagementBucketsOnProject(projectId int64, timestamp string) (model.BucketRanges, error) {
	// get score of each account on current date

	projectID := projectId
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	db := C.GetServices().Db

	var rawDateString string
	var rawBucketString string
	var result model.BucketRanges
	stmt := "select date,bucket from account_scoring_ranges where  project_id=? and date=?"
	tx := db.Raw(stmt, projectId, timestamp).Row()
	err := tx.Scan(&rawDateString, &rawBucketString)
	if err != nil {
		logCtx.WithError(err).Error("Unable to read bucket ranges data from DB")
		return model.BucketRanges{}, err
	}
	result.Date = rawDateString
	result.Ranges = make([]model.Bucket, 0)
	err = json.Unmarshal([]byte(rawBucketString), &result.Ranges)
	if err != nil {
		logCtx.WithError(err).Error("Unable to unmarshall bucket ranges ")
		return model.BucketRanges{}, err
	}

	return result, nil

}

func ComputeTrendWrapper(projectId int64, currentDate int64, countsMapDays map[string]model.LatestScore, weights *model.AccWeights) (map[string]float32, error) {

	scoreOnDays := make(map[string]float32)
	salewindow := weights.SaleWindow
	currentDateUnix := time.Unix(currentDate, 0)
	periodStartDate := currentDateUnix.AddDate(0, 0, -1*int(salewindow))

	for dayIdx := salewindow; dayIdx > 0; dayIdx-- {
		periodDateInit := periodStartDate.AddDate(0, 0, int(dayIdx)).Unix()
		dateString := U.GetDateOnlyFromTimestamp(periodDateInit)
		orderedDayString := U.GenDateStringsForLastNdays(periodDateInit, weights.SaleWindow)
		score := computeTrendInPeriod(projectId, periodDateInit, orderedDayString, countsMapDays, weights)
		scoreOnDays[dateString] = float32(score)
	}

	return scoreOnDays, nil
}

func computeTrendInPeriod(projectId int64, startDate int64, dates []string, countsMapDays map[string]model.LatestScore, weights *model.AccWeights) float64 {
	score := float64(0)
	for _, dayString := range dates {
		endDateUnix := U.GetDateFromString(dayString)
		countsPerday, ok := countsMapDays[dayString]
		if !ok {
			countsPerday.EventsCount = make(map[string]float64)
		}
		decay := model.ComputeDecayValueGivenStartEndTS(endDateUnix, startDate, weights.SaleWindow)
		accountScore := ComputeScoreWithWeightsAndCountsWithDecay(projectId, weights, countsPerday.EventsCount, decay)
		score += accountScore

	}
	return score

}

func CalculatescoresPerAccount(projectId int64, weights *model.AccWeights, currentDate int64, prevDateTotrend int64, countsMapDays map[string]model.LatestScore) (map[string]model.PerAccountScore,
	map[string]float32, float64, error) {

	var fscore float32
	var resultmap map[string]model.PerAccountScore = make(map[string]model.PerAccountScore)
	scoreOnDays := make(map[string]float32)

	accountScoreMap, err := ComputeTrendWrapper(projectId, currentDate, countsMapDays, weights)
	if err != nil {
		return nil, nil, -1, err
	}

	orderedDayString := U.GenDateStringsForLastNdays(currentDate, weights.SaleWindow)
	countsInInt := make(map[string]float64)
	for _, day := range orderedDayString {
		accountScore := accountScoreMap[day]
		if day != model.LAST_EVENT {
			var t model.PerAccountScore
			fscore = float32(normalize_score(float64(accountScore)))
			t.Score = fscore
			t.Timestamp = day
			unixDay := U.GetDateFromString(day)
			if unixDay >= prevDateTotrend {
				scoreOnDays[day] = fscore
				if debug {
					t.Debug = make(map[string]interface{})
					t.Debug["counts"] = make(map[string]int64, 0)
					t.Debug["counts"] = countsInInt
					t.Debug["date"] = day
					t.Debug["score"] = accountScore
					t.Debug["normscore"] = fscore
				}
			}
			resultmap[day] = t
		}
	}

	return resultmap, scoreOnDays, float64(fscore), nil

}

func OrderCountDays(countDays map[string]model.LatestScore) []string {

	var orderedDays []string = make([]string, len(countDays))
	dayUnix := make([]int64, len(countDays))
	daymap := make(map[int64]string, 0)

	idx := 0
	for dayString, _ := range countDays {
		d := U.GetDateFromString(dayString)
		dayUnix[idx] = d
		idx += 1
		daymap[d] = dayString
	}
	sort.Slice(dayUnix, func(i, j int) bool { return dayUnix[i] < dayUnix[j] })
	for idx, ts := range dayUnix {
		orderedDays[idx] = daymap[ts]
	}

	return orderedDays

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
	countsMap := map[string]float64{}
	err = json.Unmarshal([]byte(rt), &countsMap)
	if err != nil {
		logCtx.WithError(err).Error("Unable to unamrshall user counts data")
		return result, err
	}
	accountScore := ComputeScoreWithWeightsAndCounts(projectId, weights, countsMap, eventTS)
	result.Id = userId
	result.Score = float32(accountScore)
	result.Timestamp = eventTS
	if debug {
		result.Debug = make(map[string]interface{})
		result.Debug["counts"] = countsMap
		result.Debug["weights"] = weights
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
	var weightNamesMap map[string]string = make(map[string]string, 0)

	for _, w := range weights.WeightConfig {
		if len(w.FilterName) > 0 {
			weightNamesMap[w.WeightId] = w.FilterName
		} else {
			weightNamesMap[w.WeightId] = w.EventName
		}
	}
	for tx.Next() {
		var events_count_string string
		var le model.LatestScore
		var user_id string
		err := tx.Scan(&user_id, &events_count_string)
		if err != nil {
			logCtx.WithError(err).Error("Failed to get event counts for groups")
		}
		err = json.Unmarshal([]byte(events_count_string), &le)
		if err != nil {
			logCtx.WithError(err).Error("Unable to read counts data from DB")
			return 0, nil, err
		}

		date := U.GetDateOnlyFromTimestamp(le.Date)
		accountScore := ComputeScoreWithWeightsAndCounts(projectId, &weights, le.EventsCount, date)

		var uday model.PerUserScoreOnDay
		uday.Id = user_id
		uday.Score = float32(accountScore)
		addProperty(&uday, &le)
		addTopEvents(&uday, &le, weightNamesMap)
		uday.Debug = make(map[string]interface{})
		uday.Debug["customer_id"] = id
		uday.Debug["date"] = eventTs
		uday.Debug["counts"] = le.EventsCount
		uday.Debug["weights"] = weights
		usersOnDay = append(usersOnDay, uday)
		final_score += float32(accountScore)

	}

	return final_score, usersOnDay, nil

}

func ComputeAccountScore(weights model.AccWeights, eventsCount map[string]int64, ts string) (float32, map[string]int64, float64, error) {

	var accountScore float32
	var eventsCountMap map[string]int64 = make(map[string]int64)
	weightValue := weights.WeightConfig
	saleWindow := weights.SaleWindow
	decay_value := model.ComputeDecayValue(ts, saleWindow)

	for _, w := range weightValue {
		if ew, ok := eventsCount[w.WeightId]; ok {
			accountScore += float32(ew) * w.Weight_value
			eventsCountMap[w.WeightId] = ew
		}
	}

	account_score_after_decay := accountScore * float32(decay_value)
	return account_score_after_decay, eventsCountMap, decay_value, nil
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

	var weightNamesMap map[string]string = make(map[string]string, 0)

	for _, w := range weights.WeightConfig {
		if len(w.FilterName) > 0 {
			weightNamesMap[w.WeightId] = w.FilterName
		} else {
			weightNamesMap[w.WeightId] = w.EventName
		}
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
		ts := U.GetDateOnlyFromTimestamp(le.Date)

		accountScore, _ := ComputeAccountScoreOnLastEvent(projectId, *weights, le.EventsCount)
		accountScore = normalize_score(accountScore)

		resultPerUser.Id = userId
		resultPerUser.Score = float32(accountScore)
		addProperty(&resultPerUser, &le)
		addTopEvents(&resultPerUser, &le, weightNamesMap)
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

	var weightNamesMap map[string]string = make(map[string]string, 0)

	weights, errStatus := store.GetWeightsByProject(projectId)
	if errStatus != http.StatusFound {
		return nil, errors.New("weights not found")
	}

	for _, w := range weights.WeightConfig {
		if len(w.FilterName) > 0 {
			weightNamesMap[w.WeightId] = w.FilterName
		} else {
			weightNamesMap[w.WeightId] = w.EventName
		}
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

		ts := U.GetDateOnlyFromTimestamp(le.Date)
		resultPerUser.Timestamp = ts

		accountScore, _ := ComputeAccountScoreOnLastEvent(projectId, *weights, le.EventsCount)
		accountScore = normalize_score(accountScore)

		resultPerUser.Id = userId
		resultPerUser.Score = float32(accountScore)
		resultPerUser.TopEvents = le.TopEvents
		addProperty(&resultPerUser, &le)
		addTopEvents(&resultPerUser, &le, weightNamesMap)
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
	var weightNamesMap map[string]string = make(map[string]string, 0)

	for _, w := range weights.WeightConfig {
		if len(w.FilterName) > 0 {
			weightNamesMap[w.WeightId] = w.FilterName
		} else {
			weightNamesMap[w.WeightId] = w.EventName
		}
	}

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
		ts := U.GetDateOnlyFromTimestamp(userCounts.Date)
		r.Timestamp = ts
		accountScore := ComputeScoreWithWeightsAndCounts(projectId, &weights, userCounts.EventsCount, ts)
		accountScore = normalize_score(accountScore)

		r.Id = userKey
		r.Score = float32(accountScore)
		addProperty(&r, &userCounts)
		addTopEvents(&r, &userCounts, weightNamesMap)
		if debug {
			r.Debug = make(map[string]interface{})
			r.Debug["counts"] = userCounts.EventsCount
			r.Debug["date"] = userCounts.Date
		}
		result[userKey] = r
	}

	return result, nil
}

func (store *MemSQL) WriteScoreRanges(projectId int64, buckets []model.BucketRanges) error {
	projectID := projectId
	logFields := log.Fields{
		"project_id": projectID,
	}
	logCtx := log.WithFields(logFields)
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db
	for _, bucket := range buckets {
		if len(bucket.Ranges) > 0 {
			transTime := gorm.NowFunc()
			var bucketObj model.AccountScoreRanges

			rangesjson, err := U.EncodeStructTypeToPostgresJsonb(bucket.Ranges)
			if err != nil {
				logCtx.WithField("entity", bucket).WithError(err).Error("bucket ranhes  conversion to Jsonb failed")
			}
			bucketObj.Buckets = rangesjson
			bucketObj.ProjectID = projectId
			bucketObj.Date = bucket.Date
			bucketObj.CreatedAt = transTime
			bucketObj.UpdatedAt = transTime
			if rangesjson != nil {
				stmt := "INSERT INTO `account_scoring_ranges` (`project_id`,`date`,`bucket`,`created_at`,`updated_at`) VALUES (?,?,?,?,?) ON DUPLICATE KEY UPDATE bucket = ?, updated_at=?"
				err = db.Exec(stmt, bucketObj.ProjectID, bucketObj.Date, bucketObj.Buckets, bucketObj.CreatedAt, bucketObj.UpdatedAt, bucketObj.Buckets, bucketObj.UpdatedAt).Error
				if err != nil {
					log.WithField("entity", bucketObj).WithError(err).Error("Create Failed")
					return err
				}
			}
		}
	}

	return nil
}

func (store *MemSQL) GetAllUserEvents(projectId int64, debug bool) (map[string]map[string]model.LatestScore, map[string]int, error, int64) {
	// get score of each account on current date

	projectID := projectId
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	MAX_LIMIT := 100000
	logCtx := log.WithFields(logFields)
	db := C.GetServices().Db
	startTs := time.Now()

	stmt := fmt.Sprintf("select id,is_group_user, event_aggregate from users where json_length(event_aggregate)>=1 and project_id= ? LIMIT %d", MAX_LIMIT)
	rows, err := db.Raw(stmt, projectId).Rows()
	if err != nil {
		return nil, nil, err, 0
	}
	defer rows.Close()

	result := make(map[string]map[string]model.LatestScore, 0)
	group_user := make(map[string]int)
	for rows.Next() {
		var eventsCountString string
		var userId string
		var is_group_user_null sql.NullBool

		countsMapDays := make(map[string]model.LatestScore)
		err := rows.Scan(&userId, &is_group_user_null, &eventsCountString)
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
				t.Date = U.GetDateFromString(day)
				t.EventsCount = make(map[string]float64)
				for k, v := range countsPerday.EventsCount {
					t.EventsCount[k] = float64(v)
				}
			}
			countsMap[day] = t
		}
		result[userId] = countsMap

		is_group_user := U.IfThenElse(is_group_user_null.Valid, is_group_user_null.Bool, false).(bool)

		if is_group_user {
			group_user[userId] = 1
		} else {
			group_user[userId] = 0
		}
	}
	if rows.Err() != nil {
		logCtx.WithError(err).Error("Error while fetching rows to compute all user scores")
		return nil, nil, rows.Err(), 0
	}
	endTS := time.Now()
	timeDiff := endTS.Sub(startTs).Abs().Milliseconds()

	return result, group_user, nil, timeDiff
}

func addProperty(result *model.PerUserScoreOnDay, le *model.LatestScore) {
	result.Property = make(map[string][]string)
	for propKey, propVal := range le.Properties {
		if _, pok := result.Property[propKey]; !pok {
			result.Property[propKey] = make([]string, 0)
		}
		for pk, _ := range propVal {
			result.Property[propKey] = append(result.Property[propKey], pk)
		}
	}
}

func addTopEvents(result *model.PerUserScoreOnDay, le *model.LatestScore, weightsMap map[string]string) {
	result.TopEvents = make(map[string]float64)
	for propKey, propVal := range le.TopEvents {
		filtername, ok := weightsMap[propKey]
		if !ok {
			filtername = propKey
		}
		result.TopEvents[filtername] += propVal
	}
}

func ComputeScoreWithWeightsAndCounts(projectId int64, weights *model.AccWeights, counts map[string]float64, day string) float64 {

	var score float64
	decay := model.ComputeDecayValue(day, weights.SaleWindow)
	accountScoref, _ := ComputeAccountScoreOnLastEvent(projectId, *weights, counts)
	score = decay * accountScoref
	return score
}

func ComputeScoreWithWeightsAndCountsWithDecay(projectId int64, weights *model.AccWeights, counts map[string]float64, decay float64) float64 {

	var score float64
	accountScoref, _ := ComputeAccountScoreOnLastEvent(projectId, *weights, counts)
	score = decay * accountScoref
	return score
}

func normalize_score(x float64) float64 {
	//100 * tanh(x/100000) range 0 to 100000
	// val := x / float64(NORM_CONFIG)
	// return 100 * MM.Tanh(val)
	return x
}

func ComputeAccountScoreOnLastEvent(project_id int64, weights model.AccWeights, eventsCount map[string]float64) (float64, error) {

	var accountScore float64
	weightValue := weights.WeightConfig
	for _, w := range weightValue {
		if !w.Is_deleted {
			if ew, ok := eventsCount[w.WeightId]; ok {
				accountScore += ew * float64(w.Weight_value)
			}
		}
	}
	return accountScore, nil
}
