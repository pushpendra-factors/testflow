from datetime import datetime

from lib.utils.json import JsonUtil

import scripts
from lib.adwords.oauth_service.fetch_service import FetchService
from lib.utils.csv import CsvUtil
from .base_job import BaseJob
from .. import EXTRACT, REQUEST_COUNT, LATENCY_COUNT, LOAD, RECORDS_COUNT

from .query import QueryBuilder


class ServicesFetch(BaseJob):

    EXTRACT_FIELDS = None
    HEADERS_VMAX = None
    HEADERS_V00 = None
    REPORT = None
    MAX_VERSION = "##V01"
    current_version = "##V01"

    FIELDS = "fields"
    FIELD = "field"
    MAP = "map"
    OPERATION = "operation"
    TRANSFORM_FIELDS_V01 = []
    TRANSFORM_MAP_V01 = []

    def __init__(self, next_info):
        next_info["extract_load_timestamps"] = [next_info.get("next_timestamp")]
        super().__init__(next_info)
        # Usage - 1.Extract from system into in memory. 2. Message passing for extract-load task.
        self._rows = None


    def extract_task(self):
        if self.PROCESS_JOB:
            # Extract Phase
            self.log_status_of_job("extract", "started")
            records_metric, latency_metric = 0, 0
            start_time = datetime.now()
        
            service_query = (QueryBuilder()
                                    .Select(self.EXTRACT_FIELDS)
                                    .From(self.REPORT)
                                    # .Limit(1)
                                    .Build())

            ads_service = FetchService(scripts.adwords.CONFIG.ADWORDS_OAUTH).new_get_service(
                                                        "GoogleAdsService", self._refresh_token)
            stream = ads_service.search_stream(customer_id=self._customer_acc_id, query=service_query)
            report = self.MAX_VERSION + CsvUtil.stream_to_csv(
                                                    self.EXTRACT_FIELDS, self.HEADERS_VMAX, stream)

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
        if self.PROCESS_JOB:
            for timestamp in self._extract_load_timestamps:
                self.log_status_of_job("load", "started")
                start_time = datetime.now()
                rows, self.current_version = self.read_for_load(ran_extract, timestamp)
                end_time = datetime.now()
                latency_metric = (end_time - start_time).total_seconds()
                self.update_to_in_memory_metrics(LOAD, REQUEST_COUNT, self._project_id, self._doc_type, 1)
                self.update_to_in_memory_metrics(LOAD, LATENCY_COUNT, self._project_id, self._doc_type, latency_metric)
            
                # Load Phase
                start_time = datetime.now()
                if self.current_version == "##V00":
                    pass
                elif self.current_version == self.MAX_VERSION:
                    rows = CsvUtil.csv_to_dict_list(self.HEADERS_VMAX, rows)
                    rows = self.transform_entities(rows)
                else:
                    return

                load_response = self.add_records(rows, timestamp)
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
        for transform in self.TRANSFORM_FIELDS_V01:
            fields = transform[self.FIELDS]
            operation = transform[self.OPERATION]
            for field in fields:
                if(field in row and row[field] != ''):
                    row[field] = operation(row[field])

        for transform in self.TRANSFORM_MAP_V01:
            field = transform[self.FIELD]
            if(row[field] != ''):
                row[field] = transform[self.MAP][row[field]]
        return row

    def write_after_extract(self, rows_string):
        rows, self.current_version = CsvUtil.unmarshall(rows_string)
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
            if len(result_string) > 0 and "##V" in result_string[:10]:
                rows, version = CsvUtil.unmarshall(result_string)
            else:
                version = "##V00"
                rows = JsonUtil.read(result_string)
        return rows, version
