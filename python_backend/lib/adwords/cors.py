from scripts.adwords import DEVELOPMENT, TEST, STAGING, PRODUCTION


class Cors:

    @staticmethod
    def get_acceptable_domains(env):
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
        return (", ").join(result_domains)
