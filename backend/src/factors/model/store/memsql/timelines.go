package memsql

import (
	"encoding/json"
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) GetProfilesListByProjectId(projectID int64, payload model.TimelinePayload, profileType string, downloadLimitGiven bool) ([]model.Profile, int, string) {
	logFields := log.Fields{
		"project_id":   projectID,
		"payload":      payload,
		"profile_type": profileType,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if projectID == 0 {
		return []model.Profile{}, http.StatusBadRequest, "Project Id is Invalid"
	}

	// set Query Timezone
	timezoneString, statusCode := store.GetTimezoneForProject(projectID)
	if statusCode != http.StatusFound {
		log.WithFields(logFields).Error("Query failed. Failed to get Timezone.")
		return []model.Profile{}, statusCode, "Failed to fetch project timezone."
	}
	payload.Query.Timezone = string(timezoneString)

	// set TableProps
	if len(payload.Query.TableProps) == 0 {
		payload.Query.TableProps = store.GetTablePropsFromConfig(projectID, profileType)
	}

	if payload.Query.EventsWithProperties != nil && len(payload.Query.EventsWithProperties) > 0 {
		if !C.IsEnabledEventsFilterInSegments() {
			var profiles = make([]model.Profile, 0)
			return profiles, http.StatusBadRequest, "Event filters not enabled for the project."
		}

		if payload.SearchFilter != nil {
			searchFilters := GenerateSearchFilterQueryProperties(profileType, payload.Query.Source, payload.SearchFilter)
			payload.Query.GlobalUserProperties = append(payload.Query.GlobalUserProperties, searchFilters...)
		}

		profiles, errCode, err := store.GetAnalyzeResultForSegments(projectID, profileType, payload.Query, downloadLimitGiven)
		if errCode != http.StatusOK {
			return []model.Profile{}, errCode, err.Error()
		}

		if profileType == model.PROFILE_TYPE_ACCOUNT && payload.Query.GroupAnalysis == model.GROUP_NAME_DOMAINS {
			groupedFilters := GroupFiltersByGroupName(payload.Query.GlobalUserProperties)
			profiles, statusCode = store.AccountPropertiesForDomainsEnabled(projectID, profiles, groupedFilters, payload.Query.TableProps)
			if statusCode != http.StatusOK {
				return []model.Profile{}, statusCode, "Query Transformation Failed."
			}
		}

		returnData, err := FormatProfilesStruct(projectID, profiles, profileType, payload.Query.TableProps, payload.Query.Source)
		if err != nil {
			log.WithFields(logFields).WithField("status", err).Error("Failed to filter properties from profiles.")
			return []model.Profile{}, http.StatusInternalServerError, "Failed Formatting Profile Results"
		}
		return returnData, http.StatusFound, ""
	}

	for i, p := range payload.Query.GlobalUserProperties {
		if v, exist := model.IN_PROPERTIES_DEFAULT_QUERY_MAP[p.Property]; exist {
			v.LogicalOp = p.LogicalOp
			if p.Value == "true" {
				payload.Query.GlobalUserProperties[i] = v
			} else if p.Value == "false" || p.Value == "$none" {
				v.Operator = model.EqualsOpStr
				v.Value = "$none"
				payload.Query.GlobalUserProperties[i] = v
			}
		}
	}

	// transforming datetime filters
	addSearchFiltersToFilters := true
	groupedFilters := GroupFiltersByGroupName(payload.Query.GlobalUserProperties)
	if model.IsAccountProfiles(profileType) && model.IsDomainGroup(payload.Query.Source) && C.IsAllAccountsEnabled(projectID) {
		addSearchFiltersToFilters = false
	}
	if addSearchFiltersToFilters {
		searchFilters := GenerateSearchFilterQueryProperties(profileType, payload.Query.Source, payload.SearchFilter)
		groupedFilters[model.FILTER_TYPE_USERS] = append(groupedFilters[model.FILTER_TYPE_USERS], searchFilters...)
	}

	// to check whether the filter in account profiles is of user properties
	isUserProperty := hasUserProperty(profileType, payload.Query.GlobalUserProperties)

	// to check whether the all filters in account profiles is of user properties
	isAllUserProperties := hasAllUserProperties(payload.Query.GlobalUserProperties, profileType)

	for group, filterArray := range groupedFilters {
		for index := range filterArray {
			err := groupedFilters[group][index].TransformDateTypeFilters(timezoneString)
			if err != nil {
				log.WithFields(logFields).Error("Failed to transform payload filters.")
				return []model.Profile{}, http.StatusBadRequest, "Datetime Filters Processing Failed"
			}
		}
	}

	var params, timeWindowQParams []interface{}
	params = append(params, projectID)

	isGroupUserStmt := getGroupUserStatement(profileType, payload.Query.Source)
	sourceStmt, sourceID, err := store.GetSourceStmtWithParams(projectID, profileType, payload.Query.Source, isAllUserProperties)
	if err != nil {
		return []model.Profile{}, http.StatusBadRequest, err.Error()
	}
	if sourceID != 0 {
		params = append(params, sourceID)
	}
	timeWindowQParams = append(timeWindowQParams, params...)

	var whereStmt string
	if isUserProperty {
		whereStmt = fmt.Sprintf("WHERE users.project_id=? AND users.is_deleted = 0 %s %s", isGroupUserStmt, sourceStmt) // Common String for Queries
	} else {
		whereStmt = fmt.Sprintf("WHERE project_id=? AND is_deleted = 0 %s %s", isGroupUserStmt, sourceStmt) // Common String for Queries
	}
	// Get min and max updated_at after ordering as part of optimisation.
	limitVal := 100000
	if len(payload.Query.GlobalUserProperties) > 0 {
		limitVal = 1000000
	}
	timeWindow, errCode, errStr := store.GetTimeRangeWindow(profileType, whereStmt, limitVal, timeWindowQParams)
	if errCode != http.StatusOK {
		log.WithFields(logFields).WithField("status", errCode).Error(errStr)
		return []model.Profile{}, errCode, errStr
	}

	// Get Profiles
	var runQueryStmt string
	var queryParams []interface{}
	if model.IsAccountProfiles(profileType) && model.IsDomainGroup(payload.Query.Source) && C.IsAllAccountsEnabled(projectID) {
		runQueryStmt, queryParams, err = store.GenerateAllAccountsQueryString(projectID, payload.Query.Source, isUserProperty, isAllUserProperties, *timeWindow, groupedFilters, payload.SearchFilter, downloadLimitGiven)
		params = queryParams
	} else {
		runQueryStmt, queryParams, err = store.GenerateQueryString(projectID, profileType, payload.Query.Source, sourceStmt, isUserProperty, whereStmt, *timeWindow, groupedFilters)
		params = append(params, queryParams...)
	}
	if err != nil {
		return []model.Profile{}, http.StatusInternalServerError, err.Error()
	}

	var profiles []model.Profile
	db := C.GetServices().Db
	err = db.Raw(runQueryStmt, params...).Scan(&profiles).Error
	if err != nil {
		log.WithError(err).WithFields(logFields).WithField("status", err).Error("Failed to get profile users.")
		return []model.Profile{}, http.StatusInternalServerError, "Query Execution Failed."
	}

	// Get merged properties for all accounts
	if model.IsAccountProfiles(profileType) && C.IsDomainEnabled(projectID) && model.IsDomainGroup(payload.Query.Source) {
		if isAllUserProperties && !C.IsAllAccountsEnabled(projectID) {
			userDomains, _ := store.GetUsersAssociatedToDomain(projectID, timeWindow, groupedFilters)
			profiles = appendProfiles(profiles, userDomains)
		}
		profiles, statusCode = store.AccountPropertiesForDomainsEnabled(projectID, profiles, groupedFilters, payload.Query.TableProps)
		if statusCode != http.StatusOK {
			return []model.Profile{}, statusCode, "Query Transformation Failed."
		}
	}

	// Get Return Table Content
	returnData, err := FormatProfilesStruct(projectID, profiles, profileType, payload.Query.TableProps, payload.Query.Source)
	if err != nil {
		log.WithError(err).WithFields(logFields).WithField("status", err).Error("Failed to filter properties from profiles.")
		return []model.Profile{}, http.StatusInternalServerError, "Query formatting failed."
	}

	return returnData, http.StatusFound, ""
}

// remove none filters for building where condition
func RemoveNoneFiltersFromWhere(groupedProps map[string][]model.QueryProperty) (map[string][]model.QueryProperty, map[string][]model.QueryProperty) {
	nullPropertyMap := make(map[string][]model.QueryProperty)
	nullPropsNames := make(map[string]bool)
	if len(groupedProps) == 0 {
		return groupedProps, nullPropertyMap
	}

	for groupName, filterArray := range groupedProps {
		if groupName == model.FILTER_TYPE_USERS {
			continue
		}
		for _, filter := range filterArray {
			noneFilterCheck := filter.Value == model.PropertyValueNone
			if noneFilterCheck && filter.Entity == model.PropertyEntityUserGlobal {
				nullPropsNames[filter.Property] = true
			}
		}
	}

	resultMap := make(map[string][]model.QueryProperty)
	for groupName, filterArray := range groupedProps {
		for _, filter := range filterArray {
			if nullPropsNames[filter.Property] {
				nullPropertyMap[filter.Property] = append(nullPropertyMap[filter.Property], filter)
				continue
			}
			resultMap[groupName] = append(resultMap[groupName], filter)
		}
	}

	return resultMap, nullPropertyMap
}

// hasUserProperty checks for user properties in filters.
func hasUserProperty(profileType string, properties []model.QueryProperty) bool {
	isUserProperty := false

	if model.IsAccountProfiles(profileType) {
		for _, filter := range properties {
			if filter.Entity == model.PropertyEntityUserGroup {
				isUserProperty = true
				break
			}
		}
	}

	return isUserProperty
}

// getGroupUserStatement generates a where statement indicating whether the user is a group user or not
func getGroupUserStatement(profileType, source string) string {
	isGroupUserStmt := ""
	if model.IsDomainGroup(source) {
		return isGroupUserStmt
	}
	if model.IsAccountProfiles(profileType) {
		isGroupUserStmt = "AND users.is_group_user=1"
	} else if model.IsUserProfiles(profileType) {
		isGroupUserStmt = "AND (is_group_user=0 OR is_group_user IS NULL)"
	}
	return isGroupUserStmt
}

// GetSourceStmtWithParams generates a where statement for source of the user/account. returns a statement with source value.
func (store *MemSQL) GetSourceStmtWithParams(projectID int64, profileType, source string, hasAllUserProperties bool) (string, int, error) {
	sourceStmt := ""
	sourceID := 0
	status := http.StatusOK
	if model.IsAccountProfiles(profileType) {
		if C.IsDomainEnabled(projectID) {
			sourceStmt, sourceID, status = store.GetSourceStringForAccountsV2(projectID, source, hasAllUserProperties)
		} else {
			sourceStmt, status = store.GetSourceStringForAccountsV1(projectID, source)
		}
		if status != http.StatusOK {
			return "", 0, fmt.Errorf("failed retrieving account source")
		}
	} else if model.IsUserProfiles(profileType) {
		if model.UserSourceMap[source] == model.UserSourceWeb {
			sourceStmt = "AND (source=1 OR source IS NULL)"
		} else if model.IsSourceAllUsers(source) {
			sourceStmt = ""
		} else {
			sourceStmt = "AND source=?"
			sourceID = model.UserSourceMap[source]
		}
	}
	return sourceStmt, sourceID, nil
}

// GetTimeRangeWindow returns timestamps used for windowing the profiles listing query
func (store *MemSQL) GetTimeRangeWindow(profileType string, whereStmt string, limitVal int, timeWindowQParams []interface{}) (*model.ListingTimeWindow, int, string) {
	var timeWindow model.ListingTimeWindow
	windowSelectStr := "MIN(last_activity) AS lower_bound, MAX(last_activity) AS upper_bound"

	fromStr := fmt.Sprintf("%s AND last_activity < ?", whereStmt)
	timeWindowQParams = append(timeWindowQParams, model.FormatTimeToString(gorm.NowFunc()))

	queryStrmt := fmt.Sprintf("SELECT %s FROM (SELECT COALESCE(last_event_at, updated_at) AS last_activity FROM users %s ORDER BY last_activity DESC LIMIT %d);", windowSelectStr, fromStr, limitVal)
	db := C.GetServices().Db
	err := db.Raw(queryStrmt, timeWindowQParams...).Scan(&timeWindow).Error
	if err != nil {
		return nil, http.StatusInternalServerError, "Failed Setting Time Range."
	}
	return &timeWindow, http.StatusOK, ""
}

// buildFilterStringAndParams generates a where string and a list of parameters for property filters.
func buildFilterStringAndParams(projectID int64, groupedFilters map[string][]model.QueryProperty) (string, []interface{}, error) {
	var filterString string
	var filterParams []interface{}
	var filtersArray []string

	for group, filters := range groupedFilters {
		if group == model.FILTER_TYPE_USERS {
			continue
		}
		filtersForSource, filterParamsForSource, err := buildWhereFromProperties(projectID, filters, 0)
		if err != nil {
			return "", nil, fmt.Errorf("filters ftring build failed")
		}
		if filtersForSource == "" {
			continue
		}
		filtersForSource = "(" + filtersForSource + ")"
		filtersArray = append(filtersArray, filtersForSource)
		filterParams = append(filterParams, filterParamsForSource...)
	}

	switch len(filtersArray) {
	case 0:
		break
	case 1:
		filterString = filtersArray[0]
	default:
		filterString = strings.Join(filtersArray, " OR ")
	}

	userTypeFilters, userTypeFiltersParams, errMsg := buildWhereFromProperties(projectID, groupedFilters[model.FILTER_TYPE_USERS], 0)
	if errMsg != nil {
		return "", nil, fmt.Errorf("filters ftring build failed")
	}
	if userTypeFilters != "" {
		if filterString != "" {
			filterString = filterString + " AND (" + userTypeFilters + ")"
		} else {
			filterString = "(" + userTypeFilters + ")"
		}
		filterParams = append(filterParams, userTypeFiltersParams...)
	}

	return filterString, filterParams, nil
}

// GenerateAllAccountsQueryString generates the final query used to fetch the list of profiles.
func (store *MemSQL) GenerateAllAccountsQueryString(
	projectID int64,
	source string,
	hasUserProperty bool,
	isAllUserProperties bool,
	timeWindow model.ListingTimeWindow,
	groupedFilters map[string][]model.QueryProperty,
	searchFilter []string,
	downloadLimitGiven bool,
) (string, []interface{}, error) {
	logFields := log.Fields{
		"project_id":             projectID,
		"source":                 source,
		"has_user_property":      hasUserProperty,
		"is_all_user_properties": isAllUserProperties,
		"time_window":            timeWindow,
		"grouped_filters":        groupedFilters,
		"search_filter":          searchFilter,
	}

	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	params := []interface{}{projectID, model.UserSourceDomains, timeWindow.LowerBound, timeWindow.UpperBound}

	domainGroup, errCode := store.GetGroup(projectID, model.GROUP_NAME_DOMAINS)
	if errCode != http.StatusFound || domainGroup == nil {
		return "", params, fmt.Errorf("failed to get group while adding group info")
	}

	var isGroupUserCheck, allUsersWhere string
	if !hasUserProperty && (len(groupedFilters) > 0) {
		isGroupUserCheck = "AND u.is_group_user=1"
	}
	if isAllUserProperties && (len(groupedFilters) > 0) {
		isGroupUserCheck = "AND (u.is_group_user=0 OR u.is_group_user IS NULL)"
	}

	whereForGroups := make(map[string]string)
	paramsForGroupFilters := make(map[string][]interface{})
	index := 0

	allUserFilterArray := BuildAllUsersFilterArray(groupedFilters)
	allUsersWhere, allfilterParams, err := buildWhereFromProperties(projectID, allUserFilterArray, 0)
	if err != nil {
		return "", params, err
	}
	if len(allUserFilterArray) > 0 {
		allUsersWhere = strings.ReplaceAll(allUsersWhere, "users.properties", "properties")
		allUsersWhere = strings.ReplaceAll(allUsersWhere, "user_global_user_properties", "properties")
		allUsersWhere = "WHERE " + allUsersWhere
		params = append(params, allfilterParams...)
	}

	for groupName, filters := range groupedFilters {
		whereStr, filterParams, err := buildWhereFromProperties(projectID, filters, 0)
		if err != nil {
			return "", params, err
		}
		whereStr = strings.ReplaceAll(whereStr, "users.properties", "properties")
		whereStr = strings.ReplaceAll(whereStr, "user_global_user_properties", "properties")
		whereForGroups[groupName] = whereStr
		paramsForGroupFilters[groupName] = filterParams
		index++
	}

	whereForSearchFilters, searchFiltersParams := SearchFilterForAllAccounts(searchFilter, domainGroup.ID)
	domainQParams := []interface{}{projectID, model.UserSourceDomains}
	domainQParams = append(domainQParams, searchFiltersParams...)
	params = append(domainQParams, params...)

	// building filter steps for query
	var filterSteps string
	stepNumber := 1
	for groupName, filterString := range whereForGroups {
		if groupName == U.GROUP_NAME_DOMAINS {
			continue
		}
		isGroupStr := "is_group_user=1"
		if groupName == model.FILTER_TYPE_USERS {
			isGroupStr = "(is_group_user=0 OR is_group_user IS NULL)"
		}

		filterSteps = filterSteps + fmt.Sprintf(`, filter_%d as (
			SELECT identity, domain_name FROM all_users WHERE %s AND 
				%s
		)`, stepNumber, isGroupStr, filterString)

		params = append(params, paramsForGroupFilters[groupName]...)

		stepNumber++
	}

	// Domain Level Properties Filter Statement
	var domainFilterStmt string
	if filterStmt, exists := whereForGroups[U.GROUP_NAME_DOMAINS]; exists {
		domainFilterStmt = "AND " + filterStmt
	}

	// check if payload contains a "NULL", "notEquals", "notContains"
	requireSpecialFilter, negativeFilters := CheckForNegativeFilters(groupedFilters)

	if requireSpecialFilter {
		specialFilterParams := []interface{}{projectID, model.UserSourceDomains, timeWindow.LowerBound, timeWindow.UpperBound}
		specialFilterString, filterparams, err := BuildSpecialFilter(projectID, negativeFilters, domainGroup.ID, isGroupUserCheck)
		if err != nil {
			return "", params, err
		}
		specialFilterParams = append(specialFilterParams, filterparams...)
		params = append(params, specialFilterParams...)
		filterSteps = filterSteps + ", " + specialFilterString
	}

	// building intersect step for the query
	var intersectStep string
	delete(groupedFilters, U.GROUP_NAME_DOMAINS)
	for stepNo := 1; stepNo <= len(groupedFilters); stepNo++ {
		if len(groupedFilters) == 1 {
			var addSpecialFilter string
			if requireSpecialFilter {
				addSpecialFilter = "WHERE filter_1.identity NOT IN (SELECT identity FROM filter_special) "
			}
			intersectStep = fmt.Sprintf(`SELECT 
			filter_1.identity, filter_1.domain_name
			FROM filter_1 %s
			GROUP BY filter_1.identity`, addSpecialFilter)
			break
		}
		if stepNo == 1 {
			intersectStep = `SELECT 
			filter_1.identity, filter_1.domain_name FROM filter_1 `
		}
		intersectStep = intersectStep + fmt.Sprintf(`INNER JOIN filter_%d 
		ON filter_%d.identity = filter_%d.identity `, stepNo+1, stepNo, stepNo+1)

		if stepNo == len(groupedFilters)-1 {
			var addSpecialFilter string
			if requireSpecialFilter {
				addSpecialFilter = " WHERE filter_1.identity NOT IN (SELECT identity FROM filter_special) "
			}
			intersectStep = intersectStep + addSpecialFilter + `GROUP BY filter_1.identity`
			break
		}
	}

	if len(groupedFilters) == 0 {
		intersectStep = `SELECT 
		identity, domain_name 
		FROM all_users 
		GROUP BY identity`
	}

	var domLimit int
	if downloadLimitGiven {
		domLimit = 10000
	} else {
		domLimit = 1000
	}

	finalStep := fmt.Sprintf(`, final_step as (%s)
		SELECT 
        final_step.identity as identity, final_step.domain_name as domain_name, MAX(users.last_event_at) as last_activity 
		FROM final_step JOIN (SELECT last_event_at, group_%d_user_id FROM users WHERE users.project_id=? AND users.source!=? AND 
		last_event_at IS NOT NULL ) AS users ON 
		final_step.identity = users.group_%d_user_id
		GROUP BY identity 
		ORDER BY last_activity DESC LIMIT %d;`, intersectStep, domainGroup.ID, domainGroup.ID, domLimit)

	query := fmt.Sprintf(`WITH all_users as (
		SELECT * FROM (
		  SELECT u.properties,
		  u.last_event_at as last_activity,
		  d.id as identity,
		  d.group_%d_id as domain_name,
		  u.is_group_user,
		  u.customer_user_id
		  FROM users as u
		  JOIN (
			SELECT id, group_%d_id
            FROM users 
            WHERE project_id = ?
              AND source = ?
			  %s %s) as d
		  ON u.group_%d_user_id = d.id 
		  WHERE u.project_id = ? %s 
		    AND u.source != ? 
		    AND last_activity BETWEEN ? AND ?
		  LIMIT 10000000 
		) %s
	) %s %s`, domainGroup.ID, domainGroup.ID, domainFilterStmt, whereForSearchFilters, domainGroup.ID, isGroupUserCheck, allUsersWhere, filterSteps, finalStep)

	if len(paramsForGroupFilters[U.GROUP_NAME_DOMAINS]) > 0 {
		temp := append([]interface{}{}, params[2:]...)
		params = append(params[:2], paramsForGroupFilters[U.GROUP_NAME_DOMAINS]...)
		params = append(params, temp...)
	}

	params = append(params, projectID, model.UserSourceDomains)

	logCtx.Info("Generated query for all accounts: ", query)
	return query, params, nil
}

func GenerateSearchFilterQueryProperties(profileType string, source string, searchFilter []string) []model.QueryProperty {
	var property string
	searchFilterProperties := make([]model.QueryProperty, 0)

	if model.IsUserProfiles(profileType) {
		property = U.UP_USER_ID
	} else if model.IsAccountProfiles(profileType) {
		property = model.AccountNames[source]
	} else {
		return searchFilterProperties
	}

	logicalOp := "AND"
	for index, filterValue := range searchFilter {
		if index > 0 {
			logicalOp = "OR"
		}
		queryStruct := model.QueryProperty{
			Entity:    model.PropertyEntityUserGlobal,
			Type:      U.PropertyTypeCategorical,
			GroupName: source,
			Property:  property,
			Operator:  model.ContainsOpStr,
			Value:     filterValue,
			LogicalOp: logicalOp,
		}
		searchFilterProperties = append(searchFilterProperties, queryStruct)
	}

	return searchFilterProperties
}

func SearchFilterForAllAccounts(searchFilters []string, domainID int) (string, []interface{}) {
	var searchFilterWhere string
	var searchFilterParams []interface{}
	for index, filter := range searchFilters {
		searchFilterWhere = searchFilterWhere + fmt.Sprintf("group_%d_id RLIKE ?", domainID)
		if index < len(searchFilters)-1 {
			searchFilterWhere = searchFilterWhere + " OR "
		}
		searchFilterParams = append(searchFilterParams, model.GetDomainFromURL(filter))
	}
	if searchFilterWhere != "" {
		searchFilterWhere = fmt.Sprintf("AND (%s)", searchFilterWhere)
	}
	return searchFilterWhere, searchFilterParams
}

func BuildAllUsersFilterArray(groupedFilters map[string][]model.QueryProperty) []model.QueryProperty {
	allUsersFilters := make([]model.QueryProperty, 0)
	for groupName, filterArray := range groupedFilters {
		if groupName == U.GROUP_NAME_DOMAINS {
			continue
		}
		for _, filter := range filterArray {
			filter.LogicalOp = "OR"
			allUsersFilters = append(allUsersFilters, filter)
		}
	}

	return allUsersFilters
}

func CheckForNegativeFilters(groupedFilters map[string][]model.QueryProperty) (bool, []model.QueryProperty) {
	negativeFilterExists := false
	negativeFilters := make([]model.QueryProperty, 0)
	for _, filterArray := range groupedFilters {
		if filteredProperties := model.GetPropertyToHasNegativeFilter(filterArray); len(filteredProperties) > 0 {
			negativeFilterExists = true
			negativeFilters = append(negativeFilters, filteredProperties...)
		}
	}

	return negativeFilterExists, negativeFilters
}

func BuildSpecialFilter(projectID int64, negativeFilters []model.QueryProperty, domainGroupID int, isGroupUserCheck string) (string, []interface{}, error) {
	var buildWhereString string

	var filterParams []interface{}
	negatedFilters := make([]model.QueryProperty, 0)

	for _, filter := range negativeFilters {
		if filter.Operator == model.NotContainsOpStr && filter.Value != model.PropertyValueNone {
			filter.Operator = model.ContainsOpStr
		} else if filter.Operator == model.NotEqualOpStr && filter.Value != model.PropertyValueNone {
			filter.Operator = model.EqualsOpStr
		} else if filter.Operator == model.EqualsOpStr && filter.Value == model.PropertyValueNone {
			filter.Operator = model.NotEqualOpStr
		} else if filter.Operator == model.ContainsOpStr && filter.Value == model.PropertyValueNone {
			filter.Operator = model.NotContainsOpStr
		} else if filter.Operator == model.NotInList && filter.Value != model.PropertyValueNone {
			filter.Operator = model.InList
		}
		filter.LogicalOp = "OR"
		negatedFilters = append(negatedFilters, filter)
	}

	var err error
	if len(negatedFilters) > 0 {
		buildWhereString, filterParams, err = buildWhereFromProperties(projectID, negatedFilters, 0)
		if err != nil {
			return "", filterParams, err
		}
		buildWhereString = strings.ReplaceAll(buildWhereString, "users.properties", "properties")
		buildWhereString = strings.ReplaceAll(buildWhereString, "user_global_user_properties", "properties")
		buildWhereString = "WHERE " + buildWhereString
	}
	isGroupUserCheck = strings.ReplaceAll(isGroupUserCheck, "u.is_group_user", "is_group_user")
	isGroupUserCheck = strings.ReplaceAll(isGroupUserCheck, "u.customer_user_id", "customer_user_id")

	query := fmt.Sprintf(`filter_special as (
		SELECT 
		  * 
		FROM 
		  (
			SELECT 
			  properties, 
			  group_%d_user_id as identity,
			  last_event_at as last_activity
			FROM 
			  users 
			WHERE 
			  project_id = ? %s
			  AND source != ? 
			  AND group_%d_user_id IS NOT NULL 
			  AND last_activity BETWEEN ?
			  AND ? 
			LIMIT 
			  10000000
		  ) %s
		GROUP BY identity
	  )`, domainGroupID, isGroupUserCheck, domainGroupID, buildWhereString)

	return query, filterParams, nil
}

// GenerateQueryString generates the final query used to fetch the list of profiles.
func (store *MemSQL) GenerateQueryString(
	projectID int64,
	profileType string,
	source string,
	sourceStmt string,
	hasUserProperty bool,
	whereStmt string,
	timeWindow model.ListingTimeWindow,
	groupedFilters map[string][]model.QueryProperty,
) (string, []interface{}, error) {
	var params []interface{}
	var queryString, selectString, selectColumnsStr, fromStr, groupByStr string

	isDomainGroup := (C.IsDomainEnabled(projectID) && model.IsDomainGroup(source))

	filterString, filterParams, err := buildFilterStringAndParams(projectID, groupedFilters)
	if err != nil {
		return "", params, err
	}

	if model.IsAccountProfiles(profileType) {
		if hasUserProperty && source != "All" {
			group, errCode := store.GetGroup(projectID, source)
			if errCode != http.StatusFound || group == nil {
				return "", params, fmt.Errorf("failed to get group while adding group info")
			}
			joinStr := fmt.Sprintf("JOIN users as user_user_g ON users.id = user_user_g.group_%d_user_id", group.ID)
			whereStmt = joinStr + " " + whereStmt
		}

		selectString = "id AS identity, properties, last_activity"
		selectColumnsStr = "users.id, users.properties,  users.last_event_at as last_activity"

		// selecting property col of users in case of user props in account profiles
		if hasUserProperty {
			selectColumnsStr = selectColumnsStr + ", user_user_g.properties as user_global_user_properties"
		}

	} else if model.IsUserProfiles(profileType) {
		selectString = "COALESCE(customer_user_id, id) AS identity, ISNULL(customer_user_id) AS is_anonymous, properties"
		if filterString != "" {
			selectString = selectString + ", MAX(last_activity) AS last_activity"
		} else {
			selectString = selectString + ", last_event_at as last_activity"
		}
		selectColumnsStr = "id, customer_user_id, properties, last_event_at as last_activity"
	}

	groupByStr = "GROUP BY identity"
	timeAndRecordsLimit := "last_activity BETWEEN ? AND ? LIMIT 100000000"
	params = append(params, model.FormatTimeToString(timeWindow.LowerBound), model.FormatTimeToString(timeWindow.UpperBound))

	if filterString != "" {
		fromStr = fmt.Sprintf("(SELECT %s FROM users %s AND ", selectColumnsStr, whereStmt) +
			timeAndRecordsLimit + " ) AS select_view WHERE"

		if model.IsUserProfiles(profileType) || !isDomainGroup {
			filterString = strings.ReplaceAll(filterString, "users.", "select_view.") // Json Filters on select_view

			if hasUserProperty {
				filterString = strings.ReplaceAll(filterString, "(user_global_user_properties", "(select_view.user_global_user_properties") // Json Filters on select_view
			}
		}
	} else {
		fromStr = fmt.Sprintf("users %s AND last_activity BETWEEN ? AND ?", whereStmt)
	}

	if model.IsAccountProfiles(profileType) && isDomainGroup {
		filtersMapForWhere, nullPropertyMap := RemoveNoneFiltersFromWhere(groupedFilters)
		filterString, filterParams, err = buildFilterStringAndParams(projectID, filtersMapForWhere)
		if err != nil {
			return "", params, err
		}
		var queryParams []interface{}
		var err error
		queryString, queryParams, err = store.BuildQueryStringForDomains(projectID, filterString, hasUserProperty, source, sourceStmt, timeAndRecordsLimit, groupedFilters, nullPropertyMap)
		if err != nil {
			return "", params, err
		}
		params = append(params, queryParams...)
	} else {
		queryString = fmt.Sprintf("SELECT %s FROM %s %s %s ORDER BY last_activity DESC LIMIT 1000;", selectString, fromStr, filterString, groupByStr)
	}
	params = append(params, filterParams...)

	return queryString, params, nil
}

/*
BuildQueryStringForDomains generates the query for profiles listing for 'All Accounts' case

Sample Query :-
SELECT domain_groups.id as identity, users.properties as properties, domain_groups.updated_at as last_activity FROM (
SELECT properties, group_6_user_id FROM users WHERE project_id=2 AND source != 9 AND group_6_user_id IS NOT NULL
AND updated_at BETWEEN '2023-03-07 14:38:54.494786' AND '2023-04-07 14:38:54.494786' LIMIT 1000000) AS users JOIN (
SELECT id, updated_at FROM users WHERE project_id = 2 AND source = 9 AND is_group_user = 1 AND group_6_id IS NOT NULL
) AS domain_groups ON users.group_6_user_id = domain_groups.id WHERE JSON_EXTRACT_STRING(users.properties, "$6signal_city") = "Delhi"
GROUP BY identity ORDER BY last_activity DESC LIMIT 1000;

*/

func (store *MemSQL) BuildQueryStringForDomains(projectID int64, filterString string, hasUserProperty bool, source string, sourceStmt string,
	userTimeAndRecordsLimit string, filters map[string][]model.QueryProperty, nullFilterMap map[string][]model.QueryProperty) (string, []interface{}, error) {

	var params []interface{}

	domainGroup, errCode := store.GetGroup(projectID, U.GROUP_NAME_DOMAINS)
	if errCode != http.StatusFound || domainGroup == nil {
		return "", params, fmt.Errorf("failed to get domain group while adding group info")
	}

	whereForUserQuery := "WHERE project_id=? " + sourceStmt
	params = append(params, projectID, strconv.Itoa(model.UserSourceMap[model.UserSourceDomainsString]))

	userPropsJoin := ""

	// Join in case of "All" accounts with user properties
	if hasUserProperty {
		var errString string
		var param interface{}
		userPropsJoin, param, errString = store.GetUserPropertiesForAccounts(projectID, source)
		params = append(params, param)
		if errString != "" {
			return "", params, fmt.Errorf(errString)
		}
	}

	whereForDomainGroupQuery := fmt.Sprintf(strings.Replace(whereForUserQuery, "users.source!=", "source=",
		1) + " AND is_group_user = 1")
	whereForUserQuery = whereForUserQuery + " AND " + userTimeAndRecordsLimit
	selectUserColumnsString := fmt.Sprintf("properties, last_event_at as last_activity, group_%d_user_id, id, customer_user_id, is_group_user, group_%d_id", domainGroup.ID, domainGroup.ID)
	userQueryString := fmt.Sprintf("(SELECT " + selectUserColumnsString + " FROM users " + whereForUserQuery + " ) AS users")
	selectDomainGroupColString := fmt.Sprintf("SELECT id, group_%d_id FROM users", domainGroup.ID)
	domainGroupQueryString := "( " + selectDomainGroupColString + " " + whereForDomainGroupQuery +
		" ) AS domain_groups"
	onCondition := fmt.Sprintf("ON users.group_%d_user_id = domain_groups.id", domainGroup.ID)
	groupByStr := "GROUP BY identity"
	selectString := fmt.Sprintf("domain_groups.id AS identity, users.properties as properties, MAX(users.last_activity) AS last_activity, domain_groups.group_%d_id as domain_name", domainGroup.ID)
	selectFilterString, havingString, havingFilterParams := SelectFilterAndHavingStringsForAccounts(filters, nullFilterMap)
	if selectFilterString != "" {
		selectString = selectString + ", " + selectFilterString
		if havingFilterParams != nil {
			params = append(params, havingFilterParams...)
		}
	}
	queryString := "SELECT " + selectString + " FROM " + userQueryString + " JOIN " + domainGroupQueryString + " " +
		onCondition

	if userPropsJoin != "" {
		queryString = queryString + " " + userPropsJoin
	}

	if filterString != "" {
		queryString = queryString + " WHERE " + filterString
	}
	if selectFilterString != "" {
		queryString = queryString + " " + groupByStr + " " + havingString + " ORDER BY last_activity DESC LIMIT 1000;"
	} else {
		queryString = queryString + " " + groupByStr + " ORDER BY last_activity DESC LIMIT 1000;"
	}
	return queryString, params, nil
}

// GetTablePropsFromConfig gets us the table properties from timelines default configuration
func (store *MemSQL) GetTablePropsFromConfig(projectID int64, profileType string) []string {
	timelinesConfig, err := store.GetTimelinesConfig(projectID)
	if err != nil {
		return nil
	}
	if model.IsAccountProfiles(profileType) {
		return timelinesConfig.AccountConfig.TableProps
	} else if model.IsUserProfiles(profileType) {
		return timelinesConfig.UserConfig.TableProps
	}
	return nil
}

// hasAllUserProperties checks for all user properties in filters.
func hasAllUserProperties(filters []model.QueryProperty, profileType string) bool {
	isAllUserProperties := true

	if model.IsAccountProfiles(profileType) {
		for _, filter := range filters {

			if filter.Entity != model.PropertyEntityUserGroup {

				isAllUserProperties = false
				break
			}
		}
	}
	return isAllUserProperties
}

// Function to merge unique profiles

func appendProfiles(profiles, userDomainProfiles []model.Profile) []model.Profile {
	for _, userProfile := range userDomainProfiles {
		exists := false

		for _, profile := range profiles {
			if profile.Identity == userProfile.Identity {
				exists = true
				break
			}
		}
		if !exists {
			profiles = append(profiles, userProfile)
		}
	}
	return profiles
}

// GetUserPropertiesForAccounts generates the additional join statement required for the case when account filters has user level properties
func (store *MemSQL) GetUserPropertiesForAccounts(projectID int64, source string) (string, interface{}, string) {
	groupIdsMap, errCode := store.GetGroupNameIDMap(projectID)
	if errCode != http.StatusFound {
		return "", nil, "No CRMs for the project"
	}

	var selectColArr []string

	for groupName, id := range groupIdsMap {
		if model.IsAllowedAccountGroupNames(groupName) {
			selectColArr = append(selectColArr, fmt.Sprintf("%d", id))
		}
	}

	var selectCol, whereString, onString string
	selectCol = "properties as user_global_user_properties, is_group_user, "
	onString = "("
	whereString = "project_id = ? AND ("
	param := projectID

	for index, col := range selectColArr {
		selectCol = selectCol + fmt.Sprintf("group_%s_user_id", col)
		whereString = whereString + fmt.Sprintf("users.group_%s_id IS NOT NULL", col)
		onString = onString + fmt.Sprintf("users.id = user_user_g.group_%s_user_id", col)

		if index != len(selectColArr)-1 {
			selectCol = selectCol + ", "
			whereString = whereString + " OR "
			onString = onString + " OR "
		}
	}

	whereString = whereString + ")"
	onString = onString + ") AND (user_user_g.is_group_user = 0 OR user_user_g.is_group_user IS NULL)"

	joinStmnt := fmt.Sprintf("JOIN ( SELECT %s FROM users WHERE %s) AS user_user_g ON %s", selectCol, whereString, onString)

	return joinStmnt, param, ""
}

func (store *MemSQL) GetUsersAssociatedToDomain(projectID int64, timeWindow *model.ListingTimeWindow, groupedFilters map[string][]model.QueryProperty) ([]model.Profile, int) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	var userProfiles []model.Profile

	filterString, filterParams, err := buildFilterStringAndParams(projectID, groupedFilters)
	if err != nil || filterString == "" {
		return nil, http.StatusOK
	}

	timeAndRecordsLimit := "last_activity BETWEEN ? AND ? LIMIT 100000000"
	limitParams := []interface{}{model.FormatTimeToString(timeWindow.LowerBound), model.FormatTimeToString(timeWindow.UpperBound)}

	groupIdsMap, status := store.GetGroupNameIDMap(projectID)
	if status != http.StatusFound {
		logCtx.Error("Failed to get groups info.")
		return nil, status
	}
	if _, ok := groupIdsMap[model.GROUP_NAME_DOMAINS]; !ok {
		logCtx.Error("Domain group not found.")
		return userProfiles, http.StatusBadRequest
	}

	var selectColArr []string
	for groupName, id := range groupIdsMap {
		if model.IsAllowedAccountGroupNames(groupName) {
			selectColArr = append(selectColArr, fmt.Sprintf("group_%d_user_id IS NULL", id))
		}
	}
	colStr := strings.Join(selectColArr, " AND ")

	domainID := groupIdsMap[model.GROUP_NAME_DOMAINS]
	query := fmt.Sprintf(`SELECT domain_groups.id AS identity, 
	  user_global_user_properties as properties, 
	  MAX(users.last_activity) AS last_activity, 
	  domain_groups.group_%d_id as domain_name 
	FROM (
		SELECT id,
		  properties as user_global_user_properties, 
		  last_event_at as last_activity,
		  group_%d_user_id
		FROM users 
		WHERE project_id = ? 
		  AND customer_user_id IS NOT NULL 
		  AND group_%d_user_id IS NOT NULL 
		  AND %s 
		  AND %s
	  ) AS users 
	JOIN (
		SELECT id, group_%d_id 
		FROM users 
		WHERE project_id = ? 
		  AND source = ? 
		  AND is_group_user = 1
	  ) AS domain_groups 
	ON users.group_%d_user_id = domain_groups.id
	WHERE %s 
	GROUP BY identity 
	ORDER BY last_activity DESC 
	LIMIT 1000;`, domainID, domainID, domainID, colStr, timeAndRecordsLimit, domainID, domainID, filterString)

	queryParams := []interface{}{projectID}
	queryParams = append(queryParams, limitParams...)
	queryParams = append(queryParams, projectID, model.UserSourceDomains)
	queryParams = append(queryParams, filterParams...)
	db := C.GetServices().Db
	err = db.Raw(query, queryParams...).Scan(&userProfiles).Error
	if err != nil {
		log.WithError(err).WithFields(logFields).WithField("status", err).Error("Failed to get profile users.")
		return nil, http.StatusInternalServerError
	}

	return userProfiles, http.StatusOK
}

