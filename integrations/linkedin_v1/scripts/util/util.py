import datetime
import time
import json
import logging as log
import datetime
import requests
import logging as log
from datetime import datetime
from constants.constants import *
from custom_exception.custom_exception import CustomException
from _datetime import timedelta

class Util:
    
    @staticmethod    
    def get_separated_date(date):
        date = date.split('-')
        return date[0], date[1], date[2]

    
    @staticmethod    
    def get_split_date_from_timestamp(date):
        new_date = datetime.strptime(str(date), '%Y%m%d').date()
        return new_date.year, int(new_date.month), new_date.day

    
    @staticmethod    
    def get_timestamp(date):
        return int(datetime(date['year'],date['month'],date['day']).strftime('%Y%m%d'))

    
    @staticmethod    
    def ping_healthcheck(env, healthcheck_id, message, endpoint=''):
        message = json.dumps(message, indent=1)
        log.warning('Healthcheck ping for env %s payload %s', env, message)
        if env != 'production': 
            return

        try:
            requests.post('https://hc-ping.com/' + healthcheck_id + endpoint,
                data=message, timeout=10)
        except requests.RequestException as e:
            # Log ping failure here...
            log.error('Ping failed to healthchecks.io: %s' % e)

    
    @staticmethod
    def build_message_and_ping_slack(env, slack_url, token_failures):
        project_ids_list = []
        if env != 'production': 
            return
        
        for failure in token_failures:
            project_ids_list.append(str(failure[PROJECT_ID]))
        
        message = Util.build_slack_block(project_ids_list)
        count = 0
        response = {}
        # retrying
        while count<= 3:
            count += 1
            response = requests.post(slack_url, json=message, timeout=10)
            if response.ok:
                break
        if not response.ok:
            log.error('Ping failed to slack alerts')

    @staticmethod    
    def sort_by_timestamp(data):
        date = data['dateRange']['end']
        return Util.get_timestamp(date)

    # if last sync date was 20230101, then date start would be 20230102 adn date end would be today - 1
    # this is inclusive of date start and date end
    @staticmethod
    def get_timestamp_range_from_start_end_date(date_start, date_end):
        timestamps =[]
        num_of_days = (date_end-date_start).days + 1
        if num_of_days <=0:
            return []
        for i in range (0, num_of_days):
            date_required = date_start.strftime("%Y%m%d")
            date_start = date_start + timedelta(days=1)
            timestamps.append(date_required)
        
        return timestamps

    @staticmethod    
    def get_timestamp_range(linkedin_setting, doc_type, sync_info_with_type, input_start_timestamp=None, input_end_timestamp=None):
        timestamps =[]
        date_start = ''
        date_end = ''

        if input_end_timestamp != None:
            date_end = datetime.strptime(str(input_end_timestamp), '%Y%m%d').date()
        else:
            date_end = (datetime.now() - timedelta(days=1)).date()
        
        if input_start_timestamp != None:
            date_start = datetime.strptime(str(input_start_timestamp), '%Y%m%d').date()
        else:
            if doc_type not in sync_info_with_type:
                date_start = (datetime.now() - timedelta(days=MAX_LOOKBACK)).date()
            else:
                # date start = last sync info date + 1 day
                date_start = (datetime.strptime(str(sync_info_with_type[doc_type]), 
                                            '%Y-%m-%d') + timedelta(days=1)).date()

        timestamps = Util.get_timestamp_range_from_start_end_date(date_start, date_end)
        
        #if range greater than max lookback, return error msg
        if len(timestamps) > MAX_LOOKBACK:
            log.warning(RANGE_EXCEED_LOG.format(
                        linkedin_setting.project_id, linkedin_setting.ad_account, doc_type))
            return timestamps[-MAX_LOOKBACK:]
        return timestamps

    @staticmethod    
    def get_timestamp_range_for_company_insights(linkedin_setting, doc_type, last_timestamp, 
                                                input_start_timestamp=None, input_end_timestamp=None):
        timestamps =[]
        date_start = ''
        date_end = ''

        if input_end_timestamp != None:
            date_end = datetime.strptime(str(input_end_timestamp), '%Y%m%d').date()
        else:
            date_end = (datetime.now() - timedelta(days=1)).date()
        
        if input_start_timestamp != None:
            date_start = datetime.strptime(str(input_start_timestamp), '%Y%m%d').date()
        else:
            if last_timestamp == None or last_timestamp == 0:
                date_start = (datetime.now() - timedelta(days=MAX_LOOKBACK)).date()
            else:
                # date start = last sync info date + 1 day
                date_start = (datetime.strptime(str(last_timestamp), 
                                            '%Y-%m-%d') + timedelta(days=1)).date()

        timestamps = Util.get_timestamp_range_from_start_end_date(date_start, date_end)
        
        #if range greater than max lookback, return error msg
        if len(timestamps) > MAX_LOOKBACK:
            log.warning(RANGE_EXCEED_LOG.format(
                        linkedin_setting.project_id, linkedin_setting.ad_account, doc_type))
            return timestamps[-MAX_LOOKBACK:]
        return timestamps


    # this method is utilised by both t8 and t22 job
    # for t8 buffer is 0 and it returns nearest sunday to today
    # for t22 buffer is 2 weeks and it returns nearest sunday 2 weeks ago
    # in case there has been no backfill, for both jobs it'll pull 2 weeks of data
        # week1 and week2 in case of t8
        # week3 and week4 in case of t22
    # we have added a limitation of looking only upto 2 weeks of data from choosen sunday
    @staticmethod
    def get_timestamp_chunks_to_be_backfilled(weeks_for_buffer, last_timestamp,
                                            input_start_timestamp=None, input_end_timestamp=None):
        timerange_start_date = None
        timerange_end_date = None
        start_date_using_last_timestamp = 0
        sunday_datetime = Util.get_datetime_for_nearest_sunday_before_given_buffer(weeks_for_buffer)
        timerange_end_date = sunday_datetime.date()
        start_date_for_2_weeks = (sunday_datetime - timedelta(days=13)).date()  # by default choosing 2 weeks of data
        # this gets to Monday of 2 weeks ago from the choosen sunday
        # then we can build 2 weeks timerange of mon-sunday, monday-sunday.
        # i.e week1, week2 in case of t8, week3 and week4 in case of t22
        if last_timestamp != None and last_timestamp != 0:
            start_date_using_last_timestamp = (datetime.strptime(str(last_timestamp), '%Y-%m-%d') + timedelta(days=1)).date()
            timerange_start_date = max(start_date_for_2_weeks, start_date_using_last_timestamp)
        else:
            timerange_start_date = start_date_for_2_weeks

        if input_end_timestamp != None:
            timerange_end_date = (datetime.strptime(str(input_end_timestamp), '%Y%m%d')).date()
            if timerange_end_date.isoweekday() != 7:
                raise CustomException("Input end timestamp is not sunday", 0, MEMBER_COMPANY_INSIGHTS)
        if input_start_timestamp != None:
            timerange_start_date = (datetime.strptime(str(input_start_timestamp), '%Y%m%d')).date()
            if timerange_start_date.isoweekday() != 1:
                raise CustomException("Input start timestamp is not monday", 0, MEMBER_COMPANY_INSIGHTS)
        
        required_timerange = []
        num_of_days = (timerange_end_date-timerange_start_date).days + 1
        if num_of_days%7 != 0: # if it's a partial week, make it full week
            num_of_days = num_of_days + (7- (num_of_days%7)) 
            # week = [d3, d4...d7] -> where d3 is day 3 of the week which is wednesday
            # the above condition will convert it to week = [d1, d2......d7]
        if num_of_days <=0:
            return []
        for i in range (0, num_of_days):
            # filling backfill timestamps in reverse
            date_required = timerange_end_date.strftime("%Y%m%d")
            required_timerange.append(date_required)
            timerange_end_date = timerange_end_date - timedelta(days=1)
        
        # reversing backfill timestamps to correct order again
        required_timerange = required_timerange[::-1] 

        return Util.get_n_days_chunks_of_timestamps(required_timerange, 7)

    def get_datetime_for_nearest_sunday_before_given_buffer(weeks_for_buffer):
        current_time = datetime.now()
        current_week_day = current_time.isoweekday() # returns 1 for mon, 7 for sunday
        sunday_near_today = current_time - timedelta(days=(current_week_day%7))
        if current_week_day == 7:
            sunday_near_today = current_time - timedelta(days=7)
        sunday_near_buffer = sunday_near_today - timedelta(days=(weeks_for_buffer*7))
        return sunday_near_buffer
    
    def get_n_days_chunks_of_timestamps(timestamps, n):
        chunked_list = []
        rem = len(timestamps)%n
        if rem != 0:
            chunked_list = [timestamps[0:rem]]
            timestamps = timestamps[rem:]
        chunked_list.extend([timestamps[i * n:(i + 1) * n] for i in range((len(timestamps) + n - 1) // n )])
        return chunked_list


    def build_map_of_campaign_group_info(campaign_group_info):
        campaign_group_info_map = {}
        for campaign_group in campaign_group_info:
            campaign_group_dict_to_add = {
                'campaign_group_name': campaign_group['value']['name'],
                'campaign_group_status': campaign_group['value']['status']
            }
            campaign_group_info_map[campaign_group['id']] = campaign_group_dict_to_add
        return campaign_group_info_map
    
    def merge_2_dictionaries(dict1, dict2):
        final_dict = {}
        if len(dict1) > 0:
            for key, value in dict1.items():
                if key not in final_dict:
                    final_dict[key] = value
        
        if len(dict2) > 0:
            for key, value in dict2.items():
                if key not in final_dict:
                    final_dict[key] = value
        
        return final_dict

    def get_batch_of_ids(records, map_of_id_to_company_data):
        mapIDs = {}
        batch_of_ids = []
        len_of_batch = ORG_BATCH_SIZE
        for data in records:
            id = data['pivotValues'][0].split(':')[3]
            if id not in map_of_id_to_company_data:
                mapIDs[id]= True

        ids_list = list(mapIDs.keys())
        batch_of_ids = [",".join(ids_list[i:i + len_of_batch]) for i in range(0, len(ids_list), len_of_batch)]
        return batch_of_ids
    
    def get_non_present_ids(records, map_of_id_to_company_data):
        mapIDs = {}
        for data in records:
            id = data['pivotValues'][0].split(':')[3]
            # temporary fix start. this specific org id is causing error. Escalating the error to linkedin team
            if str(id) == '1757051':
                continue
            # temporary fix end
            if id not in map_of_id_to_company_data:
                mapIDs[id]= True

        non_present_ids = list(mapIDs.keys())
        return non_present_ids

    def get_failed_ids(ids, map_id_to_org_data):
        ids_list = ids.split(",")
        keys = map_id_to_org_data.keys()
        set_ids = set(ids_list)
        set_keys = set(keys)
        failed_ids = ",".join(set_ids-set_keys)
        return failed_ids

    @staticmethod
    def request_with_retries_and_sleep(url, headers):
        count = 0
        response = {}
        while count<= 3:
            count += 1
            try:
                response = requests.get(url, headers=headers)
                if response.ok:
                    break
                elif "Max retries exceeded" in response.text:
                    time.sleep(300)
                else:
                    time.sleep(30)
            except Exception as e:
                log.warning("Failed with exception %s", str(e))
                time.sleep(30)

        return response, count
    
    def build_url_and_headers(pivot, doc_type, linkedin_setting, start_timestamp, 
                                   request_rows_start_count, campaign_group_id=None, end_timestamp=None):
        
        start_year, start_month, start_day = Util.get_split_date_from_timestamp(start_timestamp)
        end_year, end_month, end_day = Util.get_split_date_from_timestamp(start_timestamp)

        url = INSIGHTS_REQUEST_URL_FORMAT.format(
                    pivot, start_day, start_month, 
                    start_year, end_day, end_month, end_year,
                    REQUESTED_FIELDS, linkedin_setting.ad_account,
                    request_rows_start_count, REQUESTED_ROWS_LIMIT)
        if doc_type == MEMBER_COMPANY_INSIGHTS:
            end_year, end_month, end_day = Util.get_split_date_from_timestamp(end_timestamp)
            url = COMPANY_CAMPAIGN_GROUP_INSIGHTS_REQUEST_URL_FORMAT.format(
                pivot, start_day, start_month, 
                start_year, end_day, end_month, end_year,
                REQUESTED_FIELDS, linkedin_setting.ad_account, campaign_group_id,
                request_rows_start_count, REQUESTED_ROWS_LIMIT)
            
        headers = {'Authorization': 'Bearer ' + linkedin_setting.access_token,
                    'X-Restli-Protocol-Version': PROTOCOL_VERSION, 'LinkedIn-Version': LINKEDIN_VERSION}
        return url, headers
        
    
    @staticmethod    
    def org_lookup(access_token, ids):
        url = ORG_LOOKUP_URL.format(ids)
        headers = {'Authorization': 'Bearer ' + access_token, 
                    'X-Restli-Protocol-Version': PROTOCOL_VERSION, 'LinkedIn-Version': LINKEDIN_VERSION}
        return Util.request_with_retries_and_sleep(url, headers)
    
    def build_slack_block(project_ids):
        message = {}
        blocks = [{
			"type": "header",
			"text": {
				"type": "plain_text",
				"text": "Linkedin token failures"
			}
		}]
        project_ids_str = ", ".join(project_ids)

        fields = [{
                "type": "plain_text",
                "text": project_ids_str
            }]
        section = {
            "type" : "section",
            "fields": fields
        }
        blocks.append(section)
        message["blocks"] = blocks

        return message

    def check_job_type_req(job_type_str):
        daily_job_req, t8_job_req, t22_job_req = False, False, False
        if '1' in job_type_str:
            daily_job_req = True

        if '2' in job_type_str:
            t8_job_req = True

        if '3' in job_type_str:
            t22_job_req = True

        return daily_job_req, t8_job_req, t22_job_req
    
    def get_start_and_end_timestamp_from_chunks(timerange_chunks):
        start_timestamp = timerange_chunks[0][0]
        len_chunks = len(timerange_chunks)
        end_timestamp = timerange_chunks[len_chunks-1][6]
        return start_timestamp, end_timestamp
    
    # the following function primarily validates what the valid timeranges are for weekly pulls
    # currently, it's primarily used to check for t8 job, where we check last week pull shouldn't be done on before thursday
    # we achieve this by checking t-3 not in given weekly range (mon-sun).
    @staticmethod
    def exclude_timerange_inclusive_of_day3(timeranges):
        new_timeranges = []
        date_3_days_ago = (datetime.now() - timedelta(days=3)).date().strftime("%Y%m%d")
        for timerange in timeranges:
            if date_3_days_ago not in timerange:
                new_timeranges.append(timerange)
        
        return new_timeranges