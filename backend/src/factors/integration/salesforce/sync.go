package salesforce

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"factors/config"
	C "factors/config"
	"factors/model/model"
	"factors/model/store"
	"factors/util"
	U "factors/util"

	"github.com/jinzhu/now"
	log "github.com/sirupsen/logrus"
)

const (
	// InstanceURL field instance_url
	InstanceURL                = "instance_url"
	salesforceDataServiceRoute = "/services/data/"
	/*
		refer website for api end of life
		https://developer.salesforce.com/docs/atlas.en-us.api_rest.meta/api_rest/api_rest_eol.html
	*/
	salesforceAPIVersion20 = "v20.0"
	salesforceAPIVersion54 = "v54.0"
)

//Salesforce API structs
type field map[string]interface{}

// Describe structure for salesforce describe API
type Describe struct {
	Custom bool    `json:"custom"`
	Fields []field `json:"fields"`
}

// QueryResponse structure for query API
type QueryResponse struct {
	TotalSize      int                      `json:"totalSize"`
	Done           bool                     `json:"done"`
	Records        []model.SalesforceRecord `json:"records"`
	NextRecordsURL string                   `json:"nextRecordsUrl"`
}

// ObjectStatus represents sync info from query to db
type ObjectStatus struct {
	ProjetID      int64          `json:"project_id"`
	Status        string         `json:"status"`
	DocType       string         `json:"doc_type"`
	TotalRecords  int            `json:"total_records"`
	Message       string         `json:"message,omitempty"`
	SyncAll       bool           `json:"syncall"`
	Failures      []string       `json:"failures,omitempty"`
	TotalAPICalls map[string]int `json:"total_api_calls"`
}

// JobStatus list all success and failed while sync from salesforce to db
type JobStatus struct {
	Status   string         `json:"status"`
	Success  []ObjectStatus `json:"success"`
	Failures []ObjectStatus `json:"failures"`
}

// OpportunityLeadID lead id in opportunity
const OpportunityLeadID = "opportunity_to_lead"
const OpportunityMultipleLeadID = "opportunity_to_multiple_lead"

func GetSalesforceAPIVersion(projectID int64) string {
	if C.AllowSalesforcev54APIByProjectID(projectID) {
		return salesforceAPIVersion54
	}
	return salesforceAPIVersion20
}

func buildSalesforceGETRequest(url, accessToken string) (*http.Request, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer"+" "+accessToken)
	return req, nil
}

// GETRequest performs GET request on provided url with access token
func GETRequest(url, accessToken string) (*http.Response, error) {
	req, err := buildSalesforceGETRequest(url, accessToken)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Timeout: 10 * time.Minute,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// DataServiceError impelements error interface for salesforce data api error
type DataServiceError interface{}

func getSalesforceObjectDescription(projectID int64, objectName, accessToken, instanceURL string) (*Describe, error) {
	if objectName == "" || accessToken == "" || instanceURL == "" {
		return nil, errors.New("missing required fields")
	}

	url := instanceURL + salesforceDataServiceRoute + GetSalesforceAPIVersion(projectID) + "/sobjects/" + objectName + "/describe"
	resp, err := GETRequest(url, accessToken)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errBody []DataServiceError
		json.NewDecoder(resp.Body).Decode(&errBody)

		return nil, fmt.Errorf("error while getting object description  %+v", errBody)
	}

	var jsonRespone Describe
	err = json.NewDecoder(resp.Body).Decode(&jsonRespone)
	if err != nil {
		return nil, err
	}

	return &jsonRespone, nil
}

func GetSalesforcePropertiesByDataType(projectID int64, dataType string, docTypes []int, accessToken, instanceURL string) (map[int]*map[string]bool, int) {
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "doc_type": docTypes})

	if projectID == 0 || len(docTypes) == 0 {
		logCtx.Error("Missing project_id or doc_types")
		return nil, http.StatusBadRequest
	}

	if accessToken == "" || instanceURL == "" {
		logCtx.Error("Missing salesforce access token or instance URL")
		return nil, http.StatusBadRequest
	}

	propertiesByObjectType := make(map[int]*map[string]bool)
	for i := range docTypes {
		typeAlias := model.GetSalesforceAliasByDocType(docTypes[i])
		describe, err := getSalesforceObjectDescription(projectID, typeAlias, accessToken, instanceURL)
		if err != nil {
			logCtx.WithError(err).Errorf("Failed to get salesforce object description.")
			return nil, http.StatusInternalServerError
		}

		properties := make(map[string]bool)
		for _, property := range describe.Fields {
			name := U.GetPropertyValueAsString(property["name"])
			fieldType := U.GetPropertyValueAsString(property["type"])
			if fieldType == dataType {
				properties[name] = true
			}
		}
		propertiesByObjectType[docTypes[i]] = &properties
	}

	return propertiesByObjectType, http.StatusOK
}

func getFieldsListFromDescription(description *Describe) ([]string, error) {
	var objectFields []string
	objectFieldDescriptions := description.Fields

	if len(description.Fields) == 0 {
		return objectFields, errors.New("invalid fileds on description")
	}

	for _, fieldDescription := range objectFieldDescriptions {
		if fieldName, ok := fieldDescription["name"].(string); ok {
			objectFields = append(objectFields, fieldName)
		}
	}

	if len(objectFields) == 0 {
		return objectFields, errors.New("empty field list")
	}

	return objectFields, nil
}

func getSalesforceDataByQuery(projectID int64, query, accessToken, instanceURL, dateTime string) (*QueryResponse, error) {
	if query == "" || accessToken == "" || instanceURL == "" {
		return nil, errors.New("missing required fields")
	}

	queryURL := instanceURL + salesforceDataServiceRoute + GetSalesforceAPIVersion(projectID) + "/query?q=" + query

	if dateTime != "" {
		queryURL = queryURL + "+" + "WHERE" + "+" + "LastModifiedDate" + url.QueryEscape(">"+dateTime)
	}

	resp, err := GETRequest(queryURL, accessToken)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errBody []DataServiceError
		if err := json.NewDecoder(resp.Body).Decode(&errBody); err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("error while query data %+v %d", errBody, resp.StatusCode)
	}

	var jsonResponse QueryResponse
	err = json.NewDecoder(resp.Body).Decode(&jsonResponse)
	if err != nil {
		return nil, errors.New("failed to decode response")
	}
	return &jsonResponse, nil
}

// DataClient salesforce data client handles data query from salesforce
type DataClient struct {
	ProjectID      int64
	accessToken    string
	instanceURL    string
	isFirstRun     bool
	nextBatchRoute string
	queryURL       string
	APICall        int
	ObjectName     string
}

// NewSalesforceDataClient create new instance of DataClient for fetching data from salesforce
func NewSalesforceDataClient(accessToken string, instanceURL string) (*DataClient, error) {
	if accessToken == "" || instanceURL == "" {
		return nil, errors.New("missing requied field")
	}

	dataClient := &DataClient{
		accessToken: accessToken,
		instanceURL: instanceURL,
		isFirstRun:  true,
	}

	return dataClient, nil
}

func getSalesforceObjectFieldlList(projectID int64, objectName, accessToken, instanceURL string) ([]string, error) {
	if objectName == "" || accessToken == "" || instanceURL == "" {
		return nil, errors.New("missing required field")
	}

	description, err := getSalesforceObjectDescription(projectID, objectName, accessToken, instanceURL)
	if err != nil {
		return nil, fmt.Errorf("failed to getSalesforceObjectDescription %s", err)
	}

	fields, err := getFieldsListFromDescription(description)
	if err != nil || len(fields) < 1 {
		return nil, fmt.Errorf("failed to getFieldsListFromDescription %s", err)
	}

	return fields, nil
}

