package model

import (
	"encoding/json"
	"errors"
	cacheRedis "factors/cache/redis"
	C "factors/config"
	U "factors/util"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
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
	AMPUserId          string         `gorm:"default:null";json:"amp_user_id"`
	// UserId provided by the customer.
	// An unique index is creatd on ProjectId+UserId.
	CustomerUserId string `gorm:"type:varchar(255);default:null" json:"c_uid"`
	// unix epoch timestamp in seconds.
	JoinTimestamp int64     `json:"join_timestamp"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

const usersLimitForProperties = 50000
const uniqueIndexProjectIdAmpUserId = "users_project_id_amp_user_idx"
const uniqueIndexProjectIdSegmentAnonymousId = "users_project_id_segment_anonymous_uidx"

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
	logCtx := log.WithField("project_id", user.ProjectId).
		WithField("user_id", user.ID)

	newUser, err := createUserWithError(user)
	if err == nil {
		return newUser, http.StatusCreated
	}

	if U.IsPostgresIntegrityViolationError(err) {
		if user.ID != "" {
			// Multiple requests trying to create user at the
			// same time should not lead failure permanently,
			// so get the user and return.
			existingUser, errCode := GetUser(user.ProjectId, user.ID)
			if errCode == http.StatusFound {
				// Using StatusCreated for consistency.
				return existingUser, http.StatusCreated
			}

			// Returned err_codes will be retried on queue.
			return nil, errCode
		}

		logCtx.WithError(err).Error("Failed to create user. Integrity violation.")
		return nil, http.StatusNotAcceptable
	}

	logCtx.WithError(err).Error("Failed to create user.")
	return nil, http.StatusInternalServerError
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

	properties = U.SanitizePropertiesJsonb(properties)

	// Update properties.
	newPropertiesId, statusCode := createUserPropertiesIfChanged(
		projectId, id, currentPropertiesId, properties, updateTimestamp)

	if statusCode == http.StatusBadRequest {
		return currentPropertiesId, http.StatusBadRequest
	}

	if statusCode != http.StatusCreated && statusCode != http.StatusNotModified {
		return currentPropertiesId, http.StatusInternalServerError
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

// GetUsersByCustomerUserID Gets all the users indentified by given customer_user_id in increasing order of updated_at.
func GetUsersByCustomerUserID(projectID uint64, customerUserID string) ([]User, int) {
	db := C.GetServices().Db
	logCtx := log.WithFields(log.Fields{
		"ProjectID":      projectID,
		"CustomerUserID": customerUserID,
	})

	var users []User
	if err := db.Order("created_at ASC").Where("project_id = ? AND customer_user_id = ?", projectID, customerUserID).Find(&users).Error; err != nil {
		logCtx.WithError(err).Error("Failed to get users for customer_user_id")
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

	var users []User
	if err := db.Limit(1).Where("project_id = ?", projectId).Where(
		"segment_anonymous_id = ?", segAnonId).Find(&users).Error; err != nil {
		log.WithField("project_id", projectId).WithField(
			"segment_anonymous_id", segAnonId).Error(
			"Failed to get user by segment_anonymous_id.")
		return nil, http.StatusInternalServerError
	}

	if len(users) == 0 {
		return nil, http.StatusNotFound
	}

	return &users[0], http.StatusFound
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

		cUser := &User{ProjectId: projectId, JoinTimestamp: segReqTimestamp}

		// add seg_aid, if provided and not exist already.
		if segAnonId != "" {
			cUser.SegmentAnonymousId = segAnonId
		}

		// add c_uid on create, if provided and not exist already.
		if custUserId != "" {
			cUser.CustomerUserId = custUserId
		}

		user, err := createUserWithError(cUser)
		if err != nil {
			// Get and return error is duplicate error.
			if U.IsPostgresUniqueIndexViolationError(
				uniqueIndexProjectIdSegmentAnonymousId, err) {

				user, errCode := GetUserBySegmentAnonymousId(projectId, segAnonId)
				if errCode != http.StatusFound {
					return nil, errCode
				}

				return user, http.StatusFound
			}

			logCtx.WithError(err).Error(
				"Failed to create user by segment anonymous id on CreateOrGetSegmentUser")
			return nil, http.StatusInternalServerError
		}

		return user, http.StatusCreated
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
		if U.IsPostgresUniqueIndexViolationError(uniqueIndexProjectIdAmpUserId, err) {
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

func GetUsersCachedCacheKey(projectId uint64, dateKey string) (*cacheRedis.Key, error) {
	prefix := "user:list"
	return cacheRedis.NewKey(projectId, prefix, dateKey)
}

func GetUserPropertiesByProjectCacheKey(projectId uint64, dateKey string) (*cacheRedis.Key, error) {
	prefix := "user:properties"
	return cacheRedis.NewKey(projectId, prefix, dateKey)
}

func GetValuesByUserPropertyCacheKey(projectId uint64, property_name string, dateKey string) (*cacheRedis.Key, error) {
	prefix := "user:property_values"
	return cacheRedis.NewKey(projectId, prefix, fmt.Sprintf("%s:%s", property_name, dateKey))
}

//GetRecentUserPropertyKeysWithLimits This method gets all the recent 'limit' property keys from DB for a given project
func GetRecentUserPropertyKeysWithLimits(projectID uint64, usersLimit int, propertyLimit int) ([]U.Property, error) {
	db := C.GetServices().Db

	logCtx := log.WithField("project_id", projectID)
	properties := make([]U.Property, 0)

	queryStr := "WITH recent_users AS (SELECT properties_id FROM users WHERE project_id = ? ORDER BY created_at DESC LIMIT ?)" +
		" " + "SELECT json_object_keys(user_properties.properties::json) AS key, COUNT(*) AS count, MAX(updated_timestamp) as last_seen " +
		" " + "FROM recent_users LEFT JOIN user_properties ON recent_users.properties_id = user_properties.id" +
		" " + "WHERE user_properties.project_id = ? AND user_properties.properties != 'null' GROUP BY key ORDER BY count DESC LIMIT ?;"

	rows, err := db.Raw(queryStr, projectID, usersLimit, projectID, propertyLimit).Rows()

	if err != nil {
		logCtx.WithError(err).Error("Failed to get recent user property keys.")
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var property U.Property
		if err := db.ScanRows(rows, &property); err != nil {
			logCtx.WithError(err).Error("Failed scanning rows on GetRecentUserPropertyKeysWithLimits")
			return nil, err
		}
		properties = append(properties, property)
	}

	return properties, nil
}

//GetRecentUserPropertyValuesWithLimits This method gets all the recent 'limit' property values from DB for a given project/property
func GetRecentUserPropertyValuesWithLimits(projectID uint64, propertyKey string, usersLimit, valuesLimit int) ([]U.PropertyValue, string, error) {
	db := C.GetServices().Db

	// limit on values returned.
	values := make([]U.PropertyValue, 0, 0)
	queryStmnt := "WITH recent_users AS (SELECT id FROM users WHERE project_id = ? ORDER BY created_at DESC limit ?)" +
		" " + "SELECT DISTINCT(user_properties.properties->?) AS value, 1 AS count,updated_timestamp AS last_seen, jsonb_typeof(user_properties.properties->?) AS value_type FROM recent_users" +
		" " + "LEFT JOIN user_properties ON recent_users.id = user_properties.user_id WHERE user_properties.project_id = ?" +
		" " + "AND user_properties.properties != 'null' AND user_properties.properties->? IS NOT NULL limit ?"

	queryParams := make([]interface{}, 0, 0)
	queryParams = append(queryParams, projectID, usersLimit, propertyKey, propertyKey, projectID, propertyKey, valuesLimit)

	logCtx := log.WithFields(log.Fields{"project_id": projectID, "property_key": propertyKey, "values_limit": valuesLimit})

	rows, err := db.Raw(queryStmnt, queryParams...).Rows()
	if err != nil {
		logCtx.WithError(err).Error("Failed to get property values.")
		return nil, "", err
	}
	defer rows.Close()

	for rows.Next() {
		var value U.PropertyValue
		if err := db.ScanRows(rows, &value); err != nil {
			logCtx.WithError(err).Error("Failed scanning rows on GetRecentUserPropertyValuesWithLimits")
			return nil, "", err
		}
		value.Value = U.TrimQuotes(value.Value)
		values = append(values, value)
	}

	err = rows.Err()
	if err != nil {
		logCtx.WithError(err).Error("Failed scanning rows on get property values.")
		return nil, "", err
	}

	return values, U.GetCategoryType(values), nil
}

//GetUserPropertiesByProject This method iterates over n days and gets user properties from cache for a given project
// Picks all past 24 hrs seen properties and sorts the remaining by count and returns top 'limit'
func GetUserPropertiesByProject(projectID uint64, limit int, lastNDays int) (map[string][]string, error) {
	currentDate := time.Now().UTC()
	properties := make(map[string][]string)
	if projectID == 0 {
		return properties, errors.New("invalid project on GetUserPropertiesByProject")
	}
	userProperties := make([]U.CachePropertyWithTimestamp, 0)
	for i := 0; i < lastNDays; i++ {
		currentDateOnlyFormat := currentDate.AddDate(0, 0, -i).Format("2006-01-02")
		userProperty, err := getUserPropertiesByProjectFromCache(projectID, currentDateOnlyFormat)
		if err != nil {
			return nil, err
		}
		userProperties = append(userProperties, userProperty)
	}

	userPropertiesAggregated := U.AggregatePropertyAcrossDate(userProperties)
	userPropertiesSorted := U.SortByTimestampAndCount(userPropertiesAggregated)

	if limit > 0 {
		sliceLength := len(userPropertiesSorted)
		if sliceLength > limit {
			userPropertiesSorted = userPropertiesSorted[0:limit]
		}
	}

	for _, v := range userPropertiesSorted {
		if properties[v.Category] == nil {
			properties[v.Category] = make([]string, 0)
		}
		properties[v.Category] = append(properties[v.Category], v.Name)
	}

	return properties, nil
}

func getUserPropertiesByProjectFromCache(projectID uint64, dateKey string) (U.CachePropertyWithTimestamp, error) {

	if projectID == 0 {
		return U.CachePropertyWithTimestamp{}, errors.New("invalid project on GetUserPropertiesByProjectFromCache")
	}
	PropertiesKey, err := GetUserPropertiesByProjectCacheKey(projectID, dateKey)
	if err != nil {
		return U.CachePropertyWithTimestamp{}, err
	}
	userProperties, _, err := cacheRedis.GetIfExistsPersistent(PropertiesKey)
	if err != nil {
		return U.CachePropertyWithTimestamp{}, err
	}
	if userProperties == "" {
		return U.CachePropertyWithTimestamp{}, nil
	}
	var cacheValue U.CachePropertyWithTimestamp
	err = json.Unmarshal([]byte(userProperties), &cacheValue)
	if err != nil {
		return U.CachePropertyWithTimestamp{}, err
	}
	// Not adding nil/0 check for properties list since it can be null/empty
	return cacheValue, nil
}

//GetPropertyValuesByUserProperty This method iterates over n days and gets user property values from cache for a given project/property
// Picks all past 24 hrs seen values and sorts the remaining by count and returns top 'limit'
func GetPropertyValuesByUserProperty(projectID uint64, propertyName string, limit int, lastNDays int) ([]string, error) {
	currentDate := time.Now().UTC()
	if projectID == 0 {
		return []string{}, errors.New("invalid project on GetPropertyValuesByUserProperty")
	}

	if propertyName == "" {
		return []string{}, errors.New("invalid property_name on GetPropertyValuesByUserProperty")
	}
	values := make([]U.CachePropertyValueWithTimestamp, 0)
	for i := 0; i < lastNDays; i++ {
		currentDateOnlyFormat := currentDate.AddDate(0, 0, -i).Format("2006-01-02")
		value, err := getPropertyValuesByUserPropertyFromCache(projectID, propertyName, currentDateOnlyFormat)
		if err != nil {
			return []string{}, err
		}
		values = append(values, value)
	}

	valueStrings := make([]string, 0)
	valuesAggregated := U.AggregatePropertyValuesAcrossDate(values)
	valuesSorted := U.SortByTimestampAndCount(valuesAggregated)

	for _, v := range valuesSorted {
		valueStrings = append(valueStrings, v.Name)
	}
	if limit > 0 {
		sliceLength := len(valueStrings)
		if sliceLength > limit {
			return valueStrings[0:limit], nil
		}
	}
	return valueStrings, nil
}

func getPropertyValuesByUserPropertyFromCache(projectID uint64, propertyName string, dateKey string) (U.CachePropertyValueWithTimestamp, error) {

	if projectID == 0 {
		return U.CachePropertyValueWithTimestamp{}, errors.New("invalid project on GetPropertyValuesByEventPropertyFromCache")
	}

	if propertyName == "" {
		return U.CachePropertyValueWithTimestamp{}, errors.New("invalid property_name on GetPropertyValuesByEventPropertyFromCache")
	}

	eventPropertyValuesKey, err := GetValuesByUserPropertyCacheKey(projectID, propertyName, dateKey)
	if err != nil {
		return U.CachePropertyValueWithTimestamp{}, err
	}
	values, _, err := cacheRedis.GetIfExistsPersistent(eventPropertyValuesKey)
	if err != nil {
		return U.CachePropertyValueWithTimestamp{}, err
	}
	if values == "" {
		return U.CachePropertyValueWithTimestamp{}, nil
	}
	var cacheValue U.CachePropertyValueWithTimestamp
	err = json.Unmarshal([]byte(values), &cacheValue)
	if err != nil {
		return U.CachePropertyValueWithTimestamp{}, err
	}
	// Not adding nil/0 check for properties list since it can be null/empty
	return cacheValue, nil
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

// GetDistinctCustomerUserIDSForProject Returns all distinct customer_user_id for Project.
func GetDistinctCustomerUserIDSForProject(projectID uint64) ([]string, int) {
	logCtx := log.WithFields(log.Fields{"ProjectID": projectID})
	db := C.GetServices().Db

	var customerUserIDS []string
	rows, err := db.Model(&User{}).Where("project_id = ? AND customer_user_id IS NOT NULL", projectID).Select("distinct customer_user_id").Rows()
	if err != nil {
		logCtx.WithError(err).Error("Failed to get customer user ids")
		return customerUserIDS, http.StatusInternalServerError
	}

	for rows.Next() {
		var customerUserID string
		err = rows.Scan(&customerUserID)
		if err != nil {
			logCtx.WithError(err).Error("Failed to scan customer user id")
			return customerUserIDS, http.StatusInternalServerError
		}
		customerUserIDS = append(customerUserIDS, customerUserID)
	}
	return customerUserIDS, http.StatusFound
}

func getRecentUserPropertyKeysCacheKey(projectId uint64) (*cacheRedis.Key, error) {
	prefix := "recent_properties"
	suffix := "user_properites:keys"
	return cacheRedis.NewKey(projectId, prefix, suffix)
}

func getRecentUserPropertyValuesCacheKey(projectId uint64, property string) (*cacheRedis.Key, error) {
	prefix := "recent_properties"
	suffix := fmt.Sprintf("user_properties:property:%s:values", property)
	return cacheRedis.NewKey(projectId, prefix, suffix)
}

func GetCacheRecentUserPropertyKeys(projectId uint64) (map[string][]string, error) {
	return GetCacheRecentPropertyKeys(projectId, "", PropertyEntityUser)
}

func SetCacheRecentUserPropertyKeys(projectId uint64, propsByType map[string][]string) error {
	return SetCacheRecentPropertyKeys(projectId, "", propsByType, PropertyEntityUser)
}

func GetCacheRecentUserPropertyValues(projectId uint64, property string) ([]string, error) {
	return GetCacheRecentPropertyValues(projectId, "", property, PropertyEntityUser)
}

func SetCacheRecentUserPropertyValues(projectId uint64, property string, values []string) error {
	return SetCacheRecentPropertyValues(projectId, "", property, values, PropertyEntityUser)
}

func GetRecentUserPropertyKeys(projectId uint64) (map[string][]string, int) {
	return GetRecentUserPropertyKeysWithLimitsFallback(projectId, usersLimitForProperties)
}

func GetRecentUserPropertyKeysWithLimitsFallback(projectId uint64, usersLimit int) (map[string][]string, int) {
	logCtx := log.WithField("project_id", projectId)

	if properties, err := GetCacheRecentUserPropertyKeys(projectId); err == nil {
		return properties, http.StatusFound
	} else if err != redis.ErrNil {
		logCtx.WithError(err).Error("Failed to get GetCacheRecentPropertyKeys.")
	}

	usersAfterTimestamp := U.UnixTimeBeforeDuration(24 * time.Hour)
	logCtx = log.WithFields(log.Fields{"project_id": projectId, "users_after_timestamp": usersAfterTimestamp})

	db := C.GetServices().Db

	queryStr := "WITH recent_users AS (SELECT properties_id FROM users WHERE project_id = ? AND join_timestamp >= ? ORDER BY created_at DESC LIMIT ?)" +
		" " + "SELECT user_properties.properties FROM recent_users LEFT JOIN user_properties ON recent_users.properties_id = user_properties.id" +
		" " + "WHERE user_properties.project_id = ? AND user_properties.properties != 'null';"

	rows, err := db.Raw(queryStr, projectId, usersAfterTimestamp, usersLimit, projectId).Rows()
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

	if err = SetCacheRecentUserPropertyKeys(projectId, propsByType); err != nil {
		logCtx.WithError(err).Error("Failed to SetCacheRecentUserPropertyKeys.")
	}

	return propsByType, http.StatusFound
}

func GetRecentUserPropertyValuesWithLimitsFallback(projectId uint64, propertyKey string, usersLimit, valuesLimit int) ([]string, int) {
	logCtx := log.WithFields(log.Fields{"project_id": projectId, "property_key": propertyKey, "values_limit": valuesLimit})

	if values, err := GetCacheRecentUserPropertyValues(projectId, propertyKey); err == nil {
		return values, http.StatusFound
	} else if err != redis.ErrNil {
		logCtx.WithError(err).Error("Failed to get GetCacheRecentPropertyValues.")
	}

	db := C.GetServices().Db

	// limit on values returned.
	values := make([]string, 0, 0)
	queryStmnt := "WITH recent_users AS (SELECT id FROM users WHERE project_id = ? ORDER BY created_at DESC limit ?)" +
		" " + "SELECT DISTINCT(user_properties.properties->?) AS values FROM recent_users" +
		" " + "LEFT JOIN user_properties ON recent_users.id = user_properties.user_id WHERE user_properties.project_id = ?" +
		" " + "AND user_properties.properties != 'null' AND user_properties.properties->? IS NOT NULL limit ?"

	queryParams := make([]interface{}, 0, 0)
	queryParams = append(queryParams, projectId, usersLimit, propertyKey, projectId, propertyKey, valuesLimit)

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

	if err = SetCacheRecentUserPropertyValues(projectId, propertyKey, values); err != nil {
		logCtx.WithError(err).Error("Failed to SetCacheRecentUserPropertyValues.")
	}

	return values, http.StatusFound
}

func GetRecentUserPropertyValues(projectId uint64, propertyKey string) ([]string, int) {
	return GetRecentUserPropertyValuesWithLimitsFallback(projectId, propertyKey, usersLimitForProperties, 2000)
}
