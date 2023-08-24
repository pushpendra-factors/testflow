package task

import (
	"bufio"
	"encoding/json"
	C "factors/config"
	"factors/filestore"
	"factors/model/model"
	M "factors/model/model"
	"factors/model/store"
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
	status := make(map[string]interface{})
	var err error
	modelCloudManager := configs["modelCloudManager"].(*filestore.FileManager)
	archiveCloudManager := configs["archiveCloudManager"].(*filestore.FileManager)
	diskManager := configs["diskManager"].(*serviceDisk.DiskDriver)
	day_time := configs["day_timestamp"].(int64)
	lookback := configs["lookback"].(int)
	start_time := time.Unix(day_time, 0)
	end_time := start_time.AddDate(0, 0, -1*lookback)

	tm := end_time
	db := C.GetServices().Db

	weights, errStatus := store.GetStore().GetWeightsByProject(projectId)
	if errStatus != http.StatusFound {
		status["err"] = "unable to get weights"
		return status, false
	}
	salewindow := weights.SaleWindow

	mweights, _ := CreateweightMap(weights)
	for k, v := range mweights {
		log.Infof("weight : %s, val :%v", k, v)
	}

	// prevCountsOfUser, err := store.GetStore().GetAllUserEvents(projectId, false)
	// if err != nil {
	// 	status["err"] = "unable to prev counts"
	// 	return status, false
	// }

	for tm.Unix() <= start_time.Unix() {
		dateString := GetDateOnlyFromTimestamp(tm.Unix())
		log.Infof("Pulling daily events for :%s", dateString)

		prevCountsOfUser, err := store.GetStore().GetAllUserEvents(projectId, false)
		if err != nil {
			status["err"] = "unable to prev counts"
			return status, false
		}
		err = AggDailyUsersOnDate(projectId, archiveCloudManager, modelCloudManager, diskManager, tm, db, mweights, prevCountsOfUser, salewindow)
		if err != nil {
			log.WithError(err).Errorf("Error in aggregating users : %d time: %v ", tm)
		}
		tm = tm.AddDate(0, 0, 1)
	}

	err = db.Close()
	if err != nil {
		log.Errorf("unable to close db")
		status["error"] = "unable to close db"
		return status, false
	}

	return status, true

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
	prevcountsofuser map[string]map[string]M.LatestScore, salewindow int64) error {

	day_ts := time.Unix(date.Unix(), 0)
	log.Infof("Aggregating  users :%d for %v", projectId, day_ts)
	var userGroupCount map[string]*AggEventsOnUserAndGroup = make(map[string]*AggEventsOnUserAndGroup)
	// aggregate users
	err := AggregateDailyEvents(projectId, archiveCloudManager, modelCloudManager, diskManager, date.Unix(),
		userGroupCount, weights)
	if err != nil {
		log.WithError(err).Errorf("Error in aggregating users : %d time: %d ", date)
	}

	err = WriteUserCountsToDB(projectId, date.Unix(), userGroupCount, weights, prevcountsofuser, salewindow)
	if err != nil {
		return err
	}
	return nil

}

