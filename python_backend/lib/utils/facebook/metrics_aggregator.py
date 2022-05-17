import json
import logging as log

# NOTE: SIGKILL cant capture this.
# TODO: Give summary separately. Comparing extract and load.
from lib.utils.facebook.sns_notifier import SnsNotifier
from lib.utils.facebook.task_stats import TaskStats
from lib.utils.healthchecks import HealthChecksUtil
from lib.utils.json import JsonUtil

class MetricsAggregator:
    extract_stats = None
    load_stats = None
    type_of_run = None
    permission_error_cache = {}
    etl_stats = {
        "status": "success",
        "task_stats": None,
        "failures": {},
        "success": {},
        "skipped": {},
        "token_failures": {},
    }
    env = None
    HEALTHCHECK_PING_ID = 'f2265955-a71c-42fe-a5ba-36d22a98419c'
    HEALTHCHECK_PING_ID_TOKEN_FAILURE = '2305bb1e-30db-4567-8c1a-2559ea738cbf'

    @classmethod
    def init(cls, env, type_of_run):
        cls.env = env
        sns_notifier = SnsNotifier(env, "facebook_sync")
        cls.type_of_run = type_of_run
        if type_of_run == "extract_and_load_workflow":
            cls.init_extract(sns_notifier)
            cls.init_load(sns_notifier)
        elif type_of_run == "extract_workflow":
            cls.init_extract(sns_notifier)
        else:
            cls.init_load(sns_notifier)

    @classmethod
    def init_extract(cls, sns_notifier):
        cls.extract_stats = TaskStats(sns_notifier)

    @classmethod
    def init_load(cls, sns_notifier):
        cls.load_stats = TaskStats(sns_notifier)

    # Phase - In memory or file.
    @classmethod
    def update_task_stats(cls, task, phase, metric_type, project_id, doc_type, value):
        if task == "extract":
            cls.extract_stats.update_record_stats(phase, metric_type, project_id, doc_type, value)
        elif task == "load":
            cls.load_stats.update_record_stats(phase, metric_type, project_id, doc_type, value)

    # Format of failure status: { message : { doc_type : Set() } }
    @classmethod
    def update_job_stats(cls, project_id, customer_acc_id, doc_type, status, message=""):
        if status == "failed":
            cls.etl_stats["status"] = "Failure on sync."

        if status is None:
            message = "Sync status is missing on response"
            cls.etl_stats["failures"].setdefault(message, {})
            cls.etl_stats["failures"][message].setdefault(doc_type, set())
            cls.etl_stats["failures"][message][doc_type].setdefault(project_id, set())
            cls.etl_stats["failures"][message][doc_type][project_id].add(customer_acc_id)
        elif status == "failed":
            # In our observation we have encountered that "No such object" error comes when extract has failed due to some reason(highly due to token expiry)
            # We'll keep monitoring it manually, if we figure some other cases happening, we'll seprate the two.
            if ("Error validating access token".lower() in message.lower()) or ("No such object".lower() in message.lower()):
                cls.etl_stats["token_failures"].setdefault(message, {})
                cls.etl_stats["token_failures"][message].setdefault(doc_type, set())
                cls.etl_stats["token_failures"][message][doc_type].add(project_id)
            else:
                cls.etl_stats["failures"].setdefault(message, {})
                cls.etl_stats["failures"][message].setdefault(doc_type, {})
                cls.etl_stats["failures"][message][doc_type].setdefault(project_id, set())
                cls.etl_stats["failures"][message][doc_type][project_id].add(customer_acc_id)
        elif status == "skipped":
            cls.etl_stats["skipped"].setdefault(project_id, set())
            cls.etl_stats["skipped"][project_id].add(customer_acc_id)
        else:
            cls.etl_stats["success"].setdefault(project_id, set())
            cls.etl_stats["success"][project_id].add(customer_acc_id)

    @classmethod
    def publish(cls):
        cls.publish_task_stats()
        cls.publish_job_stats()

    @classmethod
    def publish_task_stats(cls):
        if cls.type_of_run == "extract_and_load_workflow":
            cls.extract_stats.publish("extract")
            cls.load_stats.publish("load")
        elif cls.type_of_run == "extract_workflow":
            cls.extract_stats.publish("extract")
        else:
            cls.load_stats.publish("load")

    @classmethod
    def publish_job_stats(cls):
        if cls.type_of_run == "extract_and_load_workflow":
            cls.etl_stats["task_stats"] = cls.compare_load_and_extract()

        if cls.etl_stats["status"] == "success":
            HealthChecksUtil.ping(cls.env, cls.etl_stats["success"], cls.HEALTHCHECK_PING_ID)
        else:
            if len(cls.etl_stats["failures"]) != 0:
                cls.publish_to_healthcheck_failure()
                log.warning("Job has errors. Failed synced Projects and customer accounts are: %s",
                            json.dumps(cls.etl_stats["failures"], default=JsonUtil.serialize_sets))
            if len(cls.etl_stats["token_failures"]) != 0:
                cls.publish_to_healthcheck_token_failure()
                log.warning("Job has errors for token failure. Successfully synced Projects and customer accounts are: %s",
                            json.dumps(cls.etl_stats["token_failures"], default=JsonUtil.serialize_sets))

    @classmethod
    def publish_to_healthcheck_failure(cls):
        HealthChecksUtil.ping(cls.env, cls.etl_stats["failures"], cls.HEALTHCHECK_PING_ID, endpoint="/fail")
    
    @classmethod
    def publish_to_healthcheck_token_failure(cls):
        HealthChecksUtil.ping(cls.env, cls.etl_stats["token_failures"], cls.HEALTHCHECK_PING_ID_TOKEN_FAILURE, endpoint="/fail")

    @classmethod
    def compare_load_and_extract(cls):
        return cls.load_stats.processed_equal_records(cls.extract_stats)
