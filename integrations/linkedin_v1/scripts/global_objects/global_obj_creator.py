from data_service.data_service import DataService
from cache.campaign_group_info import CampaignGroupInfo
from cache.member_company_info import MemberCompany
from cache.campaign_info import CampaignInfo
from cache.creative_info import CreativeInfo
from metrics_aggregator.metrics_aggregator import MetricsAggregator
from util.linkedin_api_service import LinkedinApiService
from google_storage.google_storage import GoogleStorage

class GlobalObjects:
    campaign_group_cache = None
    campaign_cache = None
    creative_cache = None
    member_company_cache = None
    metrics_aggregator = None
    data_service = None
    linkedin_api_service = None
    google_storage = None

    def __init__(self, env, data_service_host):
        self.campaign_group_cache = CampaignGroupInfo.get_instance()
        self.campaign_cache = CampaignInfo.get_instance()
        self.creative_cache = CreativeInfo.get_instance()
        self.member_company_cache = MemberCompany.get_instance()
        self.metrics_aggregator = MetricsAggregator.get_instance()
        self.data_service = DataService.get_instance(data_service_host)
        self.linkedin_api_service = LinkedinApiService.get_instance()
        self.google_storage = GoogleStorage.get_instance(env)



