from optparse import OptionParser
import logging as log
from types import TracebackType
import requests
import json
import urllib
import sys
import time

from requests import status_codes

parser = OptionParser()
parser.add_option("--env", dest="env", default="development")
parser.add_option("--dry", dest="dry", help="", default="False")
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
parser.add_option("--project_ids", dest="project_ids", help="Allowed project_ids", default="")


APP_NAME = "hubspot_sync"
PAGE_SIZE = 100
DOC_TYPES = [ "contact", "company", "deal", "form", "form_submission" ]

METRIC_TYPE_INCR = "incr"
HEALTHCHECK_PING_ID = "87137001-b18b-474c-8bc5-63324baff2a8"

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

def create_document_in_batch(project_id, doc_type, documents, fetch_deleted_contact=False):
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

def create_all_documents(project_id, doc_type, docs, fetch_deleted_contact=False):
    if options.dry == "True":
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

def get_all_properties_by_doc_type(project_id,doc_type, api_key):
    url = "https://api.hubapi.com/properties/v1/"+doc_type+"/properties?"
    parameter_dict = { 'hapikey': api_key }
    parameters = urllib.parse.urlencode(parameter_dict)
    get_url = url + parameters
    r = get_with_fallback_retry(project_id, get_url)
    if not r.ok:
        log.error("Failure response %d from hubspot on get_properties_by_doc_type for doc type %s", r.status_code, doc_type)
        return [], r.ok

    response_dict = json.loads(r.text)
    properties = []
    for contact_property in response_dict:
        properties.append(contact_property["name"])
    return properties, r.ok

def get_with_fallback_retry(project_id, get_url):
    retries = 0
    start_time  = time.time()
    try:
        while True:
            try:
                r = requests.get(url=get_url, headers = {}, timeout=REQUEST_TIMEOUT)
                if r.status_code != 429:
                    if not r.ok:
                        if r.status_code== 414:
                            return r
                        if retries < RETRY_LIMIT:
                            log.error("Failed to get data from hubspot %d.Retries %d. Retrying in 2 seconds",r.status_code,retries)
                            time.sleep(2)
                            retries += 1
                            continue
                        log.error("Retry exhausted. Failed to get data after %d retries",retries)
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

def get_contacts_with_properties_by_id(project_id,api_key,get_url):
    batch_contact_url = "https://api.hubapi.com/contacts/v1/contact/vids/batch?"

    log.warning("Downloading contacts without properties list "+get_url)
    r = get_with_fallback_retry(project_id,get_url)
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
    batch_url = batch_contact_url +"hapikey="+ api_key + "&" + contact_ids_str
    log.warning("Downloading batch contact from url "+batch_url)
    r = get_with_fallback_retry(project_id,batch_url)
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


def sync_contacts(project_id, api_key,last_sync_timestamp, sync_all=False):
    if sync_all:
        # init sync all contacts.
        url = "https://api.hubapi.com/contacts/v1/lists/all/contacts/all?"
        log.warning("Downloading all contacts for project_id : "+ str(project_id) + ".")
    else:
        # sync recently updated and created contacts.
        url = "https://api.hubapi.com/contacts/v1/lists/recently_updated/contacts/recent?"
        log.warning("Downloading recently created or modified contacts for project_id : "+ str(project_id) + ".")

    has_more = True
    count = 0
    parameter_dict = { 'hapikey': api_key, 'count': PAGE_SIZE }
    properties, ok = get_all_properties_by_doc_type(project_id,"contacts", api_key)
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
            r = get_with_fallback_retry(project_id,get_url_with_properties)
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
            contact_dict,unmodified_response_dict, r = get_contacts_with_properties_by_id(project_id,api_key,get_url)
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
        create_all_documents(project_id, 'contact', docs)
        count = count + len(docs)
        log.warning("Downloaded and created %d contacts. total %d.", len(docs), count)
    return contact_api_calls, max_timestamp

