from constants.constants import *
from util.util import Util as U
import logging as log
from custom_exception.custom_exception import CustomException
from metrics_aggregator.metrics_aggregator import MetricsAggregator

class LinkedinApiService:
    metrics_aggregator_obj = None
    __instance = None

    @staticmethod
    def get_instance():
        if LinkedinApiService.__instance == None:
            LinkedinApiService()
        return LinkedinApiService.__instance

    def __init__(self) -> None:
        self.metrics_aggregator_obj = MetricsAggregator.get_instance()
        LinkedinApiService.__instance = self

    # We get the company name and other related data here
    # batch_of_ids - ["1,2,3", "4,5,6"] -> batch of ids of length 500
    # each batch is an string with 500 ids joined together with ','
    # Used a string because that was required in API request
    def get_company_data_from_linkedin_with_retries(self, ids_list, access_token):
        map_id_to_org_data = {}
        len_of_batch = ORG_BATCH_SIZE
        request_counter = 0

        batch_of_ids = [",".join(ids_list[i:i + len_of_batch]) for i in range(0, len(ids_list), len_of_batch)]
        
        for ids in batch_of_ids:
            response, req_count = self.org_lookup(access_token, ids)
            request_counter += req_count
            if not response.ok or 'results' not in response.json():
                # TODO: think through if exception should be placed
                err_string = ORG_DATA_FETCH_ERROR.format(response.text)
                raise CustomException(err_string, request_counter, MEMBER_COMPANY_INSIGHTS)
                
            map_id_to_org_data.update(response.json()['results'])

            # retry in case of failed ids
            # sometimes when fetching n number of ids we get result of <n ids, we are retrying for the same missing ids
            failed_ids_for_batch = U.get_failed_ids(ids, map_id_to_org_data)
            if failed_ids_for_batch != "":
                response, req_count = self.org_lookup(access_token, failed_ids_for_batch)
                request_counter += req_count
                if not response.ok or 'results' not in response.json():
                    map_id_to_org_data.update(response.json()['results'])

        self.metrics_aggregator_obj.request_counter += request_counter
        return map_id_to_org_data
    
    def org_lookup(self, access_token, ids):
        url = ORG_LOOKUP_URL.format(ids)
        headers = {'Authorization': 'Bearer ' + access_token, 
                    'X-Restli-Protocol-Version': PROTOCOL_VERSION, 'LinkedIn-Version': LINKEDIN_VERSION}
        return U.request_with_retries_and_sleep(url, headers)
    
    def get_metadata(self, linkedin_setting, url_endpoint, doc_type):
        metadata = []
        request_counter = 0
        response = {}
        project_id, ad_account, access_token = linkedin_setting.project_id, linkedin_setting.ad_account, linkedin_setting.access_token

        url = META_DATA_URL.format(ad_account, url_endpoint, META_COUNT)
        headers = {'Authorization': 'Bearer ' + access_token,
                    'X-Restli-Protocol-Version': PROTOCOL_VERSION, 'LinkedIn-Version': LINKEDIN_VERSION}
        response, req_count = U.request_with_retries_and_sleep(url, headers)
        request_counter += req_count
        if not response.ok:
            errString = API_ERROR_FORMAT.format(
                doc_type, 'metadata', response.status_code,
                response.text, project_id, ad_account)
            raise CustomException(errString, request_counter, doc_type)

        metadata.extend(response.json()[ELEMENTS])
        
        while METADATA in response.json() and NEXT_PAGE_TOKEN in response.json()[METADATA]:
            url = META_DATA_URL_PAGINATED.format(ad_account, url_endpoint, META_COUNT, response.json()[METADATA][NEXT_PAGE_TOKEN])
            response, req_count = U.request_with_retries_and_sleep(url, headers)
            request_counter += req_count
            if not response.ok:
                errString = API_ERROR_FORMAT.format(
                    doc_type, 'metadata', response.status_code,
                    response.text, project_id, ad_account)
                raise CustomException(errString, request_counter, doc_type)

            metadata.extend(response.json()[ELEMENTS])

        self.metrics_aggregator_obj.request_counter += request_counter
        return metadata


    # can't keep very long range, we might hit rate limit   
    # sample api response success -> {'elements': [{}, {}], 'paging': {}}
    def get_insights(self, linkedin_setting, timestamp, doc_type, pivot, 
                                        campaign_group_id=None, end_timestamp=None):
        if doc_type == MEMBER_COMPANY_INSIGHTS:
            log.warning(FETCH_CG_LOG_WITH_DOC_TYPE.format(
                doc_type, campaign_group_id, linkedin_setting.project_id, linkedin_setting.ad_account, timestamp))
        else:
            log.warning(FETCH_LOG_WITH_DOC_TYPE.format(
            doc_type, linkedin_setting.project_id, linkedin_setting.ad_account, timestamp))

        request_counter = 0
        records = 0
        results =[]
        request_rows_start_count = 0
        is_first_fetch = True
        is_pagination_req = False

        # following condition check if it's first pull or pagination is required.
        while is_first_fetch or is_pagination_req:
            
            is_first_fetch = False
            url, headers = U.build_url_and_headers(pivot, doc_type, linkedin_setting, timestamp, 
                                   request_rows_start_count, campaign_group_id, end_timestamp)
            response, req_count = U.request_with_retries_and_sleep(url, headers)
            request_counter += req_count
            
            if not response.ok:
                errString = API_ERROR_FORMAT.format(pivot, 'insights', response.status_code, 
                                response.text, linkedin_setting.project_id, linkedin_setting.ad_account)
                log.error(errString)
                raise CustomException(errString, request_counter, doc_type)
            
            if ELEMENTS in response.json():
                records += len(response.json()[ELEMENTS])
                results.extend(response.json()[ELEMENTS])
            is_pagination_req = len(response.json()[ELEMENTS]) == REQUESTED_ROWS_LIMIT
            request_rows_start_count += REQUESTED_ROWS_LIMIT

        
        if doc_type == MEMBER_COMPANY_INSIGHTS:
            log.warning(NUM_OF_RECORDS_CG_LOG.format(doc_type, campaign_group_id, 
                        linkedin_setting.project_id, linkedin_setting.ad_account, records))
        else:
            log.warning(NUM_OF_RECORDS_LOG.format(
                doc_type, linkedin_setting.project_id, linkedin_setting.ad_account, records))
        self.metrics_aggregator_obj.request_counter += request_counter
        return results

    def get_company_insights_for_campaign(self, linkedin_setting, timestamp, doc_type, pivot, 
                                                campaign_id, end_timestamp=None):
        log.warning(FETCH_C_LOG_WITH_DOC_TYPE.format(
            doc_type, campaign_id, linkedin_setting.project_id, linkedin_setting.ad_account, timestamp))

        request_counter = 0
        records = 0
        results =[]
        request_rows_start_count = 0
        is_first_fetch = True
        is_pagination_req = False

        # following condition check if it's first pull or pagination is required.
        while is_first_fetch or is_pagination_req:
            
            is_first_fetch = False
            url, headers = U.build_url_and_headers_campaign_company(pivot, linkedin_setting, timestamp, 
                                   request_rows_start_count, campaign_id, end_timestamp)
            response, req_count = U.request_with_retries_and_sleep(url, headers)

            request_counter += req_count
            
            if not response.ok:
                errString = API_ERROR_FORMAT.format(pivot, 'insights', response.status_code, 
                                response.text, linkedin_setting.project_id, linkedin_setting.ad_account)
                log.error(errString)
                raise CustomException(errString, request_counter, doc_type)
            
            if ELEMENTS in response.json():
                records += len(response.json()[ELEMENTS])
                results.extend(response.json()[ELEMENTS])
            is_pagination_req = len(response.json()[ELEMENTS]) == REQUESTED_ROWS_LIMIT
            request_rows_start_count += REQUESTED_ROWS_LIMIT

        
        log.warning(NUM_OF_RECORDS_C_LOG.format(doc_type, campaign_id, 
                    linkedin_setting.project_id, linkedin_setting.ad_account, records))
        self.metrics_aggregator_obj.request_counter += request_counter
        return results

    
    
    def get_ad_account_data(self, linkedin_setting):
        url = AD_ACCOUNT_URL.format(linkedin_setting.ad_account)
        headers = {'Authorization': 'Bearer ' + linkedin_setting.access_token,
                    'X-Restli-Protocol-Version': PROTOCOL_VERSION, 'LinkedIn-Version': LINKEDIN_VERSION}
        response, req_count = U.request_with_retries_and_sleep(url, headers)
        if not response.ok:
            errString = API_ERROR_FORMAT.format('ad account', 'metadata',response.status_code, 
                                response.text, linkedin_setting.project_id, linkedin_setting.ad_account)
            log.error(errString)
            raise CustomException(errString, req_count, AD_ACCOUNT)
        metadata = response.json()
        self.metrics_aggregator_obj.request_counter += req_count
        return metadata
    
    def extract_company_insights_for_all_campaign_groups(self, linkedin_setting, start_timestamp, 
                                               end_timestamp, campaign_group_ids_list):
        final_insights = []
        for campaign_group_id in campaign_group_ids_list:
            company_insights = self.extract_company_insights_for_each_campaign_group(
                                                        linkedin_setting, start_timestamp,
                                                        campaign_group_id,
                                                        end_timestamp)

            # adding campaign_group_id because data sent back doesn't contain it
            updated_insights = self.enrich_campaign_group_id_for_member_company_data(
                                        company_insights, campaign_group_id)
            final_insights.extend(updated_insights)
            
        return final_insights
    
    def extract_company_insights_for_each_campaign_group(self, linkedin_setting, start_timestamp,
                                                        campaign_group_id, end_timestamp):
        
        insights_rows = self.get_insights(linkedin_setting, start_timestamp, MEMBER_COMPANY_INSIGHTS, 
                                          'MEMBER_COMPANY', campaign_group_id, end_timestamp)

        return insights_rows

    def enrich_campaign_group_id_for_member_company_data(self, records, campaign_group_id):
        updated_records = []
        dict_to_update = {
            'campaign_group_id': campaign_group_id
        }
        for record in records:
            record.update(dict_to_update)
            updated_records.append(record)
        
        return updated_records
    
    def extract_company_insights_for_all_campaigns(self, linkedin_setting, start_timestamp, 
                                               end_timestamp, campaign_ids_list):
        final_insights = []
        for campaign_id in campaign_ids_list:
            company_insights = self.extract_company_insights_for_each_campaign(
                                                        linkedin_setting, start_timestamp,
                                                        campaign_id,
                                                        end_timestamp)

            # adding campaign_id because data sent back doesn't contain it
            updated_insights = self.enrich_campaign_id_for_member_company_data(
                                        company_insights, campaign_id)
            final_insights.extend(updated_insights)
            
        return final_insights
    
    def extract_company_insights_for_each_campaign(self, linkedin_setting, start_timestamp,
                                                        campaign_id, end_timestamp):
        
        insights_rows = self.get_company_insights_for_campaign(linkedin_setting, start_timestamp, MEMBER_COMPANY_INSIGHTS, 
                                          'MEMBER_COMPANY', campaign_id, end_timestamp)

        return insights_rows

    def enrich_campaign_id_for_member_company_data(self, records, campaign_id):
        updated_records = []
        dict_to_update = {
            'campaign_id': campaign_id
        }
        for record in records:
            record.update(dict_to_update)
            updated_records.append(record)
        
        return updated_records
    
