from optparse import OptionParser
import json
import logging as log
import csv
import datetime
import requests
import copy
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
parser.add_option('--project_ids', dest='project_ids', help='', default=None, type=str)
parser.add_option('--client_id', dest='client_id', help='',default=None, type=str)
parser.add_option('--client_secret', dest='client_secret', help='',default=None, type=str)
parser.add_option('--start_timestamp', dest='start_timestamp', help='', default=None, type=int)
parser.add_option('--end_timestamp', dest='end_timestamp', help='', default=None, type=int)
parser.add_option('--insert_metadata', dest='insert_metadata', help='', default='True')
parser.add_option('--insert_report', dest='insert_report', help='', default='True')
parser.add_option('--data_service_host', dest='data_service_host',
    help='Data service host', default='http://localhost:8089')
parser.add_option('--run_member_insights_only', dest='run_member_insights_only', help='', default='False')
# remove the following flag after testing, remove follwing flag from yaml if removing from here
parser.add_option('--member_company_project_ids', dest='member_company_project_ids', help='', default=None, type=str)

(options, args) = parser.parse_args()

APP_NAME = 'linkedin_sync'
CAMPAIGN_GROUP_INSIGHTS = 'campaign_group_insights'
CAMPAIGN_INSIGHTS = 'campaign_insights'
CREATIVE_INSIGHTS = 'creative_insights'
MEMBER_COMPANY_INSIGHTS = 'member_company_insights'
CAMPAIGN = 'campaign'
CAMPAIGNS = 'campaign'
CREATIVES = 'creative'
CAMPAIGN_GROUPS = 'campaign_group'
AD_ACCOUNT = 'ad_account'
ACCESS_TOKEN = 'int_linkedin_access_token'
REFRESH_TOKEN = 'int_linkedin_refresh_token'
LINKEDIN_AD_ACCOUNT = 'int_linkedin_ad_account'
ELEMENTS = 'elements'
PROJECT_ID = 'project_id'
CAMPAIGN_GROUP_ID = 'campaign_group_id'
CAMPAIGN_ID = 'campaign_id'
CREATIVE_ID = 'creative_id'
MAX_LOOKBACK = 30
API_REQUESTS = 'api_requests'
META_COUNT = 100
INSIGHTS_COUNT = 10000

METRIC_TYPE_INCR = 'incr'
HEALTHCHECK_PING_ID = '837dce09-92ec-4930-80b3-831b295d1a34'
start_timestamp = None
end_timestamp = None

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
        
def get_linkedin_int_settings():
    uri = '/data_service/linkedin/project/settings'
    url = options.data_service_host + uri

    response = requests.get(url)
    if not response.ok:
        log.error('Failed to get linkedin integration settings from data services')
        return 
    return response.json()

def get_linkedin_int_settings_for_projects(project_ids):
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

def get_last_sync_info(linkedin_int_setting):
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
        date = datetime.strptime(str(info['last_timestamp']), '%Y%m%d')+ timedelta(days=1)
        sync_info_with_type[info['type_alias']]= date.strftime('%Y-%m-%d')
    return sync_info_with_type, ''

def get_separated_date(date):
    date = date.split('-')
    return date[0], date[1], date[2]

def add_linkedin_documents(project_id, ad_account_id, doc_type, obj_id, value, timestamp):
    uri = '/data_service/linkedin/documents/add'
    url = options.data_service_host + uri

    payload = {
        PROJECT_ID: int(project_id),
        'customer_ad_account_id': ad_account_id,
        'type_alias': doc_type,
        'id': obj_id,
        'value': value,
        'timestamp':timestamp
    }

    response = requests.post(url, json=payload)
    if not response.ok:
        log.error('Failed to add response %s to linkedin warehouse for project %s. StatusCode:  %d, %s', 
            doc_type, project_id, response.status_code, response.text)
    
    return response


def get_timestamp(date):
    return int(datetime(date['year'],date['month'],date['day']).strftime('%Y%m%d'))

