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

	// Merge Filters
	payload.Filters = append(payload.Filters, payload.SearchFilter...)

	var tableProps []string
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
		tableProps = segmentQuery.TableProps
		segmentQuery.From = U.TimeNowZ().AddDate(0, 0, -28).Unix()
		segmentQuery.To = U.TimeNowZ().Unix()
		if segmentQuery.EventsWithProperties != nil && len(segmentQuery.EventsWithProperties) > 0 {
			if C.IsEnabledEventsFilterInSegments() {
				segmentQuery.GlobalUserProperties = append(segmentQuery.GlobalUserProperties, payload.Filters...)
				query, err := U.EncodeStructTypeToPostgresJsonb(segmentQuery)
				if err != nil {
					log.WithFields(logFields).Error("Failed to append payload filters with global properties.")
				} else {
					segment.Query = query
				}
				profiles, errCode, _ := store.GetAnalyzeResultForSegments(projectID, segment)
				if errCode != http.StatusOK {
					return nil, errCode
				}
				// Get Table Content
				returnData, err := FormatProfilesStruct(profiles, profileType, tableProps)
				if err != nil {
					log.WithFields(logFields).WithField("status", err).Error("Failed to filter properties from profiles.")
					return nil, http.StatusInternalServerError
				}
				return returnData, http.StatusFound
			} else {
				var profiles = make([]model.Profile, 0)
				return profiles, http.StatusBadRequest
			}
		} else {
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

	var selectString, isGroupUserString, sourceString string

	if profileType == model.PROFILE_TYPE_ACCOUNT {
		// Check for Enabled Groups
		groups, errCode := store.GetGroups(projectID)
		if errCode != http.StatusFound {
			log.WithField("err_code", errCode).Error("Failed to get groups while adding group info.")
		}
		groupNameIDMap := make(map[string]int)
		if len(groups) > 0 {
			for _, group := range groups {
				if model.IsAllowedAccountGroupNames(group.Name) {
					groupNameIDMap[group.Name] = group.ID
				}
			}
		}
		hubspotID, hubspotExists := groupNameIDMap[model.GROUP_NAME_HUBSPOT_COMPANY]
		salesforceID, salesforceExists := groupNameIDMap[model.GROUP_NAME_SALESFORCE_ACCOUNT]
		sixSignalID, sixSignalExists := groupNameIDMap[model.GROUP_NAME_SIX_SIGNAL]

		if !hubspotExists && !salesforceExists && !sixSignalExists {
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
		if payload.Source == model.GROUP_NAME_SIX_SIGNAL && !sixSignalExists {
			log.WithFields(logFields).Error("Salesforce Not Enabled for this project.")
			return nil, http.StatusBadRequest
		}
		selectString = "id AS identity, properties, updated_at AS last_activity"
		isGroupUserString = "is_group_user=1"
		if payload.Source == "All" {
			var sourceArr []string
			if salesforceExists {
				sourceArr = append(sourceArr, fmt.Sprintf("group_%d_id IS NOT NULL", salesforceID))
			}
			if sixSignalExists {
				sourceArr = append(sourceArr, fmt.Sprintf("group_%d_id IS NOT NULL", sixSignalID))
			}
			if hubspotExists {
				sourceArr = append(sourceArr, fmt.Sprintf("group_%d_id IS NOT NULL", hubspotID))
			}
			sourceStr := strings.Join(sourceArr, " OR ")
			if sourceStr != "" {
				sourceString = fmt.Sprintf("AND (%s)", sourceStr)
			}
		} else if (payload.Source == "All" || payload.Source == model.GROUP_NAME_HUBSPOT_COMPANY) && hubspotExists {
			sourceString = fmt.Sprintf("AND group_%d_id IS NOT NULL", hubspotID)
		} else if (payload.Source == "All" || payload.Source == model.GROUP_NAME_SALESFORCE_ACCOUNT) && salesforceExists {
			sourceString = fmt.Sprintf("AND group_%d_id IS NOT NULL", salesforceID)
		} else if (payload.Source == "All" || payload.Source == model.GROUP_NAME_SIX_SIGNAL) && sixSignalExists {
			sourceString = fmt.Sprintf("AND group_%d_id IS NOT NULL", sixSignalID)
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
		log.WithError(err).WithFields(logFields).WithField("status", err).Error("min and max updated_at couldn't be defined.")
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
		log.WithError(err).WithFields(logFields).WithField("status", err).Error("Failed to get profile users.")
		return nil, http.StatusInternalServerError
	}

	// Get Table Content
	returnData, err := FormatProfilesStruct(profiles, profileType, tableProps)
	if err != nil {
		log.WithError(err).WithFields(logFields).WithField("status", err).Error("Failed to filter properties from profiles.")
		return nil, http.StatusInternalServerError
	}
	return returnData, http.StatusFound
}

func FormatProfilesStruct(profiles []model.Profile, profileType string, tableProps []string) ([]model.Profile, error) {
	logFields := log.Fields{
		"profile_type": profileType,
	}

	if profileType == model.PROFILE_TYPE_ACCOUNT {
		companyNameProps := []string{U.UP_COMPANY, U.GP_HUBSPOT_COMPANY_NAME, U.GP_HUBSPOT_COMPANY_DOMAIN, U.GP_SALESFORCE_ACCOUNT_NAME, U.SIX_SIGNAL_NAME}
		hostNameProps := []string{U.GP_HUBSPOT_COMPANY_DOMAIN, U.GP_SALESFORCE_ACCOUNT_WEBSITE, U.SIX_SIGNAL_DOMAIN}

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
	if projectID == 0 || identity == "" || isAnonymous == "" {
		log.WithFields(logFields).Error("invalid payload.")
		return nil, http.StatusBadRequest
	}

	userId := "customer_user_id"
	if isAnonymous == "true" {
		userId = "id"
	}

	db := C.GetServices().Db
	var uniqueUser model.ContactDetails
	if err := db.Table("users").Select(`COALESCE(customer_user_id,id) AS user_id,
		ISNULL(customer_user_id) AS is_anonymous,
		properties,
		MAX(group_1_id) IS NOT NULL AS group_1,
		MAX(group_2_id) IS NOT NULL AS group_2,
		MAX(group_3_id) IS NOT NULL AS group_3,
		MAX(group_4_id) IS NOT NULL AS group_4,
		MAX(group_5_id) IS NOT NULL AS group_5,
		MAX(group_6_id) IS NOT NULL AS group_6,
		MAX(group_7_id) IS NOT NULL AS group_7,
		MAX(group_8_id) IS NOT NULL AS group_8,
		MAX(group_1_user_id) AS group_1_user_id,
		MAX(group_2_user_id) AS group_2_user_id,
		MAX(group_3_user_id) AS group_3_user_id,
		MAX(group_4_user_id) AS group_4_user_id,
		MAX(group_5_user_id) AS group_5_user_id,
		MAX(group_6_user_id) AS group_6_user_id,
		MAX(group_7_user_id) AS group_7_user_id,
		MAX(group_8_user_id) AS group_8_user_id
		`).
		Where("project_id=? AND "+userId+"=?", projectID, identity).
		Group("user_id").
		Order("updated_at desc").
		Limit(1).
		Find(&uniqueUser).Error; err != nil {
		log.WithError(err).WithFields(logFields).WithField("status", err).Error("Failed to get contact details.")
		return nil, http.StatusInternalServerError
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
	AND event_names.project_id=? 
	ORDER BY timestamp DESC;`, userId, eventNamesToExcludePlaceholders)

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

	return userActivities, uint64(webSessionCount)
}

func (store *MemSQL) GetGroupsForUserTimeline(projectID int64, userDetails model.ContactDetails) []model.GroupsInfo {
	groupsInfo := []model.GroupsInfo{}

	groups, err := store.GetGroups(projectID)
	if err != http.StatusFound {
		log.WithField("project_id", projectID).WithField("status", err).Error("Failed to get groups.")
		return groupsInfo
	}

	if len(groups) == 0 {
		return groupsInfo
	}

	groupsMap := make(map[int]string)
	for _, group := range groups {
		groupsMap[group.ID] = group.Name
	}

	for i := 1; i <= model.AllowedGroups; i++ {
		// Get the groupX field from the userDetails struct
		groupField := reflect.ValueOf(userDetails).FieldByName(fmt.Sprintf("Group%d", i))

		// Skip if groupX is 0
		if !groupField.Bool() {
			continue
		}

		groupInfo := model.GroupsInfo{GroupName: U.STANDARD_GROUP_DISPLAY_NAMES[groupsMap[i]]}

		userIDField := reflect.ValueOf(userDetails).FieldByName(fmt.Sprintf("Group%dUserID", i)) // Get the group_x_user_id field for the group
		if userIDField.String() != "" {
			associatedGroup, err := store.GetAssociatedGroup(projectID, userIDField.String(), groupsMap[i]) // Call GetAssociatedGroup method to get the associated group name for the user
			if err != nil {
				if gorm.IsRecordNotFoundError(err) {
					log.WithField("project_id", projectID).WithField("status", err).Error("Group record not found for user.")
				}
			} else {
				groupInfo.AssociatedGroup = associatedGroup // Set the associated group name for the groupInfo object
			}
		}
		groupsInfo = append(groupsInfo, groupInfo) // Append the groupInfo object to the groupsInfo slice
	}
	return groupsInfo
}

func (store *MemSQL) GetAssociatedGroup(projectID int64, userID string, groupName string) (string, error) {
	db := C.GetServices().Db
	companyQuery := "SELECT JSON_EXTRACT_STRING(properties, ?) AS associated_group FROM users WHERE project_id=? AND id=?"
	groupInfo := model.GroupsInfo{}
	if err := db.Raw(companyQuery, model.GROUP_TO_COMPANY_NAME_MAP[groupName], projectID, userID).Scan(&groupInfo).Limit(1).Error; err != nil {
		return "", err
	}
	return groupInfo.AssociatedGroup, nil
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
		log.WithField("err_code", errCode).Error("Failed to get groups.")
	}
	var groupUserString string
	groupNameIDMap := make(map[string]int)
	if len(groups) > 0 {
		for index, group := range groups {
			if model.IsAllowedAccountGroupNames(group.Name) {
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
		log.WithError(err).WithFields(logFields).Error("Failed to get account properties.")
		return nil, http.StatusInternalServerError
	}

	// Filter Properties
	nameProps := []string{U.UP_COMPANY, U.GP_HUBSPOT_COMPANY_NAME, U.GP_SALESFORCE_ACCOUNT_NAME, U.SIX_SIGNAL_NAME}
	hostNameProps := []string{U.GP_HUBSPOT_COMPANY_DOMAIN, U.GP_SALESFORCE_ACCOUNT_WEBSITE, U.SIX_SIGNAL_DOMAIN}
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

	timelinesConfig, err := store.GetTimelineConfigOfProject(projectID)
	if err != nil {
		log.WithError(err).Error("Failed to fetch timelines_config from project_settings.")
	}

	accountDetails.LeftPaneProps = GetLeftPanePropertiesFromConfig(timelinesConfig, model.PROFILE_TYPE_ACCOUNT, propertiesDecoded)
	accountDetails.Milestones = GetMilestonesFromConfig(timelinesConfig, model.PROFILE_TYPE_ACCOUNT, propertiesDecoded)

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
