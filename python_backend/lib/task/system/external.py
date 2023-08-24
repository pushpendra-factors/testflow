import logging as log
import uuid
from datetime import datetime

import requests
import time

from .base import BaseSystem
from ...utils.json import JsonUtil


# paginated strategy with support for systems with http and url only.
# May be a separate class requirement is required so that client and system implementation are decoupled.
# This is specific implementation for facebook.
# response of class methods are different.
# eg - local_storage vs external - though we have same base class being used..
class ExternalSystem(BaseSystem):
    BASE_URL_POLL_ASYNC = "https://graph.facebook.com/v17.0/{}?access_token={}"
    BASE_URL_INSIGHTS_ASYNC = "https://graph.facebook.com/v17.0/{}/insights?access_token={}"

    def read(self):
        total_requests = 0
        total_async_requests = 0
        result_records = []
        records, next_page_metadata, result_response, is_async = self.get_paginated_from_source(self.system_attributes["url"])
        total_requests += 1
        if is_async:
            total_async_requests += 1
        if not result_response.ok:
            return "", result_response, total_requests, total_async_requests
        result_records.extend(records)

        while next_page_metadata["exists"]:
            next_page_link = next_page_metadata["link"]
            records, next_page_metadata, result_response, is_async = self.get_paginated_from_source(next_page_link)
            total_requests += 1
            if is_async:
                total_async_requests += 1
            result_records.extend(records)
        return JsonUtil.create(records), result_response, total_requests, total_async_requests

    def write(self, input_string):
        input_records = JsonUtil.read(input_string)
        if len(input_records) == 0:
            curr_payload = self.get_payload_for_facebook_data_service({"id": str(uuid.uuid4())})
            curr_response = requests.post(self.system_attributes["url"], json=curr_payload)
            if not curr_response.ok:
                return
                # log.error('Failed to add response %s to facebook data service for project %s. StatusCode:  %d, %s',
                #           self.system_attributes["type_alias"], self.system_attributes["project_id"],
                #           curr_response.status_code, curr_response.json())
        for input_record in input_records:
            curr_payload = self.get_payload_for_facebook_data_service(input_record)
            curr_response = requests.post(self.system_attributes["url"], json=curr_payload)
            if not curr_response.ok:
                return
                # log.error('Failed to add response %s to facebook data service for project %s. StatusCode:  %d, %s',
                #           self.system_attributes["type_alias"], self.system_attributes["project_id"],
                #           curr_response.status_code, curr_response.json())
        return

    def get_paginated_from_source(self, url):
        result_response, is_async = self.handle_request_with_retries(url)
        if not result_response.ok:
            return [], {"exists": False, "link": ""}, result_response, is_async
        if "paging" in result_response.json() and "next" in result_response.json()["paging"]:
            next_page_metadata = {"exists": True, "link": result_response.json()["paging"]["next"]}
        else:
            next_page_metadata = {"exists": False, "link": ""}
        return result_response.json()['data'], next_page_metadata, result_response, is_async

    def get_payload_for_facebook_data_service(self, value):
        payload = {
            'project_id': int(self.system_attributes["project_id"]),
            'customer_ad_account_id': self.system_attributes["customer_account_id"],
            'type_alias': self.system_attributes["type_alias"],
            'id': value["id"],
            'value': value,
            'timestamp': self.system_attributes["timestamp"],
            'platform': 'facebook'
        }
        return payload

    def handle_request_with_retries(self, url):
        r = requests.get(url)
        if r.ok or 'OAuthException' in r.text:
            return r, False
        r = self.try_async_request(url)
        return r, True

    def try_async_request(self, url):
        r = requests.post(url=url)
        if not r.ok:
            return r
        rep_id = r.json()["report_run_id"]
        poll = True
        sleep_time = 5
        while poll and sleep_time < 65:
            time.sleep(sleep_time)
            sleep_time += 5
            r = requests.get(self.BASE_URL_POLL_ASYNC.format(rep_id, self.system_attributes["access_token"]), timeout=300)
            if not r.ok:
                return r
            if r.json()["async_status"] == "Job Completed" and r.json()["async_percent_completion"] == 100:
                poll = False
        
        r = requests.get(self.BASE_URL_INSIGHTS_ASYNC.format(rep_id, self.system_attributes["access_token"]), timeout=300)
        return r

