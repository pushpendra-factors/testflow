package tests

import (
	"encoding/json"
	"errors"
	C "factors/config"
	H "factors/handler"
	"factors/handler/helpers"
	IntSalesforce "factors/integration/salesforce"
	"factors/model/model"
	"factors/model/store"
	SDK "factors/sdk"
	"factors/task/event_user_cache"
	"factors/util"
	U "factors/util"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestSalesforceCreateSalesforceDocument(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	refreshToken := U.RandomLowerAphaNumString(5)
	instanceURL := U.RandomLowerAphaNumString(5)
	errCode := store.GetStore().UpdateAgentIntSalesforce(agent.UUID,
		refreshToken,
		instanceURL,
	)
	assert.Equal(t, http.StatusAccepted, errCode)

	_, errCode = store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		IntSalesforceEnabledAgentUUID: &agent.UUID,
	})
	assert.Equal(t, http.StatusAccepted, errCode)

	// should return list of supported doc type with timestamp 0.
	syncInfo, status := store.GetStore().GetSalesforceSyncInfo()
	assert.Equal(t, http.StatusFound, status)

	assert.Equal(t, refreshToken, syncInfo.ProjectSettings[project.ID].RefreshToken)
	assert.Equal(t, instanceURL, syncInfo.ProjectSettings[project.ID].InstanceURL)

	assert.Contains(t, syncInfo.LastSyncInfo[project.ID], model.SalesforceDocumentTypeNameContact)
	assert.Equal(t, int64(0), syncInfo.LastSyncInfo[project.ID][model.SalesforceDocumentTypeNameContact])
	assert.Contains(t, syncInfo.LastSyncInfo[project.ID], model.SalesforceDocumentTypeNameAccount)
	assert.Equal(t, int64(0), syncInfo.LastSyncInfo[project.ID][model.SalesforceDocumentTypeNameAccount])
	assert.Contains(t, syncInfo.LastSyncInfo[project.ID], model.SalesforceDocumentTypeNameLead)
	assert.Equal(t, int64(0), syncInfo.LastSyncInfo[project.ID][model.SalesforceDocumentTypeNameLead])

	// should contain opportunity by default.
	assert.Contains(t, syncInfo.LastSyncInfo[project.ID], model.SalesforceDocumentTypeNameOpportunity)

	contactID := U.RandomLowerAphaNumString(5)
	name := U.RandomLowerAphaNumString(5)

	createdDate := time.Now()

	// salesforce record with created == updated.
	jsonData := fmt.Sprintf(`{"Id":"%s", "name":"%s","CreatedDate":"%s", "LastModifiedDate":"%s"}`, contactID, name, createdDate.UTC().Format(model.SalesforceDocumentDateTimeLayout), createdDate.UTC().Format(model.SalesforceDocumentDateTimeLayout))
	salesforceDocument := &model.SalesforceDocument{
		ProjectID: project.ID,
		TypeAlias: model.SalesforceDocumentTypeNameContact,
		Value:     &postgres.Jsonb{RawMessage: json.RawMessage([]byte(jsonData))},
	}

	status = store.GetStore().CreateSalesforceDocument(project.ID, salesforceDocument)
	assert.Equal(t, http.StatusCreated, status)
	syncInfo, status = store.GetStore().GetSalesforceSyncInfo()

	// should return the latest timestamp from the database.
	assert.Equal(t, createdDate.Unix(), syncInfo.LastSyncInfo[project.ID][model.SalesforceDocumentTypeNameContact])

	// should return error on duplicate.
	status = store.GetStore().CreateSalesforceDocument(project.ID, salesforceDocument)
	assert.Equal(t, http.StatusConflict, status)

	// enrich job, create contact created and contact updated event.
	enrichStatus, _ := IntSalesforce.Enrich(project.ID)
	assert.Equal(t, project.ID, enrichStatus[0].ProjectID)
	assert.Len(t, enrichStatus, 1)
	assert.Equal(t, "success", enrichStatus[0].Status)

	eventNameCreated := fmt.Sprintf("$sf_%s_created", salesforceDocument.TypeAlias)
	eventNameUpdate := fmt.Sprintf("$sf_%s_updated", salesforceDocument.TypeAlias)
	query := model.Query{
		From: createdDate.Unix() - 500,
		To:   createdDate.Unix() + 500,
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       eventNameCreated,
				Properties: []model.QueryProperty{},
			},
			{
				Name:       eventNameUpdate,
				Properties: []model.QueryProperty{},
			},
		},
		Class: model.QueryClassInsights,

		Type:            model.QueryTypeEventsOccurrence,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	// test using query
	result, errCode, _ := store.GetStore().Analyze(project.ID, query)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, eventNameCreated, result.Rows[0][0])
	assert.Equal(t, float64(1), result.Rows[0][1])
	assert.Equal(t, eventNameUpdate, result.Rows[1][0])
	assert.Equal(t, float64(1), result.Rows[1][1])

	query = model.Query{
		From: createdDate.Unix() - 500,
		To:   createdDate.Unix() + 500,
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       eventNameCreated,
				Properties: []model.QueryProperty{},
			},
			{
				Name:       eventNameUpdate,
				Properties: []model.QueryProperty{},
			},
		},
		Class: model.QueryClassInsights,

		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	// test using query
	result, errCode, _ = store.GetStore().Analyze(project.ID, query)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, float64(1), result.Rows[0][0])

	/*
		salesforce record1 with createdDate != updatedDate
		salesforce record2 with createdDate != updatedDate
		both same id
	*/
	contactID = U.RandomLowerAphaNumString(5)
	name = U.RandomLowerAphaNumString(5)
	createdDate = createdDate.AddDate(0, 0, -10)
	updatedDate := createdDate.AddDate(0, 0, 1)

	// salesforce record1 with created != updated
	jsonData = fmt.Sprintf(`{"Id":"%s", "name":"%s","MobilePhone":1234567890,"CreatedDate":"%s", "LastModifiedDate":"%s"}`, contactID, name, createdDate.UTC().Format(model.SalesforceDocumentDateTimeLayout), updatedDate.UTC().Format(model.SalesforceDocumentDateTimeLayout))
	salesforceDocument = &model.SalesforceDocument{
		ProjectID: project.ID,
		TypeAlias: model.SalesforceDocumentTypeNameContact,
		Value:     &postgres.Jsonb{RawMessage: json.RawMessage([]byte(jsonData))},
	}

	status = store.GetStore().CreateSalesforceDocument(project.ID, salesforceDocument)
	assert.Equal(t, http.StatusCreated, status)

	// salesforce record2 with created != updated same user
	updatedDate = updatedDate.AddDate(0, 0, 1)
	jsonData = fmt.Sprintf(`{"Id":"%s", "name":"%s","CreatedDate":"%s", "LastModifiedDate":"%s"}`, contactID, name, createdDate.UTC().Format(model.SalesforceDocumentDateTimeLayout), updatedDate.Format(model.SalesforceDocumentDateTimeLayout))
	salesforceDocument = &model.SalesforceDocument{
		ProjectID: project.ID,
		TypeAlias: model.SalesforceDocumentTypeNameContact,
		Value:     &postgres.Jsonb{RawMessage: json.RawMessage([]byte(jsonData))},
	}
	status = store.GetStore().CreateSalesforceDocument(project.ID, salesforceDocument)
	assert.Equal(t, http.StatusCreated, status)

	//should return conflict on duplicate
	status = store.GetStore().CreateSalesforceDocument(project.ID, salesforceDocument)
	assert.Equal(t, http.StatusConflict, status)

	//enrich job, create contact created and contact updated event
	enrichStatus, _ = IntSalesforce.Enrich(project.ID)
	assert.Equal(t, project.ID, enrichStatus[0].ProjectID)
	assert.Len(t, enrichStatus, 1)
	assert.Equal(t, "success", enrichStatus[0].Status)

	// query count of unique users
	query = model.Query{
		From: createdDate.Unix() - 500,
		To:   updatedDate.Unix() + 500,
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       eventNameCreated,
				Properties: []model.QueryProperty{},
			},
			{
				Name:       eventNameUpdate,
				Properties: []model.QueryProperty{},
			},
		},
		Class: model.QueryClassInsights,

		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}
	// test using query
	result, errCode, _ = store.GetStore().Analyze(project.ID, query)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, float64(1), result.Rows[0][0])

	// query count of events
	query = model.Query{
		From: createdDate.Unix() - 500,
		To:   updatedDate.Unix() + 500,
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       eventNameCreated,
				Properties: []model.QueryProperty{},
			},
			{
				Name:       eventNameUpdate,
				Properties: []model.QueryProperty{},
			},
		},
		Class: model.QueryClassInsights,

		Type:            model.QueryTypeEventsOccurrence,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	// test using query
	result, errCode, _ = store.GetStore().Analyze(project.ID, query)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, eventNameCreated, result.Rows[0][0])
	assert.Equal(t, float64(1), result.Rows[0][1])
	assert.Equal(t, eventNameUpdate, result.Rows[1][0])
	assert.Equal(t, float64(3), result.Rows[1][1])

	query.GroupByProperties = []model.QueryGroupByProperty{
		{
			Entity:    model.PropertyEntityUser,
			Property:  "$user_id",
			EventName: model.UserPropertyGroupByPresent,
		},
	}

	result, errCode, _ = store.GetStore().Analyze(project.ID, query)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, eventNameCreated, result.Rows[0][0])
	assert.Equal(t, "1234567890", result.Rows[0][1])
	assert.Equal(t, eventNameUpdate, result.Rows[1][0])
	assert.Equal(t, "1234567890", result.Rows[1][1])
}

func TestSalesforceCRMSmartEvent(t *testing.T) {
	project, _, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	contactID := U.RandomLowerAphaNumString(5)
	userID1 := U.RandomLowerAphaNumString(5)
	userID2 := U.RandomLowerAphaNumString(5)
	userID3 := U.RandomLowerAphaNumString(5)
	cuid := U.RandomLowerAphaNumString(5)
	_, status := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, ID: userID1, CustomerUserId: cuid, Source: model.GetRequestSourcePointer(model.UserSourceSalesforce)})
	assert.Equal(t, http.StatusCreated, status)
	_, status = store.GetStore().CreateUser(&model.User{ProjectId: project.ID, ID: userID2, CustomerUserId: cuid, Source: model.GetRequestSourcePointer(model.UserSourceSalesforce)})
	assert.Equal(t, http.StatusCreated, status)
	_, status = store.GetStore().CreateUser(&model.User{ProjectId: project.ID, ID: userID3, CustomerUserId: cuid, Source: model.GetRequestSourcePointer(model.UserSourceSalesforce)})
	assert.Equal(t, http.StatusCreated, status)

	createdAt := time.Now().AddDate(0, 0, -11)
	updatedDate := createdAt.AddDate(0, 0, -11)
	propertyDay := "Sunday"
	jsonData := fmt.Sprintf(`{"Id":"%s", "day":"%s","CreatedDate":"%s", "LastModifiedDate":"%s"}`, contactID, propertyDay, createdAt.UTC().Format(model.SalesforceDocumentDateTimeLayout), updatedDate.Format(model.SalesforceDocumentDateTimeLayout))
	salesforceDocument := &model.SalesforceDocument{
		ProjectID: project.ID,
		TypeAlias: model.SalesforceDocumentTypeNameContact,
		Value:     &postgres.Jsonb{RawMessage: json.RawMessage([]byte(jsonData))},
	}
	status = store.GetStore().CreateSalesforceDocument(project.ID, salesforceDocument)
	assert.Equal(t, http.StatusCreated, status)
	status = store.GetStore().UpdateSalesforceDocumentBySyncStatus(project.ID, salesforceDocument, "", userID3, "", true)
	assert.Equal(t, http.StatusAccepted, status)

	createdAt = time.Now().AddDate(0, 0, -11)
	updatedDate = createdAt.AddDate(0, 0, -10)
	propertyDay = "Monday"
	jsonData = fmt.Sprintf(`{"Id":"%s", "day":"%s","CreatedDate":"%s", "LastModifiedDate":"%s"}`, contactID, propertyDay, createdAt.UTC().Format(model.SalesforceDocumentDateTimeLayout), updatedDate.Format(model.SalesforceDocumentDateTimeLayout))
	salesforceDocument = &model.SalesforceDocument{
		ProjectID: project.ID,
		TypeAlias: model.SalesforceDocumentTypeNameContact,
		Value:     &postgres.Jsonb{RawMessage: json.RawMessage([]byte(jsonData))},
	}
	status = store.GetStore().CreateSalesforceDocument(project.ID, salesforceDocument)
	assert.Equal(t, http.StatusCreated, status)
	status = store.GetStore().UpdateSalesforceDocumentBySyncStatus(project.ID, salesforceDocument, "", userID1, "", true)
	assert.Equal(t, http.StatusAccepted, status)

	createdAt = time.Now().AddDate(0, 0, -11)
	updatedDate = createdAt.AddDate(0, 0, -9)
	propertyDay = "Tuesday"
	jsonData = fmt.Sprintf(`{"Id":"%s", "day":"%s","CreatedDate":"%s", "LastModifiedDate":"%s"}`, contactID, propertyDay, createdAt.UTC().Format(model.SalesforceDocumentDateTimeLayout), updatedDate.Format(model.SalesforceDocumentDateTimeLayout))
	salesforceDocument = &model.SalesforceDocument{
		ProjectID: project.ID,
		TypeAlias: model.SalesforceDocumentTypeNameContact,
		Value:     &postgres.Jsonb{RawMessage: json.RawMessage([]byte(jsonData))},
	}
	status = store.GetStore().CreateSalesforceDocument(project.ID, salesforceDocument)
	assert.Equal(t, http.StatusCreated, status)
	status = store.GetStore().UpdateSalesforceDocumentBySyncStatus(project.ID, salesforceDocument, "", userID2, "", true)
	assert.Equal(t, http.StatusAccepted, status)

	createdAt = time.Now().AddDate(0, 0, -11)
	updatedDate = createdAt.AddDate(0, 0, -8)
	propertyDay = "Wednesday"
	jsonData = fmt.Sprintf(`{"Id":"%s", "day":"%s","CreatedDate":"%s", "LastModifiedDate":"%s"}`, contactID, propertyDay, createdAt.UTC().Format(model.SalesforceDocumentDateTimeLayout), updatedDate.Format(model.SalesforceDocumentDateTimeLayout))
	salesforceDocument = &model.SalesforceDocument{
		ProjectID: project.ID,
		TypeAlias: model.SalesforceDocumentTypeNameContact,
		Value:     &postgres.Jsonb{RawMessage: json.RawMessage([]byte(jsonData))},
	}
	status = store.GetStore().CreateSalesforceDocument(project.ID, salesforceDocument)
	assert.Equal(t, http.StatusCreated, status)

	prevDoc, status := store.GetStore().GetLastSyncedSalesforceDocumentByCustomerUserIDORUserID(project.ID, cuid, userID3, salesforceDocument.Type)
	assert.Equal(t, http.StatusFound, status)
	_, prevProperties, err := IntSalesforce.GetSalesforceDocumentProperties(project.ID, prevDoc)
	assert.Nil(t, err)
	assert.Equal(t, "Tuesday", (*prevProperties)["day"])

	prevDoc, status = store.GetStore().GetLastSyncedSalesforceDocumentByCustomerUserIDORUserID(project.ID, "", userID3, salesforceDocument.Type)
	assert.Equal(t, http.StatusFound, status)
	_, prevProperties, err = IntSalesforce.GetSalesforceDocumentProperties(project.ID, prevDoc)
	assert.Nil(t, err)
	assert.Equal(t, "Sunday", (*prevProperties)["day"])

	filter := model.SmartCRMEventFilter{
		Source:               model.SmartCRMEventSourceSalesforce,
		ObjectType:           "contact",
		Description:          "salesforce contact",
		FilterEvaluationType: model.FilterEvaluationTypeSpecific,
		Filters: []model.PropertyFilter{
			{
				Name: "day",
				Rules: []model.CRMFilterRule{
					{
						PropertyState: model.CurrentState,
						Value:         "Saturday",
						Operator:      model.COMPARE_EQUAL,
					},
					{
						PropertyState: model.PreviousState,
						Value:         "Tuesday",
						Operator:      model.COMPARE_EQUAL,
					},
				},
				LogicalOp: model.LOGICAL_OP_AND,
			},
		},
		LogicalOp:               model.LOGICAL_OP_AND,
		TimestampReferenceField: "time",
	}

	currentProperties := make(map[string]interface{})
	currentProperties["day"] = "Saturday"
	smartEvent, _, ok := IntSalesforce.GetSalesforceSmartEventPayload(project.ID, "test", contactID, userID3, salesforceDocument.Type, &currentProperties, nil, &filter)
	assert.Equal(t, true, ok)
	assert.Equal(t, "test", smartEvent.Name)
	assert.Equal(t, "Tuesday", smartEvent.Properties["$prev_salesforce_contact_day"])
	assert.Equal(t, "Saturday", smartEvent.Properties["$curr_salesforce_contact_day"])

	filter = model.SmartCRMEventFilter{
		Source:               model.SmartCRMEventSourceSalesforce,
		ObjectType:           "contact",
		Description:          "salesforce contact",
		FilterEvaluationType: model.FilterEvaluationTypeSpecific,
		Filters: []model.PropertyFilter{
			{
				Name: "day",
				Rules: []model.CRMFilterRule{
					{
						PropertyState: model.CurrentState,
						Value:         "Saturday",
						Operator:      model.COMPARE_EQUAL,
					},
					{
						PropertyState: model.PreviousState,
						Value:         "Sunday",
						Operator:      model.COMPARE_EQUAL,
					},
				},
				LogicalOp: model.LOGICAL_OP_AND,
			},
		},
		LogicalOp:               model.LOGICAL_OP_AND,
		TimestampReferenceField: "time",
	}

	smartEvent, _, ok = IntSalesforce.GetSalesforceSmartEventPayload(project.ID, "test", contactID, userID3, salesforceDocument.Type, &currentProperties, nil, &filter)
	assert.Equal(t, false, ok)
}

