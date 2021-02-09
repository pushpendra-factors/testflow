package goal_cleanup

import (
	C "factors/config"
	"factors/model/model"

	log "github.com/sirupsen/logrus"
)

//DoGoalCleanUp - Cleaning up Goal table for auto created goals for a specified project id
func DoGoalCleanUp(projectID uint64) int64 {
	db := C.GetServices().Db
	dbObj := db.Where("type = ?", "AT").Where("project_id = ?", projectID).Delete(&model.FactorsGoal{})
	if dbObj.Error != nil {
		log.WithFields(log.Fields{"projectId": projectID}).WithError(db.Error).Error(
			"Deleting from Goal Table failed")
	}
	log.WithField("ProjectId", projectID).WithField("Count", dbObj.RowsAffected).Info("Goals Deleted Count")
	return dbObj.RowsAffected
}

//DoTrackedEventsCleanUp - Cleaning up Tracked events table for auto created tracked events for a specified project id
func DoTrackedEventsCleanUp(projectID uint64) int64 {
	db := C.GetServices().Db
	var trackedEvent model.FactorsTrackedEvent
	dbObj := db.Where("type = ?", "AT").Where("project_id = ?", projectID).Delete(&trackedEvent)
	if dbObj.Error != nil {
		log.WithFields(log.Fields{"projectId": projectID}).WithError(db.Error).Error(
			"Deleting from TrackedEvents Table failed")
	}
	log.WithField("ProjectId", projectID).WithField("Count", dbObj.RowsAffected).Info("TrackedEvents Deleted Count")
	return dbObj.RowsAffected
}

//DoTrackedUserPropertiesCleanUp - Cleaning up Tracked user properties table for auto created tracked user properties for a specified project id
func DoTrackedUserPropertiesCleanUp(projectID uint64) int64 {
	db := C.GetServices().Db
	dbObj := db.Where("type = ?", "AT").Where("project_id = ?", projectID).Delete(&model.FactorsTrackedUserProperty{})
	if dbObj.Error != nil {
		log.WithFields(log.Fields{"projectId": projectID}).WithError(db.Error).Error(
			"Deleting from TrackedUserProperties Table failed")
	}
	log.WithField("ProjectId", projectID).WithField("Count", dbObj.RowsAffected).Info("TrackedUserProperties Deleted Count")
	return dbObj.RowsAffected
}
