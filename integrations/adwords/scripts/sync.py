from googleads import adwords
from googleads import oauth2
from optparse import OptionParser
import json
import logging as log
import csv
import requests
import datetime
import re
import sys
import time

parser = OptionParser()
parser.add_option("--env", dest="env", default="development")
parser.add_option("--dry", dest="dry", help="", default="False")
parser.add_option("--skip_today", dest="skip_today", help="", default="False")
parser.add_option("--developer_token", dest="developer_token", help="", default="") 
parser.add_option("--oauth_secret", dest="oauth_secret", help="", default="")
parser.add_option("--project_id", dest="project_id", help="", default=None, type=int)
parser.add_option("--data_service_host", dest="data_service_host",
    help="Data service host", default="http://localhost:8089")

ADWORDS_CLIENT_USER_AGENT = "FactorsAI (https://www.factors.ai)"
PAGE_SIZE = 200

APP_NAME = "adwords_sync"
STATUS_FAILED = "failed"
STATUS_SKIPPED = "skipped"

# Cache permission denied customer_acc_id + token and avoid 
# sync for similar requests.
# customer_acc_id:refresh_token -> user_permission_denied.
permission_error_cache = {}
request_stat = {}

class OAuthManager():
    _client_id = None
    _client_secret = None

    @classmethod
    def init(cls, secret):
        cls._secret = secret
        # throws KeyError.
        cls._client_id = secret["web"]["client_id"]
        cls._client_secret = secret["web"]["client_secret"]

    @classmethod
    def get_client_secret(cls):
        return OAuthManager._client_secret
    
    @classmethod
    def get_client_id(cls):
        return OAuthManager._client_id

def notify(env, source, message):
    if env != "production": 
        log.warning("Skipped notification for env %s payload %s", env, str(message)) 
        return

    sns_url = "https://fjnvg9a8wi.execute-api.us-east-1.amazonaws.com/v1/notify"
    payload = { "env": env, "message": message, "source": source }
    response = requests.post(sns_url, json=payload)
    if not response.ok: log.error("Failed to notify through sns.")
    return response

def first_letter_to_lower(s):
    if len(s) == 0: return ''

    f = s[0].lower()
    if len(s) == 1: return f

    return f + s[1:]

def is_valid_value_type(s):
    return isinstance(s, str) or isinstance(s, int) or isinstance(s, float) or isinstance(s, bool)

def snake_to_pascal_case(fields):
    pascals = []
    for f in fields:
        p = ''.join(x.capitalize() or '_' for x in f.split('_'))
        pascals.append(p)

    return pascals

def camel_case_to_snake_case(s):
    s1 = re.sub('(.)([A-Z][a-z]+)', r'\1_\2', s)
    return re.sub('([a-z0-9])([A-Z])', r'\1_\2', s1).lower()

def csv_to_dict_list(headers, csv_list):
    resp_rows = []

    rows = csv.reader(csv_list)
    for row in rows:
        resp = {}
        i = 0

        for col in row:
            col_striped = col.strip()
            if col_striped != '--':
                resp[headers[i]] = col_striped
            i = i + 1
        
        if len(resp) > 0:
            resp_rows.append(resp)
    
    return resp_rows
        

def get_customer_account_properties(adwords_client, customer_account_id, timestmap):
    customer_service = adwords_client.GetService('CustomerService', version='v201809')
    customer_accounts = customer_service.getCustomers()

    current_account = None
    for account in customer_accounts:
        if str(account["customerId"]) == customer_account_id:
            current_account = account

    if current_account is None:
        log.error("Customer account not found on list of accounts. Failed to get properties.")
        raise Exception("Failed to get properties. customer account"+str(customer_account_id)+" not found on list of account "+str(customer_accounts))

    properties = {}
    try:
        properties["customer_id"] = current_account["customerId"]
        properties["currency_code"] = current_account["currencyCode"]
        properties["date_timezone"] = current_account["dateTimeZone"]
        properties["can_manage_clients"] = current_account["canManageClients"]
        properties["test_account"] = current_account["testAccount"]
    except Exception as e:
        log.error("Failed to get customer account properties: %s", str(e))
        return [properties]

    return [properties], 1


