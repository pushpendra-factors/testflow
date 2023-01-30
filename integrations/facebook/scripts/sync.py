import copy
import datetime
import json
import logging as log
import sys
import traceback
from _datetime import timedelta
from datetime import datetime, date
from optparse import OptionParser
import requests
from requests.models import Response
from typing import List, Dict

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
ERR_MSG = 'err_msg'
STATUS = 'status'
PAGING = 'paging'
PROJECT_ID = 'project_id'
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
    AD_SET: 'adset',
    CAMPAIGN: 'campaign'
}

parser = OptionParser()
parser.add_option('--env', dest='env', default='development')
parser.add_option('--dry', dest='dry', help='', default='False')
parser.add_option('--skip_today', dest='skip_today', help='', default='False')
parser.add_option('--project_id', dest=PROJECT_ID, help='', default=None, type=int)
parser.add_option('--data_service_host', dest='data_service_host',
                  help='Data service host', default='http://localhost:8089')

(options, args) = parser.parse_args()


def get_datetime_from_datestring(date: str) -> datetime:
    date = date.split('-')
    date = datetime(int(date[0]), int(date[1]), int(date[2]))
    return date


def notify(env: str, source: str, message: str) -> requests.Response:
    if env != 'production':
        log.warning('Skipped notification for env %s payload %s', env, str(message))
        return None

    sns_url = 'https://fjnvg9a8wi.execute-api.us-east-1.amazonaws.com/v1/notify'
    payload = {'env': env, 'message': message, 'source': source}
    resp = requests.post(sns_url, json=payload)
    if not resp.ok:
        log.error('Failed to notify through sns.')
    return resp


def ping_healthcheck(env: str, healthcheck_id: str, message: dict, endpoint: str = '') -> None:
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


def get_time_ranges_list(date_start: str) -> List[Dict[str, str]]:
    if date_start == '0':
        date_object: date = datetime.now().date() - timedelta(days=MAX_LOOKBACK)
        days: int = MAX_LOOKBACK
    else:
        date_object = datetime.strptime(str(date_start), "%Y-%m-%d").date()
        days = (datetime.now().date() - date_object).days
        if days > MAX_LOOKBACK:
            date_object = datetime.now().date() - timedelta(days=MAX_LOOKBACK)
            days = MAX_LOOKBACK

    time_ranges: List[Dict[str, str]] = []

    for i in range(0, days):
        new_date: str = (date_object + timedelta(days=i)).strftime('%Y-%m-%d')
        time_range: Dict[str, str] = {'since': new_date, 'until': new_date}
        time_ranges.append(time_range)
    return time_ranges

def get_facebook_int_settings() -> json:
    uri: str = '/data_service/facebook/project/settings'
    url: str = options.data_service_host + uri

    response: Response = requests.get(url)
    if not response.ok:
        log.error('Failed to get facebook integration settings from data services')
        return
    return response.json()


def sync_for_project_and_customer_account(facebook_int_setting: dict, customer_account_id: str):
    project_id = facebook_int_setting[PROJECT_ID]
    facebook_int_setting_dup = copy.deepcopy(facebook_int_setting)
    facebook_int_setting_dup[FACEBOOK_AD_ACCOUNT] = customer_account_id

    sync_info_with_type: dict = get_last_sync_info(project_id, customer_account_id)
    response: dict = get_collections(facebook_int_setting_dup, sync_info_with_type)
    response[PROJECT_ID] = project_id
    response['ad_account'] = facebook_int_setting_dup[FACEBOOK_AD_ACCOUNT]
    return response

def get_last_sync_info(project_id: str, account_id: str) -> dict:
    uri: str = '/data_service/facebook/documents/last_sync_info'
    url: str = options.data_service_host + uri
    payload: Dict[str, str] = {
        PROJECT_ID: project_id,
        'account_id': account_id
    }
    resp: requests.Response = requests.get(url, json=payload)
    all_info: dict = resp.json()
    sync_info_with_type: dict = {}
    for info in all_info:
        date = datetime.strptime(str(info['last_timestamp']), '%Y%m%d') + timedelta(days=1)
        sync_info_with_type[info['type_alias'] + info['platform']] = date.strftime('%Y-%m-%d')
    return sync_info_with_type


