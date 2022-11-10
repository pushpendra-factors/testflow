package model

import (
	"fmt"
)

const (
	LEADSQUARED_LEAD           = "lead"
	LEADSQUARED_SALES_ACTIVITY = "sales_activity"
	LEADSQUARED_EMAIL_SENT     = "email_sent"
	LEADSQUARED_EMAIL_INFO     = "email_info"
	LEADSQUARED_HAD_A_CALL     = "had_a_call"
)

var LeadSquaredDocumentToQuery = map[string]string{
	LEADSQUARED_LEAD: "select * FROM `%s.%s.lead`" +
		" WHERE %v AND ProspectAutoId > %v order by ProspectAutoId asc LIMIT %v OFFSET 0",
	LEADSQUARED_SALES_ACTIVITY: "select * FROM `%s.%s.sales_activity`" +
		" WHERE %v AND CreatedOnUnix > %v order by CreatedOnUnix asc LIMIT %v OFFSET 0", // fix this
	LEADSQUARED_EMAIL_SENT: "select * FROM `%s.%s.sales_activity`" +
		" WHERE %v AND CreatedOnUnix > %v order by CreatedOnUnix asc LIMIT %v OFFSET 0", // fix this
	LEADSQUARED_EMAIL_INFO: "select * FROM `%s.%s.sales_activity`" +
		" WHERE %v AND CreatedOnUnix > %v order by CreatedOnUnix asc LIMIT %v OFFSET 0", // fix this
	LEADSQUARED_HAD_A_CALL: "select * FROM `%s.%s.sales_activity`" +
		" WHERE %v AND CreatedOnUnix > %v order by CreatedOnUnix asc LIMIT %v OFFSET 0", // fix this
}

var LeadSquaredDocumentEndpoint = map[string]string{
	LEADSQUARED_LEAD:           "/v2/LeadManagement.svc/Leads.RecentlyModified",
	LEADSQUARED_SALES_ACTIVITY: "/v2/ProspectActivity.svc/CustomActivity/RetrieveByActivityEvent",
	LEADSQUARED_EMAIL_SENT:     "/v2/ProspectActivity.svc/CustomActivity/RetrieveByActivityEvent",
	LEADSQUARED_EMAIL_INFO:     "/v2/ProspectActivity.svc/CustomActivity/RetrieveByActivityEvent",
	LEADSQUARED_HAD_A_CALL:     "/v2/ProspectActivity.svc/CustomActivity/RetrieveByActivityEvent",
}

var LeadSquaredMetadataEndpoint = map[string]string{
	LEADSQUARED_LEAD:           "/v2/LeadManagement.svc/LeadsMetaData.Get",
	LEADSQUARED_SALES_ACTIVITY: "/v2/ProspectActivity.svc/CustomActivity/GetActivitySetting",
	LEADSQUARED_EMAIL_SENT:     "/v2/ProspectActivity.svc/CustomActivity/GetActivitySetting",
	LEADSQUARED_EMAIL_INFO:     "/v2/ProspectActivity.svc/CustomActivity/GetActivitySetting",
	LEADSQUARED_HAD_A_CALL:     "/v2/ProspectActivity.svc/CustomActivity/GetActivitySetting",
}

var LeadSquaredHistoricalSyncEndpoint = map[string]string{
	LEADSQUARED_LEAD: "/v2/LeadManagement.svc/Leads.Get",
}

var LeadSquaredTableName = map[string]string{
	LEADSQUARED_LEAD:           "lead",
	LEADSQUARED_SALES_ACTIVITY: "sales_activity",
	LEADSQUARED_EMAIL_SENT:     "email_sent",
	LEADSQUARED_EMAIL_INFO:     "email_info",
	LEADSQUARED_HAD_A_CALL:     "had_a_call",
}

