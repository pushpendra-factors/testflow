'use strict';

const CONFIG = {
    development: {
        api: {
            host: "http://localhost:8090"
        }
    },
    test: {
        api: {
            host: "http://localhost:8090"
        }
    },
    production: {
        api: {
            host: "https://factors.ai"
        }
    }
}

export default CONFIG[process.env.NODE_ENV];
