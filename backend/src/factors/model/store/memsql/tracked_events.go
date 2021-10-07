package memsql

import (
	"encoding/json"
	C "factors/config"
	"factors/model/model"
	"fmt"
	"net/http"
	"time"

	U "factors/util"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) satisfiesTrackedEventForeignConstraints(trackedEvent model.FactorsTrackedEvent) int {
	_, projectErrCode := store.GetProject(trackedEvent.ProjectID)
	if projectErrCode != http.StatusFound {
		return http.StatusBadRequest
	}

	if trackedEvent.CreatedBy != nil && *trackedEvent.CreatedBy != "" {
		_, agentErrCode := store.GetAgentByUUID(*trackedEvent.CreatedBy)
		if agentErrCode != http.StatusFound {
			return http.StatusBadRequest
		}
	}
	return http.StatusOK
}

//CreateTrackedEvent - Inserts the tracked event to db
func (store *MemSQL) CreateFactorsTrackedEvent(ProjectID uint64, EventName string, agentUUID string) (int64, int) {
	if store.isActiveFactorsTrackedEventsLimitExceeded(ProjectID) {
		return 0, http.StatusBadRequest
	}
	logCtx := log.WithFields(log.Fields{"project_id": ProjectID})
	db := C.GetServices().Db
	insertType := U.UserCreated
	if agentUUID == "" {
		insertType = U.AutoTracked
	}
	transTime := gorm.NowFunc()

	eventData, err := store.GetEventName(EventName, ProjectID)
	if err != http.StatusFound {
		logCtx.Error("Get Event details failed")
		return 0, http.StatusNotFound // Janani: return error
	}
	existingFactorsTrackedEvent, dbErr := store.GetFactorsTrackedEvent(eventData.ID, ProjectID)
	if dbErr == nil {
		if existingFactorsTrackedEvent.IsActive == false {
			updatedFields := map[string]interface{}{
				"is_active":  true,
				"updated_at": transTime,
			}
			logCtx.Info("Event already exist")
			return updateFactorsTrackedEvent(existingFactorsTrackedEvent.ID, ProjectID, updatedFields)
		}
		logCtx.Error("Create tracked Event failed")
		return 0, http.StatusConflict // Janani: return error

	} else if dbErr.Error() == "record not found" {
		var trackedEvent model.FactorsTrackedEvent
		if insertType == "UC" {
			trackedEvent = model.FactorsTrackedEvent{
				ProjectID:     ProjectID,
				EventNameID:   eventData.ID,
				Type:          insertType,
				CreatedBy:     &agentUUID,
				LastTrackedAt: new(time.Time),
				IsActive:      true,
				CreatedAt:     &transTime,
				UpdatedAt:     &transTime,
			}
		} else {
			trackedEvent = model.FactorsTrackedEvent{
				ProjectID:     ProjectID,
				EventNameID:   eventData.ID,
				Type:          insertType,
				LastTrackedAt: new(time.Time),
				IsActive:      true,
				CreatedAt:     &transTime,
				UpdatedAt:     &transTime,
			}
		}

		if errCode := store.satisfiesTrackedEventForeignConstraints(trackedEvent); errCode != http.StatusOK {
			return 0, http.StatusInternalServerError
		}
		if err := db.Create(&trackedEvent).Error; err != nil {
			logCtx.WithError(dbErr).Error("Insert into tracked_events table failed")
			return 0, http.StatusInternalServerError
		}
		return int64(trackedEvent.ID), http.StatusCreated
	} else {
		logCtx.WithError(dbErr).Error("Tracked Event creation Failed")
		return 0, http.StatusInternalServerError
	}
}

// DeactivateFactorsTrackedEvent - Mark the tracked event inactive
func (store *MemSQL) DeactivateFactorsTrackedEvent(ID int64, ProjectID uint64) (int64, int) {
	logCtx := log.WithFields(log.Fields{"project_id": ProjectID})
	transTime := gorm.NowFunc()
	existingFactorsTrackedEvent, dbErr := store.GetFactorsTrackedEventByID(ID, ProjectID)
	if dbErr == nil {
		if existingFactorsTrackedEvent.IsActive == true {
			updatedFields := map[string]interface{}{
				"is_active":  false,
				"updated_at": transTime,
			}
			eventDetails, err := store.GetEventNameFromEventNameId(existingFactorsTrackedEvent.EventNameID, ProjectID)
			if err != nil {
				return 0, http.StatusBadRequest
			}
			goals, errCode := store.GetAllActiveFactorsGoals(ProjectID)
			if errCode != 302 {
				return 0, http.StatusInternalServerError
			}
			for _, goal := range goals {
				rule := model.FactorsGoalRule{}
				json.Unmarshal(goal.Rule.RawMessage, &rule)
				if rule.StartEvent == eventDetails.Name || rule.EndEvent == eventDetails.Name {
					_, errCode := store.DeactivateFactorsGoal(int64(goal.ID), goal.ProjectID)
					if errCode != 200 {
						return 0, http.StatusInternalServerError
					}
				}
			}
			return updateFactorsTrackedEvent(existingFactorsTrackedEvent.ID, ProjectID, updatedFields)
		}
		logCtx.Error("Remove Tracked Event failed")
		return 0, http.StatusConflict

	}
	logCtx.WithError(dbErr).Error("Tracked Events not found")
	return 0, http.StatusNotFound
}

