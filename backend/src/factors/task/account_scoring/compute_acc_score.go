package accountscoring

import (
	"bufio"
	"encoding/json"
	C "factors/config"
	"factors/filestore"
	"factors/model/model"
	M "factors/model/model"
	"factors/model/store"
	MS "factors/model/store/memsql"

	P "factors/pattern"
	serviceDisk "factors/services/disk"
	U "factors/util"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

const TAKETOPK int = 5

type AggEventsOnUserAndGroup struct {
	User_id     string                   `json:"uid"`
	EventsCount map[string]*M.EventAgg   `json:"eg"`
	Properties  map[string]PropAggregate `json:"prpAgg"`
	Is_group    bool                     `json:"ig"`
}

type PropAggregate struct {
	Name       string           `json:"name"`
	Properties map[string]int64 `json:"prop"`
}

type AggEventsOnGroupsPerProj struct {
	Account_id       int64                                        `json:"aid"`
	GroupEventsCount map[string]map[string]map[string]*M.EventAgg `json:"grp"`
	TimeStamp        int64                                        `json:"ets"`
}

func BuildAccScoringDaily(projectId int64, configs map[string]interface{}) (map[string]interface{}, bool) {

	logCtx := log.WithField("projectId", projectId)
	var prevCountsOfUser map[string]map[string]M.LatestScore = make(map[string]map[string]M.LatestScore)
	status := make(map[string]interface{})
	DurationMap := make(map[string]int64)
	modelCloudManager := configs["modelCloudManager"].(*filestore.FileManager)
	archiveCloudManager := configs["archiveCloudManager"].(*filestore.FileManager)
	diskManager := configs["diskManager"].(*serviceDisk.DiskDriver)
	day_time := configs["day_timestamp"].(int64)
	lookback := configs["lookback"].(int)
	start_time := time.Unix(day_time, 0)
	end_time := start_time.AddDate(0, 0, -1*lookback)

	updatedUsers := make(map[string]map[string]M.LatestScore)
	updatedGroups := make(map[string]map[string]M.LatestScore)
	tm := end_time
	db := C.GetServices().Db

	weights, err := GetWeightfromDB(projectId)
	if err != nil {
		status["err"] = "unable to get weights from DB "
		return status, false

	}
	salewindow := weights.SaleWindow

	//check if newly added project
	//if acc scoring not run before set lookback date to two times salewindow
	DurationMap["fetchDataAllusers"] = int64(0)
	prevCountsOfUser, groupUserMap, err, fetchTime := store.GetStore().GetAllUserEvents(projectId, false)
	DurationMap["fetchDataAllusers"] += fetchTime
	if err != nil {
		status["err"] = "unable to check prev counts of user "
		return status, false
	}

	if len(prevCountsOfUser) == 0 {
		logCtx.Infof("Pulling 2 times daily events, as project is empty")
		lookback := int(2 * salewindow)
		end_time := start_time.AddDate(0, 0, -1*lookback)
		tm = end_time
	}
	for _, userEventMap := range prevCountsOfUser {
		delete(userEventMap, model.LAST_EVENT)
	}

	mweights, _ := CreateweightMap(weights)
	for tm.Unix() <= start_time.Unix() {
		dateString := U.GetDateOnlyFromTimestamp(tm.Unix())
		logCtx.Infof("Pulling daily events for :%s", dateString)
		err = AggDailyUsersOnDate(projectId, archiveCloudManager, modelCloudManager,
			diskManager, tm, db, mweights, updatedUsers, updatedGroups,
			salewindow, DurationMap)
		if err != nil {
			logCtx.WithError(err).Errorf("Error in aggregating users : %d time: %v ", tm)
		}
		tm = tm.AddDate(0, 0, 1)
	}

	// from prevusers pick groups which are avaialable
	now := time.Now()
	updatedGroupsWithAllDates, err := WriteUserCountsAndRangesToDB(projectId, now.Unix(), prevCountsOfUser, updatedUsers, updatedGroups, groupUserMap, *weights, salewindow)
	if err != nil {
		logCtx.WithError(err).Errorf("error in updating user counts to DB")
	}

	logCtx.Infof("Number of updated groups:%d", len(updatedGroups))
	err = ComputeAndWriteRangesToDB(projectId, now.Unix(), int64(lookback), updatedGroupsWithAllDates, weights)
	if err != nil {
		logCtx.WithError(err).Errorf("error in updating ranges DB")
	}

	timeDiff := time.Since(now).Milliseconds()
	DurationMap["updateDB"] += timeDiff
	logCtx.WithField("stats:", DurationMap).Info("stats or time in millliseconds")
	return status, true
}

func GetWeightfromDB(projectId int64) (*M.AccWeights, error) {

	// get weights from DB
	weights, errStatus := store.GetStore().GetWeightsByProject(projectId)
	if errStatus != http.StatusFound {
		errorString := fmt.Sprintf("unable to get weights : %d", projectId)
		log.WithField("err code", errStatus).Errorf(errorString)
		return nil, fmt.Errorf("%s", errorString)
	}
	return weights, nil
}

// CreateweightMap creates a map of event name to list of rules for easy lookup
func CreateweightMap(w *M.AccWeights) (map[string][]M.AccEventWeight, int64) {
	var mwt map[string][]M.AccEventWeight = make(map[string][]M.AccEventWeight)
	// create map of event name to rules for easy lookup
	wt := w.WeightConfig
	for _, e := range wt {
		weightkey := e.EventName
		if _, ok := mwt[weightkey]; !ok {
			mwt[weightkey] = make([]M.AccEventWeight, 0)
		}
		mwt[weightkey] = append(mwt[weightkey], e)
	}
	return mwt, w.SaleWindow
}

// AggDailyUsersOnDate main entry function to aggregate users and group perday and write to file and DB
func AggDailyUsersOnDate(projectId int64, archiveCloudManager, modelCloudManager *filestore.FileManager,
	diskManager *serviceDisk.DiskDriver, date time.Time, db *gorm.DB, weights map[string][]M.AccEventWeight,
	updatedUsers, updatedGroups map[string]map[string]M.LatestScore, salewindow int64, duration map[string]int64) error {

	_, dok := duration["agg"]
	if !dok {
		duration["agg"] = int64(0)
	}
	day_ts := time.Unix(date.Unix(), 0)
	log.Infof("Aggregating  users :%d for %v", projectId, day_ts)
	var userGroupCount map[string]*AggEventsOnUserAndGroup = make(map[string]*AggEventsOnUserAndGroup)
	// aggregate users
	startAgg := time.Now()
	err := AggregateDailyEvents(projectId, archiveCloudManager, modelCloudManager, diskManager, date.Unix(),
		userGroupCount, weights)
	if err != nil {
		log.WithError(err).Errorf("Error in aggregating users : %d time: %d ", date)
	}
	endAgg := time.Since(startAgg).Milliseconds()
	duration["agg"] += endAgg
	err = FilterAndUpdateAllEvents(userGroupCount, date.Unix(), updatedUsers, updatedGroups)
	if err != nil {
		return err
	}
	return nil

}

func AggregateDailyEvents(projectId int64, archiveCloudManager,
	modelCloudManager *filestore.FileManager,
	diskManager *serviceDisk.DiskDriver, startTime int64,
	userEventGroupCount map[string]*AggEventsOnUserAndGroup, mweights map[string][]M.AccEventWeight) error {

	configs := make(map[string]interface{})
	partFilesDir, _ := (*archiveCloudManager).GetDailyEventArchiveFilePathAndName(projectId, startTime, 0, 0)
	listFiles := (*archiveCloudManager).ListFiles(partFilesDir)
	fileNamePrefix := "events"

	domain_group, status := store.GetStore().GetGroup(projectId, M.GROUP_NAME_DOMAINS)
	if status != http.StatusFound {
		e := fmt.Errorf("failed to get existing groups (%s) for project (%d)", M.GROUP_NAME_DOMAINS, projectId)
		log.WithField("err_code", status).Error(e)
		return e
	}
	domain_group_id := fmt.Sprintf("%d", domain_group.ID)
	configs["domain_group"] = domain_group_id

	for _, partFileFullName := range listFiles {
		partFNamelist := strings.Split(partFileFullName, "/")
		partFileName := partFNamelist[len(partFNamelist)-1]
		if !strings.HasPrefix(partFileName, fileNamePrefix) {
			continue
		}

		log.WithField("project", projectId).Infof("Reading daily file :%s, %s", partFilesDir, partFileName)
		file, err := (*archiveCloudManager).Get(partFilesDir, partFileName)
		if err != nil {
			log.WithField("project", projectId).WithError(err).Errorf("Unabel to pull file :%s,%s", partFilesDir, partFileName)
			return err
		}

		err = AggEventsOnUsers(file, userEventGroupCount, mweights, configs)
		if err != nil {
			return err
		}
		err = file.Close()
		if err != nil {
			return err
		}
	}

	log.Infof("Done aggregating")

	return nil
}

// AggEventsOnUsers  agg count of each event performed by each user along with users group id.
func AggEventsOnUsers(file io.ReadCloser, userGroupCount map[string]*AggEventsOnUserAndGroup, mweights map[string][]M.AccEventWeight, configs map[string]interface{}) error {
	log.Infof("Aggregating users from file")
	scanner := bufio.NewScanner(file)
	const maxCapacity = 30 * 1024 * 1024
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	eventCount := 0
	userCounts := 0
	groupCounts := 0
	lineCount := 0

	domainGroup := configs["domain_group"].(string)
	for scanner.Scan() {
		var event *P.CounterEventFormat
		line := scanner.Text()
		err := json.Unmarshal([]byte(line), &event)
		lineCount += 1
		if err != nil {
			log.WithError(err).Errorf("Error in line :%d", eventCount)
			return err
		}

		if _, ok := userGroupCount[event.UserId]; !ok {
			var ag AggEventsOnUserAndGroup
			ag.User_id = event.UserId
			ag.EventsCount = make(map[string]*M.EventAgg)
			ag.Properties = make(map[string]PropAggregate)
			userGroupCount[event.UserId] = &ag
		}

		if val, ok := userGroupCount[event.UserId]; ok {
			// get ruleIds from filter events
			ruleIds := FilterEvents(event, mweights)
			err := updateEventChannel(val, event)
			if err != nil {
				log.Error("Unable to update channel")
			}

			// get rule_ids from filter events
			// if rule_ids is already present in val.EventsCount
			// then increment the count
			// else create a new entry in val.EventsCount
			for _, ruleId := range ruleIds {
				if _, ok := val.EventsCount[ruleId]; !ok {
					var ev M.EventAgg
					ev.EventName = event.EventName
					ev.EventId = ruleId
					ev.EventCount = 0 // if eventsCount does not contain the rule id then create an entry with count 0
					userGroupCount[event.UserId].Is_group = false
					val.EventsCount[ruleId] = &ev

				}

				if _, ok := val.EventsCount[ruleId]; ok {
					t := val.EventsCount[ruleId]
					t.EventCount += 1
				}
				userCounts += 1
			}

			id := getDomainGroupUserId(event, domainGroup)
			if id != "" {
				aggGroupCounts(event, id, userGroupCount, ruleIds)
			}

		}

	}
	err := scanner.Err()
	if err != nil {
		log.WithError(err).Errorf("error in scanning events :%v", err)
		return err
	}

	log.Infof(fmt.Sprintf("Total lines counted : %d , total users counted : %d, total groups counted : %d", lineCount, userCounts, groupCounts))

	return nil

}

func getDomainGroupUserId(event *P.CounterEventFormat, gId string) string {

	switch gId {
	case "1":
		return event.Group1UserId
	case "2":
		return event.Group2UserId
	case "3":
		return event.Group3UserId
	case "4":
		return event.Group4UserId
	case "5":
		return event.Group5UserId
	case "6":
		return event.Group6UserId
	case "7":
		return event.Group7UserId
	case "8":
		return event.Group8UserId
	}
	return ""
}

func aggGroupCounts(event *P.CounterEventFormat, user_id string, userGroupCount map[string]*AggEventsOnUserAndGroup, rule_ids []string) {

	// create new user
	if _, ok := userGroupCount[user_id]; !ok {
		var ag AggEventsOnUserAndGroup
		ag.EventsCount = make(map[string]*M.EventAgg)
		ag.User_id = user_id
		ag.Is_group = true
		userGroupCount[user_id] = &ag
	}

	// if event not present create new rule in events count
	for _, rule_id := range rule_ids {
		if val, ok := userGroupCount[user_id]; ok {
			if _, ok := val.EventsCount[rule_id]; !ok {
				var ev M.EventAgg
				ev.EventName = event.EventName
				ev.EventCount = 0
				ev.EventId = rule_id
				val.EventsCount[rule_id] = &ev

			}
			if _, ok := val.EventsCount[rule_id]; ok {
				t := val.EventsCount[rule_id]
				t.EventCount += 1
			}
		}
	}

}

func filterUpdateEvents(ts int64, data []*AggEventsOnUserAndGroup, prevCountsOfUser map[string]map[string]M.LatestScore) error {

	currentDate := U.GetDateOnlyFromTimestamp(ts)

	for _, usr := range data {

		var currCounts M.LatestScore
		currCounts.Date = ts
		currCounts.EventsCount = make(map[string]float64)
		currCounts.Properties = make(map[string]map[string]int64)

		for _, usr := range usr.EventsCount {
			currCounts.EventsCount[usr.EventId] = float64(usr.EventCount)
		}
		for propKey, propVal := range usr.Properties {
			if _, pok := currCounts.Properties[propKey]; !pok {
				currCounts.Properties[propKey] = make(map[string]int64)
			}
			for pk, pv := range propVal.Properties {
				currCounts.Properties[propKey][pk] = pv
			}
		}

		if _, uok := prevCountsOfUser[usr.User_id]; !uok {
			prevCountsOfUser[usr.User_id] = make(map[string]M.LatestScore)
		}

		prevCountsOfUser[usr.User_id][currentDate] = currCounts
	}

	return nil
}

func FilterAndUpdateAllEvents(userCounts map[string]*AggEventsOnUserAndGroup, startTime int64,
	updatedUser, updatedGroups map[string]map[string]M.LatestScore) error {

	evdata := make([]*AggEventsOnUserAndGroup, 0)
	gpdata := make([]*AggEventsOnUserAndGroup, 0)

	for _, uval := range userCounts {
		if len(uval.EventsCount) > 0 {
			if uval.Is_group {
				gpdata = append(gpdata, uval)
			} else {
				evdata = append(evdata, uval)
			}
		}
	}

	err := filterUpdateEvents(startTime, evdata, updatedUser)
	if err != nil {
		return err
	}

	err = filterUpdateEvents(startTime, gpdata, updatedGroups)
	if err != nil {
		return err
	}

	return nil
}

func WriteUserCountsAndRangesToDB(projectId int64, startTime int64, prevcountsofuser,
	updatedUser, updatedGroups map[string]map[string]M.LatestScore, groupUserMap map[string]int,
	weights M.AccWeights, salewindow int64) (map[string]map[string]M.LatestScore, error) {

	updatedGroupsCopy := make(map[string]map[string]M.LatestScore)
	for id, count := range updatedGroups {
		updatedGroupsCopy[id] = make(map[string]M.LatestScore)
		updatedGroupsCopy[id] = count
	}

	// extract prev groups from prev counts and update the the groupscopy
	// this is done to compute the last event from groups
	err := extractGroupsFromPrevCounts(projectId, prevcountsofuser, updatedGroupsCopy, groupUserMap)
	if err != nil {
		e := fmt.Errorf("unable to update prev events")
		return nil, e
	}

	// get all groups last event
	groupsLastEvent, err := UpdateLastEventsDay(updatedGroupsCopy, startTime, salewindow)
	if err != nil {
		e := fmt.Errorf("unable to update last event for groups")
		return nil, e
	}

	// update users who have activity only
	err = store.GetStore().UpdateUserEventsCountGO(projectId, updatedUser)
	if err != nil {
		return nil, err
	}

	// groupsLastevent contains last event data on all users with and without activity
	// updatedGroups contains groups of users with activity
	err = store.GetStore().UpdateGroupEventsCountGO(projectId, updatedGroups, groupsLastEvent, weights)
	if err != nil {
		return nil, err
	}

	return updatedGroupsCopy, nil
}

func getChannelInfo(event *P.CounterEventFormat) string {

	if propvalKey, ok := event.EventProperties[U.EP_CHANNEL]; ok {
		propval := U.GetPropertyValueAsString(propvalKey)
		return propval
	}
	return ""
}

func updateEventChannel(user *AggEventsOnUserAndGroup, event *P.CounterEventFormat) error {

	channel := getChannelInfo(event)
	if channel != "" {
		if _, ok := user.Properties[U.EP_CHANNEL]; !ok {
			var prp PropAggregate
			prp.Name = U.EP_CHANNEL
			prp.Properties = make(map[string]int64)
			user.Properties[U.EP_CHANNEL] = prp
		}

		if _, ok := user.Properties[U.EP_CHANNEL].Properties[channel]; !ok {
			user.Properties[U.EP_CHANNEL].Properties[channel] = 0
		}
		user.Properties[U.EP_CHANNEL].Properties[channel] += 1
	}
	return nil
}

func UpdateLastEventsDay(prevCountsOfUser map[string]map[string]M.LatestScore,
	currentTS int64, saleWindow int64) (map[string]M.LatestScore, error) {

	updatedLastScore := make(map[string]M.LatestScore, 0)
	currentDate := U.GetDateOnlyFromTimestamp(currentTS)
	userIdsmap := make(map[string]bool, 0)
	for usrkey, _ := range prevCountsOfUser {
		userIdsmap[usrkey] = true
	}

	ordereddays := U.GenDateStringsForLastNdays(currentTS, saleWindow)
	for currentUser := range userIdsmap {
		var lastevent M.LatestScore
		lastevent.Properties = make(map[string]map[string]int64)
		lastevent.EventsCount = make(map[string]float64)

		properties := make(map[string]map[string]int64)
		eventsCountWithDecay := make(map[string]float64)

		if prevCountsOnday, ok := prevCountsOfUser[currentUser]; ok {
			for _, dateOfCount := range ordereddays {
				counts, ook := prevCountsOnday[dateOfCount]
				if !ook {
					counts.EventsCount = make(map[string]float64, 0)
				}
				decayval := model.ComputeDecayValue(dateOfCount, int64(saleWindow))
				if decayval > 0 {
					for eventKey, eventVal := range counts.EventsCount {
						if _, eok := eventsCountWithDecay[eventKey]; !eok {
							eventsCountWithDecay[eventKey] = 0
						}
						eventsCountWithDecay[eventKey] += decayval * eventVal
					}
				}
				for peruserProperties, eprops := range counts.Properties {
					if _, prKeyok := properties[peruserProperties]; !prKeyok {
						properties[peruserProperties] = make(map[string]int64)
					}
					for kpropval, vpropcount := range eprops {
						if _, propValOk := counts.Properties[peruserProperties][kpropval]; !propValOk {
							counts.Properties[peruserProperties][kpropval] = 0
						}
						properties[peruserProperties][kpropval] += vpropcount
					}

				}
			}
		}
		currentDateTS := U.GetDateFromString(currentDate)
		lastevent.Date = currentDateTS
		lastevent.Properties = properties
		lastevent.EventsCount = eventsCountWithDecay
		updatedLastScore[currentUser] = lastevent
	}
	return updatedLastScore, nil

}

func extractGroupsFromPrevCounts(projectId int64, prevcountsofuser, updatedGroups map[string]map[string]M.LatestScore,
	groupUserMap map[string]int) error {

	//update
	for userId, _ := range updatedGroups {
		if ud, u1ok := prevcountsofuser[userId]; u1ok {
			for uday, ucount := range ud {
				if _, uok := updatedGroups[userId][uday]; !uok {
					updatedGroups[userId][uday] = ucount
				}
			}
		}
	}

	for userId, userCounts := range prevcountsofuser {
		_, grok := updatedGroups[userId]

		if grok == false {
			if guserVal, guserOk := groupUserMap[userId]; !guserOk {
				log.WithField("projectId", projectId).WithField("user", userId).Errorf("unable to map user ")
			} else {
				if guserVal == 1 {
					updatedGroups[userId] = make(map[string]M.LatestScore)
					updatedGroups[userId] = userCounts
				}
			}
		}

	}

	log.Infof("Total number of group:%d", len(updatedGroups))
	return nil
}

func ComputeAndWriteRangesToDB(projectId int64, currentTime int64, lookback int64,
	data map[string]map[string]M.LatestScore, weights *M.AccWeights) error {

	buckets, _, err := ComputeScoreRanges(projectId, currentTime, lookback, data, weights)
	if err != nil {
		return err
	}
	for _, br := range buckets {
		log.WithField("bucket range : ", br).Info("Bucket range")
	}

	err = WriteRangestoDB(projectId, buckets)
	if err != nil {
		return err
	}

	return nil
}

func WriteRangestoDB(projectId int64, buckets []M.BucketRanges) error {

	err := store.GetStore().WriteScoreRanges(projectId, buckets)
	if err != nil {
		return err
	}

	return nil
}

func ComputeScoreRanges(projectId int64, currentTime int64, lookback int64,
	data map[string]map[string]M.LatestScore, weights *M.AccWeights) ([]M.BucketRanges, map[string]float64, error) {

	var buckets []M.BucketRanges = make([]M.BucketRanges, lookback)
	var scoresMap map[string]float64 = make(map[string]float64)
	currentDateString := U.GetDateOnlyFromTimestamp(currentTime)
	saleWindow := weights.SaleWindow
	dates := U.GenDateStringsForLastNdays(currentTime, lookback)

	for _, date := range dates {
		ids := make(map[string]bool)
		updatedUsers := make(map[string]map[string]M.LatestScore)
		ts := U.GetDateFromString(date)
		cDates := U.GenDateStringsForLastNdays(currentTime, lookback)
		for id, counts := range data {
			for _, d := range cDates {
				if _, ok := counts[d]; ok {
					ids[id] = true
				}
			}
		}
		for id, _ := range ids {
			var tmpdata map[string]M.LatestScore = make(map[string]M.LatestScore)
			tmpuserData, tok := data[id]
			if tok {
				for _, d := range cDates {
					if userOndate, dayOk := tmpuserData[d]; dayOk {
						tmpdata[d] = userOndate
					}
				}
				updatedUsers[id] = tmpdata
			}
		}
		log.Info("date : %s number of updated users : %d", date, len(updatedUsers))
		lastEventScores, err := UpdateLastEventsDay(updatedUsers, ts, saleWindow)
		if err != nil {
			log.WithField("prjoectID", projectId).Error("Unable to compute last event scores")
		}
		log.Info("number of last events score: %d", len(lastEventScores))
		scores, scoresMaponDate, err := computeScore(projectId, weights, lastEventScores)
		if date == currentDateString {
			scoresMap = scoresMaponDate
		}

		if err != nil {
			return nil, nil, err
		}
		bucket, err := ComputeBucketRanges(scores, date)
		if err != nil {
			return nil, nil, err
		}
		buckets = append(buckets, bucket)
	}

	return buckets, scoresMap, nil
}
func computeScore(projectId int64, weights *M.AccWeights, lastEvents map[string]M.LatestScore) ([]float64, map[string]float64, error) {

	scores := make([]float64, 0)
	scoresMap := make(map[string]float64, 0)
	for id, lastEvent := range lastEvents {
		score, err := MS.ComputeAccountScoreOnLastEvent(projectId, *weights, lastEvent.EventsCount)
		if err != nil {
			log.WithField("id", id).WithField("projectID", projectId).Error("Unable to compute last event scores")
		}
		scores = append(scores, score)
		scoresMap[id] = score
	}

	return scores, scoresMap, nil

}

func ComputeBucketRanges(scores []float64, date string) (M.BucketRanges, error) {

	var bucket M.BucketRanges
	bucket.Date = date
	scoresWithoutZeros := removeZeros(scores)
	if len(scores) == 0 {
		return M.BucketRanges{}, fmt.Errorf("not enough scores")
	}

	for idx := 1; idx < len(M.BUCKETRANGES); idx++ {
		var b M.Bucket
		b.Name = M.BUCKETNAMES[idx-1]
		high, err := PercentileNearestRank(scoresWithoutZeros, M.BUCKETRANGES[idx-1])
		if err != nil {
			log.Errorf("unable to compute bucket ranges")
		}
		low, err := PercentileNearestRank(scoresWithoutZeros, M.BUCKETRANGES[idx])
		if err != nil {
			log.Errorf("unable to compute bucket ranges")
		}
		b.High = high
		b.Low = low

		bucket.Ranges = append(bucket.Ranges, b)
		de := fmt.Sprintf(" data to compute buckets lenght :%d , brh:%f, brl:%f, h:%f,l:%f", len(scoresWithoutZeros), M.BUCKETRANGES[idx-1], M.BUCKETRANGES[idx], high, low)
		log.Info(de)
	}
	return bucket, nil
}

func GetEngagementBuckets(projectId int64, scoresPerUser map[string]model.PerUserScoreOnDay) (model.BucketRanges, error) {

	maxTimeStamp := int64(0)
	for _, userScores := range scoresPerUser {
		timestamp := U.GetDateFromString(userScores.Timestamp)
		maxTimeStamp = U.Max(maxTimeStamp, timestamp)
	}
	dateTofetch := U.GetDateOnlyFromTimestamp(maxTimeStamp)
	buckets, err := store.GetStore().GetEngagementBucketsOnProject(projectId, dateTofetch)
	if err != nil {
		return model.BucketRanges{}, err
	}
	return buckets, nil

}

func GetEngagement(percentile float64, buckets model.BucketRanges) string {
	for _, bucket := range buckets.Ranges {
		if bucket.Low <= percentile && percentile < bucket.High {
			return bucket.Name
		}
	}
	return "ice"
}

func GetEngagementLevelOnUser(projectId int64, event model.LatestScore, userId string, fscore float64) string {
	scoresPerUser := make(map[string]model.PerUserScoreOnDay)

	dateString := U.GetDateOnlyFromTimestamp(event.Date)
	var userOnDay model.PerUserScoreOnDay
	userOnDay.Id = userId
	userOnDay.Timestamp = dateString
	userOnDay.Score = float32(fscore)
	scoresPerUser[dateString] = userOnDay
	buckets, errGetBuckets := GetEngagementBuckets(projectId, scoresPerUser)
	if errGetBuckets != nil {
		log.WithError(errGetBuckets).Error("Error retrieving account score buckets")

	}
	engagementLevel := GetEngagement(float64(fscore), buckets)
	return engagementLevel
}