func (store *MemSQL) AccountPropertiesForDomainsEnabled(projectID int64, profiles []model.Profile, groupedFilters map[string][]model.QueryProperty, tableProps []string) ([]model.Profile, int) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	if len(profiles) < 1 {
		logCtx.Error("No domain account found.")
		return nil, http.StatusOK
	}

	groups, status := store.GetGroupNameIDMap(projectID)
	if status != http.StatusFound {
		logCtx.Error("Groups not found.")
		return nil, status
	}

	domainGroupID, domainExists := groups[model.GROUP_NAME_DOMAINS]

	if !domainExists {
		logCtx.Error("Domain group not found.")
		return nil, status
	}

	var isNullCheck string
	salesforceOpportunityGroup, salesforceOpportunityExists := groups[model.GROUP_NAME_SALESFORCE_OPPORTUNITY]
	if salesforceOpportunityExists {
		isNullCheck = fmt.Sprintf("group_%d_id IS NULL AND group_%d_user_id IS NULL AND", salesforceOpportunityGroup, salesforceOpportunityGroup)
	}
	hubspotDealsGroup, hubspotDealsExist := groups[model.GROUP_NAME_HUBSPOT_DEAL]
	if hubspotDealsExist {
		isNullCheck = isNullCheck + fmt.Sprintf(" group_%d_id IS NULL AND group_%d_user_id IS NULL AND", hubspotDealsGroup, hubspotDealsGroup)
	}

	domainIDs := make([]string, len(profiles))
	for i, profile := range profiles {
		domainIDs[i] = profile.Identity
	}

	isGroupUserString := "is_group_user=1 AND"

	var groupedFiltersByName map[string]map[string][]model.QueryProperty
	if len(groupedFilters) > 0 {
		// group these
		groupedFiltersByName = groupedFiltersByPropName(groupedFilters)
	}

	// grouping tableProps by prefix
	tablePropsMap := propsMapGroupingByName(tableProps)

	propString := fetchTablePropsForDomainsOptimised(projectID, tablePropsMap)
	userDetails := []model.User{}
	db := C.GetServices().Db

	_, domainPropsExists := tablePropsMap[model.GROUP_NAME_DOMAINS]

	// only run the query when any table property belong to groups other than $domains
	if len(tablePropsMap) > 1 || !domainPropsExists {
		err := db.Table("users").Select(fmt.Sprintf("group_%d_user_id as id, "+propString+", source", domainGroupID)).
			Where(fmt.Sprintf("project_id=? AND source!=? AND is_deleted=0 AND %s %s group_%d_user_id ", isGroupUserString, isNullCheck, domainGroupID)+" IN (?)",
				projectID, model.UserSourceDomains, domainIDs).Find(&userDetails).Error
		if err != nil {
			logCtx.WithError(err).Error("Failed to get accounts associated to domains.")
			return nil, http.StatusInternalServerError
		}
	}

	// only run the query when any table property belong to $domains
	if domainPropsExists {
		// fetch all domain properties
		domainDetails, status := store.GetDomainPropertiesByID(projectID, domainIDs)

		if status == http.StatusFound && len(domainDetails) > 0 {
			userDetails = append(userDetails, domainDetails...)
		}
	}

	// map for inList and NotInList
	fileValuesMap := make(map[string]map[string]bool)

	for _, globalFilters := range groupedFilters {
		fileValuesMap = GetFileValues(projectID, globalFilters, fileValuesMap)
	}

	// map of domain ids and their decoded merged properties
	domainsIDPropsMap := make(map[string]map[string]interface{})
	for _, userDetail := range userDetails {
		propertiesDecoded, err := U.DecodePostgresJsonb(&userDetail.Properties)
		if err != nil {
			log.Error("Unable to decode account properties.")
			return nil, http.StatusInternalServerError
		}
		if _, exists := domainsIDPropsMap[userDetail.ID]; !exists {
			domainsIDPropsMap[userDetail.ID] = make(map[string]interface{})
		}

		if _, exists := model.SourceGroupUser[*userDetail.Source]; !exists {
			continue
		}

		groupName := model.SourceGroupUser[*userDetail.Source]

		if _, tablePropGroupExist := tablePropsMap[groupName]; !tablePropGroupExist {
			continue
		}

		filterMap, groupFilterExists := groupedFiltersByName[groupName]

		for _, tablePropName := range tablePropsMap[groupName] {
			if _, propExists := domainsIDPropsMap[userDetail.ID][tablePropName]; propExists {
				continue
			}

			if _, propExists := (*propertiesDecoded)[tablePropName]; !propExists {
				continue
			}
			filterArr, filterExists := filterMap[tablePropName]
			if groupFilterExists && filterExists {
				// check for property
				for _, filter := range filterArr {
					isFound := CheckPropertyOfGivenType(projectID, filter, propertiesDecoded, fileValuesMap)
					if isFound {
						domainsIDPropsMap[userDetail.ID][tablePropName] = (*propertiesDecoded)[tablePropName]
						break
					}
				}
			} else {
				domainsIDPropsMap[userDetail.ID][tablePropName] = (*propertiesDecoded)[tablePropName]
			}
		}
	}

	for index, id := range domainIDs {
		mergedProps := domainsIDPropsMap[id]
		propsEncoded, err := U.EncodeToPostgresJsonb(&mergedProps)
		if err != nil {
			log.WithFields(logFields).Error("Failed to encode account properties.")
			return nil, http.StatusInternalServerError
		}
		profiles[index].Properties = propsEncoded
	}
	return profiles, http.StatusOK
}

