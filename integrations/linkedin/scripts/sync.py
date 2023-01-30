from optparse import OptionParser
import logging as log
import datetime
import requests
import copy
from datetime import datetime
import sys
import time
import traceback
from data_service import *
from constants import *
from util import *

parser = OptionParser()
parser.add_option('--env', dest='env', default='development')
parser.add_option('--dry', dest='dry', help='', default='False')
parser.add_option('--skip_today', dest='skip_today', help='', default='False') 
parser.add_option('--project_ids', dest='project_ids', help='', default=None, type=str)
parser.add_option('--exclude_project_ids', dest='exclude_project_ids', help='', default='', type=str)
parser.add_option('--client_id', dest='client_id', help='',default=None, type=str)
parser.add_option('--client_secret', dest='client_secret', help='',default=None, type=str)
parser.add_option('--start_timestamp', dest='start_timestamp', help='', default=None, type=int)
parser.add_option('--end_timestamp', dest='end_timestamp', help='', default=None, type=int)
parser.add_option('--insert_metadata', dest='insert_metadata', help='', default='True')
parser.add_option('--insert_report', dest='insert_report', help='', default='True')
parser.add_option('--data_service_host', dest='data_service_host',
    help='Data service host', default='http://localhost:8089')
parser.add_option('--run_member_insights_only', dest='run_member_insights_only', help='', default='False')


start_timestamp = None
end_timestamp = None

# can't keep very long range, we might hit rate limit
def get_insights(linkedin_int_setting, timestamp, doc_type, pivot, meta_request_count):
    log.warning("Fetching insights for {} started for project {} for timestamp {}".format(doc_type, linkedin_int_setting[PROJECT_ID], timestamp))
    
    start_year, start_month, start_day = get_split_date_from_timestamp(timestamp)
    end_year, end_month, end_day = get_split_date_from_timestamp(timestamp)

    request_counter = meta_request_count
    records = 0
    results =[]

    start = 0
    is_first_fetch = True
    while is_first_fetch or len(response.json()[ELEMENTS])>=INSIGHTS_COUNT:
        is_first_fetch = False
        url = INSIGHTS_REQUEST_URL_FORMAT.format(pivot, start_day, start_month, start_year, end_day, end_month, end_year,
        REQUESTED_FIELDS, linkedin_int_setting[LINKEDIN_AD_ACCOUNT], start, INSIGHTS_COUNT)
        
        headers = {'Authorization': 'Bearer ' + linkedin_int_setting[ACCESS_TOKEN]}
        response = requests.get(url, headers=headers)
        request_counter += 1
        if not response.ok:
            errString = API_ERROR_FORMAT.format(pivot, 'insights', response.status_code, response.text, linkedin_int_setting[PROJECT_ID])
            log.error(errString)
            return [], {'status': 'failed', 'errMsg': errString, API_REQUESTS: request_counter}
        if ELEMENTS in response.json():
            records += len(response.json()[ELEMENTS])
            results.extend(response.json()[ELEMENTS])
        start += INSIGHTS_COUNT

    log.warning("No of %s insights records to be inserted for project %s : %d", doc_type, linkedin_int_setting[PROJECT_ID], records)
    return results, {'status': 'success', 'errMsg': '', API_REQUESTS: request_counter}


def insert_insights(options, doc_type, project_id, ad_account, response, timestamp):
    log.warning(INSERTION_LOG.format(doc_type, 'insights', timestamp))
    if len(response) > 0:
        add_documents_response = add_all_linkedin_documents(project_id, ad_account, doc_type, response, timestamp, options)
        if not add_documents_response.ok and add_documents_response.status_code != 409:
            errString = DOC_INSERT_ERROR.format(doc_type, 'insights',add_documents_response.status, add_documents_response.text, project_id, ad_account)
            log.error(errString)
            return errString
    log.warning(INSERTION_END_LOG.format(doc_type, 'insights', timestamp))
    return ''

