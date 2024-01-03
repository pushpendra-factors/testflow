from global_objects.global_obj_creator import campaign_group_cache, campaign_cache, creative_cache, data_service_obj, linkedin_api_service
from util.util import Util as U
from constants.constants import *
from util.data_transformation import DataTransformation
from custom_exception.custom_exception import CustomException

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

    def execute(self):
        metadata = self.execute_metadata_fetch_and_transform()
        for timestamp in self.timerange_for_meta:
            data_service_obj.insert_metadata(self.meta_doc_type, self.linkedin_setting.project_id, 
                                            self.linkedin_setting.ad_account, metadata, timestamp)
        
        for timestamp in self.timerange_for_insights:
            insights = self.excute_insights_fetch_and_transform(timestamp) 
            
            data_service_obj.insert_insights(self.insights_doc_type, self.linkedin_setting.project_id, 
                                                self.linkedin_setting.ad_account, insights, timestamp)

    
    def execute_metadata_fetch_and_transform(self):
        metadata = linkedin_api_service.get_metadata(self.linkedin_setting, self.meta_endpoint, self.meta_doc_type)

        updated_meta = DataTransformation.transform_metadata_based_on_doc_type(metadata, self.meta_doc_type, 
                                                                    campaign_group_cache.campaign_group_info,
                                                                    campaign_cache.campaign_info_map,
                                                                    creative_cache.creative_info_map)
        return updated_meta
    
    def excute_insights_fetch_and_transform(self, timestamp):
        insights = linkedin_api_service.get_insights(self.linkedin_setting, timestamp,
                                self.insights_doc_type, self.pivot_insights)

        transformed_insights = DataTransformation.update_insights_with_metadata(
                                            insights, self.insights_doc_type,
                                            campaign_group_cache.campaign_group_info,
                                            campaign_cache.campaign_info_map,
                                            creative_cache.creative_info_map)
        return transformed_insights

            


