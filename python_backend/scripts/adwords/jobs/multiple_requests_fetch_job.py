import logging as log
from datetime import datetime

import scripts
from lib.adwords.oauth_service.fetch_service import FetchService
from lib.utils.json import JsonUtil
from lib.utils.sync_util import SyncUtil
from .base_job import BaseJob
from .. import EXTRACT, LATENCY_COUNT, REQUEST_COUNT, LOAD, RECORDS_COUNT
# TODO: Exception getting caught at job scheduler will lose the stats at JobLevel.


class MultipleRequestsFetchJob(BaseJob):
    # Class Variables
    FIELDS = []
    SERVICE_NAME = ""
    ENTITY_TYPE = ""

    # Constants
    TIMESTAMP_FIELD = "timestamp"

    def __init__(self, next_info):
        next_info["extract_load_timestamps"] = self.get_extract_load_timestamps(next_info)
        super().__init__(next_info)
        # Usage - 1.Extract from source into in memory. 2. Message passing for extract-load task.
        # self._report_string in reports_fetch_job is equal to self._rows in multiple_request_job.
        self._rows = None

    @staticmethod
    def get_extract_load_timestamps(next_info):
        if next_info.get("first_run"):
            return SyncUtil.get_next_timestamps_for_run(next_info.get("first_run"), None, None,
                                                        next_info.get("last_timestamp"), next_info.get("doc_type_alias")
                                                        )
        else:
            return [next_info.get("next_timestamp")]

    def process_entity(self, selector, entity):
        """ Override this in the sub classes. """
        pass

    def extract_task(self):
        if self.PROCESS_JOB:
            # Extract Phase
            self.log_status_of_job("extract", "started")
            start_time = datetime.now()

            service = FetchService(scripts.adwords.CONFIG.ADWORDS_OAUTH).get_service(self.SERVICE_NAME,
                                                                                     self._refresh_token,
                                                                                     self._customer_acc_id)
            offset = 0
            selector = {
                "fields": self.FIELDS,
                "paging": {
                    "startIndex": str(offset),
                    "numberResults": str(self.PAGE_SIZE)
                }
            }
            rows, request_metric, records_metric = self.extract_entities(service, selector)
            end_time = datetime.now()
            latency_metric = (end_time - start_time).total_seconds()
            self.update_extract_phase_metrics(EXTRACT, REQUEST_COUNT, self._project_id, self._doc_type, request_metric)
            self.update_extract_phase_metrics(EXTRACT, LATENCY_COUNT, self._project_id, self._doc_type, latency_metric)

            # Load Phase
            start_time = datetime.now()
            self.write_after_extract(rows)
            end_time = datetime.now()
            latency_metric = (end_time - start_time).total_seconds()
            self.log_status_of_job("extract", "completed")
            self.update_load_phase_metrics(EXTRACT, LATENCY_COUNT, self._project_id, self._doc_type, latency_metric)
            return

    def transform_and_load_task(self, ran_extract):
        for timestamp in self._extract_load_timestamps:
            # Extract Phase
            self.log_status_of_job("load", "started")
            start_time = datetime.now()
            rows = self.read_for_load(ran_extract, timestamp)
            end_time = datetime.now()
            latency_metric = (end_time - start_time).total_seconds()
            self.update_extract_phase_metrics(LOAD, REQUEST_COUNT, self._project_id, self._doc_type, 1)
            self.update_extract_phase_metrics(LOAD, LATENCY_COUNT, self._project_id, self._doc_type, latency_metric)

            # Load Phase
            start_time = datetime.now()
            load_response = self.add_records(rows, timestamp)
            if not load_response.ok:
                return

            end_time = datetime.now()
            latency_metric = (end_time - start_time).total_seconds()
            self.update_load_phase_metrics(LOAD, LATENCY_COUNT, self._project_id, self._doc_type, latency_metric)
            self.log_status_of_job("load", "completed")
            return

    def extract_entities(self, service, selector):
        more_pages = True
        rows = []
        total_requests = 0
        offset = 0
        total_no_of_entries = 0
        while more_pages:
            current_rows, total_no_of_entries = self.extract_entities_of_single_page(service, selector)

            offset += self.PAGE_SIZE
            selector["paging"]["startIndex"] = str(offset)
            more_pages = (offset < int(total_no_of_entries))
            rows.extend(current_rows)
            total_requests += 1

        return rows, total_requests, total_no_of_entries

    def extract_entities_of_single_page(self, service, selector):
        rows = []
        page = service.get(selector)

        # Display results.
        if "entries" in page:
            for entity in page["entries"]:
                doc = self.process_entity(selector, entity)
                rows.append(doc)
            log.warning("Entities of %s were fetched at offset: %s", self.ENTITY_TYPE, selector["paging"]["startIndex"])
        else:
            log.warning("No more entities of %s were found at offset: %s", self.ENTITY_TYPE,
                        selector["paging"]["startIndex"])

        return rows, int(page["totalNumEntries"])

    def write_after_extract(self, rows):
        self._rows = rows
        rows_string = JsonUtil.create(rows)
        for timestamp in self._extract_load_timestamps:
            job_storage = scripts.adwords.CONFIG.ADWORDS_APP.job_storage
            job_storage.write(rows_string, timestamp, self._project_id, self._customer_acc_id, self._doc_type)
            self.update_load_phase_metrics(EXTRACT, REQUEST_COUNT, self._project_id, self._doc_type, 1)
            self.update_load_phase_metrics(EXTRACT, RECORDS_COUNT, self._project_id, self._doc_type, len(rows))

    def read_for_load(self, ran_extract, timestamp):
        if ran_extract:
            rows = self._rows
        else:
            job_storage = scripts.adwords.CONFIG.ADWORDS_APP.job_storage
            rows_string = job_storage.read(timestamp, self._project_id, self._customer_acc_id, self._doc_type)
            rows = JsonUtil.read(rows_string)

        return rows
