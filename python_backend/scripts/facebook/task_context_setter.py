from lib.task.system.external import ExternalSystem
from lib.task.system.google_storage import GoogleStorage
from lib.task.system.local_storage import LocalStorage
from scripts.facebook import *
from scripts.facebook.tasks.context.extract.ad_performance_extract import AdPerformanceReportExtract
from scripts.facebook.tasks.context.extract.ad_set_info import AdSetInfoExtract
from scripts.facebook.tasks.context.extract.ad_info import AdInfoExtract
from scripts.facebook.tasks.context.extract.ad_set_performance_extract import AdSetPerformanceReportExtract
from scripts.facebook.tasks.context.extract.campaign_info_extract import CampaignInfoExtract
from scripts.facebook.tasks.context.extract.campaign_performance_extract import CampaignPerformanceReportExtract
from scripts.facebook.tasks.context.load.ad_performance_load import AdPerformanceLoad
from scripts.facebook.tasks.context.load.ad_set_info_load import AdSetInfoLoad
from scripts.facebook.tasks.context.load.ad_info_load import AdInfoLoad
from scripts.facebook.tasks.context.load.ad_set_performance_load import AdSetPerformanceLoad
from scripts.facebook.tasks.context.load.campaign_info_load import CampaignInfoLoad
from scripts.facebook.tasks.context.load.campaign_performance_load import CampaignPerformanceLoad
from scripts.facebook.tasks.extract.base_info_extract import BaseInfoExtract
from scripts.facebook.tasks.extract.base_report_extract import BaseReportExtract
from scripts.facebook.tasks.load.base_info_load import BaseInfoLoad
from scripts.facebook.tasks.load.base_report_load import BaseReportLoad


# we dont have workflow context to get tasks.


class TaskContextSetter:

    def __init__(self, last_sync_info, facebook_int_setting, env, dry, workflow_type, task_type,
                 facebook_data_service_path, input_from_timestamp, input_to_timestamp, project_min_timestamp):
        self.last_sync_info = last_sync_info
        self.facebook_int_setting = facebook_int_setting
        self.env = env
        self.dry = dry
        self.workflow_type = workflow_type
        self.task_type = task_type
        self.facebook_data_service_path = facebook_data_service_path
        self.input_from_timestamp = input_from_timestamp
        self.input_to_timestamp = input_to_timestamp
        self.project_min_timestamp = project_min_timestamp

    def get_task(self):
        if self.task_type == EXTRACT:
            if self.is_info_type():
                return BaseInfoExtract.get_instance()
            else:
                return BaseReportExtract.get_instance()
        else:
            if self.is_info_type():
                return BaseInfoLoad.get_instance()
            else:
                return BaseReportLoad.get_instance()

    def get_task_context(self):
        if self.task_type == EXTRACT:
            task_context = self.get_extract_task_context()
        else:
            task_context = self.get_load_task_context()

        task_context.add_last_sync_info(self.last_sync_info)
        task_context.add_facebook_settings(self.facebook_int_setting)
        task_context.add_source(self.get_source())
        task_context.add_destinations(self.get_destinations())
        task_context.add_env(self.env)
        task_context.add_dry(self.dry)
        task_context.add_input_from_timestamp(self.input_from_timestamp)
        task_context.add_input_to_timestamp(self.input_to_timestamp)
        task_context.add_facebook_data_service_path(self.facebook_data_service_path)
        task_context.add_project_min_timestamp(self.project_min_timestamp)
        return task_context

    def get_extract_task_context(self):
        task_name = self.last_sync_info.get("type_alias")
        if task_name == CAMPAIGN_INSIGHTS:
            return CampaignPerformanceReportExtract()
        elif task_name == AD_SET_INSIGHTS:
            return AdSetPerformanceReportExtract()
        elif task_name == AD_INSIGHTS:
            return AdPerformanceReportExtract()
        elif task_name == CAMPAIGN:
            return CampaignInfoExtract()
        elif task_name == AD_SET:
            return AdSetInfoExtract()
        else:
            return AdInfoExtract()

    def get_load_task_context(self):
        task_name = self.last_sync_info.get("type_alias")
        if task_name == CAMPAIGN_INSIGHTS:
            return CampaignPerformanceLoad()
        elif task_name == AD_SET_INSIGHTS:
            return AdSetPerformanceLoad()
        elif task_name == AD_INSIGHTS:
            return AdPerformanceLoad()
        elif task_name == CAMPAIGN:
            return CampaignInfoLoad()
        elif task_name == AD_SET:
            return AdSetInfoLoad()
        else:
            return AdInfoLoad()

    def is_report_type(self):
        return self.last_sync_info.get("type_alias") in [CAMPAIGN_INSIGHTS, AD_SET_INSIGHTS, AD_INSIGHTS]

    def is_info_type(self):
        return self.last_sync_info.get("type_alias") in [CAMPAIGN, AD_SET, AD]

    # Optimise for load in extract_and_load_workflow - localstorage
    def get_source(self):
        if self.task_type == EXTRACT:
            return ExternalSystem()
        else:
            if self.env in [DEVELOPMENT, TEST]:
                return [LocalStorage()]
            return GoogleStorage()

    # Optimise for extract in extract_and_Load_workflow - local and google storage.
    def get_destinations(self):
        if self.env in [DEVELOPMENT, TEST]:
            return [LocalStorage()]

        if self.task_type == EXTRACT:
            return [GoogleStorage()]
        elif self.task_type == LOAD:
            return [ExternalSystem()]
