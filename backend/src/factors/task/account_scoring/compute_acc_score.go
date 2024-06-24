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
	"math"

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
const DELAY_SLA int = 3

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
	numLinesPerDayMap := make(map[string]int)

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

	mweights, _ := CreateweightMap(&weights)
	for tm.Unix() <= start_time.Unix() {
		dateString := U.GetDateOnlyFromTimestamp(tm.Unix())
		logCtx.Infof("Pulling daily events for :%s", dateString)
		numLinesPerDay, err := AggDailyUsersOnDate(projectId, archiveCloudManager, modelCloudManager,
			diskManager, tm, db, mweights, updatedUsers, updatedGroups,
			salewindow, DurationMap)
		if err != nil {
			logCtx.WithError(err).Errorf("Error in aggregating users : %d time: %v ", tm)
		}
		tm = tm.AddDate(0, 0, 1)
		numLinesPerDayMap[dateString] = numLinesPerDay
	}

	// from prevusers pick groups which are avaialable
	now := time.Now()
	_, err = WriteUserCountsAndRangesToDB(projectId, now.Unix(), prevCountsOfUser, updatedUsers, updatedGroups, groupUserMap, weights, salewindow, lookback)
	if err != nil {
		logCtx.WithError(err).Errorf("error in updating user counts to DB")
	}

	logCtx.Infof("Number of updated groups:%d", len(updatedGroups))

	alldaysCehcked, err := CheckLinesReadPreviousDays(numLinesPerDayMap, lookback, start_time.Unix())

	if !alldaysCehcked {
		status["error"] = err
	}

	timeDiff := time.Since(now).Milliseconds()
	DurationMap["updateDB"] += timeDiff
	logCtx.WithField("stats:", DurationMap).Info("stats or time in millliseconds")

	return status, true
}

func GetWeightfromDB(projectId int64) (M.AccWeights, error) {

	var mweights M.AccWeights
	mweights.WeightConfig = make([]M.AccEventWeight, 0)

	// get weights from DB
	weights, errStatus := store.GetStore().GetWeightsByProject(projectId)
	if errStatus != http.StatusFound {
		errorString := fmt.Sprintf("unable to get weights : %d", projectId)
		log.WithField("err code", errStatus).Errorf(errorString)
		return M.AccWeights{}, fmt.Errorf("%s", errorString)
	}

	mweights = RemoveDeletedWeights(weights)
	return mweights, nil
}

// CreateweightMap creates a map of event name to list of rules for easy lookup
func CreateweightMap(w *M.AccWeights) (map[string][]M.AccEventWeight, int64) {
	var mwt map[string][]M.AccEventWeight = make(map[string][]M.AccEventWeight)
	// create map of event name to rules for easy lookup
	wt := w.WeightConfig
	for _, e := range wt {
		if e.Is_deleted == false {
			weightkey := e.EventName
			if _, ok := mwt[weightkey]; !ok {
				mwt[weightkey] = make([]M.AccEventWeight, 0)
			}
			mwt[weightkey] = append(mwt[weightkey], e)
		}
	}
	return mwt, w.SaleWindow
}

// AggDailyUsersOnDate main entry function to aggregate users and group perday and write to file and DB
func AggDailyUsersOnDate(projectId int64, archiveCloudManager, modelCloudManager *filestore.FileManager,
	diskManager *serviceDisk.DiskDriver, date time.Time, db *gorm.DB, weights map[string][]M.AccEventWeight,
	updatedUsers, updatedGroups map[string]map[string]M.LatestScore, salewindow int64, duration map[string]int64) (int, error) {

	_, dok := duration["agg"]
	if !dok {
		duration["agg"] = int64(0)
	}
	day_ts := time.Unix(date.Unix(), 0)
	log.Infof("Aggregating  users :%d for %v", projectId, day_ts)
	var userGroupCount map[string]*AggEventsOnUserAndGroup = make(map[string]*AggEventsOnUserAndGroup)
	// aggregate users
	startAgg := time.Now()
	numLinesPerDay, err := AggregateDailyEvents(projectId, archiveCloudManager, modelCloudManager, diskManager, date.Unix(),
		userGroupCount, weights)
	if err != nil {
		log.WithError(err).Errorf("Error in aggregating users : %d time: %d ", date)
	}
	endAgg := time.Since(startAgg).Milliseconds()
	duration["agg"] += endAgg
	err = FilterAndUpdateAllEvents(userGroupCount, date.Unix(), updatedUsers, updatedGroups)
	if err != nil {
		return 0, err
	}
	return numLinesPerDay, nil

}