var LeadSquaredDataObjectColumnsQuery = map[string]string{
	LEADSQUARED_LEAD:           "SELECT * FROM `%s.%s.INFORMATION_SCHEMA.COLUMNS` WHERE table_name = 'lead' ORDER by ordinal_position",
	LEADSQUARED_SALES_ACTIVITY: "SELECT * FROM `%s.%s.INFORMATION_SCHEMA.COLUMNS` WHERE table_name = 'sales_activity' ORDER by ordinal_position",
	LEADSQUARED_EMAIL_SENT:     "SELECT * FROM `%s.%s.INFORMATION_SCHEMA.COLUMNS` WHERE table_name = 'email_sent' ORDER by ordinal_position",
	LEADSQUARED_EMAIL_INFO:     "SELECT * FROM `%s.%s.INFORMATION_SCHEMA.COLUMNS` WHERE table_name = 'email_info' ORDER by ordinal_position",
	LEADSQUARED_HAD_A_CALL:     "SELECT * FROM `%s.%s.INFORMATION_SCHEMA.COLUMNS` WHERE table_name = 'had_a_call' ORDER by ordinal_position",
}

var LeadSquaredDocTypeIntegrationObjectMap = map[string]string{
	LEADSQUARED_LEAD:           "user",
	LEADSQUARED_SALES_ACTIVITY: "activity",
	LEADSQUARED_EMAIL_SENT:     "activity",
	LEADSQUARED_EMAIL_INFO:     "activity",
	LEADSQUARED_HAD_A_CALL:     "activity",
}

var LeadSquaredUserIdMapping = map[string]string{
	LEADSQUARED_LEAD: "ProspectID",
}

var LeadSquaredUserAutoIdMapping = map[string]string{
	LEADSQUARED_LEAD:           "ProspectAutoId",
	LEADSQUARED_SALES_ACTIVITY: "CreatedOnUnix",
	LEADSQUARED_EMAIL_SENT:     "CreatedOnUnix",
	LEADSQUARED_EMAIL_INFO:     "CreatedOnUnix",
	LEADSQUARED_HAD_A_CALL:     "CreatedOnUnix",
}

var LeadSquaredEmailMapping = map[string]string{
	LEADSQUARED_LEAD: "EmailAddress",
}

var LeadSquaredPhoneMapping = map[string]string{
	LEADSQUARED_LEAD: "Phone",
}

var LeadSquaredDocumentTypeAlias = map[string]int{
	LEADSQUARED_LEAD:           1,
	LEADSQUARED_SALES_ACTIVITY: 2,
	LEADSQUARED_EMAIL_SENT:     3,
	LEADSQUARED_EMAIL_INFO:     4,
	LEADSQUARED_HAD_A_CALL:     5,
}

var LeadSquaredDataObjectFilters = map[string]string{
	LEADSQUARED_LEAD:           "DATE(%v) = '%v'",
	LEADSQUARED_SALES_ACTIVITY: "DATE(%v) = '%v'",
	LEADSQUARED_EMAIL_SENT:     "DATE(%v) = '%v'",
	LEADSQUARED_EMAIL_INFO:     "DATE(%v) = '%v'",
	LEADSQUARED_HAD_A_CALL:     "DATE(%v) = '%v'",
}

var LeadSquaredDataObjectFiltersColumn = map[string]string{
	LEADSQUARED_LEAD:           "synced_at",
	LEADSQUARED_SALES_ACTIVITY: "synced_at",
	LEADSQUARED_EMAIL_SENT:     "synced_at",
	LEADSQUARED_EMAIL_INFO:     "synced_at",
	LEADSQUARED_HAD_A_CALL:     "synced_at",
}

var LeadSquaredTimestampMapping = map[string][]string{
	LEADSQUARED_LEAD:           []string{"CreatedOn", "CreatedOn", "ModifiedOn"},
	LEADSQUARED_SALES_ACTIVITY: []string{"CreatedOn", "ModifiedOn"},
	LEADSQUARED_EMAIL_SENT:     []string{"CreatedOn", "ModifiedOn"},
	LEADSQUARED_EMAIL_INFO:     []string{"CreatedOn", "ModifiedOn"},
	LEADSQUARED_HAD_A_CALL:     []string{"CreatedOn", "ModifiedOn"},
}

var LeadSquaredProgramIdMapping = map[string]string{
	LEADSQUARED_SALES_ACTIVITY: "ProspectActivityId",
	LEADSQUARED_EMAIL_SENT:     "ProspectActivityId",
	LEADSQUARED_EMAIL_INFO:     "ProspectActivityId",
	LEADSQUARED_HAD_A_CALL:     "ProspectActivityId",
}