func (s *DataClient) getRecordByObjectNameANDFilter(projectID int64, objectName, filterSmnt string) (*DataClient, error) {
	fields, err := getSalesforceObjectFieldlList(projectID, objectName, s.accessToken, s.instanceURL)
	if err != nil {
		return nil, err
	}

	fieldList := strings.Join(fields, ",")
	queryStmnt := fmt.Sprintf("SELECT+%s+FROM+%s+WHERE+%s", fieldList, objectName, url.QueryEscape(filterSmnt))
	queryURL := s.instanceURL + salesforceDataServiceRoute + GetSalesforceAPIVersion(projectID) + "/query?q=" + queryStmnt
	dataClient := &DataClient{
		ProjectID:      projectID,
		accessToken:    s.accessToken,
		instanceURL:    s.instanceURL,
		queryURL:       queryURL,
		isFirstRun:     true,
		nextBatchRoute: "",
		ObjectName:     objectName,
	}

	return dataClient, nil
}

func (s *DataClient) getRecordByObjectNameANDStartTimestamp(projectID int64, objectName string, lookbackTimestamp int64) (*DataClient, error) {
	fields, err := getSalesforceObjectFieldlList(projectID, objectName, s.accessToken, s.instanceURL)
	if err != nil {
		return nil, err
	}

	fieldList := strings.Join(fields, ",")
	// append all campaign memebers to campaign object
	if objectName == model.SalesforceDocumentTypeNameCampaign {
		fieldList = fieldList + ",(+SELECT+id+from+campaignmembers+)"
	}

	// append all opportunity contact roles to opportunity object
	if objectName == model.SalesforceDocumentTypeNameOpportunity {
		fieldList = fieldList + ",(+SELECT+id,isPrimary,ContactId,OpportunityId,Role+from+" + model.SalesforceChildRelationshipNameOpportunityContactRoles + "+)"
	}

	// append RelationId and Type to task or event object
	if objectName == model.SalesforceDocumentTypeNameTask || objectName == model.SalesforceDocumentTypeNameEvent {
		fieldList = fieldList + ",who.Id,who.Type"
	}

	queryStmnt := fmt.Sprintf("SELECT+%s+FROM+%s", fieldList, objectName)
	queryURL := s.instanceURL + salesforceDataServiceRoute + GetSalesforceAPIVersion(projectID) + "/query?q=" + queryStmnt

	if lookbackTimestamp > 0 {
		t := time.Unix(lookbackTimestamp, 0)
		sfFormatedTime := t.UTC().Format(model.SalesforceDocumentDateTimeLayout)
		queryURL = queryURL + "+" + "WHERE" + "+" + "LastModifiedDate" + url.QueryEscape(">"+sfFormatedTime)
	}

	queryStmnt = queryStmnt + "+" + url.QueryEscape("ORDER BY LastModifiedDate ASC")

	dataClient := &DataClient{
		ProjectID:      projectID,
		accessToken:    s.accessToken,
		instanceURL:    s.instanceURL,
		queryURL:       queryURL,
		isFirstRun:     true,
		nextBatchRoute: "",
		ObjectName:     objectName,
	}

	return dataClient, nil
}

func (s *DataClient) getNextBatch() ([]model.SalesforceRecord, bool, error) {
	if s.accessToken == "" || s.instanceURL == "" {
		return nil, true, errors.New("missing url parameters")
	}

	if !s.isFirstRun && s.nextBatchRoute == "" {
		return nil, true, nil
	}

	queryURL := ""
	if s.isFirstRun {
		if s.nextBatchRoute != "" {
			return nil, true, errors.New("invalid nextBatchRoute on first run for salesforce data client")
		}

		queryURL = s.queryURL
	} else {
		if s.nextBatchRoute == "" {
			return nil, true, errors.New("invalid nextBatchRoute in salesforce data client")
		}

		queryURL = s.instanceURL + s.nextBatchRoute
	}

	res, err := s.getRequest(queryURL)
	if err != nil {
		log.WithFields(log.Fields{"url": queryURL}).WithError(err).Warn("Failed to get salesforce data.")
		return nil, true, err
	}

	s.nextBatchRoute = res.NextRecordsURL
	s.isFirstRun = false
	log.WithFields(log.Fields{"project_id": s.ProjectID, "object_name": s.ObjectName, "total_api_call": s.APICall, "records": len(res.Records)}).
		Info("Downloaded salesforce objects.")

	return res.Records, res.Done, nil
}

func (s *DataClient) getRequest(queryURL string) (*QueryResponse, error) {
	resp, err := GETRequest(queryURL, s.accessToken)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errBody []DataServiceError
		if err := json.NewDecoder(resp.Body).Decode(&errBody); err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("error while query data %+v %d", errBody, resp.StatusCode)
	}

	var jsonResponse QueryResponse
	err = json.NewDecoder(resp.Body).Decode(&jsonResponse)
	if err != nil {
		return nil, errors.New("failed to decode response")
	}
	s.APICall++

	return &jsonResponse, nil
}

// GetObjectRecordsByIDs get list of records by Id and object type
func (s *DataClient) GetObjectRecordsByIDs(projectID int64, objectName string, IDs []string) (*DataClient, error) {
	if objectName == "" {
		return nil, errors.New("missing required fields")
	}

	fields, err := getSalesforceObjectFieldlList(projectID, objectName, s.accessToken, s.instanceURL)
	if err != nil {
		return nil, err
	}

	fieldList := strings.Join(fields, ",")
	idList := "'" + strings.Join(IDs, "','") + "'"

	queryStmnt := fmt.Sprintf("SELECT+%s+FROM+%s+WHERE+Id+IN+(%s)", fieldList, objectName, idList)
	queryURL := s.instanceURL + salesforceDataServiceRoute + GetSalesforceAPIVersion(projectID) + "/query?q=" + queryStmnt

	dataClient := &DataClient{
		ProjectID:      projectID,
		accessToken:    s.accessToken,
		instanceURL:    s.instanceURL,
		isFirstRun:     true,
		queryURL:       queryURL,
		nextBatchRoute: "",
		ObjectName:     objectName,
	}

	return dataClient, nil
}

func getCampaingMemberIDsFromCampaign(properties *model.SalesforceRecord) ([]string, error) {
	memberIDs := make([]string, 0)
	if campaignMembersInt, exist := (*properties)[model.SalesforceChildRelationshipNameCampaignMembers]; exist && campaignMembersInt != nil {
		campaignMembers, ok := campaignMembersInt.(map[string]interface{})
		if !ok {
			return nil, errors.New("failed to typecast campaignmemebers to map")
		}

		recordsInt, ok := campaignMembers["records"].([]interface{})
		if !ok {
			return nil, errors.New("failed to typecast campaignmemeber records to array of interface")
		}

		for i := range recordsInt {
			record, ok := recordsInt[i].(map[string]interface{})
			if !ok {
				return nil, errors.New("failed to typecast campaignmemeber record to map")
			}

			if record["Id"] != "" {
				memberIDs = append(memberIDs, U.GetPropertyValueAsString(record["Id"]))
			}
		}
	}

	return memberIDs, nil

}

func getSalesforceContactIDANDLeadIDFromCampaignMember(properties *model.SalesforceRecord) (string, string) {
	var contactID, leadID string

	if (*properties)["LeadId"] != "" {
		leadID = U.GetPropertyValueAsString((*properties)["LeadId"])
	}

	if (*properties)["ContactId"] != "" {
		contactID = U.GetPropertyValueAsString((*properties)["ContactId"])
	}
	return contactID, leadID
}

