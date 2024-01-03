from data_service.data_service import DataService
from cache.campaign_group_info import CampaignGroupInfo
from cache.member_company_info import MemberCompany
from cache.campaign_info import CampaignInfo
from cache.creative_info import CreativeInfo
from metrics_aggregator.metrics_aggregator import MetricsAggregator
from util.linkedin_api_service import LinkedinApiService
global data_service_obj
global campaign_group_cache
global campaign_cache
global creative_cache
global member_company_cache
global metrics_aggregator_obj
global client_id
global client_secret
campaign_group_cache = CampaignGroupInfo()
campaign_cache = CampaignInfo()
creative_cache = CreativeInfo()
member_company_cache = MemberCompany()
metrics_aggregator_obj = MetricsAggregator()
data_service_obj = DataService()
linkedin_api_service = LinkedinApiService(metrics_aggregator_obj)
client_id = ''
client_secret = ''