var LeadSquaredActorTypeMapping = map[string]string{
	LEADSQUARED_SALES_ACTIVITY: LEADSQUARED_LEAD,
	LEADSQUARED_EMAIL_SENT:     LEADSQUARED_LEAD,
	LEADSQUARED_EMAIL_INFO:     LEADSQUARED_LEAD,
	LEADSQUARED_HAD_A_CALL:     LEADSQUARED_LEAD,
}

var LeadSquaredActorIdMapping = map[string]string{
	LEADSQUARED_SALES_ACTIVITY: "RelatedProspectId",
	LEADSQUARED_EMAIL_SENT:     "RelatedProspectId",
	LEADSQUARED_EMAIL_INFO:     "RelatedProspectId",
	LEADSQUARED_HAD_A_CALL:     "RelatedProspectId",
}

func GetLeadSquaredTypeToAliasMap(aliasType map[string]int) map[int]string {
	typeAlias := make(map[int]string)
	for alias, typ := range aliasType {
		typeAlias[typ] = alias
	}
	return typeAlias
}

func GetLeadSquaredDocumentQuery(bigQueryProjectId string, schemaId string, baseQuery string, executionDate string, docType string, limit int, lastProcessedId int) string {
	if docType == LEADSQUARED_LEAD {
		return fmt.Sprintf(baseQuery, bigQueryProjectId, schemaId, GetLeadSquaredDocumentFilterCondition(docType, false, "", executionDate), lastProcessedId, limit)
	}
	if docType == LEADSQUARED_SALES_ACTIVITY {
		return fmt.Sprintf(baseQuery, bigQueryProjectId, schemaId, GetLeadSquaredDocumentFilterCondition(docType, false, "", executionDate), lastProcessedId, limit)
	}
	if docType == LEADSQUARED_EMAIL_SENT {
		return fmt.Sprintf(baseQuery, bigQueryProjectId, schemaId, GetLeadSquaredDocumentFilterCondition(docType, false, "", executionDate), lastProcessedId, limit)
	}
	if docType == LEADSQUARED_EMAIL_INFO {
		return fmt.Sprintf(baseQuery, bigQueryProjectId, schemaId, GetLeadSquaredDocumentFilterCondition(docType, false, "", executionDate), lastProcessedId, limit)
	}
	if docType == LEADSQUARED_HAD_A_CALL {
		return fmt.Sprintf(baseQuery, bigQueryProjectId, schemaId, GetLeadSquaredDocumentFilterCondition(docType, false, "", executionDate), lastProcessedId, limit)
	}
	return ""
}

func GetLeadSquaredDocumentActorId(documentType string, data []string, metadataColumns []string) string {
	actorId, exists := LeadSquaredActorIdMapping[documentType]
	if exists {
		dataObjectColumns := GetObjectDataColumns(documentType, metadataColumns)
		index, exists_index := dataObjectColumns[actorId]
		if exists_index {
			if data[index] == "<nil>" {
				return ""
			} else {
				return data[index]
			}
		}
		return ""
	}
	return ""
}

func GetLeadSquaredDocumentFilterCondition(docType string, addPrefix bool, prefix string, executionDate string) string {
	filterColumn := ""
	if addPrefix {
		filterColumn = fmt.Sprintf("%v.%v", prefix, LeadSquaredDataObjectFiltersColumn[docType])
	} else {
		filterColumn = LeadSquaredDataObjectFiltersColumn[docType]
	}
	filterCondition := fmt.Sprintf(LeadSquaredDataObjectFilters[docType], filterColumn, executionDate)

	return filterCondition
}