func TestSalesforceLastSyncedDocument(t *testing.T) {
	project, _, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	contactID1 := U.RandomLowerAphaNumString(5)
	contactID2 := U.RandomLowerAphaNumString(5)
	contactID3 := U.RandomLowerAphaNumString(5)
	contactID4 := U.RandomLowerAphaNumString(5)
	contactID5 := U.RandomLowerAphaNumString(5)
	contactID6 := U.RandomLowerAphaNumString(5)

	userID1 := U.RandomLowerAphaNumString(5)
	userID2 := U.RandomLowerAphaNumString(5)
	userID3 := U.RandomLowerAphaNumString(5)
	userID4 := U.RandomLowerAphaNumString(5)
	userID5 := U.RandomLowerAphaNumString(5)
	userID6 := U.RandomLowerAphaNumString(5)
	_, status := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, ID: userID1, Source: model.GetRequestSourcePointer(model.UserSourceSalesforce)})
	assert.Equal(t, http.StatusCreated, status)
	_, status = store.GetStore().CreateUser(&model.User{ProjectId: project.ID, ID: userID2, Source: model.GetRequestSourcePointer(model.UserSourceSalesforce)})
	assert.Equal(t, http.StatusCreated, status)
	_, status = store.GetStore().CreateUser(&model.User{ProjectId: project.ID, ID: userID3, Source: model.GetRequestSourcePointer(model.UserSourceSalesforce)})
	assert.Equal(t, http.StatusCreated, status)
	_, status = store.GetStore().CreateUser(&model.User{ProjectId: project.ID, ID: userID4, Source: model.GetRequestSourcePointer(model.UserSourceSalesforce)})
	assert.Equal(t, http.StatusCreated, status)
	_, status = store.GetStore().CreateUser(&model.User{ProjectId: project.ID, ID: userID5, Source: model.GetRequestSourcePointer(model.UserSourceSalesforce)})
	assert.Equal(t, http.StatusCreated, status)
	_, status = store.GetStore().CreateUser(&model.User{ProjectId: project.ID, ID: userID6, Source: model.GetRequestSourcePointer(model.UserSourceSalesforce)})
	assert.Equal(t, http.StatusCreated, status)

	userIDs := []string{userID1, userID2, userID3, userID4, userID5, userID6}
	contactIDs := []string{contactID1, contactID2, contactID3, contactID4, contactID5, contactID6}
	characters := []string{"A", "B", "C", "D", "E", "F"}
	days := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Friday", "Saturday"}
	var createdAt time.Time
	var updatedDate time.Time

	/*
		Summary Of synced test document
		U1(day="Sunday", character="A", type = contact)  -> U1(day="Saturday", character="G", type = contact) -> U1(day="Friday", character="H", type = lead)
		U2(day="Monday", character="B",type = contact)
		U3(day="Tuesday", character="C",type = contact)
		U4(day="Wednesday", character="D",type = contact)
		U5(day="Friday", character="E",type = contact)
		U6(day="Saturday", character="F",type = contact)
	*/
	for i := 0; i < 6; i++ {
		createdAt = time.Now().AddDate(0, 0, -20+i)
		updatedDate = createdAt.AddDate(0, 0, -20+i)
		jsonData := fmt.Sprintf(`{"Id":"%s", "character":"%s","day":"%s","CreatedDate":"%s", "LastModifiedDate":"%s"}`, contactIDs[i], characters[i], days[i], createdAt.UTC().Format(model.SalesforceDocumentDateTimeLayout), updatedDate.Format(model.SalesforceDocumentDateTimeLayout))
		salesforceDocument := &model.SalesforceDocument{
			ProjectID: project.ID,
			TypeAlias: model.SalesforceDocumentTypeNameContact,
			Value:     &postgres.Jsonb{RawMessage: json.RawMessage([]byte(jsonData))},
		}
		status = store.GetStore().CreateSalesforceDocument(project.ID, salesforceDocument)
		assert.Equal(t, http.StatusCreated, status)
		status = store.GetStore().UpdateSalesforceDocumentBySyncStatus(project.ID, salesforceDocument, "", userIDs[i], "", true)
		assert.Equal(t, http.StatusAccepted, status)
	}

	updatedDate = updatedDate.AddDate(0, 0, -1)
	jsonData := fmt.Sprintf(`{"Id":"%s", "character":"%s","day":"%s","CreatedDate":"%s", "LastModifiedDate":"%s"}`, contactID1, "G", "Saturday", createdAt.UTC().Format(model.SalesforceDocumentDateTimeLayout), updatedDate.Format(model.SalesforceDocumentDateTimeLayout))
	salesforceDocument := &model.SalesforceDocument{
		ProjectID: project.ID,
		TypeAlias: model.SalesforceDocumentTypeNameContact,
		Value:     &postgres.Jsonb{RawMessage: json.RawMessage([]byte(jsonData))},
	}
	status = store.GetStore().CreateSalesforceDocument(project.ID, salesforceDocument)
	assert.Equal(t, http.StatusCreated, status)
	status = store.GetStore().UpdateSalesforceDocumentBySyncStatus(project.ID, salesforceDocument, "", userID1, "", true)
	assert.Equal(t, http.StatusAccepted, status)

	updatedDate = updatedDate.AddDate(0, 0, -1)
	leadID1 := U.RandomLowerAphaNumString(5)
	jsonData = fmt.Sprintf(`{"Id":"%s", "character":"%s","day":"%s","CreatedDate":"%s", "LastModifiedDate":"%s"}`, leadID1, "H", "Friday", createdAt.UTC().Format(model.SalesforceDocumentDateTimeLayout), updatedDate.Format(model.SalesforceDocumentDateTimeLayout))
	salesforceDocument = &model.SalesforceDocument{
		ProjectID: project.ID,
		TypeAlias: model.SalesforceDocumentTypeNameLead,
		Value:     &postgres.Jsonb{RawMessage: json.RawMessage([]byte(jsonData))},
	}
	status = store.GetStore().CreateSalesforceDocument(project.ID, salesforceDocument)
	assert.Equal(t, http.StatusCreated, status)
	status = store.GetStore().UpdateSalesforceDocumentBySyncStatus(project.ID, salesforceDocument, "", userID1, "", true)
	assert.Equal(t, http.StatusAccepted, status)

	/*
		Last synced document of U1 and type contact
		U1(day="Saturday", character="G", type = contact)
	*/
	prevDoc, status := store.GetStore().GetLastSyncedSalesforceDocumentByCustomerUserIDORUserID(project.ID, "", userID1, model.SalesforceDocumentTypeContact)
	assert.Equal(t, http.StatusFound, status)
	_, prevProperties, err := IntSalesforce.GetSalesforceDocumentProperties(project.ID, prevDoc)
	assert.Nil(t, err)
	assert.Equal(t, "G", (*prevProperties)["character"])
	assert.Equal(t, "Saturday", (*prevProperties)["day"])

	prevDocs, status := store.GetStore().GetSyncedSalesforceDocumentByType(project.ID, []string{contactID1}, model.SalesforceDocumentTypeContact, false)
	assert.Equal(t, http.StatusFound, status)
	_, prevProperties, err = IntSalesforce.GetSalesforceDocumentProperties(project.ID, &prevDocs[len(prevDocs)-1])
	assert.Nil(t, err)
	assert.Equal(t, "G", (*prevProperties)["character"])
	assert.Equal(t, "Saturday", (*prevProperties)["day"])

	/*
		filter1:
		prev_salesforce_contact_character = "G" AND curr_salesforce_contact_character ="H"
	*/
	var filters []model.SmartCRMEventFilter
	filter := model.SmartCRMEventFilter{
		Source:               model.SmartCRMEventSourceSalesforce,
		ObjectType:           "contact",
		Description:          "salesforce contact",
		FilterEvaluationType: model.FilterEvaluationTypeSpecific,
		Filters: []model.PropertyFilter{
			{
				Name: "character",
				Rules: []model.CRMFilterRule{
					{
						PropertyState: model.CurrentState,
						Value:         "H",
						Operator:      model.COMPARE_EQUAL,
					},
					{
						PropertyState: model.PreviousState,
						Value:         "G",
						Operator:      model.COMPARE_EQUAL,
					},
				},
				LogicalOp: model.LOGICAL_OP_AND,
			},
		},
		LogicalOp:               model.LOGICAL_OP_AND,
		TimestampReferenceField: "time",
	}

	filters = append(filters, filter)
	/*
		filter2:
		prev_salesforce_contact_character = "B" AND curr_salesforce_contact_character ="I"
	*/
	filter = model.SmartCRMEventFilter{
		Source:               model.SmartCRMEventSourceSalesforce,
		ObjectType:           "contact",
		Description:          "salesforce contact",
		FilterEvaluationType: model.FilterEvaluationTypeSpecific,
		Filters: []model.PropertyFilter{
			{
				Name: "character",
				Rules: []model.CRMFilterRule{
					{
						PropertyState: model.CurrentState,
						Value:         "I",
						Operator:      model.COMPARE_EQUAL,
					},
					{
						PropertyState: model.PreviousState,
						Value:         "B",
						Operator:      model.COMPARE_EQUAL,
					},
				},
				LogicalOp: model.LOGICAL_OP_AND,
			},
		},
		LogicalOp:               model.LOGICAL_OP_AND,
		TimestampReferenceField: "time",
	}
	filters = append(filters, filter)

	/*
		filter3:
		prev_salesforce_contact_day = "Sunday" AND curr_salesforce_contact_day ="Sunday"
	*/
	filter = model.SmartCRMEventFilter{
		Source:               model.SmartCRMEventSourceSalesforce,
		ObjectType:           "contact",
		Description:          "salesforce contact",
		FilterEvaluationType: model.FilterEvaluationTypeSpecific,
		Filters: []model.PropertyFilter{
			{
				Name: "day",
				Rules: []model.CRMFilterRule{
					{
						PropertyState: model.CurrentState,
						Value:         "Sunday",
						Operator:      model.COMPARE_EQUAL,
					},
					{
						PropertyState: model.PreviousState,
						Value:         "Saturday",
						Operator:      model.COMPARE_EQUAL,
					},
				},
				LogicalOp: model.LOGICAL_OP_AND,
			},
		},
		LogicalOp:               model.LOGICAL_OP_AND,
		TimestampReferenceField: "time",
	}
	filters = append(filters, filter)

	/*
		filter1(prev_salesforce_contact_character = "G" AND curr_salesforce_contact_character ="H")
		for U1
		New incoming record(salesforce_contact_character = "H")
		Expected previous record U1(day="Sunday", character="G", type = contact)
	*/
	currentProperties := make(map[string]interface{})
	currentProperties["character"] = "H"
	smartEvent, prevProperties, ok := IntSalesforce.GetSalesforceSmartEventPayload(project.ID, "filter1", contactID1, userID1, model.SalesforceDocumentTypeContact, &currentProperties, nil, &filters[0])
	assert.Equal(t, true, ok)
	assert.Equal(t, "filter1", smartEvent.Name)
	assert.Equal(t, "G", smartEvent.Properties["$prev_salesforce_contact_character"])
	assert.Equal(t, "H", smartEvent.Properties["$curr_salesforce_contact_character"])
	//prev properties check
	assert.Equal(t, "Saturday", (*prevProperties)["day"])
	assert.Equal(t, "G", (*prevProperties)["character"])

	//Fail Test
	currentProperties = make(map[string]interface{})
	currentProperties["character"] = "G"
	smartEvent, prevProperties, ok = IntSalesforce.GetSalesforceSmartEventPayload(project.ID, "filter1", "", userID1, model.SalesforceDocumentTypeContact, &currentProperties, nil, &filters[0])
	assert.Equal(t, false, ok)
	// prev properties should be nil
	assert.Nil(t, prevProperties)

	/*
		filter2(prev_salesforce_contact_character = "B" AND curr_salesforce_contact_character ="I")
		for U2
		New incoming record(salesforce_contact_character = "I")
		Expected previous record U2(day="Monday", character="B",type = contact)
	*/
	currentProperties = make(map[string]interface{})
	currentProperties["character"] = "I"
	smartEvent, prevProperties, ok = IntSalesforce.GetSalesforceSmartEventPayload(project.ID, "filter2", contactID2, userID2, model.SalesforceDocumentTypeContact, &currentProperties, nil, &filters[1])
	assert.Equal(t, true, ok)
	assert.Equal(t, "filter2", smartEvent.Name)
	assert.Equal(t, "B", smartEvent.Properties["$prev_salesforce_contact_character"])
	assert.Equal(t, "I", smartEvent.Properties["$curr_salesforce_contact_character"])
	// prev properties check
	assert.Equal(t, "B", (*prevProperties)["character"])
	assert.Equal(t, "Monday", (*prevProperties)["day"])

	//Fail Test filter2
	currentProperties["character"] = "J"
	smartEvent, prevProperties, ok = IntSalesforce.GetSalesforceSmartEventPayload(project.ID, "filter2", contactID2, userID2, model.SalesforceDocumentTypeContact, &currentProperties, nil, &filters[1])
	assert.Equal(t, false, ok)
	// prev properties should be nil
	assert.Nil(t, prevProperties)

	/*
		filter3(prev_salesforce_contact_day = "Sunday" AND curr_salesforce_contact_day ="Sunday")
		for U1
		New incoming record(salesforce_contact_day = "Sunday")
		Expected previous record U1(day="Saturday", character="G", type = contact)
	*/
	currentProperties = make(map[string]interface{})
	currentProperties["day"] = "Sunday"
	smartEvent, prevProperties, ok = IntSalesforce.GetSalesforceSmartEventPayload(project.ID, "filter3", contactID1, userID1, model.SalesforceDocumentTypeContact, &currentProperties, nil, &filters[2])
	assert.Equal(t, true, ok)
	assert.Equal(t, "filter3", smartEvent.Name)
	assert.Equal(t, "Saturday", smartEvent.Properties["$prev_salesforce_contact_day"])
	assert.Equal(t, "Sunday", smartEvent.Properties["$curr_salesforce_contact_day"])
	//prev properties check
	assert.Equal(t, "G", (*prevProperties)["character"])
	assert.Equal(t, "Saturday", (*prevProperties)["day"])

	//Fail Test filter2
	currentProperties = make(map[string]interface{})
	currentProperties["day"] = "Monday"
	smartEvent, prevProperties, ok = IntSalesforce.GetSalesforceSmartEventPayload(project.ID, "filter2", "", userID1, model.SalesforceDocumentTypeContact, &currentProperties, nil, &filters[2])
	assert.Equal(t, false, ok)
	// prev properties should be nil
	assert.Nil(t, prevProperties)
}

func TestSalesforceSameUserSmartEvent(t *testing.T) {

	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	r := gin.Default()
	H.InitAppRoutes(r)

	filter := model.SmartCRMEventFilter{
		Source:               model.SmartCRMEventSourceSalesforce,
		ObjectType:           "contact",
		Description:          "salesforce contact",
		FilterEvaluationType: model.FilterEvaluationTypeSpecific,
		Filters: []model.PropertyFilter{
			{
				Name: "character",
				Rules: []model.CRMFilterRule{
					{
						PropertyState: model.CurrentState,
						Value:         "I",
						Operator:      model.COMPARE_EQUAL,
					},
					{
						PropertyState: model.PreviousState,
						Value:         "B",
						Operator:      model.COMPARE_EQUAL,
					},
				},
				LogicalOp: model.LOGICAL_OP_AND,
			},
		},
		LogicalOp:               model.LOGICAL_OP_AND,
		TimestampReferenceField: model.TimestampReferenceTypeDocument,
	}

	smartEventName := "Event 1"
	requestPayload := make(map[string]interface{})
	requestPayload["name"] = smartEventName
	requestPayload["expr"] = filter

	w := sendCreateSmartEventFilterReq(r, project.ID, agent, &requestPayload)
	assert.Equal(t, http.StatusCreated, w.Code)
	var responsePayload H.APISmartEventFilterResponePayload
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	err = json.Unmarshal(jsonResponse, &responsePayload)
	assert.Nil(t, err)
	stringCompEventNameId := responsePayload.EventNameID
	assert.NotEqual(t, 0, stringCompEventNameId)

	contactID := U.RandomLowerAphaNumString(5)
	name := U.RandomLowerAphaNumString(5)

	createdDate := time.Now()

	jsonData := fmt.Sprintf(`{"Id":"%s", "name":"%s","CreatedDate":"%s", "LastModifiedDate":"%s","character":"B"}`, contactID, name, createdDate.UTC().Format(model.SalesforceDocumentDateTimeLayout), createdDate.UTC().Format(model.SalesforceDocumentDateTimeLayout))
	salesforceDocumentPrev := &model.SalesforceDocument{
		ProjectID: project.ID,
		TypeAlias: model.SalesforceDocumentTypeNameContact,
		Value:     &postgres.Jsonb{RawMessage: json.RawMessage([]byte(jsonData))},
	}
	status := store.GetStore().CreateSalesforceDocument(project.ID, salesforceDocumentPrev)
	assert.Equal(t, http.StatusCreated, status)

	userID1 := U.RandomLowerAphaNumString(5)
	_, status = store.GetStore().CreateUser(&model.User{ProjectId: project.ID, ID: userID1, Source: model.GetRequestSourcePointer(model.UserSourceSalesforce)})
	assert.Equal(t, http.StatusCreated, status)
	eventID1 := U.RandomLowerAphaNumString(10)
	store.GetStore().UpdateSalesforceDocumentBySyncStatus(project.ID, salesforceDocumentPrev, eventID1, userID1, "", true)

	currentProperties := make(map[string]interface{})
	currentProperties["character"] = "I"
	jsonData = fmt.Sprintf(`{"Id":"%s", "name":"%s","CreatedDate":"%s", "LastModifiedDate":"%s","character":"%s"}`, contactID, name, createdDate.UTC().Format(model.SalesforceDocumentDateTimeLayout), createdDate.AddDate(0, 0, 1).UTC().Format(model.SalesforceDocumentDateTimeLayout), currentProperties["$salesforce_contact_character"])
	currentSalesforceDocument := &model.SalesforceDocument{
		ProjectID: project.ID,
		Type:      model.SalesforceDocumentTypeContact,
		Value:     &postgres.Jsonb{RawMessage: json.RawMessage([]byte(jsonData))},
	}
	salesforceSmartEventName := &IntSalesforce.SalesforceSmartEventName{
		EventName: smartEventName,
		Filter:    &filter,
		Type:      model.TYPE_CRM_SALESFORCE,
	}

	eventName1 := "ev1"
	eventName, status := store.GetStore().CreateOrGetEventName(&model.EventName{ProjectId: project.ID, Name: eventName1, Type: model.TYPE_USER_CREATED_EVENT_NAME})
	assert.Equal(t, http.StatusCreated, status)
	_, errCode := store.GetStore().CreateEvent(&model.Event{
		ProjectId:   project.ID,
		EventNameId: eventName.ID,
		UserId:      userID1,
		Timestamp:   createdDate.Unix(),
	})
	assert.Equal(t, http.StatusCreated, errCode)

	eventID2 := U.RandomLowerAphaNumString(10)
	IntSalesforce.TrackSalesforceSmartEvent(project.ID, salesforceSmartEventName, eventID2, contactID, userID1, currentSalesforceDocument.Type,
		&currentProperties, nil, createdDate.AddDate(0, 0, 2).Unix())

	query := model.Query{
		From: createdDate.Unix(),
		To:   createdDate.AddDate(0, 0, 5).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			model.QueryEventWithProperties{
				Name:       eventName1,
				Properties: []model.QueryProperty{},
			},
			model.QueryEventWithProperties{
				Name:       smartEventName,
				Properties: []model.QueryProperty{},
			},
		},
		Class: model.QueryClassFunnel,

		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	result, errCode, _ := store.GetStore().Analyze(project.ID, query)
	assert.Equal(t, http.StatusOK, errCode)
	assert.NotNil(t, result)
	assert.Equal(t, float64(1), result.Rows[0][0])
	assert.Equal(t, float64(1), result.Rows[0][1])

	// no previous record will ruturn true for all not equal to any value
	filter = model.SmartCRMEventFilter{
		Source:               model.SmartCRMEventSourceSalesforce,
		ObjectType:           "contact",
		Description:          "salesforce user created",
		FilterEvaluationType: model.FilterEvaluationTypeSpecific,
		Filters: []model.PropertyFilter{
			{
				Name: "day",
				Rules: []model.CRMFilterRule{
					{
						PropertyState: model.CurrentState,
						Value:         U.PROPERTY_VALUE_ANY,
						Operator:      model.COMPARE_EQUAL,
					},
					{
						PropertyState: model.PreviousState,
						Value:         U.PROPERTY_VALUE_ANY,
						Operator:      model.COMPARE_NOT_EQUAL,
					},
				},
				LogicalOp: model.LOGICAL_OP_AND,
			},
		},
		LogicalOp:               model.LOGICAL_OP_AND,
		TimestampReferenceField: "time",
	}

	cuid := "123-4567"
	userID2 := "123-234-455"
	currentProperties = make(map[string]interface{})
	currentProperties["day"] = "Sunday"
	_, _, ok := IntSalesforce.GetSalesforceSmartEventPayload(project.ID, "test", cuid, userID2, model.SalesforceDocumentTypeContact, &currentProperties, nil, &filter)
	assert.Equal(t, true, ok)

	// if property value is nil
	prevProperties := make(map[string]interface{})
	prevProperties["day"] = nil
	_, _, ok = IntSalesforce.GetSalesforceSmartEventPayload(project.ID, "test", cuid, userID2, model.SalesforceDocumentTypeContact, &currentProperties, &prevProperties, &filter)
	assert.Equal(t, true, ok)
}

func TestSalesforceEventUserPropertiesState(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	refreshToken := U.RandomLowerAphaNumString(5)
	instancURL := U.RandomLowerAphaNumString(5)
	errCode := store.GetStore().UpdateAgentIntSalesforce(agent.UUID,
		refreshToken,
		instancURL,
	)
	assert.Equal(t, http.StatusAccepted, errCode)

	_, errCode = store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		IntSalesforceEnabledAgentUUID: &agent.UUID,
	})
	assert.Equal(t, http.StatusAccepted, errCode)

	cuID := getRandomEmail()
	firstPropTimestamp := time.Now().Unix()
	createdUserID, status := store.GetStore().CreateUser(&model.User{
		ProjectId:      project.ID,
		JoinTimestamp:  firstPropTimestamp,
		CustomerUserId: cuID,
		Source:         model.GetRequestSourcePointer(model.UserSourceSalesforce),
	})
	assert.Equal(t, http.StatusCreated, status)
	assert.NotEmpty(t, createdUserID)

	properties := &postgres.Jsonb{RawMessage: []byte(`{"name":"user1","city":"bangalore"}`)}
	_, status = store.GetStore().UpdateUserProperties(project.ID, createdUserID, properties, firstPropTimestamp)
	assert.Equal(t, http.StatusAccepted, status)

	contactID := U.RandomLowerAphaNumString(7)
	name := U.RandomLowerAphaNumString(3)
	createdDate := time.Now()

	// salesforce record
	jsonData := fmt.Sprintf(`{"Id":"%s", "name":"%s","CreatedDate":"%s", "LastModifiedDate":"%s","Email":"%s"}`, contactID, name, createdDate.UTC().Format(model.SalesforceDocumentDateTimeLayout), createdDate.UTC().Format(model.SalesforceDocumentDateTimeLayout), cuID)
	salesforceDocument := &model.SalesforceDocument{
		ProjectID: project.ID,
		TypeAlias: model.SalesforceDocumentTypeNameLead,
		Value:     &postgres.Jsonb{RawMessage: json.RawMessage([]byte(jsonData))},
	}

	status = store.GetStore().CreateSalesforceDocument(project.ID, salesforceDocument)
	assert.Equal(t, http.StatusCreated, status)

	//should return error on duplicate
	status = store.GetStore().CreateSalesforceDocument(project.ID, salesforceDocument)
	assert.Equal(t, http.StatusConflict, status)

	//enrich job, create contact created and contact updated event
	enrichStatus, _ := IntSalesforce.Enrich(project.ID)
	assert.Equal(t, project.ID, enrichStatus[0].ProjectID)
	assert.Len(t, enrichStatus, 1)
	assert.Equal(t, "success", enrichStatus[0].Status)
	query := model.Query{
		From: createdDate.Unix() - 500,
		To:   createdDate.Unix() + 500,
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       "$sf_lead_created",
				Properties: []model.QueryProperty{},
			},
		},
		Class: model.QueryClassFunnel,
		GroupByProperties: []model.QueryGroupByProperty{
			{
				Entity:         model.PropertyEntityUser,
				Property:       "city",
				EventName:      "$sf_lead_created",
				EventNameIndex: 1,
			},
		},

		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	result, status, _ := store.GetStore().Analyze(project.ID, query)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, "city", result.Headers[0])
	assert.Equal(t, "bangalore", result.Rows[1][0])
	assert.Equal(t, float64(1), result.Rows[1][1])

	query = model.Query{
		From: createdDate.Unix() - 500,
		To:   createdDate.Unix() + 500,
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       "$sf_lead_created",
				Properties: []model.QueryProperty{},
			},
		},
		Class: model.QueryClassFunnel,
		GroupByProperties: []model.QueryGroupByProperty{
			{
				Entity:         model.PropertyEntityUser,
				Property:       "$user_id",
				EventName:      "$sf_lead_created",
				EventNameIndex: 1,
			},
		},

		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	result, status, _ = store.GetStore().Analyze(project.ID, query)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, cuID, result.Rows[1][0])
	assert.Equal(t, float64(1), result.Rows[1][1])
}

func sendGetCRMObjectValuesByPropertyNameReq(r *gin.Engine, projectID uint64, agent *model.Agent, objectSource, objectType, propertyName string) *httptest.ResponseRecorder {

	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
		return nil
	}

	rb := C.NewRequestBuilderWithPrefix(http.MethodGet, fmt.Sprintf("/projects/%d/v1/crm/%s/%s/properties/%s/values", projectID, objectSource, objectType, propertyName)).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error creating request")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestSalesforceObjectPropertiesAPI(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	r := gin.Default()
	H.InitAppRoutes(r)

	property1 := U.RandomLowerAphaNumString(4)
	property2 := U.RandomLowerAphaNumString(4)
	documentID := U.RandomLowerAphaNumString(4)
	createdDate := time.Now().AddDate(0, 0, -1)

	jsonData := fmt.Sprintf(`{"Id":"%s","CreatedDate":"%s", "LastModifiedDate":"%s"}`, documentID, createdDate.UTC().Format(model.SalesforceDocumentDateTimeLayout), createdDate.UTC().Format(model.SalesforceDocumentDateTimeLayout))
	salesforceDocumentPrev := &model.SalesforceDocument{
		ProjectID: project.ID,
		TypeAlias: model.SalesforceDocumentTypeNameContact,
		Value:     &postgres.Jsonb{RawMessage: json.RawMessage([]byte(jsonData))},
	}
	status := store.GetStore().CreateSalesforceDocument(project.ID, salesforceDocumentPrev)
	assert.Equal(t, http.StatusCreated, status)

	limit := 100
	for i := 0; i < limit; i++ {
		createdDate = createdDate.Add(10 * time.Second)
		value1 := fmt.Sprintf("%s_%d", property1, i)
		value2 := fmt.Sprintf("%s_%d", property2, i)
		jsonData = fmt.Sprintf(`{"Id":"%s","%s":"%s", "%s":"%s","CreatedDate":"%s", "LastModifiedDate":"%s"}`, documentID, property1, value1, property2, value2, createdDate.UTC().Format(model.SalesforceDocumentDateTimeLayout), createdDate.UTC().Format(model.SalesforceDocumentDateTimeLayout))
		salesforceDocumentPrev := &model.SalesforceDocument{
			ProjectID: project.ID,
			TypeAlias: model.SalesforceDocumentTypeNameContact,
			Value:     &postgres.Jsonb{RawMessage: json.RawMessage([]byte(jsonData))},
		}
		status := store.GetStore().CreateSalesforceDocument(project.ID, salesforceDocumentPrev)
		assert.Equal(t, http.StatusCreated, status)
	}

	var property1Values []interface{}
	var property2Values []interface{}
	w := sendGetCRMObjectValuesByPropertyNameReq(r, project.ID, agent, model.SmartCRMEventSourceSalesforce, model.SalesforceDocumentTypeNameContact, property1)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	err = json.Unmarshal(jsonResponse, &property1Values)
	assert.Nil(t, err)

	w = sendGetCRMObjectValuesByPropertyNameReq(r, project.ID, agent, model.SmartCRMEventSourceSalesforce, model.SalesforceDocumentTypeNameContact, property2)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	err = json.Unmarshal(jsonResponse, &property2Values)
	assert.Nil(t, err)
	for i := 0; i < limit; i++ {
		assert.Contains(t, property1Values, fmt.Sprintf("%s_%d", property1, i))
		assert.Contains(t, property2Values, fmt.Sprintf("%s_%d", property2, i))
	}

	for i := 0; i < 5; i++ {
		for j := 0; j < i+1; j++ {
			createdDate = createdDate.Add(10 * time.Second)
			value1 := fmt.Sprintf("%s_%d", property1, i)
			value2 := fmt.Sprintf("%s_%d", property2, i)
			jsonData = fmt.Sprintf(`{"Id":"%s","%s":"%s", "%s":"%s","CreatedDate":"%s", "LastModifiedDate":"%s"}`, documentID, property1, value1, property2, value2, createdDate.UTC().Format(model.SalesforceDocumentDateTimeLayout), createdDate.UTC().Format(model.SalesforceDocumentDateTimeLayout))
			salesforceDocumentPrev := &model.SalesforceDocument{
				ProjectID: project.ID,
				TypeAlias: model.SalesforceDocumentTypeNameContact,
				Value:     &postgres.Jsonb{RawMessage: json.RawMessage([]byte(jsonData))},
			}
			status := store.GetStore().CreateSalesforceDocument(project.ID, salesforceDocumentPrev)
			assert.Equal(t, http.StatusCreated, status)
		}
	}

	createdDate = createdDate.Add(10 * time.Second)
	value3 := "val3"
	jsonData = fmt.Sprintf(`{"Id":"%s","%s":"%s", "%s":"%s","CreatedDate":"%s", "LastModifiedDate":"%s"}`, documentID, property1, value3, property2, value3, createdDate.UTC().Format(model.SalesforceDocumentDateTimeLayout), createdDate.UTC().Format(model.SalesforceDocumentDateTimeLayout))
	salesforceDocumentPrev = &model.SalesforceDocument{
		ProjectID: project.ID,
		TypeAlias: model.SalesforceDocumentTypeNameContact,
		Value:     &postgres.Jsonb{RawMessage: json.RawMessage([]byte(jsonData))},
	}
	status = store.GetStore().CreateSalesforceDocument(project.ID, salesforceDocumentPrev)
	assert.Equal(t, http.StatusCreated, status)

	w = sendGetCRMObjectValuesByPropertyNameReq(r, project.ID, agent, model.SmartCRMEventSourceSalesforce, model.SalesforceDocumentTypeNameContact, property1)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	err = json.Unmarshal(jsonResponse, &property1Values)
	assert.Nil(t, err)

	w = sendGetCRMObjectValuesByPropertyNameReq(r, project.ID, agent, model.SmartCRMEventSourceSalesforce, model.SalesforceDocumentTypeNameContact, property2)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	err = json.Unmarshal(jsonResponse, &property2Values)
	assert.Nil(t, err)
	for i := range property1Values[:6] {
		if i == 0 {
			assert.Equal(t, "$none", property1Values[i])
			continue
		}
		assert.Equal(t, fmt.Sprintf("%s_%d", property1, 5-i), property1Values[i])
		assert.Equal(t, fmt.Sprintf("%s_%d", property2, 5-i), property2Values[i])
	}

}

