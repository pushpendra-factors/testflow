const CONFIG = {
  development: {
    backend_host: 'http://factors-dev.com:8080',
    sdk_asset_url: 'http://localhost:8090/dist/factors.prod.js',
    factors_sdk_token: 'dummy'
  },
  staging: {
    backend_host: 'https://staging-api.factors.ai',
    sdk_asset_url: 'https://staging-app.factors.ai/assets/factors.js',
    factors_sdk_token: 'we0jyjxcs0ix4ggnkptymjh48ur8y7q7'
  },
  test: {
    backend_host: 'http://localhost:8080',
    sdk_asset_url: 'http://localhost:8090/dist/factors.prod.js',
    factors_sdk_token: 'dummy'
  },
  production: {
    backend_host: 'https://api.factors.ai',
    sdk_asset_url: 'https://app.factors.ai/assets/factors.js',
    factors_sdk_token: 'we0jyjxcs0ix4ggnkptymjh48ur8y7q7'
  }
}

exports = module.exports = CONFIG;