func GetLeadSquaredDocumentMetadataQuery(docType string, bigQueryProjectId string, schemaId string, baseQuery string) (string, bool) {
	if docType == LEADSQUARED_LEAD {
		return fmt.Sprintf(baseQuery, bigQueryProjectId, schemaId), true
	}
	if docType == LEADSQUARED_SALES_ACTIVITY {
		return fmt.Sprintf(baseQuery, bigQueryProjectId, schemaId), true
	}
	if docType == LEADSQUARED_EMAIL_SENT {
		return fmt.Sprintf(baseQuery, bigQueryProjectId, schemaId), true
	}
	if docType == LEADSQUARED_EMAIL_INFO {
		return fmt.Sprintf(baseQuery, bigQueryProjectId, schemaId), true
	}
	if docType == LEADSQUARED_HAD_A_CALL {
		return fmt.Sprintf(baseQuery, bigQueryProjectId, schemaId), true
	}
	return "", false
}

func GetLeadSquaredDocumentDocumentType(documentTypeString string) int {
	docTypeId, exists := LeadSquaredDocumentTypeAlias[documentTypeString]
	if exists {
		return docTypeId
	}
	return 0
}

func GetLeadSquaredObjectDataColumns(docType string, metadataColumns []string) map[string]int {
	dataObjectColumns := make(map[string]int, 0)
	if metadataColumns != nil {
		for index, column := range metadataColumns {
			dataObjectColumns[column] = index
		}
	}
	return dataObjectColumns
}

func GetLeadSquaredActorType(documentTypeString string) int {
	actorType, exists := LeadSquaredActorTypeMapping[documentTypeString]
	if exists {
		actorTypeId, exists_actor := LeadSquaredDocumentTypeAlias[actorType]
		if exists_actor {
			return actorTypeId
		}
		return 0
	}
	return 0
}

func GetLeadSquaredDocumentValues(docType string, data []string, metadataColumns []string, metadataColumnDateTimeType map[string]bool, metadataColumnNumericalType map[string]bool) map[string]interface{} {
	values := make(map[string]interface{})
	dataObjectColumns := GetLeadSquaredObjectDataColumns(docType, metadataColumns)
	for key, index := range dataObjectColumns {
		if metadataColumnDateTimeType != nil && metadataColumnDateTimeType[key] {
			convertedTimestamp := ConvertTimestamp(data[index])
			if convertedTimestamp == 0 {
				values[key] = nil
			} else {
				values[key] = convertedTimestamp
			}
		} else if metadataColumnNumericalType != nil && metadataColumnNumericalType[key] {
			convertedNumber := ConvertToNumber(data[index])
			if convertedNumber == 0 {
				values[key] = nil
			} else {
				values[key] = convertedNumber
			}
		} else {
			if data[index] == "<nil>" {
				values[key] = ""
			} else {
				values[key] = data[index]
			}
		}
	}
	return values
}

func GetLeadSquaredUserId(documentType string, data []string, metadataColumns []string) string {
	activtyNameId, exists := LeadSquaredUserIdMapping[documentType]
	if exists {
		dataObjectColumns := GetLeadSquaredObjectDataColumns(documentType, metadataColumns)
		index, exists_index := dataObjectColumns[activtyNameId]
		if exists_index {
			if data[index] == "<nil>" {
				return ""
			} else {
				return data[index]
			}
		}
		return ""
	}
	return ""
}

func GetLeadSquaredUserAutoId(documentType string, data []string, metadataColumns []string) string {
	activtyNameId, exists := LeadSquaredUserAutoIdMapping[documentType]
	if exists {
		dataObjectColumns := GetLeadSquaredObjectDataColumns(documentType, metadataColumns)
		index, exists_index := dataObjectColumns[activtyNameId]
		if exists_index {
			if data[index] == "<nil>" {
				return ""
			} else {
				return data[index]
			}
		}
		return ""
	}
	return ""
}

func GetLeadSquaredDocumentPhone(documentType string, data []string, metadataColumns []string) string {
	activtyNameId, exists := LeadSquaredPhoneMapping[documentType]
	if exists {
		dataObjectColumns := GetLeadSquaredObjectDataColumns(documentType, metadataColumns)
		index, exists_index := dataObjectColumns[activtyNameId]
		if exists_index {
			if data[index] == "<nil>" {
				return ""
			} else {
				return data[index]
			}
		}
		return ""
	}
	return ""
}

