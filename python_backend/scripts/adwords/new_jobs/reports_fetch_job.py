from datetime import datetime

import scripts
from lib.adwords.oauth_service.fetch_service import FetchService
from lib.utils.csv import CsvUtil
from .base_job import BaseJob
from .payload import Payload
from .. import EXTRACT, REQUEST_COUNT, LATENCY_COUNT, LOAD, RECORDS_COUNT

from .query import QueryBuilder

# REPORT type is different from load ReportType. Eg - underscores.


class ReportsFetch(BaseJob):

    # New Parameters
    EXTRACT_FIELDS = None
    # VMAX = latest set of headers
    HEADERS_VMAX = None
    HEADERS_V00 = None # adwords api headers
    HEADERS_V01 = None # First iteration of ads api
    HEADERS_V02 = None # additional metrics for campaign perf, Adgroup perf and Keyword perf report. For other reports HEADERS_V01 = HEADERS_V02 = HEADERS_VMAX
    REPORT = None
    MAX_VERSION = "##V02"
    # changelog: from V01 -> V02,  2 new metrics being pulled in campaign perf, Adgroup perf and Keyword perf report.

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
        "impressions": None,
        "absolute_top_impression_percentage": None,
        "top_impression_percentage": None
    }

    FIELDS = "fields"
    FIELD = "field"
    MAP = "map"
    OPERATION = "operation"
    TRANSFORM_MAP_V01 = []

    FIELDS_WITH_STATUS = []
    FIELDS_WITH_BOOLEAN = []
    FIELDS_WITH_RESOURCE_NAME = []
    FIELDS_TO_PERCENTAGE = []
    FIELDS_WITH_INTERACTION_TYPES = []
    # PR: 6173 - Though this has change, customers are not using it.
    FIELDS_WITH_APPROVAL_STATUS = []


    def __init__(self, next_info):
        next_info["extract_load_timestamps"] = [next_info.get("next_timestamp")]
        super().__init__(next_info)
        # Usage - 1.Extract from system into in memory. 2. Message passing for extract-load task.
        self._rows = None


    def extract_task(self):
        self.log_status_of_job("extract", "started", self._next_timestamp)
        records_metric, latency_metric = 0, 0
        start_time = datetime.now()
        str_timestamp = str(self._next_timestamp)
        str_timestamp = '"' + str_timestamp[0:4] + "-" + str_timestamp[4:6] + "-" + str_timestamp[6:8] + '"'
        during = str_timestamp + " AND " + str_timestamp
        
        self.log_status_of_job("query builder", "started", self._next_timestamp)
        if self.REPORT == "click_view":
            report_query = (QueryBuilder()
                                    .Select(self.EXTRACT_FIELDS)
                                    .From(self.REPORT)
                                    .During(during)
                                    # .Limit(1)
                                    .Build())
        else:
            report_query = (QueryBuilder()
                                    .Select(self.EXTRACT_FIELDS)
                                    .From(self.REPORT)
                                    .Where('metrics.impressions > 0')
                                    # .Limit(1)
                                    .During(during)
                                    .Build())
        self.log_status_of_job("query builder", "completed", self._next_timestamp)
        
        self.log_status_of_job("fetch service", "started", self._next_timestamp)
        ads_service = FetchService(scripts.adwords.CONFIG.ADWORDS_OAUTH).new_get_service(
                                                    "GoogleAdsService", self._refresh_token, self._manager_id)
        self.log_status_of_job("fetch service", "completed", self._next_timestamp)
        
        self.log_status_of_job("search stream", "started", self._next_timestamp)
        stream = ads_service.search_stream(customer_id=self._customer_acc_id, query=report_query)
        self.log_status_of_job("search stream", "completed", self._next_timestamp)
        
        self.log_status_of_job("stream to csv", "started", self._next_timestamp)
        report = self.MAX_VERSION + CsvUtil.stream_to_csv(
                                            self.EXTRACT_FIELDS, self.HEADERS_VMAX, stream)
        self.log_status_of_job("stream to csv", "completed", self._next_timestamp)

        end_time = datetime.now()
        latency_metric = (end_time - start_time).total_seconds()
        self.log_status_of_job("upload to in memory", "started", self._next_timestamp)
        self.update_to_in_memory_metrics(EXTRACT, REQUEST_COUNT, self._project_id, self._doc_type, 1)
        self.update_to_in_memory_metrics(EXTRACT, LATENCY_COUNT, self._project_id, self._doc_type, latency_metric)
        self.log_status_of_job("upload to in memory", "completed", self._next_timestamp)

        # Load Phase
        start_time = datetime.now()
        self.log_status_of_job("load to cloud", "started", self._next_timestamp)
        self.write_after_extract(report)
        self.log_status_of_job("load to cloud", "completed", self._next_timestamp)
        end_time = datetime.now()
        latency_metric = (end_time - start_time).total_seconds()
        self.update_to_file_metrics(EXTRACT, LATENCY_COUNT, self._project_id, self._doc_type, latency_metric)
        self.log_status_of_job("extract", "completed", self._next_timestamp)
        return

    def get_payload(self, rows, version):
        headers = None
        fields_with_percentages = None
        fields_in_0_to_1 = None
        fields_to_float = None
        transform_map = None
        fields_with_status = None
        fields_with_boolean = None
        fields_with_resource_name = None
        fields_to_percentage = None
        fields_with_interaction_types = None

        if version == "##V00":
            headers = self.HEADERS_V00
            fields_with_percentages = self.FIELDS_WITH_PERCENTAGES
            fields_in_0_to_1 = self.FIELDS_IN_0_TO_1
            fields_to_float = self.FIELDS_TO_FLOAT
            fields_with_status = []
            fields_with_boolean = []
            fields_with_resource_name = []
            fields_to_percentage = []
            fields_with_interaction_types = []
            fields_with_aprroval_status = []
            transform_map = []
        elif version == "##V01":
            headers = self.HEADERS_V01
            fields_with_percentages = []
            fields_in_0_to_1 = self.FIELDS_IN_0_TO_1
            fields_in_0_to_1.update(self.FIELDS_WITH_PERCENTAGES)
            fields_to_float = self.FIELDS_TO_FLOAT
            fields_with_status = self.FIELDS_WITH_STATUS
            fields_with_boolean = self.FIELDS_WITH_BOOLEAN
            fields_with_resource_name = self.FIELDS_WITH_RESOURCE_NAME
            fields_to_percentage = self.FIELDS_TO_PERCENTAGE
            fields_with_interaction_types = self.FIELDS_WITH_INTERACTION_TYPES
            fields_with_aprroval_status = self.FIELDS_WITH_APPROVAL_STATUS
            transform_map = self.TRANSFORM_MAP_V01
        elif version == "##V02":
            headers = self.HEADERS_V02
            fields_with_percentages = []
            fields_in_0_to_1 = self.FIELDS_IN_0_TO_1
            fields_in_0_to_1.update(self.FIELDS_WITH_PERCENTAGES)
            fields_to_float = self.FIELDS_TO_FLOAT
            fields_with_status = self.FIELDS_WITH_STATUS
            fields_with_boolean = self.FIELDS_WITH_BOOLEAN
            fields_with_resource_name = self.FIELDS_WITH_RESOURCE_NAME
            fields_to_percentage = self.FIELDS_TO_PERCENTAGE
            fields_with_interaction_types = self.FIELDS_WITH_INTERACTION_TYPES
            fields_with_aprroval_status = self.FIELDS_WITH_APPROVAL_STATUS
            transform_map = self.TRANSFORM_MAP_V01
        else:
            return None

        rows = CsvUtil.csv_to_dict_list(headers, rows)
        return Payload(headers=headers, rows=rows, 
                    fields_with_percentages=fields_with_percentages,
                    fields_in_0_to_1=fields_in_0_to_1,
                    fields_to_float=fields_to_float,
                    fields_with_status=fields_with_status,
                    fields_with_boolean = fields_with_boolean,
                    fields_with_resource_name=fields_with_resource_name,
                    fields_to_percentage=fields_to_percentage,
                    fields_with_interaction_types=fields_with_interaction_types,
                    fields_with_approval_status=fields_with_aprroval_status,
                    transform_map=transform_map)

    def transform_and_load_task(self, ran_extract):
        for timestamp in self._extract_load_timestamps:
            # Extract Phase
            self.log_status_of_job("load", "started", timestamp)
            start_time = datetime.now()
            rows, version = self.read_for_load(ran_extract, timestamp)
            end_time = datetime.now()
            latency_metric = (end_time - start_time).total_seconds()
            self.update_to_in_memory_metrics(LOAD, REQUEST_COUNT, self._project_id, self._doc_type, 1)
            self.update_to_in_memory_metrics(LOAD, LATENCY_COUNT, self._project_id, self._doc_type, latency_metric)
           
            # Load Phase
            start_time = datetime.now()
            payload = self.get_payload(rows, version)
            if(payload is None):
                return

            transformed_rows = None
            if self.REPORT == "click_view":
                transformed_rows = payload.transform_entities_click_view()
            else:
                transformed_rows = payload.transform_entities()

            load_response = self.add_records(transformed_rows, timestamp)
            if load_response is None or not load_response.ok:
                self.log_status_of_job("load", "not completed", timestamp)
                return

            end_time = datetime.now()
            latency_metric = (end_time - start_time).total_seconds()
            self.update_to_file_metrics(LOAD, LATENCY_COUNT, self._project_id, self._doc_type, latency_metric)
            self.log_status_of_job("load", "completed", timestamp)
            return

    def write_after_extract(self, rows_string):
        rows, _ = CsvUtil.unmarshall(rows_string)
        self._rows = rows
        for timestamp in self._extract_load_timestamps:
            job_storage = scripts.adwords.CONFIG.ADWORDS_APP.job_storage
            job_storage.write(rows_string, timestamp, self._project_id, self._customer_acc_id, self._doc_type)
            self.update_to_file_metrics(EXTRACT, REQUEST_COUNT, self._project_id, self._doc_type, 1)
            self.update_to_file_metrics(EXTRACT, RECORDS_COUNT, self._project_id, self._doc_type, len(rows))

    def read_for_load(self, ran_extract, timestamp):
        version = self.MAX_VERSION
        if ran_extract:
            rows = self._rows
        else:
            job_storage = scripts.adwords.CONFIG.ADWORDS_APP.job_storage
            result_string = job_storage.read(timestamp, self._project_id, self._customer_acc_id, self._doc_type)
            rows, version = CsvUtil.unmarshall(result_string)
        return rows, version
