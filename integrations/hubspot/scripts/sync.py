from optparse import OptionParser
import logging as log
import requests
import json
import urllib
import sys
import time
from datetime import datetime

parser = OptionParser()
parser.add_option("--env", dest="env", default="development")
parser.add_option("--dry", dest="dry", action="store_true", help="", default=False)
parser.add_option("--first_sync", dest="first_sync",action="store_true", help="", default=False)
parser.add_option("--data_service_host", dest="data_service_host",
    help="Data service host", default="http://localhost:8089")
parser.add_option("--app_name", dest="app_name",
    help="App name", default="")
parser.add_option("--healthcheck_ping_id", dest="healthcheck_ping_id",
    help="Healthcheck ping id", default="")
parser.add_option("--batch_insert_by_project_ids", dest="batch_insert_by_project_ids",
    help="Enable batch insert for projects", default="")
parser.add_option("--batch_insert_doc_types", dest="batch_insert_doc_types",
    help="Enable batch insert for document types", default="")
parser.add_option("--enable_deleted_contacts", dest="enable_deleted_contacts", help="Enable deleted contacts flag", default=False, action="store_true")
parser.add_option("--enable_deleted_projectIDs", dest="enable_deleted_projectIDs", help="Enable deleted projectIDs", default="")
parser.add_option("--enable_company_contact_association_v2_by_project_id",
    dest="enable_company_contact_association_v2_by_project_id",help="Enable company contact association v2 for project",
    default="")
parser.add_option("--project_ids", dest="project_ids", help="Allowed project_ids", default="*")
parser.add_option("--disabled_project_ids", dest="disabled_project_ids", help="Disabled project_ids", default="")
parser.add_option("--disable_non_marketing_contacts_project_id", dest="disable_non_marketing_contacts_project_id", help="Projects to only pick marketing contacts", default="")
parser.add_option("--buffer_size_by_api_count", dest="buffer_size_by_api_count", help="Buffer size by number of api counts before performing insertion", default=1)
parser.add_option("--enable_buffer_before_insert_by_project_id", dest="enable_buffer_before_insert_by_project_id", help="Enable buffer before inserting by project id", default="")
parser.add_option("--hubspot_app_id", dest="hubspot_app_id", help="App id for hubspot access token", default="")
parser.add_option("--hubspot_app_secret", dest="hubspot_app_secret", help="App secret for hubspot access token", default="")
parser.add_option("--enable_contact_list_sync_by_project_id", dest="enable_contact_list_sync_by_project_id", help="", default="")
parser.add_option("--allowed_doc_types_sync", dest="allowed_doc_types_sync", help="", default="*")
parser.add_option("--use_sync_contact_list_v2", dest="use_sync_contact_list_v2",action="store_true", help="", default=False)
parser.add_option("--enable_owner_sync_by_project_id", dest="enable_owner_sync_by_project_id", help="", default="")
parser.add_option("--enable_sync_company_v3_by_project_id", dest="enable_sync_company_v3_by_project_id", help="Use API v3 to overcome 10K limit", default="")

APP_NAME = "hubspot_sync"
PAGE_SIZE = 100
DOC_TYPES = [ "contact", "company", "deal", "form", "form_submission", "contact_list" ]

METRIC_TYPE_INCR = "incr"
HEALTHCHECK_PING_ID = "87137001-b18b-474c-8bc5-63324baff2a8"
HEALTHCHECK_RUN_PING_ID = "745d16bc-542b-4b16-a029-05ca2c66ed8f"

API_RATE_LIMIT_TEN_SECONDLY_ROLLING = "TEN_SECONDLY_ROLLING"
API_RATE_LIMIT_DAILY = "DAILY"
API_ERROR_RATE_LIMIT = "RATE_LIMIT"
RETRY_LIMIT = 15
CONTACT_PROPERTY_KEY_LAST_MODIFIED_DATE = "lastmodifieddate"
COMPANY_PROPERTY_KEY_LAST_MODIFIED_DATE = "hs_lastmodifieddate"
RECORD_PROPERTIES_KEY = "properties"
REQUEST_TIMEOUT = 5*60 # 5 min

# Todo: Boilerplate, move this to a reusable module.
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
    if env != "production" or options.dry == True:
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

def create_document_in_batch(project_id, doc_type, documents, fetch_deleted_contact=False):
    if len(documents) == 0:
        return

    uri = "/data_service/hubspot/documents/add_batch"
    url = options.data_service_host + uri

    batched_document_payload = []
    for doc in documents:
        payload = get_document_payload(project_id,doc_type,doc,fetch_deleted_contact)
        batched_document_payload.append(payload)
    
    payload = {
        "project_id":project_id,
        "doc_type":doc_type,
        "documents":batched_document_payload
    }
    
    start_time = time.time()
    retries = 0
    while True:
        try:
            response = requests.post(url, json=payload)
            if response.ok or response.status_code == requests.codes['conflict']:
                log.warning("Successfully inserted batched %s of size %d.",doc_type, len(batched_document_payload))
            else:
                if response.status_code == requests.codes["server_error"]:
                    raise requests.exceptions.RequestException("Internal server error")
                log.error("Failed to add response %s to hubspot warehouse with uri %s: %d.", 
                    doc_type, uri, response.status_code)
            return response
        except requests.exceptions.RequestException as e:
            if retries > RETRY_LIMIT:
                raise Exception("Retry exhausted on connection error on inserting batch of document "+str(e)+ " , retries "+ str(retries))
            log.warning("Connection error occured on inserting batch of document %s retry %d, retrying in %d seconds", str(e),retries, 2)
            retries += 1
            time.sleep(2)
        finally:
            end_time = time.time()
            log.warning("Create_document_in_batch took %ds", end_time-start_time )
            log.warning("doc_type = %s", str(doc_type))

def get_document_payload(project_id, doc_type,doc, fetch_deleted_contact):
    payload = {
        "project_id": project_id,
        "type_alias": doc_type,
        "value": doc,
    }

    # Adding property value action=3 when fetching deleted contact record.
    if fetch_deleted_contact:
        payload["action"]=3
    return payload

def create_document(project_id, doc_type, doc, fetch_deleted_contact=False):
    uri = "/data_service/hubspot/documents/add"
    url = options.data_service_host + uri

    payload = get_document_payload(project_id,doc_type,doc,fetch_deleted_contact)

    start_time = time.time()
    retries = 0
    while True:
        try:
            response = requests.post(url, json=payload)
            if response.ok or response.status_code == requests.codes['conflict']:
                log.warning("Successfully inserted %s.",doc_type)
            else:
                if response.status_code == requests.codes["server_error"]:
                    raise requests.exceptions.RequestException("Internal server error")
                log.error("Failed to add response %s to hubspot warehouse with uri %s: %d.", 
                    doc_type, uri, response.status_code)
            return response
        except requests.exceptions.RequestException as e:
            if retries > RETRY_LIMIT:
                raise Exception("Retry exhausted on connection error on inserting document "+str(e)+ " , retries "+ str(retries))
            log.warning("Connection error occured on insert document %s retry %d, retrying in %d seconds", str(e),retries, 2)
            retries += 1
            time.sleep(2)
        finally:
            end_time = time.time()
            log.warning("Create_document took %ds", end_time-start_time )

def get_create_all_documents_with_buffer(project_id, doc_type, buffer_size, fetch_deleted_contact=False):
    buffered_docs = []
    def create_all_documents_with_buffer(docs, hasMore):
        nonlocal buffered_docs
        buffered_docs = buffered_docs + docs
        log.warning("Buffered %s of size %d",doc_type, len(buffered_docs) )
        if len(buffered_docs)<buffer_size and hasMore:
            return
        create_all_documents(project_id, doc_type,buffered_docs,fetch_deleted_contact)
        log.warning("Created %d %s.", len(buffered_docs),doc_type)
        buffered_docs = []
    return create_all_documents_with_buffer

def create_all_documents(project_id, doc_type, docs, fetch_deleted_contact=False):
    if options.dry == True:
        log.warning("Dry run. Skipped document upsert.")
        return

    if allow_batch_insert_by_project_id(project_id) and allow_batch_insert_doc_type(doc_type):
        return create_document_in_batch(project_id, doc_type, docs, fetch_deleted_contact)
    
    for doc in docs:
        create_document(project_id, doc_type, doc, fetch_deleted_contact)

def build_properties_param_str(properties=[]):
    param_str = ''
    for prop in properties:
        if param_str != '':
            param_str = param_str + '&'
        param_str = param_str + 'properties=' + prop
    return param_str

def get_all_properties_by_doc_type(project_id,doc_type, hubspot_request_handler):
    url = "https://api.hubapi.com/properties/v1/"+doc_type+"/properties?"
    get_url = url
    r = hubspot_request_handler(project_id, get_url)
    if not r.ok:
        log.error("Failure response %d from hubspot on get_properties_by_doc_type for doc type %s", r.status_code, doc_type)
        return [], r.ok

    response_dict = json.loads(r.text)
    properties = []
    for contact_property in response_dict:
        properties.append(contact_property["name"])
    return properties, r.ok

def get_hubspot_access_token_and_expiry_time(project_id, client_id, client_secret, refresh_token):
    access_token_url = "https://api.hubapi.com/oauth/v1/token?"
    parameter_dict = {"grant_type":"refresh_token","client_id":client_id,"client_secret":client_secret,"refresh_token":refresh_token}
    parameters = urllib.parse.urlencode(parameter_dict)
    url = access_token_url + parameters
    headers = {"Content-Type" : "application/x-www-form-urlencoded;charset=utf-8"}
    r = get_with_fallback_retry(project_id, url, requests.post, headers=headers)
    if not r.ok:
        raise Exception("Failed to get access token for project_id "+str(project_id)+" "+str(r.text))
    data = json.loads(r.text)
    access_token = data["access_token"]
    expires_in_sec = data["expires_in"]
    expire_time = time.time() + expires_in_sec
    return access_token, expire_time


