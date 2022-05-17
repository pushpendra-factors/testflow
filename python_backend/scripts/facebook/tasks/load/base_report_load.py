import logging as log
import traceback
from datetime import datetime

from lib.utils.facebook.metrics_aggregator import MetricsAggregator
from lib.utils.facebook.task_stats import TaskStats
from .base_load import BaseLoad
from ..context.load.base_load import BaseLoad as BaseLoadContext


# Not following interface rule yet for merging dependencies.
class BaseReportLoad(BaseLoad):

    @classmethod
    def get_instance(cls):
        if cls.INSTANCE is None:
            cls.INSTANCE = BaseReportLoad()
        return cls.INSTANCE

    def execute(self, task_context):
        if len(task_context.get_next_timestamps()) == 0:
            MetricsAggregator.update_job_stats(task_context.project_id, task_context.customer_account_id,
                                               task_context.type_alias, "skipped", "")
            return

        for curr_timestamp in task_context.get_next_timestamps():
            try:
                task_context.add_curr_timestamp(curr_timestamp)
                task_context.add_log("started")
                start_time = datetime.now()

                self.read_dependencies(task_context, curr_timestamp)
                self.read_current_task_records(task_context, curr_timestamp)

                end_time = datetime.now()
                latency_metric = (end_time - start_time).total_seconds()
                MetricsAggregator.update_task_stats(BaseLoadContext.TASK_TYPE, TaskStats.TO_IN_MEMORY,
                                                    TaskStats.LATENCY_COUNT,
                                                    task_context.project_id, task_context.type_alias, latency_metric)
                MetricsAggregator.update_task_stats(BaseLoadContext.TASK_TYPE, TaskStats.TO_IN_MEMORY,
                                                    TaskStats.REQUEST_COUNT,
                                                    task_context.project_id, task_context.type_alias, 1)

                self.merge_dependencies_and_current_task_records(task_context)

                start_time = datetime.now()

                self.write_to_destinations(task_context)

                task_context.add_curr_timestamp(curr_timestamp)
                task_context.add_log("completed")
                end_time = datetime.now()
                latency_metric = (end_time - start_time).total_seconds()
                MetricsAggregator.update_task_stats(BaseLoadContext.TASK_TYPE, TaskStats.TO_FILE, TaskStats.LATENCY_COUNT,
                                                    task_context.project_id, task_context.type_alias, latency_metric)

                MetricsAggregator.update_job_stats(task_context.project_id, task_context.customer_account_id,
                                                   task_context.type_alias, "success", "")
            except Exception as e:
                traceback.print_tb(e.__traceback__)
                str_exception = str(e)
                message = str_exception
                log.warning("Failed with exception: %d %s %s", task_context.project_id,
                            task_context.customer_account_id, message)
                if "No such object" in message and "HTTPStatus.PARTIAL_CONTENT" in message and "facebook_extract" in message:
                    message = "Failed to load from cloud storage"
                MetricsAggregator.update_job_stats(task_context.project_id, task_context.customer_account_id,
                                                   task_context.type_alias, "failed", message)
        return

    @staticmethod
    def read_dependencies(task_context, curr_timestamp):
        task_context.read_dependencies(curr_timestamp)

    @staticmethod
    def read_current_task_records(task_context, curr_timestamp):
        task_context.read_current_records(curr_timestamp)

    @staticmethod
    def merge_dependencies_and_current_task_records(task_context):
        task_context.merge_dependencies_and_current_task_records()

    @staticmethod
    def write_to_destinations(task_context):
        task_context.add_destination_attributes()
        task_context.write_records()
        return
