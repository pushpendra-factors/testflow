function info(message) {
    console.log("[Factors:INFO] : "+message);
}

function error(message) {
    console.error(message);
}

function debug(message) {
    if (process.env.NODE_ENV !== "production") {
        console.log("[Factors:DEBUG] : "+message)
    }
}

module.exports = exports = { info, error, debug };
