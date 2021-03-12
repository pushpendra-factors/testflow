# TODO: Check the error handling.
import requests
import logging as log

from scripts.adwords.jobs.multiple_requests_fetch_job import MultipleRequestsFetchJob


class FactorsDataService:
    data_service_path = None

    @classmethod
    def init(cls, config):
        cls.data_service_path = config.ADWORDS_APP.get_data_service_path()

    @classmethod
    def add_refresh_token(cls, session, payload):
        if session is None or session == "":
            log.error("Invalid session cookie on add_refresh_token request.")
            return

        url = cls.data_service_path + "/adwords/add_refresh_token"

        response = requests.post(url, json=payload)
        if not response.ok:
            log.error("Failed updating adwords integration with response : %d, %s",
                      response.status_code, response.text)
            return

        return response

    @classmethod
    def get_adwords_refresh_token(cls, project_id):
        url = cls.data_service_path + "/adwords/get_refresh_token"
        # project_id as str for consistency on json.
        payload = {"project_id": str(project_id)}
        response = requests.post(url, json=payload)
        if not response.ok:
            log.error("Failed getting adwords integration with response : %d, %s",
                      response.status_code, response.text)
            return
        return response

    @classmethod
    def get_last_sync_infos(cls):
        url = cls.data_service_path + "/adwords/documents/last_sync_info"

        response = requests.get(url)
        if not response.ok:
            log.error("Failed to get sync data: %d, %s",
                      response.status_code, response.text)

        log.warning("Got adwords last sync info.")
        return response.json()

    @classmethod
    def add_adwords_document(cls, project_id, customer_acc_id, doc, doc_type, timestamp):
        log.warning("Calling the adwords data service - add documents.")
        url = cls.data_service_path + "/adwords/documents/add"

        payload = {
            "project_id": project_id,
            "customer_acc_id": customer_acc_id,
            "type_alias": doc_type,
            "value": doc,
            "timestamp": timestamp,
        }

        response = requests.post(url, json=payload)
        if not response.ok:
            log.error("Failed to add response %s to adwords warehouse: %d, %s",
                      doc_type, response.status_code, response.text)

        return response

    @classmethod
    def add_all_adwords_documents(cls, project_id, customer_acc_id, docs, doc_type, timestamp):
        log.warning("Calling the adwords data service - add all documents.")
        for doc in docs:
            cls.add_adwords_document(project_id, customer_acc_id,
                                     doc, doc_type, timestamp)

    @classmethod
    def add_all_adwords_documents_for_first_run(cls, project_id, customer_acc_id, docs, doc_type):
        log.warning("Calling the adwords data service - add all documents first run.")
        for doc in docs:
            cls.add_adwords_document(project_id, customer_acc_id, doc, doc_type,
                                     doc[MultipleRequestsFetchJob.TIMESTAMP_FIELD])
