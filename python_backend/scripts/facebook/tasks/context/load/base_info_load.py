from datetime import datetime

from lib.task.system.google_storage import GoogleStorage
from lib.utils.facebook.storage_decider import StorageDecider as FacebookStorageDecider
from lib.utils.json import JsonUtil
from lib.utils.sync_util import SyncUtil
from .base_load import BaseLoad


class BaseInfoLoad(BaseLoad):
    NAME = ""
    records = None
    BACKFILL_SUPPORTED = True

    def add_records(self, records):
        self.records = records

    # Getters start here.
    def get_name(self):
        return self.NAME

    def get_next_timestamps(self):
        if self.input_from_timestamp is not None and self.input_to_timestamp is not None:
            return SyncUtil.get_next_timestamps(self.input_from_timestamp, self.input_to_timestamp)
        elif self.input_from_timestamp is not None and self.input_to_timestamp is None:
            return SyncUtil.get_next_timestamps(self.input_from_timestamp, int(datetime.utcnow().strftime('%Y%m%d')))
        elif self.is_first_run():
            return SyncUtil.get_next_timestamps(0, int(datetime.utcnow().strftime('%Y%m%d')))
        else:
            return SyncUtil.get_next_timestamps(self.last_timestamp, int(datetime.utcnow().strftime('%Y%m%d')))

    def is_first_run(self):
        return self.last_timestamp == 0

    def get_max_look_back_timestamp(self):
        return SyncUtil.get_max_look_back_timestamp()

    def get_records(self):
        return self.records

    # Check if could be done better.
    def add_source_attributes(self):
        source = self.get_source()
        if isinstance(source, GoogleStorage):
            bucket_name = FacebookStorageDecider.get_bucket_name(self.env, self.dry)
            file_path = FacebookStorageDecider.get_file_path(self.curr_timestamp, self.project_id,
                                                             self.customer_account_id, self.type_alias)
            source.set_attributes({"bucket_name": bucket_name, "file_path": file_path})
        else:
            file_path = FacebookStorageDecider.get_file_path(self.curr_timestamp, self.project_id,
                                                             self.customer_account_id, self.type_alias)
            source.set_attributes({"base_path": "/usr/local/var/factors/cloud_storage/", "file_path": file_path})
        return

    def add_destination_attributes(self):
        for destination in self.get_destinations():
            destination.set_attributes(self.get_attributes_for_destination())

    def get_attributes_for_destination(self):
        return {
            'project_id': self.project_id,
            'customer_account_id': self.customer_account_id,
            'type_alias': self.type_alias,
            'timestamp': self.curr_timestamp,
            'url': self.facebook_data_service_path + "/facebook/documents/add"
        }

    def read_records(self):
        records_string = self.source.read()
        records = JsonUtil.read(records_string)
        records = self.transform(records)
        self.records = records
        return

    def write_records(self):
        if self.dry:
            return        
        records_string = JsonUtil.create(self.records)
        for destination in self.get_destinations():
            destination.write(records_string)
        return

    def transform(self, records):
        result_records = []
        for record in records:
            record["publisher_platform"] = "facebook"
            result_records.append(record)
        return result_records
