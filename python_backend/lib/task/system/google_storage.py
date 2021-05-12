
from google.cloud import storage

from .base import BaseSystem


class GoogleStorage(BaseSystem):

    def read(self):
        client = storage.Client()
        bucket = client.get_bucket(self.system_attributes["bucket_name"])
        blob = bucket.blob(self.system_attributes["file_path"])
        if blob is None:
            return None
        return blob.download_as_text()

    # Overriding the previously present file.
    def write(self, input_string):
        if self.system_attributes["file_override"]:
            self.override_during_write(input_string)
        else:
            self.non_override_during_write(input_string)

    def override_during_write(self, input_string):
        client = storage.Client()
        bucket = client.get_bucket(self.system_attributes["bucket_name"])
        blob = bucket.blob(self.system_attributes["file_path"])
        blob.upload_from_string(input_string)
        return

    def non_override_during_write(self, input_string):
        client = storage.Client()
        bucket = client.get_bucket(self.system_attributes["bucket_name"])
        blob = bucket.blob(self.system_attributes["file_path"])
        if blob is not None:
            blob.upload_from_string(input_string)
        return
