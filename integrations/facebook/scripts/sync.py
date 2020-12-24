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
from _datetime import timedelta

parser = OptionParser()
parser.add_option("--env", dest="env", default="development")
parser.add_option("--dry", dest="dry", help="", default="False")
parser.add_option("--skip_today", dest="skip_today", help="", default="False") 
parser.add_option("--project_id", dest="project_id", help="", default=None, type=int)
parser.add_option("--data_service_host", dest="data_service_host",
    help="Data service host", default="http://localhost:8089")

(options, args) = parser.parse_args()

APP_NAME = "facebook_sync"
CAMPAIGN_INSIGHTS = "campaign_insights"
AD_SET_INSIGHTS = "ad_set_insights"
AD_INSIGHTS = "ad_insights"
CAMPAIGN = "campaign"
AD= "ad"
AD_ACCOUNT = "ad_account"
AD_SET = "ad_set"
ACCESS_TOKEN = "int_facebook_access_token"
FACEBOOK_AD_ACCOUNT = "int_facebook_ad_account"
DATA = "data"
FACEBOOK_ALL = "facebook_all"
PLATFORM = "platform"

METRIC_TYPE_INCR = "incr"
HEALTHCHECK_PING_ID = "f2265955-a71c-42fe-a5ba-36d22a98419c"

def get_datetime_from_datestring(date):
    date = date.split("-")
    date = datetime(int(date[0]),int(date[1]),int(date[2]))
    return date

def notify(env, source, message):
    if env != "production": 
        log.warning("Skipped notification for env %s payload %s", env, str(message))
        return

    sns_url = "https://fjnvg9a8wi.execute-api.us-east-1.amazonaws.com/v1/notify"
    payload = { "env": env, "message": message, "source": source }
    response = requests.post(sns_url, json=payload)
    if not response.ok: log.error("Failed to notify through sns.")
    return response

def ping_healthcheck(env, healthcheck_id, message, endpoint=""):
    message = json.dumps(message, indent=1)
    log.warning("Healthcheck ping for env %s payload %s", env, message)
    if env != "production": 
        return

    try:
        requests.post("https://hc-ping.com/" + healthcheck_id + endpoint,
            data=message, timeout=10)
    except requests.RequestException as e:
        # Log ping failure here...
        log.error("Ping failed to healthchecks.io: %s" % e)

def record_metric(metric_type, metric_name, metric_value=0):
    payload = {
        "type": metric_type,
        "name": metric_name,
        "value": metric_value,
    }

    metrics_url = options.data_service_host + "/data_service/metrics"
    response = requests.post(metrics_url, json=payload)
    if not response.ok:
        log.error("Failed to record metric %s. Error: %s", metric_name, response.text)

def get_time_ranges_list(date_start, date_stop):
    days = (date_stop - date_start).days
    time_ranges = []
    for i in range (0,days):
        new_date = (date_start + timedelta(days=i)).date()
        new_string_date = new_date.strftime("%Y-%m-%d")
        time_range = {"since":new_string_date, "until":new_string_date}
        time_ranges.append(time_range)
    return time_ranges


def get_facebook_int_settings():
    uri = "/data_service/facebook/project/settings"
    url = options.data_service_host + uri

    response = requests.get(url)
    if not response.ok:
        log.error("Failed to get facebook integration settings from data services")
        return 
    return response.json()

def get_last_sync_info(project_id, account_id):
    uri = "/data_service/facebook/documents/last_sync_info"
    url = options.data_service_host + uri
    payload = {
        'project_id': project_id,
        "account_id" : account_id
    }
    response = requests.get(url, json=payload)
    all_info = response.json()
    sync_info_with_type = {}
    for info in all_info:
        date = datetime.fromtimestamp(info["last_timestamp"])
        sync_info_with_type[info['type_alias']+info['platform']]= date.strftime("%Y-%m-%d")
    return sync_info_with_type

def getNextDate(date):
    newDate = datetime.strptime(date, '%Y-%m-%d') + timedelta(days=1)
    return newDate.strftime('%Y-%m-%d')
    
