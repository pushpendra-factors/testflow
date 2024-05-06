import logging as log
import traceback
import signal
from util.util import Util as U
from constants.constants import *
from util.data_transformation import DataTransformation
from cache.campaign_group_info import CampaignGroupInfo
from cache.campaign_info import CampaignInfo
from cache.member_company_info import MemberCompany
from metrics_aggregator.metrics_aggregator import MetricsAggregator
from util.linkedin_api_service import LinkedinApiService
from data_service.data_service import DataService

class MemberCompanyJob:
    linkedin_setting = None
    timestamp_range = []
    campaign_group_info = {}
    campaign_info = {}

    def __init__(self, linkedin_setting, timestamp_range):
        self.linkedin_setting = linkedin_setting
        self.campaign_group_info = CampaignGroupInfo.get_instance().campaign_group_info
        self.campaign_info = CampaignInfo.get_instance().campaign_info_map
        self.timestamp_range = timestamp_range
        self.metrics_aggregator_obj = MetricsAggregator.get_instance()
        self.data_service_obj = DataService.get_instance()
        self.linkedin_api_service_obj =  LinkedinApiService.get_instance()
        self.member_company_cache = MemberCompany.get_instance()

    def handle(signum, frame):
        raise Exception("Function timeout after 20 mins")

    def execute(self):
        try:
            for timestamp in self.timestamp_range:
                # timeout this function after 20 mins
                signal.signal(signal.SIGALRM, self.handle)
                signal.alarm(1200)
                # 
                company_insights = self.linkedin_api_service_obj.extract_company_insights_for_all_campaign_groups(
                                                                self.linkedin_setting, timestamp, timestamp,
                                                                self.campaign_group_info.keys())
                
                non_present_ids = U.get_non_present_ids(company_insights, self.member_company_cache.get_member_company_ids())

                self.member_company_cache.fetch_and_update_non_present_org_data_to_cache(
                                                        self.linkedin_setting.access_token, non_present_ids,
                                                        self.metrics_aggregator_obj)

                
                enriched_company_insights = DataTransformation.enrich_dependencies_to_company_insights(company_insights,
                                                                    self.campaign_group_info, 
                                                                    self.member_company_cache.member_company_map)

                self.data_service_obj.insert_insights(MEMBER_COMPANY_INSIGHTS, self.linkedin_setting.project_id, 
                                self.linkedin_setting.ad_account, enriched_company_insights, timestamp, SYNC_STATUS_T0)
        except Exception as e:
            self.metrics_aggregator_obj.update_stats(self.linkedin_setting.project_id, self.linkedin_setting.ad_account, 
                                                            e.doc_type, e.request_count, 'failed', e.message)
        self.metrics_aggregator_obj.update_stats(self.linkedin_setting.project_id, self.linkedin_setting.ad_account,
                                            MEMBER_COMPANY_INSIGHTS, 0)
        
    def execute_v1(self):
        for timestamp in self.timestamp_range:
            # timeout this function after 20 mins
            signal.signal(signal.SIGALRM, self.handle)
            signal.alarm(1200)
            # 
            company_insights = self.linkedin_api_service_obj.extract_company_insights_for_all_campaigns(
                                                            self.linkedin_setting, timestamp, timestamp,
                                                            self.campaign_info.keys())
            # rename variable
            non_present_ids = U.get_non_present_ids(company_insights, self.member_company_cache.get_member_company_ids())

            self.member_company_cache.fetch_and_update_non_present_org_data_to_cache(
                                                    self.linkedin_setting.access_token, non_present_ids,
                                                    self.metrics_aggregator_obj)

            
            enriched_company_insights = DataTransformation.enrich_dependencies_to_company_insights_v1(company_insights, 
                                                                self.member_company_cache.member_company_map,
                                                                self.campaign_group_info, self.campaign_info)

            self.data_service_obj.insert_insights(MEMBER_COMPANY_INSIGHTS, self.linkedin_setting.project_id, 
                            self.linkedin_setting.ad_account, enriched_company_insights, timestamp, SYNC_STATUS_T0)
        
        self.metrics_aggregator_obj.update_stats(self.linkedin_setting.project_id, self.linkedin_setting.ad_account,
                                            MEMBER_COMPANY_INSIGHTS, 0)
        
        