def get_and_insert_metadata(facebook_int_setting: dict, sync_info_with_type: dict) -> dict:
    status: str = ''
    errMsg: list = []
    request_counter: int = 0
    # campaign metadata
    fields_campaign: List[str] = ['id', 'name', 'account_id', 'buying_type', 'effective_status', 'spend_cap',
                                  'start_time',
                                  'stop_time']
    resp: dict
    metadata: list
    resp, metadata = get_and_insert_paginated_metadata(facebook_int_setting, CAMPAIGN, fields_campaign)
    if resp[STATUS] == 'failed':
        status = 'failed'
        errMsg.append(resp[ERR_MSG])
    elif CAMPAIGN + FACEBOOK not in sync_info_with_type:
        backfill_metadata(facebook_int_setting, CAMPAIGN, metadata)
    request_counter += resp[API_REQUESTS]

    # adset metadata
    fields_adset: List[str] = ['id', 'account_id', 'campaign_id', 'configured_status', 'daily_budget',
                               'effective_status',
                               'end_time', 'name', 'start_time', 'stop_time']

    resp, metadata = get_and_insert_paginated_metadata(facebook_int_setting, AD_SET, fields_adset)
    if resp[STATUS] == 'failed':
        status = 'failed'
        errMsg.append(resp[ERR_MSG])
    elif AD_SET + FACEBOOK not in sync_info_with_type:
        backfill_metadata(facebook_int_setting, AD_SET, metadata)
    request_counter += resp[API_REQUESTS]

    if status == 'failed':
        return {STATUS: 'failed', ERR_MSG: errMsg, API_REQUESTS: request_counter}
    return {('%s' % STATUS): 'success', ('%s' % ERR_MSG): '', API_REQUESTS: request_counter}


def get_and_insert_paginated_metadata(facebook_int_setting: dict, doc_type: str, fields: List[str]) -> (dict, list):
    request_counter: int = 0
    metadata: list = []
    records_counter: int
    log.warning("Fetching %s metadata started for Project %s", doc_type, facebook_int_setting[PROJECT_ID])
    url: str = 'https://graph.facebook.com/v15.0/{}/{}s?fields={}&&access_token={}&&limit=1000'.format(
        facebook_int_setting[FACEBOOK_AD_ACCOUNT], doc_type_map[doc_type], fields, facebook_int_setting[ACCESS_TOKEN])
    resp: requests.Response = requests.get(url)
    request_counter += 1
    if not resp.ok:
        err_string = 'Failed to get {}s metadata from facebook. StatusCode: {}. Error: {}. Project_id: {}'.format(
            doc_type, resp.status_code, resp.text, facebook_int_setting[PROJECT_ID])
        log.error(err_string)
        log.warning("Fetching %s metadata ended for Project %s", doc_type, facebook_int_setting[PROJECT_ID])
        return {STATUS: 'failed', ERR_MSG: err_string, API_REQUESTS: request_counter}, metadata
    for data in resp.json()[DATA]:
        timestamp = int(datetime.now().strftime('%Y%m%d'))
        add_facebook_document(facebook_int_setting[PROJECT_ID], facebook_int_setting[FACEBOOK_AD_ACCOUNT],
                              doc_type, data['id'], data, timestamp, FACEBOOK)
    metadata.extend(resp.json()[DATA])
    records_counter = len(metadata)

    # paging
    if ('%s' % PAGING) not in resp.json():
        records_log_string: str = "No. of {} metdata records fetch for Project {} : {}" \
            .format(doc_type, facebook_int_setting[PROJECT_ID], records_counter)
        log.warning(records_log_string)
        log.warning("Fetching %s metadata ended for Project %s", doc_type, facebook_int_setting[PROJECT_ID])
        return {STATUS: 'success', ERR_MSG: '', API_REQUESTS: request_counter}, metadata
    while 'next' in resp.json()[PAGING]:
        url = resp.json()[PAGING]['next']
        resp = requests.get(url)
        request_counter += 1
        if not resp.ok:
            err_string = 'Failed to get {}s metadata from facebook post pagination. StatusCode: {}. Error: {}. Project_id: {}'.format(
                doc_type, resp.status_code, resp.text, facebook_int_setting[PROJECT_ID])
            log.error(err_string)
            log.warning("Fetching %s metadata ended for Project %s", doc_type, facebook_int_setting[PROJECT_ID])
            return {STATUS: 'failed', ERR_MSG: err_string, API_REQUESTS: request_counter}, metadata
        for data in resp.json()[DATA]:
            timestamp = int(datetime.now().strftime('%Y%m%d'))
            add_facebook_document(facebook_int_setting[PROJECT_ID],
                                  facebook_int_setting[FACEBOOK_AD_ACCOUNT], doc_type,
                                  data['id'], data, timestamp, FACEBOOK)
        metadata.extend(resp.json()[DATA])
    records_counter = len(metadata)
    records_log_string = "No. of {} metdata records fetch for Project {} : {}".format(doc_type, facebook_int_setting[
        PROJECT_ID], records_counter)
    log.warning(records_log_string)
    log.warning("Fetching %s metadata ended for Project %s", doc_type, facebook_int_setting[PROJECT_ID])
    return {STATUS: 'success', ERR_MSG: '', API_REQUESTS: request_counter}, metadata


