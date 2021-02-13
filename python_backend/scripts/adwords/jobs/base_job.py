import logging as log


# IMP: Take care of status.
class BaseJob:
    PAGE_SIZE = 200

    def __init__(self, next_info):
        self._project_id = next_info.get("project_id")
        self._customer_account_id = next_info.get("customer_acc_id")
        self._refresh_token = next_info.get("refresh_token")
        self._timestamp = next_info.get("next_timestamp")
        self._doc_type = next_info.get("doc_type_alias")
        # self._status = {"project_id": self._project_id, "timestamp": self._timestamp, "doc_type": self._doc_type,
        #                 "status": "success"}

    def start(self):
        log.warning("ETL for project: %s, cutomer_account_id: %s, document_type: %s, timestamp: %s",
                    str(self._project_id), self._customer_acc_id, self._doc_type, str(self._timestamp))
        """  Override this in the sub classes. """
