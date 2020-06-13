package session

import (
	"errors"
	"fmt"
	"net/http"
	"sync"

	log "github.com/sirupsen/logrus"

	C "factors/config"
	M "factors/model"
	U "factors/util"
)

type NextSessionInfo struct {
	UserId         string `json:"user_id"`
	StartTimestamp int64  `json:"start_timestamp"`
}

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
		Where(fmt.Sprintf("project_id = ? AND event_name_id != ? AND session_id IS %s",
			sessionExistStr), projectId, sessionEventNameId).
		Select(selectStmnt)

	if maxLookbackTimestamp > 0 {
		query = query.Where("timestamp > ?", maxLookbackTimestamp)
	}

	err := query.Find(&nextSessionInfoList).Error
	if err != nil {
		log.WithError(err).Error("Failed to get next session info for old users.")
		return nextSessionInfoList, http.StatusInternalServerError
	}

	return nextSessionInfoList, http.StatusFound
}

func getNextSessionInfo(projectId, sessionEventNameId uint64,
	maxLookbackTimestamp int64) ([]NextSessionInfo, int) {

	// old users.
	nextSessionInfoList, status := getNextSessionInfoFromDB(projectId, true,
		sessionEventNameId, maxLookbackTimestamp)
	if status != http.StatusFound {
		return []NextSessionInfo{}, http.StatusInternalServerError
	}

	// users with session.
	usersWithSession := make(map[string]bool)
	for _, nsi := range nextSessionInfoList {
		usersWithSession[nsi.UserId] = true
	}

	// new users, not part of users_with_session map. need this
	// filter as even old user won't have session for last 30 mins.
	newUsersNextSessionInfo, status := getNextSessionInfoFromDB(projectId, false,
		sessionEventNameId, maxLookbackTimestamp)
	if status != http.StatusFound {
		return nextSessionInfoList, http.StatusInternalServerError
	}

	for i := range newUsersNextSessionInfo {
		if _, exists := usersWithSession[newUsersNextSessionInfo[i].UserId]; !exists {
			nextSessionInfoList = append(nextSessionInfoList, newUsersNextSessionInfo[i])
		}
	}

	return nextSessionInfoList, http.StatusFound
}

// addSessionByProjectId - Adds session to all users of project.
func addSessionByProjectId(projectId uint64, maxLookbackTimestamp int64) int {
	logCtx := log.WithField("project_id", projectId).
		WithField("max_lookback", maxLookbackTimestamp)

	logCtx.Info("Adding session for project.")

	sessionEventName, errCode := M.CreateOrGetSessionEventName(projectId)
	if errCode != http.StatusCreated && errCode != http.StatusConflict {
		logCtx.Error("Failed to create session event name")
		return http.StatusInternalServerError
	}

	nextSessionInfoList, errCode := getNextSessionInfo(projectId,
		sessionEventName.ID, maxLookbackTimestamp)
	if errCode == http.StatusInternalServerError {
		// log and continue if get next session info failed for a project.
		logCtx.Error("Failed to get next session info.")
		return errCode
	}

	nextSessionUserIds := make([]string, len(nextSessionInfoList))
	for i := range nextSessionInfoList {
		nextSessionUserIds = append(nextSessionUserIds, nextSessionInfoList[i].UserId)
	}

	currentTimestamp := U.TimeNowUnix()
	usersPerBatchCount := 100
	latestUserEventMap, errCode := M.GetLatestUserEventForUsersInBatch(projectId, nextSessionUserIds,
		currentTimestamp-M.OneDayInSeconds, currentTimestamp, usersPerBatchCount)
	if errCode != http.StatusFound {
		return errCode
	}

	timeTakenInMinutes := (U.TimeNowUnix() - currentTimestamp) / 60
	if timeTakenInMinutes >= 2 {
		logCtx.WithField("time_taken", timeTakenInMinutes).
			WithField("no_of_users", len(nextSessionUserIds)).
			WithField("users_per_batch", usersPerBatchCount).
			Error("Time taken to get latest event for list of users is high.")
	}

	for _, nsi := range nextSessionInfoList {
		errCode := M.AddSessionForUser(projectId, nsi.UserId, latestUserEventMap[nsi.UserId],
			nsi.StartTimestamp, sessionEventName.ID)
		if errCode == http.StatusInternalServerError || errCode == http.StatusBadRequest {
			return errCode
		}
	}

	return http.StatusOK
}

func GetAddSessionAllowedProjects(allowedProjectsList string) ([]uint64, int) {
	isAllProjects, projectsList := C.GetProjectsFromListWithAllProjectSupport(
		allowedProjectsList)

	if !isAllProjects {
		if len(projectsList) == 0 {
			return projectsList, http.StatusNotFound
		}

		return projectsList, http.StatusFound
	}

	projectIds, errCode := M.GetAllProjectIDs()
	if errCode != http.StatusInternalServerError {
		return projectIds, errCode
	}

	return projectIds, http.StatusFound
}

func setStatus(projectId uint64, statusMap *map[uint64]string, status string,
	hasFailures *bool, isFailed bool, statusLock *sync.Mutex) {

	defer (*statusLock).Unlock()

	(*statusLock).Lock()
	(*statusMap)[projectId] = status
	*hasFailures = isFailed
}

func addSessionWorker(projectId uint64, maxLookbackTimestamp int64, statusMap *map[uint64]string,
	hasFailures *bool, wg *sync.WaitGroup, statusLock *sync.Mutex) {

	defer (*wg).Done()

	errCode := addSessionByProjectId(projectId, maxLookbackTimestamp)
	if errCode == http.StatusInternalServerError {
		setStatus(projectId, statusMap, "failed", hasFailures, true, statusLock)
		return
	}

	setStatus(projectId, statusMap, "success", hasFailures, false, statusLock)
}

func AddSession(projectIds []uint64, maxLookbackTimestamp int64,
	numRoutines int) (map[uint64]string, error) {

	hasFailures := false
	statusMap := make(map[uint64]string, 0)
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

	for ci := range chunkProjectIds {
		var wg sync.WaitGroup
		wg.Add(len(chunkProjectIds[ci]))
		for pi := range chunkProjectIds[ci] {
			go addSessionWorker(chunkProjectIds[ci][pi], maxLookbackTimestamp,
				&statusMap, &hasFailures, &wg, &statusLock)
		}
		wg.Wait()
	}

	if hasFailures {
		return statusMap, errors.New("failures while adding session")
	}

	return statusMap, nil
}
