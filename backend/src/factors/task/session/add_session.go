package session

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	C "factors/config"
	M "factors/model"
	U "factors/util"
)

type NextSessionInfo struct {
	UserId         string `json:"user_id"`
	StartTimestamp int64  `json:"start_timestamp"`
}

const (
	StatusNotModified = "not_modified"
	StatusFailed      = "failed"
)

/*
Get users,
without session - where session_id null, select user_id, min(timestamp) - for session to start from first event.
with session - where session_id not null, select user_id, max(timestamp) - for session to start from last
event of user, with session, so same session could be continued based on conditions.
*/
func getNextSessionInfoFromDB(projectId uint64, withSession bool, sessionEventNameId uint64,
	maxLookbackTimestamp int64) ([]NextSessionInfo, int) {

	var nextSessionInfoList []NextSessionInfo

	sessionExistStr := "NOT NULL"
	startTimestampAggrFunc := "max"
	if !withSession {
		sessionExistStr = "NULL"
		startTimestampAggrFunc = "min"
	}
	selectStmnt := fmt.Sprintf("user_id, %s(timestamp) as start_timestamp",
		startTimestampAggrFunc)

	db := C.GetServices().Db
	query := db.Table("events").Group("user_id").
		Where("project_id = ? AND event_name_id != ?", projectId, sessionEventNameId).
		Where(fmt.Sprintf("session_id IS %s AND properties->>'%s' IS NULL",
			sessionExistStr, U.EP_SKIP_SESSION)).
		Select(selectStmnt)

	if maxLookbackTimestamp > 0 {
		query = query.Where("timestamp > ?", maxLookbackTimestamp)
	}

	err := query.Find(&nextSessionInfoList).Error
	if err != nil {
		log.WithError(err).Errorf("Failed to get next session info with session %s.", sessionExistStr)
		return nextSessionInfoList, http.StatusInternalServerError
	}

	if len(nextSessionInfoList) == 0 {
		return nextSessionInfoList, http.StatusNotFound
	}

	return nextSessionInfoList, http.StatusFound
}

/*
getNextSessionInfo - Returns user_id and start_timestamp of events for next session creation.
Users with session already (within max_lookback, if given) will have a start_timestamp of last event with session.
Users without session already (within max_lookback, if given) will have a start_timestamp of first event.
*/
func getNextSessionInfo(projectId, sessionEventNameId uint64,
	maxLookbackTimestamp int64) ([]NextSessionInfo, int) {

	logCtx := log.WithFields(log.Fields{"project_id": projectId,
		"max_lookback_timestamp": maxLookbackTimestamp})
	startTimestamp := U.TimeNowUnix()
	logCtx.Info("Started getting next session info.")

	// new users.
	nextSessionInfoList, status := getNextSessionInfoFromDB(projectId, false,
		sessionEventNameId, maxLookbackTimestamp)
	if status != http.StatusFound {
		return nextSessionInfoList, status
	}

	// users without session.
	usersWithoutSession := make(map[string]int)
	for i, nsi := range nextSessionInfoList {
		usersWithoutSession[nsi.UserId] = i
	}

	oldUsersNextSessionInfo, status := getNextSessionInfoFromDB(projectId, true,
		sessionEventNameId, maxLookbackTimestamp)
	if status == http.StatusInternalServerError {
		return nextSessionInfoList, status
	}

	// override without_session info with with_session info.
	// only if with_session info has new events without session (part of new_user_map).
	for oni := range oldUsersNextSessionInfo {
		if nni, exists := usersWithoutSession[oldUsersNextSessionInfo[oni].UserId]; exists {
			// This is to avoid having start_timestamp from with_session info,
			// which older than a day but with a new event (from without_session info)
			// lesser than an hour.
			// Avoiding this will keep min among start_timestamp low, which is start
			// timestamp for events download.
			diff := nextSessionInfoList[nni].StartTimestamp - oldUsersNextSessionInfo[oni].StartTimestamp
			if diff <= M.NewUserSessionInactivityInSeconds {
				nextSessionInfoList[nni] = oldUsersNextSessionInfo[oni]
			}
		}
	}

	logCtx = logCtx.WithField("next_session_info_list_size", len(nextSessionInfoList))

	endTimestamp := U.TimeNowUnix()
	timeTakenInMins := (endTimestamp - startTimestamp) / 60
	logCtx = logCtx.WithField("time_taken_in_mins", timeTakenInMins)
	if timeTakenInMins > 3 {
		logCtx.Error("Too much time taken for getting next session info.")
	} else {
		logCtx.Info("Got next session info.")
	}

	return nextSessionInfoList, http.StatusFound
}

