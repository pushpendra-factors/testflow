from datetime import datetime

from lib.utils.facebook.metrics_aggregator import MetricsAggregator
from lib.utils.facebook.task_stats import TaskStats
from .base_extract import BaseExtract
from ..context.extract.base_extract import BaseExtract as BaseExtractContext


class BaseInfoExtract(BaseExtract):

    @classmethod
    def get_instance(cls):
        if cls.INSTANCE is None:
            cls.INSTANCE = BaseInfoExtract()
        return cls.INSTANCE

    def execute(self, task_context):
        if not task_context.run_job():
            MetricsAggregator.update_job_stats(task_context.project_id, task_context.customer_account_id,
                                               task_context.type_alias, "skipped", "")
            return

        curr_timestamp = task_context.get_next_timestamp()
        task_context.add_curr_timestamp(curr_timestamp)
        task_context.reset_total_number_of_records()
        task_context.reset_total_number_of_async_requests()
        task_context.add_log("started")
        start_time = datetime.now()

        task_context.add_source_attributes()
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

        self.save_and_backfill_to_destinations(task_context)

        task_context.add_curr_timestamp(curr_timestamp)
        task_context.add_log("completed")
        end_time = datetime.now()
        latency_metric = (end_time - start_time).total_seconds()
        MetricsAggregator.update_task_stats(BaseExtractContext.TASK_TYPE, TaskStats.TO_FILE, TaskStats.LATENCY_COUNT,
                                            task_context.project_id, task_context.type_alias, latency_metric)
        MetricsAggregator.update_job_stats(task_context.project_id, task_context.customer_account_id,
                                           task_context.type_alias, "success", "")

    # Can be made as strategy later, whether to add or not.
    def save_and_backfill_to_destinations(self, task_context):
        for curr_timestamp in task_context.get_next_timestamps():
            self.save_to_destinations(task_context, curr_timestamp)

    # Later can add decorators for tracking.
    @staticmethod
    def save_to_destinations(task_context, curr_timestamp):
        task_context.add_curr_timestamp(curr_timestamp)
        task_context.add_destination_attributes()
        task_context.write_records()
        return
