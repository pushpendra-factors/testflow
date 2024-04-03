from util.util import Util as U
from util.linkedin_api_service import LinkedinApiService
class MemberCompany:
    member_company_map = {}
    __instance = None

    @staticmethod
    def get_instance(member_company={}):
        if MemberCompany.__instance == None:
            MemberCompany(member_company)
        return MemberCompany.__instance

    def __init__(self, member_company={}) -> None:
        self.member_company_map = member_company
        MemberCompany.__instance = self
    
    def get_member_company_data(self):
        return self.member_company_map
    
    def update_member_company_data(self, new_member_company_map={}):
        self.member_company_map = U.merge_2_dictionaries(
                self.member_company_map, new_member_company_map)
        
    def reset_member_company_data(self):
        self.member_company_map = {}
    
    def get_member_company_ids(self):
        return self.member_company_map.keys()
    
    def fetch_and_update_non_present_org_data_to_cache(self, access_token, non_present_ids, metrics_aggregator_obj):
        map_of_new_company_data = {}
        
        map_of_new_company_data = LinkedinApiService().get_company_data_from_linkedin_with_retries(
                                                                non_present_ids, access_token)
        self.update_member_company_data(map_of_new_company_data)