def get_campaigns(adwords_client, timestamp):
    if adwords_client == None:
        raise Exception('no adwords client')

    # Initialize appropriate service.
    campaign_service = adwords_client.GetService('CampaignService', version='v201809')

    fields = ["Id", "CampaignGroupId", "Name", "Status", "ServingStatus", "StartDate", "EndDate", 
    "AdServingOptimizationStatus", "Settings", "AdvertisingChannelType", "AdvertisingChannelSubType", 
    "Labels", "CampaignTrialType", "BaseCampaignId", "TrackingUrlTemplate", "FinalUrlSuffix", 
    "UrlCustomParameters", "SelectiveOptimization"]
    offset = 0
    selector = {
        'fields': fields,
        'paging': {
            'startIndex': str(offset),
            'numberResults': str(PAGE_SIZE)
        }
    }

    more_pages = True
    rows = []
    requests = 0
    while more_pages:
        page = campaign_service.get(selector)

        # Display results.
        if 'entries' in page:
            for campaign in page['entries']:
                doc = {}
                for field in fields:
                    fieldName = first_letter_to_lower(field)
                    if is_valid_value_type(campaign[fieldName]):
                        doc[camel_case_to_snake_case(fieldName)] = campaign[fieldName]
                rows.append(doc)
        else:
            log.warning('No campaigns were found.')
        offset += PAGE_SIZE
        selector['paging']['startIndex'] = str(offset)
        more_pages = offset < int(page['totalNumEntries'])
        requests = requests + 1

    return rows, requests

def get_ads(adwords_client, timestamp):
    if adwords_client == None:
        raise Exception('no adwords client')

    # Initialize appropriate service.
    ad_group_ad_service = adwords_client.GetService('AdGroupAdService', version='v201809')

    fields = ["AdGroupId", "Status", "BaseCampaignId", "BaseAdGroupId"]
    offset = 0
    selector = {
        'fields': fields,
        'paging': {
            'startIndex': str(offset),
            'numberResults': str(PAGE_SIZE)
        }
    }

    more_pages = True
    rows = []
    requests = 0
    while more_pages:
        page = ad_group_ad_service.get(selector)

        # Display results.
        if 'entries' in page:
            for ad_entry in page['entries']:
                doc = {}
                for field in fields:
                    fieldName = first_letter_to_lower(field)
                    if is_valid_value_type(ad_entry[fieldName]):
                        doc[camel_case_to_snake_case(fieldName)] = ad_entry[fieldName]
                    
                    # Add values form ad object.
                    if ad_entry['ad'] != None:
                        for field in ad_entry['ad']:
                            if is_valid_value_type(ad_entry['ad'][field]):
                                doc[camel_case_to_snake_case(field)] = ad_entry['ad'][field]

                rows.append(doc)
        else:
            log.warning('No ads were found.')
        offset += PAGE_SIZE
        selector['paging']['startIndex'] = str(offset)
        more_pages = offset < int(page['totalNumEntries'])
        requests = requests + 1

    return rows, requests


def get_ad_groups(adwords_client, timestamp):
    if adwords_client == None:
        raise Exception('no adwords client')

    # Initialize appropriate service.
    ad_group_service = adwords_client.GetService('AdGroupService', version='v201809')

    fields = ["Id", "CampaignId", "CampaignName", "Name", "Status", "Settings", "Labels", 
    "ContentBidCriterionTypeGroup", "BaseCampaignId", "BaseAdGroupId", "AdGroupType"]
    offset = 0
    selector = {
        'fields': fields,
        'paging': {
            'startIndex': str(offset),
            'numberResults': str(PAGE_SIZE)
        }
    }

    more_pages = True
    rows = []
    requests = 0
    while more_pages:
        page = ad_group_service.get(selector)

        # Display results.
        if 'entries' in page:
            for ad_group in page['entries']:
                doc = {}
                for field in fields:
                    fieldName = first_letter_to_lower(field)
                    if is_valid_value_type(ad_group[fieldName]):
                        doc[camel_case_to_snake_case(fieldName)] = ad_group[fieldName]
                rows.append(doc)
        else:
            log.warning('No ad_groups were found.')
        offset += PAGE_SIZE
        selector['paging']['startIndex'] = str(offset)
        more_pages = offset < int(page['totalNumEntries'])
        requests = requests + 1

    return rows, requests


