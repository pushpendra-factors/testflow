import signal
from util.util import Util as U
from constants.constants import *
from collections import OrderedDict
from util.data_transformation import DataTransformation
from cache.campaign_group_info import CampaignGroupInfo
from cache.campaign_info import CampaignInfo
from metrics_aggregator.metrics_aggregator import MetricsAggregator
from data_service.data_service import DataService
from util.linkedin_api_service import LinkedinApiService
from cache.member_company_info import MemberCompany
from google_storage.google_storage import GoogleStorage

class WeeklyMemberCompanyJob:
    sync_status = 0
    timestamp_range = []
    linkedin_setting =  None

    def __init__(self, linkedin_setting, timestamp_range, sync_status):
        self.linkedin_setting = linkedin_setting
        self.timestamp_range = timestamp_range
        self.sync_status = sync_status
        self.campaign_group_cache = CampaignGroupInfo.get_instance()
        self.campaign_cache = CampaignInfo.get_instance()
        self.metrics_aggregator_obj = MetricsAggregator.get_instance()
        self.data_service_obj = DataService.get_instance()
        self.linkedin_api_service_obj =  LinkedinApiService.get_instance()
        self.member_company_cache = MemberCompany.get_instance()

    
    def handle(signum, frame):
        raise Exception("Function timeout after 20 mins")
    
    def execute(self):
        # timeout this function after 20 mins
        signal.signal(signal.SIGALRM, self.handle)
        signal.alarm(1200)
        # 
        distributed_records_map_with_timestamp = OrderedDict()
        start_timestamp = self.timestamp_range[0]
        end_timestamp = self.timestamp_range[len(self.timestamp_range)-1]
        
        company_insights = self.linkedin_api_service_obj.extract_company_insights_for_all_campaign_groups(self.linkedin_setting, 
                                                                    start_timestamp, end_timestamp,
                                                                    self.campaign_group_cache.get_campaign_group_ids())
        
        non_present_ids = U.get_non_present_ids(company_insights, self.member_company_cache.get_member_company_ids())

        self.member_company_cache.fetch_and_update_non_present_org_data_to_cache(
                                                    self.linkedin_setting.access_token, non_present_ids,
                                                    self.metrics_aggregator_obj)
        
        enriched_company_insights = DataTransformation.enrich_dependencies_to_company_insights(company_insights, 
                                                                self.campaign_group_cache.campaign_group_info, 
                                                                self.member_company_cache.member_company_map)
        # split each row evenly in 7 or len_timerange parts and maps it to the timestamp
        distributed_records_map_with_timestamp = DataTransformation.distribute_data_across_timerange(
                                                    enriched_company_insights, self.timestamp_range)
        
        # here we are directly deleting and inserting the data. Existing row check is present in event creation job
        for timestamp, records in distributed_records_map_with_timestamp.items():
            # Events job should not be run at the same time as this job
            # because deletion and reading could happen at the same time
            self.data_service_obj.delete_linkedin_documents_for_doc_type_and_timestamp(
                                                        self.linkedin_setting.project_id,
                                                        self.linkedin_setting.ad_account, MEMBER_COMPANY_INSIGHTS,
                                                        timestamp)
            self.data_service_obj.insert_insights(MEMBER_COMPANY_INSIGHTS, self.linkedin_setting.project_id, 
                                        self.linkedin_setting.ad_account, records, timestamp, self.sync_status)
    
        self.metrics_aggregator_obj.update_stats(self.linkedin_setting.project_id, self.linkedin_setting.ad_account,
                                        MEMBER_COMPANY_INSIGHTS, 0)
    
    def execute_v1(self):
        # timeout this function after 20 mins
        signal.signal(signal.SIGALRM, self.handle)
        signal.alarm(1200)
        # 
        distributed_records_map_with_timestamp = OrderedDict()
        start_timestamp = self.timestamp_range[0]
        end_timestamp = self.timestamp_range[len(self.timestamp_range)-1]
        
        company_insights = self.linkedin_api_service_obj.extract_company_insights_for_all_campaigns(self.linkedin_setting, 
                                                                    start_timestamp, end_timestamp,
                                                                    self.campaign_cache.get_campaign_ids())
        # writing to google cloud
        GoogleStorage.get_instance().write(str(company_insights), self.metrics_aggregator_obj.job_type, DATA_STATE_RAW, 
                                           start_timestamp, self.linkedin_setting.project_id, 
                                           self.linkedin_setting.ad_account, MEMBER_COMPANY_INSIGHTS)
        non_present_ids = U.get_non_present_ids(company_insights, self.member_company_cache.get_member_company_ids())

        self.member_company_cache.fetch_and_update_non_present_org_data_to_cache(
                                                    self.linkedin_setting.access_token, non_present_ids,
                                                    self.metrics_aggregator_obj)
        
        enriched_company_insights = DataTransformation.enrich_dependencies_to_company_insights_v1(company_insights, 
                                                                self.member_company_cache.member_company_map,
                                                                self.campaign_group_cache.campaign_group_info, 
                                                                self.campaign_cache.campaign_info_map)
        
        # split each row evenly in 7 or len_timerange parts and maps it to the timestamp
        distributed_records_map_with_timestamp = DataTransformation.distribute_data_across_timerange(
                                                    enriched_company_insights, self.timestamp_range)
        
        # here we are directly deleting and inserting the data. Existing row check is present in event creation job
        for timestamp, records in distributed_records_map_with_timestamp.items():
            # Events job should not be run at the same time as this job
            # because deletion and reading could happen at the same time
            # writing to google cloud
            GoogleStorage.get_instance().write(str(records), self.metrics_aggregator_obj.job_type, DATA_STATE_TRANSFORMED, 
                                           timestamp, self.linkedin_setting.project_id, 
                                           self.linkedin_setting.ad_account, MEMBER_COMPANY_INSIGHTS)
            self.data_service_obj.delete_linkedin_documents_for_doc_type_and_timestamp(
                                                        self.linkedin_setting.project_id,
                                                        self.linkedin_setting.ad_account, MEMBER_COMPANY_INSIGHTS,
                                                        timestamp)
            self.data_service_obj.insert_insights(MEMBER_COMPANY_INSIGHTS, self.linkedin_setting.project_id, 
                                        self.linkedin_setting.ad_account, records, timestamp, self.sync_status)
    
        self.metrics_aggregator_obj.update_stats(self.linkedin_setting.project_id, self.linkedin_setting.ad_account,
                                        MEMBER_COMPANY_INSIGHTS, 0)
        