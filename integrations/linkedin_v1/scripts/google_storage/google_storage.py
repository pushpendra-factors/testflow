from google.cloud import storage
import logging as log

class GoogleStorage:
    client = None
    bucket = None
    is_local_run = False
    __instance = None

    @staticmethod
    def get_instance(env=""):
        if GoogleStorage.__instance == None:
            GoogleStorage(env)
        return GoogleStorage.__instance

    def __init__(self, env):
        if env != "staging" and env != "production":
            self.is_local_run = True
            GoogleStorage.__instance = self
            return
        elif env == "staging":
            prefix = "factors-staging"
        else:
            prefix = "factors-production"


        gs_bucket = prefix
        gs_bucket += "-v3"
        self.client = storage.Client()
        self.bucket = self.client.get_bucket(gs_bucket)
        GoogleStorage.__instance = self

    # Overriding the previously present file.
    @classmethod
    def write(self, input_string, job_type, data_state, timestamp, project_id, customer_acc_id, doc_type):
        if self.is_local_run:
            return
        retries = 0
        while retries <3:
            try:
                file_path = self.get_file_path(job_type, data_state, timestamp, project_id, customer_acc_id, doc_type)
                blob = self.bucket.blob(file_path)
                blob.upload_from_string(input_string)
                return
            except Exception as e:
                retries += 1
                if retries >= 3:
                    err = "Failed to upload to s3 for project {0}, ad_account {1}, timestamp {2}, doc_type {3}, job_state {4}, data_state, {5}, err {6}".format(
                        project_id, customer_acc_id, timestamp, doc_type, job_type, data_state, str(e))
                    log.warning(err)
        return
    
    # job_type: daily, t8, t22; data_state = raw/transfromed
    def get_file_path(self, job_type, data_state, timestamp, project_id, customer_acc_id, doc_type):
        return "linkedin_extract/{0}/{1}_{2}_{3}_{4}_{5}.text".format(timestamp,
                    project_id, customer_acc_id, doc_type, job_type, data_state)