func TestSalesforcePropertyDetails(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	r := gin.Default()
	H.InitAppRoutes(r)

	refreshToken := U.RandomLowerAphaNumString(5)
	instancURL := U.RandomLowerAphaNumString(5)
	errCode := store.GetStore().UpdateAgentIntSalesforce(agent.UUID,
		refreshToken,
		instancURL,
	)
	assert.Equal(t, http.StatusAccepted, errCode)

	status := IntSalesforce.CreateOrGetSalesforceEventName(project.ID)
	assert.Equal(t, http.StatusOK, status)

	// creating event property without registered event name
	createdDate := time.Now()
	eventNameCreated := model.GetSalesforceEventNameByAction(model.SalesforceDocumentTypeNameLead, model.SalesforceDocumentCreated)

	// datetime property detail
	eventNameUpdated := model.GetSalesforceEventNameByAction(model.SalesforceDocumentTypeNameLead, model.SalesforceDocumentUpdated)
	dtPropertyName1 := "last_visit"
	dtPropertyValue1 := createdDate
	dtPropertyName2 := "next_visit"
	dtPropertyValue2 := createdDate.AddDate(0, 0, 1)

	// numerical property detail
	numPropertyName1 := "lead_vists"
	numPropertyValue1 := 15
	numPropertyName2 := "lead_views"
	numPropertyValue2 := 10

	// datetime property
	dtEnKey1 := model.GetCRMEnrichPropertyKeyByType(
		model.SmartCRMEventSourceSalesforce,
		model.SalesforceDocumentTypeNameLead,
		U.GetPropertyValueAsString(dtPropertyName1),
	)
	dtEnKey2 := model.GetCRMEnrichPropertyKeyByType(
		model.SmartCRMEventSourceSalesforce,
		model.SalesforceDocumentTypeNameLead,
		U.GetPropertyValueAsString(dtPropertyName2),
	)

	// numerical property
	numEnKey1 := model.GetCRMEnrichPropertyKeyByType(
		model.SmartCRMEventSourceSalesforce,
		model.SalesforceDocumentTypeNameLead,
		U.GetPropertyValueAsString(numPropertyName1),
	)
	numEnKey2 := model.GetCRMEnrichPropertyKeyByType(
		model.SmartCRMEventSourceSalesforce,
		model.SalesforceDocumentTypeNameLead,
		U.GetPropertyValueAsString(numPropertyName2),
	)

	// datetime property details
	status = store.GetStore().CreatePropertyDetails(project.ID, eventNameCreated, dtEnKey1, U.PropertyTypeDateTime, false, false)
	assert.Equal(t, http.StatusCreated, status)
	status = store.GetStore().CreatePropertyDetails(project.ID, eventNameCreated, dtEnKey2, U.PropertyTypeDateTime, false, false)
	assert.Equal(t, http.StatusCreated, status)

	status = store.GetStore().CreatePropertyDetails(project.ID, "", dtEnKey1, U.PropertyTypeDateTime, true, false)
	assert.Equal(t, http.StatusCreated, status)
	status = store.GetStore().CreatePropertyDetails(project.ID, "", dtEnKey2, U.PropertyTypeDateTime, true, false)
	assert.Equal(t, http.StatusCreated, status)

	status = store.GetStore().CreatePropertyDetails(project.ID, eventNameUpdated, dtEnKey1, U.PropertyTypeDateTime, false, false)
	assert.Equal(t, http.StatusCreated, status)
	status = store.GetStore().CreatePropertyDetails(project.ID, eventNameUpdated, dtEnKey2, U.PropertyTypeDateTime, false, false)
	assert.Equal(t, http.StatusCreated, status)

	// numerical property details
	status = store.GetStore().CreatePropertyDetails(project.ID, eventNameCreated, numEnKey1, U.PropertyTypeNumerical, false, false)
	assert.Equal(t, http.StatusCreated, status)
	status = store.GetStore().CreatePropertyDetails(project.ID, eventNameCreated, numEnKey2, U.PropertyTypeNumerical, false, false)
	assert.Equal(t, http.StatusCreated, status)

	status = store.GetStore().CreatePropertyDetails(project.ID, "", numEnKey1, U.PropertyTypeNumerical, true, false)
	assert.Equal(t, http.StatusCreated, status)
	status = store.GetStore().CreatePropertyDetails(project.ID, "", numEnKey2, U.PropertyTypeNumerical, true, false)
	assert.Equal(t, http.StatusCreated, status)

	status = store.GetStore().CreatePropertyDetails(project.ID, eventNameUpdated, numEnKey1, U.PropertyTypeNumerical, false, false)
	assert.Equal(t, http.StatusCreated, status)
	status = store.GetStore().CreatePropertyDetails(project.ID, eventNameUpdated, numEnKey2, U.PropertyTypeNumerical, false, false)
	assert.Equal(t, http.StatusCreated, status)

	documentID := U.RandomLowerAphaNumString(4)
	jsonData := fmt.Sprintf(`{"Id":"%s","%s":"%s", "%s":"%s","%s":"%d", "%s":"%d","CreatedDate":"%s", "LastModifiedDate":"%s"}`, documentID, dtPropertyName1, dtPropertyValue1.UTC().Format(model.SalesforceDocumentDateTimeLayout), dtPropertyName2, dtPropertyValue2.UTC().Format(model.SalesforceDocumentDateTimeLayout), numPropertyName1, numPropertyValue1, numPropertyName2, numPropertyValue2, createdDate.UTC().Format(model.SalesforceDocumentDateTimeLayout), createdDate.UTC().Format(model.SalesforceDocumentDateTimeLayout))
	salesforceDocument := &model.SalesforceDocument{
		ProjectID: project.ID,
		TypeAlias: model.SalesforceDocumentTypeNameLead,
		Value:     &postgres.Jsonb{RawMessage: json.RawMessage([]byte(jsonData))},
	}
	status = store.GetStore().CreateSalesforceDocument(project.ID, salesforceDocument)
	assert.Equal(t, http.StatusCreated, status)

	allStatus, _ := IntSalesforce.Enrich(project.ID)
	for i := range allStatus {
		assert.Equal(t, U.CRM_SYNC_STATUS_SUCCESS, allStatus[i].Status)
	}

	configs := make(map[string]interface{})
	configs["rollupLookback"] = 1
	event_user_cache.DoRollUpSortedSet(configs)
	properties, err := store.GetStore().GetPropertiesByEvent(project.ID, eventNameCreated, 2500, 1)
	assert.Nil(t, err)
	assert.Contains(t, properties[U.PropertyTypeDateTime], dtEnKey1, dtEnKey2)
	assert.Contains(t, properties[U.PropertyTypeNumerical], numEnKey1, numEnKey2)

	properties, err = store.GetStore().GetUserPropertiesByProject(project.ID, 100, 10)
	assert.Nil(t, err)
	assert.Contains(t, properties[U.PropertyTypeDateTime], dtEnKey1, dtEnKey2)
	assert.Contains(t, properties[U.PropertyTypeNumerical], numEnKey1, numEnKey2)

	query := model.Query{
		From: createdDate.Unix() - 500,
		To:   createdDate.Unix() + 500,
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name: "$sf_lead_updated",
			},
		},
		GroupByProperties: []model.QueryGroupByProperty{
			{
				Entity:   model.PropertyEntityEvent,
				Property: dtEnKey1,
			},
			{
				Entity:   model.PropertyEntityEvent,
				Property: dtEnKey2,
			},
			{
				Entity:   model.PropertyEntityEvent,
				Property: numEnKey1,
			},
			{
				Entity:   model.PropertyEntityEvent,
				Property: numEnKey2,
			},
		},
		Class:           model.QueryClassEvents,
		Type:            model.QueryTypeEventsOccurrence,
		EventsCondition: model.EventCondAnyGivenEvent,
	}

	result, status, _ := store.GetStore().Analyze(project.ID, query)
	assert.Equal(t, http.StatusOK, status)
	assert.Contains(t, result.Headers, dtEnKey1, dtEnKey2, numEnKey1, numEnKey2)
	count := 0
	for i := range result.Headers[:len(result.Headers)-1] {
		if result.Headers[i] == dtEnKey1 {
			assert.Equal(t, fmt.Sprint(dtPropertyValue1.Unix()), result.Rows[0][i])
			count++
		}
		if result.Headers[i] == dtEnKey2 {
			assert.Equal(t, fmt.Sprint(dtPropertyValue2.Unix()), result.Rows[0][i])
			count++
		}

		if result.Headers[i] == numEnKey1 {
			assert.Equal(t, fmt.Sprint(numPropertyValue1), result.Rows[0][i])
			count++
		}

		if result.Headers[i] == numEnKey2 {
			assert.Equal(t, fmt.Sprint(numPropertyValue2), result.Rows[0][i])
			count++
		}
	}
	assert.Equal(t, 4, count)

}

func TestSalesforceIdentification(t *testing.T) {
	project, _, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	// custom field. Will always have priority
	testIndentificationField := map[string][]string{
		model.SalesforceDocumentTypeNameLead:        {"MobilePhone"},
		model.SalesforceDocumentTypeNameOpportunity: {"Mobile__c"},
	}

	SalesforceProjectIdentificationFieldStore := map[uint64]map[string][]string{
		project.ID: testIndentificationField,
	}

	/*
		Email Identification
	*/
	// Should return standard email field
	emailFields := model.GetSalesforceEmailFieldByProjectIDAndObjectName(project.ID, model.SalesforceDocumentTypeNameAccount, &SalesforceProjectIdentificationFieldStore)
	assert.Equal(t, 1, len(emailFields))
	assert.Equal(t, "PersonEmail", emailFields[0])
	emailFields = model.GetSalesforceEmailFieldByProjectIDAndObjectName(project.ID, model.SalesforceDocumentTypeNameContact, &SalesforceProjectIdentificationFieldStore)
	assert.Equal(t, 1, len(emailFields))
	assert.Equal(t, "Email", emailFields[0])

	// Custom field will always have priority
	emailFields = model.GetSalesforceEmailFieldByProjectIDAndObjectName(project.ID, model.SalesforceDocumentTypeNameLead, &SalesforceProjectIdentificationFieldStore)
	assert.Equal(t, 2, len(emailFields))
	assert.Equal(t, "MobilePhone", emailFields[0])
	assert.Equal(t, "Email", emailFields[1])
	emailFields = model.GetSalesforceEmailFieldByProjectIDAndObjectName(project.ID, model.SalesforceDocumentTypeNameOpportunity, &SalesforceProjectIdentificationFieldStore)
	assert.Equal(t, 1, len(emailFields))
	assert.Equal(t, "Mobile__c", emailFields[0])

	/*
		Phone Identification
	*/
	// Should return standard email field
	phoneFields := model.GetSalesforcePhoneFieldByProjectIDAndObjectName(project.ID, model.SalesforceDocumentTypeNameAccount, &SalesforceProjectIdentificationFieldStore)
	assert.Equal(t, 2, len(phoneFields))
	assert.Equal(t, "Phone", phoneFields[0])
	assert.Equal(t, "PersonMobilePhone", phoneFields[1])
	phoneFields = model.GetSalesforcePhoneFieldByProjectIDAndObjectName(project.ID, model.SalesforceDocumentTypeNameContact, &SalesforceProjectIdentificationFieldStore)
	assert.Equal(t, "Phone", phoneFields[0])
	assert.Equal(t, "MobilePhone", phoneFields[1])

	// Custom field will always have priority
	phoneFields = model.GetSalesforcePhoneFieldByProjectIDAndObjectName(project.ID, model.SalesforceDocumentTypeNameLead, &SalesforceProjectIdentificationFieldStore)
	assert.Equal(t, 3, len(phoneFields))
	assert.Equal(t, "MobilePhone", phoneFields[0])
	assert.Equal(t, "Phone", phoneFields[1])
	assert.Equal(t, "MobilePhone", phoneFields[2])
	phoneFields = model.GetSalesforcePhoneFieldByProjectIDAndObjectName(project.ID, model.SalesforceDocumentTypeNameOpportunity, &SalesforceProjectIdentificationFieldStore)
	assert.Equal(t, 1, len(phoneFields))
	assert.Equal(t, "Mobile__c", phoneFields[0])

	documentID := U.RandomLowerAphaNumString(4)
	emailAccount := getRandomEmail()
	emailContact := getRandomEmail()
	emailLead := getRandomEmail()
	emailOpportunity := getRandomEmail()
	createdDate := time.Now()
	number := U.RandomUint64()
	// Use default field
	jsonDataAccount := fmt.Sprintf(`{"Id":"%s","PersonEmail":"%s","Phone":%d,"CreatedDate":"%s", "LastModifiedDate":"%s"}`, documentID, emailAccount, number, createdDate.UTC().Format(model.SalesforceDocumentDateTimeLayout), createdDate.UTC().Format(model.SalesforceDocumentDateTimeLayout))
	salesforceDocument := &model.SalesforceDocument{
		ProjectID: project.ID,
		TypeAlias: model.SalesforceDocumentTypeNameAccount,
		Value:     &postgres.Jsonb{RawMessage: json.RawMessage([]byte(jsonDataAccount))},
	}

	status := store.GetStore().CreateSalesforceDocument(project.ID, salesforceDocument)
	assert.Equal(t, http.StatusCreated, status)
	jsonDataContact := fmt.Sprintf(`{"Id":"%s","Email":"%s","MobilePhone":%d,"CreatedDate":"%s", "LastModifiedDate":"%s"}`, documentID, emailContact, number, createdDate.UTC().Format(model.SalesforceDocumentDateTimeLayout), createdDate.UTC().Format(model.SalesforceDocumentDateTimeLayout))
	salesforceDocument = &model.SalesforceDocument{
		ProjectID: project.ID,
		TypeAlias: model.SalesforceDocumentTypeNameContact,
		Value:     &postgres.Jsonb{RawMessage: json.RawMessage([]byte(jsonDataContact))},
	}

	status = store.GetStore().CreateSalesforceDocument(project.ID, salesforceDocument)
	assert.Equal(t, http.StatusCreated, status)
	jsonDataLead := fmt.Sprintf(`{"Id":"%s","Email":"%s","CreatedDate":"%s", "LastModifiedDate":"%s"}`, documentID, emailLead, createdDate.UTC().Format(model.SalesforceDocumentDateTimeLayout), createdDate.UTC().Format(model.SalesforceDocumentDateTimeLayout))
	salesforceDocument = &model.SalesforceDocument{
		ProjectID: project.ID,
		TypeAlias: model.SalesforceDocumentTypeNameLead,
		Value:     &postgres.Jsonb{RawMessage: json.RawMessage([]byte(jsonDataLead))},
	}

	status = store.GetStore().CreateSalesforceDocument(project.ID, salesforceDocument)

	// No identification as not standard field
	jsonDataOpportunity := fmt.Sprintf(`{"Id":"%s","Email__c":"%s","CreatedDate":"%s", "LastModifiedDate":"%s"}`, documentID, emailOpportunity, createdDate.UTC().Format(model.SalesforceDocumentDateTimeLayout), createdDate.UTC().Format(model.SalesforceDocumentDateTimeLayout))
	salesforceDocument = &model.SalesforceDocument{
		ProjectID: project.ID,
		TypeAlias: model.SalesforceDocumentTypeNameOpportunity,
		Value:     &postgres.Jsonb{RawMessage: json.RawMessage([]byte(jsonDataOpportunity))},
	}
	status = store.GetStore().CreateSalesforceDocument(project.ID, salesforceDocument)
	assert.Equal(t, http.StatusCreated, status)

	allStatus, _ := IntSalesforce.Enrich(project.ID)
	for i := range allStatus {
		assert.Equal(t, U.CRM_SYNC_STATUS_SUCCESS, allStatus[i].Status)
	}

	query := model.Query{
		From: createdDate.Unix() - 500,
		To:   createdDate.Unix() + 500,
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name: U.EVENT_NAME_SALESFORCE_CONTACT_CREATED,
			}, {
				Name: U.EVENT_NAME_SALESFORCE_LEAD_CREATED,
			}, {
				Name: U.EVENT_NAME_SALESFORCE_ACCOUNT_CREATED,
			}, {
				Name: U.EVENT_NAME_SALESFORCE_OPPORTUNITY_CREATED,
			},
		},
		GroupByProperties: []model.QueryGroupByProperty{
			{
				Entity:   model.PropertyEntityUser,
				Property: U.UP_USER_ID,
			},
		},
		Class:           model.QueryClassEvents,
		Type:            model.QueryTypeEventsOccurrence,
		EventsCondition: model.EventCondAnyGivenEvent,
	}

	result, _, _ := store.GetStore().Analyze(project.ID, query)
	EventUserIDMap := make(map[string]string)
	for i := range result.Rows {
		EventUserIDMap[result.Rows[i][0].(string)] = result.Rows[i][1].(string)
		assert.Equal(t, float64(1), result.Rows[i][2])
	}

	assert.Equal(t, "$none", EventUserIDMap[U.EVENT_NAME_SALESFORCE_OPPORTUNITY_CREATED])
	assert.Equal(t, emailContact, EventUserIDMap[U.EVENT_NAME_SALESFORCE_CONTACT_CREATED])
	assert.Equal(t, emailAccount, EventUserIDMap[U.EVENT_NAME_SALESFORCE_ACCOUNT_CREATED])
	assert.Equal(t, emailLead, EventUserIDMap[U.EVENT_NAME_SALESFORCE_LEAD_CREATED])
}

func createDummySalesforceDocument(projectID uint64, value interface{}, doctTypeAlias string) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return err
	}

	salesforceDocument := &model.SalesforceDocument{
		ProjectID: projectID,
		TypeAlias: doctTypeAlias,
		Value:     &postgres.Jsonb{RawMessage: json.RawMessage([]byte(jsonData))},
	}

	status := store.GetStore().CreateSalesforceDocument(projectID, salesforceDocument)
	if status != http.StatusCreated {
		return errors.New("failed to create salesforce document")
	}

	return nil
}

