const CONFIG = {
  development: {
    backend_host: 'http://factors-dev.com:8080',
    sdk_service_host: 'http://factors-dev.com:8085',
    adwords_service_host: 'http://factors-dev.com:8091',
    sdk_asset_url: 'http://localhost:8090/dist/factors.prod.js',
    android_sdk_asset_url:
      'https://storage.googleapis.com/factors-staging-v2/sdk/android/sdk-staging-v0.1.aar',
    factors_sdk_token: 'dummy',
    facebook_app_id: '1022017331596075',
    linkedin_client_id: '861ix78kpo39ge',
    firstTimeDashboardTemplates: {
      webanalytics: '3de1776a-2b06-4223-adf3-a57012833ec5',
      websitevisitoridentification: 'd14c52fd-7856-4fda-9f2e-0f36c2dd0084',
      allpaidmarketing: '215d866d-129c-415a-a728-592672604cfa'
    }
  },
  staging: {
    backend_host: 'https://staging-api.factors.ai',
    sdk_asset_url: 'https://staging-app.factors.ai/assets/v1/factors.js',
    android_sdk_asset_url:
      'https://storage.googleapis.com/factors-staging-v2/sdk/android/sdk-staging-v0.1.aar',
    factors_sdk_token: 'we0jyjxcs0ix4ggnkptymjh48ur8y7q7',
    facebook_app_id: '1022017331596075',
    linkedin_client_id: '861ix78kpo39ge',
    firstTimeDashboardTemplates: {
      webanalytics: '3de1776a-2b06-4223-adf3-a57012833ec5',
      websitevisitoridentification: 'd14c52fd-7856-4fda-9f2e-0f36c2dd0084',
      allpaidmarketing: 'bf10934c-0128-4c1c-8948-772387c95502'
    }
  },
  test: {
    backend_host: 'http://localhost:8080',
    sdk_asset_url: 'http://localhost:8090/dist/factors.prod.js',
    android_sdk_asset_url:
      'https://storage.googleapis.com/factors-staging-v2/sdk/android/sdk-staging-v0.1.aar',
    factors_sdk_token: 'dummy'
  },
  production: {
    backend_host: 'https://api.factors.ai',
    sdk_asset_url: 'https://app.factors.ai/assets/v1/factors.js',
    android_sdk_asset_url:
      'https://storage.googleapis.com/factors-production-v2/sdk/android/sdk-production-v0.1.aar',
    factors_sdk_token: 'we0jyjxcs0ix4ggnkptymjh48ur8y7q7',
    facebook_app_id: '1022017331596075',
    linkedin_client_id: '861ix78kpo39ge',
    firstTimeDashboardTemplates: {
      webanalytics: '6d966a20-07a6-46db-b2d4-1b1d4d7fb8ec',
      websitevisitoridentification: 'b60283bc-ade3-4c12-afd1-f8a85618df91',
      allpaidmarketing: 'f8785bee-6d7d-4a03-8744-a2d31ce4dd9e'
    }
  }
};

exports = module.exports = CONFIG;
