import json
import sys

from tornado.log import logging as log

from lib.config import Config

CALLBACK_PATH = "/adwords/auth/callback"


class OauthConfig(Config):
    callback_url = None
    secret_json = None
    client_id = None
    client_secret = None
    developer_token = None

    @classmethod
    def build(cls, argv):
        cls._validate(argv)
        oauth_secret_json = json.loads(argv.oauth_secret.strip())
        cls._init(argv.host_url + CALLBACK_PATH, oauth_secret_json, argv.developer_token)

    # TODO
    @staticmethod
    def _validate(argv):
        oauth_secret_str = argv.oauth_secret.strip()
        if oauth_secret_str == "":
            log.error("Option: oauth_secret cannot be empty.")
            sys.exit(1)

        try:
            oauth_secret_json = json.loads(oauth_secret_str)
        except Exception as e:
            log.error("Failed to load oauth_secret JSON: %s", str(e))
            sys.exit(1)

        is_client_id_present = ("web" in oauth_secret_json) and ("client_id" in oauth_secret_json["web"]) and (
                    oauth_secret_json["web"]["client_id"] is not None)
        if not is_client_id_present:
            log.error("Client Id is not given in the json:")
            sys.exit(1)

        if argv.developer_token == "":
            log.error("Argument: --developer_token cannot be empty")
            sys.exit(1)

    @classmethod
    def _init(cls, callback_url, oauth_secret_json, developer_token):
        cls.callback_url = callback_url
        cls.secret_json = oauth_secret_json
        cls.client_id = oauth_secret_json["web"]["client_id"]
        cls.client_secret = oauth_secret_json["web"]["client_secret"]
        cls.developer_token = developer_token
        return
