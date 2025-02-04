package model

import (
	"encoding/json"
	"errors"
	"factors/cache"
	pCache "factors/cache/persistent"
	U "factors/util"
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// Group is an interface for groups table
type Group struct {
	ProjectID int64     `gorm:"primary_key:true;" json:"project_id"`
	ID        int       `gorm:"not null" json:"id"`
	Name      string    `gorm:"primary_key:true;" json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type GroupName struct {
	Name        string `json:"group_name"`
	DisplayName string `json:"display_name"`
	IsAccount   bool   `json:"is_account"`
}

const GROUP_NAME_HUBSPOT_COMPANY = "$hubspot_company"
const GROUP_NAME_HUBSPOT_DEAL = "$hubspot_deal"
const GROUP_NAME_SALESFORCE_ACCOUNT = "$salesforce_account"
const GROUP_NAME_SALESFORCE_OPPORTUNITY = "$salesforce_opportunity"
const GROUP_NAME_SIX_SIGNAL = "$6signal"
const GROUP_NAME_DOMAINS = "$domains"
const GROUP_NAME_LINKEDIN_COMPANY = "$linkedin_company"
const GROUP_NAME_G2 = "$g2"

// AllowedGroupNames list of allowed group names
var AllowedGroupNames = map[string]bool{
	GROUP_NAME_HUBSPOT_COMPANY:        true,
	GROUP_NAME_HUBSPOT_DEAL:           true,
	GROUP_NAME_SALESFORCE_ACCOUNT:     true,
	GROUP_NAME_SALESFORCE_OPPORTUNITY: true,
	GROUP_NAME_SIX_SIGNAL:             true,
	GROUP_NAME_LINKEDIN_COMPANY:       true,
	GROUP_NAME_G2:                     true,
}
var AccountGroupNames = map[string]bool{
	GROUP_NAME_HUBSPOT_COMPANY:    true,
	GROUP_NAME_SALESFORCE_ACCOUNT: true,
	GROUP_NAME_SIX_SIGNAL:         true,
	GROUP_NAME_LINKEDIN_COMPANY:   true,
	GROUP_NAME_G2:                 true,
}

var AllowedGroupToDomainsGroup = map[string]bool{
	GROUP_NAME_HUBSPOT_COMPANY:    true,
	GROUP_NAME_SALESFORCE_ACCOUNT: true,
	GROUP_NAME_SIX_SIGNAL:         true,
	GROUP_NAME_LINKEDIN_COMPANY:   true,
	GROUP_NAME_G2:                 true,
}

var DomainNameSourcePropertyKey = map[string][]string{
	GROUP_NAME_SIX_SIGNAL:         {U.SIX_SIGNAL_DOMAIN},
	GROUP_NAME_HUBSPOT_COMPANY:    {"$hubspot_company_domain", "$hubspot_company_website"},
	GROUP_NAME_SALESFORCE_ACCOUNT: {"$salesforce_account_website"},
}

func GetDomainNameSourcePropertyKey(groupName string) []string {
	return DomainNameSourcePropertyKey[groupName]
}

// AllowedGroups total groups allowed per project
var AllowedGroups = 8

func GetPropertiesByGroupCategoryCacheKeySortedSet(projectId int64, date string) (*cache.Key, error) {
	prefix := "SS:GN:PC"
	return cache.NewKey(projectId, prefix, date)
}

func GetValuesByGroupPropertyCacheKeySortedSet(projectId int64, date string) (*cache.Key, error) {
	prefix := "SS:GN:PV"
	return cache.NewKey(projectId, prefix, date)
}

func GetPropertiesByGroupCategoryRollUpCacheKey(projectId int64, groupName string, date string) (*cache.Key, error) {
	prefix := "RollUp:GN:PC"
	return cache.NewKey(projectId, fmt.Sprintf("%s:%s", prefix, groupName), date)
}

func GetValuesByGroupPropertyRollUpCacheKey(projectId int64, groupName string, propertyName string, date string) (*cache.Key, error) {
	prefix := "RollUp:GN:PV"
	return cache.NewKey(projectId, fmt.Sprintf("%s:%s:%s", prefix, groupName, propertyName), date)
}

func GetPropertiesByGroupFromCache(projectID int64, groupName string, dateKey string) (U.CachePropertyWithTimestamp, error) {
	logFields := log.Fields{
		"project_id": projectID,
		"group_name": groupName,
		"date_key":   dateKey,
	}
	defer LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	if projectID == 0 {
		return U.CachePropertyWithTimestamp{}, errors.New("invalid project on GetPropertiesByGroupFromCache")
	}

	if groupName == "" || groupName == "undefined" {
		return U.CachePropertyWithTimestamp{}, errors.New("invalid group_name on GetPropertiesByGroupFromCache")
	}

	groupPropertiesKey, err := GetPropertiesByGroupCategoryRollUpCacheKey(projectID, groupName, dateKey)
	if err != nil {
		return U.CachePropertyWithTimestamp{}, err
	}
	groupProperties, _, err := pCache.GetIfExists(groupPropertiesKey, true)
	if err != nil || groupProperties == "" {
		logCtx.WithField("date_key", dateKey).Info("Missing rollup cache for groups.")
		return U.CachePropertyWithTimestamp{}, nil
	}

	var cacheValue U.CachePropertyWithTimestamp
	err = json.Unmarshal([]byte(groupProperties), &cacheValue)
	if err != nil {
		return U.CachePropertyWithTimestamp{}, err
	}
	return cacheValue, nil
}

func GetPropertyValuesByGroupPropertyFromCache(projectID int64, groupName string, propertyName string, dateKey string) (U.CachePropertyValueWithTimestamp, error) {
	logFields := log.Fields{
		"project_id":    projectID,
		"group_name":    groupName,
		"property_name": propertyName,
		"date_key":      dateKey,
	}
	defer LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	if projectID == 0 {
		return U.CachePropertyValueWithTimestamp{}, errors.New("invalid project on GetPropertyValuesByGroupFromCache")
	}

	if groupName == "" {
		return U.CachePropertyValueWithTimestamp{}, errors.New("invalid event_name on GetPropertyValuesByGroupFromCache")
	}

	if propertyName == "" {
		return U.CachePropertyValueWithTimestamp{}, errors.New("invalid property_name on GetPropertyValuesByEventPropertyFromCache")
	}

	groupPropertyValuesKey, err := GetValuesByGroupPropertyRollUpCacheKey(projectID, groupName, propertyName, dateKey)
	if err != nil {
		return U.CachePropertyValueWithTimestamp{}, err
	}
	values, exists, _ := pCache.GetIfExists(groupPropertyValuesKey, true)
	if !exists || values == "" {
		return U.CachePropertyValueWithTimestamp{}, nil
	}

	var cacheValue U.CachePropertyValueWithTimestamp
	err = json.Unmarshal([]byte(values), &cacheValue)
	if err != nil {
		logCtx.WithError(err).Error("Failed to unmarshal property value from cache.")
		return U.CachePropertyValueWithTimestamp{}, err
	}
	return cacheValue, nil
}

func IsAllowedGroupName(name string) bool {
	return AllowedGroupNames[name]
}

func IsAllowedGroupForDomainsGroup(name string) bool {
	return AllowedGroupToDomainsGroup[name]
}

func IsAllowedAccountGroupNames(name string) bool {
	return AccountGroupNames[name]
}

func IsSixSignalProperty(userProperty string) bool {
	return strings.HasPrefix(userProperty, "$6Signal_")
}

func GetSixSignalDefaultUserProperties() map[string][]string {
	properties := make(map[string][]string)

	for key := range U.STANDARD_USER_PROPERTIES_DISPLAY_NAMES {
		if IsSixSignalProperty(key) {
			dataType := U.GetPropertyTypeByName(key)
			if properties[dataType] == nil {
				properties[dataType] = make([]string, 0)
			}
			properties[dataType] = append(properties[dataType], key)
		}
	}

	return properties
}

func GetSixSignalDefaultUserPropertiesDisplayNames() map[string]string {
	displayNames := make(map[string]string)
	for property, displayName := range U.STANDARD_USER_PROPERTIES_DISPLAY_NAMES {
		if IsSixSignalProperty(property) {
			displayNames[property] = displayName
		}
	}
	return displayNames
}
