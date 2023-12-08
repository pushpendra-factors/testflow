package memsql

import (
	"database/sql"
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"net/http"
	"reflect"
	"time"

	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) GetMarkedDomainsListByProjectId(projectID int64, payload model.TimelinePayloadSegment) ([]model.Profile, int, string) {
	logFields := log.Fields{
		"project_id": projectID,
		"payload":    payload,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	if projectID == 0 {
		return []model.Profile{}, http.StatusBadRequest, "Project Id is Invalid"
	}

	// return if source is not $domains
	if payload.Query.Source != model.GROUP_NAME_DOMAINS {
		logCtx.Error("Query failed. invalid source")
		return []model.Profile{}, http.StatusBadRequest, "Failed to fetch source."
	}

	oldTimelineFlow := model.TimelinePayload{
		Query:        payload.Query,
		SearchFilter: payload.SearchFilter,
	}

	// Redirect to old flow if segment_id is empty
	if payload.SegmentId == "" {
		return store.GetProfilesListByProjectId(projectID, oldTimelineFlow, model.PROFILE_TYPE_ACCOUNT)
	}

	// Redirect to old flow if additional filters exist
	segment, statusCode := store.GetSegmentById(projectID, payload.SegmentId)
	if statusCode != http.StatusFound {
		logCtx.Error("Segment not found.")
		return []model.Profile{}, statusCode, "Failed to get segment"
	}
	segmentQuery := model.Query{}
	err := U.DecodePostgresJsonbToStructType(segment.Query, &segmentQuery)
	if err != nil {
		log.WithField("project_id", projectID).Error("Unable to decode segment query")
		return []model.Profile{}, statusCode, "Failed to get segment"
	}

	additionalFiltersExist := CompareFilters(segmentQuery, payload.Query)

	if additionalFiltersExist {
		return store.GetProfilesListByProjectId(projectID, oldTimelineFlow, model.PROFILE_TYPE_ACCOUNT)
	}

	// set Query Timezone
	timezoneString, statusCode := store.GetTimezoneForProject(projectID)
	if statusCode != http.StatusFound {
		logCtx.Error("Query failed. Failed to get Timezone.")
		return []model.Profile{}, http.StatusBadRequest, "Failed to fetch project timezone."
	}
	payload.Query.Timezone = string(timezoneString)

	domainGroup, errCode := store.GetGroup(projectID, model.GROUP_NAME_DOMAINS)
	if errCode != http.StatusFound || domainGroup == nil {
		return []model.Profile{}, http.StatusInternalServerError, "failed to get group while adding group info"
	}

	// check for search filter
	whereForSearchFilters, searchFiltersParams := SearchFilterForAllAccounts(payload.SearchFilter, domainGroup.ID)

	query := fmt.Sprintf(`SELECT 
	identity, 
	host_name, 
	last_activity 
    FROM 
	( SELECT 
		id as identity, 
		group_%d_id as host_name, 
		JSON_EXTRACT_STRING(associated_segments, ?) as segment_details, 
		CASE WHEN segment_details::$last_event_at != '' THEN segment_details::$last_event_at 
		WHEN segment_details::$updated_at != '' THEN segment_details::$updated_at 
		ELSE segment_details END as last_activity 
	  FROM 
		users 
	  WHERE 
		project_id = ? 
		AND is_group_user = 1 
		AND source = ? 
		AND segment_details IS NOT NULL %s 
	  ORDER BY 
		last_activity DESC 
	  LIMIT 
		1000);`, domainGroup.ID, whereForSearchFilters)

	params := []interface{}{payload.SegmentId, projectID, model.UserSourceDomains}

	if whereForSearchFilters != "" {
		params = append(params, searchFiltersParams...)
	}

	var profiles []model.Profile
	db := C.GetServices().Db
	rows, err := db.Raw(query, params...).Rows()
	if err != nil {
		return []model.Profile{}, http.StatusInternalServerError, ""
	}

	defer rows.Close()

	for rows.Next() {
		var id string
		var host_name_null sql.NullString
		var last_activity_null sql.NullString

		if err = rows.Scan(&id,
			&host_name_null, &last_activity_null); err != nil {
			logCtx.WithFields(log.Fields{"err": err, "project_id": projectID}).Error("SQL Parse failed.")
			return []model.Profile{}, http.StatusInternalServerError, ""
		}

		host_name := U.IfThenElse(host_name_null.Valid, host_name_null.String, false).(string)
		last_activity_str := U.IfThenElse(last_activity_null.Valid, last_activity_null.String, "").(string)

		location, err := time.LoadLocation(payload.Query.Timezone)
		if err != nil {
			logCtx.WithFields(log.Fields{"err": err}).Error("Loading location failed.")
			return []model.Profile{}, http.StatusInternalServerError, ""
		}
		layout := "2006-01-02 15:04:05.000000"
		timeValue, err := time.ParseInLocation(layout, last_activity_str, location)
		if err != nil {
			logCtx.WithFields(log.Fields{"err": err}).Error("Parsing location failed.")
			return []model.Profile{}, http.StatusInternalServerError, ""
		}

		profile := model.Profile{
			Identity:     id,
			HostName:     host_name,
			LastActivity: timeValue,
		}

		profiles = append(profiles, profile)
	}

	err = rows.Err()
	if err != nil {
		// Error from DB is captured eg: timeout error
		logCtx.WithFields(log.Fields{"err": err}).Error("Error in executing timelines query for segment")
		return []model.Profile{}, http.StatusInternalServerError, ""
	}

	// Redirect to old flow if no profiles found
	if len(profiles) == 0 {
		return store.GetProfilesListByProjectId(projectID, oldTimelineFlow, model.PROFILE_TYPE_ACCOUNT)
	}

	// datetime conversion
	err = segmentQuery.TransformDateTypeFilters()
	if err != nil {
		log.WithField("project_id", projectID).Error("Failed to transform segment filters.")
		return []model.Profile{}, statusCode, "Failed to get segment"
	}
	groupedFilters := GroupFiltersByGroupName(segmentQuery.GlobalUserProperties)

	// set TableProps
	if len(payload.Query.TableProps) == 0 {
		payload.Query.TableProps = store.GetTablePropsFromConfig(projectID, model.PROFILE_TYPE_ACCOUNT)
	}

	// Get merged properties for all accounts
	profiles, statusCode = store.AccountPropertiesForDomainsEnabled(projectID, profiles, groupedFilters, payload.Query.TableProps)
	if statusCode != http.StatusOK {
		return []model.Profile{}, statusCode, "Query Transformation Failed."
	}

	// Get Return Table Content
	returnData, err := FormatProfilesStruct(projectID, profiles, model.PROFILE_TYPE_ACCOUNT, payload.Query.TableProps, payload.Query.Source)
	if err != nil {
		logCtx.WithError(err).WithField("status", err).Error("Failed to filter properties from profiles.")
		return []model.Profile{}, http.StatusInternalServerError, "Query formatting failed."
	}

	return returnData, http.StatusFound, ""
}

func CompareFilters(segmentQuery model.Query, payloadQuery model.Query) bool {
	additionalFiltersExist := false

	if len(segmentQuery.GlobalUserProperties) != len(payloadQuery.GlobalUserProperties) ||
		len(segmentQuery.EventsWithProperties) != len(payloadQuery.EventsWithProperties) ||
		segmentQuery.EventsCondition != payloadQuery.EventsCondition {
		return true
	}

	// map of property and array of model.QueryProperty
	payloadGup := make(map[string][]model.QueryProperty)
	for _, gup := range payloadQuery.GlobalUserProperties {
		if _, exists := payloadGup[gup.Property]; !exists {
			payloadGup[gup.Property] = make([]model.QueryProperty, 0)
		}
		payloadGup[gup.Property] = append(payloadGup[gup.Property], gup)
	}

	// map of property and array of model.QueryEventWithProperties
	payloadEwp := make(map[string][]model.QueryEventWithProperties)
	for _, ewp := range payloadQuery.EventsWithProperties {
		if _, exists := payloadGup[ewp.Name]; !exists {
			payloadEwp[ewp.Name] = make([]model.QueryEventWithProperties, 0)
		}
		payloadEwp[ewp.Name] = append(payloadEwp[ewp.Name], ewp)
	}

	// check for gup
	for _, query := range segmentQuery.GlobalUserProperties {
		if _, exists := payloadGup[query.Property]; !exists {
			additionalFiltersExist = true
			break
		}
		filterExists := false
		for _, mapGupArr := range payloadGup[query.Property] {
			if reflect.DeepEqual(query, mapGupArr) {
				filterExists = true
				break
			}
		}
		if !filterExists {
			additionalFiltersExist = true
			break
		}
	}

	// gup not matched
	if additionalFiltersExist {
		return additionalFiltersExist
	}

	// check for ewp
	for _, query := range segmentQuery.EventsWithProperties {
		if _, exists := payloadEwp[query.Name]; !exists {
			additionalFiltersExist = true
			break
		}
		filterExists := false
		for _, mapEwpArr := range payloadEwp[query.Name] {
			if reflect.DeepEqual(query, mapEwpArr) {
				filterExists = true
				break
			}
		}
		if !filterExists {
			additionalFiltersExist = true
			break
		}
	}

	return additionalFiltersExist
}
