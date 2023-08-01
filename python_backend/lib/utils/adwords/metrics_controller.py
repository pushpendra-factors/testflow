import json
import logging as log
import scripts
from lib.utils.healthchecks import HealthChecksUtil
from lib.utils.json import JsonUtil
from scripts.adwords import STATUS_SKIPPED, STATUS_FAILED, FAILURE_MESSAGE, SUCCESS_MESSAGE, EMPTY_RESPONSE_GSC
from lib.utils.adwords.job_task_stats import JobTaskStats

# NOTE: SIGKILL cant capture this.
# TODO: Give summary separately. Comparing extract and load.
class MetricsController:
    extract_stats = None
    load_stats = None
    type_of_run = None
    permission_error_cache = {}
    etl_stats = {
        "status": SUCCESS_MESSAGE,
        "task_stats": None,
        "failures": {},
        "token_failures": {},
        "success": {}
    }
    ADWORDS_SYNC_PING_ID = "188cbf7c-0ea1-414b-bf5c-eee47c12a0c8"
    ADWORDS_PING_ID_TOKEN_FAILURE = "e6b2efa8-ff32-41ad-b5cd-25fac93a70d9"
    GSC_SYNC_PING_ID = "914866ad-dab5-4ec9-bad1-2b6ef6eab6f5"
    GSC_PING_ID_TOKEN_FAILURE = "12132b95-3ef0-45ee-a7d7-7c3a796481f3"

    @classmethod
    def init(cls, type_of_run):
        cls.type_of_run = type_of_run
        if type_of_run == scripts.adwords.EXTRACT_AND_LOAD:
            cls.init_extract()
            cls.init_load()
        elif type_of_run == scripts.adwords.EXTRACT:
            cls.init_extract()
        else:
            cls.init_load()

    @classmethod
    def init_extract(cls):
        cls.extract_stats = JobTaskStats()

    @classmethod
    def init_load(cls):
        cls.load_stats = JobTaskStats()

    # key_string takes in customer_account_id or url_prefix
    @classmethod
    def is_permission_denied_previously(cls, project_id, key_string, refresh_token):
        key = "{0}:{1}".format(key_string, refresh_token)
        if key in cls.permission_error_cache:
            log.error("Skipping sync user permission denied already for project %s, 'key_string:refresh_token' : %s", 
                str(project_id), key)
            return True
        return False

    @classmethod
    def update_permission_cache(cls, key_string, refresh_token, message):
        key = "{0}:{1}".format(key_string, refresh_token)
        cls.permission_error_cache[key] = message


    # Phase - In memory or file.
    @classmethod
    def update_task_stats(cls, task, phase, metric_type, project_id, doc_type, value):
        if task == scripts.adwords.EXTRACT:
            cls.extract_stats.update_record_stats(phase, metric_type, project_id, doc_type, value)
        elif task == scripts.adwords.LOAD:
            cls.load_stats.update_record_stats(phase, metric_type, project_id, doc_type, value)

    # Format of failure status: { message : { doc_type : Set() } }
    @classmethod
    def update_job_stats(cls, project_id, customer_acc_id, doc_type, status, message=""):
        if status == STATUS_FAILED:
            cls.etl_stats["status"] = FAILURE_MESSAGE

        if status is None:
            message = "Sync status is missing on response"
            cls.etl_stats["failures"].setdefault(message, {})
            cls.etl_stats["failures"][message].setdefault(doc_type, set())
            cls.etl_stats["failures"][message][doc_type].add(project_id)
        elif status == STATUS_FAILED:

            if "invalid_grant" in message.lower() or "PERMISSION_DENIED".lower() in message.lower():
                cls.etl_stats["token_failures"].setdefault(message, set())
                cls.etl_stats["token_failures"][message].add(project_id)
            else:
                cls.etl_stats["failures"].setdefault(message, {})
                cls.etl_stats["failures"][message].setdefault(doc_type, set())
                cls.etl_stats["failures"][message][doc_type].add(project_id)
        else:
            cls.etl_stats["success"].setdefault(project_id, set())
            cls.etl_stats["success"][project_id].add(customer_acc_id)

    @classmethod
    def update_gsc_job_stats(cls, project_id, url, status, message=""):
        if status == STATUS_FAILED:
            cls.etl_stats["status"] = FAILURE_MESSAGE

        if status is None:
            cls.etl_stats["failures"].append("Sync status is missing on response")
        elif status == STATUS_FAILED:
            log.warning("Ashhar %s", message)
            if ("invalid_grant" in message.lower() or "PERMISSION_DENIED".lower() in message.lower() 
                or "invalid params" in message.lower() or "access_token" in message.lower() 
                or "refresh_token" in message.lower()):
                
                cls.etl_stats["token_failures"].setdefault(message, {})
                cls.etl_stats["token_failures"][message].setdefault(project_id, set())
                cls.etl_stats["token_failures"][project_id].add(url)
            
            elif EMPTY_RESPONSE_GSC in message.lower():
                cls.etl_stats["warning"].setdefault(message, set())
                cls.etl_stats["warning"][message].setdefault(project_id, set())
                cls.etl_stats["warning"][project_id].add(url)
            else:
                cls.etl_stats["failures"].setdefault(message, {})
                cls.etl_stats["failures"][message].setdefault(project_id, set())
                cls.etl_stats["failures"][project_id].add(url)
        else:
            cls.etl_stats["success"].setdefault(project_id, set())
            cls.etl_stats["success"][project_id].add(url)

