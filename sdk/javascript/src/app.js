"use strict";

import APIClient from "./api";
import * as Cookie from "./utils/cookie"

class App {
    constructor(token, config={}) {
        this.client = new APIClient(token);
        this.config = config;
    }

    setToken(token) {
        this.client = new APIClient(token);
    }

    setConfig(config) {
        this.config = config;
    }

    getClient() {
        this.client;
    }
}

// Global references.
var app = null;

function isInstalled() {
    return "Factors sdk v0.1 is installed!";
}

function isInitialized() {
    return app != null || app != undefined;
}

/**
 * Initializes sdk environment on user application. Overwrites if initialized already.
 * @param {string} token Unique application token.
 * @param {Object} appConfig Custom application configuration. i.e., { autoTrackPageView: true }
 */
function init(token, appConfig) {
    app = new App(token, appConfig);
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

export { isInstalled, app, init, track, identify, addUserProperties, Cookie };