// Fetching accounts associated to the domain

// SELECT group_6_user_id as identity, JSON_INCLUDE_MASK(properties, '{"$6Signal_name":1,"$6Signal_industry":1,
// "$6Signal_employee_range":1, "$6Signal_annual_revenue":1, "$6Signal_region":1, "$salesforce_account_name":1}"
//AS properties, source FROM `users`  WHERE (project_id='15000001' AND source!='9' AND
// is_group_user=1 AND group_6_user_id IN ('4f88f40d-c571-4bee-b456-298c533d7ef9', 'ed68f40d-c571-4bee-b456-298c533d7ef9'));

func fetchTablePropsForDomainsOptimised(projectID int64, tablePropsMap map[string][]string) string {

	propString := "properties"

	if !C.IsTimelinesTablePropsOptEnabled(projectID) {
		return propString
	}

	tablePropsToFetch := make(map[string]int)

	for groupName, propNameList := range tablePropsMap {
		if groupName == model.GROUP_NAME_DOMAINS {
			continue
		}
		for _, propName := range propNameList {
			tablePropsToFetch[propName] = 1
		}
	}

	if len(tablePropsToFetch) > 0 {
		fieldsInclude, err := json.Marshal(tablePropsToFetch)
		if err != nil {
			return propString
		}

		propString = fmt.Sprintf("JSON_INCLUDE_MASK(properties, '%s') as properties", string(fieldsInclude))
	}

	return propString
}