func AggregateDailyEvents(projectId int64, archiveCloudManager,
	modelCloudManager *filestore.FileManager,
	diskManager *serviceDisk.DiskDriver, startTime int64,
	userEventGroupCount map[string]*AggEventsOnUserAndGroup,
	mweights map[string][]M.AccEventWeight) (int, error) {

	configs := make(map[string]interface{})
	partFilesDir, _ := (*archiveCloudManager).GetDailyEventArchiveFilePathAndName(projectId, startTime, 0, 0)
	listFiles := (*archiveCloudManager).ListFiles(partFilesDir)
	fileNamePrefix := "events"

	domain_group, status := store.GetStore().GetGroup(projectId, M.GROUP_NAME_DOMAINS)
	if status != http.StatusFound {
		e := fmt.Errorf("failed to get existing groups (%s) for project (%d)", M.GROUP_NAME_DOMAINS, projectId)
		log.WithField("err_code", status).Error(e)
		return 0, e
	}
	domain_group_id := fmt.Sprintf("%d", domain_group.ID)
	configs["domain_group"] = domain_group_id
	numLines := 0
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
			return 0, err
		}

		numLinesPerFile, err := AggEventsOnUsers(file, userEventGroupCount, mweights, configs)
		if err != nil {
			return 0, err
		}
		err = file.Close()
		if err != nil {
			return 0, err
		}
		numLines += numLinesPerFile
	}

	log.Infof("Done aggregating")

	return numLines, nil
}