# can't keep very long range, we might hit rate limit
def get_insights(linkedin_int_setting, sync_info_with_type, doc_type, pivot, meta_request_count):
    log.warning("Fetching insights for %s started for project %s", doc_type, linkedin_int_setting[PROJECT_ID])
    date_start = ''
    date_end = ''
    if start_timestamp != None:
        date_start = str(datetime.strptime(str(start_timestamp), '%Y%m%d').date())
        if end_timestamp != None:
            date_end = str(datetime.strptime(str(end_timestamp), '%Y%m%d').date())
        else:
            date_end = (datetime.now() - timedelta(days=1)).strftime('%Y-%m-%d')
    else:
        if doc_type not in sync_info_with_type:
            date_start = (datetime.now() - timedelta(days=MAX_LOOKBACK)).strftime('%Y-%m-%d')
        else:
            date_start = sync_info_with_type[doc_type]
        date_end = (datetime.now() - timedelta(days=1)).strftime('%Y-%m-%d')
    
    start_year, start_month, start_day = get_separated_date(date_start)
    end_year, end_month, end_day = get_separated_date(date_end)

    request_counter = meta_request_count
    records = 0
    results =[]

    if date_start > date_end:
        errString = 'Skipped getting {} insights from linkedin. Already synced. Project_id: {}. Date start: {}, Date end: {}'.format(pivot, linkedin_int_setting[PROJECT_ID], date_start, date_end)
        return {'status': 'skipped', 'errMsg': errString, API_REQUESTS: request_counter}

    fields='totalEngagements,impressions,clicks,dateRange,landingPageClicks,costInUsd,leadGenerationMailContactInfoShares,leadGenerationMailInterestedClicks,opens,videoCompletions,videoFirstQuartileCompletions,videoMidpointCompletions,videoThirdQuartileCompletions,videoViews,externalWebsiteConversions,externalWebsitePostClickConversions,externalWebsitePostViewConversions,costInLocalCurrency,conversionValueInLocalCurrency,pivotValue'
    url = 'https://api.linkedin.com/v2/adAnalyticsV2?q=analytics&pivot={}&dateRange.start.day={}&dateRange.start.month={}&dateRange.start.year={}&dateRange.end.day={}&dateRange.end.month={}&dateRange.end.year={}&timeGranularity=DAILY&fields={}&accounts[0]=urn:li:sponsoredAccount:{}&start=0&count={}'.format(
        pivot, start_day, start_month, start_year, end_day, end_month, end_year, fields, linkedin_int_setting[LINKEDIN_AD_ACCOUNT], INSIGHTS_COUNT)
    headers = {'Authorization': 'Bearer ' + linkedin_int_setting[ACCESS_TOKEN]}
    response = requests.get(url, headers=headers)
    request_counter += 1
    if not response.ok:
        errString = 'Failed to get {} insights from linkedin. StatusCode: {}. Error: {}. Project_id: {}'.format(pivot, response.status_code, response.text, linkedin_int_setting[PROJECT_ID])
        return {'status': 'failed', 'errMsg': errString, API_REQUESTS: request_counter}
    if ELEMENTS in response.json():
        records += len(response.json()[ELEMENTS])
        results.extend(response.json()[ELEMENTS])

    start = 0
    # paging
    while len(response.json()[ELEMENTS])>=INSIGHTS_COUNT:
        start += INSIGHTS_COUNT
        fields='totalEngagements,impressions,clicks,dateRange,landingPageClicks,approximateUniqueImpressions,shares,costInUsd,leadGenerationMailContactInfoShares,leadGenerationMailInterestedClicks,oneClickLeadFormOpens,oneClickLeads,opens,videoCompletions,videoFirstQuartileCompletions,videoMidpointCompletions,videoThirdQuartileCompletions,videoViews,externalWebsiteConversions,externalWebsitePostClickConversions,externalWebsitePostViewConversions,costInLocalCurrency,conversionValueInLocalCurrency,pivotValue,pivotValues'
        url = 'https://api.linkedin.com/v2/adAnalyticsV2?q=analytics&pivot={}&dateRange.start.day={}&dateRange.start.month={}&dateRange.start.year={}&timeGranularity=DAILY&fields={}&accounts[0]=urn:li:sponsoredAccount:{}&start={}&count={}'.format(
            pivot, start_day, start_month, start_year, fields, linkedin_int_setting[LINKEDIN_AD_ACCOUNT], start, INSIGHTS_COUNT)
        headers = {'Authorization': 'Bearer ' + linkedin_int_setting[ACCESS_TOKEN]}
        response = requests.get(url, headers=headers)
        request_counter +=1
        if not response.ok:
            errString = 'Failed to get {} insights after pagination from linkedin. StatusCode: {}. Error: {}. Project_id: {}'.format(pivot, response.status_code, response.text, linkedin_int_setting[PROJECT_ID])
            return {'status': 'failed', 'errMsg': errString, API_REQUESTS: request_counter}
        if ELEMENTS in response.json():
            records += len(response.json()[ELEMENTS])
            results.extend(response.json()[ELEMENTS])

    log.warning("No of %s insights records to be inserted for project %s : %d", doc_type, linkedin_int_setting[PROJECT_ID], records)
    return results, {'status': 'success', 'errMsg': '', API_REQUESTS: request_counter}