func getAllCampaignMemberContactAndLeadRecords(projectID int64, campaignMemberIDs []string, accessToken, instanceURL string) ([]model.SalesforceRecord, []string, int, int, error) {

	logCtx := log.WithFields(log.Fields{"project_id": projectID})
	salesforceDataClient, err := NewSalesforceDataClient(accessToken, instanceURL)
	if err != nil {
		logCtx.WithError(err).Error("Failed to build salesforce data client to getAllCampaignMemberContactAndLeadRecords.")
		return nil, nil, 0, 0, err
	}

	// storing as map to avoid duplicate ids
	campaignMemberLeadIDsMap := make(map[string]bool)
	campaignMemberContactIDsMap := make(map[string]bool)

	campaingMemberAPICalls := 0
	if len(campaignMemberIDs) > 0 {

		batchedCampaignMemberIDs := U.GetStringListAsBatch(campaignMemberIDs, 50)
		for i := range batchedCampaignMemberIDs {
			paginatedCampaignMembersByID, err := salesforceDataClient.GetObjectRecordsByIDs(projectID, model.SalesforceDocumentTypeNameCampaignMember, batchedCampaignMemberIDs[i])
			if err != nil {
				logCtx.WithError(err).Error("Failed to initialize salesforce data client to getAllCampaignMemberContactAndLeadRecords.")
				return nil, nil, 0, 0, err
			}

			done := false
			for !done {
				var campaignMembers []model.SalesforceRecord
				campaignMembers, done, err = paginatedCampaignMembersByID.getNextBatch()
				if err != nil {
					logCtx.WithError(err).Error("Failed to get next batch to getAllCampaignMemberContactAndLeadRecords.")
					break // break and sync successfully pulled records
				}

				for i := range campaignMembers {
					contactID, leadID := getSalesforceContactIDANDLeadIDFromCampaignMember(&campaignMembers[i])
					if contactID != "" {
						campaignMemberContactIDsMap[contactID] = true
					}
					if leadID != "" {
						campaignMemberLeadIDsMap[leadID] = true
					}
				}
			}

			campaingMemberAPICalls += paginatedCampaignMembersByID.APICall
		}
	}

	campaignMemberLeadIDs := U.GetKeysMapAsArray(campaignMemberLeadIDsMap)
	campaignMemberContactIDs := U.GetKeysMapAsArray(campaignMemberContactIDsMap)
	// sync all campaign member if not existed since the first date of data pull
	memberObjectAPICalls := 0
	var memberRecords []model.SalesforceRecord
	var memberRecordsObjectType []string
	for campaignMemberObject, campaignMemberObjectIDs := range map[string][]string{model.SalesforceDocumentTypeNameLead: campaignMemberLeadIDs, model.SalesforceDocumentTypeNameContact: campaignMemberContactIDs} {
		batchedCampaignMemberObjectIDs := U.GetStringListAsBatch(campaignMemberObjectIDs, 50)
		for i := range batchedCampaignMemberObjectIDs {
			paginatedObjectsByID, err := salesforceDataClient.GetObjectRecordsByIDs(projectID, campaignMemberObject, batchedCampaignMemberObjectIDs[i])
			if err != nil {
				logCtx.WithFields(log.Fields{"object_name": campaignMemberObject}).WithError(err).Error("Failed to re-initialze salesforce data cleint for lead and contact ids.")
				return nil, nil, 0, 0, err
			}

			done := false
			var records []model.SalesforceRecord
			for !done {
				records, done, err = paginatedObjectsByID.getNextBatch()
				if err != nil {
					logCtx.WithFields(log.Fields{"object_name": campaignMemberObject}).WithError(err).Error("Failed to get next batch for lead and contact ids.")
					return nil, nil, 0, 0, err
				}

				for i := range records {
					memberRecords = append(memberRecords, records[i])
					memberRecordsObjectType = append(memberRecordsObjectType, campaignMemberObject)
				}

			}

			memberObjectAPICalls += paginatedObjectsByID.APICall
		}

	}

	return memberRecords, memberRecordsObjectType, campaingMemberAPICalls, memberObjectAPICalls, nil
}

func syncOpportunityPrimaryContact(projectID int64, primaryContactIDs []string, accessToken, instanceURL string) ([]string, int, bool) {
	logCtx := log.WithFields(log.Fields{"project_id": projectID})
	salesforceDataClient, err := NewSalesforceDataClient(accessToken, instanceURL)
	if err != nil {
		logCtx.WithError(err).Error("Failed to build new salesforce data client fron primary contact sync")
		return nil, 0, true
	}

	paginatedContacts, err := salesforceDataClient.GetObjectRecordsByIDs(projectID, model.SalesforceDocumentTypeNameContact, primaryContactIDs)
	if err != nil {
		logCtx.WithError(err).Error("Failed to initialize salesforce data client for sync oppportunities contact.")
		return nil, 0, true
	}

	var failures []string
	done := false
	opportunityPrimaryContact := 0
	var contactRecords []model.SalesforceRecord
	for !done {
		contactRecords, done, err = paginatedContacts.getNextBatch()
		if err != nil {
			return nil, 0, true
		}

		err = store.GetStore().BuildAndUpsertDocumentInBatch(projectID, model.SalesforceDocumentTypeNameContact, contactRecords)
		if err != nil {
			log.WithFields(log.Fields{"project_id": projectID}).Error("Failed to BuildAndUpsertDocument opportunity contact sync.")
			failures = append(failures, err.Error())
		}

	}

	opportunityPrimaryContact = paginatedContacts.APICall
	return failures, opportunityPrimaryContact, len(failures) > 0
}

// getLeadIDForOpportunityRecords sync associated leads if missing and return all lead ids
func getLeadIDForOpportunityRecords(projectID int64, records []model.SalesforceRecord, accessToken, instanceURL string) (map[string]string, map[string]map[string]bool, int, error) {
	if len(records) < 1 {
		return nil, nil, 0, nil
	}

	oppToLeadID := make(map[string]string, 0)
	oppToMultipleLeadID := make(map[string]map[string]bool, 0)
	oppIDs := make([]string, 0)
	for i := range records {
		oppID := util.GetPropertyValueAsString(records[i]["Id"])
		if oppID != "" {
			oppIDs = append(oppIDs, oppID)
			oppToLeadID[oppID] = ""
			oppToMultipleLeadID[oppID] = make(map[string]bool)
		}
	}

	leadIDForOpportunityRecordsAPICalls := 0
	batchedOppIDs := util.GetStringListAsBatch(oppIDs, 50)
	for bi := range batchedOppIDs {
		salesforceDataClient, err := NewSalesforceDataClient(accessToken, instanceURL)
		if err != nil {
			return nil, nil, 0, err
		}

		filterStmnt := "ConvertedOpportunityId IN (" + "'" + strings.Join(batchedOppIDs[bi], "','") + "')"
		paginatedLeads, err := salesforceDataClient.getRecordByObjectNameANDFilter(projectID, model.SalesforceDocumentTypeNameLead, filterStmnt)
		if err != nil {
			return nil, nil, 0, err
		}

		var objectRecords []model.SalesforceRecord
		done := false
		for !done {
			objectRecords, done, err = paginatedLeads.getNextBatch()
			if err != nil {
				return nil, nil, 0, err
			}

			for i := range objectRecords {
				leadID := util.GetPropertyValueAsString(objectRecords[i]["Id"])
				if leadID != "" {
					convertOppID := util.GetPropertyValueAsString(objectRecords[i]["ConvertedOpportunityId"])
					if convertOppID != "" {
						if leadID, exist := oppToLeadID[convertOppID]; exist && leadID != "" {
							log.WithFields(log.Fields{"lead_id": leadID}).Warn("Duplicate opportunity id on multiple leads")
						}

						oppToLeadID[convertOppID] = leadID
						oppToMultipleLeadID[convertOppID][leadID] = true
					} else {
						log.WithFields(log.Fields{"project_id": projectID}).Warn("Missing ConvertedOpportunityId on lead document")
					}

				} else {
					log.WithFields(log.Fields{"project_id": projectID}).Error("Missing lead id on lead document")
				}
			}

			err = store.GetStore().BuildAndUpsertDocumentInBatch(projectID, model.SalesforceDocumentTypeNameLead, objectRecords)
			if err != nil {
				log.WithFields(log.Fields{"project_id": projectID}).Error("Failed to BuildAndUpsertDocument opportunity lead sync .")
			}
		}
		leadIDForOpportunityRecordsAPICalls += paginatedLeads.APICall
	}

	return oppToLeadID, oppToMultipleLeadID, leadIDForOpportunityRecordsAPICalls, nil

}