func TestSalesforceSmartEventPropertyDetails(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	r := gin.Default()
	H.InitAppRoutes(r)

	filter := model.SmartCRMEventFilter{
		Source:               model.SmartCRMEventSourceSalesforce,
		ObjectType:           "contact",
		Description:          "salesforce user created",
		FilterEvaluationType: model.FilterEvaluationTypeSpecific,
		Filters: []model.PropertyFilter{
			{
				Name: "day",
				Rules: []model.CRMFilterRule{
					{
						PropertyState: model.CurrentState,
						Value:         U.PROPERTY_VALUE_ANY,
						Operator:      model.COMPARE_EQUAL,
					},
					{
						PropertyState: model.PreviousState,
						Value:         U.PROPERTY_VALUE_ANY,
						Operator:      model.COMPARE_NOT_EQUAL,
					},
				},
				LogicalOp: model.LOGICAL_OP_AND,
			},
		},
		LogicalOp:               model.LOGICAL_OP_AND,
		TimestampReferenceField: model.TimestampReferenceTypeDocument,
	}

	smartEventName := "Event 1"
	requestPayload := make(map[string]interface{})
	requestPayload["name"] = smartEventName
	requestPayload["expr"] = filter

	w := sendCreateSmartEventFilterReq(r, project.ID, agent, &requestPayload)
	assert.Equal(t, http.StatusCreated, w.Code)

	documentID := U.RandomLowerAphaNumString(4)
	emailLead := getRandomEmail()
	createdDate := time.Now()

	jsonDataContact := fmt.Sprintf(`{"Id":"%s","Email":"%s","CreatedDate":"%s", "LastModifiedDate":"%s"}`, documentID, emailLead, createdDate.UTC().Format(model.SalesforceDocumentDateTimeLayout), createdDate.UTC().Format(model.SalesforceDocumentDateTimeLayout))
	salesforceDocument := &model.SalesforceDocument{
		ProjectID: project.ID,
		TypeAlias: model.SalesforceDocumentTypeNameContact,
		Value:     &postgres.Jsonb{RawMessage: json.RawMessage([]byte(jsonDataContact))},
	}

	status := store.GetStore().CreateSalesforceDocument(project.ID, salesforceDocument)
	assert.Equal(t, http.StatusCreated, status)

	jsonDataContact = fmt.Sprintf(`{"Id":"%s","Email":"%s","day":"%s","CreatedDate":"%s", "LastModifiedDate":"%s"}`, documentID, emailLead, createdDate.Add(2*time.Second).UTC().Format(model.SalesforceDocumentDateTimeLayout), createdDate.UTC().Format(model.SalesforceDocumentDateTimeLayout), createdDate.Add(10*time.Minute).UTC().Format(model.SalesforceDocumentDateTimeLayout))
	salesforceDocument = &model.SalesforceDocument{
		ProjectID: project.ID,
		TypeAlias: model.SalesforceDocumentTypeNameContact,
		Value:     &postgres.Jsonb{RawMessage: json.RawMessage([]byte(jsonDataContact))},
	}

	status = store.GetStore().CreateSalesforceDocument(project.ID, salesforceDocument)
	assert.Equal(t, http.StatusCreated, status)

	dtEnKey1 := model.GetCRMEnrichPropertyKeyByType(
		model.SmartCRMEventSourceSalesforce,
		model.SalesforceDocumentTypeNameContact,
		"day",
	)

	dtEnKey2 := model.GetCRMEnrichPropertyKeyByType(
		model.SmartCRMEventSourceSalesforce,
		model.SalesforceDocumentTypeNameContact,
		"CreatedDate",
	)

	dtEnKey3 := model.GetCRMEnrichPropertyKeyByType(
		model.SmartCRMEventSourceSalesforce,
		model.SalesforceDocumentTypeNameContact,
		"LastModifiedDate",
	)

	_, status = store.GetStore().CreateOrGetEventName(&model.EventName{ProjectId: project.ID, Name: U.EVENT_NAME_SALESFORCE_CONTACT_CREATED, Type: model.TYPE_USER_CREATED_EVENT_NAME})
	assert.Equal(t, http.StatusCreated, status)
	_, status = store.GetStore().CreateOrGetEventName(&model.EventName{ProjectId: project.ID, Name: U.EVENT_NAME_SALESFORCE_CONTACT_UPDATED, Type: model.TYPE_USER_CREATED_EVENT_NAME})
	assert.Equal(t, http.StatusCreated, status)

	status = store.GetStore().CreatePropertyDetails(project.ID, U.EVENT_NAME_SALESFORCE_CONTACT_CREATED, dtEnKey1, U.PropertyTypeDateTime, false, false)
	assert.Equal(t, http.StatusCreated, status)
	status = store.GetStore().CreatePropertyDetails(project.ID, U.EVENT_NAME_SALESFORCE_CONTACT_UPDATED, dtEnKey1, U.PropertyTypeDateTime, false, false)
	assert.Equal(t, http.StatusCreated, status)
	status = store.GetStore().CreatePropertyDetails(project.ID, U.EVENT_NAME_SALESFORCE_CONTACT_CREATED, dtEnKey2, U.PropertyTypeDateTime, false, false)
	assert.Equal(t, http.StatusCreated, status)
	status = store.GetStore().CreatePropertyDetails(project.ID, U.EVENT_NAME_SALESFORCE_CONTACT_UPDATED, dtEnKey2, U.PropertyTypeDateTime, false, false)
	assert.Equal(t, http.StatusCreated, status)
	status = store.GetStore().CreatePropertyDetails(project.ID, U.EVENT_NAME_SALESFORCE_CONTACT_CREATED, dtEnKey3, U.PropertyTypeDateTime, false, false)
	assert.Equal(t, http.StatusCreated, status)
	status = store.GetStore().CreatePropertyDetails(project.ID, U.EVENT_NAME_SALESFORCE_CONTACT_UPDATED, dtEnKey3, U.PropertyTypeDateTime, false, false)
	assert.Equal(t, http.StatusCreated, status)

	status = store.GetStore().CreatePropertyDetails(project.ID, "", dtEnKey1, U.PropertyTypeDateTime, true, false)
	assert.Equal(t, http.StatusCreated, status)
	status = store.GetStore().CreatePropertyDetails(project.ID, "", dtEnKey2, U.PropertyTypeDateTime, true, false)
	assert.Equal(t, http.StatusCreated, status)
	status = store.GetStore().CreatePropertyDetails(project.ID, "", dtEnKey3, U.PropertyTypeDateTime, true, false)
	assert.Equal(t, http.StatusCreated, status)

	enrichStatus, _ := IntSalesforce.Enrich(project.ID)
	assert.Equal(t, project.ID, enrichStatus[0].ProjectID)
	assert.Len(t, enrichStatus, 1)
	assert.Equal(t, "success", enrichStatus[0].Status)

	configs := make(map[string]interface{})
	configs["rollupLookback"] = 1
	event_user_cache.DoRollUpSortedSet(configs)

	properties, err := store.GetStore().GetPropertiesByEvent(project.ID, U.EVENT_NAME_SALESFORCE_CONTACT_UPDATED, 2500, 1)
	assert.Nil(t, err)
	assert.Contains(t, properties[U.PropertyTypeDateTime], dtEnKey1, dtEnKey2, dtEnKey3)
	properties, err = store.GetStore().GetUserPropertiesByProject(project.ID, 100, 10)
	assert.Nil(t, err)
	assert.Contains(t, properties[U.PropertyTypeDateTime], dtEnKey1, dtEnKey2, dtEnKey3)

	properties, err = store.GetStore().GetPropertiesByEvent(project.ID, smartEventName, 2500, 1)
	assert.Nil(t, err)
	assert.Contains(t, properties[U.PropertyTypeDateTime], "$curr_salesforce_contact_day")

	query := model.Query{
		From: createdDate.Unix() - 500,
		To:   createdDate.Unix() + 5000,
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name: U.EVENT_NAME_SALESFORCE_CONTACT_CREATED,
			},
		},
		GroupByProperties: []model.QueryGroupByProperty{
			{
				Entity:   model.PropertyEntityEvent,
				Property: "$salesforce_contact_lastmodifieddate",
			},
		},
		Class:           model.QueryClassEvents,
		Type:            model.QueryTypeEventsOccurrence,
		EventsCondition: model.EventCondAnyGivenEvent,
	}

	result, _, _ := store.GetStore().Analyze(project.ID, query)
	assert.Equal(t, fmt.Sprintf("%d", createdDate.Unix()), result.Rows[0][0])
	assert.Equal(t, float64(1), result.Rows[0][1])

	query = model.Query{
		From: createdDate.Unix() - 500,
		To:   createdDate.Unix() + 5000,
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name: smartEventName,
			},
		},
		GroupByProperties: []model.QueryGroupByProperty{
			{
				Entity:   model.PropertyEntityEvent,
				Property: "$curr_salesforce_contact_day",
			},
		},
		Class:           model.QueryClassEvents,
		Type:            model.QueryTypeEventsOccurrence,
		EventsCondition: model.EventCondAnyGivenEvent,
	}

	result, _, _ = store.GetStore().Analyze(project.ID, query)
	assert.Equal(t, fmt.Sprintf("%d", createdDate.Add(2*time.Second).Unix()), result.Rows[0][0])
	assert.Equal(t, float64(1), result.Rows[0][1])
}

func TestSalesforceCampaignTest(t *testing.T) {
	project, _, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	/*

		###Campaigns
			Campaign1:
				ID: campaign1ID
				Name: campaign1Name
				CampaignMember:
					CampaignMember1:
						ID: campaignMember1ID
					CampaignMember2:
						ID:campaignMember2ID

			Campaign2:
				ID: campaign2ID
				Name: campaign2Name
					CampaignMember:
						CampaignMember3:
							ID: campaignMember3ID
						CampaignMember4:
							ID:campaignMember4ID

		###CampaignMembers:
			CampaignMember1:
				LeadID: campaignMember1LeadID
			CampaignMember2:
				ContactID: campaignMember2ContactID
			CampaignMember3:
				LeadID: campaignMember3LeadID
			CampaignMember4:
				ContactID: campaignMember4ContactID
	*/

	campaign1ID := U.RandomString(5)
	campaign1Name := U.RandomString(3)
	campaign2ID := U.RandomString(5)
	campaign2Name := U.RandomString(3)
	campaignMember1ID := U.RandomString(5)
	campaignMember2ID := U.RandomString(5)
	campaignMember3ID := U.RandomString(5)
	campaignMember4ID := U.RandomString(5)
	campaignMember1LeadID := U.RandomString(5)
	campaignMember3LeadID := U.RandomString(5)

	campaignMember2ContactID := U.RandomString(5)
	campaignMember4ContactID := U.RandomString(5)

	campaign1CreatedTimestamp := time.Now().AddDate(0, 0, -1)
	campaignMember1CreatedTimestamp := campaign1CreatedTimestamp.Add(10 * time.Second)
	campaignMember2CreatedTimestamp := campaign1CreatedTimestamp.Add(20 * time.Second)
	campaign2CreatedTimestamp := time.Now().AddDate(0, 0, -1).Add(500 * time.Second)
	campaignMember3CreatedTimestamp := campaign2CreatedTimestamp.Add(10 * time.Second)
	campaignMember4CreatedTimestamp := campaign2CreatedTimestamp.Add(20 * time.Second)

	Campaign1 := map[string]interface{}{
		"Id":   campaign1ID,
		"Name": campaign1Name,
		"CampaignMembers": IntSalesforce.RelationshipCampaignMember{
			Records: []IntSalesforce.RelationshipCampaignMemberRecord{
				{
					ID: campaignMember1ID,
				},
				{
					ID: campaignMember2ID,
				},
			},
		},
		"CreatedDate":      campaign1CreatedTimestamp.Format(model.SalesforceDocumentDateTimeLayout),
		"LastModifiedDate": campaign1CreatedTimestamp.Add(2 * time.Hour).Format(model.SalesforceDocumentDateTimeLayout),
	}

	Campaign1lead := map[string]interface{}{
		"Id":               campaignMember1LeadID,
		"Name":             "Campaign1lead",
		"CreatedDate":      time.Now().AddDate(0, 0, -2).Format(model.SalesforceDocumentDateTimeLayout),
		"LastModifiedDate": time.Now().AddDate(0, 0, -2).Add(30 * time.Second).Format(model.SalesforceDocumentDateTimeLayout),
	}

	Campaign1Contact := map[string]interface{}{
		"Id":               campaignMember2ContactID,
		"Name":             "Campaign1Contact",
		"CreatedDate":      time.Now().AddDate(0, 0, -2).Format(model.SalesforceDocumentDateTimeLayout),
		"LastModifiedDate": time.Now().AddDate(0, 0, -2).Add(30 * time.Second).Format(model.SalesforceDocumentDateTimeLayout),
	}

	campaignMember1 := map[string]interface{}{
		"Id":               campaignMember1ID,
		"CampaignId":       campaign1ID,
		"LeadId":           campaignMember1LeadID,
		"CreatedDate":      campaignMember1CreatedTimestamp.Format(model.SalesforceDocumentDateTimeLayout),
		"LastModifiedDate": campaignMember1CreatedTimestamp.Add(2 * time.Hour).Format(model.SalesforceDocumentDateTimeLayout),
	}

	campaignMember2 := map[string]interface{}{
		"Id":               campaignMember2ID,
		"CampaignId":       campaign1ID,
		"ContactId":        campaignMember2ContactID,
		"CreatedDate":      campaignMember2CreatedTimestamp.Format(model.SalesforceDocumentDateTimeLayout),
		"LastModifiedDate": campaignMember2CreatedTimestamp.Add(2 * time.Hour).Format(model.SalesforceDocumentDateTimeLayout),
	}

	Campaign2 := map[string]interface{}{
		"Id":   campaign2ID,
		"Name": campaign2Name,
		"CampaignMembers": IntSalesforce.RelationshipCampaignMember{
			Records: []IntSalesforce.RelationshipCampaignMemberRecord{
				{
					ID: campaignMember3ID,
				},
				{
					ID: campaignMember4ID,
				},
			},
		},
		"CreatedDate":      campaign2CreatedTimestamp.Format(model.SalesforceDocumentDateTimeLayout),
		"LastModifiedDate": campaign2CreatedTimestamp.Add(2 * time.Hour).Format(model.SalesforceDocumentDateTimeLayout),
	}

	Campaign2lead := map[string]interface{}{
		"Id":               campaignMember3LeadID,
		"Name":             "Campaign2lead",
		"CreatedDate":      time.Now().AddDate(0, 0, -2).Format(model.SalesforceDocumentDateTimeLayout),
		"LastModifiedDate": time.Now().AddDate(0, 0, -2).Add(30 * time.Second).Format(model.SalesforceDocumentDateTimeLayout),
	}
	Campaign2Contact := map[string]interface{}{
		"Id":               campaignMember4ContactID,
		"Name":             "Campaign2Contact",
		"CreatedDate":      time.Now().AddDate(0, 0, -2).Format(model.SalesforceDocumentDateTimeLayout),
		"LastModifiedDate": time.Now().AddDate(0, 0, -2).Add(30 * time.Second).Format(model.SalesforceDocumentDateTimeLayout),
	}

	campaignMember3 := map[string]interface{}{
		"Id":               campaignMember3ID,
		"CampaignId":       campaign2ID,
		"LeadId":           campaignMember3LeadID,
		"CreatedDate":      campaignMember3CreatedTimestamp.Format(model.SalesforceDocumentDateTimeLayout),
		"LastModifiedDate": campaignMember3CreatedTimestamp.Add(2 * time.Hour).Format(model.SalesforceDocumentDateTimeLayout),
	}

	campaignMember4 := map[string]interface{}{
		"Id":               campaignMember4ID,
		"CampaignId":       campaign2ID,
		"ContactId":        campaignMember4ContactID,
		"CreatedDate":      campaignMember4CreatedTimestamp.Format(model.SalesforceDocumentDateTimeLayout),
		"LastModifiedDate": campaignMember4CreatedTimestamp.Add(2 * time.Hour).Format(model.SalesforceDocumentDateTimeLayout),
	}

	for _, campaign := range []interface{}{Campaign1, Campaign2} {
		err := createDummySalesforceDocument(project.ID, campaign, model.SalesforceDocumentTypeNameCampaign)
		assert.Nil(t, err)
	}

	for _, campaignLead := range []interface{}{Campaign1lead, Campaign2lead} {
		err := createDummySalesforceDocument(project.ID, campaignLead, model.SalesforceDocumentTypeNameLead)
		assert.Nil(t, err)
	}

	for _, campaignContact := range []interface{}{Campaign1Contact, Campaign2Contact} {
		err := createDummySalesforceDocument(project.ID, campaignContact, model.SalesforceDocumentTypeNameContact)
		assert.Nil(t, err)
	}

	for _, campaignMember := range []interface{}{campaignMember1, campaignMember2, campaignMember3, campaignMember4} {
		err := createDummySalesforceDocument(project.ID, campaignMember, model.SalesforceDocumentTypeNameCampaignMember)
		assert.Nil(t, err)
	}

	enrichStatus, anyFailure := IntSalesforce.Enrich(project.ID)
	assert.Equal(t, false, anyFailure)
	assert.Len(t, enrichStatus, 3) // only camapaing, lead and contact
	assert.Equal(t, util.CRM_SYNC_STATUS_SUCCESS, enrichStatus[0].Status)
	assert.Equal(t, util.CRM_SYNC_STATUS_SUCCESS, enrichStatus[1].Status)
	assert.Equal(t, util.CRM_SYNC_STATUS_SUCCESS, enrichStatus[2].Status)

	query := model.Query{
		From: campaign1CreatedTimestamp.Unix() - 500,
		To:   campaign2CreatedTimestamp.Unix() + 500,
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name: U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_CREATED,
			}, {
				Name: U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_UPDATED,
			},
		},
		Class:           model.QueryClassEvents,
		Type:            model.QueryTypeEventsOccurrence,
		EventsCondition: model.EventCondAnyGivenEvent,
	}

	result, _, _ := store.GetStore().Analyze(project.ID, query)
	assert.Contains(t, []string{U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_CREATED, U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_UPDATED}, result.Rows[0][0])
	assert.Contains(t, []string{U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_CREATED, U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_UPDATED}, result.Rows[1][0])
	assert.Equal(t, float64(4), result.Rows[0][1])
	assert.Equal(t, float64(4), result.Rows[1][1])

	query = model.Query{
		From: time.Now().AddDate(0, 0, -3).Unix(),
		To:   time.Now().Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       U.EVENT_NAME_SALESFORCE_CONTACT_CREATED,
				Properties: []model.QueryProperty{},
			},
			{
				Name:       U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_CREATED,
				Properties: []model.QueryProperty{},
			},
		},
		Class: model.QueryClassFunnel,

		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	result, _, _ = store.GetStore().Analyze(project.ID, query)
	assert.Equal(t, float64(2), result.Rows[0][1])
	assert.Equal(t, "100.0", result.Rows[0][2])
	assert.Equal(t, "100.0", result.Rows[0][3])

	query.EventsWithProperties[0].Name = U.EVENT_NAME_SALESFORCE_LEAD_CREATED
	result, _, _ = store.GetStore().Analyze(project.ID, query)
	assert.Equal(t, float64(2), result.Rows[0][1])
	assert.Equal(t, "100.0", result.Rows[0][2])
	assert.Equal(t, "100.0", result.Rows[0][3])

	query.EventsWithProperties[0].Name = U.EVENT_NAME_SALESFORCE_CONTACT_CREATED
	query.EventsWithProperties[1].Name = U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_UPDATED

	result, _, _ = store.GetStore().Analyze(project.ID, query)
	assert.Equal(t, float64(2), result.Rows[0][1])
	assert.Equal(t, "100.0", result.Rows[0][2])
	assert.Equal(t, "100.0", result.Rows[0][3])

	query.EventsWithProperties[0].Name = U.EVENT_NAME_SALESFORCE_LEAD_CREATED
	query.EventsWithProperties[1].Name = U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_UPDATED
	result, _, _ = store.GetStore().Analyze(project.ID, query)
	assert.Equal(t, float64(2), result.Rows[0][1])
	assert.Equal(t, "100.0", result.Rows[0][2])
	assert.Equal(t, "100.0", result.Rows[0][3])

	query = model.Query{
		From: time.Now().AddDate(0, 0, -3).Unix(),
		To:   time.Now().Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_CREATED,
				Properties: []model.QueryProperty{},
			},
		},
		GroupByProperties: []model.QueryGroupByProperty{
			{
				Entity:   model.PropertyEntityEvent,
				Property: "$salesforce_campaign_name",
			},
		},
		Class:           model.QueryClassEvents,
		Type:            model.QueryTypeEventsOccurrence,
		EventsCondition: model.EventCondAnyGivenEvent,
	}

	result, _, _ = store.GetStore().Analyze(project.ID, query)
	assert.Equal(t, 2, len(result.Rows))
	assert.Contains(t, []string{campaign1Name, campaign2Name}, result.Rows[0][0])
	assert.Equal(t, float64(2), result.Rows[0][1])
	assert.Contains(t, []string{campaign1Name, campaign2Name}, result.Rows[1][0])
	assert.Equal(t, float64(2), result.Rows[1][1])

	/*
		Campaign1:
		Lead Name :- Campaign1lead
		Contact Name :- Campaign1Contact

		Campaign2:
		Lead Name :- Campaign2lead
		Contact Name :- Campaign2Contact
	*/
	query = model.Query{
		From: time.Now().AddDate(0, 0, -3).Unix(),
		To:   time.Now().Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_CREATED,
				Properties: []model.QueryProperty{},
			},
		},
		GroupByProperties: []model.QueryGroupByProperty{
			{
				Entity:   model.PropertyEntityEvent,
				Property: "$salesforce_campaignmember_id",
			},
			{
				Entity:   model.PropertyEntityUser,
				Property: "$salesforce_contact_name",
			},
			{
				Entity:   model.PropertyEntityUser,
				Property: "$salesforce_lead_name",
			},
		},
		Class:           model.QueryClassEvents,
		Type:            model.QueryTypeEventsOccurrence,
		EventsCondition: model.EventCondAnyGivenEvent,
	}

	result, _, _ = store.GetStore().Analyze(project.ID, query)
	success := 0
	for i := range result.Rows {
		if result.Rows[i][1] == campaignMember1ID {
			assert.Equal(t, "Campaign1lead", result.Rows[i][3])
			assert.Equal(t, float64(1), result.Rows[i][4])
			success++
		}

		if result.Rows[i][1] == campaignMember2ID {
			assert.Equal(t, "Campaign1Contact", result.Rows[i][2])
			assert.Equal(t, float64(1), result.Rows[i][4])
			success++
		}

		if result.Rows[i][1] == campaignMember3ID {
			assert.Equal(t, "Campaign2lead", result.Rows[i][3])
			assert.Equal(t, float64(1), result.Rows[i][4])
			success++
		}

		if result.Rows[i][1] == campaignMember4ID {
			assert.Equal(t, "Campaign2Contact", result.Rows[i][2])
			assert.Equal(t, float64(1), result.Rows[i][4])
			success++
		}
	}

	assert.Equal(t, 4, success)
}

func TestSalesforceGetLatestSalesforceDocument(t *testing.T) {
	project, _, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	contactID1 := U.RandomString(5)
	contact1CreatedDate := time.Now().AddDate(0, 0, -1)
	contact1LastModifiedDate := contact1CreatedDate.Add(10 * time.Second)
	contactCreated := map[string]interface{}{
		"Id":               contactID1,
		"Name":             "contact1",
		"City":             "City1",
		"CreatedDate":      contact1CreatedDate.Format(model.SalesforceDocumentDateTimeLayout),
		"LastModifiedDate": contact1LastModifiedDate.Format(model.SalesforceDocumentDateTimeLayout),
	}

	err = createDummySalesforceDocument(project.ID, contactCreated, model.SalesforceDocumentTypeNameContact)
	assert.Nil(t, err)

	contactUpdated := map[string]interface{}{
		"Id":               contactID1,
		"Name":             "contact1",
		"City":             "City2",
		"CreatedDate":      contact1CreatedDate.Format(model.SalesforceDocumentDateTimeLayout),
		"LastModifiedDate": contact1LastModifiedDate.Add(10 * time.Second).Format(model.SalesforceDocumentDateTimeLayout),
	}

	err = createDummySalesforceDocument(project.ID, contactUpdated, model.SalesforceDocumentTypeNameContact)
	assert.Nil(t, err)

	document, status := store.GetStore().GetLatestSalesforceDocumentByID(project.ID, []string{contactID1}, model.GetSalesforceDocTypeByAlias(model.SalesforceDocumentTypeNameContact), 0)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, 1, len(document))
	var contactMap map[string]interface{}
	err = json.Unmarshal(document[0].Value.RawMessage, &contactMap)
	assert.Nil(t, err)
	assert.Equal(t, contactUpdated, contactMap)

	contactID2 := U.RandomString(5)
	contact2Created := map[string]interface{}{
		"Id":               contactID2,
		"Name":             "contact2",
		"City":             "City3",
		"CreatedDate":      contact1CreatedDate.Format(model.SalesforceDocumentDateTimeLayout),
		"LastModifiedDate": contact1LastModifiedDate.Format(model.SalesforceDocumentDateTimeLayout),
	}

	err = createDummySalesforceDocument(project.ID, contact2Created, model.SalesforceDocumentTypeNameContact)
	assert.Nil(t, err)

	contact2Updated := map[string]interface{}{
		"Id":               contactID2,
		"Name":             "contact1",
		"City":             "City4",
		"CreatedDate":      contact1CreatedDate.Format(model.SalesforceDocumentDateTimeLayout),
		"LastModifiedDate": contact1LastModifiedDate.Add(10 * time.Second).Format(model.SalesforceDocumentDateTimeLayout),
	}

	err = createDummySalesforceDocument(project.ID, contact2Updated, model.SalesforceDocumentTypeNameContact)
	assert.Nil(t, err)

	document, status = store.GetStore().GetLatestSalesforceDocumentByID(project.ID, []string{contactID1, contactID2}, model.GetSalesforceDocTypeByAlias(model.SalesforceDocumentTypeNameContact), 0)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, 2, len(document))
	err = json.Unmarshal(document[0].Value.RawMessage, &contactMap)
	assert.Nil(t, err)
	failed := true
	testedContact := ""
	if contactMap["Id"] == contactID2 {
		assert.Equal(t, contact2Updated, contactMap)
		failed = false
		testedContact = contactID2
	} else if contactMap["Id"] == contactID1 {
		assert.Equal(t, contactUpdated, contactMap)
		failed = false
		testedContact = contactID1
	}
	assert.Equal(t, false, failed)

	err = json.Unmarshal(document[1].Value.RawMessage, &contactMap)
	assert.Nil(t, err)
	failed = true
	if contactMap["Id"] == contactID2 && testedContact != contactID2 {
		assert.Equal(t, contact2Updated, contactMap)
		failed = false
	} else if contactMap["Id"] == contactID1 && testedContact != contactID1 {
		assert.Equal(t, contactUpdated, contactMap)
		failed = false
	}
	assert.Equal(t, false, failed)

	lead1CreatedDate := time.Now().AddDate(0, 0, -1)
	lead1LastModifiedDate := lead1CreatedDate.Add(15 * time.Second)
	leadID1 := U.RandomString(5)
	leadCreated := map[string]interface{}{
		"Id":               leadID1,
		"Name":             "Lead1",
		"City":             "City5",
		"CreatedDate":      lead1CreatedDate.Format(model.SalesforceDocumentDateTimeLayout),
		"LastModifiedDate": lead1LastModifiedDate.Format(model.SalesforceDocumentDateTimeLayout),
	}

	err = createDummySalesforceDocument(project.ID, leadCreated, model.SalesforceDocumentTypeNameLead)
	assert.Nil(t, err)

	leadUpdated := map[string]interface{}{
		"Id":               leadID1,
		"Name":             "Lead1",
		"City":             "City6",
		"CreatedDate":      lead1CreatedDate.Format(model.SalesforceDocumentDateTimeLayout),
		"LastModifiedDate": lead1LastModifiedDate.Add(15 * time.Second).Format(model.SalesforceDocumentDateTimeLayout),
	}

	err = createDummySalesforceDocument(project.ID, leadUpdated, model.SalesforceDocumentTypeNameLead)
	assert.Nil(t, err)

	document, status = store.GetStore().GetLatestSalesforceDocumentByID(project.ID, []string{leadID1}, model.GetSalesforceDocTypeByAlias(model.SalesforceDocumentTypeNameLead), 0)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, 1, len(document))
	err = json.Unmarshal(document[0].Value.RawMessage, &contactMap)
	assert.Nil(t, err)
	assert.Equal(t, leadUpdated, contactMap)
}

