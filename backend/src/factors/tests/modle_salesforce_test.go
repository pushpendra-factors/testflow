package tests

import (
	"encoding/json"
	IntSalesforce "factors/integration/salesforce"
	M "factors/model"
	U "factors/util"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
)

func TestCreateSalesforceDocument(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	refreshToken := U.RandomLowerAphaNumString(5)
	instancUrl := U.RandomLowerAphaNumString(5)
	errCode := M.UpdateAgentIntSalesforce(agent.UUID,
		refreshToken,
		instancUrl,
	)
	assert.Equal(t, http.StatusAccepted, errCode)

	_, errCode = M.UpdateProjectSettings(project.ID, &M.ProjectSetting{
		IntSalesforceEnabledAgentUUID: &agent.UUID,
	})
	assert.Equal(t, http.StatusAccepted, errCode)

	//should return list of supported doc type with timestamp 0
	syncInfo, status := M.GetSalesforceSyncInfo()
	assert.Equal(t, http.StatusFound, status)

	assert.Equal(t, refreshToken, syncInfo.ProjectSettings[project.ID].RefreshToken)
	assert.Equal(t, instancUrl, syncInfo.ProjectSettings[project.ID].InstanceURL)

	assert.Contains(t, syncInfo.LastSyncInfo[project.ID], M.SalesforceDocumentTypeNameContact)
	assert.Equal(t, int64(0), syncInfo.LastSyncInfo[project.ID][M.SalesforceDocumentTypeNameContact])
	assert.Contains(t, syncInfo.LastSyncInfo[project.ID], M.SalesforceDocumentTypeNameAccount)
	assert.Equal(t, int64(0), syncInfo.LastSyncInfo[project.ID][M.SalesforceDocumentTypeNameAccount])
	assert.Contains(t, syncInfo.LastSyncInfo[project.ID], M.SalesforceDocumentTypeNameLead)
	assert.Equal(t, int64(0), syncInfo.LastSyncInfo[project.ID][M.SalesforceDocumentTypeNameLead])

	//should not contain opportunity by default
	assert.NotContains(t, syncInfo.LastSyncInfo[project.ID], M.SalesforceDocumentTypeNameOpportunity)

	contactId := U.RandomLowerAphaNumString(5)
	name := U.RandomLowerAphaNumString(5)
	tm := time.Now()
	createdDate := tm.UTC().Format(M.SalesforceDocumentTimeLayout)

	// salesforce record with created == updated
	jsonData := fmt.Sprintf(`{"Id":"%s", "name":"%s","CreatedDate":"%s", "LastModifiedDate":"%s"}`, contactId, name, createdDate, createdDate)
	salesforceDocument := &M.SalesforceDocument{
		ProjectID: project.ID,
		TypeAlias: M.SalesforceDocumentTypeNameContact,
		Value:     &postgres.Jsonb{RawMessage: json.RawMessage([]byte(jsonData))},
	}

	status = M.CreateSalesforceDocument(project.ID, salesforceDocument)
	assert.Equal(t, http.StatusCreated, status)
	syncInfo, status = M.GetSalesforceSyncInfo()

	//should return latest timestamp from the databse
	assert.Equal(t, tm.Unix(), syncInfo.LastSyncInfo[project.ID][M.SalesforceDocumentTypeNameContact])

	//should return error on duplicate
	status = M.CreateSalesforceDocument(project.ID, salesforceDocument)
	assert.Equal(t, http.StatusConflict, status)

	//enrich job, create contact created and contact updated event
	enrichStatus := IntSalesforce.Enrich(project.ID)
	assert.Equal(t, project.ID, enrichStatus[0].ProjectID)
	assert.Equal(t, "success", enrichStatus[0].Status)
	assert.Equal(t, "success", enrichStatus[1].Status)
	assert.Equal(t, "success", enrichStatus[2].Status)

	eventName1 := fmt.Sprintf("$sf_%s_created", salesforceDocument.TypeAlias)
	eventName2 := fmt.Sprintf("$sf_%s_updated", salesforceDocument.TypeAlias)
	query := M.Query{
		From: tm.Unix() - 500,
		To:   tm.Unix() + 500,
		EventsWithProperties: []M.QueryEventWithProperties{
			M.QueryEventWithProperties{
				Name:       eventName1,
				Properties: []M.QueryProperty{},
			},
			M.QueryEventWithProperties{
				Name:       eventName2,
				Properties: []M.QueryProperty{},
			},
		},
		Class: M.QueryClassInsights,

		Type:            M.QueryTypeEventsOccurrence,
		EventsCondition: M.EventCondAllGivenEvent,
	}

	// test using query
	result, errCode, _ := M.Analyze(project.ID, query)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, eventName1, result.Rows[0][0])
	assert.Equal(t, int64(1), result.Rows[0][1])
	assert.Equal(t, eventName2, result.Rows[1][0])
	assert.Equal(t, int64(1), result.Rows[1][1])
}
