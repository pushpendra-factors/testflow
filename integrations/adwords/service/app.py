import json
import sys
import tornado.web
from tornado import gen
from tornado.log import logging as log
from tornado.ioloop import IOLoop
import google.oauth2.credentials
import google_auth_oauthlib.flow
from googleads import oauth2
from googleads import adwords
import urllib.parse
from base64 import b64encode, b64decode
import requests
from optparse import OptionParser

# commandline options
parser = OptionParser()

parser.add_option("--port", default=8091, help="Runs on the given port.")
parser.add_option("--host_url", default="http://localhost:8091", 
    help="Self host url with protocol to refer on callbacks.")
parser.add_option("--env", default="development", help="Environment.")
parser.add_option("--developer_token", default="", help="Adwords developer token.")
parser.add_option("--api_host_url", default="http://localhost:8080", help="API host url")
parser.add_option("--app_host_url", default="http://localhost:3000", help="App host url")
parser.add_option("--oauth_secret", default="", help="OAuth2 client secret JSON string")

SESSION_COOKIE_NAME = "factors-sid"
STATUS_FAILURE = "failure"
ADWORDS_CLIENT_USER_AGENT = "FactorsAI (https://www.factors.ai)"

class App():
    _env = ""
    _app_host_url = ""
    _api_host_url = ""
    _developer_token = ""
    
    @classmethod
    def init(cls, env, developer_token, api_host_url, app_host_url):        
        cls._env = env
        cls._developer_token = developer_token
        cls._api_host_url = api_host_url
        cls._app_host_url = app_host_url

    @classmethod
    def get_app_settings_redirect_url(cls, status=None):
        url = cls._app_host_url + "/#/settings/adwords"
        if status != None:
            url = url + "?status=" + status
        return url

    @classmethod
    def get_app_login_redirect_url(cls):
        return cls._app_host_url + "/#/login"

    @classmethod
    def get_app_host_url(cls):
        return cls._app_host_url

    @classmethod
    def get_api_host(cls):
        return cls._api_host_url

    @classmethod
    def get_developer_token(cls):
        return cls._developer_token

    @classmethod
    def get_session_cookie_name(cls):
        # Warning: Any changes to the cookie name has to be in sync with
        # App Server which is setting the cookie.
        if cls._env == "production": return SESSION_COOKIE_NAME
        if cls._env == "staging": return SESSION_COOKIE_NAME+"s"
        return SESSION_COOKIE_NAME+"d"

class OAuthManager():
    _redirect_url = None
    _secret = None
    _client_id = None
    _client_secret = None

    @classmethod
    def init(cls, redirect_url, secret):
        cls._redirect_url = redirect_url
        cls._secret = secret
        # throws KeyError.
        cls._client_id = secret["web"]["client_id"]
        cls._client_secret = secret["web"]["client_secret"]

    @classmethod
    def get_flow(cls):
        # Initialize the flow using the client ID and secret downloaded earlier.
        # Note: You can use the GetAPIScope helper function to retrieve the
        # appropriate scope for AdWords or Ad Manager.
        flow = google_auth_oauthlib.flow.Flow.from_client_config(
            cls._secret, scopes=[oauth2.GetAPIScope("adwords")])

        # Indicate where the API server will redirect the user after the user completes
        # the authorization flow. The redirect URI is required.
        flow.redirect_uri = cls._redirect_url

        return flow
        
    @classmethod
    def get_authorization_url(cls, state):
        if cls._secret == None or cls._redirect_url == None:
            log.error("OAuth manager is not initialized properly")
            return ""

        flow = cls.get_flow()

        # Generate URL for request to Google's OAuth 2.0 server.
        # Use kwargs to set optional request parameters.
        authorization_url, state = flow.authorization_url(
        # Enable offline access so that you can refresh an access token without
        # re-prompting the user for permission. Recommended for web server apps.
        access_type="offline",
        # Enable incremental authorization. Recommended as a best practice.
        include_granted_scopes="true",
        state=state)

        log.info("Generated authorization URL: %s", authorization_url)

        return authorization_url

    @classmethod
    def get_client_secret(cls):
        return OAuthManager._client_secret
    
    @classmethod
    def get_client_id(cls):
        return OAuthManager._client_id