func TestSalesforceOpportunityAssociations(t *testing.T) {
	project, _, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	/*
		Use opportunity contact roles for stitch
	*/
	contactID1 := U.RandomString(5)
	contactID2 := U.RandomString(5)
	contactID3 := U.RandomString(5)
	contact1CreatedDate := time.Now()
	contact1LastModifiedDate := contact1CreatedDate.Add(10 * time.Second)
	contact1Created := map[string]interface{}{
		"Id":               contactID1,
		"Name":             "contact1",
		"City":             "City1",
		"CreatedDate":      contact1CreatedDate.Format(model.SalesforceDocumentDateTimeLayout),
		"LastModifiedDate": contact1LastModifiedDate.Format(model.SalesforceDocumentDateTimeLayout),
	}
	err = createDummySalesforceDocument(project.ID, contact1Created, model.SalesforceDocumentTypeNameContact)
	assert.Nil(t, err)

	contact2CreatedDate := time.Now()
	contact2LastModifiedDate := contact2CreatedDate.Add(10 * time.Second)
	contact2Created := map[string]interface{}{
		"Id":               contactID2,
		"Name":             "contact2",
		"City":             "City2",
		"CreatedDate":      contact2CreatedDate.Format(model.SalesforceDocumentDateTimeLayout),
		"LastModifiedDate": contact2LastModifiedDate.Format(model.SalesforceDocumentDateTimeLayout),
	}
	err = createDummySalesforceDocument(project.ID, contact2Created, model.SalesforceDocumentTypeNameContact)
	assert.Nil(t, err)

	contact3CreatedDate := time.Now()
	contact3LastModifiedDate := contact3CreatedDate.Add(10 * time.Second)
	contact3Created := map[string]interface{}{
		"Id":               contactID3,
		"Name":             "contact3",
		"City":             "City3",
		"CreatedDate":      contact3CreatedDate.Format(model.SalesforceDocumentDateTimeLayout),
		"LastModifiedDate": contact3LastModifiedDate.Format(model.SalesforceDocumentDateTimeLayout),
	}
	err = createDummySalesforceDocument(project.ID, contact3Created, model.SalesforceDocumentTypeNameContact)
	assert.Nil(t, err)

	oppID1 := U.RandomString(5)
	opp1CreatedDate := time.Now()
	opp1LastModifiedDate := opp1CreatedDate.Add(10 * time.Second)
	oppContactRoleID1 := U.RandomString(5)
	oppContactRoleID2 := U.RandomString(5)
	oppContactRoleID3 := U.RandomString(5)

	oppCreated := map[string]interface{}{
		"Id":               oppID1,
		"Name":             "opp1",
		"City":             "City1",
		"CreatedDate":      opp1CreatedDate.Format(model.SalesforceDocumentDateTimeLayout),
		"LastModifiedDate": opp1LastModifiedDate.Format(model.SalesforceDocumentDateTimeLayout),
		model.SalesforceChildRelationshipNameOpportunityContactRoles: IntSalesforce.RelationshipOpportunityContactRole{
			Records: []IntSalesforce.OpportunityContactRoleRecord{
				{
					ID:        oppContactRoleID1,
					IsPrimary: false,
					ContactID: contactID1,
				},
				{
					ID:        oppContactRoleID2,
					IsPrimary: true,
					ContactID: contactID2,
				},
				{
					ID:        oppContactRoleID3,
					IsPrimary: false,
					ContactID: contactID3,
				},
			},
		},
	}
	err = createDummySalesforceDocument(project.ID, oppCreated, model.SalesforceDocumentTypeNameOpportunity)
	assert.Nil(t, err)

	enrichStatus, _ := IntSalesforce.Enrich(project.ID)
	assert.Equal(t, project.ID, enrichStatus[0].ProjectID)
	assert.Len(t, enrichStatus, 3)
	assert.Equal(t, "success", enrichStatus[0].Status)
	assert.Equal(t, "success", enrichStatus[1].Status)

	query := model.Query{
		From: contact1CreatedDate.Unix() - 500,
		To:   opp1LastModifiedDate.Unix() + 500,
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       util.EVENT_NAME_SALESFORCE_CONTACT_CREATED,
				Properties: []model.QueryProperty{},
			},
			{
				Name:       util.EVENT_NAME_SALESFORCE_OPPORTUNITY_CREATED,
				Properties: []model.QueryProperty{},
			},
		},
		Class: model.QueryClassFunnel,

		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	result, errCode, _ := store.GetStore().Analyze(project.ID, query)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, float64(3), result.Rows[0][0])
	assert.Equal(t, float64(1), result.Rows[0][1])

	query.GroupByProperties = []model.QueryGroupByProperty{
		{
			Entity:    model.PropertyEntityUser,
			Property:  "$salesforce_contact_id",
			EventName: model.UserPropertyGroupByPresent,
		},
		{
			Entity:    model.PropertyEntityUser,
			Property:  "$salesforce_contact_city",
			EventName: model.UserPropertyGroupByPresent,
		},
	}
	result, errCode, _ = store.GetStore().Analyze(project.ID, query)
	assert.Equal(t, http.StatusOK, errCode)
	success := false
	for i := range result.Rows {
		if result.Rows[i][0] == contactID2 {
			assert.Equal(t, "City2", result.Rows[i][1])
			assert.Equal(t, float64(1), result.Rows[i][2])
			assert.Equal(t, float64(1), result.Rows[i][3])
			success = true
		}
	}
	assert.Equal(t, true, success)

	/*
		Use lead for opportunity stitching, opportunity contact roles will be skipped if record not available
	*/
	oppID2 := U.RandomString(5)
	leadID1 := U.RandomString(5)
	opp2CreatedDate := time.Now()
	opp2LastModifiedDate := opp2CreatedDate.Add(10 * time.Second)
	opp2Created := map[string]interface{}{
		"Id":                            oppID2,
		"Name":                          "opp2",
		"City":                          "City4",
		"CreatedDate":                   opp2CreatedDate.Format(model.SalesforceDocumentDateTimeLayout),
		"LastModifiedDate":              opp2LastModifiedDate.Format(model.SalesforceDocumentDateTimeLayout),
		IntSalesforce.OpportunityLeadID: leadID1,
		model.SalesforceChildRelationshipNameOpportunityContactRoles: IntSalesforce.RelationshipOpportunityContactRole{
			Records: []IntSalesforce.OpportunityContactRoleRecord{
				{
					ID:        "123456",
					IsPrimary: true,
					ContactID: "123445",
				},
			},
		},
	}

	err = createDummySalesforceDocument(project.ID, opp2Created, model.SalesforceDocumentTypeNameOpportunity)
	assert.Nil(t, err)

	lead1CreatedDate := time.Now()
	lead1LastModifiedDate := lead1CreatedDate.Add(10 * time.Second)
	lead1Created := map[string]interface{}{
		"Id":               leadID1,
		"Name":             "lead1",
		"City":             "City5",
		"CreatedDate":      lead1CreatedDate.Format(model.SalesforceDocumentDateTimeLayout),
		"LastModifiedDate": lead1LastModifiedDate.Format(model.SalesforceDocumentDateTimeLayout),
	}
	err = createDummySalesforceDocument(project.ID, lead1Created, model.SalesforceDocumentTypeNameLead)
	assert.Nil(t, err)

	enrichStatus, _ = IntSalesforce.Enrich(project.ID)
	assert.Equal(t, project.ID, enrichStatus[0].ProjectID)
	assert.Len(t, enrichStatus, 3)
	assert.Equal(t, "success", enrichStatus[0].Status)
	assert.Equal(t, "success", enrichStatus[1].Status)

	query = model.Query{
		From: contact1CreatedDate.Unix() - 500,
		To:   lead1LastModifiedDate.Unix() + 500,
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       util.EVENT_NAME_SALESFORCE_LEAD_CREATED,
				Properties: []model.QueryProperty{},
			},
			{
				Name:       util.EVENT_NAME_SALESFORCE_OPPORTUNITY_CREATED,
				Properties: []model.QueryProperty{},
			},
		},
		Class: model.QueryClassFunnel,

		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	result, errCode, _ = store.GetStore().Analyze(project.ID, query)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, float64(1), result.Rows[0][0])
	assert.Equal(t, float64(1), result.Rows[0][1])
}

func TestSalesforcePerDayBatching(t *testing.T) {
	project, _, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	status := IntSalesforce.CreateOrGetSalesforceEventName(project.ID)
	assert.Equal(t, http.StatusOK, status)

	eventName := U.EVENT_NAME_SALESFORCE_CONTACT_CREATED
	dateTimeProperty := "$salesforce_contact_lastmodifieddate"
	status = store.GetStore().CreatePropertyDetails(project.ID, eventName, dateTimeProperty, U.PropertyTypeDateTime, false, false)
	assert.Equal(t, http.StatusCreated, status)

	eventName = U.EVENT_NAME_SALESFORCE_CONTACT_UPDATED
	status = store.GetStore().CreatePropertyDetails(project.ID, eventName, dateTimeProperty, U.PropertyTypeDateTime, false, false)
	assert.Equal(t, http.StatusCreated, status)

	eventName = ""
	dateTimeProperty = "$salesforce_contact_lastmodifieddate"
	status = store.GetStore().CreatePropertyDetails(project.ID, eventName, dateTimeProperty, U.PropertyTypeDateTime, true, true)
	assert.Equal(t, http.StatusCreated, status)

	eventName = U.EVENT_NAME_SALESFORCE_LEAD_CREATED
	dateTimeProperty = "$salesforce_lead_lastmodifieddate"
	status = store.GetStore().CreatePropertyDetails(project.ID, eventName, dateTimeProperty, U.PropertyTypeDateTime, false, false)
	assert.Equal(t, http.StatusCreated, status)

	eventName = U.EVENT_NAME_SALESFORCE_LEAD_UPDATED
	status = store.GetStore().CreatePropertyDetails(project.ID, eventName, dateTimeProperty, U.PropertyTypeDateTime, false, false)
	assert.Equal(t, http.StatusCreated, status)

	eventName = ""
	dateTimeProperty = "$salesforce_lead_lastmodifieddate"
	status = store.GetStore().CreatePropertyDetails(project.ID, eventName, dateTimeProperty, U.PropertyTypeDateTime, true, true)
	assert.Equal(t, http.StatusCreated, status)

	/*
		generate per day time series -> {Day1,Day2}, {Day2,Day3},{Day3,Day4} upto current day
	*/
	startTimestamp := time.Now().AddDate(0, 0, -10) // 10 days excluding today
	startDate := time.Date(startTimestamp.UTC().Year(), startTimestamp.UTC().Month(), startTimestamp.UTC().Day(), 0, 0, 0, 0, time.UTC)
	expectedTimeSeries := [][]int64{}
	for i := 0; i < 11; i++ {
		expectedTimeSeries = append(expectedTimeSeries, []int64{startDate.AddDate(0, 0, i).Unix(), startDate.AddDate(0, 0, i+1).Unix()})
	}

	resultTimeSeries := model.GetCRMTimeSeriesByStartTimestamp(1, startTimestamp.Unix(), model.SmartCRMEventSourceSalesforce)
	assert.Equal(t, 11, len(resultTimeSeries)) // expected length 11

	for i := 0; i < 11; i++ {
		if i == 0 {
			assert.Equal(t, startTimestamp.Unix(), resultTimeSeries[i][0])
		} else {
			assert.Equal(t, expectedTimeSeries[i][0], resultTimeSeries[i][0])
		}

		assert.Equal(t, expectedTimeSeries[i][1], resultTimeSeries[i][1])
	}

	lead1CreatedDate := time.Now().AddDate(0, 0, -5)
	lead1LastModifiedDate := lead1CreatedDate.Add(10 * time.Second)
	leadID1 := U.RandomLowerAphaNumString(5)
	email := getRandomEmail()
	lead1Created := map[string]interface{}{
		"Id":               leadID1,
		"Name":             "lead1",
		"City":             "City1",
		"Email":            email,
		"CreatedDate":      lead1CreatedDate.Format(model.SalesforceDocumentDateTimeLayout),
		"LastModifiedDate": lead1LastModifiedDate.Format(model.SalesforceDocumentDateTimeLayout),
	}
	err = createDummySalesforceDocument(project.ID, lead1Created, model.SalesforceDocumentTypeNameLead)
	assert.Nil(t, err)

	contactID1 := U.RandomString(5)
	contact1CreatedDate := lead1CreatedDate.Add(500 * time.Second)
	contact1LastModifiedDate := contact1CreatedDate.Add(10 * time.Second)
	contactCreated := map[string]interface{}{
		"Id":               contactID1,
		"Name":             "contact1",
		"City":             "City2",
		"Email":            email,
		"CreatedDate":      contact1CreatedDate.Format(model.SalesforceDocumentDateTimeLayout),
		"LastModifiedDate": contact1LastModifiedDate.Format(model.SalesforceDocumentDateTimeLayout),
	}
	err = createDummySalesforceDocument(project.ID, contactCreated, model.SalesforceDocumentTypeNameContact)
	assert.Nil(t, err)

	lead1UpdatedDate := lead1CreatedDate.AddDate(0, 0, 1)
	lead1Updated := map[string]interface{}{
		"Id":               leadID1,
		"Name":             "lead1",
		"City":             "City3",
		"CreatedDate":      lead1CreatedDate.Format(model.SalesforceDocumentDateTimeLayout),
		"LastModifiedDate": lead1UpdatedDate.Format(model.SalesforceDocumentDateTimeLayout),
	}
	err = createDummySalesforceDocument(project.ID, lead1Updated, model.SalesforceDocumentTypeNameLead)
	assert.Nil(t, err)

	contact1UpdatedDate := lead1UpdatedDate.Add(500 * time.Minute)
	contactUpdated := map[string]interface{}{
		"Id":               contactID1,
		"Name":             "contact1",
		"City":             "City4",
		"CreatedDate":      contact1CreatedDate.Format(model.SalesforceDocumentDateTimeLayout),
		"LastModifiedDate": contact1UpdatedDate.Format(model.SalesforceDocumentDateTimeLayout),
	}
	err = createDummySalesforceDocument(project.ID, contactUpdated, model.SalesforceDocumentTypeNameContact)
	assert.Nil(t, err)
	enrichStatus, _ := IntSalesforce.Enrich(project.ID)
	assert.Equal(t, project.ID, enrichStatus[0].ProjectID)
	assert.Len(t, enrichStatus, 2)
	assert.Equal(t, "success", enrichStatus[0].Status)
	assert.Equal(t, "success", enrichStatus[1].Status)

	query := model.Query{
		From: lead1CreatedDate.AddDate(0, 0, -1).Unix(),
		To:   contact1UpdatedDate.AddDate(0, 0, 1).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       U.EVENT_NAME_SALESFORCE_LEAD_UPDATED,
				Properties: []model.QueryProperty{},
			},
			{
				Name:       U.EVENT_NAME_SALESFORCE_CONTACT_UPDATED,
				Properties: []model.QueryProperty{},
			},
		},
		Class:           model.QueryClassEvents,
		Type:            model.QueryTypeEventsOccurrence,
		EventsCondition: model.EventCondAnyGivenEvent,
	}
	analyzeResult, status, _ := store.GetStore().Analyze(project.ID, query)
	assert.Equal(t, http.StatusOK, status)
	assert.Len(t, analyzeResult.Rows, 2)
	assert.Equal(t, float64(3), analyzeResult.Rows[0][1])
	assert.Equal(t, float64(3), analyzeResult.Rows[1][1])
	assert.Subset(t, []interface{}{analyzeResult.Rows[0][0], analyzeResult.Rows[1][0]},
		[]interface{}{U.EVENT_NAME_SALESFORCE_LEAD_UPDATED, U.EVENT_NAME_SALESFORCE_CONTACT_UPDATED})

	query = model.Query{
		From: lead1CreatedDate.AddDate(0, 0, -1).Unix(),
		To:   contact1UpdatedDate.AddDate(0, 0, 1).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       U.EVENT_NAME_SALESFORCE_LEAD_UPDATED,
				Properties: []model.QueryProperty{},
			},
		},

		Type:            model.QueryTypeEventsOccurrence,
		EventsCondition: model.EventCondEachGivenEvent,
		Class:           model.QueryClassEvents,
		GroupByProperties: []model.QueryGroupByProperty{
			{
				Entity:         model.PropertyEntityUser,
				Property:       "$salesforce_lead_city",
				EventName:      U.EVENT_NAME_SALESFORCE_LEAD_UPDATED,
				EventNameIndex: 1,
			},
			{
				Entity:         model.PropertyEntityUser,
				Property:       "$salesforce_contact_city",
				EventName:      U.EVENT_NAME_SALESFORCE_LEAD_UPDATED,
				EventNameIndex: 1,
			},
		},
	}

	result, status := store.GetStore().RunEventsGroupQuery([]model.Query{query}, project.ID)
	assert.Equal(t, http.StatusOK, status)

	rows := result.Results[0].Rows
	assert.Equal(t, 2, len(rows))
	assert.Subset(t, []interface{}{rows[0][2], rows[1][2]}, []interface{}{"City1", "City3"})
	assert.Subset(t, []interface{}{rows[0][3], rows[1][3]}, []interface{}{"$none", "City2"})

	query = model.Query{
		From: lead1CreatedDate.AddDate(0, 0, -1).Unix(),
		To:   contact1UpdatedDate.AddDate(0, 0, 1).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       U.EVENT_NAME_SALESFORCE_CONTACT_UPDATED,
				Properties: []model.QueryProperty{},
			},
		},

		Type:            model.QueryTypeEventsOccurrence,
		EventsCondition: model.EventCondEachGivenEvent,
		Class:           model.QueryClassEvents,
		GroupByProperties: []model.QueryGroupByProperty{
			{
				Entity:         model.PropertyEntityUser,
				Property:       "$salesforce_lead_city",
				EventName:      U.EVENT_NAME_SALESFORCE_CONTACT_UPDATED,
				EventNameIndex: 1,
			},
		},
	}

	result, status = store.GetStore().RunEventsGroupQuery([]model.Query{query}, project.ID)
	assert.Equal(t, http.StatusOK, status)

	rows = result.Results[0].Rows
	assert.Equal(t, 2, len(rows))
	assert.Subset(t, []interface{}{rows[0][2], rows[1][2]}, []interface{}{"City1", "City3"})

	// TestCampaignMember update on daily baisis
	campaign1CreatedTimestamp := time.Now().AddDate(0, 0, -5)
	campaign1ID := U.RandomString(5)
	campaign1Name := U.RandomString(3)
	campaignMember1ID := U.RandomString(5)
	campaignMember1LeadID := U.RandomString(5)
	Campaign1 := map[string]interface{}{
		"Id":   campaign1ID,
		"Name": campaign1Name,
		"CampaignMembers": IntSalesforce.RelationshipCampaignMember{
			Records: []IntSalesforce.RelationshipCampaignMemberRecord{
				{
					ID: campaignMember1ID,
				},
			},
		},
		"Members":          1,
		"CreatedDate":      campaign1CreatedTimestamp.Format(model.SalesforceDocumentDateTimeLayout),
		"LastModifiedDate": campaign1CreatedTimestamp.Add(2 * time.Hour).Format(model.SalesforceDocumentDateTimeLayout),
	}
	err = createDummySalesforceDocument(project.ID, Campaign1, model.SalesforceDocumentTypeNameCampaign)
	assert.Nil(t, err)

	campaignMember1CreatedTimestamp := campaign1CreatedTimestamp.Add(10 * time.Minute)
	campaignMember1 := map[string]interface{}{
		"Id":               campaignMember1ID,
		"CampaignId":       campaign1ID,
		"LeadId":           campaignMember1LeadID,
		"City":             "Mumbai",
		"CreatedDate":      campaignMember1CreatedTimestamp.Format(model.SalesforceDocumentDateTimeLayout),
		"LastModifiedDate": campaignMember1CreatedTimestamp.Add(2 * time.Hour).Format(model.SalesforceDocumentDateTimeLayout),
	}
	err = createDummySalesforceDocument(project.ID, campaignMember1, model.SalesforceDocumentTypeNameCampaignMember)
	assert.Nil(t, err)

	Campaign1lead := map[string]interface{}{
		"Id":               campaignMember1LeadID,
		"Name":             "Campaign1lead",
		"CreatedDate":      campaignMember1CreatedTimestamp.Format(model.SalesforceDocumentDateTimeLayout),
		"LastModifiedDate": campaignMember1CreatedTimestamp.Add(30 * time.Second).Format(model.SalesforceDocumentDateTimeLayout),
	}
	err = createDummySalesforceDocument(project.ID, Campaign1lead, model.SalesforceDocumentTypeNameLead)
	assert.Nil(t, err)

	Campaign1 = map[string]interface{}{
		"Id":   campaign1ID,
		"Name": campaign1Name,
		"CampaignMembers": IntSalesforce.RelationshipCampaignMember{
			Records: []IntSalesforce.RelationshipCampaignMemberRecord{
				{
					ID: campaignMember1ID,
				},
			},
		},
		"Members":          3,
		"CreatedDate":      campaign1CreatedTimestamp.Format(model.SalesforceDocumentDateTimeLayout),
		"LastModifiedDate": campaign1CreatedTimestamp.AddDate(0, 0, 1).Add(2 * time.Hour).Format(model.SalesforceDocumentDateTimeLayout),
	}
	err = createDummySalesforceDocument(project.ID, Campaign1, model.SalesforceDocumentTypeNameCampaign)
	assert.Nil(t, err)

	campaignMember1 = map[string]interface{}{
		"Id":               campaignMember1ID,
		"CampaignId":       campaign1ID,
		"LeadId":           campaignMember1LeadID,
		"City":             "Delhi",
		"CreatedDate":      campaignMember1CreatedTimestamp.Format(model.SalesforceDocumentDateTimeLayout),
		"LastModifiedDate": campaignMember1CreatedTimestamp.AddDate(0, 0, 1).Add(2 * time.Hour).Format(model.SalesforceDocumentDateTimeLayout),
	}
	err = createDummySalesforceDocument(project.ID, campaignMember1, model.SalesforceDocumentTypeNameCampaignMember)
	assert.Nil(t, err)
	enrichStatus, _ = IntSalesforce.Enrich(project.ID)
	assert.Equal(t, project.ID, enrichStatus[0].ProjectID)
	assert.Equal(t, "success", enrichStatus[0].Status)
	assert.Equal(t, "success", enrichStatus[1].Status)

	query = model.Query{
		From: lead1CreatedDate.AddDate(0, 0, -1).Unix(),
		To:   contact1UpdatedDate.AddDate(0, 0, 1).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_UPDATED,
				Properties: []model.QueryProperty{},
			},
		},
		Class:           model.QueryClassEvents,
		Type:            model.QueryTypeEventsOccurrence,
		EventsCondition: model.EventCondAnyGivenEvent,
	}
	analyzeResult, status, _ = store.GetStore().Analyze(project.ID, query)
	assert.Equal(t, http.StatusOK, status)
	assert.Len(t, analyzeResult.Rows, 1)
	assert.Equal(t, float64(3), analyzeResult.Rows[0][0])

	query = model.Query{
		From: lead1CreatedDate.AddDate(0, 0, -1).Unix(),
		To:   contact1UpdatedDate.AddDate(0, 0, 1).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name: U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_UPDATED,
			},
		},
		GroupByProperties: []model.QueryGroupByProperty{
			{
				Property: "$salesforce_campaign_members",
				Entity:   "event",
				Type:     "categorical",
			},
		},
		Class:           model.QueryClassEvents,
		Type:            model.QueryTypeEventsOccurrence,
		EventsCondition: model.EventCondAnyGivenEvent,
	}
	analyzeResult, status, _ = store.GetStore().Analyze(project.ID, query)
	assert.Equal(t, http.StatusOK, status)
	assert.Subset(t, []interface{}{[]interface{}{"1", float64(2)}, []interface{}{"3", float64(1)}}, analyzeResult.Rows)
}