func getOpportunityPrimaryContactIDs(projectID int64, oppRecords []model.SalesforceRecord) []string {
	primaryContacts := make([]string, 0)
	for i := range oppRecords {
		opportunityContactRolesInt := oppRecords[i][model.SalesforceChildRelationshipNameOpportunityContactRoles]
		if opportunityContactRolesInt == nil {
			log.WithFields(log.Fields{"project_id": projectID, "doc_id": oppRecords[i]["Id"]}).Warn("Missing opportunity contact roles")
			continue
		}

		opportunityContactRolesMap := opportunityContactRolesInt.(map[string]interface{})
		opportunityContactRoleRecords, ok := opportunityContactRolesMap["records"].([]interface{})
		if !ok {
			log.WithFields(log.Fields{"project_id": projectID, "doc_id": oppRecords[i]["Id"]}).Warn("Failed to typecast opportunity contact role records")
			continue
		}

		primaryContact := false
		for i := range opportunityContactRoleRecords {
			contactRole := opportunityContactRoleRecords[i].(map[string]interface{})
			if contactRole["IsPrimary"] == true {
				contactID := util.GetPropertyValueAsString(contactRole["ContactId"])
				if contactID != "" {
					primaryContacts = append(primaryContacts, contactID)
					primaryContact = true
					break
				} else {
					log.WithFields(log.Fields{"project_id": projectID, "doc_id": oppRecords[i]["Id"]}).Error("Missing primary contact id on opportunity contact roles.")
				}
			}
		}

		if len(opportunityContactRoleRecords) > 0 && !primaryContact {
			log.WithFields(log.Fields{"project_id": projectID, "doc_id": oppRecords[i]}).Warn("Missing primary contact. Skipping contact association.")
		}
	}

	return primaryContacts

}

func syncOpporunitiesUsingAssociations(projectID int64, accessToken, instanceURL string, timestamp int64) ([]string, int, int, int, error) {
	logCtx := log.WithFields(log.Fields{"project_id": projectID})
	allowedObject := model.GetSalesforceDocumentTypeAlias(projectID)
	salesforceDataClient, err := NewSalesforceDataClient(accessToken, instanceURL)
	if err != nil {
		logCtx.WithError(err).Error("Failed to build salesforceDataClient for opportunity sync.")
		return nil, 0, 0, 0, err
	}

	paginatedOpportunitiesByStartTimestamp, err := salesforceDataClient.getRecordByObjectNameANDStartTimestamp(projectID, model.SalesforceDocumentTypeNameOpportunity, timestamp)
	if err != nil {
		logCtx.WithError(err).Error("Failed to initialize salesforce data client for opportunity sync.")
		return nil, 0, 0, 0, err
	}

	done := false
	var objectRecords []model.SalesforceRecord
	var failures []string

	opportunityAPICalls := 0
	leadIDForOpportunityRecordsAPICalls := 0
	opportunityPrimaryContactAPICalls := 0
	for !done {
		objectRecords, done, err = paginatedOpportunitiesByStartTimestamp.getNextBatch()
		if err != nil {
			logCtx.WithError(err).Error("Failed to getNextBatch on opportunity sync.")
			return failures, 0, 0, 0, err
		}

		var oppToLeadIDs map[string]string
		var oppToMultipleLeadID map[string]map[string]bool
		if _, exist := allowedObject[model.SalesforceDocumentTypeNameLead]; exist {
			oppToLeadIDs, oppToMultipleLeadID, leadIDForOpportunityRecordsAPICalls, err = getLeadIDForOpportunityRecords(projectID, objectRecords, accessToken, instanceURL)
			if err != nil {
				logCtx.WithError(err).Error("Failed to get lead converted opportunity id for opportunity sync.")
			}
		}

		for i := range objectRecords {
			if _, exist := allowedObject[model.SalesforceDocumentTypeNameLead]; exist {
				oppID := util.GetPropertyValueAsString(objectRecords[i]["Id"])
				leadID := (oppToLeadIDs)[oppID]
				if leadID == "" {
					logCtx.WithFields(log.Fields{"opportunity_id": oppID}).Warn("Missing lead id for opportunity. Skipping adding lead id to opportunity.")
				} else {
					objectRecords[i][OpportunityLeadID] = leadID
				}

				if len(oppToMultipleLeadID[oppID]) > 0 {
					objectRecords[i][OpportunityMultipleLeadID] = oppToMultipleLeadID[oppID]
				}
			}

		}
		err = store.GetStore().BuildAndUpsertDocumentInBatch(projectID, model.SalesforceDocumentTypeNameOpportunity, objectRecords)
		if err != nil {
			logCtx.WithError(err).Error("Failed to BuildAndUpsertDocument for opportunity sync .")
			failures = append(failures, err.Error())
		}

		// only sync object if allowed by the project, will fallback to leads if not allowed
		if _, exist := allowedObject[model.SalesforceDocumentTypeNameContact]; exist {
			primaryContactIDs := getOpportunityPrimaryContactIDs(projectID, objectRecords)
			if len(primaryContactIDs) < 1 {
				continue
			}

			allFailures, apiCalls, failure := syncOpportunityPrimaryContact(projectID, primaryContactIDs, accessToken, instanceURL)
			if failure {
				failures = append(failures, allFailures...)
			}
			opportunityPrimaryContactAPICalls = apiCalls
		}

	}
	opportunityAPICalls = paginatedOpportunitiesByStartTimestamp.APICall

	return failures, opportunityAPICalls, leadIDForOpportunityRecordsAPICalls, opportunityPrimaryContactAPICalls, nil
}

// getLeadIDOrContactIDForActivityRecords sync associated leads/contacts if missing and return all lead/contact ids
func getLeadIDAndContactIDForActivityRecords(projectID int64, records []model.SalesforceRecord) ([]string, []string) {
	logCtx := log.WithFields(log.Fields{"project_id": projectID})

	leadIDs := make([]string, 0)
	contactIDs := make([]string, 0)

	if len(records) == 0 {
		return leadIDs, contactIDs
	}

	for i := range records {
		if records[i]["Who"] == nil {
			continue
		}

		whos := records[i]["Who"].(map[string]interface{})
		if U.GetPropertyValueAsString(whos["Type"]) == U.CapitalizeFirstLetter(model.SalesforceDocumentTypeNameLead) {
			leadIDs = append(leadIDs, U.GetPropertyValueAsString(whos["Id"]))
		} else if U.GetPropertyValueAsString(whos["Type"]) == U.CapitalizeFirstLetter(model.SalesforceDocumentTypeNameContact) {
			contactIDs = append(contactIDs, U.GetPropertyValueAsString(whos["Id"]))
		} else {
			logCtx.WithFields(log.Fields{"record_id": records[i]["Id"], "who_id": U.GetPropertyValueAsString(whos["Id"]), "who_type": U.GetPropertyValueAsString(whos["Type"])}).Error("Invalid objectName in getLeadIDOrContactIDForActivityRecords.")
		}
	}

	return leadIDs, contactIDs
}

