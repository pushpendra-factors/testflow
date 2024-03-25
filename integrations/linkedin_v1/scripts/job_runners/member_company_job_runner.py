import logging as log
import traceback
from util.util import Util as U
from constants.constants import *
from jobs.member_company import MemberCompanyJob
from custom_exception.custom_exception import CustomException
from cache.campaign_group_info import CampaignGroupInfo
from cache.campaign_info import CampaignInfo
from metrics_aggregator.metrics_aggregator import MetricsAggregator
from data_service.data_service import DataService

class MemberCompanyJobRunner:
    linkedin_setting = None
    last_timestamp = None
    input_start_timestamp, input_end_timestamp = None, None

    def __init__(self, linkedin_setting, last_timestamp, input_start_timestamp, input_end_timestamp):
        self.linkedin_setting = linkedin_setting
        self.last_timestamp = last_timestamp
        self.input_start_timestamp = input_start_timestamp
        self.input_end_timestamp = input_end_timestamp
        self.campaign_group_cache = CampaignGroupInfo.get_instance()
        self.campaign_cache = CampaignInfo.get_instance()
        self.metrics_aggregator_obj = MetricsAggregator.get_instance()
        self.data_service_obj = DataService.get_instance()
        self.metrics_aggregator_obj.job_type = "daily"

    def execute(self):
        try:
            timestamp_range = U.get_timestamp_range_for_company_insights(self.linkedin_setting, MEMBER_COMPANY_INSIGHTS, 
                                                                self.last_timestamp, self.input_start_timestamp, 
                                                                self.input_end_timestamp)
            if len(timestamp_range) == 0:
                return
            
            is_valid = self.data_service_obj.validate_company_data_pull(self.linkedin_setting.project_id, 
                                                        self.linkedin_setting.ad_account, timestamp_range[0], 
                                                        timestamp_range[len(timestamp_range)-1], 0)
            if not is_valid:
                raise CustomException("failed in validation of timeranges", 0, MEMBER_COMPANY_INSIGHTS)
            
            self.campaign_group_cache.get_campaign_group_info_from_db(
                                                self.linkedin_setting.project_id, self.linkedin_setting.ad_account, 
                                                timestamp_range[0], timestamp_range[len(timestamp_range)-1])
            
            if len(self.campaign_group_cache.campaign_group_info) == 0:
                err_string = "No campaign_data found for project {}, ad account {} for range {} to {}".format(
                                        self.linkedin_setting.project_id, self.linkedin_setting.ad_account, 
                                        timestamp_range[0], timestamp_range[len(timestamp_range)-1])
                raise CustomException(err_string, 0, MEMBER_COMPANY_INSIGHTS)
            
            MemberCompanyJob(self.linkedin_setting, timestamp_range).execute()
            self.campaign_group_cache.reset_campaign_group_data()
        
        except Exception as e:
            traceback.print_tb(e.__traceback__)
            self.metrics_aggregator_obj.update_stats(self.linkedin_setting.project_id, self.linkedin_setting.ad_account, 
                                                            e.doc_type, e.request_count, 'failed', e.message)
            
    def execute_v1(self):
        try:
            timestamp_range = U.get_timestamp_range_for_company_insights(self.linkedin_setting, MEMBER_COMPANY_INSIGHTS, 
                                                                self.last_timestamp, self.input_start_timestamp, 
                                                                self.input_end_timestamp)
            if len(timestamp_range) == 0:
                return
            
            is_valid = self.data_service_obj.validate_company_data_pull(self.linkedin_setting.project_id, 
                                                        self.linkedin_setting.ad_account, timestamp_range[0], 
                                                        timestamp_range[len(timestamp_range)-1], 0)
            if not is_valid:
                raise CustomException("failed in validation of timeranges", 0, MEMBER_COMPANY_INSIGHTS)
            
            self.campaign_group_cache.get_campaign_group_info_from_db(
                                                self.linkedin_setting.project_id, self.linkedin_setting.ad_account, 
                                                timestamp_range[0], timestamp_range[len(timestamp_range)-1])
            self.campaign_cache.get_campaign_info_from_db(
                                                self.linkedin_setting.project_id, self.linkedin_setting.ad_account, 
                                                timestamp_range[0], timestamp_range[len(timestamp_range)-1])
            
            if len(self.campaign_group_cache.campaign_group_info) == 0 or len(self.campaign_cache.campaign_info_map) == 0:
                err_string = "No campaign_data found for project {}, ad account {} for range {} to {}".format(
                                        self.linkedin_setting.project_id, self.linkedin_setting.ad_account, 
                                        timestamp_range[0], timestamp_range[len(timestamp_range)-1])
                raise CustomException(err_string, 0, MEMBER_COMPANY_INSIGHTS)
            
            MemberCompanyJob(self.linkedin_setting, timestamp_range).execute_v1()
            self.campaign_group_cache.reset_campaign_group_data()
            self.campaign_cache.reset_campaign_data()
        
        except Exception as e:
            traceback.print_tb(e.__traceback__)
            self.metrics_aggregator_obj.update_stats(self.linkedin_setting.project_id, self.linkedin_setting.ad_account, 
                                                            e.doc_type, e.request_count, 'failed', e.message)