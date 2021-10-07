package postgres

import (
	"encoding/json"
	"errors"
	cacheRedis "factors/cache/redis"
	C "factors/config"
	"factors/metrics"
	"factors/model/model"
	"factors/util"
	U "factors/util"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/imdario/mergo"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

const usersLimitForProperties = 50000
const uniqueIndexProjectIdAmpUserId = "users_project_id_amp_user_idx"
const uniqueIndexProjectIdSegmentAnonymousId = "users_project_id_segment_anonymous_uidx"

// createUserWithError - Returns error during create to match
// with constraint errors.
func (pg *Postgres) createUserWithError(user *model.User) (*model.User, error) {
	logCtx := log.WithField("project_id", user.ProjectId)

	if user.ProjectId == 0 {
		logCtx.Error("Failed to create user. ProjectId not provided.")
		return nil, errors.New("invalid project_id")
	}

	// Add id with our uuid generator, if not given.
	if user.ID == "" {
		user.ID = U.GetUUID()
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

	properties, errCode := pg.UpdateUserProperties(user.ProjectId, user.ID, newUserPropertiesJsonb, user.JoinTimestamp)
	if errCode == http.StatusInternalServerError {
		return nil, errors.New("failed to update user properties")
	}

	if properties != nil {
		user.Properties = *properties
	}

	return user, nil
}

func (pg *Postgres) CreateUser(user *model.User) (string, int) {
	logCtx := log.WithField("project_id", user.ProjectId).
		WithField("user_id", user.ID)

	if user.SegmentAnonymousId != "" || user.AMPUserId != "" {
		// Corresponding create methods should be used for
		// users from different platform.
		logCtx.Error("Unsupported user of create_user method.")
		return "", http.StatusBadRequest
	}

	newUser, err := pg.createUserWithError(user)
	if err == nil {
		return newUser.ID, http.StatusCreated
	}

	if U.IsPostgresIntegrityViolationError(err) {
		if user.ID != "" {
			// Multiple requests trying to create user at the
			// same time should not lead failure permanently,
			// so get the user and return.
			existingUser, errCode := pg.GetUser(user.ProjectId, user.ID)
			if errCode == http.StatusFound {
				// Using StatusCreated for consistency.
				return existingUser.ID, http.StatusCreated
			}

			return "", errCode
		}

		logCtx.WithError(err).Error("Failed to create user. Integrity violation.")
		return "", http.StatusNotAcceptable
	}

	logCtx.WithError(err).Error("Failed to create user.")
	return "", http.StatusInternalServerError
}

// UpdateUser updates user fields by Id.
func (pg *Postgres) UpdateUser(projectId uint64, id string,
	user *model.User, updateTimestamp int64) (*model.User, int) {

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

	_, errCode := pg.UpdateUserProperties(projectId, id, &userProperties, updateTimestamp)
	if errCode != http.StatusAccepted && errCode != http.StatusNotModified {
		return nil, http.StatusInternalServerError
	}

	return &updatedUser, http.StatusAccepted
}

// UpdateUserProperties only if there is a change in properties values.
func (pg *Postgres) UpdateUserProperties(projectId uint64, id string,
	newProperties *postgres.Jsonb, updateTimestamp int64) (*postgres.Jsonb, int) {

	return pg.UpdateUserPropertiesV2(projectId, id, newProperties, updateTimestamp)
}

func (pg *Postgres) mergeNewPropertiesWithCurrentUserProperties(projectID uint64, userID string,
	currentProperties *postgres.Jsonb, currentUpdateTimestamp int64,
	newProperties *postgres.Jsonb, newUpdateTimestamp int64,
) (*postgres.Jsonb, int) {
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
			pg.UpdateCacheForUserProperties(userID, projectID, currentPropertiesMap, true)
			return currentProperties, http.StatusNotModified
		}
	}

	pg.UpdateCacheForUserProperties(userID, projectID, mergedPropertiesMap, false)
	mergedPropertiesJSON, err := U.EncodeToPostgresJsonb(&mergedPropertiesMap)
	if err != nil {
		logCtx.Error("Failed to marshal new properties merged to current user properties.")
		return nil, http.StatusInternalServerError
	}

	return mergedPropertiesJSON, http.StatusOK
}

func (pg *Postgres) getUsersForMergingPropertiesByCustomerUserID(projectID uint64, customerUserID string) ([]model.User, int) {
	logCtx := log.WithField("project_id", projectID).WithField("customer_user_id", customerUserID)

	if projectID == 0 || customerUserID == "" {
		logCtx.Error("Invalid values for arguments.")
		return []model.User{}, http.StatusBadRequest
	}

	// Users are returned in increasing order of created_at. For user_properties created at same unix time,
	// older user order will help in ensuring the order while merging properties.
	//	users, errCode := pg.GetUsersByCustomerUserID(projectID, customerUserID)
	var limit uint64 = 10000
	var numUsers uint64 = 100
	users, errCode := pg.GetSelectedUsersByCustomerUserID(projectID, customerUserID, limit, numUsers)
	if errCode == http.StatusInternalServerError || errCode == http.StatusNotFound {
		return users, errCode
	}

	usersLength := len(users)
	if usersLength > model.MaxUsersForPropertiesMerge {
		// If number of users to merge are more than max allowed, merge for oldest max/2 and latest max/2.
		users = append(users[0:model.MaxUsersForPropertiesMerge/2],
			users[usersLength-model.MaxUsersForPropertiesMerge/2:usersLength]...)
	}
	metrics.Increment(metrics.IncrUserPropertiesMergeCount)

	return users, http.StatusFound
}