func TestSalesforceOpportunitySkipOnUnsyncedLead(t *testing.T) {
	project, _, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	leadCreatedDate := time.Now().AddDate(0, 0, -4)
	leadLastModifiedDate := time.Now().AddDate(0, 0, -1)
	leadDocument := map[string]interface{}{
		"Id":               "1",
		"Name":             "lead1",
		"CreatedDate":      leadCreatedDate.Format(model.SalesforceDocumentDateTimeLayout),
		"LastModifiedDate": leadLastModifiedDate.Format(model.SalesforceDocumentDateTimeLayout),
	}
	err = createDummySalesforceDocument(project.ID, leadDocument, model.SalesforceDocumentTypeNameLead)
	assert.Nil(t, err)

	oppCreatedDate := time.Now().AddDate(0, 0, -3)
	oppLastModifiedDate := time.Now().AddDate(0, 0, -3)
	leadDocument = map[string]interface{}{
		"Id":                            "1",
		"Name":                          "lead1",
		"CreatedDate":                   oppCreatedDate.Format(model.SalesforceDocumentDateTimeLayout),
		"LastModifiedDate":              oppLastModifiedDate.Format(model.SalesforceDocumentDateTimeLayout),
		IntSalesforce.OpportunityLeadID: "1",
	}
	err = createDummySalesforceDocument(project.ID, leadDocument, model.SalesforceDocumentTypeNameOpportunity)
	assert.Nil(t, err)

	enrichStatus, _ := IntSalesforce.Enrich(project.ID)
	assert.Equal(t, project.ID, enrichStatus[0].ProjectID)
	assert.Len(t, enrichStatus, 3)
	assert.Equal(t, "success", enrichStatus[0].Status)
	assert.Equal(t, "success", enrichStatus[1].Status)
	assert.Equal(t, "success", enrichStatus[2].Status)

	query := model.Query{
		From: leadCreatedDate.AddDate(0, 0, -1).Unix(),
		To:   oppLastModifiedDate.AddDate(0, 0, 1).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name: U.EVENT_NAME_SALESFORCE_LEAD_CREATED,
			},
			{
				Name: U.EVENT_NAME_SALESFORCE_OPPORTUNITY_CREATED,
			},
		},
		Class:           model.QueryClassEvents,
		Type:            model.QueryTypeEventsOccurrence,
		EventsCondition: model.EventCondAnyGivenEvent,
	}
	analyzeResult, status, _ := store.GetStore().Analyze(project.ID, query)
	assert.Equal(t, http.StatusOK, status)
	assert.Len(t, analyzeResult.Rows, 1)
	// no opportunity available
	assert.Equal(t, U.EVENT_NAME_SALESFORCE_LEAD_CREATED, analyzeResult.Rows[0][0])
	assert.Equal(t, float64(1), analyzeResult.Rows[0][1])

	// failed opportunity should be process now
	enrichStatus, _ = IntSalesforce.Enrich(project.ID)
	assert.Equal(t, project.ID, enrichStatus[0].ProjectID)
	assert.Len(t, enrichStatus, 2)
	assert.Equal(t, "success", enrichStatus[0].Status)
	assert.Equal(t, "success", enrichStatus[1].Status)

	analyzeResult, status, _ = store.GetStore().Analyze(project.ID, query)
	assert.Equal(t, http.StatusOK, status)
	assert.Len(t, analyzeResult.Rows, 2)
	assert.Equal(t, U.EVENT_NAME_SALESFORCE_LEAD_CREATED, analyzeResult.Rows[0][0])
	assert.Equal(t, float64(1), analyzeResult.Rows[0][1])

	assert.Equal(t, U.EVENT_NAME_SALESFORCE_OPPORTUNITY_CREATED, analyzeResult.Rows[1][0])
	assert.Equal(t, float64(1), analyzeResult.Rows[1][1])

}
func TestSalesforceGetDocumentsByTypeForSyncWithTimeRange(t *testing.T) {
	project, _, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	createdDate := time.Now().AddDate(0, 0, -3)
	leadDocument := map[string]interface{}{
		"Id":               1,
		"Name":             "lead1",
		"CreatedDate":      createdDate.Format(model.SalesforceDocumentDateTimeLayout),
		"LastModifiedDate": createdDate.Add(30 * time.Second).Format(model.SalesforceDocumentDateTimeLayout),
	}
	err = createDummySalesforceDocument(project.ID, leadDocument, model.SalesforceDocumentTypeNameLead)
	assert.Nil(t, err)

	lastModifiedDate := createdDate.AddDate(0, 0, 1)
	leadDocument = map[string]interface{}{
		"Id":               1,
		"Name":             "lead1",
		"CreatedDate":      createdDate.Format(model.SalesforceDocumentDateTimeLayout),
		"LastModifiedDate": lastModifiedDate.Add(30 * time.Second).Format(model.SalesforceDocumentDateTimeLayout),
	}
	err = createDummySalesforceDocument(project.ID, leadDocument, model.SalesforceDocumentTypeNameLead)
	assert.Nil(t, err)

	lastModifiedDate = lastModifiedDate.AddDate(0, 0, 1)
	leadDocument = map[string]interface{}{
		"Id":               1,
		"Name":             "lead1",
		"CreatedDate":      createdDate.Format(model.SalesforceDocumentDateTimeLayout),
		"LastModifiedDate": lastModifiedDate.Add(30 * time.Second).Format(model.SalesforceDocumentDateTimeLayout),
	}
	err = createDummySalesforceDocument(project.ID, leadDocument, model.SalesforceDocumentTypeNameLead)
	assert.Nil(t, err)

	documents, status := store.GetStore().GetSalesforceDocumentsByTypeForSync(project.ID, model.SalesforceDocumentTypeLead, 0, 0)
	assert.Equal(t, http.StatusFound, status)
	assert.Len(t, documents, 3)

	resultTimeSeries := model.GetCRMTimeSeriesByStartTimestamp(1, createdDate.Unix(), model.SmartCRMEventSourceSalesforce)
	assert.Equal(t, 4, len(resultTimeSeries))

	documents, status = store.GetStore().GetSalesforceDocumentsByTypeForSync(project.ID, model.SalesforceDocumentTypeLead, resultTimeSeries[0][0], resultTimeSeries[0][1])
	assert.Equal(t, http.StatusFound, status)
	assert.Len(t, documents, 1)
	documents, status = store.GetStore().GetSalesforceDocumentsByTypeForSync(project.ID, model.SalesforceDocumentTypeLead, resultTimeSeries[1][0], resultTimeSeries[1][1])
	assert.Equal(t, http.StatusFound, status)
	assert.Len(t, documents, 1)
	documents, status = store.GetStore().GetSalesforceDocumentsByTypeForSync(project.ID, model.SalesforceDocumentTypeLead, resultTimeSeries[2][0], resultTimeSeries[2][1])
	assert.Equal(t, http.StatusFound, status)
	assert.Len(t, documents, 1)

	// for 2 dates
	documents, status = store.GetStore().GetSalesforceDocumentsByTypeForSync(project.ID, model.SalesforceDocumentTypeLead, resultTimeSeries[1][0], resultTimeSeries[2][1])
	assert.Equal(t, http.StatusFound, status)
	assert.Len(t, documents, 2)
}

func TestSalesforceOfflineTouchPoint(t *testing.T) {

	project, _, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	_, status := store.GetStore().CreateOrGetOfflineTouchPointEventName(project.ID)
	if status != http.StatusFound && status != http.StatusConflict && status != http.StatusCreated {
		fmt.Println("failed to create event name on SF for offline touch point")
		return
	}

	campaign1CreatedTimestamp := time.Now().AddDate(0, 0, -5)
	campaign1ID := U.RandomString(5)
	campaign1Name := "Webinar_" + U.RandomString(3)
	campaignMember1ID := U.RandomString(5)
	campaignMember1LeadID := U.RandomString(5)
	Campaign1 := map[string]interface{}{
		"Id":   campaign1ID,
		"Name": campaign1Name,
		"CampaignMembers": IntSalesforce.RelationshipCampaignMember{
			Records: []IntSalesforce.RelationshipCampaignMemberRecord{
				{
					ID: campaignMember1ID,
				},
			},
		},
		"Members":          1,
		"CreatedDate":      campaign1CreatedTimestamp.Format(model.SalesforceDocumentDateTimeLayout),
		"LastModifiedDate": campaign1CreatedTimestamp.Add(2 * time.Hour).Format(model.SalesforceDocumentDateTimeLayout),
	}
	err = createDummySalesforceDocument(project.ID, Campaign1, model.SalesforceDocumentTypeNameCampaign)
	assert.Nil(t, err)

	// TestCampaignMember

	campaignMember1CreatedTimestamp := campaign1CreatedTimestamp.Add(10 * time.Minute)
	campaignMember1 := map[string]interface{}{
		"Id":               campaignMember1ID,
		"CampaignId":       campaign1ID,
		"CampaignName":     campaign1Name,
		"LeadId":           campaignMember1LeadID,
		"City":             "Mumbai",
		"Status":           "NotResponded",
		"HasResponded":     false,
		"CreatedDate":      campaignMember1CreatedTimestamp.Format(model.SalesforceDocumentDateTimeLayout),
		"LastModifiedDate": campaignMember1CreatedTimestamp.Add(2 * time.Hour).Format(model.SalesforceDocumentDateTimeLayout),
	}
	err = createDummySalesforceDocument(project.ID, campaignMember1, model.SalesforceDocumentTypeNameCampaignMember)
	assert.Nil(t, err)

	campaignMemberDocuments, status := store.GetStore().GetLatestSalesforceDocumentByID(project.ID, []string{util.GetPropertyValueAsString(campaignMember1ID)},
		model.SalesforceDocumentTypeCampaignMember, campaign1CreatedTimestamp.Add(10*time.Hour).Unix())
	assert.Equal(t, status, http.StatusFound)

	// len(campaignMemberDocuments) > 0 && timestamp sorted desc
	enCampaignMemberProperties, _, err := IntSalesforce.GetSalesforceDocumentProperties(project.ID, &campaignMemberDocuments[0])
	assert.Nil(t, err)

	trackPayload := &SDK.TrackPayload{
		ProjectId:       project.ID,
		EventProperties: *enCampaignMemberProperties,
		RequestSource:   model.UserSourceSalesforce,
	}

	filter1 := model.TouchPointFilter{
		Property:  "$salesforce_campaign_name",
		Operator:  "contains",
		Value:     "Webinar",
		LogicalOp: "AND",
	}

	rule := model.SFTouchPointRule{
		Filters:           []model.TouchPointFilter{filter1},
		TouchPointTimeRef: model.SFCampaignMemberCreated, // SFCampaignMemberResponded
		PropertiesMap:     map[string]string{"$campaign": "$salesforce_campaignmember_campaignname"},
	}

	trackResponse, err := IntSalesforce.CreateTouchPointEvent(project, trackPayload, &campaignMemberDocuments[0], rule)
	assert.Nil(t, err)
	assert.NotNil(t, trackResponse)

	event, errCode := store.GetStore().GetEventById(project.ID, trackResponse.EventId, trackResponse.UserId)
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, event)
	eventPropertiesBytes, err := event.Properties.Value()
	var eventPropertiesMap map[string]interface{}
	_ = json.Unmarshal(eventPropertiesBytes.([]byte), &eventPropertiesMap)
	assert.Equal(t, eventPropertiesMap["$campaign"], campaign1Name)
}

func TestSalesforceOfflineTouchPointDecode(t *testing.T) {

	project, _, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	filter1 := model.TouchPointFilter{
		Property:  "$salesforce_campaign_name",
		Operator:  "contains",
		Value:     "Webinar",
		LogicalOp: "AND",
	}

	rule := model.SFTouchPointRule{
		Filters:           []model.TouchPointFilter{filter1},
		TouchPointTimeRef: model.SFCampaignMemberCreated, // SFCampaignMemberResponded
		PropertiesMap:     map[string]string{"$campaign": "$salesforce_campaignmember_campaignname"},
	}

	// creating manual rule
	touchPointRules := make(map[string][]model.SFTouchPointRule)
	touchPointRules["sf_touch_point_rules"] = []model.SFTouchPointRule{rule}

	// adding json rule
	project.SalesforceTouchPoints = postgres.Jsonb{RawMessage: json.RawMessage(`{"sf_touch_point_rules":[{"filters":[{"pr":"$salesforce_campaign_type","op":"equals","va":"Field Events","lop":"AND"},{"pr":"$salesforce_campaign_name","op":"contains","va":"Sendoso","lop":"AND"}],"touch_point_time_ref":"campaign_member_created_date","properties_map":{"campaign_name":"$salesforce_campaign_name","campaign_id":"$salesforce_campaign_id","source":"$salesforce_campaign_type"}},{"filters":[{"pr":"$salesforce_campaign_type","op":"equals","va":"Others","lop":"AND"}],"touch_point_time_ref":"campaign_member_created_date","properties_map":{"campaign_name":"$salesforce_campaign_name","campaign_id":"$salesforce_campaign_id","source":"$salesforce_campaign_type"}},{"filters":[{"pr":"$salesforce_campaign_type","op":"equals","va":"Nurture","lop":"AND"}],"touch_point_time_ref":"campaign_member_first_responded_date","properties_map":{"campaign_name":"$salesforce_campaign_name","campaign_id":"$salesforce_campaign_id","source":"$salesforce_campaign_type"}}]}`)}
	store.GetStore().UpdateProject(project.ID, project)

	project, errCode := store.GetStore().GetProject(project.ID)
	if errCode != http.StatusFound {
		return
	}
	if &project.SalesforceTouchPoints != nil && !U.IsEmptyPostgresJsonb(&project.SalesforceTouchPoints) {

		var touchPointRules map[string][]model.SFTouchPointRule
		err := U.DecodePostgresJsonbToStructType(&project.SalesforceTouchPoints, &touchPointRules)
		assert.Nil(t, err)

		rules := touchPointRules["sf_touch_point_rules"]

		assert.Equal(t, len(rules), 3)
	}
}

func querySingleEventWithBreakdownByUserProperty(projectID uint64, eventName string, propertyName string, from, to int64) (map[string]interface{}, int) {
	query := model.Query{
		Class: model.QueryClassEvents,
		From:  from,
		To:    to,
		Type:  model.QueryTypeEventsOccurrence,
		EventsWithProperties: []model.QueryEventWithProperties{
			{Name: eventName},
		},
		GroupByProperties: []model.QueryGroupByProperty{
			{
				Property:  propertyName,
				Entity:    model.PropertyEntityUser,
				EventName: eventName,
			},
		},
		EventsCondition: model.EventCondEachGivenEvent,
	}

	results, status := store.GetStore().RunEventsGroupQuery([]model.Query{query}, projectID)
	if status != http.StatusOK {
		return nil, status
	}

	result := results.Results[0]
	resultMap := make(map[string]interface{}, 0)
	for i := range result.Rows {
		row := result.Rows[i]
		resultMap[util.GetPropertyValueAsString(row[2])] = row[3]
	}

	return resultMap, status
}