// select id, properties, source from users where project_id=2 and source = 9
// and id in ('dom-id');
func (store *MemSQL) GetDomainPropertiesByID(projectID int64, domainIDs []string) ([]model.User, int) {
	var domainsDetail []model.User
	db := C.GetServices().Db
	err := db.Table("users").Select("id, properties, source").
		Where("project_id=? AND source=? AND is_deleted=false AND id IN (?)",
			projectID, model.UserSourceDomains, domainIDs).Find(&domainsDetail).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return []model.User{}, http.StatusNotFound
		}
		log.WithError(err).Error("Failed to get domain properties")
		return []model.User{}, http.StatusInternalServerError
	}

	return domainsDetail, http.StatusFound
}

func (store *MemSQL) GetSourceStringForAccountsV2(projectID int64, source string, isAllUserProperties bool) (string, int, int) {
	var sourceString string
	groupName := source
	if model.IsDomainGroup(source) {
		groupName = model.GROUP_NAME_DOMAINS
	}
	group, errCode := store.GetGroup(projectID, groupName)
	if errCode != http.StatusFound || group == nil {
		log.WithField("err_code", errCode).Error("Failed to get domain group while adding group info.")
		return sourceString, 0, http.StatusBadRequest
	}

	if model.IsDomainGroup(source) {
		sourceString = "AND users.source!=?"
		if isAllUserProperties {
			sourceString = sourceString + " " + fmt.Sprintf("AND (users.group_%d_id IS NOT NULL OR (users.customer_user_id IS NOT NULL AND users.group_%d_user_id IS NOT NULL))", group.ID, group.ID)
		} else {
			sourceString = sourceString + " " + fmt.Sprintf("AND users.group_%d_id IS NOT NULL", group.ID)
		}
	} else {
		if model.IsAllowedGroupName(source) && source == group.Name {
			sourceString = fmt.Sprintf("AND users.source!=? AND users.group_%d_id IS NOT NULL", group.ID)
		} else {
			log.WithField("err_code", errCode).Error(fmt.Sprintf("%s not enabled for this project.", source))
			return sourceString, 0, http.StatusBadRequest
		}
	}
	return sourceString, model.UserSourceDomains, http.StatusOK
}

// SelectFilterAndHavingStringsForAccounts generates a SELECT statement for filters and a HAVING clause for the case when filter properties are from multiple sources
func SelectFilterAndHavingStringsForAccounts(filtersMap map[string][]model.QueryProperty, nullFilterMap map[string][]model.QueryProperty) (string, string, []interface{}) {
	index := 1
	filterArray := make([]string, 0)
	havingArray := make([]string, 0)
	propMap := make(map[string]bool)
	var filterParams []interface{}
	for group, filterArr := range filtersMap {
		if group == model.FILTER_TYPE_USERS {
			continue
		}
		for _, filter := range filterArr {
			if exists := propMap[filter.Property]; exists {
				continue
			}

			// for case where multiple values for a property and a $none
			propArray, nullPropexists := nullFilterMap[filter.Property]
			if nullPropexists && len(propArray) > 1 {
				// select string
				if filter.Type == U.PropertyTypeNumerical {
					filterStr := fmt.Sprintf("CASE WHEN JSON_GET_TYPE(JSON_EXTRACT_JSON(properties, '%s')) = 'double' THEN  MAX(JSON_EXTRACT_DOUBLE(properties, '%s')) ELSE false END as filter_key_%d", filter.Property, filter.Property, index)
					filterArray = append(filterArray, filterStr)
				} else {
					filterStr := fmt.Sprintf("MAX(JSON_EXTRACT_STRING(properties, '%s')) as filter_key_%d", filter.Property, index)
					filterArray = append(filterArray, filterStr)
				}

				var havingProps []string

				for idx, prop := range propArray {
					propertyOp := getOp(prop.Operator, prop.Type)
					var propString string
					if idx > 0 {
						propString = fmt.Sprintf("%s ", prop.LogicalOp)
					}
					if prop.Value == model.PropertyValueNone && (prop.Operator == model.EqualsOpStr || prop.Operator == model.ContainsOpStr) {
						propString = propString + fmt.Sprintf("(filter_key_%d IS NULL OR filter_key_%d='')", index, index)
					} else if prop.Value == model.PropertyValueNone {
						propString = propString + fmt.Sprintf("filter_key_%d IS NOT NULL", index)
					} else {
						propString = propString + fmt.Sprintf("filter_key_%d %s ? ", index, propertyOp)
						filterParams = append(filterParams, prop.Value)
					}
					havingProps = append(havingProps, propString)
				}
				havingFilterStr := "(" + strings.Join(havingProps, " ") + ")"
				havingArray = append(havingArray, havingFilterStr)

				delete(nullFilterMap, filter.Property)
				index += 1
				propMap[filter.Property] = true
				continue
			}

			filterValNull := filter.Value == model.PropertyValueNone && (filter.Operator == model.EqualsOpStr || filter.Operator == model.ContainsOpStr)
			filterStr := fmt.Sprintf("MAX(JSON_EXTRACT_STRING(properties, '%s')) as filter_key_%d", filter.Property, index)
			filterArray = append(filterArray, filterStr)
			if filterValNull && len(filterArr) == 1 {
				havingArray = append(havingArray, fmt.Sprintf("(filter_key_%d IS NULL OR filter_key_%d='')", index, index))
			} else {
				havingArray = append(havingArray, fmt.Sprintf("filter_key_%d IS NOT NULL", index))
			}
			index += 1
			propMap[filter.Property] = true
		}
	}
	var selectFilterString, havingString string
	if len(filterArray) > 0 {
		selectFilterString = strings.Join(filterArray, ", ")
		havingString = "HAVING " + strings.Join(havingArray, " AND ")
	}
	return selectFilterString, havingString, filterParams
}

func (store *MemSQL) GetGroupNameIDMap(projectID int64) (map[string]int, int) {
	groups, errCode := store.GetGroups(projectID)
	if errCode != http.StatusFound {
		log.WithField("err_code", errCode).Error("Failed to get groups while adding group info.")
		return nil, errCode
	}
	groupNameIDMap := make(map[string]int)
	if len(groups) > 0 {
		for _, group := range groups {
			if group.Name == model.GROUP_NAME_DOMAINS || model.IsAllowedGroupName(group.Name) {
				groupNameIDMap[group.Name] = group.ID
			}
		}
	}
	return groupNameIDMap, http.StatusFound
}

