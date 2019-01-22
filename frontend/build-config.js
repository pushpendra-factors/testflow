const CONFIG = {
  development: {
    backend_host: 'http://localhost:8080',
    sdk_asset_url: 'http://localhost:8090/dist/factors.prod.js'
  },
  staging: {
    backend_host: '', // Uses window.origin.
    sdk_asset_url: 'https://slashbit.co/assets/factors.js'
  },
  test: {
    backend_host: 'http://localhost:8080',
    sdk_asset_url: 'http://localhost:8090/dist/factors.prod.js'
  },
  production: {
    backend_host: 'https://factors.ai',
    // Use factors.ai for production for rotating asset whenever required.
    sdk_asset_url: 'https://factors.ai/assets/factors.js'
  }
}

exports = module.exports = CONFIG;
