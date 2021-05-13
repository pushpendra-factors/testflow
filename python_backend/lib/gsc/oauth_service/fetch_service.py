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
        authorization_url = "https://www.googleapis.com/oauth2/v4/token"
        r = requests.post(authorization_url, data=params)
        if r.ok:
                access_token = r.json()['access_token']
        else:
                return None
        credentials = AccessTokenCredentials(access_token=access_token, user_agent=GSC_CLIENT_USER_AGENT)
        http = httplib2.Http()
        http = credentials.authorize(http)
        return build('webmasters', 'v3', http=http)
    