func syncMissingObjectsForSalesforceActivities(projectID int64, documentIDs []string, objectName string, accessToken, instanceURL string) ([]string, int, bool) {
	logCtx := log.WithFields(log.Fields{"project_id": projectID})
	salesforceDataClient, err := NewSalesforceDataClient(accessToken, instanceURL)
	if err != nil {
		logCtx.WithError(err).Error("Failed to build new salesforce data client on syncMissingObjectsForSalesforceActivities")
		return nil, 0, true
	}

	if objectName != model.SalesforceDocumentTypeNameLead && objectName != model.SalesforceDocumentTypeNameContact {
		logCtx.Error("Invalid docType for salesforce activities in syncMissingObjectsForSalesforceActivities.")
		return nil, 0, true
	}

	var failures []string
	done := false
	activityIDs := 0

	if len(documentIDs) == 0 {
		logCtx.Info("No documentIDs to process in syncMissingObjectsForSalesforceActivities.")
		return nil, 0, false
	}

	docs, errCode := store.GetStore().GetSyncedSalesforceDocumentByType(projectID, documentIDs, model.GetSalesforceDocTypeByAlias(objectName), true)
	if errCode != http.StatusFound && errCode != http.StatusNotFound {
		logCtx.Error(fmt.Sprintf("Failed to get salesforce %s documents in syncMissingObjectsForSalesforceActivities.", objectName))
		return nil, 0, true
	}

	docIDs := make([]string, 0)
	for i := range docs {
		docIDs = append(docIDs, docs[i].ID)
	}

	missingDocIDs := U.StringSliceDiff(documentIDs, docIDs)
	if len(missingDocIDs) < 1 {
		return nil, 0, false
	}

	paginatedObjects, err := salesforceDataClient.GetObjectRecordsByIDs(projectID, objectName, missingDocIDs)
	if err != nil {
		logCtx.WithError(err).Error(fmt.Sprintf("Failed to initialize salesforce data client for sync activities %s.", objectName))
		return nil, 0, true
	}

	var records []model.SalesforceRecord
	for !done {
		records, done, err = paginatedObjects.getNextBatch()
		if err != nil {
			return nil, 0, true
		}

		err = store.GetStore().BuildAndUpsertDocumentInBatch(projectID, objectName, records)
		if err != nil {
			log.WithFields(log.Fields{"project_id": projectID}).Error(fmt.Sprintf("Failed to BuildAndUpsertDocument activities %s sync.", objectName))
			failures = append(failures, err.Error())
		}
	}

	activityIDs = paginatedObjects.APICall

	return failures, activityIDs, len(failures) > 0
}

func syncTasks(projectID int64, accessToken, instanceURL string, timestamp int64) ([]string, []string, []string, int, error) {
	logCtx := log.WithFields(log.Fields{"project_id": projectID})
	salesforceDataClient, err := NewSalesforceDataClient(accessToken, instanceURL)
	if err != nil {
		logCtx.WithError(err).Error("Failed to build salesforceDataClient for task sync.")
		return nil, nil, nil, 0, err
	}

	paginatedTasksByStartTimestamp, err := salesforceDataClient.getRecordByObjectNameANDStartTimestamp(projectID, model.SalesforceDocumentTypeNameTask, timestamp)
	if err != nil {
		logCtx.WithError(err).Error("Failed to initialize salesforce data client for task sync.")
		return nil, nil, nil, 0, err
	}

	done := false
	var objectRecords []model.SalesforceRecord
	var failures []string

	taskAPICalls := 0

	taskLeads := make([]string, 0)
	taskContacts := make([]string, 0)

	for !done {
		objectRecords, done, err = paginatedTasksByStartTimestamp.getNextBatch()
		if err != nil {
			logCtx.WithError(err).Error("Failed to getNextBatch on task sync.")
			return failures, nil, nil, 0, err
		}

		if len(objectRecords) == 0 {
			continue
		}

		leadIDs, contactIDs := getLeadIDAndContactIDForActivityRecords(projectID, objectRecords)

		if len(leadIDs) > 0 {
			taskLeads = append(taskLeads, leadIDs...)
		} else if len(contactIDs) > 0 {
			taskContacts = append(taskContacts, contactIDs...)
		} else {
			logCtx.Info("No leads or contacts associated with tasks.")
			continue
		}

		err = store.GetStore().BuildAndUpsertDocumentInBatch(projectID, model.SalesforceDocumentTypeNameTask, objectRecords)
		if err != nil {
			logCtx.WithError(err).Error("Failed to BuildAndUpsertDocument for task sync.")
			failures = append(failures, err.Error())
		}
	}
	taskAPICalls = paginatedTasksByStartTimestamp.APICall

	return failures, taskLeads, taskContacts, taskAPICalls, nil
}

func syncEvents(projectID int64, accessToken, instanceURL string, timestamp int64) ([]string, []string, []string, int, error) {
	logCtx := log.WithFields(log.Fields{"project_id": projectID})
	salesforceDataClient, err := NewSalesforceDataClient(accessToken, instanceURL)
	if err != nil {
		logCtx.WithError(err).Error("Failed to build salesforceDataClient for event sync.")
		return nil, nil, nil, 0, err
	}

	paginatedEventsByStartTimestamp, err := salesforceDataClient.getRecordByObjectNameANDStartTimestamp(projectID, model.SalesforceDocumentTypeNameEvent, timestamp)
	if err != nil {
		logCtx.WithError(err).Error("Failed to initialize salesforce data client for event sync.")
		return nil, nil, nil, 0, err
	}

	done := false
	var objectRecords []model.SalesforceRecord
	var failures []string

	eventAPICalls := 0

	eventLeads := make([]string, 0)
	eventContacts := make([]string, 0)

	for !done {
		objectRecords, done, err = paginatedEventsByStartTimestamp.getNextBatch()
		if err != nil {
			logCtx.WithError(err).Error("Failed to getNextBatch on event sync.")
			return failures, nil, nil, 0, err
		}

		if len(objectRecords) == 0 {
			continue
		}

		leadIDs, contactIDs := getLeadIDAndContactIDForActivityRecords(projectID, objectRecords)

		if len(leadIDs) > 0 {
			eventLeads = append(eventLeads, leadIDs...)
		} else if len(contactIDs) > 0 {
			eventContacts = append(eventContacts, contactIDs...)
		} else {
			logCtx.Info("No leads or contacts associated with events.")
			continue
		}

		err = store.GetStore().BuildAndUpsertDocumentInBatch(projectID, model.SalesforceDocumentTypeNameEvent, objectRecords)
		if err != nil {
			logCtx.WithError(err).Error("Failed to BuildAndUpsertDocument for event sync.")
			failures = append(failures, err.Error())
		}
	}
	eventAPICalls = paginatedEventsByStartTimestamp.APICall

	return failures, eventLeads, eventContacts, eventAPICalls, nil
}