def backfill_metadata(facebook_int_setting: dict, doc_type: str, metadata: list):
    records_counter: int = 0
    for days in range(1, MAX_LOOKBACK + 1):
        timestamp = int((datetime.now() - timedelta(days=days)).strftime('%Y%m%d'))
        for data in metadata:
            add_facebook_document(facebook_int_setting[PROJECT_ID], facebook_int_setting[FACEBOOK_AD_ACCOUNT],
                                  doc_type, data['id'], data, timestamp, FACEBOOK)
            records_counter += 1
    records_log_string = "No of records backfilled of type {} metadata from Project {}: {}"\
        .format(doc_type, facebook_int_setting[PROJECT_ID], records_counter)
    log.warning(records_log_string)


def get_collections(facebook_int_setting: dict, sync_info_with_type: dict) -> dict:
    response: dict = {STATUS: ''}
    status: str = ''
    err_msg: list = []
    request_counter: int = 0
    try:
        res: dict = get_and_insert_metadata(facebook_int_setting, sync_info_with_type)
        request_counter += res[API_REQUESTS]
        if res[STATUS] == 'failed':
            status = 'failed'
            err_msg.append(res[ERR_MSG])

        if CAMPAIGN_INSIGHTS + FACEBOOK not in sync_info_with_type:
            res_campaign: dict = get_campaign_insights(facebook_int_setting[PROJECT_ID],
                                                       facebook_int_setting[FACEBOOK_AD_ACCOUNT],
                                                       facebook_int_setting[ACCESS_TOKEN], '0')
        else:
            res_campaign = get_campaign_insights(facebook_int_setting[PROJECT_ID],
                                                 facebook_int_setting[FACEBOOK_AD_ACCOUNT],
                                                 facebook_int_setting[ACCESS_TOKEN],
                                                 sync_info_with_type[CAMPAIGN_INSIGHTS + FACEBOOK])
        request_counter += res_campaign[API_REQUESTS]
        if res_campaign[STATUS] == 'failed':
            status = 'failed'
            err_msg.append(res_campaign[ERR_MSG])

        if AD_SET_INSIGHTS + FACEBOOK not in sync_info_with_type:
            res_adset: dict = get_adset_insights(facebook_int_setting[PROJECT_ID],
                                                 facebook_int_setting[FACEBOOK_AD_ACCOUNT],
                                                 facebook_int_setting[ACCESS_TOKEN], '0')
        else:
            res_adset = get_adset_insights(facebook_int_setting[PROJECT_ID],
                                           facebook_int_setting[FACEBOOK_AD_ACCOUNT],
                                           facebook_int_setting[ACCESS_TOKEN],
                                           sync_info_with_type[AD_SET_INSIGHTS + FACEBOOK])
        request_counter += res_adset[API_REQUESTS]
        if res_adset[STATUS] == 'failed':
            status = 'failed'
            err_msg.append(res_adset[ERR_MSG])

        if AD_INSIGHTS + FACEBOOK not in sync_info_with_type:
            res_ad: dict = get_ad_insights(facebook_int_setting[PROJECT_ID], facebook_int_setting[FACEBOOK_AD_ACCOUNT],
                                           facebook_int_setting[ACCESS_TOKEN], '0')
        else:
            res_ad = get_ad_insights(facebook_int_setting[PROJECT_ID], facebook_int_setting[FACEBOOK_AD_ACCOUNT],
                                     facebook_int_setting[ACCESS_TOKEN], sync_info_with_type[AD_INSIGHTS + FACEBOOK])
        request_counter += res_ad[API_REQUESTS]
        if res_ad[STATUS] == 'failed':
            status = 'failed'
            err_msg.append(res_ad[ERR_MSG])

    except Exception as e:
        traceback.print_tb(e.__traceback__)
        response[STATUS] = 'failed'
        response['msg'] = 'Failed with exception ' + str(e)
        response[API_REQUESTS] = request_counter
        return response
    if status == 'failed':
        response[STATUS] = 'failed'
        response['msg'] = err_msg
        response[API_REQUESTS] = request_counter
        return response
    response[STATUS] = 'success'
    response[API_REQUESTS] = request_counter
    return response


