import operator
from datetime import datetime

from googleads import adwords

import scripts
from lib.adwords.oauth_service.fetch_service import FetchService
from lib.utils.csv import CsvUtil
from lib.utils.adwords.format import FormatUtil
from lib.utils.string import StringUtil
from .base_job import BaseJob
from .. import EXTRACT, REQUEST_COUNT, LATENCY_COUNT, LOAD, RECORDS_COUNT


# REPORT type is different from load ReportType. Eg - underscores.


class ReportsFetch(BaseJob):
    QUERY_FIELDS = []
    REPORT = ""
    DEFAULT_FLOAT = 0.000
    DEFAULT_NUMERATOR_FLOAT = 0.0
    DEFAULT_DENOMINATOR_FLOAT = 1.0
    DEFAULT_DECIMAL_PLACES = 3

    # Currently only commonFields Transformation is being done and Expression = Total--- = impressions/share---
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
         RESULT_FIELD: "total_search_rank_lost_top_impression", OPERATION: operator.truediv}
    ]

    FIELDS_WITH_PERCENTAGES = {
        "search_click_share": None,
        "search_impression_share": None,
        "search_absolute_top_impression_share": None,
        "search_budget_lost_impression_share": None,
        "search_rank_lost_impression_share": None
    }

    FIELDS_IN_0_TO_1 = {
        "search_top_impression_share": None,
        "search_budget_lost_top_impression_share": None,
        "search_budget_lost_absolute_top_impression_share": None,
        "search_rank_lost_top_impression_share": None,
        "search_rank_lost_absolute_top_impression_share": None
    }

    FIELDS_TO_FLOAT = {
        "impressions": None
    }

    def __init__(self, next_info):
        next_info["extract_load_timestamps"] = [next_info.get("next_timestamp")]
        super().__init__(next_info)
        # Usage - 1.Extract from system into in memory. 2. Message passing for extract-load task.
        self._rows = None

    def extract_task(self):
        # Extract Phase
        self.log_status_of_job("extract", "started")
        records_metric, latency_metric = 0, 0
        start_time = datetime.now()
        str_timestamp = str(self._next_timestamp)
        during = str_timestamp + "," + str_timestamp
        downloader = FetchService(scripts.adwords.CONFIG.ADWORDS_OAUTH).get_report_downloader(self._refresh_token,
                                                                                              self._customer_acc_id)
        fields = StringUtil.snake_to_pascal_case(self.QUERY_FIELDS)
        if self.REPORT == "CLICK_PERFORMANCE_REPORT":
            report_query = (adwords.ReportQueryBuilder()
                            .Select(*fields)
                            .From(self.REPORT)
                            .During(during).Build())
        else:
            report_query = (adwords.ReportQueryBuilder()
                        .Select(*fields)
                        .From(self.REPORT)
                        .Where('Impressions').GreaterThan(0)
                        .During(during).Build())
        report = downloader.DownloadReportAsStringWithAwql(
            report_query, "CSV", skip_report_header=True,
            skip_column_header=True, skip_report_summary=True)
        end_time = datetime.now()
        latency_metric = (end_time - start_time).total_seconds()
        self.update_to_in_memory_metrics(EXTRACT, REQUEST_COUNT, self._project_id, self._doc_type, 1)
        self.update_to_in_memory_metrics(EXTRACT, LATENCY_COUNT, self._project_id, self._doc_type, latency_metric)

        # Load Phase
        start_time = datetime.now()
        self.write_after_extract(report)
        end_time = datetime.now()
        latency_metric = (end_time - start_time).total_seconds()
        self.update_to_file_metrics(EXTRACT, LATENCY_COUNT, self._project_id, self._doc_type, latency_metric)
        self.log_status_of_job("extract", "completed")
        return

    def transform_and_load_task(self, ran_extract):
        for timestamp in self._extract_load_timestamps:
            # Extract Phase
            self.log_status_of_job("load", "started")
            start_time = datetime.now()
            rows = self.read_for_load(ran_extract, timestamp)
            end_time = datetime.now()
            latency_metric = (end_time - start_time).total_seconds()
            self.update_to_in_memory_metrics(LOAD, REQUEST_COUNT, self._project_id, self._doc_type, 1)
            self.update_to_in_memory_metrics(LOAD, LATENCY_COUNT, self._project_id, self._doc_type, latency_metric)

            # Load Phase
            start_time = datetime.now()
            rows = CsvUtil.csv_to_dict_list(self.QUERY_FIELDS, rows)
            rows = self.transform_entities(rows)

            transformed_rows = self.transform_entities(rows)
            load_response = self.add_records(transformed_rows, timestamp)
            if load_response is None or not load_response.ok:
                self.log_status_of_job("load", "not completed")
                return

            end_time = datetime.now()
            latency_metric = (end_time - start_time).total_seconds()
            self.update_to_file_metrics(LOAD, LATENCY_COUNT, self._project_id, self._doc_type, latency_metric)
            self.log_status_of_job("load", "completed")
            return

    def transform_entities(self, rows):
        transformed_rows = []
        for row in rows:
            transformed_rows.append(self.transform_entity(row))
        return transformed_rows

    def transform_entity(self, row):
        for transform in self.TRANSFORM_AND_ADD_NEW_FIELDS:
            field1_name = transform[self.OPERAND1]
            field2_name = transform[self.OPERAND2]
            operation = transform[self.OPERATION]
            result_field_name = transform[self.RESULT_FIELD]
            if field1_name in row and field2_name in row:
                field1_value = ReportsFetch.get_transformed_values(field1_name, row.get(field1_name, ""))
                field2_value = ReportsFetch.get_transformed_values(field2_name, row.get(field2_name, ""))
                transformed_value = self.get_transformed_value_for_arithmetic_operator(field1_value, field2_value,
                                                                                       operation)
                row[result_field_name] = transformed_value
        return row

    def write_after_extract(self, rows_string):
        rows, _ = CsvUtil.unmarshall(rows_string)
        self._rows = rows
        for timestamp in self._extract_load_timestamps:
            job_storage = scripts.adwords.CONFIG.ADWORDS_APP.job_storage
            job_storage.write(rows_string, timestamp, self._project_id, self._customer_acc_id, self._doc_type)
            self.update_to_file_metrics(EXTRACT, REQUEST_COUNT, self._project_id, self._doc_type, 1)
            self.update_to_file_metrics(EXTRACT, RECORDS_COUNT, self._project_id, self._doc_type, len(rows))

    def read_for_load(self, ran_extract, timestamp):
        if ran_extract:
            rows = self._rows
        else:
            job_storage = scripts.adwords.CONFIG.ADWORDS_APP.job_storage
            result_string = job_storage.read(timestamp, self._project_id, self._customer_acc_id, self._doc_type)
            rows, _ = CsvUtil.unmarshall(result_string)
        return rows

    # Internal Methods for Transformation. Please check context.
    @staticmethod
    def get_transformed_values(field_name, value):
        response_value = value
        if field_name in ReportsFetch.FIELDS_WITH_PERCENTAGES:
            response_value = FormatUtil.get_numeric_from_percentage_string(value)
        elif field_name in ReportsFetch.FIELDS_IN_0_TO_1:
            response_value = FormatUtil.get_numeric_multiplied_by_100(value)
        elif field_name in ReportsFetch.FIELDS_TO_FLOAT:
            response_value = float(value)
        return response_value

    @staticmethod
    def get_transformed_value_for_arithmetic_operator(field1_value, field2_value, operation):
        if operation == operator.truediv and field2_value == 0:
            return ReportsFetch.DEFAULT_FLOAT
        return round(operation(field1_value, field2_value), ReportsFetch.DEFAULT_DECIMAL_PLACES)
