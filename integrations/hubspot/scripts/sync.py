from optparse import OptionParser
import logging as log
import requests
import json
import urllib
import sys

parser = OptionParser()
parser.add_option("--env", dest="env", default="development")
parser.add_option("--dry", dest="dry", help="", default="False")
parser.add_option("--data_service_host", dest="data_service_host",
    help="Data service host", default="http://localhost:8089")

APP_NAME = "hubspot_sync"
PAGE_SIZE = 50
DOC_TYPES = [ "contact", "company", "deal", "form", "form_submission" ]

METRIC_TYPE_INCR = "incr"
HEALTHCHECK_PING_ID = "87137001-b18b-474c-8bc5-63324baff2a8"

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
    if env != "production": 
        log.warning("Skipped healthcheck ping for env %s payload %s", env, str(message))
        return

    try:
        requests.post("https://hc-ping.com/" + healthcheck_id + endpoint,
            data=json.dumps(message, indent=1), timeout=10)
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

def create_document(project_id, doc_type, doc):
    uri = "/data_service/hubspot/documents/add"
    url = options.data_service_host + uri

    payload = {
        "project_id": project_id,
        "type_alias": doc_type,
        "value": doc,
    }

    response = requests.post(url, json=payload)
    if not response.ok:
        log.error("Failed to add response %s to hubspot warehouse with uri %s: %d.", 
            doc_type, uri, response.status_code)
    
    return response

def create_all_documents(project_id, doc_type, docs):
    if options.dry == "True":
        log.warning("Dry run. Skipped document upsert.")
        return

    for doc in docs:
        create_document(project_id, doc_type, doc)

def build_properties_param_str(properties=[]):
    param_str = ''
    for prop in properties:
        if param_str != '':
            param_str = param_str + '&'
        param_str = param_str + 'properties=' + prop
    return param_str

def get_all_properties_by_doc_type(doc_type, api_key):
    url = "https://api.hubapi.com/properties/v1/"+doc_type+"/properties?"
    parameter_dict = { 'hapikey': api_key }
    parameters = urllib.parse.urlencode(parameter_dict)
    get_url = url + parameters
    r = requests.get(url= get_url, headers = {})
    if not r.ok:
        log.error("Failure response %d from hubspot on get_properties_by_doc_type for doc type %s", r.status_code, doc_type)
        return [], r.ok

    response_dict = json.loads(r.text)
    properties = []
    for contact_property in response_dict:
        properties.append(contact_property["name"])
    return properties, r.ok


def sync_contacts(project_id, api_key, sync_all=False):    
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
    properties, ok = get_all_properties_by_doc_type("contacts", api_key)
    if not ok:
        log.error("Failure loading properties for project_id %d on sync_contacts", project_id)
        return

    while has_more:
        parameters = urllib.parse.urlencode(parameter_dict)
        get_url = url + parameters

        # contacts api uses property instead of properties in query parameter
        properties_str = "&".join([ "property="+property_name for property_name in properties ])
        get_url = get_url + '&' + properties_str

        log.warning("Downloading contacts for project_id %d from url %s.", project_id, get_url)
        r = requests.get(url= get_url, headers = {})
        if r.status_code == 429:
            raise Exception("hubspot api rate limit exceeded for project "+str(project_id))
        if not r.ok:
            log.error("Failure response %d from hubspot on sync_contacts", r.status_code)
            break
        response_dict = json.loads(r.text)

        has_more = response_dict['has-more']
        docs = response_dict['contacts']
        if sync_all:
            parameter_dict['vidOffset'] = response_dict['vid-offset']
        else:
            if parameter_dict.get('timeOffset') == response_dict.get('time-offset'):
                log.warning("same offset on consective pages on recent contacts sync %s, %s and has more is %s", 
                    parameter_dict.get('timeOffset'), response_dict.get('time-offset'), response_dict.get('has-more'))
                raise Exception("Same offset for consecutive pages on sync_contacts")
            parameter_dict['timeOffset'] = response_dict['time-offset']

        create_all_documents(project_id, 'contact', docs)
        count = count + len(docs)
        log.warning("Downloaded and created %d contacts. total %d.", len(docs), count)


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

    for url in urls:
        count = 0
        parameter_dict = {'hapikey': api_key, 'limit': PAGE_SIZE}

        # mandatory property needed on response, returns no properties if not given.
        properties = []
        if sync_all:
            properties, ok = get_all_properties_by_doc_type("deals", api_key)
            if not ok:
                log.error("Failure loading properties for project_id %d on sync_deals", project_id)
                break

        has_more = True
        while has_more:
            parameters = urllib.parse.urlencode(parameter_dict)
            get_url = url + parameters

            # List of all properties to get, returns empty properties if not given.
            if sync_all:
                get_url = get_url + '&' + build_properties_param_str(properties)
                get_url = get_url + '&includeAssociations=true'

            log.warning("Downloading deals for project_id %d from url %s.", project_id, get_url)
            r = requests.get(url= get_url, headers = {})
            if r.status_code == 429: 
                raise Exception("Hubspot API rate limit exceeded for project "+str(project_id))
            if not r.ok:
                log.error("Failure response %d from hubspot on sync_deals", r.status_code)
                break
            response_dict = json.loads(r.text)

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


