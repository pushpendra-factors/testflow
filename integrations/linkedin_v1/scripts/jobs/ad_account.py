
from global_objects.global_obj_creator import metrics_aggregator_obj, data_service_obj, linkedin_api_service
from util.linkedin_api_service import LinkedinApiService
from constants.constants import *
from _datetime import datetime
class AdAccountJob:
    linkedin_setting = None
    input_timestamp = None

    def __init__(self, linkedin_setting, input_timestamp) -> None:
        self.linkedin_setting, self.input_timestamp = linkedin_setting, input_timestamp
    
    def execute(self):
        try:
            metadata = linkedin_api_service.get_ad_account_data(self.linkedin_setting)

            timestamp = int(datetime.now().strftime('%Y%m%d'))
            if self.input_timestamp != None:
                timestamp = self.input_timestamp
            
            data_service_obj.add_linkedin_documents(
                            self.linkedin_setting.project_id, self.linkedin_setting.ad_account,
                            AD_ACCOUNT, str(metadata['id']),
                            metadata, timestamp)
        except Exception as e:
            metrics_aggregator_obj.update_stats(self.linkedin_setting.project_id, self.linkedin_setting.ad_account, 
                                                            e.doc_type, e.request_count, 'failed', e.message)