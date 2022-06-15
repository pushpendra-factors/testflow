package model

import (
	"fmt"
)

const (
	LEADSQUARED_LEAD = "lead"
)

var LeadSquaredDocumentToQuery = map[string]string{
	LEADSQUARED_LEAD: "select * FROM `%s.%s.lead`" +
		" WHERE %v AND ProspectAutoId > %v order by ProspectAutoId asc LIMIT %v OFFSET 0",
}
var LeadSquaredDocumentEndpoint = map[string]string{
	LEADSQUARED_LEAD: "/v2/LeadManagement.svc/Leads.RecentlyModified",
}

var LeadSquaredMetadataEndpoint = map[string]string{
	LEADSQUARED_LEAD: "/v2/LeadManagement.svc/LeadsMetaData.Get",
}

var LeadSquaredHistoricalSyncEndpoint = map[string]string{
	LEADSQUARED_LEAD: "/v2/LeadManagement.svc/Leads.Get",
}

var LeadSquaredTableName = map[string]string{
	LEADSQUARED_LEAD: "lead",
}

var LeadSquaredDataObjectColumnsQuery = map[string]string{
	LEADSQUARED_LEAD: "SELECT * FROM `%s.%s.INFORMATION_SCHEMA.COLUMNS` WHERE table_name = 'lead' ORDER by ordinal_position",
}

var LeadSquaredDocTypeIntegrationObjectMap = map[string]string{
	LEADSQUARED_LEAD: "user",
}

var LeadSquaredUserIdMapping = map[string]string{
	LEADSQUARED_LEAD: "ProspectID",
}

var LeadSquaredUserAutoIdMapping = map[string]string{
	LEADSQUARED_LEAD: "ProspectAutoId",
}

var LeadSquaredEmailMapping = map[string]string{
	LEADSQUARED_LEAD: "EmailAddress",
}

var LeadSquaredPhoneMapping = map[string]string{
	LEADSQUARED_LEAD: "Phone",
}

var LeadSquaredDocumentTypeAlias = map[string]int{
	LEADSQUARED_LEAD: 1,
}

var LeadSquaredDataObjectFilters = map[string]string{
	LEADSQUARED_LEAD: "DATE(%v) = '%v'",
}

var LeadSquaredDataObjectFiltersColumn = map[string]string{
	LEADSQUARED_LEAD: "synced_at",
}

var LeadSquaredTimestampMapping = map[string][]string{
	LEADSQUARED_LEAD: []string{"CreatedOn", "CreatedOn", "ModifiedOn"},
}

func GetLeadSquaredDocumentQuery(bigQueryProjectId string, schemaId string, baseQuery string, executionDate string, docType string, limit int, lastProcessedId int) string {
	if docType == LEADSQUARED_LEAD {
		return fmt.Sprintf(baseQuery, bigQueryProjectId, schemaId, GetLeadSquaredDocumentFilterCondition(docType, false, "", executionDate), lastProcessedId, limit)
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