class APIClientWithSession():
    @staticmethod
    def add_refresh_token(session, payload):
        if session == None or session == "":
            log.error("Invalid session cookie on add_refresh_token request.")
            return

        url = App.get_api_host() + "/integrations/adwords/add_refresh_token"
        
        cookies = {}
        cookies[App.get_session_cookie_name()] = session
        response = requests.post(url, json=payload, cookies=cookies)
        if not response.ok:
            log.error("Failed updating adwords integration with response : %d, %s", 
                response.status_code, response.json())
            return

        return response

    @staticmethod
    def get_adwords_refresh_token(session, project_id):
        url = App.get_api_host() + "/integrations/adwords/get_refresh_token"
        cookies = {}
        cookies[App.get_session_cookie_name()] = session
        # project_id as str for consistency on json.
        payload = { "project_id": str(project_id) }
        response = requests.post(url, json=payload, cookies=cookies)
        if not response.ok:
            log.error("Failed getting adwords integration with response : %d, %s", 
                response.status_code, response.json())
            return

        return response

class OAuthRedirectHandler(tornado.web.RequestHandler):
    @gen.coroutine
    def get(self):
        project_id = ""
        try:
            project_id = self.get_query_argument("pid")            
        except Exception as e:
            log.error("No project_id given on query param: %s", e)
            self.redirect(App.get_app_login_redirect_url(), True)
            return
        
        session_cookie_name = App.get_session_cookie_name()
        session_cookie_str = self.get_cookie(session_cookie_name)
        if session_cookie_str == None:
            log.error("Session %s cookie not found on oauth redirect handler.", session_cookie_name)
            self.redirect(App.get_app_login_redirect_url(), True)
            return

        # add project_id and session to state of auth url.
        session_cookie = urllib.parse.unquote(session_cookie_str)
        state = { "project_id": project_id, "session": session_cookie }
        state = b64encode(json.dumps(state).encode())

        self.redirect(OAuthManager.get_authorization_url(state), True)
        return

class OAuthCallbackHandler(tornado.web.RequestHandler):
    @gen.coroutine
    def get(self):
        code = ""
        try:
            code = self.get_argument("code")
        except tornado.web.MissingArgumentError:
            self.redirect(App.get_app_settings_redirect_url(STATUS_FAILURE), True)
            return
        except Exception as e:
            log.error("Query param code is not available on callback. %s", str(e))
            self.redirect(App.get_app_settings_redirect_url(STATUS_FAILURE), True)
            return

        state = ""
        try:
            state = self.get_argument("state")
        except tornado.web.MissingArgumentError:
            self.redirect(App.get_app_settings_redirect_url(STATUS_FAILURE), True)
            return
        except Exception as e:
            log.error("Query param state is not available on callback. %s", str(e))
            self.redirect(App.get_app_settings_redirect_url(STATUS_FAILURE), True)
            return
            
        flow = OAuthManager.get_flow()
        flow.fetch_token(code=code)

        if flow.credentials.refresh_token == None or flow.credentials.refresh_token == "":
            log.error("No refresh token on callback.")
            self.redirect(App.get_app_settings_redirect_url("ACCESS_TOKEN_FAILURE"), True)
            return

        state_payload = json.loads(b64decode(state).decode())

        project_id = state_payload.get("project_id")
        if project_id == None or project_id == "":
            log.error("Empty project_id from state of callback.")
            self.redirect(App.get_app_settings_redirect_url(STATUS_FAILURE), True)
            return
        
        # Create adwords integration request with session cookie.
        APIClientWithSession.add_refresh_token(
            state_payload.get("session"), 
            {
                "project_id": project_id, 
                "refresh_token": flow.credentials.refresh_token
            }
        )                
        
        self.redirect(App.get_app_settings_redirect_url(), True)
        return

