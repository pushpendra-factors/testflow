import io
import logging as log
import operator
from datetime import datetime

import scripts
import hashlib
from lib.gsc.oauth_service.fetch_service import FetchService
from lib.utils.csv import CsvUtil
from lib.utils.adwords.format import FormatUtil
from lib.utils.string import StringUtil
from lib.utils.json import JsonUtil
from lib.utils.time import TimeUtil
from .base_job import BaseJob
from .. import EXTRACT, REQUEST_COUNT, LATENCY_COUNT, LOAD, RECORDS_COUNT



# REPORT type is different from load ReportType. Eg - underscores.


class ReportsFetch(BaseJob):
    DIMENSIONS = []

    def __init__(self, next_info):
        next_info["extract_load_timestamps"] = [next_info.get("next_timestamp")]
        super().__init__(next_info)
        # Usage - 1.Extract from source into in memory. 2. Message passing for extract-load task.
        self._rows = None

    def extract_task(self):
        # Extract Phase
        self.log_status_of_job("extract", "started")
        records_metric, latency_metric = 0, 0
        start_time = datetime.now()
        str_timestamp = TimeUtil.convert_timestamp_to_gsc_date_parameter(self._next_timestamp)
        during = str_timestamp + "," + str_timestamp
        downloader = FetchService(scripts.gsc.CONFIG.GSC_OAUTH).get_webmasters_service(self._refresh_token)
        if downloader is None:
            self.log_status_of_job("extract", "not completed")
            raise Exception("Unable to generate google services client")

        response_rows = []
        row_start = 0
        request = {
            'startDate': str_timestamp,
            'endDate': str_timestamp,
            'dimensions': self.DIMENSIONS,
            'rowLimit': 1000,
            'startRow': row_start
        }
        response = downloader.searchanalytics().query(
            siteUrl=self._url_prefix, body=request).execute()
        if ('rows' not in response) or (len(response['rows']) == 0):
            err_string = "Empty response from api for date " + str_timestamp
            days_difference = TimeUtil.get_difference_from_current_day(str_timestamp)
            if days_difference <= 5:
                self.log_status_of_job("extract", "not completed")
                raise Exception(err_string)
            else:
                log.warning(err_string)

        # pagination 
        while 'rows' in response:
            response_rows= response_rows + response['rows']
            row_start += 1000
            request['startRow']= row_start
            response = downloader.searchanalytics().query(
            siteUrl=self._url_prefix, body=request).execute()
            if ('rows' not in response) or (len(response['rows']) == 0):
                log.warning("search_console: response: "+str(response) + " project_id: " + str(self._project_id) +" siteUrl: " + self._url_prefix + " request: " + str(request))

        # adding hash
        for i in range(len(response_rows)):
            hashKey = ""
            for j in range(len(self.DIMENSIONS)):
                hashKey += response_rows[i]["keys"][j]
            hash_object = hashlib.md5(hashKey.encode())
            response_rows[i]["id"] = hash_object.hexdigest()

        # Load Phase
        start_time = datetime.now()
        self.write_after_extract(response_rows)
        end_time = datetime.now()
        latency_metric = (end_time - start_time).total_seconds()
        self.update_to_file_metrics(EXTRACT, LATENCY_COUNT, self._project_id, self._url_prefix, latency_metric)
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
            self.update_to_in_memory_metrics(LOAD, REQUEST_COUNT, self._project_id, self._url_prefix, 1)
            self.update_to_in_memory_metrics(LOAD, LATENCY_COUNT, self._project_id, self._url_prefix, latency_metric)
            if rows is None:
                self.log_status_of_job("load", "not completed")
                return
            # Load Phase
            start_time = datetime.now()
            
            transformed_rows = self.transform_entities(rows)
            
            load_response = self.add_records(transformed_rows, timestamp)
            if load_response is None or not load_response.ok:
                self.log_status_of_job("load", "not completed")
                return

            end_time = datetime.now()
            latency_metric = (end_time - start_time).total_seconds()
            self.update_to_file_metrics(LOAD, LATENCY_COUNT, self._project_id, self._url_prefix, latency_metric)
            self.log_status_of_job("load", "completed")
            return

    def write_after_extract(self, rows):
        self._rows = rows
        rows_string = JsonUtil.create(rows)
        for timestamp in self._extract_load_timestamps:
            job_storage = scripts.gsc.CONFIG.GSC_APP.job_storage
            job_storage.write_gsc(rows_string, timestamp, self._project_id, self._url_prefix, self._doc_type)
            self.update_to_file_metrics(EXTRACT, REQUEST_COUNT, self._project_id, self._doc_type, 1)
            self.update_to_file_metrics(EXTRACT, RECORDS_COUNT, self._project_id, self._doc_type, len(rows))
    
    def read_for_load(self, ran_extract, timestamp):
        if ran_extract:
            rows = self._rows
        else:
            job_storage = scripts.gsc.CONFIG.GSC_APP.job_storage
            rows_string = job_storage.read_gsc(timestamp, self._project_id, self._url_prefix, self._doc_type)
            rows = JsonUtil.read(rows_string)
        return rows

    def transform_entities(self, rows):
        transformed_rows = []
        for row in rows:
            transformed_row = {}
            for key in row:
                if key == "keys":
                    for index in range(len(self.DIMENSIONS)):
                        transformed_row[self.DIMENSIONS[index]] = row["keys"][index]
                else:
                    transformed_row[key] = row[key]
            transformed_rows.append(transformed_row)
        
        return transformed_rows

        