// getAllEventsAsUserEventsMap - Returns a map of user:[events...] withing given period,
// excluding session event and event with session_id.
func getAllEventsAsUserEventsMap(projectId, sessionEventNameId uint64,
	startTimestamp, endTimestamp int64) (*map[string][]M.Event, int, int) {

	var userEventsMap map[string][]M.Event

	var events []M.Event
	if startTimestamp == 0 || endTimestamp == 0 {
		return &userEventsMap, 0, http.StatusBadRequest
	}

	logCtx := log.WithFields(log.Fields{"project_id": projectId,
		"start_timestamp": startTimestamp, "end_timestamp": endTimestamp})

	queryStartTime := U.TimeNowUnix()
	db := C.GetServices().Db
	if err := db.Order("timestamp ASC").
		Where("project_id = ? AND event_name_id != ? AND timestamp BETWEEN ? AND ?",
			projectId, sessionEventNameId, startTimestamp, endTimestamp).
		Find(&events).Error; err != nil {

		logCtx.WithError(err).Error("Failed to get all events of project.")
		return &userEventsMap, 0, http.StatusInternalServerError
	}

	queryTimeInSecs := U.TimeNowUnix() - queryStartTime
	logCtx = logCtx.WithField("no_of_events", len(events)).
		WithField("time_taken_in_secs", queryTimeInSecs)

	if queryTimeInSecs >= (2 * 60) {
		logCtx.Error("Too much time taken to download events on get_all_events_as_user_map.")
	}

	// Todo(Dinesh): Build user events map on scan row, to save memory.
	userEventsMap = make(map[string][]M.Event)
	for i := range events {
		if _, exists := userEventsMap[events[i].UserId]; !exists {
			userEventsMap[events[i].UserId] = make([]M.Event, 0, 0)
		}

		userEventsMap[events[i].UserId] = append(userEventsMap[events[i].UserId], events[i])
	}

	logCtx.WithField("no_of_users", len(userEventsMap)).
		Info("Got all events on get_all_events_as_user_map.")

	return &userEventsMap, len(events), http.StatusFound
}

