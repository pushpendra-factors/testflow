from app.handlers import adwords_handlers

ROUTES = [
    (r"/adwords/auth/redirect", adwords_handlers.OAuthRedirectHandler),
    (r"/adwords/auth/callback", adwords_handlers.OAuthCallbackHandler),
    (r"/adwords/get_customer_accounts", adwords_handlers.GetCustomerAccountsHandler)
]