def get_hubspot_request_handler(project_id, refresh_token, api_key):
    client_id, client_secret = get_hubspot_app_credentials()
    access_token = ""
    access_token_expire_time = None
    def hubspot_request_handler(project_id, url, request = requests.get, json=None, headers = None):
        if refresh_token == "" and api_key =="":
            raise Exception("Missing api key and refresh token")

        if refresh_token == "":
            log.warning("Using api key for project_id %d",project_id)
            parameter_dict = {'hapikey': api_key}
            parameters = urllib.parse.urlencode(parameter_dict)
            request_url = url +"&"+ parameters
            return get_with_fallback_retry(project_id, request_url, request, json, headers)

        if client_id =="" or client_secret =="":
            raise Exception("Missing app credentials")
        nonlocal access_token
        nonlocal access_token_expire_time
        log.warning("Using access token for project_id %d",project_id)
        if access_token == "" or access_token_expire_time == None or access_token_expire_time - time.time() < 10:
            access_token, access_token_expire_time = get_hubspot_access_token_and_expiry_time(project_id, client_id, client_secret, refresh_token)

        headers = {"Authorization": "Bearer " + access_token}
        return get_with_fallback_retry(project_id, url, request, json, headers)
    return hubspot_request_handler

def get_with_fallback_retry(project_id, get_url, request = requests.get, json_object=None, headers = None):
    retries = 0
    start_time  = time.time()
    try:
        while True:
            try:
                r = request(url=get_url, headers = headers, json=json_object, timeout=REQUEST_TIMEOUT)
                if r.status_code != 429:
                    if not r.ok:
                        if r.status_code== 414 or r.status_code == 404:
                            return r
                        if r.status_code == 400:
                            err_json = json.loads(r.text)
                            if err_json.get("status") == "error" and ("Unknown Contacts Search API failure" in err_json.get("message")):
                                raise Exception("Failed to get data from hubspot. User does not have permissions")
                        if retries < RETRY_LIMIT:
                            log.error("Failed to get data from hubspot %d.Retries %d. Retrying in 2 seconds %s",r.status_code,retries, r.text)
                            time.sleep(2)
                            retries += 1
                            continue
                        log.error("Retry exhausted. Failed to get data after %d retries",retries)
                        raise Exception("Retry exhausted. Failed to get data after "+str(retries)+" retries")
                    return r
                res_json = r.json()
                if res_json["errorType"] == API_ERROR_RATE_LIMIT:
                    if res_json["policyName"] == API_RATE_LIMIT_TEN_SECONDLY_ROLLING:
                        if retries > RETRY_LIMIT:
                            log.error("Retry exhausted on %s for project_id %d.",API_RATE_LIMIT_TEN_SECONDLY_ROLLING,project_id)
                            raise Exception("Retry exhausted with "+str(retries)+" retries "+str(res_json))

                        log.warning("Hubspot API limit exceeed %s retry %d, retrying in 2 seconds",API_RATE_LIMIT_TEN_SECONDLY_ROLLING, retries)
                        retries += 1
                        time.sleep(2)
                        continue
                    elif res_json["policyName"] == API_RATE_LIMIT_DAILY:
                        raise Exception("Hubspot API daily rate limit exceeded " + str(res_json))
                    else:
                        raise Exception("Unknown error occured on errorType RATE_LIMIT " + str(res_json))
                else:
                    raise Exception("Unknown error occured "+str(res_json))
            except requests.exceptions.RequestException as e:
                if retries > RETRY_LIMIT:
                    raise Exception("Retry exhausted on connection error "+str(e)+ " , retries "+ str(retries))
                log.warning("Connection error occured %s retry %d, retrying in %d seconds", str(e),retries, 2)
                retries += 1
                time.sleep(2)
    finally:
        end_time = time.time()
        log.warning("Request took %d sec", end_time - start_time )

def get_batch_documents_max_timestamp(project_id,docs, object_type, max_timestamp):
    last_modified_key = ""
    if object_type =="contacts":
        last_modified_key = CONTACT_PROPERTY_KEY_LAST_MODIFIED_DATE
    if object_type =="companies":
        last_modified_key = COMPANY_PROPERTY_KEY_LAST_MODIFIED_DATE

    for doc in docs:
        object_properties = doc[RECORD_PROPERTIES_KEY]
        if last_modified_key not in object_properties:
                log.error("Missing lastmodified in contacts for project_id %d.",project_id)
                return max_timestamp
        doc_last_modified_timestamp = int(object_properties[last_modified_key]["value"])
        if max_timestamp== 0 :
            max_timestamp = doc_last_modified_timestamp
        elif max_timestamp < doc_last_modified_timestamp:
            max_timestamp = doc_last_modified_timestamp
    return max_timestamp

def should_continue_contact_historical_data(project_id, object_dict, last_sync_timestamp):
    docs = object_dict["contacts"]

    curr_timestamp = 0
    if len(docs) < 1:
        return False, False
    for i in range(len(docs)):
        if RECORD_PROPERTIES_KEY not in docs[i]:
            log.error("Unknow error for project_id %d. Continue to pull all data", project_id)
            return False, True
        else:
            object_properties = docs[i][RECORD_PROPERTIES_KEY]
            if CONTACT_PROPERTY_KEY_LAST_MODIFIED_DATE not in object_properties:
                log.error("Missing lastmodified in contacts for project_id %d. Continue to pull all data",project_id)
                return False, True
            else:
                doc_last_modified_timestamp = int(object_properties[CONTACT_PROPERTY_KEY_LAST_MODIFIED_DATE]["value"])
                log.error("last modified timestamp %d",doc_last_modified_timestamp)
                if i==0:
                    curr_timestamp = doc_last_modified_timestamp

                if doc_last_modified_timestamp <= curr_timestamp:
                    curr_timestamp = doc_last_modified_timestamp
                else:
                    log.error("Invalid order of records for project_id %d.Received timestamp %d Continue to pull all data",project_id, doc_last_modified_timestamp)
                    return False, True
                if last_sync_timestamp > doc_last_modified_timestamp:
                    return True, False

    return False, False

def get_contacts_with_properties_by_id(project_id,get_url, hubspot_request_handler):
    batch_contact_url = "https://api.hubapi.com/contacts/v1/contact/vids/batch?"

    log.warning("Downloading contacts without properties list "+get_url)
    r = hubspot_request_handler(project_id,get_url)
    if not r.ok:
        return {},{}, r
    response_dict,unmodified_dict = json.loads(r.text),json.loads(r.text)
    if "contacts" not in response_dict:
        raise Exception("Missing contacts property key on contacts")
    contacts = response_dict["contacts"]
    contact_ids = []
    for contact in contacts:
        if "vid" not in contact:
            log.error("Missing contact vid on contacts api")
            continue
        contact_ids.append(contact["vid"])
    contact_ids_str = "&".join([ "vid="+str(id) for id in contact_ids ])
    batch_url = batch_contact_url + "&" + contact_ids_str
    log.warning("Downloading batch contact from url "+batch_url)
    r = hubspot_request_handler(project_id,batch_url)
    if not r.ok:
        log.error("Failure getting batch contacts for project_id %d on sync_contacts", project_id)
        return {},{},r
    batch_contact_dict = json.loads(r.text)

    for contact in contacts:
        if "vid" not in contact:
            log.error("Missing contact vid on get batch contacts")
            continue
        contact_id = str(contact["vid"])
        log.info("Inserting properties into contact "+ contact_id)
        if contact_id in batch_contact_dict:
            if RECORD_PROPERTIES_KEY not in batch_contact_dict[str(contact["vid"])]:
                log.error("Missing properties key on batch contact")
                continue
            contact[RECORD_PROPERTIES_KEY] = batch_contact_dict[str(contact["vid"])][RECORD_PROPERTIES_KEY]
        else :
            log.error("Missing contact %s in batch contact processing ",contact_id)

    response_dict["contacts"] = contacts

    return response_dict, unmodified_dict, r

def add_contactId( email, project_id, engagement, hubspot_request_handler):
    get_url = "https://api.hubapi.com/contacts/v1/contact/email/" + email + "/profile?"
    r  = hubspot_request_handler(project_id, get_url)
    if not r.ok:
        log.error("Failure response %d from hubspot on contactID", r.status_code)
        return

    response = json.loads(r.text)
    engagements = engagement["engagement"]
    if engagements["type"] == "INCOMING_EMAIL":
        engagement["metadata"]["from"]["contactId"] = response["vid"]
    elif engagements["type"] == "EMAIL":
        engagement["metadata"]["to"][0]["contactId"] = response["vid"]

