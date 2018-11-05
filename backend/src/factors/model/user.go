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
	ProjectId uint64 `gorm:"primary_key:true;" json:"project_id"`
	// UserId provided by the customer.
	// An unique index is creatd on ProjectId+UserId.
	CustomerUserId string `json:"c_uid"`

	// JsonB of postgres with gorm. https://github.com/jinzhu/gorm/issues/1183
	Properties postgres.Jsonb `json:"properties,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

func CreateUser(user *User) (*User, int) {
	db := C.GetServices().Db

	log.WithFields(log.Fields{"user": &user}).Info("Creating user")

	// Input Validation. (ID is to be auto generated).
	if user.ID != "" {
		log.Error("CreateUser Failed. ID provided.")
		return nil, http.StatusBadRequest
	}

	if err := db.Create(user).Error; err != nil {
		log.WithFields(log.Fields{"user": &user, "error": err}).Error("CreateUser Failed")
		return nil, http.StatusInternalServerError
	} else {
		return user, DB_SUCCESS
	}
}

func GetUser(projectId uint64, id string) (*User, int) {
	db := C.GetServices().Db

	var user User
	if err := db.Where("project_id = ?", projectId).Where("id = ?", id).First(&user).Error; err != nil {
		return nil, 404
	} else {
		return &user, DB_SUCCESS
	}
}

func GetUsers(projectId uint64, offset uint64, limit uint64) ([]User, int) {
	db := C.GetServices().Db

	var users []User
	if err := db.Order("created_at").Offset(offset).Where("project_id = ?", projectId).Limit(limit).Find(&users).Error; err != nil {
		return nil, 404
	} else {
		return users, DB_SUCCESS
	}
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

	return &user, DB_SUCCESS
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
	} else {
		return &updatedUser, DB_SUCCESS
	}
}

// Todo(Dinesh): Remove this method. Use UpdateUser to update any field by id.
func UpdateCustomerUserIdById(projectId uint64, id string, customerUserId string) (*User, int) {
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

	var user User
	if err := db.Model(&user).Where("project_id = ?", projectId).Where("id = ?", cleanId).Update("customer_user_id", customerUserId).Error; err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Failed updating customer_user_id by user_id")
		return nil, http.StatusInternalServerError
	} else {
		return &user, DB_SUCCESS
	}
}