func TestSalesforceGroups(t *testing.T) {
	project, _, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	accountID1 := "acc1_" + getRandomName()
	accountID2 := "acc2_" + getRandomName()

	createdDate := time.Now().UTC().AddDate(0, 0, -3)
	processRecords := make([]map[string]interface{}, 0)
	processRecordsType := make([]string, 0)
	document := map[string]interface{}{
		"Id":               accountID1,
		"Name":             "account1",
		"CreatedDate":      createdDate.Format(model.SalesforceDocumentDateTimeLayout),
		"LastModifiedDate": createdDate.Add(30 * time.Second).Format(model.SalesforceDocumentDateTimeLayout),
	}
	processRecords = append(processRecords, document)
	processRecordsType = append(processRecordsType, model.SalesforceDocumentTypeNameAccount)

	document = map[string]interface{}{
		"Id":               accountID2,
		"Name":             "account2",
		"CreatedDate":      createdDate.Format(model.SalesforceDocumentDateTimeLayout),
		"LastModifiedDate": createdDate.Add(30 * time.Second).Format(model.SalesforceDocumentDateTimeLayout),
	}
	processRecords = append(processRecords, document)
	processRecordsType = append(processRecordsType, model.SalesforceDocumentTypeNameAccount)

	leadID1 := "acc1_lead1_" + getRandomName()
	leadID2 := "acc2_lead1_" + getRandomName()
	document = map[string]interface{}{
		"Id":                 leadID1,
		"Name":               "lead1",
		"ConvertedAccountId": accountID1,
		"CreatedDate":        createdDate.Format(model.SalesforceDocumentDateTimeLayout),
		"LastModifiedDate":   createdDate.Add(30 * time.Second).Format(model.SalesforceDocumentDateTimeLayout),
	}
	processRecords = append(processRecords, document)
	processRecordsType = append(processRecordsType, model.SalesforceDocumentTypeNameLead)

	document = map[string]interface{}{
		"Id":                 leadID2,
		"Name":               "lead2",
		"ConvertedAccountId": accountID2,
		"CreatedDate":        createdDate.Format(model.SalesforceDocumentDateTimeLayout),
		"LastModifiedDate":   createdDate.Add(30 * time.Second).Format(model.SalesforceDocumentDateTimeLayout),
	}
	processRecords = append(processRecords, document)
	processRecordsType = append(processRecordsType, model.SalesforceDocumentTypeNameLead)

	contactID1 := "acc1_contact1_" + getRandomName()
	contactID2 := "acc2_contact1_" + getRandomName()
	document = map[string]interface{}{
		"Id":               contactID1,
		"Name":             "contact1",
		"AccountId":        accountID1,
		"CreatedDate":      createdDate.Format(model.SalesforceDocumentDateTimeLayout),
		"LastModifiedDate": createdDate.Add(30 * time.Second).Format(model.SalesforceDocumentDateTimeLayout),
	}
	processRecords = append(processRecords, document)
	processRecordsType = append(processRecordsType, model.SalesforceDocumentTypeNameContact)

	document = map[string]interface{}{
		"Id":               contactID2,
		"Name":             "contact2",
		"AccountId":        accountID2,
		"CreatedDate":      createdDate.Format(model.SalesforceDocumentDateTimeLayout),
		"LastModifiedDate": createdDate.Add(30 * time.Second).Format(model.SalesforceDocumentDateTimeLayout),
	}
	processRecords = append(processRecords, document)
	processRecordsType = append(processRecordsType, model.SalesforceDocumentTypeNameContact)

	opportunityID1 := "acc1_opp1_" + getRandomName()
	opportunityID2 := "acc2_opp1_" + getRandomName()
	opportunityID3 := "acc1_opp2_" + getRandomName()
	opportunityID4 := "acc2_opp2_" + getRandomName()

	document = map[string]interface{}{
		"Id":               opportunityID1,
		"Name":             "opportunity1",
		"AccountId":        accountID1,
		"CreatedDate":      createdDate.Format(model.SalesforceDocumentDateTimeLayout),
		"LastModifiedDate": createdDate.Add(30 * time.Second).Format(model.SalesforceDocumentDateTimeLayout),
	}
	processRecords = append(processRecords, document)
	processRecordsType = append(processRecordsType, model.SalesforceDocumentTypeNameOpportunity)

	document = map[string]interface{}{
		"Id":               opportunityID2,
		"Name":             "opportunity2",
		"AccountId":        accountID2,
		"CreatedDate":      createdDate.Format(model.SalesforceDocumentDateTimeLayout),
		"LastModifiedDate": createdDate.Add(30 * time.Second).Format(model.SalesforceDocumentDateTimeLayout),
	}
	processRecords = append(processRecords, document)
	processRecordsType = append(processRecordsType, model.SalesforceDocumentTypeNameOpportunity)

	document = map[string]interface{}{
		"Id":               opportunityID3,
		"Name":             "opportunity3",
		"AccountId":        accountID1,
		"CreatedDate":      createdDate.Format(model.SalesforceDocumentDateTimeLayout),
		"LastModifiedDate": createdDate.Add(30 * time.Second).Format(model.SalesforceDocumentDateTimeLayout),
	}
	processRecords = append(processRecords, document)
	processRecordsType = append(processRecordsType, model.SalesforceDocumentTypeNameOpportunity)

	document = map[string]interface{}{
		"Id":               opportunityID4,
		"Name":             "opportunity4",
		"AccountId":        accountID2,
		"CreatedDate":      createdDate.Format(model.SalesforceDocumentDateTimeLayout),
		"LastModifiedDate": createdDate.Add(30 * time.Second).Format(model.SalesforceDocumentDateTimeLayout),
	}
	processRecords = append(processRecords, document)
	processRecordsType = append(processRecordsType, model.SalesforceDocumentTypeNameOpportunity)

	for i := range processRecords {
		err = createDummySalesforceDocument(project.ID, processRecords[i], processRecordsType[i])
		assert.Nil(t, err, fmt.Sprintf("doc_type %s", processRecordsType[i]))
	}
	enrichStatus, _ := IntSalesforce.Enrich(project.ID)
	assert.Equal(t, project.ID, enrichStatus[0].ProjectID)
	assert.Len(t, enrichStatus, 6) // group account status and group opportunity status
	assert.Equal(t, "success", enrichStatus[0].Status)
	assert.Equal(t, "success", enrichStatus[1].Status)
	assert.Equal(t, "success", enrichStatus[2].Status)
	assert.Equal(t, "success", enrichStatus[3].Status)
	assert.Equal(t, "success", enrichStatus[4].Status)
	assert.Equal(t, "success", enrichStatus[5].Status)

	account1GroupUserId := ""
	account2GroupUserId := ""
	opportunity1GroupUserId := ""
	opportunity2GroupUserId := ""
	opportunity3GroupUserId := ""
	opportunity4GroupUserId := ""
	for i := range processRecords {
		docType := model.GetSalesforceDocTypeByAlias(processRecordsType[i])
		documents, status := store.GetStore().GetLatestSalesforceDocumentByID(project.ID, []string{util.GetPropertyValueAsString(processRecords[i]["Id"])}, docType, 0)
		assert.Equal(t, http.StatusFound, status)
		assert.NotEmpty(t, documents[0].UserID)
		if documents[0].Type == model.SalesforceDocumentTypeAccount || documents[0].Type == model.SalesforceDocumentTypeOpportunity {
			assert.NotEqual(t, "", documents[0].GroupUserID)
			assert.NotEqual(t, "", documents[0].UserID)
			groupUser, _ := store.GetStore().GetUser(project.ID, documents[0].GroupUserID)
			assert.Equal(t, model.UserSourceSalesforce, *groupUser.Source)
			if documents[0].Type == model.SalesforceDocumentTypeOpportunity {
				assert.Equal(t, "", groupUser.Group1ID)
				assert.NotEqual(t, "", groupUser.ID)
				if documents[0].ID == opportunityID1 {
					opportunity1GroupUserId = groupUser.ID
				} else if documents[0].ID == opportunityID2 {
					opportunity2GroupUserId = groupUser.ID
				} else if documents[0].ID == opportunityID3 {
					opportunity3GroupUserId = groupUser.ID
				} else {
					opportunity4GroupUserId = groupUser.ID
				}
			} else {
				if documents[0].ID == accountID1 {
					assert.Equal(t, "account1", groupUser.Group1ID)
					account1GroupUserId = groupUser.ID
				} else {
					assert.Equal(t, "account2", groupUser.Group1ID)
					account2GroupUserId = groupUser.ID
				}
			}

		} else {
			nonGroupUser, _ := store.GetStore().GetUser(project.ID, documents[0].UserID)
			assert.Equal(t, false, *nonGroupUser.IsGroupUser)
			if documents[0].ID == contactID1 || documents[0].ID == leadID1 ||
				documents[0].ID == opportunityID1 || documents[0].ID == opportunityID3 {
				assert.Equal(t, "account1", nonGroupUser.Group1ID)
			} else {
				assert.Equal(t, "account2", nonGroupUser.Group1ID)
			}
			assert.NotEqual(t, "", nonGroupUser.Group1UserID)
		}
	}

	result, status := querySingleEventWithBreakdownByUserProperty(project.ID, U.EVENT_NAME_SALESFORCE_LEAD_CREATED,
		"$salesforce_lead_id", createdDate.Unix()-500, createdDate.Add(30*time.Second).Unix()+500)
	assert.Equal(t, http.StatusOK, status)
	assert.Len(t, result, 2)
	assert.Equal(t, float64(1), result[leadID1])
	assert.Equal(t, float64(1), result[leadID2])

	result, status = querySingleEventWithBreakdownByUserProperty(project.ID, U.EVENT_NAME_SALESFORCE_LEAD_UPDATED,
		"$salesforce_lead_id", createdDate.Unix()-500, createdDate.Add(30*time.Second).Unix()+500)
	assert.Equal(t, http.StatusOK, status)
	assert.Len(t, result, 2)
	assert.Equal(t, float64(2), result[leadID1])
	assert.Equal(t, float64(2), result[leadID2])

	result, status = querySingleEventWithBreakdownByUserProperty(project.ID, U.EVENT_NAME_SALESFORCE_ACCOUNT_CREATED,
		"$salesforce_account_id", createdDate.Unix()-500, createdDate.Add(30*time.Second).Unix()+500)
	assert.Equal(t, http.StatusOK, status)
	assert.Len(t, result, 2)
	assert.Equal(t, float64(1), result[accountID1])
	assert.Equal(t, float64(1), result[accountID2])

	result, status = querySingleEventWithBreakdownByUserProperty(project.ID, U.EVENT_NAME_SALESFORCE_ACCOUNT_UPDATED,
		"$salesforce_account_id", createdDate.Unix()-500, createdDate.Add(30*time.Second).Unix()+500)
	assert.Equal(t, http.StatusOK, status)
	assert.Len(t, result, 2)
	assert.Equal(t, float64(2), result[accountID1])
	assert.Equal(t, float64(2), result[accountID2])

	result, status = querySingleEventWithBreakdownByUserProperty(project.ID, U.GROUP_EVENT_NAME_SALESFORCE_ACCOUNT_CREATED,
		"$salesforce_account_id", createdDate.Unix()-500, createdDate.Add(30*time.Second).Unix()+500)
	assert.Equal(t, http.StatusOK, status)
	assert.Len(t, result, 2)
	assert.Equal(t, float64(1), result[accountID1])
	assert.Equal(t, float64(1), result[accountID2])

	result, status = querySingleEventWithBreakdownByUserProperty(project.ID, U.GROUP_EVENT_NAME_SALESFORCE_ACCOUNT_UPDATED,
		"$salesforce_account_id", createdDate.Unix()-500, createdDate.Add(30*time.Second).Unix()+500)
	assert.Equal(t, http.StatusOK, status)
	assert.Len(t, result, 2)
	assert.Equal(t, float64(2), result[accountID1])
	assert.Equal(t, float64(2), result[accountID2])

	result, status = querySingleEventWithBreakdownByUserProperty(project.ID, U.GROUP_EVENT_NAME_SALESFORCE_OPPORTUNITY_CREATED,
		"$salesforce_opportunity_id", createdDate.Unix()-500, createdDate.Add(30*time.Second).Unix()+500)
	assert.Equal(t, http.StatusOK, status)
	assert.Len(t, result, 4)
	assert.Equal(t, float64(1), result[opportunityID1])
	assert.Equal(t, float64(1), result[opportunityID2])
	assert.Equal(t, float64(1), result[opportunityID3])
	assert.Equal(t, float64(1), result[opportunityID4])

	result, status = querySingleEventWithBreakdownByUserProperty(project.ID, U.GROUP_EVENT_NAME_SALESFORCE_OPPORTUNITY_UPDATED,
		"$salesforce_opportunity_id", createdDate.Unix()-500, createdDate.Add(30*time.Second).Unix()+500)
	assert.Equal(t, http.StatusOK, status)
	assert.Len(t, result, 4)
	assert.Equal(t, float64(2), result[opportunityID1])
	assert.Equal(t, float64(2), result[opportunityID2])
	assert.Equal(t, float64(2), result[opportunityID3])
	assert.Equal(t, float64(2), result[opportunityID4])

	// Two new update on the account. Account name will not be updated on group id
	processRecords = []map[string]interface{}{}
	processRecordsType = []string{}
	document = map[string]interface{}{
		"Id":               accountID1,
		"Name":             "account1.1",
		"city":             "A",
		"CreatedDate":      createdDate.Format(model.SalesforceDocumentDateTimeLayout),
		"LastModifiedDate": createdDate.AddDate(0, 0, 1).Format(model.SalesforceDocumentDateTimeLayout),
	}
	processRecords = append(processRecords, document)
	processRecordsType = append(processRecordsType, model.SalesforceDocumentTypeNameAccount)

	document = map[string]interface{}{
		"Id":               accountID2,
		"Name":             "account2.2",
		"City":             "B",
		"CreatedDate":      createdDate.Format(model.SalesforceDocumentDateTimeLayout),
		"LastModifiedDate": createdDate.AddDate(0, 0, 1).Format(model.SalesforceDocumentDateTimeLayout),
	}
	processRecords = append(processRecords, document)
	processRecordsType = append(processRecordsType, model.SalesforceDocumentTypeNameAccount)

	for i := range processRecords {
		err = createDummySalesforceDocument(project.ID, processRecords[i], processRecordsType[i])
		assert.Nil(t, err, fmt.Sprintf("doc_type %s", processRecordsType[i]))
	}

	// check before for property A
	result, status = querySingleEventWithBreakdownByUserProperty(project.ID, U.GROUP_EVENT_NAME_SALESFORCE_ACCOUNT_UPDATED,
		"$salesforce_account_city", createdDate.AddDate(0, 0, 1).Unix(), createdDate.AddDate(0, 0, 1).Unix()+10)
	assert.Equal(t, http.StatusOK, status)
	assert.Len(t, result, 0)

	enrichStatus, _ = IntSalesforce.Enrich(project.ID)
	assert.Equal(t, project.ID, enrichStatus[0].ProjectID)
	assert.Len(t, enrichStatus, 2) // includes group account status
	assert.Equal(t, "success", enrichStatus[0].Status)
	assert.Equal(t, "success", enrichStatus[1].Status)

	result, status = querySingleEventWithBreakdownByUserProperty(project.ID, U.GROUP_EVENT_NAME_SALESFORCE_ACCOUNT_UPDATED,
		"$salesforce_account_id", createdDate.AddDate(0, 0, 1).Unix(), createdDate.AddDate(0, 0, 1).Unix()+10)
	assert.Equal(t, http.StatusOK, status)
	assert.Len(t, result, 2)
	assert.Equal(t, float64(1), result[accountID1])
	assert.Equal(t, float64(1), result[accountID2])

	result, status = querySingleEventWithBreakdownByUserProperty(project.ID, U.GROUP_EVENT_NAME_SALESFORCE_ACCOUNT_UPDATED,
		"$salesforce_account_city", createdDate.AddDate(0, 0, 1).Unix(), createdDate.AddDate(0, 0, 1).Unix()+10)
	assert.Equal(t, http.StatusOK, status)
	assert.Len(t, result, 2)
	assert.Equal(t, float64(1), result["A"])
	assert.Equal(t, float64(1), result["B"])

	for id, docType := range map[string]int{
		leadID1:        model.SalesforceDocumentTypeLead,
		leadID2:        model.SalesforceDocumentTypeLead,
		contactID1:     model.SalesforceDocumentTypeContact,
		contactID2:     model.SalesforceDocumentTypeContact,
		accountID1:     model.SalesforceDocumentTypeAccount,
		accountID2:     model.SalesforceDocumentTypeAccount,
		opportunityID1: model.SalesforceDocumentTypeOpportunity,
		opportunityID2: model.SalesforceDocumentTypeOpportunity,
		opportunityID3: model.SalesforceDocumentTypeOpportunity,
		opportunityID4: model.SalesforceDocumentTypeOpportunity,
	} {
		documents, status := store.GetStore().GetLatestSalesforceDocumentByID(project.ID, []string{util.GetPropertyValueAsString(id)}, docType, 0)
		assert.Equal(t, http.StatusFound, status)
		lasteDocument := documents[len(documents)-1]
		if docType == model.SalesforceDocumentTypeAccount || docType == model.SalesforceDocumentTypeOpportunity {
			groupUser, status := store.GetStore().GetUser(project.ID, lasteDocument.GroupUserID)
			assert.Equal(t, http.StatusFound, status)
			assert.Equal(t, model.UserSourceSalesforce, *groupUser.Source)

			if docType == model.SalesforceDocumentTypeOpportunity {
				assert.Equal(t, "", groupUser.Group1ID)
				assert.NotEqual(t, "", groupUser.ID)
				if documents[0].ID == opportunityID1 {
					assert.Equal(t, opportunity1GroupUserId, groupUser.ID)
				} else if documents[0].ID == opportunityID2 {
					assert.Equal(t, opportunity2GroupUserId, groupUser.ID)
				} else if documents[0].ID == opportunityID3 {
					assert.Equal(t, opportunity3GroupUserId, groupUser.ID)
				} else {
					assert.Equal(t, opportunity4GroupUserId, groupUser.ID)
				}
			} else {
				if documents[0].ID == accountID1 {
					assert.Equal(t, "account1", groupUser.Group1ID)
					assert.Equal(t, account1GroupUserId, groupUser.ID)
				} else {
					assert.Equal(t, "account2", groupUser.Group1ID)
					assert.Equal(t, account2GroupUserId, groupUser.ID)
				}
			}

		} else {
			assert.Equal(t, "", lasteDocument.GroupUserID)
			nonGroupUser, status := store.GetStore().GetUser(project.ID, lasteDocument.UserID)
			assert.Equal(t, http.StatusFound, status)
			if lasteDocument.ID == leadID1 || lasteDocument.ID == contactID1 {
				assert.Equal(t, "account1", nonGroupUser.Group1ID)
			} else {
				assert.Equal(t, "account2", nonGroupUser.Group1ID)
			}
			assert.NotEqual(t, "", nonGroupUser.Group1UserID)
		}
	}

	// Verify group relationships
	for id, docType := range map[string]int{
		accountID1:     model.SalesforceDocumentTypeAccount,
		accountID2:     model.SalesforceDocumentTypeAccount,
		opportunityID1: model.SalesforceDocumentTypeOpportunity,
		opportunityID2: model.SalesforceDocumentTypeOpportunity,
		opportunityID3: model.SalesforceDocumentTypeOpportunity,
		opportunityID4: model.SalesforceDocumentTypeOpportunity,
	} {
		documents, status := store.GetStore().GetLatestSalesforceDocumentByID(project.ID, []string{util.GetPropertyValueAsString(id)},
			docType, 0)
		assert.Equal(t, http.StatusFound, status)

		if docType == model.SalesforceDocumentTypeAccount {
			groupRelationships, status := store.GetStore().GetGroupRelationshipByUserID(project.ID, documents[0].GroupUserID)
			assert.Equal(t, http.StatusFound, status)
			assert.Equal(t, 2, len(groupRelationships))
			for i := range groupRelationships {
				if groupRelationships[i].LeftGroupUserID == account1GroupUserId {
					if groupRelationships[i].RightGroupUserID == opportunity1GroupUserId {

					} else {
						assert.Equal(t, groupRelationships[i].RightGroupUserID, opportunity3GroupUserId)
					}
				} else {
					if groupRelationships[i].RightGroupUserID == opportunity2GroupUserId {

					} else {
						assert.Equal(t, groupRelationships[i].RightGroupUserID, opportunity4GroupUserId)
					}
				}
			}

		} else {
			groupRelationships, status := store.GetStore().GetGroupRelationshipByUserID(project.ID, documents[0].GroupUserID)
			assert.Equal(t, http.StatusFound, status)
			assert.Equal(t, 1, len(groupRelationships))
			assert.Equal(t, groupRelationships[0].LeftGroupUserID, documents[0].GroupUserID)
			if documents[0].ID == opportunityID1 || documents[0].ID == opportunityID3 {
				assert.Equal(t, groupRelationships[0].RightGroupUserID, account1GroupUserId)
			} else {
				assert.Equal(t, groupRelationships[0].RightGroupUserID, account2GroupUserId)
			}
		}
	}

	// Update already synced account record group status
	accountID3 := "account3"
	document = map[string]interface{}{
		"Id":               accountID3,
		"Name":             "account3",
		"City":             "A",
		"CreatedDate":      createdDate.AddDate(0, 0, -5).Format(model.SalesforceDocumentDateTimeLayout),
		"LastModifiedDate": createdDate.AddDate(0, 0, -5).Format(model.SalesforceDocumentDateTimeLayout),
	}

	err = createDummySalesforceDocument(project.ID, document, model.SalesforceDocumentTypeNameAccount)
	assert.Nil(t, err)
	salesforceDocument := model.SalesforceDocument{
		ID:        accountID3,
		ProjectID: project.ID,
		Type:      model.SalesforceDocumentTypeAccount,
		Action:    model.SalesforceDocumentCreated,
		Timestamp: createdDate.AddDate(0, 0, -5).Unix(),
	}
	status = store.GetStore().UpdateSalesforceDocumentBySyncStatus(project.ID, &salesforceDocument, "123", "1234", "", true)
	assert.Equal(t, http.StatusAccepted, status)

	leadID3 := "lead3"
	document = map[string]interface{}{
		"Id":                 leadID3,
		"Name":               "lead3",
		"ConvertedAccountId": accountID3,
		"City":               "A",
		"CreatedDate":        createdDate.AddDate(0, 0, -2).Format(model.SalesforceDocumentDateTimeLayout),
		"LastModifiedDate":   createdDate.AddDate(0, 0, -2).Format(model.SalesforceDocumentDateTimeLayout),
	}
	err = createDummySalesforceDocument(project.ID, document, model.SalesforceDocumentTypeNameLead)
	assert.Nil(t, err)
	enrichStatus, _ = IntSalesforce.Enrich(project.ID)
	assert.Equal(t, project.ID, enrichStatus[0].ProjectID)
	assert.Len(t, enrichStatus, 1)
	assert.Equal(t, "success", enrichStatus[0].Status)

	documents, status := store.GetStore().GetSyncedSalesforceDocumentByType(project.ID, []string{accountID3}, model.SalesforceDocumentTypeAccount, false)
	assert.Equal(t, http.StatusFound, status)
	assert.NotEqual(t, "", documents[0].GroupUserID)
	groupUser, status := store.GetStore().GetUser(project.ID, documents[0].GroupUserID)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, true, *groupUser.IsGroupUser)
	assert.Equal(t, "account3", groupUser.Group1ID)
	assert.Equal(t, documents[0].GroupUserID, groupUser.ID)
	assert.Equal(t, model.UserSourceSalesforce, *groupUser.Source)
	account3GroupUserID := groupUser.ID

	documents, status = store.GetStore().GetSyncedSalesforceDocumentByType(project.ID, []string{leadID3}, model.SalesforceDocumentTypeLead, false)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, "", documents[0].GroupUserID)
	groupUser, status = store.GetStore().GetUser(project.ID, documents[0].UserID)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, false, *groupUser.IsGroupUser)
	assert.Equal(t, "account3", groupUser.Group1ID)
	assert.Equal(t, account3GroupUserID, groupUser.Group1UserID)
}