def sync_engagements(project_id, refresh_token, api_key, last_sync_timestamp=0):
    page_count = 100
    get_url = "https://api.hubapi.com/engagements/v1/engagements/recent/modified?"+"count="+str(page_count)+"&since="+str(last_sync_timestamp)
    final_url = get_url
    has_more = True
    engagement_api_calls = 0
    latest_timestamp = None
    buffer_size = page_count * get_buffer_size_by_api_count()
    create_all_engagement_documents_with_buffer = get_create_all_documents_with_buffer(project_id, 'engagement', buffer_size)
    hubspot_request_handler = get_hubspot_request_handler(project_id, refresh_token, api_key)
    call_disposition = get_call_disposition(project_id, hubspot_request_handler)
    while has_more:
        log.warning("Downloading engagements for project_id %d from url %s.", project_id, final_url)
        r = hubspot_request_handler(project_id, final_url)
        if not r.ok:
            log.error("Failure response %d from hubspot on sync_engagements", r.status_code)
            break
        engagement_api_calls += 1
        filter_engagements = []
        if not r.ok:
            log.error("Failure response %d from hubspot on sync_engagements", r.status_code)
            latest_timestamp  = last_sync_timestamp
            break
        else:
            response = json.loads(r.text)
            for engagement in response['results']:
                # pick first timestamp in reverse chronological order
                if latest_timestamp is None and "lastUpdated" in engagement["engagement"]:
                    latest_timestamp = engagement["engagement"]["lastUpdated"]
                engagements = engagement["engagement"]
                if engagements["type"] == "CALL":
                    add_disposition_label(engagement, call_disposition)
                    filter_engagements.append(engagement)
                elif engagements["type"] == "MEETING":
                    filter_engagements.append(engagement)
                elif engagements["type"] == "INCOMING_EMAIL":
                    if "metadata" in engagement and "from" in engagement["metadata"] and "email" in engagement["metadata"]["from"]:
                        add_contactId( engagement["metadata"]["from"]["email"], project_id, engagement, hubspot_request_handler)
                    filter_engagements.append(engagement)
                elif engagements["type"] == "EMAIL":
                    if "metadata" in engagement and "to" in engagement["metadata"] and len(engagement["metadata"]["to"])>0 and "email" in engagement["metadata"]["to"][0]:
                        add_contactId( engagement["metadata"]["to"][0]["email"], project_id, engagement, hubspot_request_handler)
                    filter_engagements.append(engagement)
        response = json.loads(r.text)
        if 'hasMore' in response and 'offset' in response:
            has_more = response['hasMore']
            final_url = get_url + "&offset=" + str(response['offset'])
        else :
            has_more = False
        if allow_buffer_before_insert_by_project_id(project_id):
            create_all_engagement_documents_with_buffer(filter_engagements,has_more)
            log.warning("Downloaded %d engagements.", len(filter_engagements))
        else:
            create_all_documents(project_id, 'engagement', filter_engagements)
            log.warning("Downloaded and created %d engagements.", len(filter_engagements))
    create_all_engagement_documents_with_buffer([],False) ## flush any remaining docs in memory
    return engagement_api_calls,latest_timestamp

def add_disposition_label(engagement, call_disposition):
    if "metadata" in engagement and "disposition" in engagement["metadata"]:
        disposition_label = call_disposition.get(engagement["metadata"]["disposition"])
        if disposition_label != None:
            engagement["metadata"]["disposition_label"] = disposition_label

def get_call_disposition(project_id, hubspot_request_handler):
        get_disposition_label_url = "https://api.hubapi.com/calling/v1/dispositions?"
        r = hubspot_request_handler(project_id, get_disposition_label_url)
        if not r.ok:
            log.error("Failure response %d from engagement dispositions on get_call_disposition", r.status_code)
            return
        response = json.loads(r.text)
        disposition_values = {}
        for data in response:
            if data["id"] and data["label"]:
                disposition_values[data["id"]] = data["label"]
        return disposition_values

def is_marketing_contact(doc):
    if "properties" in doc:
        properties = doc["properties"]
        if "hs_marketable_status" in properties:
            return properties["hs_marketable_status"]["value"] =='true'
    return False


def get_filtered_contacts_project_id(project_id, docs):
    if disable_non_marketing_contacts_by_project_id(project_id):
        log.warning("Filtering only marketing contacts")
        marketing_docs = []
        for doc in docs:
            if is_marketing_contact(doc):
                marketing_docs.append(doc)
        log.warning("Filtered marketing contacts %d",len(marketing_docs))
        return marketing_docs
    return docs

def sync_contacts(project_id, refresh_token, api_key, last_sync_timestamp, sync_all=False):
    if sync_all:
        # init sync all contacts.
        url = "https://api.hubapi.com/contacts/v1/lists/all/contacts/all?"
        log.warning("Downloading all contacts for project_id : "+ str(project_id) + ".")
    else:
        # sync recently updated and created contacts.
        url = "https://api.hubapi.com/contacts/v1/lists/recently_updated/contacts/recent?"
        log.warning("Downloading recently created or modified contacts for project_id : "+ str(project_id) + ".")

    buffer_size = PAGE_SIZE * get_buffer_size_by_api_count()
    create_all_contact_documents_with_buffer = get_create_all_documents_with_buffer(project_id,"contact",buffer_size)
    has_more = True
    count = 0
    hubspot_request_handler = get_hubspot_request_handler(project_id, refresh_token, api_key)
    parameter_dict = {'count': PAGE_SIZE, 'formSubmissionMode': 'all' }
    properties, ok = get_all_properties_by_doc_type(project_id,"contacts", hubspot_request_handler)
    if not ok:
        log.error("Failure loading properties for project_id %d on sync_contacts", project_id)
        return 0, 0

    contact_api_calls = 0
    ordered_historical_data_failure= False
    max_timestamp = 0
    err_url_too_long = False
    while has_more:
        parameters = urllib.parse.urlencode(parameter_dict)
        get_url = url + parameters

        # contacts api uses property instead of properties in query parameter
        properties_str = "&".join([ "property="+property_name for property_name in properties ])
        get_url_with_properties = get_url + '&' + properties_str

        log.warning("Downloading contacts for project_id %d from url %s.", project_id, get_url_with_properties)
        response_dict = {}
        unmodified_response_dict={}
        if err_url_too_long == False:
            r = hubspot_request_handler(project_id, get_url_with_properties)
            if not r.ok:
                if r.status_code == 414:
                    log.error("Failure response %d from hubspot on sync_contacts, using fallback logic", r.status_code)
                    err_url_too_long= True
                else:
                    log.error("Failure response %d from hubspot on sync_contacts", r.status_code)
                    break
            else:
                response_dict = json.loads(r.text)

        if err_url_too_long == True:
            contact_dict,unmodified_response_dict, r = get_contacts_with_properties_by_id(project_id,get_url, hubspot_request_handler)
            if not r.ok:
                log.error("Failure response %d from hubspot on batch sync_contacts", r.status_code)
                break
            response_dict = contact_dict

        contact_api_calls +=1

        has_more = response_dict['has-more']
        docs = response_dict['contacts']
        validate_order_dict = unmodified_response_dict if err_url_too_long == True else response_dict
        if sync_all == False and ordered_historical_data_failure == False:
            should_stop, ordered_historical_data_failure  = should_continue_contact_historical_data(project_id,
            validate_order_dict,last_sync_timestamp-(10800*1000)) # fallback to 3hrs since hubspot api may not have updated latest change on bulk updates
            if should_stop:
                has_more = False
        if sync_all:
            parameter_dict['vidOffset'] = response_dict['vid-offset']
        else:
            if parameter_dict.get('vidOffset') == response_dict.get('vid-offset'):
                log.warning("same offset on consective pages on recent contacts sync %s, %s,  and has more is %s", 
                    parameter_dict.get('timeOffset'), response_dict.get('time-offset'), response_dict.get('has-more'))
                raise Exception("Same offset for consecutive pages on sync_contacts")
            parameter_dict['timeOffset'] = response_dict['time-offset']
            parameter_dict['vidOffset'] = response_dict['vid-offset']

        max_timestamp = get_batch_documents_max_timestamp(project_id, docs,"contacts",max_timestamp)
        docs = get_filtered_contacts_project_id(project_id, docs)
        count = count + len(docs)
        if allow_buffer_before_insert_by_project_id(project_id):
            create_all_contact_documents_with_buffer(docs,has_more)
            log.warning("Downloaded %d contacts. total %d.", len(docs), count)
        else:
            create_all_documents(project_id, 'contact', docs)
            log.warning("Downloaded and created %d contacts. total %d.", len(docs), count)
    
    create_all_contact_documents_with_buffer([],False) ## flush any remainig docs in memory
    return contact_api_calls, max_timestamp

def get_all_contact_lists_info(project_id, refresh_token, api_key):
    url = "https://api.hubapi.com/contacts/v1/lists?"
    all_contact_lists = {}
    get_url = url
    hubspot_request_handler = get_hubspot_request_handler(project_id, refresh_token, api_key)
    contact_list_api_calls = 0

    has_more = True
    while has_more:
        log.warning("Downloading contact list for project_id %d %s",project_id, get_url)
        r = hubspot_request_handler(project_id, get_url)
        if not r.ok:
            log.error("Failure response %d from hubspot on sync_contact_lists", r.status_code)
            return
        response_dict = json.loads(r.text)
        contact_lists = response_dict["lists"]
        for contact_list in contact_lists:
            all_contact_lists[contact_list["listId"]] = contact_list

        contact_list_api_calls += 1

        has_more = response_dict.get('has-more')
        if has_more:
            get_url = url +"&"+"offset="+str(response_dict["offset"])
    
    return all_contact_lists, contact_list_api_calls

