function info(message) {
    console.log(message);
}

function error(message) {
    console.error(message);
}

function debug(message) {
    if (process.env.NODE_ENV !== "production") {
        console.log(messag)
    }
}

export { info, error, debug };
