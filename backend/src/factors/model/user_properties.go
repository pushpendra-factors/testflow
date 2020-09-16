package model

import (
	"encoding/json"
	cacheRedis "factors/cache/redis"
	C "factors/config"
	U "factors/util"
	"sort"

	"fmt"
	"net"
	"net/http"
	"reflect"
	"time"

	"github.com/imdario/mergo"
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
	Properties       postgres.Jsonb `json:"properties"`
	UpdatedTimestamp int64          `gorm:"not null;default:0" json:"updated_timestamp"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

type CachedVisitedUsersList struct {
	Users map[string]bool
}

// indexed hubspot user property.
const UserPropertyHubspotContactLeadGUID = "$hubspot_contact_lead_guid"
const MaxUsersForPropertiesMerge = 100

func createUserPropertiesIfChanged(projectId uint64, userId string,
	currentPropertiesId string, newProperties *postgres.Jsonb, timestamp int64) (string, int) {

	if U.IsEmptyPostgresJsonb(newProperties) {
		return currentPropertiesId, http.StatusNotModified
	}

	currentProperties := &postgres.Jsonb{}
	var statusCode int

	var currentPropertiesMap map[string]interface{}
	var newPropertiesMap map[string]interface{}
	var mergedPropertiesMap map[string]interface{}

	shouldMergeUserProperties := false
	json.Unmarshal((*newProperties).RawMessage, &newPropertiesMap)
	if currentPropertiesId != "" {
		var currentPropertiesRecord *UserProperties
		currentPropertiesRecord, statusCode = GetUserPropertiesRecord(
			projectId, userId, currentPropertiesId)

		if statusCode == http.StatusInternalServerError {
			log.WithField("current_properties_id", currentPropertiesId).Error(
				"Failed to GetUserProperties on createUserPropertiesIfChanged.")
			return "", http.StatusInternalServerError
		}

		currentProperties = &currentPropertiesRecord.Properties

		json.Unmarshal((*currentProperties).RawMessage, &currentPropertiesMap)
		// init mergedProperties with currentProperties.
		json.Unmarshal((*currentProperties).RawMessage, &mergedPropertiesMap)
		if statusCode == http.StatusFound {
			// Overwrite the keys only, if the update is future,
			// else only add new keys.
			if timestamp >= currentPropertiesRecord.UpdatedTimestamp {
				mergo.Merge(&mergedPropertiesMap, newPropertiesMap, mergo.WithOverride)
			} else {
				mergo.Merge(&mergedPropertiesMap, newPropertiesMap)
			}

			// Using merged properties for equality check to achieve
			// currentPropertiesMap {x: 1, y: 2} newPropertiesMap {x: 1} -> true
			if reflect.DeepEqual(currentPropertiesMap, mergedPropertiesMap) {
				UpdateCacheForUserProperties(userId, projectId, currentPropertiesMap, true)
				return currentPropertiesId, http.StatusNotModified
			}
			// If not equal, trigger user properties merge.
			shouldMergeUserProperties = true
		}
	} else {
		mergedPropertiesMap = newPropertiesMap
	}

	UpdateCacheForUserProperties(userId, projectId, mergedPropertiesMap, false)
	// Overwrite only given values.
	updatedPropertiesBytes, err := json.Marshal(mergedPropertiesMap)
	if err != nil {
		return "", http.StatusInternalServerError
	}
	if shouldMergeUserProperties && isMergeEnabledForProjectID(projectId) {
		newPropertiesID, errCode := MergeUserPropertiesForUserID(projectId, userId,
			postgres.Jsonb{RawMessage: json.RawMessage(updatedPropertiesBytes)}, currentPropertiesId, timestamp, false, false)

		// Return only if merge is successfull. Else do usual update.
		if errCode == http.StatusCreated || errCode == http.StatusNotModified {
			return newPropertiesID, errCode
		}
	}

	return updateUserPropertiesForUser(projectId, userId,
		postgres.Jsonb{RawMessage: json.RawMessage(updatedPropertiesBytes)}, timestamp, false)
}

func BackFillUserDataInCacheFromDb(projectid uint64, currentDate time.Time, usersProcessedLimit int, propertiesLimit int, valuesLimit int) {

	logCtx := log.WithFields(log.Fields{
		"project_id": projectid,
	})
	logCtx.Info("Refresh User Properties Cache started")
	currentDateFormat := currentDate.AddDate(0, 0, -1).Format(U.DATETIME_FORMAT_YYYYMMDD)
	var userPropertiesTillDate U.CachePropertyWithTimestamp
	userPropertiesTillDate.Property = make(map[string]U.PropertyWithTimestamp)
	propertyCacheKey, err := GetUserPropertiesCategoryByProjectRollUpCacheKey(projectid, currentDateFormat)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get property cache key - getuserpropertiesbyproject")
	}
	logCtx.WithField("dateFormat", currentDateFormat).Info("Begin: User Properties - DB query")
	begin := U.TimeNow()
	properties, err := GetRecentUserPropertyKeysWithLimits(projectid, usersProcessedLimit, propertiesLimit)
	end := U.TimeNow()
	logCtx.WithFields(log.Fields{"dateFormat": currentDateFormat, "timeTaken": end.Sub(begin).Milliseconds()}).Info("End: User Properties - DB query")
	if err != nil {
		logCtx.WithError(err).Error("Failed to get property keys - getrecentuserpropertykeyswithlimits")
	}
	for _, propertyValue := range properties {
		logCtx.WithFields(log.Fields{"dateFormat": currentDateFormat, "property": propertyValue.Key}).Info("Begin: Get user Property values DB call")
		begin := U.TimeNow()
		values, category, err := GetRecentUserPropertyValuesWithLimits(projectid, propertyValue.Key, usersProcessedLimit, valuesLimit)
		end := U.TimeNow()
		logCtx.WithFields(log.Fields{"dateFormat": currentDateFormat, "property": propertyValue.Key, "timeTaken": end.Sub(begin).Milliseconds()}).Info("End: Get user Property values DB call")
		if err != nil {
			logCtx.WithError(err).Error("Failed to get property values - getrecentuserpropertyvalueswithlimits")
		}
		categoryMap := make(map[string]int64)
		categoryMap[category] = propertyValue.Count
		userPropertiesTillDate.Property[propertyValue.Key] = U.PropertyWithTimestamp{
			category,
			categoryMap,
			U.CountTimestampTuple{
				int64(propertyValue.LastSeen),
				propertyValue.Count}}
		var PropertyValues U.CachePropertyValueWithTimestamp
		PropertyValues.PropertyValue = make(map[string]U.CountTimestampTuple)
		if category == U.PropertyTypeCategorical {
			PropertyValuesKey, err := GetValuesByUserPropertyRollUpCacheKey(projectid, propertyValue.Key, currentDateFormat)
			if err != nil {
				logCtx.WithError(err).Error("Failed to get property cache key - getvaluesbyuserproperty")
			}
			for _, value := range values {
				if value.Value != "" {
					PropertyValues.PropertyValue[value.Value] = U.CountTimestampTuple{
						int64(value.LastSeen),
						value.Count}
				}
			}
			enPropertyValuesCache, err := json.Marshal(PropertyValues)
			if err != nil {
				logCtx.WithError(err).Error("Failed to marshal property value - getvaluesbyuserproperty")
			}
			begin := U.TimeNow()
			err = cacheRedis.SetPersistent(PropertyValuesKey, string(enPropertyValuesCache), U.EVENT_USER_CACHE_EXPIRY_SECS)
			end := U.TimeNow()
			logCtx.WithFields(log.Fields{"timeTaken": end.Sub(begin).Milliseconds()}).Info("End:UP:BS")
			if err != nil {
				logCtx.WithError(err).Error("Failed to set cache property value - getvaluesbyuserproperty")
			}
		}
	}
	enPropertiesCache, err := json.Marshal(userPropertiesTillDate)
	if err != nil {
		logCtx.WithError(err).Error("Failed to marshal property key - getuserpropertiesbyproject")
	}
	logCtx.Info("Begin:UP:BS")
	begin = U.TimeNow()
	err = cacheRedis.SetPersistent(propertyCacheKey, string(enPropertiesCache), U.EVENT_USER_CACHE_EXPIRY_SECS)
	end = U.TimeNow()
	logCtx.WithFields(log.Fields{"timeTaken": end.Sub(begin).Milliseconds()}).Info("End:UP:BS")
	if err != nil {
		logCtx.WithError(err).Error("Failed to set property key - getuserpropertiesbyproject")
	}
	logCtx.Info("Refresh Event Properties Cache Done !!!")
}

func UpdateCacheForUserProperties(userId string, projectid uint64, updatedProperties map[string]interface{}, redundantProperty bool) {
	// TODO: Remove this check after enabling caching realtime.
	if !C.GetIfRealTimeEventUserCachingIsEnabled(projectid) {
		return
	}

	// If the cache is empty / cache is updated from more than 1 day - repopulate cache
	logCtx := log.WithFields(log.Fields{
		"project_id": projectid,
	})
	currentTime := U.TimeNow()
	currentTimeDatePart := currentTime.Format(U.DATETIME_FORMAT_YYYYMMDD)
	// Store Last updated from DB in cache as a key. and check and refresh cache accordingly
	usersCacheKey, err := GetUsersCachedCacheKey(projectid, currentTimeDatePart)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get property cache key - getuserscachedcachekey")
	}

	begin := U.TimeNow()
	isNewUser, err := cacheRedis.PFAddPersistent(usersCacheKey, userId)
	end := U.TimeNow()
	logCtx.WithField("timeTaken", end.Sub(begin).Milliseconds()).Info("US:List")
	if err != nil {
		logCtx.WithError(err).Error("Failed to get users from cache - getuserscachedcachekey")
	}

	if redundantProperty == true && isNewUser == false {
		return
	}
	keysToIncr := make([]*cacheRedis.Key, 0)
	propertiesToIncr := make([]*cacheRedis.Key, 0)
	valuesToIncr := make([]*cacheRedis.Key, 0)
	for property, value := range updatedProperties {
		category := U.GetPropertyTypeByKeyValue(property, value)
		var propertyValue string
		if category == U.PropertyTypeUnknown && reflect.TypeOf(value).Kind() == reflect.Bool {
			category = U.PropertyTypeCategorical
			propertyValue = fmt.Sprintf("%v", value)
		}
		if reflect.TypeOf(value).Kind() == reflect.String {
			propertyValue = value.(string)
		}
		propertyCategoryKey, err := GetUserPropertiesCategoryByProjectCacheKey(projectid, property, category, currentTimeDatePart)
		if err != nil {
			logCtx.WithError(err).Error("Failed to get cache key - property category")
			return
		}
		propertiesToIncr = append(propertiesToIncr, propertyCategoryKey)
		if category == U.PropertyTypeCategorical {
			valueKey, err := GetValuesByUserPropertyCacheKey(projectid, property, propertyValue, currentTimeDatePart)
			if err != nil {
				logCtx.WithError(err).Error("Failed to get cache key - values")
				return
			}
			valuesToIncr = append(valuesToIncr, valueKey)
		}
	}
	keysToIncr = append(keysToIncr, propertiesToIncr...)
	keysToIncr = append(keysToIncr, valuesToIncr...)
	begin = U.TimeNow()
	counts, err := cacheRedis.IncrPersistentBatch(keysToIncr...)
	end = U.TimeNow()
	logCtx.WithField("timeTaken", end.Sub(begin).Milliseconds()).Info("US:Incr")
	if err != nil {
		logCtx.WithError(err).Error("Failed to increment keys")
		return
	}

	// The following code is to support/facilitate cleanup
	newPropertiesCount := int64(0)
	newValuesCount := int64(0)
	for _, value := range counts[0:len(propertiesToIncr)] {
		if value == 1 {
			newPropertiesCount++
		}
	}
	for _, value := range counts[len(propertiesToIncr) : len(propertiesToIncr)+len(valuesToIncr)] {
		if value == 1 {
			newValuesCount++
		}
	}

	countsInCache := make(map[*cacheRedis.Key]int64)
	if newPropertiesCount > 0 {
		propertiesCountKey, err := GetUserPropertiesCategoryByProjectCountCacheKey(projectid, currentTimeDatePart)
		if err != nil {
			logCtx.WithError(err).Error("Failed to get cache key - propertiesCount")
			return
		}
		countsInCache[propertiesCountKey] = newPropertiesCount
	}
	if newValuesCount > 0 {
		valuesCountKey, err := GetValuesByUserPropertyCountCacheKey(projectid, currentTimeDatePart)
		if err != nil {
			logCtx.WithError(err).Error("Failed to get cache key - valuesCount")
			return
		}
		countsInCache[valuesCountKey] = newValuesCount
	}
	if len(countsInCache) > 0 {
		begin := U.TimeNow()
		err = cacheRedis.IncrByBatchPersistent(countsInCache)
		end := U.TimeNow()
		logCtx.WithField("timeTaken", end.Sub(begin).Milliseconds()).Info("C:US:Incr")
		if err != nil {
			logCtx.WithError(err).Error("Failed to increment keys")
			return
		}
	}
}

// MergeUserPropertiesForProjectID Used in one time script to back run merge for all existing users.
func MergeUserPropertiesForProjectID(projectID uint64, dryRun bool) int {
	logCtx := log.WithFields(log.Fields{
		"Method":    "MergeUserPropertiesForProjectID",
		"ProjectID": projectID,
	})

	if !isMergeEnabledForProjectID(projectID) {
		logCtx.Info("User properties merge is not enabled for the project")
		return http.StatusNotModified
	}

	customerUserIDs, errCode := GetDistinctCustomerUserIDSForProject(projectID)
	if errCode != http.StatusFound {
		logCtx.Error("Error while getting distinct customer user ids")
		return errCode
	} else if len(customerUserIDs) == 0 {
		return http.StatusNotModified
	} else {
		logCtx.Infof("%d unique customer user ids to be merged", len(customerUserIDs))
	}

	for _, customerUserID := range customerUserIDs {
		// Trigger merge for any one of the users for customer user id.
		logCtx = logCtx.WithFields(log.Fields{"CustomerUserID": customerUserID})
		user, errCode := GetUserLatestByCustomerUserId(projectID, customerUserID)
		if errCode != http.StatusFound {
			logCtx.Infof("Error or no users found for customer user id")
			continue
		}
		_, errCode = MergeUserPropertiesForUserID(projectID, user.ID, postgres.Jsonb{}, "", U.TimeNowUnix(), dryRun, true)
		if errCode != http.StatusCreated && errCode != http.StatusNotModified {
			logCtx.Error("Failed to merge user properties")
		}
	}
	return http.StatusCreated
}

// MergeUserPropertiesForUserID Run's user properties merge using customer_user_id of the given user.
// This will create new properties for all the users of customer_user_id and update the users table
// for all except for the user it is being called for.
func MergeUserPropertiesForUserID(projectID uint64, userID string, updatedProperties postgres.Jsonb,
	currentPropertiesID string, timestamp int64, dryRun bool, updateCalledUser bool) (string, int) {
	logCtx := log.WithFields(log.Fields{
		"Method":    "MergeUserPropertiesForUserID",
		"ProjectID": projectID,
	})
	user, errCode := GetUser(projectID, userID)
	if errCode != http.StatusFound {
		logCtx.Errorf("User not found with id %s", userID)
		return currentPropertiesID, http.StatusInternalServerError
	} else if user.CustomerUserId == "" {
		return currentPropertiesID, http.StatusNotAcceptable
	}
	customerUserID := user.CustomerUserId

	logCtx = logCtx.WithFields(log.Fields{"CustomerUserID": user.CustomerUserId})
	if !isMergeEnabledForProjectID(projectID) {
		logCtx.Infof("User merge properties not enabled for the project")
		return currentPropertiesID, http.StatusNotAcceptable
	}

	// Users are returned in increasing order of created_at. For user_properties created at same unix time,
	// older user order will help in ensuring the order while merging properties.
	users, errCode := GetUsersByCustomerUserID(projectID, customerUserID)
	usersLength := len(users)
	if errCode == http.StatusInternalServerError {
		logCtx.Error("Error while getting users for customer_user_id")
		return currentPropertiesID, http.StatusInternalServerError
	} else if errCode == http.StatusNotFound {
		logCtx.Error("No users found for customer_user_id")
		return currentPropertiesID, http.StatusNotAcceptable
	} else if usersLength == 1 {
		return currentPropertiesID, http.StatusNotAcceptable
	} else if usersLength > 10 {
		logCtx.Infof("User properties merge triggered for more than 10 users. Count: %d", usersLength)
	}

	if usersLength > MaxUsersForPropertiesMerge {
		// If number of users to merge are more than max allowed, merge for oldest max/2 and latest max/2.
		users = append(users[0:MaxUsersForPropertiesMerge/2], users[usersLength-MaxUsersForPropertiesMerge/2:usersLength]...)
	}
	logCtx.Infof("%d users found to be merged for customer user id %s", len(users), customerUserID)

	initialPropertiesVisitedMap := make(map[string]bool)
	for _, property := range U.USER_PROPERTIES_MERGE_TYPE_INITIAL {
		initialPropertiesVisitedMap[property] = false
	}

	var userPropertiesRecords []*UserProperties
	for _, user := range users {
		userPropertiesRecord, errCode := GetUserPropertiesRecord(projectID, user.ID, user.PropertiesId)
		if errCode != http.StatusFound {
			logCtx.Error("Failed to get user properties record for user")
			return currentPropertiesID, http.StatusInternalServerError
		}
		userPropertiesRecords = append(userPropertiesRecords, userPropertiesRecord)
	}

	// Sort user properties records by UpdatedTimestamp in ascending order.
	sort.Slice(userPropertiesRecords, func(i, j int) bool {
		return userPropertiesRecords[i].UpdatedTimestamp < userPropertiesRecords[j].UpdatedTimestamp
	})

	// Append called user's updatedPropertiesMap after sorting to ensure it's always the last.
	userPropertiesRecords = append(userPropertiesRecords, &UserProperties{
		UserId:           userID,
		ProjectId:        projectID,
		Properties:       updatedProperties,
		UpdatedTimestamp: timestamp,
	})

	mergedUserProperties := make(map[string]interface{})
	mergedUserPropertiesValues := make(map[string][]interface{})
	var mergedUpdatedTimestamp int64
	for _, userPropertiesRecord := range userPropertiesRecords {
		userProperties, err := U.DecodePostgresJsonb(&userPropertiesRecord.Properties)
		if err != nil {
			logCtx.WithError(err).Error("Failed to decode user property")
			return currentPropertiesID, http.StatusInternalServerError
		}
		if userPropertiesRecord.UpdatedTimestamp > mergedUpdatedTimestamp {
			mergedUpdatedTimestamp = userPropertiesRecord.UpdatedTimestamp
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
	mergeAddTypeUserProperties(&mergedUserProperties, userPropertiesRecords)

	// Additional check for properties that can be added. If merge is triggered for users with same set of properties,
	// value of properties that can be added will change after addition. Below check is to avoid update in such case.
	if !anyPropertyChanged(mergedUserPropertiesValues, len(users)) {
		logCtx.Infof("Skipping merge as none of the properties changed %s", mergedUserPropertiesValues)
		return currentPropertiesID, http.StatusNotModified
	}
	mergedUserProperties[U.UP_MERGE_TIMESTAMP] = U.TimeNowUnix()
	SanitizeAddTypeProperties(projectID, users, &mergedUserProperties)

	mergedUserPropertiesJSON, err := U.EncodeToPostgresJsonb(&mergedUserProperties)
	if err != nil {
		logCtx.WithError(err).Errorf("Failed to encode merged user properties. %v", mergedUserProperties)
		return currentPropertiesID, http.StatusInternalServerError
	}
	if dryRun {
		logCtx.Infof("DryRun: Merge will be triggered for %d users with customer user id %s. Merged property: %v",
			len(users), customerUserID, mergedUserProperties)
		return currentPropertiesID, http.StatusNotModified
	}

	calledUserNewPropertyID := currentPropertiesID
	for _, user := range users {
		propertyID, errCode := updateUserPropertiesForUser(projectID, user.ID, *mergedUserPropertiesJSON,
			mergedUpdatedTimestamp, user.ID != userID || updateCalledUser)
		// Even if merge failed for some user, return correct status for the called user.
		if errCode != http.StatusCreated {
			if user.ID == userID {
				// Failed for the called user. Return errCode as is.
				return currentPropertiesID, errCode
			} else if calledUserNewPropertyID != currentPropertiesID {
				// Merge failed some user but for called user, it was successfully merged.
				return calledUserNewPropertyID, http.StatusCreated
			}

			return calledUserNewPropertyID, http.StatusNotAcceptable
		}
		if user.ID == userID {
			calledUserNewPropertyID = propertyID
		}
	}
	return calledUserNewPropertyID, http.StatusCreated
}

// updateUserPropertiesForUser Creates new UserProperties entry and updates properties_id in user table. Returns new properties_id.
func updateUserPropertiesForUser(projectID uint64, userID string, userProperties postgres.Jsonb, timestamp int64, updateUser bool) (string, int) {
	logCtx := log.WithFields(log.Fields{
		"Method":    "updateUserPropertiesForUser",
		"ProjectID": projectID,
		"UserID":    userID,
	})
	userPropertiesRecord := UserProperties{
		UserId:           userID,
		ProjectId:        projectID,
		Properties:       userProperties,
		UpdatedTimestamp: timestamp,
	}

	db := C.GetServices().Db
	if err := db.Create(&userPropertiesRecord).Error; err != nil {
		logCtx.WithError(err).Error("Failed to create new user properties")

		// Return bad request to skip retry.
		if U.IsPostgresUnsupportedUnicodeError(err) {
			return "", http.StatusBadRequest
		}
		return "", http.StatusInternalServerError
	}

	if updateUser {
		if err := db.Model(&User{}).Where("project_id = ? AND id = ?", projectID, userID).
			Update("properties_id", userPropertiesRecord.ID).Error; err != nil {

			logCtx.WithError(err).Error("Failed to update propertyID for user")
			return userPropertiesRecord.ID, http.StatusInternalServerError
		}
	}
	return userPropertiesRecord.ID, http.StatusCreated
}

func isEmptyPropertyValue(propertyValue interface{}) bool {
	if propertyValue == nil {
		return true
	}

	// Check only for string empty case.
	// For floats / integers hard to decide whether it was intentionally set as 0.
	switch propertyValue.(type) {
	case string:
		return propertyValue.(string) == ""
	default:
		return false
	}
}

func anyPropertyChanged(propertyValuesMap map[string][]interface{}, numUsers int) bool {
	for property := range propertyValuesMap {
		if len(propertyValuesMap[property]) < numUsers {
			// Some new property was added which is missing for one or more users.
			return true
		} else if len(propertyValuesMap[property]) < 2 {
			continue
		}
		initialValue := propertyValuesMap[property][0]
		for _, propertyValue := range propertyValuesMap[property][1:] {
			if propertyValue != initialValue {
				return true
			}
		}
	}
	return false
}

// Checks if merge is enabled for the project based on global config.
func isMergeEnabledForProjectID(projectID uint64) bool {

	allProjects, mergeEnabledProjectIDsMap, _ := C.GetProjectsFromListWithAllProjectSupport(
		C.GetConfig().MergeUspProjectIds, "")
	if allProjects {
		return true
	}
	if _, ok := mergeEnabledProjectIDsMap[projectID]; ok {
		return true
	}
	return false
}

// addValuesForProperty To add old and new value for the user property type add.
// Adding 0.1 + 0.2 will result in 0.30000000000000004 as explained https://floating-point-gui.de/
// Round off values with precision to avoid this.
func addValuesForProperty(oldValue interface{}, newValue float64, addOld bool) float64 {
	var addedValue float64
	var err error
	if addOld {
		addedValue, err = U.FloatRoundOffWithPrecision(oldValue.(float64)+newValue, 2)
		if err != nil {
			// If error in round off, use as is.
			addedValue = oldValue.(float64) + newValue
		}
	} else {
		addedValue, err = U.FloatRoundOffWithPrecision(newValue, 2)
		if err != nil {
			addedValue = newValue
		}
	}
	return addedValue
}

// Initializes merged properties with the one being updated which will be the last of `userPropertiesRecords`.
// Now for every user, add value if:
//     1. Not set already from the one being updated.
//     2. User property is not merged before i.e. $merge_timestamp is not set.
//     3. Value is greater than the value already set. Add the difference then. (This should ideally not happen)
func mergeAddTypeUserProperties(mergedProperties *map[string]interface{}, userPropertiesRecords []*UserProperties) {
	// Last record in the array would be the latest one.
	latestPropertiesRecord := userPropertiesRecords[len(userPropertiesRecords)-1]
	latestPropertiesMap, err := U.DecodePostgresJsonb(&latestPropertiesRecord.Properties)
	if err != nil {
		log.WithError(err).Error("Failed to decode user property")
		return
	}

	// Boolean map to indicate whether merged value is used at least once.
	mergedValueAddedOnce := make(map[string]bool)
	for _, property := range U.USER_PROPERTIES_MERGE_TYPE_ADD {
		mergedValueAddedOnce[property] = false
	}

	// Cases to consider:
	//    1. What if latestPropertiesMap has one of add type property missing? Add full on first encounter. And add diff after that.
	//    2. Already merged property with value more than latestProperty value? Add difference.
	//    3. Already merged property with value less than latestProperty value? Do nothing.
	//    4. Is not a merged property. Probably for a new user? Add full as is.
	//    5. Does ordering matter while parsing non latest properties? No.
	for _, property := range U.USER_PROPERTIES_MERGE_TYPE_ADD {
		if _, found := (*latestPropertiesMap)[property]; !found {
			continue
		}
		(*mergedProperties)[property] = (*latestPropertiesMap)[property]
		if _, isLatestMerged := (*latestPropertiesMap)[U.UP_MERGE_TIMESTAMP]; isLatestMerged {
			// Since latest properties is also a merged property, set mergedValueAddedOnce true
			// to avoid another merged property getting added which would double the value otherwise.
			mergedValueAddedOnce[property] = true
		}
	}

	// Loop over all records except last record.
	for _, userPropertiesRecord := range userPropertiesRecords[:len(userPropertiesRecords)-1] {
		userProperties, err := U.DecodePostgresJsonb(&userPropertiesRecord.Properties)
		if err != nil {
			log.WithError(err).Error("Failed to decode user property")
			return
		}

		_, isMergedBefore := (*userProperties)[U.UP_MERGE_TIMESTAMP]
		for _, property := range U.USER_PROPERTIES_MERGE_TYPE_ADD {
			mergedValue, mergedExists := (*mergedProperties)[property]
			userValue, userValueExists := (*userProperties)[property]
			if isMergedBefore {
				if !mergedValueAddedOnce[property] && userValueExists {
					// Merged values must be added full at least once. Since not added already, add full here.
					(*mergedProperties)[property] = addValuesForProperty(mergedValue, userValue.(float64), mergedExists)
					mergedValueAddedOnce[property] = true
				} else if mergedExists && userValueExists && userValue.(float64)-mergedValue.(float64) > 0 {
					// Add the difference of values to mergedValues.
					(*mergedProperties)[property] = addValuesForProperty(mergedValue, userValue.(float64)-mergedValue.(float64), true)
				} else if userValueExists && !mergedExists && !mergedValueAddedOnce[property] {
					// mergedValue does not exists. Which means this property was not present in the latest or has
					// not been added so far. Add the values as is to initialize.
					(*mergedProperties)[property] = addValuesForProperty(0, userValue.(float64), false)
					mergedValueAddedOnce[property] = true
				}
			} else if userValueExists {
				(*mergedProperties)[property] = addValuesForProperty(mergedValue, (*userProperties)[property].(float64), mergedExists)
			}
		}
	}
}

// SanitizeAddTypeProperties To fix bad values for add type properties like $page_count, $session_count.
//   1. Counts all sessions for users of customer_user_id and set it as session_count.
//   2. Generate random value from 1 to 5 * session_count and set as page_count.
//   3. Generate random value from 1 to 5 min * session_count and set as session_spent_time.
// TODO(prateek): Remove once older values are fixed using script.
func SanitizeAddTypeProperties(projectID uint64, users []User, propertiesMap *map[string]interface{}) {
	logCtx := log.WithFields(log.Fields{
		"Method":    "SanitizeAddTypeProperties",
		"ProjectID": projectID,
	})
	var userIDs []string
	for _, user := range users {
		userIDs = append(userIDs, user.ID)
	}

	var propertyValue interface{}
	var found bool
	propertyValue, found = (*propertiesMap)[U.UP_SESSION_COUNT]
	if !found {
		propertyValue, found = (*propertiesMap)[U.UP_PAGE_COUNT]
		if !found {
			return
		}
	}

	acceptablePropertyLength := 3 // Up to 999.
	if len(fmt.Sprint(propertyValue)) <= acceptablePropertyLength {
		return
	}
	sessionEvent, errCode := GetSessionEventName(projectID)
	if errCode != http.StatusFound {
		return
	}
	sessionCount, errCode := GetEventCountOfUsersByEventName(projectID, userIDs, sessionEvent.ID)
	if errCode != http.StatusFound {
		return
	}

	if _, found := (*propertiesMap)[U.UP_SESSION_COUNT]; found {
		sanitizedValue := float64(sessionCount)
		logCtx.Infof("Updating value for $session_count from %v to %v", (*propertiesMap)[U.UP_SESSION_COUNT], sanitizedValue)
		(*propertiesMap)[U.UP_SESSION_COUNT] = sanitizedValue
	}
	if _, found := (*propertiesMap)[U.UP_PAGE_COUNT]; found {
		sanitizedValue := float64(sessionCount * uint64(U.RandomIntInRange(1, 5))) // 1 to 5 pages.
		logCtx.Infof("Updating value for $page_count from %v to %v", (*propertiesMap)[U.UP_PAGE_COUNT], sanitizedValue)
		(*propertiesMap)[U.UP_PAGE_COUNT] = sanitizedValue
	}
	if _, found := (*propertiesMap)[U.UP_TOTAL_SPENT_TIME]; found {
		sanitizedValue := float64(sessionCount * uint64(U.RandomIntInRange(60, 300))) // 1 to 5 mins.
		logCtx.Infof("Updating value for $session_spent_time from %v to %v", (*propertiesMap)[U.UP_TOTAL_SPENT_TIME], sanitizedValue)
		(*propertiesMap)[U.UP_TOTAL_SPENT_TIME] = sanitizedValue
	}
}

func GetUserProperties(projectId uint64, userId string, id string) (*postgres.Jsonb, int) {
	userPropertiesRecord, errCode := GetUserPropertiesRecord(projectId, userId, id)
	if userPropertiesRecord == nil {
		return nil, errCode
	}

	return &userPropertiesRecord.Properties, http.StatusFound
}

func GetUserPropertiesRecord(projectId uint64, userId string, id string) (*UserProperties, int) {
	logCtx := log.WithFields(log.Fields{"project_id": projectId, "user_id": userId, "user_properties_id": id})

	var userProperties UserProperties

	db := C.GetServices().Db
	if err := db.Where("project_id = ?", projectId).Where("user_id = ?", userId).Where(
		"id = ?", id).First(&userProperties).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}
		logCtx.WithError(err).Error("Getting user properties record using projectId, userId, userPropertiesId failed")
		return nil, http.StatusInternalServerError
	}

	return &userProperties, http.StatusFound
}

func FillLocationUserProperties(properties *U.PropertiesMap, clientIP string) error {
	geo := C.GetServices().GeoLocation

	// ClientIP unavailable.
	if clientIP == "" {
		return fmt.Errorf("invalid IP, failed adding geolocation properties")
	}

	city, err := geo.City(net.ParseIP(clientIP))
	if err != nil {
		log.WithFields(log.Fields{"clientIP": clientIP}).WithError(err).Error(
			"Failed to get city information from geodb")
		return err
	}

	// Using en -> english name.
	if countryName, ok := city.Country.Names["en"]; ok && countryName != "" {
		if c, ok := (*properties)[U.UP_COUNTRY]; !ok || c == "" {
			(*properties)[U.UP_COUNTRY] = countryName
		}
	}

	if cityName, ok := city.City.Names["en"]; ok && cityName != "" {
		if c, ok := (*properties)[U.UP_CITY]; !ok || c == "" {
			(*properties)[U.UP_CITY] = cityName
		}
	}

	return nil
}

func fillUserPropertiesFromFormSubmitEventProperties(properties *U.PropertiesMap,
	formSubmitProperties *U.PropertiesMap) {

	for k, v := range *formSubmitProperties {
		if U.IsFormSubmitUserProperty(k) {
			(*properties)[k] = v
		}
	}
}

func FillUserPropertiesAndGetCustomerUserIdFromFormSubmit(projectId uint64, userId string,
	properties, formSubmitProperties *U.PropertiesMap) (string, int) {

	logCtx := log.WithFields(log.Fields{"project_id": projectId, "user_id": userId})

	user, errCode := GetUser(projectId, userId)
	if errCode != http.StatusFound {
		logCtx.Error("Failed to get latest user properties on fill form submitted properties.")
		return "", http.StatusInternalServerError
	}

	logCtx = logCtx.WithFields(log.Fields{"existing_user_properties": user.Properties,
		"form_event_properties": formSubmitProperties})

	userProperties, err := U.DecodePostgresJsonb(&user.Properties)
	if err != nil {
		logCtx.Error("Failed to decoding latest user properties on fill form submitted properties.")
	}

	formPropertyEmail, formPropertyEmailExists := (*formSubmitProperties)[U.UP_EMAIL]
	userPropertyEmail, userPropertyEmailExists := (*userProperties)[U.UP_EMAIL]

	formPropertyPhone, formPropertyPhoneExists := (*formSubmitProperties)[U.UP_PHONE]
	userPropertyPhone, userPropertyPhoneExists := (*userProperties)[U.UP_PHONE]

	if formPropertyEmailExists && userPropertyEmailExists {
		if userPropertyEmail != formPropertyEmail {
			logCtx.WithField("tag", "different_email_seen_for_customer_user_id").
				Warn("Different email seen on form event. User property not updated.")
			return "", http.StatusBadRequest
		}

		// form event email is same as user properties, update other user properties.
		fillUserPropertiesFromFormSubmitEventProperties(properties, formSubmitProperties)
		return U.GetPropertyValueAsString(formPropertyEmail), http.StatusOK
	}

	if formPropertyPhoneExists && userPropertyPhoneExists {
		if userPropertyPhone != formPropertyPhone {
			logCtx.WithField("tag", "different_phone_seen_for_customer_user_id").
				Warn("Different phone seen on form event. User property not updated.")
			return "", http.StatusBadRequest
		}

		// form event phone is same as user propertie, update other user properties.
		fillUserPropertiesFromFormSubmitEventProperties(properties, formSubmitProperties)
		return U.GetPropertyValueAsString(formPropertyPhone), http.StatusOK
	}

	if !formPropertyEmailExists && !formPropertyPhoneExists {
		return "", http.StatusBadRequest
	}

	var identity string
	if formPropertyEmailExists {
		identity = U.GetPropertyValueAsString(formPropertyEmail)
	} else if formPropertyPhoneExists {
		identity = U.GetPropertyValueAsString(formPropertyPhone)
	}

	fillUserPropertiesFromFormSubmitEventProperties(properties, formSubmitProperties)
	return identity, http.StatusOK
}

func GetUserPropertyRecordsByUserId(projectId uint64, userId string) ([]UserProperties, int) {
	db := C.GetServices().Db
	logCtx := log.WithFields(log.Fields{"project_id": projectId, "user_id": userId})

	var userProperties []UserProperties
	if err := db.Where("project_id = ? AND user_id = ?", projectId, userId).Find(&userProperties).Error; err != nil {
		logCtx.WithError(err).Error("Getting user property records by user_id failed")
		return nil, http.StatusInternalServerError
	}

	if len(userProperties) == 0 {
		return nil, http.StatusNotFound
	}

	return userProperties, http.StatusFound
}

func OverwriteUserProperties(projectId uint64, userId string,
	id string, propertiesJsonb *postgres.Jsonb) int {

	if projectId == 0 || userId == "" || id == "" {
		return http.StatusBadRequest
	}

	propertiesJsonb = U.SanitizePropertiesJsonb(propertiesJsonb)
	db := C.GetServices().Db
	if err := db.Model(&UserProperties{}).
		Where("project_id = ? AND user_id = ? AND id = ?", projectId, userId, id).
		Update("properties", propertiesJsonb).Error; err != nil {

		log.WithFields(log.Fields{"project_id": projectId, "id": id}).
			WithError(err).Error("Failed to replace properties.")
		return http.StatusInternalServerError
	}

	return http.StatusAccepted
}

// Updates given property with value on all user properties for the given user
// adds property if not exist.
func UpdatePropertyOnAllUserPropertyRecords(projectId uint64, userId string,
	property string, value interface{}) int {

	userPropertyRecords, errCode := GetUserPropertyRecordsByUserId(projectId, userId)
	if errCode == http.StatusInternalServerError {
		return errCode
	} else if errCode == http.StatusNotFound {
		return http.StatusBadRequest
	}

	logCtx := log.WithFields(log.Fields{"project_id": projectId, "user_id": userId})

	for _, userProperties := range userPropertyRecords {
		var propertiesMap map[string]interface{}

		if !U.IsEmptyPostgresJsonb(&userProperties.Properties) {
			err := json.Unmarshal(userProperties.Properties.RawMessage, &propertiesMap)
			if err != nil {
				logCtx.Error("Failed to update user property record. JSON unmarshal failed.")
				continue
			}
		} else {
			propertiesMap = make(map[string]interface{}, 0)
		}

		// update not required. Not using AddToPostgresJsonb
		// for this check.
		if pValue, _ := propertiesMap[property]; pValue == value {
			continue
		}

		logCtx = logCtx.WithFields(log.Fields{"properties_id": userProperties.ID, "property": property, "value": value})

		propertiesMap[property] = value
		properitesBytes, err := json.Marshal(propertiesMap)
		if err != nil {
			// log and continue update next user property.
			logCtx.Error("Failed to update user property record. JSON marshal failed.")
			continue
		}
		updatedProperties := postgres.Jsonb{RawMessage: json.RawMessage(properitesBytes)}

		// Triggers multiple updates.
		errCode := OverwriteUserProperties(projectId, userId, userProperties.ID, &updatedProperties)
		if errCode == http.StatusInternalServerError {
			logCtx.WithError(err).Error("Failed to update user property record. DB query failed.")
			continue
		}
	}

	return http.StatusAccepted
}

func GetUserPropertiesRecordsByProperty(projectId uint64,
	key string, value interface{}) ([]UserProperties, int) {

	logCtx := log.WithField("project_id", projectId).WithField(
		"key", key).WithField("value", value)

	db := C.GetServices().Db
	var userProperties []UserProperties
	// $$$ is a gorm alias for ? jsonb operator.
	err := db.Order("created_at").Where("project_id=?", projectId).Where(
		"properties->? $$$ ?", key, value).Find(&userProperties).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to get user properties by key.")
		return nil, http.StatusInternalServerError
	}

	if len(userProperties) == 0 {
		return nil, http.StatusNotFound
	}

	return userProperties, http.StatusFound
}

func UpdateUserPropertiesForSession(projectID uint64,
	sessionUserPropertiesRecordMap *map[string]SessionUserProperties) int {

	logCtx := log.WithField("project_id", projectID)

	hasFailure := false
	for userPropertiesID, sessionUserProperties := range *sessionUserPropertiesRecordMap {
		logCtx.WithField("user_properties_id", userPropertiesID)

		userProperties, errCode := GetUserPropertiesRecord(projectID,
			sessionUserProperties.UserID, userPropertiesID)
		if errCode != http.StatusFound {
			logCtx.WithField("err_code", errCode).
				Error("Failed to get user properties record.")
			hasFailure = true
			continue
		}

		userPropertiesMap, err := U.DecodePostgresJsonb(&userProperties.Properties)
		if err != nil {
			logCtx.WithError(err).
				Error("Failed to decode user properties on UpdateUserPropertiesForSession.")
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

		(*userPropertiesMap)[U.UP_PAGE_COUNT] = existingPageCount + sessionUserProperties.SessionPageCount
		(*userPropertiesMap)[U.UP_TOTAL_SPENT_TIME] = existingTotalSpentTime + sessionUserProperties.SessionPageSpentTime
		(*userPropertiesMap)[U.UP_SESSION_COUNT] = sessionUserProperties.SessionCount

		userPropertiesJsonb, err := U.EncodeToPostgresJsonb(userPropertiesMap)
		if err != nil {
			logCtx.WithError(err).
				Error("Failed to encode user properties json after adding new session count.")
			hasFailure = true
			continue
		}

		errCode = OverwriteUserProperties(projectID, sessionUserProperties.UserID,
			userPropertiesID, userPropertiesJsonb)
		if errCode != http.StatusAccepted {
			logCtx.WithField("err_code", errCode).
				Error("Failed to overwrite user properties record.")
			hasFailure = true
			continue
		}
	}

	if hasFailure {
		return http.StatusInternalServerError
	}

	return http.StatusAccepted
}
