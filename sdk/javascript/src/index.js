"use strict";

var App = require("./app");

// Global reference.
var app = new App();

/**
 * Initializes sdk environment on user application. Overwrites if initialized already.
 * 
 * @param {string} appToken Unique application token.
 */
function init(appToken) {
    return app.init(appToken);
}

/**
 * Clears existing SDK environment, both API token and cookies. 
 */
function reset() { app.reset(); }

/**
 * Track events on user application.
 * @param {string} eventName
 * @param {Object} eventProperties 
 */
function track(eventName, eventProperties={}) {
    return app.track(eventName, eventProperties, false);
}

/**
 * Identify user with original userId from the application.
 * @param {string} customerUserId Actual id of the user from the application.
 */
function identify(customerUserId) {
    return app.identify(customerUserId);
}

/**
 * Add additional user properties.
 * @param {Object} properties 
 */
function addUserProperties(properties={}) {
    return app.addUserProperties(properties);
}

let exposed = { init, reset, track, identify, addUserProperties };
if (process.env.NODE_ENV === "development") {
    exposed["test"] = require("./test/suite.js");
    exposed["app"] = app;
}
module.exports = exports = exposed;

