package session

import (
	"errors"
	"net/http"
	"sync"

	log "github.com/sirupsen/logrus"

	C "factors/config"
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

	userEventsMap, noOfEventsDownloaded, errCode := store.GetStore().GetAllEventsForSessionCreationAsUserEventsMap(projectId,
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
	var minOfSessionAddedLastEventTimestamp int64

	for userID, events := range *userEventsMap {
		noOfProcessedEvents, noOfCreated, isContinuedFirst,
			noOfUserPropUpdates, errCode := store.GetStore().AddSessionForUser(projectId, userID, events,
			bufferTimeBeforeSessionCreateInSecs, sessionEventName.ID)
		if errCode == http.StatusInternalServerError || errCode == http.StatusBadRequest {
			return status, http.StatusInternalServerError
		}

		lastEventTimestamp := events[len(events)-1].Timestamp
		if minOfSessionAddedLastEventTimestamp == 0 {
			minOfSessionAddedLastEventTimestamp = lastEventTimestamp
		} else if lastEventTimestamp < minOfSessionAddedLastEventTimestamp {
			minOfSessionAddedLastEventTimestamp = lastEventTimestamp
		}

		if isContinuedFirst {
			noOfSessionsContinued++
		}

		noOfSessionsCreated = noOfSessionsCreated + noOfCreated
		noOfEventsProcessedForSession = noOfEventsProcessedForSession + noOfProcessedEvents
		noOfUserPropertiesUpdates = noOfUserPropertiesUpdates + noOfUserPropUpdates
	}

	// Update next sessions start timestamp with min of last session
	// added event timestamp across all users for the project.
	errCode = store.GetStore().UpdateNextSessionStartTimestampForProject(
		projectId, minOfSessionAddedLastEventTimestamp)
	if errCode != http.StatusAccepted {
		return status, http.StatusInternalServerError
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