def get_deleted_contacts(project_id, api_key):
    url = "https://api.hubapi.com/crm/v3/objects/contacts/?"
    parameter_dict = {'archived': "true", 'hapikey': api_key,"limit":100}
    parameters = urllib.parse.urlencode(parameter_dict)
    final_url = url + parameters    
    has_more = True
    count_api_calls = 0
    while has_more:
        log.warning("Downloading deleted contacts from url: %s", final_url)
        res = get_with_fallback_retry(project_id, final_url)
        if not res.ok:
            raise Exception('Failed to get deleted contacts: ' + str(res.status_code))

        response_dict = json.loads(res.text)
        count_api_calls+=1
        docs = response_dict.get('results')
        if not docs:
            raise Exception('Found empty response for deleted_contacts')
        create_all_documents(project_id, 'contact', docs, True)
        has_more = response_dict.get('paging')
        if has_more:
            next_property = has_more.get('next')
            if not next_property:
                continue
            next_link = next_property.get('link')
            if not next_link:
                raise Exception('Found empty link value to fetch deleted contacts')
            final_url = next_link + "&" + "hapikey=" + api_key
    return count_api_calls

## https://community.hubspot.com/t5/APIs-Integrations/Deals-Endpoint-Returning-414/m-p/320468/highlight/true#M30810
def get_deals_with_properties(project_id,get_url):
    param_dict_include_all_properties = {
        "includeAllProperties" : True,
        "allPropertiesFetchMode" : "latest_version",
    }

    parameters = urllib.parse.urlencode(param_dict_include_all_properties)
    url = get_url + "&" + parameters

    return get_with_fallback_retry(project_id, url)


def sync_deals(project_id, api_key, sync_all=False):
    if sync_all:
        urls = [ "https://api.hubapi.com/deals/v1/deal/paged?" ]
        log.warning("Downloading all deals for project_id : "+ str(project_id) + ".")
    else:
        urls = [
            "https://api.hubapi.com/deals/v1/deal/recent/created?", # created
            "https://api.hubapi.com/deals/v1/deal/recent/modified?", # modified
        ]
        log.warning("Downloading recently created or modified deals for project_id : "+ str(project_id) + ".")

    deal_api_calls = 0
    err_url_too_long = False
    for url in urls:
        count = 0
        parameter_dict = {'hapikey': api_key, 'limit': PAGE_SIZE}

        # mandatory property needed on response, returns no properties if not given.
        properties = []
        if sync_all:
            properties, ok = get_all_properties_by_doc_type(project_id,"deals", api_key)
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
                r = get_with_fallback_retry(project_id, get_url_with_properties)
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
                r = get_deals_with_properties(project_id,get_url)
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

            create_all_documents(project_id, 'deal', docs)
            count = count + len(docs)
            log.warning("Downloaded and created %d deals. total %d.", len(docs), count)
    return deal_api_calls


def get_company_contacts(project_id, api_key, company_id):
    if api_key == "" or not company_id:
        raise Exception("invalid api_key or company_id")
    
    contacts = []
    url = "https://api.hubapi.com/companies/v2/companies/"+str(company_id)+"/contacts?"
    parameter_dict = { 'hapikey': api_key }
    parameters = urllib.parse.urlencode(parameter_dict)
    get_url = url + parameters
    log.warning("Downloading company contacts from url %s.", get_url)
    r = get_with_fallback_retry(project_id, get_url)
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
def fill_contacts_for_companies(project_id, api_key, docs):
    company_contacts_api_calls = 0
    for doc in docs:
        company_id = doc.get("companyId")
        contacts = get_company_contacts(project_id, api_key, company_id)
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

