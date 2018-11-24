const CONFIG = {
    development: {
       sdk_asset_url: 'http://localhost:8090/dist/factors.prod.js'
    },
    test: {
        sdk_asset_url: 'http://localhost:8090/dist/factors.prod.js'
    },
    production: {
        sdk_asset_url: 'https://factors.ai/assets/factors.prod.js'
    }
}

exports = module.exports = CONFIG;
