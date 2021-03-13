import logging as log

import scripts
from lib.adwords.oauth_service.fetch_service import FetchService
from .base_job import BaseJob


class GetCustomerAccountPropertiesJob(BaseJob):

    def __init__(self, next_info):
        super().__init__(next_info)

    def get_customer_account(self, customer_service):
        customer_accounts = customer_service.getCustomers()

        current_account = None
        for account in customer_accounts:
            if str(account["customerId"]) == self._customer_account_id:
                current_account = account

        if current_account is None:
            log.error("Customer account not found on list of accounts. Failed to get properties.")
            raise Exception("Failed to get properties. customer account" + str(
                self._customer_account_id) + " not found on list of account " + str(customer_accounts))

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
        customer_service = FetchService(scripts.adwords.CONFIG.ADWORDS_OAUTH).get_customer_accounts(self._refresh_token)
        current_account = self.get_customer_account(customer_service)
        return GetCustomerAccountPropertiesJob.get_response(current_account)