def get_click_performance_report(adwords_client, timestamp):
    if adwords_client == None:
        raise Exception('no adwords client')

    if (timestamp == None or timestamp == ""):
        raise Exception("invalid date string for report download")
    
    str_timestamp = str(timestamp)
    during = str_timestamp + "," + str_timestamp
    downloader = adwords_client.GetReportDownloader(version='v201809')

    query_fields = ['ad_format', 'ad_group_id', 'ad_network_type_1', 'ad_network_type_2', 
    'aoi_most_specific_target_id', 'campaign_id', 'click_type', 'creative_id', 'criteria_parameters', 
    'date', 'device', 'gcl_id', 'page', 'slot', 'user_list_id'] 
    fields = snake_to_pascal_case(query_fields)
    
    # Create report query.
    report_query = (adwords.ReportQueryBuilder()
        .Select(*fields)
        .From('CLICK_PERFORMANCE_REPORT')
        .Where('CampaignStatus').In('ENABLED', 'PAUSED')
        .During(during).Build())

    report = downloader.DownloadReportAsStringWithAwql(
        report_query, 'CSV', skip_report_header=True, 
        skip_column_header=True, skip_report_summary=True)

    lines = report.split('\n')
    return csv_to_dict_list(query_fields, lines), 1



def get_campaign_performance_report(adwords_client, timestamp):
    if adwords_client == None:
            raise Exception('no adwords client')

    if (timestamp == None or timestamp == ""):
        raise Exception("invalid date string for report download")
    
    str_timestamp = str(timestamp)
    during = str_timestamp + "," + str_timestamp
    downloader = adwords_client.GetReportDownloader(version='v201809')
    
    query_fields = ['active_view_impressions', 'active_view_measurability', 'active_view_measurable_cost', 
    'active_view_measurable_impressions', 'active_view_viewability', 'advertising_channel_sub_type', 'all_conversion_rate', 'conversion_rate', 'cost_per_conversion', 
    'all_conversion_value', 'all_conversions', 'amount', 'average_cost', 'average_position', 'average_time_on_site', 
    'base_campaign_id', 'bounce_rate', 'budget_id', 'campaign_id', 'campaign_name', 'campaign_status', 'campaign_trial_type', 'click_assisted_conversion_value', 
    'click_assisted_conversions', 'click_assisted_conversions_over_last_click_conversions', 'clicks', 'conversion_value', 'conversions', 
    'cost', 'start_date', 'end_date', 'engagements', 'gmail_forwards', 'gmail_saves', 'gmail_secondary_clicks', 'impression_assisted_conversions', 
    'impression_reach', 'impressions', 'interaction_types', 'interactions', 'invalid_clicks', 'is_budget_explicitly_shared', 'url_custom_parameters', 
    'value_per_all_conversion', 'video_quartile_100_rate', 'video_quartile_25_rate', 'video_quartile_50_rate', 'video_quartile_75_rate',
    'video_view_rate', 'video_views', 'view_through_conversions']
    fields = snake_to_pascal_case(query_fields)

    # Create report query.
    report_query = (adwords.ReportQueryBuilder()
        .Select(*fields)
        .From('CAMPAIGN_PERFORMANCE_REPORT')
        .Where('CampaignStatus').In('ENABLED', 'PAUSED')
        .During(during).Build())

    report = downloader.DownloadReportAsStringWithAwql(report_query, 'CSV', 
        skip_report_header=True, skip_column_header=True, skip_report_summary=True)
    
    lines = report.split('\n')
    return csv_to_dict_list(query_fields, lines), 1

def get_ad_performance_report(adwords_client, timestamp):
    if adwords_client == None:
            raise Exception('no adwords client')

    if (timestamp == None or timestamp == ""):
        raise Exception("invalid date string for report download")
    
    str_timestamp = str(timestamp)
    during = str_timestamp + "," + str_timestamp
    downloader = adwords_client.GetReportDownloader(version='v201809')
    
    query_fields = ['id', 'account_currency_code', 'account_descriptive_name', 'active_view_impressions', 
        'active_view_measurability', 'active_view_measurable_cost', 'active_view_measurable_impressions',
        'active_view_viewability', 'ad_group_id', 'all_conversion_rate', 'conversion_rate', 'cost_per_conversion', 
        'all_conversion_value', 'all_conversions', 'average_cost', 'average_position', 'average_time_on_site',
        'bounce_rate', 'click_assisted_conversion_value', 'click_assisted_conversions', 
        'click_assisted_conversions_over_last_click_conversions',
        'clicks', 'conversion_value', 'conversions', 'cost', 'date', 'engagements', 'gmail_forwards',
        'gmail_saves', 'gmail_secondary_clicks', 'impression_assisted_conversions', 'impressions', 'interaction_types',
        'interactions', 'value_per_all_conversion', 'video_quartile_100_rate', 'video_quartile_25_rate', 
        'video_quartile_50_rate', 'video_quartile_75_rate', 'video_view_rate', 'video_views', 'view_through_conversions']
    fields = snake_to_pascal_case(query_fields)

    # Create report query.
    report_query = (adwords.ReportQueryBuilder()
        .Select(*fields)
        .From('AD_PERFORMANCE_REPORT')
        .Where('CampaignStatus').In('ENABLED', 'PAUSED')
        .During(during).Build())

    report = downloader.DownloadReportAsStringWithAwql(report_query, 'CSV', 
        skip_report_header=True, skip_column_header=True, skip_report_summary=True)
    
    lines = report.split('\n')
    return csv_to_dict_list(query_fields, lines), 1