def sync_contact_lists_contact_ids(project_id, refresh_token, api_key, contact_lists_info):
    # all contacts endpoint
    url = "https://api.hubapi.com/contacts/v1/lists/all/contacts/all?"

    parameters_dict = {"showListMemberships":"true","count":100}
    required_properties = ["email","phone"]

    url = url + urllib.parse.urlencode(parameters_dict)
    url = url + "&"+ "&".join([ "property="+property_name for property_name in required_properties ])

    buffer_size = PAGE_SIZE * get_buffer_size_by_api_count()
    create_all_contact_list_documents_with_buffer = get_create_all_documents_with_buffer(project_id,"contact_list", buffer_size)

    hubspot_request_handler = get_hubspot_request_handler(project_id, refresh_token, api_key)

    get_url = url
    contact_api_calls = 0
    total_documents_created = 0
    has_more = True
    while has_more:
        log.warning("Downloading contacts for contact lists for project_id %d from url %s.", project_id, get_url)
        r = hubspot_request_handler(project_id, get_url)
        if not r.ok:
            log.error("Failure response %d from hubspot on sync_contacts", r.status_code)
            break

        response_dict = json.loads(r.text)
        contacts = response_dict['contacts']
        for contact in contacts:
            list_memberships = contact.get("list-memberships")
            for membership in list_memberships:
                if membership["static-list-id"] not in contact_lists_info.keys():
                    continue

                if membership["is-member"]:
                    contact_list = {
                        "contact_id": contact["vid"],
                        "contact_timestamp": membership["timestamp"],
                    }

                    contact_list.update(contact_lists_info[membership["static-list-id"]])
                    if allow_buffer_before_insert_by_project_id(project_id):
                        create_all_contact_list_documents_with_buffer([contact_list], True)
                    else:
                        create_all_documents(project_id, 'contact_list', [contact_list])
                    log.warning("Downloaded contact_list %d contact_id %d for project_id %d", membership["static-list-id"], contact["vid"], project_id)
                    total_documents_created += 1

        contact_api_calls += 1
        has_more = response_dict.get('has-more')
        if has_more:
            vid_offset = response_dict.get("vid-offset")
            if not vid_offset:
                break
            get_url = url + "&vidOffset="+ str(vid_offset)
    
    log.warning("Downloaded %d contacts for contact_list %d for project_id %d", total_documents_created, membership["static-list-id"], project_id)
    create_all_contact_list_documents_with_buffer([], False) ## flush any remainig docs in memory
    return contact_api_calls

def sync_contact_lists_with_recent_contact_ids(project_id, refresh_token, api_key, contact_lists_info, last_sync_timestamp=0):
    log.warning("Downloading recently created or modified contacts for project_id : "+ str(project_id) + ".")
    hubspot_request_handler = get_hubspot_request_handler(project_id, refresh_token, api_key)

    buffer_size = PAGE_SIZE * get_buffer_size_by_api_count()
    create_all_contact_list_documents_with_buffer = get_create_all_documents_with_buffer(project_id,"contact_list", buffer_size)

    contact_api_calls = 0
    total_documents_created = 0
    
    for list_id in contact_lists_info.keys():
        url = "https://api.hubapi.com/contacts/v1/lists/"+str(list_id)+"/contacts/recent?"
        parameters_dict = {"showListMemberships":"true","count":100}
        url = url + urllib.parse.urlencode(parameters_dict)
        get_url = url

        has_more = True
        pagination_timestamp = int(time.time() * 1000)

        while has_more and pagination_timestamp > last_sync_timestamp - (7200*1000):
            log.warning("Downloading recent contacts for contact lists for project_id %d from url %s.", project_id, get_url)
            r = hubspot_request_handler(project_id, get_url)
            if not r.ok:
                log.error("Failure response %d from hubspot on sync_contacts", r.status_code)
                break

            response_dict = json.loads(r.text)
            contacts = response_dict['contacts']
            for contact in contacts:
                list_memberships = contact.get("list-memberships")
                for membership in list_memberships:
                    if membership["static-list-id"] == list_id and membership["is-member"]:
                        contact_list = {
                            "contact_id": contact["vid"],
                            "contact_timestamp": membership["timestamp"],
                        }

                        contact_list.update(contact_lists_info[list_id])
                        if allow_buffer_before_insert_by_project_id(project_id):
                            create_all_contact_list_documents_with_buffer([contact_list], True)
                        else:
                            create_all_documents(project_id, 'contact_list', [contact_list])
                        log.warning("Downloaded contact_list %d contact_id %d for project_id %d", membership["static-list-id"], contact["vid"], project_id)
                        total_documents_created += 1

            contact_api_calls +=1
            has_more = response_dict.get('has-more')
            if has_more:
                vid_offset = response_dict.get("vid-offset")
                if not vid_offset:
                    break
                get_url = url + "&vidOffset="+ str(vid_offset)

                pagination_timestamp = int(response_dict.get("time-offset"))
                if not pagination_timestamp:
                    break
                get_url = get_url + "&timeOffset="+ str(pagination_timestamp)
    
    log.warning("Downloaded %d contacts for contact_list %d for project_id %d", total_documents_created, membership["static-list-id"], project_id)
    create_all_contact_list_documents_with_buffer([], False) ## flush any remainig docs in memory
    return contact_api_calls


def sync_all_contact_lists_v2(project_id, refresh_token, api_key, last_sync_timestamp=0):
    start_timestamp  = round(time.time() * 1000)

    contact_lists_info, contact_list_api_calls = get_all_contact_lists_info(project_id, refresh_token, api_key)
    
    all_contacts_api_calls = 0
    recent_contacts_api_calls = 0

    if last_sync_timestamp == 0:
        last_sync_timestamp = start_timestamp - (24*3600*1000)

    if last_sync_timestamp == 0:
        all_contacts_api_calls = sync_contact_lists_contact_ids(project_id, refresh_token, api_key, contact_lists_info)
    else:
        recent_contacts_api_calls = sync_contact_lists_with_recent_contact_ids(project_id, refresh_token, api_key, contact_lists_info, last_sync_timestamp)
    
    log.warning("Downloaded contact_lists with id and timestamp for project_id %d", project_id)

    total_api_calls = contact_list_api_calls + all_contacts_api_calls + recent_contacts_api_calls

    return total_api_calls, start_timestamp

def add_paginated_associations(project_id, to_object_ids, next_page_url, hubspot_request_handler):
    while True:
        url = next_page_url
        log.warning("Downloading paginated associations for project_id %d url %s", project_id,next_page_url)
        r  = hubspot_request_handler(project_id, url)
        if not r.ok:
            raise Exception("failed to get next page on paginated associations "+str(r.text))
        data = json.loads(r.text)
        results = data["results"]
        for association in results:
            to_object_ids.append(association["id"])
        total_associated_ids = len(results)
        if "paging" in data and "next" in data["paging"]:
            pagination = data["paging"]["next"]
            if "link" in pagination and pagination["link"] != "":
                next_page_url = pagination["link"]
                continue
        break

    if len(to_object_ids)==0:
        log.warning("Received empty result on paginated associations.")
    return to_object_ids, total_associated_ids

def validate_association_errors(association_response):
    if "errors" not in association_response:
        return True
    for error in association_response["errors"]:
        # ignore error if no contact is associated to company
        if error["subCategory"] != "crm.associations.NO_ASSOCIATIONS_FOUND":
            raise Exception("Unknown error occured in association erros "+str(error))
    return True

def fill_contacts_for_companies(project_id, docs, hubspot_request_handler):
    if use_company_contact_association_v2_by_project_id(project_id):
        return fill_contacts_for_companies_v2(project_id, docs, hubspot_request_handler)
    return fill_contacts_for_companies_v1(project_id, docs, hubspot_request_handler)


def fill_contacts_for_companies_v2(project_id, companies, hubspot_request_handler):
    companies_ids = []
    for company in companies:
        companies_ids.append(company["companyId"])
    
    associations, api_calls = get_associations(project_id, "company", companies_ids, "contact", hubspot_request_handler)
    for i in range(len(companies)):
        company_id = str(companies[i]["companyId"])
        if company_id in associations:
            contact_ids = []
            for id in associations[company_id]:
                contact_ids.append(int(id)) ## store as integer
            companies[i]["contactIds"] = contact_ids
    return companies, api_calls

def get_associations(project_id, from_object_name, from_object_ids, to_object_name, hubspot_request_handler):
    url = "https://api.hubapi.com/crm/v3/associations/"+from_object_name+"/"+to_object_name+"/batch/read"
    get_url = url +"?"
    ids_payload = []
    for id in from_object_ids:
        ids_payload.append({"id":id})
    payload = {"inputs":ids_payload}
    api_count = 0
    log.warning("Downloading %s %s using association for project_id %d url %s", from_object_name, to_object_name, project_id, url)
    r = hubspot_request_handler(project_id, get_url, requests.post, json=payload)
    if not r.ok:
        log.warning("Failed to get %s %s id using associations for project_id %d %s",from_object_name, to_object_name, project_id, str(r.text))
        raise Exception("Failed to get "+from_object_name+" "+to_object_name+" using association")

    # Response structure
    # {
    #    "status":
    #    "result":[
    #        { 
    #           "from":{
    #                     "id":from_id
    #                  },
    #            "to":[
    #                   {
    #                      "id":to_id
    #                   }
    #                 ],
    #            "paging": {
    #                   "next": {
    #                       "after": "",
    #                       "link": ""
    #                   }
    #             }
    #        }
    #    ],
    #    "numErrors":
    #    "errors":[{"subCategory":""}]
    # }

    api_count+=1
    data = json.loads(r.text)
    validate_association_errors(data)
    results = data["results"]
    associations = {} ##{from_object_id:[]to_object_id}

    total_associated_ids = 0
    for association in results:
        from_object_id = association["from"]["id"]
        to_object_ids = []
        for to_object in association["to"]:
            to_object_ids.append(to_object["id"])
            total_associated_ids+=1
        
        if "paging" in association and "next" in association["paging"]:
            pagination = association["paging"]["next"]
            if "link" in pagination and pagination["link"]!= "":
                to_object_ids, paginated_associated_ids = add_paginated_associations(project_id, 
                    to_object_ids, pagination["link"], hubspot_request_handler)
                total_associated_ids += paginated_associated_ids
                api_count+=1
        associations[from_object_id] = to_object_ids
    log.warning("Downloaded %s %s using association total %d",from_object_name, to_object_name, total_associated_ids)

    return associations, api_count

