import requests
import logging as log
import time
import string
import random
import requests
import logging as log
import time
import string
import random
from constants import *
from datetime import datetime
from _datetime import timedelta

class DataService:

    data_service_host = ''

    def __init__(self, options):
        self.data_service_host = options.data_service_host

    
    def get_linkedin_int_settings(self):
        uri = '/data_service/linkedin/project/settings'
        url = self.data_service_host + uri

        response = requests.get(url)
        if not response.ok:
            log.error('Failed to get linkedin integration settings from data services')
            return None, 'Failed to get linkedin integration settings from data services'
        return response.json(), ''

    def get_linkedin_int_settings_for_projects(self, project_ids):
        uri = '/data_service/linkedin/project/settings/projects'
        url = self.data_service_host + uri
        
        project_ids_arr = project_ids.split(',')
        payload = {
            'project_ids': project_ids_arr
        }

        response = requests.get(url, json=payload)
        if not response.ok:
            log.error('Failed to get linkedin integration settings for projects from data services')
            return (None, 
                    'Failed to get linkedin integration settings for projects from data services')
        return response.json(), ''

    
    def get_last_sync_info(self, linkedin_setting, start_timestamp=None, end_timestamp=None):
        uri = '/data_service/linkedin/documents/last_sync_info'
        url = self.data_service_host + uri

        payload = {
            PROJECT_ID: linkedin_setting.project_id,
            'customer_ad_account_id': linkedin_setting.ad_account
        }
        response = requests.get(url,json=payload)
        if not response.ok:
            log.error('Failed to get linkedin last sync info from data services')
            return [], 'Failed to get linkedin last sync info from data services'
        all_info = response.json()
        sync_info_with_type = {}
        for info in all_info:
            date = datetime.strptime(str(info['last_timestamp']), '%Y%m%d')
            if start_timestamp != None:
                date = datetime.strptime(str(start_timestamp), '%Y%m%d')
            sync_info_with_type[info['type_alias']]= date.strftime('%Y-%m-%d')
            if info['type_alias'] == 'member_company_insights':
                sync_info_with_type['last_backfill_timestamp'] = info['last_backfill_timestamp'] 
        timestamp_exceed = DataService.is_custom_timerange_exceeding_lookback(start_timestamp, end_timestamp)
        if timestamp_exceed:
            return [], "Given custom timerange exceeds max lookback" 
        return sync_info_with_type, ''


    
    def add_linkedin_documents(self, project_id, ad_account_id, 
                                doc_type, obj_id, value, 
                                timestamp, is_backfill=False):
        uri = '/data_service/linkedin/documents/add'
        url = self.data_service_host + uri

        payload = {
            PROJECT_ID: int(project_id),
            'customer_ad_account_id': ad_account_id,
            'type_alias': doc_type,
            'id': obj_id,
            'value': value,
            'timestamp': timestamp,
            'is_backfilled': is_backfill
        }

        response = requests.post(url, json=payload)
        if not response.ok:
            log.error(
            'Failed to add response %s to linkedin warehouse for project %s. StatusCode:  %d, %s',
                doc_type, project_id, response.status_code, response.text)
        
        return response

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
                                                    timestamp, is_backfill=False):
        #filling empty object when no data
        if len(docs) == 0:
            obj_id, empty_object = self.get_empty_object_with_req_ids()
            return self.add_linkedin_documents(project_id, customer_acc_id, doc_type,
                                                obj_id, empty_object, int(timestamp))
        response = {}
        for i in range(0, len(docs), BATCH_SIZE):
            batch = docs[i:i+BATCH_SIZE]
            response = self.add_multiple_linkedin_documents(project_id, customer_acc_id,
                                        doc_type,batch, timestamp, is_backfill)
            if not response.ok:
                return response

        return response

    
    def add_multiple_linkedin_documents(self, project_id, ad_account_id, 
                                                doc_type, docs, 
                                                timestamp, is_backfill=False):
        uri = '/data_service/linkedin/documents/add_multiple'
        url = self.data_service_host + uri

        batch_of_payloads = [self.get_payload_for_linkedin(project_id,
                                ad_account_id, doc_type, doc, 
                                timestamp, is_backfill) for doc in docs]

        retries = 0
        response = {}
        while retries < 3:
            response = requests.post(url, json=batch_of_payloads)
            if not response.ok:
                log.error(
                "Linkedin etl - Failed to add response %s to adwords warehouse for retry: %d, %s, %d",
                    doc_type, response.status_code, response.text, retries)
                time.sleep(2)
            else:
                return response
            retries += 1
        log.error("Linkedin etl - Failed to add response to adwords - Missing data.")
        return response
    
    def get_payload_for_linkedin(self, project_id, ad_account_id, doc_type, value, 
                                                timestamp, is_backfill=False):
        return {
            PROJECT_ID: int(project_id),
            'customer_ad_account_id': ad_account_id,
            'type_alias': doc_type,
            'id': str(value['id']),
            'value': value,
            'timestamp': int(timestamp),
            'is_backfilled': is_backfill
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
        response = requests.delete(url, json=payload)
        if not response.ok:
            log.error(
            'Failed to delete documents for project %s. Doc_type: %s, timestamp: %d, StatusCode:  %d',
                project_id, doc_type, int(timestamp), response.status_code)
        
        return response
    
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