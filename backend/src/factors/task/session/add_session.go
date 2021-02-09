package session

import (
	"errors"
	"fmt"
	"net/http"
	"sync"

	log "github.com/sirupsen/logrus"

	C "factors/config"
	"factors/model/model"
	"factors/model/store"
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
	maxLookbackTimestamp, hardLimitTimestamp int64) ([]NextSessionInfo, int) {

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

	if hardLimitTimestamp > 0 {
		query = query.Where("timestamp <= ?", hardLimitTimestamp)
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
	maxLookbackTimestamp, hardLimitTimestamp int64) ([]NextSessionInfo, int) {

	logCtx := log.WithFields(log.Fields{"project_id": projectId,
		"max_lookback_timestamp": maxLookbackTimestamp})
	startTimestamp := U.TimeNowUnix()
	logCtx.Info("Started getting next session info.")

	// new users.
	nextSessionInfoList, status := getNextSessionInfoFromDB(projectId, false,
		sessionEventNameId, maxLookbackTimestamp, hardLimitTimestamp)
	if status != http.StatusFound {
		return nextSessionInfoList, status
	}

	// users without session.
	usersWithoutSession := make(map[string]int)
	for i, nsi := range nextSessionInfoList {
		usersWithoutSession[nsi.UserId] = i
	}

	oldUsersNextSessionInfo, status := getNextSessionInfoFromDB(projectId, true,
		sessionEventNameId, maxLookbackTimestamp, hardLimitTimestamp)
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
			if diff <= model.NewUserSessionInactivityInSeconds {
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
	startTimestamp, endTimestamp int64) (*map[string][]model.Event, int, int) {

	logCtx := log.WithFields(log.Fields{"project_id": projectId,
		"start_timestamp": startTimestamp, "end_timestamp": endTimestamp})

	var userEventsMap map[string][]model.Event
	var events []model.Event
	if startTimestamp == 0 || endTimestamp == 0 {
		logCtx.Error("Invalid start_timestamp or end_timestamp.")
		return &userEventsMap, 0, http.StatusInternalServerError
	}

	queryStartTime := U.TimeNowUnix()
	db := C.GetServices().Db
	// Ordered by timestamp, created_at to fix the order for events with same
	// timestamp, as event timestamp is in seconds. This fixes the invalid first
	// event used for enrichment.
	excludeSkipSessionCondition := fmt.Sprintf("(properties->>'%s' IS NULL OR properties->>'%s' = ?)",
		U.EP_SKIP_SESSION, U.EP_SKIP_SESSION)
	if err := db.Order("timestamp, created_at ASC").
		Where("project_id = ? AND event_name_id != ? AND timestamp BETWEEN ? AND ?"+" AND "+excludeSkipSessionCondition,
			projectId, sessionEventNameId, startTimestamp, endTimestamp, false).
		Find(&events).Error; err != nil {

		logCtx.WithError(err).Error("Failed to get all events of project.")
		return &userEventsMap, 0, http.StatusInternalServerError
	}

	if len(events) == 0 {
		return &userEventsMap, 0, http.StatusNotFound
	}

	queryTimeInSecs := U.TimeNowUnix() - queryStartTime
	logCtx = logCtx.WithField("no_of_events", len(events)).
		WithField("time_taken_in_secs", queryTimeInSecs)

	if queryTimeInSecs >= (2 * 60) {
		logCtx.Error("Too much time taken to download events on get_all_events_as_user_map.")
	}

	userEventsMap = make(map[string][]model.Event)
	for i := range events {
		if _, exists := userEventsMap[events[i].UserId]; !exists {
			userEventsMap[events[i].UserId] = make([]model.Event, 0, 0)
		} else {
			// Event with session should be added as first event, if available.
			// To support continuation of the session.
			currentUserEventHasSession := events[i].SessionId != nil
			firstUserEventHasSession := userEventsMap[events[i].UserId][0].SessionId != nil
			userHasNoSessionEvent := firstUserEventHasSession && len(userEventsMap[events[i].UserId]) > 1
			if !userHasNoSessionEvent && firstUserEventHasSession && currentUserEventHasSession {
				// Add current event as first event.
				userEventsMap[events[i].UserId][0] = events[i]
				continue
			}
		}

		userEventsMap[events[i].UserId] = append(userEventsMap[events[i].UserId], events[i])
	}

	logCtx.WithField("no_of_users", len(userEventsMap)).
		Info("Got all events on get_all_events_as_user_map.")

	return &userEventsMap, len(events), http.StatusFound
}

// addSessionByProjectId - Adds session to all users of project.
func addSessionByProjectId(projectId uint64, maxLookbackTimestamp, startTimestamp,
	endTimestamp, bufferTimeBeforeSessionCreateInSecs int64) (*Status, int) {

	status := &Status{}

	logCtx := log.WithField("project_id", projectId).
		WithField("max_lookback", maxLookbackTimestamp).
		WithField("start_timestamp", startTimestamp).
		WithField("end_timestamp", endTimestamp).
		WithField("buffer_time_in_mins", bufferTimeBeforeSessionCreateInSecs)

	logCtx.Info("Adding session for project.")

	sessionEventName, errCode := store.GetStore().CreateOrGetSessionEventName(projectId)
	if errCode != http.StatusCreated && errCode != http.StatusConflict {
		logCtx.Error("Failed to create session event name.")
		return status, http.StatusInternalServerError
	}

	var eventsDownloadStartTimestamp, eventsDownloadEndTimestamp int64

	if startTimestamp > 0 && endTimestamp > 0 {
		// Use the specific window, if given.
		eventsDownloadStartTimestamp = startTimestamp
		eventsDownloadEndTimestamp = endTimestamp
	} else {
		eventsDownloadStartTimestamp, errCode = store.GetStore().GetNextSessionStartTimestampForProject(projectId)
		if errCode != http.StatusFound {
			logCtx.Error("Failed to get last min session timestamp of user for project.")
			return status, http.StatusInternalServerError
		}

		// Log and limit the startTimestamp to maxLookbackTimestamp configured.
		// Case: When one single user has an event which is very old, could cause
		// scanning and downloading of all events from that timestamp on next run.
		if maxLookbackTimestamp > 0 && eventsDownloadStartTimestamp < maxLookbackTimestamp {
			logCtx.Warn("Events download start timestamp is greater than maxLookbackTimestamp.")
			eventsDownloadStartTimestamp = maxLookbackTimestamp
		}

		logCtx = logCtx.WithField("start_timestamp", eventsDownloadStartTimestamp)
		eventsDownloadEndTimestamp = U.TimeNowUnix() - bufferTimeBeforeSessionCreateInSecs
	}

	userEventsMap, noOfEventsDownloaded, errCode := getAllEventsAsUserEventsMap(projectId,
		sessionEventName.ID, eventsDownloadStartTimestamp, eventsDownloadEndTimestamp)
	if errCode == http.StatusInternalServerError {
		logCtx.Error("Failed to get user events map on add session for project.")
		return status, http.StatusInternalServerError
	}
	if errCode == http.StatusNotFound {
		return status, http.StatusNotModified
	}

	status.NoOfEvents = noOfEventsDownloaded
	status.NoOfUsers = len(*userEventsMap)

	var noOfEventsProcessedForSession, noOfSessionsCreated, noOfSessionsContinued, noOfUserPropertiesUpdates int
	for userID, events := range *userEventsMap {
		noOfProcessedEvents, noOfCreated, isContinuedFirst,
			noOfUserPropUpdates, errCode := store.GetStore().AddSessionForUser(projectId, userID, events,
			bufferTimeBeforeSessionCreateInSecs, sessionEventName.ID)
		if errCode == http.StatusInternalServerError || errCode == http.StatusBadRequest {
			return status, http.StatusInternalServerError
		}

		currentNextSessionStartTimestamp, errCode := store.GetStore().GetNextSessionStartTimestampForProject(projectId)
		if errCode != http.StatusFound {
			return status, http.StatusInternalServerError
		}

		// Update the next_session_timestamp for project with session added oldest event across users.
		if currentNextSessionStartTimestamp > events[len(events)-1].Timestamp {
			errCode = store.GetStore().UpdateNextSessionStartTimestampForProject(projectId, events[len(events)-1].Timestamp)
			if errCode != http.StatusAccepted {
				return status, http.StatusInternalServerError
			}
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

	if noOfSessionsCreated == 0 {
		return status, http.StatusNotModified
	}

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

	projectIds, errCode := store.GetStore().GetAllProjectIDs()
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

func addSessionWorker(projectId uint64, maxLookbackTimestamp, startTimestamp, endTimestamp,
	bufferTimeBeforeSessionCreateInSecs int64, statusMap *map[uint64]Status,
	hasFailures *bool, wg *sync.WaitGroup, statusLock *sync.Mutex) {

	defer (*wg).Done()

	logCtx := log.WithField("project_id", projectId)

	execStartTimestamp := U.TimeNowUnix()
	status, errCode := addSessionByProjectId(projectId, maxLookbackTimestamp, startTimestamp,
		endTimestamp, bufferTimeBeforeSessionCreateInSecs)
	logCtx = logCtx.WithField("time_taken_in_secs", U.TimeNowUnix()-execStartTimestamp).
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

func AddSession(projectIds []uint64, maxLookbackTimestamp, startTimestamp, endTimestamp,
	bufferTimeBeforeSessionCreateInMins int64, numRoutines int) (map[uint64]Status, error) {

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
			go addSessionWorker(chunkProjectIds[ci][pi], maxLookbackTimestamp, startTimestamp,
				endTimestamp, bufferTimeBeforeSessionCreateInSecs,
				&statusMap, &hasFailures, &wg, &statusLock)
		}
		wg.Wait()
	}

	if hasFailures {
		return statusMap, errors.New("failures while adding session")
	}

	return statusMap, nil
}
