import logging as log
import traceback
from typing import List

import scripts
from lib.utils.facebook.metrics_aggregator import MetricsAggregator
from scripts.facebook import *
from scripts.facebook.task_context_setter import TaskContextSetter


# Later - precise system requirements and task phases of workflow. Eg - extract just to s3 Raw, transform
# and then multiple stage merges or transformation and loads.
# Later check how to handle permission denied.
# Imp migration - Following has to be handled properly if only few of the records/reports are to be handled.
# TODO IMP - Handle migrations properly.
# This might fail if metadata job itself is not run using run_job.
class JobSchedulerAndRunner:
    TASKS_WITH_INC_EXECUTION_ORDER = [AD, AD_SET, CAMPAIGN, CAMPAIGN_INSIGHTS, AD_SET_INSIGHTS, AD_INSIGHTS]

    @classmethod
    def sync(cls, facebook_int_setting: dict, sync_info_with_type: dict):
        facebook_config = scripts.facebook.CONFIG.FACEBOOK_APP
        ordered_last_sync_infos = JobSchedulerAndRunner.get_ordered_last_sync_infos(
            facebook_int_setting.get(PROJECT_ID),
            facebook_int_setting.get(FACEBOOK_AD_ACCOUNT),
            sync_info_with_type)
        project_min_timestamp = JobSchedulerAndRunner.get_project_min_timestamp(ordered_last_sync_infos)
        for task_name in WORKFLOW_TO_TASKS[facebook_config.type_of_run]:
            for last_sync_info in ordered_last_sync_infos:
                try:
                    task_context_setter = TaskContextSetter(last_sync_info, facebook_int_setting,
                                                            facebook_config.env, facebook_config.dry,
                                                            facebook_config.type_of_run, task_name,
                                                            facebook_config.get_data_service_path(),
                                                            facebook_config.last_timestamp,
                                                            facebook_config.to_timestamp,
                                                            project_min_timestamp)
                    task_context = task_context_setter.get_task_context()
                    task = task_context_setter.get_task()
                    task.execute(task_context)
                except Exception as e:
                    traceback.print_tb(e.__traceback__)
                    str_exception = str(e)
                    message = str_exception
                    log.warning("Failed with exception: %d %s %s %s", facebook_int_setting["project_id"],
                                facebook_int_setting["int_facebook_ad_account"], last_sync_info.get("type_alias"), message)
                    MetricsAggregator.update_job_stats(facebook_int_setting["project_id"],
                                                       facebook_int_setting["int_facebook_ad_account"],
                                                       last_sync_info.get("type_alias"), "failed", message)


    @staticmethod
    def get_project_min_timestamp(sync_infos):
        min_timestamp = 80001229
        for sync_info in sync_infos:
            if "insights" in sync_info[TYPE_ALIAS]:
                min_timestamp = min(sync_info[LAST_TIMESTAMP], min_timestamp)
        return min_timestamp

    # Manually defining order of tasks.
    @staticmethod
    def get_ordered_last_sync_infos(project_id, customer_account_id, last_sync_infos: dict):
        all_last_sync_infos: List = []
        for task_name in JobSchedulerAndRunner.TASKS_WITH_INC_EXECUTION_ORDER:
            if task_name in last_sync_infos:
                all_last_sync_infos.append(last_sync_infos[task_name])
            else:
                last_sync_info = {
                    PROJECT_ID: project_id,
                    CUSTOMER_ACCOUNT_ID: customer_account_id,
                    TYPE_ALIAS: task_name,
                    LAST_TIMESTAMP: 0
                }
                all_last_sync_infos.append(last_sync_info)
        return all_last_sync_infos
