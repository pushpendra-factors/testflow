package tests

import (
	"factors/model/model"
	"testing"

	T "factors/task"

	"github.com/stretchr/testify/assert"
)

func TestMarketoSyncLogic(t *testing.T) {
	project, _, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	projectId := project.ID
	docType := model.MARKETO_TYPE_NAME_PROGRAM_MEMBERSHIP
	queryResult := make([][]string, 0)
	columnNamesFromMetadata := make([]string, 0)
	columnNamesFromMetadataDateTime := make(map[string]bool)
	columnNamesFromMetadataNumerical := make(map[string]bool)

	// Program Membership tests
	// program_membership_created with leadId, progId
	/* will enable these when enabling program membership
	queryResult = append(queryResult, []string{"1", "1", "", "", "2022-01-02 15:04:05 +0000 UTC", "", "Invited",
		"", "", "", "", "", "", "", "ProgName", "", "", "", "", "", "", ""})
	suc, fail := T.InsertIntegrationDocument(projectId, docType, queryResult, columnNamesFromMetadata, columnNamesFromMetadataDateTime, columnNamesFromMetadataNumerical)
	assert.Equal(t, suc, 1)
	assert.Equal(t, fail, 0)
	// program_membership_created with leadId, progId1
	queryResult = make([][]string, 0)
	queryResult = append(queryResult, []string{"1", "2", "", "", "2022-01-02 15:04:05 +0000 UTC", "", "Invited",
		"", "", "", "", "", "", "", "ProgName", "", "", "", "", "", "", ""})
	suc, fail = T.InsertIntegrationDocument(projectId, docType, queryResult, columnNamesFromMetadata, columnNamesFromMetadataDateTime, columnNamesFromMetadataNumerical)
	assert.Equal(t, suc, 1)
	assert.Equal(t, fail, 0)
	// program_membership_created with leadId1, progId
	queryResult = make([][]string, 0)
	queryResult = append(queryResult, []string{"2", "1", "", "", "2022-01-02 15:04:05 +0000 UTC", "", "Invited",
		"", "", "", "", "", "", "", "ProgName", "", "", "", "", "", "", ""})
	suc, fail = T.InsertIntegrationDocument(projectId, docType, queryResult, columnNamesFromMetadata, columnNamesFromMetadataDateTime, columnNamesFromMetadataNumerical)
	assert.Equal(t, suc, 1)
	assert.Equal(t, fail, 0)

	// program_membership_created with leadId, progId different timestamp
	queryResult = make([][]string, 0)
	queryResult = append(queryResult, []string{"1", "1", "", "", "2022-01-03 15:04:05 +0000 UTC", "", "Invited",
		"", "", "", "", "", "", "", "ProgName", "", "", "", "", "", "", ""})
	suc, fail = T.InsertIntegrationDocument(projectId, docType, queryResult, columnNamesFromMetadata, columnNamesFromMetadataDateTime, columnNamesFromMetadataNumerical)
	assert.Equal(t, suc, 1)
	assert.Equal(t, fail, 0)
	// program_membership_updated with leadId, progId same timestamp
	queryResult = make([][]string, 0)
	queryResult = append(queryResult, []string{"1", "1", "", "", "2022-01-03 15:04:05 +0000 UTC", "", "Invited",
		"", "", "", "", "", "", "", "ProgName", "", "", "", "", "", "", ""})
	suc, fail = T.InsertIntegrationDocument(projectId, docType, queryResult, columnNamesFromMetadata, columnNamesFromMetadataDateTime, columnNamesFromMetadataNumerical)
	assert.Equal(t, suc, 1)
	assert.Equal(t, fail, 0)
	// program_membership_updated with leadId, progId same timestamp - error case - conflict
	queryResult = make([][]string, 0)
	queryResult = append(queryResult, []string{"1", "1", "", "", "2022-01-03 15:04:05 +0000 UTC", "", "Invited",
		"", "", "", "", "", "", "", "ProgName", "", "", "", "", "", "", ""})
	suc, fail = T.InsertIntegrationDocument(projectId, docType, queryResult, columnNamesFromMetadata, columnNamesFromMetadataDateTime, columnNamesFromMetadataNumerical)
	assert.Equal(t, suc, 0)
	assert.Equal(t, fail, 1)
	*/

	// LEAD tests
	// first lead/ created_at = updated_at
	columnNamesFromMetadata = []string{"id", "email", "phone", "prop1", "prop2", "created_at", "updated_at"}
	columnNamesFromMetadataDateTime["created_at"] = true
	columnNamesFromMetadataDateTime["updated_at"] = true
	docType = model.MARKETO_TYPE_NAME_LEAD
	queryResult = make([][]string, 0)
	queryResult = append(queryResult, []string{"", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "leadId", "email", "phone",
		"v1", "v2", "2022-01-03 15:04:05 +0000 UTC", "2022-01-03 15:04:05 +0000 UTC"})
	suc, fail := T.InsertIntegrationDocument(projectId, docType, queryResult, columnNamesFromMetadata, columnNamesFromMetadataDateTime, columnNamesFromMetadataNumerical)
	assert.Equal(t, suc, 1)
	assert.Equal(t, fail, 0)

	// first lead/ created_at != updated_at
	columnNamesFromMetadata = []string{"id", "email", "phone", "prop1", "prop2", "created_at", "updated_at"}
	columnNamesFromMetadataDateTime["created_at"] = true
	columnNamesFromMetadataDateTime["updated_at"] = true
	docType = model.MARKETO_TYPE_NAME_LEAD
	queryResult = make([][]string, 0)
	queryResult = append(queryResult, []string{"", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "leadId1", "email", "phone",
		"v1", "v2", "2022-01-03 15:04:05 +0000 UTC", "2022-01-04 15:04:05 +0000 UTC"})
	suc, fail = T.InsertIntegrationDocument(projectId, docType, queryResult, columnNamesFromMetadata, columnNamesFromMetadataDateTime, columnNamesFromMetadataNumerical)
	assert.Equal(t, suc, 1)
	assert.Equal(t, fail, 0)

	// same lead id but different updated_at timestamp
	columnNamesFromMetadata = []string{"id", "email", "phone", "prop1", "prop2", "created_at", "updated_at"}
	columnNamesFromMetadataDateTime["created_at"] = true
	columnNamesFromMetadataDateTime["updated_at"] = true
	docType = model.MARKETO_TYPE_NAME_LEAD
	queryResult = make([][]string, 0)
	queryResult = append(queryResult, []string{"", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "leadId1", "email", "phone",
		"v1", "v2", "2022-01-03 15:04:05 +0000 UTC", "2022-01-04 15:04:06 +0000 UTC"})
	suc, fail = T.InsertIntegrationDocument(projectId, docType, queryResult, columnNamesFromMetadata, columnNamesFromMetadataDateTime, columnNamesFromMetadataNumerical)
	assert.Equal(t, suc, 1)
	assert.Equal(t, fail, 0)

	// same lead id and same updated_at timestamp
	columnNamesFromMetadata = []string{"id", "email", "phone", "prop1", "prop2", "created_at", "updated_at"}
	columnNamesFromMetadataDateTime["created_at"] = true
	columnNamesFromMetadataDateTime["updated_at"] = true
	docType = model.MARKETO_TYPE_NAME_LEAD
	queryResult = make([][]string, 0)
	queryResult = append(queryResult, []string{"", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "leadId1", "email", "phone",
		"v1", "v2", "2022-01-03 15:04:05 +0000 UTC", "2022-01-04 15:04:06 +0000 UTC"})
	suc, fail = T.InsertIntegrationDocument(projectId, docType, queryResult, columnNamesFromMetadata, columnNamesFromMetadataDateTime, columnNamesFromMetadataNumerical)
	assert.Equal(t, suc, 0)
	assert.Equal(t, fail, 1)

	columnNamesFromMetadataDateTime = make(map[string]bool)
	columnNamesFromMetadataDateTime["date1"] = true
	columnNamesFromMetadataDateTime["date2"] = true
	columnNamesFromMetadataNumerical = make(map[string]bool)
	columnNamesFromMetadataNumerical["num1"] = true
	columnNamesFromMetadataNumerical["num2"] = true

	suc, fail = T.InsertPropertyDataTypes(columnNamesFromMetadataDateTime, columnNamesFromMetadataNumerical, docType, projectId)
	assert.Equal(t, suc, 7)
	assert.Equal(t, fail, 0)

	suc, fail = T.InsertPropertyDataTypes(columnNamesFromMetadataDateTime, columnNamesFromMetadataNumerical, docType, projectId)
	assert.Equal(t, suc, 7)
	assert.Equal(t, fail, 0)

}
