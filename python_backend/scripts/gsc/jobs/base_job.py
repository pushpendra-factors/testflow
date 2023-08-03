import logging as log

import scripts
from lib.data_services.factors_data_service import FactorsDataService
from scripts.gsc import TO_IN_MEMORY, TO_FILE, SUCCESS_MESSAGE, EXTRACT, RECORDS_COUNT, LOAD, REQUEST_COUNT
import uuid


# Note: If the number of custom paths exceed 5 in the subClasses. Move it to strategic pattern.


class BaseJob:
    PAGE_SIZE = 200
    PROCESS_JOB = True

    def __init__(self, next_info):
        self._project_id = next_info.get("project_id")
        self._url_prefix = next_info.get("url_prefix")
        self._refresh_token = next_info.get("refresh_token")
        self._last_extract_timestamp = next_info.get("last_timestamp")
        self._next_timestamp = next_info.get("next_timestamp")
        self._extract_load_timestamps = next_info.get("extract_load_timestamps")
        self._first_run = next_info.get("first_run")
        self._doc_type = next_info.get("type")

    def start(self):
        log.warning("ETL for project: %s, url: %s, doc_type: %s, timestamp: %s",
                    str(self._project_id), self._url_prefix, str(self._doc_type), str(self._next_timestamp))
        self.execute()

    # TODO handle error cases
    def execute(self):
        metrics_controller = scripts.gsc.CONFIG.GSC_APP.metrics_controller
        if scripts.gsc.CONFIG.GSC_APP.type_of_run == scripts.gsc.EXTRACT_AND_LOAD:
            self.extract_and_load_task()
        elif scripts.gsc.CONFIG.GSC_APP.type_of_run == scripts.gsc.EXTRACT:
            self.extract_task()
        else:
            self.transform_and_load_task(False)
        metrics_controller.update_gsc_job_stats(self._project_id, self._url_prefix, self._doc_type, SUCCESS_MESSAGE)

    def extract_and_load_task(self):
        self.extract_task()
        self.transform_and_load_task(True)

    def extract_task(self):
        """ Override this method to provide extract functionality. """
        pass

    def transform_and_load_task(self, ran_extract):
        """ Override this method to provide transform and load functionality. """
        pass

    def add_records(self, records, timestamp):
        if len(records) > 0:
            if scripts.gsc.CONFIG.GSC_APP.dry:
                log.error("Dry run. Skipped add gsc documents to db.")
                response = None
            else:
                response = FactorsDataService.add_all_gsc_documents(self._project_id, self._url_prefix, self._doc_type, records, timestamp)
                if not response.ok:
                    return response
        else:
            response = FactorsDataService.add_gsc_document(self._project_id, self._url_prefix, self._doc_type, {"id": str(uuid.uuid4())}, timestamp)
        self.update_to_file_metrics(LOAD, REQUEST_COUNT, self._project_id, self._url_prefix, 1)
        self.update_to_file_metrics(LOAD, RECORDS_COUNT, self._project_id, self._url_prefix, len(records))
        return response

    def log_status_of_job(self, job_type, status):
        log.warning("%s %s of job for Project Id: %s, Timestamp: %d, URL: %s", status, job_type, self._project_id,
                    self._next_timestamp, self._url_prefix)

    @staticmethod
    def update_to_in_memory_metrics(task, metric_type, project_id, url_prefix, value):
        scripts.gsc.CONFIG.GSC_APP.metrics_controller.update_task_stats(task, TO_IN_MEMORY, metric_type,
                                                                                project_id, url_prefix, value)

    # REQUESTS_COUNT - computed at job level not data service(add_gsc_doc) level
    @staticmethod
    def update_to_file_metrics(task, metric_type, project_id, url_prefix, value):
        scripts.gsc.CONFIG.GSC_APP.metrics_controller.update_task_stats(task, TO_FILE, metric_type,
                                                                                project_id, url_prefix, value)