def sync_companies(project_id, api_key,last_sync_timestamp, sync_all=False):
    if sync_all:
        urls = [ "https://api.hubapi.com/companies/v2/companies/paged?" ]
        log.warning("Downloading all companies for project_id : "+ str(project_id) + ".")
    else:
        urls = [ "https://api.hubapi.com/companies/v2/companies/recent/modified?" ] # both created and modified. 
        log.warning("Downloading recently created or modified companies for project_id : "+ str(project_id) + ".")

    companies_api_calls = 0
    companies_contacts_api_calls = 0
    max_timestamp = 0
    for url in urls:
        count = 0
        parameter_dict = {'hapikey': api_key, 'limit': PAGE_SIZE}

        properties = []
        if sync_all:
            properties, ok = get_all_properties_by_doc_type(project_id,"companies", api_key)
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
            r = get_with_fallback_retry(project_id, get_url)
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
            _, companies_contacts_api_calls = fill_contacts_for_companies(project_id, api_key, docs)
            create_all_documents(project_id, 'company', docs)
            count = count + len(docs)
            log.warning("Downloaded and created %d companies. total %d.", len(docs), count)
    return companies_api_calls, companies_contacts_api_calls, max_timestamp

def sync_forms(project_id, api_key):
    url = "https://api.hubapi.com/forms/v2/forms?"
    parameter_dict = {'hapikey': api_key }
    parameters = urllib.parse.urlencode(parameter_dict)
    get_url = url + parameters

    count = 0
    log.warning("Downloading forms for project_id %d from url %s.", project_id, get_url)
    r = get_with_fallback_retry(project_id, get_url)
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
def sync_form_submissions(project_id, api_key):
    forms = get_forms(project_id)

    if len(forms) == 0: 
        log.warning("No forms to sync on sync_form_submissions")

    for form in forms:
        form_id = form.get("id")
        if form_id == None:
            log.warning("id is missing on from document returned by get forms for project %d", project_id)
            continue

        url = "https://api.hubapi.com/form-integrations/v1/submissions/forms/"+form_id+"?"
        parameter_dict = { 'hapikey': api_key }
        parameters = urllib.parse.urlencode(parameter_dict)
        get_url = url + parameters
        
        count = 0
        log.warning("Downloading form submissions for project_id %d from url %s.", project_id, get_url)
        r = get_with_fallback_retry(project_id, get_url)
        if not r.ok:
            log.error("Failure response %d from hubspot on sync_form_submissions", r.status_code)
            return
        response = json.loads(r.text)
        docs = response.get("results")
        if docs == None:
            raise Exception("results key missing on form submissions api response")

        # Adding 'formId' to docs as API response 
        # doesn't have it.
        for doc in docs: doc["formId"] = form_id

        create_all_documents(project_id, 'form_submission', docs)
        count = count + len(docs)
        log.warning("Downloaded and created %d form submisssions. total %d.", 
            len(docs), count)


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

def get_allowed_list_with_all_element_support(list_string):
    if list_string =="*":
        return True, {}
    elements = [s.strip() for s in list_string.split(",")]
    elements_map = {}
    for element in elements:
        elements_map[element] = True
    return False,elements_map

def allow_sync_by_project_id(project_id):
    all_projects, allowed_projects = get_allowed_list_with_all_element_support(options.project_ids)
    if all_projects:
        return True
    return str(project_id) in allowed_projects

def allow_delete_api_by_project_id(project_id):
    if not options.enable_deleted_contacts:
        return False
    all_projects, allowed_projects = get_allowed_list_with_all_element_support(options.enable_deleted_projectIDs)
    if all_projects:
        return True
    return str(project_id) in allowed_projects

def allow_batch_insert_doc_type(doc_type):
    all_doc_type, allowed_doc_type = get_allowed_list_with_all_element_support(options.batch_insert_doc_types)
    if all_doc_type:
        return True
    return str(doc_type) in allowed_doc_type