func GetDateOnlyFromTimestamp(ts int64) string {
	year, month, date := time.Unix(ts, 0).UTC().Date()
	data := fmt.Sprintf("%d%02d%02d", year, month, date)
	return data

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

		log.Infof("Reading daily file :%s, %s", partFilesDir, partFileName)
		file, err := (*archiveCloudManager).Get(partFilesDir, partFileName)
		if err != nil {
			log.WithError(err).Errorf("Unabel to pull file :%s,%s", partFilesDir, partFileName)
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

	// if domainVal, ok := configs["domain_group"].(string); ok {
	// 	domainGroup = domainVal
	// } else {
	// 	domainGroup = ""
	// }

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

			if !event.IsGroupUser {
				if event.Group1UserId != "" {
					groupCounts += 1
					aggGroupCounts(event, event.Group1UserId, userGroupCount, ruleIds)
				}
				if event.Group2UserId != "" {
					groupCounts += 1
					aggGroupCounts(event, event.Group2UserId, userGroupCount, ruleIds)
				}
				if event.Group3UserId != "" {
					groupCounts += 1
					aggGroupCounts(event, event.Group3UserId, userGroupCount, ruleIds)
				}
				if event.Group4UserId != "" {
					groupCounts += 1
					aggGroupCounts(event, event.Group4UserId, userGroupCount, ruleIds)
				}

				if event.Group5UserId != "" {
					groupCounts += 1
					aggGroupCounts(event, event.Group5UserId, userGroupCount, ruleIds)
				}
				if event.Group6UserId != "" {
					groupCounts += 1
					aggGroupCounts(event, event.Group6UserId, userGroupCount, ruleIds)
				}
				if event.Group7UserId != "" {
					groupCounts += 1
					aggGroupCounts(event, event.Group7UserId, userGroupCount, ruleIds)
				}
				if event.Group8UserId != "" {
					groupCounts += 1
					aggGroupCounts(event, event.Group8UserId, userGroupCount, ruleIds)
				}
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

func WriteUserCountsToDB(projectId int64, startTime int64,
	userCounts map[string]*AggEventsOnUserAndGroup, weights map[string][]M.AccEventWeight,
	prevcountsofuser map[string]map[string]M.LatestScore, salewindow int64) error {

	evdata := make([]*AggEventsOnUserAndGroup, 0)
	gpdata := make([]*AggEventsOnUserAndGroup, 0)

	mweights := make(map[string]bool, 0)

	for _, v := range weights {
		for _, r := range v {
			mweights[r.WeightId] = true
		}
	}

	for _, uval := range userCounts {
		if len(uval.EventsCount) > 0 {
			if uval.Is_group {
				gpdata = append(gpdata, uval)
			} else {
				evdata = append(evdata, uval)
			}
		}
	}

	evDataLastEvent, err := UpdateLastEventsDay(prevcountsofuser, evdata, startTime, salewindow)
	if err != nil {
		e := fmt.Errorf("unable to update last event for users")
		return e
	}
	gpDataLastEvent, err := UpdateLastEventsDay(prevcountsofuser, gpdata, startTime, salewindow)
	if err != nil {
		e := fmt.Errorf("unable to update last event for groups")
		return e
	}

	evData := transformUserActivityies(projectId, evdata, startTime)
	gpData := transformUserActivityies(projectId, gpdata, startTime)

	if len(evdata) > 0 {
		err := store.GetStore().UpdateUserEventsCount(evData, evDataLastEvent)
		if err != nil {
			return err
		}
	}

	if len(gpdata) > 0 {

		err := store.GetStore().UpdateGroupEventsCount(gpData, gpDataLastEvent)
		if err != nil {
			return err
		}
	}
	return nil
}

func transformUserActivityies(projectId int64, ev []*AggEventsOnUserAndGroup, ts int64) []M.EventsCountScore {

	var events []M.EventsCountScore
	events = make([]M.EventsCountScore, 0)
	dateString := GetDateOnlyFromTimestamp(ts)
	for _, usr := range ev {
		var usrCount M.EventsCountScore
		usrCount.UserId = usr.User_id
		usrCount.EventScore = make(map[string]int64)

		for evKey, evVal := range usr.EventsCount {
			usrCount.EventScore[evKey] = evVal.EventCount
		}
		usrCount.Property = make(map[string]map[string]int64)
		for _, ev := range usr.Properties {
			propsName := ev.Name
			props := ev.Properties
			if _, ok := usrCount.Property[propsName]; !ok {
				usrCount.Property[propsName] = make(map[string]int64)
			}

			// for each property val add the count
			for k, v := range props {
				if _, ok := usrCount.Property[propsName][k]; !ok {
					usrCount.Property[propsName][k] = 0
				}
				usrCount.Property[propsName][k] = v
			}
		}
		usrCount.IsGroup = usr.Is_group
		usrCount.DateStamp = dateString
		usrCount.ProjectId = projectId
		events = append(events, usrCount)
	}

	return events
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

func UpdateCountsWithDecayToUpdateDB(CurrUserCounts map[string]*AggEventsOnUserAndGroup, weights map[string][]M.AccEventWeight,
	prevcountsofuser map[string]map[string]M.LatestScore, salewindow int64, currentts int64) []M.DbUpdateAccScoring {

	mweights := make(map[string]bool, 0)
	for _, v := range weights {
		for _, r := range v {
			mweights[r.WeightId] = true
		}
	}

	if len(prevcountsofuser) == 0 {
		prevcountsofuser = make(map[string]map[string]M.LatestScore)
	}

	currentDate := GetDateOnlyFromTimestamp(currentts)
	CurrDateTS := M.GetDateFromString(currentDate)
	userWithUpdatedCounts := make([]M.DbUpdateAccScoring, 0)

	userIdsmap := make(map[string]bool, 0)
	for usrkey, _ := range prevcountsofuser {
		userIdsmap[usrkey] = true
	}
	for usrkey, _ := range CurrUserCounts {
		userIdsmap[usrkey] = true
	}

	for userId, _ := range userIdsmap {

		var updateCounts M.DbUpdateAccScoring
		var LastEvent M.LatestScore
		var CurrEvent M.LatestScore
		updateCounts.Userid = userId
		LastEvent.EventsCount = make(map[string]float64)
		LastEvent.Properties = make(map[string]map[string]int64)
		eventsCountWithDecay := make(map[string]float64)
		updateCounts.IsGroup = false

		if _, ok := prevcountsofuser[userId]; !ok {
			prevcountsofuser[userId] = make(map[string]M.LatestScore)
		}

		if _, ok := CurrUserCounts[userId]; ok {
			currUser := CurrUserCounts[userId]
			updateCounts.IsGroup = currUser.Is_group
			CurrEvent.Date = CurrDateTS
			CurrEvent.EventsCount = make(map[string]float64)
			CurrEvent.Properties = make(map[string]map[string]int64)

			// create the current event with date and actual count
			for _, ev := range currUser.EventsCount {
				eid := ev.EventId
				if _, ok := mweights[eid]; ok {
					if _, ok := CurrEvent.EventsCount[eid]; !ok {
						CurrEvent.EventsCount[eid] = 0
					}
					CurrEvent.EventsCount[eid] = float64(ev.EventCount)

				}
			}

			for _, ev := range currUser.Properties {
				propsName := ev.Name
				props := ev.Properties
				if _, ok := CurrEvent.Properties[propsName]; !ok {
					CurrEvent.Properties[propsName] = make(map[string]int64)
				}

				// for each property val add the count
				for k, v := range props {
					if _, ok := CurrEvent.Properties[propsName][k]; !ok {
						CurrEvent.Properties[propsName][k] = 0
					}
					CurrEvent.Properties[propsName][k] = v
				}
			}
		}

		// update prevuser with event counts of  curr date
		prevcountsofuser[userId][currentDate] = CurrEvent
		properties := make(map[string]map[string]int64)
		prevCountsOnday := prevcountsofuser[userId]
		for dateOfCount, counts := range prevCountsOnday {
			dateTs := M.GetDateFromString(dateOfCount)

			if dateTs <= currentts {
				decayval := M.ComputeDecayValue(dateOfCount, salewindow)
				if decayval > 0 {
					for eventKey, eventVal := range counts.EventsCount {
						if _, eok := eventsCountWithDecay[eventKey]; !eok {
							eventsCountWithDecay[eventKey] = 0
						}
						eventsCountWithDecay[eventKey] += eventVal

					}
				}

				// for each property take the property val and update the counts
				for propKey, propCounts := range counts.Properties {
					if _, pok := properties[propKey]; !pok {
						properties[propKey] = make(map[string]int64)
					}
					for k, v := range propCounts {
						if _, kok := properties[propKey][k]; !kok {
							properties[propKey][k] = 0
						}
						properties[propKey][k] = properties[propKey][k] + v
					}

				}

				for evK, evVal := range eventsCountWithDecay {
					if evVal == 0 {
						delete(eventsCountWithDecay, evK)
					}
				}

			}
		}

		//take topK properties
		for _, propSortVal := range properties {
			pq := make(map[string]int, 0)
			propSortedKeys := make([]string, 0)
			for p, q := range propSortVal {
				pq[p] = int(q)
			}
			propSortedKeysSlice := U.SortOnPriorityTable(pq, true)

			for idx, propKeyRanked := range propSortedKeysSlice {
				if idx < TAKETOPK {
					propSortedKeys = append(propSortedKeys, propKeyRanked)
				}
			}
			keystoretain := make(map[string]bool)
			for _, propKeytoRetain := range propSortedKeys {
				keystoretain[propKeytoRetain] = true
			}

			for propKeyInCounts, _ := range propSortVal {
				if _, propRetainOk := keystoretain[propKeyInCounts]; !propRetainOk {
					delete(propSortVal, propKeyInCounts)
				}
			}
		}
		LastEvent.Date = currentts
		LastEvent.EventsCount = eventsCountWithDecay
		LastEvent.Properties = properties

		updateCounts.Date = currentDate
		updateCounts.CurrEventCount = CurrEvent
		updateCounts.Lastevent = LastEvent
		userWithUpdatedCounts = append(userWithUpdatedCounts, updateCounts)
	}

	return userWithUpdatedCounts

}

func UpdateLastEventsDay(prevCountsOfUser map[string]map[string]M.LatestScore, data []*AggEventsOnUserAndGroup,
	currentTS int64, saleWindow int64) (map[string]M.LatestScore, error) {

	updatedLastScore := make(map[string]M.LatestScore, 0)
	currentDate := GetDateOnlyFromTimestamp(currentTS)
	currentUserMap := make(map[string]*AggEventsOnUserAndGroup)

	userIdsmap := make(map[string]bool, 0)
	for usrkey, _ := range prevCountsOfUser {
		userIdsmap[usrkey] = true
	}
	for _, usrval := range data {
		userIdsmap[usrval.User_id] = true
		currentUserMap[usrval.User_id] = usrval
	}

	ordereddays := GenDateStringsForLastNdays(currentTS, saleWindow)
	for currentUser := range userIdsmap {
		var lastevent M.LatestScore
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

		decayvalcurrent := model.ComputeDecayValue(currentDate, int64(saleWindow))
		if eventCountPerUser, ok := currentUserMap[currentUser]; ok {

			counts := eventCountPerUser.EventsCount
			for _, eventCount := range counts {
				eventKey := eventCount.EventId
				if _, eok := eventsCountWithDecay[eventKey]; !eok {
					eventsCountWithDecay[eventKey] = 0
				}
				eventsCountWithDecay[eventKey] += decayvalcurrent * float64(eventCount.EventCount)
			}

			for _, eprops := range eventCountPerUser.Properties {
				propName := eprops.Name
				if _, prKeyok := properties[propName]; !prKeyok {
					properties[propName] = make(map[string]int64)
				}

				for k, v := range eprops.Properties {
					if _, propValOk := properties[propName][k]; !propValOk {
						properties[propName][k] = 0
					}
					properties[propName][k] += v
				}

			}

		}

		currentDateTS := M.GetDateFromString(currentDate)
		lastevent.Date = currentDateTS
		lastevent.Properties = properties
		lastevent.EventsCount = make(map[string]float64)
		lastevent.EventsCount = eventsCountWithDecay
		updatedLastScore[currentUser] = lastevent
	}
	return updatedLastScore, nil

}

func GenDateStringsForLastNdays(currDate int64, salewindow int64) []string {

	dateStrings := make([]string, 0)

	currts := time.Unix(currDate, 0)
	for idx := salewindow; idx > 0; idx-- {
		ts := currts.AddDate(0, 0, -1*int(idx))
		dstring := GetDateOnlyFromTimestamp(ts.Unix())
		dateStrings = append(dateStrings, dstring)
	}
	return dateStrings
}
