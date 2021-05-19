package memsql

import (
	"encoding/json"
	"errors"
	cacheRedis "factors/cache/redis"
	C "factors/config"
	"factors/metrics"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"net/http"

	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/imdario/mergo"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

const usersLimitForProperties = 50000

const constraintViolationError = "constraint violation"

func isConstraintViolationError(err error) bool {
	return err.Error() == constraintViolationError
}

func satisfiesUserConstraints(user model.User) int {
	logCtx := log.WithFields(log.Fields{
		"method":     "satisfiesUserConstraints",
		"project_id": user.ProjectId,
		"id":         user.ID,
	})

	if exists := existsUserForProject(user.ProjectId, user.ID); exists {
		logCtx.Error("duplicate user id being inserted for proejct")
	}

	// Unique (project_id, segment_anonymous_id) constraint.
	if user.SegmentAnonymousId != "" {
		if exists := existsUserWithSegmentAnonymousID(user.ProjectId, user.SegmentAnonymousId); exists {
			logCtx.WithField("seg_anon_id", user.SegmentAnonymousId).Error("segment anonymous id constraint violation")
			return http.StatusBadRequest
		}
	}

	// Unique (project_id, amp_user_id) constraint.
	if user.AMPUserId != "" {
		if exists := existsUserWithAMPUserID(user.ProjectId, user.AMPUserId); exists {
			logCtx.WithField("amp_user_id", user.AMPUserId).Error("amp user id constraint violation")
			return http.StatusBadRequest
		}
	}

	return http.StatusOK
}

// createUserWithError - Returns error during create to match
// with constraint errors.
func (store *MemSQL) createUserWithError(user *model.User) (*model.User, error) {
	defer model.LogOnSlowExecution(time.Now())

	logCtx := log.WithField("project_id", user.ProjectId)

	if user.ProjectId == 0 {
		logCtx.Error("Failed to create user. ProjectId not provided.")
		return nil, errors.New("invalid project_id")
	}

	// Add id with our uuid generator, if not given.
	if user.ID == "" {
		user.ID = U.GetUUID()
	}

	if errCode := satisfiesUserConstraints(*user); errCode != http.StatusOK {
		return user, errors.New("constraint violation")
	}

	// Add join timestamp before creation.
	// Increamenting count based on EventNameId, not by EventName.
	if user.JoinTimestamp <= 0 {
		// Default to 60 seconds earlier than now, so that
		// if event is also created simultaneously
		// user join is earlier.
		user.JoinTimestamp = time.Now().Unix() - 60
	}
	if user.PropertiesUpdatedTimestamp <= 0 {
		// Initializing properties updated timestamp at time of creation.
		user.PropertiesUpdatedTimestamp = user.JoinTimestamp
	}
	// adds join timestamp to user properties.
	newUserProperties := map[string]interface{}{
		U.UP_JOIN_TIME: user.JoinTimestamp,
	}

	// Add identification properties, if available.
	// Prioritize identity properties if customer_user_id is provided
	if user.CustomerUserId != "" {
		identityProperties, err := model.GetIdentifiedUserProperties(user.CustomerUserId)
		if err != nil {
			return nil, errors.New("failed to get identity properties")
		}

		// add identity properties to new properties.
		for k, v := range identityProperties {
			newUserProperties[k] = v
		}
	}

	newUserPropertiesJsonb, err := U.AddToPostgresJsonb(&user.Properties, newUserProperties, true)
	if err != nil {
		return nil, err
	}

	// Discourage direct properties update. Update always through
	// UpdateUserProperties method. Setting empty JSON intentionally,
	// to keep the assumption of not null of properties after create.
	user.Properties = postgres.Jsonb{json.RawMessage("{}")}
	db := C.GetServices().Db
	if err := db.Create(user).Error; err != nil {
		return nil, err
	}

	propertiesId, properties, errCode := store.UpdateUserProperties(user.ProjectId, user.ID, newUserPropertiesJsonb, user.JoinTimestamp)
	if errCode == http.StatusInternalServerError {
		return nil, errors.New("failed to update user properties")
	}

	user.PropertiesId = propertiesId
	if properties != nil {
		user.Properties = *properties
	}

	return user, nil
}

func (store *MemSQL) CreateUser(user *model.User) (*model.User, int) {
	defer model.LogOnSlowExecution(time.Now())

	logCtx := log.WithField("project_id", user.ProjectId).
		WithField("user_id", user.ID)

	newUser, err := store.createUserWithError(user)
	if err == nil {
		return newUser, http.StatusCreated
	}

	if IsDuplicateRecordError(err) || isConstraintViolationError(err) {
		if user.ID != "" {
			// Multiple requests trying to create user at the
			// same time should not lead failure permanently,
			// so get the user and return.
			existingUser, errCode := store.GetUser(user.ProjectId, user.ID)
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
func (store *MemSQL) UpdateUser(projectId uint64, id string,
	user *model.User, updateTimestamp int64) (*model.User, int) {
	defer model.LogOnSlowExecution(time.Now())

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

	// Discourage direct properties update.
	// Update always through UpdateUserProperties method.
	userProperties := user.Properties
	// Properties column will not be added as part update
	// when set with empty postgres jsonb as value. Tested.
	user.Properties = postgres.Jsonb{}

	var updatedUser model.User
	db := C.GetServices().Db
	if err := db.Model(&model.User{}).Where("project_id = ?", projectId).Where("id = ?",
		cleanId).Updates(user).Error; err != nil {

		log.WithFields(log.Fields{"user": user}).WithError(err).Error("Failed updating fields by user_id")
		return nil, http.StatusInternalServerError
	}

	_, _, errCode := store.UpdateUserProperties(projectId, id, &userProperties, updateTimestamp)
	if errCode != http.StatusAccepted && errCode != http.StatusNotModified {
		return nil, http.StatusInternalServerError
	}

	return &updatedUser, http.StatusAccepted
}

// UpdateUserProperties only if there is a change in properties values.
func (store *MemSQL) UpdateUserProperties(projectId uint64, id string,
	newProperties *postgres.Jsonb, updateTimestamp int64) (string, *postgres.Jsonb, int) {
	defer model.LogOnSlowExecution(time.Now())

	var userPropertiesID string
	var userProperties *postgres.Jsonb

	// TODO(Dinesh): Remove the block of code after deprecating user_properties
	// table permanantly. Kept this for backward compatability of data.
	var statusV1 int
	if !C.IsUserPropertiesTableWriteDeprecated(projectId) {
		currentPropertiesId, status := getUserPropertiesId(projectId, id)
		if status != http.StatusFound {
			return "", userProperties, status
		}

		userPropertiesID, statusV1 = store.UpdateUserPropertiesByCurrentProperties(projectId, id,
			currentPropertiesId, newProperties, updateTimestamp)
	}

	var statusV2 int
	if C.IsOnTableUserPropertiesWriteAllowed(projectId) {
		userProperties, statusV2 = store.UpdateUserPropertiesV2(projectId, id, newProperties, updateTimestamp)
	}

	if !C.IsUserPropertiesTableWriteDeprecated(projectId) {
		// Return status of V1 till deprecation.
		return userPropertiesID, userProperties, statusV1
	}

	return userPropertiesID, userProperties, statusV2
}

// UpdateUser
func (store *MemSQL) UpdateUserPropertiesByCurrentProperties(projectId uint64, id string,
	currentPropertiesId string, properties *postgres.Jsonb, updateTimestamp int64) (string, int) {
	defer model.LogOnSlowExecution(time.Now())

	if updateTimestamp == 0 {
		return "", http.StatusBadRequest
	}

	properties = U.SanitizePropertiesJsonb(properties)

	// Update properties.
	newPropertiesId, statusCode := store.createUserPropertiesIfChanged(
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
	if err := db.Model(&model.User{}).Where("project_id = ?", projectId).Where("id = ?",
		id).Update("properties_id", newPropertiesId).Error; err != nil {

		log.WithFields(log.Fields{"projectId": projectId,
			"id": id}).WithError(err).Error("Failed updating propertyId")
		return "", http.StatusInternalServerError
	}

	return newPropertiesId, http.StatusAccepted
}

func getUserPropertiesId(projectId uint64, id string) (string, int) {
	defer model.LogOnSlowExecution(time.Now())

	var user model.User
	db := C.GetServices().Db
	if err := db.Select("properties_id").Where("project_id = ?", projectId).Where("id = ?", id).First(&user).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return "", http.StatusNotFound
		}
		return "", http.StatusInternalServerError
	}

	return user.PropertiesId, http.StatusFound
}

func existsUserForProject(projectID uint64, id string) bool {
	defer model.LogOnSlowExecution(time.Now())

	logCtx := log.WithFields(log.Fields{"project_id": projectID, "user_id": id})

	var user model.User
	db := C.GetServices().Db
	if err := db.Limit(1).Where("project_id = ? AND id = ?", projectID, id).Select("id").Find(&user).Error; err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			logCtx.WithError(err).Error("Failed to check user using user_id")
		}
		return false
	}

	if user.ID != "" {
		return true
	}
	return false
}

func (store *MemSQL) GetUser(projectId uint64, id string) (*model.User, int) {
	defer model.LogOnSlowExecution(time.Now())

	logCtx := log.WithFields(log.Fields{"project_id": projectId, "user_id": id})

	var user model.User
	db := C.GetServices().Db
	if err := db.Where("project_id = ?", projectId).Where("id = ?", id).First(&user).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}
		logCtx.WithError(err).Error("Failed to get user using user_id")
		return nil, http.StatusInternalServerError
	}

	if C.ShouldUseUserPropertiesTableForRead(projectId) {
		if user.PropertiesId != "" {
			properties, errCode := store.GetUserProperties(projectId, id, user.PropertiesId)
			if errCode != http.StatusFound {
				return nil, errCode
			}
			user.Properties = *properties
		}
	}

	return &user, http.StatusFound
}

