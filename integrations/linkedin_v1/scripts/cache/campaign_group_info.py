from util.util import Util as U
from constants.constants import *
from data_service.data_service import DataService
class CampaignGroupInfo:
    campaign_group_info = {}
    __instance = None

    @staticmethod
    def get_instance(campaign_group_info={}):
        if CampaignGroupInfo.__instance == None:
            CampaignGroupInfo(campaign_group_info)
        return CampaignGroupInfo.__instance

    def __init__(self, campaign_group_info={}) -> None:
        self.campaign_group_info = campaign_group_info
        CampaignGroupInfo.__instance = self


    def get_campaign_group_data(self):
        return self.campaign_group_info
    
    def get_campaign_group_ids(self):
        return self.campaign_group_info.keys()
    
    def update_campaign_group_data(self, new_campaign_group_info={}):
        self.campaign_group_info = U.merge_2_dictionaries(
                self.campaign_group_info, new_campaign_group_info)
        
    def reset_campaign_group_data(self):
        self.campaign_group_info = {}
        
    def get_campaign_group_info_from_db(self, project_id, ad_account_id, 
                                                    start_timestamp, input_end_timestamp):
        data_service_obj = DataService.get_instance()

        self.campaign_group_info = data_service_obj.get_campaign_group_data_for_given_timerange(
                                                    project_id, ad_account_id, 
                                                    start_timestamp, input_end_timestamp)