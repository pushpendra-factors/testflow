package memsql

import (
	"encoding/json"
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

// Sample Query Accounts Timeline:
/* SELECT id, JSON_EXTRACT_STRING(properties, '$company') AS name, updated_at AS last_activity FROM `users`
   WHERE (
	   project_id = ? AND
		(is_group_user=1 OR is_group_user IS NOT NULL) AND
		(group_X_id IS NOT NULL OR group_Y_id IS NOT NULL) AND
		(JSON_EXTRACT_STRING(users.properties, '$country') = ?) AND
		updated_at BETWEEN ? AND ?
	)
	ORDER BY last_activity DESC LIMIT 1000
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

	var isGroupUserString string
	var sourceString string

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
		_, hubspotExists := groupNameIDMap[model.GROUP_NAME_HUBSPOT_COMPANY]
		_, salesforceExists := groupNameIDMap[model.GROUP_NAME_SALESFORCE_ACCOUNT]

		if !hubspotExists && !salesforceExists {
			log.Error("Failed to get a group user.")
			return nil, http.StatusBadRequest
		}
		// Where Strings
		isGroupUserString = "(is_group_user=1 OR is_group_user IS NOT NULL)"
		if payload.Source == "All" && hubspotExists && salesforceExists {
			sourceString = fmt.Sprintf(" AND (group_%d_id IS NOT NULL OR group_%d_id IS NOT NULL)", groupNameIDMap[model.GROUP_NAME_HUBSPOT_COMPANY], groupNameIDMap[model.GROUP_NAME_SALESFORCE_ACCOUNT])
		} else if (payload.Source == "All" || payload.Source == model.GROUP_NAME_HUBSPOT_COMPANY) && hubspotExists {
			sourceString = fmt.Sprintf(" AND group_%d_id IS NOT NULL", groupNameIDMap[model.GROUP_NAME_HUBSPOT_COMPANY])
		} else if (payload.Source == "All" || payload.Source == model.GROUP_NAME_SALESFORCE_ACCOUNT) && salesforceExists {
			sourceString = fmt.Sprintf(" AND group_%d_id IS NOT NULL", groupNameIDMap[model.GROUP_NAME_SALESFORCE_ACCOUNT])
		}
	} else if profileType == model.PROFILE_TYPE_USER {
		isGroupUserString = "(is_group_user=0 OR is_group_user IS NULL)"
		if model.UserSourceMap[payload.Source] == model.UserSourceWeb {
			sourceString = "AND (source=" + strconv.Itoa(model.UserSourceMap[payload.Source]) + " OR source IS NULL)"
		} else if payload.Source == "All" {
			sourceString = ""
		} else {
			sourceString = "AND source=" + strconv.Itoa(model.UserSourceMap[payload.Source])
		}
	}

	whereStr := []string{"project_id = ? AND", isGroupUserString, sourceString}
	parameters := []interface{}{projectID}

	if len(payload.Filters) > 0 {
		filterString, filterParams, err := buildWhereFromProperties(projectID, payload.Filters, 0)
		if filterString != "" {
			filterString = " AND " + filterString
		}
		if err != nil {
			return nil, http.StatusBadRequest
		}
		whereStr = append(whereStr, filterString)
		parameters = append(parameters, filterParams...)
	}

	whereString := strings.Join(whereStr, " ")
	parameters = append(parameters, gorm.NowFunc())

	type MinMaxTime struct {
		MinUpdatedAt time.Time `json:"min_updated_at"`
		MaxUpdatedAt time.Time `json:"max_updated_at"`
	}
	var minMax MinMaxTime
	// Get min and max updated_at for 100k after
	// ordering as part of optimisation.
	db := C.GetServices().Db
	err := db.Raw(`SELECT MIN(updated_at) AS min_updated_at, MAX(updated_at) AS max_updated_at 
		FROM (SELECT updated_at FROM users WHERE `+whereString+` AND updated_at < ? 
		ORDER BY updated_at DESC LIMIT 100000)`, parameters...).
		Scan(&minMax).Error
	if err != nil {
		log.WithField("status", err).Error("min and max updated_at couldn't be defined.")
		return nil, http.StatusInternalServerError
	}

	var profiles []model.Profile
	parameters = parameters[:len(parameters)-1]
	parameters = append(parameters, minMax.MinUpdatedAt, minMax.MaxUpdatedAt)

	var selectString string
	if profileType == model.PROFILE_TYPE_ACCOUNT {
		selectString = fmt.Sprintf("id AS identity, JSON_EXTRACT_STRING(properties, '%s') AS name, updated_at AS last_activity", U.UP_COMPANY)
	} else if profileType == model.PROFILE_TYPE_USER {
		selectString = fmt.Sprintf("COALESCE(customer_user_id, id) AS identity, ISNULL(customer_user_id) AS is_anonymous, JSON_EXTRACT_STRING(properties, '%s') AS country, MAX(updated_at) AS last_activity", U.UP_COUNTRY)
	}

	err = db.Table("users").Select(selectString).Where(whereString+` AND updated_at BETWEEN ? AND ?`, parameters...).Group("identity").Order("last_activity DESC").Limit(1000).Find(&profiles).Error
	if err != nil {
		log.WithField("status", err).Error("Failed to get profile users.")
		return nil, http.StatusInternalServerError
	}
	return profiles, http.StatusFound
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
		JSON_EXTRACT_STRING(properties, ?) AS name,
		JSON_EXTRACT_STRING(properties, ?) AS company,
		JSON_EXTRACT_STRING(properties, ?) AS email,
		JSON_EXTRACT_STRING(properties, ?) AS country,
		MAX(JSON_EXTRACT_STRING(properties, ?)) AS web_sessions_count,
		MAX(JSON_EXTRACT_STRING(properties, ?)) AS time_spent_on_site,
		MAX(JSON_EXTRACT_STRING(properties, ?)) AS number_of_page_views,
		GROUP_CONCAT(group_1_id) IS NOT NULL AS group_1,
		GROUP_CONCAT(group_2_id) IS NOT NULL AS group_2,
		GROUP_CONCAT(group_3_id) IS NOT NULL AS group_3,
		GROUP_CONCAT(group_4_id) IS NOT NULL AS group_4`,
		U.UP_NAME, U.UP_COMPANY, U.UP_EMAIL, U.UP_COUNTRY, U.UP_SESSION_COUNT, U.UP_TOTAL_SPENT_TIME, U.UP_PAGE_COUNT).
		Where("project_id=? AND "+userId+"=?", projectID, identity).
		Group("user_id").
		Order("updated_at desc").
		Limit(1).
		Find(&uniqueUser).Error
	if err != nil {
		log.WithField("status", err).Error("Failed to get contact details.")
		return nil, http.StatusInternalServerError
	}

	activities, sessionCount := store.GetUserActivitiesAndSessionCount(projectID, identity, userId)
	uniqueUser.UserActivity = activities
	uniqueUser.WebSessionsCount = sessionCount
	uniqueUser.GroupInfos = store.GetGroupsForUserTimeline(projectID, uniqueUser)

	return &uniqueUser, http.StatusFound
}

func (store *MemSQL) GetUserActivitiesAndSessionCount(projectID int64, identity string, userId string) ([]model.UserActivity, float64) {
	var userActivities []model.UserActivity
	webSessionCount := 0

	db := C.GetServices().Db
	str := []string{`SELECT event_names.name AS event_name, events1.timestamp AS timestamp, events1.properties AS properties
        FROM (SELECT project_id, event_name_id, timestamp, properties FROM events WHERE 
			project_id=? AND timestamp <= ? AND 
			user_id IN (SELECT id FROM users WHERE project_id=? AND`, userId, `= ?)  LIMIT 5000) AS events1
        LEFT JOIN event_names 
        ON events1.event_name_id=event_names.id 
        WHERE events1.project_id=? 
		ORDER BY timestamp DESC;`}
	eventsQuery := strings.Join(str, " ")
	rows, err := db.Raw(eventsQuery, projectID, gorm.NowFunc().Unix(), projectID, identity, projectID).Rows()
	if err != nil {
		log.WithError(err).Error("Failed to get events")
		return []model.UserActivity{}, float64(webSessionCount)
	}
	// User Activity
	for rows.Next() {
		var userActivity model.UserActivity
		if err := db.ScanRows(rows, &userActivity); err != nil {
			log.WithError(err).Error("Failed scanning events list")
			return []model.UserActivity{}, float64(webSessionCount)
		}
		// Session Count workaround
		if userActivity.EventName == U.EVENT_NAME_SESSION {
			webSessionCount += 1
		}

		properties, err := U.DecodePostgresJsonb(&userActivity.Properties)
		if err != nil {
			log.WithError(err).Error("Failed decoding event properties")
		} else {
			// Display Names
			userActivity.DisplayName = store.GetDisplayNameForTimelineEvents(projectID, userActivity.EventName, properties)
			// Filtered Properties
			userActivity.Properties = GetFilteredProperties(userActivity.EventName, properties)
		}
		userActivities = append(userActivities, userActivity)
	}
	return userActivities, float64(webSessionCount)
}

func (store *MemSQL) GetGroupsForUserTimeline(projectID int64, userDetails model.ContactDetails) []model.GroupsInfo {
	var groupsInfo []model.GroupsInfo
	groups, errCode := store.GetGroups(projectID)
	if errCode != http.StatusFound {
		log.WithField("status", errCode).Error("Failed to get groups while adding group info.")
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

func (store *MemSQL) GetDisplayNameForTimelineEvents(projectId int64, eventName string, properties *map[string]interface{}) string {
	var displayName string
	standardDisplayNames := U.STANDARD_EVENTS_DISPLAY_NAMES
	_, projectDisplayNames := store.GetDisplayNamesForAllEvents(projectId)
	if (*properties)[U.EP_IS_PAGE_VIEW] == true {
		displayName = "Page View"
	} else if standardDisplayNames[eventName] != "" {
		displayName = standardDisplayNames[eventName]
	} else if projectDisplayNames[eventName] != "" {
		displayName = projectDisplayNames[eventName]
	} else {
		displayName = eventName
	}
	return displayName
}

func GetFilteredProperties(eventName string, properties *map[string]interface{}) postgres.Jsonb {
	var returnProperties postgres.Jsonb
	eventNamePropertiesMap := map[string][]string{
		U.EVENT_NAME_SESSION:                           {U.EP_PAGE_COUNT, U.EP_CHANNEL, U.EP_CAMPAIGN, U.SP_SESSION_TIME, U.EP_TIMESTAMP, U.EP_REFERRER_URL},
		U.EVENT_NAME_FORM_SUBMITTED:                    {U.EP_FORM_NAME, U.EP_PAGE_URL, U.EP_TIMESTAMP},
		U.EVENT_NAME_OFFLINE_TOUCH_POINT:               {U.EP_CHANNEL, U.EP_CAMPAIGN, U.EP_TIMESTAMP},
		U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_CREATED: {"$salesforce_campaign_name", model.EP_SFCampaignMemberStatus, U.EP_TIMESTAMP},
		U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_UPDATED: {"$salesforce_campaign_name", model.EP_SFCampaignMemberStatus, U.EP_TIMESTAMP},
	}
	pageViewPropsList := []string{U.EP_IS_PAGE_VIEW, U.EP_PAGE_SPENT_TIME, U.EP_PAGE_SCROLL_PERCENT, U.EP_PAGE_LOAD_TIME}
	filteredProperties := make(map[string]interface{})
	_, eventExistsInMap := eventNamePropertiesMap[eventName]
	if (*properties)[U.EP_IS_PAGE_VIEW] == true {
		for _, prop := range pageViewPropsList {
			_, propExists := (*properties)[prop]
			if propExists {
				filteredProperties[prop] = (*properties)[prop]
			}
		}
		propertiesJSON, err := json.Marshal(filteredProperties)
		if err != nil {
			log.WithError(err).Error("filter properties marshal error.")
		}
		returnProperties = postgres.Jsonb{RawMessage: propertiesJSON}
	} else if eventExistsInMap {
		for _, prop := range eventNamePropertiesMap[eventName] {
			_, propExists := (*properties)[prop]
			if propExists {
				filteredProperties[prop] = (*properties)[prop]
			}
		}
		propertiesJSON, err := json.Marshal(filteredProperties)
		if err != nil {
			log.WithError(err).Error("filter properties marshal error.")
		}
		returnProperties = postgres.Jsonb{RawMessage: propertiesJSON}
	} else {
		returnProperties = postgres.Jsonb{RawMessage: json.RawMessage(`{}`)}
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

	db := C.GetServices().Db
	var accountDetails model.AccountDetails
	err := db.Table("users").Select("JSON_EXTRACT_STRING(properties, ?) AS name, JSON_EXTRACT_STRING(properties, ?) AS country", U.UP_COMPANY, U.UP_COUNTRY).
		Where("project_id=? AND id=?", projectID, id).
		Order("updated_at desc").
		Limit(1).
		Find(&accountDetails).Error
	if err != nil {
		log.WithField("status", err).Error("Failed to get account properties.")
		return nil, http.StatusInternalServerError
	}

	groups, errCode := store.GetGroups(projectID)
	if errCode != http.StatusFound {
		log.WithField("status", errCode).Error("Failed to get groups.")
	}
	var groupUserString string
	if len(groups) > 0 {
		for index, group := range groups {
			if index == 0 {
				groupUserString = fmt.Sprintf("group_%d_user_id='%s' ", group.ID, id)
			} else {
				groupUserString = groupUserString + fmt.Sprintf(" OR group_%d_user_id='%s' ", group.ID, id)
			}
		}
	}

	queryStr := []string{"SELECT JSON_EXTRACT_STRING(properties, ?) AS user_name, id AS user_id FROM users WHERE project_id = ? AND", groupUserString}
	query := strings.Join(queryStr, " ")
	rows, err := db.Raw(query, U.UP_NAME, projectID).Rows()
	if err != nil {
		log.WithError(err).Error("Failed to get associated users")
		return nil, http.StatusInternalServerError
	}
	var accountTimeline []model.UserTimeline
	for rows.Next() {
		var userTimeline model.UserTimeline
		if err := db.ScanRows(rows, &userTimeline); err != nil {
			log.WithError(err).Error("Error scanning associated users list")
			return nil, http.StatusInternalServerError
		}
		activities, _ := store.GetUserActivitiesAndSessionCount(projectID, userTimeline.UserId, "id")
		userTimeline.UserActivities = activities
		accountTimeline = append(accountTimeline, userTimeline)
	}
	accountDetails.AccountTimeline = accountTimeline
	return &accountDetails, http.StatusFound
}