func syncActivities(ps *model.SalesforceProjectSettings, accessToken, objectName string, timestamp int64) ([]string, int, int, int, error) {
	logCtx := log.WithFields(log.Fields{"project_id": ps.ProjectID, "doc_type": objectName})
	allowedObject := model.GetSalesforceDocumentTypeAlias(ps.ProjectID)

	leadAllowed := false
	if _, exist := allowedObject[model.SalesforceDocumentTypeNameLead]; exist {
		leadAllowed = true

	}

	contactAllowed := false
	if _, exist := allowedObject[model.SalesforceDocumentTypeNameContact]; exist {
		contactAllowed = true

	}

	if objectName == model.SalesforceDocumentTypeNameTask {
		failures, taskLeadIDs, taskContactIDs, taskAPICalls, err := syncTasks(ps.ProjectID, accessToken, ps.InstanceURL, timestamp)
		if err != nil {
			logCtx.WithError(err).Error("Failure on sync tasks.")
			return failures, 0, 0, 0, err
		}

		leadIDTaskAPICalls := 0
		if leadAllowed {
			leadFailures, leadAPICalls, failure := syncMissingObjectsForSalesforceActivities(ps.ProjectID, taskLeadIDs, model.SalesforceDocumentTypeNameLead, accessToken, ps.InstanceURL)
			if failure {
				failures = append(failures, leadFailures...)
			}
			leadIDTaskAPICalls = leadAPICalls
		}

		contactIDTaskAPICalls := 0
		if contactAllowed {
			contactFailures, contactAPICalls, failure := syncMissingObjectsForSalesforceActivities(ps.ProjectID, taskContactIDs, model.SalesforceDocumentTypeNameContact, accessToken, ps.InstanceURL)
			if failure {
				failures = append(failures, contactFailures...)
			}
			contactIDTaskAPICalls = contactAPICalls
		}

		return failures, taskAPICalls, leadIDTaskAPICalls, contactIDTaskAPICalls, nil
	}

	if objectName == model.SalesforceDocumentTypeNameEvent {
		failures, eventLeadIDs, eventContactIDs, eventAPICalls, err := syncEvents(ps.ProjectID, accessToken, ps.InstanceURL, timestamp)
		if err != nil {
			logCtx.WithError(err).Error("Failure on sync events.")
			return failures, 0, 0, 0, err
		}

		leadIDEventAPICalls := 0
		if leadAllowed {
			leadFailures, leadAPICalls, failure := syncMissingObjectsForSalesforceActivities(ps.ProjectID, eventLeadIDs, model.SalesforceDocumentTypeNameLead, accessToken, ps.InstanceURL)
			if failure {
				failures = append(failures, leadFailures...)
			}
			leadIDEventAPICalls = leadAPICalls
		}

		contactIDEventAPICalls := 0
		if contactAllowed {
			contactFailures, contactAPICalls, failure := syncMissingObjectsForSalesforceActivities(ps.ProjectID, eventContactIDs, model.SalesforceDocumentTypeNameContact, accessToken, ps.InstanceURL)
			if failure {
				failures = append(failures, contactFailures...)
			}
			contactIDEventAPICalls = contactAPICalls
		}

		return failures, eventAPICalls, leadIDEventAPICalls, contactIDEventAPICalls, nil
	}

	return nil, 0, 0, 0, errors.New("Invalid docType in syncActivities.")
}

func syncByType(ps *model.SalesforceProjectSettings, accessToken, objectName string, timestamp int64) (ObjectStatus, error) {
	var salesforceObjectStatus ObjectStatus
	salesforceObjectStatus.ProjetID = ps.ProjectID
	salesforceObjectStatus.DocType = objectName
	salesforceObjectStatus.TotalAPICalls = make(map[string]int)

	logCtx := log.WithFields(log.Fields{"project_id": ps.ProjectID, "doc_type": objectName})

	if objectName == model.SalesforceDocumentTypeNameOpportunity && config.UseOpportunityAssociationByProjectID(ps.ProjectID) {
		failures, opportunityAPICalls, leadIDForOpportunityRecordsAPICall, opportunityPrimaryContact, err := syncOpporunitiesUsingAssociations(ps.ProjectID, accessToken, ps.InstanceURL, timestamp)
		if err != nil {
			logCtx.WithError(err).Error("Failure on sync opportunities.")
			salesforceObjectStatus.Failures = append(salesforceObjectStatus.Failures, failures...)
			return salesforceObjectStatus, err
		}

		salesforceObjectStatus.TotalAPICalls["opportunityAPICalls"] = opportunityAPICalls
		salesforceObjectStatus.TotalAPICalls["leadIDForOpportunityRecordsAPICalls"] = leadIDForOpportunityRecordsAPICall
		salesforceObjectStatus.TotalAPICalls["opportunityPrimaryContactAPICalls"] = opportunityPrimaryContact
		return salesforceObjectStatus, nil
	}

	if (objectName == model.SalesforceDocumentTypeNameTask && config.IsAllowedSalesforceActivityTasksByProjectID(ps.ProjectID)) || (objectName == model.SalesforceDocumentTypeNameEvent && config.IsAllowedSalesforceActivityEventsByProjectID(ps.ProjectID)) {
		failures, activitiesAPICalls, leadIDForActivitiesRecordsAPICall, contactIDForActivitiesRecordsAPICall, err := syncActivities(ps, accessToken, objectName, timestamp)
		if err != nil {
			logCtx.WithError(err).Error("Failure on sync activities.")
			salesforceObjectStatus.Failures = append(salesforceObjectStatus.Failures, failures...)
			return salesforceObjectStatus, err
		}

		salesforceObjectStatus.TotalAPICalls[fmt.Sprintf("%sAPICalls", objectName)] = activitiesAPICalls
		salesforceObjectStatus.TotalAPICalls[fmt.Sprintf("leadIDFor%sRecordsAPICalls", U.CapitalizeFirstLetter(objectName))] = leadIDForActivitiesRecordsAPICall
		salesforceObjectStatus.TotalAPICalls[fmt.Sprintf("contactIDFor%sRecordsAPICalls", U.CapitalizeFirstLetter(objectName))] = contactIDForActivitiesRecordsAPICall
		return salesforceObjectStatus, nil
	}

	salesforceDataClient, err := NewSalesforceDataClient(accessToken, ps.InstanceURL)
	if err != nil {
		logCtx.WithError(err).Error("Failed to build salesforceDataClient.")
		return salesforceObjectStatus, err
	}

	allCampaignMemberIDs := make([]string, 0)
	allCampaignIDs := make(map[string]bool)

	paginatedObjectsByStartTimestamp, err := salesforceDataClient.getRecordByObjectNameANDStartTimestamp(ps.ProjectID, objectName, timestamp)
	if err != nil {
		logCtx.WithError(err).Error("Failed to initialize salesforce data client.")
		return salesforceObjectStatus, err
	}

	done := false
	var objectRecords []model.SalesforceRecord
	for !done {
		objectRecords, done, err = paginatedObjectsByStartTimestamp.getNextBatch()
		if err != nil {
			logCtx.WithError(err).Error("Failed to getNextBatch.")
			return salesforceObjectStatus, err
		}

		var failures []string
		for i := range objectRecords {
			// get campaing memeber ids from the campaign to sync missing leads,contacts and campaign members associated with the campaign
			if objectName == model.SalesforceDocumentTypeNameCampaign {
				campaignMemberIDs, err := getCampaingMemberIDsFromCampaign(&objectRecords[i])
				if err != nil {
					logCtx.WithError(err).Error("Failed to get campaign member ids from campaign.")
				} else {
					allCampaignMemberIDs = append(allCampaignMemberIDs, campaignMemberIDs...)
				}

			}

			if objectName == model.SalesforceDocumentTypeNameCampaignMember {
				campaignID := util.GetPropertyValueAsString(objectRecords[i]["CampaignId"])

				if campaignID != "" {
					allCampaignIDs[campaignID] = true
				} else {
					logCtx.WithError(err).Error("Missing campaign Id from campaign member record.")
				}
				campaignMemberIDs := util.GetPropertyValueAsString(objectRecords[i]["Id"])
				if campaignMemberIDs != "" {
					allCampaignMemberIDs = append(allCampaignMemberIDs, campaignMemberIDs)
				} else {
					logCtx.WithError(err).Error("Missing campaign member Id from campaign member record.")
				}
			}
		}

		err = store.GetStore().BuildAndUpsertDocumentInBatch(ps.ProjectID, objectName, objectRecords)
		if err != nil {
			logCtx.WithError(err).Error("Failed to BuildAndUpsertDocument.")
			failures = append(failures, err.Error())
		}

		salesforceObjectStatus.Failures = append(salesforceObjectStatus.Failures, failures...)
	}

	salesforceObjectStatus.TotalAPICalls[objectName] = paginatedObjectsByStartTimestamp.APICall

	// sync missing lead or contact id if not available from first date of data pull
	if objectName == model.SalesforceDocumentTypeNameCampaign || objectName == model.SalesforceDocumentTypeNameCampaignMember {

		campaignMemberRecords, recordObjectType, campaingMemberAPICalls, memberObjectAPICalls, err := getAllCampaignMemberContactAndLeadRecords(ps.ProjectID, allCampaignMemberIDs, accessToken, ps.InstanceURL)
		if err != nil {
			logCtx.WithError(err).Error("Failed to getAllCampaignMemberContactAndLeadRecords")
			return salesforceObjectStatus, err
		}
		salesforceObjectStatus.TotalAPICalls["CampaignMemberAPICalls"] = campaingMemberAPICalls
		salesforceObjectStatus.TotalAPICalls["MemberObjectAPICalls"] = memberObjectAPICalls

		objectRecordsMap := make(map[string][]model.SalesforceRecord, 0)
		for i := range campaignMemberRecords {
			if _, exist := objectRecordsMap[recordObjectType[i]]; !exist {
				objectRecordsMap[recordObjectType[i]] = make([]model.SalesforceRecord, 0)
			}
			objectRecordsMap[recordObjectType[i]] = append(objectRecordsMap[recordObjectType[i]], campaignMemberRecords[i])
		}

		var failures []string
		for objectName := range objectRecordsMap {
			err = store.GetStore().BuildAndUpsertDocumentInBatch(ps.ProjectID, objectName, objectRecordsMap[objectName])
			if err != nil {
				logCtx.WithError(err).Error("Failed to insert campaign members on BuildAndUpsertDocument.")
				failures = append(failures, err.Error())
			}
		}
	}

	// sync missing campaign and campaignmember if not available from first date of data pull
	if objectName == model.SalesforceDocumentTypeNameCampaignMember || objectName == model.SalesforceDocumentTypeNameCampaign {
		docIDs := make([]string, 0)
		var docObjectName string
		// sync missing campaign from campaignmember
		if objectName == model.SalesforceDocumentTypeNameCampaignMember {
			for campaignID := range allCampaignIDs {
				docIDs = append(docIDs, campaignID)
			}
			docObjectName = model.SalesforceDocumentTypeNameCampaign
		}
		// sync missing campaignmember from campaign
		if objectName == model.SalesforceDocumentTypeNameCampaign {
			for _, memberID := range allCampaignMemberIDs {
				docIDs = append(docIDs, memberID)
			}
			docObjectName = model.SalesforceDocumentTypeNameCampaignMember
		}

		batchedDocIDs := U.GetStringListAsBatch(docIDs, 50)
		for i := range batchedDocIDs {
			paginatedObjectByID, err := salesforceDataClient.GetObjectRecordsByIDs(ps.ProjectID, docObjectName, batchedDocIDs[i])
			if err != nil {
				logCtx.WithError(err).Error("Failed to re-initialize salesforce data client.")
				return salesforceObjectStatus, err
			}

			var campaignRecords []model.SalesforceRecord
			done = false
			for !done {
				campaignRecords, done, err = paginatedObjectByID.getNextBatch()
				if err != nil {
					logCtx.WithError(err).Error("Failed to getNextBatch.")
					return salesforceObjectStatus, err
				}

				err = store.GetStore().BuildAndUpsertDocumentInBatch(ps.ProjectID, docObjectName, campaignRecords)
				if err != nil {
					logCtx.WithError(err).Error("Failed to insert unsynced campaing related document on BuildAndUpsertDocument.")
				}
			}
			salesforceObjectStatus.TotalAPICalls[docObjectName] += paginatedObjectByID.APICall
		}
	}

	return salesforceObjectStatus, nil
}

