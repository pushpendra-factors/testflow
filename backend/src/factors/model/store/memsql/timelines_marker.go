package memsql

import (
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"net/http"
	"reflect"
	"time"

	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) GetMarkedDomainsListByProjectId(projectID int64, payload model.TimelinePayload, downloadLimitGiven bool) ([]model.Profile, int, string) {
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

	// Redirect to old flow if segment_id is empty
	if payload.SegmentId == "" {
		return store.GetProfilesListByProjectId(projectID, payload, model.PROFILE_TYPE_ACCOUNT, downloadLimitGiven)
	}

	// Redirect to old flow if additional filters exist
	segment, statusCode := store.GetSegmentById(projectID, payload.SegmentId)
	if statusCode != http.StatusFound {
		logCtx.Error("Segment not found.")
		return []model.Profile{}, statusCode, "Failed to get segment"
	}

	lastRunTime, lastRunStatusCode := store.GetMarkerLastForAllAccounts(projectID)

	// for case - segment is updated but all_run for the day is yet to run
	if lastRunStatusCode != http.StatusFound || segment.UpdatedAt.After(lastRunTime) {
		return store.GetProfilesListByProjectId(projectID, payload, model.PROFILE_TYPE_ACCOUNT, downloadLimitGiven)
	}

	// if segment.UpdatedAt
	segmentQuery := model.Query{}
	err := U.DecodePostgresJsonbToStructType(segment.Query, &segmentQuery)
	if err != nil {
		log.WithField("project_id", projectID).Error("Unable to decode segment query")
		return []model.Profile{}, statusCode, "Failed to get segment"
	}

	additionalFiltersExist := CompareFilters(segmentQuery, payload.Query)

	if additionalFiltersExist {
		return store.GetProfilesListByProjectId(projectID, payload, model.PROFILE_TYPE_ACCOUNT, downloadLimitGiven)
	}

	// set Query Timezone
	timezoneString, statusCode := store.GetTimezoneForProject(projectID)
	if statusCode != http.StatusFound {
		logCtx.Error("Query failed. Failed to get Timezone.")
		return []model.Profile{}, statusCode, "Failed to fetch project timezone."
	}
	payload.Query.Timezone = string(timezoneString)

	domainGroup, errCode := store.GetGroup(projectID, model.GROUP_NAME_DOMAINS)
	if errCode != http.StatusFound || domainGroup == nil {
		return []model.Profile{}, http.StatusInternalServerError, "failed to get group while adding group info"
	}

	// check for search filter
	whereForSearchFilters, searchFiltersParams := SearchFilterForAllAccounts(payload.SearchFilter, domainGroup.ID)

	var domLimit int
	if downloadLimitGiven {
		domLimit = 10000
	} else {
		domLimit = 1000
	}

	query := fmt.Sprintf(`WITH step_0 as (
		SELECT 
		  id as identity, 
		  group_%d_id as host_name 
		FROM 
		  users 
		WHERE 
		  project_id = ? 
		  AND JSON_EXTRACT_STRING(
			associated_segments, ?
		  ) IS NOT NULL 
		  AND is_group_user = 1 
		  AND source = ? 
		  AND group_%d_id IS NOT NULL %s
		LIMIT 
		  100000
	  ) 
	  SELECT 
		identity, 
		host_name, 
		MAX(users.last_event_at) as last_activity 
	  FROM 
		step_0 
		JOIN (
		  SELECT 
			last_event_at, 
			group_%d_user_id 
		  FROM 
			users 
		  WHERE 
			users.project_id = ? 
			AND users.source != ? 
			AND last_event_at IS NOT NULL
		) AS users ON step_0.identity = users.group_%d_user_id 
	  GROUP BY 
		identity 
	  ORDER BY 
		last_activity DESC 
	  LIMIT 
		%d;`, domainGroup.ID, domainGroup.ID, whereForSearchFilters,
		domainGroup.ID, domainGroup.ID, domLimit)

	params := []interface{}{projectID, payload.SegmentId, model.UserSourceDomains}

	if whereForSearchFilters != "" {
		params = append(params, searchFiltersParams...)
	}

	params = append(params, projectID, model.UserSourceDomains)

	var profiles []model.Profile
	db := C.GetServices().Db
	err = db.Raw(query, params...).Scan(&profiles).Error
	if err != nil {
		return []model.Profile{}, http.StatusInternalServerError, "Error in fetching rows for all accounts"
	}

	// Redirect to old flow if no profiles found
	if len(profiles) == 0 {
		return store.GetProfilesListByProjectId(projectID, payload, model.PROFILE_TYPE_ACCOUNT, downloadLimitGiven)
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
