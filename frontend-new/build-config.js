const CONFIG = {
  development: {
    backend_host: 'http://factors-dev.com:8080',
    sdk_service_host: 'http://factors-dev.com:8085',
    adwords_service_host: 'http://factors-dev.com:8091',
    sdk_asset_url: 'http://localhost:8090/dist/factors.prod.js',
    android_sdk_asset_url: 'https://storage.googleapis.com/factors-staging-v2/sdk/android/sdk-staging-v0.1.aar',
    factors_sdk_token: 'dummy',
    facebook_app_id: '1022017331596075',
    linkedin_client_id: '861ix78kpo39ge',
  }, 
  staging: {
    backend_host: 'https://staging-api.factors.ai',
    sdk_asset_url: 'https://staging-app.factors.ai/assets/v1/factors.js',
    android_sdk_asset_url: 'https://storage.googleapis.com/factors-staging-v2/sdk/android/sdk-staging-v0.1.aar',
    factors_sdk_token: 'we0jyjxcs0ix4ggnkptymjh48ur8y7q7',
    facebook_app_id: '1022017331596075',
    linkedin_client_id: '861ix78kpo39ge',
  },
  test: {
    backend_host: 'http://localhost:8080',
    sdk_asset_url: 'http://localhost:8090/dist/factors.prod.js',
    android_sdk_asset_url: 'https://storage.googleapis.com/factors-staging-v2/sdk/android/sdk-staging-v0.1.aar',
    factors_sdk_token: 'dummy'
  },
  production: {
    backend_host: 'https://api.factors.ai',
    sdk_asset_url: 'https://app.factors.ai/assets/v1/factors.js',
    android_sdk_asset_url: 'https://storage.googleapis.com/factors-production-v2/sdk/android/sdk-production-v0.1.aar',
    factors_sdk_token: 'we0jyjxcs0ix4ggnkptymjh48ur8y7q7',
    facebook_app_id: '1022017331596075',
    linkedin_client_id: '861ix78kpo39ge',
  }
};

exports = module.exports = CONFIG;