// GetAllFactorsTrackedEventsByProject - get all the tracked events by project
func (store *MemSQL) GetAllFactorsTrackedEventsByProject(ProjectID uint64) ([]model.FactorsTrackedEventInfo, int) {
	db := C.GetServices().Db

	queryStr := fmt.Sprintf("WITH tracked_events AS( SELECT * FROM factors_tracked_events WHERE project_id = %d LIMIT %d) "+
		"SELECT tracked_events.*, event_names.name FROM tracked_events INNER JOIN event_names "+
		"ON tracked_events.event_name_id = event_names.id AND event_names.project_id = %d;", ProjectID, 10000, ProjectID)
	trackedEvents := make([]model.FactorsTrackedEventInfo, 0)
	rows, err := db.Raw(queryStr).Rows()
	if err != nil {
		log.WithError(err).Error("Tracked events get all failed")
		return nil, http.StatusInternalServerError
	}

	var trackedEvent model.FactorsTrackedEventInfo
	for rows.Next() {
		if dbErr := db.ScanRows(rows, &trackedEvent); dbErr != nil {
			log.WithError(dbErr).Error("Tracked events get all parsing failed")
			return nil, http.StatusInternalServerError
		}
		trackedEvents = append(trackedEvents, trackedEvent)
	}
	return trackedEvents, http.StatusFound
}

// GetAllActiveFactorsTrackedEventsByProject - get all the tracked events by project
func (store *MemSQL) GetAllActiveFactorsTrackedEventsByProject(ProjectID uint64) ([]model.FactorsTrackedEventInfo, int) {
	db := C.GetServices().Db

	queryStr := fmt.Sprintf("WITH tracked_events AS( SELECT * FROM factors_tracked_events WHERE project_id = %d AND is_active = true LIMIT %d) "+
		"SELECT tracked_events.*, event_names.name FROM tracked_events INNER JOIN event_names "+
		"ON tracked_events.event_name_id = event_names.id AND event_names.project_id = %d;", ProjectID, 10000, ProjectID)
	trackedEvents := make([]model.FactorsTrackedEventInfo, 0)
	rows, err := db.Raw(queryStr).Rows()
	if err != nil {
		log.WithError(err).Error("Tracked events get all failed")
		return nil, http.StatusInternalServerError
	}

	var trackedEvent model.FactorsTrackedEventInfo
	for rows.Next() {
		if dbErr := db.ScanRows(rows, &trackedEvent); dbErr != nil {
			log.WithError(dbErr).Error("Tracked events get all parsing failed")
			return nil, http.StatusInternalServerError
		}
		trackedEvents = append(trackedEvents, trackedEvent)
	}
	return trackedEvents, http.StatusFound
}

// GetFactorsTrackedEvent - Get details of tracked event
func (store *MemSQL) GetFactorsTrackedEvent(EventNameID string, ProjectID uint64) (*model.FactorsTrackedEvent, error) {
	logCtx := log.WithFields(log.Fields{"project_id": ProjectID})
	db := C.GetServices().Db

	var trackedEvent model.FactorsTrackedEvent
	if err := db.Where("event_name_id = ?", EventNameID).Where("project_id = ?", ProjectID).Take(&trackedEvent).Error; err != nil {
		logCtx.WithFields(log.Fields{"ProjectID": ProjectID}).WithError(err).Error(
			"Getting tracked event failed on GetFactorsTrackedEvent")
		if gorm.IsRecordNotFoundError(err) {
			return nil, err
		}
		return nil, err
	}
	return &trackedEvent, nil
}

// GetFactorsTrackedEventByID - Get details of tracked event
func (store *MemSQL) GetFactorsTrackedEventByID(ID int64, ProjectID uint64) (*model.FactorsTrackedEvent, error) {
	logCtx := log.WithFields(log.Fields{"project_id": ProjectID})
	db := C.GetServices().Db

	var trackedEvent model.FactorsTrackedEvent
	if err := db.Where("id = ?", ID).Where("project_id = ?", ProjectID).Take(&trackedEvent).Error; err != nil {
		logCtx.WithFields(log.Fields{"ProjectID": ProjectID}).WithError(err).Error(
			"Getting tracked event failed on GetFactorsTrackedEvent")
		if gorm.IsRecordNotFoundError(err) {
			return nil, err
		}
		return nil, err
	}
	return &trackedEvent, nil
}

func updateFactorsTrackedEvent(FactorsTrackedEventID uint64, ProjectID uint64, updatedFields map[string]interface{}) (int64, int) {
	logCtx := log.WithFields(log.Fields{"project_id": ProjectID})
	db := C.GetServices().Db
	dbErr := db.Model(&model.FactorsTrackedEvent{}).Where("project_id = ? AND id = ?", ProjectID, FactorsTrackedEventID).Update(updatedFields).Error
	if dbErr != nil {
		logCtx.WithError(dbErr).Error("updating tracked_event failed")
		return 0, http.StatusInternalServerError
	}
	return int64(FactorsTrackedEventID), http.StatusOK
}

func (store *MemSQL) isActiveFactorsTrackedEventsLimitExceeded(ProjectID uint64) bool {
	trackedEvents, errCode := store.GetAllActiveFactorsTrackedEventsByProject(ProjectID)
	if errCode != http.StatusFound {
		return true
	}
	if len(trackedEvents) >= C.GetFactorsTrackedEventsLimit() {
		return true
	}
	return false
}