// TokenError implements error interface for token api error
type TokenError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// GetAccessToken gets new salesforce access token by refresh token
func GetAccessToken(ps *model.SalesforceProjectSettings, redirectURL string) (string, string, error) {
	logCtx := log.WithFields(log.Fields{"project_id": ps.ProjectID})

	if ps == nil || redirectURL == "" {
		return "", "", errors.New("invalid project setting or redirect url")
	}

	queryParams := fmt.Sprintf("grant_type=%s&refresh_token=%s&client_id=%s&client_secret=%s&redirect_uri=%s",
		"refresh_token", ps.RefreshToken, C.GetSalesforceAppId(), C.GetSalesforceAppSecret(), redirectURL)
	url := RefreshTokenURL + "?" + queryParams

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return "", "", err
	}

	client := &http.Client{
		Timeout: 10 * time.Minute,
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var errBody TokenError
		json.NewDecoder(resp.Body).Decode(&errBody)
		return "", "", fmt.Errorf("error while query data %s : %s", errBody.Error, errBody.ErrorDescription)
	}

	var jsonResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&jsonResponse)
	if err != nil {
		return "", "", err
	}

	accessToken, exists := jsonResponse["access_token"].(string)
	if !exists && accessToken == "" {
		return "", "", errors.New("failed to get access token by refresh token")
	}

	instanceURL, exists := jsonResponse["instance_url"].(string)
	if !exists && instanceURL == "" {
		logCtx.Error("Failed to get instance_url in GetAccessToken method.")
		return "", "", errors.New("failed to get instance_url")
	}

	if ps.InstanceURL != instanceURL {
		projectSetting, errCode := store.GetStore().GetProjectSetting(ps.ProjectID)
		if errCode != http.StatusFound {
			logCtx.Error("Failed to fetch Project Setting in GetAccessToken method for project.")
		} else {
			errCode := store.GetStore().UpdateAgentSalesforceInstanceURL(*projectSetting.IntSalesforceEnabledAgentUUID,
				instanceURL)
			if errCode != http.StatusAccepted {
				logCtx.WithFields(log.Fields{"Agent_uuid": projectSetting.IntSalesforceEnabledAgentUUID,
					"instanceURL": instanceURL})
				logCtx.Error("Failed to update instanceURL for agent in GetAccessToken method.")
			}
		}
	}

	return accessToken, instanceURL, nil
}

// CreateOrGetSalesforceEventName makes sure salesforce event name exists
func CreateOrGetSalesforceEventName(projectID int64) int {
	logCtx := log.WithFields(log.Fields{"project_id": projectID})

	for _, doctype := range model.GetSalesforceAllowedObjects(projectID) {
		if skipObjectEvent(doctype) {
			continue
		}

		typAlias := model.GetSalesforceAliasByDocType(doctype)
		eventName := model.GetSalesforceEventNameByAction(typAlias, model.SalesforceDocumentCreated)
		_, status := store.GetStore().CreateOrGetEventName(&model.EventName{
			ProjectId: projectID,
			Name:      eventName,
			Type:      model.TYPE_USER_CREATED_EVENT_NAME,
		})

		if status != http.StatusFound && status != http.StatusConflict && status != http.StatusCreated {
			logCtx.Error("Failed to create event name on SyncDatetimeAndNumericalProperties.")
			return http.StatusInternalServerError
		}

		eventName = model.GetSalesforceEventNameByAction(typAlias, model.SalesforceDocumentUpdated)
		_, status = store.GetStore().CreateOrGetEventName(&model.EventName{
			ProjectId: projectID,
			Name:      eventName,
			Type:      model.TYPE_USER_CREATED_EVENT_NAME,
		})

		if status != http.StatusFound && status != http.StatusConflict && status != http.StatusCreated {
			logCtx.Error("Failed to create updated event name on SyncDatetimeAndNumericalProperties.")
			return http.StatusInternalServerError
		}
	}

	if !C.IsAllowedSalesforceGroupsByProjectID(projectID) {
		return http.StatusOK
	}

	/*
		Create group and its events
	*/
	_, status := store.GetStore().CreateGroup(projectID, model.GROUP_NAME_SALESFORCE_ACCOUNT, model.AllowedGroupNames)
	if status != http.StatusCreated && status != http.StatusConflict {
		return http.StatusInternalServerError
	}

	_, status = store.GetStore().CreateGroup(projectID, model.GROUP_NAME_SALESFORCE_OPPORTUNITY, model.AllowedGroupNames)
	if status != http.StatusCreated && status != http.StatusConflict {
		return http.StatusInternalServerError
	}

	for _, eventName := range []string{U.GROUP_EVENT_NAME_SALESFORCE_ACCOUNT_CREATED,
		U.GROUP_EVENT_NAME_SALESFORCE_ACCOUNT_UPDATED, U.GROUP_EVENT_NAME_SALESFORCE_OPPORTUNITY_CREATED, U.GROUP_EVENT_NAME_SALESFORCE_OPPORTUNITY_UPDATED} {
		_, status = store.GetStore().CreateOrGetEventName(&model.EventName{
			ProjectId: projectID,
			Name:      eventName,
			Type:      model.TYPE_USER_CREATED_EVENT_NAME,
		})

		if status != http.StatusFound && status != http.StatusConflict && status != http.StatusCreated {
			logCtx.WithFields(log.Fields{"event_name": eventName}).Error("Failed to create salesforce group event name.")
			return http.StatusInternalServerError
		}
	}

	return http.StatusOK
}

