import itertools

from lib.task.system.google_storage import GoogleStorage
from lib.utils.facebook.storage_decider import StorageDecider as FacebookStorageDecider
from lib.utils.json import JsonUtil
from lib.utils.sync_util import SyncUtil
from .base_load import BaseLoad


class BaseReportLoad(BaseLoad):
    NAME = ""
    FIELDS = []
    SEGMENTS = []
    METRICS = []
    records = None

    # Getters start here.
    def get_name(self):
        return self.NAME

    # fields + metrics.
    def get_fields(self):
        return list(itertools.chain(self.FIELDS, self.METRICS, self.SEGMENTS))

    def get_next_timestamps(self):
        if self.input_from_timestamp is not None and self.input_to_timestamp is not None:
            return SyncUtil.get_next_timestamps(self.input_from_timestamp, self.input_to_timestamp)
        elif self.input_from_timestamp is not None and self.input_to_timestamp is None:
            return SyncUtil.get_next_timestamps(self.input_from_timestamp)
        else:
            return SyncUtil.get_next_timestamps(self.last_timestamp)

    def read_dependencies(self, curr_timestamp):
        pass

    def read_current_records(self, curr_timestamp):
        pass

    def merge_dependencies_and_current_task_records(self):
        pass

    # Check if could be done better.
    def add_source_attributes(self):
        self.add_source_attributes_for_type_alias(self.type_alias)
        return

    def add_source_attributes_for_type_alias(self, type_alias):
        source = self.get_source()
        if isinstance(source, GoogleStorage):
            bucket_name = FacebookStorageDecider.get_bucket_name(self.env, self.dry)
            file_path = FacebookStorageDecider.get_file_path(self.curr_timestamp, self.project_id,
                                                             self.customer_account_id, type_alias)
            source.set_attributes({"bucket_name": bucket_name, "file_path": file_path})
        else:
            file_path = FacebookStorageDecider.get_file_path(self.curr_timestamp, self.project_id,
                                                             self.customer_account_id, type_alias)
            source.set_attributes({"base_path": "/usr/local/var/factors/cloud_storage/", "file_path": file_path})
        return

    @staticmethod
    def transform_video_attributes(record):
        for attribute in ["video_p50_watched_actions", "video_p25_watched_actions", "video_30_sec_watched_actions",
                          "video_p100_watched_actions", "video_p75_watched_actions"]:
            if record.get(attribute):
                result_value = 0
                for ad_action_state in record.get(attribute):
                    result_value += int(ad_action_state["value"])
                record[attribute] = result_value
        return record

    @staticmethod
    def transform_action_attributes(record):
        if record.get("actions") is None:
            return record

        for action_hash in record.get("actions"):
            record["action_"+action_hash["action_type"]] = action_hash["value"]

        return record

    def add_destination_attributes(self):
        for destination in self.get_destinations():
            destination.set_attributes(self.get_attributes_for_destination())

    def get_attributes_for_destination(self):
        return {
            "project_id": self.project_id,
            'customer_account_id': self.customer_account_id,
            'type_alias': self.type_alias,
            'timestamp': self.curr_timestamp,
            'url': self.facebook_data_service_path + "/facebook/documents/add"
        }

    def write_records(self):
        records_string = JsonUtil.create(self.records)
        for destination in self.get_destinations():
            destination.write(records_string)
        return