def get_org_data_from_linkedin_with_retries(idStrArray, access_token):
    map_id_to_org_data = {}
    index = 0
    retry_failed_IDs = True
    request_counter = 0

    while index < len(idStrArray):
        url = 'https://api.linkedin.com/v2/organizationsLookup?ids=List({})'.format(idStrArray[index])
        headers = {'Authorization': 'Bearer ' + access_token, 'X-Restli-Protocol-Version': '2.0.0'}
        response = requests.get(url, headers=headers)
        request_counter += 1

        if not response.ok or 'results' not in response.json():
            return {}, request_counter, "Failed getting organisation data with error {}".format(response.text)
        map_id_to_org_data.update(response.json()['results'])

        failedIDs = ""
        idArray = idStrArray[index].split(",")
        for id in idArray:
            if id not in map_id_to_org_data:
                if failedIDs == '':
                    failedIDs += id
                else:
                    failedIDs += (',' + id)
        
        # if failedIDs still exist and we have already retried for the same then move to the next index
        if failedIDs == "" or (not retry_failed_IDs):
            index +=1
            retry_failed_IDs = True # setting it as true for the next set of IDs we can retry again in case of failures
        else:
            idStrArray[index] = failedIDs
            retry_failed_IDs = False #setting as false so that we don't retry again

    return map_id_to_org_data, request_counter, ""
            

# it take organization id for report rows and fetches org details like name location domain and append it to the member company report rows
def update_org_data(records, access_token):
    mapIDs = {}
    request_counter = 0
    for data in records:
        id = data['pivotValue'].split(':')[3]
        mapIDs[id]= True
    idStr = ''
    idCount = 0
    idStrArray = []
    for key in mapIDs:
        idCount += 1
        if idStr == '':
            idStr += key
        else:
            idStr += (',' + key)
        if idCount >=500:
            idStrArray.append(idStr)
            idStr = ""
            idCount = 0
    
    if idStr != "":
        idStrArray.append(idStr)

    map_id_to_org_data, request_counter, errString = get_org_data_from_linkedin_with_retries(idStrArray, access_token)
    if errString != "":
        return [], request_counter, errString

    failedIDs = ""
    for data in records:
        id = data['pivotValue'].split(':')[3]
        data['id'] = id
        if id not in map_id_to_org_data:
            if failedIDs == "":
                failedIDs += id
            else:
                failedIDs += (',' + id)
            data['vanityName'] = '$none'
            data['localizedName'] = '$none'
            data['localizedWebsite'] = '$none'
            data['preferredCountry'] = '$none'
            data['companyHeadquarters'] = '$none'
       
        else:
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

    if failedIDs != "":
        return records, request_counter, "Failed getting organisation data for ids {}".format(failedIDs)
    return records, request_counter, ''

def get_company_name_and_insert(options, doc_type, project_id, ad_account, access_token, response, request_counter, timestamp):
    if len(response) !=0:
        records, requests, errString = update_org_data(response, access_token)
        # in case where records are returned, we insert them in the db and then ping healthcheck with a failure for loggging purpses 
        if errString != '' and len(records) == 0:
            return {'status': 'failed', 'errMsg': errString, API_REQUESTS: request_counter + requests}
    else:
        log.warning('No data found for member company insights for project {} and Ad account {}'.format(project_id, ad_account))
    
    insert_err = insert_insights(options, doc_type, project_id, ad_account, records, timestamp)
    if insert_err != '':
        return {'status': 'failed', 'errMsg': insert_err, API_REQUESTS: request_counter + requests}
    # we are allowing sync for companies which we were not able to get the metadata for but also notifying in healthchecks with failed orgIDS 
    if errString != '':
        log.error(errString)
    return {'status': 'success', 'errMsg': '', API_REQUESTS: request_counter + requests}