// GetSourceStringForAccountsV1 returns a source statement for the case when domains is disabled.
func (store *MemSQL) GetSourceStringForAccountsV1(projectID int64, source string) (string, int) {
	logFields := log.Fields{
		"projectID": projectID,
		"source":    source,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	var sourceString string
	// Check for Enabled Groups
	groupNameIDMap, errCode := store.GetGroupNameIDMap(projectID)
	if errCode != http.StatusFound {
		log.WithField("err_code", errCode).Error("Failed to get groups while adding group info.")
		return sourceString, http.StatusBadRequest
	}

	var crmNames []string
	var crmIDs []int
	var crmExists []bool
	crmGroups := make([]string, 0, len(model.AccountGroupNames))
	for key := range model.AllowedGroupNames {
		crmGroups = append(crmGroups, key)
	}
	for _, crmName := range crmGroups {
		crmID, exists := groupNameIDMap[crmName]
		if exists {
			crmIDs = append(crmIDs, crmID)
			crmNames = append(crmNames, crmName)
			crmExists = append(crmExists, exists)
		}
	}

	if len(crmExists) == 0 {
		log.WithFields(logFields).Error("No CRMs Enabled for this project.")
		return sourceString, http.StatusBadRequest
	}
	for i, crmName := range crmNames {
		if source == crmName && !crmExists[i] {
			log.WithFields(logFields).Error(crmName + " Not Enabled for this project.")
			return sourceString, http.StatusBadRequest
		}
	}

	var sourceArr []string
	for i, crmName := range crmNames {
		if model.IsDomainGroup(source) || source == crmName {
			if crmExists[i] {
				sourceArr = append(sourceArr, fmt.Sprintf("group_%d_id IS NOT NULL", crmIDs[i]))
			}
		}
	}
	if len(sourceArr) == 0 {
		return "", http.StatusBadRequest
	}
	sourceStr := strings.Join(sourceArr, " OR ")
	sourceString = fmt.Sprintf("AND (%s)", sourceStr)

	return sourceString, http.StatusOK
}

// FormatProfilesStruct transforms the results into a processed version suitable for the response payload.
func FormatProfilesStruct(projectID int64, profiles []model.Profile, profileType string, tableProps []string, source string) ([]model.Profile, error) {
	if profiles == nil {
		return []model.Profile{}, nil
	}

	if model.IsAccountProfiles(profileType) {
		formatAccountProfilesList(profiles, tableProps, source)
	} else if model.IsUserProfiles(profileType) {
		formatUserProfilesList(profiles, tableProps)
	}

	return profiles, nil
}

func formatAccountProfilesList(profiles []model.Profile, tableProps []string, source string) {
	logFields := log.Fields{
		"profile_type": model.PROFILE_TYPE_ACCOUNT,
	}

	for index, profile := range profiles {
		properties, err := U.DecodePostgresJsonb(profile.Properties)
		if err != nil {
			log.WithFields(logFields).WithFields(log.Fields{"identity": profile.Identity}).WithError(err).Error("Failed decoding account properties.")
			continue
		}

		filterTableProps := filterPropertiesByKeys(properties, tableProps)
		profiles[index].TableProps = filterTableProps

		if profiles[index].DomainName == "" {
			continue
		}
	}
}

func formatUserProfilesList(profiles []model.Profile, tableProps []string) {
	logFields := log.Fields{
		"profile_type": model.PROFILE_TYPE_USER,
	}
	for index, profile := range profiles {
		properties, err := U.DecodePostgresJsonb(profile.Properties)
		if err != nil {
			log.WithFields(logFields).WithFields(log.Fields{"identity": profile.Identity}).WithError(err).Error("Failed decoding account properties.")
			continue
		}

		filterTableProps := filterPropertiesByKeys(properties, tableProps)
		profiles[index].TableProps = filterTableProps
	}
}

func filterPropertiesByKeys(properties *map[string]interface{}, keys []string) map[string]interface{} {
	filteredProps := make(map[string]interface{})
	for _, prop := range keys {
		if value, exists := (*properties)[prop]; exists {
			filteredProps[prop] = value
		}
	}
	return filteredProps
}

func (store *MemSQL) GetProfileUserDetailsByID(projectID int64, identity string, isAnonymous string) (*model.ContactDetails, int, string) {
	logFields := log.Fields{
		"project_id":   projectID,
		"id":           identity,
		"is_anonymous": isAnonymous,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if projectID == 0 || identity == "" || isAnonymous == "" {
		log.WithFields(logFields).Error("invalid payload.")
		return nil, http.StatusBadRequest, "Invalid payload."
	}
	isAnon := isAnonymous == "true"
	userId := model.COLUMN_NAME_CUSTOMER_USER_ID
	if isAnon {
		userId = model.COLUMN_NAME_ID
	}

	db := C.GetServices().Db
	var uniqueUser model.ContactDetails
	if err := db.Table("users").Select("COALESCE(customer_user_id,id) AS user_id, ISNULL(customer_user_id) AS is_anonymous, properties").
		Where("project_id=? AND "+userId+"=?", projectID, identity).
		Group("user_id").
		Order("last_event_at DESC").
		Limit(1).
		Find(&uniqueUser).Error; err != nil {
		log.WithError(err).WithFields(logFields).WithField("status", err).Error("Failed to get contact details.")
		return nil, http.StatusInternalServerError, "User Query Failed."
	}

	propertiesDecoded, err := U.DecodePostgresJsonb(uniqueUser.Properties)
	if err != nil {
		log.WithFields(logFields).WithError(err).Error("Failed decoding user properties.")
	} else {
		if name, exists := (*propertiesDecoded)[U.UP_NAME]; exists {
			uniqueUser.Name = fmt.Sprintf("%s", name)
		}
		timelinesConfig, err := store.GetTimelinesConfig(projectID)
		if err != nil {
			log.WithField("status", err).WithError(err).Error("Failed to fetch timelines_config from project_settings.")
		}
		uniqueUser.LeftPaneProps = FilterConfiguredProperties(timelinesConfig.UserConfig.TableProps, propertiesDecoded)
		uniqueUser.Milestones = FilterConfiguredProperties(timelinesConfig.UserConfig.Milestones, propertiesDecoded)
	}

	if activities, err := store.GetUserActivities(projectID, identity, userId); err == nil {
		uniqueUser.UserActivity = activities
	}

	uniqueUser.Account, err = store.GetAssociatedDomainForUser(projectID, identity, isAnon)
	if err != nil {
		log.WithField("status", err).WithError(err).Error("associated account could not be fetched.")
		uniqueUser.Account = ""
	}

	return &uniqueUser, http.StatusFound, ""
}

func (store *MemSQL) GetUserActivities(projectID int64, identity string, userId string) ([]model.UserActivity, error) {
	logFields := log.Fields{
		"project_id": projectID,
		"id":         identity,
	}

	var userActivities []model.UserActivity

	eventNamesToExcludePlaceholders := strings.Repeat("?,", len(model.ExcludedEvents)-1) + "?"

	eventsQuery := fmt.Sprintf(`SELECT events1.id AS event_id,
		event_names.name AS event_name, 
        event_names.type as event_type, 
        events1.timestamp AS timestamp, 
        events1.properties AS properties 
    FROM (
        SELECT project_id, id, event_name_id, timestamp, properties 
        FROM events 
        WHERE project_id=? AND timestamp <= ? 
        AND user_id IN (
            SELECT id FROM users WHERE project_id=? AND %s = ? ORDER BY last_event_at DESC LIMIT 100
        ) AND event_name_id NOT IN (
            SELECT id FROM event_names WHERE project_id=? AND name IN (%s)
        ) 
        ORDER BY timestamp DESC LIMIT 5000) AS events1 
    LEFT JOIN event_names
    ON events1.event_name_id=event_names.id 
    WHERE event_names.project_id=?;`, userId, eventNamesToExcludePlaceholders)

	excludedEventNamesArgs := make([]interface{}, len(model.ExcludedEvents))
	for i, name := range model.ExcludedEvents {
		excludedEventNamesArgs[i] = name
	}

	queryArgs := []interface{}{
		projectID,
		gorm.NowFunc().Unix(),
		projectID,
		identity,
		projectID,
	}
	queryArgs = append(queryArgs, excludedEventNamesArgs...)
	queryArgs = append(queryArgs, projectID)

	db := C.GetServices().Db
	rows, err := db.Raw(eventsQuery, queryArgs...).Rows()

	if err != nil || rows.Err() != nil {
		log.WithFields(logFields).WithError(err).WithError(rows.Err()).Error("Failed to get events")
		return []model.UserActivity{}, err
	}

	// Event Display Names
	eventDisplaNames := U.STANDARD_EVENTS_DISPLAY_NAMES
	errCode, projectDisplayNames := store.GetDisplayNamesForAllEvents(projectID)
	if errCode != http.StatusFound {
		log.WithError(err).WithField("project_id", projectID).Error("Error fetching display names for the project")
	} else {
		for key, value := range projectDisplayNames {
			eventDisplaNames[key] = value
		}
	}

	defer rows.Close()

	// Get Timeline Config For Event Properties
	timelinesConfig, err := store.GetTimelinesConfig(projectID)
	if err != nil {
		log.WithError(err).Error("Failed to fetch timelines_config from project_settings.")
	}

	for rows.Next() {
		var userActivity model.UserActivity
		if err := db.ScanRows(rows, &userActivity); err != nil {
			log.WithFields(logFields).WithError(err).Error("Failed scanning events list")
			return []model.UserActivity{}, err
		}

		properties, err := U.DecodePostgresJsonb(userActivity.Properties)
		if err != nil {
			log.WithFields(logFields).WithError(err).Error("Failed decoding event properties")
		} else {
			// Set AliasName
			// Virtual Events Case: Set AliasName to $page_url
			userActivity.AliasName = model.SetAliasName(userActivity.EventName, userActivity.EventType, properties)

			// Display Names
			if (*properties)[U.EP_IS_PAGE_VIEW] == true {
				userActivity.DisplayName = "Page View"
			} else {
				userActivity.DisplayName = model.SetEventDisplaName(userActivity.EventName, &eventDisplaNames)
			}

			// Event Icon
			if (*properties)[U.EP_IS_PAGE_VIEW] == true {
				// Page View Icon
				userActivity.Icon = "window"
			} else {
				userActivity.Icon = model.SetEventIcon(userActivity.EventName)
			}

			// Filtered Properties
			eventConfigProps := timelinesConfig.EventsConfig[userActivity.EventName]
			if (*properties)[U.EP_IS_PAGE_VIEW] == true {
				eventConfigProps = timelinesConfig.EventsConfig["PageView"]
			}
			userActivity.Properties = GetFilteredEventProperties(userActivity.EventName, userActivity.EventType, properties, eventConfigProps)
		}
		userActivities = append(userActivities, userActivity)
	}

	return userActivities, nil
}

func GetFilteredEventProperties(eventName string, eventType string, properties *map[string]interface{}, eventConfigProps []string) *postgres.Jsonb {
	var returnProperties *postgres.Jsonb
	filteredProperties := make(map[string]interface{})
	if len(eventConfigProps) > 0 {
		for _, prop := range eventConfigProps {
			if value, propExists := (*properties)[prop]; propExists {
				filteredProperties[prop] = value
			}
		}
	} else if model.IsEventNameTypeSmartEvent(eventType) {
		for key, value := range *properties {
			if strings.Contains(key, "$curr_") || strings.Contains(key, "$prev_") {
				filteredProperties[key] = value
			}
		}
	}
	if len(filteredProperties) > 0 {
		propertiesJSON, err := json.Marshal(filteredProperties)
		if err != nil {
			log.WithError(err).Error("filter properties marshal error.")
		}
		returnProperties = &postgres.Jsonb{RawMessage: propertiesJSON}
	} else {
		returnProperties = nil
	}
	return returnProperties
}

func (store *MemSQL) GetProfileAccountDetailsByID(projectID int64, id string, groupName string) (*model.AccountDetails, int, string) {
	logFields := log.Fields{
		"project_id": projectID,
		"id":         id,
		"group_name": groupName,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if projectID == 0 {
		log.Error("Invalid project ID.")
		return nil, http.StatusBadRequest, "Invalid project ID."
	}
	if id == "" {
		log.Error("Invalid account ID.")
		return nil, http.StatusBadRequest, "Invalid account ID."
	}
	if groupName == "" {
		log.Error("Invalid group name.")
		return nil, http.StatusBadRequest, "Invalid group name."
	}

	propertiesDecoded := make(map[string]interface{})
	var status int
	var accountDetails model.AccountDetails

	var group *model.Group
	if model.IsDomainGroup(groupName) {
		group, status = store.GetGroup(projectID, U.GROUP_NAME_DOMAINS)
	} else {
		group, status = store.GetGroup(projectID, groupName)
	}
	if status != http.StatusFound || group == nil {
		return nil, status, "Failed to get group"
	}
	groupUserString := fmt.Sprintf("group_%d_user_id=? ", group.ID)
	params := []interface{}{projectID, id}

	timelinesConfig, err := store.GetTimelinesConfig(projectID)
	if err != nil {
		log.WithError(err).Error("Failed to fetch timelines_config from project_settings.")
	}

	if C.IsDomainEnabled(projectID) {
		accountDetails, status = store.AccountPropertiesForDomainsEnabledV2(projectID, id, groupName, timelinesConfig)
	} else {
		groupUserString, propertiesDecoded, params, status = store.AccountPropertiesForDomainsDisabledV1(projectID, id)
		accountDetails = FormatAccountDetails(projectID, propertiesDecoded, groupName, accountDetails.DomainName)
		accountDetails.LeftPaneProps = FilterConfiguredProperties(timelinesConfig.AccountConfig.TableProps, &propertiesDecoded)
		accountDetails.Milestones = FilterConfiguredProperties(timelinesConfig.AccountConfig.Milestones, &propertiesDecoded)
	}

	if C.IsDomainEnabled(projectID) && status == http.StatusNotFound {
		accountDetails, propertiesDecoded, status = store.GetUserDetailsAssociatedToDomain(projectID, id)
		accountDetails = FormatAccountDetails(projectID, propertiesDecoded, groupName, accountDetails.DomainName)
		accountDetails.LeftPaneProps = make(map[string]interface{})
		accountDetails.Milestones = make(map[string]interface{})

	}

	if status != http.StatusOK {
		return nil, status, "Accounts Query Processing Failed"
	}

	if model.IsDomainGroup(groupName) {
		domainName, err := model.ConvertDomainIdToHostName(id)
		if err != nil || domainName == "" {
			log.WithFields(logFields).WithError(err).Error("Couldn't translate ID to Hostname")
		} else {
			accountDetails.DomainName = domainName
		}

	}

	additionalProp := timelinesConfig.AccountConfig.UserProp
	selectStrAdditionalProp := ""
	if additionalProp != "" {
		selectStrAdditionalProp = fmt.Sprintf("JSON_EXTRACT_STRING(properties, '%s') as additional_prop,", additionalProp)
	}

	// Timeline Query
	query := fmt.Sprintf(`SELECT COALESCE(JSON_EXTRACT_STRING(properties, '%s'), customer_user_id, id) AS user_name, %s
        COALESCE(customer_user_id, id) AS user_id, 
        ISNULL(customer_user_id) AS is_anonymous,
		LAST(properties, last_event_at) as user_properties
    FROM users 
    WHERE project_id = ?
	  AND (is_group_user = 0 OR is_group_user IS NULL)
	  AND (%s)
	  AND last_event_at IS NOT NULL
    GROUP BY user_id 
    ORDER BY last_event_at DESC 
    LIMIT 26;`, U.UP_NAME, selectStrAdditionalProp, groupUserString)

	// Get Timeline for <=25 users
	db := C.GetServices().Db
	rows, err := db.Raw(query, params...).Rows()
	if err != nil || rows.Err() != nil {
		log.WithFields(logFields).WithError(err).WithError(rows.Err()).Error("Failed to get associated users")
		return nil, http.StatusInternalServerError, "Accounts Query Building Failed"
	}
	defer U.CloseReadQuery(rows, nil)

	var accountTimeline []model.UserTimeline

	var numUsers int // Counter for the number of users processed

	for rows.Next() {
		var userTimeline model.UserTimeline
		if err := db.ScanRows(rows, &userTimeline); err != nil {
			log.WithFields(logFields).WithError(err).Error("Error scanning associated users list")
			return nil, http.StatusInternalServerError, "Accounts Query Execution Failed"
		}

		// Determine column to use as the user ID based on IsAnonymous
		var userIDStr string
		if userTimeline.IsAnonymous {
			userIDStr = model.COLUMN_NAME_ID
		} else {
			userIDStr = model.COLUMN_NAME_CUSTOMER_USER_ID
		}

		// Get the user's activities
		if activities, err := store.GetUserActivities(projectID, userTimeline.UserId, userIDStr); err == nil {
			userTimeline.UserActivities = activities
		}

		// Get filtered user properties
		if U.IsEmptyPostgresJsonb(userTimeline.UserProperties) {
			log.WithField("properties", userTimeline.UserProperties).Error("Empty or nil properties for user.")
		}
		propertiesDecoded, err := U.DecodePostgresJsonb(userTimeline.UserProperties)
		if err != nil {
			log.WithFields(logFields).WithError(err).Error("Failed decoding user properties.")
		}
		userTimeline.FilteredUserProperties = FilterConfiguredProperties(timelinesConfig.UserConfig.TableProps, propertiesDecoded)

		accountTimeline = append(accountTimeline, userTimeline)

		// Increment the number of users processed
		numUsers++
	}

	// Log a warning if users are greater than 25
	if numUsers > 25 {
		log.WithFields(logFields).Warn("Number of users greater than 25")
	}

	accountDetails.AccountTimeline = accountTimeline

	intentTimeline, err := store.GetIntentTimeline(projectID, groupName, id)
	if err != nil {
		log.WithFields(logFields).WithError(err).Error("Error retrieving intent timeline")
	} else {
		accountDetails.AccountTimeline = append(accountDetails.AccountTimeline, intentTimeline)
	}
	return &accountDetails, http.StatusFound, ""
}

// GetAccountOverview gives us a compiled response for account overview
func (store *MemSQL) GetAccountOverview(projectID int64, id, groupName string) (model.Overview, int, string) {
	logFields := log.Fields{
		"project_id": projectID,
		"id":         id,
	}

	overview := model.Overview{}

	// Get Users Count and Total Active Time
	queryParams := []interface{}{projectID} // Initialize with projectID
	groupUserString := ""
	var params []interface{}
	grpName := groupName
	if model.IsDomainGroup(groupName) {
		grpName = U.GROUP_NAME_DOMAINS
	}
	group, errGetGroup := store.GetGroup(projectID, grpName)
	if group != nil {
		groupUserString = fmt.Sprintf("group_%d_user_id=? ", group.ID)
		params = append(params, id)
	}

	db := C.GetServices().Db

	var errGetCount error
	if errGetGroup != http.StatusFound {
		errGetCount = fmt.Errorf("error retrieving parameters")
		log.WithFields(logFields).Error("Error retrieving parameters")
	} else {
		overviewQuery := fmt.Sprintf(`
		SELECT COUNT(DISTINCT(id)) AS users_count, 
			SUM(JSON_EXTRACT_STRING(properties, '%s')) AS time_active 
		FROM (
			SELECT LAST(id, last_event_at) AS id, properties 
			FROM users 
			WHERE project_id = ?
			  AND (is_group_user=0 OR is_group_user IS NULL)
			  AND (%s)
			  AND customer_user_id IS NOT NULL 
			GROUP BY customer_user_id
			UNION 
			SELECT id, properties 
			FROM users 
		  	WHERE project_id = ?
			  AND (is_group_user=0 OR is_group_user IS NULL)
			  AND (%s) 
			  AND customer_user_id IS NULL
		);`,
			U.UP_TOTAL_SPENT_TIME, groupUserString, groupUserString,
		)
		queryParams = append(queryParams, params...)      // append groupUserString params
		queryParams = append(queryParams, queryParams...) // double the queryParams

		errGetCount = db.Raw(overviewQuery, queryParams...).Scan(&overview).Error
		if errGetCount != nil {
			log.WithFields(logFields).WithError(errGetCount).Error("Error retrieving users count and active time")
		}
	}

	// Get Account Engagement Score and Trends
	accountScore, _, _, errGetScore := store.GetPerAccountScore(projectID, time.Now().Format("20060102"), id, model.NUM_TREND_DAYS, false)
	if errGetScore != nil {
		log.WithFields(logFields).WithError(errGetScore).Error("Error retrieving account score")
	} else {
		overview.ScoresList = accountScore.Trend
		overview.LastEventTS = accountScore.LastEventTimeStamp
	}

	// To Get Engagement Level And Score from Domain Properties Column
	var domainUser model.User
	if err := db.Model(&model.User{}).
		Where("project_id = ? AND id = ? AND source=?", projectID, id, model.UserSourceDomains).
		Select("properties").
		Limit(1).
		Find(&domainUser).
		Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			log.WithFields(logFields).WithError(errGetScore).Error("no record found for given domain")
		}
		log.WithFields(logFields).WithError(err).Error("Failed to get properties of given domain.")
	}

	if U.IsEmptyPostgresJsonb(&domainUser.Properties) {
		log.WithFields(logFields).WithField("properties", domainUser.Properties).Error("Empty or nil properties for user.")
	}
	propertiesDecoded, err := U.DecodePostgresJsonb(&domainUser.Properties)
	if err != nil {
		log.WithFields(logFields).WithError(err).Error("Failed decoding user properties.")
	}

	if score, exists := (*propertiesDecoded)[U.DP_ENGAGEMENT_SCORE]; exists {
		overview.Temperature = score.(float64)
	}
	if level, exists := (*propertiesDecoded)[U.DP_ENGAGEMENT_LEVEL]; exists {
		overview.Engagement = level.(string)
	}

	// Get Top Pages and Top Users
	topPages, errGetTopPages := store.GetTopPages(projectID, id, group.ID)
	if errGetTopPages != nil {
		log.WithFields(logFields).WithError(errGetScore).Error("Error getting top pages")
	} else {
		overview.TopPages = topPages
	}

	topUsers, errGetTopUsers := store.GetTopUsers(projectID, id, group.ID)
	if errGetTopUsers != nil {
		log.WithFields(logFields).WithError(errGetScore).Error("Error getting top users")
	} else {
		overview.TopUsers = topUsers
	}

	if errGetScore != nil && errGetCount != nil && errGetTopUsers != nil && errGetTopPages != nil {
		return overview, http.StatusInternalServerError, "error getting overview"
	}

	return overview, http.StatusOK, ""
}

// GetTopPages gives us a list of top pages with visited by the users associated to a group/domain ordered by number of visits
func (store *MemSQL) GetTopPages(projectID int64, id string, groupID int) ([]model.TopPage, error) {
	queryParams := []interface{}{projectID, id}
	groupUserString := fmt.Sprintf("users.group_%d_user_id=? ", groupID)

	queryStmt := fmt.Sprintf(`SELECT JSON_EXTRACT_STRING(events.properties, '%s') AS page_url,
          COUNT(events.id) AS views,
          COUNT(DISTINCT(COALESCE(users.customer_user_id, users.id))) AS users_count,
          SUM(JSON_EXTRACT_STRING(events.properties, '%s')) AS total_time,
          AVG(JSON_EXTRACT_STRING(events.properties, '%s')) AS avg_scroll_percent
        FROM users 
        JOIN events 
        ON users.id=events.user_id AND users.project_id=events.project_id
        WHERE users.project_id=?
          AND (users.is_group_user = 0 OR users.is_group_user IS NULL)
          AND (%s)
          AND JSON_EXTRACT_STRING(events.properties,  '%s') = 'true'
        GROUP BY page_url
        ORDER BY views DESC
        LIMIT 30;`, U.EP_PAGE_URL, U.EP_PAGE_SPENT_TIME, U.EP_PAGE_SCROLL_PERCENT, groupUserString, U.EP_IS_PAGE_VIEW)

	db := C.GetServices().Db
	var topPages []model.TopPage
	if err := db.Raw(queryStmt, queryParams...).Scan(&topPages).Error; err != nil {
		logFields := log.Fields{
			"project_id": projectID,
			"id":         id,
		}
		log.WithFields(logFields).WithError(err).Error("Error executing top pages query")
		return nil, fmt.Errorf("error executing top pages query")
	}

	return topPages, nil
}

// GetTopUsers gives us a list of top users associated to a group/domain ordered by activity
func (store *MemSQL) GetTopUsers(projectID int64, id string, groupID int) ([]model.TopUser, error) {
	queryParams := []interface{}{projectID, id}
	groupUserStmt := fmt.Sprintf("users.group_%d_user_id=? ", groupID)

	// known users
	topUsers, err := store.GetTopKnownUsers(queryParams, groupUserStmt)
	if err != nil {
		return nil, err
	}

	// anonymous users
	topAnonymousUsers, err := store.GetTopAnonymousUsers(queryParams, groupUserStmt)
	if err != nil {
		return nil, err
	}

	// Combine the results of known and anonymous users
	if topAnonymousUsers.AnonymousUsersCount > 0 {
		topUsers = append(topUsers, topAnonymousUsers)
	}

	return topUsers, nil
}

// GetTopUsers gives us a list of top identified users associated to a group/domain ordered by activity
func (store *MemSQL) GetTopKnownUsers(queryParams []interface{}, groupUserStmt string) ([]model.TopUser, error) {
	db := C.GetServices().Db
	queryStmt := fmt.Sprintf(`SELECT COALESCE(JSON_EXTRACT_STRING(users.properties, '%s'), users.customer_user_id) AS name,
      COUNT(events.id) as num_page_views,
      MAX(JSON_EXTRACT_STRING(users.properties, '%s')) as active_time,
      COUNT(DISTINCT(events.event_name_id)) as num_of_pages
    FROM users
    JOIN events
    ON users.id=events.user_id AND users.project_id=events.project_id
    WHERE users.project_id=?
      AND (users.is_group_user = 0 OR users.is_group_user IS NULL)
      AND (%s)
      AND users.customer_user_id IS NOT NULL
      AND JSON_EXTRACT_STRING(events.properties, '%s') = 'true'
    GROUP BY name
    ORDER BY num_page_views DESC 
    LIMIT 30;`, U.UP_NAME, U.UP_TOTAL_SPENT_TIME, groupUserStmt, U.EP_IS_PAGE_VIEW)

	var topUsers []model.TopUser
	if err := db.Raw(queryStmt, queryParams...).Scan(&topUsers).Error; err != nil {
		logFields := log.Fields{
			"project_id": queryParams[0].(int64),
			"id":         queryParams[1].(string),
		}
		log.WithFields(logFields).WithError(err).Error("Error executing top users query")
		return nil, fmt.Errorf("error executing top users query")
	}

	return topUsers, nil
}

// GetTopUsers gives us a cumulative record of unidentified/anonymous users associated to a group/domain
func (store *MemSQL) GetTopAnonymousUsers(queryParams []interface{}, groupUserStmt string) (model.TopUser, error) {
	db := C.GetServices().Db
	queryStmt := fmt.Sprintf(`SELECT COUNT(DISTINCT(users.id)) AS anonymous_users_count,
      COUNT(events.id) as num_page_views,
      SUM(JSON_EXTRACT_STRING(users.properties, '%s')) as active_time,
      COUNT(DISTINCT(events.event_name_id)) as num_of_pages
    FROM users
    JOIN events
    ON users.id=events.user_id AND users.project_id=events.project_id
    WHERE users.project_id=?
      AND (users.is_group_user = 0 OR users.is_group_user IS NULL)
      AND (%s)
      AND users.customer_user_id IS NULL
      AND JSON_EXTRACT_STRING(events.properties, '%s') = 'true'
    LIMIT 1;`, U.UP_TOTAL_SPENT_TIME, groupUserStmt, U.EP_IS_PAGE_VIEW)

	var topAnonymousUsers model.TopUser
	if err := db.Raw(queryStmt, queryParams...).Scan(&topAnonymousUsers).Error; err != nil {
		logFields := log.Fields{
			"project_id": queryParams[0].(int64),
			"id":         queryParams[1].(string),
		}
		log.WithFields(logFields).WithError(err).Error("Error executing anonymous users query")
		return model.TopUser{}, fmt.Errorf("error executing anonymous users query")
	}
	if topAnonymousUsers.AnonymousUsersCount > 0 {
		topAnonymousUsers.Name = fmt.Sprintf("%d Anonymous Users", topAnonymousUsers.AnonymousUsersCount)
	}

	return topAnonymousUsers, nil
}

func (store *MemSQL) GetIntentTimeline(projectID int64, groupName string, id string) (model.UserTimeline, error) {
	intentTimeline := model.UserTimeline{
		UserId:         id,
		IsAnonymous:    false,
		UserName:       model.GROUP_ACTIVITY_USERNAME,
		UserActivities: []model.UserActivity{},
	}

	if model.IsDomainGroup(groupName) {
		intentTimeline.ExtraProp = "All"
		groupNameIDMap, status := store.GetGroupNameIDMap(projectID)
		if status != http.StatusFound {
			return intentTimeline, fmt.Errorf("failed to retrieve GroupNameID map")
		}
		// Fetch accounts associated with the domain
		associatedAccounts, status := store.GetAccountsAssociatedToDomain(projectID, id, groupNameIDMap[model.GROUP_NAME_DOMAINS])
		if status != http.StatusFound {
			return intentTimeline, fmt.Errorf("failed to fetch associated accounts for domain ID %v", id)
		}
		// Fetch user activities for each associated account
		for _, user := range associatedAccounts {
			intentActivities, err := store.GetUserActivities(projectID, user.ID, model.COLUMN_NAME_ID)
			if err != nil {
				return intentTimeline, fmt.Errorf("failed to retrieve user activities for user ID %v", user.ID)
			}
			intentTimeline.UserActivities = append(intentTimeline.UserActivities, intentActivities...)
		}
	} else {
		if groupDisplayName, exists := U.STANDARD_GROUP_DISPLAY_NAMES[groupName]; exists {
			intentTimeline.ExtraProp = groupDisplayName
		} else {
			intentTimeline.ExtraProp = groupName
		}
		// Fetch user activities for the given account ID
		intentActivities, err := store.GetUserActivities(projectID, id, model.COLUMN_NAME_ID)
		if err != nil {
			return intentTimeline, fmt.Errorf("failed to retrieve user activities for user ID %v", id)
		}
		intentTimeline.UserActivities = intentActivities
	}

	return intentTimeline, nil
}

func (store *MemSQL) AccountPropertiesForDomainsEnabledV2(projectID int64, id, groupName string, timelinesConfig model.TimelinesConfig) (model.AccountDetails, int) {
	accountGroupDetails := make([]model.User, 0)
	if model.IsDomainGroup(groupName) {
		groupNameIDMap, status := store.GetGroupNameIDMap(projectID)
		if status != http.StatusFound {
			return model.AccountDetails{}, status
		}
		// Fetching accounts associated to the domain
		accountGroupDetails, status = store.GetAccountsAssociatedToDomain(projectID, id, groupNameIDMap[model.GROUP_NAME_DOMAINS])

		// fetch all domain properties
		domain, domStatus := store.GetDomainPropertiesByID(projectID, []string{id})
		if domStatus != http.StatusFound {
			log.Error("Failed to get account properties.")
			return model.AccountDetails{}, http.StatusInternalServerError
		}
		accountGroupDetails = append(accountGroupDetails, domain...)

		// return not found if no associated accounts found
		if status != http.StatusFound && len(accountGroupDetails) < 1 {
			return model.AccountDetails{}, http.StatusNotFound
		}
	} else {
		if !model.IsAllowedGroupName(groupName) {
			log.Error("Invalid group name.")
			return model.AccountDetails{}, http.StatusBadRequest
		}
		// Filter Properties
		properties, status := store.GetUserPropertiesByUserID(projectID, id)
		if status != http.StatusFound {
			log.Error("Failed to get account properties.")
			return model.AccountDetails{}, http.StatusInternalServerError
		}
		if properties != nil {
			var user model.User
			user.Properties = *properties
			src := model.GroupUserSource[groupName]
			user.Source = &src
			accountGroupDetails = append(accountGroupDetails, user)
		}
	}
	accountDetails, status := FetchAccountDetailsFromProps(projectID, groupName, timelinesConfig, accountGroupDetails)
	if status != http.StatusOK {
		return accountDetails, status
	}
	return accountDetails, http.StatusOK
}

func FetchAccountDetailsFromProps(projectID int64, groupName string, timelinesConfig model.TimelinesConfig,
	accountGroupDetails []model.User) (model.AccountDetails, int) {

	var accountDetails model.AccountDetails
	milestonesMap := propsMapGroupingByName(timelinesConfig.AccountConfig.Milestones)
	leftPaneMap := propsMapGroupingByName(timelinesConfig.AccountConfig.TableProps)

	filteredProperties := make(map[string]interface{})
	milestoneProperties := make(map[string]interface{})
	for _, accountGroupDetail := range accountGroupDetails {
		props, err := U.DecodePostgresJsonb(&accountGroupDetail.Properties)
		if err != nil {
			log.Error("Unable to decode account properties.")
			return accountDetails, http.StatusInternalServerError
		}

		// add name and hostname
		if accountDetails.DomainName == "" && accountDetails.Name == "" {
			accountDetails = FormatAccountDetails(projectID, *props, groupName, accountDetails.DomainName)
		}

		// allow only source from groups: skip in case of eg source = 1
		source := *accountGroupDetail.Source
		if _, exist := model.SourceGroupUser[source]; !exist {
			continue
		}
		accountSource := model.SourceGroupUser[source]

		if milestoneProps, exist := milestonesMap[accountSource]; exist {
			for _, prop := range milestoneProps {
				if _, exists := (*props)[prop]; !exists {
					continue
				}
				value := (*props)[prop]
				maxValMilestone := computeMaxMilestone(milestoneProperties, value, prop)

				if maxValMilestone <= 0 {
					continue
				}
				milestoneProperties[prop] = maxValMilestone
			}
		}

		// continue if all the leftpane props are found
		if len(filteredProperties) == len(timelinesConfig.AccountConfig.TableProps) {
			continue
		}

		if _, exist := leftPaneMap[accountSource]; !exist {
			continue
		}
		leftPaneProps := leftPaneMap[accountSource]
		for _, prop := range leftPaneProps {
			if val, valueExists := filteredProperties[prop]; valueExists && val != "" {
				continue
			}
			if value, exists := (*props)[prop]; exists && value != "" {
				filteredProperties[prop] = value
			}
		}
	}
	accountDetails.Milestones = milestoneProperties
	accountDetails.LeftPaneProps = filteredProperties
	return accountDetails, http.StatusOK
}

func computeMaxMilestone(milestoneProperties map[string]interface{},
	value interface{}, propName string) int64 {
	var err error

	var propertyValue int64
	if intVal, ok := value.(int64); ok {
		propertyValue = intVal
	} else if stringVal, ok := value.(string); ok {
		if stringVal != "" {
			propertyValue, err = strconv.ParseInt(stringVal, 10, 64)
			if err != nil {
				log.WithError(err).Error("Failed to convert property value to int type")
			}
		}
	} else if floatVal, ok := value.(float64); ok {
		propertyValue = int64(floatVal)
	}
	maxValue := propertyValue
	if existingVal, valueExists := milestoneProperties[propName]; valueExists && existingVal != "" {
		if (propertyValue) < existingVal.(int64) {
			maxValue = existingVal.(int64)
		}
	}
	return maxValue
}

func (store *MemSQL) AccountPropertiesForDomainsDisabledV1(projectID int64, id string) (string, map[string]interface{}, []interface{}, int) {
	var groupUserString string
	propertiesDecoded := make(map[string]interface{}, 0)
	var params []interface{}
	params = append(params, projectID)
	groupNameIDMap, errCode := store.GetGroupNameIDMap(projectID)
	if errCode != http.StatusFound {
		log.WithField("err_code", errCode).Error("Failed to get groups.")
		return groupUserString, propertiesDecoded, params, http.StatusNotFound
	}
	firstVal := false
	for name, groupID := range groupNameIDMap {
		if model.IsAllowedAccountGroupNames(name) {
			if !firstVal {
				groupUserString = fmt.Sprintf("group_%d_user_id=? ", groupID)
				firstVal = true
				params = append(params, id)
			} else {
				groupUserString = groupUserString + fmt.Sprintf(" OR group_%d_user_id=? ", groupID)
				params = append(params, id)
			}
		}
	}
	properties, status := store.GetUserPropertiesByUserID(projectID, id)
	if status != http.StatusFound {
		log.Error("Failed to get account properties.")
		return groupUserString, propertiesDecoded, params, http.StatusInternalServerError
	}
	props, err := U.DecodePostgresJsonb(properties)
	if err != nil {
		log.WithError(err).Error("Failed to decode account properties.")
		return groupUserString, propertiesDecoded, params, http.StatusInternalServerError
	}
	propertiesDecoded = *props

	return groupUserString, propertiesDecoded, params, http.StatusOK
}

func (store *MemSQL) GetUserDetailsAssociatedToDomain(projectID int64, id string) (model.AccountDetails, map[string]interface{}, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"id":         id,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	var accountDetails model.AccountDetails
	var paramsQuery []interface{}

	domainGroup, errCode := store.GetGroup(projectID, model.GROUP_NAME_DOMAINS)
	if errCode != http.StatusFound {
		logCtx.Error("Domain group not found")
		return accountDetails, nil, errCode
	}

	// Fetching accounts associated to the domain
	// SELECT domain_group.group_1_id AS domain_name, users.properties AS properties FROM (SELECT group_1_id, id FROM users WHERE project_id = 1000000 AND source = 9
	// AND id = "aa5f9e4e-e516-481f-86cd-bd42debda12c") AS domain_group JOIN (SELECT group_1_user_id, properties FROM users WHERE project_id = 1000000 AND (is_group_user = 0 OR
	// is_group_user IS NULL) AND group_1_user_id IS NOT NULL) AS users ON domain_group.id = users.group_1_user_id LIMIT 1;
	db := C.GetServices().Db
	query := fmt.Sprintf("SELECT domain_group.group_%d_id AS domain_name, users.properties AS properties FROM "+
		"(SELECT group_%d_id, id FROM users WHERE project_id = ? AND source = ? AND id = ?) AS domain_group JOIN "+
		"(SELECT group_%d_user_id, properties FROM users WHERE project_id = ? AND (is_group_user = 0 OR is_group_user IS NULL) "+
		"AND customer_user_id IS NOT NULL AND group_%d_user_id IS NOT NULL) AS users ON domain_group.id = users.group_%d_user_id LIMIT 1", domainGroup.ID,
		domainGroup.ID, domainGroup.ID, domainGroup.ID, domainGroup.ID)
	paramsQuery = append(paramsQuery, projectID, model.UserSourceDomains, id, projectID)
	err := db.Raw(query, paramsQuery...).Scan(&accountDetails).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to get groups.")
		return accountDetails, nil, http.StatusInternalServerError
	}

	props, err := U.DecodePostgresJsonb(accountDetails.Properties)
	if err != nil {
		log.Error("Unable to decode account properties.")
		return accountDetails, nil, http.StatusInternalServerError
	}

	propertiesDecoded := *props

	return accountDetails, propertiesDecoded, http.StatusOK
}

