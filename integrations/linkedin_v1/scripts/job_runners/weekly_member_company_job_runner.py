import logging as log
import traceback
from util.util import Util as U
from constants.constants import *
from jobs.weekly_member_company import WeeklyMemberCompanyJob
from custom_exception.custom_exception import CustomException
from global_objects.global_obj_creator import metrics_aggregator_obj, data_service_obj, campaign_group_cache, linkedin_api_service, member_company_cache
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
        metrics_aggregator_obj.job_type = job_type
    
    def execute(self):
        try:
            timestamp_range_chunks = U.get_timestamp_chunks_to_be_backfilled(self.weeks_for_buffer, self.last_timestamp, 
                                                                    self.input_start_timestamp, self.input_end_timestamp)
            valid_timestamp_range_chunks = U.exclude_timerange_inclusive_of_day3(timestamp_range_chunks)
            
            for timestamp_range in valid_timestamp_range_chunks:
                is_valid = data_service_obj.validate_company_data_pull(self.linkedin_setting.project_id, 
                                                        self.linkedin_setting.ad_account, timestamp_range[0], 
                                                        timestamp_range[len(timestamp_range)-1], self.sync_status)
                if not is_valid:
                    raise CustomException("failed in validation of timeranges", 0, MEMBER_COMPANY_INSIGHTS)
                
                campaign_group_info = data_service_obj.get_campaign_group_data_for_given_timerange(
                                                        self.linkedin_setting.project_id, self.linkedin_setting.ad_account, 
                                                        timestamp_range[0], timestamp_range[len(timestamp_range)-1])
                if len(campaign_group_info) == 0:
                    err_string = "No campaign_data found for project {}, ad account {} for range {} to {}".format(
                                        self.linkedin_setting.project_id, self.linkedin_setting.ad_account, 
                                        timestamp_range[0], timestamp_range[len(timestamp_range)-1])
                    raise CustomException(err_string, 0, MEMBER_COMPANY_INSIGHTS)
                
                campaign_group_cache.campaign_group_info = campaign_group_info

                WeeklyMemberCompanyJob(self.linkedin_setting, timestamp_range, self.sync_status).execute()
                campaign_group_cache.reset_campaign_group_data()
        
        except Exception as e:
            traceback.print_tb(e.__traceback__)
            metrics_aggregator_obj.update_stats(self.linkedin_setting.project_id, self.linkedin_setting.ad_account, 
                                                MEMBER_COMPANY_INSIGHTS, e.request_count, 
                                                'failed', e.message)