def get_metadata(ad_account, access_token, url_endpoint, doc_type, project_id):
    metadata = []
    request_counter = 0
    is_first_fetch = True
    response = {}

    start = 0
    while is_first_fetch or len(response.json()[ELEMENTS])>=META_COUNT:
        is_first_fetch = False
        url = META_DATA_URL.format(url_endpoint, ad_account, start, META_COUNT)
        headers = {'Authorization': 'Bearer ' + access_token}
        response = requests.get(url, headers=headers)
        request_counter += 1
        if not response.ok:
            errString = API_ERROR_FORMAT.format(doc_type, 'metadata', response.status_code, response.text, project_id)
            return metadata, errString, request_counter
        metadata.extend(response.json()[ELEMENTS])
        start +=META_COUNT
    return metadata, '', request_counter

def insert_metadata(options, doc_type, project_id, ad_account, response, timestamp, extraMeta):
    log.warning(INSERTION_LOG.format(doc_type, 'metadata', timestamp))
    for data in response:
        data.update(extraMeta[str(data['id'])])
    add_documents_response = add_all_linkedin_documents(project_id, ad_account, doc_type, response, timestamp, options)
    return add_documents_response

def get_campaign_group_data(options, linkedin_int_setting, sync_info_with_type, meta):
    log.warning("Fetching metadata for campaign group started for project %s", linkedin_int_setting[PROJECT_ID])
    metadata, errString, request_counter = get_metadata(linkedin_int_setting[LINKEDIN_AD_ACCOUNT], linkedin_int_setting[ACCESS_TOKEN], 'adCampaignGroupsV2', CAMPAIGN_GROUPS, linkedin_int_setting[PROJECT_ID])
    if errString != '':
        return {'status': 'failed', 'errMsg': errString, API_REQUESTS: request_counter}
    for data in metadata:
        meta[str(data['id'])] = {CAMPAIGN_GROUP_ID: str(data['id']), 'campaign_group_name': data['name'], 'campaign_group_status': data['status']}
    
    if options.insert_metadata != 'False' and len(metadata) > 0:
        log.warning("No of metadata records for campaign group to be inserted for project %s : %d", linkedin_int_setting[PROJECT_ID], len(metadata))
        timestamp_range_for_meta = get_timestamp_range(CAMPAIGN_GROUPS, sync_info_with_type, start_timestamp, end_timestamp)
        for timestamp in timestamp_range_for_meta:
            insert_response = insert_metadata(options, CAMPAIGN_GROUPS, linkedin_int_setting[PROJECT_ID], linkedin_int_setting[LINKEDIN_AD_ACCOUNT], metadata, timestamp, meta)
            if not insert_response.ok and insert_response.status != 409:
                errString = DOC_INSERT_ERROR.format(CAMPAIGN_GROUPS, "metadata", insert_response.status, insert_response.text, linkedin_int_setting[PROJECT_ID], linkedin_int_setting[LINKEDIN_AD_ACCOUNT])
                log.error(errString)
                return  {'status': 'failed', 'errMsg': errString, API_REQUESTS: request_counter}
        log.warning("Insertion metadata for campaign group ended for project %s", linkedin_int_setting[PROJECT_ID])

    timestamp_range_for_insights = get_timestamp_range(CAMPAIGN_GROUP_INSIGHTS, sync_info_with_type, start_timestamp, end_timestamp)
    for timestamp in timestamp_range_for_insights:
        results, resp = get_insights(linkedin_int_setting, timestamp, CAMPAIGN_GROUP_INSIGHTS, 'CAMPAIGN_GROUP', request_counter)
        if resp['status'] == 'failed' or resp['errMsg'] != '':
            return resp
        request_counter = resp[API_REQUESTS]
        results = update_result_with_metadata(results, CAMPAIGN_GROUP_INSIGHTS, meta, {}, {})
            
        errString = insert_insights(options, CAMPAIGN_GROUP_INSIGHTS,linkedin_int_setting[PROJECT_ID],linkedin_int_setting[LINKEDIN_AD_ACCOUNT], results, timestamp)
        if errString != '':
            return {'status': 'failed', 'errMsg': errString, API_REQUESTS: request_counter}
    return {'status': 'success', 'errMsg': '', API_REQUESTS: request_counter}