func (store *MemSQL) GetUsers(projectId uint64, offset uint64, limit uint64) ([]model.User, int) {
	defer model.LogOnSlowExecution(time.Now())

	var users []model.User
	db := C.GetServices().Db
	if err := db.Order("created_at").Offset(offset).
		Where("project_id = ?", projectId).Limit(limit).Find(&users).Error; err != nil {
		return nil, http.StatusInternalServerError
	}
	if len(users) == 0 {
		return nil, http.StatusNotFound
	}
	return users, http.StatusFound
}

// GetUsersByCustomerUserID Gets all the users indentified by given customer_user_id in increasing order of updated_at.
func (store *MemSQL) GetUsersByCustomerUserID(projectID uint64, customerUserID string) ([]model.User, int) {
	defer model.LogOnSlowExecution(time.Now())

	logCtx := log.WithFields(log.Fields{
		"ProjectID":      projectID,
		"CustomerUserID": customerUserID,
	})

	var users []model.User
	db := C.GetServices().Db
	if err := db.Order("created_at ASC").
		Where("project_id = ? AND customer_user_id = ?", projectID, customerUserID).
		Find(&users).Error; err != nil {

		logCtx.WithError(err).Error("Failed to get users for customer_user_id")
		return nil, http.StatusInternalServerError
	}
	if len(users) == 0 {
		return nil, http.StatusNotFound
	}

	return users, http.StatusFound
}

func (store *MemSQL) GetUserLatestByCustomerUserId(projectId uint64, customerUserId string) (*model.User, int) {
	defer model.LogOnSlowExecution(time.Now())

	var user model.User
	db := C.GetServices().Db
	if err := db.Order("created_at DESC").Where("project_id = ?", projectId).
		Where("customer_user_id = ?", customerUserId).
		First(&user).Error; err != nil {

		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}
		return nil, http.StatusInternalServerError
	}
	return &user, http.StatusFound
}

