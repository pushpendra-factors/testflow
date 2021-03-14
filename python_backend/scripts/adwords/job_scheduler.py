import sys

import logging as log

import scripts
from lib.data_services.factors_data_service import FactorsDataService
from lib.sns_notifier import SnsNotifier
from lib.utils.healthchecks import HealthchecksUtil
from lib.utils.time import TimeUtil
from scripts.adwords.jobs.ad_groups_job import GetAdGroupsJob
from scripts.adwords.jobs.ad_performance_reports_job import AdPerformanceReportsJob
from scripts.adwords.jobs.ads_job import GetAdsJob
from scripts.adwords.jobs.campaign_performance_reports_job import CampaignPerformanceReportsJob
from scripts.adwords.jobs.campaigns_job import GetCampaignsJob
from scripts.adwords.jobs.click_performance_report_job import ClickPerformanceReportsJob
from scripts.adwords.jobs.customer_account_properties_job import GetCustomerAccountPropertiesJob
from scripts.adwords.jobs.keywords_performance_report_job import KeywordPerformanceReportsJob
from scripts.adwords.jobs.search_performance_reports_job import SeachPerformanceReportsJob
from . import STATUS_FAILED, STATUS_SKIPPED, APP_NAME, etl_record_stats, CUSTOMER_ACCOUNT_PROPERTIES, CAMPAIGNS, ADS, \
    AD_GROUPS, CLICK_PERFORMANCE_REPORT, CAMPAIGN_PERFORMANCE_REPORT, AD_PERFORMANCE_REPORT, \
    AD_GROUP_PERFORMANCE_REPORT, SEARCH_PERFORMANCE_REPORT, KEYWORD_PERFORMANCE_REPORT, \
    HEALTHCHECKS_ADWORDS_SYNC_PING_ID
from .jobs.ad_group_performance_report_job import AdGroupPerformanceReportJob
from .jobs.reports_fetch_job import ReportsFetch


