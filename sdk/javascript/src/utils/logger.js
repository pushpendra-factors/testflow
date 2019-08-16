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

function errorLine(err) {
    let line = '';
    if (typeof(err) == "string") line = err;
    if (err instanceof Error && err.message) line = err.message;
    if (line != '') console.error(line);
}

module.exports = exports = { info, error, debug, errorLine };
