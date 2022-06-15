package model

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

var MarketoDocumentToQuery = map[string]string{
	MARKETO_TYPE_NAME_PROGRAM_MEMBERSHIP: "SELECT pm.id as lead_id, pm.program_id as program_id, pm.acquired_by, pm.is_exhausted, pm.membership_date, pm.nurture_cadence, pm.progression_status," +
		"pm.reached_success, pm.reached_success_date, pm.stream, p.channel AS program_channel, p.created_at AS program_created_at, p.description AS program_description, p.end_date AS program_end_date, " +
		"p.name AS program_name, p.sfdc_id AS program_sfdc_id, p.sfdc_name AS program_sfdc_name, p.start_date AS program_start_date, p.status AS program_status, " +
		"p.type AS program_type, p.url AS program_url, p.workspace AS program_workspace FROM " +
		"`%s.%s.program_membership` AS pm " +
		" left outer join " +
		" (SELECT * FROM (SELECT *, ROW_NUMBER() OVER (PARTITION BY  id ORDER BY updated_at DESC) AS row_num " +
		" FROM `%s.%s.program`)prog WHERE prog.row_num = 1)p " +
		" ON pm.program_id = p.id WHERE %v AND pm.membership_date IS NOT NULL ORDER BY pm.id, pm.program_id LIMIT %v OFFSET %v",
	MARKETO_TYPE_NAME_LEAD: "select lead_seg_agg.segment_ids,lead_seg_agg.segment_names,lead_seg_agg.segmentation_ids,lead_seg_agg.segmentation_names,l.* FROM `%s.%s.lead` AS l LEFT OUTER JOIN " +
		" (SELECT ls.id, ARRAY_AGG(DISTINCT ls.segment_id IGNORE NULLS) AS segment_ids, ARRAY_AGG(DISTINCT s.name IGNORE NULLS) AS segment_names, " +
		" ARRAY_AGG(DISTINCT sg.id IGNORE NULLS) AS segmentation_ids, ARRAY_AGG(DISTINCT sg.name IGNORE NULLS) AS segmentation_names " +
		" FROM `%s.%s.lead_segment` AS ls left outer join `%s.%s.segment` AS s on ls.segment_id = s.id " +
		" left outer join `%s.%s.segmentation` AS sg on s.segmentation_id = sg.id group by ls.id) lead_seg_agg on l.id = lead_seg_agg.id " +
		" WHERE %v order by id asc LIMIT %v OFFSET %v",
	MARKETO_TYPE_NAME_LEAD_NO_SEGMENT: "select NULL AS segment_ids, NULL AS segment_names, NULL AS segmentation_ids, NULL AS segmentation_names, p.channel AS program_channel, p.created_at AS program_created_at, p.description AS program_description, p.end_date AS program_end_date, " +
		" p.name AS program_name, p.sfdc_id AS program_sfdc_id, p.sfdc_name AS program_sfdc_name, p.start_date AS program_start_date, p.status AS program_status, " +
		" p.type AS program_type, p.url AS program_url, p.workspace,l.* FROM `%s.%s.lead` AS l " +
		" left outer join " +
		" (SELECT * FROM (SELECT *, ROW_NUMBER() OVER (PARTITION BY  id ORDER BY updated_at DESC) AS row_num " +
		" FROM `%s.%s.program`)prog WHERE prog.row_num = 1)p " +
		" ON l.acquisition_program_id = p.id " +
		" WHERE %v AND l.id > %v order by l.id asc LIMIT %v",
}

func GetMarketoDocumentFilterCondition(docType string, addPrefix bool, prefix string, executionDate string) string {
	filterColumn := ""
	if addPrefix {
		filterColumn = fmt.Sprintf("%v.%v", prefix, MarketoDataObjectFiltersColumn[docType])
	} else {
		filterColumn = MarketoDataObjectFiltersColumn[docType]
	}
	filterCondition := fmt.Sprintf(MarketoDataObjectFilters[docType], filterColumn, executionDate)

	return filterCondition
}

