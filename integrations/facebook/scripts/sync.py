from optparse import OptionParser
import json
import logging as log
import csv
import datetime
import requests
from datetime import datetime
import re
import sys
import time
import traceback
from _datetime import timedelta

parser = OptionParser()
parser.add_option('--env', dest='env', default='development')
parser.add_option('--dry', dest='dry', help='', default='False')
parser.add_option('--skip_today', dest='skip_today', help='', default='False') 
parser.add_option('--project_id', dest='project_id', help='', default=None, type=int)
parser.add_option('--data_service_host', dest='data_service_host',
    help='Data service host', default='http://localhost:8089')

(options, args) = parser.parse_args()

APP_NAME = 'facebook_sync'
CAMPAIGN_INSIGHTS = 'campaign_insights'
AD_SET_INSIGHTS = 'ad_set_insights'
AD_INSIGHTS = 'ad_insights'
CAMPAIGN = 'campaign'
AD = 'ad'
AD_ACCOUNT = 'ad_account'
AD_SET = 'ad_set'
ACCESS_TOKEN = 'int_facebook_access_token'
FACEBOOK_AD_ACCOUNT = 'int_facebook_ad_account'
DATA = 'data'
FACEBOOK = 'facebook'
PLATFORM = 'platform'
MAX_LOOKBACK = 30
API_REQUESTS = 'api_requests'

METRIC_TYPE_INCR = 'incr'
HEALTHCHECK_PING_ID = 'f2265955-a71c-42fe-a5ba-36d22a98419c'

level_breakdown = {
        AD_INSIGHTS: 'ad',
        AD_SET_INSIGHTS: 'adset',
        CAMPAIGN_INSIGHTS: 'campaign'
    }
id_fields = {
    AD_INSIGHTS: 'ad_id',
    AD_SET_INSIGHTS: 'adset_id',
    CAMPAIGN_INSIGHTS: 'campaign_id'
}
doc_type_map = {
    AD_SET : 'adset',
    CAMPAIGN: 'campaign'
}

def get_datetime_from_datestring(date):
    date = date.split('-')
    date = datetime(int(date[0]),int(date[1]),int(date[2]))
    return date

def notify(env, source, message):
    if env != 'production': 
        log.warning('Skipped notification for env %s payload %s', env, str(message))
        return

    sns_url = 'https://fjnvg9a8wi.execute-api.us-east-1.amazonaws.com/v1/notify'
    payload = { 'env': env, 'message': message, 'source': source }
    response = requests.post(sns_url, json=payload)
    if not response.ok: log.error('Failed to notify through sns.')
    return response

def ping_healthcheck(env, healthcheck_id, message, endpoint=''):
    message = json.dumps(message, indent=1)
    log.warning('Healthcheck ping for env %s payload %s', env, message)
    if env != 'production': 
        return

    try:
        requests.post('https://hc-ping.com/' + healthcheck_id + endpoint,
            data=message, timeout=10)
    except requests.RequestException as e:
        # Log ping failure here...
        log.error('Ping failed to healthchecks.io: %s' % e)

def record_metric(metric_type, metric_name, metric_value=0):
    payload = {
        'type': metric_type,
        'name': metric_name,
        'value': metric_value,
    }

    metrics_url = options.data_service_host + '/data_service/metrics'
    response = requests.post(metrics_url, json=payload)
    if not response.ok:
        log.error('Failed to record metric %s. Error: %s', metric_name, response.text)

def get_time_ranges_list(date_start):
    if date_start == 0:
        date_object = datetime.now().date() - timedelta(days=MAX_LOOKBACK)
        days = MAX_LOOKBACK
    else:
        date_object = datetime.strptime(str(date_start), "%Y-%m-%d").date()
        days = (datetime.now().date() - date_object).days
        if days > MAX_LOOKBACK:
            date_object = datetime.now().date() - timedelta(days=MAX_LOOKBACK)
            days = MAX_LOOKBACK

    time_ranges = []

    for i in range (0, days):
        new_date = (date_object + timedelta(days=i)).strftime('%Y-%m-%d')
        time_range = {'since':new_date, 'until':new_date}
        time_ranges.append(time_range)
    return time_ranges


def get_facebook_int_settings():
    uri = '/data_service/facebook/project/settings'
    url = options.data_service_host + uri

    response = requests.get(url)
    if not response.ok:
        log.error('Failed to get facebook integration settings from data services')
        return 
    return response.json()