class JobScheduler:

    def _validate(self, next_info, skip_today):
        project_id = next_info.get("project_id")
        customer_acc_id = next_info.get("customer_acc_id")
        doc_type = next_info.get("doc_type_alias")
        timestamp = next_info.get("next_timestamp")
        refresh_token = next_info.get("refresh_token")
        status = {"project_id": project_id, "timestamp": timestamp, "doc_type": doc_type, "status": "success"}

        if project_id is None or project_id is 0 or customer_acc_id is None or customer_acc_id == "" or doc_type is None or doc_type == "" or timestamp is None:
            log.error("Invalid project_id: %s or customer_account_id: %s or document_type: %s or timestamp: %s",
                      str(project_id), str(customer_acc_id), str(doc_type), str(timestamp))
            status["status"] = STATUS_FAILED
            status["message"] = "Invalid params project_id or customer_account_id or type or timestamp."
            return status

        if refresh_token is None or refresh_token == "":
            log.error("Invalid refresh token for project_id %s", refresh_token)
            status["status"] = STATUS_FAILED
            status["message"] = "Invalid refresh token."
            return status

        permission_error_key = str(customer_acc_id) + ':' + str(refresh_token)
        if permission_error_key in self.permission_error_cache:
            log.error("Skipping sync user permission as its denied already for project %s, "
                      "'customer_acc_id:refresh_token' : ""%s", str(project_id), permission_error_key)
            return status

        if skip_today and TimeUtil.is_today(timestamp):
            log.warning("Skipped sync for today for project_id %s doc_type %s.", str(project_id), doc_type)
            status["status"] = STATUS_SKIPPED
            status["message"] = "Skipped sync for today."
            return status

    def __init__(self, next_info, skip_today):
        self.permission_error_cache = {}
        self._validate(next_info, skip_today)
        self.next_info = next_info
        self.doc_type = next_info.get('doc_type_alias')
        self.customer_acc_id = next_info.get("customer_acc_id")
        self.timestamp = next_info.get("next_timestamp")
        self.project_id = next_info.get("project_id")
        self.refresh_token = next_info.get("refresh_token")
        self.skip_today = skip_today
        self.first_run = next_info.get("first_run")
        self.last_timestamp = next_info.get("last_timestamp")
        self.status = {"project_id": self.project_id, "timestamp": self.timestamp,
                       "doc_type": self.doc_type, "status": "success"}
        self.permission_error_key = str(self.customer_acc_id) + ':' + str(self.refresh_token)

    def sync(self, env, dry):
        doc_type = self.doc_type
        try:
            if doc_type == CUSTOMER_ACCOUNT_PROPERTIES:
                # docs, req_count = GetCustomerAccountPropertiesJob(self.next_info).start()
                self.status["status"] = STATUS_SKIPPED
                self.status["message"] = "CustomerAccountProperties is skipped."
                return self.status

            elif doc_type == CAMPAIGNS:
                docs, req_count = GetCampaignsJob(self.next_info).start()

            elif doc_type == ADS:
                # docs, req_count = GetAdsJob(self.next_info).start()
                self.status["status"] = STATUS_SKIPPED
                self.status["message"] = "Ads is skipped."
                return self.status

            elif doc_type == AD_GROUPS:
                docs, req_count = GetAdGroupsJob(self.next_info).start()

            elif doc_type == CLICK_PERFORMANCE_REPORT:
                docs, req_count = ClickPerformanceReportsJob(self.next_info).start()

            elif doc_type == CAMPAIGN_PERFORMANCE_REPORT:
                docs, req_count = CampaignPerformanceReportsJob(self.next_info).start()

            elif doc_type == AD_PERFORMANCE_REPORT:
                docs, req_count = AdPerformanceReportsJob(self.next_info).start()

            elif doc_type == AD_GROUP_PERFORMANCE_REPORT:
                docs, req_count = AdGroupPerformanceReportJob(self.next_info).start()

            elif doc_type == SEARCH_PERFORMANCE_REPORT:
                docs, req_count = SeachPerformanceReportsJob(self.next_info).start()

            elif doc_type == KEYWORD_PERFORMANCE_REPORT:
                docs, req_count = KeywordPerformanceReportsJob(self.next_info).start()

            else:
                log.error("Invalid document to sync from adwords: %s", str(doc_type))
                self.status["status"] = STATUS_FAILED
                self.status["message"] = "Invalid document type " + str(doc_type)
                return self.status

            etl_record_stats.update(self.project_id, doc_type, req_count)

            log.warning("Started Load of job for Project Id: %s, Timestamp: %d, Doc Type: %s", self.project_id,
                        self.timestamp, self.doc_type)
            if len(docs) > 0:
                if dry:
                    log.error("Dry run. Skipped add adwords documents to db.")
                elif ReportsFetch.doesnt_contains_historical_data(self.last_timestamp, doc_type) and self.first_run:
                    FactorsDataService.add_all_adwords_documents_for_first_run(self.project_id, self.customer_acc_id, docs,
                                                                 doc_type)
                else:
                    FactorsDataService.add_all_adwords_documents(self.project_id, self.customer_acc_id, docs,
                                                                 doc_type, self.timestamp)
            else:
                FactorsDataService.add_adwords_document(self.project_id, self.customer_acc_id, {}, doc_type,
                                                        self.timestamp)
            log.warning("Completed Load of job for Project Id: %s, Timestamp: %d, Doc Type: %s", self.project_id,
                        self.timestamp, self.doc_type)
        except Exception as e:
            str_exception = str(e)
            if "AuthorizationError.USER_PERMISSION_DENIED" in str_exception:
                self.permission_error_cache[self.permission_error_key] = str_exception
                self.status["status"] = STATUS_FAILED
                self.status["message"] = "Failed with exception: " + str_exception
                return self.status

            if "ReportDefinitionError.CUSTOMER_SERVING_TYPE_REPORT_MISMATCH" in str_exception:
                log.error("[Project: %s, Type: %s] Sync failed, Trying to download report from manager account.",
                          str(self.project_id), doc_type)
                self.status["status"] = STATUS_FAILED
                self.status["message"] = "Download failed for manager account with exception: " + str_exception
                return self.status

            log.error("[Project: %s, Type: %s] Sync failed with exception: %s", str(self.project_id), doc_type,
                      str_exception)
            if "RateExceededError.RATE_EXCEEDED" in str_exception:
                # TODO: Do not exit? Stop downloading reports.
                # Continue downloading other objects.
                SnsNotifier.notify(env, APP_NAME,
                                   {"status": STATUS_FAILED, "exception": str_exception, "requests": etl_record_stats})
                
                HealthchecksUtil.ping_healthcheck(scripts.adwords.CONFIG.ADWORDS_APP.env,
                                          HEALTHCHECKS_ADWORDS_SYNC_PING_ID,{"request": etl_record_stats.__dict__} , endpoint="/fail")
                sys.exit(0)  # Use zero exit to avoid job retry.

            self.status["status"] = STATUS_FAILED
            self.status["message"] = "Failed with exception: " + str_exception
            return self.status

        self.status["status"] = "success"
        return self.status
