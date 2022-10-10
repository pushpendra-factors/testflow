
import operator

from lib.utils.adwords.format import FormatUtil
from .fields_mapping import FieldsMapping
import logging as log

class Payload:

    headers = []
    rows = []

    DEFAULT_FLOAT = 0.000
    DEFAULT_NUMERATOR_FLOAT = 0.0
    DEFAULT_DENOMINATOR_FLOAT = 1.0
    DEFAULT_DECIMAL_PLACES = 3

    OPERAND1 = "operand1"
    OPERAND2 = "operand2"
    OPERATION = "operation"
    RESULT_FIELD = "result_field"
    TRANSFORM_AND_ADD_NEW_FIELDS = [
        {OPERAND1: "impressions", OPERAND2: "search_impression_share", RESULT_FIELD: "total_search_impression",
         OPERATION: operator.truediv},
        {OPERAND1: "impressions", OPERAND2: "search_click_share", RESULT_FIELD: "total_search_click",
         OPERATION: operator.truediv},
        {OPERAND1: "impressions", OPERAND2: "search_top_impression_share", RESULT_FIELD: "total_search_top_impression",
         OPERATION: operator.truediv},
        {OPERAND1: "impressions", OPERAND2: "search_absolute_top_impression_share",
         RESULT_FIELD: "total_search_absolute_top_impression",
         OPERATION: operator.truediv},
        {OPERAND1: "impressions", OPERAND2: "search_budget_lost_absolute_top_impression_share",
         RESULT_FIELD: "total_search_budget_lost_absolute_top_impression", OPERATION: operator.truediv},
        {OPERAND1: "impressions", OPERAND2: "search_budget_lost_impression_share",
         RESULT_FIELD: "total_search_budget_lost_impression", OPERATION: operator.truediv},
        {OPERAND1: "impressions", OPERAND2: "search_budget_lost_top_impression_share",
         RESULT_FIELD: "total_search_budget_lost_top_impression", OPERATION: operator.truediv},
        {OPERAND1: "impressions", OPERAND2: "search_rank_lost_absolute_top_impression_share",
         RESULT_FIELD: "total_search_rank_lost_absolute_top_impression", OPERATION: operator.truediv},
        {OPERAND1: "impressions", OPERAND2: "search_rank_lost_impression_share",
         RESULT_FIELD: "total_search_rank_lost_impression", OPERATION: operator.truediv},
        {OPERAND1: "impressions", OPERAND2: "search_rank_lost_top_impression_share",
         RESULT_FIELD: "total_search_rank_lost_top_impression", OPERATION: operator.truediv},
        {OPERAND1: "impressions", OPERAND2: "absolute_top_impression_percentage",
         RESULT_FIELD: "absolute_top_impressions", OPERATION: operator.mul},
        {OPERAND1: "impressions", OPERAND2: "top_impression_percentage",
         RESULT_FIELD: "top_impressions", OPERATION: operator.mul},
        {OPERAND1: "top_impressions", OPERAND2: "search_top_impression_share",
         RESULT_FIELD: "total_top_impressions", OPERATION: operator.truediv},
        {OPERAND1: "total_search_impression", OPERAND2: "search_budget_lost_impression_share",
         RESULT_FIELD: "impression_lost_due_to_budget", OPERATION: operator.mul},
        {OPERAND1: "total_search_impression", OPERAND2: "search_rank_lost_impression_share",
         RESULT_FIELD: "impression_lost_due_to_rank", OPERATION: operator.mul},
        {OPERAND1: "total_top_impressions", OPERAND2: "search_budget_lost_top_impression_share",
         RESULT_FIELD: "top_impression_lost_due_to_budget", OPERATION: operator.mul},
        {OPERAND1: "total_top_impressions", OPERAND2: "search_budget_lost_absolute_top_impression_share",
         RESULT_FIELD: "absolute_top_impression_lost_due_to_budget", OPERATION: operator.mul},
        {OPERAND1: "total_top_impressions", OPERAND2: "search_rank_lost_top_impression_share",
         RESULT_FIELD: "top_impression_lost_due_to_rank", OPERATION: operator.mul},
        {OPERAND1: "total_top_impressions", OPERAND2: "search_rank_lost_absolute_top_impression_share",
         RESULT_FIELD: "absolute_top_impression_lost_due_to_rank", OPERATION: operator.mul},
    ]

    fields_with_percentages = {}
    fields_in_0_to_1 = {}
    fields_to_float = {}

    FIELDS = "fields"
    transformations_with_fields_and_method = []
    
    FIELD = "field"
    MAP = "map"
    transform_map = []

    def __init__(self, headers, rows, 
                fields_with_percentages=[], fields_in_0_to_1=[], fields_to_float=[],
                fields_with_status=[], fields_with_boolean=[], fields_with_resource_name=[], 
                fields_to_percentage=[], fields_with_interaction_types=[], fields_with_approval_status=[], transform_map={}):
        self.headers = headers
        self.rows = rows
        self.fields_with_percentages = fields_with_percentages
        self.fields_in_0_to_1 = fields_in_0_to_1
        self.fields_to_float = fields_to_float
        self.transformations_with_fields_and_method = [
            {self.FIELDS: fields_with_status, self.OPERATION: FieldsMapping.transform_status},
            {self.FIELDS: fields_with_boolean, self.OPERATION: FieldsMapping.transform_boolean},
            {self.FIELDS: fields_with_resource_name, self.OPERATION: FieldsMapping.transform_resource_name},
            {self.FIELDS: fields_to_percentage, self.OPERATION: FieldsMapping.transform_percentage},
            {self.FIELDS: fields_with_interaction_types, self.OPERATION: FieldsMapping.transform_interaction_types},
            {self.FIELDS: fields_with_approval_status, self.OPERATION: FieldsMapping.transform_approval_status},
        ]
        self.transform_map = transform_map

    def transform_entities(self):
        transformed_rows = []
        for row in self.rows:
            transformed_rows.append(self.transform_entity(row))
        return transformed_rows

    def transform_entity(self, row):
        for transform in self.TRANSFORM_AND_ADD_NEW_FIELDS:
            field1_name = transform[self.OPERAND1]
            field2_name = transform[self.OPERAND2]
            operation = transform[self.OPERATION]
            result_field_name = transform[self.RESULT_FIELD]
            if field1_name in row and field2_name in row:
                field1_value = self.get_transformed_values(field1_name, row.get(field1_name, ""))
                field2_value = self.get_transformed_values(field2_name, row.get(field2_name, ""))
                transformed_value = self.get_transformed_value_for_arithmetic_operator(field1_value, field2_value,
                                                                                       operation)
                row[result_field_name] = transformed_value

        return self.transform_entity_with_methods_and_mapping(row)

    def transform_entities_click_view(self):
        already_present = {}
        for row in self.rows:
            if row["gcl_id"] in already_present:
                continue
            row = self.transform_entity_with_methods_and_mapping(row)
            already_present[row["gcl_id"]] = row
        return list(already_present.values())

    def transform_entity_with_methods_and_mapping(self, row): 
        for transform in self.transformations_with_fields_and_method:
            fields = transform[self.FIELDS]
            operation = transform[self.OPERATION]
            for field in fields:
                if(field in row and row[field] != ''):
                    row[field] = operation(row[field])

        for transform in self.transform_map:
            field = transform[self.FIELD]
            if(row[field] != ''):
                row[field] = transform[self.MAP][row[field]]
        return row

    def get_transformed_values(self, field_name, value):
        response_value = value
        if field_name in self.fields_with_percentages:
            response_value = FormatUtil.get_numeric_from_percentage_string(value)
        elif field_name in self.fields_in_0_to_1:
            response_value = FormatUtil.get_numeric_multiplied_by_100(value)
        elif field_name in self.fields_to_float:
            response_value = float(value)
        return response_value

    @staticmethod
    def get_transformed_value_for_arithmetic_operator(field1_value, field2_value, operation):
        if operation == operator.truediv and field2_value == 0:
            return Payload.DEFAULT_FLOAT
        return round(operation(field1_value, field2_value), Payload.DEFAULT_DECIMAL_PLACES)