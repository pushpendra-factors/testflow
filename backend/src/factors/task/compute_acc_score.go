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
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

type AggEventsPerUser struct {
	User_id      string       `json:"uid"`
	EventsCount  []M.EventAgg `json:"eg"`
	Gid1         string       `json:"uid1"`
	Gid2         string       `json:"uid2"`
	Gid3         string       `json:"uid3"`
	Gid4         string       `json:"uid4"`
	DayTimestamp int64        `json:"ts"`
}

type AggEventsOnUserAndGroup struct {
	User_id     string                 `json:"uid"`
	EventsCount map[string]*M.EventAgg `json:"eg"`
	Is_group    bool                   `json:"ig"`
}

type AggEventsOnGroupsPerProj struct {
	Account_id       int64                                        `json:"aid"`
	GroupEventsCount map[string]map[string]map[string]*M.EventAgg `json:"grp"`
	TimeStamp        int64                                        `json:"ets"`
}

type AggGroupPerProj struct {
	Project_id       int64                  `json:"pid"`
	Account_id       string                 `json:"accid"`
	Group_id         string                 `json:"gid"`
	GroupEventsCount map[string]*M.EventAgg `json:"grp"`
	TimeStamp        int64                  `json:"ets"`
}

func BuildAccScoringDaily(projectId int64, configs map[string]interface{}) (map[string]interface{}, bool) {
	status := make(map[string]interface{})

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

	mweights, _ := CreateweightMap(weights)

	for k, v := range mweights {
		log.Infof("weight : %s, val :%v", k, v)
	}

	for tm.Unix() <= start_time.Unix() {
		dateString := GetDateOnlyFromTimestamp(tm.Unix())
		log.Infof("Pulling daily events for :%s", dateString)
		err := AggDailyUsersOnDate(projectId, archiveCloudManager, modelCloudManager, diskManager, tm, db, mweights)
		if err != nil {
			log.WithError(err).Errorf("Error in aggregating users : %d time: %v ", tm)
		}
		tm = tm.AddDate(0, 0, 1)
	}

	err := db.Close()
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
	diskManager *serviceDisk.DiskDriver, date time.Time, db *gorm.DB, weights map[string][]M.AccEventWeight) error {

	day_ts := time.Unix(date.Unix(), 0)
	log.Infof("Aggregating  users :%d for %v", projectId, day_ts)
	var userGroupCount map[string]*AggEventsOnUserAndGroup = make(map[string]*AggEventsOnUserAndGroup)
	// aggregate users
	err := AggregateDailyEvents(projectId, archiveCloudManager, modelCloudManager, diskManager, date.Unix(), userGroupCount, weights)
	if err != nil {
		log.WithError(err).Errorf("Error in aggregating users : %d time: %v ", date)
	}

	err = WriteUserCountsToDB(projectId, date.Unix(), userGroupCount, weights)
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

	domain_group, status := store.GetStore().GetGroup(projectId, model.GROUP_NAME_DOMAINS)
	if status != http.StatusNotFound {
		e := fmt.Errorf("failed to get existing groups (%s) for project (%d)", model.GROUP_NAME_DOMAINS, projectId)
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
	var domainGroup string
	scanner := bufio.NewScanner(file)
	const maxCapacity = 30 * 1024 * 1024
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)
	eventCount := 0
	userCounts := 0
	groupCounts := 0
	lineCount := 0

	if domainVal, ok := configs["domain_group"].(string); ok {
		domainGroup = domainVal
	} else {
		domainGroup = ""
	}

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
			userGroupCount[event.UserId] = &ag
		}

		if val, ok := userGroupCount[event.UserId]; ok {
			// get ruleIds from filter events
			ruleIds := FilterEvents(event, mweights)

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
				if event.Group1UserId != "" && domainGroup != "1" {
					groupCounts += 1
					aggGroupCounts(event, event.Group1UserId, userGroupCount, ruleIds)
				}
				if event.Group2UserId != "" && domainGroup != "2" {
					groupCounts += 1
					aggGroupCounts(event, event.Group2UserId, userGroupCount, ruleIds)
				}
				if event.Group3UserId != "" && domainGroup != "3" {
					groupCounts += 1
					aggGroupCounts(event, event.Group3UserId, userGroupCount, ruleIds)
				}
				if event.Group4UserId != "" && domainGroup != "4" {
					groupCounts += 1
					aggGroupCounts(event, event.Group4UserId, userGroupCount, ruleIds)
				}

				if event.Group5UserId != "" && domainGroup != "5" {
					groupCounts += 1
					aggGroupCounts(event, event.Group5UserId, userGroupCount, ruleIds)
				}
				if event.Group6UserId != "" && domainGroup != "6" {
					groupCounts += 1
					aggGroupCounts(event, event.Group6UserId, userGroupCount, ruleIds)
				}
				if event.Group7UserId != "" && domainGroup != "7" {
					groupCounts += 1
					aggGroupCounts(event, event.Group7UserId, userGroupCount, ruleIds)
				}
				if event.Group8UserId != "" && domainGroup != "8" {
					groupCounts += 1
					aggGroupCounts(event, event.Group8UserId, userGroupCount, ruleIds)
				}
			} else if event.IsGroupUser {

				// get domain group user id
				if domainGroup != "" {
					id := getDomainGroupUserId(event, domainGroup)
					aggGroupCounts(event, id, userGroupCount, ruleIds)
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
	userCounts map[string]*AggEventsOnUserAndGroup, weights map[string][]M.AccEventWeight) error {

	evdata := make([]M.EventsCountScore, 0)
	gpdata := make([]M.EventsCountScore, 0)

	mweights := make(map[string]bool, 0)

	for _, v := range weights {
		for _, r := range v {
			mweights[r.WeightId] = true
		}
	}

	for _, uVal := range userCounts {
		us := uVal
		var usr M.EventsCountScore
		evs := us.EventsCount

		usr.UserId = us.User_id
		usr.EventScore = make(map[string]int64)
		usr.DateStamp = GetDateOnlyFromTimestamp(startTime)
		usr.ProjectId = projectId
		for _, eg := range evs {
			ename := eg.EventId
			if _, ok := mweights[ename]; ok {
				usr.EventScore[ename] = eg.EventCount
			}
		}

		if len(uVal.EventsCount) > 0 {
			if uVal.Is_group {
				gpdata = append(gpdata, usr)
			} else {
				evdata = append(evdata, usr)
			}
		}
	}

	if len(evdata) > 0 {
		err := store.GetStore().UpdateUserEventsCount(evdata)
		if err != nil {
			return err
		}
	}

	if len(gpdata) > 0 {
		err := store.GetStore().UpdateGroupEventsCount(gpdata)
		if err != nil {
			return err
		}
	}

	log.Infof("Total number of users and groups: %d , %d", len(evdata), len(gpdata))

	return nil
}
