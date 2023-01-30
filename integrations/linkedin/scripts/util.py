from datetime import datetime
from constants import *
from _datetime import timedelta
import requests
import json
import logging as log

def get_separated_date(date):
    date = date.split('-')
    return date[0], date[1], date[2]

def get_split_date_from_timestamp(date):
    new_date = datetime.strptime(str(date), '%Y%m%d').date()
    return new_date.year, int(new_date.month), new_date.day

def get_timestamp(date):
    return int(datetime(date['year'],date['month'],date['day']).strftime('%Y%m%d'))

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

def sort_by_timestamp(data):
    return get_timestamp(data['dateRange']['end'])

def get_timestamp_range(doc_type, sync_info_with_type, start_timestamp, end_timestamp):
    timestamps =[]
    date_start = ''
    date_end = ''

    if start_timestamp != None:
        date_start = datetime.strptime(str(start_timestamp), '%Y%m%d')
        if end_timestamp != None:
            date_end = datetime.strptime(str(end_timestamp), '%Y%m%d')
        else:
            date_end = datetime.now() - timedelta(days=1)
    else:
        if doc_type not in sync_info_with_type:
            date_start = (datetime.now() - timedelta(days=MAX_LOOKBACK))
        else:
            date_start = datetime.strptime(str(sync_info_with_type[doc_type]), '%Y-%m-%d')
        date_end = datetime.now() - timedelta(days=1)
    num_of_days = (date_end-date_start).days
    if num_of_days <=0:
        return []

    for i in range (0, num_of_days):
        date_start = date_start + timedelta(days=1)
        date_required = date_start.strftime("%Y%m%d")
        timestamps.append(date_required)
    
    #if range greater than max lookback, get latest range with length of maxlookback
    if len(timestamps) > MAX_LOOKBACK:
        timestamps = timestamps[-MAX_LOOKBACK:]
    return timestamps

def update_result_with_metadata(response, doc_type, campaign_group_meta, campaign_meta, creative_meta):
    final_response = []
    for data in response:
        id = data['pivotValue'].split(':')[3]
        data.update({'id': id})
        if doc_type == CAMPAIGN_GROUP_INSIGHTS:
            if id in campaign_group_meta:
                data.update(campaign_group_meta[id])
        elif doc_type == CAMPAIGN_INSIGHTS:
            if id in campaign_meta:
                data.update(campaign_meta[id])
                if campaign_meta[id][CAMPAIGN_GROUP_ID] in campaign_group_meta:
                    data.update(campaign_group_meta[campaign_meta[id][CAMPAIGN_GROUP_ID]])
        elif doc_type == CREATIVE_INSIGHTS:
            if id in creative_meta:
                data.update(creative_meta[id])
                if creative_meta[id][CAMPAIGN_GROUP_ID] in campaign_group_meta:
                    data.update(campaign_group_meta[creative_meta[id][CAMPAIGN_GROUP_ID]])
                if creative_meta[id][CAMPAIGN_ID] in campaign_meta:
                    data.update(campaign_meta[creative_meta[id][CAMPAIGN_ID]])
        final_response.append(data)
    
    return final_response