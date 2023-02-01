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

/*Sample Timeline Listing Queries:

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
func (store *MemSQL) GetProfilesListByProjectId(projectID int64, payload model.TimelinePayload, profileType string) ([]model.Profile, int) {
	logFields := log.Fields{
		"project_id":   projectID,
		"payload":      payload,
		"profile_type": profileType,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if projectID == 0 {
		return nil, http.StatusBadRequest
	}

	if payload.SegmentId != "" {
		segment, status := store.GetSegmentById(projectID, payload.SegmentId)
		if status != http.StatusFound {
			return nil, http.StatusBadRequest
		}
		payload.Source = segment.Type
		segmentQuery := &model.Query{}
		err := U.DecodePostgresJsonbToStructType(segment.Query, segmentQuery)
		if err != nil {
			return nil, http.StatusInternalServerError
		}
		if segmentQuery.EventsWithProperties != nil && len(segmentQuery.EventsWithProperties) > 0 {
			if C.IsEnabledEventsFilterInSegments() {
				profiles, errCode, _ := store.GetAnalyzeResultForSegments(projectID, segment)
				if errCode != http.StatusOK {
					return nil, errCode
				}
				return profiles, http.StatusFound
			} else {
				var profiles = make([]model.Profile, 0)
				return profiles, http.StatusBadRequest
			}
		} else {
			payload.Filters = append(payload.Filters, segmentQuery.GlobalUserProperties...)
		}
	}

	var selectString, isGroupUserString, sourceString string

	if profileType == model.PROFILE_TYPE_ACCOUNT {
		// Check for Enabled Groups
		groups, errCode := store.GetGroups(projectID)
		if errCode != http.StatusFound {
			log.WithField("status", errCode).Error("Failed to get groups while adding group info.")
		}
		groupNameIDMap := make(map[string]int)
		if len(groups) > 0 {
			for _, group := range groups {
				if group.Name == model.GROUP_NAME_HUBSPOT_COMPANY || group.Name == model.GROUP_NAME_SALESFORCE_ACCOUNT {
					groupNameIDMap[group.Name] = group.ID
				}
			}
		}
		hubspotID, hubspotExists := groupNameIDMap[model.GROUP_NAME_HUBSPOT_COMPANY]
		salesforceID, salesforceExists := groupNameIDMap[model.GROUP_NAME_SALESFORCE_ACCOUNT]

		if !hubspotExists && !salesforceExists {
			log.WithFields(logFields).Error("No CRMs Enabled for this project.")
			return nil, http.StatusBadRequest
		}
		if payload.Source == model.GROUP_NAME_HUBSPOT_COMPANY && !hubspotExists {
			log.WithFields(logFields).Error("Hubspot Not Enabled for this project.")
			return nil, http.StatusBadRequest
		}
		if payload.Source == model.GROUP_NAME_SALESFORCE_ACCOUNT && !salesforceExists {
			log.WithFields(logFields).Error("Salesforce Not Enabled for this project.")
			return nil, http.StatusBadRequest
		}
		selectString = "id AS identity, properties, updated_at AS last_activity"
		isGroupUserString = "is_group_user=1"
		if payload.Source == "All" && hubspotExists && salesforceExists {
			sourceString = fmt.Sprintf("AND (group_%d_id IS NOT NULL OR group_%d_id IS NOT NULL)", hubspotID, salesforceID)
		} else if (payload.Source == "All" || payload.Source == model.GROUP_NAME_HUBSPOT_COMPANY) && hubspotExists {
			sourceString = fmt.Sprintf("AND group_%d_id IS NOT NULL", hubspotID)
		} else if (payload.Source == "All" || payload.Source == model.GROUP_NAME_SALESFORCE_ACCOUNT) && salesforceExists {
			sourceString = fmt.Sprintf("AND group_%d_id IS NOT NULL", salesforceID)
		}
	} else if profileType == model.PROFILE_TYPE_USER {
		selectString = "COALESCE(customer_user_id, id) AS identity, ISNULL(customer_user_id) AS is_anonymous, properties, MAX(updated_at) AS last_activity"
		isGroupUserString = "(is_group_user=0 OR is_group_user IS NULL)"
		if model.UserSourceMap[payload.Source] == model.UserSourceWeb {
			sourceString = "AND (source=" + strconv.Itoa(model.UserSourceMap[payload.Source]) + " OR source IS NULL)"
		} else if payload.Source == "All" {
			sourceString = ""
		} else {
			sourceString = "AND source=" + strconv.Itoa(model.UserSourceMap[payload.Source])
		}
	}

	filterString, filterParams, errCode := buildWhereFromProperties(projectID, payload.Filters, 0)
	if filterString != "" {
		filterString = "(" + filterString + ")"
	}
	if errCode != nil {
		return nil, http.StatusBadRequest
	}

	// Run Queries
	type MinMaxTime struct {
		MinUpdatedAt time.Time `json:"min_updated_at"`
		MaxUpdatedAt time.Time `json:"max_updated_at"`
	}
	var minMax MinMaxTime
	var runQueryString, fromStr, groupByStr, selectColumnsStr, commonStr string
	windowSelectStr := "MIN(updated_at) AS min_updated_at, MAX(updated_at) AS max_updated_at"                  // Select Min & Max updated_at
	commonStr = fmt.Sprintf("users WHERE project_id=%d AND %s %s", projectID, isGroupUserString, sourceString) // Common String for Queries
	fromStr = fmt.Sprintf("%s AND updated_at < '%s'", commonStr, FormatTimeToString(gorm.NowFunc()))
	// Get min and max updated_at after ordering as part of optimisation.
	limitVal := 100000
	if filterString != "" {
		limitVal = 500000
	}
	runQueryString = fmt.Sprintf("SELECT %s FROM (SELECT updated_at FROM %s ORDER BY updated_at DESC LIMIT %d);", windowSelectStr, fromStr, limitVal)

	db := C.GetServices().Db
	err := db.Raw(runQueryString).Scan(&minMax).Error
	if err != nil {
		log.WithFields(logFields).WithField("status", err).Error("min and max updated_at couldn't be defined.")
		return nil, http.StatusInternalServerError
	}

	if profileType == model.PROFILE_TYPE_ACCOUNT {
		selectColumnsStr = "id, properties, updated_at"
		groupByStr = ""
	} else if profileType == model.PROFILE_TYPE_USER {
		selectColumnsStr = "id, customer_user_id, properties, updated_at"
		groupByStr = "GROUP BY identity"
	}
	if filterString != "" {
		fromStr = fmt.Sprintf("(SELECT %s FROM %s AND updated_at BETWEEN '%s' AND '%s'  LIMIT 1000000) AS select_view WHERE",
			selectColumnsStr, commonStr, FormatTimeToString(minMax.MinUpdatedAt), FormatTimeToString(minMax.MaxUpdatedAt))
		filterString = strings.ReplaceAll(filterString, "users.", "select_view.") // Json Filters on select_view
	} else {
		fromStr = fmt.Sprintf("%s AND updated_at BETWEEN '%s' AND '%s'",
			commonStr, FormatTimeToString(minMax.MinUpdatedAt), FormatTimeToString(minMax.MaxUpdatedAt))
	}
	runQueryString = fmt.Sprintf("SELECT %s FROM %s %s %s ORDER BY last_activity DESC LIMIT 1000;", selectString, fromStr, filterString, groupByStr)

	var profiles []model.Profile
	err = db.Raw(runQueryString, filterParams...).Scan(&profiles).Error
	if err != nil {
		log.WithFields(logFields).WithField("status", err).Error("Failed to get profile users.")
		return nil, http.StatusInternalServerError
	}

	// Get Table Content
	returnData, err := store.FormatProfilesStruct(projectID, profiles, profileType)
	if err != nil {
		log.WithFields(logFields).WithField("status", err).Error("Failed to filter properties from profiles.")
		return nil, http.StatusInternalServerError
	}
	return returnData, http.StatusFound
}

func (store *MemSQL) FormatProfilesStruct(projectID int64, profiles []model.Profile, profileType string) ([]model.Profile, error) {
	logFields := log.Fields{
		"project_id":   projectID,
		"profile_type": profileType,
	}
	timelinesConfig, err := store.GetTimelineConfigOfProject(projectID)
	if err != nil {
		log.WithFields(logFields).WithField("status", err).WithError(err).Error("Failed to fetch timelines_config from project_settings.")
	}

	if profileType == model.PROFILE_TYPE_ACCOUNT {
		companyNameProps := []string{U.UP_COMPANY, U.GP_HUBSPOT_COMPANY_NAME, U.GP_HUBSPOT_COMPANY_DOMAIN, U.GP_SALESFORCE_ACCOUNT_NAME}
		hostNameProps := []string{U.GP_HUBSPOT_COMPANY_DOMAIN, U.GP_SALESFORCE_ACCOUNT_WEBSITE}
		tableProps := timelinesConfig.AccountConfig.TableProps

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
			for _, prop := range companyNameProps {
				if profiles[index].Name != "" {
					break
				}
				if name, exists := (*properties)[prop]; exists {
					profiles[index].Name = fmt.Sprintf("%s", name)
				}
			}
			for _, prop := range hostNameProps {
				if profiles[index].HostName != "" {
					break
				}
				if hostname, exists := (*properties)[prop]; exists {
					profiles[index].HostName = fmt.Sprintf("%s", hostname)
				}
			}
		}
	} else if profileType == model.PROFILE_TYPE_USER {
		tableProps := timelinesConfig.UserConfig.TableProps
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

func (store *MemSQL) GetProfileUserDetailsByID(projectID int64, identity string, isAnonymous string) (*model.ContactDetails, int) {
	logFields := log.Fields{
		"project_id":   projectID,
		"id":           identity,
		"is_anonymous": isAnonymous,
	}

	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if projectID == 0 {
		log.Error("Invalid project ID.")
		return nil, http.StatusBadRequest
	}
	if identity == "" {
		log.Error("Invalid user ID.")
		return nil, http.StatusBadRequest
	}
	if isAnonymous == "" {
		log.Error("Invalid user type.")
		return nil, http.StatusBadRequest
	}

	userId := "customer_user_id"
	if isAnonymous == "true" {
		userId = "id"
	}

	db := C.GetServices().Db
	var uniqueUser model.ContactDetails
	err := db.Table("users").Select(`COALESCE(customer_user_id,id) AS user_id,
		ISNULL(customer_user_id) AS is_anonymous,
		properties,
		GROUP_CONCAT(group_1_id) IS NOT NULL AS group_1,
		GROUP_CONCAT(group_2_id) IS NOT NULL AS group_2,
		GROUP_CONCAT(group_3_id) IS NOT NULL AS group_3,
		GROUP_CONCAT(group_4_id) IS NOT NULL AS group_4`).
		Where("project_id=? AND "+userId+"=?", projectID, identity).
		Group("user_id").
		Order("updated_at desc").
		Limit(1).
		Find(&uniqueUser).Error
	if err != nil {
		log.WithFields(logFields).WithField("status", err).Error("Failed to get contact details.")
		return nil, http.StatusInternalServerError
	}

	propertiesDecoded, err := U.DecodePostgresJsonb(uniqueUser.Properties)
	if err != nil {
		log.WithFields(logFields).WithError(err).Error("Failed decoding user properties.")
	}

	if name, exists := (*propertiesDecoded)[U.UP_NAME]; exists {
		uniqueUser.Name = fmt.Sprintf("%s", name)
	}
	if company, exists := (*propertiesDecoded)[U.UP_COMPANY]; exists {
		uniqueUser.Company = fmt.Sprintf("%s", company)
	}

	uniqueUser.LeftPaneProps = store.GetLeftPaneProperties(projectID, model.PROFILE_TYPE_USER, propertiesDecoded)
	activities, _ := store.GetUserActivitiesAndSessionCount(projectID, identity, userId)
	uniqueUser.UserActivity = activities
	uniqueUser.GroupInfos = store.GetGroupsForUserTimeline(projectID, uniqueUser)

	return &uniqueUser, http.StatusFound
}

func (store *MemSQL) GetUserActivitiesAndSessionCount(projectID int64, identity string, userId string) ([]model.UserActivity, uint64) {
	logFields := log.Fields{
		"project_id": projectID,
		"id":         identity,
	}

	var userActivities []model.UserActivity
	webSessionCount := 0

	db := C.GetServices().Db
	eventsQuery := fmt.Sprintf(`SELECT event_names.name AS event_name, 
		event_names.type as event_type, 
		events1.timestamp AS timestamp, 
		events1.properties AS properties 
	FROM (
		SELECT project_id, event_name_id, timestamp, properties 
		FROM events 
		WHERE project_id=?
		AND timestamp <= ? 
		AND user_id IN (
			SELECT id FROM users WHERE project_id=? AND %s = ?
		) AND event_name_id NOT IN (
			SELECT id FROM event_names WHERE project_id=? AND name IN ('%s','%s','%s','%s','%s','%s','%s')
		) 
		LIMIT 5000) AS events1 
	LEFT JOIN event_names
	ON events1.event_name_id=event_names.id 
	AND event_names.project_id=?
	ORDER BY timestamp DESC;`, userId,
		U.EVENT_NAME_HUBSPOT_CONTACT_UPDATED,
		U.EVENT_NAME_SALESFORCE_CONTACT_UPDATED,
		U.EVENT_NAME_SALESFORCE_LEAD_UPDATED,
		U.EVENT_NAME_LEAD_SQUARED_LEAD_UPDATED,
		U.EVENT_NAME_MARKETO_LEAD_UPDATED,
		U.EVENT_NAME_SALESFORCE_ACCOUNT_UPDATED,
		U.EVENT_NAME_SALESFORCE_OPPORTUNITY_UPDATED,
	)
	rows, err := db.Raw(eventsQuery, projectID, gorm.NowFunc().Unix(), projectID, identity, projectID, projectID).Rows()

	if err != nil || rows.Err() != nil {
		log.WithFields(logFields).WithError(err).WithError(rows.Err()).Error("Failed to get events")
		return []model.UserActivity{}, uint64(webSessionCount)
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
			return []model.UserActivity{}, uint64(webSessionCount)
		}
		// Session Count workaround
		if userActivity.EventName == U.EVENT_NAME_SESSION {
			webSessionCount += 1
		}

		properties, err := U.DecodePostgresJsonb(userActivity.Properties)
		if err != nil {
			log.WithFields(logFields).WithError(err).Error("Failed decoding event properties")
		} else {
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
			} else if userActivity.EventName == U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_UPDATED {
				userActivity.AliasName = fmt.Sprintf("Interacted with %s", (*properties)[U.EP_SALESFORCE_CAMPAIGN_NAME])
			} else if userActivity.EventName == U.EVENT_NAME_HUBSPOT_CONTACT_FORM_SUBMISSION {
				userActivity.AliasName = fmt.Sprintf("%s", (*properties)[U.EP_HUBSPOT_FORM_SUBMISSION_TITLE])
			} else if userActivity.EventName == U.EVENT_NAME_HUBSPOT_ENGAGEMENT_EMAIL {
				emailSubject := "No Subject"
				if subject, exists := (*properties)[U.EP_HUBSPOT_ENGAGEMENT_SUBJECT]; exists {
					emailSubject = fmt.Sprintf("%s", subject)
				}
				userActivity.AliasName = fmt.Sprintf("%s: %s", (*properties)[U.EP_HUBSPOT_ENGAGEMENT_TYPE], emailSubject)
			} else if userActivity.EventName == U.EVENT_NAME_HUBSPOT_ENGAGEMENT_MEETING_CREATED ||
				userActivity.EventName == U.EVENT_NAME_HUBSPOT_ENGAGEMENT_MEETING_UPDATED ||
				userActivity.EventName == U.EVENT_NAME_HUBSPOT_ENGAGEMENT_CALL_CREATED ||
				userActivity.EventName == U.EVENT_NAME_HUBSPOT_ENGAGEMENT_CALL_UPDATED {
				userActivity.AliasName = fmt.Sprintf("%s", (*properties)[U.EP_HUBSPOT_ENGAGEMENT_TITLE])
			}

			// Set Icons
			if icon, exists := model.EVENT_ICONS_MAP[userActivity.EventName]; exists {
				userActivity.Icon = icon
			} else if strings.Contains(userActivity.EventName, "hubspot_") || strings.Contains(userActivity.EventName, "hs_") {
				userActivity.Icon = "hubspot_ads"
			} else if strings.Contains(userActivity.EventName, "salesforce_") || strings.Contains(userActivity.EventName, "sf_") {
				userActivity.Icon = "salesforce_ads"
			}
			// Default Icon
			if userActivity.Icon == "" {
				userActivity.Icon = "calendar_star"
			}

			// Filtered Properties
			userActivity.Properties = GetFilteredProperties(userActivity.EventName, userActivity.EventType, properties)
		}
		userActivities = append(userActivities, userActivity)
	}

	return userActivities, uint64(webSessionCount)
}

func (store *MemSQL) GetGroupsForUserTimeline(projectID int64, userDetails model.ContactDetails) []model.GroupsInfo {
	var groupsInfo []model.GroupsInfo
	groups, errCode := store.GetGroups(projectID)
	if errCode != http.StatusFound {
		log.WithField("project_id", projectID).WithField("status", errCode).Error("Failed to get groups while adding group info.")
		return []model.GroupsInfo{}
	}
	if errCode == http.StatusFound && len(groups) == 0 {
		return []model.GroupsInfo{}
	}

	groupsMap := make(map[int]string)
	for _, value := range groups {
		groupsMap[value.ID] = U.STANDARD_GROUP_DISPLAY_NAMES[value.Name]
	}

	if userDetails.Group1 {
		groupsInfo = append(groupsInfo, model.GroupsInfo{GroupName: groupsMap[1]})
	}
	if userDetails.Group2 {
		groupsInfo = append(groupsInfo, model.GroupsInfo{GroupName: groupsMap[2]})
	}
	if userDetails.Group3 {
		groupsInfo = append(groupsInfo, model.GroupsInfo{GroupName: groupsMap[3]})
	}
	if userDetails.Group4 {
		groupsInfo = append(groupsInfo, model.GroupsInfo{GroupName: groupsMap[4]})
	}
	return groupsInfo
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
		if value, propExists := (*properties)[U.EP_TIMESTAMP]; propExists {
			filteredProperties[U.EP_TIMESTAMP] = value
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

func (store *MemSQL) GetProfileAccountDetailsByID(projectID int64, id string) (*model.AccountDetails, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"id":         id,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if projectID == 0 {
		log.Error("Invalid project ID.")
		return nil, http.StatusBadRequest
	}
	if id == "" {
		log.Error("Invalid account ID.")
		return nil, http.StatusBadRequest
	}

	groups, errCode := store.GetGroups(projectID)
	if errCode != http.StatusFound {
		log.WithField("status", errCode).Error("Failed to get groups.")
	}
	var groupUserString string
	groupNameIDMap := make(map[string]int)
	if len(groups) > 0 {
		for index, group := range groups {
			if group.Name == model.GROUP_NAME_HUBSPOT_COMPANY || group.Name == model.GROUP_NAME_SALESFORCE_ACCOUNT {
				groupNameIDMap[group.Name] = group.ID
			}
			if index == 0 {
				groupUserString = fmt.Sprintf("group_%d_user_id='%s' ", group.ID, id)
			} else {
				groupUserString = groupUserString + fmt.Sprintf(" OR group_%d_user_id='%s' ", group.ID, id)
			}
		}
	}

	db := C.GetServices().Db
	var accountDetails model.AccountDetails
	err := db.Table("users").Select("properties").Where("project_id=? AND id=?", projectID, id).Limit(1).Find(&accountDetails).Error
	if err != nil {
		log.WithFields(logFields).WithField("status", err).Error("Failed to get account properties.")
		return nil, http.StatusInternalServerError
	}

	// Filter Properties
	nameProps := []string{U.UP_COMPANY, U.GP_HUBSPOT_COMPANY_NAME, U.GP_SALESFORCE_ACCOUNT_NAME}
	hostNameProps := []string{U.GP_HUBSPOT_COMPANY_DOMAIN, U.GP_SALESFORCE_ACCOUNT_WEBSITE}
	propertiesDecoded, err := U.DecodePostgresJsonb(accountDetails.Properties)
	if err != nil {
		log.WithFields(logFields).WithError(err).Error("Failed decoding account properties.")
	}
	for _, prop := range nameProps {
		if accountDetails.Name != "" {
			break
		}
		if name, exists := (*propertiesDecoded)[prop]; exists {
			accountDetails.Name = fmt.Sprintf("%s", name)
		}
	}
	for _, prop := range hostNameProps {
		if accountDetails.HostName != "" {
			break
		}
		if host, exists := (*propertiesDecoded)[prop]; exists {
			accountDetails.HostName = fmt.Sprintf("%s", host)
		}
	}

	accountDetails.LeftPaneProps = store.GetLeftPaneProperties(projectID, model.PROFILE_TYPE_ACCOUNT, propertiesDecoded)
	timelinesConfig, err := store.GetTimelineConfigOfProject(projectID)
	if err != nil {
		log.WithField("status", err).WithError(err).Error("Failed to fetch timelines_config from project_settings.")
	}
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
	WHERE project_id = ? AND (%s) 
	GROUP BY user_id 
	ORDER BY updated_at DESC 
	LIMIT 26;`, U.UP_NAME, selectStrAdditionalProp, groupUserString)

	// Get Timeline for <=25 users
	rows, err := db.Raw(query, projectID).Rows()
	if err != nil || rows.Err() != nil {
		log.WithFields(logFields).WithError(err).WithError(rows.Err()).Error("Failed to get associated users")
		return nil, http.StatusInternalServerError
	}
	defer U.CloseReadQuery(rows, nil)
	var accountTimeline []model.UserTimeline
	usersCount := 0
	for rows.Next() {
		usersCount += 1
		// Error log for Count of Users
		if usersCount == 26 {
			log.WithFields(logFields).Error("Number of users greater than 25")
			break
		}
		var userTimeline model.UserTimeline
		if err := db.ScanRows(rows, &userTimeline); err != nil {
			log.WithFields(logFields).WithError(err).Error("Error scanning associated users list")
			return nil, http.StatusInternalServerError
		}
		var userIDStr string
		if userTimeline.IsAnonymous {
			userIDStr = "id"
		} else {
			userIDStr = "customer_user_id"
		}
		activities, _ := store.GetUserActivitiesAndSessionCount(projectID, userTimeline.UserId, userIDStr)
		userTimeline.UserActivities = activities
		accountTimeline = append(accountTimeline, userTimeline)
	}
	accountDetails.AccountTimeline = accountTimeline
	return &accountDetails, http.StatusFound
}

func FormatTimeToString(time time.Time) string {
	return time.Format("2006-01-02 15:04:05.000000")
}

func (store *MemSQL) GetLeftPaneProperties(projectID int64, profileType string, propertiesDecoded *map[string]interface{}) map[string]interface{} {
	logFields := log.Fields{
		"project_id":   projectID,
		"profile_type": profileType,
	}

	filteredProperties := make(map[string]interface{})

	timelinesConfig, err := store.GetTimelineConfigOfProject(projectID)
	if err != nil {
		log.WithFields(logFields).WithField("status", err).WithError(err).Error("Failed to fetch timelines_config from project_settings.")
		return filteredProperties
	}
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
		logCtx.Error("Failed at building query. ", errMsg)
		return nil, errCode, nil
	}

	profiles := make([]model.Profile, 0)
	if segmentQuery.Caller == model.USER_PROFILE_CALLER {
		for _, profile := range result.Rows {
			var row model.Profile
			row.TableProps = make(map[string]interface{}, 0)
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
				logCtx.Error("Failed at unmarshalling props. ")
				return nil, http.StatusInternalServerError, nil
			}
			filterTableProps := make(map[string]interface{}, 0)

			for _, prop := range segmentQuery.TableProps {
				if value, exists := (props)[prop]; exists {
					filterTableProps[prop] = value
				}
			}
			row.TableProps = filterTableProps
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
				logCtx.Error("Failed at unmarshalling props.")
				return nil, http.StatusInternalServerError, nil
			}
			filterTableProps := make(map[string]interface{}, 0)

			for _, prop := range segmentQuery.TableProps {
				if value, exists := (props)[prop]; exists {
					filterTableProps[prop] = value
				}
			}
			row.TableProps = filterTableProps
			profiles = append(profiles, row)
		}
	}

	return profiles, http.StatusOK, nil
}
