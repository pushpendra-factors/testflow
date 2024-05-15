package tests

import (
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateMarkerRunSegment(t *testing.T) {

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	// Create Domain Group
	domaindGroup, status := store.GetStore().CreateOrGetDomainsGroup(project.ID)
	assert.Equal(t, http.StatusCreated, status)
	assert.NotNil(t, domaindGroup)

	// Create CRM Groups
	hubspotGroup, status := store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_HUBSPOT_COMPANY, model.AllowedGroupNames)
	assert.Equal(t, http.StatusCreated, status)
	assert.NotNil(t, hubspotGroup)
	salesforceGroup, status := store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_SALESFORCE_ACCOUNT, model.AllowedGroupNames)
	assert.NotNil(t, salesforceGroup)
	assert.Equal(t, http.StatusCreated, status)
	sixSignalGroup, status := store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_LINKEDIN_COMPANY, model.AllowedGroupNames)
	assert.NotNil(t, sixSignalGroup)
	assert.Equal(t, http.StatusCreated, status)

	updateTime := U.TimeNowZ().AddDate(0, 0, -1)

	// updating marker_last_run_all_accounts in project settings to one day before
	errCode := store.GetStore().UpdateSegmentMarkerLastRunForAllAccounts(project.ID, updateTime)
	assert.Equal(t, http.StatusAccepted, errCode)

	// fetch all segments
	allSegments, status := store.GetStore().GetAllSegments(project.ID)
	assert.Equal(t, status, http.StatusFound)

	assert.Equal(t, 4, len(allSegments[model.GROUP_NAME_DOMAINS]))

	var segmentIds []string

	for index, segment := range allSegments[model.GROUP_NAME_DOMAINS] {
		if index == 3 {
			break
		}

		segmentIds = append(segmentIds, segment.Id)
	}

	// updating marker_run_segment in segments table for selected segments
	errCode = store.GetStore().UpdateMarkerRunSegment(project.ID, segmentIds, updateTime)
	assert.Equal(t, http.StatusAccepted, errCode)

	// fetch newly updated/created segments
	segmentsMap, err := store.GetStore().GetNewlyUpdatedAndCreatedSegments()
	assert.Nil(t, err)
	assert.Equal(t, len(allSegments[model.GROUP_NAME_DOMAINS]), len(segmentsMap[project.ID]))

	// fetch segments with given segment ids
	givenSegmentsMap, status := store.GetStore().GetSegmentByGivenIds(project.ID, segmentIds)
	assert.Equal(t, status, http.StatusFound)
	assert.Equal(t, 3, len(givenSegmentsMap[model.GROUP_NAME_DOMAINS]))

	for _, segment := range givenSegmentsMap[model.GROUP_NAME_DOMAINS] {
		assert.Equal(t, updateTime, segment.MarkerRunSegment.UTC())
	}
}
