"use strict";

var App = require("./app");

// Global reference.
var app = new App();

/**
 * Initializes sdk environment on user application. Overwrites if initialized already.
 * @param {string} appToken Unique application token.
 * @param {object} opts Additional opts: {track_on_init: false}
 */
function init(appToken, opts={}) {
    try {
        return app.init(appToken, opts)
            .catch(app.handleError);
    } catch(e) {
        return app.handleError(e);
    }
}

/**
 * Clears existing SDK environment, both API token and cookies. 
 */
function reset() {
    try {
        app.reset();
    } catch(e) {
        app.handleError(e);
    }

    return;
}

/**
 * Track events on user application.
 * @param {string} eventName
 * @param {Object} eventProperties 
 */
function track(eventName, eventProperties={}) {
    try {
        app.track(eventName, eventProperties, false)
            .catch(app.handleError);
    } catch(e) {
        app.handleError(e);
    }

    return;
}
/**
 * Track visit to page as event.
 */
function page() {
    try {
        app.page().catch(app.handleError);
    } catch(e) {
        app.handleError(e);
    }

    return;
}

/**
 * Identify user with original userId from the application.
 * @param {string} customerUserId Actual id of the user from the application.
 */
function identify(customerUserId) {
    try {
        app.identify(customerUserId)
            .catch(app.handleError);
    } catch(e) {
        app.handleError(e);
    }

    return;
}

/**
 * Add additional user properties.
 * @param {Object} properties 
 */
function addUserProperties(properties={}) {
    try {
        app.addUserProperties(properties)
            .catch(app.handleError);
    } catch(e) {
        app.handleError(e);
    }

    return;
}

function getUserId() {
    return app.getUserId();
}

let exposed = { init, reset, track, page, identify, addUserProperties, getUserId };
if (process.env.NODE_ENV === "development") {
    exposed["test"] = require("./test/suite.js");
    exposed["app"] = app;
}
module.exports = exports = exposed;