def get_search_performance_report(adwords_client, timestamp):
    if adwords_client == None:
            raise Exception('no adwords client')

    if (timestamp == None or timestamp == ""):
        raise Exception("invalid date string for report download")
    
    str_timestamp = str(timestamp)
    during = str_timestamp + "," + str_timestamp
    downloader = adwords_client.GetReportDownloader(version='v201809')

    query_fields = ['ad_group_id', 'ad_group_name', 'all_conversion_rate', 'conversion_rate', 'cost_per_conversion',
    'all_conversion_value', 'all_conversions', 'average_cost', 'average_cpc', 'average_cpe', 
    'average_cpm', 'average_cpv', 'average_position', 'campaign_id', 'clicks', 'conversion_value', 'conversions',
    'cost', 'cost_per_all_conversion', 'cross_device_conversions', 'ctr', 'date', 
    'device', 'engagement_rate', 'engagements', 'external_customer_id',
    'final_url', 'impressions', 'interaction_rate', 'interaction_types', 'interactions', 'keyword_id', 
    'query', 'query_match_type_with_variant', 'tracking_url_template', 'value_per_all_conversion', 
    'value_per_conversion', 'video_quartile_100_rate', 'video_quartile_25_rate', 'video_quartile_50_rate', 
    'video_quartile_75_rate', 'video_view_rate', 'video_views', 'view_through_conversions', 'week', 'year']
    fields = snake_to_pascal_case(query_fields)
    
    # Create report query.
    report_query = (adwords.ReportQueryBuilder()
        .Select(*fields)
        .From('SEARCH_QUERY_PERFORMANCE_REPORT')
        .Where('CampaignStatus').In('ENABLED', 'PAUSED')
        .During(during).Build())

    report = downloader.DownloadReportAsStringWithAwql(
        report_query, 'CSV', skip_report_header=True, 
        skip_column_header=True, skip_report_summary=True)

    lines = report.split('\n')
    return csv_to_dict_list(query_fields, lines), 1


def get_keywords_performance_report(adwords_client, timestamp):
    if adwords_client == None:
            raise Exception('no adwords client')

    if (timestamp == None or timestamp == ""):
        raise Exception("invalid date string for report download")
    
    str_timestamp = str(timestamp)
    during = str_timestamp + "," + str_timestamp
    downloader = adwords_client.GetReportDownloader(version='v201809')
    
    query_fields = ['id', 'ad_group_id', 'all_conversion_rate', 'conversion_rate', 
    'cost_per_conversion', 'all_conversion_value', 'all_conversions', 
    'approval_status', 'average_cost', 'average_cpc', 'average_cpm', 'average_cpv', 
    'average_pageviews', 'average_position', 'average_time_on_site', 'campaign_id', 'click_assisted_conversion_value',
    'click_assisted_conversions', 'clicks', 'conversions', 'cpc_bid', 'cpc_bid_source', 'criteria',
    'ctr', 'date', 'impression_assisted_conversions', 'impressions', 'keyword_match_type']
    fields = snake_to_pascal_case(query_fields)

    # Create report query.
    report_query = (adwords.ReportQueryBuilder()
        .Select(*fields)
        .From('KEYWORDS_PERFORMANCE_REPORT')
        .Where('CampaignStatus').In('ENABLED', 'PAUSED')
        .During(during).Build())

    report = downloader.DownloadReportAsStringWithAwql(
        report_query, 'CSV', skip_report_header=True, 
        skip_column_header=True, skip_report_summary=True)

    lines = report.split('\n')
    return csv_to_dict_list(query_fields, lines), 1


