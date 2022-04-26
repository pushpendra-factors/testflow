import time
import httplib2
import requests

from googleapiclient.discovery import build
from oauth2client.client import AccessTokenCredentials
from google import oauth2
from google.oauth2.credentials import Credentials

GSC_CLIENT_USER_AGENT = "FactorsAI (https://www.factors.ai)"


class FetchService:

    def __init__(self, config):
        self.CONFIG = config
     
    def get_webmasters_service(self, refresh_token):
        params = {
                "grant_type": "refresh_token",
                "client_id": self.CONFIG.client_id,
                "client_secret": self.CONFIG.client_secret,
                "refresh_token": refresh_token
        }
        access_token = ''
        authorization_url = "https://www.googleapis.com/oauth2/v4/token"
        r = self.handle_request_with_retries(authorization_url, params)
        if r.ok:
                access_token = r.json()['access_token']
        credentials = AccessTokenCredentials(access_token=access_token, user_agent=GSC_CLIENT_USER_AGENT)
        http = httplib2.Http()
        http = credentials.authorize(http)
        return build('searchconsole', 'v1', http=http)
    
    def handle_request_with_retries(self, url, params):
        r = None
        for retry in range(3):
            r = requests.post(url, data=params)
            if r.status_code == 400 and r.json()['error'] == 'invalid_grant':
                    return r
            if r.status_code == 500:
                    time.sleep(2)
                    continue
            return r
        return r