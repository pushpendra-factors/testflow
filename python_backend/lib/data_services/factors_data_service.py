# TODO: Check the error handling.
from datetime import datetime, timedelta
from typing import Dict, List

import requests
import logging as log
import time

from requests import Response
from lib.utils.sync_util import SyncUtil

# Note: This class currently holds 2 functionalities - 1. Fetching data 2. Provide data with proper transformation(sometimes).
# TODO Add Ability to test.
class FactorsDataService:
    data_service_path = None
    BATCH_SIZE = 1000

    @classmethod
    def init(cls, data_service_path):
        cls.data_service_path = data_service_path

    @classmethod
    def add_refresh_token(cls, session, payload):
        if session is None or session == "":
            log.error("Invalid session cookie on add_refresh_token request.")
            return

        url = cls.data_service_path + "/adwords/add_refresh_token"

        # response = requests.post(url, json=payload)
        response, errMsg = SyncUtil.post_request_with_retries(url, payload)
        if response is None or errMsg != '':
            log.error(errMsg)
            return

        return response
    
    @classmethod
    def add_gsc_refresh_token(cls, session, payload):
        if session is None or session == "":
            log.error("Invalid session cookie on add_refresh_token request.")
            return

        url = cls.data_service_path + "/google_organic/add_refresh_token"

        # response = requests.post(url, json=payload)
        response, errMsg = SyncUtil.post_request_with_retries(url, payload)
        if response is None or errMsg != '':
            log.error(errMsg)
            return

        return response

    @classmethod
    def get_adwords_refresh_token(cls, project_id):
        url = cls.data_service_path + "/adwords/get_refresh_token"
        # project_id as str for consistency on json.
        payload = {"project_id": str(project_id)}
        # response = requests.post(url, json=payload)
        response, errMsg = SyncUtil.post_request_with_retries(url, payload)
        if response is None or errMsg != '':
            log.error(errMsg)
            return

        return response

    @classmethod
    def get_gsc_refresh_token(cls, project_id):
        url = cls.data_service_path + "/google_organic/get_refresh_token"
        # project_id as str for consistency on json.
        payload = {"project_id": str(project_id)}
        # response = requests.post(url, json=payload)
        response, errMsg = SyncUtil.post_request_with_retries(url, payload)
        if response is None or errMsg != '':
            log.error(errMsg)
            return

        return response

    @classmethod
    def get_last_sync_infos_for_all_projects(cls):
        url = cls.data_service_path + "/adwords/documents/last_sync_info"

        # response = requests.get(url)
        response, errMsg = SyncUtil.get_request_with_retries(url)
        if response is None or errMsg != '':
            log.error(errMsg)
            return None
        return response.json()

    @classmethod
    def get_last_sync_infos_for_project(cls, project_id):
        url = cls.data_service_path + "/adwords/documents/project_last_sync_info"
        payload = {
            "project_id": project_id
        }
        # response = requests.get(url)
        response, errMsg = SyncUtil.get_request_with_retries(url)
        if response is None or errMsg != '':
            log.error(errMsg)
            return None
        return response.json()

    @classmethod
    def add_all_adwords_documents(cls, project_id, customer_acc_id, docs, doc_type, timestamp):

        for i in range(0, len(docs), cls.BATCH_SIZE):
            batch = docs[i:i+cls.BATCH_SIZE]
            response = cls.add_multiple_adwords_document(project_id, customer_acc_id,
                                     batch, doc_type, timestamp)
            if not response.ok:
                return response

        return response

    # this throws an exception and it is caught at job scheduler level
    @classmethod
    def add_adwords_document(cls, project_id, customer_acc_id, doc, doc_type, timestamp):
        url = cls.data_service_path + "/adwords/documents/add"

        payload = cls.get_payload_for_adwords(project_id, customer_acc_id, doc, doc_type, timestamp)

        # response = requests.post(url, json=payload)
        response, errMsg = SyncUtil.post_request_with_retries(url, payload)
        if response is None or errMsg != '':
            log.error(errMsg)
            raise Exception(errMsg)
        return response

    # this throws an exception and it is caught at job scheduler level
    @classmethod
    def add_multiple_adwords_document(cls, project_id, customer_acc_id, docs, doc_type, timestamp):
        url = cls.data_service_path + "/adwords/documents/add_multiple"
        batch_of_payloads = [cls.get_payload_for_adwords(project_id, customer_acc_id,
                                    doc, doc_type, timestamp) for doc in docs]

        response, errMsg = SyncUtil.post_request_with_retries(url, batch_of_payloads)
        if response is None or errMsg != '':
            log.error(errMsg)
            raise Exception(errMsg)

        return response

    @staticmethod
    def get_payload_for_adwords(project_id, customer_acc_id, doc, doc_type, timestamp):
        return {
            "project_id": project_id,
            "customer_acc_id": customer_acc_id,
            "type_alias": doc_type,
            "value": doc,
            "timestamp": timestamp,
        }

    @classmethod
    def get_gsc_last_sync_infos_for_all_projects(cls):
        url = cls.data_service_path + "/google_organic/documents/last_sync_info"

        # response = requests.get(url)
        response, errMsg = SyncUtil.get_request_with_retries(url)
        if response is None or errMsg != '':
            log.error(errMsg)

        return response.json()

    @classmethod
    def get_gsc_last_sync_infos_for_project(cls, project_id):
        url = cls.data_service_path + "/google_organic/documents/project_last_sync_info"
        payload = {
            "project_id": project_id
        }
        # response = requests.get(url, json=payload)
        response, errMsg = SyncUtil.get_request_with_retries(url, payload)
        if response is None or errMsg != '':
            log.error(errMsg)

        return response.json()

    @classmethod
    def add_all_gsc_documents(cls, project_id, url, doc_type, docs, timestamp):

        for i in range(0, len(docs), cls.BATCH_SIZE):
            batch = docs[i:i+cls.BATCH_SIZE]
            response = cls.add_multiple_gsc_document(project_id, url, doc_type,
                                     batch, timestamp)
            if not response.ok:
                return response

        return response

    # this throws an exception and it is caught at job scheduler level
    @classmethod
    def add_gsc_document(cls, project_id, url_prefix, doc_type, doc, timestamp):
        url = cls.data_service_path + "/google_organic/documents/add"

        payload = cls.get_payload_for_gsc(project_id, url_prefix, doc_type, doc, timestamp)

        # response = requests.post(url, json=payload)
        response, errMsg = SyncUtil.post_request_with_retries(url, payload)
        if response is None or errMsg != '':
            log.error(errMsg)
            raise Exception(errMsg)

        return response

    # this throws an exception and it is caught at job scheduler level
    @classmethod
    def add_multiple_gsc_document(cls, project_id, url_prefix, doc_type, docs, timestamp):
        url = cls.data_service_path + "/google_organic/documents/add_multiple"
        batch_of_payloads = [cls.get_payload_for_gsc(project_id, url_prefix, doc_type,
                                    doc, timestamp) for doc in docs]

        # response = requests.post(url, json=batch_of_payloads)
        response, errMsg = SyncUtil.post_request_with_retries(url, batch_of_payloads)
        if response is None or errMsg != '':
            log.error(errMsg)
            raise Exception(errMsg)

        return response

    @staticmethod
    def get_payload_for_gsc(project_id, url, doc_type, doc, timestamp):
        return {
            "project_id": project_id,
            "url_prefix": url,
            "type": doc_type,
            "value": doc,
            "timestamp": timestamp,
            "id": doc["id"]
        }
    # facebook related processing.
    @classmethod
    def get_facebook_settings(cls):
        url: str = cls.data_service_path + "/facebook/project/settings"

        # response: Response = requests.get(url)
        response, errMsg = SyncUtil.get_request_with_retries(url)
        if response is None or errMsg != '':
            log.error(errMsg)
            return
        return response.json()

    # Add sample response
    # Add failure handling.
    
    # this throws an exception and it is caught at pipeline app level
    @classmethod
    def get_facebook_last_sync_info(cls, project_id, customer_account_id) -> dict:
        url: str = cls.data_service_path + "/facebook/documents/last_sync_info"
        payload: Dict[str, str] = {
            "project_id": project_id,
            "account_id": customer_account_id
        }
        # resp: requests.Response = requests.get(url, json=payload)
        resp, errMsg = SyncUtil.get_request_with_retries(url, payload)
        if resp is None or errMsg != '':
            log.error(errMsg)
            raise Exception(errMsg)
        
        all_info: List = resp.json()
        sync_info_with_type: dict = {}
        for info in all_info:
            sync_info_with_type[info['type_alias']] = info
        return sync_info_with_type

    