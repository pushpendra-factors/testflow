from googleads import oauth2

from google.ads.googleads.client import GoogleAdsClient

ADWORDS_CLIENT_USER_AGENT = "FactorsAI (https://www.factors.ai)"


class FetchService:

    # New Version
    NEW_VERSION = "v13"

    # Old Version
    VERSION = "v201809"

    def __init__(self, config):
        self.CONFIG = config

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