func (store *MemSQL) GetExistingCustomerUserID(projectId uint64, arrayCustomerUserID []string) (map[string]string, int) {
	defer model.LogOnSlowExecution(time.Now())

	customerUserIDMap := make(map[string]string)
	if len(arrayCustomerUserID) == 0 {
		return nil, http.StatusBadRequest
	}

	db := C.GetServices().Db
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

func existsUserWithSegmentAnonymousID(projectID uint64, segAnonID string) bool {
	defer model.LogOnSlowExecution(time.Now())

	var user model.User
	db := C.GetServices().Db
	if err := db.Limit(1).Where("project_id = ?", projectID).Where(
		"segment_anonymous_id = ?", segAnonID).Select("id").Find(&user).Error; err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			log.WithField("project_id", projectID).WithField("segment_anonymous_id", segAnonID).
				Error("Failed to get count of users by segment_anonymous_id.")
		}
		return false
	}

	if user.ID != "" {
		return true
	}
	return false
}

func (store *MemSQL) GetUserBySegmentAnonymousId(projectId uint64, segAnonId string) (*model.User, int) {
	defer model.LogOnSlowExecution(time.Now())

	var users []model.User
	db := C.GetServices().Db
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
func (store *MemSQL) CreateOrGetUser(projectId uint64, custUserId string) (*model.User, int) {
	defer model.LogOnSlowExecution(time.Now())

	logCtx := log.WithFields(log.Fields{"project_id": projectId,
		"provided_c_uid": custUserId})
	if custUserId == "" {
		logCtx.Error("No customer user id given")
		return nil, http.StatusBadRequest
	}

	var user *model.User
	var errCode int
	user, errCode = store.GetUserLatestByCustomerUserId(projectId, custUserId)
	if errCode == http.StatusFound {
		return user, http.StatusOK
	}

	if errCode == http.StatusInternalServerError {
		logCtx.WithField("err_code", errCode).Error(
			"Failed to fetching user with provided c_uid.")
		return nil, errCode
	}

	cUser := &model.User{ProjectId: projectId, CustomerUserId: custUserId}

	user, errCode = store.CreateUser(cUser)
	if errCode != http.StatusCreated {
		logCtx.WithField("err_code", errCode).Error(
			"Failed creating user with c_uid. get_segment_user failed.")
		return nil, errCode
	}
	return user, errCode
}

// GetAllUserIDByCustomerUserID returns all users with same customer_user_id
func (store *MemSQL) GetAllUserIDByCustomerUserID(projectID uint64, customerUserID string) ([]string, int) {
	defer model.LogOnSlowExecution(time.Now())

	if projectID == 0 || customerUserID == "" {
		return nil, http.StatusBadRequest
	}

	db := C.GetServices().Db

	var users []model.User
	if err := db.Table("users").Select("distinct(id)").
		Where("project_id = ? AND customer_user_id=?", projectID, customerUserID).
		Find(&users).Error; err != nil {

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
func (store *MemSQL) CreateOrGetSegmentUser(projectId uint64, segAnonId, custUserId string,
	requestTimestamp int64) (*model.User, int) {
	defer model.LogOnSlowExecution(time.Now())

	logCtx := log.WithFields(log.Fields{"project_id": projectId, "seg_aid": segAnonId,
		"provided_c_uid": custUserId})

	// seg_aid not provided.
	if segAnonId == "" && custUserId == "" {
		logCtx.Error("No segment user id or customer user id given")
		return nil, http.StatusBadRequest
	}

	var user *model.User
	var errCode int
	// fetch user by seg_aid, if given.
	// Unique (project_id, segment_anonymous_id) constraint.
	if segAnonId != "" {
		user, errCode = store.GetUserBySegmentAnonymousId(projectId, segAnonId)
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
			user, errCode = store.GetUserLatestByCustomerUserId(projectId, custUserId)
			if errCode == http.StatusFound {
				return user, http.StatusOK
			}

			if errCode == http.StatusInternalServerError {
				logCtx.WithField("err_code", errCode).Error(
					"Failed to fetching user with segment provided c_uid.")
				return nil, errCode
			}
		}

		cUser := &model.User{ProjectId: projectId, JoinTimestamp: requestTimestamp}

		// add seg_aid, if provided and not exist already.
		if segAnonId != "" {
			cUser.SegmentAnonymousId = segAnonId
		}

		user, err := store.createUserWithError(cUser)
		if err != nil {
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
		logCtx.Warn("Different customer_user_id seen for existing user with segment_anonymous_id.")
	}

	// provided and fetched c_uid are same.
	return user, http.StatusOK
}

func existsUserWithAMPUserID(projectId uint64, ampUserId string) bool {
	defer model.LogOnSlowExecution(time.Now())

	var user model.User
	db := C.GetServices().Db

	if err := db.Limit(1).Where("project_id = ? AND amp_user_id = ?",
		projectId, ampUserId).Select("id").Find(&user).Error; err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			log.WithField("project_id", projectId).WithField("amp_user_id", ampUserId).
				Error("Failed to get count of users by amp_user_id")
		}
		return false
	}

	if user.ID != "" {
		return true
	}
	return false
}

func getUserByAMPUserId(projectId uint64, ampUserId string) (*model.User, int) {
	defer model.LogOnSlowExecution(time.Now())

	logCtx := log.WithField("project_id", projectId).WithField(
		"amp_user_id", ampUserId)

	var users []model.User

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

func (store *MemSQL) CreateOrGetAMPUser(projectId uint64, ampUserId string, timestamp int64) (*model.User, int) {
	defer model.LogOnSlowExecution(time.Now())

	if projectId == 0 || ampUserId == "" {
		return nil, http.StatusBadRequest
	}

	logCtx := log.WithField("project_id",
		projectId).WithField("amp_user_id", ampUserId)

	// Unique (project_id, amp_user_id) constraint.
	user, errCode := getUserByAMPUserId(projectId, ampUserId)
	if errCode == http.StatusInternalServerError {
		return nil, errCode
	}

	if errCode == http.StatusFound {
		return user, errCode
	}

	user, err := store.createUserWithError(&model.User{ProjectId: projectId,
		AMPUserId: ampUserId, JoinTimestamp: timestamp})
	if err != nil {
		logCtx.WithError(err).Error(
			"Failed to create user by amp user id on CreateOrGetAMPUser")
		return nil, http.StatusInternalServerError
	}

	return user, http.StatusCreated
}

//GetRecentUserPropertyKeysWithLimits This method gets all the recent 'limit' property keys from DB for a given project
func (store *MemSQL) GetRecentUserPropertyKeysWithLimits(projectID uint64, usersLimit int, propertyLimit int, seedDate time.Time) ([]U.Property, error) {
	defer model.LogOnSlowExecution(time.Now())

	properties := make([]U.Property, 0)
	db := C.GetServices().Db
	startTime := seedDate.AddDate(0, 0, -7).Unix()
	endTime := seedDate.Unix()
	logCtx := log.WithField("project_id", projectID)

	var queryParams []interface{}
	queryStmnt := fmt.Sprintf("WITH recent_user_events AS (SELECT user_id, FIRST(user_properties, FROM_UNIXTIME(events.timestamp)) AS user_properties, FIRST(timestamp, FROM_UNIXTIME(events.timestamp)) AS timestamp FROM events"+" "+
		"WHERE project_id = ? AND timestamp > ? AND timestamp <= ? GROUP BY user_id ORDER BY user_id, timestamp DESC LIMIT %d)"+" "+
		"SELECT user_properties, timestamp as last_seen FROM recent_user_events"+" "+
		"WHERE user_properties != 'null' AND user_properties IS NOT NULL;", usersLimit)

	queryParams = make([]interface{}, 0, 0)
	queryParams = append(queryParams, projectID, startTime, endTime)

	if C.ShouldUseUserPropertiesTableForRead(projectID) {
		queryStmnt = fmt.Sprintf(" WITH recent_users AS (SELECT user_id, FIRST(user_properties_id, FROM_UNIXTIME(events.timestamp)) AS user_properties_id FROM events "+
			"WHERE project_id = ? AND timestamp > ? AND timestamp <= ? GROUP BY user_id ORDER BY user_id, timestamp DESC LIMIT %d) "+
			"SELECT user_properties.properties, updated_timestamp as last_seen FROM recent_users "+
			"LEFT OUTER JOIN user_properties ON recent_users.user_properties_id = user_properties.id  "+
			"WHERE user_properties.project_id = ? AND user_properties.properties != 'null' AND user_properties.properties IS NOT NULL;", usersLimit)
		queryParams = make([]interface{}, 0, 0)
		queryParams = append(queryParams, projectID, startTime, endTime, projectID)
	}

	rows, err := db.Raw(queryStmnt, queryParams...).Rows()
	if err != nil {
		logCtx.WithError(err).Error("Failed to get recent user property keys.")
		return nil, err
	}
	defer rows.Close()

	propertiesCounts := make(map[string]map[string]int64)
	for rows.Next() {
		var lastSeen int64
		var properties postgres.Jsonb
		if err := rows.Scan(&properties, &lastSeen); err != nil {
			logCtx.WithError(err).Error("Failed scanning rows on GetRecentUserPropertyKeysWithLimits")
			return nil, err
		}
		propertiesMap, err := U.DecodePostgresJsonbAsPropertiesMap(&properties)
		if err != nil {
			logCtx.WithError(err).Error("Failed to decode properties on GetRecentUserPropertyKeysWithLimits")
			return nil, err
		}

		for key := range *propertiesMap {
			if _, found := propertiesCounts[key]; found {
				propertiesCounts[key]["count"]++
				propertiesCounts[key]["last_seen"] = U.Max(propertiesCounts[key]["last_seen"], lastSeen)
			} else {
				propertiesCounts[key] = map[string]int64{
					"count":     1,
					"last_seen": lastSeen,
				}
			}
		}
	}

	for propertyKey := range propertiesCounts {
		properties = append(properties, U.Property{
			Key:      propertyKey,
			LastSeen: uint64(propertiesCounts[propertyKey]["last_seen"]),
			Count:    propertiesCounts[propertyKey]["count"]})
	}

	sort.Slice(properties, func(i, j int) bool {
		return properties[i].Count > properties[j].Count
	})

	return properties[:U.MinInt(propertyLimit, len(properties))], nil
}

// GetRecentUserPropertyValuesWithLimits This method gets all the recent 'limit' property values
// from DB for a given project/property
func (store *MemSQL) GetRecentUserPropertyValuesWithLimits(projectID uint64, propertyKey string,
	usersLimit, valuesLimit int, seedDate time.Time) ([]U.PropertyValue, string, error) {
	defer model.LogOnSlowExecution(time.Now())

	// limit on values returned.
	values := make([]U.PropertyValue, 0, 0)
	startTime := seedDate.AddDate(0, 0, -7).Unix()
	endTime := seedDate.Unix()

	var queryParams []interface{}
	queryStmnt := fmt.Sprintf(" WITH recent_user_events AS (SELECT user_id, user_properties, timestamp FROM events"+" "+
		"WHERE project_id = ? AND timestamp > ? AND timestamp <= ? ORDER BY user_id, timestamp DESC LIMIT %d)"+" "+
		"SELECT JSON_EXTRACT_STRING(user_properties, ?) AS value, COUNT(*) AS count, MAX(timestamp) AS last_seen, MAX(JSON_GET_TYPE(JSON_EXTRACT_STRING(user_properties, ?))) AS value_type FROM recent_user_events"+" "+
		"WHERE user_properties != 'null' AND JSON_EXTRACT_STRING(user_properties, ?) IS NOT NULL GROUP BY value limit %d;", usersLimit, valuesLimit)

	queryParams = make([]interface{}, 0, 0)
	queryParams = append(queryParams, projectID, startTime, endTime, propertyKey, propertyKey, propertyKey)

	if C.ShouldUseUserPropertiesTableForRead(projectID) {
		queryStmnt = fmt.Sprintf(" WITH recent_users AS (SELECT user_id, FIRST(user_properties_id, FROM_UNIXTIME(events.timestamp)) AS user_properties_id FROM events "+
			"WHERE project_id = ? AND timestamp > ? AND timestamp <= ? GROUP BY user_id ORDER BY user_id, timestamp DESC LIMIT %d) "+
			"SELECT JSON_EXTRACT_STRING(user_properties.properties, ?) AS value, COUNT(*) AS count, MAX(updated_timestamp) AS last_seen, MAX(JSON_GET_TYPE(JSON_EXTRACT_STRING(user_properties.properties, ?))) AS value_type FROM recent_users "+
			"LEFT JOIN user_properties ON recent_users.user_properties_id = user_properties.id WHERE user_properties.project_id = ? "+
			"AND user_properties.properties != 'null' AND JSON_EXTRACT_STRING(user_properties.properties, ?) IS NOT NULL GROUP BY value limit %d;",
			usersLimit, valuesLimit)

		queryParams = make([]interface{}, 0, 0)
		queryParams = append(queryParams, projectID, startTime, endTime, propertyKey, propertyKey, projectID, propertyKey)
	}

	logCtx := log.WithFields(log.Fields{"project_id": projectID, "property_key": propertyKey, "values_limit": valuesLimit})

	db := C.GetServices().Db
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
func (store *MemSQL) GetUserPropertiesByProject(projectID uint64, limit int, lastNDays int) (map[string][]string, error) {
	defer model.LogOnSlowExecution(time.Now())

	properties := make(map[string][]string)
	if projectID == 0 {
		return properties, errors.New("invalid project on GetUserPropertiesByProject")
	}
	currentDate := model.OverrideCacheDateRangeForProjects(projectID)
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

	propertyDetails, propertyDetailsStatus := store.GetAllPropertyDetailsByProjectID(projectID, "", true)

	for _, v := range userPropertiesSorted {
		category := v.Category

		if propertyDetailsStatus == http.StatusFound {
			if pType, exist := (*propertyDetails)[v.Name]; exist {
				category = pType
			}
		}

		if properties[category] == nil {
			properties[category] = make([]string, 0)
		}
		properties[category] = append(properties[category], v.Name)
	}

	return properties, nil
}

func getUserPropertiesByProjectFromCache(projectID uint64, dateKey string) (U.CachePropertyWithTimestamp, error) {
	defer model.LogOnSlowExecution(time.Now())

	logCtx := log.WithFields(log.Fields{
		"project_id": projectID,
	})
	if projectID == 0 {
		return U.CachePropertyWithTimestamp{}, errors.New("invalid project on GetUserPropertiesByProjectFromCache")
	}

	PropertiesKey, err := model.GetUserPropertiesCategoryByProjectRollUpCacheKey(projectID, dateKey)
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

//GetPropertyValuesByUserProperty This method iterates over n days and gets user property values
// from cache for a given project/property. Picks all past 24 hrs seen values and sorts the
// remaining by count and returns top 'limit'
func (store *MemSQL) GetPropertyValuesByUserProperty(projectID uint64,
	propertyName string, limit int, lastNDays int) ([]string, error) {
	defer model.LogOnSlowExecution(time.Now())

	if projectID == 0 {
		return []string{}, errors.New("invalid project on GetPropertyValuesByUserProperty")
	}

	if propertyName == "" {
		return []string{}, errors.New("invalid property_name on GetPropertyValuesByUserProperty")
	}
	currentDate := model.OverrideCacheDateRangeForProjects(projectID)
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

func getPropertyValuesByUserPropertyFromCache(projectID uint64, propertyName string,
	dateKey string) (U.CachePropertyValueWithTimestamp, error) {
	defer model.LogOnSlowExecution(time.Now())

	logCtx := log.WithFields(log.Fields{
		"project_id": projectID,
	})
	if projectID == 0 {
		return U.CachePropertyValueWithTimestamp{},
			errors.New("invalid project on GetPropertyValuesByUserPropertyFromCache")
	}

	if propertyName == "" {
		return U.CachePropertyValueWithTimestamp{},
			errors.New("invalid property_name on GetPropertyValuesByUserPropertyFromCache")
	}

	eventPropertyValuesKey, err := model.GetValuesByUserPropertyRollUpCacheKey(projectID, propertyName, dateKey)
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

func (store *MemSQL) GetLatestUserPropertiesOfUserAsMap(projectId uint64, id string) (*map[string]interface{}, int) {
	defer model.LogOnSlowExecution(time.Now())

	logCtx := log.WithField("project_id", projectId).WithField("id", id)

	user, errCode := store.GetUser(projectId, id)
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
func (store *MemSQL) GetDistinctCustomerUserIDSForProject(projectID uint64) ([]string, int) {
	defer model.LogOnSlowExecution(time.Now())

	logCtx := log.WithFields(log.Fields{"ProjectID": projectID})

	var customerUserIDS []string
	db := C.GetServices().Db
	rows, err := db.Model(&model.User{}).
		Where("project_id = ? AND customer_user_id IS NOT NULL", projectID).
		Select("distinct customer_user_id").Rows()
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
func (store *MemSQL) GetUserIdentificationPhoneNumber(projectID uint64, phoneNo string) (string, string) {
	defer model.LogOnSlowExecution(time.Now())

	if len(phoneNo) < 5 {
		return "", ""
	}

	pPhoneNo := U.GetPossiblePhoneNumber(phoneNo)
	existingPhoneNo, errCode := store.GetExistingCustomerUserID(projectID, pPhoneNo)
	if errCode == http.StatusFound {
		for i := range pPhoneNo {
			if userID, exist := existingPhoneNo[pPhoneNo[i]]; exist {
				return pPhoneNo[i], userID
			}
		}
	}

	return phoneNo, ""
}

func (store *MemSQL) FixAllUsersJoinTimestampForProject(db *gorm.DB, projectId uint64, isDryRun bool) error {
	defer model.LogOnSlowExecution(time.Now())

	userRows, err := db.Raw("SELECT id, join_timestamp FROM users WHERE project_id = ?", projectId).Rows()
	defer userRows.Close()
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("SQL Query failed.")
		return err
	}
	for userRows.Next() {
		var userId string
		var joinTimestamp int64
		if err = userRows.Scan(&userId, &joinTimestamp); err != nil {
			log.WithFields(log.Fields{"err": err}).Error("SQL Parse failed.")
			return err
		}
		type Result struct {
			Timestamp int64
		}
		var result Result
		db.Raw("SELECT MIN(timestamp) as Timestamp FROM events WHERE user_id = ? AND project_id = ?", userId, projectId).Scan(&result)
		if result.Timestamp > 0 && result.Timestamp < joinTimestamp {
			newJoinTimestamp := result.Timestamp - 60
			log.WithFields(log.Fields{
				"userId":            userId,
				"userJoinTimestamp": joinTimestamp,
				"minEventTimestamp": result.Timestamp,
				"newJoinTimestamp":  newJoinTimestamp,
			}).Error("Need to update.")
			if !isDryRun {
				db.Exec("UPDATE users SET join_timestamp=? WHERE project_id=? AND id=?", newJoinTimestamp, projectId, userId)
				log.Info(fmt.Sprintf("Updated %s", userId))
			}
		}
	}
	return nil
}

func (store *MemSQL) GetUserPropertiesByUserID(projectID uint64, id string) (*postgres.Jsonb, int) {
	defer model.LogOnSlowExecution(time.Now())

	logCtx := log.WithField("project_id", projectID).WithField("user_id", id)

	if projectID == 0 || id == "" {
		logCtx.Error("Invalid values on arguments.")
		return nil, http.StatusBadRequest
	}

	var user model.User
	db := C.GetServices().Db
	if err := db.Model(&model.User{}).Where("project_id = ? AND id = ?", projectID, id).
		Select("properties").Find(&user).Error; err != nil {

		logCtx.WithError(err).Error("Failed to get properties of user.")
		return nil, http.StatusInternalServerError
	}

	if U.IsEmptyPostgresJsonb(&user.Properties) {
		logCtx.WithField("properties", user.Properties).Error("Empty or nil properties for user.")
		return nil, http.StatusNotFound
	}

	return &user.Properties, http.StatusFound
}

// GetUserByPropertyKey - Returns first user which has the
// given property with value. No specific order.
func (store *MemSQL) GetUserByPropertyKey(projectID uint64,
	key string, value interface{}) (*model.User, int) {
	defer model.LogOnSlowExecution(time.Now())

	logCtx := log.WithField("project_id", projectID).WithField(
		"key", key).WithField("value", value)

	var user model.User
	// $$$ is a gorm alias for ? jsonb operator.
	db := C.GetServices().Db
	err := db.Limit(1).Where("project_id=?", projectID).Where(
		"JSON_EXTRACT_STRING(properties, ?) = ?", key, value).Find(&user).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}

		logCtx.WithError(err).Error("Failed to get user id by key.")
		return nil, http.StatusInternalServerError
	}

	return &user, http.StatusFound
}

func mergeUserPropertiesByCustomerUserID(projectID uint64, users []model.User) (*map[string]interface{}, int) {
	logCtx := log.WithField("project_id", projectID).WithField("users", users)
	usersLength := len(users)
	if usersLength == 0 {
		logCtx.Error("No users for merging the user_properties.")
		return nil, http.StatusInternalServerError
	}

	initialPropertiesVisitedMap := make(map[string]bool)
	for _, property := range U.USER_PROPERTIES_MERGE_TYPE_INITIAL {
		initialPropertiesVisitedMap[property] = false
	}

	// Order the properties before merging the properties to
	// ensure the precendence of value.
	sort.Slice(users, func(i, j int) bool {
		return users[i].PropertiesUpdatedTimestamp < users[j].PropertiesUpdatedTimestamp
	})

	mergedUserProperties := make(map[string]interface{})
	mergedUserPropertiesValues := make(map[string][]interface{})
	var mergedUpdatedTimestamp int64
	for i := range users {
		user := users[i]
		userProperties, err := U.DecodePostgresJsonb(&user.Properties)
		if err != nil {
			logCtx.WithField("user_properties", user.Properties).
				Error("Failed to decode user properties on merge.")
			return &mergedUserProperties, http.StatusInternalServerError
		}
		if user.PropertiesUpdatedTimestamp > mergedUpdatedTimestamp {
			mergedUpdatedTimestamp = user.PropertiesUpdatedTimestamp
		}

		for property := range *userProperties {
			mergedUserPropertiesValues[property] = append(mergedUserPropertiesValues[property], (*userProperties)[property])
			isAlreadySet, isInitialProperty := initialPropertiesVisitedMap[property]
			if isInitialProperty {
				if !isAlreadySet {
					// For initial properties, set only once for earliest user.
					mergedUserProperties[property] = (*userProperties)[property]
					initialPropertiesVisitedMap[property] = true
				}
			} else if !U.StringValueIn(property, U.USER_PROPERTIES_MERGE_TYPE_ADD[:]) &&
				!isEmptyPropertyValue((*userProperties)[property]) {
				// For all other properties, overwrite with the latest user property.
				mergedUserProperties[property] = (*userProperties)[property]
			}
		}
	}

	// Handle merge for add type properties separately.
	userPropertiesToBeMerged := make([]postgres.Jsonb, 0, 0)
	for i := range users {
		userPropertiesToBeMerged = append(userPropertiesToBeMerged, users[i].Properties)
	}
	mergeAddTypeUserProperties(&mergedUserProperties, userPropertiesToBeMerged)

	// Additional check for properties that can be added. If merge is triggered for users with same set of properties,
	// value of properties that can be added will change after addition. Below check is to avoid update in such case.
	if !anyPropertyChanged(mergedUserPropertiesValues, len(users)) {
		return &mergedUserProperties, http.StatusOK
	}
	mergedUserProperties[U.UP_MERGE_TIMESTAMP] = U.TimeNowUnix()

	// Disabled temporary sanitization.
	// pg.SanitizeAddTypeProperties(projectID, users, &mergedUserProperties)

	return &mergedUserProperties, http.StatusOK
}

func (store *MemSQL) getUsersForMergingPropertiesByCustomerUserID(projectID uint64,
	customerUserID string) ([]model.User, int) {
	defer model.LogOnSlowExecution(time.Now())

	logCtx := log.WithField("project_id", projectID).
		WithField("customer_user_id", customerUserID)

	if projectID == 0 || customerUserID == "" {
		logCtx.Error("Invalid values for arguments.")
		return []model.User{}, http.StatusBadRequest
	}

	// Users are returned in increasing order of created_at. For user_properties created at same unix time,
	// older user order will help in ensuring the order while merging properties.
	users, errCode := store.GetUsersByCustomerUserID(projectID, customerUserID)
	if errCode == http.StatusInternalServerError || errCode == http.StatusNotFound {
		return users, errCode
	}

	usersLength := len(users)
	if usersLength > 10 {
		metrics.Increment(metrics.IncrUserPropertiesMergeMoreThan10)
	}

	if usersLength > model.MaxUsersForPropertiesMerge {
		// If number of users to merge are more than max allowed, merge for oldest max/2 and latest max/2.
		users = append(users[0:model.MaxUsersForPropertiesMerge/2],
			users[usersLength-model.MaxUsersForPropertiesMerge/2:usersLength]...)
	}
	metrics.Increment(metrics.IncrUserPropertiesMergeCount)

	return users, http.StatusFound
}

func (store *MemSQL) mergeNewPropertiesWithCurrentUserProperties(projectID uint64, userID string,
	currentProperties *postgres.Jsonb, currentUpdateTimestamp int64,
	newProperties *postgres.Jsonb, newUpdateTimestamp int64) (*postgres.Jsonb, int) {
	defer model.LogOnSlowExecution(time.Now())

	logCtx := log.WithField("project_id", projectID)

	var newPropertiesMap map[string]interface{}
	err := json.Unmarshal((*newProperties).RawMessage, &newPropertiesMap)
	if err != nil {
		logCtx.WithError(err).WithField("new_properties", newPropertiesMap).
			Error("Failed to unmarshal new properties of user.")
		return nil, http.StatusInternalServerError
	}

	if len(newPropertiesMap) == 0 {
		return nil, http.StatusNotModified
	}

	if currentProperties == nil {
		logCtx.WithField("current_properties", currentProperties).
			Error("User properties of existing user is empty.")
		return nil, http.StatusInternalServerError
	}

	// Initialize merged user properties with current user_properties.
	var mergedPropertiesMap map[string]interface{}
	err = json.Unmarshal((*currentProperties).RawMessage, &mergedPropertiesMap)
	if err != nil {
		logCtx.WithError(err).WithField("current_properties", currentProperties).
			Error("Failed to unmarshal current properties of user.")
		return nil, http.StatusInternalServerError
	}

	var currentPropertiesMap map[string]interface{}
	json.Unmarshal((*currentProperties).RawMessage, &currentPropertiesMap)

	// Overwrite the keys only, if the update is future, else only add new keys.
	if newUpdateTimestamp >= currentUpdateTimestamp {
		mergo.Merge(&mergedPropertiesMap, newPropertiesMap, mergo.WithOverride)

		// For fixing the meta identifier object which was a string earlier and changed to JSON.
		// Mergo doesn't consider change in datatype as change in the same key.
		if _, exists := newPropertiesMap[U.UP_META_OBJECT_IDENTIFIER_KEY]; exists {
			mergedPropertiesMap[U.UP_META_OBJECT_IDENTIFIER_KEY] = newPropertiesMap[U.UP_META_OBJECT_IDENTIFIER_KEY]
		}
	} else {
		mergo.Merge(&mergedPropertiesMap, newPropertiesMap)
	}

	// Using merged properties for equality check to achieve
	// currentPropertiesMap {x: 1, y: 2} newPropertiesMap {x: 1} -> true
	if reflect.DeepEqual(currentPropertiesMap, mergedPropertiesMap) {
		if len(currentPropertiesMap) > len(mergedPropertiesMap) {
			store.UpdateCacheForUserProperties(userID, projectID, currentPropertiesMap, true)
			return currentProperties, http.StatusNotModified
		}
	}

	store.UpdateCacheForUserProperties(userID, projectID, mergedPropertiesMap, false)
	mergedPropertiesJSON, err := U.EncodeToPostgresJsonb(&mergedPropertiesMap)
	if err != nil {
		logCtx.Error("Failed to marshal new properties merged to current user properties.")
		return nil, http.StatusInternalServerError
	}

	return mergedPropertiesJSON, http.StatusOK
}

// UpdateUserPropertiesV2 - Merge new properties with the existing properties of user and also
// merge the properties of user with same customer_user_id, then updates properties on users table.
func (store *MemSQL) UpdateUserPropertiesV2(projectID uint64, id string,
	newProperties *postgres.Jsonb, newUpdateTimestamp int64) (*postgres.Jsonb, int) {
	defer model.LogOnSlowExecution(time.Now())

	logCtx := log.WithField("project_id", projectID).WithField("id", id).
		WithField("new_properties", newProperties).
		WithField("new_update_timestamp", newUpdateTimestamp)

	newProperties = U.SanitizePropertiesJsonb(newProperties)

	user, errCode := store.GetUser(projectID, id)
	if errCode != http.StatusFound {
		return nil, http.StatusInternalServerError
	}

	newPropertiesMergedJSON, errCode := store.mergeNewPropertiesWithCurrentUserProperties(projectID, id,
		&user.Properties, user.PropertiesUpdatedTimestamp, newProperties, newUpdateTimestamp)
	if errCode == http.StatusNotModified {
		return &user.Properties, http.StatusNotModified
	}
	if errCode != http.StatusOK {
		logCtx.Error("Failed merging current properties with new properties on update_properties v2.")
		return nil, http.StatusInternalServerError
	}

	// Skip merge by customer_user_id, if customer_user_id is not available.
	if user.CustomerUserId == "" {
		errCode = store.OverwriteUserPropertiesByID(projectID, id, newPropertiesMergedJSON, true, newUpdateTimestamp)
		if errCode == http.StatusInternalServerError || errCode == http.StatusBadRequest {
			return nil, http.StatusInternalServerError
		}

		return newPropertiesMergedJSON, http.StatusAccepted
	}

	users, errCode := store.getUsersForMergingPropertiesByCustomerUserID(projectID, user.CustomerUserId)
	if errCode != http.StatusFound {
		logCtx.Error("Failed to get user by customer_user_id for merging user properties.")
		return &user.Properties, http.StatusInternalServerError
	}

	// Skip merge by customer_user_id, if only the current user has the customer_user_id.
	if len(users) == 1 {
		errCode = store.OverwriteUserPropertiesByID(projectID, id, newPropertiesMergedJSON, true, newUpdateTimestamp)
		if errCode == http.StatusInternalServerError || errCode == http.StatusBadRequest {
			return nil, http.StatusInternalServerError
		}

		return newPropertiesMergedJSON, http.StatusAccepted
	}

	// map[user_id]map[property]value
	userPropertiesOriginalValues := make(map[string]map[string]interface{}, 0)

	// Update the merged properties to the current user's object before passing it on
	// for merging by customer user id. The merge method orders by updated timestamp
	// before merging.
	for i := range users {
		if users[i].ID == id {
			users[i].Properties = *newPropertiesMergedJSON
			users[i].PropertiesUpdatedTimestamp = newUpdateTimestamp
		}

		// Create a map of original values to overwrite, as part of
		// skipping user properties merge by customer user_id.
		userPropertiesMap, err := U.DecodePostgresJsonb(&users[i].Properties)
		if err != nil {
			logCtx.WithField("user_id", users[i].ID).Error("Failed to decode existing user_properties.")
			continue
		}

		for _, property := range model.UserPropertiesToSkipOnMergeByCustomerUserID {
			if _, exists := (*userPropertiesMap)[property]; exists {
				if _, userExists := userPropertiesOriginalValues[users[i].ID]; !userExists {
					userPropertiesOriginalValues[users[i].ID] = make(map[string]interface{}, 0)
				}
				userPropertiesOriginalValues[users[i].ID][property] = (*userPropertiesMap)[property]
			}
		}
	}

	mergedByCustomerUserIDMap, errCode := mergeUserPropertiesByCustomerUserID(projectID, users)
	if errCode != http.StatusOK {
		return nil, http.StatusInternalServerError
	}

	// Overwrite filtered users with same customer_user_id, with the newly
	// merged user_properties by customer_user_id.
	var hasFailure bool
	var mergedPropertiesOfUserJSON *postgres.Jsonb
	for _, user := range users {
		// Overwrite the merged user_properites with original values.
		mergedPropertiesAfterSkipMap := *mergedByCustomerUserIDMap
		if _, userExists := userPropertiesOriginalValues[user.ID]; userExists {
			for _, property := range model.UserPropertiesToSkipOnMergeByCustomerUserID {
				mergedPropertiesAfterSkipMap[property] = userPropertiesOriginalValues[user.ID][property]
			}
		}

		mergedPropertiesAfterSkipJSON, err := U.EncodeToPostgresJsonb(&mergedPropertiesAfterSkipMap)
		if err != nil {
			logCtx.WithError(err).Error("Failed to marshal user properties merged by customer_user_id")
			return nil, http.StatusInternalServerError
		}

		if user.ID == id {
			// Merged user_properties by customer_user_id and original values of
			// properties for user to return.
			// This makes sure the event level user_properties also contain event user's
			// properties original values are preserved. i.e $hubspot_contact_lead_guid.
			mergedPropertiesOfUserJSON = mergedPropertiesAfterSkipJSON
		}

		errCode = store.OverwriteUserPropertiesByID(projectID, user.ID,
			mergedPropertiesAfterSkipJSON, true, newUpdateTimestamp)
		if errCode == http.StatusInternalServerError || errCode == http.StatusBadRequest {
			logCtx.WithField("user_id", user.ID).Error("Failed to update merged user properties on user.")
			hasFailure = true
		}
	}

	if hasFailure {
		return nil, http.StatusInternalServerError
	}

	return mergedPropertiesOfUserJSON, http.StatusAccepted
}

// OverwriteUserPropertiesByCustomerUserID - Update the properties column value
// of all users which has the given customer_user_id, with given properties JSON.
func (store *MemSQL) OverwriteUserPropertiesByCustomerUserID(projectID uint64,
	customerUserID string, properties *postgres.Jsonb, updateTimestamp int64) int {
	defer model.LogOnSlowExecution(time.Now())

	logCtx := log.WithField("project_id", projectID).WithField("customer_user_id", customerUserID)

	if properties == nil {
		logCtx.Error("Failed to overwrite properties. Nil properties.")
		return http.StatusBadRequest
	}

	db := C.GetServices().Db
	err := db.Model(&model.User{}).
		Where("project_id = ? AND customer_user_id = ?", projectID, customerUserID).
		Update(map[string]interface{}{
			"properties":                   properties,
			"properties_updated_timestamp": updateTimestamp,
		}).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to overwrite user properteis.")
		return http.StatusInternalServerError
	}

	return http.StatusAccepted
}

func (store *MemSQL) OverwriteUserPropertiesByID(projectID uint64, id string,
	properties *postgres.Jsonb, withUpdateTimestamp bool, updateTimestamp int64) int {
	defer model.LogOnSlowExecution(time.Now())

	logCtx := log.WithField("project_id", projectID).WithField("id", id).
		WithField("update_timestamp", updateTimestamp)

	if projectID == 0 || id == "" {
		logCtx.Error("Failed to overwrite properties. Empty or nil properties.")
		return http.StatusBadRequest
	}

	if properties == nil {
		logCtx.Error("Failed to overwrite properties. Empty or nil properties.")
		return http.StatusBadRequest
	}

	if withUpdateTimestamp && updateTimestamp == 0 {
		logCtx.Error("Invalid update_timestamp.")
		return http.StatusBadRequest
	}

	update := map[string]interface{}{"properties": properties}
	if updateTimestamp > 0 {
		update["properties_updated_timestamp"] = updateTimestamp
	}

	db := C.GetServices().Db
	err := db.Model(&model.User{}).Where("project_id = ? AND id = ?", projectID, id).Update(update).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to overwrite user properteis.")
		return http.StatusInternalServerError
	}

	return http.StatusAccepted
}
