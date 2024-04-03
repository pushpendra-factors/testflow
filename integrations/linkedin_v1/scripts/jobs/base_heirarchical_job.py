
import signal
from util.util import Util as U
from constants.constants import *
from util.data_transformation import DataTransformation
from cache.campaign_group_info import CampaignGroupInfo
from cache.campaign_info import CampaignInfo
from cache.creative_info import CreativeInfo
from cache.member_company_info import MemberCompany
from metrics_aggregator.metrics_aggregator import MetricsAggregator
from util.linkedin_api_service import LinkedinApiService
from data_service.data_service import DataService

class BaseHeirarchicalJob:
    meta_doc_type = ''
    insights_doc_type = ''
    meta_endpoint = ''
    pivot_insights = ''
    linkedin_setting = None
    sync_info_with_type = {}
    input_start_timestamp, input_end_timestamp = None, None
    timerange_for_meta = []
    timerange_for_insights = []
    campaign_group_cache = None
    campaign_cache = None
    creative_cache = None
    metrics_aggregator_obj = None
    data_service_obj = None
    linkedin_api_service_obj = None

    def __init__(self, meta_doc_type, insights_doc_type, meta_endpoint, pivot_insights,
                linkedin_setting, sync_info_with_type, input_start_timestamp, input_end_timestamp):
        self.meta_doc_type = meta_doc_type
        self.insights_doc_type = insights_doc_type
        self.meta_endpoint = meta_endpoint
        self.pivot_insights = pivot_insights
        self.linkedin_setting = linkedin_setting
        self.sync_info_with_type = sync_info_with_type
        self.input_start_timestamp = input_start_timestamp
        self.input_end_timestamp = input_end_timestamp
        self.timerange_for_meta = U.get_timestamp_range(self.linkedin_setting, self.meta_doc_type, 
                                        self.sync_info_with_type, self.input_start_timestamp, self.input_end_timestamp)
        self.timerange_for_insights = U.get_timestamp_range(self.linkedin_setting, self.insights_doc_type,
                                        self.sync_info_with_type, self.input_start_timestamp, self.input_end_timestamp)
        self.campaign_group_cache = CampaignGroupInfo.get_instance()
        self.campaign_cache = CampaignInfo.get_instance()
        self.creative_cache = CreativeInfo.get_instance()
        self.metrics_aggregator_obj = MetricsAggregator.get_instance()
        self.data_service_obj = DataService.get_instance()
        self.linkedin_api_service_obj =  LinkedinApiService.get_instance()

    def handle(signum, frame):
        raise Exception("Function timeout after 20 mins")
    
    def execute(self):
        metadata = self.execute_metadata_fetch_and_transform()
        for timestamp in self.timerange_for_meta:
            # timeout this function after 20 mins
            signal.signal(signal.SIGALRM, self.handle)
            signal.alarm(1200)
            # 
            self.data_service_obj.insert_metadata(self.meta_doc_type, self.linkedin_setting.project_id, 
                                            self.linkedin_setting.ad_account, metadata, timestamp)
        
        for timestamp in self.timerange_for_insights:
            # timeout this function after 20 mins
            signal.signal(signal.SIGALRM, self.handle)
            signal.alarm(1200)
            # 
            insights = self.excute_insights_fetch_and_transform(timestamp) 
            
            self.data_service_obj.insert_insights(self.insights_doc_type, self.linkedin_setting.project_id, 
                                                self.linkedin_setting.ad_account, insights, timestamp)

    
    def execute_metadata_fetch_and_transform(self):
        metadata = self.linkedin_api_service_obj.get_metadata(self.linkedin_setting, self.meta_endpoint, self.meta_doc_type)

        updated_meta = DataTransformation.transform_metadata_based_on_doc_type(metadata, self.meta_doc_type, 
                                                                    self.campaign_group_cache.campaign_group_info,
                                                                    self.campaign_cache.campaign_info_map,
                                                                    self.creative_cache.creative_info_map)
        return updated_meta
    
    def excute_insights_fetch_and_transform(self, timestamp):
        insights = self.linkedin_api_service_obj.get_insights(self.linkedin_setting, timestamp,
                                self.insights_doc_type, self.pivot_insights)

        transformed_insights = DataTransformation.update_insights_with_metadata(
                                            insights, self.insights_doc_type,
                                            self.campaign_group_cache.campaign_group_info,
                                            self.campaign_cache.campaign_info_map,
                                            self.creative_cache.creative_info_map)
        return transformed_insights

            


