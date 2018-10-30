"use strict";

import * as Request from "./request";
import APIClient from "./api";
import config from "./config";

class App {
    constructor() {
        this.token = "";
        this.config = {};
    }

    setProperties(token, config={}) {
        this.token = token;
        this.config = config;
    }

    setToken(token) {
        this.token = token;
    }

    getToken() {
        return token;
    }
}

// Global references.
var app = new App();
var apiClient = new APIClient(config.api.host);

function isInstalled() {
    return "Factors sdk v0.1 is installed!";
}

/**
 * Initializes sdk environment on user application.
 * @param {string} token Unique application token.
 * @param {Object} appConfig Custom application configuration. i.e., { autoTrackPageView: true }
 */
function init(token, appConfig) {
    app.setProperties(token, appConfig);
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
            if exist:
                payload.user_id = cookie._fuid
        
		payload.event_name = eventName
        payload.properties = eventProperties

        request /track with payload
            response == 200 && response.body.user_id && response.body.user_id != "":
                cookie._fuid = response.body.user_id

	Todo(Dinesh): Do we need a _fident cookie for flaging identified?
     */

    apiClient.postEvent(
        app.getToken(),
        eventProperties
    );
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

export { isInstalled, init, track, identify, addUserProperties };

