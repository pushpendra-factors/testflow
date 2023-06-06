package memsql

import (
	"database/sql"
	"encoding/base64"
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
	logFields := log.Fields{
		"err": err,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	return err.Error() == constraintViolationError
}

// createUserWithError - Returns error during create to match
// with constraint errors.
func (store *MemSQL) createUserWithError(user *model.User) (*model.User, error) {
	logFields := log.Fields{
		"user": user,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)

	if user.ProjectId == 0 {
		logCtx.Error("Failed to create user. ProjectId not provided.")
		return nil, errors.New("invalid project_id")
	}

	if user.Source == nil {
		logCtx.Error("Failed to create user. User source not provided.")
		return nil, errors.New("user source missing")
	}

	allProjects, projectIDsMap, _ := C.GetProjectsFromListWithAllProjectSupport(C.GetConfig().CaptureSourceInUsersTable, "")
	if !allProjects && !projectIDsMap[user.ProjectId] {
		user.Source = nil
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
		if C.AllowIdentificationOverwriteUsingSource(user.ProjectId) {
			user.CustomerUserIdSource = user.Source
		}
	}

	// removing U.UP_SESSION_COUNT, from user properties.
	delete(newUserProperties, U.UP_SESSION_COUNT)

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

	// Log for analysis.
	log.WithField("project_id", user.ProjectId).WithField("tag", "create_user").Info("Created user.")

	properties, errCode := store.UpdateUserProperties(user.ProjectId, user.ID, newUserPropertiesJsonb, user.JoinTimestamp)
	if errCode == http.StatusInternalServerError {
		return nil, errors.New("failed to update user properties")
	}

	if properties != nil {
		user.Properties = *properties
	}

	return user, nil
}

func (store *MemSQL) CreateUser(user *model.User) (string, int) {
	logFields := log.Fields{
		"user": user,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)

	if user.SegmentAnonymousId != "" || user.AMPUserId != "" {
		// Corresponding create methods should be used for
		// users from different platform.
		logCtx.Error("Unsupported user of create_user method.")
		return "", http.StatusBadRequest
	}

	newUser, err := store.createUserWithError(user)
	if err == nil {
		return newUser.ID, http.StatusCreated
	}

	if IsDuplicateRecordError(err) {
		if user.ID != "" {
			// Multiple requests trying to create user at the
			// same time should not lead failure permanently,
			// so get the user and return.
			return user.ID, http.StatusCreated
		}

		logCtx.WithError(err).Error("Failed to create user. Integrity violation.")
		return "", http.StatusNotAcceptable
	}

	logCtx.WithError(err).Error("Failed to create user.")
	return "", http.StatusInternalServerError
}

// UpdateUser updates user fields by Id.
func (store *MemSQL) UpdateUser(projectId int64, id string,
	user *model.User, updateTimestamp int64) (*model.User, int) {
	logFields := log.Fields{
		"project_id":       projectId,
		"id":               id,
		"user":             user,
		"update_timestamp": updateTimestamp,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

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

	if user.CustomerUserId != "" {
		props, err := U.DecodePostgresJsonb(&userProperties)
		if err != nil {
			log.WithFields(logFields).WithError(err).Error("Failed to Decode user properties in user update.")
			return nil, http.StatusInternalServerError
		}
		propsMap := *props
		propsMap[U.UP_USER_ID] = user.CustomerUserId
		if U.IsEmail(user.CustomerUserId) {
			propsMap[U.UP_EMAIL] = user.CustomerUserId
		}
		propsByte, err := json.Marshal(propsMap)
		if err != nil {
			log.WithFields(logFields).WithError(err).Error("Failed to marshal user properties in user update.")
			return nil, http.StatusInternalServerError
		}
		userProperties = postgres.Jsonb{RawMessage: propsByte}
	}

	_, errCode := store.UpdateUserProperties(projectId, id, &userProperties, updateTimestamp)
	if errCode != http.StatusAccepted && errCode != http.StatusNotModified {
		return nil, http.StatusInternalServerError
	}

	return &updatedUser, http.StatusAccepted
}

// UpdateUserProperties only if there is a change in properties values.
func (store *MemSQL) UpdateUserProperties(projectId int64, id string,
	newProperties *postgres.Jsonb, updateTimestamp int64) (*postgres.Jsonb, int) {
	logFields := log.Fields{
		"project_id":       projectId,
		"id":               id,
		"new_properties":   newProperties,
		"update_timestamp": updateTimestamp,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	return store.UpdateUserPropertiesV2(projectId, id, newProperties, updateTimestamp, "", "")
}

func (store *MemSQL) IsUserExistByID(projectID int64, id string) int {
	logFields := log.Fields{
		"project_id": projectID,
		"id":         id,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)

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

func (store *MemSQL) GetUser(projectId int64, id string) (*model.User, int) {
	params := log.Fields{"project_id": projectId, "user_id": id}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &params)
	logCtx := log.WithFields(params)

	var user model.User
	db := C.GetServices().Db
	if err := db.Limit(1).Where("project_id = ?", projectId).
		Where("id = ?", id).Find(&user).Error; err != nil {

		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}

		logCtx.WithError(err).Error("Failed to get user using user_id")
		return nil, http.StatusInternalServerError
	}

	return &user, http.StatusFound
}

func (store *MemSQL) GetUsers(projectId int64, offset uint64, limit uint64) ([]model.User, int) {
	logFields := log.Fields{
		"project_id": projectId,
		"offset":     offset,
		"limit":      limit,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

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
func (store *MemSQL) GetUsersByCustomerUserID(projectID int64, customerUserID string) ([]model.User, int) {
	logFields := log.Fields{
		"project_id":       projectID,
		"customer_user_id": customerUserID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)

	var users []model.User
	db := C.GetServices().Db
	if err := db.Where("project_id = ? AND customer_user_id = ?", projectID, customerUserID).
		Find(&users).Error; err != nil {

		logCtx.WithError(err).Error("Failed to get users for customer_user_id")
		return nil, http.StatusInternalServerError
	}
	if len(users) == 0 {
		return nil, http.StatusNotFound
	}

	// Sort by created_at ASC
	sort.Slice(users, func(i, j int) bool {
		return users[i].CreatedAt.Before(users[j].CreatedAt)
	})

	return users, http.StatusFound
}

// GetSelectedUsersByCustomerUserID gets selected (top 50 & bottom 50) users identified by given customer_user_id in increasing order of updated_at.
func (store *MemSQL) GetSelectedUsersByCustomerUserID(projectID int64, customerUserID string, limit uint64, numUsers uint64) ([]model.User, int) {
	logFields := log.Fields{
		"project_id":       projectID,
		"customer_user_id": customerUserID,
		"limit":            limit,
		"num_users":        numUsers,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db
	logCtx := log.WithFields(logFields)

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
	if err := db.Where("project_id = ? AND id IN ( ? )", projectID, userIDs).Find(&users).Error; err != nil {
		logCtx.WithError(err).Error("Failed to get selected users for id")
		return nil, http.StatusInternalServerError
	}

	// sort by created_at ASC
	sort.Slice(users, func(i, j int) bool {
		return users[i].CreatedAt.Before(users[j].CreatedAt)
	})

	return users, http.StatusFound
}

func (store *MemSQL) getLatestUserIDByCustomerUserId(projectId int64,
	customerUserId string, requestSource int) (*model.User, int) {

	logFields := log.Fields{
		"project_id":       projectId,
		"customer_user_id": customerUserId,
		"request_source":   requestSource,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	var user model.User
	db := C.GetServices().Db
	if !C.CheckRestrictReusingUsersByCustomerUserId(projectId) {
		if err := db.Limit(1).Select("id").Order("created_at DESC").Where("project_id = ?", projectId).
			Where("customer_user_id = ?", customerUserId).Find(&user).Error; err != nil {
			if gorm.IsRecordNotFoundError(err) {
				return nil, http.StatusNotFound
			}
			return nil, http.StatusInternalServerError
		}
	} else {
		var userSourceWhereCondition string
		if requestSource == model.UserSourceWeb {
			userSourceWhereCondition = "source = ? OR source IS NULL"
		} else {
			userSourceWhereCondition = "source = ?"
		}
		if err := db.Limit(1).Select("id").Order("created_at DESC").Where("project_id = ?", projectId).
			Where("customer_user_id = ?", customerUserId).Where(userSourceWhereCondition, requestSource).
			Find(&user).Error; err != nil {
			if gorm.IsRecordNotFoundError(err) {
				return nil, http.StatusNotFound
			}
			return nil, http.StatusInternalServerError
		}
	}
	return &user, http.StatusFound
}

// GetUserLatestByCustomerUserId - Gets latest user's id first to avoid huge
// sorting with all columns. and uses the users.id to fetch all columns required.
func (store *MemSQL) GetUserLatestByCustomerUserId(projectId int64,
	customerUserId string, requestSource int) (*model.User, int) {
	user, status := store.getLatestUserIDByCustomerUserId(projectId, customerUserId, requestSource)
	if status != http.StatusFound {
		return nil, status
	}

	return store.GetUser(projectId, user.ID)
}

func (store *MemSQL) GetExistingUserByCustomerUserID(projectId int64, arrayCustomerUserID []string, source ...int) (map[string]string, int) {
	logFields := log.Fields{
		"project_id":             projectId,
		"array_customer_user_id": arrayCustomerUserID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	customerUserIDMap := make(map[string]string)
	if len(arrayCustomerUserID) == 0 {
		return nil, http.StatusBadRequest
	}

	db := C.GetServices().Db
	queryStmnt := "SELECT" + " " + "DISTINCT(customer_user_id), id" + " FROM " + "users" + " WHERE " + "project_id = ? AND customer_user_id IN ( ? )"
	queryParams := []interface{}{projectId, arrayCustomerUserID}

	sourceStmnt := ""
	sourceParams := []interface{}{}
	if len(source) == 1 {
		sourceStmnt = " source = ? "
		if source[0] == model.UserSourceWeb {
			sourceStmnt = " ( source = ? OR source is null ) "
		}
		sourceParams = append(sourceParams, source[0])
	} else if len(source) > 1 {
		sourceStmnt = " source IN (?) "
		for i := range source {
			if source[i] == model.UserSourceWeb {
				sourceStmnt = " ( source IN (?) OR source is null ) "
				break
			}
		}
		sourceParams = append(sourceParams, source)
	}

	if sourceStmnt != "" {
		queryStmnt = queryStmnt + " AND " + sourceStmnt
		queryParams = append(queryParams, sourceParams...)
	}

	rows, err := db.Raw(queryStmnt, queryParams...).Rows()
	if err != nil {
		log.WithError(err).Error("Failed to get customer_user_id.")
		return nil, http.StatusInternalServerError
	}

	defer rows.Close()

	for rows.Next() {
		var customerUserID string
		var userID string
		if err := rows.Scan(&customerUserID, &userID); err != nil {
			log.WithError(err).Error("Failed scanning rows on GetExistingCustomerUserID")
			return nil, http.StatusInternalServerError
		}
		customerUserIDMap[userID] = customerUserID
	}

	if len(customerUserIDMap) == 0 {
		return nil, http.StatusNotFound
	}

	return customerUserIDMap, http.StatusFound
}

func (store *MemSQL) GetUserBySegmentAnonymousId(projectId int64, segAnonId string) (*model.User, int) {
	logFields := log.Fields{
		"project_id":  projectId,
		"seg_anon_id": segAnonId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

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

// GetAllUserIDByCustomerUserID returns all users with same customer_user_id
func (store *MemSQL) GetAllUserIDByCustomerUserID(projectID int64, customerUserID string) ([]string, int) {
	logFields := log.Fields{
		"project_id":       projectID,
		"customer_user_id": customerUserID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

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

func getUserIDByAmpUserID(ampUserID string) string {
	if ampUserID == "" {
		return ""
	}

	return "amp" + "-" + ampUserID
}

func getUserIDBySegementAnonymousID(segAnonID string) string {
	if segAnonID == "" {
		return ""
	}

	return "seg" + "-" + segAnonID
}

func getDomainUserIDByDomainGroupIndexANDDomainName(domainGroupIndex int, domainName string) string {
	if domainGroupIndex == 0 || domainName == "" {
		return ""
	}

	key := fmt.Sprintf("%d-%s", domainGroupIndex, domainName)

	enKey := base64.StdEncoding.EncodeToString([]byte(key))
	return "dom-" + enKey
}

// CreateOrGetSegmentUser create or updates(c_uid) and returns user by segement_anonymous_id
// and/or customer_user_id.
func (store *MemSQL) CreateOrGetSegmentUser(projectId int64, segAnonId, custUserId string,
	requestTimestamp int64, requestSource int) (*model.User, int) {
	logFields := log.Fields{
		"project_id":        projectId,
		"cust_user_id":      custUserId,
		"seg_anon_id":       segAnonId,
		"request_timestamp": requestTimestamp,
		"request_source":    requestSource,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)

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
			user, errCode = store.GetUserLatestByCustomerUserId(projectId, custUserId, requestSource)
			if errCode == http.StatusFound {
				return user, http.StatusOK
			}

			if errCode == http.StatusInternalServerError {
				logCtx.WithField("err_code", errCode).Error(
					"Failed to fetching user with segment provided c_uid.")
				return nil, errCode
			}
		}

		cUser := &model.User{ProjectId: projectId, JoinTimestamp: requestTimestamp, Source: &requestSource}
		// add seg_aid, if provided and not exist already.
		if segAnonId != "" {
			cUser.SegmentAnonymousId = segAnonId
			if C.AllowMergeAmpIDAndSegmentIDWithUserIDByProjectID(projectId) {
				cUser.ID = getUserIDBySegementAnonymousID(segAnonId)
			}
		}
		if custUserId != "" {
			cUser.CustomerUserId = custUserId
		}

		user, err := store.createUserWithError(cUser)
		if err != nil {
			if C.AllowMergeAmpIDAndSegmentIDWithUserIDByProjectID(projectId) {
				if IsDuplicateRecordError(err) {
					if cUser.ID != "" {
						// Multiple requests trying to create user at the
						// same time should not lead failure permanently,
						// so get the user and return.
						return cUser, http.StatusCreated
					}

					logCtx.WithError(err).Error("Failed to create segment user. Integrity violation.")
					return nil, http.StatusNotAcceptable
				}
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
	if user.CustomerUserId != "" && (user.CustomerUserId != custUserId) {
		logCtx.Warn("Different customer_user_id seen for existing user with segment_anonymous_id.")
	}

	// provided and fetched c_uid are same.
	return user, http.StatusOK
}

// CreateOrGetDomainGroupUser creates or get domain group user by group name and domain name
func (store *MemSQL) CreateOrGetDomainGroupUser(projectID int64, groupName string, domainName string,
	requestTimestamp int64, requestSource int) (string, int) {
	logFields := log.Fields{
		"project_id":        projectID,
		"group_name":        groupName,
		"domain_name":       domainName,
		"request_timestamp": requestTimestamp,
		"request_source":    requestSource,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)

	if projectID == 0 || groupName == "" || domainName == "" {
		logCtx.Error("Invalid parameters.")
		return "", http.StatusBadRequest
	}

	user, status := store.GetGroupUserByGroupID(projectID, groupName, domainName)
	if status != http.StatusFound && status != http.StatusNotFound {
		logCtx.Error("Failed to check for exisistence of group user.")
		return "", http.StatusInternalServerError
	}

	if status == http.StatusFound {
		return user.ID, http.StatusFound
	}

	group, status := store.GetGroup(projectID, groupName)
	if status != http.StatusFound && status != http.StatusNotFound {
		logCtx.Error("Failed to get on create group user.")
		return "", http.StatusInternalServerError
	}

	isGroupUser := true
	userID, status := store.CreateGroupUser(&model.User{
		ID:            getDomainUserIDByDomainGroupIndexANDDomainName(group.ID, domainName),
		ProjectId:     projectID,
		IsGroupUser:   &isGroupUser,
		JoinTimestamp: requestTimestamp,
		Source:        &requestSource,
	}, groupName, domainName)
	if status != http.StatusCreated && status != http.StatusConflict {
		logCtx.WithFields(log.Fields{"err_code": status}).Error("Failed to create domain group user.")
		return "", http.StatusInternalServerError
	}

	return userID, status
}

func (store *MemSQL) GetUserIDByAMPUserID(projectId int64, ampUserId string) (string, int) {
	logFields := log.Fields{
		"project_id":  projectId,
		"amp_user_id": ampUserId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

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

		logCtx.WithError(err).Error("Failed to get user by amp_user_id")
		return "", http.StatusInternalServerError
	}

	if user.ID == "" {
		return "", http.StatusNotFound
	}

	model.SetCacheUserIDByAMPUserID(projectId, ampUserId, user.ID)

	return user.ID, http.StatusFound
}

func (store *MemSQL) CreateOrGetAMPUser(projectId int64, ampUserId string, timestamp int64, requestSource int) (string, int) {
	logFields := log.Fields{
		"project_id":     projectId,
		"amp_user_id":    ampUserId,
		"timestamp":      timestamp,
		"request_source": requestSource,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	if projectId == 0 || ampUserId == "" {
		return "", http.StatusBadRequest
	}

	logCtx := log.WithFields(logFields)

	// Unique (project_id, amp_user_id) constraint.
	userID, errCode := store.GetUserIDByAMPUserID(projectId, ampUserId)
	if errCode == http.StatusInternalServerError {
		return "", errCode
	}
	if errCode == http.StatusFound {
		return userID, errCode
	}

	cUser := model.User{ProjectId: projectId,
		AMPUserId:     ampUserId,
		JoinTimestamp: timestamp,
		Source:        &requestSource,
	}

	if C.AllowMergeAmpIDAndSegmentIDWithUserIDByProjectID(projectId) {
		cUser.ID = getUserIDByAmpUserID(ampUserId)
	}

	user, err := store.createUserWithError(&cUser)
	if err != nil {
		if C.AllowMergeAmpIDAndSegmentIDWithUserIDByProjectID(projectId) {
			if IsDuplicateRecordError(err) {
				if cUser.ID != "" {
					// Multiple requests trying to create user at the
					// same time should not lead failure permanently,
					// so get the user and return.
					return cUser.ID, http.StatusCreated
				}

				logCtx.WithError(err).Error("Failed to create amp user. Integrity violation.")
				return "", http.StatusNotAcceptable
			}
		}

		logCtx.WithError(err).Error(
			"Failed to create user by amp user id on CreateOrGetAMPUser")
		return "", http.StatusInternalServerError
	}

	return user.ID, http.StatusCreated
}

// GetRecentUserPropertyKeysWithLimits This method gets all the recent 'limit' property keys from DB for a given project
func (store *MemSQL) GetRecentUserPropertyKeysWithLimits(projectID int64, usersLimit int, propertyLimit int, seedDate time.Time) ([]U.Property, error) {
	logFields := log.Fields{
		"project_id":     projectID,
		"users_limit":    usersLimit,
		"property_limit": propertyLimit,
		"seed_date":      seedDate,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	properties := make([]U.Property, 0)
	db := C.GetServices().Db
	startTime := seedDate.AddDate(0, 0, -7).Unix()
	endTime := seedDate.Unix()
	logCtx := log.WithFields(logFields)

	var queryParams []interface{}
	queryStmnt := fmt.Sprintf("WITH recent_user_events AS (SELECT user_id, FIRST(user_properties, FROM_UNIXTIME(events.timestamp)) AS user_properties, FIRST(timestamp, FROM_UNIXTIME(events.timestamp)) AS timestamp FROM events"+" "+
		"WHERE project_id = ? AND timestamp > ? AND timestamp <= ? GROUP BY user_id ORDER BY user_id, timestamp DESC LIMIT %d)"+" "+
		"SELECT user_properties, timestamp as last_seen FROM recent_user_events"+" "+
		"WHERE user_properties != 'null' AND user_properties IS NOT NULL;", usersLimit)

	queryParams = make([]interface{}, 0, 0)
	queryParams = append(queryParams, projectID, startTime, endTime)

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
func (store *MemSQL) GetRecentUserPropertyValuesWithLimits(projectID int64, propertyKey string,
	usersLimit, valuesLimit int, seedDate time.Time) ([]U.PropertyValue, string, error) {
	logFields := log.Fields{
		"project_id":   projectID,
		"users_limit":  usersLimit,
		"property_key": propertyKey,
		"seed_date":    seedDate,
		"values_limit": valuesLimit,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

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

	logCtx := log.WithFields(logFields)
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

// Gets userProperties - sorted by count and time. Update list with required ones.
func (store *MemSQL) GetRequiredUserPropertiesByProject(projectID int64, limit int, lastNDays int) (map[string][]string, map[string]string, error) {
	logFields := log.Fields{
		"project_id":  projectID,
		"limit":       limit,
		"last_n_days": lastNDays,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	properties, err := store.GetUserPropertiesByProject(projectID, 2500, lastNDays)
	if err != nil {
		return properties, make(map[string]string), err
	}

	// We defined few properties. It needs to be classified into right category.
	// add mandatory properties And remove unnecessary properties.
	properties = U.ClassifyDateTimePropertyKeys(&properties)
	U.FillMandatoryDefaultUserProperties(&properties)
	_, overrides := store.GetPropertyOverridesByType(projectID, U.PROPERTY_OVERRIDE_BLACKLIST, model.GetEntity(true))
	U.FilterDisabledCoreUserProperties(overrides, &properties)

	// Adding Property To Displayname Hash.
	resultantPropertyToDisplayName := make(map[string]string)

	_, userPropertiesToDisplayNames := store.GetDisplayNamesForAllUserProperties(projectID)
	standardUserPropertiesToDisplayNames := U.STANDARD_USER_PROPERTIES_DISPLAY_NAMES
	_, crmSpecificPropertiesToDisplayNames := store.GetDisplayNamesForObjectEntities(projectID)
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
func (store *MemSQL) GetUserPropertiesByProject(projectID int64, limit int, lastNDays int) (map[string][]string, error) {
	logFields := log.Fields{
		"project_id":  projectID,
		"limit":       limit,
		"last_n_days": lastNDays,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

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

func getUserPropertiesByProjectFromCache(projectID int64, dateKey string) (U.CachePropertyWithTimestamp, error) {
	logFields := log.Fields{
		"project_id": projectID,
		"date_key":   dateKey,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)
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

// GetPropertyValuesByUserProperty This method iterates over n days and gets user property values
// from cache for a given project/property. Picks all past 24 hrs seen values and sorts the
// remaining by count and returns top 'limit'
func (store *MemSQL) GetPropertyValuesByUserProperty(projectID int64,
	propertyName string, limit int, lastNDays int) ([]string, error) {
	logFields := log.Fields{
		"project_id":    projectID,
		"limit":         limit,
		"last_n_days":   lastNDays,
		"property_name": propertyName,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

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

func getPropertyValuesByUserPropertyFromCache(projectID int64, propertyName string,
	dateKey string) (U.CachePropertyValueWithTimestamp, error) {
	logFields := log.Fields{
		"project_id":    projectID,
		"date_key":      dateKey,
		"property_name": propertyName,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)
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

func (store *MemSQL) GetLatestUserPropertiesOfUserAsMap(projectID int64, id string) (*map[string]interface{}, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"id":         id,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)

	var user model.User
	db := C.GetServices().Db
	if err := db.Limit(1).Where("project_id = ? AND id = ?", projectID, id).
		Select("properties").Find(&user).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}

		logCtx.WithError(err).Error("Failed to get user_properties by id.")
		return nil, http.StatusInternalServerError
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
func (store *MemSQL) GetDistinctCustomerUserIDSForProject(projectID int64) ([]string, int) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)

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
func (store *MemSQL) GetUserIdentificationPhoneNumber(projectID int64, phoneNo string) (string, string) {
	logFields := log.Fields{
		"project_id": projectID,
		"phone_no":   phoneNo,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	if len(phoneNo) < 5 {
		return "", ""
	}

	pPhoneNo := U.GetPossiblePhoneNumber(phoneNo)
	existingPhoneNo, errCode := store.GetExistingUserByCustomerUserID(projectID, pPhoneNo)
	if errCode == http.StatusFound {
		for i := range pPhoneNo {

			for userID := range existingPhoneNo {
				if existingPhoneNo[userID] == pPhoneNo[i] {
					return pPhoneNo[i], userID
				}
			}
		}
	}

	return phoneNo, ""
}

func (store *MemSQL) FixAllUsersJoinTimestampForProject(db *gorm.DB, projectId int64, isDryRun bool) error {
	logFields := log.Fields{
		"project_id": projectId,
		"db":         db,
		"is_dry_run": isDryRun,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	userRows, err := db.Raw("SELECT id, join_timestamp FROM users WHERE project_id = ?", projectId).Rows()
	defer userRows.Close()
	if err != nil {
		log.WithError(err).Error("SQL Query failed.")
		return err
	}
	for userRows.Next() {
		var userId string
		var joinTimestamp int64
		if err = userRows.Scan(&userId, &joinTimestamp); err != nil {
			log.WithError(err).Error("SQL Parse failed.")
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
				rows, err := db.Raw("UPDATE users SET join_timestamp=? WHERE project_id=? AND id=?", newJoinTimestamp, projectId, userId).Rows()
				if err != nil {
					log.WithError(err).Error("Error on update of FixAllUsersJoinTimestampForProject.")
					continue
				}
				defer rows.Close()
				log.Info(fmt.Sprintf("Updated %s", userId))
			}
		}
	}
	return nil
}

func (store *MemSQL) GetUserPropertiesByUserID(projectID int64, id string) (*postgres.Jsonb, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"id":         id,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)

	if projectID == 0 || id == "" {
		logCtx.Error("Invalid values on arguments.")
		return nil, http.StatusBadRequest
	}

	var user model.User
	db := C.GetServices().Db
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
func (store *MemSQL) GetUserByPropertyKey(projectID int64,
	key string, value interface{}) (*model.User, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"key":        key,
		"value":      value,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)

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

func (store *MemSQL) getUsersForMergingPropertiesByCustomerUserID(projectID int64,
	customerUserID string, includeUser *model.User) ([]model.User, int) {
	logFields := log.Fields{
		"project_id":       projectID,
		"customer_user_id": customerUserID,
		"include_user":     includeUser,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)

	if projectID == 0 || customerUserID == "" {
		logCtx.Error("Invalid values for arguments.")
		return []model.User{}, http.StatusBadRequest
	}

	// For user_properties created at same unix time, older user order will help in
	// ensuring the order while merging properties.
	// users, errCode := store.GetUsersByCustomerUserID(projectID, customerUserID)
	var limit uint64 = 10000
	var numUsers uint64 = 100
	users, errCode := store.GetSelectedUsersByCustomerUserID(projectID, customerUserID, limit, numUsers)
	if errCode == http.StatusInternalServerError || errCode == http.StatusNotFound {
		return users, errCode
	}

	usersLength := len(users)
	if usersLength > model.MaxUsersForPropertiesMerge {
		// If number of users to merge are more than max allowed, merge for oldest max/2 and latest max/2.
		users = append(users[0:model.MaxUsersForPropertiesMerge/2],
			users[usersLength-model.MaxUsersForPropertiesMerge/2:usersLength]...)
	}

	if includeUser != nil {
		// include user if not found in the list
		userExistByCustomerUserID := false
		for i := range users {
			if users[i].ID == includeUser.ID {
				userExistByCustomerUserID = true
				break
			}
		}

		if !userExistByCustomerUserID {
			users = append(users, *includeUser)
		}
	}

	metrics.Increment(metrics.IncrUserPropertiesMergeCount)

	return users, http.StatusFound
}

func (store *MemSQL) mergeNewPropertiesWithCurrentUserProperties(projectID int64, userID string,
	currentProperties *postgres.Jsonb, currentUpdateTimestamp int64,
	newProperties *postgres.Jsonb, newUpdateTimestamp int64, source string, objectType string,
) (*postgres.Jsonb, int) {
	logFields := log.Fields{
		"project_id":               projectID,
		"user_id":                  userID,
		"current_properties":       currentProperties,
		"current_update_timestamp": currentUpdateTimestamp,
		"new_properties":           newProperties,
		"new_update_timestamp":     newUpdateTimestamp,
		"source":                   source,
		"object_type":              objectType,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)

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

	overwriteProperties := false
	useSourcePropertyOverwrite := C.UseSourcePropertyOverwriteByProjectIDs(projectID)
	if useSourcePropertyOverwrite {
		overwriteProperties, err = model.CheckForCRMUserPropertiesOverwrite(source, objectType, newPropertiesMap, currentPropertiesMap)
		if err != nil {
			logCtx.WithError(err).Error("Failed to get overwriteProperties flag value.")
		}
	}

	// Overwrite the keys only, if the update is future, else only add new keys.
	if newUpdateTimestamp >= currentUpdateTimestamp {
		mergo.Merge(&mergedPropertiesMap, newPropertiesMap, mergo.WithOverride)
		// For fixing the meta identifier object which was a string earlier and changed to JSON.
		// Mergo doesn't consider change in datatype as change in the same key.
		if _, exists := newPropertiesMap[U.UP_META_OBJECT_IDENTIFIER_KEY]; exists {
			mergedPropertiesMap[U.UP_META_OBJECT_IDENTIFIER_KEY] = newPropertiesMap[U.UP_META_OBJECT_IDENTIFIER_KEY]
		}
	} else {
		if useSourcePropertyOverwrite && model.IsCRMSource(source) {
			for property := range newPropertiesMap {
				if model.IsEmptyPropertyValue(newPropertiesMap[property]) && !U.IsCRMPropertyKey(property) {
					continue
				}

				if U.IsCRMPropertyKeyBySource(source, property) && overwriteProperties {
					mergedPropertiesMap[property] = newPropertiesMap[property]
				} else {
					if _, exists := mergedPropertiesMap[property]; !exists {
						mergedPropertiesMap[property] = newPropertiesMap[property]
					}
				}
			}
		} else {
			mergo.Merge(&mergedPropertiesMap, newPropertiesMap)
		}
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
		logCtx.WithError(err).Error("Failed to marshal new properties merged to current user properties.")
		return nil, http.StatusInternalServerError
	}

	return mergedPropertiesJSON, http.StatusOK
}

// UpdateUserPropertiesV2 - Merge new properties with the existing properties of user and also
// merge the properties of user with same customer_user_id, then updates properties on users table.
func (store *MemSQL) UpdateUserPropertiesV2(projectID int64, id string,
	newProperties *postgres.Jsonb, newUpdateTimestamp int64, sourceValue string, objectType string) (*postgres.Jsonb, int) {
	logFields := log.Fields{
		"project_id":           projectID,
		"id":                   id,
		"new_properties":       newProperties,
		"new_update_timestamp": newUpdateTimestamp,
		"source_value":         sourceValue,
		"object_type":          objectType,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	defer model.LogOnSlowExecutionWithParams(time.Now(), nil)

	logCtx := log.WithFields(logFields)

	newProperties = U.SanitizePropertiesJsonb(newProperties)

	user, errCode := store.GetUser(projectID, id)
	if errCode != http.StatusFound {
		return nil, http.StatusInternalServerError
	}

	newPropertiesMergedJSON, errCode := store.mergeNewPropertiesWithCurrentUserProperties(projectID, id,
		&user.Properties, user.PropertiesUpdatedTimestamp, newProperties, newUpdateTimestamp, sourceValue, objectType)
	if errCode == http.StatusNotModified {
		return &user.Properties, http.StatusNotModified
	}
	if errCode != http.StatusOK {
		logCtx.WithField("err_code", errCode).Error("Failed merging current properties with new properties on update_properties v2.")
		return nil, http.StatusInternalServerError
	}

	// Skip merge by customer_user_id, if customer_user_id is not available.
	if user.CustomerUserId == "" {
		errCode = store.OverwriteUserPropertiesByID(projectID, id, newPropertiesMergedJSON, true, newUpdateTimestamp, sourceValue)
		if errCode == http.StatusInternalServerError || errCode == http.StatusBadRequest {
			return nil, http.StatusInternalServerError
		}

		return newPropertiesMergedJSON, http.StatusAccepted
	}

	users, errCode := store.getUsersForMergingPropertiesByCustomerUserID(projectID, user.CustomerUserId, user)
	if errCode != http.StatusFound {
		logCtx.WithField("err_code", errCode).Error("Failed to get user by customer_user_id for merging user properties.")
		return &user.Properties, http.StatusInternalServerError
	}

	// Skip merge by customer_user_id, if only the current user has the customer_user_id.
	if len(users) == 1 {
		errCode = store.OverwriteUserPropertiesByID(projectID, id, newPropertiesMergedJSON, true, newUpdateTimestamp, sourceValue)
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
			logCtx.WithError(err).WithField("user_id", users[i].ID).Error("Failed to decode existing user_properties.")
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

	mergedByCustomerUserIDMap, errCode := model.MergeUserPropertiesByCustomerUserID(projectID, users, user.CustomerUserId, sourceValue, objectType)
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

		// removing U.UP_SESSION_COUNT, from user properties.
		delete(*mergedPropertiesAfterSkipMap, U.UP_SESSION_COUNT)

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

		errCode = store.OverwriteUserPropertiesByID(projectID, user.ID,
			mergedPropertiesAfterSkipJSON, true, newUpdateTimestamp, sourceValue)
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
func (store *MemSQL) OverwriteUserPropertiesByCustomerUserID(projectID int64,
	customerUserID string, properties *postgres.Jsonb, updateTimestamp int64) int {
	logFields := log.Fields{
		"project_id":       projectID,
		"customer_user_id": customerUserID,
		"properties":       properties,
		"update_timestamp": updateTimestamp,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)

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

	// Log for analysis.
	log.WithField("project_id", projectID).WithField("tag", "update_user").Info("Updated user.")

	return http.StatusAccepted
}

func (store *MemSQL) OverwriteUserPropertiesByIDInBatch(batchedOverwriteUserPropertiesByIDParams []model.OverwriteUserPropertiesByIDParams) bool {
	defer model.LogOnSlowExecutionWithParams(time.Now(), nil)

	db := C.GetServices().Db
	dbTx := db.Begin()
	if dbTx.Error != nil {
		log.WithError(dbTx.Error).Error("Failed to begin transaction in OverwriteUserPropertiesByIDInBatch.")
		return true
	}

	hasFailure := false
	for i := range batchedOverwriteUserPropertiesByIDParams {
		projectID := batchedOverwriteUserPropertiesByIDParams[i].ProjectID
		userID := batchedOverwriteUserPropertiesByIDParams[i].UserID
		userProperties := batchedOverwriteUserPropertiesByIDParams[i].UserProperties
		withUpdateTimestamp := batchedOverwriteUserPropertiesByIDParams[i].WithUpdateTimestamp
		updateTimestamps := batchedOverwriteUserPropertiesByIDParams[i].UpdateTimestamp
		source := batchedOverwriteUserPropertiesByIDParams[i].Source

		status := store.overwriteUserPropertiesByIDWithTransaction(projectID, userID, userProperties,
			withUpdateTimestamp, updateTimestamps, source, dbTx)
		if status != http.StatusAccepted {
			log.WithFields(log.Fields{"overwrite_user_properties_by_id_params": batchedOverwriteUserPropertiesByIDParams[i], "err_code": status}).
				Error("Failed to overwrite user properties in batch using OverwriteUserPropertiesByIDInBatch.")
			hasFailure = true
		}
	}

	err := dbTx.Commit().Error
	if err != nil {
		log.WithError(err).Error("Failed to commit in OverwriteUserPropertiesByIDInBatch.")
		hasFailure = true
	}

	return hasFailure
}

func (store *MemSQL) OverwriteUserPropertiesByID(projectID int64, id string,
	properties *postgres.Jsonb, withUpdateTimestamp bool, updateTimestamp int64, source string) int {
	db := C.GetServices().Db
	return store.overwriteUserPropertiesByIDWithTransaction(projectID, id, properties, withUpdateTimestamp, updateTimestamp, source, db)
}

func (store *MemSQL) overwriteUserPropertiesByIDWithTransaction(projectID int64, id string,
	properties *postgres.Jsonb, withUpdateTimestamp bool, updateTimestamp int64, source string, dbTx *gorm.DB) int {
	logFields := log.Fields{
		"project_id":            projectID,
		"id":                    id,
		"properties":            properties,
		"update_timestamp":      updateTimestamp,
		"with_update_timestamp": withUpdateTimestamp,
		"source":                source,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)

	if projectID == 0 || id == "" {
		logCtx.Error("Failed to overwrite properties. Empty or nil properties.")
		return http.StatusBadRequest
	}

	if properties == nil {
		logCtx.Error("Failed to overwrite properties. Empty or nil properties.")
		return http.StatusBadRequest
	}

	if dbTx == nil {
		logCtx.Error("Missing db method.")
		return http.StatusBadRequest
	}

	if withUpdateTimestamp && updateTimestamp == 0 {
		logCtx.Error("Invalid update_timestamp.")
		return http.StatusBadRequest
	}

	currentPropertiesUpdatedTimestamp, status := store.GetPropertiesUpdatedTimestampOfUser(projectID, id)
	if status != http.StatusFound {
		logCtx.WithField("errr_code", status).Error("Failed to get propertiesUpdatedTimestamp for the user.")
		return http.StatusBadRequest
	}

	// Explicit cleanup for removing unsupported characters.
	properties.RawMessage = U.CleanupUnsupportedCharOnStringBytes(properties.RawMessage)

	update := map[string]interface{}{"properties": properties}
	if updateTimestamp > 0 && updateTimestamp > currentPropertiesUpdatedTimestamp {
		if C.UseSourcePropertyOverwriteByProjectIDs(projectID) {
			if !model.BlacklistUserPropertiesUpdateTimestampBySource[source] {
				update["properties_updated_timestamp"] = updateTimestamp
			}
		} else {
			update["properties_updated_timestamp"] = updateTimestamp
		}
	}

	err := dbTx.Model(&model.User{}).Limit(1).
		Where("project_id = ? AND id = ?", projectID, id).Update(update).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to overwrite user properties.")
		return http.StatusInternalServerError
	}

	// Log for analysis.
	log.WithField("project_id", projectID).WithField("tag", "update_user").Info("Updated user.")

	return http.StatusAccepted
}

func (store *MemSQL) GetPropertiesUpdatedTimestampOfUser(projectId int64, id string) (int64, int) {
	logFields := log.Fields{
		"project_id": projectId,
		"id":         id,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db
	logCtx := log.WithFields(logFields)

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

func (store *MemSQL) UpdateCacheForUserProperties(userId string, projectID int64,
	updatedProperties map[string]interface{}, redundantProperty bool) {
	logFields := log.Fields{
		"project_id":         projectID,
		"user_id":            userId,
		"updated_properties": updatedProperties,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	// If the cache is empty / cache is updated from more than 1 day - repopulate cache
	logCtx := log.WithFields(logFields)
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

	nonGroupProperties := model.FilterGroupPropertiesFromUserProperties(updatedProperties)
	keysToIncrSortedSet := make([]cacheRedis.SortedSetKeyValueTuple, 0)
	propertiesToIncrSortedSet := make([]cacheRedis.SortedSetKeyValueTuple, 0)
	valuesToIncrSortedSet := make([]cacheRedis.SortedSetKeyValueTuple, 0)
	for property, value := range nonGroupProperties {
		category := store.GetPropertyTypeByKeyValue(projectID, "", property, value, true)
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
func (store *MemSQL) UpdateUserPropertiesForSession(projectID int64,
	sessionUserPropertiesRecordMap *map[string]model.SessionUserProperties) int {
	logFields := log.Fields{"project_id": projectID}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	return store.updateUserPropertiesForSessionV2(projectID, sessionUserPropertiesRecordMap)
}

// GetCustomerUserIDAndUserPropertiesFromFormSubmit return customer_user_id na and validated user_properties from form submit properties
func (store *MemSQL) GetCustomerUserIDAndUserPropertiesFromFormSubmit(projectID int64, userID string,
	formSubmitProperties *U.PropertiesMap) (string, *U.PropertiesMap, int) {
	logFields := log.Fields{
		"project_id":             projectID,
		"user_id":                userID,
		"form_submit_properties": formSubmitProperties,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)

	existingUserProperties, errCode := store.GetUserPropertiesByUserID(projectID, userID)
	if errCode != http.StatusFound {
		logCtx.WithField("err_code", errCode).Error("Failed to get latest user properties on fill form submitted properties.")
		return "", nil, http.StatusInternalServerError
	}

	logCtx = logCtx.WithFields(log.Fields{"existing_user_properties": existingUserProperties,
		"form_event_properties": formSubmitProperties})

	userProperties, err := U.DecodePostgresJsonb(existingUserProperties)
	if err != nil {
		logCtx.WithError(err).Error("Failed to decoding latest user properties on fill form submitted properties.")
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

func validateAndLogPageCountAndPageSpentTimeDisparity(logCtx *log.Entry,
	existingUserProperties *postgres.Jsonb,
	newPageCount, newTotalSpentTime float64) {

	existingUserPropertiesMap, err := U.DecodePostgresJsonbAsPropertiesMap(existingUserProperties)
	if err != nil {
		logCtx.WithError(err).Error("Failed to decode existing user_proeprties to validate.")
	}

	existingPageCount, err := U.GetPropertyValueAsFloat64((*existingUserPropertiesMap)[U.UP_PAGE_COUNT])
	if err != nil {
		logCtx.WithError(err).Error("Failed to convert page_count to float64.")
	}
	if existingPageCount > 0 && existingPageCount > newPageCount {
		logCtx.WithFields(log.Fields{
			"existing": existingPageCount,
			"new":      newPageCount,
		}).Warn("Existing value of page_count is greater than new value")
	}

	existingTotalSpentTime, err := U.GetPropertyValueAsFloat64((*existingUserPropertiesMap)[U.UP_TOTAL_SPENT_TIME])
	if err != nil {
		logCtx.WithError(err).Error("Failed to convert existing total_spent_time to float64.")
	}
	if existingTotalSpentTime > 0 && existingTotalSpentTime > newTotalSpentTime {
		logCtx.WithFields(log.Fields{
			"existing": existingTotalSpentTime,
			"new":      newTotalSpentTime,
		}).Warn("Existing value of total_spent_time greater than new value")
	}
}

func (store *MemSQL) updateUserPropertiesForSessionV2(projectID int64,
	sessionUserPropertiesRecordMap *map[string]model.SessionUserProperties) int {
	logFields := log.Fields{"project_id": projectID}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)
	latestSessionUserPropertiesByUserID := make(map[string]model.LatestUserPropertiesFromSession, 0)
	sessionUpdateUserIDs := map[string]bool{}

	hasFailure := false
	for eventID, sessionUserProperties := range *sessionUserPropertiesRecordMap {
		if sessionUserProperties.IsSessionEvent {
			continue
		}
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

		userPropertiesJsonb, err := U.EncodeToPostgresJsonb(userPropertiesMap)
		if err != nil {
			logCtx.WithError(err).
				Error("Failed to encode user properties json after adding new session count.")
			hasFailure = true
			continue
		}

		errCode := store.OverwriteEventUserPropertiesByID(projectID,
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

	store.updateSessionEventUserProperties(projectID, sessionUserPropertiesRecordMap, &latestSessionUserPropertiesByUserID)

	errCode := store.updateLatestUserPropertiesForSessionIfNotUpdatedV2(
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

func (store *MemSQL) updateSessionEventUserProperties(projectID int64, sessionUserPropertiesRecordMap *map[string]model.SessionUserProperties,
	latestSessionUserPropertiesByUserID *map[string]model.LatestUserPropertiesFromSession) {

	logFields := log.Fields{"project_id": projectID}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)
	userIDToChannelProperty := store.getIntitialAndLatestChannelMap(sessionUserPropertiesRecordMap)

	for eventID, sessionUserProperties := range *sessionUserPropertiesRecordMap {
		if !sessionUserProperties.IsSessionEvent {
			continue
		}
		logCtx.WithField("event_id", eventID)

		userProperties := sessionUserProperties.EventUserProperties
		isEmptyUserProperties := userProperties == nil || U.IsEmptyPostgresJsonb(userProperties)
		if isEmptyUserProperties {
			logCtx.WithField("user_properties", userProperties).Error("Empty user properties on event.")
			continue
		}

		userPropertiesMap, err := U.DecodePostgresJsonb(userProperties)
		if err != nil {
			logCtx.WithError(err).
				Error("Failed to decode event user properties on UpdateUserPropertiesForSession.")
			continue
		}

		if channel, exists := (*userPropertiesMap)[U.UP_INITIAL_CHANNEL]; !exists || channel == "" {
			(*userPropertiesMap)[U.UP_INITIAL_CHANNEL] = userIDToChannelProperty[sessionUserProperties.UserID].InitialChannel
		}
		(*userPropertiesMap)[U.UP_LATEST_CHANNEL] = sessionUserProperties.SessionChannel

		latestUserProperties := (*latestSessionUserPropertiesByUserID)[sessionUserProperties.UserID]
		latestUserProperties.InitialChannel = userIDToChannelProperty[sessionUserProperties.UserID].InitialChannel
		latestUserProperties.LatestChannel = userIDToChannelProperty[sessionUserProperties.UserID].LatestChannel
		(*latestSessionUserPropertiesByUserID)[sessionUserProperties.UserID] = latestUserProperties

		userPropertiesJsonb, err := U.EncodeToPostgresJsonb(userPropertiesMap)
		if err != nil {
			logCtx.WithError(err).
				Error("Failed to encode user properties json after adding new session count.")
			continue
		}

		errCode := store.OverwriteEventUserPropertiesByID(projectID,
			sessionUserProperties.UserID, eventID, userPropertiesJsonb)
		if errCode != http.StatusAccepted {
			logCtx.WithField("err_code", errCode).Error("Failed to overwrite event user properties.")
			continue
		}
	}
}

func (store *MemSQL) getIntitialAndLatestChannelMap(sessionUserPropertiesRecordMap *map[string]model.SessionUserProperties) map[string]model.ChannelUserProperty {
	userIDToChannelProperties := make(map[string]model.ChannelUserProperty)
	userIDToSessionUserPropertiesMap := make(map[string][]model.SessionUserProperties)
	for _, sessionUserProperties := range *sessionUserPropertiesRecordMap {
		if !sessionUserProperties.IsSessionEvent {
			continue
		}
		userIDToSessionUserPropertiesMap[sessionUserProperties.UserID] = append(userIDToSessionUserPropertiesMap[sessionUserProperties.UserID], sessionUserProperties)
	}
	for userID, sessionUserPropertiesArr := range userIDToSessionUserPropertiesMap {
		sort.Slice(sessionUserPropertiesArr[:], func(i, j int) bool {
			return sessionUserPropertiesArr[i].SessionEventTimestamp < sessionUserPropertiesArr[j].SessionEventTimestamp
		})
		userIDToChannelProperties[userID] = model.ChannelUserProperty{
			InitialChannel: sessionUserPropertiesArr[0].SessionChannel,
			LatestChannel:  sessionUserPropertiesArr[len(sessionUserPropertiesArr)-1].SessionChannel,
		}
	}
	return userIDToChannelProperties
}

func (store *MemSQL) updateLatestUserPropertiesForSessionIfNotUpdatedV2(
	projectID int64,
	sessionUpdateUserIDs map[string]bool,
	latestSessionUserPropertiesByUserID *map[string]model.LatestUserPropertiesFromSession,
) int {

	logFields := log.Fields{"project_id": projectID}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)

	var hasFailure bool
	overwriteUserPropertiesByIDParamsInBatch := make([]model.OverwriteUserPropertiesByIDParams, 0)

	for userID := range sessionUpdateUserIDs {
		logCtx = logCtx.WithField("user_id", userID)

		existingUserProperties, errCode := store.GetUserPropertiesByUserID(projectID, userID)
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

		validateAndLogPageCountAndPageSpentTimeDisparity(logCtx, existingUserProperties,
			sessionUserProperties.PageCount, sessionUserProperties.TotalSpentTime)

		newUserProperties := map[string]interface{}{
			U.UP_TOTAL_SPENT_TIME: sessionUserProperties.TotalSpentTime,
			U.UP_PAGE_COUNT:       sessionUserProperties.PageCount,
		}
		existingUserPropertiesMap, err := U.DecodePostgresJsonb(existingUserProperties)
		if err != nil {
			logCtx.WithError(err).
				Error("Failed to decode existing user properites.")
			hasFailure = true
			continue
		} else if sessionUserProperties.InitialChannel != "" {
			if channel, exists := (*existingUserPropertiesMap)[U.UP_INITIAL_CHANNEL]; !exists || channel == "" {
				newUserProperties[U.UP_INITIAL_CHANNEL] = sessionUserProperties.InitialChannel
			}
			newUserProperties[U.UP_LATEST_CHANNEL] = sessionUserProperties.LatestChannel
		}

		userPropertiesJsonb, err := U.AddToPostgresJsonb(existingUserProperties, newUserProperties, true)
		if err != nil {
			logCtx.WithError(err).
				Error("Failed to add new user properites to existing user properites.")
			hasFailure = true
			continue
		}

		if C.GetSessionBatchTransactionBatchSize() > 0 {
			overwriteUserPropertiesByIDParamsInBatch = append(overwriteUserPropertiesByIDParamsInBatch,
				model.OverwriteUserPropertiesByIDParams{
					ProjectID:           projectID,
					UserID:              userID,
					UserProperties:      userPropertiesJsonb,
					WithUpdateTimestamp: false,
					UpdateTimestamp:     0,
				})
			continue
		}

		errCode = store.OverwriteUserPropertiesByID(projectID, userID, userPropertiesJsonb, false, 0, "")
		if errCode != http.StatusAccepted {
			logCtx.WithField("err_code", errCode).Error("Failed to overwrite user properties record.")
			hasFailure = true
			continue
		}
	}

	if C.GetSessionBatchTransactionBatchSize() > 0 {
		batcheGetOverwriteUserPropertiesByIDParams := model.GetOverwriteUserPropertiesByIDParamsInBatch(overwriteUserPropertiesByIDParamsInBatch,
			C.GetSessionBatchTransactionBatchSize())
		for i := range batcheGetOverwriteUserPropertiesByIDParams {
			hasFailure = store.OverwriteUserPropertiesByIDInBatch(batcheGetOverwriteUserPropertiesByIDParams[i])
			if hasFailure {
				logCtx.Error("Failed to overwrite user properties record in batch process.")
			}
		}

	}

	if hasFailure {
		return http.StatusInternalServerError
	}

	return http.StatusAccepted
}

func shouldAllowCustomerUserID(current, incoming string) bool {
	logFields := log.Fields{
		"current":  current,
		"incoming": incoming,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
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
func (store *MemSQL) UpdateIdentifyOverwriteUserPropertiesMeta(projectID int64, customerUserID, userID, pageURL, source string, userProperties *postgres.Jsonb, timestamp int64, isNewUser bool) error {
	logFields := log.Fields{
		"project_id":       projectID,
		"customer_user_id": customerUserID,
		"user_id":          userID,
		"page_url":         pageURL,
		"user_properties":  userProperties,
		"source":           source,
		"time_stamp":       timestamp,
		"is_new_user":      isNewUser,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if projectID == 0 || customerUserID == "" {
		return errors.New("invalid or empty parameter")
	}
	if source == "" {
		return errors.New("source missing")
	}

	logCtx := log.WithFields(logFields)
	var existingUserProperties *map[string]interface{}
	var errCode int
	if !isNewUser {
		existingUserProperties, errCode = store.GetLatestUserPropertiesOfUserAsMap(projectID, userID)
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

	_, isExistingIdentifier := (*metaObj)[customerUserID]

	(*metaObj)[customerUserID] = *customerUserIDMeta

	// overwrite on existing identifier should count in increase in identification
	if !isExistingIdentifier && len(*metaObj) > 10 {
		logCtx.WithFields(log.Fields{"meta_object": metaObj}).Info("Number of identification exceeded.")
	}

	return model.UpdateUserPropertiesIdentifierMetaObject(userProperties, metaObj)
}

func (store *MemSQL) addGroupUserPropertyDetailsToCache(projectID int64, groupName,
	groupUserID string, properties *map[string]interface{}) {

	logCtx := log.WithField("properties", properties).WithField("project_id", projectID)

	if projectID == 0 || groupName == "" || groupUserID == "" || properties == nil {
		logCtx.Error("Invalid params on add group user property details to cache.")
		return
	}

	propertiesToIncrSortedSet := make([]cacheRedis.SortedSetKeyValueTuple, 0)
	valuesToIncrSortedSet := make([]cacheRedis.SortedSetKeyValueTuple, 0)

	_, status := store.GetGroup(projectID, groupName)
	if status != http.StatusFound {
		logCtx.Info("Unavailable group on addGroupUserPropertyDetailsToCache.")
		return
	}

	currentTime := U.TimeNowZ()
	currentTimeDatePart := currentTime.Format(U.DATETIME_FORMAT_YYYYMMDD)
	for property, value := range *properties {
		if value == nil {
			continue
		}

		// TODO: Add support type support on property details for group properties.
		// Using user_properties temporarily.
		category := store.GetPropertyTypeByKeyValue(projectID, "", property, value, true)
		var propertyValue string
		if category == U.PropertyTypeUnknown && reflect.TypeOf(value).Kind() == reflect.Bool {
			category = U.PropertyTypeCategorical
			propertyValue = fmt.Sprintf("%v", value)
		}
		if reflect.TypeOf(value).Kind() == reflect.String {
			propertyValue = value.(string)
		}

		groupPropertyCategoryKeySortedSet, err := model.GetPropertiesByGroupCategoryCacheKeySortedSet(projectID, currentTimeDatePart)
		if err != nil {
			logCtx.WithError(err).Error("Failed to get cache key - group property category")
			return
		}
		propertiesToIncrSortedSet = append(propertiesToIncrSortedSet, cacheRedis.SortedSetKeyValueTuple{
			Key:   groupPropertyCategoryKeySortedSet,
			Value: fmt.Sprintf("%s:SS-GN-PC:%s:%s", groupName, category, property),
		})
		if category == U.PropertyTypeCategorical {
			if propertyValue != "" {
				groupValueKeySortedSet, err := model.GetValuesByGroupPropertyCacheKeySortedSet(projectID, currentTimeDatePart)
				if err != nil {
					logCtx.WithError(err).Error("Failed to get cache key - group values")
					return
				}
				valuesToIncrSortedSet = append(valuesToIncrSortedSet, cacheRedis.SortedSetKeyValueTuple{
					Key:   groupValueKeySortedSet,
					Value: fmt.Sprintf("%s:SS-GN-PC:%s:SS-GN-PV:%s", groupName, property, propertyValue),
				})
			}
		}
	}
	begin := U.TimeNowZ()
	keysToIncrSortedSet := make([]cacheRedis.SortedSetKeyValueTuple, 0)
	keysToIncrSortedSet = append(keysToIncrSortedSet, propertiesToIncrSortedSet...)
	keysToIncrSortedSet = append(keysToIncrSortedSet, valuesToIncrSortedSet...)
	if len(keysToIncrSortedSet) <= 0 {
		return
	}
	_, err := cacheRedis.ZincrPersistentBatch(false, keysToIncrSortedSet...)
	end := U.TimeNowZ()
	metrics.Increment(metrics.IncrGroupCacheCounter)
	metrics.RecordLatency(metrics.LatencyGroupCache, float64(end.Sub(begin).Milliseconds()))
	if err != nil {
		logCtx.WithError(err).Error("Failed to increment keys")
		return
	}
}

func (store *MemSQL) CreateGroupUser(user *model.User, groupName, groupID string) (string, int) {
	logFields := log.Fields{
		"user":       user,
		"group_name": groupName,
		"group_id":   groupID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)
	group, status := store.GetGroup(user.ProjectId, groupName)
	if status != http.StatusFound {
		if status == http.StatusNotFound {
			logCtx.WithField("err_code", status).Error("Group is missing on CreateGroupUser.")
			return "", http.StatusBadRequest
		}

		logCtx.WithField("err_code", status).Error("Failed to get group on CreateGroupUser.")
		return "", http.StatusInternalServerError
	}

	isGroupUser := true
	groupUser := &model.User{
		ProjectId:                  user.ProjectId,
		IsGroupUser:                &isGroupUser,
		Properties:                 user.Properties,
		PropertiesUpdatedTimestamp: user.PropertiesUpdatedTimestamp,
		JoinTimestamp:              user.JoinTimestamp,
		Source:                     user.Source,
	}

	if user.ID != "" {
		groupUser.ID = user.ID
	}

	if groupID != "" {
		groupIndex := fmt.Sprintf("group_%d_id", group.ID)
		processed, _, err := model.SetUserGroupFieldByColumnName(groupUser, groupIndex, groupID, false)
		if err != nil {
			logCtx.WithError(err).Error("Failed process group id on group user.")
			return "", http.StatusInternalServerError
		}

		if !processed {
			logCtx.WithError(err).Error("Failed to process group_id on group user.")
			return "", http.StatusInternalServerError

		}
	} else {
		logCtx.Warning("Skip associating group_id")
	}

	return store.CreateUser(groupUser)
}

func (store *MemSQL) UpdateUserGroup(projectID int64, userID, groupName, groupID, groupUserID string, overwrite bool) (*model.User, int) {
	logFields := log.Fields{
		"project_id":    projectID,
		"user_id":       userID,
		"group_name":    groupName,
		"group_id":      groupID,
		"group_user_id": groupUserID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	if projectID == 0 || userID == "" || groupName == "" || groupUserID == "" {
		logCtx.Error("Invalid parameters.")
		return nil, http.StatusBadRequest
	}

	if groupName == model.GROUP_NAME_DOMAINS {
		logCtx.Error("domains group not allowed.")
		return nil, http.StatusBadRequest
	}

	group, status := store.GetGroup(projectID, groupName)
	if status != http.StatusFound {
		if status == http.StatusNotFound {
			logCtx.WithField("err_code", status).Error("Group is missing on UpdateUserGroup.")
			return nil, http.StatusBadRequest
		}

		logCtx.WithField("err_code", status).Error("Failed to get group on UpdateUserGroup.")
		return nil, http.StatusInternalServerError
	}

	user, status := store.GetUserWithoutProperties(projectID, userID)
	if status != http.StatusFound {
		logCtx.WithField("err_code", status).Error("Failed to get user for group association.")
		return nil, http.StatusInternalServerError
	}

	if user.IsGroupUser != nil && *user.IsGroupUser {
		logCtx.Error("Cannot update group user.")
		return nil, http.StatusBadRequest
	}

	isGroupUser := false
	user.IsGroupUser = &isGroupUser
	return store.updateUserGroup(projectID, user, userID, group.ID, groupID, groupUserID, overwrite)
}

func (store *MemSQL) updateUserGroup(projectID int64, user *model.User, userID string, groupIndex int, groupID, groupUserID string, overwrite bool) (*model.User, int) {
	logFields := log.Fields{
		"project_id":    projectID,
		"group_index":   groupIndex,
		"group_id":      groupID,
		"group_user_id": groupUserID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	if projectID == 0 || user == nil || groupIndex == 0 || groupUserID == "" {
		logCtx.Error("Invalid parameters on updateUserGroup.")
		return nil, http.StatusBadRequest
	}
	logFields["user_id"] = user.ID

	if user.IsGroupUser == nil {
		logCtx.Error("Missing user.")
		return nil, http.StatusBadRequest
	}

	var IDUpdated, userIDUpdated, processed bool
	var err error
	if groupID != "" {
		groupIndexColumn := fmt.Sprintf("group_%d_id", groupIndex)
		processed, IDUpdated, err = model.SetUserGroupFieldByColumnName(user, groupIndexColumn, groupID, overwrite)
		if err != nil {
			logCtx.WithError(err).Error("Failed to update user by group id.")
			return nil, http.StatusInternalServerError
		}
		if !processed {
			logCtx.Error("Missing tag in struct for group id.")
			return nil, http.StatusInternalServerError
		}
	}

	groupUserIndexColumn := fmt.Sprintf("group_%d_user_id", groupIndex)
	processed, userIDUpdated, err = model.SetUserGroupFieldByColumnName(user, groupUserIndexColumn, groupUserID, overwrite)
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
	return store.UpdateUser(projectID, userID, user, user.PropertiesUpdatedTimestamp)
}

func (store *MemSQL) UpdateGroupUserDomainsGroup(projectID int64, groupUserID, groupUserGroupName, domainsUserID, domainsGroupID string, overwrite bool) (*model.User, int) {
	logFields := log.Fields{
		"project_id":            projectID,
		"group_user_id":         groupUserID,
		"domains_user_id":       domainsUserID,
		"group_user_group_name": groupUserGroupName,
		"domains_group_id":      domainsGroupID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	if projectID == 0 || groupUserID == "" || groupUserGroupName == "" || domainsGroupID == "" || domainsUserID == "" {
		logCtx.Error("Invalid parameters.")
		return nil, http.StatusBadRequest
	}

	group, status := store.GetGroup(projectID, model.GROUP_NAME_DOMAINS)
	if status != http.StatusFound {
		if status == http.StatusNotFound {
			logCtx.WithField("err_code", status).Error("domains group not found.")
			return nil, http.StatusBadRequest
		}

		logCtx.WithField("err_code", status).Error("Failed to get domains group.")
		return nil, http.StatusInternalServerError
	}

	user, status := store.GetUserWithoutProperties(projectID, groupUserID)
	if status != http.StatusFound {
		logCtx.WithField("err_code", status).Error("Failed to get user for group association.")
		return nil, http.StatusInternalServerError
	}

	if user.IsGroupUser == nil || *user.IsGroupUser == false {
		logCtx.Error("Cannot update non group user.")
		return nil, http.StatusBadRequest
	}

	isGroupUser := true
	user.IsGroupUser = &isGroupUser
	return store.updateUserGroup(projectID, user, groupUserID, group.ID, domainsGroupID, domainsUserID, overwrite)
}

func (store *MemSQL) updateUserDomainsGroup(projectID int64, userIDs []string, domainGroupIndex int, domainsUserID string) int {
	logFields := log.Fields{
		"project_id":         projectID,
		"user_ids":           userIDs,
		"domain_user_id":     domainsUserID,
		"domain_group_index": domainGroupIndex,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	if projectID == 0 || len(userIDs) == 0 || domainGroupIndex <= 0 || domainsUserID == "" {
		logCtx.Error("Invalid parameters.")
		return http.StatusBadRequest
	}

	groupUserIDColumn := fmt.Sprintf("group_%d_user_id", domainGroupIndex)

	db := C.GetServices().Db
	update := map[string]string{
		groupUserIDColumn: domainsUserID,
	}

	err := db.Model(model.User{}).Where("project_id = ? AND id IN ( ? )", projectID, userIDs).Updates(update).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to updateUserDomainsGroup.")
		return http.StatusInternalServerError
	}

	return http.StatusAccepted
}

func (store *MemSQL) UpdateUserGroupProperties(projectID int64, userID string,
	newProperties *postgres.Jsonb, updateTimestamp int64) (*postgres.Jsonb, int) {
	logFields := log.Fields{
		"project_id":       projectID,
		"user_id":          userID,
		"new_properties":   newProperties,
		"update_timestamp": updateTimestamp,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	if projectID == 0 || userID == "" || newProperties == nil {
		logCtx.Error("Invalid parameters.")
		return nil, http.StatusBadRequest
	}

	user, errCode := store.GetUser(projectID, userID)
	if errCode != http.StatusFound {
		logCtx.WithField("err_code", errCode).Error("Failed to get user on UpdateUserGroupProperties.")
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

	errCode = store.OverwriteUserPropertiesByID(projectID, user.ID, mergedPropertiesJSON, true, newUpdateTimestamp, "")
	if errCode == http.StatusInternalServerError || errCode == http.StatusBadRequest {
		logCtx.WithField("err_code", errCode).WithField("user_id", user.ID).Error("Failed to update user properties on group user.")
		return nil, http.StatusInternalServerError
	}

	return mergedPropertiesJSON, http.StatusAccepted
}

func (store *MemSQL) UpdateGroupUserGroupId(projectID int64, userID string,
	groupID string, columnName string) int {
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "id": userID, "group_id": groupID, "columnName": columnName})
	if projectID == 0 || userID == "" || groupID == "" || columnName == "" {
		logCtx.Error("Invalid parameters.")
		return http.StatusBadRequest
	}
	updatedField := map[string]interface{}{
		columnName: groupID,
	}
	db := C.GetServices().Db
	err := db.Table("users").Where("project_id = ? AND id = ? AND is_group_user = true", projectID, userID).
		Updates(updatedField).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to update groupID.")
		return http.StatusInternalServerError
	}
	return http.StatusAccepted
}

// GetUserWithoutProperties Gets the user without properties
func (store *MemSQL) GetUserWithoutProperties(projectID int64, id string) (*model.User, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"user_id":    id,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)
	if projectID <= 0 || id == "" {
		logCtx.Error("Invalid parameters.")
		return nil, http.StatusBadRequest
	}

	var user model.User
	db := C.GetServices().Db
	requiredFields := []string{}
	for _, field := range db.NewScope(&model.User{}).GetStructFields() {
		if field.DBName != "properties" {
			requiredFields = append(requiredFields, field.DBName)
		}
	}

	selectColumns := strings.Join(requiredFields, ",")

	if err := db.Where("project_id = ? AND id = ?", projectID, id).Select(selectColumns).
		Find(&user).Error; err != nil {

		logCtx.WithError(err).Error("Failed to get users for customer_user_id")
		return nil, http.StatusInternalServerError
	}

	if user.ID == "" {
		return nil, http.StatusNotFound
	}

	return &user, http.StatusFound
}

func (store *MemSQL) GetCustomerUserIdFromUserId(projectID int64, id string) (string, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"user_id":    id,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)
	if projectID == 0 || id == "" {
		logCtx.Error("Invalid parameters.")
		return "", http.StatusBadRequest
	}

	var user model.User

	db := C.GetServices().Db
	if err := db.Limit(1).Where("project_id = ? AND id = ?", projectID, id).Select("customer_user_id").Find(&user).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return "", http.StatusNotFound
		}

		logCtx.WithError(err).Error("Getting customer_user_id failed on GetCustomerUserIdFromUserId")
		return "", http.StatusInternalServerError
	}
	return user.CustomerUserId, http.StatusFound
}

func (store *MemSQL) PullUsersRowsForWIV2(projectID int64, startTime, endTime int64, dateField string, source int, group int) (*sql.Rows, *sql.Tx, error) {
	logFields := log.Fields{
		"project_id": projectID,
		"start_time": startTime,
		"end_time":   endTime,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	whereGroupStmt := fmt.Sprintf("(is_group_user=1 AND group_%d_id IS NOT NULL)", group)
	if group == 0 {
		whereGroupStmt = "(is_group_user=0 OR is_group_user IS NULL)"
	}
	rawQuery := fmt.Sprintf("SELECT COALESCE(customer_user_id, id) AS user_id, properties,ISNULL(customer_user_id) AS is_anonymous, join_timestamp AS join_timestamp, "+
		"COALESCE(CASE WHEN UNIX_TIMESTAMP(JSON_EXTRACT_STRING(properties, '%s'))>0 THEN UNIX_TIMESTAMP(JSON_EXTRACT_STRING(properties, '%s')) ELSE JSON_EXTRACT_STRING(properties, '%s') END,0) AS timestamp FROM users "+
		"WHERE %s AND project_id=%d AND source=%d AND UNIX_TIMESTAMP(created_at) BETWEEN %d AND %d AND updated_at<NOW() AND timestamp>0 "+
		"LIMIT %d",
		dateField, dateField, dateField, whereGroupStmt, projectID, source, startTime, endTime, model.UsersPullLimit+1)

	rows, tx, err, _ := store.ExecQueryWithContext(rawQuery, []interface{}{})
	return rows, tx, err
}

func (store *MemSQL) PullUsersRowsForWIV1(projectID int64, startTime, endTime int64, dateField string, source int, group int) (*sql.Rows, *sql.Tx, error) {
	logFields := log.Fields{
		"project_id": projectID,
		"start_time": startTime,
		"end_time":   endTime,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	whereGroupStmt := fmt.Sprintf("(is_group_user=1 AND group_%d_id IS NOT NULL)", group)
	if group == 0 {
		whereGroupStmt = "(is_group_user=0 OR is_group_user IS NULL)"
	}
	rawQuery := fmt.Sprintf("SELECT COALESCE(customer_user_id, id) as user_id, properties,ISNULL(customer_user_id) AS is_anonymous, "+
		"MIN(join_timestamp) as join_timestamp FROM users "+
		"WHERE %s AND project_id=%d AND source=%d AND JSON_EXTRACT_STRING(properties, '%s')>=%d AND JSON_EXTRACT_STRING(properties, '%s')<=%d AND updated_at<NOW() "+
		"GROUP BY user_id  ORDER BY join_timestamp LIMIT %d",
		whereGroupStmt, projectID, source, dateField, startTime, dateField, endTime, model.UsersPullLimit+1)

	rows, tx, err, _ := store.ExecQueryWithContext(rawQuery, []interface{}{})
	return rows, tx, err
}

func (store *MemSQL) AssociateUserDomainsGroup(projectID int64, requestUserID string, requestGroupName, requestGroupUserID string) int {

	logFields := log.Fields{"project_id": projectID, "request_user_id": requestUserID, "request_group_name": requestGroupName,
		"request_group_user_id": requestGroupUserID, "source": source}
	logCtx := log.WithFields(logFields)

	if projectID == 0 || requestUserID == "" || source == "" {
		logCtx.Error("Invalid parameters.")
		return http.StatusBadRequest
	}

	if requestGroupName != "" && !U.ContainsStringInArray(model.AccountGroupAssociationPrecedence, requestGroupName) {
		logCtx.Error("Invalid group name on AssociateUserDomainsGroup.")
		return http.StatusBadRequest
	}

	groups, status := store.GetGroups(projectID)
	if status != http.StatusFound {
		logCtx.WithFields(log.Fields{"err_code": status}).Error("Failed to get groups on AssociateUserDomainsGroup.")
		return status
	}

	groupIDMap := make(map[string]int)
	for i := range groups {
		groupIDMap[groups[i].Name] = groups[i].ID
	}

	if groupIDMap[model.GROUP_NAME_DOMAINS] == 0 && requestGroupName != "" && groupIDMap[requestGroupName] == 0 {
		logCtx.Error("Missing domains group or request group.")
		return http.StatusBadRequest
	}

	requestUser, status := store.GetUserWithoutProperties(projectID, requestUserID)
	if status != http.StatusFound {
		logCtx.WithFields(log.Fields{"err_code": status}).Error("Failed to get user in AssociateUserDomainsGroup.")
		return http.StatusInternalServerError
	}

	users := []model.User{*requestUser}
	if requestUser.CustomerUserId != "" {
		usersByCustomerUserID, status := store.GetUsersByCustomerUserID(projectID, requestUser.CustomerUserId)
		if status != http.StatusFound {
			logCtx.WithFields(log.Fields{"err_code": status}).Error("Failed to get users by customers on AssociateUserDomainsGroup.")
			return http.StatusInternalServerError
		}

		if len(users) > 100 {
			logCtx.WithFields(log.Fields{"customer_user_id": requestUser.CustomerUserId}).Error("Number of users by customer user id exceeds 100 in AssociateUserDomainsGroup.")
		}

		users = usersByCustomerUserID
	}

	requiredGroupIDsByPrecedence := make([]int, 0)
	for _, requiredGroup := range model.AccountGroupAssociationPrecedence {
		if groupIDMap[requiredGroup] != 0 {
			requiredGroupIDsByPrecedence = append(requiredGroupIDsByPrecedence, groupIDMap[requiredGroup])
		}
	}

	userGroupUserIDs, err := model.GetUserGroupUserIDsByGroupIDs(projectID, users, requiredGroupIDsByPrecedence)
	if err != nil {
		logCtx.WithError(err).Error("Failed to GetUserGroupUserIDsByGroupIDs on AssociateUserDomainsGroup.")
		return http.StatusInternalServerError
	}

	if requestGroupName != "" && requestGroupUserID != "" {
		userGroupUserIDs[groupIDMap[requestGroupName]] = requestGroupUserID
	}

	leadingGroupUserID := ""
	leadingGroupName := ""
	for i := range requiredGroupIDsByPrecedence {
		if groupUserID, exist := userGroupUserIDs[requiredGroupIDsByPrecedence[i]]; exist && groupUserID != "" {
			leadingGroupUserID = groupUserID

			for groupName, groupIndex := range groupIDMap {
				if groupIndex == requiredGroupIDsByPrecedence[i] {
					leadingGroupName = groupName
				}
			}

			break
		}
	}

	emailDomainUserID := ""
	if C.AllowEmailDomainsByProjectID(projectID) &&
		(leadingGroupName == model.GROUP_NAME_SIX_SIGNAL || leadingGroupName == "") &&
		requestUser.CustomerUserId != "" {
		if U.IsEmail(requestUser.CustomerUserId) {
			domainName := U.GetDomainGroupDomainName(projectID, U.GetEmailDomain(requestUser.CustomerUserId))
			if domainName != "" {
				emailDomainUserID, status = store.CreateOrGetDomainGroupUser(projectID, model.GROUP_NAME_DOMAINS, domainName,
					U.TimeNowUnix(), model.GetGroupUserSourceByGroupName(model.GROUP_NAME_DOMAINS))
				if status != http.StatusCreated && status != http.StatusFound {
					logCtx.WithFields(log.Fields{"err_code": status}).Error("Failed to check for group user by group id in createOrGetDomainUserIDByProperties using email domain.")
					return http.StatusInternalServerError
				}
			} else {
				logCtx.WithFields(log.Fields{"email": requestUser.CustomerUserId, "email_domain": domainName}).Warning("Failed to get domain name from email.")
			}
		}
	}

	if leadingGroupUserID == "" && emailDomainUserID == "" {
		logCtx.Warning("No leading group user id found.")
		return http.StatusNotModified
	}

	domainUserID := ""
	if emailDomainUserID == "" {
		groupUser, status := store.GetUser(projectID, leadingGroupUserID)
		if status != http.StatusFound {
			logCtx.WithFields(log.Fields{"group_user_id": leadingGroupUserID}).Error("Failed to get group user.")
			return http.StatusInternalServerError
		}

		domainUserID, err = model.GetGroupUserDomainsUserID(groupUser, groupIDMap[model.GROUP_NAME_DOMAINS])
		if err != nil && err != model.ErrMissingDomainUserID {
			logCtx.WithFields(log.Fields{"group_user": groupUser, "domain_group_id": groupIDMap[model.GROUP_NAME_DOMAINS]}).WithError(err).
				Error("Failed to get domains user id.")
			return http.StatusInternalServerError
		}

		if domainUserID == "" {
			// backfill domain user if available in properties
			var groupID string
			domainUserID, groupID, status = store.createOrGetDomainUserIDByProperties(projectID, leadingGroupName, groupUser.Properties)
			if status != http.StatusFound && status != http.StatusCreated && status != http.StatusNotFound {
				logCtx.WithFields(log.Fields{"group_user": groupUser, "domain_group_id": groupIDMap[model.GROUP_NAME_DOMAINS]}).WithError(err).
					Error("Failed to get domains user id after createOrGetDomainUserIDByProperties.")
				return http.StatusInternalServerError
			}

			// in case no domain property present skip processing
			if status == http.StatusNotFound {
				return http.StatusOK
			}

			// associate domain user id to group user to avoid going through fallback logic later.
			_, status = store.UpdateGroupUserDomainsGroup(projectID, groupUser.ID, leadingGroupName, domainUserID, groupID, true)
			if status != http.StatusAccepted && status != http.StatusNotModified {
				logCtx.WithFields(log.Fields{"domain_user_id": domainUserID, "group_user_id": groupUser.ID, "group_name": leadingGroupName, "group_id": groupID}).
					WithError(err).Error("Failed to update group user association with domains group user on AssociateUserDomainsGroup.")
				return http.StatusInternalServerError
			}
		}
	} else {
		domainUserID = emailDomainUserID
	}

	userIDs := []string{}
	for i := range users {
		userIDs = append(userIDs, users[i].ID)
	}

	status = store.updateUserDomainsGroup(projectID, userIDs, groupIDMap[model.GROUP_NAME_DOMAINS], domainUserID)
	if status != http.StatusAccepted && status != http.StatusNotModified {
		logCtx.WithFields(log.Fields{"user_id": userIDs, "domain_group_index": groupIDMap[model.GROUP_NAME_DOMAINS], "domain_user_id": domainUserID}).
			Error("Failed to update domains user id on user.")
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

func (store *MemSQL) createOrGetDomainUserIDByProperties(projectID int64, groupName string, properties postgres.Jsonb) (string, string, int) {

	logFields := log.Fields{"project_id": projectID, "group_name": groupName}
	logCtx := log.WithFields(logFields)
	if projectID == 0 || groupName == "" || U.IsEmptyPostgresJsonb(&properties) {
		logCtx.Error("Invalid parameters")
		return "", "", http.StatusBadRequest
	}

	propertyKey := model.GetDomainNameSourcePropertyKey(groupName)
	if propertyKey == "" {
		logCtx.Error("Empty property key on createOrGetDomainUserIDByProperties.")
		return "", "", http.StatusInternalServerError
	}

	propertiesMap, err := U.DecodePostgresJsonbAsPropertiesMap(&properties)
	if err != nil {
		logCtx.WithError(err).Error("Failed to decode properties in createOrGetDomainUserIDByProperties.")
		return "", "", http.StatusInternalServerError
	}

	domainName := U.GetPropertyValueAsString((*propertiesMap)[propertyKey])
	if domainName == "" {
		logCtx.WithFields(log.Fields{"properties": propertiesMap}).Warn("No domain name found. Skip processing domain.")
		return "", "", http.StatusNotFound
	}

	cleanedDomainName := U.GetDomainGroupDomainName(projectID, domainName)
	if cleanedDomainName == "" {
		logCtx.WithFields(log.Fields{"properties": propertiesMap}).Error("No domain name found after domain name cleaining. Skip processing domain.")
		return "", "", http.StatusNotFound
	}

	groupUserID, status := store.CreateOrGetDomainGroupUser(projectID, groupName, cleanedDomainName, U.TimeNowUnix(), model.GetGroupUserSourceByGroupName(groupName))
	if status != http.StatusCreated && status != http.StatusFound {
		logCtx.WithFields(log.Fields{"err_code": status}).Error("Failed to check for group user by group id in createOrGetDomainUserIDByProperties.")
		return "", "", http.StatusInternalServerError
	}

	return groupUserID, cleanedDomainName, status
}
