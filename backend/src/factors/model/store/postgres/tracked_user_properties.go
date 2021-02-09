package postgres

import (
	"encoding/json"
	C "factors/config"
	"factors/model/model"
	"net/http"
	"time"

	U "factors/util"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

//CreateFactorsTrackedUserProperty - Inserts the tracked event to db
func (pg *Postgres) CreateFactorsTrackedUserProperty(ProjectID uint64, UserPropertyName string, agentUUID string) (int64, int) {
	if pg.isActiveFactorsTrackedUserPropertiesLimitExceeded(ProjectID) {
		return 0, http.StatusBadRequest
	}
	logCtx := log.WithFields(log.Fields{"project_id": ProjectID})
	db := C.GetServices().Db
	insertType := U.UserCreated
	if agentUUID == "" {
		insertType = U.AutoTracked
	}
	transTime := gorm.NowFunc()
	if !pg.IsUserPropertyValid(ProjectID, UserPropertyName) {
		return 0, http.StatusNotFound
	}
	existingFactorsTrackedUserProperty, dbErr := pg.GetFactorsTrackedUserProperty(UserPropertyName, ProjectID)
	if dbErr == nil {
		if existingFactorsTrackedUserProperty.IsActive == false {
			updatedFields := map[string]interface{}{
				"is_active":  true,
				"updated_at": transTime,
			}
			return updateFactorsTrackedUserProperty(existingFactorsTrackedUserProperty.ID, ProjectID, updatedFields)
		}
		logCtx.Error("Create tracked user property failed")
		return 0, http.StatusConflict // Janani: return error

	} else if dbErr.Error() == "record not found" {
		var trackedUserProperty model.FactorsTrackedUserProperty
		if insertType == "UC" {
			trackedUserProperty = model.FactorsTrackedUserProperty{
				ProjectID:        ProjectID,
				UserPropertyName: UserPropertyName,
				Type:             insertType,
				LastTrackedAt:    new(time.Time),
				CreatedBy:        &agentUUID,
				IsActive:         true,
				CreatedAt:        &transTime,
				UpdatedAt:        &transTime,
			}
		} else {
			trackedUserProperty = model.FactorsTrackedUserProperty{
				ProjectID:        ProjectID,
				UserPropertyName: UserPropertyName,
				Type:             insertType,
				LastTrackedAt:    new(time.Time),
				IsActive:         true,
				CreatedAt:        &transTime,
				UpdatedAt:        &transTime,
			}
		}
		if err := db.Create(&trackedUserProperty).Error; err != nil {
			logCtx.WithError(dbErr).Error("Insert into tracked_user_property table failed")
			return 0, http.StatusInternalServerError
		}
		return int64(trackedUserProperty.ID), http.StatusCreated
	} else {
		logCtx.WithError(dbErr).Error("Tracked User Property creation Failed")
		return 0, http.StatusInternalServerError
	}
}

// RemoveFactorsTrackedUserProperty - Mark the tracked event inactive
func (pg *Postgres) RemoveFactorsTrackedUserProperty(ID int64, ProjectID uint64) (int64, int) {
	logCtx := log.WithFields(log.Fields{"project_id": ProjectID})
	transTime := gorm.NowFunc()
	existingFactorsTrackedUserProperty, dbErr := pg.GetFactorsTrackedUserPropertyByID(ID, ProjectID)
	if dbErr == nil {
		if existingFactorsTrackedUserProperty.IsActive == true {
			updatedFields := map[string]interface{}{
				"is_active":  false,
				"updated_at": transTime,
			}
			goals, errCode := pg.GetAllActiveFactorsGoals(ProjectID)
			if errCode != 302 {
				return 0, http.StatusInternalServerError
			}
			for _, goal := range goals {
				rule := model.FactorsGoalRule{}
				json.Unmarshal(goal.Rule.RawMessage, &rule)
				if isUserPropertyInList(rule.Rule.StartEnUserFitler, existingFactorsTrackedUserProperty.UserPropertyName) ||
					isUserPropertyInList(rule.Rule.EndEnUserFitler, existingFactorsTrackedUserProperty.UserPropertyName) ||
					isUserPropertyInList(rule.Rule.GlobalFilters, existingFactorsTrackedUserProperty.UserPropertyName) {
					_, errCode := pg.DeactivateFactorsGoal(int64(goal.ID), goal.ProjectID)
					if errCode != 200 {
						return 0, http.StatusInternalServerError
					}
				}
			}
			return updateFactorsTrackedUserProperty(existingFactorsTrackedUserProperty.ID, ProjectID, updatedFields)
		}
		logCtx.Error("Remove Tracked User Property failed")
		return 0, http.StatusConflict // Janani: return error

	}
	logCtx.WithError(dbErr).Error("Tracked User Property not found")
	return 0, http.StatusNotFound
}

func isUserPropertyInList(properties []model.KeyValueTuple, searchKey string) bool {
	for _, property := range properties {
		if property.Key == searchKey {
			return true
		}
	}
	return false
}

// GetAllFactorsTrackedUserPropertiesByProject - get all the tracked user properties by project
func (pg *Postgres) GetAllFactorsTrackedUserPropertiesByProject(ProjectID uint64) ([]model.FactorsTrackedUserProperty, int) {
	logCtx := log.WithFields(log.Fields{"project_id": ProjectID})
	db := C.GetServices().Db

	var trackedUserProperties []model.FactorsTrackedUserProperty
	if err := db.Limit(1000).Where("project_id = ?", ProjectID).Find(&trackedUserProperties).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusFound
		}
		logCtx.WithError(err).Error("Get All Tracked User Properties failed")
		return nil, http.StatusInternalServerError
	}
	return trackedUserProperties, http.StatusFound
}

