"use strict";

var Cookie = require("./utils/cookie");
const logger = require("./utils/logger");
const constant = require("./constant");

var App = require("./app");

// Common methods.

function _updateCookieIfUserIdInResponse(response){
    if (response && response.body && response.body.user_id) {
        let cleanUserId = response.body.user_id.trim();
        
        if (cleanUserId) 
            Cookie.setEncoded(constant.cookie.USER_ID, cleanUserId, constant.cookie.EXPIRY);
    }
    return response; // To continue chaining.
}

function _throwErrorOnFailureResponse(response, message="Request failed.") {
    if (response.status < 200 || response.status > 308) 
        throw new Error("FactorsRequestError: "+message);
    
    return response;
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


// Global reference.
var app = new App();

// Exposed methods.

/**
 * Prints SDK information, if installed.
 */
function isInstalled() {
    return "Factors sdk v0.1 is installed!";
}

/**
 * Initializes sdk environment on user application. Overwrites if initialized already.
 * 
 * @param {string} appToken Unique application token.
 */
function init(appToken) {
    appToken = _validatedStringArg("token", appToken);
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

    eventName = _validatedStringArg("event_name", eventName)

    let payload = {};
    _updatePayloadWithUserIdFromCookie(payload);
    payload.event_name = eventName;
    payload.event_properties = eventProperties;

    return app.client.track(payload)
        .then(_updateCookieIfUserIdInResponse)
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
    
    customerUserId = _validatedStringArg("customer_user_id", customerUserId);
    
    let payload = {};
    _updatePayloadWithUserIdFromCookie(payload);
    payload.c_uid = customerUserId;
    
    return app.client.identify(payload)
        .then(_updateCookieIfUserIdInResponse)
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
    _updatePayloadWithUserIdFromCookie(payload);
    payload.properties = properties;

    return app.client.addUserProperties(payload)
        .then(_updateCookieIfUserIdInResponse)
        .catch(logger.error);
}

let exposed = { isInstalled, init, reset, track, identify, addUserProperties };
if (process.env.NODE_ENV === "development") {
    exposed["test"] = require("./test/suite.js");
}
module.exports = exports = exposed;