def insert_insights(doc_type, project_id, ad_account, response, campaign_group_meta, campaign_meta, creative_meta):
    for data in response:
        timestamp = get_timestamp(data['dateRange']['end'])
        id = data['pivotValue'].split(':')[3]
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

        add_documents_response = add_linkedin_documents(project_id, ad_account, doc_type, id, data, timestamp)
        if not add_documents_response.ok:
            return
    log.warning("Insertion of insights of %s ended for project %s", doc_type, project_id)

# it take organization id for report rows and fetches org details like name location domain and append it to the member company report rows
def update_org_data(records, access_token):
    mapIDs = {}
    for data in records:
        id = data['pivotValue'].split(':')[3]
        mapIDs[id]= True
    idStr = ''
    for key in mapIDs:
        if idStr == '':
            idStr += key
        else:
            idStr += (',' + key)
    
    url = 'https://api.linkedin.com/v2/organizationsLookup?ids=List({})'.format(idStr)
    headers = {'Authorization': 'Bearer ' + access_token, 'X-Restli-Protocol-Version': '2.0.0'}
    response = requests.get(url, headers=headers)
    if not response.ok or 'results' not in response.json():
        return [], 1,  "Failed getting organisation data"
    map_id_to_org_data = response.json()['results']
    for data in records:
        id = data['pivotValue'].split(':')[3]
        if id not in map_id_to_org_data:
            return "Failed getting organisation data for id {}".format(id)
        if 'vanityName' in map_id_to_org_data[id]:
            data['vanityName'] = map_id_to_org_data[id]['vanityName']
        else:
            data['vanityName'] = '$none'

        if 'localizedName' in map_id_to_org_data[id]:
            data['localizedName'] = map_id_to_org_data[id]['localizedName']
        else:
            data['localizedName'] = '$none'

        if 'localizedWebsite' in map_id_to_org_data[id]:
            data['localizedWebsite'] = map_id_to_org_data[id]['localizedWebsite']
        else:
            data['localizedWebsite'] = '$none'
        
        if 'name' in map_id_to_org_data[id] and 'preferredLocale' in map_id_to_org_data[id]['name'] and 'country' in map_id_to_org_data[id]['name']['preferredLocale']:
            data['preferredCountry'] = map_id_to_org_data[id]['name']['preferredLocale']['country']
        else:
            data['preferredCountry'] = '$none'
        
        if 'locations' in map_id_to_org_data[id]:
            for location in map_id_to_org_data[id]['locations']:
                if 'locationType' in location and location['locationType'] == 'HEADQUARTERS' and 'address' in location and 'country' in location['address']:
                    data['companyHeadquarters'] = location['address']['country']
                    break
                else:
                    data['companyHeadquarters'] = '$none'

    return records, 1, ''