// UpdateUserPropertiesV2 - Merge new properties with the existing properties of user and also
// merge the properties of user with same customer_user_id, then updates properties on users table.
func (pg *Postgres) UpdateUserPropertiesV2(projectID uint64, id string,
	newProperties *postgres.Jsonb, newUpdateTimestamp int64) (*postgres.Jsonb, int) {
	logCtx := log.WithField("project_id", projectID).WithField("id", id).
		WithField("new_properties", newProperties).WithField("new_update_timestamp", newUpdateTimestamp)

	newProperties = U.SanitizePropertiesJsonb(newProperties)

	user, errCode := pg.GetUser(projectID, id)
	if errCode != http.StatusFound {
		return nil, http.StatusInternalServerError
	}

	newPropertiesMergedJSON, errCode := pg.mergeNewPropertiesWithCurrentUserProperties(projectID, id,
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
		errCode = pg.OverwriteUserPropertiesByID(projectID, id, newPropertiesMergedJSON, true, newUpdateTimestamp)
		if errCode == http.StatusInternalServerError || errCode == http.StatusBadRequest {
			return nil, http.StatusInternalServerError
		}

		return newPropertiesMergedJSON, http.StatusAccepted
	}

	users, errCode := pg.getUsersForMergingPropertiesByCustomerUserID(projectID, user.CustomerUserId)
	if errCode != http.StatusFound {
		logCtx.Error("Failed to get user by customer_user_id for merging user properties.")
		return &user.Properties, http.StatusInternalServerError
	}

	// Skip merge by customer_user_id, if only the current user has the customer_user_id.
	if len(users) == 1 {
		errCode = pg.OverwriteUserPropertiesByID(projectID, id, newPropertiesMergedJSON, true, newUpdateTimestamp)
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

	mergedByCustomerUserIDMap, errCode := model.MergeUserPropertiesByCustomerUserID(projectID, users, user.CustomerUserId)
	if errCode != http.StatusOK {
		return nil, http.StatusInternalServerError
	}

	// Overwrite filtered users with same customer_user_id, with the newly
	// merged user_properties by customer_user_id.
	var hasFailure bool
	var mergedPropertiesOfUserJSON *postgres.Jsonb
	for _, user := range users {
		mergedPropertiesAfterSkipMap := U.GetFilteredMapBySkipList(mergedByCustomerUserIDMap, model.UserPropertiesToSkipOnMergeByCustomerUserID)
		if _, userExist := userPropertiesOriginalValues[user.ID]; userExist {

			for property := range userPropertiesOriginalValues[user.ID] {
				(*mergedPropertiesAfterSkipMap)[property] = userPropertiesOriginalValues[user.ID][property]
			}
		}

		mergedPropertiesAfterSkipJSON, err := U.EncodeToPostgresJsonb(mergedPropertiesAfterSkipMap)
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

		errCode = pg.OverwriteUserPropertiesByID(projectID, user.ID,
			mergedPropertiesAfterSkipJSON, true, newUpdateTimestamp)
		if errCode == http.StatusInternalServerError || errCode == http.StatusBadRequest {
			logCtx.WithField("err_code", errCode).WithField("user_id", user.ID).Error("Failed to update merged user properties on user.")
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
func (pg *Postgres) OverwriteUserPropertiesByCustomerUserID(projectID uint64,
	customerUserID string, properties *postgres.Jsonb, updateTimestamp int64) int {
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

func (pg *Postgres) OverwriteUserPropertiesByID(projectID uint64, id string,
	properties *postgres.Jsonb, withUpdateTimestamp bool, updateTimestamp int64) int {

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

	currentPropertiesUpdatedTimestamp, status := pg.GetPropertiesUpdatedTimestampOfUser(projectID, id)
	if status != http.StatusFound {
		logCtx.WithField("status", status).Error("Failed to get propertiesUpdatedTimestamp for the user.")
		return http.StatusBadRequest
	}

	update := map[string]interface{}{"properties": properties}
	if updateTimestamp > 0 && updateTimestamp > currentPropertiesUpdatedTimestamp {
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

func (pg *Postgres) IsUserExistByID(projectID uint64, id string) int {
	defer model.LogOnSlowExecutionWithParams(time.Now(), nil)

	logCtx := log.WithFields(log.Fields{"project_id": projectID, "user_id": id})

	var user model.User
	db := C.GetServices().Db
	if err := db.Limit(1).Where("project_id = ? AND id = ?", projectID, id).
		Select("id").Find(&user).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return http.StatusNotFound
		}

		logCtx.WithError(err).Error("Failed to check is user exists by id.")
		return http.StatusInternalServerError
	}

	if user.ID == "" {
		return http.StatusNotFound
	}

	return http.StatusFound
}

func (pg *Postgres) GetPropertiesUpdatedTimestampOfUser(projectId uint64, id string) (int64, int) {
	db := C.GetServices().Db
	logCtx := log.WithFields(log.Fields{"project_id": projectId, "user_id": id})

	var user model.User
	if err := db.Limit(1).Where("project_id = ?", projectId).Where("id = ?", id).
		Select("properties_updated_timestamp").Find(&user).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return 0, http.StatusNotFound
		}
		logCtx.WithError(err).Error("Failed to get properties_updated_timestamp using user_id.")
		return 0, http.StatusInternalServerError
	}

	return user.PropertiesUpdatedTimestamp, http.StatusFound
}

func (pg *Postgres) GetUser(projectId uint64, id string) (*model.User, int) {
	db := C.GetServices().Db
	logCtx := log.WithFields(log.Fields{"project_id": projectId, "user_id": id})

	var user model.User
	if err := db.Where("project_id = ?", projectId).Where("id = ?", id).First(&user).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}
		logCtx.WithError(err).Error("Failed to get user using user_id")
		return nil, http.StatusInternalServerError
	}

	return &user, http.StatusFound
}

func (pg *Postgres) GetUsers(projectId uint64, offset uint64, limit uint64) ([]model.User, int) {
	db := C.GetServices().Db

	var users []model.User
	if err := db.Order("created_at").Offset(offset).Where("project_id = ?", projectId).Limit(limit).Find(&users).Error; err != nil {
		return nil, http.StatusInternalServerError
	}
	if len(users) == 0 {
		return nil, http.StatusNotFound
	}
	return users, http.StatusFound
}

// GetUsersByCustomerUserID Gets all the users indentified by given customer_user_id in increasing order of updated_at.
func (pg *Postgres) GetUsersByCustomerUserID(projectID uint64, customerUserID string) ([]model.User, int) {
	db := C.GetServices().Db
	logCtx := log.WithFields(log.Fields{
		"ProjectID":      projectID,
		"CustomerUserID": customerUserID,
	})

	var users []model.User
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

// GetSelectedUsersByCustomerUserID gets selected (top 50 & bottom 50) users identified by given customer_user_id in increasing order of updated_at.
func (pg *Postgres) GetSelectedUsersByCustomerUserID(projectID uint64, customerUserID string, limit uint64, numUsers uint64) ([]model.User, int) {
	db := C.GetServices().Db
	logCtx := log.WithFields(log.Fields{
		"ProjectID":      projectID,
		"CustomerUserID": customerUserID,
	})

	var ids []model.User
	if err := db.Limit(limit).Order("created_at ASC").
		Where("project_id = ? AND customer_user_id = ?", projectID, customerUserID).
		Select("id").Find(&ids).Error; err != nil {
		logCtx.WithError(err).Error("Failed to get selected users for customer_user_id")
		return nil, http.StatusInternalServerError
	}

	pulledUsersCount := len(ids)
	if pulledUsersCount > 10 {
		// Log based metric has been created using this log entry.
		logCtx.WithField("UsersCount", pulledUsersCount).
			Info("No.of users with same customer_user_id has exceeded 10.")
	}

	var userIDs []string
	if len(ids) >= int(numUsers) {
		for i := 0; i < int(numUsers/2); i++ {
			userIDs = append(userIDs, ids[i].ID)
		}

		for i := len(ids) - 1; i >= len(ids)-int(numUsers/2); i-- {
			userIDs = append(userIDs, ids[i].ID)
		}
	} else {
		for i := 0; i < len(ids); i++ {
			userIDs = append(userIDs, ids[i].ID)
		}
	}

	var users []model.User
	if err := db.Order("created_at ASC").
		Where("project_id = ? AND id IN ( ? )", projectID, userIDs).
		Find(&users).Error; err != nil {
		logCtx.WithError(err).Error("Failed to get selected users for id")
		return nil, http.StatusInternalServerError
	}

	return users, http.StatusFound
}

func (pg *Postgres) GetUserLatestByCustomerUserId(projectId uint64, customerUserId string) (*model.User, int) {
	db := C.GetServices().Db

	var user model.User
	if err := db.Order("created_at DESC").Where("project_id = ?", projectId).Where(
		"customer_user_id = ?", customerUserId).First(&user).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}
		return nil, http.StatusInternalServerError
	}
	return &user, http.StatusFound
}