def add_adwords_document(project_id, customer_acc_id, doc, doc_type, timestamp):    
    uri = "/data_service/adwords/documents/add"
    url = options.data_service_host + uri

    payload = {
        "project_id": project_id,
        "customer_acc_id": customer_acc_id,
        "type_alias": doc_type,
        "value": doc,
        "timestamp": timestamp,
    }

    response = requests.post(url, json=payload)
    if not response.ok:
        log.error("Failed to add response %s to adwords warehouse: %d, %s", 
            doc_type, response.status_code, response.json())
    
    return response

def add_all_adwords_documents(project_id, customer_acc_id, docs, doc_type, timestamp):
    # Add each doc from adwords response which is list of docs.
    for doc in docs:
        add_adwords_document(project_id, customer_acc_id, 
            doc, doc_type, timestamp)

def get_last_sync_info():
    uri = "/data_service/adwords/documents/last_sync_info"
    url = options.data_service_host + uri

    response = requests.get(url)
    if not response.ok:
        log.error("Failed to get sync data: %d, %s", 
            response.status_code, response.json())
    
    return response

def get_adwords_timestamp_from_datetime(dt):
    if dt == None: return
    dt_year = str(dt.year)
    if len(dt_year) == 1: dt_year = '0'+dt_year

    dt_month = str(dt.month)
    if len(dt_month) == 1: dt_month = '0'+dt_month

    dt_day = str(dt.day)
    if len(dt_day) == 1: dt_day = '0'+dt_day
    
    return int(dt_year+dt_month+dt_day)


def get_datetime_from_adwords_timestamp(timestamp):
    if timestamp == None: return
    return datetime.datetime.strptime(str(timestamp), '%Y%m%d')

def inc_day_adwords_timestamp(timestamp):
    start_datetime = get_datetime_from_adwords_timestamp(timestamp)
    return get_adwords_timestamp_from_datetime(start_datetime + datetime.timedelta(days=1))


def get_adwords_timestamp_range(from_timestamp, to_timestamp=None):
    date_range = []
    if from_timestamp == None:
        return date_range
    
    # if to_timestamp not given: range till yesterday. 
    if to_timestamp == None:
        to_timestamp = get_adwords_timestamp_before_days(1)

    start_timestamp = from_timestamp
    while start_timestamp <= to_timestamp:
        date_range.append(start_timestamp)
        start_timestamp = inc_day_adwords_timestamp(start_timestamp)
    
    return date_range

def get_adwords_timestamp_before_days(days):
    return get_adwords_timestamp_from_datetime(
        datetime.datetime.utcnow() - datetime.timedelta(days=days))

def is_today(timestamp):
    todays_timestamp = int(time.strftime('%Y%m%d'))
    return timestamp == todays_timestamp

def track_request_stats(req_stat_map, project_id, doc_type, count):
    project_key = 'projects'
    if req_stat_map.get(project_key) == None:
        req_stat_map[project_key] = {}

    if req_stat_map[project_key].get(project_id) == None:
        req_stat_map[project_key][project_id] = {}

    if req_stat_map[project_key][project_id].get(doc_type) == None:
        req_stat_map[project_key][project_id][doc_type] = 0

    # project level count.
    req_stat_map[project_key][project_id][doc_type] += count

    total_key = 'total_by_type'
    if req_stat_map.get(total_key) == None:
        req_stat_map[total_key] = {}

    if req_stat_map[total_key].get(doc_type) == None:
        req_stat_map[total_key][doc_type] = 0

    # total count.
    req_stat_map[total_key][doc_type] += count