def get_company_name_and_insert(doc_type, project_id, ad_account, access_token, response, request_counter):
    if len(response) !=0:
        records, requests, errString = update_org_data(response, access_token)
        if errString != '':
            return {'status': 'failed', 'errMsg': errString, API_REQUESTS: request_counter + requests}
    else:
        log.warning('No data found for member company insights for project {} and Ad account {}'.format(project_id, ad_account))
    
    insert_insights(MEMBER_COMPANY_INSIGHTS,linkedin_int_setting[PROJECT_ID],linkedin_int_setting[LINKEDIN_AD_ACCOUNT], records, {}, {}, {})
    return {'status': 'success', 'errMsg': '', API_REQUESTS: request_counter + requests}

def get_metadata(ad_account, access_token, url_endpoint, doc_type, project_id):
    metadata = []
    request_counter = 0
    url = 'https://api.linkedin.com/v2/{}?q=search&search.account.values[0]=urn:li:sponsoredAccount:{}&start=0&count={}'.format(url_endpoint,ad_account, META_COUNT)
    headers = {'Authorization': 'Bearer ' + access_token}
    response = requests.get(url, headers=headers)
    request_counter += 1
    if not response.ok:
        errString = 'Failed to get {} metadata from linkedin. StatusCode: {}. Error: {}. Project_id: {}'.format(doc_type, response.status_code, response.text, project_id)
        return metadata, errString, request_counter
    metadata.extend(response.json()[ELEMENTS])
    
    # paging
    start = 0
    while len(response.json()[ELEMENTS])>=META_COUNT:
        start +=META_COUNT
        url = 'https://api.linkedin.com/v2/{}?q=search&search.account.values[0]=urn:li:sponsoredAccount:{}&start={}&count={}'.format(url_endpoint, ad_account, start, META_COUNT)
        headers = {'Authorization': 'Bearer ' + access_token}
        response = requests.get(url, headers=headers)
        request_counter += 1
        if not response.ok:
            errString = 'Failed to get {} metadata from linkedin. StatusCode: {}. Error: {}. Project_id: {}'.format(doc_type, response.status_code, response.text, project_id)
            return metadata, errString, request_counter
        metadata.extend(response.json()[ELEMENTS])
    return metadata, '', request_counter

def insert_metadata(doc_type, project_id, ad_account, response, timestamp, extraMeta):
    for data in response:
        data.update(extraMeta[str(data['id'])])
        add_documents_response = add_linkedin_documents(project_id, ad_account, doc_type, str(data['id']),data, timestamp)
        if not add_documents_response.ok:
            return

def enrich_metadata_previous_dates(doc_type, project_id, ad_account, metadata, meta):
    days = MAX_LOOKBACK
    end_date = datetime.now()
    if start_timestamp != None:
        if end_timestamp != None:
            days = end_timestamp - start_timestamp
            end_date = datetime.strptime(str(end_timestamp), '%Y%m%d')
        else:
            days = int((datetime.now()).strftime('%Y%m%d')) - start_timestamp
    if days == 0:
        log.warning("No of metadata records to be backfilled for %s for project %s : %d", doc_type, project_id, 0)
        return
    log.warning("No of metadata records to be backfilled for %s for project %s : %d", doc_type, project_id, len(metadata)*(days))
    for i in range (1, days+1):
        timestamp = int((end_date-timedelta(days=i)).strftime('%Y%m%d'))
        insert_metadata(doc_type, project_id, ad_account, metadata, timestamp, meta)

