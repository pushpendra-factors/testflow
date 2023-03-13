from app.handlers import adwords_handlers, v1_adwords_handlers, v1_gsc_handlers

ROUTES = [
    (r"/", adwords_handlers.DefaultHandler),
    (r"/adwords/auth/redirect", adwords_handlers.OAuthRedirectHandler),
    (r"/adwords/auth/callback", adwords_handlers.OAuthCallbackHandler),

    (r"/adwords/v1/auth/redirect", v1_adwords_handlers.OAuthRedirectV1Handler),
    (r"/adwords/v1/auth/callback", v1_adwords_handlers.OAuthCallbackV1Handler),
    (r"/adwords/v1/get_customer_accounts", v1_adwords_handlers.GetCustomerAccountsV1Handler),

    (r"/google_organic/v1/auth/redirect", v1_gsc_handlers.OAuthRedirectV1Handler),
    (r"/google_organic/v1/auth/callback", v1_gsc_handlers.OAuthCallbackV1Handler),
    (r"/google_organic/v1/get_google_organic_urls", v1_gsc_handlers.GetGSCURLsV1Handler),
]
