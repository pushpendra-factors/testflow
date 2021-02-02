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

func GetIdentifiedUserPropertiesAsJsonb(customerUserId string) (*postgres.Jsonb, error) {
	if customerUserId == "" {
		return nil, errors.New("invalid customer user id")
	}

	properties := map[string]interface{}{
		U.UP_USER_ID: customerUserId,
	}

	if U.IsEmail(customerUserId) {
		properties[U.UP_EMAIL] = customerUserId
	}

	return U.EncodeToPostgresJsonb(&properties)
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

	// Prioritize identity properties if customer_user_id is provided
	if user.CustomerUserId != "" {
		identityProperties, err := GetIdentifiedUserPropertiesAsJsonb(user.CustomerUserId)
		if err != nil {
			return nil, errors.New("failed to get identity properties")
		}

		propertiesMap, err := U.DecodePostgresJsonb(identityProperties)
		if err != nil {
			return nil, errors.New("failed to decode identity properties")
		}

		properties, err := U.AddToPostgresJsonb(&user.Properties, *propertiesMap, true)
		if err != nil {
			return nil, errors.New("failed to add identity properties on user properties")
		}
		user.Properties = *properties
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

func GetExistingCustomerUserID(projectId uint64, arrayCustomerUserID []string) (map[string]string, int) {
	db := C.GetServices().Db

	customerUserIDMap := make(map[string]string)

	if len(arrayCustomerUserID) == 0 {
		return nil, http.StatusBadRequest
	}

	queryStmnt := "SELECT" + " " + "DISTINCT(customer_user_id), id" + " FROM " + "users" + " WHERE " + "project_id = ? AND customer_user_id IN ( ? )"
	rows, err := db.Raw(queryStmnt, projectId, arrayCustomerUserID).Rows()
	if err != nil {
		log.WithError(err).Error("Failed to get customer_user_id.")
		return nil, http.StatusInternalServerError
	}

	defer rows.Close()

	for rows.Next() {
		var customerID string
		var userID string
		if err := rows.Scan(&customerID, &userID); err != nil {
			log.WithError(err).Error("Failed scanning rows on GetExistingCustomerUserID")
			return nil, http.StatusInternalServerError
		}
		customerUserIDMap[customerID] = userID
	}

	return customerUserIDMap, http.StatusFound
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

// GetAllUserIDByCustomerUserID returns all users with same customer_user_id
func GetAllUserIDByCustomerUserID(projectID uint64, customerUserID string) ([]string, int) {
	if projectID == 0 || customerUserID == "" {
		return nil, http.StatusBadRequest
	}

	db := C.GetServices().Db

	var users []User
	if err := db.Table("users").Select("distinct(id)").Where("project_id = ? AND customer_user_id=?", projectID, customerUserID).Find(&users).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}
		return nil, http.StatusInternalServerError
	}

	if len(users) == 0 {
		return nil, http.StatusNotFound
	}

	var userIDs []string
	for i := range users {
		userIDs = append(userIDs, users[i].ID)
	}

	return userIDs, http.StatusFound
}

// CreateOrGetSegmentUser create or updates(c_uid) and returns user by segement_anonymous_id
// and/or customer_user_id.
func CreateOrGetSegmentUser(projectId uint64, segAnonId, custUserId string, requestTimestamp int64) (*User, int) {
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

		cUser := &User{ProjectId: projectId, JoinTimestamp: requestTimestamp}

		// add seg_aid, if provided and not exist already.
		if segAnonId != "" {
			cUser.SegmentAnonymousId = segAnonId
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

	// same seg_aid with different c_uid. log error. return user.
	if user.CustomerUserId != custUserId {
		logCtx.Error("Different customer_user_id seen for existing user with segment_anonymous_id.")
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

// Today's cache keys
func GetUsersCachedCacheKey(projectId uint64, dateKey string) (*cacheRedis.Key, error) {
	prefix := "US:LIST"
	return cacheRedis.NewKey(projectId, prefix, dateKey)
}

func GetUserPropertiesCategoryByProjectCacheKey(projectId uint64, property string, category string, dateKey string) (*cacheRedis.Key, error) {
	prefix := "US:PC"
	return cacheRedis.NewKey(projectId, prefix, fmt.Sprintf("%s:%s:%s", dateKey, category, property))

}

func GetValuesByUserPropertyCacheKey(projectId uint64, property_name string, value string, dateKey string) (*cacheRedis.Key, error) {
	prefix := "US:PV"
	return cacheRedis.NewKey(projectId, fmt.Sprintf("%s:%s", prefix, property_name), fmt.Sprintf("%s:%s", dateKey, value))
}

// Rollup cache keys

func GetUserPropertiesCategoryByProjectRollUpCacheKey(projectId uint64, dateKey string) (*cacheRedis.Key, error) {
	prefix := "RollUp:US:PC"
	return cacheRedis.NewKey(projectId, prefix, dateKey)

}

func GetValuesByUserPropertyRollUpCacheKey(projectId uint64, property_name string, dateKey string) (*cacheRedis.Key, error) {
	prefix := "RollUp:US:PV"
	return cacheRedis.NewKey(projectId, fmt.Sprintf("%s:%s", prefix, property_name), dateKey)
}

// Today's cache keys count
func GetUserPropertiesCategoryByProjectCountCacheKey(projectId uint64, dateKey string) (*cacheRedis.Key, error) {
	prefix := "C:US:PC"
	return cacheRedis.NewKeyWithAllProjectsSupport(projectId, prefix, dateKey)
}

func GetValuesByUserPropertyCountCacheKey(projectId uint64, dateKey string) (*cacheRedis.Key, error) {
	prefix := "C:US:PV"
	return cacheRedis.NewKeyWithAllProjectsSupport(projectId, prefix, dateKey)
}

//GetRecentUserPropertyKeysWithLimits This method gets all the recent 'limit' property keys from DB for a given project
func GetRecentUserPropertyKeysWithLimits(projectID uint64, usersLimit int, propertyLimit int, seedDate time.Time) ([]U.Property, error) {
	properties := make([]U.Property, 0)
	db := C.GetServices().Db
	startTime := seedDate.AddDate(0, 0, -7).Unix()
	endTime := seedDate.Unix()
	logCtx := log.WithField("project_id", projectID)
	queryStr := " WITH recent_users AS (SELECT DISTINCT ON(user_id) user_id, user_properties_id FROM events " +
		"WHERE project_id = ? AND timestamp > ? AND timestamp <= ? ORDER BY user_id, timestamp DESC LIMIT ?) " +
		"SELECT json_object_keys(user_properties.properties::json) AS key, COUNT(*) AS count, MAX(updated_timestamp) as last_seen FROM recent_users " +
		"LEFT OUTER JOIN user_properties ON recent_users.user_properties_id = user_properties.id  " +
		"WHERE user_properties.project_id = ? AND user_properties.properties != 'null' GROUP BY key ORDER BY count DESC LIMIT ?;"

	rows, err := db.Raw(queryStr, projectID, startTime, endTime, usersLimit, projectID, propertyLimit).Rows()

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
func GetRecentUserPropertyValuesWithLimits(projectID uint64, propertyKey string, usersLimit, valuesLimit int, seedDate time.Time) ([]U.PropertyValue, string, error) {

	// limit on values returned.
	values := make([]U.PropertyValue, 0, 0)
	startTime := seedDate.AddDate(0, 0, -7).Unix()
	endTime := seedDate.Unix()
	db := C.GetServices().Db
	queryStmnt := " WITH recent_users AS (SELECT DISTINCT ON(user_id) user_id, user_properties_id FROM events " +
		"WHERE project_id = ? AND timestamp > ? AND timestamp <= ? ORDER BY user_id, timestamp DESC LIMIT ?) " +
		"SELECT user_properties.properties->? AS value, COUNT(*) AS count, MAX(updated_timestamp) AS last_seen, MAX(jsonb_typeof(user_properties.properties->?)) AS value_type FROM recent_users " +
		"LEFT JOIN user_properties ON recent_users.user_properties_id = user_properties.id WHERE user_properties.project_id = ? " +
		"AND user_properties.properties != 'null' AND user_properties.properties->? IS NOT NULL GROUP BY value limit ?;"

	queryParams := make([]interface{}, 0, 0)
	queryParams = append(queryParams, projectID, startTime, endTime, usersLimit, propertyKey, propertyKey, projectID, propertyKey, valuesLimit)

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

	return values, U.GetCategoryType(propertyKey, values), nil
}

//GetUserPropertiesByProject This method iterates over n days and gets user properties from cache for a given project
// Picks all past 24 hrs seen properties and sorts the remaining by count and returns top 'limit'
func GetUserPropertiesByProject(projectID uint64, limit int, lastNDays int) (map[string][]string, error) {
	properties := make(map[string][]string)
	if projectID == 0 {
		return properties, errors.New("invalid project on GetUserPropertiesByProject")
	}
	currentDate := OverrideCacheDateRangeForProjects(projectID)
	userProperties := make([]U.CachePropertyWithTimestamp, 0)
	for i := 0; i < lastNDays; i++ {
		currentDateOnlyFormat := currentDate.AddDate(0, 0, -i).Format(U.DATETIME_FORMAT_YYYYMMDD)
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
	logCtx := log.WithFields(log.Fields{
		"project_id": projectID,
	})
	if projectID == 0 {
		return U.CachePropertyWithTimestamp{}, errors.New("invalid project on GetUserPropertiesByProjectFromCache")
	}

	PropertiesKey, err := GetUserPropertiesCategoryByProjectRollUpCacheKey(projectID, dateKey)
	if err != nil {
		return U.CachePropertyWithTimestamp{}, err
	}
	userProperties, _, err := cacheRedis.GetIfExistsPersistent(PropertiesKey)
	if err != nil {
		return U.CachePropertyWithTimestamp{}, err
	}
	if userProperties == "" {
		logCtx.WithField("date_key", dateKey).Info("MISSING ROLLUP US:PC")
		return U.CachePropertyWithTimestamp{}, nil
	}
	var cacheValue U.CachePropertyWithTimestamp
	err = json.Unmarshal([]byte(userProperties), &cacheValue)
	if err != nil {
		return U.CachePropertyWithTimestamp{}, err
	}
	return cacheValue, nil
}

//GetPropertyValuesByUserProperty This method iterates over n days and gets user property values from cache for a given project/property
// Picks all past 24 hrs seen values and sorts the remaining by count and returns top 'limit'
func GetPropertyValuesByUserProperty(projectID uint64, propertyName string, limit int, lastNDays int) ([]string, error) {
	if projectID == 0 {
		return []string{}, errors.New("invalid project on GetPropertyValuesByUserProperty")
	}

	if propertyName == "" {
		return []string{}, errors.New("invalid property_name on GetPropertyValuesByUserProperty")
	}
	currentDate := OverrideCacheDateRangeForProjects(projectID)
	values := make([]U.CachePropertyValueWithTimestamp, 0)
	for i := 0; i < lastNDays; i++ {
		currentDateOnlyFormat := currentDate.AddDate(0, 0, -i).Format(U.DATETIME_FORMAT_YYYYMMDD)
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
	logCtx := log.WithFields(log.Fields{
		"project_id": projectID,
	})
	if projectID == 0 {
		return U.CachePropertyValueWithTimestamp{}, errors.New("invalid project on GetPropertyValuesByUserPropertyFromCache")
	}

	if propertyName == "" {
		return U.CachePropertyValueWithTimestamp{}, errors.New("invalid property_name on GetPropertyValuesByUserPropertyFromCache")
	}

	eventPropertyValuesKey, err := GetValuesByUserPropertyRollUpCacheKey(projectID, propertyName, dateKey)
	if err != nil {
		return U.CachePropertyValueWithTimestamp{}, err
	}
	values, _, err := cacheRedis.GetIfExistsPersistent(eventPropertyValuesKey)
	if err != nil {
		return U.CachePropertyValueWithTimestamp{}, err
	}
	if values == "" {
		logCtx.WithField("date_key", dateKey).Info("MISSING ROLLUP US:PV")
		return U.CachePropertyValueWithTimestamp{}, nil
	}
	var cacheValue U.CachePropertyValueWithTimestamp
	err = json.Unmarshal([]byte(values), &cacheValue)
	if err != nil {
		return U.CachePropertyValueWithTimestamp{}, err
	}
	return cacheValue, nil
}

func GetLatestUserPropertiesOfUserAsMap(projectId uint64, id string) (*map[string]interface{}, int) {
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

// GetUserIdentificationPhoneNumber tries various patterns of phone number if exist in db and return the phone no based on priority
func GetUserIdentificationPhoneNumber(projectID uint64, phoneNo string) (string, string) {
	pPhoneNo := U.GetPossiblePhoneNumber(phoneNo)
	existingPhoneNo, errCode := GetExistingCustomerUserID(projectID, pPhoneNo)
	if errCode == http.StatusFound {
		for i := range pPhoneNo {
			if userID, exist := existingPhoneNo[pPhoneNo[i]]; exist {
				return pPhoneNo[i], userID
			}
		}
	}

	return phoneNo, ""
}