def get_campaign_insights(project_id: str, ad_account_id: str, access_token: str, date_start: str) -> dict:
    fields: List[str] = ['account_currency', 'ad_id', 'ad_name', 'adset_name', 'campaign_name', 'adset_id',
                         'campaign_id',
                         'clicks', 'conversions',
                         'cost_per_conversion', 'cost_per_ad_click', 'date_start', 'cpc', 'cpm', 'cpp', 'ctr',
                         'date_stop', 'frequency', 'impressions', 'inline_post_engagement', 'social_spend', 'spend',
                         'unique_clicks', 'reach']
    return fetch_and_insert_insights(project_id, ad_account_id, access_token, CAMPAIGN_INSIGHTS, fields, date_start)


def get_adset_insights(project_id: str, ad_account_id: str, access_token: str, date_start: str) -> dict:
    fields: List[str] = ['account_currency', 'ad_id', 'ad_name', 'adset_name', 'campaign_name', 'adset_id',
                         'campaign_id',
                         'clicks', 'conversions',
                         'cost_per_conversion', 'cost_per_ad_click', 'cpc', 'cpm', 'cpp', 'ctr',
                         'date_start', 'date_stop', 'frequency', 'impressions', 'inline_post_engagement',
                         'social_spend', 'spend',
                         'unique_clicks', 'reach']
    return fetch_and_insert_insights(project_id, ad_account_id, access_token, AD_SET_INSIGHTS, fields, date_start)


def get_ad_insights(project_id: str, ad_account_id: str, access_token: str, date_start: str) -> dict:
    fields: List[str] = ['account_currency', 'ad_id', 'ad_name', 'adset_name', 'campaign_name', 'adset_id',
                         'campaign_id',
                         'clicks', 'conversions',
                         'cost_per_conversion', 'cost_per_ad_click', 'cpc', 'cpm', 'cpp', 'ctr',
                         'date_start', 'date_stop', 'frequency', 'impressions', 'inline_post_engagement',
                         'social_spend', 'spend',
                         'unique_clicks', 'reach']
    return fetch_and_insert_insights(project_id, ad_account_id, access_token, AD_INSIGHTS, fields, date_start)


