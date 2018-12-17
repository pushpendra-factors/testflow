package model

import (
	C "factors/config"
	"net/http"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

type User struct {
	// Composite primary key with project_id and random uuid.
	ID string `gorm:"primary_key:true;uuid;default:uuid_generate_v4()" json:"id"`
	// Below are the foreign key constraints added in creation script.
	// project_id -> projects(id)
	ProjectId    uint64 `gorm:"primary_key:true;" json:"project_id"`
	PropertiesId string `json:"properties_id"`
	// Not part of table, but part of json. Stored in UserProperties table.
	Properties postgres.Jsonb `gorm:"-" json:"properties"`

	// UserId provided by the customer.
	// An unique index is creatd on ProjectId+UserId.
	CustomerUserId string    `json:"c_uid"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func CreateUser(user *User) (*User, int) {
	db := C.GetServices().Db

	log.WithFields(log.Fields{"user": &user}).Info("Creating user")

	// Input Validation. (ID is to be auto generated).
	if user.ID != "" {
		log.Error("CreateUser Failed. ID provided.")
		return nil, http.StatusBadRequest
	}
	if user.ProjectId == 0 {
		log.Error("CreateUser Failed. ProjectId not provided.")
		return nil, http.StatusBadRequest
	}

	if err := db.Create(user).Error; err != nil {
		log.WithFields(log.Fields{"user": &user, "error": err}).Error("CreateUser Failed")
		return nil, http.StatusInternalServerError
	}
	propertiesId, success := createUserProperties(user.ProjectId, user.ID, user.Properties)
	if success != http.StatusCreated {
		return nil, http.StatusInternalServerError
	}

	if err := db.Model(&user).Update("properties_id", propertiesId).Error; err != nil {
		log.WithFields(log.Fields{"user": user, "error": err}).Error("Failed updating propertyId")
		return nil, http.StatusInternalServerError
	}

	return user, http.StatusCreated
}

func GetUser(projectId uint64, id string) (*User, int) {
	db := C.GetServices().Db

	var user User
	if err := db.Where("project_id = ?", projectId).Where("id = ?", id).First(&user).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}
		return nil, http.StatusInternalServerError
	}
	if user.PropertiesId != "" {
		properties, errCode := getUserProperties(projectId, id, user.PropertiesId)
		if errCode != http.StatusFound {
			return nil, errCode
		}
		user.Properties = *properties
	}

	return &user, http.StatusFound
}

func GetUsers(projectId uint64, offset uint64, limit uint64) ([]User, int) {
	db := C.GetServices().Db

	var users []User
	if err := db.Order("created_at").Offset(offset).Where("project_id = ?", projectId).Limit(limit).Find(&users).Error; err != nil {
		return nil, http.StatusInternalServerError
	}
	if len(users) == 0 {
		return nil, http.StatusNotFound
	}
	return users, http.StatusFound
}

func GetUserLatestByCustomerUserId(projectId uint64, customerUserId string) (*User, int) {
	db := C.GetServices().Db

	var user User
	err := db.Where("project_id = ?", projectId).Where("customer_user_id = ?", customerUserId).Last(&user).Error

	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}
		return nil, http.StatusInternalServerError
	}
	return &user, http.StatusFound
}

// UpdateUser updates user fields by Id.
func UpdateUser(projectId uint64, id string, user *User) (*User, int) {
	db := C.GetServices().Db

	// Todo(Dinesh): Move to validations.
	// Ref: https://github.com/qor/validations
	if projectId == 0 {
		return nil, http.StatusBadRequest
	}

	// Todo(Dinesh): Move to validations.
	cleanId := strings.TrimSpace(id)
	if len(cleanId) == 0 {
		return nil, http.StatusBadRequest
	}

	if user.ProjectId != 0 || user.ID != "" {
		log.WithFields(log.Fields{"user": user}).Error("Bad Request. Tried updating ID or ProjectId.")
		return nil, http.StatusBadRequest
	}

	var updatedUser User
	if err := db.Model(&updatedUser).Where("project_id = ?", projectId).Where("id = ?", cleanId).Updates(user).Error; err != nil {
		log.WithFields(log.Fields{"user": user, "error": err}).Error("Failed updating fields by user_id")
		return nil, http.StatusInternalServerError
	}
	// Update properties
	propertiesId, success := createUserProperties(projectId, id, user.Properties)
	if success != http.StatusCreated {
		return nil, http.StatusInternalServerError
	}
	if err := db.Model(&updatedUser).Update("properties_id", propertiesId).Error; err != nil {
		log.WithFields(log.Fields{"user": user, "error": err}).Error("Failed updating propertyId")
		return nil, http.StatusInternalServerError
	}
	return &updatedUser, http.StatusAccepted
}
