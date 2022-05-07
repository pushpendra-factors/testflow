import logging as log

import scripts
from lib.data_services.factors_data_service import FactorsDataService
from scripts.adwords import TO_IN_MEMORY, TO_FILE, SUCCESS_MESSAGE, EXTRACT, RECORDS_COUNT, LOAD, REQUEST_COUNT


# Note: If the number of custom paths exceed 5 in the subClasses. Move it to strategic pattern.


class BaseJob:
    PROCESS_JOB = True

    def __init__(self, next_info):
        self._project_id = next_info.get("project_id")
        self._customer_acc_id = next_info.get("customer_acc_id")
        self._refresh_token = next_info.get("refresh_token")
        self._last_extract_timestamp = next_info.get("last_timestamp")
        self._next_timestamp = next_info.get("next_timestamp")
        self._extract_load_timestamps = next_info.get("extract_load_timestamps")
        self._doc_type = next_info.get("doc_type_alias")
        self._first_run = next_info.get("first_run")
        self._manager_id = next_info.get("manager_id")

    def start(self):
        log.warning("ETL for project: %s, cutomer_account_id: %s, document_type: %s, timestamp: %s",
                    str(self._project_id), self._customer_acc_id, self._doc_type, str(self._next_timestamp))
        self.execute()

    # TODO handle error cases
    def execute(self):
        metrics_controller = scripts.adwords.CONFIG.ADWORDS_APP.metrics_controller
        if scripts.adwords.CONFIG.ADWORDS_APP.type_of_run == scripts.adwords.EXTRACT_AND_LOAD:
            self.extract_and_load_task()
        elif scripts.adwords.CONFIG.ADWORDS_APP.type_of_run == scripts.adwords.EXTRACT:
            self.extract_task()
        else:
            self.transform_and_load_task(False)
        metrics_controller.update_job_stats(self._project_id, self._customer_acc_id, self._doc_type, SUCCESS_MESSAGE)

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
            if scripts.adwords.CONFIG.ADWORDS_APP.dry:
                log.error("Dry run. Skipped add adwords documents to db.")
                response = None
            else:
                response = FactorsDataService.add_all_adwords_documents(self._project_id, self._customer_acc_id, records,
                                                             self._doc_type, timestamp)
        else:
            response = FactorsDataService.add_adwords_document(self._project_id, self._customer_acc_id, {}, self._doc_type,
                                                    timestamp)
        self.update_to_file_metrics(LOAD, REQUEST_COUNT, self._project_id, self._doc_type, 1)
        self.update_to_file_metrics(LOAD, RECORDS_COUNT, self._project_id, self._doc_type, len(records))
        return response

    def log_status_of_job(self, job_type, status):
        log.warning("%s %s of job for Project Id: %s, Timestamp: %d, Doc Type: %s", status, job_type, self._project_id,
                    self._next_timestamp, self._doc_type)

    @staticmethod
    def update_to_in_memory_metrics(task, metric_type, project_id, doc_type, value):
        scripts.adwords.CONFIG.ADWORDS_APP.metrics_controller.update_task_stats(task, TO_IN_MEMORY, metric_type,
                                                                                project_id, doc_type, value)

    # REQUESTS_COUNT - computed at job level not data service(add_adwords_doc) level
    @staticmethod
    def update_to_file_metrics(task, metric_type, project_id, doc_type, value):
        scripts.adwords.CONFIG.ADWORDS_APP.metrics_controller.update_task_stats(task, TO_FILE, metric_type,
                                                                                project_id, doc_type, value)
