package memsql

import (
	"encoding/json"
	"errors"
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

/*
Sample Timeline Listing Queries:

// Users Listing Without Filters
SELECT MIN(updated_at) AS min_updated_at, MAX(updated_at) AS max_updated_at FROM (SELECT updated_at FROM users WHERE project_id=11000005 AND (is_group_user=0 OR is_group_user IS NULL) AND (source=1 OR source IS NULL) AND updated_at < '2022-09-15 13:07:24.336972' ORDER BY updated_at DESC LIMIT 100000);
SELECT COALESCE(customer_user_id, id) AS identity, ISNULL(customer_user_id) AS is_anonymous, JSON_EXTRACT_STRING(properties, '$country') AS country, MAX(updated_at) AS last_activity FROM users WHERE project_id=11000005 AND (is_group_user=0 OR is_group_user IS NULL) AND (source=1 OR source IS NULL) AND updated_at BETWEEN '2022-09-15 13:07:24.044412' AND '2022-09-15 13:07:24.322378'  GROUP BY identity ORDER BY last_activity DESC LIMIT 1000;
// Users Listing With Filters
SELECT MIN(updated_at) AS min_updated_at, MAX(updated_at) AS max_updated_at FROM (SELECT updated_at FROM users WHERE project_id=11000005 AND (is_group_user=0 OR is_group_user IS NULL) AND (source=1 OR source IS NULL) AND updated_at < '2022-09-15 14:11:44.769131' ORDER BY updated_at DESC LIMIT 500000);
SELECT COALESCE(customer_user_id, id) AS identity, ISNULL(customer_user_id) AS is_anonymous, JSON_EXTRACT_STRING(properties, '$country') AS country, MAX(updated_at) AS last_activity FROM (SELECT id, customer_user_id, properties, updated_at FROM users WHERE project_id=11000005 AND (is_group_user=0 OR is_group_user IS NULL) AND (source=1 OR source IS NULL) AND updated_at BETWEEN '2022-09-15 13:07:24.044412' AND '2022-09-15 13:07:24.322378'  LIMIT 1000000) AS select_view WHERE ((JSON_EXTRACT_STRING(select_view.properties, '$country') = 'Ukraine')) GROUP BY identity ORDER BY last_activity DESC LIMIT 1000;

// Users Listing Without Filters
SELECT MIN(updated_at) AS min_updated_at, MAX(updated_at) AS max_updated_at FROM (SELECT updated_at FROM users WHERE project_id=11000006 AND is_group_user=1 AND (group_1_id IS NOT NULL OR group_2_id IS NOT NULL) AND updated_at < '2022-09-15 13:23:20.702165' ORDER BY updated_at DESC LIMIT 100000);
SELECT id AS identity, properties, updated_at AS last_activity FROM users WHERE project_id=11000006 AND is_group_user=1 AND (group_1_id IS NOT NULL OR group_2_id IS NOT NULL) AND updated_at BETWEEN '2022-09-15 13:23:20.480649' AND '2022-09-15 13:23:20.615161'   ORDER BY last_activity DESC LIMIT 1000;
// Users Listing With Filters
SELECT MIN(updated_at) AS min_updated_at, MAX(updated_at) AS max_updated_at FROM (SELECT updated_at FROM users WHERE project_id=11000006 AND is_group_user=1 AND (group_1_id IS NOT NULL OR group_2_id IS NOT NULL) AND updated_at < '2022-09-15 13:23:20.702165' ORDER BY updated_at DESC LIMIT 500000);
SELECT id AS identity, properties, updated_at AS last_activity FROM (SELECT id, properties, updated_at FROM users WHERE project_id=11000006 AND is_group_user=1 AND (group_1_id IS NOT NULL OR group_2_id IS NOT NULL) AND updated_at BETWEEN '2022-09-15 13:23:20.480649' AND '2022-09-15 13:23:20.615161'  LIMIT 1000000) AS select_view WHERE ((JSON_EXTRACT_STRING(select_view.properties, '$salesforce_account_billingcountry') = 'India') OR (JSON_EXTRACT_STRING(select_view.properties, '$hubspot_company_country') = 'US'))  ORDER BY last_activity DESC LIMIT 1000;
*/
func (store *MemSQL) GetProfilesListByProjectId(projectID int64, payload model.TimelinePayload, profileType string) ([]model.Profile, int, string) {
	logFields := log.Fields{
		"project_id":   projectID,
		"payload":      payload,
		"profile_type": profileType,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if projectID == 0 {
		return nil, http.StatusBadRequest, "Project Id is Invalid"
	}

	var tableProps []string
	if payload.SegmentId != "" {
		segment, status := store.GetSegmentById(projectID, payload.SegmentId)
		if status != http.StatusFound {
			return nil, http.StatusBadRequest, "Invalid segment."
		}
		segmentQuery := &model.Query{}
		err := U.DecodePostgresJsonbToStructType(segment.Query, segmentQuery)
		if err != nil {
			return nil, http.StatusInternalServerError, "Segment Processing Failed"
		}
		payload.Source = segment.Type
		segmentQuery.Source = segment.Type
		tableProps = segmentQuery.TableProps
		segmentQuery.From = U.TimeNowZ().AddDate(0, 0, -28).Unix()
		segmentQuery.To = U.TimeNowZ().Unix()
		if !C.AllowEventAnalyticsGroupsByProjectID(projectID) || segmentQuery.Caller == model.USER_PROFILE_CALLER {
			segmentQuery.GroupAnalysis = ""
		} else {
			segmentQuery.GroupAnalysis = segment.Type
		}
		if segmentQuery.EventsWithProperties != nil && len(segmentQuery.EventsWithProperties) > 0 {
			if C.IsEnabledEventsFilterInSegments() {
				if payload.Filters != nil {
					segmentQuery.GlobalUserProperties = append(segmentQuery.GlobalUserProperties, payload.Filters...)
				}
				if payload.SearchFilter != nil {
					segmentQuery.GlobalUserProperties = append(segmentQuery.GlobalUserProperties, payload.SearchFilter...)
				}
				err := segmentQuery.TransformDateTypeFilters()
				if err != nil {
					log.WithFields(logFields).Error("Failed to transform query payload filters.")
					return nil, http.StatusBadRequest, "Segment Filters Processing Failed"
				}
				query, err := U.EncodeStructTypeToPostgresJsonb(segmentQuery)
				if err != nil {
					log.WithFields(logFields).Error("Failed to append payload filters with global properties.")
				} else {
					segment.Query = query
				}
				profiles, errCode, _ := store.GetAnalyzeResultForSegments(projectID, segment)
				if errCode != http.StatusOK {
					return nil, errCode, "Segment Query Failed"
				}
				// Get Table Content
				returnData, err := FormatProfilesStruct(profiles, profileType, tableProps, payload.Source)
				if err != nil {
					log.WithFields(logFields).WithField("status", err).Error("Failed to filter properties from profiles.")
					return nil, http.StatusInternalServerError, "Failed Formatting Profile Results"
				}
				return returnData, http.StatusFound, ""
			} else {
				var profiles = make([]model.Profile, 0)
				return profiles, http.StatusBadRequest, "Event filters not enabled for the project."
			}
		} else if segment.Type != "" {
			payload.Filters = append(payload.Filters, segmentQuery.GlobalUserProperties...)
		}
	} else {
		timelinesConfig, err := store.GetTimelineConfigOfProject(projectID)
		if err != nil {
			log.WithFields(logFields).WithField("status", err).WithError(err).Error("Failed to fetch timelines_config from project_settings.")
		}
		if profileType == model.PROFILE_TYPE_ACCOUNT {
			tableProps = timelinesConfig.AccountConfig.TableProps
		} else if profileType == model.PROFILE_TYPE_USER {
			tableProps = timelinesConfig.UserConfig.TableProps
		}
	}

	timezoneString, statusCode := store.GetTimezoneForProject(projectID)
	if statusCode != http.StatusFound {
		log.WithFields(logFields).Error("Query failed. Failed to get Timezone.")
		return nil, http.StatusBadRequest, "Failed to fetch project timezone."
	}

	// transforming datetime filters
	groupedFilters := GroupFiltersByPrefix(payload.Filters)
	groupedFilters[model.FILTER_TYPE_USERS] = append(groupedFilters[model.FILTER_TYPE_USERS], payload.SearchFilter...)

	// to check whether the filter in account profiles is of user properties
	isUserProperty := hasUserProperty(profileType, payload.Filters)

	// to check whether the all filters in account profiles is of user properties
	isAllUserProperties := hasAllUserProperties(payload.Filters, profileType)

	for group, filterArray := range groupedFilters {
		for index := range filterArray {
			err := groupedFilters[group][index].TransformDateTypeFilters(timezoneString)
			if err != nil {
				log.WithFields(logFields).Error("Failed to transform payload filters.")
				return nil, http.StatusBadRequest, "Datetime Filters Processing Failed"
			}
		}
	}

	var selectString, isGroupUserString, sourceString string
	var domainGroupId, status int
	var params, minMaxParamsQ []interface{}
	params = append(params, projectID)
	if profileType == model.PROFILE_TYPE_ACCOUNT {
		isGroupUserString = "AND users.is_group_user=1"
		selectString = "id AS identity, properties, updated_at AS last_activity"
		if C.IsDomainEnabled(projectID) {
			sourceString, domainGroupId, params, status = store.GetSourceStringForAccountsV2(projectID, payload.Source, isAllUserProperties)
		} else {
			// Check for Enabled Groups
			groupNameIDMap, errCode := store.GetGroupNameIDMap(projectID)
			if errCode != http.StatusFound {
				log.WithField("err_code", errCode).Error("Failed to get groups while adding group info.")
				return nil, http.StatusBadRequest, "Failed processing groups information."
			}
			sourceString, status = GetSourceStringForAccountsV1(groupNameIDMap, payload.Source)
		}
		if status != http.StatusOK {
			return nil, status, "Failed retrieving account source."
		}
	} else if profileType == model.PROFILE_TYPE_USER {
		selectString = "COALESCE(customer_user_id, id) AS identity, ISNULL(customer_user_id) AS is_anonymous, properties, MAX(updated_at) AS last_activity"
		isGroupUserString = "AND (is_group_user=0 OR is_group_user IS NULL)"
		if model.UserSourceMap[payload.Source] == model.UserSourceWeb {
			sourceString = "AND (source=1 OR users.source IS NULL)"
		} else if payload.Source == "All" {
			sourceString = ""
		} else {
			sourceString = "AND source=?"
			params = append(params, strconv.Itoa(model.UserSourceMap[payload.Source]))
		}
	}
	minMaxParamsQ = append(minMaxParamsQ, params...)

	var filterString string
	var filterParams []interface{}
	var filtersArray []string
	for group, filters := range groupedFilters {
		if group == model.FILTER_TYPE_USERS {
			continue
		}
		filtersForSource, filterParamsForSource, err := buildWhereFromProperties(projectID, filters, 0)
		if err != nil {
			return nil, http.StatusBadRequest, "Query building failed."
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
		return nil, http.StatusBadRequest, "Query building failed."
	}
	if userTypeFilters != "" {
		if filterString != "" {
			filterString = filterString + " AND (" + userTypeFilters + ")"
		} else {
			filterString = "(" + userTypeFilters + ")"
		}
		filterParams = append(filterParams, userTypeFiltersParams...)
	}

	if payload.Source == "All" {
		isGroupUserString = ""
	}

	// Run Queries
	type MinMaxTime struct {
		MinUpdatedAt time.Time `json:"min_updated_at"`
		MaxUpdatedAt time.Time `json:"max_updated_at"`
	}
	var minMax MinMaxTime
	var runQueryString, fromStr, groupByStr, selectColumnsStr, commonStr string
	windowSelectStr := "MIN(updated_at) AS min_updated_at, MAX(updated_at) AS max_updated_at" // Select Min & Max updated_at
	// commonStr for additional join stmt
	if isUserProperty {
		commonStr = fmt.Sprintf("WHERE users.project_id=? %s %s", isGroupUserString, sourceString) // Common String for Queries
	} else {
		commonStr = fmt.Sprintf("WHERE project_id=? %s %s", isGroupUserString, sourceString) // Common String for Queries
	}
	fromStr = fmt.Sprintf("%s AND updated_at < ?", commonStr)
	minMaxParamsQ = append(minMaxParamsQ, FormatTimeToString(gorm.NowFunc()))
	// Get min and max updated_at after ordering as part of optimisation.
	limitVal := 100000
	if filterString != "" {
		limitVal = 1000000
	}
	runQueryString = fmt.Sprintf("SELECT %s FROM (SELECT updated_at FROM users %s ORDER BY updated_at DESC LIMIT %d);", windowSelectStr, fromStr, limitVal)
	db := C.GetServices().Db
	err := db.Raw(runQueryString, minMaxParamsQ...).Scan(&minMax).Error
	if err != nil {
		log.WithError(err).WithFields(logFields).WithField("status", err).Error("min and max updated_at couldn't be defined.")
		return nil, http.StatusInternalServerError, "Failed Setting Time Range."
	}

	// Join condition is case of user property filter for accounts
	if isUserProperty && payload.Source != "All" {
		group, errCode := store.GetGroup(projectID, payload.Source)
		if errCode != http.StatusFound || group == nil {
			log.WithField("err_code", errCode).Error("Failed to get group while adding group info.")
			return nil, errCode, "Failed to get group while adding group info."
		}
		joinStr := fmt.Sprintf("JOIN users as user_user_g ON users.id = user_user_g.group_%d_user_id", group.ID)
		commonStr = joinStr + " " + commonStr
	}

	if profileType == model.PROFILE_TYPE_ACCOUNT {
		selectColumnsStr = "users.id, users.properties, users.updated_at "
		groupByStr = ""
		// selecting property col of users in case of user props in account profiles
		if isUserProperty {
			selectColumnsStr = selectColumnsStr + ", user_user_g.properties as user_global_user_properties"
		}
	} else if profileType == model.PROFILE_TYPE_USER {
		selectColumnsStr = "id, customer_user_id, properties, updated_at"
		groupByStr = "GROUP BY identity"
	}
	timeAndRecordsLimit := fmt.Sprintf("users.updated_at BETWEEN ? AND ? LIMIT 100000000")
	params = append(params, FormatTimeToString(minMax.MinUpdatedAt), FormatTimeToString(minMax.MaxUpdatedAt))
	isDomainGroup := (C.IsDomainEnabled(projectID) && payload.Source == "All")
	if filterString != "" {
		fromStr = fmt.Sprintf("(SELECT %s FROM users %s AND ", selectColumnsStr, commonStr) +
			timeAndRecordsLimit + " ) AS select_view WHERE"
		if profileType == model.PROFILE_TYPE_USER || !isDomainGroup {
			filterString = strings.ReplaceAll(filterString, "users.", "select_view.") // Json Filters on select_view
			if isUserProperty {
				filterString = strings.ReplaceAll(filterString, "(user_global_user_properties", "(select_view.user_global_user_properties") // Json Filters on select_view
			}
		}
	} else {
		fromStr = fmt.Sprintf("users %s AND updated_at BETWEEN ? AND ?",
			commonStr)
	}
	if profileType == model.PROFILE_TYPE_ACCOUNT && isDomainGroup {
		whereForUserQuery := "WHERE project_id=? " + sourceString
		params = append(params, projectID, strconv.Itoa(model.UserSourceMap[model.UserSourceDomainsString]))
		userPropertiesJoin := ""
		// Join in case of "All" accounts with user properties
		if isUserProperty {
			var errString string
			var param interface{}
			userPropertiesJoin, param, errString = store.userPropertiesForAccounts(projectID, payload.Source)
			params = append(params, param)
			if errString != "" {
				return nil, http.StatusNotFound, errString
			}
		}
		runQueryString = BuildQueryStringForDomains(filterString, whereForUserQuery, domainGroupId, timeAndRecordsLimit, groupedFilters, userPropertiesJoin)
	} else {
		runQueryString = fmt.Sprintf("SELECT %s FROM %s %s %s ORDER BY last_activity DESC LIMIT 1000;", selectString, fromStr, filterString, groupByStr)
	}
	params = append(params, filterParams...)
	var profiles []model.Profile
	err = db.Raw(runQueryString, params...).Scan(&profiles).Error
	if err != nil {
		log.WithError(err).WithFields(logFields).WithField("status", err).Error("Failed to get profile users.")
		return nil, http.StatusInternalServerError, "Query Execution Failed."
	}

	// Get merged properties for all accounts
	if profileType == model.PROFILE_TYPE_ACCOUNT && isDomainGroup {
		tablePropsHasUserProp := tablePropsHasUserProperty(tableProps)
		if isAllUserProperties {
			userParams := []interface{}{projectID, FormatTimeToString(minMax.MinUpdatedAt), FormatTimeToString(minMax.MaxUpdatedAt),
				projectID, model.UserSourceDomains}
			userParams = append(userParams, filterParams...)
			userDomains, _ := store.GetUsersAssociatedToDomain(projectID, timeAndRecordsLimit, filterString, userParams)
			profiles = appendProfiles(profiles, userDomains)
		}
		profiles, status = store.AccountPropertiesForDomainsEnabled(projectID, profiles, tablePropsHasUserProp)
		if status != http.StatusOK {
			return nil, status, "Query Transformation Failed."
		}
	}
	// Get Table Content
	returnData, err := FormatProfilesStruct(profiles, profileType, tableProps, payload.Source)
	if err != nil {
		log.WithError(err).WithFields(logFields).WithField("status", err).Error("Failed to filter properties from profiles.")
		return nil, http.StatusInternalServerError, "Query formatting failed."
	}
	return returnData, http.StatusFound, ""
}

// hasUserProperty checks for user properties in filters.
func hasUserProperty(profileType string, properties []model.QueryProperty) bool {
	isUserProperty := false

	if profileType == model.PROFILE_TYPE_ACCOUNT {
		for _, filter := range properties {
			if filter.Entity == model.PropertyEntityUserGroup {
				isUserProperty = true
				break
			}
		}
	}

	return isUserProperty
}

// hasAllUserProperties checks for all user properties in filters.
func hasAllUserProperties(filters []model.QueryProperty, profileType string) bool {
	isAllUserProperties := true

	if profileType == model.PROFILE_TYPE_ACCOUNT {
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

func (store *MemSQL) userPropertiesForAccounts(projectID int64, source string) (string, interface{}, string) {
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

func (store *MemSQL) GetUsersAssociatedToDomain(projectID int64, timeAndRecordsLimit string, filterString string, userParams []interface{}) ([]model.Profile, int) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	groupIdsMap, status := store.GetGroupNameIDMap(projectID)
	if status != http.StatusFound {
		logCtx.Error("Domain group not found.")
		return nil, status
	}

	var selectColArr []string
	for groupName, id := range groupIdsMap {
		if model.IsAllowedAccountGroupNames(groupName) {
			selectColArr = append(selectColArr, fmt.Sprintf("group_%d_user_id IS NULL", id))
		}
	}

	var userProfiles []model.Profile

	if filterString == "" || timeAndRecordsLimit == "" {
		return userProfiles, http.StatusOK
	}

	colStr := strings.Join(selectColArr, " AND ")
	if _, ok := groupIdsMap[model.GROUP_NAME_DOMAINS]; !ok {
		return userProfiles, http.StatusBadRequest
	}
	domainID := groupIdsMap[model.GROUP_NAME_DOMAINS]
	query := fmt.Sprintf(`SELECT domain_groups.id AS identity, user_global_user_properties as properties, 
		domain_groups.updated_at AS last_activity, domain_groups.group_%d_id as host_name FROM (
		SELECT properties as user_global_user_properties, group_%d_user_id, id FROM users 
		WHERE project_id = ? AND customer_user_id IS NOT NULL AND group_%d_user_id IS NOT NULL 
		AND %s AND %s) AS users JOIN (SELECT updated_at, id, group_%d_id FROM 
		users WHERE project_id = ? AND source = ? AND is_group_user = 1) AS 
		domain_groups ON users.group_%d_user_id = domain_groups.id WHERE %s 
		GROUP BY identity ORDER BY last_activity DESC LIMIT 1000;`, domainID, domainID, domainID, colStr, timeAndRecordsLimit, domainID, domainID, filterString)

	db := C.GetServices().Db
	err := db.Raw(query, userParams...).Scan(&userProfiles).Error
	if err != nil {
		log.WithError(err).WithFields(logFields).WithField("status", err).Error("Failed to get profile users.")
		return nil, http.StatusInternalServerError
	}

	return userProfiles, http.StatusOK
}

func (store *MemSQL) AccountPropertiesForDomainsEnabled(projectID int64, profiles []model.Profile, hasUserProp bool) ([]model.Profile, int) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	if len(profiles) < 1 {
		logCtx.Error("No domain account found.")
		return nil, http.StatusOK
	}

	domainGroup, status := store.GetGroup(projectID, model.GROUP_NAME_DOMAINS)
	if status != http.StatusFound {
		logCtx.Error("Domain group not found.")
		return nil, status
	}

	domainIDs := make([]string, len(profiles))
	for i, profile := range profiles {
		domainIDs[i] = profile.Identity
	}

	isGroupUserString := "is_group_user=1 AND"
	if hasUserProp {
		isGroupUserString = ""
	}

	// Fetching accounts associated to the domain
	// SELECT group_6_user_id as identity, properties FROM `users`  WHERE (project_id='15000001' AND source!='9' AND
	// is_group_user=1 AND group_6_user_id IN ('4f88f40d-c571-4bee-b456-298c533d7ef9', 'ed68f40d-c571-4bee-b456-298c533d7ef9'));
	var accountGroupDetails []model.Profile
	db := C.GetServices().Db
	err := db.Table("users").Select(fmt.Sprintf("group_%d_user_id as identity, properties", domainGroup.ID)).
		Where(fmt.Sprintf("project_id=? AND source!=? AND %s group_%d_user_id", isGroupUserString, domainGroup.ID)+" IN (?)",
			projectID, model.UserSourceDomains, domainIDs).Find(&accountGroupDetails).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to get accounts associated to domains.")
		return nil, http.StatusInternalServerError
	}

	// map of domain ids and their decoded merged properties
	domainsIDPropsMap := make(map[string]map[string]interface{})
	for _, accountDetails := range accountGroupDetails {
		propertiesDecoded, err := U.DecodePostgresJsonb(accountDetails.Properties)
		if err != nil {
			log.Error("Unable to decode account properties.")
			return nil, http.StatusInternalServerError
		}
		if _, exists := domainsIDPropsMap[accountDetails.Identity]; !exists {
			domainsIDPropsMap[accountDetails.Identity] = (*propertiesDecoded)
		} else {
			domainsIDPropsMap[accountDetails.Identity] = U.MergeJSONMaps(domainsIDPropsMap[accountDetails.Identity], *propertiesDecoded)
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
func (store *MemSQL) GetSourceStringForAccountsV2(projectID int64, source string, isAllUserProperties bool) (string, int, []interface{}, int) {
	var sourceString string
	var domainGroupId int
	var params []interface{}
	params = append(params, projectID)
	if source == "All" {
		group, errCode := store.GetGroup(projectID, model.GROUP_NAME_DOMAINS)
		if errCode != http.StatusFound || group == nil {
			log.WithField("err_code", errCode).Error("Failed to get domain group while adding group info.")
			return sourceString, domainGroupId, nil, http.StatusBadRequest
		}
		domainGroupId = group.ID
		sourceString = "AND users.source!=?"
		if isAllUserProperties {
			sourceString = sourceString + " " + fmt.Sprintf("AND (users.group_%d_id IS NOT NULL OR (users.customer_user_id IS NOT NULL AND users.group_%d_user_id IS NOT NULL))", domainGroupId, domainGroupId)
		} else {
			sourceString = sourceString + " " + fmt.Sprintf("AND users.group_%d_id IS NOT NULL", domainGroupId)
		}
		params = append(params, model.UserSourceDomains)
	} else {
		group, errCode := store.GetGroup(projectID, source)
		if errCode != http.StatusFound || group == nil {
			log.WithField("err_code", errCode).Error("Failed to get group while adding group info.")
			return sourceString, domainGroupId, nil, http.StatusBadRequest
		}
		if model.IsAllowedAccountGroupNames(source) && source == group.Name {
			sourceString = fmt.Sprintf("AND users.source!=? AND users.group_%d_id IS NOT NULL", group.ID)
			params = append(params, model.UserSourceDomains)
		} else {
			log.WithField("err_code", errCode).Error(fmt.Sprintf("%s not enabled for this project.", source))
			return sourceString, domainGroupId, nil, http.StatusBadRequest
		}
	}
	return sourceString, domainGroupId, params, http.StatusOK
}

func SelectFilterAndHavingStringsForAccounts(filtersMap map[string][]model.QueryProperty) (string, string) {
	index := 1
	filterArray := make([]string, 0)
	havingArray := make([]string, 0)
	propMap := make(map[string]bool)
	for group, filterArr := range filtersMap {
		if group == model.FILTER_TYPE_USERS {
			continue
		}
		for _, filter := range filterArr {
			if exists := propMap[filter.Property]; exists {
				continue
			}
			filterStr := fmt.Sprintf("MAX(JSON_EXTRACT_STRING(properties, '%s')) as filter_key_%d", filter.Property, index)
			filterArray = append(filterArray, filterStr)
			havingArray = append(havingArray, fmt.Sprintf("filter_key_%d IS NOT NULL", index))
			index += 1
			propMap[filter.Property] = true
		}
	}
	var selectFilterString, havingString string
	if len(filterArray) > 0 {
		selectFilterString = strings.Join(filterArray, ", ")
		havingString = "HAVING " + strings.Join(havingArray, " AND ")
	}
	return selectFilterString, havingString
}

// SELECT domain_groups.id as identity, users.properties as properties, domain_groups.updated_at as last_activity FROM (
// SELECT properties, group_6_user_id FROM users WHERE project_id=2 AND source != 9 AND group_6_user_id IS NOT NULL
// AND updated_at BETWEEN '2023-03-07 14:38:54.494786' AND '2023-04-07 14:38:54.494786' LIMIT 1000000) AS users JOIN (
// SELECT id, updated_at FROM users WHERE project_id = 2 AND source = 9 AND is_group_user = 1 AND group_6_id IS NOT NULL
// ) AS domain_groups ON users.group_6_user_id = domain_groups.id WHERE JSON_EXTRACT_STRING(users.properties, "$6signal_city") = "Delhi"
// GROUP BY identity ORDER BY last_activity DESC LIMIT 1000;
func BuildQueryStringForDomains(filterString string, whereForUserQuery string, domainGroupId int,
	userTimeAndRecordsLimit string, filters map[string][]model.QueryProperty, userPropsJoin string) string {
	whereForDomainGroupQuery := fmt.Sprintf(strings.Replace(whereForUserQuery, "users.source!=", "source=",
		1) + " AND is_group_user = 1")
	whereForUserQuery = whereForUserQuery + " AND " + userTimeAndRecordsLimit
	selectUserColumnsString := fmt.Sprintf("properties, updated_at, group_%d_user_id, id, customer_user_id, is_group_user, group_%d_id", domainGroupId, domainGroupId)
	userQueryString := fmt.Sprintf("(SELECT " + selectUserColumnsString + " FROM users " + whereForUserQuery + " ) AS users")
	selectDomainGroupColString := fmt.Sprintf("SELECT id, group_%d_id FROM users", domainGroupId)
	domainGroupQueryString := "( " + selectDomainGroupColString + " " + whereForDomainGroupQuery +
		" ) AS domain_groups"
	onCondition := fmt.Sprintf("ON users.group_%d_user_id = domain_groups.id", domainGroupId)
	groupByStr := "GROUP BY identity"
	selectString := fmt.Sprintf("domain_groups.id AS identity, users.properties as properties, MAX(users.updated_at) AS last_activity, domain_groups.group_%d_id as host_name", domainGroupId)

	var selectFilterString, havingString string
	selectFilterString, havingString = SelectFilterAndHavingStringsForAccounts(filters)
	if selectFilterString != "" {
		selectString = selectString + ", " + selectFilterString
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
	return queryString
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
			if group.Name == model.GROUP_NAME_DOMAINS || model.IsAllowedAccountGroupNames(group.Name) {
				groupNameIDMap[group.Name] = group.ID
			}
		}
	}
	return groupNameIDMap, http.StatusFound
}

func GetSourceStringForAccountsV1(groupNameIDMap map[string]int, source string) (string, int) {
	logFields := log.Fields{
		"group_id_map": groupNameIDMap,
		"source":       source,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	var sourceString string

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
		if source == "All" || source == crmName {
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

func FormatProfilesStruct(profiles []model.Profile, profileType string, tableProps []string, source string) ([]model.Profile, error) {
	logFields := log.Fields{
		"profile_type": profileType,
	}

	if profileType == model.PROFILE_TYPE_ACCOUNT {
		var companyNameProps, hostNameProps []string
		if model.IsAllowedAccountGroupNames(source) {
			hostNameProps = []string{model.HostNameGroup[source]}
			companyNameProps = []string{model.AccountNames[source], U.UP_COMPANY}
		} else {
			companyNameProps = model.NameProps
			hostNameProps = model.HostNameProps
		}

		for index, profile := range profiles {
			filterTableProps := make(map[string]interface{}, 0)
			properties, err := U.DecodePostgresJsonb(profile.Properties)
			if err != nil {
				log.WithFields(logFields).WithFields(log.Fields{"identity": profile.Identity}).WithError(err).Error("Failed decoding account properties.")
				continue
			}

			// Filter Table Props
			for _, prop := range tableProps {
				if value, exists := (*properties)[prop]; exists {
					filterTableProps[prop] = value
				}
			}
			profiles[index].TableProps = filterTableProps

			// Filter Company Name and Hostname
			nameProps := append(companyNameProps, hostNameProps...)

			for _, prop := range nameProps {
				if profiles[index].Name == "" {
					if name, exists := (*properties)[prop]; exists {
						profiles[index].Name = fmt.Sprintf("%s", name)
					}
				}
			}

			for _, prop := range hostNameProps {
				if profiles[index].HostName == "" {
					if hostname, exists := (*properties)[prop]; exists {
						profiles[index].HostName = fmt.Sprintf("%s", hostname)
					}
				}
			}

			if profiles[index].Name == "" && profiles[index].HostName != "" {
				profiles[index].Name = profiles[index].HostName
			}
		}
	} else if profileType == model.PROFILE_TYPE_USER {
		for index, profile := range profiles {
			filterTableProps := make(map[string]interface{}, 0)
			properties, err := U.DecodePostgresJsonb(profile.Properties)
			if err != nil {
				log.WithFields(logFields).WithFields(log.Fields{"identity": profile.Identity}).WithError(err).Error("Failed decoding account properties.")
				continue
			}

			// Filter Table Props
			for _, prop := range tableProps {
				if value, exists := (*properties)[prop]; exists {
					filterTableProps[prop] = value
				}
			}
			profiles[index].TableProps = filterTableProps
		}
	}
	return profiles, nil
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
		Order("updated_at desc").
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
		if company, exists := (*propertiesDecoded)[U.UP_COMPANY]; exists {
			uniqueUser.Company = fmt.Sprintf("%s", company)
		}
		timelinesConfig, err := store.GetTimelineConfigOfProject(projectID)
		if err != nil {
			log.WithField("status", err).WithError(err).Error("Failed to fetch timelines_config from project_settings.")
		}
		uniqueUser.LeftPaneProps = GetLeftPanePropertiesFromConfig(timelinesConfig, model.PROFILE_TYPE_USER, propertiesDecoded)
		uniqueUser.Milestones = GetMilestonesFromConfig(timelinesConfig, model.PROFILE_TYPE_USER, propertiesDecoded)
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

	eventNamesToExclude := []string{
		U.EVENT_NAME_HUBSPOT_CONTACT_UPDATED,
		U.EVENT_NAME_SALESFORCE_CONTACT_UPDATED,
		U.EVENT_NAME_SALESFORCE_LEAD_UPDATED,
		U.EVENT_NAME_LEAD_SQUARED_LEAD_UPDATED,
		U.EVENT_NAME_MARKETO_LEAD_UPDATED,
		U.EVENT_NAME_SALESFORCE_ACCOUNT_UPDATED,
		U.EVENT_NAME_SALESFORCE_OPPORTUNITY_UPDATED,
		U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_UPDATED,
		U.EVENT_NAME_SALESFORCE_TASK_UPDATED,
		U.EVENT_NAME_SALESFORCE_EVENT_UPDATED,
		U.EVENT_NAME_HUBSPOT_ENGAGEMENT_MEETING_UPDATED,
		U.EVENT_NAME_HUBSPOT_ENGAGEMENT_CALL_UPDATED,
		U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_UPDATED,
		U.GROUP_EVENT_NAME_HUBSPOT_DEAL_UPDATED,
		U.GROUP_EVENT_NAME_SALESFORCE_ACCOUNT_UPDATED,
		U.GROUP_EVENT_NAME_SALESFORCE_OPPORTUNITY_UPDATED,
		U.GROUP_EVENT_NAME_G2_ALL,
	}

	eventNamesToExcludePlaceholders := strings.Repeat("?,", len(eventNamesToExclude)-1) + "?"
	eventsQuery := fmt.Sprintf(`SELECT event_names.name AS event_name, 
        event_names.type as event_type, 
        events1.timestamp AS timestamp, 
        events1.properties AS properties 
    FROM (
        SELECT project_id, event_name_id, timestamp, properties 
        FROM events 
        WHERE project_id=? AND timestamp <= ? 
        AND user_id IN (
            SELECT id FROM users WHERE project_id=? AND %s = ?
        ) AND event_name_id NOT IN (
            SELECT id FROM event_names WHERE project_id=? AND name IN (%s)
        ) 
        LIMIT 5000) AS events1 
    LEFT JOIN event_names
    ON events1.event_name_id=event_names.id 
    AND event_names.project_id=?;`, userId, eventNamesToExcludePlaceholders)

	excludedEventNamesArgs := make([]interface{}, len(eventNamesToExclude))
	for i, name := range eventNamesToExclude {
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

	// User Activity
	standardDisplayNames := U.STANDARD_EVENTS_DISPLAY_NAMES
	errCode, projectDisplayNames := store.GetDisplayNamesForAllEvents(projectID)
	if errCode != http.StatusFound {
		log.WithError(err).WithField("project_id", projectID).Error("Error fetching display names for the project")
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
			// Virtual Events Case: Replace event_name with $page_url
			if userActivity.EventType == model.TYPE_FILTER_EVENT_NAME {
				if pageURL, exists := (*properties)[U.EP_PAGE_URL]; exists {
					userActivity.AliasName = fmt.Sprintf("%s", pageURL)
				}
			}
			// Display Names
			if (*properties)[U.EP_IS_PAGE_VIEW] == true {
				userActivity.DisplayName = "Page View"
				// Page View Icon
				userActivity.Icon = "window"
			} else if standardDisplayNames[userActivity.EventName] != "" {
				userActivity.DisplayName = standardDisplayNames[userActivity.EventName]
			} else if projectDisplayNames[userActivity.EventName] != "" {
				userActivity.DisplayName = projectDisplayNames[userActivity.EventName]
			} else {
				userActivity.DisplayName = userActivity.EventName
			}
			// Alias Names
			if userActivity.EventName == U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_CREATED {
				userActivity.AliasName = fmt.Sprintf("Added to %s", (*properties)[U.EP_SALESFORCE_CAMPAIGN_NAME])
			} else if userActivity.EventName == U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_RESPONDED_TO_CAMPAIGN {
				userActivity.AliasName = fmt.Sprintf("Responded to %s", (*properties)[U.EP_SALESFORCE_CAMPAIGN_NAME])
			} else if userActivity.EventName == U.EVENT_NAME_HUBSPOT_CONTACT_FORM_SUBMISSION {
				userActivity.AliasName = fmt.Sprintf("%s", (*properties)[U.EP_HUBSPOT_FORM_SUBMISSION_TITLE])
			} else if userActivity.EventName == U.EVENT_NAME_HUBSPOT_ENGAGEMENT_EMAIL {
				emailSubject := "No Subject"
				if subject, exists := (*properties)[U.EP_HUBSPOT_ENGAGEMENT_SUBJECT]; exists {
					if !(subject == nil || subject == "") {
						emailSubject = fmt.Sprintf("%s", subject)
					}
				}
				userActivity.AliasName = fmt.Sprintf("%s: %s", (*properties)[U.EP_HUBSPOT_ENGAGEMENT_TYPE], emailSubject)
			} else if userActivity.EventName == U.EVENT_NAME_HUBSPOT_ENGAGEMENT_MEETING_CREATED ||
				userActivity.EventName == U.EVENT_NAME_HUBSPOT_ENGAGEMENT_CALL_CREATED {
				userActivity.AliasName = fmt.Sprintf("%s", (*properties)[U.EP_HUBSPOT_ENGAGEMENT_TITLE])
			} else if userActivity.EventName == U.EVENT_NAME_SALESFORCE_TASK_CREATED {
				userActivity.AliasName = fmt.Sprintf("Created Task - %s", (*properties)[U.EP_SF_TASK_SUBJECT])
			} else if userActivity.EventName == U.EVENT_NAME_SALESFORCE_EVENT_CREATED {
				userActivity.AliasName = fmt.Sprintf("Created Event - %s", (*properties)[U.EP_SF_EVENT_SUBJECT])
			} else if userActivity.EventName == U.EVENT_NAME_HUBSPOT_CONTACT_LIST {
				userActivity.AliasName = fmt.Sprintf("Added to Hubspot List - %s", (*properties)[U.EP_HUBSPOT_CONTACT_LIST_LIST_NAME])
			}
			// Set Icons
			if icon, exists := model.EVENT_ICONS_MAP[userActivity.EventName]; exists {
				userActivity.Icon = icon
			} else if strings.Contains(userActivity.EventName, "hubspot_") || strings.Contains(userActivity.EventName, "hs_") {
				userActivity.Icon = "hubspot"
			} else if strings.Contains(userActivity.EventName, "salesforce_") || strings.Contains(userActivity.EventName, "sf_") {
				userActivity.Icon = "salesforce"
			} else if strings.Contains(userActivity.EventName, "linkedin_") || strings.Contains(userActivity.EventName, "li_") {
				userActivity.Icon = "linkedin"
			}
			// Default Icon
			if userActivity.Icon == "" {
				userActivity.Icon = "calendar-star"
			}

			// Filtered Properties
			userActivity.Properties = GetFilteredProperties(userActivity.EventName, userActivity.EventType, properties)
		}
		userActivities = append(userActivities, userActivity)
	}

	return userActivities, nil
}

func GetFilteredProperties(eventName string, eventType string, properties *map[string]interface{}) *postgres.Jsonb {
	var returnProperties *postgres.Jsonb
	filteredProperties := make(map[string]interface{})
	filterProps, eventExistsInMap := model.HOVER_EVENTS_NAME_PROPERTY_MAP[eventName]
	if (*properties)[U.EP_IS_PAGE_VIEW] == true {
		for _, prop := range model.PAGE_VIEW_HOVERPROPS_LIST {
			if value, propExists := (*properties)[prop]; propExists {
				filteredProperties[prop] = value
			}
		}
	} else if eventExistsInMap {
		for _, prop := range filterProps {
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
	isUserDetails := false
	var accountDetails model.AccountDetails

	var group *model.Group
	if groupName == "All" {
		group, status = store.GetGroup(projectID, U.GROUP_NAME_DOMAINS)
	} else {
		group, status = store.GetGroup(projectID, groupName)
	}
	if status != http.StatusFound || group == nil {
		return nil, status, "Failed to get group"
	}
	groupUserString := fmt.Sprintf("group_%d_user_id=? ", group.ID)
	params := []interface{}{projectID, id}

	if C.IsDomainEnabled(projectID) {
		propertiesDecoded, isUserDetails, status = store.AccountPropertiesForDomainsEnabledV2(projectID, id, groupName)
	} else {
		groupUserString, propertiesDecoded, params, status = store.AccountPropertiesForDomainsDisabledV1(projectID, id)
	}

	if isUserDetails {
		accountDetails, propertiesDecoded, status = store.GetUserDetailsAssociatedToDomain(projectID, id)
	}

	if status != http.StatusOK {
		return nil, status, "Accounts Query Processing Failed"
	}

	accountDetails = FormatAccountDetails(projectID, propertiesDecoded, groupName, accountDetails.HostName)

	timelinesConfig, err := store.GetTimelineConfigOfProject(projectID)
	if err != nil {
		log.WithError(err).Error("Failed to fetch timelines_config from project_settings.")
	}

	accountDetails.LeftPaneProps = GetLeftPanePropertiesFromConfig(timelinesConfig, model.PROFILE_TYPE_ACCOUNT, &propertiesDecoded)
	accountDetails.Milestones = GetMilestonesFromConfig(timelinesConfig, model.PROFILE_TYPE_ACCOUNT, &propertiesDecoded)

	additionalProp := timelinesConfig.AccountConfig.UserProp
	selectStrAdditionalProp := ""
	if additionalProp != "" {
		selectStrAdditionalProp = fmt.Sprintf("JSON_EXTRACT_STRING(properties, '%s') as additional_prop,", additionalProp)
	}

	// Timeline Query
	query := fmt.Sprintf(`SELECT COALESCE(JSON_EXTRACT_STRING(properties, '%s'), customer_user_id, id) AS user_name, %s
        COALESCE(customer_user_id, id) AS user_id, 
        ISNULL(customer_user_id) AS is_anonymous 
    FROM users 
    WHERE project_id = ?
	  AND (is_group_user = 0 OR is_group_user IS NULL)
	  AND (%s)
    GROUP BY user_id 
    ORDER BY updated_at DESC 
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

	if C.IsScoringEnabled(projectID) {
		overview, err := store.GetAccountOverview(projectID, id, groupName)
		if err != nil {
			log.WithFields(logFields).WithError(err)
		} else {
			accountDetails.Overview = overview
		}
	}

	return &accountDetails, http.StatusFound, ""
}

func (store *MemSQL) GetAccountOverview(projectID int64, id, groupName string) (model.Overview, error) {
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
	if groupName == "All" {
		grpName = U.GROUP_NAME_DOMAINS
	}
	group, errGetGroup := store.GetGroup(projectID, grpName)
	if group != nil {
		groupUserString = fmt.Sprintf("group_%d_user_id=? ", group.ID)
		params = append(params, id)
	}

	var errGetCount error
	if errGetGroup != http.StatusFound {
		errGetCount = fmt.Errorf("error retrieving parameters")
		log.WithFields(logFields).Error("Error retrieving parameters")
	} else {
		overviewQuery := fmt.Sprintf(`
		SELECT COUNT(DISTINCT(id)) AS users_count, 
			SUM(JSON_EXTRACT_STRING(properties, '%s')) AS time_active 
		FROM (
			SELECT LAST(id, updated_at) AS id, properties 
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

		db := C.GetServices().Db
		errGetCount = db.Raw(overviewQuery, queryParams...).Scan(&overview).Error
		if errGetCount != nil {
			log.WithFields(logFields).WithError(errGetCount).Error("Error retrieving users count and active time")
		}
	}

	// Get Account Engagement Score and Trends
	accountScore, _, errGetScore := store.GetPerAccountScore(projectID, time.Now().Format("20060102"), id, model.NUM_TREND_DAYS, true)
	if errGetScore != nil {
		log.WithFields(logFields).WithError(errGetScore).Error("Error retrieving account score")
	} else {
		overview.Temperature = accountScore.Score
		overview.ScoresList = accountScore.Trend
	}

	if errGetScore != nil && errGetCount != nil {
		return overview, fmt.Errorf("error getting overview")
	}
	return overview, nil
}

func (store *MemSQL) GetIntentTimeline(projectID int64, groupName string, id string) (model.UserTimeline, error) {
	intentTimeline := model.UserTimeline{
		UserId:         id,
		IsAnonymous:    false,
		UserName:       model.GROUP_ACTIVITY_USERNAME,
		UserActivities: []model.UserActivity{},
	}

	if groupName == "All" || groupName == U.GROUP_NAME_DOMAINS {
		intentTimeline.AdditionalProp = "All"
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
			intentTimeline.AdditionalProp = groupDisplayName
		} else {
			intentTimeline.AdditionalProp = groupName
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

func FormatTimeToString(time time.Time) string {
	return time.Format("2006-01-02 15:04:05.000000")
}

func (store *MemSQL) AccountPropertiesForDomainsEnabledV2(projectID int64, id string, groupName string) (map[string]interface{},
	bool, int) {
	propertiesDecoded := make(map[string]interface{}, 0)
	isUserDetails := false
	if groupName == "All" {
		groupNameIDMap, status := store.GetGroupNameIDMap(projectID)
		if status != http.StatusFound {
			return propertiesDecoded, isUserDetails, status
		}
		// Fetching accounts associated to the domain
		accountGroupDetails, status := store.GetAccountsAssociatedToDomain(projectID, id, groupNameIDMap[model.GROUP_NAME_DOMAINS])
		if status != http.StatusFound {
			return propertiesDecoded, isUserDetails, status
		}

		if len(accountGroupDetails) < 1 {
			isUserDetails = true
			return propertiesDecoded, isUserDetails, status
		}

		for index, accountGroupDetail := range accountGroupDetails {
			props, err := U.DecodePostgresJsonb(&accountGroupDetail.Properties)
			if err != nil {
				log.Error("Unable to decode account properties.")
				return propertiesDecoded, isUserDetails, status
			}
			// merging all account properties
			if index == 0 {
				propertiesDecoded = *props
			} else {
				propertiesDecoded = U.MergeJSONMaps(propertiesDecoded, *props)
			}
		}
	} else {
		if !model.IsAllowedAccountGroupNames(groupName) {
			log.Error("Invalid group name.")
			return propertiesDecoded, isUserDetails, http.StatusBadRequest
		}
		// Filter Properties
		properties, status := store.GetUserPropertiesByUserID(projectID, id)
		if status != http.StatusFound {
			log.Error("Failed to get account properties.")
			return propertiesDecoded, isUserDetails, http.StatusInternalServerError
		}
		props, err := U.DecodePostgresJsonb(properties)
		if err != nil {
			log.WithError(err).Error("Failed to decode account properties.")
			return propertiesDecoded, isUserDetails, http.StatusInternalServerError
		}
		propertiesDecoded = *props
	}
	return propertiesDecoded, isUserDetails, http.StatusOK
}

func (store *MemSQL) AccountPropertiesForDomainsDisabledV1(projectID int64, id string) (string,
	map[string]interface{}, []interface{}, int) {
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
	// SELECT domain_group.group_1_id AS host_name, users.properties AS properties FROM (SELECT group_1_id, id FROM users WHERE project_id = 1000000 AND source = 9
	// AND id = "aa5f9e4e-e516-481f-86cd-bd42debda12c") AS domain_group JOIN (SELECT group_1_user_id, properties FROM users WHERE project_id = 1000000 AND (is_group_user = 0 OR
	// is_group_user IS NULL) AND group_1_user_id IS NOT NULL) AS users ON domain_group.id = users.group_1_user_id LIMIT 1;
	db := C.GetServices().Db
	query := fmt.Sprintf("SELECT domain_group.group_%d_id AS host_name, users.properties AS properties FROM "+
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
		Where("project_id=? AND source!=? AND is_group_user=1 AND "+fmt.Sprintf("group_%d_user_id", domainGroupId)+"=?", projectID, model.UserSourceDomains, id).Find(&accountGroupDetails).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to get groups.")
		return nil, http.StatusInternalServerError
	}
	return accountGroupDetails, http.StatusFound
}

func FormatAccountDetails(projectID int64, propertiesDecoded map[string]interface{},
	groupName string, hostName string) model.AccountDetails {
	var companyNameProps, hostNameProps []string
	var accountDetails model.AccountDetails

	if C.IsDomainEnabled(projectID) && groupName != "All" {
		if model.IsAllowedAccountGroupNames(groupName) {
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
		accountDetails.HostName = hostName
	} else {
		for _, prop := range hostNameProps {
			if host, exists := (propertiesDecoded)[prop]; exists {
				accountDetails.HostName = fmt.Sprintf("%s", host)
				break
			}
		}
	}

	if accountDetails.Name == "" && accountDetails.HostName != "" {
		accountDetails.Name = accountDetails.HostName
	}
	return accountDetails
}

func GetLeftPanePropertiesFromConfig(timelinesConfig model.TimelinesConfig, profileType string, propertiesDecoded *map[string]interface{}) map[string]interface{} {

	filteredProperties := make(map[string]interface{})
	var leftPaneProps []string

	if profileType == model.PROFILE_TYPE_USER {
		leftPaneProps = timelinesConfig.UserConfig.LeftpaneProps
	} else if profileType == model.PROFILE_TYPE_ACCOUNT {
		leftPaneProps = timelinesConfig.AccountConfig.LeftpaneProps
	}
	for _, prop := range leftPaneProps {
		if value, exists := (*propertiesDecoded)[prop]; exists {
			filteredProperties[prop] = value
		}
	}
	return filteredProperties
}

func GetMilestonesFromConfig(timelinesConfig model.TimelinesConfig, profileType string, propertiesDecoded *map[string]interface{}) map[string]interface{} {

	filteredProperties := make(map[string]interface{})
	var milestones []string

	if profileType == model.PROFILE_TYPE_USER {
		milestones = timelinesConfig.UserConfig.Milestones
	} else if profileType == model.PROFILE_TYPE_ACCOUNT {
		milestones = timelinesConfig.AccountConfig.Milestones
	}
	for _, prop := range milestones {
		if value, exists := (*propertiesDecoded)[prop]; exists {
			filteredProperties[prop] = value
		}
	}
	return filteredProperties
}

func (store *MemSQL) GetAnalyzeResultForSegments(projectId int64, segment *model.Segment) ([]model.Profile, int, error) {
	logFields := log.Fields{
		"project_id": projectId,
		"name":       segment.Name,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	if projectId == 0 || segment.Name == "" {
		logCtx.Error("Segment event query Failed. Invalid parameters.")
		return nil, http.StatusBadRequest, errors.New("segment event query failed. invalid parameters")
	}

	segmentQuery := &model.Query{}
	err := U.DecodePostgresJsonbToStructType(segment.Query, segmentQuery)
	if err != nil {
		logCtx.Error("failed to decode json. aborting")
		return nil, http.StatusBadRequest, errors.New("segment event query failed. invalid parameters")
	}

	result, errCode, errMsg := store.Analyze(projectId, *segmentQuery, C.EnableOptimisedFilterOnEventUserQuery(), false)
	if errCode != http.StatusOK {
		logCtx.WithField("err_code", errCode).Error("Failed at building query. ", errMsg)
		return nil, errCode, nil
	}

	profiles := make([]model.Profile, 0)
	if segmentQuery.Caller == model.USER_PROFILE_CALLER {
		for _, profile := range result.Rows {
			var row model.Profile
			row.Identity = profile[0].(string)
			var val bool
			if profile[1] == float64(1) {
				val = true
			} else {
				val = false
			}
			row.IsAnonymous = val
			row.LastActivity = profile[2].(time.Time)
			reflectProps := reflect.ValueOf(profile[3])
			props := make(map[string]interface{}, 0)
			if err := json.Unmarshal([]byte(reflectProps.String()), &props); err != nil {
				logCtx.WithError(err).Error("Failed at unmarshalling props. ")
				return nil, http.StatusInternalServerError, nil
			}
			row.Properties, err = U.EncodeToPostgresJsonb(&props)
			if err != nil {
				logCtx.Error("Failed at encoding props. ")
				return nil, http.StatusInternalServerError, nil
			}
			profiles = append(profiles, row)
		}
	} else if segmentQuery.Caller == model.ACCOUNT_PROFILE_CALLER {
		for _, profile := range result.Rows {
			var row model.Profile
			row.TableProps = make(map[string]interface{}, 0)
			row.Identity = profile[0].(string)
			row.LastActivity = profile[1].(time.Time)
			reflectProps := reflect.ValueOf(profile[2])
			props := make(map[string]interface{}, 0)
			if err := json.Unmarshal([]byte(reflectProps.String()), &props); err != nil {
				logCtx.WithError(err).Error("Failed at unmarshalling props.")
				return nil, http.StatusInternalServerError, nil
			}
			row.Properties, err = U.EncodeToPostgresJsonb(&props)
			if err != nil {
				logCtx.Error("Failed at encoding props. ")
				return nil, http.StatusInternalServerError, nil
			}
			profiles = append(profiles, row)
		}
	}

	return profiles, http.StatusOK, nil
}

func GroupFiltersByPrefix(filters []model.QueryProperty) map[string][]model.QueryProperty {
	filtersMap := make(map[string][]model.QueryProperty)

	for _, filter := range filters {
		var groupName string
		for _, prefix := range model.GroupPropertyPrefixList {
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
