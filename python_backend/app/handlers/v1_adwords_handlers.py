import json
import urllib.parse
from base64 import b64decode
import traceback

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


class OAuthRedirectV1Handler(BaseHandler):

    @gen.coroutine
    def get(self):
        try:
            project_id = self.get_query_argument("pid")
        except Exception as e:
            log.error(MISSING_ARGUMENT_ERROR.format(argument="project_id", exception=e))
            self.finish(json.dumps({'error': {'code': 400, 'message': "project_id is not provided in params."}}))
            return

        try:
            agent_uuid = self.get_query_argument("aid")
        except Exception as e:
            log.error(MISSING_ARGUMENT_ERROR.format(argument="agent_uuid", exception=e))
            self.finish(json.dumps({'error': {'code': 400, 'message': "agent_id is not provided in params."}}))
            return

        session_cookie_key = app.CONFIG.ADWORDS_APP.get_session_cookie_key()
        session_cookie_value = self.get_cookie(session_cookie_key)
        if session_cookie_value is None:
            log.error("Session %s cookie not found on oauth redirect handler.", session_cookie_key)
            self.finish(json.dumps({'error': {'code': 400, 'message': "session is not provided in params."}}))
            return

        session_cookie = urllib.parse.unquote(session_cookie_value)
        self.write({"url": AdwordsOauthService().get_authorization_url(project_id, agent_uuid, session_cookie)})
        return


class OAuthCallbackV1Handler(BaseHandler):

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


class GetCustomerAccountsV1Handler(BaseHandler):

    @gen.coroutine
    def options(self):
        self.set_status(200)
        self.finish()

    @gen.coroutine
    def post(self):
        try:
            params = json.loads(self.request.body.decode("utf-8"))
            project_id = params["project_id"]
        except Exception as e:
            self.set_status(400)
            self.finish({"message": MISSING_ARGUMENT_ERROR.format(argument="project_id", exception=e)})
            return

        session_cookie_key = app.CONFIG.ADWORDS_APP.get_session_cookie_key()
        session_cookie_value = self.get_cookie(session_cookie_key)
        if session_cookie_value is None:
            log.error("Session %s cookie not found on request.", session_cookie_key)
            self.set_status(401)
            self.finish({"message": "access unauthorized"})
            return

        # TODO Imp - Check Session is not used.
        session = urllib.parse.unquote(session_cookie_value)

        response = FactorsDataService.get_adwords_refresh_token(project_id)
        if response is None or not response.ok:
            log.error("Failure response on get_adwords_refresh_token for project: " + project_id)
            self.set_status(500)
            self.finish({})
            return

        response_body = response.json()
        refresh_token = response_body.get("refresh_token")
        if refresh_token is None:
            log.error("refresh_token not found on response.")
            self.set_status(500)
            self.finish({})
            return

        # Get customer accounts.
        seed_customer_ids = []
        final_customer_ids = []

        # manager(bool): returns if query customer id is manager or not. level 0 signifies directly related accounts of logged in user i.e no child/sub-accounts included
        query = """
            SELECT
            customer_client.manager,
            customer_client.id,
            customer_client.descriptive_name
            FROM customer_client
            WHERE customer_client.level = 0"""

        try:
            customer_service = FetchService(app.CONFIG.ADWORDS_OAUTH).new_get_service('CustomerService',refresh_token)
            googleads_service = FetchService(app.CONFIG.ADWORDS_OAUTH).new_get_service('GoogleAdsService',refresh_token) 
            customer_resource_names = (
                customer_service.list_accessible_customers().resource_names
            )
        except Exception as e:
            self.set_status(400)
            traceback.print_tb(e.__traceback__)
            log.warning("Errored during customer fetch from adwords" + str(e))
            self.finish({"message":"Error happened during fetch of customers from adwords."})
            return

        for customer_resource_name in customer_resource_names:
            customer_id = googleads_service.parse_customer_path(
                customer_resource_name
            )["customer_id"]
            seed_customer_ids.append(customer_id)

        for seed_customer_id in seed_customer_ids:
            try:
                response = googleads_service.search(
                        customer_id=str(seed_customer_id), query=query
                    )
            except Exception as e:
                traceback.print_tb(e.__traceback__)
                log.warning("Error during manager status fetch " + str(e))
                continue
            for row in response:
                customer_client = row.customer_client
                if not customer_client.manager:
                     final_customer_ids.append({"customer_id": customer_client.id, "descriptiveName": customer_client.descriptive_name})
                else:
                    log.warning("Skipping manager account")

        self.set_status(200)
        self.finish({"customer_accounts": final_customer_ids})
        return
