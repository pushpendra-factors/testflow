package model

import (
	C "factors/config"
	U "factors/util"
	"fmt"
	"net"
	"net/http"
	"time"

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
}

func createUserProperties(projectId uint64, userId string, properties postgres.Jsonb) (string, int) {
	db := C.GetServices().Db
	userProperties := UserProperties{
		UserId:     userId,
		ProjectId:  projectId,
		Properties: properties,
	}

	log.WithFields(log.Fields{"UserProperties": &userProperties}).Info("Creating user properties")

	if err := db.Create(&userProperties).Error; err != nil {
		log.WithFields(log.Fields{"userProperties": &userProperties, "error": err}).Error("createUserProperties Failed")
		return "", http.StatusInternalServerError
	}
	return userProperties.ID, DB_SUCCESS
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
	return &userProperties.Properties, DB_SUCCESS
}

func FillUserDefaultProperties(properties *U.PropertiesMap, clientIP string) error {
	geo := C.GetServices().GeoLocation

	// ClientIP unavailable.
	if clientIP == "" {
		return fmt.Errorf("invalid IP, failed adding geolocation properties")
	}

	// Added IP for internal usage.
	(*properties)[U.UP_INTERNAL_IP] = clientIP

	country, err := geo.Country(net.ParseIP(clientIP))
	if err != nil {
		log.WithFields(log.Fields{"clientIP": clientIP, "serviceError": err}).Error(
			"Failed to get country information from geodb")
		return err
	}

	(*properties)[U.UP_COUNTRY] = country.Country.IsoCode

	return nil
}