def get_campaign_data(options, linkedin_int_setting, sync_info_with_type , campaign_group_meta, meta):
    log.warning("Fetching metadata for campaign started for project %s", linkedin_int_setting[PROJECT_ID])
    metadata, errString, request_counter = get_metadata(linkedin_int_setting[LINKEDIN_AD_ACCOUNT], linkedin_int_setting[ACCESS_TOKEN], 'adCampaignsV2', CAMPAIGNS, linkedin_int_setting[PROJECT_ID])
    if errString != '':
        return {'status': 'failed', 'errMsg': errString, API_REQUESTS: request_counter}
    for data in metadata:
        campaign_group_id = str(data['campaignGroup'].split(':')[3])
        meta[str(data['id'])] = {'campaign_group_id': campaign_group_id,'campaign_id': str(data['id']), 'campaign_name': data['name'], 'campaign_status': data['status'], 'campaign_type': data['type']}

    
    if options.insert_metadata != 'False' and len(metadata) > 0:
        log.warning("No of metadata records for campaigns to be inserted for project %s : %d", linkedin_int_setting[PROJECT_ID], len(metadata))
        timestamp_range_for_meta = get_timestamp_range(CAMPAIGNS, sync_info_with_type, start_timestamp, end_timestamp)
        
        for timestamp in timestamp_range_for_meta:
            insert_response = insert_metadata(options, CAMPAIGNS, linkedin_int_setting[PROJECT_ID], linkedin_int_setting[LINKEDIN_AD_ACCOUNT], metadata, timestamp, meta)
            if not insert_response.ok and insert_response.status != 409:
                errString = DOC_INSERT_ERROR.format(CAMPAIGNS, "metadata", insert_response.status, insert_response.text, linkedin_int_setting[PROJECT_ID], linkedin_int_setting[LINKEDIN_AD_ACCOUNT])
                log.error(errString)
                return  {'status': 'failed', 'errMsg': errString, API_REQUESTS: request_counter}
        log.warning("Insertion metadata for campaign ended for project %s", linkedin_int_setting[PROJECT_ID])

    timestamp_range_for_insights = get_timestamp_range(CAMPAIGN_INSIGHTS, sync_info_with_type, start_timestamp, end_timestamp)
    for timestamp in timestamp_range_for_insights:
        results, resp = get_insights(linkedin_int_setting, timestamp, CAMPAIGN_INSIGHTS, 'CAMPAIGN', request_counter)
        if resp['status'] == 'failed' or resp['errMsg'] != '':
            return resp
        request_counter = resp[API_REQUESTS]
        results = update_result_with_metadata(results, CAMPAIGN_INSIGHTS, campaign_group_meta, meta, {})
            
        errString = insert_insights(options, CAMPAIGN_INSIGHTS,linkedin_int_setting[PROJECT_ID],linkedin_int_setting[LINKEDIN_AD_ACCOUNT], results, timestamp)
        if errString != '':
            return {'status': 'failed', 'errMsg': errString, API_REQUESTS: request_counter}
    return {'status': 'success', 'errMsg': '', API_REQUESTS: request_counter}