def get_last_sync_info(project_id, account_id):
    uri = '/data_service/facebook/documents/last_sync_info'
    url = options.data_service_host + uri
    payload = {
        'project_id': project_id,
        'account_id' : account_id
    }
    response = requests.get(url, json=payload)
    all_info = response.json()
    sync_info_with_type = {}
    for info in all_info:
        date = datetime.strptime(str(info['last_timestamp']), '%Y%m%d') + timedelta(days=1)
        sync_info_with_type[info['type_alias']+info['platform']]= date.strftime('%Y-%m-%d')
    return sync_info_with_type

def get_and_insert_metadata(facebook_int_setting, sync_info_with_type):
    status = ''
    errMsg = []
    request_counter = 0
    # campaign metadata
    fields_campaign = ['id', 'name', 'account_id', 'buying_type','effective_status','spend_cap','start_time','stop_time']
    response, metadata = get_and_insert_paginated_metadata(facebook_int_setting, CAMPAIGN, fields_campaign)
    if response['status'] == 'failed':
        status = 'failed'
        errMsg.append(response['errMsg'])
    elif (CAMPAIGN+FACEBOOK not in sync_info_with_type):
        backfill_metadata(facebook_int_setting, CAMPAIGN, metadata)
    request_counter += response[API_REQUESTS]
    
    # adset metadata
    fields_adset = ['id', 'account_id','campaign_id','configured_status', 'daily_budget', 'effective_status','end_time','name','start_time','stop_time']
    response, metadata = get_and_insert_paginated_metadata(facebook_int_setting, AD_SET, fields_adset)
    if response['status'] == 'failed':
        status = 'failed'
        errMsg.append(response['errMsg'])
    elif (AD_SET+FACEBOOK not in sync_info_with_type):
        backfill_metadata(facebook_int_setting, AD_SET, metadata)
    request_counter += response[API_REQUESTS]

    if status == 'failed':
        return {'status': 'failed', 'errMsg': errMsg, API_REQUESTS: request_counter}
    return {'status': 'success', 'errMsg': '', API_REQUESTS: request_counter}

# return statement : {'status': failed/success, 'errString': , api_requests: }, metdata
def get_and_insert_paginated_metadata(facebook_int_setting, doc_type, fields):
    request_counter = 0
    metadata = []
    url = 'https://graph.facebook.com/v9.0/{}/{}s?fields={}&&access_token={}&&limit=1000'.format(
    facebook_int_setting[FACEBOOK_AD_ACCOUNT], doc_type_map[doc_type], fields, facebook_int_setting[ACCESS_TOKEN])
    response = requests.get(url)
    request_counter +=1
    if not response.ok:
        errString = 'Failed to get {}s metadata from facebook. StatusCode: {}. Error: {}. Project_id: {}'.format(doc_type, response.status_code, response.text, facebook_int_setting['project_id'])
        log.error(errString)
        return {'status': 'failed', 'errMsg': errString, API_REQUESTS: request_counter}, metadata
    for data in response.json()[DATA]:
        timestamp = int(datetime.now().strftime('%Y%m%d'))
        add_document_response = add_facebook_document(facebook_int_setting['project_id'], facebook_int_setting[FACEBOOK_AD_ACCOUNT], doc_type, data['id'], data, timestamp, FACEBOOK)
    metadata.extend(response.json()[DATA])
    
    # paging
    if 'paging' not in response.json():
        return {'status': 'success', 'errMsg': '',  API_REQUESTS: request_counter}, metadata
    while 'next' in response.json()['paging']:
        url = response.json()['paging']['next']
        response = requests.get(url)
        request_counter +=1
        if not response.ok:
            errString = 'Failed to get {}s metadata from facebook post pagination. StatusCode: {}. Error: {}. Project_id: {}'.format(doc_type, response.status_code, response.text, facebook_int_setting['project_id'])
            log.error(errString)
            return {'status': 'failed', 'errMsg': errString,  API_REQUESTS: request_counter}, metadata
        for data in response.json()[DATA]:
            timestamp = int(datetime.now().strftime('%Y%m%d'))
            add_document_response = add_facebook_document(facebook_int_setting['project_id'], facebook_int_setting[FACEBOOK_AD_ACCOUNT], doc_type, data['id'], data, timestamp, FACEBOOK)
        metadata.extend(response.json()[DATA])

    return {'status': 'success', 'errMsg': '', API_REQUESTS: request_counter}, metadata

def backfill_metadata(facebook_int_setting, doc_type, metadata):
    for days in range(1, MAX_LOOKBACK+1):
        timestamp = int((datetime.now()- timedelta(days=days)).strftime('%Y%m%d'))
        for data in metadata:
            add_document_response = add_facebook_document(facebook_int_setting['project_id'], facebook_int_setting[FACEBOOK_AD_ACCOUNT], doc_type, data['id'], data, timestamp, FACEBOOK)