def allow_batch_insert_by_project_id(project_id):
    all_projects, allowed_projects = get_allowed_list_with_all_element_support(options.batch_insert_by_project_ids)
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
        if api_key == None:
            log.error("No api_key on project_settings of project %d", project_id)
            continue
        
        sync_info = last_sync_info.get(project_id)
        if sync_info == None:
            log.error("Last sync info missing for project %d", project_id)
            continue

        if allow_delete_api_by_project_id(project_id):
             sync_info["deleted_contacts"] = 0

        for doc_type in sync_info:
            next_sync = {}
            next_sync["project_id"] = int(project_id)
            next_sync["api_key"] = api_key
            next_sync["doc_type"] = doc_type
            # sync all, if last sync timestamp is 0.
            next_sync["sync_all"] = first_time_sync
            next_sync["last_sync_timestamp"] = sync_info[doc_type]
            next_sync_info.append(next_sync)

    return next_sync_info 


def sync(project_id, api_key, doc_type, sync_all, last_sync_timestamp):
    response = {}
    response["project_id"] = project_id
    response["doc_type"] = doc_type
    response["sync_all"] = sync_all

    max_timestamp  = 0
    try:
        if project_id == None or api_key == None or doc_type == None or sync_all == None:
            raise Exception("invalid params on sync, project_id "+str(project_id)+", api_key "+str(api_key)+", doc_type "+str(doc_type)+", sync_all "+str(sync_all))            

        if doc_type == "contact":
            response["contact_api_calls"],max_timestamp = sync_contacts(project_id, api_key, last_sync_timestamp, sync_all)
        elif doc_type == "company":        
            response["companies_api_calls"], response["companies_contacts_api_calls"],max_timestamp = sync_companies(project_id, api_key, last_sync_timestamp, sync_all)
        elif doc_type == "deal":
            response["deal_api_calls"] = sync_deals(project_id, api_key, sync_all)
        elif doc_type == "form":
            sync_forms(project_id, api_key)
        elif doc_type == "form_submission":
            sync_form_submissions(project_id, api_key)
        elif doc_type == "deleted_contacts":
            response["deleted_contacts_api_calls"] = get_deleted_contacts(project_id, api_key)
        else:
            raise Exception("invalid doc_type "+ doc_type)

    except Exception as e:
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
            response = sync(info.get("project_id"), info.get("api_key"),
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

    sync_status = {
        "next_sync_failures":next_sync_failures,
        "next_sync_success": next_sync_success
    }

    success = len(next_sync_failures)<1
    return sync_status, success

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

    app_name = options.app_name if options.first_sync else APP_NAME
    status = task_func(app_name,1,hubspot_sync,configs,True)
    if len(status)<1:
        sys.exit(0)
    
    status_msg = ""
    err = ""
    next_sync_failures = []
    next_sync_success = []
    for delta in status:
        delta_status = status[delta]

        if delta_status["success"] == False:
            status_msg = "Failures on sync."
            err = delta_status["error"] if "error" in delta_status else ""
        else:
            status_msg = "Successfully synced."

        if "status" in delta_status:
            if "next_sync_failures" in delta_status["status"]: next_sync_failures = delta_status["status"]["next_sync_failures"]
            if "next_sync_success" in delta_status["status"]: next_sync_success = delta_status["status"]["next_sync_success"]

    notification_payload = {
        "status": status_msg, 
        "failures": next_sync_failures, 
        "success": next_sync_success,
    }

    log.warning("Successfully synced. End of hubspot sync job.")
    if err!="": # append error after processing data
            next_sync_failures.insert(0,err)
    
    try:
        if options.first_sync == True:
            ping_healthcheck(options.env, options.healthcheck_ping_id, notification_payload)
        else:
            if len(next_sync_failures) > 0:
                ping_healthcheck(options.env, HEALTHCHECK_PING_ID, notification_payload, endpoint="/fail")
            else:
                ping_healthcheck(options.env, HEALTHCHECK_PING_ID, notification_payload)

    except Exception as e:
        log.warning(e)
        next_sync_failures.append(str(e))
        if options.first_sync == True:
            ping_healthcheck(options.env, options.healthcheck_ping_id, notification_payload, endpoint="/fail")
        else:
            ping_healthcheck(options.env, HEALTHCHECK_PING_ID, notification_payload, endpoint="/fail")

    sys.exit(0)