def get_creative_data(options, linkedin_int_setting, sync_info_with_type, campaign_group_meta, campaign_meta, meta):
    log.warning("Fetching metadata for creative started for project %s", linkedin_int_setting[PROJECT_ID])
    metadata, errString, request_counter = get_metadata(linkedin_int_setting[LINKEDIN_AD_ACCOUNT], linkedin_int_setting[ACCESS_TOKEN], 'adCreativesV2', CREATIVES, linkedin_int_setting[PROJECT_ID])
    if errString != '':
        return {'status': 'failed', 'errMsg': errString, API_REQUESTS: request_counter}
    for data in metadata:
        campaign_id = str(data['campaign'].split(':')[3])
        campaign_group_id = campaign_meta[campaign_id][CAMPAIGN_GROUP_ID]
        meta[str(data['id'])] = {'campaign_group_id': campaign_group_id, 'campaign_id': campaign_id ,'creative_id': str(data['id']), 'creative_status': data['status'], 'creative_type': data['type']}
   
    if options.insert_metadata != 'False' and len(metadata) > 0:
        log.warning("No of metadata records for campaigns to be inserted for project %s : %d", linkedin_int_setting[PROJECT_ID], len(metadata))
        timestamp_range_for_meta = get_timestamp_range(CREATIVES, sync_info_with_type, start_timestamp, end_timestamp)
        
        for timestamp in timestamp_range_for_meta:
            insert_response = insert_metadata(options, CREATIVES, linkedin_int_setting[PROJECT_ID], linkedin_int_setting[LINKEDIN_AD_ACCOUNT], metadata, timestamp, meta)
            if not insert_response.ok and insert_response.status != 409:
                errString = DOC_INSERT_ERROR.format(CREATIVES, "metadata", insert_response.status, insert_response.text, linkedin_int_setting[PROJECT_ID], linkedin_int_setting[LINKEDIN_AD_ACCOUNT])
                log.error(errString)
                return  {'status': 'failed', 'errMsg': errString, API_REQUESTS: request_counter}
        log.warning("Insertion metadata for creative ended for project %s", linkedin_int_setting[PROJECT_ID])
    
    timestamp_range_for_insights = get_timestamp_range(CREATIVE_INSIGHTS, sync_info_with_type, start_timestamp, end_timestamp)
    for timestamp in timestamp_range_for_insights:
        results, resp = get_insights(linkedin_int_setting, timestamp, CREATIVE_INSIGHTS, 'CREATIVE', request_counter)
        if resp['status'] == 'failed' or resp['errMsg'] != '':
            return resp
        request_counter = resp[API_REQUESTS]
        results = update_result_with_metadata(results, CREATIVE_INSIGHTS, campaign_group_meta, campaign_meta, meta)
            
        errString = insert_insights(options, CREATIVE_INSIGHTS,linkedin_int_setting[PROJECT_ID],linkedin_int_setting[LINKEDIN_AD_ACCOUNT], results, timestamp)
        if errString != '':
            return {'status': 'failed', 'errMsg': errString, API_REQUESTS: request_counter}
    return {'status': 'success', 'errMsg': '', API_REQUESTS: request_counter}

def get_ad_account_data(options, linkedin_int_setting):
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
    response = add_linkedin_documents(linkedin_int_setting[PROJECT_ID], linkedin_int_setting[LINKEDIN_AD_ACCOUNT], AD_ACCOUNT, str(metadata['id']),metadata, timestamp, options)
    if not response.ok and response.status_code != 409:
        return {'status': 'failed', 'errMsg': 'Failed inserting add accounts data', API_REQUESTS: 1}
    return {'status': 'success', 'errMsg': '', API_REQUESTS: 1}

def get_member_company_data(options, linkedin_int_setting, sync_info_with_type):
    log.warning("Fetching insights for member company started for project %s", linkedin_int_setting[PROJECT_ID])
    timestamp_range_for_insights = get_timestamp_range(MEMBER_COMPANY_INSIGHTS, sync_info_with_type, start_timestamp, end_timestamp)
    request_counter = 0
    for timestamp in timestamp_range_for_insights:
        results, resp = get_insights(linkedin_int_setting, timestamp, MEMBER_COMPANY_INSIGHTS, 'MEMBER_COMPANY', request_counter)
        if resp['status'] == 'failed' or resp['errMsg'] != '':
            return resp
        request_counter = resp[API_REQUESTS]
        if len(results) == 0:
            continue
        
        resp = get_company_name_and_insert(options, MEMBER_COMPANY_INSIGHTS,linkedin_int_setting[PROJECT_ID],linkedin_int_setting[LINKEDIN_AD_ACCOUNT],linkedin_int_setting[ACCESS_TOKEN], results, resp[API_REQUESTS], timestamp)
        if resp['status'] == 'failed' or resp['errMsg'] != '':
            return resp
        request_counter = resp[API_REQUESTS]
    return {'status': 'success', 'errMsg': '', API_REQUESTS: request_counter}

