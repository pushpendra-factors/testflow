### Setup

* Install python3
```
brew install python3
```

* Install dependencies
```
cd <path_to_factors>/integrations/service
pip install -r requirements.txt
```

### Configure OAuth client
1. Visit https://console.developers.google.com/apis/credentials
2. Change to appropriate project on the top dropdown. For development use 'My First Project'.
3. Click 'Create Credentials' > 'OAuth Client ID'. 
4. Choose 'Web application'. Add Name.
5. Add 'Authorized JavaScript origins' as below.
```
# development
http://factors-dev.com:8091

# staging
https://staging-api.factors.ai

# production
https://api.factors.ai
```
6. Add 'Authorized redirect URIs' as `<AUTHORIZED_ORIGIN>/adwords/auth/callback`.
7. Click create.
8. Click download symbol near your credential on the credentials page.

### Start
```
python auth_service.py --oauth_secret=$(cat <CLIENT_SECRET_FILE>)
```




