from googleads import oauth2, adwords

from google.ads.googleads.client import GoogleAdsClient

ADWORDS_CLIENT_USER_AGENT = "FactorsAI (https://www.factors.ai)"


class FetchService:

    # New Version
    NEW_VERSION = "v10"


    # Old Version
    VERSION = "v201809"

    def __init__(self, config):
        self.CONFIG = config

    def get_customer_accounts(self, refresh_token):
        oauth_client = oauth2.GoogleRefreshTokenClient(self.CONFIG.client_id, self.CONFIG.client_secret, refresh_token)
        adwords_client = adwords.AdWordsClient(self.CONFIG.developer_token, oauth_client, ADWORDS_CLIENT_USER_AGENT)
        return adwords_client.GetService("CustomerService", version=self.VERSION)

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

    def new_get_service(self, service_name, refresh_token, login_customer_id=None):
        # GoogleAdsClient will read the credentials dictionary
        credentials = {
            "developer_token": self.CONFIG.developer_token,
            "refresh_token": refresh_token,
            "client_id": self.CONFIG.client_id,
            "client_secret": self.CONFIG.client_secret,
            "use_proto_plus": True
        }   
        if login_customer_id != None and login_customer_id != '':
            credentials["login_customer_id"] = login_customer_id
        ads_client = GoogleAdsClient.load_from_dict(credentials)
        return ads_client.get_service(service_name, version=self.NEW_VERSION)