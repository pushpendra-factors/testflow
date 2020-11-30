package model

import (
	C "factors/config"
	"net/http"
	"time"

	U "factors/util"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

// FactorsTrackedUserProperty - DB model for table: tracked_events
type FactorsTrackedUserProperty struct {
	ID               uint64     `gorm:"primary_key:true;" json:"id"`
	ProjectID        uint64     `json:"project_id"`
	UserPropertyName string     `json:"user_property_name"`
	Type             string     `gorm:"not null;type:varchar(2)" json:"type"`
	CreatedBy        *string    `json:"created_by;default:null"`
	LastTrackedAt    *time.Time `json:"last_tracked_at"`
	IsActive         bool       `json:"is_active"`
	CreatedAt        *time.Time `json:"created_at"`
	UpdatedAt        *time.Time `json:"updated_at"`
}

//CreateFactorsTrackedUserProperty - Inserts the tracked event to db
func CreateFactorsTrackedUserProperty(ProjectID uint64, UserPropertyName string, agentUUID string) (int64, int) {
	if isActiveFactorsTrackedUserPropertiesLimitExceeded(ProjectID) {
		return 0, http.StatusBadRequest
	}
	logCtx := log.WithFields(log.Fields{"project_id": ProjectID})
	db := C.GetServices().Db
	insertType := U.UserCreated
	if agentUUID == "" {
		insertType = U.AutoTracked
	}
	transTime := gorm.NowFunc()
	if !IsUserPropertyValid(ProjectID, UserPropertyName) {
		return 0, http.StatusNotFound
	}
	existingFactorsTrackedUserProperty, dbErr := GetFactorsTrackedUserProperty(UserPropertyName, ProjectID)
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
		var trackedUserProperty FactorsTrackedUserProperty
		if insertType == "UC" {
			trackedUserProperty = FactorsTrackedUserProperty{
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
			trackedUserProperty = FactorsTrackedUserProperty{
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
func RemoveFactorsTrackedUserProperty(ID int64, ProjectID uint64) (int64, int) {
	logCtx := log.WithFields(log.Fields{"project_id": ProjectID})
	transTime := gorm.NowFunc()
	existingFactorsTrackedUserProperty, dbErr := GetFactorsTrackedUserPropertyByID(ID, ProjectID)
	if dbErr == nil {
		if existingFactorsTrackedUserProperty.IsActive == true {
			updatedFields := map[string]interface{}{
				"is_active":  false,
				"updated_at": transTime,
			}
			return updateFactorsTrackedUserProperty(existingFactorsTrackedUserProperty.ID, ProjectID, updatedFields)
		}
		logCtx.Error("Remove Tracked User Property failed")
		return 0, http.StatusConflict // Janani: return error

	}
	logCtx.WithError(dbErr).Error("Tracked User Property not found")
	return 0, http.StatusNotFound
}

// GetAllFactorsTrackedUserPropertiesByProject - get all the tracked user properties by project
func GetAllFactorsTrackedUserPropertiesByProject(ProjectID uint64) ([]FactorsTrackedUserProperty, int) {
	logCtx := log.WithFields(log.Fields{"project_id": ProjectID})
	db := C.GetServices().Db

	var trackedUserProperties []FactorsTrackedUserProperty
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
func GetAllActiveFactorsTrackedUserPropertiesByProject(ProjectID uint64) ([]FactorsTrackedUserProperty, int) {
	logCtx := log.WithFields(log.Fields{"project_id": ProjectID})
	db := C.GetServices().Db

	var trackedUserProperties []FactorsTrackedUserProperty
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
func GetFactorsTrackedUserProperty(UserPropertyName string, ProjectID uint64) (*FactorsTrackedUserProperty, error) {
	logCtx := log.WithFields(log.Fields{"project_id": ProjectID})
	db := C.GetServices().Db

	var trackedUserProperty FactorsTrackedUserProperty
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
func GetFactorsTrackedUserPropertyByID(ID int64, ProjectID uint64) (*FactorsTrackedUserProperty, error) {
	logCtx := log.WithFields(log.Fields{"project_id": ProjectID})
	db := C.GetServices().Db

	var trackedUserProperty FactorsTrackedUserProperty
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
	dbErr := db.Model(&FactorsTrackedUserProperty{}).Where("project_id = ? AND id = ?", ProjectID, FactorsTrackedUserPropertyID).Update(updatedFields).Error
	if dbErr != nil {
		logCtx.WithError(dbErr).Error("Updating tracked_user property failed")
		return 0, http.StatusInternalServerError
	}
	return int64(FactorsTrackedUserPropertyID), http.StatusOK
}

func IsUserPropertyValid(ProjectID uint64, UserProperty string) bool {
	logCtx := log.WithFields(log.Fields{"project_id": ProjectID})
	allUserPropertiesByCategory, err := GetUserPropertiesByProject(ProjectID, 10000, 30)
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

func isActiveFactorsTrackedUserPropertiesLimitExceeded(ProjectID uint64) bool {
	trackedUserProperties, errCode := GetAllActiveFactorsTrackedUserPropertiesByProject(ProjectID)
	if errCode != http.StatusFound {
		return true
	}
	if len(trackedUserProperties) >= C.GetFactorsTrackedUserPropertiesLimit() {
		return true
	}
	return false
}