def get_company_contacts(project_id, api_key, company_id):
    if api_key == "" or not company_id:
        raise Exception("invalid api_key or company_id")
    
    contacts = []
    url = "https://api.hubapi.com/companies/v2/companies/"+str(company_id)+"/contacts?"
    parameter_dict = { 'hapikey': api_key }
    parameters = urllib.parse.urlencode(parameter_dict)
    get_url = url + parameters
    log.warning("Downloading company contacts from url %s.", get_url)
    r = requests.get(url=get_url, headers = {})
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
    for doc in docs:
        company_id = doc.get("companyId")
        contacts = get_company_contacts(project_id, api_key, company_id)
        contactIds = []

        # Adding only contact ids as company contact list
        # properties are type inconsistent
        if contacts != None:
            for contact in contacts:
                vid = contact.get("vid")
                if vid == None: continue
                contactIds.append(vid)
        doc["contactIds"] = contactIds
    return docs

def sync_companies(project_id, api_key, sync_all=False):
    if sync_all:
        urls = [ "https://api.hubapi.com/companies/v2/companies/paged?" ]
        log.warning("Downloading all companies for project_id : "+ str(project_id) + ".")
    else:
        urls = [ "https://api.hubapi.com/companies/v2/companies/recent/modified?" ] # both created and modified. 
        log.warning("Downloading recently created or modified companies for project_id : "+ str(project_id) + ".")

    for url in urls:
        count = 0
        parameter_dict = {'hapikey': api_key, 'limit': PAGE_SIZE}

        properties = []
        if sync_all:
            properties, ok = get_all_properties_by_doc_type("companies", api_key)
            if not ok:
                log.error("Failure loading properties for project_id %d on sync_companies", project_id)
                return

        has_more = True
        while has_more:
            parameters = urllib.parse.urlencode(parameter_dict)
            get_url = url + parameters
            
            if sync_all:
                get_url = get_url + '&' + build_properties_param_str(properties)

            log.warning("Downloading companies for project_id %d from url %s.", project_id, get_url)
            r = requests.get(url= get_url, headers = {})
            if r.status_code == 429:
                raise Exception("Hubspot API rate limit exceeded for project "+str(project_id))
            if not r.ok:
                log.error("Failure response %d from hubspot on sync_companies", r.status_code)
                break
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
            # fills contact ids for each comapany under 'contactIds'.
            fill_contacts_for_companies(project_id, api_key, docs)
            create_all_documents(project_id, 'company', docs)
            count = count + len(docs)
            log.warning("Downloaded and created %d companies. total %d.", len(docs), count)

