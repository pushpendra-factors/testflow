import requests
import logging as log
import time
import string
import random
import logging as log
import time
import string
import random
from constants.constants import *
from datetime import datetime
from _datetime import timedelta
from util.util import Util as U
from custom_exception.custom_exception import CustomException

class DataService:

    data_service_host = ''
    is_dry_run = False
    __instance = None

    @staticmethod
    def get_instance(url='', is_dry_run=False):
        if DataService.__instance == None:
            DataService(url, is_dry_run)
        return DataService.__instance

    def __init__(self, url='', is_dry_run=False):
        self.data_service_host = url
        self.is_dry_run = is_dry_run
        DataService.__instance = self
    
    def get_linkedin_int_settings(self):
        uri = '/data_service/linkedin/project/settings'
        url = self.data_service_host + uri

        response, errMsg = U.get_request_with_retries(url)
        if errMsg != '':
            return None, errMsg

        return response.json(), ''

    def get_linkedin_int_settings_for_projects(self, project_ids):
        uri = '/data_service/linkedin/project/settings/projects'
        url = self.data_service_host + uri
        
        project_ids_arr = project_ids.split(',')
        payload = {
            'project_ids': project_ids_arr
        }

        # response = requests.get(url, json=payload)
        response, errMsg = U.get_request_with_retries(url, payload)
        if response is None or errMsg != '':
            log.error('Failed to get linkedin integration settings for projects from data services')
            return None, errMsg
        return response.json(), ''

    
    def get_last_sync_info(self, linkedin_setting, start_timestamp=None, input_end_timestamp=None):
        uri = '/data_service/linkedin/documents/ads/last_sync_info/V1'
        url = self.data_service_host + uri

        payload = {
            PROJECT_ID: linkedin_setting.project_id,
            'customer_ad_account_id': linkedin_setting.ad_account
        }

        # response = requests.get(url,json=payload)
        response, errMsg = U.get_request_with_retries(url, payload)
        if response is None or errMsg != '':
            log.error(errMsg)
            raise CustomException(errMsg, 0, 0)

        all_info = response.json()
        sync_info_with_type = {}
        for info in all_info:
            date = datetime.strptime(str(info['last_timestamp']), '%Y%m%d')
            sync_info_with_type[info['type_alias']]= date.strftime('%Y-%m-%d')
        timestamp_exceed = DataService.is_custom_timerange_exceeding_lookback(start_timestamp, input_end_timestamp)
        if timestamp_exceed:
            err_string = "Given custom timerange exceeds max lookback"
            raise CustomException(err_string, 0, 0)
        return sync_info_with_type
    
    def get_last_sync_info_for_company_data(self, linkedin_setting, start_timestamp=None, input_end_timestamp=None):
        uri = '/data_service/linkedin/documents/company/last_sync_info/V1'
        url = self.data_service_host + uri

        payload = {
            PROJECT_ID: linkedin_setting.project_id,
            'customer_ad_account_id': linkedin_setting.ad_account
        }
        # response = requests.get(url,json=payload)
        response, errMsg = U.get_request_with_retries(url, payload)
        if response is None or errMsg != '':
            log.error(errMsg)
            raise CustomException(errMsg, 0, 0)

        all_info = response.json()
        sync_info_with_type = {}
        for info in all_info:
            key = info['type_alias']
            if info['sync_type'] != 0:
                key = info['type_alias'] + ':' + str(info['sync_type'])
            if info['last_timestamp'] == 0:
                sync_info_with_type[key] = 0
            else:
                date = datetime.strptime(str(info['last_timestamp']), '%Y%m%d')
                sync_info_with_type[key]= date.strftime('%Y-%m-%d')
        timestamp_exceed = DataService.is_custom_timerange_exceeding_lookback(start_timestamp, input_end_timestamp)
        if timestamp_exceed:
            err_string = "Given custom timerange exceeds max lookback" 
            raise CustomException(err_string, 0, 0)
        return sync_info_with_type


    
    def add_linkedin_documents(self, project_id, ad_account_id, 
                                doc_type, obj_id, value, 
                                timestamp, sync_status=0):
        uri = '/data_service/linkedin/documents/add'
        url = self.data_service_host + uri

        payload = {
            PROJECT_ID: int(project_id),
            'customer_ad_account_id': ad_account_id,
            'type_alias': doc_type,
            'id': obj_id,
            'value': value,
            'timestamp': timestamp,
            'sync_status': sync_status
        }

        if self.is_dry_run:
            return
        # response = requests.post(url, json=payload)
        response, errMsg = U.post_request_with_retries(url, payload)
        if response is None or errMsg != '':
            log.warning(errMsg)
        return


    def get_empty_object_with_req_ids(self):
        obj_id = ''.join(random.choices(string.ascii_lowercase +
                             string.digits, k=8))
        cg_id = ''.join(random.choices(string.ascii_lowercase +
                             string.digits, k=8))
        c_id = ''.join(random.choices(string.ascii_lowercase +
                             string.digits, k=8))
        cr_id = ''.join(random.choices(string.ascii_lowercase +
                             string.digits, k=8))

        obj = {'campaign_group_id': cg_id, 'campaign_id': c_id, 'creative_id': cr_id}

        return obj_id, obj

    def add_all_linkedin_documents(self, project_id, customer_acc_id, doc_type, docs, 
                                                    timestamp, sync_status=0):
        #filling empty object when no data
        if len(docs) == 0:
            obj_id, empty_object = self.get_empty_object_with_req_ids()
            self.add_linkedin_documents(project_id, customer_acc_id, doc_type,
                                                obj_id, empty_object, int(timestamp))
        response = {}
        if self.is_dry_run:
            return
        for i in range(0, len(docs), BATCH_SIZE):
            batch = docs[i:i+BATCH_SIZE]
            response, err_msg = self.add_multiple_linkedin_documents(project_id, customer_acc_id,
                                        doc_type,batch, timestamp, sync_status)
            if not response.ok:
                raise CustomException(err_msg, 0, doc_type)

    
    def add_multiple_linkedin_documents(self, project_id, ad_account_id, 
                                                doc_type, docs, 
                                                timestamp, sync_status=0):
        uri = '/data_service/linkedin/documents/add_multiple'
        url = self.data_service_host + uri

        batch_of_payloads = [self.get_payload_for_linkedin(project_id,
                                ad_account_id, doc_type, doc, 
                                timestamp, sync_status) for doc in docs]


        # response = requests.post(url, json=batch_of_payloads)
        response, errMsg = U.post_request_with_retries(url, batch_of_payloads)
        if response is None or errMsg != '':
            return None, errMsg
        
        return response, ''
    
    def get_payload_for_linkedin(self, project_id, ad_account_id, doc_type, value, 
                                                timestamp, sync_status=0):
        return {
            PROJECT_ID: int(project_id),
            'customer_ad_account_id': ad_account_id,
            'type_alias': doc_type,
            'id': str(value['id']),
            'value': value,
            'timestamp': int(timestamp),
            'sync_status': sync_status
        }
    
    def update_access_token(self, project_id, access_token):
        uri = '/data_service/linkedin/access_token'
        url = self.data_service_host + uri
        
        payload = {
            PROJECT_ID: int(project_id),
            'access_token': access_token
        }

        response = requests.put(url, json=payload)
        if not response.ok:
            log.error(
            'Failed to update access token for project %s. StatusCode:  %d', 
                project_id, response.status_code)
        
        return response

    def delete_linkedin_documents_for_doc_type_and_timestamp(self, 
                                            project_id, ad_account_id, 
                                            doc_type, timestamp):
        uri = '/data_service/linkedin/documents'
        url = self.data_service_host + uri
        payload = {
            PROJECT_ID: int(project_id),
            'customer_ad_account_id': ad_account_id,
            'type_alias': doc_type,
            'timestamp': int(timestamp)
        }
        if self.is_dry_run:
            return
        response = requests.delete(url, json=payload)
        if not response.ok:
            err_string = 'Failed to delete documents for project {}. Doc_type: {}, timestamp: {}, StatusCode:  {}'.format(
                project_id, doc_type, int(timestamp), response.status_code)
            log.error(err_string)
            raise CustomException(err_string, 0, doc_type)
        
    
    def get_campaign_group_data_for_given_timerange(self, project_id, ad_account_id, 
                                                    start_timestamp, input_end_timestamp):
        uri = '/data_service/linkedin/documents/campaign_group_info'
        url = self.data_service_host + uri
        payload = {
            PROJECT_ID: int(project_id),
            'customer_ad_account_id': ad_account_id,
            'start_timestamp': start_timestamp,
            'end_timestamp': input_end_timestamp
        }
        # response = requests.get(url, json=payload)
        response, errMsg = U.get_request_with_retries(url, payload)
        if response is None or errMsg != '':
            log.error(errMsg)
            raise CustomException(errMsg, 0, MEMBER_COMPANY_INSIGHTS)
        
        campaign_group_info = response.json()
        campaign_group_id_to_info_map = U.build_map_of_campaign_group_info(campaign_group_info)
        
        return campaign_group_id_to_info_map
    
    def get_campaign_data_for_given_timerange(self, project_id, ad_account_id, 
                                                    start_timestamp, input_end_timestamp):
        uri = '/data_service/linkedin/documents/campaign_info'
        url = self.data_service_host + uri
        payload = {
            PROJECT_ID: int(project_id),
            'customer_ad_account_id': ad_account_id,
            'start_timestamp': start_timestamp,
            'end_timestamp': input_end_timestamp
        }
        response = requests.get(url, json=payload)
        if not response.ok:
            err_msg = (
            'Failed to get campaign group info from db for project %s. ad account %s, StatusCode:  %d',
                project_id, ad_account_id, response.status_code)
            log.error(err_msg)
            raise CustomException(err_msg, 0, MEMBER_COMPANY_INSIGHTS)
        
        campaign_info = response.json()
        campaign_id_to_info_map = U.build_map_of_campaign_info(campaign_info)
        
        return campaign_id_to_info_map
    
    def validate_company_data_pull(self, project_id, ad_account_id, 
                                start_timestamp, input_end_timestamp, sync_status):
        uri = '/data_service/linkedin/documents/validation'
        url = self.data_service_host + uri
        payload = {
            PROJECT_ID: int(project_id),
            'customer_ad_account_id': ad_account_id,
            'start_timestamp': start_timestamp,
            'end_timestamp': input_end_timestamp,
            'sync_status': sync_status
        }
        # response = requests.get(url, json=payload)
        response, errMsg = U.get_request_with_retries(url, payload)
        if response is None or errMsg != '':
            return False
        is_valid = response.json()
        return is_valid
        
    
    @staticmethod
    def is_custom_timerange_exceeding_lookback(input_from, input_to):
        #if no custom range given directly return false
        if input_from is None and input_to is None:
            return False
        
        date_start, date_end = "", ""
        if input_from != None:
            date_start = datetime.strptime(str(input_from), '%Y%m%d').date()
        else:
            date_start = (datetime.now() - timedelta(days=MAX_LOOKBACK)).date()

        if input_to != None:
            date_end = datetime.strptime(str(input_to), '%Y%m%d').date()
        else:
            date_end = (datetime.now() - timedelta(days=1)).date()

        num_of_days = (date_end-date_start).days + 1
        if num_of_days > MAX_LOOKBACK:
            return True
        
        return False
    
    def insert_metadata(self, doc_type, project_id, ad_account, response, timestamp):
        log.warning(INSERTION_LOG.format(doc_type, 'metadata', timestamp))
        self.add_all_linkedin_documents(project_id,
                                ad_account, doc_type, response, timestamp)


    
    def insert_insights(self, doc_type, project_id, ad_account, response, timestamp, sync_status=0):
        log.warning(INSERTION_LOG.format(doc_type, 'insights', timestamp))
        if len(response) > 0:
            self.add_all_linkedin_documents(project_id,
                                        ad_account, doc_type, response, timestamp, sync_status)
        log.warning(INSERTION_END_LOG.format(doc_type, 'insights', timestamp))

