import logging as log

from lib.utils.csv import CsvUtil
from lib.adwords.oauth_service.fetch_service import FetchService
from scripts.adwords.new_jobs.fields_mapping import FieldsMapping

import scripts

from .base_job import BaseJob
from .query import QueryBuilder


class NewGetCustomerAccountPropertiesJob(BaseJob):

    def __init__(self, next_info):
        super().__init__(next_info)

    def new_get_customer_account(self, ads_service):

        EXTRACT_FIELDS = [
            "customer.id",
            "customer.currency_code",
            "customer.time_zone",
            "customer.manager",
            "customer.test_account",
        ]

        HEADERS_VMAX = [
            "customer_id",
            "currency_code",
            "date_timezone",
            "can_manage_clients",
            "test_account"
        ]

        query = (QueryBuilder()
                            .Select(EXTRACT_FIELDS)
                            .From("customer")
                            .Build())
        
        stream = ads_service.search_stream(customer_id=self._customer_acc_id, query=query)

        current_account = None
        for batch in stream:
            for row in batch.results:
                dict = {}
                for i in range(len(HEADERS_VMAX)):
                    dict[HEADERS_VMAX[i]] = QueryBuilder.getattribute(row, EXTRACT_FIELDS[i])
                current_account = dict

        if current_account is None:
            log.error("Customer account not found on list of accounts. Failed to get properties.")
            raise Exception("Failed to get properties. customer account" + str(
                self._customer_acc_id))
        current_account["can_manage_clients"] = FieldsMapping.BOOLEAN_MAPPING[str(current_account["can_manage_clients"])]
        current_account["test_account"] = FieldsMapping.BOOLEAN_MAPPING[str(current_account["test_account"])]
        return current_account

    def get_customer_account(self, customer_service):
        customer_accounts = customer_service.getCustomers()

        current_account = None
        for account in customer_accounts:
            if str(account["customerId"]) == self._customer_acc_id:
                current_account = account

        if current_account is None:
            log.error("Customer account not found on list of accounts. Failed to get properties.")
            raise Exception("Failed to get properties. customer account" + str(
                self._customer_acc_id) + " not found on list of account " + str(customer_accounts))

        return current_account

    @staticmethod
    def get_response(current_account):
        properties = {}
        try:
            properties["customer_id"] = current_account["customerId"]
            properties["currency_code"] = current_account["currencyCode"]
            properties["date_timezone"] = current_account["dateTimeZone"]
            properties["can_manage_clients"] = current_account["canManageClients"]
            properties["test_account"] = current_account["testAccount"]
        except Exception as e:
            log.error("Failed to get customer account properties: %s", str(e))
            return [properties]

        return [properties], 1

    def start(self):
        # ads_service  = FetchService(scripts.adwords.CONFIG.ADWORDS_OAUTH).new_get_service("GoogleAdsService", self._refresh_token)
        # return self.new_get_customer_account(ads_service)
        return 
        
    def old_start(self):
        # customer_service = FetchService(scripts.adwords.CONFIG.ADWORDS_OAUTH).get_customer_accounts(self._refresh_token)
        # current_account = self.get_customer_account(customer_service)
        # return GetCustomerAccountPropertiesJob.get_response(current_account)
        return
