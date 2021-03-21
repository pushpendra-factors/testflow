from lib.utils.google_storage_file import GoogleStorageFile
from lib.utils.local_storage import LocalStorage
from scripts.adwords import DEVELOPMENT, TEST, STAGING


class JobStorage:
    storage_file = None
    expiry_time = None
    env = None

    @classmethod
    def init(cls, env, dry, project_name):
        cls.env = env
        if env in [DEVELOPMENT, TEST]:
            cls.storage_file = LocalStorage()
        else:
            bucket_name = cls.get_bucket_name(dry)
            cls.storage_file = GoogleStorageFile(project_name, bucket_name)

    @classmethod
    def get_bucket_name(cls, dry):
        if cls.env == STAGING:
            prefix = "factors-staging"
        else:
            prefix = "factors-production"

        gs_bucket = prefix
        if dry:
            gs_bucket = gs_bucket + "-tmp"
        return gs_bucket

    @classmethod
    def write(cls, input_string, timestamp, project_id, customer_acc_id, doc_type):
        file_path = JobStorage.get_file_path(timestamp, project_id, customer_acc_id, doc_type)
        cls.storage_file.write(input_string, file_path)

    @classmethod
    def read(cls, timestamp, project_id, customer_acc_id, doc_type):
        file_path = JobStorage.get_file_path(timestamp, project_id, customer_acc_id, doc_type)
        return cls.storage_file.read(file_path)

    @staticmethod
    def get_file_path(timestamp, project_id, customer_acc_id, doc_type):
        return "adwords_extract/{0}/{1}/{2}/{3}.csv".format(timestamp, project_id, customer_acc_id, doc_type)