def get_deleted_contacts(project_id, refresh_token,  api_key):
    url = "https://api.hubapi.com/crm/v3/objects/contacts/?"
    parameter_dict = {'archived': "true","limit":100}
    parameters = urllib.parse.urlencode(parameter_dict)
    final_url = url + parameters    
    has_more = True
    count_api_calls = 0
    count = 0
    
    buffer_size = PAGE_SIZE * get_buffer_size_by_api_count()
    create_all_deleted_contacts_documents_with_buffer = get_create_all_documents_with_buffer(project_id, "contact", buffer_size, True)

    hubspot_request_handler = get_hubspot_request_handler(project_id, refresh_token, api_key)
    while has_more:
        log.warning("Downloading deleted contacts from url: %s", final_url)
        res = hubspot_request_handler(project_id, final_url)
        if not res.ok:
            raise Exception('Failed to get deleted contacts: ' + str(res.status_code))

        response_dict = json.loads(res.text)
        count_api_calls+=1
        docs = response_dict.get('results')
        if not docs:
            log.warning("Found empty response for deleted_contacts")
            break

        count = count + len(docs)
        if allow_buffer_before_insert_by_project_id(project_id):
            create_all_deleted_contacts_documents_with_buffer(docs, has_more)
            log.warning("Downloaded %d deleted_contacts. total %d.", len(docs), count)
        else:
            create_all_documents(project_id, 'contact', docs, True)
            log.warning("Downloaded and created %d deleted_contacts. total %d.", len(docs), count)

        has_more = response_dict.get('paging')
        if has_more:
            next_property = has_more.get('next')
            if not next_property:
                continue
            next_link = next_property.get('link')
            if not next_link:
                raise Exception('Found empty link value to fetch deleted contacts')
            final_url = next_link
    
    create_all_deleted_contacts_documents_with_buffer([], False) ## flush any remaining docs in memory
    return count_api_calls


## https://community.hubspot.com/t5/APIs-Integrations/Deals-Endpoint-Returning-414/m-p/320468/highlight/true#M30810
def get_deals_with_properties(project_id, get_url, hubspot_request_handler):
    param_dict_include_all_properties = {
        "includeAllProperties" : True,
        "allPropertiesFetchMode" : "latest_version",
    }

    parameters = urllib.parse.urlencode(param_dict_include_all_properties)
    url = get_url + "&" + parameters

    return hubspot_request_handler(project_id, url)


def sync_deals(project_id, refresh_token, api_key, sync_all=False):
    page_count_key = ""
    page_count = 0
    if sync_all:
        urls = [ "https://api.hubapi.com/deals/v1/deal/paged?" ]
        log.warning("Downloading all deals for project_id : "+ str(project_id) + ".")
        page_count_key = "limit"
        page_count = 250 # max size
    else:
        urls = [
            "https://api.hubapi.com/deals/v1/deal/recent/created?", # created
            "https://api.hubapi.com/deals/v1/deal/recent/modified?", # modified
        ]
        log.warning("Downloading recently created or modified deals for project_id : "+ str(project_id) + ".")
        page_count_key = "count"
        page_count = 100 # max size

    buffer_size = page_count * get_buffer_size_by_api_count()
    create_all_deal_documents_with_buffer = get_create_all_documents_with_buffer(project_id,"deal",buffer_size)
    hubspot_request_handler = get_hubspot_request_handler(project_id, refresh_token, api_key)
    deal_api_calls = 0
    err_url_too_long = False
    for url in urls:
        count = 0
        parameter_dict = {page_count_key: page_count}

        # mandatory property needed on response, returns no properties if not given.
        properties = []
        if sync_all:
            properties, ok = get_all_properties_by_doc_type(project_id,"deals", hubspot_request_handler)
            if not ok:
                log.error("Failure loading properties for project_id %d on sync_deals", project_id)
                break

        has_more = True
        while has_more:
            parameters = urllib.parse.urlencode(parameter_dict)
            get_url = url + parameters

            # List of all properties to get, returns empty properties if not given.
            if sync_all:
                get_url_with_properties = get_url + '&' + build_properties_param_str(properties)
                get_url_with_properties = get_url_with_properties + '&includeAssociations=true'
            else:
                get_url_with_properties = get_url

            log.warning("Downloading deals for project_id %d from url %s.", project_id, get_url_with_properties)
            response_dict = {}
            if err_url_too_long == False:
                r = hubspot_request_handler(project_id, get_url_with_properties)
                if not r.ok:
                    if r.status_code == 414:
                        log.error("Failure response %d from hubspot on sync_deals, using fallback logic", r.status_code)
                        err_url_too_long = True
                    else:
                        log.error("Failure response %d from hubspot on sync_deals", r.status_code)
                        break
                else:
                    response_dict = json.loads(r.text)

            if err_url_too_long == True:
                r = get_deals_with_properties(project_id, get_url, hubspot_request_handler)
                if not r.ok:
                   log.error("Failure response %d from hubspot on sync deals using fallback logic", r.status_code)
                   break
                response_dict = json.loads(r.text)

            deal_api_calls +=1

            # Need this check as has-more is not standard
            # across apis :) 
            has_more = response_dict.get('has-more')
            if has_more is None:
                has_more = response_dict.get('hasMore')

            if sync_all:
                docs = response_dict['deals']
            else:
                docs = response_dict['results']
            parameter_dict['offset']= response_dict['offset']

            count = count + len(docs)
            if allow_buffer_before_insert_by_project_id(project_id):
                create_all_deal_documents_with_buffer(docs, has_more)
                log.warning("Downloaded %d deals. total %d.", len(docs), count)
            else:
                create_all_documents(project_id, 'deal', docs)
                log.warning("Downloaded and created %d deals. total %d.", len(docs), count)
    create_all_deal_documents_with_buffer([], False) ## flush any remaining docs in memory
    return deal_api_calls


def get_company_contacts(project_id, company_id, hubspot_request_handler):
    if hubspot_request_handler == None or not company_id:
        raise Exception("invalid api_key or company_id")
    
    contacts = []
    url = "https://api.hubapi.com/companies/v2/companies/"+str(company_id)+"/contacts?"
    get_url = url
    log.warning("Downloading company contacts from url %s.", get_url)
    r = hubspot_request_handler(project_id, get_url)
    if r.status_code == 429: 
        log.error("Hubspot API rate limit exceeded for project "+str(project_id))
        return contacts
    if not r.ok:
        log.error("Failure response %d from hubspot on get_company_contacts", r.status_code)
        return contacts
    try:
        response = json.loads(r.text)
    except Exception as e:
        log.error("Failed loading json response from get_company_contacts with %s.", str(e))
        return contacts
    
    return response.get("contacts")

# Fills contacts for each company on docs.
def fill_contacts_for_companies_v1(project_id, docs, hubspot_request_handler):
    company_contacts_api_calls = 0
    for doc in docs:
        company_id = doc.get("companyId")
        contacts = get_company_contacts(project_id, company_id, hubspot_request_handler)
        company_contacts_api_calls +=1
        contactIds = []

        # Adding only contact ids as company contact list
        # properties are type inconsistent
        if contacts != None:
            for contact in contacts:
                vid = contact.get("vid")
                if vid == None: continue
                contactIds.append(vid)
        doc["contactIds"] = contactIds
    return docs, company_contacts_api_calls

def sync_companies(project_id, refresh_token, api_key,last_sync_timestamp, sync_all=False):
    if use_sync_company_v3_by_project_id(project_id):
        return sync_companies_v3(project_id, refresh_token, api_key,last_sync_timestamp, sync_all)
    return sync_companies_v2(project_id, refresh_token, api_key,last_sync_timestamp, sync_all)