# todo @ashhar: merge gsc and adwords pubish functions
    @classmethod
    def publish(cls):
        cls.publish_task_stats()
        cls.publish_job_stats()
    
    @classmethod
    def publish_gsc(cls):
        cls.publish_gsc_task_stats()
        cls.publish_gsc_job_stats()

    @classmethod
    def publish_task_stats(cls):
        if cls.type_of_run == scripts.adwords.EXTRACT_AND_LOAD:
            cls.extract_stats.publish("extract")
            cls.load_stats.publish("load")
        elif cls.type_of_run == scripts.adwords.EXTRACT:
            cls.extract_stats.publish("extract")
        else:
            cls.load_stats.publish("load")

    @classmethod
    def publish_gsc_task_stats(cls):
        if cls.type_of_run == scripts.adwords.EXTRACT_AND_LOAD:
            cls.extract_stats.publish_gsc("extract")
            cls.load_stats.publish_gsc("load")
        elif cls.type_of_run == scripts.adwords.EXTRACT:
            cls.extract_stats.publish_gsc("extract")
        else:
            cls.load_stats.publish_gsc("load")

    @classmethod
    def publish_job_stats(cls):
        if cls.type_of_run == scripts.adwords.EXTRACT_AND_LOAD:
            cls.etl_stats["task_stats"] = cls.compare_load_and_extract()

        if cls.etl_stats["status"] == SUCCESS_MESSAGE or len(cls.etl_stats["failures"].keys()) == 0:
            HealthChecksUtil.ping(scripts.adwords.CONFIG.ADWORDS_APP.env, cls.etl_stats["success"], cls.ADWORDS_SYNC_PING_ID)
        else:
            HealthChecksUtil.ping(scripts.adwords.CONFIG.ADWORDS_APP.env, cls.etl_stats["failures"], cls.ADWORDS_SYNC_PING_ID, endpoint="/fail")
            log.warning("Job has errors. Failed synced Projects and customer accounts are: %s", json.dumps(cls.etl_stats["failures"], default=JsonUtil.serialize_sets))

        if len(cls.etl_stats["token_failures"].keys()) != 0:
            HealthChecksUtil.ping(scripts.adwords.CONFIG.ADWORDS_APP.env, cls.etl_stats["token_failures"], cls.ADWORDS_PING_ID_TOKEN_FAILURE, endpoint="/fail")
            log.warning("Job has token errors. Failed synced Projects and customer accounts are: %s", json.dumps(cls.etl_stats["token_failures"], default=JsonUtil.serialize_sets))

    @classmethod
    def publish_gsc_job_stats(cls):
        log.warning("Ashhar etl stats %s", json.dumps(cls.etl_stats, default=JsonUtil.serialize_sets))
        if cls.type_of_run == scripts.adwords.EXTRACT_AND_LOAD:
            cls.etl_stats["task_stats"] = cls.compare_load_and_extract()

        if cls.etl_stats["status"] == SUCCESS_MESSAGE:
            stats_to_ping = {"success": cls.etl_stats["success"], "warning": cls.etl_stats["warning"]}
            HealthChecksUtil.ping(scripts.gsc.CONFIG.GSC_APP.env, stats_to_ping, cls.GSC_SYNC_PING_ID)
        else:
            stats_to_ping = {"failures": cls.etl_stats["failures"], "warning": cls.etl_stats["warning"]}
            HealthChecksUtil.ping(scripts.gsc.CONFIG.GSC_APP.env, stats_to_ping, cls.GSC_SYNC_PING_ID, endpoint="/fail")
            log.warning("Job has errors. Failed synced Projects and customer accounts are: %s", json.dumps(stats_to_ping, default=JsonUtil.serialize_sets))

        if len(cls.etl_stats["token_failures"].keys()) != 0:
            HealthChecksUtil.ping(scripts.adwords.CONFIG.GSC_APP.env, cls.etl_stats["token_failures"], cls.GSC_PING_ID_TOKEN_FAILURE, endpoint="/fail")
            log.warning("Job has token errors. Failed synced Projects and customer accounts are: %s", json.dumps(cls.etl_stats["token_failures"], default=JsonUtil.serialize_sets))

    @classmethod
    def compare_load_and_extract(cls):
        return cls.load_stats.processed_equal_records(cls.extract_stats)