// addSessionByProjectId - Adds session to all users of project.
func addSessionByProjectId(projectId uint64, maxLookbackTimestamp,
	bufferTimeBeforeSessionCreateInSecs int64) (*Status, int) {

	status := &Status{}

	logCtx := log.WithField("project_id", projectId).
		WithField("max_lookback", maxLookbackTimestamp).
		WithField("buffer_time_in_mins", bufferTimeBeforeSessionCreateInSecs)

	logCtx.Info("Adding session for project.")

	sessionEventName, errCode := M.CreateOrGetSessionEventName(projectId)
	if errCode != http.StatusCreated && errCode != http.StatusConflict {
		logCtx.Error("Failed to create session event name.")
		return status, http.StatusInternalServerError
	}

	nextSessionInfoList, errCode := getNextSessionInfo(projectId,
		sessionEventName.ID, maxLookbackTimestamp)
	if errCode == http.StatusNotFound {
		logCtx.Info("No new events without session to process. Empty next session info.")
		return status, http.StatusNotModified
	}

	if errCode == http.StatusInternalServerError {
		// log and continue if get next session info failed for a project.
		logCtx.Error("Failed to get next session info.")
		return status, http.StatusInternalServerError
	}

	noOfUsers := len(nextSessionInfoList)

	var minNextSessionStartTimestamp int64
	var minNextSessionUserId string
	for i := range nextSessionInfoList {
		if i == 0 || nextSessionInfoList[i].StartTimestamp < minNextSessionStartTimestamp {
			minNextSessionStartTimestamp = nextSessionInfoList[i].StartTimestamp
			minNextSessionUserId = nextSessionInfoList[i].UserId
		}
	}

	logCtx = logCtx.WithField("start_timestamp", minNextSessionStartTimestamp).
		WithField("user_id", minNextSessionUserId)

	eventsDownloadEndTimestamp := U.TimeNowUnix() - bufferTimeBeforeSessionCreateInSecs
	eventsDownloadIntervalInMins := (eventsDownloadEndTimestamp - minNextSessionStartTimestamp) / 60
	if eventsDownloadIntervalInMins <= 0 {
		return status, http.StatusOK
	}

	status.EventsDownloadIntervalInMins = eventsDownloadIntervalInMins

	if minNextSessionStartTimestamp < U.UnixTimeBeforeDuration(5*time.Hour) {
		logCtx.WithField("interval_in_mins", eventsDownloadIntervalInMins).
			Info("Notification - Interval to download events is greater than 5 hours.")
	}

	// Events will be downloaded all users between min timestamp of last session
	// among all users and current timestamp - buffer window, as we don't process
	// events on buffer window (last 30 mins) for any user.
	userEventsMap, noOfEvents, errCode := getAllEventsAsUserEventsMap(
		projectId, sessionEventName.ID, minNextSessionStartTimestamp,
		eventsDownloadEndTimestamp)
	if errCode != http.StatusFound {
		logCtx.Error("Failed to get user events map on add session for project.")
		return status, http.StatusInternalServerError
	}

	status.NoOfEvents = noOfEvents
	status.NoOfUsers = noOfUsers

	var noOfEventsProcessedForSession, noOfSessionsCreated, noOfSessionsContinued, noOfUserPropertiesUpdates int
	for _, nsi := range nextSessionInfoList {
		noOfProcessedEvents, noOfCreated, isContinuedFirst, noOfUserPropUpdates, errCode := M.AddSessionForUser(projectId,
			nsi.UserId, (*userEventsMap)[nsi.UserId], bufferTimeBeforeSessionCreateInSecs,
			nsi.StartTimestamp, sessionEventName.ID)
		if errCode == http.StatusInternalServerError || errCode == http.StatusBadRequest {
			return status, http.StatusInternalServerError
		}

		if isContinuedFirst {
			noOfSessionsContinued++
		}

		noOfSessionsCreated = noOfSessionsCreated + noOfCreated
		noOfEventsProcessedForSession = noOfEventsProcessedForSession + noOfProcessedEvents
		noOfUserPropertiesUpdates = noOfUserPropertiesUpdates + noOfUserPropUpdates
	}

	status.NoOfSessionsCreated = noOfSessionsCreated
	status.NoOfSessionsContinued = noOfSessionsContinued
	status.NoOfEventsProcessed = noOfEventsProcessedForSession
	status.NoOfUserPropertiesUpdates = noOfUserPropertiesUpdates

	return status, http.StatusOK
}

func GetAddSessionAllowedProjects(allowedProjectsList, disallowedProjectsList string) ([]uint64, int) {
	isAllProjects, projectIDsMap, skipProjectIdMap := C.GetProjectsFromListWithAllProjectSupport(
		allowedProjectsList, disallowedProjectsList)

	projectIDs := C.ProjectIdsFromProjectIdBoolMap(projectIDsMap)

	if !isAllProjects {
		if len(projectIDs) == 0 {
			return projectIDs, http.StatusNotFound
		}

		return projectIDs, http.StatusFound
	}

	projectIds, errCode := M.GetAllProjectIDs()
	if errCode != http.StatusFound {
		return projectIds, errCode
	}

	if len(disallowedProjectsList) == 0 {
		return projectIds, http.StatusFound
	}

	allowedProjectIds := make([]uint64, 0, len(projectIds))
	for i, cpid := range projectIds {
		if _, exists := skipProjectIdMap[cpid]; !exists {
			allowedProjectIds = append(allowedProjectIds, projectIds[i])
		}
	}

	return allowedProjectIds, http.StatusFound
}

