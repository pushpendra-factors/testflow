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
    production: {
        api: {
            host: "http://localhost:8080"
        }
    }
}

module.exports = exports = CONFIG[process.env.NODE_ENV];