def get_collections(facebook_int_setting, sync_info_with_type, date_stop):
    response = {}
    try:
        date_stop = get_datetime_from_datestring(date_stop)
        
        get_ad_account_data(facebook_int_setting["project_id"], facebook_int_setting[ACCESS_TOKEN],
            facebook_int_setting[FACEBOOK_AD_ACCOUNT], date_stop)

        campaigns_url = "https://graph.facebook.com/v9.0/{}/campaigns?access_token={}".format(
        facebook_int_setting[FACEBOOK_AD_ACCOUNT], facebook_int_setting[ACCESS_TOKEN])
        campaigns_response = requests.get(campaigns_url)
        if campaigns_response.ok:
            campaigns = campaigns_response.json()[DATA]
            
            if (CAMPAIGN_INSIGHTS+FACEBOOK_ALL not in sync_info_with_type):
                get_campaign_data(facebook_int_setting["project_id"],facebook_int_setting[FACEBOOK_AD_ACCOUNT],
                    facebook_int_setting[ACCESS_TOKEN], campaigns, 0, date_stop)
            else:
                get_campaign_data(facebook_int_setting["project_id"],facebook_int_setting[FACEBOOK_AD_ACCOUNT],
                    facebook_int_setting[ACCESS_TOKEN], campaigns, getNextDate(sync_info_with_type[CAMPAIGN_INSIGHTS+FACEBOOK_ALL]), date_stop)

        adsets_url = "https://graph.facebook.com/v9.0/{}/adsets?access_token={}".format(
        facebook_int_setting[FACEBOOK_AD_ACCOUNT], facebook_int_setting[ACCESS_TOKEN])
        adsets_response = requests.get(adsets_url)
        if adsets_response.ok:
            adsets = adsets_response.json()[DATA]
            if (AD_SET_INSIGHTS+FACEBOOK_ALL not in sync_info_with_type):
                get_adset_data(facebook_int_setting["project_id"],facebook_int_setting[FACEBOOK_AD_ACCOUNT],
                    facebook_int_setting[ACCESS_TOKEN], adsets, 0, date_stop)
            else:
                get_adset_data(facebook_int_setting["project_id"],facebook_int_setting[FACEBOOK_AD_ACCOUNT],
                    facebook_int_setting[ACCESS_TOKEN], adsets, getNextDate(sync_info_with_type[AD_SET_INSIGHTS+FACEBOOK_ALL]), date_stop)

        ads_url = "https://graph.facebook.com/v9.0/{}/ads?access_token={}".format(
        facebook_int_setting[FACEBOOK_AD_ACCOUNT], facebook_int_setting[ACCESS_TOKEN])
        ads_response = requests.get(ads_url)
        if ads_response.ok:
            ads = ads_response.json()[DATA]
            if (AD_INSIGHTS+FACEBOOK_ALL not in sync_info_with_type):
                get_ad_data(facebook_int_setting["project_id"],facebook_int_setting[FACEBOOK_AD_ACCOUNT],
                    facebook_int_setting[ACCESS_TOKEN], ads, 0, date_stop)
            else:
                get_ad_data(facebook_int_setting["project_id"],facebook_int_setting[FACEBOOK_AD_ACCOUNT],
                    facebook_int_setting[ACCESS_TOKEN], ads, getNextDate(sync_info_with_type[AD_INSIGHTS+FACEBOOK_ALL]), date_stop)
    except Exception as e:
        response["status"] = "failed"
        response["msg"] = "Failed with exception "+str(e)
        return response
    response["status"]="success"
    return response
    
        
def get_ad_account_data(project_id, access_token, ad_account_id, date_stop):
    timestamp = int(datetime.timestamp(date_stop))
    fields_ad_account = ["id", "balance", "name","partner", "spend_cap"]
    
    url = "https://graph.facebook.com/v9.0/{}?fields={}&&access_token={}".format(
        ad_account_id, fields_ad_account, access_token)
    response = requests.get(url)
    if not response.ok:
        log.error(response.status_code, ": ", response.reason, "failed to get ad account data from facebook")
        return
    add_facebook_document(project_id, ad_account_id, AD_ACCOUNT,  ad_account_id, response.json(), timestamp, FACEBOOK_ALL)
    

def get_campaign_data(project_id, ad_account_id, access_token, campaigns, date_start, date_stop):
    timestamp = int(datetime.timestamp(date_stop))
    for campaign in campaigns:
        
        fields_campaign = ["id", "name", "account_id", "buying_type","effective_status","spend_cap","start_time","stop_time"]
        
        url = "https://graph.facebook.com/v9.0/{}?fields={}&&access_token={}".format(
        campaign["id"], fields_campaign, access_token)
        response = requests.get(url)
        if not response.ok:
            log.error("failed to get campaign data from facebook")
        if response.ok:
            add_facebook_document(project_id, ad_account_id, CAMPAIGN, campaign["id"], response.json(), timestamp, FACEBOOK_ALL)
        
        fields = ["account_currency", "ad_id","ad_name","adset_name","campaign_name","adset_id","campaign_id","clicks","conversions",
        "cost_per_conversion","cost_per_ad_click","date_start", "cpc", "cpm","cpp","ctr",
        "date_stop","frequency","impressions","inline_post_engagement","social_spend", "spend","unique_clicks","reach"]
        get_insights(project_id, ad_account_id, access_token, campaign["id"], CAMPAIGN_INSIGHTS, fields, date_start, date_stop)

def get_adset_data(project_id, ad_account_id, access_token, adsets, date_start, date_stop):
    timestamp = int(datetime.timestamp(date_stop))
    for adset in adsets:
        
        fields_adset = ["id", "account_id","campaign_id","configured_status", "daily_budget", "effective_status","end_time","name"]
        
        url = "https://graph.facebook.com/v9.0/{}?fields={}&&access_token={}".format(
        adset["id"], fields_adset, access_token)
        response = requests.get(url)
        if not response.ok:
            log.error("failed to get adset data from facebook")
        if response.ok:
            add_facebook_document(project_id, ad_account_id, AD_SET, adset["id"], response.json(), timestamp, FACEBOOK_ALL)
        
        fields = ["account_currency", "ad_id","ad_name","adset_name","campaign_name","adset_id","campaign_id","clicks","conversions",
        "cost_per_conversion","cost_per_ad_click","cpc", "cpm","cpp","ctr",
        "date_start","date_stop","frequency","impressions","inline_post_engagement","social_spend", "spend","unique_clicks","reach"]
        get_insights(project_id, ad_account_id, access_token, adset["id"], AD_SET_INSIGHTS, fields, date_start, date_stop)