func TestSalesforceUserPropertiesOverwrite(t *testing.T) {
	// Initialize the project and the user. Also capture currentTimestamp, futureTimestamp & middleTimestamp.
	currentTimestamp := time.Now().Unix()
	futureTimestamp := currentTimestamp + 10000
	middleTimestamp := currentTimestamp + 1000
	project, user, err := SetupProjectUserReturnDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)
	assert.NotNil(t, user)
	assert.NotEmpty(t, user.Properties)
	_, errCode := store.GetStore().GetUser(project.ID, user.ID)
	assert.Equal(t, http.StatusFound, errCode)

	status := IntSalesforce.CreateOrGetSalesforceEventName(project.ID)
	assert.Equal(t, http.StatusOK, status)

	eventName := U.EVENT_NAME_SALESFORCE_CONTACT_CREATED
	dateTimeProperty := "$salesforce_contact_lastmodifieddate"
	status = store.GetStore().CreatePropertyDetails(project.ID, eventName, dateTimeProperty, U.PropertyTypeDateTime, false, false)
	assert.Equal(t, http.StatusCreated, status)

	eventName = U.EVENT_NAME_SALESFORCE_CONTACT_UPDATED
	status = store.GetStore().CreatePropertyDetails(project.ID, eventName, dateTimeProperty, U.PropertyTypeDateTime, false, false)
	assert.Equal(t, http.StatusCreated, status)

	eventName = ""
	dateTimeProperty = "$salesforce_contact_lastmodifieddate"
	status = store.GetStore().CreatePropertyDetails(project.ID, eventName, dateTimeProperty, U.PropertyTypeDateTime, true, true)
	assert.Equal(t, http.StatusCreated, status)

	// Update user properties lastmodifieddate as middleTimestamp, PropertiesUpdatedTimestamp
	// as futureTimestamp.
	newProperties := &postgres.Jsonb{RawMessage: json.RawMessage([]byte(fmt.Sprintf(
		`{"country": "india", "age": 30.1, "paid": true, "$salesforce_contact_lastmodifieddate": %d}`, middleTimestamp)))}
	_, status = store.GetStore().UpdateUserProperties(project.ID, user.ID,
		newProperties, futureTimestamp)
	assert.Equal(t, http.StatusAccepted, status)
	storedUser, errCode := store.GetStore().GetUser(project.ID, user.ID)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, futureTimestamp, storedUser.PropertiesUpdatedTimestamp)
	var propertiesMap map[string]interface{}
	err = json.Unmarshal((storedUser.Properties).RawMessage, &propertiesMap)
	assert.Nil(t, err)
	storedLastModifiedDate, err := util.GetPropertyValueAsFloat64(propertiesMap["$salesforce_contact_lastmodifieddate"])
	assert.Nil(t, err)
	assert.Equal(t, middleTimestamp, int64(storedLastModifiedDate))

	// Update user property lastmodifieddate as futureTimestamp and PropertiesUpdatedTimestamp as currentTimestamp.
	// Since the source and object-type are blank, the property value and PropertiesUpdatedTimestamp should not get
	// updated.
	newProperties = &postgres.Jsonb{RawMessage: json.RawMessage([]byte(fmt.Sprintf(
		`{"$salesforce_contact_lastmodifieddate": %d}`, futureTimestamp)))}
	_, status = store.GetStore().UpdateUserProperties(project.ID, user.ID,
		newProperties, currentTimestamp)
	assert.Equal(t, http.StatusAccepted, status)
	storedUser, errCode = store.GetStore().GetUser(project.ID, user.ID)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, futureTimestamp, storedUser.PropertiesUpdatedTimestamp)
	var updatedPropertiesMap map[string]interface{}
	err = json.Unmarshal((storedUser.Properties).RawMessage, &updatedPropertiesMap)
	assert.Nil(t, err)
	storedLastModifiedDate, err = util.GetPropertyValueAsFloat64(updatedPropertiesMap["$salesforce_contact_lastmodifieddate"])
	assert.Nil(t, err)
	assert.Equal(t, middleTimestamp, int64(storedLastModifiedDate))

	// Get oldTimestamp, before the futureTimestamp.
	oldTimestamp := futureTimestamp - 1000

	// Update user properties lastmodifieddate as futureTimestamp, PropertiesUpdatedTimestamp as oldTimestamp.
	// lastmodifieddate should get updated with futureTimestamp, but PropertiesUpdatedTimestamp should remain unchanged.
	newProperties = &postgres.Jsonb{RawMessage: json.RawMessage([]byte(fmt.Sprintf(
		`{"country": "india", "age": 30.1, "paid": true, "$salesforce_contact_lastmodifieddate": %d}`, futureTimestamp)))}
	_, status = store.GetStore().UpdateUserPropertiesV2(project.ID, user.ID,
		newProperties, oldTimestamp, SDK.SourceSalesforce, model.SalesforceDocumentTypeNameContact)
	assert.Equal(t, http.StatusAccepted, status)
	storedUser, errCode = store.GetStore().GetUser(project.ID, user.ID)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, futureTimestamp, storedUser.PropertiesUpdatedTimestamp)
	err = json.Unmarshal((storedUser.Properties).RawMessage, &propertiesMap)
	assert.Nil(t, err)
	storedLastModifiedDate, err = util.GetPropertyValueAsFloat64(propertiesMap["$salesforce_contact_lastmodifieddate"])
	assert.Nil(t, err)
	assert.Equal(t, futureTimestamp, int64(storedLastModifiedDate))

	// salesforce record test -> Testing single user
	contactID := U.RandomLowerAphaNumString(5)
	name := U.RandomLowerAphaNumString(5)

	timestampT1 := time.Now().AddDate(0, 0, -20)

	// salesforce record with created == updated.
	jsonData := fmt.Sprintf(`{"Id":"%s", "name":"%s","CreatedDate":"%s", "LastModifiedDate":"%s"}`, contactID, name, timestampT1.UTC().Format(model.SalesforceDocumentDateTimeLayout), timestampT1.UTC().Format(model.SalesforceDocumentDateTimeLayout))
	salesforceDocument := &model.SalesforceDocument{
		ProjectID: project.ID,
		TypeAlias: model.SalesforceDocumentTypeNameContact,
		Value:     &postgres.Jsonb{RawMessage: json.RawMessage([]byte(jsonData))},
	}

	status = store.GetStore().CreateSalesforceDocument(project.ID, salesforceDocument)
	assert.Equal(t, http.StatusCreated, status)

	// Execute enrich job to process the contact created above
	enrichStatus, _ := IntSalesforce.Enrich(project.ID)
	for i := range enrichStatus {
		assert.Equal(t, U.CRM_SYNC_STATUS_SUCCESS, enrichStatus[i].Status)
	}

	// Verification for contact creation.
	createDocument, status := store.GetStore().GetSyncedSalesforceDocumentByType(project.ID, []string{contactID}, model.SalesforceDocumentTypeContact, false)
	assert.Equal(t, http.StatusFound, status)
	user, status = store.GetStore().GetUser(project.ID, createDocument[0].UserID)
	assert.Equal(t, http.StatusFound, status)
	properitesMap := make(map[string]interface{})
	err = json.Unmarshal(user.Properties.RawMessage, &properitesMap)
	assert.Nil(t, err)

	// Verify salesforce_contact_lastmodifieddate is set to timestampT1
	lastmodifieddateProperty := model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceSalesforce, model.SalesforceDocumentTypeNameContact,
		util.PROPERTY_KEY_LAST_MODIFIED_DATE)
	userPropertyValue, exists := properitesMap[lastmodifieddateProperty]
	assert.Equal(t, exists, true)
	assert.Equal(t, float64(timestampT1.Unix()), userPropertyValue)

	// Update user properties (“a”:1) with timestampT3
	newProperties = &postgres.Jsonb{RawMessage: json.RawMessage([]byte(`{"a": 1}`))}
	timestampT3 := timestampT1.AddDate(0, 0, 10)
	_, status = store.GetStore().UpdateUserProperties(project.ID, user.ID,
		newProperties, timestampT3.Unix())
	assert.Equal(t, http.StatusAccepted, status)
	timestampT2 := timestampT1.AddDate(0, 0, 5)

	// update contact record with (properties->‘lastmodified’:timestampT2)
	jsonData = fmt.Sprintf(`{"Id":"%s", "name":"%s","CreatedDate":"%s", "LastModifiedDate":"%s"}`, contactID, name, timestampT1.UTC().Format(model.SalesforceDocumentDateTimeLayout), timestampT2.UTC().Format(model.SalesforceDocumentDateTimeLayout))
	salesforceDocument = &model.SalesforceDocument{
		ProjectID: project.ID,
		TypeAlias: model.SalesforceDocumentTypeNameContact,
		Value:     &postgres.Jsonb{RawMessage: json.RawMessage([]byte(jsonData))},
	}

	status = store.GetStore().CreateSalesforceDocument(project.ID, salesforceDocument)
	assert.Equal(t, http.StatusCreated, status)

	// Execute enrich job to process the contact created above
	enrichStatus, _ = IntSalesforce.Enrich(project.ID)
	for i := range enrichStatus {
		assert.Equal(t, U.CRM_SYNC_STATUS_SUCCESS, enrichStatus[i].Status)
	}

	// Verify salesforce_contact_lastmodifieddate is set to timestampT2 and PropertiesUpdatedTimestamp to timestampT3.
	updateDocument, status := store.GetStore().GetSyncedSalesforceDocumentByType(project.ID, []string{contactID}, model.SalesforceDocumentTypeContact, false)
	assert.Equal(t, http.StatusFound, status)
	user, status = store.GetStore().GetUser(project.ID, updateDocument[0].UserID)
	assert.Equal(t, http.StatusFound, status)
	properitesMap = make(map[string]interface{})
	err = json.Unmarshal(user.Properties.RawMessage, &properitesMap)
	assert.Nil(t, err)
	lastmodifieddateProperty = model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceSalesforce, model.SalesforceDocumentTypeNameContact,
		util.PROPERTY_KEY_LAST_MODIFIED_DATE)
	userPropertyValue, exists = properitesMap[lastmodifieddateProperty]
	assert.Equal(t, exists, true)
	assert.Equal(t, float64(timestampT2.Unix()), userPropertyValue)
	assert.Equal(t, timestampT3.Unix(), user.PropertiesUpdatedTimestamp)

	// Salesforce record test -> Testing multi-user by customer-user-id
	createDocumentIDU1 := rand.Intn(100)
	cuid_first := getRandomEmail()

	// create contact record createDocumentIDU1 with (properties->‘lastmodified’:timestampT1) and email property
	// ("email": cuid_first)
	jsonData = fmt.Sprintf(`{"Id":"%s", "name":"%s","CreatedDate":"%s", "LastModifiedDate":"%s", "email":"%s"}`, fmt.Sprintf("%v", createDocumentIDU1), name, timestampT1.UTC().Format(model.SalesforceDocumentDateTimeLayout), timestampT1.UTC().Format(model.SalesforceDocumentDateTimeLayout), cuid_first)
	salesforceDocument = &model.SalesforceDocument{
		ProjectID: project.ID,
		TypeAlias: model.SalesforceDocumentTypeNameContact,
		Value:     &postgres.Jsonb{RawMessage: json.RawMessage([]byte(jsonData))},
	}

	status = store.GetStore().CreateSalesforceDocument(project.ID, salesforceDocument)
	assert.Equal(t, http.StatusCreated, status)

	// Execute enrich job to process the contact created above
	enrichStatus, _ = IntSalesforce.Enrich(project.ID)
	for i := range enrichStatus {
		assert.Equal(t, U.CRM_SYNC_STATUS_SUCCESS, enrichStatus[i].Status)
	}

	// Create normal user U2 (createUserU2) with same email property as that of createDocumentIDU1 ("email": cuid_first)
	userU2, errCode1 := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: cuid_first, JoinTimestamp: timestampT1.Unix(), Source: model.GetRequestSourcePointer(model.UserSourceSalesforce)})
	assert.Equal(t, http.StatusCreated, errCode1)

	// Verify lastmodifieddate user property of userU2 to be timestampT1, which is same as createDocumentIDU1
	user, status = store.GetStore().GetUser(project.ID, userU2)
	assert.Equal(t, http.StatusFound, status)
	properitesMap = make(map[string]interface{})
	err = json.Unmarshal(user.Properties.RawMessage, &properitesMap)
	assert.Nil(t, err)
	lastmodifieddateProperty = model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceSalesforce, model.SalesforceDocumentTypeNameContact,
		util.PROPERTY_KEY_LAST_MODIFIED_DATE)
	userPropertyValue, exists = properitesMap[lastmodifieddateProperty]
	assert.Equal(t, exists, true)
	assert.Equal(t, float64(timestampT1.Unix()), userPropertyValue)

	// Update user properties (“a”:1) with timestampT3 for userU2
	newProperties = &postgres.Jsonb{RawMessage: json.RawMessage([]byte(`{"a": 1}`))}
	_, status = store.GetStore().UpdateUserProperties(project.ID, user.ID,
		newProperties, timestampT3.Unix())
	assert.Equal(t, http.StatusAccepted, status)

	// create contact updated record for createDocumentIDU1 with (properties->‘lastmodified’:timestampT2)
	jsonData = fmt.Sprintf(`{"Id":"%s", "name":"%s","CreatedDate":"%s", "LastModifiedDate":"%s", "email":"%s"}`, fmt.Sprintf("%v", createDocumentIDU1), name, timestampT1.UTC().Format(model.SalesforceDocumentDateTimeLayout), timestampT2.UTC().Format(model.SalesforceDocumentDateTimeLayout), cuid_first)
	salesforceDocument = &model.SalesforceDocument{
		ProjectID: project.ID,
		TypeAlias: model.SalesforceDocumentTypeNameContact,
		Value:     &postgres.Jsonb{RawMessage: json.RawMessage([]byte(jsonData))},
	}

	status = store.GetStore().CreateSalesforceDocument(project.ID, salesforceDocument)
	assert.Equal(t, http.StatusCreated, status)

	// Execute enrich job to process the contact created above
	enrichStatus, _ = IntSalesforce.Enrich(project.ID)
	for i := range enrichStatus {
		assert.Equal(t, U.CRM_SYNC_STATUS_SUCCESS, enrichStatus[i].Status)
	}

	// Verify salesforce_contact_lastmodifieddate is set to timestampT2 for both createDocumentIDU1 and userU2.
	// Verify PropertiesUpdatedTimestamp is set to timestampT3 for both createDocumentIDU1 and userU2.
	updateDocument, status = store.GetStore().GetSyncedSalesforceDocumentByType(project.ID, []string{fmt.Sprintf("%v", createDocumentIDU1)}, model.SalesforceDocumentTypeContact, false)
	assert.Equal(t, http.StatusFound, status)
	user, status = store.GetStore().GetUser(project.ID, updateDocument[0].UserID)
	assert.Equal(t, http.StatusFound, status)
	properitesMap = make(map[string]interface{})
	err = json.Unmarshal(user.Properties.RawMessage, &properitesMap)
	assert.Nil(t, err)
	lastmodifieddateProperty = model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceSalesforce, model.SalesforceDocumentTypeNameContact,
		util.PROPERTY_KEY_LAST_MODIFIED_DATE)
	userPropertyValue, exists = properitesMap[lastmodifieddateProperty]
	assert.Equal(t, exists, true)
	assert.Equal(t, float64(timestampT2.Unix()), userPropertyValue)
	assert.Equal(t, timestampT3.Unix(), user.PropertiesUpdatedTimestamp)

	user, status = store.GetStore().GetUser(project.ID, userU2)
	assert.Equal(t, http.StatusFound, status)
	properitesMap = make(map[string]interface{})
	err = json.Unmarshal(user.Properties.RawMessage, &properitesMap)
	assert.Nil(t, err)
	lastmodifieddateProperty = model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceSalesforce, model.SalesforceDocumentTypeNameContact,
		util.PROPERTY_KEY_LAST_MODIFIED_DATE)
	userPropertyValue, exists = properitesMap[lastmodifieddateProperty]
	assert.Equal(t, exists, true)
	assert.Equal(t, float64(timestampT2.Unix()), userPropertyValue)
	assert.Equal(t, timestampT3.Unix(), user.PropertiesUpdatedTimestamp)
}

func TestSalesforceOpportunityLateIdentification(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	opportunityID := "1"
	createdDate := time.Now().UTC().AddDate(0, 0, -3)
	document := map[string]interface{}{
		"Id":               opportunityID,
		"Name":             "opportunity",
		"CreatedDate":      createdDate.Format(model.SalesforceDocumentDateTimeLayout),
		"LastModifiedDate": createdDate.Format(model.SalesforceDocumentDateTimeLayout),
	}

	err = createDummySalesforceDocument(project.ID, document, model.SalesforceDocumentTypeNameOpportunity)
	assert.Nil(t, err)

	enrichStatus, _ := IntSalesforce.Enrich(project.ID)
	assert.Equal(t, project.ID, enrichStatus[0].ProjectID)
	assert.Len(t, enrichStatus, 2)
	assert.Equal(t, "success", enrichStatus[0].Status)
	assert.Equal(t, "success", enrichStatus[0].Status) // groups opportunity

	documents, status := store.GetStore().GetSyncedSalesforceDocumentByType(project.ID, []string{opportunityID}, model.SalesforceDocumentTypeOpportunity, false)
	assert.Equal(t, http.StatusFound, status)
	assert.Len(t, documents, 1)
	assert.NotEqual(t, "", documents[0].UserID)
	assert.NotEqual(t, "", documents[0].GroupUserID)

	opportunityUserdID := documents[0].UserID
	opportunityGroupUserdID := documents[0].GroupUserID
	user, status := store.GetStore().GetUser(project.ID, opportunityUserdID)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, "", user.CustomerUserId)

	contactID := "2"
	contactEmail := getRandomEmail()
	createdDate = time.Now().UTC().AddDate(0, 0, -3)
	document = map[string]interface{}{
		"Id":               contactID,
		"Name":             "ContactUser",
		"Email":            contactEmail,
		"CreatedDate":      createdDate.Format(model.SalesforceDocumentDateTimeLayout),
		"LastModifiedDate": createdDate.Format(model.SalesforceDocumentDateTimeLayout),
	}

	err = createDummySalesforceDocument(project.ID, document, model.SalesforceDocumentTypeNameContact)
	assert.Nil(t, err)

	document = map[string]interface{}{
		"Id":               opportunityID,
		"Name":             "opportunity",
		"CreatedDate":      createdDate.Format(model.SalesforceDocumentDateTimeLayout),
		"LastModifiedDate": createdDate.Add(30 * time.Second).Format(model.SalesforceDocumentDateTimeLayout),
		model.SalesforceChildRelationshipNameOpportunityContactRoles: map[string]interface{}{
			"records": []map[string]interface{}{
				{
					"ContactId": contactID,
					"IsPrimary": true,
				},
			},
		},
	}

	err = createDummySalesforceDocument(project.ID, document, model.SalesforceDocumentTypeNameOpportunity)
	assert.Nil(t, err)

	enrichStatus, _ = IntSalesforce.Enrich(project.ID)
	assert.Equal(t, project.ID, enrichStatus[0].ProjectID)
	assert.Len(t, enrichStatus, 3)
	assert.Equal(t, "success", enrichStatus[0].Status)
	assert.Equal(t, "success", enrichStatus[0].Status) // groups opportunity
	assert.Equal(t, "success", enrichStatus[0].Status)

	documents, status = store.GetStore().GetSyncedSalesforceDocumentByType(project.ID, []string{opportunityID}, model.SalesforceDocumentTypeOpportunity, false)
	assert.Equal(t, http.StatusFound, status)
	assert.Len(t, documents, 2)
	assert.Equal(t, opportunityUserdID, documents[0].UserID)
	assert.Equal(t, opportunityGroupUserdID, documents[0].GroupUserID)

	user, status = store.GetStore().GetUser(project.ID, opportunityUserdID)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, contactEmail, user.CustomerUserId)
}

func TestSalesforceCampaignMemberCampaignAssociation(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	status := IntSalesforce.CreateOrGetSalesforceEventName(project.ID)
	assert.Equal(t, http.StatusOK, status)

	// datetime property details
	status = store.GetStore().CreatePropertyDetails(project.ID, U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_CREATED, "$salesforce_campaignmember_createddate", U.PropertyTypeDateTime, false, false)
	assert.Equal(t, http.StatusCreated, status)
	status = store.GetStore().CreatePropertyDetails(project.ID, U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_UPDATED, "$salesforce_campaignmember_createddate", U.PropertyTypeDateTime, false, false)
	assert.Equal(t, http.StatusCreated, status)
	status = store.GetStore().CreatePropertyDetails(project.ID, U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_CREATED, "$salesforce_campaignmember_lastmodifieddate", U.PropertyTypeDateTime, false, false)
	assert.Equal(t, http.StatusCreated, status)
	status = store.GetStore().CreatePropertyDetails(project.ID, U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_UPDATED, "$salesforce_campaignmember_lastmodifieddate", U.PropertyTypeDateTime, false, false)
	assert.Equal(t, http.StatusCreated, status)

	campaignID := "1"
	campaignName := "campaign"
	campaignMemberID := "campaignMember"
	campaign1CreatedTimestamp := time.Now().AddDate(0, 0, -1)

	for i := range []int{1, 2, 3, 4, 5, 6} {
		document := map[string]interface{}{
			"Id":   campaignID,
			"Name": campaignName,
			"CampaignMembers": IntSalesforce.RelationshipCampaignMember{
				Records: []IntSalesforce.RelationshipCampaignMemberRecord{
					{
						ID: campaignMemberID,
					},
				},
			},
			"value":            i,
			"CreatedDate":      campaign1CreatedTimestamp.Format(model.SalesforceDocumentDateTimeLayout),
			"LastModifiedDate": campaign1CreatedTimestamp.Add(time.Duration(i) * time.Hour).Format(model.SalesforceDocumentDateTimeLayout),
		}

		err = createDummySalesforceDocument(project.ID, document, model.SalesforceDocumentTypeNameCampaign)
		assert.Nil(t, err)
	}

	contactID := "2"
	contactEmail := getRandomEmail()
	contactcreatedDate := time.Now().UTC().AddDate(0, 0, -3)
	contact := map[string]interface{}{
		"Id":               contactID,
		"Name":             "ContactUser",
		"Email":            contactEmail,
		"CreatedDate":      contactcreatedDate.Format(model.SalesforceDocumentDateTimeLayout),
		"LastModifiedDate": contactcreatedDate.Format(model.SalesforceDocumentDateTimeLayout),
	}
	err = createDummySalesforceDocument(project.ID, contact, model.SalesforceDocumentTypeNameContact)
	assert.Nil(t, err)

	// campaign member timestamp 1 hour ahead of campaign
	campaignMemberCreatedTimestamp := campaign1CreatedTimestamp.Add(1 * time.Hour)
	campaignMember := map[string]interface{}{
		"Id":               campaignMemberID,
		"CampaignId":       campaignID,
		"ContactId":        contactID,
		"CreatedDate":      campaignMemberCreatedTimestamp.Format(model.SalesforceDocumentDateTimeLayout),
		"LastModifiedDate": campaignMemberCreatedTimestamp.Add(3 * time.Hour).Add(20 * time.Minute).Format(model.SalesforceDocumentDateTimeLayout),
	}
	err = createDummySalesforceDocument(project.ID, campaignMember, model.SalesforceDocumentTypeNameCampaignMember)
	assert.Nil(t, err)

	enrichStatus, _ := IntSalesforce.Enrich(project.ID)
	assert.Equal(t, project.ID, enrichStatus[0].ProjectID)
	assert.Len(t, enrichStatus, 2)
	// campaign member and contact
	assert.Equal(t, "success", enrichStatus[0].Status)
	assert.Equal(t, "success", enrichStatus[0].Status)

	documents, status := store.GetStore().GetSyncedSalesforceDocumentByType(project.ID, []string{campaignMemberID}, model.SalesforceDocumentTypeCampaignMember, false)
	assert.Equal(t, http.StatusFound, status)
	assert.Len(t, documents, 1)
	assert.NotEqual(t, "", documents[0].UserID)
	campaignMemberUserID := documents[0].UserID

	documents, status = store.GetStore().GetSyncedSalesforceDocumentByType(project.ID, []string{contactID}, model.SalesforceDocumentTypeContact, false)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, campaignMemberUserID, documents[0].UserID)

	query := model.Query{
		From: campaign1CreatedTimestamp.Unix() - 500,
		To:   campaign1CreatedTimestamp.AddDate(0, 0, 1).Unix() + 500,
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_CREATED,
				Properties: []model.QueryProperty{},
			},
			{
				Name:       U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_UPDATED,
				Properties: []model.QueryProperty{},
			},
		},
		Class: model.QueryClassInsights,

		Type:            model.QueryTypeEventsOccurrence,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	// test using query
	result, errCode, _ := store.GetStore().Analyze(project.ID, query)
	assert.Equal(t, http.StatusOK, errCode)
	rows := result.Rows
	sort.Slice(rows, func(i, j int) bool {
		p1, _ := U.GetPropertyValueAsFloat64(rows[i][0])
		p2, _ := U.GetPropertyValueAsFloat64(rows[j][0])
		return p1 < p2
	})
	assert.Equal(t, U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_CREATED, rows[0][0])
	assert.Equal(t, float64(1), rows[0][1])
	assert.Equal(t, U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_UPDATED, rows[1][0])
	assert.Equal(t, float64(2), rows[1][1])

	query = model.Query{
		From: campaign1CreatedTimestamp.Unix() - 500,
		To:   campaign1CreatedTimestamp.AddDate(0, 0, 1).Unix() + 500,
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_CREATED,
				Properties: []model.QueryProperty{},
			},
			{
				Name:       U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_UPDATED,
				Properties: []model.QueryProperty{},
			},
		},
		GroupByProperties: []model.QueryGroupByProperty{
			{
				Entity:   model.PropertyEntityEvent,
				Property: "$timestamp",
			},
			{
				Entity:   model.PropertyEntityEvent,
				Property: "$salesforce_campaignmember_createddate",
			},
			{
				Entity:   model.PropertyEntityEvent,
				Property: "$salesforce_campaignmember_lastmodifieddate",
			},
			{
				Entity:   model.PropertyEntityEvent,
				Property: "$salesforce_campaign_name",
			},
			{
				Entity:   model.PropertyEntityEvent,
				Property: "$salesforce_campaign_value",
			},
		},
		Class:           model.QueryClassEvents,
		Type:            model.QueryTypeEventsOccurrence,
		EventsCondition: model.EventCondAnyGivenEvent,
	}

	result, errCode, _ = store.GetStore().Analyze(project.ID, query)
	assert.Equal(t, http.StatusOK, errCode)
	rows = result.Rows
	sort.Slice(rows, func(i, j int) bool {
		eventNameI := U.GetPropertyValueAsString(rows[i][0])
		eventNameJ := U.GetPropertyValueAsString(rows[j][0])
		eventNameITimestamp, _ := U.GetPropertyValueAsFloat64(rows[i][1])
		eventNameJTimestamp, _ := U.GetPropertyValueAsFloat64(rows[j][1])
		if eventNameI < eventNameJ {
			return true
		}

		if eventNameI > eventNameJ {
			return false
		}

		return eventNameITimestamp < eventNameJTimestamp
	})

	campaign1CreatedTimestampStr := fmt.Sprintf("%d", campaignMemberCreatedTimestamp.Unix())
	campaign1ModifiedTimestampStr := fmt.Sprintf("%d", campaignMemberCreatedTimestamp.Add(3*time.Hour).Add(20*time.Minute).Unix())
	// campaign member created
	assert.Equal(t, U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_CREATED, rows[0][0])
	assert.Equal(t, campaign1CreatedTimestampStr, rows[0][1])
	assert.Equal(t, campaign1CreatedTimestampStr, rows[0][2])
	assert.Equal(t, campaign1ModifiedTimestampStr, rows[0][3])
	assert.Equal(t, campaignName, rows[0][4])
	assert.Equal(t, "4", rows[0][5])

	// campaign member updated
	assert.Equal(t, U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_UPDATED, rows[1][0])
	assert.Equal(t, campaign1CreatedTimestampStr, rows[1][1])
	assert.Equal(t, campaign1CreatedTimestampStr, rows[1][2])
	assert.Equal(t, campaign1ModifiedTimestampStr, rows[1][3])
	assert.Equal(t, campaignName, rows[0][4])
	assert.Equal(t, "4", rows[1][5])

	assert.Equal(t, U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_UPDATED, rows[2][0])
	assert.Equal(t, campaign1ModifiedTimestampStr, rows[2][1])
	assert.Equal(t, campaign1CreatedTimestampStr, rows[2][2])
	assert.Equal(t, campaign1ModifiedTimestampStr, rows[2][3])
	assert.Equal(t, campaignName, rows[2][4])
	assert.Equal(t, "4", rows[2][5])
}