func (pg *Postgres) GetExistingCustomerUserID(projectId uint64, arrayCustomerUserID []string) (map[string]string, int) {
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

func (pg *Postgres) GetUserBySegmentAnonymousId(projectId uint64, segAnonId string) (*model.User, int) {
	db := C.GetServices().Db

	var users []model.User
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

// GetAllUserIDByCustomerUserID returns all users with same customer_user_id
func (pg *Postgres) GetAllUserIDByCustomerUserID(projectID uint64, customerUserID string) ([]string, int) {
	if projectID == 0 || customerUserID == "" {
		return nil, http.StatusBadRequest
	}

	db := C.GetServices().Db

	var users []model.User
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
func (pg *Postgres) CreateOrGetSegmentUser(projectId uint64, segAnonId, custUserId string, requestTimestamp int64) (*model.User, int) {
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
	if segAnonId != "" {
		user, errCode = pg.GetUserBySegmentAnonymousId(projectId, segAnonId)
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
			user, errCode = pg.GetUserLatestByCustomerUserId(projectId, custUserId)
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
		if custUserId != "" {
			cUser.CustomerUserId = custUserId
		}

		user, err := pg.createUserWithError(cUser)
		if err != nil {
			// Get and return error is duplicate error.
			if U.IsPostgresUniqueIndexViolationError(
				uniqueIndexProjectIdSegmentAnonymousId, err) {

				user, errCode := pg.GetUserBySegmentAnonymousId(projectId, segAnonId)
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
		logCtx.Warn("Different customer_user_id seen for existing user with segment_anonymous_id.")
	}

	// provided and fetched c_uid are same.
	return user, http.StatusOK
}

func (pg *Postgres) GetUserIDByAMPUserID(projectId uint64, ampUserId string) (string, int) {
	logCtx := log.WithField("project_id", projectId).WithField(
		"amp_user_id", ampUserId)

	userID, errCode := model.GetCacheUserIDByAMPUserID(projectId, ampUserId)
	if errCode == http.StatusFound {
		return userID, errCode
	}

	db := C.GetServices().Db
	var user model.User
	err := db.Limit(1).Where("project_id = ? AND amp_user_id = ?",
		projectId, ampUserId).Select("id").Find(&user).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return "", http.StatusNotFound
		}

		logCtx.Error("Failed to get user by amp_user_id")
		return "", http.StatusInternalServerError
	}

	if user.ID == "" {
		return "", http.StatusNotFound
	}

	model.SetCacheUserIDByAMPUserID(projectId, ampUserId, user.ID)

	return user.ID, http.StatusFound
}

func (pg *Postgres) CreateOrGetAMPUser(projectId uint64, ampUserId string, timestamp int64) (string, int) {
	if projectId == 0 || ampUserId == "" {
		return "", http.StatusBadRequest
	}

	logCtx := log.WithField("project_id",
		projectId).WithField("amp_user_id", ampUserId)

	userID, errCode := pg.GetUserIDByAMPUserID(projectId, ampUserId)
	if errCode == http.StatusInternalServerError {
		return "", errCode
	}

	if errCode == http.StatusFound {
		return userID, errCode
	}

	user, err := pg.createUserWithError(&model.User{ProjectId: projectId,
		AMPUserId: ampUserId, JoinTimestamp: timestamp})
	if err != nil {
		// Get and return error is duplicate error.
		if U.IsPostgresUniqueIndexViolationError(uniqueIndexProjectIdAmpUserId, err) {
			userID, errCode := pg.GetUserIDByAMPUserID(projectId, ampUserId)
			if errCode != http.StatusFound {
				return "", errCode
			}

			return userID, http.StatusFound
		}

		logCtx.WithError(err).Error(
			"Failed to create user by amp user id on CreateOrGetAMPUser")
		return "", http.StatusInternalServerError
	}
	userID = user.ID

	return userID, http.StatusCreated
}

// GetRecentUserPropertyKeysWithLimits This method gets all the recent 'limit' property keys from DB for a given project
func (pg *Postgres) GetRecentUserPropertyKeysWithLimits(projectID uint64, usersLimit int, propertyLimit int, seedDate time.Time) ([]U.Property, error) {
	properties := make([]U.Property, 0)
	db := C.GetServices().Db
	startTime := seedDate.AddDate(0, 0, -7).Unix()
	endTime := seedDate.Unix()
	logCtx := log.WithField("project_id", projectID)

	var queryParams []interface{}
	queryStmnt := "WITH recent_user_events AS (SELECT DISTINCT ON(user_id) user_id, user_properties, timestamp FROM events" + " " +
		"WHERE project_id = ? AND timestamp > ? AND timestamp <= ? ORDER BY user_id, timestamp DESC LIMIT ?)" + " " +
		"SELECT json_object_keys(user_properties::json) AS key, COUNT(*) AS count, MAX(timestamp) as last_seen FROM recent_user_events" + " " +
		"WHERE user_properties != 'null' GROUP BY key ORDER BY count DESC LIMIT ?;"

	queryParams = make([]interface{}, 0, 0)
	queryParams = append(queryParams, projectID, startTime, endTime, usersLimit, propertyLimit)

	rows, err := db.Raw(queryStmnt, queryParams...).Rows()
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
func (pg *Postgres) GetRecentUserPropertyValuesWithLimits(projectID uint64, propertyKey string, usersLimit, valuesLimit int, seedDate time.Time) ([]U.PropertyValue, string, error) {
	// limit on values returned.
	values := make([]U.PropertyValue, 0, 0)
	startTime := seedDate.AddDate(0, 0, -7).Unix()
	endTime := seedDate.Unix()
	db := C.GetServices().Db

	var queryParams []interface{}
	queryStmnt := " WITH recent_user_events AS (SELECT DISTINCT ON(user_id) user_id, user_properties, timestamp FROM events" + " " +
		"WHERE project_id = ? AND timestamp > ? AND timestamp <= ? ORDER BY user_id, timestamp DESC LIMIT ?)" + " " +
		"SELECT user_properties->? AS value, COUNT(*) AS count, MAX(timestamp) AS last_seen, MAX(jsonb_typeof(user_properties->?)) AS value_type FROM recent_user_events" + " " +
		"WHERE user_properties != 'null' AND user_properties->? IS NOT NULL GROUP BY value limit ?;"

	queryParams = make([]interface{}, 0, 0)
	queryParams = append(queryParams, projectID, startTime, endTime, usersLimit, propertyKey, propertyKey, propertyKey, valuesLimit)

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

// Gets userProperties - sorted by count and time. Update list with required ones.
func (pg *Postgres) GetRequiredUserPropertiesByProject(projectID uint64, limit int, lastNDays int) (map[string][]string, map[string]string, error) {
	properties, err := pg.GetUserPropertiesByProject(projectID, 2500, C.GetLookbackWindowForEventUserCache())
	if err != nil {
		return properties, make(map[string]string), err
	}

	// We defined few properties. It needs to be classified into right category.
	// add mandatory properties And remove unnecessary properties.
	properties = U.ClassifyDateTimePropertyKeys(&properties)
	U.FillMandatoryDefaultUserProperties(&properties)
	U.FilterDisabledCoreUserProperties(&properties)

	// Adding Property To Displayname Hash.
	resultantPropertyToDisplayName := make(map[string]string)

	_, userPropertiesToDisplayNames := pg.GetDisplayNamesForAllUserProperties(projectID)
	standardUserPropertiesToDisplayNames := U.STANDARD_USER_PROPERTIES_DISPLAY_NAMES
	_, crmSpecificPropertiesToDisplayNames := pg.GetDisplayNamesForObjectEntities(projectID)
	for property, displayName := range userPropertiesToDisplayNames {
		resultantPropertyToDisplayName[property] = displayName
	}
	for property, displayName := range standardUserPropertiesToDisplayNames {
		resultantPropertyToDisplayName[property] = displayName
	}
	for property, displayName := range crmSpecificPropertiesToDisplayNames {
		resultantPropertyToDisplayName[property] = displayName
	}
	return properties, resultantPropertyToDisplayName, nil
}

// GetUserPropertiesByProject This method iterates over n days and gets user properties from cache for a given project
// Picks all past 24 hrs seen properties and sorts the remaining by count and returns top 'limit'
func (pg *Postgres) GetUserPropertiesByProject(projectID uint64, limit int, lastNDays int) (map[string][]string, error) {
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

	propertyDetails, propertyDetailsStatus := pg.GetAllPropertyDetailsByProjectID(projectID, "", true)

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

//GetPropertyValuesByUserProperty This method iterates over n days and gets user property values from cache for a given project/property
// Picks all past 24 hrs seen values and sorts the remaining by count and returns top 'limit'
func (pg *Postgres) GetPropertyValuesByUserProperty(projectID uint64, propertyName string, limit int, lastNDays int) ([]string, error) {
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

func (pg *Postgres) GetLatestUserPropertiesOfUserAsMap(projectId uint64, id string) (*map[string]interface{}, int) {
	logCtx := log.WithField("project_id", projectId).WithField("id", id)

	user, errCode := pg.GetUser(projectId, id)
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
func (pg *Postgres) GetDistinctCustomerUserIDSForProject(projectID uint64) ([]string, int) {
	logCtx := log.WithFields(log.Fields{"ProjectID": projectID})
	db := C.GetServices().Db

	var customerUserIDS []string
	rows, err := db.Model(&model.User{}).Where("project_id = ? AND customer_user_id IS NOT NULL", projectID).Select("distinct customer_user_id").Rows()
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
func (pg *Postgres) GetUserIdentificationPhoneNumber(projectID uint64, phoneNo string) (string, string) {
	if len(phoneNo) < 5 {
		return "", ""
	}

	pPhoneNo := U.GetPossiblePhoneNumber(phoneNo)
	existingPhoneNo, errCode := pg.GetExistingCustomerUserID(projectID, pPhoneNo)
	if errCode == http.StatusFound {
		for i := range pPhoneNo {
			if userID, exist := existingPhoneNo[pPhoneNo[i]]; exist {
				return pPhoneNo[i], userID
			}
		}
	}

	return phoneNo, ""
}

func (pg *Postgres) FixAllUsersJoinTimestampForProject(db *gorm.DB, projectId uint64, isDryRun bool) error {
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

func (pg *Postgres) GetUserPropertiesByUserID(projectID uint64, id string) (*postgres.Jsonb, int) {
	logCtx := log.WithField("project_id", projectID).WithField("user_id", id)

	if projectID == 0 || id == "" {
		logCtx.Error("Invalid values on arguments.")
		return nil, http.StatusBadRequest
	}

	db := C.GetServices().Db
	var user model.User
	if err := db.Model(&model.User{}).Where("project_id = ? AND id = ?", projectID, id).
		Select("properties").Find(&user).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}

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
func (pg *Postgres) GetUserByPropertyKey(projectID uint64,
	key string, value interface{}) (*model.User, int) {

	logCtx := log.WithField("project_id", projectID).WithField(
		"key", key).WithField("value", value)

	db := C.GetServices().Db
	var user model.User
	// $$$ is a gorm alias for ? jsonb operator.
	err := db.Limit(1).Where("project_id=?", projectID).Where(
		"properties->>? = ?", key, value).Find(&user).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}

		logCtx.WithError(err).Error("Failed to get user id by key.")
		return nil, http.StatusInternalServerError
	}

	return &user, http.StatusFound
}

func (pg *Postgres) UpdateCacheForUserProperties(userId string, projectID uint64,
	updatedProperties map[string]interface{}, redundantProperty bool) {

	// If the cache is empty / cache is updated from more than 1 day - repopulate cache
	logCtx := log.WithFields(log.Fields{
		"project_id": projectID,
	})
	currentTime := U.TimeNowZ()
	currentTimeDatePart := currentTime.Format(U.DATETIME_FORMAT_YYYYMMDD)
	// Store Last updated from DB in cache as a key. and check and refresh cache accordingly
	usersCacheKey, err := model.GetUsersCachedCacheKey(projectID, currentTimeDatePart)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get property cache key - getuserscachedcachekey")
	}

	begin := U.TimeNowZ()
	isNewUser, err := cacheRedis.PFAddPersistent(usersCacheKey, userId, 24*60*60)
	end := U.TimeNowZ()
	metrics.Increment(metrics.IncrNewUserCounter)
	metrics.RecordLatency(metrics.LatencyNewUserCache, float64(end.Sub(begin).Milliseconds()))
	if err != nil {
		logCtx.WithError(err).Error("Failed to get users from cache - getuserscachedcachekey")
	}

	if redundantProperty == true && isNewUser == false {
		return
	}
	analyticsKeysInCache := make([]cacheRedis.SortedSetKeyValueTuple, 0)
	if isNewUser {
		uniqueUsersCountKey, err := model.UserCountAnalyticsCacheKey(
			currentTimeDatePart)
		if err != nil {
			logCtx.WithError(err).Error("Failed to get cache key - uniqueEventsCountKey")
			return
		}
		analyticsKeysInCache = append(analyticsKeysInCache, cacheRedis.SortedSetKeyValueTuple{
			Key:   uniqueUsersCountKey,
			Value: fmt.Sprintf("%v", projectID),
		})

	}
	keysToIncrSortedSet := make([]cacheRedis.SortedSetKeyValueTuple, 0)
	propertiesToIncrSortedSet := make([]cacheRedis.SortedSetKeyValueTuple, 0)
	valuesToIncrSortedSet := make([]cacheRedis.SortedSetKeyValueTuple, 0)
	for property, value := range updatedProperties {
		category := pg.GetPropertyTypeByKeyValue(projectID, "", property, value, true)
		var propertyValue string
		if category == U.PropertyTypeUnknown && reflect.TypeOf(value).Kind() == reflect.Bool {
			category = U.PropertyTypeCategorical
			propertyValue = fmt.Sprintf("%v", value)
		}
		if reflect.TypeOf(value).Kind() == reflect.String {
			propertyValue = value.(string)
		}
		propertyCategoryKeySortedSet, err := model.GetUserPropertiesCategoryByProjectCacheKeySortedSet(projectID, currentTimeDatePart)
		if err != nil {
			logCtx.WithError(err).Error("Failed to get cache key - property category")
			return
		}
		propertiesToIncrSortedSet = append(propertiesToIncrSortedSet, cacheRedis.SortedSetKeyValueTuple{
			Key:   propertyCategoryKeySortedSet,
			Value: fmt.Sprintf("%s:%s", category, property),
		})
		if category == U.PropertyTypeCategorical {
			if propertyValue != "" {
				valueKeySortedSet, err := model.GetValuesByUserPropertyCacheKeySortedSet(projectID, currentTimeDatePart)
				if err != nil {
					logCtx.WithError(err).Error("Failed to get cache key - values")
					return
				}
				valuesToIncrSortedSet = append(valuesToIncrSortedSet, cacheRedis.SortedSetKeyValueTuple{
					Key:   valueKeySortedSet,
					Value: fmt.Sprintf("%s:SS-US-PV:%s", property, propertyValue),
				})
			}
		}
	}
	keysToIncrSortedSet = append(keysToIncrSortedSet, propertiesToIncrSortedSet...)
	keysToIncrSortedSet = append(keysToIncrSortedSet, valuesToIncrSortedSet...)
	begin = U.TimeNowZ()
	_, err = cacheRedis.ZincrPersistentBatch(false, keysToIncrSortedSet...)
	end = U.TimeNowZ()
	metrics.Increment(metrics.IncrUserCacheCounter)
	metrics.RecordLatency(metrics.LatencyUserCache, float64(end.Sub(begin).Milliseconds()))
	if err != nil {
		logCtx.WithError(err).Error("Failed to increment keys")
		return
	}
	if len(analyticsKeysInCache) > 0 {
		_, err = cacheRedis.ZincrPersistentBatch(true, analyticsKeysInCache...)
		if err != nil {
			logCtx.WithError(err).Error("Failed to increment keys")
			return
		}
	}
}

// UpdateUserPropertiesForSession - Updates total user properties and
// latest user properties for session.
func (pg *Postgres) UpdateUserPropertiesForSession(projectID uint64,
	sessionUserPropertiesRecordMap *map[string]model.SessionUserProperties) int {
	return pg.updateUserPropertiesForSessionV2(projectID, sessionUserPropertiesRecordMap)
}

// GetCustomerUserIDAndUserPropertiesFromFormSubmit return customer_user_id na and validated user_properties from form submit properties
func (pg *Postgres) GetCustomerUserIDAndUserPropertiesFromFormSubmit(projectID uint64, userID string,
	formSubmitProperties *U.PropertiesMap) (string, *U.PropertiesMap, int) {

	logCtx := log.WithFields(log.Fields{"project_id": projectID, "user_id": userID})

	user, errCode := pg.GetUser(projectID, userID)
	if errCode != http.StatusFound {
		logCtx.Error("Failed to get latest user properties on fill form submitted properties.")
		return "", nil, http.StatusInternalServerError
	}

	logCtx = logCtx.WithFields(log.Fields{"existing_user_properties": user.Properties,
		"form_event_properties": formSubmitProperties})

	userProperties, err := U.DecodePostgresJsonb(&user.Properties)
	if err != nil {
		logCtx.Error("Failed to decoding latest user properties on fill form submitted properties.")
	}

	formPropertyEmail := U.GetPropertyValueAsString((*formSubmitProperties)[U.UP_EMAIL])
	userPropertyEmail := U.GetPropertyValueAsString((*userProperties)[U.UP_EMAIL])

	formPropertyPhone := U.GetPropertyValueAsString((*formSubmitProperties)[U.UP_PHONE])
	userPropertyPhone := U.GetPropertyValueAsString((*userProperties)[U.UP_PHONE])

	if formPropertyEmail == "" && formPropertyPhone == "" {
		return "", nil, http.StatusBadRequest
	}

	formSubmitUserProperties := model.GetUserPropertiesFromFormSubmitEventProperties(formSubmitProperties)

	orderedIdentifierType := model.GetIdentifierPrecendenceOrderByProjectID(projectID)

	if len(orderedIdentifierType) < 1 {
		logCtx.Error("Failed getting project configured form submit identifiers.")
		return "", nil, http.StatusInternalServerError
	}

	for _, identifierType := range orderedIdentifierType {

		if identifierType == model.IdentificationTypeEmail {

			if formPropertyEmail != "" || userPropertyEmail != "" {
				identity, err := model.GetUpdatedEmailFromFormSubmit(formPropertyEmail, userPropertyEmail)
				if identity != "" {
					if err == model.ErrDifferentEmailSeen {
						logCtx.WithError(err).
							Warn("Different email seen on form event. User property not updated.")
						return identity, formSubmitUserProperties, http.StatusConflict
					}
					return identity, formSubmitUserProperties, http.StatusOK
				}
				return "", nil, http.StatusBadRequest
			}

		} else if identifierType == model.IdentificationTypePhone {

			if formPropertyPhone != "" || userPropertyPhone != "" {
				identity, err := model.GetUpdatedPhoneNoFromFormSubmit(formPropertyPhone, userPropertyPhone)
				if identity != "" {
					if err == model.ErrDifferentPhoneNoSeen {
						logCtx.WithError(err).
							Warn("Different phone seen on form event. User property not updated.")
						return identity, formSubmitUserProperties, http.StatusConflict
					}
					return identity, formSubmitUserProperties, http.StatusOK
				}
				return "", nil, http.StatusBadRequest
			}
		}
	}

	return "", nil, http.StatusBadRequest
}

func (pg *Postgres) updateUserPropertiesForSessionV2(projectID uint64,
	sessionUserPropertiesRecordMap *map[string]model.SessionUserProperties) int {

	logCtx := log.WithField("project_id", projectID)
	latestSessionUserPropertiesByUserID := make(map[string]model.LatestUserPropertiesFromSession, 0)
	sessionUpdateUserIDs := map[string]bool{}

	hasFailure := false
	for eventID, sessionUserProperties := range *sessionUserPropertiesRecordMap {
		logCtx.WithField("event_id", eventID)

		userProperties := sessionUserProperties.EventUserProperties
		isEmptyUserProperties := userProperties == nil || U.IsEmptyPostgresJsonb(userProperties)
		if isEmptyUserProperties {
			logCtx.WithField("user_properties", userProperties).Error("Empty user properties on event.")
			hasFailure = true
			continue
		}

		userPropertiesMap, err := U.DecodePostgresJsonb(userProperties)
		if err != nil {
			logCtx.WithError(err).
				Error("Failed to decode event user properties on UpdateUserPropertiesForSession.")
			hasFailure = true
			continue
		}

		var existingPageCount, existingTotalSpentTime float64
		if existingPageCountValue, exists := (*userPropertiesMap)[U.UP_PAGE_COUNT]; exists {
			existingPageCount, err = U.GetPropertyValueAsFloat64(existingPageCountValue)
			if err != nil {
				logCtx.WithError(err).
					Error("Failed to convert page_count property value as float64.")
			}
		}

		if existingTotalSpentTimeValue, exists := (*userPropertiesMap)[U.UP_TOTAL_SPENT_TIME]; exists {
			existingTotalSpentTime, err = U.GetPropertyValueAsFloat64(existingTotalSpentTimeValue)
			if err != nil {
				logCtx.WithError(err).
					Error("Failed to convert total_page_spent time property value as float64.")
			}
		}

		newPageCount := existingPageCount + sessionUserProperties.SessionPageCount
		newTotalSpentTime := existingTotalSpentTime + sessionUserProperties.SessionPageSpentTime
		newSessionCount := sessionUserProperties.SessionCount

		(*userPropertiesMap)[U.UP_PAGE_COUNT] = newPageCount
		(*userPropertiesMap)[U.UP_TOTAL_SPENT_TIME] = newTotalSpentTime
		(*userPropertiesMap)[U.UP_SESSION_COUNT] = newSessionCount

		userPropertiesJsonb, err := U.EncodeToPostgresJsonb(userPropertiesMap)
		if err != nil {
			logCtx.WithError(err).
				Error("Failed to encode user properties json after adding new session count.")
			hasFailure = true
			continue
		}

		errCode := pg.OverwriteEventUserPropertiesByID(projectID,
			sessionUserProperties.UserID, eventID, userPropertiesJsonb)
		if errCode != http.StatusAccepted {
			logCtx.WithField("err_code", errCode).Error("Failed to overwrite event user properties.")
			hasFailure = true
			continue
		}

		// Latest session based user properties state to be overwritten on
		// latest user_properties_record of the user, if not added already
		latestUserProperties := model.LatestUserPropertiesFromSession{
			PageCount:      newPageCount,
			TotalSpentTime: newTotalSpentTime,
			SessionCount:   newSessionCount,
			Timestamp:      sessionUserProperties.SessionEventTimestamp,
		}
		if _, exists := latestSessionUserPropertiesByUserID[sessionUserProperties.UserID]; !exists {
			latestSessionUserPropertiesByUserID[sessionUserProperties.UserID] = latestUserProperties
		} else {
			if sessionUserProperties.SessionEventTimestamp >
				latestSessionUserPropertiesByUserID[sessionUserProperties.UserID].Timestamp {
				latestSessionUserPropertiesByUserID[sessionUserProperties.UserID] = latestUserProperties
			}
		}

		sessionUpdateUserIDs[sessionUserProperties.UserID] = true
	}

	errCode := pg.updateLatestUserPropertiesForSessionIfNotUpdatedV2(
		projectID,
		sessionUpdateUserIDs,
		&latestSessionUserPropertiesByUserID,
	)
	hasFailure = errCode != http.StatusAccepted

	if hasFailure {
		return http.StatusInternalServerError
	}

	return http.StatusAccepted
}

func (pg *Postgres) updateLatestUserPropertiesForSessionIfNotUpdatedV2(
	projectID uint64,
	sessionUpdateUserIDs map[string]bool,
	latestSessionUserPropertiesByUserID *map[string]model.LatestUserPropertiesFromSession,
) int {

	logCtx := log.WithField("project_id", projectID)

	var hasFailure bool
	for userID := range sessionUpdateUserIDs {
		logCtx = logCtx.WithField("user_id", userID)

		user, errCode := pg.GetUser(projectID, userID)
		if errCode != http.StatusFound {
			logCtx.WithField("err_code", errCode).Error("Failed to get user_properties by user_id.")
			hasFailure = true
			continue
		}

		sessionUserProperties, exists := (*latestSessionUserPropertiesByUserID)[userID]
		if !exists {
			logCtx.Error("Latest session user properties not found for user.")
			hasFailure = true
			continue
		}

		newUserProperties := map[string]interface{}{
			U.UP_TOTAL_SPENT_TIME: sessionUserProperties.TotalSpentTime,
			U.UP_PAGE_COUNT:       sessionUserProperties.PageCount,
			U.UP_SESSION_COUNT:    sessionUserProperties.SessionCount,
		}
		userPropertiesJsonb, err := U.AddToPostgresJsonb(&user.Properties, newUserProperties, true)
		if err != nil {
			logCtx.WithError(err).
				Error("Failed to add new user properites to existing user properites.")
			hasFailure = true
			continue
		}

		errCode = pg.OverwriteUserPropertiesByID(projectID, userID, userPropertiesJsonb, false, 0)
		if errCode != http.StatusAccepted {
			logCtx.WithField("err_code", errCode).Error("Failed to overwrite user properties record.")
			hasFailure = true
			continue
		}
	}

	if hasFailure {
		return http.StatusInternalServerError
	}

	return http.StatusAccepted
}

func shouldAllowCustomerUserID(current, incoming string) bool {
	if current == "" || incoming == "" {
		return false
	}

	if U.IsEmail(current) {
		if U.IsContainsAnySubString(incoming, "@example", "@yahoo", "@gmail") {
			return false
		}
		return true
	}

	if len(incoming) > len(current) &&
		strings.Contains(incoming, current) {
		return true
	}

	return false
}

// UpdateIdentifyOverwriteUserPropertiesMeta adds overwrite information to user properties for debuging purpose. Not available while querying
func (pg *Postgres) UpdateIdentifyOverwriteUserPropertiesMeta(projectID uint64, customerUserID, userID, pageURL, source string, userProperties *postgres.Jsonb, timestamp int64, isNewUser bool) error {
	if projectID == 0 || customerUserID == "" {
		return errors.New("invalid or empty parameter")
	}
	if source == "" {
		return errors.New("source missing")
	}

	logCtx := log.WithFields(log.Fields{"project_id": projectID, "user_id": userID, "customer_user_id": customerUserID})

	var existingUserProperties *map[string]interface{}
	var errCode int
	if !isNewUser {
		existingUserProperties, errCode = pg.GetLatestUserPropertiesOfUserAsMap(projectID, userID)
		if errCode != http.StatusFound {
			logCtx.WithField("err_code", errCode).Error("Failed to get user properties as map.")
			return errors.New("failed to get user properties as map")
		}
	}

	customerUserIDMeta := &model.IdentifyMeta{
		Timestamp: timestamp,
		PageURL:   pageURL,
		Source:    source,
	}

	metaObj, err := model.GetDecodedUserPropertiesIdentifierMetaObject(existingUserProperties)
	if err != nil {
		logCtx.WithError(err).Error("Failed to GetDecodedUserPropertiesIdentifierMetaObject")
		return nil
	}

	(*metaObj)[customerUserID] = *customerUserIDMeta
	return model.UpdateUserPropertiesIdentifierMetaObject(userProperties, metaObj)
}

func (pg *Postgres) CreateGroupUser(user *model.User, groupName, groupID string) (string, int) {

	logCtx := log.WithFields(log.Fields{"project_id": user.ProjectId, "group_name": groupName, "group_id": groupID})
	if user == nil || groupName == "" {
		logCtx.Error("Invalid parameter")
		return "", http.StatusBadRequest
	}

	group, status := pg.GetGroup(user.ProjectId, groupName)
	if status != http.StatusFound {
		if status == http.StatusNotFound {
			logCtx.Error("Group is missing for CreateGroupUser.")
			return "", http.StatusBadRequest
		}

		logCtx.Error("Failed to get group.")
		return "", http.StatusInternalServerError
	}

	groupIndex := fmt.Sprintf("group_%d_id", group.ID)

	isGroupUser := true
	groupUser := &model.User{
		ProjectId:                  user.ProjectId,
		IsGroupUser:                &isGroupUser,
		Properties:                 user.Properties,
		PropertiesUpdatedTimestamp: user.PropertiesUpdatedTimestamp,
		JoinTimestamp:              user.JoinTimestamp,
	}

	processed, _, err := model.SetUserGroupFieldByColumnName(groupUser, groupIndex, groupID)
	if err != nil {
		logCtx.WithError(err).Error("Failed process group id on group user.")
		return "", http.StatusInternalServerError
	}

	if !processed {
		logCtx.WithError(err).Error("Failed to process group_id on group user.")
		return "", http.StatusInternalServerError

	}

	return pg.CreateUser(groupUser)
}

func (pg *Postgres) UpdateUserGroup(projectID uint64, userID, groupName, groupID, groupUserID string) (*model.User, int) {
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "group_name": groupName, "group_id": groupID})
	group, status := pg.GetGroup(projectID, groupName)
	if status != http.StatusFound {
		if status == http.StatusNotFound {
			logCtx.Error("Group is missing.")
			return nil, http.StatusBadRequest
		}

		logCtx.Error("Failed to get group.")
		return nil, http.StatusInternalServerError
	}

	groupIndex := fmt.Sprintf("group_%d_id", group.ID)
	groupUserIndex := fmt.Sprintf("group_%d_user_id", group.ID)

	user, status := pg.GetUser(projectID, userID)
	if status != http.StatusFound {
		return nil, http.StatusInternalServerError
	}

	if user.IsGroupUser != nil && *user.IsGroupUser {
		logCtx.Error("Cannot update group user.")
		return nil, http.StatusBadRequest
	}

	isGroupUser := false
	user.IsGroupUser = &isGroupUser
	processed, IDUpdated, err := model.SetUserGroupFieldByColumnName(user, groupIndex, groupID)
	if err != nil {
		logCtx.WithError(err).Error("Failed to update user by group id.")
		return nil, http.StatusInternalServerError
	}
	if !processed {
		logCtx.Error("Missing tag in struct for group id.")
		return nil, http.StatusInternalServerError
	}

	processed, userIDUpdated, err := model.SetUserGroupFieldByColumnName(user, groupUserIndex, groupUserID)
	if err != nil {
		logCtx.WithError(err).Error("Failed to update user by group id.")
		return nil, http.StatusInternalServerError
	}
	if !processed {
		logCtx.Error("Missing tag in struct for group id.")
		return nil, http.StatusInternalServerError
	}

	if !IDUpdated && !userIDUpdated {
		return nil, http.StatusNotModified
	}

	user.ProjectId = 0
	user.ID = ""
	return pg.UpdateUser(projectID, userID, user, user.PropertiesUpdatedTimestamp)
}

func (pg *Postgres) UpdateUserGroupProperties(projectID uint64, userID string,
	newProperties *postgres.Jsonb, updateTimestamp int64) (*postgres.Jsonb, int) {
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "user_id": userID,
		"new_user_properties": newProperties, "update_timestamp": updateTimestamp})

	if projectID == 0 || userID == "" || newProperties == nil {
		logCtx.Error("Invalid parameters.")
		return nil, http.StatusBadRequest
	}

	user, errCode := pg.GetUser(projectID, userID)
	if errCode != http.StatusFound {
		logCtx.Error("Failed to get user on UpdateUserGroupProperties.")
		return nil, http.StatusInternalServerError
	}

	incomingProperties, err := util.DecodePostgresJsonb(newProperties)
	if err != nil {
		logCtx.WithError(err).Error("Failed to decode user properties on UpdateUserGroupProperties.")
		return nil, http.StatusInternalServerError
	}

	existingProperties, err := util.DecodePostgresJsonb(&user.Properties)
	if err != nil {
		logCtx.WithField("exsting_user_properties", user.Properties).WithError(err).Error("Failed to decode user properties on UpdateUserGroupProperties.")
		return nil, http.StatusInternalServerError
	}

	overWrite := updateTimestamp >= user.PropertiesUpdatedTimestamp

	mergedProperties := make(map[string]interface{})
	for key, value := range *existingProperties {
		mergedProperties[key] = value
	}

	for key, value := range *incomingProperties {
		if value == nil {
			continue
		}

		if _, exist := mergedProperties[key]; exist {
			if overWrite {
				mergedProperties[key] = value
			}
			continue
		}
		mergedProperties[key] = value
	}

	mergedPropertiesJSON, err := U.EncodeToPostgresJsonb(&mergedProperties)
	if err != nil {
		logCtx.WithError(err).Error("Failed to marshal group user properties.")
		return nil, http.StatusInternalServerError
	}

	var newUpdateTimestamp int64
	if overWrite {
		newUpdateTimestamp = updateTimestamp
	} else {
		newUpdateTimestamp = user.PropertiesUpdatedTimestamp
	}

	errCode = pg.OverwriteUserPropertiesByID(projectID, user.ID, mergedPropertiesJSON, true, newUpdateTimestamp)
	if errCode == http.StatusInternalServerError || errCode == http.StatusBadRequest {
		logCtx.WithField("err_code", errCode).WithField("user_id", user.ID).Error("Failed to update user properties on group user.")
		return nil, http.StatusInternalServerError
	}

	return mergedPropertiesJSON, http.StatusAccepted
}
