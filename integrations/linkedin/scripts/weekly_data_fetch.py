import logging as log
import requests
from transformations import DataTransformation
from data_service import DataService
from data_insert import DataInsert
from util import Util as U
from datetime import datetime
from constants import *
from collections import OrderedDict
from data_fetch import DataFetch


class WeeklyDataFetch:
        # timerange in chunks of 7 days = get timestamp chunks
    # the above checks gets range from t-35 to (t-15 or last_backfill timestamp {whichever is greater})
    # then it divdes the timerange into 7 days chunks because we want to do distribution of data only for 7 days at a time
    # for each chunk (eg t-15 to t-9)
        # get campaign group info from db
        # for each campaign group
            # get insights for t-15 to t-8 combined and add campaign group details
        # divide each row into 7 rows
        # for each timestamp in t-15 to t-8
            # delete the existing data for that day
            # add org details to newly fetched data and insert
    @classmethod
    def weekly_job_etl_and_backfill_company_data_with_campaign_group(self, options, linkedin_setting,
                                                sync_info_with_type):
        
        last_backfill_timestamp = sync_info_with_type['last_backfill_timestamp']
        # last_backfill_timestamp is minProjectIngestionTimestamp for new integrations
        # or it's max(backfilled_timestamp) + 1

        if last_backfill_timestamp == None or last_backfill_timestamp == 0:
            return {'status': 'failed', 'errMsg': "Backfill timestamp is unavailable", 
                            API_REQUESTS: 0}
        
        # get 7 days chunks of timerange from last_backfill_timestamp to t-15
        timestamp_range_chunks = U.get_timestamp_chunks_to_be_backfilled(last_backfill_timestamp)
        campaign_group_info = []
        request_counter = 0
        records = []
        map_of_id_to_company_data = {}

        for timestamp_range_to_be_backfilled in timestamp_range_chunks:
            len_timerange = len(timestamp_range_to_be_backfilled)
            if len_timerange == 0:
                continue
            campaign_group_info, errMsg = DataService(options).get_campaign_group_data_for_given_timerange(
                    linkedin_setting.project_id, linkedin_setting.ad_account, 
                    timestamp_range_to_be_backfilled[0], timestamp_range_to_be_backfilled[len_timerange-1])
            
            if errMsg != '':
                return {'status': 'failed', 'errMsg': errMsg, 
                        API_REQUESTS: 0}
            
            distributed_records_map_with_timestamp = OrderedDict()
            start_timestamp, end_timestamp = timestamp_range_to_be_backfilled[0], timestamp_range_to_be_backfilled[len_timerange-1]
            # in this different start and end timestamps are passed, 
            # but in case of daily fetch part same timestamp is passed as start and end timestamp
            enriched_insights, resp = DataFetch.extract_and_enrich_company_data_for_all_campaigns(
                                                    linkedin_setting, start_timestamp, 
                                                    map_of_id_to_company_data, 
                                                    campaign_group_info, end_timestamp)
            if resp['status'] == 'failed' or resp['errMsg'] != '':
                return resp
            
            # split each row evenly in 7 or len_timerange parts and maps it to the timestamp
            distributed_records_map_with_timestamp = DataTransformation.distribute_data_across_timerange(
                                                        enriched_insights, timestamp_range_to_be_backfilled)
            
            # here we are directly deleting and inserting the data. Existing row check is present in event creation job
            for timestamp, records in distributed_records_map_with_timestamp.items():
                delete_response = (DataService(options)
                                    .delete_linkedin_documents_for_doc_type_and_timestamp(
                                                        linkedin_setting.project_id,
                                                        linkedin_setting.ad_account,
                                                        MEMBER_COMPANY_INSIGHTS,
                                                        timestamp))
                if not delete_response.ok:
                    return {'status': 'failed', 'errMsg': delete_response.text, 
                                API_REQUESTS: request_counter}
                insert_err = DataInsert.insert_insights(options, 
                                                MEMBER_COMPANY_INSIGHTS, linkedin_setting.project_id,
                                                linkedin_setting.ad_account, records, 
                                                timestamp, True)
                if insert_err != '':
                    return {'status': 'failed', 'errMsg': insert_err, 
                            API_REQUESTS: request_counter}
        
        return {'status': 'success', 'errMsg': '', API_REQUESTS: request_counter}
   