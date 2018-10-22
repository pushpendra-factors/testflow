var logger = require('./logger');

var isInstalled = function() {
    return "Factors sdk v0.1 is installed!";
}

module.exports = {
    logger: logger,
    isInstalled: isInstalled
};