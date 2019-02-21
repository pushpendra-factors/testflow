const CONFIG = {
  development: {
    backend_host: 'http://factors-dev.com:8080',
    sdk_asset_url: 'http://localhost:8090/dist/factors.prod.js'
  },
  staging: {
    backend_host: 'https://api.factors.ai',
    // Uses window.origin/assets/factors.js. Todo: Replace with https://app.factors.ai/assets/factors.js
    sdk_asset_url: ''
  },
  test: {
    backend_host: 'http://localhost:8080',
    sdk_asset_url: 'http://localhost:8090/dist/factors.prod.js'
  },
  production: {
    backend_host: 'https://api.factors.ai',
    // Use factors.ai for production for rotating asset whenever required.
    sdk_asset_url: 'https://factors.ai/assets/factors.js'
  }
}

exports = module.exports = CONFIG;