var MarketoDataObjectFilters = map[string]string{
	MARKETO_TYPE_NAME_PROGRAM_MEMBERSHIP: "DATE(%v) = '%v'",
	MARKETO_TYPE_NAME_LEAD:               "DATE(%v) = '%v'",
	MARKETO_TYPE_NAME_LEAD_NO_SEGMENT:    "DATE(%v) = '%v'",
}

var MarketoDataObjectFiltersColumn = map[string]string{
	MARKETO_TYPE_NAME_PROGRAM_MEMBERSHIP: "_fivetran_synced",
	MARKETO_TYPE_NAME_LEAD:               "_fivetran_synced",
	MARKETO_TYPE_NAME_LEAD_NO_SEGMENT:    "_fivetran_synced",
}

func GetMarketoDocumentQuery(bigQueryProjectId string, schemaId string, baseQuery string, executionDate string, docType string, limit int, offset int, lastProcessedRecord int) string {

	if docType == MARKETO_TYPE_NAME_PROGRAM_MEMBERSHIP {
		return fmt.Sprintf(baseQuery, bigQueryProjectId, schemaId, bigQueryProjectId, schemaId, GetMarketoDocumentFilterCondition(docType, true, "pm", executionDate), limit, offset)
	}
	if docType == MARKETO_TYPE_NAME_LEAD {
		return fmt.Sprintf(baseQuery, bigQueryProjectId, schemaId, bigQueryProjectId, schemaId, bigQueryProjectId, schemaId, bigQueryProjectId, schemaId, GetMarketoDocumentFilterCondition(docType, true, "l", executionDate), limit, offset)
	}
	if docType == MARKETO_TYPE_NAME_LEAD_NO_SEGMENT {
		return fmt.Sprintf(baseQuery, bigQueryProjectId, schemaId, bigQueryProjectId, schemaId, GetMarketoDocumentFilterCondition(docType, true, "l", executionDate), lastProcessedRecord, limit)
	}
	return ""
}

func GetMarketoDocumentMetadataQuery(docType string, bigQueryProjectId string, schemaId string, baseQuery string) (string, bool) {
	if docType == MARKETO_TYPE_NAME_LEAD {
		return fmt.Sprintf(baseQuery, bigQueryProjectId, schemaId), true
	}
	return "", false
}

var MarketoMetadataColumns = map[string]map[string]int{
	"metadata": {"table_catalog": 0, "table_schema": 1, "table_name": 2, "column_name": 3, "ordinal_position": 4, "is_nullable": 5,
		"data_type": 6, "is_generated": 7, "generation_expression": 8, "is_stored": 9, "is_hidden": 10, "is_updatable": 11,
		"is_system_defined": 12, "is_partitioning_column": 13, "clustering_ordinal_position": 14},
}

var MarketoDataObjectColumnsInValue = map[string]map[string]int{
	MARKETO_TYPE_NAME_PROGRAM_MEMBERSHIP: {"lead_id": 0, "program_id": 1, "acquired_by": 2, "is_exhausted": 3, "membership_date": 4, "nurture_cadence": 5, "progression_status": 6,
		"reached_success": 7, "reached_success_date": 8, "stream": 9, "program_channel": 10, "program_created_at": 11, "program_description": 12, "program_end_date": 13,
		"program_name": 14, "program_sfdc_id": 15, "program_sfdc_name": 16, "program_start_date": 17, "program_status": 18,
		"program_type": 19, "program_url": 20, "program_workspace": 21},
	MARKETO_TYPE_NAME_LEAD: {"segment_ids": 0, "segment_names": 1, "segmentation_ids": 2, "segmentation_names": 3, "program_channel": 4, "program_created_at": 5, "program_description": 6, "program_end_date": 7,
		"program_name": 8, "program_sfdc_id": 9, "program_sfdc_name": 10, "program_start_date": 11, "program_status": 12,
		"program_type": 13, "program_url": 14, "program_workspace": 15},
}

var MarketoDataObjectColumnsQuery = map[string]string{
	MARKETO_TYPE_NAME_LEAD: "SELECT * FROM `%s.%s.INFORMATION_SCHEMA.COLUMNS` WHERE table_name = 'lead' ORDER by ordinal_position",
}

