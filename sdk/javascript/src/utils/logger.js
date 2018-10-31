function info(message) {
    console.log("[Factors:INFO] : "+message);
}

function error(message) {
    console.error("[Factors:ERROR] : "+message);
}

function debug(message) {
    if (process.env.NODE_ENV !== "production") {
        console.log("[Factors:DEBUG] : "+message)
    }
}

export { info, error, debug };
