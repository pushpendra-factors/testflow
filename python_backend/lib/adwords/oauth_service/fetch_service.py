from googleads import oauth2, adwords

ADWORDS_CLIENT_USER_AGENT = "FactorsAI (https://www.factors.ai)"


class FetchService:
    VERSION = 'v201809'

    def __init__(self, config):
        self.CONFIG = config

    def get_customer_accounts(self, refresh_token):
        oauth_client = oauth2.GoogleRefreshTokenClient(self.CONFIG.client_id, self.CONFIG.client_secret, refresh_token)
        adwords_client = adwords.AdWordsClient(self.CONFIG.developer_token, oauth_client, ADWORDS_CLIENT_USER_AGENT)
        return adwords_client.GetService('CustomerService', version=self.VERSION)

    def get_service(self, service_name, refresh_token, customer_acc_id):
        oauth_client = oauth2.GoogleRefreshTokenClient(self.CONFIG.client_id, self.CONFIG.client_secret, refresh_token)
        adwords_client = adwords.AdWordsClient(self.CONFIG.developer_token, oauth_client, ADWORDS_CLIENT_USER_AGENT)
        adwords_client.SetClientCustomerId(customer_acc_id)
        return adwords_client.GetService(service_name, version=self.VERSION)

    def get_report_downloader(self, refresh_token, customer_acc_id):
        oauth_client = oauth2.GoogleRefreshTokenClient(self.CONFIG.client_id, self.CONFIG.client_secret, refresh_token)
        adwords_client = adwords.AdWordsClient(self.CONFIG.developer_token, oauth_client, ADWORDS_CLIENT_USER_AGENT)
        adwords_client.SetClientCustomerId(customer_acc_id)
        return adwords_client.GetReportDownloader(version=self.VERSION)