def get_collections(options, linkedin_int_setting, sync_info_with_type):
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
                res = get_ad_account_data(options, linkedin_int_setting)
                requests_counter += res[API_REQUESTS]
                if res['status'] == 'failed':
                    status = 'failed'
                    errMsgs.append(res['errMsg'])
            # don't mutate meta object, return it as a new object from get_campaign_group_function
            res = get_campaign_group_data(options, linkedin_int_setting, sync_info_with_type, campaign_group_meta)
            requests_counter += res[API_REQUESTS]
            if res['status'] == 'failed':
                status = 'failed'
                errMsgs.append(res['errMsg'])
            if res['status'] == 'skipped':
                skipMsgs.append(res['errMsg'])
            
            res = get_campaign_data(options, linkedin_int_setting, sync_info_with_type, campaign_group_meta, campaign_meta)
            requests_counter += res[API_REQUESTS]
            if res['status'] == 'failed':
                status = 'failed'
                errMsgs.append(res['errMsg'])
            if res['status'] == 'skipped':
                skipMsgs.append(res['errMsg'])
            
            res = get_creative_data(options, linkedin_int_setting, sync_info_with_type, campaign_group_meta, campaign_meta, creative_meta)
            requests_counter += res[API_REQUESTS]
            if res['status'] == 'failed':
                status = 'failed'
                errMsgs.append(res['errMsg'])
            if res['status'] == 'skipped':
                skipMsgs.append(res['errMsg'])
        
        res = get_member_company_data(options, linkedin_int_setting, sync_info_with_type)
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

def split_settings_for_multiple_ad_accounts(options, linkedin_settings):
    final_linkedin_settings = []
    failures = []
    count = 0 # for checking if any access token is updated, if yes timeout 10 mins
    for linkedin_int_setting in linkedin_settings:
        if linkedin_int_setting[PROJECT_ID] in options.exclude_project_ids:
            continue
        if linkedin_int_setting[LINKEDIN_AD_ACCOUNT] == '':
            failures.append({'status': 'failed', 'msg': 'empty ad account', PROJECT_ID: linkedin_int_setting[PROJECT_ID], AD_ACCOUNT: linkedin_int_setting[LINKEDIN_AD_ACCOUNT]})
            continue
        # validate access token
        linkedin_int_setting[ACCESS_TOKEN], err, is_prev_token_valid = validate_or_generate_access_token_from_refresh_token(linkedin_int_setting[REFRESH_TOKEN], linkedin_int_setting[ACCESS_TOKEN])
        if err == '':
            if not is_prev_token_valid:
                token_response = update_access_token(linkedin_int_setting[PROJECT_ID], linkedin_int_setting[ACCESS_TOKEN], options)
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
        linkedin_int_settings = get_linkedin_int_settings_for_projects(options)
    else:
        linkedin_int_settings= get_linkedin_int_settings(options)
    start_timestamp = options.start_timestamp
    end_timestamp = options.end_timestamp


    if(linkedin_int_settings is not None):
        failures = []
        successes = []
        token_failures = []
        final_linkedin_settings, failures_sanitization = split_settings_for_multiple_ad_accounts(options, linkedin_int_settings)
        token_failures.extend(failures_sanitization)
        for linkedin_int_setting in final_linkedin_settings:
            if start_timestamp == None:
                sync_info_with_type, err = get_last_sync_info(linkedin_int_setting, options)
                if err != '':
                    response['status'] = 'failed'
                    response['msg'] = 'Failed to get last sync info'
                else:
                    response = get_collections(options, linkedin_int_setting, sync_info_with_type)
            else:
                response = get_collections(options, linkedin_int_setting, {})
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
        if len(token_failures) > 0:
            notification_payload = {
                'status': 'Token failures', 
                'failures': token_failures,
            }
            ping_healthcheck(options.env, HEALTHCHECK_TOKEN_FAILURE_PING_ID, notification_payload, endpoint='/fail')
        sys.exit(0)