# return statement: {STATUS: failed/success, errMsg: , api_requests: }
def fetch_and_insert_insights(project_id: str, ad_account_id: str, access_token: str, doc_type: str,
                              fields_insight: list, date_start: str) -> dict:
    request_counter: int = 0
    time_ranges: List[Dict[str, str]] = get_time_ranges_list(date_start)
    breakdowns: List[str] = ['publisher_platform']
    log.warning("Fetching %s started for Project %s", doc_type, project_id)
    for time_range in time_ranges:
        url: str = 'https://graph.facebook.com/v15.0/{}/insights?breakdowns={}&&time_range={}&&fields={}&&access_token={}&&level={}&&limit=1000'.format(
            ad_account_id, breakdowns, time_range, fields_insight, access_token, level_breakdown[doc_type])
        breakdown_response: Response = requests.get(url)
        request_counter += 1
        if not breakdown_response.ok:
            errString = 'Failed to get {} from facebook. StatusCode: {} Error: {}. Project_id: {}'\
                .format(doc_type, breakdown_response.status_code, breakdown_response.text, project_id)
            log.error(errString)
            log.warning("Fetching %s ended for Project %s", doc_type, project_id)
            return {STATUS: 'failed', ERR_MSG: errString, API_REQUESTS: request_counter}

        records_counter = len(breakdown_response.json()[DATA])
        for data in breakdown_response.json()[DATA]:
            date_stop = get_datetime_from_datestring(data['date_stop'])
            timestamp = int(date_stop.strftime('%Y%m%d'))
            add_facebook_document(project_id, ad_account_id, doc_type, data[id_fields[doc_type]], data, timestamp,
                                  data['publisher_platform'])

        # paging
        if 'paging' not in breakdown_response.json():
            continue
        while 'next' in breakdown_response.json()['paging']:
            url = breakdown_response.json()['paging']['next']
            breakdown_response = requests.get(url)
            request_counter += 1
            if not breakdown_response.ok:
                errString = 'Failed to get {} from facebook post pagination. StatusCode: {} Error: {}. Project_id: {}'\
                    .format(doc_type, breakdown_response.status_code, breakdown_response.text, project_id)
                log.error(errString)
                log.warning("Fetching %s ended for Project %s", doc_type, project_id)
                return {STATUS: 'failed', ERR_MSG: errString, API_REQUESTS: request_counter}

            records_counter += len(breakdown_response.json()[DATA])
            for data in breakdown_response.json()[DATA]:
                date_stop = get_datetime_from_datestring(data['date_stop'])
                timestamp = int(date_stop.strftime('%Y%m%d'))
                add_facebook_document(project_id, ad_account_id, doc_type, data[id_fields[doc_type]], data, timestamp,
                                      data['publisher_platform'])
        records_log_string = "No of {} records fetched for Project {} and timestamp {}: {}".format(doc_type, project_id,
                                                                                                   time_range['until'],
                                                                                                   records_counter)
        log.warning(records_log_string)
    log.warning("Fetching %s ended for Project %s", doc_type, project_id)
    return {STATUS: 'success', ERR_MSG: '', API_REQUESTS: request_counter}


def add_facebook_document(project_id: str, ad_account_id: str, doc_type: str, id: str, value: dict, timestamp: int,
                          platform: str) -> Response:
    uri = '/data_service/facebook/documents/add'
    url = options.data_service_host + uri

    payload = {
        PROJECT_ID: int(project_id),
        'customer_ad_account_id': ad_account_id,
        'type_alias': doc_type,
        'id': id,
        'value': value,
        'timestamp': timestamp,
        'platform': platform,
    }
    response = requests.post(url, json=payload)
    if not response.ok:
        log.error('Failed to add response %s to facebook warehouse for project %s. StatusCode:  %d, %s',
                  doc_type, project_id, response.status_code, response.json())

    return response


if __name__ == '__main__':
    facebook_int_settings: dict = get_facebook_int_settings()

    if facebook_int_settings is not None:
        now: datetime = datetime.now()
        failures: List[dict] = []
        successes: List[dict] = []
        for facebook_int_setting in facebook_int_settings:
            customer_account_ids = facebook_int_setting[FACEBOOK_AD_ACCOUNT].split(',')
            for customer_account_id in customer_account_ids:
                response = sync_for_project_and_customer_account(facebook_int_setting, customer_account_id)
                if response[STATUS] == 'failed':
                    failures.append(response)
                else:
                    successes.append(response)

        status_msg = ''
        if len(failures) > 0:
            status_msg = 'Failures on sync.'
        else:
            status_msg = 'Successfully synced.'
        notification_payload = {
            STATUS: status_msg,
            'failures': failures,
            'success': successes,
        }

        log.warning('Successfully synced. End of facebook sync job.')
        if len(failures) > 0:
            ping_healthcheck(options.env, HEALTHCHECK_PING_ID, notification_payload, endpoint='/fail')
        else:
            ping_healthcheck(options.env, HEALTHCHECK_PING_ID, notification_payload)
        sys.exit(0)
