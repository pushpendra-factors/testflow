'use strict';

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
    staging: {
        api: {
            host: "" // uses host from init factors.init("token", {host: "api_host"})
        }
    },
    production: {
        api: {
            host: "http://app.factors.ai"
        }
    }
}

module.exports = exports = CONFIG[process.env.NODE_ENV];