def sync(env, dry, skip_today, next_info):    
    project_id = next_info.get("project_id")
    customer_acc_id = next_info.get("customer_acc_id")
    refresh_token = next_info.get("refresh_token")
    timestamp = next_info.get("next_timestamp")
    doc_type = next_info.get("doc_type_alias")

    status = { "project_id": project_id, "timestamp": timestamp, "doc_type": doc_type, "status": "success" }

    if project_id == None or project_id == 0 or customer_acc_id == None or customer_acc_id == "" or doc_type == None or doc_type == "" or timestamp == None:
        log.error("Invalid project_id: %s or customer_account_id: %s or document_type: %s or timestamp: %s", 
            str(project_id), str(customer_acc_id), str(doc_type), str(timestamp))
        status["status"] = STATUS_FAILED
        status["message"] = "Invalid params project_id or customer_account_id or type or timestamp."
        return status

    if refresh_token == None or refresh_token == "":
        log.error("Invalid refresh token for project_id %s", refresh_token)
        status["status"] = STATUS_FAILED
        status["message"] = "Invalid refresh token."
        return status

    permission_error_key = str(customer_acc_id) + ':' + str(refresh_token)
    if permission_error_key in permission_error_cache:
        log.error("Skipping sync user permission denied already for project %s, 'customer_acc_id:refresh_token' : %s", 
            str(project_id), permission_error_key)
        return status

    if skip_today and is_today(timestamp):
        log.warning("Skipped sync for today for project_id %s doc_type %s.", str(project_id), doc_type)
        status["status"] = STATUS_SKIPPED
        status["message"] = "Skipped sync for today."
        return status

    # Todo: Reuse adwords_client, cache by refresh token.
    oauth2_client = oauth2.GoogleRefreshTokenClient(OAuthManager.get_client_id(), 
        OAuthManager.get_client_secret(), refresh_token)
    adwords_client = adwords.AdWordsClient(options.developer_token, 
        oauth2_client, ADWORDS_CLIENT_USER_AGENT)
    adwords_client.SetClientCustomerId(customer_acc_id)

    log.warning("Downloading project: %s, cutomer_account_id: %s, document_type: %s, timestamp: %s",
        str(project_id), customer_acc_id, doc_type, str(timestamp))

    docs = []
    req_count = 0
    try:
        if doc_type == "customer_account_properties":
            docs, req_count = get_customer_account_properties(adwords_client, customer_acc_id, timestamp)
            
        elif doc_type == "campaigns":
            docs, req_count = get_campaigns(adwords_client, timestamp)

        elif doc_type == "ads":
            docs, req_count = get_ads(adwords_client, timestamp)

        elif doc_type == "ad_groups":
            docs, req_count = get_ad_groups(adwords_client, timestamp)

        elif doc_type == "click_performance_report":
            docs, req_count = get_click_performance_report(adwords_client, timestamp)

        elif doc_type == "campaign_performance_report":
            docs, req_count = get_campaign_performance_report(adwords_client, timestamp)
        
        elif doc_type == "ad_performance_report":
            docs, req_count = get_ad_performance_report(adwords_client, timestamp)

        elif doc_type == "search_performance_report":
            docs, req_count = get_search_performance_report(adwords_client, timestamp)

        elif doc_type == "keyword_performance_report":
            docs, req_count = get_keywords_performance_report(adwords_client, timestamp)

        else: 
            log.error("Invalid document to sync from adwords: %s", str(doc_type))
            status["status"] = STATUS_FAILED
            status["message"] = "Invalid document type " + str(doc_type)
            return 

        track_request_stats(request_stat, project_id, doc_type, req_count)  
        
        if len(docs) > 0:
            if dry: log.error("Dry run. Skipped add adwords documents to db.")
            else: add_all_adwords_documents(project_id, customer_acc_id, docs, doc_type, timestamp) 
        else:
            add_adwords_document(project_id, customer_acc_id, docs, doc_type, timestamp)

    except Exception as e:
        str_exception = str(e)
        if "AuthorizationError.USER_PERMISSION_DENIED" in str_exception:
            permission_error_cache[permission_error_key] = str_exception
            status["status"] = STATUS_FAILED
            status["message"] = "Failed with exception: " + str_exception
            return status

        if "ReportDefinitionError.CUSTOMER_SERVING_TYPE_REPORT_MISMATCH" in str_exception:
            log.error("[Project: %s, Type: %s] Sync failed, Trying to download report from manager account.", 
                str(project_id), doc_type)
            status["status"] = STATUS_FAILED
            status["message"] = "Download failed for manager account with exception: " + str_exception
            return status

        log.error("[Project: %s, Type: %s] Sync failed with exception: %s", str(project_id), doc_type, str_exception)
        if "RateExceededError.RATE_EXCEEDED" in str_exception:
            # Todo: Do not exit? Stop downloading reports. 
            # Continue downloading other objects.
            notify(env, APP_NAME, { "status": STATUS_FAILED, "exception": str_exception, "requests": request_stat })
            sys.exit(0) # Use zero exit to avoid job retry.

        status["status"] = STATUS_FAILED
        status["message"] = "Failed with exception: " + str_exception
        return status

    status["status"] = "success"
    return status