var MarketoDataObjectColumnsDatetimeType = map[string]map[string]bool{
	MARKETO_TYPE_NAME_PROGRAM_MEMBERSHIP: {"membership_date": true, "reached_success_date": true, "program_created_at": true, "program_end_date": true, "program_start_date": true},
	MARKETO_TYPE_NAME_LEAD:               {"program_created_at": true, "program_end_date": true, "program_start_date": true},
}

var MarketoDataObjectColumnsNumericalType = map[string]map[string]bool{
	MARKETO_TYPE_NAME_PROGRAM_MEMBERSHIP: {"lead_id": true, "program_id": true},
}

var DocTypeIntegrationObjectMap = map[string]string{
	MARKETO_TYPE_NAME_PROGRAM_MEMBERSHIP: "activity",
	MARKETO_TYPE_NAME_LEAD:               "user",
}

func GetObjectDataColumns(docType string, metadataColumns []string) map[string]int {
	dataObjectColumns := make(map[string]int, 0)
	for key, index := range MarketoDataObjectColumnsInValue[docType] {
		dataObjectColumns[key] = index
	}
	if metadataColumns != nil {
		for index, column := range metadataColumns {
			dataObjectColumns[column] = index + len(MarketoDataObjectColumnsInValue[docType])
		}
	}
	return dataObjectColumns
}
func GetMarketoDocumentValues(docType string, data []string, metadataColumns []string, metadataColumnDateTimeType map[string]bool, metadataColumnNumericalType map[string]bool) map[string]interface{} {
	values := make(map[string]interface{})
	dataObjectColumns := GetObjectDataColumns(docType, metadataColumns)
	for key, index := range dataObjectColumns {
		if MarketoDataObjectColumnsDatetimeType[docType][key] || (metadataColumnDateTimeType != nil && metadataColumnDateTimeType[key]) {
			convertedTimestamp := ConvertTimestamp(data[index])
			if convertedTimestamp == 0 {
				values[key] = nil
			} else {
				values[key] = convertedTimestamp
			}
		} else if MarketoDataObjectColumnsNumericalType[docType][key] || (metadataColumnNumericalType != nil && metadataColumnNumericalType[key]) {
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

func GetMarketoDocumentDocumentType(documentTypeString string) int {
	docTypeId, exists := MarketoDocumentTypeAlias[documentTypeString]
	if exists {
		return docTypeId
	}
	return 0
}

const (
	MARKETO_TYPE_NAME_PROGRAM_MEMBERSHIP = "program_membership"
	MARKETO_TYPE_NAME_LEAD               = "lead"
	MARKETO_TYPE_NAME_LEAD_NO_SEGMENT    = "lead_no_segment"
)

var MarketoDocumentTypeAlias = map[string]int{
	// MARKETO_TYPE_NAME_PROGRAM_MEMBERSHIP: 1,
	MARKETO_TYPE_NAME_LEAD: 2,
}

var MarketoActorTypeMapping = map[string]string{
	MARKETO_TYPE_NAME_PROGRAM_MEMBERSHIP: MARKETO_TYPE_NAME_LEAD,
}

var MarketoActorIdMapping = map[string]string{
	MARKETO_TYPE_NAME_PROGRAM_MEMBERSHIP: "lead_id",
}

var MarketoEmailMapping = map[string]string{
	MARKETO_TYPE_NAME_LEAD: "email",
}

var MarketoUserIdMapping = map[string]string{
	MARKETO_TYPE_NAME_LEAD: "id",
}

var MarketoPhoneMapping = map[string]string{
	MARKETO_TYPE_NAME_LEAD: "phone",
}

var MarketoTimestampMapping = map[string][]string{
	MARKETO_TYPE_NAME_PROGRAM_MEMBERSHIP: []string{"membership_date"},
	MARKETO_TYPE_NAME_LEAD:               []string{"created_at", "created_at", "updated_at"},
}

var MarketoLogValuesMapping = map[string][]string{
	MARKETO_TYPE_NAME_PROGRAM_MEMBERSHIP: []string{"lead_id", "program_id"},
	MARKETO_TYPE_NAME_LEAD:               []string{"id"},
}

var MarketoProgramIdMapping = map[string]string{
	MARKETO_TYPE_NAME_PROGRAM_MEMBERSHIP: "program_id",
}

func GetMarketoDocumentProgramId(documentType string, data []string, metadataColumns []string) string {
	activtyNameId, exists := MarketoProgramIdMapping[documentType]
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

func GetMarketoTypeToAliasMap(aliasToType map[string]int) (map[int]string, error) {
	typeToAlias := make(map[int]string)
	for typeAlias := range aliasToType {
		objectType := aliasToType[typeAlias]
		if _, exist := typeToAlias[objectType]; exist {
			return nil, errors.New("same type on alias")
		}
		typeToAlias[objectType] = typeAlias
	}
	return typeToAlias, nil
}

func GetMetadataColumnNameIndex(columnName string) int {
	metadata, exists := MarketoMetadataColumns["metadata"]
	if exists {
		index, indexExists := metadata[columnName]
		if indexExists {
			return index
		}
	}
	return -1
}

func GetMarketoActorType(documentTypeString string) int {
	actorType, exists := MarketoActorTypeMapping[documentTypeString]
	if exists {
		actorTypeId, exists_actor := MarketoDocumentTypeAlias[actorType]
		if exists_actor {
			return actorTypeId
		}
		return 0
	}
	return 0
}

func GetMarketoDocumentActorId(documentType string, data []string, metadataColumns []string) string {
	actorId, exists := MarketoActorIdMapping[documentType]
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

func GetUniqueLogValue(docType string, data []string, metadataColumns []string) string {
	index_ids, exists := MarketoLogValuesMapping[docType]
	result := ""
	for _, indexId := range index_ids {
		if exists {
			dataObjectColumns := GetObjectDataColumns(docType, metadataColumns)
			index, exists_index := dataObjectColumns[indexId]
			if exists_index {
				if result == "" {
					result = data[index]
				} else {
					result = result + "," + data[index]
				}
			}
		}
	}
	return result
}
func GetMarketoDocumentPhone(documentType string, data []string, metadataColumns []string) string {
	activtyNameId, exists := MarketoPhoneMapping[documentType]
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

func GetMarketoDocumentEmail(documentType string, data []string, metadataColumns []string) string {
	activtyNameId, exists := MarketoEmailMapping[documentType]
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

func GetMarketoUserId(documentType string, data []string, metadataColumns []string) string {
	activtyNameId, exists := MarketoUserIdMapping[documentType]
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

func GetMarketoDocumentAction(documentType string, data []string, metadataColumns []string) CRMAction {
	dataObjectColumns := GetObjectDataColumns(documentType, metadataColumns)
	created_at_index, exists_created_at_index := dataObjectColumns["created_at"]
	updated_at_index, exists_updated_at_index := dataObjectColumns["updated_at"]
	if !exists_created_at_index || !exists_updated_at_index {
		return 0
	}
	if ConvertTimestamp(data[updated_at_index]) > ConvertTimestamp(data[created_at_index]) {
		return CRMActionUpdated
	} else {
		return CRMActionCreated
	}
}

func GetMarketoDocumentTimestamp(documentType string, data []string, metadataColumns []string) []int64 {
	timestampIds, exists := MarketoTimestampMapping[documentType]
	result := make([]int64, 0)
	for _, timestampId := range timestampIds {
		if exists {
			dataObjectColumns := GetObjectDataColumns(documentType, metadataColumns)
			index, exists_index := dataObjectColumns[timestampId]
			if exists_index {
				result = append(result, ConvertTimestamp(data[index]))
			}
		}
	}
	return result
}

func ConvertTimestamp(date string) int64 {
	var err1, err2 error
	dateConverted, err := time.Parse("2006-01-02 15:04:05 +0000 UTC", date)
	if err != nil {
		dateConverted, err1 = time.Parse("2006-01-02T15:04:05", date)
		if err1 != nil {
			dateConverted, err2 = time.Parse("2006-01-02", date)
			if err2 != nil {
				return 0
			}
		}
	}
	return dateConverted.Unix()
}

func ConvertToNumber(num string) float64 {
	if s, err := strconv.ParseFloat(num, 64); err == nil {
		return s
	}
	return 0
}