def get_campaign_group_data(linkedin_int_setting, sync_info_with_type, meta):
    log.warning("Fetching metadata for campaign group started for project %s", linkedin_int_setting[PROJECT_ID])
    metadata, errString, request_counter = get_metadata(linkedin_int_setting[LINKEDIN_AD_ACCOUNT], linkedin_int_setting[ACCESS_TOKEN], 'adCampaignGroupsV2', CAMPAIGN_GROUPS, linkedin_int_setting[PROJECT_ID])
    if errString != '':
        return {'status': 'failed', 'errMsg': errString, API_REQUESTS: request_counter}
    for data in metadata:
        meta[str(data['id'])] = {CAMPAIGN_GROUP_ID: str(data['id']), 'campaign_group_name': data['name'], 'campaign_group_status': data['status']}
    timestamp = int(datetime.now().strftime('%Y%m%d'))
    if end_timestamp != None:
        timestamp = end_timestamp
    
    if options.insert_metadata != 'False':
        log.warning("No of metadata records for campaign group to be inserted for project %s : %d", linkedin_int_setting[PROJECT_ID], len(metadata))
        insert_metadata(CAMPAIGN_GROUPS, linkedin_int_setting[PROJECT_ID], linkedin_int_setting[LINKEDIN_AD_ACCOUNT], metadata, timestamp, meta)
        
        log.warning("Insertion metadata for campaign group ended for project %s", linkedin_int_setting[PROJECT_ID])

        if CAMPAIGN_GROUPS not in sync_info_with_type:
            log.warning("Backfilling campaign group metadata started for project %s", linkedin_int_setting[PROJECT_ID])
            enrich_metadata_previous_dates(CAMPAIGN_GROUPS, linkedin_int_setting[PROJECT_ID], linkedin_int_setting[LINKEDIN_AD_ACCOUNT], metadata, meta)
            log.warning("Backfilling campaign group metadata ended for project %s", linkedin_int_setting[PROJECT_ID])
    
    results, resp = get_insights(linkedin_int_setting, sync_info_with_type, CAMPAIGN_GROUP_INSIGHTS, 'CAMPAIGN_GROUP', request_counter)
    if resp['status'] == 'failed' or resp['errMsg'] != '':
        return resp
        
    insert_insights(CAMPAIGN_GROUP_INSIGHTS,linkedin_int_setting[PROJECT_ID],linkedin_int_setting[LINKEDIN_AD_ACCOUNT], results, meta, {}, {})
    return resp

def get_campaign_data(linkedin_int_setting, sync_info_with_type , campaign_group_meta, meta):
    log.warning("Fetching metadata for campaign started for project %s", linkedin_int_setting[PROJECT_ID])
    metadata, errString, request_counter = get_metadata(linkedin_int_setting[LINKEDIN_AD_ACCOUNT], linkedin_int_setting[ACCESS_TOKEN], 'adCampaignsV2', CAMPAIGNS, linkedin_int_setting[PROJECT_ID])
    if errString != '':
        return {'status': 'failed', 'errMsg': errString, API_REQUESTS: request_counter}
    for data in metadata:
        campaign_group_id = str(data['campaignGroup'].split(':')[3])
        meta[str(data['id'])] = {'campaign_group_id': campaign_group_id,'campaign_id': str(data['id']), 'campaign_name': data['name'], 'campaign_status': data['status'], 'campaign_type': data['type']}
    timestamp = int(datetime.now().strftime('%Y%m%d'))
    if end_timestamp != None:
        timestamp = end_timestamp
    
    if options.insert_metadata != 'False':
        log.warning("No of metadata records for campaign to be inserted for project %s : %d", linkedin_int_setting[PROJECT_ID], len(metadata))
        
        insert_metadata(CAMPAIGNS, linkedin_int_setting[PROJECT_ID], linkedin_int_setting[LINKEDIN_AD_ACCOUNT], metadata, timestamp, meta)
        
        log.warning("Insertion metadata for campaign ended for project %s", linkedin_int_setting[PROJECT_ID])

        if CAMPAIGNS not in sync_info_with_type:
            log.warning("Backfilling campaign metadata started for project %s", linkedin_int_setting[PROJECT_ID])
            enrich_metadata_previous_dates(CAMPAIGNS, linkedin_int_setting[PROJECT_ID], linkedin_int_setting[LINKEDIN_AD_ACCOUNT], metadata, meta)
            log.warning("Backfilling campaign metadata ended for project %s", linkedin_int_setting[PROJECT_ID])

    results, resp =  get_insights(linkedin_int_setting, sync_info_with_type, CAMPAIGN_INSIGHTS, 'CAMPAIGN', request_counter)
    if resp['status'] == 'failed' or resp['errMsg'] != '':
        return resp
    
    insert_insights(CAMPAIGN_INSIGHTS,linkedin_int_setting[PROJECT_ID],linkedin_int_setting[LINKEDIN_AD_ACCOUNT], results, campaign_group_meta, meta, {})
    return resp