def sync_companies_v2(project_id, refresh_token, api_key,last_sync_timestamp, sync_all=False):
    limit_key = ""
    if sync_all:
        urls = [ "https://api.hubapi.com/companies/v2/companies/paged?" ]
        log.warning("Downloading all companies for project_id : "+ str(project_id) + ".")
        limit_key = "limit"
    else:
        urls = [ "https://api.hubapi.com/companies/v2/companies/recent/modified?" ] # both created and modified. 
        log.warning("Downloading recently created or modified companies for project_id : "+ str(project_id) + ".")
        limit_key = "count"

    buffer_size = PAGE_SIZE * get_buffer_size_by_api_count()
    create_all_company_documents_with_buffer = get_create_all_documents_with_buffer(project_id,"company",buffer_size)

    hubspot_request_handler = get_hubspot_request_handler(project_id, refresh_token, api_key)

    companies_api_calls = 0
    companies_contacts_api_calls = 0
    max_timestamp = 0
    for url in urls:
        count = 0
        parameter_dict = {limit_key: PAGE_SIZE}

        properties = []
        if sync_all:
            properties, ok = get_all_properties_by_doc_type(project_id,"companies", hubspot_request_handler)
            if not ok:
                log.error("Failure loading properties for project_id %d on sync_companies", project_id)
                return 0, 0,0

        has_more = True
        while has_more:
            parameters = urllib.parse.urlencode(parameter_dict)
            get_url = url + parameters
            
            if sync_all:
                get_url = get_url + '&' + build_properties_param_str(properties)

            if not sync_all:
                get_url = get_url + "&since=" + str(last_sync_timestamp)

            log.warning("Downloading companies for project_id %d from url %s.", project_id, get_url)
            r = hubspot_request_handler(project_id, get_url)
            if not r.ok:
                log.error("Failure response %d from hubspot on sync_companies", r.status_code)
                break
            companies_api_calls +=1
            response_dict = json.loads(r.text)

            # Need this check as has-more is not standard
            # across apis :) 
            has_more = response_dict.get('has-more')
            if has_more is None:
                has_more = response_dict.get('hasMore')

            if sync_all:
                docs = response_dict['companies']
            else:
                docs = response_dict['results']
            parameter_dict['offset']= response_dict['offset']
            max_timestamp = get_batch_documents_max_timestamp(project_id, docs, "companies", max_timestamp)
            # fills contact ids for each comapany under 'contactIds'.
            _, api_calls = fill_contacts_for_companies(project_id, docs, hubspot_request_handler)
            companies_contacts_api_calls += api_calls
            count = count + len(docs)
            if allow_buffer_before_insert_by_project_id(project_id):
                create_all_company_documents_with_buffer(docs,has_more)
                log.warning("Downloaded %d companies. total %d.", len(docs), count)
            else:
                create_all_documents(project_id, 'company', docs)
                log.warning("Downloaded and created %d companies. total %d.", len(docs), count)
    create_all_company_documents_with_buffer([],False) ## flush any remaining docs in memory
    return companies_api_calls, companies_contacts_api_calls, max_timestamp

def get_batch_documents_max_timestamp_v3(project_id, docs, object_type, max_timestamp):
    last_modified_key = ""
    if object_type == "companies":
        last_modified_key = COMPANY_PROPERTY_KEY_LAST_MODIFIED_DATE

    for doc in docs:
        object_properties = doc[RECORD_PROPERTIES_KEY]
        if last_modified_key not in object_properties:
            log.error("Missing lastmodified in %s for project_id %d.", object_type, project_id)
            return max_timestamp
        
        last_modified_date = datetime.strptime(object_properties[last_modified_key], "%Y-%m-%dT%H:%M:%S.%fZ")
        doc_last_modified_timestamp = int(last_modified_date.timestamp())
        
        if max_timestamp== 0 :
            max_timestamp = doc_last_modified_timestamp
        elif max_timestamp < doc_last_modified_timestamp:
            max_timestamp = doc_last_modified_timestamp
    
    return max_timestamp

def fill_contacts_for_companies_v3(project_id, companies, hubspot_request_handler):
    companies_ids = []
    for company in companies:
        companies_ids.append(company["id"])
    
    associations, api_calls = get_associations(project_id, "company", companies_ids, "contact", hubspot_request_handler)
    for i in range(len(companies)):
        company_id = str(companies[i]["id"])
        if company_id in associations:
            contact_ids = []
            for id in associations[company_id]:
                contact_ids.append(int(id)) ## store as integer
            companies[i]["contactIds"] = contact_ids
    return companies, api_calls

def sync_companies_v3(project_id, refresh_token, api_key, last_sync_timestamp, sync_all=False):
    log.info("Using sync_companies_v3 for project_id : "+str(project_id)+".")

    limit = PAGE_SIZE
    if sync_all:
        url = "https://api.hubapi.com/crm/v3/objects/companies?"
        headers = None
        request = requests.get
        json_data = None
        log.warning("Downloading all companies for project_id : "+ str(project_id) + ".")
    else:
        url = "https://api.hubapi.com/crm/v3/objects/companies/search?"  # both created and modified.
        headers = {'Content-Type': 'application/json'}
        request = requests.post
        json_data = {
            "filterGroups":[
                {
                    "filters":[
                        {
                            "propertyName": COMPANY_PROPERTY_KEY_LAST_MODIFIED_DATE,
                            "operator": "GTE",
                            "value": str(last_sync_timestamp)
                        }
                    ]
                }
            ],
            "sorts": [
                {
                    "propertyName": COMPANY_PROPERTY_KEY_LAST_MODIFIED_DATE,
                    "direction": "ASCENDING"
                }
            ],
            "limit": limit
        }
        log.warning("Downloading recently created or modified companies for project_id : "+ str(project_id) + ".")

    buffer_size = PAGE_SIZE * get_buffer_size_by_api_count()
    create_all_company_documents_with_buffer = get_create_all_documents_with_buffer(project_id, "company", buffer_size)

    hubspot_request_handler = get_hubspot_request_handler(project_id, refresh_token, api_key)

    companies_api_calls = 0
    companies_contacts_api_calls = 0
    max_timestamp = 0

    count = 0
    parameter_dict = {"limit": limit}

    properties, ok = get_all_properties_by_doc_type(project_id, "companies", hubspot_request_handler)
    if not ok:
        log.error("Failure loading properties for project_id %d on sync_companies", project_id)
        return 0, 0, 0

    has_more = True
    while has_more:
        parameters = urllib.parse.urlencode(parameter_dict)
        get_url = url
        
        if sync_all:
            get_url = get_url + parameters + '&' + build_properties_param_str(properties)
        else:
            json_data["properties"] = properties

        log.warning("Downloading companies for project_id %d from url %s.", project_id, get_url)
        r = hubspot_request_handler(project_id, get_url, request=request, json=json_data, headers=headers)
        if not r.ok:
            log.error("Failure response %d from hubspot on sync_companies", r.status_code)
            break
        
        companies_api_calls +=1
        response_dict = json.loads(r.text)

        docs = response_dict['results']
        
        paging_after = ""
        if "paging" in response_dict and "next" in response_dict["paging"] and "after" in response_dict["paging"]["next"]:
            paging_after = response_dict["paging"]["next"]["after"]

        if paging_after != "":
            has_more = True
            if sync_all:
                parameter_dict["after"] = paging_after
            else:
                json_data["after"] = paging_after
        else:
            has_more = False
        
        max_timestamp = get_batch_documents_max_timestamp_v3(project_id, docs, "companies", max_timestamp)
        
        # fills contact ids for each comapany under 'contactIds'.
        _, api_calls = fill_contacts_for_companies_v3(project_id, docs, hubspot_request_handler)
        companies_contacts_api_calls += api_calls
        count = count + len(docs)
        
        if allow_buffer_before_insert_by_project_id(project_id):
            create_all_company_documents_with_buffer(docs, has_more)
            log.warning("Downloaded %d companies. total %d.", len(docs), count)
        else:
            create_all_documents(project_id, 'company', docs)
            log.warning("Downloaded and created %d companies. total %d.", len(docs), count)
    
    create_all_company_documents_with_buffer([], False) ## flush any remaining docs in memory
    return companies_api_calls, companies_contacts_api_calls, max_timestamp

def sync_forms(project_id, refresh_token, api_key):
    url = "https://api.hubapi.com/forms/v2/forms?"
    get_url = url

    hubspot_request_handler = get_hubspot_request_handler(project_id, refresh_token, api_key)

    count = 0
    log.warning("Downloading forms for project_id %d from url %s.", project_id, get_url)
    r = hubspot_request_handler(project_id, get_url)
    if not r.ok:
        log.error("Failure response %d from hubspot on sync_forms", r.status_code)
        return
    docs = json.loads(r.text)

    create_all_documents(project_id, 'form', docs)
    count = count + len(docs)
    log.warning("Downloaded and created %d forms. total %d.", len(docs), count)


def get_forms(project_id):
    uri = "/data_service/hubspot/documents/types/form?project_id="+str(project_id)
    url = options.data_service_host + uri
    log.warning("Getting form documents for project %d", project_id)

    response = requests.get(url) 
    if not response.ok:
        raise Exception('Failed to get form submissions with status '+response.status_code)

    return response.json()


# sync form_submission for all forms.
def sync_form_submissions(project_id, refresh_token, api_key):
    forms = get_forms(project_id)

    if len(forms) == 0: 
        log.warning("No forms to sync on sync_form_submissions")

    hubspot_request_handler = get_hubspot_request_handler(project_id, refresh_token, api_key)

    page_count = 50
    buffer_size = page_count * get_buffer_size_by_api_count()
    create_all_form_submissions_documents_with_buffer = get_create_all_documents_with_buffer(project_id,"form_submission",buffer_size)
    for form in forms:
        form_id = form.get("id")
        if form_id == None:
            log.warning("id is missing on from document returned by get forms for project %d", project_id)
            continue

        url = "https://api.hubapi.com/form-integrations/v1/submissions/forms/"+form_id+"?"
        parameter_dict = {"limit":50 }
        parameters = urllib.parse.urlencode(parameter_dict)
        get_url = url + parameters

        count = 0
        log.warning("Downloading form submissions for project_id %d from url %s.", project_id, get_url)
        r = hubspot_request_handler(project_id, get_url)
        if not r.ok:
            log.error("Failure response %d from hubspot on sync_form_submissions", r.status_code)
            create_all_form_submissions_documents_with_buffer([],False)
            return
        response = json.loads(r.text)
        docs = response.get("results")
        if docs == None:
            raise Exception("results key missing on form submissions api response")

        # Adding 'formId' to docs as API response 
        # doesn't have it.
        for doc in docs: doc["formId"] = form_id

        count = count + len(docs)
        if allow_buffer_before_insert_by_project_id(project_id):
            create_all_form_submissions_documents_with_buffer(docs,True)
            log.warning("Downloaded %d form submissions. total %d.", len(docs), count)
        else:
            create_all_documents(project_id, 'form_submission', docs)
            log.warning("Downloaded and created %d form submissions. total %d.", 
            len(docs), count)
    create_all_form_submissions_documents_with_buffer([],False)


