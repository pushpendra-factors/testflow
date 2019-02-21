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
	Properties         postgres.Jsonb `gorm:"-" json:"properties"`
	SegmentAnonymousId string         `gorm:"type:varchar(200);default:null" json:"seg_aid"`
	// UserId provided by the customer.
	// An unique index is creatd on ProjectId+UserId.
	CustomerUserId string `gorm:"type:varchar(255);default:null" json:"c_uid"`
	// unix epoch timestamp in seconds.
	JoinTimestamp int64     `json:"join_timestamp"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (user *User) BeforeCreate(scope *gorm.Scope) error {
	// Increamenting count based on EventNameId, not by EventName.
	if user.JoinTimestamp <= 0 {
		user.JoinTimestamp = time.Now().Unix()
	}
	return nil
}

func CreateUser(user *User) (*User, int) {
	db := C.GetServices().Db

	log.WithFields(log.Fields{"user": &user}).Info("Creating user")

	// Input Validation. (ID is to be auto generated).
	if user.ID != "" {
		log.Error("CreateUser Failed. ID not provided.")
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
	propertiesId, success := createUserPropertiesIfChanged(
		user.ProjectId, user.ID, user.PropertiesId, &user.Properties)
	if success != http.StatusCreated {
		return nil, http.StatusInternalServerError
	}

	if err := db.Model(&user).Update("properties_id", propertiesId).Error; err != nil {
		log.WithFields(log.Fields{"user": user, "error": err}).Error("Failed updating propertyId")
		return nil, http.StatusInternalServerError
	}

	return user, http.StatusCreated
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
	// Update properties.
	propertiesId, success := createUserPropertiesIfChanged(
		projectId, id, user.PropertiesId, &user.Properties)
	if success != http.StatusCreated && success != http.StatusNotModified {
		return nil, http.StatusInternalServerError
	}
	if err := db.Model(&updatedUser).Update("properties_id", propertiesId).Error; err != nil {
		log.WithFields(log.Fields{"user": user, "error": err}).Error("Failed updating propertyId")
		return nil, http.StatusInternalServerError
	}
	return &updatedUser, http.StatusAccepted
}

// UpdateUserProperties only if there is a change in properties values.
func UpdateUserProperties(projectId uint64, id string, properties *postgres.Jsonb) (string, int) {
	userPropertiesId, status := getUserPropertiesId(projectId, id)
	if status != http.StatusFound {
		return "", status
	}
	db := C.GetServices().Db
	// Update properties.
	newPropertiesId, statusCode := createUserPropertiesIfChanged(
		projectId, id, userPropertiesId, properties)
	if statusCode != http.StatusCreated && statusCode != http.StatusNotModified {
		return "", http.StatusInternalServerError
	}
	if newPropertiesId != userPropertiesId {
		user := User{ID: id, ProjectId: projectId}
		if err := db.Model(&user).Update("properties_id", newPropertiesId).Error; err != nil {
			log.WithFields(log.Fields{"user": user, "error": err}).Error("Failed updating propertyId")
			return "", http.StatusInternalServerError
		}
		return newPropertiesId, http.StatusAccepted
	}
	return userPropertiesId, http.StatusNotModified
}

func getUserPropertiesId(projectId uint64, id string) (string, int) {
	db := C.GetServices().Db

	var user User
	if err := db.Select("properties_id").Where("project_id = ?", projectId).Where("id = ?", id).First(&user).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return "", http.StatusNotFound
		}
		return "", http.StatusInternalServerError
	}

	return user.PropertiesId, http.StatusFound
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
	if err := db.Order("created_at DESC").Where("project_id = ?", projectId).Where(
		"customer_user_id = ?", customerUserId).First(&user).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}
		return nil, http.StatusInternalServerError
	}
	return &user, http.StatusFound
}

func GetUserBySegmentAnonymousId(projectId uint64, segAnonId string) (*User, int) {
	db := C.GetServices().Db

	var user User
	if err := db.Where("project_id = ?", projectId).Where(
		"segment_anonymous_id = ?", segAnonId).First(&user).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}
		return nil, http.StatusInternalServerError
	}

	return &user, http.StatusFound
}

// GetSegmentUser create or updates(c_uid) and returns user by segement_anonymous_id
// and customer_user_id.
func GetSegmentUser(projectId uint64, segAnonId, custUserId string) (*User, int) {
	// seg_aid not provided.
	if segAnonId == "" {
		log.WithFields(log.Fields{"project_id": projectId,
			"c_uid": custUserId}).Error("No segment user id given")
		return nil, http.StatusBadRequest
	}

	logCtx := log.WithFields(log.Fields{"project_id": projectId, "seg_aid": segAnonId})

	// fetch by seg_aid, create and return user if not exist.
	user, errCode := GetUserBySegmentAnonymousId(projectId, segAnonId)
	if errCode == http.StatusInternalServerError ||
		errCode == http.StatusBadRequest {
		return nil, errCode
	}

	if errCode == http.StatusNotFound {
		cUser := &User{ProjectId: projectId, SegmentAnonymousId: segAnonId}
		// add c_uid on create, if provided.
		if custUserId != "" {
			cUser.CustomerUserId = custUserId
			logCtx = logCtx.WithField("provided_c_uid", custUserId)
		}

		user, errCode = CreateUser(cUser)
		if errCode != http.StatusCreated {
			logCtx.WithField("err_code", errCode).Error("Failed creating user with c_uid. get_segment_user failed.")
			return nil, errCode
		}
		return user, http.StatusOK
	}

	logCtx = logCtx.WithField("provided_c_uid", custUserId).WithField("fetched_c_uid", user.CustomerUserId)

	// fetched c_uid empty, identify and return.
	if user.CustomerUserId == "" {
		uUser, uErrCode := UpdateUser(projectId, user.ID, &User{CustomerUserId: custUserId})
		if uErrCode != http.StatusAccepted {
			logCtx.WithField("err_code", uErrCode).Error(
				"Identify failed. Failed updating c_uid failed. get_segment_user failed.")
			return nil, uErrCode
		}
		user.CustomerUserId = uUser.CustomerUserId
	}
	// same seg_aid with different c_uid. log error. return user.
	if user.CustomerUserId != custUserId {
		logCtx.Error("Tried re-identifying with same seg_aid and different c_uid.")
	}

	// provided and fetched c_uid are same.
	return user, http.StatusOK
}