func syncSalesforcePropertyByType(projectID int64, doctTypeAlias string, fieldName, fieldType string) error {
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "field_name": fieldName, "field_type": fieldType, "doc_type_alias": doctTypeAlias})

	if fieldName == "" || fieldType == "" || projectID == 0 || doctTypeAlias == "" {
		logCtx.Error("Missing required field.")
		return errors.New("missing required field")
	}

	pType := model.GetSalesforceMappedDataType(U.GetPropertyValueAsString(fieldType))

	enKey := model.GetCRMEnrichPropertyKeyByType(
		model.SmartCRMEventSourceSalesforce,
		doctTypeAlias,
		fieldName,
	)

	eventName := model.GetSalesforceEventNameByAction(doctTypeAlias, model.SalesforceDocumentCreated)
	err := store.GetStore().CreateOrDeletePropertyDetails(projectID, eventName, enKey, pType, false, true)
	if err != nil {
		logCtx.WithFields(log.Fields{"enriched_property_key": enKey}).WithError(err).
			Error("Failed to create created event property details.")
		return err
	}

	eventName = model.GetSalesforceEventNameByAction(doctTypeAlias, model.SalesforceDocumentUpdated)
	err = store.GetStore().CreateOrDeletePropertyDetails(projectID, eventName, enKey, pType, false, true)
	if err != nil {
		logCtx.WithFields(log.Fields{"enriched_property_key": enKey}).WithError(err).
			Error("Failed to create updated event property details.")
		return err
	}

	err = store.GetStore().CreateOrDeletePropertyDetails(projectID, "", enKey, pType, true, true)
	if err != nil {
		logCtx.WithFields(log.Fields{"enriched_property_key": enKey}).WithError(err).
			Error("Failed to create user property details.")
		return err
	}

	return nil
}

func skipObjectEvent(docType int) bool {
	return docType == model.SalesforceDocumentTypeOpportunityContactRole
}

// SyncDatetimeAndNumericalProperties sync datetime and numerical properties to the property_details table
func SyncDatetimeAndNumericalProperties(projectID int64, accessToken, instanceURL string) (bool, []Status) {
	if projectID == 0 || accessToken == "" || instanceURL == "" {
		return false, nil
	}

	logCtx := log.WithFields(log.Fields{"project_id": projectID})
	status := CreateOrGetSalesforceEventName(projectID)
	if status != http.StatusOK {
		logCtx.Errorf("Failed to CreateOrGetSalesforceEventName status %d", status)
		return true, nil
	}

	var allStatus []Status
	anyFailures := false
	for _, doctype := range model.GetSalesforceAllowedObjects(projectID) {
		if skipObjectEvent(doctype) {
			continue
		}

		var status Status
		typAlias := model.GetSalesforceAliasByDocType(doctype)
		status.Type = typAlias
		status.ProjectID = projectID

		docTypeFailure := false
		describe, err := getSalesforceObjectDescription(projectID, typAlias, accessToken, instanceURL)
		if err != nil {
			logCtx.WithError(err).Error("Failed to sync datetime and numerical properties.")
			anyFailures = true
			continue
		}

		for i := range describe.Fields {
			fieldType, exist := describe.Fields[i]["type"]
			if !exist {
				logCtx.WithFields(log.Fields{"property_type": fieldType}).Error("Failed to get property type field.")
				docTypeFailure = true
				continue
			}

			fieldName, exist := describe.Fields[i]["name"]
			if !exist {
				logCtx.Error("Failed to get property name field.")
				docTypeFailure = true
				continue
			}

			if failure := syncSalesforcePropertyByType(projectID, typAlias, U.GetPropertyValueAsString(fieldName), U.GetPropertyValueAsString(fieldType)); failure != nil {
				docTypeFailure = true
			}

			label, exist := describe.Fields[i]["label"]
			if !exist {
				logCtx.Warn("Failed to get property label.")
			} else {
				logCtx.Info("Inserting display names")
				err := store.GetStore().CreateOrUpdateDisplayNameByObjectType(projectID, model.GetCRMEnrichPropertyKeyByType(
					model.SmartCRMEventSourceSalesforce,
					typAlias,
					U.GetPropertyValueAsString(fieldName),
				), typAlias, U.GetPropertyValueAsString(label), model.SmartCRMEventSourceSalesforce)
				if err != http.StatusCreated && err != http.StatusConflict {
					logCtx.Error("Failed to create or update display name")
				}
			}
		}

		if docTypeFailure {
			status.Status = U.CRM_SYNC_STATUS_FAILURES
			anyFailures = true
		} else {
			status.Status = U.CRM_SYNC_STATUS_FAILURES
		}

		allStatus = append(allStatus, status)
	}

	return anyFailures, allStatus
}

func getStartTimestamp(docType string) int64 {
	if docType != model.SalesforceDocumentTypeNameCampaignMember && docType != model.SalesforceDocumentTypeNameTask && docType != model.SalesforceDocumentTypeNameEvent {
		return 0 // 1 January 1970 00:00:00
	}

	currentTime := time.Now().AddDate(0, 0, -90).UTC()
	return now.New(currentTime).BeginningOfDay().Unix() // get from last 90 days
}

// SyncDocuments syncs from salesforce to database by doc type
func SyncDocuments(ps *model.SalesforceProjectSettings, lastSyncInfo map[string]int64, accessToken string) []ObjectStatus {
	var allObjectStatus []ObjectStatus

	for docType, timestamp := range lastSyncInfo {
		var syncAll bool
		if timestamp == 0 {
			timestamp = getStartTimestamp(docType)
			syncAll = true
		}

		objectStatus, err := syncByType(ps, accessToken, docType, timestamp)
		if err != nil || len(objectStatus.Failures) != 0 {
			log.WithFields(log.Fields{
				"project_id": ps.ProjectID,
				"doctype":    docType,
			}).WithError(err).Errorf("Failed to sync documents")

			if err != nil {
				objectStatus.Message = err.Error()
			}

			objectStatus.Status = U.CRM_SYNC_STATUS_FAILURES
		} else {
			objectStatus.Status = U.CRM_SYNC_STATUS_SUCCESS
		}

		objectStatus.SyncAll = syncAll
		allObjectStatus = append(allObjectStatus, objectStatus)
	}

	return allObjectStatus
}