def get_collections(facebook_int_setting, sync_info_with_type):
    response = {'status': ''}
    status = ''
    errMsg = []
    request_counter = 0
    try:
        res = get_and_insert_metadata(facebook_int_setting, sync_info_with_type)
        request_counter += res[API_REQUESTS]
        if res['status'] == 'failed':
            status = 'failed'
            errMsg.append(res['errMsg'])

        if (CAMPAIGN_INSIGHTS+FACEBOOK not in sync_info_with_type):
            res_campaign = get_campaign_insights(facebook_int_setting['project_id'], facebook_int_setting[FACEBOOK_AD_ACCOUNT],
                facebook_int_setting[ACCESS_TOKEN], 0)
        else:
            res_campaign = get_campaign_insights(facebook_int_setting['project_id'], facebook_int_setting[FACEBOOK_AD_ACCOUNT],
                facebook_int_setting[ACCESS_TOKEN], sync_info_with_type[CAMPAIGN_INSIGHTS+FACEBOOK])
        request_counter += res_campaign[API_REQUESTS]
        if res_campaign['status'] == 'failed':
            status = 'failed'
            errMsg.append(res_campaign['errMsg'])

        if (AD_SET_INSIGHTS+FACEBOOK not in sync_info_with_type):
            res_adset = get_adset_insights(facebook_int_setting['project_id'], facebook_int_setting[FACEBOOK_AD_ACCOUNT],
                facebook_int_setting[ACCESS_TOKEN], 0)
        else:
            res_adset = get_adset_insights(facebook_int_setting['project_id'], facebook_int_setting[FACEBOOK_AD_ACCOUNT],
                facebook_int_setting[ACCESS_TOKEN], sync_info_with_type[AD_SET_INSIGHTS+FACEBOOK])
        request_counter += res_adset[API_REQUESTS]
        if res_adset['status'] == 'failed':
            status = 'failed'
            errMsg.append(res_adset['errMsg'])

        if (AD_INSIGHTS+FACEBOOK not in sync_info_with_type):
            res_ad = get_ad_insights(facebook_int_setting['project_id'], facebook_int_setting[FACEBOOK_AD_ACCOUNT],
                facebook_int_setting[ACCESS_TOKEN], 0)
        else:
            res_ad = get_ad_insights(facebook_int_setting['project_id'], facebook_int_setting[FACEBOOK_AD_ACCOUNT],
                facebook_int_setting[ACCESS_TOKEN], sync_info_with_type[AD_INSIGHTS+FACEBOOK])
        request_counter += res_ad[API_REQUESTS]
        if res_ad['status'] == 'failed':
            status = 'failed'
            errMsg.append(res_ad['errMsg'])

    except Exception as e:
        traceback.print_tb(e.__traceback__)
        response['status'] = 'failed'
        response['msg'] = 'Failed with exception '+str(e)
        response[API_REQUESTS] = request_counter
        return response
    if status == 'failed':
        response['status'] = 'failed'
        response['msg'] = errMsg
        response[API_REQUESTS] = request_counter
        return response
    response['status']='success'
    response[API_REQUESTS] = request_counter
    return response
       

def get_campaign_insights(project_id, ad_account_id, access_token, date_start):
        
    fields = ['account_currency', 'ad_id','ad_name','adset_name','campaign_name','adset_id','campaign_id','clicks','conversions',
    'cost_per_conversion','cost_per_ad_click','date_start', 'cpc', 'cpm','cpp','ctr',
    'date_stop','frequency','impressions','inline_post_engagement','social_spend', 'spend','unique_clicks','reach']
    return fetch_and_insert_insights(project_id, ad_account_id, access_token, CAMPAIGN_INSIGHTS, fields, date_start)

def get_adset_insights(project_id, ad_account_id, access_token, date_start):

    fields = ['account_currency', 'ad_id','ad_name','adset_name','campaign_name','adset_id','campaign_id','clicks','conversions',
    'cost_per_conversion','cost_per_ad_click','cpc', 'cpm','cpp','ctr',
    'date_start','date_stop','frequency','impressions','inline_post_engagement','social_spend', 'spend','unique_clicks','reach']
    return fetch_and_insert_insights(project_id, ad_account_id, access_token, AD_SET_INSIGHTS, fields, date_start)