def get_creative_data(linkedin_int_setting, sync_info_with_type, campaign_group_meta, campaign_meta, meta):
    log.warning("Fetching metadata for creative started for project %s", linkedin_int_setting[PROJECT_ID])
    metadata, errString, request_counter = get_metadata(linkedin_int_setting[LINKEDIN_AD_ACCOUNT], linkedin_int_setting[ACCESS_TOKEN], 'adCreativesV2', CREATIVES, linkedin_int_setting[PROJECT_ID])
    if errString != '':
        return {'status': 'failed', 'errMsg': errString, API_REQUESTS: request_counter}
    for data in metadata:
        campaign_id = str(data['campaign'].split(':')[3])
        campaign_group_id = campaign_meta[campaign_id][CAMPAIGN_GROUP_ID]
        meta[str(data['id'])] = {'campaign_group_id': campaign_group_id, 'campaign_id': campaign_id ,'creative_id': str(data['id']), 'creative_status': data['status'], 'creative_type': data['type']}
    timestamp = int(datetime.now().strftime('%Y%m%d'))
    if end_timestamp != None:
        timestamp = end_timestamp
    
    if options.insert_metadata != 'False':
        log.warning("No of metadata records for creative to be inserted for project %s : %d", linkedin_int_setting[PROJECT_ID], len(metadata))
        
        insert_metadata(CREATIVES, linkedin_int_setting[PROJECT_ID], linkedin_int_setting[LINKEDIN_AD_ACCOUNT], metadata, timestamp, meta)
        log.warning("Insertion metadata for creative ended for project %s", linkedin_int_setting[PROJECT_ID])

        if CREATIVES not in sync_info_with_type:
            log.warning("Backfilling creative metadata started for project %s", linkedin_int_setting[PROJECT_ID])
            enrich_metadata_previous_dates(CREATIVES, linkedin_int_setting[PROJECT_ID], linkedin_int_setting[LINKEDIN_AD_ACCOUNT], metadata, meta)
            log.warning("Backfilling creative metadata ended for project %s", linkedin_int_setting[PROJECT_ID])
    
    results, resp = get_insights(linkedin_int_setting, sync_info_with_type, CREATIVE_INSIGHTS, 'CREATIVE', request_counter)
    if resp['status'] == 'failed' or resp['errMsg'] != '':
        return resp
    
    insert_insights(CREATIVE_INSIGHTS,linkedin_int_setting[PROJECT_ID],linkedin_int_setting[LINKEDIN_AD_ACCOUNT], results, campaign_group_meta, campaign_meta, meta)
    return resp

def get_ad_account_data(linkedin_int_setting):
    url = 'https://api.linkedin.com/v2/adAccountsV2/{}'.format(linkedin_int_setting[LINKEDIN_AD_ACCOUNT])
    headers = {'Authorization': 'Bearer ' + linkedin_int_setting[ACCESS_TOKEN]}
    response = requests.get(url, headers=headers)
    if not response.ok:
        errString = 'Failed to get ad_accounts metadata from linkedin. StatusCode: {}. Error: {}. Project_id: {}'.format(response.status_code, response.text, linkedin_int_setting[PROJECT_ID])
        return {'status': 'failed', 'errMsg': errString, API_REQUESTS: 0}
    metadata = response.json()
    timestamp = int(datetime.now().strftime('%Y%m%d'))
    if end_timestamp != None:
        timestamp = end_timestamp
    add_linkedin_documents(linkedin_int_setting[PROJECT_ID], linkedin_int_setting[LINKEDIN_AD_ACCOUNT], AD_ACCOUNT, str(metadata['id']),metadata, timestamp)
    return {'status': 'success', 'errMsg': '', API_REQUESTS: 1}

