package model

import (
	"errors"
	C "factors/config"
	U "factors/util"
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
	AMPUserId          string         `json:"amp_user_id"`
	// UserId provided by the customer.
	// An unique index is creatd on ProjectId+UserId.
	CustomerUserId string `gorm:"type:varchar(255);default:null" json:"c_uid"`
	// unix epoch timestamp in seconds.
	JoinTimestamp int64     `json:"join_timestamp"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

const usersLimitForProperties = 1000

const errorDuplicateAMPUser = "pq: duplicate key value violates unique constraint \"users_project_id_amp_user_idx\""

func isDuplicateAMPUserError(err error) bool {
	return err.Error() == errorDuplicateAMPUser
}

func (user *User) BeforeCreate(scope *gorm.Scope) error {
	// Increamenting count based on EventNameId, not by EventName.
	if user.JoinTimestamp <= 0 {
		// Default to 60 seconds earlier than now, so that if event is also created simultaneously
		// user join is earlier.
		user.JoinTimestamp = time.Now().Unix() - 60
	}

	// adds join timestamp to user properties.
	newProperties := map[string]interface{}{
		U.UP_JOIN_TIME: user.JoinTimestamp,
	}
	newPropertiesJsonb, err := U.AddToPostgresJsonb(&user.Properties, newProperties, true)
	if err != nil {
		return err
	}
	user.Properties = *newPropertiesJsonb

	return nil
}

// createUserWithError - Returns error during create to match
// with constraint errors.
func createUserWithError(user *User) (*User, error) {
	logCtx := log.WithField("project_id", user.ProjectId)

	if user.ProjectId == 0 {
		logCtx.Error("Failed to create user. ProjectId not provided.")
		return nil, errors.New("invalid project_id")
	}

	// Add id with our uuid generator, if not given.
	if user.ID == "" {
		user.ID = U.GetUUID()
	}

	db := C.GetServices().Db
	if err := db.Create(user).Error; err != nil {
		return nil, err
	}

	propertiesId, errCode := UpdateUserPropertiesByCurrentProperties(user.ProjectId, user.ID,
		user.PropertiesId, &user.Properties, user.JoinTimestamp)
	if errCode == http.StatusInternalServerError {
		return nil, errors.New("failed to update user properties")
	}

	// assign propertiesId with created propertiesId.
	user.PropertiesId = propertiesId

	return user, nil
}

func CreateUser(user *User) (*User, int) {
	logCtx := log.WithField("project_id", user.ProjectId)

	user, err := createUserWithError(user)
	if err != nil {
		if U.IsPostgresIntegrityViolationError(err) {
			return nil, http.StatusNotAcceptable
		}

		logCtx.WithError(err).Error("Failed to create user")
		return nil, http.StatusInternalServerError
	}

	return user, http.StatusCreated
}

// UpdateUser updates user fields by Id.
func UpdateUser(projectId uint64, id string, user *User, updateTimestamp int64) (*User, int) {

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
	db := C.GetServices().Db
	if err := db.Model(&updatedUser).Where("project_id = ?", projectId).Where("id = ?",
		cleanId).Updates(user).Error; err != nil {

		log.WithFields(log.Fields{"user": user}).WithError(err).Error("Failed updating fields by user_id")
		return nil, http.StatusInternalServerError
	}

	_, errCode := UpdateUserProperties(projectId, id, &user.Properties, updateTimestamp)
	if errCode != http.StatusAccepted && errCode != http.StatusNotModified {
		return nil, http.StatusInternalServerError
	}

	return &updatedUser, http.StatusAccepted
}

// UpdateUserProperties only if there is a change in properties values.
func UpdateUserProperties(projectId uint64, id string,
	properties *postgres.Jsonb, updateTimestamp int64) (string, int) {

	currentPropertiesId, status := getUserPropertiesId(projectId, id)
	if status != http.StatusFound {
		return "", status
	}

	return UpdateUserPropertiesByCurrentProperties(projectId, id,
		currentPropertiesId, properties, updateTimestamp)
}

// UpdateUser
func UpdateUserPropertiesByCurrentProperties(projectId uint64, id string,
	currentPropertiesId string, properties *postgres.Jsonb, updateTimestamp int64) (string, int) {

	if updateTimestamp == 0 {
		return "", http.StatusBadRequest
	}

	// Update properties.
	newPropertiesId, statusCode := createUserPropertiesIfChanged(
		projectId, id, currentPropertiesId, properties, updateTimestamp)
	if statusCode != http.StatusCreated && statusCode != http.StatusNotModified {
		return "", http.StatusInternalServerError
	}

	if newPropertiesId == currentPropertiesId {
		return currentPropertiesId, http.StatusNotModified
	}

	db := C.GetServices().Db
	if err := db.Model(&User{}).Where("project_id = ?", projectId).Where("id = ?",
		id).Update("properties_id", newPropertiesId).Error; err != nil {

		log.WithFields(log.Fields{"projectId": projectId,
			"id": id}).WithError(err).Error("Failed updating propertyId")
		return "", http.StatusInternalServerError
	}

	return newPropertiesId, http.StatusAccepted
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
	logCtx := log.WithFields(log.Fields{"project_id": projectId, "user_id": id})

	var user User
	if err := db.Where("project_id = ?", projectId).Where("id = ?", id).First(&user).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}
		logCtx.WithError(err).Error("Failed to get user using user_id")
		return nil, http.StatusInternalServerError
	}

	if user.PropertiesId != "" {
		properties, errCode := GetUserProperties(projectId, id, user.PropertiesId)
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

// CreateOrGetUser create or updates(c_uid) and returns user customer_user_id.
func CreateOrGetUser(projectId uint64, custUserId string) (*User, int) {
	logCtx := log.WithFields(log.Fields{"project_id": projectId,
		"provided_c_uid": custUserId})
	if custUserId == "" {
		logCtx.Error("No customer user id given")
		return nil, http.StatusBadRequest
	}

	var user *User
	var errCode int
	user, errCode = GetUserLatestByCustomerUserId(projectId, custUserId)
	if errCode == http.StatusFound {
		return user, http.StatusOK
	}

	if errCode == http.StatusInternalServerError {
		logCtx.WithField("err_code", errCode).Error(
			"Failed to fetching user with provided c_uid.")
		return nil, errCode
	}

	cUser := &User{ProjectId: projectId, CustomerUserId: custUserId}

	user, errCode = CreateUser(cUser)
	if errCode != http.StatusCreated {
		logCtx.WithField("err_code", errCode).Error(
			"Failed creating user with c_uid. get_segment_user failed.")
		return nil, errCode
	}
	return user, errCode
}

// CreateOrGetSegmentUser create or updates(c_uid) and returns user by segement_anonymous_id
// and/or customer_user_id.
func CreateOrGetSegmentUser(projectId uint64, segAnonId, custUserId string, segReqTimestamp int64) (*User, int) {
	logCtx := log.WithFields(log.Fields{"project_id": projectId, "seg_aid": segAnonId,
		"provided_c_uid": custUserId})

	// seg_aid not provided.
	if segAnonId == "" && custUserId == "" {
		logCtx.Error("No segment user id or customer user id given")
		return nil, http.StatusBadRequest
	}

	var user *User
	var errCode int
	// fetch user by seg_aid, if given.
	if segAnonId != "" {
		user, errCode = GetUserBySegmentAnonymousId(projectId, segAnonId)
		if errCode == http.StatusInternalServerError ||
			errCode == http.StatusBadRequest {
			return nil, errCode
		}
	} else {
		errCode = http.StatusNotFound
	}

	// fetch by c_uid, if user not found by seg_aid provided or c_uid provided.
	if errCode == http.StatusNotFound {
		// if found by c_uid return user, else create new user.
		if custUserId != "" {
			user, errCode = GetUserLatestByCustomerUserId(projectId, custUserId)
			if errCode == http.StatusFound {
				return user, http.StatusOK
			}

			if errCode == http.StatusInternalServerError {
				logCtx.WithField("err_code", errCode).Error(
					"Failed to fetching user with segment provided c_uid.")
				return nil, errCode
			}
		}

		cUser := &User{ProjectId: projectId}

		// add seg_aid, if provided and not exist already.
		if segAnonId != "" {
			cUser.SegmentAnonymousId = segAnonId
		}

		// add c_uid on create, if provided and not exist already.
		if custUserId != "" {
			cUser.CustomerUserId = custUserId
		}

		user, errCode = CreateUser(cUser)
		if errCode != http.StatusCreated {
			logCtx.WithField("err_code", errCode).Error(
				"Failed creating user with c_uid. get_segment_user failed.")
			return nil, errCode
		}
		return user, errCode
	}

	// No c_uid provided given, to update.
	if custUserId == "" {
		return user, http.StatusOK
	}

	logCtx = logCtx.WithField("fetched_c_uid", user.CustomerUserId)

	// fetched c_uid empty, identify and return.
	if user.CustomerUserId == "" {
		uUser, uErrCode := UpdateUser(projectId, user.ID, &User{CustomerUserId: custUserId}, segReqTimestamp)
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

func getUserByAMPUserId(projectId uint64, ampUserId string) (*User, int) {
	logCtx := log.WithField("project_id", projectId).WithField(
		"amp_user_id", ampUserId)

	var users []User

	db := C.GetServices().Db
	err := db.Limit(1).Where("project_id = ? AND amp_user_id = ?",
		projectId, ampUserId).Find(&users).Error
	if err != nil {
		logCtx.Error("Failed to get user by amp_user_id")
		return nil, http.StatusInternalServerError
	}

	if len(users) == 0 {
		return nil, http.StatusNotFound
	}

	return &users[0], http.StatusFound
}

func CreateOrGetAMPUser(projectId uint64, ampUserId string, timestamp int64) (*User, int) {
	if projectId == 0 || ampUserId == "" {
		return nil, http.StatusBadRequest
	}

	logCtx := log.WithField("project_id",
		projectId).WithField("amp_user_id", ampUserId)

	user, errCode := getUserByAMPUserId(projectId, ampUserId)
	if errCode == http.StatusInternalServerError {
		return nil, errCode
	}

	if errCode == http.StatusFound {
		return user, errCode
	}

	user, err := createUserWithError(&User{ProjectId: projectId,
		AMPUserId: ampUserId, JoinTimestamp: timestamp})
	if err != nil {
		// Get and return error is duplicate error.
		if isDuplicateAMPUserError(err) {
			user, errCode := getUserByAMPUserId(projectId, ampUserId)
			if errCode != http.StatusFound {
				return nil, errCode
			}

			return user, http.StatusFound
		}

		logCtx.WithError(err).Error(
			"Failed to create user by amp user id on CreateOrGetAMPUser")
		return nil, http.StatusInternalServerError
	}

	return user, http.StatusCreated
}

func GetRecentUserPropertyKeysWithLimits(projectId uint64, usersLimit int) (map[string][]string, int) {
	db := C.GetServices().Db

	logCtx := log.WithField("project_id", projectId)

	queryStr := "WITH recent_users AS (SELECT properties_id FROM users WHERE project_id = ? ORDER BY created_at DESC LIMIT ?)" +
		" " + "SELECT user_properties.properties FROM recent_users LEFT JOIN user_properties ON recent_users.properties_id = user_properties.id" +
		" " + "WHERE user_properties.project_id = ? AND user_properties.properties != 'null';"

	rows, err := db.Raw(queryStr, projectId, usersLimit, projectId).Rows()
	if err != nil {
		logCtx.WithError(err).Error("Failed to get recent user property keys.")
		return nil, http.StatusInternalServerError
	}
	defer rows.Close()

	propertiesMap := make(map[string]map[interface{}]bool, 0)
	for rows.Next() {
		var propertiesJson []byte
		rows.Scan(&propertiesJson)

		err := U.FillPropertyKvsFromPropertiesJson(propertiesJson, &propertiesMap, U.SamplePropertyValuesLimit)
		if err != nil {
			log.WithError(err).WithField("properties_json",
				string(propertiesJson)).Error("Failed to unmarshal json properties.")
			return nil, http.StatusInternalServerError
		}
	}

	err = rows.Err()
	if err != nil {
		logCtx.WithError(err).Error("Failed to scan recent property keys.")
		return nil, http.StatusInternalServerError
	}

	propsByType, err := U.ClassifyPropertiesType(&propertiesMap)
	if err != nil {
		logCtx.WithError(err).Error("Failed to classify properties on get recent property keys.")
		return nil, http.StatusInternalServerError
	}

	return propsByType, http.StatusFound
}

func GetRecentUserPropertyKeys(projectId uint64) (map[string][]string, int) {
	return GetRecentUserPropertyKeysWithLimits(projectId, usersLimitForProperties)
}

func GetRecentUserPropertyValuesWithLimits(projectId uint64, propertyKey string, usersLimit, valuesLimit int) ([]string, int) {
	db := C.GetServices().Db

	// limit on values returned.
	values := make([]string, 0, 0)
	queryStmnt := "WITH recent_users AS (SELECT id FROM users WHERE project_id = ? ORDER BY created_at DESC limit ?)" +
		" " + "SELECT DISTINCT(user_properties.properties->?) AS values FROM recent_users" +
		" " + "LEFT JOIN user_properties ON recent_users.id = user_properties.user_id WHERE user_properties.project_id = ?" +
		" " + "AND user_properties.properties != 'null' AND user_properties.properties->? IS NOT NULL limit ?"

	queryParams := make([]interface{}, 0, 0)
	queryParams = append(queryParams, projectId, usersLimit, propertyKey, projectId, propertyKey, valuesLimit)

	logCtx := log.WithFields(log.Fields{"project_id": projectId, "property_key": propertyKey, "values_limit": valuesLimit})

	rows, err := db.Raw(queryStmnt, queryParams...).Rows()
	if err != nil {
		logCtx.WithError(err).Error("Failed to get property values.")
		return values, http.StatusInternalServerError
	}
	defer rows.Close()

	for rows.Next() {
		var value string
		rows.Scan(&value)
		value = U.TrimQuotes(value)
		values = append(values, value)
	}

	err = rows.Err()
	if err != nil {
		logCtx.WithError(err).Error("Failed scanning rows on get property values.")
		return values, http.StatusInternalServerError
	}

	return values, http.StatusFound
}

func GetRecentUserPropertyValues(projectId uint64, propertyKey string) ([]string, int) {
	return GetRecentUserPropertyValuesWithLimits(projectId, propertyKey, usersLimitForProperties, 2000)
}

// Updates user join time with min join time among all users with same customer user id.
func UpdateUserJoinTimePropertyForCustomerUser(projectId uint64, customerUserId string) int {
	db := C.GetServices().Db

	if projectId == 0 || customerUserId == "" {
		return http.StatusBadRequest
	}

	var users []User
	if err := db.Order("join_timestamp ASC").Where("project_id = ? AND customer_user_id = ?",
		projectId, customerUserId).Find(&users).Error; err != nil {

		return http.StatusInternalServerError
	}

	if len(users) == 0 {
		return http.StatusNotFound
	}

	if len(users) == 1 {
		return http.StatusNotModified
	}

	// sorted result from DB by joinTimestamp by ASC.
	//minJoinTimestamp := users[0].JoinTimestamp

	/*for _, user := range users {
		errCode := UpdatePropertyOnAllUserPropertyRecords(projectId, user.ID, U.UP_JOIN_TIME, minJoinTimestamp)
		if errCode == http.StatusInternalServerError {
			// log failure and continue with next user.
			log.WithFields(log.Fields{"project_id": projectId, "user_id": user.ID}).Error("Failed to update user join time by customer user id.")
		}
	}*/

	return http.StatusAccepted
}

func GetUserPropertiesAsMap(projectId uint64, id string) (*map[string]interface{}, int) {
	logCtx := log.WithField("project_id", projectId).WithField("id", id)

	user, errCode := GetUser(projectId, id)
	if errCode != http.StatusFound {
		logCtx.WithField("err_code", errCode).Error(
			"Getting user failed on get user properties as map.")
		return nil, errCode
	}

	existingUserProperties, err := U.DecodePostgresJsonb(&user.Properties)
	if err != nil {
		logCtx.WithError(err).Error(
			"Unmarshaling user properties failed on get user properties as map.")
		return nil, http.StatusInternalServerError
	}

	return existingUserProperties, http.StatusFound
}
