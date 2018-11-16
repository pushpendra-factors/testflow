"use strict";

const logger = require("./utils/logger");
const util = require("./utils/util");

var _fa = require("./app");

// Global reference.
var app = new _fa.App();

/**
 * Prints SDK information, if installed.
 */
function isInstalled() {
    return "Factors sdk v0.2 is installed!";
}

/**
 * Initializes sdk environment on user application. Overwrites if initialized already.
 * 
 * @param {string} appToken Unique application token.
 */
function init(appToken) {
    appToken = util.validatedStringArg("token", appToken);
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
    if (!app || !app.isInitialized())
        throw new Error("FactorsError: SDK is not initialised with token.");

    eventName = util.validatedStringArg("event_name", eventName)

    let payload = {};
    _fa.updatePayloadWithUserIdFromCookie(payload);
    payload.event_name = eventName;
    payload.event_properties = eventProperties;

    return app.client.track(payload)
        .then(_fa.updateCookieIfUserIdInResponse)
        .catch(logger.error);
}

/**
 * Identify user with original 
 * userId from the application.
 * @param {string} customerUserId Actual id of the user from the application.
 */
function identify(customerUserId) {
    if (!app || !app.isInitialized())
        throw new Error("FactorsError: SDK is not initialised with token.");
    
    customerUserId = util.validatedStringArg("customer_user_id", customerUserId);
    
    let payload = {};
    _fa.updatePayloadWithUserIdFromCookie(payload);
    payload.c_uid = customerUserId;
    
    return app.client.identify(payload)
        .then(_fa.updateCookieIfUserIdInResponse)
        .catch(logger.error);
}

/**
 * Add additional user properties.
 * @param {Object} properties 
 */
function addUserProperties(properties={}) {
    if (!app || !app.isInitialized())
        throw new Error("FactorsError: SDK is not initialised with token.");

    if (typeof(properties) != "object")
        throw new Error("FactorsArgumentError: Properties should be an Object(key/values).");
    
    if (Object.keys(properties).length == 0)
        return Promise.reject("No changes. Empty properties.");
    
    let payload = {};
    _fa.updatePayloadWithUserIdFromCookie(payload);
    payload.properties = properties;

    return app.client.addUserProperties(payload)
        .then(_fa.updateCookieIfUserIdInResponse)
        .catch(logger.error);
}

let exposed = { isInstalled, init, reset, track, identify, addUserProperties };
if (process.env.NODE_ENV === "development") {
    exposed["test"] = require("./test/suite.js");
}
module.exports = exports = exposed;

