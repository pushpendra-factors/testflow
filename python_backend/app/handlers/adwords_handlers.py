import json
import urllib.parse
from base64 import b64decode

from tornado import gen
from tornado.log import logging as log

import app
from lib.adwords.cors import Cors
from lib.adwords.oauth_service.authorisation_flow import AuthorisationFlow as AdwordsOauthService
from lib.adwords.oauth_service.fetch_service import FetchService
from lib.data_services.factors_data_service import FactorsDataService
from lib.exception.error_message_constant import MISSING_ARGUMENT_ERROR
from lib.exception.oauth_callback_exceptions import OauthCallbackMissingParameter, OauthCallbackMissingRefreshToken
from .base_handler import BaseHandler

STATUS_FAILURE = "failure"
REFRESH_TOKEN_FAILURE = "refresh token failure"


class DefaultHandler(BaseHandler):
    @gen.coroutine
    def get(self):
        self.set_status(200)
        self.finish({"status": "I'm ok."})
        return

# TODO Delete all deprecated methods and handlers.
class OAuthRedirectHandler(BaseHandler):

    @gen.coroutine
    def get(self):
        try:
            project_id = self.get_query_argument("pid")
        except Exception as e:
            log.error(MISSING_ARGUMENT_ERROR.format(argument="project_id", exception=e))
            self.redirect(app.CONFIG.ADWORDS_APP.get_factors_login_redirect_url(), True)
            return

        try:
            agent_uuid = self.get_query_argument("aid")
        except Exception as e:
            log.error(MISSING_ARGUMENT_ERROR.format(argument="agent_uuid", exception=e))
            self.redirect(app.CONFIG.ADWORDS_APP.get_factors_login_redirect_url(), True)
            return

        session_cookie_key = app.CONFIG.ADWORDS_APP.get_session_cookie_key()
        session_cookie_value = self.get_cookie(session_cookie_key)
        if session_cookie_value is None:
            log.error("Session %s cookie not found on oauth redirect handler.", session_cookie_key)
            self.redirect(app.CONFIG.ADWORDS_APP.get_factors_login_redirect_url(), True)
            return

        session_cookie = urllib.parse.unquote(session_cookie_value)
        self.redirect(AdwordsOauthService().get_authorization_url(project_id, agent_uuid, session_cookie), True)
        return


class OAuthCallbackHandler(BaseHandler):
    @gen.coroutine
    def get(self):
        # Input params validation.
        try:
            authorisation_code = self.get_argument("code")
        except Exception as e:
            log.error(MISSING_ARGUMENT_ERROR.format(argment="code", exception=e))
            self.redirect(app.CONFIG.ADWORDS_APP.get_factors_admin_adwords_redirect_url(STATUS_FAILURE), True)
            return

        try:
            state = self.get_argument("state")
        except Exception as e:
            log.error(MISSING_ARGUMENT_ERROR.format(argument="state", exception=e))
            self.redirect(app.CONFIG.ADWORDS_APP.get_factors_admin_adwords_redirect_url(STATUS_FAILURE), True)
            return

        state_payload = json.loads(b64decode(state).decode())
        try:
            AdwordsOauthService().check_and_add_refresh_token(authorisation_code, state_payload)
        except OauthCallbackMissingParameter as e:
            self.redirect(app.CONFIG.ADWORDS_APP.get_factors_admin_adwords_redirect_url(STATUS_FAILURE), True)
            return

        except OauthCallbackMissingRefreshToken as e:
            self.redirect(app.CONFIG.ADWORDS_APP.get_factors_admin_adwords_redirect_url(REFRESH_TOKEN_FAILURE), True)
            return

        self.redirect(app.CONFIG.ADWORDS_APP.get_factors_admin_adwords_redirect_url(), True)
        return

