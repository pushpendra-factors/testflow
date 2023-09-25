from constants import *
from util import Util as U
import logging as log

class LinkedinApiHelper:

    # We get the company name and other related data here
    # batch_of_ids - ["1,2,3", "4,5,6"] -> batch of ids of length 500
    # each batch is an string with 500 ids joined together with ','
    # Used a string because that was required in API request
    def get_company_data_from_linkedin_with_retries(ids_list, access_token):
        map_id_to_org_data = {}
        len_of_batch = ORG_BATCH_SIZE
        request_counter = 0

        batch_of_ids = [",".join(ids_list[i:i + len_of_batch]) for i in range(0, len(ids_list), len_of_batch)]
        
        for ids in batch_of_ids:
            response, req_count = LinkedinApiHelper.org_lookup(access_token, ids)
            request_counter += req_count
            if not response.ok or 'results' not in response.json():
                return ({}, request_counter, ORG_DATA_FETCH_ERROR.format(
                            response.text))
            map_id_to_org_data.update(response.json()['results'])

            # retry in case of failed ids
            failed_ids_for_batch = U.get_failed_ids(ids, map_id_to_org_data)
            if failed_ids_for_batch != "":
                response, req_count = LinkedinApiHelper.org_lookup(access_token, failed_ids_for_batch)
                request_counter += req_count
                if 'results' in response.json() and len(response.json()['results']) > 0:
                    map_id_to_org_data.update(response.json()['results'])

        return map_id_to_org_data, request_counter, ""
    
    def org_lookup(access_token, ids):
        url = ORG_LOOKUP_URL.format(ids)
        headers = {'Authorization': 'Bearer ' + access_token, 
                    'X-Restli-Protocol-Version': PROTOCOL_VERSION, 'LinkedIn-Version': LINKEDIN_VERSION}
        return U.request_with_retries_and_sleep(url, headers)
    
    def fetch_and_update_org_data_to_map(access_token, records, map_of_id_to_company_data):
        
        if len(records) == 0:
            return map_of_id_to_company_data, {'status': 'success', 'errMsg': '', 
                                        API_REQUESTS: 0}
        
        non_present_ids = U.get_non_present_ids(records, map_of_id_to_company_data)
        map_of_new_company_data, request_counter, errString = (LinkedinApiHelper
                                                            .get_company_data_from_linkedin_with_retries(
                                                                non_present_ids, access_token))
        if errString != '':
            log.error(errString)
            return {}, {'status': 'failed', 'errMsg': errString, 
                    API_REQUESTS: request_counter}
        
        map_of_id_to_company_data.update(map_of_new_company_data)
        
        return map_of_id_to_company_data, {'status': 'success', 'errMsg': '', 
                                        API_REQUESTS: request_counter}