# generates next sync info with all missing timestamps 
# for each document type.
def get_next_sync_info(last_sync_info):
    last_timestmap = last_sync_info.get('last_timestamp')
    if last_timestmap == None:
        log.error("Missing last_timestamp in last sync info.")
        return

    doc_type = last_sync_info.get('doc_type_alias')
    if doc_type == None:
        log.error("Missing doc_type_alias name on last_sync_info.")
        return

    next_sync_info = []

    # For non report doc_type sync only for current timestamp.
    # as no historical data would be available.
    adwords_timestamp_today = get_adwords_timestamp_from_datetime(datetime.datetime.utcnow())
    is_non_report_doc_type = doc_type == "campaigns" or doc_type == "ads" or doc_type == "ad_groups" or doc_type == "customer_account_properties"
    if is_non_report_doc_type and last_timestmap != adwords_timestamp_today:
        sync_info = last_sync_info.copy()
        sync_info['next_timestamp'] = adwords_timestamp_today
        next_sync_info.append(sync_info)
        return next_sync_info
    
    start_timestamp = 0
    if last_timestmap == 0:
        # new projects, starts with last 30 days.
        start_timestamp = get_adwords_timestamp_before_days(30)
    else:
        start_timestamp = inc_day_adwords_timestamp(last_timestmap)

    next_timestamps = get_adwords_timestamp_range(start_timestamp)

    # For report doc_type sync for date ranges after 
    # last_timestamp till current date.
    for timestamp in next_timestamps:
        sync_info = last_sync_info.copy()
        sync_info['next_timestamp'] = timestamp
        next_sync_info.append(sync_info)

    return next_sync_info

if __name__ == "__main__":
    (options, args) = parser.parse_args()

    if options.developer_token == "":
        log.error("Option: developer_token cannot be empty.")
        sys.exit(1)

    oauth_secret_str = options.oauth_secret.strip()
    if oauth_secret_str == "":
        log.error("Option: oauth_secret cannot be empty.")
        sys.exit(1)

    try:
        # initialize client secret.
        oauth_client_secret = json.loads(oauth_secret_str)
    except Exception as e:
        log.error("Failed to load oauth_secret JSON: %s", str(e))
        sys.exit(1)

    try:
        OAuthManager.init(oauth_client_secret)
    except Exception as e:
        log.error("Failed to init oauth manager with error %s", str(e))
        sys.exit(1)

    log.warning("Started adwords sync job.")

    # using string as kubernetes options doesn't 
    # allow boolean.
    is_dry = options.dry == "True"
    skip_today = options.skip_today == "True"

    last_sync_response = get_last_sync_info()
    last_sync_infos = last_sync_response.json()
    log.warning("Got adwords last sync info.")

    next_sync_failures = []
    next_sync_skipped = []
    next_sync_success = {}
    # Todo: Use multiple python process to distrubute.
    for last_sync in last_sync_infos:
        # add next_sync_info only for the selected project.
        if options.project_id != None:
            project_id = last_sync.get("project_id")
            if project_id != options.project_id:
                continue
        
        next_sync_infos = get_next_sync_info(last_sync)
        if next_sync_infos == None: continue
        for next_sync in next_sync_infos:
            response = sync(options.env, is_dry, skip_today, next_sync)
            status = response.get("status")
            if status == None:
                next_sync_failures.append("Sync status is missing on response")
            elif status == STATUS_FAILED:
                next_sync_failures.append(response)
            elif status == STATUS_SKIPPED:
                next_sync_skipped.append(response)
            else:
                next_sync_success[next_sync.get("project_id")] = next_sync.get("customer_acc_id")

    status_msg = ""
    if len(next_sync_failures) > 0: status_msg = "Failures on sync."
    else: status_msg = "Successfully synced."
    notify_payload = {
        "status": status_msg,
        "failures": next_sync_failures,
        "skipped": next_sync_skipped,
        "success": { "projects": next_sync_success },
        "requests": request_stat,
    }
    notify(options.env, APP_NAME, notify_payload)
    log.warning("Successfully synced. End of adwords sync job.")
    sys.exit(0)