// GetAllActiveFactorsTrackedUserPropertiesByProject - get all the tracked user properties by project
func (pg *Postgres) GetAllActiveFactorsTrackedUserPropertiesByProject(ProjectID uint64) ([]model.FactorsTrackedUserProperty, int) {
	logCtx := log.WithFields(log.Fields{"project_id": ProjectID})
	db := C.GetServices().Db

	var trackedUserProperties []model.FactorsTrackedUserProperty
	if err := db.Limit(1000).Where("project_id = ?", ProjectID).Where("is_active = ?", true).Find(&trackedUserProperties).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusFound
		}
		logCtx.WithError(err).Error("Get All Tracked User Properties failed")
		return nil, http.StatusInternalServerError
	}
	return trackedUserProperties, http.StatusFound
}

// GetFactorsTrackedUserProperty - Get tracked user property
func (pg *Postgres) GetFactorsTrackedUserProperty(UserPropertyName string, ProjectID uint64) (*model.FactorsTrackedUserProperty, error) {
	logCtx := log.WithFields(log.Fields{"project_id": ProjectID})
	db := C.GetServices().Db

	var trackedUserProperty model.FactorsTrackedUserProperty
	if err := db.Where("user_property_name = ?", UserPropertyName).Where("project_id = ?", ProjectID).Take(&trackedUserProperty).Error; err != nil {
		logCtx.WithFields(log.Fields{"ProjectID": ProjectID}).WithError(err).Error(
			"Getting tracked user property failed on getFactorsTrackedUserProperty")
		if gorm.IsRecordNotFoundError(err) {
			return nil, err
		}
		return nil, err
	}
	return &trackedUserProperty, nil
}

// GetFactorsTrackedUserPropertyByID - Get tracked user property
func (pg *Postgres) GetFactorsTrackedUserPropertyByID(ID int64, ProjectID uint64) (*model.FactorsTrackedUserProperty, error) {
	logCtx := log.WithFields(log.Fields{"project_id": ProjectID})
	db := C.GetServices().Db

	var trackedUserProperty model.FactorsTrackedUserProperty
	if err := db.Where("id = ?", ID).Where("project_id = ?", ProjectID).Take(&trackedUserProperty).Error; err != nil {
		logCtx.WithFields(log.Fields{"ProjectID": ProjectID}).WithError(err).Error(
			"Getting tracked user property failed on getFactorsTrackedUserProperty")
		if gorm.IsRecordNotFoundError(err) {
			return nil, err
		}
		return nil, err
	}
	return &trackedUserProperty, nil
}

func updateFactorsTrackedUserProperty(FactorsTrackedUserPropertyID uint64, ProjectID uint64, updatedFields map[string]interface{}) (int64, int) {
	logCtx := log.WithFields(log.Fields{"project_id": ProjectID})
	db := C.GetServices().Db
	dbErr := db.Model(&model.FactorsTrackedUserProperty{}).Where("project_id = ? AND id = ?", ProjectID, FactorsTrackedUserPropertyID).Update(updatedFields).Error
	if dbErr != nil {
		logCtx.WithError(dbErr).Error("Updating tracked_user property failed")
		return 0, http.StatusInternalServerError
	}
	return int64(FactorsTrackedUserPropertyID), http.StatusOK
}

func (pg *Postgres) IsUserPropertyValid(ProjectID uint64, UserProperty string) bool {
	logCtx := log.WithFields(log.Fields{"project_id": ProjectID})
	allUserPropertiesByCategory, err := pg.GetUserPropertiesByProject(ProjectID, 10000, 30)
	if err != nil {
		log.WithError(err).Error("Get user Properties from cache failed")
		return false
	}
	userPropertiesMap := make(map[string]bool)
	for _, properties := range allUserPropertiesByCategory {
		for _, property := range properties {
			userPropertiesMap[property] = true
		}
	}
	if userPropertiesMap[UserProperty] == false {
		logCtx.Error("User Property not associated with project")
		return false
	}
	return true
}

func (pg *Postgres) isActiveFactorsTrackedUserPropertiesLimitExceeded(ProjectID uint64) bool {
	trackedUserProperties, errCode := pg.GetAllActiveFactorsTrackedUserPropertiesByProject(ProjectID)
	if errCode != http.StatusFound {
		return true
	}
	if len(trackedUserProperties) >= C.GetFactorsTrackedUserPropertiesLimit() {
		return true
	}
	return false
}