def get_ad_data(project_id, ad_account_id, access_token, ads, date_start, date_stop):
    timestamp = int(datetime.timestamp(date_stop))
    for ad in ads:
        fields_ad = ["id","adset_id","account_id","bid_amount","bid_type","campaign_id","name","status"]
        
        url = "https://graph.facebook.com/v9.0/{}?fields={}&&access_token={}".format(
        ad["id"], fields_ad, access_token)
        response = requests.get(url)
        if not response.ok:
            log.error("failed to get ad data from facebook")
        if response.ok:
            add_facebook_document(project_id, ad_account_id, AD, ad["id"], response.json(), timestamp, FACEBOOK_ALL)
        
        fields = ["account_currency", "ad_id","ad_name","adset_name","campaign_name","adset_id","campaign_id","clicks","conversions",
        "cost_per_conversion","cost_per_ad_click","cpc", "cpm","cpp","ctr",
        "date_start","date_stop","frequency","impressions","inline_post_engagement","social_spend", "spend","unique_clicks","reach"]
        get_insights(project_id, ad_account_id, access_token, ad["id"], AD_INSIGHTS, fields, date_start, date_stop)

def get_insights(project_id, ad_account_id, access_token, id, doc_type, fields_insight, date_start, date_stop):
    if date_start == 0 :
        date_start = date_stop - timedelta(days=60)
    else:
        date_start = get_datetime_from_datestring(date_start)
    time_ranges = get_time_ranges_list(date_start, date_stop)

    url = "https://graph.facebook.com/v9.0/{}/insights?time_ranges={}&&fields={}&&access_token={}".format(
    id, time_ranges, fields_insight, access_token)
    facebook_all_response = requests.get(url)
    if not facebook_all_response.ok:
        log.error("Failed to get insights from facebook")
        return

    breakdowns = ["publisher_platform"]
    url = "https://graph.facebook.com/v9.0/{}/insights?breakdowns={}&&time_ranges={}&&fields={}&&access_token={}".format(
    id, breakdowns, time_ranges, fields_insight, access_token)
    breakdown_response = requests.get(url)
    if not breakdown_response.ok:
        log.error("Failed to get insights from facebook")
        return

    for data in breakdown_response.json()[DATA]:
        date_stop = get_datetime_from_datestring(data["date_stop"])
        timestamp= int(datetime.timestamp(date_stop))
        add_document_response = add_facebook_document(project_id, ad_account_id, doc_type, id, data, timestamp, data["publisher_platform"])
        if not add_document_response.ok:
            return
    
    for data in facebook_all_response.json()[DATA]:
        date_stop = get_datetime_from_datestring(data["date_stop"])
        timestamp= int(datetime.timestamp(date_stop))
        add_document_response = add_facebook_document(project_id, ad_account_id, doc_type, id, data, timestamp, FACEBOOK_ALL)
        if not add_document_response.ok:
            return


def add_facebook_document(project_id, ad_account_id, doc_type, id, value, timestamp, platform):
    uri = "/data_service/facebook/documents/add"
    url = options.data_service_host + uri

    payload = {
        "project_id": int(project_id),
        "customer_ad_account_id": ad_account_id,
        "type_alias": doc_type,
        "id": id,
        "value": value,
        "timestamp":timestamp,
        "platform": platform,
    }
    response = requests.post(url, json=payload)
    if not response.ok:
        log.error("Failed to add response %s to facebook warehouse for project %s. StatusCode:  %d, %s", 
            doc_type, project_id, response.status_code, response.json())
    
    return response


if __name__ == "__main__":
    facebook_int_settings = get_facebook_int_settings()

    if(facebook_int_settings is not None):
        now = datetime.now()
        failures = []
        successes = []
        for facebook_int_setting in facebook_int_settings:
            sync_info_with_type = get_last_sync_info(facebook_int_setting["project_id"],facebook_int_setting["int_facebook_ad_account"])
            response = get_collections(facebook_int_setting, sync_info_with_type, now.strftime("%Y-%m-%d"))
            if(response["status"]=="failure"):
                failures.append(response)
            else:
                successes.append(response)
        status_msg = ""
        if len(failures) > 0: status_msg = "Failures on sync."
        else: status_msg = "Successfully synced."
        notification_payload = {
            "status": status_msg, 
            "failures": failures, 
            "success": successes,
        }

        log.warning("Successfully synced. End of facebook sync job.")
        if len(failures) > 0:
            ping_healthcheck(options.env, HEALTHCHECK_PING_ID, notification_payload, endpoint="/fail")
        else:
            ping_healthcheck(options.env, HEALTHCHECK_PING_ID, notification_payload)
        sys.exit(0)
        
        
