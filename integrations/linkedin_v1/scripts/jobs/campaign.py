from constants.constants import *
from .base_heirarchical_job import BaseHeirarchicalJob

class CampaignJob(BaseHeirarchicalJob):
    meta_doc_type = CAMPAIGNS
    insights_doc_type = CAMPAIGN_INSIGHTS
    meta_endpoint = URL_ENDPOINT_CAMPAIGN_META
    pivot_insights = PIVOT_CAMPAIGN

    def __init__(self, linkedin_setting, sync_info_with_type, input_start_timestamp, input_end_timestamp):
        super().__init__(self.meta_doc_type, self.insights_doc_type, self.meta_endpoint, self.pivot_insights,
                        linkedin_setting, sync_info_with_type, input_start_timestamp, input_end_timestamp)