import logging as log
import sys
import traceback

import scripts
from scripts.gsc.jobs.google_organic_sync_job import GetSearchConsoleDataJob
from scripts.gsc.jobs.google_organic_page_sync_job import GetSearchConsolePageDataJob
from . import STATUS_FAILED, CUSTOMER_ACCOUNT_PROPERTIES, CAMPAIGNS, ADS, \
    AD_GROUPS, CLICK_PERFORMANCE_REPORT, CAMPAIGN_PERFORMANCE_REPORT, AD_PERFORMANCE_REPORT, \
    AD_GROUP_PERFORMANCE_REPORT, SEARCH_PERFORMANCE_REPORT, KEYWORD_PERFORMANCE_REPORT,COMBINED_LEVEL, PAGE_LEVEL



class JobScheduler:

    @staticmethod
    def validate(next_info, skip_today):
        project_id = next_info.get("project_id")
        url_prefix = next_info.get("url_prefix")
        timestamp = next_info.get("next_timestamp")
        refresh_token = next_info.get("refresh_token")
        doc_type = next_info.get("type")
        metrics_controller = scripts.gsc.CONFIG.GSC_APP.metrics_controller

        message = ""
        if project_id == None or project_id == 0 or url_prefix == None or url_prefix == "" or doc_type == None or doc_type < COMBINED_LEVEL or doc_type > PAGE_LEVEL or timestamp == None:
            log.error("Invalid project_id: %s or url: %s or doc_type: %s or timestamp: %s",
                      str(project_id), str(url_prefix), str(doc_type), str(timestamp))
            message = "Invalid params project_id or url or type or timestamp."


        elif refresh_token is None or refresh_token == "":
            log.error("Invalid refresh token for project_id %d", project_id)
            message = "Invalid refresh token."
        if message != "":
            metrics_controller.update_job_stats(project_id, url_prefix, doc_type, STATUS_FAILED, message)
            return False

        if metrics_controller.is_permission_denied_previously(project_id, url_prefix, refresh_token):
            return False
        return True

    def __init__(self, next_info, skip_today):
        self.permission_error_cache = {}

        self.next_info = next_info
        self.url_prefix = next_info.get("url_prefix")
        self.timestamp = next_info.get("next_timestamp")
        self.project_id = next_info.get("project_id")
        self.refresh_token = next_info.get("refresh_token")
        self.doc_type = next_info.get("type")
        self.skip_today = skip_today
        self.first_run = next_info.get("first_run")
        self.last_timestamp = next_info.get("last_timestamp")
        self.status = {"project_id": self.project_id, "timestamp": self.timestamp,
                       "url_prefix": self.url_prefix, "doc_type": self.doc_type, "status": "success"}
        self.permission_error_key = str(self.url_prefix) + ":" + str(self.refresh_token)

    def sync(self, env, dry):
        project_id = self.next_info.get("project_id")
        url_prefix = self.next_info.get("url_prefix")
        refresh_token = self.next_info.get("refresh_token")
        doc_type = self.next_info.get("type")
        metrics_controller = scripts.gsc.CONFIG.GSC_APP.metrics_controller
        try:
            if doc_type == COMBINED_LEVEL:
                GetSearchConsoleDataJob(self.next_info).start()
            elif doc_type == PAGE_LEVEL:
                GetSearchConsolePageDataJob(self.next_info).start()
            else:
                status = STATUS_FAILED
                message = "Invalid document type " + str(doc_type)
                metrics_controller.update_job_stats(project_id, url_prefix, status, message)

        except Exception as e:
            traceback.print_tb(e.__traceback__)
            str_exception = str(e)
            message = str_exception
            log.warning("Failed with exception: %d %s %s", project_id, url_prefix, str_exception)
            if "AuthorizationError.USER_PERMISSION_DENIED" in str_exception:
                metrics_controller.update_permission_cache(url_prefix, refresh_token, str_exception)

            elif "ReportDefinitionError.CUSTOMER_SERVING_TYPE_REPORT_MISMATCH" in str_exception:
                message = "Download failed for manager account with exception: " + str_exception

            elif "quotaExceeded" in str_exception:
                metrics_controller.update_job_stats(project_id, url_prefix, doc_type, STATUS_FAILED, message)
                metrics_controller.publish()
                sys.exit(0)

            else:
                message = "Failed with exception: " + str_exception
            metrics_controller.update_job_stats(project_id, url_prefix, doc_type, STATUS_FAILED, message)

            return ("Failed with exception: %d %s %s", project_id, url_prefix, str_exception)