func GetLeadSquaredDocumentEmail(documentType string, data []string, metadataColumns []string) string {
	activtyNameId, exists := LeadSquaredEmailMapping[documentType]
	if exists {
		dataObjectColumns := GetLeadSquaredObjectDataColumns(documentType, metadataColumns)
		index, exists_index := dataObjectColumns[activtyNameId]
		if exists_index {
			if data[index] == "<nil>" {
				return ""
			} else {
				return data[index]
			}
		}
		return ""
	}
	return ""
}

func GetLeadSquaredDocumentAction(documentType string, data []string, metadataColumns []string) CRMAction {
	dataObjectColumns := GetLeadSquaredObjectDataColumns(documentType, metadataColumns)
	created_at_index, exists_created_at_index := dataObjectColumns["CreatedOn"]
	updated_at_index, exists_updated_at_index := dataObjectColumns["ModifiedOn"]
	if !exists_created_at_index || !exists_updated_at_index {
		return 0
	}
	if ConvertTimestamp(data[updated_at_index]) > ConvertTimestamp(data[created_at_index]) {
		return CRMActionUpdated
	} else {
		return CRMActionCreated
	}
}

func GetLeadSquaredDocumentTimestamp(documentType string, data []string, metadataColumns []string) []int64 {
	timestampIds, exists := LeadSquaredTimestampMapping[documentType]
	result := make([]int64, 0)
	for _, timestampId := range timestampIds {
		if exists {
			dataObjectColumns := GetLeadSquaredObjectDataColumns(documentType, metadataColumns)
			index, exists_index := dataObjectColumns[timestampId]
			if exists_index {
				result = append(result, ConvertTimestamp(data[index]))
			}
		}
	}
	return result
}

func GetLeadSquaredDocumentProgramId(documentType string, data []string, metadataColumns []string) string {
	activtyNameId, exists := LeadSquaredProgramIdMapping[documentType]
	if exists {
		dataObjectColumns := GetObjectDataColumns(documentType, metadataColumns)
		index, exists_index := dataObjectColumns[activtyNameId]
		if exists_index {
			if data[index] == "<nil>" {
				return ""
			} else {
				return data[index]
			}
		}
		return ""
	}
	return ""
}

type LeadsByDateRangeRequest struct {
	Parameter ParameterObj `json:"Parameter"`
	Columns   ColumnsObj   `json:"Columns"`
	Paging    PagingObj    `json:"Paging"`
}

type ParameterObj struct {
	FromDate string `json:"FromDate"`
	ToDate   string `json:"ToDate"`
}

type LeadSearchParameterObj struct {
	LookupName  string `json:"LookupName"`
	LookupValue string `json:"LookupValue"`
	SqlOperator string `json:"SqlOperator"`
}

type SalesActivitySearchParameterObj struct {
	FromDate         string `json:"FromDate"`
	ToDate           string `json:"ToDate"`
	ActivityEvent    int    `json:"ActivityEvent"`
	RemoveEmptyValue bool   `json:"RemoveEmptyValue"`
}

type ColumnsObj struct {
	IncludeCSV string `json:"Include_CSV"`
}

type PagingObj struct {
	PageIndex int `json:"PageIndex"`
	PageSize  int `json:"PageSize"`
}

type SortingObj struct {
	ColumnName string `json:"ColumnName"`
	Direction  string `json:"Direction"`
}

type LeadsByDateRangeResponse struct {
	RecordCount int            `json:"RecordCount"`
	Leads       []LeadResponse `json:"Leads"`
}

type LeadResponse struct {
	LeadPropertyList []LeadAttributeValue `json:"LeadPropertyList"`
}

type LeadAttributeValue struct {
	Attribute string `json:"Attribute"`
	Value     string `json:"Value"`
}

type SearchLeadsByCriteriaRequest struct {
	Columns   ColumnsObj             `json:"Columns"`
	Paging    PagingObj              `json:"Paging"`
	Sorting   SortingObj             `json:"Sorting"`
	Parameter LeadSearchParameterObj `json:"Parameter"`
}

type SearchSalesActivityByCriteriaRequest struct {
	Paging    PagingObj                       `json:"Paging"`
	Sorting   SortingObj                      `json:"Sorting"`
	Parameter SalesActivitySearchParameterObj `json:"Parameter"`
}
