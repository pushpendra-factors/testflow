import logging as log
import traceback
from datetime import datetime

from lib.utils.facebook.metrics_aggregator import MetricsAggregator
from lib.utils.facebook.task_stats import TaskStats
from .base_load import BaseLoad
from ..context.load.base_load import BaseLoad as BaseLoadContext


class BaseInfoLoad(BaseLoad):

    @classmethod
    def get_instance(cls):
        if cls.INSTANCE is None:
            cls.INSTANCE = BaseInfoLoad()
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
                start_time = datetime.now()

                self.read_current_task_records(task_context)

                end_time = datetime.now()
                latency_metric = (end_time - start_time).total_seconds()
                MetricsAggregator.update_task_stats(BaseLoadContext.TASK_TYPE, TaskStats.TO_IN_MEMORY,
                                                    TaskStats.LATENCY_COUNT,
                                                    task_context.project_id, task_context.type_alias, latency_metric)
                MetricsAggregator.update_task_stats(BaseLoadContext.TASK_TYPE, TaskStats.TO_IN_MEMORY,
                                                    TaskStats.REQUEST_COUNT,
                                                    task_context.project_id, task_context.type_alias, 1)

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
            message = "Timestamp: " + str(current_timestamp) + ". Message: " + str_exception
            log.warning("Failed with exception: %d %s %s", task_context.project_id,
                        task_context.customer_account_id, message)
            MetricsAggregator.update_job_stats(task_context.project_id, task_context.customer_account_id,
                                                task_context.type_alias, "failed", message)
        return

    def read_current_task_records(self, task_context):
        task_context.add_source_attributes()
        task_context.read_records()
        return

    def write_to_destinations(self, task_context):
        task_context.add_destination_attributes()
        task_context.write_records()
        return
