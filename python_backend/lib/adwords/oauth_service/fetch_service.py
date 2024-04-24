from googleads import oauth2

from google.ads.googleads.client import GoogleAdsClient

import time

ADWORDS_CLIENT_USER_AGENT = "FactorsAI (https://www.factors.ai)"


class FetchService:

    # New Version
    NEW_VERSION = "v15"

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
        
        return self.get_service_with_retries(credentials, service_name)
    
    def get_service_with_retries(self, credentials, service_name):
        ads_client = None
        service = None
        for retry in range(3):
            try:
                ads_client = GoogleAdsClient.load_from_dict(credentials)
                service = ads_client.get_service(service_name, version=self.NEW_VERSION)
                return service
            except Exception as e:
                if retry < 2:
                    time.sleep(10*(retry+1))
                else:
                    raise Exception(str(e)) 
        return service
