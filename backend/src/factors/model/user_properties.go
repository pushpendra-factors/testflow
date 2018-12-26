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
	db := C.GetServices().Db

	currentProperties := &postgres.Jsonb{}
	var statusCode int

	var currentPropertiesMap map[string]interface{}
	var newPropertiesMap map[string]interface{}
	json.Unmarshal((*newProperties).RawMessage, &newPropertiesMap)

	if currentPropertiesId != "" {
		currentProperties, statusCode = getUserProperties(
			projectId, userId, currentPropertiesId)
		json.Unmarshal((*currentProperties).RawMessage, &currentPropertiesMap)
		if statusCode == http.StatusFound {
			if reflect.DeepEqual(currentPropertiesMap, newPropertiesMap) {
				return currentPropertiesId, http.StatusNotModified
			}
		}
	}
	// Overwrite only given values.
	mergo.Merge(&currentPropertiesMap, newPropertiesMap, mergo.WithOverride)
	updatedPropertiesBytes, err := json.Marshal(currentPropertiesMap)
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
		log.WithFields(log.Fields{"userProperties": &userProperties, "error": err}).Error("createUserProperties Failed")
		return "", http.StatusInternalServerError
	}
	return userProperties.ID, http.StatusCreated
}

func getUserProperties(projectId uint64, userId string, id string) (*postgres.Jsonb, int) {
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
		log.WithFields(log.Fields{"clientIP": clientIP, "error": err}).Error(
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
