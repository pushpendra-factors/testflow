from util.util import Util as U
from constants.constants import *
from util.data_transformation import DataTransformation
from global_objects.global_obj_creator import metrics_aggregator_obj, data_service_obj, linkedin_api_service, campaign_group_cache, member_company_cache
class MemberCompanyJob:
    linkedin_setting = None
    timestamp_range = []
    campaign_group_info = {}

    def __init__(self, linkedin_setting, campaign_group_info, timestamp_range):
        self.linkedin_setting = linkedin_setting
        self.campaign_group_info = campaign_group_info
        self.timestamp_range = timestamp_range

    def execute(self):
        try:
            for timestamp in self.timestamp_range:
                company_insights = linkedin_api_service.extract_company_insights_for_all_campaigns(
                                                                self.linkedin_setting, timestamp, timestamp,
                                                                campaign_group_cache.get_campaign_group_ids())
                
                non_present_ids = U.get_non_present_ids(company_insights, member_company_cache.get_member_company_ids())

                member_company_cache.fetch_and_update_non_present_org_data_to_cache(
                                                        self.linkedin_setting.access_token, non_present_ids,
                                                        metrics_aggregator_obj)

                
                enriched_company_insights = DataTransformation.enrich_dependencies_to_company_insights(company_insights,
                                                                    campaign_group_cache.campaign_group_info, 
                                                                    member_company_cache.member_company_map)

                data_service_obj.insert_insights(MEMBER_COMPANY_INSIGHTS, self.linkedin_setting.project_id, 
                                self.linkedin_setting.ad_account, enriched_company_insights, timestamp, SYNC_STATUS_T0)
        except Exception as e:
            metrics_aggregator_obj.update_stats(self.linkedin_setting.project_id, self.linkedin_setting.ad_account, 
                                                            e.doc_type, e.request_count, 'failed', e.message)
        metrics_aggregator_obj.update_stats(self.linkedin_setting.project_id, self.linkedin_setting.ad_account,
                                            MEMBER_COMPANY_INSIGHTS, 0)
        
        