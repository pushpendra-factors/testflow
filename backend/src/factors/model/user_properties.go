package model

import (
	"encoding/json"
	C "factors/config"
	U "factors/util"
	"fmt"
	"net"
	"net/http"
	"reflect"
	"time"

	"github.com/imdario/mergo"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

type UserProperties struct {
	// Composite primary key with project_id, user_id and random uuid.
	ID string `gorm:"primary_key:true;uuid;default:uuid_generate_v4()" json:"id"`
	// Below are the foreign key constraints added in creation script.
	// project_id -> projects(id)
	// (project_id, user_id) -> users(project_id, id)
	ProjectId uint64 `gorm:"primary_key:true;" json:"project_id"`
	UserId    string `gorm:"primary_key:true;" json:"user_id"`

	// JsonB of postgres with gorm. https://github.com/jinzhu/gorm/issues/1183
	Properties postgres.Jsonb `json:"properties"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

func createUserPropertiesIfChanged(projectId uint64, userId string,
	currentPropertiesId string, newProperties *postgres.Jsonb) (string, int) {

	if U.IsEmptyPostgresJsonb(newProperties) {
		return currentPropertiesId, http.StatusNotModified
	}

	db := C.GetServices().Db

	currentProperties := &postgres.Jsonb{}
	var statusCode int

	var currentPropertiesMap map[string]interface{}
	var newPropertiesMap map[string]interface{}
	var mergedPropertiesMap map[string]interface{}

	json.Unmarshal((*newProperties).RawMessage, &newPropertiesMap)
	if currentPropertiesId != "" {
		currentProperties, statusCode = GetUserProperties(
			projectId, userId, currentPropertiesId)

		if statusCode == http.StatusInternalServerError {
			log.WithField("current_properties_id", currentPropertiesId).Error(
				"Failed to GetUserProperties on createUserPropertiesIfChanged.")
			return "", http.StatusInternalServerError
		}

		json.Unmarshal((*currentProperties).RawMessage, &currentPropertiesMap)
		// init mergedProperties with currentProperties.
		json.Unmarshal((*currentProperties).RawMessage, &mergedPropertiesMap)
		if statusCode == http.StatusFound {
			// Using merged properties for equality check to achieve
			// currentPropertiesMap {x: 1, y: 2} newPropertiesMap {x: 1} -> true
			mergo.Merge(&mergedPropertiesMap, newPropertiesMap, mergo.WithOverride)
			if reflect.DeepEqual(currentPropertiesMap, mergedPropertiesMap) {
				return currentPropertiesId, http.StatusNotModified
			}
		}
	} else {
		mergedPropertiesMap = newPropertiesMap
	}

	// Overwrite only given values.
	updatedPropertiesBytes, err := json.Marshal(mergedPropertiesMap)
	if err != nil {
		return "", http.StatusInternalServerError
	}
	userProperties := UserProperties{
		UserId:     userId,
		ProjectId:  projectId,
		Properties: postgres.Jsonb{RawMessage: json.RawMessage(updatedPropertiesBytes)},
	}
	log.WithFields(log.Fields{"UserProperties": &userProperties}).Info("Creating user properties")

	if err := db.Create(&userProperties).Error; err != nil {
		log.WithFields(log.Fields{"userProperties": &userProperties}).WithError(err).Error("createUserProperties Failed")
		return "", http.StatusInternalServerError
	}
	return userProperties.ID, http.StatusCreated
}

func GetUserProperties(projectId uint64, userId string, id string) (*postgres.Jsonb, int) {
	db := C.GetServices().Db

	var userProperties UserProperties
	if err := db.Where("project_id = ?", projectId).Where("user_id = ?", userId).Where(
		"id = ?", id).First(&userProperties).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}
		return nil, http.StatusInternalServerError
	}
	return &userProperties.Properties, http.StatusFound
}

func FillLocationUserProperties(properties *U.PropertiesMap, clientIP string) error {
	geo := C.GetServices().GeoLocation

	// ClientIP unavailable.
	if clientIP == "" {
		return fmt.Errorf("invalid IP, failed adding geolocation properties")
	}

	city, err := geo.City(net.ParseIP(clientIP))
	if err != nil {
		log.WithFields(log.Fields{"clientIP": clientIP}).WithError(err).Error(
			"Failed to get city information from geodb")
		return err
	}

	// Using en -> english name.
	if countryName, ok := city.Country.Names["en"]; ok && countryName != "" {
		if c, ok := (*properties)[U.UP_COUNTRY]; !ok || c == "" {
			(*properties)[U.UP_COUNTRY] = countryName
		}
	}

	if cityName, ok := city.City.Names["en"]; ok && cityName != "" {
		if c, ok := (*properties)[U.UP_CITY]; !ok || c == "" {
			(*properties)[U.UP_CITY] = cityName
		}
	}

	return nil
}

func GetUserPropertyRecordsByUserId(projectId uint64, userId string) ([]UserProperties, int) {
	db := C.GetServices().Db

	var userProperties []UserProperties
	if err := db.Where("project_id = ? AND user_id = ?", projectId, userId).Find(&userProperties).Error; err != nil {
		return nil, http.StatusInternalServerError
	}

	if len(userProperties) == 0 {
		return nil, http.StatusNotFound
	}

	return userProperties, http.StatusFound
}

func GetLatestUserPropertiesByUserId(projectId uint64, userId string) (*UserProperties, int) {
	db := C.GetServices().Db

	var userProperties []UserProperties
	if err := db.Limit(1).Order("created_at DESC").Where("project_id = ? AND user_id = ?",
		projectId, userId).Find(&userProperties).Error; err != nil {

		return nil, http.StatusInternalServerError
	}

	if len(userProperties) == 0 {
		return nil, http.StatusNotFound
	}

	return &userProperties[0], http.StatusFound
}

func OverwriteUserProperties(projectId uint64, userId string, id string, propertiesJsonb *postgres.Jsonb) int {
	if projectId == 0 || userId == "" || id == "" {
		return http.StatusBadRequest
	}

	db := C.GetServices().Db
	if err := db.Model(&UserProperties{}).Where("project_id = ? AND user_id = ? AND id = ?",
		projectId, userId, id).Update("properties", propertiesJsonb).Error; err != nil {
		log.WithFields(log.Fields{"project_id": projectId, "id": id}).Error("Failed to replace properties.")
		return http.StatusInternalServerError
	}

	return http.StatusAccepted
}

// Updates given property with value on all user properties for the given user
// adds property if not exist.
func UpdatePropertyOnAllUserPropertyRecords(projectId uint64, userId string,
	property string, value interface{}) int {

	userPropertyRecords, errCode := GetUserPropertyRecordsByUserId(projectId, userId)
	if errCode == http.StatusInternalServerError {
		return errCode
	} else if errCode == http.StatusNotFound {
		return http.StatusBadRequest
	}

	logCtx := log.WithFields(log.Fields{"project_id": projectId, "user_id": userId})

	for _, userProperties := range userPropertyRecords {
		var propertiesMap map[string]interface{}

		if !U.IsEmptyPostgresJsonb(&userProperties.Properties) {
			err := json.Unmarshal(userProperties.Properties.RawMessage, &propertiesMap)
			if err != nil {
				logCtx.Error("Failed to update user property record. JSON unmarshal failed.")
				continue
			}
		} else {
			propertiesMap = make(map[string]interface{}, 0)
		}

		// update not required. Not using AddToPostgresJsonb
		// for this check.
		if pValue, _ := propertiesMap[property]; pValue == value {
			continue
		}

		logCtx = logCtx.WithFields(log.Fields{"properties_id": userProperties.ID, "property": property, "value": value})

		propertiesMap[property] = value
		properitesBytes, err := json.Marshal(propertiesMap)
		if err != nil {
			// log and continue update next user property.
			logCtx.Error("Failed to update user property record. JSON marshal failed.")
			continue
		}
		updatedProperties := postgres.Jsonb{RawMessage: json.RawMessage(properitesBytes)}

		// Triggers multiple updates.
		errCode := OverwriteUserProperties(projectId, userId, userProperties.ID, &updatedProperties)
		if errCode == http.StatusInternalServerError {
			logCtx.WithError(err).Error("Failed to update user property record. DB query failed.")
			continue
		}
	}

	return http.StatusAccepted
}
