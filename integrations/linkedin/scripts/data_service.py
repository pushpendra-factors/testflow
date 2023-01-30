from constants import *
from datetime import datetime
from _datetime import timedelta
import requests
import logging as log
import time

def get_linkedin_int_settings(options):
    uri = '/data_service/linkedin/project/settings'
    url = options.data_service_host + uri

    response = requests.get(url)
    if not response.ok:
        log.error('Failed to get linkedin integration settings from data services')
        return 
    return response.json()

def get_linkedin_int_settings_for_projects(options):
    project_ids = options.project_ids
    uri = '/data_service/linkedin/project/settings/projects'
    url = options.data_service_host + uri
    project_ids_arr = project_ids.split(',')
    payload = {
        'project_ids': project_ids_arr
    }

    response = requests.get(url, json=payload)
    if not response.ok:
        log.error('Failed to get linkedin integration settings for projects from data services')
        return 
    return response.json()

def get_last_sync_info(linkedin_int_setting, options):
    uri = '/data_service/linkedin/documents/last_sync_info'
    url = options.data_service_host + uri
    payload = {
        PROJECT_ID: linkedin_int_setting[PROJECT_ID],
        'customer_ad_account_id': linkedin_int_setting[LINKEDIN_AD_ACCOUNT]
    }
    response = requests.get(url,json=payload)
    if not response.ok:
        log.error('Failed to get linkedin last sync info from data services')
        return [], 'failed'
    all_info = response.json()
    sync_info_with_type = {}
    for info in all_info:
        date = datetime.strptime(str(info['last_timestamp']), '%Y%m%d')
        sync_info_with_type[info['type_alias']]= date.strftime('%Y-%m-%d')
    return sync_info_with_type, ''


def add_linkedin_documents(project_id, ad_account_id, doc_type, obj_id, value, timestamp, options):
    uri = '/data_service/linkedin/documents/add'
    url = options.data_service_host + uri

    payload = {
        PROJECT_ID: int(project_id),
        'customer_ad_account_id': ad_account_id,
        'type_alias': doc_type,
        'id': obj_id,
        'value': value,
        'timestamp': timestamp
    }

    response = requests.post(url, json=payload)
    if not response.ok:
        log.error('Failed to add response %s to linkedin warehouse for project %s. StatusCode:  %d, %s', 
            doc_type, project_id, response.status_code, response.text)
    
    return response

def add_multiple_linkedin_documents(project_id, ad_account_id, doc_type, docs, timestamp,options):
    uri = '/data_service/linkedin/documents/add'
    url = options.data_service_host + uri

    batch_of_payloads = [get_payload_for_linkedin(project_id, ad_account_id,
                                    doc_type, doc, timestamp) for doc in docs]

    retries = 0
    response = {}
    while retries < 3:
        response = requests.post(url, json=batch_of_payloads)
        if not response.ok:
            log.error("Linkedin etl - Failed to add response %s to adwords warehouse for retry: %d, %s, %d",
                    doc_type, response.status_code, response.text, retries)
            time.sleep(2)
        else:
            return response
        retries += 1
    log.error("Linkedin etl - Failed to add response to adwords - Missing data.")
    return response

def add_all_linkedin_documents(project_id, customer_acc_id, doc_type, docs, timestamp, options):
    response = {}
    for i in range(0, len(docs), BATCH_SIZE):
        batch = docs[i:i+BATCH_SIZE]
        response = add_multiple_linkedin_documents(project_id, customer_acc_id,
                                    doc_type,batch, timestamp, options)
        if not response.ok:
            return response

    return response

def add_multiple_linkedin_documents(project_id, ad_account_id, doc_type, docs, timestamp,options):
    uri = '/data_service/linkedin/documents/add_multiple'
    url = options.data_service_host + uri

    batch_of_payloads = [get_payload_for_linkedin(project_id, ad_account_id,
                                    doc_type, doc, timestamp) for doc in docs]

    retries = 0
    response = {}
    while retries < 3:
        response = requests.post(url, json=batch_of_payloads)
        if not response.ok:
            log.error("Linkedin etl - Failed to add response %s to adwords warehouse for retry: %d, %s, %d",
                    doc_type, response.status_code, response.text, retries)
            time.sleep(2)
        else:
            return response
        retries += 1
    log.error("Linkedin etl - Failed to add response to adwords - Missing data.")
    return response

def get_payload_for_linkedin(project_id, ad_account_id, doc_type, value, timestamp):
    return {
        PROJECT_ID: int(project_id),
        'customer_ad_account_id': ad_account_id,
        'type_alias': doc_type,
        'id': str(value['id']),
        'value': value,
        'timestamp': int(timestamp)
    }

def update_access_token(project_id, access_token, options):
    uri = '/data_service/linkedin/access_token'
    url = options.data_service_host + uri

    payload = {
        PROJECT_ID: int(project_id),
        'access_token': access_token
    }

    response = requests.put(url, json=payload)
    if not response.ok:
        log.error('Failed to update access token for project %s. StatusCode:  %d', project_id, response.status_code)
    
    return response