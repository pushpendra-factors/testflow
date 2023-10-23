import logging as log
import requests
from transformations import DataTransformation
from data_service import DataService
from data_insert import DataInsert
from util import Util as U
from datetime import datetime
from constants import *
from collections import OrderedDict
from linkedin_api_helper import LinkedinApiHelper

class DataFetch:

    def get_metadata(ad_account, access_token, url_endpoint, doc_type, project_id):
        metadata = []
        request_counter = 0
        is_first_fetch = True
        response = {}

        start = 0
        while is_first_fetch or len(response.json()[ELEMENTS])>=META_COUNT:
            is_first_fetch = False
            url = META_DATA_URL.format(ad_account, url_endpoint, start, META_COUNT)
            headers = {'Authorization': 'Bearer ' + access_token,
                    'X-Restli-Protocol-Version': PROTOCOL_VERSION, 'LinkedIn-Version': LINKEDIN_VERSION}
            response, req_count = U.request_with_retries_and_sleep(url, headers)
            request_counter += req_count
            if not response.ok:
                errString = API_ERROR_FORMAT.format(
                    doc_type, 'metadata', response.status_code,
                    response.text, project_id, ad_account)
                return metadata, errString, request_counter
            metadata.extend(response.json()[ELEMENTS])
            start +=META_COUNT
        return metadata, '', request_counter


    # can't keep very long range, we might hit rate limit   
    # sample api response success -> {'elements': [{}, {}], 'paging': {}}
    def get_insights(linkedin_setting, timestamp, doc_type, pivot, meta_request_count, 
                                        campaign_group_id=None, end_timestamp=None):
        if doc_type == MEMBER_COMPANY_INSIGHTS:
            log.warning(FETCH_CG_LOG_WITH_DOC_TYPE.format(
                doc_type, campaign_group_id, linkedin_setting.project_id, timestamp))
        else:
            log.warning(FETCH_LOG_WITH_DOC_TYPE.format(
            doc_type, linkedin_setting.project_id, timestamp))

        request_counter = meta_request_count
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
                return [], {'status': 'failed', 'errMsg': errString, API_REQUESTS: request_counter}
            
            if ELEMENTS in response.json():
                records += len(response.json()[ELEMENTS])
                results.extend(response.json()[ELEMENTS])
            is_pagination_req = len(response.json()[ELEMENTS]) == REQUESTED_ROWS_LIMIT
            request_rows_start_count += REQUESTED_ROWS_LIMIT

        
        if doc_type == MEMBER_COMPANY_INSIGHTS:
            log.warning(NUM_OF_RECORDS_CG_LOG.format(
                doc_type, campaign_group_id, linkedin_setting.project_id, records))
        else:
            log.warning(NUM_OF_RECORDS_LOG.format(
                doc_type, linkedin_setting.project_id, records))
        
        return results, {'status': 'success', 'errMsg': '', API_REQUESTS: request_counter}



    def get_ad_account_data(options, linkedin_setting, input_end_timestamp):
        url = AD_ACCOUNT_URL.format(linkedin_setting.ad_account)
        headers = {'Authorization': 'Bearer ' + linkedin_setting.access_token,
                    'X-Restli-Protocol-Version': PROTOCOL_VERSION, 'LinkedIn-Version': LINKEDIN_VERSION}
        response, req_count = U.request_with_retries_and_sleep(url, headers)
        if not response.ok:
            errString = API_ERROR_FORMAT.format(
                            'ad account', 'metadata',
                            response.status_code, response.text,
                            linkedin_setting.project_id, 
                            linkedin_setting.ad_account)
            return {'status': 'failed', 'errMsg': errString, API_REQUESTS: req_count}
        metadata = response.json()
        timestamp = int(datetime.now().strftime('%Y%m%d'))
        if input_end_timestamp != None:
            timestamp = input_end_timestamp
        
        response = (DataService(options)
                    .add_linkedin_documents(
                        linkedin_setting.project_id, linkedin_setting.ad_account,
                        AD_ACCOUNT, str(metadata['id']),
                        metadata, timestamp))

        if not response.ok and response.status_code != 409:
            return {'status': 'failed',
                'errMsg': 'Failed inserting add accounts data', API_REQUESTS: 1}
        return {'status': 'success', 'errMsg': '', API_REQUESTS: 1}
    
    # flow->
    # get today's metadata
    # update heirarchical data
        # update campaign_group_meta, campaign_meta, creative_meta with heirarchichal ids and names
    # insert metadata based on last_sync_info -> current day's metadata is used as a backfill
    # get time range for which report data is to be inserted based on start timestamp or last sync info
    # get insights for given timerange and insert into db
    @classmethod
    def etl_ads_hierarchical_data(self, options,
            linkedin_setting, sync_info_with_type,
            campaign_group_meta, campaign_meta, creative_meta, 
            meta_doc_type, insights_doc_type, meta_url_endpoint,
            pivot_insights, input_end_timestamp):
        log.warning(META_FETCH_START.format(meta_doc_type, linkedin_setting.project_id))
        
        metadata, errString, request_counter = self.get_metadata(
                                                linkedin_setting.ad_account,
                                                linkedin_setting.access_token,
                                                meta_url_endpoint, meta_doc_type,
                                                linkedin_setting.project_id)
        if errString != '':
            return {'status': 'failed', 'errMsg': errString, 
                    API_REQUESTS: request_counter}
        
        updated_meta = DataTransformation.update_hierarchical_data(
                                                metadata, meta_doc_type,
                                                campaign_group_meta, campaign_meta,
                                                creative_meta)

        meta_insertion_response = self.get_timerange_and_insert_metadata(
                                                options, linkedin_setting, 
                                                metadata, updated_meta,
                                                meta_doc_type, sync_info_with_type,
                                                input_end_timestamp, request_counter)
        if meta_insertion_response['errMsg'] != '':
            return meta_insertion_response
            

        timestamp_range_for_insights, errMsg = U.get_timestamp_range(
                                                insights_doc_type,
                                                sync_info_with_type, input_end_timestamp)
        if errMsg != '':
            log.warning("Range exceeded for project_id {} for doc_type {}".format(
                        linkedin_setting.project_id, insights_doc_type))
        
        return self.get_insights_for_timerange_and_insert(
                                                options, linkedin_setting,
                                                insights_doc_type, pivot_insights, 
                                                campaign_group_meta, campaign_meta, 
                                                creative_meta, timestamp_range_for_insights,
                                                meta_insertion_response[API_REQUESTS])


    @classmethod
    def get_timerange_and_insert_metadata(self, options, linkedin_setting, metadata,
                                                updated_meta, meta_doc_type, 
                                                sync_info_with_type,
                                                input_end_timestamp, request_counter):
        log.warning(NUM_OF_RECORDS_LOG.format(meta_doc_type,
                         linkedin_setting.project_id, len(metadata)))
        
        timestamp_range_for_meta, errMsg = U.get_timestamp_range(
                                                meta_doc_type,
                                                sync_info_with_type, input_end_timestamp)
        if errMsg != '':
            log.warning("Range exceeded for project_id {} for doc_type {}".format(
                        linkedin_setting.project_id, meta_doc_type))
        for timestamp in timestamp_range_for_meta:
            
            insert_response = DataInsert.insert_metadata(
                                                options, meta_doc_type,
                                                linkedin_setting.project_id, 
                                                linkedin_setting.ad_account,
                                                metadata, timestamp, updated_meta)
            if not insert_response.ok and insert_response.status != 409:
                errString = DOC_INSERT_ERROR.format(
                                meta_doc_type, "metadata",
                                insert_response.status, insert_response.text,
                                linkedin_setting.project_id, linkedin_setting.ad_account, 
                                timestamp)
                log.error(errString)
                return  {'status': 'failed', 'errMsg': errString, 
                            API_REQUESTS: request_counter}
        log.warning(FINAL_INSERTION_END_LOG.format(meta_doc_type, 
                            'metadata', linkedin_setting.project_id))

        return {'status': 'success', 'errMsg': '', API_REQUESTS: request_counter}

    @classmethod
    def get_insights_for_timerange_and_insert(self, options, 
                                                linkedin_setting, insights_doc_type,
                                                pivot_insights, campaign_group_meta, 
                                                campaign_meta, creative_meta, 
                                                timestamp_range, request_counter):
        for timestamp in timestamp_range:
            results, resp = self.get_insights(linkedin_setting, timestamp,
                                 insights_doc_type, pivot_insights, request_counter)
            if resp['status'] == 'failed' or resp['errMsg'] != '':
                return resp
            request_counter = resp[API_REQUESTS]
            results = DataTransformation.update_result_with_metadata(
                                                results, insights_doc_type, 
                                                campaign_group_meta, campaign_meta,
                                                creative_meta)
                
            errString = DataInsert.insert_insights(
                                                options, insights_doc_type, 
                                                linkedin_setting.project_id, 
                                                linkedin_setting.ad_account, 
                                                results, timestamp)
            if errString != '':
                return {'status': 'failed', 'errMsg': errString,
                            API_REQUESTS: request_counter}
        return {'status': 'success', 'errMsg': '', API_REQUESTS: request_counter}
 
    # get timerange based on last sync info eg. (t-3 to t-1)
    # get campaign group for t-3 to t-1
    # for each timestamp in t-3 to t-1
        # for each campaign group
            # get insights 
            # add campaign group info
        # add org details and insert
    @classmethod
    def etl_member_company_data_with_campaign_group(self, options, linkedin_setting,
                                                sync_info_with_type, 
                                                input_end_timestamp):
        campaign_group_info = []
        request_counter = 0
        map_of_id_to_company_data = {}

        timestamp_range, len_timerange = U.get_timestamp_range_and_length_for_company_data(
                                                linkedin_setting.project_id, 
                                                sync_info_with_type, input_end_timestamp)
        if len_timerange == 0:
            return {'status': 'success', 'errMsg': '', API_REQUESTS: request_counter}
        
        campaign_group_info, err_msg = DataService(options).get_campaign_group_data_for_given_timerange(
                                            linkedin_setting.project_id, linkedin_setting.ad_account, 
                                            timestamp_range[0], timestamp_range[len_timerange-1])
        if err_msg != '':
            return {'status': 'failed', 'errMsg': err_msg, API_REQUESTS: 0}

        for timestamp in timestamp_range:

            final_insights, resp = self.extract_and_enrich_company_data_for_all_campaigns(
                                                    linkedin_setting, timestamp, 
                                                    map_of_id_to_company_data, 
                                                    campaign_group_info, timestamp)
            if resp['status'] == 'failed' or resp['errMsg'] != '':
                return resp
            
            insert_err = DataInsert.insert_insights(options, MEMBER_COMPANY_INSIGHTS, 
                                                linkedin_setting.project_id, linkedin_setting.ad_account, 
                                                final_insights, timestamp, False)
            if insert_err != '':
                return {'status': 'failed', 'errMsg': insert_err, 
                            API_REQUESTS: request_counter}
        
        return {'status': 'success', 'errMsg': '', API_REQUESTS: request_counter}
    
    @classmethod
    def extract_and_enrich_company_data_for_all_campaigns(self, linkedin_setting, 
                                                        timestamp, map_of_id_to_company_data,
                                                        campaign_group_info, end_timestamp):
        final_insights = []
        request_counter = 0
        for campaign_group in campaign_group_info:
            enriched_insights, resp = self.extract_and_enrich_company_data_for_each_campaign(
                                                        linkedin_setting, timestamp,
                                                        campaign_group, map_of_id_to_company_data,
                                                        end_timestamp)
            if resp['status'] == 'failed' or resp['errMsg'] != '':
                return [], resp
            request_counter = resp[API_REQUESTS]
            final_insights.extend(enriched_insights)

        return final_insights, {'status': 'success', 'errMsg': '', API_REQUESTS: request_counter}
    
    @classmethod
    def extract_and_enrich_company_data_for_each_campaign(self, linkedin_setting, timestamp,
                                                        campaign_group, map_of_id_to_company_data,
                                                        end_timestamp):
        insights_rows, resp = self.get_insights(linkedin_setting, timestamp,
                                                        MEMBER_COMPANY_INSIGHTS, 
                                                        'MEMBER_COMPANY', 0, 
                                                        campaign_group['id'], end_timestamp)
        if resp['status'] == 'failed' or resp['errMsg'] != '':
            return [], resp
        request_counter = resp[API_REQUESTS]

        map_of_id_to_company_data, resp = LinkedinApiHelper.fetch_and_update_org_data_to_map(
                                            linkedin_setting.access_token, insights_rows, 
                                            map_of_id_to_company_data)
        if resp['status'] == 'failed' or resp['errMsg'] != '':
            resp[API_REQUESTS] += request_counter
            return [], resp
        request_counter += resp[API_REQUESTS]

        enriched_insights = DataTransformation.enrich_campaign_company_fields_for_member_company_data(
                                map_of_id_to_company_data, insights_rows, campaign_group)
        return enriched_insights, {'status': 'success', 'errMsg': '', API_REQUESTS: request_counter}
    
    @classmethod
    def etl_member_company_data_old(self, options, linkedin_setting,
                                                sync_info_with_type, 
                                                end_timestamp, backfill_project_ids):
        backfill_project_ids_list = backfill_project_ids.split(",")
        is_backfill_enable_for_project = (linkedin_setting.project_id in backfill_project_ids_list
                                            or backfill_project_ids == '*')
        # for avoiding key error
        if 'last_backfill_timestamp' not in sync_info_with_type:
            sync_info_with_type['last_backfill_timestamp'] = 0
        timerange_for_insights, timerange_for_backfill, errMsg = (U
                                                .get_timestamp_ranges_for_company_insights_old(
                                                MEMBER_COMPANY_INSIGHTS, 
                                                sync_info_with_type, end_timestamp,
                                                is_backfill_enable_for_project))
        if errMsg != '':
            log.warning("Range exceeded for project_id {} for doc_type {}".format(
                        linkedin_setting.project_id, MEMBER_COMPANY_INSIGHTS))

        request_counter = 0
        for timestamp in timerange_for_insights:
            
            resp = self.etl_company_insights_for_timestamp_old(options,
                            linkedin_setting, request_counter, timestamp)
            if resp['status'] == 'failed' or resp['errMsg'] != '':
                return resp
            request_counter = resp[API_REQUESTS]
        
        
          
        if len(timerange_for_backfill) > 0:
            resp = self.delete_and_backfill_member_company_insights_old(
                                                options, linkedin_setting, 
                                                timerange_for_backfill, request_counter)
            if resp['errMsg'] != '':
                return resp
            request_counter = resp[API_REQUESTS]
        
        return {'status': 'success', 'errMsg': '', API_REQUESTS: request_counter}
    
    @classmethod
    def etl_company_insights_for_timestamp_old(self, options,
                                                linkedin_setting, request_counter, 
                                                timestamp, is_backfill=False):
        results, resp = self.get_insights_old(linkedin_setting, timestamp,
                                                MEMBER_COMPANY_INSIGHTS, 
                                                'MEMBER_COMPANY', request_counter)
        if resp['status'] == 'failed' or resp['errMsg'] != '':
            return resp
        
        return self.enrich_company_details_and_insert_data_old(options, MEMBER_COMPANY_INSIGHTS,
                                                linkedin_setting.project_id, 
                                                linkedin_setting.ad_account,
                                                linkedin_setting.access_token, 
                                                results, resp[API_REQUESTS], 
                                                timestamp, is_backfill)
    
    @classmethod
    def enrich_company_details_and_insert_data_old(self, options, 
                                                doc_type, project_id, 
                                                ad_account, access_token, 
                                                records, request_counter, 
                                                timestamp, is_backfill=False):
        updated_records = []
        if len(records) != 0:
            ids_batch = U.get_batch_of_ids_old(records)
            map_id_to_org_data, request_counter, errString = (self
                                            .get_org_data_from_linkedin_with_retries_old(
                                                ids_batch, access_token, 
                                                request_counter))
            if errString != '':
                log.error(errString)
                return {'status': 'failed', 'errMsg': errString, 
                        API_REQUESTS: request_counter}
            
            updated_records = DataTransformation.update_org_data_old(
                                                map_id_to_org_data, 
                                                records)
        else:
            log.warning(NO_DATA_MEMBER_COMPANY_LOG.format(
                                                project_id, 
                                                ad_account))
        
        insert_err = DataInsert.insert_insights(options, 
                                                doc_type, project_id,
                                                ad_account, updated_records, 
                                                timestamp, is_backfill)
        if insert_err != '':
            return {'status': 'failed', 'errMsg': insert_err, 
                        API_REQUESTS: request_counter}
        return {'status': 'success', 'errMsg': '', API_REQUESTS: request_counter}
    

    @classmethod
    def delete_and_backfill_member_company_insights_old(
                                                self, options,
                                                linkedin_setting, 
                                                backfill_timestamps, 
                                                request_counter):
        response = {}
        for backfill_timestamp in backfill_timestamps:
            delete_response = (DataService(options)
                                .delete_linkedin_documents_for_doc_type_and_timestamp(
                                                    linkedin_setting.project_id,
                                                    linkedin_setting.ad_account,
                                                    MEMBER_COMPANY_INSIGHTS,
                                                    backfill_timestamp))
            if not delete_response.ok:
                return {'status': 'failed', 'errMsg': delete_response.text, 
                            API_REQUESTS: request_counter}
            
            response = self.etl_company_insights_for_timestamp_old(
                                                    options, linkedin_setting, 
                                                    request_counter, backfill_timestamp,
                                                    True)
            if response['errMsg'] != '':
                log.warning(response['errMsg'])
                return response
            request_counter = response[API_REQUESTS]
        
        return {'status': 'success', 'errMsg': '', API_REQUESTS: request_counter}
        
            # can't keep very long range, we might hit rate limit   
    def get_insights_old(linkedin_setting, timestamp, doc_type, pivot, meta_request_count):
        log.warning(FETCH_LOG_WITH_DOC_TYPE.format(
            doc_type, linkedin_setting.project_id, timestamp))
        
        start_year, start_month, start_day = U.get_split_date_from_timestamp(timestamp)
        end_year, end_month, end_day = U.get_split_date_from_timestamp(timestamp)

        request_counter = meta_request_count
        records = 0
        results =[]

        start = 0
        is_first_fetch = True
        # following condition check if it's first pull or pagination is required.
        while is_first_fetch or len(response.json()[ELEMENTS])>=INSIGHTS_COUNT:
            is_first_fetch = False
            url = INSIGHTS_REQUEST_URL_FORMAT.format(
                    pivot, start_day, start_month, 
                    start_year, end_day, end_month, end_year,
                    REQUESTED_FIELDS, linkedin_setting.ad_account,
                    start, INSIGHTS_COUNT)
            
            headers = {'Authorization': 'Bearer ' + linkedin_setting.access_token,
                    'X-Restli-Protocol-Version': PROTOCOL_VERSION, 'LinkedIn-Version': LINKEDIN_VERSION}
            response, req_count = U.request_with_retries_and_sleep(url, headers)
            request_counter += req_count
            if not response.ok:
                errString = API_ERROR_FORMAT.format(
                                pivot, 'insights',
                                response.status_code, response.text, 
                                linkedin_setting.project_id, linkedin_setting.ad_account)
                log.error(errString)
                return [], {'status': 'failed', 'errMsg': errString,
                                API_REQUESTS: request_counter}
            if ELEMENTS in response.json():
                records += len(response.json()[ELEMENTS])
                results.extend(response.json()[ELEMENTS])
            start += INSIGHTS_COUNT

        log.warning(NUM_OF_RECORDS_LOG.format(
            doc_type, linkedin_setting.project_id, records))
        return results, {'status': 'success', 'errMsg': '',
                            API_REQUESTS: request_counter}

      # We get the company name and other related data here
    # batch_of_ids - ["1,2,3", "4,5,6"] -> batch of ids of length 500
    # each batch is an string with 500 ids joined together with ','
    # Used a string because that was required in API request
    def get_org_data_from_linkedin_with_retries_old(batch_of_ids, access_token, request_counter):
        map_id_to_org_data = {}

        for ids in batch_of_ids:
            response, req_count = U.org_lookup(access_token, ids)
            request_counter += req_count
            if not response.ok or 'results' not in response.json():
                return ({}, request_counter, ORG_DATA_FETCH_ERROR.format(
                            response.text))
            map_id_to_org_data.update(response.json()['results'])

            # retry in case of failed ids
            failed_ids_for_batch = U.get_failed_ids(ids, map_id_to_org_data)
            if failed_ids_for_batch != "":
                response, req_count = U.org_lookup(access_token, failed_ids_for_batch)
                request_counter += req_count
                if 'results' in response.json() and len(response.json()['results']) > 0:
                    map_id_to_org_data.update(response.json()['results'])

        return map_id_to_org_data, request_counter, ""
    