// AggEventsOnUsers  agg count of each event performed by each user along with users group id.
func AggEventsOnUsers(file io.ReadCloser, userGroupCount map[string]*AggEventsOnUserAndGroup,
	mweights map[string][]M.AccEventWeight, configs map[string]interface{}) (int, error) {
	log.Infof("Aggregating users from file")
	scanner := bufio.NewScanner(file)
	const maxCapacity = 30 * 1024 * 1024
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	eventCount := 0
	userCounts := 0
	groupCounts := 0
	lineCount := 0
	numLines := 0
	domainGroup := configs["domain_group"].(string)
	for scanner.Scan() {
		var event *P.CounterEventFormat
		line := scanner.Text()
		err := json.Unmarshal([]byte(line), &event)
		lineCount += 1
		if err != nil {
			log.WithError(err).Errorf("Error in line :%d", eventCount)
			return 0, err
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
			// err := updateEventChannel(val, event)
			// if err != nil {
			// 	log.Error("Unable to update channel")
			// }

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
		numLines += 1
	}
	err := scanner.Err()
	if err != nil {
		log.WithError(err).Errorf("error in scanning events :%v", err)
		return 0, err
	}

	log.Infof(fmt.Sprintf("Total lines counted : %d , total users counted : %d, total groups counted : %d", lineCount, userCounts, groupCounts))

	return numLines, nil

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
	weights M.AccWeights, salewindow int64, lookback int) (map[string]map[string]M.LatestScore, error) {

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
	groupsLastEvent, groupAllEvents, err := UpdateLastEventsDay(updatedGroupsCopy, startTime, weights, salewindow)
	if err != nil {
		e := fmt.Errorf("unable to update last event for groups")
		return nil, e
	}

	now := time.Now()
	buckets, err := ComputeAndWriteRangesToDB(projectId, now.Unix(), int64(lookback), updatedGroupsCopy, weights)
	if err != nil {
		log.WithError(err).Errorf("error in updating ranges DB")
	}

	// not updating users due to lag in DB
	// // update users who have activity only
	// err = store.GetStore().UpdateUserEventsCountGO(projectId, updatedUser)
	// if err != nil {
	// 	return nil, err
	// }

	maxBucket := getMaxBucket(buckets)

	// groupsLastevent contains last event data on all users with and without activity
	// updatedGroups contains groups of users with activity
	err = store.GetStore().UpdateGroupEventsCountGO(projectId, updatedGroups, groupsLastEvent, groupAllEvents, weights, maxBucket)
	if err != nil {
		return nil, err
	}

	return updatedGroupsCopy, nil
}

func getMaxBucket(buckets []M.BucketRanges) M.BucketRanges {
	var maxBucket M.BucketRanges
	maxTS := int64(0)
	for _, bucket := range buckets {
		if bucket.Date != "" {
			dateTs := U.GetDateFromString(bucket.Date)
			if dateTs > maxTS {
				maxBucket = bucket
				maxTS = dateTs
			}
		} else {
			log.WithField("buckets", buckets).Error("unable to decode bucket")
		}
	}
	return maxBucket
}

func getChannelInfo(event *P.CounterEventFormat) string {

	if propvalKey, ok := event.EventProperties[U.EP_CHANNEL]; ok {
		propval := U.GetPropertyValueAsString(propvalKey)
		return propval
	} else {
		log.WithField("event without channel", event).Errorf("Unable to get channel info")
	}
	return ""
}

func UpdateEventChannel_(user *AggEventsOnUserAndGroup, event *P.CounterEventFormat) error {
	return updateEventChannel(user, event)
}

func updateEventChannel(user *AggEventsOnUserAndGroup, event *P.CounterEventFormat) error {

	channel := getChannelInfo(event)
	if _, channel_ok := user.Properties[U.EP_CHANNEL]; !channel_ok {
		return nil
	}

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
	currentTS int64, weigths M.AccWeights, saleWindow int64) (map[string]M.LatestScore, map[string]M.LatestScore, error) {

	updatedLastScore := make(map[string]M.LatestScore, 0)
	updatedAllEventsScore := make(map[string]M.LatestScore, 0)

	currentDate := U.GetDateOnlyFromTimestamp(currentTS)
	userIdsmap := make(map[string]bool, 0)
	for usrkey, _ := range prevCountsOfUser {
		userIdsmap[usrkey] = true
	}

	ordereddays := U.GenDateStringsForLastNdays(currentTS, saleWindow)
	for currentUser := range userIdsmap {
		var lastevent M.LatestScore
		var lastEventDay string
		lastevent.Properties = make(map[string]map[string]int64)
		lastevent.EventsCount = make(map[string]float64)
		lastevent.TopEvents = make(map[string]float64)

		var allEventsInrange M.LatestScore
		allEventsInrange.Properties = make(map[string]map[string]int64)
		allEventsInrange.EventsCount = make(map[string]float64)
		allEventsInrange.TopEvents = make(map[string]float64)

		properties := make(map[string]map[string]int64)
		eventsCountWithDecay := make(map[string]float64)
		eventsCountWithoutDecay := make(map[string]float64)

		if prevCountsOnday, ok := prevCountsOfUser[currentUser]; ok {
			for _, dateOfCount := range ordereddays {
				counts, ook := prevCountsOnday[dateOfCount]
				if !ook {
					counts.EventsCount = make(map[string]float64, 0)
				}
				decayval := model.ComputeDecayValue(dateOfCount, int64(saleWindow))
				if decayval > 0 {
					if len(counts.EventsCount) > 0 {
						lastEventDay = dateOfCount
					}
					for eventKey, eventVal := range counts.EventsCount {
						if _, eok := eventsCountWithDecay[eventKey]; !eok {
							eventsCountWithDecay[eventKey] = 0
						}
						eventsCountWithDecay[eventKey] += decayval * eventVal
					}
				}

				// take counts of all events happened from start till today ..ignoring decay value
				for eventKey, eventVal := range counts.EventsCount {
					if _, eok := eventsCountWithoutDecay[eventKey]; !eok {
						eventsCountWithoutDecay[eventKey] = 0

					}
					eventsCountWithoutDecay[eventKey] += eventVal
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

		//take top k events contribution to score
		// using undecayed counts
		topEventsOnScore := GetTopkEventsOnCounts(eventsCountWithoutDecay, weigths, TAKETOPK)

		currentDateTS := U.GetDateFromString(currentDate)
		lastevent.Date = currentDateTS
		lastevent.Properties = properties
		lastevent.EventsCount = eventsCountWithDecay
		lastevent.TopEvents = topEventsOnScore
		lastevent.LastEventTimeStamp = lastEventDay

		allEventsInrange.Date = currentDateTS
		allEventsInrange.Properties = properties
		allEventsInrange.EventsCount = eventsCountWithoutDecay

		updatedLastScore[currentUser] = lastevent
		updatedAllEventsScore[currentUser] = allEventsInrange
	}
	return updatedLastScore, updatedAllEventsScore, nil

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
	data map[string]map[string]M.LatestScore, weights M.AccWeights) ([]M.BucketRanges, error) {

	buckets, _, err := ComputeScoreRanges(projectId, currentTime, lookback, data, weights)
	if err != nil {
		return nil, err
	}
	for _, br := range buckets {
		log.WithField("bucket range : ", br).Info("Bucket range")
	}

	err = WriteRangestoDB(projectId, buckets)
	if err != nil {
		return nil, err
	}

	return buckets, nil
}

func WriteRangestoDB(projectId int64, buckets []M.BucketRanges) error {

	err := store.GetStore().WriteScoreRanges(projectId, buckets)
	if err != nil {
		return err
	}

	return nil
}

func ComputeScoreRanges(projectId int64, currentTime int64, lookback int64,
	data map[string]map[string]M.LatestScore, weights M.AccWeights) ([]M.BucketRanges, map[string]float64, error) {

	logCtx := log.WithFields(log.Fields{
		"projectId": projectId,
	})

	var buckets []M.BucketRanges = make([]M.BucketRanges, lookback)
	var scoresMap map[string]float64 = make(map[string]float64)
	currentDateString := U.GetDateOnlyFromTimestamp(currentTime)
	saleWindow := weights.SaleWindow
	dates := U.GenDateStringsForLastNdays(currentTime, lookback)
	for _, date := range dates {
		ids := make(map[string]bool)
		updatedUsers := make(map[string]map[string]M.LatestScore)
		ts := U.GetDateFromString(date)
		cDates := U.GenDateStringsForLastNdays(currentTime, weights.SaleWindow)
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
		logCtx.Infof("date : %s number of updated users : %d", date, len(updatedUsers))
		lastEventScores, _, err := UpdateLastEventsDay(updatedUsers, ts, weights, saleWindow)
		if err != nil {
			log.WithField("prjoectID", projectId).Error("Unable to compute last event scores")
		}
		logCtx.Info("number of last events score: %d", len(lastEventScores))
		scores, scoresMaponDate, err := computeScore(projectId, weights, lastEventScores)
		if date == currentDateString {
			scoresMap = scoresMaponDate
		}

		if err != nil {
			return nil, nil, err
		}
		bucket, err := ComputeBucketRanges(projectId, scores, date)
		if err != nil {
			return nil, nil, err
		}
		buckets = append(buckets, bucket)
	}

	logCtx.WithField("buckets", buckets).Debug("computed buckets for project")

	return buckets, scoresMap, nil
}
func computeScore(projectId int64, weights M.AccWeights, lastEvents map[string]M.LatestScore) ([]float64, map[string]float64, error) {

	scores := make([]float64, 0)
	scoresMap := make(map[string]float64, 0)
	for id, lastEvent := range lastEvents {
		score, err := MS.ComputeAccountScoreOnLastEvent(projectId, weights, lastEvent.EventsCount)
		if err != nil {
			log.WithField("id", id).WithField("projectID", projectId).Error("Unable to compute last event scores")
		}
		scores = append(scores, score)
		scoresMap[id] = score
	}

	return scores, scoresMap, nil

}

func ComputeBucketRanges(projectId int64, scores []float64, date string) (M.BucketRanges, error) {

	logCtx := log.WithFields(log.Fields{
		"projectId": projectId,
	})

	var bucket M.BucketRanges
	var projectbucketRanges M.BucketRanges
	var statusCode int
	projectbucketRanges.Ranges = make([]M.Bucket, 0)

	bucket.Date = date
	scoresWithoutZeros := removeZeros(scores)
	if len(scores) == 0 {
		return M.BucketRanges{}, fmt.Errorf("not enough scores")
	}

	// get the max and min percentile ranges from custom config
	projectbucketRanges.Ranges = make([]M.Bucket, 4)
	buckets, statusCode := store.GetStore().GetEngagementLevelsByProject(projectId)
	if statusCode != http.StatusFound {
		logCtx.Info("Getting default engagement buckets")
		projectbucketRanges = model.DefaultEngagementBuckets()
		projectbucketRanges.Date = date
	} else {
		projectbucketRanges = buckets
	}

	for idx := 0; idx < len(projectbucketRanges.Ranges); idx++ {
		var b M.Bucket
		b.Name = projectbucketRanges.Ranges[idx].Name
		high, err := PercentileNearestRank(scoresWithoutZeros, projectbucketRanges.Ranges[idx].High)
		if err != nil {
			log.Errorf("unable to compute bucket ranges")
		}
		low, err := PercentileNearestRank(scoresWithoutZeros, projectbucketRanges.Ranges[idx].Low)
		if err != nil {
			log.Errorf("unable to compute bucket ranges")
		}
		b.High = math.Round(high)
		b.Low = math.Round(low)
		if b.Name == "Hot" {
			b.High += 1
		}

		bucket.Ranges = append(bucket.Ranges, b)
	}
	logCtx.WithField("bucket", bucket).Info("buckets")
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
	engagementLevel := model.GetEngagement(float64(fscore), buckets)
	return engagementLevel
}

func GetTopkEventsOnCounts(events map[string]float64, weigths M.AccWeights, topK int) map[string]float64 {

	var eventsRetained map[string]bool = make(map[string]bool)
	for _, w := range weigths.WeightConfig {
		if !w.Is_deleted {
			eventsRetained[w.WeightId] = true
		}
	}
	for evKey, _ := range events {
		if _, ok := eventsRetained[evKey]; !ok {
			delete(events, evKey)
		}
	}

	// if num events less than topk return events
	if len(events) <= topK {
		return events
	}
	result := U.TakeTopKOnSortedmap(events, topK)
	return result
}

func RemoveDeletedWeights(weights *M.AccWeights) M.AccWeights {
	log.Info("indide remove deleted weights")
	var mweights M.AccWeights

	mweights.SaleWindow = weights.SaleWindow
	mweights.WeightConfig = make([]M.AccEventWeight, 0)

	for _, w := range weights.WeightConfig {
		if !w.Is_deleted {
			mweights.WeightConfig = append(mweights.WeightConfig, w)
		} else {
			log.Infof("deleted name : %s", w.EventName)

		}
	}

	return mweights
}

func CheckLinesReadPreviousDays(numLinesPerDay map[string]int, lookback int,
	startTime int64) (bool, error) {

	slaMap := make(map[string]bool)
	lookbackMap := make(map[string]bool)

	dateStringSLA := U.GenDateStringsForLastNdays(startTime, int64(DELAY_SLA))
	dateStringLookBack := U.GenDateStringsForLastNdays(startTime, int64(lookback))

	for _, dt := range dateStringSLA {
		slaMap[dt] = true
	}

	for _, dt := range dateStringLookBack {
		lookbackMap[dt] = true
		if _, ok := slaMap[dt]; ok {
			lookbackMap[dt] = false
		}
	}

	numDaysNoData := make([]string, 0)
	for dtString, dtBool := range lookbackMap {
		if dtBool {
			dtInt, dtOk := numLinesPerDay[dtString]
			if dtOk {
				if dtInt == 0 {
					numDaysNoData = append(numDaysNoData, dtString)
				}
			} else {
				numDaysNoData = append(numDaysNoData, dtString)
			}
		}
	}

	if len(numDaysNoData) > 0 {
		n := strings.Join(numDaysNoData, " ")
		err := fmt.Errorf("no data found on :%s", n)
		return false, err
	}

	return true, nil

}
