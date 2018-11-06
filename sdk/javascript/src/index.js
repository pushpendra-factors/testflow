"use strict";

var Cookie = require("./utils/cookie");
const logger = require("./utils/logger");
const constant = require("./constant");

var App = require("./app");

// Common methods.

function _updateCookieIfUserIdInResponse(response){
    if (response && response.body && response.body.user_id) {
        let cleanUserId = response.body.user_id.trim();

        if (cleanUserId) Cookie.setEncoded(constant.cookie.USER_ID, cleanUserId);
    }
    return response; // To continue chaining.
}

function _updatePayloadWithUserIdFromCookie(payload) {
    if (Cookie.isExist(constant.cookie.USER_ID))
        payload.user_id = Cookie.getDecoded(constant.cookie.USER_ID);
    
    return payload;
}

function _validatedStringArg(name, value) {
    if (typeof(value) != "string")
        throw new Error("FactorsArgumentError: Invalid type for "+name);
    
    value = value.trim();
    if (!value) throw new Error("FactorsArgumentError: "+name+" cannot be empty.");
    
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

/**
 * Clears existing SDK environment, both API token and cookies. 
 */
function reset() {
    app.reset();
    Cookie.remove(constant.cookie.USER_ID);
}

/**
 * Track events on user application.
 * @param {string} eventName
 * @param {Object} eventProperties 
 */
function track(eventName, eventProperties={}) {
    eventName = _validatedStringArg("event_name", eventName)

    let payload = {};
    _updatePayloadWithUserIdFromCookie(payload);
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
    _updatePayloadWithUserIdFromCookie(payload);
    payload.c_uid = customerUserId;
    
    if (app && app.client.isInitialized()) {
        return app.client.identify(payload)
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
function addUserProperties(properties={}) {
    if (typeof(properties) != "object")
        throw new Error("FactorsArgumentError: Properties should be an Object(key/values).");
    
    if (Object.keys(properties).length == 0)
        return Promise.reject("No changes. Empty properties.");

    let payload = {};
    _updatePayloadWithUserIdFromCookie(payload);
    payload.properties = properties;

    if (app && app.client.isInitialized()) {
        return app.client.addUserProperties(payload)
            .then(_updateCookieIfUserIdInResponse)
            .catch(logger.error);
    } else {
        throw new Error("FactorsError: SDK is not initialised with token.");
    }
}

let exposed = { app, isInstalled, init, reset, track, identify, addUserProperties };
if (process.env.NODE_ENV === "development") exposed["test"] = require("./test/suite.js");

module.exports = exports = exposed;

