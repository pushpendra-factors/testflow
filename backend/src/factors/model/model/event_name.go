package model

import (
	"encoding/json"
	"errors"
	cacheRedis "factors/cache/redis"
	U "factors/util"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type EventName struct {
	// Composite primary key with projectId.
	ID   string `gorm:"primary_key:true;" json:"id"`
	Name string `json:"name"`
	Type string `gorm:"not null;type:varchar(2)" json:"type"`
	// Below are the foreign key constraints added in creation script.
	// project_id -> projects(id)
	ProjectId int64 `gorm:"primary_key:true;" json:"project_id"`
	// if default is not set as NULL empty string will be installed.
	FilterExpr string    `gorm:"type:varchar(500);default:null" json:"filter_expr"`
	Deleted    bool      `gorm:"not null;default:false" json:"deleted"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type CacheEventNames struct {
	EventNames []EventName
	Timestamp  int64
}

type CacheEventNamesWithTimestamp struct {
	EventNames map[string]U.CountTimestampTuple `json:"en"`
}

// AT is for all page event
// FE and SE are same
// UC is for form submit event
const TYPE_USER_CREATED_EVENT_NAME = "UC"
const TYPE_AUTO_TRACKED_EVENT_NAME = "AT"
const TYPE_FILTER_EVENT_NAME = "FE"
const TYPE_INTERNAL_EVENT_NAME = "IE"
const TYPE_CRM_SALESFORCE = "CS"
const TYPE_CRM_HUBSPOT = "CH"
const EVENT_NAME_REQUEST_TYPE_APPROX = "approx"
const EVENT_NAME_REQUEST_TYPE_EXACT = "exact"
const EVENT_NAME_TYPE_SMART_EVENT = "SE"

var ALLOWED_TYPES = [...]string{
	TYPE_USER_CREATED_EVENT_NAME,
	TYPE_AUTO_TRACKED_EVENT_NAME,
	TYPE_FILTER_EVENT_NAME,
	TYPE_INTERNAL_EVENT_NAME,
	TYPE_CRM_SALESFORCE,
	TYPE_CRM_HUBSPOT,
}

var AllowedEventNamesForHubspot = []string{
	U.EVENT_NAME_HUBSPOT_CONTACT_CREATED,
	U.EVENT_NAME_HUBSPOT_CONTACT_UPDATED,
	U.EVENT_NAME_HUBSPOT_DEAL_STATE_CHANGED,
	U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_CREATED,
	U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_UPDATED,
	U.GROUP_EVENT_NAME_HUBSPOT_DEAL_CREATED,
	U.GROUP_EVENT_NAME_HUBSPOT_DEAL_UPDATED,
	U.EVENT_NAME_HUBSPOT_CONTACT_FORM_SUBMISSION,
	U.EVENT_NAME_HUBSPOT_ENGAGEMENT_MEETING_CREATED,
	U.EVENT_NAME_HUBSPOT_ENGAGEMENT_MEETING_UPDATED,
	U.EVENT_NAME_HUBSPOT_ENGAGEMENT_CALL_CREATED,
	U.EVENT_NAME_HUBSPOT_ENGAGEMENT_CALL_UPDATED,
	U.EVENT_NAME_HUBSPOT_ENGAGEMENT_EMAIL,
}

// NOTE: This is currently being used only in kpi though.
var AllowedEventNamesForSalesforce = []string{
	U.EVENT_NAME_SALESFORCE_CONTACT_CREATED,
	U.EVENT_NAME_SALESFORCE_CONTACT_UPDATED,
	U.EVENT_NAME_SALESFORCE_LEAD_CREATED,
	U.EVENT_NAME_SALESFORCE_LEAD_UPDATED,
	U.EVENT_NAME_SALESFORCE_OPPORTUNITY_CREATED,
	U.EVENT_NAME_SALESFORCE_OPPORTUNITY_UPDATED,
}

var EventTypeToEnameType = map[string]string{
	PageViewsDisplayCategory: "AT",
}

const URI_PROPERTY_PREFIX = ":"
const EVENT_NAMES_LIMIT = 5000

// TimestampReferenceTypeDocument use document timestamp for smart event creation
const TimestampReferenceTypeDocument = "timestamp_in_track"

// SmartCRMEventFilter struct is base for CRM smart event filter
type SmartCRMEventFilter struct {
	Source                  string           `json:"source" enums:"salesforce,hubspot"`
	ObjectType              string           `json:"object_type" enums:"salesforce[account,contact,lead],hubspot[contact]"`
	Description             string           `json:"description"`
	FilterEvaluationType    string           `json:"property_evaluation_type" enums:"specific,any"` //any change or specific
	Filters                 []PropertyFilter `json:"filters"`
	TimestampReferenceField string           `json:"timestamp_reference_field" enums:"timestamp_in_track, <any property name>"`
	LogicalOp               string           `json:"logical_op" enums:"AND"`
}

// CRMFilterRule struct for filter rule
type CRMFilterRule struct {
	Operator      string        `json:"op" enums:"EQUAL,NOT EQUAL,GREATER THAN,LESS THAN,CONTAINS,NOT CONTAINS"`
	PropertyState PropertyState `json:"gen" enums:"curr,last"` // previous or current
	Value         interface{}   `json:"value"`                 // value_any or property value
}

// PropertyFilter struct hold name of the property and logical operations on rules provided
type PropertyFilter struct {
	Name      string          `json:"property_name"`
	Rules     []CRMFilterRule `josn:"rules"`
	LogicalOp string          `json:"logical_op" enums:"AND"`
}

// PropertyState holds string representing state of the property
type PropertyState string

// PropertyState represents the current or prevous state of property
const (
	CurrentState  PropertyState = "curr"
	PreviousState PropertyState = "last"
)

type EventNameWithAggregation struct {
	// Composite primary key with projectId.
	ID   string `gorm:"primary_key:true;" json:"id"`
	Name string `json:"name"`
	Type string `gorm:"not null;type:varchar(2)" json:"type"`
	// Below are the foreign key constraints added in creation script.
	// project_id -> projects(id)
	ProjectId int64 `gorm:"primary_key:true;" json:"project_id"`
	// if default is not set as NULL empty string will be installed.
	FilterExpr string    `gorm:"type:varchar(500);default:null" json:"filter_expr"`
	Deleted    bool      `gorm:"not null;default:false" json:"deleted"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	LastSeen   uint64    `json:"last_seen"`
	Count      int64     `json:"count"`
}

// Support source for CRM smart event filter
const (
	SmartCRMEventSourceSalesforce = "salesforce"
	SmartCRMEventSourceHubspot    = "hubspot"

	// smart event property prefix
	SmartCRMEventPreviousPropertyPrefix = U.NAME_PREFIX + "prev_"
	SmartCRMEventCurrentPropertyPrefix  = U.NAME_PREFIX + "curr_"

	SmartCRMEventSalesforcePrevPropertyPrefix = SmartCRMEventPreviousPropertyPrefix + SmartCRMEventSourceSalesforce + "_"
	SmartCRMEventHubspotPrevPropertyPrefix    = SmartCRMEventPreviousPropertyPrefix + SmartCRMEventSourceHubspot + "_"

	SmartCRMEventSalesforceCurrPropertyPrefix = SmartCRMEventCurrentPropertyPrefix + SmartCRMEventSourceSalesforce + "_"
	SmartCRMEventHubspotCurrPropertyPrefix    = SmartCRMEventCurrentPropertyPrefix + SmartCRMEventSourceHubspot + "_"
)

var ErrorSmartEventFiterEmptyString = errors.New("empty string")

// GetDecodedSmartEventFilterExp unmarhsal encoded CRM smart event filter exp to SmartCRMEventFilter struct
func GetDecodedSmartEventFilterExp(enFilterExp string) (*SmartCRMEventFilter, error) {
	if enFilterExp == "" {
		return nil, ErrorSmartEventFiterEmptyString
	}

	var smartCRMEventFilter SmartCRMEventFilter
	err := json.Unmarshal([]byte(enFilterExp), &smartCRMEventFilter)
	if err != nil {
		return nil, err
	}

	return &smartCRMEventFilter, nil
}

func GetCurrPropertyName(pName, source, objectType string) string {
	return getCurrPropertyName(pName, source, objectType)
}

func GetPrevPropertyName(pName, source, objectType string) string {
	return getPrevPropertyName(pName, source, objectType)
}

func getPrevPropertyName(pName, source, objectType string) string {
	if pName == "" || source == "" || objectType == "" {
		return ""
	}

	return SmartCRMEventPreviousPropertyPrefix + getCRMPropertyKeyByType(source, objectType, pName)
}

func getCurrPropertyName(pName, source, objectType string) string {
	if pName == "" || source == "" || objectType == "" {
		return ""
	}
	return SmartCRMEventCurrentPropertyPrefix + getCRMPropertyKeyByType(source, objectType, pName)
}

// GetPropertyNameByTrimmedSmartEventPropertyPrefix removes smart event property property prefix
func GetPropertyNameByTrimmedSmartEventPropertyPrefix(pName string) string {
	if strings.HasPrefix(pName, SmartCRMEventSalesforcePrevPropertyPrefix) ||
		strings.HasPrefix(pName, SmartCRMEventHubspotPrevPropertyPrefix) {
		return U.NAME_PREFIX + strings.TrimPrefix(pName, SmartCRMEventPreviousPropertyPrefix)
	}

	if strings.HasPrefix(pName, SmartCRMEventSalesforceCurrPropertyPrefix) ||
		strings.HasPrefix(pName, SmartCRMEventHubspotCurrPropertyPrefix) {
		return U.NAME_PREFIX + strings.TrimPrefix(pName, SmartCRMEventCurrentPropertyPrefix)
	}

	return pName
}

// FillSmartEventCRMProperties fills all properties from CRM smart filter to new properties
func FillSmartEventCRMProperties(newProperties, current, prev *map[string]interface{},
	filter *SmartCRMEventFilter) {

	if *newProperties == nil {
		*newProperties = make(map[string]interface{})
	}

	for i := range filter.Filters {
		if value, exists := (*current)[filter.Filters[i].Name]; exists {
			(*newProperties)[getCurrPropertyName(filter.Filters[i].Name, filter.Source, filter.ObjectType)] = value
		}
		if value, exists := (*prev)[filter.Filters[i].Name]; exists {
			(*newProperties)[getPrevPropertyName(filter.Filters[i].Name, filter.Source, filter.ObjectType)] = value
		}
	}
}

// CRMSmartEvent holds payload for creating smart event
type CRMSmartEvent struct {
	Name       string
	Properties map[string]interface{}
	Timestamp  uint64
}

// compare modes for validating properties
const (
	CompareStateCurr = "curr"
	CompareStatePrev = "prev"
	CompareStateBoth = "both"
)

// FilterEvaluationTypeSpecific for specific change in property or any change property
const (
	FilterEvaluationTypeSpecific = "specific"
	FilterEvaluationTypeAny      = "any"
)

// list of logical operators for CRM filter
const (
	LOGICAL_OP_OR  = "OR"
	LOGICAL_OP_AND = "AND"
)

func isSameSourceAndObjectType(existingFilter *SmartCRMEventFilter, incomingFilter *SmartCRMEventFilter) bool {
	if existingFilter.Source == incomingFilter.Source &&
		existingFilter.ObjectType == incomingFilter.ObjectType &&
		existingFilter.FilterEvaluationType == incomingFilter.FilterEvaluationType {
		return true
	}

	return false
}

// IsEventNameTypeSmartEvent validates event name is of type smart event
func IsEventNameTypeSmartEvent(eventType string) bool {
	return eventType == TYPE_CRM_HUBSPOT || eventType == TYPE_CRM_SALESFORCE
}

func isDuplicateTimestampReferenceField(existingFilter, incomingFilter *SmartCRMEventFilter) bool {
	return existingFilter.TimestampReferenceField == incomingFilter.TimestampReferenceField
}

func isDuplicatePropertyFilters(existingFilter, incomingFilter []PropertyFilter) bool {

	if len(existingFilter) != len(incomingFilter) {
		return false
	}

	existingRuleMap := make(map[string]bool)
	for i := range existingFilter {

		if len(existingFilter[i].Rules) < 1 { // FilterEvaluationType == any, doesn't require any specific rule
			key := existingFilter[i].Name
			existingRuleMap[key] = true
			continue
		}

		for _, rule := range existingFilter[i].Rules {
			key := fmt.Sprintf("%s:%s:%s:%s", existingFilter[i].Name, rule.PropertyState, rule.Operator, rule.Value)
			existingRuleMap[key] = true
		}
	}

	incomingRulesLen := 0
	for i := range incomingFilter {

		if len(incomingFilter[i].Rules) < 1 { // FilterEvaluationType == any, doesn't require any specific rule
			key := existingFilter[i].Name
			if _, exist := existingRuleMap[key]; !exist {
				return false
			}

			continue
		}

		for _, rule := range incomingFilter[i].Rules {
			key := fmt.Sprintf("%s:%s:%s:%s", incomingFilter[i].Name, rule.PropertyState, rule.Operator, rule.Value)
			if _, exist := existingRuleMap[key]; !exist {
				return false
			}
			incomingRulesLen++
		}
	}

	if incomingRulesLen != len(existingRuleMap) {
		return false
	}

	return true
}

// CheckSmartEventNameDuplicateFilter validates two smart event filter for duplicates.
func CheckSmartEventNameDuplicateFilter(existingFilter *SmartCRMEventFilter, incomingFilter *SmartCRMEventFilter) bool {
	if isSameSourceAndObjectType(existingFilter, incomingFilter) {
		if isDuplicatePropertyFilters(existingFilter.Filters, incomingFilter.Filters) {
			if isDuplicateTimestampReferenceField(existingFilter, incomingFilter) {
				return true
			}
		}
	}

	return false
}

// CRMFilterEvaluator evaluates a CRM filter on the properties provided. Can work in current properties or current and previous property mode
func CRMFilterEvaluator(projectID int64, currProperty, prevProperty *map[string]interface{},
	filter *SmartCRMEventFilter, compareState string) bool {
	if filter == nil {
		return false
	}

	if compareState == "" ||
		(compareState == CompareStateCurr && currProperty == nil) ||
		(compareState == CompareStatePrev && prevProperty == nil) ||
		(compareState == CompareStateBoth && (currProperty == nil || prevProperty == nil)) {
		return false
	}

	filterSkipable := filter.LogicalOp == LOGICAL_OP_OR

	anyfilterTrue := false
	for _, filterProperty := range filter.Filters { // a successfull completion of this loop implies a vaild AND or failed OR operation
		ruleSkipable := filterProperty.LogicalOp == LOGICAL_OP_OR
		var anyPrevMatch bool
		var anyCurrMatch bool

		// avoid same value in previous and current properties
		if compareState == CompareStateBoth {
			diffPropertyValue := U.GetPropertyValueAsString((*currProperty)[filterProperty.Name]) != U.GetPropertyValueAsString((*prevProperty)[filterProperty.Name])
			if !diffPropertyValue {
				if !filterSkipable {
					return false
				}
				continue
			}

			if filter.FilterEvaluationType == FilterEvaluationTypeAny {
				if diffPropertyValue {
					anyfilterTrue = true
				} else {
					if !filterSkipable {
						return false
					}
				}
				continue
			}
		}

		// cannot compare if only one is provided, return true and switch to both mode
		if (compareState == CompareStateCurr || compareState == CompareStatePrev) && filter.FilterEvaluationType == FilterEvaluationTypeAny {
			return true
		}

		for _, rule := range filterProperty.Rules { // a successfull completion of this loop implies a vaild AND or failed OR operation

			if (compareState == CompareStateCurr || compareState == CompareStateBoth) && rule.PropertyState == CurrentState {
				if !isRuleApplicable(currProperty, filterProperty.Name, &rule) {
					if !ruleSkipable && !filterSkipable {
						return false
					}
				} else {
					anyCurrMatch = true
				}
			}

			if (compareState == CompareStatePrev || compareState == CompareStateBoth) && rule.PropertyState == PreviousState {
				if !isRuleApplicable(prevProperty, filterProperty.Name, &rule) {
					if !ruleSkipable && !filterSkipable {
						return false
					}
				} else {
					anyPrevMatch = true
				}
			}
		}

		if !filterSkipable {

			// is it an OR operation ? either previous or current should have a match
			if !validateMatch(anyCurrMatch, anyPrevMatch, compareState, ruleSkipable) {
				return false
			}

		} else if validateMatch(anyCurrMatch, anyPrevMatch, compareState, ruleSkipable) {
			return true
		}
	}

	if !filterSkipable {
		return true
	} else if anyfilterTrue {
		return true
	}

	return false
}

// isRuleApplicable compare property based on rule provided
func isRuleApplicable(properties *map[string]interface{},
	propertyName string, rule *CRMFilterRule) bool {

	if propertyValue, exists := (*properties)[propertyName]; exists && propertyValue != nil {
		if comparisonOp[rule.Operator](rule.Value, propertyValue) {
			return true
		}
	} else {
		if comparisonOp[rule.Operator](rule.Value, "") {
			return true
		}
	}

	return false
}

// list of comparision operators for CRM filter
const (
	COMPARE_EQUAL        = "EQUAL"
	COMPARE_NOT_EQUAL    = "NOT EQUAL"
	COMPARE_GREATER_THAN = "GREATER THAN"
	COMPARE_LESS_THAN    = "LESS THAN"
	COMPARE_CONTAINS     = "CONTAINS"
	COMPARE_NOT_CONTAINS = "NOT CONTAINS"
)

func validateMatch(anyCurrMatch, anyPrevMatch bool, compareMode string, ruleSkipable bool) bool {
	switch compareMode {
	case CompareStateBoth:
		return (anyCurrMatch && anyPrevMatch) || (ruleSkipable && (anyCurrMatch || anyPrevMatch))
	case CompareStateCurr:
		return anyCurrMatch
	case CompareStatePrev:
		return anyPrevMatch
	default:
		return false
	}
}

// comparisonOp is map of comparision operator  and its function
var comparisonOp = map[string]func(interface{}, interface{}) bool{
	COMPARE_EQUAL: func(rValue, pValue interface{}) bool {
		if rValue == U.PROPERTY_VALUE_ANY { // should not be blank
			return pValue != ""
		}
		if reflect.ValueOf(pValue).Kind() == reflect.Bool {
			if strconv.FormatBool(pValue.(bool)) == rValue {
				return true
			} else {
				return false
			}
		}

		return rValue == pValue
	},
	COMPARE_NOT_EQUAL: func(rValue, pValue interface{}) bool {
		if rValue == U.PROPERTY_VALUE_ANY { // value not equal to anything
			return pValue == ""
		}

		if reflect.ValueOf(pValue).Kind() == reflect.Bool {
			if strconv.FormatBool(pValue.(bool)) != rValue {
				return true
			} else {
				return false
			}
		}

		return rValue != pValue
	},
	COMPARE_GREATER_THAN: func(rValue, pValue interface{}) bool {
		if reflect.ValueOf(pValue).Kind() == reflect.Bool {
			return false
		}
		intRValue, _ := U.GetPropertyValueAsFloat64(rValue)
		intpValue, _ := U.GetPropertyValueAsFloat64(pValue)
		return intpValue > intRValue
	},
	COMPARE_LESS_THAN: func(rValue, pValue interface{}) bool {
		if reflect.ValueOf(pValue).Kind() == reflect.Bool {
			return false
		}
		intRValue, _ := U.GetPropertyValueAsFloat64(rValue)
		intpValue, _ := U.GetPropertyValueAsFloat64(pValue)
		return intpValue < intRValue
	},
	COMPARE_CONTAINS: func(rValue, pValue interface{}) bool {
		if reflect.ValueOf(pValue).Kind() == reflect.Bool {
			return false
		}
		rValueStr := U.GetPropertyValueAsString(rValue)
		pValueStr := U.GetPropertyValueAsString(pValue)
		if rValueStr == "" || pValueStr == "" {
			return false
		}

		return strings.Contains(pValueStr, rValueStr)
	},
	COMPARE_NOT_CONTAINS: func(rValue, pValue interface{}) bool {
		if reflect.ValueOf(pValue).Kind() == reflect.Bool {
			return false
		}
		rValueStr := U.GetPropertyValueAsString(rValue)
		pValueStr := U.GetPropertyValueAsString(pValue)
		if pValueStr == "" {
			return true
		}

		return !strings.Contains(pValueStr, rValueStr)
	},
}

func toggleNoneOperator(operator string) string {
	if operator == COMPARE_EQUAL {
		return COMPARE_NOT_EQUAL
	}

	if operator == COMPARE_NOT_EQUAL {
		return COMPARE_EQUAL
	}

	return operator
}

/*
HandleSmartEventNoneTypeValue Convert $none to internal keyword for backend compatiblity

We use PROPERTY_VALUE_ANY const in backend for CRM rule validation
$none will be converted to ANY with logic
value != $none  -->> value == PROPERTY_VALUE_ANY
value == $none  ->-> value != PROPERTY_VALUE_ANY

*/
func HandleSmartEventNoneTypeValue(filterExpr *SmartCRMEventFilter) {
	for _, filter := range filterExpr.Filters {
		for k := range filter.Rules {
			if filter.Rules[k].Value == PropertyValueNone {
				filter.Rules[k].Operator = toggleNoneOperator(filter.Rules[k].Operator) // only toggle for equal and not equal. Other operator will be blocked on validation
				filter.Rules[k].Value = U.PROPERTY_VALUE_ANY
			}
		}
	}
}

/*
HandleSmartEventAnyTypeValue convert internal to $none keyword for frontend compatiblity

We use PROPERTY_VALUE_ANY const in backend for CRM rule validation
ANY will be converted to $none with logic
value == PROPERTY_VALUE_ANY -->> value != $none
value != PROPERTY_VALUE_ANY -->> value == $none

*/
func HandleSmartEventAnyTypeValue(filterExpr *SmartCRMEventFilter) {
	for _, filter := range filterExpr.Filters {
		for k := range filter.Rules {
			if filter.Rules[k].Value == U.PROPERTY_VALUE_ANY {
				filter.Rules[k].Operator = toggleNoneOperator(filter.Rules[k].Operator) // only toggle for equal and not equal. Other operator will be blocked on validation
				filter.Rules[k].Value = PropertyValueNone
			}
		}
	}
}

func isValidSmartCRMFilterObjectType(smartCRMFilter *SmartCRMEventFilter) bool {
	if smartCRMFilter.Source == SmartCRMEventSourceSalesforce {
		typeInt := GetSalesforceDocTypeByAlias(smartCRMFilter.ObjectType)
		if typeInt != 0 {
			return true
		}
	}

	if smartCRMFilter.Source == SmartCRMEventSourceHubspot {
		if smartCRMFilter.ObjectType == HubspotDocumentTypeNameContact ||
			smartCRMFilter.ObjectType == HubspotDocumentTypeNameDeal {
			return true
		}
	}

	return false
}

func isValidSmartCRMFilterOperator(operator string) bool {
	if _, exists := comparisonOp[operator]; !exists {
		return false
	}
	return true
}

func isValidSmartCRMFilterLogicalOp(logicalOp string) bool {
	if logicalOp != LOGICAL_OP_AND && logicalOp != LOGICAL_OP_OR {
		return false
	}
	return true
}

// Validates smart event filter expression
func IsValidSmartEventFilterExpr(smartCRMFilter *SmartCRMEventFilter) bool {
	if smartCRMFilter == nil {
		return false
	}

	if smartCRMFilter.TimestampReferenceField == "" ||
		smartCRMFilter.FilterEvaluationType != FilterEvaluationTypeSpecific &&
			smartCRMFilter.FilterEvaluationType != FilterEvaluationTypeAny {
		return false
	}

	if !isValidSmartCRMFilterObjectType(smartCRMFilter) {
		return false
	}

	if len(smartCRMFilter.Filters) < 1 {
		return false
	}

	for i := range smartCRMFilter.Filters {
		if smartCRMFilter.Filters[i].Name == "" {
			return false
		}

		if smartCRMFilter.FilterEvaluationType == FilterEvaluationTypeAny {
			if len(smartCRMFilter.Filters[i].Rules) > 0 { // for any change, rules not required
				return false
			}
			continue
		}

		if !isValidSmartCRMFilterLogicalOp(smartCRMFilter.Filters[i].LogicalOp) {
			return false
		}

		if len(smartCRMFilter.Filters[i].Rules) < 2 { // avoid single rule filter, require prev and curr property rule
			return false
		}

		var anyCurr bool
		var anyPrev bool
		for _, rule := range smartCRMFilter.Filters[i].Rules {
			if !isValidSmartCRMFilterOperator(rule.Operator) {
				return false
			}

			if rule.PropertyState == CurrentState {
				anyCurr = true
			}

			if rule.PropertyState == PreviousState {
				anyPrev = true
			}

			if rule.Value == "" {
				return false
			}

			if rule.Value == U.PROPERTY_VALUE_ANY && rule.Operator != COMPARE_EQUAL && rule.Operator != COMPARE_NOT_EQUAL {
				return false
			}
		}

		if anyCurr == false || anyPrev == false {
			return false
		}
	}

	return true
}

// IsFilterMatch checks for exact match of filter and uri passed.
// skips uri_token, if filter_token prefixed with semicolon (URI_PROPERTY_PREFIX).
func IsFilterMatch(tokenizedFilter []string, tokenizedMatchURI []string) bool {
	if len(tokenizedMatchURI) != len(tokenizedFilter) {
		return false
	}

	lastIndexTF := len(tokenizedFilter) - 1
	for i, ftoken := range tokenizedFilter {
		if !strings.HasPrefix(ftoken, URI_PROPERTY_PREFIX) {
			// filter_token is not property, should be == uri_token.
			if ftoken != tokenizedMatchURI[i] {
				return false
			}
		} else {
			// last index of filter_token as property with uri_token as "". edge case.
			if i == lastIndexTF && tokenizedMatchURI[0] == "" {
				return false
			}
		}
	}

	return true
}

// AddSmartEventReferenceMeta adds reference_id and meta for debuging purpose
func AddSmartEventReferenceMeta(properties *map[string]interface{}, eventID string) {
	if eventID != "" {
		(*properties)[U.EP_CRM_REFERENCE_EVENT_ID] = eventID
	}
}

// Today's keys
func GetPropertiesByEventCategoryCacheKey(projectId int64, event_name string, property string, category string, date string) (*cacheRedis.Key, error) {
	prefix := "EN:PC"
	return cacheRedis.NewKey(projectId, fmt.Sprintf("%s:%s", prefix, event_name), fmt.Sprintf("%s:%s:%s", date, category, property))
}
func GetEventNamesOrderByOccurrenceAndRecencyCacheKey(projectId int64, event_name string, date string) (*cacheRedis.Key, error) {
	prefix := "EN"
	return cacheRedis.NewKey(projectId, prefix, fmt.Sprintf("%s:%s", date, event_name))
}

func GetSmartEventNamesOrderByOccurrenceAndRecencyCacheKey(projectId int64, event_name string, date string) (*cacheRedis.Key, error) {
	prefix := "EN:SE"
	return cacheRedis.NewKey(projectId, prefix, fmt.Sprintf("%s:%s", date, event_name))
}

func GetValuesByEventPropertyCacheKey(projectId int64, event_name string, property_name string, value string, date string) (*cacheRedis.Key, error) {
	prefix := "EN:PV"
	return cacheRedis.NewKey(projectId, fmt.Sprintf("%s:%s:%s", prefix, event_name, property_name), fmt.Sprintf("%s:%s", date, value))
}

// For sortedsets
func GetPropertiesByEventCategoryCacheKeySortedSet(projectId int64, date string) (*cacheRedis.Key, error) {
	prefix := "SS:EN:PC"
	return cacheRedis.NewKey(projectId, fmt.Sprintf("%s", prefix), fmt.Sprintf("%s", date))
}
func GetEventNamesOrderByOccurrenceAndRecencyCacheKeySortedSet(projectId int64, date string) (*cacheRedis.Key, error) {
	prefix := "SS:EN"
	return cacheRedis.NewKey(projectId, prefix, fmt.Sprintf("%s", date))
}

func GetSmartEventNamesOrderByOccurrenceAndRecencyCacheKeySortedSet(projectId int64, date string) (*cacheRedis.Key, error) {
	prefix := "SS:EN:SE"
	return cacheRedis.NewKey(projectId, prefix, fmt.Sprintf("%s", date))
}

func GetValuesByEventPropertyCacheKeySortedSet(projectId int64, date string) (*cacheRedis.Key, error) {
	prefix := "SS:EN:PV"
	return cacheRedis.NewKey(projectId, fmt.Sprintf("%s", prefix), fmt.Sprintf("%s", date))
}

// Rollup keys
func GetPropertiesByEventCategoryRollUpCacheKey(projectId int64, event_name string, date string) (*cacheRedis.Key, error) {
	prefix := "RollUp:EN:PC"
	return cacheRedis.NewKey(projectId, fmt.Sprintf("%s:%s", prefix, event_name), date)
}
func GetEventNamesOrderByOccurrenceAndRecencyRollUpCacheKey(projectId int64, date string) (*cacheRedis.Key, error) {
	prefix := "RollUp:EN"
	return cacheRedis.NewKey(projectId, prefix, date)
}

func GetValuesByEventPropertyRollUpCacheKey(projectId int64, event_name string, property_name string, date string) (*cacheRedis.Key, error) {
	prefix := "RollUp:EN:PV"
	return cacheRedis.NewKey(projectId, fmt.Sprintf("%s:%s:%s", prefix, event_name, property_name), date)
}

// Today's keys count per project used for clean up
func GetPropertiesByEventCategoryCountCacheKey(projectId int64, dateKey string) (*cacheRedis.Key, error) {
	prefix := "C:EN:PC"
	return cacheRedis.NewKeyWithAllProjectsSupport(projectId, prefix, dateKey)
}
func GetEventNamesOrderByOccurrenceAndRecencyCountCacheKey(projectId int64, dateKey string) (*cacheRedis.Key, error) {
	prefix := "C:EN"
	return cacheRedis.NewKeyWithAllProjectsSupport(projectId, prefix, dateKey)
}

func GetValuesByEventPropertyCountCacheKey(projectId int64, dateKey string) (*cacheRedis.Key, error) {
	prefix := "C:EN:PV"
	return cacheRedis.NewKeyWithAllProjectsSupport(projectId, prefix, dateKey)

}

// Analytics Cache keys
func UniqueEventNamesAnalyticsCacheKey(dateKey string) (*cacheRedis.Key, error) {
	prefix := "SS:A:EN"
	return cacheRedis.NewKeyWithOnlyPrefix(fmt.Sprintf("%s:%s", prefix, dateKey))
}
func UserCountAnalyticsCacheKey(dateKey string) (*cacheRedis.Key, error) {
	prefix := "SS:A:UC"
	return cacheRedis.NewKeyWithOnlyPrefix(fmt.Sprintf("%s:%s", prefix, dateKey))
}
func EventsCountAnalyticsCacheKey(dateKey string) (*cacheRedis.Key, error) {
	prefix := "SS:A:EC"
	return cacheRedis.NewKeyWithOnlyPrefix(fmt.Sprintf("%s:%s", prefix, dateKey))
}
func EventCountKeyByDocumentType(documentType string, dateKey string) (*cacheRedis.Key, error) {
	prefix := "SS:A:CK"
	return cacheRedis.NewKeyWithOnlyPrefix(fmt.Sprintf("%s:%s:%s", prefix, documentType, dateKey))
}

// FillEventPropertiesByFilterExpr - Parses and fills event properties
// from tokenized_event_uri using tokenized_filter_expr.
func FillEventPropertiesByFilterExpr(eventProperties *U.PropertiesMap,
	filterExpr string, eventURL string) error {

	parsedEventURL, err := U.ParseURLStable(eventURL)
	if err != nil {
		return err
	}
	tokenizedEventURI := U.TokenizeURI(U.GetURLPathWithHash(parsedEventURL))

	parsedFilterExpr, err := U.ParseURLWithoutProtocol(filterExpr)
	if err != nil {
		return err
	}
	tokenizedFilterExpr := U.TokenizeURI(U.GetURLPathWithHash(parsedFilterExpr))

	for pos := 0; pos < len(tokenizedFilterExpr); pos++ {
		if strings.HasPrefix(tokenizedFilterExpr[pos], URI_PROPERTY_PREFIX) {
			// Adding semicolon removed filter_expr_token as key and event_uri_token as value.
			(*eventProperties)[strings.TrimPrefix(tokenizedFilterExpr[pos],
				URI_PROPERTY_PREFIX)] = tokenizedEventURI[pos]
		}
	}

	return nil
}

func isCachePrefixTypeSmartEvent(prefix string) bool {
	prefixes := strings.SplitN(prefix, ":", 2)
	if len(prefixes) == 2 && prefixes[1] == EVENT_NAME_TYPE_SMART_EVENT {
		return true
	}
	return false
}

func GetCacheEventObject(events []*cacheRedis.Key, eventCounts []string) CacheEventNamesWithTimestamp {
	eventNames := make(map[string]U.CountTimestampTuple)
	for index, eventCount := range eventCounts {
		key, value := ExtractKeyDateCountFromCacheKey(eventCount, events[index].Suffix)
		if isCachePrefixTypeSmartEvent(events[index].Prefix) {
			value.Type = EVENT_NAME_TYPE_SMART_EVENT
		}

		eventNames[key] = value
	}
	cacheEventNames := CacheEventNamesWithTimestamp{
		EventNames: eventNames}
	return cacheEventNames
}

func GetCachePropertyValueObject(values []*cacheRedis.Key, valueCounts []string) U.CachePropertyValueWithTimestamp {
	propertyValues := make(map[string]U.CountTimestampTuple)
	for index, valuesCount := range valueCounts {
		key, value := ExtractKeyDateCountFromCacheKey(valuesCount, values[index].Suffix)
		propertyValues[key] = value
	}
	cachePropertyValues := U.CachePropertyValueWithTimestamp{
		PropertyValue: propertyValues}
	return cachePropertyValues
}

func extractCategoryProperty(categoryProperty string) (string, string, string) {
	catPr := strings.SplitN(categoryProperty, ":", 3)
	return catPr[0], catPr[1], catPr[2]
}

func GetCachePropertyObject(properties []*cacheRedis.Key, propertyCounts []string) U.CachePropertyWithTimestamp {
	var dateKeyInTime time.Time
	eventProperties := make(map[string]U.PropertyWithTimestamp)
	propertyCategory := make(map[string]map[string]int64)
	for index, propertiesCount := range propertyCounts {
		dateKey, cat, pr := extractCategoryProperty(properties[index].Suffix)
		dateKeyInTime, _ = time.Parse(U.DATETIME_FORMAT_YYYYMMDD, dateKey)
		if propertyCategory[pr] == nil {
			propertyCategory[pr] = make(map[string]int64)
		}
		catCount, _ := strconv.Atoi(propertiesCount)
		propertyCategory[pr][cat] = int64(catCount)
	}
	for pr, catCount := range propertyCategory {
		cwc := make(map[string]int64)
		totalCount := int64(0)
		for cat, catCount := range catCount {
			cwc[cat] = catCount
			totalCount += catCount
		}
		prWithTs := U.PropertyWithTimestamp{CategorywiseCount: cwc,
			CountTime: U.CountTimestampTuple{Count: totalCount, LastSeenTimestamp: dateKeyInTime.Unix()}}
		eventProperties[pr] = prWithTs
	}
	cacheProperties := U.CachePropertyWithTimestamp{
		Property: eventProperties}
	return cacheProperties
}

func ExtractKeyDateCountFromCacheKey(keyCount string, cacheKey string) (string, U.CountTimestampTuple) {
	dateKey := strings.SplitN(cacheKey, ":", 2)

	keyDate, _ := time.Parse(U.DATETIME_FORMAT_YYYYMMDD, dateKey[0])
	KeyCountNum, _ := strconv.Atoi(keyCount)
	return dateKey[1], U.CountTimestampTuple{
		LastSeenTimestamp: keyDate.Unix(),
		Count:             int64(KeyCountNum),
	}
}