func (store *MemSQL) GetAccountsAssociatedToDomain(projectID int64, id string, domainGroupId int) ([]model.User, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"id":         id,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	// Fetching accounts associated to the domain
	// SELECT id, source, properties FROM users WHERE project_id = 2 AND source!=9 AND is_group_user=1 AND
	// group_6_user_id = '92d0899a-cbf3-4031-8a43-6b330e07326f'
	var accountGroupDetails []model.User
	db := C.GetServices().Db
	err := db.Table("users").Select("id, source, properties").
		Where("project_id=? AND source!=? AND is_group_user=1 AND is_deleted = false AND "+fmt.Sprintf("group_%d_user_id", domainGroupId)+"=?", projectID, model.UserSourceDomains, id).Find(&accountGroupDetails).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to get groups.")
		return []model.User{}, http.StatusInternalServerError
	}
	return accountGroupDetails, http.StatusFound
}

func FormatAccountDetails(projectID int64, propertiesDecoded map[string]interface{},
	groupName string, hostName string) model.AccountDetails {
	var companyNameProps, hostNameProps []string
	var accountDetails model.AccountDetails

	if C.IsDomainEnabled(projectID) && groupName != U.GROUP_NAME_DOMAINS {
		if model.IsAllowedGroupName(groupName) {
			hostNameProps = []string{model.HostNameGroup[groupName]}
			companyNameProps = []string{model.AccountNames[groupName], U.UP_COMPANY}
		}
	} else {
		companyNameProps = model.NameProps
		hostNameProps = model.HostNameProps
	}

	nameProps := append(companyNameProps, hostNameProps...)
	for _, prop := range nameProps {
		if name, exists := (propertiesDecoded)[prop]; exists {
			accountDetails.Name = fmt.Sprintf("%s", name)
			break
		}
	}

	if hostName != "" {
		accountDetails.DomainName = hostName
	} else {
		for _, prop := range hostNameProps {
			if host, exists := (propertiesDecoded)[prop]; exists {
				accountDetails.DomainName = fmt.Sprintf("%s", host)
				break
			}
		}
	}

	if accountDetails.Name == "" && accountDetails.DomainName != "" {
		accountDetails.Name = accountDetails.DomainName
	}
	return accountDetails
}

