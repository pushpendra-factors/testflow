"use strict";

var App = require("./app");
const logger = require("./utils/logger");



// Global reference.
var app = new App();

/**
 * Initializes sdk environment on user application. Overwrites if initialized already.
 * @param {string} appToken Unique application token.
 */
function init(appToken, opts={}) {
    app.init(appToken, opts)
        .catch(logger.errorLine);
    return;
}

/**
 * Clears existing SDK environment, both API token and cookies. 
 */
function reset() { 
    app.reset();
    return;
}

/**
 * Track events on user application.
 * @param {string} eventName
 * @param {Object} eventProperties 
 */
function track(eventName, eventProperties={}) {
    app.track(eventName, eventProperties, false)
        .catch(logger.errorLine);
    return;
}
/**
 * Track visit to page as event.
 */
function page() {
    app.page().catch(logger.errorLine);
    return;
}

/**
 * Identify user with original userId from the application.
 * @param {string} customerUserId Actual id of the user from the application.
 */
function identify(customerUserId) {
    app.identify(customerUserId)
        .catch(logger.errorLine);
    return;
}

/**
 * Add additional user properties.
 * @param {Object} properties 
 */
function addUserProperties(properties={}) {
    app.addUserProperties(properties)
        .catch(logger.errorLine);
    return;
}

let exposed = { init, reset, track, page, identify, addUserProperties };
if (process.env.NODE_ENV === "development") {
    exposed["test"] = require("./test/suite.js");
    exposed["app"] = app;
}
module.exports = exports = exposed;