def get_ad_insights(project_id, ad_account_id, access_token, date_start):

    fields = ['account_currency', 'ad_id','ad_name','adset_name','campaign_name','adset_id','campaign_id','clicks','conversions',
    'cost_per_conversion','cost_per_ad_click','cpc', 'cpm','cpp','ctr',
    'date_start','date_stop','frequency','impressions','inline_post_engagement','social_spend', 'spend','unique_clicks','reach']
    return fetch_and_insert_insights(project_id, ad_account_id, access_token, AD_INSIGHTS, fields, date_start)

# return statement: {'status': failed/success, errMsg: , api_requests: }
def fetch_and_insert_insights(project_id, ad_account_id, access_token, doc_type, fields_insight, date_start):
    request_counter = 0
    time_ranges = get_time_ranges_list(date_start)
    breakdowns = ['publisher_platform']
    for time_range in time_ranges:
        url = 'https://graph.facebook.com/v9.0/{}/insights?breakdowns={}&&time_range={}&&fields={}&&access_token={}&&level={}&&limit=1000'.format(
        ad_account_id, breakdowns, time_range, fields_insight, access_token, level_breakdown[doc_type])
        breakdown_response = requests.get(url)
        request_counter +=1
        if not breakdown_response.ok:
            errString = 'Failed to get {} insights from facebook. StatusCode: {} Error: {}. Project_id: {}'.format(doc_type, breakdown_response.status_code, breakdown_response.text, project_id)
            log.error(errString)
            return {'status': 'failed', 'errMsg': errString, API_REQUESTS: request_counter}

        for data in breakdown_response.json()[DATA]:
            date_stop = get_datetime_from_datestring(data['date_stop'])
            timestamp= int(datetime.strftime(date_stop, '%Y%m%d'))
            add_facebook_document(project_id, ad_account_id, doc_type, data[id_fields[doc_type]], data, timestamp, data['publisher_platform'])

        # paging
        if 'paging' not in breakdown_response.json():
            continue
        while 'next' in breakdown_response.json()['paging']:
            url = breakdown_response.json()['paging']['next']
            breakdown_response = requests.get(url)
            request_counter +=1
            if not breakdown_response.ok:
                errString = 'Failed to get {} insights from facebook post pagination. StatusCode: {} Error: {}. Project_id: {}'.format(doc_type, breakdown_response.status_code, breakdown_response.text, project_id)
                log.error(errString)
                return {'status': 'failed', 'errMsg': errString, API_REQUESTS: request_counter}   

            for data in breakdown_response.json()[DATA]:
                date_stop = get_datetime_from_datestring(data['date_stop'])
                timestamp= int(datetime.strftime(date_stop, '%Y%m%d'))
                add_facebook_document(project_id, ad_account_id, doc_type, data[id_fields[doc_type]], data, timestamp, data['publisher_platform'])
    return {'status': 'success', 'errMsg': '', API_REQUESTS: request_counter}


def add_facebook_document(project_id, ad_account_id, doc_type, id, value, timestamp, platform):
    uri = '/data_service/facebook/documents/add'
    url = options.data_service_host + uri

    payload = {
        'project_id': int(project_id),
        'customer_ad_account_id': ad_account_id,
        'type_alias': doc_type,
        'id': id,
        'value': value,
        'timestamp':timestamp,
        'platform': platform,
    }
    response = requests.post(url, json=payload)
    if not response.ok:
        log.error('Failed to add response %s to facebook warehouse for project %s. StatusCode:  %d, %s', 
            doc_type, project_id, response.status_code, response.json())
    
    return response


if __name__ == '__main__':
    facebook_int_settings = get_facebook_int_settings()

    if(facebook_int_settings is not None):
        now = datetime.now()
        failures = []
        successes = []
        for facebook_int_setting in facebook_int_settings:
            sync_info_with_type = get_last_sync_info(facebook_int_setting['project_id'],facebook_int_setting['int_facebook_ad_account'])
            response = get_collections(facebook_int_setting, sync_info_with_type)
            response['project_id']= facebook_int_setting['project_id']
            response['ad_account']= facebook_int_setting[FACEBOOK_AD_ACCOUNT]
            if(response['status']=='failed'):
                failures.append(response)
            else:
                successes.append(response)
        status_msg = ''
        if len(failures) > 0: status_msg = 'Failures on sync.'
        else: status_msg = 'Successfully synced.'
        notification_payload = {
            'status': status_msg, 
            'failures': failures, 
            'success': successes,
        }

        log.warning('Successfully synced. End of facebook sync job.')
        if len(failures) > 0:
            ping_healthcheck(options.env, HEALTHCHECK_PING_ID, notification_payload, endpoint='/fail')
        else:
            ping_healthcheck(options.env, HEALTHCHECK_PING_ID, notification_payload)
        sys.exit(0)
        
        
