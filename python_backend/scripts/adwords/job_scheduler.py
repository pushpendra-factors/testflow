import logging as log
import sys
import traceback

import scripts
from scripts.adwords import *
from scripts.adwords.jobs.ad_group_performance_report_job import AdGroupPerformanceReportJob
from scripts.adwords.jobs.ad_groups_job import GetAdGroupsJob
from scripts.adwords.jobs.ad_performance_reports_job import AdPerformanceReportsJob
from scripts.adwords.jobs.ads_job import GetAdsJob
from scripts.adwords.jobs.campaign_performance_reports_job import CampaignPerformanceReportsJob
from scripts.adwords.jobs.campaigns_job import GetCampaignsJob
from scripts.adwords.jobs.click_performance_report_job import ClickPerformanceReportsJob
from scripts.adwords.jobs.customer_account_properties_job import GetCustomerAccountPropertiesJob
from scripts.adwords.jobs.keywords_performance_report_job import KeywordPerformanceReportsJob
from scripts.adwords.jobs.search_performance_reports_job import SeachPerformanceReportsJob

from scripts.adwords.new_jobs.ad_group_performance_report_job import NewAdGroupPerformanceReportJob
from scripts.adwords.new_jobs.ad_groups_job import NewGetAdGroupsJob
from scripts.adwords.new_jobs.ad_performance_reports_job import NewAdPerformanceReportsJob
from scripts.adwords.new_jobs.ads_job import NewGetAdsJob
from scripts.adwords.new_jobs.campaign_performance_reports_job import NewCampaignPerformanceReportsJob
from scripts.adwords.new_jobs.campaigns_job import NewGetCampaignsJob
from scripts.adwords.new_jobs.click_performance_report_job import NewClickPerformanceReportsJob
from scripts.adwords.new_jobs.customer_account_properties_job import NewGetCustomerAccountPropertiesJob
from scripts.adwords.new_jobs.keywords_performance_report_job import NewKeywordPerformanceReportsJob

