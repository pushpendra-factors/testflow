import logging as log
import traceback
from util.util import Util as U
from constants.constants import *
from jobs.weekly_member_company import WeeklyMemberCompanyJob
from custom_exception.custom_exception import CustomException
from cache.campaign_group_info import CampaignGroupInfo
from cache.campaign_info import CampaignInfo
from metrics_aggregator.metrics_aggregator import MetricsAggregator
from data_service.data_service import DataService
class WeeklyMemberCompanyJobRunner:
    last_timestamp = ''
    weeks_for_buffer = 0
    sync_status = 0
    linkedin_setting, input_start_timestamp, input_end_timestamp = None, None, None
    
    def __init__(self, last_timestamp, weeks_for_buffer, sync_status, job_type, linkedin_setting, 
                                                    input_start_timestamp, input_end_timestamp):
        self.last_timestamp = last_timestamp
        self.weeks_for_buffer = weeks_for_buffer
        self.sync_status = sync_status
        self.linkedin_setting = linkedin_setting
        self.input_start_timestamp = input_start_timestamp
        self.input_end_timestamp = input_end_timestamp
        self.campaign_group_cache = CampaignGroupInfo.get_instance()
        self.campaign_cache = CampaignInfo.get_instance()
        self.metrics_aggregator_obj = MetricsAggregator.get_instance()
        self.data_service_obj = DataService.get_instance()
        self.metrics_aggregator_obj.job_type = job_type
    
    def execute(self):
        try:
            timestamp_range_chunks = U.get_timestamp_chunks_to_be_backfilled(self.weeks_for_buffer, self.last_timestamp, 
                                                                    self.input_start_timestamp, self.input_end_timestamp)
            valid_timestamp_range_chunks = U.exclude_timerange_inclusive_of_day3(timestamp_range_chunks)
            
            for timestamp_range in valid_timestamp_range_chunks:
                is_valid = self.data_service_obj.validate_company_data_pull(self.linkedin_setting.project_id, 
                                                        self.linkedin_setting.ad_account, timestamp_range[0], 
                                                        timestamp_range[len(timestamp_range)-1], self.sync_status)
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
                

                WeeklyMemberCompanyJob(self.linkedin_setting, timestamp_range, self.sync_status).execute()
                self.campaign_group_cache.reset_campaign_group_data()
        
        except Exception as e:
            traceback.print_tb(e.__traceback__)
            self.metrics_aggregator_obj.update_stats(self.linkedin_setting.project_id, self.linkedin_setting.ad_account, 
                                                MEMBER_COMPANY_INSIGHTS, e.request_count, 
                                                'failed', e.message)

    def execute_v1(self):
        try:
            timestamp_range_chunks = U.get_timestamp_chunks_to_be_backfilled(self.weeks_for_buffer, self.last_timestamp, 
                                                                    self.input_start_timestamp, self.input_end_timestamp)
            valid_timestamp_range_chunks = U.exclude_timerange_inclusive_of_day3(timestamp_range_chunks)
            
            for timestamp_range in valid_timestamp_range_chunks:
                is_valid = self.data_service_obj.validate_company_data_pull(self.linkedin_setting.project_id, 
                                                        self.linkedin_setting.ad_account, timestamp_range[0], 
                                                        timestamp_range[len(timestamp_range)-1], self.sync_status)
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
                

                WeeklyMemberCompanyJob(self.linkedin_setting, timestamp_range, self.sync_status).execute_v1()
                self.campaign_group_cache.reset_campaign_group_data()
                self.campaign_cache.reset_campaign_data()
        
        except Exception as e:
            traceback.print_tb(e.__traceback__)
            self.metrics_aggregator_obj.update_stats(self.linkedin_setting.project_id, self.linkedin_setting.ad_account, 
                                                MEMBER_COMPANY_INSIGHTS, e.request_count, 
                                                'failed', e.message)