def get_member_company_data(linkedin_int_setting, sync_info_with_type):
    log.warning("Fetching insights for member company started for project %s", linkedin_int_setting[PROJECT_ID])
    results, resp = get_insights(linkedin_int_setting, sync_info_with_type, MEMBER_COMPANY_INSIGHTS, 'MEMBER_COMPANY', 0)
    if resp['status'] == 'failed' or resp['errMsg'] != '':
        return resp
    
    return get_company_name_and_insert(MEMBER_COMPANY_INSIGHTS,linkedin_int_setting[PROJECT_ID],linkedin_int_setting[LINKEDIN_AD_ACCOUNT],linkedin_int_setting[ACCESS_TOKEN], results, resp[API_REQUESTS])

def get_collections(linkedin_int_setting, sync_info_with_type):
    response = {'status': 'success'}
    status = ''
    errMsgs = []
    skipMsgs = []
    campaign_group_meta = {}
    campaign_meta = {}
    creative_meta = {}
    requests_counter = 0

    try:
        if options.run_member_insights_only != 'True' and options.run_member_insights_only != True:
            # above if condition is for running member company insights for older date ranges, since we don't have doc_type segregation here
            if options.insert_metadata != 'False':
                res = get_ad_account_data(linkedin_int_setting)
                requests_counter += res[API_REQUESTS]
                if res['status'] == 'failed':
                    status = 'failed'
                    errMsgs.append(res['errMsg'])
            # don't mutate meta object, return it as a new object from get_campaign_group_function
            res = get_campaign_group_data(linkedin_int_setting, sync_info_with_type, campaign_group_meta)
            requests_counter += res[API_REQUESTS]
            if res['status'] == 'failed':
                status = 'failed'
                errMsgs.append(res['errMsg'])
            if res['status'] == 'skipped':
                skipMsgs.append(res['errMsg'])
            
            res = get_campaign_data(linkedin_int_setting, sync_info_with_type, campaign_group_meta, campaign_meta)
            requests_counter += res[API_REQUESTS]
            if res['status'] == 'failed':
                status = 'failed'
                errMsgs.append(res['errMsg'])
            if res['status'] == 'skipped':
                skipMsgs.append(res['errMsg'])
            
            res = get_creative_data(linkedin_int_setting, sync_info_with_type, campaign_group_meta, campaign_meta, creative_meta)
            requests_counter += res[API_REQUESTS]
            if res['status'] == 'failed':
                status = 'failed'
                errMsgs.append(res['errMsg'])
            if res['status'] == 'skipped':
                skipMsgs.append(res['errMsg'])
        
        if linkedin_int_setting[PROJECT_ID] in options.member_company_project_ids or options.member_company_project_ids == '*':
            res = get_member_company_data(linkedin_int_setting, sync_info_with_type)
            requests_counter += res[API_REQUESTS]
            if res['status'] == 'failed':
                status = 'failed'
                errMsgs.append(res['errMsg'])
            if res['status'] == 'skipped':
                skipMsgs.append(res['errMsg'])
        
    except Exception as e:
        traceback.print_tb(e.__traceback__)
        response['status'] = 'failed'
        response['msg'] = 'Failed with exception '+str(e)
        response[API_REQUESTS]= requests_counter
        return response
    if status == 'failed':
        response['status'] = 'failed'
        response['msg'] = errMsgs
        response[API_REQUESTS]= requests_counter
        return response
    response['status']= 'success'
    response[API_REQUESTS] = requests_counter
    response['msg'] = skipMsgs
    return response

def validate_or_generate_access_token_from_refresh_token(refresh_token, access_token):
    access_token_check_url = 'https://api.linkedin.com/v2/me?oauth2_access_token={}'.format(access_token)
    response = requests.get(access_token_check_url)
    if response.ok:
        return access_token, '', True

    url = 'https://www.linkedin.com/oauth/v2/accessToken?grant_type=refresh_token&refresh_token={}&client_id={}&client_secret={}'.format(refresh_token, options.client_id, options.client_secret)
    response = requests.get(url)
    response_json = response.json()
    if response.ok:
        return response_json['access_token'], '', False
    return '', response.text, False


