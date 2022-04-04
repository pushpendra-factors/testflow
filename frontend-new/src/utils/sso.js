let host = BUILD_CONFIG.backend_host;

export const SSO_LOGIN_URL = host + '/oauth/login?connection=google-oauth2';
export const SSO_SIGNUP_URL = host + '/oauth/signup?connection=google-oauth2';
export const SSO_ACTIVATE_URL = host + '/oauth/activate?connection=google-oauth2';