def sync_forms(project_id, api_key):
    url = "https://api.hubapi.com/forms/v2/forms?"
    parameter_dict = {'hapikey': api_key }
    parameters = urllib.parse.urlencode(parameter_dict)
    get_url = url + parameters

    count = 0
    log.warning("Downloading forms for project_id %d from url %s.", project_id, get_url)
    r = requests.get(url=get_url, headers = {})
    if r.status_code == 429:
        raise Exception("Hubspot API rate limit exceeded for project %d", project_id)
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
        r = requests.get(url=get_url, headers = {})
        if r.status_code == 429:
            raise Exception("Hubspot API rate limit exceeded for project "+str(project_id))
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


def get_sync_info():
    uri = "/data_service/hubspot/documents/sync_info"
    url = options.data_service_host + uri

    response = requests.get(url)
    if not response.ok:
        raise Exception('Failed to get sync info with status: '+str(response.status_code))

    return response.json()


def get_next_sync_info(project_settings, last_sync_info):
    next_sync_info = []
    
    for project_id in project_settings:
        settings = project_settings[project_id]
        api_key = settings.get("api_key")
        if api_key == None:
            log.error("No api_key on project_settings of project %d", project_id)
            continue
        
        sync_info = last_sync_info.get(project_id)
        if sync_info == None:
            log.error("Last sync info missing for project %d", project_id)
            continue

        for doc_type in sync_info:
            next_sync = {}
            next_sync["project_id"] = int(project_id)
            next_sync["api_key"] = api_key
            next_sync["doc_type"] = doc_type
            # sync all, if last sync timestamp is 0.
            next_sync["sync_all"] = sync_info[doc_type] == 0
            next_sync_info.append(next_sync)

    return next_sync_info 


def sync(project_id, api_key, doc_type, sync_all):
    response = {}
    response["project_id"] = project_id
    response["doc_type"] = doc_type
    response["sync_all"] = sync_all

    try:
        if project_id == None or api_key == None or doc_type == None or sync_all == None:
            raise Exception("invalid params on sync, project_id "+str(project_id)+", api_key "+str(api_key)+", doc_type "+str(doc_type)+", sync_all "+str(sync_all))            
        
        if doc_type == "contact":
            sync_contacts(project_id, api_key, sync_all)
        elif doc_type == "company":        
            sync_companies(project_id, api_key, sync_all)
        elif doc_type == "deal":
            sync_deals(project_id, api_key, sync_all)
        elif doc_type == "form":
            sync_forms(project_id, api_key)
        elif doc_type == "form_submission":
            sync_form_submissions(project_id, api_key)
        else:
            raise Exception("invalid doc_type "+ doc_type)

    except Exception as e:
        response["status"] = "failed"
        response["message"] = "Failed with exception: " + str(e)
        return response

    response["status"] = "success"
    return response

if __name__ == "__main__":
    (options, args) = parser.parse_args()
    sync_info = get_sync_info()

    project_settings = sync_info.get("project_settings")
    if project_settings == None:
        log.error("Project settings missing on get sync info response")
        sys.exit(1)
    
    last_sync_info = sync_info.get("last_sync_info")
    if last_sync_info == None:
        log.error("Last sync info missing on get sync info response")
        sys.exit(1)

    next_sync_info = get_next_sync_info(project_settings, last_sync_info)

    next_sync_failures = []
    next_sync_success = []
    for info in next_sync_info:
        response = sync(info.get("project_id"), info.get("api_key"), 
                info.get("doc_type"), info.get("sync_all"))
        if response["status"] == "failed": 
            next_sync_failures.append(response)
        else:
            next_sync_success.append(response)

    status_msg = ""
    if len(next_sync_failures) > 0: status_msg = "Failures on sync."
    else: status_msg = "Successfully synced."
    notification_payload = {
        "status": status_msg, 
        "failures": next_sync_failures, 
        "success": next_sync_success,
    }

    log.warning("Successfully synced. End of hubspot sync job.")
    if len(next_sync_failures) > 0:
        ping_healthcheck(options.env, HEALTHCHECK_PING_ID, notification_payload, endpoint="/fail")
    else:
        ping_healthcheck(options.env, HEALTHCHECK_PING_ID, notification_payload)    
    sys.exit(0)
