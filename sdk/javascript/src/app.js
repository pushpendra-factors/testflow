"use strict";

var Cookie = require("./utils/cookie");
var logger = require("./utils/logger");

var APIClient = require("./api-client");

function App(token, config={}) {
    this.client = new APIClient(token);
    this.config = config;
}

App.prototype.isInitialized = function() {}

App.prototype.init = function(token, config={}) {
    if(token) this.client.setToken(token);
    this.config = config;
}

App.prototype.getClient = function() {
    return this.client;
}

// Constants.
const COOKIE_FUID = "_fuid";

// Global reference.
var app = new App(null, {});

function updateCookieIfUserIdInResponse(response){
    if (response && response.body && response.body.user_id) { 
        Cookie.set(COOKIE_FUID, response.body.user_id);
        return true;
    } else return false;
}

function isInstalled() {
    return "Factors sdk v0.1 is installed!";
}

/**
 * Initializes sdk environment on user application. Overwrites if initialized already.
 * @param {string} token Unique application token.
 * @param {Object} appConfig Custom application configuration. i.e., { autoTrackPageView: true }
 */
function init(token, appConfig) {
    app.set(token, appConfig);
}

/**
 * Track events on user application.
 * @param {string} eventName
 * @param {Object} eventProperties 
 */
function track(eventName, eventProperties) {
    /**
    (eventName, eventProperties)
        check cookie._fuid:
            if not exist:
                cookie._fuid = (create user) response.id
                
		payload.event_name = eventName
        payload.properties = eventProperties

        request /track with payload
            response == 200 && response.body.user_id && response.body.user_id != "":
                cookie._fuid = response.body.user_id

	Todo(Dinesh): Do we need a _fident cookie for flaging identified?
     */

    let payload = {};

    if (Cookie.isExist(COOKIE_FUID)) 
        payload.user_id = Cookie.get(COOKIE_FUID);

    if (app && app.client.isInitialized()) {
        app.client.track(null, eventName, eventProperties)
            .then(updateCookieIfUserIdInResponse)
            .catch(logger.error);

    } else {
        throw new Error("Tracking failed. Factors not initialized with token.");
    }
}

/**
 * Identify user with original 
 * userId from the application.
 * @param {string} customerUserId Actual id of the user from the application.
 */
function identify(customerUserId) {
    /**
    (customerUserId)
        payload = {}
        check cookie._fuid:
            if not exist:
                cookie._fuid = (create user) response.id
        
        payload.user_id = cookie._fuid
        payload.c_uid = customerUserId

        request /identify with payload:
            // if user_id already claimed as different user.
            if response == 200 && response.body.user_id && response.body.user_id != "":
                cookie._fuid = response.body.user_id
    */
}

/**
 * Add additional user properties.
 * @param {Object} properties 
 */
function addUserProperties(properties) {}

module.exports = exports = { isInstalled, app, init, track, identify, addUserProperties, Cookie };

