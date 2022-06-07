import logging as log
import traceback
from datetime import datetime

from lib.utils.facebook.metrics_aggregator import MetricsAggregator
from lib.utils.facebook.task_stats import TaskStats
from .base_extract import BaseExtract
from ..context.extract.base_extract import BaseExtract as BaseExtractContext


# There is a difference between load and extract during execute i.e.extract has all
# execution steps in here, but load has all execution steps in context.
class BaseReportExtract(BaseExtract):

    @classmethod
    def get_instance(cls):
        if cls.INSTANCE is None:
            cls.INSTANCE = BaseReportExtract()
        return cls.INSTANCE

    def execute(self, task_context):
        if len(task_context.get_next_timestamps()) == 0:
            MetricsAggregator.update_job_stats(task_context.project_id, task_context.customer_account_id,
                                               task_context.type_alias, "skipped", "")
            return
        current_timestamp = None
        try:
            for curr_timestamp in task_context.get_next_timestamps():
                current_timestamp = curr_timestamp
                task_context.add_curr_timestamp(curr_timestamp)
                task_context.add_log("started")
                task_context.reset_total_number_of_records()
                task_context.reset_total_number_of_async_requests()
                start_time = datetime.now()

                read_records_status = task_context.read_records()
                if read_records_status != "success":
                    return

                end_time = datetime.now()
                latency_metric = (end_time - start_time).total_seconds()
                MetricsAggregator.update_task_stats(BaseExtractContext.TASK_TYPE, TaskStats.TO_IN_MEMORY,
                                                    TaskStats.LATENCY_COUNT,
                                                    task_context.project_id, task_context.type_alias, latency_metric)
                MetricsAggregator.update_task_stats(BaseExtractContext.TASK_TYPE, TaskStats.TO_IN_MEMORY,
                                                    TaskStats.REQUEST_COUNT,
                                                    task_context.project_id, task_context.type_alias, task_context.total_number_of_records)
                MetricsAggregator.update_task_stats(BaseExtractContext.TASK_TYPE, TaskStats.TO_IN_MEMORY,
                                                    TaskStats.ASYNC_REQUEST_COUNT,
                                                    task_context.project_id, task_context.type_alias, task_context.total_number_of_async_requests)

                start_time = datetime.now()

                self.save_to_destinations(task_context)

                task_context.add_log("completed")
                end_time = datetime.now()
                latency_metric = (end_time - start_time).total_seconds()
                MetricsAggregator.update_task_stats(BaseExtractContext.TASK_TYPE, TaskStats.TO_FILE,
                                                    TaskStats.LATENCY_COUNT,
                                                    task_context.project_id, task_context.type_alias, latency_metric)
                MetricsAggregator.update_job_stats(task_context.project_id, task_context.customer_account_id,
                                                   task_context.type_alias, "success", "")
        except Exception as e:
            traceback.print_tb(e.__traceback__)
            str_exception = str(e)
            message = str_exception
            message = "Timestamp: " + str(current_timestamp) + ". Message: " + str_exception
            log.warning("Failed with exception: %d %s %s", task_context.project_id,
                        task_context.customer_account_id, message)
            MetricsAggregator.update_job_stats(task_context.project_id, task_context.customer_account_id,
                                                task_context.type_alias, "failed", message)
        return

    # Later can add decorators for tracking.
    @staticmethod
    def save_to_destinations(task_context):
        task_context.add_destination_attributes()
        task_context.write_records()
        return
