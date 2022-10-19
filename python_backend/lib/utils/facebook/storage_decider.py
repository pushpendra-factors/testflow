from scripts.adwords import STAGING


class StorageDecider:

    @staticmethod
    def get_bucket_name(env, dry):
        if env == STAGING:
            prefix = "factors-staging"
        else:
            prefix = "factors-production"

        gs_bucket = prefix
        if dry:
            gs_bucket += "-tmp"
        else:
            gs_bucket += "-v3"
        return gs_bucket

    def get_file_path(timestamp, project_id, customer_acc_id, doc_type):
        return "facebook_extract/{0}/{1}/{2}/{3}.csv".format(timestamp, project_id, customer_acc_id, doc_type)