def update_access_token(project_id, access_token):
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

def split_settings_for_multiple_ad_accounts(linkedin_settings):
    final_linkedin_settings = []
    failures = []
    count = 0 # for checking if any access token is updated, if yes timeout 10 mins
    for linkedin_int_setting in linkedin_settings:
        if linkedin_int_setting[LINKEDIN_AD_ACCOUNT] == '':
            failures.append({'status': 'failed', 'msg': 'empty ad account', PROJECT_ID: linkedin_int_setting[PROJECT_ID], AD_ACCOUNT: linkedin_int_setting[LINKEDIN_AD_ACCOUNT]})
            continue
        # validate access token
        linkedin_int_setting[ACCESS_TOKEN], err, is_prev_token_valid = validate_or_generate_access_token_from_refresh_token(linkedin_int_setting[REFRESH_TOKEN], linkedin_int_setting[ACCESS_TOKEN])
        if err == '':
            if not is_prev_token_valid:
                token_response = update_access_token(linkedin_int_setting[PROJECT_ID], linkedin_int_setting[ACCESS_TOKEN])
                if not token_response.ok:
                    failures.append({'status': 'failed', 'msg': 'failed to update access token', PROJECT_ID: linkedin_int_setting[PROJECT_ID], AD_ACCOUNT: linkedin_int_setting[LINKEDIN_AD_ACCOUNT]})
                    continue
                count += 1

            # spliting 1 setting into multiple for multiple ad accounts
            ad_accounts =  linkedin_int_setting[LINKEDIN_AD_ACCOUNT].split(',')
            for account_id in ad_accounts:
                new_setting = copy.deepcopy(linkedin_int_setting)
                new_setting[LINKEDIN_AD_ACCOUNT] = account_id
                final_linkedin_settings.append(new_setting)
        else:
            failures.append({'status': 'failed', 'msg': err, PROJECT_ID: linkedin_int_setting[PROJECT_ID], AD_ACCOUNT: linkedin_int_setting[LINKEDIN_AD_ACCOUNT]})
    
    if count > 0:
        time.sleep(600)
    
    return final_linkedin_settings, failures


if __name__ == '__main__':
    (options, args) = parser.parse_args()
    
    linkedin_int_settings =[]
    if options.project_ids != None and options.project_ids != '':
        linkedin_int_settings = get_linkedin_int_settings_for_projects(options.project_ids)
    else:
        linkedin_int_settings= get_linkedin_int_settings()
    start_timestamp = options.start_timestamp
    end_timestamp = options.end_timestamp


    if(linkedin_int_settings is not None):
        failures = []
        successes = []
        final_linkedin_settings, failures_sanitization = split_settings_for_multiple_ad_accounts(linkedin_int_settings)
        failures.extend(failures_sanitization)
        for linkedin_int_setting in final_linkedin_settings:
            if start_timestamp == None:
                sync_info_with_type, err = get_last_sync_info(linkedin_int_setting)
                if err != '':
                    response['status'] = 'failed'
                    response['msg'] = 'Failed to get last sync info'
                else:
                    response = get_collections(linkedin_int_setting, sync_info_with_type)
            else:
                response = get_collections(linkedin_int_setting, {})
            response[PROJECT_ID] = linkedin_int_setting[PROJECT_ID]
            response[AD_ACCOUNT] = linkedin_int_setting[LINKEDIN_AD_ACCOUNT]
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
        log.warning('Successfully synced. End of Linkedin sync job.')
        if len(failures) > 0:
            ping_healthcheck(options.env, HEALTHCHECK_PING_ID, notification_payload, endpoint='/fail')
        else:
            ping_healthcheck(options.env, HEALTHCHECK_PING_ID, notification_payload)
        sys.exit(0)