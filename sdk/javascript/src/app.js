"use strict";

var Cookie = require("./utils/cookie");
var logger = require("./utils/logger");

var APIClient = require("./api-client");


// App class.

function App(token, config={}) {
    this.client = new APIClient(token);
    this.config = config;
}

App.prototype.isInitialized = function() {}

App.prototype.set = function(token, config={}) {
    if(token) this.client.setToken(token);
    this.config = config;
}

App.prototype.getClient = function() {
    return this.client;
}


// Common methods.

function _updateCookieIfUserIdInResponse(response){
    if (response && response.body && response.body.user_id) {
        let cleanUserId = response.body.user_id.trim();

        if (cleanUserId) Cookie.set(COOKIE_FUID, cleanUserId);
    }
}

function _validatedStringArg(name, value) {
    if (typeof(value) != "string")
        throw new TypeError("FactorsError: Invalid type for "+name);
    
    value = value.trim();
    if (!value) throw new Error("FactorsError: "+name+" cannot be empty.");
    
    return value;
}


// Exposed methods.

// Constants.
const COOKIE_FUID = "_fuid";

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
    app.set(appToken, appConfig);
}

/**
 * Track events on user application.
 * @param {string} eventName
 * @param {Object} eventProperties 
 */
function track(eventName, eventProperties) {
    eventName = _validatedStringArg("event_name", eventName)

    let payload = {};

    // Use user_id on cookie.
    if (Cookie.isExist(COOKIE_FUID)) 
        payload.user_id = Cookie.get(COOKIE_FUID);
    
    payload.event_name = eventName;
    payload.event_properties = eventProperties;

    if (app && app.client.isInitialized()) {
        app.client.track(payload)
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

module.exports = exports = { isInstalled, init, track, identify, addUserProperties };