def sync_owners(project_id, refresh_token, api_key):
    url = "https://api.hubapi.com/owners/v2/owners?"
    get_url = url

    hubspot_request_handler = get_hubspot_request_handler(project_id, refresh_token, api_key)

    buffer_size = PAGE_SIZE * get_buffer_size_by_api_count()
    create_all_owners_documents_with_buffer = get_create_all_documents_with_buffer(project_id, "owner", buffer_size)

    count = 0
    log.warning("Downloading owners for project_id %d from url %s.", project_id, get_url)
    r = hubspot_request_handler(project_id, get_url)
    if not r.ok:
        log.error("Failure response %d from hubspot on sync_owners", r.status_code)
        return
    docs = json.loads(r.text)

    create_all_owners_documents_with_buffer(docs, False)
    count = count + len(docs)
    log.warning("Downloaded and created %d owners. total %d.", len(docs), count)


def get_sync_info(sync_first_time = False):
    uri = "/data_service/hubspot/documents/sync_info?is_first_time="
    if sync_first_time == True:
        uri = uri + "true"
    else:
        uri = uri + "false"
    
    url = options.data_service_host + uri
    response = requests.get(url)
    if not response.ok:
        raise Exception('Failed to get sync info with status: '+str(response.status_code))
    return response.json()

def update_sync_status(request_payload, first_sync=False):
    uri = "/data_service/hubspot/documents/sync_info?"
    if first_sync:
        uri = uri+"is_first_time=true"
    else:
        uri = uri+"is_first_time=false"

    url = options.data_service_host + uri

    retries = 0
    while True:
        try:
            res = requests.post(url, data=json.dumps(request_payload))
            if not res.ok:
                raise Exception('Failed to send first time sync update: ' + str(res.status_code))
            return
        except Exception as e:
            if retries > RETRY_LIMIT:
                raise Exception("Retry exhausted on send first time sync update "+str(e)+ " , retries "+ str(retries))
            log.warning("Failed to send first time sync update. retrying in 2s "+str(e))
            retries += 1
            time.sleep(2)

def get_allowed_list_with_all_element_support(allowed_list_string, disallowed_list_string=""):
    disallowed_map = {}
    elements = [s.strip() for s in disallowed_list_string.split(",")]
    for element in elements:
        disallowed_map[element]= True

    if allowed_list_string =="*":
        return True, {},disallowed_map

    elements = [s.strip() for s in allowed_list_string.split(",")]
    allowed_map = {}
    for element in elements:
        if element not in disallowed_map:
            allowed_map[element]=True

    return False,allowed_map,disallowed_map

def allow_sync_by_project_id(project_id):
    all_projects, allowed_projects, disallowed_projects = get_allowed_list_with_all_element_support(options.project_ids, options.disabled_project_ids)
    if str(project_id) in disallowed_projects:
        return False

    if not all_projects:
        return str(project_id) in allowed_projects

    return True

def get_hubspot_app_credentials():
    return options.hubspot_app_id, options.hubspot_app_secret

def disable_non_marketing_contacts_by_project_id(project_id):
    all_projects, allowed_projects, _ = get_allowed_list_with_all_element_support(options.disable_non_marketing_contacts_project_id)

    if all_projects:
        return True
    return str(project_id) in allowed_projects

def get_buffer_size_by_api_count():
    return int(options.buffer_size_by_api_count)

def allow_buffer_before_insert_by_project_id(project_id):
    all_projects, allowed_projects,_ = get_allowed_list_with_all_element_support(options.enable_buffer_before_insert_by_project_id)
    if all_projects:
        return True
    return str(project_id) in allowed_projects

def allow_delete_api_by_project_id(project_id):
    if not options.enable_deleted_contacts:
        return False
    all_projects, allowed_projects,_ = get_allowed_list_with_all_element_support(options.enable_deleted_projectIDs)
    if all_projects:
        return True
    return str(project_id) in allowed_projects

def use_company_contact_association_v2_by_project_id(project_id):
    if not options.enable_company_contact_association_v2_by_project_id:
        return False
    all_projects, allowed_projects,_ = get_allowed_list_with_all_element_support(options.enable_company_contact_association_v2_by_project_id)
    if all_projects:
        return True
    return str(project_id) in allowed_projects

def allow_batch_insert_doc_type(doc_type):
    all_doc_type, allowed_doc_type,_ = get_allowed_list_with_all_element_support(options.batch_insert_doc_types)
    if all_doc_type:
        return True
    return str(doc_type) in allowed_doc_type

def allow_batch_insert_by_project_id(project_id):
    all_projects, allowed_projects,_ = get_allowed_list_with_all_element_support(options.batch_insert_by_project_ids)
    if all_projects:
        return True
    return str(project_id) in allowed_projects

def allow_contact_list_sync_by_project_id(project_id):
    if not options.enable_contact_list_sync_by_project_id:
        return False
    all_projects, allowed_projects,_ = get_allowed_list_with_all_element_support(options.enable_contact_list_sync_by_project_id)
    if all_projects:
        return True
    return str(project_id) in allowed_projects

def allow_owner_sync_by_project_id(project_id):
    if not options.enable_owner_sync_by_project_id:
        return False
    all_projects, allowed_projects,_ = get_allowed_list_with_all_element_support(options.enable_owner_sync_by_project_id)
    if all_projects:
        return True
    return str(project_id) in allowed_projects

def allowed_doc_types_sync(doc_type):
    all_doc_typ, allowed_doc_types,_ = get_allowed_list_with_all_element_support(options.allowed_doc_types_sync)
    if all_doc_typ:
        return True
    return doc_type in allowed_doc_types

def use_sync_company_v3_by_project_id(project_id):
    if not options.enable_sync_company_v3_by_project_id:
        return False
    all_projects, allowed_projects,_ = get_allowed_list_with_all_element_support(options.enable_sync_company_v3_by_project_id)
    if all_projects:
        return True
    return str(project_id) in allowed_projects

def get_next_sync_info(project_settings, last_sync_info, first_time_sync = False):
    next_sync_info = []
    for project_id in project_settings:
        if not allow_sync_by_project_id(project_id):
            continue

        settings = project_settings[project_id]
        if first_time_sync == True and settings.get("is_first_time_synced")!=False :
            continue

        if first_time_sync == False and settings.get("is_first_time_synced")!=True:
            continue

        api_key = settings.get("api_key")
        refresh_token = settings.get("refresh_token")
        if api_key == None and refresh_token == None:
            log.error("No api_key and refresh_token on project_settings of project %d", project_id)
            continue

        sync_info = last_sync_info.get(project_id)
        if sync_info == None:
            log.error("Last sync info missing for project %d", project_id)
            continue

        if allow_delete_api_by_project_id(project_id):
             sync_info["deleted_contacts"] = 0

        for doc_type in sync_info:
            if not allowed_doc_types_sync(doc_type):
                continue

            if doc_type == "contact_list" and not allow_contact_list_sync_by_project_id(project_id):
                continue
            if doc_type == "owner" and not allow_owner_sync_by_project_id(project_id):
                continue
            
            next_sync = {}
            next_sync["project_id"] = int(project_id)
            next_sync["api_key"] = api_key
            next_sync["doc_type"] = doc_type
            next_sync["refresh_token"] = refresh_token if refresh_token is not None else ""
            # sync all, if last sync timestamp is 0.
            next_sync["sync_all"] = first_time_sync
            next_sync["last_sync_timestamp"] = sync_info[doc_type]
            next_sync_info.append(next_sync)

    return next_sync_info 


def sync(project_id, refresh_token, api_key, doc_type, sync_all, last_sync_timestamp):
    response = {}
    response["project_id"] = project_id
    response["doc_type"] = doc_type
    response["sync_all"] = sync_all

    max_timestamp  = 0
    try:
        if project_id == None or api_key == None or doc_type == None or sync_all == None:
            raise Exception("invalid params on sync, project_id "+str(project_id)+", api_key "+str(api_key)+", doc_type "+str(doc_type)+", sync_all "+str(sync_all))            

        if doc_type == "contact":
            response["contact_api_calls"],max_timestamp = sync_contacts(project_id, refresh_token, api_key, last_sync_timestamp, sync_all)
        elif doc_type == "company":        
            response["companies_api_calls"], response["companies_contacts_api_calls"],max_timestamp = sync_companies(project_id, refresh_token, api_key, last_sync_timestamp, sync_all)
        elif doc_type == "deal":
            response["deal_api_calls"] = sync_deals(project_id, refresh_token, api_key, sync_all)
        elif doc_type == "contact_list":
            response["contact_list_api_calls"], max_timestamp = sync_all_contact_lists_v2(project_id, refresh_token, api_key, last_sync_timestamp)
        elif doc_type == "form":
            sync_forms(project_id, refresh_token, api_key)
        elif doc_type == "form_submission":
            sync_form_submissions(project_id, refresh_token, api_key)
        elif doc_type == "owner":
            sync_owners(project_id, refresh_token, api_key)
        elif doc_type == "deleted_contacts":
            response["deleted_contacts_api_calls"] = get_deleted_contacts(project_id, refresh_token, api_key)
        elif doc_type == "engagement":
            response["engagement_api_calls"],max_timestamp = sync_engagements(project_id, refresh_token, api_key, last_sync_timestamp)
        else:
            raise Exception("invalid doc_type "+ doc_type)

    except Exception as e:
        if str(e) == "Same offset for consecutive pages on sync_contacts":
            response["message"] = str(e)
        else:    
            response["status"] = "failed"
            response["message"] = "Failed with exception: " + str(e)
            return response

    response["status"] = "success"
    response["timestamp"]= max_timestamp
    return response

