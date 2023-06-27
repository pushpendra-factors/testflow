package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"reflect"
	"sort"
	"strings"
	"time"

	cacheRedis "factors/cache/redis"
	"factors/config"
	"factors/util"
	U "factors/util"

	"github.com/gomodule/redigo/redis"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

type User struct {
	// Composite primary key with project_id and random uuid.
	ID string `gorm:"primary_key:true;uuid;default:uuid_generate_v4()" json:"id"`
	// Below are the foreign key constraints added in creation script.
	// project_id -> projects(id)
	ProjectId                  int64          `gorm:"primary_key:true;" json:"project_id"`
	Properties                 postgres.Jsonb `json:"properties"`
	PropertiesUpdatedTimestamp int64          `json:"properties_updated_timestamp"`
	SegmentAnonymousId         string         `gorm:"type:varchar(200);default:null" json:"seg_aid"`
	AMPUserId                  string         `gorm:"default:null" json:"amp_user_id"`
	// Avoid updating group field tags
	IsGroupUser  *bool  `gorm:"default:null" json:"is_group_user"`
	Group1ID     string `gorm:"default:null;column:group_1_id" json:"group_1_id"`
	Group1UserID string `gorm:"default:null;column:group_1_user_id" json:"group_1_user_id"`
	Group2ID     string `gorm:"default:null;column:group_2_id" json:"group_2_id"`
	Group2UserID string `gorm:"default:null;column:group_2_user_id" json:"group_2_user_id"`
	Group3ID     string `gorm:"default:null;column:group_3_id" json:"group_3_id"`
	Group3UserID string `gorm:"default:null;column:group_3_user_id" json:"group_3_user_id"`
	Group4ID     string `gorm:"default:null;column:group_4_id" json:"group_4_id"`
	Group4UserID string `gorm:"default:null;column:group_4_user_id" json:"group_4_user_id"`
	Group5ID     string `gorm:"default:null;column:group_5_id" json:"group_5_id"`
	Group5UserID string `gorm:"default:null;column:group_5_user_id" json:"group_5_user_id"`
	Group6ID     string `gorm:"default:null;column:group_6_id" json:"group_6_id"`
	Group6UserID string `gorm:"default:null;column:group_6_user_id" json:"group_6_user_id"`
	Group7ID     string `gorm:"default:null;column:group_7_id" json:"group_7_id"`
	Group7UserID string `gorm:"default:null;column:group_7_user_id" json:"group_7_user_id"`
	Group8ID     string `gorm:"default:null;column:group_8_id" json:"group_8_id"`
	Group8UserID string `gorm:"default:null;column:group_8_user_id" json:"group_8_user_id"`
	// UserId provided by the customer.
	// An unique index is creatd on ProjectId+UserId.
	CustomerUserId       string `gorm:"type:varchar(255);default:null" json:"c_uid"`
	CustomerUserIdSource *int   `gorm:"default:null" json:"c_uid_source"`
	// unix epoch timestamp in seconds.
	JoinTimestamp int64     `json:"join_timestamp"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	// source of the user record (1 = WEB, 2 = HUBSPOT, 3 = SALESFORCE)
	Source         *int           `json:"source"`
	EventAggregate postgres.Jsonb `json:"event_aggregate"`
}

type LatestUserPropertiesFromSession struct {
	PageCount      float64
	TotalSpentTime float64
	SessionCount   uint64
	Timestamp      int64
	InitialChannel string
	LatestChannel  string
}

type SessionUserProperties struct {
	// Meta
	UserID                string
	SessionEventTimestamp int64

	// Current event user properties.
	EventUserProperties *postgres.Jsonb

	// Properties
	SessionCount         uint64
	SessionPageCount     float64
	SessionPageSpentTime float64
	SessionChannel       string
	IsSessionEvent       bool
}
type ChannelUserProperty struct {
	InitialChannel string
	LatestChannel  string
}

// indexed hubspot user property.
const UserPropertyHubspotContactLeadGUID = "$hubspot_contact_lead_guid"

// contact delete and merge hubspot user properties
const UserPropertyHubspotContactDeleted = "$hubspot_contact_deleted"
const UserPropertyLeadSquaredLeadDeleted = "$leadsquared_lead_is_deleted"
const UserPropertyHubspotContactMerged = "$hubspot_contact_merged"
const UserPropertyHubspotContactPrimaryContact = "$hubspot_contact_primary_contact"

var UserPropertiesToSkipOnMergeByCustomerUserID = []string{
	UserPropertyHubspotContactLeadGUID,
	U.UP_META_OBJECT_IDENTIFIER_KEY,
	UserPropertyHubspotContactDeleted,
	UserPropertyHubspotContactMerged,
	UserPropertyHubspotContactPrimaryContact,
	U.UP_SESSION_COUNT,
	UserPropertyLeadSquaredLeadDeleted,
}

var ErrDifferentEmailSeen error = errors.New("different_email_seen_for_customer_user_id")
var ErrDifferentPhoneNoSeen error = errors.New("different_phone_no_seen_for_customer_user_id")

const MaxUsersForPropertiesMerge = 100

// IdentifyMeta holds data for overwriting customer_user_id
type IdentifyMeta struct {
	Timestamp int64  `json:"timestamp"`
	PageURL   string `json:"page_url,omitempty"`
	Source    string `json:"source"`
}

// UserPropertiesMeta is a map for customer_user_id to IdentifyMeta
type UserPropertiesMeta map[string]IdentifyMeta

const (
	UserSourceWeb                   = 1
	UserSourceHubspot               = 2
	UserSourceSalesforce            = 3
	UserSourceHubspotString         = "hubspot"
	UserSourceSalesforceString      = "salesforce"
	UserSourceMarketo               = "marketo"
	UserSourceLeadSquared           = U.CRM_SOURCE_NAME_LEADSQUARED
	UserSourceSixSignalString       = "6signal"
	UserSourceSixSignal             = 8
	UserSourceDomainsString         = "domains"
	UserSourceDomains               = 9
	UserSourceLinkedinCompany       = 10
	UserSourceLinkedinCompanyString = "linkedin_company"
	UserSourceG2                    = 11
	UserSourceG2String              = "g2"
)

var UserSourceMap = map[string]int{
	"web":                           1,
	UserSourceHubspotString:         2,
	UserSourceSalesforceString:      3,
	UserSourceMarketo:               6,
	UserSourceLeadSquared:           7,
	UserSourceSixSignalString:       UserSourceSixSignal,
	UserSourceDomainsString:         UserSourceDomains,
	UserSourceLinkedinCompanyString: UserSourceLinkedinCompany,
	UserSourceG2String:              UserSourceG2,
}

var UserSourceCRM = map[string]int{
	UserSourceHubspotString:    2,
	UserSourceSalesforceString: 3,
	UserSourceMarketo:          6,
	UserSourceLeadSquared:      7,
}

var GroupUserSource = map[string]int{
	U.GROUP_NAME_SALESFORCE_ACCOUNT:     UserSourceSalesforce,
	U.GROUP_NAME_SALESFORCE_OPPORTUNITY: UserSourceSalesforce,
	U.GROUP_NAME_HUBSPOT_COMPANY:        UserSourceHubspot,
	U.GROUP_NAME_HUBSPOT_DEAL:           UserSourceHubspot,
	U.GROUP_NAME_SIX_SIGNAL:             UserSourceSixSignal,
	U.GROUP_NAME_DOMAINS:                UserSourceDomains,
	U.GROUP_NAME_LINKEDIN_COMPANY:       UserSourceLinkedinCompany,
	U.GROUP_NAME_G2:                     UserSourceG2,
}

const USERS = "users"

type OverwriteUserPropertiesByIDParams struct {
	ProjectID           int64
	UserID              string
	UserProperties      *postgres.Jsonb
	WithUpdateTimestamp bool
	UpdateTimestamp     int64
	Source              string
}

var AccountGroupAssociationPrecedence = []string{
	GROUP_NAME_SALESFORCE_ACCOUNT,
	GROUP_NAME_HUBSPOT_COMPANY,
	GROUP_NAME_SIX_SIGNAL,
}

func IsUserSourceCRM(source int) bool {
	for _, crmSource := range UserSourceCRM {
		if crmSource == source {
			return true
		}
	}
	return false
}

func GetAllCRMUserSource() []int {
	sources := make([]int, 0)
	for i := range UserSourceCRM {
		sources = append(sources, UserSourceCRM[i])
	}

	return sources
}

func FilterGroupPropertiesFromUserProperties(properties map[string]interface{}) map[string]interface{} {
	nonGroupProperties := make(map[string]interface{})
	for key, value := range properties {
		if U.DisableGroupUserPropertiesByKeyPrefix(key) {
			continue
		}
		nonGroupProperties[key] = value
	}

	return nonGroupProperties
}

func GetOverwriteUserPropertiesByIDParamsInBatch(list []OverwriteUserPropertiesByIDParams,
	batchSize int) [][]OverwriteUserPropertiesByIDParams {
	batchList := make([][]OverwriteUserPropertiesByIDParams, 0, 0)
	listLen := len(list)
	for i := 0; i < listLen; {
		next := i + batchSize
		if next > listLen {
			next = listLen
		}

		batchList = append(batchList, list[i:next])
		i = next
	}

	return batchList
}

func GetRequestSourcePointer(requestSource int) *int {
	var requestSourcePointer = requestSource
	return &requestSourcePointer
}

func GetIdentifiedUserProperties(customerUserId string) (map[string]interface{}, error) {
	if customerUserId == "" {
		return nil, errors.New("invalid customer user id")
	}

	properties := map[string]interface{}{
		U.UP_USER_ID: customerUserId,
	}

	if U.IsEmail(customerUserId) {
		properties[U.UP_EMAIL] = customerUserId
	}

	return properties, nil
}

func GetIdentifiedUserPropertiesAsJsonb(customerUserId string) (*postgres.Jsonb, error) {
	properties, err := GetIdentifiedUserProperties(customerUserId)
	if err != nil {
		return nil, err
	}

	return U.EncodeToPostgresJsonb(&properties)
}

// Today's cache keys
func GetUsersCachedCacheKey(projectId int64, dateKey string) (*cacheRedis.Key, error) {
	prefix := "US:LIST"
	return cacheRedis.NewKey(projectId, prefix, dateKey)
}

func GetUserPropertiesCategoryByProjectCacheKey(projectId int64, property string, category string, dateKey string) (*cacheRedis.Key, error) {
	prefix := "US:PC"
	return cacheRedis.NewKey(projectId, prefix, fmt.Sprintf("%s:%s:%s", dateKey, category, property))

}

func GetValuesByUserPropertyCacheKey(projectId int64, property_name string, value string, dateKey string) (*cacheRedis.Key, error) {
	prefix := "US:PV"
	return cacheRedis.NewKey(projectId, fmt.Sprintf("%s:%s", prefix, property_name), fmt.Sprintf("%s:%s", dateKey, value))
}

// sorted sets
func GetUserPropertiesCategoryByProjectCacheKeySortedSet(projectId int64, dateKey string) (*cacheRedis.Key, error) {
	prefix := "SS:US:PC"
	return cacheRedis.NewKey(projectId, prefix, fmt.Sprintf("%s", dateKey))

}

func GetValuesByUserPropertyCacheKeySortedSet(projectId int64, dateKey string) (*cacheRedis.Key, error) {
	prefix := "SS:US:PV"
	return cacheRedis.NewKey(projectId, fmt.Sprintf("%s", prefix), fmt.Sprintf("%s", dateKey))
}

// Rollup cache keys
func GetUserPropertiesCategoryByProjectRollUpCacheKey(projectId int64, dateKey string) (*cacheRedis.Key, error) {
	prefix := "RollUp:US:PC"
	return cacheRedis.NewKey(projectId, prefix, dateKey)
}

func GetValuesByUserPropertyRollUpCacheKey(projectId int64, property_name string, dateKey string) (*cacheRedis.Key, error) {
	prefix := "RollUp:US:PV"
	return cacheRedis.NewKey(projectId, fmt.Sprintf("%s:%s", prefix, property_name), dateKey)
}

// Today's cache keys count
func GetUserPropertiesCategoryByProjectCountCacheKey(projectId int64, dateKey string) (*cacheRedis.Key, error) {
	prefix := "C:US:PC"
	return cacheRedis.NewKeyWithAllProjectsSupport(projectId, prefix, dateKey)
}

func GetValuesByUserPropertyCountCacheKey(projectId int64, dateKey string) (*cacheRedis.Key, error) {
	prefix := "C:US:PV"
	return cacheRedis.NewKeyWithAllProjectsSupport(projectId, prefix, dateKey)
}

func GetUpdatedPhoneNoFromFormSubmit(formPropertyPhone, userPropertyPhone string) (string, error) {
	if userPropertyPhone != formPropertyPhone {
		if userPropertyPhone == "" {
			return U.SanitizePhoneNumber(formPropertyPhone), nil
		}

		if formPropertyPhone == "" {
			return userPropertyPhone, nil
		}

		sanitizedPhoneNumber := U.SanitizePhoneNumber(formPropertyPhone)
		if shouldAllowCustomerUserID(U.GetPropertyValueAsString(userPropertyPhone), sanitizedPhoneNumber) {
			return sanitizedPhoneNumber, ErrDifferentPhoneNoSeen
		}

		return "", nil
	}

	return formPropertyPhone, nil
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

func GetUpdatedEmailFromFormSubmit(formPropertyEmail, userPropertyEmail string) (string, error) {
	lowerCaseformPropertyEmail := U.GetEmailLowerCase(formPropertyEmail)
	lowerCaseUserPropertyEmail := U.GetEmailLowerCase(userPropertyEmail)

	if lowerCaseUserPropertyEmail != lowerCaseformPropertyEmail {

		if lowerCaseUserPropertyEmail == "" {
			return lowerCaseformPropertyEmail, nil
		}

		if lowerCaseformPropertyEmail == "" {
			return lowerCaseUserPropertyEmail, nil
		}

		// avoid free email update
		if !shouldAllowCustomerUserID(U.GetPropertyValueAsString(lowerCaseUserPropertyEmail), lowerCaseformPropertyEmail) {
			return "", ErrDifferentEmailSeen
		}

		return lowerCaseformPropertyEmail, ErrDifferentEmailSeen
	}

	return lowerCaseformPropertyEmail, nil
}

func GetUserPropertiesFromFormSubmitEventProperties(formSubmitProperties *U.PropertiesMap) *U.PropertiesMap {
	properties := make(U.PropertiesMap)
	for k, v := range *formSubmitProperties {
		if U.IsFormSubmitUserProperty(k) {
			if k == U.UP_EMAIL {
				email := U.GetEmailLowerCase(v)
				if email != "" {
					properties[k] = email
				}

			} else if k == U.UP_PHONE {
				sPhoneNo := U.SanitizePhoneNumber(v)
				if sPhoneNo != "" {
					properties[k] = sPhoneNo
				}

			} else {
				properties[k] = v
			}
		}
	}
	return &properties
}

func AnyPropertyChanged(propertyValuesMap map[string][]interface{}, numUsers int) bool {
	for property := range propertyValuesMap {
		if len(propertyValuesMap[property]) < numUsers {
			// Some new property was added which is missing for one or more users.
			return true
		} else if len(propertyValuesMap[property]) < 2 {
			continue
		}
		initialValue := propertyValuesMap[property][0]
		for _, propertyValue := range propertyValuesMap[property][1:] {
			if fmt.Sprintf("%v", propertyValue) != fmt.Sprintf("%v", initialValue) {
				return true
			}
		}
	}
	return false
}

// Initializes merged properties with the one being updated which will be the last of `userPropertiesRecords`.
// Now for every user, add value if:
//  1. Not set already from the one being updated.
//  2. User property is not merged before i.e. $merge_timestamp is not set.
//  3. Value is greater than the value already set. Add the difference then. (This should ideally not happen)
func MergeAddTypeUserProperties(mergedProperties *map[string]interface{}, existingProperties []postgres.Jsonb) {
	// Last record in the array would be the latest one.
	latestProperties := existingProperties[len(existingProperties)-1]
	latestPropertiesMap, err := U.DecodePostgresJsonb(&latestProperties)
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
	for _, userPropertiesRecord := range existingProperties[:len(existingProperties)-1] {
		userProperties, err := U.DecodePostgresJsonb(&userPropertiesRecord)
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

func IsEmptyPropertyValue(propertyValue interface{}) bool {
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

func FillLocationUserProperties(properties *util.PropertiesMap, clientIP string) (error, string) {
	geo := config.GetServices().GeoLocation

	country := ""
	// ClientIP unavailable.
	if clientIP == "" {
		return fmt.Errorf("invalid IP, failed adding geolocation properties"), country
	}

	city, err := geo.City(net.ParseIP(clientIP))
	if err != nil {
		log.WithFields(log.Fields{"clientIP": clientIP}).WithError(err).Error(
			"Failed to get city information from geodb")
		return err, country
	}

	// Using en -> english name.
	if countryName, ok := city.Country.Names["en"]; ok && countryName != "" {
		if c, ok := (*properties)[util.UP_COUNTRY]; !ok || c == "" {
			(*properties)[util.UP_COUNTRY] = countryName
			country = countryName
		}
	}

	if cityName, ok := city.City.Names["en"]; ok && cityName != "" {
		if c, ok := (*properties)[util.UP_CITY]; !ok || c == "" {
			(*properties)[util.UP_CITY] = cityName
		}
	}
	if continentName, ok := city.Continent.Names["en"]; ok && continentName != "" {
		if c, ok := (*properties)[util.UP_CONTINENT]; !ok || c == "" {
			(*properties)[util.UP_CONTINENT] = continentName
		}
	}
	postalCode := city.Postal.Code
	if postalCode != "" {
		if c, ok := (*properties)[util.UP_POSTAL_CODE]; !ok || c == "" {
			(*properties)[util.UP_POSTAL_CODE] = postalCode
		}
	}

	return nil, country
}

// GetDecodedUserPropertiesIdentifierMetaObject gets the identifier meta data from the user properties
func GetDecodedUserPropertiesIdentifierMetaObject(existingUserProperties *map[string]interface{}) (*UserPropertiesMeta, error) {
	metaObj := make(UserPropertiesMeta)
	if existingUserProperties == nil {
		return &metaObj, nil
	}

	intMetaObj, exists := (*existingUserProperties)[util.UP_META_OBJECT_IDENTIFIER_KEY]
	if !exists {
		return &metaObj, nil
	}

	metaObjMap, ok := intMetaObj.(map[string]interface{})
	if !ok {
		err := json.Unmarshal([]byte(util.GetPropertyValueAsString(intMetaObj)), &metaObj)
		if err != nil {
			log.WithError(err).Errorf("Failed to get meta data from user properties")
		}
		return &metaObj, err
	}

	var enMetaObj []byte
	enMetaObj, err := json.Marshal(metaObjMap)
	if err != nil {
		log.WithError(err).Errorf("Failed to encode meta data from user properties")
		return &metaObj, err
	}

	err = json.Unmarshal(enMetaObj, &metaObj)
	if err != nil {
		log.WithError(err).Errorf("Failed to unmarshal meta data from user properties")
	}

	return &metaObj, err
}

// UpdateUserPropertiesIdentifierMetaObject overwrites the identifier meta date in the user properties
func UpdateUserPropertiesIdentifierMetaObject(userProperties *postgres.Jsonb, metaObj *UserPropertiesMeta) error {
	if metaObj == nil {
		return errors.New("invalid meta object")
	}

	userPropertiesMap, err := U.DecodePostgresJsonbAsPropertiesMap(userProperties)
	if err != nil {
		return err
	}

	(*userPropertiesMap)[U.UP_META_OBJECT_IDENTIFIER_KEY] = *metaObj

	newUserProperties, err := U.EncodeStructTypeToPostgresJsonb(userPropertiesMap)
	if err != nil {
		return err
	}

	*userProperties = *newUserProperties
	return nil
}

func getCacheKeyForUserIDByAMPUserID(projectID int64, ampUserID string) (*cacheRedis.Key, error) {
	return cacheRedis.NewKey(projectID, "users_ampid_id", ampUserID)
}

func GetCacheUserIDByAMPUserID(projectID int64, ampUserID string) (string, int) {
	logCtx := log.WithField("project_id", projectID).WithField("amp_user_id", ampUserID)

	if projectID == 0 || ampUserID == "" {
		logCtx.Error("Invalid params on getCacheUserIDByAMPUserID.")
		return "", http.StatusInternalServerError
	}

	key, err := getCacheKeyForUserIDByAMPUserID(projectID, ampUserID)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get cache key for user_id by amp_user_id.")
		return "", http.StatusInternalServerError
	}

	userID, err := cacheRedis.Get(key)
	if err != nil {
		if err == redis.ErrNil {
			return "", http.StatusNotFound
		}

		logCtx.WithError(err).Error("Failed to user_id by amp_user_id from cache.")
		return "", http.StatusInternalServerError
	}

	return userID, http.StatusFound
}

func SetCacheUserIDByAMPUserID(projectID int64, ampUserID, userID string) int {
	logCtx := log.WithField("project_id", projectID).WithField("amp_user_id", ampUserID)

	if projectID == 0 || ampUserID == "" || userID == "" {
		logCtx.Error("Invalid params on setCacheUserIDByAMPUserID.")
		return http.StatusInternalServerError
	}

	key, err := getCacheKeyForUserIDByAMPUserID(projectID, ampUserID)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get cache key for setCacheUserIDByAMPUserID.")
		return http.StatusInternalServerError
	}

	var expiryInSecs float64 = 60 * 15 // 15 minutes.
	err = cacheRedis.Set(key, userID, expiryInSecs)
	if err != nil {
		logCtx.WithError(err).Error("Failed to set cache on setCacheUserIDByAMPUserID")
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

func MergeUserPropertiesByCustomerUserID(projectID int64, users []User, customerUserID string, source string, objectType string) (*map[string]interface{}, int) {
	logCtx := log.WithField("project_id", projectID).
		WithField("users", users).
		WithField("customer_user_id", customerUserID)

	usersLength := len(users)
	if usersLength == 0 {
		logCtx.Error("No users for merging the user_properties.")
		return nil, http.StatusInternalServerError
	}

	initialPropertiesVisitedMap := make(map[string]bool)
	for _, property := range U.USER_PROPERTIES_MERGE_TYPE_INITIAL {
		initialPropertiesVisitedMap[property] = false
	}

	// Order the properties by jointime to maintain the initial properties
	sort.Slice(users, func(i, j int) bool {
		return users[i].JoinTimestamp < users[j].JoinTimestamp
	})

	initialProperties := make(map[string]interface{})
	for i := range users {
		user := users[i]
		userProperties, err := U.DecodePostgresJsonb(&user.Properties)
		if err != nil {
			logCtx.WithField("user_properties", user.Properties).
				Error("Failed to decode user properties on initial properties.")
			return nil, http.StatusInternalServerError
		}

		for property := range *userProperties {
			isAlreadySet, isInitialProperty := initialPropertiesVisitedMap[property]
			if isInitialProperty {
				if !isAlreadySet {
					// For initial properties, set only once for earliest user.
					initialProperties[property] = (*userProperties)[property]
					initialPropertiesVisitedMap[property] = true
				}
			}
		}
	}

	// Order the properties before merging the properties to
	// ensure the precendence of value.
	sort.Slice(users, func(i, j int) bool {
		return users[i].PropertiesUpdatedTimestamp < users[j].PropertiesUpdatedTimestamp
	})

	mergedUserProperties := make(map[string]interface{})
	for property := range initialProperties {
		mergedUserProperties[property] = initialProperties[property]
	}

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

		useSourcePropertyOverwrite := config.UseSourcePropertyOverwriteByProjectIDs(projectID)
		if user.PropertiesUpdatedTimestamp > mergedUpdatedTimestamp {
			if useSourcePropertyOverwrite {
				if source != SmartCRMEventSourceHubspot && source != SmartCRMEventSourceSalesforce {
					mergedUpdatedTimestamp = user.PropertiesUpdatedTimestamp
				}
			} else {
				mergedUpdatedTimestamp = user.PropertiesUpdatedTimestamp
			}
		}

		overwriteProperties := false
		var overwritePropertiesError bool
		if useSourcePropertyOverwrite {
			overwriteProperties, err = CheckForCRMUserPropertiesOverwrite(source, objectType, *userProperties, mergedUserProperties)
			if err != nil {
				logCtx.WithField("error", err.Error()).Error("Failed to get overwriteProperties flag value.")
				overwritePropertiesError = true
			}
		}

		for property := range *userProperties {
			mergedUserPropertiesValues[property] = append(mergedUserPropertiesValues[property], (*userProperties)[property])
			if U.StringValueIn(property, U.USER_PROPERTIES_MERGE_TYPE_ADD[:]) ||
				(IsEmptyPropertyValue((*userProperties)[property]) && !U.IsCRMPropertyKey(property)) {
				continue
			}

			_, isInitialProperty := initialPropertiesVisitedMap[property]
			if !isInitialProperty {
				if useSourcePropertyOverwrite && !overwritePropertiesError {
					if U.IsCRMPropertyKeyBySource(source, property) {
						if overwriteProperties {
							mergedUserProperties[property] = (*userProperties)[property]
						} else {
							if _, exist := mergedUserProperties[property]; !exist {
								mergedUserProperties[property] = (*userProperties)[property]
							}
						}
						continue
					}
				}

				mergedUserProperties[property] = (*userProperties)[property]
			}
		}
	}

	// Handle merge for add type properties separately.
	userPropertiesToBeMerged := make([]postgres.Jsonb, 0, 0)
	for i := range users {
		userPropertiesToBeMerged = append(userPropertiesToBeMerged, users[i].Properties)
	}
	MergeAddTypeUserProperties(&mergedUserProperties, userPropertiesToBeMerged)

	// Additional check for properties that can be added. If merge is triggered for users with same set of properties,
	// value of properties that can be added will change after addition. Below check is to avoid update in such case.
	if !AnyPropertyChanged(mergedUserPropertiesValues, len(users)) {
		return &mergedUserProperties, http.StatusOK
	}
	mergedUserProperties[U.UP_MERGE_TIMESTAMP] = U.TimeNowUnix()

	// removing U.UP_SESSION_COUNT, from user properties.
	delete(mergedUserProperties, U.UP_SESSION_COUNT)

	return &mergedUserProperties, http.StatusOK
}

func getCRMTimestampValue(value interface{}) (int64, error) {
	fValue, err := util.GetPropertyValueAsFloat64(value)
	if err != nil {
		timestamp, err1 := GetTimestampForV3Records(value)
		if err1 == nil {
			return timestamp, nil
		}

		timestamp, err2 := GetSalesforceDocumentTimestamp(value) // make sure timezone info is loaded to the container
		if err2 == nil {
			return timestamp, nil
		}

		return 0, fmt.Errorf("%v %v", err1.Error(), err2.Error())
	}

	timestamp := int64(fValue)
	if timestamp >= 10000000000 { // hubspot old millisecond timestamp
		return timestamp / 1000, nil
	}

	return timestamp, nil
}

func CheckForCRMUserPropertiesOverwrite(source string, objectType string, incomingProperties map[string]interface{},
	currentProperties map[string]interface{}) (bool, error) {
	logCtx := log.WithField("source", source).
		WithField("objectType", objectType)

	if !IsCRMSource(source) {
		return false, nil
	}

	overwriteProperties := false
	if objectType == "" {
		return overwriteProperties, nil
	}

	propertySuffix := GetSourceUserPropertyOverwritePropertySuffix(source, objectType)
	lastmodifieddateProperty := GetCRMEnrichPropertyKeyByType(source, objectType, propertySuffix)
	incomingPropertyValue, err := getCRMTimestampValue(incomingProperties[lastmodifieddateProperty])
	if err != nil {
		logCtx.WithField("mergedUserProperties", incomingProperties).WithError(err).
			Error("Failed to convert incoming lastmodifieddate property value to float64 inside CheckForCRMUserPropertiesOverwrite.")
		return overwriteProperties, err
	}
	currentPropertyValue, err := getCRMTimestampValue(currentProperties[lastmodifieddateProperty])
	if err != nil {
		logCtx.WithField("userProperties", currentProperties).WithError(err).
			Error("Failed to convert current lastmodifieddate property value to float64 inside CheckForCRMUserPropertiesOverwrite.")
		return overwriteProperties, err
	}
	if currentPropertyValue < incomingPropertyValue {
		overwriteProperties = true
	}
	return overwriteProperties, nil
}

// List of constants to be upadated for each CRM
var (
	// List of property key which tracks update in the object. Property key without prefixes
	SourceUserPropertiesOverwritePropertyKeys = map[string]map[string]string{
		U.CRM_SOURCE_NAME_HUBSPOT: {
			HubspotDocumentTypeNameContact: util.PROPERTY_KEY_LAST_MODIFIED_DATE,
			HubspotDocumentTypeNameCompany: util.PROPERTY_KEY_LAST_MODIFIED_DATE_HS,
			HubspotDocumentTypeNameDeal:    util.PROPERTY_KEY_LAST_MODIFIED_DATE_HS,
		},
		U.CRM_SOURCE_NAME_SALESFORCE: {
			SalesforceDocumentTypeNameLead:        util.PROPERTY_KEY_LAST_MODIFIED_DATE,
			SalesforceDocumentTypeNameContact:     util.PROPERTY_KEY_LAST_MODIFIED_DATE,
			SalesforceDocumentTypeNameAccount:     util.PROPERTY_KEY_LAST_MODIFIED_DATE,
			SalesforceDocumentTypeNameOpportunity: util.PROPERTY_KEY_LAST_MODIFIED_DATE,
		},
		U.CRM_SOURCE_NAME_MARKETO: {
			"lead": "updated_at",
		},
		UserSourceLeadSquared: {
			"lead": "ModifiedOn",
		},
	}

	// Block updating user properties_update_timestamp. CRM uses object property key to update.
	BlacklistUserPropertiesUpdateTimestampBySource = map[string]bool{
		SmartCRMEventSourceHubspot:    true,
		SmartCRMEventSourceSalesforce: true,
		U.CRM_SOURCE_NAME_MARKETO:     true,
		UserSourceLeadSquared:         true,
	}

	/*
		CRM_SOURCE and ALLOWED_CRM_SOURCES constants should be updated for CRM source check
	*/
)

func GetSourceUserPropertyOverwritePropertySuffix(source string, objectType string) string {

	sourceKeys, ok := SourceUserPropertiesOverwritePropertyKeys[source]
	if !ok {
		return ""
	}

	return sourceKeys[objectType]
}

// SetUserGroupFieldByColumnName update user struct field by gorm column name. If value already set then it won't update the value
func SetUserGroupFieldByColumnName(user *User, columnName, value string, overwrite bool) (bool, bool, error) {

	if user == nil || columnName == "" || value == "" {
		return false, false, errors.New("invalid parameters")
	}

	if !strings.HasPrefix(columnName, "group_") {
		return false, false, errors.New("not a group field")
	}

	refUserVal := reflect.ValueOf(user)
	refUserTyp := refUserVal.Elem().Type()

	processed := false
	updated := false
	for i := 0; i < refUserVal.Elem().NumField(); i++ {
		refField := refUserTyp.Field(i)
		if tagName := refField.Tag.Get("gorm"); tagName != "" {

			// refer parseTagSetting in backend/src/factors/vendor/github.com/jinzhu/gorm/model_struct.go
			tags := strings.Split(tagName, ";")
			tagColumnName := ""
			for _, value := range tags {
				v := strings.Split(value, ":")
				if len(v) == 2 && v[0] == "column" { // `gorm:"default:null;column:group_1_user_id" json:"group_1_user_id"`
					tagColumnName = v[1]
				}
			}

			field := refUserVal.Elem().Field(i)
			currValue := field.String()
			if tagColumnName == columnName {
				if processed {
					return false, false, errors.New("duplicate tag found")
				}

				if currValue == "" || (overwrite && currValue != value) { // don't overwrite if already set

					if !field.CanSet() {
						return false, false, errors.New("cannot update field")
					}

					refUserVal.Elem().Field(i).SetString(value)
					updated = true
				}
				processed = true
			}
		}
	}

	if !processed {
		return false, false, errors.New("Failed to update user by tag")
	}

	return processed, updated, nil
}

func GetGroupUserGroupID(groupUser *User, groupIndex int) (string, error) {
	if groupUser.IsGroupUser == nil || !(*groupUser.IsGroupUser) {
		return "", errors.New("not a group user")
	}
	groupID, err := GetUserGroupID(groupUser, groupIndex)
	if err != nil {
		return "", err
	}
	if groupID == "" {
		return "", errors.New("failed to get group id for group user")
	}
	return groupID, nil
}

var ErrMissingDomainUserID = errors.New("failed to get domains user id for group user")

// GetGroupUserDomainsUserID gives domains user id from the group user
func GetGroupUserDomainsUserID(groupUser *User, groupIndex int) (string, error) {
	if groupUser.IsGroupUser == nil || !(*groupUser.IsGroupUser) {
		return "", errors.New("not a group user")
	}

	groupUserID, err := GetUserGroupUserID(groupUser, groupIndex)
	if err != nil {
		return "", err
	}

	if groupUserID == "" {
		return "", ErrMissingDomainUserID
	}

	return groupUserID, nil
}

func GetUserGroupUserID(user *User, groupIndex int) (string, error) {
	groupUserID := ""
	if groupIndex != 0 {
		groupUserID = fmt.Sprintf("group_%d_user_id", groupIndex)
	}
	return GetUserGroupColumn(user, groupUserID)
}

func GetUserGroupID(user *User, groupIndex int) (string, error) {
	groupID := ""
	if groupIndex != 0 {
		groupID = fmt.Sprintf("group_%d_id", groupIndex)
	}
	return GetUserGroupColumn(user, groupID)
}

// GetUserGroupColumn return group_<index>_id or group_<index>_user_id by group index
func GetUserGroupColumn(user *User, searchColumn string) (string, error) {

	refUserVal := reflect.ValueOf(user)
	refUserTyp := refUserVal.Elem().Type()

	value := ""
	for i := 0; i < refUserVal.Elem().NumField(); i++ {
		refField := refUserTyp.Field(i)
		if tagName := refField.Tag.Get("json"); strings.HasPrefix(tagName, "group_") {
			if searchColumn != "" && searchColumn != tagName {
				continue
			}

			field := refUserVal.Elem().Field(i)
			if field.Kind() != reflect.String {
				continue
			}

			fieldValue := field.String()
			if fieldValue != "" { // group user won't have multiple id associated and group user id are empty
				if value != "" {
					return "", errors.New("more than 1 field value found")
				}
				value = fieldValue
			}

		}

	}

	return value, nil
}

func GetUserGroupUserIDsByGroupIDs(projectID int64, users []User, requiredGroupsIndex []int) (map[int]string, error) {
	groupUserIDByGroupID := make(map[int]string)
	sort.Slice(users, func(i, j int) bool {
		return users[i].JoinTimestamp < users[j].JoinTimestamp
	})

	for i := range users {
		user := users[i]
		for _, groupIndex := range requiredGroupsIndex {
			groupUserID, err := GetUserGroupUserID(&user, groupIndex)
			if err != nil {
				log.WithFields(log.Fields{"project_id": projectID, "user_id": users[i]}).WithError(err).
					Error("Failed to get group user id on GetUserGroupUserIDsByGroupIDs.")
				return nil, err
			}

			if groupUserID != "" {
				groupUserIDByGroupID[groupIndex] = groupUserID
			}
		}
	}

	return groupUserIDByGroupID, nil
}

func IsValidUserSource(source string) bool {
	if _, exists := UserSourceMap[source]; exists {
		return true
	} else {
		return false
	}
}

func GetUserSourceByName(source string) int {
	return UserSourceMap[source]
}

func GetUserSourceName(source int) string {
	for name := range UserSourceMap {
		if UserSourceMap[name] == source {
			return name
		}
	}

	return ""
}

func GetGroupUserSourceByGroupName(groupName string) int {
	return GroupUserSource[groupName]
}

func GetGroupUserSourceNameByGroupName(groupName string) string {
	return GetUserSourceName(GroupUserSource[groupName])
}

func GetUsersForDomainUserAssociationUpdate(users []User) []User {
	if len(users) <= 100 {
		return users
	}

	// Sort by updated_at ASC
	sort.Slice(users, func(i, j int) bool {
		return users[i].UpdatedAt.Before(users[j].UpdatedAt)
	})

	var updateUsers []User
	updateUsers = users[:50]
	updateUsers = append(updateUsers, users[len(users)-50:]...)

	return updateUsers
}
