import json
from base64 import b64encode

from google_auth_oauthlib.flow import Flow
from googleads import oauth2
from tornado.log import logging as log

from lib.adwords.config.oauth_config import OauthConfig as AdwordsOauthConfig
from lib.data_services.factors_data_service import FactorsDataService
from lib.exception.error_message_constant import MISSING_ARGUMENT_IN_JSON
from lib.exception.oauth_callback_exceptions import OauthCallbackMissingParameter, OauthCallbackMissingRefreshToken
from lib.oauth_service import OAuthService as BaseOauthService

ADWORDS_CLIENT_USER_AGENT = "FactorsAI (https://www.factors.ai)"


class AuthorisationFlow(BaseOauthService):
    VERSION = "v201809"

    # GETAPIScope is giving wrong value. Hence directly entering scope value.
    def initialise_flow(self):
        flow = Flow.from_client_config(self.CONFIG.secret_json, scopes=["https://www.googleapis.com/auth/adwords"])
        flow.redirect_uri = self.CONFIG.callback_url
        return flow

    def __init__(self, config=AdwordsOauthConfig):
        self.CONFIG = config
        self.flow = self.initialise_flow()

    def get_authorization_url(self, project_id, agent_uuid, session_cookie):
        state = {"project_id": project_id, "agent_uuid": agent_uuid, "session": session_cookie}
        state = b64encode(json.dumps(state).encode())

        authorization_url, state = self.flow.authorization_url(
            access_type="offline",
            scopes=["https://www.googleapis.com/auth/adwords"],
            state=state)

        log.info("Generated authorization URL: %s", authorization_url)
        return authorization_url

    def check_and_add_refresh_token(self, authorisation_code, state_payload):

        project_id = state_payload.get("project_id")
        if project_id is None or project_id == "":
            log.error(MISSING_ARGUMENT_IN_JSON.format(argument="state", hash_name="auth callback: State"))
            raise OauthCallbackMissingParameter()

        agent_uuid = state_payload.get("agent_uuid")
        if agent_uuid is None or agent_uuid == "":
            log.error(MISSING_ARGUMENT_IN_JSON.format(argument="agent_uuid", hash_name="auth callback: State"))
            raise OauthCallbackMissingParameter("")

        session = state_payload.get("session")
        if session is None or session == "":
            log.error(MISSING_ARGUMENT_IN_JSON.format(argument="session", hash_name="auth callback: State"))
            raise OauthCallbackMissingParameter("")

        try:
            self.flow.fetch_token(code=authorisation_code)
        except Exception as e:
            log.error("Failed to fetch token on callback. %s", str(e))
            return

        if self.flow.credentials.refresh_token is None or self.flow.credentials.refresh_token == "":
            log.error("No refresh token on callback.")
            raise OauthCallbackMissingRefreshToken()

        FactorsDataService.add_refresh_token(
            session,
            {
                "project_id": project_id,
                "agent_uuid": agent_uuid,
                "refresh_token": self.flow.credentials.refresh_token
            }
        )
        return
