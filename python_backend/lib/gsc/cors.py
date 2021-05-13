from scripts.gsc import DEVELOPMENT, TEST, STAGING, PRODUCTION


class Cors:
    acceptable_origins = []

    @classmethod
    def get_cors_allowed_origin(cls, request_origin):
        if request_origin is None or len(request_origin) == 0:
            return
        for curr_origin in cls.acceptable_origins:
            if request_origin == curr_origin: return request_origin
        return

    @classmethod
    def set_acceptable_origins(cls, env):
        result_domains = []
        if env in (DEVELOPMENT, TEST):
            result_domains = ["http://factors-dev.com:3000"]
        elif env == STAGING:
            result_domains = [
                "https://staging-app.factors.ai",
                "https://tufte-staging.factors.ai",
                "https://staging-app-old.factors.ai",
                "http://factors-dev.com:3000"
            ]
        elif env == PRODUCTION:
            result_domains = [
                "https://app.factors.ai",
                "https://tufte-prod.factors.ai",
                "https://app-old.factors.ai"
            ]
        cls.acceptable_origins = result_domains
        return
