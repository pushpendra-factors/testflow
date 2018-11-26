const CONFIG = {
    development: {
       sdk_asset_url: 'http://localhost:8090/dist/factors.prod.js'
    },
    test: {
        sdk_asset_url: 'http://localhost:8090/dist/factors.prod.js'
    },
    production: {
        // Use factors.ai for production for rotating asset whenever required.
        sdk_asset_url: 'https://factors.ai/assets/factors.prod.js'
    }
}

exports = module.exports = CONFIG;
