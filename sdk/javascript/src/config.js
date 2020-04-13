'use strict';

// It is possible override api.host on init using opts. e.g factors.init("token", {host: "https://example.com"})

const CONFIG = {
    development: {
        api: {
            host: "http://localhost:8085"
        }
    },
    test: {
        api: {
            host: "http://localhost:8085"
        }
    },
    production: {
        api: {
            host: "https://api.factors.ai"
        }
    }
}

module.exports = exports = CONFIG[process.env.NODE_ENV];