class JobScheduler:

    @staticmethod
    def validate(next_info, skip_today):
        project_id = next_info.get("project_id")
        customer_acc_id = next_info.get("customer_acc_id")
        doc_type = next_info.get("doc_type_alias")
        timestamp = next_info.get("next_timestamp")
        refresh_token = next_info.get("refresh_token")
        metrics_controller = scripts.adwords.CONFIG.ADWORDS_APP.metrics_controller

        message = ""
        if project_id is None or project_id is 0 or customer_acc_id is None or customer_acc_id == "" or doc_type is None or doc_type == "" or timestamp is None:
            log.error("Invalid project_id: %s or customer_account_id: %s or document_type: %s or timestamp: %s",
                      str(project_id), str(customer_acc_id), str(doc_type), str(timestamp))
            message = "Invalid params project_id or customer_account_id or type or timestamp."


        elif refresh_token is None or refresh_token == "":
            log.error("Invalid refresh token for project_id %s", str(project_id))
            message = "Invalid refresh token."
        if message != "":
            metrics_controller.update_job_stats(project_id, customer_acc_id, doc_type, STATUS_FAILED, message)
            return False

        if metrics_controller.is_permission_denied_previously(project_id, customer_acc_id, refresh_token):
            return False
        return True

    def __init__(self, next_info, skip_today):
        self.permission_error_cache = {}

        self.next_info = next_info
        self.doc_type = next_info.get("doc_type_alias")
        self.customer_acc_id = next_info.get("customer_acc_id")
        self.timestamp = next_info.get("next_timestamp")
        self.project_id = next_info.get("project_id")
        self.refresh_token = next_info.get("refresh_token")
        self.skip_today = skip_today
        self.first_run = next_info.get("first_run")
        self.last_timestamp = next_info.get("last_timestamp")
        self.status = {"project_id": self.project_id, "timestamp": self.timestamp,
                       "doc_type": self.doc_type, "status": "success"}
        self.permission_error_key = str(self.customer_acc_id) + ":" + str(self.refresh_token)

    def sync(self, env, dry, new=False):
        doc_type = self.doc_type
        project_id = self.next_info.get("project_id")
        customer_acc_id = self.next_info.get("customer_acc_id")
        refresh_token = self.next_info.get("refresh_token")
        metrics_controller = scripts.adwords.CONFIG.ADWORDS_APP.metrics_controller
        try:
            if new:
                if doc_type == CUSTOMER_ACCOUNT_PROPERTIES:
                    NewGetCustomerAccountPropertiesJob(self.next_info).start()

                elif doc_type == CAMPAIGNS:
                    NewGetCampaignsJob(self.next_info).start()

                elif doc_type == ADS:
                    NewGetAdsJob(self.next_info).start()

                elif doc_type == AD_GROUPS:
                    NewGetAdGroupsJob(self.next_info).start()

                elif doc_type == CLICK_PERFORMANCE_REPORT:
                    NewClickPerformanceReportsJob(self.next_info).start()

                elif doc_type == CAMPAIGN_PERFORMANCE_REPORT:
                    NewCampaignPerformanceReportsJob(self.next_info).start()

                elif doc_type == AD_PERFORMANCE_REPORT:
                    NewAdPerformanceReportsJob(self.next_info).start()

                elif doc_type == AD_GROUP_PERFORMANCE_REPORT:
                    NewAdGroupPerformanceReportJob(self.next_info).start()

                elif doc_type == KEYWORD_PERFORMANCE_REPORT:
                    NewKeywordPerformanceReportsJob(self.next_info).start()

                else:
                    status = STATUS_FAILED
                    message = "Invalid document type " + str(doc_type)
                    scripts.adwords.CONFIG.ADWORDS_APP.metrics_controller.update_job_stats(project_id, customer_acc_id,
                                                                                        status, message)
            else:
                if doc_type == CUSTOMER_ACCOUNT_PROPERTIES:
                    GetCustomerAccountPropertiesJob(self.next_info).start()

                elif doc_type == CAMPAIGNS:
                    GetCampaignsJob(self.next_info).start()

                elif doc_type == ADS:
                    GetAdsJob(self.next_info).start()

                elif doc_type == AD_GROUPS:
                    GetAdGroupsJob(self.next_info).start()

                elif doc_type == CLICK_PERFORMANCE_REPORT:
                    ClickPerformanceReportsJob(self.next_info).start()

                elif doc_type == CAMPAIGN_PERFORMANCE_REPORT:
                    CampaignPerformanceReportsJob(self.next_info).start()

                elif doc_type == AD_PERFORMANCE_REPORT:
                    AdPerformanceReportsJob(self.next_info).start()

                elif doc_type == AD_GROUP_PERFORMANCE_REPORT:
                    AdGroupPerformanceReportJob(self.next_info).start()

                elif doc_type == SEARCH_PERFORMANCE_REPORT:
                    SeachPerformanceReportsJob(self.next_info).start()

                elif doc_type == KEYWORD_PERFORMANCE_REPORT:
                    KeywordPerformanceReportsJob(self.next_info).start()

                else:
                    status = STATUS_FAILED
                    message = "Invalid document type " + str(doc_type)
                    scripts.adwords.CONFIG.ADWORDS_APP.metrics_controller.update_job_stats(project_id, customer_acc_id,
                                                                                        status, message)


        except Exception as e:
            traceback.print_tb(e.__traceback__)
            str_exception = str(e)
            message = str_exception
            log.warning("Failed with exception: %d %s %s", project_id, customer_acc_id, str_exception)
            if "AuthorizationError.USER_PERMISSION_DENIED" in str_exception:
                metrics_controller.update_permission_cache(customer_acc_id, refresh_token, str_exception)

            elif "ReportDefinitionError.CUSTOMER_SERVING_TYPE_REPORT_MISMATCH" in str_exception:
                message = "Download failed for manager account with exception: " + str_exception

            elif "RateExceededError.RATE_EXCEEDED" in str_exception:
                # https://developers.google.com/adwords/api/docs/guides/rate-limits. Use zero exit to avoid job retry.
                metrics_controller.update_job_stats(project_id, customer_acc_id, doc_type, STATUS_FAILED, message)
                metrics_controller.publish()
                sys.exit(0)

            else:
                message = "Failed with exception: " + str_exception

            metrics_controller.update_job_stats(project_id, customer_acc_id, doc_type, STATUS_FAILED, message)
