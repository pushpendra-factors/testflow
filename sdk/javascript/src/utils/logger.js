function info(message) {
    console.log(message);
}

function error(message) {
    console.error(message);
}

function debug(message, trace=false) {
    if (window.FAITRACKER_DEBUG == true) {
        if (trace) console.trace("%c"+message, 'color: red');
        else console.log(message);
    }
}

function errorLine(err) {
    let line = '';
    if (typeof(err) == "string") line = err;
    if (err instanceof Error && err.message) line = err.message;
    if (line != '') console.error(line);
}

module.exports = exports = { info, error, debug, errorLine };