def requests_with_retry(method,url):
    retries = 0
    while True:
        try:
            return requests.request(method=method, url = url)
        except Exception as e:
            if retries < RETRY_LIMIT:
                log.error("Failed to perform request %s. Retrying in %dsec",e,2)
                time.sleep(2)
                continue
            else:
                raise Exception("Failed to perform request %s",e)

def get_task_detail(job_name):
    uri = "/data_service/task/details?task_name="+job_name
    url = options.data_service_host + uri

    response = requests_with_retry("GET",url)
    if not response.ok:
        raise Exception('Failed to get task details: '+str(response.status_code)+' %s'+response.text)
    return response.json()

def get_all_to_be_executed_deltas(task_id,project_id, lookback):
    uri = "/data_service/task/deltas?task_id="+str(task_id) +"&lookback="+str(lookback) +"&project_id="+str(project_id)
    url = options.data_service_host + uri

    response = requests_with_retry("GET",url)
    if not response.ok:
        raise Exception('Failed to task deltas: '+str(response.status_code)+' %s'+response.text)
    return response.json()

def insert_task_begin_record(task_id,project_id, delta):
    uri = "/data_service/task/begin?task_id="+str(task_id) +"&delta="+str(delta) +"&project_id="+str(project_id)
    url = options.data_service_host + uri

    response = requests_with_retry("POST",url)
    if response.status_code != requests.codes["created"]:
        raise Exception('Failed to insert task begin record: '+str(response.status_code)+' %s'+response.text)
    return response

def insert_task_end_record(task_id,project_id, delta):
    uri = "/data_service/task/end?task_id="+str(task_id) +"&delta="+str(delta) +"&project_id="+str(project_id)
    url = options.data_service_host + uri

    response = requests_with_retry("POST",url)
    if response.status_code != requests.codes["created"]:
        raise Exception('Failed to insert task end record: '+str(response.status_code)+' %s'+response.text)
    return response

def delete_task_end_record(task_id,project_id, delta):
    uri = "/data_service/task/end?task_id="+str(task_id) +"&delta="+str(delta) +"&project_id="+str(project_id)
    url = options.data_service_host + uri

    response = requests_with_retry("DELETE",url)
    if response.status_code != requests.codes["accepted"]:
        raise Exception('Failed to delete task delta: '+str(response.status_code)+' %s'+response.text)
    return response

def is_dependent_task_done(task_id,project_id,delta):
    uri = "/data_service/task/dependent_task_done?task_id="+str(task_id) +"&delta="+str(delta) +"&project_id="+str(project_id)
    url = options.data_service_host + uri

    response = requests_with_retry("GET",url)
    if not response.ok and response.status_code != requests.codes["not_found"]:
        raise Exception('Failed to check dependent task: '+str(response.status_code)+' %s'+response.text)
    data = json.loads(response.text)
    return data

def get_task_delta_as_time(delta):
    uri = "/data_service/task/delta_timestamp?delta="+str(delta)
    url = options.data_service_host + uri

    response = requests_with_retry("GET",url)
    if not response.ok:
        raise Exception('Failed to get timestamp for delta: '+delta+" status: "+str(response.status_code)+' %s'+response.text)
    return response

def get_task_end_timestamp(delta,frequency, frequency_interval):
    uri = "/data_service/task/delta_end_timestamp?delta="+str(delta)+"&frequency="+str(frequency)+"&frequency_interval="+frequency_interval
    url = options.data_service_host + uri

    response = requests_with_retry("GET",url)
    if not response.ok:
        raise Exception("Failed to get end timestamp for delta: "+delta+" status: "+str(response.status_code)+' %s'+response.text)
    return response

def get_pending_delta(job_name, lookback):
    task_details = get_task_detail(job_name)
    deltas = get_all_to_be_executed_deltas(task_details["task_id"],lookback)
    if len(deltas)<1:
        return task_details,0,False

    return task_details,deltas[len(deltas)-1], True # only process the latest delta

def hubspot_sync(configs):
    first_sync = configs["first_sync"]
    sync_info = get_sync_info(first_sync)

    project_settings = sync_info.get("project_settings")
    if project_settings == None:
        log.error("Project settings missing on get sync info response")
        sys.exit(1)

    last_sync_info = sync_info.get("last_sync_info")
    if last_sync_info == None:
        log.error("Last sync info missing on get sync info response")
        sys.exit(1)

    next_sync_info = get_next_sync_info(project_settings, last_sync_info, first_sync)

    log.warning("sync_info: "+str(next_sync_info))
    next_sync_failures = []
    next_sync_success = []

    for info in next_sync_info:
        log.warning("Current processing sync_info: "+str(info))
        notification_payload = {}
        try:
            response = sync(info.get("project_id"), info.get("refresh_token"), info.get("api_key"),
                    info.get("doc_type"), info.get("sync_all"), info.get("last_sync_timestamp"))
            if response["status"] == "failed":
                next_sync_failures.append(response)
                notification_payload["status"]= "Failures on sync."
                notification_payload["failures"] = [response]
                notification_payload["success"] = []
            else:
                next_sync_success.append(response)
                notification_payload["status"]= "Successfully synced."
                notification_payload["failures"] = []
                notification_payload["success"] = [response]
            
            update_sync_status(notification_payload, first_sync)
        except Exception as e:
            project_id = info.get("project_id")
            doc_type = info.get("doc_type")
            log.warning("Failed to process doc type %s for project_id %d. exception: %s", doc_type, project_id, str(e))
            next_sync_failures.append(str(project_id)+":"+doc_type+" -> "+str(e))

    success = len(next_sync_failures)<1
    status_msg = "Successfully synced." if success else "Failures on sync."

    notification_payload = {
        "status": status_msg, 
        "failures": next_sync_failures, 
        "success": next_sync_success,
    }

    log.warning("Successfully synced. End of hubspot sync job.")

    try:
        if first_sync == True:
            ping_healthcheck(options.env, options.healthcheck_ping_id, notification_payload)
        else:
            if len(next_sync_failures) > 0:
                ping_healthcheck(options.env, HEALTHCHECK_PING_ID, notification_payload, endpoint="/fail")
            else:
                ping_healthcheck(options.env, HEALTHCHECK_PING_ID, notification_payload)

    except Exception as e:
        log.warning(e)
        next_sync_failures.append(str(e))
        if first_sync == True:
            ping_healthcheck(options.env, options.healthcheck_ping_id, notification_payload, endpoint="/fail")
        else:
            ping_healthcheck(options.env, HEALTHCHECK_PING_ID, notification_payload, endpoint="/fail")
    
    ping_healthcheck(options.env, HEALTHCHECK_RUN_PING_ID, {})

    return None, True

def task_func(job_name, lookback, f, configs, latest_interval=False):
    task_details = get_task_detail(job_name)
    finalStatus = {}
    if task_details["is_project_enabled"] == True:
        finalStatus["status"] = "Call ProjectId Enabled Func"
        return finalStatus
    task_id = task_details["task_id"]

    deltas  = get_all_to_be_executed_deltas(task_id, 0,lookback)
    if len(deltas)<1:
        log.warning("No interval for processing.")
        return finalStatus

    deltas.sort()
    if latest_interval == True:
        deltas = deltas[len(deltas)-1:]
    log.warning("Deltas to be processed %s", deltas)

    for delta in deltas:
        finalStatus[str(delta)]={}
        log.warning("Checking dependency")
        done = is_dependent_task_done(task_id,0,delta)
        if done == True:
            log.warning("Processing delta %s",delta)
            insert_task_begin_record(task_id,0,delta)
            configs["start_timestamp"] = get_task_delta_as_time(delta)
            configs["end_timestamp"]=get_task_end_timestamp(delta,str(task_details["frequency"]), str(task_details["frequency_interval"]))
            try:
                status, success = f(configs)
            except Exception as e:
                delete_task_end_record(task_id,0, delta)
                finalStatus[str(delta)]["success"] = False
                finalStatus[str(delta)]["error"] = str(e)
                break
            finalStatus[str(delta)]["status"] = status
            finalStatus[str(delta)]["success"] = success
            if success == False:
                log.warning("Processing failed for delta %s",delta)
                delete_task_end_record(task_id,0, delta)
                break
            log.warning("Processing success for delta %s",delta)
            insert_task_end_record(task_id,0,delta)
        else:
            finalStatus[str(delta)]["error"] = "Dependency not done yet"
            finalStatus[str(delta)]["success"]=False
            log.warning("%s - dependency not done yet",delta)

    return finalStatus


if __name__ == "__main__":
    (options, args) = parser.parse_args()
    configs = {
        "first_sync":options.first_sync
    }

    ## prioritize app_name from environment
    app_name = options.app_name if options.app_name != "" else APP_NAME
    status = task_func(app_name,1,hubspot_sync,configs,True)
    if len(status)<1:
        sys.exit(0)
    
    err = ""
    for delta in status:
        delta_status = status[delta]

        if "error" in delta_status:
            err = delta_status["error"]

    if err!="":
        ping_healthcheck(options.env, options.healthcheck_ping_id, err, endpoint="/fail")

    sys.exit(0)