function info(message) {
    console.log(message);
}

function error(message) {
    console.error(message);
}

function debug(message) {
    if (window.FACTORS_DEBUG == true) {
        console.trace("%c"+message, 'color: red');
    }
}

module.exports = exports = { info, error, debug };