func setStatus(projectId uint64, statusMap *map[uint64]Status, status *Status,
	hasFailures *bool, isFailed bool, statusLock *sync.Mutex) {

	defer (*statusLock).Unlock()

	(*statusLock).Lock()
	(*statusMap)[projectId] = *status
	*hasFailures = isFailed
}

type Status struct {
	Status                       string `json:"status"`
	EventsDownloadIntervalInMins int64  `json:"events_download_interval_in_mins"`
	NoOfEvents                   int    `json:"no_of_events_downloaded"`
	// count of after filter events. actual no.of events processed for session.
	NoOfEventsProcessed       int `json:"no_of_events_processed"`
	NoOfUsers                 int `json:"no_of_users"`
	NoOfSessionsContinued     int `json:"no_of_sessions_continued"`
	NoOfSessionsCreated       int `json:"no_of_sessions_created"`
	NoOfUserPropertiesUpdates int `json:"no_of_user_properties_updates"`
}

func addSessionWorker(projectId uint64, maxLookbackTimestamp,
	bufferTimeBeforeSessionCreateInSecs int64, statusMap *map[uint64]Status,
	hasFailures *bool, wg *sync.WaitGroup, statusLock *sync.Mutex) {

	defer (*wg).Done()

	logCtx := log.WithField("project_id", projectId)

	startTimestamp := U.TimeNowUnix()
	status, errCode := addSessionByProjectId(projectId, maxLookbackTimestamp,
		bufferTimeBeforeSessionCreateInSecs)
	logCtx = logCtx.WithField("time_taken_in_secs", U.TimeNowUnix()-startTimestamp).
		WithField("status", status)

	if errCode != http.StatusOK {
		var isFailed bool

		if errCode == http.StatusNotModified {
			isFailed = false
			status.Status = StatusNotModified

			logCtx.Info("No events to add session.")
		} else {
			isFailed = true
			status.Status = StatusFailed

			logCtx.Error("Failed to add session.")
		}

		setStatus(projectId, statusMap, status, hasFailures, isFailed, statusLock)
		return
	}

	logCtx.Info("Added session for project.")

	status.Status = "success"
	setStatus(projectId, statusMap, status, hasFailures, false, statusLock)
}

func AddSession(projectIds []uint64, maxLookbackTimestamp,
	bufferTimeBeforeSessionCreateInMins int64,
	numRoutines int) (map[uint64]Status, error) {

	hasFailures := false
	statusMap := make(map[uint64]Status, 0)
	var statusLock sync.Mutex

	if numRoutines == 0 {
		numRoutines = 1
	}

	// breaks list of projectIds into multiple
	// chunks of lenght num_routines to run inside
	// go routines.
	chunkProjectIds := make([][]uint64, 0, 0)
	for i := 0; i < len(projectIds); {
		next := i + numRoutines
		if next > len(projectIds) {
			next = len(projectIds)
		}

		chunkProjectIds = append(chunkProjectIds, projectIds[i:next])
		i = next
	}

	bufferTimeBeforeSessionCreateInSecs := bufferTimeBeforeSessionCreateInMins * 60

	for ci := range chunkProjectIds {
		var wg sync.WaitGroup
		wg.Add(len(chunkProjectIds[ci]))
		for pi := range chunkProjectIds[ci] {
			go addSessionWorker(chunkProjectIds[ci][pi], maxLookbackTimestamp,
				bufferTimeBeforeSessionCreateInSecs,
				&statusMap, &hasFailures, &wg, &statusLock)
		}
		wg.Wait()
	}

	if hasFailures {
		return statusMap, errors.New("failures while adding session")
	}

	return statusMap, nil
}
