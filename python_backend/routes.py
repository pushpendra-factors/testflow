from app.handlers import adwords_handlers, v1_adwords_handlers

ROUTES = [
    (r"/adwords/auth/redirect", adwords_handlers.OAuthRedirectHandler),
    (r"/adwords/auth/callback", adwords_handlers.OAuthCallbackHandler),
    (r"/adwords/get_customer_accounts", adwords_handlers.GetCustomerAccountsHandler),

    (r"/adwords/v1/auth/redirect", v1_adwords_handlers.OAuthRedirectV1Handler),
    (r"/adwords/v1/auth/callback", v1_adwords_handlers.OAuthCallbackV1Handler),
    (r"/adwords/v1/get_customer_accounts", v1_adwords_handlers.GetCustomerAccountsV1Handler)
]