class GetCustomerAccountsHandler(tornado.web.RequestHandler):
    # Todo: use set_default_headers and options as BaseHandler.
    def set_default_headers(self):
        self.set_header("Access-Control-Allow-Origin", App.get_app_host_url())
        self.set_header("Access-Control-Allow-Headers", "x-requested-with, Origin, Content-Type")
        self.set_header('Access-Control-Allow-Methods', 'POST, GET, PUT, DELETE, OPTIONS')
        self.set_header('Access-Control-Allow-Credentials', 'true')

    @gen.coroutine
    def options(self):
        self.set_status(200)
        self.finish()

    @gen.coroutine
    def post(self):
        params = json.loads(self.request.body.decode("utf-8"))
        if not "project_id" in params:
            self.set_status(400)
            self.finish({ "message": "invalid project_id" })
            return
        
        session_cookie_name = App.get_session_cookie_name()
        session_cookie_str = self.get_cookie(session_cookie_name)
        if session_cookie_str == None:
            log.error("Session %s cookie not found on request.", session_cookie_name)
            self.set_status(401)
            self.finish({ "message": "access unauthorized" })
            return

        session = urllib.parse.unquote(session_cookie_str)

        response = APIClientWithSession.get_adwords_refresh_token(session, params["project_id"])
        if response == None or not response.ok:
            log.error("Failure response on get_adwords_refresh_token for project "+ params["project_id"])
            self.set_status(500)
            self.finish({})
            return

        response_body = response.json()
        refresh_token = response_body.get("refresh_token")
        if refresh_token == None:
            log.error("refresh_token not found on response.")
            self.set_status(500)
            self.finish({})
            return

        # Get customer accounts.
        oauth_client = oauth2.GoogleRefreshTokenClient(OAuthManager.get_client_id(), 
            OAuthManager.get_client_secret(), refresh_token)
        adwords_client = adwords.AdWordsClient(App.get_developer_token(), oauth_client, ADWORDS_CLIENT_USER_AGENT)
        customer_service = adwords_client.GetService('CustomerService', version='v201809')
        
        response = []
        
        customer_accounts = customer_service.getCustomers()

        if len(customer_accounts) == 0:
            self.set_status(404)
            self.finish({"message": "no customer accounts found for user on adwords"})

        for account in customer_accounts:
            resp_account = {}

            # Manager account doesn't support reports download.
            # Skip listing it.
            try:
                if account["canManageClients"]:
                    log.warning("Skipping manager accounts on get customer accounts.")
                    continue
            except Exception:
                pass

            try:
                resp_account["customer_id"] = account["customerId"]
            except Exception:
                log.error("cusomter account id is missing on response from adwords")
                continue

            try:
                if account["descriptiveName"] != None:
                    resp_account["name"] = account["descriptiveName"]
                else:
                    resp_account["name"] = ""
            except KeyError:
                resp_account["name"] = ""
            except Exception:
                log.error("descriptive name is missing on response from adwords")
                continue

            response.append(resp_account)

        self.set_status(200)
        self.finish({ "customer_accounts": response })
        return

if __name__ == "__main__":
    (options, args) = parser.parse_args()

    oauth_secret_str = options.oauth_secret.strip()
    if oauth_secret_str == "":
        log.error("Option: oauth_secret cannot be empty.")
        sys.exit(1)

    try:
        # initialize client secret.
        oauth_client_secret = json.loads(oauth_secret_str)
    except Exception as e:
        log.error("Failed to load oauth_secret JSON: %s", str(e))
        sys.exit(1)

    if options.developer_token == "":
        log.error("Argument: --developer_token cannot be empty")
        sys.exit(1)
    
    App.init(options.env, options.developer_token, options.api_host_url, options.app_host_url)
    try:
        OAuthManager.init(options.host_url + "/adwords/auth/callback", oauth_client_secret)
    except Exception as e:
        log.error("Failed to init oauth manager with error %s", str(e))
        sys.exit(1)

    routes = [
        (r"/adwords/auth/redirect", OAuthRedirectHandler),
        (r"/adwords/auth/callback", OAuthCallbackHandler),
        (r"/adwords/get_customer_accounts", GetCustomerAccountsHandler)
    ]
    application = tornado.web.Application(routes)
    application.listen(options.port)
    log.warning("Listening on port %d..", options.port)
    IOLoop.instance().start()