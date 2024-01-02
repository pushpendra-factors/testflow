from util.util import Util as U
from constants.constants import *
from collections import OrderedDict
from util.data_transformation import DataTransformation
from global_objects.global_obj_creator import metrics_aggregator_obj, data_service_obj, campaign_group_cache, linkedin_api_service, member_company_cache
class WeeklyMemberCompanyJob:
    sync_status = 0
    timestamp_range = []
    linkedin_setting =  None
    def __init__(self, linkedin_setting, timestamp_range, sync_status):
        self.linkedin_setting = linkedin_setting
        self.timestamp_range = timestamp_range
        self.sync_status = sync_status
    
    def execute(self):
        distributed_records_map_with_timestamp = OrderedDict()
        start_timestamp = self.timestamp_range[0]
        end_timestamp = self.timestamp_range[len(self.timestamp_range)-1]
        
        company_insights = linkedin_api_service.extract_company_insights_for_all_campaigns(self.linkedin_setting, 
                                                                    start_timestamp, end_timestamp,
                                                                    campaign_group_cache.get_campaign_group_ids())
        non_present_ids = U.get_non_present_ids(company_insights, member_company_cache.get_member_company_ids())

        member_company_cache.fetch_and_update_non_present_org_data_to_cache(
                                                    self.linkedin_setting.access_token, non_present_ids,
                                                    metrics_aggregator_obj)
        
        enriched_company_insights = DataTransformation.enrich_dependencies_to_company_insights(company_insights, 
                                                                campaign_group_cache.campaign_group_info, 
                                                                member_company_cache.member_company_map)
        # split each row evenly in 7 or len_timerange parts and maps it to the timestamp
        distributed_records_map_with_timestamp = DataTransformation.distribute_data_across_timerange(
                                                    enriched_company_insights, self.timestamp_range)
        
        # here we are directly deleting and inserting the data. Existing row check is present in event creation job
        for timestamp, records in distributed_records_map_with_timestamp.items():
            # Events job should not be run at the same time as this job
            # because deletion and reading could happen at the same time
            data_service_obj.delete_linkedin_documents_for_doc_type_and_timestamp(
                                                        self.linkedin_setting.project_id,
                                                        self.linkedin_setting.ad_account, MEMBER_COMPANY_INSIGHTS,
                                                        timestamp)
            data_service_obj.insert_insights(MEMBER_COMPANY_INSIGHTS, self.linkedin_setting.project_id, 
                                        self.linkedin_setting.ad_account, records, timestamp, self.sync_status)
    
        metrics_aggregator_obj.update_stats(self.linkedin_setting.project_id, self.linkedin_setting.ad_account,
                                        MEMBER_COMPANY_INSIGHTS, 0)
        