func FilterConfiguredProperties(propsToFilter []string, propertiesDecoded *map[string]interface{}) map[string]interface{} {

	filteredProperties := make(map[string]interface{})
	for _, prop := range propsToFilter {
		if value, exists := (*propertiesDecoded)[prop]; exists {
			filteredProperties[prop] = value
		}
	}
	return filteredProperties
}

func (store *MemSQL) GetAnalyzeResultForSegments(projectId int64, profileType string, query model.Query, downloadLimitGiven bool) ([]model.Profile, int, error) {
	logFields := log.Fields{
		"project_id": projectId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	if projectId == 0 {
		logCtx.Error("Segment event query Failed. Invalid projectID.")
		return nil, http.StatusBadRequest, fmt.Errorf("segment event query failed. invalid projectID")
	}

	// Update parameters
	if profileType == model.PROFILE_TYPE_ACCOUNT {
		query.Source = query.GroupAnalysis
	}
	query.Caller = profileType
	query.Class = model.QueryClassEvents
	query.From = U.TimeNowZ().AddDate(0, 0, -90).Unix()
	query.To = U.TimeNowZ().Unix()
	query.DownloadAccountsLimitGiven = downloadLimitGiven
	err := query.TransformDateTypeFilters()
	if err != nil {
		log.WithFields(logFields).Error("Failed to transform query payload filters.")
		return nil, http.StatusBadRequest, fmt.Errorf("segment filters processing failed")
	}

	result, errCode, errMsg := store.Analyze(projectId, query, C.EnableOptimisedFilterOnEventUserQuery(), false)
	if errCode != http.StatusOK {
		logCtx.WithField("err_code", errCode).Error("Failed at building query. ", errMsg)
		return nil, errCode, fmt.Errorf(errMsg)
	}
	profiles, err := FormatAnalyzeResultForProfiles(result, query.Caller)
	if err != nil {
		logCtx.Error("Failed at building query. ", err)
		return nil, errCode, err
	}

	return profiles, http.StatusOK, nil
}

func FormatAnalyzeResultForProfiles(result *model.QueryResult, profileType string) ([]model.Profile, error) {
	profiles := make([]model.Profile, 0)

	for _, profile := range result.Rows {
		var row model.Profile
		row.Identity = profile[0].(string)

		if model.IsUserProfiles(profileType) {
			isAnonymous := profile[1] == float64(1)
			row.IsAnonymous = isAnonymous
			row.LastActivity = profile[2].(time.Time)
			reflectProps := reflect.ValueOf(profile[3])
			props := make(map[string]interface{}, 0)
			if err := json.Unmarshal([]byte(reflectProps.String()), &props); err != nil {
				return nil, fmt.Errorf("failed at unmarshalling props")
			}
			var err error
			row.Properties, err = U.EncodeToPostgresJsonb(&props)
			if err != nil {
				return nil, fmt.Errorf("failed at encoding props")
			}
		} else if model.IsAccountProfiles(profileType) {
			row.TableProps = make(map[string]interface{}, 0)
			if profile[1] != nil {
				row.LastActivity = profile[1].(time.Time)
			}
			row.DomainName = profile[2].(string)
		}

		profiles = append(profiles, row)
	}

	return profiles, nil
}

func GroupFiltersByPrefix(filters []model.QueryProperty) map[string][]model.QueryProperty {
	filtersMap := make(map[string][]model.QueryProperty)

	for _, filter := range filters {
		var groupName string
		for _, prefix := range model.GroupPropertyPrefixList {
			if filter.Entity == model.PropertyEntityUserGroup {
				break
			}
			if strings.Contains(strings.ToLower(filter.Property), prefix) {
				groupName = prefix
				break
			}

		}

		if groupName == "" {
			groupName = model.FILTER_TYPE_USERS
		} else if groupName == U.LI_PROPERTIES_PREFIX {
			groupName = U.GROUP_NAME_LINKEDIN_COMPANY
		}

		filtersMap[groupName] = append(filtersMap[groupName], filter)
	}

	return filtersMap
}

func GroupFiltersByGroupName(filters []model.QueryProperty) map[string][]model.QueryProperty {
	filtersMap := make(map[string][]model.QueryProperty)

	for _, filter := range filters {
		if filter.GroupName == model.PROFILE_TYPE_USER {
			filtersMap[model.FILTER_TYPE_USERS] = append(filtersMap[model.FILTER_TYPE_USERS], filter)
			continue
		}
		filtersMap[filter.GroupName] = append(filtersMap[filter.GroupName], filter)
	}
	return filtersMap
}

func groupedFiltersByPropName(groupedFilters map[string][]model.QueryProperty) map[string]map[string][]model.QueryProperty {
	groupedByName := make(map[string]map[string][]model.QueryProperty)

	for name, queryPropArr := range groupedFilters {
		for _, queryProp := range queryPropArr {
			if _, exists := groupedByName[name]; exists {
				groupedByName[name][queryProp.Property] = append(groupedByName[name][queryProp.Property], queryProp)
				continue
			}
			groupedByName[name] = make(map[string][]model.QueryProperty)
			groupedByName[name][queryProp.Property] = append(groupedByName[name][queryProp.Property], queryProp)
		}
	}

	return groupedByName
}

func propsMapGroupingByName(tableProps []string) map[string][]string {
	tablePropsMap := make(map[string][]string)

	for _, tableProp := range tableProps {
		var groupName string
		for _, prefix := range model.GroupPropertyPrefixList {
			if strings.Contains(strings.ToLower(tableProp), prefix) {
				groupName = prefix
				break
			} else if domainPropertyExist(tableProp) {
				groupName = model.GROUP_NAME_DOMAINS
			}
		}

		if groupName == "" {
			groupName = model.FILTER_TYPE_USERS
		} else if groupName == U.LI_PROPERTIES_PREFIX {
			groupName = U.GROUP_NAME_LINKEDIN_COMPANY
		}

		tablePropsMap[groupName] = append(tablePropsMap[groupName], tableProp)
	}

	return tablePropsMap
}

func domainPropertyExist(property string) bool {
	for _, prop := range U.ALL_ACCOUNTS_TOTAL_PROPERTIES {
		if property == prop {
			return true
		}
	}
	return false
}

func tablePropsHasUserProperty(props []string) bool {
	for _, prop := range props {
		if !hasPrefixFromList(prop, model.GroupPropertyPrefixList) {
			return true
		}
	}
	return false
}

func hasPrefixFromList(s string, prefixes []string) bool {
	lowerS := strings.ToLower(s)
	for _, prefix := range prefixes {
		if strings.Contains(lowerS, prefix) {
			return true
		}
	}
	return false
}

func (store *MemSQL) UpdateConfigForEvent(projectID int64, eventName string, updatedConfig []string) (int, error) {
	logFields := log.Fields{
		"projectID": projectID,
		"eventName": eventName,
	}
	logCtx := log.WithFields(logFields)
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	if projectID == 0 || eventName == "" || len(updatedConfig) == 0 {
		logCtx.Error("Failed to update event config. Invalid parameters.")
		return http.StatusBadRequest, fmt.Errorf("invalid parameters")
	}

	db := C.GetServices().Db

	timelinesConfig, err := store.GetTimelinesConfig(projectID)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	timelinesConfig.EventsConfig[eventName] = updatedConfig

	tlConfigEncoded, err := U.EncodeStructTypeToPostgresJsonb(timelinesConfig)
	if err != nil {
		logCtx.WithError(err).Error("Default Timelines Config Encode Failed.")
		return http.StatusInternalServerError, err
	}

	updateFields := map[string]interface{}{
		"timelines_config": tlConfigEncoded,
	}

	err = db.Model(&model.ProjectSetting{}).Where("project_id = ?", projectID).Update(updateFields).Error
	if err != nil {
		switch {
		case gorm.IsRecordNotFoundError(err):
			return http.StatusNotFound, err
		default:
			logCtx.WithError(err).Error("Failed while updating segment on UpdateSegmentById.")
			return http.StatusInternalServerError, err
		}
	}

	return http.StatusOK, nil
}

func (store *MemSQL) GetConfiguredUserPropertiesWithValues(projectID int64, id string, isAnonymous bool) (map[string]interface{}, int) {
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

	userId := model.COLUMN_NAME_CUSTOMER_USER_ID
	if isAnonymous {
		userId = model.COLUMN_NAME_ID
	}

	var user model.User
	db := C.GetServices().Db
	if err := db.Model(&model.User{}).
		Where("project_id = ? AND "+userId+" = ?", projectID, id).
		Select("properties").
		Order("last_event_at DESC").
		Limit(1).
		Find(&user).
		Error; err != nil {
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
	propertiesDecoded, err := U.DecodePostgresJsonb(&user.Properties)
	if err != nil {
		log.WithFields(logFields).WithError(err).Error("Failed decoding user properties.")
	}
	timelinesConfig, err := store.GetTimelinesConfig(projectID)
	if err != nil {
		log.WithField("status", err).WithError(err).Error("Failed to fetch timelines_config from project_settings.")
	}
	filteredProperties := FilterConfiguredProperties(timelinesConfig.UserConfig.TableProps, propertiesDecoded)
	return filteredProperties, http.StatusOK
}

func (store *MemSQL) UpdateTimelineConfigForEngagementScoring(projectID int64, isScoringEnabled bool) error {
	// Fetch timelinesConfig
	timelinesConfig, err := store.GetTimelinesConfig(projectID)
	if err != nil {
		return fmt.Errorf("failed to fetch timelinesConfig")
	}

	updatedColumns, isUpdated := getUpdatedTableColumns(timelinesConfig.AccountConfig.TableProps, isScoringEnabled)

	if isUpdated {
		timelinesConfig.AccountConfig.TableProps = updatedColumns
		tlConfigEncoded, err := U.EncodeStructTypeToPostgresJsonb(timelinesConfig)
		if err != nil {
			return err
		}

		// Update project settings
		_, errCode := store.UpdateProjectSettings(projectID, &model.ProjectSetting{TimelinesConfig: tlConfigEncoded})
		if errCode != http.StatusAccepted {
			return fmt.Errorf("failed to update account table columns")
		}
	}

	return nil
}

func getUpdatedTableColumns(existingColumns []string, isScoringEnabled bool) ([]string, bool) {
	// columns bool map for lookup
	existingColsMap := make(map[string]bool)
	for _, column := range existingColumns {
		existingColsMap[column] = true
	}

	// Check if any of the engagement columns are present
	hasEngagementColumns := false
	for _, prop := range U.ENGAGEMENT_PROPERTIES {
		if existingColsMap[prop] {
			hasEngagementColumns = true
			break
		}
	}

	configUpdated := false // set flag for config update

	// Update columns list based on scoring feature status
	var updatedColumns []string
	if isScoringEnabled && !hasEngagementColumns {
		updatedColumns = addEngagementColumns(existingColumns)
		configUpdated = true
	} else if !isScoringEnabled && hasEngagementColumns {
		updatedColumns = removeEngagementColumns(existingColumns)
		configUpdated = true
	}

	return updatedColumns, configUpdated
}

func addEngagementColumns(existingColumns []string) []string {
	return append(U.ENGAGEMENT_PROPERTIES, existingColumns...)
}

func removeEngagementColumns(existingColumns []string) []string {
	engagementColumns := make(map[string]bool)
	for _, column := range U.ENGAGEMENT_PROPERTIES {
		engagementColumns[column] = true
	}
	var updatedColumns []string
	for _, column := range existingColumns {
		if !engagementColumns[column] {
			updatedColumns = append(updatedColumns, column)
		}
	}
	return updatedColumns
}

func (store *MemSQL) GetTopEventsForADomain(projectID int64, domainName string) ([]model.TimelineEvent, int) {
	var eventsList []model.TimelineEvent
	logFields := log.Fields{
		"project_id": projectID,
		"domain":     domainName,
	}

	domainGroup, errCode := store.GetGroup(projectID, model.GROUP_NAME_DOMAINS)
	if errCode != http.StatusFound {
		log.WithFields(logFields).Error("Domain group not found")
		return eventsList, http.StatusInternalServerError
	}

	eventNamesToExcludePlaceholders := strings.Repeat("?,", len(model.ExcludedEvents)-1) + "?"
	queryStmt := fmt.Sprintf(`SELECT event_names.name,
		event_names.type,
		events.id,
		events.timestamp,
		events.properties,
		COALESCE(domain_users.customer_user_id, domain_users.id) AS user_id,
		COALESCE(domain_users.name, domain_users.customer_user_id, domain_users.id) AS username,
		domain_users.is_group_user,
		ISNULL(domain_users.customer_user_id) AS is_anonymous_user
	FROM events
		JOIN (
		SELECT id,
			project_id,
			customer_user_id,
			JSON_EXTRACT_STRING(properties, '%s') AS name,
			is_group_user
			FROM users
		WHERE project_id = ?
			AND group_%d_user_id = (SELECT id FROM users WHERE project_id = ? AND group_%d_id = ? AND source = ? LIMIT 1)
		ORDER BY last_event_at DESC
		LIMIT 25
		) AS domain_users ON events.user_id = domain_users.id
		LEFT JOIN event_names ON events.event_name_id = event_names.id
		AND events.project_id = event_names.project_id
	WHERE events.project_id = ?
		AND event_names.name NOT IN (%s)
	ORDER BY events.timestamp DESC
	LIMIT 100;`, U.UP_NAME, domainGroup.ID, domainGroup.ID, eventNamesToExcludePlaceholders)

	excludedEventNamesArgs := make([]interface{}, len(model.ExcludedEvents))
	for i, name := range model.ExcludedEvents {
		excludedEventNamesArgs[i] = name
	}
	queryArgs := []interface{}{
		projectID,
		projectID,
		domainName,
		model.UserSourceDomains,
		projectID,
	}
	queryArgs = append(queryArgs, excludedEventNamesArgs...)

	db := C.GetServices().Db
	rows, err := db.Raw(queryStmt, queryArgs...).Rows()

	if err != nil || rows.Err() != nil {
		log.WithFields(logFields).WithError(err).WithError(rows.Err()).Error("Failed to get events")
		return []model.TimelineEvent{}, http.StatusInternalServerError
	}

	// Event Display Names
	eventDisplaNames := U.STANDARD_EVENTS_DISPLAY_NAMES
	errCode, projectDisplayNames := store.GetDisplayNamesForAllEvents(projectID)
	if errCode != http.StatusFound {
		log.WithError(err).WithField("project_id", projectID).Error("Error fetching display names for the project")
	} else {
		for key, value := range projectDisplayNames {
			eventDisplaNames[key] = value
		}
	}

	defer rows.Close()

	for rows.Next() {
		var event model.TimelineEvent
		if err := db.ScanRows(rows, &event); err != nil {
			log.WithFields(logFields).WithError(err).Error("Failed scanning events list")
			return []model.TimelineEvent{}, http.StatusInternalServerError
		}
		properties, err := U.DecodePostgresJsonb(event.Properties)
		if err != nil {
			log.WithFields(logFields).WithError(err).Error("Failed decoding event properties")
		} else {
			timelinesConfig, err := store.GetTimelinesConfig(projectID)
			if err != nil {
				log.WithError(err).Error("Failed to fetch timelines_config from project_settings.")
			}
			eventConfigProps := timelinesConfig.EventsConfig[event.Name]
			if (*properties)[U.EP_IS_PAGE_VIEW] == true {
				eventConfigProps = timelinesConfig.EventsConfig["PageView"]
			}

			formatEventData(&event, &eventDisplaNames, properties, eventConfigProps)
		}
		eventsList = append(eventsList, event)
	}

	return eventsList, http.StatusOK
}

func formatEventData(event *model.TimelineEvent, eventDisplaNames *map[string]string, properties *map[string]interface{}, configProperties []string) {
	// Set AliasName
	event.AliasName = model.SetAliasName(event.Name, event.Type, properties)

	// Display Names
	if (*properties)[U.EP_IS_PAGE_VIEW] == true {
		event.DisplayName = "Page View"
	} else {
		event.DisplayName = model.SetEventDisplaName(event.Name, eventDisplaNames)
	}

	// Event Icon
	if (*properties)[U.EP_IS_PAGE_VIEW] == true {
		// Page View Icon
		event.Icon = "window"
	} else {
		event.Icon = model.SetEventIcon(event.Name)
	}
	event.Properties = GetFilteredEventProperties(event.Name, event.Type, properties, configProperties)
}

func (store *MemSQL) GetConfiguredEventPropertiesWithValues(projectID int64, eventID string, eventName string) (map[string]interface{}, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"id":         eventID,
		"event_name": eventName,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)

	if projectID == 0 || eventID == "" || eventName == "" {
		logCtx.Error("Invalid values on arguments.")
		return nil, http.StatusBadRequest
	}

	timelinesConfig, err := store.GetTimelinesConfig(projectID)
	if err != nil {
		log.WithField("status", err).WithError(err).Error("Failed to fetch timelines_config from project_settings.")
	}
	propertiesToFilter, exists := timelinesConfig.EventsConfig[eventName]

	if !exists {
		logCtx.Error("config empty for given event.")
		return nil, http.StatusNotFound
	}

	var event model.Event
	db := C.GetServices().Db
	if err := db.Model(&model.Event{}).
		Where("project_id = ? AND id = ?", projectID, eventID).
		Select("properties").
		Find(&event).
		Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}

		logCtx.WithError(err).Error("Failed to get properties of user.")
		return nil, http.StatusInternalServerError
	}

	if U.IsEmptyPostgresJsonb(&event.Properties) {
		logCtx.WithField("properties", event.Properties).Error("Empty or nil properties for user.")
		return nil, http.StatusNotFound
	}
	propertiesDecoded, err := U.DecodePostgresJsonb(&event.Properties)
	if err != nil {
		log.WithFields(logFields).WithError(err).Error("Failed decoding user properties.")
	}

	filteredProperties := make(map[string]interface{})
	for _, prop := range propertiesToFilter {
		if value, exists := (*propertiesDecoded)[prop]; exists {
			filteredProperties[prop] = value
		}
	}
	return filteredProperties, http.StatusOK
}
