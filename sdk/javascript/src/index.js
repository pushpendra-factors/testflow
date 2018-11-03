"use strict";

var Cookie = require("./utils/cookie");
var logger = require("./utils/logger");

var App = require("./app");

// Constants.
const COOKIE_FUID = "_fuid";

// Common methods.

function _updateCookieIfUserIdInResponse(response){
    if (response && response.body && response.body.user_id) {
        let cleanUserId = response.body.user_id.trim();

        if (cleanUserId) Cookie.set(COOKIE_FUID, cleanUserId);
    }
    return response; // To continue chaining.
}

function _validatedStringArg(name, value) {
    if (typeof(value) != "string")
        throw new Error("FactorsError: Invalid type for "+name);
    
    value = value.trim();
    if (!value) throw new Error("FactorsError: "+name+" cannot be empty.");
    
    return value;
}

// Exposed methods.

// Global reference.
var app = new App(null, {});

/**
 * Prints SDK information, if installed.
 */
function isInstalled() {
    return "Factors sdk v0.1 is installed!";
}

/**
 * Initializes sdk environment on user application. Overwrites if initialized already.
 * @param {string} appToken Unique application token.
 * @param {Object} appConfig Custom application configuration. i.e., { autoTrackPageView: true }
 */
function init(appToken, appConfig) {
    appToken = appToken.trim();
    if(!appToken) throw new Error("FactorsError: Initialization failed. Invalid Token.");

    app.set(appToken, appConfig);
}

function reset() {
    app.reset();
    Cookie.remove(COOKIE_FUID);
}

/**
 * Track events on user application.
 * @param {string} eventName
 * @param {Object} eventProperties 
 */
function track(eventName, eventProperties={}) {
    eventName = _validatedStringArg("event_name", eventName)

    let payload = {};

    // Use user_id on cookie.
    if (Cookie.isExist(COOKIE_FUID)) 
        payload.user_id = Cookie.get(COOKIE_FUID);
    
    payload.event_name = eventName;
    payload.event_properties = eventProperties;

    if (app && app.client.isInitialized()) {
        return app.client.track(payload)
            .then(_updateCookieIfUserIdInResponse)
            .catch(logger.error);
    } else {
        throw new Error("FactorsError: SDK is not initialised with token.");
    }
}

/**
 * Identify user with original 
 * userId from the application.
 * @param {string} customerUserId Actual id of the user from the application.
 */
function identify(customerUserId) {
    customerUserId = _validatedStringArg("customer_user_id", customerUserId);
    
    let payload = {};

    // Use user_id on cookie.
    if (Cookie.isExist(COOKIE_FUID)) 
        payload.user_id = Cookie.get(COOKIE_FUID);

    payload.c_uid = customerUserId;
    
    if (app && app.client.isInitialized()) {
        app.client.identify(payload)
            .then(_updateCookieIfUserIdInResponse)
            .catch(logger.error);
    } else {
        throw new Error("FactorsError: SDK is not initialised with token.");
    }
}

/**
 * Add additional user properties.
 * @param {Object} properties 
 */
function addUserProperties(properties) {}

let exposed = { app, isInstalled, init, reset, track, identify, addUserProperties };
if (process.env.NODE_ENV === "development") exposed["test"] = require("./test/suite.js");

module.exports = exports = exposed;

