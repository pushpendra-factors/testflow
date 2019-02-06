'use strict';

// It is possible override api.host on init using opts. e.g factors.init("token", {host: "https://example.com"})

const CONFIG = {
    development: {
        api: {
            host: "http://localhost:8080"
        }
    },
    test: {
        api: {
            host: "http://localhost:8080"
        }
    },
    production: {
        api: {
            host: "https://api.factors.ai"
        }
    }
}

module.exports = exports = CONFIG[process.env.NODE_ENV];
